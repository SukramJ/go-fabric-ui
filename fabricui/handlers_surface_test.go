// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/surface"
)

// The report travels in go-fabric's own encoding: the fixture is
// surface.Report.MarshalJSON of sampleSurface.
func TestSurface_ServesTheReport(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-surface.200")
}

func TestSurface_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-surface.503")
}

// A report go-fabric cannot encode (an element of no kind) is a 500, not
// a truncated 200.
func TestSurface_UnencodableReport_Returns500(t *testing.T) {
	n := sampleNode()
	n.rep.Withheld = append(n.rep.Withheld, surface.Element{Endpoint: "x"})
	_, mux := newTestHandler(t, testOptions{node: n})
	if rec := do(t, mux, http.MethodGet, "/node/surface", ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}

// A reassembly that announces surface changes publishes surface_changed
// after endpoints_changed, with each changed endpoint's key once; one
// that announces none publishes no surface_changed.
func TestReassemble_PublishesSurfaceChanged(t *testing.T) {
	n := sampleNode()
	announce := func(context.Context) error {
		for _, key := range []string{"sensor-2", "lamp-1", "sensor-2"} {
			n.rep.Changed = append(n.rep.Changed, surface.SurfaceChange{Endpoint: key, At: testNow.Add(time.Duration(len(n.rep.Changed)) * time.Second)})
		}
		return nil
	}
	h, mux := newTestHandler(t, testOptions{node: n, reassemble: announce})
	rec := record(h)
	if code := do(t, mux, http.MethodPost, "/node/maintenance/reassemble", "").Code; code != http.StatusNoContent {
		t.Fatalf("status %d", code)
	}
	if !slices.Equal(rec.types(), []string{EventEndpointsChanged, EventSurfaceChanged}) {
		t.Fatalf("events %v", rec.types())
	}
	if ev, ok := rec.events[1].Data.(SurfaceChangedEvent); !ok || !slices.Equal(ev.Endpoints, []string{"sensor-2", "lamp-1"}) {
		t.Fatalf("payload %+v", rec.events[1].Data)
	}

	h2, mux2 := newTestHandler(t, testOptions{node: sampleNode(), reassemble: func(context.Context) error { return nil }})
	rec2 := record(h2)
	do(t, mux2, http.MethodPost, "/node/maintenance/reassemble", "")
	if !slices.Equal(rec2.types(), []string{EventEndpointsChanged}) {
		t.Fatalf("no change announced: events %v", rec2.types())
	}
}

// Applying a selection reassembles, and the reassembly's announced
// changes reach the stream the same way.
func TestExposureApply_PublishesSurfaceChanged(t *testing.T) {
	n := sampleNode()
	announce := func(context.Context) error {
		n.rep.Changed = append(n.rep.Changed, surface.SurfaceChange{Endpoint: "lamp-1", At: testNow})
		return nil
	}
	h, mux := newTestHandler(t, testOptions{node: n, exposure: newMemoryProvider(), reassemble: announce})
	rec := record(h)
	do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`)
	if !slices.Equal(rec.types(), []string{EventExposureChanged, EventEndpointsChanged, EventSurfaceChanged}) {
		t.Fatalf("events %v", rec.types())
	}
	if ev, ok := rec.events[2].Data.(SurfaceChangedEvent); !ok || !slices.Equal(ev.Endpoints, []string{"lamp-1"}) {
		t.Fatalf("payload %+v", rec.events[2].Data)
	}
}

// The node keeps the last hundred changes, so the list's length stops
// growing; the new changes are those behind the last one seen, and all
// of them when that one was dropped.
func TestNewChangeKeys_ReadsBehindTheLastChangeSeen(t *testing.T) {
	at := func(i int) surface.SurfaceChange {
		return surface.SurfaceChange{Endpoint: string(rune('a' + i)), At: testNow.Add(time.Duration(i) * time.Second)}
	}
	before := []surface.SurfaceChange{at(0), at(1), at(2)}
	cases := []struct {
		name  string
		after []surface.SurfaceChange
		want  []string
	}{
		{"none new", []surface.SurfaceChange{at(0), at(1), at(2)}, nil},
		{"appended", []surface.SurfaceChange{at(0), at(1), at(2), at(3)}, []string{"d"}},
		{"bounded", []surface.SurfaceChange{at(1), at(2), at(3), at(4)}, []string{"d", "e"}},
		{"mark dropped", []surface.SurfaceChange{at(3), at(4)}, []string{"d", "e"}},
	}
	for _, c := range cases {
		if got := newChangeKeys(before, c.after); !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
	if got := newChangeKeys(nil, before); !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Errorf("nothing before: %v", got)
	}
}
