# Phase 2 — the dispositions

Branch `review/p2` from `1056db2`; the pull request is squashed, its single
commits stay retrievable through the pull request's reference. Sources:
`notes/briefs/2026-11-18-review-p2.md`, `notes/reads/read-review-p2-inventory.md`
(script), the reviewers' reports `phase-2-X1.md`, `phase-2-X3.md`,
`phase-2-X4.md` (Fable), `phase-2-X2.md` (Opus), `phase-2-verify.md`
(Fable), the fix reports `phase-2-fix-G.md`, `phase-2-fix-W.md` (Opus),
the pre-merge read `phase-2-premerge.md` (Fable, IDs 190–199),
`decisions.md` rows 33–36. Severity is the verifier's for every medium and
for the sampled lows (9 of 14, 0 struck, 1 lowered); the rest is the
reviewer's. The public copies of X4's report, of the verification and of
the pre-merge read leave out what concerns held rows.

**Held back:** three findings of X4 belong to the held rows of phase 0
(decision 18); they are folded into those rows' fix in phase 3 and have no
row here. The red tests of the held rows exist and stay out of the
repository until that pull request.

Every fix below has a test that was red before it (the commit and the fix
report name it; decision 17); the exceptions are said in the row. "→ phase
3" means the remaining fix changes the contract (decision 13) or belongs
to a phase 3 unit.

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P2-conc-04 | medium (verifier; X1 medium, X2 low as -32) | fixed `b3b5c05`, `f3b69e0` | `NodeSwitch.Set` ends the streams over the switch; the erase flow resyncs |
| UR-P2-conc-32 | — | merged into UR-P2-conc-04 | verifier |
| UR-P2-adpt-30 | medium | fixed `120843e` | a chunked list write answers its first refused element |
| UR-P2-conc-01 | low | fixed `40400f2` | a replaced window is `window_opened` again; `docs/contract.md` says a `window_opened` may follow one |
| UR-P2-err-02 | low | fixed `9f18287` | the fresh window told by `OpenedAt` |
| UR-P2-adpt-03 | low | fixed `120843e`, `6dcaf30` | a commissioning that reached its start and then failed is answered as done; the pre-merge read's UR-P2-adpt-190 narrowed it |
| UR-P2-ui-05 | low | fixed `8b94c01` | the flow takes its own window's kind from the status; the residue (a failed status read and a replacement before the next) → phase 3 with decision 28's field |
| UR-P2-ui-06 | note | open → phase 3 | the countdown needs a remaining-seconds field (contract) |
| UR-P2-conc-07 | note (inferred) | fixed `9f18287` | red test written first: two concurrent opens answered 200 and 200 |
| UR-P2-leak-08 | note | fixed `9f18287`, `bbce296` | credentials dropped when the window ends; the page drops codes once their flow ends |
| UR-P2-err-09 | note | upstream go-fabric | the commissioner does not disarm the fail-safe on a cancelled context; issue after this pull request |
| UR-P1-err-22 | low | fixed `a0cc661` | erase detached from the request, bounded, failure logged at error |
| UR-P2-err-31 | low | fixed `a0cc661` | registry Delete and decommission the same way |
| UR-P2-bound-33 | low | fixed `4870251`, `48bb394` | a write deadline of two heartbeats per stream round |
| UR-P2-conc-34 | low | docs `56d1682`; an `Options` field → phase 3 `api` | the host's shutdown of open streams |
| UR-P2-ui-35 | note | open → phase 3 | an exposure apply whose rebuild fails needs a field (contract) |
| UR-P2-adpt-36 | note (inferred) | fixed `b2e49f1` | red test written first: an undecodable NOCResponse answered 204 |
| UR-P2-err-37 | note | upstream go-fabric, then phase 3 | the facade exports neither error that tells a refused write from a node failure |
| UR-P2-auth-60 | note (verifier; X3 low) | open → phase 3 | decision 33: the cached ACL, bindings, fabrics and NOCs gated |
| UR-P2-leak-61 | note | open → phase 3 | decision 33: endpoint 0's NOCs and ACL events gated |
| UR-P2-doc-62 | low | open → phase 3, with UR-P0-doc-22 | what a false `Admin` serves, written once the gates of decisions 19, 20, 33–35 land |
| UR-P2-leak-63 | — | merged into UR-P1-leak-66 → phase 3 | verifier; its test becomes row 66's red test |
| UR-P2-ctr-64 | note | open → phase 3 | node ids of 2^63 and more printed negative by the diagnostics (contract) |
| UR-P2-bound-65 | note | docs `56d1682` | the status route's host ports must be cheap |
| UR-P2-ui-90 | low | fixed `4d00f0e`, `ee9105b` | every keyed list keys by position or a key that cannot repeat; a standing test over every keyed `{#each}` |
| UR-P2-web-91 | low | fixed `bd232da`, `027375b` | decision 36: an unknown path is 404; changelog line for the hosts |
| UR-P2-test-95 | note | fixed `a8e1183`, `e5d289a` | guards: stream data on one line, no CORS on any route, framing and cross-origin reads in Chromium |
| UR-P2-adpt-190 | medium | fixed `6dcaf30` | pre-merge read: adpt-03's first fix answered 200 to a second commissioning of a commissioned device |
| UR-P2-ui-191 | low | fixed `0a4d51f` | pre-merge read: an endpoint's cluster panels reopened after a reload |
| UR-P2-test-192 | note | fixed `6dcaf30` | pre-merge read: a test comment said more than its test |
| UR-P2-doc-193 | note | fixed `6dcaf30` | pre-merge read: a shifted citation and a claim limited to the switch as the node source |

Decisions 33–35 (decision 8 per class) gate further routes in phase 3; no
row of their own beyond auth-60 and leak-61.

Counts: 29 findings (25 of the reviewers after two merges and three folds
into held rows, plus 4 of the pre-merge read) — 0 high, 3 medium, 12 low,
14 note. Fixed 19 (one with a part → phase 3), docs only 2 (one with a
part → phase 3), open 6 (→ phase 3), upstream 2, struck 0.

## Measured cost

| Agent | Model | Tokens | Wall clock |
| --- | --- | --- | --- |
| inventory | script | — | — |
| X1 | Fable | 325 k | 20 min |
| X2 | Opus | 262 k | 15 min |
| X3 | Fable | 306 k | 17 min |
| X4 | Fable | 360 k | 36 min |
| verifier | Fable | 194 k | 7 min |
| fix G (with the follow-up) | Opus | 506 k | 33 min |
| fix W (with two follow-ups) | Opus | 229 k | 53 min |
| pre-merge read | Fable | 195 k | 12 min |
| **sum** | | **2 377 k** | |
