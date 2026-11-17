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

`decisions.md` carries them as rows 1–5.

| Decision | Choice |
| --- | --- |
| Scope | go-fabric-ui, then go-fabric-server, then go-ha2fabric, one after the other; this plan is go-fabric-ui only. The Loom repositories are excluded |
| Cut | **Adapted, five phases**: 0 basis, 1 package pass, 2 flows and adversarial, 3 architecture, tests, docs and supply chain, 4 convergence. No matter.js parity phase; citations of go-fabric, matter.js and the reference pages stay checked as `CLAUDE.md` has it |
| Baseline | origin/main **now** (`25f6e7b4`); feature work (U12, #68) continues, each phase re-reads what landed as delta |
| Models | As go-fabric decision 24 and this repository's model table: **Opus by default** for reviews and implementation; **Fable** for the security paths' review (admin and token paths, auth, the streams, whatever renders untrusted data), the verification of every high and medium finding, every reference citation, and planning; **Sonnet or a script** for inventories |
| Process | As go-fabric decisions 20, 25, 27: review and verification per phase as a workflow (inventory → reviewer per theme → an independent verifier per finding), **at most ten agents per phase** unless the owner agrees; **one bundled fix pull request per phase**, squash, auto-merge on green CI; open security items held back from the public notes until their fix merges |
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
| `ui/src/lib/api/schema.gen.ts` (generated), `i18n/{en,de}.json` | 3 | 5 880 + 1 500 | generated: guard only; catalogues: phase 1 unit F by key, not by line |
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
| 2 Flows and adversarial | End to end across handler → `gofabric` adapter → go-fabric's facade at `v0.22.2`, and page → client → handler; one reviewer per attacker role, exploits written as failing tests (table below) |
| 3 Architecture, tests, docs, supply chain | Contract and OpenAPI against handlers, client and fixtures; the exported `fabricui` API (`Options`, the ports) and the import cut; test quality on both halves; every behavioural claim in `docs/` and every citation against code and pins; workflows, release, npm publish, the committed bundle |
| 4 Convergence | A full-cut run over the HEAD of the day, repeated until one yields no new high or medium. Includes the delta of what feature work landed and the delta when go-fabric's post-Review-II tag is pinned in `go.mod` |

### Phase 1 units

| Unit | Files | Lines (about) | Model |
| --- | --- | --- | --- |
| A gate, streams, SPA | `fabricui/handler.go`, `options.go`, `problem.go`, `stream.go`, `spa.go`, `events.go`, `ports.go`, `registry.go`, `nodeswitch.go`, `operations.go`, `doc.go` | 2 700 | **Fable** (the admin and token paths, the streams) |
| B node family | `fabricui/handlers_node.go`, `handlers_pairing.go`, `handlers_exposure.go`, `handlers_diagnostics.go`, `handlers_topology.go`, `handlers_surface.go`, `node_detail.go`, `provider.go`, `devicetype_names.go` | 2 400 | Opus; Fable overlay on `handlers_pairing.go` and the secret-bearing parts of `handlers_node.go` |
| C controller family | `fabricui/handlers_controller.go`, `controller.go`, `memory_controller.go` | 4 000 | Opus; Fable overlay on commissioning, credentials and the window routes |
| D controller adapter | `fabricui/gofabric/*.go` (6 production files) | 2 300 | Opus; Fable overlay: behaviour against go-fabric's controller facade at `v0.22.2`, and every go-fabric citation in the package |
| E client, stores, shell, theme | `ui/src/lib/api/client.ts`, `stream.ts`, `ui/src/lib/context.ts`, `store.svelte.ts`, `controller-store.svelte.ts`, `nodes-store.svelte.ts`, `qr.ts`, `scan/`, `i18n/index.ts`, `theme/ha-bridge.ts`, `theme/tokens.css`, `ui/src/shell/*` | 2 700 | **Fable** (token handling, the streams, the parent-document access of the theme bridge, the prepaint script of `shell/index.html`) |
| F node and operation pages, kit | `ui/src/lib/pages/node/*`, `pages/operation/*`, `ui/src/lib/kit/*`, the catalogues by key | 4 600 | Opus; whatever renders a host-, node- or log-supplied string is listed for the Fable overlay |
| G controller pages | `ui/src/lib/pages/controller/**` | 2 700 | Opus; same listing for the overlay |

The three Fable overlays (B, C and the render lists of F and G) are one
bounded agent: the lines named, the question, a fact note back.

### Phase 2 flows and roles

| Unit | Content | Model |
| --- | --- | --- |
| F1 pairing | Window on fresh or configured credentials → the ephemeral record → `GET /node/pairing` → the stream's window events → the Pair page's QR and countdown; close, expiry, a second node, a `NodeSwitch` swap, erase and factory reset | Fable |
| F2 commissioning | Commission page → `POST /controller/commission` → adapter → go-fabric's commissioner; the credential store, BLE decision, terms retry, progress events on the controller stream, abort and context end | Fable |
| F3 node writes | ACL, bindings, attributes, commands, fabric removal, ICD, OTA consent, exposure apply → reassemble: request decoding → adapter encoding → the facade call → the error's problem → what the page shows | Opus |
| F4 life cycle | Stream subscribe and unsubscribe against `node.Observe`, slow and vanished clients, goroutine and observer leaks, registry create, rename, delete against the handler cache, shutdown of the host | Opus |
| R1 unauthenticated client on the listener's network | What it reads and what it can drive, in token mode and in host mode | Fable |
| R2 non-admin viewer | Exists only in host mode, where `Options.Admin` answers false: what the unguarded reads and streams disclose | Fable |
| R3 malicious page or origin | Forged writes, cross-origin reads of status and streams, DNS rebinding, framing, a hostile parent or sibling frame under an Ingress | Fable |
| R4 hostile node and host data | Names, labels, attribute and event values, discovery results, mDNS records, log records, the host's warnings and counters, problem details — through JSON, the SSE framing and every page that renders them | Fable |

### Phase 3 units

| Unit | Content | Model |
| --- | --- | --- |
| contract | `contract/openapi.yaml` (87 paths, 97 operations, 48 `x-admin` marks) against the hand-written routes, the generated ones, `ui/src/lib/api/client.ts`, the fixtures and `docs/contract.md`; `ui/scripts/contract-guard.mjs` and `contract/previous/` | Opus |
| api | The exported surface of `fabricui` and `fabricui/gofabric` (`Options`, `NodeSource`, `ControllerSource`, `ExposureProvider`, `NodeRegistry`, `Operations`, `EventSource`, `NodeSwitch`), the npm package's exports (`ui/package.json`), `internal/arch` | Opus |
| test-go | Tests that assert nothing, missing negative cases, the doubles against the real facade, the `gofabric` floor of 50, flakes, the e2e tests | Opus |
| test-web | vitest (32 files), the Playwright smoke (`ui/e2e/smoke.spec.ts`, 7 tests), the i18n coverage tests, the happy-dom limits | Opus |
| docs | Every behavioural claim in `README.md`, `CONCEPT.md`, `docs/*.md`, `ui/README.md` and `changelog.md` against the code | Opus |
| cite | Every citation of go-fabric (at the tag the sentence names and at `v0.22.2`), openccu-loom `b40b289a` and matter.js `862ad75f` in code and docs: a script and Opus check every range, Fable reads each that is the reason for behaviour, each doubtful one and a sample of the rest | Opus + Fable |
| supply | `.github/workflows/*` (permissions, pins, `pull_request_target`), `release-on-tag.yml` and the npm publish, Dependabot and its auto-merge, branch protection against the jobs (ten required contexts on 2026-11-17, `gh api`), `go.mod`'s indirect set, npm licences and `THIRD-PARTY-NOTICES.md`, the committed bundle and `dist-check` | Opus |

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
3. **Verify.** An independent Fable verifier re-reads every high and
   medium against the code and the pins and tries to refute it — struck
   with a reason, or confirmed; lows and notes are verified in batches by
   Opus.
4. **Fix.** Branch `review/p<N>`; the regression test first (red, then
   green); one standing guard per defect class; the local guard set of
   `CLAUDE.md` in the background before the pull request; one pull request
   per phase, squash, auto-merge on green CI. Opus implements; a fix on
   the admin and token paths gets the Fable read before it merges
   (`CLAUDE.md`'s model table). Decisions go to the owner as a wizard, the
   recommendation first. A facade gap becomes a go-fabric issue, not a
   workaround. A confirmed vulnerability is not committed to the register
   before its fix merges.

The coordinator is Opus in a fresh session per phase; push, pull request
and tag are that session's own calls.

## Agents per phase (estimates)

| Phase | Agents | Cut |
| --- | --- | --- |
| 0 | 5 | 1 mechanical evidence (Opus, scripts), 1 inventory (Sonnet), 1 threat model (Fable), 1 fuzz and error-path coverage (Opus), 1 verifier (Fable) |
| 1 | 10 | 7 unit reviewers (2 Fable, 5 Opus), 1 Fable overlay, 2 verifiers (1 Fable for high and medium, 1 Opus for the rest); inventories by script |
| 2 | 10 | 4 flow reviewers (2 Fable, 2 Opus), 4 role reviewers (Fable), 2 Fable verifiers |
| 3 | 10 | 7 unit reviewers (Opus), 1 Fable citation read, 2 verifiers (1 Fable, 1 Opus) |
| 4 | 6 per run | 2 Fable (Go and web security paths), 2 Opus (the rest of each half), 1 delta reader (Opus; Fable where the delta is on a security path or follows a go-fabric tag), 1 Fable verifier |

Fix agents are not counted; a phase's fix wave is hand-led. Phase 0 of
go-fabric's Review II spent about 10.3 M subagent tokens on a module many
times this size (its decision 24); no token estimate is given here before
phase 0 has measured one.

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
   the host's job" as an explicit assumption of the threat model, let role
   R1 measure what it exposes and whether `docs/embedding.md` says it
   plainly, and bring anything beyond a docs fix to the wizard.
   Alternative: treat it as a defect class from the start.
2. **Is disclosure to a non-admin in scope (role R2)?** Every read and all
   three streams are unguarded in both modes (`handler.go:328`, `:346`,
   `:351-355`). *Recommendation:* yes — R2 lists per route what a non-admin
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
   much R1–R3 matter (`go-fabric-server/internal/daemon/web.go:36`,
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
10. **Phase 1 and 2 at ten agents each.** Both sit at the limit of
    decision 5. *Recommendation:* approve as cut; a second pass, if a unit
    leaves files unread, is asked for separately.
