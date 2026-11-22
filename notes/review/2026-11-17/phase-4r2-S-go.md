**No held candidate: this report carries no finding of severity high or medium. It lies in `held/` beside the other unit reports of run 2 until the coordinator publishes them.**

# Phase 4, run 2 — unit S-go (the Go security paths and the adapter against go-fabric v0.23.0)

HEAD read `792d15b` (the review branch adds notes only: `git diff --stat 792d15b HEAD` names three files under `notes/`). go-fabric `v0.23.0` (`gf:`, the `git archive` export). Scratch worktree
`<scratch>` (detached at `792d15b`, two untracked test files, left in place).
**M** = measured (line read, or test run), **I** = inferred.

Counts: 0 high, 0 medium, 2 low, 1 note.

Red run, both files:
`cd <scratch> && PATH=$HOME/go/bin:$PATH CGO_ENABLED=0 go test ./fabricui ./fabricui/gofabric -run 'TestP4R2S_' -count=1` → `FAIL` in both packages (2 tests, 2 failing; outputs below).

## Scope, listed first

| File | Lines | Read |
| --- | --- | --- |
| `fabricui/handler.go` | 831 | whole, delta first |
| `fabricui/strict.go` | 200 | whole |
| `fabricui/options.go` | 431 | whole, delta first |
| `fabricui/host.go` | 122 | whole, delta first |
| `fabricui/stream.go` | 195 | whole |
| `fabricui/events.go` | 394 | whole |
| `fabricui/problem.go` | 171 | whole |
| `fabricui/spa.go` | 219 | whole |
| `fabricui/gofabric/gofabric.go` | 811 | whole, delta first |
| `fabricui/gofabric/hooks.go` | 300 | whole, delta first |
| `fabricui/gofabric/node.go` | 867 | whole |
| `docs/embedding.md` | — | `:56-60` (Options rows), `:64-110` (Admin rights), `:440-523` (Behind an Ingress, headers and framing), delta first |
| the gate, route table, token mode, streams, a non-admin's reads outside those files | — | `handlers_controller.go:36-215`, `:350-415`, `:1780-1890`; `handlers_node.go:40-290`, `:706-728`; `operations.go:160-290`; `registry.go:128-256`, `:366-390`; `docs/contract.md:424-470` |

---

## UR-P4-leak-200 — low — a request without admin rights that is abandoned makes the status log a fabric failure and its recovery, once per pair of requests (wider reach of UR-P4-leak-91)

- **Where:** `fabricui/handlers_node.go:229-235` (the status's fabric read on `r.Context()`, then `logFabricsState`), `:710-728` (`logFabricsState`: logs when the error text differs from the last one, and stores it). Other side: `gf:node/node.go:1752-1756` (`FabricInfos` → `n.st.ListFabrics(ctx)`), `gf:store/fabric.go:269-275` (`s.db.QueryContext(ctx, …)` → `store: list fabrics: context canceled` under a cancelled context). The rule that exists for the same case elsewhere: `fabricui/problem.go:125-133` (`serverError`: a `context.Canceled` under an ended request context is an abandoned request, logged at debug, UR-P1-err-12).
- **Claim.** UR-P4-leak-91's fix logs on a change of state instead of per request. A cancelled request context is not a state of the node, but `logFabricsState` takes it for one: on a node whose store is SQLite (what both hosts run; the in-memory store of the default `fabrictest` device ignores the context, which is why the fix's own test does not see it), a request whose context ends before the fabric read answers the read with `context canceled`; the handler logs `fabricui.status.fabrics_failed` and stores the text; the next plain request logs `fabricui.status.fabrics_recovered`. Two lines per pair, as often as the client repeats it, on the open routes `GET /node/status` and `/nodes/{node}/status`.
  - M: a client that ends its sending side right after the request (TCP half-close; net/http cancels the request context on the read's EOF) does it every time: 25 pairs → 25 warnings and 25 info lines. The client still reads its answer, which then lacks `fabric_count` (25 of 25).
  - I: behind a proxy the same by aborting the request (the proxy closes its upstream connection); and the shipped page does it without intent whenever a status fetch is aborted, so an administrator's own log gains a false "fabrics_failed … context canceled" beside the rotation.
  - Reach as in UR-P4-leak-91: the ring those lines land in is read by administrators only and rotates (`operations.go:335-338`); with a file logger it is disk. No secret, no admin action.
  - Sibling path, I, not tested: `fabricui/stream.go:116-124` logs `fabricui.stream.flush_failed` at warn for every stream opened through a host middleware without `Unwrap` — per request and open to a non-admin too, but only in a setup where no stream works at all.
- **M/I:** M (test below, real node on a SQLite store through `fabrictest.Options{StoreOnDisk: true}`, `fabricui.Handler`, host mode with `Admin` false); the proxy and page variants I.
- **Test:** `<scratch>/fabricui/p4r2s_abandoned_log_test.go`, `TestP4R2S_AbandonedStatusRequestsWriteTheLogPerRequest`. Red:
  ```
  p4r2s_abandoned_log_test.go:92: 25 pairs of status requests without admin rights (one half-closed, one plain) wrote 25 fabrics_failed and 25 fabrics_recovered lines; 25 answers lacked fabric_count; first: fabricui.status.fabrics_failed error="store: list fabrics: context canceled"
  ```
- **Proposed fix:** in `status` (or `logFabricsState`) leave the state and the log alone when `errors.Is(err, context.Canceled)` (or `DeadlineExceeded`) and `r.Context().Err() != nil` — `serverError`'s abandoned rule; optionally read the fabrics on a context detached from the request with a short bound, so a half-closing client still gets `fabric_count`. Keep the test (it needs the SQLite store).
- **Contract:** no.

## UR-P4-adpt-201 — low — a node that leaves the controller's registry without a decommission raises no `node_removed`; its hooks keep emitting under the old id

- **Where:** `fabricui/gofabric/hooks.go:197-217` (`Hooks.For`: `Decommissioned` → `removed`, no `Deleted`), `:108-116` (`removed` is the only place that forgets `ids` and the pending timer), `:128-152` (`nodeID` falls back to the remembered id). Other side: `gf:controller/client/config.go:151-170` (`Decommissioned`: a registry's decommission or the peer's leave; `Deleted`, new in v0.23.0: "the node left its Registry — Registry.Forget, Registry.Erase, a decommission, a peer's leave, the cull of an expired commissionable node, or a CompleteCommissioning that failed"), `gf:controller/client/registry.go:455-469` (`remove` runs `Deleted`), `:477` (`Forget` → `deleteNode`, which runs no `Decommissioned`: `:482-520`), `:625-637` (`Erase`).
- **Relation to known rows:** UR-P4-adpt-31 fixed `Wrap`, which now chains the host's `Deleted`; its report named this as "optional, separate" and no register row carries it. This is the adapter's own subscription, a different path.
- **Claim.** M: after `Registry.Forget` of a commissioned node, `GET /controller/nodes` (the source's `Nodes()`) no longer lists it and the controller stream carries no `node_removed`; the events that do arrive (`connection`, `connection`, `node_updated`) are emitted under the removed node's id out of `Hooks.ids`, which is never cleared for it. A page keeps the node in its list until a reload, and a reload it makes on `node_updated` asks for a node that answers 404. I: `Registry.Erase` goes the same way (same `deleteNode`/`remove`, read, not run). I: once go-fabric hands the local id to a new node (`gf:controller/client/registry.go:665-671`, "a node created later may receive the same id"), hooks of that node that run before it has a peer address are published under the old node's id.
- **Precondition (why low):** something calls `Registry.Forget` or `Erase` on a commissioned node. The adapter and the handler never do; go-fabric's facade does so only at start, before `ControllerOf` (`gf:controller/ota_provider.go:265-267`, the controller's former own node); the local checkouts of go-fabric-server and go-ha2fabric do not call either (grep, unpinned, I). A host that adopts them — or a go-fabric release that removes a node this way at run time — leaves the stream silent.
- **Test:** `<scratch>/fabricui/gofabric/p4r2s_forget_test.go`, `TestP4R2S_AForgottenNodeRaisesNoNodeRemoved` (the e2e site: a real controller, a device commissioned through the source's `Pair`, then `ctrl.Registry().Forget`). Red:
  ```
  p4r2s_forget_test.go:62: the node left the registry (Forget) and the source lists 0 nodes, but no node_removed reached the stream in 3 s; events since: [connection connection node_updated]
  ```
- **Proposed fix:** `Hooks.For` sets `Deleted: func() { h.removed(id) }` and `removed` emits `node_removed` only while `ids` still holds the node (a decommission runs `Decommissioned`, then `Deleted`: one event, not two); the guard test of UR-P4-adpt-31 stays. Keep this test.
- **Contract:** no (`node_removed` exists: `contract/openapi.yaml:2117`, `docs/contract.md:936`).

## UR-P4-doc-202 — note — the third kind of proxy "passes every request"; with a name of its own that the host did not list, every request is refused, and no text says to list it any more

- **Where (docs side):** `docs/embedding.md:470-472` ("It replaces `Host` — with a name of its own or with the upstream's address: the check is constant, every request passes it whatever name the browser used."); the same wording without "passes" in `docs/contract.md:458-461` and the `AllowedHosts` godoc `fabricui/options.go:312-315`.
- **Where (code side):** `fabricui/host.go:49-60` (`nameAllowed`: a name passes only as a loopback name, an IP literal or an entry of `AllowedHosts`), `:64-76` (`admitHost`); the docs guard itself knows the condition: `fabricui/host_proxy_test.go:17-23` ("a name of its own **the host listed**"), case `"a listed name of its own"`.
- **Claim (M, both sides read).** The check is constant in both sub-cases, but only the upstream's address (and a listed name) is constantly admitted; a proxy that sends `Host: fabric-upstream` to a handler that does not list that name gets 403 `host_not_allowed` on every route. Run 1's text said "A proxy that sends a name of its own as `Host`: list that name in `AllowedHosts`"; decision 69's rewrite (delta `979a67a..792d15b`) dropped the sentence. Fails closed; a host sees it at its first request, so no unsafe setup follows — the docs lost one instruction.
- **Proposed fix:** "… with a name of its own, which the host lists in `AllowedHosts`, or with the upstream's address: the check is constant …" in `docs/embedding.md`; the other two places need no change.
- **Contract:** no.

---

## Adapter against go-fabric v0.23.0, every facade call read at the pin (no finding beyond UR-P4-adpt-201)

- `Controller.Pair` (`gf:controller/controller.go:1184-1247`, `pair_race.go:41-116`): returns the node beside the error of its start (`:1241-1243`, `pair_race.go:75-77`), which `commissionedDespite` needs; `pairOptions` (`:1262-1276`) defaults `attestation.Reject()`; the search bound only without a deadline (`:1227-1234`, decision 71's addendum).
- `CommissionDiscovered` (`gofabric.go:568-627`) against `pairOptions`: the same members, the unset policy rejecting before `Edit` (`gf:controller/attestation/policy.go:18-34`); `Registry.ForDevice` (`gf:controller/client/addresses.go:165`), `Registry.Commission` (`fabricops.go:176-259`: `ErrAlreadyCommissioned`, `Commissioned` hook, Subscribe 21.1 reported started before `Start`). Known and not re-reported: no trace entry and no metric (UR-P3-adpt-92, go-fabric-ui#81), `ErrAlreadyCommissioned` → 502 (UR-P2-adpt-190, phase-3-D).
- `Controller.Discover` (`gf:controller/code.go:106-150`): 0 is 15 minutes; the adapter always passes a bound (`rediscoverWindow`, the handler's 1–60 s).
- `Registry.Commissioned/Get/GetByAddress/Decommission/SetFabricLabel` (`registry.go:651-690`, `fabricops.go:267-277`, `label.go:128`), `Fabric.SetLabel` (`commissioner.ErrLabel`), `runFabricOperation` (`fabricops.go:70-98`: `ErrClosed`, `ErrFabricOperationInProgress` → 409 by `nodeError`).
- `client.Hooks` (`config.go:109-182`): fourteen fields, all chained by `chain` (delta, UR-P4-adpt-31's fix; the reflection guard holds the list).
- `Node.Read/Write/Invoke` (`interact.go:35-40`, `:67`, `:261-309`; `client/interaction/write.go:48-78`, `:137-208`): `Write` returns statuses or an error, never both, so `node.go:586-603` loses nothing; a zero `AttributePath` is the wildcard (`surface.go:133-136`), which `Interview` sends; `TimedTimeout` > 0 marks a write timed. `ErrRequestTooLarge`, `ErrMalformedRequest` (`interact.go:24-28`, the delta's 400 mapping), `ErrJSONValue`, `ErrICDOperation`, `ErrDisabled`, `ErrNoSession`, `ErrNotCommissioned`, `ErrClosed`: present, none deprecated.
- `Node.OpenCommissioningWindow`/`RevokeCommissioningWindow` (`window.go:48-94`, `:181`), `PingAddress` (`ping.go:32-36`), `Enable/Disable/SetAutoSubscribe/AutoSubscribe` (`node.go:665-735`), `EventHistory` (`event_history.go:100-113`), `ICD()` with `ICDRegisterOptions` (`icd.go:130-141`), `discovery.Device` (`client/discovery/scanner.go:141-154`), `controller.PairingCode` (`code.go:28-51`).
- Decision 62: no identifier the three files use carries `Deprecated:` at v0.23.0 (grep over `controller/`, `controller/client`, `controller/commissioner`, `client/discovery`, `node/`, `schema/`, `thread/`); `make lint` was not run here.

## Covered without finding

Read whole, line by line (delta first where the delta map names the file): `fabricui/handler.go`, `fabricui/strict.go`, `fabricui/options.go`, `fabricui/host.go`, `fabricui/stream.go`, `fabricui/events.go`, `fabricui/problem.go`, `fabricui/spa.go`, `fabricui/gofabric/gofabric.go`, `fabricui/gofabric/hooks.go`, `fabricui/gofabric/node.go`; the delta `979a67a..792d15b` of all eleven and of `docs/embedding.md` as a diff.

Read in the ranges named in the scope table: `fabricui/handlers_controller.go` (gate, admin table, status, controller stream), `fabricui/handlers_node.go` (status, `logFabricsState`), `fabricui/operations.go` (info, level, log stream), `fabricui/registry.go` (the lookups an open route reaches, `logSkipped`), `docs/embedding.md`, `docs/contract.md` (Admin rights, Host, Origin, Order).

Checked and holding:
- **Gate and route table:** 40 generated controller routes, 39 in `controllerAdmin`, `GET /controller/status` the one open route (script over `internal/api`'s `HandleFunc` lines against the table); the gate reads the pattern from the same `mux.Handler` lookup `ServeHTTP` uses, HEAD on a GET pattern included; node routes are gated before the node lookup; the open routes are the status, `GET /nodes`, the node stream, `/operation/info`, `GET /operation/log/level` and the controller status.
- **Host check:** runs once per request (the admission in the context, key unexported), covers the bundle through `NewHandler`; `sibling` now carries `allowedHosts` and `allowHost` (delta, UR-P4-auth-01's fix), registry nodes through it; a name passes only as `localhost`, `*.localhost`, a zone-less IP literal or a listed name; an absolute-form request target or `:authority` is what `r.Host` holds.
- **Origin and admin test:** Origin before the host's `Admin`, which is asked only with `X-Fabric-UI: 1`; `null`, userinfo, a path and an unparsable origin are refused; the `AllowHost` equality is host and port; the token compare is constant-time and the token is in the two statuses only. An origin of another local port or an IP literal passes the name rule as the contract says (decision 38); what keeps such a page out is the request header and the token, which need a preflight no route answers.
- **Streams:** the cap is taken before the subscription and released on every return; HEAD subscribes nothing; the non-admin filter wraps the handler's hub, the nodes hub and the node's observers, and the host's source is not subscribed at all; the event type is checked before it is written; the controller and log streams are behind the gate; the write deadline does not outlive the stream on a reused connection (phase 2's pre-merge measurement, `phase-2-premerge.md:110-112`, not repeated).
- **A non-admin's reads:** the node status per decisions 34, 46, 63 (`store`, warning messages, the attestation text), the controller status per decision 57 (identity zeroed, `admin` or the token), the node stream per decisions 41 and 58.
- **Bodies:** `decodeStrict` (one value, unknown and differently cased members, nested), `jsonBody`/`hasBody` (a chunked body peeked and kept), 1 MiB.
- **Bundle:** an unknown path and a directory are 404, the index alone carries the policy, `FrameAncestors` validated against any-parent entries.
- **Docs (decision 69):** the three kinds of proxy, what a rebinding page reads and writes in each, and the `AllowedHosts`/`AllowHost` godoc agree with `host.go` and `handler.go:759-773` — but for UR-P4-doc-202. A write of the shipped page carries its real origin: `fetch`'s mode is `cors`, so `Referrer-Policy: no-referrer` does not turn it into `null` (I, from the Fetch standard; the e2e-web smokes write under that header).

Not re-reported (known): UR-P3-bound-197 (one cap for open and admin streams); UR-P4-auth-05 (`*.localhost`); the admin test once per stream (`docs/contract.md:425-426`); `health_note`, `source_note`, `listen`, `exposure_reason`, the counters open to a non-admin (decisions 34, 63); UR-P3-adpt-92 and go-fabric-ui#81; decision 73's own window after a close; the three rows carried to run 2 (UR-P4-adpt-197, UR-P4-ui-61's other-tab path, UR-P4-ctr-196's mapping — `handlers_node.go:215-217` is its site).

## Not covered

- `fabricui/handlers_controller.go`, `handlers_node.go`, `operations.go`, `registry.go` outside the ranges above; `fabricui/handlers_pairing.go`, `controller.go`, `nodeswitch.go`, `memory_controller.go`, `node_detail.go`, `handlers_diagnostics.go`, `handlers_exposure.go`, `handlers_surface.go`, `handlers_topology.go`, `provider.go`, `ports.go`; `fabricui/gofabric/updates.go`, `topology.go`, `diagnostics.go` — unit G's by the brief's cut.
- `contract/openapi.yaml` against the route tables: not read; the admin table was counted against the generated routes only, the `x-admin` marks are the guard test's (`TestControllerAdminMatchesTheContract`) and unit G's.
- `internal/api/api.gen.go`: only the `HandleFunc` lines.
- `docs/embedding.md` outside the three ranges; `docs/contract.md` outside `:424-470`.
- go-fabric at v0.23.0: the facade calls listed above at their definitions and doc comments; `client/interaction/write.go` by its return sites, not whole; `controller/pair_race.go:41-116` only; no read of `controller/commissioner`.
- Not run: `make lint` (SA1019), `make e2e`, `make race`, any browser. The proxy and page variants of UR-P4-leak-200 and the `Erase` and recycled-id halves of UR-P4-adpt-201 are inferred, not run.

## Tokens

Not reported by the harness to this agent; estimate ~300 k.
