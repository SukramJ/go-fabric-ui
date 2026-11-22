// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"thingfabric.dev/go-fabric/controller"
	"thingfabric.dev/go-fabric/controller/client"
	"thingfabric.dev/go-fabric/fabrictest"

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

// An event the node delivered schedules node_updated, after the host's own
// Event hook ran: the events tab reloads on it.
func TestHooks_EventChainsAndSchedulesAnUpdate(t *testing.T) {
	h := NewHooks()
	h.delay = time.Hour
	_ = ControllerOf(&controller.Controller{}, ControllerOptions{Hooks: h})
	defer h.Close()
	var hostSaw atomic.Int32
	hooks := h.Wrap(func(string) client.Hooks {
		return client.Hooks{Event: func(uint16, uint32, uint32, client.EventData, any) { hostSaw.Add(1) }}
	})("n1")
	hooks.Event(1, 0x0039, 0x03, client.EventData{}, nil)
	h.mu.Lock()
	pending := h.pending["n1"] != nil
	h.mu.Unlock()
	if hostSaw.Load() != 1 || !pending {
		t.Fatalf("host hook ran %d times, node_updated pending %v", hostSaw.Load(), pending)
	}
	if bothEvent(nil, nil) != nil {
		t.Fatal("two absent hooks chain to one")
	}
}

// Decommissioned forgets the node: its id leaves ids and its pending
// node_updated timer is stopped, so no node_updated follows node_removed
// (UR-P1-bound-68).
func TestHooks_DecommissionedForgetsTheNode(t *testing.T) {
	h, _ := hooksOnClosedController(t)
	defer h.Close()
	h.delay = time.Hour
	h.mu.Lock()
	h.ids["n1"] = 0x11
	h.mu.Unlock()
	hooks := h.For("n1")
	hooks.Changed()
	hooks.Decommissioned()
	h.mu.Lock()
	ids, pending := len(h.ids), len(h.pending)
	h.mu.Unlock()
	if ids != 0 || pending != 0 {
		t.Fatalf("after Decommissioned: %d ids, %d timers pending", ids, pending)
	}
}

// A source without hooks releases its subscriptions through its own
// Close (UR-P1-api-67); Close is idempotent.
func TestSource_CloseWithoutHooks(t *testing.T) {
	src := ControllerOf(&controller.Controller{}, ControllerOptions{})
	s := src.(*source)
	released := 0
	s.topoMu.Lock()
	s.unwatch = func() { released++ }
	s.unobserve = func() { released++ }
	s.topoMu.Unlock()
	src.Close()
	src.Close()
	if released != 2 {
		t.Fatalf("released %d subscriptions, want 2", released)
	}
}

// Close cancels a running label push and returns only once it ended; a
// label set after Close starts none (UR-P1-conc-62).
func TestSource_CloseEndsTheLabelPush(t *testing.T) {
	s := ControllerOf(&controller.Controller{}, ControllerOptions{}).(*source)
	started := make(chan struct{})
	var ended, pushes atomic.Int32
	s.push = func(ctx context.Context, _ string) error {
		if pushes.Add(1) == 1 {
			close(started)
		}
		<-ctx.Done()
		ended.Add(1)
		return ctx.Err()
	}
	s.pushLabel()
	<-started
	s.pushLabel() // a second label during the push: one more round, not a second worker
	s.Close()
	if ended.Load() != pushes.Load() || pushes.Load() == 0 {
		t.Fatalf("Close returned with %d of %d pushes ended", ended.Load(), pushes.Load())
	}
	s.pushLabel()
	time.Sleep(10 * time.Millisecond)
	if n := pushes.Load(); n > 2 {
		t.Fatalf("%d pushes, one after Close", n)
	}
}

// Close waits for a node_updated callback that is already running: when
// it returns, nothing of the hooks touches the controller any more
// (UR-P1-conc-62).
func TestHooks_CloseWaitsForARunningTimer(t *testing.T) {
	h, _ := hooksOnClosedController(t)
	h.delay = time.Millisecond
	entered, release := make(chan struct{}), make(chan struct{})
	var done atomic.Bool
	h.fired = func() {
		close(entered)
		<-release
		done.Store(true)
	}
	h.For("n1").Changed()
	<-entered
	closed := make(chan struct{})
	go func() { h.Close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("Close returned while a timer callback runs")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	<-closed
	if !done.Load() {
		t.Fatal("Close returned before the callback ended")
	}
}

// TestHooks_WrapCallsEveryHostHook holds Wrap against go-fabric's Hooks
// by reflection: every func field of client.Hooks the host sets runs
// through the chain, so a hook a go-fabric release adds (Deleted, v0.23.0)
// fails here until chain carries it (UR-P4-adpt-31).
func TestHooks_WrapCallsEveryHostHook(t *testing.T) {
	var host client.Hooks
	hv := reflect.ValueOf(&host).Elem()
	called := map[string]bool{}
	for i := range hv.NumField() {
		f := hv.Field(i)
		if f.Kind() != reflect.Func {
			continue
		}
		name := hv.Type().Field(i).Name
		f.Set(reflect.MakeFunc(f.Type(), func([]reflect.Value) []reflect.Value {
			called[name] = true
			return nil
		}))
	}
	h, _ := hooksOnClosedController(t)
	got := reflect.ValueOf(h.Wrap(func(string) client.Hooks { return host })("n"))
	for i := range got.NumField() {
		f := got.Field(i)
		if f.Kind() != reflect.Func {
			continue
		}
		name := got.Type().Field(i).Name
		if f.IsNil() {
			t.Errorf("Wrap drops the host's %s hook", name)
			continue
		}
		args := make([]reflect.Value, f.Type().NumIn())
		for j := range args {
			args[j] = reflect.Zero(f.Type().In(j))
		}
		f.Call(args)
		if !called[name] {
			t.Errorf("Wrap's %s does not call the host's", name)
		}
	}
}

// A label set on a controller without a fabric is the controller's
// failure, 500 as PUT /controller/fabric/label declares, not a node's
// 502 (UR-P4-test-33; the status guard of main_test.go holds it).
func TestSource_FabricLabelWithoutAFabricIs500(t *testing.T) {
	src := ControllerOf(&controller.Controller{}, ControllerOptions{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, fabricui.DefaultPrefix+"/controller/fabric/label", strings.NewReader(`{"label":"x"}`))
	req.Header.Set(fabricui.RequestHeader, "1")
	req.Header.Set("Content-Type", "application/json")
	fabricui.Handler(fabricui.Options{AllowedHosts: []string{"example.com"}, Controller: src, Admin: func(*http.Request) bool { return true }}).ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("label without a fabric: %d %s", rec.Code, rec.Body)
	}
}

// A local id whose node was removed and that a node commissioned later
// holds again raises node_removed again when that node goes: Commissioned
// clears the id's mark in gone (UR-P4-adpt-201, UR-P4-adpt-393). The
// registry is a real controller's, without a device; the node id is
// placed in ids as the hooks learn it from a commissioned node.
func TestHooks_AReaddedLocalIDIsRemovedAgain(t *testing.T) {
	site := fabrictest.NewSite(t, fabrictest.Options{})
	cfg := site.ControllerConfig(fabrictest.ControllerOptions{Label: "hooks"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var ctrl *controller.Controller
	if err := site.Run(ctx, func(ctx context.Context) (err error) {
		ctrl, err = controller.New(ctx, cfg)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	h := NewHooks()
	src := ControllerOf(ctrl, ControllerOptions{Hooks: h})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		h.Close()
		_ = site.Run(ctx, ctrl.Close)
	})
	var mu sync.Mutex
	var events []fabricui.ControllerEvent
	src.Observe(func(e fabricui.ControllerEvent) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	learn := func() {
		h.mu.Lock()
		h.ids["n1"] = 0x11
		h.mu.Unlock()
	}
	removed := func() int {
		mu.Lock()
		defer mu.Unlock()
		n := 0
		for _, e := range events {
			if e.Type == fabricui.EventNodeRemoved && e.Node == 0x11 {
				n++
			}
		}
		return n
	}
	hooks := h.For("n1")
	learn()
	hooks.Decommissioned()
	hooks.Deleted()
	if got := removed(); got != 1 {
		t.Fatalf("after Decommissioned and Deleted: %d node_removed, want 1", got)
	}
	learn()
	hooks.Commissioned()
	hooks.Deleted()
	if got := removed(); got != 2 {
		t.Fatalf("after the local id was commissioned again and deleted: %d node_removed, want 2", got)
	}
}
