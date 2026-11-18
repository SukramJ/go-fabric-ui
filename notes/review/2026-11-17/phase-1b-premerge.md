# Phase 1b — pre-merge read (admin/token and stream paths)

Fable, bounded read of `git diff 869b99b..HEAD` on `review/p1b` at `9d0b45f`
(api, the three stores, context, shell, Pair, operation, DataTable,
Attributes). **M** = measured (scratch probe outside the repo:
`…/scratchpad/pm/t.mts`, `u.mts`, run with jiti on Node 22.22.1, both with
`JSON.rawJSON` and with the `JsonText` fallback) or read at a line that
alone decides; **I** = inferred. Nothing in the repo changed but this file.
No vitest file was run; the guards' results are not part of this read.

**Verdict: no blocker.** Eight new rows, all low or note; 190 and 191 are
worth taking before the merge (two or three lines each), the rest can go to
phase 3.

## The four questions

### 1 Token — holds

- **Only on admin-marked calls** (M): `client.ts:315-316` reads the token
  only under `opts.admin`; the retry repeats the same call with the same
  options (`:343`). `refreshToken` reads `/node/status`, then
  `/controller/status`, without `admin` (`:360`), so it carries no token
  and cannot itself retry (`token` is undefined at `:342`).
- **Not in a URL, log, toast or storage** (M): the stream URLs are
  `base + path` (`:426, :435, :497`), `EventSource` is opened without
  `withCredentials` (`stream.ts:84`) and carries no header; the only
  `localStorage` writer is DataTable's sort; error texts are the problem's
  detail/title or `method path: status` (`:346, :348`), the RangeError names
  the integer or node id, never a header. The token lives in `auth.token`
  and in the stores' `status.csrf_token` (as before the diff).
- **Same origin** (I): the header goes to `base + path`; the shell's base is
  relative. `credentials: "same-origin"` governs cookies, not the header — a
  host that hands `createClient` a cross-origin base sends the token there.
  Unchanged by 1b; see UR-P1-web-197.
- **Bounded** (M): a 403 `forbidden` on a call that carried a token costs at
  most one shared status read (two when the node family answers no status)
  and one repeat (`retried`, and only when the token changed, `:368`). A 403
  for another reason (same token back) is not repeated. Stream reopen:
  1/2/5/10/30 s, the last repeated, only for `readyState === CLOSED`, no
  resync on a failure (`stream.ts:122-130`). Probe: same ladder
  (`App.svelte:199-200`). One unbounded edge is the browser's own reconnect,
  UR-P1-bound-195.
- Operation's and Pair's admin detection (`Operation.svelte:27`,
  `store.svelte.ts:78-80`) only hide controls; the server still decides.
  Fail-closed gap: UR-P1-ui-196.

### 2 `json.ts` — correct on hostile input; verbatim write is safe

- **Prototype** (M): `{"__proto__":{…}}` and `constructor.prototype` become
  own members (`:100`, `defineProperty`), `Object.prototype` untouched, the
  parsed object keeps `Object.prototype`; round-trips verbatim. Scoped and
  unscoped alike.
- **Numbers** (M): `1e400`/`-1e400` and every integer-valued number beyond
  2^53 (`1.5e300`, `9007199254740993.0`, uint64 max, int64 min) are kept as
  text inside a kept member; `-0` → `0`, `1.0` → `1`, `1E+2` → `100`,
  `0.1e-400` → `0`, as `JSON.parse`. Outside a kept member the result is
  `JSON.parse`'s (rounded), and `stringifyJson` then refuses it unsent
  (RangeError) — decision 27's refusal.
- **Grammar** (M): 21 malformed or edge inputs (leading zero, `1.`, `.5`,
  `+1`, `NaN`, trailing comma, raw control character in a string, `\x`, lone
  surrogate escape, NBSP as whitespace, U+2028 after the value, trailing
  garbage) accept and reject exactly as `JSON.parse`, alone and inside an
  array. Unicode escapes in keys (`"__proto__"`) decode through
  `JSON.parse` of the token. Duplicate keys: the last wins, as `JSON.parse`.
- **Injection** (M): `stringifyJson` writes verbatim only a `JsonText` or an
  engine `rawJSON` object (`isRawJson`, by `instanceof`/`JSON.isRawJSON`,
  not by shape): a server object `{"rawJSON":"1,\"x\":2","toJSON":"x"}` is
  serialised as an ordinary object. `jsonText` accepts only one complete
  JSON value: `1,"x":2`, `1}`, `1}, {"a":1`, `"a"//`, `[1]\n,2`,
  `1e5 ,"p":1` are all refused, so a typed value cannot add a member to the
  request body. It does forward duplicate keys inside the typed value as
  typed (`{"a":…,"a":2}`): the server's decoder decides, and it is the
  administrator's own text.
- Limits and cosmetics: UR-P1-bound-194, UR-P1-ui-193.

### 3 Streams — no leak, no double subscription; three lost-event windows

- **Leak / double** (M): `openStream` drops every callback of a replaced
  source (`es === src`), clears the timer and the visibility listener on
  close (`stream.ts:149-154`); the stores guard `connect()` with `unsub`
  and `disconnect()` nulls it; the shell opens the default node's stream
  only for `family === "node" && nodeKey === undefined`, a keyed node's is
  NodeScope's own store, closed in `onDestroy`. No path holds two sources
  for one store.
- **Lost events**: UR-P1-conc-190 (first open deferred by a hidden tab),
  UR-P1-conc-191 (log resync), UR-P1-conc-192 (reload before the open).

### 4 err-04 `foreignWindow` — right at the hydrate, not re-judged after it

At the hydrate (`store.svelte.ts:326`) the flag matches decision 28 for
every pair (M, against `handlers_pairing.go:58-110` and the contract's
`window_kind`): `own`+`configured` no notice; `basic`+`window` (this
handler's window, any tab) no notice; `basic`/`enhanced`+`configured`
notice (covers decision 30's mislabel either way); `none` or absent no
notice. Status is read fresh before the pairing; Close is offered on every
open card (`Pair.svelte:197`); a non-admin sees no codes. The flag is never
set on a window opened here (`:281`). The gap is the flow that stays open
while the window under it changes: UR-P1-err-198.

## New rows

| ID | Sev | Where | What | Fix |
| --- | --- | --- | --- | --- |
| UR-P1-conc-190 | low | `ui/src/lib/api/stream.ts:103, :142-148` | **M (read).** A page loaded in a hidden tab (opened in the background, session restore) loads at mount but opens no stream (`:148`); when shown, `start()` runs with `resyncDue === false` (the hidden branch sets it only when a source was open, `:138-141`), so the first open does not resync: the state is as old as the tab was hidden until the next event. Same for the log. | `let resyncDue = !!doc?.hidden;` (or set it in the shown branch at `:142`) plus a test beside "a hidden tab holds no stream". |
| UR-P1-conc-191 | low | `ui/src/lib/context.ts:157-171` | **M (read).** (a) `last` counts only streamed records: a gap before the first streamed record (`since === undefined`, `:165`) fetches nothing, though the page holds ring records. (b) After a daemon restart the numbering starts again (`log.ts:20-26`) but `last` stays at the old maximum (`Math.max`, `:159`): every resync asks `since` beyond the new numbers and gets nothing — the start-up records, the gap that matters, until the new process passes the old count. | On resync read `logRecords({ limit })` without `since` when `last` is undefined, and let a lower streamed `seq` set `last = rec.seq`; simplest: always read the newest ring page on resync and let `appendRecord` drop or restart. |
| UR-P1-conc-192 | low | `ui/src/lib/store.svelte.ts:446-452`, `controller-store.svelte.ts:236-242`, `stream.ts:116-121` | **M (read).** `connect()` after a `disconnect()` reloads at once, while the new EventSource is still connecting, and that subscription's first open does not resync: an event between the reload's snapshot and the server-side subscription is lost (same window as the very first connect, which predates 1b). The hidden-tab reopen is right (resync on open). | Let the store ask the helper to call `onResync` on the first open too when `connectedBefore`, instead of reloading in `connect()`. |
| UR-P1-ui-193 | note | `ui/src/lib/kit/AttributeTable.svelte:16`, `kit/EventList.svelte:21` | **M.** Both print with `JSON.stringify`: exact with `JSON.rawJSON`, but on an engine without it (the `JsonText` fallback) a kept integer shows quoted, `"12345678901234567890"`. Also: an engine `rawJSON` object has a null prototype — `String(v)` or `{v}` in a template throws TypeError; no page does that today. | `stringifyJson(v, 2)` in both; a line in `json.ts`'s header on the null prototype. |
| UR-P1-bound-194 | note | `ui/src/lib/api/json.ts:79-147, :52, :204` | **M.** On the slow path only: nesting beyond about 5 000–10 000 levels and a single string beyond 8–16 MB throw RangeError where `JSON.parse` succeeds; every caller catches (client → ApiError "not JSON", stream → event skipped, Attributes → invalid-JSON toast). 300 000 kept integers parse in 1–2 s. `stringifyJson` prints a sparse array as `[1,,2]` (`:204`, `map` skips holes); no caller builds one. `jsonText` trims with `String.trim` (BOM, NBSP), harmless: the trimmed text is what is sent. | None needed; `Array.from(v, …)` at `:204` if wanted. |
| UR-P1-bound-195 | note | `ui/src/lib/api/stream.ts:116-121`, the stores' `resync()` | **I.** The backoff covers only a source that failed for good. A stream that opens and drops (a proxy cutting idle connections, a server answering 200 and closing) is retried by the browser at its own pace (about 3 s, or the server's `retry:`), and every open runs the full reload (status, fabrics, node list, three epochs → the shown page's refetch). Bounded by the browser, not by the code. | Coalesce `onResync` (at most one per few seconds), in the helper. |
| UR-P1-ui-196 | note | `ui/src/lib/pages/operation/Operation.svelte:27, :78` | **M (read).** Fail-closed only. `ControllerStatus` has `csrf_token` but no `admin`: on a controller-only host that decides admin rights itself the level control is hidden from an administrator (contract gap, phase 3). And mounted alone on a controller-only host in token mode the page reads only `node.loadStatus()` (404) and never the controller's status, so the control stays hidden. | `:78`: fall back to `controller.loadStatus()`; the `admin` field to phase 3's wizard. |
| UR-P1-web-197 | note | `ui/src/lib/api/client.ts:298-324` | **I**, unchanged by 1b. The token header follows `baseUrl`; `credentials: "same-origin"` does not restrain it. A host passing a cross-origin base sends the token there (a preflight applies, so only to an origin that allows the header). | A sentence in `ui/README.md`/`docs/embedding.md`, or refuse the header when `new URL(base, location.href).origin !== location.origin`. Phase 3. |
| UR-P1-err-198 | low | `ui/src/lib/store.svelte.ts:374-380, :387-401, :436-443` | **M (read).** `foreignWindow` is judged only in the hydrate. While the flow is `open`: (a) `syncCommissioning` after a stream gap checks only `commissioning_window_open` — own window closed and another opened meanwhile keeps the earlier codes and countdown **without the notice** (against decision 28's "freshly read"); (b) `window_opened` during the 'closed' timer only reloads the status, then the timer sets the flow idle with the "closed" toast while a window is open, and nothing hydrates. Narrow (needs a replacement window within a gap or the timer). | In both places, when the flow is `open` and the fresh status's `window_kind`/`window_deadline` no longer match the shown window, reset to idle and hydrate; cancel `closedTimer` on `window_opened`. |
| UR-P1-conc-199 | note | `ui/src/shell/App.svelte:190-205, :246` | **M (read).** `retryProbe()` during a running probe runs two chains until one timer fires (the orphaned timer is not cleared), and a probe in flight at destroy schedules a timer after `onDestroy`. Self-healing; the shell is never destroyed in practice. | A generation counter in `probe()`. |

## Not checked

Browser behaviour of `EventSource` (reconnect pace, 204) and of fetch on a
cross-origin redirect is inferred, not run. The rebuilt bundle
(`fabricui/dist`) was not read. `ui/src/shell/App.test.ts` showed as modified for a moment during the read
(the background guards, not this read) and was clean again at the end;
`notes/…/decisions.md` is modified and uncommitted, also not by this read.
