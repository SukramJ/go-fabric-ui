// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui_test

import (
	"net/http"
	"testing"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// On a node without a fabric the close of a window on fresh credentials
// answers 204, and go-fabric then reopens the node's own window on its
// configured codes — also after the operator had closed that window. The
// handler does not change this (decision 73); the contract's closeWindow
// and ConfiguredWindow's godoc say it, and a second close closes the
// node's own window (UR-P4-auth-90).
func TestEndToEnd_CloseOfAFreshWindowReopensTheOwnWindowWithoutAFabric(t *testing.T) {
	s := newSite(t, false)
	s.openStream()
	s.reopen("basic")
	s.expect(http.StatusNoContent, http.MethodPost, "/node/pairing/window/close", "")
	s.waitEvent(fabricui.EventWindowOpened)
	if st := s.status(); !st.WindowOpen || st.WindowKind != "own" {
		t.Fatalf("after the close of the fresh window: %+v, want the node's own window", st)
	}
	var p struct {
		Source string `json:"source"`
		Open   bool   `json:"window_open"`
	}
	s.getJSON("/node/pairing", &p)
	if p.Source != "configured" || !p.Open {
		t.Errorf("GET /node/pairing: %+v, want the configured codes in an open window", p)
	}
	// The second close closes the node's own window.
	s.expect(http.StatusNoContent, http.MethodPost, "/node/pairing/window/close", "")
	s.waitEvent(fabricui.EventWindowClosed)
	if st := s.status(); st.WindowOpen {
		t.Errorf("after the second close: %+v", st)
	}
}
