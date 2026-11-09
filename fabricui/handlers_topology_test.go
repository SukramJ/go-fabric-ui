// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// topologyNode3 is the sample node with a third session: a controller on
// the second fabric, quiet for longer than the idle interval it reported.
func topologyNode3() *fakeNode {
	n := sampleNode()
	n.sess.Sessions = append(n.sess.Sessions, node.SessionInfo{
		ID: 19, FabricIndex: 2, PeerNodeID: 0x0000000000000077, LocalNodeID: 0x0000000000000042,
		LastActivity: testNow.Add(-90 * time.Second), LastPeerActivity: testNow.Add(-90 * time.Second),
		Idle: 90 * time.Second, PeerIdle: 90 * time.Second, PeerIdleInterval: 30 * time.Second,
	})
	return n
}

// The node family's graph on the fixture: a bridge with two devices (one
// with a part), two fabrics and three sessions, one of them PASE.
func TestTopology_GraphOfBridgeFabricsAndSessions(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{node: topologyNode3()})
	rec := runFixture(t, mux, "node-topology.200")
	var g topologyGraph
	if err := json.Unmarshal(rec.Body.Bytes(), &g); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		ids = append(ids, n.ID)
	}
	want := []string{"node", "ep:1", "ep:2", "ep:3", "ep:4", "fabric:1", "fabric:2", "session:17", "session:19", "session:18"}
	if !slices.Equal(ids, want) {
		t.Fatalf("nodes %v, want %v", ids, want)
	}
	has := func(src, dst, kind string) bool {
		return slices.Contains(g.Edges, topologyEdge{Source: src, Target: dst, Kind: kind})
	}
	for _, e := range []topologyEdge{
		{"node", "ep:1", "contains"},
		{"ep:1", "ep:2", "contains"},
		{"ep:3", "ep:4", "contains"},
		{"fabric:1", "session:17", "contains"},
		{"session:17", "node", "session"},
		{"fabric:2", "session:19", "contains"},
		{"session:18", "node", "session"},
	} {
		if !has(e.Source, e.Target, e.Kind) {
			t.Errorf("no edge %+v", e)
		}
	}
	for _, e := range g.Edges {
		if e.Target == "session:18" {
			t.Errorf("the PASE session hangs on a fabric: %+v", e)
		}
	}
	state := map[string]string{}
	for _, n := range g.Nodes {
		state[n.ID] = n.State
		if n.Ref == nil {
			t.Errorf("%s has no ref", n.ID)
		}
	}
	if state["ep:3"] != "unreachable" || state["session:19"] != "idle" || state["session:17"] != "ok" {
		t.Fatalf("states %v", state)
	}
}

func TestTopology_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-topology.503")
}

// Before the node assembled its endpoints there is no bridge to draw.
func TestTopology_NotAssembled_Returns503(t *testing.T) {
	n := sampleNode()
	n.topo = node.BridgeTopology{}
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := do(t, mux, http.MethodGet, "/node/topology", "")
	if rec.Code != http.StatusServiceUnavailable || problemOf(t, rec).Code != codeServiceUnready {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestTopology_FabricsFail_Returns500(t *testing.T) {
	n := sampleNode()
	n.fabricsErr = errors.New("store: /var/lib/x unreadable")
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := do(t, mux, http.MethodGet, "/node/topology", "")
	if p := problemOf(t, rec); rec.Code != http.StatusInternalServerError || p.Detail != "" {
		t.Fatalf("got %d %+v", rec.Code, p)
	}
}

// A CASE session on a fabric the node does not list hangs on the node;
// the node's state follows its health sample; a session whose peer
// reported no idle interval is never idle.
func TestBuildNodeGraph_EdgeCases(t *testing.T) {
	topo := node.BridgeTopology{Aggregator: 1}
	diag := node.SessionDiagnostics{AsOf: testNow, Sessions: []node.SessionInfo{
		{ID: 5, FabricIndex: 9, PeerNodeID: 1, PeerIdle: time.Hour},
	}}
	g := buildNodeGraph(topo, nil, diag, "", false, false)
	if g.Nodes[0].State != "unknown" || len(g.Nodes) != 3 || g.Nodes[2].State != "ok" {
		t.Fatalf("nodes %+v", g.Nodes)
	}
	if want := []topologyEdge{{"node", "ep:1", "contains"}, {"session:5", "node", "session"}}; !reflect.DeepEqual(g.Edges, want) {
		t.Fatalf("edges %+v, want %+v", g.Edges, want)
	}
	if g := buildNodeGraph(topo, nil, node.SessionDiagnostics{}, "", true, false); g.Nodes[0].State != "unreachable" {
		t.Fatalf("an unhealthy node is %q", g.Nodes[0].State)
	}
	f := fabricNode(&node.FabricInfo{FabricIndex: 3, VendorID: 0xFFF1})
	if f.Label != "0xFFF1" {
		t.Fatalf("fabric label %q", f.Label)
	}
}

// The three session observers reach the stream with the session as
// GET /node/sessions lists it.
func TestStream_SessionEvents(t *testing.T) {
	n := sampleNode()
	h, _ := newTestHandler(t, testOptions{node: n})
	rec := &recorder{}
	defer h.subscribe(rec.add)()
	s := n.sess.Sessions[0]
	n.notify(func(o node.Observers) { o.OnSessionOpened(s) })
	s.Subscriptions = 3
	n.notify(func(o node.Observers) { o.OnSubscriptionsChanged(s) })
	n.notify(func(o node.Observers) { o.OnSessionClosed(s) })
	if got, want := rec.types(), []string{EventSessionOpened, EventSubscriptionsChanged, EventSessionClosed}; !slices.Equal(got, want) {
		t.Fatalf("events %v, want %v", got, want)
	}
	ev, ok := rec.events[1].Data.(SessionEvent)
	if !ok || ev.SessionID != 17 || ev.Subscriptions != 3 || ev.PeerNodeIDHex != "000000000001B669" || ev.PeerIdleSeconds != 45 {
		t.Fatalf("payload %+v", rec.events[1].Data)
	}
	b, _ := json.Marshal(ev)
	var asSession sessionResponse
	if err := json.Unmarshal(b, &asSession); err != nil || SessionEvent(asSession) != ev {
		t.Fatalf("the event's JSON is not the session's: %s", b)
	}
}

// controllerTopology is the controller-topology.200 fixture's graph as
// the port carries it: the controller, a Thread router, a sleepy end
// device, an unreachable Wi-Fi station with its access point, an
// Ethernet bridge, a Thread router and an end device no commissioned node
// matches, a border router (contract 0.8.0), and a link of every strength. It is what the go-fabric
// adapter maps its topology test literal to
// (fabricui/gofabric/topology_test.go), so the fixture holds both ends.
func controllerTopology() NetworkTopology {
	rloc := func(v uint16) *uint16 { return &v }
	lqi := func(v uint8) *uint8 { return &v }
	rssi := func(v int8) *int8 { return &v }
	const pan = "DEAD00BEEF00CAFE"
	return NetworkTopology{
		CollectedAt: testNow,
		Nodes: []NetworkNode{
			{ID: "node", Kind: "node", Label: "Controller", State: "ok", Details: []TopologyDetail{{"nodes", "4"}, {"collected_at", "2026-10-25T12:00:00Z"}}},
			{
				ID: "peer:0000000000000011", Kind: "peer", Label: "Light", State: "ok", NodeID: 0x11, Role: "router", RLOC16: rloc(0x2C00),
				Details: []TopologyDetail{
					{"network", "thread"},
					{"role", "router"},
					{"rloc16", "0x2c00"},
					{"ext_address", "1A2B3C4D5E6F7011"},
					{"ext_pan_id", pan},
					{"network_name", "Home"},
					{"addresses", "fd00::11"},
				},
			},
			{
				ID: "peer:0000000000000012", Kind: "peer", Label: "Sensor", State: "ok", NodeID: 0x12, Role: "sleepy_end_device", RLOC16: rloc(0x2C01),
				Details: []TopologyDetail{
					{"network", "thread"},
					{"role", "sleepy_end_device"},
					{"rloc16", "0x2c01"},
					{"ext_address", "1A2B3C4D5E6F7012"},
					{"ext_pan_id", pan},
					{"network_name", "Home"},
					{"addresses", "fd00::12"},
				},
			},
			{
				ID: "peer:0000000000000013", Kind: "peer", Label: "0x0000000000000013", State: "unreachable", NodeID: 0x13, Role: "station",
				SSID: "home-wifi", BSSID: "AA:BB:CC:DD:EE:FF",
				Details: []TopologyDetail{
					{"network", "wifi"}, {"role", "station"}, {"ssid", "home-wifi"}, {"bssid", "AA:BB:CC:DD:EE:FF"}, {"addresses", "192.168.1.13, fd00::13"},
				},
			},
			{
				ID: "peer:0000000000000014", Kind: "peer", Label: "Bridge", State: "ok", NodeID: 0x14, Role: "ethernet",
				Details: []TopologyDetail{{"network", "ethernet"}, {"bridge", "yes"}, {"addresses", "192.168.1.20"}},
			},
			{
				ID: "unknown_0011223344556677", Kind: "thread_router", Label: "0011223344556677", State: "unknown",
				Details: []TopologyDetail{{"ext_address", "0011223344556677"}, {"ext_pan_id", pan}, {"role", "router"}, {"network_name", "Home"}},
			},
			{
				ID: "unknown_8899AABBCCDDEEFF", Kind: "thread_end_device", Label: "8899AABBCCDDEEFF", State: "unknown",
				Details: []TopologyDetail{{"ext_address", "8899AABBCCDDEEFF"}, {"ext_pan_id", pan}},
			},
			{
				ID: "ap_AABBCCDDEEFF", Kind: "wifi_ap", Label: "home-wifi", State: "unknown", SSID: "home-wifi", BSSID: "AA:BB:CC:DD:EE:FF",
				Details: []TopologyDetail{{"bssid", "AA:BB:CC:DD:EE:FF"}, {"ssid", "home-wifi"}},
			},
			{
				ID: "br_CAFE0000CAFE0001", Kind: "border_router", Label: "otbr-kitchen", State: "ok",
				Details: []TopologyDetail{
					{"host_name", "otbr-kitchen"},
					{"vendor_name", "OpenThread"},
					{"model_name", "OTBR"},
					{"network_name", "Home"},
					{"ext_address", "CAFE0000CAFE0001"},
					{"ext_pan_id", pan},
					{"last_seen", "2026-10-25T11:55:00Z"},
				},
			},
		},
		Connections: []NetworkConnection{
			{Source: "node", Target: "peer:0000000000000011", Kind: "session"},
			{Source: "node", Target: "peer:0000000000000012", Kind: "session"},
			{Source: "node", Target: "peer:0000000000000013", Kind: "session"},
			{Source: "node", Target: "peer:0000000000000014", Kind: "session"},
			{Source: "peer:0000000000000011", Target: "peer:0000000000000012", Kind: "link", Strength: "strong", LQI: lqi(3), RSSI: rssi(-45), Label: "LQI 3/2"},
			{Source: "peer:0000000000000011", Target: "unknown_0011223344556677", Kind: "link", Strength: "medium", LQI: lqi(2), RSSI: rssi(-66), Label: "LQI 2/1"},
			{Source: "peer:0000000000000011", Target: "unknown_8899AABBCCDDEEFF", Kind: "link", Strength: "weak", LQI: lqi(1), RSSI: rssi(-88)},
			{Source: "peer:0000000000000012", Target: "unknown_0011223344556677", Kind: "link", Strength: "none", LQI: lqi(0)},
			{Source: "peer:0000000000000013", Target: "ap_AABBCCDDEEFF", Kind: "link", Strength: "unknown"},
			{Source: "peer:0000000000000011", Target: "br_CAFE0000CAFE0001", Kind: "link", Strength: "strong", LQI: lqi(3), RSSI: rssi(-50)},
		},
	}
}

// GET /controller/topology serves the source's graph (200); a source
// without one answers 501; refresh=true without admin rights 403.
func TestControllerTopology_Fixtures(t *testing.T) {
	_, mux, m := controllerHandler(t, nil)
	runFixture(t, mux, "controller-topology.501")
	runFixture(t, mux, "controller-topology.403")
	topo := controllerTopology()
	m.Topology = &topo
	runFixture(t, mux, "controller-topology.200")
}

// refresh=true needs admin rights; the plain read does not.
func TestControllerTopology_RefreshNeedsAdmin(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController(), admin: func(*http.Request) bool { return false }})
	for path, want := range map[string]int{
		"/controller/topology":               http.StatusNotImplemented,
		"/controller/topology?refresh=false": http.StatusNotImplemented,
		"/controller/topology?refresh=true":  http.StatusForbidden,
		"/controller/topology?refresh=1":     http.StatusForbidden,
	} {
		if rec := do(t, mux, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("%s: %d %s, want %d", path, rec.Code, rec.Body, want)
		}
	}
	_, admin, _ := controllerHandler(t, nil)
	if rec := do(t, admin, http.MethodGet, "/controller/topology?refresh=true", ""); rec.Code != http.StatusNotImplemented {
		t.Fatalf("admin refresh: %d %s", rec.Code, rec.Body)
	}
}

// topologySource is a controller whose source records the refresh flag
// and fails as told.
type topologySource struct {
	*MemoryController
	refreshed []bool
	err       error
}

func (s *topologySource) NetworkTopology(_ context.Context, refresh bool) (NetworkTopology, error) {
	s.refreshed = append(s.refreshed, refresh)
	return controllerTopology(), s.err
}

// The refresh flag reaches the source; a node's failure is 502.
func TestControllerTopology_RefreshAndFailure(t *testing.T) {
	src := &topologySource{MemoryController: NewMemoryController()}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
	for _, path := range []string{"/controller/topology", "/controller/topology?refresh=true"} {
		if rec := do(t, mux, http.MethodGet, path, ""); rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
	}
	if !slices.Equal(src.refreshed, []bool{false, true}) {
		t.Fatalf("refresh %v", src.refreshed)
	}
	src.err = ErrNodeFailed
	if rec := do(t, mux, http.MethodGet, "/controller/topology?refresh=true", ""); rec.Code != http.StatusBadGateway {
		t.Fatalf("failure: %d %s", rec.Code, rec.Body)
	}
}

// The topology_changed event carries an empty payload.
func TestControllerEvent_TopologyChanged(t *testing.T) {
	b, err := json.Marshal(controllerEvent(ControllerEvent{Type: EventTopologyChanged}).Data)
	if err != nil || string(b) != "{}" {
		t.Fatalf("payload %s %v", b, err)
	}
}

// controllerAdminQuery names exactly the query parameters the contract
// marks x-admin under /controller.
func TestControllerAdminQueryMatchesTheContract(t *testing.T) {
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pathRe := regexp.MustCompile(`^ {2}(/\S+):\s*$`)
	methodRe := regexp.MustCompile(`^ {4}(get|put|post|delete|patch):\s*$`)
	paramRe := regexp.MustCompile(`^\s+- \{ name: (\w+), in: query, x-admin: true`)
	got := map[string]string{}
	var path, method string
	for line := range strings.SplitSeq(string(b), "\n") {
		if m := pathRe.FindStringSubmatch(line); m != nil {
			path = m[1]
		} else if m := methodRe.FindStringSubmatch(line); m != nil {
			method = strings.ToUpper(m[1])
		} else if m := paramRe.FindStringSubmatch(line); m != nil && strings.HasPrefix(path, "/controller/") {
			got[method+" "+path] = m[1]
		}
	}
	if !reflect.DeepEqual(got, controllerAdminQuery) {
		t.Fatalf("contract %v, controllerAdminQuery %v", got, controllerAdminQuery)
	}
}
