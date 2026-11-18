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
