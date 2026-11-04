// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/SukramJ/go-fabric/contract"
	"github.com/SukramJ/go-fabric/node"
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
