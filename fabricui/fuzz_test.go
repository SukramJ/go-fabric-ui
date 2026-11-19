// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

// The standing fuzz targets of the handler (review decision 53): the
// phase 0 targets as permanent Fuzz functions. `go test` runs each on its
// seeds (hand-written edge cases and the requests of contract/fixtures);
// `make fuzz` runs each for FUZZTIME, and the nightly job does so at about
// three minutes per target.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode"
	"unicode/utf8"
)

// fuzzFixtures are the requests of contract/fixtures, sorted by file name
// so the seed corpus is the same on every run.
func fuzzFixtures(f *testing.F) []fixture {
	f.Helper()
	names, err := filepath.Glob(filepath.Join(fixtureDir, "*.json"))
	if err != nil || len(names) == 0 {
		f.Fatalf("no fixtures under %s: %v", fixtureDir, err)
	}
	sort.Strings(names)
	out := make([]fixture, 0, len(names))
	for _, name := range names {
		raw, err := os.ReadFile(name)
		if err != nil {
			f.Fatal(err)
		}
		var fx fixture
		if err := json.Unmarshal(raw, &fx); err != nil {
			f.Fatalf("%s: %v", name, err)
		}
		out = append(out, fx)
	}
	return out
}

// wantBadRequest fails unless err is a 400 bad_request apiProblem.
func wantBadRequest(t *testing.T, what string, err error) {
	t.Helper()
	p, ok := errors.AsType[*apiProblem](err)
	if !ok || p.status != http.StatusBadRequest || p.code != codeBadRequest {
		t.Fatalf("%s: error %v (%T), want a 400 bad_request problem", what, err, err)
	}
}

// wantProblem fails unless rec is a problem+json answer whose body agrees
// with the status and the X-Problem-Code header. A HEAD answer without a
// body is held to the headers only.
func wantProblem(t *testing.T, method string, rec *httptest.ResponseRecorder) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("%d with content type %q, want application/problem+json; body %.200q", rec.Code, ct, rec.Body)
	}
	code := rec.Header().Get("X-Problem-Code")
	if code == "" {
		t.Fatalf("%d problem without X-Problem-Code", rec.Code)
	}
	if method == http.MethodHead && rec.Body.Len() == 0 {
		return
	}
	var p problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("%d problem body does not decode: %v; %.200q", rec.Code, err, rec.Body)
	}
	if p.Status != rec.Code || p.Code != code {
		t.Fatalf("problem status %d code %q, answer %d X-Problem-Code %q", p.Status, p.Code, rec.Code, code)
	}
}

// FuzzParseHex: parseHex, parseHex32 and parseNodeID. A 64-bit id round
// trips through hex16, a 32-bit one through hexID and agrees with the
// 64-bit parse; every refusal is a 400 bad_request.
func FuzzParseHex(f *testing.F) {
	for _, s := range []string{
		"0", "1", "0x1", "0X1", "0xFFFFFFFF", "100000000", "FFFFFFFFFFFFFFFF", "0x10000000000000000",
		"00000000000000001", " 0XaB ", "", " ", "0x", "-1", "+1", "0x-1", "1_0", "ｆ", "0x0X1", "000000000000001B669",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := parseNodeID("id", s)
		if err != nil {
			wantBadRequest(t, "parseNodeID", err)
		} else if back, err := parseHex(hex16(v), 64); err != nil || back != v {
			t.Fatalf("%q -> %d -> %s -> %d, %v", s, v, hex16(v), back, err)
		}
		v32, err := parseHex32("cluster", s)
		if err != nil {
			wantBadRequest(t, "parseHex32", err)
			return
		}
		if back, err := parseHex(hexID(v32), 32); err != nil || back != uint64(v32) {
			t.Fatalf("%q -> %d -> %s -> %d, %v", s, v32, hexID(v32), back, err)
		}
		if v64, err := parseHex(s, 64); err != nil || v64 != uint64(v32) {
			t.Fatalf("%q: 32-bit %d, 64-bit %d, %v", s, v32, v64, err)
		}
	})
}

// FuzzQueryWindows: queryInt, eventWindow, logWindow, pathEndpoint and
// endpointOf. An accepted limit is within 1..max, an accepted endpoint a
// 16-bit number; every refusal is a 400 bad_request.
func FuzzQueryWindows(f *testing.F) {
	for _, s := range [][3]string{
		{"", "", "1"},
		{"0", "1", "0"},
		{"5", "1000", "65535"},
		{"-1", "0", "65536"},
		{"x", "2001", "-1"},
		{"9223372036854775807", "9223372036854775808", "01"},
		{" 1", "1 ", "+1"},
		{"1e3", "0x10", ""},
	} {
		f.Add(s[0], s[1], s[2], 0)
	}
	f.Add("", "", "1", -1)
	f.Add("", "", "1", 65536)
	f.Add("", "", "1", math.MaxInt)
	f.Fuzz(func(t *testing.T, since, limit, ep string, n int) {
		q := url.Values{}
		if since != "" {
			q.Set("since", since)
		}
		if limit != "" {
			q.Set("limit", limit)
		}
		r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x?"+q.Encode(), http.NoBody)
		if _, l, err := logWindow(r); err != nil {
			wantBadRequest(t, "logWindow", err)
		} else if l < 1 || l > maxLogLimit {
			t.Fatalf("log limit %d", l)
		}
		sinceQ, errS := queryInt(r, "since")
		limitQ, errL := queryInt(r, "limit")
		switch {
		case errS != nil:
			wantBadRequest(t, "queryInt since", errS)
		case errL != nil:
			wantBadRequest(t, "queryInt limit", errL)
		default:
			if _, l, err := eventWindow(sinceQ, limitQ); err != nil {
				wantBadRequest(t, "eventWindow", err)
			} else if l < 1 || l > maxEventLimit {
				t.Fatalf("event limit %d", l)
			}
		}
		r.SetPathValue("ep", ep)
		if v, err := pathEndpoint(r); err != nil {
			wantBadRequest(t, "pathEndpoint", err)
		} else if want, perr := strconv.ParseUint(ep, 10, 64); perr != nil || want != uint64(v) {
			t.Fatalf("ep %q -> %d", ep, v)
		}
		if v, err := endpointOf("endpoint", n); err != nil {
			wantBadRequest(t, "endpointOf", err)
			if n >= 0 && n <= 0xFFFF {
				t.Fatalf("endpoint %d refused", n)
			}
		} else if int(v) != n {
			t.Fatalf("endpoint %d -> %d", n, v)
		}
	})
}

// validNodeKey is the contract's NodeKey pattern spelled out by hand.
func validNodeKey(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for i := range len(s) {
		c := s[i]
		alnum := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
		if !alnum && (i == 0 || c != '.' && c != '_' && c != '-') {
			return false
		}
	}
	return true
}

// FuzzStrings: nodeLabel, checkNodeKey, parseLevel and levelName. An
// accepted label is trimmed, non-empty, at most MaxNodeLabel runes and
// free of control characters; an accepted key matches the contract's
// pattern and is no page name; a level name round trips.
func FuzzStrings(f *testing.F) {
	for _, s := range []string{
		"", " ", "Kitchen", "  Living room  ", "a\x00b", "tab\there", strings.Repeat("ü", 64), strings.Repeat("ü", 65),
		"default", "status", "node-1", "_x", "a.b_c-d", strings.Repeat("k", 64), strings.Repeat("k", 65),
		"debug", "info", "warn", "error", "INFO", "\xff",
	} {
		f.Add(s, 0)
	}
	f.Add("info", -8)
	f.Add("info", 12)
	f.Add("info", math.MinInt32)
	f.Fuzz(func(t *testing.T, s string, n int) {
		if l, err := nodeLabel(s); err == nil {
			if l == "" || l != strings.TrimSpace(l) || l != strings.TrimSpace(s) ||
				utf8.RuneCountInString(l) > MaxNodeLabel || strings.ContainsFunc(l, unicode.IsControl) {
				t.Fatalf("label %q accepted as %q", s, l)
			}
		}
		if err := checkNodeKey(s, nil); err == nil {
			if !validNodeKey(s) || slices.Contains(reservedNodeKeys, s) {
				t.Fatalf("key %q accepted", s)
			}
		} else if validNodeKey(s) && !slices.Contains(reservedNodeKeys, s) {
			t.Fatalf("key %q refused: %v", s, err)
		}
		if l, ok := parseLevel(s); ok && levelName(l) != s {
			t.Fatalf("level %q -> %v -> %q", s, l, levelName(l))
		}
		if _, ok := parseLevel(levelName(slog.Level(n))); !ok {
			t.Fatalf("level %d named %q, which does not parse", n, levelName(slog.Level(n)))
		}
	})
}

// FuzzDecodeBodies: decodeJSON and badBody into the hand-written bodies
// (the pairing window, the exposure update list, the log level and the
// node label). A refusal is a 400 or, for a body over maxBodyBytes, a 413
// problem; an accepted body writes nothing.
func FuzzDecodeBodies(f *testing.F) {
	kinds := map[string]uint8{"/pairing/window": 0, "/exposure": 1, "/exposure/preview": 1, "/operation/log/level": 2, "/label": 3, "/nodes": 3}
	for _, fx := range fuzzFixtures(f) {
		for suffix, kind := range kinds {
			if len(fx.Request.Body) > 0 && strings.HasSuffix(fx.Request.Path, suffix) {
				f.Add(kind, []byte(fx.Request.Body))
			}
		}
	}
	for kind := range uint8(4) {
		for _, b := range []string{"", "{}", "null", "[]", "{", `{"x":1}`, `{"level":"info"}{}`, `{"duration_seconds":-1}`, `{"items":null}`, `{"label":"\u0000"}`} {
			f.Add(kind, []byte(b))
		}
	}
	f.Add(uint8(3), []byte(`{"label":"`+strings.Repeat("a", maxBodyBytes)+`"}`))
	f.Fuzz(func(t *testing.T, kind uint8, body []byte) {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, DefaultPrefix+"/x", bytes.NewReader(body))
		rec := httptest.NewRecorder()
		var err error
		switch kind % 4 {
		case 0:
			var v windowRequest
			err = decodeJSON(rec, req, &v)
		case 1:
			var v updateList
			err = decodeJSON(rec, req, &v)
		case 2:
			var v logLevelRequest
			err = decodeJSON(rec, req, &v)
		default:
			var v nodeLabelRequest
			err = decodeJSON(rec, req, &v)
		}
		if err == nil {
			if rec.Body.Len() != 0 || len(rec.Header()) != 0 {
				t.Fatalf("an accepted body wrote %d bytes and headers %v", rec.Body.Len(), rec.Header())
			}
			return
		}
		badBody(rec, req, err)
		_, tooLarge := errors.AsType[*http.MaxBytesError](err)
		switch {
		case tooLarge && rec.Code != http.StatusRequestEntityTooLarge, !tooLarge && rec.Code != http.StatusBadRequest:
			t.Fatalf("refusal %v answered %d", err, rec.Code)
		}
		wantProblem(t, http.MethodPut, rec)
	})
}

// fuzzAdminHeaders marks req as an admin request: the token and the
// admin request header.
func fuzzAdminHeaders(req *http.Request) {
	req.Header.Set(TokenHeader, testToken)
	req.Header.Set("X-Fabric-UI", "1")
}

// fuzzSecrets are the string values of the body's secret members, long
// enough and with a digit, so that a match in the log is not a word of
// the log's own.
func fuzzSecrets(body []byte) []string {
	var m map[string]any
	if json.Unmarshal(body, &m) != nil {
		return nil
	}
	var out []string
	for _, k := range []string{"code", "passcode", "passphrase", "dataset_hex"} {
		var s string
		switch v := m[k].(type) {
		case string:
			s = v
		case float64:
			s = strconv.FormatFloat(v, 'f', -1, 64)
		}
		if len(s) >= 8 && strings.ContainsAny(s, "0123456789") {
			out = append(out, s)
		}
	}
	return out
}

// logTime is the time attribute of a text log line, left out before the
// log is searched for a secret.
var logTime = regexp.MustCompile(`time=\S+`)

// FuzzControllerBodies: every controller write route a fixture names,
// with a fuzzed body, on MemoryController, as an admin. No answer is a 500
// or above 503; every refusal is a problem; no secret of the body (code,
// passcode, passphrase, dataset_hex) reaches the log.
func FuzzControllerBodies(f *testing.F) {
	type route struct{ method, path string }
	var routes []route
	seen := map[route]bool{}
	for _, fx := range fuzzFixtures(f) {
		r := route{fx.Request.Method, fx.Request.Path}
		if r.method == http.MethodGet || !strings.HasPrefix(r.path, "/controller/") || len(fx.Request.Body) == 0 {
			continue
		}
		if !seen[r] {
			seen[r] = true
			routes = append(routes, r)
		}
		f.Add(uint8(slices.Index(routes, r)), []byte(fx.Request.Body)) //nolint:gosec // fewer than 256 routes
	}
	if len(routes) == 0 || len(routes) > 255 {
		f.Fatalf("%d controller write routes among the fixtures", len(routes))
	}
	for j := range routes {
		for _, b := range []string{"", "{}", "null", `{"code":"MT:Y.K9042C00KA0648G00"}`, `{"passcode":20202021}`} {
			f.Add(uint8(j), []byte(b)) //nolint:gosec // fewer than 256 routes
		}
	}
	f.Fuzz(func(t *testing.T, which uint8, body []byte) {
		rt := routes[int(which)%len(routes)]
		var log bytes.Buffer
		_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), log: &log})
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req := httptest.NewRequestWithContext(ctx, rt.method, DefaultPrefix+rt.path, bytes.NewReader(body))
		req.Host = "localhost"
		req.Header.Set("Content-Type", "application/json")
		fuzzAdminHeaders(req)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusInternalServerError || rec.Code > http.StatusServiceUnavailable {
			t.Fatalf("%s %s answered %d: %.300q", rt.method, rt.path, rec.Code, rec.Body)
		}
		if rec.Code >= 300 {
			wantProblem(t, rt.method, rec)
		}
		text := logTime.ReplaceAllString(log.String(), "")
		for _, s := range fuzzSecrets(body) {
			if strings.Contains(text, s) && !strings.Contains(rt.path, s) {
				t.Fatalf("%s %s logged the secret %q", rt.method, rt.path, s)
			}
		}
	})
}

// refreshSpy is MemoryController that notes a topology read with refresh.
type refreshSpy struct {
	*MemoryController
	refreshed atomic.Bool
}

func (s *refreshSpy) NetworkTopology(ctx context.Context, refresh bool) (NetworkTopology, error) {
	if refresh {
		s.refreshed.Store(true)
	}
	return s.MemoryController.NetworkTopology(ctx, refresh)
}

// matchPattern reports whether path matches a route pattern whose {name}
// segments match any non-empty segment.
func matchPattern(pattern, path string) bool {
	ps, xs := strings.Split(pattern, "/"), strings.Split(path, "/")
	if len(ps) != len(xs) {
		return false
	}
	for i := range ps {
		if strings.HasPrefix(ps[i], "{") && strings.HasSuffix(ps[i], "}") {
			if xs[i] == "" {
				return false
			}
		} else if ps[i] != xs[i] {
			return false
		}
	}
	return true
}

// isAdminRoute reports whether method and path name a route the contract
// marks x-admin: the node family's (under /node and /nodes/{node}), the
// registry's, the operation family's and the controller family's.
func isAdminRoute(method, path string) bool {
	rel, ok := strings.CutPrefix(path, DefaultPrefix)
	if !ok {
		return false
	}
	if method == http.MethodHead {
		method = http.MethodGet
	}
	for _, rt := range nodeRoutes {
		if rt.admin && rt.method == method && (matchPattern("/node"+rt.path, rel) || matchPattern("/nodes/{node}"+rt.path, rel)) {
			return true
		}
	}
	for _, rt := range registryRoutes {
		if rt.method == method && matchPattern("/nodes"+rt.path, rel) {
			return true
		}
	}
	for _, table := range []map[string]bool{operationAdmin, controllerAdmin} {
		for k := range table {
			m, p, _ := strings.Cut(k, " ")
			if m == method && matchPattern(p, rel) {
				return true
			}
		}
	}
	return false
}

// FuzzMux: the whole mux as Handler builds it (the API on the fake node
// with the registry, the exposure provider, MemoryOperations and
// MemoryController behind a refresh spy, and the bundle at "/"), in token
// mode and in Options.Admin mode, with and without rights, with a fuzzed
// method, request target, one header and body. No answer is a 500 or
// above 503; every answer under the prefix that is not 2xx or 3xx is a
// problem; without rights no admin route answers 2xx and no topology
// refresh reaches the controller.
func FuzzMux(f *testing.F) {
	for _, fx := range fuzzFixtures(f) {
		target := DefaultPrefix + fx.Request.Path
		f.Add(uint8(1), fx.Request.Method, target, "", "", []byte(fx.Request.Body))
		f.Add(uint8(0), fx.Request.Method, target, "", "", []byte(fx.Request.Body))
		f.Add(uint8(2), fx.Request.Method, target, "Origin", "http://localhost", []byte(fx.Request.Body))
	}
	for _, s := range []string{
		"/", "/index.html", "/assets/x.js", "/api/v1", "/api/v1/", "/api/v1/no/such", "/api/v1/node/../controller/nodes",
		"/api/v1/controller/topology?refresh=true", "/api/v1/nodes/%2e%2e/label", "/api/v1/node/stream", "//api/v1/node/status",
	} {
		f.Add(uint8(0), http.MethodGet, s, "Accept", "application/json", []byte(nil))
		f.Add(uint8(3), http.MethodHead, s, TokenHeader, "wrong", []byte(nil))
	}
	f.Fuzz(func(t *testing.T, mode uint8, method, target, hname, hval string, body []byte) {
		tokenMode, rights := mode&2 == 0, mode&1 == 1
		spy := &refreshSpy{MemoryController: NewMemoryController()}
		o := testOptions{
			exposure: newMemoryProvider(), controller: spy, operations: sampleOperations(),
			registry: newMemoryRegistry(0), nodeConfig: defaultNodeConfig(),
		}
		if !tokenMode {
			o.admin = func(*http.Request) bool { return rights }
		}
		_, mux := newTestHandler(t, o)
		mux.Handle("/", spaHandler(nil))
		if !strings.HasPrefix(target, "/") {
			target = "/" + target
		}
		path := target
		if i := strings.IndexAny(path, "?#"); i >= 0 {
			path = path[:i]
		}
		// A stream answers its headers and ends with the request.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if strings.HasSuffix(path, "/stream") {
			cancel()
		}
		req, err := http.NewRequestWithContext(ctx, method, "http://localhost"+target, bytes.NewReader(body))
		if err != nil {
			t.Skip()
		}
		req.Host = "localhost"
		if hname != "" && (rights || http.CanonicalHeaderKey(hname) != TokenHeader) {
			req.Header.Set(hname, hval)
		}
		if rights {
			fuzzAdminHeaders(req)
		} else {
			req.Header.Del(TokenHeader)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code == http.StatusInternalServerError || rec.Code > http.StatusServiceUnavailable {
			t.Fatalf("%s %s answered %d: %.300q", method, target, rec.Code, rec.Body)
		}
		underPrefix := strings.HasPrefix(req.URL.Path, DefaultPrefix+"/")
		if underPrefix && rec.Code >= 400 {
			wantProblem(t, method, rec)
		}
		if !rights {
			if rec.Code < 300 && isAdminRoute(method, req.URL.Path) {
				t.Fatalf("%s %s without rights answered %d", method, target, rec.Code)
			}
			if spy.refreshed.Load() {
				t.Fatalf("%s %s without rights refreshed the topology", method, target)
			}
		}
	})
}
