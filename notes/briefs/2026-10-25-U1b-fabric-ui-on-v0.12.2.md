# Brief U1b — go-fabric-ui v0.1.2: the handlers on go-fabric v0.12.2's facade routes (no more seam walks)

Date 2026-10-25. Repository `/home/markus/Dokumente/GitHub/go-fabric-ui`
(main = v0.1.1). Source: go-fabric `notes/briefs/reports/CK.md` §Questions
Q2 and §1 (the six routes U1 asked for), go-fabric `CHANGELOG.md` section
0.12.2 (`/home/markus/Dokumente/GitHub/go-fabric`, tag `v0.12.2`). Read
the repository's `CLAUDE.md`, `docs/reports/U1.md` §Questions (what each
workaround was), `fabricui/options.go`, `fabricui/ports.go`,
`fabricui/handlers_*.go`, `fabricui/events.go`, `fabricui/stream.go`.

**Model:** Opus. Branch `u1b/facade-routes`, one PR, no auto-merge, no
label, no attribution lines, English. Version `0.1.2` (`internal/version`,
`ui/package.json`, `changelog.md`; `contract/previous/` refreshed in the
release commit). Contract: no removed field; new optional fields only.

## 1 Replace every workaround with the facade route

- Pin go-fabric `v0.12.2`.
- `NodeSource` gains the methods the handlers now read; `*node.Node`
  still satisfies it (keep the compile-time assertion):
  - events: `Node.Observe(o)` instead of config-time observers — the
    stream subscribes on `Mount`/`Handler` and unsubscribes on close; a
    handler mounted after `Start` sees commissioning, fabric removal,
    window open/close (test).
  - window: `Node.Window()` (`WindowSnapshot` with `Kind` None/Own/Basic/
    Enhanced, `Discriminator`, `OpenedAt`, `Deadline`, `Timeout`) for
    `/node/status` and `/node/pairing/window`; the status response gains
    optional `window_kind`, `window_deadline` (contract, fixtures, page
    shows the countdown from the deadline).
  - setup payload: `Node.SetupPayload()` for `/node/pairing` instead of
    the configured-credentials re-derivation.
  - endpoints: `BridgedDevice.Clusters` (`ClusterInfo{ID, Name, Revision}`)
    and `DeviceTypeRevision` instead of the `Added.Spec()` walk; parts and
    snapshotter endpoints now carry clusters (test on a fabrictest device
    with a part).
  - fresh window credentials: `node.NewWindowCredentials(timeout)`; the
    module's own passcode rule goes.
  - sessions: `Node.SessionDiagnostics()` (`Sessions`, `Occupancy`,
    `Idle`/`PeerIdle` on the node's clock) and `Node.Clock()` instead of
    the manager aliases and wall time.
- Delete what the routes replace; `internal/arch` stays green; the
  advanced-seams mentions in `docs/embedding.md` go.

## 2 Tests, docs, report

Fixtures re-recorded only where a field was added (byte-identical
otherwise); Go handler tests and the e2e on fabrictest (`Site.AddNode`
with a `node.Config` if that reads cleaner than `Device.Node()`); vitest
for the countdown; `docs/embedding.md` and `docs/contract.md` updated;
`changelog.md` `## [0.1.2]`. Guards: `make ci` (web-check, dist-check),
`make race`, `make contract-check`, `make version-check`; `df -h /`
before, `go clean -cache -testcache` after. Report `docs/reports/U1b.md`
(PR body): per § with commit hashes; guards with HEAD; differences;
questions. Copy this brief to `notes/briefs/` in the first commit. Hand
back with the PR number and HEAD. If a permission is denied, say which
and stop; never ask the launcher to perform it.
