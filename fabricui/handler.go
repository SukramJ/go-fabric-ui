// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// TokenHeader carries the admin token in token mode (a nil
// [Options].Admin).
const TokenHeader = "X-Fabric-UI-Token" //nolint:gosec // a header name, not a credential

// handler holds what every route reads.
type handler struct {
	node     nodePort
	exposure ExposureProvider
	// ctrl is the controller family's source; nil serves no controller
	// routes. ctrlHub carries the handler's own controller events
	// (commissioning progress, update).
	ctrl    ControllerSource
	ctrlHub *Events
	admin   func(*http.Request) bool
	// token is the admin token of token mode; empty when Options.Admin is
	// set.
	token      string
	reassemble func(context.Context) error
	warnings   func(context.Context) []Warning
	// hub carries the handler's own events; extra is the host's source
	// when it is not an *Events. The node's transitions come from
	// node.Observe, per stream (subscribe).
	hub        *Events
	extra      EventSource
	prefix     string
	log        *slog.Logger
	configured bool
	allowEmpty bool
	heartbeat  time.Duration
	// fresh draws the credentials of a window on fresh credentials:
	// node.NewWindowCredentials.
	fresh func(time.Duration) (node.WindowCredentials, error)

	// mu guards ephemeral: the credentials of the window this handler
	// opened, served by GET /node/pairing while that window is open.
	mu        sync.Mutex
	ephemeral *ephemeralWindow
}

func newHandler(o Options) (*handler, error) {
	prefix := o.Prefix
	if prefix == "" {
		prefix = DefaultPrefix
	}
	if err := validPrefix(prefix); err != nil {
		return nil, err
	}
	h := &handler{
		exposure:   o.Exposure,
		ctrl:       o.Controller,
		ctrlHub:    NewEvents(),
		admin:      o.Admin,
		reassemble: o.Reassemble,
		warnings:   o.Warnings,
		prefix:     prefix,
		log:        o.Logger,
		configured: o.ConfiguredWindow,
		allowEmpty: o.AllowEmptyTopology,
		heartbeat:  15 * time.Second,
		fresh:      node.NewWindowCredentials,
	}
	if h.log == nil {
		h.log = slog.Default()
	}
	if o.Node != nil {
		h.node = nodeAdapter{n: o.Node}
	}
	if ev, ok := o.Events.(*Events); ok && ev != nil {
		h.hub = ev
	} else {
		h.hub = NewEvents()
		h.extra = o.Events
	}
	if h.admin == nil {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		h.token = hex.EncodeToString(b)
	}
	return h, nil
}

// register puts every route on mux. Route table: docs/contract.md and
// contract/openapi.yaml; loom's block it replaces is
// internal/north/rest/router.go:1328-1359 at b40b289a.
func (h *handler) register(mux *http.ServeMux) {
	p := h.prefix + "/node"
	read := func(pattern string, fn http.HandlerFunc) { mux.HandleFunc(pattern, fn) }
	write := func(pattern string, fn http.HandlerFunc) { mux.HandleFunc(pattern, h.adminOnly(fn)) }

	read("GET "+p+"/status", h.status)
	read("GET "+p+"/fabrics", h.fabrics)
	write("DELETE "+p+"/fabrics/{index}", h.removeFabric)
	write("GET "+p+"/pairing", h.pairing)
	write("POST "+p+"/pairing/window", h.openWindow)
	write("POST "+p+"/pairing/window/close", h.closeWindow)
	read("GET "+p+"/endpoints", h.endpoints)
	read("GET "+p+"/sessions", h.sessions)
	read("GET "+p+"/mdns", h.mdns)
	read("GET "+p+"/compatibility", h.compatibility)
	read("GET "+p+"/events", h.diagnosticEvents)
	read("GET "+p+"/exposure/schema", h.exposureSchema)
	read("GET "+p+"/exposure", h.exposureList)
	write("PUT "+p+"/exposure", h.exposureApply)
	write("POST "+p+"/exposure/preview", h.exposurePreview)
	write("POST "+p+"/maintenance/reassemble", h.maintenanceReassemble)
	write("POST "+p+"/maintenance/factory-reset", h.factoryReset)
	read("GET "+p+"/stream", h.stream)
	if h.ctrl != nil {
		h.mountController(mux)
	}
	// Anything else under the prefix is an API 404, not the web
	// interface's index.
	mux.HandleFunc(h.prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
	})
}

// adminOnly refuses a request the host's Admin does not admit, or, in
// token mode, one without the token. Mirrors loom's admin-gated route
// group (router.go:1336-1353, `pr.With(admin)`).
func (h *handler) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.isAdmin(r) {
			writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required",
				"this route writes or reads a pairing secret")
			return
		}
		next(w, r)
	}
}

func (h *handler) isAdmin(r *http.Request) bool {
	if h.admin != nil {
		return h.admin(r)
	}
	got := r.Header.Get(TokenHeader)
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(h.token)) == 1
}

// publish hands ev to the stream's subscribers.
func (h *handler) publish(ev Event) { h.hub.Publish(ev) }

// subscribe registers fn for every event from now on — the handler's own,
// the host's source and the node's transitions (node.Observe) — and
// returns the function that unregisters it from all three. A route that
// removes a fabric or opens a window does not publish the transition
// itself: the node reports it through the observer, once.
func (h *handler) subscribe(fn func(Event)) (unsubscribe func()) {
	unsubs := []func(){h.hub.Subscribe(fn)}
	if h.extra != nil {
		unsubs = append(unsubs, h.extra.Subscribe(fn))
	}
	if h.node != nil {
		unsubs = append(unsubs, h.node.observe(nodeObservers(fn, h.node.window().Open)))
	}
	return func() {
		for _, u := range unsubs {
			u()
		}
	}
}

// nodeUnready answers a node route while the host wired no node.
func nodeUnready(w http.ResponseWriter, r *http.Request) {
	unready(w, r, "Node not available", "the host serves no node (Options.Node is nil)")
}
