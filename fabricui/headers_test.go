// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
)

// contractOperations are every operation of contract/openapi.yaml as
// "METHOD /path", scanned as adminRoutesOfContract describes.
func contractOperations(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pathRe := regexp.MustCompile(`^ {2}(/\S+):\s*$`)
	methodRe := regexp.MustCompile(`^ {4}(get|put|post|delete|patch):\s*$`)
	var out []string
	var path string
	for line := range strings.SplitSeq(string(b), "\n") {
		if m := pathRe.FindStringSubmatch(line); m != nil {
			path = m[1]
		} else if m := methodRe.FindStringSubmatch(line); m != nil && path != "" {
			out = append(out, strings.ToUpper(m[1])+" "+path)
		}
	}
	return out
}

// Every answer of every route of the contract — JSON, problem, stream —
// carries Cache-Control: no-store (UR-P0-leak-25), X-Content-Type-Options:
// nosniff and Referrer-Policy: no-referrer (UR-P0-web-24); so does the 404
// of a path the contract does not name.
func TestAPI_EveryRouteAnswersWithTheSecurityHeaders(t *testing.T) {
	ops := contractOperations(t)
	if len(ops) < 80 {
		t.Fatalf("found %d operations in the contract; the scan is not reading it", len(ops))
	}
	_, mux := newTestHandler(t, testOptions{
		exposure: newMemoryProvider(), controller: NewMemoryController(), operations: sampleOperations(),
		registry: newMemoryRegistry(0), nodeConfig: defaultNodeConfig(),
	})
	params := strings.NewReplacer("{node}", "default", "{ep}", "1", "{id}", "0000000000000011", "{index}", "1", "{cluster_hex}", "0006")
	done, cancel := context.WithCancel(context.Background())
	cancel()
	for _, op := range append(ops, "GET /no/such/route", "GET /controller/no/such/route") {
		method, path, _ := strings.Cut(op, " ")
		ctx := context.Background()
		if strings.HasSuffix(path, "/stream") {
			ctx = done // a stream answers its headers and ends with the request
		}
		req := httptest.NewRequestWithContext(ctx, method, DefaultPrefix+params.Replace(path), http.NoBody)
		req.Header.Set(TokenHeader, testToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		for k, v := range map[string]string{
			"Cache-Control": "no-store", "X-Content-Type-Options": "nosniff", "Referrer-Policy": "no-referrer",
		} {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s (%d): %s %q, want %q", op, rec.Code, k, got, v)
			}
		}
	}
}

// A page of another origin reads no answer of the API: no route sends an
// Access-Control-* header to a request that names another origin, and no
// preflight is answered with success, so a request that needs one (the
// token header, a JSON body) is not sent (UR-P2-test-95; a guard, green).
func TestAPI_NoRouteAdmitsAnotherOrigin(t *testing.T) {
	ops := contractOperations(t)
	_, mux := newTestHandler(t, testOptions{
		exposure: newMemoryProvider(), controller: NewMemoryController(), operations: sampleOperations(),
		registry: newMemoryRegistry(0), nodeConfig: defaultNodeConfig(),
	})
	params := strings.NewReplacer("{node}", "default", "{ep}", "1", "{id}", "0000000000000011", "{index}", "1", "{cluster_hex}", "0006")
	done, cancel := context.WithCancel(context.Background())
	cancel()
	const other = "https://other.example"
	for _, op := range ops {
		method, path, _ := strings.Cut(op, " ")
		ctx := context.Background()
		if strings.HasSuffix(path, "/stream") {
			ctx = done
		}
		target := DefaultPrefix + params.Replace(path)
		req := httptest.NewRequestWithContext(ctx, method, target, http.NoBody)
		req.Header.Set("Origin", other)
		pre := httptest.NewRequestWithContext(ctx, http.MethodOptions, target, http.NoBody)
		pre.Header.Set("Origin", other)
		pre.Header.Set("Access-Control-Request-Method", method)
		pre.Header.Set("Access-Control-Request-Headers", TokenHeader+", content-type")
		for _, r := range []*http.Request{req, pre} {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, r)
			for k := range rec.Header() {
				if strings.HasPrefix(http.CanonicalHeaderKey(k), "Access-Control-") {
					t.Errorf("%s %s (%d): %s: %q", r.Method, path, rec.Code, k, rec.Header().Get(k))
				}
			}
			if r == pre && rec.Code < 300 {
				t.Errorf("preflight of %s: %d, want no success", op, rec.Code)
			}
		}
	}
}
