// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/SukramJ/go-fabric/contract"
)

// memoryProvider is the in-memory [ExposureProvider] docs/provider.md
// walks through: a fixed list of items, a schema with two columns, a
// preview that applies the updates to copies and an apply that writes
// them. An item whose verdict is unmappable cannot be enabled.
type memoryProvider struct {
	mu     sync.Mutex
	schema Schema
	items  []Item
	// applyErr and listErr make the provider fail.
	applyErr error
	listErr  error
	applied  [][]Update
}

func newMemoryProvider() *memoryProvider {
	return &memoryProvider{
		schema: Schema{
			Columns: []Column{
				{Key: "room", Label: "Room", Filter: true},
				{Key: "kind", Label: "Kind", Filter: false},
			},
			Groups: true,
			Filters: []Filter{{Field: "kind", Label: "Kind", Options: []FilterOption{
				{Value: "light", Label: "Light"}, {Value: "sensor", Label: "Sensor"},
			}}},
		},
		items: []Item{
			{
				Key: "lamp-1", Name: "Kitchen lamp", Group: "Kitchen",
				Fields:  map[string]string{"room": "Kitchen", "kind": "light"},
				Verdict: Verdict{State: contract.EligibilityMappable, DeviceType: 0x0100, Clusters: []uint32{0x0006, 0x0008}},
				Enabled: false, Source: "rule",
			},
			{
				Key: "sensor-2", Name: "Hall sensor", Group: "Hall",
				Fields: map[string]string{"room": "Hall", "kind": "sensor"},
				Verdict: Verdict{
					State: contract.EligibilityPartial, DeviceType: 0x0302, Clusters: []uint32{0x0402},
					Reason: "the battery level is not mapped",
				},
				Enabled: true, Source: "override",
			},
			{
				Key: "remote-3", Name: "Hall remote", Group: "Hall",
				Fields:  map[string]string{"room": "Hall", "kind": "remote"},
				Verdict: Verdict{State: contract.EligibilityUnmappable, Reason: "no cluster covers a remote"},
				Enabled: false, Source: "rule",
			},
		},
	}
}

func (p *memoryProvider) Schema(context.Context) (Schema, error) { return p.schema, nil }

func (p *memoryProvider) List(context.Context) ([]Item, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.listErr != nil {
		return nil, p.listErr
	}
	return cloneItems(p.items), nil
}

// Preview applies the updates to copies of the items they name and
// returns those copies; nothing is stored.
func (p *memoryProvider) Preview(_ context.Context, updates []Update) ([]Item, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Item, 0, len(updates))
	for _, u := range updates {
		i, err := p.find(u)
		if err != nil {
			return nil, err
		}
		it := cloneItems(p.items[i : i+1])[0]
		applyUpdate(&it, u)
		out = append(out, it)
	}
	return out, nil
}

// Apply checks every update before it writes any, so a batch applies
// whole or not at all.
func (p *memoryProvider) Apply(_ context.Context, updates []Update) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.applyErr != nil {
		return 0, p.applyErr
	}
	idx := make([]int, 0, len(updates))
	for _, u := range updates {
		i, err := p.find(u)
		if err != nil {
			return 0, err
		}
		idx = append(idx, i)
	}
	for n, i := range idx {
		applyUpdate(&p.items[i], updates[n])
	}
	p.applied = append(p.applied, slices.Clone(updates))
	return len(updates), nil
}

func (p *memoryProvider) find(u Update) (int, error) {
	i := slices.IndexFunc(p.items, func(it Item) bool { return it.Key == u.Key })
	if i < 0 {
		return 0, fmt.Errorf("%w: %q", ErrUnknownItem, u.Key)
	}
	if u.Enabled && p.items[i].Verdict.State == contract.EligibilityUnmappable {
		return 0, fmt.Errorf("%w: %q cannot be exposed: %s", ErrInvalidUpdate, u.Key, p.items[i].Verdict.Reason)
	}
	return i, nil
}

func applyUpdate(it *Item, u Update) {
	it.Enabled = u.Enabled
	if u.Name != nil {
		it.Name = *u.Name
	}
}

func cloneItems(in []Item) []Item {
	out := slices.Clone(in)
	for i := range out {
		out[i].Fields = maps.Clone(out[i].Fields)
		out[i].Verdict.Clusters = slices.Clone(out[i].Verdict.Clusters)
	}
	return out
}
