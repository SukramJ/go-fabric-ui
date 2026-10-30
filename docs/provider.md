# Writing an ExposureProvider

The selection page knows nothing about a host's model. It renders what the
host's `fabricui.ExposureProvider` describes:

```go
type ExposureProvider interface {
    Schema(ctx context.Context) (Schema, error)
    List(ctx context.Context) ([]Item, error)
    Preview(ctx context.Context, updates []Update) ([]Item, error)
    Apply(ctx context.Context, updates []Update) (applied int, err error)
}
```

## Schema

`Schema` says how to show the list:

- `Columns` — the host's columns in display order. Each names a key of
  `Item.Fields` and carries its label in the host's language. `Filter:
  true` makes the page offer a value filter for the column, built from the
  distinct values the items carry.
- `Groups` — group the items by `Item.Group` (one section per group, in the
  order of the first item of each).
- `Filters` — fixed filters with the options the host names; an item
  matches an option when `Item.Fields[Field]` equals its value.

A field can hold several values — the labels an item carries. The host
puts them in `Item.Lists[key]`, one entry per value (`[]string{"Evening",
"Living, dining"}`), and `Multi: true` on the column says so: the page
shows each value as a chip, and a value filter on the column (`Filter:
true`) offers the individual values across all items, not the
combinations, and matches an item whose list holds the chosen value.
`Multi: true` on a fixed `Filter` makes it match the same way. An entry
has no separator semantics, so a value may contain a comma; an empty
entry is dropped and an empty value never matches. `Item.Fields[key]` may
still carry a display string for the same key; for a `Multi` column or
filter `Lists` wins whenever it has the key. `Item.ListValues(key)`
returns the list as the page reads it.

Deprecated, accepted until v0.2.0: a host that leaves `Lists` without the
key carries the values in `Item.Fields[key]` as one comma-separated
string (`"Kitchen, Evening"`); the page splits it on commas and trims
each value of spaces, so a value with a comma splits into two. Without
`Multi` a comma is an ordinary character and the filter compares the
whole string.

The page always shows, beside the host's columns, a select box, the state
icon (exposed, partially mappable, available, cannot be mapped), the name
and the verdict's device type. The search matches the name, the group, the
device type and every field.

## Items

| Field | Meaning |
| --- | --- |
| `Key` | Opaque to the page; it comes back unchanged in an `Update`. |
| `Name` | The name the item is exposed under. The page can rename it. |
| `Group` | The section under `Schema.Groups`. |
| `Fields` | The values of the schema's columns. |
| `Lists` | The values of the list-valued (`Multi`) columns and filters, one entry per value; wins over `Fields` for the same key. Optional. |
| `Verdict` | go-fabric's `eligibility.Verdict`: state, device type, clusters, reason. An unmappable item cannot be selected. |
| `Enabled` | Whether the item is exposed. |
| `Source` | Where the verdict comes from (a rule, an override, a label, a suggestion), in the host's words; the detail drawer shows it. |

## Preview and apply

The page collects changes (`Update{Key, Enabled, Name}`; a nil `Name`
leaves the name) and shows them before anything is written:

- `Preview` is a dry run. It returns the touched items as they would be
  after the updates — their verdict and whether they would be enabled —
  and changes nothing. The page marks an item whose verdict moves. A host
  whose verdicts depend on each other (one item suppressing another)
  returns every item the change affects, not only the named ones.
- `Apply` writes the updates and returns how many it applied. Whether a
  batch is atomic is the provider's decision; the in-memory example checks
  every update before it writes any.

Wrap `fabricui.ErrUnknownItem` for a key the provider does not know and
`fabricui.ErrInvalidUpdate` for an update it refuses (enabling an
unmappable item); both answer 400 with the error's text. Any other error
answers 500 and stays in the log. After `Apply` the handler calls
`Options.Reassemble` so the node serves the new selection before the
response, and publishes `exposure_changed` with the keys, then
`endpoints_changed`; an endpoint whose surface the reassembly changed
reaches the stream as `surface_changed` from the node's observer. A host that
changes its selection elsewhere publishes the same event through its
`fabricui.Events`.

## An in-memory provider

The handler tests run on this provider
([`fabricui/memory_provider_test.go`](https://github.com/SukramJ/go-fabric-ui/blob/main/fabricui/memory_provider_test.go));
a host's provider has the same shape over its own store:

```go
type memoryProvider struct {
    mu     sync.Mutex
    schema fabricui.Schema
    items  []fabricui.Item
}

func (p *memoryProvider) Schema(context.Context) (fabricui.Schema, error) { return p.schema, nil }

func (p *memoryProvider) List(context.Context) ([]fabricui.Item, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    return cloneItems(p.items), nil
}

func (p *memoryProvider) Preview(_ context.Context, updates []fabricui.Update) ([]fabricui.Item, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    out := make([]fabricui.Item, 0, len(updates))
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

func (p *memoryProvider) Apply(_ context.Context, updates []fabricui.Update) (int, error) {
    p.mu.Lock()
    defer p.mu.Unlock()
    idx := make([]int, 0, len(updates))
    for _, u := range updates { // check all before writing any
        i, err := p.find(u)
        if err != nil {
            return 0, err
        }
        idx = append(idx, i)
    }
    for n, i := range idx {
        applyUpdate(&p.items[i], updates[n])
    }
    return len(updates), nil
}

func (p *memoryProvider) find(u fabricui.Update) (int, error) {
    i := slices.IndexFunc(p.items, func(it fabricui.Item) bool { return it.Key == u.Key })
    if i < 0 {
        return 0, fmt.Errorf("%w: %q", fabricui.ErrUnknownItem, u.Key)
    }
    if u.Enabled && p.items[i].Verdict.State == contract.EligibilityUnmappable {
        return 0, fmt.Errorf("%w: %q cannot be exposed", fabricui.ErrInvalidUpdate, u.Key)
    }
    return i, nil
}

func applyUpdate(it *fabricui.Item, u fabricui.Update) {
    it.Enabled = u.Enabled
    if u.Name != nil {
        it.Name = *u.Name
    }
}
```

A real provider builds `Verdict` with go-fabric's `eligibility` package
(`eligibility.Classify`, or the verdict a host's own rule derives) and keeps
its selection where the host keeps its state.
