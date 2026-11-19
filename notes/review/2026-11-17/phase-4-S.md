**HELD: this report carries one held candidate (UR-P4-auth-03, medium, security). It stays in `held/` until its fix merges (decisions 5, 14).**

# Phase 4, run 1 — unit S (security paths)

HEAD read `979a67a`. go-fabric `v0.23.0` (`gf:`, the `git archive` export), delta `v0.22.2..v0.23.0` by `git diff`. Scratch worktree
`(scratch worktree)` (detached at `979a67a`, three untracked test files, left in place).
**M** = measured (line read, or test run), **I** = inferred.

Counts: 0 high, 1 medium (held candidate), 2 low, 2 note.

Red run, all three files together:
`cd <scratch> && PATH=$HOME/go/bin:$PATH go test ./fabricui -run 'TestP4S_' -count=1` → `FAIL` (4 tests, 4 failing; outputs below).

---

## UR-P4-auth-03 — medium — held candidate — a proxy that rewrites `Host` to the upstream's address voids the Host check; the docs name only the proxy that sends a name

- **Where:** `fabricui/host.go:46-57` (`nameAllowed`: `localhost`, `*.localhost`, any IP literal), `:61-73` (`admitHost`); docs `docs/embedding.md:455-462` ("A proxy that sends a name of its own as `Host`: list that name in `AllowedHosts`. A proxy that forwards the browser's `Host` unchanged: set `AllowHost`"), `docs/contract.md:435-443`, godoc `fabricui/options.go` `AllowedHosts` ("a host behind a proxy lists the name the proxy sends as Host").
- **Claim.** A wider reach of UR-P0-auth-20 (decisions 37, 61), not the decided rule itself. The docs know two proxies: one that sends a *name* of its own, one that forwards the browser's `Host`. The third is the common default: a proxy that sends the upstream's *address* as `Host` (`127.0.0.1:8099`, `localhost:8099`). Then every request reaches the handler under a loopback name or an IP literal, rules (a)/(b) admit it, and the Host check never sees the browser's name. Consequences, all measured in the test with `AllowedHosts: ["ui.lan"]` set as the docs say:
  - a page under any name that resolves to the proxy (the rebinding case of UR-P0-auth-20) gets `GET /node/status` 200 with `csrf_token`;
  - with the token every admin **GET** answers it — `GET /node/pairing` 200 with the passcode — because a same-origin GET carries no `Origin` (`handler.go:753-756` passes a request without one). Admin GETs with an effect (`GET /controller/discovery`, `GET /controller/topology?refresh=true`) go the same way;
  - writes stay refused: a POST/PUT/DELETE of that page carries `Origin: http://<its name>`, which no rule admits (`handler.go:761-765`) → `origin_not_allowed`. That refusal is also why a host behind such a proxy lists its real name in `AllowedHosts`, believes the setup is complete, and keeps the read hole.
  - Host mode: the same when `Admin` answers "came through the proxy" and the proxy does not authenticate per name.
- **Preconditions (why medium, not high):** the proxy answers a name it does not know (a default server), over plain HTTP (under TLS the rebound name fails the certificate), and does not itself authenticate; token mode, or host mode with `Admin` by the proxy's source address.
- **M/I:** M — the handler's behaviour (test below; `httputil.ReverseProxy` with `Rewrite: pr.SetURL(target)` sends the target's host as `Host`). I — that nginx (`proxy_set_header Host $proxy_host` is `proxy_pass`'s default) and Apache (`ProxyPass` without `ProxyPreserveHost On`) behave so; from the products' documentation, not a pin. I — the browser side (no `Origin` on a same-origin GET; rebinding itself), as held for UR-P0-auth-20 in `held/phase-2-X4.md`.
- **Test:** `<scratch>/fabricui/p4s_rewriting_proxy_test.go`, `TestP4S_AProxyThatRewritesHostVoidsTheHostCheck`. Red:
  ```
  p4s_rewriting_proxy_test.go:55: status under an unlisted name through the proxy: 200, token issued: true
  p4s_rewriting_proxy_test.go:60: pairing under an unlisted name through the proxy: 200, passcode in the answer: true
  ```
  The test asserts 403 on both, which no docs-only fix turns green; with a docs fix it becomes the documented demonstration (expectation inverted, or kept as a test of the option below).
- **Proposed fix (owner's choice, two shapes):**
  1. Docs and godoc only: `docs/embedding.md` "Host names", `docs/contract.md` "Host", `Options.AllowedHosts` godoc — "a proxy that sends the upstream's address as `Host` (nginx's `proxy_pass` default, Apache without `ProxyPreserveHost`) makes every request arrive under an admitted address: the check is then the proxy's — it must refuse names it does not serve — or the proxy is set to forward `Host` and the name is listed." A changelog line for the hosts.
  2. In addition an `Options` field that drops rules (a) and (b) for a host that knows it sits behind a proxy (names from `AllowedHosts`/`AllowHost` only); then the red test turns green with that field set. New `Options` field, no contract artefact.
- **Contract:** no (docs; shape 2 adds an `Options` field only).

---

## UR-P4-auth-01 — low — a non-default node's handler has no `AllowedHosts`: under a listed `Origin` its status and stream answer a non-admin

- **Where:** `fabricui/handler.go:596-626` (`sibling` copies `admin`, `token`, `streams`, `hostDone` … but not `allowedHosts`/`allowHost`, set only at `:315-316`); used by `handler.go:746` (`requestAdmin`) → `:761` → `host.go:56` (`h.allowedHosts[name]` on a nil map) from `handlers_node.go:150` (status) and `stream.go:40` (stream). Registry nodes take the same path (`registry.go:182`, `:195`).
- **Claim (M).** Decision 38: on an open route an `Origin` passes when its host meets the name rules, "the list" included. For `/nodes/{node}/status` and `/nodes/{node}/stream` of every node but the default the list is empty, so a request under a name of `Options.AllowedHosts` that carries an `Origin` of that name is an admin's on `/node/status` and `/nodes/default/status` and a non-admin's on `/nodes/second/status`: `admin:false`, `config.store` and warning texts withheld, the fixed attestation sentence, and a node stream filtered to three event types. The admin routes are not affected (their gate runs on the default handler, `handler.go:537-545`), loopback and IP-literal origins neither. It fails closed. Who meets it: a client that sends `Origin` on a GET (a host's own client, a proxy that adds it); browsers send none on a same-origin GET (I), so the shipped page does not.
- **Test:** `<scratch>/fabricui/p4s_sibling_origin_test.go`. Red:
  ```
  --- FAIL: TestP4S_SiblingStatusAdminUnderAListedOrigin
      p4s_sibling_origin_test.go:25: GET /nodes/second/status: 200, want admin true: {"enabled":true,…,"admin":false,…}
  --- FAIL: TestP4S_SiblingStreamFilteredUnderAListedOrigin
      p4s_sibling_origin_test.go:76: /nodes/second/stream: an admin's stream got [window_closed], want [exposure_changed window_closed]
  ```
- **Proposed fix:** `sibling` copies `allowedHosts` and `allowHost` (or the admin test of an open route always runs on the default handler, as the gate does); keep both tests, plus one over a registry node.
- **Contract:** no.

## UR-P4-bound-02 — low — a commissioning without `timeout_seconds` has no bound; the contract names "the controller's default", which go-fabric v0.23.0 says does not exist

- **Where:** `fabricui/handlers_controller.go:1420-1427` (a deadline only when the request sends `timeout_seconds`), `:1454`, `:1463`; adapter `fabricui/gofabric/gofabric.go:354` (`s.ctrl.Pair(ctx, …)`), `:618` (`reg.Commission(ctx, …)`), neither adds one; contract `contract/openapi.yaml:3526` ("The whole commissioning's bound; absent is the controller's default."). Other side: `gf:controller/commissioner/commissioner.go:326-331` ("ctx bounds the whole call … the total by the device's MaxCumulativeFailsafeSeconds (up to about 18 hours) … a caller that wants a shorter bound gives ctx a deadline (review II, R2-P6-bound-1812)"), `gf:controller/controller.go:1166-1170` and `:1225-1235` (`DefaultPairTimeout` bounds the *search* only). The shipped page sends no `timeout_seconds` (grep over `ui/src`: only the window's field).
- **Claim.** M: the ctx both facade calls get has no deadline. I: a device that answers each step just inside its timeout holds `POST /controller/commission` (one goroutine, the claimed node id, the page's spinner) for as long as the admin's connection stays, up to what the device itself states. Admin-started, device-controlled; no secret, no other role. v0.22.2 behaved the same; v0.23.0 states the caller's duty (delta of `commissioner.go`).
- **Test:** `<scratch>/fabricui/p4s_commission_bound_test.go`, `TestP4S_CommissionWithoutTimeoutHasNoBound`. Red:
  ```
  p4s_commission_bound_test.go:42: Pair: the commissioning's ctx has no deadline
  p4s_commission_bound_test.go:45: CommissionDiscovered: the commissioning's ctx has no deadline
  ```
- **Proposed fix:** a default deadline in the handler when the request names none (the contract's maximum, 600 s, is the natural one), and the contract sentence says that number.
- **Contract:** yes (the description of `CommissionRequest.timeout_seconds`; no guard break).

## UR-P4-web-04 — note — I — the token's same-origin test resolves against `location`, `fetch` against the document base

- **Where:** `ui/src/lib/api/client.ts:670-678` (`new URL(base, location.href)`), against `:466` (`fetchImpl(base + path, …)`, which the browser resolves against `document.baseURI`). A wider reach of UR-P3-web-194.
- **Claim.** In a host's own page with `<base href>` of another origin, a relative `baseUrl` passes `isSameOrigin` and the request, token header included, goes to the base's origin. Not the shipped page: its policy has `base-uri 'none'` (`fabricui/spa.go:146`). Not executed.
- **Proposed fix:** resolve against `document.baseURI` when there is a document. **Contract:** no.

## UR-P4-auth-05 — note — I — `*.localhost` is admitted on the assumption that the client resolves it to loopback

- **Where:** `fabricui/host.go:50`. Browsers that pin `*.localhost` to loopback make the rule safe; a client that asks its resolver (older WebViews, non-browser clients) lets whoever answers its DNS rebind `x.localhost`. No public zone exists under `.localhost`, so this needs the victim's resolver. No change proposed beyond a godoc clause; decision 37 (a) stands. **Contract:** no.

---

## Adapter against go-fabric v0.22.2 → v0.23.0 (no finding beyond UR-P4-bound-02)

- `controller.Pair` (`gf:controller/controller.go:1184-1247`): the adapter passes the template; `pairOptions` (`:1262-1276`) still defaults `attestation.Reject()`.
- `CommissionDiscovered` (`gofabric.go:568-627`) against `pairOptions`: same members (passcode, policy, network, terms, progress), the unset policy rejecting (`:609-612`; `gf:controller/attestation/policy.go:18-34`, `Reject` is `OnFailure=&false`) before `Edit`. New in v0.23.0 and not used, no defect: `CommissionOptions.OnboardingCode`, `Result/Record.AttestedVendorID/ProductID` (the page shows the announced ids only; a feature, not a finding).
- `OpenCommissioningWindow`/`RevokeCommissioningWindow` (`gf:controller/client/window.go`): comments only in the delta. v0.23.0's kept onboarding codes (`client.Node.OnboardingCode`, `SecretCode`) are not read by the adapter; `nodeRecord` (`node.go:60-74`) maps named fields only.
- No go-fabric error of the pairing path formats a passcode or code (grep over `controller/`, `controller/commissioner`, `controller/client`, `client/discovery` at v0.23.0); the adapter replaces `ParsePairingCode`'s error (`gofabric.go:333`).
- `softwareupdate.Manager` (`QueuedUpdates`, `QueryUpdates`, `AddUpdateConsent`, `Observe`, `ErrMismatch`), `Registry.Decommission/SetFabricLabel/GetByAddress/Commissioned`, `Fabric.SetLabel` (`commissioner.ErrLabel`): signatures and the mapped errors unchanged for the adapter's use.
- Not run here: `make lint` (SA1019, decision 62) and the adapter's e2e — the coordinator's guard set.

## Covered without finding

Line by line (delta first): `fabricui/host.go`, `fabricui/handler.go`, `fabricui/strict.go`, `fabricui/stream.go`, `fabricui/events.go`, `fabricui/problem.go`, `fabricui/spa.go`, `fabricui/handlers_pairing.go`, `fabricui/gofabric/gofabric.go`, `fabricui/gofabric/updates.go`, `ui/src/lib/api/stream.ts`, `ui/src/lib/api/client.ts:354-690`.
By delta (`f4b8573..979a67a`) plus the sections named: `fabricui/options.go` (delta, `:354-417`), `fabricui/handlers_controller.go` (delta, `:30-400` gate and status, `:1394-1530` commission and discovery), `fabricui/handlers_node.go:28-330` (status), `fabricui/operations.go:166-287`, `fabricui/registry.go:177-252`, `fabricui/gofabric/node.go` (delta, record, window, events, ping), `fabricui/controller.go` (delta), `ui/src/lib/store.svelte.ts`, `ui/src/lib/controller-store.svelte.ts`, `ui/src/lib/context.ts`, `ui/src/shell/route.ts` (delta each), `ui/src/shell/App.svelte` (delta, `:60-215`).
Checked and holding: the open routes are exactly the contract's eight (script over `contract/openapi.yaml`); the gate runs before node lookup, parameters and body on every admin route, HEAD included; `Host` parsing (ports, brackets, trailing dot, zone, empty); the `Origin` rule incl. `AllowHost`'s equality; token compare; the stream's type check and one-line data, the non-admin allow-list, host events admin-only; the cap and its 503; the non-admin status per decisions 34, 46, 57, 63; the two `{@html}` sinks (`Pair.svelte:174`, `Window.svelte:102`, both `renderQrSvg`, unchanged) and no other raw-HTML or script sink (grep); dynamic `href`s are constant hash routes plus an id; the token leaves the client only to a same-origin URL.
Not re-reported (known): `health_note`, `source_note`, `exposure_reason` open to a non-admin (phase-3-premerge, decision 63's scope); the stream cap shared by open and admin streams (UR-P3-bound-197); the six carried rows.

## Tokens

Not reported by the harness to this agent; estimate ~330 k.
