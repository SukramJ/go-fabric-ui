// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"sync"
	"time"

	"github.com/SukramJ/go-fabric/controller/client"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// updateDelay coalesces a node's reports into one node_updated: a
// subscribed node reports every change, and a page reloads the node on
// each event.
const updateDelay = 500 * time.Millisecond

// Hooks turn the controller's node hooks (client.Hooks) into the
// controller stream's events: node_added on Commissioned, node_removed
// on Decommissioned, connection, subscription, and node_updated for a
// structure change or a report (coalesced). The host builds them before
// controller.New, wires [Hooks.For] or [Hooks.Wrap] into
// controller.Config.HooksFor, passes them to [ControllerOf], and calls
// [Hooks.Close] before it closes the controller. Events before
// ControllerOf, after Close, or while the controller has no registry are
// dropped. It is safe for concurrent use.
type Hooks struct {
	mu     sync.Mutex
	src    *source
	closed bool
	// ids are the node ids by local id, kept so a Decommissioned hook,
	// which runs once the peer address is gone, still names its node.
	ids map[string]uint64
	// pending are the coalescing timers of node_updated, by local id. A
	// timer's function emits only while it is still the one in pending.
	pending map[string]*time.Timer
	// delay is updateDelay; a test shortens it.
	delay time.Duration
}

// NewHooks returns hooks bound to no controller yet.
func NewHooks() *Hooks {
	return &Hooks{ids: map[string]uint64{}, pending: map[string]*time.Timer{}, delay: updateDelay}
}

// bind sets the source the hooks publish to; from then on hooks fire
// into it. The node ids of the commissioned nodes are learnt at once.
func (h *Hooks) bind(s *source) {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.src = s
	h.mu.Unlock()
	if reg := s.ctrl.Registry(); reg != nil {
		for _, n := range reg.Commissioned() {
			h.nodeID(n.ID())
		}
	}
}

// Close stops the pending node_updated timers and drops every event from
// now on. The host calls it before closing the controller; it is
// idempotent.
func (h *Hooks) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
	h.src = nil
	for local, t := range h.pending {
		t.Stop()
		delete(h.pending, local)
	}
}

// source is the bound source, nil while unbound or closed.
func (h *Hooks) source() *source {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil
	}
	return h.src
}

// nodeID resolves a local id to the node id, remembering it; false while
// unbound, closed, without a registry, or before the node is
// commissioned and known.
func (h *Hooks) nodeID(local string) (uint64, bool) {
	s := h.source()
	if s == nil {
		return 0, false
	}
	h.mu.Lock()
	id, known := h.ids[local]
	h.mu.Unlock()
	reg := s.ctrl.Registry()
	if reg == nil {
		return 0, false
	}
	if n := reg.Get(local); n != nil {
		if addr, ok := n.PeerAddress(); ok {
			h.mu.Lock()
			h.ids[local] = addr.NodeID
			h.mu.Unlock()
			return addr.NodeID, true
		}
	}
	return id, known
}

func (h *Hooks) emit(local string, e fabricui.ControllerEvent) {
	id, ok := h.nodeID(local)
	if !ok {
		return
	}
	s := h.source()
	if s == nil {
		return
	}
	e.Node = id
	s.publish(e)
}

// updated schedules one node_updated for the node.
func (h *Hooks) updated(local string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed || h.src == nil || h.pending[local] != nil {
		return
	}
	var t *time.Timer
	t = time.AfterFunc(h.delay, func() {
		h.mu.Lock()
		current := h.pending[local] == t
		if current {
			delete(h.pending, local)
		}
		h.mu.Unlock()
		// A timer Close stopped too late, or one replaced, emits nothing.
		if current {
			h.emit(local, fabricui.ControllerEvent{Type: fabricui.EventNodeUpdated})
		}
	})
	h.pending[local] = t
}

// For is the hooks of the node with local id: controller.Config.HooksFor
// when the host has no hooks of its own.
func (h *Hooks) For(id string) client.Hooks {
	conn := func(s client.ConnectionState) {
		h.emit(id, fabricui.ControllerEvent{Type: fabricui.EventConnection, Connection: connectionOf(s)})
	}
	return client.Hooks{
		EndpointInstalled:           func(*client.Endpoint) { h.updated(id) },
		ClusterInstalled:            func(*client.Endpoint, *client.Cluster) { h.updated(id) },
		ClusterReplaced:             func(*client.Endpoint, *client.Cluster) { h.updated(id) },
		ClusterDeleted:              func(*client.Endpoint, *client.Cluster) { h.updated(id) },
		Changed:                     func() { h.updated(id) },
		Seeded:                      func() { h.updated(id) },
		ConfigurationVersionChanged: func(*client.Endpoint) { h.updated(id) },
		ConnectionStateChanged:      conn,
		SubscriptionStatusChanged: func(active bool) {
			h.emit(id, fabricui.ControllerEvent{Type: fabricui.EventSubscription, Active: active})
		},
		Commissioned:   func() { h.emit(id, fabricui.ControllerEvent{Type: fabricui.EventNodeAdded}) },
		Decommissioned: func() { h.emit(id, fabricui.ControllerEvent{Type: fabricui.EventNodeRemoved}) },
	}
}

// Wrap chains these hooks after the host's own: each hook of next(id)
// runs, then the stream's. next may be nil.
func (h *Hooks) Wrap(next func(id string) client.Hooks) func(id string) client.Hooks {
	if next == nil {
		return h.For
	}
	return func(id string) client.Hooks { return chain(next(id), h.For(id)) }
}

func chain(a, b client.Hooks) client.Hooks {
	return client.Hooks{
		EndpointInstalled:           both1(a.EndpointInstalled, b.EndpointInstalled),
		ClusterInstalled:            both2(a.ClusterInstalled, b.ClusterInstalled),
		ClusterReplaced:             both2(a.ClusterReplaced, b.ClusterReplaced),
		ClusterDeleted:              both2(a.ClusterDeleted, b.ClusterDeleted),
		Changed:                     both0(a.Changed, b.Changed),
		ConnectionStateChanged:      both1(a.ConnectionStateChanged, b.ConnectionStateChanged),
		SubscriptionStatusChanged:   both1(a.SubscriptionStatusChanged, b.SubscriptionStatusChanged),
		SubscriptionAlive:           both0(a.SubscriptionAlive, b.SubscriptionAlive),
		Event:                       a.Event,
		Seeded:                      both0(a.Seeded, b.Seeded),
		Decommissioned:              both0(a.Decommissioned, b.Decommissioned),
		Commissioned:                both0(a.Commissioned, b.Commissioned),
		ConfigurationVersionChanged: both1(a.ConfigurationVersionChanged, b.ConfigurationVersionChanged),
	}
}

func both0(a, b func()) func() {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	return func() { a(); b() }
}

func both1[T any](a, b func(T)) func(T) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	return func(v T) { a(v); b(v) }
}

func both2[T, U any](a, b func(T, U)) func(T, U) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	return func(v T, w U) { a(v, w); b(v, w) }
}

func connectionOf(s client.ConnectionState) fabricui.ConnectionState {
	switch s {
	case client.Connected:
		return fabricui.ConnectionConnected
	case client.Reconnecting:
		return fabricui.ConnectionReconnecting
	case client.WaitingForDeviceDiscovery:
		return fabricui.ConnectionWaitingForDiscovery
	case client.Disconnected:
	}
	return fabricui.ConnectionDisconnected
}
