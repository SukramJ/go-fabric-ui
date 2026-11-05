// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/SukramJ/go-fabric/node"
)

// secondNode is the second node of the two-node tests: one fabric, no
// health probe, so a route's answer tells which node served it.
func secondNode() *fakeNode {
	n := sampleNode()
	n.fabricList = n.fabricList[:1]
	n.healthOK = false
	n.addr = "[::]:5541"
	return n
}

func TestNodes_Fixtures(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{
		second: secondNode(), exposure: newMemoryProvider(),
		exposureReason: "This host builds its endpoints without a selection.",
	})
	for _, name := range []string{"nodes.200", "nodes-status.200", "nodes-status.404"} {
		runFixture(t, mux, name)
	}
}

// Every route of the node family answers under /nodes/{node} for the node
// the key names, and under /node for the default.
func TestNodes_RoutesServeTheKeyedNode(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{second: secondNode()})
	count := func(path string) int {
		t.Helper()
		rec := do(t, mux, http.MethodGet, path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body)
		}
		var body struct {
			Fabrics []json.RawMessage `json:"fabrics"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return len(body.Fabrics)
	}
	if a, b, c := count("/node/fabrics"), count("/nodes/default/fabrics"), count("/nodes/second/fabrics"); a != 2 || b != 2 || c != 1 {
		t.Fatalf("fabrics: default %d, keyed default %d, second %d", a, b, c)
	}
	// Every route of the table is mounted under both prefixes.
	for _, rt := range nodeRoutes {
		if rt.path == "/stream" {
			continue
		}
		path := strings.NewReplacer("{index}", "2", "{ep}", "2").Replace(rt.path)
		rec := do(t, mux, rt.method, "/nodes/nowhere"+path, "{}")
		if rec.Code != http.StatusNotFound || problemOf(t, rec).Title != "No such node" {
			t.Fatalf("%s /nodes/nowhere%s: %d %s", rt.method, path, rec.Code, rec.Body)
		}
	}
}

// The selection, the reassembly and the host's events belong to the
// default node; another node serves none and says why.
func TestNodes_OnlyTheDefaultServesTheSelection(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{second: secondNode(), exposure: newMemoryProvider(), reassemble: func(context.Context) error { return nil }})
	if rec := do(t, mux, http.MethodGet, "/node/exposure", ""); rec.Code != http.StatusOK {
		t.Fatalf("default: %d", rec.Code)
	}
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/nodes/second/exposure"},
		{http.MethodPost, "/nodes/second/maintenance/reassemble"},
	} {
		if rec := do(t, mux, c.method, c.path, ""); rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
}

// A node's admin routes take the token GET /node/status issues.
func TestNodes_AdminRoutesShareTheToken(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{second: secondNode()})
	if rec := do(t, mux, http.MethodPost, "/nodes/second/pairing/window/close", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("with the token: %d %s", rec.Code, rec.Body)
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, DefaultPrefix+"/nodes/second/pairing/window/close", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("without the token: %d", rec.Code)
	}
}

// Each node's stream carries its own node's transitions alone.
func TestNodes_StreamPerNode(t *testing.T) {
	first, second := sampleNode(), secondNode()
	h, _ := newTestHandler(t, testOptions{node: first, second: second})
	r := record(h.keyed["second"])
	first.sample(node.HealthSample{Healthy: false, Note: "first"})
	second.sample(node.HealthSample{Healthy: true, Note: "second"})
	if got := r.types(); !slices.Equal(got, []string{EventHealthSample}) {
		t.Fatalf("second's stream: %v", got)
	}
	if ev := r.events[0].Data.(HealthSampleEvent); ev.Note != "second" {
		t.Fatalf("payload %+v", ev)
	}
}

// stubSource is a NodeSource whose methods are never called: the tests
// that list nodes only.
type stubSource struct{ NodeSource }

func TestNodeEntries(t *testing.T) {
	for _, c := range []struct {
		name string
		o    Options
		err  string
	}{
		{"bad key", Options{Nodes: []NodeEntry{{Key: "a/b"}}}, "does not match"},
		{"empty key", Options{Nodes: []NodeEntry{{Key: ""}}}, "does not match"},
		{"page name", Options{Nodes: []NodeEntry{{Key: "status"}}}, "page name"},
		{"twice", Options{Nodes: []NodeEntry{{Key: "a"}, {Key: "a"}}}, "twice"},
		{"default twice", Options{Node: stubSource{}, Nodes: []NodeEntry{{Key: DefaultNodeKey}}}, "twice"},
	} {
		if _, err := newHandler(c.o); err == nil || !strings.Contains(err.Error(), c.err) {
			t.Errorf("%s: %v", c.name, err)
		}
	}
	// Node first under the default key, then Nodes in their order; a
	// node without a source is listed as not enabled.
	h, err := newHandler(Options{Node: stubSource{}, Nodes: []NodeEntry{{Key: "b", Label: "B", Kind: "device"}, {Key: "a"}}})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.register(mux)
	rec := do(t, mux, http.MethodGet, "/nodes", "")
	if want := `{"nodes":[{"key":"default","label":"default","enabled":true,"managed":false,"exposure":false},{"key":"b","label":"B","kind":"device","enabled":false,"managed":false,"exposure":false},{"key":"a","label":"a","enabled":false,"managed":false,"exposure":false}]}`; strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("got %s", rec.Body)
	}
	// Without Node the first of Nodes is the default.
	h, err = newHandler(Options{Nodes: []NodeEntry{{Key: "x", Source: stubSource{}}}})
	if err != nil {
		t.Fatal(err)
	}
	if h.node == nil {
		t.Fatal("the first of Nodes is not the default")
	}
	// Without any node the list is empty, not null.
	_, mux = newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodGet, "/nodes", ""); strings.TrimSpace(rec.Body.String()) != `{"nodes":[]}` {
		t.Fatalf("no node: %s", rec.Body)
	}
}

// The node family's and the operation routes' admin flags are the
// contract's x-admin, under /node, /nodes/{node} and /operation.
func TestNodeAndOperationAdminMatchesTheContract(t *testing.T) {
	want := contractAdminRoutes(t, func(path string) bool {
		return strings.HasPrefix(path, "/node/") || path == "/nodes" || strings.HasPrefix(path, "/nodes/") || strings.HasPrefix(path, "/operation/")
	})
	var got []string
	for _, rt := range nodeRoutes {
		if rt.admin {
			got = append(got, rt.method+" /node"+rt.path, rt.method+" /nodes/{node}"+rt.path)
		}
	}
	for _, rt := range registryRoutes {
		got = append(got, rt.method+" /nodes"+rt.path)
	}
	for r := range operationAdmin {
		got = append(got, r)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("admin routes differ from the contract's x-admin\n got: %v\nwant: %v", got, want)
	}
}

// health_sample and diagnostic_event reach the stream from the node's
// callbacks, and the stream's end unregisters them.
func TestStream_HealthSampleAndDiagnosticEvent(t *testing.T) {
	n := sampleNode()
	h, _ := newTestHandler(t, testOptions{node: n})
	r := &recorder{}
	unsub := h.subscribe(r.add)
	n.sample(node.HealthSample{Healthy: false, Note: "listener closed"})
	n.trace(node.DiagnosticEvent{At: testNow, Kind: "session", Severity: "info", Message: "A controller opened a session.", Detail: map[string]string{"peer": "1"}})
	if got := r.types(); !slices.Equal(got, []string{EventHealthSample, EventDiagnosticEvent}) {
		t.Fatalf("types %v", got)
	}
	hs, _ := json.Marshal(r.events[0].Data)
	if want := `{"healthy":false,"note":"listener closed","sampled_at":"2026-10-25T12:00:00Z"}`; string(hs) != want {
		t.Fatalf("health_sample %s", hs)
	}
	de, _ := json.Marshal(r.events[1].Data)
	if want := `{"at":"2026-10-25T12:00:00Z","kind":"session","severity":"info","message":"A controller opened a session.","detail":{"peer":"1"}}`; string(de) != want {
		t.Fatalf("diagnostic_event %s", de)
	}
	unsub()
	if c := n.callbackCount(); c != 0 {
		t.Fatalf("%d callbacks left after the stream closed", c)
	}
}
