# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

One web interface for every daemon built on
[go-fabric](https://github.com/SukramJ/go-fabric), maintained here and
embedded by the projects (go-ha2fabric, go-fabric-server, openccu-loom).
`CONCEPT.md` is the short design; the full one is go-fabric's
`docs/concept-v9-fabric-ui.md`. **State:** wave U1 (the node page family)
is `v0.1.0`, patched in `v0.1.1` (brief U1a) and `v0.1.2` (brief U1b, go-fabric v0.12.2's facade routes) and `v0.1.3` (brief F2, list-valued fields); wave U2 adds the controller family.

## Commands

```sh
make ci              # build vet test lint web-check dist-check — what the pipeline runs
make race            # go test -race (CGO_ENABLED=1 for this run only)
make e2e             # the end-to-end test: a fabrictest controller commissions through the handler
make fmt             # gofumpt -w .
make web-install     # npm ci in ui/
make web             # rebuild the committed bundle fabricui/dist
make dist-check      # rebuild into a temp dir, fail on any byte of difference
make contract-check  # no removal against contract/previous, every fixture used on both sides
make notices-check   # every bundled npm package is in THIRD-PARTY-NOTICES.md
make version-check   # internal/version == ui/package.json == changelog section
```

The web half lives under `ui/` (npm scripts: `dev`, `build`, `check`,
`test`, `gen:types`). A change under `ui/src/lib` or `ui/src/shell` is
followed by `make web` in the same commit, or `dist-check` fails. A change
to `contract/openapi.yaml` is followed by `make gen-types`; the fixtures'
responses are recorded from the handlers with
`go test ./fabricui -run <Test> -update` and reviewed before committing.

## Rules

- **Two halves, one contract.** `contract/openapi.yaml` is the single
  source; Go handlers and the TypeScript client are tested against the
  same fixtures under `contract/fixtures/`. A route is never removed in a
  minor version.
- **Go imports only go-fabric's facade** (`node`, `eligibility`,
  `contract`), never a package below it (`bridge`, `secure`, `im`, …). A
  facade gap is reported to go-fabric (wave CK), not worked around through
  a seam. An import guard (`internal/arch`) holds it.
- **No host knowledge.** What is exposable, how it is named and grouped,
  comes through `ExposureProvider`; the pages render schema and items.
  Home Assistant, CCU and MQTT terms do not appear in this repository.
  Two exceptions, decided 2026-10-25 (brief U1): the design tokens keep
  Home Assistant's CSS variable names and the theme bridge detects an
  Ingress panel, because that is how the pages adopt a host theme; and the
  ecosystem labels on the compatibility page name the controller vendors
  (Apple Home, Google Home, Amazon Alexa, Home Assistant) as go-fabric's
  vendor table names them.
- **The bundle is committed** (`fabricui/dist/`) and guarded: pinned
  versions, `npm ci`, byte-identical rebuild in CI. Sourcemaps are not
  committed.
- **Dependencies are decisions.** Go: go-fabric and the standard library.
  npm: MIT/Apache-2.0/BSD/ISC only (ISC accepted 2026-10-25 for
  `@lucide/svelte`, as the family accepted it for coder/websocket); every
  package that lands in the bundle is
  listed in `THIRD-PARTY-NOTICES.md`.
- **Security.** Writing routes only with `Options.Admin`; the standalone
  shell sends a header token on writes; secrets (passcodes) appear only in
  admin routes and never in logs.
- **64-bit identifiers** travel as hex strings in JSON; a JS number never
  carries a node or fabric id.
- **Trademark.** "Matter" is a trademark of the Connectivity Standards
  Alliance and is not part of this project's name, its packages or its
  routes; it appears only descriptively. See the README.
- **Language.** Everything in the repository is English. No attribution
  lines in commits or pull requests.
- **Every new `.go` file** opens with the Apache-2.0 SPDX header; `make
  lint` and the licence-header test fail otherwise.

## The reference for the pages

openccu-loom at `b40b289a` (`assets/ui/src/routes/matter/*.svelte`,
`assets/ui/src/lib/stores/matter.svelte.ts`, `internal/north/rest/handlers/matter*.go`,
`assets/ui/src/app.css`, `assets/ui/src/lib/theme/ha-bridge.ts`) is where
the node pages come from; cite `file:line` at that commit when porting.
It is not a gold standard for Matter behaviour — that stays matter.js,
reached only through go-fabric.
