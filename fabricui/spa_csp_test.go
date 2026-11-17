// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// sha256Source is the CSP hash-source of s.
func sha256Source(s string) string {
	sum := sha256.Sum256([]byte(s))
	return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}

// cspDirectives splits a policy into its directives by name.
func cspDirectives(t *testing.T, policy string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for d := range strings.SplitSeq(policy, ";") {
		name, value, _ := strings.Cut(strings.TrimSpace(d), " ")
		if name != "" {
			out[name] = value
		}
	}
	return out
}

func getPath(h http.Handler, p string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, p, nil))
	return rec
}

// The index carries a CSP whose script-src admits the inline pre-paint
// script by its hash and nothing inline else, framing by the same origin
// only (UR-P0-web-23, UR-P0-web-24); every response of the bundle carries
// nosniff and no-referrer.
func TestSPA_IndexCarriesTheCSP(t *testing.T) {
	inline := "\n      document.documentElement.dataset.skin = \"ha\";\n    "
	h := spaHandlerFS(fstest.MapFS{
		"index.html":      {Data: []byte("<!doctype html><head><script>" + inline + "</script><script type=\"module\" src=\"./assets/app-1.js\"></script></head>")},
		"assets/app-1.js": {Data: []byte("console.log(1)")},
	}, nil)
	for _, p := range []string{"/", "/deep/link"} {
		rec := getPath(h, p)
		d := cspDirectives(t, rec.Header().Get("Content-Security-Policy"))
		for name, want := range map[string]string{
			"default-src":     "'self'",
			"script-src":      "'self' " + sha256Source(inline),
			"style-src":       "'self'",
			"img-src":         "'self'",
			"object-src":      "'none'",
			"base-uri":        "'none'",
			"form-action":     "'self'",
			"frame-ancestors": "'self'",
		} {
			if d[name] != want {
				t.Errorf("%s: %s %q, want %q", p, name, d[name], want)
			}
		}
	}
	if csp := getPath(h, "/").Header().Get("Content-Security-Policy"); strings.Contains(csp, "unsafe") {
		t.Errorf("policy admits something unsafe: %s", csp)
	}
	for _, p := range []string{"/", "/assets/app-1.js"} {
		rec := getPath(h, p)
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Errorf("%s: headers %v", p, rec.Header())
		}
	}
}

// Options.FrameAncestors adds parents after 'self' (UR-P0-web-23).
func TestSPA_FrameAncestors(t *testing.T) {
	h := spaHandlerFS(fstest.MapFS{"index.html": {Data: []byte("<!doctype html>")}}, []string{"https://dash.example:8123", "http://10.0.0.2"})
	d := cspDirectives(t, getPath(h, "/").Header().Get("Content-Security-Policy"))
	if got := d["frame-ancestors"]; got != "'self' https://dash.example:8123 http://10.0.0.2" {
		t.Fatalf("frame-ancestors %q", got)
	}
	srv := httptest.NewServer(Handler(Options{FrameAncestors: []string{"https://dash.example"}}))
	defer srv.Close()
	resp, err := srv.Client().Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if d := cspDirectives(t, resp.Header.Get("Content-Security-Policy")); d["frame-ancestors"] != "'self' https://dash.example" {
		t.Fatalf("Handler: frame-ancestors %q", d["frame-ancestors"])
	}
}

// The not-built page shares the index's path and its policy, with its own
// inline style admitted by hash.
func TestSPA_NotBuiltCarriesTheCSP(t *testing.T) {
	rec := getPath(spaHandlerFS(fstest.MapFS{"README": {Data: []byte("x")}}, nil), "/")
	style := notBuiltHTML[strings.Index(notBuiltHTML, "<style>")+len("<style>") : strings.Index(notBuiltHTML, "</style>")]
	d := cspDirectives(t, rec.Header().Get("Content-Security-Policy"))
	if rec.Code != http.StatusServiceUnavailable || d["frame-ancestors"] != "'self'" || d["style-src"] != "'self' "+sha256Source(style) {
		t.Fatalf("%d %v", rec.Code, d)
	}
}

// The committed bundle's index: the policy admits exactly its inline
// pre-paint script, read here independently of the handler's scan, so a
// change of the script cannot leave a stale hash behind.
func TestSPA_EmbeddedIndexScriptHash(t *testing.T) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		t.Fatal(err)
	}
	b, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	page := string(b)
	start := strings.Index(page, "<script>")
	end := strings.Index(page, "</script>")
	if start < 0 || end < start || strings.Count(page, "<script>") != 1 {
		t.Fatal("dist/index.html holds no single inline <script>; the pre-paint script moved")
	}
	want := "'self' " + sha256Source(page[start+len("<script>"):end])
	d := cspDirectives(t, getPath(spaHandler(nil), "/").Header().Get("Content-Security-Policy"))
	if d["script-src"] != want {
		t.Fatalf("script-src %q, want %q", d["script-src"], want)
	}
}

// Each FrameAncestors entry is a CSP host-source or scheme-source; New
// (through Handler) and Mount refuse anything else (UR-P0-web-23).
func TestOptions_FrameAncestorsValidated(t *testing.T) {
	for _, ok := range []string{
		"https:", "https://dash.example", "https://*.example.com", "http://10.0.0.2:8123",
		"dash.example", "dash.example:*", "https://dash.example/panel/", "*",
	} {
		if err := Mount(http.NewServeMux(), Options{FrameAncestors: []string{ok}}); err != nil {
			t.Errorf("%q refused: %v", ok, err)
		}
	}
	for _, bad := range []string{
		"", "'self'", "'none'", "https://a b", "a;b", "https://x/;y", "a,b", "*.", "https://", "https://dash.example:x",
		"https://[::1]", "javascript:alert(1)", "dash.example\n",
	} {
		if err := Mount(http.NewServeMux(), Options{FrameAncestors: []string{bad}}); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("Handler accepted an invalid FrameAncestors entry")
		}
	}()
	Handler(Options{FrameAncestors: []string{"'self'"}})
}
