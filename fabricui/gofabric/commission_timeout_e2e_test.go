// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/fabrictest"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// The port's CommissionTimeout reaches go-fabric's
// PairOptions.CommissionTimeout on both paths: a device that falls silent
// once PASE is established, under a ctx without a deadline, ends the
// commissioning at the bound with context.DeadlineExceeded — without the
// bound the step ends when its exchange gives up on the peer, after
// 13.5 s on the site — and nothing is commissioned (decision 80,
// go-fabric#554; go-fabric's own test of the bound,
// controller/pair_commission_timeout_test.go:19-71 at v0.24.1).
func TestCommissionTimeout_ReachesGoFabric(t *testing.T) {
	for _, path := range []string{"code", "discovered"} {
		t.Run(path, func(t *testing.T) {
			s := newCtrlSite(t)
			sp, err := s.dev.Node().SetupPayload()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var once sync.Once
			o := fabricui.PairOptions{CommissionTimeout: 5 * time.Second, Progress: func(st fabricui.PairingStep) {
				if st.Name == "PASE" && st.Phase == "succeeded" {
					once.Do(func() { s.site.Net().Set(fabrictest.NetOptions{Loss: 1}) })
				}
			}}
			var perr error
			if err := s.site.Run(ctx, func(ctx context.Context) error {
				if path == "code" {
					_, perr = s.src.Pair(ctx, sp.Manual, o)
					return nil
				}
				found, err := s.src.Discover(ctx, fabricui.DiscoverFilter{}, 5*time.Second)
				if err != nil {
					return err
				}
				if len(found) == 0 {
					return errors.New("no device discovered")
				}
				_, perr = s.src.CommissionDiscovered(ctx, found[0], sp.Passcode, o)
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if !errors.Is(perr, context.DeadlineExceeded) || len(s.src.Nodes()) != 0 {
				t.Fatalf("error %v, %d nodes; want the bound's DeadlineExceeded and none", perr, len(s.src.Nodes()))
			}
		})
	}
}
