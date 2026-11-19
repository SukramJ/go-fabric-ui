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

// go-fabric's QueryUpdates returns the updates of the products whose
// check succeeded with a joined error that names no node: Check keeps
// them and asks every other node on its own, so the failed node is named
// and a passing failure costs nothing (decision 70, UR-P4-adpt-32).
func TestUpdates_CheckNamesTheFailedNode(t *testing.T) {
	var asked []uint64
	u := &updates{
		nodes: func() []productNode { return []productNode{{1, "A"}, {2, "B"}, {3, "C"}} },
		query: func(_ context.Context, only uint64, all bool) ([]softwareupdate.PeerUpdate, error) {
			if all {
				return []softwareupdate.PeerUpdate{peerUpdate(1, 2)}, errors.New("product 0xFFF1/0x8002: download failed")
			}
			asked = append(asked, only)
			if only == 2 {
				return nil, errors.New("download failed")
			}
			return []softwareupdate.PeerUpdate{peerUpdate(only, 5)}, nil
		},
	}
	found, err := u.Check(context.Background())
	failures, ok := errors.AsType[fabricui.UpdateCheckFailures](err)
	if !ok || len(failures) != 1 || failures[0].NodeID != 2 {
		t.Fatalf("err %v", err)
	}
	if len(found) != 2 || found[0].NodeID != 1 || found[1].NodeID != 3 || found[1].SoftwareVersion != 5 {
		t.Errorf("found %+v", found)
	}
	if len(asked) != 2 || asked[0] != 2 || asked[1] != 3 {
		t.Errorf("asked %v, want the nodes without an update found", asked)
	}
}

// When every node's own check succeeds the whole check's failure passed:
// no error.
func TestUpdates_CheckAPassingFailure(t *testing.T) {
	u := &updates{
		nodes: func() []productNode { return []productNode{{1, "A"}} },
		query: func(_ context.Context, _ uint64, all bool) ([]softwareupdate.PeerUpdate, error) {
			if all {
				return nil, errors.New("once")
			}
			return nil, nil
		},
	}
	if found, err := u.Check(context.Background()); err != nil || len(found) != 0 {
		t.Errorf("found %v, err %v", found, err)
	}
}

// A cancelled request ends the check without asking node by node.
func TestUpdates_CheckCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	u := &updates{
		nodes: func() []productNode { t.Error("nodes asked after the cancel"); return nil },
		query: func(context.Context, uint64, bool) ([]softwareupdate.PeerUpdate, error) {
			return nil, context.Canceled
		},
	}
	if _, err := u.Check(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("err %v", err)
	}
}

// One failing product among many nodes costs one further check per
// product, not per node: the failure is given to every node of the
// product, and a product without an update asks no more of its nodes
// (UR-P4-adpt-192).
func TestUpdates_CheckAsksOneNodePerProduct(t *testing.T) {
	nodes := []productNode{{1, "A"}}
	for i := uint64(2); i <= 30; i++ {
		p := "B"
		if i%3 == 0 {
			p = "A"
		}
		nodes = append(nodes, productNode{i, p})
	}
	var asked []uint64
	u := &updates{
		nodes: func() []productNode { return nodes },
		query: func(_ context.Context, only uint64, all bool) ([]softwareupdate.PeerUpdate, error) {
			if all {
				return nil, errors.New("OTA image: the ledger record states no checksum")
			}
			asked = append(asked, only)
			if only == 1 {
				return nil, errors.New("OTA image: the ledger record states no checksum")
			}
			return nil, nil
		},
	}
	_, err := u.Check(context.Background())
	if len(asked) != 2 || asked[0] != 1 || asked[1] != 2 {
		t.Errorf("asked %v, want one node per product: [1 2]", asked)
	}
	failures, _ := errors.AsType[fabricui.UpdateCheckFailures](err)
	if len(failures) != 11 {
		t.Errorf("%d failures, want the eleven nodes of product A: %v", len(failures), err)
	}
	for _, f := range failures {
		if f.NodeID != 1 && f.NodeID%3 != 0 {
			t.Errorf("node %d of product B failed", f.NodeID)
		}
	}
}

// When the product's node finds an update on its own, the product's
// other nodes are checked too, so none of their updates is lost.
func TestUpdates_CheckAProductWithAnUpdateAsksItsOtherNodes(t *testing.T) {
	u := &updates{
		nodes: func() []productNode { return []productNode{{1, "A"}, {2, "A"}} },
		query: func(_ context.Context, only uint64, all bool) ([]softwareupdate.PeerUpdate, error) {
			if all {
				return nil, errors.New("once")
			}
			return []softwareupdate.PeerUpdate{peerUpdate(only, 4)}, nil
		},
	}
	found, err := u.Check(context.Background())
	if err != nil || len(found) != 2 {
		t.Errorf("found %+v, err %v", found, err)
	}
}
