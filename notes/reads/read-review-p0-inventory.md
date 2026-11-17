# Review phase 0 — inventory (read-only map)

- Repository HEAD: `ee88372f2802851a07a9e0adc0e9a75e56c3f788` (worktree `/home/markus/Dokumente/GitHub/go-fabric-ui-review`). Code baseline `25f6e7b4ad743d25932406b88e386d945f47c54f`; `git diff --name-only 25f6e7b4 HEAD` lists only `notes/review/2026-11-17/README.md` and `notes/review/2026-11-17/decisions.md`, so code at HEAD equals the baseline.
- Host pins: `go-fabric-server@d61a545`, `go-ha2fabric@ec7deea` (exports under the session scratchpad `pins/`). Both `go.mod:7` require `github.com/SukramJ/go-fabric-ui v0.9.2`; neither has a `replace` line (grep of `replace` in both `go.mod` is empty).
- Paths are relative to the repo root unless prefixed with a host pin. No opinions below; "unclear" marks what I did not resolve.
- Method note: the contract rows (operation, `x-admin`, request body, SSE line) were extracted by a script over `contract/openapi.yaml` lines 1..(components start); the code-side route/gate sets were compared to them by script (result under "Cross-check summary").

---

## 1. Routes and their gate

### 1.1 Registration mechanism

| Fact | Where |
|---|---|
| `register` puts every route on a `*http.ServeMux`; `read` = plain `mux.HandleFunc`, `write` = `mux.HandleFunc(pattern, h.adminOnly(fn))` | `fabricui/handler.go:326-329` |
| Node family: `nodeRoutes` table (24 entries), each registered twice: `<prefix>/node<path>` and `<prefix>/nodes/{node}<path>`; the `{node}` form resolves `h.nodeHandler(r.PathValue("node"))`, 404 problem when none | table `fabricui/handler.go:286-309`; loop `:331-345`; nodeHandler call `:338-343` |
| `nodeRoute.admin` flag selects `write` over `read` | `fabricui/handler.go:274-280` (type), `:332-335` |
| Registry routes (3), all `adminOnlyBecause(..., "this route creates, renames or deletes a node")` | table `fabricui/handler.go:314-318`; loop `:347-350` |
| `GET <prefix>/nodes` plain | `fabricui/handler.go:346` |
| Operation routes: 4 plain, `PUT /operation/log/level` via `adminOnlyBecause` | `fabricui/handler.go:351-355` |
| `operationAdmin` map `{"PUT /operation/log/level": true}` declared; referenced in non-test code: nowhere (grep `operationAdmin`); referenced in `fabricui/handlers_nodes_test.go:180` | `fabricui/handler.go:320-321` |
| Controller family mounted only when `h.ctrl != nil` | `fabricui/handler.go:356-358` |
| Catch-all `<prefix>/` -> 404 problem "No such route" (not in contract) | `fabricui/handler.go:359-363` |
| Gate function `adminOnlyBecause`: `!h.isAdmin(r)` -> 403 problem "Admin rights required" | `fabricui/handler.go:477-485` |
| `isAdmin`: `h.admin != nil` -> `h.admin(r)`; else header `X-Fabric-UI-Token` (const `TokenHeader`, `handler.go:26`) non-empty and `subtle.ConstantTimeCompare` with `h.token` | `fabricui/handler.go:487-493` |
| Token: 16 random bytes hex, generated only when `h.admin == nil` | `fabricui/handler.go:265-269`; copied into per-node handlers `:402` |
| Token mode / Admin doc | `fabricui/options.go:218-225`; contract `contract/openapi.yaml:19-21` |
| Controller gate `controllerGate`: looks up the matched generated pattern, `controllerAdmin[route] && !isAdmin` -> 403; `controllerAdminQuery` (`GET /controller/topology`, param `refresh`) -> 403 only if `ParseBool(refresh)` is true and not admin; then `MaxBytesReader(maxBodyBytes)`; then `sub.ServeHTTP` | `fabricui/handlers_controller.go:120-142` (gate `:125-129`, query `:130-136`, body `:137-139`) |
| `controllerAdmin` table (22 entries) | `fabricui/handlers_controller.go:51-74` |
| `controllerAdminQuery` | `fabricui/handlers_controller.go:79-81` |
| Controller mount: generated routes on a private `sub` mux via `api.HandlerWithOptions`; `sub` 404 for other `/controller/` paths; `mux.Handle(prefix+"/controller/", controllerGate(sub))` | `fabricui/handlers_controller.go:98-117` |
| Body bound constant `maxBodyBytes = 1 << 20` | `fabricui/problem.go:31-33` |
| `Mount(mux, o)`: `newHandler` + `h.register(mux)` only (no SPA); recovers a duplicate-pattern panic | `fabricui/options.go:307-320` |
| `Handler(o)`: new `ServeMux`, `h.register`, then `mux.Handle("/", spaHandler())`; invalid prefix panics | `fabricui/options.go:322-333` |
| `validPrefix`: must start `/`, not end `/`, no space/`{`/`}` | `fabricui/options.go:335-340` |
| SPA handler `spaHandler`/`spaHandlerFS`: embeds `dist` (`//go:embed all:dist`, `spa.go:15-16`); `index.html`/empty/trailing-slash path -> `serveIndex`; file open error or directory -> `serveIndex`; else file with `Cache-Control` and `http.FileServer`; no method check | `fabricui/spa.go:26-58` |
| `serveIndex`: `index.html` with `Content-Type`+`no-store`; 503 + inline `notBuiltHTML` when the bundle is missing | `fabricui/spa.go:68-80` |
| SSE loop `serveSSE` shared by `/node/stream`, `/controller/stream`, `/operation/log/stream`; buffer 64, `push` drops when full, heartbeat `h.heartbeat` (15 s at `handler.go:250`) | `fabricui/stream.go:17`, `:34-78`; heartbeat `fabricui/handler.go:250` |

### 1.2 Node family (each row = `/node<p>` and `/nodes/{node}<p>`; contract lines given as `/node` / `/nodes/{node}`)

| # | Method | Path (`/node`+p, and `/nodes/{node}`+p) | Handler | Registered | Gate in code | Stream | Reads body (code) | Body in contract | In contract | x-admin in contract | Gated in code |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | GET | `/node/status` and `/nodes/{node}/status` | `status` fabricui/handlers_node.go:137 | handler.go:286 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L41 / L574 | n | n |
| 2 | GET | `/node/fabrics` and `/nodes/{node}/fabrics` | `fabrics` fabricui/handlers_node.go:238 | handler.go:287 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L52 / L589 | n | n |
| 3 | DELETE | `/node/fabrics/{index}` and `/nodes/{node}/fabrics/{index}` | `removeFabric` fabricui/handlers_node.go:270 | handler.go:288 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | no | n | y L65 / L606 | y L69 / L611 | y |
| 4 | GET | `/node/pairing` and `/nodes/{node}/pairing` | `pairing` fabricui/handlers_pairing.go:58 | handler.go:289 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | no | n | y L82 / L627 | y L86 / L632 | y |
| 5 | POST | `/node/pairing/window` and `/nodes/{node}/pairing/window` | `openWindow` fabricui/handlers_pairing.go:134 | handler.go:290 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | fabricui/handlers_pairing.go:140 (decodeJSON) | y L101 / L651 | y L96 / L645 | y L100 / L650 | y |
| 6 | POST | `/node/pairing/window/close` and `/nodes/{node}/pairing/window/close` | `closeWindow` fabricui/handlers_pairing.go:244 | handler.go:291 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | no | n | y L118 / L671 | y L122 / L676 | y |
| 7 | GET | `/node/endpoints` and `/nodes/{node}/endpoints` | `endpoints` fabricui/handlers_diagnostics.go:229 | handler.go:292 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L129 / L686 | n | n |
| 8 | GET | `/node/endpoints/{ep}` and `/nodes/{node}/endpoints/{ep}` | `nodeEndpoint` fabricui/handlers_node.go:484 | handler.go:293 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L149 / L703 | n | n |
| 9 | GET | `/node/endpoints/{ep}/events` and `/nodes/{node}/endpoints/{ep}/events` | `nodeEndpointEvents` fabricui/handlers_node.go:554 | handler.go:294 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L181 / L722 | n | n |
| 10 | GET | `/node/surface` and `/nodes/{node}/surface` | `surfaceReport` fabricui/handlers_surface.go:16 | handler.go:295 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L204 / L742 | n | n |
| 11 | GET | `/node/sessions` and `/nodes/{node}/sessions` | `sessions` fabricui/handlers_diagnostics.go:113 | handler.go:296 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L223 / L758 | n | n |
| 12 | GET | `/node/topology` and `/nodes/{node}/topology` | `topology` fabricui/handlers_topology.go:88 | handler.go:297 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L235 / L774 | n | n |
| 13 | GET | `/node/mdns` and `/nodes/{node}/mdns` | `mdns` fabricui/handlers_diagnostics.go:160 | handler.go:298 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L257 / L791 | n | n |
| 14 | POST | `/node/mdns/reannounce` and `/nodes/{node}/mdns/reannounce` | `reannounce` fabricui/handlers_diagnostics.go:172 | handler.go:299 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | no | n | y L269 / L807 | y L279 / L812 | y |
| 15 | GET | `/node/compatibility` and `/nodes/{node}/compatibility` | `compatibility` fabricui/handlers_diagnostics.go:335 | handler.go:300 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L285 / L821 | n | n |
| 16 | GET | `/node/events` and `/nodes/{node}/events` | `diagnosticEvents` fabricui/handlers_diagnostics.go:374 | handler.go:301 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L297 / L837 | n | n |
| 17 | GET | `/node/exposure/schema` and `/nodes/{node}/exposure/schema` | `exposureSchema` fabricui/handlers_exposure.go:90 | handler.go:302 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L309 / L853 | n | n |
| 18 | GET | `/node/exposure` and `/nodes/{node}/exposure` | `exposureList` fabricui/handlers_exposure.go:117 | handler.go:303 -> :331-345 | plain (`read`, handler.go:328) | no | no | n | y L322 / L870 | n | n |
| 19 | PUT | `/node/exposure` and `/nodes/{node}/exposure` | `exposureApply` fabricui/handlers_exposure.go:189 | handler.go:304 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | fabricui/handlers_exposure.go:133 (decodeJSON in readUpdates) | y L339 / L890 | y L334 / L884 | y L338 / L889 | y |
| 20 | POST | `/node/exposure/preview` and `/nodes/{node}/exposure/preview` | `exposurePreview` fabricui/handlers_exposure.go:158 | handler.go:305 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | fabricui/handlers_exposure.go:133 (decodeJSON in readUpdates) | y L361 / L916 | y L356 / L910 | y L360 / L915 | y |
| 21 | POST | `/node/maintenance/reassemble` and `/nodes/{node}/maintenance/reassemble` | `maintenanceReassemble` fabricui/handlers_node.go:293 | handler.go:306 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | no | n | y L377 / L935 | y L381 / L940 | y |
| 22 | POST | `/node/maintenance/factory-reset` and `/nodes/{node}/maintenance/factory-reset` | `factoryReset` fabricui/handlers_node.go:317 | handler.go:307 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | fabricui/handlers_node.go:325-326 (Decode, error dropped) | y L393 / L956 | y L388 / L950 | y L392 / L955 | y |
| 23 | POST | `/node/maintenance/erase` and `/nodes/{node}/maintenance/erase` | `eraseNode` fabricui/handlers_node.go:364 | handler.go:308 -> :331-345 | `adminOnly` (handler.go:329, :472-474) | no | fabricui/handlers_node.go:376-377 (Decode, error dropped) | y L420 / L977 | y L405 / L971 | y L419 / L976 | y |
| 24 | GET | `/node/stream` and `/nodes/{node}/stream` | `stream` fabricui/stream.go:24 | handler.go:309 -> :331-345 | plain (`read`, handler.go:328) | SSE (stream.go:44-47) | no | n | y L433 / L993 | n | n |

### 1.3 Registry, node list, operation routes, fall-throughs

| # | Method | Path | Handler | Registered | Gate in code | Stream | Reads body (code) | In contract | x-admin in contract | Gated in code |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | GET | `/nodes` | `listNodes` fabricui/handler.go:452 | handler.go:346 | plain | no | no | y L479 | n | n |
| 2 | POST | `/nodes` | `createNode` fabricui/registry.go:260 | handler.go:315, :347-350 | `adminOnlyBecause` | no | fabricui/registry.go:246 (decodeJSON in readNodeLabel; called registry.go:265) | y L497 (body L509) | y L508 | y |
| 3 | PUT | `/nodes/{node}/label` | `renameNode` fabricui/registry.go:290 | handler.go:316, :347-350 | `adminOnlyBecause` | no | fabricui/registry.go:246 (readNodeLabel, called registry.go:295) | y L549 (body L558) | y L557 | y |
| 4 | DELETE | `/nodes/{node}` | `deleteNode` fabricui/registry.go:309 | handler.go:317, :347-350 | `adminOnlyBecause` | no | no | y L528 | y L538 | y |
| 5 | GET | `/operation/info` | `operationInfo` fabricui/operations.go:180 | handler.go:351 | plain | no | no | y L1006 | n | n |
| 6 | GET | `/operation/log/level` | `logLevel` fabricui/operations.go:193 | handler.go:352 | plain | no | no | y L1018 | n | n |
| 7 | PUT | `/operation/log/level` | `setLogLevel` fabricui/operations.go:202 | handler.go:353 | `adminOnlyBecause` | no | fabricui/operations.go:208 (decodeJSON) | y L1029 (body L1034) | y L1033 | y |
| 8 | GET | `/operation/log` | `logRecords` fabricui/operations.go:227 | handler.go:354 | plain | no | no (query `since`,`limit`, operations.go:246-268) | y L1050 | n | n |
| 9 | GET | `/operation/log/stream` | `logStream` fabricui/operations.go:274 | handler.go:355 | plain | SSE | no | y L1071 (SSE L1085) | n | n |
| 10 | any | `<prefix>/` (unmatched) | inline 404 problem fabricui/handler.go:361-363 | handler.go:361 | none | no | no | n (not in contract) | n | n |
| 11 | any | `/` (unmatched, `Handler` only) | `spaHandler()` fabricui/spa.go:26 | fabricui/options.go:331 | none | no | no | n (not in contract) | n | n |

### 1.4 Controller family (only when `Options.Controller != nil`, handler.go:356-358; contract lines are under `contract/openapi.yaml`)

| # | Method | Path (under prefix) | Handler (`fabricui/handlers_controller.go`) | Generated route (`internal/api/api.gen.go`) | Gate in code | Stream | Reads body (code) | In contract | x-admin in contract | Gated in code |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | GET | `/controller/status` | `GetControllerStatus` :300 | :3008 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1089 | n | n |
| 2 | GET | `/controller/nodes` | `ListControllerNodes` :325 | :3009 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1104 | n | n |
| 3 | GET | `/controller/nodes/{id}` | `GetControllerNode` :419 | :3011 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1117 | n | n |
| 4 | DELETE | `/controller/nodes/{id}` | `DecommissionNode` :550 | :3010 | controllerGate -> `controllerAdmin` L52 (:125-129) | no | no | y L1129 | y L1132 | y |
| 5 | GET | `/controller/nodes/{id}/endpoints/{ep}` | `GetNodeEndpoint` :497 | :3012 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1145 | n | n |
| 6 | GET | `/controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}` | `GetNodeCluster` :515 | :3013 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1175 | n | n |
| 7 | GET | `/controller/nodes/{id}/events` | `ListNodeEvents` :538 | :3014 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1199 | n | n |
| 8 | POST | `/controller/nodes/{id}/interview` | `InterviewNode` :599 | :3015 | controllerGate -> `controllerAdmin` L53 (:125-129) | no | no | y L1224 | y L1227 | y |
| 9 | POST | `/controller/nodes/{id}/ping` | `PingNode` :565 | :3016 | controllerGate -> `controllerAdmin` L71 (:125-129) | no | no | y L1238 | y L1241 | y |
| 10 | POST | `/controller/nodes/{id}/enable` | `EnableNode` :610 | :3017 | controllerGate -> `controllerAdmin` L54 (:125-129) | no | no | y L1270 | y L1273 | y |
| 11 | POST | `/controller/nodes/{id}/disable` | `DisableNode` :621 | :3018 | controllerGate -> `controllerAdmin` L55 (:125-129) | no | no | y L1284 | y L1287 | y |
| 12 | POST | `/controller/nodes/{id}/auto-subscribe` | `SetNodeAutoSubscribe` :632 | :3019 | controllerGate -> `controllerAdmin` L56 (:125-129) | no | api.gen.go:6280 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1298 | y L1301 | y |
| 13 | POST | `/controller/nodes/{id}/window` | `OpenNodeWindow` :648 | :3020 | controllerGate -> `controllerAdmin` L57 (:125-129) | no | api.gen.go:6742 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1317 | y L1320 | y |
| 14 | POST | `/controller/nodes/{id}/window/close` | `CloseNodeWindow` :673 | :3021 | controllerGate -> `controllerAdmin` L58 (:125-129) | no | no | y L1340 | y L1343 | y |
| 15 | POST | `/controller/nodes/{id}/attributes` | `ReadNodeAttributes` :702 | :3022 | controllerGate passes (no table entry); bounds body :137-139 | no | api.gen.go:6214 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1354 | n | n |
| 16 | PUT | `/controller/nodes/{id}/attributes` | `WriteNodeAttribute` :747 | :3023 | controllerGate -> `controllerAdmin` L59 (:125-129) | no | api.gen.go:6247 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1378 | y L1381 | y |
| 17 | POST | `/controller/nodes/{id}/commands` | `InvokeNodeCommand` :771 | :3024 | controllerGate -> `controllerAdmin` L60 (:125-129) | no | api.gen.go:6372 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1401 | y L1404 | y |
| 18 | GET | `/controller/nodes/{id}/acl` | `GetNodeACL` :870 | :3025 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1424 | n | n |
| 19 | PUT | `/controller/nodes/{id}/acl` | `SetNodeACL` :915 | :3026 | controllerGate -> `controllerAdmin` L61 (:125-129) | no | api.gen.go:6181 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1437 | y L1440 | y |
| 20 | GET | `/controller/nodes/{id}/bindings` | `GetNodeBindings` :991 | :3027 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1464 | n | n |
| 21 | PUT | `/controller/nodes/{id}/bindings` | `SetNodeBindings` :1039 | :3028 | controllerGate -> `controllerAdmin` L62 (:125-129) | no | api.gen.go:6339 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1477 | y L1480 | y |
| 22 | GET | `/controller/nodes/{id}/icd` | `GetNodeIcd` :1118 | :3029 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1500 | n | n |
| 23 | POST | `/controller/nodes/{id}/icd/register` | `RegisterNodeIcd` :1126 | :3030 | controllerGate -> `controllerAdmin` L63 (:125-129) | no | api.gen.go:6618 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1515 | y L1518 | y |
| 24 | POST | `/controller/nodes/{id}/icd/unregister` | `UnregisterNodeIcd` :1141 | :3031 | controllerGate -> `controllerAdmin` L64 (:125-129) | no | api.gen.go:6654 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1538 | y L1541 | y |
| 25 | GET | `/controller/nodes/{id}/fabrics` | `ListNodeFabrics` :1179 | :3032 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1560 | n | n |
| 26 | DELETE | `/controller/nodes/{id}/fabrics/{index}` | `RemoveNodeFabric` :1207 | :3033 | controllerGate -> `controllerAdmin` L65 (:125-129) | no | no | y L1580 | y L1583 | y |
| 27 | POST | `/controller/commission` | `CommissionNode` :1259 | :3034 | controllerGate -> `controllerAdmin` L66 (:125-129) | no | api.gen.go:5836 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1593 | y L1596 | y |
| 28 | GET | `/controller/discovery` | `DiscoverDevices` :1338 | :3035 | controllerGate -> `controllerAdmin` L67 (:125-129) | no | no | y L1624 | y L1627 | y |
| 29 | GET | `/controller/credentials` | `ListCredentials` :1390 | :3037 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1652 | n | n |
| 30 | PUT | `/controller/credentials` | `SetCredential` :1412 | :3038 | controllerGate -> `controllerAdmin` L68 (:125-129) | no | api.gen.go:5917 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1664 | y L1667 | y |
| 31 | DELETE | `/controller/credentials` | `RemoveCredential` :1452 | :3036 | controllerGate -> `controllerAdmin` L69 (:125-129) | no | no | y L1680 | y L1683 | y |
| 32 | PUT | `/controller/fabric/label` | `SetFabricLabel` :1473 | :3039 | controllerGate -> `controllerAdmin` L70 (:125-129) | no | api.gen.go:6022 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1702 | y L1705 | y |
| 33 | GET | `/controller/updates` | `ListUpdates` :1488 | :3040 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1718 | n | n |
| 34 | POST | `/controller/updates/consent` | `ConsentUpdate` :1519 | :3041 | controllerGate -> `controllerAdmin` L72 (:125-129) | no | api.gen.go:6924 (BindJSONBody); 1 MiB MaxBytesReader at handlers_controller.go:137-139 | y L1729 | y L1732 | y |
| 35 | POST | `/controller/updates/check` | `CheckUpdates` :1545 | :3042 | controllerGate -> `controllerAdmin` L73 (:125-129) | no | no | y L1747 | y L1750 | y |
| 36 | GET | `/controller/topology` | `GetControllerTopology` :1570 | :3043 | controllerGate -> `controllerAdminQuery` L79-81: open unless `refresh` parses true (:130-136) | no | no | y L1762 | query param `refresh` L1789 | partial (query) |
| 37 | GET | `/controller/diagnostics` | `GetControllerDiagnostics` :1729 | :3044 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1800 | n | n |
| 38 | GET | `/controller/mdns` | `GetControllerMdns` :1826 | :3045 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1822 | n | n |
| 39 | GET | `/controller/events` | `ListControllerDiagnosticEvents` :1843 | :3046 | controllerGate passes (no table entry); bounds body :137-139 | no | no | y L1842 | n | n |
| 40 | GET | `/controller/stream` | `ControllerStream` :1636 | :3047 | controllerGate passes (no table entry); bounds body :137-139 | SSE (serveSSE via controllerStream, :1631-1639) | no | y L1860 | n | n |

### 1.5 Cross-check summary (computed over the tables above)

| Check | Result | Evidence |
|---|---|---|
| Contract operations (paths under `contract/openapi.yaml` `paths:`) | 97 | script over `contract/openapi.yaml` (first `components:` at the end of paths) |
| Routes served by code: node family 24 x 2 + registry 3 + `GET /nodes` 1 + operation 5 + controller 40 | 97 | `handler.go:286-309` x2, `:315-317`, `:346`, `:351-355`, `api.gen.go:3008-3047` |
| In code, not in contract (excluding SPA and prefix catch-all) | 0 | set difference |
| In contract, not in code | 0 | set difference |
| Operations with `x-admin: true` in contract | 46 (10 node-family ops x 2 = 20; registry 3; `PUT /operation/log/level` 1; controller 22) plus one `x-admin` query parameter (`refresh` on `GET /controller/topology`, `contract/openapi.yaml:1789`) | script over contract |
| Routes gated in code | 46 plus the `refresh` parameter gate | `adminOnly` rows, `adminOnlyBecause` rows, `controllerAdmin` (22), `controllerAdminQuery` |
| Contract `x-admin` vs code gate, per route | equal on all 97 (no row has y/n or n/y) | tables 1.2-1.4 |
| `controllerAdmin` keys vs contract controller `x-admin` ops | identical sets (22 / 22) | script; test `TestControllerAdminMatchesTheContract` named at `handlers_controller.go:48-50` |
| Ungated routes that carry a secret field | `GET /node/status`, `GET /nodes/{node}/status` (`csrf_token`), `GET /controller/status` (`csrf_token`) — see section 4 | `handlers_node.go:138`, `handlers_controller.go:320` |
| Streams | 4 (`/node/stream`, `/nodes/{node}/stream`, `/operation/log/stream`, `/controller/stream`), all without a gate in code and without `x-admin` in the contract | table 1.2 row 24 (`stream`), table 1.3 row 9, table 1.4 `/controller/stream` |

### 1.6 Every response header the code sets (`Header().Set` / `hdr.Set`; grep over `fabricui/*.go`, non-test)

| Header = value | file:line | Condition |
|---|---|---|
| `Content-Type: application/json; charset=utf-8` | `fabricui/problem.go:118` | `writeJSON` (all JSON answers); `WriteHeader(200)` at `:119` |
| `Content-Type: application/json; charset=utf-8` | `fabricui/handlers_surface.go:28` | `GET .../surface` (writes body itself, `WriteHeader(200)` `:29`) |
| `Content-Type: application/json; charset=utf-8` | `fabricui/registry.go:284` | `POST /nodes` 201 (`WriteHeader` `:285`) |
| `Content-Type: application/problem+json` | `fabricui/problem.go:72` | `writeProblem` |
| `X-Problem-Code: <code>` | `fabricui/problem.go:73` | `writeProblem` |
| `Retry-After: 30` | `fabricui/handlers_pairing.go:158` | `openWindow` when `!h.allowEmpty` and the node topology has no device (`handlers_pairing.go:154-160`) |
| `X-Content-Type-Options: nosniff` | `fabricui/spa.go:37` | every SPA request |
| `Cache-Control: <cacheControlFor(path)>` (`public, max-age=31536000, immutable` for `assets/`, else `public, max-age=3600, must-revalidate`) | `fabricui/spa.go:54`, rule `:62-67` | existing non-directory file |
| `Content-Type: text/html; charset=utf-8`, `Cache-Control: no-store` | `fabricui/spa.go:70-71` | `serveIndex` |
| `Content-Type: text/event-stream`, `Cache-Control: no-store`, `X-Accel-Buffering: no` | `fabricui/stream.go:45-47` | every SSE response |

Read but empty (headers): grep over `fabricui internal` (non-test, non-generated) for `Access-Control`, `Content-Security`, `X-Frame`, `Referrer-Policy`, `Set-Cookie`/`SetCookie`, `Strict-Transport`, `Permissions-Policy`, `Cross-Origin`, `CheckOrigin`, `Origin` as a request header: none. `ui/src/shell/index.html` has no `<meta http-equiv>` and no CSP. `Last-Event-ID`: no match in Go or TS.

### 1.7 What `Mount` registers vs `Handler`

| Entry point | Registers | file:line |
|---|---|---|
| `fabricui.Mount(mux, o)` | all of section 1.2-1.4 under `o.Prefix` (default `/api/v1`, `options.go:22`) plus the `<prefix>/` catch-all; no SPA | `fabricui/options.go:307-320`, `handler.go:326-364` |
| `fabricui.Handler(o)` | the same on a private mux, plus `/` -> SPA | `fabricui/options.go:324-333` |
| `New`-style constructor | `fabricui.New` does not exist in `fabricui/*.go` (grep `func New` lists `NewEvents`, `NewNodeSwitch`, `NewMemory*`, `NewMemoryOperations` only) | `fabricui/events.go:205`, `nodeswitch.go:51`, `operations.go:315`, `memory_controller.go:41,199` |


---

## 2. Parse sites

### 2.1 Go (`fabricui/`, `fabricui/gofabric/`, `internal/api/api.gen.go`; non-test)

| # | file:line | What is parsed | From | Bound / limit applied (line) |
|---|---|---|---|---|
| G1 | `fabricui/problem.go:125-129` `decodeJSON` | JSON body into a struct | request body | `MaxBytesReader` 1 MiB `problem.go:126` (const `:33`); `DisallowUnknownFields` `:128`. Callers: `handlers_pairing.go:140` (window), `handlers_exposure.go:133` (via `readUpdates`, called by apply/preview), `registry.go:246` (label), `operations.go:208` (log level) |
| G2 | `fabricui/handlers_node.go:325-326` | JSON `{confirm}` for factory-reset | request body | `MaxBytesReader` `:325`; `Decode` error discarded (`_ =`); no `DisallowUnknownFields`; value compared to `factoryResetConfirmation` `:327` |
| G3 | `fabricui/handlers_node.go:376-377` | JSON `{confirm}` for erase | request body | same as G2 (`:376`), compared to `eraseConfirmation` `:378` |
| G4 | `internal/api/api.gen.go:5836, 5917, 6022, 6181, 6214, 6247, 6280, 6339, 6372, 6618, 6654, 6742, 6924` `runtime.BindJSONBody` | controller request bodies (13 operations, see table 1.4) | request body | `MaxBytesReader` 1 MiB in `controllerGate` `handlers_controller.go:137-139`; behaviour of `runtime.BindJSONBody` itself (unknown fields, trailing data): unclear (oapi-codegen runtime, not in repo) |
| G5 | `internal/api/api.gen.go:1904 (kind), 1917 (id), 1992 (filter), 2005 (value), 2018 (timeout), 2507 (since), 2520 (limit), 2824 (refresh)` `BindQueryParameterWithOptions` | query parameters of the controller routes | query | generated type binding only (`string`/`integer`/`boolean`); value ranges applied in handler bodies, see G9-G12 |
| G6 | `internal/api/api.gen.go:2105-2768` (`BindStyledParameterWithOptions` for `id`), `:2426, :2461` (`ep`, integer), `:2470` (`cluster_hex`, string), `:2586` (`index`, integer) | path parameters of controller routes | path | generated; `id`/`cluster_hex` re-parsed by G8 |
| G7 | `fabricui/handlers_controller.go:131` `strconv.ParseBool(r.URL.Query().Get(q))` | `refresh` on `GET /controller/topology` | query | gate condition `err == nil && on` (`:131`); parse error leaves the gate open |
| G8 | `fabricui/handlers_controller.go:208-218` `parseHex`, `:220-226` `parseHex32`, `:228-234` `parseNodeID` | hex ids (node id 64-bit, cluster/attribute/command/device-type 32-bit) | path (`id` call `:255`; `cluster_hex` `:524`), body (`:691, :695, :784, :788, :935, :949, :963, :1055, :1076, :1528`) | trims space and `0x`; length `> bits/4` rejected `:210`; `ParseUint(t,16,bits)` `:213` |
| G9 | `fabricui/handlers_node.go:441-451` `queryInt` (`strconv.Atoi`) | `since`, `limit` | query (callers `handlers_node.go:564, :569`, `operations.go:247, :251`) | `eventWindow`: `since >= 0` `:406`, `limit` 1..1000 `:413` (const `:401`); `logWindow`: `since >= 0` `operations.go:256`, `limit` 1..2000 `:263` (const `:153`) |
| G10 | `fabricui/handlers_node.go:455-461` `pathEndpoint` (`ParseUint(ep,10,16)`) | endpoint number | path `{ep}` (callers `handlers_node.go:489, :559`) | 16-bit by `ParseUint` bitSize |
| G11 | `fabricui/handlers_node.go:275-279` | fabric index | path `{index}` | `ParseUint(...,10,8)`; rejects 0 and 255 `:276` |
| G12 | `fabricui/handlers_controller.go:1259-1284` `CommissionNode` | `code` (QR payload or manual code), `passcode`, `timeout_seconds`, `terms_and_conditions` | request body | `passcode` 1..99999998 `:1272`; terms version/acks 0..65535 `:1274-1275`; timeout 1..600 `:1279`; `code` only `TrimSpace`d `:1264, :1302`, no length bound seen beyond the 1 MiB body |
| G13 | `fabricui/gofabric/gofabric.go:291-295` `controller.ParsePairingCode(code)` | the QR payload / manual pairing code | request body via host adapter | parser is in go-fabric (`go-fabric@v0.22.2`), not read: unclear |
| G14 | `fabricui/handlers_controller.go:1338-1360` `DiscoverDevices` | `filter`, `value`, `timeout` | query | `filter.Valid()` `:1342`; `value` 0..0xFFFFFFFF `:1347`; `timeout` 1..60 `:1354`; `identifierOf` bit-width checks `fabricui/gofabric/gofabric.go:383-392` (12-bit / 4-bit) |
| G15 | `fabricui/handlers_controller.go:1412-1450` `SetCredential` incl. `hex.DecodeString(*b.DatasetHex)` `:1438` | Wi-Fi `ssid`/`passphrase`, Thread `dataset_hex`, `id` | request body | `id` 1..64 runes `:1421`; `ssid` non-empty `:1426`; `dataset_hex` non-empty `:1435` and hex-valid `:1438-1440`; no length bound on `passphrase`/dataset beyond 1 MiB |
| G16 | `fabricui/gofabric/gofabric.go:88-96` `thread.DecodeDataset(dataset)` | Thread operational dataset bytes (checks a network name) | request body (decoded hex), via `gofabric.ThreadDatasetName` | go-fabric decoder (not read); name must be non-empty `:93` |
| G17 | `fabricui/registry.go:96-107` `nodeLabel` | node label | request body | trim; 1..`MaxNodeLabel` (64, `registry.go:29`) runes `:99-102`; no control characters `:103` |
| G18 | `fabricui/registry.go:217` `r.PathValue("node")`, `fabricui/handler.go:338` `r.PathValue("node")` | node key | path | looked up in maps (`h.keyed` `registry.go:218`, `nodeHandler` `handler.go:338`); the pattern `nodeKeyPattern` `handler.go:153` is applied to configured keys (`checkNodeKey` `:140-150`) not to this path value |
| G19 | `fabricui/operations.go:111-118` `parseLevel` | log level name | request body `level` | exact match against `logLevels` `operations.go:90-96` |
| G20 | `fabricui/handler.go:491-492` | `X-Fabric-UI-Token` request header | header | non-empty; `ConstantTimeCompare` with `h.token`; no length bound |
| G21 | `fabricui/spa.go:35, :38-52` | request URL path used as an `fs.FS` name | URL path | `strings.TrimPrefix(r.URL.Path,"/")`; `sub.Open(path)`; directory or error -> index |
| G22 | `fabricui/options.go:335-340` `validPrefix` | `Options.Prefix` | host config | start `/`, no trailing `/`, no space/`{`/`}` |
| G23 | `fabricui/handler.go:153` `regexp.MustCompile`, `:140-150` | `NodeEntry.Key` values | host config | pattern `^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`, reserved names `:157`, duplicates |
| G24 | `fabricui/handlers_controller.go:880, 1017, 1173, 1193, 1233` `json.Unmarshal` | node ACL list, bindings list, `CurrentFabricIndex`, fabrics list, `NOCResponse` | peer node's attribute/command result handed over by the host's `ControllerSource` | none seen |
| G25 | `fabricui/handlers_node.go:540` `json.Unmarshal(b,&rep)` | the handler's own marshalled surface report | host `Node` | none (self-produced) |
| G26 | `fabricui/node_detail.go:295` `jsonOf` -> `json.Unmarshal(raw, v)`; `fabricui/gofabric/node.go:119, 383, 494` (`client.DecodeJSON` + `json.Unmarshal`) | node attribute / event-list / fabric-index values | node attribute values read by the adapter | none seen |
| G27 | `fabricui/gofabric/gofabric.go:437-445` `discovery.ParseTxtDecimal(vendor/product, 0xFFFF)` on the `VP` TXT record | mDNS TXT `VP` (vendor+product) | LAN mDNS | bound 0xFFFF (`:439, :442`) |
| G28 | `fabricui/memory_controller.go:1020, 1026` | RemoveFabric payload / fabrics of the in-memory test double | test double | none |

### 2.2 TypeScript (`ui/src`, excluding `*.test.ts`, `*.gen.ts`, `ui/src/test`)

| # | file:line | What is parsed | From | Bound / limit applied (line) |
|---|---|---|---|---|
| T1 | `ui/src/lib/api/stream.ts:77, :121, :144` `JSON.parse(String(msg.data))` | SSE event data (node, controller, log stream) | SSE (`EventSource(url,{withCredentials:true})` `:70, :114, :138`) | none seen on size; non-JSON skipped in `try/catch` `:76-80, :120-124, :143-147`; result cast to `StreamEvent`/`ControllerStreamEvent`/`LogRecord` `:82, :126, :148` without a shape check |
| T2 | `ui/src/lib/api/stream.ts:19-37, :88-98, :73, :117` | event type names | SSE | allow-lists `STREAM_EVENTS` / `CONTROLLER_STREAM_EVENTS`; one listener per listed type; log stream listens for `log_record` only `:141` |
| T3 | `ui/src/lib/api/client.ts:304-305` `JSON.parse(text)` | every API response body | HTTP response | none seen; 204 short-circuit `:303`; `isProblem` check `:453-455`; `ApiError` message `p.detail \|\| p.title` `:307-308` |
| T4 | `ui/src/lib/api/client.ts:332, :336, :378` | `csrf_token` from status responses stored in module-local `auth.token` | HTTP response | none; sent as header `client.ts:296` on `admin: true` calls |
| T5 | `ui/src/lib/api/client.ts:413-420, :423, :442-448` `URLSearchParams` | builds `filter/value/timeout`, `kind/id`, `since/limit` queries | page state | builders (no parse); path segments via `encodeURIComponent` `:326, :340, :368, :369, :385, :410, :451` |
| T6 | `ui/src/shell/route.ts:22-44` `routeOf(hash)` | URL fragment: family, page, `?tab=`, controller node id, node key, endpoint number | `location.hash` (reads `ui/src/shell/App.svelte:71, :72`) | node id `/^[0-9a-f]{1,16}$/i` `:28`; endpoint `/^\d+$/` and `<= 0xffff` `:42`; `decodeURIComponent(rest[0])` `:37` with no try/catch seen; `tab` `:27` taken as-is (consumers `Node.svelte:35 isTab`, `Nodes.svelte`) |
| T7 | `ui/src/shell/App.svelte:126, :131, :138`; `ui/src/lib/pages/controller/Nodes.svelte:74` | `location.hash = ...` writes | page state / API data (`n.node_id_hex`) | writes, not parses |
| T8 | `ui/src/lib/context.ts:79, :92-97` | `TopologyRef` (`page`, `id`) -> hash route | API data (topology) | `/^\d+$/` on `ref.id` for `endpoints` `:95`; `page` checked against `REF_PAGES` `:96`; `controller_node` id appended unchecked `:94` |
| T9 | `ui/src/lib/scan/index.ts:11` `isPayload` | text read from the camera frame | camera (`getUserMedia` `:49`; `BarcodeDetector` `:19`; `jsQR` `:40`) | regex `/^MT:[0-9A-Z.-]+$/`, no length bound; poll every 200 ms `:77` |
| T10 | `ui/src/lib/qr.ts:46-57` `qrCreate(payload, {errorCorrectionLevel:"M"})` | `qr_code` payload string turned into an SVG string | API data | none seen (library call) |
| T11 | `ui/src/lib/kit/DataTable.svelte:62-63` `localStorage.getItem(...)` + `JSON.parse` | persisted sort/search `{sortKey,sortAsc,query}` under key `fabric-ui.table:<persistKey>` | `localStorage` | `try/catch` `:61-65`; parsed object spread into state `{...base, ...parsed}` without type check; writes `:79-82`; keys used: `node-fabrics`, `node-selection`, `node-sessions`, `controller-nodes`, `controller-peers`, `controller-sessions` (`Fabrics.svelte:182`, `Selection.svelte:363`, `Diagnostics.svelte:218`, `Nodes.svelte:131`, `controller/Diagnostics.svelte:216, :237`) |
| T12 | `ui/src/lib/kit/utils.ts:34-40` `loadLS`, `:43-49` `saveLS` | `localStorage` string read/write | `localStorage` | `try/catch`; no caller of `loadLS`/`saveLS` in `ui/src/lib` or `ui/src/shell` (grep) |
| T13 | `ui/src/lib/kit/utils.ts:17-30` `makeTextMatcher` -> `new RegExp(q, "i")` | the user's search text as a regular expression | text input | `try/catch` `:21-24`; no length or complexity bound |
| T14 | `ui/src/lib/pages/controller/node/Attributes.svelte:60` `JSON.parse(edit)` | attribute value typed by the user | text input | `try/catch` `:59-64`; none on size |
| T15 | `ui/src/lib/pages/controller/format.ts:45-49` `parseHexId` (`parseInt(t,16)`) | cluster/attribute hex typed by the user | text input | `/^[0-9a-f]{1,8}$/i` `:47` |
| T16 | `ui/src/lib/pages/controller/node/Acl.svelte:48-49, :62, :66` | subjects (`split(/[\s,]+/)`, `/^(0x)?[0-9a-f]{1,16}$/i`), target endpoint `Number(...)` | text input | subject regex `:49`; `Number(targetEndpoint)` unchecked at parse |
| T17 | `ui/src/lib/pages/controller/node/Bindings.svelte:56-59` | binding target (`/^g:\d+$/i`, node id regex), `Number(d.endpoint)`, `Number(target.slice(2))` | text input | node id regex `:57`; numbers unchecked at parse |
| T18 | `ui/src/lib/pages/controller/Commission.svelte:82-84` `network.split(":")`; `:98-101` `Number(v)`, `Number(a)`; `:180-183` `Number(passcode)` | wifi/thread selection value; T&C version/ack; passcode | select / text input | T&C `Number.isInteger` and `0..0xffff` `:100`; passcode `Number.isInteger` and `1..99999998` `:182` |
| T19 | `ui/src/lib/pages/controller/Settings.svelte:53` `secret.replace(/\s+/g,"")` | Thread dataset hex typed/pasted | text input | none client-side (hex validity checked by server G15) |
| T20 | `ui/src/lib/pages/controller/node/Window.svelte:20, :26`; `ui/src/lib/pages/node/Pair.svelte:112` `Number(v)` | window duration | fixed `<select>` options | option list constants (`Window.svelte:20`) |
| T21 | `ui/src/lib/pages/node/Surface.svelte:63, :65, :99`; `ui/src/lib/pages/node/surface.ts:23`; `ui/src/lib/pages/node/Endpoint.svelte:116, :119` `Number(id)` | cluster/device-type id strings from the surface report / endpoint detail | API data | `hexId` requires `startsWith("0x")` and `Number.isFinite` `surface.ts:24`; others none |
| T22 | `ui/src/lib/pages/controller/node/Attributes.svelte:34` `Number(endpoint)`; `Inventory.svelte:77-78` `key.split("/")` + `Number(ep)` | endpoint from a select; internal map key | page state | none |
| T23 | `ui/src/lib/pages/node/fields.ts:21` `field.split(",")` | list-valued exposure item field (deprecated comma form) | API data (host exposure item) | trims, drops empties |
| T24 | `ui/src/lib/i18n/index.ts:19` `s.replace(/\{(\w+)\}/g, ...)`; `:48` `navigator.languages[0]` | translation placeholders; browser language | catalogue string + caller vars; browser | placeholder name `\w+`; locale reduced to `"de"`/`"en"` `:49` |
| T25 | `ui/src/lib/theme/ha-bridge.ts:63, :84-86, :92`; `ui/src/shell/index.html:56-73` | parent-document CSS custom properties and background colour; `parseInt`/`Number` of `#rgb`, `rgb()` | parent frame (`window.parent.document`, `getComputedStyle`) | `parseRgb` accepts `#rgb(a)`, `#rrggbb`, `rgb()/rgba()` only `:78-92`; `NaN` check `:87`; wrapped in `try/catch` |
| T26 | `ui/src/lib/pages/controller/Diagnostics.svelte:106, :112` `Number(bool)` | boolean flags to sort keys | API data | n/a (coercion) |

Read but empty (TS parse): `postMessage`, `addEventListener("message")`, `sessionStorage`, `document.cookie`, `window.name`, `atob`/`btoa`, `new URL`, `BigInt(`, `parseFloat`, `FileReader`, `window.open`, `eval`, `import.meta` in `ui/src/lib` and `ui/src/shell` (grep): no match. `new URL`/`URLSearchParams` appear only as builders (T5) and one reader of `?tab=` (T6).


---

## 3. Render sinks (web) — `ui/src/lib`, `ui/src/shell`, no tests

### 3.1 Sink inventory

| Sink class | file:line | Value and where it comes from | Notes (as read) |
|---|---|---|---|
| `{@html}` | `ui/src/lib/pages/node/Pair.svelte:150` | `qrSvg` = `renderQrSvg(win.qr_code)` (`Pair.svelte:91`); `win` is the `Window`/`Pairing` object from the API (`store.svelte.ts:246-247, :270-274`) | `renderQrSvg` `ui/src/lib/qr.ts:46-57` builds the string from `px` (default 200), `total` (module count + 8), and a path made of `M`/`m`/`h` commands over the module matrix of `qrCreate(payload,…)`; the payload text is not interpolated into the string (`qr.ts:51-57`) |
| `{@html}` | `ui/src/lib/pages/controller/node/Window.svelte:66` | `qrSvg` = `renderQrSvg(win.qr_code)` (`Window.svelte:21`), `win` from `client.openNodeWindow` `:26` | same function |
| `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `document.write`, `srcdoc`, `eval`, `new Function`, `setTimeout("string")`, `createElement("script")`, `window.open`, `use:` action, `@attach`, `<iframe>/<object>/<embed>/<img>`, `src=`, `action=` | none | grep over `ui/src/lib`, `ui/src/shell` (`*.ts`, `*.svelte`, non-test) | only `document.createElement("canvas")` at `ui/src/lib/scan/index.ts:54` |
| `window.location.assign(href)` | `ui/src/lib/kit/Graph.svelte:29` (default of prop `follow`), used `:145` on Enter | `href` = `hrefOf(node)`; `hrefOf` = `refHref(n.ref, routes)` (`ui/src/lib/pages/node/Topology.svelte:51`; `ui/src/lib/pages/controller/Network.svelte:78`, peers only) | `refHref` `ui/src/lib/context.ts:92-97` returns `routes.<x> + id` or a `routes` entry (hash routes); `follow` is passed down but no shell file supplies one (grep `follow` in `ui/src/shell`: none), so the default applies |
| `location.hash =` | `ui/src/shell/App.svelte:126, :131, :138`; `ui/src/lib/pages/controller/Nodes.svelte:74` | `routesForNode(...)`, `routes.controllerNode + n.node_id_hex` | hash writes |
| `target="_blank"` with `rel="noopener noreferrer"` | `ui/src/lib/pages/node/EcosystemHints.svelte:48` (href `link.href` from the static table `ui/src/lib/pages/node/ecosystems.ts:15-32`, https constants); `ui/src/lib/pages/node/NodeManage.svelte:106` (href const `HINT_SOURCE` `:22-23`) | static https URLs | no `target="_blank"` without `rel` found |
| dynamic `href=` (hash routes + data) | `ui/src/lib/pages/node/Endpoint.svelte:153` (`parent_endpoint_id`), `:159` (`parts` entries); `ui/src/lib/pages/node/Endpoints.svelte:96` (`s.number`), `:111` (`ep.endpoint_id`); `ui/src/lib/pages/controller/Nodes.svelte:161` (`n.node_id_hex` + `?tab=window`); `ui/src/lib/pages/controller/Diagnostics.svelte:132` (`s.peer_node_id_hex`), `:148` (`p.node_id_hex`); `ui/src/lib/pages/controller/Commission.svelte:262` (`result.node.node_id_hex`); `ui/src/lib/kit/Graph.svelte:266` (`{href}` from `hrefOf`); `ui/src/lib/kit/Tabs.svelte:53` (`item.href`) | `routes.<x>` string prefix (starts with `#/`) + a number or hex id from the API | tab hrefs in `App.svelte:94-96, :142-148`, `NodeFamily.svelte:19`, `Nodes.svelte:40` are `routes.*` values |
| `style=` / `style:` with data | `ui/src/lib/pages/controller/node/Endpoints.svelte:32` (`margin-left: ${depth*1.5}rem`, `depth` from the endpoint tree); `ui/src/lib/kit/Graph.svelte:176` (`style:height` from `layout.height`); `ui/src/lib/kit/Spinner.svelte:11` (`--fui-size:{size}px`, prop) | numbers | `Pair.svelte:141` is a static style string |
| SVG attributes from data | `ui/src/lib/kit/Graph.svelte:177` (`viewBox` from `view.*`), `:190` (`d={e.path}`), `:193-195, :225-227` (stroke from constant maps), `:191-192, :207-208` (`data-edge-kind`, `data-strength`, `data-id={n.id}`, `data-state={n.state}`) | layout numbers; `n.id`, `n.state`, `edge.kind/strength` from the topology response | |
| CSS custom property injection | `ui/src/lib/theme/ha-bridge.ts:120-121` (`ourRoot.style.setProperty(name, value)` / `removeProperty`), `:130` (`style.colorScheme`); `ui/src/shell/index.html:62` (`root.style.setProperty(vars[i], value)`), `:80` (`root.style.colorScheme`) | `name` from the constant list `HA_THEME_VARS` (`ha-bridge.ts:11`) / `vars` (`index.html:21-53`); `value` = `getComputedStyle(parentDoc.documentElement).getPropertyValue(name).trim()` of the parent frame (`ha-bridge.ts:115-119`, `index.html:59-61`) | applied only when `isEmbedded()` (`ha-bridge.ts:54-67`) or, in the prepaint script, when framed and the path contains `/api/hassio_ingress/` or the parent has `<home-assistant>` (`index.html:55-58`) |
| class / dataset toggles | `ui/src/lib/theme/ha-bridge.ts:129` (`classList.toggle("dark")`), `:156-157` (`dataset.skin`); `ui/src/shell/App.svelte:176`; `ui/src/shell/index.html:19, :87` | booleans / constant `"ha"` | |
| class strings from data | `ui/src/lib/pages/node/Diagnostics.svelte:132-138, :273` and `ui/src/lib/pages/controller/Diagnostics.svelte:257` (`eventClass(event.severity)`, ternary over constant class strings); `ui/src/lib/kit/DataTable.svelte:182, :201, :208` (`col.headClass`, `rowClass(row)`, `col.cellClass` from the page's column definition) | `severity` from API; class strings from code | |
| DOM-touching code | `ui/src/lib/scan/index.ts:50-51` (`video.srcObject = stream`, `setAttribute("playsinline","true")`), `:56-62`; `ui/src/lib/theme/ha-bridge.ts:138-139` (`MutationObserver` on the parent `documentElement`), `:141-142` (`focus`, `visibilitychange` listeners) | camera stream; parent frame | |
| clipboard writes | `ui/src/lib/pages/node/Pair.svelte:59`; `ui/src/lib/pages/controller/node/Window.svelte:49`; `ui/src/lib/pages/operation/Operation.svelte:102` | `win.manual_code`/`win.qr_code`; log lines (`recordLine`) | |
| default toast / confirm fallbacks | `ui/src/lib/context.ts:127-131` (`console.info/error(m)`), `:142` (`window.confirm(title\n\nbody)`) | message strings | shell passes its own toast/confirm (`App.svelte:57`) |

### 3.2 `ui/src/shell/index.html` prepaint script and `ha-bridge.ts`: what they read from the parent frame

| Fact | file:line |
|---|---|
| Inline `<script>` (IIFE) runs before first paint; sets `data-skin="ha"` on `<html>` | `ui/src/shell/index.html:14-19` |
| Reads `prefers-color-scheme` | `ui/src/shell/index.html:20` |
| When `window.self !== window.top` (framed): reads `window.parent.document`; proceeds if `location.pathname` contains `/api/hassio_ingress/` or `parentDoc.querySelector("home-assistant")` | `ui/src/shell/index.html:55-58` |
| Reads `parentWin.getComputedStyle(parentDoc.documentElement)`; for each of 31 listed names copies a non-empty value to `root.style.setProperty` | `ui/src/shell/index.html:21-53` (list), `:59-63` (copy) |
| Reads `--primary-background-color` (else `ps.backgroundColor`), parses `#rgb`/`#rrggbb` (`parseInt(...,16)`) or `rgb(a)()` (`Number`), computes luminance, sets `dark` class and `colorScheme` | `ui/src/shell/index.html:64-81, :87` |
| All wrapped in `try/catch` (cross-origin parent leaves `prefers-color-scheme`) | `ui/src/shell/index.html:54, :84-86` |
| `isEmbedded()`: `window.self === window.top` -> false; true when path contains `HA_INGRESS_SEGMENT`; else `window.parent.document.querySelector("home-assistant") !== null` in `try/catch` | `ui/src/lib/theme/ha-bridge.ts:54-67` (segment const `:46`) |
| `startHaBridge()`: reads `window.parent` computed style of its `documentElement` for each `HA_THEME_VARS` name (copy or remove), reads `--primary-background-color`/`backgroundColor`, `parseRgb`, toggles `dark`; `MutationObserver` on parent root (`style`,`class` attributes) and `focus`/`visibilitychange` re-apply | `ui/src/lib/theme/ha-bridge.ts:106-149` (copy `:115-121`, observer `:138-139`) |
| `parseRgb`: `#rgb`/`#rrggbb` hex via `parseInt`; `rgb()/rgba()` via regex `/rgba?\(\s*([\d.]+)[\s,]+([\d.]+)[\s,]+([\d.]+)/i` and `Number` | `ui/src/lib/theme/ha-bridge.ts:77-94` |
| `postMessage`, `message` listeners, `window.name`, cookies: none | grep, none |

### 3.3 Strings from a node, the host, a log record, discovery, or a problem detail — where they are rendered

Svelte text interpolation `{expr}` is used for every row marked "text"; "attr" = element attribute, "title" = `title=` attribute, "aria" = `aria-label`. No row reaches `{@html}` (only the two QR sites in 3.1).

#### Kit components (`ui/src/lib/kit/`)

| file:line | Field(s) | Origin | Sink |
|---|---|---|---|
| `AttributeTable.svelte:31-36, :44-46` | `a.id_hex`, `a.name`, `format(a.value)`, `imStatusName(a.status)`, `statusHex` | node attribute (controller read / endpoint cluster) | text |
| `ClusterView.svelte:23-28, :36, :48` | `cluster.id_hex`, `cluster.name`, `feature_map_hex`, feature `f.code`/`f.title`, event `e.id_hex`/`e.name` | node cluster description | text; `:48` also `title={e.id_hex}` |
| `EventList.svelte:28-40` | `e.number`, `priority`, `cluster_hex`, `event_hex`, `name`, `time(e.epoch_ms)`, `format(e.data)` | node event log | text; `datetime={...toISOString()}` `:36` |
| `TagChips.svelte:22` | `tag.namespace`, `tag.tag`, `title(tag)` | node endpoint tags | text + title |
| `Graph.svelte:174, :197, :211, :235-237, :250, :254-255, :260-261, :266` | graph `label` prop; edge title; `nodeAria(n)` (`:159`, uses `n.label`); `ln.text`/`n.label`; `selected.label`; `d.key`/`d.value` details; `href` | topology response (node or controller) | aria, SVG `<title>`, SVG text, text, href |
| `DataTable.svelte:182, :189, :193, :201, :208-211, :222` | `col.label`, `col.get?.(row)`, `group.key`, `data-label={col.label}`, `aria-sort`, `rowClass`, `cellClass`, `headClass` | page column definitions + row data (any page below) | text, attr, class |
| `ErrorState.svelte:17`; `EmptyState.svelte:16, :18`; `LoadingState.svelte:14`; `PageHeader.svelte:21, :23`; `Badge.svelte:24`; `Icon.svelte:12`; `Select.svelte:34, :44, :47`; `Tabs.svelte:49, :53-62` | `message`, `description`, `label`, `title`, `subtitle`, `placeholder`, `option.label`, `item.label/badge/href/icon` | props from pages: error text (problem detail via `errorText`), node/host strings (e.g. node-selector labels from `GET /nodes`, `App.svelte:221-223`) | text, title, aria, href |

#### Node pages (`ui/src/lib/pages/node/`)

| file:line | Field(s) | Origin | Sink |
|---|---|---|---|
| `HostWarnings.svelte:19-21` | `w.code`, `w.message`, `w.since` (via `toLocaleString`, in `t()` var) | host `Options.Warnings` via `Status.warnings` | text |
| `HostCounters.svelte:23-25` | `c.code`, `c.value`, `c.message` | host `Options.Counters` via `Status.counters` | text |
| `Status.svelte:72, :82, :118-130` | `s.attestation_error` (`handlers_node.go:175`, Go `err.Error()`), `s.health_note` (title), `listen_addr`, `boot_reason` (i18n key built from it `:~130`), surface `profiles.join`/`extensions.join` inside `t()` vars | node status response | text; title (`:82`) |
| `NodeConfig.svelte:35-44` | `config.label`, `listen`, `discriminator`, `vendor_name`, `product_name`, `store`, `source_note` (via `note`) | host `Options.NodeConfig` via `Status.config` | text |
| `NodeManage.svelte:50, :79, :95` | node `label` in `t()` toast vars; input `value={label}` | node registry entry / operator input | toast text, input value |
| `Pair.svelte:157, :167, :183` (+ `:150` html) | `win.manual_code`, `win.qr_code`, `addedLabel` (`node.commissioning.addedFabricLabel`, `:47`) | pairing window / fabric label from SSE | text |
| `Fabrics.svelte:189-198` | `vendorName(fabric)`, `fabric.label`, `fabric_index`, `nodeIdHex(fabric)` | node fabrics | text (`DataTable` cells) |
| `Endpoints.svelte:97, :111, :119, :129` | `s.label`, `ep.name`, `device_type_name`, `cluster.name` | node endpoints | text; hrefs `:96, :111` |
| `Endpoint.svelte:139-211` | `d.endpoint_id`, `d.name`, `dt.name`/`dt.id_hex` (text + `title={dt.id_hex}` `:142`), `d.scope` (in `t()`), `l.label`/`l.value` (`:168`), `d.tags`, `c.name`/`c.id_hex`, `dec.policy`, `dec.reason`, `element(dec)` | node endpoint detail / surface decisions | text, title |
| `Selection.svelte:246-263, :271, :278, :288, :324-333, :388, :411, :417, :423` | `current(item).name`, `verdict.device_type_name`, column values `col.get?.(item)`, `node.status.exposure_reason`, `warnings` (HostWarnings), filter `f.label`, `after.name`, `verdict.reason`, `drawer.name`, `item.source` | host exposure items (`Options.Exposure` list), host `ExposureReason` | text; `:243` aria/title = `si.label` (a `t()` catalogue string, `Selection.svelte:114-121`) |
| `Surface.svelte:123-157, :171-176, :203, :249-250, :267-279` | `clusterLabel`/`deviceTypeLabel`, `d.verdict`, `d.policy`, `r.policies` (title `:203`), `endpointLabel`, `element(...)`, `c.before`/`c.after` (+ `shortDigest`), `ch.before`/`ch.after` | node surface report | text; title `:203` |
| `Diagnostics.svelte:142-149, :185-186, :193-195, :229-230, :242-252, :272-274` | session fields; mDNS `finding.severity/message`; `svc.service_type`, `port`, `addresses`, `subtypes`; ecosystem `eco.ecosystem`, `finding.message`; events `event.at/kind/message` | node diagnostics (sessions, mDNS, ecosystem findings, trace) | text; class `:273` |
| `EcosystemHints.svelte:48, :56` | `finding.message`; `link.href` (static) | node diagnostics / static table | text; href |
| `Topology.svelte` (via `Graph`) | node `label`, `details` | node topology | see Graph row |

#### Operation page

| file:line | Field(s) | Origin | Sink |
|---|---|---|---|
| `operation/Operation.svelte:128` | `info.started_at` | host `Operations.Info()` | title attribute |
| `operation/Operation.svelte:157-161` | `rec.time`, `rec.level`, `rec.msg`, each `rec.attrs` `{k}={v}` | log record (REST `GET /operation/log` and SSE `log_record`) | text |
| `operation/Operation.svelte:102`; `operation/log.ts:20-23` | `recordLine(r)` (time, level, `msg`, `k=v`) | log record | clipboard text; filter haystack |

#### Controller pages (`ui/src/lib/pages/controller/`)

| file:line | Field(s) | Origin | Sink |
|---|---|---|---|
| `Commission.svelte:262, :265, :271, :278-280, :299-305, :314` | `nodeName(result.node)`, `failure` (error text), `ble.reason`, `step.number/sub/name/phase`, discovery `d.device_name`/`instance_name`, `vendorLabel`, `long_discriminator`, `addresses.join`, passcode input `value` | commission response; SSE `commissioning_progress`; mDNS discovery results (`GET /controller/discovery`) | text, href (`:262`), input value |
| `Nodes.svelte:141-161` | `nodeName(n)`, `vendorLabel`, `connectionLabel`, `n.addresses.join(", ")`, `n.node_id_hex` | controller node list | text, href |
| `Node.svelte:71` | `nodeName(detail.node)` | controller node | text |
| `Diagnostics.svelte:129-163, :200-201, :256-258` | session/peer fields, `p.load_error`, `operational_address/host`, finding `severity/message`, event `kind/message` | controller diagnostics (sessions, peers, mDNS findings, trace) | text; class `:257`; href `:132, :148` |
| `Settings.svelte:115-122, :149-152` | credential `c.kind`, `c.id`, `c.name` (SSID / Thread network name), status `compressed_fabric_id_hex`, `vendorLabel`, `node_count` | controller credentials / status | text |
| `Network.svelte:95-106` | graph; `error` | controller topology | via Graph; text |
| `node/Acl.svelte:121-122`; `node/Bindings.svelte:99`; `node/Endpoints.svelte:35-54`; `node/Fabrics.svelte:68-72`; `node/Icd.svelte:74`; `node/Inventory.svelte:105-162` (names, tags, labels, `ep.conformance.status`, `v.device_type/requirement/detail`, cluster/feature names); `node/Overview.svelte:66-117` (vendor, `product_name`, `software_version`, `serial_number`, `addresses`, `operational_host`, `ping.error`, per-address `a.ip`/`a.error`, features); `node/Updates.svelte:61-62`; `node/Attributes.svelte:92` (`JSON.stringify(result.value, null, 2)`); `node/Window.svelte:69-73` (`win.manual_code`, `win.qr_code`, `win.discriminator`; html `:66`) | controller node data read from the peer node through the host | text (Attributes: text in a block) |

#### Shell

| file:line | Field(s) | Origin | Sink |
|---|---|---|---|
| `shell/Toaster.svelte:16` | `e.message` | `ApiError.message` = problem `detail` or `title` (`client.ts:307-308`), `errorText` (`store.svelte.ts:419-421`), or `t()` strings; also SSE-derived (`controller-store.svelte.ts:182-183`: `update_version`, `update_error` inside `t()` toast) | text |
| `shell/ConfirmDialog.svelte:21, :23, :26, :28` | `current.title`, `current.body`, labels | callers: `c.name` (`Settings.svelte:69`), `errorText(err)` (`controller/node/Icd.svelte:41`), node label (`NodeManage.svelte:62-69`, `Fabrics.svelte`) | text |
| `shell/NodeCreateDialog.svelte:50` | `error` | `errorText` | text |
| `shell/App.svelte:221-223, :249` | node-selector options (labels from `GET /nodes`), `nodeKey` | host/registry | `<option>` text, text |
| `lib/store.svelte.ts:314`; `lib/controller-store.svelte.ts:115, :182-183` | `fabric_index`; `message(err)`; `update_version`/`update_error` | SSE/API | toast text |

Read but empty (render sinks): `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `document.write`, `srcdoc`, `window.open`, `eval(`, `new Function`, string-argument `setTimeout/setInterval`, `createElement("script")`, `use:`, `<svelte:head>`, `<iframe>`, `<img>`, `src=`, `action=`, `target="_blank"` without `rel`.


---

## 4. Secrets in the data

Gate column refers to section 1 (tables 1.2-1.4). "Token mode" = `Options.Admin == nil`.

### 4.1 Fields

| # | Secret field | Go (struct / set) | Contract schema | TS type / use | Carried by (route or event) | Gate of that route/stream |
|---|---|---|---|---|---|---|
| S1 | admin/CSRF token `csrf_token` | `statusResponse.CSRFToken` `fabricui/handlers_node.go:49`, set `:138`; `ControllerStatus.CsrfToken` `internal/api/api.gen.go:959`, set `fabricui/handlers_controller.go:320`; generated `handler.go:265-269`, copied to sub-handlers `handler.go:402` | `Status.csrf_token` `contract/openapi.yaml:1974`; `ControllerStatus.csrf_token` `:2706` | `ui/src/lib/api/schema.gen.ts:2044, :2766`; stored `client.ts:332, :336, :378` (`auth.token`), sent as `X-Fabric-UI-Token` `client.ts:296`; UI reads presence for `isAdmin` (`shell/App.svelte:107`, `pages/node/NodeManage.svelte:27`, `pages/node/Diagnostics.svelte:26`) | `GET /node/status`, `GET /nodes/{node}/status` (1.2 rows 1), `GET /controller/status` (1.4 row 1) | none (plain / open); the token exists only when `h.admin == nil` (`handler.go:265`), the field is `omitempty`, empty otherwise |
| S2 | passcode + discriminator + QR code + manual code of the node's pairing | `pairingResponse` `fabricui/handlers_pairing.go:38-46` (`Discriminator` `:39`, `Passcode` `:40`, `QRCode` `:41`, `ManualCode` `:42`); built `:69-72` (window opened by the handler) and `:83-86` (configured setup payload) | `Pairing` `contract/openapi.yaml:2042-2049` (`passcode` `:2047`, `qr_code` `:2048`, `manual_code` `:2049`) | `schema.gen.ts:2118-2123`; hydrate `ui/src/lib/store.svelte.ts:262-274`; rendered `pages/node/Pair.svelte:157, :167` (+ QR `:150`); passcode not rendered (grep `passcode` in `pages/node/*.svelte`: none) | `GET /node/pairing`, `GET /nodes/{node}/pairing` (1.2 row 4) | `adminOnly` (x-admin `contract/openapi.yaml:86`) |
| S3 | passcode + discriminator + QR + manual code of a window the handler opens | `windowResponse` `fabricui/handlers_pairing.go:118-125`; built `:214-215`, `:236-237` | `Window` `contract/openapi.yaml:2057-2065` (`passcode` `:2062`) | `schema.gen.ts:2134-2140`; `store.svelte.ts:245-248` keeps `win`; rendered as S2 | `POST /node/pairing/window`, `POST /nodes/{node}/pairing/window` (1.2 row 5) | `adminOnly` (x-admin `:100`) |
| S4 | passcode + discriminator + QR + manual code of a controller node's window | `api.NodeWindow` fields `internal/api/api.gen.go:1323-1326`; built `fabricui/handlers_controller.go:667-668`; port struct `fabricui/controller.go:627-629`; adapter `fabricui/gofabric/node.go:659` | `NodeWindow` `contract/openapi.yaml:3030-3037` (`passcode` `:3036`) | `schema.gen.ts:3056-3059`; `pages/controller/node/Window.svelte:3` (comment: passcode "is not stored"), rendered `:66, :69, :71, :73` (QR svg, manual code, QR text, discriminator; passcode not rendered) | `POST /controller/nodes/{id}/window` (1.4 row 13) | `controllerAdmin` L57 (x-admin `contract/openapi.yaml:1320`) |
| S5 | passcode, QR/manual `code`, `discovered.long_discriminator` in the commission request | `api` `CommissionRequest` `api.gen.go:801-806`; read `fabricui/handlers_controller.go:1259-1308`; port `CommissionDiscovered(..., passcode uint32, ...)` `fabricui/controller.go:42`; adapter `fabricui/gofabric/gofabric.go:291-314, :480-513` | `CommissionRequest` `contract/openapi.yaml:3199-3207` (`code` `:3202`, `passcode` `:3205`), `DiscoveredRef` `:3193-3198` | `schema.gen.ts:3199-3204`; `pages/controller/Commission.svelte:56` (`passcodes` state), `:180-190`, `:214` (code input), `:216` (scan fills `code`) | request body of `POST /controller/commission` (1.4 row 27) | `controllerAdmin` L66 (x-admin `contract/openapi.yaml:1596`) |
| S6 | Wi-Fi `passphrase`, Thread `dataset_hex` (write-only) | `api.CredentialRequest` `api.gen.go:995-1001`; read `fabricui/handlers_controller.go:1412-1450`; port `SetWiFi`/`SetThread` `fabricui/controller.go:782-785`; getters returning the secret to the commissioner `fabricui/gofabric/gofabric.go:53-54`; stored in memory by `fabricui/memory_controller.go:30-35, :69-95, :118-130` (test double) | `CredentialRequest` `contract/openapi.yaml:3257-3265` (`passphrase` `:3264`, `dataset_hex` `:3265`) | `schema.gen.ts:3259-3261`; `pages/controller/Settings.svelte:25` (`secret`), `:51, :53`, password inputs `:134, :136`, cleared `:59` | request body of `PUT /controller/credentials` (1.4 row 30) | `controllerAdmin` L68 (x-admin `contract/openapi.yaml:1667`) |
| S7 | credential names (SSID of a Wi-Fi credential, network name of a Thread dataset) — no secret field | `Credential` list built in `handlers_controller.go:1390-1410` | `Credential` `contract/openapi.yaml:3242-3250`; summary "never a secret" `:1655` | `Settings.svelte:115-122` | `GET /controller/credentials` | none (open) |
| S8 | `discriminator` alone (window event) | `WindowEvent` `fabricui/events.go:72-75`, built `:335` | `WindowEvent` `contract/openapi.yaml:2530-2535`; stream description "No payload carries a passcode" `:471` | `schema.gen.ts:2595` | SSE events `window_opened`, `window_closed` on `/node/stream`, `/nodes/{node}/stream` | none |
| S9 | `discriminator` (node config, no passcode on that route) | `NodeConfigView.Discriminator` `fabricui/options.go:162`; `statusResponse` config `fabricui/handlers_node.go:88, :146` | `NodeConfig` `contract/openapi.yaml:2322-2338`; "never carries the passcode" `:2327` | `pages/node/NodeConfig.svelte:23` (`String(config.discriminator)`) | `GET /node/status` (`config`), `GET /nodes/{node}/status` | none |
| S10 | `discriminator` of a controller node record (commissionable advert) | `fabricui/handlers_controller.go:371` | `NodeRecord` `contract/openapi.yaml:2784-2797` (`:2797`) | `schema.gen.ts:2866` | `GET /controller/nodes/{id}` | none |
| S11 | in-memory test double constants: passcode `20202021`, manual `34970112332`, QR `MT:Y.K9042C00KA0648G00`, seeded Wi-Fi `"secret-passphrase"` | `fabricui/memory_controller.go:137-139, :211, :412-415, :1060` | n/a | n/a | served only by a host using `NewMemoryController` | as per route |
| S12 | PAKE verifier, Thread network key as a separate field, setup PIN other than above | no field with these names in Go, contract or TS (grep `verifier`, `Verifier`, `PAKE`, `network_key`, `NetworkKey`: none) | | | | |
| S13 | error text echoed in problem details on the commission/credential path | `controllerError` `fabricui/handlers_controller.go:165-195` writes `err.Error()` for `ErrInvalidArgument`/`ErrInvalidCredential` (`:186`), `ErrNodeFailed`/deadline (`:192-193`); adapter wraps go-fabric errors `fabricui/gofabric/gofabric.go:362-376` (`%w: %w` with the underlying error) | | `ApiError.message` `client.ts:307-308` -> toast `shell/Toaster.svelte:16` | responses of `POST /controller/commission`, `PUT /controller/credentials` | gated routes; whether the underlying go-fabric error text can contain a code or dataset: unclear (go-fabric not read) |

### 4.2 Log calls (`slog.*`, `log.*`, `Logger`, `console.*`) and whether their arguments contain a field above or a whole request/record

Doc statements: `Options.Logger` "Nothing a handler logs carries a passcode" `fabricui/options.go:254-256`; contract `contract/openapi.yaml:1058, :1322`; `fabricui/handlers_controller.go:1255-1257` and `:1410-1411`.

| file:line | Message | Arguments | Contains a secret field or whole request/record? |
|---|---|---|---|
| `fabricui/handlers_pairing.go:79` | `fabricui.pairing.codes_unavailable` | `error` text | error text only (source error not read: unclear) |
| `fabricui/handlers_pairing.go:180` | `fabricui.window.blocked` | `error` text | as above |
| `fabricui/handlers_pairing.go:197-199` | `fabricui.window.opened` | `discriminator` (int), `duration_seconds`, `configured` (bool) | logs the window's discriminator; passcode, QR and manual code not passed |
| `fabricui/handlers_diagnostics.go:178` | `fabricui.mdns.reannounced` | none | no |
| `fabricui/handlers_exposure.go:214` | `fabricui.exposure.reassemble_failed` | `error` text | no secret field; error from host `Reassemble` |
| `fabricui/handlers_exposure.go:219` | `fabricui.exposure.applied` | `applied` count | no |
| `fabricui/handlers_node.go:285` | `fabricui.fabric.removed` | `fabric_index` | no |
| `fabricui/handlers_node.go:343-344` | `fabricui.factory_reset.partial` | `removed`, `of`, `fabric_indexes` | no |
| `fabricui/handlers_node.go:350` | `fabricui.factory_reset.done` | `removed` | no |
| `fabricui/handlers_node.go:388` | `fabricui.erase.done` | none | no |
| `fabricui/operations.go:222` | `fabricui.operation.log_level` | `level` name | no |
| `fabricui/problem.go:104-105` (`serverError`, used by many handlers) | the call-site title | `error` text, `method`, `path` (`r.URL.Path`, no query) | error text of any failing port call; request body not logged |
| `fabricui/handlers_controller.go:192` | `fabricui.controller.node_failed` | `path` (`r.URL.Path`), `error` text | error text of a node/commission failure (see S13: unclear whether it can carry a code) |
| `fabricui/stream.go:67` | `fabricui.stream.encode_failed` | `type`, `error` | no |
| `fabricui/registry.go:147` | `fabricui.registry.entry_skipped` | `error` | no |
| `fabricui/registry.go:282, :321` | `fabricui.registry.created` / `deleted` | `key` | no |
| `fabricui/gofabric/gofabric.go:307` | `fabricui.pair.ble_skipped` | `reason` | no |
| `fabricui/gofabric/gofabric.go:581` | `fabricui.gofabric.label_push_failed` | `error` | no |
| `ui/src/lib/context.ts:128-130` | default toast | `console.info(m)`, `console.error(m)` with the toast message | message text (problem detail / `t()` strings); used only when the host passes no toast (shell passes one, `App.svelte:57`) |

Read but empty (logs): `log.Print*`, `log.Fatal*`, `fmt.Print*` to stdout/stderr, `os.Std*` in `fabricui` non-test: none (the only `fmt.Fprint*` calls write to the HTTP response, `stream.go:49, :61, :70`). `console.*` in `ui/src` outside `context.ts:128-130`: none (non-test).


---

## 5. The hosts' Options call sites

Exports: `.../scratchpad/pins/go-fabric-server@d61a545` and `.../pins/go-ha2fabric@ec7deea` (prefixes below as `go-fabric-server@d61a545/` and `go-ha2fabric@ec7deea/`).

### 5.1 go-fabric-server@d61a545

| Fact | file:line |
|---|---|
| go.mod pin | `go-fabric-server@d61a545/go.mod:7` `github.com/SukramJ/go-fabric-ui v0.9.2`; no `replace` |
| The only construction of `fabricui.Options` and the only call of `fabricui.Handler` (no `fabricui.New`, no `fabricui.Mount`) | `go-fabric-server@d61a545/internal/daemon/web.go:36-39`, `:58` (grep for `fabricui.Options`, `fabricui.Mount`, `fabricui.Handler`, `fabricui.New(` over non-test `*.go`) |
| Fields set at construction: `Events: ev`, `Logger: log`, `Controller: ctrl`, `Operations: ops`, `ExposureReason: exposureReason` | `go-fabric-server@d61a545/internal/daemon/web.go:36-39`; `exposureReason` defined `internal/daemon/operation.go:17-` (const text not read: unclear); `ev := fabricui.NewEvents()` `internal/daemon/run.go:392` |
| Set when a bridge node exists (`def != nil`): `Node: def.sw` (a `NodeSwitch`, `fabricui.NewNodeSwitch(run.node)` `internal/daemon/nodes.go:209`), `NodeConfig: defaultNodeConfig(def.spec)`, `Erase: life.eraser(def)`, `Registry: reg` (when `reg != nil`) | `go-fabric-server@d61a545/internal/daemon/web.go:40-49`; `defaultNodeConfig` `:77-89` |
| `Warnings: life.warnings` when `life != nil` | `go-fabric-server@d61a545/internal/daemon/web.go:50-52` |
| `Admin: admin` from `webAdmin(c)` | `go-fabric-server@d61a545/internal/daemon/web.go:53-57` |
| Not set: `Prefix` (default `/api/v1`), `Exposure`, `Nodes`, `Reassemble`, `Counters`, `ConfiguredWindow`, `AllowEmptyTopology` | absent from `web.go:36-57` |
| `Admin` func logic (`webAdmin`): `--web-admin-source` set -> `func(r) bool { return fromAddr(r, src) }` (request `RemoteAddr` equals the configured IP, `netip.ParseAddrPort(r.RemoteAddr)`, `:110-116`); else `--web-admin` -> `nil` (the handler's token mode); else `func(*http.Request) bool { return false }` (read-only) | `go-fabric-server@d61a545/internal/daemon/web.go:91-107` |
| Mount: the handler is served at the root of its own `http.Server{Handler: h, ReadHeaderTimeout: 10s}`; `h` = `fabricui.Handler(o)` (API under `/api/v1` plus SPA at `/`); no mux, no path prefix, no other middleware | `go-fabric-server@d61a545/internal/daemon/web.go:58, :128` |
| Wrapper when `--web-admin-source` is set: every request whose peer address differs from the source gets `403 "the pages answer the Ingress proxy only"`; `RemoteAddr` only (no `X-Forwarded-For` read) | `go-fabric-server@d61a545/internal/daemon/web.go:59-72`, `fromAddr` `:110-116` |
| Listen address and TLS: `net.ListenConfig.Listen(ctx, "tcp", c.Web)`; `srv.Serve(ln)` (plain HTTP); `ServeTLS`, `ListenAndServeTLS`, `tls.` : no match in `internal/daemon` | `go-fabric-server@d61a545/internal/daemon/web.go:124-133` |
| Flags: `--web` (string, default `""` = pages off), `--web-admin` (bool, default false), `--web-admin-source` (string); env `FABRIC_SERVER_<FLAG>` | `go-fabric-server@d61a545/internal/daemon/run.go:174-176`, env naming `:132` |
| Call chain: `if c.Web != "" { serveWeb(ctx, c, life, def, reg, ctrlPages, ev, ops, log) }` | `go-fabric-server@d61a545/internal/daemon/run.go:393-398`; `ctrlPages` `:252, :352` |
| Add-on wiring: `FABRIC_SERVER_WEB="0.0.0.0:8099"`, `FABRIC_SERVER_WEB_ADMIN_SOURCE="172.30.32.2"` | `go-fabric-server@d61a545/script/run.sh:56-57` |
| Add-on manifest: `host_network: true`, `ingress: true`, `ingress_port: 8099`; comment "the pages ... listen on the host too; the daemon answers only the Ingress proxy (172.30.32.2)" | `go-fabric-server@d61a545/addon/config.yaml:33, :44-45`; comment `:27-32` |
| Auth in front, as read: the source-address filter of `--web-admin-source` (Ingress proxy address) or none; no CORS, no cookie, no TLS in `internal/daemon` (grep `tls|TLS|Cors|CORS|Access-Control` found only an unrelated `Authorization: Bearer` at `internal/daemon/run.go:585`) | `go-fabric-server@d61a545/internal/daemon/web.go`, `run.go:585` |
| Without either flag: `Options.Admin` returns false for every request (read-only pages); with `--web-admin` and no source: token mode on a listener at `c.Web` with no filter | `go-fabric-server@d61a545/internal/daemon/web.go:103-106, :59-61` |

### 5.2 go-ha2fabric@ec7deea

| Fact | file:line |
|---|---|
| go.mod pin | `go-ha2fabric@ec7deea/go.mod:7` `github.com/SukramJ/go-fabric-ui v0.9.2`; no `replace` |
| Construction of `fabricui.Options` (one) and the handler call (`fabricui.Handler(s.opts.UI)`); no `fabricui.New`, no `Mount` | `go-ha2fabric@ec7deea/cmd/ha2fabric/main.go:112-116`; `go-ha2fabric@ec7deea/internal/web/web.go:69` |
| Fields set: `Node: def.sw`, `Exposure: def.prov`, `Admin: admin(cfg)`, `Reassemble: def.Resync`, `Events: def.events`, `Warnings: warnings(def, fl)`, `Counters: counters(def, fl)`, `Erase: def.erase`, `Logger: logger`, `Operations: web.NewOperations(ring, version.Version, started)`, `Registry: fl`, `NodeConfig: nodeConfig(cfg)` | `go-ha2fabric@ec7deea/cmd/ha2fabric/main.go:112-116` |
| Not set: `Prefix`, `Controller`, `Nodes`, `ExposureReason`, `ConfiguredWindow`, `AllowEmptyTopology` | absent from `main.go:112-116` |
| `web.New` fills `UI.Logger` from `Options.Logger` when nil; `Server.Handler()` = `s.filter(fabricui.Handler(s.opts.UI))` | `go-ha2fabric@ec7deea/internal/web/web.go:50-57, :68-70` |
| Client filter `filter`: when `Allow` is non-empty, `net.SplitHostPort(r.RemoteAddr)`, `net.ParseIP`, must equal one of `Allow`, else `403 "forbidden"`; empty `Allow` passes everything | `go-ha2fabric@ec7deea/internal/web/web.go:59-63, :76-89` |
| Server: `http.Server{Addr, Handler, ReadHeaderTimeout: 10s, ReadTimeout: 30s, IdleTimeout: 120s}`, no `WriteTimeout` (comment: the event stream stays open); `net.ListenConfig.Listen(ctx, "tcp", s.opts.Listen)`; plain HTTP, no TLS | `go-ha2fabric@ec7deea/internal/web/web.go:93-108` |
| Mount: root of its own server (API at `/api/v1`, SPA at `/`); no mux, prefix or other middleware | `go-ha2fabric@ec7deea/internal/web/web.go:68-70, :96-100` |
| `Admin` func `admin(cfg)`: `cfg.Addon` -> `web.From(config.IngressProxy)` (client IP equals `172.30.32.2`); `cfg.WebAdmin` -> `nil` (token mode); else `web.ReadOnly` (admits nobody) | `go-ha2fabric@ec7deea/cmd/ha2fabric/main.go:185-194`; `web.From` `internal/web/web.go:126-133`; `ReadOnly` `:137`; `IngressProxy` `internal/config/config.go:63`; `Addon` = `SUPERVISOR_TOKEN` present `config.go:212` (comment `:74-76`) |
| Config: `WEB_LISTEN` / `--web-listen` default `127.0.0.1:8099` (empty disables); `WEB_ALLOW` / `--web-allow` default `""` (all clients); `WEB_ADMIN` / `--web-admin` default `false` | `go-ha2fabric@ec7deea/internal/config/config.go:48, :196-198`; validation `:286-288, :303-310` |
| Listener is started only when `cfg.WebListen != ""` | `go-ha2fabric@ec7deea/cmd/ha2fabric/main.go:111-118` |
| Add-on wiring: `WEB_LISTEN="0.0.0.0:8486"`, `WEB_ALLOW="172.30.32.2"`; manifest `host_network: true`, `ingress: true`, `ingress_port: 8486` | `go-ha2fabric@ec7deea/script/run.sh:70-71`; `go-ha2fabric@ec7deea/addon/config.yaml:37, :43-44` |
| Container image default `WEB_LISTEN=0.0.0.0:8099` (no `WEB_ALLOW` set there) | `go-ha2fabric@ec7deea/Dockerfile:31` |
| Auth in front, as read: the `Allow` source filter and, in add-on mode, `Admin` by Ingress source address; no CORS, no TLS in `internal/web` | `go-ha2fabric@ec7deea/internal/web/web.go:76-89, :96-108` |

### 5.3 Pins at a glance

| Host | go-fabric-ui version in `go.mod` | Admin mode in add-on | Admin mode stand-alone default | Listen (add-on) | Listen default |
|---|---|---|---|---|---|
| go-fabric-server@d61a545 | v0.9.2 (`go.mod:7`) | `func`: peer address == `172.30.32.2` (`web.go:97-102`) | none unless `--web-admin` (then token mode), else read-only (`web.go:103-106`) | `0.0.0.0:8099` (`script/run.sh:56`) | `--web` empty: pages off (`run.go:174`) |
| go-ha2fabric@ec7deea | v0.9.2 (`go.mod:7`) | `web.From("172.30.32.2")` (`main.go:187-188`) | token mode if `WEB_ADMIN=true`, else `web.ReadOnly` (`main.go:189-192`) | `0.0.0.0:8486` (`script/run.sh:70`) | `127.0.0.1:8099` (`config.go:48`) |


---

## Counts

| Section | Count |
|---|---|
| 1. Routes | 97 contract operations = 97 code routes (24 node-family x 2 = 48, registry 3, `GET /nodes` 1, operation 5, controller 40); 46 gated (`adminOnly` 20, registry `adminOnlyBecause` 3, `PUT /operation/log/level` 1, `controllerAdmin` 22) + 1 query-parameter gate (`refresh`); 4 SSE routes; 26 routes that read a body (node 10, registry 2, operation 1, controller 13); 0 code/contract mismatches; 13 `Header().Set`/`hdr.Set` call sites, 6 distinct header names; 2 entry points (`Mount`, `Handler`) |
| 2. Parse sites | Go 28 rows (G1-G28); TS 26 rows (T1-T26) |
| 3. Render sinks | `{@html}` 2 sites (both the generated QR SVG); `target="_blank"` 2 (both with `rel`); dynamic `href` 10 sites; CSS custom-property injection 2 sites (`ha-bridge.ts`, `index.html`); string-render tables 3.3: 5 groups (kit, node pages, operation, controller pages, shell) |
| 4. Secrets | 13 field rows (S1-S13) of which 6 carry a secret value (S1 token, S2 pairing codes+passcode, S3 window codes+passcode, S4 node window codes+passcode, S5 commission code/passcode, S6 Wi-Fi/Thread secret); 19 log-call rows; 3 ungated routes return the admin token (`GET /node/status`, `GET /nodes/{node}/status`, `GET /controller/status`) |
| 5. Hosts | 2 hosts, 2 `Options` constructions, both call `fabricui.Handler`, both pin `v0.9.2` |

## Read but empty (greps that found nothing)

- Go, `fabricui` + `internal` non-test: `Access-Control`, `Content-Security`, `X-Frame`, `Referrer-Policy`, `Set-Cookie`/`SetCookie`, `Strict-Transport`, `Permissions-Policy`, `Cross-Origin`, `CheckOrigin`; `Last-Event-ID`/`LastEventID`; `base64.` (no use in `fabricui`, `internal/api` generated code excluded); `url.Parse`, `FormValue`, `ParseForm`, `ReadAll`, `ParseDuration`, `time.Parse`, `Sscanf`; `log.Print*`/`log.Fatal*`; `verifier`/`PAKE`/`network_key`; `fabricui.New(` (no such function).
- Hand-written decoders of setup codes, QR payloads, Thread datasets, hex 64-bit ids inside this repo: setup code/QR parsing is delegated (`controller.ParsePairingCode` `fabricui/gofabric/gofabric.go:292`, `thread.DecodeDataset` `:89`); the only hand-written hex id parser is `parseHex` (`handlers_controller.go:208`).
- TS (`ui/src/lib`, `ui/src/shell`, non-test): `postMessage`, `addEventListener("message")`, `sessionStorage`, `document.cookie`, `window.name`, `atob`/`btoa`, `new URL`, `BigInt(`, `parseFloat`, `FileReader`, `window.open`, `eval(`, `new Function`, `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `document.write`, `srcdoc`, `use:` actions, `<iframe>`, `<img>`, `src=`, `action=`, `target="_blank"` without `rel`, `import.meta`, `console.*` outside `context.ts:128-130`; callers of `loadLS`/`saveLS` (defined, never called).
- `ui/src/shell/index.html`: no CSP meta, no `http-equiv`.

## Unresolved / unclear (as read)

- `runtime.BindJSONBody` behaviour (unknown fields, trailing data) for the 13 generated controller bodies (G4): oapi-codegen runtime not in the repo.
- Whether error text from go-fabric (`controller.ParsePairingCode`, `thread.DecodeDataset`, node errors) can contain a code or dataset when echoed in problem details (S13): go-fabric not read.
- `exposureReason` text in the server (`internal/daemon/operation.go:17`) and the sources of the error values logged at `handlers_pairing.go:79, :180`.
- Whether `HA_THEME_VARS` (`ha-bridge.ts:11`) and the inline `vars` list (`index.html:21-53`) are identical: the repo names a test (`theme.test.ts`, per `index.html` comment) but I did not compare the lists.
