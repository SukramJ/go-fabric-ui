# Review phase 2 — inventory (script, read-only map)

Generated at `1056db24a68eaf891149dd6383eb038dc9eb2624` by grep over the hand-written Go (`fabricui/`, without
tests and generated code) and web source (`ui/src`, as phase 1b). A map,
never a citation: every row is a grep hit, not a reading.

## 1. Routes and gates

### contract: operationId and x-admin (in order)

```
19: Admin routes (`x-admin: true`) answer 403 unless the host's admin
43: operationId: getNodeStatus
54: operationId: listFabrics
67: operationId: removeFabric
69: x-admin: true
84: operationId: getPairing
86: x-admin: true
98: operationId: openWindow
100: x-admin: true
120: operationId: closeWindow
122: x-admin: true
131: operationId: listEndpoints
151: operationId: getEndpointDetail
183: operationId: listEndpointEvents
206: operationId: getNodeSurface
225: operationId: listSessions
237: operationId: getNodeTopology
259: operationId: getMdns
271: operationId: reannounce
279: x-admin: true
287: operationId: getCompatibility
299: operationId: listDiagnosticEvents
311: operationId: getExposureSchema
324: operationId: listExposure
336: operationId: applyExposure
338: x-admin: true
358: operationId: previewExposure
360: x-admin: true
379: operationId: reassemble
381: x-admin: true
390: operationId: factoryReset
392: x-admin: true
407: operationId: erase
419: x-admin: true
435: operationId: stream
481: operationId: listNodes
499: operationId: createNode
508: x-admin: true
530: operationId: deleteNode
538: x-admin: true
551: operationId: renameNode
557: x-admin: true
576: operationId: getNodeStatusForNode
591: operationId: listFabricsForNode
608: operationId: removeFabricForNode
611: x-admin: true
629: operationId: getPairingForNode
632: x-admin: true
647: operationId: openWindowForNode
650: x-admin: true
673: operationId: closeWindowForNode
676: x-admin: true
688: operationId: listEndpointsForNode
705: operationId: getEndpointDetailForNode
724: operationId: listEndpointEventsForNode
744: operationId: getNodeSurfaceForNode
760: operationId: listSessionsForNode
776: operationId: getNodeTopologyForNode
793: operationId: getMdnsForNode
809: operationId: reannounceForNode
812: x-admin: true
823: operationId: getCompatibilityForNode
839: operationId: listDiagnosticEventsForNode
855: operationId: getExposureSchemaForNode
872: operationId: listExposureForNode
886: operationId: applyExposureForNode
889: x-admin: true
912: operationId: previewExposureForNode
915: x-admin: true
937: operationId: reassembleForNode
940: x-admin: true
952: operationId: factoryResetForNode
955: x-admin: true
973: operationId: eraseForNode
976: x-admin: true
995: operationId: streamForNode
1008: operationId: getOperationInfo
1020: operationId: getLogLevel
1031: operationId: setLogLevel
1033: x-admin: true
1052: operationId: listLogRecords
1073: operationId: logStream
1091: operationId: getControllerStatus
1106: operationId: listControllerNodes
1119: operationId: getControllerNode
1131: operationId: decommissionNode
1132: x-admin: true
1147: operationId: getNodeEndpoint
1177: operationId: getNodeCluster
1201: operationId: listNodeEvents
1226: operationId: interviewNode
1227: x-admin: true
1240: operationId: pingNode
1241: x-admin: true
1272: operationId: enableNode
1273: x-admin: true
1286: operationId: disableNode
1287: x-admin: true
1300: operationId: setNodeAutoSubscribe
1301: x-admin: true
1319: operationId: openNodeWindow
1320: x-admin: true
1342: operationId: closeNodeWindow
1343: x-admin: true
1356: operationId: readNodeAttributes
1380: operationId: writeNodeAttribute
1381: x-admin: true
1403: operationId: invokeNodeCommand
1404: x-admin: true
1426: operationId: getNodeAcl
1439: operationId: setNodeAcl
1440: x-admin: true
1466: operationId: getNodeBindings
1479: operationId: setNodeBindings
1480: x-admin: true
1502: operationId: getNodeIcd
1517: operationId: registerNodeIcd
1518: x-admin: true
1540: operationId: unregisterNodeIcd
1541: x-admin: true
1562: operationId: listNodeFabrics
1582: operationId: removeNodeFabric
1583: x-admin: true
1595: operationId: commissionNode
1596: x-admin: true
1626: operationId: discoverDevices
1627: x-admin: true
1654: operationId: listCredentials
1666: operationId: setCredential
1667: x-admin: true
1682: operationId: removeCredential
1683: x-admin: true
1704: operationId: setFabricLabel
1705: x-admin: true
1720: operationId: listUpdates
1731: operationId: consentUpdate
1732: x-admin: true
1749: operationId: checkUpdates
1750: x-admin: true
1764: operationId: getControllerTopology
1789: - { name: refresh, in: query, x-admin: true, schema: { type: boolean }, description: "Re-read the nodes' network diagnostics before answering; admin only." }
1802: operationId: getControllerDiagnostics
1824: operationId: getControllerMdns
1844: operationId: listControllerDiagnosticEvents
1862: operationId: controllerStream
```

### gate sites

```
fabricui/handler.go:436: write := func(pattern string, fn http.HandlerFunc) { handle(pattern, h.adminOnly(fn)) }
fabricui/handler.go:455: handle(rt.method+" "+h.prefix+"/nodes"+rt.path, h.adminOnlyBecause(func(w http.ResponseWriter, r *http.Request) { rt.serve(h, w, r) },
fabricui/handler.go:460: handle("PUT "+h.prefix+"/operation/log/level", h.adminOnlyBecause(h.setLogLevel, "this route changes the daemon's log level"))
fabricui/handler.go:603:// adminOnly refuses a request the host's Admin does not admit, or, in
fabricui/handler.go:606:func (h *handler) adminOnly(next http.HandlerFunc) http.HandlerFunc {
fabricui/handler.go:607: return h.adminOnlyBecause(next, "this route writes or reads a pairing secret")
fabricui/handler.go:610:// adminOnlyBecause is adminOnly with the refusal's detail.
fabricui/handler.go:611:func (h *handler) adminOnlyBecause(next http.HandlerFunc, detail string) http.HandlerFunc {
fabricui/handler.go:613: if !h.isAdmin(r) {
fabricui/handler.go:621:func (h *handler) isAdmin(r *http.Request) bool {
fabricui/handler.go:623: return h.admin(r)
fabricui/handlers_controller.go:119: return h.controllerGate(sub)
fabricui/handlers_controller.go:122:// controllerGate refuses an admin route the request may not use before
fabricui/handlers_controller.go:124:func (h *handler) controllerGate(sub *http.ServeMux) http.Handler {
fabricui/handlers_controller.go:129: if controllerAdmin[route] && !h.isAdmin(r) {
fabricui/handlers_controller.go:134: if q, ok := controllerAdminQuery[route]; ok && !h.isAdmin(r) {
fabricui/handlers_node.go:141: admin := h.admin(r)
```

### headers

```
fabricui/handler.go:27:const TokenHeader = "X-Fabric-UI-Token" //nolint:gosec // a header name, not a credential
fabricui/handler.go:379: hdr.Set("Cache-Control", "no-store")
fabricui/handler.go:388: hdr.Set("X-Content-Type-Options", "nosniff")
fabricui/handlers_surface.go:28: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/options.go:225: // requires it in the X-Fabric-UI-Token header of every such request.
fabricui/options.go:270: // the web interface [Handler] serves: its Content-Security-Policy says
fabricui/problem.go:73: w.Header().Set("Content-Type", "application/problem+json")
fabricui/problem.go:127: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/registry.go:304: w.Header().Set("Content-Type", "application/json; charset=utf-8")
fabricui/spa.go:60: w.Header().Set("Cache-Control", cacheControlFor(path))
fabricui/spa.go:96: hdr.Set("Content-Type", "text/html; charset=utf-8")
fabricui/spa.go:97: hdr.Set("Cache-Control", "no-store")
fabricui/spa.go:98: hdr.Set("Content-Security-Policy", i.csp)
fabricui/stream.go:51: hdr.Set("Content-Type", "text/event-stream")
fabricui/stream.go:52: hdr.Set("Cache-Control", "no-store")
```

## 2. Secret flows (X1)

### secrets: Go

```
fabricui/controller.go:27: // routes; the credential and update routes also ask Credentials and
fabricui/controller.go:34: // Pair commissions the device a QR payload or a manual pairing code
fabricui/controller.go:35: // names. The code carries the passcode; it is never logged.
fabricui/controller.go:41: // its passcode.
fabricui/controller.go:42: CommissionDiscovered(ctx context.Context, d Discovered, passcode uint32, opts PairOptions) (ControllerNode, error)
fabricui/controller.go:48: // Credentials is the host's store of Wi-Fi and Thread credentials;
fabricui/controller.go:50: Credentials() CredentialStore
fabricui/controller.go:193: // Thread RLOC16, nil when unknown; SSID and BSSID the Wi-Fi network.
fabricui/controller.go:196: SSID, BSSID string
fabricui/controller.go:262: // carries the window's passcode.
fabricui/controller.go:337: // Discriminator, VendorID, ProductID and DeviceType are what the node
fabricui/controller.go:339: Discriminator *uint16
fabricui/controller.go:624:// WindowInfo is an open commissioning window on a node. Passcode is a
fabricui/controller.go:628: ManualCode, QRCode string
fabricui/controller.go:629: Passcode uint32
fabricui/controller.go:630: Discriminator uint16
fabricui/controller.go:679: // WiFiID and ThreadID name a stored credential ([CredentialStore])
fabricui/controller.go:709:// network_only and named credential, the code's discovery capabilities —
fabricui/controller.go:724:// central; the operator's network_only; the device's QR code without the
fabricui/controller.go:725:// BLE bit; no credential named, and a device on BLE has no network yet.
fabricui/controller.go:752:// "long_discriminator", "short_discriminator", "vendor" or "device_type",
fabricui/controller.go:753:// with Value the discriminator, vendor id or device type.
fabricui/controller.go:765: LongDiscriminator uint16
fabricui/controller.go:775:// CredentialStore is the host's store of network credentials for a
fabricui/controller.go:776:// commissioning: Wi-Fi SSIDs with their passphrase, Thread operational
fabricui/controller.go:777:// datasets. The routes read names only; a secret goes in and never comes
fabricui/controller.go:779:type CredentialStore interface {
fabricui/controller.go:780: // List is every stored credential, by kind and id; Name is the SSID
fabricui/controller.go:782: List(ctx context.Context) ([]Credential, error)
fabricui/controller.go:783: SetWiFi(ctx context.Context, id, ssid, passphrase string) error
fabricui/controller.go:784: // SetThread stores an operational dataset (MeshCoP TLVs); an
fabricui/controller.go:785: // invalid one wraps [ErrInvalidCredential].
fabricui/controller.go:786: SetThread(ctx context.Context, id string, dataset []byte) error
fabricui/controller.go:787: // Remove forgets a credential; one the store does not hold wraps
fabricui/controller.go:788: // [ErrUnknownCredential].
fabricui/controller.go:789: Remove(ctx context.Context, kind CredentialKind, id string) error
fabricui/controller.go:792:// CredentialKind is "wifi" or "thread".
fabricui/controller.go:793:type CredentialKind string
fabricui/controller.go:795:// The credential kinds.
fabricui/controller.go:797: CredentialWiFi CredentialKind = "wifi"
fabricui/controller.go:798: CredentialThread CredentialKind = "thread"
fabricui/controller.go:801:// Credential is a stored credential's name, never its secret.
fabricui/controller.go:802:type Credential struct {
fabricui/controller.go:803: Kind CredentialKind
fabricui/controller.go:808: // (contract 0.7.1, Credential.default); the pages preselect it. A
fabricui/controller.go:908: // ErrInvalidCredential: a credential the store refuses (400).
fabricui/controller.go:909: ErrInvalidCredential = errors.New("fabricui: invalid credential")
fabricui/controller.go:910: // ErrUnknownCredential: no stored credential with the kind and id
fabricui/controller.go:912: ErrUnknownCredential = errors.New("fabricui: no such credential")
fabricui/events.go:57:// A payload never carries a passcode.
fabricui/events.go:73: Discriminator uint16 `json:"discriminator,omitempty"`
fabricui/events.go:335: return WindowEvent{Discriminator: w.Discriminator, DurationSeconds: durationSeconds(w.Timeout)}
fabricui/gofabric/gofabric.go:12:// identifiers client/discovery, the Thread dataset codec thread, and the
fabricui/gofabric/gofabric.go:15:// icdmanagement and operationalcredentials). A host passes it in
fabricui/gofabric/gofabric.go:24:// src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{Hooks: hooks, Credentials: store, Topology: svc})
fabricui/gofabric/gofabric.go:53:// Credentials is the host's credential store as a commissioning reads it:
fabricui/gofabric/gofabric.go:54:// the routes' [fabricui.CredentialStore] and the secrets a pairing that
fabricui/gofabric/gofabric.go:55:// names a credential needs. *fabricui.MemoryCredentials satisfies it.
fabricui/gofabric/gofabric.go:56:type Credentials interface {
fabricui/gofabric/gofabric.go:57: fabricui.CredentialStore
fabricui/gofabric/gofabric.go:58: WiFi(id string) (ssid, passphrase string, ok bool)
fabricui/gofabric/gofabric.go:59: Thread(id string) (dataset []byte, ok bool)
fabricui/gofabric/gofabric.go:72: // Credentials is the host's Wi-Fi and Thread credential store; nil
fabricui/gofabric/gofabric.go:73: // serves none (the credential routes answer 503).
fabricui/gofabric/gofabric.go:74: Credentials Credentials
fabricui/gofabric/gofabric.go:80: // it logs carries a passcode or a credential.
fabricui/gofabric/gofabric.go:90:// ThreadDatasetName is a Thread operational dataset's network name, read
fabricui/gofabric/gofabric.go:92:// OperationalDataset.decode); a dataset that does not decode wraps
fabricui/gofabric/gofabric.go:93:// fabricui.ErrInvalidCredential. It is the threadName of
fabricui/gofabric/gofabric.go:94:// fabricui.NewMemoryCredentials for a store that checks what it keeps.
fabricui/gofabric/gofabric.go:95:func ThreadDatasetName(dataset []byte) (string, error) {
fabricui/gofabric/gofabric.go:96: ds, err := thread.DecodeDataset(dataset)
fabricui/gofabric/gofabric.go:98: return "", fmt.Errorf("%w: %w", fabricui.ErrInvalidCredential, err)
fabricui/gofabric/gofabric.go:101: return "", fmt.Errorf("%w: the dataset names no network", fabricui.ErrInvalidCredential)
fabricui/gofabric/gofabric.go:266:// credentials the request names.
fabricui/gofabric/gofabric.go:271: if s.opts.Credentials == nil {
fabricui/gofabric/gofabric.go:272: return nil, nil, fmt.Errorf("%w: the host stores no credentials", fabricui.ErrInvalidArgument)
fabricui/gofabric/gofabric.go:277: ssid, pass, ok := s.opts.Credentials.WiFi(o.WiFiID)
fabricui/gofabric/gofabric.go:279: return nil, nil, fmt.Errorf("%w: wifi %q", fabricui.ErrUnknownCredential, o.WiFiID)
fabricui/gofabric/gofabric.go:281: wifi = &commissioner.WiFiNetwork{SSID: ssid, Credentials: pass}
fabricui/gofabric/gofabric.go:284: ds, ok := s.opts.Credentials.Thread(o.ThreadID)
fabricui/gofabric/gofabric.go:286: return nil, nil, fmt.Errorf("%w: thread %q", fabricui.ErrUnknownCredential, o.ThreadID)
fabricui/gofabric/gofabric.go:288: th = &commissioner.ThreadNetwork{OperationalDataset: ds}
fabricui/gofabric/gofabric.go:293:// capabilityBLE is the BLE bit of a QR payload's discovery capabilities
fabricui/gofabric/gofabric.go:299:// allows every transport; a QR code allows BLE with bit 1, go-fabric
fabricui/gofabric/gofabric.go:301:// credential — go-fabric refuses BLE without a network
fabricui/gofabric/gofabric.go:322:// rule it out, the code allows it and a named credential puts the device
fabricui/gofabric/gofabric.go:331: return nil, fmt.Errorf("%w: not a QR payload or a manual pairing code", fabricui.ErrInvalidArgument)
fabricui/gofabric/gofabric.go:419: case "long_discriminator":
fabricui/gofabric/gofabric.go:421: return discovery.Identifier{}, fmt.Errorf("%w: a long discriminator has 12 bits", fabricui.ErrInvalidArgument)
fabricui/gofabric/gofabric.go:423: return discovery.ByLongDiscriminator(uint16(f.Value)), nil
fabricui/gofabric/gofabric.go:424: case "short_discriminator":
fabricui/gofabric/gofabric.go:426: return discovery.Identifier{}, fmt.Errorf("%w: a short discriminator has 4 bits", fabricui.ErrInvalidArgument)
fabricui/gofabric/gofabric.go:428: return discovery.ByShortDiscriminator(uint8(f.Value)), nil
fabricui/gofabric/gofabric.go:461: InstanceName: d.InstanceID, Hostname: d.Hostname, LongDiscriminator: d.D, CommissioningMode: d.CM,
fabricui/gofabric/gofabric.go:544:// CommissionDiscovered commissions a discovered device with its passcode
fabricui/gofabric/gofabric.go:550:func (s *source) CommissionDiscovered(ctx context.Context, d fabricui.Discovered, passcode uint32, o fabricui.PairOptions) (fabricui.ControllerNode, error) {
fabricui/gofabric/gofabric.go:553: if d.LongDiscriminator == 0 {
fabricui/gofabric/gofabric.go:556: found, err := s.ctrl.Discover(ctx, discovery.ByLongDiscriminator(d.LongDiscriminator), rediscoverWindow)
fabricui/gofabric/gofabric.go:581: Passcode: passcode, OnAttestationFailure: po.OnAttestationFailure,
fabricui/gofabric/gofabric.go:708:func (s *source) Credentials() fabricui.CredentialStore {
fabricui/gofabric/gofabric.go:709: if s.opts.Credentials == nil {
fabricui/gofabric/gofabric.go:712: return s.opts.Credentials
fabricui/gofabric/node.go:22: "github.com/SukramJ/go-fabric/cluster/spec/operationalcredentials"
fabricui/gofabric/node.go:64: Discriminator: r.Discriminator, VendorID: r.VendorID, ProductID: r.ProductID, DeviceType: r.DeviceType,
fabricui/gofabric/node.go:487: path := fabricui.AttributePath{Endpoint: 0, Cluster: operationalcredentials.ClusterID, Attribute: operationalcredentials.AttrCurrentFabricIndex}
fabricui/gofabric/node.go:659: ManualCode: cw.ManualCode, QRCode: cw.QRCode, Passcode: cw.Passcode, Discriminator: cw.Discriminator,
fabricui/gofabric/topology.go:182: d.add("ssid", n.SSID)
fabricui/gofabric/topology.go:183: d.add("bssid", n.BSSID)
fabricui/gofabric/topology.go:196: Role: role, RLOC16: n.Rloc16, SSID: n.SSID, BSSID: n.BSSID, Details: d.list(),
fabricui/gofabric/topology.go:265: d.add("ssid", n.SSID)
fabricui/gofabric/topology.go:266: d.add("bssid", n.BSSID)
fabricui/gofabric/topology.go:287:// accessPointOf is a Wi-Fi access point, labelled by the SSID its
fabricui/gofabric/topology.go:288:// stations name, else by its BSSID.
fabricui/gofabric/topology.go:290: label := n.SSID
fabricui/gofabric/topology.go:292: label = n.BSSID
fabricui/gofabric/topology.go:295: d.add("bssid", n.BSSID)
fabricui/gofabric/topology.go:296: d.add("ssid", n.SSID)
fabricui/gofabric/topology.go:297: return fabricui.NetworkNode{ID: n.ID, Kind: "wifi_ap", Label: label, State: "unknown", SSID: n.SSID, BSSID: n.BSSID, Details: d.list()}
fabricui/gofabric/updates.go:66:// image or a credential beyond the version reaches the event.
fabricui/handler.go:27:const TokenHeader = "X-Fabric-UI-Token" //nolint:gosec // a header name, not a credential
fabricui/handler.go:58: // fresh draws the credentials of a window on fresh credentials:
fabricui/handler.go:59: // node.NewWindowCredentials.
fabricui/handler.go:60: fresh func(time.Duration) (node.WindowCredentials, error)
fabricui/handler.go:93: // mu guards ephemeral: the credentials of the window this handler
fabricui/handler.go:96: ephemeral *ephemeralWindow
fabricui/handler.go:266: fresh: node.NewWindowCredentials,
fabricui/handlers_controller.go:29:// OperationalCredentials Fabrics (0x003E/0x0001), CurrentFabricIndex
fabricui/handlers_controller.go:39: clusterOperationalCredentials uint32 = 0x003E
fabricui/handlers_controller.go:45:// maxLongDiscriminator is the largest 12-bit long discriminator.
fabricui/handlers_controller.go:46:const maxLongDiscriminator = 0x0FFF
fabricui/handlers_controller.go:72: "PUT /controller/credentials": true,
fabricui/handlers_controller.go:73: "DELETE /controller/credentials": true,
fabricui/handlers_controller.go:187: case errors.Is(err, ErrUnknownCredential):
fabricui/handlers_controller.go:188: writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such credential", err.Error())
fabricui/handlers_controller.go:189: case errors.Is(err, ErrInvalidArgument), errors.Is(err, ErrInvalidCredential):
fabricui/handlers_controller.go:318: Credentials: s.src.Credentials() != nil,
fabricui/handlers_controller.go:375: Discriminator: optUint(rec.Discriminator),
fabricui/handlers_controller.go:669: // The answer carries the passcode; it is not logged.
fabricui/handlers_controller.go:671: ManualCode: w.ManualCode, QrCode: w.QRCode, Passcode: int(w.Passcode),
fabricui/handlers_controller.go:672: Discriminator: int(w.Discriminator), VendorID: int(w.VendorID), ProductID: int(w.ProductID),
fabricui/handlers_controller.go:939:// group id, a PASE entry's a passcode id: neither names a node), and an
fabricui/handlers_controller.go:1207: raw, err := readOne(ctx, n, AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeCurrentFabricIndex}, true)
fabricui/handlers_controller.go:1227: raw, err := readOne(ctx, n, AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeFabrics}, false)
fabricui/handlers_controller.go:1266: res, err := n.Invoke(ctx, CommandPath{Endpoint: 0, Cluster: clusterOperationalCredentials, Command: commandRemoveFabric}, payload, 0)
fabricui/handlers_controller.go:1300:// passcode, and reports the commissioning on the stream. Neither the code
fabricui/handlers_controller.go:1301:// nor the passcode reaches a log line or an event.
fabricui/handlers_controller.go:1305: return nil, badRequest("the body {code} or {discovered, passcode} is required")
fabricui/handlers_controller.go:1313: case b.Discovered != nil && b.Passcode == nil:
fabricui/handlers_controller.go:1314: return nil, badRequest("discovered needs passcode")
fabricui/handlers_controller.go:1315: case b.Passcode != nil && (*b.Passcode < 1 || *b.Passcode > 99999998):
fabricui/handlers_controller.go:1316: return nil, badRequest("passcode: not 1 to 99999998")
fabricui/handlers_controller.go:1320: case b.Discovered != nil && b.Discovered.LongDiscriminator != nil &&
fabricui/handlers_controller.go:1321: (*b.Discovered.LongDiscriminator < 0 || *b.Discovered.LongDiscriminator > maxLongDiscriminator):
fabricui/handlers_controller.go:1322: return nil, badRequest("discovered.long_discriminator: %d is not 0 to %d", *b.Discovered.LongDiscriminator, maxLongDiscriminator)
fabricui/handlers_controller.go:1351: if b.Discovered.LongDiscriminator != nil {
fabricui/handlers_controller.go:1352: // 0 is a valid discriminator; the port cannot yet tell it from
fabricui/handlers_controller.go:1354: d.LongDiscriminator = uint16(*b.Discovered.LongDiscriminator) //nolint:gosec // range-checked above
fabricui/handlers_controller.go:1356: n, err = s.src.CommissionDiscovered(ctx, d, uint32(*b.Passcode), opts) //nolint:gosec // bounded above
fabricui/handlers_controller.go:1415: LongDiscriminator: int(d.LongDiscriminator), CommissioningMode: int(d.CommissioningMode),
fabricui/handlers_controller.go:1428:// --- credentials ---
fabricui/handlers_controller.go:1430:func (s *controllerServer) credentials() (CredentialStore, error) {
fabricui/handlers_controller.go:1431: c := s.src.Credentials()
fabricui/handlers_controller.go:1433: return nil, serviceUnready("No credential store", "the host stores no Wi-Fi or Thread credentials")
fabricui/handlers_controller.go:1438:func (s *controllerServer) ListCredentials(ctx context.Context, _ api.ListCredentialsRequestObject) (api.ListCredentialsResponseObject, error) {
fabricui/handlers_controller.go:1439: c, err := s.credentials()
fabricui/handlers_controller.go:1447: out := api.ListCredentials200JSONResponse{Credentials: []api.Credential{}}
fabricui/handlers_controller.go:1449: c := api.Credential{Kind: api.CredentialKind(e.Kind), ID: e.ID, Name: e.Name}
fabricui/handlers_controller.go:1453: out.Credentials = append(out.Credentials, c)
fabricui/handlers_controller.go:1458:// SetCredential stores a credential; the passphrase and the dataset are
fabricui/handlers_controller.go:1460:func (s *controllerServer) SetCredential(ctx context.Context, req api.SetCredentialRequestObject) (api.SetCredentialResponseObject, error) {
fabricui/handlers_controller.go:1461: c, err := s.credentials()
fabricui/handlers_controller.go:1473: case api.CredentialRequestKindWifi:
fabricui/handlers_controller.go:1474: if b.Ssid == nil || *b.Ssid == "" {
fabricui/handlers_controller.go:1475: return nil, badRequest("ssid is required for a Wi-Fi credential")
fabricui/handlers_controller.go:1478: if b.Passphrase != nil {
fabricui/handlers_controller.go:1479: pass = *b.Passphrase
fabricui/handlers_controller.go:1481: err = c.SetWiFi(ctx, b.ID, *b.Ssid, pass)
fabricui/handlers_controller.go:1482: case api.CredentialRequestKindThread:
fabricui/handlers_controller.go:1483: if b.DatasetHex == nil || *b.DatasetHex == "" {
fabricui/handlers_controller.go:1484: return nil, badRequest("dataset_hex is required for a Thread dataset")
fabricui/handlers_controller.go:1486: ds, herr := hex.DecodeString(*b.DatasetHex)
fabricui/handlers_controller.go:1488: return nil, badRequest("dataset_hex is not hex")
fabricui/handlers_controller.go:1497: return api.SetCredential204Response{}, nil
fabricui/handlers_controller.go:1500:func (s *controllerServer) RemoveCredential(ctx context.Context, req api.RemoveCredentialRequestObject) (api.RemoveCredentialResponseObject, error) {
fabricui/handlers_controller.go:1501: c, err := s.credentials()
fabricui/handlers_controller.go:1508: if err := c.Remove(ctx, CredentialKind(req.Params.Kind), req.Params.ID); err != nil {
fabricui/handlers_controller.go:1511: return api.RemoveCredential204Response{}, nil
fabricui/handlers_controller.go:1641: Role: optString(n.Role), Ssid: optString(n.SSID), Bssid: optString(n.BSSID),
fabricui/handlers_node.go:59: // the host supplies none. Never the passcode.
fabricui/handlers_node.go:89: Discriminator uint16 `json:"discriminator"`
fabricui/handlers_node.go:147: Discriminator: c.Discriminator, VendorID: c.VendorID, VendorName: c.VendorName,
fabricui/handlers_pairing.go:26:// ephemeralWindow is a window this handler opened on fresh credentials.
fabricui/handlers_pairing.go:27:type ephemeralWindow struct {
fabricui/handlers_pairing.go:28: creds node.WindowCredentials
fabricui/handlers_pairing.go:29: qr string
fabricui/handlers_pairing.go:34:// discriminator and passcode they encode — of the window this handler
fabricui/handlers_pairing.go:35:// opened on fresh credentials, else the configured ones
fabricui/handlers_pairing.go:39: Discriminator *uint16 `json:"discriminator,omitempty"`
fabricui/handlers_pairing.go:40: Passcode *uint32 `json:"passcode,omitempty"`
fabricui/handlers_pairing.go:41: QRCode string `json:"qr_code"`
fabricui/handlers_pairing.go:42: ManualCode string `json:"manual_code"`
fabricui/handlers_pairing.go:57:// whose 503 for a node without commissioning credentials it keeps.
fabricui/handlers_pairing.go:68: if eph := h.activeEphemeral(win); eph != nil {
fabricui/handlers_pairing.go:69: pc, disc := eph.creds.Passcode, eph.creds.Discriminator
fabricui/handlers_pairing.go:71: Discriminator: &disc, Passcode: &pc,
fabricui/handlers_pairing.go:72: QRCode: eph.qr, ManualCode: eph.manual, Source: sourceWindow,
fabricui/handlers_pairing.go:80: unready(w, r, "Pairing codes not available", "the node has no valid commissioning credentials configured")
fabricui/handlers_pairing.go:83: pc, disc := sp.Passcode, sp.Discriminator
fabricui/handlers_pairing.go:85: Discriminator: &disc, Passcode: &pc,
fabricui/handlers_pairing.go:86: QRCode: sp.QR, ManualCode: sp.Manual,
fabricui/handlers_pairing.go:91:// activeEphemeral returns the window this handler opened while the node
fabricui/handlers_pairing.go:93:// discriminator whose deadline has not passed on the node's clock — and
fabricui/handlers_pairing.go:97:func (h *handler) activeEphemeral(win node.WindowSnapshot) *ephemeralWindow {
fabricui/handlers_pairing.go:100: eph := h.ephemeral
fabricui/handlers_pairing.go:104: if !win.Open || win.Kind != node.WindowBasic || win.Discriminator != eph.creds.Discriminator ||
fabricui/handlers_pairing.go:106: h.ephemeral = nil
fabricui/handlers_pairing.go:118:// node through the window, the discriminator and passcode included.
fabricui/handlers_pairing.go:120: Discriminator *uint16 `json:"discriminator,omitempty"`
fabricui/handlers_pairing.go:121: Passcode *uint32 `json:"passcode,omitempty"`
fabricui/handlers_pairing.go:123: QRCode string `json:"qr_code"`
fabricui/handlers_pairing.go:124: ManualCode string `json:"manual_code"`
fabricui/handlers_pairing.go:131:// Retry-After while no endpoint is exposed, and fresh credentials unless
fabricui/handlers_pairing.go:132:// [Options].ConfiguredWindow. The passcode is in the response only — never
fabricui/handlers_pairing.go:184: unready(w, r, title, "the node has no commissioning credentials configured")
fabricui/handlers_pairing.go:198: slog.Int("discriminator", int(*res.Discriminator)), slog.Int("duration_seconds", int(res.DurationSeconds)),
fabricui/handlers_pairing.go:204:// credentials and answers with its setup payload (node.SetupPayload).
fabricui/handlers_pairing.go:214: pc, disc := sp.Passcode, sp.Discriminator
fabricui/handlers_pairing.go:215: return windowResponse{Discriminator: &disc, Passcode: &pc, QRCode: sp.QR, ManualCode: sp.Manual}, nil
fabricui/handlers_pairing.go:218:// openFresh opens the node's own window on credentials drawn for it by
fabricui/handlers_pairing.go:219:// node.NewWindowCredentials: a fresh discriminator, passcode and salt.
fabricui/handlers_pairing.go:220:// Mirrors loom matter_window_adapter.go:143-200 (openEphemeral).
fabricui/handlers_pairing.go:226: qr, manual, err := h.node.pairingCodesFor(creds)
fabricui/handlers_pairing.go:234: h.ephemeral = &ephemeralWindow{creds: creds, qr: qr, manual: manual}
fabricui/handlers_pairing.go:236: pc, disc := creds.Passcode, creds.Discriminator
fabricui/handlers_pairing.go:237: return windowResponse{Discriminator: &disc, Passcode: &pc, QRCode: qr, ManualCode: manual}, nil
fabricui/handlers_pairing.go:250: prev := h.ephemeral
fabricui/handlers_pairing.go:257: // the credentials held before the close are forgotten: a fresh window
fabricui/handlers_pairing.go:259: // window this close did end anyway is forgotten by activeEphemeral.
fabricui/handlers_pairing.go:260: h.forgetEphemeralIf(prev)
fabricui/handlers_pairing.go:277:// forgetEphemeralIf drops the credentials of the window this handler
fabricui/handlers_pairing.go:279:func (h *handler) forgetEphemeralIf(prev *ephemeralWindow) {
fabricui/handlers_pairing.go:281: if h.ephemeral == prev {
fabricui/handlers_pairing.go:282: h.ephemeral = nil
fabricui/memory_controller.go:17:// MemoryCredentials is an in-memory [CredentialStore]: a host without a
fabricui/memory_controller.go:21:type MemoryCredentials struct {
fabricui/memory_controller.go:24: wifi map[string]wifiCredential
fabricui/memory_controller.go:25: thread map[string]threadCredential
fabricui/memory_controller.go:27: defaults map[Credential]bool
fabricui/memory_controller.go:30:type wifiCredential struct{ ssid, passphrase string }
fabricui/memory_controller.go:32:type threadCredential struct {
fabricui/memory_controller.go:34: dataset []byte
fabricui/memory_controller.go:37:// NewMemoryCredentials returns an empty store. threadName reads a Thread
fabricui/memory_controller.go:38:// dataset's network name and refuses a dataset no device could join
fabricui/memory_controller.go:39:// (gofabric.ThreadDatasetName, over go-fabric's thread); nil
fabricui/memory_controller.go:40:// names a dataset by its id and accepts any non-empty one.
fabricui/memory_controller.go:41:func NewMemoryCredentials(threadName func(dataset []byte) (string, error)) *MemoryCredentials {
fabricui/memory_controller.go:42: return &MemoryCredentials{threadName: threadName, wifi: map[string]wifiCredential{}, thread: map[string]threadCredential{}, defaults: map[Credential]bool{}}
fabricui/memory_controller.go:46:// Credential.Default List reports while the entry is stored: a test seam
fabricui/memory_controller.go:47:// for a host's own default (contract 0.7.1, Credential.default).
fabricui/memory_controller.go:48:func (m *MemoryCredentials) MarkDefault(kind CredentialKind, id string) {
fabricui/memory_controller.go:51: m.defaults[Credential{Kind: kind, ID: id}] = true
fabricui/memory_controller.go:54:// List implements [CredentialStore]: Wi-Fi first, each kind by id.
fabricui/memory_controller.go:55:func (m *MemoryCredentials) List(context.Context) ([]Credential, error) {
fabricui/memory_controller.go:58: out := make([]Credential, 0, len(m.wifi)+len(m.thread))
fabricui/memory_controller.go:60: out = append(out, Credential{Kind: CredentialWiFi, ID: id, Name: m.wifi[id].ssid, Default: m.defaults[Credential{Kind: CredentialWiFi, ID: id}]})
fabricui/memory_controller.go:63: out = append(out, Credential{Kind: CredentialThread, ID: id, Name: m.thread[id].name, Default: m.defaults[Credential{Kind: CredentialThread, ID: id}]})
fabricui/memory_controller.go:68:// SetWiFi implements [CredentialStore].
fabricui/memory_controller.go:69:func (m *MemoryCredentials) SetWiFi(_ context.Context, id, ssid, passphrase string) error {
fabricui/memory_controller.go:70: if id == "" || ssid == "" {
fabricui/memory_controller.go:71: return fmt.Errorf("%w: a Wi-Fi credential needs an id and an SSID", ErrInvalidCredential)
fabricui/memory_controller.go:75: m.wifi[id] = wifiCredential{ssid: ssid, passphrase: passphrase}
fabricui/memory_controller.go:79:// SetThread implements [CredentialStore].
fabricui/memory_controller.go:80:func (m *MemoryCredentials) SetThread(_ context.Context, id string, dataset []byte) error {
fabricui/memory_controller.go:81: if id == "" || len(dataset) == 0 {
fabricui/memory_controller.go:82: return fmt.Errorf("%w: a Thread dataset needs an id and a dataset", ErrInvalidCredential)
fabricui/memory_controller.go:86: n, err := m.threadName(dataset)
fabricui/memory_controller.go:94: m.thread[id] = threadCredential{name: name, dataset: slices.Clone(dataset)}
fabricui/memory_controller.go:98:// Remove implements [CredentialStore].
fabricui/memory_controller.go:99:func (m *MemoryCredentials) Remove(_ context.Context, kind CredentialKind, id string) error {
fabricui/memory_controller.go:103: case CredentialWiFi:
fabricui/memory_controller.go:108: case CredentialThread:
fabricui/memory_controller.go:114: return fmt.Errorf("%w: %s %q", ErrUnknownCredential, kind, id)
fabricui/memory_controller.go:117:// WiFi is the stored Wi-Fi credential id names, for a commissioning.
fabricui/memory_controller.go:118:func (m *MemoryCredentials) WiFi(id string) (ssid, passphrase string, ok bool) {
fabricui/memory_controller.go:122: return c.ssid, c.passphrase, ok
fabricui/memory_controller.go:125:// Thread is the stored Thread dataset id names, for a commissioning.
fabricui/memory_controller.go:126:func (m *MemoryCredentials) Thread(id string) (dataset []byte, ok bool) {
fabricui/memory_controller.go:130: return slices.Clone(c.dataset), ok
fabricui/memory_controller.go:134:// codes of the Matter specification's onboarding chapter (passcode
fabricui/memory_controller.go:135:// 20202021, discriminator 3840), which no real device uses.
fabricui/memory_controller.go:137: memoryPasscode = 20202021
fabricui/memory_controller.go:138: memoryManualCode = "34970112332"
fabricui/memory_controller.go:139: memoryQRCode = "MT:Y.K9042C00KA0648G00"
fabricui/memory_controller.go:151:// discovery finds, a credential store with the Thread entry mesh as its
fabricui/memory_controller.go:167: // request names a credential and network_only is off (contract
fabricui/memory_controller.go:184: creds *MemoryCredentials
fabricui/memory_controller.go:206: creds: NewMemoryCredentials(nil),
fabricui/memory_controller.go:211: _ = m.creds.SetWiFi(context.Background(), "home", "HomeNet", "secret-passphrase")
fabricui/memory_controller.go:213: m.creds.MarkDefault(CredentialThread, "mesh")
fabricui/memory_controller.go:260:// or QR code commissions it once; another code that reads as a code finds
fabricui/memory_controller.go:266: return nil, fmt.Errorf("%w: not a QR payload or a manual pairing code", ErrInvalidArgument)
fabricui/memory_controller.go:274: if code != memoryManualCode && code != memoryQRCode {
fabricui/memory_controller.go:292:// has: a central, network_only and a named credential, in the
fabricui/memory_controller.go:375: LongDiscriminator: memoryLongDisc, VendorID: 0xFFF1, ProductID: 0x8002, DeviceType: 0x010A,
fabricui/memory_controller.go:383: case "long_discriminator":
fabricui/memory_controller.go:384: if f.Value != uint32(d.LongDiscriminator) {
fabricui/memory_controller.go:387: case "short_discriminator":
fabricui/memory_controller.go:388: if f.Value != uint32(d.LongDiscriminator>>8) {
fabricui/memory_controller.go:406:func (m *MemoryController) CommissionDiscovered(_ context.Context, d Discovered, passcode uint32, o PairOptions) (ControllerNode, error) {
fabricui/memory_controller.go:412: if passcode != memoryPasscode {
fabricui/memory_controller.go:415: return nil, fmt.Errorf("%w: PASE failed: the device refused the passcode", ErrNodeFailed)
fabricui/memory_controller.go:441:// Credentials implements [ControllerSource].
fabricui/memory_controller.go:442:func (m *MemoryController) Credentials() CredentialStore { return m.creds }
fabricui/memory_controller.go:658: s.values[AttributePath{0, clusterOperationalCredentials, attributeCurrentFabricIndex}] = json.RawMessage(`1`)
fabricui/memory_controller.go:659: s.values[AttributePath{0, clusterOperationalCredentials, attributeFabrics}] = json.RawMessage(fmt.Sprintf(
fabricui/memory_controller.go:682: {ID: clusterOperationalCredentials, Name: "OperationalCredentials", Revision: 2},
fabricui/memory_controller.go:759: DiscoveredAt: memoryTime, CommissionedAt: memoryTime, Discriminator: ptr(uint16(memoryLongDisc)),
fabricui/memory_controller.go:1050: case p.Endpoint == 0 && p.Cluster == clusterOperationalCredentials && p.Command == commandRemoveFabric:
fabricui/memory_controller.go:1058: path := AttributePath{0, clusterOperationalCredentials, attributeFabrics}
fabricui/memory_controller.go:1094: ManualCode: memoryManualCode, QRCode: memoryQRCode, Passcode: memoryPasscode, Discriminator: memoryLongDisc,
fabricui/nodeswitch.go:104:func (s *NodeSwitch) OpenBasicWindow(ctx context.Context, creds node.WindowCredentials) error {
fabricui/nodeswitch.go:112:func (s *NodeSwitch) PairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
fabricui/operations.go:26:// passcode.
fabricui/options.go:36: OpenBasicWindow(ctx context.Context, creds node.WindowCredentials) error
fabricui/options.go:38: PairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
fabricui/options.go:147:// carries no passcode — that stays on the pairing page behind admin
fabricui/options.go:163: Discriminator uint16
fabricui/options.go:256: // a handler logs carries a passcode.
fabricui/options.go:259: // window with its configured passcode and discriminator
fabricui/options.go:261: // fresh random credentials (node.OpenBasicWindow), which no earlier
fabricui/ports.go:25: openBasicWindow(ctx context.Context, creds node.WindowCredentials) error
fabricui/ports.go:29: pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error)
fabricui/ports.go:87:func (a nodeAdapter) openBasicWindow(ctx context.Context, creds node.WindowCredentials) error {
fabricui/ports.go:101:func (a nodeAdapter) pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
fabricui/registry.go:20:// its label. Everything else — key, port, discriminator, passcode, store
fabricui/registry.go:39:// family and not [DefaultNodeKey] — and its port, discriminator, passcode
fabricui/registry.go:48: // keeps its state (a window's credentials, its stream's subscribers)
```

### secrets: web

```
ui/src/lib/api/client.ts:107:export type Credential = S["Credential"];
ui/src/lib/api/client.ts:108:export type CredentialList = S["CredentialList"];
ui/src/lib/api/client.ts:109:export type CredentialRequest = S["CredentialRequest"];
ui/src/lib/api/client.ts:143:export type DiscoverFilter = { filter?: "any" | "long_discriminator" | "short_discriminator" | "vendor" | "device_type"; value?: number; timeout?: number };
ui/src/lib/api/client.ts:272: credentials(): Promise<CredentialList>;
ui/src/lib/api/client.ts:273: setCredential(req: CredentialRequest): Promise<void>;
ui/src/lib/api/client.ts:274: removeCredential(kind: Credential["kind"], id: string): Promise<void>;
ui/src/lib/api/client.ts:323: credentials: "same-origin",
ui/src/lib/api/client.ts:484: credentials: () => call("GET", "/controller/credentials"),
ui/src/lib/api/client.ts:485: setCredential: (req) => call("PUT", "/controller/credentials", { admin: true, body: req }),
ui/src/lib/api/client.ts:486: removeCredential: (kind, id) =>
ui/src/lib/api/client.ts:487: call("DELETE", `/controller/credentials?${new URLSearchParams({ kind, id }).toString()}`, { admin: true }),
ui/src/lib/api/stream.ts:92:// (credentials: "same-origin"; UR-P1-web-116).
ui/src/lib/controller-store.svelte.ts:2:// the credentials and the update queue, and the progress of a
ui/src/lib/controller-store.svelte.ts:10: type Credential,
ui/src/lib/controller-store.svelte.ts:53: let credentials = $state<Credential[]>([]);
ui/src/lib/controller-store.svelte.ts:54: let credentialsError = $state<string | null>(null);
ui/src/lib/controller-store.svelte.ts:80: let credentialsGen = 0;
ui/src/lib/controller-store.svelte.ts:117: async function loadCredentials() {
ui/src/lib/controller-store.svelte.ts:118: const mine = ++credentialsGen;
ui/src/lib/controller-store.svelte.ts:119: credentialsError = null;
ui/src/lib/controller-store.svelte.ts:121: const list = (await client.credentials()).credentials;
ui/src/lib/controller-store.svelte.ts:122: if (mine === credentialsGen) credentials = list;
ui/src/lib/controller-store.svelte.ts:124: if (mine === credentialsGen) credentialsError = message(err);
ui/src/lib/controller-store.svelte.ts:271: get credentials() { return credentials; },
ui/src/lib/controller-store.svelte.ts:272: get credentialsError() { return credentialsError; },
ui/src/lib/controller-store.svelte.ts:285: loadCredentials,
ui/src/lib/pages/controller/Commission.svelte:2: discovery found and its passcode; the stored credential that puts a
ui/src/lib/pages/controller/Commission.svelte:5: the decoder is a chunk loaded on demand). The code and the passcode go
ui/src/lib/pages/controller/Commission.svelte:11: explicit input: the host's default credential, or the only one, is
ui/src/lib/pages/controller/Commission.svelte:19: the code or passcode is held for that retry only. -->
ui/src/lib/pages/controller/Commission.svelte:57: let passcodes = $state<Record<string, string>>({});
ui/src/lib/pages/controller/Commission.svelte:61: if (controller.status?.capabilities.credentials) void controller.loadCredentials().then(preselect);
ui/src/lib/pages/controller/Commission.svelte:65: // picked already; "No network credential" stays a choice.
ui/src/lib/pages/controller/Commission.svelte:68: const list = controller.credentials;
ui/src/lib/pages/controller/Commission.svelte:75: ...controller.credentials.map((c) => ({
ui/src/lib/pages/controller/Commission.svelte:77: label: `${t(`controller.credential.kind_${c.kind}`)} · ${c.name}${c.default ? ` (${t("controller.credential.default")})` : ""}`,
ui/src/lib/pages/controller/Commission.svelte:80: const bleHint = $derived(!!controller.status?.capabilities.ble && network === "" && !networkOnly && controller.credentials.length > 0);
ui/src/lib/pages/controller/Commission.svelte:181: const pin = Number((passcodes[d.instance_name] ?? "").trim());
ui/src/lib/pages/controller/Commission.svelte:183: toast.error(t("controller.commission.passcode_invalid"));
ui/src/lib/pages/controller/Commission.svelte:190: await commission({ discovered: { instance_name: d.instance_name, long_discriminator: d.long_discriminator }, passcode: pin });
ui/src/lib/pages/controller/Commission.svelte:192: // passcode, a timeout, terms required) it stays for the retry.
ui/src/lib/pages/controller/Commission.svelte:194: passcodes = { ...passcodes, [d.instance_name]: "" };
ui/src/lib/pages/controller/Commission.svelte:201: { key: "discriminator", label: t("controller.commission.col_discriminator"), numeric: true, get: (d) => d.long_discriminator },
ui/src/lib/pages/controller/Commission.svelte:306: {:else if col.key === "discriminator"}
ui/src/lib/pages/controller/Commission.svelte:307: <span class="font-mono text-xs">{d.long_discriminator}</span>
ui/src/lib/pages/controller/Commission.svelte:316: aria-label={t("controller.commission.passcode")}
ui/src/lib/pages/controller/Commission.svelte:317: placeholder={t("controller.commission.passcode")}
ui/src/lib/pages/controller/Commission.svelte:318: value={passcodes[d.instance_name] ?? ""}
ui/src/lib/pages/controller/Commission.svelte:319: oninput={(e) => (passcodes = { ...passcodes, [d.instance_name]: (e.currentTarget as HTMLInputElement).value })}
ui/src/lib/pages/controller/Settings.svelte:2: Thread credentials (names only; a secret is entered once and never
ui/src/lib/pages/controller/Settings.svelte:8: import type { Credential, CredentialRequest } from "../../api/client";
ui/src/lib/pages/controller/Settings.svelte:22: let kind = $state<Credential["kind"]>("wifi");
ui/src/lib/pages/controller/Settings.svelte:24: let ssid = $state("");
ui/src/lib/pages/controller/Settings.svelte:31: if (controller.status?.capabilities.credentials) void controller.loadCredentials();
ui/src/lib/pages/controller/Settings.svelte:47: async function addCredential() {
ui/src/lib/pages/controller/Settings.svelte:48: const req: CredentialRequest = { kind, id: id.trim() };
ui/src/lib/pages/controller/Settings.svelte:50: req.ssid = ssid.trim();
ui/src/lib/pages/controller/Settings.svelte:51: req.passphrase = secret;
ui/src/lib/pages/controller/Settings.svelte:53: // An operational dataset is TLV bytes: hex digits, two per byte.
ui/src/lib/pages/controller/Settings.svelte:56: toast.error(t("controller.settings.dataset_invalid"));
ui/src/lib/pages/controller/Settings.svelte:59: req.dataset_hex = hex;
ui/src/lib/pages/controller/Settings.svelte:63: await client.setCredential(req);
ui/src/lib/pages/controller/Settings.svelte:64: toast.success(t("controller.settings.credential_saved"));
ui/src/lib/pages/controller/Settings.svelte:65: id = ssid = secret = "";
ui/src/lib/pages/controller/Settings.svelte:66: await controller.loadCredentials();
ui/src/lib/pages/controller/Settings.svelte:74: async function removeCredential(c: Credential) {
ui/src/lib/pages/controller/Settings.svelte:75: const ok = await confirm({ title: t("controller.settings.credential_remove"), body: c.name, destructive: true, confirmLabel: t("common.remove") });
ui/src/lib/pages/controller/Settings.svelte:78: await client.removeCredential(c.kind, c.id);
ui/src/lib/pages/controller/Settings.svelte:79: await controller.loadCredentials();
ui/src/lib/pages/controller/Settings.svelte:86: { value: "wifi", label: t("controller.credential.kind_wifi") },
ui/src/lib/pages/controller/Settings.svelte:87: { value: "thread", label: t("controller.credential.kind_thread") },
ui/src/lib/pages/controller/Settings.svelte:90: const columns: DataColumn<Credential>[] = $derived([
ui/src/lib/pages/controller/Settings.svelte:113: <Card class="p-4" data-testid="credentials">
ui/src/lib/pages/controller/Settings.svelte:114: <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.settings.credentials")}</h3>
ui/src/lib/pages/controller/Settings.svelte:115: {#if s && !s.capabilities.credentials}
ui/src/lib/pages/controller/Settings.svelte:118: {#if controller.credentialsError}
ui/src/lib/pages/controller/Settings.svelte:119: <ErrorState message={controller.credentialsError} onRetry={() => void controller.loadCredentials()} />
ui/src/lib/pages/controller/Settings.svelte:121: <DataTable rows={controller.credentials} {columns} rowKey={(c) => `${c.kind}:${c.id}`} emptyMessage={t("controller.settings.credentials_empty")}>
ui/src/lib/pages/controller/Settings.svelte:126: <Badge variant="muted">{t(`controller.credential.kind_${c.kind}`)}</Badge>
ui/src/lib/pages/controller/Settings.svelte:130: <Button size="sm" variant="ghost" class="text-[var(--fui-error-text)]" onclick={() => void removeCredential(c)}>{t("common.remove")}</Button>
ui/src/lib/pages/controller/Settings.svelte:135: <form class="mt-4 flex flex-wrap items-center gap-2" onsubmit={(e) => { e.preventDefault(); void addCredential(); }}>
ui/src/lib/pages/controller/Settings.svelte:141: // A passphrase is never sent as a dataset, nor the other way.
ui/src/lib/pages/controller/Settings.svelte:142: if (v !== kind) ssid = secret = "";
ui/src/lib/pages/controller/Settings.svelte:149: <Input class="w-40" bind:value={ssid} placeholder={t("controller.settings.ssid")} aria-label={t("controller.settings.ssid")} />
ui/src/lib/pages/controller/Settings.svelte:150: <Input class="w-48" type="password" autocomplete="new-password" bind:value={secret} placeholder={t("controller.settings.passphrase")} aria-label={t("controller.settings.pa
ui/src/lib/pages/controller/Settings.svelte:152: <Input class="w-72 font-mono" type="password" autocomplete="off" bind:value={secret} placeholder={t("controller.settings.dataset")} aria-label={t("controller.settings.data
ui/src/lib/pages/controller/Settings.svelte:154: <Button type="submit" disabled={busy || !id.trim()}>{t("controller.settings.credential_add")}</Button>
ui/src/lib/pages/controller/Settings.svelte:171: {#each (["ota", "credentials", "ble", "diagnostics"] as const).filter((k) => s.capabilities[k]) as k (k)}<Badge variant="muted">{t(`controller.settings.cap.${k}`)}</Badge>
ui/src/lib/pages/controller/node/Window.svelte:3: passcode stays on this page; it is not stored. The codes leave the
ui/src/lib/pages/controller/node/Window.svelte:49: const qrSvg = $derived(win ? renderQrSvg(win.qr_code) : null);
ui/src/lib/pages/controller/node/Window.svelte:104: <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.manual_code")}</dt>
ui/src/lib/pages/controller/node/Window.svelte:105: <dd class="font-mono">{win.manual_code} <Button size="sm" variant="ghost" onclick={() => void copy(win!.manual_code)}>{t("common.copy")}</Button></dd>
ui/src/lib/pages/controller/node/Window.svelte:106: <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.qr_code")}</dt>
ui/src/lib/pages/controller/node/Window.svelte:107: <dd class="font-mono text-xs">{win.qr_code}</dd>
ui/src/lib/pages/controller/node/Window.svelte:108: <dt class="text-[var(--ha-secondary-text-color)]">{t("controller.window.discriminator")}</dt>
ui/src/lib/pages/controller/node/Window.svelte:109: <dd>{win.discriminator}</dd>
ui/src/lib/pages/node/NodeConfig.svelte:3: from, the label, the address, the discriminator, vendor and product
ui/src/lib/pages/node/NodeConfig.svelte:4: and the store; never the passcode, which stays on the Pair page. A
ui/src/lib/pages/node/NodeConfig.svelte:23: [t("node.config.discriminator"), String(config.discriminator)],
ui/src/lib/pages/node/Pair.svelte:101: const qrSvg = $derived(win?.qr_code ? renderQrSvg(win.qr_code) : null);
ui/src/lib/pages/node/Pair.svelte:178: <p class="mb-1 text-xs text-[var(--ha-secondary-text-color)]">{t("node.pair.manual_code")}</p>
ui/src/lib/pages/node/Pair.svelte:181: {win.manual_code}
ui/src/lib/pages/node/Pair.svelte:183: <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_manual_code")} onclick={() => void copy(win.manual_code)}>
ui/src/lib/pages/node/Pair.svelte:191: <p class="break-all font-mono text-xs text-[var(--ha-secondary-text-color)]">{win.qr_code}</p>
ui/src/lib/pages/node/Pair.svelte:192: <Button variant="ghost" size="sm" aria-label={t("node.pair.copy_qr_payload")} onclick={() => void copy(win.qr_code)}>
ui/src/lib/qr.ts:3:// assets/ui/src/lib/qr.ts:1-95 at b40b289a (qrcode, MIT): the matrix from
ui/src/lib/qr.ts:4:// QRCode.create, one horizontal path segment per run of dark modules.
ui/src/lib/qr.ts:7:import { create as qrCreate } from "qrcode";
ui/src/lib/scan/index.ts:22: const formats = (await ctor.getSupportedFormats?.()) ?? ["qr_code"];
ui/src/lib/scan/index.ts:23: return formats.includes("qr_code") ? new ctor({ formats: ["qr_code"] }) : null;
ui/src/lib/store.svelte.ts:332: discriminator: p.discriminator,
ui/src/lib/store.svelte.ts:333: passcode: p.passcode,
ui/src/lib/store.svelte.ts:335: qr_code: p.qr_code,
ui/src/lib/store.svelte.ts:336: manual_code: p.manual_code,
ui/src/lib/theme/ha-bridge.ts:183: if (skin === "ha") root.dataset.skin = "ha";
ui/src/lib/theme/ha-bridge.ts:184: else delete root.dataset.skin;
ui/src/shell/index.html:19: root.dataset.skin = "ha";
```

## 3. Writes and the adapter (X2)

### handler → port calls

```
fabricui/handler.go:641: unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
fabricui/handlers_controller.go:1336: s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("started"), Step: pairingStepOf(st)}})
fabricui/handlers_controller.go:1342: s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("started"), BLE: &bleDecision{Used: d.Used, Reason: string(d.Reason)}}})
fabricui/handlers_controller.go:1344: s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("started")}})
fabricui/handlers_controller.go:1359: s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("failed"), Detail: ptr(progressDetail(err))}})
fabricui/handlers_controller.go:1362: s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("succeeded"), NodeIDHex: ptr(hex16(n.ID()))}})
fabricui/handlers_controller.go:1715: unsubs := []func(){h.ctrlHub.Subscribe(fn)}
fabricui/handlers_controller.go:1717: unsubs = append(unsubs, h.ctrl.Observe(func(e ControllerEvent) { fn(controllerEvent(e)) }))
fabricui/handlers_controller.go:1718: if h.ctrl.Capabilities().Diagnostics {
fabricui/handlers_controller.go:1719: if u := h.ctrl.OnDiagnosticEvent(func(e DiagnosticEvent) { fn(controllerTraceEvent(e)) }); u != nil {
fabricui/handlers_exposure.go:95: s, err := h.exposure.Schema(r.Context())
fabricui/handlers_exposure.go:122: items, err := h.exposure.List(r.Context())
fabricui/handlers_exposure.go:171: items, err := h.exposure.Preview(r.Context(), updates)
fabricui/handlers_exposure.go:202: applied, err := h.exposure.Apply(r.Context(), updates)
fabricui/handlers_node.go:207: if items, err := h.exposure.List(r.Context()); err == nil {
fabricui/operations.go:167: out := logLevelResponse{Level: levelName(h.ops.LogLevel()), Levels: make([]string, 0, len(logLevels))}
fabricui/operations.go:185: in := h.ops.Info()
fabricui/operations.go:218: if err := h.ops.SetLogLevel(l); err != nil {
fabricui/operations.go:237: recs := h.ops.Log(since, limit)
fabricui/operations.go:283: return h.ops.ObserveLog(func(rec LogRecord) {
fabricui/registry.go:140: list := h.registry.List()
fabricui/registry.go:177: if h.regGen != gen && !slices.ContainsFunc(h.registry.List(), func(x NodeEntry) bool { return x.Key == e.Key }) {
fabricui/registry.go:289: e, err := h.registry.Create(r.Context(), NodeSpec{Label: label})
fabricui/registry.go:319: if err := h.registry.Rename(r.Context(), key, label); err != nil {
fabricui/registry.go:334: if err := h.registry.Delete(r.Context(), key); err != nil {
fabricui/registry.go:353: h.nodesHub.Publish(Event{Type: EventNodesChanged, Data: NodesChangedEvent{Keys: keys}})
```

### contexts

```
fabricui/gofabric/gofabric.go:120: s.ctx, s.cancel = context.WithCancel(context.Background())
fabricui/gofabric/gofabric.go:670: ctx, cancel := context.WithTimeout(s.ctx, labelPushTimeout)
fabricui/gofabric/node.go:711: rctx, cancel := context.WithTimeout(ctx, pingTimeout)
fabricui/handler.go:401: req, err := http.NewRequestWithContext(context.Background(), method, path, http.NoBody)
fabricui/handlers_controller.go:196: h.log.WarnContext(r.Context(), "fabricui.controller.node_failed", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
fabricui/handlers_controller.go:1329: ctx, cancel = context.WithTimeout(ctx, time.Duration(*b.TimeoutSeconds)*time.Second)
fabricui/handlers_diagnostics.go:178: h.node.reannounce(r.Context())
fabricui/handlers_diagnostics.go:179: h.log.InfoContext(r.Context(), "fabricui.mdns.reannounced")
fabricui/handlers_diagnostics.go:343: if infos, err := h.node.fabrics(r.Context()); err == nil {
fabricui/handlers_diagnostics.go:356: h.log.WarnContext(r.Context(), "fabricui.compatibility.fabrics_failed", slog.String("error", err.Error()))
fabricui/handlers_exposure.go:95: s, err := h.exposure.Schema(r.Context())
fabricui/handlers_exposure.go:122: items, err := h.exposure.List(r.Context())
fabricui/handlers_exposure.go:171: items, err := h.exposure.Preview(r.Context(), updates)
fabricui/handlers_exposure.go:202: applied, err := h.exposure.Apply(r.Context(), updates)
fabricui/handlers_exposure.go:216: ctx, cancel := detached(r.Context())
fabricui/handlers_exposure.go:219: h.log.WarnContext(r.Context(), "fabricui.exposure.reassemble_failed", slog.String("error", err.Error()))
fabricui/handlers_exposure.go:224: h.log.InfoContext(r.Context(), "fabricui.exposure.applied", slog.Int("applied", applied))
fabricui/handlers_node.go:157: res.Warnings = h.warnings(r.Context())
fabricui/handlers_node.go:165: res.Counters = h.counters(r.Context())
fabricui/handlers_node.go:199: if fabrics, err := h.node.fabrics(r.Context()); err == nil {
fabricui/handlers_node.go:204: h.log.WarnContext(r.Context(), "fabricui.status.fabrics_failed", slog.String("error", err.Error()))
fabricui/handlers_node.go:207: if items, err := h.exposure.List(r.Context()); err == nil {
fabricui/handlers_node.go:248: infos, err := h.node.fabrics(r.Context())
fabricui/handlers_node.go:285: if err := h.node.removeFabric(r.Context(), uint8(idx)); err != nil {
fabricui/handlers_node.go:290: h.log.InfoContext(r.Context(), "fabricui.fabric.removed", slog.Int("fabric_index", int(idx)))
fabricui/handlers_node.go:306: ctx, cancel := detached(r.Context())
fabricui/handlers_node.go:343: infos, err := h.node.fabrics(r.Context())
fabricui/handlers_node.go:351: ctx, cancel := detached(r.Context())
fabricui/handlers_node.go:358: h.log.WarnContext(r.Context(), "fabricui.factory_reset.partial",
fabricui/handlers_node.go:365: h.log.InfoContext(r.Context(), "fabricui.factory_reset.done", slog.Int("removed", len(removed)))
fabricui/handlers_node.go:400: if err := erase(r.Context()); err != nil {
fabricui/handlers_node.go:404: h.log.InfoContext(r.Context(), "fabricui.erase.done")
fabricui/handlers_node.go:510: d, err := h.node.endpointDetail(r.Context(), ep)
fabricui/handlers_node.go:616: return context.WithTimeout(context.WithoutCancel(ctx), detachedTimeout)
fabricui/handlers_pairing.go:79: h.log.WarnContext(r.Context(), "fabricui.pairing.codes_unavailable", slog.String("error", err.Error()))
fabricui/handlers_pairing.go:180: h.log.WarnContext(r.Context(), "fabricui.window.blocked", slog.String("error", err.Error()))
fabricui/handlers_pairing.go:197: h.log.InfoContext(r.Context(), "fabricui.window.opened",
fabricui/handlers_pairing.go:206: if err := h.node.openOwnWindow(r.Context(), timeout); err != nil {
fabricui/handlers_pairing.go:211: _ = h.node.closeWindow(r.Context())
fabricui/handlers_pairing.go:230: if err := h.node.openBasicWindow(r.Context(), creds); err != nil {
fabricui/handlers_pairing.go:252: if err := h.node.closeWindow(r.Context()); err != nil {
fabricui/handlers_topology.go:100: fabrics, err := h.node.fabrics(r.Context())
fabricui/memory_controller.go:211: _ = m.creds.SetWiFi(context.Background(), "home", "HomeNet", "secret-passphrase")
fabricui/memory_controller.go:212: _ = m.creds.SetThread(context.Background(), "mesh", []byte{0x03, 0x04, 'M', 'e', 's', 'h'})
fabricui/operations.go:222: h.log.InfoContext(r.Context(), "fabricui.operation.log_level", "level", levelName(l))
fabricui/operations.go:282: h.serveSSE(r.Context(), w, func(fn func(Event)) func() {
fabricui/problem.go:105: if errors.Is(err, context.Canceled) && r.Context().Err() != nil {
fabricui/problem.go:108: h.log.DebugContext(r.Context(), "fabricui.request.abandoned", slog.String("error", err.Error()),
fabricui/problem.go:113: h.log.ErrorContext(r.Context(), title, slog.String("error", err.Error()),
fabricui/registry.go:289: e, err := h.registry.Create(r.Context(), NodeSpec{Label: label})
fabricui/registry.go:302: h.log.InfoContext(r.Context(), "fabricui.registry.created", slog.String("key", e.Key))
fabricui/registry.go:319: if err := h.registry.Rename(r.Context(), key, label); err != nil {
fabricui/registry.go:334: if err := h.registry.Delete(r.Context(), key); err != nil {
fabricui/registry.go:345: h.log.InfoContext(r.Context(), "fabricui.registry.deleted", slog.String("key", key))
fabricui/stream.go:31: h.serveSSE(r.Context(), w, h.subscribe)
```

## 4. Life cycle (X2)

### goroutines, observers, locks, channels

```
fabricui/controller.go:56: Observe(fn func(ControllerEvent)) (unsubscribe func())
fabricui/controller.go:276: SetAutoSubscribe(ctx context.Context, on bool) error
fabricui/controller.go:741:// name verbatim, plus go-fabric's Discovery 2.1, PASE 2.2 and Subscribe
fabricui/events.go:188:// own: Subscribe registers fn for every event from now on and returns the
fabricui/events.go:191: Subscribe(fn func(Event)) (unsubscribe func())
fabricui/events.go:199: mu sync.Mutex
fabricui/events.go:207:// Subscribe implements [EventSource].
fabricui/events.go:208:func (e *Events) Subscribe(fn func(Event)) (unsubscribe func()) {
fabricui/events.go:259:// unobserve := n.Observe(fabricui.NodeObservers(push, n.Window().Open))
fabricui/events.go:264: var mu sync.Mutex
fabricui/gofabric/gofabric.go:147: mu sync.Mutex
fabricui/gofabric/gofabric.go:163: labelMu sync.Mutex
fabricui/gofabric/gofabric.go:166: pushes sync.WaitGroup
fabricui/gofabric/gofabric.go:529: defer s.mu.Unlock()
fabricui/gofabric/gofabric.go:586: AutoSubscribe: po.AutoSubscribe,
fabricui/gofabric/gofabric.go:649: defer s.labelMu.Unlock()
fabricui/gofabric/gofabric.go:659: go s.pushRounds() //nolint:contextcheck // the push outlives the request by design
fabricui/gofabric/gofabric.go:674: cancel()
fabricui/gofabric/gofabric.go:702: s.cancel()
fabricui/gofabric/gofabric.go:723:func (s *source) Observe(fn func(fabricui.ControllerEvent)) (unsubscribe func()) {
fabricui/gofabric/gofabric.go:724: return s.events.Subscribe(func(e fabricui.Event) {
fabricui/gofabric/hooks.go:33: mu sync.Mutex
fabricui/gofabric/hooks.go:46: running sync.WaitGroup
fabricui/gofabric/hooks.go:121: defer h.mu.Unlock()
fabricui/gofabric/hooks.go:170: defer h.mu.Unlock()
fabricui/gofabric/node.go:712: defer cancel()
fabricui/gofabric/node.go:736: sem := make(chan struct{}, pingParallel)
fabricui/gofabric/node.go:737: var wg sync.WaitGroup
fabricui/gofabric/node.go:799:func (w *node) SetAutoSubscribe(ctx context.Context, on bool) error {
fabricui/gofabric/node.go:800: return nodeError(w.n.SetAutoSubscribe(ctx, on))
fabricui/gofabric/topology.go:89: topoMu sync.Mutex
fabricui/gofabric/updates.go:68: unobserve := m.Observe(func(e softwareupdate.UpdateEvent) { s.publish(updateEvent(&e)) })
fabricui/handler.go:82: regMu sync.Mutex
fabricui/handler.go:90: done chan struct{}
fabricui/handler.go:95: mu sync.Mutex
fabricui/handler.go:202: // interface nil check and panic on Subscribe (UR-P1-err-03).
fabricui/handler.go:525: s.done = make(chan struct{})
fabricui/handler.go:526: s.stop = sync.OnceFunc(func() { close(s.done) })
fabricui/handler.go:639: unsubs := []func(){h.hub.Subscribe(fn)}
fabricui/handler.go:641: unsubs = append(unsubs, h.nodesHub.Subscribe(fn))
fabricui/handler.go:644: unsubs = append(unsubs, h.extra.Subscribe(fn))
fabricui/handlers_controller.go:636:func (s *controllerServer) SetNodeAutoSubscribe(ctx context.Context, req api.SetNodeAutoSubscribeRequestObject) (api.SetNodeAutoSubscribeResponseObject, error) {
fabricui/handlers_controller.go:644: if err := n.SetAutoSubscribe(ctx, req.Body.Enabled); err != nil {
fabricui/handlers_controller.go:647: return api.SetNodeAutoSubscribe204Response{}, nil
fabricui/handlers_controller.go:1330: defer cancel()
fabricui/handlers_controller.go:1715: unsubs := []func(){h.ctrlHub.Subscribe(fn)}
fabricui/handlers_controller.go:1717: unsubs = append(unsubs, h.ctrl.Observe(func(e ControllerEvent) { fn(controllerEvent(e)) }))
fabricui/handlers_exposure.go:217: defer cancel()
fabricui/handlers_node.go:307: defer cancel()
fabricui/handlers_node.go:352: defer cancel()
fabricui/handlers_pairing.go:99: defer h.mu.Unlock()
fabricui/memory_controller.go:22: mu sync.Mutex
fabricui/memory_controller.go:50: defer m.mu.Unlock()
fabricui/memory_controller.go:57: defer m.mu.Unlock()
fabricui/memory_controller.go:74: defer m.mu.Unlock()
fabricui/memory_controller.go:93: defer m.mu.Unlock()
fabricui/memory_controller.go:101: defer m.mu.Unlock()
fabricui/memory_controller.go:120: defer m.mu.Unlock()
fabricui/memory_controller.go:128: defer m.mu.Unlock()
fabricui/memory_controller.go:153:// commissioning reports the steps Discovery 2.1, PASE 2.2 and Subscribe
fabricui/memory_controller.go:181: mu sync.Mutex
fabricui/memory_controller.go:231: defer m.mu.Unlock()
fabricui/memory_controller.go:243: defer m.mu.Unlock()
fabricui/memory_controller.go:254: defer m.mu.Unlock()
fabricui/memory_controller.go:284:// StepSubscribe; controller/commissioner/progress.go:78-80 at v0.18.0).
fabricui/memory_controller.go:288: memoryStepSubscribe = PairingStep{Number: 21, Sub: 1, Name: "Subscribe"}
fabricui/memory_controller.go:340:// and Subscribe.
fabricui/memory_controller.go:362: report(memoryStepSubscribe, "started")
fabricui/memory_controller.go:363: report(memoryStepSubscribe, "succeeded")
fabricui/memory_controller.go:448:func (m *MemoryController) Observe(fn func(ControllerEvent)) (unsubscribe func()) {
fabricui/memory_controller.go:449: return m.subs.Subscribe(func(e Event) {
fabricui/memory_controller.go:568: defer u.m.mu.Unlock()
fabricui/memory_controller.go:597: defer u.m.mu.Unlock()
fabricui/memory_controller.go:649: mu sync.Mutex
fabricui/memory_controller.go:776: defer n.mu.Unlock()
fabricui/memory_controller.go:784: defer n.mu.Unlock()
fabricui/memory_controller.go:792: defer n.mu.Unlock()
fabricui/memory_controller.go:798: defer n.mu.Unlock()
fabricui/memory_controller.go:807: defer n.mu.Unlock()
fabricui/memory_controller.go:813: defer n.mu.Unlock()
fabricui/memory_controller.go:819: defer n.mu.Unlock()
fabricui/memory_controller.go:858: defer n.mu.Unlock()
fabricui/memory_controller.go:890: defer n.mu.Unlock()
fabricui/memory_controller.go:907: defer n.mu.Unlock()
fabricui/memory_controller.go:958: defer n.mu.Unlock()
fabricui/memory_controller.go:1163:func (n *memoryNode) SetAutoSubscribe(_ context.Context, on bool) error {
fabricui/memory_controller.go:1173: defer n.mu.Unlock()
fabricui/memory_controller.go:1182: defer n.mu.Unlock()
fabricui/memory_controller.go:1195: defer n.mu.Unlock()
fabricui/nodeswitch.go:128:func (s *NodeSwitch) Observe(o node.Observers) (unsubscribe func()) { return s.Current().Observe(o) }
fabricui/operations.go:303: mu sync.Mutex
fabricui/operations.go:361: defer m.mu.Unlock()
fabricui/operations.go:376: defer m.mu.Unlock()
fabricui/options.go:42: Observe(o node.Observers) (unsubscribe func())
fabricui/ports.go:113:func (a nodeAdapter) observe(o node.Observers) (unsubscribe func()) { return a.n.Observe(o) }
fabricui/registry.go:175: defer h.regMu.Unlock()
fabricui/stream.go:62: ch := make(chan Event, streamBuffer)
```

## 5. Browser side (X4)

### client transport

```
ui/src/lib/api/client.ts:146:export const TOKEN_HEADER = "X-Fabric-UI-Token";
ui/src/lib/api/client.ts:174: readonly baseUrl: string;
ui/src/lib/api/client.ts:175: /** Sets the admin token sent on admin routes; status() sets it from csrf_token. */
ui/src/lib/api/client.ts:236: /** The controller's status; it sets the admin token from csrf_token, as status() does. 404 when the host serves no controller. */
ui/src/lib/api/client.ts:272: credentials(): Promise<CredentialList>;
ui/src/lib/api/client.ts:294: * A client of the node routes under `baseUrl` (the mount prefix, e.g.
ui/src/lib/api/client.ts:298:export function createClient(baseUrl: string, fetchImpl: Fetch = (i, init) => fetch(i, init)): FabricClient {
ui/src/lib/api/client.ts:299: const base = baseUrl.replace(/\/+$/, "");
ui/src/lib/api/client.ts:313: const headers: Record<string, string> = { Accept: "application/json" };
ui/src/lib/api/client.ts:314: if (opts.body !== undefined) headers["Content-Type"] = "application/json";
ui/src/lib/api/client.ts:316: if (token) headers[TOKEN_HEADER] = token;
ui/src/lib/api/client.ts:321: headers,
ui/src/lib/api/client.ts:323: credentials: "same-origin",
ui/src/lib/api/client.ts:360: const s = await call<{ csrf_token?: string }>("GET", p);
ui/src/lib/api/client.ts:361: if (s.csrf_token) auth.token = s.csrf_token;
ui/src/lib/api/client.ts:389: baseUrl: base,
ui/src/lib/api/client.ts:397: if (s.csrf_token) auth.token = s.csrf_token;
ui/src/lib/api/client.ts:439: if (s.csrf_token) auth.token = s.csrf_token;
ui/src/lib/api/client.ts:484: credentials: () => call("GET", "/controller/credentials"),
ui/src/lib/api/client.ts:485: setCredential: (req) => call("PUT", "/controller/credentials", { admin: true, body: req }),
ui/src/lib/api/client.ts:487: call("DELETE", `/controller/credentials?${new URLSearchParams({ kind, id }).toString()}`, { admin: true }),
ui/src/lib/api/stream.ts:55:/** The part of EventSource the helper uses, so a test can hand in a fake. */
ui/src/lib/api/stream.ts:56:export interface EventSourceLike {
ui/src/lib/api/stream.ts:64:export type EventSourceFactory = (url: string) => EventSourceLike;
ui/src/lib/api/stream.ts:92:// (credentials: "same-origin"; UR-P1-web-116).
ui/src/lib/api/stream.ts:93:const defaultOpen: EventSourceFactory = (u) => new EventSource(u);
ui/src/lib/api/stream.ts:107: open: EventSourceFactory,
ui/src/lib/api/stream.ts:111: let es: EventSourceLike | null = null;
ui/src/lib/api/stream.ts:205: open: EventSourceFactory = defaultOpen,
ui/src/lib/api/stream.ts:249: open: EventSourceFactory = defaultOpen,
ui/src/lib/api/stream.ts:271: open: EventSourceFactory = defaultOpen,
ui/src/lib/context.ts:121: /** Subscribes to the log stream; the default opens an EventSource on the client's log stream URL. */
ui/src/lib/controller-store.svelte.ts:2:// the credentials and the update queue, and the progress of a
ui/src/lib/controller-store.svelte.ts:36: /** Subscribes to the controller stream, as NodeStoreDeps.subscribe; the default opens an EventSource on the client's URL. */
ui/src/lib/controller-store.svelte.ts:53: let credentials = $state<Credential[]>([]);
ui/src/lib/controller-store.svelte.ts:54: let credentialsError = $state<string | null>(null);
ui/src/lib/controller-store.svelte.ts:80: let credentialsGen = 0;
ui/src/lib/controller-store.svelte.ts:118: const mine = ++credentialsGen;
ui/src/lib/controller-store.svelte.ts:119: credentialsError = null;
ui/src/lib/controller-store.svelte.ts:121: const list = (await client.credentials()).credentials;
ui/src/lib/controller-store.svelte.ts:122: if (mine === credentialsGen) credentials = list;
ui/src/lib/controller-store.svelte.ts:124: if (mine === credentialsGen) credentialsError = message(err);
ui/src/lib/controller-store.svelte.ts:271: get credentials() { return credentials; },
ui/src/lib/controller-store.svelte.ts:272: get credentialsError() { return credentialsError; },
ui/src/lib/kit/DataTable.svelte:2: at b40b289a: sortable headers, an optional search box, the sort
ui/src/lib/pages/controller/Commission.svelte:61: if (controller.status?.capabilities.credentials) void controller.loadCredentials().then(preselect);
ui/src/lib/pages/controller/Commission.svelte:68: const list = controller.credentials;
ui/src/lib/pages/controller/Commission.svelte:75: ...controller.credentials.map((c) => ({
ui/src/lib/pages/controller/Commission.svelte:80: const bleHint = $derived(!!controller.status?.capabilities.ble && network === "" && !networkOnly && controller.credentials.length > 0);
ui/src/lib/pages/controller/Settings.svelte:2: Thread credentials (names only; a secret is entered once and never
ui/src/lib/pages/controller/Settings.svelte:31: if (controller.status?.capabilities.credentials) void controller.loadCredentials();
ui/src/lib/pages/controller/Settings.svelte:113: <Card class="p-4" data-testid="credentials">
ui/src/lib/pages/controller/Settings.svelte:114: <h3 class="mb-2 font-medium text-[var(--ha-primary-text-color)]">{t("controller.settings.credentials")}</h3>
ui/src/lib/pages/controller/Settings.svelte:115: {#if s && !s.capabilities.credentials}
ui/src/lib/pages/controller/Settings.svelte:118: {#if controller.credentialsError}
ui/src/lib/pages/controller/Settings.svelte:119: <ErrorState message={controller.credentialsError} onRetry={() => void controller.loadCredentials()} />
ui/src/lib/pages/controller/Settings.svelte:121: <DataTable rows={controller.credentials} {columns} rowKey={(c) => `${c.kind}:${c.id}`} emptyMessage={t("controller.settings.credentials_empty")}>
ui/src/lib/pages/controller/Settings.svelte:171: {#each (["ota", "credentials", "ble", "diagnostics"] as const).filter((k) => s.capabilities[k]) as k (k)}<Badge variant="muted">{t(`controller.settings.cap.${k}`)}</Badge>
ui/src/lib/pages/node/Diagnostics.svelte:27: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/Fabrics.svelte:30: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/NodeManage.svelte:27: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/node/Selection.svelte:38: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
ui/src/lib/pages/operation/Operation.svelte:27: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true || !!controller.status?.csrf_token);
ui/src/lib/store.svelte.ts:49: * Subscribes to the event stream; the default opens an EventSource on
ui/src/lib/store.svelte.ts:78:export function isAdminStatus(s: Pick<Status, "csrf_token" | "admin"> | null | undefined): boolean {
ui/src/lib/store.svelte.ts:79: return !!s?.csrf_token || s?.admin === true;
ui/src/shell/App.svelte:112: const isAdmin = $derived(!!node.status?.csrf_token || node.status?.admin === true);
```

### sinks and frames

```
ui/src/lib/kit/DataTable.svelte:4: the stored value is read with a shape check, since localStorage is
ui/src/lib/kit/DataTable.svelte:41: /** When set, the sort persists under this localStorage key. */
ui/src/lib/kit/DataTable.svelte:65: const raw = localStorage.getItem("fabric-ui.table:" + persistKey);
ui/src/lib/kit/DataTable.svelte:89: localStorage.setItem("fabric-ui.table:" + persistKey, JSON.stringify(snapshot));
ui/src/lib/kit/Graph.svelte:29: let { graph, hrefOf = () => undefined, follow = (href) => window.location.assign(href), label, class: className }: Props = $props();
ui/src/lib/kit/Tabs.svelte:78: <a href={item.href} aria-current={isActive ? "page" : undefined} class={itemClass(isActive)} onclick={() => onSelect?.(item.key)}>
ui/src/lib/kit/utils.ts:47:/** Reads a string from localStorage, `fallback` when absent or unavailable. */
ui/src/lib/kit/utils.ts:50: return localStorage.getItem(key) ?? fallback;
ui/src/lib/kit/utils.ts:56:/** Best-effort write to localStorage. */
ui/src/lib/kit/utils.ts:59: localStorage.setItem(key, value);
ui/src/lib/pages/controller/Commission.svelte:266: <a class="ml-2 font-medium text-[var(--fui-primary-text)] hover:underline" href={routes.controllerNode + result.node.node_id_hex}>{nodeName(result.node)}</a>
ui/src/lib/pages/controller/Diagnostics.svelte:133: <a class="font-mono text-[var(--fui-primary-text)] hover:underline" href={routes.controllerNode + s.peer_node_id_hex}>0x{s.peer_node_id_hex}</a>
ui/src/lib/pages/controller/Diagnostics.svelte:149: <a class="text-[var(--fui-primary-text)] hover:underline" href={routes.controllerNode + p.node_id_hex} data-testid="peer-link">{peerName(p)}</a>
ui/src/lib/pages/controller/Node.svelte:84: <a href={routes.controllerNodes} class="text-sm text-[var(--fui-primary-text)] hover:underline">← {t("controller.node.back")}</a>
ui/src/lib/pages/controller/Nodes.svelte:76: location.hash = routes.controllerNode + n.node_id_hex;
ui/src/lib/pages/controller/Nodes.svelte:124: <a href={routes.controllerCommission} class="ml-auto text-sm font-medium text-[var(--fui-primary-text)] hover:underline">{t("controller.nodes.commission")}</a>
ui/src/lib/pages/controller/Nodes.svelte:166: <a class="inline-flex items-center px-2 text-sm text-[var(--fui-primary-text)] hover:underline" href={`${routes.controllerNode}${n.node_id_hex}?tab=window`}>{t("controller.no
ui/src/lib/pages/controller/node/Window.svelte:102: <div class="w-48">{@html qrSvg}</div>
ui/src/lib/pages/node/EcosystemHints.svelte:48: <a class="text-[var(--fui-primary-text)] underline" href={link.href} target="_blank" rel="noopener noreferrer">{t(`node.pair.hints.${hint.ecosystem}.${link.key}`)}</a>
ui/src/lib/pages/node/Endpoint.svelte:150: <a href={routes.endpoints} class="text-sm text-[var(--fui-primary-text)] hover:underline">← {t("node.endpoint.back")}</a>
ui/src/lib/pages/node/Endpoint.svelte:175: <span>{t("node.endpoint.parent")} <a class="text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + d.parent_endpoint_id}>#{d.parent_endpoint_id}</a></span>
ui/src/lib/pages/node/Endpoint.svelte:181: <a class="ml-1 text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + p}>#{p}</a>
ui/src/lib/pages/node/Endpoints.svelte:101: <a class="rounded border border-[var(--ha-divider-color)] px-3 py-1 text-sm text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + s.number}>
ui/src/lib/pages/node/Endpoints.svelte:116: <a class="font-medium text-[var(--fui-primary-text)] hover:underline" href={routes.endpoint + ep.endpoint_id}>{ep.name}</a>
ui/src/lib/pages/node/Fabrics.svelte:255: <a href={routes.pair} data-testid="share-link" class="mt-3 inline-flex items-center text-sm font-medium text-[var(--fui-primary-text)] hover:underline">
ui/src/lib/pages/node/NodeManage.svelte:106: <a class="mt-1 inline-block text-sm text-[var(--fui-primary-text)] underline" href={HINT_SOURCE} target="_blank" rel="noopener noreferrer">
ui/src/lib/pages/node/Pair.svelte:174: <div class="h-auto max-w-full">{@html qrSvg}</div>
ui/src/lib/theme/ha-bridge.ts:57: if (window.self === window.top) return false;
ui/src/lib/theme/ha-bridge.ts:61: if (window.location.pathname.includes(HA_INGRESS_SEGMENT)) return true;
ui/src/lib/theme/ha-bridge.ts:63: return window.parent.document.querySelector("home-assistant") !== null;
ui/src/lib/theme/ha-bridge.ts:136: const parentWin = window.parent;
ui/src/shell/App.svelte:76: let route = $state<Route>(routeOf(location.hash));
ui/src/shell/App.svelte:77: const onHash = () => (route = routeOf(location.hash));
ui/src/shell/App.svelte:131: location.hash = key === defaultKey ? ui.routes.status : routesForNode(ui.routes, key).status;
ui/src/shell/App.svelte:136: location.hash = routesForNode(ui.routes, created.key).selection;
ui/src/shell/App.svelte:143: location.hash = ui.routes.status;
ui/src/shell/index.html:55: if (window.self !== window.top) {
ui/src/shell/index.html:56: var parentWin = window.parent;
ui/src/shell/index.html:58: if (location.pathname.indexOf("/api/hassio_ingress/") !== -1 || parentDoc.querySelector("home-assistant")) {
```
