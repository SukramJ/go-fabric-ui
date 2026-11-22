# Phase 4 — convergence run 2, the dispositions

Branch `review/p4r2` from `792d15b` (origin/main after #82, release
v0.11.0), rebased onto origin/main `0fb2ee0`; HEAD `dbb51ab` when written
(the hashes below are those after the rebase; the pull request is
squashed, its single commits stay retrievable through the pull request's
reference). go-fabric pin `v0.23.0` → `v0.24.1` (decision 80). Decisions
74–82. Contract 0.11.0 → 0.12.0, additive (no accepted break); release
`v0.12.0` (decision 76, in place of decision 75's `v0.11.1`). Sources:
`notes/briefs/2026-11-19-review-p4r2.md` (with its fix-wave table),
`notes/reads/read-review-p4r2-delta.md` (script), the unit reports
`phase-4r2-S-go.md`, `phase-4r2-S-web.md` (Fable), `phase-4r2-G.md`,
`phase-4r2-W.md` (Opus), `phase-4r2-verify.md` (Fable: the verification
and its "Disposition" of severities, the gap read, IDs 320–322, the
pre-merge read, IDs 390–396, and the second look at the go-fabric v0.24.1
lift, IDs 400–407). Severity is the verifier's for the sampled lows (12
of 17: 11 confirmed, ui-306 re-rated down to note; ui-233 re-rated up to
low; 0 struck), the verifier's for 320–322, 390–396 and 400–407; the rest
is the reviewer's.

**No held row.** Run 2 found no high and no medium, so no report was a
held candidate; every `phase-4r2-*` report publishes with this pull
request (scratch paths replaced by `<scratch>`).

**Run 2 is the last.** It found no new high or medium (README "Done": a
convergence run over the then-current HEAD, with go-fabric's
post-Review-II tag pinned, yields none), and neither did its pre-merge
read or the second look. Phase 4, and with it the go-fabric-ui review,
ends with this pull request. Four notes stay open and are carried
(below); README "Done" asks each finding to close as fixed, struck,
accepted or upstream, so each of the four needs its issue or a recorded
decision when it is carried.

## Carried rows

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-adpt-197 | low | fixed `b251e4a`; superseded `3424872` | first keyed on the numeric `SoftwareVersion`; then the re-check is gone whole: go-fabric#555's per-product errors (v0.24.0) give each node its product's cause, one `QueryUpdates`, one `check_finished` (closes UR-P4-adpt-192's re-query too) |
| UR-P4-ui-61 (other tab) | low | fixed `39512e6` | a rebuild elsewhere, seen as `surface_changed` or `endpoints_changed`, clears the not-in-effect notice; its rest is UR-P4-ui-396 |
| UR-P4-ctr-196 (mapping) | note | fixed `75497bc`, `9f53b1a`, `21b1771` | `window_kind` `other` for a kind the contract does not name (contract 0.12.0), not go-fabric's `WindowKind(n)` text; the page shows it as any open window |

## Issues

| Issue | disposition | where |
| --- | --- | --- |
| go-fabric-ui#81 | closed by this pull request, `787352c` | `CommissionDiscovered` through go-fabric's `Controller.PairDiscovered` (go-fabric#553, v0.24.0): the template's attestation policy, zero value rejecting, so UR-P1-adpt-60's restatement is dropped; the pairing in the diagnostic trace and `pair.failures` (closes UR-P3-adpt-92); BLE and Address cleared |
| go-fabric#554 | used, `949261f` | `PairOptions.CommissionTimeout`: a code's commissioning bounded at 600 s after the search (decision 71's intent, UR-P4-bound-02's code path) |
| go-fabric#555 | used, `3424872` | `QueryUpdates`' per-product errors (UR-P4-adpt-192, -197) |

## Findings of run 2

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-leak-200 | low | fixed `81ec227`, `fef3059` | an abandoned status request is not a fabric failure: `serverError`'s abandoned rule before the state and the log (UR-P4-leak-91's reach) |
| UR-P4-adpt-201 | low | fixed `79d0572` | `For` subscribes `Deleted`; exactly one `node_removed` on the decommission, leave and forget paths |
| UR-P4-doc-202 | note | docs `2a5267b` | `docs/embedding.md`: a proxy's own name for `Host` is listed in `AllowedHosts`; the 0.12.0 changelog corrects 0.11.0's sentence |
| UR-P4-conc-230 | low | fixed `2ea0296` | the status's three readers share one count; an overtaken read leaves neither "loading" nor a stale stream |
| UR-P4-err-231 | low | fixed `90f175a` | the token reload's own status reads carry `retried: true` |
| UR-P4-ui-232 | note | fixed `2893335` | the Operation tab shows on an answer only; an undecided probe is asked again |
| UR-P4-ui-233 | low (re-rated from note) | fixed `8be37b8` | a refused stream is shown, reloads the status and reopens when a status grants the rights again |
| UR-P4-ui-234 | note | fixed `a42ec34` | a QR payload the encoder refuses shows the codes without the picture |
| UR-P4-web-235 | note | fixed `00af194` | `renameNode` and `deleteNode` check the node key as `forNode` does |
| UR-P4-ctr-260 | low | fixed `75497bc`, `56c44b8` | `x-go-type: int64` on the discovery `value`; `make test-386` holds it |
| UR-P4-err-261 | low | fixed `032281e` | the factory reset passes over a fabric removed meanwhile, as go-fabric's `FactoryReset` does |
| UR-P4-ctr-262 | note | fixed `82bb0c7` | failed attributes sorted with `cmp.Compare` |
| UR-P4-doc-263 | note | docs `322f888`, `5a93a01` (decision 78) | the comma-separated Multi form stays accepted, deprecated, no removal date; no behaviour change |
| UR-P4-doc-264 | note | docs `c2a99fe` | `NetworkNode.Kind`'s godoc names `border_router` |
| UR-P4-test-265 | note | fixed `b1ff600` | a skipped entry is logged exactly once, and again on return |
| UR-P4-test-266 | note | fixed `453e0a9` | every hand-written route is an operation of the contract |
| UR-P4-api-267 | note | fixed `b650950` (decision 79), `b217f6c` (decision 82) | 501 and 503 only where the operation declares them, 500 elsewhere; a shared 500 on every controller operation; the status guard sees every answer |
| UR-P4-ui-290 | low | fixed `03920ac` | the update check's error is status text on `--fui-error-text`; a guard holds the rule over every page |
| UR-P4-ui-291 | low | fixed `75497bc`, `a57886b`, `1e927a7` (decision 76) | `Problem.code` `icd_multi_admin`, `fabricui.ErrICDMultiAdmin` wrapping `ErrConflict`; the page asks only on that code |
| UR-P4-ui-292 | low | fixed `8d66c8a` | a node row follows the host's `controllerNode` route |
| UR-P4-ui-293 | low | fixed `b8a70c3` | the controller node page follows a new tab for the same node |
| UR-P4-ui-294 | note | fixed `53898a9` | the fabric removal is busy while one is in flight |
| UR-P4-test-295 | low | fixed `a70134d` | the contract guard reports a new required request field |
| UR-P4-test-296 | low | fixed `a70134d` | the contract guard reads the members of an inline object |
| UR-P4-ui-297 | note | fixed `d3aad5b` | a node label's 64 counted in characters, as the server counts |
| UR-P4-ui-298 | note | fixed `6517c54` | the erase poll ends with the Fabrics page |
| UR-P4-ui-299 | low | fixed `5d6f5bc`, `838b4e9` (decision 77) | button labels reach 4.5:1 on their fill in the four built-in palettes; a guard test computes the ratios |
| UR-P4-ui-300 | low | fixed `5d6f5bc` (decision 77) | an empty state's description is secondary text |
| UR-P4-ui-301 | low | fixed `9a93e50` | the diagnostics load the node list for the peers' names |
| UR-P4-ui-302 | low | fixed `b0e5928`, `cc7c096`, `063b028` | the adapter leaves an epoch above the int64 range out (`epoch_ms` absent); the page shows a time outside the Date range as its number |
| UR-P4-ui-303 | note | fixed `9a93e50` | a session's peer links only to a node the controller holds |
| UR-P4-ui-304 | note | fixed `e9648d1` | a column key named twice is shown once, the first standing |
| UR-P4-ui-305 | note | fixed `5bf805c` | a failed topology reload keeps the drawn graph and toasts |
| UR-P4-ui-306 | note (re-rated from low) | fixed `f2031d1` | the endpoint tree lists every entry once, an orphan and a cycle too (hardening: go-fabric v0.23.0 cannot deliver the input) |
| UR-P4-test-320 | low | fixed `8532f30` | the node stream route's own admin test, over HTTP without the token |
| UR-P4-test-321 | note | fixed `8532f30` | the non-admin controller status checked member by member |
| UR-P4-test-322 | note | fixed `6effafe` | six client rules held by a test each |

## Pre-merge read

The verifier resumed, over `792d15b..d7bc3f3` (before the rebase).

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-doc-390 | low | docs `0a4b7c1` | `docs/contract.md`'s controller route table lists the shared 500 |
| UR-P4-api-391 | note | **open, note** | decision 79's rest: a host's own `ControllerSource` can still answer 409, 502 or 404 on an operation that does not declare it (13 answers by a temporary guard case); the shipped adapter only through `Ping` under a host's request deadline. Carried to the next review of the hosts (go-fabric-server's plan: a host source's error kinds), or a go-fabric-ui issue |
| UR-P4-test-392 | note | fixed `bd3a294` | a case with a live request and a `DeadlineExceeded` from the fake node expects `fabrics_failed` |
| UR-P4-adpt-393 | note | fixed `0a4b7c1` (godoc), `bd3a294` (test) | the `Hooks` godoc names leave and forget; removed → added → removed on one local id is tested; the growth of `gone` stays (a short string per removed node, the verifier's call) |
| UR-P4-test-394 | note | fixed `bd3a294` | the status guard's unnamed-path branch has a test |
| UR-P4-doc-395 | note | docs `0a4b7c1` | three changelog phrases match the code |
| UR-P4-ui-396 | note | **open, note** | the not-in-effect notice is cleared by any `surface_changed`, possibly one of a failed rebuild (I: whether a failing assembly raises the observer at go-fabric's tag was not read). Carried as a go-fabric-ui issue (clear on `endpoints_changed` only, once the observer's behaviour is read) |

## Second look: go-fabric v0.24.1 and #81

The verifier resumed, over `f1d2321..5dcf396` (the lift's six commits).

| ID | sev | disposition | where |
| --- | --- | --- | --- |
| UR-P4-bound-400 | low | fixed `dbb51ab` | the adapter takes the smaller of the port's `CommissionTimeout` and a positive one of the host's `Pair` template; godoc and changelog say so |
| UR-P4-test-401 | low | fixed `dbb51ab` | a discovered commissioning with a template naming a BLE central and an address |
| UR-P4-test-402 | note | fixed `dbb51ab` | each failed node carries exactly its own product's cause |
| UR-P4-test-403 | note | **open, note** | no test lets go-fabric's real update manager fail one product and reads the route's `errors` and the single `check_finished`; the seam test stands. Carried to the next review of the hosts (an image source that refuses the download is a host's fixture), or a go-fabric-ui issue |
| UR-P4-err-404 | note | **open, note** | `PairDiscovered`'s `ErrConfiguration` refusals (no address, a closed controller) answer 400 `invalid_argument` on the discovered path, where the registry path answered 502. Whether `Discover` can return a device without an address was not read (I). Carried as a go-fabric-ui issue (or a go-fabric question, if the error should be split) |
| UR-P4-adpt-405 | note | fixed `dbb51ab` | `Hooks.nodeID` resolves only a commissioned node |
| UR-P4-err-406 | note | fixed `dbb51ab` | go-fabric's `softwareupdate.ErrNoNode` on a consent is `ErrNodeNotFound`, 404 |
| UR-P4-doc-407 | note | fixed `dbb51ab` | the consent texts name the condition (the manager knows the running version); `Check` returns the whole error when no product names a node; contract: the 204's description only |

## Open owner question

- **The consent route's 204 for a version at or below the running one.**
  go-fabric v0.24.0's `AddUpdateConsent` keeps no consent to a version at
  or below the one the node runs (while it knows that version) and answers
  `false`, without error — indistinguishable from a consent it keeps for a
  node not eligible for updates. The route keeps answering 204 and its
  description, the adapter's godoc and the changelog say what the manager
  keeps (`34a897a`, `dbb51ab`): status quo, no behaviour change. A 400 for
  such a version would need the adapter to tell the two `false` cases
  apart (or go-fabric to) and is an owner decision; not taken in this
  pull request.

## Measured cost

| Agent | Model | Tokens |
| --- | --- | --- |
| S-go | Fable | ~300 k (estimate) |
| S-web | Fable | ~200 k (estimate) |
| G Go rest (incl. coverage follow-up) | Opus | ~225 k (estimate) |
| W web rest (incl. coverage pass) | Opus | ~260 k (estimate) |
| verifier | Fable | ~190 k (estimate) |
| gap read (the verifier, resumed) | Fable | ~110 k (estimate) |
| pre-merge read (the verifier, redone) | Fable | ~190 k (estimate) |
| second look (the verifier, resumed) | Fable | ~150 k (estimate) |
| fix G, fix W, the v0.24.1 lift | Opus | not in the reports |
| **sum of the reports' figures** | | **~1 625 k** |

Every figure is the agent's own estimate; the harness's exact numbers
went to the coordinator and are not recorded here.

## Guards

At `d7bc3f3` (before the rebase) green, as the pre-merge read states;
at `5dcf396` (after the go-fabric v0.24.1 lift) and at `dbb51ab` (after
the second look's fixes) the full local guard set exits 0: `make ci`,
`make race`, `make cover-check`, `make vuln licenses secrets`,
`GOARCH=386 go build ./...`, `make test-386`, `make notices-check`,
`make version-check`, `make contract-check`, `make web-check dist-check`,
`make e2e`, `make e2e-web`.

## Differences from the brief

- Decision 75 planned `v0.11.1` with the contract unchanged; UR-P4-ui-291
  needed a problem code (decision 76), so contract 0.12.0 and `v0.12.0`;
  decision 82 added the shared 500 to the same minor.
- UR-P4-ctr-196's value is `other`, not the brief's example `unknown`.
- UR-P4-ui-302's adapter half leaves the epoch out rather than clamping it.
- Decision 80 named go-fabric `v0.24.0`; the pin is `v0.24.1`, the last
  version under `github.com/SukramJ/go-fabric` (moved to
  `thingfabric.dev/go-fabric` from v0.25.0). The path's deprecation is a
  `go.mod` notice, not an identifier; no deprecated identifier is used
  (decision 62, `make lint`).
- UR-P4-adpt-197's fix (`b251e4a`) was superseded in the same pull
  request by go-fabric#555's errors (`3424872`).
- A second look by the verifier after the lift (IDs 400–407), outside the
  brief's 390–399; the verifier resumed, so the cap held.

## Hosts' obligations (0.12.0)

Additive over 0.11.0; `changelog.md` 0.12.0 ("Hosts") and
`docs/contract.md` ("Versioning"). go-fabric v0.24.1 is the minimum a host
builds with (its own v0.24.0 breaks, the store schema `v12` among them,
are the host's at its pin lift); `fabricui.PairOptions.CommissionTimeout`
is new, and a shorter host bound wins over the handler's 600 s; a client
tolerates `icd_multi_admin`, `window_kind` `other` and a 500 on every
controller operation.

## Carried after this pull request

UR-P4-api-391, UR-P4-test-403 → the next review of the hosts
(go-fabric-server's plan), or a go-fabric-ui issue; UR-P4-ui-396,
UR-P4-err-404 → a go-fabric-ui issue. The consent route's 400 → the
owner. Wave U12 (go-fabric-ui#68) follows `v0.12.0` in a fresh session
(decision 81).

Counts: 3 carried rows (all fixed), 37 run-2 findings (0 high, 0 medium,
18 low, 19 note), 7 pre-merge findings (1 low, 6 note), 8 second-look
findings (2 low, 6 note). By status: fixed 46, docs 5, open 4 (notes),
partly 0, struck 0.
