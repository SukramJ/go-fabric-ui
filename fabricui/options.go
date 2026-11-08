// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/SukramJ/go-fabric/clock"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"
)

// DefaultPrefix is the path the API is mounted under when
// [Options].Prefix is empty.
const DefaultPrefix = "/api/v1"

// NodeSource is what the node page family reads and drives: the subset of
// *node.Node's methods the handlers call. *node.Node satisfies it; a test
// or a host wrapping its node may supply another implementation.
type NodeSource interface {
	Health() (node.HealthSample, bool)
	LocalAddr() string
	FabricInfos(ctx context.Context) ([]node.FabricInfo, error)
	RemoveFabric(ctx context.Context, fabricIndex uint8) error
	OpenOwnWindow(ctx context.Context, timeout time.Duration) error
	CloseWindow(ctx context.Context) error
	Window() node.WindowSnapshot
	OpenBasicWindow(ctx context.Context, creds node.WindowCredentials) error
	SetupPayload() (node.SetupPayload, error)
	PairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
	Topology() node.BridgeTopology
	SessionDiagnostics() node.SessionDiagnostics
	Clock() clock.Clock
	Observe(o node.Observers) (unsubscribe func())
	MDNSEnabled() bool
	MDNSServices() []node.MDNSService
	DiagnoseMDNS() []node.MDNSFinding
	DiagnosticEvents() []node.DiagnosticEvent
	// Surface is the node's surface report: the specification level and
	// its origin, the controller profiles and the extensions its served
	// surface is composed under, with the decisions, the withheld elements
	// and the announced changes (go-fabric node/surface.go:372, v0.15.2).
	Surface() surface.Report
	// ReadCluster and ReadAttribute read the values the node serves through
	// its own read path, as a controller's read reaches them; Events lists
	// the events its event log holds (go-fabric node/served.go:53-94 and
	// :164-184 at 894adbcd). The endpoint detail and its events read them.
	ReadCluster(ctx context.Context, endpoint uint16, cluster uint32) ([]node.AttributeValue, error)
	ReadAttribute(ctx context.Context, endpoint uint16, cluster, attribute uint32) (node.AttributeValue, error)
	Events(q node.EventQuery) []node.Event
	// Endpoints lists every endpoint the node serves, the root first, with
	// its parent, its PartsList and DeviceTypeList, its name and scope
	// (go-fabric node/endpoints.go:14-66 at 82591648): the endpoint tree
	// of the endpoint detail and the root and Aggregator of the endpoint
	// list.
	Endpoints() []node.EndpointInfo
	// OnHealthSample and OnDiagnosticEvent register a callback for every
	// verdict of the node's health probe and every entry of its diagnostic
	// trace, and return the function that unregisters it (go-fabric
	// node/health.go:93 and node/diagnostics.go:40 at 19588e9b); the
	// stream carries them as health_sample and diagnostic_event (contract
	// 0.5.0). Both run on the goroutine that produced the value and must
	// not block.
	OnHealthSample(fn func(node.HealthSample)) (unsubscribe func())
	OnDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func())
	// AttestationError is why the node's attestation files did not load
	// under Attestation.BlockCommissioningOnError — the node runs on the
	// test chain and refuses every commissioning — nil when they loaded
	// (go-fabric node/attestation.go:55-58 at v0.18.0); the status page
	// shows it (contract 0.7.0, Status.attestation_error).
	AttestationError() error
	// BootReason is the GeneralDiagnostics BootReason the node was built
	// with (node/node.go:1138-1141); Status.boot_reason.
	BootReason() uint8
	// SessionParameters are the session parameters the node reports to
	// every PASE and CASE peer (node/sessions.go:110-118); GET
	// /node/sessions serves them as parameters.
	SessionParameters() node.SessionParameters
	// TriggerReannounce announces the node's DNS-SD records now, a no-op
	// on an advertiser that cannot (node/reannounce.go:12-26); POST
	// /node/mdns/reannounce.
	TriggerReannounce(ctx context.Context)
}

// DefaultNodeKey is the key GET /nodes lists [Options].Node under.
const DefaultNodeKey = "default"

// NodeEntry is one device node of the node family: a key the routes
// /nodes/{key}/… carry, what the selector shows, and the node.
type NodeEntry struct {
	// Key identifies the node in the routes and the shell's hash routes;
	// stable across restarts. It matches ^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$
	// and is not a page name of the node family (status, pair, fabrics,
	// selection, endpoints, endpoint, surface, diagnostics, topology).
	Key string
	// Label is the host's name for the node; empty shows the key.
	Label string
	// Kind is what the host says the node is ("bridge", "device"); empty
	// says nothing.
	Kind string
	// Source is the node; nil lists it as not enabled, and its routes
	// answer as the family's routes without a node.
	Source NodeSource
	// Exposure, Reassemble and Events are the node's own selection, its
	// rebuild and its host events, as [Options].Exposure, Reassemble and
	// Events are the default node's (contract 0.6.0): the exposure and
	// maintenance routes and the stream under /nodes/{key}/… use them, and
	// GET /nodes reports exposure true with a selection. Nil answers as
	// the Options-level field nil does: 503 on the exposure routes and
	// exposure false. For the default node the Options-level fields win;
	// an entry's own serve only when they are nil.
	Exposure   ExposureProvider
	Reassemble func(ctx context.Context) error
	Events     EventSource
	// Erase erases the node and ends it, as [Options].Erase does the
	// default node's (contract 0.7.0): POST /nodes/{key}/maintenance/erase.
	// Nil answers 501 there and reports Status.erase false. For the
	// default node the Options-level field wins.
	Erase func(ctx context.Context) error
	// Config is the node's configuration as the status page shows it,
	// read-only (contract 0.6.0, Status.config); nil omits the block.
	// A registry node's Source defaults to [NodeConfigSourceRegistry],
	// a configured node's to [NodeConfigSourceConfiguration].
	Config *NodeConfigView
}

// The sources of a node's configuration (NodeConfigView.Source).
const (
	// NodeConfigSourceConfiguration is a node the host configures (its
	// flags, its add-on options): nothing of it is editable in the page.
	NodeConfigSourceConfiguration = "configuration"
	// NodeConfigSourceRegistry is a node of the host's registry: the host
	// allocated its values, and only its label is editable.
	NodeConfigSourceRegistry = "registry"
)

// NodeConfigView is what the host shows of a node's configuration on its
// status page (contract 0.6.0): one shape for every node, read-only. It
// carries no passcode — that stays on the pairing page behind admin
// rights.
type NodeConfigView struct {
	// Source is [NodeConfigSourceConfiguration] or
	// [NodeConfigSourceRegistry]; empty takes the node's own.
	Source string
	// SourceNote is the host's wording of where a configured node's
	// values come from ("add-on options", "the --bridge-* flags"); empty
	// says nothing.
	SourceNote string
	// Label is the node's configured name; empty omits it (a registry
	// node's is its entry's label).
	Label string
	// Listen is the address the node listens on, as configured
	// ("[::]:5540").
	Listen        string
	Discriminator uint16
	VendorID      uint16
	// VendorName and ProductName are the configured names; empty omits
	// them.
	VendorName  string
	ProductID   uint16
	ProductName string
	// Store is where the node keeps its state (a file name, not a
	// secret); empty omits it.
	Store string
}

// Options configures the handlers.
type Options struct {
	// Node is the node the node family reads and drives; nil serves the
	// family as disabled (status reports enabled false, every other node
	// route answers 503 service_unready). It is the default node: the
	// /node/* routes serve it, and GET /nodes lists it first under
	// [DefaultNodeKey].
	Node NodeSource
	// Nodes are the device nodes beside Node, in the selector's order,
	// each served under /nodes/{key}/… (contract 0.5.0). With Node nil
	// the first of them is the default node and the /node/* routes serve
	// it. Options.Exposure, Reassemble and Events belong to the default
	// node; another node serves the selection its entry carries (0.6.0).
	// Keys are unique. Nodes and Registry are exclusive.
	Nodes []NodeEntry
	// Registry is the host's node registry (contract 0.6.0): the nodes
	// beside Node that the operator creates, renames and deletes in the
	// page (POST /nodes, PUT /nodes/{key}/label, DELETE /nodes/{key}).
	// Set, GET /nodes lists Node under [DefaultNodeKey] and then
	// Registry.List on every request, and GET /node/status reports
	// node_registry. It needs Node — the first node is the host's
	// configuration, every further one registry state — and refuses
	// Nodes beside it: one source for the list. Nil answers 503 on the
	// three writing routes and lists Node and Nodes as in 0.5.0.
	Registry NodeRegistry
	// Operations is the daemon's log level, log ring and version, the
	// Operation page's /operation/* routes; nil answers 503 there.
	Operations Operations
	// ExposureReason says why the host serves no selection (Exposure is
	// nil), in the host's language; GET /node/status reports it and the
	// selection page shows it. Empty gives no reason.
	ExposureReason string
	// Exposure is the host's selection; nil answers 503 on the exposure
	// routes and hides the selection page.
	Exposure ExposureProvider
	// NodeConfig is the default node's configuration as its status page
	// shows it, read-only (contract 0.6.0, Status.config); nil omits the
	// block. Its Source defaults to [NodeConfigSourceConfiguration].
	NodeConfig *NodeConfigView
	// Controller is the controller the controller family reads and
	// drives (gofabric.ControllerOf over a *controller.Controller); nil
	// serves no controller family, and every /controller/* path answers
	// 404 like an unknown path.
	Controller ControllerSource
	// Admin decides whether a request may write (every POST, PUT and
	// DELETE route) or read a secret (GET /node/pairing), and whether it
	// may use a controller route the contract marks x-admin. A host behind a
	// proxy that authenticates (an Ingress) answers from what the proxy
	// vouches for. Nil selects the token mode: the handler issues a
	// per-process random token in GET /node/status (csrf_token) and
	// requires it in the X-Fabric-UI-Token header of every such request.
	Admin func(*http.Request) bool
	// Reassemble rebuilds the node's endpoints from the host's current
	// selection: POST /node/maintenance/reassemble, and after an exposure
	// change. Nil answers 503 on the maintenance route and skips the
	// rebuild after an exposure change.
	Reassemble func(ctx context.Context) error
	// Erase erases the default node — its fabrics, its store and its
	// endpoint records — and ends it: POST /node/maintenance/erase
	// (contract 0.7.0). go-fabric's Node.FactoryReset does the erase and
	// stops the node, which is single-use (node/factory_reset.go:20-45 at
	// v0.18.0), so the host's Erase also brings a node back — a new node
	// on the erased store, started, as fabrictest.Device.FactoryReset
	// does (fabrictest/device.go:396-416) — or exits for its supervisor to
	// restart it. The handler keeps reading the NodeSource it was given:
	// a host that builds a new node hands over a source that follows it
	// — a [NodeSwitch] over its first node, Set after the new node
	// started. Nil answers 501 not_implemented and reports Status.erase
	// false, so the page offers no erase. POST /node/maintenance/factory-reset is a different
	// action: it removes every fabric and leaves the node running.
	Erase func(ctx context.Context) error
	// Events is a source of events beside the handler's own and the
	// node's: what the host publishes (an exposure change it made outside
	// the page). The node's fabric and window transitions reach the
	// stream through Node.Observe, registered for each stream while it is
	// open, so a handler mounted after the node started sees them too.
	Events EventSource
	// Prefix is the path the API is mounted under; empty is
	// [DefaultPrefix]. It starts with "/" and does not end with one.
	Prefix string
	// Logger receives the handlers' logs; nil is slog.Default(). Nothing
	// a handler logs carries a passcode.
	Logger *slog.Logger
	// ConfiguredWindow makes POST /node/pairing/window open the node's own
	// window with its configured passcode and discriminator
	// (Node.OpenOwnWindow). The default opens it with
	// fresh random credentials (node.OpenBasicWindow), which no earlier
	// reader of the configured codes can use.
	ConfiguredWindow bool
	// AllowEmptyTopology lets POST /node/pairing/window open a window
	// while the node exposes no endpoint yet. By default that is refused
	// with 503 and Retry-After: a controller that subscribes before the
	// first endpoint appears caches an empty parts list.
	AllowEmptyTopology bool
	// Warnings reports what the host found half-valid in its own
	// configuration — a label that resolves to nothing, a rule that
	// matches no item — so the status and selection pages show it rather
	// than only the host's log. It is called on every GET /node/status;
	// nil omits the status response's warnings.
	Warnings func(ctx context.Context) []Warning
}

// Warning is one finding of the host about its own configuration.
type Warning struct {
	// Code identifies the kind of finding, stable across releases of the
	// host ("label_unresolved"); the page shows it as a chip.
	Code string `json:"code"`
	// Message says what is wrong, already in the host's language.
	Message string `json:"message"`
	// Since is when the host first saw the finding; zero omits it.
	Since time.Time `json:"since,omitzero"`
}

// Mount registers the API routes on mux under o.Prefix. It fails on an
// invalid prefix and when mux already holds one of the patterns.
func Mount(mux *http.ServeMux, o Options) (err error) {
	h, err := newHandler(o)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("fabricui: mount: %v", r)
		}
	}()
	h.register(mux)
	return nil
}

// Handler returns the API routes and the embedded web interface: the API
// under o.Prefix, the bundle at "/" with a fallback to its index for any
// other path. An invalid prefix is a programming error and panics.
func Handler(o Options) http.Handler {
	h, err := newHandler(o)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	h.register(mux)
	mux.Handle("/", spaHandler())
	return mux
}

func validPrefix(p string) error {
	if !strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") || strings.ContainsAny(p, " {}") {
		return errors.New("fabricui: Options.Prefix must start with \"/\", not end with one, and hold no pattern characters")
	}
	return nil
}
