// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// deadlineSource records the deadline of a commissioning's ctx: zero
// without one.
type deadlineSource struct {
	*MemoryController
	pair, discovered *time.Duration
}

func left(ctx context.Context) time.Duration {
	if d, ok := ctx.Deadline(); ok {
		return time.Until(d)
	}
	return 0
}

func (d deadlineSource) Pair(ctx context.Context, code string, o PairOptions) (ControllerNode, error) {
	*d.pair = left(ctx)
	return d.MemoryController.Pair(ctx, code, o)
}

func (d deadlineSource) CommissionDiscovered(ctx context.Context, dev Discovered, passcode uint32, o PairOptions) (ControllerNode, error) {
	*d.discovered = left(ctx)
	return d.MemoryController.CommissionDiscovered(ctx, dev, passcode, o)
}

// Without timeout_seconds a discovered device's commissioning is bounded
// at 600 s; a code's is not, so go-fabric's one-minute bound on the
// search stays (decision 71, UR-P4-bound-02). With timeout_seconds both
// are bounded by it.
func TestCommission_DefaultBoundOnTheDiscoveredPathOnly(t *testing.T) {
	var pair, disc time.Duration
	src := deadlineSource{MemoryController: NewMemoryController(), pair: &pair, discovered: &disc}
	_, mux := newTestHandler(t, testOptions{controller: src})
	do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332"}`)
	do(t, mux, http.MethodPost, "/controller/commission", `{"discovered":{"instance_name":"1A2B3C4D5E6F7788"},"passcode":11223344}`)
	if pair != 0 {
		t.Errorf("Pair without timeout_seconds has a deadline in %v; go-fabric's search bound applies only without one", pair)
	}
	if disc < 590*time.Second || disc > 600*time.Second {
		t.Errorf("CommissionDiscovered without timeout_seconds: deadline in %v, want 600 s", disc)
	}
	do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":30}`)
	if pair < 25*time.Second || pair > 30*time.Second {
		t.Errorf("Pair with timeout_seconds 30: deadline in %v", pair)
	}
}
