# Phase 1b — fix wave, agent W1

Branch `review/p1b`, base `869b99b`, HEAD `026bc33`
(17 commits, all local; nothing pushed). Files touched: W1's list only, plus
their tests, the i18n catalogues, and two new helpers in owned directories
(`ui/src/lib/api/json.ts`, `ui/src/shell/focus-trap.ts`). `contract/`,
`fabricui/`, `docs/` and `fabricui/dist` untouched; `make web` not run.

## Per finding

| ID | Result | Commit | Test (file › describe) |
| --- | --- | --- | --- |
| UR-P1-ctr-101 (+162) | fixed | `56ea1d5` | `api/json.test.ts` › parseJson (native and fallback), stringifyJson, "the client's generic values", "the stream's cluster_event"; `pages/controller/node/Attributes.test.ts` › "UR-P1-ctr-101: a 64-bit value read and written back" |
| UR-P1-conc-102 | fixed | `7b8fc84` (helper), `2aa1cef` (stores, log) | `api/stream-life.test.ts` › "UR-P1-conc-102 …"; `store-review.test.ts` › "UR-P1-conc-102: a store reloads after a gap in its stream" |
| UR-P1-bound-103 | fixed (decision 29) | `7b8fc84` (hidden tab), `ac77145` (on demand) | `stream-life.test.ts` › "UR-P1-bound-103: a hidden tab holds no stream"; `shell/App.test.ts` › "UR-P1-bound-103 …" |
| UR-P1-ui-104 | fixed | `f7156c8` | `shell/dialogs.test.ts` › "UR-P1-ui-104: a double click does not pass both steps", "… the confirm button is armed after the guard" |
| UR-P1-conc-105 (+160) | fixed, both sites | `5c8473f` | `scan/start.test.ts`, `pages/controller/QrScan.test.ts` › "UR-P1-conc-105 …" |
| UR-P1-err-106 | fixed | `c49693b` | `api/client-errors.test.ts` › "UR-P1-err-106: answers that are not JSON" |
| UR-P1-err-107 | fixed | `ac77145` | `shell/App.test.ts` › "UR-P1-err-107: a failed probe is repeated" |
| UR-P1-err-108 | fixed | `49d7772` | `client-errors.test.ts` › "UR-P1-err-108: a rotated token" |
| UR-P1-conc-109 | fixed | `a9b3be3` | `store-review.test.ts` › "UR-P1-conc-109 …" |
| UR-P1-conc-110 | fixed | `33b4d13` | `store-review.test.ts` › "UR-P1-conc-110 …" |
| UR-P1-bound-111 | fixed | `597a463` | `shell/toast.test.ts` › "UR-P1-bound-111 …" |
| UR-P1-ui-112 | fixed | `033724a` | `shell/App.test.ts` › "UR-P1-ui-112: an inherited name is no page" |
| UR-P1-ctr-113 | not fixed: phase 3 per the brief (contract) | — | — |
| UR-P1-web-114 | fixed | `b1295bc` | `kit/DataTable.test.ts` › "UR-P1-web-114 …" |
| UR-P1-web-115 | partly as proposed (see below) | `7b57f41` | `client-errors.test.ts` › "UR-P1-web-115 …"; `context-ref.test.ts` |
| UR-P1-web-116 | fixed | `7b8fc84` | `stream-life.test.ts` › "UR-P1-web-116 …" |
| UR-P1-ui-117 | fixed | `f7156c8` | `shell/dialogs.test.ts` › "UR-P1-ui-117 …" |
| UR-P1-ui-118 | fixed (bridge and prepaint) | `220af9d` | `theme/theme.test.ts` › "UR-P1-ui-118 …" |
| UR-P1-test-119 | done with the fixes: stream error/open fakes, non-JSON answers, inherited page names, `startScan`, `confirm`/`toast` tests now exist | — | the tests above |
| UR-P1-err-04 page part | fixed with the verifier's three corrections | `33b4d13` | `pages/node/Pair.test.ts` › "UR-P1-err-04: a window opened elsewhere" |
| UR-P1-ui-132 | fixed | `33b4d13` | `Pair.test.ts` › "UR-P1-ui-132 …" |
| UR-P1-ui-133 (Pair part) | fixed | `33b4d13` | `Pair.test.ts` › "UR-P1-ui-133: a viewer without admin rights" |
| UR-P1-ui-141 (Pair part) | fixed | `33b4d13` | `Pair.test.ts` › "UR-P1-ui-141 …" |
| UR-P1-ui-161 | fixed (verifier's variant: clear, reset cluster) | `449a447` | `Attributes.test.ts` › "UR-P1-ui-161 …" |
| verifier note `Attributes.svelte:46` | fixed | `449a447` | `Attributes.test.ts` › "a read the node answers with neither value nor status" |

Every test was run red against the code before its fix (the fix stashed or
the file restored from HEAD), then green. `026bc33` only renames an unused
parameter and drops a lint comment.

## How, where it differs from the proposals

- **101/162.** `lib/api/json.ts`: own small parser (no dependence on reviver
  `context.source`); an integer that `JSON.parse` would round is kept as
  `JSON.rawJSON(src)` where the engine has it (so the kit's
  `JSON.stringify` displays stay exact without touching W2's files), else a
  `JsonText` (fallback, tested). Kept only under `value`, `data`, `response`
  of the routes that carry generic values (read, invoke, endpoint and
  cluster detail, the two event routes, stream `cluster_event`), so typed
  numbers (`Counter.value`, `EventRecord.number`) stay numbers. Every body
  goes through `stringifyJson`: raw values verbatim, a JS number that is an
  unsafe integer refused with `RangeError` before sending (decision 27's
  refusal; covers `invokeCommand`). Attributes sends the edit field's text
  verbatim via `jsonText(edit)` after checking it is JSON. New exports from
  `./client`: `JsonText`, `isRawJson`, `jsonText`, `parseJson`,
  `stringifyJson` (npm surface, additive). In a webview without
  `JSON.rawJSON`, the kit's displays (W2's `AttributeTable`, `EventList`,
  `ClusterView`) show such an integer as a quoted string — exact, but
  quoted; the Attributes tab uses `stringifyJson` and shows it bare.
- **115.** `forNode` refuses a key outside the contract's NodeKey pattern;
  `refHref` links a controller node only by a 64-bit hex id. For the
  controller node routes only an empty id and the dot segments `.`/`..`
  are refused: the contract's `NodeID` has **no pattern**, and the fixture
  `controller-node.400` sends `xyz` and expects the server's 400 — a full
  hex check broke that fixture's client test. `../x` is encoded to one
  segment and stays the server's.
- **102/103.** Stream helper reopens on error in state CLOSED (backoff 1,
  2, 5, 10, 30 s), closes while hidden, `onResync` on every open but the
  first. Stores: `subscribe(onEvent, onResync?)` (a host's own function
  gets the second argument; old ones ignore it), `disconnect()`, `resync()`.
  Log stream: the default `subscribeLog` reads `logRecords({since})` after
  a gap (the page merges by seq). Shell: the default node's stream only
  while the node family shows the default node, the controller's only on
  the controller family.
- **104.** Confirm button disabled for 500 ms after a question shows, and
  `createConfirm` ignores an earlier `true`. Disabled (not only ignored) so
  the Playwright smoke's two-step clicks wait instead of being swallowed
  (`e2e/smoke.spec.ts:146-147`, `:207-208`). `make e2e-web` not run.
- **err-04.** Not a new phase: `Commissioning.foreignWindow` on phase
  `open`, so `window_closed`, the idle check and the page's branches need
  nothing new (correction 3). Status read fresh in `hydrateCommissioning`
  (correction 2); `basic`/`enhanced` tested literally; notice text says the
  codes "may not open it" (correction 1). `store.test.ts`'s hydrate test now
  answers `node-pairing.200-window`: `node-pairing.200` names a closed
  window, which rule 1 hydrates to nothing.
- **133 (Pair).** The store exposes `isAdmin` and `isAdminStatus()`
  (`store.svelte.ts`). W2 fixes 133 on the other pages; the coordinator may
  want them on this one helper.

## i18n keys added (en and de)

`controller.attributes.no_answer`, `shell.probe_failed`,
`common.seconds_short` (`{n} s`), `node.pair.foreign_window`,
`node.pair.admin_required`, `node.pair.window_open_left`,
`node.pair.window_open_now`. `common.seconds_short` is G-web A's proposed
name for 141, so W2 (Status.svelte) may add the same key: identical text
`{n} s` on both sides keeps the merge trivial.

## Guards (in `ui/`, at HEAD `026bc33`)

- `npx vitest run`: 45 files, 499 tests passed (baseline at `869b99b`: 33
  files, 433 tests).
- `npm run check` (svelte-check `--fail-on-warnings`): 0 errors, 0 warnings.
- ESLint: not configured in `ui/` (no config, no binary, no npm script);
  the repo's `make lint` is Go only. Not run.
- Not run, by the brief: `make web`, `make dist-check`, `make e2e-web`.

## For the owner or the coordinator

1. **101, docs:** the verifier's point 6 — `docs/contract.md:192-195`
   should say what a JS client does with `value` (lossless parse, verbatim
   write). Not a W1 file; left for the coordinator.
2. **115:** the contract's `NodeID` has no pattern; a pattern (64-bit hex)
   would be a contract change (phase 3) and would let the client refuse
   more. Today only dot segments are refused.
3. **104:** please run `make e2e-web` after the rebuild; the smoke's
   confirm clicks now wait 500 ms for the button to arm.
4. **113:** stays phase 3 (contract), as the brief says.

## Follow-up: pre-merge rows (`phase-1b-premerge.md`)

On HEAD `0a48bb5` (W2's commits, dist `9d0b45f`), one commit per row, each
test red before its fix and green after. No `make` run (a guard run may be
in progress in this worktree).

| ID | Commit | Test (file › describe) | How |
| --- | --- | --- | --- |
| UR-P1-conc-190 | `4b62df9` | `api/stream-life.test.ts` › "UR-P1-conc-190: a page loaded hidden resyncs on its first open" | `resyncDue` starts true when the document is hidden at subscribe |
| UR-P1-bound-195 | `14b387b` | `stream-life.test.ts` › "UR-P1-bound-195 …" | the helper coalesces `onResync`: at most one per `RESYNC_MIN_MS` (5 s), opens within it share one trailing call |
| UR-P1-conc-192 | `f2b1bc1` | `store-review.test.ts` › "UR-P1-conc-192: connecting again reloads on the new stream's open, not before it" | new `StreamOptions.resyncOnFirstOpen`; the stores' default subscription passes `connectedBefore` and no longer reloads in `connect()`; a host's own `subscribe` keeps the immediate reload (it has no open to wait for) |
| UR-P1-conc-191 | `59f0204` | `store-review.test.ts` › "UR-P1-conc-191: the log's resync without a streamed record, and after a restart" | the log resync reads the newest 500 ring records (no `since`); the page merges by seq. The earlier log test now expects `limit=500` |
| UR-P1-err-198 | `24e3626` | `store-review.test.ts` › "UR-P1-err-198: a window replaced under an open flow is judged afresh" | the store remembers the shown window's kind and deadline (own window: now + duration); `syncCommissioning` resets and re-hydrates when a fresh status names another window (kind differs or deadline moved > 5 s); `window_opened` clears the closed timer and, on an open flow, reloads the status and syncs |
| UR-P1-conc-199 | `b7a9a64` | `shell/App.test.ts` › "UR-P1-conc-199: one probe chain, none after the shell is gone" | `probeGen` in `probe()`, bumped by each probe and by `onDestroy`; a superseded chain schedules nothing |

`9366d66` types the test's fetch stand-in for svelte-check.

Guards in `ui/` at HEAD `9366d66`: `npm run check` 0 errors, 0 warnings.
`npx vitest run`: first full run 550/551 — the one failure is W2's
`Operation.test.ts` "keeps at most the newest KEEP_RECORDS while paused …
(UR-P1-bound-134)", a 20 s timeout under load average ~11; the file alone
passes (11/11, 31 s). Second full run: the same 550/551, the same test
(32 s, load average ~16 while another guard run was going). None of these
rows touch the Operation page or its test's path (`testUI` injects its own
`subscribeLog`); the test is CPU-bound (2000 records) and needs a quieter
machine or a longer per-test timeout (W2's file).

Note for 191: after a daemon restart the ring's new numbers (1, 2, …) are
handed to the page beside the old ones; how the page orders two sequences
is UR-P1-ui-143 (W2's page).
