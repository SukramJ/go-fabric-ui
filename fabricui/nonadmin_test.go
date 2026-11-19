// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

// contractNonAdminEvents is the x-non-admin-events list of the node
// stream's operations, which both carry the same.
func contractNonAdminEvents(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`(?m)^\s+x-non-admin-events: \[([^\]]*)\]`)
	ms := re.FindAllStringSubmatch(string(b), -1)
	if len(ms) != 2 {
		t.Fatalf("found x-non-admin-events %d times, want 2 (stream, streamForNode)", len(ms))
	}
	var out []string
	for s := range strings.SplitSeq(ms[0][1], ",") {
		out = append(out, strings.TrimSpace(s))
	}
	if ms[1][1] != ms[0][1] {
		t.Fatalf("the two stream operations differ: %q, %q", ms[0][1], ms[1][1])
	}
	slices.Sort(out)
	return out
}

// nodeEventConstants are the values of the Event* constants of events.go
// and registry.go: every type the node stream carries.
func nodeEventConstants(t *testing.T) []string {
	t.Helper()
	var out []string
	fs := token.NewFileSet()
	for _, file := range []string{"events.go", "registry.go"} {
		f, err := parser.ParseFile(fs, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			g, ok := d.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST {
				continue
			}
			for _, s := range g.Specs {
				vs := s.(*ast.ValueSpec)
				for i, n := range vs.Names {
					if !strings.HasPrefix(n.Name, "Event") || i >= len(vs.Values) {
						continue
					}
					if lit, ok := vs.Values[i].(*ast.BasicLit); ok {
						out = append(out, strings.Trim(lit.Value, `"`))
					}
				}
			}
		}
	}
	slices.Sort(out)
	return out
}

func TestNodeStreamAllowListMatchesTheContract(t *testing.T) {
	consts := nodeEventConstants(t)
	if len(consts) < 15 {
		t.Fatalf("found %d event constants; the scan is not reading the files", len(consts))
	}
	keys := slices.Sorted(func(yield func(string) bool) {
		for k := range nodeStreamEvents {
			if !yield(k) {
				return
			}
		}
	})
	if !slices.Equal(keys, consts) {
		t.Fatalf("nodeStreamEvents keys differ from the Event constants\n got: %v\nwant: %v", keys, consts)
	}
	var allowed []string
	for k, v := range nodeStreamEvents {
		if v {
			allowed = append(allowed, k)
		}
	}
	slices.Sort(allowed)
	if want := contractNonAdminEvents(t); !slices.Equal(allowed, want) {
		t.Fatalf("non-admin events %v, contract %v", allowed, want)
	}
}

// A request without admin rights receives the allowed types of the
// handler's own hub, nothing of the host's source — not even a type the
// list allows — and an admin request receives all.
func TestNodeStream_NonAdminFilter(t *testing.T) {
	host := NewEvents()
	h, _ := newTestHandler(t, testOptions{events: host})
	var plain, full recorder
	defer h.subscribeAs(false, plain.add)()
	defer h.subscribeAs(true, full.add)()
	h.publish(Event{Type: EventWindowClosed, Data: WindowEvent{}})
	h.publish(Event{Type: EventFabricAdded, Data: FabricEvent{FabricIndex: 1}})
	host.Publish(Event{Type: EventWindowOpened, Data: WindowEvent{}})
	host.Publish(Event{Type: EventExposureChanged, Data: ExposureEvent{}})
	if got := plain.types(); !slices.Equal(got, []string{EventWindowClosed}) {
		t.Fatalf("non-admin received %v", got)
	}
	if got := full.types(); len(got) != 4 {
		t.Fatalf("admin received %v", got)
	}
}

// The node stream route itself makes a request without admin rights a
// filtered stream: the decision that selects subscribeAs's filter, not
// only the filter (UR-P4-test-320). In token mode without the token and
// in host mode with an admin check that refuses, the stream receives
// window_closed and not fabric_added.
func TestNodeStream_RouteFiltersARequestWithoutAdminRights(t *testing.T) {
	for _, c := range []struct {
		name string
		o    testOptions
		hdr  map[string]string
	}{
		{"token mode, no token", testOptions{}, nil},
		{"host mode, admin refused", testOptions{admin: func(*http.Request) bool { return false }}, map[string]string{RequestHeader: "1"}},
	} {
		h, mux := newTestHandler(t, c.o)
		h.heartbeat = time.Hour
		srv := httptest.NewServer(mux)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/node/stream", http.NoBody)
		for k, v := range c.hdr {
			req.Header.Set(k, v)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d", c.name, res.StatusCode)
		}
		waitFor(t, func() bool { return subsOf(h.hub) == 1 })
		h.publish(Event{Type: EventFabricAdded, Data: FabricEvent{FabricIndex: 1}})
		h.publish(Event{Type: EventWindowClosed, Data: WindowEvent{}})
		var types []string
		sc := bufio.NewScanner(res.Body)
		for sc.Scan() {
			if v, ok := strings.CutPrefix(sc.Text(), "event: "); ok {
				types = append(types, v)
				if v == EventWindowClosed {
					break
				}
			}
		}
		res.Body.Close()
		cancel()
		srv.Close()
		if !slices.Equal(types, []string{EventWindowClosed}) {
			t.Errorf("%s: the stream received %v, want only window_closed", c.name, types)
		}
	}
}

// The handler's own events no longer reach the host's *Events.
func TestNodeStream_HostEventsAreNotTheHandlersHub(t *testing.T) {
	host := NewEvents()
	h, _ := newTestHandler(t, testOptions{events: host})
	var seen recorder
	defer host.Subscribe(seen.add)()
	h.publish(Event{Type: EventWindowClosed, Data: WindowEvent{}})
	if got := seen.types(); len(got) != 0 {
		t.Fatalf("the host's source received %v", got)
	}
}

// Without admin rights the node status carries no store and every
// warning's code alone; with them both as before.
func TestStatus_NonAdminIsReduced(t *testing.T) {
	warn := func(context.Context) []Warning {
		return []Warning{{Code: "label_unresolved", Message: "secret detail", Since: time.Unix(1, 0)}}
	}
	cfg := &NodeConfigView{Listen: ":5540", Store: "/data/node.json"}
	_, mux := newTestHandler(t, testOptions{warnings: warn, nodeConfig: cfg})
	get := func(token string) map[string]json.RawMessage {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, DefaultPrefix+"/node/status", http.NoBody)
		req.Host = "example.com"
		if token != "" {
			req.Header.Set(TokenHeader, token)
		}
		rec := serve(mux, req)
		var m map[string]json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatalf("%d %s", rec.Code, rec.Body)
		}
		return m
	}
	plain := get("")
	if w := string(plain["warnings"]); w != `[{"code":"label_unresolved","message":""}]` {
		t.Fatalf("non-admin warnings %s", w)
	}
	if strings.Contains(string(plain["config"]), "store") {
		t.Fatalf("non-admin config %s", plain["config"])
	}
	full := get(testToken)
	if !strings.Contains(string(full["warnings"]), "secret detail") || !strings.Contains(string(full["config"]), `"store":"/data/node.json"`) {
		t.Fatalf("admin status %s %s", full["warnings"], full["config"])
	}
}

// Without admin rights the controller status keeps enabled, capabilities
// and the auth members; every other member is empty or 0, each checked,
// not a sample (decision 57, UR-P4-test-321).
func TestControllerStatus_NonAdminIsZeroed(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), admin: func(*http.Request) bool { return false }})
	rec := do(t, mux, http.MethodGet, "/controller/status", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var st map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &st)
	if st["fabric_id_hex"] != "" || st["node_count"] != float64(0) || st["label"] != "" || st["admin"] != false || st["enabled"] != true || st["capabilities"] == nil {
		t.Fatalf("non-admin controller status %v", st)
	}
	for k, v := range st {
		switch k {
		case "enabled", "capabilities", "admin", "csrf_token":
			continue
		}
		if v != "" && v != float64(0) && v != nil && v != false {
			t.Errorf("non-admin controller status carries %s = %v", k, v)
		}
	}
}

// serve runs one request through mux.
func serve(mux http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}
