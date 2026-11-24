# go-fabric-ui

[![ci](https://github.com/thingfabric/go-fabric-ui/actions/workflows/ci.yml/badge.svg)](https://github.com/thingfabric/go-fabric-ui/actions/workflows/ci.yml)
[![codeql](https://github.com/thingfabric/go-fabric-ui/actions/workflows/codeql.yml/badge.svg)](https://github.com/thingfabric/go-fabric-ui/actions/workflows/codeql.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

One web interface for every daemon built on
[go-fabric](https://github.com/thingfabric/go-fabric): health, pairing codes and
the commissioning window, fabrics, which host items are exposed, endpoints
and diagnostics on the node side; commissioning, discovery and the
commissioned nodes on the controller side. Maintained once, embedded by
each project.

The terms the pages use — the node family and the controller family,
Topology versus Network, the window kinds — and the Matter terms behind
them are defined in go-fabric's [glossary](https://docs.thingfabric.dev/reference/glossary/).

> **Status: `v0.14.1`** (contract 0.14.0) of
> [concept v9](https://github.com/thingfabric/go-fabric/blob/main/docs/concept-v9-fabric-ui.md).
> Both page families: the node family (status, pairing, fabrics,
> selection, endpoints, surface, diagnostics, topology) and the controller
> family (nodes, commissioning, a node's endpoints, attributes, ACL,
> bindings, window, ICD, fabrics and updates, settings, network,
> diagnostics, the onboarding inventory). The releases, one line per wave:
>
> - `v0.1.0`–`v0.1.3` — U1: the node family.
> - `v0.2.0`–`v0.3.0` — U2: the controller family; the patches carry the
>   surface page and go-fabric v0.14.0 to v0.16.0.
> - `v0.4.0`, `v0.4.1` — U3a, U3b: the node family's topology view and
>   the controller family's network graph.
> - `v0.5.0` — U4b, U4a, U5a: a node's inventory from the controller's
>   image, the node family's endpoint page, the event log of both
>   families; operation — node selector, Operation page, ping, QR scan,
>   pairing hints, live status, a Playwright smoke per page.
> - `v0.6.0`, `v0.6.1` — U6a: nodes created, renamed and deleted in the
>   page; go-fabric v0.18.0.
> - `v0.7.0` — U7: completeness against go-fabric v0.18.0 (attestation
>   error and boot reason, session parameters, re-announce, erase,
>   commissioning steps, terms and conditions, three node events).
> - `v0.7.1` — U8: the Commission page's BLE decision, `NodeSwitch`.
> - `v0.8.0` — U9: go-fabric v0.19.1, `terms_required` with a retry, the
>   host's counters, Thread border routers.
> - `v0.8.1` — U10: go-fabric v0.20.0, a node's operational host.
> - `v0.9.0`–`v0.9.2` — U11: the controller family's Diagnostics tab;
>   go-fabric v0.21.0 to v0.22.2.
> - `v0.10.0` — review phases 0 to 3: admin rights on the reading routes
>   and streams, a Host check on every route, strict request bodies,
>   `NewHandler`, stream limits; go-fabric v0.23.0; see the
>   [changelog](changelog.md).
> - `v0.11.0`, `v0.12.0` — review phase 4: the update check's errors per
>   node, a discovered device commissioned through go-fabric's pairing
>   pipeline, a code's commissioning bounded after the search; go-fabric
>   v0.24.1 in `v0.12.0`.
> - `v0.12.1` — go-fabric v0.25.0 under its new module path
>   `thingfabric.dev/go-fabric`.
> - `v0.13.0` — U12: the controller family's Inventory tab — onboarding
>   codes kept, shown, printed and attached (issue #68).
> - `v0.14.0` — M1: the move to `thingfabric.dev/go-fabric-ui`, the npm
>   scope `@thingfabric/fabric-ui` and the problem `type` URIs on
>   `thingfabric.dev` (contract 0.14.0).
> - `v0.14.1` — go-fabric v0.26.0.

## Two artefacts, one version

- **Go module** `thingfabric.dev/go-fabric-ui`, package `fabricui`:
  `Mount(mux, Options)` puts the HTTP API on a `*node.Node` and, through
  `fabricui/gofabric.ControllerOf`, on a `*controller.Controller`;
  `NewHandler(Options)` (or `Handler`, which panics on invalid options)
  serves the API and the built bundle, so a daemon
  without its own web app gets the whole interface from one import and
  needs no Node.js toolchain. What a host exposes is its own business: it
  implements `ExposureProvider` (schema, list, preview, apply) and the
  selection page renders it.
- **npm package** `@thingfabric/fabric-ui`: the Svelte 5 pages, the UI kit,
  the design tokens (Home Assistant's variable names, with a theme bridge
  that adopts the live theme inside an Ingress panel) and the generated
  TypeScript client, for a project with its own single-page app.

A Go daemon embeds the whole interface in one line, before or after the
node started:

```go
mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Exposure: provider}))
```

[`docs/embedding.md`](docs/embedding.md) covers the options, the npm
package and running behind an Ingress; [`docs/provider.md`](docs/provider.md)
how to write the selection's provider.

## The contract

`contract/openapi.yaml` is the single source of the API (routes under
`/api/v1/node/` and `/api/v1/controller/`; the controller handlers are
generated from it with oapi-codegen). The Go handlers
and the TypeScript client are tested against the same golden fixtures in
`contract/fixtures/`, and a guard fails on any removal against the last
release's copy. [`docs/contract.md`](docs/contract.md) has the route
table, the admin modes, the problem codes and the versioning rule.

## Building

```sh
make ci          # the main checks: build vet test lint generate-check web-check dist-check
make race        # the Go tests with the race detector
make fuzz        # each fuzz target for FUZZTIME (default 30s); go test runs their seeds
make help        # every target
make web-install # npm ci in ui/
make web         # rebuild fabricui/dist (committed; dist-check fails on any byte of difference)
```

`make ci` is not the whole pipeline: CI also runs `make race`,
`make cover-check`, `make vuln licenses secrets pins-check`,
`GOARCH=386 make build`, and after `make web-install` `make npm-audit`,
`make notices-check`, `make version-check`, `make contract-check` and
`make e2e-web`; CodeQL runs for Go and TypeScript. The nightly workflow
adds the macOS test leg, `make fuzz` at three minutes per target and a
full `npm audit` that reports without failing.

## Consumers

| Project | Uses |
| --- | --- |
| [go-ha2fabric](https://github.com/thingfabric/go-ha2fabric) | Go module with the bundle; Home Assistant exposure provider |
| [openccu-loom](https://github.com/SukramJ/openccu-loom) | npm package in its own app; Go module for the handlers |
| [go-fabric-server](https://github.com/thingfabric/go-fabric-server) | Go module with the bundle; both page families |

## Why "Matter" is not in the name

"Matter" and the Matter logo are trademarks of the Connectivity Standards
Alliance. Their use in a product, project or repository name, as on a
device, is reserved to Alliance members and certified products under the
*Connectivity Standards Alliance Trademark and Logo Usage Guidelines and
Terms* (<https://csa-iot.org/resources/governing-documents/>; the terms are
restated in go-fabric's
[`licenses/NOTICE-matter.js.txt`](https://github.com/thingfabric/go-fabric/blob/main/licenses/NOTICE-matter.js.txt):
"Only the Alliance and its members may use Alliance trademarks and logos,
including, without limitation, the Matter trademarks and logos"). This
project is neither, so the mark is not part of its name; the word appears
only descriptively, as in "a user interface for a Matter node", and never to
suggest certification or endorsement. Nothing built with this project is
certified by the Alliance, and using it does not convey any right to
describe a device, product or service as Matter compliant or certified.
See go-fabric's README, section "Not certified".

## Licence

Apache-2.0, see [`LICENSE`](./LICENSE). Contributions need the
[CLA](./CLA.md). [AI policy](./AI_POLICY.md).
