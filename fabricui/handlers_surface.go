// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"net/http"
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
