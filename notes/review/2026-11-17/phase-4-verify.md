**HELD: carries the held candidate UR-P4-auth-03 (medium, security). Stays in `held/` until its fix merges (decisions 5, 14).**

# Phase 4, run 1 — verification

Verifier (Fable), independent of S, G, W. Code read at `979a67a` in the
scratch worktrees `<scratchpad>/p4wt/{S,G,W}` (detached), go-fabric
`v0.23.0` at `<scratchpad>/pins/gf` (`gf:`). Every red test re-run by me
on 2026-11-19; the review worktree untouched. **M** measured, **I**
inferred. One test added by me in the S scratch worktree
(`fabricui/p4v_fixed_name_proxy_test.go`), left in place.

Scope: the three mediums; five lows (S 2 of 2, G 1 of 1, W 2 of 4 —
decision 17's sample); notes not read.

| ID | Reviewer | Verdict | My severity |
| --- | --- | --- | --- |
| UR-P4-auth-03 | S, medium, held | CONFIRMED, claim widened, fix shape 2 does not close it | medium |
| UR-P4-adpt-31 | G, medium | CONFIRMED | medium (lower edge) |
| UR-P4-adpt-32 | G, medium | CONFIRMED, trigger corrected | medium |
| UR-P4-auth-01 | S, low | CONFIRMED | low |
| UR-P4-bound-02 | S, low | CONFIRMED, proposed fix has a side effect | low |
| UR-P4-ctr-30 | G, low | CONFIRMED | low (upper edge) |
| UR-P4-ui-60 | W, low | CONFIRMED | low |
| UR-P4-test-62 | W, low | CONFIRMED, one count wrong | low |

Struck: none. False positives in the Opus sample: 0 of 3 (G 1, W 2);
two reports carry a wrong detail (adpt-32's trigger, test-62's count).

## Mediums

### UR-P4-auth-03 — CONFIRMED, medium, held — and wider than reported

- **Re-run (M):** `go test ./fabricui -run TestP4S_AProxyThatRewritesHostVoidsTheHostCheck -count=1`
  → FAIL: `status under an unlisted name through the proxy: 200, token
  issued: true`; `pairing under an unlisted name through the proxy: 200,
  passcode in the answer: true`.
- **The read is reachable as claimed (M).** `host.go:53-55` admits any
  IP literal, `:50` `localhost`; `handlers_node.go:147` puts the token
  into every admitted status (token mode); `handler.go:753-756` passes a
  request without `Origin`; `GET /node/pairing` then answers the passcode
  on the token alone. Writes are refused: `handler.go:761-765` admits no
  `Origin` of the rebound name (no rule a–c, not `byPredicate`). I could
  not refute any step on the handler's side. I: a same-origin `fetch`
  GET carries no `Origin`; nginx's `proxy_pass` default (`$proxy_host`)
  and Apache without `ProxyPreserveHost` send the upstream's address —
  not pinned, consistent with what I know of both products.
- **Wider reach of UR-P0-auth-20, not accepted before: yes.** Decision
  37's shape rests on `held/phase-0-verify-held.md:49-51` ("an IP
  literal, a loopback name or an allowed name cannot be made to appear
  under the attacker's name") and remark 3 `:55-56` ("a reverse proxy
  passes its own name in `Host`"). Decision 61 measured the one proxy of
  the pinned hosts (Ingress forwards `Host`) and added `AllowHost`.
  Nothing in decisions 37/61, `held/phase-3-contract-held.md` or
  `docs/embedding.md:455-462` weighs a proxy whose `Host` does not
  depend on the browser's. The docs name the case and tell the host to
  list the name, without saying what that does to the check.
- **Correction — the class is larger than rules (a)/(b) (M, my test).**
  `TestP4V_AProxyThatSendsItsOwnListedNameVoidsTheHostCheckToo`: a proxy
  that sends a fixed name of its own (`ui.lan`, listed in
  `AllowedHosts`) — the docs' own first case, `embedding.md:458-459`,
  and the godoc of `AllowedHosts` (`options.go:300-302`) — answers the
  rebound name 200 with the token as well; the control (the proxy
  forwards `Host`) is 403. So every proxy that *replaces* `Host`, by an
  address or by a name, makes the handler's check constant. The handler
  cannot see the browser's name there at all.
- **Severity: medium stands.** Preconditions: a proxy that replaces
  `Host` and answers a server name it does not know (catch-all vhost),
  plain HTTP, no authentication at the proxy; token mode, or host mode
  with `Admin` by source address. `embedding.md:463-467` describes host
  mode with a proxy that authenticates the user and token mode "for a
  daemon that serves the page directly" — the documented host-mode setup
  is not reachable (the rebound name has no session at the proxy), token
  mode behind a proxy is undocumented but not refused. Neither local
  host checkout (`go-fabric-server`, `go-ha2fabric`; checkouts, not
  pins) documents a reverse proxy or sets `AllowedHosts`. Effect: every
  admin GET incl. the pairing passcode and the GETs with an effect; no
  write. Not high: three deployment preconditions and no write.
- **Fix shapes and their cost.**
  1. *Docs and godoc* (S's shape 1) — `docs/embedding.md` "Host names",
     `docs/contract.md:435-443`, godoc of `AllowedHosts`/`AllowHost`,
     changelog. No contract artefact, no code. **It must be worded on
     the class**, not on the address case: "a proxy that replaces `Host`
     — with a name of its own or the upstream's address — makes this
     check constant; the check is then the proxy's (no catch-all server,
     refuse unknown names), or the proxy forwards `Host` and the host
     lists the name or sets `AllowHost`." The current sentence
     "A proxy that sends a name of its own as `Host`: list that name"
     has to change, not only gain a neighbour. Both red tests stay as
     demonstrations with the expectation inverted. Cost: small; closes
     nothing in code, which is correct here — the information is not in
     the request.
  2. *An `Options` field that drops rules (a)/(b)* (S's shape 2) —
     new public field, validation, tests, the fuzz invariants of
     `FuzzHostCheck`, docs. **It does not close the finding:** behind an
     address-rewriting proxy it turns every request into 403 until the
     operator either forwards `Host` (then shape 1's advice already
     suffices and the field adds a fail-closed default for that host) or
     sets a fixed name (then the check is void again, my test). Its only
     value is fail-closed discovery for a host that knows it is proxied.
     Not recommended as the fix; at most on top of 1, owner's call.
  Not proposed by S, noted only: reading `X-Forwarded-Host` as a second
  name to check would be a heuristic (nginx does not send it by
  default).
- **Contract:** no.

### UR-P4-adpt-31 — CONFIRMED, medium (lower edge)

- **Re-run (M):** `go test ./fabricui/gofabric -run TestP4G -count=1` →
  FAIL `Wrap drops the host's Deleted hook`.
- `gf:controller/client/config.go` `Hooks` has 14 func fields,
  `hooks.go:229-243` copies 13; `gf:controller/client/registry.go:466`
  calls `n.cfg.Hooks.Deleted`. `Hooks.For` sets no `Deleted` itself, so
  the only loss is the host's.
- Refutation tried on the precondition: the local `go-fabric-server`
  checkout does call `pageHooks.Wrap(cfg.HooksFor)`
  (`internal/daemon/run.go:285`), but pins go-fabric `v0.21.0` and its
  hooks (`controllerapi/api.go:214`) set no `Deleted`. So no host is
  hit today; the first host that adopts `Deleted` on its lift loses it
  silently. Real defect with a precondition → medium by the brief's
  scale, at its lower edge; a low would be defensible.
- Fix as proposed (`Deleted: both0(...)` plus the reflection test as the
  guard). Contract: no.

### UR-P4-adpt-32 — CONFIRMED, medium — the trigger is the image download, not the ledger lookup

- **Re-run (M):** `go test ./fabricui -run TestP4G -count=1` → FAIL
  `check with one failed product: 502 {… "code":"node_error"}`;
  `-run TestP4GUpdateCheckStatusGuard` → `PASS`, then `status guard: 1
  answered status(es) the contract does not declare … POST
  /controller/updates/check answered 502`, `FAIL`.
- **Against go-fabric v0.23.0 (M, read):**
  `gf:controller/softwareupdate/manager.go:1015-1029` — a product whose
  check errs is logged, joined into `failed`, and the loop goes on;
  `:1029` returns `out, true, failed`; `QueryUpdates` (`:947-949`)
  returns both. The adapter (`updates.go:46-49`) drops `found` on any
  error and maps it through `nodeError` (`gofabric.go:788`) to
  `ErrNodeFailed` → 502; `contract/openapi.yaml:1974-1981` declares 200
  403 500 503. The docs sentence G quotes is at `manager.go:943-946`.
- **Correction to the claim.** "One vendor's DCL or ledger lookup
  failing" is not the trigger: `Images.CheckForUpdate`
  (`gf:…/images.go:838`) returns `(UpdateInfo, bool)` — a lookup that
  fails is "no update", no error. The only error of
  `checkProductForUpdates` (`manager.go:1066-1078`) is
  `Images.DownloadUpdate`: an update *was found* and its image could not
  be fetched or stored (dead OTA URL, checksum, storage).
- **User harm (M for the page, I end to end):** `Updates.svelte:29-36`
  fills the consent buttons only from the check's answer. While product
  A's image cannot be downloaded, node B's page answers every check with
  a toast "the node failed: …" naming another product, and B's found
  update cannot be consented to from the page — for as long as A's
  download fails, which for a dead URL is permanent. The
  `UpdateAvailable` events go out, but the page does not read them into
  the list. The harm G inferred holds; its cause is narrower and more
  persistent than "DCL unreachable".
- Fix as proposed (return `found` with the error; a declared status or
  an additive field). Contract: yes.

## Sampled lows

### UR-P4-auth-01 — CONFIRMED, low
Re-run (M): both `TestP4S_Sibling…` FAIL (`"admin":false` on
`/nodes/second/status`; stream `[window_closed]`). `handler.go:596-617`
copies neither `allowedHosts` nor `allowHost`; `host.go:56` reads the nil
map. Fails closed, the gate itself runs on the default handler. Detail:
`originAllowed`'s predicate branch reads the request's admission, not
`h.allowHost`, so only `allowedHosts` is load-bearing; copying both is
still right.

### UR-P4-bound-02 — CONFIRMED, low — mind the fix
Re-run (M): FAIL, both `ctx has no deadline`. `handlers_controller.go:1420-1427`,
`gf:controller/commissioner/commissioner.go:326-331`,
`gf:controller/controller.go:1164-1170`, contract `:3526` as cited.
**Side effect of the proposed fix (M, read):** `Pair` applies
`DefaultPairTimeout` (1 min, `controller.go:98`) to the *search* only
when ctx has no deadline (`:1231-1234`). A default 600 s deadline from
the handler stretches "no device answers this code" from one minute to
ten. The fix must keep the search's bound (bound the commissioning, not
the search — in the adapter, or a go-fabric facade gap if `PairOptions`
cannot say it; it has no timeout field at v0.23.0).

### UR-P4-ctr-30 — CONFIRMED, low (upper edge)
Re-run (M): `GOARCH=386 go test ./fabricui -run 'TestP4G(Uint32|Consent)'`
→ FAIL: `"data_version":-559038737`, `"id":-918528`; `cannot unmarshal
number 3000000000 into … target_software_version of type int`.
`handlers_controller.go:546-557`, `:1689` (the existing 32-bit comment),
`ci.yml:127-128` (386 is built, not tested). Wrong numbers on a platform
the pipeline names; display and one rare write, hence low.

### UR-P4-ui-60 — CONFIRMED, low
Re-run (M): `npx vitest run src/review-p4/w-pages.test.ts -t UR-P4-ui-60`
→ FAIL, still `n":300` after 10 s. `Pair.svelte:57` derives from
`windowSecondsLeft(node.status, node.statusAt)` with no reactive clock;
`Status.svelte:35-40` ticks. The store reloads the status on
`window_closed` (`store.svelte.ts:453-455`), so the stale positive
"left" ends with that event when the stream is up; the frozen number in
between stands.

### UR-P4-test-62 — CONFIRMED, low — one count wrong
Re-run (M): `npx vitest run scripts/review-p4-W.test.mjs -t UR-P4-test-62`
→ 3 failed (`diffContracts` returns `[]`). `contract-guard.mjs:53-56`
compares parameter schemas only; no line reads `required` of a parameter
or of `requestBody`. The contract has 11 `in: query` parameters (W: six),
2 with `required: true`, and 23 of 26 request bodies required — the case
is live either way.

## Disposition

1. UR-P4-auth-03 — medium, held, **owner decision (wizard)**:
   recommended shape 1 (docs, godoc, changelog) worded on the whole
   class "a proxy that replaces `Host`"; the existing sentence for the
   fixed-name proxy is rewritten; both tests kept as inverted
   demonstrations. Shape 2 not as the fix. Stays in `held/` until merged.
2. UR-P4-adpt-31 — medium → fix wave G: the field plus the reflection
   guard.
3. UR-P4-adpt-32 — medium → fix wave G, contract 0.11.0; the register
   row names the download as the trigger.
4. UR-P4-auth-01 — low → fix wave G; a registry-node test beside the two.
5. UR-P4-bound-02 — low → fix wave G; the search keeps its one-minute
   bound, the contract sentence says the number.
6. UR-P4-ctr-30 — low → fix wave G; a 386 test run in CI.
7. UR-P4-ui-60, UR-P4-test-62 — low → fix wave W; the register corrects
   test-62's count.
8. Not sampled, verified by their regression tests (decision 17):
   UR-P4-ui-61, UR-P4-test-63 and every note.
9. **Run 2 follows** (README "Done"): run 1 has three confirmed mediums.

## What unit S's partial coverage leaves unread

Not reviewed by me, listed from S's own coverage section against the
brief's unit. Go: `handlers_controller.go` outside `:30-400` and
`:1394-1530` — about three quarters of the file, with the stored Wi-Fi
and Thread credentials' routes and what their reads return, ACL and
binding writes, decommission, the fabric label, update consent and the
controller's topology and diagnostics answers; `handlers_node.go:330-698`
(reassemble, factory reset, erase, the endpoint and event reads);
`operations.go` outside `:166-287` (the log buffer and what a log record
may carry to the log stream — whether a token, passcode or credential
can reach a line); `registry.go` outside `:177-252` (create, rename,
delete of managed nodes, key and label validation); `options.go` outside
the delta; `nodeswitch.go` and `provider.go` (the handler swap and the
`Mount`/`Handler` wiring the Host check hangs on); the adapter beyond
`gofabric.go`/`updates.go` — `gofabric/node.go` outside the delta,
`hooks.go`, `topology.go`, `diagnostics.go` (what of a network's
credentials or addresses reaches an answer or event); the generated
strict server's error paths in `internal/api`; the log calls on the
security paths as a class. Web: `client.ts:1-353`; `api/json.ts`;
`nodes-store.svelte.ts` (the third store, not in the brief's "two");
the two named stores, `route.ts`, `context.ts` and `App.svelte` outside
the delta and `:60-215`; `shell/main.ts` and `shell/index.html` (the
inline pre-paint script the policy hashes); `qr.ts` (`renderQrSvg`, the
producer behind both `{@html}` sinks — S checked the sinks, not the
function); `Pair.svelte`/`Window.svelte` beyond the sink lines (how long
the passcode and code stay in state). Aspects: host mode end to end with
a real `Admin`; the stream cap under admin and open streams together
(known row); `make lint` and the adapter's e2e, which S names as not
run.
