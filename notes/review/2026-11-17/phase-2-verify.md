# Phase 2 — verification

Public copy; the parts on held rows are left out (decisions 5, 14, 18, 25). Verifier
(Fable), HEAD `1056db2`, go-fabric `v0.22.2` (`gf:`), hosts `srv:` `d61a545`,
`ha2f:` `ec7deea`. Reports read: `held/phase-2-X1.md` … `-X4.md`; registers
of phases 0, 1a, 1b; the held rows.

**M** = re-read at the line and/or re-run; **I** = inferred. Re-run in the
reviewers' scratch worktrees `p2wt/x1…x4` (all detached at `1056db2`): every
Go test `TestX1_*`, `TestX2_*`, `TestX3_*`, `TestX4*` and the two vitest
files (`x1-pairing.test.ts`, `x4-keys.test.ts`). Every test the reports call
red is red, every one they call green is green, with the reported texts.
**Not re-run:** X4's Playwright specs (Chromium, Firefox) — their browser
results are taken from the report; the handler half of each was re-run in Go.
No probe of my own was needed; nothing was written outside this file.

## 1. Findings verified

| ID | reviewer sev | verdict | my sev | deciding lines | M/I | contract | held |
| --- | --- | --- | --- | --- | --- | --- | --- |
| UR-P2-conc-04 (X1) | medium | **confirmed; UR-P2-conc-32 merged into it** | medium | `fabricui/nodeswitch.go:128` (`Observe` on the target of that moment), `fabricui/handler.go:646-647` (once per stream), `fabricui/stream.go:89-95` (ends on the request or `h.done` only); `ui/src/lib/pages/node/Fabrics.svelte:104-115` (poll, `loadStatus`, `loadFabrics` — no `resync`, no reconnect), `ui/src/lib/store.svelte.ts:445-473` (only `resync` judges the flow); hosts `srv:internal/daemon/nodes.go:296`, `ha2f:cmd/ha2fabric/nodes.go:136` | M (read + both Go tests + vitest red) | no | no |
| UR-P2-conc-32 (X2) | low | **merged into UR-P2-conc-04** (same defect, the server half) | — | as above | M | no | no |
| UR-P2-adpt-30 (X2) | medium | **confirmed** | medium | `fabricui/gofabric/node.go:585-588` (`statuses[0]` before the error is looked at); `gf:client/interaction/write.go:259-275` (REPLACE-ALL, then one append per element), `:191-198` (one status per path), `:23` (only ACL and Extension are not chunked); `ui/…/controller/node/Bindings.svelte:51`, `Attributes.svelte:104` | M (read + test red: status 0, 255 of 256) | no | no |
| UR-P2-err-02 (X1) | low | confirmed | low | `fabricui/handlers_pairing.go:104` (`win.Discriminator != eph.creds.Discriminator` forgets); `gf:node/commissioning.go:278-284` (`setAnnounced` only after a publish that worked), `:373-375`; `gf:bridge/bridge.go:2352-2357` | M (read + test red: `discriminator 0`, source `configured`) | no | no |
| UR-P2-ui-05 (X1) | low | confirmed | low | `ui/src/lib/store.svelte.ts:281` (`shown.kind = "own"` after every POST), `:293-297` (`otherWindow`), `:462-468`; `fabricui/handlers_node.go:192` (`basic` for a fresh window) | M (read + vitest red: `idle`, 1 read) | no | no |
| UR-P1-err-22 (X2, assigned) | low (was note) | confirmed | low (borderline note) | `fabricui/handlers_node.go:400` (`erase(r.Context())`) against `:351` (`detached`); `fabricui/problem.go:105-111` (debug only); both pinned hosts detach themselves (`srv:…/nodes.go:259`, `ha2f:…/nodes.go:226`) | M (read + test red) | no | no |
| UR-P2-err-31 (X2) | low | confirmed | low | `fabricui/registry.go:334`; `ha2f:cmd/ha2fabric/nodes.go:555`, `:557` (request's context), `:632-646` (`leave` warns and counts), `:564-571` (files and record go anyway); `srv:internal/daemon/registry.go:300` detaches | M handler (test red), M host (read); the decommission half I | no | no |
| UR-P2-bound-33 (X2) | low | confirmed; **not** UR-P0-bound-29 | low | `fabricui/stream.go:89-121` (no write deadline; the heartbeat is a write in the same loop); `ha2f:internal/web/web.go:93-99` (comment rests on the heartbeat), `srv:internal/daemon/web.go:128` | M (read + test red: 1 subscription after 100 heartbeats) | no | no |
| UR-P2-auth-60 (X3) | low | **facts confirmed, lowered** | note | `fabricui/handlers_controller.go:519-536`, `:542-552`; neither route in `controllerAdmin` (`:55-78`) | M (both tests red, memory controller and real adapter) | yes | no |
| UR-P2-doc-62 (X3) | low | confirmed | low | `docs/embedding.md:35`, `:362-365`; `docs/contract.md:367` and `fabricui/options.go:219-220` ("every POST, PUT and DELETE") against `controllerAdmin` (`handlers_controller.go:55-78`: `PUT …/attributes` is there, `POST …/attributes` is not); `srv:internal/daemon/run.go:174`, `ha2f:cmd/ha2fabric/main.go:184` ("read-only") | M (read) | no | no |
| UR-P2-ui-90 (X4) | low | confirmed | low | `ui/src/lib/kit/EventList.svelte:28`, `…/controller/node/Events.svelte:34-38` (append, no merge), `Endpoints.svelte:36`, `kit/AttributeTable.svelte:31`, `kit/ClusterView.svelte:47`; `fabricui/gofabric/node.go:171-183`; `gf:controller/client/cluster.go:192-204` (sorted, repeats kept), `gf:controller/client/event_history.go:52-66` | M (7 vitests red, Go test red); browser run not repeated | no | no |
| UR-P2-web-91 (X4) | low | confirmed | low | `fabricui/spa.go:42-59` (anything that is no file → the index), `ui/vite.config.ts:46` (`base: "./"`) | M (read + test red for `/a/`, `/a/b`, `/assets/`, `/node/status`) | no (see section 4) | no |

Sample: 9 lows re-read (X1 2 of 4, X2 3 of 5, X3 2 of 2, X4 2 of 3, one
held candidate on top).

## 2. The mediums

### UR-P2-conc-04 = UR-P2-conc-32 — confirmed, medium; ID 04 survives

Ran: `TestX1_StreamFollowsTheNodeAcrossAnErase` (no `fabric_added`; pairing
answers `configured 20202021` while the page holds `75500059`),
`TestX2_StreamAcrossEraseSeesTheNewNode` (no `commissioning_complete`,
30 s), the vitest "the erased node's window codes leave the page" (red).

Tried to refute:
- *The hosts end the stream themselves.* No. Both keep the listener and the
  handler and call `Set` (`srv:…/nodes.go:296`, `ha2f:…/nodes.go:136`).
  Only go-ha2fabric's registry nodes list `Source` nil during the erase,
  which rebuilds the handler and ends that node's streams (X2's coverage
  note); the default node of both hosts and go-fabric-server's registry
  nodes are affected.
- *The page reconnects anyway.* Only when the tab is hidden and shown, or
  the shell switches family (`store.svelte.ts:476-493`, `stream.ts:158-170`).
  Going from Fabrics to Pair inside the node family keeps the stream. The
  erase flow calls `loadStatus` and `loadFabrics`, not `resync`
  (`Fabrics.svelte:113-115`), and `loadStatus` does not judge the flow.
- *It is documented, so accepted.* `nodeswitch.go:25-29` and
  `docs/embedding.md:157-160` say "until the page reconnects" — the premise
  that it does is what is false. That is a wrong statement, not an accepted
  limit.

Severity: the two reviewers disagree (X1 medium, X2 low). Medium stands:
deterministic after every erase on both pinned hosts, and it breaks the one
flow that follows an erase — pairing the new node never reaches `success`,
the fabric list and status stay, and an open flow keeps showing codes of a
node that no longer exists. Against it: nothing is lost and a reload heals
it; low would be defensible if the owner reads the comment as intent.

Merge: X1's row carries both halves and both tests; X2 adds the health and
trace observers through the same switch, the second fix option (the switch
moves its registrations) and its Go test. Fix advice: `Set` ends the streams
of handlers over that switch (the browser's `EventSource` reopens, the
store's `resync` runs) **and** the erase flow calls `node.resync()`; take
UR-P2-ui-05 in the same change, because `resync` runs straight into it.
Moving the registrations alone leaves the page's stale flow. The two doc
sentences are reference statements (CLAUDE.md model table: a Fable fact
note). Contract: no.

### UR-P2-adpt-30 — confirmed, medium

Ran: `TestX2_ChunkedListWriteReportsLaterFailure` — status `0x0`, 255 of
256 labels.

Tried to refute:
- *go-fabric returns an error when one path fails.* No: `Write` returns the
  statuses and a nil error (`gf:client/interaction/write.go:191-204`); the
  check is the caller's (`WriteStatuses.Err`, `:82-89`), which the adapter
  does not call.
- *The first status is the failing one.* For a chunked list the first is the
  REPLACE-ALL with the empty list (`:266-268`), which succeeds.
- *Only an exotic node gets there.* The ordinary case is worse than the
  test's: a binding list with one target the node refuses, or a full
  binding table. REPLACE-ALL has already emptied the list, the appends
  before the refused one land, the page toasts "saved". The Bindings tab
  reloads and shows the partial list; the Attributes tab does not.

Medium stands (success shown on a write that destroyed part of the list;
precondition: the node refuses one element). The ACL is not affected
(`write.go:23`). Fix as proposed: the first non-success status; an empty
status list with a nil error is `ErrNodeFailed`. `WriteStatus.status`
already exists in the answer — contract: no.

### Held candidates

Three held candidates of X4 were read; none is a row of its own. They are
published with the pull request that fixes the rows they belong to.

## 3. Duplicates, merges, re-reports

| Finding | Against | Result |
| --- | --- | --- |
| UR-P2-conc-32 (X2) | UR-P2-conc-04 (X1) | **same defect → merged, 04 survives** |
| UR-P2-leak-63 (X3) | UR-P1-leak-66 (open → phase 3) | **re-report**: same site (`gf:controller/diagnostics.go:250-270` → the ungated trace), same two fix options. New is only the measurement and that decision 20's wording names the log, not the traces. Merge into row 66: its test becomes row 66's red test, the wording point goes to phase 3's wizard. No new row |
| UR-P2-bound-33 (X2) | UR-P0-bound-29 | **distinct**: 29 is the missing cap on open streams (a new 503); 33 is a stream that never ends for a reader that stopped reading. A cap would not fix 33, a write deadline does not fix 29. Own row |
| UR-P2-auth-60 (X3) | UR-P0-auth-26 | **distinct, but not a defect of the gate**: row 26 is "four routes make the controller *send*", and its own proposal offered "serve ACL, bindings and fabrics to a non-admin from the last read, or gate them". Decision 19 chose the gate; nothing at HEAD implements it. That the cached image shows the same lists is a decision-8 disclosure class (X3's I-b), the same kind as UR-P2-leak-61, which X3 rates note. Lowered to note; own row, to the wizard with decision 19 |
| UR-P2-conc-01 (X1) | UR-P1-err-198 (fixed), UR-P1-err-04 | distinct: 198 made the page react to `window_opened` and a resync; 01 is that the server sends no event for open → open (`fabricui/events.go:281-293`). A new path, own row |
| UR-P2-err-02 (X1) | UR-P1-err-04 (decision 28) | distinct cause (the handler's own window forgotten on a discriminator of 0), same symptom; own row. A fix that identifies the window by `OpenedAt` and kind also narrows 04 |
| UR-P2-conc-07 (X1) | UR-P1-conc-05 (fixed) | same family, other pair (open/open); own row |
| UR-P2-err-31 (X2) | UR-P1-err-22 | same class, other route; own row, one fix (`detached`) |
| UR-P2-doc-62 (X3) | UR-P0-doc-22 (open → phase 3) | distinct (the converse: what a false `Admin` serves); same paragraph, same pull request |
| UR-P2-bound-65 (X3) | UR-P0-bound-29 | distinct site (status polling, not streams) |
| UR-P2-ui-90 (X4) | UR-P1-ui-130, -131, -139 (fixed) | same class, other lists; own row. The standing guard of that class did not cover them: phase 2's fix should extend the guard (every keyed `each` over data of a node), not only the five lists |
| UR-P2-ui-05 (X1) | UR-P1-err-198's fix `24e3626` | a regression of that fix; own row |

No finding of the four reports repeats a register row unchanged except
UR-P2-leak-63.

## 4. Contract (decision 13) and disposition, all findings

"p2" = this phase's pull request. Go half and web half are disjoint paths.

| ID | sev (final) | disposition | contract | note |
| --- | --- | --- | --- | --- |
| UR-P2-conc-04 | medium | **fix in p2** (Go + web) | no | with conc-32; correct `nodeswitch.go:25-29`, `docs/embedding.md:157-160` |
| UR-P2-conc-32 | — | **merged into 04** | — | |
| UR-P2-conc-01 | low | **fix in p2** (Go) | no | a second `window_opened` is within the event's meaning; one sentence in `docs/contract.md`. If the owner reads that sentence as a contract line, it is the only p2 item for the wizard |
| UR-P2-err-02 | low | **fix in p2** (Go) | no | |
| UR-P2-adpt-03 | low | **fix in p2** (Go, adapter) | no | a 200 where 500/502 was answered for a commissioning that happened; both statuses are in the contract. Not sampled |
| UR-P2-ui-05 | low | **fix in p2** (web) | no | the no-contract variant; `window_kind` in the window answer → phase 3 with decision 28's field |
| UR-P2-ui-06 | note | **phase 3** | yes (`window_remaining_seconds`) | the `Date`-header variant would be p2, but it reads a header a proxy may rewrite |
| UR-P2-conc-07 | note | **fix in p2** (Go) | no | inferred: needs its red test first, else struck (decision 17) |
| UR-P2-leak-08 | note | **fix in p2** | no | server part (forget on the close transition); the page parts are small |
| UR-P2-err-09 | note | **upstream go-fabric** | no | the commissioner does not disarm on a cancelled context |
| UR-P2-adpt-30 | medium | **fix in p2** (Go, adapter) | no | |
| UR-P1-err-22 | low | **fix in p2** (Go) | no | port doc and `docs/embedding.md` say "detached, two minutes" |
| UR-P2-err-31 | low | **fix in p2** (Go) | no | host observation to go-ha2fabric's review (`fleet.Delete`) |
| UR-P2-bound-33 | low | **fix in p2** (Go) | no | `SetWriteDeadline` per round, `ErrNotSupported` ignored |
| UR-P2-conc-34 | low | **docs in p2**; `Options.Done` → phase 3 `api` | no (docs) | host observation to both hosts' reviews. Not sampled |
| UR-P2-ui-35 | note | **phase 3** | yes (field) | |
| UR-P2-adpt-36 | note | **fix in p2** (Go) | no | inferred: red test first, else struck |
| UR-P2-err-37 | note | **upstream go-fabric** (facade question), then phase 3 | status 502 → 400 for that case | `controller/client` at `v0.22.2` exports neither `ErrRequestTooLarge` nor `ErrMalformedRequest` (grep) — a facade gap, not to be worked around |
| UR-P2-auth-60 | note (lowered) | **phase 3** (wizard, with decision 19; class I-b) | yes | |
| UR-P2-leak-61 | note | **phase 3** (wizard, class C); the `Nocs` question → go-fabric `cite` | yes if filtered | |
| UR-P2-doc-62 | low | **phase 3**, with UR-P0-doc-22 | no | its text lists the classes as decided, and the two wrong sentences become true with decision 19 — writing it now would be written twice |
| UR-P2-leak-63 | — | **merged into UR-P1-leak-66** → phase 3 | yes if gated | |
| UR-P2-ctr-64 | note | **phase 3** | yes | |
| UR-P2-bound-65 | note | **docs in p2** | no | one sentence in `docs/embedding.md` |
| UR-P2-ui-90 | low | **fix in p2** (web) | no | plus the guard of the class |
| UR-P2-web-91 | low | **fix in p2** (Go) | no | bundle paths are not in `contract/openapi.yaml`; but a 200 becomes a 404 or a redirect and `docs/embedding.md:374` ("every deep path") changes — say it in the changelog for the hosts |
| UR-P2-test-95 | note | **p2 in part** | no | the framing guard (T10), the no-CORS guard (T8/T9) and `TestX4_Guard_StreamDataStaysOnOneLine` can go in now. |

Struck: none. The red tests that belong to held rows are not findings; they
stay out of the repository until phase 3.

## 5. X3's decision-8 classes — the facts under three of them

Policy not judged. Sample: A-b, J, L.

| Class | Row it rests on | Checked | Result |
| --- | --- | --- | --- |
| A-b node configuration | `status.config` carries `store`; both hosts pass a path | `fabricui/handlers_node.go:82-93` (`Store` on the wire), `:145-149` (copied without a look at the caller's rights); `srv:internal/daemon/web.go:87` (`Store: s.Store`), `ha2f:cmd/ha2fabric/main.go:178` (`Store: cfg.StorePath`) | holds |
| J controller diagnostics and topology | `ssid`, `bssid` on `GET /controller/topology` | `internal/api/api.gen.go:1558-1559`, `:1596-1597`; `fabricui/handlers_controller.go:1641`; `fabricui/gofabric/topology.go:182-183`, `:196`; the route is not in `controllerAdmin` and only `refresh=true` is gated (`:83-85`) | holds |
| L credential names | `GET /controller/credentials` ungated, `name` is the SSID or the Thread network name | `contract/openapi.yaml:1652-1655` (no `x-admin` on the GET; `:1667`, `:1683` on PUT and DELETE), `fabricui/handlers_controller.go:1438-1456`; name source `fabricui/memory_controller.go:60`, `:63`, adapter `fabricui/gofabric/gofabric.go:100-103` (Thread `NetworkName`) | holds; the adapter's Wi-Fi name was not traced to its line |

One wording point for the wizard: class I-b's "otherwise decision 19 is half
done" is the reviewer's reading. Row 26 and decision 19 are about the
controller *sending*; whether the cached lists are disclosed is a separate
answer (section 3).

## 6. False-positive rate

- Mediums (4 reported, 3 distinct): 0 struck, 0 lowered. One merged
  (conc-32 → conc-04), one folded into a held row.
- Sampled lows (9): 0 struck, 1 lowered to note (UR-P2-auth-60: facts right,
  class wrong). By reviewer: X1 0 of 2, X2 (Opus) 0 of 3, X3 1 of 2 lowered,
  X4 0 of 2.
- Re-reports: 1 (UR-P2-leak-63, a note, said by the reviewer to extend
  row 66).
- Citations: all deciding lines re-read were right; `docs/embedding.md:362-365`
  in doc-62 is the paragraph's first four lines (it runs to `:365`), fine.

Final counts after merges and folds: 30 reported → 25 rows for the register
(2 merged, 3 folded into held rows): 0 high, 2 medium (conc-04, adpt-30),
11 low, 12 note. Fix in p2: 17 (2 of them docs only, 1 in part); phase 3: 6;
upstream go-fabric: 2.
