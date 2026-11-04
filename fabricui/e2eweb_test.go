// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// e2eWebAddr names the address TestServeE2EWeb listens on; unset, the test
// is skipped. make e2e-web sets it through ui/playwright.config.ts.
const e2eWebAddr = "FABRIC_UI_E2E_ADDR"

// TestServeE2EWeb is the server of the browser smoke tests (ui/e2e,
// make e2e-web): the committed bundle and the API on the test doubles the
// fixtures come from — two device nodes (the second with a selection
// reason), the memory selection, the memory controller and the operations
// ring — until the process is interrupted. It is not a test of its own.
func TestServeE2EWeb(t *testing.T) {
	addr := os.Getenv(e2eWebAddr)
	if addr == "" {
		t.Skip(e2eWebAddr + " is not set: the browser smoke tests start this server")
	}
	ops := sampleOperations()
	_, mux := newTestHandler(t, testOptions{
		node: sampleNode(), second: secondNode(), exposure: newMemoryProvider(),
		controller: NewMemoryController(), operations: ops,
		exposureReason: "This host builds its endpoints without a selection.",
	})
	mux.Handle("/", spaHandler())
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	// A record now and then, so the log stream has something to append.
	go func() {
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-tick.C:
				ops.Record(now, slog.LevelInfo, "heartbeat of the e2e server", nil)
			}
		}
	}()
	t.Logf("serving on http://%s", ln.Addr())
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}
