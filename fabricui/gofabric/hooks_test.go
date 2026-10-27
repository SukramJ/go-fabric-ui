// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/controller"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// hooksOnClosedController are hooks bound to a controller without a
// registry — what a closed or never-started controller answers — with a
// short coalescing delay, and a counter of the events they publish.
func hooksOnClosedController(t *testing.T) (*Hooks, *atomic.Int32) {
	t.Helper()
	h := NewHooks()
	h.delay = time.Millisecond
	src := ControllerOf(&controller.Controller{}, ControllerOptions{Hooks: h})
	var events atomic.Int32
	src.Observe(func(fabricui.ControllerEvent) { events.Add(1) })
	return h, &events
}

// Regression (PR #12 CI): a node_updated timer that fires after the
// controller's registry is gone drops its event instead of panicking, and
// Close stops the pending timers so none fires.
func TestHooks_TimerAfterCloseDropsTheEvent(t *testing.T) {
	h, events := hooksOnClosedController(t)
	hooks := h.For("n1")
	hooks.Changed()
	h.Close()
	hooks.Changed() // after Close: not scheduled
	time.Sleep(20 * time.Millisecond)
	if n := events.Load(); n != 0 {
		t.Fatalf("%d events after Close", n)
	}
	h.mu.Lock()
	pending := len(h.pending)
	h.mu.Unlock()
	if pending != 0 {
		t.Fatalf("%d timers pending after Close", pending)
	}
	h.Close() // idempotent
}

// Without Close, a timer firing on a controller without a registry drops
// the event; so does every direct hook.
func TestHooks_NoRegistryDropsEvents(t *testing.T) {
	h, events := hooksOnClosedController(t)
	hooks := h.For("n1")
	hooks.Changed()
	hooks.Commissioned()
	hooks.SubscriptionStatusChanged(true)
	time.Sleep(20 * time.Millisecond)
	if n := events.Load(); n != 0 {
		t.Fatalf("%d events without a registry", n)
	}
}

// Hooks firing concurrently with Close neither race nor panic.
func TestHooks_ConcurrentCloseAndUpdates(t *testing.T) {
	h, _ := hooksOnClosedController(t)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			hooks := h.For(string(rune('a' + i)))
			for range 50 {
				hooks.Changed()
				hooks.ConnectionStateChanged(0)
			}
		})
	}
	wg.Go(h.Close)
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
	h.Close()
}
