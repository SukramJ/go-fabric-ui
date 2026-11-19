// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/SukramJ/go-fabric/controller/softwareupdate"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// updates is the controller's softwareupdate.Manager as a
// fabricui.SoftwareUpdateSource: QueuedUpdates, QueryUpdates over the
// ledgers and the stored images, AddUpdateConsent (matter.js
// SoftwareUpdateManager through go-fabric's controller/softwareupdate).
type updates struct {
	s *source
	m *softwareupdate.Manager
	// query is the manager's QueryUpdates over every node (all) or one;
	// nodes the commissioned nodes with their product. Test seams.
	query func(ctx context.Context, only uint64, all bool) ([]softwareupdate.PeerUpdate, error)
	nodes func() []productNode
}

// productNode is a node and its product as the update manager groups
// nodes for a check: vendor, product and numeric software version.
type productNode struct {
	id      uint64
	product string
}

func newUpdates(s *source, m *softwareupdate.Manager) *updates {
	u := &updates{s: s, m: m}
	u.query = func(ctx context.Context, only uint64, all bool) ([]softwareupdate.PeerUpdate, error) {
		if all {
			return m.QueryUpdates(ctx, nil, true)
		}
		n, ok := s.lookup(only)
		if !ok {
			return nil, nil
		}
		return m.QueryUpdates(ctx, n.n, true)
	}
	u.nodes = func() []productNode {
		nodes := s.Nodes()
		out := make([]productNode, 0, len(nodes))
		for _, n := range nodes {
			b := n.Basic()
			// The numeric SoftwareVersion, as QueryUpdates groups
			// (controller/softwareupdate/manager.go:995 at v0.23.0), not
			// the display string (UR-P4-adpt-197).
			var version uint32
			if w, ok := n.(*node); ok {
				version = w.softwareVersion()
			}
			out = append(out, productNode{id: n.ID(), product: fmt.Sprintf("%04X/%04X/%d", b.VendorID, b.ProductID, version)})
		}
		return out
	}
	return u
}

func (u *updates) Queue() []fabricui.PendingUpdate {
	q := u.m.QueuedUpdates()
	out := make([]fabricui.PendingUpdate, 0, len(q))
	for _, p := range q {
		e := fabricui.PendingUpdate{
			NodeID: p.Peer.NodeID, VendorID: p.VendorID, ProductID: p.ProductID,
			TargetSoftwareVersion: p.TargetSoftwareVersion,
			// "in-progress" on the manager, "in_progress" on the wire.
			Status:     strings.ReplaceAll(string(p.Status), "-", "_"),
			ProgressAt: p.LastProgressUpdateTime, Transferred: p.BDXTransferred, Length: p.BDXDataLength,
		}
		if p.LastProgressStatus != nil {
			e.ProgressStatus = p.LastProgressStatus.String()
		}
		out = append(out, e)
	}
	return out
}

// Check is QueryUpdates over every node. go-fabric passes over a
// product whose check failed and returns the updates of the others with
// the joined error, which names no node (softwareupdate.Manager.
// QueryUpdates; go-fabric#555 asks for an error per product). So on an
// error one node per product among the nodes without an update found is
// checked again on its own, as the manager groups a check by vendor,
// product and software version: when that check fails, every node of the
// product is one of the fabricui.UpdateCheckFailures; when it finds an
// update, the product's other nodes are checked too. The updates found
// are returned with the failures (contract 0.11.0, decision 70). Each
// check of one node ends with its own check_finished event.
func (u *updates) Check(ctx context.Context) ([]fabricui.AvailableUpdate, error) {
	found, err := u.query(ctx, 0, true)
	if err == nil {
		return availableOf(found), nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	have := map[uint64]bool{}
	for _, f := range found {
		have[f.Peer.NodeID] = true
	}
	// The nodes without an update found, by product, in the registry's
	// order.
	var order []string
	groups := map[string][]uint64{}
	for _, n := range u.nodes() {
		if have[n.id] {
			continue
		}
		if _, ok := groups[n.product]; !ok {
			order = append(order, n.product)
		}
		groups[n.product] = append(groups[n.product], n.id)
	}
	var failed fabricui.UpdateCheckFailures
	for _, product := range order {
		ids := groups[product]
		one, err := u.query(ctx, ids[0], false)
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err != nil {
			for _, id := range ids {
				failed = append(failed, fabricui.UpdateCheckFailure{NodeID: id, Err: err})
			}
			continue
		}
		found = append(found, one...)
		if len(one) == 0 {
			// No update for the product: its other nodes have none either.
			continue
		}
		for _, id := range ids[1:] {
			more, err := u.query(ctx, id, false)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if err != nil {
				failed = append(failed, fabricui.UpdateCheckFailure{NodeID: id, Err: err})
				continue
			}
			found = append(found, more...)
		}
	}
	if len(failed) == 0 {
		// Every failure was passing: each node's own check succeeded.
		return availableOf(found), nil
	}
	return availableOf(found), failed
}

func availableOf(found []softwareupdate.PeerUpdate) []fabricui.AvailableUpdate {
	out := make([]fabricui.AvailableUpdate, 0, len(found))
	for _, f := range found {
		out = append(out, fabricui.AvailableUpdate{
			NodeID: f.Peer.NodeID, VendorID: f.Info.VendorID, ProductID: f.Info.ProductID,
			SoftwareVersion: f.Info.SoftwareVersion, SoftwareVersionString: f.Info.SoftwareVersionString,
		})
	}
	return out
}

// watchUpdates subscribes the source to the update manager's events
// (softwareupdate.Manager.Observe, controller/softwareupdate/
// observe.go:130-139 at v0.18.0: in order, one at a time, on the
// manager's goroutine and off its locks): each is an update event on the
// controller stream with its kind, the node, the target version for the
// kinds that carry one and the error's text. Nothing of a consent, an
// image or a credential beyond the version reaches the event.
func (s *source) watchUpdates(m *softwareupdate.Manager) {
	unobserve := m.Observe(func(e softwareupdate.UpdateEvent) { s.publish(updateEvent(&e)) })
	s.topoMu.Lock()
	s.unobserve = unobserve
	s.topoMu.Unlock()
}

// updateEvent is a manager's event as the port's update event. Target
// is the found version of UpdateAvailable, the consent's of
// UpdateConsentChanged and the queued entry's otherwise; a check of
// every node carries none (observe.go:99-103).
func updateEvent(e *softwareupdate.UpdateEvent) fabricui.ControllerEvent {
	out := fabricui.ControllerEvent{Type: fabricui.EventUpdate, Node: e.Peer.NodeID, UpdateKind: e.Kind.String()}
	if e.Kind != softwareupdate.UpdateCheckFinished && e.Target.TargetSoftwareVersion != 0 {
		v := e.Target.TargetSoftwareVersion
		out.UpdateVersion = &v
	}
	if e.Err != nil {
		out.UpdateErr = e.Err.Error()
	}
	return out
}

// Consent records the consent (softwareupdate.Manager.AddUpdateConsent).
// A vendor or product that differs from the node's
// (softwareupdate.ErrMismatch) is the request's error, decided before
// any node is asked: an invalid argument, not the node's failure.
func (u *updates) Consent(ctx context.Context, nodeID uint64, vendorID, productID uint16, version uint32) error {
	n, ok := u.s.lookup(nodeID)
	if !ok {
		return fabricui.ErrNodeNotFound
	}
	addr, _ := n.n.PeerAddress()
	_, err := u.m.AddUpdateConsent(ctx, addr, softwareupdate.Target{VendorID: vendorID, ProductID: productID, TargetSoftwareVersion: version})
	if errors.Is(err, softwareupdate.ErrMismatch) {
		return fmt.Errorf("%w: %w", fabricui.ErrInvalidArgument, err)
	}
	// Any other refusal is the manager's (500, as the route declares),
	// never a node's 502 or 409 (UR-P4-test-33).
	return err
}
