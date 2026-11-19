# Phase 3 — unit A, contract and api

Reviewer A (Opus), HEAD read `f4b8573`, go-fabric pin `v0.22.2`. Scratch
worktree with the red tests:
(scratch worktree)
(`fabricui/p3a_internal_test.go`, `ui/scripts/p3a-guard.test.mjs`;
`ui/node_modules` is a symlink to the review worktree's, read only). Run:
`go test ./fabricui -run TestP3A_ -count=1 -v` and, in `ui/`,
`node --test scripts/p3a-guard.test.mjs`. No held candidate among the
findings below.

## Findings

### UR-P3-ctr-01 — low — every body route answers 413, and no operation declares it

- **Where:** `fabricui/problem.go:143-157` (node family `decodeJSON` →
  `badBody` → 413 `payload_too_large`), `fabricui/handlers_controller.go:124-126`
  (`controllerGate` bounds the body; the generated bind error goes to
  `badBody`, `:100`); contract: the problem code is in the enum
  (`contract/openapi.yaml:1939`) and documented (`docs/contract.md:410`),
  but none of the 26 operations with a `requestBody` lists `"413"`. Same
  class: `GET /controller/topology?refresh=maybe` answers 400 (generated
  parameter binding, `handlers_controller.go:106-108`) and the operation
  lists 200, 403, 501, 502 only (`openapi.yaml:1786-1795`); `GET
  /node/surface` writes 500 on an encode error
  (`fabricui/handlers_surface.go:25`, `docs/contract.md:28` says 200, 500,
  503) and the contract lists 200, 503.
- **Claim:** the per-operation responses are not the statuses the
  handlers write. A client generated from the contract has no type for a
  413; the TS client survives only because it reads any problem
  generically.
- **M:** test and output below.
- **What decisions 37–44 add to the same gap:** 403 on *every* operation,
  `GET /node/status` and the bundle included (decision 37's Host check;
  `/node/status` lists only 200 today, `openapi.yaml:41-51`); 415 and 403
  (Origin) on admin and body routes (decision 38); 405 with `Allow` on
  every path (decision 43); 503 with `Retry-After` on the four streams
  (decision 42; `GET /node/stream` and `/controller/stream` list only
  200). The contract guard only checks that nothing is removed, so it
  would not notice any of these missing either.
- **Fix:** in the phase-3 bump, add `"413"` to every operation with a
  body and the three single statuses above; for the decided cross-cutting
  statuses, a `components/responses` entry each and a Go guard test in
  the style of the test below (every op with a body declares 413 and
  415, every x-admin op 403, every op 403 for the Host check, every stream
  503). Alternative: a `default: Problem` response on every operation —
  shorter, but it says less, and it changes the generated response types
  in `internal/api`.
- **Contract:** yes (additive responses).

```go
// fabricui/p3a_internal_test.go (scratch)
func TestP3A_TooLargeIsDeclared(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), operations: sampleOperations()})
	big := `{"label":"` + strings.Repeat("x", 1<<20+10) + `"}`
	for _, route := range []string{"PUT /operation/log/level", "PUT /controller/fabric/label"} {
		method, path, _ := strings.Cut(route, " ")
		req := httptest.NewRequest(method, DefaultPrefix+path, strings.NewReader(big))
		req.Header.Set(TokenHeader, testToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("%s: %d, want the measured 413", route, rec.Code)
		}
	}
	st, body := contractStatuses(t) // line scan: path at 2, method at 4, status at 8 spaces, requestBody
	var undeclared []string
	for op := range body {
		if !st[op]["413"] {
			undeclared = append(undeclared, op)
		}
	}
	if len(undeclared) > 0 {
		t.Errorf("%d of %d operations with a request body answer 413 and do not declare it, e.g. %v", len(undeclared), len(body), undeclared[:3])
	}
	if !st["GET /controller/topology"]["400"] {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, DefaultPrefix+"/controller/topology?refresh=maybe", nil))
		t.Errorf("GET /controller/topology?refresh=maybe answers %d; the contract declares no 400", rec.Code)
	}
}
```
```
=== RUN   TestP3A_TooLargeIsDeclared
    p3a_internal_test.go:100: 26 of 26 operations with a request body answer 413 and do not declare it, e.g. [POST /nodes/{node}/pairing/window POST /nodes/{node}/maintenance/factory-reset PUT /operation/log/level]
    p3a_internal_test.go:105: GET /controller/topology?refresh=maybe answers 400; the contract declares no 400
--- FAIL: TestP3A_TooLargeIsDeclared (0.03s)
```

### UR-P3-test-02 — low — the guard compares against 0.8.1 though 0.9.0 is released

- **Where:** `contract/previous/openapi.yaml:4` (`version: 0.8.1`);
  `contract/openapi.yaml:4` (`0.9.0`), released in tags `v0.9.0`,
  `v0.9.1`, `v0.9.2` (`git show v0.9.0:contract/previous/openapi.yaml`
  is 0.8.1 too); `docs/contract.md:454` ("the copy of the last
  release") and `:465-466` ("The release commit that tags a version
  copies the contract to `contract/previous/`"); `ui/package.json:62`.
- **Claim:** what the docs say happens in the release commit has been
  done at the *next* contract bump instead (each "Contract 0.x is
  additive over …" paragraph, `docs/contract.md:467-504`). So everything
  0.9.0 added — three controller routes, `controller_diagnostic_event`,
  `ControllerCapabilities.diagnostics` — can be removed with `make
  contract-check` green, and so can anything released in between. For
  the phase-3 bump: unless the fix wave's first commit copies 0.9.0 into
  `previous/`, the gates and field changes of decisions 19–44 are judged
  against 0.8.1.
- **M:** test and output below.
- **Fix:** first commit of the phase-3 bump: `git show
  v0.9.2:contract/openapi.yaml > contract/previous/openapi.yaml`. Guard:
  `version-check` (or the guard) fails when `previous/`'s `info.version`
  is not the contract version of the latest tag. Correct either the docs
  sentence or the practice.
- **Contract:** no (guard and docs).

```js
// ui/scripts/p3a-guard.test.mjs (scratch), tests 1–2
test("previous is the contract of the last release tag", () => {
  const tag = execSync("git describe --tags --abbrev=0", { encoding: "utf8" }).trim();
  const released = execSync(`git show ${tag}:contract/openapi.yaml`, { encoding: "utf8" });
  const ver = /\n  version: (\S+)/.exec(released)[1];
  assert.equal(prev().info.version, ver, `last tag ${tag} shipped contract ${ver}`);
});
test("removing a released 0.9.0 route is a break", () => {
  const c = cur();
  delete c.paths["/controller/diagnostics"];
  delete c.components.schemas.ControllerDiagnostics;
  assert.notDeepEqual(diffContracts(prev(), c), [], "the guard passes the removal of /controller/diagnostics");
});
```
```
not ok 1 - previous is the contract of the last release tag
  expected: '0.9.0'
  actual: '0.8.1'
not ok 2 - removing a released 0.9.0 route is a break
  error: 'the guard passes the removal of /controller/diagnostics'
```

### UR-P3-test-03 — low — the guard does not catch the breaking changes decisions 34, 43 and 44 make

- **Where:** `ui/scripts/contract-guard.mjs:17-59` (`diffContracts`
  checks removed paths, methods, statuses, schemas, properties and newly
  required fields — nothing else); the decided shapes: decision 44
  "`fabric_count` absent on a failed read" (`Status.required` holds it,
  `openapi.yaml:1954`), decision 34 "each warning's kind only"
  (`Warning.required: [code, message]`, `:2008`), decision 44 "64-bit
  values as strings" and "a hex pattern on `NodeID`", decision 43
  "endpoint `maximum: 65534`" and "`unknown` in the ACL enums".
- **Claim:** a response field that stops being required, a property
  whose type changes, a request parameter or field that gets narrower
  (pattern, maximum) and a new enum value on a response field all break
  a client (the TS types change from `number` to `number | undefined`; a
  strict generated client rejects a missing required field). The guard
  reports none of them, and `docs/contract.md:449-453` says nothing about
  them either. The phase-3 bump makes each of them once.
- **M:** tests 3–5 below. **I:** that the 64-bit change would be done as
  a type change; done as new `*_hex`/`*_str` twins beside the integers
  (the pattern the contract already uses, `openapi.yaml:13-15`) it is
  additive.
- **Fix:** (a) the decided shapes, done in a compatible way: 64-bit values
  as *new* string fields beside the existing integers; `Warning.message`
  kept required and sent as `""` to a non-admin *or* made optional with a
  `docs/contract.md` line naming the break; `fabric_count` made optional
  with a contract.md line. (b) In the guard: report a required field that
  became optional on a schema a response reaches (reuse the reachability
  walk of `closedResponseSchemas`, `:79-102`), a changed `type`, and a new
  `pattern`/`maximum`/`minimum`/`maxLength` on a request parameter or
  request body field. The release accepts each listed break by a line in
  `docs/contract.md`, as the newly-required case already works
  (`:50-55`).
- **Contract:** yes for (a); (b) is a guard change.

```js
// ui/scripts/p3a-guard.test.mjs (scratch), tests 3–5
test("a response field that stops being required is reported", () => {
  const c = cur();
  c.components.schemas.Status.required = c.components.schemas.Status.required.filter((f) => f !== "fabric_count");
  c.components.schemas.Warning.required = ["code"];
  assert.notDeepEqual(diffContracts(cur(), c), []);
});
test("a field whose type changes is reported", () => {
  const c = cur();
  c.components.schemas.Fabric.properties.fabric_id = { type: "string" };
  assert.notDeepEqual(diffContracts(cur(), c), []);
});
test("a tightened request parameter is reported", () => {
  const c = cur();
  c.components.parameters.NodeID.schema.pattern = "^[0-9A-Fa-f]{1,16}$";
  assert.notDeepEqual(diffContracts(cur(), c), []);
});
```
```
not ok 3 - a response field that stops being required is reported
  error: 'Expected "actual" not to be strictly deep-equal to: []'
not ok 4 - a field whose type changes is reported
  error: 'Expected "actual" not to be strictly deep-equal to: []'
not ok 5 - a tightened request parameter is reported
  error: 'Expected "actual" not to be strictly deep-equal to: []'
# tests 5
# pass 0
# fail 5
```

### UR-P3-ctr-04 — note — decided shapes 20, 38 and 41 need a stream that carries a header; the client's streams cannot carry one

- **Where:** `ui/src/lib/api/stream.ts:64` (`EventSourceFactory = (url)
  => EventSourceLike`), `:93` (`new EventSource(u)`), `:202-284`
  (`subscribeStream`, `subscribeControllerStream`, `subscribeLogStream`
  take a URL and that factory); `client.ts:426`, `:435`, `:497` (the three
  `*StreamUrl`); server: `fabricui/handler.go:635-641` (`isAdmin`: in
  token mode the header is the only signal).
- **Claim (I):** decision 20 moves only the log stream to `fetch`.
  Decision 41 gates the controller stream as well, and filters the node
  stream by admin rights. In token mode an `EventSource` request carries
  no token, so (a) the gated controller stream answers 403 to the admin
  page too, and (b) the filtered node stream always treats the page as
  a non-admin: the admin's page loses `fabric_*`, `session_*`,
  `endpoints_changed` and the rest, with nothing to show for it. In host
  mode, decision 38's required header (if the admin decision on streams
  includes it) does the same. The exported `EventSourceFactory` type is
  part of the npm API (`./client`) and has no way to receive a header.
- **Fix:** read all three streams through `fetch` with the client's
  headers. Additive npm shape: `FabricClient.openStream:
  EventSourceFactory` (a fetch-based `EventSourceLike` that sends
  `TOKEN_HEADER` and decision 38's header), used as the default `open`
  by the three `subscribe*` helpers when they get a client. Server
  shape: the stream filter uses the same admin decision as the routes.
  If the owner prefers to keep `EventSource` for the node stream in host
  mode, the filter may accept `Admin(r)` without decision 38's header
  for that GET — no other origin can read an SSE answer without CORS,
  and decision 37's Host check covers rebinding — but that needs to be
  decided, not left to chance. Contract: the stream operations name the
  header and 403/503.
- **Contract:** yes (stream responses and headers); npm API addition.

### UR-P3-api-05 — low — `Handler`'s doc and `Options.FrameAncestors`' doc contradict the code; there is no constructor that returns the error

- **Where:** `fabricui/options.go:333-335` (Handler: "the bundle at "/"
  with a fallback to its index for any other path. An invalid prefix is a
  programming error and panics"); `fabricui/spa.go:26-28`, `:48-64` (any
  other path 404 since decision 36); `fabricui/options.go:274-275`
  ("New and [Mount] refuse anything else") — `fabricui.New` does not
  exist (`go doc ./fabricui`: only `Handler`, `Mount`); `handler.go:225-235`,
  `:168-200` (`newHandler` fails on FrameAncestors, Registry with Nodes,
  Registry without Node, any bad or duplicate node key, all of which
  `Handler` turns into a panic, `options.go:337-340`). `docs/embedding.md:223`,
  `:425` say it correctly ("`Handler` panics").
- **Claim:** a host reading the godoc expects an index for deep links and
  a panic only for a bad prefix. Both pinned hosts use `Handler`
  (decision 23). A host whose `FrameAncestors` or node list comes from
  user configuration (an add-on's options) can only validate it by
  calling `Mount` on a throwaway mux, or by recovering the panic.
  Decisions 37 (`AllowedHosts`), 42 (the stream cap) and the proposed
  `Options.Done` (UR-P2-conc-34) add three more validated fields, so the
  number of ways `Handler` can panic grows.
- **M:** test and output below.
- **Fix:** add `func NewHandler(o Options) (http.Handler, error)`, with
  `Handler` calling it and panicking on the error (or `func (o Options)
  Validate() error`). Correct both godocs: `Handler` 404s an unknown path
  and panics on any invalid option; `FrameAncestors` names `Handler`,
  `NewHandler` and `Mount`.
- **Contract:** no (Go API addition).

```go
func TestP3A_HandlerDocMatchesBehaviour(t *testing.T) {
	for name, o := range map[string]Options{
		"frame ancestor *":   {FrameAncestors: []string{"*"}},
		"registry and nodes": {Registry: newMemoryRegistry(0), Nodes: []NodeEntry{{Key: "a"}}},
		"bad node key":       {Nodes: []NodeEntry{{Key: "status"}}},
	} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s: Handler panics (%v); its doc names only an invalid prefix, and no error-returning constructor exists", name, r)
				}
			}()
			_ = Handler(o)
		}()
	}
	rec := httptest.NewRecorder()
	Handler(Options{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/some/deep/path", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("GET /some/deep/path: %d; Handler's doc promises \"a fallback to its index for any other path\"", rec.Code)
	}
}
```
```
=== RUN   TestP3A_HandlerDocMatchesBehaviour
    p3a_internal_test.go:122: frame ancestor *: Handler panics (fabricui: Options.FrameAncestors: "*" admits any parent; name the parent's origin); its doc names only an invalid prefix, and no error-returning constructor exists
    p3a_internal_test.go:122: registry and nodes: Handler panics (fabricui: Options.Registry and Options.Nodes are both set; the registry lists the nodes beside Options.Node); its doc names only an invalid prefix, and no error-returning constructor exists
    p3a_internal_test.go:122: bad node key: Handler panics (fabricui: node key "status" is a page name of the node family); its doc names only an invalid prefix, and no error-returning constructor exists
    p3a_internal_test.go:131: GET /some/deep/path: 404; Handler's doc promises "a fallback to its index for any other path"
--- FAIL: TestP3A_HandlerDocMatchesBehaviour (0.00s)
```
(The 404 is the decided behaviour; that line of the test shows the doc
is wrong, not the code.)

### UR-P3-ctr-06 — note — body members match case-insensitively; decision 43's "strict" will miss it, and the generated bodies have no place for the check

- **Where:** `fabricui/problem.go:143-148` (`json.Decoder` with
  `DisallowUnknownFields`; encoding/json matches member names
  case-insensitively); `internal/api/api.gen.go:6181` (`runtime.BindJSONBody`),
  runtime `v1.8.0` `bindjson.go:36-45` (plain `json.NewDecoder(...).Decode`,
  unknown members ignored, no option).
- **Claim:** `{"LABEL":"Kitchen"}` creates a node and `{"Level":"debug"}`
  sets the level, although the contract names neither member. Decision
  43's "unknown fields are 400" done with `DisallowUnknownFields` alone
  still accepts these. On the controller family the generated
  `BindJSONBody` cannot be made strict through `oapi-codegen.yaml`, and
  `internal/api` is never hand-edited. So the check must run before the
  generated code: in `controllerGate` (`handlers_controller.go:111-128`),
  decode the body once against the operation's request type with an
  exact-name, no-trailing-data decoder and put the bytes back. Its
  `validatedBody` hook (`bindjson.go:38-41`) is the runtime's way to hand
  the result on.
- **M:** test below. Not a security matter: the gate, the confirm string
  and the types are unaffected.
- **Fix:** one strict decoder for both families: exact member names
  (compare the decoded object's keys against the struct's json tags, or
  decode into `map[string]json.RawMessage` first), unknown members and
  trailing data 400; called in `decodeJSON` and in `controllerGate` for
  routes with a body. Not a red test in this phase's sense: its owner is
  UR-P0-ctr-28/UR-P1-ctr-17's fix, which should include this test.
- **Contract:** with decision 43 (request schemas `additionalProperties:
  false`; allowed for request bodies by the guard, `contract-guard.mjs:65-75`.
  Caution: `AclEntryList` is both the PUT body and the GET answer of
  `/controller/nodes/{id}/acl`, `openapi.yaml:1433`, `:1450`, so closing it
  trips `closedResponseSchemas`. Close the request through a separate
  schema).

```go
func TestP3A_BodyMemberNamesAreExact(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{registry: newMemoryRegistry(0), operations: sampleOperations()})
	for _, c := range []struct{ route, body string }{
		{"POST /nodes", `{"LABEL":"Kitchen"}`},
		{"PUT /operation/log/level", `{"Level":"debug"}`},
	} {
		method, path, _ := strings.Cut(c.route, " ")
		req := httptest.NewRequest(method, DefaultPrefix+path, strings.NewReader(c.body))
		req.Header.Set(TokenHeader, testToken)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s %s: %d, want 400 (no member %q in the contract)", c.route, c.body, rec.Code, strings.Split(c.body, `"`)[1])
		}
	}
}
```
```
=== RUN   TestP3A_BodyMemberNamesAreExact
    p3a_internal_test.go:150: POST /nodes {"LABEL":"Kitchen"}: 201, want 400 (no member "LABEL" in the contract)
    p3a_internal_test.go:150: PUT /operation/log/level {"Level":"debug"}: 200, want 400 (no member "Level" in the contract)
--- FAIL: TestP3A_BodyMemberNamesAreExact (0.00s)
```

### UR-P3-ctr-07 — note — the client's admin flag is a hand-kept list, and decisions 19–40 and 38 change it in about thirty places

- **Where:** `ui/src/lib/api/client.ts:305-316` (the token header only
  when the call passes `admin: true` and a token is held), `:398-497`
  (the flag per call); no test ties it to the contract's `x-admin` (Go
  has `TestNodeAndOperationAdminMatchesTheContract`,
  `fabricui/handlers_nodes_test.go:167-187`, and
  `TestControllerAdminMatchesTheContract`,
  `handlers_controller_test.go:195-210`, for the server side only).
- **Claim:** **M** (read, each x-admin operation against its call at
  `client.ts:398-497`) — consistent today: every x-admin op is called with
  `admin: true`, and the only other flagged call is `refresh=true` on the
  topology. **I** — decisions 19, 20, 33, 35 and 40 turn most GET routes
  of both families into `x-admin`. Each call the fix wave misses works
  in host mode and gives 403 in token mode only (the standalone shell).
  Decision 38's host-mode header is not sent by the client at all: `auth.token`
  is undefined in host mode, so after 38 every admin call of the npm
  client in host mode is refused, unless the client changes in the same
  release. A `Mount` host with its own client has to add the header too
  (changelog, `docs/embedding.md`).
- **Fix:** drop the per-call flag: send `TOKEN_HEADER` (when held) and
  decision 38's fixed header on *every* request of the client. The token
  is the same process-wide secret, already sent same-origin
  (`credentials: "same-origin"`), and a non-admin route ignores it. If the
  flag stays, derive it: `gen:types` emits the list of x-admin operations
  and a vitest checks that the client follows it. Keep the token
  same-origin (UR-P1-web-197 below).
- **Contract:** no (client); the header is decision 38's contract change.

### UR-P3-api-08 — note — decision 41's "host events admin-only" cannot be applied by source while a host's `*Events` becomes the handler's own hub

- **Where:** `fabricui/handler.go:216-227` (`hubOf`: a host's `*Events`
  *is* the handler's hub), `:610-628` (`subscribe` keeps hub, nodesHub,
  extra, node apart); `handlers_exposure.go:211`, `:221`,
  `handlers_node.go:314` (the handler publishes `exposure_changed`,
  `endpoints_changed` into that hub); `options.go:248-253` and
  `docs/embedding.md:41` ("beside the handler's own").
- **Claim:** **M** (read). With `Options.Events` an `*Events`, which is
  what the docs suggest (`NewEvents`), the handler's own events and the
  host's share one hub, so the stream cannot tell a host event from the
  handler's by where it came from. A pure type allow-list (decision 41's
  "one allow-list of event types per class") lets through a host event
  whose type is on the non-admin list. That is harmless today, but it is
  not what "host events are admin-only" says.
- **Fix:** always give the handler a hub of its own and subscribe a
  host's `*Events` as `extra`, so the filter drops `extra` for a
  non-admin. Note that a host that today subscribes to its own `*Events`
  to see the handler's `exposure_changed` stops seeing it. Nothing in the
  docs promises that, but say so in the changelog.
- **Contract:** no.

## The carried `api` rows — judgement and proposed shape

| Row | Judgement | Proposed shape | Contract |
| --- | --- | --- | --- |
| UR-P1-web-20 (note) | Holds. `anyParent` (`spa.go:185-200`) refuses `*.com`-style single-label wildcards only; `https://*.co.uk` and, more to the point, `https://*.github.io` (PSL private section: anyone's page) pass `validFrameAncestors`. A suffix list is a new dependency (`golang.org/x/net/publicsuffix`) for one check | Keep decision 22's `*.domain`; add one sentence to the `FrameAncestors` godoc and `docs/embedding.md:424-427`: a wildcard admits every name below it, so never put one over a public or shared suffix (`*.co.uk`, `*.github.io`). If the owner wants the class closed: refuse *every* wildcard host (only named origins). That is stdlib-only and can break a host that relies on `*.domain` today (decision 22 allowed it), so it is a wizard question | no |
| UR-P1-conc-21 (note) | Holds: `managedHandler` calls `h.registry.List()` under `regMu` (`registry.go:170-175`) on the path a deletion overtook; the port's doc (`registry.go:32-44`) says the methods are called concurrently, nothing about locks. A `List` that waits on anything that waits on `regMu` (another request's `registryEntries`, `:150-156`) deadlocks | Do not call the host under the lock: read `regGen` under the lock; if it moved, unlock, call `List`, re-lock and re-check `regGen` (loop until stable); only then decide cached or uncached. Add to `NodeRegistry`'s doc: "List must not block on a request to this handler" | no |
| UR-P1-api-171 (low) | Holds at `f4b8573`: `pages/index.ts:15-22` exports no controller Diagnostics page; `lib/index.ts:5-7` re-exports `./pages` and `./api` with `export *`, and `api/client.ts:117` exports the type `ControllerDiagnostics`, so the obvious name would clash (TS2308). No other page/api name clashes (checked by script) | `export { default as ControllerDiagnosticsPage } from "./controller/Diagnostics.svelte"`, README list; a vitest that every page the shell imports from `lib/pages/**` (`shell/App.svelte`) is exported from `./pages` — that is the guard that would have caught it | no (npm addition) |
| UR-P1-web-197 (note) | Holds (`client.ts:298-324`). It becomes more relevant with UR-P3-ctr-07's fix (header on every call) | Send `TOKEN_HEADER` only when `new URL(base, globalThis.location?.href).origin === location.origin` (no `location`, as in SSR or tests: send it). Decision 38's fixed host-mode header is not a secret and is always sent. One sentence in `ui/README.md`: a cross-origin `baseUrl` gets no token | no |
| UR-P2-conc-34 (low) | Holds; docs part done (`docs/embedding.md`, `56d1682`). `serveSSE` ends on the request context, `h.done` (registry deletion) or the switch (`stream.go:76-95`) | `Options.Done <-chan struct{}`: closed by the host (in `srv.RegisterOnShutdown`), it ends every open stream (node, nodes/{key}, controller, log) as `h.done` does; nil keeps today's behaviour. Additive, validated by nothing, so it does not add to UR-P3-api-05's panics. Decision 42's cap belongs in the same `serveSSE` place | no (Go API addition) |
| Decision 24 (CSP builder) | Not needed. `contentSecurityPolicy` (`spa.go:129-150`) hashes *fabric-ui's* `index.html` inline script; a `Mount` host's page is its own document with its own inline code, so the builder's output is wrong for it. What such a host needs is to know what the pages need | Keep it unexported. Add to `ui/README.md` what the pages need from a host's CSP: `'self'` scripts and styles, no inline script, styles only through the CSSOM (no `style-src 'unsafe-inline'`), the QR codes as inline SVG through `{@html}` (`pages/node/Pair.svelte:174`, `controller/node/Window.svelte:102`; no `data:`/`blob:` images), the camera scan needs no source; and that `Mount` only validates `FrameAncestors` (`options.go:280-281`) | no |

## Covered without finding

- **Route parity, both ways (M):** every handler pattern is a contract
  operation and every non-controller contract operation has a
  hand-written pattern (`TestP3A_RoutesAndContractAgree`, scratch,
  green). The controller family is generated, so it matches by
  construction. Worth adding as a standing guard: today's
  `TestAPI_EveryRouteAnswersWithTheSecurityHeaders`
  (`headers_test.go:42-72`) also passes for an operation that only
  reaches the catch-all 404.
- **x-admin parity:** node, registry, operation and controller tables
  against the contract (`handlers_nodes_test.go:167-187`,
  `handlers_controller_test.go:195-210`); the `refresh` query mark
  (`controllerAdminQuery`, `:81-83`). The map script's
  "getControllerTopology admin" is the parameter's mark, not the
  operation's. The line scanner (`handlers_controller_test.go:168-193`)
  would count a block-style `x-admin: true` on a parameter as the
  operation's; today every parameter mark is inline.
- **Client against contract (read):** every `client.ts` route exists in
  the contract with the method used; the admin flags are consistent today
  (see UR-P3-ctr-07).
- **Fixtures:** 142, each route and status in the contract (one, 
  `controller-unknown.404`, is a deliberate non-route), each named by a
  Go and a TS test (script). The checks use the name only, not that the
  fixture is run, and `runFixture` (`support_test.go:74-108`) and
  `fixture()` (`ui/src/test/support.ts:20`) are the runners. 24
  operations have no fixture: the `/nodes/{node}/…` mirrors and the
  three streams.
- **`docs/contract.md` route tables against the contract:** admin column
  and statuses agree but for `/node/surface` (folded into UR-P3-ctr-01).
  The parenthesised `(403)`/`(404)` are per-method and correct. 30
  operations are not in the tables by row (the per-node mirrors are
  described in prose, `:82-90`; `enable/disable`, `icd/*`, `updates/*`
  are in combined rows).
- **`Options.Admin` godoc** "every POST, PUT and DELETE route"
  (`options.go:219-226`) is contradicted today by the ungated `POST
  /controller/nodes/{id}/attributes`. Not a new finding: decision 19 /
  UR-P0-auth-26 gate it. The godoc needs rewriting with decisions 19–40
  anyway (it names only "write or read a secret"); UR-P2-doc-62 carries
  it.
- **`NodeSource`, `NodeSwitch`** (every method forwarded in the
  interface's order, `nodeswitch.go:95-205`), **`ExposureProvider`**,
  **`Operations`**, **`EventSource`**, **`NodeRegistry`** error mapping
  (`registry.go:252-262`, `:340-344`): docs match the code. **`gofabric`**:
  `ControllerOf`, `Source.Close`, `ControllerOptions`, `Hooks` docs match
  (`gofabric.go:64-136`). The `Pair` attestation default belongs to the
  held row (decision 25), not reviewed here.
- **`internal/arch`:** facade, adapter and generated-runtime lists
  (`arch_test.go:30-72`) match `CLAUDE.md`; `TestCheckClassifies` holds
  the classes.
- **npm exports** (`ui/package.json` `exports`, `files` without tests):
  every subpath resolves to an index that exists. The shell imports only
  what the package exports, apart from the controller Diagnostics page
  (UR-P1-api-171).
- **Generated server:** `internal/api/oapi-codegen.yaml` limited to the
  `controller` tag. The gate runs before binding and decoding
  (`handlers_controller.go:111-128`), and HEAD on a gated GET pattern is
  gated (the mux's GET pattern matches HEAD).
- **Decisions 37–44 against the contract**, beyond the findings:
  decision 43's 405 and the per-endpoint and fabric 404s are additive;
  `maximum: 65534` and the `NodeID` pattern narrow a request (UR-P3-test-03).
  Decision 44's `admin` on the controller status, `auto_subscribe`, the
  commissioning id, the window's own/other and seconds left are additive
  optional fields. Decision 37's 403 needs the contract's top-level
  description (`openapi.yaml:18-20`), which today speaks only of x-admin.
  The guard's rule "a minor only adds" does not cover gating: the gates of
  19–40 take reads away from a non-admin, so the versioning paragraph
  (`docs/contract.md:449-466`) needs a line saying that is a deliberate
  break in this release.

## Counts

high 0 · medium 0 · low 4 (UR-P3-ctr-01, -test-02, -test-03, -api-05) ·
note 4 (UR-P3-ctr-04, -ctr-06, -ctr-07, -api-08). Contract changes
proposed: ctr-01, test-03 (a), ctr-04, ctr-06 (with decision 43); none for
api-05, ctr-07, api-08 and the carried rows. Held candidates: none.
