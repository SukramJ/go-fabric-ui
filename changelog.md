# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
[Semantic Versioning](https://semver.org/). The Go module and the npm package
`@sukramj/fabric-ui` share one version.

## [Unreleased]

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
