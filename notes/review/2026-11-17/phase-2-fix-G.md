# Phase 2 — fix agent G (Go half)

Branch `review/p2`, base `29cf56b`,
HEAD `48bb394`. Every fix red first (run before the fix, failure text
recorded below in short), then green.

Guards at `48bb394`: `go test -race ./...` ok; `make lint` 0 issues;
`make contract-check` ok; `make e2e-web` 42 passed (Chromium; run without
`make web`, against the committed bundle). `contract/openapi.yaml`
untouched; `ui/` untouched; worktree clean.

| Finding | Result | Commit | Test (red before the fix) |
| --- | --- | --- | --- |
| UR-P2-conc-04 (server half) | fixed: `NodeSwitch.Set` closes a channel the node stream selects on (read before the stream subscribes, so a Set in between ends it too); controller and log streams pass nil. `nodeswitch.go` comment and `docs/embedding.md` (Erase and a node that comes back) corrected | `b3b5c05` | `TestEndToEnd_EraseEndsTheStreamAndTheNextFollowsTheNewNode` (red: "the stream opened before the erase is still open on the erased node"); the e2e `site` got `openStream()`/`ended` |
| UR-P2-conc-01 | fixed: the observer tells windows by kind and `OpenedAt`; open→open with another window is `window_opened`, a repeated snapshot stays silent. Handler streams start from the node's snapshot (internal `nodeObservers`); public `NodeObservers(push, open)` unchanged, a window open at registration is taken by its first snapshot. `docs/contract.md` Stream: "A `window_opened` may follow a `window_opened` …" | `40400f2` | `TestNodeObservers_AReplacedWindowIsWindowOpened` (red: one window_opened of two). The X1 e2e test was not ported: it imports `secure/spake2`, `cluster/spec/administratorcommissioning`, `controller/client`, which the arch guard forbids in `fabricui` tests |
| UR-P2-err-02 | fixed: `ephemeralWindow.openedAt` from the snapshot after the open; `servesEphemeral` compares kind and `OpenedAt`, the discriminator only when the snapshot names one | `9f18287` | `TestPairing_KeepsTheFreshWindowWhoseRecordWasNotPublished` (red: source configured). Fake node gained `unannounced` |
| UR-P2-conc-07 (inferred) | red test written → fixed: `openMu` from the window check to the stored credentials | `9f18287` | `TestOpenWindow_ConcurrentOpensOneWins` (red: `[200 200]`). Fake node gained `beforeBasic` |
| UR-P2-leak-08 (server part) | fixed: `forgetOnClose` registers an observer per fresh window; the transition that ends it (closed, replaced, deadline) drops the credentials; the observer unregisters itself off the node's goroutine | `9f18287` | `TestPairing_ForgetsTheFreshCredentialsWhenTheWindowCloses` (red: condition not reached) |
| UR-P2-adpt-30 | fixed: first non-success status of the chunked write; empty statuses with nil error → `ErrNodeFailed` | `120843e` | `TestEndToEnd_ChunkedListWriteAnswersTheRefusedElement` (gofabric; red: success) |
| UR-P2-adpt-03 | fixed: `commissionedDespite` — an error with a commissioned node (`IsCommissioned`) answers the node, logs `fabricui.pair.start_failed` at warn; both `Pair` and `CommissionDiscovered` (which now also forgets the cached device) | `120843e` | `TestPair_ACommissioningThatHappenedIsNoFailure` (red: context canceled) |
| UR-P2-adpt-36 (inferred) | red test written → fixed: `HasResponse` with a NOCResponse that does not decode answers 502 | `b2e49f1` | `TestRemoveNodeFabric_UndecodedNOCResponseIs502` (red: 204; the port-level double returns what the adapter returns on a decode error) |
| UR-P1-err-22 | fixed: erase on `detached`; failure via new `detachedError` (always error level). `Options.Erase` doc + `docs/embedding.md` (Erasing a node): detached, two minutes, failure at error | `a0cc661` | `TestErase_OutlivesTheRequest` (red: context canceled); also asserts level=ERROR for a host error wrapping `context.Canceled` |
| UR-P2-err-31 | fixed with the same change: `DELETE /nodes/{node}` (registry `Delete`) and `DELETE /controller/nodes/{id}` (`Decommission`, the inferred second site) detached; port docs say so | `a0cc661` | `TestRegistry_DeleteOutlivesTheRequest`, `TestDecommissionNode_OutlivesTheRequest` (both red: context canceled) |
| UR-P2-bound-33 | fixed: `SetWriteDeadline(now + 2·heartbeat)` before every write, `ErrNotSupported` ignored | `4870251`, `48bb394` (test cleanup: closing the response drained the stream for 60 s) | `TestStream_StalledReaderIsReaped` (red: condition not reached) |
| UR-P2-web-91 (decision 36) | fixed: `/` and `/index.html` serve the index; a file itself; anything else (`/x`, `/a/`, `/assets/`, `/node/status`, `/assets/nope.js`) a plain 404 with the common headers. Changelog [Unreleased] Changed line for the hosts; `docs/embedding.md` intro and "Response headers and framing" no longer say every path / every deep path. Existing tests that expected the index under `/deep/link` and `/assets/` changed to the root | `bd232da` | `TestSPA_OnlyTheRootServesTheIndex`, `TestSPA_TheIndexCanLoadItsAssets` (the X4 test, embedded bundle; red for `/a/`, `/a/b`, `/assets/`, `/node/status`). e2e-web still green (all its `goto` are `/` + hash) |
| UR-P2-conc-34 | docs: `docs/embedding.md` after the observer paragraph — Shutdown waits for open streams; the base-context + `RegisterOnShutdown` pattern | `56d1682` | guard of the pattern `TestStream_EndsOnTheHostsShutdown` (green, docs finding) |
| UR-P2-bound-65 | docs: `docs/embedding.md` after the counters paragraph — four host ports per status request, keep them cheap. Checked against `handlers_node.go` (warnings, counters, `fabrics`→`FabricInfos`, `exposure.List`) | `56d1682` | none (docs) |
| UR-P2-test-95 | partly: `TestStream_DataStaysOnOneLine` (X4's guard), `TestAPI_NoRouteAdmitsAnotherOrigin` (Go form of the no-CORS guard: no `Access-Control-*` on any contract route for a foreign `Origin`, no preflight answered 2xx). **Not done:** the Playwright guards T10 (framing) and T8/T9 (browser, second origin) — they live under `ui/e2e`, outside G's paths; the Go side of framing is already held by `TestSPA_FrameAncestors`/`TestSPA_IndexCarriesTheCSP`. | `a8e1183` | guards, green |

Changelog: `ddfca4f` adds Fixed lines for conc-04, conc-01, err-02,
conc-07, leak-08, adpt-30, adpt-03, adpt-36, err-22/31, bound-33.

Struck: none (both inferred findings got a red test).

## For the coordinator

- The page half of conc-04 (W): the server now ends the stream on `Set`;
  the store's reconnect + `resync` must run on that end (the
  `EventSource` reopens on its own; the decision whether the store resyncs
  on an `onerror`/reopen is W's).
- Fable reads due (CLAUDE.md model table): the corrected sentence in
  `fabricui/nodeswitch.go` and `docs/embedding.md` (verifier: reference
  statements); the new citations `gofabric/node.go` (go-fabric
  `client/interaction/write.go:259-275` at v0.22.2) and
  `gofabric/gofabric.go` `commissionedDespite` (`controller/client/fabricops.go:229-235`
  at v0.22.2); the `spa.go` comment ("the 404 for an unknown path is this
  module's", the loom citation kept).
- `NodeObservers` (public) cannot see a replacement of a window already
  open when it is called (it gets `open bool` only); `ObserveNode` and the
  handler's streams can. A snapshot parameter would be an API change
  (phase 3 `api` if wanted).
- `forgetOnClose` registers one node observer per fresh window; it stays
  on a node that stops without a transition (erase), released with that
  node.
- `make e2e-web` was run (it only builds the test binary and runs the
  committed bundle); `fabricui/dist` not rebuilt.

## Follow-up after the pre-merge read (base `027375b`)

| Finding | Result | Commit | Test |
| --- | --- | --- | --- |
| UR-P2-adpt-190 | fixed: `markSubscribe` wraps this call's progress hook (in `CommissionDiscovered` after the host's `Edit`) and marks step Subscribe 21.1 started; `commissionedDespite` requires the mark besides `IsCommissioned`. A host `Edit` that replaces `Progress` on the `Pair` path loses the mark → the error is answered as before (conservative) | `6dcaf30` | `TestCommissionDiscovered_ACommissionedNodeAgainIsNoSuccess` (red: "answered success"); `TestPair_ACommissioningThatHappenedIsNoFailure` stays green |
| UR-P2-test-192 | comment cut after "answered with success" | same | — |
| UR-P2-doc-193 | `nodeswitch.go` cites `registry.go:168-190`; `nodeswitch.go` and `docs/embedding.md` say Set ends the node streams served while the switch is `Options.Node` or a `NodeEntry.Source` | same | — |

Guards: `go test -race ./...` ok, `make lint` 0 issues. `ui/` and `fabricui/dist` untouched.
