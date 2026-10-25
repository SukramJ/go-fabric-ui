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
| `/node/status` | GET | – | 200 | `Health`, `LocalAddr`, `Topology`, `MDNSEnabled`, `CommissioningWindow`, `FabricInfos`; the provider's `List` for `enabled_count`; `Options.Warnings` for `warnings` |
| `/node/fabrics` | GET | – | 200, 500, 503 | `FabricInfos`, `eligibility.VendorName` |
| `/node/fabrics/{index}` | DELETE | yes | 204, 400, 403, 500, 503 | `RemoveFabric` |
| `/node/pairing` | GET | yes | 200, 403, 503 | `PairingCodes`, or the codes of the window this handler opened |
| `/node/pairing/window` | POST | yes | 200, 400, 403, 409, 500, 503 | `OpenBasicWindow` on fresh credentials and `PairingCodesFor`, or `CommissioningWindow().OpenOwnWindow` with `Options.ConfiguredWindow` |
| `/node/pairing/window/close` | POST | yes | 204, 403, 500, 503 | `CommissioningWindow().RevokeWindow` |
| `/node/endpoints` | GET | – | 200, 503 | `Topology` |
| `/node/sessions` | GET | – | 200, 503 | `Sessions` (session table, occupancy), `Subscriptions` (per session) |
| `/node/mdns` | GET | – | 200, 503 | `MDNSEnabled`, `MDNSServices`, `DiagnoseMDNS` |
| `/node/compatibility` | GET | – | 200, 503 | `FabricInfos`, `Topology`, `eligibility.Compat` |
| `/node/events` | GET | – | 200, 503 | `DiagnosticEvents` (newest first) |
| `/node/exposure/schema` | GET | – | 200, 500, 503 | provider `Schema` |
| `/node/exposure` | GET | – | 200, 500, 503 | provider `List` |
| `/node/exposure` | PUT | yes | 200, 204, 400, 403, 500, 503 | provider `Apply`, then `Options.Reassemble` |
| `/node/exposure/preview` | POST | yes | 200, 400, 403, 500, 503 | provider `Preview` |
| `/node/maintenance/reassemble` | POST | yes | 204, 403, 500, 503 | `Options.Reassemble` |
| `/node/maintenance/factory-reset` | POST | yes | 204, 400, 403, 500, 503 | `FabricInfos`, `RemoveFabric` for each |
| `/node/stream` | GET | – | 200 (`text/event-stream`) | `Options.Events`, the handler's own events |

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
  `GET /node/pairing` and `POST /node/pairing/window`, and only for a
  window this handler opened on fresh credentials. No event and no log
  line carries it; the tests assert both.
- **Stream**: `event:` is `fabric_added`, `fabric_removed`,
  `window_opened`, `window_closed`, `endpoints_changed` or
  `exposure_changed`; `data:` is JSON (`FabricEvent`, `WindowEvent`, `{}`,
  `ExposureEvent`). A comment line arrives every 15 s.

## Admin rights

A writing route (POST, PUT, DELETE) and `GET /node/pairing` run only for an
admin; anything else answers `403 forbidden`. Two modes:

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

## Versioning

The Go module and the npm package share one version. A minor version adds
paths, methods, responses, fields and optional request fields; it never
removes or renames one and never makes an optional request field required.
`contract/previous/openapi.yaml` is the copy of the last release, and the
contract guard (`make contract-check`, CI job `contract`) fails on a
removal against it. The release commit that tags a version copies the
contract to `contract/previous/`.

## Controller routes (planned)

The controller family (wave U2) is listed in the contract with
`x-status: planned` and no schemas: `/controller/commission`,
`/controller/discovery`, `/controller/nodes`, `/controller/nodes/{id}`,
`/controller/fabric/label`.
