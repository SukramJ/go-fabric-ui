# Brief U4a — the node family's device details and the event log of both families (go-fabric-ui)

Date 2026-11-04. Concept `docs/concept-u4-device-details.md` (German; read
A–F in full first). This wave is D.3: the node side with values and names,
`attribute_changed`, the event routes of both families and the event tab.
**Model:** Opus. Repo `/home/markus/Dokumente/GitHub/go-fabric-ui`,
worktree `../go-fabric-ui-u4a`, branch `u4/node-details-events`, cut from
main after U4b (39373b8). Pins: go-fabric main **894adbcd** (the M9 module
wave, #391: the facade reads, element names, events, the controller's
event history; `go doc` after the bump), matterjs-server 0bd06f22 for the
comparison only. U4b's brief `notes/briefs/2026-11-03-U4b-controller-details.md`
is the pattern (sections, fixtures, guards, hand-back). One wave, one pull
request, one hand-back. No tag; version stays **0.5.0 unreleased**
(`changelog.md`, `ui/package.json`, `internal/version`); contract
**0.4.1** additive (`contract/openapi.yaml:4`), `contract/previous/openapi.yaml`
= the 0.4.0 copy for `make contract-check`.

## 0 Working rules

Commits as `SukramJ <sukramj@icloud.com>` (`git var GIT_AUTHOR_IDENT`
first), `<area>: <what> (U4a)`, no attribution lines. Guards under
`flock ~/.cache/go-fabric-scratch/guards.lock` with
`export PATH=$HOME/go/bin:$PATH GOTMPDIR=$HOME/.cache/gotmp`, logs and
`summary.log` (HEAD before/after, one `exit` line per guard) in
`~/.cache/go-fabric-scratch/u4a/`; read logs by grep. No change in
go-fabric or go-fabric-server: a gap there is a go-fabric issue text in the
hand-back, not a seam. Every go-fabric line you build on is read at
894adbcd and cited `file:line` in the hand-back.

## 1 Dependency

`go get github.com/SukramJ/go-fabric@894adbcd`
(`v0.16.1-0.20261104133901-894adbcd4cd3`; `changelog.md` `### Changed`
line). go-fabric's CHANGELOG `[Unreleased]` since 53b0d93d names the
breaks (`Registry.ForDevice(ctx, d, id)`, `SubscribeRequest.Updated`
returns error, `Images.DownloadUpdate(force)`, `store.Port` embeds
`Eraser`, `WiFiNetworkCommissioning.OnCommissioningComplete(ctx, …)`):
apply what the build and the tests demand, one report line each.

The calls this wave builds on (verified at 894adbcd with `go doc`):

- `node.Node.ReadCluster(ctx, endpoint, cluster) ([]node.AttributeValue, error)`
  and `ReadAttribute`; `node.AttributeValue{Endpoint, Cluster, Attribute,
  Value any, DataVersion}` with `JSON() (json.RawMessage, error)`; an
  unserved path is a `contract.StatusError`, before Start `node.ErrNotStarted`.
- `node.Node.Events(node.EventQuery{Endpoint *uint16, Cluster *uint32,
  Event *uint32, MinNumber uint64, Limit int}) []node.Event`
  (`Event{Number, Priority, Endpoint, Cluster, Event, Time, Payload any}`,
  oldest first, `Limit` keeps the newest).
- `node.Observers.OnAttributeChanged func(node.AttributeChange)` and
  `OnEvent func(node.Event)` (runs on the emitter's goroutine).
- `schema.AttributeName`, `schema.EventName`, `schema.CommandName`,
  `schema.ResponseCommandName` `(cluster, id uint32) (string, bool)`;
  `schema.ClusterFeature.Title`.
- `client.Node.EventHistory(since uint64, limit int) []client.RecordedEvent`
  (`RecordedEvent{EventData; Decoded any}`, oldest first;
  `client.Config.EventHistory`, `DefaultEventHistory` 100);
  `client.DecodeJSON(client.EventField(cluster, event), v, client.NameBased)`
  renders a decoded payload as the attribute path does today
  (`fabricui/gofabric/node.go:331-392` `clusterDetail`).

## 2 Contract 0.4.1 (additive; concept C.1)

- Type `EventRecord{number, priority, endpoint, cluster_hex, event_hex,
  name?, epoch_ms?, data?}` — one shape for both families; `data` is JSON
  by field name, absent without payload or when the decode fails (U4b
  follow-up 4, below). `priority` as the string enum the stream uses for
  diagnostics (`debug|info|critical`), coverage line in `coverage.test.ts`.
- `GET /node/endpoints/{n}` → `EndpointDetail` (node flavour: `decisions[]`
  from the surface report, no `conformance`; `clusters[]` **with values**
  from `Node.ReadCluster` and names from `schema`; `tags[]` from Descriptor
  TagList, `labels[]` from FixedLabel/UserLabel; `device_types[]` with
  revision; `parts[]` = direct children, see follow-up 2; endpoint 0 and
  the aggregator answer too). 404 for an endpoint the node does not serve
  (`contract.StatusError` UNSUPPORTED_ENDPOINT), 503 before Start.
- `GET /node/endpoints/{n}/events` → `EventRecord[]` (`Node.Events` with
  the endpoint filter; query `since` (number, optional) and `limit`
  (optional, default 100, max 1000); newest last as the facade returns
  them). `/node/events` stays the diagnostic trace — do not touch it.
- `GET /controller/nodes/{id}/events` → `EventRecord[]` from
  `EventHistory(since, limit)`; 404 for an unknown node, as the other
  controller routes.
- Stream event `attribute_changed {endpoint, cluster_hex, attributes[]}`
  (ids, no values) from `Observers.OnAttributeChanged`, published through
  `fabricui.NodeObservers` (`fabricui/events.go:131-170`) like
  `surface_changed`; `docs/contract.md` stream table and the embedding
  guide sentence.
- `make generate`, `make gen-types`, fixtures recorded with
  `go test ./fabricui -run <Test> -update` and reviewed.

## 3 Go half (concept C.2)

- `fabricui.NodeSource` (`fabricui/options.go:25`) gains the read and event
  methods `*node.Node` satisfies; `nodePort`/`nodeAdapter`
  (`fabricui/ports.go:18-52`) follow; handlers in `fabricui/handlers_node.go`
  beside the endpoints handler; `fabricui.EndpointDetail`/`ClusterDetail`
  are reused (U4b, `fabricui/controller.go`), `fabricui.EventRecord` new.
- Node-side detail built from `Node.Topology()` (structure, device types,
  parts, parent) joined with `Node.ReadCluster` per listed cluster
  (values, DataVersion) and `Node.Surface()` (decisions). Names:
  `schema.AttributeName/EventName/CommandName/ResponseCommandName`,
  cluster name and feature titles as U4b's `clusterDetail` does; a cluster
  whose read fails is listed without values and the error noted in the
  hand-back.
- Controller adapter `fabricui/gofabric/node.go`: `Events(since, limit)`
  through `EventHistory`, names via `schema.EventName`, `data` through
  `DecodeJSON(EventField(...), …)`.
- **U4b follow-ups (the four, from PR #36's review):** (1)
  `ClusterDetail` answers 404, not 503, for a listed cluster before the
  Descriptor is read — return `ErrNotRead` while `ep.Cluster(0x001D)` is
  nil; (2) `parts` = direct children (`ep.Parts()`); the schema description
  says so and `MemoryController` follows; (3) `Inventory.svelte:33-48`
  gets a sequence guard so an older reload cannot land last; (4) a
  `DecodeJSON` failure leaves `value` absent, not `null`.
- Test doubles: the node-side double the handler tests use and
  `MemoryController` grow with the ports (one fixture-shaped endpoint with
  values, two events each), so the Go handler tests and the TS page tests
  share one shape.

## 4 Svelte half (concept C.3)

- Node page `#/node/endpoint/<n>` (routes key `endpoint`, `context.ts`;
  `refHref` maps topology refs `{page:"endpoints", id}` to it): header with
  device types `name rN`, tags, labels, reachable; the clusters in the
  shared `ClusterView` + `AttributeTable` (node flavour: values inline);
  the host's `decisions` as in the surface page. Reload on
  `attribute_changed` for its endpoint (refetch the endpoint; ids only,
  no values on the stream) and on `endpoints_changed`/`surface_changed`.
- `node/Endpoints.svelte`: revision per device type, the link to the
  detail page, rows for endpoint 0 and the aggregator.
- Events: kit `EventList.svelte` (number, time, priority badge, endpoint,
  cluster·event name, data as formatted JSON, newest first in the view);
  node page section "Events" on the endpoint page; controller
  `pages/controller/Node.svelte` tab **events** (`controller.node.tab.events`,
  coverage line), reload on `node_updated`, "since" kept so a reload fetches
  only the newer ones.
- Client `ui/src/lib/api/client.ts`: `nodeEndpointDetail`, `nodeEndpointEvents`,
  `controllerNodeEvents`; test table rows in `client.test.ts` /
  `controller-client.test.ts`; `stream.ts` knows `attribute_changed`.
- Fixtures (every one used on both sides: `api/fixtures.test.ts`,
  `fabricui/misc_test.go:308`): `node-endpoint.200`, `node-endpoint.404`,
  `node-endpoint-events.200`, `controller-node-events.200`; page tests for
  the endpoint page, the events tab and the Endpoints list; i18n keys in
  `en.json` and `de.json`, `catalogs.test.ts` green.
- e2e (`fabricui/gofabric/e2e_test.go` and the node-side e2e if one
  exists): the real adapters against a `fabrictest` pair — the node's
  endpoint detail carries OnOff's value, and a toggled OnOff reaches the
  controller's event history (OnOff has no event: use a cluster of the
  bridged device that emits one, or the root's StartUp from a restart) and
  the node's `attribute_changed`.

## 5 Docs and changelog

`docs/contract.md` (routes, stream event, version 0.4.1), `docs/embedding.md`
only if the host wiring changes (it should not: `NodeObservers` carries
the new observer), `changelog.md` `[0.5.0]` `### Added` (routes, types,
page, tab, components), `### Changed` (go-fabric bump; the U4b
follow-ups), `README.md` feature line.

## 6 Guards

`make ci` (build, vet, test, lint, web-check, dist-check, contract-check,
version-check, generate-check, notices-check), `make race`, `make e2e`,
`make cover-check` if the Makefile has it — under the flock, logs
`~/.cache/go-fabric-scratch/u4a/`, summary `u4a/summary.log`. `make web`
in the same commit as any `ui/src` change (dist-check).

## 7 Out of scope

Stream values, writes from the inventory or the endpoint page, Playwright
(U5), a tag-name table, any go-fabric change, the tag.

## Hand-back

Per §: what landed (files, routes, components, tests by name); guards
table (guard, HEAD, exit); differences from this brief; decisions; left
out; open questions; go-fabric issue texts for gaps found; every go-fabric
line cited, re-read at 894adbcd. Open the pull request against main
(English, hand-back as body, no labels, auto-merge off) and hand back
with its number.
