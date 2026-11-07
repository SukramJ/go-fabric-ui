// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SukramJ/go-fabric/controller/topology"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// controllerID is the controller's node in the graph (the port's
// NetworkNode.ID for the controller).
const controllerID = "node"

// NetworkTopology implements [fabricui.ControllerSource] over the host's
// topology service (ControllerOptions.Topology): its snapshot as it
// stands, or with refresh the snapshot after the nodes' network
// diagnostics were read again. Without a service the route answers 501.
func (s *source) NetworkTopology(ctx context.Context, refresh bool) (fabricui.NetworkTopology, error) {
	svc := s.opts.Topology
	if svc == nil {
		return fabricui.NetworkTopology{}, fabricui.ErrNotImplemented
	}
	var t topology.Topology
	if refresh {
		var err error
		if t, err = svc.Refresh(ctx); err != nil {
			// Refresh fails only when ctx ends while the reads run.
			return fabricui.NetworkTopology{}, fmt.Errorf("%w: %w", fabricui.ErrNodeFailed, err)
		}
	} else {
		t = svc.Topology()
	}
	root := "Controller"
	if f := s.Fabric(); f.Label != "" {
		root = f.Label
	}
	names := map[uint64]string{}
	for _, n := range s.Nodes() {
		if b := n.Basic(); b.NodeLabel != "" {
			names[n.ID()] = b.NodeLabel
		} else if b.ProductName != "" {
			names[n.ID()] = b.ProductName
		}
	}
	return networkTopologyOf(t, root, names), nil
}

// watchTopology subscribes the source to the service's changes: every
// snapshot that moved is a topology_changed on the controller stream.
// The callback only publishes, it never calls Refresh (the service's
// rule for its observers).
func (s *source) watchTopology(svc *topology.Service) {
	unsubscribe := svc.OnChanged(func(topology.Topology) {
		s.publish(fabricui.ControllerEvent{Type: fabricui.EventTopologyChanged})
	})
	s.topoMu.Lock()
	s.unwatch = unsubscribe
	s.topoMu.Unlock()
}

// closeTopology releases the subscriptions of watchTopology and
// watchUpdates; Hooks.Close runs it when hooks are bound, a test without
// hooks calls it. It is idempotent.
func (s *source) closeTopology() {
	s.topoMu.Lock()
	unsubscribe, unobserve := s.unwatch, s.unobserve
	s.unwatch, s.unobserve = nil, nil
	s.topoMu.Unlock()
	if unsubscribe != nil {
		unsubscribe()
	}
	if unobserve != nil {
		unobserve()
	}
}

// topoState is the subscription state of a source; embedded in source:
// unwatch releases the topology service's, unobserve the update
// manager's.
type topoState struct {
	topoMu    sync.Mutex
	unwatch   func()
	unobserve func()
}

// networkTopologyOf is go-fabric's topology snapshot as the port's graph:
// the controller as the root node with a session edge to every
// commissioned node, the commissioned nodes as peers, the Thread
// neighbours no commissioned node matches as thread_router or
// thread_end_device, the access points as wifi_ap, and one link edge
// per connection. root labels the controller, names the commissioned
// nodes by node id (as the Nodes page names them; the hex id otherwise).
func networkTopologyOf(t topology.Topology, root string, names map[uint64]string) fabricui.NetworkTopology {
	ids := make(map[string]string, len(t.Nodes))
	peers := 0
	for i := range t.Nodes {
		n := &t.Nodes[i]
		ids[n.ID] = graphID(n)
		if n.Kind == topology.KindMatter {
			peers++
		}
	}
	out := fabricui.NetworkTopology{
		CollectedAt: t.CollectedAt,
		Nodes:       make([]fabricui.NetworkNode, 0, len(t.Nodes)+1),
		Connections: make([]fabricui.NetworkConnection, 0, peers+len(t.Connections)),
	}
	out.Nodes = append(out.Nodes, fabricui.NetworkNode{
		ID: controllerID, Kind: "node", Label: root, State: "ok",
		Details: []fabricui.TopologyDetail{
			{Key: "nodes", Value: fmt.Sprint(peers)},
			{Key: "collected_at", Value: t.CollectedAt.UTC().Format(time.RFC3339)},
		},
	})
	for i := range t.Nodes {
		n := &t.Nodes[i]
		switch n.Kind {
		case topology.KindMatter:
			out.Nodes = append(out.Nodes, peerOf(n, ids[n.ID], names))
			out.Connections = append(out.Connections, fabricui.NetworkConnection{Source: controllerID, Target: ids[n.ID], Kind: "session"})
		case topology.KindThreadUnknown:
			out.Nodes = append(out.Nodes, threadUnknownOf(n))
		case topology.KindWiFiAP:
			out.Nodes = append(out.Nodes, accessPointOf(n))
		}
	}
	for i := range t.Connections {
		out.Connections = append(out.Connections, linkOf(&t.Connections[i], ids))
	}
	return out
}

// graphID is a node's id in the port's graph: peer:<hex node id> for a
// commissioned node, the topology's own id (unknown_…, ap_…) otherwise.
func graphID(n *topology.Node) string {
	if n.Kind == topology.KindMatter {
		return "peer:" + hex16(n.NodeID)
	}
	return n.ID
}

func hex16(v uint64) string { return fmt.Sprintf("%016X", v) }

// peerOf is a commissioned node: its role is the Thread or Wi-Fi role,
// else its network (an Ethernet node is a peer with role ethernet).
func peerOf(n *topology.Node, id string, names map[uint64]string) fabricui.NetworkNode {
	label := names[n.NodeID]
	if label == "" {
		label = "0x" + hex16(n.NodeID)
	}
	state := "ok"
	if !n.Available {
		state = "unreachable"
	}
	role := string(n.Role)
	if role == "" {
		role = string(n.NetworkType)
	}
	var d details
	d.add("network", string(n.NetworkType))
	d.add("role", string(n.Role))
	if n.Rloc16 != nil {
		d.add("rloc16", fmt.Sprintf("0x%04x", *n.Rloc16))
	}
	d.addHex("ext_address", n.ExtAddress)
	d.addHex("ext_pan_id", n.ExtPanID)
	d.add("network_name", n.NetworkName)
	d.add("ssid", n.SSID)
	d.add("bssid", n.BSSID)
	if n.IsBridge {
		d.add("bridge", "yes")
	}
	if len(n.Addresses) > 0 {
		addrs := make([]string, len(n.Addresses))
		for i, a := range n.Addresses {
			addrs[i] = a.String()
		}
		d.add("addresses", strings.Join(addrs, ", "))
	}
	return fabricui.NetworkNode{
		ID: id, Kind: "peer", Label: label, State: state, NodeID: n.NodeID,
		Role: role, RLOC16: n.Rloc16, SSID: n.SSID, BSSID: n.BSSID, Details: d.list(),
	}
}

// threadUnknownOf is a Thread neighbour no commissioned node matches: a
// router when a node saw it rx-on-when-idle (the topology's RoleRouter),
// else an end device.
func threadUnknownOf(n *topology.Node) fabricui.NetworkNode {
	kind := "thread_end_device"
	if n.Role == topology.RoleRouter {
		kind = "thread_router"
	}
	var d details
	d.addHex("ext_address", n.ExtAddress)
	d.addHex("ext_pan_id", n.ExtPanID)
	d.add("role", string(n.Role))
	return fabricui.NetworkNode{ID: n.ID, Kind: kind, Label: hex16(n.ExtAddress), State: "unknown", Details: d.list()}
}

// accessPointOf is a Wi-Fi access point, labelled by the SSID its
// stations name, else by its BSSID.
func accessPointOf(n *topology.Node) fabricui.NetworkNode {
	label := n.SSID
	if label == "" {
		label = n.BSSID
	}
	var d details
	d.add("bssid", n.BSSID)
	d.add("ssid", n.SSID)
	return fabricui.NetworkNode{ID: n.ID, Kind: "wifi_ap", Label: label, State: "unknown", SSID: n.SSID, BSSID: n.BSSID, Details: d.list()}
}

// linkOf is one connection as a link edge: its strength, and LQI and
// RSSI of the direction that strength comes from (the strongest, as the
// topology folds it). The label shows both directions when they report
// and differ: LQI on Thread, RSSI on Wi-Fi.
func linkOf(c *topology.Connection, ids map[string]string) fabricui.NetworkConnection {
	e := fabricui.NetworkConnection{Source: mapID(ids, c.Source), Target: mapID(ids, c.Target), Kind: "link", Strength: string(c.Strength)}
	if c.Strength == "" {
		e.Strength = string(topology.StrengthUnknown)
	}
	thread := c.Network == topology.NetworkThread
	d := c.SourceToTarget
	if d == nil || (c.TargetToSource != nil && d.Strength != c.Strength && c.TargetToSource.Strength == c.Strength) {
		d = c.TargetToSource
	}
	if d != nil {
		if thread {
			lqi := d.LQI
			e.LQI = &lqi
		}
		if d.RSSI != nil {
			rssi := *d.RSSI
			e.RSSI = &rssi
		}
	}
	a, b := c.SourceToTarget, c.TargetToSource
	switch {
	case a == nil || b == nil:
	case thread && a.LQI != b.LQI:
		e.Label = fmt.Sprintf("LQI %d/%d", a.LQI, b.LQI)
	case !thread && a.RSSI != nil && b.RSSI != nil && *a.RSSI != *b.RSSI:
		e.Label = fmt.Sprintf("RSSI %d/%d dBm", *a.RSSI, *b.RSSI)
	}
	return e
}

func mapID(ids map[string]string, id string) string {
	if m, ok := ids[id]; ok {
		return m
	}
	return id
}

// details collects a node's details in order, each only when set.
type details []fabricui.TopologyDetail

func (d *details) add(key, value string) {
	if value != "" {
		*d = append(*d, fabricui.TopologyDetail{Key: key, Value: value})
	}
}

func (d *details) addHex(key string, v uint64) {
	if v != 0 {
		d.add(key, hex16(v))
	}
}

func (d details) list() []fabricui.TopologyDetail {
	if d == nil {
		return []fabricui.TopologyDetail{}
	}
	return d
}
