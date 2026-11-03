# Brief U3b — the topology view, controller family (`v0.4.1`, contract 0.3.1)

Date 2026-11-03. Concept: `docs/concept-u3-topology.md` (German; §C.2
:115-127, §C.3 :142-145, §D :162-167 describe this wave; §F carries the
owner's decisions). Fact map: the launcher's
`~/.cache/go-fabric-scratch/read-u3b.md` (go-fabric at the merge commit of
pull request #384 = `c91e5b7f`, go-fabric-ui `v0.4.0` untagged at
ba04b8b0, go-fabric-server `v0.4.9`); every `file:line` there is a map —
verify what you build on. One wave, one pull request, one hand-back.
**Model:** Opus. Conventional commits, English, no attribution lines, no
auto-merge, no label. Before every commit run `git var GIT_AUTHOR_IDENT` and
expect `SukramJ <sukramj@icloud.com>`; commit with
`-c user.name=SukramJ -c user.email=sukramj@icloud.com`. go-fabric's token
rules (`/home/markus/Dokumente/GitHub/go-fabric/CLAUDE.md` §"Token use"):
read slices, guards in the background with a summary log
(`~/.cache/go-fabric-scratch/u3b/summary.log`, HEAD before and after, one
`exit` line per guard), a short report `docs/reports/U3b.md` (per §, guards
with HEAD, differences, decisions, questions). This repository's own rules
are `CLAUDE.md` here: Go imports only go-fabric's facade (`fabricui/gofabric`
is the one adapter and may import `controller/topology` from this wave on),
the bundle is committed, fixtures on both sides, 64-bit ids as hex strings.
Worktree `../go-fabric-ui-u3b`, branch `u3b/topology-controller`.

## 0 Decisions (owner 2026-11-03 in concept §F and U3a §0; launcher 2026-11-03)

- **0.1** (owner) One contract type `TopologyGraph` for both families; the
  controller graph is built in go-fabric `controller/topology` (landed in
  pull request #384) and mapped into the port's own types by the adapter.
  The contract schema of U3a stays; this wave adds one enum value.
- **0.2** (owner) `v0.4.1`; the contract goes 0.3.0 → 0.3.1 (an addition:
  `TopologyEdge.strength` gains `none`; `/controller/topology` loses
  `x-status: planned`). The release tag is the owner's.
- **0.3** (launcher) The `topology.Service` is the **host's**:
  `topology.Changes.Install(&cfg)` must run before `controller.New`, like
  `gofabric.Hooks.Wrap`, and someone must `Close` it — both are the host's
  wiring. The adapter takes it as `ControllerOptions.Topology
  *topology.Service`; nil keeps `ErrNotImplemented` (501), which is what
  go-fabric-server serves until its S-wave.
- **0.4** (launcher) The graph has the controller as its root `node` with a
  `session` edge to every commissioned node (contract text
  `openapi.yaml:903-909`: "nodes are the controller and its commissioned
  nodes"); matterjs-server draws no controller, our contract does. Links
  between nodes are `link` edges from `topology.Connection`.
- **0.5** (launcher) No change under `contract/previous/`. The 501 response
  stays in the contract for a host without a topology service.

## 1 Contract 0.3.1 (`contract/openapi.yaml`)

- `info.version` 0.3.1 (:4). `/controller/topology` (:897-923): drop
  `x-status: planned` (:901); description (:903-909) now states the
  mapping of §2.2 in two sentences (ids, kinds, edges, `refresh`), and that
  501 `not_implemented` is a host without a topology service. 200/403/501/502
  stay.
- `TopologyEdge.strength` (:1170) enum `[strong, medium, weak, none, unknown]`
  with `none` described as "link: observed dead (LQI 0)" — matterjs-server's
  `TopologyStrength` (`packages/ws-client/src/models/model.ts:258` at
  0bd06f22), carried by `topology.StrengthNone`.
- Fixtures `contract/fixtures/controller-topology.200.json` (the graph of
  §2.3's test: controller, a Thread router peer, a Thread end-device peer,
  one `unknown_…` neighbour, a Wi-Fi peer with its `ap_…`, an Ethernet
  peer; every strength value including `none` on one edge) and
  `controller-topology.403.json`; delete nothing. `make generate` and
  `make gen-types` regenerate `internal/api/api.gen.go` (strength consts
  :231-236) and `ui/src/lib/api/schema.gen.ts` (:1149 union); the
  contract guard (`ui/scripts/contract-guard.mjs`) must stay green against
  `contract/previous/` (0.2.0).

## 2 Go module (`fabricui/`, `fabricui/gofabric/`)

- **2.1 go.mod**: `go get github.com/SukramJ/go-fabric@c91e5b7f` (a
  pseudo-version of go-fabric main; no go-fabric tag before the owner's
  ecosystem test), `go mod tidy`; `internal/arch/arch_test.go:51-65`
  allowlist gains `github.com/SukramJ/go-fabric/controller/topology`.
- **2.2 Adapter** `fabricui/gofabric/topology.go` (new) + `gofabric.go`:
  `ControllerOptions.Topology *topology.Service` (:52-73, doc after 0.3; the
  package doc's wiring example :9-16 gains `changes.Install(&cfg)` before
  `controller.New` and `Topology: svc`). `NetworkTopology(ctx, refresh)`
  (:507-513): nil service → `ErrNotImplemented`; `refresh` →
  `svc.Refresh(ctx)` (its error → `ErrNodeFailed`-class as the handler maps
  it to 502, `handlers_controller.go:179-181`), else `svc.Topology()`; then
  `networkTopologyOf(t topology.Topology, ...)` — a pure function, tested
  on a `topology.Topology` literal:
  - root: `NetworkNode{ID: "node", Kind: "node", Label: <fabric label as
    Fabric() reports it, else "Controller">, State: "ok"}`; details:
    `nodes` (count), `collected_at`.
  - a `topology.KindMatter` node → `ID "peer:" + hex16(NodeID)`, `Kind
    "peer"`, `Label` as `Nodes()` names the node (the same label the
    controller Nodes page shows — read `gofabric.go:155-190` for its
    source; fall back to the hex id), `State` ok / unreachable from
    `Available`, `NodeID`, `Role` = `Role` when set else `NetworkType`,
    `RLOC16` = `Rloc16`, `SSID`, `BSSID`; details in this order, each only
    when set: `network` (NetworkType), `role`, `rloc16` (`0x%04x`),
    `ext_address` (16 hex), `ext_pan_id` (16 hex), `network_name`, `ssid`,
    `bssid`, `bridge` ("yes" when IsBridge), `addresses` (joined by ", ").
  - `KindThreadUnknown` → `ID` as given (`unknown_…`), `Kind
    "thread_router"` when `Role == topology.RoleRouter`
    (`controller/topology/build.go:114-116`: a neighbour seen
    rx-on-when-idle), else `"thread_end_device"`; `Label` = the 16-hex
    ext address; `State "unknown"`; details `ext_address`, `ext_pan_id`,
    `role`.
  - `KindWiFiAP` → `ID` as given (`ap_…`), `Kind "wifi_ap"`, `Label` =
    SSID when the connected stations name one, else the BSSID; `State
    "unknown"`; details `bssid`, `ssid`.
  - the contract's `ethernet` kind is not produced: an Ethernet peer is a
    `peer` with role `ethernet` (say so in the contract description).
  - edges: one `{Source: "node", Target: peer id, Kind: "session"}` per
    commissioned node (no strength); one `{Kind: "link"}` per
    `topology.Connection` with `Source`/`Target` mapped through the same id
    rule, `Strength` = `Connection.Strength` (including `none`), `LQI`/`RSSI`
    from the direction whose strength equals the connection's (matterjs-
    server summarises the strongest, `NetworkTopologyService.ts:486-501` at
    0bd06f22), `Label` "LQI a/b" or "RSSI a/b dBm" when both directions
    report and differ.
  - `OnChanged`: on bind (`ControllerOf`, when `opts.Topology != nil`)
    subscribe `svc.OnChanged(func(topology.Topology) { s.publish(
    fabricui.ControllerEvent{Type: fabricui.EventTopologyChanged}) })`;
    the unsubscribe runs in `Hooks.Close` (`hooks.go:68-78`) when hooks are
    bound, else is kept on the source and released by a new unexported
    close the tests call — do not add `Close` to the port. The callback
    must not call `Refresh` (service rule).
- **2.3 Handler**: `handlers_controller.go:1383-1426` `networkGraph` stays;
  `TopologyEdgeStrength` cast now admits `none`. Tests: replace
  `TestControllerTopology_NotImplemented` (`handlers_topology_test.go:162-184`)
  by the 200 path against the fixture (a `topologySource` returning the
  §2.2 graph — the same literal as the adapter test's expected output, so
  the fixture is checked end to end), keep a 501 case with
  `MemoryController`, add the 403 case against `controller-topology.403.json`,
  keep `TestControllerEvent_TopologyChanged`. Adapter test in
  `fabricui/gofabric/topology_test.go` on a `topology.Topology` literal and,
  in `e2e_test.go` (fabrictest site, :45-62), one case: `Changes.Install`
  + `topology.New` + `ControllerOf(..., Topology: svc)`; commission a node;
  `NetworkTopology(ctx, false)` holds `node` and the peer with a `session`
  edge; `svc.Close()` at the end. `MemoryController.NetworkTopology`
  (`memory_controller.go:343-348`) returns the fixture graph when a new
  field `Topology *NetworkTopology` is set, else `ErrNotImplemented`.

## 3 npm package (`ui/`)

- **3.1 Page**: `pages/controller/Nodes.svelte` gets page tabs `list` /
  `network` (`?tab=network`, the per-node `Tabs` pattern
  `pages/controller/Node.svelte:12,25,68`; route helper in `lib/routes`);
  the network tab is `pages/controller/Network.svelte` (new, exported in
  `pages/index.ts`): `client.controllerTopology()` on load, admin-only
  "Refresh" button (`controllerTopology(true)`; hidden when not admin as
  the shell's other admin controls), `LoadingState`/`ErrorState`/
  `EmptyState`, `Graph` with `hrefOf` → `routes.controllerNode(node_id_hex)`
  for `peer` nodes (`ref.page === "controller_node"`), 501 shown as the
  `EmptyState` "the host serves no network graph" (the fixture's text).
- **3.2 Store**: `lib/controller-store.svelte.ts` gains `graphEpoch` +
  `scheduleGraphReload` (300 ms, as `store.svelte.ts:129-135`) and
  `applyEvent` `case "topology_changed"` (:125-128) bumps it; the page
  reloads on it exactly as `pages/node/Topology.svelte:44-47`.
- **3.3 Kit**: `Graph.svelte` draws a `link` edge's `strength`: stroke width
  3/2/1 for strong/medium/weak, `none` dotted and muted, `unknown` 1 plain;
  `session` keeps its dash; a `title` on the edge with LQI/RSSI/label.
  `graph-layout.ts` unchanged unless the controller graph's columns need it
  (root `node`, peers column 1, APs/unknowns column 2 — check
  `columnsOf` :87 gives that; the golden test gains the
  `controller-topology.200` fixture).
- **3.4 i18n** en/de: `controller.nodes.tab.list|network`,
  `controller.network.{title,hint,empty,not_served,graph_label,refresh}`,
  `graph.strength.{strong,medium,weak,none,unknown}`, `graph.detail.*` for
  every new detail key (`network, role, rloc16, ext_address, ext_pan_id,
  network_name, ssid, bssid, bridge, addresses, nodes, collected_at`).
  `catalogs.test.ts` (:41-46) also walks the `controller-topology.200`
  fixture's kinds, states and detail keys and the strength enum.
- **3.5 Tests**: `controller-client.test.ts` 200 + 403 + 501 against the
  fixtures; `Graph.test.ts` strength styling; the store's epoch.

## 4 Docs

`docs/contract.md` (:95, :171, :229 route row: served, 0.3.1; stream
:267), `docs/embedding.md:125` (the facade-call row: `Topology` option,
host wiring `Changes.Install` → `controller.New` → `topology.New` →
`ControllerOf` → `svc.Close` on shutdown), `docs/concept-u3-topology.md`
§C.1 :94 strength list gains `none` and §D :162-164 gets the "landed"
note; `changelog.md` `[Unreleased]` → `## [0.4.1] - <date>` Added
(`/controller/topology` 200, Network tab, `none`), Changed (contract
0.3.1, go-fabric pseudo-version), and the `### Planned` entry (:33-38)
goes; `internal/version/version.go:11` + `ui/package.json:3` → 0.4.1
(`make version-check`). `docs/reports/U3b.md`.

## 5 Guards

As U3a §5: `make ci` whole (build vet test lint web-check dist-check),
`make race`, `make e2e`, `make contract-check`, `make version-check`,
`make notices-check`, `make generate-check`, all in the background with the
summary log; read a guard's log only on failure, with grep. Run guards one
at a time, under `flock ~/.cache/go-fabric-scratch/guards.lock` (a go-fabric
agent may run its own).

## 6 Out of scope

go-fabric-server's S-wave (its `controllerSource`, `get_network_topology`
on `controller/topology`, go.mod bumps, `web.go:20` comment), the
border-router registry, Playwright (U5), U4. No go-fabric change: a gap in
`controller/topology` is a question in the hand-back, not a local patch.

## Hand-back

The report `docs/reports/U3b.md` (per §, guards with HEAD, differences from
this brief with reasons, decisions, open questions, what was left out), the
pull request without auto-merge, the branch pushed; the hand-back message
is the report's essentials with the PR link.
