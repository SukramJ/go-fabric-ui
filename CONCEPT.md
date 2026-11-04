# Concept

The design is [concept v9](https://github.com/SukramJ/go-fabric/blob/main/docs/concept-v9-fabric-ui.md)
in go-fabric (German working paper). In short:

- **Role.** One web interface for the daemons of the go-fabric family,
  maintained here and only embedded by the projects. Two page families:
  *node* (health, pairing, fabrics, exposure, endpoints, diagnostics) and
  *controller* (commission, discovery, nodes, subscriptions, bindings/ACL,
  multi-admin window, OTA, Thread).
- **Shape.** A Go module (`fabricui`: handlers on the `node.Node` facade,
  no seam below it; the contract; the built bundle under `fabricui/dist/`
  embedded with `go:embed`) and an npm package `@sukramj/fabric-ui` (Svelte 5,
  Vite, Tailwind 4; the pages, the kit, the tokens, the theme bridge, the
  generated client). One version for both, one tag releases both.
- **Host-specific only through `ExposureProvider`** (`Schema`, `List`,
  `Preview`, `Apply`): the selection page renders the host's columns and
  verdicts (`eligibility.Verdict`) and shows where each verdict comes from.
- **Contract first.** `contract/openapi.yaml` with sections `node` and
  `controller`; routes `/api/v1/node/…` and `/api/v1/controller/…`; golden
  fixtures under `contract/fixtures/` read by the Go and the TypeScript
  tests; no removal in a minor version.
- **The bundle is committed**, because a Go module is fetched from the git
  tag and a consumer without Node.js must receive it; a CI job rebuilds it
  and fails when the committed bundle is not byte-identical.
- **Origin.** The node pages and their handlers are extracted from
  openccu-loom (`assets/ui/src/routes/matter/`, `/api/v1/matter/*` at
  `b40b289a`), made host-neutral; loom migrates onto this package in a wave
  of its own.

## Waves

| Wave | Content | Result |
| --- | --- | --- |
| U1 | node family: contract v1, Go module with handlers and provider, Svelte library and standalone shell, dist guard, release with npm publish | `v0.1.0` |
| U2 | controller family: contract 0.2.0, the `ControllerSource` port and the `gofabric.ControllerOf` adapter, handlers generated from the contract (oapi-codegen), four pages, family detection in the shell (brief U2, `docs/concept-u2-controller-family.md`) | `v0.2.0` |
| U3a | the topology view of the node family: contract 0.3.0 (`/node/topology`, session stream events, `/controller/topology` planned), the `Graph` kit component, the Topology page (brief U3a, `docs/concept-u3-topology.md`) | `v0.4.0` |
| U3b | the topology view of the controller family: contract 0.3.1 (`/controller/topology` served, link strength `none`), `ControllerOptions.Topology` over go-fabric's `controller/topology`, the Network tab of the nodes page (brief U3b) | `v0.4.1` |
| U4b | the controller family's device details: contract 0.4.0 (`GET /controller/nodes/{id}/endpoints/{ep}` and `…/clusters/{cluster_hex}` from the controller's image, `EndpointDetail`, `ClusterDetail`), the Inventory tab and the kit's `ClusterView`, `AttributeTable`, `TagChips` (brief U4b, `docs/concept-u4-device-details.md`) | `v0.5.0` (unreleased) |
| U4a | the node family's device details and the event log of both families: contract 0.4.1 (`GET /node/endpoints/{ep}` with values, `GET /node/endpoints/{ep}/events`, `GET /controller/nodes/{id}/events`, `EventRecord`, the stream's `attribute_changed`), the endpoint page, the events tab and the kit's `EventList` (brief U4a) | `v0.5.0` (unreleased) |
| U5 | Betrieb: Live-Status, Betriebsseite (Log-Level, Log-Stream, Info), Ping, QR-Scan, Ökosystem-Hinweise, Node-Wähler für mehrere Geräte-Nodes, Playwright-Smoke je Seite | `docs/concept-u5-operation.md` |
