// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// A request under a name of Options.AllowedHosts with an Origin of that
// name is an admin's on every node, not only the default one: sibling
// carries allowedHosts and allowHost (UR-P4-auth-01).
func TestSibling_StatusAdminUnderAListedOrigin(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return true }, second: secondNode()})
	hdr := map[string]string{"Host": "example.com", "Origin": "http://example.com", RequestHeader: "1"}
	for _, p := range []string{"/node/status", "/nodes/default/status", "/nodes/second/status"} {
		rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+p, "", hdr)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"admin":true`) {
			t.Errorf("GET %s: %d, want admin true: %s", p, rec.Code, rec.Body)
		}
	}
}

// The same on the node stream: the second node's stream under a listed
// Origin is an admin's, unfiltered (UR-P4-auth-01).
func TestSibling_StreamUnfilteredUnderAListedOrigin(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return true }, second: secondNode()})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	for _, tc := range []struct {
		path string
		hub  *handler
	}{{"/node/stream", h}, {"/nodes/second/stream", h.keyed["second"]}} {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+tc.path, http.NoBody)
		req.Host = "example.com"
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set(RequestHeader, "1")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		br := bufio.NewReader(res.Body)
		if l, _ := br.ReadString('\n'); !strings.HasPrefix(l, ": connected") {
			t.Fatalf("%s: first line %q", tc.path, l)
		}
		tc.hub.publish(Event{Type: EventExposureChanged, Data: ExposureEvent{Keys: []string{}}})
		tc.hub.publish(Event{Type: EventWindowClosed, Data: WindowEvent{}})
		var got []string
		for {
			l, err := br.ReadString('\n')
			if err != nil {
				break
			}
			if strings.HasPrefix(l, "event: ") {
				got = append(got, strings.TrimSpace(strings.TrimPrefix(l, "event: ")))
				if got[len(got)-1] == EventWindowClosed {
					break
				}
			}
		}
		cancel()
		_ = res.Body.Close()
		if len(got) != 2 || got[0] != EventExposureChanged {
			t.Errorf("%s: an admin's stream got %v, want [exposure_changed window_closed]", tc.path, got)
		}
	}
}

// sibling copies every field of the default node's handler but those
// that are the node's own or the default node's alone: a field added to
// handler fails here until sibling carries it or the list names it.
func TestSibling_CopiesEverySharedField(t *testing.T) {
	h, _ := newTestHandler(t, testOptions{
		admin: func(*http.Request) bool { return true }, second: secondNode(),
		controller: NewMemoryController(), operations: sampleOperations(),
		allowHost: func(*http.Request) bool { return false },
	})
	own := map[string]bool{
		// the node's own
		"node": true, "exposure": true, "hooks": true, "hub": true, "extra": true, "config": true,
		"done": true, "stop": true, "mu": true, "ephemeral": true, "openMu": true,
		"logMu": true, "fabricsErr": true, "skipped": true,
		// the default handler's alone: the controller family, the node
		// list and the registry's cache
		"ctrl": true, "nodes": true, "served": true, "keyed": true, "regMu": true, "regCache": true, "regGen": true,
	}
	parent := reflect.ValueOf(h).Elem()
	child := reflect.ValueOf(h.keyed["second"]).Elem()
	for i := range parent.NumField() {
		name := parent.Type().Field(i).Name
		if own[name] || parent.Field(i).IsZero() {
			continue
		}
		if child.Field(i).IsZero() {
			t.Errorf("sibling drops %s", name)
		}
	}
}

// A node the registry created is served by a sibling too: under a listed
// Origin it answers an admin as the default node does (UR-P4-auth-01).
func TestSibling_RegistryNodeAdminUnderAListedOrigin(t *testing.T) {
	reg := newMemoryRegistry(0)
	if _, err := reg.Create(context.Background(), NodeSpec{Label: "Kitchen"}); err != nil {
		t.Fatal(err)
	}
	_, mux := newTestHandler(t, testOptions{
		exposure: newMemoryProvider(), registry: reg, nodeConfig: defaultNodeConfig(),
		admin: func(*http.Request) bool { return true },
	})
	hdr := map[string]string{"Host": "example.com", "Origin": "http://example.com", RequestHeader: "1"}
	rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/nodes/kitchen/status", "", hdr)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"admin":true`) {
		t.Errorf("GET /nodes/kitchen/status: %d, want admin true: %s", rec.Code, rec.Body)
	}
}
