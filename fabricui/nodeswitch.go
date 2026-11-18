// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SukramJ/go-fabric/clock"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"
)

// NodeSwitch is a NodeSource whose target the host swaps: the pages keep
// one source while the host erases its node and brings a new one
// (Options.Erase, NodeEntry.Erase). Every method delegates to the current
// target; Set replaces it atomically, after the new node started. Set is
// called with a live node only — a host whose node does not come back
// keeps the old target in place, and the pages read what a stopped node
// answers: its last health verdict (go-fabric node/health.go:74-84 at
// v0.18.0) and the erased store's errors (node.go:1119-1121;
// factory_reset.go:66-67 closes it); what the host says about the failure
// is the host's. Set ends every node stream (GET /node/stream and its
// per-node form) a handler serves while the switch is its Options.Node or
// a NodeEntry.Source: a stream's observers were
// registered on the node it was opened on (events.go:152-166), so the
// browser's EventSource reconnects and its new stream observes the new
// node, and the page reads the node again on that reconnection. The
// page's erase flow polls the status, which reads through the switch,
// until the node answers. As a registry node's
// NodeEntry.Source the same *NodeSwitch is the same value across a Set,
// so the node keeps its handler (registry.go:168-190); listing a new
// *node.Node as Source instead gets a fresh handler and is the contract's
// other path.
//
// A NodeSource that gains a method fails this package's build until
// NodeSwitch delegates it too, so a host that follows its node writes no
// forwarder of its own. It is safe for concurrent use.
type NodeSwitch struct {
	// cur points at the interface value, not an atomic.Value: a Store of
	// another concrete type than the first panics there, and a host may
	// swap a *node.Node for a wrapper or a test double.
	cur atomic.Pointer[NodeSource]
	// mu orders Set against swapped: swap is closed by the Set that
	// replaces the target a stream observes, and replaced by a new one.
	mu   sync.Mutex
	swap chan struct{}
}

var _ NodeSource = (*NodeSwitch)(nil)

// NewNodeSwitch returns a switch over initial. It panics on nil: the
// handlers decide "no node" once, when they are built (a nil
// Options.Node answers 503), so a switch without a target would fail
// inside the first request instead.
func NewNodeSwitch(initial NodeSource) *NodeSwitch {
	if initial == nil {
		panic("fabricui: NewNodeSwitch: nil source")
	}
	s := &NodeSwitch{swap: make(chan struct{})}
	s.cur.Store(&initial)
	return s
}

// Set makes n the target of every later call; a call already running
// finishes on the old target. It panics on nil, for NewNodeSwitch's
// reason.
func (s *NodeSwitch) Set(n NodeSource) {
	if n == nil {
		panic("fabricui: NodeSwitch.Set: nil source")
	}
	s.mu.Lock()
	s.cur.Store(&n)
	old := s.swap
	s.swap = make(chan struct{})
	s.mu.Unlock()
	close(old)
}

// swapped is closed by the next Set. A stream reads it before it
// registers its observers on Current: a Set in between closes the channel
// it holds, so the stream ends instead of staying on the old target.
func (s *NodeSwitch) swapped() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.swap
}

// Current is the target calls reach now.
func (s *NodeSwitch) Current() NodeSource { return *s.cur.Load() }

// The NodeSource methods follow, in the interface's order (options.go),
// each on the current target.

// Health implements [NodeSource] on the current target.
func (s *NodeSwitch) Health() (node.HealthSample, bool) { return s.Current().Health() }

// LocalAddr implements [NodeSource] on the current target.
func (s *NodeSwitch) LocalAddr() string { return s.Current().LocalAddr() }

// FabricInfos implements [NodeSource] on the current target.
func (s *NodeSwitch) FabricInfos(ctx context.Context) ([]node.FabricInfo, error) {
	return s.Current().FabricInfos(ctx)
}

// RemoveFabric implements [NodeSource] on the current target.
func (s *NodeSwitch) RemoveFabric(ctx context.Context, fabricIndex uint8) error {
	return s.Current().RemoveFabric(ctx, fabricIndex)
}

// OpenOwnWindow implements [NodeSource] on the current target.
func (s *NodeSwitch) OpenOwnWindow(ctx context.Context, timeout time.Duration) error {
	return s.Current().OpenOwnWindow(ctx, timeout)
}

// CloseWindow implements [NodeSource] on the current target.
func (s *NodeSwitch) CloseWindow(ctx context.Context) error { return s.Current().CloseWindow(ctx) }

// Window implements [NodeSource] on the current target.
func (s *NodeSwitch) Window() node.WindowSnapshot { return s.Current().Window() }

// OpenBasicWindow implements [NodeSource] on the current target.
func (s *NodeSwitch) OpenBasicWindow(ctx context.Context, creds node.WindowCredentials) error {
	return s.Current().OpenBasicWindow(ctx, creds)
}

// SetupPayload implements [NodeSource] on the current target.
func (s *NodeSwitch) SetupPayload() (node.SetupPayload, error) { return s.Current().SetupPayload() }

// PairingCodesFor implements [NodeSource] on the current target.
func (s *NodeSwitch) PairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
	return s.Current().PairingCodesFor(creds)
}

// Topology implements [NodeSource] on the current target.
func (s *NodeSwitch) Topology() node.BridgeTopology { return s.Current().Topology() }

// SessionDiagnostics implements [NodeSource] on the current target.
func (s *NodeSwitch) SessionDiagnostics() node.SessionDiagnostics {
	return s.Current().SessionDiagnostics()
}

// Clock implements [NodeSource] on the current target.
func (s *NodeSwitch) Clock() clock.Clock { return s.Current().Clock() }

// Observe implements [NodeSource] on the current target.
func (s *NodeSwitch) Observe(o node.Observers) (unsubscribe func()) { return s.Current().Observe(o) }

// MDNSEnabled implements [NodeSource] on the current target.
func (s *NodeSwitch) MDNSEnabled() bool { return s.Current().MDNSEnabled() }

// MDNSServices implements [NodeSource] on the current target.
func (s *NodeSwitch) MDNSServices() []node.MDNSService { return s.Current().MDNSServices() }

// DiagnoseMDNS implements [NodeSource] on the current target.
func (s *NodeSwitch) DiagnoseMDNS() []node.MDNSFinding { return s.Current().DiagnoseMDNS() }

// DiagnosticEvents implements [NodeSource] on the current target.
func (s *NodeSwitch) DiagnosticEvents() []node.DiagnosticEvent { return s.Current().DiagnosticEvents() }

// Surface implements [NodeSource] on the current target.
func (s *NodeSwitch) Surface() surface.Report { return s.Current().Surface() }

// ReadCluster implements [NodeSource] on the current target.
func (s *NodeSwitch) ReadCluster(ctx context.Context, endpoint uint16, cluster uint32) ([]node.AttributeValue, error) {
	return s.Current().ReadCluster(ctx, endpoint, cluster)
}

// ReadAttribute implements [NodeSource] on the current target.
func (s *NodeSwitch) ReadAttribute(ctx context.Context, endpoint uint16, cluster, attribute uint32) (node.AttributeValue, error) {
	return s.Current().ReadAttribute(ctx, endpoint, cluster, attribute)
}

// Events implements [NodeSource] on the current target.
func (s *NodeSwitch) Events(q node.EventQuery) []node.Event { return s.Current().Events(q) }

// Endpoints implements [NodeSource] on the current target.
func (s *NodeSwitch) Endpoints() []node.EndpointInfo { return s.Current().Endpoints() }

// OnHealthSample implements [NodeSource] on the current target.
func (s *NodeSwitch) OnHealthSample(fn func(node.HealthSample)) (unsubscribe func()) {
	return s.Current().OnHealthSample(fn)
}

// OnDiagnosticEvent implements [NodeSource] on the current target.
func (s *NodeSwitch) OnDiagnosticEvent(fn func(node.DiagnosticEvent)) (unsubscribe func()) {
	return s.Current().OnDiagnosticEvent(fn)
}

// AttestationError implements [NodeSource] on the current target.
func (s *NodeSwitch) AttestationError() error { return s.Current().AttestationError() }

// BootReason implements [NodeSource] on the current target.
func (s *NodeSwitch) BootReason() uint8 { return s.Current().BootReason() }

// SessionParameters implements [NodeSource] on the current target.
func (s *NodeSwitch) SessionParameters() node.SessionParameters {
	return s.Current().SessionParameters()
}

// TriggerReannounce implements [NodeSource] on the current target.
func (s *NodeSwitch) TriggerReannounce(ctx context.Context) { s.Current().TriggerReannounce(ctx) }
