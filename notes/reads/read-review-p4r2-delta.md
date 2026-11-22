# Read: review phase 4 run 2 — the delta since run 1's read

Script output (coordinator, 2026-11-19), HEAD `792d15b` (origin/main after #82, v0.11.0). A map, never a citation.

## Hand-written files changed since run 1's read (`979a67a..HEAD`), generated, dist and lock files excluded

| File | + | - |
| --- | --- | --- |
| `contract/previous/openapi.yaml` | 357 | 34 |
| `internal/statusguard/statusguard.go` | 184 | 0 |
| `fabricui/sibling_test.go` | 164 | 0 |
| `fabricui/gofabric/updates_test.go` | 142 | 0 |
| `fabricui/gofabric/updates.go` | 111 | 5 |
| `fabricui/host_proxy_test.go` | 94 | 0 |
| `changelog.md` | 88 | 0 |
| `ui/src/lib/token-path.test.ts` | 85 | 0 |
| `fabricui/log_once_test.go` | 84 | 0 |
| `fabricui/updates_check_test.go` | 84 | 0 |
| `contract/openapi.yaml` | 81 | 42 |
| `fabricui/handlers_controller.go` | 77 | 25 |
| `docs/contract.md` | 74 | 11 |
| `fabricui/gofabric/window_kind_e2e_test.go` | 63 | 0 |
| `ui/src/lib/pages/node/Pair.test.ts` | 63 | 2 |
| `fabricui/gofabric/hooks_test.go` | 61 | 0 |
| `fabricui/commission_bound_test.go` | 57 | 0 |
| `fabricui/gofabric/failsafe_e2e_test.go` | 55 | 0 |
| `ui/src/lib/pages/controller/Node.test.ts` | 48 | 0 |
| `fabricui/window_close_e2e_test.go` | 42 | 0 |
| `fabricui/int64_test.go` | 41 | 0 |
| `ui/src/lib/i18n/coverage.test.ts` | 41 | 8 |
| `ui/src/lib/pages/node/Fabrics.test.ts` | 39 | 0 |
| `ui/scripts/contract-guard.mjs` | 39 | 13 |
| `fabricui/handlers_node.go` | 38 | 8 |
| `ui/scripts/contract-guard.test.mjs` | 37 | 0 |
| `fabricui/fabric_not_found_test.go` | 34 | 0 |
| `ui/src/lib/pages/node/Pair.svelte` | 34 | 2 |
| `ui/src/lib/store-scoped-stream.test.ts` | 33 | 0 |
| `ui/src/lib/store.svelte.ts` | 28 | 0 |
| `fabricui/controller.go` | 25 | 1 |
| `docs/embedding.md` | 25 | 9 |
| `fabricui/registry.go` | 24 | 1 |
| `fabricui/options.go` | 20 | 6 |
| `fabricui/gofabric/main_test.go` | 19 | 0 |
| `ui/src/lib/store-pairing.test.ts` | 17 | 0 |
| `ui/src/lib/pages/controller/Settings.test.ts` | 15 | 0 |
| `ui/src/lib/pages/index.test.ts` | 15 | 0 |
| `ui/src/lib/pages/node/Status.test.ts` | 15 | 0 |
| `ui/src/lib/api/client.ts` | 15 | 4 |
| `fabricui/handler.go` | 14 | 7 |
| `internal/statusobs/statusobs.go` | 13 | 0 |
| `fabricui/status_guard_test.go` | 13 | 147 |
| `ui/src/lib/pages/controller/node/Updates.svelte` | 11 | 2 |
| `ui/src/lib/pages/node/HostWarnings.svelte` | 7 | 1 |
| `fabricui/gofabric/gofabric.go` | 7 | 3 |
| `ui/src/lib/pages/controller/node/Acl.svelte` | 6 | 4 |
| `fabricui/gofabric/adapter_paths_test.go` | 5 | 1 |
| `Makefile` | 5 | 2 |
| `ui/README.md` | 5 | 2 |
| `script/coverfloor/floors.txt` | 4 | 0 |
| `ui/src/lib/pages/index.ts` | 4 | 0 |
| `fabricui/host.go` | 4 | 1 |
| `ui/src/lib/i18n/de.json` | 4 | 1 |
| `ui/src/lib/i18n/en.json` | 4 | 1 |
| `fabricui/handlers_pairing.go` | 4 | 2 |
| `fabricui/gofabric/hooks.go` | 3 | 0 |
| `.github/workflows/ci.yml` | 2 | 0 |
| `CLAUDE.md` | 2 | 2 |
| `fabricui/gofabric/e2e_test.go` | 2 | 2 |
| `fabricui/nodeswitch.go` | 2 | 2 |
| `fabricui/handlers_diagnostics.go` | 1 | 1 |
| `fabricui/misc_test.go` | 1 | 1 |
| `internal/version/version.go` | 1 | 1 |
| `ui/package.json` | 1 | 1 |
| `ui/src/lib/pages/controller/Settings.svelte` | 1 | 1 |
| `ui/src/lib/pages/node/Fabrics.svelte` | 1 | 1 |

## By unit

Files are cut by the brief's unit table; this list only says which changed.

## go-fabric

Pin unchanged at `v0.23.0` (`go.mod:6`); origin/main of go-fabric is 8 commits past it, no newer tag (go-fabric-ui#81 waits).
