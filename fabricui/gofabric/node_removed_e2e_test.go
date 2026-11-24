// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"thingfabric.dev/go-fabric/node"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// pairOne commissions the site's device through the source and waits for
// its node_added.
func (s *ctrlSite) pairOne(ctx context.Context) {
	s.t.Helper()
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		s.t.Fatal(err)
	}
	var perr error
	_ = s.site.Run(ctx, func(ctx context.Context) error {
		_, perr = s.src.Pair(ctx, sp.Manual, fabricui.PairOptions{})
		return nil
	})
	if perr != nil || len(s.src.Nodes()) != 1 {
		s.t.Fatalf("pair: %v, %d nodes", perr, len(s.src.Nodes()))
	}
	s.waitEvent(fabricui.EventNodeAdded)
}

// exactlyOneRemoved waits for node_removed and then requires that no
// second one follows within settle.
func (s *ctrlSite) exactlyOneRemoved(how string, settle time.Duration) {
	s.t.Helper()
	deadline := time.After(10 * time.Second)
	var seen []string
	removed := 0
	var quiet <-chan time.Time
	for {
		select {
		case e := <-s.events:
			seen = append(seen, e)
			if e == fabricui.EventNodeRemoved {
				removed++
				if removed == 1 {
					quiet = time.After(settle)
				}
			}
		case <-quiet:
			if removed != 1 {
				s.t.Fatalf("%s: %d node_removed, want exactly one; events %v", how, removed, seen)
			}
			return
		case <-deadline:
			s.t.Fatalf("%s: no node_removed in 10 s; events %v", how, seen)
		}
	}
}

// A node that leaves the controller's registry raises exactly one
// node_removed on each path: a decommission (Decommissioned, then
// Deleted), the peer's leave (the same pair, the node's sessions going
// in between) and Registry.Forget, which runs Deleted only
// (UR-P4-adpt-201).
func TestNodeRemoved_ExactlyOncePerPath(t *testing.T) {
	const settle = 1500 * time.Millisecond
	t.Run("decommission", func(t *testing.T) {
		s := newCtrlSite(t)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
		defer cancel()
		s.pairOne(ctx)
		id := s.src.Nodes()[0].ID()
		if err := s.site.Run(ctx, func(ctx context.Context) error { return s.src.Decommission(ctx, id) }); err != nil {
			t.Fatalf("decommission: %v", err)
		}
		s.exactlyOneRemoved("decommission", settle)
	})
	t.Run("leave", func(t *testing.T) {
		s := newCtrlSite(t)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
		defer cancel()
		s.pairOne(ctx)
		// The device flushes its Leave event to an established
		// subscription before the fabric goes, as go-fabric's own leave
		// test sets it up (controller/client/
		// client_node_parity_matterjs_test.go:281-310 at v0.23.0).
		n := s.src.Nodes()[0]
		if err := s.site.Eventually(ctx, "the default subscription", n.SubscriptionActive); err != nil {
			t.Fatal(err)
		}
		if err := s.site.Settle(ctx); err != nil {
			t.Fatal(err)
		}
		if err := s.site.Run(ctx, func(ctx context.Context) error {
			return s.dev.Act(func(n *node.Node) error {
				fabrics, err := n.FabricInfos(ctx)
				if err != nil {
					return err
				}
				if len(fabrics) != 1 {
					return fmt.Errorf("%d fabrics on the device, want 1", len(fabrics))
				}
				return n.RemoveFabric(ctx, fabrics[0].FabricIndex)
			})
		}); err != nil {
			t.Fatalf("the device leaves: %v", err)
		}
		s.exactlyOneRemoved("leave", settle)
	})
	t.Run("forget", func(t *testing.T) {
		s := newCtrlSite(t)
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
		defer cancel()
		s.pairOne(ctx)
		reg := s.ctrl.Registry()
		nodes := reg.Commissioned()
		if len(nodes) != 1 {
			t.Fatalf("%d commissioned", len(nodes))
		}
		var err error
		_ = s.site.Run(ctx, func(ctx context.Context) error {
			err = reg.Forget(ctx, nodes[0])
			return nil
		})
		if err != nil {
			t.Fatalf("forget: %v", err)
		}
		if n := len(s.src.Nodes()); n != 0 {
			t.Fatalf("the source still lists %d nodes after Forget", n)
		}
		s.exactlyOneRemoved("forget", settle)
	})
}
