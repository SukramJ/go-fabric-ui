# Phase 1a — unit S-go (Go security paths)

HEAD `49cc081`, go-fabric `v0.22.2` (the pinned export). Reviewer: Fable.
**M** = read at the line or run; **I** = inferred. Probes ran in a scratch
copy of the module (`scratchpad/sgo`), never in the repo; their code is at
the end. No finding is a new aspect of UR-P0-auth-20 or UR-P0-web-21.

## Findings

### UR-P1-conc-01 — `HEAD` on a stream route leaks its goroutine and subscriptions for the life of the process

- **Severity:** medium
- **Where:** `fabricui/stream.go:31-86` (loop), reached by `handler.go:313` (`GET …/stream`, both registrations), `handler.go:448` + `operations.go:274-284` (log stream), `handlers_controller.go:1631-1634` (controller stream)
- **Claim:** A `GET` pattern also matches `HEAD`, so `serveSSE` runs for `HEAD`. net/http discards a `HEAD` response's body writes without error, so neither the heartbeat write (`stream.go:61`) nor `Flush` (`:82`) ever fails; the loop can only end through `ctx.Done()`. Once the client has sent one more byte on the connection (a pipelined request, or a keep-alive client reusing it) the server's background read holds that byte and no longer sees the close, so the context is never cancelled. Result: the goroutine, the 64-event channel, the hub subscription and the node's three observers (`events.go:151-167`) stay until the process ends. No rights needed (the streams are ungated), no connection has to be held open — this is not bounded by a cap on *open* streams (UR-P0-bound-29) unless the cap's slot is then lost for good, which turns it into a permanent refusal of streams.
- **State:** M for the node stream (probe 1: `HEAD` + one later byte + close → subscriber still registered 2 s and 100 heartbeats later; `GET` under the same input ends on its write error; `HEAD` followed by a second request from `http.Client` on the same connection: that request times out and `httptest.Server.Close` reports the connection `active`). I for the log and controller streams (same function, same patterns: `api.gen.go:3008ff` registers `GET …`).
- **Fix:** in `serveSSE`'s callers (or one wrapper), answer a method other than `GET` with the headers only and return — `if r.Method != http.MethodGet { … return }` before subscribing; the controller stream needs the method passed into `controllerStream`. Optionally a write deadline per frame (`rc.SetWriteDeadline`) so a stalled reader ends too.
- **Contract change:** no.

### UR-P1-api-02 — `Mount` and `Handler` panic on prefixes `validPrefix` admits

- **Severity:** low
- **Where:** `fabricui/options.go:342-347`; `fabricui/handler.go:343` (`h.routes()` runs before the `recover` of `:349-353`), `:418`
- **Claim:** `validPrefix` refuses only a space and braces. A tab (`"/a\tb"`), an unclean path (`"/a//b"`, `"/a/../b"`, `"/a/."`) pass it and then panic in `mux.Handle` on the API's own mux, outside `mount`'s `recover`: `Mount` panics although its doc says "It fails on invalid options" and returns an error. `"/a%zz"` mounts without error and can never be requested.
- **State:** M (probe 2).
- **Fix:** `validPrefix`: require `path.Clean(p) == p`, no `%`, no white space or control character (or: `url.Parse` round-trip); and build the routes inside the recovered section.
- **Contract change:** no.
- **Red test:** `Mount(http.NewServeMux(), Options{Prefix: "/a//b"})` returns an error and does not panic.

### UR-P1-err-03 — a typed-nil `*Events` as `Options.Events` / `NodeEntry.Events` makes every stream request panic

- **Severity:** low
- **Where:** `fabricui/handler.go:190-195` (`hubOf`), `:603-605`; `fabricui/events.go:208-209`
- **Claim:** `hubOf` tests `ev != nil` and then returns the same typed-nil as `extra`; `h.extra != nil` is true for the interface, and `(*Events)(nil).Subscribe` dereferences nil. Every `GET …/stream` of that node panics (recovered by net/http, the connection is aborted); the half-check suggests the nil case was meant to be handled.
- **State:** M (probe 3: nil pointer dereference).
- **Fix:** in `hubOf`: `if ev, ok := src.(*Events); ok { if ev == nil { return NewEvents(), nil }; return ev, nil }`.
- **Contract change:** no.
- **Red test:** a handler built with `var ev *Events; Options{Events: ev}` serves `GET /node/stream` with 200.

### UR-P1-err-04 — `GET …/pairing` presents the configured codes as the open window's when the handler holds no credentials for that window

- **Severity:** low
- **Where:** `fabricui/handlers_pairing.go:77-88` (fallback), `:97-110` (`activeEphemeral`)
- **Claim:** Whenever a window is open that is not the one this handler opened on fresh credentials, the answer is `source: configured, window_open: true` with the configured passcode, QR and manual code — also when `win.Kind` says the open window is `WindowBasic` (fresh credentials of another handler of the same node, of the host, or of this node's handler before the registry rebuilt it, `registry.go:165-176`) or `WindowEnhanced` (an administrator's). The codes shown cannot commission through that window, and nothing in the body says so. `win.Kind` is read (`:104`) but only to forget, never to answer.
- **State:** M for the code path; I for the page's rendering (1b).
- **Fix:** report the window's kind in the pairing answer (the status already has `window_kind`, `handlers_node.go:191`) and set `window_open` semantics accordingly, or answer `window_open: false`-equivalent for a window whose codes the handler cannot name.
- **Contract change:** yes if a field is added; no if only the page reads `status.window_kind`.
- **Red test:** fake node with `Window{Open: true, Kind: node.WindowEnhanced}` and no ephemeral: the pairing answer must not claim the configured codes for the open window.

### UR-P1-conc-05 — `closeWindow` can drop the credentials of a window opened concurrently

- **Severity:** low
- **Where:** `fabricui/handlers_pairing.go:249-254` (close, then `forgetEphemeral`), `:230-235` (open, then store)
- **Claim:** Close and forget are two steps without a common lock. Request A closes the node's window (`:249`); request B opens a fresh one and stores its credentials (`:230-235`); A then runs `forgetEphemeral` (`:254`) and wipes B's. The window stays open on credentials nobody can read again: `GET …/pairing` falls to UR-P1-err-04's answer. Needs two admin requests within the gap.
- **State:** I (read, not reproduced).
- **Fix:** hold `h.mu` across close-and-forget and across open-and-store, or forget only when the stored discriminator is the one that was open before the close.
- **Contract change:** no.
- **Red test:** a fake node whose `closeWindow` blocks on a channel; open a fresh window while it blocks, release, then `GET /node/pairing` answers `source: window`.

### UR-P1-conc-06 — a registry node's cached handler keeps the `Reassemble` and `Erase` it was first built with

- **Severity:** low
- **Where:** `fabricui/registry.go:125-127` (`sameEntry`), `:165-176` (`managedHandler`); `fabricui/handler.go:487-512` (`sibling`)
- **Claim:** The cache hit compares `Source`, `Exposure` and `Events` and refreshes only `config` (`:170`). `Reassemble` and `Erase` are copied once in `sibling`. A host whose `List` later returns the same source with `Erase` or `Reassemble` set, cleared or bound to new state keeps getting the first closures called, and `Status.erase` (`handlers_node.go:138`) stays what it was. `NodeRegistry.List`'s doc names the three compared fields as what keeps the handler, not that the other two are frozen.
- **State:** M (read).
- **Fix:** store the entry's `reassemble`/`erase` atomically on every lookup, as `config` is; or document them as fixed for the handler's life.
- **Contract change:** no.
- **Red test:** registry lists a node first with `Erase: nil`, then the same `Source` with `Erase` set: `GET /nodes/{key}/status` reports `erase: true`.

### UR-P1-err-07 — the status swallows a failed fabric read and reports `fabric_count: 0`

- **Severity:** low
- **Where:** `fabricui/handlers_node.go:198-200`, `:201-211`
- **Claim:** `fabrics` failing leaves `FabricCount` at 0 (not `omitempty`), without a log line: the status says "no fabric" where the truth is "unknown" (an erased store, `nodeswitch.go:20-23`, answers errors exactly here). The exposure list's error is at least visible as an absent `enabled_count`.
- **State:** M for the code; I for what the page concludes from 0.
- **Fix:** make the count absent on error (pointer, as `enabled_count`) or log at warn; the page then shows "unknown".
- **Contract change:** yes if `fabric_count` becomes optional; no for the log line.
- **Red test:** fake node whose `fabrics` errors: the status body has no `fabric_count` (or a warn record is logged).

### UR-P1-err-08 — factory reset is bound to the request's context: a dropped connection leaves a partial reset

- **Severity:** low
- **Where:** `fabricui/handlers_node.go:333-349`
- **Claim:** The loop removes fabric by fabric with `r.Context()`. A client that goes away mid-loop (tab closed, proxy timeout) cancels the remaining removals; the node stays paired to some controllers, reported only by `fabricui.factory_reset.partial` in the log — the case the handler's own comment calls the one to avoid (`:341-342`).
- **State:** M for the context; I for go-fabric's `RemoveFabric` honouring the cancellation.
- **Fix:** `ctx := context.WithoutCancel(r.Context())` with a timeout of its own for the loop.
- **Contract change:** no.
- **Red test:** fake node that cancels the request's context after the first `removeFabric`: every fabric is still removed.

### UR-P1-leak-09 — the ungated controller stream carries the update error's raw text

- **Severity:** low
- **Where:** `fabricui/handlers_controller.go:1716-1719` (`update_error`); source `fabricui/gofabric/updates.go:82-84` (`e.Err.Error()`); contrast `handlers_controller.go:1318-1336`
- **Claim:** A failed commissioning reaches the stream as a kind only, "so no error text … travels in an event". An `update` event's `update_error` is go-fabric's error text verbatim — for `UpdateCheckFinished` the joined errors of every product whose check failed (`gf:controller/softwareupdate/manager.go:854`, `observe.go:36`, `:111-112`), i.e. network and URL detail — on a stream every non-admin reads (`threat-model.md` 2.3). Not in the phase 0 list of what the streams disclose.
- **State:** M for the path; I for what the joined errors contain (not read below `manager.go:854`).
- **Fix:** carry a kind word (`not_applied`, `cancelled`, `check_failed`) as `progressDetail` does and log the text; or gate with the log stream (decision 20).
- **Contract change:** the field stays; its content narrows (a line in `docs/contract.md`).
- **Red test:** a source emitting `ControllerEvent{Type: EventUpdate, UpdateErr: "Get https://…: dial tcp 10.0.0.1"}`: the frame holds no address.

### UR-P1-err-10 — `discovered.long_discriminator` is cast to `uint16` unchecked

- **Severity:** low
- **Where:** `fabricui/handlers_controller.go:1305-1307`; used at `fabricui/gofabric/gofabric.go:485-492`
- **Claim:** Every other integer of the body is range-checked (`:1272-1280`); this one is narrowed with a `nolint` that asserts 12 bits. `65536+n` or a negative value wraps to another discriminator, and with an empty `instance_name` the adapter commissions the first device it finds under the wrapped value (`gofabric.go:492`).
- **State:** M.
- **Fix:** `badRequest` unless `0 ≤ v ≤ 4095`; consider refusing an empty `instance_name` together with a discriminator.
- **Contract change:** no (the schema's range, enforced).
- **Red test:** `POST /controller/commission {"discovered":{"instance_name":"x","long_discriminator":69632},"passcode":20202021}` answers 400.

### UR-P1-conc-11 — a deleted registry node's handler can be re-cached, and its open streams outlive it

- **Severity:** note
- **Where:** `fabricui/registry.go:180-190` (`nodeHandler`: list, then `managedHandler`), `:133-160`, `:308-323`
- **Claim:** (a) `nodeHandler` lists, a concurrent `DELETE` completes and clears the cache, then `managedHandler` builds and caches a handler on the deleted node's entry; it is served once and stays cached until the next `registryEntries`. (b) `deleteNode` drops the cache entry but nothing ends the streams open on that handler: goroutine, subscription and observers on a stopped node stay until the client leaves.
- **State:** I (a), M (b, absence).
- **Fix:** (a) look up and build under one `regMu` section that re-reads the list; (b) a per-handler `done` channel closed on delete and selected in `serveSSE`.
- **Contract change:** no.
- **Red test:** open `GET /nodes/{key}/stream`, `DELETE /nodes/{key}`: the stream ends.

### UR-P1-err-12 — a request the client abandoned is answered 500 and logged at error

- **Severity:** note
- **Where:** `fabricui/handlers_controller.go:194-195`; `fabricui/gofabric/gofabric.go:619-620` (passes `context.Canceled` through unclassified)
- **Claim:** `context.Canceled` is in no branch of `controllerError`, so each abandoned live read or commissioning writes `Controller failed` at error level with the error text into the log (the ring, UR-P0-leak-27) and a 500 nobody reads.
- **State:** M.
- **Fix:** a branch for `context.Canceled`: no body, debug level.
- **Contract change:** no.
- **Red test:** a source returning `context.Canceled`: no record at error level.

### UR-P1-ctr-13 — a known path with the wrong method answers 404 `not_found`

- **Severity:** note
- **Where:** `fabricui/handler.go:454-456`; `fabricui/handlers_controller.go:112-114`
- **Claim:** The method-less catch-alls match before the mux's own 405, so `POST /node/status`, `OPTIONS /node/pairing`, `POST /controller/discovery` answer 404 "No such route" without `Allow` (probe 4). Harmless for the gate (nothing is served), wrong for a client that probes.
- **State:** M.
- **Fix:** in the catch-all, ask the mux for the path under `GET`/`POST`/`PUT`/`DELETE` and answer 405 with `Allow` when one matches; or state it in `docs/contract.md`.
- **Contract change:** a new status if changed.
- **Red test:** `POST /node/status` answers 405 with `Allow: GET, HEAD`.

### UR-P1-bound-14 — stream: the first `Flush` error is dropped, and an invalid event type is logged raw once per open stream

- **Severity:** note
- **Where:** `fabricui/stream.go:52`, `:69-72`
- **Claim:** (a) Behind a `ResponseWriter` that cannot flush (a host middleware without `Unwrap`) `:52` ignores `ErrNotSupported`; the stream then ends silently at the first event or heartbeat (`:82`), and the page reconnects every 15 s with no log line. (b) `fabricui.stream.invalid_type` writes the host-supplied type unbounded, once per event and per open stream, into a log that is itself streamed.
- **State:** M (read).
- **Fix:** (a) fail the request (500, logged) when the first flush is not supported; (b) drop in `Events.Publish`/at subscribe time once, truncate the type.
- **Contract change:** no.
- **Red test:** `serveSSE` on a writer without `Flush` answers an error instead of 200.

### UR-P1-doc-15 — the stream's comment says the oldest events are dropped; the code drops the newest

- **Severity:** note
- **Where:** `fabricui/stream.go:14-17` against `:34-39`
- **Claim:** `push` discards the incoming event when the channel is full.
- **State:** M.
- **Fix:** correct the comment (or drop the oldest, which keeps the latest state change).
- **Contract change:** no.
- **Red test:** none (docs); checked against the code by the implementer.

### UR-P1-web-16 — `FrameAncestors` admits a wildcard over a top-level label

- **Severity:** note
- **Where:** `fabricui/spa.go:154-155`, `:176-185`
- **Claim:** Decision 22 refuses an entry that admits any parent; `https://*.com` (or `*.net`, `*.local`) passes both checks and admits nearly any.
- **State:** M (regex read; `(?:\*\.)?[A-Za-z0-9-]+` needs one label only).
- **Fix:** require at least two labels after `*.`.
- **Contract change:** no.
- **Red test:** `validFrameAncestors([]string{"https://*.com"})` returns an error.

### UR-P1-ctr-17 — the hand-written bodies accept data after the first JSON value

- **Severity:** note (new aspect of UR-P0-ctr-28, which named the generated bodies)
- **Where:** `fabricui/problem.go:125-130`
- **Claim:** `decodeJSON` decodes one value and never checks for more: `{"level":"debug"}{"level":"nope"} trailing` sets the level and answers 200 (probe 5).
- **State:** M.
- **Fix:** after `Decode`, require `dec.Decode(&struct{}{}) == io.EOF`; settle with UR-P0-ctr-28's one rule.
- **Contract change:** yes if made strict (decision 13).
- **Red test:** the body above answers 400.

### UR-P1-api-18 — concurrent commissionings share one progress stream without a correlation

- **Severity:** note
- **Where:** `fabricui/handlers_controller.go:1289-1314`
- **Claim:** `started` and step events carry neither a node nor a request id; two commissionings at once (two tabs) interleave their steps and each page shows the other's.
- **State:** I (read; whether the adapter serialises commissionings is G-go B's).
- **Fix:** an id per commissioning in the events, or refuse a second with 409.
- **Contract change:** yes.
- **Red test:** two concurrent commissions: each event names its commissioning.

## Read and found nothing

| File / function | What was checked |
| --- | --- |
| `handler.go:289-325`, `:414-458` route tables and registration | every row's gate against the contract's `x-admin` (46 marked operations, script over `contract/openapi.yaml`; held by `handlers_nodes_test.go:166-185`, `handlers_controller_test.go:157-207`); both `{node}` registrations wrap the **root** handler's gate before the node lookup; loop variable capture (Go ≥ 1.22) |
| `handler.go:566-587` `adminOnly`, `isAdmin` | gate before any body; empty token refused; constant-time compare |
| `handler.go:269-273`, `:487-512`, `registry.go:165-176` token life cycle | one token per process, copied to siblings and registry handlers, absent in host mode; `rand.Read` cannot fail at go 1.27 |
| Path and method canonicalisation (probe 4) | `HEAD` on an admin `GET` → 403; trailing slash, case, `%70airing`, `/nodes/default/…` → 404 or 403, never served |
| `handlers_controller.go:98-142` `controllerGate` | `sub.Handler` and `sub.ServeHTTP` resolve alike; `refresh`: gate and binding both `strconv.ParseBool` on `r.URL.Query()`; two values → 400 by the runtime (`bindparam.go:539-541`); `refresh=1` gated; body bound before binding |
| `handlers_controller.go:161-197` `controllerError`, `problem.go:103-108` `serverError` | 500 carries no detail; 502 detail is UR-P0-leak-31 |
| `handlers_node.go:137-213` status | token and `admin` as the threat model has them; what it discloses is the 3.2 list |
| `handlers_node.go:293-390` maintenance | gate, confirmation word, nil node / nil erase order (decode strictness: UR-P0-ctr-28) |
| `handlers_node.go:465-477` `writeAPIError` | mapping |
| `handlers_pairing.go` logs (`:79`, `:180`, `:197-199`) | **resolves the threat model's open X1 item:** `node.New` validates the passcode (`gf:node/surface.go:124`, `config.go:654`), so a `*node.Node`'s `SetupPayload` cannot fail with the "invalid passcode %d" texts of `gf:secure/setup/setup.go:84`, `:129`, `:250`; fresh credentials come from `node.NewWindowCredentials`. No passcode reaches a log line from these handlers on a real node. A host's own `NodeSource` could return such a text — its matter |
| `handlers_controller.go:648-671`, `:1259-1316`, `:1338-1464` window, commission, discovery, credentials | passcode, codes, passphrase, dataset in admin answers only; none logged or published; `progressDetail` kind only |
| `handlers_controller.go:300-323`, `:1636-1722` status, controller stream | every event type `[a-z_]+` (all 27 constants); unsubscribe on return |
| `stream.go:31-86` | unsubscribe deferred on every exit; `push` cannot panic or block (channel never closed); write and flush errors end the loop; type check before the frame |
| `events.go` | `Publish` calls outside the lock; a late call after unsubscribe lands in a buffered channel nobody reads; payloads carry no secret (window: discriminator) |
| `registry.go:95-107`, `:211-255`, `:259-308` | label validation, target resolution after the gate, error mapping |
| `nodeswitch.go` | atomic pointer, nil refused; a call spanning `Set` mixes targets harmlessly |
| `operations.go` | bounds of `since`/`limit`; level parse; stream reuse (gating: UR-P0-leak-27) |
| `spa.go` | path handling (invalid names and directories → index), CSP builder, `anyParent`, `$` without `(?m)` |
| `options.go`, `ports.go`, `problem.go`, `doc.go` | nothing beyond the findings above |

## Probes (scratch copy, package `fabricui`)

```go
func subsOf(h *handler) int { h.hub.mu.Lock(); defer h.hub.mu.Unlock(); return len(h.hub.subs) }

// 1 — UR-P1-conc-01. Output: HEAD,"X": while open 1, 2s after close 1 (leaked);
// HEAD,"" and GET: 0 after close.
func TestProbeHeadStream(t *testing.T) {
	for _, extra := range []string{"", "X"} {
		h, mux := newTestHandler(t, testOptions{})
		h.heartbeat = 20 * time.Millisecond
		srv := httptest.NewServer(mux)
		for _, method := range []string{"HEAD", "GET"} {
			conn, _ := net.Dial("tcp", srv.Listener.Addr().String())
			fmt.Fprintf(conn, "%s %s/node/stream HTTP/1.1\r\nHost: x\r\n\r\n", method, DefaultPrefix)
			time.Sleep(200 * time.Millisecond)
			fmt.Fprint(conn, extra)
			time.Sleep(200 * time.Millisecond)
			before := subsOf(h)
			conn.Close()
			time.Sleep(2 * time.Second)
			t.Logf("%s, trailing byte %q: while open %d, 2s after close %d", method, extra, before, subsOf(h))
		}
		srv.CloseClientConnections()
		go srv.Close()
	}
}

// 2 — UR-P1-api-02. "/a\tb", "/a//b", "/a/../b", "/a/." panic; "/a%zz" mounts.
func TestProbePrefix(t *testing.T) {
	for _, p := range []string{"/a\tb", "/a//b", "/a/../b", "/a/.", "/a%zz"} {
		func() {
			defer func() { if r := recover(); r != nil { t.Logf("%q PANICS: %v", p, r) } }()
			t.Logf("%q err=%v", p, Mount(http.NewServeMux(), Options{Prefix: p}))
		}()
	}
}

// 3 — UR-P1-err-03. Output: nil pointer dereference.
func TestProbeTypedNilEvents(t *testing.T) {
	var ev *Events
	defer func() { t.Logf("recovered: %v", recover()) }()
	_, mux := newTestHandler(t, testOptions{events: ev})
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(ctx, "GET", DefaultPrefix+"/node/stream", nil))
}

// 4 — gate and UR-P1-ctr-13, without a token. Output:
// POST /node/status 404 · HEAD /node/pairing 403 · GET /node/pairing/ 404 · GET /NODE/pairing 404 ·
// GET /node/%70airing 403 · GET /nodes/default/pairing 403 · HEAD /controller/discovery 403 ·
// POST /controller/discovery 404 · topology?refresh=false&refresh=true 400 · ?refresh=1 403 ·
// ?REFRESH=true 501 (not the parameter) · OPTIONS /node/pairing 404 · POST /nodes 403
func TestProbeMethods(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController()})
	for _, c := range [][2]string{{"POST", "/node/status"}, {"HEAD", "/node/pairing"} /* … as listed */} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(c[0], "http://x"+DefaultPrefix+c[1], nil))
		t.Logf("%s %s -> %d %s", c[0], c[1], rec.Code, rec.Header().Get("X-Problem-Code"))
	}
}

// 5 — UR-P1-ctr-17. Output: 200 {"level":"debug",…}
func TestProbeTrailing(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{operations: sampleOperations()})
	rec := do(t, mux, "PUT", "/operation/log/level", `{"level":"debug"}{"level":"nope"} trailing`)
	t.Logf("%d %s", rec.Code, rec.Body)
}
```

## Count

18 findings: 0 high · 1 medium · 9 low · 8 note. HELD-CLASS: 0.
