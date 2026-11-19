# Owner decisions, whole-codebase review

Each put as a wizard step with the recommendation first; the owner's choice
is the row. IDs are `UR-…` (`README.md`, "Finding IDs"). Decisions of
go-fabric's Review II are cited as "go-fabric decision n"
(`go-fabric/notes/review/2026-11-13/decisions.md`).

## 2026-11-17, before phase 0

| # | Subject | Decision |
| --- | --- | --- |
| 1 | Scope | **go-fabric-ui, then go-fabric-server, then go-ha2fabric, sequentially**; this review and its plan are go-fabric-ui only. The Loom repositories are excluded |
| 2 | Cut | **Adapted, five phases.** 0 Basis: the baseline pinned; mechanical evidence (`make race`, govulncheck, fuzzing where input is parsed, coverage including error paths, `npm audit`, svelte-check, Playwright); a threat model for the HTTP handler, the admin and token paths, the SSE stream, the SPA (XSS, CSRF) and the embedding host boundary. 1 Package pass: every hand-written Go package and the `ui/` source. 2 Flows and adversarial: end-to-end flows across handler → `gofabric` adapter → go-fabric's facade; attacker roles such as an unauthenticated LAN client, an authenticated low-privilege viewer if one exists, a malicious page or origin, hostile node data rendered in the pages. 3 Architecture, tests, docs, supply chain: contract and OpenAPI, the `fabricui.Options` API, test quality, docs against code, workflows, release, the committed bundle. 4 Convergence: repeated until a run yields no new high or medium, and including the delta when go-fabric's post-Review-II tag is pinned. **No matter.js parity phase**; where the pages mirror go-fabric, matter.js or the reference pages, the citations stay checked as `CLAUDE.md` has it |
| 3 | Baseline | **origin/main now**, fetched and recorded: `25f6e7b4ad743d25932406b88e386d945f47c54f`. The open branch `wave/U12-onboarding-inventory` (issue #68) is feature work that continues; each phase re-reads what landed as delta |
| 4 | Models | **As go-fabric decision 24 and this repository's model table**: Opus is the default for reviews and implementation; Fable for the review of the security paths (the admin and token paths, authentication, the stream, anything that renders untrusted data), the verification of every high and medium finding, every reference citation, and planning; Sonnet or a script for inventories |
| 5 | Process | **As go-fabric decisions 20, 25 and 27**: review and verification per phase as a workflow (inventory → a reviewer per theme → an independent verifier per finding); at most ten agents per phase unless the owner agrees; one bundled fix pull request per phase, squash, auto-merge on green CI; open security items are held back from the public notes until the pull request that fixes them |
| 6 | Agent caps and sessions (this machine is slower) | **At most five agents per phase for review and verification** — replaces the ten of decision 5 — **and at most two for implementation** (a fix wave). Themes are merged and findings verified in batches, not one verifier per finding. **One fresh session per phase.** The plan for go-fabric-server is written while go-fabric-ui is in phase 4 |
| 7 | Token mode (open question 1) | **An assumption of the threat model**: whoever reaches the listener is an admin in token mode; restricting who reaches it is the host's job. Unit X3 measures what it exposes and whether `docs/embedding.md` says it plainly; anything beyond a docs fix goes to the wizard |
| 8 | Disclosure to a non-admin (open question 2) | **In scope**: X3 lists per route and stream what a non-admin learns; the owner decides per class |
| 9 | Fuzz budget (open question 3) | **Ad hoc targets in phase 0** for the hand-written parse sites and one HTTP-level target over the mux, five minutes each locally, committed only on a hit; a standing fuzz job is decided in phase 3 |
| 10 | Playwright (open question 4) | **Baseline evidence, not proof of a security property**; a browser-only exploit gets its own Playwright test, a guard once its fix lands |
| 11 | `npm audit` (open question 5) | **Runs in phase 0**; an advisory on a bundled package is a finding, on a development dependency a note; a workflow job is decided in phase 3 |
| 12 | The hosts as input (open question 6) | **A read-only Sonnet inventory** of how go-fabric-server and go-ha2fabric set `Options` in phase 0, verified by the threat model's author; host findings go to those repositories' reviews |
| 13 | Fixes that change the contract (open question 7) | **Each to the wizard; those accepted are bundled in phase 3's pull request** with a contract bump and a line in `docs/contract.md` for the hosts |
| 14 | Held-back reports (open question 8) | **An untracked directory of the review worktree** listed in `.git/info/exclude`, named in the phase's hand-over |
| 15 | Tags (open question 9) | **One release after phase 3's pull request, one after convergence if it changed code**; a security fix that cannot wait gets its own patch release with the hosts' bump |
| 16 | Phase 1 under the cap (open question 10) | **Two halves**: 1a Go and 1b web, each its own workflow of at most five agents, its own fresh session, its own branch (`review/p1a`, `review/p1b`) and pull request; G-go and G-web are each read by two Opus agents |
| 17 | Verification of low and note | **The regression-test rule plus a sample**: a low or note is verified by its fix's regression test (red before the fix; one that cannot be made red is struck, a docs or style finding without a test is checked against the code by the implementer and struck with a reason if wrong), **and** the phase's Fable verifier re-reads a sample of the lows — about one in five per unit, at least two — so the false-positive rate of the Opus reviewers is measured |

## 2026-11-17, after phase 0's verification

Put as one wizard of four questions, the recommendation first. The two held
rows are named by ID only until their fix merges (decisions 5 and 14).

| # | Subject | Decision |
| --- | --- | --- |
| 18 | The held rows UR-P0-auth-20 and UR-P0-web-21, and the verifier's observation next to 20 | **Bundled in phase 3's pull request** with the other contract changes (decision 13), not a patch release of their own (the recommendation was a phase 0 patch under decision 15). The rows stay held until that pull request; UR-P0-doc-22 waits with them, because its text depends on their fix |
| 19 | UR-P0-auth-26, the four ungated controller routes that make the controller read from a node | **All four gated** (`x-admin` on ACL, bindings, fabrics and `POST …/attributes`), as `refresh=true` already is; a non-admin no longer sees those tabs. Contract change, phase 3 |
| 20 | UR-P0-leak-27, the daemon's log to everyone | **Gated**: `x-admin` on `GET /operation/log` and the log stream; the client reads the log stream through `fetch` with the token header instead of `EventSource`. Contract and web change, phase 3 |
| 21 | UR-P0-web-23, framing | **`frame-ancestors 'self'` by default plus an `Options` field** for further parent origins of a host that frames the page from elsewhere. No contract field; lands in phase 0's fix wave |

## 2026-11-17, the fix wave's questions

| # | Subject | Decision |
| --- | --- | --- |
| 22 | `Options.FrameAncestors` and an entry that admits any parent | **Refused**: `New` and `Mount` fail on `*`, a host of `*` with any scheme or port, and every scheme-source (`https:`); only a named origin (also `*.domain`) passes |
| 23 | `Mount` as one pattern `Prefix/` (UR-P0-api-32): a host route registered under the prefix after `Mount` shadows the API's | **Accepted** as implemented and documented; both pinned hosts use `Handler`, not `Mount` |
| 24 | Exporting the CSP builder for a host that serves the page itself beside `Mount` | **Not now**; phase 3's api unit takes it up if a host needs it |

## 2026-11-18, after phase 1a's verification

| # | Subject | Decision |
| --- | --- | --- |
| 25 | One held row of phase 1a (medium; no contract change) | **Bundled in phase 3's pull request** with the held rows of decision 18, not fixed in 1a (the recommendation was the 1a pull request). The row stays held until then; the go-fabric issue for the facade gap it shows is filed after that pull request merges |
| 26 | UR-P1-doc-41, UR-P1-doc-69: `CLAUDE.md` and the adapter's package doc name fewer allowed imports than `internal/arch` | **The docs follow the guard**: `CLAUDE.md` and the package doc list what `internal/arch/arch_test.go` allows; no code change |

## 2026-11-18, after phase 1b's verification

| # | Subject | Decision |
| --- | --- | --- |
| 27 | UR-P1-ctr-101 (high; with UR-P1-ctr-162): 64-bit attribute values rounded by the client's parse and written back by the Attributes tab | **Lossless parse in 1b**, no contract change: the client keeps integers outside the safe range exactly, the write sends the edited text as typed, and a value that would be rounded is refused. Not held back: an administrator's own data error, not an attacker's path |
| 28 | The page part of UR-P1-err-04 | **A notice in 1b, the field in phase 3**: when a freshly read `window_kind` says the open window is not the page's own and the pairing answers `source: configured`, the Pair page shows the codes with a notice that the window was opened elsewhere and they may not apply, and offers Close; the exact answer is a contract field for phase 3's wizard |
| 29 | UR-P1-bound-103: two or three streams per tab whatever page is shown | **Streams on demand**, no contract change: a family's stream only while one of its pages is shown, closed while the tab is hidden and reopened with a reload when it is shown again; together with UR-P1-conc-102's error and open hooks |
| 30 | go-fabric v0.22.2 reports an administrator's basic window as `WindowEnhanced` | **A go-fabric issue after the 1b pull request merges**; not held back |
| 31 | UR-P1-ui-172: under the HA skin, which the standalone shell uses too, Home Assistant's warning, success and primary colours fail AA as small text | **Text-safe variants**: the HA colours stay for areas and icons; status text uses `--fui-*-text` variants of at least 4.5:1, which a host's theme under an Ingress panel still overrides. No contract change, in 1b |
| 32 | UR-P1-ui-142: the kit's search ran the user's term as a regular expression | **Kept as fixed**: literal text, `\|` separates alternatives; a `changelog.md` line tells the hosts |

## 2026-11-18, after phase 2's verification

Put as one wizard of four questions, the recommendation first. Rows 33–35
answer decision 8 per class (classes as in `phase-2-X3.md` section 6); every
gate is a contract change and lands in phase 3's pull request (decision 13).

| # | Subject | Decision |
| --- | --- | --- |
| 33 | Decision 8, classes I-b, L and C (part): the controller's cached ACL, bindings, fabrics and NOCs of every node (UR-P2-auth-60), the credential names of `GET /controller/credentials`, endpoint 0's NOCs and ACL events on the node | **All three gated** (`x-admin`), as decision 19 did for the live reads. Phase 3 |
| 34 | Decision 8, class A-b: `status.config.store` and the text of the host's warnings | **Withheld from a non-admin**: the status leaves out `config.store` and gives each warning's kind only. Phase 3 |
| 35 | Decision 8, classes B, F, H, J, D and K: fabrics and sessions, the exposure item list, the controller inventory, diagnostics and topology, the node and controller traces | **Gated, everything but the status** (the recommendation was open and documented, the traces with decision 20): a non-admin keeps the node status, the node list and the operation info; the read-only page shows little else. Phase 3, with rows 19, 20 and 33 |
| 36 | UR-P2-web-91: every unknown path is answered with the index, whose relative asset paths then fail | **404 for an unknown path**: only the prefix root serves the index (the page routes by hash); a `changelog.md` line for the hosts, `docs/embedding.md` drops "every deep path". In phase 2 |

## 2026-11-18, before phase 3

Put as one wizard of eight questions in three steps, the recommendation
first; the owner took every recommendation. The inventory of what phase 3
carries is `notes/reads/read-review-p3-carryover.md` (46 rows). The two
held rows are named by ID only until their fix merges (decisions 5, 14,
18); the shape of their fix is in the review worktree's `held/`.

| # | Subject | Decision |
| --- | --- | --- |
| 37 | The held row UR-P0-auth-20, the verifier's observation next to it and one held finding of phase 2 | **The recommended shape**, in phase 3's pull request: one check for every mode and every route, a new `Options` field, a new 403. Held until then |
| 38 | The held row UR-P0-web-21 and two held findings of phase 2 | **The recommended shape**, in phase 3's pull request: three handler-side checks, a new 415 and a new 403, a request header a host's own client must send. Held until then |
| 39 | Order of phase 3's two jobs | **Reviews first, then one fix wave**: the four review units and the verifier read `f4b8573`; one fix wave (at most two agents) then implements the carried rows and the new findings together, with one contract bump |
| 40 | Decision 8, classes C and E and M (not named by decision 35) | **C and E gated, M open**: the node's endpoints, attribute values, surface and mDNS become `x-admin` (decision 33's endpoint-0 rule is subsumed); `window_opened` stays readable |
| 41 | The streams under decisions 20, 33–35 and 40 | **The node stream filtered, the controller stream gated**: a non-admin receives on the node stream only the event types of the classes it may read (status, `nodes_changed`, `window_opened`); host events from `Options.Events` are admin-only; the controller stream and the log stream are `x-admin`. One allow-list of event types per class, held by a guard test |
| 42 | UR-P0-bound-29 | **A per-handler cap on open streams**, answered 503 with `Retry-After`, the cap an `Options` field with a default |
| 43 | UR-P0-ctr-28, UR-P1-ctr-17, -ctr-13, -err-33, -err-36, -adpt-35, -ctr-32 | **All land**: strict bodies everywhere (unknown fields and trailing data are 400; factory reset and erase stop discarding the decode error); 405 with `Allow`; an unknown fabric index 404 on the node route (the controller route after go-fabric exports the error); compatibility 500 on a store error; endpoint `maximum: 65534`; `unknown` in the ACL enums |
| 44 | Additive fields | **All four groups land**: the pairing window's own/other answer (decision 28, UR-P2-ui-05) and the seconds left (UR-P2-ui-06); 64-bit values as strings (UR-P1-ctr-113, -177, UR-P2-ctr-64) and a hex pattern on `NodeID` (UR-P1-web-115); `admin` on the controller status (UR-P1-ui-133, -196), `auto_subscribe` (UR-P1-ctr-168), a commissioning id on progress events (UR-P1-api-18); `fabric_count` absent on a failed read (UR-P1-err-07), a per-endpoint status on bindings (UR-P1-bound-37), a failed rebuild on exposure apply (UR-P2-ui-35) |

## 2026-11-18, after phase 3's verification

Put as one wizard of twelve questions in three steps, the recommendation
first; the owner took every recommendation. Sources: the four unit reports
and the verification of phase 3 (published with phase 3's pull request).

| # | Subject | Decision |
| --- | --- | --- |
| 45 | UR-P3-ctr-04: the streams and the admin test | **One admin test for every route, the streams included; the client reads all three streams through `fetch`**; the npm package gains an additive `fetch`-based stream factory, `EventSourceFactory` stays |
| 46 | UR-P3-test-03: shapes the contract guard cannot see as breaking | **Additive twins, optional fields**: a 64-bit value gets a string field beside the old number (the number deprecated until a major); `Warning.message` stays required and is empty for a non-admin; `fabric_count` becomes optional with a changelog line; the guard learns required→optional, type changes and tighter patterns and maximums |
| 47 | UR-P3-ctr-01: undeclared statuses | **Every status per operation, through shared response components**, and a Go guard that every status a route answers is declared (extending the routes-and-contract test of unit A) |
| 48 | UR-P3-api-05: panicking constructors | **`NewHandler(o) (http.Handler, error)`** and an error-returning `Mount` variant; the panicking ones stay, documented; the godocs corrected |
| 49 | UR-P3-doc-72: `Prefix` with `Handler` | **`Handler` refuses a non-default `Prefix`**; `Prefix` stays for `Mount`; `docs/embedding.md` says so |
| 50 | UR-P1-web-20: a wildcard over a public suffix | **A docs and godoc warning only**; decision 22 stands, no public-suffix list |
| 51 | UR-P2-conc-34: the host's shutdown | **`Options.Done <-chan struct{}`** ends every stream and its observers; the documented base-context pattern stays as an alternative |
| 52 | Handler corrections (decision 13): UR-P1-adpt-40, -ui-178, -ui-179, UR-P3-adpt-91 | **All land**: the fabric label is refused over 32 bytes (400; a go-fabric question, its controller counts runes); a discovered device's invalid passcode is 400 in the handler and refused by the page; the client-state errors of the controller facade are 409 (declared where missing); the ACL add form drops `pase`, display and enum keep it |
| 53 | Decision 9: a standing fuzz job | **Seeds and a nightly job**: the eight phase 0 targets as `Fuzz*` functions with seed corpora, run by every `go test`, extended to the shapes of decisions 37, 38, 43, 44; `make fuzz` uses `FUZZTIME`; a nightly job at about three minutes per target |
| 54 | Decision 11, UR-P3-sup-70: `npm audit` | **A failing `npm audit --omit=dev --audit-level=low` in the web job**, a full audit in the nightly job that reports without failing |
| 55 | UR-P3-sup-69: the generator's requirements | **A separate tools module** for oapi-codegen; the module's `go.mod` keeps only runtime dependencies |
| 56 | Repository settings (UR-P3-sup-61, -63, -64, -66) | **All four set**: the six unbound required checks bound to GitHub Actions; a ruleset on `v*` tags (and the release checks the tag is on `main`); an `npm` environment for the publish job; CodeQL for the web half as a required context. Settings, not code: applied by the owner's session with the owner's confirmation |

## 2026-11-19, the contract design note's open points

Put as one wizard of four questions, the recommendation first; the owner
took every recommendation. Source: `notes/briefs/2026-11-18-review-p3-contract.md`
"Open points" (Fable).

| # | Subject | Decision |
| --- | --- | --- |
| 57 | OPEN-1: `GET /controller/status` in class H (decision 35) carries `admin` (decision 44) and the token | **Open, the rest zeroed**: a non-admin gets `enabled`, `capabilities`, `admin` and the token as they are, every other required member empty or 0 (decision 46's pattern); no guard break |
| 58 | OPEN-2/3: `GET /operation/log/level` and the `window_closed` event, named by no decision | **Both open**: the log level route keeps no `x-admin`; `window_closed` joins `window_opened` and `nodes_changed` on the node stream's non-admin allow-list |
| 59 | OPEN-4: the commissioning id of UR-P1-api-18 | **Chosen by the client**: an optional `CommissionRequest.commissioning_id`, echoed on the progress events; additive |
| 60 | OPEN-5/6: 405, the unknown-path 404, and decision 48's `Mount` variant | **At path level**: a `MethodNotAllowed` component, a line in `info.description` and a Go guard on `Allow`, not per operation; `Mount` already returns an error, so only `NewHandler(o) (http.Handler, error)` is new |

## 2026-11-19, the fix wave's first measurement

| # | Subject | Decision |
| --- | --- | --- |
| 61 | Held row UR-P0-auth-20 (decision 37): a measurement of the pinned hosts' proxy contradicts one premise of the decided shape (detail in `held/`) | **A per-request predicate beside the name list**: a new `Options` field lets a host admit a request by its own trusted signal; the check stays in every mode and on every route. Fable fixes name and signature in the held note. Held until the fix merges |
| 62 | Owner rule, 2026-11-19: deprecated go-fabric identifiers | **None is used**, also not in tests. Held by staticcheck SA1019 in `make lint` (enabled, not excluded; clean at the v0.22.2 pin). Phase 4's lift to go-fabric v0.23.0 migrates every identifier v0.23.0 deprecates in the same change, by the migrations in go-fabric's `docs/guides/migrating-to-the-facade.md`; no `//nolint:staticcheck` for SA1019 |

## 2026-11-19, the pre-merge read

| # | Subject | Decision |
| --- | --- | --- |
| 63 | UR-P3-web-195: `attestation_error` on the node status carries the loader's error text to a non-admin, the class decision 34 withheld | **A fixed sentence for a non-admin**: the field stays present when the node refuses commissioning, its text a fixed neutral sentence; an administrator gets the loader's text. Additive, no guard break; a `docs/contract.md` line |
