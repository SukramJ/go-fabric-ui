// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// streamBuffer is how many events a slow stream client may fall behind
// before the oldest are dropped; a page reloads what an event names, so a
// dropped event costs a stale view until the next one.
const streamBuffer = 64

// stream serves GET /node/stream: server-sent events from the handler's
// own routes, the node's transitions (node.Observe) and the host's
// source, with a comment line every 15 s so a proxy keeps the connection
// open. It replaces loom's WebSocket topics
// (internal/north/rest/handlers/matter_events.go:62-82 at b40b289a).
func (h *handler) stream(w http.ResponseWriter, r *http.Request) {
	rc := http.NewResponseController(w)
	ch := make(chan Event, streamBuffer)
	push := func(ev Event) {
		select {
		case ch <- ev:
		default:
		}
	}
	// The node's observers are registered for this stream while it is
	// open (node.Observe) and removed when it closes.
	defer h.subscribe(push)()

	hdr := w.Header()
	hdr.Set("Content-Type", "text/event-stream")
	hdr.Set("Cache-Control", "no-store")
	hdr.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
		return
	}
	_ = rc.Flush()

	tick := time.NewTicker(h.heartbeat)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
		case ev := <-ch:
			data, err := json.Marshal(ev.Data)
			if err != nil {
				h.log.WarnContext(r.Context(), "fabricui.stream.encode_failed", "type", ev.Type, "error", err.Error())
				continue
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, data); err != nil {
				return
			}
		}
		if err := rc.Flush(); err != nil {
			return
		}
	}
}
