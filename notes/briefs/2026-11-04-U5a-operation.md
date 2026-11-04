<!-- go-fabric-ui issue #39, copied verbatim as the wave's brief. -->

# Brief U5a — operation: the UI half (go-fabric-ui)

Date 2026-11-04. Concept `docs/concept-u5-operation.md` (German; read A–F
in full first). This wave is D.1: contract 0.5.0, the ports, the node
selector, the operation page, ping, the stream events, the QR scan, the
ecosystem hints and the Playwright smoke per page — everything the UI half
can prove with test doubles before the go-fabric `logring` package
(go-fabric issue #393, wave D.2) and the server wave (D.3) exist.
**Model:** Opus. Repo `/home/markus/Dokumente/GitHub/go-fabric-ui`,
worktree `../go-fabric-ui-u5a`, branch `u5/operation`, cut from main after
the concept merged. Pins: go-fabric main 19588e9b (`go doc` after a bump
only if the build needs one; `Node.OnHealthSample` `node/health.go:93`,
`Node.OnDiagnosticEvent` `node/diagnostics.go:40` exist there), matterjs-server
0bd06f22 for the comparison only, matter.js `docs/ECOSYSTEMS.md:85-131` at
862ad75f for the hint texts. Map `~/.cache/go-fabric-scratch/read-u5.md`
(a map, not a citation). U4a's brief
`notes/briefs/2026-11-04-U4a-node-details-events.md` is the pattern. One
wave, one pull request, one hand-back. No tag; version stays **0.5.0
unreleased**; contract **0.5.0** (`contract/openapi.yaml:4`),
`contract/previous/openapi.yaml` = the 0.4.1 copy.

## 0 Working rules

Commits as `SukramJ <sukramj@icloud.com>` (`git var GIT_AUTHOR_IDENT`
first), `<area>: <what> (U5a)`, no attribution lines, stage named paths.
Guards under `flock ~/.cache/go-fabric-scratch/guards.lock` with
`export PATH=$HOME/go/bin:$PATH GOTMPDIR=$HOME/.cache/gotmp`, logs and
`summary.log` in `~/.cache/go-fabric-scratch/u5a/`. No change in go-fabric
or go-fabric-server: a gap there is an issue text in the hand-back. Every
go-fabric line you build on is read at 19588e9b and cited `file:line`.
Official text in English.

## 1 Contract 0.5.0 (additive; concept C.1)

- **Node selector.** `GET /nodes` → `NodeSummaryList{nodes[]{key, label,
  kind, enabled}}`; every node route also under `/nodes/{node}/…` (same
  operation, `operationId` suffixed `ForNode`, parameter `NodeKey`), the
  `/node/*` routes stay and mean the default node. Node stream per node
  (`/nodes/{node}/stream`).
- **Operation.** `GET /operation/log/level` → `LogLevel{level, levels[]}`,
  `PUT /operation/log/level` (`x-admin`), `GET /operation/log` →
  `LogRecordList{records[]{seq, time, level, msg, attrs}}` with `since`,
  `limit` (default 200, max 2000), `GET /operation/log/stream` (SSE
  `log_record`, heartbeat as the other streams), `GET /operation/info` →
  `OperationInfo{version, go_fabric, uptime_s, started_at}`. 503 when no
  `Operations` port is attached (as the node family answers without a
  node).
- **Ping.** `POST /controller/nodes/{id}/ping` (`x-admin`) →
  `PingResult{method: "matter-read", addresses[]{ip, reachable, rtt_ms?}}`;
  404 unknown node.
- **Stream.** Node stream kinds `health_sample {healthy, note, sampled_at}`
  and `diagnostic_event` (the `DiagnosticEvent` shape the trace route
  serves); `docs/contract.md` stream table.
- `NodeStatus.exposure_reason` (optional string): why the host has no
  selection, shown by the Selection page.
- `make generate`, `make gen-types`, fixtures recorded and reviewed, every
  fixture used on both sides.

## 2 Go half (concept C.2)

- `NodeSource` (`fabricui/options.go:25`) **gains** `OnHealthSample(func(node.HealthSample)) (unsubscribe func())`
  and `OnDiagnosticEvent(func(node.DiagnosticEvent)) (unsubscribe func())`
  — breaking for hosts (owner decision F.3), `changelog.md` `### Changed`
  **Breaking** line with the one-line migration (`*node.Node` satisfies it;
  a wrapper forwards). `fabricui.NodeObservers` publishes both as stream
  events.
- `Options.Nodes map[string]NodeSource` + `Options.NodeLabels` (or one
  `NodeEntry{Source, Label, Kind}` slice — pick the one shape and say why);
  `Options.Node` stays the default node; handlers registered per key; the
  test doubles cover two nodes.
- New port `fabricui.Operations` (`Options.Operations`): `LogLevel()
  slog.Level`, `SetLogLevel(slog.Level) error`, `Log(since uint64, limit int)
  []LogRecord`, `ObserveLog(func(LogRecord)) (unsubscribe func())`,
  `Info() OperationInfo`; `fabricui.MemoryOperations` as the double (ring of
  N, level var); the SSE stream from `ObserveLog` with the existing stream
  machinery (buffer, slow-client drop).
- `ControllerNode.Ping(ctx) (PingResult, error)`: adapter
  `fabricui/gofabric/node.go` reads BasicInformation DataModelRevision
  (0x0028/0x0000) once per known address of the node through the existing
  targeted read, measures the round trip; `MemoryController` answers a
  fixture result.
- Tests: handler tests per route and status, two-node registration, stream
  events, the ping adapter against a `fabrictest` pair in
  `fabricui/gofabric/e2e_test.go` (one case: ping answers reachable for
  the commissioned node).

## 3 Svelte half (concept C.3)

- Shell: node selector in the header when `/nodes` lists more than one;
  routes `#/node/<key>/…` (default node without key); store per node.
- Page **Operation** (both families, `pages/operation/Operation.svelte`):
  level select (admin), log view (ring + SSE append, pause, level/text
  filter, copy), info card; 503 → the known empty state.
- Status and Diagnostics reload on `health_sample` / `diagnostic_event`
  (coalesced); Refresh stays.
- Controller node page: Ping button (admin) with a per-address result and
  the sentence that it is a Matter read.
- Commission page: "Scan" button → lazy chunk `ui/src/lib/scan/` loading a
  decoder (MIT/Apache-2.0; `THIRD-PARTY-NOTICES.md` + `licenses/npm/`,
  `notices-check`), native `BarcodeDetector` when present; on an insecure
  context an explanation instead of the button; the result fills `code`.
  `make web` in the same commit; `dist-check` green; report the bundle
  size before/after (main chunk flat, scan chunk separate).
- Pair page: ecosystem hints (Apple, Google, Alexa, SmartThings, Home
  Assistant) as static copy in `en.json`/`de.json` from ECOSYSTEMS.md:85-131
  with links, no version numbers; the existing compatibility beside them.
- Selection page without exposure: shows `exposure_reason`.
- Playwright: `ui/e2e/` (Chromium only, `@playwright/test` devDependency),
  one smoke per page (loads, renders the fixture, no console errors)
  against a Go e2e server (`cmd` or test binary serving `fabricui.Handler`
  with `MemoryController`, `MemoryOperations` and the node double);
  `make e2e-web`; CI job `e2e-web` in `.github/workflows/ci.yml` (browser
  cached). `TestWorkflow…`-style guard if one exists for jobs.
- i18n keys in both catalogues, coverage lines; vitest page tests per new
  page/component.

## 4 Docs and changelog

`docs/contract.md` (routes, selector, streams, 0.5.0), `docs/embedding.md`
(the `NodeSource` additions, `Options.Nodes`, `Operations` port, the
`logring` sentence pointing at go-fabric #393), `changelog.md` `[0.5.0]`
`### Added` / `### Changed` (Breaking line), `README.md` feature list,
`CONCEPT.md` row already there.

## 5 Guards

`make ci`, `make race`, `make e2e`, `make e2e-web`, `make cover-check` if
present, `make contract-check`, `make notices-check`, `make version-check`
— under the flock, logs `u5a/`, summary `u5a/summary.log`.

## 6 Out of scope

The `logring` package (go-fabric #393), the server's multi-node hosting
and wiring (wave D.3), ICMP, log files, stream values, full Playwright
flows, the tag.

## Hand-back

Per §: what landed (files, routes, components, tests by name); guards
table (guard, HEAD, exit); bundle sizes; differences from this brief;
decisions; left out; open questions; go-fabric / go-fabric-server issue
texts for gaps; every go-fabric line cited, re-read at 19588e9b. Open the
pull request against main (English, hand-back as body, no labels,
auto-merge off) and hand back with its number.

