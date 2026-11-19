# Phase 3 — the dispositions

Branch `review/p3` from `f4b8573`, HEAD `512f2bc` when written; the pull
request is squashed, its single commits stay retrievable through the pull
request's reference. go-fabric pin `v0.22.2`, lifted to `v0.23.0` at the
end (decision 64). Decisions 37–64. Contract
0.9.0 → 0.10.0 with three accepted breaks (`info.x-accepted-breaks` in
`contract/openapi.yaml`): `Status.fabric_count` no longer required, the
`NodeID` path parameter narrower (pattern), the `EndpointNumber` path
parameter narrower (maximum). Sources: `notes/briefs/2026-11-18-review-p3.md`,
`notes/briefs/2026-11-18-review-p3-contract.md` (Fable),
`notes/reads/read-review-p3-carryover.md` (script, 46 rows), the unit
reports `phase-3-A.md`, `phase-3-B.md`, `phase-3-C.md` (Opus),
`phase-3-D.md` (Fable), `phase-3-verify.md` (Fable; its "Disposition"),
the pre-merge read `phase-3-premerge.md` (Fable, IDs 190–197), the held
notes now public `phase-3-carryover-held.md` and `phase-3-contract-held.md`,
`decisions.md` rows 37–63. Severity is the verifier's for the sampled
lows (9 of 9 confirmed, 0 struck, none changed) and the pre-merge
reader's for 190–197; the rest is the reviewer's.

**Held rows, now public.** The three held rows of phases 0 and 1a
(UR-P0-auth-20, UR-P0-web-21, UR-P1-adpt-60) are fixed in this pull
request, so every phase 3 report publishes with it. The public copy of
`phase-3-D.md` gives UR-P3-adpt-92 in the reduced wording of
`phase-3-verify.md` "Item 2"; scratch paths are replaced by "(scratch
worktree)". The held files of phases 0–2 stay held.

"→ phase 4" means the remaining fix uses what go-fabric `v0.23.0`
exports; the pin is lifted in this pull request (decision 64, with
decision 62's migration), the fixes are phase 4.

## Findings of phase 3

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P3-ctr-01 | low | fixed `841b8fb`, `bba61bb` | decision 47: every status per operation through shared responses; a Go guard that every answered status is declared (it found 409/502 on three routes). Takes doc-78 |
| UR-P3-test-02 | low | fixed `5272f7c` | `contract/previous/` is the 0.9.0 baseline from `v0.9.2`. Takes sup-60 |
| UR-P3-test-03 | low | fixed `5272f7c`, `841b8fb` | the guard learns required→optional, type changes, tighter patterns and maximums; decision 46's twins and optional fields |
| UR-P3-ctr-04 | note | fixed `19fb493`, `562dffc`, `9350659` | decision 45: one admin test on every route, the streams included; the client reads all three streams through `fetch` |
| UR-P3-api-05 | low | fixed `91062a5` | decision 48: `NewHandler(o) (http.Handler, error)`; godocs corrected. Takes doc-74 |
| UR-P3-ctr-06 | note | fixed `ec57fba` | strict bodies: exact member names, nested objects included |
| UR-P3-ctr-07 | note | fixed `562dffc` | the request header and token on every request; with UR-P1-web-197 |
| UR-P3-api-08 | note | fixed `19fb493` | the host's events no longer the handler's own hub; host events admin-only |
| UR-P3-test-30 | low | fixed `7436888` | the adapter's error, update status, ICD mode and binding list mappings under test |
| UR-P3-test-31 | low | fixed `e2df2f7` | floors at the measurement; a package without a line fails |
| UR-P3-test-32 | low | fixed `ec57fba` | the fake node's windows open at their own time |
| UR-P3-test-33 | low | fixed `d1c117e` | template keys enumerated |
| UR-P3-test-34 | note | fixed `ec57fba` | the stream guard checks every line's kind |
| UR-P3-test-35 | note | fixed `d1c117e` | each-keys allowed per block, whole |
| UR-P3-test-36 | note | fixed `d1c117e`, `3682e3c` | flows watched for console errors and policy violations |
| UR-P3-test-37 | — | merged into UR-P3-sup-71 | verifier |
| UR-P3-sup-60 | — | merged into UR-P3-test-02 | verifier |
| UR-P3-sup-61 | low | fixed `6996388` | verify, release and publish jobs with their own rights; the `npm` environment is decision 56 (open) |
| UR-P3-sup-62 | low | fixed `331e963` | every action pinned by commit SHA, `make pins-check` |
| UR-P3-sup-63 | note | fixed `6996388`; ruleset open (decision 56) | the release checks the tag is on `main` with passed checks; the `v*` tag ruleset is an owner setting |
| UR-P3-sup-64 | note | open (decision 56) | the six required checks bound to GitHub Actions: an owner setting, applied by the owner's session after confirmation |
| UR-P3-sup-65 | low | fixed `0c3623a` | go-fabric bumps excluded from Dependabot auto-merge |
| UR-P3-sup-66 | note | fixed `7b332ab` | CodeQL for `javascript-typescript`; the required context is decision 56 (open) |
| UR-P3-sup-67 | low | partly `4256928` | tailwindcss in the bundle table; the guard sits in the Makefile (`notices-check`) |
| UR-P3-sup-68 | note | docs `4256928` | stale Go rows dropped from the notices |
| UR-P3-sup-69 | note | fixed `491414b` | decision 55: oapi-codegen in a tools module of its own |
| UR-P3-sup-70 | note | fixed `e847214` | decision 54: `npm audit` of the runtime packages in the web job, the whole tree nightly |
| UR-P3-sup-71 | note | docs `5a78459` | `make ci` described as it is; `FUZZTIME` used by `make fuzz` (decision 53). Takes test-37 |
| UR-P3-doc-72 | low | fixed `91062a5`, `16e3f64` | decision 49: `Handler` and `NewHandler` refuse a non-default `Prefix`; docs say so |
| UR-P3-doc-73 | low | docs `75d030d`, `512f2bc` | README, CONCEPT and the CLAUDE.md state line at v0.10.0 |
| UR-P3-doc-74 | — | merged into UR-P3-api-05 | verifier |
| UR-P3-doc-75 | note | docs `16e3f64` | "a 500 problem carries no detail" |
| UR-P3-doc-76 | note | docs `16e3f64` | the admin routes as the contract marks them |
| UR-P3-doc-77 | note | docs `16e3f64` | `FrameAncestors` in the Options table, the wildcard rule written |
| UR-P3-doc-78 | — | merged into UR-P3-ctr-01 | verifier |
| UR-P3-doc-90 | low | docs `16e3f64` | `logring` has landed, cited at v0.22.2 |
| UR-P3-adpt-91 | low | fixed `ec57fba` | decision 52: the controller client's own refusals are 409 |
| UR-P3-adpt-92 | note | upstream go-fabric; docs `16e3f64` | a discovered device's commissioning writes no pairing entry and no metric; a facade entry for a discovered device and a passcode (go-fabric issue with decision 25) |
| UR-P3-cite-93 | note | docs `e6bbb8d` | go-fabric ranges with their tags |
| UR-P3-cite-94 | note | docs `5a6d1fa` | openccu-loom ranges corrected |
| UR-P3-cite-95 | note | docs `5a6d1fa` | the toast store's loom comment corrected |
| UR-P3-cite-96 | note | docs `e6bbb8d` | go-fabric-server lines and pins |
| UR-P3-cite-97 | note | docs `e6bbb8d` | concept pages pinned |
| UR-P3-web-190 | medium | fixed `3dab117` | pre-merge read: the status is repeated with its token; the ACL page asks when its own id is unknown |
| UR-P3-web-191 | medium | partly `3dab117` → phase 4 | pre-merge read: a refused admin stream refreshes the token and reopens once; the node stream (an open route) is not reopened on a token change after a daemon restart and stays filtered until a reload |
| UR-P3-test-192 | low | fixed `3b9aaf4` | pre-merge read: a `NewHandler` test of the Host check on the bundle |
| UR-P3-test-193 | low | fixed `d597108` | pre-merge read: a web test of the store row for an administrator |
| UR-P3-web-194 | low | fixed `3dab117` | pre-merge read: the token's origin by URL resolution |
| UR-P3-leak-195 | low | fixed `fca5fbe` | pre-merge read; decision 63: `attestation_error` is a fixed sentence without admin rights (decision 63 and the commit call it UR-P3-web-195) |
| UR-P3-word-196 | note | fixed `3b9aaf4` | pre-merge read: test names and comments say what is checked |
| UR-P3-bound-197 | note | accepted (decision 42) | pre-merge read: one stream cap per handler, open and admin streams together; no fix proposed |

Not a row: B's proposal of a standing fuzz job, decision 53 — `8e31057`,
`5a78459`, `62436df` (thirteen targets with seed corpora, `make fuzz`
with `FUZZTIME`, a nightly job). Decision 24 (an exported CSP builder):
not now (unit A: a `Mount` host's page is its own document; what the
pages need is in `ui/README.md`). Decision 56 (repository settings,
sup-61, -63, -64, -66) is open: applied by the owner's session after
the owner confirms.

## Carried rows (46)

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P0-auth-20 | high | fixed `c3c4534`, `3b9aaf4` | decisions 37, 61: a Host check on every route in every mode, `AllowedHosts` and `AllowHost`; 403 `host_not_allowed` |
| UR-P0-web-21 | medium | fixed `933137a`, `562dffc` | decision 38: the request header, the Origin check, 415 on another media type; the client sends the header |
| UR-P1-adpt-60 | medium | fixed `7ad26f4` | decision 25: `CommissionDiscovered` applies the template's attestation policy, an unset one rejecting |
| UR-P0-doc-22 | medium | docs `16e3f64` | what `Admin` must guarantee and what a false one serves |
| UR-P0-auth-26 | medium | fixed `841b8fb`, `19fb493` | decision 19: the four routes gated |
| UR-P0-leak-27 | medium | fixed `19fb493`, `562dffc` | decision 20: log and log stream gated, the client's log stream through `fetch` |
| UR-P0-ctr-28 | low | fixed `ec57fba` | decision 43: strict bodies everywhere |
| UR-P0-bound-29 | low | fixed `91062a5` | decision 42: `MaxStreams`, 503 `too_many_streams` with `Retry-After` |
| UR-P0-leak-31 | note | fixed `19fb493` | with leak-27 |
| UR-P1-err-04 | low | fixed `ec57fba`, `9350659` | `window_match` read by the page |
| UR-P1-err-07 | low | fixed `ec57fba` | `fabric_count` absent on a failed read (accepted break) |
| UR-P1-leak-09 | low | fixed `19fb493` | decision 20 |
| UR-P1-adpt-65 | low | fixed `ec57fba` | a presence flag for the discovered device's long discriminator |
| UR-P1-ctr-32 | low | fixed `841b8fb` | `unknown` in the ACL enums |
| UR-P1-err-33 | low | partly `ec57fba` → phase 4 | the node route's unknown fabric is 404; the controller route's waits on the error go-fabric v0.23.0 exports |
| UR-P1-adpt-35 | low | fixed `ec57fba` | endpoint 65535 is 400, `maximum: 65534` (accepted break) |
| UR-P1-err-36 | low | fixed `ec57fba` | compatibility 500 on a failed fabric read |
| UR-P1-bound-37 | low | fixed `ec57fba`, `5e34fd3` | a per-endpoint status on bindings |
| UR-P1-web-20 | note | docs `91062a5`, `16e3f64` | decision 50: a godoc and docs warning on a wildcard over a public suffix |
| UR-P1-conc-21 | note | docs `e6bbb8d` | `NodeRegistry`'s doc: its methods must not call back into the handler |
| UR-P1-ctr-13 | note | fixed `ec57fba` | 405 with `Allow` at path level (decision 60) |
| UR-P1-ctr-17 | note | fixed `ec57fba` | with ctr-28 |
| UR-P1-api-18 | note | fixed `ec57fba`, `9350659` | decision 59: `commissioning_id` chosen by the client, echoed on progress |
| UR-P1-ui-38 | note | fixed `77e35df`, `9a184ea` | one device type name across the node family |
| UR-P1-adpt-40 | note | fixed `ec57fba` | a fabric label over 32 bytes is 400; go-fabric v0.23.0's controller checks bytes too (`commissioner.MaxLabelLength`), the question is answered |
| UR-P1-leak-66 | note | fixed `19fb493` | decision 20; takes UR-P2-leak-63 |
| UR-P1-ctr-113 | low | fixed `841b8fb`, `9350659` | `number_str` twin |
| UR-P1-ui-133 | low | fixed `841b8fb`, `9350659` | `admin` on the controller status |
| UR-P1-ctr-168 | low | fixed `ec57fba`, `5e34fd3` | `auto_subscribe` |
| UR-P1-api-171 | low | open → phase 4 | the controller Diagnostics page is still not exported from `./pages` (name clash with the type); no commit of this wave touches it |
| UR-P1-web-115 | note | fixed `841b8fb` | a hex pattern on `NodeID` (accepted break) |
| UR-P1-ctr-177 | note | fixed `ec57fba` | with ctr-113 |
| UR-P1-ui-178 | note | fixed `ec57fba`, `5e34fd3` | an invalid passcode is 400 and refused by the page |
| UR-P1-ui-179 | note | fixed `5e34fd3` | the ACL add form drops `pase`; display and enum keep it |
| UR-P1-ui-196 | note | fixed `841b8fb`, `9350659` | with ui-133 |
| UR-P1-web-197 | note | fixed `562dffc`, `3dab117` | the token only to a base of the page's origin |
| UR-P2-ui-05 | low | fixed `ec57fba`, `9350659` | decision 28's `window_match` |
| UR-P2-ui-06 | note | fixed `ec57fba`, `9350659`, `9a184ea` | `window_remaining_seconds` |
| UR-P2-conc-34 | low | fixed `91062a5` | decision 51: `Options.Done` |
| UR-P2-ui-35 | note | fixed `ec57fba`, `9350659` | a failed rebuild on exposure apply (`reassembled`) |
| UR-P2-err-37 | note | → phase 4 | waits on the errors go-fabric v0.23.0 now exports (#544) |
| UR-P2-auth-60 | note | fixed `19fb493` | decisions 33, 40 |
| UR-P2-leak-61 | note | fixed `19fb493` | decisions 33, 40 |
| UR-P2-doc-62 | low | docs `16e3f64` | with UR-P0-doc-22 |
| UR-P2-leak-63 | — | merged into UR-P1-leak-66 | fixed with it |
| UR-P2-ctr-64 | note | fixed `841b8fb`, `ec57fba` | 64-bit values as string twins |

Counts: 97 rows. Phase 3's own: 51 IDs, of which 4 merged; 47 findings
(39 of the reviewers — 16 low, 23 note — and 8 of the pre-merge read —
2 medium, 4 low, 2 note). Carried: 46 (1 high, 5 medium, 18 low, 21
note, 1 merged). By status over all 97 rows: fixed 68, partly 3
(sup-67, web-191, err-33), docs 16, → phase 4 2 (err-37, api-171), open
owner setting 1 (sup-64), upstream 1 (adpt-92), accepted 1 (bound-197),
merged 5, struck 0.

## Measured cost

| Agent | Model | Tokens |
| --- | --- | --- |
| A contract and api | Opus | 228 k |
| B tests | Opus | 216 k |
| C docs and supply | Opus | 206 k |
| D cite | Fable | 309 k |
| verifier | Fable | 159 k |
| contract design note | Fable | ~193 k |
| fix G (incl. its two helpers F and D) | — | ~465 k |
| fix W | — | ~372 k |
| pre-merge read | Fable | 184 k |
| fix round after the pre-merge read | — | ~152 k |
| **sum** | | **~2 484 k** |

## Guards

At `512dd13` (go-fabric v0.23.0 and the two CI fixes; first at `512f2bc` on v0.22.2), local, all exit 0: `make ci`, `make race`, `make cover-check`, `make vuln licenses secrets`, `GOARCH=386 make build`, `make notices-check`, `make version-check`, `make contract-check`, `make e2e`, `make e2e-web` (44/44).

## Differences from the brief

- The go-fabric pin is lifted to `v0.23.0` in this pull request
  (decision 64), not in phase 4. Its timing exposed a lost update in
  the adapter's label push (a fabric label set twice ended at the
  first); fixed with it.
- The pull request's first CI run failed two jobs the local guards had
  passed: `web` (the contract guard's test reads the newest tag; the
  job's checkout had no tags → `fetch-depth: 0`) and `e2e-web` (1 in
  about 45 runs: the shell chose the Operation page before its probes
  answered and read the log without the token → it waits for them; a
  shell test holds it).
- G spawned two helper agents (F and D) beyond the cap of two
  implementing agents (decision 39); their tokens are in G's line.
- A measurement of the pinned hosts' proxy contradicted one premise of
  decision 37's Host rule → decision 61 (a per-request predicate,
  `AllowHost`, beside the name list).
- The pre-merge read's UR-P3-leak-195 needed an owner choice → decision
  63.
- W corrected tests the 0.10.0 fixtures had turned false (`9a184ea`,
  `eaa30a2`) and the smokes' wait (`3682e3c`).

## Upstream

- go-fabric v0.23.0 closed #539, #543 (UR-P2-err-09) and #544
  (UR-P2-err-37); its pin is lifted here (decision 64), the fixes using
  it are phase 4 (UR-P1-err-33, UR-P2-err-37). It also answers
  UR-P1-adpt-40's question: the controller checks the label in bytes.
- To file after the merge: decision 25's facade gap together with
  UR-P3-adpt-92 (one issue).

## Hosts' obligations (0.10.0)

- Contract 0.10.0: a request without admin rights loses the gated
  routes (decisions 19, 20, 33–35, 40, 41) and gets reduced statuses
  (decisions 46, 57, 63).
- Every host client sends the request header `X-Fabric-UI: 1` on every
  request (and in token mode the token), the streams included, so it
  reads the streams through `fetch`.
- A host reached under a name that is not loopback or an IP literal —
  the Ingress case — lists it in `Options.AllowedHosts` or admits it with
  `Options.AllowHost`; else every route is 403 `host_not_allowed`.
- `Options.MaxStreams` (default 64), `Options.Done` for shutdown, and
  `NewHandler` returning the error `Handler` panics on (a non-default
  `Prefix` is one).

## Lessons

- An implementer's brief forbids sub-agents explicitly.
- Streams over `fetch` never let Playwright's `networkidle` settle; a
  smoke waits for no pending request but the streams.
- A harness may refuse git in another worktree; an agent hands back a
  branch instead.
- Measure what a proxy sends as `Host` before deciding a Host rule.
