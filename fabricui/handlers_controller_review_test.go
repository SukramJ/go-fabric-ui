// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

// nullSource hands out the memory controller's nodes wrapped in nullNode.
type nullSource struct {
	*MemoryController
	nulls   map[AttributePath]bool
	invokes *atomic.Int32
}

func (s *nullSource) Node(id uint64) (ControllerNode, bool) {
	n, ok := s.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	return nullNode{ControllerNode: n, nulls: s.nulls, invokes: s.invokes}, true
}

// nullNode answers JSON null for the paths in nulls with status 0, as
// go-fabric renders a TLV null the node reports (valuejson.go:82 at
// v0.22.2), and counts the commands it is asked to invoke.
type nullNode struct {
	ControllerNode
	nulls   map[AttributePath]bool
	invokes *atomic.Int32
}

func (n nullNode) Read(ctx context.Context, paths []AttributePath, filtered bool) ([]AttributeValue, error) {
	vals, err := n.ControllerNode.Read(ctx, paths, filtered)
	for i := range vals {
		if n.nulls[vals[i].Path] {
			vals[i].Value = json.RawMessage("null")
			vals[i].Status = 0
		}
	}
	return vals, err
}

func (n nullNode) Invoke(ctx context.Context, path CommandPath, payload json.RawMessage, timed time.Duration) (InvokeResult, error) {
	n.invokes.Add(1)
	return n.ControllerNode.Invoke(ctx, path, payload, timed)
}

// UR-P1-err-31: a node that reports null for CurrentFabricIndex, the ACL,
// the Fabrics list or a Binding list — none of them nullable — is a node
// error (502), never index 0 or an empty list; the own-fabric guard holds.
func TestController_NullForANonNullableAttributeIs502(t *testing.T) {
	const id = "0000000000000011"
	current := AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeCurrentFabricIndex}
	fabrics := AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeFabrics}
	acl := AttributePath{Endpoint: 0, Cluster: clusterAccessControl, Attribute: attributeACL}
	binding := AttributePath{Endpoint: 1, Cluster: clusterBinding, Attribute: attributeBinding}
	for _, c := range []struct {
		name         string
		null         AttributePath
		method, path string
	}{
		{"current fabric index, remove", current, http.MethodDelete, "/controller/nodes/" + id + "/fabrics/1"},
		{"current fabric index, list", current, http.MethodGet, "/controller/nodes/" + id + "/fabrics"},
		{"fabrics", fabrics, http.MethodGet, "/controller/nodes/" + id + "/fabrics"},
		{"acl", acl, http.MethodGet, "/controller/nodes/" + id + "/acl"},
		{"binding", binding, http.MethodGet, "/controller/nodes/" + id + "/bindings"},
	} {
		t.Run(c.name, func(t *testing.T) {
			var invokes atomic.Int32
			src := &nullSource{MemoryController: NewMemoryController(), nulls: map[AttributePath]bool{c.null: true}, invokes: &invokes}
			_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
			rec := do(t, mux, c.method, c.path, "")
			if rec.Code != http.StatusBadGateway {
				t.Fatalf("%s %s: %d %s, want 502", c.method, c.path, rec.Code, rec.Body)
			}
			if invokes.Load() != 0 {
				t.Fatalf("%d command(s) invoked", invokes.Load())
			}
		})
	}
}
