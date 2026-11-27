# Read: #85 close-out (ui-396, err-404, api-391, test-403) at go-fabric v0.28.0

Date 2026-11-27 (Fable). Source: the module cache
`thingfabric.dev/go-fabric@v0.28.0`, cited `gf:<path>:<line>`; this repo at
`51bd01c` (`v0.15.0`). Nothing was run: every statement is a read. What was
not read is marked **not read**; what follows from reads but was not
executed is marked **inferred**.

## a) ui-396 — does a failing assembly raise the observer?

### The hooks

- The handler's stream turns `node.Observers.OnSurfaceChanged` into
  `surface_changed`, one event per call (`fabricui/events.go:343-345`),
  registered with `node.Observe` (`fabricui/ports.go:113`).
- `endpoints_changed` has no go-fabric source. The handler publishes it
  itself, in two places, both only after `reassemble(ctx)` returned nil:
  `fabricui/handlers_exposure.go:223-229` (`PUT /node/exposure`; on an
  error it logs `fabricui.exposure.reassemble_failed` and publishes
  nothing) and `fabricui/handlers_node.go:367-372`
  (`POST /node/maintenance/reassemble`; on an error it answers 500 before
  the publish).
- `reassemble` is the host's `Options.Reassemble` / `NodeEntry.Reassemble`
  (`fabricui/options.go:122,235`). What a host calls inside it is host
  code: **not read**, not answerable from go-fabric.

### go-fabric v0.28.0

- `OnSurfaceChanged` is raised in exactly one function,
  `Node.announceSurface` (`gf:node/surface_stability.go:188`, the hook at
  `:277-285`). It has two callers: `Node.startLocked`
  (`gf:node/node.go:1447`) and `Aggregator.reassembleLocked`
  (`gf:node/aggregator.go:1659`). `grep` for `announceSurface(` and
  `br.Reassemble` over `node/*.go` (non-test) finds no other.
- `reassembleLocked` (`gf:node/aggregator.go:1652-1663`): when
  `br.Reassemble` fails it sets `behind` and returns at `:1653-1656`,
  **before** `announceSurface` at `:1659`. A reassembly that fails raises
  no `OnSurfaceChanged`.
- Inside `announceSurface` the hook runs only for an endpoint whose
  digest differs from the persisted one (`:206-213`, the `drifts`); an
  endpoint without a persisted surface — its first assembly — is stored
  without an announcement (`:204-205`), and no drift returns early
  (`:215-217`). So a successful reassembly raises `OnSurfaceChanged` zero
  or more times.
- A `Partition.Batch` can return an error although a reassembly inside it
  succeeded and announced (`gf:node/aggregator.go:635-733`):
  - `:693` `a.reassemble` succeeds (announces), then `fnErr`, a call's
    error or `storeErr` make the result a `*BatchError`
    (`batchResult`, `:820-826`);
  - `:693` succeeds (announces), the second reassembly at `:731` fails:
    `BatchError.Reassembly` is set, the first one's `OnSurfaceChanged`
    calls have already run.
  Whether a host's `Reassemble` returns that error to the handler is host
  code: **not read**.
- `OnSurfaceChanged` is not tied to the selection: `Node.Start` raises it
  for a surface that changed against the previous run
  (`gf:node/node.go:1445-1447`), and "an Add or Remove from another
  goroutine during fn reassembles on its own" (`gf:node/aggregator.go:632-634`).
- Whether the removal of an endpoint changes a surface that is announced
  (the Aggregator's or the root's): **not read**.

### Result per path

| Path | `surface_changed` | `endpoints_changed` |
| --- | --- | --- |
| Rebuild through a route, `reassemble` returns nil | 0..n (one per endpoint whose surface changed against the persisted one) | 1 |
| Rebuild through a route, the bridge reassembly fails (`aggregator.go:1653-1656`) | none | none |
| Rebuild through a route, `reassemble` returns an error after a reassembly succeeded (`:693` then `:731` or `:826`), if the host returns the batch's error | 0..n possible | none |
| An assembly the host runs on its own, or `Node.Start` | 0..n | none (unless the host publishes it through its `Events`) |

### Verdict

"Clear on `endpoints_changed` only" is correct.

- The register's suspected case as worded — a failed reassembly's
  `surface_changed` — does not exist for a reassembly that fails in the
  bridge (`aggregator.go:1653-1656`).
- `surface_changed` is still no sign that the stored selection is in
  effect: it can precede a rebuild error (`:693` + `:731`/`:826`), and it
  is raised by assemblies that have nothing to do with the selection
  (`node.go:1447`, `aggregator.go:632-634`). It is also not raised by
  every successful rebuild (`surface_stability.go:204-205,215-217`).
- `endpoints_changed` is published only after `reassemble` returned nil
  (`handlers_exposure.go:228`, `handlers_node.go:372`): the one event
  that states what the notice asks.
- Consequence for `ui/src/lib/store.svelte.ts:480-490`: split the two
  cases. `endpoints_changed`: `exposureNotInEffect = false`, `loadStatus`,
  `topologyEpoch++`. `surface_changed`: `loadStatus`, `topologyEpoch++`,
  the notice untouched. The comment at `:482-486` ("An assembly ran … so
  the stored selection is in effect") is wrong for `surface_changed` and
  is rewritten from this note.
- Cost: a rebuild the host runs on its own no longer clears the notice in
  an open page unless the host publishes `endpoints_changed` through its
  `fabricui.Events`. `docs/contract.md:265-267` says today "the routes
  publish `endpoints_changed` alone"; `docs/provider.md:87-89` names only
  the selection event for a host. A sentence in `docs/provider.md` /
  `docs/embedding.md` (a host that rebuilds on its own publishes
  `endpoints_changed`) is the host's obligation of this wave.
- Test that changes: `ui/src/lib/pages/node/Fabrics.test.ts:236-251`
  loops over both event types and expects the notice cleared.

## b) err-404 — `Discover` without an address, `ErrConfiguration` on the discovered path

### Can `Discover` return a device without an address? No.

- `Controller.Discover` → `DiscoverWith` (`gf:controller/code.go:106-108,130`)
  searches with one transport, `discovery.IPTransport(sc)` (`:140`), as a
  `FindAll` (`:161-165`) or, with `First`, a `FindInstance` (`:147-153`);
  both take the candidate's `Device` as the transport reported it
  (`gf:client/discovery/session.go:331`).
- `ipTransport.FindContinuously` reports a device only
  `if len(d.Addresses) > 0` (`gf:client/discovery/session.go:58-65`); the
  godoc says the same (`gf:controller/code.go:128-129`, "a device is
  reported once it has an address"). The scanner below it delivers on the
  same condition (`gf:client/discovery/scanner.go:607-610`, `:875-885`)
  and hands each waiter a clone (`:301-305`), so a later expiry in the
  cache does not empty a device already returned.
- An address is built from an A/AAAA record whose IP is valid
  (`ipOf`, `gf:client/discovery/service.go:581-592`; the one caller of
  `updateAddress`, `:566-568`) and the SRV entry's port (`:596-597`); an
  SRV record with port 0 or no target is not installed
  (`gf:client/discovery/name.go:144-149,386`). So "an address without IP
  or port" does not leave the discovery either. How `srvEntry.port` is
  set from the installed record: **not read** beyond `service.go:222,432`
  (`updateService(r.TTL, r.SRV.Target, r.SRV.Port, …)`).
- `client.Device` is `discovery.Device` (`gf:controller/client/surface.go:120`).

### `ErrConfiguration` cases that can reach `PairDiscovered`

`Controller.PairDiscovered` (`gf:controller/controller.go:1261-1285`),
through `pair` (`:1318-1350`):

| # | Line | Case | Reaches go-fabric-ui's discovered path? |
| --- | --- | --- | --- |
| 1 | `:1268` | `PairOptions.BLE` set | no: the adapter clears it (`fabricui/gofabric/gofabric.go:613`, `opts.BLE, opts.Address … = nil, nil`) |
| 2 | `:1270` | `PairOptions.Address` set | no: same line |
| 3 | `:1272` | the device has no address | no: the device is the last `Discover`'s answer (`gofabric.go:554-562,566-574`) or a fresh one (`:598-606`), and `Discover` returns none without an address (above) |
| 4 | `:1275` | invalid passcode (`setup.IsValidPasscode`) | no through the handler: `fabricui/handlers_controller.go:1482-1485,1970` refuses 0, above 99999998 and the listed codes first; go-fabric's list is `gf:secure/setup/setup.go:334-347,364-374` — the same set (99999999 and 0 fall to the range check). A caller of the port other than the handler could reach it |
| 5 | `:1279` | an address without IP or port | no (above) |
| 6 | `:86`, returned at `:1323-1326` (`c.open()` fails) | the controller is closed | **yes** |
| 7 | `:86`, returned by `Discover` at `gf:controller/code.go:131-134` | the controller is closed, on the adapter's re-discovery (`gofabric.go:598-601`, mapped by `commissionError`) | **yes** |

`controller/client` and `controller/commissioner` do not name
`ErrConfiguration` (`grep -l` over both, non-test: no file), so
`Registry.ForDevice` and `Registry.Commission` (`controller.go:1341-1348`)
add no further case.

The code path has the same closed case (`Pair` → `pair` → `:1323-1326`;
`gofabric.go:364-366` → `commissionError` → 400); its other cases are in
the re-check below.

### The mapping today

- `commissionError` (`fabricui/gofabric/gofabric.go:444-445`):
  `ErrPairingCode` and `ErrConfiguration` → `ErrInvalidArgument` →
  400 "Invalid request" (`fabricui/handlers_controller.go:241-242`),
  progress detail `invalid_argument` (`:1570-1571`).
- The registry path's 502 for a closed controller: from the register
  (`notes/review/2026-11-17/phase-4r2-verify.md:580-587`); the old code
  was **not read** again.

### Re-check after the parallel read (2026-11-27)

- The discovery route answers a closed controller with 500 today: the
  adapter's `Discover` returns go-fabric's error as it is
  (`fabricui/gofabric/gofabric.go:486-489`; `errClosed` from
  `gf:controller/code.go:131-134`), the handler passes it on
  (`fabricui/handlers_controller.go:1604-1606`), and `controllerError`
  has no kind for it (`:251-252`, the default).
- A host template's `ErrConfiguration` reaches the code path, not
  `PairDiscovered`. `Pair` starts from the template (`gofabric.go:349`,
  `opts := s.opts.Pair`) and clears only `BLE`, which it sets again when
  the BLE decision is "used" (`:350-354`); the template's `Address`
  stays. So `gf:controller/controller.go:1213` (BLE and Address both
  set) and `DeviceAt`'s "no IP or no port" (`:1416`, called at `:1217`)
  are reachable with a template that carries an `Address`, and answer
  400 for a valid request today. `:1224` (BLE without a network) is not:
  the decision uses BLE only with a credential (`gofabric.go:324-327,348`).
  `PairDiscovered` gets both fields cleared (`gofabric.go:613`).
- A code's own faults are not `ErrConfiguration`: the adapter refuses an
  unparsable code itself (`gofabric.go:340-343`); that
  `ParsePairingCode`'s errors all wrap `ErrPairingCode`: **not read**.

### Verdict (revised): `ErrConfiguration` gets no port kind — 500

The first verdict of this note (closed controller → 502 by a
`Registry() == nil` probe, the rest stays 400) is withdrawn.

- No `ErrConfiguration` that can arrive is the request's fault: on the
  discovered path only the closed controller arrives (rows 6, 7); on the
  code path the closed controller and a host template (above). "The rest
  stays 400" was wrong for the template.
- 502 `node_error` is "the interaction with a controller's node failed;
  `detail` is the node's answer" (`docs/contract.md:589-591`). A closed
  controller or a wrong template is no node's answer; it is the host's
  state, which is what 500 with the cause in the log says
  (`docs/contract.md:531-533`).
- The same closed controller already answers 500 on `discoverDevices`;
  one state, one status.
- No probe, no race, no question to go-fabric needed: `commissionError`
  drops `controller.ErrConfiguration` from the `ErrInvalidArgument` case
  (`gofabric.go:444-445`), on both paths; `ErrPairingCode` stays 400.
  The error then falls to the function's last line — which wraps
  `ErrNodeFailed` (`:451`) — so `ErrConfiguration` needs its own case
  that returns the error without a port kind. Progress detail becomes
  `failed` (`handlers_controller.go:1578-1579`). `commissionNode`
  declares 500.
- The one loss: an invalid passcode handed to the port by a caller other
  than the handler (row 4) would be 500. The handler is the port's only
  caller in this repo (`handlers_controller.go:1553`) and refuses it
  first (`:1482-1485`).
- The registry path's old 502 is from the register only; the old code
  was not read, and the reasons above hold without it.

## c) api-391 — what was read, and the owner's decision

### What was read

- `controllerError` maps the port's kinds centrally
  (`fabricui/handlers_controller.go:218-254`): `ErrNodeNotFound`,
  `ErrEndpointNotFound`, `ErrClusterNotFound`, `ErrNoDevice`,
  `ErrUnknownCredential` → 404; `ErrInvalidArgument`,
  `ErrInvalidCredential` → 400; `ErrICDMultiAdmin`, `ErrConflict`,
  `*ErrTermsRequired` → 409; `ErrNodeFailed`,
  `context.DeadlineExceeded` → 502; anything else 500.
- 501 and 503 are already closed: only the declaring operations convert
  them (`notImplemented`, `notRead`, `:256-273`), elsewhere they fall to
  500 (decision 79, `notes/review/2026-11-17/decisions.md:201`), every
  controller operation declares 500 (decision 82, `:217`), and
  `TestController_NotImplementedAndNotReadOnlyWhereDeclared`
  (`fabricui/undeclared_status_test.go:84-116`) holds it.
- The onboarding routes are closed for all kinds: `onboardingError` maps
  what the operation declares and wraps the rest in `opaqueError`, which
  has no `Unwrap`, so `controllerError` answers 500
  (`handlers_controller.go:2211-2245`; U12 decision 2;
  `docs/contract.md:1068`). `checkUpdates` is closed the same way by hand
  (`:1801-1806`).
- `docs/contract.md:522-533` (Problems): a 500 carries no detail, a 502
  the node's answer; `problem-conflict` (`:563-567`) is worded for the
  pairing routes and the routes under `/controller/nodes/{id}`.
- Declared statuses per controller operation, scanned from
  `contract/openapi.yaml` 0.15.0 with the status guard's own patterns
  (`internal/statusguard/statusguard.go:150-153`), against the handler
  sites that return a port's error unfiltered:

| Operation | Site | Port call | Undeclared kinds that reach `controllerError` |
| --- | --- | --- | --- |
| `getNodeEndpoint` | `:645-647` | `EndpointDetail` | 409, 502 |
| `getNodeCluster` | `:667-669` | `ClusterDetail` | 409, 502 |
| `pingNode` | `:717-719` | `Ping` | 409, 502 |
| `discoverDevices` | `:1604-1606` | `Discover` | 404, 409 |
| `setFabricLabel` | `:1727-1732` | `SetFabricLabel` | 404, 502 (409 declared since 0.15.0) |
| `getControllerTopology` | `:1835-1837` | `NetworkTopology` | 404, 409 |
| `listCredentials` | `:1640-1642` | `CredentialStore.List` | 404, 409, 502 |
| `setCredential` | `:1678-1692` | `SetWiFi` / `SetThread` | 404, 409, 502 |
| `removeCredential` | `:1705-1706` | `Remove` | 409, 502 |
| `consentUpdate` | `:1790-1791` | `Updates.Consent` | 409, 502 |

  The first six rows are the ports `errController` overrides
  (`undeclared_status_test.go:16-80`): 12 answers now, 13 before 0.15.0
  declared the label's 409 — **inferred** to be the register's "13
  answers"; the temporary guard case is not in the tree and was not run.
  The last four rows (the credential and update ports) are beyond that
  count. Operations without a fallible port call at the read sites
  (`getControllerNode`, `listNodeEvents`, `getNodeIcd`, status, lists,
  diagnostics, stream) are not affected; their bodies were read only as
  far as the `grep` of error returns shows.
- The shipped adapter reaches exactly one of these: `Ping` returns the
  request context's error when the host's deadline ends it
  (`fabricui/gofabric/node.go:719-722`), which `controllerError` answers
  502 on `pingNode` (`:248-250`), undeclared.

### Decision (owner, 2026-11-27): variant C

Proposed were B (close every affected operation to 500, no contract
change) and A (declare every status); the owner took B with `pingNode`
declared instead of closed.

- `pingNode` declares 409 and 502 (contract 0.16.0, additive, one route,
  `POST /controller/nodes/{id}/ping`; fixtures `.409` and `.502`). 502 is
  the answer the shipped adapter gives (`gofabric/node.go:719-722`), and
  every other acting route under `/controller/nodes/{id}` declares both
  (the table's scan).
- Every other row of the table — the credential and update operations
  included — is closed: the handler passes the port's error through one
  helper beside `opaqueError` that keeps the kinds the operation declares
  and answers 500 for the rest.
- The temporary guard case becomes a permanent test in
  `fabricui/undeclared_status_test.go`: every port kind against every
  fallible controller route, the credential and update sources failing
  too. What judges it is the status guard's report of an answer whose
  status the operation does not declare
  (`internal/statusguard/statusguard.go:130-132`; the parallel read's
  `:131-133` is that block, one line off).
- Release `v0.16.0`.

### Rejected

- Variant A — declare all 22 undeclared statuses on the ten operations:
  it would declare statuses without a meaning (404 on `discoverDevices`
  or `listCredentials`, 409 on `getControllerTopology`) and tie the
  contract to whatever a host's source returns.
- Variant B pure (`pingNode` 500, no contract change).

## d) test-403 — where a product's check fails at v0.28.0

- `Manager.checkProductForUpdates` (`gf:controller/softwareupdate/manager.go:1970-2015`)
  returns an error at one place: `Images.DownloadUpdate` failing
  (`:1980-1981`). Its caller joins it per product and goes on with the
  next (`:1907-1911`, `productCheckFailed`). The parallel read's
  `:1976-1982` is that block.
- The image store is the host's to give: `controller.OTAConfig.Images`
  (`gf:controller/ota.go:39-41`), built by `softwareupdate.NewImages`
  from an `ImagesConfig{Blobs, Ledger, …}`
  (`gf:controller/softwareupdate/images.go:320-369`). So no host fixture
  is needed, only a commissioned node (the check runs per product of the
  nodes).
- Route (i), a ledger record without a checksum: refused at
  `images.go:1046,1053`. `ImagesConfig.Ledger` is a `*dcl.OTAVersions`
  (`images.go:25`, import; the field in `:320-354`), and `dcl` is in
  neither the adapter's nor the tests' import set
  (`internal/arch/arch_test.go:51-…`, `:79-…`, `:135-154`; `grep dcl`: no hit) —
  it would need the guard widened.
- Route (ii), an offline store: `Ledger == nil` is offline
  (`images.go:365`), and offline `downloadUpdate` fails for a stored
  image it cannot read (`:1098-1100`), one it must not replace
  (`:1101-1103`) or none valid (`:1105-1106`); before that
  `DownloadUpdate` refuses a stored image that is neither host-supplied
  nor bound to a ledger checksum (`:1047-1053`, `hostSupplied`
  `:117-119`). All of it is `controller/softwareupdate`, which the
  adapter's set allows (`arch_test.go:55`).
- **Not verified:** which stored blob `CheckForUpdate`
  (`images.go:861-…`) offers for a product, and so which of route (ii)'s
  failures a test's own `Blobs` provokes for one product while another
  product's image stays valid. The implementer establishes it by the
  test, not from this note.

## Not read

- A host's `Options.Reassemble` (go-ha2fabric, go-fabric-server,
  openccu-loom): whether it returns a `*BatchError` with a successful
  first reassembly.
- Whether removing an endpoint changes an announced surface.
- `srvEntry.port`'s assignment; the registry path's old code; whether
  every error of `ParsePairingCode` wraps `ErrPairingCode`.
- `Images.CheckForUpdate` (which stored image it offers).
- `context.Canceled`'s way through the strict handler's error function
  (relevant to variant B's helper: it must not turn a client's
  cancellation into a logged 500 where it is not one today).
- matter.js: nothing was read; the matter.js lines above are go-fabric's
  own comments, not verified.
- The parallel read's note itself: it was overwritten before it was
  read; its claims reached this note through the coordinator's message
  and were checked as above.
