// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"errors"
	"testing"
)

// The port's trace is newest first, as go-fabric's ring returns it; the
// memory controller's is too (UR-P1-adpt-30).
func TestMemoryController_DiagnosticEventsNewestFirst(t *testing.T) {
	events := NewMemoryController().DiagnosticEvents()
	if len(events) < 2 {
		t.Fatalf("%d events", len(events))
	}
	for i := 1; i < len(events); i++ {
		if events[i-1].At.Before(events[i].At) {
			t.Fatalf("event %d (%s) is older than event %d (%s)", i-1, events[i-1].At, i, events[i].At)
		}
	}
}

// A consent whose vendor or product differs from the node's is an
// invalid argument, as go-fabric's softwareupdate.ErrMismatch through the
// adapter (UR-P1-err-63); the matching one is recorded.
func TestMemoryController_ConsentRefusesAMismatch(t *testing.T) {
	u := NewMemoryController().SoftwareUpdates()
	if err := u.Consent(t.Context(), 0x11, 0xFFF1, 0x9999, 2); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("product mismatch: %v", err)
	}
	if err := u.Consent(t.Context(), 0x11, 0x1234, 0x8001, 2); !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("vendor mismatch: %v", err)
	}
	if err := u.Consent(t.Context(), 0x11, 0xFFF1, 0x8001, 2); err != nil {
		t.Fatalf("matching consent: %v", err)
	}
}

// What the memory controller hands out is the caller's own: editing it
// in place leaves the controller's state alone, and so does editing what
// a Write was given (UR-P1-conc-70).
func TestMemoryController_ReturnsNoAliases(t *testing.T) {
	m := NewMemoryController()
	cn, _ := m.Node(0x11)
	path := AttributePath{Endpoint: 1, Cluster: 0x0006, Attribute: 0x0000}

	vals, err := cn.Read(t.Context(), []AttributePath{path}, false)
	if err != nil || len(vals) != 1 || len(vals[0].Value) == 0 {
		t.Fatalf("read %v, %v", vals, err)
	}
	want := string(vals[0].Value)
	vals[0].Value[0] = 'X'
	if again, _ := cn.Read(t.Context(), []AttributePath{path}, false); string(again[0].Value) != want {
		t.Fatalf("Read aliases the stored value: %s", again[0].Value)
	}

	in := json.RawMessage(`true`)
	if _, err := cn.Write(t.Context(), path, in, 0); err != nil {
		t.Fatal(err)
	}
	in[0] = 'X'
	if again, _ := cn.Read(t.Context(), []AttributePath{path}, false); string(again[0].Value) != "true" {
		t.Fatalf("Write keeps the caller's bytes: %s", again[0].Value)
	}

	eps := cn.Endpoints()
	var ep *NodeEndpoint
	for i := range eps {
		if len(eps[i].Clusters) > 0 && len(eps[i].DeviceTypes) > 0 {
			ep = &eps[i]
			break
		}
	}
	if ep == nil {
		t.Fatal("no endpoint with clusters and device types")
	}
	number, cluster, dt := ep.Number, ep.Clusters[0].ID, ep.DeviceTypes[0].ID
	ep.Clusters[0].ID, ep.DeviceTypes[0].ID = 0xDEAD, 0xBEEF
	for _, e := range cn.Endpoints() {
		if e.Number == number && (e.Clusters[0].ID != cluster || e.DeviceTypes[0].ID != dt) {
			t.Fatalf("Endpoints aliases the stored endpoint: %+v", e)
		}
	}

	rec := cn.Record()
	if len(rec.Addresses) == 0 {
		t.Fatal("no address")
	}
	addr := rec.Addresses[0]
	rec.Addresses[0] = "edited"
	if cn.Record().Addresses[0] != addr {
		t.Fatal("Record aliases the stored addresses")
	}

	props := cn.Properties()
	if len(props.DeviceTypes) == 0 {
		t.Fatal("no device types")
	}
	first := props.DeviceTypes[0]
	props.DeviceTypes[0] = 0xBEEF
	if cn.Properties().DeviceTypes[0] != first {
		t.Fatal("Properties aliases the stored device types")
	}
}
