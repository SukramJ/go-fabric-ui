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
// fixtures come from — two device nodes (the second, a node of the memory
// registry, with a selection reason; the registry creates and deletes
// further ones, contract 0.6.0), the default node's configuration block,
// the memory selection, the memory controller with the network graph of
// the topology fixture and the operations ring — until the process is
// interrupted. The default node reports a refused attestation and the boot
// reason 5, and its erase removes its fabrics (contract 0.7.0); the
// second node is clean. The controller has a Bluetooth LE central
// (contract 0.7.1). It is not a test of its own.
func TestServeE2EWeb(t *testing.T) {
	addr := os.Getenv(e2eWebAddr)
	if addr == "" {
		t.Skip(e2eWebAddr + " is not set: the browser smoke tests start this server")
	}
	ops := sampleOperations()
	ctrl := NewMemoryController()
	topo := controllerTopology()
	ctrl.Topology = &topo
	// A host with a central (contract 0.7.1): the Commission page
	// preselects the default credential and states the BLE decision.
	ctrl.BLE = true
	// The second node is registry state without a selection, as the
	// selector's and the second node's smokes expect it.
	reg := newMemoryRegistry(0)
	second := secondNode()
	reg.nodes = []NodeEntry{{
		Key: "second", Label: "Second bridge", Kind: "bridge", Source: portSource{fakeNode: second},
		Config: &NodeConfigView{Listen: second.addr, Discriminator: 3541, VendorID: 0xFFF1, VendorName: "go-fabric", ProductID: 0x8000, ProductName: "Bridge", Store: "bridge-second.db"},
	}}
	def := sampleNode()
	def.attestationErr = errors.New("node: attestation files: open /data/dac.pem: no such file or directory")
	def.boot = 5
	erase := func(context.Context) error {
		def.mu.Lock()
		def.fabricList = nil
		def.mu.Unlock()
		return nil
	}
	_, mux := newTestHandler(t, testOptions{
		node: def, exposure: newMemoryProvider(), registry: reg, nodeConfig: defaultNodeConfig(),
		controller: ctrl, operations: ops, erase: erase,
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
