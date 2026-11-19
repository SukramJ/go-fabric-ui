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
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// TokenHeader carries the admin token in token mode (a nil
// [Options].Admin).
const TokenHeader = "X-Fabric-UI-Token" //nolint:gosec // a header name, not a credential

// RequestHeader, with the value "1", marks a request of the web
// interface's client: in host mode (a non-nil [Options].Admin) a request
// has admin rights only with it and Admin's yes. The shipped client sends
// it on every request.
const RequestHeader = "X-Fabric-UI"

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
	token string
	// hooks are the host's rebuild and erase of this node (Options.
	// Reassemble and Erase, NodeEntry's); a registry node's are replaced
	// on every lookup, as config is (UR-P1-conc-06).
	hooks    atomic.Pointer[nodeHooks]
	warnings func(context.Context) []Warning
	counters func(context.Context) []Counter
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
	// regGen counts the registry deletions (under regMu): a lookup that a
	// deletion overtook caches nothing (UR-P1-conc-11).
	regGen uint64
	// done is closed by stop when the registry deletes this node: its
	// open streams end (UR-P1-conc-11). Nil, never closed, on every other
	// handler.
	done chan struct{}
	stop func()

	// mu guards ephemeral: the credentials of the window this handler
	// opened, served by GET /node/pairing while that window is open.
	mu        sync.Mutex
	ephemeral *ephemeralWindow
	// openMu serialises POST /node/pairing/window from its check of the
	// node's window to the stored credentials.
	openMu sync.Mutex

	// streams counts the open event streams of every kind and node,
	// shared by the siblings (Options.MaxStreams); hostDone is
	// Options.Done.
	streams  *streamCount
	hostDone <-chan struct{}

	// allowedHosts and allowHost are Options.AllowedHosts and AllowHost:
	// the Host check (hostCheck).
	allowedHosts map[string]bool
	allowHost    func(*http.Request) bool
}

// streamCount is the open streams and their bound.
type streamCount struct {
	open atomic.Int64
	max  int64
}

// acquire takes one stream if fewer than max are open.
func (c *streamCount) acquire() bool {
	if c.open.Add(1) > c.max {
		c.open.Add(-1)
		return false
	}
	return true
}

func (c *streamCount) release() { c.open.Add(-1) }

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

// hubOf is the hub a node's handler publishes into, always its own, and
// the host's source beside it, which reaches admin requests only
// (UR-P3-api-08): a host's *Events no longer carries the handler's events.
func hubOf(src EventSource) (hub *Events, extra EventSource) {
	if ev, ok := src.(*Events); ok && ev == nil {
		// A typed-nil *Events is no source: as extra it would pass an
		// interface nil check and panic on Subscribe (UR-P1-err-03).
		return NewEvents(), nil
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
	if err := validFrameAncestors(o.FrameAncestors); err != nil {
		return nil, err
	}
	allowed, err := allowedHostSet(o.AllowedHosts)
	if err != nil {
		return nil, err
	}
	maxStreams := o.MaxStreams
	switch {
	case maxStreams < 0:
		return nil, errors.New("fabricui: Options.MaxStreams is negative")
	case maxStreams == 0:
		maxStreams = DefaultMaxStreams
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
		warnings:       o.Warnings,
		counters:       o.Counters,
		prefix:         prefix,
		log:            o.Logger,
		configured:     o.ConfiguredWindow,
		allowEmpty:     o.AllowEmptyTopology,
		heartbeat:      15 * time.Second,
		fresh:          node.NewWindowCredentials,
		streams:        &streamCount{max: int64(maxStreams)},
		hostDone:       o.Done,
		allowedHosts:   allowed,
		allowHost:      o.AllowHost,
	}
	h.hooks.Store(&nodeHooks{reassemble: o.Reassemble, erase: o.Erase})
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
	// admin is the contract's x-admin: the route answers 403 to a request
	// without admin rights.
	admin bool
	serve func(*handler, http.ResponseWriter, *http.Request)
}

// nodeRoutes is the node family's route table: docs/contract.md and
// contract/openapi.yaml; loom's block it replaces is
// internal/north/rest/router.go:1328-1359 at b40b289a.
var nodeRoutes = []nodeRoute{
	{"GET", "/status", false, (*handler).status},
	{"GET", "/fabrics", true, (*handler).fabrics},
	{"DELETE", "/fabrics/{index}", true, (*handler).removeFabric},
	{"GET", "/pairing", true, (*handler).pairing},
	{"POST", "/pairing/window", true, (*handler).openWindow},
	{"POST", "/pairing/window/close", true, (*handler).closeWindow},
	{"GET", "/endpoints", true, (*handler).endpoints},
	{"GET", "/endpoints/{ep}", true, (*handler).nodeEndpoint},
	{"GET", "/endpoints/{ep}/events", true, (*handler).nodeEndpointEvents},
	{"GET", "/surface", true, (*handler).surfaceReport},
	{"GET", "/sessions", true, (*handler).sessions},
	{"GET", "/topology", true, (*handler).topology},
	{"GET", "/mdns", true, (*handler).mdns},
	{"POST", "/mdns/reannounce", true, (*handler).reannounce},
	{"GET", "/compatibility", true, (*handler).compatibility},
	{"GET", "/events", true, (*handler).diagnosticEvents},
	{"GET", "/exposure/schema", true, (*handler).exposureSchema},
	{"GET", "/exposure", true, (*handler).exposureList},
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
var operationAdmin = map[string]bool{
	"PUT /operation/log/level":  true,
	"GET /operation/log":        true,
	"GET /operation/log/stream": true,
}

// register puts the API on mux (see mount) and panics where mount fails.
func (h *handler) register(mux *http.ServeMux) {
	if err := h.mount(mux); err != nil {
		panic(err)
	}
}

// mount puts the API on mux under the one pattern prefix + "/", which
// serves a mux of the API's own (routes). One registration on the host's
// mux either succeeds or leaves it as it was: first every route pattern
// is looked up on mux, and one the host already serves under the prefix
// fails the mount before anything is registered; then the single
// registration's panic (a conflicting host pattern) is recovered into the
// error. A pattern the host registers under the prefix after the mount is
// more specific than prefix + "/" and shadows the API's route.
func (h *handler) mount(mux *http.ServeMux) (err error) {
	// The API's own mux is built inside the recovered section too: a
	// pattern it refuses is an error of Mount, not a panic (UR-P1-api-02).
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fabricui: mount: %v; nothing was registered", r)
		}
	}()
	api, patterns := h.routes()
	for _, p := range patterns {
		if q := servedPattern(mux, p); q != "" && h.underPrefix(q) {
			return fmt.Errorf("fabricui: mount: the mux already serves %q for the API route %q; nothing was registered", q, p)
		}
	}
	mux.Handle(h.prefix+"/", apiHeaders(h.hostCheck(api)))
	return nil
}

// apiHeaders sets the headers every answer of the API carries before the
// route writes: no-store, because answers carry pairing codes, a token or
// the daemon's state, which no cache keeps (UR-P0-leak-25); nosniff and
// no-referrer as on the bundle (UR-P0-web-24).
func apiHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := w.Header()
		hdr.Set("Cache-Control", "no-store")
		setCommonHeaders(hdr)
		if observeStatus == nil {
			next.ServeHTTP(w, r)
			return
		}
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if !sw.pathLevel {
			observeStatus(r.Method, r.URL.Path, sw.status)
		}
	})
}

// observeStatus, set by the package's tests only, sees every status the
// API answers: the guard that each is declared for its operation in the
// contract (decision 47, TestMain).
var observeStatus func(method, path string, status int)

// statusWriter records the status written through it.
type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
	// pathLevel marks an answer no operation gives: noRoute's 404 and
	// 405, and the 404 of an absent family.
	pathLevel bool
}

func (s *statusWriter) WriteHeader(code int) {
	if !s.wrote {
		s.status, s.wrote = code, true
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusWriter) Write(b []byte) (int, error) {
	s.wrote = true
	return s.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the writer's flush and
// deadlines.
func (s *statusWriter) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// setCommonHeaders sets what every response of the handler carries, the
// API's and the bundle's: nosniff and no-referrer (UR-P0-web-24).
func setCommonHeaders(hdr http.Header) {
	hdr.Set("X-Content-Type-Options", "nosniff")
	hdr.Set("Referrer-Policy", "no-referrer")
}

// servedPattern is the pattern mux serves a request for p with — p's
// method, or GET, and its path with every wildcard a plain segment — or
// "" when mux has none.
func servedPattern(mux *http.ServeMux, p string) string {
	method, path, ok := strings.Cut(p, " ")
	if !ok {
		method, path = http.MethodGet, p
	}
	path = wildcard.ReplaceAllString(path, "x")
	req, err := http.NewRequestWithContext(context.Background(), method, path, http.NoBody)
	if err != nil {
		return ""
	}
	_, q := mux.Handler(req)
	return q
}

// wildcard is a pattern's {name} or {name...} segment.
var wildcard = regexp.MustCompile(`\{[^}]*\}`)

// underPrefix reports whether the pattern q (method and host optional)
// matches paths under the API's prefix.
func (h *handler) underPrefix(q string) bool {
	if _, rest, ok := strings.Cut(q, " "); ok {
		q = rest
	}
	if i := strings.IndexByte(q, '/'); i >= 0 {
		q = q[i:]
	}
	return q == h.prefix || strings.HasPrefix(q, h.prefix+"/")
}

// routes builds every route on a mux of the API's own and returns it with
// the patterns it holds: the node family under /node for the default node
// and under /nodes/{node} for each node, GET /nodes, the operation routes
// and the controller family.
func (h *handler) routes() (api http.Handler, patterns []string) {
	h.mountNodes()
	mux := http.NewServeMux()
	handle := func(pattern string, fn http.Handler) {
		mux.Handle(pattern, fn)
		patterns = append(patterns, pattern)
	}
	read := func(pattern string, fn http.HandlerFunc) { handle(pattern, h.openRoute(fn)) }
	write := func(pattern string, fn http.HandlerFunc) { handle(pattern, h.adminOnly(fn)) }

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
		handle(rt.method+" "+h.prefix+"/nodes"+rt.path, h.adminOnlyBecause(func(w http.ResponseWriter, r *http.Request) { rt.serve(h, w, r) },
			"this route creates, renames or deletes a node"))
	}
	read("GET "+h.prefix+"/operation/info", h.operationInfo)
	read("GET "+h.prefix+"/operation/log/level", h.logLevel)
	handle("PUT "+h.prefix+"/operation/log/level", h.adminOnlyBecause(h.setLogLevel, "this route changes the daemon's log level"))
	write("GET "+h.prefix+"/operation/log", h.logRecords)
	write("GET "+h.prefix+"/operation/log/stream", h.logStream)
	if h.ctrl != nil {
		handle(h.prefix+"/controller/", h.controllerRoutes())
	}
	// Anything else under the prefix is an API 405 or 404, not the web
	// interface's index.
	catchAll := h.prefix + "/"
	handle(catchAll, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		noRoute(w, r, mux, catchAll)
	}))
	return mux, patterns
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
// warnings and counters, logger, window options and registry on the node of e, with
// e's selection, rebuild and events.
func (h *handler) sibling(e nodeEntry) *handler {
	s := &handler{
		node:           e.port,
		exposure:       e.exposure,
		registry:       h.registry,
		nodesHub:       h.nodesHub,
		admin:          h.admin,
		token:          h.token,
		warnings:       h.warnings,
		counters:       h.counters,
		ctrlHub:        h.ctrlHub,
		prefix:         h.prefix,
		log:            h.log,
		configured:     h.configured,
		allowEmpty:     h.allowEmpty,
		heartbeat:      h.heartbeat,
		fresh:          h.fresh,
		ops:            h.ops,
		exposureReason: h.exposureReason,
		streams:        h.streams,
		hostDone:       h.hostDone,
	}
	s.hub, s.extra = hubOf(e.events)
	s.config.Store(e.config)
	s.hooks.Store(&nodeHooks{reassemble: e.reassemble, erase: e.erase})
	if e.managed {
		s.done = make(chan struct{})
		s.stop = sync.OnceFunc(func() { close(s.done) })
	}
	return s
}

// nodeHooks are a node's rebuild and erase; either may be nil.
type nodeHooks struct {
	reassemble func(context.Context) error
	// erase is the host's erase of this node; nil answers 501.
	erase func(context.Context) error
}

// reassembleFn is the node's rebuild; nil without one.
func (h *handler) reassembleFn() func(context.Context) error {
	if k := h.hooks.Load(); k != nil {
		return k.reassemble
	}
	return nil
}

// eraseFn is the node's erase; nil without one.
func (h *handler) eraseFn() func(context.Context) error {
	if k := h.hooks.Load(); k != nil {
		return k.erase
	}
	return nil
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
	return h.adminOnlyBecause(next, "this route needs admin rights")
}

// adminOnlyBecause is adminOnly with the refusal's detail: the Origin,
// then the admin test, then the body's media type, as every admin route
// checks them (admitAdmin).
func (h *handler) adminOnlyBecause(next http.HandlerFunc, detail string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.admitAdmin(w, r, detail) || !jsonBody(w, r) {
			return
		}
		next(w, r)
	}
}

// openRoute checks an open route's body media type (jsonBody).
func (h *handler) openRoute(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if jsonBody(w, r) {
			next(w, r)
		}
	}
}

// admitAdmin answers 403 origin_not_allowed for an Origin the handler
// does not admit and 403 forbidden for a request without admin rights,
// and reports whether the request may go on.
func (h *handler) admitAdmin(w http.ResponseWriter, r *http.Request, detail string) bool {
	if !h.originAllowed(r) {
		writeProblem(w, r, http.StatusForbidden, codeOriginNotAllowed, "Origin not allowed", "the handler does not admit this origin")
		return false
	}
	if !h.isAdmin(r) {
		writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required", detail)
		return false
	}
	return true
}

// requestAdmin is the one admin test of a request on an open route: the
// status's admin field, its reduced answer and the node stream's filter.
// An Origin the handler does not admit makes the request a non-admin's.
func (h *handler) requestAdmin(r *http.Request) bool { return h.originAllowed(r) && h.isAdmin(r) }

// originAllowed reports whether r's Origin, when it carries one, names a
// host the name rules admit, or, for a request Options.AllowHost
// admitted, the request's own Host and port; "null" and an unparsable
// Origin are not.
func (h *handler) originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" || u.User != nil {
		return false
	}
	if h.nameAllowed(hostName(u.Host)) {
		return true
	}
	a, _ := admissionOf(r)
	return a.byPredicate && strings.EqualFold(u.Host, r.Host)
}

// isAdmin is the admin test: in host mode RequestHeader set to "1" and
// the host's Admin; in token mode the token.
func (h *handler) isAdmin(r *http.Request) bool {
	if h.admin != nil {
		return r.Header.Get(RequestHeader) == "1" && h.admin(r)
	}
	got := r.Header.Get(TokenHeader)
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(h.token)) == 1
}

// publish hands ev to the stream's subscribers.
func (h *handler) publish(ev Event) { h.hub.Publish(ev) }

// subscribe registers fn for every event from now on as an admin request
// receives them (subscribeAs).
func (h *handler) subscribe(fn func(Event)) (unsubscribe func()) {
	return h.subscribeAs(true, fn)
}

// subscribeAs registers fn for every event from now on — the handler's own,
// the host's source and the node's transitions (node.Observe, the health
// probe and the diagnostic trace: observeNode) — and
// returns the function that unregisters it from all three. A route that
// removes a fabric or opens a window does not publish the transition
// itself: the node reports it through the observer, once. Without admin
// rights fn receives the types nodeStreamEvents allows and nothing of the
// host's source.
func (h *handler) subscribeAs(admin bool, fn func(Event)) (unsubscribe func()) {
	if !admin {
		all := fn
		fn = func(ev Event) {
			if nodeStreamEvents[ev.Type] {
				all(ev)
			}
		}
	}
	unsubs := []func(){h.hub.Subscribe(fn)}
	if h.nodesHub != nil {
		unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
	}
	if h.extra != nil && admin {
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
