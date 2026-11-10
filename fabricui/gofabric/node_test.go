// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"net/netip"
	"testing"

	"github.com/SukramJ/go-fabric/controller/client"
)

// TestNodeRecord_OperationalHost: the record's proven operational host
// (go-fabric v0.20.0) reaches the port beside the commissionable
// advertisement's hostname, and a record stored before it stays empty.
func TestNodeRecord_OperationalHost(t *testing.T) {
	r := client.Record{
		FabricIndexOnPeer: 2,
		Addresses:         []client.NetworkAddress{{IP: netip.MustParseAddr("fd00::11"), Port: 5540}},
		Hostname:          "light-1.local.",
		OperationalHost:   "0123456789abcdef.local",
	}
	got := nodeRecord(r)
	if got.OperationalHost != "0123456789abcdef.local" || got.Hostname != "light-1.local." || got.FabricIndexOnPeer != 2 {
		t.Fatalf("record = %+v", got)
	}
	if len(got.Addresses) != 1 || got.Addresses[0] != "[fd00::11]:5540" {
		t.Fatalf("addresses = %v", got.Addresses)
	}
	r.OperationalHost = ""
	if got := nodeRecord(r); got.OperationalHost != "" {
		t.Fatalf("operational host = %q, want empty", got.OperationalHost)
	}
}
