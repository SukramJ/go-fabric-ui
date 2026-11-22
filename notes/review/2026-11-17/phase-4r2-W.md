# Phase 4, run 2 — unit W (web rest)

Not a held candidate: no finding of this unit is a security finding of
high or medium. Reviewer W (Opus), HEAD read `792d15b` (review worktree at
`89c098b`, notes only on top). Scratch worktree `<scratchpad>/p4r2wt/W`
(detached at `792d15b`, `npm ci`); red tests in
`ui/src/review-p4r2/w-pages.test.ts` and `ui/scripts/review-p4r2-W.test.mjs`,
run with `npx vitest run <file>` (focused files only). The coverage pass
the coordinator asked for (decision 74, full cut) adds IDs 299–306 with
their red tests in `ui/src/review-p4r2/w-cover.test.ts`.

Counts: high 0, medium 0, low 11, note 6 (first pass low 6, note 3; coverage pass low 5, note 3). Contract changes: one proposed
fix (ui-291) would add a problem code (additive); everything else no.

## Scope

Listed first (`git ls-files ui/src ui/e2e ui/scripts`, minus S-web's files
and `schema.gen.ts`): pages (`pages/controller/**`, `pages/node/**` except
`Pair.svelte`, `pages/operation/**`, `pages/index.ts`), `kit/**`,
`i18n/{index.ts,en.json,de.json}`, `context.ts`, `countdown.ts`, `qr.ts`,
`scan/index.ts`, `theme/{tokens.css,index.ts}`, `index.ts`, `api/index.ts`,
shell files other than `App.svelte`/`route.ts` (`ConfirmDialog`,
`confirm.svelte.ts`, `dialogs`, `focus-trap.ts`, `NodeCreateDialog`,
`NodeFamily`, `NodeScope`, `Toaster`, `toast.svelte.ts`, `index.html`,
`main.ts`, `app.css`), `test/**`, `e2e/{smoke,origins}.spec.ts`,
`scripts/{contract-guard.mjs,contract-guard.test.mjs,check-notices.mjs,gen-types.mjs}`,
and the web tests that hold a guard.

Delta first (decision 66): every file of the delta map in this unit, line by
line — `pages/controller/node/{Updates,Acl}.svelte`,
`pages/node/{HostWarnings,Fabrics}.svelte`, `pages/controller/Settings.svelte`,
`pages/index.ts`, `i18n/{en,de}.json` (the five changed keys),
`scripts/contract-guard.mjs` (whole file), `ui/README.md`, `ui/package.json`,
and the delta's tests as guards: `contract-guard.test.mjs`,
`i18n/coverage.test.ts` (whole file), `pages/controller/Node.test.ts`
(new blocks), `pages/node/{Fabrics,Status}.test.ts` (new blocks),
`pages/controller/Settings.test.ts`, `pages/index.test.ts`. The S-web-owned
delta tests (`token-path.test.ts`, `Pair.test.ts`, `store-*.test.ts`) were
not read as targets.

## Findings

### UR-P4-ui-290 — low — the update check's error text is not text-safe under the HA skin

- **Where:** `ui/src/lib/pages/controller/node/Updates.svelte:89` (new in
  the delta, decision 70's per-node error).
- **Claim:** the line "The check failed for this node: …" is small status
  text (`text-sm`, `role="alert"`) coloured `text-[var(--ha-error-color)]`.
  Decision 31 (UR-P1-ui-172) keeps the HA colours for areas and icons and
  puts status text on `--fui-*-text`; every other error sentence of the
  pages does (`Settings.svelte:125`, `Commission.svelte:272`, `:287`,
  `Attributes.svelte:126`, …). Under the HA skin `--ha-error-color` falls
  back to `#db4437` (`theme/tokens.css:127`): 4.29:1 on the light card
  `#ffffff`, 3.97:1 on the dark card `#1c1c1c` — below 4.5:1. A new path of
  UR-P1-ui-172. No guard catches a page using an HA status colour for text
  (`theme/tokens.test.ts` checks the tokens, not their use). **M.**
- **Test:** `ui/src/review-p4r2/w-pages.test.ts` › "UR-P4-ui-290 … uses a
  text-safe colour token, which reaches 4.5:1 under the HA skin" — red:
  `expected { token, ok } to deeply equal { token: '--fui-error-text', ok: true }`,
  received `ok: false` (token `--ha-error-color`, 4.29:1).
- **Fix:** `text-[var(--fui-error-text)]` on that line; optionally a source
  guard that no `text-[var(--ha-(error|warning|success|primary|info)-color)]`
  sits on an element holding text (icons excepted by an allow-list).
- **Contract:** no.

### UR-P4-ui-291 — low — every 409 of the ICD registration asks the multi-admin question

- **Where:** `ui/src/lib/pages/controller/node/Icd.svelte:40-44`;
  server side `fabricui/gofabric/gofabric.go:783-790` (`nodeError`),
  `fabricui/handlers_controller.go:234-235`.
- **Claim:** the page treats any 409 on `POST …/icd/register` as the
  multi-admin refusal: it asks "other administrators may not support LIT"
  (`controller.icd.multi_admin`) and on yes sends the registration again
  with `allow_multi_admin: true`. Phase 1b/2 recorded that the 409 there
  was only `ICDMultiAdminError` (`phase-1b-G-web-B.md:202`,
  `phase-2-X2.md:187`); since UR-P3-adpt-91 (`ec57fba`) the adapter also
  maps `client.ErrICDOperation` (a call the client's state refuses, e.g.
  no subscription), `ErrDisabled` and `ErrFabricOperationInProgress` to
  `ErrConflict` → 409 `conflict`, the same code. A disabled node now gets
  the LIT question and a second registration that fails the same way. The
  page was not adjusted when the mapping widened. **M** (page, against a
  409 `conflict` answer); the server reaching 409 for those errors **M**
  by the code read.
- **Test:** `w-pages.test.ts` › "UR-P4-ui-291 … does not ask the multi-admin
  question for a conflict of the client's own state" — red:
  `expected "vi.fn()" to not be called with arguments: [ ObjectContaining{ title: "controller.icd.multi_admin" } ]`.
- **Fix:** give the multi-admin refusal its own problem code (e.g.
  `icd_multi_admin`, 409, additive) in `controllerError` and ask only on it;
  any other 409 is a plain error toast. A page-only stopgap (matching the
  detail text) is brittle.
- **Contract:** yes, additive (a problem code), if fixed as proposed.

### UR-P4-ui-292 — low — a node row navigates by `location.hash`, ignoring the host's routes

- **Where:** `ui/src/lib/pages/controller/Nodes.svelte:76-78` (`open`),
  against `context.ts:24` ("a host maps them onto its router") and the
  page's own `follow` prop (`Nodes.svelte:25-26`), which `Graph` uses
  (`kit/Graph.svelte:29`, default `location.assign`).
- **Claim:** a row click sets `location.hash = routes.controllerNode + id`.
  With the default `#/…` routes that works; a host that maps
  `controllerNode` onto its router (`"/app/controller/node/"`) gets
  `<page>#/app/controller/node/<id>` instead, which its router does not
  know; the `follow` a host passes is not used. Every other link of the
  page is an `href` (`:129`, `:171`) and follows the host's route. **M.**
- **Test:** `w-pages.test.ts` › "UR-P4-ui-292 … follows the host's
  controllerNode route, not a hash built from it" — red:
  `expected '#/app/controller/node/0000000000000011' not to be '#/app/controller/node/0000000000000011'`.
- **Fix:** `(follow ?? ((h) => location.assign(h)))(routes.controllerNode + n.node_id_hex)`,
  as `Graph` does.
- **Contract:** no.

### UR-P4-ui-293 — low — the controller node page does not follow a new `tab` for the same node

- **Where:** `ui/src/lib/pages/controller/Node.svelte:35-36` (`active`
  seeded once, `svelte-ignore state_referenced_locally`); the shell keys
  the page by node id only (`shell/App.svelte:340-341`).
- **Claim:** `#/controller/node/<id>?tab=acl` reached from
  `…?tab=overview` of the same node (back/forward, a pasted link, a host
  router that passes `tab`) leaves the page on the old tab. `Nodes.svelte:33-39`
  resets its tab when the prop changes; `Node.svelte` does not. **M.**
- **Test:** `w-pages.test.ts` › "UR-P4-ui-293 … shows the tab a new route
  names for the same node" — red: `Unable to find an element by:
  [data-testid="acl-entries"]` after `rerender({ tab: "acl" })`; the
  control "the acl tab shows its entries when the page opens on it" passes.
- **Fix:** the `Nodes.svelte` pattern: `chosen` state reset by an effect on
  `tab`, `active = chosen ?? (isTab(tab) ? tab : "overview")`.
- **Contract:** no.

### UR-P4-test-295 — low — the contract guard misses a new required request field

- **Where:** `ui/scripts/contract-guard.mjs:80-86`; its header `:4`.
- **Claim:** the header says the guard fails on "a request field, parameter
  or body that became required (or is new and required)". For fields the
  loop flags a requirement only when `(schema.properties ?? {})[req]`
  existed in `prev`: a field added to a request schema together with
  `required: [it]` passes. An old client never sends it, so every request
  of it is refused (strict bodies, 400). UR-P4-test-62's fix covered
  parameters and bodies, not this. **M.**
- **Test:** `ui/scripts/review-p4r2-W.test.mjs` › "a new required field of
  a request schema" — red: `expected [] to include 'schema Req: field b became required'`.
- **Fix:** for a schema in `requestNamed`, flag every `now.required` entry
  not in `wasRequired`, whether or not the field existed; keep today's
  rule for response-only schemas. A case in `contract-guard.test.mjs`.
- **Contract:** no (guard only).

### UR-P4-test-296 — low — the contract guard does not look inside an inline object

- **Where:** `ui/scripts/contract-guard.mjs:61-87` (removal and
  required checks only at the top level of a component schema),
  `compareNode` `:117-135` (recurses into inline properties for type and
  range only).
- **Claim:** a field removed from, or no longer required in, an inline
  object nested in a component schema passes. Live: `Status.surface` is
  such an object with eight required members (`contract/openapi.yaml`,
  schema `Status`, property `surface`); deleting `Status.surface.level`
  from the contract yields no break. A client generated from 0.11.0 reads
  `status.surface.level` as always present. **M.**
- **Test:** `review-p4r2-W.test.mjs` › "a field removed from an inline
  object inside a response schema" — red: `expected 0 to be greater than 0`;
  › "on the real contract: Status.surface.level removed passes" — red:
  `expected [] to not deeply equal []`.
- **Fix:** in `compareNode`, for an inline object (no `$ref`) report a
  removed property (`${at}: field x removed`), and on the response side a
  member that left `required`, on the request side one that joined it
  (the 295 rule); tests for both.
- **Contract:** no (guard only).

### UR-P4-ui-294 — note — the controller's fabric removal has no busy state (sibling of UR-P1-ui-145)

- **Both sides:** `pages/controller/node/Fabrics.svelte:33-48`, `:77` (no
  `busy`, the button never disabled) against the node page's fix of
  UR-P1-ui-145 (`pages/node/Fabrics.svelte`, busy on every action). With a
  host `confirm` that does not serialise, a double click sends two
  `DELETE …/fabrics/{index}`; the second answers 404 since 0.11.0 and
  toasts an error. Test (optional for a note):
  `w-pages.test.ts` › "UR-P4-ui-294 …" — red: `expected [ 2, 2 ] to have a length of 1`.
- **Fix:** a `busy` (or per-index set) as `Nodes.svelte`'s `act`.

### UR-P4-ui-297 — note — the node label's 64 is counted in UTF-16 units, the server counts runes

- **Both sides:** `pages/node/NodeManage.svelte:37` and
  `shell/NodeCreateDialog.svelte:19` (`label.trim().length <= 64`, UTF-16
  code units) against `fabricui/registry.go:107`
  (`utf8.RuneCountInString(l) > MaxNodeLabel`) and the contract's
  `maxLength: 64` ("1 to 64 characters", `openapi.yaml:2633`; JSON Schema
  counts code points). A label of 33–64 characters outside the BMP (emoji)
  is valid on the server and refused by the page (the button stays
  disabled). The opposite direction of UR-P4-ui-64.
- **Fix:** count `[...label.trim()].length`.

### UR-P4-ui-298 — note — the erase poll can outlive the Fabrics page

- **Both sides:** `pages/node/Fabrics.svelte:98-100` (`onDestroy` clears
  `poll`) against `:106-113`: the timer callback sets `poll = null` before
  it awaits `client.status()`; a page destroyed during that await is not
  cleared, and the `catch` schedules `waitForNode()` again, so the page's
  poll keeps running every 2 s while the node is down and finally calls
  `node.resync()` for a page that is gone. **I** (read; bounded by the node
  coming back).
- **Fix:** a `destroyed` flag checked before rescheduling and before
  `resync`.

## Findings of the coverage pass

### UR-P4-ui-299 — low — the label on a filled button fails AA in the shell's skin

- **Where:** `ui/src/lib/kit/Button.svelte:22` (`default`: `--ha-on-primary-color`
  on `--ha-primary-color`), `:25` (`destructive`: `text-white` on
  `--ha-error-color`); tokens `theme/tokens.css:67` (`--ha-on-primary-color:
  #ffffff`, not redefined under `html.dark`), `:86` (dark primary teal-400),
  `:96`, `:118`, `:127`, `:136`; the shell always sets `data-skin="ha"`
  (`shell/index.html`, pre-paint script).
- **Claim:** every primary button (Save, Commission, Apply, …) and the
  confirm dialog's destructive button (`shell/ConfirmDialog.svelte:43`)
  carry `text-sm` labels below 4.5:1: under the HA skin with Home
  Assistant's fallbacks white on `#009ac7` is 3.26:1, white on `#db4437`
  4.29:1; a host on the standalone palette in dark mode gets white on
  teal-400, 1.87:1, and white on `#f87171`, 2.77:1. The phase 1b fix of
  UR-P1-ui-172 left this pair out on purpose ("white text on filled
  primary/error buttons … out of this decision's scope",
  `phase-1b-fix-W2.md`), and no register row carries it. **M** (WCAG
  formula over the token values).
- **Test:** `w-cover.test.ts` › "UR-P4-ui-299 … reaches 4.5:1 on the
  primary and the destructive fill in every palette" — red:
  `expected [ [ 'HA skin primary', 3.26 ], …(3) ] to deeply equal []`
  (the four pairs above).
- **Fix:** an owner question beside decision 31: a darker fill for the
  label (e.g. `--fui-primary-fill` mixed towards black as the text tokens
  are, the HA colour kept for icons and borders), and in standalone dark a
  dark label on the light teal (`--ha-on-primary-color: #0f172a` under
  `html.dark`); a token test over fill and label in the four palettes.
- **Contract:** no.

### UR-P4-ui-300 — low — an empty state's description, the denied notice's hint, is in the disabled colour

- **Where:** `ui/src/lib/kit/EmptyState.svelte:18`
  (`text-[var(--ha-disabled-text-color)]`); used with text by
  `kit/DeniedState.svelte` (`common.denied.*_hint`, the reason a page is
  refused) and every `emptyDescription`.
- **Claim:** the description is meaningful text in the disabled colour:
  `#bdbdbd` on the HA skin's white card 1.88:1, `#9ca3af` on the
  standalone card 2.54:1 (`theme/tokens.css:51`, `:126`). The class of
  UR-P1-ui-172 on a token decision 31 did not cover. **M.**
- **Test:** `w-cover.test.ts` › "UR-P4-ui-300 … is readable text, not the
  disabled colour" — red: `expected { ha: false, standalone: false } to
  deeply equal { ha: true, standalone: true }`.
- **Fix:** `--ha-secondary-text-color` for the description (the icon may
  keep the disabled colour).
- **Contract:** no.

### UR-P4-ui-301 — low — the controller's diagnostics name a peer only when another page loaded the nodes

- **Where:** `ui/src/lib/pages/controller/Diagnostics.svelte:92-96`
  (`peerName` reads `controller.nodes`), `:73` (`onMount` loads only the
  three diagnostics routes); `controller-store.svelte.ts:153`, `:255`
  (nodes load on a stream reload or a resync).
- **Claim:** opened first (a bookmark, a host mounting the page alone),
  the peers' links read `0x0000000000000011` instead of the node's name,
  which the page's own comment promises, until a stream event reloads the
  node list. **M.**
- **Test:** `w-cover.test.ts` › "UR-P4-ui-301 …" — red: expected text
  `Kitchen`, received `0x0000000000000011`.
- **Fix:** `if (controller.nodes.length === 0) void controller.loadNodes()`
  in `onMount` (an admin page; the nodes route is the admin's too).
- **Contract:** no.

### UR-P4-ui-302 — low — an event timestamp outside the Date range takes the events list down

- **Where:** `ui/src/lib/kit/EventList.svelte:39`
  (`new Date(e.epoch_ms).toISOString()`); the value is the remote node's
  `EpochTimestamp` cast to int64 (`fabricui/gofabric/node.go:431-433`),
  contract `epoch_ms` `format: int64` (`openapi.yaml:3294`).
- **Claim:** a node that reports an epoch timestamp beyond ±8.64e15 ms (a
  wrong unit, a corrupt or hostile value) makes `toISOString` throw
  `RangeError: Invalid time value` while the list renders: the controller
  node's Events tab shows nothing more, and a reload brings the same
  event back. Hostile node data rendered (decision 2's attacker role);
  `formatDate` beside it does not throw. **M** (page); the reach from a
  real node **I** (go-fabric passes the value through, read at the adapter
  only).
- **Test:** `w-cover.test.ts` › "UR-P4-ui-302 … does not take the
  controller's events tab down" — red: `expected [ 'Invalid time value' ]
  to deeply equal []`.
- **Fix:** `datetime` only when `Number.isFinite(new Date(ms).getTime())`;
  the time text falls back to the raw number.
- **Contract:** no.

### UR-P4-ui-306 — low — the controller's endpoint tree hides an orphan and recurses without end on a self-owned entry

- **Where:** `ui/src/lib/pages/controller/node/Endpoints.svelte:15-24`.
- **Claim:** the tree walks only from endpoints without a parent: an
  endpoint whose `parent` is not in the list (or a cycle) is not shown at
  all; an entry whose `parent` equals its own number, when that number is
  also reached from a root (a repeated number — the contract makes the
  list unique nowhere), recurses until `Maximum call stack size exceeded`
  and the tab renders nothing. The parent comes from the node's PartsList
  through go-fabric's `Owner()` (`fabricui/gofabric/node.go:158-165`):
  node data. **M** (page); whether go-fabric can hand such a list over **I**.
- **Test:** `w-cover.test.ts` › "UR-P4-ui-306 … lists every endpoint, an
  orphan …" — red: `Unable to find an element by: [data-testid="endpoint-7"]`;
  › "survives an endpoint that names itself as its parent" — red:
  `expected [ 'Maximum call stack size exceeded' ] to deeply equal []`.
- **Fix:** walk with a `seen` set (by list index), then append every entry
  not reached at depth 0.
- **Contract:** no.

### UR-P4-ui-303 — note — a session's peer always links to a controller node page

- **Both sides:** `pages/controller/Diagnostics.svelte:138` links every
  session's `peer_node_id_hex` to `routes.controllerNode`, PASE sessions
  (`is_pase`, peer id 0) and peers the controller has not commissioned
  included; the peers table beside it links only `registered` peers
  (`:152-157`). The link leads to a "No such node" error page.
- **Fix:** link only when the peer is a registered peer of the same
  answer (or a listed node), otherwise plain text.

### UR-P4-ui-304 — note — a host schema naming a column twice takes the selection page down

- **Both sides:** `pages/node/Selection.svelte:248-256` builds a column
  `field:<key>` per schema column; `kit/DataTable.svelte:190`, `:215` key
  them by `col.key`; the contract's `ExposureSchema.columns` does not say
  the keys are unique (`openapi.yaml:2736-2746`, `Column` `:2712-2719`), and
  the each-keys guard allows the key as "the columns the page defines".
  A provider listing one key twice: `each_key_duplicate`, nothing renders.
  Test (optional for a note): `w-cover.test.ts` › "UR-P4-ui-304 …" — red:
  `expected [ 'each_key_duplicate' ] to deeply equal []`.
- **Fix:** drop repeated schema keys in the page (first wins), or key the
  columns by position in DataTable; a sentence in the contract.

### UR-P4-ui-305 — note — a failed reload of the node topology blanks the drawn graph

- **Both sides:** `pages/node/Topology.svelte:33-46` sets `error` on any
  failed load, and the `{:else if error}` branch replaces the graph; the
  comment says "a reload keeps the drawn graph until the new one arrives".
  `pages/controller/Network.svelte:41-58` keeps the graph and toasts. A
  stream-driven reload that fails once (`graphEpoch`) leaves an error box
  instead of the last graph. Test (optional): `w-cover.test.ts` ›
  "UR-P4-ui-305 …" — red: `expected null not to be null`.
- **Fix:** as `Network.svelte`: with a graph drawn, toast the error.

## Covered without finding

Line by line: the delta files above (beyond the findings); 
`pages/controller/{Commission,Settings,Nodes,Node}.svelte`,
`pages/controller/node/{Attributes,Bindings (to :80),Icd,Fabrics,Inventory,Events,Overview (script)}.svelte`,
`pages/controller/{format.ts,QrScan.svelte (script)}`,
`pages/node/{Selection.svelte (:1-260),Fabrics.svelte (:1-150),Status.svelte (:20-80),NodeManage.svelte (:1-70)}`,
`pages/operation/{Operation.svelte (:1-130),log.ts}`, `kit/{DataTable.svelte (:1-130),data-table.ts,utils.ts,events.ts}`,
`scan/index.ts (:30-94)`, `shell/NodeCreateDialog.svelte (:1-40)`,
`context.ts` (routes, `refHref`, `createFabricUI`), `index.ts`,
`e2e/origins.spec.ts`, `e2e/smoke.spec.ts (:1-60)`,
`scripts/{check-notices.mjs,gen-types.mjs}`, the `notices-check` target
and `vite.config.ts`'s vendor list; guards: `i18n/coverage.test.ts`,
`contract-guard.test.mjs`, `each-keys.test.ts (:1-50)`,
`i18n/catalogs.test.ts (:1-40)`, `theme/tokens.test.ts (:40-146)`,
`test/{support.ts,Harness.svelte,NodeHarness.svelte}`.
A script over every literal `t("key", {…})` call checked that each passes
the placeholders its English text names: none missing.

## Covered in the coverage pass

Read fully after the first report, without a finding beyond 299–306:
`pages/controller/{Diagnostics,Network}.svelte`,
`pages/controller/node/Endpoints.svelte`, the markup of `Bindings.svelte`
and `Overview.svelte`,
`pages/node/{Diagnostics,Surface,Endpoint,Endpoints,Topology,NodeConfig,EcosystemHints,HostCounters}.svelte`,
`pages/node/{fields,surface,ecosystems}.ts`, the rest of `Selection.svelte`,
node `Fabrics.svelte`, `Operation.svelte` and `DataTable.svelte`;
`kit/{Graph,Tabs,Select,Button,Badge,Card,ClusterView,AttributeTable,EventList,TagChips,Icon,Input,PageHeader,PageShell,Spinner,EmptyState,ErrorState,LoadingState,DeniedState}.svelte`,
`kit/{graph-layout,icons,im-status,dates,tabs,index}.ts`;
`theme/tokens.css` (all four palettes), `theme/theme.test.ts` (`:160-200`);
`shell/{ConfirmDialog,NodeFamily,NodeScope,Toaster,NodeCreateDialog}.svelte`,
`shell/{confirm.svelte.ts,toast.svelte.ts,focus-trap.ts,index.html,main.ts,app.css}`;
`e2e/smoke.spec.ts` (`:60-237`) and `playwright.config.ts` (serial, one
worker, so the order-dependent erase and commission smokes hold).

## Not covered

None in the unit's scope. Read as targets only by unit S-web, not here:
`lib/api/{client,stream,json}.ts`, the token path, the three stores,
`shell/{App.svelte,route.ts}`, `theme/ha-bridge.ts`, `pages/node/Pair.svelte`,
`pages/controller/node/Window.svelte` (read here only where a page of this
unit calls them). The page and kit tests outside the delta were read as
helpers and for the guards they hold, not each line by line; the
catalogues were checked by the guards and the placeholder script, not key
by key; `qr.ts`, `countdown.ts` and `i18n/index.ts` were read in run 1 and
did not move (delta map), and were skimmed only here.

## Observations, not findings

- `controller/node/Overview.svelte`: after `setNodeAutoSubscribe` the
  button's label comes from `detail`, which reloads only on a stream event
  about the node; whether go-fabric emits one for a node with no
  subscription was not checked at the pin (**I**). The memory controller
  publishes `EventSubscription` (`fabricui/memory_controller.go:1167`).
- `notices-check` removes `$(CHECK_DIR)/notices` but not
  `$(CHECK_DIR)/vendor.json`; a build that stopped writing the list would be
  checked against the previous run's file (local runs only; CI starts clean).
- `kit/graph-layout.ts` `layoutGraph` returns `nodes.map(…)`, so a repeated
  topology node id would repeat a key in `Graph.svelte`'s keyed each; the
  contract says the id is "Unique in the graph" (`openapi.yaml:2475`), so
  the page trusts it (whether every server path keeps it: units S-go/G).
- The eleven controller-node tab smokes (`e2e/smoke.spec.ts:30-34`) assert
  only the page title; a tab whose content fails is caught only by the
  console-error check.
- `Operation.svelte` `togglePause` merges the held records by `seq`
  (`mergeRecords`); a daemon restart while paused interleaves two
  sequences, which `appendRecord` avoids for the live path.

Tokens: about 150k for the first pass, about 110k for the coverage pass (about 260k in all).
