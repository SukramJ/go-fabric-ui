// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"testing"

	"github.com/SukramJ/go-fabric/controller/softwareupdate"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

func peerUpdate(id uint64, v uint32) softwareupdate.PeerUpdate {
	return softwareupdate.PeerUpdate{Peer: softwareupdate.Peer{NodeID: id}, Info: softwareupdate.SoftwareUpdateInfo{VendorID: 0xFFF1, ProductID: 0x8001, SoftwareVersion: v}}
}

// productFailed is go-fabric's error of one product's failed check for
// the nodes.
func productFailed(cause error, nodes ...uint64) error {
	pe := &softwareupdate.ProductCheckError{VendorID: 0xFFF1, ProductID: 0x8002, SoftwareVersion: 1, Err: cause}
	for _, id := range nodes {
		pe.Peers = append(pe.Peers, softwareupdate.Peer{NodeID: id})
	}
	return pe
}

// go-fabric's QueryUpdates returns the updates of the products whose
// check succeeded with one ProductCheckError per failed product: Check
// keeps the updates and names every node of a failed product with its
// product's cause, without checking any node again (decision 70,
// UR-P4-adpt-192, decision 80).
func TestUpdates_CheckNamesTheNodesOfEachFailedProduct(t *testing.T) {
	download := errors.New("download failed")
	checksum := errors.New("the ledger record states no checksum")
	calls := 0
	u := &updates{query: func(context.Context) ([]softwareupdate.PeerUpdate, error) {
		calls++
		return []softwareupdate.PeerUpdate{peerUpdate(1, 2)}, errors.Join(productFailed(download, 2, 4), productFailed(checksum, 3))
	}}
	found, err := u.Check(context.Background())
	failures, ok := errors.AsType[fabricui.UpdateCheckFailures](err)
	if !ok || len(failures) != 3 {
		t.Fatalf("err %v", err)
	}
	want := []struct {
		id    uint64
		cause error
	}{{2, download}, {4, download}, {3, checksum}}
	for i, w := range want {
		if failures[i].NodeID != w.id || !errors.Is(failures[i].Err, w.cause) {
			t.Errorf("failure %d: %+v, want node %d with %v", i, failures[i], w.id, w.cause)
		}
	}
	if len(found) != 1 || found[0].NodeID != 1 {
		t.Errorf("found %+v", found)
	}
	if calls != 1 {
		t.Errorf("%d checks, want the one over every node", calls)
	}
}

// An error that names no product fails the whole check: no node is
// blamed for it.
func TestUpdates_CheckAnErrorWithoutAProduct(t *testing.T) {
	u := &updates{query: func(context.Context) ([]softwareupdate.PeerUpdate, error) {
		return []softwareupdate.PeerUpdate{peerUpdate(1, 2)}, errors.New("manager closed")
	}}
	found, err := u.Check(context.Background())
	if _, partial := errors.AsType[fabricui.UpdateCheckFailures](err); err == nil || partial || found != nil {
		t.Errorf("found %v, err %v; want the whole check failed", found, err)
	}
}

// A cancelled request ends the check with the cancellation.
func TestUpdates_CheckCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u := &updates{query: func(context.Context) ([]softwareupdate.PeerUpdate, error) {
		return nil, errors.Join(productFailed(context.Canceled, 1))
	}}
	if _, err := u.Check(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("err %v", err)
	}
	if _, partial := errors.AsType[fabricui.UpdateCheckFailures](func() error { _, err := u.Check(ctx); return err }()); partial {
		t.Error("a cancelled check names failed nodes")
	}
}
