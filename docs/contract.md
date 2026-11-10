# The contract

`contract/openapi.yaml` (OpenAPI 3.1) is the one source of the HTTP API.
The Go handlers (`fabricui`) and the TypeScript client (`ui/src/lib/api`,
generated with `npm run gen:types`) are tested against the same golden
fixtures under `contract/fixtures/`: one request/response pair per route
and status. The Go tests compare the handler's answer byte for byte after
JSON normalisation; the TypeScript tests answer the client from the same
files. A test on each side fails when a fixture is used by no test on that
side.

Paths are relative to the mount prefix (`Options.Prefix`, default
`/api/v1`).

## Node routes (v1)

| Route | Method | Admin | Answers | Reads in go-fabric |
| --- | --- | --- | --- | --- |
| `/node/status` | GET | – | 200 | `Health`, `LocalAddr`, `Topology`, `MDNSEnabled`, `Window` (open, duration, `window_kind`, `window_deadline`), `FabricInfos`; the provider's `List` for `enabled_count`; `Options.Warnings` for `warnings`; `Surface` for `surface`; `AttestationError` for `attestation_error`, `BootReason` for `boot_reason`, `Options.Erase`/`NodeEntry.Erase` for `erase` (0.7.0) |
| `/node/fabrics` | GET | – | 200, 500, 503 | `FabricInfos`, `eligibility.VendorName` |
| `/node/fabrics/{index}` | DELETE | yes | 204, 400, 403, 500, 503 | `RemoveFabric` |
| `/node/pairing` | GET | yes | 200, 403, 503 | `SetupPayload`, or the credentials of the window this handler opened while `Window` still reports it |
| `/node/pairing/window` | POST | yes | 200, 400, 403, 409, 500, 503 | `NewWindowCredentials`, `PairingCodesFor` and `OpenBasicWindow`, or `CommissioningWindow().OpenOwnWindow` and `SetupPayload` with `Options.ConfiguredWindow`; `Window` for the duration |
| `/node/pairing/window/close` | POST | yes | 204, 403, 500, 503 | `Node.CloseWindow` |
| `/node/endpoints` | GET | – | 200, 503 | `Topology` (`Clusters`, `DeviceTypeRevision` of every endpoint, parts included) |
| `/node/endpoints/{ep}` | GET | – | 200, 400, 404, 500, 503 | `ReadCluster` (the Descriptor, then every cluster of its ServerList), `ReadAttribute` (the parts' PartsList), `Topology`, `Surface`, `schema` names — served from 0.4.1 |
| `/node/endpoints/{ep}/events` | GET | – | 200, 400, 503 | `Events` with the endpoint's filter, `since` and `limit` — served from 0.4.1 |
| `/node/surface` | GET | – | 200, 500, 503 | `Surface` — the `surface.Report` in go-fabric's own JSON encoding (`MarshalJSON`) |
| `/node/sessions` | GET | – | 200, 503 | `SessionDiagnostics` (sessions with subscription counts and idle ages on the node's clock, occupancy); `SessionParameters` for `parameters` (0.7.0: the intervals, revisions, `specification_version_hex` as 8 hex digits, paths per invoke, transports, `tcp` and `max_tcp_message_size` with TCP) |
| `/node/topology` | GET | – | 200, 500, 503 | `Topology`, `FabricInfos`, `SessionDiagnostics`, `Health`, `LocalAddr` → `TopologyGraph` (v0.4.0); 503 also before the node assembled (`Topology().Aggregator` 0) |
| `/node/mdns` | GET | – | 200, 503 | `MDNSEnabled`, `MDNSServices`, `DiagnoseMDNS` |
| `/node/mdns/reannounce` | POST | yes | 204, 403, 503 | `TriggerReannounce` (0.7.0; a no-op on an advertiser that cannot) |
| `/node/compatibility` | GET | – | 200, 503 | `FabricInfos`, `Topology`, `eligibility.Compat` |
| `/node/events` | GET | – | 200, 503 | `DiagnosticEvents` (newest first) |
| `/node/exposure/schema` | GET | – | 200, 500, 503 | provider `Schema` |
| `/node/exposure` | GET | – | 200, 500, 503 | provider `List` |
| `/node/exposure` | PUT | yes | 200, 204, 400, 403, 500, 503 | provider `Apply`, then `Options.Reassemble` |
| `/node/exposure/preview` | POST | yes | 200, 400, 403, 500, 503 | provider `Preview` |
| `/node/maintenance/reassemble` | POST | yes | 204, 403, 500, 503 | `Options.Reassemble` |
| `/node/maintenance/factory-reset` | POST | yes | 204, 400, 403, 500, 503 | `FabricInfos`, `RemoveFabric` for each |
| `/node/maintenance/erase` | POST | yes | 204, 400, 403, 500, 501, 503 | `Options.Erase` (0.7.0), the host's `Node.FactoryReset` and its way back to a node; body `{"confirm":"erase-node"}`; 501 `not_implemented` without one |
| `/node/stream` | GET | – | 200 (`text/event-stream`) | `Observe`, `OnHealthSample`, `OnDiagnosticEvent` (per stream), `Options.Events`, the handler's own events |

`GET /node/status` also reports `exposure_reason` (0.5.0): why the host
serves no selection (`Options.ExposureReason`), absent with a selection;
and from 0.6.0 `node_registry`, `admin` and `config` (see
[Node registry](#node-registry-060)).

From 0.7.0 the status reports `attestation_error` — the text of
`Node.AttestationError`, present when the node's attestation files did
not load under `BlockCommissioningOnError` and it refuses every
commissioning — and `boot_reason`, the GeneralDiagnostics BootReasonEnum
value (0–6) the node was built with; both absent without a node. `erase`
is present and true when the host wires an erase for the node. Two
maintenance routes differ: `POST /node/maintenance/factory-reset` removes
every fabric and the node keeps running; `POST /node/maintenance/erase`
runs the host's `Erase` — go-fabric's `Node.FactoryReset` erases the
store and the endpoint records and ends the node, and the host builds a
new one on the erased store or restarts. After its 204 the handler keeps
reading the `NodeSource` it was given, so a host that builds a new node
hands over a source that follows it (`fabricui.NodeSwitch`, 0.7.1); the
page polls `GET /node/status` until it answers again.

## The node selector (0.5.0)

| Route | Method | Admin | Answers | Reads |
| --- | --- | --- | --- | --- |
| `/nodes` | GET | – | 200 | `Options.Node` (key `default`) and `Options.Nodes`, the default first: `{nodes: [{key, label, kind?, enabled}]}` |
| `/nodes/{node}/…` | as `/node/…` | as `/node/…` | as `/node/…`, and 404 for a key `/nodes` does not list | the node of that key |

Every node route is served twice: under `/node/…` for the **default node**
— `Options.Node`, else the first of `Options.Nodes` — and under
`/nodes/{node}/…` for every node by its key, the default included; the
operation ids of the second set carry the suffix `ForNode`. A key matches
`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$` and is no page name of the node family
(`status`, `pair`, …), so the shell's hash route `#/node/<key>/<page>` stays
unambiguous. The selection (`Options.Exposure`), `Options.Reassemble` and
`Options.Events` belong to the default node; another node answers its
exposure routes 503 and reports `exposure: false` with the host's
`exposure_reason` — unless its `NodeEntry` carries a selection of its own
(0.6.0). One admin token covers every node. A node listed
without a source (`NodeEntry.Source` nil) is `enabled: false`, and its
routes answer as the family's without a node. The standalone shell shows a
node selector in its header when the list holds more than one node.

## Node registry (0.6.0)

| Route | Method | Admin | Answers | Reads |
| --- | --- | --- | --- | --- |
| `/nodes` | GET | – | 200 | `Options.Node` (key `default`), then `Options.Registry.List()` on every request: `{nodes: [{key, label, kind?, enabled, managed, exposure}]}` |
| `/nodes` | POST | yes | 201 `NodeSummary`, 400, 403, 409 `node_limit`, 500, 503 | `NodeRegistry.Create(ctx, NodeSpec{Label})` with `{label}` |
| `/nodes/{node}/label` | PUT | yes | 204, 400, 403, 404, 409 `node_is_default`, 500, 503 | `NodeRegistry.Rename(ctx, key, label)` with `{label}` |
| `/nodes/{node}` | DELETE | yes | 204, 403, 404, 409 `node_is_default`, 500, 503 | `NodeRegistry.Delete(ctx, key)` |

The first node is the host's configuration (`Options.Node`, the key
`default`); every further node is registry state in the host's store,
created, renamed and deleted through the page. `Options.Registry` needs
`Options.Node` and excludes `Options.Nodes` — `Handler` panics and
`Mount` fails otherwise. With it `GET /nodes` lists the default and then
the registry's nodes, each `managed: true`; an entry whose key is invalid,
a page name or already listed is logged and left out. A registry node's
routes answer under `/nodes/{node}/…` from the moment `List` names it, and
its handler keeps its state (the window it opened, its stream's
subscribers) while the entry's `Source`, `Exposure` and `Events` stay the
same values. Without a registry the three writing routes answer 503 and
everything else is 0.5.0.

- **Label** (`NodeLabelRequest`): trimmed, 1 to 64 characters, no control
  character; an unknown field is 400. The host allocates the key (a slug
  of the label, unique, of the key pattern above), the port, the
  discriminator, the passcode and the store, and starts the node in
  `Create`.
- **Errors**: a host error wrapping `fabricui.ErrNodeLimit` is 409
  `node_limit` with its text, one wrapping `fabricui.ErrDefaultNode` 409
  `node_is_default`, any other 500 without its text. The handler answers
  409 `node_is_default` itself for a configured node (the default), and 404
  for a key the registry does not list. A key `Create` returns that the
  pattern refuses is 500.
- **Delete is final**: the host stops the node and removes its fabrics —
  every controller that paired it loses it —, its store and its
  selection. The shell asks twice, naming the label and the fabric count.
- **`nodes_changed`** follows each of the three writes on every node's
  stream, `{keys: []string}`: a page reloads `GET /nodes`.
- **Per-node selection**: `NodeEntry.Exposure`, `Reassemble` and `Events`
  (a registry's entries and `Options.Nodes` alike) are that node's
  selection, rebuild and host events under `/nodes/{node}/…`;
  `NodeSummary.exposure` and the node's `Status.exposure` report it. The
  Options-level ones stay the default node's.
- **Status**: `node_registry: true` with a registry (absent without), so
  the shell offers "Create node…"; `admin` (host mode only) whether the
  host's admin check admits this request — in token mode `csrf_token`
  says it.

### The configuration block

`GET /node/status` and `GET /nodes/{node}/status` carry `config` when the
host describes the node (`Options.NodeConfig` for the default,
`NodeEntry.Config` for any other; nil omits it, so a 0.5.0 host is
unchanged). One shape for every node, read-only:

| Field | Meaning |
| --- | --- |
| `source` | `configuration` (the host's configuration; nothing editable in the page) or `registry` (allocated by the host; only the label is editable, `PUT /nodes/{node}/label`). Empty in the view takes the node's own. |
| `source_note` | The host's wording of where a configured node's values come from ("add-on options", "the --bridge-* flags"); optional. |
| `label` | The configured name; a registry node's is its entry's label, so a rename shows at once. Optional. |
| `listen` | The address the node listens on, as configured. |
| `discriminator`, `vendor_id`, `product_id` | Numbers; `vendor_name`, `product_name` optional. |
| `store` | Where the node keeps its state (a file name); optional. |

It never carries the passcode — that stays with the admin route `GET
/node/pairing`.

## Operation (0.5.0)

| Route | Method | Admin | Answers | Reads |
| --- | --- | --- | --- | --- |
| `/operation/info` | GET | – | 200, 503 | `Operations.Info`: `{version, go_fabric, uptime_s, started_at}` |
| `/operation/log/level` | GET | – | 200, 503 | `Operations.LogLevel`: `{level, levels[]}` |
| `/operation/log/level` | PUT | yes | 200, 400, 403, 500, 503 | `Operations.SetLogLevel` with `{level}` |
| `/operation/log` | GET | – | 200, 400, 503 | `Operations.Log(since, limit)`: `{records: [{seq, time, level, msg, attrs}]}`, oldest first; `limit` 1..2000, default 200 |
| `/operation/log/stream` | GET | – | 200 (`text/event-stream`), 503 | `Operations.ObserveLog`: `event: log_record` with a `LogRecord`, a comment line every 15 s, the oldest dropped for a client 64 records behind |

The levels are slog's four, lower-case: `debug`, `info`, `warn`, `error`;
a level between two names reads as the lower one. Every route answers 503
`service_unready` while the host attaches no `Options.Operations`; the
shell then hides the Operation tab. The page belongs to neither family: a
daemon with a controller only has it as well.

A route whose dependency the host did not wire answers
`503 service_unready`: no node (`Options.Node` nil), no selection
(`Options.Exposure` nil), no `Options.Reassemble`. `GET /node/status`
answers 200 with `enabled: false` instead, because the status is what tells
a page the family is off.

Details the fixtures pin:

- **Window duration** 180..900 s; 0 or absent is 900 (Matter §11.19.8.1).
  409 when a window is open or commissioning is blocked; 503 with
  `Retry-After: 30` while the node exposes no endpoint (a controller that
  subscribes then caches an empty parts list), unless
  `Options.AllowEmptyTopology`.
- **Factory reset** needs the body `{"confirm":"remove-all-fabrics"}`; a
  generic confirmation or an empty body answers 400 and removes nothing. A
  partial reset answers 500 naming how many fabrics went.
- **Selection writes**: an empty batch is 204 for `PUT` and an empty list
  for the preview. A provider error wrapping `fabricui.ErrUnknownItem` or
  `fabricui.ErrInvalidUpdate` is 400 with its text; any other is 500.
  After `Apply` the handler runs `Options.Reassemble`; a failed reassembly
  is logged and does not fail the persisted change.

## Identifiers, secrets, events

- **64-bit ids** (fabric id, node id, session peer and local node id)
  travel exactly as 16-digit uppercase hex in `*_hex`; the numeric twins
  are informational — a JSON number carries 53 bits and a browser rounds
  larger ids.
- **Secrets**: the passcode appears only in the answers of the admin routes
  `GET /node/pairing` and `POST /node/pairing/window` — the configured one
  (`Node.SetupPayload`) or that of the window this handler opened on fresh
  credentials. No event and no log line carries it; the tests assert both.
- **Stream**: `event:` is `fabric_added`, `fabric_removed`,
  `fabric_updated` (0.7.0), `commissioning_complete` (0.7.0),
  `window_opened`, `window_closed`, `endpoints_changed`,
  `exposure_changed`, `surface_changed`, `session_opened`,
  `session_closed`, `subscriptions_changed`, `attribute_changed`
  (0.4.1), `cluster_event` (0.7.0), `health_sample` or `diagnostic_event`
  (0.5.0); `data:` is JSON
  (`FabricEvent`, `WindowEvent`, `{}`, `ExposureEvent`,
  `SurfaceChangedEvent`, `SessionEvent`, `AttributeChangedEvent`,
  `EventRecord`, `HealthSampleEvent`, `DiagnosticEvent`), and
  `nodes_changed` (0.6.0, `NodesChangedEvent`, see
  [Node registry](#node-registry-060)). A comment line arrives every 15 s.
  The fabric, window and surface events come from
  the node (`Node.Observe`, registered per stream), once per transition,
  whoever caused it. A commissioning is `fabric_added` — the AddNOC that
  installs the fabric — then `commissioning_complete`, go-fabric's
  `Observers.OnCommissioningComplete` after a successful
  CommissioningComplete, with `fabric_index` alone. `fabric_updated` is
  `Observers.OnFabricUpdated` (an UpdateNOC, UpdateFabricLabel or
  SetVIDVerificationStatement), its payload as `fabric_added`'s.
  `cluster_event` is `Observers.OnEvent`: every event a cluster server
  emitted and the node's event log recorded, in the shape
  `GET /node/endpoints/{ep}/events` lists it, delivered on the emitter's
  goroutine. `surface_changed` is go-fabric's
  `Observers.OnSurfaceChanged`: one event per endpoint whose served surface
  an assembly changed, after the node recorded the change in
  `Surface().Changed` and on the assembly's goroutine; `endpoints` names
  that endpoint's key (`""` for the root endpoint). It arrives for a
  reassembly this handler runs (`POST /node/maintenance/reassemble`,
  `PUT /node/exposure`) and for one the host runs on its own; the routes
  publish `endpoints_changed` alone. The session events are go-fabric's
  `Observers.OnSessionOpened`, `OnSessionClosed` and
  `OnSubscriptionsChanged`: CASE sessions only (a commissioning's PASE
  session raises none), the payload the session as `/node/sessions` lists
  it with its subscription count after the change. `attribute_changed` is
  `Observers.OnAttributeChanged`: one event per cluster of one endpoint
  whose attributes a controller's write, a command or the host changed,
  `{endpoint, cluster_hex, attributes[]}` with the attribute ids in hex
  and never a value — a page refetches what it shows. A burst of changes
  is a burst of events; a stream that falls 64 events behind drops the
  oldest. `health_sample` (0.5.0) is every verdict of the node's health
  probe (`Node.OnHealthSample`): `{healthy, note, sampled_at}`, the time
  the handler received it on the node's clock — the status page reloads on
  a burst of them. `diagnostic_event` (0.5.0) is every entry the node adds
  to the trace `GET /node/events` serves (`Node.OnDiagnosticEvent`), in its
  shape — the diagnostics page reloads on a burst of them. Both callbacks
  are registered per stream and removed when it closes.

## The endpoint detail and the event log (0.4.1)

`GET /node/endpoints/{ep}` answers the `EndpointDetail` the controller
family's inventory answers (one type for both families), filled from what
the node serves, read through its own read path as a controller's read
reaches it (`Node.ReadCluster`, no access control): the Descriptor's
DeviceTypeList with each revision and its TagList (empty while the node
serves none, which is every bridged endpoint's case today), the FixedLabel
then UserLabel entries, every server cluster of the ServerList in id order
as a `ClusterDetail` **with** its values (JSON by field name, as
`AttributeValue.value`), its data version, its feature map with the
features the schema names for its bits, and the attributes, accepted and
generated commands and events named by go-fabric's schema. A bridged
endpoint adds its name (the NodeLabel it serves), Reachable, its `scope`,
its parent and its parts from `Node.Topology`; the root endpoint and the
Aggregator, which the topology does not list, are named by their first
device type, and their `parts` are the PartsList entries that no other
entry's PartsList names — the direct children, as everywhere in
`EndpointDetail`. `decisions` are the surface report's decisions on the
endpoint: the root's under the key `""`, the Aggregator's under
`aggregator`, a bridged endpoint's under its stable key. A cluster whose
read fails is listed with the schema's name and revision and no values.
404 for an endpoint the node does not serve, 503 before the node started.

`GET /node/endpoints/{ep}/events` and `GET /controller/nodes/{id}/events`
answer an `EventRecordList` of `EventRecord{number, priority, endpoint,
cluster_hex, event_hex, name?, epoch_ms?, data?}`, oldest first:
`priority` is `debug`, `info` or `critical`, `name` the schema's, `data`
the event's fields by name — absent without data or when they do not
render. `since` is the lowest event number returned (a page passes the
last number it shows plus one), `limit` keeps the newest (default 100,
1 to 1000, else 400). The node's log (`Node.Events`) is bounded and
volatile, emptied when the node stops; an endpoint it no longer serves
keeps its events there, so an unknown endpoint is an empty list. The
controller's history (`client.Node.EventHistory`) holds the events the
node's default subscription, its bootstrap read and the read after
commissioning delivered — the newest 100 by default, in memory only; a
host's own reads do not reach it, and an unknown node is 404. An event
the node delivers raises the node's `node_updated`, so the events tab
fetches the newer ones on it.

## The topology graph

`GET /node/topology` and `GET /controller/topology` (from U3b) answer a
`TopologyGraph{as_of, family, nodes[], edges[]}`. A node has an `id`, a
`kind`, a `label`, a `state` (`ok`, `unreachable`, `idle`, `unknown`), a
`ref{page, id}` to the page that already lists it and an ordered
`details[]` of `{key, value}` a page shows without knowing the kind; the
typed fields (`endpoint`, `fabric_index`, `session_id`, `node_id_hex`,
`reachable`, `is_pase`, `subscriptions`, `role`, `rloc16`, `ssid`,
`bssid`) are present only for the kinds they belong to. The node family's
graph:

- `node` (state from the health sample) `contains` `ep:<aggregator>`,
  which `contains` each device (`ep:<n>`, kind `device`, state from
  `Reachable`), which `contains` its parts (kind `part`);
- each `fabric:<index>` `contains` the CASE sessions on it
  (`session:<id>`), and every session has a `session` edge to `node`; a
  PASE session (fabric index 0) hangs on `node` alone;
- a session is `idle` when its peer has been quiet longer than the idle
  interval the peer reported in its handshake;
- refs: `status` (the node), `endpoints` (endpoint number), `fabrics`
  (fabric index), `diagnostics` (session id); `controller_node`
  (`node_id_hex`) on the controller's graph.

The controller family's graph (go-fabric's `controller/topology`, through
`fabricui/gofabric`):

- `node` is the controller (labelled by its fabric label, else
  `Controller`; details `nodes`, `collected_at`) with a `session` edge to
  every commissioned node, `peer:<node_id_hex>` (state `ok` or
  `unreachable`; `role` its Thread or Wi-Fi role, else its network — an
  Ethernet node is a `peer` with role `ethernet`, not the `ethernet`
  kind; details `network`, `role`, `rloc16`, `ext_address`,
  `ext_pan_id`, `network_name`, `ssid`, `bssid`, `bridge`, `addresses`,
  each only when known);
- a Thread neighbour no commissioned node matches is
  `unknown_<extended address>`, `thread_router` when a node saw it
  rx-on-when-idle, else `thread_end_device` (details `ext_address`,
  `ext_pan_id`, `role`, and `network_name` when go-fabric fills it — its
  own, or from 0.8.0 the border-router registry's for its extended PAN
  id); an access point is `ap_<BSSID>`, `wifi_ap`, labelled by its SSID;
  both have state `unknown`;
- a Thread neighbour go-fabric's border-router registry knows (0.8.0,
  go-fabric `topology.KindBorderRouter`, issue #56) is
  `br_<extended address>`, `border_router`, labelled by its mDNS host
  name without the trailing dot and `.local` (go-fabric strips it as
  matterjs-server does), else by its extended address; state `ok` while
  the registry saw it within its 24-hour expiry (`brdiscovery.StaleRetention`)
  before the snapshot, else `unknown`; details `host_name`, `vendor_name`,
  `model_name`, `network_name`, `ext_address`, `ext_pan_id`, `last_seen`
  (RFC 3339, UTC), each only when known. The Network page draws it square
  and dashed and names every kind it draws in a legend under the graph;
- a node of a kind the adapter does not know yet (one go-fabric adds
  later) is never dropped: it is drawn by its network — on Thread as a
  neighbour is, on Wi-Fi as `wifi_ap`, otherwise as `ethernet` — with
  state `unknown` and its go-fabric kind as the first detail, `kind`;
- a `link` per Thread pair and Wi-Fi station carries `strength`
  (`strong`, `medium`, `weak`, `none` — observed dead —, `unknown` —
  no measurement), the `lqi` (Thread) and `rssi` of the direction the
  strength comes from, and a `label` with both directions when they
  differ.

Subscriptions are counts per session; per-session subscription paths,
partitions and tags wait for go-fabric (brief U3a decision 0.4).

## Admin rights

A writing route (POST, PUT, DELETE), `GET /node/pairing` and every
controller route marked `x-admin` (among them `GET /controller/discovery`)
run only for an admin; anything else answers `403 forbidden`, before the
request's parameters or body are read. `GET /controller/status` serves the
token of token mode as `csrf_token` too. Two modes:

- **Host mode** — `Options.Admin` decides per request. A host behind an
  authenticating proxy (a Home Assistant Ingress, where the proxy's source
  address is the check) answers from what the proxy vouches for.
- **Token mode** — `Options.Admin` is nil. The handler draws a random token
  per process and serves it as `csrf_token` in `GET /node/status`; an admin
  request carries it in the `X-Fabric-UI-Token` header. A page of another
  origin in the same browser cannot read the status, so it cannot forge an
  admin request. The standalone shell does this on its own.

## Problems

Errors are RFC 9457 `application/problem+json` with `type`, `title`,
`status`, `detail`, `instance` and `code`; the code also travels in
`X-Problem-Code`. A 5xx problem carries no detail — the cause, which can
name paths, goes to the log.

### problem-bad_request
The request body or a path parameter is invalid; `detail` says what.

### problem-forbidden
The request needs admin rights (see above).

### problem-conflict
A commissioning window is already open, or commissioning is blocked.

### problem-not_found
No such route under the API prefix.

### problem-service_unready
A dependency is not wired, or the node is not ready for the action.

### problem-internal
The node or the provider failed; the log has the cause.

### problem-payload_too_large
The request body exceeds 1 MiB.

### problem-node_error
The interaction with a controller's node failed (502); `detail` is the
node's answer.

### problem-node_limit
`POST /nodes`: the host serves no further node (409, 0.6.0); `detail` is
the host's text.

### problem-node_is_default
`PUT /nodes/{node}/label` or `DELETE /nodes/{node}` on a node that is the
host's configuration rather than registry state (409, 0.6.0).

### problem-terms_required
`POST /controller/commission`: the device requires an accepted version of
its terms and conditions and the request sent no acknowledgement, or one
below that version (409, 0.8.0). go-fabric ends such a commissioning after
its information step, before ArmFailSafe and AddNOC
(`commissioner.TermsRequiredError`), so the device stays commissionable.
Besides the problem's members it carries `version` (the device's
TCMinRequiredVersion), `acknowledgements` (always `null`: no device
attribute names the bits — the terms the user accepts set them) and
`deadline` (the device's TCUpdateDeadline in Matter epoch seconds, from
2000-01-01T00:00:00Z; `null` when the attribute is null). A second request
with `terms_and_conditions` of that version or later commissions the
device. The port's error is `*fabricui.ErrTermsRequired`.

### problem-not_implemented
The host does not serve the route (501): `GET /controller/topology`
from a host without a topology service, and `POST
/node/maintenance/erase` from a host that wires no `Erase` (0.7.0). Its
detail is a fixed sentence, never an error's text.

## Versioning

The Go module and the npm package share one version. A minor version adds
paths, methods, responses, fields and optional request fields; it never
removes or renames one and never makes an optional request field required.
A patch may add the optional field or event field a fix needs to state
its behaviour; it adds no route and no problem code.
`contract/previous/openapi.yaml` is the copy of the last release, and the
contract guard (`make contract-check`, CI job `contract`) fails on a
removal against it. The guard also fails on a schema a response is decoded
through that sets `additionalProperties: false` — a generated client turns
it into a strict model, and the next added field, a minor here, breaks
every installed client (issue #27); only a request body, which the handler
decodes and whose unknown fields it rejects, may be closed. Its tests check
every fixture's response against the route's schema, so a field an encoder
writes and the contract does not name fails there. An operation marked
`x-status: planned` promises no answer: when it is served its placeholder responses may go, its path and
method may not. The release commit that tags a version copies the
contract to `contract/previous/`.

Contract 0.7.0 (wave U7) is additive over 0.6.0, which
`contract/previous/` holds as release 0.6.1 shipped it: two admin routes
per node (`mdns/reannounce`, `maintenance/erase`), three status fields,
the sessions' `parameters`, three node stream events,
`CommissionRequest.terms_and_conditions` and the controller events'
`step` and `update_*` fields. No problem code is new.

Contract 0.7.1 (wave U8) is additive over 0.7.0, which
`contract/previous/` holds as release 0.7.0 shipped it:
`ControllerEvent.ble`, `Credential.default`, two descriptions
(`ControllerCapabilities.ble`, `CommissionRequest.network_only`). No
problem code is new.

Contract 0.8.0 (wave U9) is additive over 0.7.1, which
`contract/previous/` holds as release 0.7.1 shipped it: the problem code
`terms_required` (409 on `POST /controller/commission`) with its
`version`, `acknowledgements` and `deadline` members on `Problem`, the
`commissioning_progress` failure kind `terms_required`, `Status.counters`
with the `Counter` schema, and the topology node kind `border_router`.

Contract 0.8.1 (wave U10) is additive over 0.8.0, which
`contract/previous/` holds as release 0.8.0 shipped it:
`NodeRecord.operational_host`, the operational SRV target that owned the
node's address when it was last reached (go-fabric 0.20.0's
`client.Record.OperationalHost`), absent when unknown or stored before
go-fabric 0.20.0, and not `hostname`, the commissionable advertisement's
host. No problem code is new.

## Controller routes (v0.2.0)

Served only when the host passes `Options.Controller`; without it every
`/controller/*` path answers `404 not_found`, which is how the shell hides
the family. `{id}` is the node id as hex (`node_id_hex`; case and leading
zeros free). The handlers' router, path and query binding, request bodies
and response types are generated from the contract (`internal/api`,
`make generate`, guarded by `make generate-check`); the handler bodies are
hand-written.

| Route | Method | Admin | Answers |
| --- | --- | --- | --- |
| `/controller/status` | GET | – | 200 |
| `/controller/nodes` | GET | – | 200 |
| `/controller/nodes/{id}` | GET | – | 200, 400, 404 |
| `/controller/nodes/{id}/endpoints/{ep}` | GET | – | 200, 400, 404 — served from 0.4.0 |
| `/controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}` | GET | – | 200, 400, 404, 503 — served from 0.4.0 |
| `/controller/nodes/{id}/events` | GET | – | 200, 400, 404 — served from 0.4.1 |
| `/controller/nodes/{id}` | DELETE | yes | 204, 400, 403, 404, 500, 502 |
| `/controller/nodes/{id}/interview` | POST | yes | 204, 400, 403, 404, 502 |
| `/controller/nodes/{id}/ping` | POST | yes | 200, 400, 403, 404 — served from 0.5.0 |
| `/controller/nodes/{id}/enable`, `/disable` | POST | yes | 204, 400, 403, 404, 500 |
| `/controller/nodes/{id}/auto-subscribe` | POST | yes | 204, 400, 403, 404, 500 |
| `/controller/nodes/{id}/window` | POST | yes | 200, 400, 403, 404, 502 |
| `/controller/nodes/{id}/window/close` | POST | yes | 204, 400, 403, 404, 502 |
| `/controller/nodes/{id}/attributes` | POST | – | 200, 400, 404, 502 |
| `/controller/nodes/{id}/attributes` | PUT | yes | 200, 400, 403, 404, 502 |
| `/controller/nodes/{id}/commands` | POST | yes | 200, 400, 403, 404, 502 |
| `/controller/nodes/{id}/acl` | GET / PUT | – / yes | 200, 400, (403), 404, 502 |
| `/controller/nodes/{id}/bindings` | GET / PUT | – / yes | 200, 400, (403), 404, 502 |
| `/controller/nodes/{id}/icd` | GET | – | 200, 400, 404 |
| `/controller/nodes/{id}/icd/register`, `/unregister` | POST | yes | 200, 400, 403, 404, 409, 502 |
| `/controller/nodes/{id}/fabrics` | GET | – | 200, 400, 404, 502 |
| `/controller/nodes/{id}/fabrics/{index}` | DELETE | yes | 204, 400, 403, 404, 502 |
| `/controller/commission` | POST | yes | 200, 400, 403, 404, 409 (0.8.0), 502, 503 |
| `/controller/discovery` | GET | yes | 200, 400, 403, 500 |
| `/controller/credentials` | GET | – | 200, 500, 503 |
| `/controller/credentials` | PUT / DELETE | yes | 204, 400, 403, (404), 500, 503 |
| `/controller/fabric/label` | PUT | yes | 204, 400, 403, 500 |
| `/controller/updates` | GET | – | 200 |
| `/controller/updates/consent`, `/check` | POST | yes | 200/204, 400, 403, 404, 500, 503 |
| `/controller/topology` | GET | `refresh=true` only | 200, 403, 501, 502 — served from 0.3.1; 501 `not_implemented` from a host without a topology service |
| `/controller/stream` | GET | – | 200 (`text/event-stream`) |

Details the fixtures pin:

- **Reads are targeted** (owner decision 0.4): `POST …/attributes` takes
  1 to 32 concrete paths (`endpoint`, `cluster_hex`, `attribute_hex`); no
  wildcard. A path the node refuses answers its status in the entry. A
  generic write and an invoke are admin routes; a refusal by the node is
  the `status` of a 200, not a problem. A read-only inventory from the
  controller's image is served by the endpoint and cluster routes (0.4.0,
  owner decision U4 0.1): nothing is read from the node for them.
- **The inventory** (0.4.0): `GET …/endpoints/{ep}` answers an
  `EndpointDetail` — the Descriptor's DeviceTypeList in its order with
  each revision, its TagList (empty without one), the FixedLabel then
  UserLabel entries, its parts (the direct children on the image's tree),
  BridgedDeviceBasicInformation Reachable (true on an endpoint that does
  not serve it), the device type `conformance` verdict (`unknown`,
  `incomplete`, `clean`, `violated`, with the violations and the
  clusters not judged inside) and every server cluster as a
  `ClusterDetail` without values. `GET …/clusters/{cluster_hex}` answers
  that cluster with the values the image holds: revision,
  `feature_map_hex` with the features it sets (`code`, `title` where a
  definition names it, `bit`), `data_version`, the AttributeList with
  names and values, the accepted and generated commands and the
  EventList where the peer reports one, each named where the cluster has
  a generated definition. An endpoint or cluster the image lacks is 404;
  a cluster the ServerList names that the controller has not read yet,
  and any cluster while the controller has not read the endpoint's
  Descriptor, is 503 `service_unready`. A value that does not render is
  absent, not `null`. The page reloads both on the node's
  `node_updated`. `NodeEndpoint` carries `reachable` where the endpoint
  serves BridgedDeviceBasicInformation, and its `device_types` are the
  whole DeviceTypeList.
- **Values** are JSON by the generated cluster definitions, a struct
  keyed by field name (`{"fabricIndex": 1}`); cluster, attribute, command
  and device type ids travel as `0x`-prefixed hex.
- **Ping** (0.5.0): `POST …/ping` is a Matter read of the node's
  BasicInformation DataModelRevision (0x0028/0x0000), not an ICMP echo —
  go-fabric-server's `ping_node` measures the same. `{method:
  "matter-read", per_address, reachable, rtt_ms?, error?, addresses[]}`:
  go-fabric's controller reads through the node's session and cannot aim a
  read at one address, so `per_address` is false and the one outcome
  stands for every address the record lists. A failed read is 200 with
  `reachable: false` and the reason, not an error.
- **ACL and bindings** are written as one list. A subject naming the node
  itself is left out of an ACL write, and an entry left without subjects
  dropped, as matterjs-server does. The page warns before a list that
  keeps no Administer entry for this controller.
- **Our fabric** is removed with `DELETE /controller/nodes/{id}`;
  `DELETE …/fabrics/{index}` with our index is 400.
- **Commissioning**: `{code}` or `{discovered, passcode}`, not both; a
  code that does not parse is 400, no device answering is 404, a node
  that fails (PASE, attestation) is 502 with the node's error as detail.
  `terms_and_conditions` `{version, acknowledgements}` (0.7.0, each
  0–65535, else 400) is the Terms and Conditions acknowledgement the
  commissioning sends with SetTCAcknowledgements, chip-tool's
  `--tc-acknowledgements-version` and `--tc-acknowledgements`; a device
  that requires terms fails the commissioning without it, or with a
  version below its own, with 409 `terms_required` (0.8.0; until 0.7.1 a
  502 at CommissioningComplete). The Commission page then opens the
  acknowledgement block with the device's version filled in, the
  acknowledgements field focused and the deadline as a date when the
  device names one, and its button sends the same commissioning again
  with the acknowledgement; the code or passcode is held for that one
  retry only. Bluetooth LE (0.7.1):
  a commissioning from a code searches over the host's central only when
  the code allows BLE (a manual code, or a QR code with the BLE bit),
  `network_only` is off and the request names a stored credential with
  `wifi_id` or `thread_id` — a device on BLE has no network yet, and no
  credential named means none, never the host's default. Otherwise it runs
  over the IP network and says why: the stream's `ble` reason is the first
  of `no_central`, `network_only`, `code_without_ble`, `no_network` that
  holds, and the go-fabric adapter logs `fabricui.pair.ble_skipped` with
  it when a central is not used. A commissioning of a discovered device
  carries no decision.
- **Credentials**: `GET` answers kind, id and name (SSID or Thread network
  name), never a passphrase or a dataset; `default: true` (0.7.1) marks
  the host's default entry, which the Commission page preselects (else
  the only entry).

## Identifiers, secrets and events of the controller family

- **Secrets**: a commissioning's code and passcode, a window's passcode
  and a credential's passphrase or dataset appear only in admin requests
  and the window's answer; no event and no log line carries them (the
  tests assert it).
- **Stream** (`GET /controller/stream`): `event:` is `node_added`,
  `node_updated`, `node_removed`, `connection`, `subscription`,
  `commissioning_progress` or `update`; `data:` is a `ControllerEvent`
  (`node_id_hex`, and `connection`, `active`, `phase`/`detail` by type).
  A commissioning's progress carries the failure's kind
  (`invalid_argument`, `no_device`, `timeout`, `node_failed`, and from
  0.8.0 `terms_required`), not its text. From 0.7.0 every step the controller reports
  (`PairOptions.Progress`, go-fabric `controller.PairingStep`: matter.js's
  step names plus Discovery 2.1, PASE 2.2 and Subscribe 21.1) is a
  `commissioning_progress` with `phase` `started` and `step`
  `{number, sub, name, phase}`, the step's phase `started`, `succeeded`,
  `failed` or `skipped` and no error text; a client that reads `phase`
  alone sees `started`. From 0.7.1 a commissioning from a code reports,
  once and before the first step, a `commissioning_progress` with `phase`
  `started` and `ble` `{used, reason}` — the reason word only, present
  when `used` is false. `update` comes from the source, not the handler
  (0.7.0): the go-fabric adapter observes the controller's update manager
  (`softwareupdate.Manager.Observe`) and emits each event with
  `update_kind` (`available`, `consent_changed`, `check_finished`,
  `applying`, `done`, `failed`, `queued`, `transfer_started`),
  `update_version` for the kinds that carry a target and `update_error`
  for `check_finished` and `failed`; a controller without an update
  manager emits none. `topology_changed` (`{}`) tells a page to reload
  `/controller/topology`. A comment line arrives every 15 s.
- **`node_error`** (502): the interaction with the node failed. Unlike
  the other 5xx problems it carries a detail — the node's answer, which
  names no host path.
