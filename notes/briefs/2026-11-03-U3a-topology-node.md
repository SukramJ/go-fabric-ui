# Brief U3a — the topology view, node family (`v0.4.0`, contract 0.3.0)

Date 2026-11-03. Concept: `docs/concept-u3-topology.md` (German; §F carries
the owner's decisions of 2026-11-03). Fact maps: the launcher's
`read-u3-ui-server.md` and `read-u3-facade.md` (go-fabric `v0.16.0`,
go-fabric-ui `v0.3.0`, go-fabric-server `v0.4.9`); every `file:line` there is
a map — verify what you build on. One wave, one pull request, one hand-back.
**Model:** Opus. Conventional commits, English, no attribution lines, no
auto-merge, no label. Before every commit run `git var GIT_AUTHOR_IDENT` and
expect `SukramJ <sukramj@icloud.com>`; commit with
`-c user.name=SukramJ -c user.email=sukramj@icloud.com`. go-fabric's token
rules (`/home/markus/Dokumente/GitHub/go-fabric/CLAUDE.md` §"Token use"):
read slices, guards in the background with a summary log
(`~/.cache/go-fabric-scratch/u3a/summary.log`, HEAD before and after, one
`exit` line per guard), a short report `docs/reports/U3a.md` (per §, guards
with HEAD, differences, decisions, questions). This repository's own rules
are `CLAUDE.md` here: Go imports only go-fabric's facade, the bundle is
committed, fixtures on both sides, 64-bit ids as hex strings.

## 0 Decisions (owner, 2026-11-03)

- **0.1** One contract type `TopologyGraph` for both families, built
  server-side in Go from the facade calls, drawn by one kit component in
  hand-written SVG (no graph library — concept v13 E.4).
- **0.2** The controller-side graph builder lives in go-fabric
  (`controller/topology`, a separate go-fabric wave). This wave declares
  `/controller/topology` in the contract with its schema and
  `x-status: planned` (501), and adds the port method; U3b fills it.
- **0.3** Version `v0.4.0`, contract 0.3.0.
- **0.4** Facade gaps (subscription enumeration, partition/scope, TagList,
  node-wide attribute observer) are a go-fabric issue, not worked around
  here: subscriptions stay counts per session in this wave.

## 1 Contract 0.3.0 (`contract/openapi.yaml`)

- `GET /node/topology` → `TopologyGraph` (200; 503 before the node assembled,
  as `/node/endpoints` does — reuse its problem code).
- `GET /controller/topology?refresh=` → `TopologyGraph`; `x-status: planned`,
  answers 501 until U3b; `refresh=true` is `x-admin`. The schema is complete now.
- `/node/stream` description gains `session_opened`, `session_closed`,
  `subscriptions_changed` with payload `SessionEvent` (the `Session` schema
  of `/node/sessions`, `openapi.yaml:1037-1052`); `/controller/stream` gains
  `topology_changed` (`ControllerEvent` with a new type value).
- Schemas: `TopologyGraph{as_of, family, nodes[], edges[]}`,
  `TopologyNode{id, kind, label, state, ref?, details[], …typed optionals}`,
  `TopologyEdge{source, target, kind, strength?, lqi?, rssi?, label?}`,
  `TopologyRef{page, id}`, `TopologyDetail{key, value}` — the enums and
  fields of concept §C.1. Responses open to unknown fields as the 0.2.9
  guard requires.
- Node ids in the graph are strings: `node`, `ep:<n>`, `fabric:<index>`,
  `session:<id>`, `peer:<node_id_hex>`.
- `contract/previous/` stays 0.2.0; `make contract-check` green. Fixtures:
  `node-topology.200` (a bridge with two devices, one with a part, two
  fabrics, three sessions incl. one PASE), `node-topology.503`,
  `controller-topology.501`. Version 0.3.0 in the contract; `v0.4.0` in
  `internal/version`, `ui/package.json`, lock file, changelog section.
- `make gen-types`, `make generate` (controller stubs), `make generate-check`.

## 2 Go module (`fabricui/`)

- `handlers_topology.go`: `GET /node/topology` from `NodeSource.Topology()`,
  `FabricInfos`, `SessionDiagnostics()` (`fabricui/options.go:27-51`). Graph:
  `node` → `ep:<aggregator>` → devices (`kind: device`, `state` from
  `Reachable`, `details`: device type name + revision, cluster count, key) →
  parts (`kind: part`); `fabric:<i>` (`details`: label, vendor id, node id
  hex, fabric id hex) → `session:<id>` (`kind: session`, `state: idle` when
  `PeerIdle` exceeds `PeerIdleInterval`, `details`: peer node id hex,
  PASE/CASE, subscriptions, idle seconds) → edge `session` to `node`. A PASE
  session (fabric index 0) hangs on `node` without a fabric. Every node gets
  a `ref` to the page that already lists it (`endpoints`, `fabrics`,
  `diagnostics`).
- `events.go` / `stream.go`: the three session events from
  `Observe(node.Observers{OnSessionOpened, OnSessionClosed,
  OnSubscriptionsChanged})` (`node/observers.go:91-100`, CASE only for
  opened — say so in the contract description), payload `SessionEvent`
  mapped as `sessionResponse` does today (`handlers_diagnostics.go`).
- `controller.go`: `ControllerSource.NetworkTopology(ctx context.Context,
  refresh bool) (NetworkTopology, error)` with small own types
  (`NetworkTopology{CollectedAt, Nodes, Connections}` mirroring the graph
  schema, no go-fabric type) and `ControllerEventTopologyChanged`. The
  generated handler answers 501 `not_implemented` (problem code) in this
  wave; `memory_controller.go` and `fabricui/gofabric` return
  `ErrNotImplemented` — no diagnostics reads here (0.2).
- Tests: handler tests on the fixtures (`support_test.go` pattern,
  `TestEveryFixtureIsUsedByTheGoTests` must keep passing), stream test for
  the three events, `e2e_test.go` extended: the fabrictest controller's
  session appears in `/node/topology` as a `session` node under its fabric.
- Coverage floors (`script/coverfloor/floors.txt`) hold or rise.

## 3 npm package (`ui/`)

- `src/lib/kit/Graph.svelte` + `graph-layout.ts`: layered layout as a pure
  function (columns by depth from `node`, rows in input order, cubic edge
  paths, sizes from label length), golden test of positions on the
  `node-topology.200` fixture; `viewBox` pan/zoom (wheel + drag, buttons);
  node shape per `kind`, border colour per `state` from `theme/tokens.css`;
  selection opens a side panel with `details[]` and a link that follows
  `ref` through `shell/route.ts`; keyboard: Tab across nodes, Enter follows
  `ref`, Escape clears; `aria-label` per node.
- `src/lib/pages/node/Topology.svelte`: loads `/node/topology`; subscribes to
  the stream and reloads (debounced like `store.svelte.ts:103-107`) on the
  three session events plus `endpoints_changed`, `fabric_added`,
  `fabric_removed`; `LoadingState`/`EmptyState`/`ErrorState` from the kit.
  Export from `pages/index.ts`; shell menu entry "Topology", route
  `#/node/topology` (`shell/route.ts`, `App.svelte` tabs array); i18n en/de
  with the coverage test.
- `src/lib/api`: regenerate `schema.gen.ts`; client methods `topology()`,
  `controllerTopology(refresh)`; stream parser accepts the new events.
- Tests: page test on the fixtures through `test/Harness.svelte`, graph
  component test, layout golden, stream test, `fixtures.test.ts` covers the
  new fixtures. `make web` after the source change (bundle committed,
  `dist-check`).

## 4 Docs

- `docs/contract.md`: the new routes, events and the planned route;
  `README.md` page list; `CONCEPT.md` wave list (U3a); `changelog.md`
  section `[0.4.0]` (Added: topology page, contract 0.3.0; Planned:
  `/controller/topology`); `docs/reports/U3a.md`.
- `THIRD-PARTY-NOTICES.md` unchanged (no new npm package — if you need one,
  stop and report instead).

## 5 Guards

`make ci` whole (build vet test lint web-check dist-check), `make race`,
`make e2e`, `make contract-check`, `make version-check`, `make notices-check`,
`make generate-check`, all in the background with the summary log; read a
guard's log only on failure, with grep. Run guards one at a time.

## 6 Out of scope

The controller graph builder and `/controller/topology` 200 (U3b after
go-fabric's `controller/topology`), the server wave, Playwright (U5), live
attribute values (U4), per-session subscription paths, partition/scope and
tags (go-fabric issue). No change under `contract/previous/`.

## Hand-back

`docs/reports/U3a.md` per §: what landed with the commit, guards with HEAD
and exit, differences from this brief with the reason, decisions taken,
open questions, what was left out. Pull request against `main` without
auto-merge; report its URL, HEAD and the worktree through SubagentHandback.
