// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"sync"
	"sync/atomic"
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
	// (commissioning progress).
	ctrl    ControllerSource
	ctrlHub *Events
	admin   func(*http.Request) bool
	// token is the admin token of token mode; empty when Options.Admin is
	// set.
	token      string
	reassemble func(context.Context) error
	// erase is the host's erase of this node (Options.Erase,
	// NodeEntry.Erase); nil answers 501.
	erase    func(context.Context) error
	warnings func(context.Context) []Warning
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
	// registry is the host's node registry (Options.Registry), shared by
	// every node's handler; nil without one. nodesHub carries
	// nodes_changed to every node's stream; nil without a registry.
	registry NodeRegistry
	nodesHub *Events
	// config is this node's configuration view (Status.config); nil
	// omits it. A registry node's is replaced on every lookup, so a
	// rename shows at once (managedHandler).
	config atomic.Pointer[NodeConfigView]
	// regMu guards regCache: the handler of each registry node by key,
	// with the entry it was built on (the default node's handler only).
	regMu    sync.Mutex
	regCache map[string]managedNode

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
	// exposure, reassemble and events are the node's own (NodeEntry,
	// contract 0.6.0).
	exposure   ExposureProvider
	reassemble func(context.Context) error
	events     EventSource
	// erase is the node's erase (NodeEntry.Erase, contract 0.7.0).
	erase func(context.Context) error
	// config is the node's configuration view (Status.config).
	config *NodeConfigView
	// managed is a node of the registry: the page may rename and delete
	// it.
	managed bool
}

// entryOf is e in the handler's terms.
func entryOf(e *NodeEntry, managed bool) nodeEntry {
	ne := nodeEntry{key: e.Key, label: e.Label, kind: e.Kind, exposure: e.Exposure, reassemble: e.Reassemble, events: e.Events, erase: e.Erase, managed: managed}
	if e.Config != nil {
		c := *e.Config
		if c.Source == "" {
			c.Source = NodeConfigSourceConfiguration
			if managed {
				c.Source = NodeConfigSourceRegistry
			}
		}
		if c.Label == "" && managed {
			c.Label = e.Label
		}
		ne.config = &c
	}
	switch p := e.Source.(type) {
	case nil:
	case nodePort:
		// A double of this package's tests that is a nodePort already;
		// no type outside the package can implement the unexported
		// methods.
		ne.port = p
	default:
		ne.port = nodeAdapter{n: p}
	}
	return ne
}

// checkNodeKey refuses a key that does not match the contract's pattern,
// is a page name of the node family or is in seen.
func checkNodeKey(key string, seen map[string]bool) error {
	switch {
	case !nodeKeyPattern.MatchString(key):
		return fmt.Errorf("fabricui: node key %q does not match %s", key, nodeKeyPattern)
	case slices.Contains(reservedNodeKeys, key):
		return fmt.Errorf("fabricui: node key %q is a page name of the node family", key)
	case seen[key]:
		return fmt.Errorf("fabricui: node key %q is listed twice", key)
	}
	return nil
}

// nodeKeyPattern is the contract's NodeKey pattern.
var nodeKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// reservedNodeKeys are the node family's page names: the shell's hash
// route #/node/<key>/<page> cannot tell a key from one of them.
var reservedNodeKeys = []string{"status", "pair", "fabrics", "selection", "endpoints", "endpoint", "surface", "diagnostics", "topology"}

// nodeEntries are o's configured nodes, the default first, validated;
// the registry's are listed per request (registryEntries).
func nodeEntries(o *Options) ([]nodeEntry, error) {
	if o.Registry != nil {
		switch {
		case len(o.Nodes) > 0:
			return nil, errors.New("fabricui: Options.Registry and Options.Nodes are both set; the registry lists the nodes beside Options.Node")
		case o.Node == nil:
			return nil, errors.New("fabricui: Options.Registry needs Options.Node, the default node the host configures")
		}
	}
	in := o.Nodes
	if o.Node != nil {
		in = append([]NodeEntry{{Key: DefaultNodeKey, Source: o.Node}}, in...)
	}
	out := make([]nodeEntry, 0, len(in))
	seen := map[string]bool{}
	for i := range in {
		if err := checkNodeKey(in[i].Key, seen); err != nil {
			return nil, err
		}
		seen[in[i].Key] = true
		out = append(out, entryOf(&in[i], false))
	}
	return out, nil
}

// hubOf is the hub a node's handler publishes into and the host's source
// beside it: the host's *Events itself, else a hub of the handler's own
// and src.
func hubOf(src EventSource) (hub *Events, extra EventSource) {
	if ev, ok := src.(*Events); ok && ev != nil {
		return ev, nil
	}
	return NewEvents(), src
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
	// The default node out of Options.Nodes serves its entry's
	// selection, rebuild and events where the Options-level ones are nil.
	if o.Node == nil && len(nodes) > 0 {
		d := nodes[0]
		if o.Exposure == nil {
			o.Exposure = d.exposure
		}
		if o.Reassemble == nil {
			o.Reassemble = d.reassemble
		}
		if o.Events == nil {
			o.Events = d.events
		}
		if o.Erase == nil {
			o.Erase = d.erase
		}
		if o.NodeConfig == nil {
			o.NodeConfig = d.config
		}
	}
	if c := o.NodeConfig; c != nil && c.Source == "" {
		cc := *c
		cc.Source = NodeConfigSourceConfiguration
		o.NodeConfig = &cc
	}
	h := &handler{
		registry:       o.Registry,
		ops:            o.Operations,
		exposureReason: o.ExposureReason,
		nodes:          nodes,
		exposure:       o.Exposure,
		ctrl:           o.Controller,
		ctrlHub:        NewEvents(),
		admin:          o.Admin,
		reassemble:     o.Reassemble,
		erase:          o.Erase,
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
	h.hub, h.extra = hubOf(o.Events)
	h.config.Store(o.NodeConfig)
	if h.registry != nil {
		h.nodesHub = NewEvents()
		h.regCache = map[string]managedNode{}
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
	{"POST", "/mdns/reannounce", true, (*handler).reannounce},
	{"GET", "/compatibility", false, (*handler).compatibility},
	{"GET", "/events", false, (*handler).diagnosticEvents},
	{"GET", "/exposure/schema", false, (*handler).exposureSchema},
	{"GET", "/exposure", false, (*handler).exposureList},
	{"PUT", "/exposure", true, (*handler).exposureApply},
	{"POST", "/exposure/preview", true, (*handler).exposurePreview},
	{"POST", "/maintenance/reassemble", true, (*handler).maintenanceReassemble},
	{"POST", "/maintenance/factory-reset", true, (*handler).factoryReset},
	{"POST", "/maintenance/erase", true, (*handler).eraseNode},
	{"GET", "/stream", false, (*handler).stream},
}

// registryRoutes are the node registry's routes under /nodes (contract
// 0.6.0), every one an admin route.
var registryRoutes = []nodeRoute{
	{"POST", "", true, (*handler).createNode},
	{"PUT", "/{node}/label", true, (*handler).renameNode},
	{"DELETE", "/{node}", true, (*handler).deleteNode},
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
			n, ok := h.nodeHandler(r.PathValue("node"))
			if !ok {
				writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+r.PathValue("node"))
				return
			}
			rt.serve(n, w, r)
		})
	}
	read("GET "+h.prefix+"/nodes", h.listNodes)
	for _, rt := range registryRoutes {
		mux.HandleFunc(rt.method+" "+h.prefix+"/nodes"+rt.path, h.adminOnlyBecause(func(w http.ResponseWriter, r *http.Request) { rt.serve(h, w, r) },
			"this route creates, renames or deletes a node"))
	}
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

// mountNodes builds the handler of each configured node: h itself for the
// default — h.node, which a test may have set after newHandler — and a
// sibling for every other, which shares h's admin check, token and options
// and has a hub, a pairing state and the selection of its entry. A
// registry node's handler is built on its first request (managedHandler).
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
	for i := range served {
		if i == 0 {
			h.keyed[served[i].key] = h
			continue
		}
		h.keyed[served[i].key] = h.sibling(served[i])
	}
}

// sibling is the handler of another node: h's admin check, token, host
// warnings, logger, window options and registry on the node of e, with
// e's selection, rebuild and events.
func (h *handler) sibling(e nodeEntry) *handler {
	s := &handler{
		node:           e.port,
		exposure:       e.exposure,
		reassemble:     e.reassemble,
		erase:          e.erase,
		registry:       h.registry,
		nodesHub:       h.nodesHub,
		admin:          h.admin,
		token:          h.token,
		warnings:       h.warnings,
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
	s.hub, s.extra = hubOf(e.events)
	s.config.Store(e.config)
	return s
}

// deviceNodeSummary is one entry of GET /nodes and the answer of POST
// /nodes.
type deviceNodeSummary struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Kind    string `json:"kind,omitempty"`
	Enabled bool   `json:"enabled"`
	// Managed is a node of the registry, which the page may rename and
	// delete (0.6.0).
	Managed bool `json:"managed"`
	// Exposure is whether the node serves a selection (0.6.0), as its
	// Status.exposure.
	Exposure bool `json:"exposure"`
}

// summaryOf is e as GET /nodes lists it; the label is the key when the
// host names none.
func summaryOf(e nodeEntry) deviceNodeSummary {
	label := e.label
	if label == "" {
		label = e.key
	}
	return deviceNodeSummary{Key: e.key, Label: label, Kind: e.kind, Enabled: e.port != nil, Managed: e.managed, Exposure: e.exposure != nil}
}

type deviceNodeList struct {
	Nodes []deviceNodeSummary `json:"nodes"`
}

// listNodes serves GET /nodes: the configured nodes in the options' order,
// the default first, then the registry's; empty when the host serves
// none.
func (h *handler) listNodes(w http.ResponseWriter, _ *http.Request) {
	reg := h.registryEntries()
	out := deviceNodeList{Nodes: make([]deviceNodeSummary, 0, len(h.served)+len(reg))}
	for i := range h.served {
		e := h.served[i]
		if i == 0 {
			// The default node serves Options.Exposure (h.exposure).
			e.exposure = h.exposure
		}
		out.Nodes = append(out.Nodes, summaryOf(e))
	}
	for i := range reg {
		out.Nodes = append(out.Nodes, summaryOf(entryOf(&reg[i], true)))
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
	if h.nodesHub != nil {
		unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
	}
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
