// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"net/http"

	"github.com/SukramJ/go-fabric/surface"
)

// surfaceReport serves GET /node/surface: the node's surface report in
// go-fabric's own JSON encoding (surface.Report.MarshalJSON, go-fabric
// surface/policy.go:400-432) — the effective surface per endpoint, the
// decision ledger, the withheld elements and the announced changes. The
// page reads the shape go-fabric writes; this handler adds nothing.
func (h *handler) surfaceReport(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	b, err := json.Marshal(h.node.surface())
	if err != nil {
		// A report with an element of no kind or no origin has no
		// encoding (surface/element.go:135-140).
		h.serverError(w, r, "Failed to encode the surface report", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(b, '\n'))
}

// announcedChanges is the node's announced surface changes before a
// reassembly; nil without a node.
func (h *handler) announcedChanges() []surface.SurfaceChange {
	if h.node == nil {
		return nil
	}
	return h.node.surface().Changed
}

// publishSurfaceChanged publishes surface_changed with the endpoint keys
// of the changes the node announced since before was read, and nothing
// when there are none. go-fabric has no hook for an announced change; the
// report's Changed list is the source (node/surface_stability.go:281-297).
func (h *handler) publishSurfaceChanged(before []surface.SurfaceChange) {
	if h.node == nil {
		return
	}
	if keys := newChangeKeys(before, h.node.surface().Changed); len(keys) > 0 {
		h.publish(Event{Type: EventSurfaceChanged, Data: SurfaceChangedEvent{Endpoints: keys}})
	}
}

// newChangeKeys returns the endpoint keys, each once and in order, of the
// changes in after that come behind the last change of before. The node
// appends each change and keeps only the last hundred
// (node/surface_stability.go:31-32, :282-289), so a count taken before
// the reassembly stops growing at the bound; the last change seen is the
// mark instead. A mark that is no longer in after (dropped by the bound)
// makes every change new.
func newChangeKeys(before, after []surface.SurfaceChange) []string {
	start := 0
	if n := len(before); n > 0 {
		last := &before[n-1]
		for i := len(after) - 1; i >= 0; i-- {
			if sameChange(last, &after[i]) {
				start = i + 1
				break
			}
		}
	}
	var keys []string
	seen := map[string]bool{}
	for i := start; i < len(after); i++ {
		if k := after[i].Endpoint; !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	return keys
}

func sameChange(a, b *surface.SurfaceChange) bool {
	return a.Endpoint == b.Endpoint && a.Before == b.Before && a.After == b.After && a.At.Equal(b.At)
}
