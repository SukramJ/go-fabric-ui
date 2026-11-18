# Phase 2 — X3, a client without admin rights

Reviewer X3 (Fable). HEAD read
`1056db2`; go-fabric `v0.22.2` (`gf:`), go-fabric-server `d61a545` (`srv:`),
go-ha2fabric `ec7deea` (`ha2f:`). Paths without a prefix are this repository
at `1056db2`. **M** = read at the line or run; **I** = inferred.

Roles: **R1** an unauthenticated client on the listener's network (token
mode and host mode); **R2** a non-admin viewer (host mode, `Options.Admin`
answers false). In host mode with `Admin` false for everyone R1 = R2.

Tests live in the scratch worktree `…/scratchpad/p2wt/x3` (detached at
`1056db2`), never in the review worktree:

| File | What | Today |
| --- | --- | --- |
| `fabricui/x3_gate_test.go` | the gate sweep (section 3) and the `refresh` variants | pass |
| `fabricui/x3_cached_test.go` | UR-P2-auth-60; the token-serving routes | **fail** / pass |
| `fabricui/x3_e2e_test.go` | what R2 reads off a real go-fabric node during and after a commissioning (dump) | pass (no passcode, QR or manual code in any answer or event) |
| `fabricui/gofabric/x3_dump_test.go` | what a tokenless client reads off a real controller; UR-P2-auth-60 on the real adapter | **fail** |
| `fabricui/gofabric/x3_trace_test.go` | UR-P1-leak-66 measured end to end (note 63) | **fail** |

`x3_trace_test.go` needs three lines in the scratch copy of
`fabricui/gofabric/e2e_test.go`: `wifiDriver.ConnectNetwork` returns
`7, x3ConnectFails, 0` when the package variable `x3ConnectFails` is set.

No finding of this unit is a held candidate.

## 1. Disclosure table (deliverable a)

"Type" is the response type. **S** = a secret; **I** = a sensitive
identifier (names a person's network, device, peer or a host path);
— = neither. Every row is reachable by R2, and by R1 wherever the host puts
no filter in front (section 5). All M; rows marked *run* were also read off
a real node or controller (`x3_e2e_test.go`, `x3_dump_test.go`).

### 1.1 Node family — `/node/…` and `/nodes/{node}/…` (same handlers, `handler.go:438-452`)

| Route | Type, lines | Fields R2 reads | Mark |
| --- | --- | --- | --- |
| `GET …/status` | `statusResponse` `handlers_node.go:27-78`, built `:138-217` | enabled, listening, `listen_addr`, healthy, `health_note`, endpoint, fabric and enabled counts, advertising, window open / duration / `window_kind` / `window_deadline`, exposure, `exposure_reason`, `admin`, `node_registry`, erase, `boot_reason`, `attestation_error` (go-fabric's error text, `:176`), surface summary `:95-104` | I: `listen_addr`; `attestation_error` can name a file path |
| — `config` | `nodeConfig` `:82-93`, `:145-149` | source, source note, label, `listen`, **`discriminator`**, vendor and product id and name, **`store`** — both hosts pass the store *path* (`srv:internal/daemon/web.go:87`, `ha2f:cmd/ha2fabric/main.go:178`) | I: discriminator, store path |
| — `warnings`, `counters` | `Warning` `options.go:296-304`, `Counter` `:307-315`; `handlers_node.go:157`, `:165` | code, message, since; code, value, message — the host's text; go-fabric-server puts `key + ": " + err.Error()` there (`srv:internal/daemon/nodes.go:312`) | I: host error text |
| — `csrf_token` | `:49`, `:139` | token mode only: the admin token (S1, decision 7) | **S** in token mode |
| `GET …/fabrics` | `fabricResponse` `handlers_node.go:223-234`, `:243-268` | fabric index, fabric id, node id (number and hex), vendor id and name, label, compressed fabric id, **root public key** | I: who commissioned the node |
| `GET …/sessions` | `sessionResponse` `handlers_diagnostics.go:23-36`, `occupancyResponse` `:60-65`, `sessionParameters` `:78-89` | session id, fabric index, peer and local node id, PASE flag, subscription count, activity times, idle seconds; occupancy; session parameters | I: peers' node ids |
| `GET …/endpoints` | `endpointResponse` `:196-208`, `structuralEndpoint` `:217-223` | endpoint ids and tree, device type, reachable, name, `source` = `{"key": host's stable key}` (`:286`), cluster ids, names, revisions | I: host item keys and names |
| `GET …/endpoints/{ep}` | `EndpointDetail` (`controller.go:403`), `readEndpointDetail` `node_detail.go:58-110` | the above plus tags, labels, and **every attribute value of every server cluster** (`node_detail.go:94`, `Node.ReadCluster`, no access control, `gf:node/served.go:53-58`). On endpoint 0 (*run*): BasicInformation (serial number, unique id, node label, location), GeneralDiagnostics `NetworkInterfaces` (**MAC, IPv4 and IPv6 addresses**), NetworkCommissioning `Networks` (network id), OperationalCredentials **`Nocs` (the NOC of every fabric in full), `Fabrics`, `TrustedRootCertificates`**, AccessControl (`Acl` reads `[]` — withheld), GroupKeyManagement (`GroupKeyMap`, `GroupTable`), AdministratorCommissioning window status | I: NOCs and root certificates, MAC and addresses, serial number. No key material: a NOC is a certificate |
| `GET …/endpoints/{ep}/events` | `api.EventRecord` `handlers_node.go:438-454`, `:570` | number, priority, endpoint, cluster, event, name, time, **data by field name**. On endpoint 0 (*run*): `AccessControlEntryChanged` with `latestValue` = the whole ACL entry (privilege, auth mode, **subjects**, targets) and the admin's node id | I: ACL entries (note 61) |
| `GET …/surface` | go-fabric's `surface.Report` as JSON `handlers_surface.go:16-30` | level, profiles, extensions, decisions (cluster, element, verdict, policy, reason), withheld, changed | — |
| `GET …/topology` | `topologyGraph` `handlers_topology.go:37-80`, `:88` | node address, endpoints with name, device type and **host key** (`:239`), fabrics (label, vendor, node id, fabric id `:260-266`), sessions (peer node id, PASE, subscriptions) | I: as fabrics, sessions, endpoints |
| `GET …/mdns` | `mdnsResponse` `handlers_diagnostics.go:132-156`, `:161` | advertising; per service: type, instance name, host name, port, addresses, subtypes, TXT; findings | I, but broadcast on the link by design |
| `GET …/compatibility` | `compatibilityResponse` `:310-330`, `:336` | ecosystem, vendor id, fabric index, fabric label; findings | I: fabric labels |
| `GET …/events` (trace) | `TraceEvent` `events.go:114-120`; `handlers_diagnostics.go:379` | at, kind, severity, message, detail map — at `v0.22.2` the node's detail keys are `peer` (a controller's **address**), `session_id`, `max_errors`, `lockout_duration` (`gf:bridge/securechannel.go:396`, `:474`, `:606-608`, `:1169-1171`, `gf:bridge/conn.go:580-582`) | I: peer addresses; PASE lockout state |
| `GET …/exposure/schema` | `Schema` `provider.go:49-96`; `handlers_exposure.go:90` | columns, groups, filters and their options — the host's vocabulary | — |
| `GET …/exposure` | `itemResponse` `handlers_exposure.go:24-35`, `:117` | per host item: key, name, group, fields, lists, verdict (state, device type, clusters, reason), enabled, source — **the host's whole item list**, exposed or not | I: the household's device inventory |
| `GET …/pairing` | — | 403 (*run*) | never |

### 1.2 Node list and operation

| Route | Type, lines | Fields | Mark |
| --- | --- | --- | --- |
| `GET /nodes` | `deviceNodeSummary` `handler.go:556`, `:586` | key, label, kind, enabled, managed, exposure | — |
| `GET /operation/info` | `operationInfoResponse` `operations.go:143-148`, `:180` | version, go-fabric version, uptime, started at | — (version disclosure) |
| `GET /operation/log/level` | `logLevelResponse` `operations.go:122-125`, `:193` | level, levels | — |
| `GET /operation/log`, `/operation/log/stream` | `logRecordResponse` `:131-137`, `:227`, `:274` | decided: UR-P0-leak-27, decision 20 | not re-reported |

### 1.3 Controller family (the routes not in `controllerAdmin`, `handlers_controller.go:55-78`)

| Route | Type, lines | Fields | Mark |
| --- | --- | --- | --- |
| `GET /controller/status` | `GetControllerStatus200JSONResponse` `:304-327` | fabric id, compressed fabric id, fabric index, the controller's node id, vendor, label, node count, capabilities; `csrf_token` in token mode (`:324`) | I; **S** in token mode |
| `GET /controller/nodes` | `ControllerNodeSummary` `:337-354` | node id, vendor, product, node label, connection, subscription, disabled, icd, **addresses**, commissioned at | I |
| `GET /controller/nodes/{id}` | `ControllerNodeDetail` `:363-421` | the summary; record: fabric indexes, addresses, hostname, operational host, discovered and commissioned at, **discriminator** (`:375`), announced vendor, product, device type, device name; properties incl. **serial number** (`:389`), software version; endpoints with cluster ids, names, features | I |
| `GET …/nodes/{id}/endpoints/{ep}` | `api.EndpointDetail` `:432-466`, `:501` | device types, tags, labels, parts, conformance verdict and violations, clusters with attribute ids and names, **no values** (`gofabric/node.go:234-253`) | — |
| `GET …/endpoints/{ep}/clusters/{cluster_hex}` | `api.ClusterDetail` `:469-490`, `:519-536` | **every attribute value the controller's image holds** (`:484`, `gofabric/node.go:371`). *Run* on endpoint 0: `Acl` with subjects, `Nocs` (NOC in full), `Fabrics`, `TrustedRootCertificates`, `NetworkInterfaces` (MAC, addresses), `Networks`, `GroupKeyMap`, BasicInformation (serial number, unique id) | I — finding 60 |
| `GET …/nodes/{id}/events` | `api.EventRecord`, `:542-552`, `gofabric/node.go:420` | the node's event history with data; *run*: `AccessControlEntryChanged` with the entry | I — finding 60 |
| `GET …/nodes/{id}/icd` | `ICDState` (`controller.go:636`), `:1157` | supported, LIT supported, registered, awake, available, operating mode, next and last check-in — no key | — |
| `GET …/acl`, `…/bindings`, `…/fabrics`, `POST …/attributes` | live reads | decided: UR-P0-auth-26, decision 19 | not re-reported |
| `GET /controller/credentials` | `api.Credential` `:1438-1456` | kind, id, **name** (SSID, Thread network name), default; no passphrase, no dataset | I |
| `GET /controller/updates` | `api.PendingUpdate` `:1536-1557` | node id, vendor, product, target version, status, progress, byte counts | — |
| `GET /controller/topology` | `api.TopologyGraph` `:1630-1668` | per node: id, kind, label, state, role, **`ssid`, `bssid`** (`:1641`), node id, RLOC16, details (key and value — *run*: network, bridge, **addresses**); edges with LQI, RSSI | I: Wi-Fi SSID and BSSID, Thread and IP addresses |
| `GET /controller/diagnostics` | `api.Session` `:1841-1862`, `api.ControllerPeer` `:1865-1892` | sessions (ids, node ids, transport, resumed, times); peers (node id, state, **operational address and host, discovered names**, check-in times, load error text) | I |
| `GET /controller/mdns` | `api.MdnsFinding` `:1897-1909` | severity, code, message, service | — |
| `GET /controller/events` (trace) | `api.DiagnosticEvent` `:1914-1932` | at, kind, severity, message, detail — at `v0.22.2`: pairing `code_kind` and the code's **discriminator** (`gf:controller/diagnostics.go:245-246`), a failed step's and a failed pairing's **error text** (`:250-251`, `:262-270`), node ids, session ids, DNS-SD names and targets | I — UR-P1-leak-66 (known); note 63 |
| `GET /controller/discovery` | — | 403 (*run*) | never |

### 1.4 Streams, per event type

One loop for all four (`stream.go:60`); R2 opens each without a gate.

**Node stream** `GET /node/stream`, `/nodes/{node}/stream` (`stream.go:27`; sources `handler.go:638-654`):

| Event | Payload, lines | Fields | Mark |
| --- | --- | --- | --- |
| `fabric_added`, `fabric_updated` | `FabricEvent` `events.go:64-69`, `:338-345` | fabric index, vendor id and name, label | I: label |
| `fabric_removed`, `commissioning_complete` | `FabricEvent` `:270-277` | fabric index | — |
| `window_opened` | `WindowEvent` `:72-75`, `:334-336` | **discriminator**, duration — also of a window the handler opened on fresh credentials | I (known, threat model 3.2) |
| `window_closed` | `WindowEvent{}` | — | — |
| `session_opened`, `session_closed`, `subscriptions_changed` | `SessionEvent` `:172-185` | as `GET …/sessions` | I |
| `attribute_changed` | `AttributeChangedEvent` `:96-100` | endpoint, cluster, attribute ids — **no value** | — |
| `cluster_event` | `api.EventRecord` `:278-280`, `:315-317` | as `GET …/endpoints/{ep}/events`: **event data in full**, of every endpoint; *run*: `AccessControlEntryChanged` with the ACL entry | I — note 61 |
| `surface_changed` | `SurfaceChangedEvent` `:87-89` | endpoint | — |
| `endpoints_changed` | `struct{}` `handlers_node.go:314`, `handlers_exposure.go:221` | — | — |
| `exposure_changed` | `ExposureEvent` `events.go:79-81`, `handlers_exposure.go:211` | the changed items' **keys** | I: host item keys |
| `health_sample` | `HealthSampleEvent` `:106-110` | healthy, note, sampled at | — |
| `diagnostic_event` | `TraceEvent` `:114-120` | as `GET …/events` | I: peer addresses |
| `nodes_changed` | `NodesChangedEvent` `registry.go:88-90`, `:353` | node keys | — |
| anything the host publishes | `Options.Events` (`handler.go:643-645`) | the host's payload, unfiltered | the host's |

**Controller stream** `GET /controller/stream` (`handlers_controller.go:1707-1728`):

| Event | Payload | Fields | Mark |
| --- | --- | --- | --- |
| `node_added`, `node_updated`, `node_removed` | `controllerEventData` `:1735-1748` | node id | — |
| `connection`, `subscription` | same, `:1774-1778` | node id, connection state / active | — |
| `commissioning_progress` | same, `:1336-1362` | phase, step (number, sub, name, phase), BLE decision, failure **kind only** (`:1368-1384`), the new node's id | — |
| `update` | same, `:1787-1790` | node id, kind, target version, **`update_error` = go-fabric's error text** (`gofabric/updates.go:84-86`) | I: error text |
| `topology_changed` | same | — | — |
| `controller_diagnostic_event` | `TraceEvent`, `:1937-1941` | as `GET /controller/events`, **with the pairing error text** | I — UR-P1-leak-66; note 63 |

**Log stream**: `log_record` (`operations.go:274-284`) — decided (decision 20).

### 1.5 Never reaches R2 (M, and *run*)

Passcode, QR and manual code of the configured payload and of a window:
admin routes only; absent from every answer and event of both dumps
(`x3_e2e_test.go`, `x3_dump_test.go` search for them). Passphrase and
dataset: no response field (`handlers_controller.go:1447-1454`). The admin token
in host mode: not drawn (`handler.go:281-285`), `csrf_token` omitted.

## 2. What each role drives (deliverable b)

| Role, mode | Writes | Makes the daemon send | Holds |
| --- | --- | --- | --- |
| R1, token mode | everything: the token is handed out by `GET /node/status`, `/nodes/{node}/status` for every key (default, configured, registry) and `GET /controller/status` — *run*, `TestX3_TokenServedBy`: 4 of 4 paths | everything | everything (decision 7; UR-P0-doc-22 open) |
| R2, and R1 in host mode with `Admin` false | **none** — section 3 | the four live reads (UR-P0-auth-26, decided). Nothing else: `GET /controller/topology` without `refresh` serves the service's snapshot (`gofabric/topology.go:38-40`); `GET /controller/mdns` asks the DNS-SD cache (`gf:controller/diagnostics.go:545-563`); the node routes read the node's own state | streams without a cap (UR-P0-bound-29, known); a live read for as long as the node takes (the request's context, `handlers_controller.go:726` — with decision 19 it moves behind the gate) |

Further, for the record (no finding, see note 65): each `GET …/status`
calls four host ports — `Warnings`, `Counters`, `Node.Fabrics`,
`Exposure.List` (`handlers_node.go:157`, `:165`, `:199`, `:207`); each
`GET …/endpoints/{ep}` reads every cluster of the endpoint
(`node_detail.go:87-103`); a first request for a registry node builds its
handler (`handler.go:445`), bounded by the registry. Lists are bounded
(`operations.go:262-265`, `handlers_node.go:420-432`).

## 3. Gate correctness (deliverable c) — no hole found

`TestX3_GateSweep_HostModeNonAdmin` and
`TestX3_GateSweep_TokenModeWithoutToken` (9 header variants) pass. The
handler is built as the hosts build it (`register` → `mount`,
`handler.go:368`) with a node, a registry node, a selection, erase,
reassemble, the memory controller and operations. For each of the
contract's 46 `x-admin` operations (scanned from `contract/openapi.yaml`;
48 marks = 46 operations, the `refresh` parameter, and the prose line 19):

| Check | Result |
| --- | --- |
| the route with a valid body, node keys `default`, `kitchen` (registry), `nope`, `%64efault`, `default%2F` | 403 every time; 403 before the 404 of an unknown node, so no key oracle (`handler.go:436`, `:444-449`) |
| `HEAD` on the admin `GET`s (`…/pairing`, `/controller/discovery`) | 403 |
| path variants: trailing `/`, `//`, `/.`, `/x/..`, a doubled or dotted segment after the prefix, `%2F`, `%00`, `;x=1`, `%3Bx`, upper case, a percent-encoded first letter of the last segment, a leading `/` or `/x/..` — each with the route's method, `GET`, `HEAD`, `POST` | a 2xx only where the contract has an operation without `x-admin` for that method and path (for example `GET …/exposure` beside `PUT`); otherwise 301, 307, 404 or 405. An encoded segment (`/%63ommission`) matches the pattern and is gated |
| methods on the canonical path: `GET HEAD POST PUT PATCH DELETE OPTIONS TRACE CONNECT post Post` | as above |
| token mode: no header, empty, wrong last byte, a prefix, padded with spaces, a wrong value before the right one, the token in the query, in a cookie, in `Authorization` | 403 |
| host mode, `Admin` false, with the right token header | 403 (`handler.go:621-624`: the token is not consulted) |
| body read on a refused request | never (a counting reader; the only reads are the ungated `POST …/attributes`) |
| state after ~30 000 refused requests: fabrics, selection, node list, log level, controller nodes, credentials, update queue, the cached ACL, window state, reannounce, erase, reassemble and registry counters | unchanged |
| `GET /controller/topology?refresh=…`, 22 spellings (`TestX3_RefreshVariants`) | `true 1 t T TRUE True`, percent-encoded key or value, first of two values: 403. `false&true`, empty, `yes`, `on`, `+true`: 400 from the binder. `Refresh=true`, `refresh[]=true`, a `;` pair: the parameter is not set for gate **and** binder (both read `r.URL.Query()`), the route answers unrefreshed |

Why it holds (M): the node family wraps each pattern (`handler.go:436`,
`:611-619`), so the gate is per registered pattern, both registrations
alike (`:438-452`); the controller gate keys on the pattern the sub-mux
will serve (`handlers_controller.go:126-129`), so it cannot disagree with
the dispatch; a path the mux cleans is answered by its redirect and runs no
handler.

## 4. Findings

### UR-P2-auth-60 — the controller's cached image serves what decision 19 gates

- **Severity:** low. Extends **UR-P0-auth-26** (a wider reach of the same
  data; decision 19).
- **Where:** `fabricui/handlers_controller.go:519-536` (`GetNodeCluster`,
  values at `:484`), `:542-552` (`ListNodeEvents`); adapter
  `fabricui/gofabric/node.go:256`, `:371`, `:420`; neither route is in
  `controllerAdmin` (`handlers_controller.go:55-78`).
- **Claim (M, run):** decision 19 gates `GET …/acl`, `…/bindings`,
  `…/fabrics` and `POST …/attributes` and records "a non-admin no longer
  sees those tabs". The same lists stay readable without admin rights from
  the controller's image: `GET /controller/nodes/{id}/endpoints/0/clusters/0000001F`
  (the ACL with subjects), `…/endpoints/{ep}/clusters/0000001E` (bindings),
  `…/endpoints/0/clusters/0000003E` (fabrics, and each NOC), and the ACL's
  changes through `GET …/nodes/{id}/events`. The gate of decision 19 then
  stops the controller from *sending*, not a non-admin from *reading*.
- **Test** (`fabricui/x3_cached_test.go`; helpers `x3New`, `send` in
  `x3_gate_test.go`):

```go
func TestX3_CachedImageServesWhatDecision19Gates(t *testing.T) {
	e := x3New(t, func(*http.Request) bool { return false })
	base := DefaultPrefix + "/controller/nodes/" + e.nodeID
	for _, c := range []struct{ path, marker, what string }{
		{"/endpoints/0/clusters/0000001F", `"subjects"`, "the ACL (AccessControl 0x001F Acl)"},
		{"/endpoints/1/clusters/0000001E", `"Binding"`, "the bindings (Binding 0x001E)"},
		{"/endpoints/0/clusters/0000003E", `"Fabrics"`, "the fabrics (OperationalCredentials 0x003E)"},
	} {
		rec := e.send(http.MethodGet, base+c.path, "")
		if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), c.marker) {
			t.Errorf("a non-admin reads %s: GET %s -> %d %.300s", c.what, c.path, rec.Code, rec.Body)
		}
	}
}
```

```
x3_cached_test.go:25: a non-admin reads the ACL (AccessControl 0x001F Acl): GET /endpoints/0/clusters/0000001F -> 200 {"accepted":[],"attributes":[{"id":0,"id_hex":"0x0000","value":[{"privilege":5,"authMode":2,"subjects":[112233],"targets":null,"fabricIndex":1}]}],…
x3_cached_test.go:25: a non-admin reads the bindings (Binding 0x001E): GET /endpoints/1/clusters/0000001E -> 200 {"accepted":[],"attributes":[{"id":0,"id_hex":"0x0000","value":[{"node":18,"endpoint":1,"cluster":6,"fabricIndex":1}]}],…
--- FAIL: TestX3_CachedImageServesWhatDecision19Gates
```

  On the real adapter (`fabricui/gofabric/x3_dump_test.go`,
  `TestX3_E2E_ControllerNonAdminReads`, a fabrictest device commissioned
  through the handler, then requests without the token):

```
x3_dump_test.go:85: GET …/endpoints/0/clusters/0000001F without the token serves the node's ACL: {"accepted":[],"attributes":[{"id":0,"id_hex":"0x0000","name":"Acl","value":[{"authMode":2,"fabricIndex":1,"privilege":5,"subjects":[17627619866457343044],"targets":null}]},…
--- FAIL: TestX3_E2E_ControllerNonAdminReads
```

  The same run's `…/clusters/0000003E` holds `Nocs` (the NOC in full),
  `Fabrics`, `TrustedRootCertificates`; `…/nodes/{id}/events` holds
  `AccessControlEntryChanged` with `latestValue.subjects`.
- **Fix (owner, decision 8 class I-b below):** either `x-admin` on
  `GET …/clusters/{cluster_hex}` (the only cached route with values) and on
  `GET …/nodes/{id}/events`, or the handler withholds, for a non-admin, the
  clusters decision 19 is about (0x001F, 0x001E, 0x003E) and their events.
  Phase 3, with decision 19.
- **Contract:** yes (new `x-admin` marks, or a documented omission).

### UR-P2-leak-61 — the node family serves a non-admin what the node gives only an administrator

- **Severity:** note (a decision-8 class with a specific; no defect of the
  gate).
- **Where:** `fabricui/node_detail.go:87-103`; `fabricui/handlers_node.go:570`;
  `fabricui/events.go:278-280`.
- **Claim (M, run — `x3_e2e_test.go` dump):** `GET /node/endpoints/0`
  serves OperationalCredentials `Nocs` with each fabric's NOC, and
  `GET /node/endpoints/0/events` and the stream's `cluster_event` serve
  `AccessControlEntryChanged` with the complete entry
  (`"latestValue":{"authMode":2,"fabricIndex":1,"privilege":5,"subjects":[112233],"targets":null}`).
  Over the wire both need Administer privilege and the reader's own fabric.
  The `Acl` attribute itself reads `[]` on the same route
  (`gf:node/served.go:55-57`: "fabric-sensitive fields withheld"), so the
  event log discloses what the attribute read withholds. Certificates and
  node ids, no key.
- **Fix:** owner question, class C below. If the answer is "not to a
  non-admin": drop events of AccessControl (0x001F) and the values of
  0x001F / 0x003E on endpoint 0 for a non-admin, in the three places above.
  Whether go-fabric's `ReadCluster` should withhold `Nocs.noc` as it
  withholds the ACL is a question for go-fabric (phase 3 `cite`).
- **Contract:** yes if filtered (a value absent for a non-admin).

### UR-P2-doc-62 — the docs do not say what a false `Admin` still serves

- **Severity:** low. Docs finding, no failing test possible (decision 17:
  checked against the code — sections 1 and 3 are the check).
- **Where:** `docs/embedding.md:35`, `:362-365`; `docs/contract.md:365-380`;
  `fabricui/options.go:219-226`.
- **Claim (M):** the three places describe `Admin` as who may *write and
  read the pairing secret*. None states the converse: with `Admin` false —
  or with no `Admin` and no token — every route without `x-admin` and all
  four streams answer **any client that reaches the listener**: the node's
  configuration with the store path and discriminator, fabrics, sessions
  and peers, every attribute value and event, the host's item list, the
  controller's inventory with addresses, SSIDs and cached ACLs, both traces
  and (until decision 20) the log; and until decision 19 such a client makes
  the controller send. Both hosts call that mode "read-only"
  (`srv:internal/daemon/run.go:174`, `ha2f:cmd/ha2fabric/main.go:183-184`)
  and one serves it unfiltered on its `--web` address
  (`srv:internal/daemon/web.go:58-61`, `:106`). Two sentences are also wrong
  as written today: `docs/contract.md:367` ("A writing route (POST, PUT,
  DELETE)") and `options.go:219-220` ("every POST, PUT and DELETE route")
  — `POST /controller/nodes/{id}/attributes` is not gated (decision 19 makes
  them true).
- **State of UR-P0-doc-22 at `1056db2` (not re-reported):** unchanged.
  `docs/embedding.md:364-365` still says only "Token mode is for a daemon
  that serves the page directly"; `docs/contract.md:376-380` still rests the
  mode on "a page of another origin … cannot read the status"; no sentence
  says that whoever reaches the listener is an admin. The contract names two
  token routes (`:371`, `:377`); there are three patterns and one per node
  key (section 2).
- **Fix:** with UR-P0-doc-22's paragraph, one more in `docs/embedding.md`
  ("What a client without admin rights reads") that lists the classes of
  section 6 as decided, and the sentence that `Admin` authorises, it does
  not authenticate or filter: restricting who reaches the listener is the
  host's. Hand both to the hosts' flag help (HR-2, HR-3).
- **Contract:** no.

### UR-P2-leak-63 — decision 20 as worded leaves the pairing error text on the trace

- **Severity:** note. Extends **UR-P1-leak-66** (known, open → phase 3); new
  here: the measurement, and that the decision's wording does not reach it.
- **Where:** `gf:controller/diagnostics.go:250-251`, `:262-270`;
  `fabricui/gofabric/diagnostics.go:105-108`;
  `fabricui/handlers_controller.go:1914-1932`, `:1937-1941`; the comment at
  `:1366-1367` ("no error text that might echo a code travels in an event")
  holds for `commissioning_progress` only.
- **Claim (M, run):** decision 20 gates `GET /operation/log` and the log
  stream. The same text is on `GET /controller/events` and on
  `controller_diagnostic_event`, neither of which the decision names. A
  BLE commissioning whose device refuses the Wi-Fi connect with a
  `DebugText`:

```go
func TestX3_FailedCommissioningTextReachesANonAdmin(t *testing.T) {
	x3ConnectFails = "psk-mismatch-from-device"
	t.Cleanup(func() { x3ConnectFails = "" })
	s := newCtrlSiteWith(t, ctrlSiteOptions{ble: true})
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	status, out := s.call(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q,"wifi_id":"lab"}`, sp.Manual))
	if status < 400 {
		t.Fatalf("the commissioning did not fail: %d %s", status, out)
	}
	resp, err := s.srv.Client().Get(s.srv.URL + "/api/v1/controller/events") // no token
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	for _, text := range []string{"psk-mismatch-from-device", `WiFi network \"lab\"`} {
		if strings.Contains(string(body), text) {
			t.Errorf("GET /controller/events without the token carries %s", text)
		}
	}
}
```

```
x3_trace_test.go:44: GET /controller/events without the token carries psk-mismatch-from-device
x3_trace_test.go:44: GET /controller/events without the token carries WiFi network \"lab\"
x3_trace_test.go:49: trace entry: {"at":"2026-01-01T00:00:00Z","detail":{"err":"commissioner: Commissionee failed to connect to WiFi network \"lab\": psk-mismatch-from-device","step":"NetworkCommissioning.Wifi"},"kind":"pairing","message":"Pairing failed at step NetworkCommissioning.Wifi: commissioner: …","severity":"error"}
x3_trace_test.go:49: trace entry: {…"detail":{"code_kind":"manual","discriminator":"15"},"kind":"pairing","message":"Pairing started.",…}
--- FAIL: TestX3_FailedCommissioningTextReachesANonAdmin
```

  No pairing error of `gf:controller` at `v0.22.2` formats a passcode, a
  code, a passphrase or a dataset (grep over `Errorf`, `newError`,
  `errors.New` in `controller/`, `client/discovery/`): SSID, Thread network
  name (`gf:controller/commissioner/network.go:108`), the device's text.
- **Fix:** phase 3, with decision 20: say whether the two traces
  (`GET /node/events`, `GET /controller/events`, `diagnostic_event`,
  `controller_diagnostic_event`) follow the log behind the gate, or the
  adapter drops `Detail["err"]` and the sentence's tail for kind `pairing`
  (UR-P1-leak-66's second option). Threat model 3.2's "Never disclosed … a
  failed commissioning's error text on the stream (kind only)" is true of
  `commissioning_progress` and false of the stream.
- **Contract:** yes if gated.

### UR-P2-ctr-64 — a node id of 2^63 or more is a negative number in `GET /controller/diagnostics`

- **Severity:** note (outside the unit; seen in the dump).
- **Where:** `fabricui/handlers_controller.go:1845`, `:1847`
  (`int(c.PeerNodeID)`, `int(c.LocalNodeID)` under a `nolint`); generated
  type `internal/api/api.gen.go:1449`, `:1457` (`int`); contract
  `contract/openapi.yaml:2542`, `:2544`.
- **Claim (M, run):** `x3_dump_test.go`:
  `"local_node_id":-819124207252208572,"local_node_id_hex":"F4A1E2DF97CF0444"`.
  The node family's twin is `uint64` and positive
  (`handlers_diagnostics.go:26`, `:28`). go-fabric's operational node ids
  are random 64-bit values, so half of them print negative.
  "Informational; use `…_hex`" — but a negative id is wrong, not imprecise.
- **Fix:** drop the two numeric fields from the controller's `Session`, or
  type them so the generator emits `uint64`. Phase 3 contract unit.
- **Contract:** yes.

### UR-P2-bound-65 — a status request runs four host ports, for anyone

- **Severity:** note. T15; adds to UR-P0-bound-29 (streams), not the same
  site.
- **Where:** `fabricui/handlers_node.go:157`, `:165`, `:199`, `:207`.
- **Claim (M):** `GET …/status` is ungated and polled; each call runs the
  host's `Warnings`, `Counters`, `Node.Fabrics` and `Exposure.List` on the
  request's goroutine, the last only to count enabled items. Nothing bounds
  the rate; what a call costs is the host's (go-ha2fabric's `List` is its
  item model). No send on the network was found behind them at the pins.
- **Fix:** none proposed in this module beyond a sentence in
  `docs/embedding.md` that the four ports are called per status request by
  any client and must be cheap; a count on `ExposureProvider` would be a
  port change.
- **Contract:** no.

## 5. R1 under the hosts as configured (M, `srv:` and `ha2f:` at their pins)

| Host, configuration | R1 is | Lines |
| --- | --- | --- |
| go-fabric-server, `--web` alone | R2 for the whole network: every class of section 1, controller family included; the live reads | `srv:internal/daemon/web.go:58-61`, `:106`; `run.go:174` |
| go-fabric-server, `--web-admin` | admin (token mode) | `web.go:103-104` |
| go-fabric-server, `--web-admin-source` (add-on) | nothing: 403 for every other peer on every route | `web.go:59-72` |
| go-ha2fabric, default | nothing off-host (`127.0.0.1:8099`) | `ha2f:internal/config/config.go:48` |
| go-ha2fabric, container image without `WEB_ALLOW` | R2 for the network, with the selection's item list, the log; no controller family | HR-2; `ha2f:cmd/ha2fabric/main.go:185-193` |
| go-ha2fabric, `WEB_ADMIN` | admin (token mode) | `main.go:189-190` |
| go-ha2fabric, add-on | nothing (`WEB_ALLOW` peer filter) | threat model section 4 |

Neither host has an authenticated non-admin viewer: R2 exists only as "R1
where nobody is admin". HR-2 and HR-3 stand as written.

## 6. Decision 8 — the classes the owner decides

Each: what R2 learns (section 1), and a proposal. "Open" = stays readable.

| # | Class | Routes and events | Sensitive content | Proposal |
| --- | --- | --- | --- | --- |
| A | Node status | `…/status` without `config`, `warnings`, `counters` | listen address, attestation error text | open (the page needs it to say "off") |
| A-b | Node configuration and host findings | `status.config`, `status.warnings`, `status.counters` | **store path**, discriminator, host error text | owner: withhold `config.store` and the warnings' text from a non-admin, or accept and document |
| B | Fabrics, sessions, topology, compatibility | `…/fabrics`, `…/sessions`, `…/topology`, `…/compatibility`; `fabric_*`, `session_*`, `subscriptions_changed` | who commissioned the node, peers' node ids, root public keys | owner |
| C | Node model and values | `…/endpoints`, `…/endpoints/{ep}`, `…/endpoints/{ep}/events`, `…/surface`; `cluster_event`, `attribute_changed`, `surface_changed` | all attribute values; on endpoint 0 NOCs, root certificates, MAC and addresses, serial number; ACL entries through events (note 61) | owner: at least endpoint 0's 0x001F and 0x003E and their events |
| D | Node trace and health | `…/events`; `diagnostic_event`, `health_sample` | controllers' addresses, PASE lockout state | with decision 20 (note 63) |
| E | Node mDNS | `…/mdns` | what the node broadcasts anyway | open |
| F | Exposure | `…/exposure/schema`, `…/exposure`; `exposure_changed`; `endpoints[].source.key`, topology `key` | the host's whole item list, exposed or not | owner — the widest host-data class |
| G | Node list, operation info and level | `/nodes`, `/operation/info`, `/operation/log/level`; `nodes_changed` | version | open |
| H | Controller inventory without values | `/controller/status`, `/nodes`, `/nodes/{id}`, `…/endpoints/{ep}`, `/updates`, `…/icd`; `node_*`, `connection`, `subscription`, `update`, `commissioning_progress` | fabric and node ids, addresses, hostnames, serial numbers, the record's discriminator, `update_error` text | owner |
| I-b | Controller cached values and node events | `…/clusters/{cluster_hex}`, `…/nodes/{id}/events` | ACL, bindings, fabrics, NOCs, network ids, MAC and addresses of every commissioned node | **gate** — finding 60; otherwise decision 19 is half done |
| J | Controller diagnostics and topology | `/controller/diagnostics`, `/topology` (cached), `/mdns`; `topology_changed` | sessions, peers' addresses and hosts, **SSID and BSSID**, Thread data | owner |
| K | Controller trace | `/controller/events`; `controller_diagnostic_event` | pairing error text, the code's discriminator, SSID, device text | with decision 20 (note 63, UR-P1-leak-66) |
| L | Credential names | `GET /controller/credentials` | SSID, Thread network name | owner: gate (the Commission page, its only reader, is an admin page) |
| M | Window transitions | `window_opened` | discriminator and duration of a window, also the handler's own | open (the discriminator is in mDNS while the window is open) |
| — | Host events | whatever `Options.Events` publishes | the host's | document: the stream is public to every page reader |

Decided already, not in the table: the log (decision 20), the four live
reads (decision 19).

## Covered without finding

- **Gate, node family:** the 15 admin rows of `nodeRoutes` and the three
  registry routes under `/node`, `/nodes/{node}` for a configured, a
  registry, an unknown and an encoded key; `PUT /operation/log/level`.
  403 before body, node lookup and handler build. Section 3.
- **Gate, controller family:** the 22 table routes and `refresh`; method,
  HEAD, path, encoding and query variants; pattern lookup against dispatch
  (`handlers_controller.go:126`), redirects, the catch-all 404, CONNECT.
- **Token check:** empty, wrong, prefix, padded, second value, query,
  cookie, `Authorization`; host mode ignores the header.
- **No side effect on refusal:** state snapshot and counters equal after
  the sweep; no body read.
- **Secrets on ungated routes and streams:** passcode, QR and manual code
  absent from every answer and event of a real node during and after a
  commissioning and of a real controller after one; passphrase and dataset
  have no response field; `commissioning_progress` carries the kind only.
- **What R2 drives:** no write; no send beyond the decided live reads —
  cached topology, cached mDNS diagnosis, cached ICD state, cached update
  queue read at the adapter and at go-fabric.
- **Token mode, R1:** the token on all status patterns, every node key
  (decision 7; measured, not re-reported).
- **Hosts:** `Admin`, the filters and the listen addresses of both pins
  re-read for section 5; no difference from threat model section 4.
- **Docs:** `docs/embedding.md` and `docs/contract.md` against sections
  1–3 (finding 62; UR-P0-doc-22's state).
- **Not re-reported:** the held rows; UR-P0-auth-26, UR-P0-leak-27, -31,
  UR-P0-bound-29, UR-P1-leak-09, -66; HR-1 to HR-7.

## Counts

High 0, medium 0, low 2 (UR-P2-auth-60, UR-P2-doc-62), note 4
(UR-P2-leak-61, -63, UR-P2-ctr-64, UR-P2-bound-65). Held candidates: none.
