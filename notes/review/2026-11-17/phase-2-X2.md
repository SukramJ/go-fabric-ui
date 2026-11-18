# Phase 2 — X2 write and life-cycle flows

Reviewer X2 (Opus). HEAD `1056db2`; go-fabric `v0.22.2` (`gf:`), hosts
`srv:` `d61a545`, `ha2f:` `ec7deea`. Tests written and run in the scratch
worktree `p2wt/x2` only (files `fabricui/x2_internal_test.go`,
`fabricui/x2_e2e_test.go`, `fabricui/x2_race_test.go`,
`fabricui/gofabric/x2_test.go`); nothing in the review worktree but this
file.

**Held candidates: none.**

Counts: 1 medium, 5 low (one of them the assigned UR-P1-err-22), 3 note.

| ID | sev | one line |
| --- | --- | --- |
| UR-P2-adpt-30 | medium | a chunked list write (Binding, any list through `PUT …/attributes`) answers the REPLACE-ALL's status; a refused element still shows "saved" |
| UR-P1-err-22 (assigned) | low (was note) | confirmed: erase on the request's context; a client that leaves cancels the host's erase, logged at debug only |
| UR-P2-err-31 | low | same class, wider reach: `DELETE /nodes/{key}` runs `NodeRegistry.Delete` on the request's context; reachable on go-ha2fabric |
| UR-P2-conc-32 | low | a stream open across `NodeSwitch.Set` stays on the erased node for good; the new node's events never reach it |
| UR-P2-bound-33 | low | a client that stops reading holds its stream's goroutine and subscriptions forever; the heartbeat blocks too (new path of UR-P0-bound-29) |
| UR-P2-conc-34 | low | the host's `Shutdown` waits its whole budget for every open stream; streams and observers outlive it; nothing in the API or docs ends them |
| UR-P2-ui-35 | note | exposure apply answers 200 and the page toasts "saved" when the rebuild failed |
| UR-P2-adpt-36 | note | remove-fabric on a controller node reports success when the NOCResponse does not decode (I) |
| UR-P2-err-37 | note | a write go-fabric refuses before sending (`ErrRequestTooLarge`, `ErrMalformedRequest`) answers 502 "The node failed" (I) |

---

## UR-P2-adpt-30 — a chunked list write reports success when an element was refused

- **Severity:** medium (a page shows success on failure, and the node is left with a partial list; precondition: the node refuses one element — table full, constraint)
- **Where:** `fabricui/gofabric/node.go:585-588`; `gf:client/interaction/write.go:259-275` (a list not ACL/Extension is REPLACE-ALL + one append per element), `:191-198` (one status per path, concatenated); `ui/src/lib/pages/controller/node/Bindings.svelte:50-52`, `Attributes.svelte:103-105` (toast "saved"/"written" on `status === 0`)
- **Claim:** `Write` returns `statuses[0]`, which for a chunked list is the REPLACE-ALL's status. A later append the node refuses is dropped. Reaches `PUT /controller/nodes/{id}/bindings` (`bindingList`, chunked: Binding is not excluded) and every list written through `PUT /controller/nodes/{id}/attributes`. The ACL is not chunked (excluded upstream), so `PUT …/acl` is not affected. Same function: an empty status list with a nil error answers success (`:595`).
- **M** — run. 256 labels to the root's UserLabel LabelList (limit 255, `gf:cluster/core/user_label.go:22`, `:98-100`):
  ```go
  // fabricui/gofabric/x2_test.go (gofabric_test), after commissioning as TestEndToEnd_Controller
  labels := make([]string, 0, 256)
  for i := range 256 { labels = append(labels, fmt.Sprintf(`{"label":"l%d","value":"v"}`, i)) }
  body := `{"endpoint":0,"cluster_hex":"0x0041","attribute_hex":"0x0000","value":[` + strings.Join(labels, ",") + `]}`
  s.json(http.MethodPut, "/controller/nodes/"+id+"/attributes", body, http.StatusOK, &wr)
  // read back 0x0041/0x0000 into got
  if wr.Status == 0 && len(got) != 256 { t.Fatalf("the write answered success (status 0) but the node holds %d of 256 labels", len(got)) }
  ```
  ```
  === RUN   TestX2_ChunkedListWriteReportsLaterFailure
      x2_test.go:52: write answered status 0x0; the node holds 255 labels
      x2_test.go:54: the write answered success (status 0) but the node holds 255 of 256 labels
  --- FAIL: TestX2_ChunkedListWriteReportsLaterFailure (1.01s)
  ```
- **Fix:** answer the first non-success status of `statuses` (else `statuses[0]`); an empty status list with nil error is `ErrNodeFailed`. Tried in scratch: the test passes with status `0x89` (ResourceExhausted). The pages may re-read after a non-zero status, since the list is now partial (Bindings already reloads; Attributes does not).
- **Contract:** no.

## UR-P1-err-22 (assigned) — erase runs on the request's context: confirmed

- **Severity:** low (reviewer; the pre-merge read had note). Not reachable on the pinned hosts, which detach inside their erase (`srv:internal/daemon/nodes.go:259`, `ha2f:cmd/ha2fabric/nodes.go:226`); reachable for a host that follows the documented pattern (`docs/embedding.md:150-165`, `fabricui/e2e_test.go:80-86`: the context goes straight to `FactoryReset`).
- **Where:** `fabricui/handlers_node.go:400`; against factory reset `:351` (`detached`, `51d7f79`) and reassemble `:306`; the failure's log `fabricui/problem.go:105-111`
- **Claim:** a client that leaves mid-erase (tab closed, proxy timeout) cancels the host's erase: `gf:node/factory_reset.go:45-68` runs RemoveFabric, stop and store erase on that context, so the node can stop half-reset. The resulting `context.Canceled` is logged at **debug** as `fabricui.request.abandoned`, so the operator never learns the erase broke.
- **M** — run:
  ```go
  // fabricui/x2_internal_test.go:40
  erase := func(ctx context.Context) error {
      close(started)
      select {
      case <-ctx.Done(): got = ctx.Err(); return got
      case <-time.After(500 * time.Millisecond): return nil
      }
  }
  _, mux := newTestHandler(t, testOptions{erase: erase, log: &log})
  leaving(t, mux, http.MethodPost, "/node/maintenance/erase", `{"confirm":"erase-node"}`, started) // cancels the request once erase started
  if got != nil { t.Errorf("the erase's context ended with the request: %v; log: %q", got, …) }
  ```
  ```
  --- FAIL: TestX2_EraseRunsDetachedFromTheRequest (0.00s)
      x2_internal_test.go:57: the erase's context ended with the request: context canceled; log: "… level=DEBUG msg=fabricui.request.abandoned error=\"context canceled\" method=POST path=/api/v1/node/maintenance/erase"
  ```
- **Fix:** `ctx, cancel := detached(r.Context()); defer cancel(); erase(ctx)` in `eraseNode`, as factory reset; log a failed erase at error (it is no longer an abandoned request). Say in the `Erase` doc and in `docs/embedding.md` that the handler runs it detached, bounded by two minutes.
- **Contract:** no.

## UR-P2-err-31 — registry Delete (and decommission) on the request's context

- **Severity:** low. New finding naming UR-P1-err-22 (wider reach).
- **Where:** `fabricui/registry.go:334`; host: `ha2f:cmd/ha2fabric/nodes.go:542-575` (`leave(ctx, …)` `:555`, `s.close(ctx)` `:557`, both on the request's context; errors only warned, then the files and record go, `:564-571`); `srv:internal/daemon/registry.go:300` detaches itself. Inferred second site: `fabricui/handlers_controller.go:559` → `fabricui/gofabric/gofabric.go:613` → `gf:controller/client/fabricops.go:245-256` (peer RemoveFabric, then `deleteNode(ctx, …)` on the same context).
- **Claim:** a client that leaves during `DELETE /nodes/{key}` on go-ha2fabric cancels the node's leave of its fabrics; Delete still removes the store and record and returns nil. The node is gone while every controller keeps it, with no Leave event. A decommission cancelled between RemoveFabric and the node's deletion leaves an uncommissioned record (I).
- **M** (handler; host read) — run:
  ```go
  // fabricui/x2_internal_test.go:83 — memoryRegistry whose Delete waits 500 ms or until ctx ends
  do(t, mux, http.MethodPost, "/nodes", `{"label":"Garden"}`)
  leaving(t, mux, http.MethodDelete, "/nodes/garden", "", reg.started)
  if reg.got != nil { t.Errorf("the registry's Delete ended with the request: %v; still listed: %v", reg.got, len(reg.List()) == 1) }
  ```
  ```
  --- FAIL: TestX2_RegistryDeleteRunsDetachedFromTheRequest (0.00s)
      x2_internal_test.go:91: the registry's Delete ended with the request: context canceled; still listed: true
  ```
- **Fix:** `detached(r.Context())` for `registry.Delete` (and for `Decommission`); the `NodeRegistry.Delete` doc says it runs detached. Host observation for go-ha2fabric's review: `fleet.Delete` should detach as `served.erase` does.
- **Contract:** no.

## UR-P2-conc-32 — a stream open across an erase stays on the erased node

- **Severity:** low. Overlaps X1's "erase" item on the Pair page side; this is the stream's life cycle.
- **Where:** `fabricui/nodeswitch.go:128` (and the health and trace observers through the same switch), `fabricui/events.go:152-160`, `fabricui/stream.go:89-121`; the docs' promise `fabricui/nodeswitch.go:25-29`, `docs/embedding.md:157-160` ("until the page reconnects"); the page `ui/src/lib/pages/node/Fabrics.svelte:104-117` (polls the status, never reopens the stream)
- **Claim:** the stream's observers are registered on the node current at subscribe time. After `Set`, nothing ends the stream (the heartbeat keeps it alive), so the page never reconnects: every later transition of the new node — `fabric_added`, `commissioning_complete`, `window_opened` — is lost until a reload. Reaches the default node on both pinned hosts and registry nodes on go-fabric-server (same `*NodeSwitch` value, so `sameEntry` keeps the handler, `fabricui/registry.go:183`).
- **M** — run (existing site helper, stream opened before the erase):
  ```go
  // fabricui/x2_e2e_test.go (fabricui_test)
  s := newSiteWith(t, true, true)
  s.expect(http.StatusNoContent, http.MethodPost, "/node/maintenance/erase", `{"confirm":"erase-node"}`)
  c := s.site.AddController(fabrictest.ControllerOptions{Label: "After"})
  if _, err := c.Commission(ctx, s.device, fabrictest.CommissionOptions{}); err != nil { t.Fatal(err) }
  s.getJSON("/node/status", &st)        // fabric_count=1: the new node is commissioned
  s.waitEvent("commissioning_complete") // never arrives
  ```
  ```
  === RUN   TestX2_StreamAcrossEraseSeesTheNewNode
      x2_e2e_test.go:31: status after commissioning the new node: fabric_count=1
      x2_e2e_test.go:32: no commissioning_complete event on the stream
  --- FAIL: TestX2_StreamAcrossEraseSeesTheNewNode (30.18s)
  ```
- **Fix:** either the switch moves its registrations: `NodeSwitch` records each `Observe`/`OnHealthSample`/`OnDiagnosticEvent` registration and on `Set` unregisters it from the old target and registers it on the new one (the stream keeps running); or `Set` closes a channel the streams of handlers over that switch select on, so the page reconnects. The first needs no page change. Correct the two doc sentences.
- **Contract:** no.

## UR-P2-bound-33 — a stalled reader is never reaped

- **Severity:** low. New path of UR-P0-bound-29 (no cap on open streams): even with a cap, stalled streams would hold it permanently.
- **Where:** `fabricui/stream.go:89-121` (no write deadline; heartbeat `:96-99` is a write in the same loop); hosts without `WriteTimeout`: `srv:internal/daemon/web.go:128`, `ha2f:internal/web/web.go:93-99`, whose comment says "the handler's heartbeat ends it on a dead client"
- **Claim:** a client that keeps the TCP connection open and stops reading (no RST, no FIN) fills the socket buffers; the next `Fprint` blocks for the life of the connection. The goroutine, the 64-event channel and every subscription (hub, host source, `node.Observe`, health, trace) stay. Any client on the listener's network can do this on the ungated node, controller and log streams. A heartbeat cannot detect it, because the heartbeat is blocked too.
- **M** — run (raw TCP client, receive buffer 4 KiB, never reads after the headers; heartbeat 50 ms; 64 KiB host events for 3 s):
  ```go
  // fabricui/x2_internal_test.go:98
  time.Sleep(2 * time.Second) // after the fill
  if n := subsOf(ev); n != 0 { t.Fatalf("a client that stopped reading still holds %d subscription(s) after %d heartbeats", …) }
  ```
  ```
  --- FAIL: TestX2_StalledReaderIsReaped (5.01s)
      x2_internal_test.go:135: a client that stopped reading still holds 1 subscription(s) after 100 heartbeats
  ```
- **Fix:** `_ = rc.SetWriteDeadline(time.Now().Add(2 * h.heartbeat))` at the top of each loop round in `serveSSE` (ignoring `ErrNotSupported`). Tried in scratch: the test passes and the existing `TestStream*` tests stay green.
- **Contract:** no.

## UR-P2-conc-34 — the host's Shutdown waits for every open stream

- **Severity:** low
- **Where:** `fabricui/stream.go:89-95` (ends on the request's context or `h.done` only); `fabricui/options.go:323-343` (no `Close`, no context in `Options`); hosts: `srv:internal/daemon/web.go:135-138`, `ha2f:internal/web/web.go:109-113` (both `Shutdown` with 5 s); `docs/embedding.md` says nothing on shutdown and streams
- **Claim:** `http.Server.Shutdown` does not cancel active requests, so an open page holds every shutdown for the whole 5 s budget, and the stream goroutines with their node observers outlive `Shutdown` while the host goes on to stop its nodes.
- **M** — run:
  ```go
  // fabricui/x2_internal_test.go:141
  srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
  // one GET /node/stream open
  err = srv.Shutdown(ctx /* 1 s */)
  if errors.Is(err, context.DeadlineExceeded) { t.Fatalf("Shutdown waited for the open stream until its budget ended") }
  ```
  ```
  --- FAIL: TestX2_ShutdownWithAnOpenStream (1.00s)
      x2_internal_test.go:162: Shutdown: context deadline exceeded after 1.001s; stream subscriptions left: 1
      x2_internal_test.go:164: Shutdown waited for the open stream until its budget ended
  ```
- **Fix:** docs first: `docs/embedding.md` tells the host to serve with `BaseContext` returning a context it cancels in `srv.RegisterOnShutdown`. Tried in scratch (`TestX2_ShutdownWithBaseContextRemedy`): `Shutdown: <nil> after 2ms; subscriptions left: 0`. A handler-side `Options.Done` would be an API addition (phase 3 `api`). Host observation for both hosts' reviews.
- **Contract:** no (docs); an `Options` field would be API.

## UR-P2-ui-35 — exposure apply shows success when the rebuild failed (note)

- **Where:** `fabricui/handlers_exposure.go:212-225` (rebuild error only warned, 200 `{applied}`); `ui/src/lib/pages/node/Selection.svelte:225-227` (toast "saved")
- **Claim:** **M** (read). Deliberate (comment `:186-188`, after loom), but the page cannot tell that the change is persisted and *not* in effect: the endpoints stay as they were until the next reassembly, and no `endpoints_changed` follows. A field (e.g. `reassembled: false`) or a warning on the status page would let it say so.
- **Fix:** a contract field, or reuse the host's `Warnings`. **Contract:** yes (field) → phase 3.

## UR-P2-adpt-36 — remove-fabric success when the NOCResponse does not decode (note)

- **Where:** `fabricui/handlers_controller.go:1270-1279`; `fabricui/gofabric/node.go:641-646` (a decode error leaves `Response` nil and `HasResponse` true)
- **Claim:** **I.** `json.Unmarshal(nil)` fails, so the NOCResponse status check is skipped and 204 is answered. Only a node answering a malformed NOCResponse reaches it.
- **Fix:** with `HasResponse` and no decoded status, answer `ErrNodeFailed`. **Contract:** no.

## UR-P2-err-37 — a write refused before sending answers 502 (note)

- **Where:** `fabricui/gofabric/node.go:592-593`, `gofabric.go:734-747` (everything but `ErrJSONValue`/`ICDMultiAdminError` is `ErrNodeFailed`); `gf:client/interaction/write.go:121-123`, `:229-280` (`ErrRequestTooLarge` for a timed write that does not fit one message; `ErrMalformedRequest`)
- **Claim:** **I.** A request that never left the controller answers 502 "The node failed" with go-fabric's text. `PUT …/attributes` with `timed_ms` and a list larger than one message is the reachable case.
- **Fix:** map `interaction.ErrRequestTooLarge` / `ErrMalformedRequest` to `ErrInvalidArgument` (400), if the facade exports them (else a go-fabric facade question). **Contract:** no.

---

## Covered without finding

**Writes**
- ACL: `handlers_controller.go:943-1006` decode, own-node CASE subject dropped; adapter `node.go:503-527` (`FabricIndex` ours, nullable targets), ACL not chunked (`gf:…/write.go:259`); read fabric-filtered `:894`; page lock-out confirm `Acl.svelte:87-97`, reload after save.
- Bindings: `:1078-1134` target shape checks; read chunked at `maxReadPaths`; page reload (`Bindings.svelte`) — apart from adpt-30.
- Attributes and commands: `:690-819`; `AttributeAt`/`CommandAt` are concrete paths (`gf:controller/client/surface.go:134-147`), 0xFFFF is not a wildcard; timed from the definition (`node.go:583`, `:632`); `ErrJSONValue` → 400 (`gofabric.go:738-739`); `StatusError` → status in the body.
- Fabric removal: node family `handlers_node.go:275-292` (1..254), factory reset detached `:351`; controller `:1250-1280` own-fabric guard via `ownIndex` (0 refused).
- ICD: `:1165-1193`, `node.go:835-855`; `ICDMultiAdminError` → 409 and the page's retry with `allow_multi_admin` (`Icd.svelte:40-44`); no other path maps to 409 there.
- OTA consent: `:1567-1591` ranges and node existence; page `Updates.svelte:41-53`.
- Exposure apply and reassemble: decode strict (`problem.go:134-139`), empty batch 204, rebuild detached (`handlers_exposure.go:216`, `handlers_node.go:306`); store `saveBulk` keeps pending on failure.
- Error mapping: `controllerError` `handlers_controller.go:168-201`; `serverError` detail empty.
- Routing to the right node: `nodeHandler`/`sibling` (`registry.go:198-211`, `handler.go:501-529`) carry the entry's port, exposure, hooks; the default node's Options-level fallbacks `handler.go:227-244`.

**Life cycle**
- Subscribe/unsubscribe: `serveSSE` defers the unsubscribe of all four sources (`stream.go:71`, `handler.go:638-654`, `events.go:152-167`); non-blocking push; HEAD answers without subscribing (node, log, controller `handlers_controller.go:1687-1705`).
- Controller stream: `subscribeController` `:1714-1729`, adapter `Observe` and `OnDiagnosticEvent` release (`gofabric.go:723-729`, `diagnostics.go:101-103`); `Source.Close` waits for the label push (`gofabric.go:697-706`).
- Registry vs cache: create, rename (handler kept, config and hooks refreshed `registry.go:183-186`), delete (`regGen`, stop `:338-344`), deletion overtaking a lookup (`:177-181`), a list that drops a key (`:155-162`). `go test -race -run 'Stream|Registry|NodeSwitch|Nodes|Exposure|Erase|Factory'`: ok; churn probe `TestX2_RegistryChurnUnderRace` (4 workers × 30 create/stream/rename/status/list/delete), `-race`: ok, no data race.
- Hosts' registry entries keep `Source`/`Exposure`/`Events` stable (`srv:…/registry.go:193-208`, `ha2f:…/nodes.go:435-449`), so handlers are not rebuilt per request. go-ha2fabric lists `Source` nil during an erase (`:441`), which rebuilds the handler twice and so happens to end that node's streams; the default node is not affected that way (conc-32).
- Unbounded growth: `regCache` bounded by the registry, `keyed` fixed at mount, hub subscribers by open streams (bound-33, UR-P0-bound-29).
- T20 on the write paths: answers decoded by go-fabric's IM layer; a bad list is 502 (`:899-901`, `:1056-1058`); no own bound beyond go-fabric's message size.
