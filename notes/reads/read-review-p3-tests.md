# Review phase 3 — tests, guards and workflows map (script, read-only map)

Generated at `f4b8573` by the coordinator's script over a `git archive` export; per test file the counts of tests, skips, sleeps; a heuristic list of Go tests without a visible assertion (many use a fixture helper — check before reporting); web test counts; the coverage floors; every `uses:` and `permissions` line of the workflows. A map, never a citation: every row is a grep or script hit, not a reading.

## Go test files: tests, t.Skip, time.Sleep, t.Parallel
fabricui/devicetype_names_test.go tests=1 skip=0 sleep=0 parallel=0 lines=25
fabricui/e2e_test.go tests=4 skip=0 sleep=0 parallel=0 lines=517
fabricui/e2eweb_test.go tests=1 skip=1 sleep=0 parallel=0 lines=99
fabricui/error_branches_test.go tests=3 skip=0 sleep=0 parallel=0 lines=77
fabricui/gofabric/e2e_test.go tests=10 skip=0 sleep=3 parallel=0 lines=1223
fabricui/gofabric/gofabric_test.go tests=12 skip=0 sleep=0 parallel=0 lines=302
fabricui/gofabric/hooks_test.go tests=8 skip=0 sleep=4 parallel=0 lines=203
fabricui/gofabric/node_test.go tests=1 skip=0 sleep=0 parallel=0 lines=34
fabricui/gofabric/ping_test.go tests=4 skip=0 sleep=1 parallel=0 lines=97
fabricui/gofabric/topology_test.go tests=6 skip=0 sleep=0 parallel=0 lines=237
fabricui/handlers_controller_errors_test.go tests=3 skip=0 sleep=0 parallel=0 lines=250
fabricui/handlers_controller_review_test.go tests=8 skip=0 sleep=0 parallel=0 lines=348
fabricui/handlers_controller_test.go tests=21 skip=0 sleep=0 parallel=0 lines=749
fabricui/handlers_diagnostics_test.go tests=13 skip=0 sleep=0 parallel=0 lines=147
fabricui/handlers_endpoint_test.go tests=12 skip=0 sleep=0 parallel=0 lines=338
fabricui/handlers_exposure_test.go tests=30 skip=0 sleep=0 parallel=0 lines=381
fabricui/handlers_nodes_test.go tests=8 skip=0 sleep=0 parallel=0 lines=213
fabricui/handlers_node_test.go tests=30 skip=0 sleep=0 parallel=0 lines=585
fabricui/handlers_operation_test.go tests=7 skip=0 sleep=0 parallel=0 lines=180
fabricui/handlers_pairing_test.go tests=29 skip=0 sleep=0 parallel=0 lines=433
fabricui/handlers_surface_test.go tests=4 skip=0 sleep=0 parallel=0 lines=74
fabricui/handlers_topology_test.go tests=11 skip=0 sleep=0 parallel=0 lines=344
fabricui/headers_test.go tests=2 skip=0 sleep=0 parallel=0 lines=113
fabricui/memory_controller_test.go tests=3 skip=0 sleep=0 parallel=0 lines=107
fabricui/memory_provider_test.go tests=0 skip=0 sleep=0 parallel=0 lines=156
fabricui/misc_test.go tests=17 skip=0 sleep=1 parallel=0 lines=510
fabricui/nodeswitch_test.go tests=2 skip=0 sleep=0 parallel=0 lines=87
fabricui/registry_test.go tests=16 skip=0 sleep=0 parallel=0 lines=703
fabricui/spa_csp_test.go tests=8 skip=0 sleep=0 parallel=0 lines=239
fabricui/stream_test.go tests=6 skip=0 sleep=5 parallel=0 lines=222
fabricui/support_test.go tests=0 skip=0 sleep=0 parallel=0 lines=874
internal/arch/arch_test.go tests=3 skip=0 sleep=0 parallel=0 lines=214

## Go tests whose body has no t.Error/t.Fatal/require/assert/cmp (heuristic)
fabricui/handlers_topology_test.go:84 TestTopology_NoNode_Returns503(t
fabricui/handlers_topology_test.go:253 TestControllerTopology_Fixtures(t
fabricui/handlers_surface_test.go:18 TestSurface_ServesTheReport(t
fabricui/handlers_surface_test.go:23 TestSurface_NoNode_Returns503(t
fabricui/handlers_pairing_test.go:28 TestPairing_Configured_Returns200(t
fabricui/handlers_pairing_test.go:34 TestPairing_NoCredentials_Returns503(t
fabricui/handlers_pairing_test.go:74 TestPairing_IsAdminOnly(t
fabricui/handlers_pairing_test.go:98 TestOpenWindow_DurationTooShort_Returns400(t
fabricui/handlers_operation_test.go:32 TestOperation_Fixtures(t
fabricui/handlers_nodes_test.go:28 TestNodes_Fixtures(t
fabricui/handlers_node_test.go:29 TestStatus_NoNode_Returns200Disabled(t
fabricui/handlers_node_test.go:35 TestStatus_WithNode_Returns200WithSnapshot(t
fabricui/handlers_node_test.go:192 TestFabrics_NoNode_Returns503(t
fabricui/handlers_node_test.go:248 TestFabrics_MultipleRecords_Returns200(t
fabricui/handlers_node_test.go:285 TestRemoveFabric_ZeroIndex_Returns400(t
fabricui/handlers_exposure_test.go:25 TestExposureSchema_NoProvider_Returns503(t
fabricui/handlers_exposure_test.go:38 TestExposureSchema_Returns200(t
fabricui/handlers_exposure_test.go:73 TestExposureList_Returns200(t
fabricui/handlers_exposure_test.go:258 TestExposurePreview_Returns200(t
fabricui/handlers_exposure_test.go:264 TestExposurePreview_UnknownKey_Returns400(t
fabricui/handlers_diagnostics_test.go:20 TestSessions_ReportPerPeerIdleAge(t
fabricui/handlers_diagnostics_test.go:26 TestSessions_NoNode_Returns503(t
fabricui/handlers_diagnostics_test.go:56 TestDiagnosticEvents_ServeTheTraceNewestFirst(t
fabricui/handlers_diagnostics_test.go:69 TestEndpoints_ListBridgedDevicesAndParts(t
fabricui/handlers_diagnostics_test.go:74 TestMdns_ReportsTheAdvertisement(t
fabricui/handlers_diagnostics_test.go:79 TestCompatibility_JoinsFabricsAndEndpoints(t
fabricui/handlers_controller_test.go:120 TestControllerFixtures_BLE(t
fabricui/gofabric/hooks_test.go:69 TestHooks_ConcurrentCloseAndUpdates(t

## web tests: file, it/test count, skip/only, setTimeout
ui/e2e/origins.spec.ts cases=2 skip=0 timeout=1
ui/e2e/smoke.spec.ts cases=9 skip=0 timeout=0
ui/src/lib/api/client-errors.test.ts cases=8 skip=0 timeout=0
ui/src/lib/api/client.test.ts cases=5 skip=0 timeout=0
ui/src/lib/api/controller-client.test.ts cases=3 skip=0 timeout=0
ui/src/lib/api/fixtures.test.ts cases=1 skip=0 timeout=0
ui/src/lib/api/json.test.ts cases=8 skip=0 timeout=0
ui/src/lib/api/schema.test.ts cases=1 skip=0 timeout=0
ui/src/lib/api/stream-life.test.ts cases=7 skip=0 timeout=0
ui/src/lib/api/stream.test.ts cases=3 skip=0 timeout=0
ui/src/lib/context-ref.test.ts cases=1 skip=0 timeout=0
ui/src/lib/countdown.test.ts cases=3 skip=0 timeout=0
ui/src/lib/each-keys.test.ts cases=3 skip=0 timeout=0
ui/src/lib/i18n/catalogs.test.ts cases=7 skip=0 timeout=0
ui/src/lib/i18n/coverage.test.ts cases=8 skip=0 timeout=0
ui/src/lib/kit/ClusterView.test.ts cases=1 skip=0 timeout=0
ui/src/lib/kit/DataTable.test.ts cases=4 skip=0 timeout=0
ui/src/lib/kit/graph-layout.test.ts cases=6 skip=0 timeout=0
ui/src/lib/kit/Graph.test.ts cases=7 skip=0 timeout=0
ui/src/lib/kit/json-display.test.ts cases=2 skip=0 timeout=0
ui/src/lib/kit/kit.test.ts cases=23 skip=0 timeout=0
ui/src/lib/kit/Tabs.test.ts cases=2 skip=0 timeout=0
ui/src/lib/kit/TagChips.test.ts cases=1 skip=0 timeout=0
ui/src/lib/kit/utils.test.ts cases=2 skip=0 timeout=0
ui/src/lib/pages/controller/Commission.test.ts cases=17 skip=0 timeout=1
ui/src/lib/pages/controller/Diagnostics.test.ts cases=5 skip=0 timeout=0
ui/src/lib/pages/controller/Network.test.ts cases=10 skip=0 timeout=0
ui/src/lib/pages/controller/node/Attributes.test.ts cases=5 skip=0 timeout=0
ui/src/lib/pages/controller/node/Endpoints.test.ts cases=2 skip=0 timeout=0
ui/src/lib/pages/controller/node/Events.test.ts cases=2 skip=0 timeout=1
ui/src/lib/pages/controller/Nodes.test.ts cases=5 skip=0 timeout=0
ui/src/lib/pages/controller/Node.test.ts cases=29 skip=0 timeout=2
ui/src/lib/pages/controller/QrScan.test.ts cases=2 skip=0 timeout=0
ui/src/lib/pages/controller/Settings.test.ts cases=6 skip=0 timeout=0
ui/src/lib/pages/node/Diagnostics.test.ts cases=22 skip=0 timeout=1
ui/src/lib/pages/node/Endpoint.test.ts cases=11 skip=0 timeout=0
ui/src/lib/pages/node/Fabrics.test.ts cases=19 skip=0 timeout=1
ui/src/lib/pages/node/fields.test.ts cases=6 skip=0 timeout=0
ui/src/lib/pages/node/NodeManage.test.ts cases=7 skip=0 timeout=0
ui/src/lib/pages/node/Pair.test.ts cases=19 skip=0 timeout=0
ui/src/lib/pages/node/Selection.test.ts cases=21 skip=0 timeout=0
ui/src/lib/pages/node/Status.test.ts cases=14 skip=0 timeout=0
ui/src/lib/pages/node/Surface.test.ts cases=11 skip=0 timeout=1
ui/src/lib/pages/node/Topology.test.ts cases=7 skip=0 timeout=0
ui/src/lib/pages/operation/Operation.bound.test.ts cases=1 skip=0 timeout=0
ui/src/lib/pages/operation/Operation.test.ts cases=11 skip=0 timeout=0
ui/src/lib/scan/scan.test.ts cases=1 skip=0 timeout=0
ui/src/lib/scan/start.test.ts cases=3 skip=0 timeout=0
ui/src/lib/store-pairing.test.ts cases=7 skip=0 timeout=0
ui/src/lib/store-review.test.ts cases=15 skip=0 timeout=0
ui/src/lib/store.test.ts cases=18 skip=0 timeout=0
ui/src/lib/theme/theme.test.ts cases=23 skip=0 timeout=1
ui/src/lib/theme/tokens.test.ts cases=2 skip=0 timeout=0
ui/src/shell/App.test.ts cases=5 skip=0 timeout=2
ui/src/shell/dialogs.test.ts cases=5 skip=0 timeout=0
ui/src/shell/route.test.ts cases=5 skip=0 timeout=0
ui/src/shell/toast.test.ts cases=2 skip=0 timeout=0

## coverage floors
# Statement coverage floors per package (script/coverfloor). A ratchet:
# raise a floor with the change that raises the coverage.
fabricui            88
fabricui/gofabric   50

## workflows: uses and permissions
.github/workflows/ci.yml:20:permissions:
.github/workflows/ci.yml:45:      - uses: actions/checkout@v7.0.1
.github/workflows/ci.yml:46:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:70:      - uses: actions/checkout@v7.0.1
.github/workflows/ci.yml:73:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:105:      - uses: actions/checkout@v7.0.1
.github/workflows/ci.yml:106:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:119:      - uses: actions/checkout@v7.0.1
.github/workflows/ci.yml:120:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:133:      - uses: actions/checkout@v7
.github/workflows/ci.yml:134:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:145:      - uses: actions/checkout@v7
.github/workflows/ci.yml:146:      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7.1.0
.github/workflows/ci.yml:166:      - uses: actions/checkout@v7
.github/workflows/ci.yml:167:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:171:      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7.1.0
.github/workflows/ci.yml:180:        uses: actions/cache@55cc8345863c7cc4c66a329aec7e433d2d1c52a9 # v6.1.0
.github/workflows/ci.yml:193:      - uses: actions/checkout@v7
.github/workflows/ci.yml:194:      - uses: actions/setup-go@v7.0.0
.github/workflows/ci.yml:198:      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7.1.0
.github/workflows/nightly.yml:20:permissions:
.github/workflows/nightly.yml:32:      - uses: actions/checkout@v7.0.1
.github/workflows/nightly.yml:33:      - uses: actions/setup-go@v7.0.0
.github/workflows/codeql.yml:16:permissions:
.github/workflows/codeql.yml:32:      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
.github/workflows/codeql.yml:33:      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
.github/workflows/codeql.yml:38:        uses: github/codeql-action/init@33d7b85da1a04171e211befe2c739303b9a4c066 # v4.39.3
.github/workflows/codeql.yml:46:        uses: github/codeql-action/analyze@33d7b85da1a04171e211befe2c739303b9a4c066 # v4.39.3
.github/workflows/release-on-tag.yml:19:permissions:
.github/workflows/release-on-tag.yml:37:      - uses: actions/checkout@v7
.github/workflows/release-on-tag.yml:44:      - uses: actions/setup-go@v7.0.0
.github/workflows/release-on-tag.yml:48:      - uses: actions/setup-node@949feb2413d6458794dcd2491c4babbbce0c15c1 # v7.1.0
.github/workflows/release-on-tag.yml:66:        uses: softprops/action-gh-release@8a0cd651e5b2365bd640854a3fa5dfc70e4875f3  # v3
.github/workflows/dependabot-auto-merge.yml:4:on: pull_request_target
.github/workflows/dependabot-auto-merge.yml:6:permissions:
.github/workflows/dependabot-auto-merge.yml:17:        uses: dependabot/fetch-metadata@25dd0e34f4fe68f24cc83900b1fe3fe149efef98
