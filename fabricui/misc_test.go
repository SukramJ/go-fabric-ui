// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
)

// *node.Node is a NodeSource: a facade change that breaks the subset fails
// here, not in a host.
var _ NodeSource = (*node.Node)(nil)

// Ports TestActorFromRequest_NoContext_ReturnsAnonymous: in token mode a
// write without the token is refused before it reaches the node.
func TestAdmin_TokenModeRefusesWithoutTheToken(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-fabric-remove.403")
	if len(n.removed) != 0 {
		t.Fatal("a refused request removed a fabric")
	}
}

// Ports TestActorFromRequest_IdentityWithoutSubject_ReturnsAnonymous: a
// wrong token is no token.
func TestAdmin_TokenModeRefusesAWrongToken(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, DefaultPrefix+"/node/pairing/window/close", nil)
	req.Header.Set(TokenHeader, "not-the-token")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || problemOf(t, rec).Code != codeForbidden {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Ports TestActorFromRequest_WithAuthenticatedIdentity: the host's check
// decides and the token is not consulted.
func TestAdmin_HostCheckDecides(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{admin: func(r *http.Request) bool { return r.Header.Get("X-Host-Admin") == "yes" }})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, DefaultPrefix+"/node/pairing/window/close", nil)
	req.Header.Set(TokenHeader, testToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("token without the host's consent: %d, want 403", rec.Code)
	}
	req.Header.Set("X-Host-Admin", "yes")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("host admin: %d, want 204", rec.Code)
	}
}

// Ports TestPublishMatterEventTypeCarriesTheMatterPrefix: the stream names
// events by the contract's types, delivers the handler's own, the host's
// and the node's (an observer registered while the stream is open), and
// keeps the connection alive with comments.
func TestStream_DeliversEventsAndHeartbeats(t *testing.T) {
	host := &recordingSource{}
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{events: host, node: n})
	h.heartbeat = 20 * time.Millisecond
	base := h.hub.subscribers()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/node/stream", http.NoBody)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	sc := bufio.NewScanner(resp.Body)
	next := func() string {
		t.Helper()
		for sc.Scan() {
			if line := sc.Text(); line != "" {
				return line
			}
		}
		t.Fatalf("stream ended: %v", sc.Err())
		return ""
	}
	if line := next(); line != ": connected" {
		t.Fatalf("first line %q", line)
	}
	waitFor(t, func() bool { return h.hub.subscribers() == base+1 && host.subscribers() == 1 && n.observerCount() == 1 })
	h.publish(Event{Type: EventFabricAdded, Data: FabricEvent{FabricIndex: 3, Label: "Home"}})
	if a, b := next(), next(); a != "event: fabric_added" || b != `data: {"fabric_index":3,"label":"Home"}` {
		t.Fatalf("event lines %q %q", a, b)
	}
	host.publish(Event{Type: EventExposureChanged, Data: ExposureEvent{Keys: []string{"lamp-1"}}})
	if a, b := next(), next(); a != "event: exposure_changed" || b != `data: {"keys":["lamp-1"]}` {
		t.Fatalf("event lines %q %q", a, b)
	}
	n.notify(func(o node.Observers) { _ = o.OnFabricRemoved(node.FabricInfo{FabricIndex: 2}) })
	if a, b := next(), next(); a != "event: fabric_removed" || b != `data: {"fabric_index":2}` {
		t.Fatalf("event lines %q %q", a, b)
	}
	if line := next(); line != ": heartbeat" {
		t.Fatalf("line %q, want a heartbeat", line)
	}
	cancel()
	waitFor(t, func() bool { return h.hub.subscribers() == base && host.subscribers() == 0 && n.observerCount() == 0 })
}

// recordingSource is a host EventSource that is not an *Events.
type recordingSource struct{ ev Events }

func (s *recordingSource) Subscribe(fn func(Event)) func() {
	if s.ev.subs == nil {
		s.ev.subs = map[uint64]func(Event){}
	}
	return s.ev.Subscribe(fn)
}
func (s *recordingSource) publish(ev Event) { s.ev.Publish(ev) }
func (s *recordingSource) subscribers() int { return s.ev.subscribers() }

func (e *Events) subscribers() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.subs)
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not reached")
		}
		time.Sleep(time.Millisecond)
	}
}

// The node's observers report its transitions — a commissioning's
// fabric_added, a removal, a window opened and closed — once each: a
// repeated snapshot of the same window state is no transition, and a route
// that causes one does not publish it a second time. The unsubscribe
// removes the observer from the node.
func TestSubscribe_ObservesTheNodeOnce(t *testing.T) {
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{node: n})
	rec := &recorder{}
	unsub := h.subscribe(rec.add)
	if n.observerCount() != 1 {
		t.Fatalf("%d observers registered, want 1", n.observerCount())
	}
	n.notify(func(o node.Observers) {
		_ = o.OnFabricAdded(node.FabricInfo{FabricIndex: 4, VendorID: 0x1349, Label: "Home"})
	})
	n.notify(func(o node.Observers) { _ = o.OnWindowTransition(node.WindowSnapshot{Open: true, Discriminator: 77}) })
	n.notify(func(o node.Observers) { _ = o.OnWindowTransition(node.WindowSnapshot{Open: true, Discriminator: 77}) })
	do(t, mux, http.MethodPost, "/node/pairing/window/close", "")
	do(t, mux, http.MethodDelete, "/node/fabrics/1", "")
	want := []string{EventFabricAdded, EventWindowOpened, EventWindowClosed, EventFabricRemoved}
	if !slices.Equal(rec.types(), want) {
		t.Fatalf("events %v, want %v", rec.types(), want)
	}
	if fe, ok := rec.events[0].Data.(FabricEvent); !ok || fe.VendorName == "" {
		t.Fatalf("fabric_added payload %+v", rec.events[0].Data)
	}
	if we, ok := rec.events[1].Data.(WindowEvent); !ok || we.Discriminator != 77 {
		t.Fatalf("window_opened payload %+v", rec.events[1].Data)
	}
	unsub()
	if n.observerCount() != 0 {
		t.Fatalf("%d observers left after the unsubscribe", n.observerCount())
	}
}

// NodeObservers is the route a host's own channel takes, and it yields
// what the handler's stream yields: the same transitions, through the same
// node, give the same events with the same payloads, a repeated window
// snapshot included.
func TestNodeObservers_YieldWhatTheStreamYields(t *testing.T) {
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{node: n})
	stream, host := &recorder{}, &recorder{}
	defer h.subscribe(stream.add)()
	defer n.observe(NodeObservers(host.add, n.window().Open))()

	n.notify(func(o node.Observers) {
		_ = o.OnFabricAdded(node.FabricInfo{FabricIndex: 4, VendorID: 0x1349, Label: "Home"})
	})
	n.notify(func(o node.Observers) {
		_ = o.OnWindowTransition(node.WindowSnapshot{Open: true, Discriminator: 77, Timeout: 300 * time.Second})
	})
	n.notify(func(o node.Observers) {
		_ = o.OnWindowTransition(node.WindowSnapshot{Open: true, Discriminator: 77, Timeout: 300 * time.Second})
	})
	do(t, mux, http.MethodPost, "/node/pairing/window/close", "")
	do(t, mux, http.MethodDelete, "/node/fabrics/1", "")

	want := []Event{
		{Type: EventFabricAdded, Data: FabricEvent{FabricIndex: 4, VendorID: 0x1349, VendorName: eligibility.VendorName(0x1349), Label: "Home"}},
		{Type: EventWindowOpened, Data: WindowEvent{Discriminator: 77, DurationSeconds: 300}},
		{Type: EventWindowClosed, Data: WindowEvent{}},
		{Type: EventFabricRemoved, Data: FabricEvent{FabricIndex: 1}},
	}
	if !slices.Equal(host.events, want) {
		t.Fatalf("NodeObservers yielded %+v, want %+v", host.events, want)
	}
	if !slices.Equal(stream.events, host.events) {
		t.Fatalf("the stream yielded %+v, NodeObservers %+v", stream.events, host.events)
	}
}

func TestMount_RegistersTheRoutesAndRefusesABadPrefix(t *testing.T) {
	mux := http.NewServeMux()
	if err := Mount(mux, Options{Prefix: "/ui-api"}); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ui-api/node/status", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"enabled":false`) {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
	if err := Mount(mux, Options{Prefix: "/ui-api"}); err == nil {
		t.Fatal("a second mount on the same patterns succeeded")
	}
	for _, p := range []string{"api", "/api/", "/a {x}"} {
		if err := Mount(http.NewServeMux(), Options{Prefix: p}); err == nil {
			t.Fatalf("prefix %q accepted", p)
		}
	}
}

func TestSPA_ServesTheBundleWithItsCacheHeaders(t *testing.T) {
	h := spaHandlerFS(fstest.MapFS{
		"index.html":       {Data: []byte("<!doctype html><title>shell</title>")},
		"assets/app-1.js":  {Data: []byte("console.log(1)")},
		"favicon.svg":      {Data: []byte("<svg/>")},
		"assets/dir/x.css": {Data: []byte("a{}")},
	})
	get := func(p string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, p, nil))
		return rec
	}
	for _, p := range []string{"/", "/index.html", "/deep/link", "/assets/"} {
		rec := get(p)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "shell") || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s: %d %q %s", p, rec.Code, rec.Header().Get("Cache-Control"), rec.Body)
		}
	}
	if rec := get("/assets/app-1.js"); rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Fatalf("hashed asset cache %q", rec.Header().Get("Cache-Control"))
	}
	if rec := get("/favicon.svg"); rec.Header().Get("Cache-Control") != "public, max-age=3600, must-revalidate" {
		t.Fatalf("plain file cache %q", rec.Header().Get("Cache-Control"))
	}
}

func TestSPA_NotBuiltAnswers503(t *testing.T) {
	rec := httptest.NewRecorder()
	spaHandlerFS(fstest.MapFS{"README": {Data: []byte("x")}}).ServeHTTP(rec,
		httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), "Bundle not built") {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestHandler_ServesAPIAndBundle(t *testing.T) {
	srv := httptest.NewServer(Handler(Options{}))
	defer srv.Close()
	for _, p := range []string{"/api/v1/node/status", "/"} {
		resp, err := srv.Client().Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if p == "/api/v1/node/status" && resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d", p, resp.StatusCode)
		}
	}
	if _, err := fs.Stat(distFS, "dist"); err != nil {
		t.Fatalf("dist is not embedded: %v", err)
	}
}

// TestEveryFixtureIsUsedByTheGoTests fails when a contract fixture is not
// named by any Go test; the TypeScript suite holds the same check for its
// side (ui/src/lib/api/fixtures.test.ts).
func TestEveryFixtureIsUsedByTheGoTests(t *testing.T) {
	names, err := filepath.Glob(filepath.Join(fixtureDir, "*.json"))
	if err != nil || len(names) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}
	tests, _ := filepath.Glob("*_test.go")
	var src strings.Builder
	for _, f := range tests {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		src.Write(b)
	}
	for _, n := range names {
		base := strings.TrimSuffix(filepath.Base(n), ".json")
		if !regexp.MustCompile(`"` + regexp.QuoteMeta(base) + `"`).MatchString(src.String()) {
			t.Errorf("fixture %s is used by no Go test", base)
		}
	}
}
