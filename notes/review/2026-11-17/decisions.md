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
