// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"thingfabric.dev/go-fabric/controller/client"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// SetOnboardingCode's refusals map by their go-fabric kind: the typed
// mismatch by its field bit with both values, an unreadable code with the
// wrapped reason, a node no longer commissioned or erased a conflict,
// anything else as it is (500).
func TestOnboardingError(t *testing.T) {
	for _, c := range []struct {
		bit   client.OnboardingChecks
		field string
	}{
		{client.OnboardingCheckedVendor, fabricui.OnboardingFieldVendor},
		{client.OnboardingCheckedProduct, fabricui.OnboardingFieldProduct},
		{client.OnboardingCheckedDiscriminator, fabricui.OnboardingFieldDiscriminator},
	} {
		err := onboardingError(&client.OnboardingMismatchError{Field: c.bit, Code: 3, Node: 15})
		mm, ok := errors.AsType[*fabricui.OnboardingMismatch](err)
		if !ok || mm.Field != c.field || mm.Code != 3 || mm.Node != 15 {
			t.Fatalf("bit %d: %v", c.bit, err)
		}
	}
	odd := &client.OnboardingMismatchError{Field: 64}
	if err := onboardingError(odd); !errors.Is(err, odd) {
		t.Fatalf("an unknown field: %v", err)
	}
	if _, ok := errors.AsType[*fabricui.OnboardingMismatch](onboardingError(odd)); ok {
		t.Fatal("an unknown field is mapped")
	}
	reason := errors.New("setup: bad check digit")
	inv, ok := errors.AsType[*fabricui.OnboardingInvalid](onboardingError(fmt.Errorf("%w: %w", client.ErrOnboardingCode, reason)))
	if !ok || inv.Reason != reason.Error() {
		t.Fatalf("invalid %v", inv)
	}
	if inv, _ := errors.AsType[*fabricui.OnboardingInvalid](onboardingError(client.ErrOnboardingCode)); inv == nil || inv.Reason != "not an onboarding code" {
		t.Fatalf("bare sentinel %v", inv)
	}
	for _, err := range []error{client.ErrNotCommissioned, client.ErrClosed} {
		if !errors.Is(onboardingError(err), fabricui.ErrConflict) {
			t.Fatalf("%v is no conflict", err)
		}
	}
	store := errors.New("store: disk full")
	if err := onboardingError(store); !errors.Is(err, store) || err.Error() != store.Error() {
		t.Fatal("a store error is changed")
	}
}

// Without ControllerOptions.OnboardingCodes the three methods serve
// nothing (the handler answers 501 before it calls them).
func TestOnboarding_OffByDefault(t *testing.T) {
	s := &source{}
	if s.Capabilities().OnboardingCodes || s.OnboardingInventory() != nil {
		t.Fatal("served without the option")
	}
	if _, err := s.SetOnboardingCode(context.Background(), 1, "34970112332"); !errors.Is(err, fabricui.ErrNotImplemented) {
		t.Fatalf("set: %v", err)
	}
	if err := s.ClearOnboardingCode(context.Background(), 1); !errors.Is(err, fabricui.ErrNotImplemented) {
		t.Fatalf("clear: %v", err)
	}
}

// R8 of brief U12, deterministically (U12-R-03): the inventory is sorted
// by node id whatever order go-fabric's registry lists it in.
func TestSortByNodeID(t *testing.T) {
	entries := []fabricui.OnboardingEntry{{NodeID: 0x30}, {NodeID: 0x05}, {NodeID: 0xFFFF_FFFF_0000_0001}, {NodeID: 0x11}}
	sortByNodeID(entries)
	got := make([]uint64, 0, len(entries))
	for _, e := range entries {
		got = append(got, e.NodeID)
	}
	if fmt.Sprint(got) != fmt.Sprint([]uint64{0x05, 0x11, 0x30, 0xFFFF_FFFF_0000_0001}) {
		t.Fatalf("order %x", got)
	}
}

// U12-R-07: a code stored on a node a fresh inventory no longer lists is
// no ErrNodeNotFound (the attach happened) but an error of no kind, 500.
func TestEntryOf(t *testing.T) {
	inv := []fabricui.OnboardingEntry{{NodeID: 1}, {NodeID: 2, HasCode: true}}
	if e, err := entryOf(inv, 2); err != nil || !e.HasCode {
		t.Fatalf("entry %v, %v", e, err)
	}
	_, err := entryOf(inv, 3)
	if err == nil || errors.Is(err, fabricui.ErrNodeNotFound) || errors.Is(err, fabricui.ErrConflict) {
		t.Fatalf("a gone node: %v", err)
	}
}
