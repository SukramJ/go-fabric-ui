# Phase 1a — G-go-A (Go handlers outside the security paths)

HEAD `49cc081`, go-fabric `v0.22.2` (pinned export). Proving tests were run in a
scratch copy of the repo (`fabricui/zz_ggoa_test.go`, not in the repo); the code
is quoted under the findings that use it.

Gate cross-check: `controllerAdmin` (`handlers_controller.go:51-74`) holds 22
routes, the contract has 22 `x-admin` operations under `/controller`
(`openapi.yaml:1131-1750`), plus the `refresh` query mark (`:1789`, `:79-81`).
They match (M). No route in my slices reaches an admin action without the gate.

## Findings

### UR-P1-adpt-30 — GET /controller/events serves the trace oldest first

- **Severity:** medium
- **Where:** `fabricui/handlers_controller.go:1849` (the reversal); cause `fabricui/gofabric/diagnostics.go:85-94`; go-fabric `diagevent/ring.go:210-231`, `controller/diagnostics.go:140-142`
- **Claim:** The port says `DiagnosticEvents` is oldest first (`controller.go:71-73`), so the handler reverses it. go-fabric's `Controller.DiagnosticEvents` is `Ring.Snapshot`, which is **newest first** (`ring.go:210`, index `next-1-i`). The adapter passes it through unchanged and calls it "oldest first". The route therefore serves the oldest event first, against the contract (`openapi.yaml:1846-1849`, "newest first"). The page renders it in that order (`ui/src/lib/pages/controller/Diagnostics.svelte:60`). The in-memory controller is oldest first, so the handler tests pass. The e2e test checks only membership (`gofabric/e2e_test.go:311-319`).
- **M/I:** M (read the ring, the adapter and the handler)
- **Fix:** reverse in the adapter so the port's promise holds (or change the port doc to newest first and drop the reversal). Add an e2e assertion on order.
- **Contract change:** no

### UR-P1-err-31 — an undecodable read turns into "no ACL", "no fabrics" and fabric index 0; the own-fabric guard falls

- **Severity:** medium (precondition: a node whose value go-fabric's definitions cannot decode)
- **Where:** `handlers_controller.go:1172-1176` (`ownIndex`), `:1219`, `:1192-1201`, `:879-883`, `:1016-1019`; cause `gofabric/node.go:462-465` (a decode failure becomes the value `null`)
- **Claim:** The adapter renders a value it cannot decode as `null`. `json.Unmarshal("null", &x)` leaves `x` alone and returns no error. So:
  - `ownIndex` answers 0, and `RemoveNodeFabric`'s "this controller's fabric" guard (`:1219`) never matches. Our own fabric is then removed with RemoveFabric.
  - `ListNodeFabrics` answers `[]` and marks no fabric as `own`, so the page offers our fabric for removal.
  - `GetNodeACL` answers 200 `{"entries":[]}`, not the 502 the code intends (`:880-881`). A following `PUT …/acl` replaces the node's real list.
  - `GetNodeBindings` treats it the same way.
- **M/I:** M. Test (scratch): a node wrapper that returns `null` for CurrentFabricIndex and ACL gives `DELETE …/fabrics/1` (our index) **204**, and `GET …/acl` gives **200 `{"entries":[]}`**:
  ```go
  func (n nullNode) Read(ctx context.Context, paths []AttributePath, f bool) ([]AttributeValue, error) {
  	vals, err := n.ControllerNode.Read(ctx, paths, f)
  	for i := range vals { p := vals[i].Path
  		if (p.Cluster == clusterOperationalCredentials && p.Attribute == attributeCurrentFabricIndex) ||
  			(p.Cluster == clusterAccessControl && p.Attribute == attributeACL) { vals[i].Value = json.RawMessage("null") } }
  	return vals, err
  }
  // DELETE /controller/nodes/0000000000000011/fabrics/1 -> 204 (want 400); GET …/acl -> 200 {"entries":[]} (want 502)
  ```
- **Fix:** in `readOne`, treat a `null` value as `ErrNodeFailed` (none of ACL, Fabrics, Binding or CurrentFabricIndex is nullable). In `ownIndex`, refuse 0. In the adapter (G-go-B), answer a decode failure as an error or a status, not `null`.
- **Contract change:** no

### UR-P1-ctr-32 — GET …/acl answers `privilege: ""` / `auth_mode: ""` for a value outside the enum

- **Severity:** low
- **Where:** `handlers_controller.go:885-887`
- **Claim:** `privilegeNames[e.Privilege]` and `authModeNames[e.AuthMode]` answer the zero string for a value the maps do not hold (0, a future value). That violates the contract's enums (`openapi.yaml:3113-3114`), and the TS client's union type does not cover it.
- **M/I:** M. Scratch test: write `[{"privilege":6,"authMode":4,…}]` to the memory node, then GET answers `{"auth_mode":"","privilege":"",…}`.
- **Fix:** answer 502 for an entry outside the enums, or add `unknown` to both enums (contract).
- **Contract change:** no for the 502 fix; yes for `unknown`
- **Red test:** a memory node's ACL with privilege 6 → `GET …/acl` is not 200 with `"privilege":""`.

### UR-P1-err-33 — an unknown fabric index is 500 on the node route and 502 on the controller route

- **Severity:** low
- **Where:** `handlers_node.go:280-282`; `handlers_controller.go:1230-1235`
- **Claim:**
  - Node route: `Node.RemoveFabric` answers `store.ErrFabricNotFound` for an unknown index (go-fabric `node/node.go:1146-1149`). The facade does not export that sentinel, so the handler logs a server error and answers 500 "Failed to remove fabric". This happens on an ordinary race: another admin, or the controller itself, removed the fabric first.
  - Controller route: the node answers NOCResponse `InvalidFabricIndex` (11), and the handler answers 502 "The node failed". The contract lists 404 for that route (`openapi.yaml:1590`). The same error class gets three different statuses.
- **M/I:** M. Scratch test: `DELETE /controller/nodes/0000000000000011/fabrics/7` → 502 `RemoveFabric answered NOCResponse status 11`. The node route is M by reading.
- **Fix:** map NOCResponse 11 to 404 `not_found`. Ask go-fabric (wave CK) to export `node.ErrFabricNotFound`, then map it to 404 and add 404 to `removeFabric`/`removeFabricForNode`.
- **Contract change:** yes (node routes gain 404)
- **Red test:** memory controller `DELETE …/fabrics/7` expects 404.

### UR-P1-adpt-34 — SetNodeACL drops a Group subject that equals the node id

- **Severity:** low
- **Where:** `handlers_controller.go:939-944`
- **Claim:** "A subject naming the node itself is left out" runs for every auth mode. A Group entry's subject is a 16-bit group id. On a node with a small node id (common for nodes migrated from a sequential allocator), group N on node N is silently dropped. If it was the entry's only subject, the whole entry goes, and the write still answers 200 `{"status":0}`.
- **M/I:** M. Scratch test: a PUT with a CASE admin entry plus `{"auth_mode":"group","subjects_hex":["0000000000000011"]}` on node 0x11 → the GET shows the group entry is gone. How often small node ids occur is I.
- **Fix:** apply the self-filter only when `mode == 2` (CASE).
- **Contract change:** no
- **Red test:** the PUT above keeps the group entry.

### UR-P1-adpt-35 — `GET /node/endpoints/65535/events` returns every endpoint's events

- **Severity:** low
- **Where:** `handlers_node.go:455-461` (`pathEndpoint` accepts 0xFFFF), `:579`; go-fabric `im/eventlog.go:441-469` (0xFFFF is the endpoint wildcard), `node/served.go:176-185`
- **Claim:** The route promises the events of one endpoint. Endpoint 0xFFFF is not a valid endpoint, and go-fabric reads it as "any endpoint", so the route answers the whole log. The contract's `EndpointNumber` allows 65535 (`openapi.yaml:1905-1910`).
- **M/I:** M (read). No harm beyond the wrong scope: the log carries no access control either way.
- **Fix:** refuse 65535 with 400 in `pathEndpoint`, and/or set `maximum: 65534` in the contract.
- **Contract change:** yes if the maximum changes
- **Red test:** a fake event log with events on endpoints 1 and 2; `/node/endpoints/65535/events` answers 400 or an empty list.

### UR-P1-err-36 — /node/compatibility swallows the fabric store error

- **Severity:** low
- **Where:** `handlers_diagnostics.go:342-352`
- **Claim:** When `FabricInfos` fails, the report is 200 with `ecosystems: []` and findings computed as if no controller were commissioned. `/node/fabrics` and `/node/topology` answer 500 for the same failure (`handlers_node.go:243-246`, `handlers_topology.go:100-103`).
- **M/I:** M. Scratch test with `fabricsErr` set: compatibility 200 `{"ecosystems":[],"endpoint_count":3,"findings":[]}`, topology 500, fabrics 500.
- **Fix:** answer `serverError` as the sibling routes do, and add 500 to `getCompatibility`.
- **Contract change:** yes (a 500 response listed)
- **Red test:** `fabricsErr` set → `GET /node/compatibility` is 500.

### UR-P1-bound-37 — GET …/bindings reads every Binding endpoint in one unbounded read, and a refused endpoint vanishes

- **Severity:** low
- **Where:** `handlers_controller.go:996-1015`; also `:710` with `maxReadPaths = 32`
- **Claim:**
  - One `Read` gets as many paths as the image has Binding endpoints, with no cap. The adapter does not chunk (`gofabric/node.go:448-456`). A node that refuses a large path list fails the whole route with 502. The spec's guaranteed minimum is fewer than 32 paths (I, not checked against the spec text here).
  - A path the node answers with a status is skipped (`:1013-1015`), so the endpoint is missing from the answer rather than marked.
- **M/I:** M for the code; the device limit is I
- **Fix:** chunk the reads (or read per endpoint), and carry a per-endpoint `status` (contract) or answer 502 for a refused endpoint.
- **Contract change:** yes if a status field is added
- **Red test:** a memory node whose Binding read answers 0x7E on one endpoint → the endpoint is present with its status.

### UR-P1-ui-38 — one endpoint, two device type names

- **Severity:** note
- **Where:** `devicetype_names.go:12-36`, `:43-54`; `node_detail.go:137-143`, `:300-307`; `handlers_diagnostics.go:253-258`
- **Claim:** `GET /node/endpoints` and the exposure verdicts name 0x0100 "On/Off Light" (operator labels). `GET /node/endpoints/{ep}` and `structural[].device_types` name the same type "OnOffLight" (`schema.DeviceTypeName`). The same page shows both.
- **M/I:** M (`go run` on the pin: `schema.DeviceTypeName(0x0100)` = "OnOffLight", 0x0076 = "SmokeCoAlarm")
- **Fix:** one naming function for every node-family field, or document that `device_types[].name` is the schema name.
- **Contract change:** no
- **Red test:** the endpoint list's `device_type_name` equals the detail's `device_types[0].name` for 0x0100.

### UR-P1-conc-39 — the exposure apply reassembles on the request context

- **Severity:** note
- **Where:** `handlers_exposure.go:212-217`
- **Claim:** The provider has already persisted the change. `h.reassemble(r.Context())` then runs, so a client that disconnects (or a proxy timeout) can cancel the reassembly. That failure is logged only (by design, `TestExposureApply_ReassembleFailureKeepsTheChange`), and the endpoints stay stale until the next reassemble. Whether the host's reassemble honours cancellation is up to the host.
- **M/I:** I
- **Fix:** `h.reassemble(context.WithoutCancel(r.Context()))`, perhaps with its own timeout. Same for `maintenanceReassemble`.
- **Contract change:** no
- **Red test:** a reassemble stub that fails on `ctx.Err()`; cancel the request after `Apply` → reassemble still runs.

### UR-P1-adpt-40 — the fabric label limit counts runes

- **Severity:** note
- **Where:** `handlers_controller.go:1477`; go-fabric `controller/commissioner/fabric.go:98-124`
- **Claim:** 32 runes is matter.js's JS-string length. The FabricDescriptorStruct `Label` constraint `max 32` (`cluster/spec/operationalcredentials/definition_gen.go:126`) is, as I read the data model, a byte length. A 20-character label with umlauts passes here and in go-fabric, then a node may refuse UpdateFabricLabel. The push is in the background, so the refusal appears only in the log.
- **M/I:** I (the byte semantics are not checked against the spec text in this unit)
- **Fix:** upstream question (go-fabric wave CK); then `len(label) > 32` here.
- **Contract change:** no
- **Red test:** a label of 17 "ü" → 400.

### UR-P1-doc-41 — CLAUDE.md lists a narrower facade than the guard enforces

- **Severity:** note
- **Where:** `CLAUDE.md` Rules ("node, eligibility, contract"); `internal/arch/arch_test.go:30-42`, `:151`
- **Claim:** `node_detail.go`, `devicetype_names.go` (`schema`) and `handlers_node.go`/`handlers_surface.go` (`surface`) import packages that the guard allows (`schema`, `surface`, `clock`) but the rule text does not name.
- **M/I:** M
- **Fix:** list the five packages in CLAUDE.md as the guard's message does.
- **Contract change:** no
- **Red test:** none (doc); optionally a test that greps CLAUDE.md for every `facade` key.

## Read and found nothing

- `handlers_exposure.go`: `itemsOf`, `listsOf` (nil→[]), `exposureSchema` (nil lists filled), `readUpdates` (empty key 400; body bounded by `decodeJSON`), `providerError` mapping, `exposurePreview` (empty batch), `exposureApply` 204/200. The Verdict zero value is `unmappable` (go-fabric `contract/matter.go:916`), inside the enum.
- `handlers_surface.go`: `surfaceReport`.
- `provider.go`: interface, `ListValues`.
- `devicetype_names.go`: `deviceTypeName` (apart from 38).
- `handlers_diagnostics.go`: `sessionOf`, `sessionParametersOf`, `sessions`, `mdns` (adapter fills every list, `ports.go:118-155`), `reannounce`, `endpoints`, `structuralEndpoints`, `flattenTopology`, `clustersOf`, `diagnosticEvents` (node ring newest first, as the contract says).
- `handlers_topology.go`: `topology`, `buildNodeGraph` (stable order; a PASE session appears once; ids unique), `nodeState`, `countDevices`, `deviceNode`, `fabricNode`, `sessionNode`.
- `node_detail.go`: `readEndpointDetail` (0x7F → 404, race → 404), `treeParent`, `childrenOf`, `deviceTypesOf`, `findDevice`, `servedCluster`, `failedAttributes`, `named`, `tagsOf`, `labelsOf`, `valueOf`, `eventRecordOf`.
- `handlers_node.go` slices: `statusResponse`/`summariseSurface`, `fabrics`, `removeFabric` index parsing (1..254), `hex64`, `eventWindow`, `eventRecords`, `queryInt`, `nodeEndpoint`, `surfaceKeyOf`, `decisionsOf`, `nodeEndpointEvents` (apart from 33, 35).
- `handlers_controller.go` slices: `hex16`, `hexID`, `parseHex` (width checked before `ParseUint`, bit size passed), `parseHex32`, `parseNodeID`, `endpointOf`, `timedOf` (0..65535), `node`, the opt helpers, `ListControllerNodes`, `nodeSummary`, `nodeDetail`, `endpointDetail`, `clusterDetail`, `elements`, `GetNodeEndpoint`, `GetNodeCluster`, `ListNodeEvents`, `DecommissionNode`, `PingNode`/`pingResult`, `Interview`/`Enable`/`Disable`/`SetNodeAutoSubscribe`, `attributePathOf`, `ReadNodeAttributes` (≤ 32 paths, filter default true), `WriteNodeAttribute` (`null` value allowed, absent refused), `InvokeNodeCommand`, `readOne`, the ACL/Binding field names against go-fabric's camelCase rendering (`cluster/spec/json.go:682`; `accesscontrol`, `binding`, `operationalcredentials` definitions: all match), the privilege/auth-mode numbers (1-5, 1-3: match), `SetNodeBindings` (unicast/group shape checks), the ICD routes, `ListUpdates`, `ConsentUpdate` (int64 bound for 32-bit platforms), `CheckUpdates`, `GetControllerTopology`/`networkGraph`, `GetControllerDiagnostics`, `wireTime`, `controllerSessionOf`, `controllerPeerOf`, `GetControllerMdns`, `controllerTraceEvent`, `diagnosticsServed`.
- Tests read: `handlers_exposure_test.go`, `handlers_diagnostics_test.go`, `handlers_topology_test.go`, `handlers_surface_test.go`, `devicetype_names_test.go`, `handlers_endpoint_test.go`, the controller fixtures list and diagnostics tests in `handlers_controller_test.go`, `gofabric/e2e_test.go:280-330`.

**Count:** 0 high, 2 medium, 6 low, 4 note (12).
