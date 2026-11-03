# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
[Semantic Versioning](https://semver.org/). The Go module and the npm package
`@sukramj/fabric-ui` share one version.

## [Unreleased]

## [0.4.1] - 2026-11-03

The topology view of the controller family (brief U3b, concept
`docs/concept-u3-topology.md`); contract 0.3.1, additive over 0.3.0.

### Added

- `GET /controller/topology` answers 200: the controller's network graph
  from go-fabric's `controller/topology` — the controller with a session
  to each commissioned node (`peer:<node_id_hex>`), the Thread neighbours
  no node matches (`thread_router`, `thread_end_device`), the Wi-Fi access
  points (`wifi_ap`) and a `link` per Thread pair and Wi-Fi station with
  its strength, LQI and RSSI. `gofabric.ControllerOptions.Topology` takes
  the host's `*topology.Service`; its changes reach the controller stream
  as `topology_changed`. Without a service the route still answers 501.
- `TopologyEdge.strength` gains `none` (a link observed dead, LQI 0).
- The **Network** tab of the controller's nodes page (`ControllerNodes`
  prop `tab`, route `#/controller/nodes?tab=network`, routes key
  `controllerNetwork`; `ControllerNetwork` alone): the graph, an admin
  "Read diagnostics again" (`refresh=true`), a reload on the controller
  store's new `graphEpoch`. The kit's `Graph` draws a link's strength
  (width, dotted when dead) with its quality as the title.
- `MemoryController.Topology`: the graph the in-memory controller serves.
- Fixtures `controller-topology.200` and `controller-topology.403`.

### Changed

- Contract 0.3.1: `/controller/topology` is served (no `x-status:
  planned`).
- go-fabric `v0.16.1-0.20261103130909-c91e5b7f4a2f` (main at `c91e5b7f`,
  `controller/topology`), a pseudo-version until go-fabric's next tag.

## [0.4.0] - 2026-11-03

The topology view of the node family (brief U3a, concept
`docs/concept-u3-topology.md`); contract 0.3.0, additive over 0.2.0.

### Added

- `GET /node/topology`: the node as a graph (`TopologyGraph`) — the node,
  its Aggregator, the bridged devices and their parts, the fabrics and the
  sessions their controllers hold — built from `Topology`, `FabricInfos`
  and `SessionDiagnostics`; 503 without a node or before it assembled.
- The node stream carries `session_opened` (CASE sessions only, as go-fabric
  reports them), `session_closed` and `subscriptions_changed` with a
  `SessionEvent` (the session as `/node/sessions` lists it);
  `fabricui.NodeObservers` yields them too.
- The **Topology** page (`NodeTopology`, route `#/node/topology`, routes key
  `topology`) and the kit's `Graph` component with its pure layout
  (`layoutGraph`): hand-written SVG, pan and zoom, a details panel that
  links each node to the page that lists it, keyboard access. It reloads
  on the store's new `graphEpoch`.
- Client `topology()` and `controllerTopology(refresh)`; problem code
  `not_implemented` (501); the catalogue test (`catalogs.test.ts`).
- `GET /controller/topology` declared with its final schema (then planned,
  501), `refresh=true` admin-only, and the controller stream's
  `topology_changed`; served from 0.4.1.

### Changed

- **Breaking for a host with its own `ControllerSource`:** the port gains
  `NetworkTopology(ctx, refresh) (NetworkTopology, error)`; returning
  `fabricui.ErrNotImplemented` keeps the route at 501. A source wrapping
  `gofabric.ControllerOf` inherits it.

## [0.3.0] - 2026-11-02

Built on go-fabric v0.16.0, whose facade removed what `NodeSource` and the
device type names rested on. The routes, their bodies and the npm package's
pages are unchanged.

### Changed

- go-fabric `v0.16.0` (was `v0.15.6`), which raises `modernc.org/sqlite` to
  1.61.0, `modernc.org/libc` to 1.79.0, `go-humanize` to 1.1.0 and
  `go-strftime` to 1.1.0 in the build list.
- **Breaking for a host with its own `NodeSource`:**
  `CommissioningWindow() *node.CommissioningWindow` is replaced by
  `OpenOwnWindow(ctx, timeout) error` and `CloseWindow(ctx) error`.
  `*node.Node` satisfies the interface as before.
- `fabricui.NodeObservers`: the hooks return nothing, as go-fabric's
  `node.Observers` now declares them.
- `POST /node/pairing/window/close` closes the window through
  `Node.CloseWindow` (matter.js `endCommissioning`) instead of
  `RevokeWindow`: a PASE session already established and an armed fail-safe
  now stay with their commissioner, where they were evicted and expired
  before. The route's status codes are unchanged. A host that needs the
  eviction wraps its node in a `NodeSource` whose `CloseWindow` revokes.
- `device_type_name` keeps its values: go-fabric removed
  `contract.DeviceTypeName` and its label table, so the 25 operator labels
  ("On/Off Light") move to `fabricui`, and every other device type reads
  `schema.DeviceTypeName`. The import guard allows `schema` beside the facade.

## [0.2.9] - 2026-11-01

Every response schema of the contract is open to unknown fields again, and
a guard keeps it so; the npm package is published through trusted
publishing.

### Fixed

- `contract/openapi.yaml`: the nine schemas of the surface report
  (`SurfaceReport`, `SurfaceEffective`, `SurfaceEndpoint`,
  `SurfaceDeviceType`, `SurfaceCluster`, `SurfaceElement`,
  `SurfaceDecision`, `SurfaceChange`, `SurfaceElementChange`) no longer set
  `additionalProperties: false`. A client generated from a closed schema
  decodes it strictly, so the next field go-fabric adds to
  `surface.Report`'s JSON — additive for the contract guard, a minor for the
  changelog — made every installed client reject the whole
  `GET /node/surface` answer. No other response schema was closed; request
  bodies keep the keyword, where the handler is the decoder. Closes #27.
- The contract guard (`make contract-check`) fails on a schema reachable
  from a response that sets `additionalProperties: false`, and its tests
  check every recorded fixture's response against the route's schema, so a
  field the encoder writes and the contract does not name is caught at test
  time instead of on a client's wire.

### Changed

- go-fabric `v0.15.6` (was `v0.15.4`); oapi-codegen runtime `v1.8.0` and
  generator `v2.8.2` (was `v1.7.0`, `v2.8.0`) — the regenerated handlers
  bind request bodies through `runtime.BindJSONBody`, the same decoding.
  npm: `@lucide/svelte` 1.66.0, `svelte` 5.57.3, `vite` 8.3.5,
  `@types/node` 26.6.6; the bundle is rebuilt. `minimatch`, reached only
  through `openapi-typescript`'s `@redocly/openapi-core` (the type
  generator, not the bundle), is overridden to 11.0.1 for GHSA-g3hv-75wg-m249
  (`overrides` in `ui/package.json`; `gen:types` produces the same file).
  TypeScript stays at 6.x: 7.0 is a major whose toolchain `svelte-check`
  4.7 does not declare. Left as reported: `deepmerge` ≤ 4.3.1 (low,
  GHSA-gcqx-w48m-w2h4) under `@sveltejs/vite-plugin-svelte`, no fixed
  version published.
- The release workflow publishes `@sukramj/fabric-ui` through npm's trusted
  publishing: npmjs accepts the job's GitHub OIDC token (Trusted Publisher on
  the package, set to `SukramJ/go-fabric-ui` and `release-on-tag.yml`) and
  signs the provenance from it. The `NPM_TOKEN` secret, which npm limited to
  a short lifetime, is no longer read; a missing trust configuration fails
  the publish instead of skipping it. The job installs npm 11 first, which
  trusted publishing requires.

## [0.2.8] - 2026-10-30

`surface_changed` comes from the node itself, so every assembly that
changes an endpoint's surface reaches the stream; the Surface page shows the
root endpoint.

### Changed

- `surface_changed` comes from the node's observer
  (`Observers.OnSurfaceChanged`, registered per stream through
  `NodeObservers`): one event per changed endpoint, so a reassembly the
  host runs on its own reaches the stream too. The handler no longer
  compares `Surface().Changed` before and after its own reassemblies.
- The status page's warning count reads "{count} warnings" — any policy's
  `Warn` counts, not only a profile's.
- The Surface page shows the root endpoint (`effective.root`) as the first
  endpoint card, titled "Root endpoint"; its decisions are listed under
  that title.
- go-fabric `v0.15.4` (was `v0.15.3`), which adds `Observers.OnSurfaceChanged`.

## [0.2.7] - 2026-10-30

The node's surface report gets its own page, and a reassembly that changes
an endpoint's surface reaches the stream and the Endpoints page.

### Added

- `GET /node/surface`: go-fabric's `surface.Report` of the node in
  go-fabric's own JSON encoding — the effective surface per endpoint, the
  decision ledger, the withheld elements and the announced surface changes
  (contract schema `SurfaceReport`).
- The Surface page (`NodeSurface`; `#/node/surface` and a tab in the
  standalone shell, `Routes.surface`): the policies' level, profiles,
  extensions and digest; per endpoint its device types and clusters with
  revision, feature map and element counts; the decisions, filterable by
  verdict and endpoint, the silent keeps hidden by default; the withheld
  elements; the announced changes with their element changes.
- The stream event `surface_changed` (`SurfaceChangedEvent`, the endpoint
  keys): published after `endpoints_changed` when a reassembly this
  handler runs announced surface changes.
- The Endpoints page marks an endpoint whose surface changed in this
  process ("surface changed"), and refetches on `endpoints_changed` and
  `surface_changed`.

### Changed

- The status page's `warning_count` counts `Warn` decisions — the verdict,
  not the reason's text (go-fabric v0.15.3's `surface.Warn`).
- The German status text says "Pin" for a level taken from the pin.
- go-fabric `v0.15.3` (PR #345) (was `v0.15.2`), which adds `surface.Warn`.

## [0.2.6] - 2026-10-30

The node status page shows the policies go-fabric v0.15 composes the served
surface under: the specification level, the controller profiles and the
enabled extensions.

### Added

- A "Surface" row on the node status page, below the counts: the
  specification level with its origin (set, the pin, or the profiles that
  imply it), the selected controller profiles, the enabled
  ahead-of-matter.js extensions, and — only when not zero — the withheld
  elements, the profile warnings and the surface changes announced in this
  process.
- `GET /node/status` reports a `surface` object (absent without a node):
  `level`, `level_origin`, `profiles`, `extensions`, `assembled`,
  `withheld_count`, `warning_count`, `changed_count`, read from
  `node.Node.Surface()`.

### Changed

- go-fabric `v0.15.2` (was `v0.15.1`), which adds `surface.Report.Extensions`.
- `NodeSource` gains `Surface() surface.Report`. `*node.Node` satisfies it
  at go-fabric v0.15.2; a host that implements `NodeSource` with its own
  type adds the method.
- The import guard allows `github.com/SukramJ/go-fabric/surface` as a facade
  type package: it is the type of `node.Config.Surface` and of
  `Node.Surface()`'s result.

## [0.2.5] - 2026-10-30

`fabricui/gofabric` builds against go-fabric v0.15.1, and the selection page
hides what the host cannot expose.

### Changed

- The selection page hides items whose verdict is unmappable — nothing the
  host could expose now — by default. A toggle above the list ("Show items
  that cannot be mapped (n)") brings them back, with their reason and the
  disabled checkbox as before; the value filters still offer every item's
  values. The legend keeps all four states.
- go-fabric `v0.15.1` (was `v0.14.0`): the served surface as a value under
  policies (ADR 0030) — the node's spec level, ahead-of-matter.js opt-ins and
  controller profiles sit in `node.Config.Surface`; nothing in this module's
  routes changes.

## [0.2.4] - 2026-10-29

`fabricui/gofabric` builds against go-fabric v0.14.0, and a host gets the
node's transitions over its own channel through an exported route
(issue #19).

### Added

- `NodeObservers(push, open) node.Observers`: the node's fabric and window
  transitions as the node page family's events, with the stream's
  payloads — the conversion the handler registers for each open stream,
  which since v0.2.0 is the transitions' only route. A host that carries
  the events over a channel of its own (a WebSocket hub, an audit log)
  registers it with `node.Observe` instead of rebuilding it, and does not
  publish the transitions into `Events`, which would hand every stream
  each one twice.

### Changed

- go-fabric `v0.13.3` → `v0.14.0`. The controller adapter decodes a Thread
  dataset with `github.com/SukramJ/go-fabric/thread`, where go-fabric
  v0.13.7 moved the codec from `controller/thread` (every identifier
  unchanged); `fabricui/gofabric` did not compile against go-fabric
  v0.13.7 or later before. The import guard allows `thread` in its place.

## [0.2.3] - 2026-10-27

### Changed

- Build tooling: `@lucide/svelte` 1.57.0 → 1.59.0, `svelte-check` and
  `vitest` minor updates (Dependabot #3). The rebuilt bundle is
  byte-identical; `THIRD-PARTY-NOTICES.md` records the new `@lucide/svelte`
  version.

## [0.2.2] - 2026-10-27

The pages take Home Assistant's theme from the first frame: inside an
Ingress panel they wear the live theme (colours and light/dark, a custom
theme included), and standalone they wear Home Assistant's default light
and dark palettes, following `prefers-color-scheme`. Checked in headless
Firefox inside a same-origin fake Home Assistant parent with a light, a
dark and a custom theme, and standalone.

### Fixed

- No flash of the wrong palette on a cold load. The shell's pre-paint
  script sets the Home Assistant skin before the first paint and, inside
  Home Assistant's Ingress frame, copies the parent's theme variables and
  takes its light/dark from the parent's background — before, the first
  frames showed the standalone teal/white palette, and dark followed the
  operating system instead of Home Assistant until the bundle had run.
- The theme bridge drops a variable the parent no longer defines (a custom
  theme switched back to the default), so the token falls back to Home
  Assistant's default instead of keeping the old theme's value.
- The underline tabs no longer show a vertical scrollbar: the active
  marker overlapped the list's border by one pixel, which overflowed the
  scrollable tab list.

### Added

- `--ha-on-primary-color`, the label on a primary fill: white by default,
  Home Assistant's `--text-primary-color` under the HA skin, so a light
  primary in a custom theme keeps a readable button label. The primary
  button uses it.
- Under the HA skin, native controls (checkboxes, radios, scrollbars) take
  `accent-color: var(--ha-primary-color)` and the skin's `color-scheme`.
  A host on another skin (`data-skin` other than `ha`) is unchanged.

### Deprecated

- The comma form of `Item.Fields` for a `Multi` column, announced in
  v0.2.1 for removal in this release, stays accepted; its removal moves to
  the next release that carries breaking changes for go-fabric-server and
  go-ha2fabric.

## [0.2.1] - 2026-10-27

### Fixed

- The module compiles on 32-bit targets again: two range checks in the
  controller handlers compared an `int` with `0xFFFFFFFF`, which overflows
  where `int` is 32 bits (found by go-fabric-server's armv7 add-on image,
  report S4). CI now builds with `GOARCH=386`. The comma form of
  `Item.Fields` for a `Multi` column stays accepted in this release; its
  removal moves to v0.2.2, after go-fabric-server v0.4.0 has bumped.

## [0.2.0] - 2026-10-27

The controller page family (wave U2, brief U2): every go-fabric
embedding that runs a controller gets the same pages for its nodes, as
every node embedding gets the node pages. The HTTP contract goes to 0.2.0
with the controller section; the node section is unchanged.

### Added

- **Contract 0.2.0**: the controller routes under `/api/v1/controller/`
  (status, nodes, a node's detail, decommission, interview,
  enable/disable, auto-subscribe, the commissioning window, targeted
  attribute reads, a generic write and invoke, ACL, bindings, ICD, the
  node's fabrics, commissioning, discovery, credentials, the fabric label,
  the update queue, the stream) with request and response schemas, hex
  ids, `x-admin` flags and the problem code `node_error` (502). Fixtures
  for every controller response.
- `fabricui.ControllerSource` and `ControllerNode`, the port with its own
  value types (no go-fabric type in it), and `Options.Controller`; nil
  serves no controller family (`/controller/*` answers 404).
- `fabricui/gofabric.ControllerOf(*controller.Controller,
  ControllerOptions)`, the adapter over go-fabric's controller facade;
  `gofabric.Hooks` deliver the nodes' transitions to the stream
  (`controller.Config.HooksFor`, `Wrap` keeps the host's own hooks;
  `Close` before the controller's stops their timers);
  `gofabric.ThreadDatasetName` names a Thread dataset by go-fabric's
  decoder.
- The controller handlers are generated from the contract with
  oapi-codegen (pinned as a `go.mod` tool; `make generate`,
  `make generate-check` in CI); the handler bodies are hand-written.
- `fabricui.NewMemoryController` and `fabricui.NewMemoryCredentials`,
  in-memory implementations for tests and demonstrations.
- npm: the pages `ControllerNodes`, `ControllerCommission`,
  `ControllerNode` (tabs overview, endpoints, attributes, ACL, bindings,
  window, ICD, fabrics, updates) and `ControllerSettings`;
  `createControllerStore` (`ui.controller`); the client's controller
  methods; `subscribeControllerStream`; the routes `controllerNodes`,
  `controllerCommission`, `controllerSettings`, `controllerNode`; i18n
  keys `controller.*` and `shell.*` in English and German.
- The standalone shell shows the families the host serves, by presence
  (`/node/status`, `/controller/status`; a 404 hides a family, as does a
  disabled node next to a controller), with hash routes
  `#/controller/<page>` and `#/controller/node/<id>`.

### Changed

- go-fabric `v0.12.2` → `v0.13.3`.
- The contract guard lets an operation marked `x-status: planned` drop
  its placeholder responses when it is served; its path and method stay
  guarded.

### Deprecated

- The comma-separated `Item.Fields[key]` for a `Multi` column or filter
  is removed in v0.2.1, after go-fabric-server and go-ha2fabric have
  bumped; fill `Item.Lists` instead.

### Removed

- `Events.Observers` (deprecated since v0.1.2): the stream observes the
  node itself. A host that still calls it drops the call.

## [0.1.3] - 2026-10-26

List-valued item fields travel as JSON arrays, so a value with a comma —
a label named "Living, dining" — stays one chip and one filter option
(issue #10, brief F2). The HTTP contract only gains an optional field;
hosts that fill `Fields` alone keep working unchanged.

### Added

- `Item.Lists map[string][]string`, answered as `lists` on
  `GET /node/exposure` and `POST /node/exposure/preview` (absent when the
  provider leaves it empty): the values of a list-valued field, one entry
  per value, no separator semantics. A `Multi` column or filter reads
  `Lists[key]`; when the key is in both `Lists` and `Fields`, `Lists`
  wins.
- `Item.ListValues(key)` returns the list as the page reads it.

### Deprecated

- The comma-separated `Item.Fields[key]` for a `Multi` column or filter.
  It stays accepted as the fallback when `Lists` has no such key until
  v0.2.0; `Fields[key]` may keep carrying a display string.

## [0.1.2] - 2026-10-25

The handlers read go-fabric v0.12.2's facade routes instead of the U1
workarounds (brief U1b, go-fabric report CK Q2). The HTTP contract only
gains optional fields; the config-time observer wiring is deprecated.

### Added

- `GET /node/status` answers `window_kind` (`none`, `own`, `basic`,
  `enhanced`) and `window_deadline` (RFC 3339, the node's clock; absent
  while no window is open). The Status page and the pairing flow count the
  window down from the deadline, so a reload shows what is left.
- `GET /node/endpoints` answers `device_type_revision`, and every
  endpoint's clusters carry `name` and `revision` — parts and a host
  snapshotter's endpoints included, which listed no clusters before. The
  list is the mounted servers in ServerList order, Descriptor and
  BridgedDeviceBasicInformation among them.
- `GET /node/pairing` answers the configured `discriminator` and
  `passcode` (`Node.SetupPayload`), and `POST /node/pairing/window` with
  `Options.ConfiguredWindow` does too; both are admin routes.

### Changed

- Requires go-fabric v0.12.2. `NodeSource` gains `Window`,
  `SetupPayload`, `SessionDiagnostics`, `Clock` and `Observe` and drops
  `PairingCodes`, `Sessions` and `Subscriptions`; `*node.Node` satisfies
  it.
- The stream registers an observer with `Node.Observe` while it is open
  and removes it on close: a handler built after `Node.Start` reports
  commissionings, fabric removals and window transitions, each once
  (routes no longer publish them themselves).
- Fresh window credentials come from `node.NewWindowCredentials`; session
  idle ages from `Node.SessionDiagnostics` on the node's clock (0 for a
  session without traffic yet). The window's duration is
  `WindowSnapshot.Timeout`.
- The credentials of a window this handler opened are served while
  `Node.Window` reports that basic window (its discriminator, before its
  deadline), not on a timer of the handler's own.

### Deprecated

- `Events.Observers` returns its argument unchanged: `Node.Observe`
  replaces it, and a stream reports each transition once whether or not
  the host still wires it into `node.Config.Observers`. Delete that line;
  the method goes in v0.2.0. `Options.Events` stays for the host's own
  events.

## [0.1.1] - 2026-10-25

The first consumer's findings (go-ha2fabric, report H3): list-valued
fields and the host's configuration warnings (brief U1a). Additions only;
contract v1 keeps every field.

### Added

- **List-valued fields:** `Column.Multi` and `Filter.Multi` (JSON
  `multi`, optional, default false). A multi column shows its
  comma-separated value as chips; its value filter offers the individual
  values across all items instead of their combinations; a multi filter
  matches an item whose list holds the chosen value (values trimmed, an
  empty value never matches).
- **Host warnings:** `Options.Warnings func(ctx) []Warning` with
  `Warning{Code, Message, Since}`. `GET /node/status` answers them as
  `warnings` (absent without the hook, `[]` when it finds nothing); the
  Status page shows them in a card above the status card, the Selection
  page collapsed above the table. The npm client exports the `Warning`
  type.

## [0.1.0] - 2026-10-25

The node page family (wave U1 of concept v9), extracted from openccu-loom
at `b40b289a` and made host-neutral.

### Added

- **Contract v1** (`contract/openapi.yaml`): the node routes under
  `/api/v1/node/` — status, fabrics (list, remove), pairing (codes, open
  and close the commissioning window), endpoints, sessions, mDNS,
  compatibility, events, the selection (schema, list, preview, apply),
  maintenance (reassemble, factory reset with a written confirmation) and
  the server-sent event stream; the controller routes listed as planned.
  Golden fixtures under `contract/fixtures/` are read by the Go and the
  TypeScript tests; a guard fails on a removal against
  `contract/previous/`.
- **Go module** `fabricui`: `Mount` and `Handler` on the go-fabric node
  facade (`NodeSource`, which `*node.Node` satisfies), `ExposureProvider`
  for the host's selection, `Events` (wires `node.Observers` into the
  stream), the admin gate (`Options.Admin`, or a per-process token in
  token mode), the embedded bundle with SPA fallback and cache headers,
  an import guard holding the module on the facade, and an end-to-end test
  that commissions a fabrictest controller through a window the handler
  opened.
- **npm package** `@sukramj/fabric-ui`: the six node pages (Status, Pair,
  Fabrics with maintenance, Selection with preview before apply,
  Endpoints, Diagnostics), the UI kit, the design tokens with Home
  Assistant's variable names and the theme bridge, the client generated
  from the contract, the stream helper, English and German catalogues.
- **Standalone shell** built into `fabricui/dist/` and committed; CI
  rebuilds it and fails on any byte of difference.
- **Release**: one tag releases both artefacts; the npm publish runs when
  the `NPM_TOKEN` secret is set.
