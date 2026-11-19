# Read: review phase 4 — the delta since each earlier read

Script output (coordinator, 2026-11-19), HEAD `979a67a` (origin/main after #78). A map, never a citation.

## Hand-written files changed since phase 2's read (`f4b8573..HEAD`), generated and dist excluded

| File | + | - |
| --- | --- | --- |
| `fabricui/fuzz_test.go` | 525 | 0 |
| `contract/openapi.yaml` | 357 | 34 |
| `fabricui/host_test.go` | 313 | 0 |
| `fabricui/gofabric/adapter_paths_e2e_test.go` | 252 | 0 |
| `fabricui/routes_contract_test.go` | 225 | 0 |
| `ui/src/lib/api/client.ts` | 224 | 65 |
| `fabricui/fuzz_checks_test.go` | 204 | 0 |
| `fabricui/handler.go` | 203 | 41 |
| `fabricui/strict.go` | 200 | 0 |
| `ui/src/lib/token-path.test.ts` | 194 | 0 |
| `fabricui/nonadmin_test.go` | 193 | 0 |
| `ui/src/lib/contract-0.10.test.ts` | 192 | 0 |
| `fabricui/handlers_controller.go` | 187 | 67 |
| `ui/src/lib/api/stream.ts` | 173 | 5 |
| `fabricui/status_guard_test.go` | 164 | 0 |
| `ui/scripts/contract-guard.mjs` | 161 | 7 |
| `fabricui/admin_request_test.go` | 153 | 0 |
| `ui/src/lib/api/stream-fetch.test.ts` | 146 | 0 |
| `fabricui/host.go` | 119 | 0 |
| `fabricui/handlers_node.go` | 110 | 33 |
| `fabricui/options_streams_test.go` | 102 | 0 |
| `fabricui/gofabric/fuzz_test.go` | 97 | 0 |
| `fabricui/options.go` | 87 | 28 |
| `ui/src/shell/App.test.ts` | 86 | 7 |
| `ui/scripts/contract-guard.test.mjs` | 86 | 3 |
| `ui/src/lib/i18n/coverage.test.ts` | 83 | 20 |
| `fabricui/gofabric/e2e_test.go` | 67 | 3 |
| `ui/e2e/smoke.spec.ts` | 65 | 41 |
| `ui/src/lib/api/client.test.ts` | 57 | 6 |
| `ui/src/lib/controller-store.svelte.ts` | 53 | 2 |
| `ui/src/lib/store.svelte.ts` | 51 | 16 |
| `fabricui/handlers_pairing_test.go` | 51 | 0 |
| `fabricui/gofabric/adapter_paths_test.go` | 49 | 0 |
| `fabricui/gofabric/gofabric.go` | 44 | 21 |
| `ui/src/lib/each-keys.test.ts` | 39 | 22 |
| `ui/src/lib/pages/operation/Operation.svelte` | 35 | 9 |
| `fabricui/support_test.go` | 34 | 7 |
| `ui/src/lib/kit/events.ts` | 32 | 5 |
| `ui/src/shell/App.svelte` | 32 | 10 |
| `ui/src/lib/pages/controller/node/Acl.svelte` | 27 | 9 |
| `fabricui/stream_test.go` | 25 | 7 |
| `fabricui/events.go` | 24 | 0 |
| `fabricui/problem.go` | 21 | 8 |
| `fabricui/handlers_node_test.go` | 21 | 8 |
| `ui/src/lib/api/client-errors.test.ts` | 20 | 6 |
| `ui/src/lib/pages/controller/Settings.svelte` | 20 | 2 |
| `fabricui/stream.go` | 20 | 1 |
| `fabricui/handlers_pairing.go` | 18 | 7 |
| `ui/src/lib/pages/controller/Commission.test.ts` | 18 | 5 |
| `ui/src/lib/kit/DeniedState.svelte` | 18 | 0 |
| `fabricui/controller.go` | 17 | 13 |
| `ui/src/lib/countdown.ts` | 16 | 0 |
| `ui/src/lib/api/controller-client.test.ts` | 15 | 3 |
| `ui/src/lib/pages/node/NodeManage.test.ts` | 14 | 2 |
| `ui/src/lib/pages/node/Selection.svelte` | 14 | 1 |
| `ui/src/lib/i18n/en.json` | 13 | 0 |
| `ui/src/lib/i18n/de.json` | 13 | 0 |
| `ui/src/lib/pages/node/Endpoint.svelte` | 12 | 5 |
| `ui/src/lib/pages/controller/Commission.svelte` | 12 | 3 |
| `script/coverfloor/floors.txt` | 11 | 3 |
| `fabricui/gofabric/node.go` | 10 | 6 |
| `ui/src/lib/pages/node/Diagnostics.test.ts` | 10 | 2 |
| `script/coverfloor/main.go` | 10 | 2 |
| `fabricui/misc_test.go` | 10 | 2 |
| `fabricui/handlers_topology_test.go` | 9 | 9 |
| `ui/src/lib/pages/node/Status.svelte` | 9 | 5 |
| `fabricui/node_detail.go` | 9 | 2 |
| `fabricui/handlers_exposure.go` | 9 | 2 |
| `ui/src/lib/pages/node/Pair.test.ts` | 8 | 7 |
| `ui/src/lib/store-review.test.ts` | 8 | 6 |
| `ui/src/shell/NodeFamily.svelte` | 8 | 2 |
| `ui/src/lib/pages/controller/node/Bindings.svelte` | 8 | 2 |
| `ui/src/test/support.ts` | 8 | 0 |
| `fabricui/memory_controller.go` | 8 | 0 |
| `ui/src/lib/pages/node/Fabrics.svelte` | 7 | 1 |
| `fabricui/handlers_controller_review_test.go` | 6 | 3 |
| `ui/src/lib/pages/controller/node/Overview.svelte` | 6 | 2 |
| `fabricui/registry_test.go` | 6 | 2 |
| `ui/src/lib/pages/node/Topology.svelte` | 6 | 0 |
| `ui/src/lib/pages/node/Surface.svelte` | 6 | 0 |
| `ui/src/lib/pages/node/Endpoints.svelte` | 6 | 0 |
| `ui/src/lib/pages/node/Diagnostics.svelte` | 6 | 0 |
| `ui/src/lib/pages/controller/Nodes.svelte` | 6 | 0 |
| `ui/src/lib/pages/controller/Node.svelte` | 6 | 0 |
| `ui/src/lib/pages/controller/Diagnostics.svelte` | 6 | 0 |
| `fabricui/handlers_endpoint_test.go` | 5 | 6 |
| `ui/src/lib/pages/controller/node/Events.svelte` | 5 | 4 |
| `fabricui/handlers_diagnostics.go` | 4 | 4 |
| `ui/src/lib/store.test.ts` | 4 | 3 |
| `ui/src/lib/pages/node/Status.test.ts` | 4 | 3 |
| `ui/src/lib/pages/node/Endpoint.test.ts` | 4 | 3 |
| `script/gen-api.sh` | 4 | 3 |
| `ui/src/shell/toast.svelte.ts` | 4 | 2 |
| `fabricui/handlers_exposure_test.go` | 4 | 2 |
| `ui/src/lib/context.ts` | 4 | 1 |
| `fabricui/e2e_test.go` | 4 | 0 |
| `ui/src/lib/pages/node/Selection.test.ts` | 3 | 2 |
| `ui/src/lib/pages/node/Fabrics.test.ts` | 3 | 2 |
| `ui/src/lib/kit/EventList.svelte` | 3 | 2 |
| `ui/src/lib/i18n/catalogs.test.ts` | 3 | 2 |
| `fabricui/handlers_controller_test.go` | 3 | 2 |
| `ui/src/lib/pages/node/EcosystemHints.svelte` | 3 | 1 |
| `fabricui/registry.go` | 3 | 1 |
| `fabricui/gofabric/topology_test.go` | 3 | 1 |
| `ui/src/lib/store-pairing.test.ts` | 2 | 2 |
| `ui/src/lib/pages/node/Pair.svelte` | 2 | 2 |
| `ui/src/lib/pages/controller/Node.test.ts` | 2 | 2 |
| `fabricui/handlers_nodes_test.go` | 2 | 2 |
| `fabricui/handlers_diagnostics_test.go` | 2 | 2 |
| `ui/src/shell/route.ts` | 2 | 1 |
| `ui/src/lib/kit/Spinner.svelte` | 1 | 1 |
| `internal/version/version.go` | 1 | 1 |
| `fabricui/handlers_surface.go` | 1 | 1 |
| `fabricui/gofabric/hooks_test.go` | 1 | 1 |
| `ui/src/lib/kit/index.ts` | 1 | 0 |
| `fabricui/handlers_operation_test.go` | 1 | 0 |
| `fabricui/error_branches_test.go` | 0 | 16 |

## Production files changed since the baseline (`25f6e7b..HEAD`) but not since `f4b8573`

- `fabricui/gofabric/diagnostics.go`
- `fabricui/gofabric/hooks.go`
- `fabricui/gofabric/topology.go`
- `fabricui/gofabric/updates.go`
- `fabricui/nodeswitch.go`
- `fabricui/operations.go`
- `fabricui/spa.go`
- `ui/src/lib/api/index.ts`
- `ui/src/lib/api/json.ts`
- `ui/src/lib/kit/AttributeTable.svelte`
- `ui/src/lib/kit/Badge.svelte`
- `ui/src/lib/kit/Button.svelte`
- `ui/src/lib/kit/ClusterView.svelte`
- `ui/src/lib/kit/DataTable.svelte`
- `ui/src/lib/kit/ErrorState.svelte`
- `ui/src/lib/kit/Graph.svelte`
- `ui/src/lib/kit/Select.svelte`
- `ui/src/lib/kit/Tabs.svelte`
- `ui/src/lib/kit/TagChips.svelte`
- `ui/src/lib/kit/data-table.ts`
- `ui/src/lib/kit/dates.ts`
- `ui/src/lib/kit/graph-layout.ts`
- `ui/src/lib/kit/utils.ts`
- `ui/src/lib/pages/controller/Network.svelte`
- `ui/src/lib/pages/controller/QrScan.svelte`
- `ui/src/lib/pages/controller/format.ts`
- `ui/src/lib/pages/controller/node/Attributes.svelte`
- `ui/src/lib/pages/controller/node/Endpoints.svelte`
- `ui/src/lib/pages/controller/node/Fabrics.svelte`
- `ui/src/lib/pages/controller/node/Icd.svelte`
- `ui/src/lib/pages/controller/node/Inventory.svelte`
- `ui/src/lib/pages/controller/node/Updates.svelte`
- `ui/src/lib/pages/controller/node/Window.svelte`
- `ui/src/lib/pages/node/HostWarnings.svelte`
- `ui/src/lib/pages/node/NodeConfig.svelte`
- `ui/src/lib/pages/node/NodeManage.svelte`
- `ui/src/lib/pages/operation/log.ts`
- `ui/src/lib/scan/index.ts`
- `ui/src/lib/theme/ha-bridge.ts`
- `ui/src/lib/theme/tokens.css`
- `ui/src/shell/ConfirmDialog.svelte`
- `ui/src/shell/NodeCreateDialog.svelte`
- `ui/src/shell/confirm.svelte.ts`
- `ui/src/shell/focus-trap.ts`
- `ui/src/shell/index.html`

## go-fabric v0.22.2 → v0.23.0: facade packages the adapter imports

- `client/discovery`: +/- 704 / 116 (non-test lines)
- `clock`: +/- 5 / 5 (non-test lines)
- `cluster/spec`: +/- 22411 / 1893 (non-test lines)
- `cluster/spec/accesscontrol`: +/- 155 / 0 (non-test lines)
- `cluster/spec/basicinformation`: +/- 126 / 10 (non-test lines)
- `cluster/spec/binding`: +/- 51 / 5 (non-test lines)
- `cluster/spec/icdmanagement`: +/- 116 / 13 (non-test lines)
- `cluster/spec/operationalcredentials`: +/- 383 / 4 (non-test lines)
- `contract`: +/- 780 / 141 (non-test lines)
- `controller`: +/- 10121 / 1693 (non-test lines)
- `controller/attestation`: +/- 181 / 167 (non-test lines)
- `controller/ble/bletest`: +/- 75 / 15 (non-test lines)
- `controller/client`: +/- 3087 / 370 (non-test lines)
- `controller/commissioner`: +/- 1959 / 399 (non-test lines)
- `controller/softwareupdate`: +/- 799 / 106 (non-test lines)
- `controller/topology`: +/- 214 / 24 (non-test lines)
- `device`: +/- 1604 / 614 (non-test lines)
- `diag`: +/- 0 / 0 (non-test lines)
- `eligibility`: +/- 3 / 3 (non-test lines)
- `fabrictest`: +/- 252 / 55 (non-test lines)
- `node`: +/- 5257 / 940 (non-test lines)
- `schema`: +/- 5034 / 868 (non-test lines)
- `surface`: +/- 94 / 24 (non-test lines)
- `thread`: +/- 169 / 33 (non-test lines)

## Render sinks and raw HTML (`{@html`, innerHTML) in ui/src

- `ui/src/lib/pages/node/Pair.svelte:174`
- `ui/src/lib/pages/controller/node/Window.svelte:102`
