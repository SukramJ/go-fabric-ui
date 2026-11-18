# Phase 1b — S-web (Fable)

HEAD read `007ca96`, branch `review/p1b`. Unit:
`ui/src/lib/api/{client,stream,index}.ts`, `context.ts`, the three stores,
`qr.ts`, `scan/index.ts`, `i18n/index.ts`, `theme/ha-bridge.ts`,
`ui/src/shell/*`, and in pages and kit only the sinks and the storage of the
inventory's sections 1 and 5. Paths below are under `ui/src/` unless they
start with `fabricui/`, `contract/` or `docs/`.

**M** = read at the line or run. **I** = follows from measured lines plus
browser behaviour not executed here. Run: one scratch vitest file outside
the repo (`…/scratchpad/sweb/src/probe.test.ts`, 3 tests) against
`routeOf`, `createClient`, `createTranslator`, `parseRgb`; nothing in the
repo was changed.

Counts: 0 high, 1 medium, 13 low, 5 note.

## What holds (no finding)

- **Token.** Lives in the `auth` closure (`lib/api/client.ts:291`) and, as
  `csrf_token`, in the status objects of the stores; never in a URL
  (`streamUrl`, `logStreamUrl`, `controllerStreamUrl` `:365`, `:374`,
  `:433` carry none), never in `ApiError.message` (`:308`), a toast, a log
  line or storage (grep: the only `localStorage` use is
  `lib/kit/DataTable.svelte:62`, `:80`). Sent only on calls marked `admin`
  (`:296`); the marks agree with `fabricui/handler.go:301-326`, `:455`, `:460` and
  `fabricui/handlers_controller.go:55-78`, `:83-85` route by route (M).
- **Render sinks.** The two `{@html}` take `renderQrSvg`'s string, which
  interpolates numbers only (`lib/qr.ts:53-56`; `size` is never passed by a
  caller). Every dynamic `href` is a `routes.*` constant plus an id; the
  external ones are constants with `rel="noopener noreferrer"`. Toasts,
  confirm title and body, the create dialog's error are text
  interpolation (`shell/Toaster.svelte:16`, `shell/ConfirmDialog.svelte:21`,
  `:23`, `shell/NodeCreateDialog.svelte:50`). `format` substitutes in one
  pass, so a value holding `{name}` is not expanded again
  (`lib/i18n/index.ts:19-23`).
- **Route.** A node key from the hash is used only when `GET /nodes` lists
  it (`shell/App.svelte:102-104`); a controller node id is pattern-checked
  (`shell/route.ts:28`); `tab` is checked against a constant list
  (`lib/pages/controller/Node.svelte:35`).
- **Theme bridge and prepaint.** Constant name list, `setProperty` on
  custom properties only, cross-origin parent caught
  (`lib/theme/ha-bridge.ts:108-134`, `shell/index.html:54-86`). Nothing
  beyond the threat model's 2.5.
- **Stream ownership.** One subscription per store (`unsub` guard,
  `lib/store.svelte.ts:363`, `lib/controller-store.svelte.ts:207`); a node
  switch destroys `NodeScope`, which closes its store and timers
  (`shell/NodeScope.svelte:35`, `lib/store.svelte.ts:368-375`). A late
  response after a switch lands in the discarded store.

## Findings

### UR-P1-ctr-101 — a 64-bit attribute value is rounded on read and altered on write

Severity **medium**.

| | |
| --- | --- |
| Where | `lib/api/client.ts:305` (`JSON.parse` of every answer), `:300` (`JSON.stringify` of every body); `lib/pages/controller/node/Attributes.svelte:47` (`edit = JSON.stringify(result.value)`), `:59` (`JSON.parse(edit)`), `:67` (write), `:92` (display); `contract/openapi.yaml:3063`, `:2981` (`value` is raw JSON) |
| Claim | The rule "a JS number never carries a node or fabric id" (`docs/contract.md:193-196`, CLAUDE.md) holds for the named fields but not for generic attribute, event and command values. They arrive as JSON numbers at full 64-bit width (go-fabric renders an unsigned integer as `json.Number`, seen in `internal/valuejson/valuejson.go:184` at `v0.22.2`; the citation is phase 3's) and are parsed to doubles. The Attributes tab shows the rounded value, pre-fills the edit field with it and writes it back: a read-then-write of the ACL (`0x001F/0x0000`, subjects), a Binding list (node), or any struct holding a node, fabric or group-key id above 2^53 writes a different id than the node held, with a success toast |
| M / I | M for the page path and the wire form; I for a concrete node (any fabric whose commissioner assigns node ids above 2^53) |
| Fix | In the client: parse `value` losslessly (integers outside the safe range kept as their source text, e.g. a `BigInt`-aware parse of the body or a reviver using the source text) and send the edit field's text as written. Or on the wire: integers above 2^53 as strings |
| Contract | No for the client-side fix; yes if the wire form changes (decision 13) |
| Red test | vitest: a fixture answer `{"values":[{…,"value":{"subjects":[18446744073709551615]}}]}`; read, then write unchanged; the sent body must contain `18446744073709551615`. Today it contains `18446744073709552000` |

### UR-P1-conc-102 — a stream that fails once is dead for good, and a reconnect resyncs nothing

Severity **low**.

| | |
| --- | --- |
| Where | `lib/api/stream.ts:55-58` (`EventSourceLike` has no error or open hook), `:70-84`, `:114-128`, `:138-150`; `lib/store.svelte.ts:362-366`, `lib/controller-store.svelte.ts:206-210` |
| Claim | (1) An `EventSource` whose request is answered with anything but 200 `text/event-stream` (a proxy's 502 or 503 while the daemon restarts, a 404) does not reconnect; nothing listens for `error`, `unsub` stays set, so the store believes it is connected and the page stays stale without a sign until it is reloaded. (2) After a reconnect the browser does by itself, and after the server dropped events for a full client channel (`fabricui/stream.go:63-68`, the `default` branch), no store reloads: status, fabrics, node list and the commissioning flow keep what they had |
| M / I | M for the missing hooks; I for the browser's "fail the connection" on a non-200 answer |
| Fix | Add `onopen`/`onerror` to the helper: on `error` with `readyState === CLOSED` reopen with backoff; on every `open` after the first call a `resync` the store passes (status, fabrics, lists) |
| Contract | No |
| Red test | vitest with a fake source that emits `error` in state closed: the helper must open a second source; on the second `open` the node store must call `client.status` again |

### UR-P1-bound-103 — two or three tabs exhaust the browser's connections to the listener

Severity **low**.

| | |
| --- | --- |
| Where | `shell/App.svelte:192-197` (node and controller stream open as soon as each family is there, closed only in `onDestroy` `:199-205`), `shell/NodeScope.svelte:33` (a keyed node's stream beside the default node's), `lib/context.ts:147` (log stream on the Operation page) |
| Claim | A tab holds two streams at rest and three on a keyed node or on the Operation page, whatever page is shown; the default node's stream stays open while a keyed node is viewed. Over HTTP/1.1 (the Go listener without TLS, an Ingress proxy) a browser allows six connections per origin across all tabs: two tabs on a keyed node, or three tabs at rest, leave none for `fetch`, and every page of the origin hangs on its next call |
| M / I | M for the counts; I for the limit |
| Fix | One stream per tab (the families multiplexed, the node by parameter), or close the streams of a hidden tab (`visibilitychange`) and resync on return (with 102) |
| Contract | Yes for a multiplexed stream; no for closing on hidden |
| Red test | Playwright: three pages on the same server, then a `GET …/status` from the third must answer within a second |

### UR-P1-ui-104 — a double click passes both steps of a two-step confirmation

Severity **low**.

| | |
| --- | --- |
| Where | `shell/confirm.svelte.ts:7-15`, `shell/ConfirmDialog.svelte:25-29`; the two-step callers `lib/pages/node/Fabrics.svelte:115-128` (erase), `lib/pages/node/NodeManage.svelte:62-75` (delete node) |
| Claim | The second question replaces the first in the same dialog in the same task turn: `answer(true)` resolves the first promise, the caller's continuation calls `ask` again, and the confirm button of the second question is rendered where the first one's was (right-aligned, same row). The second click of a double click lands on it. The second step exists to stop exactly an unintended confirmation |
| M / I | M for the sequence; I for the timing in a browser |
| Fix | Ignore clicks for a short time after `current` changes, or make the second step a typed confirmation (the node's label), or swap the button order in the second step |
| Contract | No |
| Red test | Playwright: `dblclick` on the first dialog's confirm button; `POST …/maintenance/erase` must not be sent |

### UR-P1-conc-105 — the camera is not released when a scan fails to start or is cancelled while starting

Severity **low**.

| | |
| --- | --- |
| Where | `lib/scan/index.ts:49-53` (stream acquired, then `video.play()` and `nativeDetector()` awaited before `stop` exists, `:57`); `lib/pages/controller/QrScan.svelte:31`, `:45-51` |
| Claim | (1) If `video.play()` rejects after `getUserMedia` resolved, `startScan` throws and nobody holds the stream: its tracks are never stopped. (2) Cancel is offered while starting (`QrScan.svelte:61-62`); it runs with `stop === null`, sets `open = false`, and when `startScan` resolves afterwards `stop` is assigned to a scan nobody shows — the camera stays on and the 200 ms loop runs until the component is destroyed. If the component was destroyed meanwhile (`onDestroy(cancel)` ran earlier), never; a code read later is delivered to the destroyed page |
| M / I | M (read) |
| Fix | In `startScan`: `try { await video.play(); … } catch (e) { stop tracks; throw e }`. In `QrScan`: a `cancelled` flag checked after the await, stopping at once |
| Contract | No |
| Red test | vitest with a fake `mediaDevices` and a `play` that rejects: every track's `stop` must have been called. Second: `cancel()` before `startScan` resolves, then resolve: `stop` called |

### UR-P1-err-106 — an answer that is not JSON, or empty, is not an `ApiError`

Severity **low**.

| | |
| --- | --- |
| Where | `lib/api/client.ts:303-310`, `:335-336`, `:377-378` |
| Claim | `JSON.parse` runs before `res.ok` is looked at and outside any guard. A proxy's HTML error page throws `SyntaxError` (run: `SyntaxError: Unexpected token '<', "<html>502</html>" is not valid JSON`, not an `ApiError`), so the status is lost: the shell's 404 and 503 tests (`shell/App.svelte:159`, `:170`, `lib/controller-store.svelte.ts:81`) do not match and the operator reads a parser message. A 200 with an empty body returns `undefined` and `status()` throws `TypeError` on `s.csrf_token` (run). `isProblem` (`:454-456`) accepts any object with `code` and `title`; `detail` of another type becomes the message unchanged |
| M / I | M (run) |
| Fix | Parse inside `try`; on failure throw `ApiError(res.status, "<method> <path>: <status>")`; a 2xx without a body where one is required is an `ApiError` too; `isProblem` checks the three strings |
| Contract | No |
| Red test | vitest: a fake fetch answering 502 `text/html` must reject with `ApiError` of status 502 |

### UR-P1-err-107 — the shell probes each family once; a failed controller probe is never repeated

Severity **low**.

| | |
| --- | --- |
| Where | `lib/controller-store.svelte.ts:75-88` (any error but a 404 leaves `present` at `null`), `shell/App.svelte:186-188` (probes run once on mount), `:78` (`probed`), `:239-240` |
| Claim | When the first `GET /controller/status` fails for any reason but 404 (the daemon still starting behind a proxy, a network error), `present` stays `null` for the page's lifetime: the controller family never appears, and on a controller-only host (`nodeServed === false`) the shell shows "Loading…" for ever, since `probed` never turns true and nothing retries. `statusError` is set but the shell does not show it |
| M / I | M (read) |
| Fix | Retry the probes with backoff while one is undecided, and show an error state with a retry instead of the loading state when a probe failed |
| Contract | No |
| Red test | vitest on `App`: controller status answers 502 once, then 200; the controller tab must appear |

### UR-P1-err-108 — a rotated token is not picked up on a 403

Severity **low**.

| | |
| --- | --- |
| Where | `lib/api/client.ts:296`, `:306-309` (no handling of 403), `:336`, `:378` (the token changes only when a status call happens to run) |
| Claim | The handler draws a new token per process (`fabricui/handler.go:284`). After a restart the page keeps the old one; every admin call answers 403 "Admin rights required" until some event or page mount reloads a status. A reconnected stream triggers no reload (102), so on a quiet daemon the admin controls stay shown and fail until the page is reloaded |
| M / I | M for the lines; I for the sequence |
| Fix | In `call`: on a 403 `forbidden` of an admin call with a token, reload the status once and repeat the call once if the token changed |
| Contract | No |
| Red test | vitest: fake fetch with token `a`, then the server switches to `b` and answers 403 for `a`; `closeWindow()` must succeed after one status call |

### UR-P1-conc-109 — an older answer overwrites a newer one in the node and controller stores

Severity **low**.

| | |
| --- | --- |
| Where | `lib/store.svelte.ts:109-120`, `:122-132`, `:134-146`; `lib/controller-store.svelte.ts:75-110`; against the guard `lib/nodes-store.svelte.ts:11`, `:15-22` |
| Claim | `loadStatus`, `loadFabrics`, `loadExposure`, `loadNodes` carry no generation. Stream events start several at once (`fabric_added` and `window_closed` each call `loadStatus`, `:303-304`, `:330`, plus the health timer); answers on separate connections may arrive out of order, and the older one is stored last — e.g. `commissioning_window_open: true` after the close, until the next event. `statusLoading` is cleared by whichever returns first. The node list has the guard; these do not |
| M / I | M for the missing guard; I for the reordering |
| Fix | The generation counter of `nodes-store` in each loader |
| Contract | No |
| Red test | vitest: two `loadStatus` calls whose promises resolve in reverse order; the store must hold the second call's answer |

### UR-P1-conc-110 — `hydrateCommissioning` writes the flow without looking again after its awaits

Severity **low**.

| | |
| --- | --- |
| Where | `lib/store.svelte.ts:254-280` (phase checked at `:255`, state written at `:267-279` after two awaits) |
| Claim | While `client.status()` or `client.pairing()` is in flight, a `fabric_added` (phase `success`, `:302`), a `window_closed` or the user's `closeWindow` can change the flow; the write at `:267` then puts it back to `open` with the codes and starts a countdown for a window that is closed or already used |
| M / I | M for the lines; I for the interleaving |
| Fix | Re-check `commissioning.phase === "idle"` (or a generation) after the last await |
| Contract | No |
| Red test | vitest: start `hydrateCommissioning` with a deferred `pairing()`, apply `fabric_added`, resolve; the phase must stay `success` |

### UR-P1-bound-111 — error toasts never leave and have no cap

Severity **low**.

| | |
| --- | --- |
| Where | `shell/toast.svelte.ts:12-16`; sources driven by the stream: `lib/controller-store.svelte.ts:115` (every `update` event whose reload fails), `:183` (`update_error`, free text of any length) |
| Claim | An error toast is removed only by a click. Each `update` event while `GET /controller/updates` fails adds one, each failed update one more with the error text unbounded; the list and the DOM grow without limit and cover the page's lower right |
| M / I | M (read) |
| Fix | Cap the list (drop the oldest beyond ~5), collapse an identical message, clamp the text |
| Contract | No |
| Red test | vitest: 50 `update` events with a failing `updates()`; at most the cap of toasts |

### UR-P1-ui-112 — a crafted link selects an inherited property as the page component

Severity **low**.

| | |
| --- | --- |
| Where | `shell/route.ts:31`, `:52` (the page is any path segment); `shell/App.svelte:152` (`route.page in controllerPages`), `:207`, `:265`; `shell/NodeFamily.svelte:22-23`, `:33` (`route.page in pages`) |
| Claim | `in` is true for inherited names. `#/controller/__proto__` gives `controllerPage = "__proto__"` and `ControllerCurrent = Object.prototype`; `#/controller/constructor` gives `Object`; `#/node/<default key>/__proto__` the same in `NodeFamily` (run: `routeOf` returns these pages and `"__proto__" in {…}` is true). Rendering a non-function as a component throws inside the render effect: the link opens a blank or broken page instead of the first page. No script runs; the value is never data |
| M / I | M for the parse and the lookup (run); I for the render error |
| Fix | `Object.hasOwn(pages, route.page)` at the three places |
| Contract | No |
| Red test | vitest on `NodeFamily` and `App` with `#/controller/__proto__`: the nodes page renders |

### UR-P1-ctr-113 — the event number is a 64-bit counter in a JS number

Severity **low**.

| | |
| --- | --- |
| Where | `contract/openapi.yaml:2995` (`number: integer`), `fabricui/controller.go:523` (`uint64`), `lib/api/client.ts:86` (`since?: number`), `:443`; users `lib/kit/EventList.svelte:27` (each key), `lib/pages/controller/node/Events.svelte:34`, `lib/pages/node/Endpoint.svelte:60`, `:102` |
| Claim | A node that reports event numbers above 2^53 (its choice; the controller passes them on) gets them rounded: neighbours collapse to one number, which is the keyed-each key, the de-duplication test and the `since` cursor — duplicate keys in the list and a cursor that skips or repeats |
| M / I | M for the types; I for the effect in the list |
| Fix | `number_hex` (or a decimal string) beside `number`, and `since` as a string; the page keys and pages by it |
| Contract | Yes (additive) |
| Red test | vitest: two records `9007199254740993` and `9007199254740994` render two rows and `since` of the next call equals the last number exactly |

### UR-P1-web-114 — the table preferences are read from storage without a shape check

Severity **low**.

| | |
| --- | --- |
| Where | `lib/kit/DataTable.svelte:62-63` (`JSON.parse(raw) as Partial<Persisted>` spread into the state), `:130-131` (`query.trim()`, `makeTextMatcher(query)`), `lib/kit/utils.ts:17-24` (`new RegExp`) |
| Claim | `localStorage` is shared by everything of the origin (under an Ingress: the host's frontend and every other panel). A stored `{"query":5}` makes `query.trim` throw in the derived on every render of that table; a stored catastrophic pattern is compiled and run against every cell on every load. Both persist across reloads until the key is removed by hand. The search text itself is also kept in storage across sessions, which the header comment does not say |
| M / I | M (read); who writes the key is the threat model's "other content of the host's origin" |
| Fix | Take only `typeof` checked fields, a length cap on `query`, `sortKey` only when a column has it; do not persist `query` |
| Contract | No |
| Red test | vitest: `localStorage["fabric-ui.table:t1"] = '{"query":5,"sortKey":{}}'`, render: the table shows its rows |

### UR-P1-web-115 — ids are not validated where the library takes them

Severity **note**.

| | |
| --- | --- |
| Where | `lib/api/client.ts:326`, `:451` (`encodeURIComponent` leaves `.` and `..`); `lib/context.ts:94` (`routes.controllerNode + ref.id`, no pattern, while `:95` checks the endpoint) |
| Claim | `forNode("..")` requests `/api/v1/nodes/../status`, which the browser sends as `/api/v1/status`; `controllerNode("..")` as `/api/v1/controller/` (run). The shell cannot reach either (`shell/App.svelte:102-104`, `shell/route.ts:28`; the server's key pattern starts with a letter or digit, `fabricui/handler.go:162`), and no existing route is hit by one dot segment. A host with its own router and path-style `routes` passes what its router gives; `refHref` then builds a link from an unchecked `ref.id` |
| M / I | M (run) |
| Fix | Refuse a key or id that does not match the contract's patterns in `forNode`, `node()` and `refHref` |
| Contract | No |
| Red test | vitest: `forNode("..")` throws; `refHref({page:"controller_node", id:"../x"})` is `undefined` |

### UR-P1-web-116 — the streams ask for credentials across origins, `fetch` does not

Severity **note**.

| | |
| --- | --- |
| Where | `lib/api/stream.ts:70`, `:114`, `:138` (`withCredentials: true`); `lib/api/client.ts:301` (`credentials: "same-origin"`) |
| Claim | No effect in the shell (same origin). A host that gives the client a base URL of another origin gets cookies sent on the three streams and not on the calls; the two halves should say the same |
| M / I | M |
| Fix | Drop `withCredentials`, or derive both from one option |
| Contract | No |
| Red test | vitest: the default factory is called with `withCredentials` unset |

### UR-P1-ui-117 — the shell's dialogs do not take the focus

Severity **note**.

| | |
| --- | --- |
| Where | `shell/ConfirmDialog.svelte:13-33`, `shell/NodeCreateDialog.svelte:35-58` |
| Claim | `aria-modal="true"` without a focus move (confirm), a focus trap, an Escape handler or an inert background. After a keyboard-triggered destructive action the focus stays on the trigger under the overlay: Enter asks again (`confirm.svelte.ts:9` answers the open question with false), Tab reaches the page's controls behind the dialog |
| M / I | M (read) |
| Fix | A native `<dialog>` with `showModal()`, or focus the cancel button on open, trap Tab, close on Escape |
| Contract | No |
| Red test | vitest: after `ask`, `document.activeElement` is inside the dialog; Escape resolves false |

### UR-P1-ui-118 — the bridge reads light or dark only from hex and `rgb()`

Severity **note**.

| | |
| --- | --- |
| Where | `lib/theme/ha-bridge.ts:77-93`, `:125-131`; the same in `shell/index.html:64-73` |
| Claim | A theme whose `--primary-background-color` is a colour name, `hsl()` or `color-mix()` yields `null` (run: `parseRgb("black")`, `parseRgb("hsl(0 0% 7%)")`): the copied variables are the theme's, the `.dark` class stays the OS preference's, and since the media listener is not registered when embedded (`shell/App.svelte:182-185`) it does not follow that either. The `rgb()` branch of the bridge does not check for `NaN` as the prepaint script does (`index.html:74`) |
| M / I | M (run) |
| Fix | Resolve the colour through the browser (a probe element's computed `background-color`, which is always `rgb()`), in both places |
| Contract | No |
| Red test | vitest: a parent whose variable is `black` sets `.dark` |

### UR-P1-test-119 — what the unit's tests do not hold

Severity **note**.

| | |
| --- | --- |
| Where | `lib/api/stream.test.ts:4-16` (the fake has no error or open), `lib/api/client.test.ts` (no answer that is not JSON), `shell/route.test.ts` (no inherited page name), `lib/scan/scan.test.ts` (`isPayload` only; `startScan` has no test), no test of `shell/confirm.svelte.ts` or `shell/toast.svelte.ts` |
| Claim | The red tests of 102, 105, 106, 111 and 112 have no neighbour today |
| M / I | M |
| Fix | With each fix |
| Contract | No |
| Red test | — |

## Not settled

- **112**: whether Svelte's production build throws or renders nothing for
  a non-function component was not run (a component test from outside the
  repo could not load the package's `.svelte` dependencies).
- **103**: whether the hosts' proxies speak HTTP/2 to the browser (then no
  limit of six) is the hosts' reviews'.
- **101**: which other pages show a rounded value (`ClusterView`,
  `AttributeTable`, `EventList`, the invoke result) is G-web's; the client
  line is the common cause.
- **102**: the browser's behaviour on a non-200 answer is from the
  EventSource specification, not from a run.
- `Fabric.node_id` as a number is the page's fallback only
  (`lib/pages/node/Fabrics.svelte:40-49`, with a badge when inexact); not
  reported.
- `createTranslator` returns an inherited property for a key such as
  `constructor` (run: a function comes back); every dynamic key in the
  unit's files has a constant prefix, so no site was found and it is not a
  finding.
