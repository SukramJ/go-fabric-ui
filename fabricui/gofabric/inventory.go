// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"context"
	"errors"
	"slices"

	"thingfabric.dev/go-fabric/controller"
	"thingfabric.dev/go-fabric/controller/client"

	"thingfabric.dev/go-fabric-ui/fabricui"
)

// The onboarding inventory (contract 0.13.0): go-fabric's
// controller.Controller.Inventory and client.Node's SetOnboardingCode and
// ClearOnboardingCode. Served only with ControllerOptions.OnboardingCodes;
// without it the handler answers 501 before it reaches these methods,
// which answer nothing and ErrNotImplemented anyway.

// OnboardingInventory implements [fabricui.ControllerSource]: the
// controller's inventory, sorted by node id as Nodes is (go-fabric lists
// it in the registry's order).
func (s *source) OnboardingInventory() []fabricui.OnboardingEntry {
	if !s.opts.OnboardingCodes {
		return nil
	}
	inv := s.ctrl.Inventory()
	out := make([]fabricui.OnboardingEntry, 0, len(inv))
	for i := range inv {
		out = append(out, onboardingEntryOf(&inv[i]))
	}
	sortByNodeID(out)
	return out
}

// sortByNodeID orders the entries by node id, the order of the adapter's
// Nodes and so of GET /controller/nodes (R8 of brief U12).
func sortByNodeID(entries []fabricui.OnboardingEntry) {
	slices.SortFunc(entries, func(a, b fabricui.OnboardingEntry) int {
		switch {
		case a.NodeID < b.NodeID:
			return -1
		case a.NodeID > b.NodeID:
			return 1
		}
		return 0
	})
}

// onboardingEntryOf is the port's entry of go-fabric's; a revealed code
// is bound to no variable.
func onboardingEntryOf(e *controller.InventoryEntry) fabricui.OnboardingEntry {
	out := fabricui.OnboardingEntry{
		NodeID: e.NodeID, Label: e.Label,
		VendorName: e.VendorName, ProductName: e.ProductName, SerialNumber: e.SerialNumber,
		VendorID: e.VendorID, ProductID: e.ProductID,
		HasCode:    e.HasCode,
		QRCode:     fabricui.NewSecretText(e.QRCode.Reveal()),
		ManualCode: fabricui.NewSecretText(e.ManualCode.Reveal()),
		Checks: fabricui.OnboardingChecks{
			Vendor:        e.Checks&client.OnboardingCheckedVendor != 0,
			Product:       e.Checks&client.OnboardingCheckedProduct != 0,
			Discriminator: e.Checks&client.OnboardingCheckedDiscriminator != 0,
		},
		StoredAt: e.StoredAt,
	}
	switch e.Source {
	case client.OnboardingPaired:
		out.Source = fabricui.OnboardingPaired
	case client.OnboardingAttached:
		out.Source = fabricui.OnboardingAttached
	}
	return out
}

// SetOnboardingCode implements [fabricui.ControllerSource]:
// client.Node.SetOnboardingCode, then the node's entry from a fresh
// inventory.
func (s *source) SetOnboardingCode(_ context.Context, id uint64, code string) (fabricui.OnboardingEntry, error) {
	if !s.opts.OnboardingCodes {
		return fabricui.OnboardingEntry{}, fabricui.ErrNotImplemented
	}
	n, ok := s.lookup(id)
	if !ok {
		return fabricui.OnboardingEntry{}, fabricui.ErrNodeNotFound
	}
	if _, err := n.n.SetOnboardingCode(code); err != nil {
		return fabricui.OnboardingEntry{}, onboardingError(err)
	}
	return entryOf(s.OnboardingInventory(), id)
}

// errStoredNodeGone is a code stored on a node that a fresh inventory no
// longer lists (decommissioned, or left without a peer address, between
// the attach and the read). The code is stored, so the attach is no
// "no such node"; the error is of no kind (500).
var errStoredNodeGone = errors.New("gofabric: the code is stored, but its node is not in the inventory")

// entryOf is the node's entry of a fresh inventory, after an attach.
func entryOf(inv []fabricui.OnboardingEntry, id uint64) (fabricui.OnboardingEntry, error) {
	for i := range inv {
		if inv[i].NodeID == id {
			return inv[i], nil
		}
	}
	return fabricui.OnboardingEntry{}, errStoredNodeGone
}

// ClearOnboardingCode implements [fabricui.ControllerSource].
func (s *source) ClearOnboardingCode(_ context.Context, id uint64) error {
	if !s.opts.OnboardingCodes {
		return fabricui.ErrNotImplemented
	}
	n, ok := s.lookup(id)
	if !ok {
		return fabricui.ErrNodeNotFound
	}
	return n.n.ClearOnboardingCode()
}

// onboardingError maps SetOnboardingCode's refusals to the port's kinds,
// by their go-fabric kind and not through nodeError, whose default is the
// node's failure (502) — wrong for a local store. A mismatch carries its
// field and both values; an invalid code the text of the setup error
// go-fabric wraps, which names the rule and never the passcode. No error
// text is cut. A store error is returned as it is (500).
func onboardingError(err error) error {
	if mm, ok := errors.AsType[*client.OnboardingMismatchError](err); ok {
		out := &fabricui.OnboardingMismatch{Code: mm.Code, Node: mm.Node}
		switch mm.Field {
		case client.OnboardingCheckedVendor:
			out.Field = fabricui.OnboardingFieldVendor
		case client.OnboardingCheckedProduct:
			out.Field = fabricui.OnboardingFieldProduct
		case client.OnboardingCheckedDiscriminator:
			out.Field = fabricui.OnboardingFieldDiscriminator
		default:
			return err
		}
		return out
	}
	switch {
	case errors.Is(err, client.ErrOnboardingCode):
		return &fabricui.OnboardingInvalid{Reason: invalidReason(err)}
	case errors.Is(err, client.ErrNotCommissioned), errors.Is(err, client.ErrClosed):
		return fabricui.ErrConflict
	}
	return err
}

// invalidReason is the text of the setup error go-fabric wraps beside
// client.ErrOnboardingCode (fmt.Errorf("%w: %w", ErrOnboardingCode, err)):
// the element of its Unwrap() []error that is not the sentinel.
func invalidReason(err error) string {
	var joined interface{ Unwrap() []error }
	if errors.As(err, &joined) {
		for _, e := range joined.Unwrap() {
			if e != nil && !errors.Is(e, client.ErrOnboardingCode) {
				return e.Error()
			}
		}
	}
	return "not an onboarding code"
}
