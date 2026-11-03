// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
)

// The kinds, states and edge kinds of the node family's graph (the
// contract's TopologyNode and TopologyEdge enums).
const (
	topoKindNode       = "node"
	topoKindAggregator = "aggregator"
	topoKindDevice     = "device"
	topoKindPart       = "part"
	topoKindFabric     = "fabric"
	topoKindSession    = "session"

	topoStateOK          = "ok"
	topoStateUnreachable = "unreachable"
	topoStateIdle        = "idle"
	topoStateUnknown     = "unknown"

	topoEdgeContains = "contains"
	topoEdgeSession  = "session"
)

// topologyRef is the page that already lists a graph node and the id it
// lists it by.
type topologyRef struct {
	Page string `json:"page"`
	ID   string `json:"id"`
}

// topologyDetail is one key/value pair a page shows for a graph node.
type topologyDetail struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// topologyNode is one node of GET /node/topology (the contract's
// TopologyNode): the typed fields only for the kinds they belong to.
type topologyNode struct {
	ID            string           `json:"id"`
	Kind          string           `json:"kind"`
	Label         string           `json:"label"`
	State         string           `json:"state"`
	Ref           *topologyRef     `json:"ref,omitempty"`
	Details       []topologyDetail `json:"details"`
	Endpoint      *uint16          `json:"endpoint,omitempty"`
	FabricIndex   *uint8           `json:"fabric_index,omitempty"`
	SessionID     *uint16          `json:"session_id,omitempty"`
	NodeIDHex     string           `json:"node_id_hex,omitempty"`
	Reachable     *bool            `json:"reachable,omitempty"`
	IsPASE        *bool            `json:"is_pase,omitempty"`
	Subscriptions *int             `json:"subscriptions,omitempty"`
}

// topologyEdge is one edge of the graph (the contract's TopologyEdge).
type topologyEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Kind   string `json:"kind"`
}

// topologyGraph is the body of GET /node/topology (the contract's
// TopologyGraph with family "node").
type topologyGraph struct {
	AsOf   string         `json:"as_of"`
	Family string         `json:"family"`
	Nodes  []topologyNode `json:"nodes"`
	Edges  []topologyEdge `json:"edges"`
}

// topology serves GET /node/topology: the node, its bridge and the
// controllers talking to it as one graph, built from the three facade
// reads the endpoints, fabrics and sessions routes already serve
// (node.Topology, FabricInfos, SessionDiagnostics) — concept U3 §C.2. The
// graph is assembled here rather than in the page so both families share
// one contract type (brief U3a decision 0.1).
func (h *handler) topology(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	topo := h.node.topology()
	if topo.Aggregator == 0 {
		// node.Topology is empty before the node assembled its endpoints
		// (go-fabric node/lifecycle.go BridgeTopology.Aggregator).
		unready(w, r, "Node not assembled", "the node has not assembled its endpoints yet")
		return
	}
	fabrics, err := h.node.fabrics(r.Context())
	if err != nil {
		h.serverError(w, r, "Failed to read the fabrics", err)
		return
	}
	healthy, _, healthOK := h.node.health()
	writeJSON(w, buildNodeGraph(topo, fabrics, h.node.sessions(), h.node.localAddr(), healthOK, healthy))
}

// buildNodeGraph is the node family's graph:
//
//	node ─contains→ ep:<aggregator> ─contains→ device ─contains→ part
//	fabric:<index> ─contains→ session:<id> ─session→ node
//
// A session whose fabric the node does not list (a PASE session, fabric
// index 0) hangs on the node alone. Nodes come in that order: the node, the
// aggregator, each device followed by its parts, each fabric, the sessions
// by fabric and then those without one — the order the page lays rows in.
func buildNodeGraph(topo node.BridgeTopology, fabrics []node.FabricInfo, diag node.SessionDiagnostics, addr string, healthOK, healthy bool) topologyGraph {
	g := topologyGraph{
		AsOf:   diag.AsOf.UTC().Format(time.RFC3339),
		Family: "node",
		Nodes:  []topologyNode{},
		Edges:  []topologyEdge{},
	}
	edge := func(src, dst, kind string) {
		g.Edges = append(g.Edges, topologyEdge{Source: src, Target: dst, Kind: kind})
	}

	root := topologyNode{
		ID: "node", Kind: topoKindNode, Label: "Node", State: nodeState(healthOK, healthy),
		Ref: &topologyRef{Page: "status", ID: ""}, Details: []topologyDetail{},
	}
	if addr != "" {
		root.Details = append(root.Details, topologyDetail{"address", addr})
	}
	root.Details = append(root.Details,
		topologyDetail{"endpoints", strconv.Itoa(countDevices(topo.Devices))},
		topologyDetail{"fabrics", strconv.Itoa(len(fabrics))},
		topologyDetail{"sessions", strconv.Itoa(len(diag.Sessions))},
	)
	g.Nodes = append(g.Nodes, root)

	aggID := endpointNodeID(topo.Aggregator)
	g.Nodes = append(g.Nodes, topologyNode{
		ID: aggID, Kind: topoKindAggregator, Label: "Aggregator", State: topoStateOK,
		Ref:      &topologyRef{Page: "endpoints", ID: strconv.Itoa(int(topo.Aggregator))},
		Details:  []topologyDetail{{"endpoint", strconv.Itoa(int(topo.Aggregator))}, {"devices", strconv.Itoa(len(topo.Devices))}},
		Endpoint: ptr(topo.Aggregator),
	})
	edge("node", aggID, topoEdgeContains)

	var walk func(parent string, devs []node.BridgedDevice, kind string)
	walk = func(parent string, devs []node.BridgedDevice, kind string) {
		for i := range devs {
			d := &devs[i]
			id := endpointNodeID(d.Endpoint)
			g.Nodes = append(g.Nodes, deviceNode(d, id, kind))
			edge(parent, id, topoEdgeContains)
			walk(id, d.Parts, topoKindPart)
		}
	}
	walk(aggID, topo.Devices, topoKindDevice)

	listed := make(map[uint8]bool, len(fabrics))
	for i := range fabrics {
		f := &fabrics[i]
		listed[f.FabricIndex] = true
		g.Nodes = append(g.Nodes, fabricNode(f))
	}
	// Sessions by fabric in the fabrics' order, then those on no listed
	// fabric, each group by session id as SessionDiagnostics lists them.
	for i := range fabrics {
		for j := range diag.Sessions {
			if s := &diag.Sessions[j]; s.FabricIndex == fabrics[i].FabricIndex && !s.PASE {
				id := sessionNodeID(s.ID)
				g.Nodes = append(g.Nodes, sessionNode(s))
				edge(fabricNodeID(s.FabricIndex), id, topoEdgeContains)
				edge(id, "node", topoEdgeSession)
			}
		}
	}
	for j := range diag.Sessions {
		if s := &diag.Sessions[j]; s.PASE || !listed[s.FabricIndex] {
			id := sessionNodeID(s.ID)
			g.Nodes = append(g.Nodes, sessionNode(s))
			edge(id, "node", topoEdgeSession)
		}
	}
	return g
}

// nodeState is the node's state from its health sample: unknown without
// one, unreachable when the probe found the listener down.
func nodeState(ok, healthy bool) string {
	switch {
	case !ok:
		return topoStateUnknown
	case healthy:
		return topoStateOK
	default:
		return topoStateUnreachable
	}
}

func countDevices(devs []node.BridgedDevice) int {
	n := len(devs)
	for i := range devs {
		n += countDevices(devs[i].Parts)
	}
	return n
}

func endpointNodeID(ep uint16) string { return "ep:" + strconv.Itoa(int(ep)) }
func fabricNodeID(idx uint8) string   { return "fabric:" + strconv.Itoa(int(idx)) }
func sessionNodeID(id uint16) string  { return "session:" + strconv.Itoa(int(id)) }

// deviceNode is a bridged device or a part: named by the host (its
// NodeLabel), reachable as a controller reads it, with its device type and
// revision, its cluster count and its stable key.
func deviceNode(d *node.BridgedDevice, id, kind string) topologyNode {
	typeName := deviceTypeName(d.DeviceType)
	label := d.Name
	if label == "" {
		label = typeName
	}
	state := topoStateOK
	if !d.Reachable {
		state = topoStateUnreachable
	}
	details := []topologyDetail{
		{"endpoint", strconv.Itoa(int(d.Endpoint))},
		{"device_type", typeName},
	}
	if d.DeviceTypeRevision != 0 {
		details = append(details, topologyDetail{"device_type_revision", strconv.Itoa(int(d.DeviceTypeRevision))})
	}
	details = append(details, topologyDetail{"clusters", strconv.Itoa(len(d.Clusters))})
	if d.Key != nil {
		details = append(details, topologyDetail{"key", d.Key.String()})
	}
	return topologyNode{
		ID: id, Kind: kind, Label: label, State: state,
		Ref:      &topologyRef{Page: "endpoints", ID: strconv.Itoa(int(d.Endpoint))},
		Details:  details,
		Endpoint: ptr(d.Endpoint), Reachable: ptr(d.Reachable),
	}
}

// fabricNode is one fabric the node is commissioned into, labelled by its
// label, else by its controller's vendor (eligibility.VendorName, the id
// itself for a vendor the table does not name).
func fabricNode(f *node.FabricInfo) topologyNode {
	vendor := eligibility.VendorName(f.VendorID)
	label := f.Label
	if label == "" {
		label = vendor
	}
	details := []topologyDetail{}
	if f.Label != "" {
		details = append(details, topologyDetail{"label", f.Label})
	}
	details = append(details,
		topologyDetail{"vendor", vendor},
		topologyDetail{"vendor_id", fmt.Sprintf("0x%04X", f.VendorID)},
		topologyDetail{"node_id", hex64(f.NodeID)},
		topologyDetail{"fabric_id", hex64(f.FabricID)},
	)
	return topologyNode{
		ID: fabricNodeID(f.FabricIndex), Kind: topoKindFabric, Label: label, State: topoStateOK,
		Ref:         &topologyRef{Page: "fabrics", ID: strconv.Itoa(int(f.FabricIndex))},
		Details:     details,
		FabricIndex: ptr(f.FabricIndex), NodeIDHex: hex64(f.NodeID),
	}
}

// sessionNode is one secure session: labelled by its peer's node id (a
// commissioning session by its kind), idle when the peer has been quiet
// longer than the idle interval it reported in its handshake — a session
// whose peer reported none is never idle.
func sessionNode(s *node.SessionInfo) topologyNode {
	kind, label := "CASE", hex64(s.PeerNodeID)
	if s.PASE {
		kind, label = "PASE", "PASE"
	}
	state := topoStateOK
	if s.PeerIdleInterval > 0 && s.PeerIdle > s.PeerIdleInterval {
		state = topoStateIdle
	}
	return topologyNode{
		ID: sessionNodeID(s.ID), Kind: topoKindSession, Label: label, State: state,
		Ref: &topologyRef{Page: "diagnostics", ID: strconv.Itoa(int(s.ID))},
		Details: []topologyDetail{
			{"session_type", kind},
			{"peer_node_id", hex64(s.PeerNodeID)},
			{"subscriptions", strconv.Itoa(s.Subscriptions)},
			{"peer_idle_seconds", strconv.FormatInt(int64(s.PeerIdle.Seconds()), 10)},
		},
		SessionID: ptr(s.ID), FabricIndex: ptr(s.FabricIndex), NodeIDHex: hex64(s.PeerNodeID),
		IsPASE: ptr(s.PASE), Subscriptions: ptr(s.Subscriptions),
	}
}
