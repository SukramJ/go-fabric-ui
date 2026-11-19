# Phase 4, run 1 — unit W (web rest)

Not a held candidate: no finding of this unit is a security finding of
high or medium. Reviewer W (Opus), HEAD read `979a67a`, scratch worktree
`<scratchpad>/p4wt/W` (detached at `979a67a`, `npm ci`; red tests under
`ui/src/review-p4/w-pages.test.ts` and `ui/scripts/review-p4-W.test.mjs`).
Delta first (decision 66): every `ui/` file of the delta map's first table
outside unit S's paths, line by line; then the rest of the unit by slices.

Counts: high 0, medium 0, low 4, note 3. Contract changes: none.

## Findings

### UR-P4-ui-60 — low — the viewer's window countdown stands still

- **Where:** `ui/src/lib/pages/node/Pair.svelte:57`, `:113`.
- **Claim:** without admin rights the Pair page shows "window open, N s
  left" from `windowSecondsLeft(node.status, node.statusAt)`, a `$derived`
  with no reactive clock (the function's `now` defaults to `Date.now()`,
  which Svelte does not track). The number is computed once per status
  read and stays: 300 s still reads 300 s after ten seconds, and after the
  window has closed by its timer the page keeps a positive "left" until
  the next status read. `Status.svelte:35`, `:39` keeps a ticking `now` for the
  same value; the Pair page has none. Pre-existing with `secondsUntil`
  before 0.10.0; not in any register. **M.**
- **Test:** `ui/src/review-p4/w-pages.test.ts` › "UR-P4-ui-60 … counts down
  while the page is shown" — red: `expect(element).not.toHaveTextContent(/n.":300/)`,
  received `node.pair.window_open_left:{"left":"common.seconds_short:{\"n\":300}"}`
  after `advanceTimersByTimeAsync(10_000)`.
- **Fix:** a `now` state ticked by a 1 s interval while `viewerWindowOpen`
  (as `Status.svelte`), passed as the third argument; cleared on destroy.
- **Contract:** no.

### UR-P4-ui-61 — low — "not in effect" outlives a successful rebuild

- **Where:** `ui/src/lib/store.svelte.ts:97`, `:268`; shown at
  `ui/src/lib/pages/node/Selection.svelte:333`; `ui/src/lib/pages/node/Fabrics.svelte:64`.
- **Claim:** `exposureNotInEffect` is set only by `saveBulk` and never
  cleared otherwise. After an apply answered `reassembled: false`, a
  successful reassemble from the Fabrics page (the same node store) rebuilds
  the node with the stored selection, yet the Selection page keeps the
  warning "stored, but … not in effect yet" for the life of the store; the
  same after a successful apply from another tab (`exposure_changed` only
  reloads the items). The notice then contradicts the node. **M** (the
  reassemble path); the other-tab path **I**.
- **Test:** `ui/src/review-p4/w-pages.test.ts` › "UR-P4-ui-61 … goes when the
  node is reassembled from the Fabrics page" — red: `expected true to be false`
  at `ui.node.exposureNotInEffect` after the `reassemble_done` toast.
- **Fix:** a store method `reassembled()` (or the reassemble call moved into
  the store) that clears the flag on success; clear it as well when a
  resync or `surface_changed`/`endpoints_changed` follows a rebuild, or
  state in the notice that it describes the last apply from this page.
- **Contract:** no.

### UR-P4-test-62 — low — the contract guard misses required request parts

- **Where:** `ui/scripts/contract-guard.mjs:53-56` (parameters: schema only),
  `:27-49` (paths, methods, responses), `parameterIndex` `:136-151`.
- **Claim:** the guard's header (`:1-12`) says it fails on "a request field
  that became required"; for a request it compares only component-schema
  `required` lists and parameter schemas. Three breaks of an old client pass
  silently: a query parameter whose `required` goes false → true, a new
  operation-level parameter with `required: true`, and a `requestBody` whose
  `required` goes false → true. The contract has six query parameters, two
  of them required (`DELETE /controller/credentials` `kind`, `id`), so the
  case is live. **M.**
- **Test:** `ui/scripts/review-p4-W.test.mjs` › "UR-P4-test-62 …" — three
  red cases: `diffContracts(prev, cur)` returns `[]` for each.
- **Fix:** in `diffContracts`, per previous operation: every current
  parameter with `required: true` that was absent or not required in
  `prev` → `"<METHOD> <path>: parameter <name> became required"`; the same
  for `requestBody.required`; resolve `$ref` parameters through
  `components.parameters`. A test per case in `contract-guard.test.mjs`.
- **Contract:** no (guard only).

### UR-P4-test-63 — low — the i18n guard sees only the first literal of a `t()` call

- **Where:** `ui/src/lib/i18n/coverage.test.ts:103`.
- **Claim:** "define every literal key the code passes to t()" extracts keys
  with `/\bt\(\s*"([a-z0-9_.]+)"/`: the key in a ternary's second branch
  and every key handed to `t()` through a variable are never checked. Today
  that is nine keys: `node.endpoints.unreachable` (`Endpoint.svelte:175`),
  `controller.endpoints.unreachable` (`node/Inventory.svelte:110`,
  `node/Endpoints.svelte:40`), `controller.inventory.fixed_label`
  (`Endpoint.svelte:198`, `node/Inventory.svelte:129`), the four
  `controller.nodes.{removed,interviewed,enabled,disabled_done}` through
  `act(…, key)` (`Nodes.svelte:103-177`, `t(done…)` at `:85`), and
  `inventory.{accepted,generated,events}` through `ClusterView.svelte:17-19`,
  `:43`. All nine exist in both catalogues today (checked by grep); a typo
  in any of them would pass the guard and render the raw key. **M.**
- **Test:** `ui/scripts/review-p4-W.test.mjs` › "UR-P4-test-63 …" — red:
  the guard's regex over `t(ok ? "node.endpoints.reachable" : "node.endpoints.unreachable_typo")`
  does not yield the second key.
- **Fix:** collect every string literal of the catalogue's namespaces
  (`"(?:common|node|controller|operation|events|graph|inventory)\.[a-z0-9_.]+"`)
  in the sources, minus an explicit allow-list (`dates.ts:10`'s
  `"common.locale"` is one already present), and require each in both
  catalogues.
- **Contract:** no.

### UR-P4-ui-64 — note — the fabric label's hint says characters, the rule is bytes

- **Both sides:** `ui/src/lib/i18n/en.json:297` "Up to 32 characters" (de
  `:297` "Bis zu 32 Zeichen") and `Settings.svelte:121` `maxlength={32}`
  (UTF-16 code units) against `api/client.ts:195` `FABRIC_LABEL_MAX_BYTES = 32`
  and `Settings.svelte:38` (UTF-8 bytes, decision 52). A label of 17
  umlauts fits the hint and the field and is refused with
  `controller.settings.label_too_long` ("longer than 32 bytes").
- **Fix:** the hint in bytes ("Up to 32 bytes (fewer with accents or
  emoji)"), both catalogues.

### UR-P4-ui-65 — note — a viewer sees a warning's bare code

- **Both sides:** `HostWarnings.svelte:20-21` renders `w.code` as a badge and
  `w.message` as text; for a request without admin rights `message` is
  empty (contract `Warning.message`, `openapi.yaml:2268`; decision 34). A
  viewer's status page then shows a row of host-defined machine codes with
  no sentence. **I** (no host's code list read).
- **Fix:** without admin rights one line under the card, e.g.
  `node.warnings.viewer` "The host reports configuration findings; an
  administrator sees the details", or hide the empty span.

### UR-P4-ui-66 — note — the ACL lock-out question comes before the refusal

- **Both sides:** `controller/node/Acl.svelte:105-115`: `save()` asks the
  destructive lock-out confirmation first (`keepsUs`) and only then refuses
  a list with an `unknown` entry (`writableEntries`,
  `controller.acl.unknown_entry`). The operator confirms a lock-out for a
  write that is never sent.
- **Fix:** the `writableEntries` check first.

## Read without finding (delta, line by line)

`ui/src/lib/countdown.ts`, `context.ts`, `kit/events.ts`, `kit/EventList.svelte`,
`kit/DeniedState.svelte`, `kit/Spinner.svelte`, `kit/index.ts`,
`i18n/en.json`, `i18n/de.json` (the 13 new keys, parity),
`pages/controller/{Commission,Diagnostics,Node,Nodes,Settings}.svelte`,
`pages/controller/node/{Bindings,Events,Overview}.svelte`,
`pages/node/{Diagnostics,EcosystemHints,Endpoint,Endpoints,Fabrics,Selection,Status,Surface,Topology}.svelte`,
`pages/operation/Operation.svelte`, `shell/NodeFamily.svelte`, `shell/route.ts`,
`shell/toast.svelte.ts` (comment), `test/support.ts`, `ui/scripts/contract-guard.mjs`
(beyond -62), `ui/e2e/smoke.spec.ts` (`watch`, `settled`), the delta's tests
as guards: `contract-0.10.test.ts`, `i18n/coverage.test.ts` (beyond -63),
`each-keys.test.ts`, `catalogs.test.ts`, `contract-guard.test.mjs`,
`Commission.test.ts` (the `run()` helper), page tests of Diagnostics,
Endpoint, Fabrics, NodeManage, Pair, Selection, Status, controller Node.
The stores' and App's logic as read for state only
(`store.svelte.ts` status/window/apply paths, `controller-store.svelte.ts`
run id, resync, connect) — no state defect beyond -61.

Rest of the unit, by slices: `kit/{Tabs,HostWarnings}`, `pages/node/{Pair,HostWarnings}`,
`pages/controller/Commission` (terms, passcode, run id), `pages/controller/node/Acl`,
the template-literal and indirect `t()` sites. Not read again line by line
beyond those: the files of the baseline-only list, which phases 1b–3 read and
which did not move since.

## Observations, not findings

- `newCommissioningId()` (`controller-store.svelte.ts:343-347`) falls back to
  `Date.now()`+`Math.random()` where `crypto.randomUUID` is absent (a page over
  plain HTTP, not a secure context): within the pattern, and the id is a
  correlation id, not a secret.
- The pages a non-admin cannot read show `DeniedState`, but their `onMount`
  loaders still send the gated reads (answered 403, dropped). Only reachable
  when a host mounts such a page alone; the shell's `NodeFamily` redirects to
  status first.
- No Playwright smoke covers a non-admin page (`DeniedState`); the vitest
  page tests do.

Tokens: about 160k.
