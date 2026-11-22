// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/controller"
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
// controller/pair_commission_timeout_test.go:19-71 at v0.24.1). The
// handler's 600 s does not replace a host template's shorter bound
// (UR-P4-bound-400).
func TestCommissionTimeout_ReachesGoFabric(t *testing.T) {
	for name, c := range map[string]struct {
		discovered     bool
		port, template time.Duration
	}{
		"code":                         {port: 5 * time.Second},
		"discovered":                   {discovered: true, port: 5 * time.Second},
		"code, the host's bound stays": {port: 600 * time.Second, template: 5 * time.Second},
	} {
		t.Run(name, func(t *testing.T) {
			s := newCtrlSiteWith(t, ctrlSiteOptions{pair: func(p *controller.PairOptions) { p.CommissionTimeout = c.template }})
			sp, err := s.dev.Node().SetupPayload()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var once sync.Once
			o := fabricui.PairOptions{CommissionTimeout: c.port, Progress: func(st fabricui.PairingStep) {
				if st.Name == "PASE" && st.Phase == "succeeded" {
					once.Do(func() { s.site.Net().Set(fabrictest.NetOptions{Loss: 1}) })
				}
			}}
			var perr error
			if err := s.site.Run(ctx, func(ctx context.Context) error {
				if !c.discovered {
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

// The discovered path reaches the device at its discovered addresses: a
// host template that names a BLE central and an address (both
// ErrConfiguration beside PairDiscovered, go-fabric
// controller/controller.go:1265-1269 at v0.24.1) still commissions it
// (UR-P4-test-401).
func TestCommissionDiscovered_IgnoresTheTemplatesBLEAndAddress(t *testing.T) {
	s := newCtrlSiteWith(t, ctrlSiteOptions{ble: true, pair: func(p *controller.PairOptions) {
		p.Address = &net.UDPAddr{IP: net.IPv6loopback, Port: 5540}
	}})
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Minute)
	defer cancel()
	var perr error
	if err := s.site.Run(ctx, func(ctx context.Context) error {
		found, err := s.src.Discover(ctx, fabricui.DiscoverFilter{}, 5*time.Second)
		if err != nil {
			return err
		}
		if len(found) == 0 {
			return errors.New("no device discovered")
		}
		_, perr = s.src.CommissionDiscovered(ctx, found[0], sp.Passcode, fabricui.PairOptions{})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if perr != nil || len(s.src.Nodes()) != 1 {
		t.Fatalf("error %v, %d nodes; want the device commissioned", perr, len(s.src.Nodes()))
	}
}
