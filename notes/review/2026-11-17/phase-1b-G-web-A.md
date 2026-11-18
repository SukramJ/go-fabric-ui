# Phase 1b — G-web A (node pages, operation page, kit, countdown)

Reviewer G-web A (Opus), branch `review/p1b` at `007ca96`. Files read in
full: `ui/src/lib/pages/node/*`, `pages/operation/*`, `ui/src/lib/kit/*`,
`ui/src/lib/countdown.ts` and their vitest files, `ui/src/test/support.ts`;
`contract/openapi.yaml` and `store.svelte.ts` by slice. Probes ran as a
scratch vitest file outside the repo (config `root: ui/`, include only the
probe), see "Probes". IDs `UR-P1-<theme>-130…145`.

## UR-P1-err-04 — the page part (settled)

- **Answer:** no page reads `status.window_kind`, `pairing.source` or
  `pairing.window_open`. `grep -rn window_kind ui/src/lib` finds only
  `schema.gen.ts`. `Pair.svelte:27-35` calls `node.hydrateCommissioning()`,
  which (`store.svelte.ts:254-279`) checks `st.enabled` and
  `st.commissioning_window_open` (`:261`), fetches `client.pairing()`
  (`:262`) and puts its `qr_code`/`manual_code` into phase `open`
  unconditionally. `Pair.svelte:122-176` then shows them as the open
  window's codes with the countdown. `Status.svelte:93-99` shows only "window
  open" and the seconds left. Nothing tells the user that the codes cannot
  commission through that window.
- **State:** M (read; run: probe P2, a status with `window_kind: "enhanced"`
  and a pairing answer `source: configured, window_open: true` renders
  manual code `34970112332` in phase `open`).
- **Page-only fix (no contract change):** `window_kind` is already in
  `Status` (`openapi.yaml:1967-1970`), so the page can do this without a
  contract change. In `hydrateCommissioning`, after `client.pairing()`:
  (1) `if (!p.window_open) return;` (the window closed between the two
  reads); (2) when `p.source === "configured"` and `st.window_kind` is
  present and not `"own"` (`basic` or `enhanced`), do not enter `open`.
  Enter a new phase `foreign` with no codes. `Pair.svelte` renders it as a
  card: "A commissioning window is open that this page did not open
  ({kind}); its codes are not known here. Close it to open a new one." It
  has the existing Close button (admin). Also check `source` when
  `window_kind` is absent, which means an older server: keep today's
  behaviour. A `basic` window this handler opened answers `source: window`
  and still shows its codes. Optional: `Status.svelte:93-99` names the
  kind in the window badge. Two catalogue keys (en, de). The fix lives in
  `store.svelte.ts`, which is S-web's file. Only the rendering is in
  `Pair.svelte`.
- **Red test:** Pair on status `{commissioning_window_open: true,
  window_kind: "enhanced"}` and pairing `{source: "configured",
  window_open: true}` shows no manual code and no QR, and shows the
  foreign-window notice. With `window_kind: "own"` it shows the codes as
  today.

## Findings

### UR-P1-ui-130 — two mDNS findings with the same code and no service crash the Diagnostics page

- **Severity:** medium
- **Where:** `ui/src/lib/pages/node/Diagnostics.svelte:183` (key
  `finding.code + (finding.service ?? "")`). The same key is at
  `pages/controller/Diagnostics.svelte:198`, which is G-web B's file.
- **Claim:** go-fabric's `DiagnoseMDNS` appends `mdns.DiagnoseInterfaces`.
  At `v0.22.2` that gives one finding per named interface
  (`gf:mdns/diagnose_interfaces.go:48-74`, codes `interface_not_found`,
  `interface_filtered`, `interface_no_addresses`, no `Service`). Two
  misnamed interfaces in `[MDNS].Interfaces` therefore yield two findings
  with the key `interface_not_found`. Svelte 5 throws `each_key_duplicate`
  for a duplicate key in production as well as in dev
  (`svelte/src/internal/client/dom/blocks/each.js:357-363`). The
  Diagnostics page fails on exactly the misconfiguration it is there to
  explain. The compatibility list keys `finding.ecosystem + finding.code`
  (`:249`). That is unique at `v0.22.2` (`gf:eligibility/compat.go:115-170`),
  but the contract does not promise it.
- **State:** M (run: probe P1, `Svelte error: each_key_duplicate … duplicate
  key interface_not_found at indexes 0 and 1`, in `Card.svelte` in
  `Diagnostics.svelte`). Which go-fabric lines produce the pair was read.
  The page state after the throw is I.
- **Fix:** key both lists by position (`(i)`), as the events list beside
  them already does (`:270`). Apply the same to the controller page.
- **Contract change:** no.
- **Red test:** Diagnostics with `mdns.findings = [{error,
  interface_not_found, "eth9"}, {error, interface_not_found, "wlan9"}]`
  shows both messages and the sessions card.

### UR-P1-ui-131 — a cluster_event that lands during an incremental events read is listed twice and crashes the endpoint page

- **Severity:** low
- **Where:** `ui/src/lib/pages/node/Endpoint.svelte:58-69`
  (`loadEvents(false)` appends `l.events` without a dedupe, `:64`),
  `:97-105` (cluster_event append); `ui/src/lib/kit/EventList.svelte:27`
  (keyed by `e.number`)
- **Claim:** `loadEvents(false)` takes `last` before the request. If event
  N > last arrives on the stream while the request is in flight, `:103`
  appends it. The response also contains N, and `:64` appends it again. The
  keyed each in `EventList` then throws `each_key_duplicate`. The trigger is
  a burst of `attribute_changed` (the refetch) followed within one round
  trip by an event of the same endpoint, for example a switch pressed twice.
- **State:** M (run: probe P3, `duplicate key 12 at indexes 0 and 1`).
- **Fix:** merge by number in `loadEvents`, as `mergeRecords` does for the
  log: drop the numbers already present and sort by number.
- **Contract change:** no.
- **Red test:** the incremental read is held open, cluster_event 12 is
  applied, then the read answers `[12]`: exactly one `event-12`, and no
  error.

### UR-P1-ui-132 — the Pair page ignores a window opened elsewhere while it is mounted

- **Severity:** low
- **Where:** `ui/src/lib/pages/node/Pair.svelte:27-35` (hydrates once, on
  mount); `ui/src/lib/store.svelte.ts:321-323` (`window_opened` only
  reloads the status)
- **Claim:** A window opened while the Pair page shows the idle form
  (another tab, another admin, the host) does not move the page. It keeps
  offering "Open window" until the user leaves and comes back. The
  `window_closed` path does end an open flow (`store.svelte.ts:328-342`),
  so the two directions are not symmetric.
- **State:** I (read; no hydrate call on `window_opened` in the store or
  the page).
- **Fix:** on `window_opened`, when `commissioning.phase === "idle"`, call
  `hydrateCommissioning()` after the status reload. That path also takes
  the err-04 rule above.
- **Contract change:** no.
- **Red test:** Pair mounted with no window, then `applyEvent({type:
  "window_opened"})` with status and pairing now open: the codes appear.

### UR-P1-ui-133 — admin controls are shown to a non-admin on four pages; the Pair page's 403 reads as a load error

- **Severity:** low
- **Where:** `Pair.svelte:116`, `:173` (open, close), `:97-98` (generic
  `node.pair.load_error`); `Fabrics.svelte:204`, `:216`, `:220`, `:225`
  (remove, reassemble, reset, erase); `Selection.svelte:306-312`, `:339`
  (preview, apply); `Operation.svelte:138` (level). The routes behind them
  are all `x-admin` in the contract. Contrast `Diagnostics.svelte:26`,
  `:170` and `NodeManage.svelte:27`, `:91`, which gate with
  `!!status.csrf_token || status.admin === true`.
- **Claim:** In host mode with `Admin` answering false, a viewer gets
  buttons whose only outcome is a 403 toast. On the Pair page with a window
  open, `GET /node/pairing` is admin-only, so the hydrate fails and the
  whole flow is replaced by "could not load" with no reason given.
  `threat-model.md` 3.2 states that "the admin controls are hidden by
  `status.admin` or the token's presence". That holds for two pages, not
  six.
- **State:** M (read; contract `x-admin` list).
- **Fix:** one `isAdmin` derived value in the context (it is repeated three
  times today: `App.svelte:107`, Diagnostics, NodeManage). Hide or disable
  the controls on the four pages. Pair: for a non-admin, show the window
  state (open, seconds left from the status) and say that the codes need
  admin rights; do not fetch `/pairing`.
- **Contract change:** no.
- **Red test:** each page rendered on a status with `admin: false` and no
  `csrf_token` shows none of the write buttons. Pair with an open window
  makes no `GET /node/pairing`.

### UR-P1-bound-134 — the paused log keeps every streamed record without bound

- **Severity:** low
- **Where:** `ui/src/lib/pages/operation/Operation.svelte:34`, `:40-45`;
  release at `:74-81`
- **Claim:** While paused, `append` pushes into `held` with no cap. The
  visible list is capped at `KEEP_RECORDS` (`log.ts:179-187`). A tab left
  paused at level `debug` grows without limit, and the resume merges and
  sorts the whole backlog in one step.
- **State:** M (read).
- **Fix:** cap `held` at `KEEP_RECORDS`, dropping the oldest. Keep
  `heldCount` as the total received, so the banner can say "n received, the
  newest 2000 kept".
- **Contract change:** no.
- **Red test:** pause, deliver 2500 records, resume: at most 2000 records
  are rendered, and `held` never holds more than 2000.

### UR-P1-ui-135 — the Selection preview goes stale: Apply writes what was not previewed

- **Severity:** low
- **Where:** `ui/src/lib/pages/node/Selection.svelte:150-153`,
  `:162-166` (bulk and drawer edits leave `previewItems` as it is);
  `:179-190` (`apply` → `node.saveBulk()`, which sends the current
  `pending`, `store.svelte.ts` `saveBulk`); `:316-343` (preview card)
- **Claim:** After Preview, the save bar and the table stay usable. A bulk
  expose or a drawer save changes `pending`, and the preview card keeps
  showing the old verdicts. Apply then writes the new set. The page's own
  "Preview before Apply" step (header comment `:7-9`) is bypassed without
  the user noticing.
- **State:** M (read).
- **Fix:** clear `previewItems` on every `markDirty` from the page (bulk,
  drawer), or snapshot `pending` at preview time and apply that snapshot.
- **Contract change:** no.
- **Red test:** mark A, Preview, then bulk-mark B: the preview card is
  gone, or Apply sends only A.

### UR-P1-ui-136 — the Selection drawer is not a dialog: no focus handling, and Escape does nothing

- **Severity:** low
- **Where:** `ui/src/lib/pages/node/Selection.svelte:372-440`; Escape
  handler on the backdrop at `:374-382`
- **Claim:** The drawer opens as a plain `<aside>`. Focus is not moved into
  it, not kept in it, and not returned to the row's button on close. There
  is no `role="dialog"` or `aria-modal`. The Escape handler sits on a
  `tabindex="-1"` backdrop that is never focused, so Escape from the name
  field or anywhere else does not close the drawer. Keyboard users tab
  through the table behind it.
- **State:** M (read).
- **Fix:** `role="dialog" aria-modal="true" aria-labelledby` on the aside.
  Focus the name input on open, restore focus to the opener on close, and
  handle Escape on the aside or on `window` while it is open. A native
  `<dialog>` with `showModal()` gives all of this.
- **Contract change:** no.
- **Red test:** open the drawer: `document.activeElement` is the name
  input. Escape keydown on it closes the drawer and focus returns to the
  name button.

### UR-P1-ui-137 — Endpoints, Surface and Topology reload without a sequence guard; one reassembly fetches twice

- **Severity:** note
- **Where:** `Endpoints.svelte:30-45`, `:51-54`; `Surface.svelte:30-49`,
  `:54-57`; `Topology.svelte:29-40`, `:44-47`. Trigger:
  `store.svelte.ts` `endpoints_changed`/`surface_changed` each do
  `topologyEpoch++` without the debounce the graph epoch has.
- **Claim:** A reassembly emits both events, so `topologyEpoch` moves twice
  and two `load()` calls overlap. Refresh plus a stream move does the same.
  Whichever answer arrives last wins, older or not. On Topology, Refresh is
  enabled during a reload (`:31` keeps `loading` false once a graph is
  drawn). Diagnostics (`:39-60`) and Endpoint (`:179-206` of the page) have
  the `seq` guard; these three do not.
- **State:** I (read).
- **Fix:** the `mine === seq` guard of `Diagnostics.svelte:43-61` in all
  three. Optionally debounce `topologyEpoch` as `graphEpoch` is.
- **Contract change:** no.
- **Red test:** two loads with the first answer delayed past the second:
  the page shows the second answer.

### UR-P1-ui-138 — Tabs: tab roles without the tab pattern

- **Severity:** note
- **Where:** `ui/src/lib/kit/Tabs.svelte:49-65`
- **Claim:** `role="tablist"`/`role="tab"` with `aria-selected`, but no
  `aria-controls`, no `tabpanel`, no arrow-key movement and no roving
  `tabindex`: every tab is a Tab stop. On the `href` variant, links with
  `role="tab"` navigate, which screen readers do not expect of a tab.
- **State:** M (read).
- **Fix:** route tabs: drop the tab roles and use
  `<nav aria-label>` + `aria-current="page"`. Button tabs: arrow keys,
  roving `tabindex`, `aria-controls`.
- **Contract change:** no.
- **Red test:** route variant renders `aria-current="page"` on the active
  link and no `role="tab"`. Button variant: ArrowRight moves focus to the
  next tab.

### UR-P1-ui-139 — Graph keys edges by source and target, which the contract does not make unique

- **Severity:** note
- **Where:** `ui/src/lib/kit/Graph.svelte:35`, `:185`
- **Claim:** `TopologyNode.id` is "unique in the graph"
  (`openapi.yaml:2192`). `TopologyEdge` says nothing about uniqueness
  (`:2210-2219`). Two edges between the same pair (say a `session` and a
  `link`) collide in `edgeOf` (the second wins) and in the keyed each,
  which throws `each_key_duplicate` as in UR-P1-ui-130. Whether a server
  emits such a pair at `v0.22.2` was not checked (controller topology:
  G-web B, server).
- **State:** I.
- **Fix:** key by index, or by `source>target>kind`. State the uniqueness
  of edges in the contract if it is meant (that is a contract text change).
- **Contract change:** no for the key; yes if uniqueness is stated.
- **Red test:** a graph with edges `a>b contains` and `a>b link` renders
  both paths without an error.

### UR-P1-ui-140 — catalogue keys built from contract strings without an enum, without a fallback

- **Severity:** note
- **Where:** `Diagnostics.svelte:242`, `:251`; `EcosystemHints.svelte:38`
  (all `node.diag.ecosystem.${…}`)
- **Claim:** `Ecosystem.ecosystem` and `CompatFinding.ecosystem` are
  `type: string` with no enum (`openapi.yaml:2272`, `:2280`). A value the
  catalogue does not name renders as the raw key
  `node.diag.ecosystem.<value>`. At `v0.22.2` all seven values are in
  `en.json`. `Graph.svelte:82-86` (`detailName`) already shows the pattern
  that falls back to the value itself.
- **State:** M (read).
- **Fix:** the `detailName` pattern (key unchanged → show the value), or an
  enum in the contract.
- **Contract change:** no (yes if an enum is added).
- **Red test:** a compatibility answer with `ecosystem: "tuya"` shows
  `tuya`, not `node.diag.ecosystem.tuya`.

### UR-P1-ui-141 — hard-coded English and units in the kit and two pages

- **Severity:** note
- **Where:** `ui/src/lib/kit/TagChips.svelte:15-16` (title "namespace … ·
  tag … · mfg …"); `Pair.svelte:144` (`{remaining}s`);
  `Status.svelte:98` (`({windowLeft}s)`)
- **Claim:** These strings bypass `t`. The seconds suffix is not a unit
  symbol that every locale writes the same way, and screen readers read
  "300s" as "300 s".
- **State:** M (read).
- **Fix:** catalogue keys (`kit.tag_title`, `common.seconds_short`) with
  `{n}`.
- **Contract change:** no.
- **Red test:** under `echoT`, the tag chip's title and the countdown render
  their keys.

### UR-P1-ui-142 — the selection search runs the user's term as a regular expression on every keystroke

- **Severity:** note
- **Where:** `ui/src/lib/kit/utils.ts:17-30` (`makeTextMatcher`);
  `ui/src/lib/pages/node/Selection.svelte:99-110`
- **Claim:** A term that is a valid but catastrophic regex (`(a+)+$`, typed
  one character at a time) runs against every field of every item, and a
  long `aaaa…b` host name freezes the tab. The user does this to
  themselves, but on host-supplied text. `DataTable`'s own search
  (persisted, so it would replay on mount) is used by no page.
- **State:** I (behaviour of backtracking regexes; not run).
- **Fix:** use the regex only when the term contains `|` and no nested
  quantifier, or debounce and cap the time per match. The simplest option
  is to keep substring matching and treat `|` as OR.
- **Contract change:** no.
- **Red test:** `makeTextMatcher("(a+)+$")("a".repeat(40) + "b")` returns
  within 50 ms.

### UR-P1-ui-143 — the log page re-sorts the whole ring per streamed record; a daemon restart mixes two sequences

- **Severity:** note
- **Where:** `ui/src/lib/pages/operation/Operation.svelte:46`;
  `ui/src/lib/pages/operation/log.ts:13-18`
- **Claim:** (a) Each record builds a Map and sorts up to 2000 records, so
  at `debug` the cost is O(n log n) per record on the main thread. (b)
  `seq` restarts at 1 with a new daemon process. After the log stream
  reconnects, new records with the same `seq` replace old ones in the Map
  and sort among them.
- **State:** (a) M read; (b) I (reconnect behaviour is `stream.ts`, S-web).
- **Fix:** (a) append when `rec.seq > last.seq`, merge only otherwise.
  (b) clear `records` when a record arrives with `seq` lower than the
  first held, or key by `started_at` + `seq` from `operationInfo`.
- **Contract change:** no.
- **Red test:** (b) records 1..5, then a record with seq 2 and a later time:
  the list shows 6 entries (or only the new process's records), not 5.

### UR-P1-ui-144 — Endpoint keeps the first endpoint's baselines when its prop changes without a remount

- **Severity:** note
- **Where:** `ui/src/lib/pages/node/Endpoint.svelte:83-84` (`startAttr`
  of the first `endpoint`), `:76-79` (refetch on the prop), `detail` not
  cleared
- **Claim:** The shell remounts per endpoint (`shell/NodeFamily.svelte:29`
  `{#key route.endpoint}`). A host that passes a new `endpoint` to a mounted
  page instead sees the old endpoint's detail until the new answer
  arrives. `startAttr` compares the new endpoint's epoch with the old
  one's, so the page may refetch at once and on every later move.
- **State:** I (read; the kit is exported for hosts, `lib/index.ts`).
- **Fix:** document that the page must be keyed by endpoint, or reset
  `detail`, `events` and the baselines inside the `endpoint` effect.
- **Contract change:** no.
- **Red test:** render with `endpoint: 2`, rerender with `endpoint: 3`: no
  `endpoint-title` of 2 is visible after the rerender starts.

### UR-P1-ui-145 — Fabrics: Remove has no busy state

- **Severity:** note
- **Where:** `ui/src/lib/pages/node/Fabrics.svelte:143-159`, `:204`
- **Claim:** Reassemble, reset and erase set `busy`; `unpair` does not. A
  double click on Remove opens two confirmations (with a host `confirm`
  that does not serialise them). Confirming both sends two `DELETE
  /node/fabrics/{i}`, and the second answers an error toast after the first
  succeeded.
- **State:** I (depends on the host's `confirm`).
- **Fix:** a per-row pending set, with Remove disabled while its delete
  runs.
- **Contract change:** no.
- **Red test:** two clicks on Remove with a confirm that resolves true
  send one DELETE.

## Read and found nothing

| File | What was checked |
| --- | --- |
| `Fabrics.svelte`, `fields.ts` | 64-bit rule: `node_id_hex` first, the number only as a flagged fallback (`:40-49`); no fabric or node id parsed to a number anywhere in these files |
| `Endpoint.svelte`, `Surface.svelte`, `surface.ts` | `hexId` and `Number(id)` only on 16/32-bit ids (cluster, device type, element) |
| `Status.svelte`, `countdown.ts` | interval cleared on unmount; `secondsUntil` never negative, null on a bad deadline; `boot_reason` 0–6 all in the catalogue |
| `Diagnostics.svelte` | `seq` guard; the quiet reload; admin-gated reannounce with a busy state; events keyed by index |
| `NodeManage.svelte` | admin gate, two confirmations, busy, length bound 64 as the field |
| `HostWarnings`, `HostCounters`, `NodeConfig`, `EcosystemHints`, `ecosystems.ts` | optional fields read as optional; passcode not shown |
| `Operation.svelte`, `log.ts` | subscription closed on unmount; 503/404 → unavailable; the level reverted on refusal |
| `DataTable.svelte`, `data-table.ts` | sort compare (nulls last, numeric), `aria-sort`, group order, row activation skips controls |
| `Graph.svelte`, `graph-layout.ts` | node ids unique by contract; Enter, Space, Escape, Tab; pointer capture released; zoom bounds |
| `Select`, `Input`, `Button`, `Badge`, `Card`, `EmptyState`, `ErrorState`, `LoadingState`, `Icon`, `icons.ts`, `im-status.ts`, `PageHeader`, `PageShell`, `Spinner`, `ClusterView`, `AttributeTable`, `TagChips`, `EventList` | props, keys, the fallback icon, the IM status map |

## Probes

Scratch file `…/scratchpad/p1b/probe.test.ts` (outside the repo), vitest
with `root: ui/`:

- P1: Diagnostics with two findings `{error, interface_not_found}` and no
  service → `each_key_duplicate … interface_not_found at indexes 0 and 1`.
- P2: Pair on status `{commissioning_window_open: true, window_kind:
  "enhanced"}` and pairing `{source: "configured", window_open: true}` →
  manual code `34970112332` rendered, phase `open`.
- P3: Endpoint 2, `attribute_changed` → incremental read held open;
  `cluster_event` 12 applied; the read answers `[12]` →
  `each_key_duplicate … 12 at indexes 0 and 1`.

## Count

16 findings: 0 high · 1 medium · 6 low · 9 note. Plus
the page part of UR-P1-err-04, settled above: the page reads none of
`window_kind`, `source` and `window_open`; the fix needs no contract
change.
