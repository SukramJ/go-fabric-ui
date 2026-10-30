// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/SukramJ/go-fabric/node"
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

// surface_changed comes from the node's observer, one per changed
// endpoint, whichever assembly made the change: a route's reassembly
// reaches each open stream once — the route publishes endpoints_changed
// alone — and a reassembly the host runs on its own reaches it too.
func TestSubscribe_SurfaceChangedFromTheObserver(t *testing.T) {
	n := sampleNode()
	assemble := func(context.Context) error {
		// The node reports the change on the assembly's goroutine, after
		// it recorded it (go-fabric node/observers.go:75-84).
		n.notify(func(o node.Observers) { _ = o.OnSurfaceChanged(surface.SurfaceChange{Endpoint: "lamp-1", At: testNow}) })
		return nil
	}
	h, mux := newTestHandler(t, testOptions{node: n, exposure: newMemoryProvider(), reassemble: assemble})
	rec := &recorder{}
	defer h.subscribe(rec.add)()

	if code := do(t, mux, http.MethodPost, "/node/maintenance/reassemble", "").Code; code != http.StatusNoContent {
		t.Fatalf("status %d", code)
	}
	do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`)
	_ = assemble(context.Background()) // the host's own reassembly

	want := []string{
		EventSurfaceChanged, EventEndpointsChanged,
		EventExposureChanged, EventSurfaceChanged, EventEndpointsChanged,
		EventSurfaceChanged,
	}
	if !slices.Equal(rec.types(), want) {
		t.Fatalf("events %v, want %v", rec.types(), want)
	}
	for _, i := range []int{0, 3, 5} {
		if ev, ok := rec.events[i].Data.(SurfaceChangedEvent); !ok || !slices.Equal(ev.Endpoints, []string{"lamp-1"}) {
			t.Fatalf("event %d payload %+v", i, rec.events[i].Data)
		}
	}
}
