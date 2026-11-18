// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
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

// UR-P1-adpt-34: only a CASE subject names a node; a Group entry whose
// group id equals the node id is kept.
func TestSetNodeACL_KeepsAGroupSubjectEqualToTheNodeID(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	body := `{"entries":[{"privilege":"administer","auth_mode":"case","subjects_hex":["0000000000000020"],"targets":[]},` +
		`{"privilege":"operate","auth_mode":"group","subjects_hex":["0000000000000011"],"targets":[]}]}`
	if r := do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/acl", body); r.Code != http.StatusOK {
		t.Fatalf("acl %d %s", r.Code, r.Body)
	}
	r := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/acl", "")
	if !strings.Contains(r.Body.String(), `"auth_mode":"group"`) {
		t.Fatalf("the group entry is gone: %s", r.Body)
	}
}

// UR-P1-ctr-32: an ACL entry whose privilege or auth mode is outside the
// contract's enums is a node error, not "" in the answer.
func TestGetNodeACL_UnknownPrivilegeOrAuthModeIs502(t *testing.T) {
	for _, entry := range []string{
		`[{"privilege":6,"authMode":2,"subjects":[32],"targets":null,"fabricIndex":1}]`,
		`[{"privilege":5,"authMode":4,"subjects":[32],"targets":null,"fabricIndex":1}]`,
	} {
		src := &rawSource{
			MemoryController: NewMemoryController(),
			raw:              map[AttributePath]json.RawMessage{{Endpoint: 0, Cluster: clusterAccessControl, Attribute: attributeACL}: json.RawMessage(entry)},
		}
		_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
		r := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/acl", "")
		if r.Code != http.StatusBadGateway {
			t.Fatalf("%s: %d %s, want 502", entry, r.Code, r.Body)
		}
	}
}

// rawSource hands out the memory controller's nodes wrapped in rawNode.
type rawSource struct {
	*MemoryController
	raw map[AttributePath]json.RawMessage
}

func (s *rawSource) Node(id uint64) (ControllerNode, bool) {
	n, ok := s.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	return rawNode{ControllerNode: n, raw: s.raw}, true
}

// rawNode answers the given JSON for the paths in raw.
type rawNode struct {
	ControllerNode
	raw map[AttributePath]json.RawMessage
}

func (n rawNode) Read(ctx context.Context, paths []AttributePath, filtered bool) ([]AttributeValue, error) {
	vals, err := n.ControllerNode.Read(ctx, paths, filtered)
	for i := range vals {
		if v, ok := n.raw[vals[i].Path]; ok {
			vals[i].Value, vals[i].Status = v, 0
		}
	}
	return vals, err
}
