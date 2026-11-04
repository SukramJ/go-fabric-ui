# Brief U4b — the controller family's device details (go-fabric-ui)

Date 2026-11-03. Concept `docs/concept-u4-device-details.md` (German; read
it in full first: A–F). This wave is D.1: the controller side, adapter
and pages, from the cached image only. **Model:** Opus. Repo
`/home/markus/Dokumente/GitHub/go-fabric-ui`, worktree `../go-fabric-ui-u4b`,
branch `u4/controller-details`, cut from main after the concept merged.
Pins: go-fabric main 53b0d93d (`go doc` after the bump; the facade import
rule of `internal/arch/arch_test.go:29-70` is unchanged), matterjs-server
0bd06f22 for the comparison only. Map
`~/.cache/go-fabric-scratch/read-u4.md` §1 and §3 (a map, not a citation:
verify every line you build on). One wave, one pull request, one hand-back.
No tag (the owner tags); version 0.5.0 in `changelog.md`, `ui/package.json`
and `internal/version` (`make version-check`), contract 0.4.0
(`contract/openapi.yaml:4`), `contract/previous/openapi.yaml` = the 0.3.1
copy for `make contract-check`.

## 1 Dependency

`go get github.com/SukramJ/go-fabric@53b0d93d` (pseudo-version; `[Unreleased]`
`### Changed` line). Apply the migrations go-fabric's CHANGELOG
`[Unreleased]` names since c91e5b7f (`Service.AddSource`, phase 30's
`device.Options.DeviceTypes/Tags`, `#387`): read them, apply what the build
demands, one report line each.

## 2 Contract 0.4.0 (additive; concept C.1)

- Types `ClusterDetail`, `EndpointDetail` (controller flavour: `conformance`,
  no `decisions`), `EventRecord` is **not** in this wave (U4a). Field names
  and optionality exactly as C.1; `value` is JSON by field name as
  `AttributeValue.value` (`contract.md:262-264`); `name` optional.
- Routes `GET /controller/nodes/{id}/endpoints/{ep}` → `EndpointDetail`
  (structure, no attribute values) and
  `GET /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}` →
  `ClusterDetail` with values **from the cache** (owner decision F.0.1; no
  wire read, no admin; a node the controller has no image of answers 404,
  a cluster the endpoint lacks 404, the server's own failure 503 — follow
  the existing controller routes' error shapes).
- `make generate` (the Go handlers in `internal/api`, never hand-edited),
  `make gen-types` (`schema.gen.ts`), `docs/contract.md` paragraph for the
  two routes and the amended 0.4 sentence ("reads are targeted … a
  read-only inventory from the controller's image is served by the
  endpoint and cluster routes").

## 3 Go half (concept C.2)

- Port `fabricui.ControllerNode` gets `EndpointDetail(ep uint16)
  (EndpointDetail, error)` and `ClusterDetail(ep uint16, cluster uint32)
  (ClusterDetail, error)`; `fabricui/controller.go` types mirror the
  contract; handlers in `fabricui/handlers_controller.go` beside
  `nodeDetail` (`:346-398`).
- Adapter `fabricui/gofabric/node.go`: from `client.Endpoint`
  (`controller/client/node.go:56-148`) and `client.Cluster`
  (`controller/client/cluster.go:207-398`): `Revision()`, `FeatureMap()`
  (hex), features as `{code, title, bit}` (`Features()` codes +
  `Definition().Features` for title and bit, `cluster/spec/model.go:332`;
  codes only where the definition is nil), `AttributeList()` with
  `Value()` + `client.DecodeJSON(client.AttributeField(cluster, attr),
  node, client.NameBased)` as `Read` does today (`:140-172`), `Version()`,
  `AcceptedCommands()/GeneratedCommands()` with names from
  `Definition()`, event ids with names; DeviceTypeList and TagList from
  the Descriptor attributes 0 and 4 (`Cluster(0x1D).Decoded`), Reachable
  from BridgedDeviceBasicInformation 0x0039/0x0011 where the endpoint has
  it, `Endpoint.Conformance()` (`controller/client/conformance.go:93`)
  as `conformance{status, violations[], unjudged[]}`, parts as numbers,
  `Owner()` as parent, `Name()`.
- `MemoryController` (`fabricui/memory_controller.go`) grows with the
  port: a fixture-shaped node with two endpoints, one cluster with three
  attributes, a tag and a feature map, so the Go handler tests and the
  TS page tests share one shape.
- No change to `Hooks` (events are U4a); `node_updated` already reloads
  the node's pages.

## 4 Svelte half (concept C.3)

- `pages/controller/Node.svelte`: tab **inventory** (`controller.node.tab.inventory`
  in `en.json`/`de.json` and the coverage line `coverage.test.ts:46-61`):
  endpoint list with device types `name rN` and tags as chips, then per
  endpoint the clusters; a cluster expands lazily
  (`client.controllerClusterDetail(nodeId, ep, clusterHex)`) into the
  shared `ClusterView`.
- Kit components `ui/src/lib/kit/ClusterView.svelte` (header: name, id hex,
  revision, feature map hex, feature chips `code · title`, data version)
  and `AttributeTable.svelte` (id hex, name, value as formatted JSON,
  status), `TagChips.svelte`; written for both families (U4a reuses them
  with the node flavour).
- `node/Endpoints.svelte` gets the DeviceTypeList (all types, each `rN`),
  Reachable badge per bridged endpoint, and a link into the inventory tab;
  the **attributes** tab stays as it is (F.1.6).
- Client `ui/src/lib/api/client.ts`: `controllerEndpointDetail`,
  `controllerClusterDetail`; `api/controller-client.test.ts` table rows.
- Fixtures (both sides must use every one: `api/fixtures.test.ts`,
  `fabricui/misc_test.go:308`): `controller-node-endpoint.200`,
  `controller-node-endpoint.404`, `controller-node-cluster.200`,
  `controller-node-cluster.404`; page tests in
  `pages/controller/Node.test.ts` (inventory tab renders the fixture,
  lazy load counts one request per expanded cluster, 404 shows the error
  state); i18n keys in both catalogues, `catalogs.test.ts` green.
- e2e (`fabricui/gofabric/e2e_test.go`): one case over `MemoryController`'s
  replacement — the real adapter against a `fabrictest` node with a
  bridged OnOff endpoint: the cluster detail of OnOff carries `OnOff`'s
  value and the feature chips.

## 5 Docs and changelog

`docs/contract.md` (routes, the 0.4 amendment, version 0.4.0),
`docs/embedding.md` if the host wiring changes (it should not),
`changelog.md` `[0.5.0]` (unreleased date "unreleased"; the owner tags):
`### Added` routes/types/tab/components, `### Changed` the go-fabric bump.
`README.md` feature list line.

## 6 Guards

`make ci` (Go + vitest + lint + contract-check + version-check +
generate-check), `make race`, `make cover-check`, under
`flock ~/.cache/go-fabric-scratch/guards.lock`, `export PATH=$HOME/go/bin:$PATH`,
logs `~/.cache/go-fabric-scratch/u4b/`, summary `u4b/summary.log` (HEAD
before/after, one `exit` line per guard). Commits as
`SukramJ <sukramj@icloud.com>` (`git var GIT_AUTHOR_IDENT` first), no
attribution lines, one PR against main (English, hand-back shape as body,
no labels, auto-merge off).

## 7 Out of scope

Events and the event tab (U4a), the node family's detail page (U4a),
stream values, writes from the inventory, Playwright (U5), any change in
go-fabric or go-fabric-server.

## Hand-back

Per §: what landed (files, routes, components, tests by name); guards
table (guard, HEAD, exit); differences from the brief; decisions; left
out; open questions; every go-fabric line cited, re-read at 53b0d93d.
