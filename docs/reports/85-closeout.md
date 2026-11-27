# Report 85-closeout — the #85 close-out (go-fabric-ui v0.16.0, contract 0.16.0)

Brief: `notes/briefs/2026-11-27-85-closeout.md`; fact note
`notes/reads/read-85-closeout.md` (Fable, 2026-11-27). Branch
`wave/85-closeout`, cut from main `51bd01c` (v0.15.0); first commit
`26cf5ef` carries the notes. Built against go-fabric v0.28.0 (module
cache), read-only. Register rows
`notes/review/2026-11-17/phase-4r2-register.md:99,104,115,116` closed.

## Per section

### 1. ui-396 — the notice clears on `endpoints_changed` only — `a0a85e6`

`ui/src/lib/store.svelte.ts`: the case split; `endpoints_changed` clears
the notice, reloads the status and bumps the topology; `surface_changed`
does the latter two only. Comments rewritten from note a ("Verdict").
`Fabrics.test.ts`: the loop is two cases (`endpoints_changed` clears,
`surface_changed` leaves `true`). Docs: `docs/provider.md`,
`docs/embedding.md`, `docs/contract.md` (events paragraph). `make web`.

Bite: the old fall-through restored — `FAIL … > stays on the stream's
surface_changed`, `AssertionError: expected false to be true`.

### 2. err-404 — `ErrConfiguration` is 500, not 400 — `92c29a5`

`fabricui/gofabric/gofabric.go` `commissionError`: `ErrConfiguration`
leaves the `ErrInvalidArgument` case and returns the error without a
port kind; `ErrPairingCode` stays 400. Comment from note b ("Verdict
(revised)"). Tests: `TestCommissionError_ConfigurationIsTheHosts`
(neither `ErrInvalidArgument` nor `ErrNodeFailed`),
`TestSource_ClosedControllerIsNoRequestFault` (a real controller, closed:
`Pair`, `CommissionDiscovered` with a cached device and with the
re-discovery branch), `TestEndToEnd_ClosedControllerCommissionIs500`
(handler: 500, progress detail `failed`).

Bite: `commissionError` as before — `gofabric_test.go:197: configuration:
fabricui: invalid argument: …`, `:243` for each of the three paths, and
`e2e_test.go:1377: closed controller: 400 {…"code":"bad_request"}`.

### 3. api-391 — variant C — `e8dfda8`

Contract 0.16.0: `pingNode` declares 409 and 502 (description names
both); `contract/previous/openapi.yaml` is release 0.15.0's; fixtures
`controller-node-ping.409` and `.502` recorded with `-update` from
`errController` (`TestControllerFixtures_PingRefusedAndFailed`), used in
`controller-client.test.ts`; `make generate`, `make gen-types`, `make
web` (bundle unchanged).

`fabricui/handlers_controller.go`: `declared(err, kinds...)` beside
`opaqueError` keeps an `*apiProblem`, a `context.Canceled` (so
`serverError`'s abandoned-request path stays, UR-P1-err-12) and the
kinds passed; anything else becomes `opaqueError` (500, logged). The
groups `kinds400`, `kinds404`, `kinds409`, `kinds502` are the kinds
`controllerError` answers with each status. Sites, each with the groups
of its declared statuses (read from `contract/openapi.yaml`):
`getNodeEndpoint` 400/404, `getNodeCluster` 400/404 (after `notRead`),
`pingNode` 400/404/409/502, `discoverDevices` 400/502, `setFabricLabel`
400 (after its own 409), `getControllerTopology` 400/502 (after
`notImplemented`), `listCredentials` none, `setCredential` 400,
`removeCredential` 400/404, `consentUpdate` 400/404. Port godoc in
`fabricui/controller.go`; `docs/contract.md` (route row, versioning
entry, "A source's error kinds"), `docs/embedding.md`.

Guard: `TestController_SourceKindsOnlyWhereDeclared`
(`fabricui/undeclared_status_test.go`) — the seven kinds against 35
routes (`controllerRoutes` and twelve beside them), `errController` now
also failing `CommissionDiscovered`, `Credentials()` and
`SoftwareUpdates()`; TestMain's status guard judges.

Bites: the helper dropped at `getNodeEndpoint` — `status guard: … GET
/controller/nodes/{id}/endpoints/{ep} answered 409` and `… 502`; at
`listCredentials` — 404, 409, 502; at `setCredential` — 404, 409, 502;
at `consentUpdate` — 409, 502; the 502 removed from `pingNode` — `POST
/controller/nodes/{id}/ping answered 502`.

### 4. test-403 — the real update manager fails one product — `8b05411`

`fabricui/gofabric/updates_e2e_test.go` (new):
`TestEndToEnd_UpdateCheckFailsOneProduct`. Two devices of two products
(0x8101, 0x8102) with an OTA requestor, commissioned; `controller.
OTAConfig{Images: softwareupdate.NewImages({Blobs: MemoryBlobs}),
AllowTestImages: true}`; a newer local image of each product stored with
`Images.Store` (built with the node facade's `node.CreateOTAImage`), then
the second product's stored bytes changed in `MemoryBlobs` (last payload
byte flipped, header intact). The route answers one update (the first
node) and one `errors` entry (the second node); the stream carries one
`check_finished`. `updates.go` unchanged. The import guard unchanged.

Bites: `Check` returning `nil` for the failures — `updates_e2e_test.go:103:
errors [], want the node …`; a second `query` after the first —
`updates_e2e_test.go:121: 2 check_finished, want 1`. (A second query
before the first fails on `errors` instead: the first check removes the
invalid stored image.)

### Release — `f880039`

`internal/version`, `ui/package.json` (and the lock's two lines),
`changelog.md` `[0.16.0]`, contract 0.16.0 (in `e8dfda8`), `CLAUDE.md`
state paragraph.

## Guards

At HEAD `f880039` (before and after the run), on macOS with `gmake`:
`make ci` 0, `make race` 0, `make cover-check` 0 (floors unchanged),
`make vuln licenses secrets` 0, `GOOS=linux GOARCH=386 make build` 0,
`make web-install` 0, `make notices-check` 0, `make version-check` 0,
`make contract-check` 0, `make e2e-web` 0, `make e2e` 0,
`make generate-check` 0. `make test-386` runs on CI only (darwin). The
register-and-report commit after it touches notes and this report only:
no second run.

## Differences from the brief

- `declared` takes groups of kinds by status (the kinds `controllerError`
  answers with that status), not single kinds per site, so no status an
  operation declares and answered before changes. For `pingNode` that is
  wider than the brief's list (`ErrNodeNotFound`, `ErrConflict`,
  `ErrNodeFailed`, `DeadlineExceeded`): the other 404 kinds and the 400
  kinds keep their declared 404 and 400.
- test-403 lives in a new e2e file beside `e2e_test.go`, not in
  `updates_test.go` (it needs the site and the handler).
- `ErrConfiguration`'s test is in `gofabric_test.go` and `e2e_test.go`.

## Decisions

- A `context.Canceled` passes `declared` unchanged (the note's open
  point): `serverError` logs a client that left at debug, as before.
- `listCredentials` declares no 4xx/5xx beyond 500/503: every source
  error there is 500.

## Left out

- Nothing of the brief.

## Questions

- None blocking. Fable: the go-fabric statements listed below against the
  note.

## The hosts' obligations

- A host that rebuilds the node on its own publishes `endpoints_changed`
  through `fabricui.Events` after its rebuild returned nil, or an open
  page keeps the "not in effect" notice until a rebuild through the
  routes (go-ha2fabric, go-fabric-server, openccu-loom: whether they
  rebuild outside the routes is not read).
- A host's own `ControllerSource`, `CredentialStore` or
  `SoftwareUpdateSource`: an undeclared kind is 500 now (godoc per
  method).
- A commissioning on a closed controller, or with a `Pair` template that
  go-fabric refuses as `ErrConfiguration`, answers 500 (was 400).

## go-fabric statements written in this wave (for Fable)

- `ui/src/lib/store.svelte.ts:480-498` (both comments), note a.
- `fabricui/gofabric/gofabric.go:439-450`, note b.
- `fabricui/gofabric/gofabric_test.go:204-207`,
  `fabricui/gofabric/e2e_test.go:1359-1363`, note b rows 6, 7.
- `fabricui/gofabric/updates_e2e_test.go:23-34`, note d;
  `:47-49` (`manager.go:1971-1975`, read for the test, not in the note).
- `docs/provider.md:89-95`, `docs/embedding.md:148-153`,
  `docs/contract.md:265-270`, note a.
