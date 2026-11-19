**HELD: carries UR-P4-doc-190 (medium, security, same class as UR-P4-auth-03). Stays in `held/` until its fix merges (decisions 5, 14).**

# Phase 4 — pre-merge read of the fix wave

Reader (Fable), `git diff 979a67a..92bbf10` without `fabricui/dist`, the
two generated files; go-fabric `v0.23.0` at `<scratchpad>/pins/gf`
(`gf:`). Tests of mine in the scratch worktrees, moved to `92bbf10`
(detached): `<scratchpad>/p4wt/S/fabricui/p4pm_test.go`,
`…/S/fabricui/gofabric/p4pm_test.go`, `…/p4pm_window_e2e_test.go`,
`<scratchpad>/p4wt/W/ui/src/lib/p4pm.test.ts`; the run-1 red tests of S
and W moved to `<scratchpad>/p4wt/old-{s,w}-tests`. No make target run,
the review worktree untouched. **M** measured, **I** inferred.

Counts: 0 high, 1 medium (held), 4 low, 2 note. **Not ready to merge
as it stands: UR-P4-doc-190 is a sentence of the new public docs.**

| ID | Severity | Where | One line |
| --- | --- | --- | --- |
| UR-P4-doc-190 | medium, held | `docs/embedding.md` "Host names", changelog, `docs/contract.md` | "a proxy that forwards `Host` keeps the check … or set `AllowHost`": with `AllowHost` by source address every name passes, writes included |
| UR-P4-doc-191 | low | embedding, contract, changelog | "a node without a fabric has its own window open again after a close" is true for a fresh window only |
| UR-P4-adpt-192 | low | `gofabric/updates.go` `Check` | one product without a servable image → every check asks every other node again, one by one |
| UR-P4-ctr-193 | low | `handlers_pairing.go:103` | `window_match` false under `administrator_basic`, whose codes are the configured ones |
| UR-P4-web-195 | low | `store.svelte.ts` `reopenOnNewToken` | a non-default node's store opens its stream twice on every mount |
| UR-P4-test-194 | note | `sibling_test.go` | the reflection guard does not see `token` and eight other fields |
| UR-P4-ctr-196 | note | contract "Versioning" | a new response enum value is not among what a minor may add |

## Findings

### UR-P4-doc-190 — medium — held — forwarding plus `AllowHost` does not "keep the check"

- **Both sides.** `docs/embedding.md` (new): "A proxy that forwards the
  browser's `Host` unchanged keeps the check: list the names the browser
  uses in `AllowedHosts`, or set `AllowHost` and answer it from what
  only the proxy produces, such as its source address". Changelog
  0.11.0 "Hosts", first item: "…or it forwards the browser's `Host` and
  the host lists the names or sets `AllowHost`". `docs/contract.md`
  "Host" keeps "behind a proxy that forwards … it sets `AllowHost`".
  Code: `host.go:69-70` — the predicate admits whatever the name is;
  `handler.go:764-765` — for a request the predicate admitted, an
  `Origin` equal to `Host` passes.
- **Claim.** Only the *list* keeps the check behind a forwarding proxy.
  `AllowHost` answered from the proxy's source address is as constant as
  a replaced `Host` — and wider than UR-P4-auth-03: the rebound page's
  own `Origin` equals its `Host`, so its **writes** pass too. Decision
  61 accepted that for the Ingress (the proxy authenticates the user);
  the new paragraph, written to explain rebinding, now says the
  opposite of it. A host behind a forwarding proxy without
  authentication reads "keeps the check" and is unprotected.
- **Test (M):** `TestP4PM_AllowHostBySourceAdmitsAReboundName` (token
  mode, predicate true by `RemoteAddr`) — red:
  `status under a rebound name: 200, token issued: true`;
  `an admin write under a rebound name with its own Origin: 204`.
- **Severity.** Same preconditions as auth-03 (catch-all proxy, plain
  HTTP, no authentication at the proxy), effect larger (writes). Not a
  code defect: decision 61's shape. The defect is the sentence.
- **Fix (docs only, this pull request).** Three kinds of proxy, not
  two: forwards `Host` and the host *lists* the names — the handler
  checks; forwards `Host` and the host sets `AllowHost` from the proxy's
  source address — every name the proxy forwards passes, reads and
  writes, the check is the proxy's (its authentication, its server
  names); replaces `Host` — as written. Same wording in the changelog
  item and one clause in `docs/contract.md` "Host" and the godoc of
  `AllowHost`. `host_proxy_test.go` gains the fourth case and holds the
  new sentence. No contract artefact, within decision 69.

### UR-P4-doc-191 — low — the reopened own window is stated for every close

- **Both sides.** `docs/embedding.md:54`: "**Either way**, a node
  without a fabric has its own window … open again once the opened
  window closes or expires". `contract/openapi.yaml` `closeWindow`: "On
  a node without a fabric go-fabric then reopens the node's own
  window". `docs/contract.md` route row and "From 0.11.0 …": "After
  `POST /node/pairing/window/close` a node without a fabric has its own
  window open again". Changelog: "after a window the page opened closes
  or expires". Other side: `gf:node/commissioning.go:155-160`
  (`transition` "reports whether a [Node.OpenBasicWindow] window just
  closed"), `:244` (`return endedBasic && !window.IsOpen()`), `:728-731`
  ("that closes a **basic** window reopens its own window"); and this
  repository's own `fabricui/e2e_test.go:332-337` (M, existing): the
  close of the own window on the uncommissioned node leaves
  `WindowOpen` false, kind `none`.
- **Claim.** Only a window on fresh credentials (`OpenBasicWindow`)
  comes back as the own window. With `ConfiguredWindow` the page opens
  the own window; its close or expiry leaves the node without a window,
  the opposite of "either way". The godoc of `ConfiguredWindow`
  ("that window" = the fresh one) and the second-close sentences are
  right; the page reads the status and is right.
- **Fix:** "the close or expiry of a window on fresh credentials" in
  the four places; drop "Either way" and say that with
  `ConfiguredWindow` the closed window stays closed. Contract: text only.

### UR-P4-adpt-192 — low — the node-by-node re-check runs on every check while one product fails

- **Where:** `fabricui/gofabric/updates.go` `Check` (the loop over
  `u.nodes()`), `gf:controller/softwareupdate/manager.go:991-1029`
  (a query of one node runs that node's product check: ledger lookup,
  then `DownloadUpdate`), `:964-971`, `images.go:992-996` (an update
  whose ledger record states no checksum is refused with `ErrOTA` —
  deterministic, every time).
- **Claim.** The fallback asks *every* node without an update found,
  not the failed product's: with one product in that state each
  `POST /controller/updates/check` costs 1 + (nodes without an update)
  product checks in sequence inside the request, each with its own
  ledger lookup, a repeated download attempt per node of the failing
  product, one `check_finished` event per node, and per open admin page
  one `GET /controller/updates` per event
  (`controller-store.svelte.ts:211-216`). I: whether go-fabric caches a
  ledger answer was not read. The state is permanent until the vendor's
  record changes, so this is the steady state, not the exception. A
  client that leaves mid-way is answered 500 and logged (I, by reading
  `CheckUpdates`: a context error is not a `UpdateCheckFailures`).
- **Test (M):** `TestP4PM_CheckAsksEveryNodeAgain` — red: `one failing
  product, 30 nodes: 30 further checks of one node each`.
- **Events, concurrency (asked for):** each one-node query emits its own
  `check_finished` with the node as peer (`manager.go:955-962`), as the
  docs say; found nodes are skipped, so no `available` event repeats;
  `queueUpdate` (`:1123-1144`) deduplicates a queued target. Two checks
  at once double the cost, nothing else. No defect there.
- **Fix.** The gap is go-fabric's: `QueryUpdates`' joined error names
  no product. README "Method" step 4 — a facade gap is an issue, not a
  workaround: file it (an error per product, `errors.As`-able with
  vendor and product id). Until it is pinned, ask one node per product
  (vendor, product, software version — the manager's own group key) and
  give its result to the product's other nodes.
- **Contract:** no.

### UR-P4-ctr-193 — low — `window_match` is false under an administrator's basic window

- **Both sides.** `fabricui/handlers_pairing.go:100-103`: "The
  configured codes commission through the node's own window only
  (decision 28)", `WindowMatch = ptr(win.Kind == node.WindowOwn)`.
  `gf:node/commissioning.go:215-230`: an administrator's basic window
  is "PASE with the configured passcode and the commissionable record
  under the configured discriminator"; `gf:node/observers.go:246-251`.
  And the wave's own contract text: `administrator_basic` "on the
  node's own passcode".
- **Claim.** Since v0.23.0 tells this window from an enhanced one, the
  answer "a window is open on other credentials" is wrong for it: the
  Pair page marks the configured codes as not valid for the open window
  while they are what a commissioner enters. Present at `979a67a`
  already; the wave names the kind and its passcode without moving this
  line (decision 30's carried row).
- **Test (M):** `TestP4PM_WindowMatchUnderAnAdministratorsBasicWindow`
  (the wave's e2e plus `GET /node/pairing`) — red: `source "configured"
  open true window_match false`.
- **Fix:** `win.Kind == node.WindowOwn || win.Kind == node.WindowAdministratorBasic`,
  the comment, the test. **Contract:** no (the field's sentence holds).

### UR-P4-web-195 — low — a non-default node's stream is opened twice

- **Where:** `ui/src/lib/store.svelte.ts` `reopenOnNewToken`,
  `connect` (`streamToken = status?.csrf_token`);
  `ui/src/shell/NodeScope.svelte:32-33` (`loadStatus()` unawaited, then
  `connect()`); `api/client.ts:363-364` (one token for every node's
  client).
- **Claim.** The store remembers the token of *its own status* at the
  open, not the token the stream was opened with. A scoped store opens
  its stream before its first status, with the shared token the client
  already holds; the status then brings the same token, `streamToken`
  is still undefined, and the stream is closed and opened again — on
  every switch to a non-default node, each open counting against
  `MaxStreams` until the abort lands.
- **Test (M):** `p4pm.test.ts` — red: two `/nodes/second/stream t1`.
- **Fix:** compare with the token the client sent at the open (a getter
  on the client, or `openStream` reporting it), not with
  `status?.csrf_token`. The three cases of `token-path.test.ts` stay.
- No loop: after the reopen `streamToken` equals the status's token
  (read), and host mode never reopens (`!s.csrf_token`).

### UR-P4-test-194 — note — what the sibling guard cannot see

`TestSibling_CopiesEverySharedField` skips a field that is zero on the
parent. In its set-up that is (M, `TestP4PM_SiblingGuardBlindFields`):
`token` (the set-up is host mode), `warnings`, `counters`, `configured`,
`allowEmpty`, `exposureReason`, `registry`, `nodesHub`, `hostDone` —
all copied by `sibling` today, none guarded; a future shared field is
guarded only if this one test happens to set it. Fix: fail on a zero
parent field that an explicit list does not excuse, and a second run in
token mode.

### UR-P4-ctr-196 — note — the versioning rule and a new enum value

`contract/openapi.yaml` "Versioning": a minor "adds paths, methods,
response fields and optional request fields". `a44044a` adds a value to
the response enum `Status.window_kind`; a client generated with closed
enums fails the whole status on it, and the guard does not report enum
widening. In practice 0.11.0 worsens nothing: v0.10.0 on go-fabric
v0.23.0 already sent `administrator-basic` (`handlers_node.go:218` at
`979a67a`, `gf:node/observers.go:266`), a value outside its own enum;
the shipped 0.10.0 page compares kinds as strings
(`store.svelte.ts:345`) and reads an unknown one as another window. Fix:
one sentence — a response enum may gain a value, a client tolerates an
unknown one — in the contract's Versioning paragraph and
`docs/contract.md`; and say in the 0.11.0 lines that 0.10.0 sent the
hyphenated value. Also: a future go-fabric kind reaches the wire as
`WindowKind(5)`; map an unnamed kind to a fixed value.
`format: int64` is no wire change; a Go client generated from the
contract changes those fields' type (`int` → `int64`) — said in the
changelog for the server side only.

## The five focus points

1. **Admin, token, Host paths.** auth-01: `sibling` carries both fields;
   the three behavioural tests and the registry-node test hold (re-run,
   M); guard see -194. web-191: correct for the default store, no loop,
   host mode untouched; -195 for a scoped store. web-04: correct —
   resolves against `document.baseURI` only when a `<base href>`
   exists, compares with the page's origin; an opaque `location`
   still fails closed. auth-05: the godoc line is right (RFC 6761
   §6.3).
2. **auth-03's public docs.** The class "a proxy that replaces `Host`"
   is stated correctly in `embedding.md`, the godoc of `AllowedHosts`,
   `contract.md` and the changelog, and `host_proxy_test.go` measures
   the three cases and holds two sentences (re-run, M). Nothing beyond
   decision 69: no `Options` field, no contract artefact, the commit
   message neutral, committed notes carry the ID only. Wrong is the
   neighbouring sentence on the forwarding proxy: -190.
3. **go-fabric statements against v0.23.0.** Right: the proxy class
   (ours); `Commissioning.WindowOnDemand` (`gf:node/config.go:509-521`);
   `administrator-basic` on the node's own passcode and discriminator
   (`observers.go:246-251`, `commissioning.go:215-230`);
   `client.ErrRequestTooLarge` / `ErrMalformedRequest` as refusals
   before anything is sent (`gf:controller/client/interact.go:15-28` —
   the docs' "a timed write" is one of its three cases, the others a
   response-suppressed write and a value that fits no message);
   `node.ErrFabricNotFound` (`gf:node/facade.go:103-105`,
   `node.go:1719`); NOCResponse InvalidFabricIndex = 11
   (`gf:cluster/core/operational_credentials.go:430`); `Pair`'s
   one-minute search bound only without a deadline
   (`controller.go:1227-1234`); no exported `PairDiscovered` at the pin,
   and no new text names one. Too wide: the reopened own window, -191.
   Not checked (no v0.22.2 export read): "go-fabric 0.22 reported it as
   `enhanced`".
4. **adpt-32.** Result and mapping right: updates kept, a failed node
   named, `errors` optional, a whole failure 500 (declared), the status
   guard now runs over the adapter's tests. Cost: -192.
5. **Contract 0.11.0.** `contract/previous/` is byte for byte `v0.10.0`'s
   contract (M, `diff`); `x-accepted-breaks` is gone as it must; every
   change is an added optional field, an added declared status use, a
   description or a `format`. The enum value: -196.

## Disposition

1. UR-P4-doc-190 — **before the merge**, docs only (G), then one more
   look at the paragraph by this reader; held until merged.
2. UR-P4-doc-191 — before the merge, the same docs commit (G).
3. UR-P4-ctr-193 — before the merge, one line and its test (G).
4. UR-P4-web-195 — before the merge (W), small.
5. UR-P4-adpt-192 — go-fabric issue now; the per-product grouping in
   this pull request if cheap, else a register row carried to run 2.
6. UR-P4-test-194, UR-P4-ctr-196 — notes; -196's sentence fits the docs
   commit.
7. Run 2 is due in any case (run 1's mediums); -190 adds a medium to
   the same held class, not a new class.

## Second look at f787172

Bounded: `git diff 92bbf10..f787172` over `docs`, `changelog.md`,
`fabricui/options.go`, `host.go`, `host_proxy_test.go`,
`contract/openapi.yaml`, plus `6b1f428` and `6b81d06`. Re-run in
`<scratchpad>/p4wt/S` at `f787172` (M): `go test ./fabricui -run
TestHostCheck_` ok; `go test ./fabricui/gofabric -run
'TestUpdates_|WindowReadsAdministratorBasic'` ok. My earlier `p4pm`
tests moved to `<scratchpad>/p4wt/old-pm-tests`.

| Item | Verdict |
| --- | --- |
| (a) UR-P4-doc-190 and UR-P4-auth-03 in the docs | **OK** |
| (b) UR-P4-doc-191 | **OK** |
| (c) UR-P4-ctr-193 (`6b81d06`) | **OK** |
| (c) UR-P4-adpt-192 (`6b1f428`) | **UR-P4-adpt-197, low** — the key is not the manager's |

- **(a) OK.** `docs/embedding.md` "Host names" now names three kinds of
  proxy and says for each whose check it is: listed names — the
  handler's, "the one setup"; `AllowHost` — every forwarded name passes,
  reads and writes, because the page's `Origin` equals its `Host`
  (`host.go:69-70`, `originAllowed`); replaced `Host` — constant, reads.
  The closing sentence gives the two remedies (the proxy authenticates,
  or serves only its own names). The `AllowHost` option row and godoc,
  `docs/contract.md` "Host" and the changelog item (its headline
  included) say the same. `TestHostCheck_AllowHostBySourceAdmitsEveryForwardedName`
  measures the second kind and the docs guard holds four sentences.
  Nothing beyond decisions 61 and 69: no new option, no contract
  artefact, the Ingress named only as the case whose proxy
  authenticates; no ID or held detail beyond what the paragraph itself
  must say.
- **(b) OK.** Every statement is now tied to "a window on fresh
  credentials (`basic`)": `closeWindow`'s description and 204 text, the
  route row and both paragraphs of `docs/contract.md`, the
  `ConfiguredWindow` row and godoc, the changelog. "The close of the
  node's own window leaves no window open" matches
  `gf:node/commissioning.go:155-160`, `:244` and `e2e_test.go:332-337`.
- **(c) ctr-193 OK.** `handlers_pairing.go:103-105`: own or
  administrator-basic; the e2e holds it.

### UR-P4-adpt-197 — low — the adapter groups by the version *string*, go-fabric by the version *number*

- **Both sides.** `fabricui/gofabric/updates.go` `newUpdates`:
  `fmt.Sprintf("%04X/%04X/%s", b.VendorID, b.ProductID, b.SoftwareVersion)`,
  where `Basic().SoftwareVersion` is a string filled from
  `AttrSoftwareVersionString` (`fabricui/controller.go:371`,
  `gofabric/node.go:132`). `gf:controller/softwareupdate/manager.go:991`:
  the manager's group key is vendor, product and `d.softwareVersion`,
  the uint32 (`strconv.FormatUint(uint64(d.softwareVersion), 10)`), and
  the applicability of an update is decided on that number
  (`images.go:861`).
- **Claim.** The comment, `docs/contract.md` and the commit say "as the
  manager groups a check"; it is not the same key. Two nodes of one
  product whose numbers differ and whose strings are equal (the string
  is the vendor's free text) fall into one group: when the first has no
  update, the second is not asked — its failed check is then in neither
  `updates` nor `errors`, and if that was the only failure the route
  answers 200 without `errors`; when the first fails, the second is
  named failed without having been checked. Rare precondition, silent
  when met. I: no test — the seam takes the key as given; the defect is
  the line that builds it.
- **Fix:** key on the numeric SoftwareVersion (read
  `AttrSoftwareVersion` from the same cache in `nodes()`), a test with
  two nodes of equal string and different number. Small; this pull
  request or a register row for run 2. Contract: no.

Disposition after the second look: UR-P4-doc-190 and -191 are closed by
`54a41e0` and `cd7139b`; the held class may merge. UR-P4-adpt-197 does
not block the merge.
