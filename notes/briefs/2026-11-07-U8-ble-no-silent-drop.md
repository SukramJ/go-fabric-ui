<!-- Written 2026-11-08; §11 decided by the owner the same day, and
     fabricui.NodeSwitch (§3.5) added on that decision. The map
     go-fabric-server notes/reads/read-s7-ble-pairing.md is an inventory,
     not a citation: every file:line below was re-read at the pins named
     here. Issue #49's file:line were taken at v0.6.0; they are re-cited at
     main. -->

# Brief U8 — the Commission page's BLE decision is stated, never silent (go-fabric-ui)

Date 2026-11-08. Issue SukramJ/go-fabric-ui#49 (host report
SukramJ/go-fabric-server#48, confirmed on the installation 2026-11-06: the
same device paired over BLE once the Thread credential was selected). Its
proposal is a suggestion; this brief settles the rule (§1); §11 holds the
owner's decisions of 2026-11-08, and §3.5 the second thing U8 ships:
`fabricui.NodeSwitch`, the swappable `NodeSource` both erasing hosts
(go-ha2fabric E1, go-fabric-server S7) designed as a copy of their own. **Model:** Opus. Repo
`~/Documents/GitHub/go-fabric-ui`, worktree `../go-fabric-ui-u8`, branch
`u8/ble-no-silent-drop`, cut from main `7200da3` (= v0.7.0). Pins:
go-fabric **v0.18.0** (local `f1e9677e` = v0.18.0+1; `git diff --stat
v0.18.0 HEAD` over `controller/controller.go`, `controller/pair_race.go`,
`controller/commissioner/progress.go`, `client/discovery/session.go`,
`controller/ble/central.go`, `fabrictest/ble.go` is empty, so the lines
below are v0.18.0's), go-fabric-server main `b947c93` (v0.6.1, read-only:
the `/ws` rule this brief must agree with; its own fix is wave S7, another
brief), matterjs-server `0bd06f2` (the comparison target for how a bridge
host decides). Pattern: `notes/briefs/2026-11-07-U7-completeness.md`. One
wave, one pull request, one hand-back. Version **0.7.1** with the release
commit folded in (§7, §11 3); contract **0.7.1** (`contract/openapi.yaml:4`
holds 0.7.0); `contract/previous/openapi.yaml` = the contract as released
in 0.7.0 (today it holds 0.6.0, `contract/previous/openapi.yaml:4` — copy
in the first commit, as U7 did).

**Hosts today** (read-only, 2026-11-08): go-fabric-server wires the pages'
source with `Credentials: credentialStore{api}` and `Pair:
api.PairOptions()` (`internal/daemon/controllerui.go:30-36`), so the pages'
`Pair.BLE` is the daemon's central (`controllerapi/host.go:162-168`);
its `credentialStore.List` maps `api.Credentials()` to `{Kind, ID, Name}`
(`controllerui.go:66-72`, `controllerapi/host.go:51-67`: the stored
entries only, an entry Home Assistant stored without an id under its id
`default`, `host.go:69-70`, `settings.go:21-23`, `cmd_server.go:195-200`).
go-ha2fabric has no controller family. §5 cites these lines. For the
node family both hosts pass a `*node.Node` today (go-ha2fabric
`cmd/ha2fabric/main.go:112-116`, the fleet `nodes.go:232-237`;
go-fabric-server `internal/daemon/web.go:39`, `registry.go:202` — the
two host briefs' lines, not re-read here) and their erase waves replace
it with a following source (§3.5).

**Not this repository's rule:** what a host does when its node fails to
come back after an erase is the hosts' (owner, 2026-11-08: the daemon
stays up, the host's status shows the error). `NodeSwitch`'s doc (§3.5)
must not assume a restart succeeds: `Set` is called with a live source
only; until then the old target answers.

## 0 Working rules

Commits as `SukramJ <sukramj@icloud.com>` (`git var GIT_AUTHOR_IDENT`
first), `<area>: <what> (U8)`, no attribution lines, stage named paths.
Guards under `flock ~/.cache/go-fabric-scratch/guards.lock` with
`export PATH=$HOME/go/bin:$PATH GOTMPDIR=$HOME/.cache/gotmp`, logs and
`summary.log` in `~/.cache/go-fabric-scratch/u8/`. No change in go-fabric
or go-fabric-server: a gap there is an issue text in the hand-back (§5,
§11). Every go-fabric line you build on is read at v0.18.0 and cited
`file:line`. Official text in English. The code and the passcode reach no
log line and no event (`CLAUDE.md`, Security) — the new log line and event
carry a reason word, nothing from the request.

## 1 The rule, as verified against the code

**Today.** `source.Pair` sets the central on the per-call copy only when
`s.opts.Pair.BLE != nil && !o.NetworkOnly && (pc.Manual ||
pc.Capabilities&capabilityBLE != 0) && (wifi != nil || th != nil)`
(`fabricui/gofabric/gofabric.go:273-277`; `capabilityBLE = 1 << 1`,
`:255-257`); `networkOf` answers no network for empty ids (`:229-232`), so
a stored credential that is not picked has no effect; the function has no
log call (`:264-289`; the adapter's only one is `:542`). The page's select
starts at "No network credential" (`Commission.svelte:30`, `:49-50`) and
names an id only when picked (`:54-58`); `capabilities.ble` is
`Pair.BLE != nil` regardless (`gofabric.go:162-164`,
`handlers_controller.go:307-311`). go-fabric then never learns of the
central: the BLE scanner is built only from `PairOptions.BLE`
(`controller/pair_race.go:49-53`), BLE without a network is
`ErrConfiguration` (`controller/controller.go:674-676`, field doc
`:587-593`), and a QR code without bit 1 drops BLE in
`SelectTransports` even when the scanner exists (`client/discovery/
session.go:73-76`, `:78-93`; mirrored by the pre-check). go-fabric's own
log lines cannot say it: `discovery.initiating`'s note "BLE not
requested" needs a BLE transport that was not selected (`session.go:
391-404`), `ble_not_enabled` a QR code naming BLE (`:409-411`). Nothing in
go-fabric is wrong (issue, "Why"; confirmed go-fabric-server#48, comment
2026-11-06 18:24).

**`/ws` today.** `bleNetwork` resolves the ids given, else the `default`
entries (`controllerapi/cmd_commission.go:76-93`, `settings.go:145-181`),
then the same four-input decision (`cmd_commission.go:132-135`), no log
(`:106-143`). matterjs-server: `wifi_credentials_id ??
DEFAULT_CREDENTIAL_ID`, an unknown explicit id refused, the credentials
applied when `!network_only && bleEnabled`
(`packages/ws-controller/src/server/WebSocketControllerHandler.ts:
1311-1348`; `:1315-1316`, `:1319-1324`, `:1328-1340`).

**Settled (U8).** The decision function stays the one both paths share;
what changes is that it is *stated* at three places and that the page
puts the network in before the operator starts. Inputs: **A** central
(`s.opts.Pair.BLE != nil`), **B** code allows BLE (`pc.Manual ||
pc.Capabilities&2 != 0`), **C** network from the request's named ids
(`networkOf`, unchanged), **D** `o.NetworkOnly`.

| A | B | C | D | BLE | Stated by U8 |
| --- | --- | --- | --- | --- | --- |
| no | any | any | any | no | event `ble{used:false, reason:no_central}`; no log line (the capability already says it; a QR code with bit 1 gets go-fabric's `ble_not_enabled`, `session.go:409-411`) |
| yes | any | any | yes | no | log `fabricui.pair.ble_skipped reason=network_only`; event `ble{used:false, reason:network_only}` |
| yes | no | any | no | no | `reason=code_without_ble` (the device's QR says it does not advertise on BLE) |
| yes | yes | none | no | no | `reason=no_network` — the issue's case; the page has already preselected or hinted (§4) |
| yes | yes | named | no | **yes** | event `ble{used:true}`; go-fabric's `Discovery.BLE` 2.1 step follows (`pair_race.go:25-26`, `:52`, `:243`) and the ESPHome central's `proxy_taken` (issue, "Why") |
| any | any | named, not stored | any | — | 404 `not_found` before the pairing, unchanged (`gofabric.go:239-250`, `handlers_controller.go:178-179`); no store with an id named: 400 (`gofabric.go:233-235`) |

Precedence when several reasons hold: `network_only` (the operator's
word) > `code_without_ble` (the device's) > `no_network`. Emitted once per
`Pair`, before go-fabric's first step; `CommissionDiscovered`
(`gofabric.go:442`) has no BLE option (the device came from a DNS-SD
discovery) and emits nothing — the page shows no BLE line in that flow.

**Not in the rule, with the reasons (decided, §11 1):**

- *No server-side default fallback on the pages' path.* `/ws` falls back
  because its schema has no selection (Home Assistant's `matter/set_thread`
  stores without an id — the issue comment's citation, `api.py:195-204` at
  core `75684ea787c`, not re-read here). The pages have a select: the
  network is a visible, explicit input, and "No network credential" must
  mean what it says. A fallback would make that option mean "the host's
  default, if any", which the page cannot show (the list carries no
  default flag today) — a convenience seam. Instead the list says which
  entry is the host's default (`Credential.default`, §2) and the page
  preselects it (§4); a stored Home Assistant dataset is listed under id
  `default` with its network name (`host.go:51-67`) and so is preselected
  when it is alone or flagged.
- *No refusal before the start (no new problem code).* A manual code says
  nothing about where the device is, and a QR code's BLE bit is the
  factory's: an already-commissioned device opened for a second fabric is
  on IP whatever the code says, which is why go-fabric keeps the IP
  transport regardless of the bit (`session.go:85-86`, matter.js's
  selection, `pair_race.go:34-36`). A refusal would make `network_only`
  mandatory for every Ethernet device on a host with a central; neither
  `/ws` nor matterjs-server refuses. The pairing runs, and says why BLE is
  not part of it.

**Best architecture?** Yes: the decision stays one function on both
paths (the port's inputs, go-fabric's refusal untouched); every drop is
visible before (page), during (event) and after (log) the pairing; the
only new notion — which entry is the host's default — is a fact the host
already has (`defaultCredentialID`) and now states instead of the port
guessing it from an id. The alternatives bought convenience with a hidden
input or correctness with an extra click for the common Ethernet case.

## 2 Contract 0.7.1 (additive; `make contract-check` refuses removals, `ui/scripts/contract-guard.mjs:22-59`)

- **`ControllerEvent.ble`** (`openapi.yaml:3162-3177`) → new
  `BLEDecision{used: boolean (required), reason: enum [no_central,
  network_only, code_without_ble, no_network] (present when used is
  false)}`: on `commissioning_progress` with `phase: started`, once, before
  the first `step` — the shape of `step` (`:3171`, `:3178-3186`: carried
  while the top-level phase stays `started`, so an old client keeps reading
  "started", `controller-store.svelte.ts:133-145`).
- **`Credential.default`** (schema `Credential`; port `controller.go:
  652-657`): boolean, optional, "the host's default entry: the one a
  commissioning started without a choice would use on the host's own
  interfaces; the pages preselect it". Absent = false.
- **`ControllerCapabilities.ble`** (`:2607`) description rewritten: "The
  host has a Bluetooth LE central. A commissioning searches over it when
  the code allows BLE, `network_only` is off and the request names a
  credential; the `commissioning_progress` event's `ble` says what was
  decided." The field's value is unchanged.
- **`CommissionRequest.network_only`** (`:3055`) description gains "and
  the event's `ble` reason is `network_only`".
- **Problem codes:** none new (`:1859-1861`; §1).
- `make generate`, `make gen-types`; fixtures re-recorded with `-update`
  and reviewed: `controller-credentials.200.json` (mesh gains `default:
  true`, §3.2), new `controller-status.200-ble.json` (a memory controller
  with a central: `capabilities.ble: true`) — used by the Go handler test
  and the vitest page test (§3.4, §4), as `TestEveryFixtureIsUsedByTheGoTests`
  and `fixtures.test.ts` require.

## 3 Go half

### 3.1 Ports — how each addition stays compatible

- **`PairOptions.Transport func(BLEDecision)`** (`fabricui/controller.go:
  565-582`): a struct field, as `Progress` is (`:577-581`), so
  `ControllerSource.Pair` keeps its signature. Doc: "Transport, when set,
  is called once on the commissioning's goroutine, before the first
  Progress report, with whether the search runs over Bluetooth LE and, if
  not, why; it must not block. A source without a central reports
  `Used: false, Reason: BLESkipNoCentral`." `BLEDecision{Used bool;
  Reason BLESkipReason}`; `type BLESkipReason string` with
  `BLESkipNoCentral = "no_central"`, `BLESkipNetworkOnly =
  "network_only"`, `BLESkipCode = "code_without_ble"`, `BLESkipNoNetwork
  = "no_network"` — the contract's enum, one vocabulary. Not a step: go-
  fabric's `Discovery.BLE 2.1 skipped` means "lost the race"
  (`pair_race.go:202`, `:218-220`); the port never synthesises a step
  go-fabric did not report.
- **`Credential.Default bool`** (`controller.go:652-657`): a struct
  field; a host that sets nothing lists no default (go-fabric-server sets
  it in S7, §5). The `Credentials` lookup interface (`gofabric.go:47-55`)
  is unchanged: `WiFi(id)`/`Thread(id)` by the named id only; the port
  gives `""` no meaning (issue, open question 2 — answered by the list,
  not the lookup).
- **`MemoryController.BLE bool`** (`memory_controller.go`, beside
  `Topology` — set before the handler serves): `Capabilities()` returns it
  (`:209`); `Pair` (`:233-247`) reports through `o.Transport` with the
  inputs it has — `Used = m.BLE && !o.NetworkOnly && (o.WiFiID != "" ||
  o.ThreadID != "")`, reasons `no_central`, `network_only`, `no_network`
  in that order; `code_without_ble` is the adapter's alone, because only
  go-fabric parses a code (say so in the comment). `record` keeps
  `Transport` out of `lastPair` as it does `Progress` (`:259-263`).
  **`MemoryCredentials.MarkDefault(kind, id)`** (`:21-119`): a test seam
  that sets the flag `List` reports; `NewMemoryController` marks `mesh`
  (`:183-184`), so the fixtures and the browser smoke have a default among
  two entries.

### 3.2 Handlers and the stream

- `CommissionNode` (`handlers_controller.go:1251-1300`): `opts.Transport
  = func(d BLEDecision) { publish commissioning_progress{Phase: started,
  BLE: &bleDecision{Used, Reason}} }`, set beside `opts.Progress`
  (`:1277-1283`). `controllerEventData` (`:1639-1650`) gains `BLE
  *bleDecision json:"ble,omitempty"`; `bleDecision{Used bool
  json:"used"; Reason string json:"reason,omitempty"}` beside
  `pairingStep` (`:1653-1658`). `progressDetail` (`:1305-1318`) unchanged.
- Credentials route: the list maps `Default` → `default` (omit when
  false, so a recorded fixture of a host without defaults is unchanged).
- Status (`:307-311`): unchanged code; the new fixture only.

### 3.3 Adapter `fabricui/gofabric`

- **`bleFor(central bool, pc controller.PairingCode, networkOnly,
  network bool) fabricui.BLEDecision`**: the pure decision of §1, the
  precedence in its `switch`. `Pair` (`gofabric.go:264-289`) calls it
  after `networkOf`; `opts.BLE = s.opts.Pair.BLE` iff `d.Used` (replacing
  `:273-277`); then, in this order: `if s.opts.Pair.BLE != nil && !d.Used
  { s.log.Info("fabricui.pair.ble_skipped", slog.String("reason",
  string(d.Reason))) }` — the operator's question is "why was my central
  not used", so the line exists only when one exists; `if o.Transport !=
  nil { o.Transport(d) }` — always, the page's question is answered in
  every case. Then `progressFor`, `editFor` as today (`:278-279`). The
  doc comment `:259-263` is rewritten to cite `/ws` at main `b947c93`
  (`cmd_commission.go:132-135`) and this brief's §1, not v0.3.2.
- `CommissionDiscovered` (`:442-482`): unchanged; its comment says the
  decision does not apply.

### 3.4 Tests

- **Adapter table test `TestBLEFor`** (`gofabric_test.go`, beside
  `TestEditFor`/`TestProgressFor`, `:129-178`): the four inputs as a
  table — `controller.PairingCode{Manual: true}`, `{Capabilities: 2}`,
  `{Capabilities: 4}` (bits per `session.go:73-76`), central yes/no,
  network yes/no, `NetworkOnly` yes/no — every row of §1's table and the
  precedence rows (central + network_only + QR without bit → `network_only`;
  central + QR without bit + no network → `code_without_ble`). Built from
  `PairingCode` values, no parsing.
- **Adapter e2e over a real central** (`fabricui/gofabric/e2e_test.go`,
  the site `:96-102`): a site whose device is built with
  `fabrictest.UseWiFiCommissioning(cfg, driver)` (`fabrictest/ble.go:56`;
  `contract.WiFiDriver` is two methods, `contract/wifi.go:17-25`, so a
  fake of a few lines) and a source whose template carries `BLE:
  bletest.NewCentral(dev.BLE())` (`fabrictest/ble.go:28`,
  `controller/ble/bletest/bletest.go:48`; the pattern
  `controller/pair_race_test.go:29-45`) and a `MemoryCredentials` with one
  Wi-Fi entry. Two pairings through the handler: (i) `wifi_id` named →
  the stream's `ble{used:true}` and a `Discovery.BLE 2.1` step; (ii) no id
  → `ble{used:false, reason:no_network}`, no `Discovery.BLE` step, and the
  pairing still succeeds over DNS-SD (the fabrictest device is on the
  site's network, `fabrictest/ble.go:25`) — the regression test of #49.
  If the Wi-Fi site cannot be built beside the existing one (one
  `fabrictest` controller per site is the pattern, `e2e_test.go:96`),
  make it its own test function with its own site; if it still cannot
  run under `make e2e`, record the gap in the report with the line that
  blocks it.
- **Handler tests**: `TestCommission_BLEDecision` on a memory controller
  with `BLE: true` — body `{"code":…}` → events `started`,
  `started+ble{used:false, reason:no_network}`, the steps, `succeeded`;
  the wire bytes asserted as `TestCommission_TermsAndSteps` does
  (`handlers_controller_test.go:442-444`): `{"phase":"started","ble":
  {"used":false,"reason":"no_network"}}`; body with `thread_id: "mesh"`
  → `{"used":true}` and `lastPair.ThreadID`; `network_only: true` with
  an id → `network_only`. Without `BLE` → `no_central`. Fixtures:
  `controller-status.200-ble.json` recorded from the BLE memory
  controller; `controller-credentials.200.json` re-recorded with `default`
  on mesh. The secrets test (`docs/contract.md:532-535`) stays green: the
  event carries a reason word only.
- **Playwright** (`ui/e2e/smoke.spec.ts:11-39`, flow `:106-117`): the
  e2e server (`fabricui/e2eweb_test.go:40-42`) sets `ctrl.BLE = true`;
  SMOKES gains "controller commission ble" (testid `commission-ble-hint`
  is *not* shown, because the default is preselected — assert the select,
  `getByLabel("Network for the device")`, has value `thread:mesh`); a flow
  test "a commissioning without a choice uses the preselected network":
  fill the code, commission, `commission-ble` reads "Bluetooth LE: used".

### 3.5 `fabricui.NodeSwitch` — one swappable `NodeSource` for every erasing host (owner, 2026-11-08)

**Why here.** `Options.Erase`'s doc says the handler keeps the `NodeSource`
it was given and a host that builds a new node "hands over a source that
follows it" (`fabricui/options.go:231-244`); `entryOf` wraps that value
once, at construction, in `nodeAdapter{n: p}` (`handler.go:124-134`), and a
registry node's handler lives while its `Source` is the same value
(`sameEntry`, `registry.go:126-128`; `managedHandler`, `:164-176`). So every
host that erases in place writes the same delegator: go-ha2fabric's
`nodeSource{s}` (E1 brief §2, `cmd/ha2fabric/source.go`, "thirty lines",
its Q1 asks for `fabricui.FollowNode(current func() NodeSource)`),
go-fabric-server's `bridgeHalf` (S7 brief §2.1, "29 delegating methods",
its Q1 asks for `fabricui.NodeSwitch` over a `func() NodeSource`), and this
repository's own test double `followNode` (`e2e_test.go:426-495`). U8
exports it once; the hosts adopt it with the 0.7.1 bump (§5).

**Shape** (`fabricui/nodeswitch.go`, SPDX header, beside `options.go`):

- `type NodeSwitch struct { cur atomic.Pointer[NodeSource] }` — a pointer
  to the interface value, not `atomic.Value`: `atomic.Value` panics when
  two stores carry different concrete types, and a host may swap a
  `*node.Node` for a test double or a wrapper.
- `func NewNodeSwitch(initial NodeSource) *NodeSwitch` — panics on nil
  (`"fabricui: NewNodeSwitch: nil source"`).
- `func (s *NodeSwitch) Set(n NodeSource)` — the atomic swap; panics on nil
  (`"fabricui: NodeSwitch.Set: nil source"`). Why a panic and not "answers
  as absent": the handlers decide "no node" once, at construction — a nil
  `Source` gives no port and the node routes answer `nodeUnready`, 503
  `service_unready` "the host serves no node (Options.Node is nil)"
  (`handler.go:124-134`, `:519-521`); a non-nil one is wrapped in
  `nodeAdapter`, which calls the methods without a nil check
  (`ports.go:66-69`). A nil target would therefore panic inside the first
  request, not answer 503; the type refuses it at the call that is wrong.
- `func (s *NodeSwitch) Current() NodeSource`.
- Every method of `NodeSource` (`options.go:27-90`: **29 methods** — the
  count `sed -n 28,90p fabricui/options.go | grep -v '^\s*//' | grep -c '('`
  = 29, the same as S7's) as one line each delegating to `s.Current()`,
  in the interface's order; the shape of `followNode` (`e2e_test.go:434-495`).
- `var _ NodeSource = (*NodeSwitch)(nil)`.

**Form: hand-written, not generated.** 29 one-line methods are ~90 lines.
Completeness is the compiler's: a method `NodeSource` gains and
`NodeSwitch` lacks fails `make build` in this repository — which is the
point (S7 Q1: "the widening of `NodeSource` then stops reaching hosts";
`docs/embedding.md:96-102` already states the compile check as the port's
design). A generator would need a go/types tool (`make generate` is
oapi-codegen over the contract only, `Makefile:106-112`, `script/gen-api.sh`)
— a dependency decision for ninety lines the compiler already guards.
The mistake a generator would prevent and the compiler would not (a
forwarder calling the wrong method) is caught by the end-to-end test
below, which drives every node route through the switch.

**Doc comment** (the type's; it is what `docs/embedding.md` quotes):

> NodeSwitch is a NodeSource whose target the host swaps: the pages keep
> one source while the host erases its node and brings a new one (Options.Erase,
> NodeEntry.Erase). Every method delegates to the current target; Set
> replaces it atomically, after the new node started. Set is called with a
> live node only — a host whose node does not come back keeps the old
> target in place, and the pages read what a stopped node answers: its
> last health verdict (go-fabric node/health.go:74-84 at v0.18.0) and the
> erased store's errors (node.go:1119-1121; factory_reset.go:66-67 closes
> it); what the host says about the failure is the host's. A stream open
> across Set keeps the node it was opened on: its observers were registered
> on that node when the stream began (events.go:152-166), so the new node's
> transitions reach the page when it reconnects; the page's erase flow polls
> the status, which reads through the switch, until the node answers.
> As a registry node's NodeEntry.Source the same *NodeSwitch is the same
> value across a Set, so the node keeps its handler (registry.go:163-177);
> listing a new *node.Node as Source instead gets a fresh handler and is
> the contract's other path.

**Why the old target and not an empty 503 source** (owner's question of
2026-11-08, both weighed): no nil-safe empty `NodeSource` exists in the
code (`var _ NodeSource` is implemented only by `followNode` in the tests;
503 comes from `h.node == nil`, decided at construction, above). Building
one would be a second 29-method type in this repository — two shapes for
one port — and its 503 would say "the host serves no node (Options.Node
is nil)", a false statement about a node that failed to restart. The
hosts' rule is that *their* status shows *the error*; the switch cannot
carry it, so it carries the node that has it: the stopped one. Best
architecture: yes — one type, no second shape, no sentence that lies.

**Tests** (`fabricui/nodeswitch_test.go`, and `make race`, `Makefile:44-45`):

- `TestNodeSwitch_SwapUnderConcurrentReader`: two doubles
  `type stubSource struct{ fabricui.NodeSource; addr string }` (the
  embedded nil interface is the partial double; `LocalAddr` overridden),
  eight reader goroutines calling `sw.LocalAddr()` and `sw.Current()` in a
  loop while the test goroutine alternates `Set(a)`/`Set(b)` a thousand
  times; every read is `a.addr` or `b.addr`, never `""`; `-race` is the
  assertion that matters.
- `TestNodeSwitch_NilRefused`: `NewNodeSwitch(nil)` and `Set(nil)` panic
  with the messages above.
- **End-to-end**: `e2e_test.go:76` becomes `sw :=
  fabricui.NewNodeSwitch(s.device.Node()); opts.Node = sw; opts.Erase =
  func(ctx) error { err := s.device.FactoryReset(ctx); if err == nil {
  sw.Set(s.device.Node()) }; return err }`; `followNode` (`:426-495`) is
  deleted. `TestEndToEnd_EraseBringsAFreshNode` then proves the switch under
  every route the erase flow touches, and the test is the pattern the
  embedding page cites. Whether `Device.Node()` after a *failed*
  `FactoryReset` is still the old node is unverified (`fabrictest/device.go:
  396-416` not read); the `err == nil` guard makes the test independent of it.

## 4 Svelte half

- `Commission.svelte`: after `loadCredentials()` in `onMount` (`:44-47`),
  preselect when `network === ""`: the entry with `default`, else the only
  entry, as `kind:id` (`:49-52` shape); the operator can still pick "No
  network credential". The option label of a default entry appends
  `controller.credential.default` ("default"). Below the select
  (`:155-158`) a hint, testid `commission-ble-hint`, when
  `controller.status?.capabilities.ble && network === "" && !networkOnly
  && controller.credentials.length > 0`: "Bluetooth LE is used only with a
  network credential." Under the progress (`:181-203`), testid
  `commission-ble`, from `controller.progress.ble`: `used` → "Bluetooth
  LE: used"; else `controller.commission.ble_skipped.<reason>` — four
  texts; `no_network` reads "Bluetooth LE not used: no network credential
  was picked", which with "no device answered" is the answer #48 lacked.
- `controller-store.svelte.ts` (`:23-27`, `:37`, `:133-145`):
  `CommissionProgress.ble: BLEDecision | null`, set from `ev.data.ble`,
  cleared by the commissioning's own `started` with the steps.
- `en.json` (`:71-73` neighbours), `de.json`; `make web` in the same
  commit.
- vitest (`Commission.test.ts:19-29` pattern): "preselects the only
  credential" (a one-entry credentials answer; the sent body carries
  `thread_id`), "preselects the default among several" (the re-recorded
  fixture), "shows the hint while no network is picked on a host with a
  central" (`controller-status.200-ble`), "says why Bluetooth LE was not
  used" (the stream event from §3.4's wire bytes); a store test for `ble`.

## 5 Hosts' obligations

- **go-fabric-server** (main `b947c93`; its wave S7, brief there): bump to
  go-fabric-ui 0.7.1; `credentialStore.List` (`internal/daemon/
  controllerui.go:66-72`) sets `Default: e.ID == defaultCredentialID`
  (`controllerapi/settings.go:21-23`) — the pages then preselect the entry
  Home Assistant stored; the `commission_with_code` log line
  (go-fabric-server#48, "Left here") uses the same reason words
  `network_only`, `code_without_ble`, `no_network` under its own name, so
  one grep answers "why not BLE" on both paths; the default fallback in
  `bleNetwork` (`cmd_commission.go:76-93`) stays, with the line stating
  `default` as the source of the network. Hand-over text in the hand-back,
  ready to paste into S7.
- **Both hosts, the node family (§3.5):** E1 (go-ha2fabric) and S7
  (go-fabric-server) bump to go-fabric-ui **0.7.1** and pass a
  `*fabricui.NodeSwitch` as `Options.Node` (E1 `main.go:112-116`, S7
  `web.go:39`) and as `NodeEntry.Source` where the node is erased in
  place (S7 `registry.go:202`; E1's fleet hands over the raw node and
  takes the fresh-handler path, its §2 — both paths stay in the
  contract), calling `Set(newNode)` after the new node started and
  never on a failed restart. go-ha2fabric's `cmd/ha2fabric/source.go`
  (`nodeSource{s}`) and go-fabric-server's 29 delegating methods on
  `bridgeHalf` go; `clock`/`surface` imports taken for them go with
  them (E1 §7, S7 §2.1 arch). Hand-over text for both in the hand-back.
- **go-ha2fabric**, the controller family: nothing (it has none).

## 6 Docs and changelog

`docs/contract.md`: the Commissioning bullet (`:521-529`) gets §1's rule
in three sentences and the `ble` event; the Credentials bullet
(`:527-528`) `default`; "Identifiers, secrets and events of the
controller family" (`:530`–) the `ble` field; Problems (`:355-399`)
unchanged; Versioning (`:401-425`): "Contract 0.7.1 (wave U8) is additive
over 0.7.0: `ControllerEvent.ble`, `Credential.default`, two
descriptions. No problem code is new." and, after the paragraph's
minor-version sentence (`:403-405`), the one sentence decided in §11 3:
"A patch may add the optional field or event field a fix needs to state
its behaviour; it adds no route and no problem code." `docs/embedding.md`: the
`Credentials` row (`:213`) — `Default` is the host's statement, the pages
preselect it, the lookup is by the named id; the `Pair` row (`:214`) — the
central and §1's rule in one sentence, `PairOptions.Transport`;
`fabricui.pair.ble_skipped` named under the `Logger` row (`:215`).
`docs/embedding.md`, the node family: the `Node` row (`:28`) gains "or
a `*NodeSwitch`, the source a host swaps after an erase"; the compile-
check paragraph (`:96-107`) one sentence — a host that follows its node
uses `NodeSwitch` and writes no forwarder; and under "Erasing a node"
(`:111-128`) a new subsection **"Erase and a node that comes back"**
replacing the sentences from "The handler keeps the `NodeSource`" to
"over `Device.Node`)" (`:122-125`): the host builds the switch over its
first node, hands it to `Options.Node`/`NodeEntry.Source`, and after
`FactoryReset` and the new node's start calls `Set`; a failed restart
calls nothing and the pages read the stopped node (what the host shows
about it is the host's); a stream open across `Set` keeps the node it
was opened on until the page reconnects (owner-accepted 2026-11-07 for
`Options.Erase`), the erase flow's status polling reads through the
switch; the registry's other path (a new `*node.Node` as `Source`, a
fresh handler) keeps its sentence (`:125-127`). The end-to-end test's
name stays as the reference.
`changelog.md` `[0.7.1]` **Added**: "`fabricui.NodeSwitch`
(`NewNodeSwitch`, `Set`, `Current`): a `NodeSource` whose target the
host swaps after an erase, so the pages keep one source while the node
is remade; a stream open across the swap keeps the node it was opened
on until the page reconnects." **Fixed**: "The Commission page no longer drops
Bluetooth LE silently when no network credential is picked (#49): the
page preselects the host's default or only credential, hints while none
is picked on a host with a central, and the commissioning's progress says
whether BLE was used and why not (`ble` on `commissioning_progress`,
`fabricui.pair.ble_skipped` in the log). `Credential.default` names the
host's default entry." `README.md`: no change. `docs/reports/U8.md` in
the fixed shape (U7: Summary, §1 verification table, Per section with
commits, Guards, Differences from the brief, Decisions taken, Open
questions, Left out, Hosts).

## 7 Commit order (each green under `make ci`)

1. `contract: 0.7.1 — the BLE decision on commissioning_progress, Credential.default (U8)` — `openapi.yaml`, `contract/previous` = 0.7.0's contract, `make generate`, `make gen-types`; schemas only, so the generated strict server compiles; no fixture yet.
2. `fabricui: PairOptions.Transport, BLEDecision, Credential.Default, the memory controller's central, handler and stream, fixtures (U8)` — §3.1, §3.2, handler tests, the two fixtures.
3. `gofabric: bleFor, the ble_skipped line, the decision on the stream; table test, e2e over bletest (U8)` — §3.3, §3.4.
4. `fabricui: NodeSwitch — the NodeSource a host swaps after an erase; the e2e erase test over it (U8)` — §3.5: `nodeswitch.go`, `nodeswitch_test.go`, `e2e_test.go` (`followNode` deleted). Its own commit, before the UI commit: Go only, no bundle, bisectable on its own.
5. `ui: preselection, hint, the decision under the progress, store, i18n (U8)` — §4 with `make web`; vitest; Playwright.
6. `docs: contract, embedding, changelog (U8)` — §6, the Versioning sentence and the NodeSwitch subsection included.
7. `report: U8 (U8)`.
8. `chore(release): 0.7.1 — wave U8, contract 0.7.1` — `internal/version`, `ui/package.json:3`, the changelog section dated; as U7's `0ecef14` was folded into #52. The tag is the owner's after the merge.

## 8 Guards

Run: `make ci`, `make race`, `make e2e`, `make e2e-web`,
`make contract-check`, `make generate-check`, `make cover-check`,
`make version-check` (0.7.1), `make notices-check` (`ui/package.json`
moves). Skip with the reason in the report: nothing else applies (no
image, no Docker in this repository).

## 9 The architecture check

- **Better than the comparison targets.** matterjs-server and `/ws` decide
  the same four inputs and say nothing (`WebSocketControllerHandler.ts:
  1328-1340`; `cmd_commission.go:132-135`); their default fallback is a
  hidden input a UI cannot show. The pages make the input visible
  (preselection, hint), carry the decision on the stream the page already
  follows, and log it once with a reason; the default becomes a listed
  fact. No refusal that `/ws` does not have, so a client moving between
  the two paths meets one rule.
- **Better than before.** A stored credential now takes effect without a
  click when it is the only or the default one; a dropped BLE is never
  silent; `capabilities.ble` says what it means.
- **The right cut.** go-fabric keeps every Matter rule (transport
  selection by the QR bits, the refusal of BLE without a network,
  `controller.go:674-676`); the adapter decides from host-side inputs
  only, in one pure function, and reports — it synthesises no step; the
  page reflects, it does not decide; go-fabric-server states which entry
  is its default and wires nothing else. The port's one new callback is
  the shape `Progress` already has.
- **The right cut, the node family (§3.5).** The following source is
  lifecycle glue every erasing host writes identically (E1 §2, S7 §2.4
  say so themselves); it lives once, beside the port it implements, and
  a widening of `NodeSource` is absorbed here, where the compiler checks
  it, not in two hosts. The lifecycle itself (what is erased, how a node
  is remade, what a failed restart shows) stays the hosts'; the type
  carries no policy, only the current node.
- **Each design choice** carries its own "best architecture?" line in §1
  (the rule, no fallback, no refusal), §3.1 (a field, not a step; the
  list, not the lookup) and §3.5 (one type, hand-written, the old target
  on a failed restart).

## 10 Left out

- A refusal problem code (`network_required`) and the server-side default
  fallback on the pages' path: decided against, §1, §11 1.
- `NodeSwitch` re-registering observers on `Set` so an open stream
  follows the new node without a reconnect: the stream holds the
  unsubscribe functions (`events.go:152-166`), so the switch would have
  to own a subscription table; owner-accepted limitation (2026-11-07),
  stated in the doc, reopened only if a page shows a stale node after
  an erase.
- An empty `NodeSource` answering 503 for a node that did not come
  back: §3.5, rejected with the reason there.
- A "Bluetooth LE" row on the Operation page's log view: the line is a
  normal `slog` Info record; the page already shows records.
- Whether go-fabric's commissioner skips a Wi-Fi credential on a Thread-
  only device (the "Thread device onto Wi-Fi by accident" worry of the
  issue's open question 1): not read (`controller/commissioner/flow.go`
  network steps). Read it during the wave; if the flow provisions by the
  device's NetworkCommissioning features, one sentence in `docs/
  embedding.md`'s `Pair` row; if not, a go-fabric question in the report,
  not a check in the adapter.

## 11 Decided (owner, 2026-11-08)

1. **Neither fallback nor refusal** — the explicit network with the
   preselected default or only entry, the hint, the event and the log
   line (§1, §4); the alternatives (a default fallback as `/ws`, a 400
   `network_required`) stay rejected with §1's reasons.
2. **`Credential.default` in U8** — one optional field on both sides,
   go-fabric-server's S7 sets it (§2, §3.1, §5).
3. **Release 0.7.1**, with the one Versioning sentence in
   `docs/contract.md` (§6): a patch may add the optional field a fix
   needs to state its behaviour, no route, no problem code.
4. **`fabricui.NodeSwitch` in U8** (§3.5) — one exported following source
   for both erasing hosts, its own commit before the UI commit (§7); E1
   and S7 adopt it with the 0.7.1 bump (§5).

Unverified in this brief, to be read during the wave: the commissioner's
network-step behaviour (§10); that `getByLabel("Network for the device")`
reaches the kit `Select` (`Commission.svelte:155`, `ariaLabel`); that a
Wi-Fi fabrictest site runs beside the adapter's existing one (§3.4);
whether `fabrictest.Device.Node()` after a failed `FactoryReset` is the
old node (§3.5 — the test guards with `err == nil` either way); the host
briefs' lines for their wiring (`main.go:112-116`, `nodes.go:232-237`,
`web.go:39`, `registry.go:202`) are those briefs', not re-read. The
Home Assistant citation (`api.py:195-204` at core `75684ea787c`) is the
issue comment's, not re-read here, and nothing in this repository
depends on it.

## Hand-back

Per §: what landed (files, fields, components, tests by name); §1's table
with each row's test; guards table (guard, HEAD, exit); bundle sizes;
differences from this brief; decisions; left out; open questions; the
go-fabric-server S7 text (§5) and the `NodeSwitch` adoption text for E1
and S7 (§5) ready to paste; the §10 reading's result;
every go-fabric line cited, re-read at v0.18.0. Open the pull request
against main (English, hand-back as body, no labels, auto-merge off) and
hand back with its number.
