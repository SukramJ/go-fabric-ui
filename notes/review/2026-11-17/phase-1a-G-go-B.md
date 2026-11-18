# Phase 1a — G-go-B: controller port, memory controller, go-fabric adapter

HEAD `49cc081`, go-fabric `v0.22.2` (export, `gf:` below). Scratch proofs in
the session scratchpad (`gbtest/`, a module with `replace` onto the
worktree); their code is quoted where a finding rests on them.

### (held) — 1 row held back (decision 25), published with the pull request that fixes it

### UR-P1-conc-61 — Two label writes: the older one wins and is persisted

- **Severity:** low
- **Where:** `fabricui/gofabric/gofabric.go:566`, `:573-583`; `gf:controller/client/label.go:128-133`
- **Claim:** `SetFabricLabel` sets the label, answers, and starts a goroutine whose `Registry.SetFabricLabel(ctx, label)` first calls `f.SetLabel(label)` again with the captured label. Goroutine start order is not the request order, so after `a` then `b` the fabric (persisted record, nodes' push) often ends at `a` while the second answer said `b`.
- **M** — scratch test, 300 rounds of `SetFabricLabel(a); SetFabricLabel(b)`, 2 ms later `Fabric().Label`: `label "a1", want "b1"`; reverted in 138–187 of 300 rounds.
- **Fix:** the goroutine pushes only (do not re-set the label): one worker per source that reconciles the current label, coalescing requests; or ask go-fabric for a push-only call.
- **Contract change:** no.
- **Red test:** two `SetFabricLabel` in a row on the fabrictest controller, `Fabric().Label` is the second, 100 rounds.

### UR-P1-conc-62 — The label push and a running hooks timer outlive Hooks.Close and race ctrl.Close

- **Severity:** low
- **Where:** `fabricui/gofabric/gofabric.go:573-583` (2 min, `context.Background()`), `:576`; `fabricui/gofabric/hooks.go:144-155`, `:108`; `gf:controller/controller.go:641` (`Registry()` reads `c.reg` plainly), `:570-573` (`Close` writes `c.reg = nil` unlocked)
- **Claim:** the shutdown order the package documents (`svc.Close()`, `hooks.Close()`, then `ctrl.Close`, `gofabric.go:22`) stops nothing of the push goroutine, and `Hooks.Close` stops timers but does not wait for a callback already past its `current` check. Both call `s.ctrl.Registry()` (and the push then the registry, its fabrics' DB and nodes) concurrently with `ctrl.Close`: a data race on `c.reg`, and a push into a closing registry. The adapter's `reg == nil` checks only work if the nil is observed through a happens-before edge, which there is not.
- **M** for the lines; the race **I** (the scratch `-race` run did not hit it on the fake site, where the push ends at once).
- **Fix:** give the source a `Close` (or let `Hooks.Close` do it) that cancels a source-wide context the push derives from and waits for it (`sync.WaitGroup`); make `Hooks.Close` wait for running timer callbacks.
- **Contract change:** no.
- **Red test:** `-race`: a push blocked on a node (a site device that delays UpdateFabricLabel), then `hooks.Close(); ctrl.Close()`.

### UR-P1-err-63 — Consent with a mismatching vendor/product answers 502 "The node failed"; the memory controller accepts it

- **Severity:** low
- **Where:** `fabricui/gofabric/updates.go:195-202` → `nodeError` (`gofabric.go:613-626`); `gf:controller/softwareupdate/manager.go:1241-1250` (`ErrTransferOptions`, `ErrMismatch`); `fabricui/handlers_controller.go:191-193`; `fabricui/memory_controller.go:582-586`
- **Claim:** `AddUpdateConsent`'s argument errors (`ErrMismatch`, `ErrTransferOptions`) become `ErrNodeFailed`: a 502 `node_error` plus a `node_failed` warning in the log for a request error the node never saw. The memory controller accepts any vendor/product, so the handler tests and fixtures cannot show the adapter's answer.
- **M** (lines read).
- **Fix:** map `softwareupdate.ErrMismatch` and `ErrTransferOptions` to `ErrInvalidArgument` (400) in `Consent`; give the memory controller the same refusal for a node whose vendor/product differs.
- **Contract change:** no (400 is the route's existing problem; check the contract lists it for `POST /controller/updates/consent`).
- **Red test:** memory controller `Consent(0x11, 0xFFF1, 0x9999, 2)` → `ErrInvalidArgument`.

### UR-P1-bound-64 — The discovered-device cache is never bounded or aged, and a stale entry beats a fresh discovery

- **Severity:** low
- **Where:** `fabricui/gofabric/gofabric.go:124-127`, `:414-418`, `:480-497`, `:526-528`
- **Claim:** every `Discover` adds each answer by instance name; only a successful `CommissionDiscovered` deletes one. Instance names change with every commissioning window, and anything on the LAN can advertise commissionable instances, so the map grows for the process's life (admin-triggered). `CommissionDiscovered` prefers a cached `client.Device` of any age over rediscovery, so a device whose address changed since is commissioned against stale addresses and fails instead of being found again.
- **M** (lines); growth and stale use **I**.
- **Fix:** keep only the last discovery's answers (replace the map per `Discover`) with a timestamp; use an entry younger than e.g. `rediscoverWindow`×N, else rediscover.
- **Contract change:** no.
- **Red test:** two `Discover` with disjoint answers → the cache holds only the second's.

### UR-P1-adpt-65 — Long discriminator 0 is treated as "absent"

- **Severity:** low
- **Where:** `fabricui/gofabric/gofabric.go:485-487`; `fabricui/handlers_controller.go:1304-1307`
- **Claim:** the handler turns the optional `long_discriminator` into a `uint16` where 0 stands for "not sent"; the adapter then answers `ErrNoDevice` (404) without searching. 0 is a valid 12-bit discriminator, so a device with discriminator 0 that the adapter no longer caches (another process, UR-P1-bound-64's fix, a restart) cannot be commissioned from a discovered entry.
- **M** (lines).
- **Fix:** a presence flag (`*uint16` or `HasLongDiscriminator`) in `fabricui.Discovered` for this input, or rediscover by instance name when the discriminator is absent.
- **Contract change:** no (port only).
- **Red test:** `CommissionDiscovered` with `{InstanceName: "", LongDiscriminator: 0}` on a site device with discriminator 0 commissions it.

### UR-P1-leak-66 — Pairing failure text reaches the ungated diagnostic trace (extends UR-P0-leak-31)

- **Severity:** note
- **Where:** `gf:controller/diagnostics.go:248-251`, `:262-270` (`err` text in Message and Detail); `fabricui/gofabric/diagnostics.go:87-101` (verbatim); `fabricui/handlers_controller.go:1866-1870`; `GET /controller/events` and `GET /controller/stream` absent from `controllerAdmin` (`handlers_controller.go:51-74`)
- **Claim:** UR-P0-leak-31 covered problem details and the log. The same commissioner error text (an SSID, the commissionee's `DebugText`) is also recorded by go-fabric's `Pair` into the controller's trace, which the adapter forwards as is and the handler serves to every reader, newest first and live. `CommissionDiscovered` does not go through the trace (it calls the registry directly, so its failures are not recorded there at all).
- **M** for the sites read.
- **Fix:** decide it with UR-P0-leak-31/27 (decision 20): gate the trace route and event, or strip `Detail["err"]` and the sentence's error tail in `diagnosticEventOf` for kind `pairing`.
- **Contract change:** yes if gated.
- **Red test:** a failed `Pair` with a Wi-Fi credential → no event of `GET /controller/events` holds the SSID.

### UR-P1-api-67 — A source without Hooks cannot release its subscriptions

- **Severity:** note
- **Where:** `fabricui/gofabric/gofabric.go:59-64`, `:105-110`; `fabricui/gofabric/topology.go:69-83`
- **Claim:** `ControllerOf` subscribes to the topology service and the update manager; only `Hooks.Close` releases them. `ControllerOptions.Hooks` is documented as optional, and `closeTopology` is unexported, so a host without hooks leaks both observers until the service or controller closes.
- **M**.
- **Fix:** return a source with an exported `Close` (also the hook for UR-P1-conc-62's push), or document that `Hooks` is required for a clean shutdown.
- **Contract change:** no (Go API).
- **Red test:** `ControllerOf` with `Topology` and nil `Hooks`, closed through the new API → the service has no observer.

### UR-P1-bound-68 — Hooks.ids is never pruned

- **Severity:** note
- **Where:** `fabricui/gofabric/hooks.go:35-37`, `:112-117`, `:179`
- **Claim:** a node id is remembered per local id forever; `Decommissioned` emits but does not delete. Growth is one entry per commissioning in the process's life.
- **M**.
- **Fix:** delete the entry after the `node_removed` emit (and stop its pending timer, which can otherwise emit `node_updated` after `node_removed` — **I**).
- **Contract change:** no.
- **Red test:** after `Decommissioned`, `len(h.ids) == 0` and no `node_updated` follows.

### UR-P1-doc-69 — The adapter's import list in CLAUDE.md does not match the guard

- **Severity:** note
- **Where:** `CLAUDE.md` Rules ("controller facade, thread, cluster definitions"); `internal/arch/arch_test.go:53-68`; `fabricui/gofabric/gofabric.go:4-10`
- **Claim:** no breach: every adapter import (`client/discovery`, `controller`, `controller/client`, `controller/commissioner`, `controller/topology`, `controller/softwareupdate`, `diag`, `thread`, `cluster/spec/…`, facade `schema`) is on the guard's list. But `diag` and `client/discovery` are outside `controller/…`, which CLAUDE.md names, and the package doc omits `client/discovery`, `commissioner` and `schema`.
- **M**.
- **Fix:** name the guard's list (or "see `internal/arch`") in CLAUDE.md and the package doc.
- **Contract change:** no.
- **Red test:** none possible (doc); struck unless a doc guard is added.

### UR-P1-conc-70 — MemoryController hands out and keeps aliased slices; one read after unlock

- **Severity:** note
- **Where:** `fabricui/memory_controller.go:930-931` (Read returns the stored `RawMessage`), `:940-954` (Write stores the caller's bytes), `:806-809` (`Endpoints` clones one level: `Clusters`, `DeviceTypes`, `Parent` shared), `:767-776` (`Record().Addresses`, `Properties().DeviceTypes` shared), `:478-493` (`m.fabric.CompressedFabricID` read after `m.mu` is released)
- **Claim:** the type says "safe for concurrent use"; it is while no caller mutates what it receives. A handler or host test that edits a returned slice in place edits the controller's state. The adapter returns fresh values everywhere, so a caller that works against the memory controller works against the adapter; the reverse risk is a test passing on shared state.
- **M**.
- **Fix:** `slices.Clone` the bytes in Read/Write, deep-clone `Endpoints`, `Record`, `Properties`; read the compressed id under the lock.
- **Contract change:** no.
- **Red test:** mutate `Read(...)[0].Value[0]`, read again → unchanged.

## Read and found nothing

- `fabricui/controller.go` whole: port types, error kinds, `ErrTermsRequired`, `EventPriorityName`.
- `fabricui/gofabric/gofabric.go`: `ThreadDatasetName` (decoder errors carry offsets and lengths only, `gf:thread/tlv.go:33-63`, `dataset.go:182`, `:261`; no dataset byte echoed), `Fabric`/`beUint64`, `Capabilities`, `nodeOf`, `Nodes`, `lookup`, `networkOf` (errors name the id, never a secret), `bleFor`, `Pair` (parser error replaced, `:292-295`; ctx passed through; log carries the reason only), `progressFor`, `termsOf`, `commissionError`, `identifierOf` (bounds), `Discover`, `discoveredOf`, `cutPlus`, `Decommission`, `Credentials`, `SoftwareUpdates`, `Observe`, `nodeError`.
- `fabricui/gofabric/node.go`: conversions (`nodeRecord`, `Properties`, nil pointers checked), `imageValue`/`endpointValue`, `Basic`, `Endpoints`, `parentOf`, `deviceTypes`, `EndpointDetail`, `ClusterDetail` (`ErrNotRead` vs `ErrClusterNotFound`), `tagsOf`, `labelsOf`, `conformanceOf`, `clusterDetail`, `commands`, `Events`, `Read` (concrete paths, `NoFabricFilter`), `fabricIndex`, `aclList` (null subjects/targets, our index in every entry, against `accesscontrol.AccessControlEntryStruct`), `bindingList`, `Write` (path check for typed lists, timed by definition), `Invoke`, window, `Interview` (zero path = full wildcard, `gf:controller/client/surface.go:29`, `:134-142`), `Ping`/`pingAddresses` (bounded parallelism, ctx), `pingSession`, ICD calls.
- `fabricui/gofabric/hooks.go`: registration, `bind` after Close, timer coalescing and Close, `chain`/`both*`, `connectionOf`; no lock held while publishing; `fabricui.Events.Publish` copies subscribers under its lock and calls them outside (`fabricui/events.go:222-232`).
- `fabricui/gofabric/topology.go`: `NetworkTopology`, `watchTopology`, `networkTopologyOf` (empty snapshot, unknown kinds kept), `peerOf`, `threadUnknownOf`, `borderRouterOf`, `otherOf`, `accessPointOf`, `linkOf`, `details`.
- `fabricui/gofabric/diagnostics.go`: `diagnosticsOf`, `peerStateOf`, `DiagnoseMDNS`, `OnDiagnosticEvent` (beyond UR-P1-leak-66).
- `fabricui/gofabric/updates.go`: `Queue`, `Check`, `watchUpdates`, `updateEvent`.
- `fabricui/memory_controller.go`: `MemoryCredentials` (locks on every map, dataset cloned in and out), Pair/commission/Discover/CommissionDiscovered, Decommission, SetFabricLabel (handler bounds the label, `handlers_controller.go:1471-1478`), Diagnostics, node methods' locking, Invoke, Ping, ICD.
- Tests: `gofabric_test.go`, `hooks_test.go`, `node_test.go`, `ping_test.go`, `topology_test.go`, `e2e_test.go` harness and test list; `go vet` clean, `go test -race ./fabricui/gofabric` passes, coverage 71.4 %.
- Import rule: no breach (UR-P1-doc-69).

Count: 0 high, 1 medium (held), 5 low, 5 note (11).
