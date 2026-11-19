// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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

// bindingNode serves a Binding cluster on endpoints 1..count, answers a
// status for the endpoints in refused and records the size of each read.
type bindingNode struct {
	ControllerNode
	count   int
	refused map[uint16]uint8
	reads   *[]int
}

func (n bindingNode) Endpoints() []NodeEndpoint {
	eps := make([]NodeEndpoint, 0, n.count)
	for i := 1; i <= n.count; i++ {
		eps = append(eps, NodeEndpoint{Number: uint16(i), Clusters: []NodeCluster{{ID: clusterBinding}}})
	}
	return eps
}

func (n bindingNode) Read(_ context.Context, paths []AttributePath, _ bool) ([]AttributeValue, error) {
	*n.reads = append(*n.reads, len(paths))
	out := make([]AttributeValue, 0, len(paths))
	for _, p := range paths {
		if st, ok := n.refused[p.Endpoint]; ok {
			out = append(out, AttributeValue{Path: p, Status: st})
			continue
		}
		out = append(out, AttributeValue{Path: p, Value: json.RawMessage("[]")})
	}
	return out, nil
}

type bindingSource struct {
	*MemoryController
	node bindingNode
}

func (s *bindingSource) Node(id uint64) (ControllerNode, bool) {
	n, ok := s.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	b := s.node
	b.ControllerNode = n
	return b, true
}

// UR-P1-bound-37: the Binding read is chunked at maxReadPaths, and an
// endpoint the node refuses is a node error, not left out of an answer
// the contract says holds every Binding endpoint.
func TestGetNodeBindings_ChunkedAndRefusedCarriesItsStatus(t *testing.T) {
	var reads []int
	src := &bindingSource{MemoryController: NewMemoryController(), node: bindingNode{count: 2*maxReadPaths + 1, reads: &reads}}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
	r := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/bindings", "")
	if r.Code != http.StatusOK || strings.Count(r.Body.String(), `"endpoint":`) != 2*maxReadPaths+1 {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
	for _, n := range reads {
		if n > maxReadPaths {
			t.Fatalf("reads of %v paths, want at most %d each", reads, maxReadPaths)
		}
	}

	src = &bindingSource{MemoryController: NewMemoryController(), node: bindingNode{count: 2, refused: map[uint16]uint8{2: 0x7E}, reads: &reads}}
	_, mux = newTestHandler(t, testOptions{noNode: true, controller: src})
	// A refused endpoint is listed with its status and no bindings
	// (contract 0.10.0, UR-P1-bound-37).
	r = do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/bindings", "")
	if r.Code != http.StatusOK || !strings.Contains(r.Body.String(), `{"bindings":[],"endpoint":2,"status":126}`) {
		t.Fatalf("a refused endpoint: %d %s", r.Code, r.Body)
	}
}

// UR-P1-err-10: a long discriminator is 12 bits; a value outside 0..4095
// is refused, not narrowed to another discriminator.
func TestCommission_LongDiscriminatorOutOfRangeIs400(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	for _, v := range []string{"69632", "4096", "-1"} {
		body := `{"discovered":{"instance_name":"x","long_discriminator":` + v + `},"passcode":20202021}`
		if r := do(t, mux, http.MethodPost, "/controller/commission", body); r.Code != http.StatusBadRequest {
			t.Fatalf("long_discriminator %s: %d %s, want 400", v, r.Code, r.Body)
		}
	}
}

// cancelNode answers every read with context.Canceled, as the adapter
// passes a cancelled request's error through.
type cancelNode struct{ ControllerNode }

func (cancelNode) Read(context.Context, []AttributePath, bool) ([]AttributeValue, error) {
	return nil, context.Canceled
}

type cancelSource struct{ *MemoryController }

func (s cancelSource) Node(id uint64) (ControllerNode, bool) {
	n, ok := s.MemoryController.Node(id)
	return cancelNode{n}, ok
}

// UR-P1-err-12: a request its client abandoned is no server error: it is
// not logged at error level.
func TestController_AbandonedRequestIsNotLoggedAsError(t *testing.T) {
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: cancelSource{NewMemoryController()}, log: &log})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://x"+DefaultPrefix+"/controller/nodes/0000000000000011/acl", http.NoBody)
	mux.ServeHTTP(httptest.NewRecorder(), req)
	if strings.Contains(log.String(), "level=ERROR") {
		t.Fatalf("an abandoned request logged at error level:\n%s", log.String())
	}
}

// undecodedSource hands out the memory controller's nodes wrapped in
// undecodedNode.
type undecodedSource struct{ *MemoryController }

func (s undecodedSource) Node(id uint64) (ControllerNode, bool) {
	n, ok := s.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	return undecodedNode{n}, true
}

// undecodedNode answers RemoveFabric with a NOCResponse the adapter could
// not decode: HasResponse with no fields (gofabric/node.go Invoke).
type undecodedNode struct{ ControllerNode }

func (n undecodedNode) Invoke(ctx context.Context, path CommandPath, payload json.RawMessage, timed time.Duration) (InvokeResult, error) {
	if path.Cluster == clusterOperationalCredentials && path.Command == commandRemoveFabric {
		return InvokeResult{HasResponse: true, Command: 0x08}, nil
	}
	return n.ControllerNode.Invoke(ctx, path, payload, timed)
}

// A RemoveFabric whose NOCResponse does not decode is no success: its
// status is unknown, so the answer is 502, not 204 (UR-P2-adpt-36).
func TestRemoveNodeFabric_UndecodedNOCResponseIs502(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: undecodedSource{NewMemoryController()}})
	rec := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/fabrics", "")
	var fl struct {
		Fabrics []struct {
			FabricIndex int  `json:"fabric_index"`
			Own         bool `json:"own"`
		} `json:"fabrics"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &fl); err != nil {
		t.Fatal(err)
	}
	other := 0
	for _, f := range fl.Fabrics {
		if !f.Own {
			other = f.FabricIndex
		}
	}
	if other == 0 {
		t.Fatalf("no fabric but the controller's own: %s", rec.Body)
	}
	if rec := do(t, mux, http.MethodDelete, "/controller/nodes/0000000000000011/fabrics/"+strconv.Itoa(other), ""); rec.Code != http.StatusBadGateway {
		t.Fatalf("remove: %d %s, want 502", rec.Code, rec.Body)
	}
}

// slowDecommission is the memory controller whose Decommission takes a
// while and stops when its context ends.
type slowDecommission struct {
	*MemoryController
	started chan struct{}
	got     error
}

func (s *slowDecommission) Decommission(ctx context.Context, id uint64) error {
	close(s.started)
	select {
	case <-ctx.Done():
		s.got = ctx.Err()
		return s.got
	case <-time.After(200 * time.Millisecond):
		return s.MemoryController.Decommission(ctx, id)
	}
}

// A decommission runs detached from its request, as the registry's
// Delete: a client that leaves between the node's RemoveFabric and its
// deletion would leave an uncommissioned record (UR-P2-err-31).
func TestDecommissionNode_OutlivesTheRequest(t *testing.T) {
	src := &slowDecommission{MemoryController: NewMemoryController(), started: make(chan struct{})}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
	leaving(t, mux, http.MethodDelete, "/controller/nodes/0000000000000011", "", src.started)
	if src.got != nil {
		t.Fatalf("the decommission's context ended with the request: %v", src.got)
	}
}
