# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); the project uses
[Semantic Versioning](https://semver.org/). The Go module and the npm package
`@sukramj/fabric-ui` share one version.

## [Unreleased]

## [0.9.1] - 2026-11-12

### Changed

- Built with Go 1.27.4: `GO-2026-6859` (`net/http`, fixed in 1.27.4); `go.mod`, the workflows and the Dockerfiles move together.
- go-fabric `v0.22.0` (milestone 7, phase 28: BLE and Wi-Fi commissioning
  on the device side). Nothing these pages read changes; the end-to-end
  test's BLE site migrates from the removed
  `fabrictest.UseWiFiCommissioning` to `node.Network.WiFi.Driver` and gives
  its device a radio with `fabrictest.Topology.BLE`, which
  `fabrictest.Device.BLE` now requires. No migration for an embedder.

## [0.9.0] - 2026-11-10

go-fabric v0.21.0 and the controller family's Diagnostics tab (wave U11,
issue #60, brief `notes/briefs/2026-11-10-U11-controller-diagnostics.md`);
contract 0.9.0, additive over 0.8.1.

### Added

- The controller family's **Diagnostics** tab (`#/controller/diagnostics`,
  "Diagnose" in German): the DNS-SD findings, one row per peer (connection
  state, reachability, operational address and host, time offline, ICD
  check-ins, sessions; a commissioned node links to its page), the open
  sessions with transport, establishment time, resumption and a lost peer,
  and the recent events newest first. It reloads on a burst of
  `controller_diagnostic_event`; a source that serves no diagnostics hides
  the tab and the page shows an empty state.
- `GET /controller/diagnostics`, `GET /controller/mdns` and
  `GET /controller/events` (contract 0.9.0), with the schemas
  `ControllerDiagnostics`, `ControllerPeer` and `ControllerMdns`; the
  stream event `controller_diagnostic_event`; the required
  `ControllerCapabilities.diagnostics`; the optional `Session` fields
  `created_at`, `resumed`, `transport` and `peer_lost` (the controller's
  sessions only); the `DiagnosticEvent` kinds `reachability`, `icd`,
  `ota` and `camera`. A source without diagnostics answers the three
  routes with 501 `not_implemented`.
- `fabricui.ControllerSource` gains `Diagnostics`, `DiagnoseMDNS`,
  `DiagnosticEvents` and `OnDiagnosticEvent`, with the port types
  `ControllerDiagnostics`, `ControllerSession`, `SessionOccupancy`,
  `ControllerPeer`, `MDNSFinding` and `DiagnosticEvent`; a source of its
  own implements them, and `fabricui.ControllerCapabilities` gains
  `Diagnostics`, the flag a source serving them sets. The
  `gofabric` adapter sets it and maps go-fabric's controller diagnostics one to one;
  `MemoryController.NoDiagnostics` stands for a source without them.

### Changed

- go-fabric v0.20.0 → v0.21.0. The controller serves diagnostics — the
  sessions and connection state of every peer, a bounded trace with
  reachability, ICD, OTA and camera events, and the browse-side mDNS
  finding `no_operational_names` (issue 483) — which this release's tab
  shows. A peer learns its proven operational host from discovery, not
  only when a session is established, and while another responder's
  cache-flush hides that host a reconnect tries the last operational
  address first (issue 488).

## [0.8.1] - 2026-11-10

go-fabric v0.20.0 and the node record's operational host (wave U10, brief
`notes/briefs/2026-11-10-U10-gofabric-0.20.0.md`); contract 0.8.1,
additive over 0.8.0.

### Added

- `NodeRecord.operational_host` on `GET /controller/nodes/{node}` and
  `fabricui.NodeRecord.OperationalHost`: the operational SRV target that
  owned the node's address when it was last reached, which go-fabric
  0.20.0 records as `client.Record.OperationalHost`; absent when unknown
  or stored before go-fabric 0.20.0, and not `hostname`, the
  commissionable advertisement's host. The controller node page's overview
  shows it as "Operational host" when set.

### Changed

- go-fabric v0.19.1 → v0.20.0. v0.19.2: the fabric-leaving flush of a
  node's factory reset or fabric removal sends the Leave event and the
  attributes owed in one report and waits at most 500 ms, so a reset no
  longer waits about 14 s. v0.20.0: the controller reconnects to a node
  through the operational host its last session proved first, so a Thread
  node behind a border router that answers its instance with its own host
  no longer costs every reconnect 45 s.

### Fixed

- `contract/openapi.yaml`: the descriptions of `SessionParameters.supported_transports` and `TopologyGraph.as_of` are quoted — inside a flow mapping their comma split them, so the generated `TopologyGraph.AsOf` comment read "When the graph was read" and a strict parser (OpenCCU-Loom's contract merge) refused the file (#58). No change in substance.

## [0.8.0] - 2026-11-09

go-fabric v0.19.1: the Terms and Conditions retry, the host's counters and
the Thread border routers (wave U9, brief
`notes/briefs/2026-11-09-U9-gofabric-0.19.1.md`, issues #55 and #56);
contract 0.8.0, additive over 0.7.1.

### Added

- `POST /controller/commission` answers a device that requires an accepted
  version of its terms and conditions with 409 `terms_required` (#55),
  carrying the device's `version`, `acknowledgements` (null) and
  `deadline` (Matter epoch seconds, or null); the port's error is
  `*fabricui.ErrTermsRequired`, which the go-fabric adapter makes from
  `commissioner.TermsRequiredError`, and the failed
  `commissioning_progress` carries the kind `terms_required`. The
  Commission page opens the acknowledgement block with the version filled
  in, the acknowledgements field focused and the deadline as a date, and
  one click sends the same commissioning again with the acknowledgement.
  The in-memory controller's `Terms` stands for such a device.
- `Options.Counters` and `Counter{Code, Value, Message}`: the host's own
  counters — go-ha2fabric's `discards`, say — as `counters` on
  `GET /node/status` and `/nodes/{node}/status`, absent without the hook;
  the Status page shows them in a small table under the warnings.
- The Network page draws the Thread border routers go-fabric's registry
  knows as `border_router` nodes (#56), labelled by their host name, `ok`
  while seen within the registry's expiry, with host, vendor, model and
  network name and when they were last seen; Thread neighbours carry
  their `network_name` where go-fabric fills it, and a legend under the
  graph names every kind it draws. A node of a kind the adapter does not
  know yet is drawn by its network instead of being dropped.

### Changed

- go-fabric v0.19.1 (was v0.18.0). The go-fabric adapter sets the request's
  Terms and Conditions acknowledgement on `PairOptions.TermsAndConditions`
  (and the commissioner's options for a discovered device) instead of
  through `Edit`; the host's `Edit` runs after it, as go-fabric runs it,
  and the template's acknowledgement applies when the request sends none.

## [0.7.1] - 2026-11-08

The Commission page's Bluetooth LE decision is stated, never silent (wave
U8, brief `notes/briefs/2026-11-07-U8-ble-no-silent-drop.md`); contract
0.7.1, additive over 0.7.0.

### Added

- `fabricui.NodeSwitch` (`NewNodeSwitch`, `Set`, `Current`): a
  `NodeSource` whose target the host swaps after an erase, so the pages
  keep one source while the node is remade; a stream open across the swap
  keeps the node it was opened on until the page reconnects.

### Fixed

- The Commission page no longer drops Bluetooth LE silently when no
  network credential is picked (#49): the page preselects the host's
  default or only credential, hints while none is picked on a host with a
  central, and the commissioning's progress says whether BLE was used and
  why not (`ble` on `commissioning_progress`, `fabricui.pair.ble_skipped`
  in the log). `Credential.default` names the host's default entry.

## [0.7.0] - 2026-11-07

Completeness against go-fabric v0.18.0 (wave U7, brief
`notes/briefs/2026-11-07-U7-completeness.md`); contract 0.7.0, additive
over 0.6.0.

### Added

- `POST /node/mdns/reannounce` and `/nodes/{node}/mdns/reannounce`
  (admin): the node announces its DNS-SD records now
  (`Node.TriggerReannounce`); a Re-announce button on the diagnostics
  page.
- `POST /node/maintenance/erase` and `/nodes/{node}/maintenance/erase`
  (admin, body `{"confirm":"erase-node"}`): the host's erase of a node,
  `Options.Erase` and `NodeEntry.Erase` — go-fabric's `Node.FactoryReset`
  and the host's way back to a node (`fabrictest.Device.FactoryReset` is
  the pattern); 501 `not_implemented` without one. `Status.erase` says
  whether it is wired; the fabrics page offers it behind two
  confirmations and polls the status until the node answers again. The
  factory-reset route is unchanged: it removes the fabrics and keeps the
  node.
- `Status.attestation_error` (`Node.AttestationError`) and
  `Status.boot_reason` (`Node.BootReason`), shown on the status page.
- `SessionList.parameters` (`Node.SessionParameters`), a collapsed block
  on the diagnostics page.
- Node stream events `fabric_updated`, `commissioning_complete` and
  `cluster_event` (`node.Observers` `OnFabricUpdated`,
  `OnCommissioningComplete`, `OnEvent`); the node store reloads, toasts
  and hands a cluster event to its endpoint's page.
- `CommissionRequest.terms_and_conditions` and `PairOptions.
  TermsAndConditions`: the Terms and Conditions acknowledgement a
  commissioning sends (go-fabric `commissioner.Options.
  TermsAndConditions`, set by the adapter through the commissioner's
  `Edit`); a disclosure on the commission page.
- `PairOptions.Progress` and `ControllerEvent.Step`: each step of a
  commissioning (go-fabric `controller.PairingStep`) is a
  `commissioning_progress` with `phase` `started` and `step`; the
  commission page lists them.
- `ControllerEvent.UpdateKind`, `UpdateVersion`, `UpdateErr` and the
  stream's `update_kind`, `update_version`, `update_error`: the adapter
  observes `softwareupdate.Manager.Observe`; the page toasts a finished or
  failed update.

### Changed

- **Breaking for a host that wraps its node:** `NodeSource` gains
  `AttestationError`, `BootReason`, `SessionParameters` and
  `TriggerReannounce`. `*node.Node` has them; a wrapper forwards them
  (`docs/embedding.md`).
- The controller stream's `update` events come from the
  `ControllerSource`, not the handler: the handler no longer publishes
  one after `POST /controller/updates/check` or `/consent`. A host with a
  `ControllerSource` of its own emits them; the go-fabric adapter and the
  memory controller do.
- The node stream reports a commissioning as `fabric_added`, then
  `commissioning_complete`.

## [0.6.1] - 2026-11-07

### Changed

- go-fabric `v0.18.0` (milestone 10 phases 36–39 and the matter.js pin lift
  862ad75f → db021da0). What these pages read moves with it: the node's UTC
  is an offset over its clock (`cluster/core`, `node`; SetUTCTime sets it,
  the host clock never); the projected OccupancySensing's default selection
  is PassiveInfrared with OccupancyEvent, FeatureMap `0x202`
  (`cluster/measurement`); the OTA provider is the controller itself —
  `softwareupdate.Config.ProviderNodeID` is RootNodeID, endpoint
  `controller.OTAProviderEndpoint` — and `controller.OwnNode` and
  `node/controllernode` are removed; the OTA requestor's
  `node.OTAConfig.TransferProtocols` (HTTPS, opt-in);
  `schema.SchemaSnapshotSHA256` moves with the pin lift. No migration: no
  removed symbol is used here.

## [0.6.0] - 2026-11-05

Several nodes per host, created in the page (wave U6a, issue #47, concept
`docs/concept-u6-nodes.md`); contract 0.6.0, additive over 0.5.0. The
first node stays the host's configuration; every further node is registry
state the operator creates, renames and deletes. A host without
`Options.Registry` serves 0.5.0's routes as before.

### Added

- The node registry port: `fabricui.NodeRegistry` (`List`, `Create(ctx,
  NodeSpec{Label})`, `Rename`, `Delete`), `Options.Registry` (needs
  `Options.Node`; `Handler` panics and `Mount` fails beside `Options.Nodes`), the
  errors `ErrNodeLimit` and `ErrDefaultNode`, `MaxNodeLabel`.
- `POST /nodes` (admin, 201 `NodeSummary`; 409 `node_limit`), `PUT
  /nodes/{node}/label` (admin, 204) and `DELETE /nodes/{node}` (admin,
  204; 409 `node_is_default`), each 503 without a registry; the problem
  codes `node_limit` and `node_is_default`; the request body
  `NodeLabelRequest` (1 to 64 characters).
- `GET /nodes` lists the registry's nodes after the default, read on every
  request; `NodeSummary.managed` and `NodeSummary.exposure`.
- `Status.node_registry` (with a registry) and `Status.admin` (host mode:
  whether the host's admin check admits the request).
- The stream event `nodes_changed` (`NodesChangedEvent{keys}`) on every
  node's stream after each registry write.
- Per-node selection: `NodeEntry.Exposure`, `Reassemble` and `Events` serve
  the exposure and maintenance routes and the stream under
  `/nodes/{key}/…` (for `Options.Nodes` and the registry's entries alike);
  the Options-level ones stay the default node's.
- The configuration block: `Status.config` (`NodeConfig{source,
  source_note?, label?, listen, discriminator, vendor_id, vendor_name?,
  product_id, product_name?, store?}`), from `Options.NodeConfig` for the
  default node and `NodeEntry.Config` for any other (`NodeConfigView`);
  never the passcode. The status page shows it read-only and says where
  the values come from.
- Shell: "Create node…" in the node selector (with `node_registry` and
  admin rights; the selector shows from one node then), a label dialog,
  the new node's selection page with a hint to the Pair page; a keyed
  route whose node is gone goes to the default node.
- A managed node's status page: rename, delete after two confirmations
  naming its label and fabric count, and the several-nodes ecosystem hint
  (Alexa pairs reliably only on port 5540 and refuses bridges above about
  80–100 devices; home-assistant-matter-hub's bridge configuration page).
- The npm package: the context's `nodes` store (`list`, `loaded`, `find`,
  `load`, `create`, `rename`, `remove`), the client's `createNode`,
  `renameNode` and `deleteNode`, `NodeStoreDeps.onNodesChanged`, the
  `NodeConfig` and `NodesChangedEvent` types.
- The e2e server's registry double and the Playwright smoke: create a node
  from the selector, find it listed, delete it from its status page.

### Changed

- `contract/previous/openapi.yaml` is the released 0.5.0 contract.

## [0.5.0] - 2026-11-05

The controller family's device details (brief U4b, concept
`docs/concept-u4-device-details.md`, wave D.1); contract 0.4.0, additive
over 0.3.1. The node family's device details and the event log of both
families (brief U4a, wave D.3); contract 0.4.1, additive over 0.4.0.
Operation — the UI half (brief U5a, concept `docs/concept-u5-operation.md`,
wave D.1); contract 0.5.0, additive over 0.4.1.

### Added

- The node selector (contract 0.5.0): `GET /nodes` (`NodeSummaryList`,
  the default node first) and every node route also under
  `/nodes/{node}/…` (operation ids suffixed `ForNode`; 404 for an unlisted
  key). `Options.Nodes []NodeEntry{Key, Label, Kind, Source}` beside
  `Options.Node`, which stays the default node under `/node/*` and the
  key `default`. The shell shows a selector in its header when there are
  two nodes and routes a node's pages under `#/node/<key>/…`; the client's
  `forNode(key)` and `nodes()`.
- `Status.exposure_reason` and `Options.ExposureReason`: the selection page
  of a node without a selection says why instead of hiding.
- The **Operation** page and its routes (0.5.0): `GET /operation/info`,
  `GET`/`PUT /operation/log/level`, `GET /operation/log` (`since`,
  `limit` 1..2000, default 200) and `GET /operation/log/stream`
  (`log_record`), 503 without a port. New port `fabricui.Operations`
  (`Options.Operations`) with `LogRecord` and `OperationInfo`,
  `fabricui.MemoryOperations` as its double, `fabricui.GoFabricVersion`.
  The page sets the level (admin) and shows the ring and the stream with
  pause, a level and a text filter and copy; its tab shows when the
  host attaches the port.
- `POST /controller/nodes/{id}/ping` (admin, `PingResult`): a Matter read
  of BasicInformation DataModelRevision through the node's session with
  its round trip, not an ICMP echo, as go-fabric-server's `ping_node`;
  `ControllerNode.Ping`, in the go-fabric adapter and `MemoryController`.
  A Ping button on the controller's node overview with the outcome per
  address and the sentence what was measured.
- Stream events `health_sample` (`HealthSampleEvent`) and
  `diagnostic_event` (the trace entry's shape, `fabricui.TraceEvent`),
  from `Node.OnHealthSample` and `Node.OnDiagnosticEvent`, registered per
  stream; `fabricui.ObserveNode` registers them beside `NodeObservers`
  for a host's own channel. The status page reloads on a burst of health
  samples, the diagnostics page in place on a burst of trace entries.
- Camera QR scan on the commissioning page: the decoder (`jsqr` 1.4.0,
  Apache-2.0, in `THIRD-PARTY-NOTICES.md`) and the scan are a chunk loaded
  on the first scan, the browser's `BarcodeDetector` used when it reads QR
  codes; outside a secure context the page explains why there is no
  button. The payload fills the code field.
- Pairing hints per ecosystem (Apple, Google, Alexa, SmartThings, Home
  Assistant) below the pairing flow, from matter.js `docs/ECOSYSTEMS.md`,
  with the ecosystems already holding a fabric marked and their
  compatibility findings beside them.
- A Playwright smoke per page (`ui/e2e`, Chromium only, `make e2e-web`,
  CI job `e2e-web`) against `TestServeE2EWeb`, a test binary serving the
  bundle on the handler tests' doubles.
- Fixtures `nodes.200`, `nodes-status.200`, `nodes-status.404`,
  `operation-info.200`, `operation-info.503`, `operation-log.200`,
  `operation-log.400`, `operation-log-level.200`,
  `operation-log-level-set.200`, `operation-log-level-set.400`,
  `operation-log-level-set.403`, `controller-node-ping.200`,
  `controller-node-ping.404`.

- `GET /controller/nodes/{id}/endpoints/{ep}` (`EndpointDetail`) and
  `GET /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}`
  (`ClusterDetail`): a node's inventory read from the controller's image,
  never from the node and without admin (owner decision U4 0.1 amends
  0.4). An endpoint carries the Descriptor's DeviceTypeList with each
  revision, its TagList, the FixedLabel and UserLabel entries, its parts,
  BridgedDeviceBasicInformation Reachable, the device type conformance
  verdict and its server clusters without values; a cluster its revision,
  feature map in hex with the features it sets (code, title, bit), data
  version, attributes with names and values (JSON by field name),
  accepted and generated commands and the EventList, named where go-fabric
  has a generated definition. 404 for an endpoint or cluster the image
  lacks, 503 while the ServerList names a cluster the controller has not
  read yet.
- Contract types `EndpointDetail`, `EndpointDeviceType`, `SemanticTag`,
  `EndpointLabel`, `ConformanceVerdict`, `ConformanceViolation`,
  `ClusterDetail`, `ClusterFeature`, `ClusterAttribute`, `ClusterElement`;
  `NodeEndpoint.reachable` where the endpoint serves
  BridgedDeviceBasicInformation.
- The port `fabricui.ControllerNode` gains `EndpointDetail` and
  `ClusterDetail` with their types, and the errors `ErrEndpointNotFound`,
  `ErrClusterNotFound` (404) and `ErrNotRead` (503);
  `gofabric.ControllerOf` serves them from `client.Endpoint` and
  `client.Cluster`, and `MemoryController`'s light carries a tag, a
  conformance verdict and OnOff's inventory.
- The **Inventory** tab of the controller's node page (`tab=inventory`):
  per endpoint its device types, tags, labels, Reachable, conformance and
  clusters; a cluster loads its values when it is opened, and the tab
  reloads on the node's `node_updated`. The endpoints tab shows Reachable
  per bridged endpoint and links each endpoint into the inventory.
- Kit components `ClusterView`, `AttributeTable` and `TagChips`, for both
  families; the client's `controllerEndpointDetail` and
  `controllerClusterDetail`.
- Fixtures `controller-node-endpoint.200`, `controller-node-endpoint.404`,
  `controller-node-cluster.200`, `controller-node-cluster.404`.
- `GET /node/endpoints/{ep}` (`EndpointDetail`, 0.4.1): one endpoint the
  node serves, read through its own read path (go-fabric
  `Node.ReadCluster`) — the Descriptor's DeviceTypeList with each
  revision and its TagList, the FixedLabel and UserLabel entries, every
  server cluster with its values, data version, feature map and the
  names go-fabric's schema gives its attributes, commands and events; a
  bridged endpoint's name, Reachable, `scope`, parent and parts; the
  surface report's `decisions` on the endpoint. The root endpoint and the
  Aggregator answer too. 404 for an endpoint the node does not serve, 503
  before the node started.
- `GET /node/endpoints/{ep}/events` and `GET /controller/nodes/{id}/events`
  (`EventRecordList` of `EventRecord`, 0.4.1, `since` and `limit`): the
  node's event log for one endpoint (`Node.Events`) and the events the
  controller received from a node (`client.Node.EventHistory`), oldest
  first, each with its priority, name and data by field name.
- The stream event `attribute_changed` (`AttributeChangedEvent`:
  endpoint, cluster and the attribute ids, never a value) from go-fabric's
  `node.Observers.OnAttributeChanged`, through `fabricui.NodeObservers`.
- `NodeSource` gains `ReadCluster`, `ReadAttribute` and `Events`, which
  `*node.Node` has; `ControllerNode` gains `Events`, served by
  `gofabric.ControllerOf` and `MemoryController` (two events).
  `EndpointDetail.Scope`, `EventRecord`, `EventPriorityName`.
- The node family's endpoint page `NodeEndpointDetail`
  (`#/node/endpoint/<n>`, routes key `endpoint`): device types with
  revisions, tags (with the reason when there are none), labels, scope,
  parent and parts, the clusters with their values, the surface decisions
  and the endpoint's events; it refetches on the endpoint's
  `attribute_changed` (coalesced per 250 ms in the store's
  `attributeEpoch`) and on `endpoints_changed`/`surface_changed`, the
  events after the last one only. The endpoint list links every endpoint,
  the root endpoint and the Aggregator to it and shows each device type's
  revision; the topology's endpoint refs lead to it.
- The **Events** tab of the controller's node page (`tab=events`): the
  events the controller received, newest first, fetching the newer ones on
  the node's `node_updated`.
- Kit component `EventList`; the client's `nodeEndpointDetail`,
  `nodeEndpointEvents` and `controllerNodeEvents`.
- Fixtures `node-endpoint.200`, `node-endpoint.404`,
  `node-endpoint-events.200`, `controller-node-events.200`.

### Changed

- **Breaking** for a host with its own `NodeSource`: the interface gains
  `OnHealthSample(func(node.HealthSample)) func()` and
  `OnDiagnosticEvent(func(node.DiagnosticEvent)) func()` (owner decision
  U5 F.3). `*node.Node` satisfies both; a wrapper forwards them to its
  node. Likewise `ControllerNode` gains `Ping(ctx) (PingResult, error)`;
  `gofabric.ControllerOf` and `MemoryController` implement it.
- The selection page loads the selection only once the status says the
  host serves one, so a node without a selection is never asked for one.
- An event a node delivers to the controller raises the node's
  `node_updated` (`gofabric.Hooks`), and `Hooks.Wrap` chains the host's
  own `Event` hook instead of replacing ours with it.
- `GET /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}`
  answers 503, not 404, while the controller has not read the endpoint's
  Descriptor; a value that does not render is absent, not `null`.
  `EndpointDetail.parts` are the direct children (the contract says so),
  in `MemoryController` too. The Inventory tab drops a reload's answer
  that arrives after a newer one's.
- The endpoints of `GET /controller/nodes/{id}` list every device type of
  the Descriptor's DeviceTypeList, each with its revision, not only the
  endpoint's own.
- **Breaking** for a host with its own `NodeSource`: the interface gains
  `Endpoints() []node.EndpointInfo` (issue #43). `*node.Node` satisfies
  it; a wrapper forwards it to its node.
- The node family's endpoint detail takes its parent, direct children,
  device types, name and scope from the endpoint's `node.EndpointInfo`
  (`Node.Endpoints`) — the root and the Aggregator from the same source as
  a bridged endpoint — instead of reading every part's PartsList;
  Reachable stays from the topology. Which endpoints carry
  `parent_endpoint_id` is unchanged: none for the root and the
  Aggregator. `GET /node/endpoints` adds `structural` (`StructuralEndpoint`,
  contract 0.5.0): the root and the Aggregator from that tree, which the
  Endpoints page links instead of inferring the Aggregator from the
  bridged endpoints' parents.
- A status column by name: an attribute `Node.ReadCluster` answers with a
  failure is listed with its IM `status` and no `value` instead of being
  left out, and the attribute table of both families shows a known status
  by name (`inventory.status.*`, en and de) beside its code. The
  cluster's `data_version` is a value's, never a failed attribute's 0.
- The ping is measured per address (go-fabric #398): the go-fabric
  adapter probes every address of the node's record on its own with
  `client.Node.PingAddress` — an empty read over the held session sent to
  that address, at most four at once, nothing migrates — and answers
  `per_address: true`; the result is reachable when one address answered,
  with the fastest round trip. `PingAddress` (contract and port) adds an
  optional `error` per address. A node without a session or without a
  known address keeps the single read through the session
  (`per_address: false`). `MemoryController`'s light lists a second,
  silent address (`192.168.1.11:5540`); the node overview shows each
  address with its badge, round trip and reason.
- go-fabric `v0.17.0` (milestones 8 and 9: the certification harness with
  the Go controller, the test register over every matter.js root, the second
  compositions of the reference daemon, determinism, and the module side of
  these pages — `Node.ReadAttribute`/`ReadCluster`/`Events`/`Endpoints`,
  `Node.PingAddress`, `logring`); replaces the pseudo-version of `8feaa508`;
  no migration.

## [0.4.1] - 2026-11-03

The topology view of the controller family (brief U3b, concept
`docs/concept-u3-topology.md`); contract 0.3.1, additive over 0.3.0.

### Added

- `GET /controller/topology` answers 200: the controller's network graph
  from go-fabric's `controller/topology` — the controller with a session
  to each commissioned node (`peer:<node_id_hex>`), the Thread neighbours
  no node matches (`thread_router`, `thread_end_device`), the Wi-Fi access
  points (`wifi_ap`) and a `link` per Thread pair and Wi-Fi station with
  its strength, LQI and RSSI. `gofabric.ControllerOptions.Topology` takes
  the host's `*topology.Service`; its changes reach the controller stream
  as `topology_changed`. Without a service the route still answers 501.
- `TopologyEdge.strength` gains `none` (a link observed dead, LQI 0).
- The **Network** tab of the controller's nodes page (`ControllerNodes`
  prop `tab`, route `#/controller/nodes?tab=network`, routes key
  `controllerNetwork`; `ControllerNetwork` alone): the graph, an admin
  "Read diagnostics again" (`refresh=true`), a reload on the controller
  store's new `graphEpoch`. The kit's `Graph` draws a link's strength
  (width, dotted when dead) with its quality as the title.
- `MemoryController.Topology`: the graph the in-memory controller serves.
- Fixtures `controller-topology.200` and `controller-topology.403`.

### Changed

- Contract 0.3.1: `/controller/topology` is served (no `x-status:
  planned`).
- go-fabric `v0.16.1-0.20261103130909-c91e5b7f4a2f` (main at `c91e5b7f`,
  `controller/topology`), a pseudo-version until go-fabric's next tag.

## [0.4.0] - 2026-11-03

The topology view of the node family (brief U3a, concept
`docs/concept-u3-topology.md`); contract 0.3.0, additive over 0.2.0.

### Added

- `GET /node/topology`: the node as a graph (`TopologyGraph`) — the node,
  its Aggregator, the bridged devices and their parts, the fabrics and the
  sessions their controllers hold — built from `Topology`, `FabricInfos`
  and `SessionDiagnostics`; 503 without a node or before it assembled.
- The node stream carries `session_opened` (CASE sessions only, as go-fabric
  reports them), `session_closed` and `subscriptions_changed` with a
  `SessionEvent` (the session as `/node/sessions` lists it);
  `fabricui.NodeObservers` yields them too.
- The **Topology** page (`NodeTopology`, route `#/node/topology`, routes key
  `topology`) and the kit's `Graph` component with its pure layout
  (`layoutGraph`): hand-written SVG, pan and zoom, a details panel that
  links each node to the page that lists it, keyboard access. It reloads
  on the store's new `graphEpoch`.
- Client `topology()` and `controllerTopology(refresh)`; problem code
  `not_implemented` (501); the catalogue test (`catalogs.test.ts`).
- `GET /controller/topology` declared with its final schema (then planned,
  501), `refresh=true` admin-only, and the controller stream's
  `topology_changed`; served from 0.4.1.

### Changed

- **Breaking for a host with its own `ControllerSource`:** the port gains
  `NetworkTopology(ctx, refresh) (NetworkTopology, error)`; returning
  `fabricui.ErrNotImplemented` keeps the route at 501. A source wrapping
  `gofabric.ControllerOf` inherits it.

## [0.3.0] - 2026-11-02

Built on go-fabric v0.16.0, whose facade removed what `NodeSource` and the
device type names rested on. The routes, their bodies and the npm package's
pages are unchanged.

### Changed

- go-fabric `v0.16.0` (was `v0.15.6`), which raises `modernc.org/sqlite` to
  1.61.0, `modernc.org/libc` to 1.79.0, `go-humanize` to 1.1.0 and
  `go-strftime` to 1.1.0 in the build list.
- **Breaking for a host with its own `NodeSource`:**
  `CommissioningWindow() *node.CommissioningWindow` is replaced by
  `OpenOwnWindow(ctx, timeout) error` and `CloseWindow(ctx) error`.
  `*node.Node` satisfies the interface as before.
- `fabricui.NodeObservers`: the hooks return nothing, as go-fabric's
  `node.Observers` now declares them.
- `POST /node/pairing/window/close` closes the window through
  `Node.CloseWindow` (matter.js `endCommissioning`) instead of
  `RevokeWindow`: a PASE session already established and an armed fail-safe
  now stay with their commissioner, where they were evicted and expired
  before. The route's status codes are unchanged. A host that needs the
  eviction wraps its node in a `NodeSource` whose `CloseWindow` revokes.
- `device_type_name` keeps its values: go-fabric removed
  `contract.DeviceTypeName` and its label table, so the 25 operator labels
  ("On/Off Light") move to `fabricui`, and every other device type reads
  `schema.DeviceTypeName`. The import guard allows `schema` beside the facade.

## [0.2.9] - 2026-11-01

Every response schema of the contract is open to unknown fields again, and
a guard keeps it so; the npm package is published through trusted
publishing.

### Fixed

- `contract/openapi.yaml`: the nine schemas of the surface report
  (`SurfaceReport`, `SurfaceEffective`, `SurfaceEndpoint`,
  `SurfaceDeviceType`, `SurfaceCluster`, `SurfaceElement`,
  `SurfaceDecision`, `SurfaceChange`, `SurfaceElementChange`) no longer set
  `additionalProperties: false`. A client generated from a closed schema
  decodes it strictly, so the next field go-fabric adds to
  `surface.Report`'s JSON — additive for the contract guard, a minor for the
  changelog — made every installed client reject the whole
  `GET /node/surface` answer. No other response schema was closed; request
  bodies keep the keyword, where the handler is the decoder. Closes #27.
- The contract guard (`make contract-check`) fails on a schema reachable
  from a response that sets `additionalProperties: false`, and its tests
  check every recorded fixture's response against the route's schema, so a
  field the encoder writes and the contract does not name is caught at test
  time instead of on a client's wire.

### Changed

- go-fabric `v0.15.6` (was `v0.15.4`); oapi-codegen runtime `v1.8.0` and
  generator `v2.8.2` (was `v1.7.0`, `v2.8.0`) — the regenerated handlers
  bind request bodies through `runtime.BindJSONBody`, the same decoding.
  npm: `@lucide/svelte` 1.66.0, `svelte` 5.57.3, `vite` 8.3.5,
  `@types/node` 26.6.6; the bundle is rebuilt. `minimatch`, reached only
  through `openapi-typescript`'s `@redocly/openapi-core` (the type
  generator, not the bundle), is overridden to 11.0.1 for GHSA-g3hv-75wg-m249
  (`overrides` in `ui/package.json`; `gen:types` produces the same file).
  TypeScript stays at 6.x: 7.0 is a major whose toolchain `svelte-check`
  4.7 does not declare. Left as reported: `deepmerge` ≤ 4.3.1 (low,
  GHSA-gcqx-w48m-w2h4) under `@sveltejs/vite-plugin-svelte`, no fixed
  version published.
- The release workflow publishes `@sukramj/fabric-ui` through npm's trusted
  publishing: npmjs accepts the job's GitHub OIDC token (Trusted Publisher on
  the package, set to `SukramJ/go-fabric-ui` and `release-on-tag.yml`) and
  signs the provenance from it. The `NPM_TOKEN` secret, which npm limited to
  a short lifetime, is no longer read; a missing trust configuration fails
  the publish instead of skipping it. The job installs npm 11 first, which
  trusted publishing requires.

## [0.2.8] - 2026-10-30

`surface_changed` comes from the node itself, so every assembly that
changes an endpoint's surface reaches the stream; the Surface page shows the
root endpoint.

### Changed

- `surface_changed` comes from the node's observer
  (`Observers.OnSurfaceChanged`, registered per stream through
  `NodeObservers`): one event per changed endpoint, so a reassembly the
  host runs on its own reaches the stream too. The handler no longer
  compares `Surface().Changed` before and after its own reassemblies.
- The status page's warning count reads "{count} warnings" — any policy's
  `Warn` counts, not only a profile's.
- The Surface page shows the root endpoint (`effective.root`) as the first
  endpoint card, titled "Root endpoint"; its decisions are listed under
  that title.
- go-fabric `v0.15.4` (was `v0.15.3`), which adds `Observers.OnSurfaceChanged`.

## [0.2.7] - 2026-10-30

The node's surface report gets its own page, and a reassembly that changes
an endpoint's surface reaches the stream and the Endpoints page.

### Added

- `GET /node/surface`: go-fabric's `surface.Report` of the node in
  go-fabric's own JSON encoding — the effective surface per endpoint, the
  decision ledger, the withheld elements and the announced surface changes
  (contract schema `SurfaceReport`).
- The Surface page (`NodeSurface`; `#/node/surface` and a tab in the
  standalone shell, `Routes.surface`): the policies' level, profiles,
  extensions and digest; per endpoint its device types and clusters with
  revision, feature map and element counts; the decisions, filterable by
  verdict and endpoint, the silent keeps hidden by default; the withheld
  elements; the announced changes with their element changes.
- The stream event `surface_changed` (`SurfaceChangedEvent`, the endpoint
  keys): published after `endpoints_changed` when a reassembly this
  handler runs announced surface changes.
- The Endpoints page marks an endpoint whose surface changed in this
  process ("surface changed"), and refetches on `endpoints_changed` and
  `surface_changed`.

### Changed

- The status page's `warning_count` counts `Warn` decisions — the verdict,
  not the reason's text (go-fabric v0.15.3's `surface.Warn`).
- The German status text says "Pin" for a level taken from the pin.
- go-fabric `v0.15.3` (PR #345) (was `v0.15.2`), which adds `surface.Warn`.

## [0.2.6] - 2026-10-30

The node status page shows the policies go-fabric v0.15 composes the served
surface under: the specification level, the controller profiles and the
enabled extensions.

### Added

- A "Surface" row on the node status page, below the counts: the
  specification level with its origin (set, the pin, or the profiles that
  imply it), the selected controller profiles, the enabled
  ahead-of-matter.js extensions, and — only when not zero — the withheld
  elements, the profile warnings and the surface changes announced in this
  process.
- `GET /node/status` reports a `surface` object (absent without a node):
  `level`, `level_origin`, `profiles`, `extensions`, `assembled`,
  `withheld_count`, `warning_count`, `changed_count`, read from
  `node.Node.Surface()`.

### Changed

- go-fabric `v0.15.2` (was `v0.15.1`), which adds `surface.Report.Extensions`.
- `NodeSource` gains `Surface() surface.Report`. `*node.Node` satisfies it
  at go-fabric v0.15.2; a host that implements `NodeSource` with its own
  type adds the method.
- The import guard allows `github.com/SukramJ/go-fabric/surface` as a facade
  type package: it is the type of `node.Config.Surface` and of
  `Node.Surface()`'s result.

## [0.2.5] - 2026-10-30

`fabricui/gofabric` builds against go-fabric v0.15.1, and the selection page
hides what the host cannot expose.

### Changed

- The selection page hides items whose verdict is unmappable — nothing the
  host could expose now — by default. A toggle above the list ("Show items
  that cannot be mapped (n)") brings them back, with their reason and the
  disabled checkbox as before; the value filters still offer every item's
  values. The legend keeps all four states.
- go-fabric `v0.15.1` (was `v0.14.0`): the served surface as a value under
  policies (ADR 0030) — the node's spec level, ahead-of-matter.js opt-ins and
  controller profiles sit in `node.Config.Surface`; nothing in this module's
  routes changes.

## [0.2.4] - 2026-10-29

`fabricui/gofabric` builds against go-fabric v0.14.0, and a host gets the
node's transitions over its own channel through an exported route
(issue #19).

### Added

- `NodeObservers(push, open) node.Observers`: the node's fabric and window
  transitions as the node page family's events, with the stream's
  payloads — the conversion the handler registers for each open stream,
  which since v0.2.0 is the transitions' only route. A host that carries
  the events over a channel of its own (a WebSocket hub, an audit log)
  registers it with `node.Observe` instead of rebuilding it, and does not
  publish the transitions into `Events`, which would hand every stream
  each one twice.

### Changed

- go-fabric `v0.13.3` → `v0.14.0`. The controller adapter decodes a Thread
  dataset with `github.com/SukramJ/go-fabric/thread`, where go-fabric
  v0.13.7 moved the codec from `controller/thread` (every identifier
  unchanged); `fabricui/gofabric` did not compile against go-fabric
  v0.13.7 or later before. The import guard allows `thread` in its place.

## [0.2.3] - 2026-10-27

### Changed

- Build tooling: `@lucide/svelte` 1.57.0 → 1.59.0, `svelte-check` and
  `vitest` minor updates (Dependabot #3). The rebuilt bundle is
  byte-identical; `THIRD-PARTY-NOTICES.md` records the new `@lucide/svelte`
  version.

## [0.2.2] - 2026-10-27

The pages take Home Assistant's theme from the first frame: inside an
Ingress panel they wear the live theme (colours and light/dark, a custom
theme included), and standalone they wear Home Assistant's default light
and dark palettes, following `prefers-color-scheme`. Checked in headless
Firefox inside a same-origin fake Home Assistant parent with a light, a
dark and a custom theme, and standalone.

### Fixed

- No flash of the wrong palette on a cold load. The shell's pre-paint
  script sets the Home Assistant skin before the first paint and, inside
  Home Assistant's Ingress frame, copies the parent's theme variables and
  takes its light/dark from the parent's background — before, the first
  frames showed the standalone teal/white palette, and dark followed the
  operating system instead of Home Assistant until the bundle had run.
- The theme bridge drops a variable the parent no longer defines (a custom
  theme switched back to the default), so the token falls back to Home
  Assistant's default instead of keeping the old theme's value.
- The underline tabs no longer show a vertical scrollbar: the active
  marker overlapped the list's border by one pixel, which overflowed the
  scrollable tab list.

### Added

- `--ha-on-primary-color`, the label on a primary fill: white by default,
  Home Assistant's `--text-primary-color` under the HA skin, so a light
  primary in a custom theme keeps a readable button label. The primary
  button uses it.
- Under the HA skin, native controls (checkboxes, radios, scrollbars) take
  `accent-color: var(--ha-primary-color)` and the skin's `color-scheme`.
  A host on another skin (`data-skin` other than `ha`) is unchanged.

### Deprecated

- The comma form of `Item.Fields` for a `Multi` column, announced in
  v0.2.1 for removal in this release, stays accepted; its removal moves to
  the next release that carries breaking changes for go-fabric-server and
  go-ha2fabric.

## [0.2.1] - 2026-10-27

### Fixed

- The module compiles on 32-bit targets again: two range checks in the
  controller handlers compared an `int` with `0xFFFFFFFF`, which overflows
  where `int` is 32 bits (found by go-fabric-server's armv7 add-on image,
  report S4). CI now builds with `GOARCH=386`. The comma form of
  `Item.Fields` for a `Multi` column stays accepted in this release; its
  removal moves to v0.2.2, after go-fabric-server v0.4.0 has bumped.

## [0.2.0] - 2026-10-27

The controller page family (wave U2, brief U2): every go-fabric
embedding that runs a controller gets the same pages for its nodes, as
every node embedding gets the node pages. The HTTP contract goes to 0.2.0
with the controller section; the node section is unchanged.

### Added

- **Contract 0.2.0**: the controller routes under `/api/v1/controller/`
  (status, nodes, a node's detail, decommission, interview,
  enable/disable, auto-subscribe, the commissioning window, targeted
  attribute reads, a generic write and invoke, ACL, bindings, ICD, the
  node's fabrics, commissioning, discovery, credentials, the fabric label,
  the update queue, the stream) with request and response schemas, hex
  ids, `x-admin` flags and the problem code `node_error` (502). Fixtures
  for every controller response.
- `fabricui.ControllerSource` and `ControllerNode`, the port with its own
  value types (no go-fabric type in it), and `Options.Controller`; nil
  serves no controller family (`/controller/*` answers 404).
- `fabricui/gofabric.ControllerOf(*controller.Controller,
  ControllerOptions)`, the adapter over go-fabric's controller facade;
  `gofabric.Hooks` deliver the nodes' transitions to the stream
  (`controller.Config.HooksFor`, `Wrap` keeps the host's own hooks;
  `Close` before the controller's stops their timers);
  `gofabric.ThreadDatasetName` names a Thread dataset by go-fabric's
  decoder.
- The controller handlers are generated from the contract with
  oapi-codegen (pinned as a `go.mod` tool; `make generate`,
  `make generate-check` in CI); the handler bodies are hand-written.
- `fabricui.NewMemoryController` and `fabricui.NewMemoryCredentials`,
  in-memory implementations for tests and demonstrations.
- npm: the pages `ControllerNodes`, `ControllerCommission`,
  `ControllerNode` (tabs overview, endpoints, attributes, ACL, bindings,
  window, ICD, fabrics, updates) and `ControllerSettings`;
  `createControllerStore` (`ui.controller`); the client's controller
  methods; `subscribeControllerStream`; the routes `controllerNodes`,
  `controllerCommission`, `controllerSettings`, `controllerNode`; i18n
  keys `controller.*` and `shell.*` in English and German.
- The standalone shell shows the families the host serves, by presence
  (`/node/status`, `/controller/status`; a 404 hides a family, as does a
  disabled node next to a controller), with hash routes
  `#/controller/<page>` and `#/controller/node/<id>`.

### Changed

- go-fabric `v0.12.2` → `v0.13.3`.
- The contract guard lets an operation marked `x-status: planned` drop
  its placeholder responses when it is served; its path and method stay
  guarded.

### Deprecated

- The comma-separated `Item.Fields[key]` for a `Multi` column or filter
  is removed in v0.2.1, after go-fabric-server and go-ha2fabric have
  bumped; fill `Item.Lists` instead.

### Removed

- `Events.Observers` (deprecated since v0.1.2): the stream observes the
  node itself. A host that still calls it drops the call.

## [0.1.3] - 2026-10-26

List-valued item fields travel as JSON arrays, so a value with a comma —
a label named "Living, dining" — stays one chip and one filter option
(issue #10, brief F2). The HTTP contract only gains an optional field;
hosts that fill `Fields` alone keep working unchanged.

### Added

- `Item.Lists map[string][]string`, answered as `lists` on
  `GET /node/exposure` and `POST /node/exposure/preview` (absent when the
  provider leaves it empty): the values of a list-valued field, one entry
  per value, no separator semantics. A `Multi` column or filter reads
  `Lists[key]`; when the key is in both `Lists` and `Fields`, `Lists`
  wins.
- `Item.ListValues(key)` returns the list as the page reads it.

### Deprecated

- The comma-separated `Item.Fields[key]` for a `Multi` column or filter.
  It stays accepted as the fallback when `Lists` has no such key until
  v0.2.0; `Fields[key]` may keep carrying a display string.

## [0.1.2] - 2026-10-25

The handlers read go-fabric v0.12.2's facade routes instead of the U1
workarounds (brief U1b, go-fabric report CK Q2). The HTTP contract only
gains optional fields; the config-time observer wiring is deprecated.

### Added

- `GET /node/status` answers `window_kind` (`none`, `own`, `basic`,
  `enhanced`) and `window_deadline` (RFC 3339, the node's clock; absent
  while no window is open). The Status page and the pairing flow count the
  window down from the deadline, so a reload shows what is left.
- `GET /node/endpoints` answers `device_type_revision`, and every
  endpoint's clusters carry `name` and `revision` — parts and a host
  snapshotter's endpoints included, which listed no clusters before. The
  list is the mounted servers in ServerList order, Descriptor and
  BridgedDeviceBasicInformation among them.
- `GET /node/pairing` answers the configured `discriminator` and
  `passcode` (`Node.SetupPayload`), and `POST /node/pairing/window` with
  `Options.ConfiguredWindow` does too; both are admin routes.

### Changed

- Requires go-fabric v0.12.2. `NodeSource` gains `Window`,
  `SetupPayload`, `SessionDiagnostics`, `Clock` and `Observe` and drops
  `PairingCodes`, `Sessions` and `Subscriptions`; `*node.Node` satisfies
  it.
- The stream registers an observer with `Node.Observe` while it is open
  and removes it on close: a handler built after `Node.Start` reports
  commissionings, fabric removals and window transitions, each once
  (routes no longer publish them themselves).
- Fresh window credentials come from `node.NewWindowCredentials`; session
  idle ages from `Node.SessionDiagnostics` on the node's clock (0 for a
  session without traffic yet). The window's duration is
  `WindowSnapshot.Timeout`.
- The credentials of a window this handler opened are served while
  `Node.Window` reports that basic window (its discriminator, before its
  deadline), not on a timer of the handler's own.

### Deprecated

- `Events.Observers` returns its argument unchanged: `Node.Observe`
  replaces it, and a stream reports each transition once whether or not
  the host still wires it into `node.Config.Observers`. Delete that line;
  the method goes in v0.2.0. `Options.Events` stays for the host's own
  events.

## [0.1.1] - 2026-10-25

The first consumer's findings (go-ha2fabric, report H3): list-valued
fields and the host's configuration warnings (brief U1a). Additions only;
contract v1 keeps every field.

### Added

- **List-valued fields:** `Column.Multi` and `Filter.Multi` (JSON
  `multi`, optional, default false). A multi column shows its
  comma-separated value as chips; its value filter offers the individual
  values across all items instead of their combinations; a multi filter
  matches an item whose list holds the chosen value (values trimmed, an
  empty value never matches).
- **Host warnings:** `Options.Warnings func(ctx) []Warning` with
  `Warning{Code, Message, Since}`. `GET /node/status` answers them as
  `warnings` (absent without the hook, `[]` when it finds nothing); the
  Status page shows them in a card above the status card, the Selection
  page collapsed above the table. The npm client exports the `Warning`
  type.

## [0.1.0] - 2026-10-25

The node page family (wave U1 of concept v9), extracted from openccu-loom
at `b40b289a` and made host-neutral.

### Added

- **Contract v1** (`contract/openapi.yaml`): the node routes under
  `/api/v1/node/` — status, fabrics (list, remove), pairing (codes, open
  and close the commissioning window), endpoints, sessions, mDNS,
  compatibility, events, the selection (schema, list, preview, apply),
  maintenance (reassemble, factory reset with a written confirmation) and
  the server-sent event stream; the controller routes listed as planned.
  Golden fixtures under `contract/fixtures/` are read by the Go and the
  TypeScript tests; a guard fails on a removal against
  `contract/previous/`.
- **Go module** `fabricui`: `Mount` and `Handler` on the go-fabric node
  facade (`NodeSource`, which `*node.Node` satisfies), `ExposureProvider`
  for the host's selection, `Events` (wires `node.Observers` into the
  stream), the admin gate (`Options.Admin`, or a per-process token in
  token mode), the embedded bundle with SPA fallback and cache headers,
  an import guard holding the module on the facade, and an end-to-end test
  that commissions a fabrictest controller through a window the handler
  opened.
- **npm package** `@sukramj/fabric-ui`: the six node pages (Status, Pair,
  Fabrics with maintenance, Selection with preview before apply,
  Endpoints, Diagnostics), the UI kit, the design tokens with Home
  Assistant's variable names and the theme bridge, the client generated
  from the contract, the stream helper, English and German catalogues.
- **Standalone shell** built into `fabricui/dist/` and committed; CI
  rebuilds it and fails on any byte of difference.
- **Release**: one tag releases both artefacts; the npm publish runs when
  the `NPM_TOKEN` secret is set.
