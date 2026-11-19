# Review phase 3 — fix wave step 0: the contract shape

Design note for the two implementers (G: Go, contract, docs; W: `ui/**`).
Tree: `review/p3` at `5272f7c`. Baseline `contract/previous/openapi.yaml`
= release v0.9.2's contract 0.9.0 = the current contract
(`contract/openapi.yaml:4`). Decisions: `notes/review/2026-11-17/decisions.md`
rows 13, 15, 19–56. Held rows (UR-P0-auth-20, UR-P0-web-21, UR-P1-adpt-60)
appear here by ID only; their shapes are in
`notes/review/2026-11-17/held/phase-3-contract-held.md` under the labels
**H-a … H-e** used below. Rows marked **OPEN-n** rest on a recommendation
(last section) until the owner confirms.

## 1. Per operation — `x-admin`

97 operations, 46 `x-admin` today; 43 become `x-admin`, 8 stay open.
A non-admin gets `403` with the existing problem code `forbidden`
(`docs/contract.md:395`), before parameters or body are read
(`docs/contract.md:370-372`).

### Re-gated (`-` → `x-admin: true`)

Node family: each row is two operations, `/node/…` and its
`/nodes/{node}/…` twin (`…ForNode`); lines are `contract/openapi.yaml`.

| Path (GET) | operationId (+ `ForNode`) | Lines | Decision (class) |
| --- | --- | --- | --- |
| `/node/fabrics` | `listFabrics` | 52, 589 | 35 (B) |
| `/node/endpoints` | `listEndpoints` | 129, 686 | 40 (C) |
| `/node/endpoints/{ep}` | `getEndpointDetail` | 149, 703 | 40 (C), 33 |
| `/node/endpoints/{ep}/events` | `listEndpointEvents` | 181, 722 | 40 (C), 33 |
| `/node/surface` | `getNodeSurface` | 204, 742 | 40 (C) |
| `/node/sessions` | `listSessions` | 223, 758 | 35 (B) |
| `/node/topology` | `getNodeTopology` | 235, 774 | 35 (B) |
| `/node/mdns` | `getMdns` | 257, 791 | 40 (E) |
| `/node/compatibility` | `getCompatibility` | 285, 821 | 35 (B) |
| `/node/events` | `listDiagnosticEvents` | 297, 837 | 35 (D) |
| `/node/exposure/schema` | `getExposureSchema` | 309, 853 | 35 (F) |
| `/node/exposure` | `listExposure` | 322, 870 | 35 (F) |

| Method, path | operationId | Line | Decision (class) |
| --- | --- | --- | --- |
| GET `/operation/log` | `listLogRecords` | 1050 | 20 |
| GET `/operation/log/stream` | `logStream` | 1071 | 20 |
| GET `/controller/nodes` | `listControllerNodes` | 1104 | 35 (H) |
| GET `/controller/nodes/{id}` | `getControllerNode` | 1117 | 35 (H) |
| GET `/controller/nodes/{id}/endpoints/{ep}` | `getNodeEndpoint` | 1145 | 35 (H) |
| GET `/controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}` | `getNodeCluster` | 1175 | 33 (I-b) |
| GET `/controller/nodes/{id}/events` | `listNodeEvents` | 1199 | 33 (I-b) |
| POST `/controller/nodes/{id}/attributes` | `readNodeAttributes` | 1354 | 19 |
| GET `/controller/nodes/{id}/acl` | `getNodeAcl` | 1424 | 19 |
| GET `/controller/nodes/{id}/bindings` | `getNodeBindings` | 1464 | 19 |
| GET `/controller/nodes/{id}/icd` | `getNodeIcd` | 1500 | 35 (H) |
| GET `/controller/nodes/{id}/fabrics` | `listNodeFabrics` | 1560 | 19 |
| GET `/controller/credentials` | `listCredentials` | 1652 | 33 (L) |
| GET `/controller/updates` | `listUpdates` | 1718 | 35 (H) |
| GET `/controller/topology` | `getControllerTopology` | 1762 | 35 (J) |
| GET `/controller/diagnostics` | `getControllerDiagnostics` | 1800 | 35 (J) |
| GET `/controller/mdns` | `getControllerMdns` | 1822 | 35 (J) |
| GET `/controller/events` | `listControllerDiagnosticEvents` | 1842 | 35 (K) |
| GET `/controller/stream` | `controllerStream` | 1860 | 41 |

- `getControllerTopology`: the operation mark replaces the parameter
  mark on `refresh` (`openapi.yaml:1789`); drop it and the entry of
  `controllerAdminQuery` (`fabricui/handlers_controller.go:83-85`).
- Go tables that follow: `nodeRoutes` (`fabricui/handler.go:304-329`),
  `operationAdmin` (`:340`), `controllerAdmin`
  (`handlers_controller.go:56-79`); their contract tests already exist
  (`handlers_controller_test.go:195`, `handlers_nodes_test.go:167`).
- W: the client sends its admin headers on **every** request (section 6);
  no hand-kept admin list (UR-P3-ctr-07).

### Stay open

| Method, path | operationId | Line | What a non-admin gets |
| --- | --- | --- | --- |
| GET `/node/status`, `/nodes/{node}/status` | `getNodeStatus`, `…ForNode` | 41, 574 | the status without `config.store`; every `warnings[]` entry with `code`, `message: ""`, no `since` (decision 34). `counters` unchanged |
| GET `/nodes` | `listNodes` | 479 | unchanged (class G) |
| GET `/operation/info` | `getOperationInfo` | 1006 | unchanged (class G) |
| GET `/operation/log/level` | `getLogLevel` | 1018 | unchanged — **OPEN-2** |
| GET `/controller/status` | `getControllerStatus` | 1089 | reduced — **OPEN-1** |
| GET `/node/stream`, `/nodes/{node}/stream` | `stream`, `streamForNode` | 433, 993 | filtered, section 2 |

## 2. Node stream allow-list (decision 41)

| Class | Event types | Non-admin |
| --- | --- | --- |
| G node list | `nodes_changed` | yes |
| M window | `window_opened`, `window_closed` (**OPEN-3**) | yes |
| B | `fabric_added`, `fabric_removed`, `fabric_updated`, `commissioning_complete`, `session_opened`, `session_closed`, `subscriptions_changed` | no |
| C | `cluster_event`, `attribute_changed`, `surface_changed`, `endpoints_changed` | no |
| D | `health_sample`, `diagnostic_event` | no |
| F | `exposure_changed` | no |
| host | anything from `Options.Events`, whatever its type | no |

- Class A (status) has no event type of its own; the non-admin page
  reloads the status on the two window events and on every (re)open.
- **Where it lives:** one table in `fabricui/events.go`, beside the
  `Event*` constants (`events.go:17-54`): `nodeStreamEvents
  map[string]bool` (type → a non-admin receives it), every node stream
  type present, an absent type admin-only.
- **Contract side:** `x-non-admin-events: [nodes_changed, window_opened,
  window_closed]` on `stream` and `streamForNode`. Guard test
  `TestNodeStreamAllowListMatchesTheContract`: the table's keys = the
  `Event*` constants plus `EventNodesChanged` (`registry.go:86`); its
  true keys = the extension; a host event of an allowed type does not
  reach a non-admin.
- **Host events by source, not by type** (UR-P3-api-08): the handler
  always owns its hub; a host's `*Events` is subscribed as `extra`
  (today it *is* the hub, `fabricui/handler.go:201-211`) and `extra` is
  dropped for a non-admin. Changelog line: a host reading its own
  `*Events` no longer sees the handler's events. Check both pinned hosts
  first.
- The admin test is evaluated once, when the stream opens, by the same
  function as the routes (decision 45; `isAdmin`, `handler.go:624-630`).

## 3. Statuses

Shared `components/responses` (today only `Problem`,
`openapi.yaml:1921-1926`):

| Name | Status | Headers | Used for |
| --- | --- | --- | --- |
| `Problem` (kept) | 400, 403, 404, 409, 413, 415, 500, 501, 502, 503 | – | every problem without a header of its own; existing references stay |
| `MethodNotAllowed` (new) | 405 | `Allow` (required) | path level, see below |
| `StreamUnavailable` (new) | 503 | `Retry-After` (integer seconds; present with `too_many_streams`) | the four stream operations |

| Status | Problem code | When | Declared on |
| --- | --- | --- | --- |
| 400 | `bad_request` | strict body: unknown member, member name in another case (UR-P3-ctr-06), data after the value, wrong type; factory reset and erase no longer discard the decode error; invalid parameter; label over 32 bytes; invalid passcode (section 4) | every operation with a body or a parameter; new on `getControllerTopology` (`refresh=maybe`, `openapi.yaml:1791-1798`) |
| 403 | `forbidden` | not an admin | every `x-admin` operation (the 43 above gain it) |
| 403 | **H-a** | held row UR-P0-auth-20: see held/ | every operation |
| 403 | **H-b** | held row UR-P0-web-21: see held/ | every `x-admin` operation |
| 404 | `not_found` | unknown fabric index | new on `removeFabric` (`openapi.yaml:65-81`); `removeFabricForNode` and `removeNodeFabric` list 404 already (the controller route answers it once go-fabric exports the error; not in this wave) |
| 404 | `not_found` | unknown path under the prefix | not per operation; `info.description` |
| 405 | `method_not_allowed` (new) | a served path with another method; `Allow` = the path's contract methods, plus `HEAD` where `GET` is | not per operation (**OPEN-5**); `info.description` + guard |
| 409 | `conflict` | the controller's own state refuses (UR-P3-adpt-91: `ErrICDOperation`, and the two siblings `ErrDisabled`, `ErrFabricOperationInProgress`) | new on the 14 operations under `/controller/nodes/{id}` that list 502 and no 409: `decommissionNode`, `interviewNode`, `openNodeWindow`, `closeNodeWindow`, `readNodeAttributes`, `writeNodeAttribute`, `invokeNodeCommand`, `getNodeAcl`, `setNodeAcl`, `getNodeBindings`, `setNodeBindings`, `unregisterNodeIcd`, `listNodeFabrics`, `removeNodeFabric` |
| 413 | `payload_too_large` | body over 1 MiB (`docs/contract.md:410`) | all 26 operations with a `requestBody` (none lists it today, UR-P3-ctr-01) |
| 415 | `unsupported_media_type` (new) | held row UR-P0-web-21: see held/ | the same 26 |
| 500 | `internal` | store error | new on `getCompatibility`, `…ForNode` (UR-P1-err-36); encode error on `getNodeSurface`, `…ForNode` (`docs/contract.md:28` says 500 already) |
| 503 | `too_many_streams` (new) + `Retry-After: 5` | the handler's open streams are at `Options.MaxStreams` | `stream`, `streamForNode`, `logStream`, `controllerStream` via `StreamUnavailable` (`logStream`'s existing 503 `service_unready` moves under the same component) |

- `Problem.code` enum (`openapi.yaml:1939`) gains `method_not_allowed`,
  `too_many_streams`, `unsupported_media_type`, **H-a**, **H-b**; each
  gets a `### problem-<code>` section in `docs/contract.md`.
  `problem-conflict` (`docs/contract.md:398-399`) gains the controller's
  client-state case.
- Request bodies: every schema only a request reaches gets
  `additionalProperties: false` (the guard allows it there,
  `docs/contract.md:461-464`); `info.description` states the strict
  rule for all bodies, nested shared objects included.
- Go guard (decision 47), extending unit A's line scan of the contract:
  every operation with a body declares 400, 413, 415; every `x-admin`
  operation 403; every operation 403 (H-a); every stream 503; no handler
  answers a status its operation does not declare; `Allow` of a 405
  equals the path's methods.

## 4. Fields and twins

New fields are optional in the schema (decision 46) and always sent
by a 0.10.0 server unless "absent when" says otherwise.

| Schema (line) | Field | Type | Notes |
| --- | --- | --- | --- |
| `Status` (1952) | `fabric_count` | integer | **leaves `required`** (`:1954`); absent when the fabric read failed (UR-P1-err-07) |
| `Status` | `window_remaining_seconds` | integer, `minimum: 0` | seconds until the open window's timer closes it, on the node's clock when the answer was written; absent while no window is open (UR-P2-ui-06) |
| `Pairing` (2042) | `window_remaining_seconds` | integer, `minimum: 0` | as on `Status` |
| `Pairing` | `window_match` | boolean | true: the open window admits the codes of this answer; false: a window is open on other credentials (opened elsewhere); absent while no window is open (decision 28, UR-P2-ui-05) |
| `NodeConfig` (2322) | `store` | unchanged | already optional (`:2343`); not sent to a non-admin |
| `Warning` (2006) | `message` | unchanged, stays required | `""` for a non-admin; description says so |
| `EventRecord` (2990) | `number_str` | string, `pattern: "^(0\|[1-9][0-9]{0,19})$"` | the event number in decimal, exact; `number` gets `deprecated: true` (UR-P1-ctr-113, -177) |
| parameter `EventSinceStr` (new) | `since_str`, query | string, same pattern | beside `EventSince` on `listEndpointEvents`, `…ForNode`, `listNodeEvents` (`openapi.yaml:193`, `728`, `1211`); when present it is used and `since` ignored |
| `Session` (2119), `SessionEvent` (2535) | `peer_node_id`, `local_node_id` | integer + `format: uint64`, `deprecated: true` | the `_hex` twins exist; the fix is the unsigned Go type on the controller side (UR-P2-ctr-64). No new twin |
| `Fabric` (2021) | `fabric_id`, `node_id` | + `deprecated: true` | twins exist |
| parameter `NodeID` (1899) | – | + `pattern: "^[0-9A-Fa-f]{1,16}$"` | UR-P1-web-115; accepted break |
| parameter `EndpointNumber` (1905) | – | `maximum: 65534` | UR-P1-adpt-35; accepted break. Body fields named `endpoint` keep 65535 |
| `AclEntry` (3110) | `privilege`, `auth_mode` | enums + `unknown` | response only from now on (UR-P1-ctr-32) |
| `AclEntryWrite`, `AclWriteRequest` (new) | as `AclEntry` / `AclEntryList` | closed; enums without `unknown`, `auth_mode` keeps `pase` | request body of `setNodeAcl` (`openapi.yaml:1450`). The add form (W) offers `case` and `group` only (decision 52) |
| `EndpointBindings` (3140) | `status` | integer | the node's status for this endpoint's read, absent on success, `bindings` then `[]` (as `AttributeValue.status`, `:3064`; UR-P1-bound-37) |
| `EndpointBindingsWrite` (new) | `endpoint`, `bindings` | closed | request body of `setNodeBindings` (`openapi.yaml:1486`) |
| `Applied` (2507) | `reassembled` | boolean | false: stored, the rebuild failed, not in effect (UR-P2-ui-35) |
| `ControllerStatus` (2692) | `admin` | boolean | as `Status.admin` (`:1976`): present in host mode, absent in token mode |
| `ControllerNodeSummary` (2761) | `auto_subscribe` | boolean | the setting, not the live subscription (UR-P1-ctr-168) |
| `ControllerEvent` (3316) | `commissioning_id` | string | on every `commissioning_progress` (UR-P1-api-18) |
| `CommissionRequest` (3199) | `commissioning_id` | string, `pattern: "^[A-Za-z0-9_-]{1,64}$"`, optional | the client's id for the events; absent: the server draws one — **OPEN-4** |
| `CommissionRequest` | `passcode` | unchanged (1..99999998) | description + handler 400 for 11111111, 22222222, … 88888888, 12345678, 87654321; the page refuses the same list (UR-P1-ui-178) |
| `FabricLabelRequest` (3266) | `label` | `maxLength: 32` unchanged | JSON Schema counts characters, so the bound stays as the necessary one and the description says "at most 32 bytes of UTF-8"; the handler answers 400 on `len(label) > 32` (UR-P1-adpt-40) |

Port only, no contract: a presence flag for the discovered device's long
discriminator (UR-P1-adpt-65); one device type naming function
(UR-P1-ui-38).

## 5. `Options` and constructors

| Name | Type | Default | Godoc one-liner |
| --- | --- | --- | --- |
| `MaxStreams` | `int` | 0 = `DefaultMaxStreams` (64); negative is an error | "MaxStreams caps the event streams this handler serves at once, every kind and node together; one more answers 503 with Retry-After." |
| `Done` | `<-chan struct{}` | nil | "Done, when closed, ends every open stream and removes its observers; a host closes it from its server's shutdown hook." |
| **H-c**, **H-e** | see held/ | see held/ | held row UR-P0-auth-20 (decisions 37, 61): two fields |

- `func NewHandler(o Options) (http.Handler, error)` — new; `Handler`
  calls it and panics on the error (stays, documented).
- `Mount(mux, o) error` already returns the error
  (`fabricui/options.go:325-331`): no new variant — **OPEN-6**.
- `Handler` / `NewHandler` refuse `Prefix` other than `""` or
  `DefaultPrefix` (decision 49); `Mount` keeps it.
- Godocs to correct: `Handler` ("fallback to its index … An invalid
  prefix … panics", `options.go:333-335`: unknown path is 404, any
  invalid option panics); `FrameAncestors` names "New"
  (`options.go:274-275`: it is `Handler`, `NewHandler`, `Mount`) and
  gains the public-suffix warning (decision 50); `Events`
  (`options.go:248-253`): admin-only on the stream.
- New exported constants: `DefaultMaxStreams`, **H-d** (the request
  header's name, beside `TokenHeader`, `handler.go:27`).

## 6. npm package (decision 45)

- New in `./client` (`ui/src/lib/api/stream.ts`, exported through
  `api/index.ts`):
  `fetchEventSource(init?: { headers?: () => Record<string, string>; fetch?: Fetch }): EventSourceFactory`
  — an `EventSourceLike` over `fetch` (`credentials: "same-origin"`):
  `open` on a 200 `text/event-stream`, `error` and `readyState` 2 on any
  other answer or a read error, `close()` aborts; parses comments and
  multi-line `data:`.
- `FabricClient.openStream: EventSourceFactory` — `fetchEventSource`
  bound to the client's headers and fetch. The three stores pass it
  (`store.svelte.ts:508`, `controller-store.svelte.ts:246`,
  `context.ts:162` pass `undefined` today).
- `EventSourceFactory` (`stream.ts:64`), `EventSourceLike` and the
  `subscribe*` signatures (`stream.ts:202-207`, `:246-251`, `:268-273`)
  stay; their default stays the browser's `EventSource`, which carries no
  header: such a stream is a non-admin's (node) or 403 (controller, log).
  Changelog and `ui/README.md` say so.
- Headers on every request: the token header only to a same-origin base
  (UR-P1-web-197); **H-d** always.

## 7. Version and accepted breaks

- **Contract 0.10.0**, module and npm release v0.10.0. A patch adds no
  problem code (`docs/contract.md:451-452`) and this wave adds five; the
  project is before 1.0 and the guard's `info.x-accepted-breaks`
  (`docs/contract.md:456-458`) is the way a minor carries a deliberate
  break (decisions 13, 15: one bump, one release after the pull
  request).
- `info.x-accepted-breaks`, three entries, exactly:
  1. `schema Status: field fabric_count no longer required`
  2. `parameter NodeID: narrower (pattern)`
  3. `parameter EndpointNumber: narrower (maximum)`
- Not reported by the guard, docs only: the `x-admin` marks, the strict
  controller bodies, new enum values on response fields, `format` and
  `deprecated`. Nothing else may be reported: no property joins a
  `required` list that existed, no type changes, no status goes.
- `docs/contract.md`, after line 512:

  > Contract 0.10.0 (review phase 3) is over 0.9.0, which
  > `contract/previous/` holds as release 0.9.2 shipped it. Three breaks
  > are made on purpose and listed in `info.x-accepted-breaks`:
  > `Status.fabric_count` is optional (absent when the fabric read
  > failed), the `NodeID` path parameter takes 1 to 16 hex digits only,
  > and the `EndpointNumber` path parameter ends at 65534. One more the
  > guard does not see: 43 reading operations and the controller and log
  > streams are `x-admin` now. A client without admin rights keeps
  > `GET /node/status`, `/nodes/{node}/status`, `/nodes`,
  > `/operation/info`, `/operation/log/level`, `/controller/status` and
  > the node stream, which then carries `nodes_changed`, `window_opened`
  > and `window_closed` only; its status has no `config.store` and every
  > warning's `message` is empty. An admin request carries the admin
  > headers on every route, the streams included, so a client reads them
  > through `fetch`. Every request body is strict and JSON. New: the
  > statuses 405, 413, 415 and the stream's 503 with `Retry-After`; the
  > problem codes `method_not_allowed`, `too_many_streams`,
  > `unsupported_media_type`, [H-a], [H-b]; the fields of section 4.

  G fills the bracketed names and the sentences for the held rows from
  held/ when the rows are released. `info.description`
  (`openapi.yaml:9-21`) gets the same facts: accepted breaks, strict
  bodies, 404/405, the admin paragraph.

## 8. Open points

All six taken by the owner as recommended: decisions 57–60 (`decisions.md`, 2026-11-19).

| # | Point | Recommendation |
| --- | --- | --- |
| OPEN-1 | `GET /controller/status` is in class H, which decision 35 gates, yet decision 44 puts `admin` on it and it serves the token (`docs/contract.md:371-372`, `openapi.yaml:2706`) | Stays open. A non-admin gets `enabled`, `capabilities`, `admin`/`csrf_token` as they are and every other required member zero (`""`, `0`) — decision 46's `Warning.message` pattern, no guard break |
| OPEN-2 | `GET /operation/log/level` is class G (open), but decision 35's sentence names only status, node list and operation info | Stays open, unchanged |
| OPEN-3 | Decision 41 names `window_opened`; `window_closed` is in no class and carries `{}` (`fabricui/events.go:316`) | A non-admin receives it; without it the status's window flag goes stale |
| OPEN-4 | Decision 44 asks for a commissioning id on the events; the POST answers only when the commissioning ends (`openapi.yaml:1599-1601`), so the page cannot learn a server-drawn id in time | The optional request field of section 4 |
| OPEN-5 | Decision 47 says "every status per operation"; 405 and the unknown-path 404 are answered by no operation | Path level: the `MethodNotAllowed` component, `info.description`, and the Go guard on `Allow` |
| OPEN-6 | Decision 48 asks for an error-returning `Mount` variant; `Mount` returns an error today (`options.go:325`) | Nothing to add; only `NewHandler` is new |
