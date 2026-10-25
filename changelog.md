# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
[Semantic Versioning](https://semver.org/). The Go module and the npm package
`@sukramj/fabric-ui` share one version.

## [Unreleased]

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
