// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A request without admin rights does not write a warning into the log
// on every call while the node's fabric read fails (a stopped node
// behind a NodeSwitch, a locked store) — the log ring only administrators
// read is not rotated by whoever polls the open status: the warning is
// written when the failure starts (UR-P4-leak-91).
func TestStatus_AFabricReadFailureIsLoggedOnce(t *testing.T) {
	var log bytes.Buffer
	n := sampleNode()
	n.fabricsErr = errors.New("store: closed")
	_, mux := newTestHandler(t, testOptions{node: n, log: &log, admin: func(*http.Request) bool { return false }, second: secondNode()})
	const calls = 50
	for range calls {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, DefaultPrefix+"/node/status", http.NoBody)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), `"admin":true`) {
			t.Fatalf("status %d %s", rec.Code, rec.Body)
		}
	}
	if got := strings.Count(log.String(), "fabricui.status.fabrics_failed"); got > 1 {
		t.Errorf("%d status requests without admin rights wrote %d warnings (fabricui.status.fabrics_failed); want one per failure, not one per request", calls, got)
	}
	// The end of the failure is logged once, and the next failure again.
	status := func() {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, DefaultPrefix+"/node/status", http.NoBody))
	}
	n.mu.Lock()
	n.fabricsErr = nil
	n.mu.Unlock()
	status()
	status()
	n.mu.Lock()
	n.fabricsErr = errors.New("store: closed")
	n.mu.Unlock()
	status()
	if got := strings.Count(log.String(), "fabricui.status.fabrics_recovered"); got != 1 {
		t.Errorf("recovered logged %d times, want 1", got)
	}
	if got := strings.Count(log.String(), "fabricui.status.fabrics_failed"); got != 2 {
		t.Errorf("failed logged %d times over two failures, want 2", got)
	}
}

// The same through the registry: an entry the handler skips is logged
// once, not on every GET /nodes (UR-P4-leak-91).
func TestRegistry_ASkippedEntryIsLoggedOnce(t *testing.T) {
	var log bytes.Buffer
	reg := newMemoryRegistry(0)
	if _, err := reg.Create(context.Background(), NodeSpec{Label: "ok"}); err != nil {
		t.Fatal(err)
	}
	reg.mu.Lock()
	reg.nodes = append(reg.nodes, NodeEntry{Key: "status", Label: "a page name"})
	reg.mu.Unlock()
	_, mux := newTestHandler(t, testOptions{registry: reg, log: &log, admin: func(*http.Request) bool { return false }})
	const calls = 50
	for range calls {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, DefaultPrefix+"/nodes", http.NoBody)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("nodes %d %s", rec.Code, rec.Body)
		}
	}
	if got := strings.Count(log.String(), "fabricui.registry.entry_skipped"); got > 1 {
		t.Errorf("%d GET /nodes without admin rights wrote %d warnings (fabricui.registry.entry_skipped); want one per entry, not one per request", calls, got)
	}
}
