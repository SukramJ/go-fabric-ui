// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui_test

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/SukramJ/go-fabric/device"
	"github.com/SukramJ/go-fabric/fabrictest"
	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

type logBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *logBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *logBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// A request without admin rights that ends its sending side right after
// the request (or is aborted behind a proxy) has a cancelled context; a
// node's SQLite store answers the status's fabric read with "context
// canceled". That is an abandoned request, not a state of the node: the
// status logs no fabrics_failed for it and no fabrics_recovered after it,
// as serverError treats an abandoned request (UR-P1-err-12,
// UR-P4-leak-200). The count is not fixed (how many requests the cancel
// reaches in time varies); one pair at most may log.
func TestStatus_AnAbandonedRequestIsNotAFabricFailure(t *testing.T) {
	site := fabrictest.NewSite(t, fabrictest.Options{StoreOnDisk: true})
	plug, err := device.NewOnOffPlugInUnit(device.Options{StableKey: device.StringKey("plug:1"), Name: "Plug"})
	if err != nil {
		t.Fatal(err)
	}
	dev := site.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{plug}})
	var log logBuf
	srv := httptest.NewServer(fabricui.Handler(fabricui.Options{
		Node:   dev.Node(),
		Logger: slog.New(slog.NewTextHandler(&log, nil)),
		Admin:  func(*http.Request) bool { return false },
	}))
	t.Cleanup(srv.Close)
	addr := srv.Listener.Addr().String()

	halfClosed := func() string {
		c, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if _, err := io.WriteString(c, "GET /api/v1/node/status HTTP/1.1\r\nHost: "+addr+"\r\nConnection: close\r\n\r\n"); err != nil {
			t.Fatal(err)
		}
		_ = c.(*net.TCPConn).CloseWrite()
		b, _ := io.ReadAll(c)
		return string(b)
	}
	plain := func() {
		resp, err := srv.Client().Get(srv.URL + "/api/v1/node/status")
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
	const pairs = 25
	missing := 0
	for range pairs {
		if !strings.Contains(halfClosed(), `"fabric_count"`) {
			missing++
		}
		plain()
	}
	failed := strings.Count(log.String(), "fabricui.status.fabrics_failed")
	recovered := strings.Count(log.String(), "fabricui.status.fabrics_recovered")
	if failed > 1 || recovered > 1 {
		i := strings.Index(log.String(), "fabricui.status.fabrics_failed")
		line := log.String()[i:]
		if j := strings.IndexByte(line, '\n'); j >= 0 {
			line = line[:j]
		}
		t.Errorf("%d pairs of status requests without admin rights (one half-closed, one plain) wrote %d fabrics_failed and %d fabrics_recovered lines; %d answers lacked fabric_count; first: %s",
			pairs, failed, recovered, missing, line)
	}
}
