# Threat model, phase 0

Code at baseline `25f6e7b4`. Pins: go-fabric `v0.22.2`, go-fabric-server
`d61a545`, go-ha2fabric `ec7deea` (`git archive` exports; both hosts pin
go-fabric-ui `v0.9.2`, `go.mod:7` each), oapi-codegen runtime `v1.8.0`
(`go.mod:7`, read in the module cache). Paths without a prefix are this
repository; `gf:` is go-fabric, `srv:` go-fabric-server, `ha2f:`
go-ha2fabric.

**M** = measured (the line was read, the statement is what the line does).
**I** = inferred (follows from measured lines plus browser, proxy or host
behaviour that is not in a pin; to be shown by a test in the unit named).
Nothing below was executed; no exploit was run.

**Held back: 2 rows** (decisions 5 and 14), IDs 20 and 21, in
`held/threat-model-held.md` (git-excluded). The per-threat table marks where
they sit (T3, T7) without their detail.

## 1. Assets

| # | Asset | Where it lives | Who is meant to reach it |
| --- | --- | --- | --- |
| A1 | The admin capability: every write, the pairing secret, the controller's admin routes | `fabricui/handler.go:477-493`, `fabricui/handlers_controller.go:51-81`, `:120-142` | Whom `Options.Admin` admits; in token mode whoever holds the token |
| A2 | The per-process token (token mode only) | drawn `handler.go:265-269` (16 random bytes, hex), copied to each node's handler `:402`; in the page: module state `ui/src/lib/api/client.ts:291`, `:336` | See assumption S1 |
| A3 | The node's pairing secret: passcode, QR and manual code, of the configured payload and of a window the handler opened | `fabricui/handlers_pairing.go:68-88` (GET), `:134-200` (window) | Admin only (`handler.go:289-291`) |
| A4 | A controller node's window codes; the commission request's code and passcode | `handlers_controller.go:648-668`, `:1259-1308` | Admin only (`:57`, `:66`) |
| A5 | Wi-Fi passphrase, Thread dataset | write-only request `handlers_controller.go:1412-1447`; never in a response struct (list `:1390-1407` carries kind, id, name, default) | Admin writes; nobody reads back |
| A6 | Node and fabric state: fabrics, ACL, bindings, sessions, attribute values, the exposure selection, the registry | the write routes of `handler.go:286-318` and `handlers_controller.go:51-74` | Admin to change; reading is per class, section 3.2 |
| A7 | Host data: the daemon's log, warnings, counters, node configuration (listen address, store file, discriminator), host item names | `fabricui/operations.go:227-284`, `fabricui/handlers_node.go:143-167` | Undecided per class (decision 8) |
| A8 | The origin the page runs in. Under an Ingress it is the host's origin (`docs/embedding.md:349-350`), so script injection in a page is script in the host's frontend | `ui/src/shell/index.html`, the bundle | The page's own code only |
| A9 | Availability of the daemon and of the nodes it controls | streams `fabricui/stream.go:31-78`; live reads `fabricui/gofabric/node.go:448-455` | — |

## 2. Trust boundaries and entry points

### 2.1 The HTTP handler

| Entry | Gate | Lines | State |
| --- | --- | --- | --- |
| Node family, 24 routes, each under `/node` and `/nodes/{node}` | per row: plain or `adminOnly` | table `handler.go:286-309`, loop `:331-345`, wrappers `:328-329` | M |
| Registry (3), `PUT /operation/log/level` | `adminOnlyBecause` | `handler.go:314-318`, `:347-350`, `:353` | M |
| `GET /nodes`, `/operation/info`, `/operation/log/level`, `/operation/log`, `/operation/log/stream` | none | `handler.go:346`, `:351-355` | M |
| Controller family, 40 generated routes | `controllerGate`: 22 routes by table, one query parameter (`refresh`), then the 1 MiB bound | `handlers_controller.go:51-81`, `:120-142` | M |
| Anything else under the prefix | 404 problem, the path reflected as JSON text | `handler.go:361-363` | M |
| `/` (`Handler` only) | none; the bundle with index fallback, no method check | `fabricui/options.go:331`, `fabricui/spa.go:34-56` | M |

The gate runs before a body is read on every admin route (`handler.go:479-484`,
`handlers_controller.go:125-129` before `:137-140`). `isAdmin` has no input
besides `Options.Admin(r)` or the token header (`handler.go:487-493`) (M).

### 2.2 The two admin modes

| Mode | Selected by | Check | What the status says |
| --- | --- | --- | --- |
| Host mode | `Options.Admin != nil` | `h.admin(r)`, nothing else (`handler.go:488-490`) | `admin: <bool>` per request (`handlers_node.go:139-142`); no token (`handler.go:265` not taken; `csrf_token` is `omitempty`, `handlers_node.go:49`) |
| Token mode | `Options.Admin == nil` (`options.go:218-225`) | header `X-Fabric-UI-Token` non-empty and equal in constant time (`handler.go:491-492`) | `csrf_token` (`handlers_node.go:138`) |

**Assumption S1 (decision 7).** In token mode whoever reaches the listener
is an admin: the token is issued without any check by **three** ungated
routes — `GET /node/status`, `GET /nodes/{node}/status` (both
`handlers_node.go:138` through `handler.go:286`, `:331-345`; the per-node
handler carries the same token, `handler.go:402`) and
`GET /controller/status` (`handlers_controller.go:319-320`). The inventory's
count of three is confirmed (M). Restricting who reaches the listener is the
host's job.

**Do the docs say it plainly? No.** `docs/embedding.md:356-359` says only
"Token mode is for a daemon that serves the page directly";
`docs/contract.md:370-374` describes the token as what stops "a page of
another origin in the same browser"; `options.go:218-225` describes the
mechanism. None states that any client that reaches the listener obtains the
token and with it every admin route (finding 22).

**Assumption S2.** In host mode the handler adds nothing to the host's
answer: no token, no origin check. What `Options.Admin` reads must therefore
itself be safe against a request the user's browser sends without the user
meaning to. The docs do not say so (`docs/embedding.md:356-359`; grep for
"cross", "forge", "SameSite" in `docs/embedding.md`, `ui/README.md`: only
`docs/contract.md:373`, which is about token mode).

### 2.3 The four SSE streams

| Stream | Route | Gate | Payload source |
| --- | --- | --- | --- |
| Node, default | `GET /node/stream` | none (`handler.go:309`) | handler events, `node.Observe`, the host's `Options.Events` (`fabricui/events.go:263-308`) |
| Node, keyed | `GET /nodes/{node}/stream` | none (same row, `handler.go:337-344`) | as above, of that node |
| Controller | `GET /controller/stream` | none (not in `handlers_controller.go:51-74`) | commissioning progress `handlers_controller.go:1290-1314`, trace `:1866-1870` |
| Log | `GET /operation/log/stream` | none (`handler.go:355`) | every record the host's ring takes (`operations.go:274-284`) |

One loop serves all four (`stream.go:31-78`): 64-event channel per client,
drop when full (`:33-39`), `no-store` (`:46`), heartbeat 15 s
(`handler.go:250`). There is no cap on the number of open streams
(`stream.go:31-42`, `events.go:208-220`; M). Frame: `event: <type>` and
`data: <json.Marshal>` (`stream.go:65-70`); the data cannot hold a line
break, the type is written as given.

### 2.4 The SPA

| Property | Fact | Lines | State |
| --- | --- | --- | --- |
| Response headers | `X-Content-Type-Options: nosniff` on bundle responses only; API responses carry `Content-Type` and, for a problem, `X-Problem-Code` | `spa.go:37`; `fabricui/problem.go:72-73`, `:118` | M |
| Caching | index `no-store`; `assets/` one year immutable; other files one hour; JSON answers: no `Cache-Control` at all | `spa.go:71`, `:61-66`; `problem.go:117-121`, `:71-76` | M |
| Framing | no `X-Frame-Options`, no `frame-ancestors` | grep over `fabricui`, none | M |
| CSP | none as a header, none as `<meta>`; the index has one inline script, so a CSP needs its hash | `spa.go:68-79`; `ui/src/shell/index.html:9-89`; `fabricui/dist/index.html` (one `<script>` without `src`) | M |
| CORS | no `Access-Control-*` header is ever sent | grep, none | M |
| Path handling | `fs.FS` name from the URL path; an invalid name or a directory falls back to the index | `spa.go:38-52` | M |
| HTML sinks | two `{@html}`, both the QR SVG built from the module matrix; the payload text is not interpolated | `ui/src/lib/pages/node/Pair.svelte:150`, `ui/src/lib/pages/controller/node/Window.svelte:66`, `ui/src/lib/qr.ts:46-57`; grep for `innerHTML`, `outerHTML`, `insertAdjacentHTML`, `srcdoc`, `document.write`, `eval(`, `new Function`: none | M |
| Navigation sinks | dynamic `href` values are a constant `#/…` route plus an id (`ui/src/lib/context.ts:92-97`, routes `:60-65`); `Graph` follows with `location.assign` (`ui/src/lib/kit/Graph.svelte:29`, `:145`) | | M for the prefix; every call site: 1b S-web |
| Token handling | kept in a closure, sent only on calls marked admin, `credentials: "same-origin"` | `client.ts:291-302` | M |

### 2.5 The embedding host boundary

| Boundary | Fact | Lines | State |
| --- | --- | --- | --- |
| `Options.Admin` | the only authority in host mode; called per request, also by the status route | `handler.go:488-490`, `handlers_node.go:139-142` | M |
| `Mount` beside a host's routes | registers the API and a catch-all `<prefix>/` on the host's mux, no SPA; a duplicate pattern panics inside `register` and is recovered into an error after earlier patterns were registered | `options.go:307-320`, `handler.go:326-364` | M |
| Same origin as the host | the token (token mode) and every response are readable by any script of that origin; the handler cannot tell them apart | consequence of 2.2 | I |
| Theme bridge | when framed, reads `window.parent.document`'s computed style for a constant list of 31 names and copies the values to its own root with `style.setProperty`; a cross-origin parent throws and is ignored | `ui/src/lib/theme/ha-bridge.ts:11-43`, `:106-150`; `isEmbedded` `:54-68` | M |
| Prepaint script | the same before first paint, inline | `ui/src/shell/index.html:17-88` (list `:21-53`, copy `:59-63`, `try` `:54`, `:84-86`) | M |
| What crosses | only a same-origin parent can be read, and a same-origin parent could script the page anyway; the values land in custom properties, not in markup. `isEmbedded` is also true for any framed page whose path contains `/api/hassio_ingress/` (`ha-bridge.ts:61`), which changes the skin only | | M / I |
| No message channel | no `postMessage`, no `message` listener, no cookie, no `window.name` | grep over `ui/src/lib`, `ui/src/shell`: none | M |

### 2.6 Untrusted data rendered in pages

Sources: a node's own strings (labels, names, attribute and event values,
`DebugText`), other fabrics' labels, mDNS records, the host's warnings,
counters, exposure items and log records, problem details. All reach the
pages as JSON (`encoding/json`, `problem.go:120`) or SSE data
(`stream.go:65`) and are rendered by Svelte text interpolation; the two
`{@html}` sites take none of them (2.4). The per-site list is the
inventory's section 3.3, to be re-read in 1b S-web; phase 0 re-read the sink
classes, not every site.

## 3. Attacker roles

### 3.1 R1 — unauthenticated client on the listener's network

| Mode | Reads | Drives |
| --- | --- | --- |
| Token mode | everything: the status hands out the token (S1) | everything |
| Host mode, `Admin` answers false | the ungated classes of 3.2 | what 3.2 lists under "drives" |
| Host mode, `Admin` by source address | as the address decides; a host that filters the listener by the same address (both pinned add-on configurations, section 4) serves this role nothing | — |

### 3.2 R2 — non-admin viewer (host mode, `Admin` answers false)

Per route class. "Discloses" is what the response type carries; rows marked
M were read at the line, the others are the contract's and the inventory's
shape and are measured field by field in X3.

| Class | Routes (ungated) | Discloses | State |
| --- | --- | --- | --- |
| Node status and configuration | `GET …/status` | enabled, listen address, boot reason, attestation error text, window state, and from the host: label, listen, **discriminator**, vendor and product, **store file name**, warnings, counters (`handlers_node.go:143-176`) | M |
| Node fabrics, sessions | `…/fabrics`, `…/sessions` | who commissioned the node, peers | X3 |
| Node model and values | `…/endpoints`, `…/endpoints/{ep}`, `…/events`, `…/surface`, `…/compatibility`, `…/topology`, `…/mdns` | device model, attribute and event values, mDNS services and addresses | X3 |
| Exposure | `…/exposure/schema`, `…/exposure` | the host's item list (names of what the host could expose) | X3 |
| Node list | `GET /nodes` | keys, labels (`handler.go:452-466`) | M |
| Operation | `/operation/info`, `/operation/log/level`, `/operation/log` | version, uptime, level, **the daemon's log records with all attributes** (`operations.go:227-241`, `:44-53`) | M |
| Controller inventory | `/controller/status`, `/nodes`, `/nodes/{id}`, `…/endpoints/…`, `…/clusters/…`, `…/events`, `/updates`, `/topology` (cached), `/diagnostics`, `/mdns`, `/events` | fabric and node ids, addresses, node attribute values, sessions and peers, the trace — which at `v0.22.2` records a pairing's **discriminator** (`gf:controller/diagnostics.go:245-246`, passed through `handlers_controller.go:1855-1857`) | M for the trace; rest X3 |
| Controller live reads | `GET …/acl`, `…/bindings`, `…/fabrics`, `POST …/attributes` | the node's ACL, bindings, fabrics of every ecosystem (read unfiltered, `handlers_controller.go:1184`), any attribute by path (`:702-724`) | M |
| Credentials | `GET /controller/credentials` | kind, id, **name** (SSID, Thread network name), default (`handlers_controller.go:1399-1405`) | M |
| Streams | the four of 2.3 | fabric, session, attribute, cluster-event and window transitions (window: discriminator and duration, `events.go:72-75`, `:335`), commissioning phase and step, the trace, every log record | M for window and log; rest X3 |
| Never disclosed | — | passcode, QR and manual code (admin routes only: `handler.go:289-291`, `handlers_controller.go:57`); passphrase and dataset (no response field, `handlers_controller.go:1399-1405`); a failed commissioning's error text on the stream (kind only, `:1320-1336`) | M |

Drives: opens streams without a cap (2.3); makes the controller send reads
to a node through the live-read class (finding 26). In the UI the admin
controls are hidden by `status.admin` or the token's presence
(`ui/src/shell/App.svelte:107`), which is presentation, not a control.

### 3.3 R3 — malicious page or origin

| Vector | Token mode | Host mode |
| --- | --- | --- |
| Forged write | needs the header token; a custom header forces a preflight the handler never answers (no CORS, 2.4) — M for the lines, I for the browser | held back (decision 18) |
| Cross-origin read of status, token, streams | blocked by the same-origin policy only: no CORS header is sent, so `fetch` and `EventSource` of another origin get no data (I) | same |
| (held back, decision 18) | T3 | T3 |
| Framing, clickjacking | the page can be framed by any origin (2.4) and runs with admin rights inside the frame (finding 23) | under an Ingress the frame is intended; the parent is same-origin |
| Hostile parent or sibling frame under an Ingress | n/a | a same-origin parent or sibling is already inside the trust boundary (A8); a cross-origin parent can read nothing (2.5). Other content of the host's origin sharing `localStorage` (`ui/src/lib/kit/DataTable.svelte:62-63`): table preferences only — 1b |
| Crafted link into the page | hash route parsing (`ui/src/shell/route.ts:22-45`): ids are pattern-checked (`:28`, `:42`); `decodeURIComponent` at `:37` is unguarded (finding 30) | same |

### 3.4 R4 — hostile node or host data

| Path | Control | State |
| --- | --- | --- |
| JSON bodies | `encoding/json` on the way out (`problem.go:120`, `:75`); 64-bit ids as hex strings | M |
| SSE framing | data is `json.Marshal` output, no raw line break (`stream.go:65`, `:70`); the event type is not checked (finding 33) | M |
| Render sinks | text interpolation everywhere but the two QR sites, which do not take data text (2.4) | M for the class; every site: 1b |
| Link targets from data | constant `#/` prefix plus id (`context.ts:92-97`) | M |
| A node's strings in errors | a commissionee's `DebugText` (up to 128 characters, `gf:controller/commissioner/wire.go:139`) is concatenated into the error (`gf:controller/commissioner/network.go:271`, `:296`, `:354`, `:375`), which becomes the 502 detail and a log attribute (`handlers_controller.go:191-193`) and is shown as text (`client.ts:307-308`, `ui/src/shell/Toaster.svelte:16`) | M |
| A node's JSON into Go | `json.Unmarshal` of node-supplied lists without a size bound of its own (`handlers_controller.go:880`, `:1193`) | M for the absence; the bound of the read below it: 1a G-go |
| Size and shape in the page | `JSON.parse` of responses and events without a shape check (`client.ts:304-305`, `ui/src/lib/api/stream.ts:77`) | M; effect: 1b |
| User regex over data | `new RegExp(term)` on the search text (`ui/src/lib/kit/utils.ts:17-24`); self-inflicted only | M |

## 4. The hosts as configured (decision 12)

The inventory's section 5 was re-read line by line against the exports;
every row holds except one line range (section 8).

| | go-fabric-server `d61a545` | go-ha2fabric `ec7deea` |
| --- | --- | --- |
| Entry | `fabricui.Handler(o)` at the root of its own server (`srv:internal/daemon/web.go:58`, `:128`) | `fabricui.Handler` behind `filter` (`ha2f:internal/web/web.go:68-70`) |
| `Options.Admin` | always set or deliberately nil: `--web-admin-source` → peer address equals it (`web.go:97-102`, `fromAddr` `:110-116`); `--web-admin` → nil, token mode (`:103-104`); else a func that returns false (`:106`) | add-on (`SUPERVISOR_TOKEN` set, `ha2f:internal/config/config.go:212`) → peer address equals `172.30.32.2` (`ha2f:cmd/ha2fabric/main.go:187-188`, `web.go:126-133`, `config.go:63`); `WEB_ADMIN` → nil (`main.go:189-190`); else `ReadOnly` (`web.go:137`) |
| Auth in front | with `--web-admin-source`: every other peer gets 403 for every route (`web.go:59-72`); otherwise nothing | `WEB_ALLOW` peer filter for every route, off when empty (`web.go:76-89`); otherwise nothing |
| Listens | `--web`, default off (`srv:internal/daemon/run.go:174`); add-on `0.0.0.0:8099`, source `172.30.32.2`, `host_network`, Ingress (`srv:script/run.sh:56-57`, `srv:addon/config.yaml:33`, `:44-45`) | default `127.0.0.1:8099` (`config.go:48`); add-on `0.0.0.0:8486` with `WEB_ALLOW=172.30.32.2` (`ha2f:script/run.sh:70-71`, `ha2f:addon/config.yaml:37`, `:43-44`); container image `0.0.0.0:8099` without `WEB_ALLOW` (`ha2f:Dockerfile:31`) |
| Transport | plain HTTP, `ReadHeaderTimeout` only (`web.go:124-130`) | plain HTTP, read-header, read and idle timeouts, no write timeout (`web.go:96-100`) |
| Ports wired | `Controller`, `Operations`, `Events`, `Warnings`, `Node`, `NodeConfig`, `Erase`, `Registry`, `ExposureReason` (`web.go:36-52`) | `Node`, `Exposure`, `Reassemble`, `Events`, `Warnings`, `Counters`, `Erase`, `Operations`, `Registry`, `NodeConfig` (`main.go:112-116`); no controller family |

Roles as the hosts produce them (M): neither host has an authenticated
non-admin viewer. Under the add-on every request is the Ingress proxy's and
is an admin; stand-alone the choice is read-only for everyone who reaches
the listener, or token mode for everyone who reaches it.

**For the hosts' reviews** (not go-fabric-ui findings):

| # | Host | Observation | Lines |
| --- | --- | --- | --- |
| HR-1 | both | held back until the fix of T7 (decision 18); goes to the hosts' reviews with it | — |
| HR-2 | go-ha2fabric | The container image listens on every interface with no client filter; with `WEB_ADMIN` unset that is the read classes of 3.2 for the whole network, the log included; with it set, token mode for the whole network | `ha2f:Dockerfile:31`, `main.go:189-192`, `web.go:77-79` |
| HR-3 | go-fabric-server | `--web-admin` with `--web` on a routable address is token mode without a filter; the flag's help text says "the page's token mode", not who becomes an admin | `srv:run.go:174-175`, `web.go:59-61`, `:103-104` |
| HR-4 | both | The log ring the Operation page serves is the daemon's whole logger; nothing in go-fabric-ui gates or filters it, so each host owns what its log attributes carry | `srv:internal/daemon/operation.go:26-29`; `ha2f:main.go:115`; here `operations.go:227-284` |
| HR-5 | both | Plain HTTP: stand-alone the token and the pairing codes cross the network in clear | `srv:web.go:128-130`; `ha2f:web.go:96-108` |
| HR-6 | go-fabric-server | No read or idle timeout on the pages' server | `srv:web.go:128` |
| HR-7 | go-ha2fabric | `web.New` drops an `Allow` entry that does not parse; if none parses the filter is off. The daemon's own configuration refuses such a value first (`config.go:308-311`), so this bites only another caller of the package | `ha2f:web.go:59-63`, `:77-79` |

## 5. Per threat

| # | Threat | Entry point | Existing control | Gap | Read further in |
| --- | --- | --- | --- | --- | --- |
| T1 | A client without rights calls an admin route | every write, `GET …/pairing`, the 22 controller routes | `adminOnly` and `controllerGate` before any body (`handler.go:477-485`, `handlers_controller.go:125-129`); table held against the contract's `x-admin` by test (`handlers_controller.go:47-50`) | none seen; path canonicalisation, method matching and the `{node}` double registration under the gate | 1a S-go; phase 0 fuzz over the mux |
| T2 | A client on the network becomes admin in token mode | the three status routes | none, by decision 7 (S1) | the docs do not say it (finding 22) | X3; phase 3 docs |
| T3 | held back until its fix (decision 18) | — | — | **1 row held back** | X4 |
| T4 | Token guessed or compared unsafely | `X-Fabric-UI-Token` | 128 random bits (`handler.go:266-268`), constant-time compare, empty refused (`handler.go:491-492`); no rotation for the process's life | none | 1a S-go |
| T5 | Token or pairing secret left in a cache or a log | JSON answers | index and streams `no-store`; the handlers log neither (`handlers_pairing.go:197-199` logs the discriminator only; `options.go:254-255`) | JSON answers carry no `Cache-Control` (finding 25) | 1a S-go; X1 |
| T6 | A non-admin reads more than the owner intends | the ungated classes of 3.2 | passcode, codes, passphrase and dataset are on admin routes or in no response | the log (finding 27) and live reads (finding 26); the rest is decision 8, per class | X3 |
| T7 | held back until its fix (decision 18) | — | — | **1 row held back** | X4; the hosts' reviews (HR-1) |
| T8 | Forged write from another origin in token mode | every admin route | the header token: a custom header cannot be sent cross-origin without a preflight, and none is answered (2.4) | see T3 | X4 |
| T9 | Cross-origin read of status, token or a stream | the status routes, the four streams | no CORS header anywhere (2.4) | see T3 | X4 |
| T10 | Clickjacking the standalone page | `/` | none | no framing control (finding 23) | X4, with a Playwright test (decision 10) |
| T11 | Script injection through node, host or log data | every page | text interpolation; two `{@html}` sites that take no data text; constant link prefixes | no CSP as a second line (finding 24); every sink | 1b S-web; X4 |
| T12 | Forged stream frames | `event:` line | data is JSON (`stream.go:65`) | type unchecked (finding 33) | 1a S-go |
| T13 | Hostile parent frame under an Ingress | theme bridge, prepaint script | only a same-origin parent is readable; values go to custom properties from a constant name list; `try/catch` | none seen | 1b S-web |
| T14 | Secret in an error text | 4xx and 502 details, the log | 5xx carries no detail (`problem.go:103-107`); the adapter replaces the pairing-code parser's error (`fabricui/gofabric/gofabric.go:292-295`); the stream carries a kind only (`handlers_controller.go:1320-1336`) | resolved in section 6: no code, passphrase or dataset; an SSID and a device's text do travel (finding 31) | X1 |
| T15 | Resource exhaustion by a client without rights | streams; `GET /operation/log` (limit ≤ 2000, `operations.go:153`, `:262-265`); live reads (≤ 32 paths, `handlers_controller.go:45`) | bodies 1 MiB (`problem.go:33`, `:126`; `handlers_controller.go:137-139`); 64-event buffer with drop | no cap on open streams (finding 29); live reads ungated (finding 26); the hosts set no write timeout | X2; X3 |
| T16 | Body smuggles a field the handler ignores | the 13 generated bodies, factory reset, erase | hand-written routes refuse unknown fields (`problem.go:128`) | the generated ones do not (finding 28) | 1a G-go; phase 3 contract |
| T17 | `Mount` leaves a host's mux half-registered | `Mount` | the panic is recovered (`options.go:312-316`) | the patterns registered before it stay (finding 32) | phase 3 api |
| T18 | A crafted link breaks or misroutes the page | the hash route | id patterns (`route.ts:28`, `:42`) | finding 30 | 1b S-web |
| T19 | Path traversal out of the bundle | `/` | `fs.FS` refuses invalid names; error → index (`spa.go:43-47`) | none seen | 1a S-go; phase 0 fuzz |
| T20 | A node's oversized or malformed answer | adapter and handler `json.Unmarshal`, `JSON.parse` | 502 on a list that does not decode (`handlers_controller.go:881-883`) | no size bound of its own | 1a G-go; 1b |
| T21 | The committed bundle differs from its source | `fabricui/dist/` | `make dist-check` | — | phase 3 supply |

## 6. The inventory's four unresolved items

| Item | Answer | Lines | State |
| --- | --- | --- | --- |
| `runtime.BindJSONBody` | `json.NewDecoder(r.Body).Decode(v)`; the validated-body branch needs a middleware that is not installed (`handlers_controller.go:100-103` passes no middleware). So: unknown fields ignored, data after the first JSON value ignored, no content-type check. The size bound is the gate's (`handlers_controller.go:137-139`), 1 MiB, and holds | runtime `v1.8.0` `bindjson.go:36-45`; `internal/api/api.gen.go:5836` | M |
| go-fabric error text in problem details | **Pairing code:** not echoed — the adapter parses first and answers a fixed text (`gofabric.go:292-295`); go-fabric's decoder does print an invalid passcode (`gf:secure/setup/decode.go:193`, `:428`) but that text is discarded there. **Dataset:** not echoed — the Thread decoder's errors carry sizes, offsets and type names (`gf:thread/tlv.go:33-44`, `gf:thread/dataset.go:182`, `:261`, `gf:thread/view.go:118-121`), wrapped at `gofabric.go:89-92`. **Passphrase:** no error in `gf:controller/commissioner/*.go` formats one (grep over `Errorf` and `newError`). **What does travel:** the Wi-Fi SSID and the commissionee's `DebugText` (`gf:controller/commissioner/network.go:271`, `:296`; Thread: status and `DebugText`, `:354`, `:375`), as the 502 detail of an admin route and as a log attribute (`handlers_controller.go:191-193`) | as cited | M for the sites read; the passcode on the discovered path (`handlers_controller.go:1305`) and every other commissioner error: X1 |
| The server's `exposureReason` | a constant sentence, no data: "the server keeps no host item model; selection is the client's" | `srv:internal/daemon/operation.go:20` (comment `:17-19`) | M |
| `HA_THEME_VARS` vs the inline list | identical, 31 names in the same order (scripted diff of `index.html:22-52` against `ha-bridge.ts:12-42`); a test holds them together (`ui/src/lib/theme/theme.test.ts:123-129`) | as cited | M |

Not resolved here: the sources of the error values logged at
`handlers_pairing.go:79` and `:180` (go-fabric's `SetupPayload` and window
errors) — X1.

## 7. Findings

IDs 20 and 21 are held back. "Contract" says whether the fix changes what a
client or host can observe (decision 13: such a fix goes to the wizard).

| ID | Sev. | Where | Claim | State | Proposed fix | Contract |
| --- | --- | --- | --- | --- | --- | --- |
| UR-P0-doc-22 | medium | `docs/embedding.md:356-359`, `docs/contract.md:370-374`, `fabricui/options.go:218-225` | Token mode's consequence is not stated: the token is served by three ungated routes (`handlers_node.go:138`, `handlers_controller.go:319-320`), so every client that reaches the listener is an admin. The docs describe the token only as a defence against another origin's page | M | One plain paragraph in `docs/embedding.md` and on `Options.Admin`: token mode authenticates nobody; bind to loopback or put a filter in front. Hand the same sentence to the hosts for their flag help (HR-3) | no |
| UR-P0-web-23 | medium | `fabricui/spa.go:37`, `:54`, `:70-71` (all headers the bundle sets) | No framing control. In token mode any origin can frame the standalone page, and the framed page holds admin rights (`client.ts:296`, `:336`) | M (absence); the click-through is I, to be shown in X4 | `Content-Security-Policy: frame-ancestors 'self'` on the index. An Ingress parent is same-origin (`docs/embedding.md:349-350`), so `'self'` keeps it working; a host with another parent needs an option | header only; an option if added |
| UR-P0-web-24 | low | `spa.go:68-79`; `problem.go:117-121` | No CSP on the index, no `Referrer-Policy`, and `nosniff` on bundle responses only, not on API answers. Under an Ingress the page shares the host's origin (A8), which is what a second line of defence would protect | M | A CSP with the inline script's hash (`index.html:9-89`), `object-src 'none'`, `base-uri 'none'`; `nosniff` and `Referrer-Policy: no-referrer` on every response. The camera scan and the inline styles need checking against it in 1b | no |
| UR-P0-leak-25 | low | `problem.go:117-121`, `:71-76`; `fabricui/handlers_surface.go:28`, `fabricui/registry.go:284` | JSON answers carry no `Cache-Control`, among them the pairing secret (`handlers_pairing.go:70-74`, `:84-88`) and the token (`handlers_node.go:138`); the index and the streams do set `no-store` (`spa.go:71`, `stream.go:46`) | M | `Cache-Control: no-store` in `writeJSON`, `writeProblemBody`, the two hand-set sites and, for the generated routes, in `controllerGate` | no |
| UR-P0-auth-26 | medium | `handlers_controller.go:702-724`, `:870-876`, `:991-1000`, `:1179-1185`; absent from `:51-74` | Four ungated controller routes make the controller read from a node on request (`fabricui/gofabric/node.go:448-455`): up to 32 arbitrary attribute paths per call (`:45`), fabric filter off on request (`:721`). The same handler gates `refresh=true` because it "makes the controller talk to its nodes" (`:130-134`) | M | Decision 8 for the class. Proposal: `x-admin` on `POST …/attributes`; serve ACL, bindings and fabrics to a non-admin from the last read, or gate them | yes (new `x-admin` marks) |
| UR-P0-leak-27 | medium | `handler.go:354-355`; `operations.go:227-241`, `:274-284` | The daemon's log is served to everyone, with every attribute (`operations.go:44-53`). The port promises only "no passcode" (`operations.go:25-26`). What this module itself puts there includes port error texts (`problem.go:104-105`) and the node-failed text (`handlers_controller.go:192`), which can hold an SSID and a device-supplied string (section 6) | M | Decision 8 for the class. Proposal: `x-admin` on `GET /operation/log` and the log stream; or state in `docs/embedding.md` that the ring is public to every page reader (HR-4) | yes if gated |
| UR-P0-ctr-28 | low | runtime `bindjson.go:36-45`; `api.gen.go:5836` and 12 more; `handlers_node.go:325-326`, `:376-377` | Body strictness differs: the hand-written routes refuse unknown fields (`problem.go:128`); the 13 generated bodies ignore them and anything after the first value; factory reset and erase discard the decode error. The plan's "Bodies" fact (`README.md`, "Facts the cut rests on") states `DisallowUnknownFields` for the controller gate, which bounds the size only | M | Decide one rule. Either a strict decode before the generated binding, or a line in `docs/contract.md` that the controller routes ignore unknown fields. Correct the plan's bullet | yes if made strict |
| UR-P0-bound-29 | low | `stream.go:31-42`; `events.go:208-220`; `handler.go:309`, `:355` | No cap on open streams: each costs a goroutine, a 64-event channel and a subscription, on four ungated routes | M (absence); exhaustion I | A per-handler cap answered with 503 and `Retry-After`; the hosts' server timeouts are theirs (HR-6) | a new 503 case |
| UR-P0-ui-30 | low | `ui/src/shell/route.ts:37` | `decodeURIComponent` on a hash segment without `try`: a link such as `#/node/%` raises `URIError` where the shell derives its route | M (the throw); what the page then shows is I | Catch and fall back to the default node's status | no |
| UR-P0-leak-31 | note | `handlers_controller.go:185-193`; `gf:controller/commissioner/network.go:271`, `:296`, `:354`, `:375` | Problem details of the commission and credential routes carry go-fabric's error text verbatim. Read at `v0.22.2` it holds no code, passphrase or dataset, but an SSID and the commissionee's `DebugText`; the same text is logged, and the log is ungated (finding 27) | M for the sites read | Keep the detail on the admin answer; log the error kind and step, not the text, or settle it with finding 27. X1 reads the remaining commissioner errors | no |
| UR-P0-api-32 | note | `options.go:307-320`; `handler.go:326-364` | `Mount` recovers a duplicate-pattern panic after part of the routes are on the host's mux; the error does not say the mux was changed | M | Register on a scratch mux first, or document that a failed `Mount` leaves the mux unusable | no |
| UR-P0-web-33 | note | `stream.go:70`; `events.go:222-232` | The SSE event type is written unchecked; a type with a line break from a host's `Options.Events` would forge a frame. Every type this module publishes is a constant | M | Refuse or drop an event whose type is not `[a-z_]+` | no |

Counts, public: 4 medium, 5 low, 3 note. Held: 2 rows.

## 8. Inventory check

Re-read against the code and the host exports: every line this file cites,
and about 60 further inventory rows (sections 1.1, 1.5, 1.6, 3.1, 3.2, 4.1,
4.2, 5). Errors found: 2, both line ranges, no wrong fact.

| Inventory row | Says | Is |
| --- | --- | --- |
| 3.2, inline script | `ui/src/shell/index.html:14-19` | the `<script>` is `:9-89`, the IIFE `:17-88`; `:19` is the `data-skin` line |
| 5.1, `exposureReason` | `internal/daemon/operation.go:17-` | the constant is `:20`; `:17-19` is its comment |

Confirmed where the brief asked: three ungated routes serve the token; four
SSE routes on one loop; 13 response-header call sites with 6 names; both
hosts' `Admin`, filter, listen address and add-on wiring as section 5 has
them.
