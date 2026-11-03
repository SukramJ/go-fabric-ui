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

// GET /controller/topology is planned: the in-memory controller builds no
// graph and the route answers 501.
func TestControllerTopology_NotImplemented(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	runFixture(t, mux, "controller-topology.501")
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

// topologySource is a controller whose source builds a graph.
type topologySource struct {
	*MemoryController
	got  bool
	topo NetworkTopology
}

func (s *topologySource) NetworkTopology(_ context.Context, refresh bool) (NetworkTopology, error) {
	s.got = refresh
	return s.topo, nil
}

// A source's graph reaches the page as the contract's TopologyGraph: the
// node id as hex with a ref to the node's page, the link's quality.
func TestControllerTopology_MapsTheSourcesGraph(t *testing.T) {
	rloc, lqi, rssi := uint16(0x2C01), uint8(3), int8(-61)
	src := &topologySource{MemoryController: NewMemoryController(), topo: NetworkTopology{
		CollectedAt: testNow,
		Nodes: []NetworkNode{
			{ID: "node", Kind: "node", Label: "Controller", State: "ok"},
			{ID: "peer:0000000000000011", Kind: "thread_router", Label: "Light", State: "ok", NodeID: 0x11, Role: "router", RLOC16: &rloc, Details: []TopologyDetail{{"channel", "15"}}},
		},
		Connections: []NetworkConnection{{Source: "node", Target: "peer:0000000000000011", Kind: "link", Strength: "strong", LQI: &lqi, RSSI: &rssi}},
	}}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
	rec := do(t, mux, http.MethodGet, "/controller/topology?refresh=true", "")
	if rec.Code != http.StatusOK || !src.got {
		t.Fatalf("got %d %s (refresh %v)", rec.Code, rec.Body, src.got)
	}
	want := `{"as_of":"2026-10-25T12:00:00Z","edges":[{"kind":"link","lqi":3,"rssi":-61,"source":"node","strength":"strong","target":"peer:0000000000000011"}],` +
		`"family":"controller","nodes":[{"details":[],"id":"node","kind":"node","label":"Controller","state":"ok"},` +
		`{"details":[{"key":"channel","value":"15"}],"id":"peer:0000000000000011","kind":"thread_router","label":"Light","node_id_hex":"0000000000000011",` +
		`"ref":{"id":"0000000000000011","page":"controller_node"},"rloc16":11265,"role":"router","state":"ok"}]}`
	if got := string(normaliseJSON(t, rec.Body.Bytes())); got != want {
		t.Fatalf("graph\n got: %s\nwant: %s", got, want)
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
