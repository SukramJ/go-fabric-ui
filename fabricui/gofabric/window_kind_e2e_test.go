// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"fmt"
	"net/http"
	"testing"

	"thingfabric.dev/go-fabric/node"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// An administrator's basic window — OpenBasicCommissioningWindow from a
// paired controller, on a node that serves the Basic feature — reads
// window_kind administrator_basic on the node's status, as go-fabric
// v0.23.0 reports it (node.WindowAdministratorBasic; decision 30: v0.22.2
// said enhanced). basic stays the host's window on fresh credentials.
func TestEndToEnd_AnAdministratorsBasicWindowReadsAdministratorBasic(t *testing.T) {
	s := newCtrlSiteWith(t, ctrlSiteOptions{configure: func(cfg *node.Config) { cfg.Commissioning.BasicCommissioning = true }})
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var n struct {
		Node struct {
			NodeIDHex string `json:"node_id_hex"`
		} `json:"node"`
	}
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.Manual), http.StatusOK, &n)
	s.waitEvent(fabricui.EventSubscription)
	var res struct {
		Status int `json:"status"`
	}
	// AdministratorCommissioning (0x003C) OpenBasicCommissioningWindow (0x01).
	s.json(http.MethodPost, "/controller/nodes/"+n.Node.NodeIDHex+"/commands",
		`{"endpoint":0,"cluster_hex":"0x003C","command_hex":"0x01","payload":{"commissioningTimeout":180}}`, http.StatusOK, &res)
	if res.Status != 0 {
		t.Fatalf("OpenBasicCommissioningWindow status %#x", res.Status)
	}
	var st struct {
		Open bool   `json:"commissioning_window_open"`
		Kind string `json:"window_kind"`
	}
	s.json(http.MethodGet, "/node/status", "", http.StatusOK, &st)
	if !st.Open || st.Kind != "administrator_basic" {
		t.Fatalf("after an administrator's basic window: open %v, kind %q", st.Open, st.Kind)
	}
	// The window runs on the node's configured passcode and
	// discriminator, so the configured codes of GET /node/pairing
	// commission through it: window_match is true (UR-P4-ctr-193).
	var p struct {
		Source string `json:"source"`
		Open   bool   `json:"window_open"`
		Match  *bool  `json:"window_match"`
	}
	s.json(http.MethodGet, "/node/pairing", "", http.StatusOK, &p)
	if !p.Open || p.Source != "configured" || p.Match == nil || !*p.Match {
		t.Errorf("GET /node/pairing under administrator_basic: source %q open %v window_match %v, want true", p.Source, p.Open, p.Match != nil && *p.Match)
	}
}
