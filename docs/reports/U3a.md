# Report U3a — the topology view, node family (go-fabric-ui v0.4.0)

Brief: `notes/briefs/2026-11-03-U3a-topology-node.md`. Concept:
`docs/concept-u3-topology.md` (§C, §F). Branch `u3/topology`, based on
`f832db1`. Built against go-fabric `v0.16.0` (the go.mod pin, no replace);
go-fabric and go-fabric-server read, not changed.

## Summary

Contract 0.3.0 (additive over 0.2.0): `GET /node/topology` serves the node
as a `TopologyGraph`, built server-side from `Topology`, `FabricInfos` and
`SessionDiagnostics`. The node stream carries `session_opened`,
`session_closed` and `subscriptions_changed`. `GET /controller/topology` is
declared with its final schema, `x-status: planned` and 501
`not_implemented`, and `ControllerSource` gains `NetworkTopology`. The npm
package adds the hand-written SVG `Graph` kit component with a pure, golden-tested
layout, plus the **Topology** page under `#/node/topology`. No new
dependency; `contract/previous/` unchanged.

## Per section: what landed

### §1 Contract 0.3.0 — `7ec0094`

- `GET /node/topology` → `TopologyGraph` (200, 500, 503).
- `GET /controller/topology?refresh=` → `TopologyGraph`, `x-status:
  planned`, 403/501/502; `refresh` carries `x-admin` on the parameter.
- `TopologyGraph`, `TopologyNode` (all kinds and typed optionals of §C.1),
  `TopologyEdge`, `TopologyRef`, `TopologyDetail`, `SessionEvent` (the
  `Session` fields). No schema is closed.
- Stream descriptions: three session events on `/node/stream` (CASE only
  for opened), `topology_changed` on `/controller/stream`. Problem code
  `not_implemented`.
- Fixtures `node-topology.200` (bridge with two devices, one with a part,
  two fabrics, three sessions incl. one PASE), `node-topology.503` and
  `controller-topology.501`, each recorded with `-update` and reviewed.
- Version 0.3.0 in the contract; `v0.4.0` in `internal/version`,
  `ui/package.json`, the lock file and the changelog. `make gen-types`.

### §2 Go module — `b2618ec`

- `handlers_topology.go`: `node` `contains` `ep:<agg>` `contains` the devices
  `contains` their parts. Each `fabric:<i>` `contains` its CASE `session:<id>`,
  and every session has a `session` edge to `node`. A PASE session, or one on
  an unlisted fabric, hangs on `node` alone. The details are the ones the
  brief names, and every node has a `ref`.
- `events.go`: `EventSessionOpened`, `EventSessionClosed` and
  `EventSubscriptionsChanged` with an exported `SessionEvent`.
  `NodeObservers` yields them, and so the stream does too. `sessionOf` is
  shared with `/node/sessions`.
- `controller.go`: `ControllerSource.NetworkTopology(ctx, refresh)`, the
  port's own types `NetworkTopology{CollectedAt, Nodes, Connections}`,
  `NetworkNode`, `NetworkConnection` and `TopologyDetail`, plus
  `EventTopologyChanged` and `ErrNotImplemented`. The memory controller and
  `fabricui/gofabric` return `ErrNotImplemented`, and `controllerError` maps
  it to 501.
- `controllerGate` refuses `refresh=true` without admin rights
  (`controllerAdminQuery`), and a test holds it against the contract.
- `internal/api` is regenerated (`make generate`).
- Tests:
  - handler fixtures
  - 503 before assembly and 500 on a fabric error
  - graph edge cases
  - the stream session events
  - the 501, 403 and 200 mapping of the controller route
  - `topology_changed` payload
  - e2e: the commissioner's CASE session is under its fabric in
    `/node/topology`, and `session_opened` reached the stream. The
    controller e2e answers 501.
- Coverage: `fabricui` 94.0 % (floor 88), `fabricui/gofabric` 56.2 % (floor 50).

### §3 npm package — `26751d3`

- `kit/graph-layout.ts` is a pure function:
  - Columns are the signed depth from `node`: following an edge from its source
    goes right, from its target goes left. So the bridge lies right of the node
    and the sessions and fabrics left of it.
  - Rows keep input order, edges are cubic paths, and node sizes come from the
    label length.
  - A golden test fixes every position of `node-topology.200`.
- `kit/Graph.svelte`:
  - pan and zoom on the `viewBox` (wheel, drag, buttons, fit)
  - a shape per kind and a border colour per state from the tokens
  - a side panel with `details[]` and the ref's link
  - keyboard: Tab focuses and selects a node, Enter follows its ref, Space
    selects, Escape clears
  - `aria-label` per node
- `pages/node/Topology.svelte`: loading, error and empty states from the kit.
  It reloads when the store's new `graphEpoch` moves. The store debounces it
  at 300 ms on the three session events, `endpoints_changed`, `fabric_added`
  and `fabric_removed`.
- Wiring:
  - `pages/index.ts` (`NodeTopology`), the shell tab, route
    `#/node/topology`, routes key `topology`
  - `refHref(ref, routes)` in the context
  - `kit/index.ts` exports
  - en/de catalogues (`graph.*`, `node.topology.*`)
- Client `topology()` and `controllerTopology(refresh)`. The stream helpers
  accept the new events. The controller store ignores `topology_changed`
  until U3b.
- Tests: client fixtures, a refresh token test, stream, store debounce,
  route, layout golden, `Graph`, `Topology` page and `catalogs.test.ts`.
  The bundle is rebuilt (`make web`).

### §4 Docs — `4f9f34f` and the report commit

- `docs/contract.md`: the routes, the events, a "topology graph" section and
  `problem-not_implemented`.
- `docs/embedding.md`: the session events and the planned controller route.
- `README.md`, `ui/README.md`, `CONCEPT.md` (wave U3a), `CLAUDE.md` state
  line.
- `changelog.md` `[0.4.0]`: Added, Planned (`/controller/topology`), and
  Changed (port method). This report.

## Guards

All 7 guards ran at HEAD `4f9f34f`, one at a time with `flock`; the summary
is in `~/.cache/go-fabric-scratch/u3a/summary.log`.

| Guard | Exit |
| --- | --- |
| `make ci` (build vet test lint generate-check web-check dist-check) | 0 |
| `make race` | 0 |
| `make e2e` | 0 |
| `make contract-check` | 0 |
| `make version-check` (0.4.0) | 0 |
| `make notices-check` | 0 |
| `make generate-check` | 0 |

`make cover-check` was green on the Go state of `b2618ec`. The report commit
touches docs only, and the full set is re-run at the final HEAD before the
push.

## Differences from the brief

1. **503 before assembly.** `/node/endpoints` answers 503 only without a
   node. As the brief asks, `/node/topology` also answers 503 while
   `Topology().Aggregator` is 0, with the same code `service_unready` and a
   different title. The fixture records the no-node case; a test covers the
   other one.
2. **The constant is named `EventTopologyChanged`**, not
   `ControllerEventTopologyChanged`, to match the existing controller
   `Event*` constants.
3. **The controller handler maps a source's graph to 200 already**, tested
   with a fake source. The 501 comes from `ErrNotImplemented`, which both
   sources in this repository return; it is not a fixed answer.
4. **Commit split.** `internal/api/api.gen.go` is in the §2 commit, not §1,
   so that every commit builds: the regenerated interface needs the new
   handler method. At `7ec0094` alone, `generate-check` is red.
5. **No catalogue coverage test existed in the tracked tree.** The
   `.gitignore` rule `coverage.*` swallows
   `ui/src/lib/i18n/coverage.test.ts`. An untracked copy sits in the main
   checkout `../go-fabric-ui`. This wave adds `catalogs.test.ts`: the same
   keys in en and de, every literal key in the catalogue, and the topology
   kinds, states and detail keys.
6. **The node gets a ref too** (`status`), and its state follows the health
   sample: `unknown` without a sample, `unreachable` when unhealthy.
7. **Reload path.** The page does not open a stream of its own. The store's
   single subscription moves a debounced `graphEpoch`, as `topologyEpoch`
   does for Endpoints and Surface.
8. **Links.** `ref` resolves through the context's `Routes` (`refHref`),
   which a host can map. `shell/route.ts` already parses `#/node/topology`;
   only its test changes.
9. **Bundle.** The vendor chunk's hash changed: the bundle now uses more of
   the Svelte runtime. No package was added, and `notices-check` is green.

## Decisions taken

- **Layout by signed depth.** Controllers are left of the node and the
  bridge is right of it. A node the walk does not reach takes the column of
  its kind, else the column left of every other.
- **Session nodes.** The label is the peer node id hex, or `PASE`. The
  state is `idle` only when the peer reported an idle interval and has been
  quiet longer than that interval.
- **Fabric label.** The fabric's label, else `eligibility.VendorName`
  (which returns the id itself for an unknown vendor).
- **Admin on the parameter.** `refresh` carries `x-admin` inside an inline
  parameter map, so the operation-level scan is unaffected.
  `controllerAdminQuery` is held against the contract by its own test.
- **501 detail.** The 501 problem has a fixed, generic detail and never
  carries an error's text.
- **No 403 or 200 controller fixtures** (concept §C.1 lists them; the brief
  does not). Both cases are tested in Go.

## Open questions

1. Should `.gitignore`'s `coverage.*` be narrowed (for example to
   `/coverage.*`) and the main checkout's untracked `coverage.test.ts`
   committed or dropped? `catalogs.test.ts` covers the same ground.
2. Adding a method to `ControllerSource` breaks a host that implements the
   port itself (the changelog says so). go-fabric-server embeds the adapter
   (`internal/daemon/controllerui.go`, `labelled`) and is unaffected.

## Left out

- The controller graph builder and `/controller/topology` 200 (U3b).
- The Network tab on the controller Nodes page (U3b).
- Playwright (U5) and live values (U4).
- Per-session subscription paths, partition/scope and tags (go-fabric issue,
  decision 0.4).
- No change under `contract/previous/`, and no new npm package.
