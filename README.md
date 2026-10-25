# go-fabric-ui

[![ci](https://github.com/SukramJ/go-fabric-ui/actions/workflows/ci.yml/badge.svg)](https://github.com/SukramJ/go-fabric-ui/actions/workflows/ci.yml)
[![codeql](https://github.com/SukramJ/go-fabric-ui/actions/workflows/codeql.yml/badge.svg)](https://github.com/SukramJ/go-fabric-ui/actions/workflows/codeql.yml)
[![License: Apache-2.0](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](./LICENSE)

One web interface for every daemon built on
[go-fabric](https://github.com/SukramJ/go-fabric): health, pairing codes and
the commissioning window, fabrics, which host items are exposed, endpoints
and diagnostics on the node side; commissioning, discovery and the
commissioned nodes on the controller side. Maintained once, embedded by
each project.

> **Status: concept.** This repository was created from
> [concept v9](https://github.com/SukramJ/go-fabric/blob/main/docs/concept-v9-fabric-ui.md)
> on 2026-10-25; wave U1 delivers the node page family. Nothing runs yet.

## Two artefacts, one version

- **Go module** `github.com/SukramJ/go-fabric-ui`, package `fabricui`:
  `Mount(mux, Options)` puts the HTTP API on a `*node.Node`;
  `Handler(Options)` serves the API and the built bundle, so a daemon
  without its own web app gets the whole interface from one import and
  needs no Node.js toolchain. What a host exposes is its own business: it
  implements `ExposureProvider` (schema, list, preview, apply) and the
  selection page renders it.
- **npm package** `@sukramj/fabric-ui`: the Svelte 5 pages, the UI kit,
  the design tokens (Home Assistant's variable names, with a theme bridge
  that adopts the live theme inside an Ingress panel) and the generated
  TypeScript client, for a project with its own single-page app
  (openccu-loom).

The contract between them is `contract/openapi.yaml`; both sides are tested
against the same fixtures.

## Consumers

| Project | Uses |
| --- | --- |
| [go-ha2fabric](https://github.com/SukramJ/go-ha2fabric) | Go module with the bundle; Home Assistant exposure provider |
| [openccu-loom](https://github.com/SukramJ/openccu-loom) | npm package in its own app; Go module for the handlers |
| [go-fabric-server](https://github.com/SukramJ/go-fabric-server) | Go module with the bundle; both page families |

## Why "Matter" is not in the name

"Matter" and the Matter logo are trademarks of the Connectivity Standards
Alliance. Their use in a product, project or repository name, as on a
device, is reserved to Alliance members and certified products under the
*Connectivity Standards Alliance Trademark and Logo Usage Guidelines and
Terms* (<https://csa-iot.org/resources/governing-documents/>; the terms are
restated in go-fabric's
[`licenses/NOTICE-matter.js.txt`](https://github.com/SukramJ/go-fabric/blob/main/licenses/NOTICE-matter.js.txt):
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
