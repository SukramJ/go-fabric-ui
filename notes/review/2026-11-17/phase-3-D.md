# Review phase 3 — unit D (cite), Fable

HEAD read `f4b8573`. Pins: go-fabric `v0.22.2` (`gf:`), openccu-loom
`b40b289a` (`loom:`), matter.js `862ad75f` (`mjs:`, docs only). Older
go-fabric tags read with `git show <tag>:<path>`; go-fabric-server,
matterjs-server, matter.js sources and connectedhomeip read from the local
clones at the commit the sentence names (no pin of this review — marked
"local clone"). Map: `notes/reads/read-review-p3-cite.md`.

M = measured (lines read on both sides, or a test run); I = inferred.

## Findings

### UR-P3-doc-90 — low — `docs/embedding.md` says go-fabric's `logring` has not landed

- **Where:** `docs/embedding.md:248-253` (handed over by unit C).
- **Sentence:** "go-fabric's `logring` package is that handler once it
  lands (go-fabric issue #393, concept U5 C.4); until then a host writes
  the adapter itself".
- **Cited side (M):** `gf:logring/logring.go:4-43` (package doc: a
  `slog.Handler` wrapping the host's handler, a `slog.LevelVar` gate, a
  bounded ring), `:116` `New(inner slog.Handler, opts Options) *Handler`,
  `:142`/`:146` `Level`/`SetLevel`, `:239` `Records(since uint64, limit
  int) []Record` (oldest first, newest `limit`), `:270` `Dropped`, `:331`
  `Observe(fn func(Record)) (unsubscribe func())`. This repo's own
  `changelog.md:616` lists `logring` among what a go-fabric bump brought.
- **Verdict:** the package exists at v0.22.2; the sentence is stale and
  tells a host to write what the module ships. The port's five methods
  (`LogLevel`, `SetLogLevel`, `Log`, `ObserveLog`, `Info`) still need a
  thin host adapter over the handler (`Record` → `LogRecord`, `Info`), so
  "is that handler" is right and "once it lands … until then" is not.
- **Fix (sentence):** "go-fabric's `logring` package is that handler
  (`logring.New`, `Handler.Level`/`SetLevel`, `Records`, `Observe`;
  go-fabric v0.22.2 `logring/logring.go:116`, `:142-146`, `:239`, `:331`):
  a host maps its `Record` onto `fabricui.LogRecord` and adds `Info`."
- **Contract:** no.

### UR-P3-adpt-91 — low — an ICD call the client's state refuses is answered as the node's failure

- **Where:** `fabricui/gofabric/gofabric.go:753-766` (`nodeError`), reached
  from `node.go:848`, `:860`, `:862`.
- **Claim:** the facade defines `client.ErrICDOperation` as a refusal of
  the client's own state — `gf:controller/client/icd_errors.go:80-83`
  ("a registration exists, another operation runs, no subscription, the
  peer offline or without CIP"), returned by `Register`
  (`gf:controller/client/icd.go:645-654`) and `Unregister` (`:818-826`).
  `nodeError` maps only `*ICDMultiAdminError` to `fabricui.ErrConflict`;
  every other `ErrICDOperation` becomes `ErrNodeFailed` (502), although
  the port defines `ErrConflict` as "the node refuses the action in its
  state (409)" (`fabricui/controller.go:903-905`) and nothing was sent to
  the node. A second click on "register" reads as a failed node.
- **M:** red test, scratch copy of `f4b8573`
  (`fabricui/gofabric/p3d_red_test.go`, package `gofabric`):
  `nodeError(fmt.Errorf("%w: a registration exists", client.ErrICDOperation))`
  must be `errors.Is(…, fabricui.ErrConflict)`; today:
  `nodeError(ErrICDOperation) = fabricui: the node failed: client: ICD
  operation refused: a registration exists, want ErrConflict` — FAIL.
- **I (same class, not tested):** `client.ErrDisabled`
  (`gf:controller/client/config.go:199`) on an interaction with a disabled
  node, and `client.ErrFabricOperationInProgress`
  (`gf:controller/client/registry.go:71`) on a second decommission, take
  the same path to 502. `ErrAlreadyCommissioned` is known
  (UR-P2-adpt-190) and not re-reported.
- **Fix:** `errors.Is(err, client.ErrICDOperation)` → `ErrConflict` in
  `nodeError` (it covers `*ICDMultiAdminError`, which "is an
  ErrICDOperation", `icd_errors.go:25-27`); `ErrICDAbandoned` stays as it
  is. Decide the two inferred ones with it.
- **Contract:** no — `POST …/icd/register` already lists 409
  (`contract/openapi.yaml:1533`).

### UR-P3-adpt-92 — note — held candidate: `CommissionDiscovered` leaves out what `Controller.Pair` records

- **Held candidate**: same function as the held UR-P1-adpt-60; a wider
  reach, not the attestation default.
- **Where:** `fabricui/gofabric/gofabric.go:567-619`.
- **Claim (M, lines):** `Controller.Pair` wraps the progress hook with
  `tracePairing` and ends with `recordPairFailed` / `recordPaired`
  (`gf:controller/controller.go:770-787`, `:795`): the pairing's start,
  each failed step, the paired node or the failed pairing enter
  `Controller.DiagnosticEvents`, and a failure counts once as
  `pair.failures`. `tracePairing`, `recordPaired` and `recordPairFailed`
  have no caller but `Pair` (`gf:controller/controller.go:782-795`,
  `controller/diagnostics.go:235-274`). `CommissionDiscovered` calls
  `Registry().ForDevice` and `Commission` directly, so a commissioning of
  a discovered device leaves no pairing entry in the trace that
  `GET /controller/events` serves and no metric; it also reports no
  Discovery 2.1 step (`Pair` with a known address reports it skipped,
  `:823`).
- **Fix:** a facade entry for a discovered device and a passcode
  (go-fabric issue); until then one sentence in
  `docs/embedding.md`'s `POST /controller/commission` row.
- **Contract:** no.

### UR-P3-cite-93 — note — go-fabric ranges without a tag that are right only at an older tag

Each sentence names no tag, so it reads at the pin (`go.mod` v0.22.2); the
construct exists unchanged, at other lines. M.

| Citing line | Cited | At v0.22.2 the range is | The construct at v0.22.2 |
| --- | --- | --- | --- |
| `fabricui/controller.go:693-694` (`PairOptions.Progress`) | `controller/controller.go:611-618` | `Controller.Close`'s tail (`c.brs.Close()` …) | `controller/controller.go:720-727` (right at v0.18.0, as `gofabric.go:366` says) |
| `fabricui/options.go:80-81` (`BootReason`) | `node/node.go:1138-1141` | the doc of `Node.RemoveFabric` | `node/node.go:1195-1199` (right at v0.18.0) |
| `fabricui/options.go:83-84` (`SessionParameters`) — map: out-of-range | `node/sessions.go:110-118` | file has 112 lines | `node/sessions.go:105-112` (v0.18.0: `:110-112`) |
| `fabricui/handlers_surface.go:23-24`, `fabricui/support_test.go:598-599` (`marshalEnum`) | `surface/element.go:135-140` | starts in the previous function's `}` | `surface/element.go:137-142` (right at v0.15.2) |

- **Fix:** the corrected ranges, or "at v0.18.0" / "at v0.15.2" added as
  the neighbouring comments do.
- **Contract:** no.

### UR-P3-cite-94 — note — openccu-loom ranges that miss the construct they name

M, at `b40b289a`.

| Citing line | Sentence names | Cited | What the cited lines are | Corrected |
| --- | --- | --- | --- | --- |
| `fabricui/problem.go:36-37` | `Details` | `internal/north/rest/problem/problem.go:50-59` | the `Type` constants (`feature_unavailable`, `pairing_*`) | `:78-89` |
| `fabricui/problem.go:49` | `Write`, `New` | `problem.go:61-85` | `FeatureRef`, `FieldError`, half of `Details` | `:91-118` |
| `fabricui/handlers_node.go:318-320` | the confirmation word's reason | `matter_maintenance.go:58-61` | the doc of `MatterFactoryReset` | `:54-57` |
| `fabricui/handlers_node.go:220-221` | `MatterFabricResponse` | `matter.go:232-255` | starts in `MatterFabricStore`'s `}`, ends one line early | `internal/north/rest/handlers/matter.go:234-256` |
| `fabricui/handlers_pairing.go:176-178` | `ErrBridgeTopologyNotReady` | `matter.go:131-139` | `ErrCommissioningBlocked`'s `var`, then the doc without the `var` | `:134-140` |
| `fabricui/handlers_pairing.go:18-19` | the duration bounds | `matter.go:169-180` | the body decode, then the bounds cut off | `:174-183` |
| `ui/src/lib/kit/Spinner.svelte:1` | the file | `Spinner.svelte:1-100` | the file has 99 lines | `:1-99` |

The sentences themselves hold (the bounds 180..900 and the 900 default at
`matter.go:174-178`; the confirmation word at `matter_maintenance.go:54-57`;
`Details` without `errors`/`feature` at `problem.go:79-89`).

- **Contract:** no.

### UR-P3-cite-95 — note — the toast store's comment gives loom three behaviours loom does not have

- **Where:** `ui/src/shell/toast.svelte.ts:2-4`.
- **Sentence:** "Mirrors openccu-loom
  assets/ui/src/lib/stores/toast.svelte.ts:16-59 at b40b289a: info and
  success leave after 4 s, warnings and errors stay; at most MAX_TOASTS at
  once, a repeated message once, a long one cut."
- **Cited side (M):** `loom:assets/ui/src/lib/stores/toast.svelte.ts:33-47`
  — `ttl` 4000 unless the severity is `warn` or `error` (sticky). The file
  (70 lines) has no cap, no de-duplication and no truncation; it has a
  `warn` severity and a `detail` line this store does not have
  (`toast.svelte.ts:7`: `"info" | "success" | "error"`).
- **Fix (sentence):** "Mirrors … :16-59 at b40b289a for the lifetimes
  (info and success leave after 4 s, errors stay), without loom's `warn`
  severity and `detail` line. This module's own: at most MAX_TOASTS at
  once, a repeated message once, a long one cut (UR-P1-bound-111)."
- **Contract:** no.

### UR-P3-cite-96 — note — go-fabric-server line numbers of the ping, and citations of unpinned sources without a commit

- **Where / M (local clone, `e8bcd86`, the commit `node.go:710` names;
  identical outline at v0.3.2):**
  - `fabricui/gofabric/node.go:689-690`: "bounds each attempt
    (controllerapi/cmd_node.go:187)". `:187` is the `}` of the attempts
    clamp; the bound `context.WithTimeout(ctx, 10*time.Second)` is `:192`.
  - `fabricui/gofabric/node.go:786-787`: "(cmd_node.go:190-193: err ==
    nil)". `err == nil` is `:195`; the read is `:193`. Corrected: `:193-195`.
- **No commit named** (the sentence is right at the commit its neighbours
  name; add it): `gofabric.go:458`, `:564`, `:622`, `:645`,
  `node.go:660`, `:678`, `:842`, `:853`, `handlers_controller.go:573`
  (go-fabric-server, all right at v0.3.2); `handlers_controller.go:947`
  (`ControllerCommandHandler.ts:2618-2668`, matterjs-server — right at
  `0bd06f2`: `setAclEntry` `:2624-2668`, the self-ACL filter `:2639-2652`).
- **Contract:** no.

### UR-P3-cite-97 — note — the concept pages' citations are snapshots without a pin for this repository

- **Where:** `docs/concept-u2…u6*.md` (95 map rows; 12 read).
- **M:** `docs/concept-u3-topology.md:17` `openapi.yaml:1037-1052`
  (sessions' subscription counter) is `LogLevelRequest` today; `:29`
  `openapi.yaml:297-312` (the stream's events) is `listDiagnosticEvents`;
  `docs/concept-u4-device-details.md:40` `controller/client/config.go:123`
  (`Hooks.Event`) is `SubscriptionAlive` at v0.22.2 (`Event` is `:134`);
  `docs/concept-u5-operation.md:32` `node/config.go:217-220` (`Logger`) is
  `Observers`, and its sentence "go-fabric selbst hat weder LevelVar noch
  Ring" is no longer true (UR-P3-doc-90); `concept-u3-topology.md:194`
  `im/subscription/manager.go:846` is `EventFiring`'s doc. Still exact at
  v0.22.2: `concept-u4:38` `conformance.go:93`, `:50` `schema/lookup.go:20`,
  `:199` `node/lifecycle.go:129-133`, `concept-u3:42`
  `threadnetworkdiagnostics/server_gen.go:22,28,85` and
  `wifinetworkdiagnostics/server_gen.go:21,25`.
- **Verdict:** each page names its date and go-fabric commit ("Stand …,
  go-fabric main a6ce6272"), so these are history like `docs/reports/*`,
  but the citations of this repository's own files carry no commit.
- **Fix:** one header line per concept page: "Citations are of the day
  above (this repository at `<commit>`), not maintained." No range edits.
  For unit C: the pages are German, against the repository's language rule.
- **Contract:** no.

## Fact notes on the handed-over docs rows (unit C)

1. **`docs/embedding.md:251-252`, logring** — stale; UR-P3-doc-90.
2. **`docs/contract.md:656-657`, controller `occupancy` all zero** —
   holds at v0.22.2. `gf:controller/diagnostics.go:46-62` builds
   `diag.SessionDiagnostics{AsOf: now}` and never sets `Occupancy`; its
   doc `:33-35` ("Occupancy stays zero (Capacity 0 …) and each session's
   Subscriptions 0"); `gf:diag/diag.go:57-60`, `:120-122`. The two
   neighbouring claims hold too: `last_activity` zero and `idle_seconds` 0
   (`diag.go:85-87`, `:92-93`), `subscriptions` 0 (`:78-80`). The three
   line citations of that entry are exact (`diag/diag.go:50`, `:68`,
   `:130`; `mdns/diagnose_browse.go:14`; `diagevent/ring.go:49-58`).
3. **`docs/embedding.md:288-311`, the adapter's facade table** — holds
   row by row against `fabricui/gofabric` and the facade at v0.22.2, with
   one gap:
   - `Hooks`: all 13 fields of `client.Hooks`
     (`gf:controller/client/config.go:98-161`) are chained by `Wrap`
     (`hooks.go:228-244`); events without a registry are dropped
     (`hooks.go:139-142`). Holds.
   - `Pair`: the acknowledgement is set before the template's `Edit`, as
     `gf:controller/controller.go:835-851` (`pairOptions`: options, the
     attestation default, then `opts.Edit`) runs it. Holds. The cited
     `controller/commissioner/network.go:217-235`, `:308-325` "at v0.18.0"
     are also exact at v0.22.2 (the Wi-Fi and Thread feature checks,
     `:222-226`, `:317-321`).
   - `Logger`: the three logged reasons are the three `bleFor` can give
     with a central present (`gofabric.go:306-318`, `:344-346`). Holds.
   - Routes: every named call is the adapter's (`gofabric.go:177-191`,
     `:217-256`, `:329-362`, `:459-474`, `:567-619`, `:623-659`;
     `node.go:448-479`, `:553-604`, `:630-687`, `:814-863`;
     `updates.go:26-105`; `topology.go:27-67`; `diagnostics.go:19-103`).
     `controller/topology/service.go:57-60` and `:163-165` are exact at
     v0.22.2; `controller/diagnostics.go:46`, `:547`, `:140`, `:154`
     ("at v0.21.0") are exact at v0.22.2.
   - **Gap:** the `POST /controller/commission` row does not say that the
     discovered-device path goes around `Pair`'s records
     (held UR-P1-adpt-60, UR-P3-adpt-92). The `GET /controller/discovery`
     row could add that the handler always passes a timeout (1–60 s,
     default 3 s, `handlers_controller.go:1412-1418`), since the facade's
     0 means 15 minutes (`gf:controller/code.go:92-99`).

## Fact notes on the carried rows

### UR-P1-adpt-40 — the fabric label's limit: three different units, bytes decide on the device

- **This repo:** `handlers_controller.go:1537`
  `utf8.RuneCountInString(label) > 32` (runes).
- **go-fabric controller:** `gf:controller/commissioner/fabric.go:98-101`,
  `:123-126` — `MaxLabelLength = 32` "in characters",
  `utf8.RuneCountInString` (runes).
- **matter.js (local clone, `862ad75f`):**
  `packages/protocol/src/fabric/Fabric.ts:194-196` `label.length > 32` —
  UTF-16 code units, neither runes nor bytes.
- **The device side:** go-fabric's own node refuses in bytes —
  `gf:cluster/core/operational_credentials.go:2228` `len(req.Label) > 32`,
  answered with IM InvalidCommand (its comment `:2229-2232` cites chip);
  connectedhomeip (local clone) `src/credentials/FabricTable.h:48`
  `kFabricLabelMaxLengthInBytes = 32`, checked at `FabricTable.cpp:237`,
  `:2142`. The generated struct decoder counts UTF-16 units
  (`gf:cluster/spec/codec.go:199-216`), so the byte check is the
  handler's.
- **Verdict:** the claim is right. A label of 17 "ü" (17 runes, 34 bytes)
  passes this handler and `Fabric.SetLabel`, is persisted as the
  controller's label, and every go-fabric or chip node refuses
  UpdateFabricLabel in the background push — visible only as
  `fabricui.gofabric.label_push_failed`. The Matter data model's string
  length is in octets (Core §7.19.2 "String", from memory — not read at a
  pin; the two device implementations above are what was read).
- **Decision for the fix wave:** `len(label) > 32` (bytes) in the handler
  is the only limit every device accepts and is stricter than go-fabric's,
  so no label passes here that go-fabric refuses. The comment at
  `handlers_controller.go:1528-1530` then reads: "the longest fabric label
  in bytes — what a node accepts (go-fabric
  `cluster/core/operational_credentials.go:2228`); go-fabric's controller
  counts runes (`commissioner.MaxLabelLength`), matter.js UTF-16 units
  (`Fabric.ts:194`)". An upstream question to go-fabric (runes on the
  controller, bytes on the node) belongs with it. Contract: the schema's
  `maxLength: 32` (`openapi.yaml:3270`) counts characters in JSON Schema —
  a description line, no shape change. Severity stays note.

### UR-P1-ui-178 — trivially invalid passcodes

- **Decides:** `gf:secure/setup/setup.go:330-346` (`trivialPINs`:
  00000000, 11111111 … 99999999, 12345678, 87654321) and `:348-361`
  (`IsValidSetupPIN`: 1..99999998 and not in the list), mirroring chip
  `IsValidSetupPIN` and matter.js `INVALID_PASSCODES`
  (`gf:secure/setup/decode.go:91-95`). Matter Core §5.1.7.1 "Invalid
  Passcodes" is the specification's list (from memory, not read at a pin).
- **What the facade does:** a pairing *code* is validated —
  `gf:controller/code.go:53-58` (`ParsePairingCode` … "each validating the
  passcode"), so `Pair` refuses such a code with `ErrPairingCode` (400
  here). A *passcode* for a discovered device is not:
  `commissioner.Options.Passcode` goes to PASE unchecked (no caller of
  `IsValidSetupPIN` under `gf:controller/`), and the adapter adds no check
  (`gofabric.go:596-598`).
- **Verdict:** the claim is right and harmless: no conforming device has
  such a passcode, so the commissioning fails at PASE after the
  discovery, with a 502 in place of a 400. The fix belongs in the handler
  as well as the page (the list above, 400 `passcode`), since the page is
  not the only client; `Number()` accepting `0x10`/`1e3` is the page's
  alone. The contract's `1..99999998` stays; an added refusal of eleven
  values is no shape change. Severity stays note.

### UR-P1-ui-179 — ACL auth mode `pase`

- **Decides:** `gf:cluster/core/access_control.go:699-700` (rule 4:
  "AuthMode != PASE on every entry — PASE auth is only valid during
  commissioning, never in the persisted ACL", after chip
  `AccessControl.cpp` `Entry::IsValid` and matter.js
  `AccessControlServer.ts:165-265`) and `:749-750`: a write with
  `AuthMode=PASE` is answered CONSTRAINT_ERROR. The enum itself keeps the
  value (`gf:cluster/spec/accesscontrol/definition_gen.go:92`), so a read
  can in principle show it.
- **Verdict:** the claim is right: an entry with `pase` can be read, never
  written; every conforming node refuses it. Remove `pase` from the add
  form's `MODES` (`Acl.svelte:34`, `:112`) and keep the label for display;
  the contract's enum (`openapi.yaml:3115`) stays, because the read side
  shares the schema. The handler's comment at
  `handlers_controller.go:942-944` ("a PASE entry's [subject is] a
  passcode id") stays true for display. Severity stays note. Contract: no.

## (d) The adapter against the controller facade at v0.22.2 — read without finding

Besides UR-P3-adpt-91 and -92 and the held rows:

- **Zero values and defaults.** `Controller.Discover` timeout 0 = 15 min
  (`code.go:92-99`): the adapter never passes 0 (handler default 3 s,
  `rediscoverWindow` 5 s). `OpenCommissioningWindow` timeout 0 =
  `DefaultWindowTimeout`, zero `WindowOptions` = enhanced
  (`window.go:46-55`, `:74-97`): as the adapter's comment says.
  `WriteOptions`/`InvokeOptions` `TimedTimeout` 0 = default timed timeout,
  a timeout marks the interaction timed (`client/interaction/write.go:49-54`,
  `invoke.go:31-42`): as `node.go:583`, `:640` use them.
  `ICDRegisterOptions` zero = own node id, Permanent, the trusted vendors
  (`icd.go:128-140`). `CommissionOptions.AutoSubscribe` nil = subscribe
  (`fabricops.go:124-126`). `ForDevice(…, "")` = an allocated id
  (`addresses.go:150-152`).
- **Error types.** `ErrPairingCode`, `ErrConfiguration`, `ErrNoDevice`,
  `*TermsRequiredError` (`controller.go:754-768`, `errors.go:158-176`),
  `ErrLabel` (`fabric.go:103-104`, `:123-126`), `ErrMismatch`
  (`softwareupdate/manager.go:55-57`, `:1236-1240`), `*StatusError`,
  `ErrJSONValue`, `ErrNoSession`/`ErrNotCommissioned`/`ErrClosed` of
  `PingAddress` (`ping.go:44-48`) are mapped as the facade defines them.
  `commissionedDespite` matches `Registry.Commission`'s "an error of its
  initial read is returned and the node stays commissioned"
  (`fabricops.go:156-157`, `:230-234`).
- **Context use.** The label push runs on the source's context, not the
  request's, as go-fabric-server's does (`cmd_server.go:256`); `Ping`
  bounds the round with 10 s and returns the request's own cancellation;
  `Discover`'s end of time is no error (`code.go:104-105`).
- **Callbacks.** `Manager.Observe`'s fn may call back but not `Close`
  (`observe.go:131-138`): the adapter's only publishes.
  `Controller.OnDiagnosticEvent` runs inline and must not call
  `SessionDiagnostics`/`DiagnoseMDNS` (`diagnostics.go:144-153`): the
  adapter's fn converts and hands over; the two calls are made from the
  request only (`diagnostics.go:19-21`, `:76-83` of the adapter). Hooks
  run after the node's lock is released (`config.go:88-97`), so
  `Registry.Get` and `PeerAddress` in `Hooks.nodeID` are allowed; the
  `Decommissioned` hook starts no fabric operation (`config.go:140-148`).
- **Ownership.** `Registry.Commissioned` returns a fresh slice
  (`registry.go:590-605`); `DiagnosticEvents` "a copy the caller may keep"
  (`diagnostics.go:133-141`, `diagevent/ring.go:210-231`); `EventHistory`
  "a copy the caller may keep" (`event_history.go:100-107`); the adapter
  copies `DeviceTypes`, `Features`, ACL subjects before handing them on.
  `diag.PeerInfo`'s `SessionIDs`/`DiscoveredNames` are passed through
  from a snapshot built per call (`controller/diagnostics.go:46-62`).
- **Nil accessors.** `Controller.Fabric()` and `Registry()` return fields
  that are never nil on a built controller (`controller.go:639-648`); the
  adapter's nil checks are dead but harmless. `SoftwareUpdates()` is nil
  without `Config.OTA` (`ota.go:125-129`), `Node.ICD()` nil for a node no
  registry holds (`icd.go:285-286`): both handled.

## Covered without finding

Map rows are named by the citing line. "Range and construct" = the cited
lines exist and hold the named construct; "sentence" = the claim was read
against the lines.

- **go-fabric, sentence read at v0.22.2 (exact or within two lines):**
  `ui/src/lib/kit/im-status.ts:2`; `Endpoints.svelte:62-63` (both);
  `support_test.go:498`, `:608`; `options.go:50`, `:54`, `:61`, `:68`
  (both), `:77`, `:88`, `:235`, `:238`; `nodeswitch.go:23`, `:25`;
  `node_detail.go:42`, `:49`, `:52`, `:111`, `:171`, `:311`;
  `memory_controller.go:284`; `handlers_surface_test.go:47`;
  `handlers_surface.go:13`; `handlers_node.go:114`, `:116`, `:376`, `:503`,
  `:573`, `:610`; `handlers_diagnostics.go:76`, `:92`, `:171`;
  `handlers_controller_test.go:715`; `handlers_controller_review_test.go:35`;
  `handlers_controller.go:540`; `events.go:251`, `:253` (both), `:255`;
  `e2e_test.go:454` (both); `controller.go:744`, `:885`, `:931`;
  `gofabric/updates.go:62`, `:77`; `gofabric/topology.go:218`, `:226`;
  `gofabric/node.go:416`, `:589`; `gofabric/gofabric_test.go:189`;
  `gofabric/gofabric.go:301`, `:368`, `:400`, `:415`, `:772`;
  `gofabric/e2e_test.go:863` (`ahead.go:47`), `:915`, `:1085`, `:1132`;
  `gofabric/diagnostics.go:87`; `internal/arch/arch_test.go:38`, `:77`;
  `docs/embedding.md:290`, `:310`, `:311`; `docs/contract.md:650`, `:667`,
  `:677`.
- **go-fabric, right at the tag the sentence names (the construct is
  unchanged at v0.22.2, at other lines):** `memory_controller.go:155`
  (v0.18.0 `controller.go:623-631`; now `:732-740`); `nodeswitch.go:24`
  (v0.18.0 `node/node.go:1119-1121`, `FabricInfos`);
  `gofabric/gofabric.go:303` (v0.18.0 `:674-676`; now `:807-809`), `:366`
  (v0.18.0 `:611-618`; now `:720-727`), `:399` (v0.19.1 `:654-663`; now
  `:710-719` — the map resolved this one to this repository by mistake),
  `:402` (v0.19.1 `:752-768`; now `:835-851`); `gofabric/e2e_test.go:863`
  (v0.18.0 `node/config.go:402-411`; now `:434-443`), `:1049` (v0.18.0
  `contract/wifi.go:17-25`; now `:21-25`).
- **openccu-loom, range and construct (whole-file ports checked against
  the file length, handler citations against the function or type at the
  range):** `toast.svelte.ts:2` (range; sentence is UR-P3-cite-95),
  `store.test.ts:3`, `store.svelte.ts:3`, `:328`, `qr.ts:3`,
  `theme/ha-bridge.ts:6`, `Status.svelte:3`, `Selection.svelte:5`,
  `Pair.svelte:3`, `Fabrics.svelte:3`, `Endpoints.svelte:3`,
  `Diagnostics.svelte:4`, the kit's 17 other files (`utils.ts`, `tabs.ts`,
  `icons.ts`, `data-table.ts`, `Tabs`, `Select`, `PageShell`, `PageHeader`,
  `LoadingState`, `Input`, `Icon`, `ErrorState`, `EmptyState`,
  `DataTable`, `Card`, `Button`, `Badge`), `api/client.ts:5`;
  `stream.go:26`, `spa.go:30`, `ports.go:116`, `handlers_pairing.go:41`,
  `:60`, `:149`, `:150`, `:246`, `:273`; `handlers_node.go:241`, `:273`,
  `:296`, `:326`; `handlers_exposure.go:115`; `handlers_diagnostics.go:20`,
  `:111`, `:131`, `:143`, `:151`, `:160`, `:193`, `:226`, `:227`, `:309`,
  `:318`, `:334`; `handler.go:303`, `:608`; `events.go:113`;
  `problem.go:112`, `:134`, `:142`, `:151`. The ten "ambiguous" `matter.go`
  and `write.go` rows all mean `internal/north/rest/handlers/…`; the map
  resolved seven loom `matter.go` rows to `gf:contract/matter.go`, which
  is the script's error, not the comment's.
- **openccu-loom, sentence read:** `Diagnostics.svelte:122` (five minutes,
  `MatterDiagnostics.svelte:64-71`); `spa.go:73` (a year / an hour with
  revalidation, `spa.go:80-85`); `handlers_node.go:137` (200 with
  `enabled:false`, `matter_exposures.go:363-371`);
  `handlers_exposure.go:187` (empty batch 204 at `:268-270`; the failed
  reassembly logged, `:231-239`); `handlers_diagnostics.go:59`
  (`matter.go:333-346`), `:378` (503 without a node, `matter.go:580-585`);
  `handlers_pairing.go:60` (503 without a passcode, `matter.go:53-58`);
  `i18n/index.ts:4` (`i18n.ts:36`, `:8688-8719`: format, the fallback
  chain, `catalogKeys`).
- **go-fabric-server (local clone), sentence read at the named commit:**
  `gofabric.go:245`, `:325` (the four inputs, `cmd_commission.go:132-135`
  at `b947c93`), `:564` (no cleanup of the node after a failed
  commissioning there either, `:222-232`), `:645`; `node.go:446`, `:549`,
  `:550`, `:628`, `:710`, `:812` (`cmd_icd.go:48-73`).
- **matterjs-server (local clone, `0bd06f2`):** `gofabric/topology.go:225`
  (`NetworkTopologyService.ts:452-478`, `mapExternal` with
  `stripMdnsHostname`). go-fabric's own comment names the helper's
  definition instead (`gf:controller/topology/build.go:92-94`,
  `topology-utils.ts:1114-1116`); both are right.
- **matter.js (local clone, `862ad75f`):** `handlers_controller.go:1529`
  (`Fabric.ts:193-196`; see UR-P1-adpt-40).
- **This repository:** `nodeswitch.go:29` (`events.go:152-166`), `:35`
  (`registry.go:168-190`).
- **Not read:** `NodeManage.svelte:9` and the `docs/concept-u6-nodes.md`
  rows on home-assistant-matter-hub and go-ha2fabric (no pin, other
  repositories' docs); `docs/reports/*` (historical); the concept pages
  beyond the 12 of UR-P3-cite-97; `gofabric/e2e_test.go:1050`
  (`controller/ble_test.go:24-38`, a test pattern).

## Counts

- Map rows in scope (code, tests, `docs/embedding.md`, `docs/contract.md`):
  185; read: 183 (2 not read, above). Concept pages: 12 of 95 sampled.
- Without finding: 161. With a finding: 22 rows in 5 citation findings
  (UR-P3-cite-93: 5 rows, -94: 7, -95: 1, -96: 2 wrong lines + 10 without
  a commit; -97: 5 of the 12 sampled concept rows, counted apart).
- Findings: 8 — low 2 (UR-P3-doc-90, UR-P3-adpt-91), note 6 (one a held
  candidate, UR-P3-adpt-92). High and medium: none.
- Carried rows judged: 3 (all three claims confirmed, severities kept).
- Red test: 1 (UR-P3-adpt-91), run in a scratch copy, FAIL as stated; the
  review worktree is untouched but for this file.
