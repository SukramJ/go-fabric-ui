// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"thingfabric.dev/go-fabric/controller/client"

	"github.com/SukramJ/go-fabric-ui/fabricui"
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
