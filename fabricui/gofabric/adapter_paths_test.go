// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"thingfabric.dev/go-fabric/controller/client"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// nodeError maps an interaction's error to the port's kinds: a value that
// does not fit its type, or a request go-fabric refuses before sending
// it, is an invalid argument (400), an ICD refusal over
// other administrators a conflict (409), a cancellation stays itself,
// anything else is the node's failure (502) (UR-P3-test-30).
func TestNodeError_MapsTheKinds(t *testing.T) {
	multiAdmin := &client.ICDMultiAdminError{AdminVendorIDs: []uint16{0xFFF1, 0xFFF2}}
	for _, c := range []struct {
		name string
		in   error
		want error
	}{
		{"json value", fmt.Errorf("write: %w", client.ErrJSONValue), fabricui.ErrInvalidArgument},
		// Refused before sending: the request's error (UR-P2-err-37).
		{"request too large", fmt.Errorf("write: %w", client.ErrRequestTooLarge), fabricui.ErrInvalidArgument},
		{"malformed request", fmt.Errorf("write: %w", client.ErrMalformedRequest), fabricui.ErrInvalidArgument},
		{"icd multi-admin", fmt.Errorf("register: %w", multiAdmin), fabricui.ErrICDMultiAdmin},
		{"icd refused by the client", fmt.Errorf("register: %w", client.ErrICDOperation), fabricui.ErrConflict},
		{"node disabled", fmt.Errorf("read: %w", client.ErrDisabled), fabricui.ErrConflict},
		{"fabric operation in progress", fmt.Errorf("remove: %w", client.ErrFabricOperationInProgress), fabricui.ErrConflict},
		{"other", errors.New("timeout"), fabricui.ErrNodeFailed},
		{"cancelled", context.Canceled, context.Canceled},
	} {
		got := nodeError(c.in)
		if !errors.Is(got, c.want) || !errors.Is(got, c.in) {
			t.Errorf("%s: %v, want %v wrapping %v", c.name, got, c.want, c.in)
		}
		for _, other := range []error{fabricui.ErrInvalidArgument, fabricui.ErrConflict, fabricui.ErrICDMultiAdmin, fabricui.ErrNodeFailed} {
			if !errors.Is(c.want, other) && errors.Is(got, other) {
				t.Errorf("%s: %v is also %v", c.name, got, other)
			}
		}
	}
	if nodeError(nil) != nil {
		t.Error("nodeError(nil) is not nil")
	}
}

// An event's epoch timestamp is the node's data: one above the int64
// range is left out, not cast to a negative time (UR-P4-ui-302).
func TestEpochMS_LeavesOutAValueOutsideInt64(t *testing.T) {
	for _, c := range []struct {
		in   *uint64
		want *int64
	}{
		{nil, nil},
		{ptrTo(uint64(1_700_000_000_000)), ptrTo(int64(1_700_000_000_000))},
		{ptrTo(uint64(math.MaxInt64)), ptrTo(int64(math.MaxInt64))},
		{ptrTo(uint64(math.MaxInt64) + 1), nil},
		{ptrTo(uint64(math.MaxUint64)), nil},
	} {
		got := epochMS(c.in)
		if (got == nil) != (c.want == nil) || (got != nil && *got != *c.want) {
			t.Errorf("epochMS(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func ptrTo[T any](v T) *T { return &v }

// The commissioning's bound after the search: the smaller positive of
// the port's and the host template's (UR-P4-bound-400).
func TestCommissionBound(t *testing.T) {
	for _, c := range []struct{ port, template, want time.Duration }{
		{0, 0, 0},
		{600 * time.Second, 0, 600 * time.Second},
		{0, 30 * time.Second, 30 * time.Second},
		{600 * time.Second, 30 * time.Second, 30 * time.Second},
		{30 * time.Second, 600 * time.Second, 30 * time.Second},
	} {
		if got := commissionBound(c.port, c.template); got != c.want {
			t.Errorf("commissionBound(%v, %v) = %v, want %v", c.port, c.template, got, c.want)
		}
	}
}
