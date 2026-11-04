// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// sampleOperations is the operations double of the fixtures: a ring of
// three that took four records (the first fell out), at level info.
func sampleOperations() *MemoryOperations {
	ops := NewMemoryOperations(3, OperationInfo{
		Version: "0.5.0", GoFabric: "v0.16.1-0.20261104133901-894adbcd4cd3", StartedAt: testNow.Add(-90 * time.Minute),
	})
	ops.Now = func() time.Time { return testNow }
	ops.Record(testNow.Add(-4*time.Second), slog.LevelInfo, "node started", map[string]string{"port": "5540"})
	ops.Record(testNow.Add(-3*time.Second), slog.LevelWarn, "mdns: only a link-local address", nil)
	ops.Record(testNow.Add(-2*time.Second), slog.LevelDebug, "dropped below the level", nil)
	ops.Record(testNow.Add(-2*time.Second), slog.LevelInfo, "session opened", map[string]string{"peer.node": "000000000001B669", "fabric": "1"})
	ops.Record(testNow.Add(-time.Second), slog.LevelError, "subscription failed", map[string]string{"error": "timeout"})
	return ops
}

func TestOperation_Fixtures(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{operations: sampleOperations()})
	for _, name := range []string{
		"operation-info.200",
		"operation-log.200",
		"operation-log.400",
		"operation-log-level.200",
		"operation-log-level-set.400",
		"operation-log-level-set.403",
		"operation-log-level-set.200",
	} {
		runFixture(t, mux, name)
	}
	_, mux = newTestHandler(t, testOptions{})
	runFixture(t, mux, "operation-info.503")
}

// Every operation route answers 503 without the port.
func TestOperation_NoPort_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/operation/info", ""},
		{http.MethodGet, "/operation/log/level", ""},
		{http.MethodPut, "/operation/log/level", `{"level":"debug"}`},
		{http.MethodGet, "/operation/log", ""},
		{http.MethodGet, "/operation/log/stream", ""},
	} {
		rec := do(t, mux, c.method, c.path, c.body)
		if rec.Code != http.StatusServiceUnavailable || problemOf(t, rec).Code != codeServiceUnready {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
}

func TestOperation_LogWindow(t *testing.T) {
	ops := sampleOperations()
	_, mux := newTestHandler(t, testOptions{operations: ops})
	for _, c := range []struct {
		query string
		code  int
		count int
	}{
		{"", http.StatusOK, 3},
		{"?limit=1", http.StatusOK, 1},
		{"?since=4", http.StatusOK, 1},
		{"?since=5", http.StatusOK, 0},
		{"?limit=2001", http.StatusBadRequest, 0},
		{"?since=-1", http.StatusBadRequest, 0},
		{"?since=x", http.StatusBadRequest, 0},
	} {
		rec := do(t, mux, http.MethodGet, "/operation/log"+c.query, "")
		if rec.Code != c.code {
			t.Fatalf("%q: %d %s", c.query, rec.Code, rec.Body)
		}
		if c.code == http.StatusOK && strings.Count(rec.Body.String(), `"seq"`) != c.count {
			t.Fatalf("%q: %s", c.query, rec.Body)
		}
	}
}

func TestOperation_LogStream(t *testing.T) {
	ops := sampleOperations()
	h, mux := newTestHandler(t, testOptions{operations: ops})
	h.heartbeat = time.Hour
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/operation/log/stream", http.NoBody)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
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
	waitFor(t, func() bool { return ops.observers() == 1 })
	ops.Record(testNow, slog.LevelWarn, "window opened", map[string]string{"discriminator": "1234"})
	if a, b := next(), next(); a != "event: log_record" ||
		b != `data: {"seq":5,"time":"2026-10-25T12:00:00Z","level":"warn","msg":"window opened","attrs":{"discriminator":"1234"}}` {
		t.Fatalf("event lines %q %q", a, b)
	}
	cancel()
	waitFor(t, func() bool { return ops.observers() == 0 })
}

func TestMemoryOperations(t *testing.T) {
	ops := NewMemoryOperations(0, OperationInfo{})
	if ops.capacity != DefaultMemoryLogCapacity || ops.LogLevel() != slog.LevelInfo {
		t.Fatalf("defaults: %d %v", ops.capacity, ops.LogLevel())
	}
	if _, ok := ops.Record(testNow, slog.LevelDebug, "below", nil); ok {
		t.Fatal("a record below the level was taken")
	}
	_ = ops.SetLogLevel(slog.LevelDebug)
	rec, ok := ops.Record(testNow, slog.LevelDebug, "now taken", map[string]string{"k": "v"})
	if !ok || rec.Seq != 1 {
		t.Fatalf("record %+v %v", rec, ok)
	}
	got := ops.Log(0, 0)
	got[0].Attrs["k"] = "changed"
	if ops.Log(0, 0)[0].Attrs["k"] != "v" {
		t.Fatal("Log hands out the ring's own attributes")
	}
	if in := ops.Info(); in.Uptime != 0 {
		t.Fatalf("uptime without a start: %v", in.Uptime)
	}
}

func TestLevelNames(t *testing.T) {
	for l, want := range map[slog.Level]string{
		slog.LevelDebug - 4: "debug", slog.LevelDebug: "debug", slog.LevelInfo - 1: "debug",
		slog.LevelInfo: "info", slog.LevelInfo + 2: "info", slog.LevelWarn: "warn", slog.LevelError: "error", slog.LevelError + 8: "error",
	} {
		if got := levelName(l); got != want {
			t.Errorf("levelName(%v) = %q, want %q", l, got, want)
		}
	}
	if _, ok := parseLevel("WARN"); ok {
		t.Error("the level names are lower-case")
	}
}

// The test binary is built in module mode, so its build information names
// the go-fabric module go.mod requires.
func TestGoFabricVersion(t *testing.T) {
	if v := GoFabricVersion(); !strings.HasPrefix(v, "v0.") {
		t.Fatalf("GoFabricVersion() = %q", v)
	}
}

// observers is how many ObserveLog callbacks are registered.
func (m *MemoryOperations) observers() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.obs)
}
