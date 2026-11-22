// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"net/http"
	"testing"

	"thingfabric.dev/go-fabric/node"
)

// window_kind carries the contract's names only: go-fabric's named kinds
// by their names, a kind the contract does not name (one a later
// go-fabric adds) as other, never go-fabric's own text such as
// "WindowKind(9)" (contract 0.12.0, UR-P4-ctr-196).
func TestStatus_WindowKindIsAContractName(t *testing.T) {
	for kind, want := range map[node.WindowKind]string{
		node.WindowOwn:                "own",
		node.WindowBasic:              "basic",
		node.WindowEnhanced:           "enhanced",
		node.WindowAdministratorBasic: "administrator_basic",
		node.WindowKind(9):            "other",
	} {
		n := sampleNode()
		n.open, n.kind, n.duration = true, kind, 180
		_, mux := newTestHandler(t, testOptions{node: n})
		rec := do(t, mux, http.MethodGet, "/node/status", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("%v: %d %s", kind, rec.Code, rec.Body)
		}
		var s struct {
			WindowKind string `json:"window_kind"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
			t.Fatalf("%v: %v %s", kind, err, rec.Body)
		}
		if s.WindowKind != want {
			t.Errorf("%v: window_kind %q, want %q", kind, s.WindowKind, want)
		}
	}
}
