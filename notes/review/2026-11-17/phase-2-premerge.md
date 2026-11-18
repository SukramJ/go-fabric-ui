# Phase 2 — pre-merge read (Fable)

Public copy. Read: `git diff 29cf56b..e5d289a -- .
':!fabricui/dist'` on `review/p2`; go-fabric `v0.22.2` (`gf:`), the hosts at
the brief's pins (`srv:`, `ha2f:`). Probes ran in the scratch worktree
`…/scratchpad/p2wt/x2` at `e5d289a`, never here. M = measured (read or
run), I = inferred.

Probe files left untracked in the scratch worktree:
`fabricui/gofabric/pm_test.go`, `fabricui/pm_internal_test.go`,
`ui/src/lib/pages/node/pm-endpoint.test.ts`. The reviewer's
`fabricui/x2_*_test.go` there redeclare names the fix wave committed
(`leaving`, `slowDeleteRegistry`), so the `fabricui` package only builds
with them overlaid out: `go test -overlay …/scratchpad/pm_overlay.json`.

## Findings

### UR-P2-adpt-190 — medium — a commissioning that did not happen is answered 200 (regression of the adpt-03 fix)

- Where: `fabricui/gofabric/gofabric.go:755-761` (`commissionedDespite`),
  called at `:352` (`Pair`) and `:591` (`CommissionDiscovered`).
- Claim (M): the test is "the call returned an error and the node has a
  peer address". That is also true for a node that was commissioned
  *before* the call. `gf:controller/client/addresses.go:187-193`
  (`findOrCreate`) answers the existing node for a device whose instance
  id a record holds, and `gf:controller/client/fabricops.go:171-173`
  refuses it with `ErrAlreadyCommissioned`; `ErrFabricOperationInProgress`
  (`:92-93`) and `ErrClosed` (`:89-90`) return the same way. The adapter
  now answers each of them as a commissioning that succeeded, logs
  `fabricui.pair.start_failed` for a start that never ran, and (`:594`)
  forgets the cached device. At `29cf56b` the same request answered
  `ErrNodeFailed` (502).
- Measured: `TestPM_CommissionDiscoveredOfACommissionedNode`
  (`pm_test.go`): `Discover`, `CommissionDiscovered`, then
  `CommissionDiscovered` again with the same device. Output:
  `WARN fabricui.pair.start_failed node=peer1 error="client: node is
  already commissioned: peer1"`, `n2=true err2=<nil> nodes=1`, FAIL
  ("answered success (nothing was commissioned)").
- `Pair` (I): `gf:controller/controller.go:827-829` returns `n, err` for
  every `Commission` error on the known-address path, and
  `gf:controller/pair_race.go:71-73` returns the race's winner with its
  error; the same check decides. Not run.
- Fix: decide by what this call did, not by the node's state. The
  adapter already wraps the progress hook (`progressFor`): note there
  that step Subscribe 21.1 started — go-fabric reports it only after the
  record took the peer address (`gf:controller/client/fabricops.go:214-230`)
  — and let `commissionedDespite` require that mark. The cited range
  `fabricops.go:229-235` and the comment's sentence stay true. Red test
  first: the probe above, plus the existing
  `TestPair_ACommissioningThatHappenedIsNoFailure` staying green.
- Contract: no.

### UR-P2-ui-191 — low — the endpoint page's cluster panels keep their open state by position

- Where: `ui/src/lib/pages/node/Endpoint.svelte:196` (`{#each d.clusters
  as c, ci (ci)}` around `<details open=…>` at `:198`).
- Claim (M): the `<details>` hold the user's open/closed state in the DOM
  and the page replaces `detail` in place on every `attribute_changed` of
  the endpoint and every topology change (`:103-111`), without unmounting
  the list. When the cluster list gains or loses an entry before a panel
  (a reassembly, a surface policy change), the state stays at the index:
  the panel the user closed opens again and its neighbour closes. With
  the former key (`c.id`) the state moved with its cluster.
- Measured: `pm-endpoint.test.ts` — fixture endpoint 2, close the last
  cluster's panel, answer the next detail read with a cluster `0x0003`
  added in front, apply `attribute_changed`: `expected true to be false`
  (the closed panel is open again).
- Fix: key by a stable identity, deduplicated, as `DataTable` does:
  `{#each keyedRows(d.clusters, (c) => c.id_hex) as r (r.key)}` with
  `r.row` in the body; `each-keys.test.ts` gets the allow-list entry for
  that key beside `lib/kit/DataTable.svelte :: r.key` (`:29`).
- Contract: no.

### UR-P2-test-192 — note — a test comment said more than its test

- Where: `fabricui/headers_test.go:75-77`.
- Claim (M, read): the comment of the no-CORS guard (UR-P2-test-95) went
  beyond what the guard asserts. Fix: end the sentence at the read; the
  test body stays.

### UR-P2-doc-193 — note — two reference points in `nodeswitch.go`

- `fabricui/nodeswitch.go:34` cites `registry.go:163-177` for "the node
  keeps its handler". This diff added two lines above it
  (`registry.go:60-61`); `managedHandler` with its comment is now
  `registry.go:168-190` (func at `:175`), and `163-167` is the tail of
  the function before it (M). `events.go:152-166` (`:28`) still holds.
- `nodeswitch.go:26-27` and `docs/embedding.md:183`: "Set ends every
  stream a handler serves over the switch". True when the switch itself
  is `Options.Node` or `NodeEntry.Source` (`stream.go:40-47` asserts
  `nodeAdapter.n.(*NodeSwitch)`), which is what both hosts do
  (`srv:internal/daemon/web.go:43`, `registry.go:203`;
  `ha2f:cmd/ha2fabric/main.go:113`, `nodes.go:442`). A switch inside a
  host's own `NodeSource` wrapper, and a host's `ObserveNode` over a
  switch, are not ended (M, read). Fix: "…over the switch handed to
  `Options.Node` or `NodeEntry.Source` itself", one clause in both places.

## Holds

1. Admin/token paths and streams
- `NodeSwitch.Set` (`nodeswitch.go:70-89`): target stored and channel
  replaced under one mutex, `swapped()` reads under the same; a stream
  reads the channel before it subscribes (`stream.go:31-34`), so a `Set`
  before the read gives the new node and a `Set` after it ends the
  stream. Concurrent `Set`s each close their own channel once. No lock is
  held across `close` or a callback. Holds.
- Write deadline (`stream.go:89-95`, `:124`, `:142`): set before the
  headers and every write; `ErrNotSupported` ignored. Go's server does
  not carry the deadline into the next request of a kept-alive
  connection: `TestPM_StreamLeavesNoWriteDeadlineOnTheConnection` ends
  the stream server-side, waits six heartbeats, and the next request on
  the same connection answers 200 (M). The per-write deadline replaces a
  host's `WriteTimeout` for the stream's life, which is what a stream
  needs. Holds.
- Fresh window (`handlers_pairing.go`): `openMu` then `h.mu`, never the
  reverse; the observer callback takes `h.mu` and calls only the node's
  clock; `h.mu` is never held across `window()`, `observe` or an open.
  `OpenedAt` is the stored value copied into each snapshot
  (`gf:bridge/commissioning_window.go:290-297`, written only at `:463`
  and `:631`), so `Equal` and the `windowID` comparison are stable for
  one window. `forgetOnClose`: the three orders (transition before the
  unsubscribe is stored, after it, window ended before registration)
  each unregister exactly once. The observer left on an erased node is
  released with that node (G's note). Holds. Residual, not a finding: an
  `eph.openedAt` of zero with a snapshot discriminator of 0 accepts any
  basic window; it needs the window to change between the open and the
  read two lines later.
- conc-01 (`events.go:283-312`): the own window restarted is
  `window_opened` again, since go-fabric restamps `OpenedAt`
  (`gf:bridge/commissioning_window.go:457-463`); an administrator's
  window replaces the node's own without the node being closed first
  from the observer's view (`:617-631`). The `docs/contract.md:221-224`
  sentence holds.
- Detached erase, registry `Delete`, `Decommission`: `WithoutCancel` plus
  two minutes, cancelled on return; the handler still waits, so nothing
  runs past the handler. A host's `Shutdown` waits up to that bound, as
  for the factory reset already. `deleteNode` keeps 409 for
  `ErrDefaultNode`. Holds.

2. Citations and sentences
- `gf:client/interaction/write.go:259-275` at v0.22.2: the unchunked
  branch and REPLACE-ALL plus one append per element; `ListChunked`
  (`:20-23`) excludes ACL and Extension; one status per response path
  (`:187-194`). Holds.
- `gf:controller/client/fabricops.go:229-235`: Subscribe started, the
  node's start, failed and returned, succeeded. Holds (the use made of
  it: UR-P2-adpt-190).
- `spa.go:26-33`: index for `/` and `/index.html`, assets `./assets/…`
  (`fabricui/dist/index.html:8`, `:105-108`). Holds; the loom lines were
  not re-read (no export of that pin in this session), the 404 is stated
  as this module's.
- `docs/embedding.md`: intro, status-route paragraph, shutdown pattern,
  detached paragraph, `:401`. Hold against the code.

3. Index keys (UR-P2-ui-90), every changed list read
- No local row state (text, badges, links, buttons on shared state):
  `EventList:29` (new events enter at the front, rows are stateless),
  `AttributeTable:31`, `ClusterView:35/47`, controller `Endpoints:31/36/53`,
  node `Endpoints:99/112/132`, `Overview:98`, `Updates:66/78`,
  `Diagnostics:199/236/249`, `EcosystemHints:53`, `NodeConfig:42`,
  `Selection:370/400`, `Surface:142/158`, `Endpoint:176`.
- State held outside the row, by identity: `Inventory:98/147` (maps by
  endpoint and cluster id), `Bindings:101` (drafts by endpoint,
  controlled inputs), `Selection:329` (`filterValues` by field),
  `Select:46` (`bind:value` on the select).
- `Surface:222`: the cards are `<details>` with a `DataTable` each, but
  every reload unmounts the list (`loading`), so no state survives a list
  change under either key.
- `DataTable` `keyedRows`: unique keys unchanged, a repeated key gets a
  suffix by position; per group. Holds.

4. web-91: the page requests `./assets/*`, `./favicon.svg`, `./api/v1`
and hash routes only; no `history` API, no path navigation. Both add-ons
have no `ingress_entry` (root). `srv`'s `/health` is on its own mux. HEAD
and query strings do not change the path match; `//` is cleaned by the
mux before the handler. No legitimate non-root, non-file request found.

5. Held rows: commit messages `29cf56b..e5d289a`, changelog, docs, code
comments and tests read; nothing points at a held row. `ui/e2e/origins.spec.ts` carries reads and framing
only.
