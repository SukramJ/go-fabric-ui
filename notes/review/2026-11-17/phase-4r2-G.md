# Review phase 4, run 2 — unit G (Go rest)

Reviewer G (Opus), HEAD read `792d15b`, go-fabric pin `v0.23.0` (`gf:` = the
`git archive` export). Not a held candidate: no finding of this unit is a
security finding. Red tests in the scratch worktree
`<scratchpad>/p4r2wt/G` (detached at `792d15b`), files `fabricui/p4r2g_*_test.go`, the mutation of UR-P4-test-265 restored after its run;
the worktree is left in place.

Counts: high 0, medium 0, low 2, note 6 (263–267 added by the follow-up read that closed "Not covered").

## Scope (listed before reading)

Hand-written Go outside unit S-go: `fabricui/{controller,devicetype_names,doc,handlers_controller,handlers_diagnostics,handlers_exposure,handlers_node,handlers_pairing,handlers_surface,handlers_topology,memory_controller,node_detail,nodeswitch,operations,ports,provider,registry}.go`; `fabricui/gofabric/{diagnostics,topology,updates}.go`; `internal/{statusguard,statusobs,version}`; `script/coverfloor/main.go`, `script/gen-api.sh`, `script/extract-release-notes.sh`; `contract/openapi.yaml` against the handlers; the Go tests that hold a guard. Delta first: `internal/statusguard/statusguard.go`, `internal/statusobs/statusobs.go`, `fabricui/gofabric/updates.go`, and the `979a67a..792d15b` diff of `handlers_controller.go`, `controller.go`, `registry.go`, `handlers_node.go`, `handlers_pairing.go`, `nodeswitch.go`, `handlers_diagnostics.go`, `internal/version`, `script/coverfloor/floors.txt`, and the contract.

## UR-P4-ctr-260 — low — 32-bit hosts: a discovery filter value ≥ 2^31 is refused with 400 (a wider reach of UR-P4-ctr-30)

- **Where:** `contract/openapi.yaml:1862-1865` (`GET /controller/discovery` parameter `value`: `{ type: integer, minimum: 0 }`, no `x-go-type: int64`); generated `internal/api/api.gen.go:1824` (`Value *int`); `fabricui/handlers_controller.go:1542` (the range check already says "int64: int is 32 bits on arm/386", but binding fails before it).
- **Claim:** UR-P4-ctr-30's fix gave the uint32-ranged response fields and the consent's request field `x-go-type: int64` (`contract/openapi.yaml:22-25`, `:3649`), but not this query parameter. On a 386/arm build a `device_type` filter of a vendor-specific device type (vendor prefix ≥ 0x8000, e.g. the test vendors' `0xFFF1xxxx`) overflows the generated `int` and is answered 400 "Invalid request parameter" before the handler runs. The npm client offers the filter (`ui/src/lib/api/client.ts:150`, `DiscoverFilter.filter "device_type"`), so a host client or a future page cannot discover such devices on the armv7 add-on.
- **M/I:** measured under `GOARCH=386` (passes on amd64).
- **Test:** `fabricui/p4r2g_discovery_test.go`, `GOARCH=386 CGO_ENABLED=0 go test -run TestP4R2G ./fabricui`:
  ```
  --- FAIL: TestP4R2GDiscoveryHighDeviceTypeOn32Bit (0.00s)
      p4r2g_discovery_test.go:17: device_type 0xFFF10000: 400 {... "detail":"Invalid format for parameter value: error binding string parameter: value '4293984256' overflows destination of type: int", ...}
  ```
- **Fix:** `x-go-type: int64` on the parameter's schema (as `target_software_version`), `make generate`, `make gen-types` (no TS change), drop the `int64(...)` cast and its comment at `:1542`; the test joins `int64_test.go` (run by `make test-386`).
- **Contract:** no wire change; the contract file gains the generator annotation.

## UR-P4-err-261 — low — factory reset stops "incomplete" at a fabric removed meanwhile (the factory-reset path of UR-P1-err-33)

- **Where:** `fabricui/handlers_node.go:414-424` (`factoryReset`'s loop: any `removeFabric` error ends the reset with 500); compare `fabricui/handlers_node.go:338-344` (`removeFabric` maps `node.ErrFabricNotFound` to 404, UR-P1-err-33) and go-fabric's own reset, which passes over it: `gf:node/factory_reset.go:78` (`err != nil && !errors.Is(err, store.ErrFabricNotFound)`); `gf:node/facade.go:103-105`, `gf:node/node.go:1719` ("An unknown index is [ErrFabricNotFound]").
- **Claim:** the reset lists the fabrics, then removes them one by one, detached. A controller that removes its own fabric between the list and its turn (a decommission from that controller, or a concurrent `DELETE /node/fabrics/{index}`) makes `RemoveFabric` answer `ErrFabricNotFound`; the loop treats that as a failure, logs `factory_reset.partial`, answers 500 "Factory reset incomplete: removed n of m", and leaves every later fabric in place — the node stays paired to controllers the operator asked to remove, until the reset is run again. The fabric is gone, which is what the reset wants.
- **M/I:** measured with the test node's `removeErr` (status and stop); the later fabrics left in place are read from the loop (I).
- **Test:** `fabricui/p4r2g_reset_test.go`, `go test -run TestP4R2GFactory ./fabricui`:
  ```
  --- FAIL: TestP4R2GFactoryResetSkipsAFabricGoneMeanwhile (0.00s)
      p4r2g_reset_test.go:23: status 500 {..."title":"Factory reset incomplete: removed 0 of 2 fabrics","status":500,...}, want 204
  ```
- **Fix:** in the loop, `errors.Is(err, node.ErrFabricNotFound)` → continue (not counted as removed, not a failure), as go-fabric's `FactoryReset` does; the regression test also checks that a later fabric is removed (a fake that fails one index with `ErrFabricNotFound`).
- **Contract:** no.

## UR-P4-ctr-262 — note — 32-bit hosts: a cluster whose reads all failed lists a manufacturer-specific attribute first

- **Where:** `fabricui/node_detail.go:237` (`failedAttributes`: `slices.SortFunc(..., func(a, b ClusterAttribute) int { return int(a.ID) - int(b.ID) })`).
- **Claim:** `int(uint32)` is negative for an id ≥ 2^31 on 386/arm (and the difference can overflow), so `0xFFF10000` sorts before `0x0000` in `GET /node/endpoints/{ep}` when every attribute of a cluster failed. Order only; the same class as UR-P4-ctr-30, missed by its sweep because it is a comparator, not a field.
- **M/I:** measured under `GOARCH=386`.
- **Test:** `fabricui/p4r2g_sort_test.go`:
  ```
  --- FAIL: TestP4R2GFailedAttributesSortOn32Bit (0.00s)
      p4r2g_sort_test.go:20: order 0xfff10000, 0x0; want 0x0, 0xfff10000
  ```
- **Fix:** `cmp.Compare(a.ID, b.ID)`.
- **Contract:** no.

## UR-P4-doc-263 — note — the comma-separated Multi form is "accepted until v0.2.0", and still accepted at v0.11.0

- **Both sides:** `fabricui/provider.go:146-147` ("Deprecated form: the comma-separated Fields[key] for a Multi column is accepted until v0.2.0") and `docs/provider.md:41` ("Deprecated, accepted until v0.2.0"); against `internal/version/version.go:11` (`0.11.0`), `Item.ListValues` (`fabricui/provider.go:148-165`, still splitting `Fields[key]`) and the page's split (`ui/src/lib/pages/node/fields.ts:21`).
- **Claim:** the removal date passed nine minors ago and the form is still served; a host reading the godoc either thinks it is gone or that its removal is overdue in the next release. A route-free form, so not a contract question, but the godoc is the hosts' reference.
- **Fix:** name the real plan — "accepted until the next major (v1.0.0)" — or remove the form in a minor with a changelog line (then also `fields.ts`); the owner's choice.
- **Contract:** no.

## UR-P4-doc-264 — note — `NetworkNode.Kind`'s godoc omits `border_router`

- **Both sides:** `fabricui/controller.go:185-187` ("one of the contract's TopologyNode kinds: node, peer, thread_router, thread_end_device, wifi_ap, ethernet"); `contract/openapi.yaml:2476` (the enum has `border_router`, 0.8.0, described at `:2019`) and the adapter's `borderRouterOf` (`fabricui/gofabric/topology.go:248`, `Kind: "border_router"`).
- **Claim:** a host writing its own `ControllerSource` learns the kinds from this godoc and does not know the border-router node it may emit.
- **Fix:** add `border_router` (and the sentence of `:2019`: labelled by the mDNS host name, `ok` while seen within 24 h).
- **Contract:** no.

## UR-P4-test-265 — note — the skipped-entry log guard passes when nothing is logged

- **Where:** `fabricui/log_once_test.go:81` (`TestRegistry_ASkippedEntryIsLoggedOnce`: `if got > 1`); the guarded code `fabricui/registry.go:375-390` (`logSkipped`). No other test names `fabricui.registry.entry_skipped`.
- **Claim:** the test is named "logged once" but asserts "at most once": a `logSkipped` that never logs passes, and so does one that forgets the "logged again when it returns" rule. The operator then never learns that the host's registry lists an entry the handler drops. (`TestStatus_AFabricReadFailureIsLoggedOnce` has the same `> 1` at `:35`, but its later `!= 2` and `!= 1` counts close the gap there.)
- **M/I:** measured by mutation in the scratch worktree: the `h.log.Warn(...)` loop of `logSkipped` replaced by `_ = fresh`; `go test -run 'Registry|LogOnce|Skipped|Nodes' ./fabricui` → `ok` (file restored afterwards).
- **Fix:** require exactly 1, then drop the entry from the registry's list, list once, add it back, list again, and require 2.
- **Contract:** no.

## UR-P4-test-266 — note — no guard holds that every served route is in the contract

- **Where:** `internal/statusguard/statusguard.go:123-125` (an answer whose path no operation matches is skipped: `if best == nil { continue }`); `fabricui/routes_contract_test.go` (`TestEveryPathAnswers405WithItsMethods` and `fabricui/host_test.go:243` `TestAllowHost_EveryModeAndRoute` walk the contract's operations, not the handler's); `fabricui/handlers_nodes_test.go:167` and `fabricui/handlers_controller_test.go:195` compare the **admin** routes only; the route tables `fabricui/handler.go:361` (`nodeRoutes`), `:390` (`registryRoutes`), `:397` (`operationAdmin`).
- **Claim:** a non-admin hand-written route added to `nodeRoutes` without its contract operation passes every guard: the status guard ignores its answers, and the contract-side walks never reach it. The controller family cannot drift (generated), the node and operation families can. Holds today: the scratch test below passes.
- **M/I:** inferred by reading the guards; today's consistency measured.
- **Test (the proposed guard, green today):** `fabricui/p4r2g_routes_test.go` (`TestP4R2GEveryHandWrittenRouteIsInTheContract`, every `nodeRoutes` entry under `/node` and `/nodes/{node}` and every `registryRoutes` entry is an operation of the contract); extend it to the non-admin operation routes, or make the status guard fail on an answered path no operation matches (other than the path-level 404/405 it already excludes).
- **Contract:** no.

## UR-P4-api-267 — note — the port's error godocs promise statuses the contract declares on few routes

- **Both sides:** `fabricui/controller.go:942` ("ErrNotImplemented: the source does not serve the route yet (501)"), `:948` ("ErrNotRead … (503)"), mapped for every controller route by `fabricui/handlers_controller.go:211-245` (`controllerError`); against the contract, which declares 501 only on `getControllerTopology`, `getControllerDiagnostics`, `getControllerMdns`, `listControllerDiagnosticEvents` and 503 on the controller node routes only on `getNodeCluster` (e.g. `pingNode` declares 200 400 403 404).
- **Claim:** the adapter and `MemoryController` return these errors only where declared, so the status guard is green; a host's own `ControllerSource` that follows the godoc ("does not serve the route yet") on any other method answers a status the contract does not declare for it (decision 47), and no guard can see a host's source.
- **M/I:** inferred by reading.
- **Fix:** say per error where it is answered ("from NetworkTopology, Diagnostics…: 501"; "from ClusterDetail: 503") and that elsewhere it is a 500; or map them to 500 outside the operations that declare them.
- **Contract:** no (docs), or the declared statuses if the owner prefers to declare 501 everywhere.

## Covered without finding

Delta, line by line: `internal/statusguard/statusguard.go` (recorder, best-literal match, HEAD→GET, `ScanOperations` indentation rules, `-short` and failed-run skip; the coverage floor line), `internal/statusobs/statusobs.go`, `fabricui/gofabric/updates.go` (seams, `Check`'s per-product re-check against `gf:controller/softwareupdate/manager.go:948-1003` — grouping, `only`, suppressed run, unknown node; `Consent`'s 500; `updateEvent`), the delta hunks of `fabricui/handlers_controller.go` (`optUint64`/`nonZero64`, `RemoveNodeFabric`'s `nocInvalidFabricIndex` = NodeOperationalCertStatusEnum 11, `CommissionNode`'s 600 s default on the discovered path, `CheckUpdates`' partial/total split and 500, the int64 conversions), `fabricui/controller.go` (`UpdateCheckFailures`), `fabricui/handlers_node.go` (`window_kind` mapping, `removeFabric` 404, `logFabricsState`), `fabricui/handlers_pairing.go` (`window_match` under `administrator_basic`), `fabricui/registry.go` (`logSkipped`), `fabricui/nodeswitch.go`, `fabricui/handlers_diagnostics.go`, `internal/version/version.go`, `script/coverfloor/floors.txt`; the contract's 0.11.0 diff against the handlers (statuses of `checkUpdates`, `removeNodeFabric`, `closeWindow`, `commissionNode.timeout_seconds`, `window_kind` enum).
Then: `fabricui/handlers_controller.go` in full (error mapping per operation checked against the declared statuses with the adapter's `nodeError`/`commissionError` mappings — only `context.Canceled` reaches an undeclared 500, already dropped by run 1), `fabricui/handlers_node.go` from line 100, `fabricui/handlers_pairing.go`, `fabricui/node_detail.go`, `fabricui/handlers_exposure.go`, `fabricui/handlers_surface.go`, `fabricui/handlers_topology.go`, `fabricui/handlers_diagnostics.go`, `fabricui/registry.go`, `fabricui/operations.go`, `fabricui/nodeswitch.go` (method bodies), `fabricui/gofabric/diagnostics.go`, `script/coverfloor/main.go`. Tests holding a guard read: `fabricui/status_guard_test.go`, `fabricui/gofabric/main_test.go`, `fabricui/int64_test.go`, `fabricui/commission_bound_test.go`, `fabricui/fabric_not_found_test.go`.

Follow-up read (the former "Not covered", line by line): `fabricui/ports.go`, `fabricui/provider.go` (finding 263), `fabricui/devicetype_names.go`, `fabricui/doc.go`, `script/gen-api.sh`, `script/extract-release-notes.sh` (exact `## [X.Y.Z]` match, previous-version link), `fabricui/handlers_node.go` 1-99, `fabricui/controller.go` in full (findings 264, 267), `fabricui/memory_controller.go` second half (Diagnostics, mDNS, events, updates, the memory nodes' Read/Write/Invoke incl. NOCResponse InvalidFabricIndex, ICD), `fabricui/gofabric/topology.go` tail (`linkOf`, `mapID`, `details`). Guard tests, each judged for bite: `fabricui/sibling_test.go` (fills every shared field by reflection and fatals on one it cannot fill — bites), `fabricui/updates_check_test.go` (partial/whole/no-failure statuses — bites), `fabricui/gofabric/updates_test.go` (holds `Check`'s algorithm through its seams; the real `newUpdates` closures — the product key and `QueryUpdates(nil|node)` — are not under test, which is the carried UR-P4-adpt-197's ground), `fabricui/host_proxy_test.go` (three proxy kinds measured, four docs sentences — bites), `fabricui/log_once_test.go` (finding 265), `fabricui/window_close_e2e_test.go`, `fabricui/gofabric/hooks_test.go` (`Wrap` by reflection over every `client.Hooks` func field — bites; a future hook with a result would panic the test, which fails loudly), `fabricui/gofabric/failsafe_e2e_test.go` (fatals when the cut came too late — self-checking), `fabricui/gofabric/window_kind_e2e_test.go`, `fabricui/gofabric/adapter_paths_test.go` (each kind exclusive of the others), `fabricui/routes_contract_test.go` (finding 266), `fabricui/fuzz_test.go`, `fabricui/fuzz_checks_test.go`, `fabricui/gofabric/fuzz_test.go` (oracles beyond "no panic": round-trips, 400/413 problem shape, no secret in the log, no admin answer without rights; `FuzzDecodeStrict` checks only agreement on accepted input, the refusals are `TestBodies_AreStrict`'s), `internal/arch/arch_test.go` (classifier table plus a walk that fatals under ten files — bites).

Considered and dropped: `openWindow` answers a stopped or not-started node (`node.ErrStopped`, `ErrNotStarted`) 500 rather than 503 (declared either way; no host path shown); `Check` re-checks one node of every product without an update, also of products whose check succeeded (extra DCL lookups and `check_finished` events, documented in its comment); `MemoryOperations.Record` can hand two concurrent records to observers out of sequence (test double; the page reads `seq`); a registry entry whose `Source` is not comparable rebuilds its handler per request and so re-logs `fabrics_failed` per request (documented trade-off of `sameValue`).

## Not covered

None of the unit's files. Out of the unit and not read: the remaining Go tests that hold no guard named by the brief (handler tests in the existing patterns, `support_test.go` beyond the fakes used above), and the memory nodes' static spec data (`memorySensor`, `memoryPlug`) beyond their shape.

## Tokens

About 150k for the first pass and about 75k for the follow-up, about 225k in all (estimate; no counter available to the agent).
