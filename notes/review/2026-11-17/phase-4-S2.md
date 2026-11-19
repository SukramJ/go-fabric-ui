**No held candidate: this report carries no finding of severity high or medium. It lies in `held/` beside unit S's report, whose gaps it reads (decision 72).**

# Phase 4, run 1 — unit S2 (the security reader of unit S's gaps)

HEAD read `979a67a`. go-fabric `v0.23.0` (`gf:`, the `git archive` export). Scope: the list "What unit S's partial coverage leaves unread" of `held/phase-4-verify.md`, line by line. Scratch worktree
`(scratch worktree)` (detached at `979a67a`, four untracked test files `fabricui/p4s2_*_test.go`, left in place).
**M** = measured (line read, or test run), **I** = inferred.

Counts: 0 high, 0 medium, 2 low, 1 note.

Runs: `cd <scratch> && PATH=$HOME/go/bin:$PATH go test ./fabricui -run 'TestP4S2_' -count=1 -v` → 5 tests, 3 red (the findings below), 2 green (the two measurements under "Aspects").

---

## UR-P4-auth-90 — low — on a node without a fabric, the close (and the expiry) of a fresh window puts the node back into its own 48-hour window on the configured passcode; the close answers 204 "Closed.", and an operator's earlier close is undone

- **Where:** `fabricui/handlers_pairing.go:286-304` (`closeWindow`: one `h.node.closeWindow`, then 204), `:255` ff. (`openFresh`); godoc `fabricui/options.go:269-274` (`ConfiguredWindow`: "The default opens it with fresh random credentials (node.OpenBasicWindow), which no earlier reader of the configured codes can use"); `docs/embedding.md:54`; contract `contract/openapi.yaml:152-159` ("Close the open commissioning window." / `"204": { description: Closed. }`); `docs/contract.md:26`. Other side: `gf:node/commissioning.go:654-659` (`OpenBasicWindow`: "Once it closes, the configured credentials are back: a node without a fabric reopens its own window under the configured discriminator (not with [Commissioning].WindowOnDemand)"), `:728-731` (`CloseWindow`: "A node without a fabric that closes a basic window reopens its own window, as when that window closes on its timeout"), `:714-719` (`OpenOwnWindow`: "timeout 0 is 48 h while the node has no fabric"), `gf:node/config.go:509-521` (`WindowOnDemand`).
- **Claim.** In the default mode (fresh credentials) on a node that has no fabric and does not set `WindowOnDemand`:
  - M: `POST /node/pairing/window/close` on the fresh window answers 204; the next status reads `commissioning_window_open: true`, `window_kind: "own"`, `window_remaining_seconds: 172800`, and `GET /node/pairing` answers `source: configured`, `window_open: true`, `window_match: true` — the node accepts PASE on its configured passcode for 48 hours.
  - M: this holds also when the operator had closed the node's own window before (`window_kind: "none"` after that first close, measured): one open and close of a fresh window undoes it. A second `POST …/close` closes the own window again (measured in `p4s2_hostmode_e2e_test.go`: after two closes a fresh window opens, after one it answers 409 "already open").
  - I (go-fabric's doc lines above, not run): the same on the fresh window's timeout, with nobody at the page.
  - Neither side of the contract says so: the route's 204 says "Closed.", and the godoc's promise for the default mode ("no earlier reader of the configured codes can use") holds only while the fresh window is open. A client other than the shipped page that takes the 204 as "not commissionable" is wrong for a node without a fabric.
  - I (read of `ui/src/lib/store.svelte.ts:390-396`, `:434-447`, `:348-388`): the shipped page does show it — the own window's `window_opened` re-hydrates the flow and the Pair page shows the configured codes with a 172800 s countdown and Close again. Phase 2 saw the expiry half and judged the page consistent (`phase-2-X1.md:678-680`, covered without finding); the close route's answer, the undone close and the godoc sentence are not in a register.
- **Why low, not medium:** the node is in the state every node without a fabric starts in (its own window, go-fabric's documented design, which the host can turn off with `WindowOnDemand`); the attacker needs the configured passcode and the network; the shipped page shows the open window. What is wrong is the answer and the documentation, and that a deliberate close does not hold. The verifier may weigh the undone close higher.
- **Test:** `<scratch>/fabricui/p4s2_window_close_test.go`, `TestP4S2_CloseOfAFreshWindowLeavesTheNodeOpenOnTheConfiguredCodes` (real node through `fabrictest`, `fabricui.Handler`, host mode). Red:
  ```
  p4s2_window_close_test.go:45: close of the fresh window answered 204 ""; the status then: open true, kind "own", 172800 s left, 0 fabrics
  p4s2_window_close_test.go:55: GET /node/pairing after the close: source "configured", window_open true, window_match true: the configured codes commission
  ```
- **Proposed fix (owner's choice):**
  1. Docs and contract text: `ConfiguredWindow`'s godoc, `docs/embedding.md` (the `ConfiguredWindow` row and a line under the pairing section), the `closeWindow` description and `docs/contract.md` — "on a node without a fabric the node's own window on the configured codes comes back when a fresh window closes or expires, and after the last fabric is removed (factory reset, `DELETE …/fabrics/{index}`); a host whose operators drive the window from the page sets go-fabric's `Commissioning.WindowOnDemand`". The red test becomes the documented demonstration (expectation inverted).
  2. In addition the close route closes until `Window().Open` is false (a second `CloseWindow` when the first left the own window open), so an operator's Close means closed; the expiry path stays go-fabric's and is covered by 1. With this the red test turns green as written.
- **Contract:** yes for the description of `closeWindow` (text only, no guard break); shape 2 changes behaviour within the declared statuses.

## UR-P4-leak-91 — low — a request without admin rights writes one warning per call into the log while a node's fabric read fails or the registry lists an unusable key

- **Where:** `fabricui/handlers_node.go:226-231` (`fabricui.status.fabrics_failed`, on the open `GET /node/status` and `/nodes/{node}/status`), `fabricui/registry.go:150-155` (`fabricui.registry.entry_skipped`, in `registryEntries`, which runs on the open `GET /nodes` and on every request under `/nodes/{key}/…`, `handler.go:684`, `registry.go:177-252`). The ring those lines land in is read by administrators only (decision 20; `operations.go:27-41`, `MemoryOperations` keeps the newest `capacity` records, `:335-338`).
- **Claim (M).** Both lines are written per request, not per failure. The two states are ones the docs name as reachable: a stopped node behind a `NodeSwitch` whose host did not `Set` a new one answers "the erased store's errors" (`nodeswitch.go:21-25`), a locked store the same; a registry entry with an invalid or colliding key is "logged and left out" (`registry.go:136-139`). While either lasts, anyone who reaches the listener — no rights needed, both routes are open — writes a warning per request; the shipped page does it too with every status reload (`store.svelte.ts:227-233`, per health sample). The ring rotates and the records an administrator would look for are gone; with a file logger the volume is the requester's to choose. No secret is disclosed: the error text (store paths) stays in the admin-only log.
- **Test:** `<scratch>/fabricui/p4s2_log_test.go`. Red:
  ```
  --- FAIL: TestP4S2_ANonAdminWritesTheLogOncePerStatusRequest
      p4s2_log_test.go:35: 50 status requests without admin rights wrote 50 warnings (fabricui.status.fabrics_failed); want one per failure, not one per request
  --- FAIL: TestP4S2_ANonAdminWritesTheLogOncePerNodeListRequest
      p4s2_log_test.go:61: 50 GET /nodes without admin rights wrote 50 warnings (fabricui.registry.entry_skipped); want one per entry, not one per request
  ```
- **Proposed fix:** log on the change of state, not on the request: the status remembers the last fabric-read error text per handler and logs when it changes (and once when it clears); `registryEntries` remembers the set of skipped keys and logs a key when it enters the set. Keep both tests.
- **Contract:** no.

## UR-P4-doc-92 — note — the comment over `ListNodeEvents` says "no admin"; the route is `x-admin`

- **Where:** `fabricui/handlers_controller.go:614-616` ("no interaction, no admin") against `:86` (`"GET /controller/nodes/{id}/events": true` in `controllerAdmin`) and `contract/openapi.yaml:1354-1361` (`x-admin: true`, decision 35). The gate holds (measured in the sweep below: 403 for a request without rights); the comment is from before 0.10.0.
- **Proposed fix:** drop "no admin" from the comment. **Contract:** no.

---

## Aspects the verifier named, measured without finding

- **Host mode end to end with a real `Admin`** — `<scratch>/fabricui/p4s2_hostmode_e2e_test.go`, `TestP4S2_HostModeEndToEnd`, green: `fabricui.Handler` over a real node (`fabrictest`), a second node entry, `MemoryController`, `MemoryOperations`, `Warnings`, `NodeConfig`; `Admin` answers from a header only the proxy sets. Every operation of `contract/openapi.yaml` (scanned, ≥ 90) as a request without the role, alone and with five decoys (a token header, the role in another case, the role in a list value, a look-alike header name, `X-Forwarded-For`/`Forwarded` of loopback): every `x-admin` operation 403 `forbidden`, the open ones exactly eight and 200, no `csrf_token` in host mode, and none of the planted secrets in any answer (the node window's passcode, QR and manual code; the controller window's; a Wi-Fi passphrase and SSID; a log record's message and attribute; `config.store`; a warning's text). After the refused requests the admin's window still serves and the log level is unchanged. The role without `X-Fabric-UI: 1` is refused; a foreign `Origin` with the role is refused on an admin read and reduced on the status. Every admin GET answers without 403 or 500. A non-admin's node stream on both nodes carries only `window_closed`/`window_opened` across an admin's close and open and none of the secrets; the controller and log streams answer it 403. The handler's log at debug carries none of the secrets.
  (One probe of mine was wrong and is dropped: a header value with a trailing space is the same value after HTTP's trimming.)
- **The stream cap under admin and open streams together** (known row UR-P3-bound-197, accepted by decision 42, not re-reported) — `p4s2_streamcap_test.go`, green as a measurement: 64 node streams without rights fill the cap; an administrator's node, controller and log streams then answer 503 with `Retry-After`; the other routes go on answering; a freed slot is reused at once.
- **`make lint`** in the scratch worktree at `979a67a`: exit 0, `0 issues.` (SA1019 included, decision 62). **`make e2e`**: exit 0, both packages (`fabricui` 4 tests, `fabricui/gofabric` 13 tests, the adapter's attestation refusal among them).

## Covered without finding

Read whole, line by line: `fabricui/handlers_controller.go:1-30`, `:200-260`, `:400-1394`, `:1530-2079` (with S's ranges the file is read); `fabricui/handlers_node.go:300-698`; `fabricui/operations.go:1-166`, `:287-418`; `fabricui/registry.go:1-177`, `:252-367`; `fabricui/options.go:1-417`; `fabricui/nodeswitch.go`; `fabricui/provider.go`; `fabricui/gofabric/node.go`, `hooks.go`, `topology.go`, `diagnostics.go`; `ui/src/lib/api/client.ts:1-353`; `ui/src/lib/api/json.ts`; `ui/src/lib/qr.ts`; `ui/src/lib/nodes-store.svelte.ts`; `ui/src/lib/store.svelte.ts`; `ui/src/lib/controller-store.svelte.ts`; `ui/src/lib/context.ts`; `ui/src/shell/route.ts`; `ui/src/shell/main.ts`; `ui/src/shell/index.html`; `ui/src/lib/pages/node/Pair.svelte`; `ui/src/lib/pages/controller/node/Window.svelte`.

Checked and holding:
- **Stored credentials:** `GET /controller/credentials` returns kind, id, name (the SSID or the dataset's network name) and `default` only (`handlers_controller.go:1545-1563`, `memory_controller.go:55-66`); passphrase and dataset have no read route; the adapter reads them only in `networkOf` for a commissioning (`gofabric.go:267-292`), whose errors name the credential id, not its content; `ThreadDatasetName`'s and `gf:thread`'s decode errors carry lengths and TLV names, no bytes (`gf:thread/dataset.go:81`, `:186`, `:265`).
- **What a log line may carry:** every log call of `fabricui/` and `fabricui/gofabric/` (grep, 20 sites) logs an error text, a path (`r.URL.Path`, never the query), a key, a count or a level; none logs a body, a header or a response. A JSON value that does not fit (`client.ErrJSONValue`) is `ErrInvalidArgument` → 400 to the administrator who sent it, not logged (`gofabric.go:776-777`, `handlers_controller.go:232-233`); only `ErrNodeFailed` and unknown errors are logged, and no go-fabric error under `controller/`, `thread/` formats an SSID, a passphrase or a dataset (grep at v0.23.0). The log buffer itself (`MemoryOperations`) stores what the host records; the stream and the ring are `x-admin`. Refusals (403) are not logged at all — so a refused requester cannot write the log that way; the two per-request lines are UR-P4-leak-91.
- **ACL, bindings, fabric removal, decommission, label, updates:** all behind the gate (`controllerAdmin`, sweep above); values validated before the port (`parseNodeID`, `parseHex32`, `endpointOf`, the group and index ranges, 32 bytes of label); the own fabric cannot be removed through `DELETE …/fabrics/{index}` (`:1343-1349`); decommission and the registry's delete run detached and bounded.
- **Topology and diagnostics answers** (`topology.go`, `diagnostics.go`): addresses, extended PAN id, network name, SSID/BSSID and host names reach `x-admin` answers only; `topology.Node` and `diag` carry no network key or credential; the events they cause (`topology_changed`, `controller_diagnostic_event`) go to the admin-only controller stream.
- **Node side:** factory reset and erase need their confirmation word, decode strictly, run detached; endpoint and event reads are `x-admin` and bound (`pathEndpoint`, `eventWindow`, `eventSince`); after an erase the handler's kept window credentials are not served (`servesEphemeral`, and `TestEndToEnd_EraseBringsAFreshNode` green).
- **`NodeSwitch`:** every `NodeSource` method delegates to the current target; `Set` and `swapped` are ordered by one mutex. `provider.go` is the `ExposureProvider` port and its types — no wiring; the `Mount`/`NewHandler`/`Handler` wiring the Host check hangs on is `options.go:354-404` (`hostCheck` wraps the whole mux in `NewHandler`).
- **Generated strict server (`internal/api`):** its three error hooks are all set by `controllerRoutes` (`handlers_controller.go:139-147`) — parameter errors 400 problem, body errors `badBody`, handler errors `controllerError`; the generator's `http.Error` defaults (`api.gen.go:3144`, `:6898`) are not reachable; binding and decoding run behind `controllerGate`.
- **Web:** a node key from the hash reaches a request only when the node list holds it, is checked against `NODE_KEY_PATTERN` in `forNode` and escaped in the path (`App.svelte:117-118`, `client.ts:521-534`); a controller node id is matched and normalised by `routeOf`; `tab` selects from a fixed list. `renderQrSvg` builds its markup from numbers only (module coordinates, sizes) — the payload never enters the string. `parseJson` keeps `__proto__` an own property, its token expressions are linear. `index.html`'s pre-paint script reads a same-origin parent's computed style and sets custom properties one by one (`setProperty`), no markup, no URL. The node store drops the codes on success, close, expiry and a status that shows the window gone; `Window.svelte` drops them on expiry, close and unmount, and the page is keyed by node id (`App.svelte:340`, `Node.svelte` `{#key}`), so one node's codes are not shown under another.

Not re-reported (known): UR-P3-bound-197 (measured above); `Hooks.Wrap` and `Deleted` (UR-P4-adpt-31); the 32-bit conversions (UR-P4-ctr-30); the ACL lock-out question (UR-P4-ui-66); `listen`, `discriminator`, `source_note`, `health_note`, `exposure_reason` and the counters open to a non-admin (decisions 34, 63; `phase-2-X3.md` class A-b).

## Coverage, honestly

- **Whole:** the files and ranges of the list above.
- **Partly:** `internal/api/api.gen.go` (8040 generated lines) — the server wrapper's and the strict handler's error paths by pattern (grep over every `ErrorHandlerFunc`, `http.Error`, `WriteHeader`) and two operations read in full (`RemoveCredential`, `CommissionNode`), not every operation. `ui/src/shell/App.svelte` — `:1-60` and `:215-357` here, `:60-215` and the delta by S; `:117-118` and `:152` by grep. `fabricui/handlers_pairing.go` — S's file; I read `:40-135` and `:286-326` for UR-P4-auth-90. `fabricui/gofabric/gofabric.go` — S's file; I read `:45-105`, `:262-300`, `:417-456`, `:772-807` for the credential and error paths.
- **go-fabric at v0.23.0:** `node/commissioning.go:640-760`, `node/config.go:505-575`, `thread/dataset.go:66-84`, and greps over `controller/` and `thread/` for errors that format credentials; not a full read of the facade (unit S's adapter section stands).
- **Not done:** the expiry half of UR-P4-auth-90 was not run (go-fabric's doc and phase 2's observation); the page side of it is read, not run in a browser; no Playwright run; the adapter's own error texts under a failing real device were read at their sources, not provoked.

## Tokens

Not reported by the harness to this agent; estimate ~260 k.
