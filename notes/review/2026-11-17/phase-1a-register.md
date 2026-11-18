# Phase 1a — the dispositions

Branch `review/p1a` from `49cc081`; the pull request is squashed, its single
commits stay retrievable through the pull request's reference. Sources:
`notes/briefs/2026-11-18-review-p1a.md`,
`notes/reads/read-review-p1a-inventory.md` (script), the reviewers' reports
`phase-1a-S-go.md` (Fable), `phase-1a-G-go-A.md`, `phase-1a-G-go-B.md`
(Opus), `phase-1a-verify.md` (Fable), `decisions.md` rows 25–26. The Fable pre-merge read of the fixes on the security paths
(plan step 4) found nothing that blocks and added IDs 19–22. Severity is
the verifier's for every medium and for the sampled lows (6 of 20, 0
struck); the rest is the reviewer's.

**1 row held back** (decisions 5, 14 and 25): one medium, in
`G-go-B`'s range. Its report lives in the review worktree's git-excluded
`held/` directory and is published with phase 3's pull request, which fixes
it.

Every fix below has a regression test that was red before it (the commit
names the test; decision 17). "→ phase 3" means the remaining fix changes
the contract (decision 13) and goes to phase 3's wizard.

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| (held) | medium | open → phase 3 | decision 25 |
| UR-P1-conc-01 | medium | fixed `b06f171` | `HEAD` on the node, log and controller streams answers the headers and subscribes nothing |
| UR-P1-adpt-30 | medium | fixed `8aa0194`, `ca0bf61` | the trace is served as go-fabric gives it, newest first; port doc, adapter comment and memory controller say and do the same |
| UR-P1-err-31 | low (verifier; reviewer medium) | fixed `2aae95c` | a TLV null for a non-nullable read, and own fabric index 0, are node errors (502); the own-fabric guard holds. The adapter's `null` branch is unreachable at `v0.22.2` and stays |
| UR-P1-conc-19 | low | fixed `5b15b22` | from the pre-merge read: a registry entry whose Source, Exposure or Events change under the same key left the replaced handler's streams open |
| UR-P1-api-02 | low | fixed `8e69fbd` | |
| UR-P1-err-03 | low | fixed `8e69fbd` | |
| UR-P1-err-04 | low | open → phase 1b / 3 | the page reading the window's kind is 1b; a field is a contract change |
| UR-P1-conc-05 | low | fixed `65d0b12` | |
| UR-P1-conc-06 | low | fixed `ab581a6` | |
| UR-P1-err-07 | low | log part fixed `12e7c6e`; optional `fabric_count` → phase 3 | |
| UR-P1-err-08 | low | fixed `51d7f79` | factory reset runs detached from the request, bounded at two minutes |
| UR-P1-leak-09 | low | open → phase 3 | with UR-P0-leak-27/-31 (decision 20) |
| UR-P1-err-10 | low | fixed `a075a4e` | 400 outside 0..4095 |
| UR-P1-conc-61 | low | fixed `e7108b6`, `06d0cfb` | one label push worker per source; a write during a push gets one more round |
| UR-P1-conc-62 | low | fixed `e7108b6` | `Source.Close`; `Hooks.Close` waits for a running callback |
| UR-P1-err-63 | low | fixed `ca0bf61`, `2a57071` | a vendor or product mismatch is 400; the memory controller refuses it too |
| UR-P1-bound-64 | low | fixed `4814bf8` | the discovery cache holds the last discovery for two minutes |
| UR-P1-adpt-65 | low | open → phase 3 | needs a presence flag on the port's discovered device; bundled with phase 3's commissioning changes |
| UR-P1-ctr-32 | low | 502 part fixed `e791329`; enum `unknown` → phase 3 | |
| UR-P1-err-33 | low | open → phase 3 / upstream | node route 404 is a contract change; the controller route needs go-fabric's fabric-not-found error, which the facade does not export (issue with phase 3) |
| UR-P1-adpt-34 | low | fixed `e791329` | |
| UR-P1-adpt-35 | low | fixed `a92a4e8` (no-contract part) | 65535 answers the unknown endpoint's empty list; 400 or `maximum: 65534` → phase 3 |
| UR-P1-err-36 | low | log part fixed `12e7c6e`; a 500 → phase 3 | |
| UR-P1-bound-37 | low | fixed `b80bd20` (no-contract part) | chunked read, a refused endpoint is 502; a per-endpoint status → phase 3 |
| UR-P1-web-20 | note | open → phase 3 api | from the pre-merge read: `FrameAncestors` admits a wildcard over a public suffix (`https://*.co.uk`); refusing it needs a suffix list |
| UR-P1-conc-21 | note | open → phase 3 api | from the pre-merge read: on the path a deletion overtook, `NodeRegistry.List` is called under the handler's registry lock; the port's doc does not forbid a call back into the handler |
| UR-P1-err-22 | note | open → phase 2 X2 | from the pre-merge read: erase still runs on the request's context, unlike factory reset (`51d7f79`) |
| UR-P1-conc-11 | note | fixed `6fd086b` | |
| UR-P1-err-12 | note | fixed `d8f19ce` | still 500, logged at debug |
| UR-P1-ctr-13 | note | open → phase 3 | 405 is a new status |
| UR-P1-bound-14 | note | fixed `a2669d1` | |
| UR-P1-doc-15 | note | fixed `a2669d1` | comment checked against the code |
| UR-P1-web-16 | note | fixed `b8ca059`, `4fcdeaa` | extends decision 22 |
| UR-P1-ctr-17 | note | open → phase 3 | with UR-P0-ctr-28 |
| UR-P1-api-18 | note | open → phase 3 | a correlation field |
| UR-P1-ui-38 | note | open → phase 3 | the fixtures carry both names; either single name changes a fixture |
| UR-P1-conc-39 | note | fixed `51d7f79` | |
| UR-P1-adpt-40 | note | open → phase 3 `cite` | rune or byte limit of the fabric label: a reference question for Fable |
| UR-P1-doc-41 | note | fixed `8089673` | decision 26 |
| UR-P1-leak-66 | note | open → phase 3 | with UR-P0-leak-27/-31 (decision 20) |
| UR-P1-api-67 | note | fixed `e7108b6` | `ControllerOf` returns `gofabric.Source` (port plus `Close`); existing callers compile |
| UR-P1-bound-68 | note | fixed `e7108b6` | |
| UR-P1-doc-69 | note | fixed `237f9c6`, `8089673` | decision 26; same root as doc-41 |
| UR-P1-conc-70 | note | fixed `ca0bf61` | |

Counts: 45 findings (41 of the reviewers, 4 of the pre-merge read) — 3
medium (1 held), 22 low (UR-P1-err-31 lowered by the verifier), 20 note.
Fixed 31 (5 of them in their no-contract part only, the rest → phase 3),
open 14 (the held row included), struck 0.

## Measured cost

| Agent | Model | Tokens | Wall clock |
| --- | --- | --- | --- |
| inventory | script | — | — |
| S-go | Fable | 221 k | 14 min |
| G-go A | Opus | 234 k | 8 min |
| G-go B | Opus | 200 k | 7 min |
| verifier | Fable | 138 k | 6 min |
| fix H | Opus | 208 k | 20 min |
| fix A | Opus | 161 k | 12 min |
| pre-merge read | Fable | 70 k | 1.5 min |
| **sum** | | **1 232 k** | |
