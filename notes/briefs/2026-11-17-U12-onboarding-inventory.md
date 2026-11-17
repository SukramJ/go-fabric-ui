# Brief U12 — the controller family's Inventory page: onboarding codes kept, shown, printed, attached (issue #68; contract 0.10.0, release 0.10.0)

Date 2026-11-17. Repo: go-fabric-ui (`/home/markus/Dokumente/GitHub/go-fabric-ui`). Branch `wave/U12-onboarding-inventory` off `origin/main` (25f6e7b). CLAUDE.md rules; no attribution lines; English; one PR; do not push or tag.

**Blocked on a go-fabric tag.** The facade this wave uses is merged in go-fabric (PR #520) but in no tag; go-fabric-ui pins v0.22.2. This brief is designed against go-fabric `origin/main` at `dd17ab06` (`git describe`: v0.22.2-22). Owner decision 2026-11-17: go-fabric is tagged after its Review II; **implementation starts with that tag**, and its first step is the "Re-check at the tag" list below. Nothing in this repo imports an untagged go-fabric.

## Why
Issue #68 (owner): a commissioned device's QR or manual pairing code is on a slip of paper that gets lost. go-fabric can keep the code with the node and list it. The page shows one row per commissioned node — name, vendor, product, serial number, the codes, how the code came to be stored — prints one label per node, and attaches a code to a node that has none (typed, scanned, read from a photograph). Codes are secrets: with one and a factory reset anyone can commission the device.

## Sources (read these; nothing else beyond what they point to)
- go-fabric at `dd17ab06` (`git -C /home/markus/Dokumente/GitHub/go-fabric show dd17ab06:<path>`; at the tag: the module cache): `controller/inventory.go`, `controller/client/onboarding.go`, `controller/controller.go:144-155,1123-1125`, `controller/client/fabricops.go:238-240,254-272`, `controller/client/registry.go:477-505`, `controller/client/node.go:735-766`, `secure/setup/decode.go:372-392`, `secure/setup/error_text_test.go:13-18`, `docs/security.md:188,193,216`.
- This repo: `fabricui/controller.go:23-80` (`ControllerSource`), `:359-369` (`NodeBasic`), `:895-919` (port errors); `fabricui/handlers_controller.go:36-80` (`controllerAdmin`, `controllerAdminQuery`), `:98-145` (mount and gate), `:163-197` (`controllerError`), `:550-559` (`DecommissionNode`); `fabricui/problem.go`; `fabricui/handlers_controller_test.go:157-200` (the x-admin scan), `:259` (`TestController_SecretsStayOutOfLogsAndEvents`); `fabricui/gofabric/gofabric.go:58-66` (`ControllerOptions`), `:163-165` (`Capabilities`), `:538-552` (`Decommission`), `:613-625` (`nodeError`); `fabricui/memory_controller.go:160-237`; `contract/openapi.yaml:1114-1140` (`/controller/nodes/{id}`), `:1651-1700` (`/controller/credentials`, the write-only secret route), `:1928-1950` (`Problem`), `:2707-2714` (`ControllerCapabilities`), `:2761` (`ControllerNodeSummary`); `docs/contract.md` "Admin rights" (:359), "Problems" (:376), "Versioning" (:438), "Identifiers, secrets and events of the controller family" (:673); `ui/src/lib/qr.ts` (58 lines), `ui/src/lib/scan/index.ts` (81), `ui/src/lib/pages/controller/QrScan.svelte` (67), `ui/src/lib/pages/controller/Nodes.svelte:90-98` (`decommission`), `ui/src/shell/App.svelte:62,141-150` (`controllerPages`, `controllerTabs`), `ui/src/lib/context.ts:36-64` (routes), `ui/src/lib/i18n/{en,de}.json`.

## go-fabric facts this rests on (at `dd17ab06`)
1. `controller.Controller.Inventory() []InventoryEntry` (`controller/inventory.go:80-122`): every commissioned node in the order of `Controller.Nodes`; nil once Close has begun; a node without a peer address is left out (`:88-91`); a stored code that no longer decodes is listed as none (`:113-118`).
2. `InventoryEntry` (`controller/inventory.go:27-55`): `Node`, `NodeID uint64`, `Label` (NodeLabel), `VendorName`, `ProductName`, `SerialNumber` (Basic Information, "" when the image holds none), `VendorID`, `ProductID` (0 unknown), `HasCode`, `QRCode`, `ManualCode` (`client.SecretCode`), `Source`, `Checks`, `StoredAt`. `QRCode` is zero when the stored code is a manual pairing code: no QR code is made up (`:43-45`, `controller/client/onboarding.go:173-185`). A stored QR code yields both forms; the manual form is 11 digits, or 21 when the code's commissioning flow is not Standard (`onboarding.go:190-199`).
3. `client.SecretCode` (`onboarding.go:88-145`): redacted (`"redacted"`) under every fmt verb, slog, encoding/json and any `encoding.TextMarshaler`; `Reveal()` is the one way to the text; no `UnmarshalText`; not comparable. `InventoryEntry` and `OnboardingCode` redact themselves the same way (`inventory.go:57-70`, `onboarding.go:164-171`).
4. `client.OnboardingSource` (`onboarding.go:61-71`): `OnboardingPaired = 1` (the device accepted the passcode), `OnboardingAttached = 2` (added afterwards, passcode unproven). `client.OnboardingChecks` (`:73-86`), a bit set: `OnboardingCheckedVendor = 1`, `OnboardingCheckedDiscriminator = 2`, `OnboardingCheckedProduct = 4`; 0 for a paired code and for an attached one whose node had nothing to compare with (`:158-161`).
5. `client.Node.SetOnboardingCode(code string) (OnboardingCode, error)` (`onboarding.go:251-292`): trims, validates with the decoders `Controller.Pair` uses, replaces a stored code. Errors: `ErrOnboardingCode` (`:51-54`; unreadable, bad checksum or passcode — wraps the `setup` error), `ErrNotCommissioned` (`controller/client/registry.go:83`), `ErrOnboardingMismatch` (`:55-57`; nothing stored), a store error, `ErrClosed` for a node erased meanwhile (`:398-400`). The very code the node was paired with stays `OnboardingPaired` (`:284-286`). No check proves the passcode or the unit (`:265-270`).
6. The mismatch is an error **text**, no typed error (`onboarding.go:319,324,335`): `"client: onboarding code names another device: its vendor id is 0x%04X, the node's is 0x%04X"`, `"… its product id is 0x%04X, the node's is 0x%04X"`, `"… its discriminator is %d, the node advertised %d"`. The discriminator check runs only for a record last seen commissionable in Basic mode (`:329`); a manual code's four bits are compared with the advertised one's upper four (`:331-333`).
7. A manual code's non-digit characters are ignored (`secure/setup/decode.go:373,381-387`); a QR code must start with `MT:` (`onboarding.go:216`). `setup`'s errors name the rule and never the passcode (`secure/setup/error_text_test.go:13-18`, `decode.go:193,428`); they can name a vendor or product id and one offending character with its position (`decode.go:103,227`).
8. `client.Node.ClearOnboardingCode() error` (`onboarding.go:294-310`): removes the stored code, also an unreadable one; nil on an erased node.
9. **The code does not outlive the node.** `Registry.Decommission` (`fabricops.go:260-272`) ends in `deleteNode` (`registry.go:482-503`), which calls `Node.Erase` (`node.go:740-766`): `n.onboarding = nil` (`:747`) and `store.EraseNode` (`:762`). `Registry.Forget` takes the same path (`registry.go:477`). A removal the peer refused keeps the node and so its code (`fabricops.go:258-259`). This repo's `Decommission` calls `Registry.Decommission` (`fabricui/gofabric/gofabric.go:547`).
10. `controller.Config.KeepOnboardingCodes` (`controller.go:144-155`), off by default: `Pair` keeps the code when the commissioning succeeded and the device advertised Basic commissioning mode — not for an Enhanced Commissioning Window, not for a pairing by address, not when `PairOptions.Edit` replaced the passcode (`controller.go:1123-1125`, `fabricops.go:238-240`, `onboarding.go:365-388`). The controller exposes no getter for the setting.
11. `docs/security.md:216`: the store holds the code in plain text unless sealed (`sqlstore.go:52-77`: the code's value is a sealed row with a `Sealer`); "An API that serves the inventory hands out setup passcodes: give it the authorisation an export of the fabric's keys would need, and call `Reveal()` only where the code is shown or printed".

### Re-check at the tag (first step of the implementation; a difference stops the wave and goes back to planning)
- R1. Names and signatures of facts 1, 2, 4, 5, 8: `Controller.Inventory`, every `InventoryEntry` field, the two `OnboardingSource` values, the three `OnboardingChecks` bits and their numeric values, `Node.SetOnboardingCode`, `Node.ClearOnboardingCode`, `Node.OnboardingCode`.
- R2. `SecretCode` still redacts under fmt, slog and JSON and `Reveal()` still returns the text (fact 3).
- R3. The three mismatch texts of fact 6 — or a typed mismatch error if go-fabric took hand-over H1; then decision 6 switches to its fields.
- R4. `ErrOnboardingCode`'s wrapped `setup` texts still carry no passcode (fact 7; `error_text_test.go` present and passing at the tag).
- R5. `Decommission` and `Forget` still erase the code (fact 9), and `SetOnboardingCode` on an erased node still fails rather than stores.
- R6. `Config.KeepOnboardingCodes` keeps its name, default and the three exclusions (fact 10); whether a getter appeared (open question Q3).
- R7. The tag's number and its `CHANGELOG.md` section, for `go.mod`, the changelog line and CLAUDE.md's state sentence; any other facade change between v0.22.2 and the tag that touches `fabricui/gofabric`.
- R8. `Inventory()` still skips nothing but nodes without a peer address, and still returns the entries in `Nodes()` order.

## Decisions (do not reopen)
1. **Contract 0.10.0, additive; three routes, all `x-admin`, all `x-no-store`.**
   - `GET /controller/inventory` → `OnboardingInventory` `{ entries: OnboardingInventoryEntry[] }`. Admin although it is a GET, as `GET /node/pairing` and `GET /controller/discovery` are: the answer carries setup passcodes. One route, not a public list plus a secret one — the non-secret columns are already on `GET /controller/nodes`.
   - `PUT /controller/nodes/{id}/onboarding-code`, body `OnboardingCodeRequest` `{ code: string }` → 200 with the node's `OnboardingInventoryEntry` (the row the page replaces; it shows `source` and `checks` of what was stored).
   - `DELETE /controller/nodes/{id}/onboarding-code` → 204, also when the node had none. In scope because a stored secret needs a way out that is not "decommission the node", and a wrong attached code is only otherwise replaceable.
   - The code travels in a request body or a response body, never in a path or a query.
   - Schema names carry `Onboarding`: "inventory" already names the endpoint and cluster inventory of contract 0.4.0 (`EndpointDetail`, `ClusterDetail`, the node page's Inventory tab).
2. **Problems.** Two new codes in `Problem.code` (an enum value added, as `node_limit` and `terms_required` were):
   - `onboarding_code_invalid` — 400, port `ErrOnboardingCode`. Title "Not an onboarding code"; `detail` is go-fabric's reason (fact 7). Distinct from `bad_request` (a malformed body, an empty `code`, a bad `{id}`), which the page words differently.
   - `onboarding_code_mismatch` — 409, port `ErrOnboardingMismatch`. Title "The code names another device"; `detail` is go-fabric's sentence after the sentinel, e.g. `its vendor id is 0x1234, the node's is 0xFFF1`. Nothing was stored.
   - Unknown node: 404 `not_found` (`ErrNodeNotFound`, existing). Node not commissioned or erased meanwhile: 409 `conflict` (`ErrConflict`, existing) with a fixed detail. Source without the capability: 501 `not_implemented` on all three. Store failure: 500 `internal`, no detail, logged with the error's text (a store error carries no code: `SecretCode` redacts, and the code is never an argument of the wrapped error — R2).
   - No structured members on the mismatch problem in 0.10.0 (see H1).
3. **Is the differing value of a mismatch a secret? No.** Vendor and product id are the node's public identity (already on `GET /controller/nodes`); the discriminator is what a commissionable device advertises to everyone in range. None of them is the passcode, and none narrows it. They may stand in the problem's `detail` on this admin route and on the page. They are still not logged: the handler logs nothing for a 4xx of these routes.
4. **`Cache-Control: no-store` on every answer of the three routes**, problems and the 403 included. Mechanism: a vendor extension `x-no-store: true` beside `x-admin` in the contract, a table `controllerNoStore` in `fabricui/handlers_controller.go`, the header set in `controllerGate` before the admin check; a test scans the contract against the table as `TestControllerAdminMatchesTheContract` does. Not a `headers:` object on the 200: that would cover neither the problems nor the 403. No API answer sets `Cache-Control` today (only `stream.go:46` and `spa.go`); this wave does not change the others.
5. **Capability and opt-in.** `ControllerCapabilities.onboarding_codes` (bool, required in 0.10.0 with the fixtures updated, as `diagnostics` was in 0.9.0). The `gofabric` adapter reports it from a new `ControllerOptions.OnboardingCodes bool`, **false by default**: a host that upgrades does not start handing out passcodes; serving them is the host's stated decision (fact 11). False: the three routes answer 501 with the fixed sentence, the tab is hidden, `has_onboarding_code` is absent.
6. **Port (`fabricui/controller.go`).** fabricui imports no `controller/client`, so the port has its own redacting type:
   ```go
   // SecretText is a text the port carries and nothing renders: redacted
   // under fmt, slog, encoding/json and encoding.TextMarshaler. Reveal is
   // the one way to it; the zero value is no text.
   type SecretText struct{ /* text behind a pointer, a [0]func() to forbid == */ }
   func NewSecretText(s string) SecretText
   func (s SecretText) Reveal() string
   func (s SecretText) IsZero() bool
   // String, GoString, LogValue, MarshalText: "redacted" ("" when zero)

   type OnboardingSource string // "paired" | "attached"
   type OnboardingChecks struct{ Vendor, Product, Discriminator bool }

   type OnboardingEntry struct {
       NodeID                                uint64
       Label                                 string
       VendorName, ProductName, SerialNumber string
       VendorID, ProductID                   uint16
       HasCode                               bool
       QRCode, ManualCode                    SecretText // QRCode zero for a stored manual code
       Source                                OnboardingSource
       Checks                                OnboardingChecks
       StoredAt                              time.Time
   }

   // on ControllerSource
   OnboardingInventory() []OnboardingEntry
   SetOnboardingCode(ctx context.Context, id uint64, code string) (OnboardingEntry, error)
   ClearOnboardingCode(ctx context.Context, id uint64) error

   // on ControllerCapabilities
   OnboardingCodes bool

   var ErrOnboardingCode = errors.New("fabricui: not an onboarding code")
   var ErrOnboardingMismatch = errors.New("fabricui: the onboarding code names another device")
   ```
   `SetOnboardingCode` takes the code as a plain string, as `Pair` does. The handler checks `Capabilities().OnboardingCodes` first (501) and is the one place outside the adapter that calls `Reveal()`: where it fills the response. `OnboardingEntry` gets `String`/`GoString`/`LogValue` that name the node id and "code redacted"/"no code", as go-fabric's entry does. The wrapped port errors carry the reason after the sentinel (`fmt.Errorf("%w: %s", ErrOnboardingMismatch, reason)`); the handler puts the text after the sentinel into `detail`.
7. **Adapter (`fabricui/gofabric`, a new `inventory.go`).** `OnboardingInventory`: `ctrl.Inventory()` mapped one to one; `NewSecretText(e.QRCode.Reveal())` in one expression, the revealed string bound to no variable; `Source` 1→`paired`, 2→`attached`; `Checks` bit by bit. `SetOnboardingCode`: `lookup(id)` (404), `n.n.SetOnboardingCode(code)`, then the node's entry from a fresh `ctrl.Inventory()` (by node id; missing → `ErrNodeNotFound`). Errors by `errors.Is`, not through `nodeError` (whose default is 502 `node_error`, wrong for a local store): `client.ErrOnboardingCode` → `ErrOnboardingCode` with the text after go-fabric's sentinel; `client.ErrOnboardingMismatch` → `ErrOnboardingMismatch` likewise; `client.ErrNotCommissioned`, `client.ErrClosed` → `ErrConflict`; anything else unwrapped (500). `ClearOnboardingCode`: `lookup`, `n.n.ClearOnboardingCode()`. With `ControllerOptions.OnboardingCodes` false the three methods are not reached (the handler's 501); they answer `nil` / `ErrNotImplemented` anyway.
8. **The removal warning needs no secret route.** `ControllerNodeSummary.has_onboarding_code` (optional boolean, 0.10.0; existence only, no secret): the `GET /controller/nodes` handler derives it from `OnboardingInventory()`'s `HasCode` when the capability is on, and omits the field otherwise. No new method on `ControllerNode`.
9. **Memory port (`fabricui/memory_controller.go`).** A map node id → stored code, `NoOnboardingCodes` to switch the capability off (as `NoDiagnostics`), `Capabilities().OnboardingCodes = !m.NoOnboardingCodes`. `SetOnboardingCode` there: trimmed text starting `MT:` and matching the scan module's payload shape is a QR code (manual code = a fixed test value held beside it), 11 or 21 digits after dropping non-digits a manual code (no QR), anything else `ErrOnboardingCode: Invalid pairing code`; a settable `OnboardingMismatch string` makes the next attach answer `ErrOnboardingMismatch` with that reason. `Decommission` deletes the code. It validates shape only — it is the test port, not a second decoder; the real decoders are tested through the adapter's e2e.
10. **Fixtures**, recorded from the memory port, codes are the spec's public test values (`MT:Y.K9042C00KA0648G00`, `34970112332`), never a real device's: `controller-inventory.200.json` (three nodes: a paired QR code with both forms; an attached 11-digit manual code, `checks.discriminator` true and `checks.vendor`/`checks.product` false — an 11-digit code carries no vendor or product id —, no `qr_code`; a node without a code), `controller-inventory.403.json`, `controller-inventory.501.json`, `controller-node-onboarding-code-set.200.json`, `.400-invalid.json`, `.409-mismatch.json`, `.404.json`, `controller-node-onboarding-code-clear.204.json`; `controller-status.200.json` and `controller-status.200-ble.json` gain the capability, `controller-nodes.200.json` one `has_onboarding_code`. `contract-check` on both sides.
11. **UI.**
    - **Placement:** a fifth controller tab `inventory` after `nodes` (en "Inventory", de "Inventar"), route `#/controller/inventory` (`Routes.controllerInventory`), page `ui/src/lib/pages/controller/OnboardingInventory.svelte` (not `Inventory.svelte`: `controller/node/Inventory.svelte` exists). In `controllerTabs` only while `capabilities.onboarding_codes`; the page itself shows the kit's "not served" state on a 501 and an "admin rights required" state on a 403 instead of a toast.
    - **State:** the entries live in the page component, not in `controller-store`; they are dropped when the page is left. Never in `localStorage`/`sessionStorage`, the location hash, a toast, the console or an error text. The client's three calls pass `cache: "no-store"`. The page refetches on `node_added` / `node_removed` of the controller stream and after its own writes.
    - **Table (DataTable):** name (NodeLabel, falling back as `nodeName` does; links to `#/controller/node/<id>`), vendor, product, serial number, code column: a badge — "Accepted at pairing" (`paired`), "Attached, unproven" (`attached`) with chips for the checks passed (vendor, product, discriminator) or "no identity check was possible", "No code" — and the stored date. Action per row: "Attach code" (none stored) or "Replace" and "Remove code" (confirm).
    - **Codes are shown on request:** a row opens (button "Show code") to the QR image (`renderQrSvg(entry.qr_code)` from `ui/src/lib/qr.ts`) and the manual code with a copy button (the clipboard refusal outside a secure context as the pairing page words it). An entry without `qr_code` shows the manual code and the sentence that no QR image exists for a manual code. Nothing is drawn for closed rows.
    - **Print:** a button "Print labels" calls `window.print()`; the page renders a `.print-labels` block (screen: `display: none`) with one label per node that has a code — name, QR image when there is one, manual code — each `break-inside: avoid`, in a wrapping grid. `@media print` (in `ui/src/lib/theme/tokens.css`, the first print rule of the bundle) hides the shell, the tabs, the table and the toasts and shows the block; black on white whatever the theme. Nodes without a code get no label.
    - **Attach dialog**, three ways into one text field, then one "Attach" button (nothing is sent before it):
      1. type: the field takes `MT:…` or the 11/21 digits with any separators (fact 7);
      2. camera: the existing `QrScan.svelte` (`onCode` fills the field); outside a secure context it shows the new sentence instead of the button;
      3. photograph: `<input type="file" accept="image/*">` without `capture`, so a phone offers both its camera and its gallery; a new `decodeImage(file: Blob): Promise<string | null>` in `ui/src/lib/scan/index.ts` — `createImageBitmap`, the native detector when present, else a canvas scaled to at most 1600 px on its longer side and `jsQR` with `inversionAttempts: "attemptBoth"`, the result kept only when `isPayload`. The photograph is decoded in the browser and uploaded nowhere. It works without a secure context (where `BarcodeDetector` is absent, `jsQR` is the path — the implementation verifies this on a plain-HTTP LAN address).
      The dialog states: a live camera needs a secure page (HTTPS or localhost); a photograph works on any page. A refusal stays in the dialog: the localized title by problem code and, under it, the module's reason (`detail`, English, as go-fabric words it). Success: the row is replaced, a toast without the code.
    - **Decoder and renderer:** `jsqr` 1.4.0 and `qrcode` 1.5.4, both pinned in `ui/package.json:70-71`, listed in `THIRD-PARTY-NOTICES.md:22-23`, used by `ui/src/lib/scan/index.ts:7,40` and `ui/src/lib/qr.ts:7,47`. No new dependency; `decodeImage` stays in the `scan` chunk (dynamic import).
    - **Removal warning:** `Nodes.svelte:90-98` `decommission()` — the confirm body gains a second sentence when `n.has_onboarding_code`. The Inventory page offers no decommission.
    - **i18n** (en / de):
      - `controller.tab.inventory`: Inventory / Inventar
      - `controller.inventory.intro`: The onboarding code of each commissioned device, kept so that it survives the paper it came on. A code is a secret: with it and a factory reset anyone can commission the device. / Der Onboarding-Code jedes gekoppelten Geräts, aufbewahrt, damit er den Zettel überlebt. Ein Code ist ein Geheimnis: Mit ihm und einem Werksreset kann jeder das Gerät koppeln.
      - `controller.inventory.col_name|col_vendor|col_product|col_serial|col_code`: Name, Vendor, Product, Serial number, Code / Name, Hersteller, Produkt, Seriennummer, Code
      - `controller.inventory.source_paired`: Accepted at pairing / Beim Koppeln akzeptiert
      - `controller.inventory.source_attached`: Attached, unproven / Nachträglich hinterlegt, ungeprüft
      - `controller.inventory.source_attached_hint`: A commissioned device offers no way to test a passcode. / Ein gekoppeltes Gerät bietet keinen Weg, einen Passcode zu prüfen.
      - `controller.inventory.check_vendor|check_product|check_discriminator`: Vendor matches, Product matches, Discriminator matches / Hersteller stimmt, Produkt stimmt, Discriminator stimmt
      - `controller.inventory.checks_none`: No identity check was possible. / Keine Identitätsprüfung möglich.
      - `controller.inventory.no_code`: No code / Kein Code
      - `controller.inventory.stored_at`: Stored {when} / Gespeichert {when}
      - `controller.inventory.show|hide`: Show code, Hide code / Code anzeigen, Code verbergen
      - `controller.inventory.manual_code`: Manual code / Manueller Code
      - `controller.inventory.no_qr`: A manual code carries too little for a QR code; none is made up. / Ein manueller Code enthält zu wenig für einen QR-Code; es wird keiner erfunden.
      - `controller.inventory.attach|replace|remove`: Attach code, Replace, Remove code / Code hinterlegen, Ersetzen, Code entfernen
      - `controller.inventory.attach_title`: Attach the code of {name} / Code von {name} hinterlegen
      - `controller.inventory.attach_hint`: Enter the QR payload (MT:…) or the manual pairing code printed on the device, scan it, or choose a photograph of it. / QR-Inhalt (MT:…) oder den aufgedruckten manuellen Code eingeben, scannen oder ein Foto davon wählen.
      - `controller.inventory.photo`: Choose or take a photograph / Foto wählen oder aufnehmen
      - `controller.inventory.photo_none`: No onboarding QR code was found in the photograph. / Im Foto wurde kein Onboarding-QR-Code gefunden.
      - `controller.inventory.camera_note`: The live camera needs a secure page (HTTPS or localhost). A photograph works on any page and is read in the browser; it is not uploaded. / Die Live-Kamera braucht eine sichere Seite (HTTPS oder localhost). Ein Foto funktioniert auf jeder Seite und wird im Browser gelesen; es wird nicht hochgeladen.
      - `controller.inventory.attached`: The code is stored. / Der Code ist gespeichert.
      - `controller.inventory.refused_invalid`: This is not an onboarding code. / Das ist kein Onboarding-Code.
      - `controller.inventory.refused_mismatch`: The code names another device; nothing was stored. / Der Code gehört zu einem anderen Gerät; es wurde nichts gespeichert.
      - `controller.inventory.remove_confirm`: Remove the stored code of {name}? / Den gespeicherten Code von {name} entfernen?
      - `controller.inventory.removed`: The code is removed. / Der Code ist entfernt.
      - `controller.inventory.print`: Print labels / Etiketten drucken
      - `controller.inventory.print_empty`: No node has a code to print. / Kein Knoten hat einen Code zum Drucken.
      - `controller.inventory.admin_required`: The inventory shows secrets and needs admin rights. / Das Inventar zeigt Geheimnisse und braucht Admin-Rechte.
      - `controller.inventory.not_served`: This host does not serve the inventory. / Dieser Host stellt das Inventar nicht bereit.
      - `controller.nodes.remove_confirm_code`: Its stored onboarding code is deleted with it. / Der gespeicherte Onboarding-Code wird mit gelöscht.
12. **Export: printing is the export of this wave; a file export is left out.** The issue states a constraint ("an export is an explicit action"), not a request for a file. Print and the per-row copy are explicit actions the user starts. A CSV or JSON download would put every passcode into a file outside the database's protection (fact 11) for no stated need; it is a later wave's decision with its own route.
13. **Release 0.10.0:** `go.mod` to the go-fabric tag; `changelog.md` `[0.10.0]`; version files; `contract/previous` = the 0.9.0 contract; `docs/contract.md` a section "Onboarding inventory (v0.10.0)" (routes, the two problems, `x-no-store`, the secrets bullet extended to the three routes) ; `docs/embedding.md` the three source methods, `ControllerOptions.OnboardingCodes` and the hosts' obligations below; CLAUDE.md's state sentence and the Security rule ("secrets (passcodes, onboarding codes)"); README status.

## Proposed contract (`contract/openapi.yaml`, version 0.10.0)
```yaml
  /controller/inventory:
    get:
      tags: [controller]
      operationId: getOnboardingInventory
      x-admin: true
      x-no-store: true
      summary: Every commissioned node with the onboarding code kept for it (0.10.0). The codes are secrets — the answer is never cached.
      responses:
        "200":
          description: The inventory, in the order of `GET /controller/nodes`.
          content:
            application/json:
              schema: { $ref: "#/components/schemas/OnboardingInventory" }
        "403": { $ref: "#/components/responses/Problem" }
        "500": { $ref: "#/components/responses/Problem" }
        "501": { $ref: "#/components/responses/Problem" }
  /controller/nodes/{id}/onboarding-code:
    parameters:
      - $ref: "#/components/parameters/NodeID"
    put:
      tags: [controller]
      operationId: setNodeOnboardingCode
      x-admin: true
      x-no-store: true
      summary: Attach the code printed on the device to a commissioned node (0.10.0); it replaces a stored one. The code is validated and compared with what the node is known to be; its passcode stays unproven.
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/OnboardingCodeRequest" }
      responses:
        "200":
          description: The node's entry with the stored code.
          content:
            application/json:
              schema: { $ref: "#/components/schemas/OnboardingInventoryEntry" }
        "400": { $ref: "#/components/responses/Problem" }   # bad_request, onboarding_code_invalid
        "403": { $ref: "#/components/responses/Problem" }
        "404": { $ref: "#/components/responses/Problem" }
        "409": { $ref: "#/components/responses/Problem" }   # onboarding_code_mismatch, conflict
        "500": { $ref: "#/components/responses/Problem" }
        "501": { $ref: "#/components/responses/Problem" }
    delete:
      tags: [controller]
      operationId: clearNodeOnboardingCode
      x-admin: true
      x-no-store: true
      summary: Remove the node's stored onboarding code (0.10.0); 204 also when it had none.
      responses:
        "204": { description: The node has no stored code. }
        "400": { $ref: "#/components/responses/Problem" }
        "403": { $ref: "#/components/responses/Problem" }
        "404": { $ref: "#/components/responses/Problem" }
        "500": { $ref: "#/components/responses/Problem" }
        "501": { $ref: "#/components/responses/Problem" }

# components.schemas
    OnboardingInventory:
      type: object
      description: "The commissioned nodes with their kept onboarding codes (0.10.0, go-fabric `controller.Controller.Inventory`)."
      required: [entries]
      properties:
        entries:
          type: array
          items: { $ref: "#/components/schemas/OnboardingInventoryEntry" }
    OnboardingInventoryEntry:
      type: object
      required: [node_id_hex, has_code]
      properties:
        node_id_hex: { type: string, description: The node id, 16 uppercase hex digits. }
        node_label: { type: string, description: "The node's NodeLabel; absent when unset or not read yet." }
        vendor_id: { type: integer, minimum: 0, maximum: 65535, description: Absent when unknown. }
        vendor_name: { type: string }
        product_id: { type: integer, minimum: 0, maximum: 65535, description: Absent when unknown. }
        product_name: { type: string }
        serial_number: { type: string }
        has_code: { type: boolean, description: "A code is stored; the fields below are absent without one." }
        qr_code: { type: string, description: "Secret. The QR code's text (`MT:…`), drawn by the page; absent when the stored code is a manual pairing code, from which no QR code is made up." }
        manual_code: { type: string, description: "Secret. The manual pairing code, 11 or 21 digits without separators." }
        source: { $ref: "#/components/schemas/OnboardingSource" }
        checks: { $ref: "#/components/schemas/OnboardingChecks" }
        stored_at: { type: string, format: date-time }
    OnboardingSource:
      type: string
      enum: [paired, attached]
      description: "`paired`: the device accepted the code's passcode at commissioning. `attached`: added afterwards, its passcode unproven."
    OnboardingChecks:
      type: object
      description: "The identity checks an attached code passed; a check runs only when the node's side of it is known. All false for a paired code. No check proves the passcode or the unit."
      required: [vendor, product, discriminator]
      properties:
        vendor: { type: boolean, description: "The code's vendor id equals the node's." }
        product: { type: boolean, description: "The code's product id equals the node's; compared only together with the vendor id." }
        discriminator: { type: boolean, description: "The code's discriminator equals the one the node advertised while commissionable." }
    OnboardingCodeRequest:
      type: object
      required: [code]
      properties:
        code: { type: string, minLength: 1, maxLength: 512, description: "Secret. A QR code's text (`MT:…`) or a manual pairing code; separators in a manual code are ignored." }

# additions to existing schemas
#   Problem.code enum: + onboarding_code_invalid, onboarding_code_mismatch
#   ControllerCapabilities: + onboarding_codes (boolean, required): "The source serves the onboarding inventory (0.10.0); false: the three routes answer 501."
#   ControllerNodeSummary: + has_onboarding_code (boolean, optional): "An onboarding code is stored for the node (0.10.0); absent when the source serves no inventory. Decommissioning the node deletes it."
#   info.description: one sentence on `x-no-store: true` — the answer, a problem included, carries `Cache-Control: no-store`.
```

## Hand-over to go-fabric (before its tag, if the owner agrees — Q1)
- **H1.** A typed mismatch: `ErrOnboardingMismatch` carries its values only as text (fact 6). A `*client.OnboardingMismatchError{Field, Code, Node}` (still `errors.Is` the sentinel) would let the contract carry `field`, `code_value`, `node_value` and the page word the reason in the user's language. Without it the page shows go-fabric's English sentence, and the adapter cuts it after the sentinel's text — a string operation on an error message, which R3 guards.

## Guards
- `export PATH=$PATH:$HOME/go/bin`. `make ci`, `make race`, `make contract-check`, `make generate-check`, `make version-check`, `make notices-check` (unchanged list), `npm test` in `ui/`, `make web`, `make e2e`, `make e2e-web` — in the background with a summary log.
- **New, each with its bite line:**
  - G1 `TestController_SecretsStayOutOfLogsAndEvents` extended: after `GET /controller/inventory`, a successful attach, an invalid one, a mismatch and a store failure, neither the captured log nor any stream event contains the QR text, the manual code or the request's code. Bites when a handler logs a body or an error that wraps the code.
  - G2 `TestSecretTextRedacts`: `%v`, `%+v`, `%#v`, `%s` of a `SecretText`, of an `OnboardingEntry` and of a struct holding one in an unexported field; `slog` text and JSON handlers; `json.Marshal`. Bites when the port type loses a method.
  - G3 `TestControllerNoStoreMatchesTheContract` (the scan) and a handler test: `Cache-Control: no-store` on the 200, the 403, the 400, the 409 and the 501 of the three routes.
  - G4 the three routes in `controllerAdmin`; the existing scan test covers it; fixture `controller-inventory.403.json`.
  - G5 adapter e2e (`fabricui/gofabric/e2e_test.go` pattern, `KeepOnboardingCodes` on): a paired node's entry is `paired` with both codes; an attach of another product's code answers `onboarding_code_mismatch` with go-fabric's sentence; after `DELETE /controller/nodes/{id}` the inventory no longer lists the node.
  - G6 UI: `OnboardingInventory.test.ts` — no QR image for an entry without `qr_code`; the insecure-context sentence; the refusal's reason in the dialog; nothing written to `localStorage`. `scan.test.ts` — `decodeImage` returns null for a non-payload QR code. `Nodes.test.ts` — the second sentence of the confirm only with `has_onboarding_code`. The Playwright smoke opens the tab.
  - G7 `catalogs.test.ts` / `coverage.test.ts`: every new key in en and de.

## What a host must do (go-fabric-server wires it; `docs/embedding.md`)
1. Set `gofabric.ControllerOptions.OnboardingCodes = true` to serve the inventory at all — and only behind an access control fit for an export of the fabric's keys (fact 11). In host mode `Options.Admin` decides who reads the codes. **Token mode is no access control**: the token is served by `GET /controller/status` to whoever reaches the handler; it stops a cross-origin page, not a person on the network. A host in token mode that sets the option hands the passcodes to everyone who can open the UI, and must restrict that itself (bind address, proxy authentication).
2. Set `controller.Config.KeepOnboardingCodes` if codes of newly paired devices are to be kept; without it the inventory holds only attached codes. Configure a `Sealer` if the store's secrets are to be sealed at rest.
3. Not log request or response bodies of the three routes in its own middleware or proxy, and not cache them; the handler sends `Cache-Control: no-store`, a proxy has to honour it.
4. A host with its own `ControllerSource` implements the three methods (or returns `OnboardingCodes: false` and `ErrNotImplemented`) — the interface grows, as it did in 0.9.0.
5. Serve the UI over HTTPS (or localhost) for the live camera; the photograph path and typing work without.

## Left out
A file export (decision 12). A structured mismatch (H1). An event on the controller stream for a code stored or removed (the page refetches after its own write; a second browser sees it on its next load). Bulk attach. Editing NodeLabel from this page. A setting in the UI for `KeepOnboardingCodes` (host configuration).

## Open questions for the owner
- **Q1.** H1 — ask go-fabric for the typed mismatch error before its tag, or ship 0.10.0 with the English sentence in `detail`?
- **Q2.** Token mode: is the opt-in of decision 5 plus the obligation of host point 1 enough, or shall the handler refuse the three routes outright (501) whenever `Options.Admin` is nil?
- **Q3.** Shall the page say whether codes of newly paired devices are kept? The controller has no getter for `Config.KeepOnboardingCodes` (fact 10); it would need a second `ControllerOptions` field the host sets, or a getter in go-fabric.
- **Q4.** Label format: the wrapping grid on the browser's paper size, or one label per page for a label printer (a `@page` size the user picks)?
- **Q5.** `DELETE …/onboarding-code` and "Remove code" are beyond the issue's text (decision 1) — keep?
- **Q6.** Codes behind "Show code" per row (decision 11) or always visible in the table, as the issue's row description can be read?

## Report (fixed shape, short)
Branch, commits, the result of the re-check at the tag (R1–R8, each "as designed" or the difference), guards with HEAD, differences from this brief, left out, questions. Close nothing; the session closes #68 with the PR.
