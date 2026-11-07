<!-- Written 2026-11-07 on the owner's decision of the same day. The map
     notes/reads/read-u7-gofabric-coverage.md is an inventory, not a
     citation: every file:line below was re-read at the pins named here. -->

# Brief U7 — completeness against go-fabric v0.18.0 (go-fabric-ui)

Date 2026-11-07. Owner decision 2026-11-07: the five scope items below;
the controller-side camera and groups out; §11's four questions answered
the same day, each with the recommendation (§11, "Decided"). **Model:** Opus. Repo
`~/Documents/GitHub/go-fabric-ui`, worktree `../go-fabric-ui-u7`, branch
`u7/completeness`, cut from main after the 0.6.1 release merged. Pins:
go-fabric **v0.18.0** (`f9a81ae4`; `go.mod` already requires it, PR #50),
go-fabric-ui main `7fac9cd`, connectedhomeip `9fb3392c307` (the local
checkout, for the comparison in §11 only; go-fabric's own chip citations
are at `6170af84`). Pattern: `notes/briefs/2026-11-04-U5a-operation.md`.
One wave, one pull request, one hand-back. No tag; version **0.7.0
unreleased** (after 0.6.1, `internal/version`, `ui/package.json`,
`changelog.md`); contract **0.7.0** (`contract/openapi.yaml:4`);
`contract/previous/openapi.yaml` = the contract as released in 0.6.1
(today it still holds 0.5.0, `contract/previous/openapi.yaml:4` — copy in
the first commit).

**Hosts today** (read-only, 2026-11-07): go-ha2fabric main `34075a6`
(0.5.3: go-fabric v0.18.0, go-fabric-ui v0.6.1) has U6b — the node
registry `internal/nodes`, `cmd/ha2fabric/nodes.go` (`served` with its
`*node.Node` at `:31`, the `fleet` as `fabricui.NodeRegistry`, `:99-128`),
and `fabricui.Options{Node: def.n, …, Registry: fl, NodeConfig:
nodeConfig(cfg)}` at `cmd/ha2fabric/main.go:108-111`. go-fabric-server
main `b947c93` (go-fabric v0.18.0, go-fabric-ui v0.6.1, `go.mod:6-7`)
wires the registry at `internal/daemon/web.go:42` (`o.Node = def.node`
`:39`, `o.NodeConfig` `:40`). §3.1 and §5 cite these lines.

## 0 Working rules

Commits as `SukramJ <sukramj@icloud.com>` (`git var GIT_AUTHOR_IDENT`
first), `<area>: <what> (U7)`, no attribution lines, stage named paths.
Guards under `flock ~/.cache/go-fabric-scratch/guards.lock` with
`export PATH=$HOME/go/bin:$PATH GOTMPDIR=$HOME/.cache/gotmp`, logs and
`summary.log` in `~/.cache/go-fabric-scratch/u7/`. No change in go-fabric,
go-fabric-server or go-ha2fabric: a gap there is an issue text in the
hand-back (§11's is filed: SukramJ/go-fabric#449). Every go-fabric line you build on
is read at v0.18.0 and cited `file:line`. Official text in English.

## 1 Scope as verified against the code

| Item | Map said | Verified | U7 does |
| --- | --- | --- | --- |
| 1 Observers | `OnFabricUpdated`, `OnCommissioningComplete`, `OnEvent` unread | `node/observers.go:63,71,127`; the UI sets 8 of 11 hooks (`fabricui/events.go:244-283`). go-fabric closes U2's controller gap too: `controller.PairOptions.Progress` (`controller/controller.go:611-618`, `PairingStep = commissioner.Step`, `commissioner/progress.go:67-73`) and `softwareupdate.Manager.Observe` (`controller/softwareupdate/observe.go:139`); the UI uses neither (handler-side publishes, `handlers_controller.go:1274-1290`, `:1510`, `:1530`) | three node events, step-level commissioning progress, update events from the manager (§2, §3) |
| 2 Diagnostics | `AttestationError`, `BootReason`, `SessionParameters` unread | `node/attestation.go:55-58`, `node/node.go:1138-1141`, `node/sessions.go:110-118` (`n.sec` is wired in `New`, `node/node.go:360`, so the read is safe on any node the host hands over) | status fields and a sessions block (§2, §3) |
| 3a Reannounce | `TriggerReannounce` unread | `node/reannounce.go:12-26`: a no-op on an advertiser without the method | admin route beside `GET /node/mdns` (§2) |
| 3b Factory reset | `Node.FactoryReset` unread | True, but the UI already has `POST /node/maintenance/factory-reset` = remove every fabric (`handlers_node.go:285-325`, `docs/contract.md:39`, confirmation `remove-all-fabrics`, `openapi.yaml:2292-2296`). `Node.FactoryReset` is different: it erases the store and endpoint records and **ends the node**; the host builds a new one (`node/factory_reset.go:20-45`, `fabrictest/device.go:396-416`). go-fabric-server's registry Delete already calls it (`go-fabric-server internal/daemon/registry.go:303`) | a second, host-wired action `erase` (§2, §3.1); the existing route stays as it is |
| 4 Surface | `Report.Policies`, `.Effective` unread | **Already served.** `GET /node/surface` writes `surface.Report.MarshalJSON` (`fabricui/handlers_surface.go:21`), which encodes `Policies` as hex and `Effective` (`go-fabric surface/policy.go:406-417`); the page renders both (`ui/src/lib/pages/node/Surface.svelte:203`, `:211-216`; contract `SurfaceEffective`, `ui/src/lib/api/schema.gen.ts:2375`). The map's "no" is a grep artefact: no Go selector reads the fields | nothing; the report records the finding |
| 5 Terms and Conditions | `TermsAndConditions` unreferenced | go-fabric offers the **acknowledgement** (`commissioner.Options.TermsAndConditions`, `commissioner/commissioner.go:154-163`, `:662-665`; flow step 8.3, `flow.go:710-742`; reachable from the facade only through `PairOptions.Edit`, `controller.go:619-620`). It offers no **requirement**: the flow reads none of TCMinRequiredVersion/TCAcknowledgementsRequired (`grep` over `controller/commissioner`, zero hits; the attribute ids exist, `cluster/spec/generalcommissioning/definition_gen.go:35-39`), and a refused CommissioningComplete is an untyped `ErrCommissioning` message (`flow.go:409-419`). chip at `9fb3392c307` is the same: the acknowledgement is a parameter (`src/controller/CHIPDeviceController.cpp:3996-4006`, chip-tool `--tc-acknowledgements`/`--tc-acknowledgements-version`, `examples/chip-tool/commands/pairing/PairingCommand.h:274-278`), no read of the requirement in `src/controller` | 5a in: the acknowledgement as an optional request field, exactly what chip-tool offers. 5b out: the device-stated requirement and the retry flow need go-fabric (SukramJ/go-fabric#449, §11) |

## 2 Contract 0.7.0 (additive; `make contract-check` refuses removals, `ui/scripts/contract-guard.mjs:22-59`)

- **`Status`** (`openapi.yaml:1771-1813`): `attestation_error` (string; the node runs on the test chain and refuses commissioning until the files are fixed; absent when nil — `AttestationError`), `boot_reason` (integer 0–6, GeneralDiagnostics BootReasonEnum, `generaldiagnostics/definition_gen.go:77-83`; absent without a node), `erase` (boolean: the host wires `Erase`, so the page shows the button; absent without one — the capability-flag shape of `exposure`, `node_registry`, `:1791`, `:1795`).
- **`SessionList`** (`handlers_diagnostics.go:66-69`): `parameters` → new `SessionParameters{idle_interval_ms, active_interval_ms, active_threshold_ms, data_model_revision, interaction_model_revision, specification_version_hex (8 hex digits), max_paths_per_invoke, supported_transports (integer bitmap), tcp (boolean, Parameters.SupportsTCP), max_tcp_message_size (present when tcp)}` — the fields of `secure/sessionparams/sessionparams.go:49-66`.
- **`POST /node/mdns/reannounce`** (`x-admin`, `operationId reannounce`) → 204, 403, 503; `/nodes/{node}/mdns/reannounce` (`reannounceForNode`, 404 for an unknown key), as every node route is mirrored (`openapi.yaml:878-891` pattern).
- **`POST /node/maintenance/erase`** (`x-admin`, `operationId erase`), body `EraseRequest{confirm: const "erase-node"}` (names the action, as `remove-all-fabrics` does, `handlers_node.go:282-285`; a route of its own, not a flag on the factory-reset route — owner, 2026-11-07) → 204 (store and endpoint records erased, the node stopped; the host brings a node back or restarts), 400 (not confirmed), 403, 500, 501 `not_implemented` (no `Erase` wired), 503 (no node); `/nodes/{node}/maintenance/erase` (`eraseForNode`, 404).
- **Node stream** (`openapi.yaml:394-397`, `docs/contract.md:183-218`): `fabric_updated` (`FabricEvent`, as `fabric_added`), `commissioning_complete` (`FabricEvent` with `fabric_index` alone, as `fabric_removed`, `events.go:250-251`), `cluster_event` (`EventRecord`, `openapi.yaml:2718-2731` — one shape for both families, as the event log route).
- **`CommissionRequest`** (`openapi.yaml:2927-2937`): `terms_and_conditions` → new `TermsAcknowledgement{version (integer 0–65535), acknowledgements (integer 0–65535)}`, both required inside the object, the object optional.
- **`ControllerEvent`** (`openapi.yaml:3035-3043`): `step` → new `PairingStep{number, sub, name, phase: enum [started, succeeded, failed, skipped]}` on `commissioning_progress` while the top-level `phase` stays `started` (an old client keeps reading "started", `controller-store.svelte.ts:131`); `update_kind` (enum `[available, consent_changed, check_finished, applying, done, failed, queued, transfer_started]` = `softwareupdate.UpdateEventKind`, `observe.go:29-59`), `update_version` (integer, the target version when the kind carries one), `update_error` (string, no secret) on `update`.
- **Problem codes:** none new (`openapi.yaml:1770`; `problem.go:15-23`, `registry.go:75-76`).
- `make generate`, `make gen-types`, fixtures recorded with `-update` and reviewed, every fixture used on both sides.

## 3 Go half

### 3.1 Ports — how each addition stays compatible

- **`NodeSource` grows by four methods** (`fabricui/options.go:25-73`): `AttestationError() error`, `BootReason() uint8`, `SessionParameters() node.SessionParameters`, `TriggerReannounce(ctx context.Context)`. `*node.Node` satisfies all four; both hosts hand it over directly (go-ha2fabric `cmd/ha2fabric/main.go:109` `Node: def.n`, `nodes.go:31`, and each registry node `nodes.go:190`; go-fabric-server `internal/daemon/nodes.go:98`, `web.go:39`, `registry.go:202`), so nothing breaks in the family. A host wrapping its node re-forwards, as 0.5.0 asked for the two callbacks (`docs/embedding.md:85-93`) — the changelog says so under a Breaking line. **Decided (owner, 2026-11-07), closing U3a's open question 2** (`docs/reports/U3a.md:188-191`): the port follows the facade, one shape; a wrapper host pays one forwarding line per method on a bump. Optional interfaces (`n.(interface{ TriggerReannounce(context.Context) })`, go-fabric's own idiom for its advertiser, `node/reannounce.go:23`) were weighed and rejected for the port: two shapes, and a wrapper that forgets a method loses a page field silently instead of failing to compile.
- **`Options.Erase func(ctx) error`, `NodeEntry.Erase`** — the shape of `Reassemble` (`options.go:89`, `handler.go:105`, `:209-210`, `:233`): a host lifecycle action the UI cannot run on a `NodeSource`, because `Node.FactoryReset` ends the node (`factory_reset.go:31-36`). nil → 501 and `Status.erase` false. The host's callback is `fabrictest.Device.FactoryReset` by meaning (`fabrictest/device.go:396-416`): `node.FactoryReset`, a new node on the erased store, start — or the daemon's exit for a supervisor restart (§5). The nodeAdapter keeps serving the old `*node.Node` after a reset of `Options.Node`; the page expects that and polls `GET /node/status` until it answers again (§4).
- **`nodePort`/`nodeAdapter`** (`ports.go:18-56`, `:58-180`): `attestationError()`, `bootReason()`, `sessionParameters()`, `reannounce(ctx)`; forwarders only.
- **`PairOptions`** (`controller.go:565-571`) gains two **struct fields**, so the `ControllerSource` methods keep their signatures: `TermsAndConditions *TermsAcknowledgement{Version, Acknowledgements uint16}` and `Progress func(PairingStep)` (`PairingStep{Number, Sub uint8; Name, Phase string}`, the port's own type). `ControllerEvent` (`controller.go:680-690`) gains `Step *PairingStep`, `UpdateKind string`, `UpdateVersion *uint32`, `UpdateErr string`.
- **`update` events come from the source, not the handler (owner, 2026-11-07).** Today the handler publishes a kind-less `update` after a check and a consent (`handlers_controller.go:1510`, `:1530`); the adapter now observes the manager (§3.3) and the memory controller emits its own, so the handler's two publishes go — the rule `events.go:233-235` already states why a second source hands each event twice. A host with a `ControllerSource` of its own emits them itself (changelog Changed line; neither host has one: go-fabric-server uses the adapter, `controllerui.go:29-30`; go-ha2fabric has no controller).

### 3.2 Handlers and the stream

- `status` (`handlers_node.go:124-190`): `AttestationError` → `attestation_error` when non-nil, `BootReason`, `Erase != nil` → `erase`.
- `sessions` (`handlers_diagnostics.go:76-91`): `parameters` from `SessionParameters()`; `specification_version_hex` as `%08X`.
- `reannounce`: admin gate, `h.node == nil` → `nodeUnready`, else `reannounce(r.Context())`, 204. Route table `handler.go:274-297`.
- `erase`: body decoded as `factoryReset` does (`handlers_node.go:291-306`), word `erase-node`; `h.erase == nil` → 501 `not_implemented` with detail "the host wires no Erase" (the `Reassemble` sibling answers 503 at `:268-270` — 501 is the right code for a capability the host does not have, the U3b precedent `handlers_controller.go:184`); error → 500; logs `fabricui.erase.done`.
- `NodeObservers` (`events.go:244-283`): `OnFabricUpdated` → `fabric_updated`/`fabricEvent(f)`; `OnCommissioningComplete(fabricIndex)` → `commissioning_complete`/`FabricEvent{FabricIndex}`; `OnEvent(e)` → `cluster_event`/`eventRecordOf(&e)` (`node_detail.go:313`). `OnEvent` runs on the emitter's goroutine (`observers.go:127-136`); push never blocks (the stream drops at 64 behind, as today). Rewrite the comment `events.go:226-228` ("A commissioning is fabric_added …"): `fabric_added` is the AddNOC, `commissioning_complete` the end.
- Controller `commission` (`handlers_controller.go:1273-1290`): `opts.Progress` publishes `commissioning_progress{phase: started, step}` per report; `pairOptionsOf` (`:1234-1244`) maps `terms_and_conditions`. The stream forwarder (`:1647`) maps `Step`, `UpdateKind`, `UpdateVersion`, `UpdateErr`.

### 3.3 Adapter `fabricui/gofabric`

- `Pair` (`gofabric.go:260-283`): on the per-call copy `opts.Progress = func(s controller.PairingStep)` → port step (phase by `commissioner.StepStarted…StepSkipped`, `progress.go:15-18`); the terms through an `Edit` chain that calls the host's `Edit` first, then sets `o.TermsAndConditions = &commissioner.TermsAndConditionsAcknowledgement{Version, Acknowledgements}` — factored as a pure `editFor(host func(*commissioner.Options), tc *fabricui.TermsAcknowledgement) func(*commissioner.Options)` so a unit test holds it. `CommissionDiscovered` (`:429-437`): `copts.Options.Progress` and `.TermsAndConditions` set before `po.Edit`.
- `source` subscribes `Manager.Observe` (`observe.go:130-139`: ordered, one at a time, off the manager's locks) when `s.ctrl.SoftwareUpdates() != nil` (`gofabric.go:508-514`) and emits `ControllerEvent{Type: EventUpdate, Node: e.Peer.NodeID, UpdateKind, UpdateVersion (Target's version for the kinds that carry one), UpdateErr: e.Err text}` through its hub (`:517-523`); the unsubscribe runs in `Hooks.Close` (`hooks.go:72`). No passcode or credential can reach these fields.
- `MemoryController` (`memory_controller.go:225`, `:302`): records the last `PairOptions` (test seam), reports a fixed step sequence through `opts.Progress` (Discovery 2.1, PASE 2.2, Subscribe 21.1 — `progress.go:78-80`, `controller.go:623-631`), emits `update` kinds `check_finished` after a check and `consent_changed` after a consent.
- `fakeNode` (`fabricui/support_test.go:136-181`): `attestationErr`, `bootReason`, `params`, `reannounced int`; `Erase` is an `Options` field, so the test handler takes a func. The e2e web server (`e2eweb_test.go:31-60`) seeds the default node with an attestation error ("node: attestation files: …") and boot reason 5, the second node clean, and wires an `Erase` that flips the fake to zero fabrics.

### 3.4 Tests

- Handler tests on the fakes, one per response: status with and without `attestation_error` (fixtures `node-status.200.json` re-recorded with `boot_reason`, new `node-status.200-attestation.json`), `node-sessions.200.json` re-recorded with `parameters`, `node-mdns-reannounce.204.json`/`.503.json` (+403 asserted without a fixture, as the window routes do), `node-maintenance-erase.204.json`/`.400.json`/`.501.json`, `nodes-…` 404 twins; stream: `fabric_updated`, `commissioning_complete`, `cluster_event` through the fake's observers, order and payload; controller: `terms_and_conditions` reaches the recorded `PairOptions`, step events carry `phase: started` + `step`, `update` events carry `update_kind`; the handler no longer publishes `update` itself. `TestEveryFixtureIsUsedByTheGoTests` and `fixtures.test.ts` hold both sides.
- Adapter: `editFor` unit test (host `Edit` runs first, terms set after); the update observer on a `MemoryController`-free path needs a manager — if `fabrictest` gives the gofabric e2e (`fabricui/gofabric/e2e_test.go`) a controller with `SoftwareUpdates() != nil`, assert one `update` event per check; else record the gap in the report.
- `make e2e` (`fabricui/e2e_test.go:45-64`, `:185-188`): the fabrictest controller's commissioning yields `fabric_added` then `commissioning_complete` on the stream (`s.events`); **Erase**: `Options.Erase: s.device.FactoryReset` with `Options.Node` a `NodeSource` that delegates to `s.device.Node()` (the host pattern of §5) — after 204, `GET /node/status` answers with `fabric_count` 0 and the controller's node is gone. If `fabrictest` lets a device's node state `Commissioning.TermsAndConditions` (`node/config.go:402-411`; check `fabrictest.Options`/`Topology` for a `node.Config` hook — unverified), commission it through the handler with `terms_and_conditions{1,1}` and, without them, assert 502 `node_error` whose detail names the GeneralCommissioning code (`definition_gen.go:110-112`); else say so in the report.
- Playwright (`ui/e2e/smoke.spec.ts:20-48`): SMOKES gain "node status attestation" (`node-attestation` testid), "node diagnostics reannounce" (the button), "node fabrics erase" (the button, shown because `erase` is true), "controller commission terms" (the disclosure); flows: reannounce → toast; erase → two confirmations → toast; commission with terms on the memory controller → step list visible.

## 4 Svelte half

- `Status.svelte` (`:53-94`): destructive Badge + text for `attestation_error` (testid `node-attestation`); boot reason line with labels `node.status.boot_reason.0…6` (the seven spec values, `definition_gen.go:77-83` — page text, no cluster logic; the handlers cannot import `cluster/spec/*`, `internal/arch` `:30-64`).
- `Diagnostics.svelte`: discovery section (`:129-150`) gets "Re-announce" (admin, `outline`, toast); sessions section (`:157-175`) a collapsed "Session parameters" block.
- `Fabrics.svelte` (`:59-69`, `:157`): "Erase node" `outline-destructive` beside the reset button, only with `status.erase`; two confirmations naming the label and the fabric count (`NodeManage.svelte:62-75`), word `erase-node` in the request; after 204 a persistent notice "erased — waiting for the node" and status polling until 200 (or the host's restart).
- `store.svelte.ts`: `fabric_updated` → reload fabrics; `commissioning_complete` → toast + reload status and fabrics; `cluster_event` → the endpoint page's event list (`Endpoint.svelte`) prepends when the endpoint matches, else ignored.
- `controller-store.svelte.ts` (`:130-131`): `step` collected into `progress.steps[]`; `Commission.svelte` (`:109`) renders the list under the phase; `update_kind` done/failed → toast and reload of the updates tab; `Commission.svelte` (`:52`, `:125`): a `<details>` "Terms and conditions acknowledgement" with two integer fields, sent only when both are set, hint: what chip-tool's `--tc-acknowledgements` flags send.
- `en.json`, `de.json`; `make web` in the same commit.

## 5 Hosts' obligations

- **go-ha2fabric** (main `34075a6`, U6b): nothing to compile — the configured node is a `*node.Node` (`cmd/ha2fabric/main.go:109` `Node: def.n`, `nodes.go:31`), each registry node too (`nodes.go:190` `e.Source = m.s.n`; `Registry: fl, NodeConfig: nodeConfig(cfg)`, `main.go:108-111`). `Erase` is a decision for its owner, for the configured node (`Options.Erase`) and the registry's (`NodeEntry.Erase`, `fleet.entry` `nodes.go:178-196`): the fleet's Delete today leaves every fabric and removes the files without `FactoryReset` (`nodes.go:270-301`, `leave` `:304-319`); an Erase would be `node.FactoryReset`, then `serve` again on the erased store (`nodes.go:45`, `fleet.start` `:153`) or the daemon's exit for the supervisor restart — or nothing (`erase` false, no button). Follow-up issue there, referencing this wave.
- **go-fabric-server** (main `b947c93`): nothing to compile (`internal/daemon/nodes.go:98`, `web.go:39`, `registry.go:202`; the adapter at `controllerui.go:30`; the registry wired at `web.go:42`). Registry nodes: wire `NodeEntry.Erase` on the Delete path's `FactoryReset` (`registry.go:303`) plus a restart of the half and `publish()` (`registry.go:196-207` — the handler re-adapts a changed `Source`, `fabricui/registry.go:44-49`); default node (`web.go:39-40`) as go-ha2fabric. Follow-up issue there.

## 6 Docs and changelog

`docs/contract.md`: route table (`:15-46`: reannounce, erase, status and sessions rows), stream (`:183-218`: three events; the `fabric_added` sentence), controller events (`:486-500`: `step`, `update_*`), Problems unchanged, Versioning note 0.7.0. `docs/embedding.md`: the four `NodeSource` methods with the wrapper paragraph (`:85-93` pattern), `Options.Erase`/`NodeEntry.Erase` with the fabrictest reference, `PairOptions` fields, `update` events now the source's. `docs/provider.md`: no change (the exposure provider is untouched) — say so in the report. `changelog.md` `[0.7.0]` Added (routes, fields, events) / Changed (Breaking: `NodeSource` methods; `update` events from the source). `README.md` feature list one line. `docs/reports/U7.md` in the fixed shape (U3b: Summary, Per section with commits, Guards, Differences from the brief, Decisions taken, Open questions, Left out) plus the §1 verification table.

## 7 Commit order (each green under `make ci`)

1. `contract: 0.7.0 — reannounce, erase, status and session fields, three node events, terms and step (U7)` — `openapi.yaml`, `contract/previous` = 0.6.1's contract, `make generate`, `make gen-types`; no fixture yet. The controller side changes schemas only (no new operation), so the generated strict server still compiles; the node routes are hand-written and land in commit 2 — until then the contract names two routes the handler does not serve, which no guard refuses (`contract-check` compares against `previous`, the fixture tests cover fixtures that exist). If `make ci` still fails on commit 1 alone, fold 1 and 2 and say so.
2. `fabricui: ports, handlers, stream, fixtures (U7)` — §3.1, §3.2, `fakeNode`, `MemoryController`, fixtures, handler tests.
3. `gofabric: progress, terms, update observer (U7)` — §3.3, `editFor` test, e2e assertions.
4. `ui: status, diagnostics, fabrics, commission, stores, i18n (U7)` — §4 with `make web`; Playwright smokes.
5. `docs: contract, embedding, changelog, README (U7)`.
6. `report: U7 (U7)`.

## 8 Guards

Run: `make ci`, `make race`, `make e2e`, `make e2e-web`, `make contract-check`, `make generate-check`, `make cover-check`, `make version-check` (the version moves to 0.7.0). Skip with the reason in the report: `make notices-check` (no npm package added or changed — run it if `ui/package.json` moved), the release dry run (no tag in this wave).

## 9 The architecture check

- **Better than the comparison targets.** matterjs-server at `0bd06f2` is a controller: its stream has `node_added`/`node_updated` (`packages/ws-controller/src/server/WebSocketControllerHandler.ts:455-472`), its `commission_with_code`/`commission_on_network` commands (`:960-963`) report no step and take no TC acknowledgement (grep over `packages/ws-controller/src`, zero hits); it has no device-side pages at all, so the node stream and the erase have no comparison there. chip-tool offers the TC acknowledgement as two flags (`PairingCommand.h:274-278`) — the UI offers the same two values, in the one dialog, and nothing it cannot verify. Step-level pairing progress under matter.js's step names (`controller.go:623-631`) is what no UI of the comparison set shows.
- **Better than before.** The stream names the end of a commissioning instead of its AddNOC; the status shows the two conditions an operator could not see (a refused attestation chain, the boot reason); `update` has one source; a wrapper host's obligation is written down with a decision instead of an open question.
- **The right cut.** go-fabric keeps every Matter rule (what a reset erases, when TC is sent); the UI reads the facade and adds no interpretation (the boot-reason labels are page text over a fixed spec enum); the host owns the node's lifecycle (`Erase`), as it owns `Reassemble`; go-fabric-server wires, it does not design. The one thing the UI needs and the facade lacks — the device's TC requirement — goes to go-fabric (§11), not around it.

## 10 Left out

- Controller-side camera (`Controller.CameraRequestor`, `OnCameraSignal`, `controller/serve.go:276-287`): a concept of its own (owner, 2026-11-07).
- Groups (`Node.Groups` `node/node.go:1157`, `Controller.Groups`/`SetGroupKey` `controller/groups.go:83-89`): a concept of its own (owner, 2026-11-07).
- The TC retry flow (problem code, dialog from the device's requirement): waits on SukramJ/go-fabric#449 (§11).
- The map's other "no" rows (ICD, OTA provider status, Peers/Bindings, CommissionerControl, `Controller.VendorInfo`, …): not in the owner's decision; a later inventory wave.

## 11 Decided (owner, 2026-11-07)

Each the recommendation; the alternatives are closed:

1. **Erase is its own route** `POST /node/maintenance/erase`, confirmation word `erase-node`, 501 when no `Erase` is wired (§2, §3.2); the factory-reset route stays as it is.
2. **`NodeSource` grows** by the four methods (§3.1); no optional interfaces on the source.
3. **`update` events come from `softwareupdate.Manager.Observe`** through the adapter (§3.3); the handler's own publishes go (§3.1).
4. **Item 5a stays in U7** (the `terms_and_conditions` field and the dialog, §2, §4); the TC-requirement read is the go-fabric hand-over below.

Hand-over to go-fabric (per "Work that starts in a host", go-fabric `CLAUDE.md:288-304` at v0.18.0), **filed 2026-11-07 as SukramJ/go-fabric#449** (https://github.com/SukramJ/go-fabric/issues/449) with the text below, every `file:line` re-read at the pins; the UI's follow-up (the retry flow, §10) references it:

> **Commissioner: the device's Terms and Conditions requirement as a typed result.**
> **Why.** A controller UI must offer the acknowledgement when a device requires it. The facade offers the acknowledgement (`commissioner.Options.TermsAndConditions`, `controller/commissioner/commissioner.go:154-163`, `:662-665`; step 8.3 `flow.go:710-742`; from `controller.Pair` only through `PairOptions.Edit`, `controller/controller.go:619-620`, "a seam") but no way to learn that a device requires one or which: the flow reads none of GeneralCommissioning's TC attributes (ids `cluster/spec/generalcommissioning/definition_gen.go:35-39`), and the device's refusal is `ErrCommissioning` with a message (`flow.go:409-419`), its code (`CommissioningErrorRequiredTcNotAccepted` 5, `TcAcknowledgementsNotReceived` 6, `TcMinVersionNotMet` 7, `definition_gen.go:110-112`) not typed.
> **Sources at the pins.** connectedhomeip `9fb3392c307` (go-fabric's citations are at `6170af84` — re-read there): the acknowledgement is a parameter, skipped without it (`src/controller/CHIPDeviceController.cpp:3996-4006`; `AutoCommissioner.cpp:497-499`); no read of TCMinRequiredVersion/TCAcknowledgementsRequired in the commissioning sources (`src/controller/*.cpp`, `*.h`: zero hits; the names appear only in the generated cluster bindings under `src/controller/data_model`, `java`, `python`) — so this is ahead of chip and of matter.js (which serves no TC, `packages/node/src/behaviors/general-commissioning/GeneralCommissioningServer.ts:31` at `862ad75f`, as `node/config.go:406-407` cites). The device side wires its terms into GeneralCommissioning at `node/root.go:663-670` (the TC server `cluster/core/general_commissioning_tc.go`; register row `BD-chip-GeneralCommissioning-TermsAndConditions`, `cluster/core/ahead.go:47`).
> **Scope and facade shape.** (a) `controller.PairOptions.TermsAndConditions *commissioner.TermsAndConditionsAcknowledgement` as a first-class field (the seam stays). (b) In the read-commissioning-info step, read TCMinRequiredVersion, TCAcknowledgementsRequired, TCUpdateDeadline when the root's GeneralCommissioning FeatureMap has TC. (c) A typed `*commissioner.TermsRequiredError{Version, Acknowledgements uint16; Deadline *uint32}` returned before AddNOC when the device requires terms and the options carry none or a lower version — fail early, with the requirement, instead of the device's refusal at CommissioningComplete. (d) A typed `*commissioner.GeneralCommissioningError{Code}` wrapping `ErrCommissioning` for every refused GeneralCommissioning command.
> **Parity tests to port.** The CGEN TC cases the harness already runs against `--tc-required` (`CHANGELOG.md` 0.18.0, `examples/reference-bridge/main.go:142`, `surface.go:284-288`; `internal/chiptool` family `CGENTC`, `families_table_test.go:455`): one case where the runner omits the acknowledgement and asserts the typed error with version 1, bit 1.
> **`by_design.md` row.** `BD-Matter-Commissioner-TermsRequirementRead`: the commissioner reads the device's TC requirement and fails early with it; chip and matter.js do not.
> **Acceptance.** `go test ./controller/commissioner/...` holds the typed errors; the reference bridge with `--tc-required` is refused early with `TermsRequiredError{1, 1}` and commissioned with `TermsAndConditions{1, 1}`; `CHANGELOG.md` and `docs/` name the field.
> **Stays host-side (go-fabric-ui).** The request field and the dialog (U7, 5a); the retry flow: problem code `terms_required` (409) carrying `{version, acknowledgements, deadline}` from the typed error, the dialog pre-filled, a second `POST /controller/commission` with the acknowledgement; the terms text itself (the vendor's DCL entry) is outside both modules.
> **Open.** Whether the requirement read belongs in step 1.x (info) or as its own step with a matter.js-style name; whether `Deadline` is rendered (Matter epoch seconds).

## Hand-back

Per §: what landed (files, routes, components, tests by name); the §1
verification table with the report's finding on item 4; guards table
(guard, HEAD, exit); bundle sizes; differences from this brief; decisions;
left out; open questions; the go-fabric issue (SukramJ/go-fabric#449, filed) and the two host
follow-up texts (§5), ready to file; every go-fabric line cited, re-read at
v0.18.0. Open the pull request against main (English, hand-back as body,
no labels, auto-merge off) and hand back with its number.
