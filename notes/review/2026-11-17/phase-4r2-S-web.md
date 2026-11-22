No held candidate: this report carries no high and no medium.

# Phase 4, run 2 — unit S-web (the web security paths)

Fable, 2026-11-19. Read at `792d15b` (worktree HEAD `89c098b`, notes only on
top). Brief `notes/briefs/2026-11-19-review-p4r2.md`. Scratch worktree
`<scratchpad>/p4r2wt/S-web` (detached at `792d15b`, `npm ci`), focused
vitest files only. M = measured (a test run or a line read on both sides),
I = inferred.

**Result: 0 high, 0 medium, 2 low, 4 note.** The security properties of the
unit hold (see "Covered without finding"); both lows are liveness defects
of the client, neither gives a role anything it must not have.

## Scope, listed first

Unit files (11): `ui/src/lib/api/client.ts` (701), `stream.ts` (452),
`json.ts` (213), `ui/src/lib/store.svelte.ts` (639),
`controller-store.svelte.ts` (347), `nodes-store.svelte.ts` (62),
`ui/src/shell/App.svelte` (357), `shell/route.ts` (54),
`ui/src/lib/theme/ha-bridge.ts` (186), `ui/src/lib/pages/node/Pair.svelte`
(247), `ui/src/lib/pages/controller/node/Window.svelte` (116).
The token path beyond them: `ui/src/lib/context.ts` (log subscription,
routes), `ui/src/shell/NodeScope.svelte`, `ui/src/shell/NodeFamily.svelte`,
`ui/src/lib/qr.ts` (what the two sinks render), `ui/src/lib/token-path.test.ts`,
`store-scoped-stream.test.ts`. Delta first: `client.ts`, `store.svelte.ts`,
`Pair.svelte` by `git diff 979a67a 792d15b`, then every file whole.
Go, only to judge the client: `fabricui/handler.go` (construction, admin
gate `:700-783`, reserved keys `:196-213`), `handlers_node.go:140-200`
(status), `problem.go` codes.

## Findings

### UR-P4-conc-230 — low — M — `hydrateCommissioning` overtakes a `loadStatus` in flight: the status stays "loading", and a new token does not reopen the node stream

- **Where:** `ui/src/lib/store.svelte.ts:364-366` (`hydrateCommissioning`
  bumps `statusGen` and sets the status itself), `:148-168` (`loadStatus`
  drops an overtaken answer — and with it `statusLoading = false` in its
  `finally` and `reopenOnNewToken` at `:157`), `:536-540`.
- **Claim:** `statusGen` is shared by two readers of which only one owns
  `statusLoading` and the reopen. When a hydrate starts while a
  `loadStatus` is in flight:
  1. `statusLoading` stays `true` until some later `loadStatus` ends. The
     shell does exactly this on a start at `#/node/pair`:
     `App.svelte:198` `void node.loadStatus()`, then the Pair page mounts
     and hydrates (`Pair.svelte:36-38`).
  2. If the hydrate's own read fails (a 502, a network error), the store
     ends with `status === null`, `statusLoading === true`,
     `statusError === null` although the overtaken read answered 200:
     `Selection.svelte:57` then waits for good, `Fabrics.svelte:37` and
     `Selection.svelte:48` never load the status again, `Status.svelte:54`
     shows the loading state.
  3. After a daemon restart (token mode, a new token per process) the
     node stream reopens with the stale token and is answered filtered.
     If the resync's `loadStatus` (`:502`) is overtaken by a hydrate — the
     Pair page mounted, or a `window_opened` on an idle flow (`:456`) —
     the client holds the new token but the stream is not reopened: it
     stays a non-admin's (status, `nodes_changed`, window events only)
     until the next `loadStatus`. A narrower rest of UR-P3-web-191 /
     `a548f8c`; `openWindow` (`:321-325`) sets a status without the
     reopen too, but does not drop another reader's.
- **Test:** `ui/src/lib/p4r2-sweb-store.test.ts` (scratch), three cases, all red:
  ```
  FAIL … statusLoading ends when both have answered
  AssertionError: expected true to be false
  FAIL … a failed hydrate read does not leave the store without a status and loading for good
  + "loading": true, + "status": false   (expected status true, loading false, error null)
  FAIL … the node stream is reopened when the overtaken status brought a new token
  AssertionError: expected [ 't1' ] to deeply equal [ 't1', 't2' ]
  ```
- **Proposed fix:** one private `readStatus()` used by `loadStatus`,
  `hydrateCommissioning` and `openWindow`: the newest answer sets the
  status and calls `reopenOnNewToken`; `statusLoading` counts the reads in
  flight (false when none is), so a dropped answer cannot leave it set; a
  hydrate whose read fails does not discard an older read's good answer.
- **Contract:** no.

### UR-P4-err-231 — low — M (the hang), I (reach) — the token reload waits on itself when the status route is refused 403 `forbidden` too

- **Where:** `ui/src/lib/api/client.ts:495` (`call` awaits
  `refreshToken` on a 403 `forbidden` that carried a token), `:506-524`
  (`refreshToken` reads the status through the same `call`, without
  `retried`).
- **Claim:** the refresh's own `GET /node/status` (and `/controller/status`)
  carries the stale token. If it is answered 403 with a problem of code
  `forbidden`, the inner `call` awaits `refreshing` — the promise that is
  waiting for this very call. It never settles; `refreshing` is never
  reset, so every later call refused the same way hangs with it: pages stay
  in their loading state with no error, a stream's refresh (`:403`) never
  ends. The shipped handler does not answer its status routes with
  `forbidden` (`fabricui/handlers_node.go:149`, an open route; the Host and
  Origin refusals have their own codes), so the precondition is a layer in
  front that answers in the contract's problem shape — a host's own gate
  or `fetchImpl` in token mode. Phase 1b's pre-merge read held "cannot
  itself retry" on the ground that the refresh carried no token
  (`phase-1b-premerge.md:19-22`); since 0.10.0 every request carries it
  (UR-P3-ctr-07), and the ground is gone.
- **Test:** `ui/src/lib/api/p4r2-sweb-client.test.ts` (scratch), red:
  ```
  FAIL … a call refused 403 forbidden settles when the status is refused 403 forbidden as well
  Expected: "ApiError 403"   Received: "never settled"
  ```
- **Proposed fix:** the refresh's two reads go out with `retried: true`
  (they fail as an `ApiError`, the loop takes the next source, the caller
  gets its 403).
- **Contract:** no.

### UR-P4-ui-232 — note — M — the Operation tab shows when `GET /operation/info` did not answer

- **Where:** `ui/src/shell/App.svelte:229-236`; the file's header `:9`
  ("its tab shows when GET /operation/info answers") and `:75`.
- **Claim:** `operationServed = !(ApiError with 403 | 404 | 503)`: a
  network error (no `ApiError`) and any other status (a proxy's 502, a 500)
  set it `true`. The probe runs once and is not repeated, so the tab then
  leads to a page whose every read fails; a transient failure at start on a
  host that does serve operation happens to come out right, one on a host
  without it does not.
- **Test:** `ui/src/shell/p4r2-sweb-app.test.ts` (scratch), two cases red:
  `expected [ <a href="#/operation" …> ] to deeply equal []` (a network
  error; a 502).
- **Proposed fix:** `true` only on an answer; an undecided answer repeats
  with the families' backoff (`probe`), as UR-P1-err-107 did for them.
- **Contract:** no.

### UR-P4-ui-233 — note — M (read) — a refused stream ends without a word while the status still says admin

- **Where:** `ui/src/lib/store.svelte.ts:561`, `:591` (`streamDenied` of
  the node store: no reader in `ui/src`), `ui/src/shell/App.svelte:333-335`
  (the controller's is shown only under `!controller.isAdmin`),
  `ui/src/lib/context.ts:165` (the log subscription passes no `onDenied`),
  `ui/src/lib/api/stream.ts:312-317`.
- **Claim:** a 403 on a stream ends the subscription for good (decided
  with UR-P3-web-191: "a second 403 … ends as today"). In host mode
  (`Options.Admin`, no token to refresh) a session that ends while the
  controller family is shown gets 403 `forbidden` on the stream's next
  reopen; the status in the store still says admin, so the pages stay and
  show no event (commissioning progress among them) and nothing says why;
  the stream is not reopened when the session is back until the family is
  left and entered again. No right is gained. Not the decided shape's
  fault, its missing second half: the store could reload its status on
  `onDenied`, which turns the page to the notice it already has.
- **Contract:** no.

### UR-P4-ui-234 — note — I — the controller's Window tab throws on a payload the QR encoder refuses

- **Where:** `ui/src/lib/pages/controller/node/Window.svelte:49`
  (`win ? renderQrSvg(win.qr_code) : null`), against
  `ui/src/lib/pages/node/Pair.svelte:128` (`win?.qr_code ? … : null`);
  `ui/src/lib/qr.ts:48`.
- **Claim:** `qrcode`'s `create` throws on an empty text ("No input
  text", M: run in the scratch worktree) and on one over the code's
  capacity (2400 characters at level M, M). `NodeWindow.qr_code` is a
  required string without a length (`contract/openapi.yaml:3327-3330`); an
  empty one throws inside the `$derived` and takes the tab down instead of
  showing the manual code. Pair guards the empty case, not the long one.
  go-fabric always answers a payload, so this needs a broken source; the
  sink itself stays safe (the SVG is built from numbers only).
- **Proposed fix:** a `try` in `renderQrSvg`'s two callers (or in it,
  answering `null`), the codes shown without the picture.
- **Contract:** no.

### UR-P4-web-235 — note — I — `renameNode` and `deleteNode` take a key without the pattern `forNode` checks

- **Where:** `ui/src/lib/api/client.ts:582-583` against `:527-530`
  (`forNode`) and `:462-466` (the controller id and endpoint checks,
  UR-P1-web-115).
- **Claim:** `encodeURIComponent` leaves `.` and `..`, which the URL
  parser resolves: `deleteNode("..")` is `DELETE <base>/`, `renameNode("..", l)`
  is `PUT <base>/label`. Neither is a route (405, 404), and the pages pass
  keys of `GET /nodes` only, so nothing is reachable today; the same
  check as `forNode`'s keeps it so for a host's own caller.
- **Contract:** no.

## Covered without finding

Every file below was read whole.

- **`api/client.ts`** — the token goes into a header only, never a URL,
  storage or log; only to a base or stream URL of the page's origin,
  resolved as fetch resolves it (`<base href>` included, UR-P4-web-04's fix
  read and its test); `X-Fabric-UI: 1` on every request and stream;
  `credentials: "same-origin"`; the 403 retry runs once per call and only
  on a changed token (a refused write was not executed, so its repeat is
  safe); `statusWithToken` repeats once; path building: node key by
  pattern, controller id by pattern before sending, endpoint numeric and
  bounded, every other segment and query through `encodeURIComponent` /
  `URLSearchParams`; the `/endpoints/` test cannot hit a node key, because
  the Go side refuses a key that is a page name (`handler.go:196-213`);
  problems are read by shape and rendered as text; a non-JSON answer
  keeps its status. The stream wrapper: one reopen, listeners carried
  over, nothing dispatched after `close()`.
- **`api/stream.ts`** — `FetchEventSource`: non-200 or another media type
  is an error with status, `Retry-After` and code; the line parser (CR, LF,
  CRLF split over chunks, comments, multi-line data, `id` with NUL,
  a pending frame dropped at the end) as EventSource; `openStream`: 403
  ends, 503 honours `Retry-After`, backoff, hidden tab, the resync
  throttle, a stale source's events ignored (`es === src`); data that is
  not JSON is skipped.
- **`api/json.ts`** — the slow path against `JSON.parse` (numbers, strings,
  literals, trailing input, `__proto__` as an own property, duplicate
  keys); the regular expressions are linear; `jsonText` validates before
  it is sent verbatim; `stringifyJson` refuses an unsafe integer. The
  depth and size limits are UR-P1-bound-194 (struck), not re-reported.
- **`store.svelte.ts`** — beyond 230: codes dropped on success, close and
  expiry (UR-P2-leak-08); no pairing read without admin rights; the
  generation counters of fabrics and exposure; `reopenOnNewToken` cannot
  loop with the shipped client (`statusWithToken` sets the token before
  the store compares; UR-P4-web-195's fix read). UR-P4-ui-61's other-tab
  path is carried, not re-reported.
- **`controller-store.svelte.ts`** — no stream without admin rights; a
  foreign commissioning's progress is not shown (decision 59); the
  commissioning id matches the contract's pattern in both branches.
- **`nodes-store.svelte.ts`** — generation, a failed list is empty.
- **`shell/App.svelte`, `route.ts`, `NodeScope.svelte`, `NodeFamily.svelte`**
  — a hash cannot name a node the list does not hold, a page or tab outside
  the own-property tables, or a controller id outside the pattern; the
  admin streams (controller, log) open only after a status brought the
  token (`controllerFamily`, `probed`); the status route answers 200 on
  every host shape (`handlers_node.go:146-193`), so token mode always has a
  token source; a denied probe is a notice and is not repeated.
- **`theme/ha-bridge.ts`** — reads a same-origin parent only, inside
  `try`; copies a fixed list of custom properties; a cross-origin parent
  is inert.
- **`Pair.svelte`, `Window.svelte`, `qr.ts`** — the two `{@html}` render a
  string built from module coordinates (numbers) only; payload, manual
  code and labels are text nodes; the decision 73 notice and the viewer's
  clock (the delta) read; no code is stored, logged or put in a URL.
- **`context.ts`** — the log stream through the client's `openStream`;
  `refHref` admits ids by pattern only.

## Not covered

- `ui/src/lib/api/schema.gen.ts` (generated), `api/index.ts`, `lib/index.ts`
  (re-exports): not read.
- The kit's and the other pages' use of the stores, `Operation.svelte`
  beyond `:96-110`, `shell/main.ts`, `index.html`, the dialogs, the toaster:
  unit W. Read of them: only the lines cited above.
- The unit's test files other than `token-path.test.ts` and
  `store-scoped-stream.test.ts` (`client*.test.ts`, `stream*.test.ts`,
  `json.test.ts`, `store*.test.ts`, `App.test.ts` beyond its scaffolding,
  `route.test.ts`, `theme.test.ts`, `Pair.test.ts`): not read as guards
  (unit W's "web tests that hold a guard").
- The Go side beyond the lines named under "Scope": unit S-go.
- No browser run: the findings and the holds are from vitest (happy-dom)
  and reading; redirect handling of `fetch` with the token header was
  reasoned, not measured (the handler answers no cross-origin redirect).
- The pins `gf:`, `loom:`, `mjs:` were not needed and not opened; the
  port comments' loom citations in the unit's files were not re-checked.

## Tokens

About 200 k (estimate; the harness reports the exact number to the coordinator).
