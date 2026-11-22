**No held candidate: the verifier found no high and no medium in run 2. It lies in `held/` beside the unit reports until the coordinator publishes them.**

# Phase 4, run 2 — verification

Verifier (Fable), 2026-11-19. HEAD read `792d15b`, go-fabric `v0.23.0`
(`gf:`). Reports verified: `phase-4r2-S-go.md`, `phase-4r2-S-web.md`,
`phase-4r2-G.md`, `phase-4r2-W.md` (final, 11 low). Red tests re-run in
the reviewers' scratch worktrees (`<scratchpad>/p4r2wt/<unit>`), focused
runs only; logs in `<scratchpad>/verify/`. M = re-measured here, I = read
or inferred.

**Result: 0 high, 0 medium.** 12 lows verified of 17 (all of S-go, S-web
and G; 6 of W's 11): 11 confirmed, 1 re-rated down (ui-306 → note). One
note re-rated up (ui-233 → low). Nothing struck. README "Done" holds for
this run: no new high or medium.

## Verdicts

| ID | reported | verdict | ground |
| --- | --- | --- | --- |
| UR-P4-leak-200 | low | **confirmed, low** | M: red again (23 of 25 pairs wrote both lines; the half-close races the read, so not always 25). `handlers_node.go:229-235`, `:710-728`: a cancelled request context is stored as the node's state. It is UR-P4-leak-91's reach, not a new class — the fix `f20b416` is bypassed by a client that abandons every other request — and it gains nothing beyond leak-91 (log lines in the admins' ring or on disk, no secret, no action). Same severity as leak-91. Fix as proposed (`serverError`'s abandoned rule) |
| UR-P4-adpt-201 | low | **confirmed, low; a new row** | M: red again (no `node_removed` in 3 s; `connection connection node_updated` under the old id). Not UR-P4-adpt-31: that row is `Wrap`/`chain` losing the **host's** `Deleted`; this is the adapter's own `For` (`hooks.go:197-217`) never subscribing it. Run 1 named it only as "optional, separate" in `phase-4-G.md:21`; no register row carries it. Names adpt-31 as the brief asks. Not medium: at v0.23.0 a commissioned node leaves without `Decommissioned` only through `Registry.Forget`/`Erase` (`gf:controller/client/registry.go:477`, `:625-637`); go-fabric calls `Forget` only at start (`gf:controller/ota_provider.go:267`), the adapter never, the hosts' checkouts never (grep, unpinned, I). The other `deleteNode` sites are a decommission (hook runs first, `fabricops.go:277`, `:406`), a peer's leave (`reactions.go:141-147`), and nodes never announced as added (`complete.go:88`, `addresses.go:119`). Effect is a stale row in the page, no right |
| UR-P4-doc-202 | note | **confirmed, note** | Both sides read (`docs/embedding.md:470-472`, `host.go:49-76`). Fails closed, seen at the first request; the paragraph's first sentence still names `AllowedHosts`. Not the class of run 1's held mediums (those misled into an open setup) |
| UR-P4-conc-230 | low | **confirmed, low** | M: three cases red. `store.svelte.ts:148-168`, `:364-366`. Every reader of `statusLoading` tests it with `status === null` (`Selection.svelte:48`, `:57`, `Fabrics.svelte:37`, `Status.svelte:54`, `Operation.svelte:103`), so the stuck flag shows only in case 2 (the hydrate's read fails after the overtaken read succeeded) and heals at the next `loadStatus`. Case 3 (stream not reopened on a new token) needs a restart plus a hydrate inside the resync's read and heals the same way: a narrower rest of UR-P3-web-191, not its medium |
| UR-P4-err-231 | low | **confirmed, low** | M: red ("never settled"). `client.ts:495`, `:506-524`: the refresh's own read awaits `refreshing`. The shipped handler never answers a status route `forbidden`, and without a token there is no refresh, so it needs token mode behind a layer that answers in the contract's problem shape. The hang is permanent for the page once hit; reach keeps it low |
| UR-P4-ui-233 | note | **re-rated: low** | Read: `stream.ts:312-317` ends for good, `store.svelte.ts:561` sets `streamDenied`, which nothing in `ui/src` reads for the node store; `App.svelte:333-335` shows the controller's only under `!controller.isAdmin`. A user-visible defect with a precondition (host mode, the admin right lost while the controller family is shown): events stop, nothing says so, and the stream does not come back with the right. No right gained, so not medium. Needs a red test in the fix wave (none written; the brief asks one for a low) |
| UR-P4-ui-232, -234, web-235 | note | not sampled | — |
| UR-P4-ctr-260 | low | **confirmed, low** | M: red under `GOARCH=386` (400 "overflows destination of type: int"). `contract/openapi.yaml:1862-1865`, `handlers_controller.go:1542`. Only `device_type` can exceed 2^31 (discriminator and vendor are 16 bit). No wire change |
| UR-P4-err-261 | low | **confirmed, low** | M: red (500 "removed 0 of 2"). `handlers_node.go:414-424` against `gf:node/factory_reset.go:78`, which passes over `ErrFabricNotFound`. Not medium: the race window is one list-to-remove span, the answer says "incomplete" and is logged, and a second reset completes |
| UR-P4-ctr-262 | note | confirmed, note | M: red under 386; order only |
| UR-P4-api-267 | note | **confirmed, note** | `controller.go:942`, `:948`, `handlers_controller.go:211-245` map 501/503 on every controller route; the contract declares 501 on four controller reads (and the node erase). Only a host's own `ControllerSource` can reach it; adapter and memory controller do not. A godoc line, not a defect of a shipped path |
| UR-P4-doc-263, -264, test-265, -266 | note | not sampled | — |
| UR-P4-ui-291 | low | **confirmed, low; contract yes** | M: red. `Icd.svelte:40-44` asks on any 409; `gofabric.go:783-790` maps four errors to `ErrConflict`. See "ui-291" below |
| UR-P4-test-296 | low | **confirmed, low** | M: three script cases red (295's one and 296's two). `contract-guard.mjs:67-68` checks removal at a component's top level only; `compareNode` `:131-134` skips a property that is gone and never looks at `required`. A guard gap, no break today |
| UR-P4-test-295 | low | confirmed, low | M: red; `:86` requires the field to have existed |
| UR-P4-ui-299 | low | **confirmed, low; a new row, not decision 31's shape** | M: red. Decision 31 is about status **text on the card** and keeps the HA colours "for areas"; it says nothing on a label on a fill, and `phase-1b-fix-W2.md:133-134` only recorded the pair as outside it. So a finding, and an owner question (the fix changes the fill or the label colour of every primary button). Under the shell (always `data-skin="ha"`) 3.26:1 and 4.29:1; the 1.87:1 case needs a host on the standalone palette in dark. Same class and severity as UR-P1-ui-172 (low) |
| UR-P4-ui-300 | low | **confirmed, low; not more** | M: red. `EmptyState.svelte:18`. In the denied notice the refusal itself (`common.denied.<code>`) is the message line in the secondary colour and readable; only the hint is in the disabled colour. Nothing is hidden from a role and the notice still says what happened |
| UR-P4-ui-302 | low | **confirmed, low; reachable at v0.23.0** | M: red. go-fabric hands the wire value over unchecked: `gf:im/client.go:678-682` (`uint64Field` only), `gf:client/interaction/report.go:427`; the adapter casts to `int64` (`node.go:431-433`), so a value ≥ 2^63 arrives negative — also outside the Date range. One commissioned node's Events tab, no right: low |
| UR-P4-ui-306 | low | **re-rated: note** | M: red at the page, but go-fabric v0.23.0 cannot deliver the input: endpoints are a map by number (`gf:controller/client/node.go:405-415`, no repeated number), a parts list naming the endpoint itself or the root claims nothing (`gf:controller/client/structure.go:634-640`), longer cycles are refused (`:947-950`, `closesCycle` `:974-982`), and an owner is placed before it owns (`:945`). Only a host's own `ControllerSource` reaches it. Worth the `seen` set as hardening |
| UR-P4-ui-290, -292, -293, -301 | low | re-run red (M), code not read | all four fail as reported in the same run; not judged beyond that |

## ui-291: the contract change

**Yes.** The page cannot tell the multi-admin refusal from the other
three 409s without the server saying so: all four are `409` with code
`conflict`, and the detail text is not a contract. Minimal additive shape:

- `Problem.code` gains `icd_multi_admin` (`contract/openapi.yaml:2217`),
  answered 409 by `registerNodeIcd` only; the status is already declared
  (`:1750`). A response enum value, which a client
  tolerates by UR-P4-ctr-196's docs line (`75988c5`: a minor may add one).
- Port: a new `fabricui.ErrICDMultiAdmin` that wraps `ErrConflict`, so a
  host's `errors.Is(err, ErrConflict)` keeps holding; `controllerError`
  tests it before `ErrConflict`; the adapter returns it for
  `*client.ICDMultiAdminError`, the memory controller for its refusal.
- Page: asks only on `code === "icd_multi_admin"`; any other 409 is a toast.
- `docs/contract.md`: the problem's section.

Consequence for decision 75: this is a contract change of the kind it
names for ctr-196 (0.12.0, `v0.12.0`). To keep `v0.11.1` with the
contract unchanged, ui-291 waits for the next minor; a detail-text match
is not a sound stopgap.

## Notes for the fix wave

- **adpt-201's fix needs one more look (I).** "Emit only while `ids`
  holds the node" is not enough if a hook between `Decommissioned` and
  `Deleted` re-learns the id: on a peer's leave the node still has its
  peer address when `Disconnect` runs (`gf:controller/client/reactions.go:141-147`),
  and `nodeID` (`hooks.go:128-152`) writes `ids` again from
  `reg.Get(local).PeerAddress()`. Then `Deleted` would emit a second
  `node_removed`. A per-node "removed" mark, or a test of both the
  decommission and the leave path asserting exactly one event, closes it.
- leak-200's test is racy by nature (23 or 25 of 25): assert "more than
  one", as it does, not a count.

## Coverage audit against the brief's unit table

- **S-go:** the eleven named files whole, the three `docs/embedding.md`
  ranges, the adapter's facade calls at the pin. Left: `make lint`
  (SA1019, decision 62) not run — checked by grep only;
  `gf:controller/commissioner` not read.
- **S-web:** the eleven named files whole. No gap in the unit's files.
- **G:** every hand-written Go file outside S-go is named as read after
  the follow-up (checked against `git ls-tree 792d15b`). Left: the
  contract against the **node** family's handlers was read as the 0.11.0
  diff plus the status guard, not shape by shape.
- **W:** "Not covered: none" overstates a little. Not read by its own
  lists: `pages/node/Status.svelte:1-19`, `:81-144`,
  `pages/node/NodeManage.svelte:71-111`, `scan/index.ts:1-29`,
  `pages/controller/QrScan.svelte` markup; `qr.ts`, `countdown.ts`,
  `i18n/index.ts` skimmed. **Closed here:** I read those four ranges —
  text through `t()` only, no raw HTML, the scan's generation counter
  and `onDestroy` hold; no finding.
- **Real gap left: the guard tests of the security paths, judged for
  bite by no unit.** S-web left its own tests to W, W read the S-web
  tests "not as targets"; S-go read no test, G read the tests of its
  own files. Unnamed in any report: Go `nonadmin_test.go`,
  `admin_request_test.go`, `options_streams_test.go`, `stream_test.go`,
  `headers_test.go`, `spa_csp_test.go`, `registry_test.go`; web
  `api/client-errors.test.ts`, `api/stream-fetch.test.ts`,
  `api/stream-life.test.ts`, `store-review.test.ts`,
  `store-pairing.test.ts`, `contract-0.10.test.ts`. The code they guard
  was read whole, so this is about a guard that would not bite
  (the class of UR-P4-test-265), not about an unread path. I checked
  only that the tests exist by name (the non-admin filter, the stream
  cap, the admin routes' header, origin and media type).
- Outside every unit by the brief's cut, changed in the delta: `Makefile`
  (+5/-2), `.github/workflows/ci.yml` (+2), `changelog.md`,
  `docs/contract.md` outside `:424-470`. Not read by anyone in run 2.

## Tokens

About 190 k (estimate; the harness reports the exact number to the coordinator).

## Gap read

Verifier, resumed, 2026-11-19. Scratch worktree `<scratchpad>/p4r2wt/V`
(detached at `792d15b`; `ui/node_modules` linked from unit W's; one
untracked test file, left in place; every mutation reverted,
`git status` clean but for those). Still no high and no medium, so no
held candidate. New rows: one low, two notes.

### Guard tests judged for bite

Read whole: `fabricui/nonadmin_test.go`, `admin_request_test.go`,
`options_streams_test.go`, `stream_test.go`, `headers_test.go`. By their
test names plus mutation: `spa_csp_test.go`, `registry_test.go` and the
six web files. A mutation is one changed line, the suite run, the line
restored (`<scratchpad>/verify/mut.py`, `mutw.py`, `mutw2.py`).

Go, `go test ./fabricui`, 17 mutations, **15 killed**:

| Mutation | Killed by |
| --- | --- |
| the host's source also to a non-admin stream (`handler.go:815`) | `TestNodeStream_NonAdminFilter` |
| the non-admin type filter off (`handler.go:806`) | `TestNodeStream_NonAdminFilter` |
| `requestAdmin` without the Origin test (`handler.go:753`) | `TestOpenRoute_ForeignOriginIsANonAdmin` |
| an Origin with userinfo, with a path (`handler.go:765`) | `FuzzOrigin` (seed corpus) |
| host mode without the request header (`handler.go:779`) | `TestAdminRoutes_…` (two), `TestAdmin_HostCheckDecides` |
| under `AllowHost`, any Origin (`handler.go:772`) | `TestAllowHost_OriginUnderAnAdmittedHost`, `FuzzOrigin` |
| the attestation text to a non-admin (`handlers_node.go`) | `TestStatus_AttestationBootReasonAndErase` |
| `config.store` to a non-admin | `TestStatus_NonAdminIsReduced`, `TestRegistry_Fixtures` |
| CSP: `'unsafe-inline'` in `script-src`, `object-src *`, `base-uri *`, `frame-ancestors *` (`spa.go:141-148`) | `TestSPA_IndexCarriesTheCSP` (all four), `TestSPA_FrameAncestors`, `TestSPA_NotBuiltCarriesTheCSP`, `TestSPA_EmbeddedIndexScriptHash` |
| a stream never gives its place back (`stream.go:92`) | `TestStreams_MaxStreams`, `TestStreams_DoneEndsThem` |

Two survived — UR-P4-test-320 and -321 below. By reading, no further gap:
`headers_test.go` walks every operation of the contract and fails under
80 found; `stream_test.go`'s framing test checks every line's kind;
`sibling_test.go` (run 1's guard, read by G) covers a registry node's
handler sharing the cap and the host check.

Web, the whole vitest suite (643 tests), 13 mutations, **7 killed**:
the token to any origin (`client.test.ts`, `token-path.test.ts`), no
request header (four files, 69 tests), a node key unchecked
(`client-errors.test.ts`), a stream's 403 retried and `Retry-After`
ignored (`stream-fetch.test.ts`), the codes kept after a fabric was added
(`store-pairing.test.ts`), the pairing read without admin rights
(`pages/node/Pair.test.ts` — not by the store's own tests). Six survived:
UR-P4-test-322.

### UR-P4-test-320 — low — no test holds the node stream route's own admin test

- **Where:** `fabricui/stream.go:40-41` (`admin := h.requestAdmin(r)`,
  then `subscribeAs(admin, …)`); the guard `fabricui/nonadmin_test.go:109-126`
  calls `h.subscribeAs(false, …)` directly, and every stream test over
  HTTP sends the token (`stream_test.go:137`, `:168`, `:195`,
  `options_streams_test.go:21`).
- **Claim:** the filter is tested, the decision that selects it is not.
  This line is the one place where a request without admin rights on the
  open routes `GET /node/stream` and `/nodes/{node}/stream` becomes a
  filtered stream (decisions 41, 58). **M:** with the line changed to
  `admin := true || h.requestAdmin(r)` the whole `./fabricui` suite
  passes (`ok`, 4.4 s). Today's code is right; nothing would report its
  regression, which hands every event of the handler and the host's
  source to a viewer. The Playwright origin smoke
  (`ui/e2e/origins.spec.ts:42-73`) is a cross-origin read, not this.
- **Test (the proposed guard, green today, red under the mutation):**
  `<scratch V>/fabricui/p4r2v_guard_test.go`,
  `TestP4R2V_NodeStreamRouteFiltersWithoutTheToken` — a stream opened
  over HTTP without the token receives `window_closed` and not
  `fabric_added`. Under the mutation:
  `--- FAIL: TestP4R2V_NodeStreamRouteFiltersWithoutTheToken`.
- **Fix:** take the test into `nonadmin_test.go`; a second case for a
  registry node's route and one in host mode (`Admin` false) cost little.
- **Contract:** no.

### UR-P4-test-321 — note — the non-admin controller status is checked for three of its members

- **Where:** `fabricui/nonadmin_test.go:175-186` (`fabric_id_hex`,
  `node_count`, `label`); `fabricui/handlers_controller.go:371-378`.
- **Claim (M):** a non-admin branch that also sets `NodeIDHex`,
  `VendorID` and `CompressedFabricIDHex` passes the suite. The branch is a
  fresh struct, so a regression needs someone to add a member there;
  hence a note. Decision 57 says "identity zeroed", the test says three.
- **Test (proposed guard, green today, red under the mutation):** same
  file, `TestP4R2V_ControllerStatusNonAdminEveryMemberZero` — every member
  but `enabled`, `capabilities`, `admin`, `csrf_token` is its zero value.
- **Contract:** no.

### UR-P4-test-322 — note — six client rules no web test holds

- **Claim (M, each mutation passes all 643 tests):**
  1. `credentials: "same-origin"` on a call → `"include"`
     (`client.ts:476`; the streams' twin is tested, `stream-life.test.ts:115`).
  2. the 403 retry on any 403, not only code `forbidden` (`client.ts:495`).
  3. the 403 retry without `!opts.retried` (`client.ts:495`); the test
     "does not repeat when the status brings no other token" ends it
     another way.
  4. a stream answer of another media type taken as a stream (`stream.ts:163`).
  5. the resync's fabric read without admin rights (`store.svelte.ts:504`).
  6. the controller stream opened without admin rights
     (`controller-store.svelte.ts:273`).
- **Why a note:** none is the control — the server refuses 5 and 6, the
  token and the request header (the client's two real duties) are held —
  but 1 decides whether cookies go to a base of another origin, and 3 is
  the bound UR-P4-err-231's fix leans on (`retried: true`).
- **Fix:** one assertion each beside the existing cases; 1 and 3 first.
- **Contract:** no.

### The delta outside every unit

Read as `git diff 979a67a 792d15b`:

- **`Makefile`, `.github/workflows/ci.yml`:** `test-386` is the whole
  suite under `GOARCH=386`, a step of the build job after the 386 build;
  `make ci` does not run it and its help line says so. No finding.
- **`changelog.md` 0.11.0:** the entries match the register's rows and
  the code where checked (`currentToken`, `NodeStore.reassemble`, the
  four i18n keys, the 500 of a failed check, "No problem code is new").
  One sentence is UR-P4-doc-202's second site, stronger than the docs':
  "A proxy that replaces it — with a name of its own or the upstream's
  address — makes the check **pass every request**". doc-202's fix should
  say it in 0.11.1's entry; the released entry stays as written.
- **`docs/contract.md` outside `:424-470`** (the route table's three
  rows, "From 0.11.0", the versioning rule, the controller details):
  consistent with the handlers and the contract as the units read them;
  the product key "(vendor, product, software version)" is the carried
  UR-P4-adpt-197's site. No new finding.

### `make lint`

`PATH=$HOME/go/bin:$PATH make lint` in `<scratch V>` at `792d15b`:
gofumpt and tidy checks pass, `golangci-lint run`: **0 issues**, exit 0.
Decision 62 holds at v0.23.0 (no SA1019).

### Gaps still open

None of the two named. Not done: the six web files and
`spa_csp_test.go`, `registry_test.go` line by line (judged by names and
mutation); `gf:controller/commissioner` (unread, S-go).

Tokens for the gap read: about 110 k (estimate).

## Pre-merge read

Reader (Fable), 2026-11-19, redone after the first reader ended without
writing. `git diff 792d15b..d7bc3f3` without `notes/`, `fabricui/dist`,
`internal/api/api.gen.go`, `ui/src/lib/api/schema.gen.ts`,
`ui/package-lock.json`; go-fabric `v0.23.0` at `<scratchpad>/pins/gf`
(`gf:`). Mutations and one temporary test in `<scratch V>` (detached at
`d7bc3f3`, every file restored, the test removed, `git status` clean but
for `ui/node_modules`); no make target, the review worktree untouched but
for this section. The guards at `d7bc3f3` were green before and not run
again. **M** measured, **I** inferred.

**Result: 0 high, 0 medium, 1 low, 6 notes. No held candidate. Ready to
merge;** UR-P4-doc-390 is one table of `docs/contract.md` and fits this
pull request.

| ID | Severity | Where | One line |
| --- | --- | --- | --- |
| UR-P4-doc-390 | low | `docs/contract.md:766-799` | the controller route table still lists no 500 for the 29 operations that declare it since 0.12.0 |
| UR-P4-api-391 | note | `handlers_controller.go:214-245`, `controller.go:936-946` | decision 79's rest: a source's other error kinds still answer statuses an operation does not declare |
| UR-P4-test-392 | note | `handlers_node.go:720` | no test holds the request-context half of the abandoned rule |
| UR-P4-adpt-393 | note | `gofabric/hooks.go:21-22`, `:119-120`, `:135` | `gone` never shrinks, its clearing has no test, the type's godoc still says "node_removed on Decommissioned" |
| UR-P4-test-394 | note | `internal/statusguard/statusguard.go:127` | the guard's new branch (a path no operation names) is held by no test |
| UR-P4-doc-395 | note | `changelog.md:18`, `:70-72`, `:84` | three changelog phrases say more than the code |
| UR-P4-ui-396 | note | `ui/src/lib/store.svelte.ts:487` | the not-in-effect notice is cleared by any `surface_changed`, one of a failed rebuild too |

### The focus points, each read

| Point | Verdict | Ground |
| --- | --- | --- |
| Go security paths | **unchanged** | `handler.go`, `stream.go`, `strict.go`, `host.go`, `options.go`, `spa.go`, `events.go` are not in the diff; the fix wave adds their guards only (test-320, -321, both with bite, below) |
| conc-230 `readStatus` | **holds** | `store.svelte.ts:157-188`: one counter for the three readers, `mine > statusApplied` drops an older answer, a non-owner's failure neither bumps it nor nulls the status, the reader count ends `statusLoading`. Read through the four orders (older first, newer first, owner fails after a newer success, non-owner fails before an older success). M by the first reader: P6, P7, P11 killed |
| err-231 `refreshToken` | **holds** | `client.ts:512-529`: the reload's own reads carry `retried: true`, so they cannot wait on `refreshing`; the retry stays bounded by `!opts.retried` (`:499`). P1, P10 killed |
| ui-233 reopen | **holds, bounded** | node `store.svelte.ts:569-580`, controller `controller-store.svelte.ts:99-108`, `:287-290`. No reopen loop: a reopen needs `streamDenied !== null && !wasAdmin && admin`, `connect` sets `streamDenied` null, and a second refusal finds `wasAdmin` true. With `reopenOnNewToken` in the same answer no double open (the first reconnect clears `streamDenied`). Shown in `NodeFamily.svelte:33-37` and `App.svelte:350-354`. P3–P5 killed. The server never answers the node stream 403 `forbidden` (it filters, test-320), so the node store's path is the host and origin refusals and a proxy's |
| web-235 key check | **holds** | `client.ts:464-468`: the segment after `/nodes/` against `NODE_KEY_PATTERN`; `renameNode`/`deleteNode` encode the key (`:589-590`), so `/`, `?`, `#` arrive as `%…` and fail the pattern, `""`, `.` and `..` fail its first character; `/nodes` itself (list, create) has no segment. P2 killed |
| adpt-201 exactly once | **holds on all three paths** | `hooks.go:117-128`. M: with the `gone` test off only the **leave** case goes red (two `node_removed`), so the verifier's caution was real and the test bites it; with `Deleted` unsubscribed the forget case goes red. Decommission needs no mark (the id is gone before `Deleted`, nothing learns it again). `chain` carries `Deleted` (`:268`). A failed commissioning and a culled commissionable node run `removed` without a known id: nothing is emitted. 8 runs of the test green |
| leak-200 abandoned rule | **hides no store failure** | `handlers_node.go:719-725`: skipped only when the request's context is done **and** the error is a context error; a store error of another kind is logged also for an abandoned request, a context error under a live request too. One rest: UR-P4-test-392. 30 runs of the test green |
| api-267, decision 82 | **holds as decided** | 40 controller operations, all declare 500, 29 did not (M, by script over both contract files); `notImplemented`/`notRead` only at the declaring operations; the only producers are `gofabric/node.go:275`, `:279` (ClusterDetail), `gofabric/topology.go:29`, `memory_controller.go:460`. F5–F7 killed. Rest: UR-P4-api-391, UR-P4-doc-390 |
| ui-291 end to end | **holds** | contract: `Problem.code` + the 409's text on `registerNodeIcd`; `gf:controller/client/icd_errors.go:41` (the error unwraps to `ErrICDOperation`, so the adapter's test comes first, `gofabric.go:775-779`); `controller.go:943` wraps `ErrConflict`; `handlers_controller.go:235-237` before the conflict case; memory controller `:1195`; fixture recorded; `Icd.svelte:42` asks on the code only and never twice (`!allowMultiAdmin`). G4, F4, F8, P8 killed |
| contract 0.12.0 | **additive; `previous/` is v0.11.0** | `git diff v0.11.0:contract/openapi.yaml HEAD:contract/previous/openapi.yaml` is empty. Structural diff by script: no operation, status, schema or field removed; added are the 500s, two response enum values (`icd_multi_admin`, `other`), `x-go-type` on one parameter, descriptions. The 409 of `registerNodeIcd` went from the shared response to an inline one of the same media type and schema |
| `window_kind` `other` | **holds** | `handlers_node.go:748-755`; the pairing handler decides on go-fabric's kinds, not the wire name (`handlers_pairing.go:105`, `:136`, `:273`), so an unnamed kind is "codes do not match", the careful side. F3 killed |
| docs on security paths | **true** | `docs/embedding.md:470-473` against `host.go:47-60`: an unlisted name is refused unless it is a loopback name, an IP literal (the bullet's "upstream's address") or `AllowHost` admits it; `host_proxy_test.go` holds the sentence and the new case. Changelog "Hosts" item 1 says the same |
| ui-299 token design | **sound for the four palettes** | `tokens.css:80-90`, `:120-125`, `:169-180`. The test's cascade model (root, dark, HA, HA dark in source order) is the real one: `html.dark` and `html[data-skin="ha"]` have equal specificity, the HA block comes later and sets all six tokens, HA dark sets none. `Button.svelte` reads only these for a fill; no other label on a fill in `ui/src` (grep `text-white`, `bg-[var(--ha-…-color)]`: none). A host's live theme under the bridge is not covered by the 4.5:1 and the comment says so (`:172-174`) |
| the other fixes | read, no finding | err-261 (`gf:node/factory_reset.go:78`, F2 killed), adpt-197 (`gf:controller/softwareupdate/manager.go:995`, G6 killed), ui-302 both halves (F13 killed), ctr-260/-262, ui-232, -234, -290, -292, -293, -294, -297 (`registry.go:107` counts runes), -298, -301, -303, -304, -305, -306, test-295/-296 |

### Mutations (Go, this read)

`<scratchpad of this session>/mutg.py`, one changed line each,
`go test ./fabricui` (or the adapter's named tests), the line restored.
19 mutants, **16 killed**: G1 the `gone` test off, G2 `Deleted` not
subscribed, G4 multi-admin as plain conflict, G6 product key without the
version, F1 abandoned rule off, F2 reset stops on a fabric gone, F3 an
unnamed window kind as go-fabric's text, F4 the `icd_multi_admin` case
off, F5/F5b 501 and 503 on every route again, F6 the cluster route
without 503, F7 the topology without 501, F8 memory controller plain
conflict, F11 the node stream always admin (test-320's guard), F12 the
non-admin controller status with the node id (test-321's guard), F13 an
epoch above int64 cast. **3 survived:** F1b, G3, F9 — the three test
notes below. The web batch (11 of 11 killed) is the first reader's.

### UR-P4-doc-390 — low — the controller route table omits the 500

- **Both sides.** `docs/contract.md:766-799`: e.g. `/controller/status`
  "200, 403", `/controller/nodes/{id}/ping` "200, 400, 403, 404",
  `/controller/commission` "…, 502, 503". `contract/openapi.yaml`: every
  one of the 40 controller operations declares 500 (M); and the same
  file's "Versioning" paragraph (`:748-752`) says so.
- **Claim.** The table was the exact list per route up to 0.11.0 (the
  rows that declared 500 name it: `:772`, `:775`, `:776`, `:789-792`,
  `:794`); after decision 82 it is stale in the rows of the 29
  operations. A host reading the table for the statuses to handle misses
  the 500.
- **Fix:** 500 in those rows, or one sentence above the table ("every
  row also answers 500"). Contract: no.

### UR-P4-api-391 — note — decision 79's rest: the other kinds

- **Where:** `fabricui/handlers_controller.go:214-245`; the godoc of
  `ErrConflict` (`controller.go:936`, "409") and `ErrNodeFailed`
  (`:944`, "502"), stated for every method.
- **Claim (M):** a temporary test as
  `TestController_NotImplementedAndNotReadOnlyWhereDeclared`, with the
  other kinds, makes the status guard report 13 answers: 409 and 502 on
  `POST …/ping`, on `GET …/endpoints/{ep}` and its cluster route, 404,
  409 and 502 on `PUT /controller/fabric/label`, 404 and 409 on the
  discovery and the topology. Only a host's own `ControllerSource`
  reaches them; the shipped adapter does not, but for one edge (I): its
  `Ping` returns the request context's error (`gofabric/node.go:728-730`),
  so under a host's request deadline the ping answers 502.
- **Why a note:** the class of UR-P4-api-267, which decision 79 closed
  for two kinds on purpose. Either the godoc names the routes per kind,
  or `controllerError` learns the operation's declared set. Contract: no.

### UR-P4-test-392 — note — the abandoned rule's context half is untested

- **Claim (M):** `if true && (errors.Is(err, context.Canceled) || …)` —
  the rule without `ctx.Err() != nil` — passes `./fabricui`. That variant
  is the one that would hide a failure: a read that ends in a deadline of
  its own under a live request would no longer be logged.
- **A rest of the rule as written (I):** a host that puts a deadline on
  the request (`http.TimeoutHandler`) and a fabric read that hangs: the
  context is done and the error is `DeadlineExceeded`, so nothing is
  logged, every time. Before the fix this logged once. The request is
  answered by the host's 503, so it is seen there.
- **Fix:** one case with a live request and a `context.DeadlineExceeded`
  from the fake node, expecting `fabrics_failed`. Contract: no.

### UR-P4-adpt-393 — note — `gone`: growth, an untested branch, a stale godoc

- `hooks.go:120`: every `Deleted` leaves an entry, also for a culled
  commissionable node and a failed commissioning; `:135` is the only
  removal. The registry never hands a local id out twice
  (`gf:controller/client/registry.go:231-238`), so the entries stay for
  the process's life (a short string each) and `added`'s clearing is
  reached only with a caller-chosen id (`gf:…/registry.go:738-741`).
- **M:** without `delete(h.gone, local)` the adapter's tests pass.
- `hooks.go:21-22`: "node_removed on Decommissioned" — now also on
  `Deleted`.
- **Fix:** the godoc line; a unit test of removed → added → removed on
  one local id; the growth can stay. Contract: no.

### UR-P4-test-394 — note — the status guard's new branch has no test

- **Claim (M):** with `statusguard.go:127` replaced by a no-op,
  `./fabricui` and `./internal/...` pass. `changelog.md:108-109` names
  the behaviour. `TestEveryHandWrittenRouteIsInTheContract` holds the
  route table itself, so this is the second line of test-266 only.
- **Fix:** one case in `internal/statusguard`'s test. Contract: no.

### UR-P4-doc-395 — note — three changelog phrases

- `:70-72` "a refused event or log stream is now shown; the status is
  reloaded and the stream reopens when admin rights return": the log
  stream is shown only (`Operation.svelte:98`, `context.ts:165-169` — no
  status reload, no reopen); `ui/README.md` says it right.
- `:18` "0.11.0's line above": the 0.11.0 section is below.
- `:84` "in every theme": the four built-in palettes
  (`tokens.test.ts`); a host's live theme is the `-color` override's
  case (`tokens.css:172-174`). `ui/README.md` says "every palette".

### UR-P4-ui-396 — note — the notice and a failed rebuild's events (I)

- `store.svelte.ts:487` clears `exposureNotInEffect` on every
  `endpoints_changed` and `surface_changed`. `endpoints_changed` is sent
  after a rebuild that succeeded only (`handlers_exposure.go:227-229`,
  `handlers_node.go:372`); `surface_changed` is go-fabric's observer per
  endpoint an assembly changed (`docs/contract.md:260-267`), which a
  rebuild that fails part-way may have raised before it failed. Such an
  event read after the `PUT`'s answer (`reassembled: false`, `:304`)
  clears the notice of the very apply that failed. Not measured: whether
  a failing assembly raises the observer at v0.23.0 was not read. The
  fix wave's own comment names the opposite gap (a rebuild that changes
  nothing clears nothing).
- **Fix, if wanted:** clear on `endpoints_changed` only. Contract: no.

### Not covered

- The web tests added by the wave, line by line (judged by the first
  reader's mutation batch and the green suite): `App.test.ts`,
  `Node.test.ts`, `client-rules.test.ts`, `store-status-readers.test.ts`,
  `Diagnostics`, `Endpoints`, `Events`, `Fabrics`, `Selection`,
  `Topology`, `Pair`, `NodeManage`, `dialogs`, `Operation` tests,
  `contract-guard.test.mjs`.
- Go tests `log_once_test.go` (test-265) and `routes_contract_test.go`
  beyond its one function; `internal/api/api.gen.go` and
  `schema.gen.ts` (generated, `generate-check` green).
- No web mutation of my own; no browser run; the 4.5:1 ratios were not
  recomputed by hand (the guard test computes them, its model was read).
- Whether a failing assembly raises `OnSurfaceChanged` (UR-P4-ui-396),
  and `gf:controller/commissioner` (open since run 2).
- `notes/` of the wave (the register is still to be written).

### Tokens

About 190 k (estimate; the harness reports the exact number to the coordinator).

## Second look: go-fabric v0.24.1 and #81

Reader (Fable), 2026-11-22. `git diff 5dcf396~6..5dcf396` (f1d2321,
787352c, 949261f, 3424872, 34a897a, 5dcf396), whole, with the generated
files only as far as the description lines. go-fabric `v0.24.1` as a
`git archive` at `<scratchpad>/gf2` (`gf:`), its working tree not read;
`v0.23.0` of two files the same way for the "since v0.24.0" statements.
Tests and mutations in `<scratch V>` (detached at `5dcf396`, every file
restored, `git status` clean); nothing run in the review worktree.
M = measured here, I = read or inferred.

**Result: 0 high, 0 medium; 2 low, 6 notes. No held candidate.** Every
`gf:` citation and every go-fabric statement of the six commits holds at
v0.24.1.

### The focus points, each read

- **Attestation on the discovered path: fail-closed.** The adapter hands
  the template's policy on unchanged (`gofabric.go:594`); go-fabric
  applies `attestation.Reject()` to the zero value before `Edit`
  (`gf:controller/controller.go:1385-1397`), on the known-device branch
  of the one pipeline (`:1337-1347`). No request field reaches the
  policy. M: the test's `unset` case is red when the call goes back to
  `ForDevice` + `Commission` without the restatement (mutant M1), so
  UR-P1-adpt-60's class is held by a test again, now at go-fabric's side.
- **BLE and Address cleared: right.** `PairDiscovered` refuses both with
  `ErrConfiguration` (`gf:controller/controller.go:1265-1269`), so a
  template with a central would fail every discovered commissioning
  without the clearing. Not held by a test (UR-P4-test-401).
- **The 600 s bound starts after the search.** The handler sets no ctx
  deadline for a code without `timeout_seconds`
  (`handlers_controller.go:1472-1495`), so the search keeps
  `DefaultPairTimeout`, one minute (`gf:controller/controller.go:1232-1236`,
  `:100`); the bound is taken per attempt when a candidate is reached
  (`gf:controller/pair_race.go:93-99`), and on the known-device branch
  after the skipped Discovery (`gf:controller/controller.go:1337-1339`).
  A commissioning from a code therefore ends at most 60 s + 600 s after
  the request (plus the fail-safe's disarm). With `timeout_seconds` the
  ctx deadline bounds search and commissioning on both paths and
  `CommissionTimeout` stays 0 (M: `TestCommission_DefaultBounds`, mutants
  M4a, M4b killed). A bound that ends the commissioning is
  `context.DeadlineExceeded` (`gf:controller/controller.go:1359-1370`):
  502 and the progress detail `timeout`, as a `timeout_seconds` deadline
  (`handlers_controller.go:240`, `:1551`).
- **The update check's mapping.** One `QueryUpdates`, one
  `UpdateCheckFinished` (`gf:controller/softwareupdate/manager.go:1846-1861`);
  every target of a failed product is in its `Peers`
  (`check_error.go:63-69`, `manager.go:1907-1911`), the groups are
  disjoint from the updates found, so no node is dropped or listed twice;
  an error without a `ProductCheckError` is returned as it is and the
  handler answers 500 (`handlers_controller.go:1776-1780`). M: mutants
  M5b–M5f killed.
- **adpt-197's code is gone whole:** `productNode`, `nodes`, the `only`
  branch of `query`, `node.softwareVersion`, `export_test.go`,
  `updates_key_e2e_test.go`; `git grep` finds no reader and no statement
  of the re-check left outside the 0.11.0 history lines.
- **The consent 204 statement is true** (`gf:…/manager.go:2362-2391`: not
  stored, `false`, no error), with the nuance of UR-P4-doc-407.
- **Contract: description-only.** Two description lines
  (`openapi.yaml:2007`, `:3588`) and their generated twins; `info.version`
  stays 0.12.0 (unreleased); no shape, status or enum moves. The
  discovered path's new "Discovery 2.1 skipped" report uses the declared
  phase (`openapi.yaml:3745`), and `docs/embedding.md:370` says so. No
  accepted break, no further version step.
- **Deprecations.** None of v0.24.0's deprecated identifiers
  (`gf:CHANGELOG.md`, "Deprecated" of 0.24.0) is used in a `.go` file
  here (`git grep`, I; `make lint` is the guard run's).

### Citations and reference statements

All six ranges name their construct at v0.24.1:
`controller/controller.go:1135-1149` (`CommissionTimeout`), `:1243-1285`
(`PairDiscovered`), `controller/softwareupdate/check_error.go:8-59`,
`manager.go:1829-1843` (`QueryUpdates`), `manager.go:2345-2354`
(`AddUpdateConsent`'s rules), `controller/pair_commission_timeout_test.go:19-71`
(the 13.5 s is that file's `:19-21`).

Statements read against the tag, all true: `PairDiscovered` is `Pair`'s
pipeline with the trace, `pair.failures` and Discovery skipped
(`controller.go:1243-1259`, `:1304-1314`); `CommissionTimeout` is
independent of the search's bound (`:1135-1149`); `QueryUpdates` joins
one `*ProductCheckError` per failed product naming its nodes
(`manager.go:1829-1840`); no consent to a version at or below the running
one since v0.24.0 (`gf:CHANGELOG.md` 0.24.0 "Changed"; absent at
v0.23.0, M); v0.24.1 is the last version under the path and `go.mod`
carries the `// Deprecated:` line `go get` and `go list -m -u` report;
store schema `v12` is a v0.24.0 break. No statement about
`Node.PeerAddress` on a failed load is in the diff; the behaviour is
UR-P4-adpt-405.

### Mutations (this read)

`<scratchpad>/mut/mut.py`, one change each, the adapter's and the
handler's named tests, the file restored. 13 mutants, **10 killed**: M1
`PairDiscovered` back to the registry path, M3a/M3b the bound not handed
on (`Pair`, `CommissionDiscovered`), M4a/M4b the handler's bound never or
always, M5b an error without a product as a partial result, M5c no
cancellation check, M5d/M5e the first peer or product only, M5f the
updates found dropped. **3 survived:** M2 (UR-P4-test-401), M5a
(UR-P4-test-402), M4c — the handler's bound on the discovered path too,
equivalent beside that path's 600 s ctx.

### UR-P4-bound-400 — low — the handler's 600 s replaces a host's own bound on the code path

- **Both sides.** `gofabric.go:342-344`: a positive port value overwrites
  the template's `CommissionTimeout`; the handler sends 600 s for every
  code without `timeout_seconds` (`handlers_controller.go:1493-1495`).
  On the discovered path the handler sends none, so the template's bound
  applies there.
- **Claim.** A host that sets `ControllerOptions.Pair.CommissionTimeout`
  (new in go-fabric v0.24.0, part of the template the godoc calls "of
  every commissioning") below 600 s gets it on the discovered path and
  with `timeout_seconds`, and 600 s on the code path without. Nothing is
  unbounded; the godoc (`controller.go:711-717`: "zero leaves the
  source's own default") and the changelog do not say that a positive
  value wins over a shorter one.
- **Fix:** the smaller of the two in the adapter, or one sentence in the
  godoc and `docs/embedding.md`'s `Pair` row. Contract: no.

### UR-P4-test-401 — low — no test holds the cleared BLE and Address

- M: mutant M2 (`opts.BLE`, `opts.Address` left as the template's)
  passes the adapter's tests: no site's template names a central or an
  address on the discovered path.
- **Claim.** Without the clearing every discovered commissioning of a
  host with a BLE central is 400 (`gf:controller/controller.go:1265-1269`).
  The line is right today and unguarded.
- **Fix:** the discovered commissioning once with a template that names a
  central (the BLE tests' fake). Contract: no.

### UR-P4-test-402 — note — "its product's cause" is not held

- M: mutant M5a (`Err: err`, the whole joined error, for every node)
  passes: the test asks `errors.Is(failure.Err, cause)`, which the joined
  error answers for both causes. The route's `errors[].error` would then
  carry every product's text for each node. Assert the text, or that the
  other product's cause is absent.

### UR-P4-test-403 — note — the real manager's error reaches the route in no test

- The mapping is tested on the `query` seam with a hand-built
  `ProductCheckError`; `TestEndToEnd_UpdateEventsFromTheManager` runs a
  check that finds nothing. No test lets go-fabric's manager fail one
  product (an image source that refuses the download) and reads the
  route's `errors` and the single `check_finished`. The type is
  go-fabric's own, so the seam is sound; the end-to-end half is open.

### UR-P4-err-404 — note — `PairDiscovered`'s refusals are 400 on the discovered path

- `gf:controller/controller.go:1270-1279`, `:86`: a discovered device
  without an address, an address without IP or port and a closed
  controller are `ErrConfiguration`, which `commissionError` maps to
  `ErrInvalidArgument` (`gofabric.go:426-427`): 400 "Invalid request"
  and the progress detail `invalid_argument` for a request that was
  valid. Before, the registry path answered these as the node's failure
  (the closed controller explicitly, 502). 400 is declared; whether
  `Discover` can return a device without an address was not read (I).
  The passcode case cannot arrive: the handler refuses it first.

### UR-P4-adpt-405 — note — `Hooks.nodeID` reads `PeerAddress`'s `ok` as "commissioned"

- `hooks.go:153-173` (godoc: "before the node is commissioned and
  known") against `gf:controller/client/record.go:714-737`: since
  v0.24.0 a node that failed to load answers its stored address, and
  go-fabric's changelog sends such a caller to `IsCommissioned`. The
  routes are unaffected (`lookup` asks `IsCommissioned`, `Nodes` ranges
  `Commissioned()`); a hook of a failed node (its `Deleted` on `Forget`)
  may now publish an event under an id no list ever showed. Not run (I).
  Fix, if wanted: `IsCommissioned` beside the address, or the godoc.

### UR-P4-err-406 — note — `ErrNoNode` on the consent route is 500

- `gf:controller/softwareupdate/manager.go:809-824`, new in v0.24.0:
  `AddUpdateConsent` answers `ErrNoNode` for an address no node has
  (before: stored, `false`). `updates.go:133-145` finds the node first,
  so only a removal between the two reaches it, and it is then 500, not
  the route's 404. The godoc added by 34a897a names the two `false`
  cases and not this one.

### UR-P4-doc-407 — note — two nuances of the new texts

- The consent is "not kept" only while the manager knows the version the
  node runs (`gf:…/manager.go:2362-2391`, `known`; `:779-800`): for a
  node whose Basic Information is not cached the consent is stored. The
  204's description and `updates.go:125-129` state it without the
  condition.
- `updates.go:76-80`: were every `ProductCheckError` without `Peers`,
  `Check` would return an empty `UpdateCheckFailures` as a non-nil error
  and the route 200 without `errors`. go-fabric builds a group only with
  a target (`gf:…/manager.go:1877-1891`), so it cannot arrive at
  v0.24.1; `len(failed) == 0 → err` would make it independent of that.

### Not covered

- `fabricui/dist` and the generated files beyond the two description
  lines; `make lint`, `make ci`, `contract-check` (the guard run's).
- go-fabric's race internals (`client/discovery` session and
  `CommissioningRace`) beyond `Timeout`'s meaning; BLE commissioning
  under the bound; the fail-safe's disarm after the bound.
- No run with a failed-load node (adpt-405), none with a device without
  an address (err-404), none of the real manager failing a product
  (test-403).
- v0.24.0's other behaviour changes of `controller/softwareupdate` (the
  queue's phases, the 24 h consent, `UpdateConsentChanged` on a node
  found gone) against the Updates page: only the queue's three status
  words and the eight event kinds were compared (unchanged).
- The web half: no page changed in the six commits.

### Tokens

About 150 k (estimate; the harness reports the exact number to the coordinator).
