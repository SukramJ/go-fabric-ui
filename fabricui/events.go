// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"sync"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
)

// Event types the stream carries.
const (
	EventFabricAdded      = "fabric_added"
	EventFabricRemoved    = "fabric_removed"
	EventWindowOpened     = "window_opened"
	EventWindowClosed     = "window_closed"
	EventEndpointsChanged = "endpoints_changed"
	EventExposureChanged  = "exposure_changed"
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

// nodeObservers turns the node's fabric and window transitions into
// stream events for push: fabric_added, fabric_removed, window_opened and
// window_closed. open is the window state the subscriber last saw, so a
// repeated snapshot of the same state (an administrator's AdminFabricIndex
// cleared) is not a transition. A commissioning is fabric_added: the
// AddNOC that installs the fabric is its last step a page can name.
func nodeObservers(push func(Event), open bool) node.Observers {
	var mu sync.Mutex
	return node.Observers{
		OnFabricAdded: func(f node.FabricInfo) error {
			push(Event{Type: EventFabricAdded, Data: fabricEvent(f)})
			return nil
		},
		OnFabricRemoved: func(f node.FabricInfo) error {
			push(Event{Type: EventFabricRemoved, Data: FabricEvent{FabricIndex: f.FabricIndex}})
			return nil
		},
		OnWindowTransition: func(w node.WindowSnapshot) error {
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
			return nil
		},
	}
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
