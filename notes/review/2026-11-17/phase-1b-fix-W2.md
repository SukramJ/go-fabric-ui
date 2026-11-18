# Phase 1b — fix wave, agent W2

Branch `review/p1b-w2` from `869b99b`, worktree
`/home/markus/Dokumente/GitHub/go-fabric-ui/.claude/worktrees/agent-a17472218e8838b61`.
Local commits only (no push, PR or tag). `contract/openapi.yaml` and
`fabricui/dist` untouched; no `make web`. Files touched: only W2's
(pages except Pair/QrScan/Attributes, kit except DataTable, theme/tokens.css),
their tests, the catalogues and `i18n/coverage.test.ts` (UR-P1-test-173).

## Commits (oldest first)

fda6ee0 ui-130 · b638400 ui-131 · 7253164 ui-144 · f7fda28 ui-133 ·
302c720 bound-134 + ui-143 · 013ea77 ui-135 + ui-136 · 80bdd47 ui-137 ·
fdda463 ui-138 · 09cef98 ui-139 · ce73f4f ui-140 · 719d7fd ui-141 ·
41b27b1 ui-142 · 6f256c0 ui-145 · 816de41 ui-163 + conc-180 (Node) ·
a124104 ui-181 + conc-180 (Network) · a829368 ui-164 + conc-165 ·
51cbbc0 ui-166 · 59b29ca ui-167 · be2984f ui-169 · 6e24ef1 ui-170 ·
a684c8a ui-172 (partial) · ee3ff55 ui-174 + ui-175 · 9d0cb5d ui-176 ·
6f4359c ui-182 · 7c77476 test-173

Every fix: the test was run red first, then the fix, then green.

## G-web A (130–145, 132 excepted)

| ID | Result | Commit | Test (file › name) |
| --- | --- | --- | --- |
| ui-130 | fixed, both Diagnostics pages and the node compat list keyed by index | fda6ee0 | node Diagnostics.test.ts › "renders two interface findings with the same code and no service"; controller Diagnostics.test.ts › "renders two findings with the same code and no service" |
| ui-131 | fixed, incremental events merged by number | b638400 | Endpoint.test.ts › "lists a cluster_event that lands during an incremental read once" |
| ui-133 | fixed for Fabrics, Selection, Operation (Pair is W1's) | f7fda28 | Fabrics/Selection/Operation.test.ts › "… admin rights (UR-P1-ui-133)" |
| bound-134 | fixed, held capped, banner names the drop (operation.paused_dropped) | 302c720 | Operation.test.ts › "keeps at most the newest KEEP_RECORDS while paused …" |
| ui-135 | fixed, preview cleared on bulk change and drawer save | 013ea77 | Selection.test.ts › "drops the preview when the pending set changes after it" |
| ui-136 | fixed, role=dialog + aria-modal, focus in/trap/return, Escape anywhere in it | 013ea77 | Selection.test.ts › "opens the drawer as a dialog …" |
| ui-137 | fixed, seq guard on Endpoints, Surface, Topology (no topologyEpoch debounce: store is W1's) | 80bdd47 | Diagnostics.test.ts › "Endpoints — overlapping loads"; Surface.test.ts, Topology.test.ts › "… overlapping loads" |
| ui-138 | fixed: route strips → nav + aria-current; button strips → roving tabindex, arrows/Home/End; no aria-controls (no page renders a tabpanel) | fdda463 | kit/Tabs.test.ts (2 cases) |
| ui-139 | fixed, LaidEdge.index, keyed and looked up by it | 09cef98 | Graph.test.ts › "draws two edges with the same source and target …" |
| ui-140 | fixed in Diagnostics (detailName pattern); EcosystemHints keys only its own static list | ce73f4f | node Diagnostics.test.ts › "shows an unknown ecosystem as its value, a known one by its label" |
| ui-141 | fixed for TagChips and Status (kit.tag_title, kit.tag_title_mfg, common.seconds_short); Pair is W1's | 719d7fd | kit/TagChips.test.ts; Status.test.ts countdown cases |
| ui-142 | fixed: no regex, `|` splits literal alternatives (red run took 8.9 s) | 41b27b1 | kit/utils.test.ts (2 cases) |
| ui-143 | (a) append without sort — cost item, no red test; (b) fixed, a restart drops the previous process's records | 302c720 | Operation.test.ts › "starts over when a restarted daemon numbers its records from the start again" |
| ui-144 | fixed, endpoint effect resets detail/error/events and both baselines | 7253164 | Endpoint.test.ts › "drops the previous endpoint and its baselines when the prop changes" |
| ui-145 | fixed, per-row removing set | 6f256c0 | Fabrics.test.ts › "sends one DELETE for two clicks on Remove" |

## G-web B (163–182; 160/162 W1; 168, 171, 177, 178, 179 phase 3)

Corrected file:line at 869b99b (under ui/src/lib/):

| ID | Result | Commit | Test | Corrected file:line |
| --- | --- | --- | --- | --- |
| ui-163 | fixed per the verifier: nodeId effect clears detail/error/focus, tab body {#key nodeId}, load() seq-guarded | 816de41 | Node.test.ts › "drops the previous node's tab state when nodeId changes in place" | pages/controller/Node.svelte:41-48 (load), :50-53 (nodeId effect), :74-102 (tabs); node/Acl.svelte:45, node/Bindings.svelte:34, node/Fabrics.svelte:31, node/Icd.svelte:30 (onMount only); node/Attributes.svelte:18 |
| ui-164 | fixed, parseUint16 (endpoint 0..65534, group 0..65535) | a829368 | Node.test.ts › "refuses an ACL target endpoint …", "refuses a binding endpoint or group out of range" | node/Acl.svelte:62; node/Bindings.svelte:56, :59 |
| conc-165 | fixed, busy disables add/remove until write + reload | a829368 | Node.test.ts › "sends one binding write at a time" | node/Bindings.svelte:41-50, :100, :109 |
| ui-166 | fixed: Close always offered (with hint), countdown, codes leave at zero | 51cbbc0 | Node.test.ts › "offers Close after the tab was left and shown again", "takes the codes off the page when the window's time is up" | node/Window.svelte:16, :59-79 (:64, :78); tab unmount Node.svelte:74-102 |
| ui-167 | fixed page-only (loading / queue_unknown) | 59b29ca | Node.test.ts › "does not claim an empty queue when the queue could not be read" | node/Updates.svelte:19, :21, :54-55; controller-store.svelte.ts:111-117 |
| ui-169 | fixed, busy is a Set | be2984f | Nodes.test.ts › "keeps a node's actions disabled while its call runs …" | pages/controller/Nodes.svelte:49, :77-87, :158-171 |
| ui-170 | fixed, device dropped only when result is set | 6e24ef1 | Commission.test.ts › "keeps the device listed when its pairing fails" | pages/controller/Commission.svelte:189-191 |
| ui-172 | partial: ours fixed, the host's left | a684c8a | theme/tokens.test.ts › "reaches 4.5:1 on the card background", light and dark | standalone theme/tokens.css:50-51, :54; HA skin :93, :102-104; text uses e.g. pages/controller/Diagnostics.svelte:123, :195; node/Diagnostics.svelte:180, :246 |
| test-173 | fixed: typed Records per enum, en and de; placeholder parity added (green today). Red = npm run check with the old three diagnostic kinds | 7c77476 | coverage.test.ts › "name every value of the contract's enumerations the pages translate", "use the same placeholders in English and German" | i18n/coverage.test.ts:43-56; i18n/catalogs.test.ts:51-52, :96 |
| ui-174 | fixed, badge + key | ee3ff55 | Settings.test.ts › "shows the diagnostics capability" | pages/controller/Settings.svelte:155 |
| ui-175 | fixed: kind change clears SSID and secret; dataset must be hex bytes | ee3ff55 | Settings.test.ts › "clears the typed secret when the kind changes, and refuses a dataset …" | pages/controller/Settings.svelte:25, :49-54, :130 |
| ui-176 | fixed, destructive confirm before forget | 9d0cb5d | Node.test.ts › "asks first, and sends nothing when the answer is no" | node/Icd.svelte:51-61, :86 |
| conc-180 | fixed (Node and Network) | 816de41, a124104 | Node.test.ts › "shows the newer detail when an older reload answers last"; Network.test.ts › "shows the newer graph when an older reload answers last" | pages/controller/Node.svelte:41-48, :56-59; pages/controller/Network.svelte:37-49, :54-65, :70-72 |
| ui-181 | fixed, failed reload with a graph → toast | a124104 | Network.test.ts › "keeps the drawn graph and toasts when a stream reload fails" | pages/controller/Network.svelte:37-49 (:46), :94 |
| ui-182 | fixed in all pages and kit (kit/dates.ts, locale from new key common.locale) | 6f4359c | Node.test.ts › "formats the commissioning time in the UI's language, not the runtime's" (en and de: red on any machine) | pages/controller/Commission.svelte:108; pages/controller/Diagnostics.svelte:88, :256; node/Icd.svelte:80-81; node/Overview.svelte:77; same pattern also kit/EventList.svelte:19, pages/node/Diagnostics.svelte:279, HostWarnings.svelte:22, Surface.svelte:273, operation/Operation.svelte:173 |
| ctr-168, api-171, ctr-177, ui-178, ui-179 | not fixed: phase 3 per the brief | — | — | — |

## Partial / not fixed

- ui-172: ours (standalone light warning #d97706 3.19, success #059669 3.77, info sky-600 4.02) now amber-700 / emerald-700 / sky-700 (5.0 / 5.5 / ≥4.5). Not ours: HA skin literals tokens.css:93, :102-104 (#009ac7 3.26, #db4437 4.29, #ffa600 1.96, #43a047 3.30) are Home Assistant's default palette, verbatim; left.
- ui-138: no aria-controls (no tabpanel in any consumer).
- ui-143(a): cost item; done without a red test.

## Guards (HEAD 7c77476)

- npx vitest run (ui/): 37 files, 476 tests passed.
- npm run check: 0 errors, 0 warnings (970 files).
- Lint: ui/package.json has no lint script.
- Not run (as instructed): make web, dist-check, make e2e-web. fabricui/dist is stale until the coordinator's rebuild.

## For the owner / coordinator

1. ui-172: the standalone shell uses data-skin="ha", i.e. HA's default palette, which fails AA as small text; keep, own standalone skin, or text-safe variants — owner decision.
2. ui-133 Operation: admin = node csrf_token/admin or controller csrf_token. ControllerStatus has no admin field, so a controller-only host in host mode shows the level read-only even to admins; ControllerStatus.admin would be a contract change (phase 3). The page reads the node status when mounted with neither status loaded.
3. New catalogue keys (W2): operation.paused_dropped, common.seconds_short, kit.tag_title, kit.tag_title_mfg, controller.window.close_hint/.expired/.left, controller.updates.queue_unknown, controller.settings.cap.diagnostics, controller.settings.dataset_invalid, controller.icd.forget_confirm/_body, common.locale. W1's Pair part of ui-141 can use common.seconds_short ("{n} s").
4. ui-182: common.locale is a catalogue's BCP 47 tag; a host catalogue for another language must name it, else dates follow the runtime locale.
5. ui-142: the search no longer takes regexes (only `|` as OR) — behaviour change vs the port, also for DataTable's search.
6. ui-166: Close is offered without a known window; the server's answer to a close with none open is untested (worst case a toast). A read route would be a contract change.
7. ui-138: route tab strips (shell family/page tabs, Nodes list/network) are nav links now, not role=tab; Network.test.ts adjusted; the Playwright smoke does not query tab roles (grep, not run).

## Follow-up: decision 31 (ui-172), ui-193, ui-196

Branch review/p1b-w2 rebased onto review/p1b (59f0204); the earlier W2
commits were already there (cherry-picked) and were dropped by the
rebase as applied. New commits since review/p1b's HEAD, oldest first:

- da26304 UR-P1-bound-134 test only: the 2000-row paused-backlog test hit the
  20 s default under a full parallel run (25.7 s); it gets 60 s, its waitFor 20 s.
- d6ad8db UR-P1-ui-172 (decision 31): status text tokens. Test:
  theme/tokens.test.ts "reach 4.5:1 on the card and on a badge's tint",
  standalone light/dark, HA skin light/dark (red: tokens absent; then
  standalone light failed on the 15% tint: primary 3.76, error 3.82,
  warning 4.08, success 4.43).
- 3c328f3 UR-P1-ui-193: kit/utils.ts formatJson → stringifyJson(v, 2) in
  AttributeTable and EventList. Test: kit/json-display.test.ts (2 cases).
- f3e1493 UR-P1-ui-196: Operation, mounted alone, reads the controller's
  status after a node status that yields nothing. Test: Operation.test.ts
  "learns the token from the controller's status when the node family answers 404".

### Decision 31: design
- New tokens --fui-{primary,error,warning,success,info}-text, for small
  status text only; --ha-*-color and the HA-remapped Tailwind ramps stay
  for areas, borders, icons.
- HA skin: color-mix of the host colour towards #000000 (light: primary 70%,
  error 80%, warning 50%, success 70%, info 65%) and towards #ffffff (dark:
  75%, error 70%). Why color-mix: the bridge copies --warning-color & co.
  live, so a custom HA theme's hue still drives the text without asking
  the host for any new variable; a literal fallback would ignore it.
- Override: each token reads --fui-<status>-text-color first, for a host
  whose colours the mix does not carry to AA (set on any ancestor).
- Standalone: light 80% towards black (the plain colours fail on a
  badge's tint); dark as they are.
- Checked against the card AND a badge's 15% tint, in all four palettes.
- Switched to the text tokens: every small-text use in W2's files (Badge,
  Button outline-destructive, Tabs active, ErrorState message, Graph link,
  and the page texts: 30 files). Icons keep --ha-*-color (Selection state
  icons, ErrorState icon). No Tailwind text-<status>-<step> utility is
  used anywhere in the sources (grep), so none needed switching.
- W1's files that need the same change (please apply,
  text-[var(--ha-X-color)] → text-[var(--fui-X-text)]):
  - ui/src/lib/pages/node/Pair.svelte:108 (load error text)
  - ui/src/lib/pages/controller/node/Attributes.svelte:126 (status text)
  - ui/src/shell/NodeCreateDialog.svelte:52 (error text)
  - Pair.svelte:204 is a 4xl decorative check mark (aria-hidden): an icon, leave.
- Not covered: white text on filled primary/error buttons (text on a fill,
  not on the card) — out of this decision's scope.

### ui-193 note
stringifyJson refuses an unsafe JS number; for display formatJson falls
back to JSON.stringify so a page never fails.

### Guards (HEAD f3e1493)
- npx vitest run: 50 files, 555 tests passed.
- npm run check: 0 errors, 0 warnings (985 files).
- No lint script; no make web.
