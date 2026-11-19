# Phase 3 — verification (Fable)

Verifier of `phase-3-A.md`, `-B.md`, `-C.md`, `-D.md`. Code read at
`f4b8573` (the review worktree; `git diff f4b8573 -- fabricui ui/src contract .github`
is empty). Pins: go-fabric `v0.22.2` (`gf:`), openccu-loom `b40b289a`,
matter.js docs `862ad75f`. Local clones read without a pin of this review:
connectedhomeip at `c55ddad4480`, matter.js at `862ad75f` through
`git show` (its HEAD is `b8979071`). Go 1.27.4.

No high and no medium was reported, so the verification is the sample of
lows (decision 17), the held candidate, three notes with design
consequences, the merges and D's three carried rows.

M = I ran it or read both sides' lines. Runs:

- `p3wt/A`: `go test ./fabricui -run TestP3A_ -count=1 -v` — three tests
  FAIL as the report prints them; `node --test scripts/p3a-guard.test.mjs`
  — 5 of 5 fail as printed.
- `p3wt/B`: the `OpenedAt` case of `servesEphemeral` deleted,
  `go test -count=1 ./fabricui/...` — both packages `ok`; reverted
  (`git status` clean). `go test -coverpkg=./... -coverprofile` and
  `go run ./script/coverfloor` — 94.7 / 81.2 / 44.8 / 0.0, as printed.

## Verdict table — the low sample

| ID | Reviewer sev | Verdict | Final sev | Deciding lines |
| --- | --- | --- | --- | --- |
| UR-P3-ctr-01 (A) | low | confirmed (M) | low | `fabricui/problem.go:143-157` (`decodeJSON`, `badBody` → 413); `handlers_controller.go:104` (`RequestErrorHandlerFunc` → `badBody`), `:110-112` (parameter error → 400); `contract/openapi.yaml`: 26 `requestBody`, no `"413"` on any operation, `payload_too_large` only in the enum `:1939`; topology lists 200/403/501/… and no 400 (`:1789-1797`); `handlers_surface.go:25` (500) against `openapi.yaml:215-217`. Test re-run: "26 of 26 …", "answers 400; the contract declares no 400" |
| UR-P3-test-03 (A) | low | confirmed (M) | low | `ui/scripts/contract-guard.mjs:18-63`: removed path, method, status, schema, field, and newly required — nothing on `required` shrinking, `type`, `pattern`, `maximum`, `enum`. Tests 3–5 re-run, all fail. `docs/contract.md:449-456` names none of these cases |
| UR-P3-test-31 (B) | low | confirmed (M) | low | `script/coverfloor/floors.txt:3-4` (88, 50); `script/coverfloor/main.go:64-66` (`floor, ok := floors[p]` … `if ok && pct+1e-9 < floor`: a package without a line never fails). Measured again: 94.7 %, 81.2 %, `internal/api` 44.8 % floor 0 |
| UR-P3-test-32 (B) | low | confirmed (M) | low | `fabricui/support_test.go:399` (`w.OpenedAt = testNow` for every window); `handlers_pairing.go:123-124`; `gf:node/commissioning.go:361-362` (`snap.OpenedAt, … = window.Timing()` per window); `gf:node/observers.go:200-203`. Mutation re-run: green without the case |
| UR-P3-sup-61 (C) | low | confirmed (M for the file, I for the impact) | low | `.github/workflows/release-on-tag.yml:19-21` (workflow-level `contents: write`, `id-token: write`), one job `:32`, `:37` (`checkout@v7`, no `persist-credentials: false`), `:44` (`setup-go@v7.0.0`), `:59-62` (`make web-install`, `make dist-check`), `:75-76` (`npm install -g npm@11`), `:82` (publish). The report's `:57-58` and `:59-60` are two lines early; the claim stands |
| UR-P3-sup-65 (C) | low | confirmed (M for the config, I for the event) | low | `.github/dependabot.yml:21-22` (`go-minor-patch: update-types: ["minor","patch"]`); `dependabot-auto-merge.yml:19` (only `semver-major` excluded); `go.mod:6` (`v0.22.2`). Refutation tried: a guard that pins the go-fabric version — none: `TestGoFabricVersion` accepts any `v0.` (`handlers_operation_test.go:169-171`), and nothing compares the "at v0.22.2" citations with `go.mod`. `git log --author=dependabot` has no go-fabric bump yet |
| UR-P3-doc-72 (C) | low | confirmed (M, read) | low | `ui/src/shell/App.svelte:62` (`createClient("./api/v1")`, fixed); `fabricui/options.go:336-345` (`Handler`: API under `o.Prefix`, SPA at `/`); `spa.go:58-66` (no such bundle file → 404); `docs/embedding.md:42`, `:380-382`. Nothing in `spa.go` writes the prefix into the page |
| UR-P3-adpt-91 (D) | low | confirmed (M, read) | low | `fabricui/gofabric/gofabric.go:753-766` (only `*ICDMultiAdminError` → `ErrConflict`; else `ErrNodeFailed`); `gf:controller/client/icd_errors.go:79-82` (`ErrICDOperation`: "a call its state refuses"); `gf:controller/client/icd.go:645-654`, `:818-826`; `fabricui/controller.go:903-905`; `gofabric/node.go:848`, `:860-862`. See Merges for the fix's scope |
| UR-P3-doc-90 (D) | low | confirmed (M) | low | `docs/embedding.md:250-252` ("once it lands … until then a host writes the adapter itself"); `gf:logring/logring.go:4-12`, `:116`, `:142`, `:146`, `:239`, `:270`, `:331` — all exact |

Nine of nine sampled lows confirmed, none struck, no severity changed.
The measured false-positive rate of the sample is 0 of 9 (A 0/2, B 0/2,
C 0/3, D 0/2).

Citation precision (not a finding, for the fix wave): A's `handler.go`
line numbers are loose. `hubOf` is `:201-211` (report: `:216-227`),
`newHandler` starts at `:213`, `isAdmin` is `:624-630` (report:
`:635-641`), `subscribe` is `:641-657` (report: `:610-628`);
`docs/contract.md`'s release sentence is `:464-465` (A: `:465-466`; C has
it right). Every claim behind those numbers holds.

## Item 2 — the held candidate UR-P3-adpt-92

**Real: yes (M, lines).** `gf:controller/controller.go:779-787` (`Pair`
ends with `recordPairFailed` / `recordPaired`), `:795` (`tracePairing`
wraps the progress hook); `gf:controller/diagnostics.go:235-277`. The
three functions have no other caller in the non-test tree (grep).
`fabricui/gofabric/gofabric.go:567-619` calls `Registry().ForDevice` and
`reg.Commission` and never `Pair`. So a commissioning of a discovered
device writes no `pairing` entry and no `pair.failures`.

**What still enters the trace.** The other kinds are recorded by hooks,
not by `Pair`: the session ("A secure session with node … was
established", `gf:controller/diagnostics.go:292`), reachability,
discovery. A discovered commissioning is therefore visible in
`GET /controller/events` as a new session, not as a pairing. D's "leaves
no pairing entry" is exact; "no entry at all" would be wrong, and D does
not say it.

**Security item: no.** No role gets a secret or an action. The trace is
an operator's diagnostics ring (bounded, in memory), not an access
decision and not an audit log the module promises. The commissioning is
an `x-admin` action either way.

**Must it stay held with UR-P1-adpt-60: no.** The fact that would point
at adpt-60 — the discovered path goes around `Controller.Pair` — is
already public: the function's own comment (`gofabric.go:560-565`) and
`docs/embedding.md:306` ("`Controller.Pair`, or `Registry().ForDevice`
and `Commission` for a discovered device"). A row that says "this path
records no pairing entry and no `pair.failures`" adds nothing to that.

Conditions for the public copy:

- The row's first line and D's fact note 3 ("goes around `Pair`'s
  defaults and records (held UR-P1-adpt-60 …)") drop the words "class
  of UR-P1-adpt-60", "same cause" and "`Pair`'s defaults". "Defaults" is
  the hint; "records" is not.
- The fix sentence reads "a facade entry for a discovered device and a
  passcode (go-fabric issue)", without "with the fix of UR-P1-adpt-60".

**Severity: note.** **Disposition: upstream**, plus one sentence in
`docs/embedding.md`'s `POST /controller/commission` row. The adapter
cannot close it: `Controller.record` is unexported and the facade's
exported diagnostics are read-only (`DiagnosticEvents`,
`OnDiagnosticEvent`, `gf:controller/diagnostics.go:140`, `:154`).
Decision 25's fix (the attestation default in `CommissionDiscovered`)
leaves this gap open. File it in the same go-fabric issue as decision
25's facade gap, after the phase-3 pull request merges — an issue of its
own before that would name the seam on its own.

## Item 3 — notes with design consequences

### UR-P3-ctr-04 — confirmed (M); the consequence is larger than decision 20 says

The bundled client uses the browser's `EventSource` for all three
streams:

- `ui/src/lib/api/stream.ts:93` — `const defaultOpen: EventSourceFactory = (u) => new EventSource(u)`.
- `:205`, `:249`, `:271` — `open: EventSourceFactory = defaultOpen` on
  `subscribeStream`, `subscribeControllerStream`, `subscribeLogStream`.
- The only callers pass `undefined` for it: `store.svelte.ts:508` (node),
  `controller-store.svelte.ts:246` (controller), `context.ts:162` (log).
- `stream.ts:64` — the exported factory type takes a URL only.
- Server: `fabricui/handler.go:624-630` — without `Options.Admin` the
  token header is the only admin signal; no cookie exists in the package.

What decisions 20 and 41 then need:

1. **Token mode.** Decision 20 moves the log stream to `fetch`. Decision
   41 makes the controller stream `x-admin` too, so it needs the same
   move, or the admin's own page gets 403. The node stream is not gated
   but filtered by admin rights, so over `EventSource` every page is a
   non-admin and loses every event outside the non-admin allow-list
   without an error. All three must carry the header.
2. **Host mode.** `Admin(r)` reads what the browser sends anyway
   (Ingress headers, cookies), so `EventSource` keeps working — until
   decision 38's fixed header applies. The controller and log streams
   are admin routes after 20 and 41, so 38 covers them. For the node
   stream decisions 37–44 do not say whether the filter's admin test
   includes 38's header. **That is the open choice (wizard).**
3. **Recommendation:** one admin test for routes and filter, and all
   three streams over `fetch` with the client's headers. An exception for
   the node stream would be a second definition of "admin".
4. **What the move costs.** The browser's own reconnect and
   `Last-Event-ID` go; the helper already reopens with `BACKOFF_MS`
   (`stream.ts:90`) and resyncs on open, so the reconnect is the helper's
   in every case. A fetch stream needs an abort on unsubscribe and on
   `pauseWhenHidden`, and a parser for SSE frames (comments, multi-line
   `data:`), which wants its own tests against `stream.go`'s framing.
5. **npm API.** `EventSourceFactory` stays (a host's own factory keeps
   working but must now send the headers itself — changelog and
   `ui/README.md`). Additive: a fetch-based factory from the client, the
   default when a client is passed. Contract: the stream operations name
   the header, 403 and decision 42's 503.

### UR-P3-api-08 — confirmed (M)

- `fabricui/handler.go:201-211` — `hubOf`: a host's non-nil `*Events`
  **is** the handler's hub (`return ev, nil`); any other `EventSource`
  becomes `extra`.
- `:633` — `publish` writes into that hub; `handlers_exposure.go:211`,
  `:221`, `handlers_node.go:314` publish `exposure_changed` and
  `endpoints_changed` there.
- `:641-657` — `subscribe` keeps hub, `nodesHub`, `extra` and the node
  apart, so a filter by source is possible only for `extra`.

Consequence for decision 41: with `Options.Events` an `*Events` (what
`NewEvents` and `docs/embedding.md:41` suggest) "host events are
admin-only" can only be done by type, and a host event of an allowed
type passes. A's fix is the right shape: the handler always owns a hub,
a host's `*Events` is subscribed as `extra`, the filter drops `extra`
for a non-admin. Two things for the fix wave:

- A host that subscribes to its own `*Events` to see the handler's
  events stops seeing them. go-fabric-server passes `Events: ev`
  (`internal/daemon/web.go:37`, local clone, unpinned); whether it reads
  from `ev` was not checked here — check both pinned hosts before the
  change, and write the changelog line.
- The allow-list guard test of decision 41 needs a case "a host event
  of an allowed type does not reach a non-admin".

No owner choice: decision 41 already says what must hold.

### UR-P3-ctr-06 — confirmed (M); both library claims are right

- encoding/json: `$GOROOT/src/encoding/json/decode.go:48` ("… match is
  preferred over a case-insensitive one"); `DisallowUnknownFields` does
  not change that. Test re-run: `{"LABEL":"Kitchen"}` → 201,
  `{"Level":"debug"}` → 200.
- oapi-codegen runtime `v1.8.0` (`go.mod:7`), `bindjson.go:36-45`:
  `BindJSONBody` binds a `validatedBody` exactly when a kin-openapi
  validator ran, else `json.NewDecoder(r.Body).Decode(v)` — unknown
  members ignored, case-insensitive, trailing data unread, no option.
  Its own doc says so (`:29-35`). The generated strict handlers call it
  (`internal/api/api.gen.go:6181` and 12 more).

Consequences for decision 43:

1. `DisallowUnknownFields` alone is not "strict": case variants pass,
   and `Decoder.Decode` reads one value and leaves trailing data. The
   node family needs the exact-name and end-of-input checks added to
   `decodeJSON`.
2. The controller family cannot be made strict in the generated code.
   The check runs before it, in `controllerGate`
   (`handlers_controller.go:124-128`, where the body is already
   bounded): read the bounded body, check it against the operation's
   request type, put the bytes back.
3. `encoding/json/v2` (case-sensitive by default, `RejectUnknownMembers`)
   is still behind `GOEXPERIMENT=jsonv2` at Go 1.27.4
   (`encoding/json/v2/doc.go:5`), so it is not an option.
4. A's caution holds and matters: `AclEntryList` is both request and
   response; closing it trips `closedResponseSchemas`
   (`contract-guard.mjs:79-102`). The request needs a schema of its own.

Stays a note; its test belongs to the fix of UR-P0-ctr-28 /
UR-P1-ctr-17. No separate owner choice.

## Item 4 — merges

| Pair | Verdict | Survivor |
| --- | --- | --- |
| UR-P3-test-02 (A) and UR-P3-sup-60 (C) | One finding. Both cite `contract/previous/openapi.yaml:4` (0.8.1) against `contract/openapi.yaml:4` (0.9.0) and tags `v0.9.0`–`v0.9.2`; I read `git show v0.9.2:contract/previous/openapi.yaml` (0.8.1) and `…:contract/openapi.yaml` (0.9.0) | **UR-P3-test-02** (low; it carries the red test). It takes from sup-60: `README.md:84-85`, the missing step in `release-on-tag.yml`, the right lines `docs/contract.md:464-465`, and the stricter guard (byte-equal to the newest tag's contract). UR-P3-sup-60 is closed as merged |
| UR-P3-adpt-91 (D) and UR-P3-test-30 (B), `nodeError` part | Not one finding: adpt-91 is a wrong mapping, test-30 is missing tests. They share the regression test | **Both survive.** UR-P3-adpt-91 owns the behaviour; the `nodeError` table test B proposes is its regression test and must include `ErrICDOperation → ErrConflict`. B's wording "check `errors.Is` against … `ErrNodeFailed`" must not pin today's 502 for `ErrICDOperation`. UR-P3-test-30 keeps `Queue`, `ICD`, `bindingList`, `fabricIndex` and the 0 % methods |
| UR-P3-api-05 (A) and UR-P3-doc-74 (C) | doc-74 is the first half of api-05 (`options.go:333-335`, the index fallback) | **UR-P3-api-05** (low). It takes C's replacement sentence. UR-P3-doc-74 is closed as merged |

Two more overlaps found while reading:

| Pair | Verdict | Survivor |
| --- | --- | --- |
| UR-P3-doc-78 (C) and UR-P3-ctr-01 (A) | The 500 of `/node/surface` missing in the contract is in both (`handlers_surface.go:25`, `openapi.yaml:215-217`, `docs/contract.md:28`); A says "folded into UR-P3-ctr-01" | **UR-P3-ctr-01**; doc-78 closed as merged (it adds `/nodes/{node}/surface`) |
| UR-P3-test-37 (B) and UR-P3-sup-71 (C), `FUZZTIME` | The same dead variable, `Makefile:10` | **UR-P3-sup-71**; test-37 closed as merged. What happens to the variable follows the fuzz decision (wizard) |

On adpt-91's scope: `ErrICDOperation` also covers "no subscription" and
"the peer offline" (`icd_errors.go:79-82`). D maps the whole error to
409. That fits the facade's own definition ("a call its state refuses")
and nothing is sent to the node in any of those cases, so 409 is the
better of the two; the detail text tells them apart. The two inferred
siblings (`ErrDisabled`, `ErrFabricOperationInProgress`) were not
verified here; if their routes do not list 409 today, mapping them is a
contract change and goes to the wizard (decision 13).

## Item 5 — D's carried-row verdicts

### UR-P1-adpt-40 — verdict confirmed (M); the section number is not

| Side | Line read | Unit |
| --- | --- | --- |
| this repo | `fabricui/handlers_controller.go:1531`, `:1537` (`utf8.RuneCountInString(req.Body.Label) > maxFabricLabel`) | runes |
| go-fabric controller | `gf:controller/commissioner/fabric.go:98-101`, `:123-126` | runes |
| matter.js | `packages/protocol/src/fabric/Fabric.ts:194` at `862ad75f` (`label.length > 32`) | UTF-16 units |
| go-fabric node | `gf:cluster/core/operational_credentials.go:2228` (`len(req.Label) > 32` → InvalidCommand), decoder `gf:cluster/spec/codec.go:199-216` (UTF-16 units) | bytes |
| connectedhomeip (local, `c55ddad4480`) | `src/credentials/FabricTable.h:48` (`kFabricLabelMaxLengthInBytes = 32`), `FabricTable.cpp:237`, `:2142` | bytes |

All of D's lines are exact. The claim and the fix (`len(label) > 32` in
the handler; note) stand. One addition: a matter.js device counts UTF-16
units, so "every go-fabric or chip node refuses" is right and "every
node" would not be.

**Spec section: corrected to unverified.** D wrote "Core §7.19.2
'String'". At the sources I can read, §7.19.2 is not a section named
String: go-fabric cites §7.19.2 for the Matter epoch
(`gf:controller/camera/service.go:20`), §7.19.2.31-34 for ids
(`gf:cluster/spec/ids.go:12`), and §7.19.2.40 for the character string's
IS1 rule (`gf:tlv/decode.go:212`); matter.js cites §7.19.2.23 for Fabric
Index and §7.19.2.54 for Price. So the character string appears to be a
subsection of §7.19.2 (§7.19.2.40 by go-fabric's comment), and that its
length constraint counts octets is **unverified** — no local source
states it with a section number. The fix must not put a spec section
into the code comment; the two device implementations are the citation.

### UR-P1-ui-178 — verdict confirmed (M), one line corrected

- `gf:secure/setup/setup.go:333-346` (the twelve values), `:351-361`
  (`IsValidSetupPIN`), `gf:secure/setup/decode.go:91-95`
  (`IsValidPasscode` is `IsValidSetupPIN`), `:192`, `:427` (the QR and
  the manual code call it); `gf:controller/code.go:53-58`.
- **Correction:** "no caller of `IsValidSetupPIN` under `gf:controller/`"
  is true by name only. `gf:controller/client/window.go:237` calls
  `setup.IsValidPasscode` — when it draws a passcode for a window it
  opens, not on the commissioning path. No check of a passcode handed to
  `commissioner.Options.Passcode` exists (grep over `controller/` and
  `secure/pase`), so the verdict stands.
- This repo: `fabricui/handlers_controller.go:1326-1328` checks the
  range only ("passcode: not 1 to 99999998").
- **Spec section §5.1.7.1:** not read in the specification. Two
  implementations name it for exactly this list — connectedhomeip
  `src/setup_payload/SetupPayload.h:268` ("invalid passcode values as
  defined in 5.1.7.1") and matter.js
  `packages/types/src/schema/PairingCodeSchema.ts:158` (HEAD
  `b8979071`; "§ 5.1.7.1", also "§ 5.1.1.6"). Corroborated at second
  hand, the title "Invalid Passcodes" **unverified**.

Severity stays note; fix in the handler and the page as D says.

### UR-P1-ui-179 — verdict confirmed (M)

`gf:cluster/core/access_control.go:699-700` (rule 4), `:749-750`
(`AuthMode=PASE is forbidden in ACL` → CONSTRAINT_ERROR);
`gf:cluster/spec/accesscontrol/definition_gen.go:92`;
`ui/src/lib/pages/controller/node/Acl.svelte:34`, `:112`;
`contract/openapi.yaml:3115`. All exact. chip `AccessControl.cpp` and
matter.js `AccessControlServer.ts:165-265` were read only as go-fabric's
comment cites them, not at their source. No spec section is quoted.
Severity stays note; `pase` leaves the add form, the enum stays.

## Disposition

43 findings reported (A 8, B 8, C 19, D 8): low 17, note 26. After four
merges (sup-60, doc-74, doc-78, test-37) 39 remain: **low 16, note 23**; high 0, medium 0; struck by
refutation 0.

"Wizard" = an owner choice decisions 37–44 do not cover, or a contract
change that decision 13 sends to the wizard, or a decision the plan
defers to phase 3 (9, 11).

### Unit A

| ID | Final sev | Disposition |
| --- | --- | --- |
| UR-P3-ctr-01 | low | **Wizard** (decision 13: contract yes): explicit responses per operation with a Go guard (recommended) or `default: Problem`. Then fix wave. Takes UR-P3-doc-78 |
| UR-P3-test-02 | low | **Fix wave**, first commit of the bump (copy the `v0.9.2` contract to `previous/`), guard in `version-check`, docs sentence. Takes UR-P3-sup-60 |
| UR-P3-test-03 | low | Guard part (b): **fix wave**. Shape part (a): **wizard** (decision 13) — 64-bit values as new string twins (additive, recommended) or a type change; `Warning.message` kept required and empty, or optional; `fabric_count` optional with a `docs/contract.md` line |
| UR-P3-ctr-04 | note | **Wizard**: the node stream's admin test in host mode, and all three streams over `fetch` (recommended); npm API addition. Then fix wave (web and contract) |
| UR-P3-api-05 | low | Godoc: **fix wave**. `NewHandler(o) (http.Handler, error)` (recommended) or `Options.Validate()`: **wizard** (new exported Go API). Takes UR-P3-doc-74 |
| UR-P3-ctr-06 | note | **Fix wave**, inside decision 43's strict bodies (pre-check in `controllerGate`, a request schema of its own for the ACL) |
| UR-P3-ctr-07 | note | **Fix wave** (web): the headers on every request, or a derived flag with a vitest; together with UR-P1-web-197's same-origin rule. Token path: Fable read before merge |
| UR-P3-api-08 | note | **Fix wave**, inside decision 41 (own hub, host `*Events` as `extra`); changelog line; check both hosts first |

### Unit B

| ID | Final sev | Disposition |
| --- | --- | --- |
| UR-P3-test-30 | low | **Fix wave** (Go tests); the `nodeError` table is shared with UR-P3-adpt-91 |
| UR-P3-test-31 | low | **Fix wave** (floors 93 / 80, fail on a package without a line) |
| UR-P3-test-32 | low | **Fix wave** (the fake's window clock and derived fields, two tests) |
| UR-P3-test-33 | low | **Fix wave** (web) — not sampled |
| UR-P3-test-34 | note | **Fix wave** (Go test) |
| UR-P3-test-35 | note | **Fix wave** (web test) |
| UR-P3-test-36 | note | **Fix wave** (Playwright helper) |
| UR-P3-test-37 | — | **Merged** into UR-P3-sup-71 |
| (proposal) standing fuzz job | — | **Wizard** (decision 9 defers it to phase 3): B recommends seeds in every `go test` plus a nightly `make fuzz` |

### Unit C

| ID | Final sev | Disposition |
| --- | --- | --- |
| UR-P3-sup-60 | — | **Merged** into UR-P3-test-02 |
| UR-P3-sup-61 | low | **Fix wave** (three jobs, `persist-credentials: false`, npm pinned). The optional environment with a required reviewer is an owner setting |
| UR-P3-sup-62 | low | **Fix wave** (SHA pins and a guard) — not sampled |
| UR-P3-sup-63 | note | **Fix wave** (ancestor and check-run test in the release job); the `v*` tag ruleset is an owner setting (**wizard**) |
| UR-P3-sup-64 | note | **Wizard** (owner setting, no file): bind the six checks to GitHub Actions |
| UR-P3-sup-65 | low | **Fix wave** (go-fabric out of the auto-merge) |
| UR-P3-sup-66 | note | **Fix wave** (CodeQL `javascript-typescript`); making it a required check is an owner setting |
| UR-P3-sup-67 | low | **Fix wave** (notices row, licence text, guard list) — not sampled |
| UR-P3-sup-68 | note | **Docs-only** |
| UR-P3-sup-69 | note | **Wizard**: the generator in a module of its own, or accepted and recorded |
| UR-P3-sup-70 | note | **Wizard** (decision 11 defers the job to phase 3): C recommends a failing `npm audit --omit=dev` in `web` and a non-failing full audit nightly |
| UR-P3-sup-71 | note | **Docs-only** (README and help text). Takes UR-P3-test-37; `FUZZTIME` follows the fuzz decision |
| UR-P3-doc-72 | low | **Wizard**: `Handler` refuses a non-default `Prefix` (a behaviour change for a host) or the docs say "`Prefix` only with `Mount`" |
| UR-P3-doc-73 | low | **Docs-only** — not sampled |
| UR-P3-doc-74 | — | **Merged** into UR-P3-api-05 |
| UR-P3-doc-75 | note | **Docs-only** |
| UR-P3-doc-76 | note | **Docs-only**, after the gates of decisions 19–40 land |
| UR-P3-doc-77 | note | **Docs-only** |
| UR-P3-doc-78 | — | **Merged** into UR-P3-ctr-01 |

### Unit D

| ID | Final sev | Disposition |
| --- | --- | --- |
| UR-P3-doc-90 | low | **Docs-only** (D's sentence) |
| UR-P3-adpt-91 | low | **Fix wave** (`ErrICDOperation` → `ErrConflict`, red test first). The two inferred siblings: fix wave if their routes list 409, else wizard |
| UR-P3-adpt-92 | note | **Upstream** (go-fabric facade, in decision 25's issue after the merge) plus one docs sentence. Not a security item; public in the reduced wording of item 2 |
| UR-P3-cite-93 | note | **Docs-only** (comment ranges or tags) |
| UR-P3-cite-94 | note | **Docs-only** |
| UR-P3-cite-95 | note | **Docs-only** |
| UR-P3-cite-96 | note | **Docs-only** |
| UR-P3-cite-97 | note | **Docs-only** (one header line per concept page) |

### Carried rows judged in this phase

| Row | Verdict here | Disposition |
| --- | --- | --- |
| UR-P1-adpt-40 (note) | confirmed; spec section unverified | **Fix wave** (`len(label) > 32`), a description line in the contract, no spec section in the comment; **upstream** question (runes on the controller, bytes on the node). It narrows a request, so it is on the wizard's list under decision 13 (recommendation: land) |
| UR-P1-ui-178 (note) | confirmed; one line corrected; §5.1.7.1 at second hand | **Fix wave** (handler 400 and page); same wizard line as above |
| UR-P1-ui-179 (note) | confirmed | **Fix wave** (web) |
| A's carried `api` rows (UR-P1-web-20, -conc-21, -api-171, -web-197, UR-P2-conc-34, decision 24) | not verified here (no sample fell on them) | A's shapes with an owner choice: UR-P1-web-20 (refuse every wildcard, or a docs sentence) and UR-P2-conc-34 (`Options.Done`, a new field) — **wizard**; the others fix wave or docs |

### Wizard items, in one list

1. UR-P3-ctr-04 — the node stream's admin test in host mode; all three
   streams over `fetch`.
2. UR-P3-test-03 (a) — how the breaking shapes of decisions 34, 43, 44
   are written (twins or type change; `Warning.message`; `fabric_count`).
3. UR-P3-ctr-01 — explicit responses or `default: Problem`.
4. UR-P3-api-05 — `NewHandler` or `Options.Validate`.
5. UR-P3-doc-72 — `Handler` refuses a non-default `Prefix`, or docs.
6. Standing fuzz job (decision 9).
7. UR-P3-sup-70 — the `npm audit` job (decision 11).
8. UR-P3-sup-69 — the generator's module.
9. Owner settings: UR-P3-sup-64 (checks bound to Actions), UR-P3-sup-63
   (tag ruleset), and optionally the `npm` environment (sup-61) and the
   CodeQL context (sup-66).
10. Carried: UR-P1-web-20 (wildcards), UR-P2-conc-34 (`Options.Done`),
    UR-P1-adpt-40 and UR-P1-ui-178 (narrowed requests, decision 13).
11. Conditional: UR-P3-adpt-91's two inferred siblings, only if 409 is
    not yet on their routes.
