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
| `/node/pairing/window/close` | POST | yes | 204, 403, 500, 503 | `CommissioningWindow().RevokeWindow` |
| `/node/endpoints` | GET | – | 200, 503 | `Topology` (`Clusters`, `DeviceTypeRevision` of every endpoint, parts included) |
| `/node/surface` | GET | – | 200, 500, 503 | `Surface` — the `surface.Report` in go-fabric's own JSON encoding (`MarshalJSON`) |
| `/node/sessions` | GET | – | 200, 503 | `SessionDiagnostics` (sessions with subscription counts and idle ages on the node's clock, occupancy) |
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
  `exposure_changed` or `surface_changed`; `data:` is JSON (`FabricEvent`,
  `WindowEvent`, `{}`, `ExposureEvent`, `SurfaceChangedEvent`). A comment
  line arrives every 15 s. The fabric, window and surface events come from
  the node (`Node.Observe`, registered per stream), once per transition,
  whoever caused it. `surface_changed` is go-fabric's
  `Observers.OnSurfaceChanged`: one event per endpoint whose served surface
  an assembly changed, after the node recorded the change in
  `Surface().Changed` and on the assembly's goroutine; `endpoints` names
  that endpoint's key (`""` for the root endpoint). It arrives for a
  reassembly this handler runs (`POST /node/maintenance/reassemble`,
  `PUT /node/exposure`) and for one the host runs on its own; the routes
  publish `endpoints_changed` alone.

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

## Versioning

The Go module and the npm package share one version. A minor version adds
paths, methods, responses, fields and optional request fields; it never
removes or renames one and never makes an optional request field required.
`contract/previous/openapi.yaml` is the copy of the last release, and the
contract guard (`make contract-check`, CI job `contract`) fails on a
removal against it. An operation marked `x-status: planned` promises no
answer: when it is served its placeholder responses may go, its path and
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
| `/controller/stream` | GET | – | 200 (`text/event-stream`) |

Details the fixtures pin:

- **Reads are targeted** (owner decision 0.4): `POST …/attributes` takes
  1 to 32 concrete paths (`endpoint`, `cluster_hex`, `attribute_hex`); no
  wildcard, no inventory. A path the node refuses answers its status in
  the entry. A generic write and an invoke are admin routes; a refusal by
  the node is the `status` of a 200, not a problem.
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
  text. A comment line arrives every 15 s.
- **`node_error`** (502): the interaction with the node failed. Unlike
  the other 5xx problems it carries a detail — the node's answer, which
  names no host path.
