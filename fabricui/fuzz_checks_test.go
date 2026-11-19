// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// FuzzHostCheck: the Host check (decisions 37, 61) on any Host value, with
// a predicate that records its calls. An admitted name passes the name
// rules or the predicate admitted it; the predicate is asked at most once,
// never for an empty Host nor for a name the rules admit; a refusal is a
// 403 host_not_allowed without a token.
func FuzzHostCheck(f *testing.F) {
	for _, s := range []string{
		"", "localhost", "LOCALHOST:8080", "a.localhost.", "127.0.0.1:80", "[::1]:1", "::1", "[fe80::1%eth0]",
		"example.com", "Example.COM.:443", "unlisted.example", "localhost.example", "[", "]:", "1.2.3.4.5", "@:", "%00",
	} {
		f.Add(s, false)
		f.Add(s, true)
	}
	h, mux := newTestHandler(f, testOptions{})
	f.Fuzz(func(t *testing.T, host string, admit bool) {
		var calls atomic.Int64
		h.allowHost = func(*http.Request) bool { calls.Add(1); return admit }
		name := hostName(host)
		byName := h.nameAllowed(name)
		a, ok := h.admitHost(&http.Request{Host: host})
		switch {
		case name == "" && (ok || calls.Load() != 0):
			t.Fatalf("empty name %q admitted %v, %d calls", host, ok, calls.Load())
		case byName && (!ok || a.byPredicate || calls.Load() != 0):
			t.Fatalf("%q by name: ok %v, byPredicate %v, %d calls", host, ok, a.byPredicate, calls.Load())
		case !byName && name != "" && (ok != admit || calls.Load() != 1 || a.byPredicate != admit):
			t.Fatalf("%q by predicate %v: ok %v, %d calls", host, admit, ok, calls.Load())
		}
		if byName && name != "localhost" && !strings.HasSuffix(name, ".localhost") && name != "example.com" {
			if addr, err := netip.ParseAddr(name); err != nil || addr.Zone() != "" {
				t.Fatalf("%q admitted by name as %q", host, name)
			}
		}
		req := httptest.NewRequest(http.MethodGet, DefaultPrefix+"/node/status", http.NoBody)
		req.Host = host
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if refused := rec.Header().Get("X-Problem-Code") == codeHostNotAllowed; refused == ok {
			t.Fatalf("%q: %d with admission %v", host, rec.Code, ok)
		}
		if !ok && strings.Contains(rec.Body.String(), "csrf_token") {
			t.Fatalf("%q refused with a token", host)
		}
	})
}

// FuzzOrigin: an admitted Origin is http or https with a host the name
// rules admit, or the request's own Host under a predicate's admission
// (decisions 38, 61).
func FuzzOrigin(f *testing.F) {
	for _, s := range []string{
		"", "null", "http://localhost", "https://example.com", "http://example.com:8080", "http://other.example",
		"ftp://localhost", "http://", "http://user@localhost", "http://localhost/x", "http://[::1]:80", "HTTP://LOCALHOST", "http://unlisted.example:8123",
	} {
		f.Add(s, false)
		f.Add(s, true)
	}
	h, _ := newTestHandler(f, testOptions{})
	f.Fuzz(func(t *testing.T, origin string, byPredicate bool) {
		req := httptest.NewRequest(http.MethodPost, DefaultPrefix+"/node/pairing/window/close", http.NoBody)
		req.Host = "unlisted.example:8123"
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		req = req.WithContext(contextWithAdmission(req, byPredicate))
		ok := h.originAllowed(req)
		if origin == "" {
			if !ok {
				t.Fatal("no Origin refused")
			}
			return
		}
		if !ok {
			return
		}
		lower := strings.ToLower(origin)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			t.Fatalf("%q admitted without http(s)", origin)
		}
		rest := origin[strings.Index(origin, "://")+3:]
		if !h.nameAllowed(hostName(rest)) && (!byPredicate || !strings.EqualFold(rest, req.Host)) {
			t.Fatalf("%q admitted (byPredicate %v)", origin, byPredicate)
		}
	})
}

// FuzzMediaType: a body passes jsonBody only under application/json; a
// request without a body always passes; the body read after a pass is the
// body sent (decision 38).
func FuzzMediaType(f *testing.F) {
	for _, s := range []string{"", "application/json", "Application/JSON; charset=utf-8", "text/plain", "application/json-patch+json", "application/x-www-form-urlencoded", ";", "application/json;;"} {
		f.Add(s, []byte(`{}`), false)
		f.Add(s, []byte{}, true)
	}
	f.Fuzz(func(t *testing.T, ct string, body []byte, unknownLength bool) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		if unknownLength {
			req.ContentLength = -1
		}
		if ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		rec := httptest.NewRecorder()
		ok := jsonBody(rec, req)
		if len(body) == 0 && !ok {
			t.Fatalf("no body refused under %q", ct)
		}
		if ok && len(body) > 0 && !strings.EqualFold(strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]), "application/json") {
			t.Fatalf("body admitted under %q", ct)
		}
		if !ok && (rec.Code != http.StatusUnsupportedMediaType || rec.Header().Get("X-Problem-Code") != codeUnsupportedMediaType) {
			t.Fatalf("refusal %d", rec.Code)
		}
		if ok {
			var got bytes.Buffer
			_, _ = got.ReadFrom(req.Body)
			if !bytes.Equal(got.Bytes(), body) {
				t.Fatalf("body %q read back as %q", body, got.Bytes())
			}
		}
	})
}

// FuzzDecodeStrict: what decodeStrict accepts, encoding/json accepts too
// and decodes to the same value; nothing it accepts holds a second value
// (decision 43).
func FuzzDecodeStrict(f *testing.F) {
	for _, s := range []string{
		`{"label":"a"}`, `{"Label":"a"}`, `{"label":"a"} {}`, `{"label":"a","x":1}`, `{}`, `[]`, `null`,
		`{"code":"1","terms_and_conditions":{"version":1,"acknowledgements":2}}`, `{"code":"1","terms_and_conditions":{"VERSION":1,"acknowledgements":2}}`,
	} {
		f.Add([]byte(s))
	}
	type terms struct {
		Version         int `json:"version"`
		Acknowledgement int `json:"acknowledgements"`
	}
	type body struct {
		Label string           `json:"label"`
		Code  string           `json:"code"`
		Terms *terms           `json:"terms_and_conditions"`
		Map   map[string]terms `json:"map"`
		Raw   json.RawMessage  `json:"raw"`
		List  []terms          `json:"list"`
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var strict, lenient body
		if decodeStrict(data, &strict) != nil {
			return
		}
		if err := json.Unmarshal(data, &lenient); err != nil {
			t.Fatalf("strict accepted %q, encoding/json refuses: %v", data, err)
		}
		a, _ := json.Marshal(strict)
		b, _ := json.Marshal(lenient)
		if !bytes.Equal(a, b) {
			t.Fatalf("%q decoded as %s strictly, %s leniently", data, a, b)
		}
	})
}

// contextWithAdmission is req's context with the Host check's admission.
func contextWithAdmission(req *http.Request, byPredicate bool) context.Context {
	return context.WithValue(req.Context(), hostAdmissionKey{}, hostAdmission{byPredicate: byPredicate})
}

// FuzzEventSince: since_str is a decimal event number without leading
// zeros that round-trips exactly, else 400 (decision 44).
func FuzzEventSince(f *testing.F) {
	for _, s := range []string{"0", "1", "01", "18446744073709551615", "18446744073709551616", "-1", "", " 1", "1e3", "９"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := eventSince(7, &s)
		if err != nil {
			if p, ok := errors.AsType[*apiProblem](err); !ok || p.status != http.StatusBadRequest {
				t.Fatalf("%q: %v", s, err)
			}
			return
		}
		if strconv.FormatUint(v, 10) != s {
			t.Fatalf("%q read as %d", s, v)
		}
	})
}
