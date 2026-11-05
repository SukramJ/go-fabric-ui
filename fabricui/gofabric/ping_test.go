// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"fmt"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/controller/client"
)

// fakePing answers PingAddress for the addresses in rtt and fails every
// other one as client.Node.PingAddress does: ErrUnreachable naming it.
func fakePing(rtt map[netip.AddrPort]time.Duration) func(context.Context, netip.AddrPort) (time.Duration, error) {
	return func(_ context.Context, a netip.AddrPort) (time.Duration, error) {
		if d, ok := rtt[a]; ok {
			return d, nil
		}
		return 0, fmt.Errorf("%w: %s", client.ErrUnreachable, a)
	}
}

// Two addresses, one answering and one silent: one entry each in the
// record's order, the silent one with its reason, the result reachable
// with the answering address's round trip and measured per address.
func TestPingAddresses_OneReachableOneSilent(t *testing.T) {
	up, silent := netip.MustParseAddrPort("[fd00::11]:5540"), netip.MustParseAddrPort("192.168.1.11:5540")
	res, ok := pingAddresses(context.Background(), []netip.AddrPort{silent, up}, fakePing(map[netip.AddrPort]time.Duration{up: 3 * time.Millisecond}))
	if !ok || !res.PerAddress || !res.Reachable || res.RTT != 3*time.Millisecond || res.Error != "" || len(res.Addresses) != 2 {
		t.Fatalf("%v %+v", ok, res)
	}
	if a := res.Addresses[0]; a.Address != "192.168.1.11:5540" || a.Reachable || a.RTT != 0 || a.Error != "client: address unreachable: 192.168.1.11:5540" {
		t.Fatalf("silent %+v", a)
	}
	if a := res.Addresses[1]; a.Address != "[fd00::11]:5540" || !a.Reachable || a.RTT != 3*time.Millisecond || a.Error != "" {
		t.Fatalf("up %+v", a)
	}
}

// The fastest answering address is the result's round trip; no address
// answering is an unreachable result with a reason, not an error.
func TestPingAddresses_FastestAndNoneAnswering(t *testing.T) {
	a, b := netip.MustParseAddrPort("[fd00::1]:5540"), netip.MustParseAddrPort("[fd00::2]:5540")
	res, ok := pingAddresses(context.Background(), []netip.AddrPort{a, b}, fakePing(map[netip.AddrPort]time.Duration{a: 9 * time.Millisecond, b: 2 * time.Millisecond}))
	if !ok || !res.Reachable || res.RTT != 2*time.Millisecond {
		t.Fatalf("%v %+v", ok, res)
	}
	res, ok = pingAddresses(context.Background(), []netip.AddrPort{a, b}, fakePing(nil))
	if !ok || !res.PerAddress || res.Reachable || res.RTT != 0 || res.Error != "none of the node's 2 addresses answered" {
		t.Fatalf("%v %+v", ok, res)
	}
}

// No address, or a probe that fails for the node — no session, not
// commissioned, closed — sends the caller back to the session read.
func TestPingAddresses_FallsBackForTheNode(t *testing.T) {
	a := netip.MustParseAddrPort("[fd00::1]:5540")
	if _, ok := pingAddresses(context.Background(), nil, fakePing(nil)); ok {
		t.Fatal("no address measured per address")
	}
	for _, err := range []error{client.ErrNoSession, client.ErrNotCommissioned, client.ErrClosed} {
		fail := func(context.Context, netip.AddrPort) (time.Duration, error) { return 0, err }
		if _, ok := pingAddresses(context.Background(), []netip.AddrPort{a}, fail); ok {
			t.Fatalf("%v measured per address", err)
		}
	}
}

// At most pingParallel probes run at once.
func TestPingAddresses_Bounded(t *testing.T) {
	var addrs []netip.AddrPort
	for i := range 3 * pingParallel {
		addrs = append(addrs, netip.AddrPortFrom(netip.AddrFrom4([4]byte{10, 0, 0, byte(i + 1)}), 5540))
	}
	var running, peak atomic.Int32
	ping := func(context.Context, netip.AddrPort) (time.Duration, error) {
		n := running.Add(1)
		for {
			p := peak.Load()
			if n <= p || peak.CompareAndSwap(p, n) {
				break
			}
		}
		time.Sleep(2 * time.Millisecond)
		running.Add(-1)
		return time.Millisecond, nil
	}
	res, ok := pingAddresses(context.Background(), addrs, ping)
	if !ok || len(res.Addresses) != len(addrs) || peak.Load() > pingParallel {
		t.Fatalf("%v %d entries, peak %d", ok, len(res.Addresses), peak.Load())
	}
}
