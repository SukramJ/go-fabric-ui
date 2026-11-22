// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"
	"net/http"
	"testing"

	"thingfabric.dev/go-fabric/node"
)

// A fabric another admin removed between the list and the removal is
// go-fabric's node.ErrFabricNotFound: 404, not 500 (UR-P1-err-33).
func TestRemoveFabric_GoneMeanwhileIs404(t *testing.T) {
	n := sampleNode()
	n.removeErr = fmt.Errorf("remove: %w", node.ErrFabricNotFound)
	_, mux := newTestHandler(t, testOptions{node: n})
	if rec := do(t, mux, http.MethodDelete, "/node/fabrics/1", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("status %d %s, want 404", rec.Code, rec.Body)
	}
}

// The controller route: a node that answers RemoveFabric with NOCResponse
// InvalidFabricIndex holds no such fabric: 404 not_found, as the contract
// says from 0.11.0 (UR-P1-err-33).
func TestRemoveNodeFabric_UnknownIndexIs404(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController()})
	rec := do(t, mux, http.MethodDelete, "/controller/nodes/0000000000000011/fabrics/7", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d %s, want 404", rec.Code, rec.Body)
	}
}
