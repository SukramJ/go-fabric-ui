// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

// Host mode with an Admin that admits every request: a request without
// the request header, a body that is not JSON and a foreign Origin are
// each refused before the route runs (decision 38).
func TestAdminRoutes_HeaderOriginAndMediaType(t *testing.T) {
	cross := func(ct string) map[string]string {
		h := map[string]string{"Origin": "https://other.example", "Sec-Fetch-Site": "cross-site"}
		if ct != "" {
			h["Content-Type"] = ct
		}
		return h
	}
	admin := func(*http.Request) bool { return true }

	t.Run("factory reset, foreign Origin, text/plain body", func(t *testing.T) {
		n := sampleNode()
		_, mux := newTestHandler(t, testOptions{node: n, admin: admin})
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/maintenance/factory-reset", `{"confirm":"remove-all-fabrics"}`, cross("text/plain;charset=UTF-8"))
		if rec.Code != http.StatusForbidden || len(n.removed) != 0 {
			t.Errorf("→ %d, fabrics removed %v; want 403", rec.Code, n.removed)
		}
	})
	t.Run("erase, foreign Origin, text/plain body", func(t *testing.T) {
		erased := 0
		_, mux := newTestHandler(t, testOptions{admin: admin, erase: func(context.Context) error { erased++; return nil }})
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/maintenance/erase", "{\"confirm\":\"erase-node\",\"x\":\"=\"}\r\n", cross("text/plain"))
		if rec.Code != http.StatusForbidden || erased != 0 {
			t.Errorf("→ %d, Erase called %d times; want 403", rec.Code, erased)
		}
	})
	t.Run("commission, foreign Origin, text/plain body", func(t *testing.T) {
		ctrl := NewMemoryController()
		_, mux := newTestHandler(t, testOptions{admin: admin, controller: ctrl})
		before := ctrl.Nodes()
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/controller/commission", "{\"code\":\"34970112332\",\"x\":\"=\"}\r\n", cross("text/plain"))
		if rec.Code != http.StatusForbidden || len(ctrl.Nodes()) != len(before) {
			t.Errorf("→ %d, controller nodes %d → %d; want 403", rec.Code, len(before), len(ctrl.Nodes()))
		}
	})
	t.Run("bodiless window close", func(t *testing.T) {
		_, mux := newTestHandler(t, testOptions{admin: admin})
		if rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window/close", "", cross("")); rec.Code != http.StatusForbidden {
			t.Errorf("→ %d; want 403", rec.Code)
		}
	})
	t.Run("bodiless updates check", func(t *testing.T) {
		_, mux := newTestHandler(t, testOptions{admin: admin, controller: NewMemoryController()})
		if rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/controller/updates/check", "", cross("")); rec.Code != http.StatusForbidden {
			t.Errorf("→ %d; want 403", rec.Code)
		}
	})
	t.Run("window open, foreign Origin, text/plain body", func(t *testing.T) {
		n := sampleNode()
		_, mux := newTestHandler(t, testOptions{node: n, admin: admin})
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window", `{}`, cross("text/plain;charset=UTF-8"))
		if rec.Code != http.StatusForbidden || n.window().Open {
			t.Errorf("→ %d, open=%v; want 403", rec.Code, n.window().Open)
		}
	})
	for name, c := range map[string]struct{ path, body, ct string }{
		"JSON as text/plain on a strict route": {"/node/pairing/window", "{\"timeout_seconds\":180,\"x\":\"=\"}\r\n", "text/plain"},
		"urlencoded":                           {"/node/maintenance/erase", "confirm=erase-node", "application/x-www-form-urlencoded"},
	} {
		t.Run(name, func(t *testing.T) {
			_, mux := newTestHandler(t, testOptions{admin: admin, erase: func(context.Context) error { return nil }})
			rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+c.path, c.body, cross(c.ct))
			if rec.Code != http.StatusForbidden || rec.Header().Get("X-Problem-Code") != codeOriginNotAllowed {
				t.Errorf("→ %d %s, want 403 origin_not_allowed", rec.Code, rec.Header().Get("X-Problem-Code"))
			}
		})
	}
	// Without an Origin: the request header decides, then the media type.
	t.Run("no header", func(t *testing.T) {
		_, mux := newTestHandler(t, testOptions{admin: admin})
		if rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window/close", "", nil); rec.Code != http.StatusForbidden || rec.Header().Get("X-Problem-Code") != codeForbidden {
			t.Errorf("→ %d; want 403 forbidden", rec.Code)
		}
	})
	for ct, want := range map[string]int{"": http.StatusUnsupportedMediaType, "text/plain": http.StatusUnsupportedMediaType, "APPLICATION/JSON; charset=utf-8": http.StatusOK} {
		_, mux := newTestHandler(t, testOptions{admin: admin})
		hdr := map[string]string{RequestHeader: "1"}
		if ct != "" {
			hdr["Content-Type"] = ct
		}
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window", `{}`, hdr)
		if rec.Code != want {
			t.Errorf("Content-Type %q: %d %s, want %d", ct, rec.Code, rec.Body, want)
		}
	}
}

// The admin GETs with an effect need the request header in host mode:
// without it they are refused.
func TestAdminRoutes_GETWithAnEffectNeedsTheHeader(t *testing.T) {
	ctrl := NewMemoryController()
	topo := controllerTopology()
	ctrl.Topology = &topo
	_, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return true }, controller: ctrl})
	for _, p := range []string{"/controller/topology?refresh=true", "/controller/discovery"} {
		rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+p, "", map[string]string{"Referer": "http://other.example/", "Accept": "image/avif,image/webp,*/*"})
		if rec.Code != http.StatusForbidden {
			t.Errorf("GET %s without the header → %d; want 403", p, rec.Code)
		}
	}
}

// On an open route a foreign Origin makes the request a non-admin's.
func TestOpenRoute_ForeignOriginIsANonAdmin(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return true }})
	rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{RequestHeader: "1", "Origin": "null"})
	if rec.Code != http.StatusOK || !containsJSON(rec.Body.String(), `"admin":false`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{RequestHeader: "1", "Origin": "http://example.com"})
	if !containsJSON(rec.Body.String(), `"admin":true`) {
		t.Fatalf("listed origin: %s", rec.Body)
	}
}

func containsJSON(body, member string) bool { return strings.Contains(body, member) }

// Admitted by the predicate, an Origin equal to the Host passes an admin
// write; another Origin is 403 origin_not_allowed; an Origin under the
// name rules passes as before (decision 61).
func TestAllowHost_OriginUnderAnAdmittedHost(t *testing.T) {
	pred, _ := countingPredicate(true)
	_, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return true }, allowHost: pred})
	for origin, want := range map[string]int{
		"http://unlisted.example:8123": http.StatusNoContent,
		"http://UNLISTED.example:8123": http.StatusNoContent,
		"http://unlisted.example:9999": http.StatusForbidden,
		"http://other.example:8123":    http.StatusForbidden,
		"http://localhost:8123":        http.StatusNoContent,
		"http://192.168.1.2":           http.StatusNoContent,
	} {
		rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window/close", "",
			map[string]string{"Host": "unlisted.example:8123", "Origin": origin, RequestHeader: "1"})
		if rec.Code != want || (want == http.StatusForbidden && rec.Header().Get("X-Problem-Code") != codeOriginNotAllowed) {
			t.Errorf("Origin %s: %d %s, want %d", origin, rec.Code, rec.Body, want)
		}
	}
}
