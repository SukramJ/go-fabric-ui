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
	// bound is the CommissionTimeout of the last Pair.
	bound *time.Duration
}

func left(ctx context.Context) time.Duration {
	if d, ok := ctx.Deadline(); ok {
		return time.Until(d)
	}
	return 0
}

func (d deadlineSource) Pair(ctx context.Context, code string, o PairOptions) (ControllerNode, error) {
	*d.pair = left(ctx)
	*d.bound = o.CommissionTimeout
	return d.MemoryController.Pair(ctx, code, o)
}

func (d deadlineSource) CommissionDiscovered(ctx context.Context, dev Discovered, passcode uint32, o PairOptions) (ControllerNode, error) {
	*d.discovered = left(ctx)
	return d.MemoryController.CommissionDiscovered(ctx, dev, passcode, o)
}

// Without timeout_seconds a discovered device's commissioning is bounded
// at 600 s; a code's ctx is not, so go-fabric's one-minute bound on the
// search stays, and its commissioning after the search is bounded at
// 600 s by PairOptions.CommissionTimeout (decision 71, UR-P4-bound-02;
// decision 80). With timeout_seconds both are bounded by it alone.
func TestCommission_DefaultBounds(t *testing.T) {
	var pair, disc, bound time.Duration
	src := deadlineSource{MemoryController: NewMemoryController(), pair: &pair, discovered: &disc, bound: &bound}
	_, mux := newTestHandler(t, testOptions{controller: src})
	do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332"}`)
	do(t, mux, http.MethodPost, "/controller/commission", `{"discovered":{"instance_name":"1A2B3C4D5E6F7788"},"passcode":11223344}`)
	if pair != 0 {
		t.Errorf("Pair without timeout_seconds has a deadline in %v; go-fabric's search bound applies only without one", pair)
	}
	if bound != 600*time.Second {
		t.Errorf("Pair without timeout_seconds: CommissionTimeout %v, want 600 s", bound)
	}
	if disc < 590*time.Second || disc > 600*time.Second {
		t.Errorf("CommissionDiscovered without timeout_seconds: deadline in %v, want 600 s", disc)
	}
	do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":30}`)
	if pair < 25*time.Second || pair > 30*time.Second {
		t.Errorf("Pair with timeout_seconds 30: deadline in %v", pair)
	}
	if bound != 0 {
		t.Errorf("Pair with timeout_seconds 30: CommissionTimeout %v, want none beside the deadline", bound)
	}
}
