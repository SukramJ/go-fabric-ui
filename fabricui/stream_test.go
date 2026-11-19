// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
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

// A client that keeps its connection open and stops reading fills the
// socket buffers; the stream's next write blocks, the heartbeat with it.
// A write deadline per round ends such a stream, and its subscriptions
// go (UR-P2-bound-33).
func TestStream_StalledReaderIsReaped(t *testing.T) {
	ev := NewEvents()
	h, mux := newTestHandler(t, testOptions{events: ev})
	h.heartbeat = 50 * time.Millisecond
	srv := httptest.NewServer(mux)
	defer srv.Close()
	conn, err := net.Dial("tcp", srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if tc, ok := conn.(*net.TCPConn); ok {
		_ = tc.SetReadBuffer(4096)
	}
	fmt.Fprintf(conn, "GET %s/node/stream HTTP/1.1\r\nHost: x\r\n%s: %s\r\n\r\n", DefaultPrefix, TokenHeader, testToken)
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close(); _ = resp.Body.Close() }() // the conn first: Close drains the body
	// From here on the client reads nothing.
	big := strings.Repeat("x", 64<<10)
	for stop := time.Now().Add(time.Second); time.Now().Before(stop); {
		ev.Publish(Event{Type: "host_note", Data: big})
		time.Sleep(time.Millisecond)
	}
	waitFor(t, func() bool { return subsOf(ev) == 0 })
}

// The shutdown pattern docs/embedding.md gives the hosts: a base context
// cancelled in RegisterOnShutdown ends the open streams, so Shutdown does
// not wait its budget for them and their subscriptions go
// (UR-P2-conc-34).
func TestStream_EndsOnTheHostsShutdown(t *testing.T) {
	ev := NewEvents()
	_, mux := newTestHandler(t, testOptions{events: ev})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	base, stopStreams := context.WithCancel(context.Background())
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second, BaseContext: func(net.Listener) context.Context { return base }}
	srv.RegisterOnShutdown(stopStreams)
	go func() { _ = srv.Serve(ln) }()
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String()+DefaultPrefix+"/node/stream", http.NoBody)
	req.Header.Set(TokenHeader, testToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	waitFor(t, func() bool { return subsOf(ev) == 1 })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown with an open stream: %v", err)
	}
	waitFor(t, func() bool { return subsOf(ev) == 0 })
}

// A host's event data cannot break the stream's framing: a string with
// line breaks is escaped, raw JSON with line breaks is compacted, and
// invalid raw JSON drops the event (UR-P2-test-95; a guard, green).
func TestStream_DataStaysOnOneLine(t *testing.T) {
	src := &recordingSource{}
	h, mux := newTestHandler(t, testOptions{events: src})
	h.heartbeat = time.Hour
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/node/stream", http.NoBody)
	req.Header.Set(TokenHeader, testToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	waitFor(t, func() bool { return src.subscribers() == 1 })
	forged := "x\n\nevent: fabric_removed\ndata: {\"fabric_index\":1}\n\n\r\revent: y "
	src.publish(Event{Type: EventEndpointsChanged, Data: map[string]string{"k": forged}})
	src.publish(Event{Type: EventEndpointsChanged, Data: json.RawMessage("{\n\"a\":\n1\n}")})
	src.publish(Event{Type: EventEndpointsChanged, Data: json.RawMessage("{}\n\nevent: fabric_removed\ndata: {}")})
	src.publish(Event{Type: EventExposureChanged, Data: struct{}{}})
	var types []string
	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		line := sc.Text()
		if typ, ok := strings.CutPrefix(line, "event: "); ok {
			types = append(types, typ)
		}
		if data, ok := strings.CutPrefix(line, "data: "); ok && strings.ContainsRune(data, '\r') {
			t.Errorf("a data line holds a carriage return: %q", line)
		}
		if line == "event: "+EventExposureChanged {
			break
		}
	}
	if want := []string{EventEndpointsChanged, EventEndpointsChanged, EventExposureChanged}; !slices.Equal(types, want) {
		t.Errorf("event lines %v, want %v", types, want)
	}
}
