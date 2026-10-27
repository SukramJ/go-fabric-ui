// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/SukramJ/go-fabric/cluster/spec/accesscontrol"
	"github.com/SukramJ/go-fabric/controller/client"

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
