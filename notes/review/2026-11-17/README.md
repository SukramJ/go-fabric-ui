# Whole-codebase review, November 2026

Owner decision 2026-11-17. Goal: every weakness of go-fabric-ui found and
eliminated, over every aspect, before the hosts' own reviews
(go-fabric-server, then go-ha2fabric) build on it. The model is go-fabric's
Review II (`go-fabric/notes/review/2026-11-13/`), adapted: this module is a
web interface on go-fabric's facade, so the cuts are the HTTP handler, the
controller adapter, the contract and the Svelte pages, not a protocol stack.

This file is the plan. It holds no finding. Every statement about the code
carries a `file:line` read at the baseline; every size is an estimate and
says so.

## Pins and baseline

| Pin | Value | Source |
| --- | --- | --- |
| Baseline HEAD | `25f6e7b4ad743d25932406b88e386d945f47c54f` (origin/main after #73, 2026-11-17) | `git rev-parse origin/main` after `git fetch` |
| Review branch | `review/p0`, worktree `../go-fabric-ui-review` | — |
| go-fabric | `v0.22.2` | `go.mod:6` |
| Contract | `0.9.0` | `contract/openapi.yaml:4`; module version `0.9.2` (`ui/package.json:3`) |
| Reference pages | openccu-loom `b40b289a` | `CLAUDE.md`, "The reference for the pages" |
| matter.js (pairing hints only) | `862ad75f` | `ui/src/lib/pages/node/ecosystems.ts:1-3` |

Reviewers read go-fabric, openccu-loom and matter.js as `git archive`
exports of these pins, never the sibling checkouts' working trees.

In flight and not part of the baseline: `wave/U12-onboarding-inventory`
(issue #68; three notes commits ahead of main on 2026-11-17, no code yet).
What it lands is read as delta by the next phase (decision 3).

## Owner decisions (2026-11-17)

`decisions.md` carries them as rows 1–6.

| Decision | Choice |
| --- | --- |
| Scope | go-fabric-ui, then go-fabric-server, then go-ha2fabric, one after the other; this plan is go-fabric-ui only. The Loom repositories are excluded |
| Cut | **Adapted, five phases**: 0 basis, 1 package pass, 2 flows and adversarial, 3 architecture, tests, docs and supply chain, 4 convergence. No matter.js parity phase; citations of go-fabric, matter.js and the reference pages stay checked as `CLAUDE.md` has it |
| Baseline | origin/main **now** (`25f6e7b4`); feature work (U12, #68) continues, each phase re-reads what landed as delta |
| Models | As go-fabric decision 24 and this repository's model table: **Opus by default** for reviews and implementation; **Fable** for the security paths' review (admin and token paths, auth, the streams, whatever renders untrusted data), the verification of every high and medium finding, every reference citation, and planning; **Sonnet or a script** for inventories |
| Process | As go-fabric decisions 20, 25, 27: review and verification per phase as a workflow (inventory → reviewer per theme → an independent verifier per finding), **at most five agents per phase for review and verification and at most two for a fix wave** (decision 6: this machine is slower; replaces the ten of decision 5); verification in batches; **one bundled fix pull request per phase**, squash, auto-merge on green CI; open security items held back from the public notes until their fix merges |
| Sessions | **One fresh session per phase.** The plan for go-fabric-server is written while go-fabric-ui is in phase 4 |
| Done | Every finding down to note ends fixed, struck with a reason, accepted by a recorded decision, or handed to go-fabric with an issue; one standing guard per defect class; convergence runs until one yields no new high or medium |

## What is in this directory, and what is not

The plan (this file), the owner's decisions (`decisions.md`), the
dispositions of every finding (`phase-<N>-register.md`) and the reviewers'
and fix agents' working reports (`phase-<N>-<unit>.md`). Inventories go to
`notes/reads/read-review-p<N>-<unit>.md`, as `CLAUDE.md` places first reads.

A report that mentions a security item still open is not committed until
the pull request that fixes it; the register then says how many rows are
held back (decision 5). Where the held-back files live meanwhile is open
question 8.

## The tree the phases are sized from

Measured at the baseline with `git ls-files | xargs wc -l`; "about" marks a
sum rounded by hand.

| Area | Files | Lines (about) | Read by |
| --- | --- | --- | --- |
| `fabricui/` production Go | 23 | 9 000 | phase 1, in full |
| `fabricui/gofabric/` production Go (the controller adapter) | 6 | 2 300 | phase 1, in full |
| Go tests under `fabricui/…` | 25 | 7 800 | phase 3 (quality); phase 1 reads the tests of the unit it reviews |
| `internal/arch`, `internal/version`, `script/` | 6 | 600 | phase 3 |
| `internal/api/api.gen.go` (oapi-codegen, generated) | 1 | 6 950 | not read line by line; `make generate-check` and the phase 3 contract unit hold it |
| `contract/openapi.yaml`, `contract/previous/`, 142 fixtures | 144 | 3 350 + 3 240 + fixtures | phase 3 |
| `ui/src/lib` and `ui/src/shell`, hand-written, without tests | 95 | 9 000 | phase 1, in full |
| `ui/src` tests (`*.test.ts`, `ui/src/test/`), `ui/e2e` | 38 | 4 600 | phase 3 |
| `ui/src/lib/api/schema.gen.ts` (generated), `i18n/{en,de}.json` | 3 | 5 880 + 1 500 | generated: guard only; catalogues: phase 1 unit G-web by key, not by line |
| `fabricui/dist/` (committed bundle, 616 KiB) | 7 | — | `make dist-check`; phase 3 supply chain |
| `docs/`, `README.md`, `CONCEPT.md`, `changelog.md`, `ui/README.md` | 20 | 5 000 | phase 3 |
| `.github/workflows/`, `dependabot.yml`, `Makefile`, `.golangci.yaml`, `ui/` build config | 14 | 900 | phase 3 |

Facts the cut rests on, read at the baseline:

- **Gate.** Routes are registered either plain or behind `adminOnly`
  (`fabricui/handler.go:328-329`); the node family's table marks which
  (`handler.go:285-310`), the registry's three routes and
  `PUT /operation/log/level` are admin (`handler.go:314-318`, `:353`), the
  controller family is gated by a table held against the contract's
  `x-admin` (`fabricui/handlers_controller.go:51-74`, gate `:120-142`).
  Every other route, the streams included (`handler.go:309`, `:351-355`),
  is served without a check.
- **Two admin modes.** `Options.Admin` decides when set
  (`handler.go:487-490`); nil selects token mode (`options.go:218-225`): a
  16-byte random token per process (`handler.go:265-269`), compared in
  constant time against the `X-Fabric-UI-Token` header
  (`handler.go:491-492`) and served as `csrf_token` by `GET /node/status`
  and `GET /controller/status` (`handlers_node.go:138`,
  `handlers_controller.go:320`), both unguarded reads. The contract
  describes token mode as a defence against a page of another origin
  (`docs/contract.md:370-374`). The handler has no user model of its own;
  authentication is the host's or its proxy's (`docs/embedding.md:356-359`).
- **Bodies.** JSON with `DisallowUnknownFields` under a 1 MiB bound
  (`fabricui/problem.go:33`, `:125-128`; the controller gate
  `handlers_controller.go:137-139`).
- **Streams.** One SSE loop for the node stream, the controller stream and
  the log stream, a 64-event buffer per client that drops when full
  (`fabricui/stream.go:17`, `:34-39`, `fabricui/operations.go:274-284`).
- **SPA.** The embedded bundle with an index fallback; the only response
  header it sets beside caching is `X-Content-Type-Options`
  (`fabricui/spa.go:37`, `:54`).
- **Rendering.** Two `{@html}` sites, both the QR SVG the page builds
  itself (`ui/src/lib/pages/node/Pair.svelte:150`,
  `ui/src/lib/pages/controller/node/Window.svelte:66`, from
  `ui/src/lib/qr.ts:46-57`); no `innerHTML` in `ui/src`.
- **Fuzz.** No fuzz target exists; the pipeline's header says why
  (`.github/workflows/ci.yml:9-11`), and `FUZZTIME` in `Makefile:10` has no
  target that uses it.
- **Coverage floors.** `fabricui` 88, `fabricui/gofabric` 50
  (`script/coverfloor/floors.txt:3-4`).
- **npm.** `npm ci --no-audit` (`Makefile:83`); no audit step in a
  workflow.
- **Imports.** The facade and adapter allow-lists are a test
  (`internal/arch/arch_test.go:30-66`).

## Phases

| Phase | Content |
| --- | --- |
| 0 Basis | Pin the baseline (above). Mechanical evidence, each with HEAD and a summary line: `make ci`, `make race`, `make cover-check` plus an error-path coverage reading (which `writeProblem` / `controllerError` branches no test reaches), `make vuln licenses secrets`, `GOARCH=386 make build`, `make contract-check`, `make e2e`, `make e2e-web`, `npm audit` (all and `--omit=dev`), `svelte-check`, vitest; ad hoc fuzzing where input is parsed (open question 3). The threat model, new (`notes/review/2026-11-17/threat-model.md`): the HTTP handler, the admin and token paths, the three SSE streams, the SPA (XSS, CSRF, framing, caching), the embedding host boundary (`Options.Admin`, `Mount` beside a host's own routes, the Ingress theme bridge), secrets (passcodes, the onboarding codes, Wi-Fi and Thread credentials) |
| 1 Package pass | Every hand-written Go package and the `ui/` source, unit by unit (table below) |
| 2 Flows and adversarial | End to end across handler → `gofabric` adapter → go-fabric's facade at `v0.22.2`, and page → client → handler; the attacker roles in two units, exploits written as failing tests (table below) |
| 3 Architecture, tests, docs, supply chain | Contract and OpenAPI against handlers, client and fixtures; the exported `fabricui` API (`Options`, the ports) and the import cut; test quality on both halves; every behavioural claim in `docs/` and every citation against code and pins; workflows, release, npm publish, the committed bundle |
| 4 Convergence | A full-cut run over the HEAD of the day, repeated until one yields no new high or medium. Includes the delta of what feature work landed and the delta when go-fabric's post-Review-II tag is pinned in `go.mod` |

### Phase 1 units

Four reviewers and one verifier (decision 6). The security paths are cut
out of the packages they sit in, so that Fable reads them whole and Opus
reads the rest.

| Unit | Files | Lines (about) | Model |
| --- | --- | --- | --- |
| S-go, the Go security paths | `fabricui/handler.go`, `options.go`, `problem.go`, `stream.go`, `spa.go`, `events.go`, `ports.go`, `registry.go`, `nodeswitch.go`, `operations.go`, `doc.go`, `handlers_pairing.go`; of `handlers_node.go` the status and maintenance routes; of `handlers_controller.go` the gate, status, commissioning, credentials and window routes | 3 600 | **Fable** |
| S-web, the web security paths | `ui/src/lib/api/client.ts`, `stream.ts`, `ui/src/lib/context.ts`, `store.svelte.ts`, `controller-store.svelte.ts`, `nodes-store.svelte.ts`, `qr.ts`, `scan/`, `i18n/index.ts`, `theme/ha-bridge.ts`, `ui/src/shell/*` (the prepaint script of `index.html` included); in every page and kit component the sinks that render a host-, node- or log-supplied string, from the inventory's list | 2 700 + the sinks | **Fable** |
| G-go, the Go rest | The rest of `handlers_node.go` and `handlers_controller.go`; `handlers_exposure.go`, `handlers_diagnostics.go`, `handlers_topology.go`, `handlers_surface.go`, `node_detail.go`, `provider.go`, `devicetype_names.go`, `controller.go`, `memory_controller.go`; `fabricui/gofabric/*.go` (6 production files) | 7 700 | Opus |
| G-web, the web rest | `ui/src/lib/pages/node/*`, `pages/operation/*`, `pages/controller/**`, `ui/src/lib/kit/*`, `theme/tokens.css`, the catalogues by key | 7 300 | Opus |

The adapter's behaviour against go-fabric's controller facade at `v0.22.2`
and its go-fabric citations are read by Fable in phase 3's `cite` unit, not
here. G-go and G-web are large for one reader each; open question 10.

### Phase 2 units

Four reviewers and one verifier. A role's exploit is written as a failing
test.

| Unit | Content | Model |
| --- | --- | --- |
| X1 secret flows | **Pairing:** window on fresh or configured credentials → the ephemeral record → `GET /node/pairing` → the stream's window events → the Pair page's QR and countdown; close, expiry, a second node, a `NodeSwitch` swap, erase and factory reset. **Commissioning:** Commission page → `POST /controller/commission` → adapter → go-fabric's commissioner; the credential store, the BLE decision, the terms retry, progress events on the controller stream, abort and context end | Fable |
| X2 write and life-cycle flows | **Writes:** ACL, bindings, attributes, commands, fabric removal, ICD, OTA consent, exposure apply → reassemble: request decoding → adapter encoding → the facade call → the error's problem → what the page shows. **Life cycle:** stream subscribe and unsubscribe against `node.Observe`, slow and vanished clients, goroutine and observer leaks, registry create, rename, delete against the handler cache, the host's shutdown | Opus |
| X3 a client without admin rights | Two roles on the HTTP side: an unauthenticated client on the listener's network, in token mode and in host mode; and a non-admin viewer, who exists only in host mode where `Options.Admin` answers false — what each reads, through the unguarded routes and streams, and what each can drive | Fable |
| X4 the browser | Two roles on the page's side: a malicious page or origin (forged writes, cross-origin reads of status and streams, DNS rebinding, framing, a hostile parent or sibling frame under an Ingress); and hostile node and host data (names, labels, attribute and event values, discovery results, mDNS records, log records, the host's warnings and counters, problem details) through JSON, the SSE framing and every page that renders them | Fable |

### Phase 3 units

Four reviewers and one verifier.

| Unit | Content | Model |
| --- | --- | --- |
| contract and api | `contract/openapi.yaml` (87 paths, 97 operations, 48 `x-admin` marks) against the hand-written routes, the generated ones, `ui/src/lib/api/client.ts`, the fixtures and `docs/contract.md`; `ui/scripts/contract-guard.mjs` and `contract/previous/`. The exported surface of `fabricui` and `fabricui/gofabric` (`Options`, `NodeSource`, `ControllerSource`, `ExposureProvider`, `NodeRegistry`, `Operations`, `EventSource`, `NodeSwitch`), the npm package's exports (`ui/package.json`), `internal/arch` | Opus |
| tests | Go: tests that assert nothing, missing negative cases, the doubles against the real facade, the `gofabric` floor of 50, flakes, the e2e tests. Web: vitest (32 files), the Playwright smoke (`ui/e2e/smoke.spec.ts`, 7 tests), the i18n coverage tests, the happy-dom limits | Opus |
| docs and supply | Every behavioural claim in `README.md`, `CONCEPT.md`, `docs/*.md`, `ui/README.md` and `changelog.md` against the code. `.github/workflows/*` (permissions, pins, `pull_request_target`), `release-on-tag.yml` and the npm publish, Dependabot and its auto-merge, branch protection against the jobs (ten required contexts on 2026-11-17, `gh api`), `go.mod`'s indirect set, npm licences and `THIRD-PARTY-NOTICES.md`, the committed bundle and `dist-check` | Opus |
| cite | Every citation of go-fabric (at the tag the sentence names and at `v0.22.2`), openccu-loom `b40b289a` and matter.js `862ad75f` in code and docs. The coordinator's script lists them and checks that each range exists; Fable reads each that is the reason for behaviour, each doubtful one and a sample of the rest, and with them the adapter's behaviour against go-fabric's controller facade | Fable |

## Aspects

Authentication and the admin gate · cross-origin and browser security (CSRF,
XSS, framing, caching) · secrets in responses, events and logs · untrusted
data from nodes and hosts · concurrency and life cycle · resource bounds ·
error paths (fail open, the problem mapping, nil ports) · correctness
against go-fabric's facade · contract consistency across both halves · page
logic, state and i18n · API and architecture · tests and guards · supply
chain and release · docs and citations against code.

## Method per phase

1. **Map.** A script or Sonnet inventory per unit
   (`notes/reads/read-review-p<N>-<unit>.md`), `file:line` for every fact:
   routes and their gate, goroutines, locks, subscriptions, parse sites,
   log calls, every sink that renders a string. A map, never a citation.
2. **Review.** One reviewer per unit, briefed with the map, the file list
   and the pinned sources; the model as the tables say. Findings into
   `phase-<N>-<unit>.md`: ID, severity (high / medium / low / note),
   `file:line`, the claim, the reference line where one is the reason, the
   proposed fix. Measured and inferred are reported apart.
3. **Verify.** One independent Fable verifier per phase re-reads every
   high and medium against the code and the pins, in batches by unit, and
   tries to refute each — struck with a reason, or confirmed. A low or a
   note has no verifier of its own: the fix's regression test, red before
   the fix, is its verification, and one that cannot be made red is struck.
4. **Fix.** Branch `review/p<N>`; the regression test first (red, then
   green); **at most two implementing agents, on disjoint paths** (the Go
   half and the web half; decision 6); one standing guard per defect class; the local guard set of
   `CLAUDE.md` in the background before the pull request; one pull request
   per phase, squash, auto-merge on green CI. Opus implements; a fix on
   the admin and token paths gets the Fable read before it merges
   (`CLAUDE.md`'s model table). Decisions go to the owner as a wizard, the
   recommendation first. A facade gap becomes a go-fabric issue, not a
   workaround. A confirmed vulnerability is not committed to the register
   before its fix merges.

The coordinator is Opus in **one fresh session per phase** (decision 6);
the state is in these notes and the memory file, never in a long context.
Push, pull request and tag are that session's own calls. The plan for
go-fabric-server is written while this review is in phase 4.

## Agents per phase (estimates)

Caps (decision 6): five for review and verification, two for a fix wave.
Inventories are scripts run by the coordinator wherever a script can
produce them; a Sonnet inventory counts as an agent.

| Phase | Review and verification | Cut | Fix wave |
| --- | --- | --- | --- |
| 0 | 4 | 1 inventory (Sonnet: routes and gates, parse sites, render sinks, the hosts' `Options` call sites), 1 mechanical evidence with fuzzing and the error-path coverage reading (Opus), 1 threat model (Fable), 1 verifier (Fable) | ≤ 2 |
| 1 | 5 | 2 Fable (S-go, S-web), 2 Opus (G-go, G-web), 1 Fable verifier | ≤ 2 |
| 2 | 5 | 3 Fable (X1, X3, X4), 1 Opus (X2), 1 Fable verifier | ≤ 2 |
| 3 | 5 | 3 Opus (contract and api, tests, docs and supply), 1 Fable (cite), 1 Fable verifier | ≤ 2 |
| 4 | 4 per run | 1 Fable (the security paths of both halves), 2 Opus (the Go rest, the web rest, each with the delta of its half; the delta of a go-fabric tag on the adapter goes to the Fable reader), 1 Fable verifier | ≤ 2 |

Phase 0 of go-fabric's Review II spent about 10.3 M subagent tokens on a
module many times this size (its decision 24); no token estimate is given
here before phase 0 has measured one.

## Finding IDs

`UR-P<N>-<theme>-<nn>`. Themes:

| Theme | Covers |
| --- | --- |
| `auth` | the admin gate, token mode, host mode, what a role reaches |
| `web` | browser-side security: CSRF, XSS and unsafe rendering, framing, response headers, caching |
| `leak` | a secret or internal detail in a response, an event, a log line or the bundle |
| `conc` | races, locks, goroutine and subscription life cycle |
| `bound` | resource bounds: bodies, lists, streams, buffers, timeouts |
| `err` | error paths: fail open, wrong problem, swallowed error, nil port |
| `adpt` | the adapter and handlers against go-fabric's facade: wrong call, wrong encoding, a facade gap |
| `ctr` | contract, handlers, client and fixtures disagreeing |
| `ui` | page logic, state, i18n, accessibility |
| `api` | exported Go and npm surface, package cut |
| `test` | tests and guards |
| `sup` | workflows, dependencies, release, the committed bundle |
| `doc` | docs against code |
| `cite` | a citation that does not hold at its pin |

Status: open / fixed (#PR) / struck (reason) / accepted (decision number,
and where it is documented) / upstream (go-fabric issue).

## Done

Every finding down to note has one of the four closing statuses; every
defect class found has one standing guard in `make ci` or a workflow; a
convergence run over the then-current HEAD — with go-fabric's
post-Review-II tag pinned — yields no new high or medium; the held-back
rows are published with the pull request or release that carries their fix.

## Open questions for the owner

Each with the recommendation first; none is decided here.

1. **Is token mode meant to authenticate anybody?** The token is served by
   an unguarded read (`handlers_node.go:138`, `handlers_controller.go:320`)
   and the contract describes it as a cross-origin defence
   (`docs/contract.md:370-374`). *Recommendation:* record "whoever reaches
   the listener is an admin in token mode; restricting who reaches it is
   the host's job" as an explicit assumption of the threat model, let unit
   X3 measure what it exposes and whether `docs/embedding.md` says it
   plainly, and bring anything beyond a docs fix to the wizard.
   Alternative: treat it as a defect class from the start.
2. **Is disclosure to a non-admin in scope (unit X3)?** Every read and all
   three streams are unguarded in both modes (`handler.go:328`, `:346`,
   `:351-355`). *Recommendation:* yes — X3 lists per route what a non-admin
   learns, and the owner decides per class whether it is by design.
3. **Fuzz budget.** No target exists (`ci.yml:9-11`). *Recommendation:*
   phase 0 writes ad hoc targets for the hand-written parse sites
   (`parseHex`, `parseHex32`, `parseNodeID` at
   `handlers_controller.go:208-228`, `logWindow` at `operations.go:246`,
   `decodeJSON` at `problem.go:125`) and one HTTP-level target over the
   whole mux, five minutes each locally, committed only if one finds
   something; a standing fuzz job is decided in phase 3.
4. **Do Playwright runs count as evidence?** *Recommendation:* as the
   baseline's mechanical evidence, yes (`e2e-web` is a required check); as
   evidence of a security property, no — the smoke is seven tests. A role's
   exploit that needs a real browser is written as its own Playwright test
   and becomes a guard only when its fix lands.
5. **`npm audit`.** *Recommendation:* run in phase 0 as evidence; an
   advisory on a package in the bundle is a finding, one on a development
   dependency a note; whether a workflow runs it is decided in phase 3.
6. **The hosts as input.** How `Options.Admin` is really set decides how
   much X3 and X4 matter (`go-fabric-server/internal/daemon/web.go:36`,
   `go-ha2fabric/cmd/ha2fabric/main.go:112`). *Recommendation:* a read-only
   Sonnet inventory of both call sites in phase 0, verified by the threat
   model's author; what it finds in a host is noted for that host's review,
   not fixed here.
7. **A fix that changes the contract.** Gating a route that is open today
   turns a 200 into a 403, which the rule "a route is never removed in a
   minor version" does not cover. *Recommendation:* each goes to the
   wizard; those accepted are bundled in phase 3's pull request with a
   contract bump and a line in `docs/contract.md` for the hosts.
8. **Where held-back reports live.** *Recommendation:* an untracked
   directory of the review worktree listed in `.git/info/exclude` (no
   change to `.gitignore`), named in the phase's hand-over.
9. **Tags.** *Recommendation:* one release after phase 3's pull request
   and one after convergence if it changed code; a security fix that
   cannot wait gets a patch release of its own, the hosts' bump with it.
10. **Phase 1 under the cap of five.** It leaves two Opus readers with
    about 7 700 lines of Go and 7 300 lines of pages each, which costs
    depth. *Recommendation:* run phase 1 as two halves, 1a Go and 1b web,
    each a workflow of at most five agents in its own fresh session with
    its own pull request (as go-fabric decision 120 halved its phase 6),
    so that G-go and G-web are each read by two agents. Alternative: one
    workflow as cut above, with a second pass over whatever a reader
    reports unread.
