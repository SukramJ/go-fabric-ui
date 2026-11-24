# Brief M1: go-fabric-ui moves to thingfabric.dev (v0.14.0, contract 0.14.0)

Owner, 2026-11-24: the repository was transferred to
`github.com/thingfabric/go-fabric-ui` (old URLs redirect); the module path
becomes `thingfabric.dev/go-fabric-ui` (go-import served by
`thingfabric/thingfabric.github.io`, like `thingfabric.dev/go-fabric`).
Decisions (owner wizard 2026-11-24): npm scope `@thingfabric/fabric-ui`;
problem `type` URIs on the own domain; only this repository, go-fabric-server
and go-ha2fabric move (go-hamapping, go-hamqtt, go-ha-catalog stay). Model:
Opus for every role (owner rule 2026-11-24, Opus instead of Fable).

Branch `chore/thingfabric-move` from `origin/main` (`109cf14`). First commit:
this brief.

## Work

| # | What |
| --- | --- |
| 1 | `go.mod` `module thingfabric.dev/go-fabric-ui`; every import; `internal/arch` and any test that names the path; `make generate` if the generator embeds it; `gofumpt` |
| 2 | Problem `type` URIs: `problemTypeBase` (`fabricui/problem.go:48`) → `https://thingfabric.dev/go-fabric-ui/problems#problem-`, the anchor names unchanged; every fixture in `contract/fixtures`; `contract/openapi.yaml` where it states the base; `docs/contract.md`. Contract **0.14.0**, breaking for a client that compares `type` as a string — say so in `changelog.md` and `docs/contract.md`; `contract/previous` and `make contract-check` as the project does a version step |
| 3 | npm: `ui/package.json` `name` `@thingfabric/fabric-ui`, `package-lock.json` (regenerate with `npm install --package-lock-only --ignore-scripts`), the release workflow's step names and texts, README / docs install lines; `make web` (the bundle names the package?) and `make dist-check` |
| 4 | Links: `github.com/SukramJ/go-fabric-ui` → `github.com/thingfabric/go-fabric-ui`; `github.com/SukramJ/go-fabric` (stale) → `github.com/thingfabric/go-fabric` (a URL) or `thingfabric.dev/go-fabric` (an import path); `github.com/SukramJ/go-fabric-server` and `go-ha2fabric` → `github.com/thingfabric/…`; `openccu-loom`, `go-hamapping`, `go-hamqtt`, `go-ha-catalog` stay under `SukramJ`. Leave `notes/` history as it is (only this brief and the report are new) |
| 5 | Version `v0.14.0`: `internal/version`, `ui/package.json` `version`, `changelog.md` section `[0.14.0] - 2026-11-24` (Changed: module path, npm scope, problem URIs — breaking — with a "What a consumer must do" list: `go get thingfabric.dev/go-fabric-ui@v0.14.0` and replace imports; `npm i @thingfabric/fabric-ui`; problem `type` base), `CLAUDE.md` / `README.md` state lines; `make version-check` |
| 6 | For the coordinator (not in this repository): a list of every problem type — anchor name, title, HTTP status, the `docs/contract.md` heading — written to `notes/briefs/2026-11-24-M1-problem-types.md`; it becomes the page `thingfabric.dev/go-fabric-ui/problems` |

## Rules

No sub-agents; no push, pull request, tag or npm publish (the coordinator's;
the npm org and trusted publisher are the owner's); never `git stash`;
`PATH=$HOME/go/bin:$PATH`. Guards of `CLAUDE.md` before the report: `make ci`,
`make race`, `make e2e`, `make e2e-web`, `make contract-check`,
`make notices-check`, `make version-check`, `make generate-check`,
`make cover-check`, `make fuzz-smoke`, `make vuln licenses secrets`, the
386 build/test if the Makefile has it — output to log files under
`/tmp/claude-1000/-home-markus-Dokumente-GitHub/8fe7f689-a76f-424d-88a1-c6d98442f53c/scratchpad/M1/`
(slow machine; read summaries). A `grep -rn -i sukramj` over the tree outside
`notes/` lists only what row 4 keeps. Report
`notes/briefs/2026-11-24-M1-report.md` (commits, guards with HEAD, per row,
what stays under SukramJ and why, questions), committed last. Stop when it is
written; do not leave background work running.
