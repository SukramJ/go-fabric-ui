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

	"github.com/SukramJ/go-fabric/node"
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
	CommissioningWindow() *node.CommissioningWindow
	OpenBasicWindow(ctx context.Context, creds node.WindowCredentials) error
	PairingCodes() (qr, manual string, err error)
	PairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
	Topology() node.BridgeTopology
	Sessions() *node.SessionManager
	Subscriptions() *node.SubscriptionManager
	MDNSEnabled() bool
	MDNSServices() []node.MDNSService
	DiagnoseMDNS() []node.MDNSFinding
	DiagnosticEvents() []node.DiagnosticEvent
}

// Options configures the handlers.
type Options struct {
	// Node is the node the node family reads and drives; nil serves the
	// family as disabled (status reports enabled false, every other node
	// route answers 503 service_unready).
	Node NodeSource
	// Exposure is the host's selection; nil answers 503 on the exposure
	// routes and hides the selection page.
	Exposure ExposureProvider
	// Admin decides whether a request may write (every POST, PUT and
	// DELETE route) or read a secret (GET /node/pairing). A host behind a
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
	// Events is a source of events beside the handler's own. An [*Events]
	// whose Observers the host wired into node.Config carries the node's
	// fabric and window transitions; the handler then does not publish
	// them a second time.
	Events EventSource
	// Prefix is the path the API is mounted under; empty is
	// [DefaultPrefix]. It starts with "/" and does not end with one.
	Prefix string
	// Logger receives the handlers' logs; nil is slog.Default(). Nothing
	// a handler logs carries a passcode.
	Logger *slog.Logger
	// ConfiguredWindow makes POST /node/pairing/window open the node's own
	// window with its configured passcode and discriminator
	// (CommissioningWindow().OpenOwnWindow). The default opens it with
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
