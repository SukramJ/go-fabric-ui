// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"errors"
	"strings"

	"github.com/SukramJ/go-fabric/eligibility"
)

// ExposureProvider is the host's side of the selection page: which items it
// could expose, how the page shows them, what a change would do and the
// change itself. The page knows nothing about the host's model; it renders
// [Schema] and [Item] as they come.
//
// Every method is called on a request goroutine and may be called
// concurrently. Errors that wrap [ErrUnknownItem] or [ErrInvalidUpdate]
// answer 400 with the error text; any other error answers 500 and stays in
// the log.
type ExposureProvider interface {
	// Schema describes the list: its columns, its filters and whether the
	// items are grouped.
	Schema(ctx context.Context) (Schema, error)
	// List returns every item the host could expose, in the order the page
	// shows them.
	List(ctx context.Context) ([]Item, error)
	// Preview is a dry run: the items the updates touch, as they would be
	// after Apply — the verdict each would carry and whether it would be
	// enabled. It changes nothing.
	Preview(ctx context.Context, updates []Update) ([]Item, error)
	// Apply writes the updates and returns how many it applied. Whether a
	// batch is applied atomically is the provider's decision; the handler
	// reports the count it returns.
	Apply(ctx context.Context, updates []Update) (applied int, err error)
}

// Errors a provider wraps to answer 400 instead of 500.
var (
	// ErrUnknownItem marks an update whose key the provider does not know.
	ErrUnknownItem = errors.New("fabricui: unknown item")
	// ErrInvalidUpdate marks an update the provider refuses as malformed.
	ErrInvalidUpdate = errors.New("fabricui: invalid update")
)

// Schema describes the selection list.
type Schema struct {
	// Columns are the host's columns, in display order. Each one names a
	// key of [Item].Fields.
	Columns []Column `json:"columns"`
	// Groups reports whether the page groups the items by [Item].Group.
	Groups bool `json:"groups"`
	// Filters are fixed filters the page offers above the list.
	Filters []Filter `json:"filters"`
}

// Column is one host column of the selection list.
type Column struct {
	// Key names the [Item].Fields entry the column shows.
	Key string `json:"key"`
	// Label is the column header, already in the host's language.
	Label string `json:"label"`
	// Filter offers a value filter for the column, built by the page from
	// the distinct values the listed items carry.
	Filter bool `json:"filter"`
	// Multi marks a list-valued field: [Item].Lists carries the values,
	// one entry per value. The page shows each value as a chip, a value
	// filter on the column offers the individual values across all items
	// rather than their combinations, and it matches an item that holds
	// the chosen value in its list. See [Item.ListValues] for the
	// deprecated comma-separated form in [Item].Fields.
	Multi bool `json:"multi,omitempty"`
}

// Filter is a fixed filter with the options the host names.
type Filter struct {
	// Field is the [Item].Fields key the filter matches.
	Field string `json:"field"`
	// Label is the filter's name.
	Label string `json:"label"`
	// Options are the values the filter offers. An item matches an option
	// when its field equals the option's value, or, with Multi, when the
	// option's value is one of the item's [Item.ListValues] for the field.
	Options []FilterOption `json:"options"`
	// Multi matches by membership in the item's list for the field
	// ([Item.ListValues]); an empty value never matches.
	Multi bool `json:"multi,omitempty"`
}

// FilterOption is one value of a [Filter].
type FilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Verdict is go-fabric's eligibility verdict: whether an item can be
// exposed, as which device type with which clusters, and why not.
type Verdict = eligibility.Verdict

// Item is one exposable host item.
type Item struct {
	// Key identifies the item to the provider. It is opaque to the page
	// and travels back unchanged in an [Update].
	Key string
	// Name is the name the item is exposed under.
	Name string
	// Group is the group the page lists the item under when
	// [Schema].Groups is set.
	Group string
	// Fields are the values of the schema's columns, by column key. For
	// a list-valued (Multi) column it may carry the display string; the
	// values themselves travel in Lists.
	Fields map[string]string
	// Lists are the values of the list-valued (Multi) columns and filters,
	// by field key: one entry per value, with no separator semantics, so
	// a value may contain a comma ("Living, dining"). When a key is in
	// both Lists and Fields, Lists wins for a Multi column or filter.
	Lists map[string][]string
	// Verdict is the item's eligibility.
	Verdict Verdict
	// Enabled reports whether the item is exposed.
	Enabled bool
	// Source says where the verdict comes from — a rule, an override, a
	// label, a suggestion — in the host's words, so the page can show why
	// an item is classified as it is.
	Source string
}

// Update is one change of an item.
type Update struct {
	// Key is the [Item].Key the change applies to.
	Key string `json:"key"`
	// Enabled is whether the item is to be exposed.
	Enabled bool `json:"enabled"`
	// Name, when set, renames the item; nil leaves its name alone.
	Name *string `json:"name,omitempty"`
}

// ListValues is the list a Multi column or filter reads for key, as the
// page reads it: Lists[key] when the key is present there, its empty
// entries dropped; otherwise the comma split of Fields[key], each value
// trimmed of spaces, empty ones dropped.
//
// Deprecated form: the comma-separated Fields[key] for a Multi column is
// accepted until v0.2.0; a host fills Lists instead.
func (it *Item) ListValues(key string) []string {
	if list, ok := it.Lists[key]; ok {
		out := make([]string, 0, len(list))
		for _, v := range list {
			if v != "" {
				out = append(out, v)
			}
		}
		return out
	}
	out := []string{}
	for v := range strings.SplitSeq(it.Fields[key], ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
