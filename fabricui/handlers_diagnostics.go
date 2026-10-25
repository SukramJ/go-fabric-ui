// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"net/http"
	"slices"
	"time"

	"github.com/SukramJ/go-fabric/contract"
	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
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
}

// sessions serves GET /node/sessions: whether a commissioned controller is
// still talking to the node. A controller that vanished without closing
// its session only stops sending, so the peer idle age is the load-bearing
// field. Mirrors loom matter.go:374-400 (MatterSessions).
func (h *handler) sessions(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	now := h.now()
	infos, occ := h.node.sessions()
	out := make([]sessionResponse, 0, len(infos))
	for _, s := range infos {
		out = append(out, sessionResponse{
			SessionID:        s.SessionID,
			FabricIndex:      s.FabricIndex,
			PeerNodeID:       s.PeerNodeID,
			PeerNodeIDHex:    hex64(s.PeerNodeID),
			LocalNodeID:      s.LocalNodeID,
			LocalNodeIDHex:   hex64(s.LocalNodeID),
			IsPASE:           s.IsPASE,
			Subscriptions:    s.Subscriptions,
			LastActivity:     s.LastActivity.UTC().Format(time.RFC3339),
			LastPeerActivity: s.LastPeerActivity.UTC().Format(time.RFC3339),
			IdleSeconds:      int64(now.Sub(s.LastActivity).Seconds()),
			PeerIdleSeconds:  int64(now.Sub(s.LastPeerActivity).Seconds()),
		})
	}
	writeJSON(w, sessionList{Sessions: out, Occupancy: occ})
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

// clusterResponse is one cluster of an endpoint. Name and revision are
// absent: the facade does not name clusters (a CK question).
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
	EndpointID       uint16            `json:"endpoint_id"`
	ParentEndpointID uint16            `json:"parent_endpoint_id"`
	DeviceType       uint16            `json:"device_type"`
	DeviceTypeName   string            `json:"device_type_name"`
	Reachable        bool              `json:"reachable"`
	Name             string            `json:"name"`
	Source           map[string]string `json:"source"`
	Clusters         []clusterResponse `json:"clusters"`
}

type endpointList struct {
	Endpoints []endpointResponse `json:"endpoints"`
}

// endpoints serves GET /node/endpoints: what a controller sees, without
// commissioning one first. Mirrors loom matter.go:496-506
// (MatterEndpoints) and cmd/openccu-loom/daemon_matter.go:1307-1360
// (matterEndpointInspector), on node.Topology.
func (h *handler) endpoints(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	writeJSON(w, endpointList{Endpoints: flattenTopology(h.node.topology())})
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
				EndpointID:       d.Endpoint,
				ParentEndpointID: parent,
				DeviceType:       d.DeviceType,
				DeviceTypeName:   contract.DeviceTypeName(d.DeviceType),
				Reachable:        d.Reachable,
				Name:             d.Name,
				Source:           map[string]string{},
				Clusters:         clustersOf(d.Added),
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

// clustersOf lists the server clusters an added endpoint's source serves,
// by id and sorted; empty for a part and for a host's own snapshotter
// endpoint, which carry no Added.
func clustersOf(ep node.Endpoint) []clusterResponse {
	out := []clusterResponse{}
	if ep == nil {
		return out
	}
	src := ep.Spec().Source
	if src == nil {
		return out
	}
	var ids []uint32
	for _, cs := range src.MatterClusterServers() {
		if cs == nil {
			continue
		}
		if id := cs.MatterClusterID(); !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	for _, id := range ids {
		out = append(out, clusterResponse{ID: id})
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

// diagnosticEvent is one entry of GET /node/events (loom matter.go:558-565).
type diagnosticEvent struct {
	At       string            `json:"at"`
	Kind     string            `json:"kind"`
	Severity string            `json:"severity"`
	Message  string            `json:"message"`
	Detail   map[string]string `json:"detail,omitempty"`
}

type diagnosticEventList struct {
	Events []diagnosticEvent `json:"events"`
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
	out := make([]diagnosticEvent, 0, len(events))
	for _, e := range events {
		out = append(out, diagnosticEvent{
			At:       e.At.UTC().Format(time.RFC3339),
			Kind:     string(e.Kind),
			Severity: string(e.Severity),
			Message:  e.Message,
			Detail:   e.Detail,
		})
	}
	writeJSON(w, diagnosticEventList{Events: out})
}
