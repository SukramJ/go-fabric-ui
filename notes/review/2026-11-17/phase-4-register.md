# Phase 4 — convergence run 1, the dispositions

Branch `review/p4` from `979a67a` (origin/main after #78, release
v0.10.0), HEAD `f787172` when written; the pull request is squashed, its
single commits stay retrievable through the pull request's reference.
go-fabric pin `v0.23.0`. Decisions 65–73 and the addendum to 71. Contract
0.10.0 → 0.11.0, additive (no accepted break). Sources:
`notes/briefs/2026-11-19-review-p4.md`, `notes/reads/read-review-p4-delta.md`
(script), the unit reports `phase-4-S.md`, `phase-4-S2.md` (Fable),
`phase-4-G.md`, `phase-4-W.md` (Opus), `phase-4-verify.md` (Fable; its
"Disposition"), the pre-merge read `phase-4-premerge.md` (Fable, IDs
190–199). Severity is the verifier's for the mediums and the sampled lows
(8 of 8 confirmed, 0 struck) and the pre-merge reader's for 190–199; the
rest is the reviewer's.

**Held rows, now public.** UR-P4-auth-03 (medium, run 1) and UR-P4-doc-190
(medium, pre-merge) were held until this pull request; their fixes
(decision 69) are docs for the whole class and merge here, so every
phase 4 report publishes with it (scratch paths replaced). The held
files of phases 0–2 stay held.

**Run 1 is not the last.** It found three new mediums (auth-03, adpt-31,
adpt-32), so a convergence run 2 reads the merged HEAD (README "Done").

## Carried rows

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P1-err-33 | low | fixed `4e032dd` | the node route maps go-fabric's `node.ErrFabricNotFound`, the controller route the NOCResponse `InvalidFabricIndex` to 404 |
| UR-P2-err-37 | note | fixed `8c7fbdf` | go-fabric's `client.ErrRequestTooLarge` / `ErrMalformedRequest` are 400, not the node's 502 |
| decision 30 | — | fixed `a44044a`, `8340339`, `e38a336` | v0.23.0 reports an administrator's basic window as its own kind: `window_kind` `administrator_basic` (contract 0.11.0); the page needs no change |
| UR-P2-err-09 | note | fixed `cda23ec` (test) | v0.23.0 disarms the fail-safe on a cancelled context; the adapter's cancel path is held by a test |
| UR-P1-api-171 | low | fixed `b88d168` | `ControllerDiagnosticsPage` from `./pages` and the package root |
| UR-P3-web-191 | medium | fixed `a548f8c` | the node stream reopens when a status read brings another token |

## Findings of run 1

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-auth-01 | low | fixed `ae29fea` | `sibling` carries `AllowedHosts` and `AllowHost`; a reflection guard over the per-handler fields |
| UR-P4-bound-02 | low | fixed `2820389` | decision 71 and its addendum: 600 s on the discovered path; the code path waits on go-fabric#554 |
| UR-P4-auth-03 | medium | docs `a15b0cc` (decision 69) | a proxy that replaces `Host` makes the check constant; `docs/embedding.md`, `docs/contract.md`, the `AllowedHosts` godoc, the changelog; `host_proxy_test.go` asserts the docs sentences |
| UR-P4-web-04 | note | fixed `3194bcf` | the token's same-origin test resolves a relative base against `<base href>`, as fetch does |
| UR-P4-auth-05 | note | docs `a46642d` | the godoc says what admitting `*.localhost` assumes |
| UR-P4-ctr-30 | low | fixed `71863a8`, `9ce9060` | `format: int64` on the uint32/uint64 fields; the whole Go suite runs under `GOARCH=386` (`make test-386`, CI) |
| UR-P4-adpt-31 | medium | fixed `6f98ddc` | `Wrap` chains `Deleted`; a reflection guard over every `client.Hooks` field |
| UR-P4-adpt-32 | medium | fixed `71863a8`, `8d7dd2c`, `60e6cd6` | decision 70: a partial result with `errors[]`; the adapter re-checks node by node; the page shows a node's failed check |
| UR-P4-test-33 | note | fixed `8d7dd2c` | the status guard (`internal/statusguard`) also sees the adapter's tests; it found two more undeclared 502s (now 500) |
| UR-P4-doc-34 | note | fixed `5d04184` | comment |
| UR-P4-ui-60 | low | fixed `743eb14` | the viewer's countdown ticks |
| UR-P4-ui-61 | low | partly `c68debf` | a reassemble clears the notice; an apply from another tab can leave it (no event says whether it was rebuilt) |
| UR-P4-test-62 | low | fixed `5033927` | the contract guard reports a parameter or a body that became required |
| UR-P4-test-63 | low | fixed `70e7e63` | the i18n guard reads every key literal and every template prefix |
| UR-P4-ui-64 | note | fixed `546fd56` | the label hint says bytes |
| UR-P4-ui-65 | note | fixed `5172347` | warnings without a message say an administrator sees the details |
| UR-P4-ui-66 | note | fixed `a53b94c` | the unknown entry is refused before the lock-out question |
| UR-P4-auth-90 | low | docs `ea1f1ef`, fixed `71f6031` | decision 73: said in the contract, godoc and docs; the page shows the own window back with its codes |
| UR-P4-leak-91 | low | fixed `f20b416` | logged on a change of state, not per request |
| UR-P4-doc-92 | note | fixed `5d04184` | comment |

## Pre-merge read

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-doc-190 | medium | docs `54a41e0` | held until this merge: only the name list keeps the check behind a forwarding proxy; `AllowHost` answered from the proxy's source address leaves the check to the proxy (decision 61's case), writes included; `docs/embedding.md`, `docs/contract.md`, the godoc and the changelog name three kinds of proxy; the docs guard holds four sentences |
| UR-P4-doc-191 | low | docs `cd7139b` | the own window comes back only after a window on fresh credentials; after a `ConfiguredWindow` it stays closed |
| UR-P4-adpt-192 | low | fixed `6b1f428`; upstream go-fabric#555 | one node per product is checked again; go-fabric#555 asks for per-product errors |
| UR-P4-ctr-193 | low | fixed `6b81d06` | `window_match` true under `administrator_basic` |
| UR-P4-test-194 | note | fixed `41d06c6` | the sibling guard fills every shared field before it compares |
| UR-P4-web-195 | low | fixed `f787172` | a scoped node's stream opens once, compared with the token the client sent |
| UR-P4-ctr-196 | note | docs `75988c5`; rest → run 2 | a minor may add a response enum value, a client tolerates it; mapping a future unnamed go-fabric window kind to a fixed value is a contract choice left to run 2 |

Second look at `f787172` (the verifier, resumed): doc-190 and doc-191 closed, ctr-193 OK; one new row:

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-adpt-197 | low | → run 2 | the re-check's product key uses the software version string; go-fabric's `QueryUpdates` groups by the numeric version — key on `SoftwareVersion` |

## Measured cost

| Agent | Model | Tokens |
| --- | --- | --- |
| S security | Fable | 307 k |
| G Go rest | Opus | 230 k |
| W web rest | Opus | 179 k |
| verifier | Fable | 132 k |
| S2 the gaps (decision 72) | Fable | 332 k |
| fix G (stage 1, 2, docs) | Opus | ~406 k |
| fix W (incl. follow-up) | Opus | ~225 k |
| pre-merge read (the verifier, resumed) | Fable | ~140 k |
| **sum** | | **~1 950 k** |

## Guards

At `92bbf10` and again at `f787172` (after the pre-merge fixes), local, all exit 0: `make ci`, `make race`, `make cover-check`,
`make vuln licenses secrets`, `GOARCH=386 make build`, `make test-386`
(new), `make notices-check`, `make version-check`, `make contract-check`,
`make e2e`, `make e2e-web` (44/44).

## Differences from the brief

- Decision 71 named "the commission-by-code path"; the code path is
  `Pair`, whose search bound a deadline would switch off — the 600 s
  sit on the discovered path (addendum in `decisions.md`).
- Decision 70's "per-node field" is a list `errors[]` beside `updates`:
  an `AvailableUpdate` needs a found update.
- The contract moved once after G's stage 1 (`a44044a`,
  `administrator_basic`), still additive.
- A fifth review agent (S2, decision 72) read what unit S left unread;
  the pre-merge read is the verifier resumed, so the cap held.

## Upstream

- go-fabric#555 (filed 2026-11-19): `QueryUpdates` names no product in its joined error (UR-P4-adpt-192).
- go-fabric#554 (filed 2026-11-19): a `PairOptions` timeout for the
  commissioning after the search (decision 71).
- go-fabric-ui#81: `CommissionDiscovered` through go-fabric's
  `Controller.PairDiscovered` (go-fabric#552, unreleased) — closes
  UR-P3-adpt-92 and the adapter's own attestation rule of UR-P1-adpt-60
  once the release is pinned; run 2's fix wave if it is out by then.

## Hosts' obligations (0.11.0)

Additive over 0.10.0; `changelog.md` 0.11.0 and `docs/contract.md`
"From 0.11.0". A host behind a proxy that replaces `Host` reads
`docs/embedding.md` "Host names": the check is then the proxy's.

## Carried to run 2

UR-P4-adpt-197; UR-P4-ui-61's other-tab path; UR-P4-ctr-196's mapping of a
future go-fabric window kind; go-fabric-ui#81 if go-fabric releases
`PairDiscovered`. Run 2 reads the merged HEAD in a fresh session.

Counts: 6 carried rows (all fixed), 22 run-1 findings (3 medium, 9 low,
10 note), 8 pre-merge findings (1 medium, 4 low, 2 note, plus 197 low).
By status: fixed 25, docs 6, partly 1 (ui-61), → run 2 1 (adpt-197),
struck 0.
