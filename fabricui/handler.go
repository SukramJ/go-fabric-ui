// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
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
	// ops is the daemon's operation (Options.Operations); nil answers 503
	// on the /operation/* routes.
	ops            Operations
	exposureReason string
	// nodes are the node family's nodes as the options list them, the
	// default first; served is that list as register mounted it, keyed
	// the handler of each by key (the default's is h).
	nodes  []nodeEntry
	served []nodeEntry
	keyed  map[string]*handler

	// mu guards ephemeral: the credentials of the window this handler
	// opened, served by GET /node/pairing while that window is open.
	mu        sync.Mutex
	ephemeral *ephemeralWindow
}

// nodeEntry is one node of the family in the handler's terms.
type nodeEntry struct {
	key, label, kind string
	// port is nil for a node listed without a source.
	port nodePort
}

// nodeKeyPattern is the contract's NodeKey pattern.
var nodeKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// reservedNodeKeys are the node family's page names: the shell's hash
// route #/node/<key>/<page> cannot tell a key from one of them.
var reservedNodeKeys = []string{"status", "pair", "fabrics", "selection", "endpoints", "endpoint", "surface", "diagnostics", "topology"}

// nodeEntries are o's nodes, the default first, validated.
func nodeEntries(o *Options) ([]nodeEntry, error) {
	in := o.Nodes
	if o.Node != nil {
		in = append([]NodeEntry{{Key: DefaultNodeKey, Source: o.Node}}, in...)
	}
	out := make([]nodeEntry, 0, len(in))
	seen := map[string]bool{}
	for _, e := range in {
		switch {
		case !nodeKeyPattern.MatchString(e.Key):
			return nil, fmt.Errorf("fabricui: node key %q does not match %s", e.Key, nodeKeyPattern)
		case slices.Contains(reservedNodeKeys, e.Key):
			return nil, fmt.Errorf("fabricui: node key %q is a page name of the node family", e.Key)
		case seen[e.Key]:
			return nil, fmt.Errorf("fabricui: node key %q is listed twice", e.Key)
		}
		seen[e.Key] = true
		ne := nodeEntry{key: e.Key, label: e.Label, kind: e.Kind}
		if e.Source != nil {
			ne.port = nodeAdapter{n: e.Source}
		}
		out = append(out, ne)
	}
	return out, nil
}

func newHandler(o Options) (*handler, error) {
	prefix := o.Prefix
	if prefix == "" {
		prefix = DefaultPrefix
	}
	if err := validPrefix(prefix); err != nil {
		return nil, err
	}
	nodes, err := nodeEntries(&o)
	if err != nil {
		return nil, err
	}
	h := &handler{
		ops:            o.Operations,
		exposureReason: o.ExposureReason,
		nodes:          nodes,
		exposure:       o.Exposure,
		ctrl:           o.Controller,
		ctrlHub:        NewEvents(),
		admin:          o.Admin,
		reassemble:     o.Reassemble,
		warnings:       o.Warnings,
		prefix:         prefix,
		log:            o.Logger,
		configured:     o.ConfiguredWindow,
		allowEmpty:     o.AllowEmptyTopology,
		heartbeat:      15 * time.Second,
		fresh:          node.NewWindowCredentials,
	}
	if h.log == nil {
		h.log = slog.Default()
	}
	if len(nodes) > 0 {
		h.node = nodes[0].port
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

// nodeRoute is one route of the node family: served under /node for the
// default node and under /nodes/{node} for every node by its key.
type nodeRoute struct {
	method, path string
	// admin routes write or read a pairing secret (contract x-admin).
	admin bool
	serve func(*handler, http.ResponseWriter, *http.Request)
}

// nodeRoutes is the node family's route table: docs/contract.md and
// contract/openapi.yaml; loom's block it replaces is
// internal/north/rest/router.go:1328-1359 at b40b289a.
var nodeRoutes = []nodeRoute{
	{"GET", "/status", false, (*handler).status},
	{"GET", "/fabrics", false, (*handler).fabrics},
	{"DELETE", "/fabrics/{index}", true, (*handler).removeFabric},
	{"GET", "/pairing", true, (*handler).pairing},
	{"POST", "/pairing/window", true, (*handler).openWindow},
	{"POST", "/pairing/window/close", true, (*handler).closeWindow},
	{"GET", "/endpoints", false, (*handler).endpoints},
	{"GET", "/endpoints/{ep}", false, (*handler).nodeEndpoint},
	{"GET", "/endpoints/{ep}/events", false, (*handler).nodeEndpointEvents},
	{"GET", "/surface", false, (*handler).surfaceReport},
	{"GET", "/sessions", false, (*handler).sessions},
	{"GET", "/topology", false, (*handler).topology},
	{"GET", "/mdns", false, (*handler).mdns},
	{"GET", "/compatibility", false, (*handler).compatibility},
	{"GET", "/events", false, (*handler).diagnosticEvents},
	{"GET", "/exposure/schema", false, (*handler).exposureSchema},
	{"GET", "/exposure", false, (*handler).exposureList},
	{"PUT", "/exposure", true, (*handler).exposureApply},
	{"POST", "/exposure/preview", true, (*handler).exposurePreview},
	{"POST", "/maintenance/reassemble", true, (*handler).maintenanceReassemble},
	{"POST", "/maintenance/factory-reset", true, (*handler).factoryReset},
	{"GET", "/stream", false, (*handler).stream},
}

// operationAdmin are the /operation routes the contract marks x-admin.
var operationAdmin = map[string]bool{"PUT /operation/log/level": true}

// register puts every route on mux: the node family under /node for the
// default node and under /nodes/{node} for each node, GET /nodes, the
// operation routes and the controller family.
func (h *handler) register(mux *http.ServeMux) {
	h.mountNodes()
	read := func(pattern string, fn http.HandlerFunc) { mux.HandleFunc(pattern, fn) }
	write := func(pattern string, fn http.HandlerFunc) { mux.HandleFunc(pattern, h.adminOnly(fn)) }

	for _, rt := range nodeRoutes {
		gate := read
		if rt.admin {
			gate = write
		}
		gate(rt.method+" "+h.prefix+"/node"+rt.path, func(w http.ResponseWriter, r *http.Request) { rt.serve(h, w, r) })
		gate(rt.method+" "+h.prefix+"/nodes/{node}"+rt.path, func(w http.ResponseWriter, r *http.Request) {
			n, ok := h.keyed[r.PathValue("node")]
			if !ok {
				writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+r.PathValue("node"))
				return
			}
			rt.serve(n, w, r)
		})
	}
	read("GET "+h.prefix+"/nodes", h.listNodes)
	read("GET "+h.prefix+"/operation/info", h.operationInfo)
	read("GET "+h.prefix+"/operation/log/level", h.logLevel)
	mux.HandleFunc("PUT "+h.prefix+"/operation/log/level", h.adminOnlyBecause(h.setLogLevel, "this route changes the daemon's log level"))
	read("GET "+h.prefix+"/operation/log", h.logRecords)
	read("GET "+h.prefix+"/operation/log/stream", h.logStream)
	if h.ctrl != nil {
		h.mountController(mux)
	}
	// Anything else under the prefix is an API 404, not the web
	// interface's index.
	mux.HandleFunc(h.prefix+"/", func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
	})
}

// mountNodes builds the handler of each node: h itself for the default —
// h.node, which a test may have set after newHandler — and a sibling for
// every other, which shares h's admin check, token and options and has a
// hub, a pairing state and no selection of its own.
func (h *handler) mountNodes() {
	served := slices.Clone(h.nodes)
	switch {
	case len(served) > 0:
		served[0].port = h.node
	case h.node != nil:
		served = []nodeEntry{{key: DefaultNodeKey, port: h.node}}
	}
	h.served = served
	h.keyed = make(map[string]*handler, len(served))
	for i, e := range served {
		if i == 0 {
			h.keyed[e.key] = h
			continue
		}
		h.keyed[e.key] = h.sibling(e.port)
	}
}

// sibling is the handler of another node: h's admin check, token, host
// warnings, logger and window options on node p.
func (h *handler) sibling(p nodePort) *handler {
	return &handler{
		node:           p,
		admin:          h.admin,
		token:          h.token,
		warnings:       h.warnings,
		hub:            NewEvents(),
		ctrlHub:        h.ctrlHub,
		prefix:         h.prefix,
		log:            h.log,
		configured:     h.configured,
		allowEmpty:     h.allowEmpty,
		heartbeat:      h.heartbeat,
		fresh:          h.fresh,
		ops:            h.ops,
		exposureReason: h.exposureReason,
	}
}

// deviceNodeSummary is one entry of GET /nodes.
type deviceNodeSummary struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Kind    string `json:"kind,omitempty"`
	Enabled bool   `json:"enabled"`
}

type deviceNodeList struct {
	Nodes []deviceNodeSummary `json:"nodes"`
}

// listNodes serves GET /nodes: the nodes in the options' order, the
// default first; empty when the host serves none.
func (h *handler) listNodes(w http.ResponseWriter, _ *http.Request) {
	out := deviceNodeList{Nodes: make([]deviceNodeSummary, 0, len(h.served))}
	for _, e := range h.served {
		label := e.label
		if label == "" {
			label = e.key
		}
		out.Nodes = append(out.Nodes, deviceNodeSummary{Key: e.key, Label: label, Kind: e.kind, Enabled: e.port != nil})
	}
	writeJSON(w, out)
}

// adminOnly refuses a request the host's Admin does not admit, or, in
// token mode, one without the token. Mirrors loom's admin-gated route
// group (router.go:1336-1353, `pr.With(admin)`).
func (h *handler) adminOnly(next http.HandlerFunc) http.HandlerFunc {
	return h.adminOnlyBecause(next, "this route writes or reads a pairing secret")
}

// adminOnlyBecause is adminOnly with the refusal's detail.
func (h *handler) adminOnlyBecause(next http.HandlerFunc, detail string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.isAdmin(r) {
			writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required", detail)
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
// the host's source and the node's transitions (node.Observe, the health
// probe and the diagnostic trace: observeNode) — and
// returns the function that unregisters it from all three. A route that
// removes a fabric or opens a window does not publish the transition
// itself: the node reports it through the observer, once.
func (h *handler) subscribe(fn func(Event)) (unsubscribe func()) {
	unsubs := []func(){h.hub.Subscribe(fn)}
	if h.extra != nil {
		unsubs = append(unsubs, h.extra.Subscribe(fn))
	}
	if h.node != nil {
		unsubs = append(unsubs, observeNode(h.node, fn))
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
