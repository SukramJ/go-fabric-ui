// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/SukramJ/go-fabric/cluster/spec/accesscontrol"
	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/controller/commissioner"
	"github.com/SukramJ/go-fabric/controller/softwareupdate"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// A dataset's network name comes from go-fabric's decoder; one that does
// not decode, or names no network, is an invalid credential.
func TestThreadDatasetName(t *testing.T) {
	// Network Name TLV (3) "Mesh", PAN ID TLV (1) 0x1234.
	ds, _ := hex.DecodeString("03044d657368" + "01021234")
	if name, err := ThreadDatasetName(ds); err != nil || name != "Mesh" {
		t.Fatalf("name %q, %v", name, err)
	}
	for _, bad := range []string{"0304", "01021234"} {
		b, _ := hex.DecodeString(bad)
		if _, err := ThreadDatasetName(b); !errors.Is(err, fabricui.ErrInvalidCredential) {
			t.Fatalf("%s: %v", bad, err)
		}
	}
	store := fabricui.NewMemoryCredentials(ThreadDatasetName)
	if err := store.SetThread(t.Context(), "t", []byte{0x03}); !errors.Is(err, fabricui.ErrInvalidCredential) {
		t.Fatalf("store accepted a broken dataset: %v", err)
	}
}

// Wrap runs the host's hook, then the stream's; a nil side is left out.
func TestHooks_WrapChainsTheHostsHooks(t *testing.T) {
	h := NewHooks()
	var host []string
	next := func(string) client.Hooks {
		return client.Hooks{
			Changed:                   func() { host = append(host, "changed") },
			SubscriptionStatusChanged: func(bool) { host = append(host, "subscription") },
			ClusterInstalled:          func(*client.Endpoint, *client.Cluster) { host = append(host, "cluster") },
		}
	}
	got := h.Wrap(next)("n1")
	got.Changed()
	got.SubscriptionStatusChanged(true)
	got.ClusterInstalled(nil, nil)
	got.ConnectionStateChanged(client.Connected)
	got.Commissioned()
	if len(host) != 3 {
		t.Fatalf("host hooks %v", host)
	}
	if h.Wrap(nil)("n1").Commissioned == nil {
		t.Fatal("Wrap(nil) has no hooks")
	}
	// Unbound hooks drop their events.
	if _, ok := h.nodeID("n1"); ok {
		t.Fatal("an unbound hook resolved a node")
	}
}

func TestConnectionOf(t *testing.T) {
	for s, want := range map[client.ConnectionState]fabricui.ConnectionState{
		client.Connected:                 fabricui.ConnectionConnected,
		client.Disconnected:              fabricui.ConnectionDisconnected,
		client.Reconnecting:              fabricui.ConnectionReconnecting,
		client.WaitingForDeviceDiscovery: fabricui.ConnectionWaitingForDiscovery,
	} {
		if got := connectionOf(s); got != want {
			t.Errorf("%v: %s", s, got)
		}
	}
}

// The ACL list carries our fabric's index in every entry and null for an
// absent member, the wildcard; the routes' ids are the definition's.
func TestACLList(t *testing.T) {
	ep := uint16(1)
	list := aclList([]fabricui.ACLEntry{
		{Privilege: 5, AuthMode: 2, Subjects: []uint64{0x1B669}},
		{Privilege: 3, AuthMode: 2, Targets: []fabricui.ACLTarget{{Endpoint: &ep}}},
	}, 3)
	if len(list) != 2 || list[0].FabricIndex != 3 || list[0].Targets.Null != true || list[0].Subjects.Value[0] != 0x1B669 {
		t.Fatalf("entry 0 %+v", list[0])
	}
	if list[1].Subjects.Null != true || list[1].Targets.Value[0].Endpoint.Value != 1 || !list[1].Targets.Value[0].Cluster.Null {
		t.Fatalf("entry 1 %+v", list[1])
	}
	if accesscontrol.ClusterID != 0x001F || accesscontrol.AttrAcl != 0 {
		t.Fatal("AccessControl ids moved")
	}
}

func TestIdentifierOf(t *testing.T) {
	for _, f := range []fabricui.DiscoverFilter{
		{Kind: "any"},
		{Kind: "long_discriminator", Value: 3840},
		{Kind: "short_discriminator", Value: 15},
		{Kind: "vendor", Value: 0xFFF1},
		{Kind: "device_type", Value: 0x10A},
	} {
		if _, err := identifierOf(f); err != nil {
			t.Errorf("%+v: %v", f, err)
		}
	}
	for _, f := range []fabricui.DiscoverFilter{
		{Kind: "long_discriminator", Value: 0x1000},
		{Kind: "short_discriminator", Value: 16},
		{Kind: "vendor", Value: 0x10000},
		{Kind: "colour"},
	} {
		if _, err := identifierOf(f); !errors.Is(err, fabricui.ErrInvalidArgument) {
			t.Errorf("%+v: %v", f, err)
		}
	}
}

// termsOf is the request's acknowledgement, else the template's; the
// host's Edit is not involved (it runs after, in go-fabric's Pair).
func TestTermsOf(t *testing.T) {
	tmpl := &commissioner.TermsAndConditionsAcknowledgement{Version: 9, Acknowledgements: 9}
	got := termsOf(&fabricui.TermsAcknowledgement{Version: 1, Acknowledgements: 3}, tmpl)
	if got == nil || *got != (commissioner.TermsAndConditionsAcknowledgement{Version: 1, Acknowledgements: 3}) {
		t.Fatalf("request's terms: %+v", got)
	}
	if termsOf(nil, tmpl) != tmpl {
		t.Fatal("without the request's terms the template's are not kept")
	}
	if termsOf(nil, nil) != nil {
		t.Fatal("no terms anywhere is an acknowledgement")
	}
}

// commissionError makes go-fabric's *commissioner.TermsRequiredError, also
// wrapped, the port's *ErrTermsRequired with its version and deadline.
func TestCommissionErrorTermsRequired(t *testing.T) {
	deadline := uint32(852076800)
	err := commissionError(fmt.Errorf("pair: %w", &commissioner.TermsRequiredError{Version: 2, Deadline: &deadline}))
	var tr *fabricui.ErrTermsRequired
	if !errors.As(err, &tr) || tr.Version != 2 || tr.Deadline == nil || *tr.Deadline != deadline {
		t.Fatalf("%v", err)
	}
	if errors.Is(err, fabricui.ErrNodeFailed) {
		t.Fatal("a terms requirement is a node failure")
	}
}

// bleFor is brief U8 §1's table: every row, and the precedence when
// several reasons hold — network_only before code_without_ble before
// no_network. Built from PairingCode values (bit 1 of a QR code's
// discovery capabilities is BLE, client/discovery/session.go:73-76 at
// v0.18.0), no parsing.
func TestBLEFor(t *testing.T) {
	manual := controller.PairingCode{Manual: true}
	qrBLE := controller.PairingCode{Capabilities: 2}
	qrIP := controller.PairingCode{Capabilities: 4}
	used := fabricui.BLEDecision{Used: true}
	skip := func(r fabricui.BLESkipReason) fabricui.BLEDecision { return fabricui.BLEDecision{Reason: r} }
	for _, c := range []struct {
		name                 string
		central              bool
		pc                   controller.PairingCode
		networkOnly, network bool
		want                 fabricui.BLEDecision
	}{
		{"no central, manual, network", false, manual, false, true, skip(fabricui.BLESkipNoCentral)},
		{"no central, network only", false, qrBLE, true, false, skip(fabricui.BLESkipNoCentral)},
		{"network only, manual", true, manual, true, true, skip(fabricui.BLESkipNetworkOnly)},
		{"network only over a QR without BLE", true, qrIP, true, true, skip(fabricui.BLESkipNetworkOnly)},
		{"QR without BLE", true, qrIP, false, true, skip(fabricui.BLESkipCode)},
		{"QR without BLE and no network", true, qrIP, false, false, skip(fabricui.BLESkipCode)},
		{"manual, no network", true, manual, false, false, skip(fabricui.BLESkipNoNetwork)},
		{"QR with BLE, no network", true, qrBLE, false, false, skip(fabricui.BLESkipNoNetwork)},
		{"manual, network", true, manual, false, true, used},
		{"QR with BLE, network", true, qrBLE, false, true, used},
	} {
		if got := bleFor(c.central, c.pc, c.networkOnly, c.network); got != c.want {
			t.Errorf("%s: %+v, want %+v", c.name, got, c.want)
		}
	}
}

// progressFor reports each step to the template's hook, then to the
// port's in the port's shape, the phase by name.
func TestProgressFor(t *testing.T) {
	var host []controller.PairingStep
	var port []fabricui.PairingStep
	fn := progressFor(func(s controller.PairingStep) { host = append(host, s) }, func(s fabricui.PairingStep) { port = append(port, s) })
	fn(controller.PairingStep{Number: 2, Sub: 2, Name: commissioner.StepPASE, Phase: commissioner.StepStarted})
	fn(controller.PairingStep{Number: 2, Sub: 2, Name: commissioner.StepPASE, Phase: commissioner.StepFailed, Err: errors.New("refused")})
	fn(controller.PairingStep{Number: 8, Sub: 3, Name: "GeneralCommissioning.ConfigureTCAcknowledgements", Phase: commissioner.StepSkipped})
	want := []fabricui.PairingStep{
		{Number: 2, Sub: 2, Name: "PASE", Phase: "started"},
		{Number: 2, Sub: 2, Name: "PASE", Phase: "failed"},
		{Number: 8, Sub: 3, Name: "GeneralCommissioning.ConfigureTCAcknowledgements", Phase: "skipped"},
	}
	if len(host) != 3 || !slices.Equal(port, want) {
		t.Fatalf("host %d, port %+v", len(host), port)
	}
	if progressFor(nil, nil) != nil {
		t.Fatal("no hook is a hook")
	}
}

// A manager's event is an update event with its kind, node, target
// version — none for a check of every node — and error text.
func TestUpdateEvent(t *testing.T) {
	peer := softwareupdate.Peer{NodeID: 0x11}
	target := softwareupdate.Target{VendorID: 0xFFF1, ProductID: 0x8001, TargetSoftwareVersion: 2}
	for _, c := range []struct {
		in   softwareupdate.UpdateEvent
		kind string
		ver  uint32
		err  string
	}{
		{softwareupdate.UpdateEvent{Kind: softwareupdate.UpdateAvailable, Peer: peer, Target: target}, "available", 2, ""},
		{softwareupdate.UpdateEvent{Kind: softwareupdate.UpdateCheckFinished, Err: errors.New("ledger down")}, "check_finished", 0, "ledger down"},
		{softwareupdate.UpdateEvent{Kind: softwareupdate.UpdateFailed, Peer: peer, Target: target, Err: softwareupdate.ErrUpdateCancelled}, "failed", 2, softwareupdate.ErrUpdateCancelled.Error()},
		{softwareupdate.UpdateEvent{Kind: softwareupdate.UpdateTransferStarted, Peer: peer, Target: target}, "transfer_started", 2, ""},
	} {
		got := updateEvent(&c.in)
		var ver uint32
		if got.UpdateVersion != nil {
			ver = *got.UpdateVersion
		}
		if got.Type != fabricui.EventUpdate || got.Node != c.in.Peer.NodeID || got.UpdateKind != c.kind || ver != c.ver || got.UpdateErr != c.err {
			t.Errorf("%v: %+v", c.in.Kind, got)
		}
	}
}
