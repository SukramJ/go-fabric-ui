# Phase 3, unit B: tests and guards

Reviewer B (Opus), session of 2026-11-18. HEAD read: `f4b8573`. Pins:
go-fabric `v0.22.2` (`gf:`), the export in the session scratchpad.
Demonstrations ran in a scratch worktree (`p3wt/B`, detached at `f4b8573`),
never in the review worktree. Every mutation was reverted afterwards. The
saved diffs are under the session scratchpad: `B-mut-adapter.diff`,
`B-mut-nodeerror.diff` and `B-mut-stream.diff`.

No finding here belongs to a held class. None re-reports a row of the four
registers or of the carry-over inventory.

Counts: **high 0, medium 0, low 4, note 4**. A fuzz job is proposed at the
end, as decision 9 asks.

## Findings

### UR-P3-test-30: low: four adapter paths of `fabricui/gofabric` run under no test

- **Where:**
  - `fabricui/gofabric/gofabric.go:753-766` (`nodeError`, 28.6 % covered)
  - `fabricui/gofabric/updates.go:26-34` (`updates.Queue`, 0 %)
  - `fabricui/gofabric/node.go:814-838` (`ICD`, 17.6 %)
  - `fabricui/gofabric/node.go:539-545` (`bindingList`, 0 %)
  - also at 0 %: `RegisterICD`, `UnregisterICD`, `Interview`, `pingSession`, `Enable`, `Disable`
  - `fabricui/gofabric/node.go:483` (`fabricIndex`'s fallback read, 18.2 %)
- **Claim:** The adapter turns go-fabric's facade into the port. Several of
  its mappings are not tested at all. The handler tests run on the fake node
  and on `MemoryController`, which never go through this code. The
  fabrictest end-to-end tests do not reach these paths. Each mapping below
  can be broken and every Go test stays green:
  - **Write errors.** A write value that does not fit the attribute's type
    (`client.ErrJSONValue`) is mapped to `ErrInvalidArgument`, which the
    handler answers with 400. An ICD multi-admin refusal is mapped to
    `ErrConflict`, which the handler answers with 409.
  - **OTA status.** `in-progress` (`gf:controller/softwareupdate/manager.go:123`)
    is rewritten to the contract's `in_progress`.
  - **ICD mode.** The ICD operating mode is mapped to `sit` or `lit`.
  - **Bindings.** The Binding list is built for a write.
- **M/I:** M.
- **Demonstration:** three mutations, run against `go test ./fabricui/...`.
  - Mutation 1: `bindingList` returns `out[:0]`, so a bindings write would
    wipe the device's list. `ICD` reports `lit` for mode 0. `Queue` no
    longer rewrites `-` to `_`.
    ```
    -	return out
    +	return out[:0]
    -		if mode == 1 {
    +		if mode == 0 {
    -			Status:     strings.ReplaceAll(string(p.Status), "-", "_"),
    +			Status:     strings.TrimSpace(string(p.Status)),
    ok  	github.com/SukramJ/go-fabric-ui/fabricui	4.288s
    ok  	github.com/SukramJ/go-fabric-ui/fabricui/gofabric	4.365s
    ```
  - Mutation 2: the `ErrJSONValue` and `ICDMultiAdminError` cases are
    deleted from `nodeError`. A mistyped value is then a 502 instead of a
    400, and an ICD refusal a 502 instead of a 409.
    ```
    -	case errors.Is(err, client.ErrJSONValue):
    -		return fmt.Errorf("%w: %w", fabricui.ErrInvalidArgument, err)
    -	if _, ok := errors.AsType[*client.ICDMultiAdminError](err); ok {
    -		return fmt.Errorf("%w: %w", fabricui.ErrConflict, err)
    -	}
    ok  	github.com/SukramJ/go-fabric-ui/fabricui/gofabric	3.402s
    ```
  - A caveat on the bindings mutation: go-fabric's server overwrites the
    fabric index of a fabric-scoped write. Dropping only `FabricIndex`
    therefore proves nothing, which is why the demonstration empties the
    list instead.
- **Proposed fix:**
  - A table test of `nodeError`: wrap `client.ErrJSONValue` and build a
    `*client.ICDMultiAdminError`, then check `errors.Is` against
    `ErrInvalidArgument`, `ErrConflict` and `ErrNodeFailed`.
  - Lift the status mapping out of `Queue` and test it against every
    `softwareupdate.PendingStatus` constant. The test should fail on a
    value that is not in the contract enum.
  - A fabrictest end-to-end test: write the bindings of the site's plug
    through `PUT …/bindings`, then read them back. Another for an
    attribute write with a mistyped value, which must answer 400.
  - An ICD state test wherever fabrictest can host an ICD device.
    Otherwise a test of the mode mapping alone.
- **Contract:** no.

### UR-P3-test-31: low: the coverage floors trail the measured coverage by up to 31 points, and three packages have no floor

- **Where:**
  - `script/coverfloor/floors.txt:3-4`
  - `script/coverfloor/main.go` (the floor check, near the `pct+1e-9 < floor` comparison)
- **Claim:** The floor is called a ratchet ("raise a floor with the change
  that raises the coverage"), but it was not raised. It no longer stops a
  drop:
  - `fabricui/gofabric` measures **81.2 %** against a floor of 50.
  - `fabricui` measures **94.7 %** against a floor of 88.
  - Packages without a line in `floors.txt` are listed with "floor 0.0"
    and never fail. That covers `internal/api` (44.8 %, the generated
    handlers), `script/coverfloor`, and any package added later.

  A change that deletes the gofabric end-to-end tests of one family loses
  far less than 31 points, so `cover-check` stays green. UR-P3-test-30
  shows that the adapter's weakest paths are already uncovered.
- **M/I:** M.
- **Demonstration:** `go test -coverpkg=./... -coverprofile … ./...` then
  `go run ./script/coverfloor`, at `f4b8573`:
  ```
    fabricui                       94.7%  floor  88.0%
    fabricui/gofabric              81.2%  floor  50.0%
    internal/api                   44.8%  floor   0.0%
    script/coverfloor               0.0%  floor   0.0%
  ```
- **Proposed fix:**
  - Set the floors to the measurement rounded down, minus about 1 point
    for platform noise: `fabricui 93`, `fabricui/gofabric 80`.
  - Make `coverfloor` fail on a package of the profile that has no line in
    `floors.txt`. `internal/api` and `script/coverfloor` can be listed with
    an explicit 0 and a comment.
- **Contract:** no.

### UR-P3-test-32: low: the fake node's window keeps one `OpenedAt` for every window, so the check that tells a replaced window apart is never exercised

- **Where:**
  - `fabricui/support_test.go:384-403` (`fakeNode.window`: `w.OpenedAt = testNow` for every window)
  - `fabricui/handlers_pairing.go:123-124` (`servesEphemeral`: the `win.OpenedAt` comparison)
  - Against go-fabric: `gf:node/commissioning.go:356-366`. `snapshot` takes
    `OpenedAt` from `window.Timing()` for each window, and
    `gf:node/observers.go:200-203` says when a window opened.
- **Claim:** In go-fabric, each window has its own `OpenedAt`. In the fake,
  every window ever opened carries `testNow`, so a window that replaced
  another looks the same as the one it replaced.
  - **The check.** `servesEphemeral` tells the page's fresh window (and
    with it the passcode the page shows) from a later basic window by
    `OpenedAt`. This is the only rule that decides when the later window
    has the same or an unpublished (0) discriminator
    (`handlers_pairing.go:113-118`).
  - **Deleting it stays green.** The whole `fabricui` package passes
    without that case, including the fabrictest end-to-end tests.
  - **The cost.** If the check were lost, the page would show the old
    passcode for a window that does not accept it.
  - **Other divergences, not material today.** The fake reports `Own:
    true` for any kind and `CommissioningMode` 0 with a non-zero
    discriminator, which go-fabric never does: `Own` is false exactly for
    `WindowEnhanced`, and the discriminator is 0 when the mode is 0
    (`gf:node/observers.go:181-196`). No handler reads `Own` or
    `CommissioningMode` today. The next one that does will be tested
    against impossible states.
- **M/I:** M for the mutation. I for the cost to the user.
- **Demonstration:**
  ```
  -	case !eph.openedAt.IsZero() && !win.OpenedAt.Equal(eph.openedAt):
  -		return false
  ok  	github.com/SukramJ/go-fabric-ui/fabricui	5.145s
  ok  	github.com/SukramJ/go-fabric-ui/fabricui/gofabric	6.338s
  ```
  The control: deleting the next case, the `Deadline` check, does fail
  `TestPairing_ForgetsAFreshWindowTheNodeNoLongerServes`. The tests do
  reach this function.
- **Proposed fix:**
  - Give the fake a window clock that advances on every open.
  - Derive `Own`, `Kind`, `CommissioningMode` and `Discriminator` as go-fabric's
    `snapshot` does.
  - Add a test: a fresh window is replaced by a basic window with the same
    discriminator, after which `GET /node/pairing` carries no passcode. Add
    the same case where the replacing window's discriminator is 0.
- **Contract:** no.

### UR-P3-test-33: low: the i18n coverage tests check template-literal keys only by prefix; two contract enums, and three lists typed as arrays, can lose a key unnoticed

- **Where:**
  - `ui/src/lib/i18n/coverage.test.ts:49-52` (the prefix check)
  - the call sites:
    - `ui/src/lib/pages/node/Surface.svelte` (`t(\`node.surface.kind.${…}\`)`, `t(\`node.surface.verdict.${…}\`)`)
    - `ui/src/lib/pages/controller/node/Overview.svelte:112` (`controller.overview.flag.${f}`)
  - `ui/src/lib/i18n/catalogs.test.ts:86-100` (`BLEDecision["reason"][]`, boot reasons and pairing phases as literal arrays)
  - the enums in the contract: `contract/openapi.yaml:2649`, `:2661`, `:2667`, `:2691`
- **Claim:** For a `t(\`prefix.${x}\`)` call, the test only asks whether
  *some* key starts with the prefix.
  - **Records exist for some enums.** Most enums the pages translate are
    also listed through a `Record<Enum, true>`, which fails the type check
    when the contract adds a value.
  - **Two have none.** The surface report's `kind` (`device-type, cluster,
    feature, attribute, command, event`) and `verdict` (`keep, warn,
    remove, refuse`) are not listed that way. Either catalogue can lose
    one of their keys, and the page then shows the raw key, for example
    `node.surface.verdict.warn`.
  - **Arrays do not check completeness.** The Bluetooth LE reasons are an
    array of the union type. The comment calls them "The contract's
    BLEDecision.reason enum", but an array of the union is not checked for
    completeness: a new contract value type-checks and has no label. The
    same holds for the hard-coded boot reasons 0–6 and the pairing phases.
- **M/I:** M for the deletion. I for the array typing: these are type-level
  semantics, and `svelte-check` was not run against a changed `schema.gen.ts`.
- **Demonstration:** `node.surface.verdict.warn` and `node.surface.kind.event`
  were deleted from both `en.json` and `de.json`, then the suite was run.
  - The i18n tests together with Surface and Selection:
    ```
     Test Files  4 passed (4)
          Tests  47 passed (47)
    ```
  - The whole vitest suite:
    ```
     Test Files  56 passed (56)
          Tests  581 passed (581)
    ```
- **Proposed fix:** In `coverage.test.ts`, add `Record`s over
  `SurfaceElement["kind"]`, `["verdict"]` (and `origin` if a page labels
  it) and the change kinds. Make the BLE reasons, the pairing phases and
  the boot reasons `Record`s over their contract types, as the other
  enums already are. Better still, have the template-literal check fail
  for any prefix that no `Record` or constant list enumerates.
- **Contract:** no.

### UR-P3-test-34: note: `TestStream_DataStaysOnOneLine` does not check that the data stays on one line

- **Where:**
  - `fabricui/stream_test.go:185-222`
  - `fabricui/stream.go:137` (`json.Marshal(ev.Data)`)
- **Claim:** The guard of UR-P2-test-95 checks three things: the sequence of
  `event:` lines, and that no `data:` line holds a `\r`. It does not check
  that a frame is one `data:` line, or that the data parses as JSON.
  - **What survives.** Encoding with `json.MarshalIndent` passes the guard,
    yet it would cut every multi-key event short in the browser:
    `EventSource` would see `data: {`, and the indented continuation lines
    would be dropped as unknown fields.
  - **Who catches it.** Other tests of the package do, so the package
    bites. The guard named for the class does not.
- **M/I:** M.
- **Demonstration:** the diff is `B-mut-stream.diff`.
  ```
  -			data, err := json.Marshal(ev.Data)
  +			data, err := json.MarshalIndent(ev.Data, "", " ")
  go test -run TestStream_DataStaysOnOneLine ./fabricui/   →  ok
  go test ./fabricui/...  →  FAIL TestStream_DeliversEventsAndHeartbeats, TestControllerStream_DeliversEvents, … (7 tests)
  ```
- **Proposed fix:** In the guard, require every line of the stream to be
  one of four kinds: `event: <type>`, `data: <valid JSON>`, empty, or a
  `:` comment. Also require each frame to carry exactly one `data:` line.
- **Contract:** no.

### UR-P3-test-35: note: the each-keys guard allows a key by file and expression, not by each block

- **Where:**
  - `ui/src/lib/each-keys.test.ts:16-34` (`UNIQUE`, keyed `"file :: key"`)
- **Claim:** A second `{#each}` in an allow-listed file passes the guard if
  it reuses an allowed key name, even when its list comes from node data.
  This is the class of UR-P1-ui-130 and UR-P2-ui-90. For example,
  `Overview.svelte :: f` is allowed for the constant flags, and a new
  `{#each ping.addresses.map((x) => x.address) as f (f)}` is accepted. The
  ping addresses can repeat.
- **M/I:** M.
- **Demonstration:**
  - Mutation, adding one line to `Overview.svelte`:
    ```
    +          <p>{#each ping.addresses.map((x) => x.address) as f (f)}<span>{f}</span>{/each}</p>
    npx vitest run src/lib/each-keys.test.ts   →  Tests 3 passed (3)
    ```
  - Control: the same line keyed `(addr)` fails the guard, which reports
    `"lib/pages/controller/node/Overview.svelte :: addr …"`.
- **Proposed fix:** Key the allow list by the whole `{#each …}` tag (the
  list expression and the key). Alternatively, fail when one entry
  matches more than one tag.
- **Contract:** no.

### UR-P3-test-36: note: the five Playwright flow tests check neither console errors nor policy violations

- **Where:** `ui/e2e/smoke.spec.ts`
  - the per-page smokes, `:59-72`, which collect console errors, `pageerror`
    and CSP violations
  - the flows, `:116-212`
- **Claim:** The per-page smokes collect console errors, `pageerror` and
  CSP violations. The flows do less:
  - "node selector" (`:116`) collects nothing.
  - The other four collect `pageerror` only: node create/delete (`:128`),
    re-announce (`:155`), commissioning with terms (`:167`) and the BLE
    preselection (`:187`).
  - The erase test (`:202`) collects `pageerror` only too.

  The states these flows reach are seen in a browser only here: the
  confirm dialogs, toasts, the commissioning step list and the erased
  page. A CSP violation or a console error in any of them would pass.
  This is decision 10's baseline evidence, so it is a note.
- **M/I:** I: read only, no Playwright run was spent.
- **Both sides:**
  - `smoke.spec.ts:61-64,69-71` (`console` and `violations()` in the smokes)
  - `smoke.spec.ts:130,157,169,189,204` (`pageerror` only in the flows)
- **Proposed fix:** One helper, `watch(page)`, that returns `errors` and
  `violations()`. Use it in every test, and assert both at the end.
- **Contract:** no.

### UR-P3-test-37: note: `FUZZTIME` is a dead variable

- **Where:** `Makefile:10` (`FUZZTIME ?= 10000x`)
- **Claim:** The variable has been in the Makefile since the skeleton
  (`406eb28`), but no target uses it and the tree has no `Fuzz*` function.
  It suggests a fuzz target that does not exist. The proposal below gives
  it a use; failing that, delete it.
- **M/I:** M (`grep -rn "func Fuzz"` finds nothing; `git log -S FUZZTIME` shows only `406eb28` and `b0bb2c6`).
- **Proposed fix:** see the fuzz proposal.
- **Contract:** no.

## A standing fuzz job (decision 9): yes, in two parts

Phase 0 ran eight native targets for 5 minutes each, with no hit, and
deleted them afterwards (`phase-0-evidence.md:112-134`). A standing job is
worth having now: the phase 3 fix wave rewrites exactly the parse sites
those targets covered.

- Strict bodies with trailing data, and 405 with `Allow` (decision 43).
- A new 415 and a required request header (decision 38).
- A hex pattern on `NodeID`, 64-bit values as strings (decision 44).
- An admin check for every mode and route (decision 37).

**Part 1, committed targets with seed corpora (every `go test`, no cost
worth counting).** The seeds of a `Fuzz*` function run as ordinary tests,
so each target becomes a regression property. The targets, in
`fabricui/fuzz_test.go` and `fabricui/gofabric/fuzz_test.go`, with the
phase 0 properties extended to phase 3's new shapes:

1. `FuzzDecodeBodies`: `decodeJSON` and `badBody`. Every refusal is 400,
   413 or 415 `application/problem+json`, with a matching `X-Problem-Code`.
   Trailing data and unknown fields are refused.
2. `FuzzControllerBodies`: every controller write route a fixture names,
   on `MemoryController`. No 5xx outside 501–503, and every non-2xx is a
   problem.
3. `FuzzMux`: the whole mux in token mode and in `Options.Admin` mode.
   Every refusal is a problem. No `x-admin` route answers 2xx without
   rights. The security headers are on every answer (the property of
   `TestAPI_EveryRouteAnswersWithTheSecurityHeaders`, under fuzzed paths
   and methods).
4. `FuzzParseHex`: `parseHex`, `parseNodeID`, and the contract's new
   `NodeID` pattern. A value accepted by Go matches the pattern, and the
   reverse.
5. `FuzzQueryWindows`: `queryInt`, `eventWindow`, `pathEndpoint`,
   `logWindow`, `endpointOf`.
6. `FuzzStrings`: `nodeLabel`, `checkNodeKey`, `parseLevel`.
7. `FuzzThreadDatasetName`, in gofabric.
8. `FuzzDiscoveredAndCode`, in gofabric.

**Part 2, a nightly run.** Add a target `make fuzz`, which finally uses
`FUZZTIME`. It loops over the targets, since `go test -fuzz` takes one at
a time. Add a job to `nightly.yml` with `FUZZTIME=3m`, about 25 minutes
on one runner. A hit fails the nightly run. The failing input goes under
`testdata/fuzz/` with its fix, so it becomes a seed.

Not on pull requests, where the seeds are enough. No web fuzzing: the
client's parse sites are `JSON.parse` of the server's own answers.

## Covered without finding

- **The "no assertion" heuristic list** (`read-review-p3-tests.md`). All 27
  hits except one call `runFixture`, which compares the status and the
  normalised body against the fixture (`support_test.go:74-112`). The
  exception, `TestHooks_ConcurrentCloseAndUpdates`, is a race and panic
  test under `-race`, which CI runs (`ci.yml` test job).
- **Every fixture is run, not only named.** I instrumented `runFixture`
  with a `TestMain` report. All 142 fixtures in `contract/fixtures` go
  through `runFixture` during `go test ./fabricui/`, so none is "used" only
  through a string. The demonstration was reverted.
- **Trailing data on a 200** fails 30+ tests of the package (a mutation of
  `writeJSON` that appends a second JSON document). `runFixture` alone
  would not notice it, because `normaliseJSON` reads only the first value,
  but the package bites.
- **Guards that bite under mutation:**
  - `TestAPI_NoRouteAdmitsAnotherOrigin` fails when `apiHeaders` echoes
    the `Origin` as `Access-Control-Allow-Origin`.
  - `TestAPI_EveryRouteAnswersWithTheSecurityHeaders` fails when
    `no-store` is skipped on `…/log`.
  - The gate on node routes is bitten through
    `TestNodes_AdminRoutesShareTheToken`: the `/nodes/{node}` routes were
    all served as `read`.
  - The controller gate is held by
    `TestController_AdminRoutesRefuseWithoutRights` over every
    `controllerAdmin` route in both modes, plus
    `TestControllerAdminMatchesTheContract`.
  - The `each-keys` guard catches a new key from data, as the control in
    UR-P3-test-35 shows.
  - The `servesEphemeral` deadline check fails one test when deleted.
- **Flakes and order.** I ran
  `-race -count=20 -cpu 1,4 -run 'TestStream_|TestHooks_|TestPing|TestMisc|TestSubscribe|TestNodeSwitch'`
  on `fabricui` and `fabricui/gofabric` (85 s), then
  `go test -shuffle=on ./fabricui/ ./fabricui/gofabric/ ./internal/...`.
  Both were green, with no failure. The sleeps that do the waiting
  (`hooks_test.go:41,62,170`) wait for an event that must *not* happen,
  so a slow machine only weakens them, it does not make them fail.
- **`App.test.ts:92-112`** (UR-P1-conc-199) does bite. The first backoff
  is 1000 ms after a 100 ms probe, inside the 1500 ms wait
  (`App.svelte:163`).
- **The Playwright order.** `workers: 1`, `fullyParallel: false` and
  `retries: 0` (`playwright.config.ts:14-17`). The tests that change state
  run last or restore it (node create and delete; erase at the end).
  Only `--repeat-each` would expose the order, so no finding.
- **`fixtureFetch`** answers without `Content-Type` or `X-Problem-Code`.
  The client reads neither (`client.ts`, grep), so there is no divergence
  that matters.
- **`MemoryController`** accepts any JSON value on a write, where gofabric
  would answer 400. The handler's error mapping is tested on its own
  (`handlers_controller_test.go:362-364`), so the gap is the adapter's,
  filed as UR-P3-test-30.
- **The e2e web server** (`e2eweb_test.go`) builds its mux as `Handler`
  does: `register` plus `spaHandler` (`options.go:336-344`).
- **What happy-dom cannot show** (CSP, frames, cross-origin reads, the
  camera) is in the Playwright smokes and `origins.spec.ts`, as decision
  10 intends. The gaps there are filed as UR-P3-test-36.
- **Not run:** `make e2e-web`, so the Playwright findings are read only;
  `svelte-check`.
