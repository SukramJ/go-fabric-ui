# Phase 2 fix wave, agent W (web half)

Branch `review/p2-web` in `…/scratchpad/p2wt/web`, base `29cf56b`. Five
commits, `ui/src/**` only, no `make web` (the coordinator rebuilds
`fabricui/dist`), no contract change, no new i18n key. Each fix: red vitest
first, then green.

| Commit | Finding |
| --- | --- |
| `95381ab` | UR-P2-ui-90: the five lists the reviewer named, plus the Events tab merge |
| `2c12a9d` | UR-P2-ui-90: the standing guard, and the lists it found |
| `7c1ebc1` | UR-P2-ui-05 |
| `60bf1f2` | UR-P2-conc-04, page half |
| `64c32f7` | UR-P2-leak-08, page parts |

## UR-P2-conc-04 (page half): fixed, `60bf1f2`

- `pages/node/Fabrics.svelte`: at the erase's 204 the page calls
  `node.resetCommissioning()`, so the erased node's window, codes and
  countdown leave the store. When the polled status answers, it calls
  `node.resync()` (status plus `syncCommissioning`, fabrics, exposure, the
  epochs, the node list) instead of `loadStatus` and `loadFabrics`.
- A stream the server ends: `api/stream.ts` already handles it. EventSource
  raises `error` in state CONNECTING, the browser reopens, and the open
  resyncs (phase 1b, `stream-life.test.ts` "resyncs when the browser's own
  reconnect opens the source again"). I added a test of the same through the
  store's default subscription with a stubbed `EventSource`. This needed no
  code change.
- Tests: `Fabrics.test.ts` › "drops the erased node's codes at once and
  resyncs when the node answers" (**red before**: phase `open`).
  `store-pairing.test.ts` › "a resync after the erase replaces the erased
  node's codes" (X1's red test, adapted to `resync`) and "a stream the
  server ends is reopened, and its open resyncs". Both were green before
  the change and stay as guards. X1's original store test was only red
  because the page called `loadStatus` and `loadFabrics`; the red test is
  now the page test.

## UR-P2-ui-05: fixed, `7c1ebc1` (no contract change)

- `store.svelte.ts`: `openWindow` no longer writes `kind: "own"` with a
  deadline taken from the browser clock. It sets `shown = { adopt: true }`
  and reads the status right after the POST. `otherWindow` adopts the kind
  and deadline (on the node's clock) from the first open status it sees and
  answers "same window". If that read fails, the next status (a
  `window_opened` or a resync) adopts instead. A later status of another
  kind or deadline still replaces the codes. `hydrateCommissioning` uses
  the same `windowOf`.
- Residual: if the status read right after the POST fails and another
  window replaces this one before the next status, that window is taken
  for the page's own. Only `window_kind` in the window answer (phase 3,
  decision 28) closes this.
- Tests (`store-pairing.test.ts`): "its own window_opened keeps the codes
  the POST answered and reads none" (X1's test), "a resync keeps the codes
  of its own window and reads none", and "a window opened while the status
  was not read is taken as its own at the next status" are **red before**.
  "a later window of another kind still replaces the codes" was green
  before and guards UR-P1-err-198.

## UR-P2-leak-08 (page parts): fixed, `64c32f7`

- `store.svelte.ts`: `fabric_added` sets `window: null, remaining: null` in
  phase `success`. Pair's success view renders neither.
- `pages/controller/Commission.svelte`: closing the terms block
  (`termsOpen` false) clears `retry`, which held the refused code. A new
  search keeps only the `passcodes` of devices it still lists.
- The erase drops the old node's codes (conc-04 above).
- Tests: `store-pairing.test.ts` › "a fabric added ends the flow without
  the window's codes". `Commission.test.ts` › "drops the code held for a
  terms retry when the terms block is closed" and "drops the passcode typed
  for a device a new search no longer lists". All **red before**.

## UR-P2-ui-90: fixed, `95381ab` and `2c12a9d`

- `95381ab`: `kit/EventList.svelte`, `pages/controller/node/Endpoints.svelte`
  (endpoints, device types, clusters), `kit/AttributeTable.svelte` and
  `kit/ClusterView.svelte` (accepted, generated, events) now key by index.
  The controller's Events tab merges by number and reads from the highest
  number shown. The merge moved from `node/Endpoint.svelte` to the new
  `kit/events.ts` (`mergeEvents`, `lastNumber`), used by both pages.
  Tests, X4's adapted (all **red before**, each_key_duplicate):
  `controller/node/Events.test.ts` › "lists two events with the same
  number" and "lists events delivered out of number order once across the
  next reload". `controller/node/Endpoints.test.ts` › "shows an endpoint
  whose DeviceTypeList names a device type twice" and "… server list names
  a cluster twice". `kit/ClusterView.test.ts` › "shows a {attributes,
  accepted, generated, events} list that names an id twice".
- The guard. Phase 1b left **no standing guard** of this class in code:
  the fixes for -130, -131 and -139 each added only a regression test for
  their own list. I built one: `src/lib/each-keys.test.ts` reads every
  keyed `{#each}` in `src/**/*.svelte`. Each must key by its index or by a
  key in an allow list that states why it cannot repeat (page constants,
  object or Map keys, the toaster's counter, merged log `seq`, the graph's
  index and layout ids, DataTable's `keyedRows`). A key from node,
  controller or host data is not allowed. A second test fails on stale
  allow-list entries, and a third checks the tag parser. Red before: 24
  offenders plus a stale entry.
- `2c12a9d`: the offenders the guard found now key by position: Inventory
  (endpoints, clusters, features), ClusterView features, Bindings, Overview
  ping addresses, Updates (queue and available), the node family's
  Endpoints, Surface, Endpoint (parts, clusters), Diagnostics (services,
  parameters, ecosystems), EcosystemHints findings, Selection (filters,
  preview, legend), NodeConfig, and Select options. DataTable's row key
  could also repeat (controller Fabrics `fabric_index`, Commission
  `instance_name`, sessions and so on). DataTable now renders rows under
  `keyedRows(rows, rowKey)`, which appends a count to a repeated key.
  Tests: `each-keys.test.ts` (above) and `DataTable.test.ts` › "renders
  both rows, flat and grouped" (**red before**). The two keys X4 left
  untraced (ping addresses, discovered instances) are covered.
- Helper: `uncaught()` (X4's collector) moved to `src/test/support.ts`.

## Guards

At HEAD `64c32f7`: `npx vitest run` passes 56 files and 580 tests.
`npx svelte-check --tsconfig ./tsconfig.json --fail-on-warnings` reports
992 files, 0 errors, 0 warnings. The working tree is clean apart from the
`ui/node_modules` symlink.

## Follow-up: UR-P2-test-95, the two browser guards (`19e7236`)

- `ui/e2e/origins.spec.ts` contains two of X4's guards, cut down, and
  nothing else from that spec. The spec starts its own server on
  `localhost:<ephemeral>` as the second origin; the e2e server runs on
  `127.0.0.1`. The existing project (Chromium) is kept and no project was
  added.
  - "a page of another origin cannot frame the shell; a same-origin parent
    can" (T10): two cross-origin frames are refused and the console reports
    `frame-ancestors`. A same-origin parent page frames the shell, and the
    framed shell shows the Fabrics page.
  - "a page of another origin reads no status and no stream" (T8/T9, read
    parts only): `fetch` of the status is blocked with and without
    credentials. A `no-cors` fetch comes back opaque and empty. `EventSource`
    on the node stream errors. A `<script src>` of the status fails to load.
    The write attempts and the token-header requests in X4's version were
    left out.
- `make e2e-web` against the committed bundle, without `make web`: 44
  passed, the 2 new tests included.

## Follow-up: UR-P2-ui-191 (`5d19993`, on `027375b`)

- `pages/node/Endpoint.svelte`: the cluster panels were keyed by
  position. They now key by `keyedRows(d.clusters, x => x.id_hex)`, and
  `each-keys.test.ts` has an allow-list entry for the key ("keyedRows
  suffixes a repeated cluster id").
- The new key alone did not fix it. The `open={…}` expression shared its
  template effect with `data-testid`, so every reload wrote `open` back
  to every panel. The open state is now set once, when the panel is
  created, through an attachment (`{@attach initiallyOpen}`, the `≤ 6
  clusters` rule read with `untrack`).
- Test: `Endpoint.test.ts` › "keeps a closed panel closed when a cluster is
  added in front", red before (the closed panel was open again). After the
  fix, `npx vitest run` passes 56 files and 581 tests, and svelte-check
  reports 0 errors and 0 warnings.
