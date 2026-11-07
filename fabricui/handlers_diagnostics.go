// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/internal/api"
)

// sessionResponse is one session of GET /node/sessions. Mirrors loom's
// MatterSessionResponse (internal/north/rest/handlers/matter.go:320-331 at
// b40b289a): the node ids exactly as hex strings (the numeric twins are
// informational), the idle ages in seconds.
type sessionResponse struct {
	SessionID        uint16 `json:"session_id"`
	FabricIndex      uint8  `json:"fabric_index"`
	PeerNodeID       uint64 `json:"peer_node_id"`
	PeerNodeIDHex    string `json:"peer_node_id_hex"`
	LocalNodeID      uint64 `json:"local_node_id"`
	LocalNodeIDHex   string `json:"local_node_id_hex"`
	IsPASE           bool   `json:"is_pase"`
	Subscriptions    int    `json:"subscriptions"`
	LastActivity     string `json:"last_activity"`
	LastPeerActivity string `json:"last_peer_activity"`
	IdleSeconds      int64  `json:"idle_seconds"`
	PeerIdleSeconds  int64  `json:"peer_idle_seconds"`
}

// sessionOf is one session as GET /node/sessions and the stream's session
// events carry it.
func sessionOf(s *node.SessionInfo) sessionResponse {
	return sessionResponse{
		SessionID:        s.ID,
		FabricIndex:      s.FabricIndex,
		PeerNodeID:       s.PeerNodeID,
		PeerNodeIDHex:    hex64(s.PeerNodeID),
		LocalNodeID:      s.LocalNodeID,
		LocalNodeIDHex:   hex64(s.LocalNodeID),
		IsPASE:           s.PASE,
		Subscriptions:    s.Subscriptions,
		LastActivity:     s.LastActivity.UTC().Format(time.RFC3339),
		LastPeerActivity: s.LastPeerActivity.UTC().Format(time.RFC3339),
		IdleSeconds:      int64(s.Idle.Seconds()),
		PeerIdleSeconds:  int64(s.PeerIdle.Seconds()),
	}
}

// occupancyResponse is how much of the session-id space is in use,
// including ids staked by handshakes that never completed (loom
// matter.go:333-347).
type occupancyResponse struct {
	Live     int `json:"live"`
	Reserved int `json:"reserved"`
	Capacity int `json:"capacity"`
	Free     int `json:"free"`
}

type sessionList struct {
	Sessions  []sessionResponse `json:"sessions"`
	Occupancy occupancyResponse `json:"occupancy"`
	// Parameters are the session parameters the node reports (0.7.0).
	Parameters *sessionParameters `json:"parameters,omitempty"`
}

// sessionParameters is the contract's SessionParameters: the fields of
// go-fabric's sessionparams.Parameters (secure/sessionparams/
// sessionparams.go:49-66 at v0.18.0) as Node.SessionParameters returns
// them, the specification version as 8 hex digits.
type sessionParameters struct {
	IdleIntervalMS           uint32  `json:"idle_interval_ms"`
	ActiveIntervalMS         uint32  `json:"active_interval_ms"`
	ActiveThresholdMS        uint16  `json:"active_threshold_ms"`
	DataModelRevision        uint16  `json:"data_model_revision"`
	InteractionModelRevision uint16  `json:"interaction_model_revision"`
	SpecificationVersionHex  string  `json:"specification_version_hex"`
	MaxPathsPerInvoke        uint16  `json:"max_paths_per_invoke"`
	SupportedTransports      uint16  `json:"supported_transports"`
	TCP                      bool    `json:"tcp"`
	MaxTCPMessageSize        *uint32 `json:"max_tcp_message_size,omitempty"`
}

// sessionParametersOf is p on the wire; max_tcp_message_size only with
// TCP, as the structure encodes it (sessionparams.go:59-63).
func sessionParametersOf(p *node.SessionParameters) *sessionParameters {
	out := &sessionParameters{
		IdleIntervalMS: p.SessionIdleInterval, ActiveIntervalMS: p.SessionActiveInterval,
		ActiveThresholdMS: p.SessionActiveThreshold, DataModelRevision: p.DataModelRevision,
		InteractionModelRevision: p.InteractionModelRevision,
		SpecificationVersionHex:  fmt.Sprintf("%08X", p.SpecificationVersion),
		MaxPathsPerInvoke:        p.MaxPathsPerInvoke, SupportedTransports: p.SupportedTransports,
		TCP: p.SupportsTCP(),
	}
	if out.TCP {
		out.MaxTCPMessageSize = &p.MaxTCPMessageSize
	}
	return out
}

// sessions serves GET /node/sessions: whether a commissioned controller is
// still talking to the node. A controller that vanished without closing
// its session only stops sending, so the peer idle age is the load-bearing
// field. Mirrors loom matter.go:374-400 (MatterSessions), on
// node.SessionDiagnostics: the idle ages are measured on the node's clock,
// 0 for a session that has not carried traffic yet.
func (h *handler) sessions(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	diag := h.node.sessions()
	out := make([]sessionResponse, 0, len(diag.Sessions))
	for i := range diag.Sessions {
		out = append(out, sessionOf(&diag.Sessions[i]))
	}
	occ := diag.Occupancy
	params := h.node.sessionParameters()
	writeJSON(w, sessionList{Sessions: out, Occupancy: occupancyResponse{
		Live: occ.Live, Reserved: occ.Reserved, Capacity: occ.Capacity, Free: occ.Free,
	}, Parameters: sessionParametersOf(&params)})
}

// mdnsService is one advertised record (loom matter.go:404-413).
type mdnsService struct {
	ServiceType  string            `json:"service_type"`
	InstanceName string            `json:"instance_name"`
	HostName     string            `json:"host_name"`
	Port         uint16            `json:"port"`
	Addresses    []string          `json:"addresses"`
	Subtypes     []string          `json:"subtypes"`
	TXT          map[string]string `json:"txt"`
}

// mdnsFinding explains one reason a controller might not discover the
// node (loom matter.go:415-422).
type mdnsFinding struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Service  string `json:"service,omitempty"`
}

// mdnsResponse is the body of GET /node/mdns (loom matter.go:424-429).
type mdnsResponse struct {
	Advertising bool          `json:"advertising"`
	Services    []mdnsService `json:"services"`
	Findings    []mdnsFinding `json:"findings"`
}

// mdns serves GET /node/mdns: what is actually announced, and the
// conditions that keep a controller from listing the node. Mirrors loom
// matter.go:445-457 (MatterMdns).
func (h *handler) mdns(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	writeJSON(w, h.node.mdns())
}

// reannounce serves POST /node/mdns/reannounce (contract 0.7.0): the node
// announces its DNS-SD records now (Node.TriggerReannounce, go-fabric
// node/reannounce.go:12-26 at v0.18.0), a no-op on an advertiser that
// cannot.
func (h *handler) reannounce(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	h.node.reannounce(r.Context())
	h.log.InfoContext(r.Context(), "fabricui.mdns.reannounced")
	w.WriteHeader(http.StatusNoContent)
}

// clusterResponse is one cluster server of an endpoint, named and revised
// from go-fabric's generated schema (node.ClusterInfo); name and revision
// are absent for a cluster the schema does not know.
type clusterResponse struct {
	ID       uint32 `json:"id"`
	Name     string `json:"name,omitempty"`
	Revision uint16 `json:"revision,omitempty"`
}

// endpointResponse is one bridged endpoint of GET /node/endpoints. Mirrors
// loom's MatterEndpointInfo (matter.go:465-477) without its host fields:
// the host's name for the endpoint is name, and its identity travels as
// the opaque source map ({"key": the stable key}).
type endpointResponse struct {
	EndpointID       uint16 `json:"endpoint_id"`
	ParentEndpointID uint16 `json:"parent_endpoint_id"`
	DeviceType       uint16 `json:"device_type"`
	DeviceTypeName   string `json:"device_type_name"`
	// DeviceTypeRevision is the revision the endpoint's Descriptor serves
	// for its device type; absent for a type the schema does not know.
	DeviceTypeRevision uint16            `json:"device_type_revision,omitempty"`
	Reachable          bool              `json:"reachable"`
	Name               string            `json:"name"`
	Source             map[string]string `json:"source"`
	Clusters           []clusterResponse `json:"clusters"`
}

type endpointList struct {
	Endpoints  []endpointResponse   `json:"endpoints"`
	Structural []structuralEndpoint `json:"structural"`
}

// structuralEndpoint is one endpoint of the node's endpoint tree no
// bridged device declares (the contract's StructuralEndpoint).
type structuralEndpoint struct {
	EndpointID       uint16                   `json:"endpoint_id"`
	ParentEndpointID *uint16                  `json:"parent_endpoint_id,omitempty"`
	Name             string                   `json:"name"`
	DeviceTypes      []api.EndpointDeviceType `json:"device_types"`
	Parts            []uint16                 `json:"parts"`
}

// endpoints serves GET /node/endpoints: what a controller sees, without
// commissioning one first. Mirrors loom matter.go:496-506
// (MatterEndpoints) and cmd/openccu-loom/daemon_matter.go:1307-1360
// (matterEndpointInspector), on node.Topology; the root and the
// Aggregator from node.Endpoints.
func (h *handler) endpoints(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	bridged := flattenTopology(h.node.topology())
	writeJSON(w, endpointList{Endpoints: bridged, Structural: structuralEndpoints(h.node.endpointInfos(), bridged)})
}

// structuralEndpoints are the endpoints of infos (Node.Endpoints, the root
// first, then by number) that bridged does not list — the root and the
// Aggregator — with their tree parent, name, device types and direct
// children as the endpoint detail renders them (readEndpointDetail).
func structuralEndpoints(infos []node.EndpointInfo, bridged []endpointResponse) []structuralEndpoint {
	out := []structuralEndpoint{}
	for i := range infos {
		info := &infos[i]
		if slices.ContainsFunc(bridged, func(e endpointResponse) bool { return e.EndpointID == info.Endpoint }) {
			continue
		}
		e := structuralEndpoint{
			EndpointID: info.Endpoint, ParentEndpointID: treeParent(info), Name: info.Name,
			DeviceTypes: []api.EndpointDeviceType{}, Parts: childrenOf(infos, info.Endpoint),
		}
		dts := deviceTypesOf(info.DeviceTypes)
		for _, dt := range dts {
			e.DeviceTypes = append(e.DeviceTypes, api.EndpointDeviceType{ID: int(dt.ID), IDHex: hexID(dt.ID), Name: optString(dt.Name), Revision: int(dt.Revision)})
		}
		if e.Name == "" && len(dts) > 0 {
			e.Name = dts[0].Name
		}
		out = append(out, e)
	}
	return out
}

// flattenTopology lists every bridged endpoint, parts after the device
// that holds them; a top-level device's parent is the Aggregator.
func flattenTopology(topo node.BridgeTopology) []endpointResponse {
	out := []endpointResponse{}
	var walk func(parent uint16, devs []node.BridgedDevice)
	walk = func(parent uint16, devs []node.BridgedDevice) {
		for i := range devs {
			d := &devs[i]
			ep := endpointResponse{
				EndpointID:         d.Endpoint,
				ParentEndpointID:   parent,
				DeviceType:         d.DeviceType,
				DeviceTypeName:     deviceTypeName(d.DeviceType),
				DeviceTypeRevision: d.DeviceTypeRevision,
				Reachable:          d.Reachable,
				Name:               d.Name,
				Source:             map[string]string{},
				Clusters:           clustersOf(d.Clusters),
			}
			if d.Key != nil {
				ep.Source["key"] = d.Key.String()
			}
			out = append(out, ep)
			walk(d.Endpoint, d.Parts)
		}
	}
	walk(topo.Aggregator, topo.Devices)
	return out
}

// clustersOf lists the cluster servers mounted on an endpoint
// (node.BridgedDevice.Clusters) in the order its Descriptor's ServerList
// lists them — for every endpoint, parts and a host's own snapshotter
// endpoints included.
func clustersOf(infos []node.ClusterInfo) []clusterResponse {
	out := make([]clusterResponse, 0, len(infos))
	for _, c := range infos {
		out = append(out, clusterResponse{ID: c.ID, Name: c.Name, Revision: c.Revision})
	}
	return out
}

// ecosystemResponse is one fabric classified by its controller's vendor
// (loom matter.go:509-515).
type ecosystemResponse struct {
	Ecosystem   string `json:"ecosystem"`
	VendorID    uint16 `json:"vendor_id"`
	FabricIndex uint8  `json:"fabric_index"`
	Label       string `json:"label,omitempty"`
}

// compatFinding warns about one ecosystem and device type (loom
// matter.go:517-523).
type compatFinding struct {
	Ecosystem  string `json:"ecosystem"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	DeviceType uint16 `json:"device_type,omitempty"`
}

type compatibilityResponse struct {
	Ecosystems    []ecosystemResponse `json:"ecosystems"`
	EndpointCount int                 `json:"endpoint_count"`
	Findings      []compatFinding     `json:"findings"`
}

// compatibility serves GET /node/compatibility: the commissioned fabrics
// joined with the exposed endpoints, the one place both halves are known.
// Mirrors loom cmd/openccu-loom/daemon_matter.go:1362-1418
// (matterCompatibilityReporter) on eligibility.Compat.
func (h *handler) compatibility(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	out := compatibilityResponse{Ecosystems: []ecosystemResponse{}, Findings: []compatFinding{}}
	var vendorIDs []uint16
	if infos, err := h.node.fabrics(r.Context()); err == nil {
		for _, f := range infos {
			vendorIDs = append(vendorIDs, f.VendorID)
			out.Ecosystems = append(out.Ecosystems, ecosystemResponse{
				Ecosystem:   string(eligibility.EcosystemForVendor(f.VendorID)),
				VendorID:    f.VendorID,
				FabricIndex: f.FabricIndex,
				Label:       f.Label,
			})
		}
	}
	deviceTypes := map[uint16]int{}
	for _, ep := range flattenTopology(h.node.topology()) {
		deviceTypes[ep.DeviceType]++
		out.EndpointCount++
	}
	for _, f := range eligibility.Compat(vendorIDs, deviceTypes, out.EndpointCount) {
		out.Findings = append(out.Findings, compatFinding{
			Ecosystem: string(f.Ecosystem), Code: f.Code, Message: f.Message, DeviceType: f.DeviceType,
		})
	}
	writeJSON(w, out)
}

type diagnosticEventList struct {
	Events []TraceEvent `json:"events"`
}

// diagnosticEvents serves GET /node/events: the pairing, session and
// discovery trace, newest first. A host without a node answers 503 rather
// than an empty list — an empty trace on a disabled node would read as
// "nothing happened" (loom matter.go:572-600).
func (h *handler) diagnosticEvents(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	events := h.node.diagnosticEvents()
	out := make([]TraceEvent, 0, len(events))
	for i := range events {
		out = append(out, traceEventOf(&events[i]))
	}
	writeJSON(w, diagnosticEventList{Events: out})
}
