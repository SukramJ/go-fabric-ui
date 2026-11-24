// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"thingfabric.dev/go-fabric/fabrictest"
	"thingfabric.dev/go-fabric/node"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

type inventoryEntry struct {
	NodeIDHex  string  `json:"node_id_hex"`
	HasCode    bool    `json:"has_code"`
	QRCode     *string `json:"qr_code"`
	ManualCode *string `json:"manual_code"`
	Source     string  `json:"source"`
	Checks     *struct {
		Vendor, Product, Discriminator bool
	} `json:"checks"`
}

func (s *ctrlSite) inventory() []inventoryEntry {
	s.t.Helper()
	var inv struct {
		Entries []inventoryEntry `json:"entries"`
	}
	s.json(http.MethodGet, "/controller/inventory", "", http.StatusOK, &inv)
	return inv.Entries
}

func (s *ctrlSite) nodeIDs() []string {
	s.t.Helper()
	var list struct {
		Nodes []struct {
			NodeIDHex string `json:"node_id_hex"`
			HasCode   *bool  `json:"has_onboarding_code"`
		} `json:"nodes"`
	}
	s.json(http.MethodGet, "/controller/nodes", "", http.StatusOK, &list)
	ids := make([]string, 0, len(list.Nodes))
	for _, n := range list.Nodes {
		if n.HasCode == nil {
			s.t.Fatalf("node %s without has_onboarding_code", n.NodeIDHex)
		}
		ids = append(ids, n.NodeIDHex)
	}
	return ids
}

// G5 of brief U12 (contract 0.13.0) on go-fabric's controller with
// KeepOnboardingCodes: a node paired by its QR code is listed paired
// with both forms; another product's code answers
// onboarding_code_mismatch with field product and both ids, nothing
// stored; a text that is no code onboarding_code_invalid with go-fabric's
// reason and without the text; the inventory is in the order of GET
// /controller/nodes; a removed code and a decommissioned node are gone.
func TestEndToEnd_OnboardingInventory(t *testing.T) {
	s := newCtrlSiteWith(t, ctrlSiteOptions{onboarding: true})
	var st struct {
		Capabilities struct {
			OnboardingCodes bool `json:"onboarding_codes"`
		} `json:"capabilities"`
	}
	s.json(http.MethodGet, "/controller/status", "", http.StatusOK, &st)
	if !st.Capabilities.OnboardingCodes {
		t.Fatal("capabilities.onboarding_codes is false")
	}
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var n struct {
		Node struct {
			NodeIDHex string `json:"node_id_hex"`
		} `json:"node"`
	}
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.QR), http.StatusOK, &n)
	first := n.Node.NodeIDHex
	inv := s.inventory()
	if len(inv) != 1 || inv[0].NodeIDHex != first || !inv[0].HasCode || inv[0].Source != "paired" ||
		inv[0].QRCode == nil || *inv[0].QRCode != sp.QR || inv[0].ManualCode == nil || *inv[0].ManualCode != strings.ReplaceAll(sp.Manual, "-", "") {
		t.Fatalf("after pairing by QR code: %+v (QR %s, manual %s)", inv, sp.QR, sp.Manual)
	}

	// Another product of the same vendor, on another discriminator.
	other := s.site.AddDevice(fabrictest.Topology{Configure: func(cfg *node.Config) {
		cfg.ProductDescription.ProductID = 0x8002
		cfg.Commissioning.Discriminator = 1234
	}})
	osp, err := other.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var mm struct {
		Code      string `json:"code"`
		Field     string `json:"field"`
		CodeValue int    `json:"code_value"`
		NodeValue int    `json:"node_value"`
	}
	s.json(http.MethodPut, "/controller/nodes/"+first+"/onboarding-code", fmt.Sprintf(`{"code":%q}`, osp.QR), http.StatusConflict, &mm)
	if mm.Code != "onboarding_code_mismatch" || mm.Field != "product" || mm.CodeValue != 0x8002 || mm.NodeValue != 0x8001 {
		t.Fatalf("mismatch %+v", mm)
	}
	if inv := s.inventory(); inv[0].QRCode == nil || *inv[0].QRCode != sp.QR || inv[0].Source != "paired" {
		t.Fatalf("a refused code was stored: %+v", inv)
	}

	const notACode = "MT:NOT-A-CODE-4711"
	var bad struct {
		Code   string `json:"code"`
		Detail string `json:"detail"`
	}
	s.json(http.MethodPut, "/controller/nodes/"+first+"/onboarding-code", fmt.Sprintf(`{"code":%q}`, notACode), http.StatusBadRequest, &bad)
	if bad.Code != "onboarding_code_invalid" || bad.Detail == "" || strings.Contains(bad.Detail, "NOT-A-CODE") || strings.Contains(bad.Detail, "4711") {
		t.Fatalf("invalid %+v", bad)
	}

	// The second device, commissioned by its manual code: two nodes in
	// the order of GET /controller/nodes.
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, osp.Manual), http.StatusOK, &n)
	second := n.Node.NodeIDHex
	inv = s.inventory()
	invIDs := make([]string, 0, len(inv))
	for _, e := range inv {
		invIDs = append(invIDs, e.NodeIDHex)
	}
	if ids := s.nodeIDs(); !slices.Equal(ids, invIDs) || len(ids) != 2 {
		t.Fatalf("inventory order %v, nodes %v", invIDs, ids)
	}
	for _, e := range inv {
		if e.NodeIDHex == second && (e.QRCode != nil || e.ManualCode == nil || e.Source != "paired") {
			t.Fatalf("a node paired by its manual code: %+v", e)
		}
	}

	// Attaching the second device's QR code to it: its passcode stays
	// unproven, its identity checks pass.
	var e inventoryEntry
	s.json(http.MethodPut, "/controller/nodes/"+second+"/onboarding-code", fmt.Sprintf(`{"code":%q}`, osp.QR), http.StatusOK, &e)
	if e.Source != "attached" || e.QRCode == nil || *e.QRCode != osp.QR || e.Checks == nil || !e.Checks.Vendor || !e.Checks.Product {
		t.Fatalf("attached %+v", e)
	}
	s.json(http.MethodDelete, "/controller/nodes/"+second+"/onboarding-code", "", http.StatusNoContent, nil)
	for _, e := range s.inventory() {
		if e.NodeIDHex == second && e.HasCode {
			t.Fatalf("a removed code: %+v", e)
		}
	}

	s.json(http.MethodDelete, "/controller/nodes/"+first, "", http.StatusNoContent, nil)
	s.waitEvent(fabricui.EventNodeRemoved)
	for _, e := range s.inventory() {
		if e.NodeIDHex == first {
			t.Fatalf("a decommissioned node is listed: %+v", e)
		}
	}
	s.json(http.MethodPut, "/controller/nodes/"+first+"/onboarding-code", fmt.Sprintf(`{"code":%q}`, sp.QR), http.StatusNotFound, nil)
}
