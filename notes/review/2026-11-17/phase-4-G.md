# Review phase 4, run 1 — unit G (Go rest)

Reviewer G (Opus), HEAD read `979a67a`, go-fabric pin `v0.23.0` (`gf:` = the
`git archive` export). Not a held candidate: no finding of this unit is a
security finding. Red tests in the scratch worktree
`<scratchpad>/p4wt/G` (detached at `979a67a`), files `p4g_*_test.go`; the
worktree is left in place.

Counts: high 0, medium 2, low 1, note 2.

## UR-P4-adpt-31 — medium — `Hooks.Wrap` drops the host's `Deleted` hook (v0.23.0 delta)

- **Where:** `fabricui/gofabric/hooks.go:228-244` (`chain`); `gf:controller/client/config.go:161-170` (`Hooks.Deleted`, new in v0.23.0; absent from `go-fabric@v0.22.2/controller/client/config.go`).
- **Claim:** `chain` copies thirteen fields of `client.Hooks` by name; v0.23.0 added a fourteenth, `Deleted`, and the lift did not add it. A host that wires its own hooks through `Hooks.Wrap(next)` — the documented path for a host with hooks of its own (`hooks.go:23-26`, `:219-221`: "each hook of next(id) runs, then the stream's") — silently loses its `Deleted` callback: go-fabric's registry calls `n.cfg.Hooks.Deleted` (`gf:controller/client/registry.go:466`) and finds nil. Whatever the host does on a node's deletion (forget, erase, cull, failed `CompleteCommissioning`, decommission) never runs. Nothing guards the field list, so the next go-fabric hook goes the same way.
- **M/I:** measured (red test); the host impact is inferred (precondition: a host that sets `Deleted` and uses `Wrap`).
- **Test:** `fabricui/gofabric/p4g_hooks_test.go` (`TestP4GWrapKeepsEveryHostHook`, reflection over every func field of `client.Hooks`):
  ```
  --- FAIL: TestP4GWrapKeepsEveryHostHook (0.00s)
      p4g_hooks_test.go:27: Wrap drops the host's Deleted hook
  ```
- **Fix:** `Deleted: both0(a.Deleted, b.Deleted)` in `chain`; keep the reflection test as the guard so a future hook field fails the build's tests. (Optional, separate: whether `For` should also forget `ids`/`pending` on `Deleted`, which covers the removals `Decommissioned` does not.)
- **Contract:** no.

## UR-P4-adpt-32 — medium — one product's failed update check answers the whole check 502, undeclared, and drops the updates found

- **Where:** `fabricui/gofabric/updates.go:45-49` (`Check`: `if err != nil { return nil, nodeError(err) }`); `fabricui/handlers_controller.go:1702-1709` (`CheckUpdates` returns on any error); `contract/openapi.yaml` `POST /controller/updates/check` declares 200 403 500 503.
- **Claim:** `softwareupdate.Manager.QueryUpdates` checks product by product, passes over a product whose check failed, and returns the updates of the others **together with** the joined error; go-fabric's doc names exactly this misuse: "a caller that treats any error as fatal discards the updates found besides" (`gf:controller/softwareupdate/manager.go:938-947`, `:1015-1019`). The adapter does that: one vendor's DCL or ledger lookup failing hides every available update of every other node, and the error goes through `nodeError` to `ErrNodeFailed`, so the route answers **502 node_error** — a status the operation does not declare (decision 47). The status guard did not see it because `MemoryController`'s `Check` never fails and the adapter package's tests are not observed (UR-P4-test-33).
- **M/I:** measured at the handler with a port whose `Check` fails as the adapter's does (the 502 and the guard); the partial-result loss measured by reading the adapter against `gf:` (inferred end to end — go-fabric's `query` hook is unexported, so no adapter-level red test without a DCL double).
- **Test:** `fabricui/p4g_updates_test.go`:
  ```
  --- FAIL: TestP4GUpdateCheckWithOneFailedProduct (0.00s)
      p4g_updates_test.go:37: check with one failed product: 502 {... "code":"node_error"}
  ```
  and `TestP4GUpdateCheckStatusGuard` (passes on today's 502) then fails the run through `TestMain`:
  ```
  PASS
  status guard: 1 answered status(es) the contract does not declare for the operation:
    - POST /controller/updates/check answered 502 (/api/v1/controller/updates/check)
  FAIL	github.com/SukramJ/go-fabric-ui/fabricui
  ```
- **Fix:** the port's `Check` returns the updates found and the error; the adapter returns `found, err` (logged per product by go-fabric already); the handler answers 200 with the list when it is non-empty or the error is partial, and an additive field (e.g. `check_error` / `partial: true`) says a product's check failed; a total failure stays an error with a declared status (declare 502 on the operation, or map it to 500). Either way the contract changes.
- **Contract:** yes (declared status and/or an additive field; 0.11.0, decision 67).

## UR-P4-ctr-30 — low — 32-bit hosts: uint32 values become negative numbers, and a consent for a version ≥ 2^31 cannot be sent

- **Where:** generated `int` fields (`internal/api/api.gen.go:806`, `:859`, `:1145`, `:1149`, `:1477`, `:1746`, …) filled by `int(uint32)` in `fabricui/handlers_controller.go:546` (cluster `id`), `:551` (`data_version`), `:557` (attribute `id`), `:1654` (`target_software_version`), `:1715` (`software_version`), `nonZero`/`optUint` (`:331-343`: `device_type`, `pairing_hint`, `specification_version`), `fabricui/handlers_node.go:500` (event `number`); the request side `UpdateConsentRequest.TargetSoftwareVersion int`. The pipeline builds for 386 "the armv7 add-on and arm/386 hosts of the consumers" (`.github/workflows/ci.yml:127-128`), and `ConsentUpdate` already guards the 32-bit `int` (`handlers_controller.go:1689`, comment "int is 32 bits on arm/386").
- **Claim:** on a 32-bit build every uint32 at or above 2^31 is written as a negative JSON number: a cluster's `data_version` (a random 32-bit start value per the specification, so about half of all clusters), manufacturer-specific cluster and attribute ids (`0xFFF1FC00` → `-918528`), device types, software versions. The hex twins keep the ids exact, but `data_version`, `software_version` and `target_software_version` have no twin. On the request side, a `target_software_version` ≥ 2^31 does not decode into the 32-bit `int` (400), so the update the list offers cannot be consented to.
- **M/I:** measured under `GOARCH=386` (the tests pass on amd64).
- **Test:** `fabricui/p4g_int32_test.go`, `GOARCH=386 go test ./fabricui -run TestP4G`:
  ```
  --- FAIL: TestP4GUint32SurvivesA32BitInt (0.00s)
      p4g_int32_test.go:19: data_version: {...,"data_version":-559038737,...,"id":-918528,"id_hex":"0xFFF1FC00",...}
  --- FAIL: TestP4GConsentHighVersionDecodes (0.00s)
      p4g_int32_test.go:33: decode: json: cannot unmarshal number 3000000000 into Go struct field UpdateConsentRequest.target_software_version of type int
  ```
- **Fix:** `format: int64` on the contract's uint32-ranged integers (or `x-go-type: int64`), `make generate`, and `int64(...)` at the conversion sites; a `GOARCH=386` run of these tests in CI (the 386 job only builds today). The wire shape is unchanged.
- **Contract:** no wire change; the contract file gains `format` annotations (generator input).

## UR-P4-test-33 — note — the status guard (decision 47) sees only `MemoryController`'s errors

- **Where:** `fabricui/status_guard_test.go:28-46` (`TestMain` of package `fabricui` only); `fabricui/handler.go:436-451` (`observeStatus` unexported, unset in `fabricui/gofabric`); the adapter-only error paths, e.g. `fabricui/gofabric/updates.go:47-49` (UR-P4-adpt-32), `fabricui/gofabric/gofabric.go` `SetFabricLabel` "the controller has no fabric" → `ErrNodeFailed` → 502 on `PUT /controller/fabric/label` (declares 204 400 403 413 415 500).
- **Claim:** the guard observes the answers of the `fabricui` package's tests, whose controller is `MemoryController`; the real adapter's mappings (`nodeError`, `commissionError`) are exercised by `fabricui/gofabric`'s tests (`adapter_paths_e2e_test.go`, `e2e_test.go`) through the same handler, but nothing observes their statuses. A status only the adapter produces escapes the guard, as UR-P4-adpt-32 shows.
- **M/I:** measured for updates/check (UR-P4-adpt-32's test); the fabric-label path inferred by reading (precondition rare).
- **Fix:** an exported test hook (e.g. an internal `fabricuitest` package or an `export_test`-style `SetStatusObserver` in an `internal/` package) used by `fabricui/gofabric`'s `TestMain` with the same check; or a table test that drives each port error kind through every controller operation's handler and requires the status declared.
- **Contract:** no (the guard); the statuses it would find are.

## UR-P4-doc-34 — note — `NodeSwitch` names the browser's `EventSource` as what reconnects

- **Both sides:** `fabricui/nodeswitch.go:29-31` "so the browser's EventSource reconnects and its new stream observes the new node"; `ui/src/lib/api/stream.ts:93-98` (decision 45: all three streams through `fetch`) "it does not reconnect by itself (the subscribe functions do)".
- **Fix:** "so the page's stream subscription reconnects".
- **Contract:** no.

## Covered without finding

Delta first, line by line: `fabricui/handlers_controller.go` (the gate's strict body pass, admin context, status for a non-admin, `since_str`, commissioning id, passcode list, label bytes, bindings per-endpoint status, session ids), `fabricui/handlers_node.go` (status reductions, `remainingSeconds`, fabric 404, strict confirm bodies, `eventSince`, `queryString`), `fabricui/handlers_pairing.go`, `fabricui/problem.go`, `fabricui/strict.go` (non-security: `decodeStrict`, `exactNames`, `jsonFields` — no `UnmarshalJSON`/union type in `internal/api` that its field walk would misjudge; `hasBody`, `readBody`, `allowedMethods`, `noRoute`), `fabricui/controller.go`, `fabricui/node_detail.go`, `fabricui/handlers_exposure.go`, `fabricui/handlers_diagnostics.go`, `fabricui/registry.go`, `fabricui/handlers_surface.go`, `fabricui/memory_controller.go`, `fabricui/gofabric/node.go`, `script/coverfloor/main.go`, `script/gen-api.sh`, `internal/version/version.go`; tests holding a guard: `fabricui/status_guard_test.go`, `fabricui/routes_contract_test.go`, `fabricui/fuzz_test.go`, `fabricui/fuzz_checks_test.go` (FuzzHostCheck/FuzzOrigin read for guard shape only, security is unit S's).
Then: `fabricui/gofabric/hooks.go`, `updates.go`, `topology.go` (watch/close), `fabricui/nodeswitch.go`, `fabricui/operations.go` (log window), `fabricui/handler.go` (status writer, mount; security not), `fabricui/events.go` (types and allow-list shape only), `internal/arch/arch_test.go`. Read in part only (no finding, not line by line): `fabricui/gofabric/gofabric.go` (error mapping, `Decommission`, label push, `CommissionDiscovered` head), `fabricui/gofabric/diagnostics.go`, `fabricui/handlers_topology.go`, `fabricui/spa.go`, `fabricui/provider.go`, `fabricui/ports.go`, `fabricui/devicetype_names.go`.

Considered and dropped: the controller routes without a declared 500 answer one only for a client that left (`context.Canceled` passed through `nodeError`/`commissionError`, logged at debug) — nobody reads it; a registry `List` that races a concurrent `Create` can stop a just-built managed handler (`registry.go:140-158`) — window too narrow to show without a host double, not reported.

## Tokens

About 190k (estimate; no counter available to the agent).
