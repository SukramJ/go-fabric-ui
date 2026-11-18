# Phase 1b — G-web B (controller pages, tokens, catalogues)

Reviewer: Opus. HEAD `007ca96`, branch `review/p1b`. Scope: `ui/src/lib/pages/controller/**`,
`pages/index.ts`, `lib/index.ts`, `theme/tokens.css`, `shell/app.css`, their vitest files,
`i18n/en.json` / `de.json` by key. IDs `UR-P1-<theme>-160…182`.

Scratch tests (outside the repo): `/tmp/claude-1000/-home-markus-Dokumente-GitHub/fc0f62c8-1d8c-48ba-8069-435b70bf4869/scratchpad/vt/`
(`attr.test.ts`, `acl.test.ts`, run with `vitest.scratch.config.ts` there, root `ui/`).
The `file:line` citations of this report are often off (the verifier,
`phase-1b-verify.md`, "New while verifying"); the corrected lines are in
`phase-1b-fix-W2.md`.

Counts: 1 high, 3 medium, 10 low, 9 note.

---

## UR-P1-ctr-162 — high — a 64-bit attribute value travels as a JS number (display and write-back)

- **Where:** `ui/src/lib/api/client.ts:305` (`JSON.parse(text)`), `ui/src/lib/pages/controller/node/Attributes.svelte:47`, `:60`, `:92`; contract `contract/openapi.yaml:3062` (`AttributeValue.value`, raw JSON), `:3077` (`WriteRequest.value`).
- **Claim:** go-fabric renders 64-bit integers "in full as plain decimals" (go-fabric v0.22.2 `controller/client/json.go:34`), and the adapter passes them through (`fabricui/gofabric/node.go:462`). The client parses the answer with `JSON.parse`, so any value above 2^53 is rounded before the page sees it: `JSON.parse('[1311768467463790321]')` gives `1311768467463790300` (0x123456789ABCDEF1 → 0x123456789ABCDF00). The page shows the rounded number (`:92`), seeds the edit field with it (`:47`), and a write sends it back (`:60`, `JSON.parse(edit)`). Attributes that carry node or fabric ids are exactly the ones an administrator reads here: AccessControl ACL (0x001F/0x0000) subjects, OperationalCredentials Fabrics (0x003E/0x0001) NodeID/FabricID, Binding (0x001E/0x0000) node. Reading the ACL, editing one field and writing it back rewrites every other administrator's subject node id to a different id — and, if the controller's own id is above 2^53, its own Administer entry, which locks it out. The ACL tab's lockout guard (`Acl.svelte:180`) does not apply to this path. This breaks the rule "a JS number never carries a node or fabric id". The inventory's `ClusterView` (kit, G-web A's unit) displays the same parsed values.
- **M/I:** M (read; the rounding run with node). The lockout consequence is I.
- **Fix:** parse attribute/response `value` losslessly: either the contract renders integers outside ±2^53 as strings in `value` (contract change; go-fabric's EncodeJSON must accept the string form back), or the client parses `value` with a reviver using `JSON.parse`'s `context.source` (or a small lossless parser) and keeps big integers as `bigint`/string, serialising them back verbatim on write. Until then the writer should refuse to send a value that contains an integer above `Number.MAX_SAFE_INTEGER` it did not get verbatim.
- **Contract change:** yes for the string form; no for a lossless client parse.
- **Red test:** answer a read of 0x001F/0x0000 with `value` `[{"1":5,"2":2,"3":[1311768467463790321],"4":null,"254":1}]` as raw text; read, click write unchanged; expect the PUT body text to contain `1311768467463790321` (today `1311768467463790300`).

## UR-P1-ui-161 — medium — the attribute writer writes to the path in the inputs, not to the path it read

- **Where:** `ui/src/lib/pages/controller/node/Attributes.svelte:55-67` (`write()` calls `path()`), `:30-35`, `:82-84`, `:87-96`.
- **Claim:** after a read, the result card and its write field stay while the endpoint select, cluster select and attribute input remain editable. `write()` builds its path from the current inputs, so a value read from one attribute is written to another the operator only typed into the attribute field (or a cluster of another endpoint: changing the endpoint keeps `cluster`, which then may not exist there). Measured: read 0x0028/0x0005 (`"Kitchen"`), change the attribute input to `0`, click write → `PUT …/attributes` body `attribute_hex: "0x0000"`, value `"Kitchen"`.
- **M/I:** M (scratch `attr.test.ts`, fails as described).
- **Fix:** keep the path the read used with the result (`result` already has `endpoint`, `cluster_hex`, `attribute_hex`) and write to that; clear `result` when any input changes; reset `cluster` when the endpoint changes.
- **Contract change:** no.
- **Red test:** the scratch test above: read 0x0028/0x0005, change the attribute to 0, write; expect the PUT's `attribute_hex` to be `0x0005` (or the write button gone).

## UR-P1-ui-163 — medium — a changed `nodeId` prop keeps the previous node's tab state and writes it to the new node

- **Where:** `ui/src/lib/pages/controller/Node.svelte:142`, `:166-169` (the page reloads on a new `nodeId`); `node/Acl.svelte:136-145` (`onMount` only), likewise `node/Bindings.svelte:34`, `node/Fabrics.svelte:31`, `node/Icd.svelte:224`, `node/Window.svelte:130`, `node/Attributes.svelte:18` (endpoint seeded once).
- **Claim:** `Node.svelte` handles a new `nodeId` in place (an `$effect` reloads the detail), and `ControllerNode` is exported for hosts (`pages/index.ts:21`, README "props `nodeId`"). The tabs load only on mount, so the open tab keeps the old node's data. Measured: ACL tab on node …11, rerender with nodeId …12, add an entry, save → `PUT /controller/nodes/0000000000000012/acl` with node …11's entries (subject `000000000001B669`, `fabric_index` 1). The standalone shell is not affected (`shell/App.svelte:259` keys on `route.nodeId`); a host that routes without a `{#key}` is. Also in `Node.svelte:157-164` the old node's in-flight `load()` can land after the new one (no sequence guard), and `detail`/`focus` are not cleared.
- **M/I:** M (scratch `acl.test.ts`, first case).
- **Fix:** wrap the tab body in `{#key nodeId}` inside `Node.svelte` (and clear `detail`, `error`, `focus`, guard `load()` with a sequence number), or make every tab reload on `nodeId` change.
- **Contract change:** no.
- **Red test:** render `NodeHarness` on …11 tab `acl`, rerender with …12 (an ACL fixture for …12 with a different subject), save; expect the PUT to …12 to carry …12's list (today …11's).

## UR-P1-conc-160 — medium — the camera stays on when the scan is cancelled or the page left while it starts

- **Where:** `ui/src/lib/pages/controller/QrScan.svelte:23-43` (`start`), `:45-49` (`cancel`), `:51` (`onDestroy(cancel)`), `:62` (Cancel shown while `starting`); `ui/src/lib/scan/index.ts:48-80`.
- **Claim:** `stop` is assigned only when `startScan` resolves, which waits for the permission prompt and `video.play()`. A Cancel click or a navigation (onDestroy) during that time runs `cancel()` with `stop === null`; when `startScan` resolves afterwards the stream is live, the 200 ms decode loop runs, and nothing shows it. After onDestroy nothing ever calls `stop` (camera on until the tab closes); after Cancel the hidden scanner can still fill the code field via `onCode`, and a second "Scan" overwrites `stop`, leaking the first stream for good. Cancel during the `import()` instead throws "no video element" (`:30`) and toasts a spurious `scan_failed`.
- **M/I:** M (read); not run (needs a getUserMedia mock).
- **Fix:** a `cancelled`/generation flag set by `cancel()`; after `await startScan` call the returned stop at once when cancelled or destroyed; skip the error toast when cancelled. Or pass an `AbortSignal` into `startScan` and check it after `getUserMedia`.
- **Contract change:** no.
- **Red test:** mock `scan.startScan` with a deferred promise resolving to a spy `stop`; click Scan, click Cancel (or unmount), resolve; expect `stop` called once and no `onCode`/toast afterwards.

## UR-P1-ui-164 — low — ACL and binding numbers reach the server unvalidated

- **Where:** `ui/src/lib/pages/controller/node/Acl.svelte:162`; `node/Bindings.svelte:56`, `:59`.
- **Claim:** the ACL target endpoint and the binding endpoint are `Number(input)` with no check: `"abc"` becomes `NaN`, serialised as `null` (measured: entry `targets: [{ endpoint: null }]` in the PUT); `"1.5"`, `"-1"`, `"70000"` pass as such. The binding group `g:\d+` accepts any length (`g:70000`, group ids are 16-bit). The contract gives `AclTarget.endpoint`/`BindingTarget.endpoint`/`group` no bounds (`openapi.yaml:3108`, `:3136-3137`), so the server-side answer depends on go-fabric's encoder.
- **M/I:** M (scratch `acl.test.ts`, second case; bindings by read).
- **Fix:** accept `^\d{1,5}$` within 0..65534 for endpoints and 0..65535 for groups before adding; and `minimum/maximum` in the contract (contract change, optional).
- **Contract change:** no for the page; yes if bounds are added.
- **Red test:** ACL tab, target endpoint `abc`, add; expect `controller.acl.invalid_target` and no entry added (today an entry with `endpoint: null` is written).

## UR-P1-conc-165 — low — binding writes have no busy state; two quick removes undo each other

- **Where:** `ui/src/lib/pages/controller/node/Bindings.svelte:41-50`, `:100`, `:109`.
- **Claim:** `write()` sends the whole endpoint list computed from `eb.bindings` at click time, and no button is disabled while a write runs. Removing entry A and then entry B before the reload lands writes `[…without A]` and then `[…with A, without B]`: A comes back. Add has the same race against a remove. Every other admin write on these pages has a `busy` guard.
- **M/I:** M (read).
- **Fix:** a `busy` flag that disables the add and remove buttons until `load()` after the write completes.
- **Contract change:** no.
- **Red test:** a list of two bindings, hold the first PUT's answer, click both removes; expect one PUT (or the second list to lack both).

## UR-P1-ui-166 — low — an opened node window cannot be closed after leaving the tab, and expired codes stay on screen

- **Where:** `ui/src/lib/pages/controller/node/Window.svelte:130`, `:173-193`.
- **Claim:** `win` is local; the tab unmounts on a tab switch (`Node.svelte:190-218`), and the Close button is rendered only `{:else}` (window known). After switching back, only "Open" shows; opening again fails while the window is open, and there is no way to revoke the window from the page. Conversely, after `timeout_seconds` the codes and QR stay displayed as if valid.
- **M/I:** M (read).
- **Fix:** always offer "Close window" (the route is idempotent enough to call), and clear `win` (or mark it expired) after `timeout_seconds` with a countdown (`countdown.ts` exists).
- **Contract change:** no (a GET of the window state would be one).
- **Red test:** open a window, switch to overview and back; expect a close control.

## UR-P1-ui-167 — low — the updates tab claims "no queued updates" before the queue loaded or after it failed

- **Where:** `ui/src/lib/pages/controller/node/Updates.svelte:102`, `:104`, `:133-138`; store `controller-store.svelte.ts:111-117`.
- **Claim:** while `controller.updates` is `null` (loading, or `loadUpdates` failed — it only toasts), `queue` is `[]` and the card prints `controller.updates.queue_empty`, a statement about the node that was not fetched. `Diagnostics.svelte:404-405` documents the rule this breaks.
- **M/I:** M (read).
- **Fix:** a loading state while `updates === null` and no load has finished; an error state on failure (keep the error in the store).
- **Contract change:** no.
- **Red test:** answer `GET /controller/updates` with 500; expect no `controller.updates.queue_empty` text.

## UR-P1-ctr-168 — low — the auto-subscribe button reads the live subscription, not the setting

- **Where:** `ui/src/lib/pages/controller/node/Overview.svelte:238-239`; contract `openapi.yaml:2772`, `:1295-1307`.
- **Claim:** the toggle sends `!n.subscription_active`. `subscription_active` is "the default subscription is established" — false while a node with auto-subscribe on is reconnecting or waiting for discovery. Then the button offers "subscribe" and sends `enabled: true`; the operator cannot turn auto-subscribe off for a node that is down. The contract has no field for the setting.
- **M/I:** I (from the contract's field meaning; not run against a live node).
- **Fix:** add the setting (`auto_subscribe`) to `ControllerNodeSummary`/detail and drive the button from it.
- **Contract change:** yes.
- **Red test:** detail with `connection: "reconnecting"`, `subscription_active: false`, `auto_subscribe: true`; expect the button to send `enabled: false`.

## UR-P1-ui-169 — low — the node list's busy marker covers one node at a time

- **Where:** `ui/src/lib/pages/controller/Nodes.svelte:49`, `:77-87`, `:158-173`.
- **Claim:** `busy` holds one id. Acting on node A, then on B sets `busy = B`, re-enabling A's buttons while A's call runs (a second interview/disable/decommission of A can be sent); A's `finally` then sets `busy = null` and re-enables B's buttons while B runs.
- **M/I:** M (read).
- **Fix:** a `Set` of busy ids.
- **Contract change:** no.
- **Red test:** hold A's interview answer, click B's interview, then A's again; expect one POST for A.

## UR-P1-ui-170 — low — a failed discovered-device pairing removes the device from the list

- **Where:** `ui/src/lib/pages/controller/Commission.svelte:189-191`.
- **Claim:** after `commission()` the device is dropped and its passcode cleared whatever the outcome (`commission` swallows the error). A wrong passcode or a timeout means a new 5 s search to retry; after `terms_required` the device vanishes although the retry is pending.
- **M/I:** M (read).
- **Fix:** drop the device only when `result` is set after the call.
- **Contract change:** no.
- **Red test:** discovery answers one device, commission answers 502; expect the row still listed after the failure.

## UR-P1-api-171 — low — the controller Diagnostics page is not exported for hosts

- **Where:** `ui/src/lib/pages/index.ts:18-22`; `ui/src/lib/index.ts:31-33`; shell import `shell/App.svelte:45`; `ui/README.md:52-60`.
- **Claim:** contract 0.9.0 added the controller Diagnostics tab and `routes.controllerDiagnostics` (`context.ts:42`), but `pages/index.ts` exports no component for it, so an embedding host cannot mount it. The natural name `ControllerDiagnostics` collides with the type `ControllerDiagnostics` that `lib/index.ts:33` re-exports from `api` (`api/client.ts:116`) — likely why it is missing. The README's controller page list omits it as well.
- **M/I:** M (read).
- **Fix:** export it under a non-colliding name (e.g. `ControllerDiagnosticsPage`) and list it in the README.
- **Contract change:** no (npm API addition).
- **Red test:** an import test: `import { ControllerDiagnosticsPage } from "@sukramj/fabric-ui"` resolves.

## UR-P1-ui-172 — low — warning, success, error and primary tokens fail text contrast under the HA skin the shell uses

- **Where:** `ui/src/lib/theme/tokens.css:93`, `:102-104`; used as small text e.g. `pages/controller/Diagnostics.svelte:281-283` (event kind chip), `:355` (`discovery_ok`), every `text-[var(--ha-primary-color)]` link.
- **Claim:** on white (standalone shell sets `data-skin="ha"`, `tokens.css:11-12`): warning `#ffa600` 1.96:1 (1.77:1 on its 15% chip), success `#43a047` 3.30:1, primary `#009ac7` 3.26:1, error `#db4437` 4.29:1; standalone skin warning `#d97706` 3.19:1, success `#059669` 3.77:1. WCAG AA for normal text is 4.5:1. The values are Home Assistant's defaults (verbatim port), so inside an HA panel the host's theme rules; the standalone shell inherits the failure.
- **M/I:** M (computed with the WCAG formula; scratch `contrast.js`).
- **Fix:** text-safe variants (`--ha-warning-text-color` etc., darkened mixes such as `color-mix(in srgb, var(--ha-warning-color) 60%, #000)`) for text and chips; keep the raw tokens for fills and icons.
- **Contract change:** no.
- **Red test:** a theme test that resolves the HA-skin fallbacks and asserts ≥ 4.5:1 for each text token on `--ha-card-background-color`.

## UR-P1-test-173 — low — the i18n guards hand-list most enums, so a new contract value passes

- **Where:** `ui/src/lib/i18n/coverage.test.ts:38-56`; `ui/src/lib/i18n/catalogs.test.ts:86-92`.
- **Claim:** only conformance status, event priority (`catalogs.test.ts:50-51`) and BLE reasons are typed against the contract; connection state, ACL privilege/mode, ICD mode, update status, credential kind, step phase and the diagnostic event kinds are literal arrays (the latter lists 3 of the contract's 7 kinds, `openapi.yaml:2300`). The template check (`coverage.test.ts:40`) only asserts that some key starts with the prefix. Today every contract enum value has a key in both catalogues (checked by script), but a value added to the contract would render as a raw key with green tests. No test compares `{placeholder}` sets between en and de, or the names a call site passes (today consistent, checked by script).
- **M/I:** M (read and script).
- **Fix:** `Record<Schema["field"], true>` per enum (as `catalogs.test.ts:50` does), and a placeholder-parity test over both catalogues.
- **Contract change:** no.
- **Red test:** a `Record<DiagnosticEvent["kind"], true>` over `node.diag.kind_*` fails `svelte-check` today with 4 kinds missing from the record, and fails again whenever the contract adds a kind before the catalogue has its key.

## UR-P1-ui-174 — note — the capability badges omit `diagnostics`

- **Where:** `ui/src/lib/pages/controller/Settings.svelte:155`; contract `openapi.yaml:2711`.
- **Claim:** the list is `["ota", "credentials", "ble"]`; 0.9.0's `capabilities.diagnostics` is not shown and has no `controller.settings.cap.diagnostics` key.
- **M/I:** M. **Fix:** add it and the key. **Contract change:** no. **Red test:** status with `diagnostics: true`; expect the badge.

## UR-P1-ui-175 — note — switching the credential kind carries the typed secret into the other field

- **Where:** `ui/src/lib/pages/controller/Settings.svelte:25`, `:49-54`, `:132-137`.
- **Claim:** one `secret` backs both the Wi-Fi passphrase and the Thread dataset; a passphrase typed before switching to Thread is sent as `dataset_hex` (whitespace stripped), and a dataset as a passphrase the other way. Neither is validated (hex, length) before the write.
- **M/I:** M (read). **Fix:** clear `secret` (and `ssid`) on a kind change; check `^[0-9a-f]+$` and even length for the dataset. **Contract change:** no. **Red test:** type a passphrase, switch to Thread; expect an empty field.

## UR-P1-ui-176 — note — "forget" ICD registration runs without confirmation

- **Where:** `ui/src/lib/pages/controller/node/Icd.svelte:280`.
- **Claim:** `unregister(true)` (forget without telling the node; styled destructive) runs on one click; the node keeps a check-in registration for a client that no longer listens. Fabric removal and decommission ask first.
- **M/I:** M. **Fix:** `confirm({ destructive: true })` first. **Contract change:** no. **Red test:** click forget with confirm false; expect no POST.

## UR-P1-ctr-177 — note — the event number is a JS number in `since = last + 1`

- **Where:** `ui/src/lib/pages/controller/node/Events.svelte:218-220`; contract `openapi.yaml:2995` (`EventRecord.number: integer`, Matter EventNumber is 64-bit).
- **Claim:** not a node or fabric id, and practical counters stay below 2^53, but the contract allows any 64-bit value and the increment would be wrong above 2^53.
- **M/I:** I. **Fix:** none needed now; if the contract moves 64-bit counters to strings, follow. **Contract change:** yes if acted on. **Red test:** —

## UR-P1-ui-178 — note — trivially invalid passcodes pass the page's check

- **Where:** `ui/src/lib/pages/controller/Commission.svelte:180-181`.
- **Claim:** only `1..99999998` is checked (as the contract, `openapi.yaml:3205`). The Matter spec's invalid passcodes (11111111 … 88888888, 12345678, 87654321) go to the server and fail later at PASE. Also `Number()` accepts `0x10`, `1e3`. Spec reading for phase 3 `cite`.
- **M/I:** I. **Fix:** `^\d{1,8}$` plus the invalid set. **Contract change:** no. **Red test:** passcode `12345678`; expect `passcode_invalid` and no POST.

## UR-P1-ui-179 — note — the ACL editor offers auth mode `pase`

- **Where:** `ui/src/lib/pages/controller/node/Acl.svelte:134`.
- **Claim:** to my reading of the Matter spec an ACL entry with AuthMode PASE cannot be written (the PASE grant is implicit during commissioning); offering it only produces a node error. Needs a phase 3 `cite` check before acting.
- **M/I:** I. **Fix:** drop `pase` from the add form (keep it for display). **Contract change:** no. **Red test:** the mode select offers case and group only.

## UR-P1-conc-180 — note — node detail and network graph loads have no sequence guard

- **Where:** `ui/src/lib/pages/controller/Node.svelte:157-164`, `:172-175`; `ui/src/lib/pages/controller/Network.svelte:218-246`.
- **Claim:** a stream reload, a retry and (network) the admin refresh can overlap; the older answer can land last. Diagnostics, Events and Inventory guard with a sequence number; these two do not. The server coalesces `node_updated` (`fabricui/gofabric/hooks.go:15`), so the window is small.
- **M/I:** I. **Fix:** the same `seq` pattern. **Contract change:** no. **Red test:** two loads with the first answered last; expect the second's detail.

## UR-P1-ui-181 — note — a failed graph reload replaces the drawn graph with an error

- **Where:** `ui/src/lib/pages/controller/Network.svelte:219-227`, `:275-279`.
- **Claim:** the comment says a reload keeps the drawn graph; a stream-triggered reload that fails sets `error`, and the `{:else if error}` branch precedes the graph, so the graph disappears.
- **M/I:** M (read). **Fix:** with a graph present, toast the error (as `refresh()` does) instead of setting `error`. **Contract change:** no. **Red test:** first topology 200, second 500 after a `graphEpoch` move; expect the graph still rendered.

## UR-P1-ui-182 — note — dates format in the browser's locale, not the UI's

- **Where:** `ui/src/lib/pages/controller/Commission.svelte:108`; `Diagnostics.svelte:248`, `:416`; `node/Icd.svelte:274-275`; `node/Overview.svelte:235`.
- **Claim:** `toLocaleDateString()`/`toLocaleString()` without a locale; with the UI switched to German on an English browser (or the reverse) dates come out in the other language.
- **M/I:** M (read). **Fix:** pass the translator's locale (expose it on the context). **Contract change:** no. **Red test:** locale `de`, `navigator.language` `en-US`; expect `dd.mm.yyyy`.

---

## Checked without finding

- Catalogues: 750 keys each, same key set; no placeholder differs between en and de; every literal `t()` call passes exactly the placeholders its key names (script over `lib/**`); every contract enum behind a template key in my pages has a key in both catalogues; German strings left English are technical terms only (Endpoint, Cluster, Fabric, Thread, Version).
- Commission: double submit guarded (`running` on every start button; implicit form submission is blocked while the submit button is disabled); the terms retry holds the bare request and re-applies the current network and terms; the 409 on ICD register is only the multi-admin conflict (`fabricui/gofabric/gofabric.go:743`).
- Every node and fabric id the controller pages read or write is a hex string (ACL subjects, binding nodes, links, keys); the only JS-number ids are inside attribute values (UR-P1-ctr-162).
- Diagnostics finding keys (`code + service`) are unique for go-fabric v0.22.2's `mdns.Diagnose` (one finding per code per service).
- Non-admin disabled states: the context carries no role, so the pages cannot disable writes; the server's 403 is toasted. By design, not reported.

## Unsettled

- UR-P1-ctr-162's severity rests on how often a 64-bit value above 2^53 appears in the attributes an operator would write back (ACL subjects of other ecosystems' controllers are random 64-bit ids); the verifier may lower it to medium.
- UR-P1-ui-178/-179 depend on Matter spec text (phase 3 `cite`).
- `Commission.svelte:105-108` treats `deadline` as Matter-epoch seconds as the contract says (`openapi.yaml:1951`); whether TCUpdateDeadline is an absolute time is a `cite` question, not checked.
