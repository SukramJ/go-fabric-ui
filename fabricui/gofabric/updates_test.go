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
		nodes: func() []uint64 { return []uint64{1, 2, 3} },
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
		nodes: func() []uint64 { return []uint64{1} },
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
		nodes: func() []uint64 { t.Error("nodes asked after the cancel"); return nil },
		query: func(context.Context, uint64, bool) ([]softwareupdate.PeerUpdate, error) {
			return nil, context.Canceled
		},
	}
	if _, err := u.Check(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("err %v", err)
	}
}
