// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"testing"
	"time"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// The request's context ends after the device took our certificates and
// before CommissioningComplete (step Reconnect 18.1): the commissioning
// fails, and go-fabric v0.23.0 disarms the device's fail-safe on the
// cancelled context, so the device rolls the fabric back and a second
// attempt right after it commissions the device, instead of meeting
// BusyWithOtherAdmin until the fail-safe expires (UR-P2-err-09).
func TestPair_ACancelledCommissioningLeavesNoFailSafeArmed(t *testing.T) {
	s := newCtrlSite(t)
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	var first, second error
	var fabricsAfterCut int
	_ = s.site.Run(ctx, func(ctx context.Context) error {
		cut, abort := context.WithCancel(ctx)
		defer abort()
		_, first = s.src.Pair(cut, sp.Manual, fabricui.PairOptions{Progress: func(st fabricui.PairingStep) {
			if st.Number == 18 && st.Sub == 1 && st.Phase == "started" {
				abort() // the request ends here
			}
		}})
		fi, _ := s.dev.Node().FabricInfos(ctx)
		fabricsAfterCut = len(fi)
		_, second = s.src.Pair(ctx, sp.Manual, fabricui.PairOptions{})
		return nil
	})
	if first == nil {
		t.Fatal("the cut commissioning succeeded; the test cut too late")
	}
	if fabricsAfterCut != 0 {
		t.Errorf("the device holds %d fabric(s) after the cut, want the fail-safe's rollback", fabricsAfterCut)
	}
	if second != nil {
		t.Fatalf("the attempt after the cut: %v", second)
	}
	if n := len(s.src.Nodes()); n != 1 {
		t.Fatalf("the source lists %d nodes, want 1", n)
	}
}
