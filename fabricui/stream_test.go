// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func subsOf(ev *Events) int {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	return len(ev.subs)
}

func logObserversOf(m *MemoryOperations) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.obs)
}

// UR-P1-conc-01: a HEAD on a stream route answers the headers and ends.
// net/http discards a HEAD answer's body without an error, so a stream
// served for HEAD never noticed a client that had sent one more byte
// before it left: its subscription stayed for the life of the process.
func TestStream_HEADDoesNotSubscribe(t *testing.T) {
	ops := sampleOperations()
	h, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), operations: ops})
	h.heartbeat = 20 * time.Millisecond
	srv := httptest.NewServer(mux)
	// A leaked stream keeps its connection active, and Close waits for
	// it: close in the background so the failure is reported, not hung.
	defer func() { srv.CloseClientConnections(); go srv.Close() }()

	cases := []struct {
		path  string
		count func() int
	}{
		{"/node/stream", func() int { return subsOf(h.hub) }},
		{"/controller/stream", func() int { return subsOf(h.ctrlHub) }},
		{"/operation/log/stream", func() int { return logObserversOf(ops) }},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			conn, err := net.Dial("tcp", srv.Listener.Addr().String())
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(conn, "HEAD %s%s HTTP/1.1\r\nHost: x\r\n\r\n", DefaultPrefix, c.path)
			time.Sleep(100 * time.Millisecond)
			fmt.Fprint(conn, "X")
			time.Sleep(100 * time.Millisecond)
			_ = conn.Close()
			deadline := time.Now().Add(time.Second)
			for c.count() != 0 {
				if time.Now().After(deadline) {
					t.Fatalf("HEAD %s: %d subscriber(s) a second after the client left, want 0", c.path, c.count())
				}
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}

// plainWriter is a ResponseWriter that cannot flush: a host middleware
// that wraps the writer without Unwrap.
type plainWriter struct{ rec *httptest.ResponseRecorder }

func (p plainWriter) Header() http.Header         { return p.rec.Header() }
func (p plainWriter) Write(b []byte) (int, error) { return p.rec.Write(b) }
func (p plainWriter) WriteHeader(code int)        { p.rec.WriteHeader(code) }

// UR-P1-bound-14 (a): a writer that cannot flush ends the stream at its
// first heartbeat; that is logged, not silent.
func TestStream_UnflushableWriterIsLogged(t *testing.T) {
	var log bytes.Buffer
	h, _ := newTestHandler(t, testOptions{log: &log})
	h.heartbeat = 5 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	h.serveSSE(ctx, plainWriter{httptest.NewRecorder()}, nil, func(func(Event)) func() { return func() {} })
	if !strings.Contains(log.String(), "level=WARN") || !strings.Contains(log.String(), "fabricui.stream.flush_failed") {
		t.Fatalf("no warn record for a writer that cannot flush; log:\n%s", log.String())
	}
}

// UR-P1-bound-14 (b): an invalid event type is logged bounded, not as
// the host handed it.
func TestStream_InvalidTypeIsLoggedBounded(t *testing.T) {
	var log bytes.Buffer
	h, _ := newTestHandler(t, testOptions{log: &log})
	h.heartbeat = time.Hour
	ctx, cancel := context.WithCancel(context.Background())
	long := strings.Repeat("A", 10000)
	h.serveSSE(ctx, httptest.NewRecorder(), nil, func(fn func(Event)) func() {
		fn(Event{Type: long})
		go func() { time.Sleep(50 * time.Millisecond); cancel() }()
		return func() {}
	})
	if !strings.Contains(log.String(), "fabricui.stream.invalid_type") {
		t.Fatalf("no invalid_type record; log:\n%s", log.String())
	}
	if n := log.Len(); n > 1000 {
		t.Fatalf("the invalid_type record is %d bytes long, want the type truncated", n)
	}
}
