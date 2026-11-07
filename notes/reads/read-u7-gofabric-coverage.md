# Read U7: go-fabric coverage in go-fabric-ui

Inventory only, no judgement. Sources: go-fabric `git describe --tags` = v0.18.0; go-fabric-ui `git describe --tags` = v0.6.0; go-fabric-ui `go.mod:6` requires `github.com/SukramJ/go-fabric v0.17.0` (not v0.18.0). Paths in Part 2/3 are relative to go-fabric-ui; Part 1 paths relative to go-fabric. Doc column = first line of the doc comment block directly above, as written; `(none)` = no comment directly above.

## Part 1 - go-fabric observable surface

| package | symbol | file:line | doc comment (first line) |
|---|---|---|---|
| node | Node.AttestationError | node/attestation.go:58 | AttestationError is the load error of [Attestation].Files under |
| node | Node.Peers | node/client.go:172 | Peers returns the client stack's peer set — the nodes of our fabrics |
| node | Node.Bindings | node/client.go:183 | Bindings returns the binding manager: it resolves the entries of the |
| node | Node.CommissionerControl | node/commissioner_control.go:56 | CommissionerControl returns the CommissionerControl server — nil unless |
| node | Node.OpenBasicWindow | node/commissioning.go:500 | OpenBasicWindow opens the node's own commissioning window with its own |
| node | Node.OpenOwnWindow | node/commissioning.go:525 | OpenOwnWindow opens the node's own commissioning window with the |
| node | Node.CloseWindow | node/commissioning.go:543 | CloseWindow closes the open commissioning window, whoever opened it — |
| node | Node.PairingCodesFor | node/credentials.go:90 | PairingCodesFor returns the QR code payload and the manual pairing code |
| node | Node.DiagnosticEvents | node/diagnostics.go:30 | DiagnosticEvents returns the node's diagnostic trace, newest first: a |
| node | Node.OnDiagnosticEvent | node/diagnostics.go:40 | OnDiagnosticEvent registers fn for every event the node records from |
| node | Node.OnDiagnosticEventAsync | node/diagnostics.go:53 | OnDiagnosticEventAsync is [Node.OnDiagnosticEvent] with fn queued on |
| node | Node.Endpoints | node/endpoints.go:66 | Endpoints returns every endpoint the node serves now — the root first, |
| node | Node.FactoryReset | node/factory_reset.go:45 | FactoryReset returns the node to its factory state and stops it: every |
| node | Node.OnHealthSample | node/health.go:93 | OnHealthSample registers fn for every verdict of the health probe |
| node | Node.OnHealthSampleAsync | node/health.go:100 | OnHealthSampleAsync is [Node.OnHealthSample] with fn queued on the |
| node | Node.Health | node/health.go:109 | Health returns the health probe's last verdict; false before the first |
| node | Node.ICD | node/icd.go:262 | ICD returns the node's ICD handle; nil for a node that is no ICD. |
| node | Node.SetupPayload | node/inspect.go:42 | SetupPayload returns the configured setup payload: passcode, |
| node | Node.SessionDiagnostics | node/inspect.go:130 | SessionDiagnostics returns the node's sessions, each with its |
| node | Node.Clock | node/inspect.go:192 | Clock returns the node's time source: [Config].Clock, or [clock.System] |
| node | Node.Topology | node/lifecycle.go:173 | Topology returns the host's view of the bridge as the node serves it |
| node | Node.MDNSEnabled | node/mdns.go:73 | MDNSEnabled reports whether the node's advertiser publishes on the |
| node | Node.MDNSServices | node/mdns.go:77 | MDNSServices returns the records the node's advertiser currently |
| node | Node.DiagnoseMDNS | node/mdns.go:83 | DiagnoseMDNS inspects the records the node currently advertises and |
| node | Node.Start | node/node.go:769 | Start brings the node online: the bridge starts, and the daemon's |
| node | Node.Stop | node/node.go:920 | Stop takes the node offline and releases what New and Start began, in |
| node | Node.Run | node/node.go:1041 | Run starts the node, waits for ctx to end and stops it — matter.js |
| node | Node.Bridge | node/node.go:1053 | Bridge returns the bridge, for a host that needs a port the facade does |
| node | Node.LocalAddr | node/node.go:1058 | LocalAddr returns the bound `host:port` of the node's operational UDP |
| node | Node.Root | node/node.go:1061 | Root returns the root endpoint's servers. |
| node | Node.Aggregator | node/node.go:1065 | Aggregator returns the Aggregator endpoint's bridged-device set — nil |
| node | Node.Sessions | node/node.go:1072 | Sessions returns the secure-session manager, for a host's diagnostics |
| node | Node.Subscriptions | node/node.go:1079 | Subscriptions returns the subscription manager, for a host's |
| node | Node.RemoveFabric | node/node.go:1090 | RemoveFabric removes a fabric without a commissioner — a host's revoke |
| node | Node.SetCommissioningBlocked | node/node.go:1104 | SetCommissioningBlocked refuses every new commissioning while err is |
| node | Node.FabricInfos | node/node.go:1119 | FabricInfos returns the installed fabrics in the store's order (by |
| node | Node.UniqueID | node/node.go:1135 | UniqueID returns the BasicInformation UniqueID the node serves: the |
| node | Node.BootReason | node/node.go:1141 | BootReason returns the GeneralDiagnostics BootReason the node was built |
| node | Node.Store | node/node.go:1144 | Store returns the credential store the node runs on. |
| node | Node.Persistence | node/node.go:1150 | Persistence is the endpoint-level persistence (scene tables, bridged |
| node | Node.Groups | node/node.go:1157 | Groups returns the group state the node runs on — [Config].Groups, or |
| node | Node.Window | node/observers.go:249 | Window returns the commissioning window as the node serves it now — |
| node | Node.Observe | node/observers.go:316 | Observe registers o's hooks on a running (or not yet started) node: |
| node | Node.ObserveAsync | node/observers.go:342 | ObserveAsync is [Node.Observe] with every hook of o queued on the node's |
| node | Node.OnOTAProviderStatus | node/ota_provider.go:113 | OnOTAProviderStatus registers fn for every update state a requestor of |
| node | Node.OnOTAProviderStatusAsync | node/ota_provider.go:123 | OnOTAProviderStatusAsync is [Node.OnOTAProviderStatus] with fn queued on |
| node | Node.OTAProvider | node/ota_provider.go:136 | OTAProvider returns the OTA Software Update Provider — nil unless |
| node | Node.OTAProviderEndpoint | node/ota_provider.go:140 | OTAProviderEndpoint is the provider endpoint's number, false without a |
| node | Node.BDX | node/ota_provider.go:151 | BDX is the BDX responder the provider serves its images through — the |
| node | Node.TriggerReannounce | node/reannounce.go:22 | TriggerReannounce announces the node's DNS-SD records now — for a host |
| node | Node.ReadAttribute | node/served.go:61 | ReadAttribute returns the value one attribute of the node serves, read |
| node | Node.ReadCluster | node/served.go:85 | ReadCluster returns every attribute one cluster of an endpoint serves, |
| node | Node.Events | node/served.go:176 | Events returns the events the node's event log holds that match q, |
| node | Node.SessionParameters | node/sessions.go:110 | SessionParameters are the session parameters the node reports to every |
| node | Node.Surface | node/surface.go:396 | Surface reports the node's surface: the digest of [Config].Surface, the |
| controller | Controller.Discover | controller/code.go:106 | Discover searches DNS-SD for the commissionable devices id selects |
| controller | Controller.DiscoverWith | controller/code.go:130 | DiscoverWith is [Controller.Discover] with options: with |
| controller | Controller.Close | controller/controller.go:474 | Close tears the stack down in the reverse order of [New]: the update |
| controller | Controller.Registry | controller/controller.go:542 | Registry is the registry of the paired nodes: Commissioned lists them, |
| controller | Controller.Nodes | controller/controller.go:545 | Nodes are the commissioned nodes, in the registry's order. |
| controller | Controller.Fabric | controller/controller.go:549 | Fabric is our fabric: its ID, our node ID, the root of trust. A seam of |
| controller | Controller.Peers | controller/controller.go:557 | Peers is the peer set the registry and the commissioner reach nodes |
| controller | Controller.Clock | controller/controller.go:562 | Clock is the time source of the controller's stack ([Config.Clock], the |
| controller | Controller.TCP | controller/controller.go:566 | TCP is the controller's TCP configuration ([Config.TCP]): whether it |
| controller | Controller.Addr | controller/controller.go:571 | Addr is the controller's UDP address: where its sessions, its |
| controller | Controller.Commissioner | controller/controller.go:575 | Commissioner is the commissioner the registry pairs through. A seam of |
| controller | Controller.Pair | controller/controller.go:658 | Pair finds the commissionable device a QR code ("MT:…") or a manual |
| controller | Controller.Erase | controller/erase.go:30 | Erase is the controller's factory reset: it closes the controller and |
| controller | Controller.Groups | controller/groups.go:83 | Groups is the controller's group manager: the key sets it holds for the |
| controller | Controller.SetGroupKey | controller/groups.go:89 | SetGroupKey keeps a key set in the controller's own tables and maps |
| controller | Controller.SoftwareUpdates | controller/ota.go:124 | SoftwareUpdates is the update manager — matter.js's SoftwareUpdateManager |
| controller | Controller.BDX | controller/ota.go:128 | BDX is the provider's BDX responder, the transfers a node downloads an |
| controller | Controller.CameraRequestor | controller/serve.go:276 | CameraRequestor is the WebRtcTransportRequestor the controller serves |
| controller | Controller.OnCameraSignal | controller/serve.go:287 | OnCameraSignal calls fn for every command a camera invokes on the |
| controller | Controller.TCPAddr | controller/tcp.go:94 | TCPAddr is the controller's TCP listener address, nil without |
| controller | Controller.VendorInfo | controller/vendorinfo.go:41 | VendorInfo is the controller's vendor information service, built on the |
| node | Observers.OnFabricAdded | node/observers.go:43 | OnFabricAdded runs after an AddNOC installed a fabric and the |
| node | Observers.OnFabricRemoved | node/observers.go:53 | OnFabricRemoved runs after RemoveFabric (or the fail-safe rollback |
| node | Observers.OnFabricUpdated | node/observers.go:63 | OnFabricUpdated runs after an UpdateNOC (or the fail-safe rollback |
| node | Observers.OnCommissioningComplete | node/observers.go:71 | OnCommissioningComplete runs after a successful |
| node | Observers.OnWindowTransition | node/observers.go:76 | OnWindowTransition runs on every commissioning-window transition |
| node | Observers.OnSurfaceChanged | node/observers.go:86 | OnSurfaceChanged runs after an assembly changed an endpoint's served |
| node | Observers.OnSessionOpened | node/observers.go:93 | OnSessionOpened runs after a CASE session was established — by a |
| node | Observers.OnSessionClosed | node/observers.go:97 | OnSessionClosed runs after a session OnSessionOpened reported closed, |
| node | Observers.OnSubscriptionsChanged | node/observers.go:102 | OnSubscriptionsChanged runs after a subscription was added to or |
| node | Observers.OnAttributeChanged | node/observers.go:117 | OnAttributeChanged runs after a cluster server on any endpoint — |
| node | Observers.OnEvent | node/observers.go:127 | OnEvent runs after a cluster server on any endpoint emitted an |
| node | HealthSample.Healthy | bridge/health_probe.go:42 | (none) |
| node | HealthSample.Note | bridge/health_probe.go:45 | Note is the stable, English machine string. The tracker's scoring |
| node | DiagnosticEvent.At | diagevent/ring.go:59 | (none) |
| node | DiagnosticEvent.Kind | diagevent/ring.go:60 | (none) |
| node | DiagnosticEvent.Severity | diagevent/ring.go:61 | (none) |
| node | DiagnosticEvent.Message | diagevent/ring.go:64 | Message is one sentence an operator can act on. It is not a log |
| node | DiagnosticEvent.Detail | diagevent/ring.go:67 | Detail carries the identifiers that make the message specific — |
| node | SessionDiagnostics.AsOf | node/inspect.go:64 | AsOf is when the snapshot was taken, on the node's clock |
| node | SessionDiagnostics.Sessions | node/inspect.go:66 | Sessions are the open sessions, by session id. |
| node | SessionDiagnostics.Occupancy | node/inspect.go:68 | Occupancy is the session table's id usage. |
| node | MDNSFinding.Severity | mdns/diagnose.go:27 | (none) |
| node | MDNSFinding.Code | mdns/diagnose.go:30 | Code is a stable machine-readable identifier; the message is |
| node | MDNSFinding.Message | mdns/diagnose.go:31 | (none) |
| node | MDNSFinding.Service | mdns/diagnose.go:34 | Service names the advertised record the finding applies to, empty |
| surface | Report.Policies | surface/policy.go:295 | (none) |
| surface | Report.Level | surface/policy.go:298 | Level is the specification level the surface was composed at: the |
| surface | Report.LevelOrigin | surface/policy.go:301 | LevelOrigin names where Level comes from ([Policies.LevelOrigin]): |
| surface | Report.Profiles | surface/policy.go:304 | Profiles are the selected controller profiles ([Policies].Profiles), |
| surface | Report.Extensions | surface/policy.go:308 | Extensions are the enabled "Ahead of matter.js" register IDs |
| surface | Report.Effective | surface/policy.go:309 | (none) |
| surface | Report.Decisions | surface/policy.go:311 | Decisions has at least one row per candidate element ([Evaluate]). |
| surface | Report.Withheld | surface/policy.go:314 | Withheld are the elements the level policy withheld its endpoint |
| surface | Report.Assembled | surface/policy.go:321 | Assembled reports that Effective was read from the assembled |
| surface | Report.Changed | surface/policy.go:324 | Changed are the surface changes the reporter announced in this |

Type locations: `node.DiagnosticEvent` = `diagevent.Event` (node/facade.go:64, diagevent/ring.go:58); `node.MDNSFinding` = `mdns.Finding` (node/mdns.go:30, mdns/diagnose.go:26); `node.HealthSample` = `bridge.HealthSample` (node/health.go:16, bridge/health_probe.go:41); `node.Observers` node/observers.go:38; `node.SessionDiagnostics` node/inspect.go:61; `surface.Report` surface/policy.go:294. Rows: 111 (methods Node 55, Controller 21; fields Observers 11, HealthSample 2, DiagnosticEvent 5, SessionDiagnostics 3, MDNSFinding 4, Report 10).

## Part 2 - go-fabric symbols go-fabric-ui reads (non-test Go; comment-only mentions excluded)

Method calls on `*node.Node` go through `NodeSource` (fabricui/options.go:25-73) and are in Part 3. Feed = route of the handler named in the enclosing-function column (route table fabricui/handler.go:274-297 for /node and /nodes/{node}; controller routes contract/openapi.yaml:997-1694).

| symbol | first use | enclosing | feeds | uses |
|---|---|---|---|---|
| schema.DeviceTypeName | fabricui/devicetype_names.go:54 | deviceTypeName | device_type_name: GET /node/endpoints (handlers_diagnostics.go:225), GET /node/exposure (handlers_exposure.go:74), GET /node/topology (handlers_topology.go:221), GET /node/endpoints/{ep} (node_detail.go:140) | 3; also gofabric/node.go:198, node_detail.go:302 |
| schema.DeviceTypeID | fabricui/devicetype_names.go:54 | deviceTypeName | same as schema.DeviceTypeName | 1 |
| eligibility.Verdict | fabricui/provider.go:100 | type Verdict | ExposureProvider items: GET/PUT /node/exposure, POST /node/exposure/preview | 1 |
| softwareupdate.Manager | fabricui/gofabric/updates.go:21 | type updates | GET /controller/updates, POST /controller/updates/check | 1 |
| softwareupdate.Target | fabricui/gofabric/updates.go:64 | Consent | POST /controller/updates/consent | 1 |
| topology.Topology | fabricui/gofabric/topology.go:31 | NetworkTopology | GET /controller/topology | 3; also gofabric/topology.go:61, gofabric/topology.go:95 |
| topology.Service | fabricui/gofabric/topology.go:60 | watchTopology | topology_changed event on GET /controller/stream (gofabric/topology.go:62) | 2; also gofabric/gofabric.go:78 |
| topology.KindMatter | fabricui/gofabric/topology.go:101 | networkTopologyOf | GET /controller/topology graph | 3; also gofabric/topology.go:120, gofabric/topology.go:138 |
| topology.KindThreadUnknown | fabricui/gofabric/topology.go:123 | networkTopologyOf | GET /controller/topology graph | 1 |
| topology.KindWiFiAP | fabricui/gofabric/topology.go:125 | networkTopologyOf | GET /controller/topology graph | 1 |
| topology.Node | fabricui/gofabric/topology.go:137 | graphID | GET /controller/topology graph | 4; also gofabric/topology.go:148, gofabric/topology.go:191, gofabric/topology.go:205 |
| topology.RoleRouter | fabricui/gofabric/topology.go:193 | threadUnknownOf | GET /controller/topology graph | 1 |
| topology.Connection | fabricui/gofabric/topology.go:220 | linkOf | GET /controller/topology graph | 1 |
| topology.StrengthUnknown | fabricui/gofabric/topology.go:223 | linkOf | GET /controller/topology graph | 1 |
| topology.NetworkThread | fabricui/gofabric/topology.go:225 | linkOf | GET /controller/topology graph | 1 |
| controller.PairOptions | fabricui/gofabric/gofabric.go:69 | type ControllerOptions | POST /controller/commission (template ControllerOptions.Pair) | 1 |
| thread.DecodeDataset | fabricui/gofabric/gofabric.go:87 | ThreadDatasetName | exported helper gofabric.ThreadDatasetName; no caller in non-test code (memory_controller.go:37 comment) | 1 |
| controller.Controller | fabricui/gofabric/gofabric.go:98 | ControllerOf | adapter over the whole /controller/* family | 2; also gofabric/gofabric.go:114 |
| commissioner.WiFiNetwork | fabricui/gofabric/gofabric.go:224 | networkOf | POST /controller/commission (networkOf) | 3; also gofabric/gofabric.go:231, gofabric/gofabric.go:238 |
| commissioner.ThreadNetwork | fabricui/gofabric/gofabric.go:224 | networkOf | POST /controller/commission (networkOf) | 3; also gofabric/gofabric.go:232, gofabric/gofabric.go:245 |
| controller.ParsePairingCode | fabricui/gofabric/gofabric.go:260 | Pair | POST /controller/commission | 1 |
| controller.ErrPairingCode | fabricui/gofabric/gofabric.go:287 | commissionError | problem mapping of POST /controller/commission | 1 |
| controller.ErrConfiguration | fabricui/gofabric/gofabric.go:287 | commissionError | problem mapping of POST /controller/commission | 1 |
| controller.ErrNoDevice | fabricui/gofabric/gofabric.go:289 | commissionError | problem mapping of POST /controller/commission | 1 |
| discovery.Identifier | fabricui/gofabric/gofabric.go:297 | identifierOf | GET /controller/discovery (identifierOf) | 5; also gofabric/gofabric.go:303, gofabric/gofabric.go:308, gofabric/gofabric.go:313, gofabric/gofabric.go:319 |
| discovery.AnyCommissionable | fabricui/gofabric/gofabric.go:300 | identifierOf | GET /controller/discovery | 1 |
| discovery.ByLongDiscriminator | fabricui/gofabric/gofabric.go:305 | identifierOf | GET /controller/discovery; POST /controller/commission (rediscover, gofabric.go:405) | 2; also gofabric/gofabric.go:405 |
| discovery.ByShortDiscriminator | fabricui/gofabric/gofabric.go:310 | identifierOf | GET /controller/discovery | 1 |
| discovery.ByVendor | fabricui/gofabric/gofabric.go:315 | identifierOf | GET /controller/discovery | 1 |
| discovery.ByDeviceType | fabricui/gofabric/gofabric.go:317 | identifierOf | GET /controller/discovery | 1 |
| discovery.ParseTxtDecimal | fabricui/gofabric/gofabric.go:358 | discoveredOf | GET /controller/discovery result fields (discoveredOf) | 2; also gofabric/gofabric.go:361 |
| commissioner.Options | fabricui/gofabric/gofabric.go:429 | CommissionDiscovered | POST /controller/commission (CommissionDiscovered) | 1 |
| commissioner.ErrLabel | fabricui/gofabric/gofabric.go:482 | SetFabricLabel | PUT /controller/fabric/label error mapping | 1 |
| schema.DeviceTypeIDOf | fabricui/gofabric/node.go:197 | deviceTypeName | device types on GET /controller/nodes/{id}/endpoints/{ep} (gofabric/node.go:197) and GET /node/endpoints/{ep} (node_detail.go:301) | 2; also node_detail.go:301 |
| schema.ClusterName | fabricui/gofabric/node.go:211 | clusterName | cluster names: GET /controller/nodes/{id}/endpoints/{ep}[/clusters/{hex}], GET /node/endpoints/{ep} | 2; also node_detail.go:180 |
| schema.ClusterFeatures | fabricui/gofabric/node.go:348 | clusterDetail | features: /controller/nodes/{id}/endpoints/{ep}/clusters/{hex}; /node/endpoints/{ep} | 2; also node_detail.go:192 |
| schema.ClusterFeature | fabricui/gofabric/node.go:351 | clusterDetail | /controller/nodes/{id}/endpoints/{ep}/clusters/{hex} | 1 |
| schema.EventName | fabricui/gofabric/node.go:428 | Events | event names: GET /controller/nodes/{id}/events; GET /node/endpoints/{ep}/events (and endpoint detail) | 3; also node_detail.go:222, node_detail.go:318 |
| node.WindowCredentials | fabricui/handler.go:55 | type handler | GET /node/pairing, POST /node/pairing/window | 8; also options.go:35, options.go:37, ports.go:25, ports.go:29, ports.go:80 (+2) |
| node.NewWindowCredentials | fabricui/handler.go:240 | newHandler | GET /node/pairing, POST /node/pairing/window | 1 |
| node.HealthSample | fabricui/options.go:28 | type NodeSource | GET /node/status (handlers_node.go:156); stream health_sample (events.go:120) | 6; also options.go:71, ports.go:52, ports.go:158, events.go:120, events.go:141 |
| node.FabricInfo | fabricui/options.go:30 | type NodeSource | GET /node/fabrics; GET /node/topology; stream fabric_added/fabric_removed | 8; also ports.go:21, ports.go:66, events.go:247, events.go:250, events.go:303 (+2) |
| node.WindowSnapshot | fabricui/options.go:34 | type NodeSource | GET /node/status, GET /node/pairing, POST /node/pairing/window; stream window_opened/window_closed | 6; also ports.go:24, ports.go:78, events.go:253, events.go:299, handlers_pairing.go:97 |
| node.SetupPayload | fabricui/options.go:36 | type NodeSource | GET /node/pairing | 3; also ports.go:28, ports.go:92 |
| node.BridgeTopology | fabricui/options.go:38 | type NodeSource | GET /node/topology, GET /node/endpoints, GET /node/endpoints/{ep} | 7; also ports.go:30, ports.go:98, node_detail.go:147, handlers_node.go:448, handlers_topology.go:118 (+1) |
| node.SessionDiagnostics | fabricui/options.go:39 | type NodeSource | GET /node/sessions, GET /node/topology | 4; also ports.go:34, ports.go:102, handlers_topology.go:118 |
| node.Observers | fabricui/options.go:41 | type NodeSource | GET /node/stream (NodeObservers) | 5; also ports.go:38, ports.go:106, events.go:244, events.go:246 |
| node.MDNSService | fabricui/options.go:43 | type NodeSource | GET /node/mdns (type in NodeSource only) | 1 |
| node.MDNSFinding | fabricui/options.go:44 | type NodeSource | GET /node/mdns (type in NodeSource only) | 1 |
| node.DiagnosticEvent | fabricui/options.go:45 | type NodeSource | GET /node/events; stream diagnostic_event | 8; also options.go:72, ports.go:40, ports.go:53, ports.go:150, ports.go:162 (+2) |
| surface.Report | fabricui/options.go:50 | type NodeSource | GET /node/surface; summary in GET /node/status; decisions in GET /node/endpoints/{ep} | 4; also ports.go:42, ports.go:152, handlers_node.go:103 |
| node.AttributeValue | fabricui/options.go:55 | type NodeSource | GET /node/endpoints/{ep} | 11; also options.go:56, node_detail.go:175, node_detail.go:181, node_detail.go:211, node_detail.go:228 (+5) |
| node.EventQuery | fabricui/options.go:57 | type NodeSource | GET /node/endpoints/{ep}/events | 4; also ports.go:49, ports.go:166, handlers_node.go:515 |
| node.Event | fabricui/options.go:57 | type NodeSource | GET /node/endpoints/{ep}/events | 2; also node_detail.go:313 |
| node.EndpointInfo | fabricui/options.go:63 | type NodeSource | GET /node/endpoints, GET /node/endpoints/{ep} | 7; also ports.go:33, ports.go:100, node_detail.go:68, node_detail.go:114, node_detail.go:126 (+1) |
| eligibility.VendorName | fabricui/handlers_controller.go:304 | GetControllerStatus | vendor_name: GET /controller/status (handlers_controller.go:304), /controller/nodes/{id}/fabrics (:1193), GET /controller/discovery (:1346), GET /node/fabrics (handlers_node.go:231), GET /node/topology (handlers_topology.go:253), stream (events.go:307) | 6; also handlers_controller.go:1193, handlers_controller.go:1346, events.go:307, handlers_node.go:231, handlers_topology.go:253 |
| surface.SurfaceChange | fabricui/events.go:266 | NodeObservers | stream surface_changed | 2; also events.go:294 |
| node.SessionInfo | fabricui/events.go:269 | NodeObservers | GET /node/sessions, GET /node/topology; stream session_opened/closed, subscriptions_changed | 5; also events.go:272, events.go:275, handlers_topology.go:280, handlers_diagnostics.go:38 |
| node.AttributeChange | fabricui/events.go:278 | NodeObservers | stream attribute_changed | 2; also events.go:285 |
| node.DeviceTypeInfo | fabricui/node_detail.go:137 | deviceTypesOf | GET /node/endpoints/{ep} | 1 |
| node.BridgedDevice | fabricui/node_detail.go:147 | findDevice | GET /node/topology, GET /node/endpoints, GET /node/endpoints/{ep} | 11; also node_detail.go:148, node_detail.go:148, node_detail.go:149, node_detail.go:149, handlers_topology.go:152 (+5) |
| schema.ClusterRevision | fabricui/node_detail.go:183 | servedCluster | GET /node/endpoints/{ep} | 1 |
| schema.AttributeName | fabricui/node_detail.go:210 | servedCluster | GET /node/endpoints/{ep} | 2; also node_detail.go:233 |
| schema.CommandName | fabricui/node_detail.go:220 | servedCluster | GET /node/endpoints/{ep} | 1 |
| schema.ResponseCommandName | fabricui/node_detail.go:221 | servedCluster | GET /node/endpoints/{ep} | 1 |
| surface.Warn | fabricui/handlers_node.go:114 | summariseSurface | warning count in GET /node/status surface summary | 1 |
| node.ErrNotStarted | fabricui/handlers_node.go:408 | writeAPIError | problem mapping of GET /node/endpoints/{ep}[/events] | 1 |
| node.AggregatorSurfaceKey | fabricui/handlers_node.go:453 | surfaceKeyOf | decisions in GET /node/endpoints/{ep} | 1 |
| node.WindowBasic | fabricui/handlers_pairing.go:104 | activeEphemeral | GET /node/pairing | 1 |
| node.ErrCommissioningWindowAlreadyOpen | fabricui/handlers_pairing.go:176 | openWindow | POST /node/pairing/window problem mapping | 1 |
| node.ErrAdministratorWindowOpen | fabricui/handlers_pairing.go:176 | openWindow | POST /node/pairing/window problem mapping | 1 |
| node.ErrCommissioningBlocked | fabricui/handlers_pairing.go:178 | openWindow | POST /node/pairing/window problem mapping | 1 |
| node.ErrCommissioningWindowNotConfigured | fabricui/handlers_pairing.go:183 | openWindow | POST /node/pairing/window problem mapping | 1 |
| node.ClusterInfo | fabricui/handlers_diagnostics.go:247 | clustersOf | GET /node/endpoints | 1 |
| eligibility.EcosystemForVendor | fabricui/handlers_diagnostics.go:294 | compatibility | GET /node/compatibility | 1 |
| eligibility.Compat | fabricui/handlers_diagnostics.go:306 | compatibility | GET /node/compatibility | 1 |

Rows: 76.

Other go-fabric packages imported by non-test code (outside the Part 2 regex): 
| package | import line | where |
|---|---|---|
| controller/client | fabricui/gofabric/gofabric.go:39; hooks.go:10; node.go:23 | controller adapter: node registry, reads, writes, invoke, ICD, hooks |
| controller/topology | fabricui/gofabric/gofabric.go:41; topology.go:13 | GET /controller/topology |
| controller/softwareupdate | fabricui/gofabric/updates.go:10 | GET /controller/updates |
| controller/commissioner | fabricui/gofabric/gofabric.go:40 | POST /controller/commission |
| client/discovery | fabricui/gofabric/gofabric.go:37 | GET /controller/discovery |
| thread | fabricui/gofabric/gofabric.go:42 | ThreadDatasetName |
| cluster/spec | fabricui/gofabric/node.go:17 | controller attribute/command read and write |
| cluster/spec/{accesscontrol,basicinformation,binding,icdmanagement,operationalcredentials} | fabricui/gofabric/node.go:18-22 | ACL, bindings, ICD, node fabrics routes |
| clock | fabricui/options.go:15 | NodeSource.Clock |
| contract | fabricui/node_detail.go:13 | contract.StatusError, node_detail.go:41 |
| controller (methods) | fabricui/gofabric/gofabric.go:38 | Discover, Registry, Fabric, Pair, SoftwareUpdates (see Part 3) |

## Part 3 - join: Part 1 symbols vs go-fabric-ui reads

Method rows: yes = a call on a `*node.Node` / `*controller.Controller` receiver (node via `NodeSource`, options.go:25-73). Field rows: yes = a selector or composite-literal key on that type. Checked by `grep -rnE` over fabricui and internal, non-test, excluding api.gen.go, and reading each hit.

| package | symbol | read by go-fabric-ui |
|---|---|---|
| node | Node.AttestationError | no |
| node | Node.Peers | no |
| node | Node.Bindings | no |
| node | Node.CommissionerControl | no |
| node | Node.OpenBasicWindow | yes (fabricui/ports.go:81) |
| node | Node.OpenOwnWindow | yes (fabricui/ports.go:85) |
| node | Node.CloseWindow | yes (fabricui/ports.go:89) |
| node | Node.PairingCodesFor | yes (fabricui/ports.go:95) |
| node | Node.DiagnosticEvents | yes (fabricui/ports.go:150) |
| node | Node.OnDiagnosticEvent | yes (fabricui/ports.go:163) |
| node | Node.OnDiagnosticEventAsync | no |
| node | Node.Endpoints | yes (fabricui/ports.go:100; fabricui/node_detail.go:67) |
| node | Node.FactoryReset | no |
| node | Node.OnHealthSample | yes (fabricui/ports.go:159) |
| node | Node.OnHealthSampleAsync | no |
| node | Node.Health | yes (fabricui/ports.go:60) |
| node | Node.ICD | no |
| node | Node.SetupPayload | yes (fabricui/ports.go:92) |
| node | Node.SessionDiagnostics | yes (fabricui/ports.go:102) |
| node | Node.Clock | yes (fabricui/ports.go:104) |
| node | Node.Topology | yes (fabricui/ports.go:98; fabricui/node_detail.go:83) |
| node | Node.MDNSEnabled | yes (fabricui/ports.go:113) |
| node | Node.MDNSServices | yes (fabricui/ports.go:114) |
| node | Node.DiagnoseMDNS | yes (fabricui/ports.go:139) |
| node | Node.Start | no |
| node | Node.Stop | no |
| node | Node.Run | no |
| node | Node.Bridge | no |
| node | Node.LocalAddr | yes (fabricui/ports.go:64) |
| node | Node.Root | no |
| node | Node.Aggregator | no |
| node | Node.Sessions | no |
| node | Node.Subscriptions | no |
| node | Node.RemoveFabric | yes (fabricui/ports.go:71) |
| node | Node.SetCommissioningBlocked | no |
| node | Node.FabricInfos | yes (fabricui/ports.go:67) |
| node | Node.UniqueID | no |
| node | Node.BootReason | no |
| node | Node.Store | no |
| node | Node.Persistence | no |
| node | Node.Groups | no |
| node | Node.Window | yes (fabricui/ports.go:78) |
| node | Node.Observe | yes (fabricui/ports.go:106) |
| node | Node.ObserveAsync | no |
| node | Node.OnOTAProviderStatus | no |
| node | Node.OnOTAProviderStatusAsync | no |
| node | Node.OTAProvider | no |
| node | Node.OTAProviderEndpoint | no |
| node | Node.BDX | no |
| node | Node.TriggerReannounce | no |
| node | Node.ReadAttribute | no (declared in NodeSource fabricui/options.go:56; no call site in non-test code) |
| node | Node.ReadCluster | yes (fabricui/node_detail.go:59; :94) |
| node | Node.Events | yes (fabricui/ports.go:167) |
| node | Node.SessionParameters | no |
| node | Node.Surface | yes (fabricui/ports.go:152) |
| controller | Controller.Discover | yes (fabricui/gofabric/gofabric.go:329; :405) |
| controller | Controller.DiscoverWith | no |
| controller | Controller.Close | no (named only in a comment, fabricui/gofabric/gofabric.go:21) |
| controller | Controller.Registry | yes (fabricui/gofabric/gofabric.go:175; hooks.go:61) |
| controller | Controller.Nodes | no (the .Nodes() hits are ControllerSource.Nodes, not *controller.Controller) |
| controller | Controller.Fabric | yes (fabricui/gofabric/gofabric.go:134; :203; :477) |
| controller | Controller.Peers | no |
| controller | Controller.Clock | no (ports.go:104 is node.Clock) |
| controller | Controller.TCP | no |
| controller | Controller.Addr | no |
| controller | Controller.Commissioner | no |
| controller | Controller.Pair | yes (fabricui/gofabric/gofabric.go:273) |
| controller | Controller.Erase | no |
| controller | Controller.Groups | no |
| controller | Controller.SetGroupKey | no |
| controller | Controller.SoftwareUpdates | yes (fabricui/gofabric/gofabric.go:510) |
| controller | Controller.BDX | no |
| controller | Controller.CameraRequestor | no |
| controller | Controller.OnCameraSignal | no |
| controller | Controller.TCPAddr | no |
| controller | Controller.VendorInfo | no |
| node | Observers.OnFabricAdded | yes (fabricui/events.go:247) |
| node | Observers.OnFabricRemoved | yes (fabricui/events.go:250) |
| node | Observers.OnFabricUpdated | no |
| node | Observers.OnCommissioningComplete | no |
| node | Observers.OnWindowTransition | yes (fabricui/events.go:253) |
| node | Observers.OnSurfaceChanged | yes (fabricui/events.go:266) |
| node | Observers.OnSessionOpened | yes (fabricui/events.go:269) |
| node | Observers.OnSessionClosed | yes (fabricui/events.go:272) |
| node | Observers.OnSubscriptionsChanged | yes (fabricui/events.go:275) |
| node | Observers.OnAttributeChanged | yes (fabricui/events.go:278) |
| node | Observers.OnEvent | no |
| node | HealthSample.Healthy | yes (fabricui/ports.go:61; events.go:121) |
| node | HealthSample.Note | yes (fabricui/ports.go:61; events.go:121) |
| node | DiagnosticEvent.At | yes (fabricui/events.go:111) |
| node | DiagnosticEvent.Kind | yes (fabricui/events.go:112) |
| node | DiagnosticEvent.Severity | yes (fabricui/events.go:113) |
| node | DiagnosticEvent.Message | yes (fabricui/events.go:114) |
| node | DiagnosticEvent.Detail | yes (fabricui/events.go:115) |
| node | SessionDiagnostics.AsOf | yes (fabricui/handlers_topology.go:120) |
| node | SessionDiagnostics.Sessions | yes (fabricui/handlers_topology.go:139; handlers_diagnostics.go:82) |
| node | SessionDiagnostics.Occupancy | yes (fabricui/handlers_diagnostics.go:86) |
| node | MDNSFinding.Severity | yes (fabricui/ports.go:141) |
| node | MDNSFinding.Code | yes (fabricui/ports.go:142) |
| node | MDNSFinding.Message | yes (fabricui/ports.go:143) |
| node | MDNSFinding.Service | yes (fabricui/ports.go:144) |
| surface | Report.Policies | no |
| surface | Report.Level | yes (fabricui/handlers_node.go:105) |
| surface | Report.LevelOrigin | yes (fabricui/handlers_node.go:106) |
| surface | Report.Profiles | yes (fabricui/handlers_node.go:107) |
| surface | Report.Extensions | yes (fabricui/handlers_node.go:108) |
| surface | Report.Effective | no |
| surface | Report.Decisions | yes (fabricui/handlers_node.go:113) |
| surface | Report.Withheld | yes (fabricui/handlers_node.go:110) |
| surface | Report.Assembled | yes (fabricui/handlers_node.go:109) |
| surface | Report.Changed | yes (fabricui/handlers_node.go:111) |

Rows: 111; "no" rows: 52.

## Part 4 - go-fabric 0.18.0 additions (CHANGELOG.md `## [0.18.0]`, lines 13-92)

Selection: `### Added` / `### Changed` bullets whose leading package token is `node`, `cluster/core`, `controller`, `schema` (none lead with `surface` or `eligibility`; the `### Removed`/`### Fixed` sections are excluded). Referenced = grep of fabricui and internal, Go, non-test, for the named symbols.

| CHANGELOG.md line | section | package | symbols named (as written) | referenced by go-fabric-ui |
|---|---|---|---|---|
| 28 | Added | `cluster/core`, `node` | `node.Commissioning.TermsAndConditions`, `core.GeneralCommissioningConfig.TermsAndConditions`, `SetTCAcknowledgements`, `store.SettingTermsAcceptance`, `commissioner.Options.TermsAndConditions`, `controller.PairOptions.Edit` | `TermsAndConditions`: no. `controller.PairOptions` and its `.Edit`: yes (fabricui/gofabric/gofabric.go:69, :435-436, host-supplied template) |
| 29 | Added | `cluster/core`, `node` | `node.TimePeerReader`, `core.TimePeerReader`, `node.NTPClient`, `core.NTPClient`, `core.TimeSyncAttemptTimeout`, `TimeSynchronizationConfig.PlatformTime`, `TimeSynchronization.Start`/`Stop` | no |
| 30 | Added | `cluster/core`, `node` | `core.OTARequestorConfig.TransferProtocols`, `node.OTAConfig.TransferProtocols`, `node.OTADownloadHTTPS`, `node.OTAHTTPClient`, `HTTPSMaxImageSize`, `HTTPSIdleTimeout` | no |
| 32 | Added | `controller` | `Config.Camera`, `CameraControllerEndpoint`, `Controller.CameraRequestor`, `Controller.OnCameraSignal`, `OTAProviderEndpoint`, `Controller.BDX`, `Controller.Groups`, `Controller.SetGroupKey`, `Config.GroupTransport`, `Controller.Addr` | no (grep hit `.BDX` fabricui/gofabric/updates.go:33 is `p.BDXTransferred`/`BDXDataLength`, a different symbol) |
| 41 | Added | `node`, `cluster/wire` | `node.Commissioning.BasicCommissioning`, `wire.NewAdministratorCommissioningWith`, `wire.AdministratorCommissioningConfig` | no (UI calls `Node.OpenBasicWindow`, fabricui/ports.go:81, which this bullet does not name) |
| 50 | Changed | `cluster/core` | `GeneralDiagnostics.EnableTestEventTriggers` | no |
| 61 | Changed | `schema` | `schema.SchemaSnapshotSHA256` | no (UI uses other `schema.` symbols, Part 2) |
| 64 | Changed | `cluster/core`, `node` | `SetUTCTime`, `core.TimeSynchronization.MatterEpochSeconds`, `AttachMatterTime`, `clock.MatterTime`, `clock.MatterTimeOf`, `PlatformTime` | no |
| 68 | Changed | `controller` | `softwareupdate.Config.ProviderNodeID`, `controller.OTAProviderEndpoint` | `ProviderNodeID`/`OTAProviderEndpoint`: no |

Rows: 9.

## Part 5 - go-fabric-ui's own open list

Grep `left out|not yet|later|open|U7|backlog` (case-insensitive) over `docs/concept-u*.md`, `docs/reports/*.md`, `CONCEPT.md`, `README.md`, `notes/briefs/*.md`: 30 hits. Zero hits in `docs/concept-u*.md`, `CONCEPT.md`, `README.md`. Zero hits for `U7` or `backlog` anywhere. Report files present: U1, U1a, U1b, U2, U3a, U3b (no U4/U5 report). Hits that only use "open" for a commissioning window or an SSE stream are omitted; the rest are quoted.

| file:line | quote |
|---|---|
| docs/reports/U1.md:88-92 | "## Left out - Controller family (U2), loom's frame-shell detection (`lite-shell`), audit log, single-item exposure PUT, Playwright e2e, visual regression, CodeQL for JavaScript/TypeScript (would add a required check)." |
| docs/reports/U1.md:96-112 | "### go-fabric facade gaps (CK)" items 1-4: runtime observers, window discriminator, configured setup payload, cluster and device-type names (item 4 continues past line 112) |
| docs/reports/U1b.md:126 | "(system clocks in practice); if needed later, an optional" |
| docs/reports/U2.md:148-152 | "## Left out - Camera QR scanning, OTBR/Thread border-router pages, go-fabric-server binding the family (S4), the node family on generated handlers (all out of scope by the brief)." |
| docs/reports/U2.md:171-175 | "3. **Facade gaps for the stream.** go-fabric offers no commissioning progress hook and no observer on `softwareupdate.Manager`." |
| docs/reports/U3a.md:184-191 | "## Open questions" 1: `.gitignore` `coverage.*` narrowing; 2: adding a method to `ControllerSource` breaks a host implementing the port itself |
| docs/reports/U3a.md:193-200 | "## Left out - The controller graph builder and `/controller/topology` 200 (U3b). - The Network tab on the controller Nodes page (U3b). - Playwright (U5) and live values (U4). - Per-session subscription paths, partition/scope and tags (go-fabric issue, decision 0.4). - No change under `contract/previous/`, and no new npm package." |
| docs/reports/U3b.md:149, :158 | "(open question 1)", "`ErrNotImplemented` (open question 3)" |
| docs/reports/U3b.md:189-205 | "## Open questions" 1: `none` on an edge (go-fabric `Build` never yields it); 2: an admin signal for the pages (status field for admin rights); 3: the 501 detail "the host's controller does not serve this route yet"; 4: go-fabric-server's S-wave wiring of `topology.Changes.Install` |
| docs/reports/U3b.md:207-212 | "## Left out - go-fabric-server's S-wave, the border-router registry, Playwright (U5), U4, the release tag (the owner's). - No go-fabric change, no change under `contract/previous/`, no new npm package." |
| notes/briefs/2026-10-26-F2-family-followups.md:7 | "the add-on Bluetooth option is a later bundle (S4)." |
| notes/briefs/2026-10-25-U1-fabric-ui-node-family.md:161 | "it is skipped (the owner adds the token later)." |
| notes/briefs/2026-11-03-U3a-topology-node.md:141; U3b brief :202; U4b brief :133; U5a brief :143-144; U4a brief :177-178 | each ends with the report shape "open questions, what was left out" (process text, no item) |
