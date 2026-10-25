# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

One web interface for every daemon built on
[go-fabric](https://github.com/SukramJ/go-fabric), maintained here and
embedded by the projects (go-ha2fabric, go-fabric-server, openccu-loom).
`CONCEPT.md` is the short design; the full one is go-fabric's
`docs/concept-v9-fabric-ui.md`. **State: concept** (created 2026-10-25);
wave U1 delivers the node page family as `v0.1.0`.

## Commands

```sh
make ci        # build vet test lint — what the pipeline runs for the Go module
make race      # go test -race (CGO_ENABLED=1 for this run only)
make fmt       # gofumpt -w .
```

The web half lives under `ui/` (npm scripts: `dev`, `build`, `check`,
`test`); `make web` builds it into `fabricui/dist/`, `make dist-check`
fails when the committed bundle differs from a fresh build. (U1 adds them.)

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
- **The bundle is committed** (`fabricui/dist/`) and guarded: pinned
  versions, `npm ci`, byte-identical rebuild in CI. Sourcemaps are not
  committed.
- **Dependencies are decisions.** Go: go-fabric and the standard library.
  npm: MIT/Apache-2.0/BSD only; every package that lands in the bundle is
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
