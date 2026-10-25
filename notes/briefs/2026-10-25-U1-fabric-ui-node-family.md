# Brief U1 — go-fabric-ui: the node page family (contract v1, Go module, Svelte library, standalone shell, release)

Date 2026-10-25. Repository `/home/markus/Dokumente/GitHub/go-fabric-ui`
(main = skeleton commit; read `README.md`, `CONCEPT.md`, `CLAUDE.md`
first). Design: go-fabric `docs/concept-v9-fabric-ui.md` (German; §0
decisions, B architecture, B.4 route table, B.5 build and release, D, E) —
read it in full. Read map `read-u1-loom-matter-ui.md` (copy to
`notes/reads/` in the first commit; a map with file:line, verify before
acting). Source of the pages: **openccu-loom at `b40b289a`**
(`/home/markus/Dokumente/GitHub/openccu-loom`; read only via
`git -C … show b40b289a:<path>` or the checkout, never modify it): the five
pages `assets/ui/src/routes/matter/*.svelte`, store
`assets/ui/src/lib/stores/matter.svelte.ts`, types `api/matter-types.ts`,
handlers `internal/north/rest/handlers/matter*.go` (+ 72 tests), kit
`assets/ui/src/lib/components/ui/`, tokens `assets/ui/src/app.css:60-128`,
theme bridge `assets/ui/src/lib/theme/ha-bridge.ts`, `lib/qr.ts`,
`lib/i18n.ts` (the `matter.*` strings), vitest setup. Loom is Apache-2.0
with the same owner; porting is allowed, cite `file:line` at `b40b289a` in a
comment at the top of every ported file. go-fabric `v0.12.0`
(`/home/markus/Dokumente/GitHub/go-fabric`): the `node.Node` facade
(`node/node.go`, `node/commissioning.go`, `node/credentials.go`,
`node/health.go`, `node/diagnostics.go`), `eligibility.Verdict`,
`fabrictest` for an end-to-end test, `docs/guides/observability.md` and
`docs/guides/ports.md` for what the facade offers.

**Model:** Opus. Branch `u1/node-family`, conventional commits, one pull
request, no auto-merge, no label; the owner merges and tags `v0.1.0`.
English. Never add attribution lines. go-fabric's `CLAUDE.md` token rules
apply. **Hard rules:** (1) Go imports only go-fabric's facade (`node`,
`eligibility`, `contract`, `clock` if needed); an import guard
`internal/arch` holds it; a facade gap is a Question for wave CK, not a
seam. (2) No host terms (Home Assistant, CCU, MQTT, central, channel) in
code, routes, strings or docs; the selection page is driven by
`ExposureProvider.Schema`. (3) "matter" is not in package, route or
file names (it may appear descriptively in prose). (4) 64-bit ids as hex
strings. (5) Dependencies: Go — go-fabric and stdlib only; npm — pinned
exact versions, MIT/Apache-2.0/BSD, every runtime package in
`THIRD-PARTY-NOTICES.md` with its licence text under `licenses/` when the
licence requires it. (6) Secrets (passcode) only in admin routes, never in
logs; tests assert it.

## 1 Contract (`contract/`)

- `openapi.yaml` 3.1, sections `node` (this wave) and `controller` (paths
  listed with `x-status: planned`, no schemas yet). Routes exactly as the
  concept's B.4 table: `/api/v1/node/status`, `/fabrics`, `/fabrics/{index}`
  (DELETE), `/pairing` (admin), `/pairing/window` + `/window/close`
  (admin), `/endpoints`, `/sessions`, `/mdns`, `/compatibility`, `/events`,
  `/exposure/schema`, `/exposure` (GET, PUT admin), `/exposure/preview`
  (POST admin), `/maintenance/reassemble`, `/maintenance/factory-reset`
  (admin, confirmation body), `/stream` (SSE). Schemas from loom's
  `matter-types.ts` and the handler structs, host-neutral: `endpoint`
  carries `name` and an opaque `source` map instead of
  `device_address`/`channel_address`; `exposure` items are
  `{key, name, group, fields: {col: string}, verdict: {state, device_type,
  device_type_name, clusters, reason}, enabled, source}`; `schema` is
  `{columns: [{key, label, filter: bool}], groups: bool, filters: [...]}`.
  Every 64-bit field has a `_hex` twin; numeric twins are documented as
  informational.
- `fixtures/*.json`: one request/response pair per route and status (200,
  204, 400, 403, 409, 503). The Go handler tests and the TS client tests
  read these files; a test fails when a fixture is unused by either side.
- A contract guard: a test that diffs `openapi.yaml` against the previous
  tag's copy (committed under `contract/previous/`) and fails on a removed
  path, method or response field.

## 2 Go module (`fabricui/`)

- `Options{Node NodeSource; Exposure ExposureProvider; Admin func(*http.Request) bool;
  Reassemble func(ctx) error; Events EventSource; Prefix string; Logger *slog.Logger}`,
  `Mount(mux *http.ServeMux, o Options) error`, `Handler(o Options) http.Handler`
  (API + the embedded bundle with SPA fallback and cache headers as loom's
  `internal/north/ui/spa.go`), `ExposureProvider`, `Schema`, `Item`,
  `Update`, `Verdict` (= `eligibility.Verdict`'s fields), `NodeSource`
  (the method subset of `*node.Node`; compile-time assertion in a test
  that `*node.Node` satisfies it).
- Handlers ported from loom's `matter*.go` with their tests (72): same
  statuses and bodies, renamed; `Admin` gate on every writing route and on
  `/pairing`; `503 service_unready` when a dependency is nil; the
  factory-reset confirmation body; window open 409 when open; durations
  bounded as loom bounds them (cite). Where loom reads a bridge seam
  (sessions occupancy, mDNS advertiser), read go-fabric's facade
  (`Node.Health`, diagnostics, `node/metrics`); what the facade does not
  offer is a Question (CK) and the route answers what it can with the
  field documented as optional.
- SSE `/stream`: events fabric_added/removed, window_opened/closed,
  endpoints_changed, exposure_changed from `node.Observers`/the facade's
  hooks plus `EventSource` for the host's exposure changes; heartbeat
  comment every 15 s; a test with `httptest`.
- `internal/arch` import guard; `internal/version` constant `0.1.0`
  (`make version-check` ties it to `ui/package.json` and `changelog.md`).
- A `fabricui/dist/` placeholder page ("bundle not built") until §4
  commits the real bundle; `Handler` serves it with 503 as loom does.
- End-to-end on `fabrictest`: a site node, `Handler` on `httptest`, GET
  status/fabrics/pairing, open window, a site controller commissions
  through the opened window, fabrics shows one entry, DELETE removes it,
  SSE delivered fabric_added and fabric_removed.

## 3 Svelte library and shell (`ui/`)

- `package.json` name `@sukramj/fabric-ui`, version `0.1.0`, `type: module`,
  `exports` for the library (pages, kit, theme, client, i18n) and `files`;
  Svelte 5, Vite, Tailwind 4, `@tailwindcss/vite`, `qrcode`, `lucide-svelte`
  (or `@lucide/svelte`), `openapi-typescript`, vitest + testing-library +
  happy-dom — versions pinned exactly (no `^`), taken from loom's
  `assets/ui/package.json` where they exist; `bits-ui` only if a kit
  component needs it (say which). `.npmrc` only if required (loom's
  legacy-peer-deps note).
- `src/lib/kit/`: the kit components the pages use (Badge, Button, Card,
  DataTable, EmptyState, ErrorState, LoadingState, Icon, Input, PageHeader,
  PageShell, Select, Tabs), ported from loom with their tests where they
  exist.
- `src/lib/theme/tokens.css` (loom `app.css:60-128` verbatim semantics;
  the `--ha-*` names, dark block, HA skin) and `ha-bridge.ts` ported;
  pages use the tokens, not raw Tailwind colours (loom's pages use
  `text-slate-*`; port them onto the tokens so the HA skin applies).
- `src/lib/i18n/`: `en.json`, `de.json` with the `matter.*` strings
  renamed `node.*`; `t()` as loom's, provided through Svelte context by
  the host (the shell provides the default); a coverage test that every
  key used by a page exists in both catalogues.
- `src/lib/api/`: client generated from `contract/openapi.yaml`
  (`npm run gen:types`; the generated file is committed and a test fails
  when stale), `createClient(baseUrl, fetch)`; SSE helper.
- `src/lib/pages/node/`: Status (loom `Matter.svelte` content without the
  tabs), Pair, Fabrics (+ maintenance), Selection (loom
  `MatterExposureList` made generic over `schema`; search, filters from
  `schema.filters`, grouping when `schema.groups`, dirty map, Preview
  before Apply with the changed verdicts shown, save/discard bar),
  Endpoints, Diagnostics (sessions, mDNS, compatibility, events). Each
  page takes `client` and host hooks (`toast`, `confirm`) through context.
  Tests ported from loom's `*.test.ts` (vitest, happy-dom) against the
  fixtures.
- `src/shell/`: the standalone app — `index.html`, hash routing, a
  header with the page tabs, `PageShell`, toast and confirm, the theme
  bridge started, the admin header token read from `/node/status`
  (`csrf_token` field; the Go side issues a per-process random token and
  requires it on writes when `Options.Admin` is nil — document both
  modes). `vite.config.ts`: `base: './'`, `outDir ../fabricui/dist`,
  no sourcemaps in the committed output, vendor chunk.
- Docs under `ui/README.md`: embedding the library in an existing app
  (what loom will do: provide `t`, toast, confirm, mount the pages, feed
  SSE into its hub).

## 4 Build, guards, release

- `Makefile`: `web-install` (`npm ci`), `web` (build into `fabricui/dist`),
  `web-check` (svelte-check + vitest), `dist-check` (rebuild into a temp
  dir and `diff -r` against the committed `fabricui/dist`; fails on any
  byte), `version-check`, `ci` = build vet test lint web-check dist-check.
- `.github/workflows/ci.yml`: add `web` (node pinned, `npm ci`, check,
  test, build, `dist-check`), `coverage` (floors as the siblings),
  `fuzz-smoke` if a decoder exists (else say none), `contract` (fixture
  use on both sides, openapi previous-tag diff). Keep the job names the
  branch protection requires (`build`, `lint`, `test (…)`, `codeql (go)`,
  `security (vuln + licenses + secrets)`); list new required names in the
  report so the owner adds them.
- `release-on-tag.yml` in the family's form: tag == `internal/version` ==
  `ui/package.json` == changelog section; GitHub release with notes;
  `npm publish --access public --provenance` with `NPM_TOKEN` — the step
  is conditional on the secret being set and says so in the summary when
  it is skipped (the owner adds the token later).
- `THIRD-PARTY-NOTICES.md` with every runtime npm package (from the built
  bundle's `vendor` chunk list) and the Go modules; `licenses/` texts.
- Dependabot: add the `npm` ecosystem for `/ui`.

## 5 Docs and report

`README.md` (status to first release; the two artefacts; embedding in
three lines for a Go daemon; the contract), `docs/embedding.md` (Go
module, npm, Ingress notes: same-origin theme bridge, admin gate by the
host), `docs/contract.md` (route table, versioning rule), `docs/provider.md`
(writing an `ExposureProvider`, with the schema/preview semantics and an
in-memory example used by the tests), `changelog.md` `## [0.1.0]`. Report
`docs/reports/U1.md` (and PR body): per § what landed with commit hashes;
guards with HEAD (`make ci`, `race`, `dist-check`, `web-check`, the
contract guard, the e2e); differences from the brief; decisions; left out;
**questions** (go-fabric facade gaps with the symbol needed → CK; anything
loom did that needs an owner decision); the list of required-check names.
Copy this brief to `notes/briefs/2026-10-25-U1-fabric-ui-node-family.md`
in the first commit. If a permission is denied, say which and stop that
step; never ask the launcher to perform it. Hand back with the PR number,
HEAD, and the Questions section verbatim.
