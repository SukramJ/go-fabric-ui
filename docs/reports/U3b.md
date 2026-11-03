# Report U3b — the topology view, controller family (go-fabric-ui v0.4.1)

Brief: `notes/briefs/2026-11-03-U3b-topology-controller.md`. Concept:
`docs/concept-u3-topology.md` (§C.2, §C.3, §D). Branch
`u3b/topology-controller`, based on `73afa9d` (main `ba04b8b`). Built
against go-fabric main `c91e5b7f` (pseudo-version
`v0.16.1-0.20261103130909-c91e5b7f4a2f`, no replace); go-fabric and
go-fabric-server read, not changed; `contract/previous/` unchanged.

## Summary

Contract 0.3.1, additive over 0.3.0: `GET /controller/topology` is served,
and `TopologyEdge.strength` gains `none`. `gofabric.ControllerOptions.Topology`
takes the host's `*topology.Service`. The adapter maps its snapshot into the
port's graph and publishes `topology_changed` on every change. The nodes page
gets the tabs **List** and **Network**. The kit's `Graph` draws a link's
strength. No new dependency.

## Per section: what landed

### §1 Contract 0.3.1 — `ac2fc9f`

- `info.version` 0.3.1. `/controller/topology` has no `x-status: planned`.
  Its description states the mapping: ids, kinds, edges, `refresh`, and
  501 for a host without a topology service. 200/403/501/502 stay.
- `strength` enum `[strong, medium, weak, none, unknown]`, with `none`
  meaning observed dead (LQI 0).
- `internal/api/api.gen.go` (`TopologyEdgeStrengthNone`) and
  `ui/src/lib/api/schema.gen.ts` are regenerated.

### §2 Go module — `3670d07`

- go.mod: go-fabric at `c91e5b7f` (`go mod tidy` drops the indirect
  `cenkalti/backoff`). The `internal/arch` allowlist gains
  `controller/topology`.
- `fabricui/gofabric/topology.go`:
  - `NetworkTopology`: no service gives `ErrNotImplemented`. `refresh` calls
    `svc.Refresh`, whose error is wrapped as `ErrNodeFailed` (502); without
    it the adapter calls `svc.Topology()`.
  - `networkTopologyOf` is a pure function that maps the snapshot as brief
    §2.2 says. The root is labelled by the fabric label, else `Controller`.
    Peers are named as the Nodes page names them (`NodeLabel`, else
    `ProductName`, else `0x<hex>`). Link edges carry the strength and the
    LQI/RSSI of the direction that strength comes from. The label shows
    both directions when they differ.
  - `watchTopology` subscribes `OnChanged`, which publishes
    `topology_changed`. `closeTopology` releases the subscription, and
    `Hooks.Close` calls it. If the hooks were already closed when
    `ControllerOf` binds them, the subscription is released at once.
- `gofabric.go`: the `ControllerOptions.Topology` doc, and the package
  doc's wiring (`changes.Install(&cfg)`, `topology.New`, `Topology: svc`,
  `svc.Close()`).
- `MemoryController.Topology *NetworkTopology`: when it is set, the
  in-memory controller serves that graph; otherwise it answers 501. The
  port docs name `none`.
- Fixtures `controller-topology.200` and `.403`, recorded with `-update`
  and reviewed. Tests:
  - handler fixtures 200/403/501
  - the refresh flag reaching the source, and 502
  - `TopologyEdge` `none`
  - the adapter literal, served through `fabricui.Handler`, equals the
    200 fixture. This is the end-to-end check: the handler test's port
    literal must give the same fixture.
  - `linkOf` direction and label rules
  - the subscription release
  - e2e: `Changes.Install` → `controller.New` → `topology.New` →
    `ControllerOf(…, Topology: svc)`. After commissioning, the graph has
    `node`, `peer:<id>` and a `session` edge between them. `svc.Close()`
    runs on cleanup.

### §3 npm package — `6645f05`

- `Nodes.svelte` gets the tabs `list` and `network`, with the prop `tab`.
  Tabs link to `routes.controllerNodes` and the new
  `routes.controllerNetwork` (`#/controller/nodes?tab=network`).
  `shell/route.ts` reads `?tab=` on every controller page, and `App.svelte`
  passes it.
- `Network.svelte` (exported as `ControllerNetwork`):
  - loads the graph, with loading, error and empty states
  - shows 501 as the `controller.network.not_served` empty state, without
    the Refresh button
  - "Read diagnostics again" sends `refresh=true` with the token. A refusal
    shows a toast and keeps the graph.
  - only a `peer` links to its page
  - reloads when the controller store's `graphEpoch` moves
- `controller-store.svelte.ts`: `graphEpoch` and `scheduleGraphReload`
  (300 ms). `topology_changed` bumps the epoch, and `close` clears the
  timer.
- `Graph.svelte`: a link's stroke width is 3/2/1 for strong/medium/weak;
  `none` is 1, dotted (`1 4`) and muted; `unknown` is 1 and plain; a
  session keeps `5 4`. Each link has a `<title>` with strength, LQI, RSSI
  and label. `graph-layout.ts` exports `TOPOLOGY_STRENGTHS` and its layout
  is unchanged: `columnsOf` gives root 0, peers 1, and neighbours and APs 2.
- en/de: `controller.nodes.tab.*`, `controller.nodes.tabs`,
  `controller.network.*`, `graph.strength.*` and twelve `graph.detail.*`
  keys. `catalogs.test.ts` walks both graph fixtures' kinds, states, detail
  keys and strengths, plus the strength enum.
- Tests:
  - client 200/403/501
  - layout golden on `controller-topology.200`
  - `Graph` strength styling
  - the store epoch
  - the route
  - the Network page: graph, links, reload, refresh refused, 501, empty,
    error
  - the Nodes tabs
- The bundle is rebuilt (`make web`).

### §4 Docs — `559b52f` and the report commit

`docs/contract.md` (the controller graph section, the problem, the route
row), `docs/embedding.md` (the facade row with the host wiring),
`docs/concept-u3-topology.md` (`none`, the "landed" note), `changelog.md`
`[0.4.1]`, `internal/version` and `ui/package.json` + lock at 0.4.1,
`README.md`, `ui/README.md`, `CONCEPT.md` and `CLAUDE.md` state lines.
This report.

## Guards

The guards ran one at a time under `flock` at HEAD `559b52f` (HEAD before
and after the run); the summary is in `~/.cache/go-fabric-scratch/u3b/summary.log`.

| Guard | HEAD | Exit |
| --- | --- | --- |
| `make ci` (build vet test lint generate-check web-check dist-check) | `559b52f` | 0 |
| `make race` | `559b52f` | 0 |
| `make e2e` | `559b52f` | 0 |
| `make contract-check` | `559b52f` | 0 |
| `make version-check` (0.4.1) | `559b52f` | 0 |
| `make notices-check` | `559b52f` | 0 |
| `make generate-check` | `559b52f` | 0 |
| `make cover-check` (`fabricui` 94.0 %, `fabricui/gofabric` 63.5 %) | `559b52f` | 0 |

The report commit touches only this file. The guards were not re-run for it.

## Differences from the brief

1. **Two `unknown_` neighbours, not one.** Five strengths need five links,
   and three Thread nodes plus one Wi-Fi link give only four. The second
   neighbour also covers both mappings of an unknown neighbour (router and
   end device).
2. **The fixture's `none` link is written by hand.** go-fabric's `Build`
   never yields a `Connection` of strength `none`:
   - `connectionOf` drops a pair whose directions are all dead
     (`controller/topology/thread.go:166-183`).
   - A Wi-Fi link is never `none` (`build.go:141`, `rssiStrength`).

   `none` occurs only on a `Direction`. The adapter test's literal says so
   (open question 1).
3. **Refresh is not hidden from non-admins.** The pages have no admin
   signal: with host-decided rights nothing tells the page. The shell's
   other admin controls (decommission, interview, …) are not hidden
   either. A 403 shows as a toast and keeps the graph. The button is
   hidden on 501.
4. **The 501 text.** The page's empty state is the catalogue's "The host
   serves no network graph." (`controller.network.not_served`). The
   problem's detail stays the generic one shared by every
   `ErrNotImplemented` (open question 3).
5. **Commit split.** The fixtures are in the §2 commit, because they are
   recorded from the handler tests with `-update`. Their TypeScript use
   comes in §3, so at `3670d07` alone the TS fixture-usage check is red.
6. **No `lib/routes`.** The route keys live in `Routes`
   (`lib/context.ts`). `controllerNetwork` is added there. Hosts pass
   `Partial<Routes>`, so the addition is non-breaking.
7. **The e2e case sits in `TestEndToEnd_Controller`.** The topology
   service is wired in the shared `newCtrlSite`, so every controller e2e
   runs with it. The check replaces the former 501 assertion.
8. **Changelog `[0.4.0]`.** Its `### Planned` entry goes, as the brief
   says. One `Added` line keeps the history: the route was declared, then
   planned, in 0.4.0.

## Decisions taken

- **LQI only on Thread links.** A Wi-Fi `Direction.LQI` is 0 by
  definition. A Wi-Fi edge carries RSSI only, and its label shows RSSI.
- **Direction choice.** Source-to-target is taken when its strength
  matches the connection's, else target-to-source when that one matches,
  else whichever exists. An empty strength is drawn as `unknown`
  (defensive; `Build` never yields one).
- **Peer `role`** is the Thread or Wi-Fi role, else the network type
  (`ethernet`, `unknown`). The `details[].role` appears only when go-fabric
  has a role.
- **Subscription lifetime.** The subscription is made in `ControllerOf`
  before the hooks bind, and released by `Hooks.Close` or the unexported
  `closeTopology`. The port has no `Close`.
- **Extra key** `controller.nodes.tabs`: the tab strip's aria label, as
  `controller.node.tabs`.

## Open questions

1. **`none` on an edge.** go-fabric's `Build` never yields it (difference
   2). Should go-fabric keep dead pairs as `none` connections, as
   matterjs-server's dashboard type admits? Or should the contract carry
   each direction's strength instead? Today the enum value is reachable
   only from a hand-built `Topology`.
2. **An admin signal for the pages.** A status field telling the page
   whether the viewer has admin rights would let Refresh and the other
   admin controls hide, rather than answer 403.
3. **The 501 detail** "the host's controller does not serve this route
   yet": should "yet" go, now that 501 means a host without a topology
   service? Changing it changes the 501 fixture's body only.
4. **go-fabric-server's S-wave** has to wire `topology.Changes.Install`
   between `pageHooks.Wrap` and `controller.New` (`run.go:246-252`). It
   must also close the service before `ctrl.Close`, and pass `Topology:` in
   `controllerSource` (`controllerui.go:25-33`).

## Left out

- go-fabric-server's S-wave, the border-router registry, Playwright (U5),
  U4, the release tag (the owner's).
- No go-fabric change, no change under `contract/previous/`, no new npm
  package.
