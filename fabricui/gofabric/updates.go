// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
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

func (u *updates) Check(ctx context.Context) ([]fabricui.AvailableUpdate, error) {
	found, err := u.m.QueryUpdates(ctx, nil, true)
	if err != nil {
		return nil, nodeError(err)
	}
	out := make([]fabricui.AvailableUpdate, 0, len(found))
	for _, f := range found {
		out = append(out, fabricui.AvailableUpdate{
			NodeID: f.Peer.NodeID, VendorID: f.Info.VendorID, ProductID: f.Info.ProductID,
			SoftwareVersion: f.Info.SoftwareVersion, SoftwareVersionString: f.Info.SoftwareVersionString,
		})
	}
	return out, nil
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
// every node carries none (observe.go:98-104).
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

func (u *updates) Consent(ctx context.Context, nodeID uint64, vendorID, productID uint16, version uint32) error {
	n, ok := u.s.lookup(nodeID)
	if !ok {
		return fabricui.ErrNodeNotFound
	}
	addr, _ := n.n.PeerAddress()
	_, err := u.m.AddUpdateConsent(ctx, addr, softwareupdate.Target{VendorID: vendorID, ProductID: productID, TargetSoftwareVersion: version})
	return nodeError(err)
}
