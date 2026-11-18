# Phase 1b — the dispositions

Branch `review/p1b` from `007ca96`; the pull request is squashed, its
single commits stay retrievable through the pull request's reference.
Sources: `notes/briefs/2026-11-18-review-p1b.md`,
`notes/reads/read-review-p1b-inventory.md` (script), the reviewers' reports
`phase-1b-S-web.md` (Fable), `phase-1b-G-web-A.md`, `phase-1b-G-web-B.md`
(Opus), `phase-1b-verify.md` (Fable), the fix reports `phase-1b-fix-W1.md`,
`phase-1b-fix-W2.md` (Opus), the pre-merge read `phase-1b-premerge.md`
(Fable, IDs 190–199), `decisions.md` rows 27–32. Severity is the
verifier's for every high and medium and for the sampled lows (7 of 29,
0 struck); the rest is the reviewer's. G-web B's `file:line` citations are
often off; `phase-1b-fix-W2.md` has the corrected lines.

**No row held back**: no finding of phase 1b touches a held class.

Every fix below has a vitest that was red before it (the commit and the
fix report name it; decision 17); the exceptions are said in the row.
"→ phase 3" means the remaining fix changes the contract (decision 13) or
belongs to a phase 3 unit.

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P1-ctr-101 | high (verifier; S-web medium, G-web B high as -162) | fixed `56ea1d5`, `ecf971d` | decision 27: lossless parse of generic values, edited value sent as typed, unsafe number refused; `docs/contract.md` says so |
| UR-P1-ctr-162 | — | merged into UR-P1-ctr-101 | verifier |
| UR-P1-bound-103 | medium (verifier; reviewer low) | fixed `7b8fc84`, `ac77145` | decision 29: streams on demand |
| UR-P1-conc-105 | medium (verifier; S-web low, G-web B medium as -160) | fixed `5c8473f` | camera released on cancel, destroy and a rejected `play()` |
| UR-P1-conc-160 | — | merged into UR-P1-conc-105 | verifier |
| UR-P1-ui-130 | medium | fixed `2fa7785` | both Diagnostics pages keyed by index |
| UR-P1-conc-102 | low | fixed `7b8fc84`, `2aa1cef` | |
| UR-P1-ui-104 | low | fixed `f7156c8` | the confirm button is armed after 500 ms |
| UR-P1-err-106 | low | fixed `c49693b` | |
| UR-P1-err-107 | low | fixed `ac77145` | |
| UR-P1-err-108 | low | fixed `49d7772` | one retry, only on a changed token |
| UR-P1-conc-109 | low | fixed `a9b3be3` | |
| UR-P1-conc-110 | low | fixed `33b4d13` | |
| UR-P1-bound-111 | low | fixed `597a463` | |
| UR-P1-ui-112 | low | fixed `033724a`, `0a48bb5` | |
| UR-P1-ctr-113 | low | open → phase 3 | the event number as a string is a contract change; with UR-P1-ctr-177 |
| UR-P1-web-114 | low | fixed `b1295bc` | |
| UR-P1-ui-131 | low | fixed `b1d236a` | |
| UR-P1-ui-132 | low | fixed `33b4d13` | |
| UR-P1-ui-133 | low | fixed `33b4d13`, `1f9609e`; controller `admin` field → phase 3 | with UR-P1-ui-196 |
| UR-P1-bound-134 | low | fixed `a021db2`, `aaf560b` | |
| UR-P1-ui-135 | low | fixed `25d31ef` | |
| UR-P1-ui-136 | low | fixed `25d31ef` | |
| UR-P1-ui-161 | low (verifier; reviewer medium) | fixed `449a447` | the verifier's variant |
| UR-P1-ui-163 | low (verifier; reviewer medium) | fixed `3bfaf44` | the verifier's variant |
| UR-P1-ui-164 | low | fixed `81ff32e` | |
| UR-P1-conc-165 | low | fixed `81ff32e` | |
| UR-P1-ui-166 | low | fixed `fd69bf3` | |
| UR-P1-ui-167 | low | fixed `d802c2e` | page part; no contract change |
| UR-P1-ctr-168 | low | open → phase 3 | needs a contract field for the auto-subscribe setting |
| UR-P1-ui-169 | low | fixed `5dba282` | |
| UR-P1-ui-170 | low | fixed `8fc3a32` | |
| UR-P1-api-171 | low | open → phase 3 `api` | exporting the controller Diagnostics page; name clash with the type |
| UR-P1-ui-172 | low | fixed `ee50f69`, `77e86e5`, `c307443` | decision 31: `--fui-*-text` tokens |
| UR-P1-test-173 | low | fixed `4b812aa` | red step was `npm run check` |
| UR-P1-conc-190 | low | fixed `4b62df9` | pre-merge read |
| UR-P1-conc-191 | low | fixed `59f0204` | pre-merge read |
| UR-P1-conc-192 | low | fixed `f2b1bc1` | pre-merge read |
| UR-P1-err-198 | low | fixed `24e3626` | pre-merge read |
| UR-P1-web-115 | note | fixed `7b57f41` (narrowed) | a hex pattern on the contract's `NodeID` → phase 3 |
| UR-P1-web-116 | note | fixed `7b8fc84` | |
| UR-P1-ui-117 | note | fixed `f7156c8` | |
| UR-P1-ui-118 | note | fixed `220af9d` | |
| UR-P1-test-119 | note | fixed with the tests of 102–118 | tests over the fixes |
| UR-P1-ui-137 | note | fixed `eb7cf05` | |
| UR-P1-ui-138 | note | fixed `c9f6e5d` | link strips are `nav` + `aria-current` |
| UR-P1-ui-139 | note | fixed `43b12a4` | |
| UR-P1-ui-140 | note | fixed `83bd929` | |
| UR-P1-ui-141 | note | fixed `e24a025`, `33b4d13` | |
| UR-P1-ui-142 | note | fixed `ee23d8c` | decision 32 |
| UR-P1-ui-143 | note | fixed `a021db2` | (a) a cost item, no red test possible |
| UR-P1-ui-144 | note | fixed `eac33c2` | |
| UR-P1-ui-145 | note | fixed `11aff4a` | |
| UR-P1-ui-174 | note | fixed `dfa14f6` | |
| UR-P1-ui-175 | note | fixed `dfa14f6` | |
| UR-P1-ui-176 | note | fixed `ab7d9c9` | |
| UR-P1-ctr-177 | note | open → phase 3 | with UR-P1-ctr-113 |
| UR-P1-ui-178 | note | open → phase 3 `cite` | the invalid passcodes rest on a reading of the specification |
| UR-P1-ui-179 | note | open → phase 3 `cite` | whether an ACL entry may name `pase` |
| UR-P1-conc-180 | note | fixed `3bfaf44`, `8f08d52` | |
| UR-P1-ui-181 | note | fixed `8f08d52` | |
| UR-P1-ui-182 | note | fixed `041ef34` | catalogue key `common.locale` |
| UR-P1-ui-193 | note | fixed `e3263e2` | pre-merge read |
| UR-P1-bound-194 | note | struck | the reader proposes no change: the slow path's limits are far above any answer, and every caller catches the `RangeError` |
| UR-P1-bound-195 | note | fixed `14b387b` | pre-merge read |
| UR-P1-ui-196 | note | fixed `8829c9a`; `admin` on the controller status → phase 3 | pre-merge read |
| UR-P1-web-197 | note | open → phase 3 `api` | the token header follows a host's `baseUrl`; a docs sentence or an origin check |
| UR-P1-conc-199 | note | fixed `b7a9a64`, `9366d66` | pre-merge read |

New while verifying, not a finding of this repository: go-fabric v0.22.2
reports an administrator's basic window as `WindowEnhanced`
(`node/commissioning.go:364-366` against `node/observers.go:225-228`) →
a go-fabric issue after this pull request (decision 30).

Counts: 66 findings (56 of the reviewers after two merges, 10 of the
pre-merge read) — 1 high, 3 medium, 33 low, 29 note. Fixed 58 (3 of them
with a part → phase 3), open 7 (→ phase 3), struck 1.

## Measured cost

| Agent | Model | Tokens | Wall clock |
| --- | --- | --- | --- |
| inventory | script | — | — |
| S-web | Fable | 201 k | 10 min |
| G-web A | Opus | 235 k | 13 min |
| G-web B | Opus | 203 k | 11 min |
| verifier | Fable | 205 k | 12 min |
| fix W1 (with the follow-up) | Opus | 344 k | 64 min |
| fix W2 (with the follow-up) | Opus | 417 k | 74 min |
| pre-merge read | Fable | 126 k | 6 min |
| **sum** | | **1 731 k** | |
