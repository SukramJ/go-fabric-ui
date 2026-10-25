// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"sync"
	"sync/atomic"

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
// [Events.Publish], and wires the node's fabric and window transitions
// into it with [Events.Observers]. It is safe for concurrent use.
type Events struct {
	mu        sync.Mutex
	next      uint64
	subs      map[uint64]func(Event)
	observing atomic.Bool
	// windowOpen is the last window state an observer reported, so a
	// repeated snapshot of the same state is not a transition.
	windowOpen atomic.Bool
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

// Observers returns next with the fabric and window observers chained so
// that every transition also reaches this source: fabric_added,
// fabric_removed, window_opened and window_closed. next's own observers
// run first and their error is returned. A host assigns the result to
// node.Config.Observers before node.New.
//
// Once Observers has been called the handler stops publishing these
// events itself, so a route that removes a fabric or opens a window does
// not report the transition twice.
func (e *Events) Observers(next node.Observers) node.Observers {
	e.observing.Store(true)
	out := next
	out.OnFabricAdded = func(f node.FabricInfo) error {
		err := call(next.OnFabricAdded, f)
		e.Publish(Event{Type: EventFabricAdded, Data: fabricEvent(f)})
		return err
	}
	out.OnFabricRemoved = func(f node.FabricInfo) error {
		err := call(next.OnFabricRemoved, f)
		e.Publish(Event{Type: EventFabricRemoved, Data: FabricEvent{FabricIndex: f.FabricIndex}})
		return err
	}
	out.OnWindowTransition = func(w node.WindowSnapshot) error {
		err := call(next.OnWindowTransition, w)
		if e.windowOpen.Swap(w.Open) != w.Open {
			if w.Open {
				e.Publish(Event{Type: EventWindowOpened, Data: WindowEvent{Discriminator: w.Discriminator}})
			} else {
				e.Publish(Event{Type: EventWindowClosed, Data: WindowEvent{}})
			}
		}
		return err
	}
	return out
}

// observingNode reports whether [Events.Observers] wired the node into
// this source.
func (e *Events) observingNode() bool { return e.observing.Load() }

func call[T any](fn func(T) error, v T) error {
	if fn == nil {
		return nil
	}
	return fn(v)
}

func fabricEvent(f node.FabricInfo) FabricEvent {
	return FabricEvent{
		FabricIndex: f.FabricIndex,
		VendorID:    f.VendorID,
		VendorName:  eligibility.VendorName(f.VendorID),
		Label:       f.Label,
	}
}
