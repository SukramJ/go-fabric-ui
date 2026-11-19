# Review phase 3 — contract and api map (script, read-only map)

Generated at `f4b8573` by the coordinator's script over a `git archive` export; operations of `contract/openapi.yaml` with their `x-admin` mark, the hand-written route patterns, the client's request sites, the fixture count, the npm exports. A map, never a citation: every row is a grep or script hit, not a reading.

## contract operations (line: path method operationId x-admin)
41: /node/status: get: getNodeStatus 
52: /node/fabrics: get: listFabrics 
65: /node/fabrics/{index}: delete: removeFabric admin
82: /node/pairing: get: getPairing admin
96: /node/pairing/window: post: openWindow admin
118: /node/pairing/window/close: post: closeWindow admin
129: /node/endpoints: get: listEndpoints 
149: /node/endpoints/{ep}: get: getEndpointDetail 
181: /node/endpoints/{ep}/events: get: listEndpointEvents 
204: /node/surface: get: getNodeSurface 
223: /node/sessions: get: listSessions 
235: /node/topology: get: getNodeTopology 
257: /node/mdns: get: getMdns 
269: /node/mdns/reannounce: post: reannounce admin
285: /node/compatibility: get: getCompatibility 
297: /node/events: get: listDiagnosticEvents 
309: /node/exposure/schema: get: getExposureSchema 
322: /node/exposure: get: listExposure 
334: /node/exposure: put: applyExposure admin
356: /node/exposure/preview: post: previewExposure admin
377: /node/maintenance/reassemble: post: reassemble admin
388: /node/maintenance/factory-reset: post: factoryReset admin
405: /node/maintenance/erase: post: erase admin
433: /node/stream: get: stream 
479: /nodes: get: listNodes 
497: /nodes: post: createNode admin
528: /nodes/{node}: delete: deleteNode admin
549: /nodes/{node}/label: put: renameNode admin
574: /nodes/{node}/status: get: getNodeStatusForNode 
589: /nodes/{node}/fabrics: get: listFabricsForNode 
606: /nodes/{node}/fabrics/{index}: delete: removeFabricForNode admin
627: /nodes/{node}/pairing: get: getPairingForNode admin
645: /nodes/{node}/pairing/window: post: openWindowForNode admin
671: /nodes/{node}/pairing/window/close: post: closeWindowForNode admin
686: /nodes/{node}/endpoints: get: listEndpointsForNode 
703: /nodes/{node}/endpoints/{ep}: get: getEndpointDetailForNode 
722: /nodes/{node}/endpoints/{ep}/events: get: listEndpointEventsForNode 
742: /nodes/{node}/surface: get: getNodeSurfaceForNode 
758: /nodes/{node}/sessions: get: listSessionsForNode 
774: /nodes/{node}/topology: get: getNodeTopologyForNode 
791: /nodes/{node}/mdns: get: getMdnsForNode 
807: /nodes/{node}/mdns/reannounce: post: reannounceForNode admin
821: /nodes/{node}/compatibility: get: getCompatibilityForNode 
837: /nodes/{node}/events: get: listDiagnosticEventsForNode 
853: /nodes/{node}/exposure/schema: get: getExposureSchemaForNode 
870: /nodes/{node}/exposure: get: listExposureForNode 
884: /nodes/{node}/exposure: put: applyExposureForNode admin
910: /nodes/{node}/exposure/preview: post: previewExposureForNode admin
935: /nodes/{node}/maintenance/reassemble: post: reassembleForNode admin
950: /nodes/{node}/maintenance/factory-reset: post: factoryResetForNode admin
971: /nodes/{node}/maintenance/erase: post: eraseForNode admin
993: /nodes/{node}/stream: get: streamForNode 
1006: /operation/info: get: getOperationInfo 
1018: /operation/log/level: get: getLogLevel 
1029: /operation/log/level: put: setLogLevel admin
1050: /operation/log: get: listLogRecords 
1071: /operation/log/stream: get: logStream 
1089: /controller/status: get: getControllerStatus 
1104: /controller/nodes: get: listControllerNodes 
1117: /controller/nodes/{id}: get: getControllerNode 
1129: /controller/nodes/{id}: delete: decommissionNode admin
1145: /controller/nodes/{id}/endpoints/{ep}: get: getNodeEndpoint 
1175: /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}: get: getNodeCluster 
1199: /controller/nodes/{id}/events: get: listNodeEvents 
1224: /controller/nodes/{id}/interview: post: interviewNode admin
1238: /controller/nodes/{id}/ping: post: pingNode admin
1270: /controller/nodes/{id}/enable: post: enableNode admin
1284: /controller/nodes/{id}/disable: post: disableNode admin
1298: /controller/nodes/{id}/auto-subscribe: post: setNodeAutoSubscribe admin
1317: /controller/nodes/{id}/window: post: openNodeWindow admin
1340: /controller/nodes/{id}/window/close: post: closeNodeWindow admin
1354: /controller/nodes/{id}/attributes: post: readNodeAttributes 
1378: /controller/nodes/{id}/attributes: put: writeNodeAttribute admin
1401: /controller/nodes/{id}/commands: post: invokeNodeCommand admin
1424: /controller/nodes/{id}/acl: get: getNodeAcl 
1437: /controller/nodes/{id}/acl: put: setNodeAcl admin
1464: /controller/nodes/{id}/bindings: get: getNodeBindings 
1477: /controller/nodes/{id}/bindings: put: setNodeBindings admin
1500: /controller/nodes/{id}/icd: get: getNodeIcd 
1515: /controller/nodes/{id}/icd/register: post: registerNodeIcd admin
1538: /controller/nodes/{id}/icd/unregister: post: unregisterNodeIcd admin
1560: /controller/nodes/{id}/fabrics: get: listNodeFabrics 
1580: /controller/nodes/{id}/fabrics/{index}: delete: removeNodeFabric admin
1593: /controller/commission: post: commissionNode admin
1624: /controller/discovery: get: discoverDevices admin
1652: /controller/credentials: get: listCredentials 
1664: /controller/credentials: put: setCredential admin
1680: /controller/credentials: delete: removeCredential admin
1702: /controller/fabric/label: put: setFabricLabel admin
1718: /controller/updates: get: listUpdates 
1729: /controller/updates/consent: post: consentUpdate admin
1747: /controller/updates/check: post: checkUpdates admin
1762: /controller/topology: get: getControllerTopology admin
1800: /controller/diagnostics: get: getControllerDiagnostics 
1822: /controller/mdns: get: getControllerMdns 
1842: /controller/events: get: listControllerDiagnosticEvents 
1860: /controller/stream: get: controllerStream 

## hand-written route patterns (fabricui, non-test)
fabricui/handler.go:340 var operationAdmin = map[string]bool{"PUT /operation/log/level": true}
fabricui/handler.go:450 writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+r.PathValue("node"))
fabricui/handlers_controller.go:56 "DELETE /controller/nodes/{id}":                 true,
fabricui/handlers_controller.go:57 "POST /controller/nodes/{id}/interview":         true,
fabricui/handlers_controller.go:58 "POST /controller/nodes/{id}/enable":            true,
fabricui/handlers_controller.go:59 "POST /controller/nodes/{id}/disable":           true,
fabricui/handlers_controller.go:60 "POST /controller/nodes/{id}/auto-subscribe":    true,
fabricui/handlers_controller.go:61 "POST /controller/nodes/{id}/window":            true,
fabricui/handlers_controller.go:62 "POST /controller/nodes/{id}/window/close":      true,
fabricui/handlers_controller.go:63 "PUT /controller/nodes/{id}/attributes":         true,
fabricui/handlers_controller.go:64 "POST /controller/nodes/{id}/commands":          true,
fabricui/handlers_controller.go:65 "PUT /controller/nodes/{id}/acl":                true,
fabricui/handlers_controller.go:66 "PUT /controller/nodes/{id}/bindings":           true,
fabricui/handlers_controller.go:67 "POST /controller/nodes/{id}/icd/register":      true,
fabricui/handlers_controller.go:68 "POST /controller/nodes/{id}/icd/unregister":    true,
fabricui/handlers_controller.go:69 "DELETE /controller/nodes/{id}/fabrics/{index}": true,
fabricui/handlers_controller.go:70 "POST /controller/commission":                   true,
fabricui/handlers_controller.go:71 "GET /controller/discovery":                     true,
fabricui/handlers_controller.go:72 "PUT /controller/credentials":                   true,
fabricui/handlers_controller.go:73 "DELETE /controller/credentials":                true,
fabricui/handlers_controller.go:74 "PUT /controller/fabric/label":                  true,
fabricui/handlers_controller.go:75 "POST /controller/nodes/{id}/ping":              true,
fabricui/handlers_controller.go:76 "POST /controller/updates/consent":              true,
fabricui/handlers_controller.go:77 "POST /controller/updates/check":                true,
fabricui/handlers_controller.go:84 "GET /controller/topology": "refresh",
fabricui/registry.go:246 writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+key)

## client.ts request sites

## fixtures
142

## exported Go identifiers (fabricui, fabricui/gofabric)
301

## npm exports
  "exports": {
    ".": {
      "types": "./src/lib/index.ts",
      "svelte": "./src/lib/index.ts",
      "default": "./src/lib/index.ts"
    },
    "./pages": {
      "types": "./src/lib/pages/index.ts",
      "svelte": "./src/lib/pages/index.ts",
      "default": "./src/lib/pages/index.ts"
    },
    "./kit": {
      "types": "./src/lib/kit/index.ts",
      "svelte": "./src/lib/kit/index.ts",
      "default": "./src/lib/kit/index.ts"
    },
    "./theme": {
      "types": "./src/lib/theme/index.ts",
      "default": "./src/lib/theme/index.ts"
    },
    "./theme/tokens.css": "./src/lib/theme/tokens.css",
    "./client": {
      "types": "./src/lib/api/index.ts",
      "default": "./src/lib/api/index.ts"
    },
    "./i18n": {
      "types": "./src/lib/i18n/index.ts",
      "default": "./src/lib/i18n/index.ts"
    },
    "./package.json": "./package.json"
  },
