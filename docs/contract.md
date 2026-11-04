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
| `/node/status` | GET | – | 200 | `Health`, `LocalAddr`, `Topology`, `MDNSEnabled`, `Window` (open, duration, `window_kind`, `window_deadline`), `FabricInfos`; the provider's `List` for `enabled_count`; `Options.Warnings` for `warnings`; `Surface` for `surface` |
| `/node/fabrics` | GET | – | 200, 500, 503 | `FabricInfos`, `eligibility.VendorName` |
| `/node/fabrics/{index}` | DELETE | yes | 204, 400, 403, 500, 503 | `RemoveFabric` |
| `/node/pairing` | GET | yes | 200, 403, 503 | `SetupPayload`, or the credentials of the window this handler opened while `Window` still reports it |
| `/node/pairing/window` | POST | yes | 200, 400, 403, 409, 500, 503 | `NewWindowCredentials`, `PairingCodesFor` and `OpenBasicWindow`, or `CommissioningWindow().OpenOwnWindow` and `SetupPayload` with `Options.ConfiguredWindow`; `Window` for the duration |
| `/node/pairing/window/close` | POST | yes | 204, 403, 500, 503 | `Node.CloseWindow` |
| `/node/endpoints` | GET | – | 200, 503 | `Topology` (`Clusters`, `DeviceTypeRevision` of every endpoint, parts included) |
| `/node/endpoints/{ep}` | GET | – | 200, 400, 404, 500, 503 | `ReadCluster` (the Descriptor, then every cluster of its ServerList), `ReadAttribute` (the parts' PartsList), `Topology`, `Surface`, `schema` names — served from 0.4.1 |
| `/node/endpoints/{ep}/events` | GET | – | 200, 400, 503 | `Events` with the endpoint's filter, `since` and `limit` — served from 0.4.1 |
| `/node/surface` | GET | – | 200, 500, 503 | `Surface` — the `surface.Report` in go-fabric's own JSON encoding (`MarshalJSON`) |
| `/node/sessions` | GET | – | 200, 503 | `SessionDiagnostics` (sessions with subscription counts and idle ages on the node's clock, occupancy) |
| `/node/topology` | GET | – | 200, 500, 503 | `Topology`, `FabricInfos`, `SessionDiagnostics`, `Health`, `LocalAddr` → `TopologyGraph` (v0.4.0); 503 also before the node assembled (`Topology().Aggregator` 0) |
| `/node/mdns` | GET | – | 200, 503 | `MDNSEnabled`, `MDNSServices`, `DiagnoseMDNS` |
| `/node/compatibility` | GET | – | 200, 503 | `FabricInfos`, `Topology`, `eligibility.Compat` |
| `/node/events` | GET | – | 200, 503 | `DiagnosticEvents` (newest first) |
| `/node/exposure/schema` | GET | – | 200, 500, 503 | provider `Schema` |
| `/node/exposure` | GET | – | 200, 500, 503 | provider `List` |
| `/node/exposure` | PUT | yes | 200, 204, 400, 403, 500, 503 | provider `Apply`, then `Options.Reassemble` |
| `/node/exposure/preview` | POST | yes | 200, 400, 403, 500, 503 | provider `Preview` |
| `/node/maintenance/reassemble` | POST | yes | 204, 403, 500, 503 | `Options.Reassemble` |
| `/node/maintenance/factory-reset` | POST | yes | 204, 400, 403, 500, 503 | `FabricInfos`, `RemoveFabric` for each |
| `/node/stream` | GET | – | 200 (`text/event-stream`) | `Observe` (per stream), `Options.Events`, the handler's own events |

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
  `window_opened`, `window_closed`, `endpoints_changed`,
  `exposure_changed`, `surface_changed`, `session_opened`,
  `session_closed`, `subscriptions_changed` or `attribute_changed`
  (0.4.1); `data:` is JSON (`FabricEvent`, `WindowEvent`, `{}`,
  `ExposureEvent`, `SurfaceChangedEvent`, `SessionEvent`,
  `AttributeChangedEvent`). A comment line arrives every 15 s. The fabric, window and surface events come from
  the node (`Node.Observe`, registered per stream), once per transition,
  whoever caused it. `surface_changed` is go-fabric's
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
  oldest.

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
  Ethernet node is a `peer` with role `ethernet`, the `ethernet` kind is
  not produced; details `network`, `role`, `rloc16`, `ext_address`,
  `ext_pan_id`, `network_name`, `ssid`, `bssid`, `bridge`, `addresses`,
  each only when known);
- a Thread neighbour no commissioned node matches is
  `unknown_<extended address>`, `thread_router` when a node saw it
  rx-on-when-idle, else `thread_end_device`; an access point is
  `ap_<BSSID>`, `wifi_ap`, labelled by its SSID; both have state `unknown`;
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

### problem-not_implemented
The host's controller does not serve the route (501): today
`GET /controller/topology` from a host without a topology service. Its
detail is a fixed sentence, never an error's text.

## Versioning

The Go module and the npm package share one version. A minor version adds
paths, methods, responses, fields and optional request fields; it never
removes or renames one and never makes an optional request field required.
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
| `/controller/commission` | POST | yes | 200, 400, 403, 404, 502, 503 |
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
- **ACL and bindings** are written as one list. A subject naming the node
  itself is left out of an ACL write, and an entry left without subjects
  dropped, as matterjs-server does. The page warns before a list that
  keeps no Administer entry for this controller.
- **Our fabric** is removed with `DELETE /controller/nodes/{id}`;
  `DELETE …/fabrics/{index}` with our index is 400.
- **Commissioning**: `{code}` or `{discovered, passcode}`, not both; a
  code that does not parse is 400, no device answering is 404, a node
  that fails (PASE, attestation) is 502 with the node's error as detail.
- **Credentials**: `GET` answers kind, id and name (SSID or Thread network
  name), never a passphrase or a dataset.

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
  (`invalid_argument`, `no_device`, `timeout`, `node_failed`), not its
  text. `topology_changed` (`{}`) tells a page to reload
  `/controller/topology`. A comment line arrives every 15 s.
- **`node_error`** (502): the interaction with the node failed. Unlike
  the other 5xx problems it carries a detail — the node's answer, which
  names no host path.
