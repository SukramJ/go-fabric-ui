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

func (u *updates) Consent(ctx context.Context, nodeID uint64, vendorID, productID uint16, version uint32) error {
	n, ok := u.s.lookup(nodeID)
	if !ok {
		return fabricui.ErrNodeNotFound
	}
	addr, _ := n.n.PeerAddress()
	_, err := u.m.AddUpdateConsent(ctx, addr, softwareupdate.Target{VendorID: vendorID, ProductID: productID, TargetSoftwareVersion: version})
	return nodeError(err)
}
