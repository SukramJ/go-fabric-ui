# Report U1b — the handlers on go-fabric v0.12.2's facade routes (go-fabric-ui v0.1.2)

Brief: `notes/briefs/2026-10-25-U1b-fabric-ui-on-v0.12.2.md`. Source:
go-fabric `notes/briefs/reports/CK.md` §1 and Q2, `CHANGELOG.md` 0.12.2
(tag `v0.12.2`). Branch `u1b/facade-routes`.

## Per section: what landed

### Brief, pin — `87229f4`

go-fabric `v0.12.0` → `v0.12.2` (`go.mod`, `go.sum`); the brief copied to
`notes/briefs/`.

### §1 Every workaround replaced with its facade route — `8f1caad`, `960943b`

`NodeSource` gains `Window`, `SetupPayload`, `SessionDiagnostics`,
`Clock` and `Observe` and drops `PairingCodes`, `Sessions` and
`Subscriptions`; `*node.Node` still satisfies it (the compile-time
assertion in `misc_test.go` stays). `internal/arch` stays green: `clock`
is already on its facade list.

- **Events (U1 Q1).** `handler.subscribe` registers the hub, the host's
  `Options.Events` and `Node.Observe(nodeObservers(...))` together; the
  stream calls it on open and unsubscribes on close, so the observer lives
  exactly as long as the connection. Routes no longer publish
  `fabric_removed`, `window_opened` and `window_closed` themselves — the
  node's observer reports each transition once, whoever caused it. A
  repeated window snapshot of the same state is not a transition. Tests:
  `TestSubscribe_ObservesTheNodeOnce`, `TestStream_DeliversEventsAndHeartbeats`
  (the observer count goes 0 → 1 → 0 with the connection), and the e2e,
  whose handler is built after `AddDevice` started the node and sees the
  commissioning, the window closing after it, the fabric removal and the
  window close/open of `reopen`. `Events.Observers` is a deprecated
  pass-through (see Differences).
- **Window (U1 Q2).** `nodePort.window()` is `Node.Window()`. The status
  gains `window_kind` (`WindowKind.String()`: none/own/basic/enhanced) and
  `window_deadline` (RFC 3339, absent while no window is open); the
  duration is `WindowSnapshot.Timeout`. The pairing flow's hydration and
  the Status page count down from the deadline (`ui/src/lib/countdown.ts`,
  `secondsUntil`); without a deadline (a server before 0.1.2) the duration
  stays the upper bound. The fresh window's credentials are served while
  `Node.Window` reports a basic window with their discriminator before its
  deadline on the node's clock (`Node.Clock`) — the handler's own expiry
  timer and its hub-based forget are gone.
- **Setup payload (U1 Q3).** `GET /node/pairing` and the configured
  `POST /node/pairing/window` answer `Node.SetupPayload()`: the codes with
  the configured discriminator and passcode (admin routes; still never in
  an event or a log line).
- **Endpoints (U1 Q4).** `BridgedDevice.Clusters` (`ID`, `Name`,
  `Revision`) and `DeviceTypeRevision` replace the `Added.Spec()` walk;
  the response gains `device_type_revision`. Parts and snapshotter
  endpoints now list clusters. e2e: a `BridgedNode` with an `OnOffLight`
  part beside the plug — every endpoint lists Descriptor with names and
  revisions, the part lists OnOff.
- **Fresh credentials (U1 Q5).** `Options` default `fresh` is
  `node.NewWindowCredentials`; `freshCredentials`, `validPasscode` and
  `ephemeralIterations` are deleted. `TestFreshCredentials_ComeFromTheNode`
  checks the default against `node.ValidateSetupPIN`.
- **Sessions (U1 Q6).** `GET /node/sessions` reads
  `Node.SessionDiagnostics()`; idle ages are the node's `Idle`/`PeerIdle`
  on its clock; the handler's wall clock (`h.now`) is gone.
- **Docs.** `docs/embedding.md` (no observer wiring; what each route reads
  on the facade), `docs/contract.md` (route table, secrets, stream),
  README snippet.

Fixtures re-recorded where a field was added only:
`node-status.200` (+`window_kind`, `window_deadline`), `node-endpoints.200`
(+`device_type_revision`, cluster lists with names and revisions — the
fake's sample topology now carries them), `node-pairing.200`
(+`discriminator`, `passcode`). Every other fixture is byte-identical.
The diff removes no line but the three empty `"clusters": []`.

### Release — `54f4270` (the deprecation `960943b` follows it on the same version)

`0.1.2` in `internal/version`, `ui/package.json`, `ui/package-lock.json`,
`changelog.md`, `contract/openapi.yaml` `info.version`, README and
CLAUDE.md state lines; `contract/previous/openapi.yaml` refreshed after the
guard passed against the v0.1.1 snapshot.

## Guards (HEAD 960943b, clean tree)

| Guard | Result |
| --- | --- |
| `make ci` (build vet test lint web-check dist-check) | exit 0 |
| `make race` | exit 0 |
| `make contract-check` (before refresh, against v0.1.1) | exit 0 |
| `make contract-check` (after refresh) | exit 0 |
| `make version-check` | exit 0 |

`df -h /` before: 115 G free. `go clean -cache -testcache` after. This
report's commit is docs only; no second guard run.

## Differences from the brief

1. **`Events.Observers` is deprecated, not deleted.** go-ha2fabric
   (`cmd/ha2fabric/main.go:83`) and go-fabric-server
   (`internal/daemon/run.go:315`) call it; deleting it would break both
   builds in a patch release. It now returns its argument unchanged, so a
   host that still wires it gets each event once (from `Node.Observe`);
   it goes in v0.2.0. `TestEvents_ObserversIsAPassThrough`.
2. **Per-stream subscription.** "Subscribes on Mount/Handler and
   unsubscribes on close" is read as: the stream subscribes when a
   connection opens and unsubscribes when it closes. `Mount`/`Handler` have
   no close to unsubscribe from; a handler-lifetime observer would leak
   with every handler a host discards.
3. **Clusters are the mounted servers.** The old walk listed the source's
   application servers only; `BridgedDevice.Clusters` lists every mounted
   server in ServerList order, Descriptor and BridgedDeviceBasicInformation
   included — what a controller reads.
4. **Configured passcode served.** With `SetupPayload` the configured
   source of `GET /node/pairing` and the configured window answer carry
   `discriminator` and `passcode` (contract descriptions updated; the
   fields existed). Loom served the same.
5. **Own window's duration.** The uncommissioned node's own window reports
   its `Timeout` (on fabrictest longer than 18 h, clamped to 65535 s in the
   uint16 field); before, `RequestedDurationSeconds`.

## Questions

1. **`Events.Observers` removal in v0.2.0.** Recommendation: remove it in
   v0.2.0 (wave U2) and let go-ha2fabric and go-fabric-server drop their
   one wiring line on their next bump; the changelog announces it now.
2. **Countdown against the browser clock.** The page counts
   `window_deadline` down against `Date.now()`; a browser clock off from
   the node's shifts the countdown by that skew. Recommendation: accept
   (system clocks in practice); if needed later, an optional
   `window_remaining_seconds` computed on the node's clock is a
   non-breaking addition.
3. **`vendor_id` / `product_id` on `GET /node/pairing`.** `SetupPayload`
   carries them and loom served them; they are not in the contract.
   Recommendation: add as optional fields when a page needs them, not now.

## by_design candidates

None for go-fabric (no Matter behaviour is implemented here).
