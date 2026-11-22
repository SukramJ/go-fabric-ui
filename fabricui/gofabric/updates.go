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
	// query is the manager's QueryUpdates over every node. A test seam.
	query func(ctx context.Context) ([]softwareupdate.PeerUpdate, error)
}

func newUpdates(s *source, m *softwareupdate.Manager) *updates {
	return &updates{s: s, m: m, query: func(ctx context.Context) ([]softwareupdate.PeerUpdate, error) {
		return m.QueryUpdates(ctx, nil, true)
	}}
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
// product whose check failed, returns the updates of the others and joins
// one *softwareupdate.ProductCheckError per failed product, naming the
// product's nodes (go-fabric#555; controller/softwareupdate/
// check_error.go:8-59 and manager.go:1829-1843 at v0.24.1). Each of those
// nodes is one of the fabricui.UpdateCheckFailures, with its product's
// cause, beside the updates found (contract 0.11.0, decision 70); no node
// is checked again (UR-P4-adpt-192, decision 80). An error that names no
// product fails the whole check.
func (u *updates) Check(ctx context.Context) ([]fabricui.AvailableUpdate, error) {
	found, err := u.query(ctx)
	if err == nil {
		return availableOf(found), nil
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	products := softwareupdate.ProductCheckErrors(err)
	if len(products) == 0 {
		return nil, err
	}
	var failed fabricui.UpdateCheckFailures
	for _, p := range products {
		for _, peer := range p.Peers {
			failed = append(failed, fabricui.UpdateCheckFailure{NodeID: peer.NodeID, Err: p.Err})
		}
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
// Since go-fabric v0.24.0 the manager keeps no consent to a version at or
// below the one the node runs and answers false, as it does for a consent
// it keeps for a node not applicable for updates
// (controller/softwareupdate/manager.go:2345-2354 at v0.24.1); the two are
// not told apart, so neither is an error here.
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
