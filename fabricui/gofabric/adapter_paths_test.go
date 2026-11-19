// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/SukramJ/go-fabric/controller/client"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// nodeError maps an interaction's error to the port's kinds: a value that
// does not fit its type is an invalid argument (400), an ICD refusal over
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
		{"icd multi-admin", fmt.Errorf("register: %w", multiAdmin), fabricui.ErrConflict},
		{"other", errors.New("timeout"), fabricui.ErrNodeFailed},
		{"cancelled", context.Canceled, context.Canceled},
	} {
		got := nodeError(c.in)
		if !errors.Is(got, c.want) || !errors.Is(got, c.in) {
			t.Errorf("%s: %v, want %v wrapping %v", c.name, got, c.want, c.in)
		}
		for _, other := range []error{fabricui.ErrInvalidArgument, fabricui.ErrConflict, fabricui.ErrNodeFailed} {
			if !errors.Is(other, c.want) && errors.Is(got, other) {
				t.Errorf("%s: %v is also %v", c.name, got, other)
			}
		}
	}
	if nodeError(nil) != nil {
		t.Error("nodeError(nil) is not nil")
	}
}
