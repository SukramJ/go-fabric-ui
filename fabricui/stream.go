// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
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
	if streamHead(w, r) {
		return
	}
	h.serveSSE(r.Context(), w, h.subscribe)
}

// streamHead answers a stream request that is not a GET — a HEAD, which
// the mux matches on a GET pattern — with the stream's headers and no
// body, and reports whether it did. net/http discards a HEAD answer's
// body without an error, so a stream served for HEAD would notice a
// client's leaving only through the request's context, which a client
// that sent one more byte never cancels (UR-P1-conc-01).
func streamHead(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet {
		return false
	}
	writeStreamHeaders(w)
	return true
}

// writeStreamHeaders writes the header of a server-sent event stream.
func writeStreamHeaders(w http.ResponseWriter) {
	hdr := w.Header()
	hdr.Set("Content-Type", "text/event-stream")
	hdr.Set("Cache-Control", "no-store")
	hdr.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
}

// serveSSE serves server-sent events from subscribe until ctx ends, with
// a comment line every heartbeat: the node stream and the controller
// stream (GET /controller/stream) alike.
func (h *handler) serveSSE(ctx context.Context, w http.ResponseWriter, subscribe func(func(Event)) func()) {
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
	defer subscribe(push)()

	writeStreamHeaders(w)
	if _, err := fmt.Fprint(w, ": connected\n\n"); err != nil {
		return
	}
	_ = rc.Flush()

	tick := time.NewTicker(h.heartbeat)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
		case ev := <-ch:
			// The type goes into the frame verbatim: one that is not
			// [a-z_]+ (a line break from a host's Options.Events) would
			// forge a frame, so it is dropped and logged without its
			// payload. Every type this module publishes passes.
			if !validEventType(ev.Type) {
				h.log.WarnContext(ctx, "fabricui.stream.invalid_type", "type", ev.Type)
				continue
			}
			data, err := json.Marshal(ev.Data)
			if err != nil {
				h.log.WarnContext(ctx, "fabricui.stream.encode_failed", "type", ev.Type, "error", err.Error())
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

// validEventType reports whether t is a non-empty [a-z_]+.
func validEventType(t string) bool {
	if t == "" {
		return false
	}
	for _, c := range []byte(t) {
		if (c < 'a' || c > 'z') && c != '_' {
			return false
		}
	}
	return true
}
