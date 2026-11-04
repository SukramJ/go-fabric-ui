// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"sync"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"
)

// Event types the stream carries.
const (
	EventFabricAdded      = "fabric_added"
	EventFabricRemoved    = "fabric_removed"
	EventWindowOpened     = "window_opened"
	EventWindowClosed     = "window_closed"
	EventEndpointsChanged = "endpoints_changed"
	EventExposureChanged  = "exposure_changed"
	EventSurfaceChanged   = "surface_changed"
	// EventSessionOpened, EventSessionClosed and EventSubscriptionsChanged
	// carry a [SessionEvent]: go-fabric's node.Observers OnSessionOpened
	// (CASE sessions only), OnSessionClosed and OnSubscriptionsChanged.
	EventSessionOpened        = "session_opened"
	EventSessionClosed        = "session_closed"
	EventSubscriptionsChanged = "subscriptions_changed"
	// EventAttributeChanged carries an [AttributeChangedEvent]: go-fabric's
	// node.Observers.OnAttributeChanged, ids only.
	EventAttributeChanged = "attribute_changed"
)

// Event is one server-sent event: its type and a JSON-encodable payload.
// A payload never carries a passcode.
type Event struct {
	Type string
	Data any
}

// FabricEvent is the payload of fabric_added and fabric_removed.
type FabricEvent struct {
	FabricIndex uint8  `json:"fabric_index"`
	VendorID    uint16 `json:"vendor_id,omitempty"`
	VendorName  string `json:"vendor_name,omitempty"`
	Label       string `json:"label,omitempty"`
}

// WindowEvent is the payload of window_opened and window_closed.
type WindowEvent struct {
	Discriminator   uint16 `json:"discriminator,omitempty"`
	DurationSeconds uint16 `json:"duration_seconds,omitempty"`
}

// ExposureEvent is the payload of exposure_changed: the keys of the items
// that changed; empty when the host does not say which.
type ExposureEvent struct {
	Keys []string `json:"keys"`
}

// SurfaceChangedEvent is the payload of surface_changed: the key of the
// endpoint an assembly changed the served surface of, as the node's
// Observers.OnSurfaceChanged reports it (one endpoint per event) — the key
// of its surface.Report.Changed entry, "" for the root endpoint.
type SurfaceChangedEvent struct {
	Endpoints []string `json:"endpoints"`
}

// AttributeChangedEvent is the payload of attribute_changed: the
// attributes of one cluster of one endpoint that changed, as the node's
// Observers.OnAttributeChanged reports them (go-fabric node/observers.go
// at 894adbcd) — their ids in hex, ascending, never their values: a page
// refetches what it shows.
type AttributeChangedEvent struct {
	Endpoint   uint16   `json:"endpoint"`
	ClusterHex string   `json:"cluster_hex"`
	Attributes []string `json:"attributes"`
}

// SessionEvent is the payload of session_opened, session_closed and
// subscriptions_changed: the session as GET /node/sessions lists it, with
// its subscription count after the change.
type SessionEvent struct {
	SessionID        uint16 `json:"session_id"`
	FabricIndex      uint8  `json:"fabric_index"`
	PeerNodeID       uint64 `json:"peer_node_id"`
	PeerNodeIDHex    string `json:"peer_node_id_hex"`
	LocalNodeID      uint64 `json:"local_node_id"`
	LocalNodeIDHex   string `json:"local_node_id_hex"`
	IsPASE           bool   `json:"is_pase"`
	Subscriptions    int    `json:"subscriptions"`
	LastActivity     string `json:"last_activity"`
	LastPeerActivity string `json:"last_peer_activity"`
	IdleSeconds      int64  `json:"idle_seconds"`
	PeerIdleSeconds  int64  `json:"peer_idle_seconds"`
}

// EventSource is where the stream reads events from beside the handler's
// own: Subscribe registers fn for every event from now on and returns the
// unsubscribe function. fn must not block.
type EventSource interface {
	Subscribe(fn func(Event)) (unsubscribe func())
}

// Events is an in-process [EventSource]. A host publishes its own changes
// (an exposure change it made outside the page, a reassembly) with
// [Events.Publish]; the node's transitions reach the stream through
// Node.Observe and need no wiring. It is safe for concurrent use.
type Events struct {
	mu   sync.Mutex
	next uint64
	subs map[uint64]func(Event)
}

// NewEvents returns an empty event source.
func NewEvents() *Events { return &Events{subs: map[uint64]func(Event){}} }

// Subscribe implements [EventSource].
func (e *Events) Subscribe(fn func(Event)) (unsubscribe func()) {
	e.mu.Lock()
	id := e.next
	e.next++
	e.subs[id] = fn
	e.mu.Unlock()
	return func() {
		e.mu.Lock()
		delete(e.subs, id)
		e.mu.Unlock()
	}
}

// Publish hands ev to every subscriber, on the caller's goroutine.
func (e *Events) Publish(ev Event) {
	e.mu.Lock()
	fns := make([]func(Event), 0, len(e.subs))
	for _, fn := range e.subs {
		fns = append(fns, fn)
	}
	e.mu.Unlock()
	for _, fn := range fns {
		fn(ev)
	}
}

// NodeObservers turns the node's fabric, window, surface and session
// transitions and its attribute changes into the node page family's events
// for push: fabric_added, fabric_removed, window_opened, window_closed,
// surface_changed, session_opened, session_closed, subscriptions_changed
// and attribute_changed, with the payloads the stream carries. go-fabric reports CASE sessions only
// (node/observers.go OnSessionOpened), so a commissioning's PASE session
// raises none of the three.
// open is the window state the subscriber last saw, so a repeated snapshot
// of the same state (an administrator's AdminFabricIndex cleared) is not a
// transition. A commissioning is fabric_added: the AddNOC that installs the
// fabric is its last step a page can name. A surface change is
// surface_changed, whichever assembly made it — a route's reassembly or
// one the host runs on its own (go-fabric node/observers.go:75-84).
//
// The handler registers it with node.Observe for each open stream, and the
// transitions reach those streams alone, never [Events]. A host that
// carries the node page family's events over a channel of its own (a
// WebSocket hub, an audit log) registers it the same way:
//
//	unobserve := n.Observe(fabricui.NodeObservers(push, n.Window().Open))
//
// Publishing the transitions into the [Events] the handler reads instead
// would hand every stream each one twice.
func NodeObservers(push func(Event), open bool) node.Observers {
	var mu sync.Mutex
	return node.Observers{
		OnFabricAdded: func(f node.FabricInfo) {
			push(Event{Type: EventFabricAdded, Data: fabricEvent(f)})
		},
		OnFabricRemoved: func(f node.FabricInfo) {
			push(Event{Type: EventFabricRemoved, Data: FabricEvent{FabricIndex: f.FabricIndex}})
		},
		OnWindowTransition: func(w node.WindowSnapshot) {
			mu.Lock()
			changed := open != w.Open
			open = w.Open
			mu.Unlock()
			switch {
			case !changed:
			case w.Open:
				push(Event{Type: EventWindowOpened, Data: windowEvent(w)})
			default:
				push(Event{Type: EventWindowClosed, Data: WindowEvent{}})
			}
		},
		OnSurfaceChanged: func(c surface.SurfaceChange) {
			push(Event{Type: EventSurfaceChanged, Data: surfaceChangedEvent(c)})
		},
		OnSessionOpened: func(s node.SessionInfo) {
			push(Event{Type: EventSessionOpened, Data: SessionEvent(sessionOf(&s))})
		},
		OnSessionClosed: func(s node.SessionInfo) {
			push(Event{Type: EventSessionClosed, Data: SessionEvent(sessionOf(&s))})
		},
		OnSubscriptionsChanged: func(s node.SessionInfo) {
			push(Event{Type: EventSubscriptionsChanged, Data: SessionEvent(sessionOf(&s))})
		},
		OnAttributeChanged: func(c node.AttributeChange) {
			push(Event{Type: EventAttributeChanged, Data: attributeChangedEvent(c)})
		},
	}
}

// attributeChangedEvent is attribute_changed's payload for one change.
func attributeChangedEvent(c node.AttributeChange) AttributeChangedEvent {
	ids := make([]string, 0, len(c.Change.Attributes))
	for _, a := range c.Change.Attributes {
		ids = append(ids, hexID(a))
	}
	return AttributeChangedEvent{Endpoint: c.Endpoint, ClusterHex: hexID(c.Cluster), Attributes: ids}
}

// surfaceChangedEvent is surface_changed's payload for one change.
func surfaceChangedEvent(c surface.SurfaceChange) SurfaceChangedEvent {
	return SurfaceChangedEvent{Endpoints: []string{c.Endpoint}}
}

// windowEvent is window_opened's payload for an open window.
func windowEvent(w node.WindowSnapshot) WindowEvent {
	return WindowEvent{Discriminator: w.Discriminator, DurationSeconds: durationSeconds(w.Timeout)}
}

func fabricEvent(f node.FabricInfo) FabricEvent {
	return FabricEvent{
		FabricIndex: f.FabricIndex,
		VendorID:    f.VendorID,
		VendorName:  eligibility.VendorName(f.VendorID),
		Label:       f.Label,
	}
}
