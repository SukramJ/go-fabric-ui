// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"
	"net"
	"net/http/httptest"
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
