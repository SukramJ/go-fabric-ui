// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/controller/topology"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// topologyLiteral is go-fabric's snapshot behind the
// controller-topology.200 fixture: a Thread router (17), a sleepy end
// device (18), an unreachable Wi-Fi station (19) with its access point,
// an Ethernet bridge (20), a Thread router (with its network name) and an
// end device no commissioned node matches, and a border router the
// registry knows (contract 0.8.0); a link of every strength. Build never
// yields a connection of strength none (a pair whose directions are all
// dead is dropped), so that one link is written by hand.
func topologyLiteral() topology.Topology {
	rloc := func(v uint16) *uint16 { return &v }
	rssi := func(v int8) *int8 { return &v }
	const pan, bssid = 0xDEAD00BEEF00CAFE, "AA:BB:CC:DD:EE:FF"
	return topology.Topology{
		CollectedAt: time.Date(2026, 10, 25, 14, 0, 0, 0, time.FixedZone("CEST", 2*3600)),
		Nodes: []topology.Node{
			{
				ID: "17", Kind: topology.KindMatter, NetworkType: topology.NetworkThread, NodeID: 0x11, Role: topology.RoleRouter, Available: true,
				ExtAddress: 0x1A2B3C4D5E6F7011, Rloc16: rloc(0x2C00), ExtPanID: pan, NetworkName: "Home", Addresses: []netip.Addr{netip.MustParseAddr("fd00::11")},
			},
			{
				ID: "18", Kind: topology.KindMatter, NetworkType: topology.NetworkThread, NodeID: 0x12, Role: topology.RoleSleepyEndDevice, Available: true,
				ExtAddress: 0x1A2B3C4D5E6F7012, Rloc16: rloc(0x2C01), ExtPanID: pan, NetworkName: "Home", Addresses: []netip.Addr{netip.MustParseAddr("fd00::12")},
			},
			{
				ID: "19", Kind: topology.KindMatter, NetworkType: topology.NetworkWiFi, NodeID: 0x13, Role: topology.RoleStation,
				SSID: "home-wifi", BSSID: bssid, Addresses: []netip.Addr{netip.MustParseAddr("192.168.1.13"), netip.MustParseAddr("fd00::13")},
			},
			{
				ID: "20", Kind: topology.KindMatter, NetworkType: topology.NetworkEthernet, NodeID: 0x14, Available: true, IsBridge: true,
				Addresses: []netip.Addr{netip.MustParseAddr("192.168.1.20")},
			},
			{ID: "unknown_0011223344556677", Kind: topology.KindThreadUnknown, NetworkType: topology.NetworkThread, Role: topology.RoleRouter, ExtAddress: 0x0011223344556677, ExtPanID: pan, NetworkName: "Home"},
			{ID: "unknown_8899AABBCCDDEEFF", Kind: topology.KindThreadUnknown, NetworkType: topology.NetworkThread, ExtAddress: 0x8899AABBCCDDEEFF, ExtPanID: pan},
			{ID: "ap_AABBCCDDEEFF", Kind: topology.KindWiFiAP, NetworkType: topology.NetworkWiFi, Role: topology.RoleAP, NetworkName: bssid, SSID: "home-wifi", BSSID: bssid},
			{
				ID: "br_CAFE0000CAFE0001", Kind: topology.KindBorderRouter, NetworkType: topology.NetworkThread, Role: topology.RoleRouter,
				ExtAddress: 0xCAFE0000CAFE0001, ExtPanID: pan, NetworkName: "Home", HostName: "otbr-kitchen", VendorName: "OpenThread", ModelName: "OTBR",
				LastSeen: time.Date(2026, 10, 25, 11, 55, 0, 0, time.UTC),
			},
		},
		Connections: []topology.Connection{
			{
				Source: "17", Target: "18", Network: topology.NetworkThread, Strength: topology.StrengthStrong,
				SourceToTarget: &topology.Direction{Strength: topology.StrengthStrong, LQI: 3, RSSI: rssi(-45)},
				TargetToSource: &topology.Direction{Strength: topology.StrengthMedium, LQI: 2, RSSI: rssi(-62)},
			},
			{
				Source: "17", Target: "unknown_0011223344556677", Network: topology.NetworkThread, Strength: topology.StrengthMedium,
				SourceToTarget: &topology.Direction{Strength: topology.StrengthMedium, LQI: 2, RSSI: rssi(-66)},
				TargetToSource: &topology.Direction{Strength: topology.StrengthWeak, LQI: 1, RSSI: rssi(-80)},
			},
			{
				Source: "17", Target: "unknown_8899AABBCCDDEEFF", Network: topology.NetworkThread, Strength: topology.StrengthWeak,
				SourceToTarget: &topology.Direction{Strength: topology.StrengthWeak, LQI: 1, RSSI: rssi(-88)},
			},
			{
				Source: "18", Target: "unknown_0011223344556677", Network: topology.NetworkThread, Strength: topology.StrengthNone,
				SourceToTarget: &topology.Direction{Strength: topology.StrengthNone},
			},
			{Source: "19", Target: "ap_AABBCCDDEEFF", Network: topology.NetworkWiFi, Strength: topology.StrengthUnknown},
			{
				Source: "17", Target: "br_CAFE0000CAFE0001", Network: topology.NetworkThread, Strength: topology.StrengthStrong,
				SourceToTarget: &topology.Direction{Strength: topology.StrengthStrong, LQI: 3, RSSI: rssi(-50)},
			},
		},
	}
}

// The adapter's graph of the literal, served by the handler, is the
// controller-topology.200 fixture: go-fabric's snapshot to the contract's
// JSON end to end (the handler test holds the port literal to the same
// fixture).
func TestNetworkTopologyOf_IsTheFixture(t *testing.T) {
	got := networkTopologyOf(topologyLiteral(), "Controller", map[uint64]string{0x11: "Light", 0x12: "Sensor", 0x14: "Bridge"})
	m := fabricui.NewMemoryController()
	m.Topology = &got
	rec := httptest.NewRecorder()
	fabricui.Handler(fabricui.Options{Controller: m}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fabricui.DefaultPrefix+"/controller/topology", http.NoBody))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body)
	}
	raw, err := os.ReadFile("../../contract/fixtures/controller-topology.200.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Response struct {
			Body any `json:"body"`
		} `json:"response"`
	}
	var body any
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body, fixture.Response.Body) {
		want, _ := json.Marshal(fixture.Response.Body)
		t.Fatalf("graph\n got: %s\nwant: %s", rec.Body.Bytes(), want)
	}
}

// A border router is ok while the registry saw it within its expiry
// before the snapshot, unknown after it or never seen; without a host
// name its extended address labels it.
func TestBorderRouterOf(t *testing.T) {
	at := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	br := topology.Node{ID: "br_00000000000000AB", Kind: topology.KindBorderRouter, ExtAddress: 0xAB, LastSeen: at.Add(-borderRouterExpiry)}
	if n := borderRouterOf(&br, at); n.State != "ok" || n.Label != "00000000000000AB" || n.Kind != "border_router" {
		t.Fatalf("at the expiry %+v", n)
	}
	br.LastSeen = at.Add(-borderRouterExpiry - time.Second)
	if n := borderRouterOf(&br, at); n.State != "unknown" {
		t.Fatalf("past the expiry %+v", n)
	}
	br.LastSeen = time.Time{}
	if n := borderRouterOf(&br, at); n.State != "unknown" || len(n.Details) != 1 {
		t.Fatalf("never seen %+v", n)
	}
}

// A kind go-fabric adds after this adapter is drawn by its network with
// its kind as a detail, never dropped.
func TestNetworkTopologyOf_KeepsAnUnknownKind(t *testing.T) {
	g := networkTopologyOf(topology.Topology{Nodes: []topology.Node{
		{ID: "x_1", Kind: "future", NetworkType: topology.NetworkThread, Role: topology.RoleRouter, ExtAddress: 0x1},
		{ID: "x_2", Kind: "future", NetworkType: topology.NetworkWiFi},
		{ID: "x_3", Kind: "future"},
	}}, "Controller", nil)
	if len(g.Nodes) != 4 {
		t.Fatalf("nodes %+v", g.Nodes)
	}
	want := []struct{ kind, label string }{{"thread_router", "0000000000000001"}, {"wifi_ap", "x_2"}, {"ethernet", "x_3"}}
	for i, w := range want {
		n := g.Nodes[i+1]
		if n.Kind != w.kind || n.Label != w.label || n.State != "unknown" || n.Details[0] != (fabricui.TopologyDetail{Key: "kind", Value: "future"}) {
			t.Errorf("%d: %+v", i, n)
		}
	}
}

// A link's quality comes from the direction its strength comes from; the
// label shows both directions when they differ — RSSI on Wi-Fi, LQI on
// Thread — and an empty strength is unknown.
func TestLinkOf(t *testing.T) {
	rssi := func(v int8) *int8 { return &v }
	ids := map[string]string{"1": "peer:0000000000000001"}
	wifi := linkOf(&topology.Connection{
		Source: "1", Target: "ap_X", Network: topology.NetworkWiFi, Strength: topology.StrengthStrong,
		SourceToTarget: &topology.Direction{Strength: topology.StrengthMedium, RSSI: rssi(-80)},
		TargetToSource: &topology.Direction{Strength: topology.StrengthStrong, RSSI: rssi(-60)},
	}, ids)
	if wifi.Source != "peer:0000000000000001" || wifi.Target != "ap_X" || wifi.LQI != nil || wifi.RSSI == nil || *wifi.RSSI != -60 || wifi.Label != "RSSI -80/-60 dBm" {
		t.Fatalf("wifi %+v", wifi)
	}
	thread := linkOf(&topology.Connection{
		Source: "a", Target: "b", Network: topology.NetworkThread, Strength: topology.StrengthMedium,
		TargetToSource: &topology.Direction{Strength: topology.StrengthMedium, LQI: 2},
	}, ids)
	if thread.LQI == nil || *thread.LQI != 2 || thread.RSSI != nil || thread.Label != "" {
		t.Fatalf("thread %+v", thread)
	}
	if bare := linkOf(&topology.Connection{Source: "a", Target: "b"}, ids); bare.Strength != "unknown" || bare.LQI != nil {
		t.Fatalf("bare %+v", bare)
	}
}

// An empty snapshot is the controller alone; an unnamed fabric labels it
// "Controller" (NetworkTopology's fallback is tested end to end).
func TestNetworkTopologyOf_Empty(t *testing.T) {
	at := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	got := networkTopologyOf(topology.Topology{CollectedAt: at}, "Controller", nil)
	want := fabricui.NetworkTopology{
		CollectedAt: at,
		Nodes: []fabricui.NetworkNode{{
			ID: "node", Kind: "node", Label: "Controller", State: "ok",
			Details: []fabricui.TopologyDetail{{Key: "nodes", Value: "0"}, {Key: "collected_at", Value: "2026-11-03T09:00:00Z"}},
		}},
		Connections: []fabricui.NetworkConnection{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

// Without a service the route is 501, and the topology subscription is
// released by Hooks.Close, or by closeTopology without hooks; both are
// idempotent.
func TestTopologySubscription(t *testing.T) {
	s := &source{}
	if _, err := s.NetworkTopology(t.Context(), false); !errors.Is(err, fabricui.ErrNotImplemented) {
		t.Fatalf("no service: %v", err)
	}
	released := 0
	s.unwatch = func() { released++ }
	h := NewHooks()
	h.mu.Lock()
	h.src = s
	h.mu.Unlock()
	h.Close()
	h.Close()
	s.closeTopology()
	if released != 1 {
		t.Fatalf("released %d times", released)
	}
	// Hooks closed before ControllerOf binds: the subscription goes at once.
	s2 := &source{}
	s2.unwatch = func() { released++ }
	h2 := NewHooks()
	h2.Close()
	h2.bind(s2)
	if released != 2 {
		t.Fatalf("closed hooks kept the subscription (%d)", released)
	}
}
