// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// reqWith sends one request with the given headers; "Host" sets r.Host.
func reqWith(t *testing.T, mux http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequestWithContext(context.Background(), method, path, http.NoBody)
	} else {
		req = httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	}
	for k, v := range hdr {
		if k == "Host" {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// foreignHost is a request under a name the handler does not list, sent
// from a page of that same name.
func foreignHost(extra map[string]string) map[string]string {
	h := map[string]string{"Host": "unlisted.example:8099", "Origin": "http://unlisted.example:8099", "Sec-Fetch-Site": "same-origin"}
	for k, v := range extra {
		h[k] = v
	}
	return h
}

// isHostRefusal reports a 403 host_not_allowed.
func isHostRefusal(rec *httptest.ResponseRecorder) bool {
	return rec.Code == http.StatusForbidden && rec.Header().Get("X-Problem-Code") == codeHostNotAllowed
}

// Token mode under an unlisted Host: the status issues no token, the
// pairing codes and a factory reset are refused (decision 37).
func TestHost_TokenModeAnUnlistedHostGetsNoToken(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n, configured: true})
	rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", foreignHost(nil))
	if !isHostRefusal(rec) || strings.Contains(rec.Body.String(), "csrf_token") {
		t.Fatalf("status: %d %s", rec.Code, rec.Body)
	}
	if rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/pairing", "", foreignHost(map[string]string{TokenHeader: testToken})); !isHostRefusal(rec) {
		t.Fatalf("pairing: %d %s", rec.Code, rec.Body)
	}
	rec = reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/maintenance/factory-reset", `{"confirm":"remove-all-fabrics"}`,
		foreignHost(map[string]string{TokenHeader: testToken, "Content-Type": "application/json"}))
	if !isHostRefusal(rec) || len(n.removed) != 0 {
		t.Fatalf("factory reset: %d, removed %v", rec.Code, n.removed)
	}
}

// Every route that issues the token refuses an unlisted Host.
func TestHost_EveryStatusRouteRefusesAnUnlistedHost(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), second: secondNode()})
	for _, p := range []string{"/node/status", "/nodes/second/status", "/controller/status"} {
		rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+p, "", foreignHost(nil))
		if !isHostRefusal(rec) || strings.Contains(rec.Body.String(), testToken) {
			t.Errorf("GET %s: %d %s", p, rec.Code, rec.Body)
		}
	}
}

// What holds without the Host check: a cross-site admin request without
// the token is 403, a preflight gets an error and no Access-Control-*
// header, the status sends no Access-Control-Allow-Origin.
func TestHost_ControlsThatHold(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	cross := map[string]string{"Origin": "https://other.example", "Sec-Fetch-Site": "cross-site"}
	if rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window/close", "", cross); rec.Code != http.StatusForbidden {
		t.Errorf("cross-site admin route without the token: %d, want 403", rec.Code)
	}
	rec := reqWith(t, mux, http.MethodOptions, DefaultPrefix+"/node/pairing/window/close", "",
		map[string]string{"Origin": "https://other.example", "Access-Control-Request-Method": "POST", "Access-Control-Request-Headers": "x-fabric-ui-token"})
	if rec.Code < 400 {
		t.Errorf("preflight: %d, want an error status", rec.Code)
	}
	for k := range rec.Header() {
		if strings.HasPrefix(k, "Access-Control-") {
			t.Errorf("preflight answers %s", k)
		}
	}
	rec = reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", cross)
	if v := rec.Header().Get("Access-Control-Allow-Origin"); v != "" {
		t.Errorf("status answers Access-Control-Allow-Origin %q", v)
	}
}

// Host mode under an unlisted Host reads nothing, the streams and the
// bundle included.
func TestHost_HostModeAnUnlistedHostReadsNothing(t *testing.T) {
	ctrl := NewMemoryController()
	h, mux := newTestHandler(t, testOptions{admin: func(*http.Request) bool { return false }, controller: ctrl, operations: sampleOperations()})
	mux.Handle("/", h.hostCheck(spaHandler(nil)))
	const id = "0000000000000011"
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, DefaultPrefix + "/node/status", ""},
		{http.MethodGet, DefaultPrefix + "/node/fabrics", ""},
		{http.MethodGet, DefaultPrefix + "/controller/status", ""},
		{http.MethodGet, DefaultPrefix + "/controller/nodes/" + id + "/acl", ""},
		{http.MethodPost, DefaultPrefix + "/controller/nodes/" + id + "/attributes", `{"paths":[{"endpoint":0,"cluster":40,"attribute":5}]}`},
		{http.MethodGet, DefaultPrefix + "/operation/log", ""},
		{http.MethodHead, DefaultPrefix + "/node/stream", ""},
		{http.MethodHead, DefaultPrefix + "/controller/stream", ""},
		{http.MethodHead, DefaultPrefix + "/operation/log/stream", ""},
		{http.MethodGet, "/", ""},
	} {
		hdr := foreignHost(nil)
		if c.body != "" {
			hdr["Content-Type"] = "application/json"
		}
		if rec := reqWith(t, mux, c.method, c.path, c.body, hdr); !isHostRefusal(rec) {
			t.Errorf("%s %s: %d", c.method, c.path, rec.Code)
		}
	}
}

// Host mode with an Admin that goes by the peer's address: a request
// under an unlisted Host from that address is refused before Admin runs.
func TestHost_AddressAdminUnderAnUnlistedHostIsRefused(t *testing.T) {
	byAddress := func(r *http.Request) bool { return strings.HasPrefix(r.RemoteAddr, "192.0.2.1:") } // httptest's peer
	_, mux := newTestHandler(t, testOptions{configured: true, admin: byAddress, controller: NewMemoryController()})
	if rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/pairing", "", foreignHost(nil)); !isHostRefusal(rec) {
		t.Fatalf("pairing: %d %s", rec.Code, rec.Body)
	}
	if rec := reqWith(t, mux, http.MethodPost, DefaultPrefix+"/controller/commission", `{"code":"34970112332"}`,
		foreignHost(map[string]string{"Content-Type": "application/json"})); !isHostRefusal(rec) {
		t.Fatalf("commission: %d %s", rec.Code, rec.Body)
	}
}

// The name rules: loopback names, IP literals without a zone and listed
// names pass, with the port, brackets, case and one trailing dot ignored;
// anything else and an empty Host do not.
func TestHost_NameRules(t *testing.T) {
	h := &handler{allowedHosts: map[string]bool{"fabric.example": true}}
	for host, want := range map[string]bool{
		"localhost": true, "LOCALHOST:8080": true, "ui.localhost": true, "localhost.": true,
		"127.0.0.1:80": true, "[::1]:8099": true, "::1": true, "192.168.1.20": true, "[fe80::1%25eth0]:80": false,
		"fabric.example": true, "Fabric.Example.:443": true,
		"": false, "other.example": false, "localhost.example": false, "fabric.example.evil": false,
	} {
		if got := h.nameAllowed(hostName(host)); got != want {
			t.Errorf("%q: %v, want %v", host, got, want)
		}
	}
}

// NewHandler and Mount refuse an AllowedHosts entry that is not a host
// name.
func TestHost_AllowedHostsAreNamesOnly(t *testing.T) {
	for _, e := range []string{"", "https://a.example", "a.example:8080", "a.example/x", "*.example", "a .example", "[::1]"} {
		if _, err := NewHandler(Options{AllowedHosts: []string{e}}); err == nil {
			t.Errorf("NewHandler accepted %q", e)
		}
		if err := Mount(http.NewServeMux(), Options{AllowedHosts: []string{e}}); err == nil {
			t.Errorf("Mount accepted %q", e)
		}
	}
}

// countingPredicate is an AllowHost that admits by the peer's address and
// counts its calls.
func countingPredicate(admit bool) (func(*http.Request) bool, *atomic.Int64) {
	var n atomic.Int64
	return func(r *http.Request) bool {
		n.Add(1)
		return admit && strings.HasPrefix(r.RemoteAddr, "192.0.2.1:")
	}, &n
}

// A predicate that admits by the peer's address admits an unlisted Host,
// once per request, in both modes (decision 61).
func TestAllowHost_PredicateAdmitsAnUnlistedHost(t *testing.T) {
	for name, admin := range map[string]func(*http.Request) bool{"token mode": nil, "host mode": func(*http.Request) bool { return true }} {
		t.Run(name, func(t *testing.T) {
			pred, calls := countingPredicate(true)
			_, mux := newTestHandler(t, testOptions{admin: admin, allowHost: pred})
			rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{"Host": "unlisted.example"})
			if rec.Code != http.StatusOK || calls.Load() != 1 {
				t.Fatalf("%d, %d calls: %s", rec.Code, calls.Load(), rec.Body)
			}
			if admin == nil && !strings.Contains(rec.Body.String(), `"csrf_token":"`+testToken+`"`) {
				t.Fatalf("token mode without the token: %s", rec.Body)
			}
		})
	}
}

// A predicate that refuses leaves the refusal: no token, on the status,
// an admin route and the bundle.
func TestAllowHost_PredicateRefuses(t *testing.T) {
	pred, _ := countingPredicate(false)
	h, mux := newTestHandler(t, testOptions{allowHost: pred})
	mux.Handle("/", h.hostCheck(spaHandler(nil)))
	for _, p := range []string{DefaultPrefix + "/node/status", DefaultPrefix + "/node/pairing", "/"} {
		rec := reqWith(t, mux, http.MethodGet, p, "", map[string]string{"Host": "unlisted.example", TokenHeader: testToken})
		if !isHostRefusal(rec) || strings.Contains(rec.Body.String(), "csrf_token") {
			t.Errorf("%s: %d %s", p, rec.Code, rec.Body)
		}
	}
}

// Without a predicate the names decide; with one, a request the names
// admit never asks it.
func TestAllowHost_NilMeansTheListOnly(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for host, want := range map[string]int{"example.com": http.StatusOK, "localhost:1": http.StatusOK, "10.0.0.1": http.StatusOK, "unlisted.example": http.StatusForbidden} {
		if rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{"Host": host}); rec.Code != want {
			t.Errorf("%s: %d, want %d", host, rec.Code, want)
		}
	}
	pred, calls := countingPredicate(true)
	_, mux = newTestHandler(t, testOptions{allowHost: pred})
	for _, host := range []string{"example.com", "localhost:1", "10.0.0.1"} {
		reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{"Host": host})
	}
	if calls.Load() != 0 {
		t.Fatalf("the predicate was asked %d times for admitted names", calls.Load())
	}
}

// In both modes, every route of the contract — the streams as HEAD — and
// the bundle refuse an unlisted Host the predicate refuses.
func TestAllowHost_EveryModeAndRoute(t *testing.T) {
	ops := contractOps(t)
	for name, admin := range map[string]func(*http.Request) bool{"token mode": nil, "host mode": func(*http.Request) bool { return true }} {
		pred, _ := countingPredicate(false)
		h, mux := newTestHandler(t, testOptions{
			admin: admin, allowHost: pred, controller: NewMemoryController(), operations: sampleOperations(),
			exposure: newMemoryProvider(), registry: newMemoryRegistry(0), nodeConfig: defaultNodeConfig(),
		})
		mux.Handle("/", h.hostCheck(spaHandler(nil)))
		targets := []string{"GET /", "GET /index.html"}
		for _, op := range ops {
			m := op.method
			if strings.HasSuffix(op.path, "/stream") {
				m = http.MethodHead
			}
			targets = append(targets, m+" "+DefaultPrefix+samplePath(op.path))
		}
		for _, tg := range targets {
			method, path, _ := strings.Cut(tg, " ")
			rec := reqWith(t, mux, method, path, "", map[string]string{"Host": "unlisted.example", TokenHeader: testToken})
			if !isHostRefusal(rec) {
				t.Errorf("%s: %s %d", name, tg, rec.Code)
			}
		}
	}
}

// An empty Host is refused before the predicate is asked.
func TestAllowHost_EmptyHostIsRefusedUnasked(t *testing.T) {
	pred, calls := countingPredicate(true)
	_, mux := newTestHandler(t, testOptions{allowHost: pred})
	if rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{"Host": ""}); !isHostRefusal(rec) || calls.Load() != 0 {
		t.Fatalf("%d, %d calls", rec.Code, calls.Load())
	}
}
