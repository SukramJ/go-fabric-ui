// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// openStream opens path on srv with the token and returns the response
// once the stream answered.
func openStream(t *testing.T, ctx context.Context, srv *httptest.Server, path string) *http.Response {
	t.Helper()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+path, http.NoBody)
	req.Header.Set(TokenHeader, testToken)
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

// One stream more than Options.MaxStreams is 503 too_many_streams with
// Retry-After: 5, whichever kind it is; a stream that ends frees its
// place (decision 42).
func TestStreams_MaxStreams(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), operations: sampleOperations()})
	h.streams.max = 2
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close) // after the streams' bodies close
	ctx, cancel := context.WithCancel(context.Background())
	first := openStream(t, ctx, srv, "/node/stream")
	_ = openStream(t, context.Background(), srv, "/controller/stream")
	for _, p := range []string{"/node/stream", "/operation/log/stream", "/controller/stream"} {
		res := openStream(t, context.Background(), srv, p)
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusServiceUnavailable || res.Header.Get("Retry-After") != "5" || !strings.Contains(string(body), `"code":"too_many_streams"`) {
			t.Fatalf("%s at the limit: %d Retry-After %q %s", p, res.StatusCode, res.Header.Get("Retry-After"), body)
		}
	}
	if first.StatusCode != http.StatusOK {
		t.Fatalf("first stream %d", first.StatusCode)
	}
	cancel()
	waitFor(t, func() bool { return h.streams.open.Load() == 1 })
	if res := openStream(t, context.Background(), srv, "/node/stream"); res.StatusCode != http.StatusOK {
		t.Fatalf("after one ended: %d", res.StatusCode)
	}
}

// Closing Options.Done ends every open stream (decision 51).
func TestStreams_DoneEndsThem(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{controller: NewMemoryController()})
	done := make(chan struct{})
	h.hostDone = done
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close) // after the streams' bodies close
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := openStream(t, ctx, srv, "/controller/stream")
	waitFor(t, func() bool { return h.streams.open.Load() == 1 })
	close(done)
	if _, err := io.ReadAll(res.Body); err != nil {
		t.Fatalf("the stream did not end cleanly: %v", err)
	}
	waitFor(t, func() bool { return h.streams.open.Load() == 0 })
}

// NewHandler refuses what Handler panics on: a negative MaxStreams and a
// Prefix other than the default (decisions 48, 49); Mount keeps any
// valid prefix.
func TestNewHandler_RefusesInvalidOptions(t *testing.T) {
	for name, o := range map[string]Options{
		"negative MaxStreams": {MaxStreams: -1},
		"another prefix":      {Prefix: "/ui-api"},
	} {
		if _, err := NewHandler(o); err == nil {
			t.Errorf("%s: no error", name)
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Handler did not panic", name)
				}
			}()
			Handler(o)
		}()
	}
	if _, err := NewHandler(Options{Prefix: DefaultPrefix}); err != nil {
		t.Fatal(err)
	}
	if err := Mount(http.NewServeMux(), Options{Prefix: "/ui-api"}); err != nil {
		t.Fatal(err)
	}
}
