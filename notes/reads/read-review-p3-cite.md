# Review phase 3 — citations map (script, read-only map)

Generated at `f4b8573` by the coordinator's script over a `git archive` export; every `path:line[-line]` in hand-written code, tests and docs (not `notes/`, generated files, fixtures, `contract/previous`), resolved by path suffix against this repository at `f4b8573` and the pins go-fabric `v0.22.2` (`gf`), openccu-loom `b40b289a` (`loom`), matter.js `862ad75f` (`mjs`, `docs/` only). `tag` = the pin the words before the citation name; `status` = whether the file resolves uniquely and the range is within it. `ok` says only that the lines exist, never that they say what the sentence claims. A citation that names another go-fabric tag is checked here against `v0.22.2` only. A map, never a citation: every row is a grep or script hit, not a reading.

| src | cited | tag | resolved | status | kind |
| --- | --- | --- | --- | --- | --- |
| `ui/src/shell/toast.svelte.ts:2` | `assets/ui/src/lib/stores/toast.svelte.ts:16-59` | loom | `loom:assets/ui/src/lib/stores/toast.svelte.ts` | ok | code |
| `ui/src/lib/store.test.ts:3` | `assets/ui/src/lib/stores/matter.svelte.ts:92-160` | loom | `loom:assets/ui/src/lib/stores/matter.svelte.ts` | ok | test |
| `ui/src/lib/store.svelte.ts:3` | `assets/ui/src/lib/stores/matter.svelte.ts:23-381` | loom | `loom:assets/ui/src/lib/stores/matter.svelte.ts` | ok | code |
| `ui/src/lib/store.svelte.ts:328` | `matter.svelte.ts:282-318` | loom | `loom:assets/ui/src/lib/stores/matter.svelte.ts` | ok | code |
| `ui/src/lib/qr.ts:3` | `assets/ui/src/lib/qr.ts:1-95` | loom | `loom:assets/ui/src/lib/qr.ts` | ok | code |
| `ui/src/lib/theme/ha-bridge.ts:6` | `assets/ui/src/lib/theme/ha-bridge.ts:1-184` | loom | `loom:assets/ui/src/lib/theme/ha-bridge.ts` | ok | code |
| `ui/src/lib/pages/node/Status.svelte:3` | `assets/ui/src/routes/matter/Matter.svelte:47-100` | loom | `loom:assets/ui/src/routes/matter/Matter.svelte` | ok | code |
| `ui/src/lib/pages/node/Selection.svelte:5` | `assets/ui/src/routes/matter/MatterExposureList.svelte:1-584` | ui | `loom:assets/ui/src/routes/matter/MatterExposureList.svelte` | ok [tagged ui] | code |
| `ui/src/lib/pages/node/Pair.svelte:3` | `assets/ui/src/routes/matter/MatterPair.svelte:1-279` | loom | `loom:assets/ui/src/routes/matter/MatterPair.svelte` | ok | code |
| `ui/src/lib/pages/node/NodeManage.svelte:9` | `docs-site/docs/getting-started/bridge-configuration.md:28-33` | ui | `` | unresolved | code |
| `ui/src/lib/pages/node/Fabrics.svelte:3` | `assets/ui/src/routes/matter/MatterFabrics.svelte:1-209` | loom | `loom:assets/ui/src/routes/matter/MatterFabrics.svelte` | ok | code |
| `ui/src/lib/pages/node/Endpoints.svelte:3` | `assets/ui/src/routes/matter/MatterDiagnostics.svelte:229-260` | ui | `loom:assets/ui/src/routes/matter/MatterDiagnostics.svelte` | ok [tagged ui] | code |
| `ui/src/lib/pages/node/Endpoints.svelte:62` | `node/surface_stability.go:62-63` | ui | `gf:node/surface_stability.go` | ok [tagged ui] | code |
| `ui/src/lib/pages/node/Endpoints.svelte:63` | `persist.go:90-94` | ui | `gf:node/persist.go,examples/reference-bridge/persist.go,controller/ca/persist.go` | ambiguous(4) | code |
| `ui/src/lib/pages/node/Diagnostics.svelte:4` | `assets/ui/src/routes/matter/MatterDiagnostics.svelte:1-352` | loom | `loom:assets/ui/src/routes/matter/MatterDiagnostics.svelte` | ok | code |
| `ui/src/lib/pages/node/Diagnostics.svelte:122` | `MatterDiagnostics.svelte:60-70` | loom | `loom:assets/ui/src/routes/matter/MatterDiagnostics.svelte` | ok | code |
| `ui/src/lib/kit/utils.ts:2` | `assets/ui/src/lib/utils.ts:1-65` | loom | `loom:assets/ui/src/lib/utils.ts` | ok | code |
| `ui/src/lib/kit/tabs.ts:2` | `assets/ui/src/lib/components/ui/tabs.ts:1-19` | loom | `loom:assets/ui/src/lib/components/ui/tabs.ts` | ok | code |
| `ui/src/lib/kit/im-status.ts:2` | `im/status.go:64-101` | gf | `gf:im/status.go` | ok | code |
| `ui/src/lib/kit/icons.ts:3` | `assets/ui/src/lib/icons.ts:1-20` | loom | `loom:assets/ui/src/lib/icons.ts` | ok | code |
| `ui/src/lib/kit/data-table.ts:2` | `assets/ui/src/lib/components/ui/data-table.ts:1-55` | loom | `loom:assets/ui/src/lib/components/ui/data-table.ts` | ok | code |
| `ui/src/lib/kit/Tabs.svelte:1` | `assets/ui/src/lib/components/ui//Tabs.svelte:1-107` | loom | `loom:assets/ui/src/lib/components/ui/Tabs.svelte` | ok | code |
| `ui/src/lib/kit/Spinner.svelte:1` | `assets/ui/src/lib/components/ui//Spinner.svelte:1-100` | loom | `loom:assets/ui/src/lib/components/ui/Spinner.svelte` | ok | code |
| `ui/src/lib/kit/Select.svelte:1` | `assets/ui/src/lib/components/ui//Select.svelte:1-80` | loom | `loom:assets/ui/src/lib/components/ui/Select.svelte` | ok | code |
| `ui/src/lib/kit/PageShell.svelte:1` | `assets/ui/src/lib/components/ui//PageShell.svelte:1-44` | loom | `loom:assets/ui/src/lib/components/ui/PageShell.svelte` | ok | code |
| `ui/src/lib/kit/PageHeader.svelte:1` | `assets/ui/src/lib/components/ui//PageHeader.svelte:1-72` | loom | `loom:assets/ui/src/lib/components/ui/PageHeader.svelte` | ok | code |
| `ui/src/lib/kit/LoadingState.svelte:1` | `assets/ui/src/lib/components/ui//LoadingState.svelte:1-30` | loom | `loom:assets/ui/src/lib/components/ui/LoadingState.svelte` | ok | code |
| `ui/src/lib/kit/Input.svelte:1` | `assets/ui/src/lib/components/ui//Input.svelte:1-24` | loom | `loom:assets/ui/src/lib/components/ui/Input.svelte` | ok | code |
| `ui/src/lib/kit/Icon.svelte:1` | `assets/ui/src/lib/components/ui//Icon.svelte:1-35` | loom | `loom:assets/ui/src/lib/components/ui/Icon.svelte` | ok | code |
| `ui/src/lib/kit/ErrorState.svelte:1` | `assets/ui/src/lib/components/ui//ErrorState.svelte:1-36` | loom | `loom:assets/ui/src/lib/components/ui/ErrorState.svelte` | ok | code |
| `ui/src/lib/kit/EmptyState.svelte:1` | `assets/ui/src/lib/components/ui//EmptyState.svelte:1-46` | loom | `loom:assets/ui/src/lib/components/ui/EmptyState.svelte` | ok | code |
| `ui/src/lib/kit/DataTable.svelte:1` | `assets/ui/src/lib/components/ui/DataTable.svelte:1-469` | loom | `loom:assets/ui/src/lib/components/ui/DataTable.svelte` | ok | code |
| `ui/src/lib/kit/Card.svelte:1` | `assets/ui/src/lib/components/ui//Card.svelte:1-18` | loom | `loom:assets/ui/src/lib/components/ui/Card.svelte` | ok | code |
| `ui/src/lib/kit/Button.svelte:1` | `assets/ui/src/lib/components/ui//Button.svelte:1-72` | loom | `loom:assets/ui/src/lib/components/ui/Button.svelte` | ok | code |
| `ui/src/lib/kit/Badge.svelte:1` | `assets/ui/src/lib/components/ui//Badge.svelte:1-43` | loom | `loom:assets/ui/src/lib/components/ui/Badge.svelte` | ok | code |
| `ui/src/lib/i18n/index.ts:4` | `assets/ui/src/lib/i18n.ts:36` | loom | `loom:assets/ui/src/lib/i18n.ts` | ok | code |
| `ui/src/lib/api/client.ts:5` | `assets/ui/src/lib/api/client.ts:1806-1875` | loom | `loom:assets/ui/src/lib/api/client.ts` | ok | code |
| `internal/arch/arch_test.go:38` | `script/archcheck/layers.json:7` | ui | `gf:script/archcheck/layers.json` | ok [tagged ui] | test |
| `internal/arch/arch_test.go:77` | `fabrictest/ble.go:18-28` | ui | `gf:fabrictest/ble.go` | ok [tagged ui] | test |
| `fabricui/support_test.go:498` | `im/eventlog.go:443` | gf | `gf:im/eventlog.go` | ok | test |
| `fabricui/support_test.go:599` | `surface/element.go:135-140` | gf | `gf:surface/element.go` | ok | test |
| `fabricui/support_test.go:608` | `surface/surface.go:24` | gf | `gf:surface/surface.go` | ok | test |
| `fabricui/stream.go:26` | `internal/north/rest/handlers/matter_events.go:62-82` | loom | `loom:internal/north/rest/handlers/matter_events.go` | ok | code |
| `fabricui/spa.go:30` | `internal/north/ui/spa.go:29-111` | loom | `loom:internal/north/ui/spa.go` | ok | code |
| `fabricui/spa.go:73` | `spa.go:73-85` | loom | `loom:internal/north/ui/spa.go` | ok | code |
| `fabricui/problem.go:37` | `internal/north/rest/problem/problem.go:50-59` | loom | `loom:internal/north/rest/problem/problem.go` | ok | code |
| `fabricui/problem.go:49` | `problem.go:61-85` | loom | `loom:internal/north/rest/problem/problem.go` | ok | code |
| `fabricui/problem.go:112` | `internal/north/rest/handlers/write.go:59-81` | ui | `loom:internal/north/rest/handlers/write.go` | ok [tagged ui] | code |
| `fabricui/problem.go:134` | `write.go:96-100` | ui | `gf:im/write.go,cluster/spec/write.go,client/interaction/write.go` | ambiguous(3) | code |
| `fabricui/problem.go:142` | `write.go:109-114` | loom | `loom:internal/north/rest/handlers/write.go,internal/model/group/write.go` | ambiguous(2) | code |
| `fabricui/problem.go:151` | `write.go:44-49` | loom | `loom:internal/north/rest/handlers/write.go,internal/model/group/write.go` | ambiguous(2) | code |
| `fabricui/ports.go:116` | `cmd/openccu-loom/daemon_matter.go:1255-1305` | loom | `loom:cmd/openccu-loom/daemon_matter.go` | ok | code |
| `fabricui/options.go:50` | `node/surface.go:372` | gf | `gf:node/surface.go` | ok | code |
| `fabricui/options.go:54` | `node/served.go:53-94` | gf | `gf:node/served.go` | ok | code |
| `fabricui/options.go:61` | `node/endpoints.go:14-66` | gf | `gf:node/endpoints.go` | ok | code |
| `fabricui/options.go:68` | `node/health.go:93` | ui | `gf:node/health.go` | ok [tagged ui] | code |
| `fabricui/options.go:68` | `node/diagnostics.go:40` | ui | `gf:node/diagnostics.go` | ok [tagged ui] | code |
| `fabricui/options.go:77` | `node/attestation.go:55-58` | gf | `gf:node/attestation.go` | ok | code |
| `fabricui/options.go:81` | `node/node.go:1138-1141` | ui | `gf:node/node.go` | ok [tagged ui] | code |
| `fabricui/options.go:84` | `node/sessions.go:110-118` | ui | `gf:node/sessions.go` | out-of-range(len 113) [tagged ui] | code |
| `fabricui/options.go:88` | `node/reannounce.go:12-26` | ui | `gf:node/reannounce.go` | ok [tagged ui] | code |
| `fabricui/options.go:235` | `node/factory_reset.go:20-45` | ui | `gf:node/factory_reset.go` | ok [tagged ui] | code |
| `fabricui/options.go:238` | `fabrictest/device.go:396-416` | ui | `gf:fabrictest/device.go` | ok [tagged ui] | code |
| `fabricui/nodeswitch.go:23` | `node/health.go:74-84` | gf | `gf:node/health.go` | ok | code |
| `fabricui/nodeswitch.go:24` | `node.go:1119-1121` | gf | `gf:node/node.go,fabrictest/node.go,controller/client/node.go` | ambiguous(3) | code |
| `fabricui/nodeswitch.go:25` | `factory_reset.go:66-67` | ui | `gf:node/factory_reset.go` | ok [tagged ui] | code |
| `fabricui/nodeswitch.go:29` | `events.go:152-166` | ui | `ui:fabricui/events.go` | ok | code |
| `fabricui/nodeswitch.go:35` | `registry.go:168-190` | ui | `ui:fabricui/registry.go` | ok | code |
| `fabricui/node_detail.go:42` | `node/served.go:70-94` | gf | `gf:node/served.go` | ok | code |
| `fabricui/node_detail.go:49` | `node/served.go:20-40` | ui | `gf:node/served.go` | ok [tagged ui] | code |
| `fabricui/node_detail.go:52` | `node/endpoints.go:14-66` | ui | `gf:node/endpoints.go` | ok [tagged ui] | code |
| `fabricui/node_detail.go:111` | `node/endpoints.go:19-24` | ui | `gf:node/endpoints.go` | ok [tagged ui] | code |
| `fabricui/node_detail.go:171` | `node/served.go:20-40` | gf | `gf:node/served.go` | ok | code |
| `fabricui/node_detail.go:311` | `node/served.go:140-144` | gf | `gf:node/served.go` | ok | code |
| `fabricui/memory_controller.go:155` | `controller/controller.go:623-631` | mjs | `gf:controller/controller.go` | ok [tagged mjs] | code |
| `fabricui/memory_controller.go:284` | `controller/commissioner/progress.go:78-80` | ui | `gf:controller/commissioner/progress.go` | ok [tagged ui] | code |
| `fabricui/handlers_surface_test.go:47` | `node/observers.go:75-84` | gf | `gf:node/observers.go` | ok | test |
| `fabricui/handlers_surface.go:13` | `surface/policy.go:400-432` | ui | `gf:surface/policy.go` | ok [tagged ui] | code |
| `fabricui/handlers_surface.go:24` | `surface/element.go:135-140` | ui | `gf:surface/element.go` | ok [tagged ui] | code |
| `fabricui/handlers_pairing.go:19` | `internal/north/rest/handlers/matter.go:169-180` | loom | `loom:internal/north/rest/handlers/matter.go` | ok | code |
| `fabricui/handlers_pairing.go:41` | `matter.go:51-85` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_pairing.go:60` | `matter.go:51-85` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_pairing.go:149` | `matter.go:158-226` | ui | `gf:contract/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_pairing.go:150` | `cmd/openccu-loom/matter_window_adapter.go:95-200` | loom | `loom:cmd/openccu-loom/matter_window_adapter.go` | ok | code |
| `fabricui/handlers_pairing.go:177` | `matter.go:131-139` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_pairing.go:246` | `matter_window_adapter.go:143-200` | loom | `loom:cmd/openccu-loom/matter_window_adapter.go` | ok | code |
| `fabricui/handlers_pairing.go:273` | `matter_exposures.go:422-444` | loom | `loom:internal/north/rest/handlers/matter_exposures.go` | ok | code |
| `fabricui/handlers_node.go:114` | `surface/profile.go:259-273` | gf | `gf:surface/profile.go` | ok | code |
| `fabricui/handlers_node.go:116` | `surface/policy.go:25-30` | ui | `gf:surface/policy.go` | ok [tagged ui] | code |
| `fabricui/handlers_node.go:137` | `matter_exposures.go:363-372` | loom | `loom:internal/north/rest/handlers/matter_exposures.go` | ok | code |
| `fabricui/handlers_node.go:221` | `matter.go:232-255` | ui | `gf:contract/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_node.go:241` | `matter.go:266-296` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_node.go:273` | `matter_exposures.go:383-410` | loom | `loom:internal/north/rest/handlers/matter_exposures.go` | ok | code |
| `fabricui/handlers_node.go:296` | `matter_maintenance.go:27-48` | loom | `loom:internal/north/rest/handlers/matter_maintenance.go` | ok | code |
| `fabricui/handlers_node.go:320` | `matter_maintenance.go:58-61` | loom | `loom:internal/north/rest/handlers/matter_maintenance.go` | ok | code |
| `fabricui/handlers_node.go:326` | `matter_maintenance.go:67-116` | ui | `loom:internal/north/rest/handlers/matter_maintenance.go` | ok [tagged ui] | code |
| `fabricui/handlers_node.go:376` | `node/factory_reset.go:20-45` | ui | `gf:node/factory_reset.go` | ok [tagged ui] | code |
| `fabricui/handlers_node.go:503` | `node/surface.go:33-36` | gf | `gf:node/surface.go` | ok | code |
| `fabricui/handlers_node.go:573` | `node/served.go:159-184` | ui | `gf:node/served.go` | ok [tagged ui] | code |
| `fabricui/handlers_node.go:610` | `im/eventlog.go:443` | ui | `gf:im/eventlog.go` | ok [tagged ui] | code |
| `fabricui/handlers_exposure.go:115` | `internal/north/rest/handlers/matter_exposures.go:86-150` | ui | `loom:internal/north/rest/handlers/matter_exposures.go` | ok [tagged ui] | code |
| `fabricui/handlers_exposure.go:187` | `matter_exposures.go:253-311` | ui | `loom:internal/north/rest/handlers/matter_exposures.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:20` | `internal/north/rest/handlers/matter.go:320-331` | ui | `loom:internal/north/rest/handlers/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:59` | `matter.go:333-347` | ui | `gf:contract/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:76` | `sessionparams.go:49-66` | ui | `gf:secure/sessionparams/sessionparams.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:92` | `sessionparams.go:59-63` | ui | `gf:secure/sessionparams/sessionparams.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:111` | `matter.go:374-400` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:131` | `matter.go:404-413` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:143` | `matter.go:415-422` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:151` | `matter.go:424-429` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:160` | `matter.go:445-457` | ui | `gf:contract/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:171` | `node/reannounce.go:12-26` | ui | `gf:node/reannounce.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:193` | `matter.go:465-477` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:226` | `matter.go:496-506` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:227` | `cmd/openccu-loom/daemon_matter.go:1307-1360` | loom | `loom:cmd/openccu-loom/daemon_matter.go` | ok | code |
| `fabricui/handlers_diagnostics.go:309` | `matter.go:509-515` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_diagnostics.go:318` | `matter.go:517-523` | ui | `gf:contract/matter.go` | ok [tagged ui] | code |
| `fabricui/handlers_diagnostics.go:334` | `cmd/openccu-loom/daemon_matter.go:1362-1418` | loom | `loom:cmd/openccu-loom/daemon_matter.go` | ok | code |
| `fabricui/handlers_diagnostics.go:378` | `matter.go:572-600` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/handlers_controller_test.go:715` | `diagevent/ring.go:210` | ui | `gf:diagevent/ring.go` | ok [tagged ui] | test |
| `fabricui/handlers_controller_review_test.go:35` | `valuejson.go:82` | gf | `gf:internal/valuejson/valuejson.go` | ok | test |
| `fabricui/handlers_controller.go:540` | `controller/client/event_history.go:100-115` | gf | `gf:controller/client/event_history.go` | ok | code |
| `fabricui/handlers_controller.go:573` | `controllerapi/cmd_node.go:167-204` | ui | `` | unresolved | code |
| `fabricui/handlers_controller.go:947` | `ControllerCommandHandler.ts:2618-2668` | ui | `` | unresolved | code |
| `fabricui/handlers_controller.go:1529` | `packages/protocol/src/fabric/Fabric.ts:193-196` | ui | `` | unresolved | code |
| `fabricui/handler.go:303` | `internal/north/rest/router.go:1328-1359` | loom | `loom:internal/north/rest/router.go` | ok | code |
| `fabricui/handler.go:608` | `router.go:1336-1353` | ui | `loom:internal/north/rest/router.go` | ok [tagged ui] | code |
| `fabricui/events.go:113` | `matter.go:558-565` | loom | `loom:pkg/interfaces/matter.go,internal/north/rest/handlers/matter.go,internal/model/generic/matter.go` | ambiguous(10) | code |
| `fabricui/events.go:251` | `node/observers.go:63-71` | gf | `gf:node/observers.go` | ok | code |
| `fabricui/events.go:253` | `observers.go:121-122` | ui | `gf:node/observers.go` | ok [tagged ui] | code |
| `fabricui/events.go:255` | `node/observers.go:75-84` | gf | `gf:node/observers.go` | ok | code |
| `fabricui/e2e_test.go:454` | `node/factory_reset.go:20-45` | ui | `gf:node/factory_reset.go` | ok [tagged ui] | test |
| `fabricui/e2e_test.go:454` | `fabrictest/device.go:396-416` | ui | `gf:fabrictest/device.go` | ok [tagged ui] | test |
| `fabricui/controller.go:694` | `controller/controller.go:611-618` | ui | `gf:controller/controller.go` | ok [tagged ui] | code |
| `fabricui/controller.go:744` | `controller/commissioner/progress.go:67-73` | ui | `gf:controller/commissioner/progress.go` | ok [tagged ui] | code |
| `fabricui/controller.go:885` | `observe.go:29-59` | ui | `gf:controller/softwareupdate/observe.go` | ok [tagged ui] | code |
| `fabricui/controller.go:931` | `errors.go:158-176` | ui | `gf:thread/errors.go,thread/otbr/errors.go,controller/commissioner/errors.go` | ambiguous(6) | code |
| `fabricui/gofabric/updates.go:62` | `observe.go:130-139` | ui | `gf:controller/softwareupdate/observe.go` | ok [tagged ui] | code |
| `fabricui/gofabric/updates.go:77` | `observe.go:99-103` | ui | `gf:controller/softwareupdate/observe.go` | ok [tagged ui] | code |
| `fabricui/gofabric/topology.go:218` | `thread/brdiscovery/registry.go:44` | ui | `gf:thread/brdiscovery/registry.go` | ok [tagged ui] | code |
| `fabricui/gofabric/topology.go:225` | `NetworkTopologyService.ts:452-478` | ui | `` | unresolved | code |
| `fabricui/gofabric/topology.go:226` | `controller/topology/build.go:92-101` | gf | `gf:controller/topology/build.go` | ok | code |
| `fabricui/gofabric/node.go:416` | `controller/client/event_history.go:100-115` | ui | `gf:controller/client/event_history.go` | ok [tagged ui] | code |
| `fabricui/gofabric/node.go:446` | `controllerapi/cmd_interact.go:68-110` | gf | `` | unresolved | code |
| `fabricui/gofabric/node.go:549` | `controllerapi/cmd_interact.go:285-340` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:550` | `cmd_acl.go:37-62` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:589` | `client/interaction/write.go:259-275` | ui | `gf:client/interaction/write.go` | ok [tagged ui] | code |
| `fabricui/gofabric/node.go:628` | `controllerapi/cmd_interact.go:344-400` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:660` | `controllerapi/cmd_node.go:303-334` | gf | `` | unresolved | code |
| `fabricui/gofabric/node.go:678` | `controllerapi/cmd_node.go:99-124` | gf | `` | unresolved | code |
| `fabricui/gofabric/node.go:690` | `controllerapi/cmd_node.go:187` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:710` | `controllerapi/cmd_node.go:167-204` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:787` | `cmd_node.go:190-193` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:812` | `controllerapi/cmd_icd.go:47-74` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:842` | `controllerapi/cmd_icd.go:121-150` | ui | `` | unresolved | code |
| `fabricui/gofabric/node.go:853` | `controllerapi/cmd_icd.go:153-180` | ui | `` | unresolved | code |
| `fabricui/gofabric/gofabric_test.go:189` | `client/discovery/session.go:73-76` | ui | `gf:client/discovery/session.go` | ok [tagged ui] | test |
| `fabricui/gofabric/gofabric.go:245` | `controllerapi/cmd_node.go:24-35` | ui | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:301` | `client/discovery/session.go:73-93` | ui | `gf:client/discovery/session.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:303` | `controller/controller.go:674-676` | ui | `gf:controller/controller.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:325` | `controllerapi/cmd_commission.go:132-135` | ui | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:366` | `controller/controller.go:611-618` | ui | `gf:controller/controller.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:368` | `controller/commissioner/progress.go:15-32` | ui | `gf:controller/commissioner/progress.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:399` | `controller.go:654-663` | ui | `ui:fabricui/controller.go` | ok | code |
| `fabricui/gofabric/gofabric.go:400` | `controller/commissioner/commissioner.go:154-167` | ui | `gf:controller/commissioner/commissioner.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:402` | `controller/controller.go:752-768` | ui | `gf:controller/controller.go` | ok [tagged ui] | code |
| `fabricui/gofabric/gofabric.go:415` | `errors.go:158-176` | ui | `gf:thread/errors.go,thread/otbr/errors.go,controller/commissioner/errors.go` | ambiguous(6) | code |
| `fabricui/gofabric/gofabric.go:458` | `controllerapi/cmd_commission.go:335-347` | gf | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:564` | `controllerapi/cmd_commission.go:159-232` | ui | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:622` | `controllerapi/cmd_node.go:81-94` | gf | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:645` | `controllerapi/cmd_server.go:217-286` | ui | `` | unresolved | code |
| `fabricui/gofabric/gofabric.go:772` | `controller/client/fabricops.go:229-235` | ui | `gf:controller/client/fabricops.go` | ok [tagged ui] | code |
| `fabricui/gofabric/e2e_test.go:863` | `cluster/core/ahead.go:47` | gf | `gf:cluster/core/ahead.go` | ok | test |
| `fabricui/gofabric/e2e_test.go:915` | `controller/commissioner/errors.go:158-176` | ui | `gf:controller/commissioner/errors.go` | ok [tagged ui] | test |
| `fabricui/gofabric/e2e_test.go:1049` | `contract/wifi.go:17-25` | ui | `gf:contract/wifi.go` | ok [tagged ui] | test |
| `fabricui/gofabric/e2e_test.go:1050` | `controller/ble_test.go:24-38` | gf | `gf:controller/ble_test.go` | ok | test |
| `fabricui/gofabric/e2e_test.go:1085` | `controller/pair_race.go:25-26` | ui | `gf:controller/pair_race.go` | ok [tagged ui] | test |
| `fabricui/gofabric/e2e_test.go:1132` | `client/interaction/write.go:259-275` | ui | `gf:client/interaction/write.go` | ok [tagged ui] | test |
| `fabricui/gofabric/diagnostics.go:87` | `diagevent/ring.go:210-231` | ui | `gf:diagevent/ring.go` | ok [tagged ui] | code |
| `docs/embedding.md:290` | `controller/commissioner/network.go:217-235` | ui | `gf:controller/commissioner/network.go` | ok [tagged ui] | docs |
| `docs/embedding.md:310` | `controller/topology/service.go:57-60` | gf | `gf:controller/topology/service.go` | ok | docs |
| `docs/embedding.md:311` | `controller/diagnostics.go:46` | gf | `gf:controller/diagnostics.go` | ok | docs |
| `docs/contract.md:650` | `diag/diag.go:50` | ui | `gf:diag/diag.go` | ok [tagged ui] | docs |
| `docs/contract.md:667` | `mdns/diagnose_browse.go:14` | ui | `gf:mdns/diagnose_browse.go` | ok [tagged ui] | docs |
| `docs/contract.md:677` | `diagevent/ring.go:49-58` | ui | `gf:diagevent/ring.go` | ok [tagged ui] | docs |
| `docs/concept-u6-nodes.md:15` | `fabricui/options.go:103-109` | ui | `ui:fabricui/options.go` | ok | docs |
| `docs/concept-u6-nodes.md:17` | `internal/daemon/nodes.go:21-34` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:20` | `cmd/ha2fabric/main.go:87` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:21` | `internal/selection/selection.go:51-74` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:22` | `internal/ha/client.go:85` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:26` | `docs-site/docs/getting-started/bridge-configuration.md:3-4` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:27` | `packages/backend/src/services/bridges/bridge-service.ts:80` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:28` | `packages/common/src/bridge-data.ts:219-222` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:31` | `bridge-configuration.md:28-33` | ui | `` | unresolved | docs |
| `docs/concept-u6-nodes.md:113` | `addon/config.yaml:37` | ui | `` | unresolved | docs |
| `docs/concept-u5-operation.md:3` | `docs/roadmap.md:378` | gf | `gf:docs/roadmap.md` | ok | docs |
| `docs/concept-u5-operation.md:4` | `docs/concept-v13-roadmap-extension.md:99` | ui | `gf:docs/concept-v13-roadmap-extension.md` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:16` | `events.go:14-31` | ui | `ui:fabricui/events.go` | ok | docs |
| `docs/concept-u5-operation.md:20` | `node/health.go:93` | gf | `gf:node/health.go` | ok | docs |
| `docs/concept-u5-operation.md:21` | `node/diagnostics.go:40` | ui | `gf:node/diagnostics.go` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:22` | `fabricui/options.go:25` | ui | `ui:fabricui/options.go` | ok | docs |
| `docs/concept-u5-operation.md:25` | `controllerapi/cmd_server.go:76-102` | ui | `` | unresolved | docs |
| `docs/concept-u5-operation.md:26` | `internal/daemon/run.go:184` | ui | `` | unresolved | docs |
| `docs/concept-u5-operation.md:26` | `cmd_node.go:167` | ui | `` | unresolved | docs |
| `docs/concept-u5-operation.md:32` | `node/config.go:217-220` | ui | `gf:node/config.go` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:40` | `docs/ECOSYSTEMS.md:85-131` | mjs | `mjs:docs/ECOSYSTEMS.md` | ok | docs |
| `docs/concept-u5-operation.md:41` | `options.go:62-65` | ui | `ui:fabricui/options.go` | ok | docs |
| `docs/concept-u5-operation.md:42` | `run.go:370` | ui | `` | unresolved | docs |
| `docs/concept-u5-operation.md:114` | `ECOSYSTEMS.md:85-131` | ui | `mjs:docs/ECOSYSTEMS.md` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:131` | `node/health.go:93` | ui | `gf:node/health.go` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:131` | `node/diagnostics.go:40` | ui | `gf:node/diagnostics.go` | ok [tagged ui] | docs |
| `docs/concept-u5-operation.md:136` | `run.go:370` | ui | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:7` | `docs/roadmap.md:377` | ui | `gf:docs/roadmap.md` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:19` | `fabricui/handlers_diagnostics.go:172-198` | ui | `ui:fabricui/handlers_diagnostics.go` | ok | docs |
| `docs/concept-u4-device-details.md:22` | `contract/openapi.yaml:1454` | ui | `ui:contract/openapi.yaml` | ok | docs |
| `docs/concept-u4-device-details.md:24` | `fabricui/events.go:16-28` | ui | `ui:fabricui/events.go` | ok | docs |
| `docs/concept-u4-device-details.md:25` | `node/observers.go:116` | ui | `gf:node/observers.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:26` | `node/lifecycle.go:128-137` | ui | `gf:node/lifecycle.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:32` | `concept-u2-controller-family.md:170` | ui | `ui:docs/concept-u2-controller-family.md` | ok | docs |
| `docs/concept-u4-device-details.md:33` | `contract.md:255-258` | ui | `ui:docs/contract.md` | ok | docs |
| `docs/concept-u4-device-details.md:38` | `controller/client/conformance.go:93` | ui | `gf:controller/client/conformance.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:40` | `controller/client/config.go:123` | ui | `gf:controller/client/config.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:41` | `fabricui/gofabric/hooks.go:199` | ui | `ui:fabricui/gofabric/hooks.go` | ok | docs |
| `docs/concept-u4-device-details.md:42` | `internal/daemon/run.go:245-247` | gf | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:44` | `controllerapi/events.go:15` | ui | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:47` | `internal/arch/arch_test.go:29-42` | ui | `ui:internal/arch/arch_test.go` | ok | docs |
| `docs/concept-u4-device-details.md:50` | `schema/lookup.go:20` | ui | `gf:schema/lookup.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:52` | `schema/devicetype_requirements.go:338` | ui | `gf:schema/devicetype_requirements.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:55` | `matter-dashboard-app.ts:208-229` | ui | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:57` | `matter-cluster-view.ts:432-460` | ui | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:57` | `matter-endpoint-view.ts:89` | ui | `` | unresolved | docs |
| `docs/concept-u4-device-details.md:78` | `concept-u3-topology.md:177-182` | ui | `ui:docs/concept-u3-topology.md` | ok | docs |
| `docs/concept-u4-device-details.md:84` | `contract.md:262-264` | ui | `ui:docs/contract.md` | ok | docs |
| `docs/concept-u4-device-details.md:104` | `im/eventlog.go:19-45` | ui | `gf:im/eventlog.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:108` | `fixtures.test.ts:1-29` | ui | `ui:ui/src/lib/api/fixtures.test.ts` | ok | docs |
| `docs/concept-u4-device-details.md:108` | `fabricui/misc_test.go:308-311` | ui | `ui:fabricui/misc_test.go` | ok | docs |
| `docs/concept-u4-device-details.md:122` | `cluster/spec/model.go:332` | ui | `gf:cluster/spec/model.go` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:141` | `context.ts:58-69` | ui | `ui:ui/src/lib/context.ts` | ok | docs |
| `docs/concept-u4-device-details.md:151` | `coverage.test.ts:46-61` | ui | `ui:ui/src/lib/i18n/coverage.test.ts` | ok | docs |
| `docs/concept-u4-device-details.md:169` | `docs/roadmap.md:105` | ui | `gf:docs/roadmap.md` | ok [tagged ui] | docs |
| `docs/concept-u4-device-details.md:199` | `node/lifecycle.go:129-133` | ui | `gf:node/lifecycle.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:15` | `contract/openapi.yaml:1012-1029` | ui | `ui:contract/openapi.yaml` | ok | docs |
| `docs/concept-u3-topology.md:15` | `fabricui/handlers_diagnostics.go:167` | ui | `ui:fabricui/handlers_diagnostics.go` | ok | docs |
| `docs/concept-u3-topology.md:17` | `openapi.yaml:1037-1052` | ui | `ui:contract/openapi.yaml,contract/previous/openapi.yaml` | ambiguous(2) | docs |
| `docs/concept-u3-topology.md:22` | `node/lifecycle.go:89-132` | ui | `gf:node/lifecycle.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:23` | `node/node.go:1057` | ui | `gf:node/node.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:25` | `node/inspect.go:61-108` | ui | `gf:node/inspect.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:27` | `node/observers.go:91-100` | ui | `gf:node/observers.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:29` | `openapi.yaml:297-312` | ui | `ui:contract/openapi.yaml,contract/previous/openapi.yaml` | ambiguous(2) | docs |
| `docs/concept-u3-topology.md:29` | `fabricui/options.go:27-51` | ui | `ui:fabricui/options.go` | ok | docs |
| `docs/concept-u3-topology.md:33` | `controllerapi/cmd_topology.go:106` | ui | `` | unresolved | docs |
| `docs/concept-u3-topology.md:36` | `cmd_topology.go:52-91` | ui | `` | unresolved | docs |
| `docs/concept-u3-topology.md:38` | `NetworkTopologyService.ts:304-430` | ui | `` | unresolved | docs |
| `docs/concept-u3-topology.md:40` | `fabricui/controller.go:22-56` | ui | `ui:fabricui/controller.go` | ok | docs |
| `docs/concept-u3-topology.md:42` | `server_gen.go:22` | ui | `gf:cluster/spec/zonemanagement/server_gen.go,cluster/spec/windowcovering/server_gen.go,cluster/spec/wifinetworkmanagement/server_gen.go` | ambiguous(132) | docs |
| `docs/concept-u3-topology.md:119` | `NetworkTopologyService.ts:304-430` | mjs | `` | unresolved | docs |
| `docs/concept-u3-topology.md:122` | `cmd_topology.go:166` | ui | `` | unresolved | docs |
| `docs/concept-u3-topology.md:169` | `web.go:21` | ui | `` | unresolved | docs |
| `docs/concept-u3-topology.md:194` | `im/subscription/manager.go:846` | ui | `gf:im/subscription/manager.go` | ok [tagged ui] | docs |
| `docs/concept-u3-topology.md:200` | `cmd_topology.go:28-31` | ui | `` | unresolved | docs |
| `docs/concept-u2-controller-family.md:15` | `contract/openapi.yaml:292-332` | ui | `ui:contract/openapi.yaml` | ok | docs |
| `docs/concept-u2-controller-family.md:21` | `internal/daemon/web.go:25-34` | ui | `` | unresolved | docs |
| `docs/concept-u2-controller-family.md:29` | `cmd_acl.go:53` | ui | `` | unresolved | docs |
| `docs/reports/U8.md:31` | `client/discovery/session.go:73-93` | ui | `gf:client/discovery/session.go` | ok [tagged ui] | docs |
| `docs/reports/U8.md:32` | `controller/controller.go:674-676` | ui | `gf:controller/controller.go` | ok [tagged ui] | docs |
| `docs/reports/U8.md:33` | `controller/pair_race.go:49-53` | ui | `gf:controller/pair_race.go` | ok [tagged ui] | docs |
| `docs/reports/U8.md:35` | `controllerapi/cmd_commission.go:132-135` | gf | `` | unresolved | docs |
| `docs/reports/U8.md:151` | `controller/commissioner/network.go:217-235` | ui | `gf:controller/commissioner/network.go` | ok [tagged ui] | docs |
| `docs/reports/U8.md:154` | `ui/src/lib/kit/Select.svelte:31-35` | ui | `ui:ui/src/lib/kit/Select.svelte` | ok | docs |
| `docs/reports/U8.md:173` | `internal/daemon/controllerui.go:66-72` | ui | `` | unresolved | docs |
| `docs/reports/U8.md:174` | `controllerapi/settings.go:21-23` | ui | `` | unresolved | docs |
| `docs/reports/U8.md:178` | `cmd_commission.go:76-93` | ui | `` | unresolved | docs |
| `docs/reports/U8.md:180` | `internal/daemon/web.go:39` | ui | `` | unresolved | docs |
| `docs/reports/U8.md:181` | `registry.go:202` | ui | `ui:fabricui/registry.go` | ok | docs |
| `docs/reports/U8.md:187` | `cmd/ha2fabric/main.go:112-116` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:28` | `node/observers.go:63` | ui | `gf:node/observers.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:28` | `controller/controller.go:611-618` | ui | `gf:controller/controller.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:28` | `controller/commissioner/progress.go:15-32` | ui | `gf:controller/commissioner/progress.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:28` | `controller/softwareupdate/observe.go:29-59` | ui | `gf:controller/softwareupdate/observe.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:29` | `node/attestation.go:55-58` | ui | `gf:node/attestation.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:29` | `node/node.go:1138-1141` | ui | `gf:node/node.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:29` | `node/sessions.go:110-118` | ui | `gf:node/sessions.go` | out-of-range(len 113) [tagged ui] | docs |
| `docs/reports/U7.md:29` | `secure/sessionparams/sessionparams.go:49-66` | ui | `gf:secure/sessionparams/sessionparams.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:30` | `node/reannounce.go:12-26` | ui | `gf:node/reannounce.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:31` | `node/factory_reset.go:20-45` | ui | `gf:node/factory_reset.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:31` | `fabrictest/device.go:396-416` | ui | `gf:fabrictest/device.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:32` | `fabricui/handlers_surface.go:21` | ui | `ui:fabricui/handlers_surface.go` | ok | docs |
| `docs/reports/U7.md:33` | `controller/commissioner/commissioner.go:154-163` | ui | `gf:controller/commissioner/commissioner.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:33` | `flow.go:172` | ui | `gf:controller/commissioner/flow.go,cluster/measurement/flow.go` | ambiguous(2) | docs |
| `docs/reports/U7.md:33` | `flow.go:408-419` | ui | `gf:controller/commissioner/flow.go,cluster/measurement/flow.go` | ambiguous(2) | docs |
| `docs/reports/U7.md:33` | `cluster/spec/generalcommissioning/definition_gen.go:110-112` | ui | `gf:cluster/spec/generalcommissioning/definition_gen.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:33` | `fabrictest/device.go:59-63` | ui | `gf:fabrictest/device.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:33` | `node/config.go:402-422` | ui | `gf:node/config.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:164` | `observers.go:127-136` | ui | `gf:node/observers.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:166` | `observe.go:99-103` | ui | `gf:controller/softwareupdate/observe.go` | ok [tagged ui] | docs |
| `docs/reports/U7.md:196` | `cmd/ha2fabric/main.go:109` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:196` | `nodes.go:190` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:200` | `nodes.go:45` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:203` | `nodes.go:270-301` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:206` | `internal/daemon/web.go:39` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:206` | `registry.go:202` | ui | `ui:fabricui/registry.go` | ok | docs |
| `docs/reports/U7.md:207` | `controllerui.go:30` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:209` | `internal/daemon/registry.go:303` | ui | `` | unresolved | docs |
| `docs/reports/U7.md:211` | `web.go:39-40` | ui | `` | unresolved | docs |
| `docs/reports/U3b.md:145` | `controller/topology/thread.go:166-183` | ui | `gf:controller/topology/thread.go` | ok [tagged ui] | docs |
| `docs/reports/U3b.md:146` | `build.go:141` | ui | `gf:controller/topology/build.go` | ok [tagged ui] | docs |
| `docs/reports/U3b.md:203` | `run.go:246-252` | ui | `` | unresolved | docs |
| `docs/reports/U3b.md:205` | `controllerui.go:25-33` | ui | `` | unresolved | docs |
| `docs/reports/U1b.md:96` | `cmd/ha2fabric/main.go:83` | ui | `` | unresolved | docs |
| `docs/reports/U1b.md:97` | `internal/daemon/run.go:315` | ui | `` | unresolved | docs |

