# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

One web interface for every daemon built on
[go-fabric](https://github.com/SukramJ/go-fabric), maintained here and
embedded by the projects (go-ha2fabric, go-fabric-server, openccu-loom).
`CONCEPT.md` is the short design; the full one is go-fabric's
`docs/concept-v9-fabric-ui.md`. **State:** wave U1 (the node page family)
is `v0.1.0`, patched in `v0.1.1` (brief U1a) and `v0.1.2` (brief U1b, go-fabric v0.12.2's facade routes) and `v0.1.3` (brief F2, list-valued fields); wave U2 (brief U2) adds the controller family in `v0.2.0`; wave U3a (brief U3a) adds the node family's topology view in `v0.4.0`; wave U3b (brief U3b) the controller family's network graph in `v0.4.1`; wave U4b (brief U4b) the controller family's inventory and wave U4a (brief U4a) the node family's endpoint page and the event log of both families, and wave U5a (brief U5a) operation — node selector, Operation page, ping, QR scan, pairing hints, live status, Playwright smoke — in `v0.5.0`; wave U6a (issue #47, concept `docs/concept-u6-nodes.md`) the node registry — nodes created, renamed and deleted in the page, per-node selection, the configuration block — in `v0.6.0`; `v0.6.1` carries go-fabric v0.18.0; wave U7 (brief U7, issue-free: the inventory `notes/reads/read-u7-gofabric-coverage.md`) completeness against go-fabric v0.18.0 — reannounce, erase, status and session fields, three node events, pairing step and terms — in `v0.7.0` (contract 0.7.0).

## Commands

```sh
make ci              # build vet test lint web-check dist-check — what the pipeline runs
make race            # go test -race (CGO_ENABLED=1 for this run only)
make e2e             # the end-to-end test: a fabrictest controller commissions through the handler
make e2e-web         # the Playwright smoke per page (Chromium) against fabricui's TestServeE2EWeb
make fmt             # gofumpt -w .
make web-install     # npm ci in ui/
make web             # rebuild the committed bundle fabricui/dist
make dist-check      # rebuild into a temp dir, fail on any byte of difference
make contract-check  # no removal against contract/previous, every fixture used on both sides
make notices-check   # every bundled npm package is in THIRD-PARTY-NOTICES.md
make version-check   # internal/version == ui/package.json == changelog section
make generate        # the controller handlers (internal/api) from the contract, pinned oapi-codegen
make generate-check  # fail when internal/api is stale
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
  a seam. The one wider place is `fabricui/gofabric`, the controller
  adapter: go-fabric's controller facade (`controller`, `controller/client`,
  `controller/softwareupdate`, …), the Thread dataset codec (`thread`) and
  the generated cluster definitions it encodes ACL and Binding with. An import guard
  (`internal/arch`) holds both.
- **Controller handlers are generated.** `internal/api` is oapi-codegen's
  strict server for the contract's `controller` tag; regenerate with
  `make generate` after a contract change, never hand-edit it. The handler
  bodies (`fabricui/handlers_controller.go`) call the port and map errors
  to problems; the node family stays hand-written.
- **No host knowledge.** What is exposable, how it is named and grouped,
  comes through `ExposureProvider`; the pages render schema and items.
  Home Assistant, CCU and MQTT terms do not appear in this repository.
  Two exceptions, decided 2026-10-25 (brief U1): the design tokens keep
  Home Assistant's CSS variable names and the theme bridge detects an
  Ingress panel, because that is how the pages adopt a host theme; and the
  ecosystem labels on the compatibility page name the controller vendors
  (Apple Home, Google Home, Amazon Alexa, Home Assistant) as go-fabric's
  vendor table names them; the pairing hints (brief U5a) add SmartThings
  and quote matter.js `docs/ECOSYSTEMS.md`; the several-nodes hint on a
  managed node's status page (U6a) quotes home-assistant-matter-hub's
  bridge configuration page and links it.
- **The bundle is committed** (`fabricui/dist/`) and guarded: pinned
  versions, `npm ci`, byte-identical rebuild in CI. Sourcemaps are not
  committed.
- **Dependencies are decisions.** Go: go-fabric, the standard library and
  oapi-codegen's runtime (the generated handlers; decided in brief U2);
  the generator is a pinned `tool`.
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

## Which model does which work

As go-fabric's `CLAUDE.md` (owner, 2026-10-10; the review split 2026-11-08,
go-fabric #456), adapted: a statement about go-fabric's facade, the
reference pages (the section above) or the contract that lands in the repo
(`docs/contract.md`, `docs/embedding.md`, `docs/provider.md`, a concept
page, a report's decisions) is cited later as fact by the hosts, so the work
that produces such statements gets the strongest model, and the work a guard
checks may use a cheaper one.

| Work | Model |
| --- | --- |
| Analysis, planning, briefs, the wave report, the contract's design (a route, a problem code, an event's payload, a port on `fabricui.Options`); anything written into `docs/`, `CONCEPT.md` or a comment that cites a go-fabric, matter.js or reference-page line; rebases and merge conflicts where upstream may have overtaken the change; a question or hand-over issue to go-fabric | **Fable** (the default for this repo) |
| Implementation of a finished brief with guards behind it: handlers and the `gofabric` adapter against a contract the brief fixes, Svelte pages and stores, fixtures used on both contract sides, `make generate` and `make web`, a dependency bump, a `changelog.md` entry for a finished diff | **Opus** (also what fast mode runs on) |
| **Review of a wave's pull request, split (owner decision 2026-11-08, go-fabric #456).** The code review — correctness, concurrency, error paths, tests, the fixed report shape — is **Opus**. On top of it **Fable** reads, bounded: every new citation line against the pinned source (go-fabric at its tag, matter.js and the reference pages at their pins), every line that enters `docs/contract.md`, `docs/embedding.md` or `docs/provider.md`, the admin and token paths of the handler, and the brief's decisions against go-fabric and the hosts — the behavioural gaps a code review does not see. The load-bearing claims of the Opus review are verified before they reach the implementer. | **Opus**, then **Fable** |
| Work a test or generator checks immediately: `make generate`, `make web`, `gofumpt`, regenerating the coverage floors, a Playwright smoke or handler test in an existing pattern; search and first reading of a source tree (go-fabric, the reference pages, the hosts, this repo) whose result a stronger model or the owner reads again | **Sonnet**, including as the model of a delegated search subagent |

Sonnet never writes a fact about go-fabric or the reference pages into the
repo, and never touches the handler's admin and token paths. When in doubt
about which row a task is in, it is in the first.

## Token use

As go-fabric's `CLAUDE.md` (owner, 2026-10-14: "Wichtig ist mir eine
effiziente Nutzung der Token"). A wave costs on the order of half a million
to a million tokens per implementing agent before the owner's review; the
rules below are what keeps that from doubling.

- **Read slices, not files.** `sed -n a,bp`, `grep -n` with a tight
  pattern and `head`; a whole file only when the task is the file. Never
  re-read what is already in context; cite the earlier read.
- **Delegate first reads.** A source tree's inventory (go-fabric's facade
  at its tag, the reference pages at their pins, a host, this repo) is a
  Sonnet read written to `notes/reads/read-*.md` with `file:line` for
  every fact and no judgement; the owner spot-checks the load-bearing
  lines and verifies every fact before it enters a brief or `docs/` —
  Sonnet's result is a map, never a citation.
- **Logs by their summary.** Guards (`make ci`, `make race`, `make e2e`,
  `make e2e-web`, `make contract-check`, `make generate-check`,
  `make cover-check`, `make version-check` — `gmake` where BSD make 3.81
  misreads `#` in `$(shell …)`) run in the background with a summary log
  (HEAD before and after, one `exit` line per guard); read the summary,
  and open a guard's log only on failure, with `grep` for `FAIL`,
  `DATA RACE`, `_test.go:`, the Playwright failure lines and the error
  strings — never `cat`. The same for CI: the job log is fetched once to
  a file and grepped.
- **One turn, every independent call.** Reads, checks and background
  starts that do not depend on each other go out together; a wait is a
  `Monitor` or an `until` loop in the background, not polling by hand.
- **Short artefacts.** A brief (`notes/briefs/<date>-<wave>-<topic>.md`)
  names decisions and sections, not prose; a report
  (`docs/reports/<wave>.md`) follows the fixed shape — per section with
  the commits, guards with HEAD, differences from the brief, decisions,
  left out, questions, the hosts' obligations; status messages to the
  owner are short, outcome first, in German.
- **Skip what cannot have changed.** No `make web` or `make e2e-web` when
  no file under `ui/` moved; no `make generate` when the contract did not
  move; no second guard run for a notes-only commit — say what was
  skipped and why.
- **Edit by script.** Several replacements in several files are one
  scripted call with asserted anchors, not a chain of interactive edits;
  a generated file (`internal/api`, the bundle `fabricui/dist`) is
  regenerated, never patched.
- **Agents read what the brief names.** An implementing agent gets the
  file list and the pinned sources in the brief, reads those, and reports
  in the fixed shape; it does not survey the tree. The brief is copied to
  `notes/briefs/` in the wave's first commit. One agent per brief; two
  agents only on disjoint paths.
- **Push, pull request and tag are the session's own calls.** An
  implementing agent works locally — branch, commits, guards — and
  reports; the owner's session reads the diff, pushes, opens the pull
  request with auto-merge and tags after CI. A delegated push, pull
  request or tag is refused by the auto-mode classifier ("Create Public
  Surface", 2026-11-07), and the review above sits between the agent's
  report and the push anyway.
