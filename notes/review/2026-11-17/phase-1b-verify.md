# Phase 1b — verification

Independent verifier (Fable), branch `review/p1b` at `6d3d133` (source as
`007ca96`), go-fabric `v0.22.2` read from the tag (`gf:` below). Scope: every
high and medium of the three web reports, seven lows as the sample
(decision 17), two severities S-web flagged as arguable, and G-web A's page
fix for UR-P1-err-04. **M** = run as a scratch probe outside the repo, or
read at a line that alone decides; **I** = inferred (browser or device
behaviour not executed). Nothing in the repo was changed but this file;
`git status --short` is empty for tracked files.

Probes (all outside the repo, vitest with `root: ui/`):
`…/scratchpad/verify/` (`a.test.ts.done` V1–V4, `b.test.ts` C1–C2,
`c.test.ts` C3, output `out.txt`), the reviewers' `…/scratchpad/vt/`
(`attr.test.ts`, `acl.test.ts`) and `…/scratchpad/p1b/probe.test.ts`
re-run; a Go program on the pinned module in `…/scratchpad/verify-go/`
(`client.EncodeJSON` at `v0.22.2`).

## Verdicts

| ID | Verdict | Severity (reviewer → mine) | Reason |
| --- | --- | --- | --- |
| UR-P1-ctr-101 + UR-P1-ctr-162 | CONFIRMED, **one finding; 101 survives** (root line `client.ts:305` is S-web's), 162 merged into it with its lockout analysis | medium / high → **high** | **M**: a read of `0x001F/0x0000` with a 64-bit subject shows, seeds and writes back the rounded id with a success toast (V1); go-fabric accepts the rounded value (Go probe). The lockout is not an edge case: the controller's own node id is a random 64-bit value (`gf:controller/controller.go:347`), above 2^53 in ~99.95 % of installs; the ACL reaches the generic path with no guard (`fabricui/gofabric/node.go:556-561`, `handlers_controller.go:768`) |
| UR-P1-ui-161 | CONFIRMED | medium → **low** | **M** (`attr.test.ts` red as described). The write goes to the path the inputs show; the defect is the stale result card under a changed path. Needs an operator slip, admin only, and the typed encoder refuses a value of another type. Medium is defensible |
| UR-P1-ui-163 | CONFIRMED | medium → **low** | **M** (`acl.test.ts` case 1 red: `PUT …/0000000000000012/acl` carries node …11's entry). No host is affected today: the shell keys (`shell/App.svelte:259-261`), and the one known npm host (openccu-loom) mounts node pages only. A defect of the exported component, not of a shipped path |
| UR-P1-conc-105 + UR-P1-conc-160 | CONFIRMED, **one finding; 105 survives** (root is `startScan`'s missing cleanup and abort, `scan/index.ts:49-57`), 160 merged | low / medium → **medium** | **M**: Cancel while starting, then `startScan` resolves → `stop` never called, a late code still reaches `onCode` with a success toast; a second scan leaks the first for good (C1); destroyed while starting → `stop` never called, `onCode` runs on the destroyed page (C2); `play()` rejects → tracks not stopped (C3). Browser side **I** |
| UR-P1-ui-130 | CONFIRMED | medium → **medium** | **M** (P1 re-run: `each_key_duplicate … interface_not_found at indexes 0 and 1`). Reachable at the pin on the node page: `gf:node/mdns.go:118-124` appends `gf:mdns/diagnose_interfaces.go:45-74`, one finding per named interface, no `Service`. Svelte throws in production too (`svelte@5.57.3 …/blocks/each.js:357-363`). The controller page has the same key (`controller/Diagnostics.svelte:198`) but go-fabric's controller does not produce the pair (`gf:controller/diagnostics.go:543`): latent there |
| UR-P1-ui-104 (S-web sample) | CONFIRMED | low → low | **M** (V3): after the first confirm the same button element stays connected and shows the second question; a second click on it answers true. Double-click timing in a browser **I** |
| UR-P1-ui-112 (S-web sample) | CONFIRMED | low → low | **M** (V2): `routeOf("#/controller/__proto__")` gives page `__proto__`, `in` is true; rendering it throws `TypeError: … is not a function`, `valueOf` throws too, `constructor` and `toString` render an empty page. Note would also be defensible |
| UR-P1-web-114 (S-web sample) | CONFIRMED | low → low | **M** (V4): `fabric-ui.table:controller-nodes = {"query":5}` → `TypeError: … .trim is not a function` in `Nodes.svelte`, no rows, and the bad value stays stored. Reachable because `controller/Nodes.svelte:131-132` sets `persistKey` and `search` |
| UR-P1-ui-131 (G-web A sample) | CONFIRMED | low → low | **M** (P3 re-run: `duplicate key 12 at indexes 0 and 1`). A race of one round trip; the page breaks until reload |
| UR-P1-ui-133 (G-web A sample) | CONFIRMED | low → low | **M** (read): no admin test in `Pair.svelte`, `Fabrics.svelte`, `Selection.svelte`, `Operation.svelte` (grep: only `Diagnostics.svelte:26`, `NodeManage.svelte:27`, `shell/App.svelte:107`); `GET /node/pairing` is `x-admin` (`openapi.yaml:86`); `threat-model.md:165` says otherwise |
| UR-P1-ui-164 (G-web B sample) | CONFIRMED | low → low | **M** (`acl.test.ts` case 2 red: `targets: [{endpoint: null}]`). The report's lines are off by 100: `Acl.svelte:62`, not `:162`; `Bindings.svelte:56`, `:59` hold. Harm is small: the server refuses `1.5`, `-1`, `70000`; `null` becomes an empty target |
| UR-P1-ui-172 (G-web B sample) | CONFIRMED | low → low | **M** (recomputed): warning `#ffa600` 1.96, success `#43a047` 3.30, primary `#009ac7` 3.26, error `#db4437` 4.29 on white. Tokens at `tokens.css:93`, `:101-103` (not `:102-104`); text uses at `controller/Diagnostics.svelte:195`, `node/Diagnostics.svelte:179`, not the cited `:281-283`, `:355` (the file has 265 lines) |
| UR-P1-bound-103 (severity view) | CONFIRMED | low → **medium** | **M** for the counts (`shell/App.svelte:192-197`, `NodeScope.svelte:33`); the limit **I**. The daemon's own listener without TLS is HTTP/1.1 always, so the module's direct mode has it: three tabs of a daemon with both families hang every call, without an error. Under an Ingress the pool is the host frontend's too |
| UR-P1-err-107 (severity view) | CONFIRMED | low → low | **M** (read: `controller-store.svelte.ts:75-88`, `shell/App.svelte:78`, `:186-188`, `:239-240`). A reload recovers; no data or rights involved. Fix together with UR-P1-conc-102 |
| UR-P1-err-04, page fix (G-web A) | SOUND as an interim, **with three corrections** (below); not complete without the server | — | The rule hides codes that do work when the host enables go-fabric's basic-commissioning feature; two reads are not atomic; the new phase must be handled in three more places |

Struck: none. False positives in the sample of seven lows: none. Of the
six highs and mediums (four after merging): none struck, two lowered
(161, 163), one pair raised to its higher reading (101/162), one low
raised (103).

## Highs and mediums

### UR-P1-ctr-101 / -162 — confirmed, one finding, high

Lines read: `ui/src/lib/api/client.ts:300` (`JSON.stringify` of the body),
`:305` (`JSON.parse` of the answer), `:396-398`;
`controller/node/Attributes.svelte:45-47`, `:60`, `:67`, `:92`;
`fabricui/handlers_controller.go:726` (`fabric_filtered` defaults to
true), `:749-771` (`WriteNodeAttribute`, any path, the raw value passed
on); `fabricui/gofabric/node.go:448-479` (read, `NameBased`), `:553-561`
(a raw value goes to `client.EncodeJSON`, no exclusion for the ACL).
go-fabric: `gf:controller/client/json.go:34` is the doc sentence the
reviewer cites ("64-bit ones in full as plain decimals"); the code is
`gf:internal/valuejson/valuejson.go:181-184` (S-web's line). Both hold.

Probe V1: the attributes tab on a fetch that answers raw text with subject
`18364758544493064721`. The card shows `18364758544493064000`, the edit
field holds it, and an unchanged Write sends `…"subjects":[18364758544493064000]…`
with `controller.attributes.written`. Go probe: `client.EncodeJSON(AttributeField(0x1F, 0), …)`
accepts that value without an error; only a rounding past 2^64 is refused.

Is the lockout realistic? Yes, more than either report says.

- The ACL does reach this path. The cluster select offers `0x001F` on
  endpoint 0 (V1 lists it), and the attribute field's default `0x0000`
  **is** the ACL. Select the cluster, Read, Write.
- The read is fabric-filtered, so the list is our fabric's entries: the
  Administer entry go-fabric wrote at commissioning, subject = the
  controller's root node id (`gf:controller/commissioner/flow.go:1020`).
  G-web B's "every other administrator's subject" is imprecise: other
  fabrics' entries are neither read nor written.
- That id is `random()%0xFFFF_FFEF_FFFF_FFFF + 1`
  (`gf:controller/controller.go:347`): at most 2^53 in about one install
  of 2000. The write replaces our only Administer subject with an id
  nobody holds. Whether a device refuses an ACL without the writer is
  **I**; the specification does not require it.
- The safe path exists and is not affected: the ACL tab and
  `PUT …/acl` carry `subjects_hex` (`Acl.svelte:39`, `:66`,
  `handlers_controller.go:944-1001`), with the lockout question at
  `Acl.svelte:80-90` (not `:180`). Nothing guards the generic writer.

High: an administrator's write that looks idempotent removes the
controller's access to the device, with a success toast, on nearly every
install. Display alone (the inventory's `AttributeTable.svelte:16`, the
events) also breaks `docs/contract.md:192-195`.

### UR-P1-ui-161 — confirmed, low

`Attributes.svelte:30-35`, `:55-67`, `:82-84`, `:87-96`. `attr.test.ts`
re-run: red, `attribute_hex: "0x0000"` in the PUT after a read of
`0x0005`. The endpoint select keeps `cluster` (`:82`). The inputs are the
only path on screen (the card names none), so the write goes where the
page says; what is wrong is that a value read elsewhere is still offered.

### UR-P1-ui-163 — confirmed, low

`controller/Node.svelte:41-48` (`load`, no sequence), `:50-53` (reload on
`nodeId`), `:89-101` (tabs); `node/Acl.svelte:45`, `node/Bindings.svelte:34`
(`onMount` only). The report's `Node.svelte:142`, `:157-164`, `:166-169`,
`:190-218` do not exist (105 lines). `acl.test.ts` case 1 re-run: red as
described. `ui/README.md:55` documents the prop, and `Node.svelte` itself
reacts to a new `nodeId`, so the in-place change is meant to work. But
`docs/embedding.md:334-345` asks nothing about keying, the shell keys, and
openccu-loom's `Matter.svelte` imports no controller page.

### UR-P1-conc-105 / -160 — confirmed, one finding, medium

`scan/index.ts:49-53`, `:57-62`; `QrScan.svelte:23-43`, `:45-51`, `:62`.
C1–C3 as in the table. Beyond the reports: removing the `<video>` on
Cancel pauses it, which rejects a pending `play()` (**I**, HTML media
element removal steps) — so Cancel after the permission was granted takes
the play-rejection path of 105 and also toasts `scan_failed`. Medium: the
camera runs with nothing in the page showing it, until the tab closes,
after an ordinary Cancel or tab switch during the prompt.

### UR-P1-ui-130 — confirmed, medium

`node/Diagnostics.svelte:176-183`; `controller/Diagnostics.svelte:198`
(same expression). `gf:mdns/diagnose_interfaces.go:45-74`: three codes,
each once per name of the allow-list, `Service` empty;
`gf:node/mdns.go:118-124` appends them when the node owns its advertiser
and `MDNSEnabled` (which is `advertising`, `fabricui/ports.go:120`).
Two typos, or `docker0` and a `br-…` bridge, are enough. P1 re-run.
`gf:mdns/diagnose.go:105-190`: the other codes carry the service name.

## The sample of lows, and the two severities

One paragraph each is in the table; the probe names are there. Two
additions:

- **UR-P1-web-114** contradicts an aside in G-web A's UR-P1-ui-142
  ("`DataTable`'s own search … is used by no page"):
  `controller/Nodes.svelte:131-132` uses it, persisted. So 142's regex is
  also replayed from storage on that page.
- **UR-P1-bound-103**: medium rests on the direct listener being
  HTTP/1.1; if the coordinator weighs only proxied hosts, low stands.

## Fix notes

**UR-P1-ctr-101 (with 162).** Client-side lossless parse, **no contract
change**; not the string wire form. Reasons: go-fabric accepts a decimal
string only for a *typed unsigned* field (`gf:cluster/spec/json.go:262-281`,
Go probe: accepted); a signed 64-bit wants a JSON number (`:158-165`), and
on an untyped path a string is encoded as a UTF-8 string (`:420`). A
string form would need a go-fabric change and stay ambiguous. Complete
means all of:

1. the answer's `value` (attributes, event `data`, invoke `response`)
   keeps integers outside ±2^53 exactly — as source text or `bigint`;
2. display (`Attributes.svelte:92`, `kit/AttributeTable.svelte:16`,
   `EventList`) prints them exactly;
3. the write sends the edit field's text verbatim after checking that it
   is JSON — both `Attributes.svelte:60` and `client.ts:300` round today,
   so fixing the parse alone leaves a hand-typed id rounded;
   `invokeCommand` (`client.ts:399`) has the same body path for hosts;
4. `JSON.parse`'s `context.source` and `JSON.rawJSON` are not in every
   webview a host frames: a small own scanner, or a fallback, is needed;
5. `ui/src/test/support.ts:46` parses the sent body: the red test must
   compare the raw body text, so `Sent` needs it;
6. `docs/contract.md:192-195` says what a JS client must do with `value`
   (docs, not wire).

The stopgap both reviewers name is right and cheap: refuse a write whose
parsed value holds an integer that is not `Number.isSafeInteger`.

**UR-P1-ui-161.** Clear `result` when endpoint, cluster or attribute
changes, and reset `cluster` on an endpoint change. Do **not** take the
first half of the proposal alone (write to the path of `result` while the
inputs show another): that writes to a path nothing on screen names.
No contract change.

**UR-P1-ui-163.** `{#key nodeId}` alone is incomplete: the remounted tabs
get the old `detail` until the new one lands (`Attributes.svelte:18`,
`Bindings`). In the `nodeId` effect set `detail = null`, `error = null`,
`focus = undefined` and guard `load()` with a sequence; `detail = null`
already unmounts the tabs. No contract change.

**UR-P1-conc-105 (with 160).** Both sites: in `startScan` a
`try/catch` from `getUserMedia` on that stops the tracks, and an
`AbortSignal` (or a returned handle before the awaits); in `QrScan` a
generation set by `cancel()`, checked after the await, no toast and no
`onCode` when cancelled. A fix only in `QrScan` leaves the `play()`
rejection. No contract change.

**UR-P1-ui-130.** Index keys on both pages are right and complete for
these static lists. No contract change.

**Lows.** 104: ignoring clicks for a moment after `current` changes
covers the shell only; a host's own `confirm` is the host's. 112:
`Object.hasOwn` at `App.svelte:152` and `NodeFamily.svelte:22` is enough
(`:207` and `:23` read through the checked name). 114: as proposed; the
header comment (`DataTable.svelte:2`) must follow if `query` is no longer
kept. 131, 133, 164, 172: as proposed. 103: closing on `hidden` needs
102's resync first. None changes the contract.

**UR-P1-err-04, page fix.** Sound as a fail-safe interim, with these
corrections:

1. *It can hide codes that work.* go-fabric names every administrator
   window `enhanced` (`gf:node/commissioning.go:364-366`, `!snap.Own`),
   also the basic one of the BC feature (`gf:node/config.go:424-432`,
   off by default), which runs on the configured passcode and
   discriminator (`gf:node/commissioning.go:182-190`). With
   `window_kind: enhanced` and `source: configured` the page cannot tell
   the two apart. The notice must therefore not say the codes "are not
   known"; say they may not open this window. The exact answer needs the
   server (`pairing` naming whether its codes fit the open window, from
   `WindowSnapshot.Administrator`): a contract change, phase 3.
2. *Two reads.* `hydrateCommissioning` uses the cached status
   (`store.svelte.ts:256-260`) and a fresh pairing. Read the status fresh
   when it says a window is open, or re-check after the pairing answer.
   Test `window_kind === "basic" || "enhanced"` literally, so `none` and
   an absent field keep today's path.
3. *The new phase.* `window_closed` ends only phase `open`
   (`store.svelte.ts:333-336`), `hydrateCommissioning` returns unless
   `idle` (`:255`), and `Pair.svelte:99-178` branches on three phases: a
   `foreign` phase needs all three, or it never leaves. `closeWindow`
   (`:282-286`) already fits.

The first rule (`if (!p.window_open) return`) is right as it stands.

## New while verifying

- **New, report defect (G-web B):** many `file:line` citations point past
  the end of the file — `controller/Node.svelte` (105 lines; cited `:142`
  to `:218`), `node/Acl.svelte` (cited 100 too high throughout),
  `node/Window.svelte` (80; `:130`, `:173-193`), `node/Icd.svelte` (92;
  `:224`, `:274-280`), `node/Overview.svelte` (120; `:235-239`),
  `node/Updates.svelte` (75; `:102-138`), `node/Events.svelte` (68;
  `:218-220`), `controller/Network.svelte` (112; `:218-279`),
  `controller/Diagnostics.svelte` (265; `:281-283`, `:355`, `:404-405`,
  `:416`). The claims I checked hold at other lines; every G-web B line
  must be re-derived before the public copy.
- **New, low (I), upstream:** `gf:node/commissioning.go:364-366` reports
  an administrator's *basic* window as `WindowEnhanced`, against the
  kind's own doc (`gf:node/observers.go:225-228`, "with its PAKE
  verifier"). `fabricui/handlers_node.go:192` passes it on as
  `window_kind`. Only with `BasicCommissioning` on. A go-fabric issue.
- **New, note (M, read):** `controller/node/Attributes.svelte:46` — a
  read the node answers with neither value nor status
  (`fabricui/gofabric/node.go:472-477` leaves the path out) sets
  `result = null`: the card disappears without a word.
- **New, note:** G-web A's UR-P1-ui-142 misstates that no page uses the
  persisted `DataTable` search (`controller/Nodes.svelte:131-132` does).
