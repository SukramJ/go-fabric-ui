# Brief U2 — the controller family of go-fabric-ui (`v0.2.0`)

Date 2026-10-27. Concept: `docs/concept-u2-controller-family.md` (German;
its §F carries the owner's decisions of 2026-10-27). Fact map: the launcher's
`read-u2-controller-family.md` (go-fabric `v0.13.3`, go-fabric-ui `v0.1.3`,
go-fabric-server `v0.3.2`); every `file:line` there is a map, verify what you
build on. One wave, one pull request, one hand-back. **Model:** Opus.
Conventional commits, English, no attribution lines, no auto-merge, no label.
go-fabric's token rules (`/home/markus/Dokumente/GitHub/go-fabric/CLAUDE.md`
§"Token use"): read slices, guards in the background with a summary log,
a short report `docs/reports/U2.md` (per §, guards with HEAD, differences,
decisions, questions).

## 0 Decisions (owner, 2026-10-27)

- **0.1** REST handlers in this module over a port `ControllerSource`, with
  the adapter `fabricui.ControllerOf(*controller.Controller, ControllerOptions)`
  as the only place that knows go-fabric's facade. **The Go handler stubs are
  generated from the contract** (`oapi-codegen`: router, request/response
  types, validation); the TS client stays `openapi-typescript`. The node
  family stays hand-written in this wave.
- **0.2** One wave → `v0.2.0`.
- **0.3** Wi-Fi/Thread credentials can be created and deleted in the UI
  (admin, write-only; names readable, secrets never read back).
- **0.4** Targeted attribute reads of any path without admin; generic writes
  and invokes admin-only; no full inventory.

## 1 Contract 0.2.0 (`contract/openapi.yaml`)

The routes and admin flags of concept §C.2, with schemas for every request
and response (hex ids `*_hex`, Problem+JSON, `x-admin`). Keep the node
section byte-identical; `contract/previous/` stays the 0.1.3 baseline and
the removal test must pass. Fixtures under `contract/fixtures/` for every
controller response (recorded from the in-memory source, reviewed). Version
0.2.0 in the contract, `internal/version`, `ui/package.json`, lock file.

## 2 Go module

- `fabricui.ControllerSource`, `ControllerNode` and the small value types
  of concept §C.1 (no go-fabric type in the port; hex strings only at the
  wire, `uint64` in the port). `Options.Controller ControllerSource`
  (nil = no controller family; `/controller/*` answers 404 like unknown
  paths). `Events.Observers` removed (U1b Q1; both hosts already dropped
  it).
- `fabricui.ControllerOf(ctrl *controller.Controller, opts ControllerOptions)
  ControllerSource` in a sub-package that imports go-fabric's `controller`,
  `controller/client`, `controller/softwareupdate`, `controller/thread` and
  the generated `accesscontroldef`/`bindingdef` definitions (ACL and
  bindings are built from them and written through `Write`, as
  go-fabric-server's `cmd_acl.go` does). `ControllerOptions` carries the
  credential store (an interface the host implements; go-fabric-server has
  its settings store) and the `Observe` wiring from `client.Hooks`.
- Generated handlers: add `oapi-codegen` as a tool dependency (`go.mod`
  `tool` directive, pinned; `make generate` regenerates; a CI check fails
  when the generated file is stale, like `dist-check`). Hand-written only:
  the handler bodies that call the port and map errors to problems.
- An in-memory `ControllerSource` for tests and the standalone shell
  (`memory_controller.go`), with two nodes, one ICD, one pending update.
- Handler tests per route (admin refusal, hex ids, problem shapes), an
  e2e test over `fabrictest` with `ControllerOf` on a real
  `controller.Controller` commissioning a `fabrictest` device: nodes list,
  node detail, window, ACL write, read of an attribute, decommission.
- Secrets: PIN and credentials only in admin routes; never logged; the
  credentials GET answers names only.

## 3 npm package

- `ui/src/lib/pages/controller/`: Nodes, Commission, Node (tabs: Overview,
  Endpoints, Attributes, ACL, Bindings, Window, ICD, Fabrics, Updates),
  Settings — concept §C.3. Kit components only; DataTable for lists; the
  QR of a window via `renderQrSvg`.
- `FabricClient` gains the controller methods (generated types); a
  `createControllerStore`; SSE `/controller/stream` through the existing
  stream helper.
- Shell: families by presence (`/node/status` and `/controller/status`;
  404 hides a family), hash routes `#/controller/<page>`; i18n keys
  `controller.*` in en and de with the coverage test.
- Vitest per page, the contract fixtures read by the TS tests as today.

## 4 Docs

`docs/embedding.md` (the controller option, `ControllerOf`, the credential
store), `docs/contract.md` (the controller section replaces "planned"),
`docs/provider.md` untouched, `README.md` family list, `CONCEPT.md` wave row
U2 → `v0.2.0`, changelog `## [0.2.0] - <date>` (Added: the family; Removed:
`Events.Observers`; Changed: contract 0.2.0).

## 5 Guards

`make ci` (web-check, vitest, dist-check), race, cover-check,
contract-check (against the 0.1.3 baseline), version-check, notices-check,
licenses, vuln, secrets, the generated-handler staleness check, `npm run
build` with the bundle committed. `df -h /` first; `go clean -cache
-testcache` only when disk is tight.

## 6 Out of scope

go-fabric-server binding the family (S4), the add-on, camera-based QR
scanning, OTBR/Thread border-router pages beyond the dataset names, the
node family's move to generated handlers.
