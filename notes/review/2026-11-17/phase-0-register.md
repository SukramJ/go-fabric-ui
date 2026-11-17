# Phase 0 — the dispositions

Branch `review/p0`; the pull request is squashed, its single commits stay
retrievable through the pull request's reference. Sources:
`phase-0-evidence.md` (mechanical evidence, IDs 01–19), `threat-model.md`
(IDs 20–33), `phase-0-verify.md`, `phase-0-fix.md`, `decisions.md` rows
18–24. Severity is the verifier's for every high and medium and for the
sampled lows (0 of 3 struck); the rest is the reviewer's.

**2 rows held back** (decisions 5, 14 and 18): one high, one medium, with
one observation the verifier added to the high. Their reports live in the
review worktree's git-excluded `held/` directory and are published with
phase 3's pull request, which fixes them. The public reports were redacted
of the hints to them before this commit.

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| (held) | high | open → phase 3 | decision 18 |
| (held) | medium | open → phase 3 | decision 18 |
| UR-P0-doc-22 | medium | open → phase 3, with the held rows | decision 18: its text depends on their fix |
| UR-P0-web-23 | medium | fixed `65a9c9e`, `7bd7fb2` | `frame-ancestors 'self'` + `Options.FrameAncestors`; any-parent entries refused (decision 22) |
| UR-P0-auth-26 | medium | open → phase 3 | decision 19: all four routes gated (contract) |
| UR-P0-leak-27 | medium | open → phase 3 | decision 20: log and log stream gated, the client's log stream through `fetch` (contract, web) |
| UR-P0-web-24 | low | fixed `65a9c9e`, `73157ee`, `1c04fcf` | CSP with the pre-paint script's hash derived at start; `nosniff`, `no-referrer` on every answer; Playwright asserts no violation |
| UR-P0-leak-25 | low | fixed `73157ee` | `no-store` on every API answer |
| UR-P0-ctr-28 | low | open → phase 3's wizard | contract if made strict (decision 13); the plan's "Bodies" bullet corrected |
| UR-P0-bound-29 | low | open → phase 3's wizard | a new 503 case (contract) |
| UR-P0-ui-30 | low | fixed `3de7c98` | vitest red first |
| UR-P0-sup-02 | low | fixed `43735ec` | `make licenses` fails on an unclassified licence; commented allow-list |
| UR-P0-leak-31 | note | open → phase 3, with UR-P0-leak-27 | decision 20 |
| UR-P0-api-32 | note | fixed `37bde08`, `1c04fcf` | one pattern `Prefix/` after a lookup; late host routes shadow the API (decision 23) |
| UR-P0-web-33 | note | fixed `8d0f7ac` | red test with a line break in the type |
| UR-P0-sup-01 | note | open → Dependabot | `picomatch` ≤ 4.0.7, dev-only, no fixed release on 2026-11-17 |
| UR-P0-err-01 | note | struck | the generated strict server makes `req.Body == nil` unreachable from a request (`phase-0-fix.md`) |
| UR-P0-test-01 | note | fixed `2fd3a9a` | tests over existing behaviour; no red step possible (decision 17 applies to defects, these were gaps) |

Host observations HR-1 to HR-7 (`threat-model.md`, HR-1 held) go to the
reviews of go-fabric-server and go-ha2fabric.

## Measured cost

| Agent | Model | Tokens | Wall clock |
| --- | --- | --- | --- |
| inventory | Sonnet | 308 k | 13 min |
| mechanical evidence | Opus | 173 k | 58 min |
| threat model | Fable | 205 k | 10 min |
| verifier | Fable | 105 k | 5 min |
| fix wave | Opus | 183 k | 27 min |
| **sum** | | **974 k** | |
