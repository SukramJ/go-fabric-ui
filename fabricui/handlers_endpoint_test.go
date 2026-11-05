// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/SukramJ/go-fabric/contract"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/schema"
)

// The node's endpoint detail: the lamp's values, its scope and the surface
// report's decisions on its key (lamp-1).
func TestNodeEndpoint_Fixtures(t *testing.T) {
	for _, name := range []string{"node-endpoint.200", "node-endpoint.404", "node-endpoint-events.200"} {
		t.Run(name, func(t *testing.T) {
			_, mux := newTestHandler(t, testOptions{})
			runFixture(t, mux, name)
		})
	}
}

func TestNodeEndpoint_DecisionsByKey(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	var d struct {
		Scope     string `json:"scope"`
		Decisions []struct {
			Endpoint string `json:"endpoint"`
		} `json:"decisions"`
	}
	rec := do(t, mux, http.MethodGet, "/node/endpoints/2", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Scope != "default" || len(d.Decisions) != 3 {
		t.Fatalf("scope %q, %d decisions; body %s", d.Scope, len(d.Decisions), rec.Body)
	}
	for _, x := range d.Decisions {
		if x.Endpoint != "lamp-1" {
			t.Fatalf("decision of %q on the lamp", x.Endpoint)
		}
	}
}

// The root's decisions are the report's under "", the Aggregator's under
// node.AggregatorSurfaceKey; an endpoint the topology does not key has
// none.
func TestSurfaceKeyOf(t *testing.T) {
	topo := sampleNode().topo
	for _, c := range []struct {
		ep   uint16
		key  string
		keyd bool
	}{
		{0, "", true}, {1, node.AggregatorSurfaceKey, true}, {2, "lamp-1", true}, {4, "sensor-2/humidity", true}, {9, "", false},
	} {
		if key, ok := surfaceKeyOf(topo, c.ep); key != c.key || ok != c.keyd {
			t.Errorf("endpoint %d: %q %v, want %q %v", c.ep, key, ok, c.key, c.keyd)
		}
	}
	h, _ := newTestHandler(t, testOptions{})
	if d, err := h.decisionsOf("", false); err != nil || len(d) != 0 {
		t.Fatalf("unkeyed: %v %v", d, err)
	}
	if d, err := h.decisionsOf("", true); err != nil || len(d) != 1 {
		t.Fatalf("root: %v %v", d, err)
	}
}

func TestNodeEndpoint_Errors(t *testing.T) {
	for _, c := range []struct {
		name string
		err  error
		path string
		want int
	}{
		{"not started", node.ErrNotStarted, "/node/endpoints/2", http.StatusServiceUnavailable},
		{"read failure", errors.New("boom"), "/node/endpoints/2", http.StatusInternalServerError},
		{"bad number", nil, "/node/endpoints/70000", http.StatusBadRequest},
		{"not a number", nil, "/node/endpoints/x", http.StatusBadRequest},
		{"events bad number", nil, "/node/endpoints/x/events", http.StatusBadRequest},
		{"events bad since", nil, "/node/endpoints/2/events?since=x", http.StatusBadRequest},
		{"events negative since", nil, "/node/endpoints/2/events?since=-1", http.StatusBadRequest},
		{"events bad limit", nil, "/node/endpoints/2/events?limit=x", http.StatusBadRequest},
		{"events zero limit", nil, "/node/endpoints/2/events?limit=0", http.StatusBadRequest},
		{"events limit above 1000", nil, "/node/endpoints/2/events?limit=1001", http.StatusBadRequest},
	} {
		t.Run(c.name, func(t *testing.T) {
			n := sampleNode()
			n.detailErr = c.err
			_, mux := newTestHandler(t, testOptions{node: n})
			if rec := do(t, mux, http.MethodGet, c.path, ""); rec.Code != c.want {
				t.Fatalf("status %d, want %d: %s", rec.Code, c.want, rec.Body)
			}
		})
	}
	_, mux := newTestHandler(t, testOptions{noNode: true})
	for _, p := range []string{"/node/endpoints/2", "/node/endpoints/2/events"} {
		if rec := do(t, mux, http.MethodGet, p, ""); rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s without a node: %d", p, rec.Code)
		}
	}
}

// since and limit reach Node.Events as MinNumber and Limit; absent, the
// limit is 100.
func TestNodeEndpointEvents_Query(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	var got struct {
		Events []struct {
			Number int `json:"number"`
		} `json:"events"`
	}
	rec := do(t, mux, http.MethodGet, "/node/endpoints/2/events?since=8&limit=5", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got.Events) != 1 || got.Events[0].Number != 9 {
		t.Fatalf("since 8: %s", rec.Body)
	}
	ep := uint16(2)
	if want := (node.EventQuery{Endpoint: &ep, MinNumber: 8, Limit: 5}); !reflect.DeepEqual(n.query, want) {
		t.Fatalf("query %+v, want %+v", n.query, want)
	}
	do(t, mux, http.MethodGet, "/node/endpoints/2/events", "")
	if n.query.Limit != defaultEventLimit || n.query.MinNumber != 0 {
		t.Fatalf("default query %+v", n.query)
	}
}

// attribute_changed carries the ids of one cluster's changed attributes in
// hex, never a value.
func TestNodeObservers_AttributeChanged(t *testing.T) {
	var got []Event
	o := NodeObservers(func(e Event) { got = append(got, e) }, false)
	o.OnAttributeChanged(node.AttributeChange{Endpoint: 2, Cluster: 0x0006, Change: contract.Change{Attributes: []uint32{0x0000, 0x4003}}})
	want := []Event{{Type: EventAttributeChanged, Data: AttributeChangedEvent{Endpoint: 2, ClusterHex: "0x0006", Attributes: []string{"0x0000", "0x4003"}}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	b, _ := json.Marshal(got[0].Data)
	if string(b) != `{"endpoint":2,"cluster_hex":"0x0006","attributes":["0x0000","0x4003"]}` {
		t.Fatalf("payload %s", b)
	}
}

// The priorities' names (Matter Core §10.6.6.1).
func TestEventPriorityName(t *testing.T) {
	for p, want := range map[uint8]string{0: "debug", 1: "info", 2: "critical", 7: "debug"} {
		if got := EventPriorityName(p); got != want {
			t.Errorf("%d: %q, want %q", p, got, want)
		}
	}
}

// treeSource is a NodeSource with sampleNode's endpoint tree and topology:
// every endpoint serves a Descriptor whose ServerList names the Descriptor
// alone; an endpoint outside served answers UNSUPPORTED_ENDPOINT.
type treeSource struct {
	stubSource
	infos  []node.EndpointInfo
	served map[uint16]bool
}

func newTreeSource() *treeSource {
	s := &treeSource{infos: sampleInfos(), served: map[uint16]bool{}}
	for i := range s.infos {
		s.served[s.infos[i].Endpoint] = true
	}
	return s
}

func (s *treeSource) Endpoints() []node.EndpointInfo { return s.infos }
func (s *treeSource) Topology() node.BridgeTopology  { return sampleNode().topo }

func (s *treeSource) ReadCluster(_ context.Context, ep uint16, cluster uint32) ([]node.AttributeValue, error) {
	if !s.served[ep] {
		return nil, contract.StatusError{Status: statusUnsupportedEndpoint}
	}
	return []node.AttributeValue{
		{Endpoint: ep, Cluster: cluster, Attribute: attrServerList, Value: []uint32{clusterDescriptor}, DataVersion: 3},
	}, nil
}

// The endpoint tree comes from Node.Endpoints: the root and the Aggregator
// carry no parent (the Aggregator's is the root), a bridged device the
// Aggregator, a part its device; the parts are the direct children, not
// the full-family PartsList; Reachable from the topology.
func TestReadEndpointDetail_TreeFromEndpoints(t *testing.T) {
	src := newTreeSource()
	agg, sensor := uint16(1), uint16(3)
	for _, c := range []struct {
		ep        uint16
		name      string
		parent    *uint16
		parts     []uint16
		reachable bool
		scope     string
	}{
		{0, "RootNode", nil, []uint16{1}, true, ""},
		{1, "Aggregator", nil, []uint16{2, 3}, true, ""},
		{2, "Kitchen lamp", &agg, []uint16{}, true, "default"},
		{3, "Hall sensor", &agg, []uint16{4}, false, "default"},
		{4, "Hall humidity", &sensor, []uint16{}, true, "default"},
	} {
		d, err := readEndpointDetail(context.Background(), src, c.ep)
		if err != nil {
			t.Fatalf("endpoint %d: %v", c.ep, err)
		}
		if d.Name != c.name || !reflect.DeepEqual(d.Parent, c.parent) || !reflect.DeepEqual(d.Parts, c.parts) ||
			d.Reachable != c.reachable || d.Scope != c.scope {
			t.Errorf("endpoint %d: %+v", c.ep, d)
		}
		if want := sampleInfos()[c.ep].DeviceTypes[0]; len(d.DeviceTypes) != 1 || d.DeviceTypes[0].ID != want.ID || d.DeviceTypes[0].Revision != want.Revision {
			t.Errorf("endpoint %d device types %+v", c.ep, d.DeviceTypes)
		}
		if len(d.Clusters) != 1 || d.Clusters[0].ID != clusterDescriptor {
			t.Errorf("endpoint %d clusters %+v", c.ep, d.Clusters)
		}
	}
	if _, err := readEndpointDetail(context.Background(), src, 9); !errors.Is(err, ErrEndpointNotFound) {
		t.Fatalf("an endpoint the node does not serve: %v", err)
	}
	// Served at the Descriptor read, gone from the tree: not found.
	src.served[5] = true
	if _, err := readEndpointDetail(context.Background(), src, 5); !errors.Is(err, ErrEndpointNotFound) {
		t.Fatalf("an endpoint outside the tree: %v", err)
	}
}

// The list route renders the root and the Aggregator from Node.Endpoints.
func TestStructuralEndpoints(t *testing.T) {
	got := structuralEndpoints(sampleInfos(), flattenTopology(sampleNode().topo))
	if len(got) != 2 {
		t.Fatalf("structural %+v", got)
	}
	root, agg := got[0], got[1]
	if root.EndpointID != 0 || root.ParentEndpointID != nil || root.Name != "RootNode" || !reflect.DeepEqual(root.Parts, []uint16{1}) ||
		len(root.DeviceTypes) != 1 || root.DeviceTypes[0].IDHex != "0x0016" || root.DeviceTypes[0].Revision != 5 {
		t.Errorf("root %+v", root)
	}
	if agg.EndpointID != 1 || agg.ParentEndpointID != nil || agg.Name != "Aggregator" || !reflect.DeepEqual(agg.Parts, []uint16{2, 3}) {
		t.Errorf("aggregator %+v", agg)
	}
}

// A failed attribute — of the AttributeList or not — is listed with its
// status and no value; the DataVersion is a value's, not the failed
// attribute's 0; with every attribute failed the revision is the schema's.
func TestServedCluster_FailedAttributes(t *testing.T) {
	const onOff = 0x0006
	vals := []node.AttributeValue{
		{Cluster: onOff, Attribute: 0x0000, Status: 0x86},
		{Cluster: onOff, Attribute: attrAttributeList, Value: []uint32{0x0000, 0xFFFB, 0xFFFC, 0xFFFD}, DataVersion: 7},
		{Cluster: onOff, Attribute: attrFeatureMap, Value: uint32(1), DataVersion: 7},
		{Cluster: onOff, Attribute: attrClusterRevision, Value: uint16(6), DataVersion: 7},
		{Cluster: onOff, Attribute: 0x4000, Status: 0x8F},
	}
	c := servedCluster(onOff, vals)
	if c.DataVersion == nil || *c.DataVersion != 7 || c.Revision != 6 || c.FeatureMap != 1 {
		t.Fatalf("cluster %+v", c)
	}
	ids := make([]uint32, 0, len(c.Attributes))
	for _, a := range c.Attributes {
		ids = append(ids, a.ID)
	}
	if !reflect.DeepEqual(ids, []uint32{0x0000, 0x4000, 0xFFFB, 0xFFFC, 0xFFFD}) {
		t.Fatalf("attributes %v", ids)
	}
	if a := c.Attributes[0]; a.Status != 0x86 || a.Value != nil || a.Name != "OnOff" {
		t.Errorf("failed OnOff %+v", a)
	}
	if a := c.Attributes[1]; a.Status != 0x8F || a.Value != nil {
		t.Errorf("failed 0x4000 %+v", a)
	}
	if a := c.Attributes[3]; a.Status != 0 || string(a.Value) != "1" {
		t.Errorf("FeatureMap %+v", a)
	}

	all := servedCluster(onOff, []node.AttributeValue{{Cluster: onOff, Attribute: 0x4000, Status: 0x8F}, {Cluster: onOff, Attribute: 0x0000, Status: 0x01}})
	rev, _ := schema.ClusterRevision(onOff)
	if all.DataVersion != nil || all.Revision != rev || len(all.Attributes) != 2 || all.Attributes[0].ID != 0 || all.Attributes[0].Status != 0x01 {
		t.Fatalf("every attribute failed %+v", all)
	}
}

// The handler renders a failed attribute with status and no value, a
// value without status.
func TestNodeEndpoint_FailedAttributeStatus(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	var d struct {
		Clusters []struct {
			IDHex      string `json:"id_hex"`
			Attributes []map[string]json.RawMessage
		} `json:"clusters"`
	}
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/endpoints/2", "").Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	var seen int
	for _, c := range d.Clusters {
		for _, a := range c.Attributes {
			_, hasValue := a["value"]
			st, hasStatus := a["status"]
			if hasStatus == hasValue {
				t.Errorf("%s %s: status and value both or neither", c.IDHex, a["id_hex"])
			}
			if hasStatus {
				seen++
				if string(st) != "1" {
					t.Errorf("status %s", st)
				}
			}
		}
	}
	if seen != 1 {
		t.Fatalf("%d failed attributes, want 1", seen)
	}
}
