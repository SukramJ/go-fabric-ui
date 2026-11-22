# Brief U12 — the controller family's Inventory page: onboarding codes kept, shown, printed, attached (issue #68; contract 0.13.0, release 0.13.0)

Date 2026-11-17, rebased 2026-11-22. Repo: go-fabric-ui (`/home/markus/Dokumente/GitHub/go-fabric-ui`). Branch `wave/U12-onboarding-inventory` off `origin/main` (f4d625a, `v0.12.0`). CLAUDE.md rules; no attribution lines; English; one PR; do not push or tag.

**Unblocked.** The facade this wave uses is in go-fabric `v0.23.0` (PR #520, the typed mismatch of #537); `main` pins `v0.24.1`. The brief was designed against go-fabric `dd17ab06` and contract 0.9.0; the re-check at the tag is done (below) and the wave starts with the implementation.

## Rebase 2026-11-22 (Fable; decisions 81 and 85)
Re-check R1–R8 at go-fabric `v0.24.1`: `notes/reads/read-U12-recheck.md`. No difference the decisions and Q1–Q6 cannot absorb. What this rebase changed, and why:
1. **Versions.** Contract 0.10.0 → **0.13.0**, release → **v0.13.0**, additive over 0.12.0 (0.10.0–0.12.0 went to the review, decisions 37–82). `contract/previous` = the 0.12.0 contract. `go.mod` stays (v0.24.1 is pinned).
2. **Q1 applied (R3).** go-fabric has `*client.OnboardingMismatchError{Field, Code, Node}`. Decisions 2, 6, 7, 9, 10, 11 and the contract carry `field`, `code_value`, `node_value`; three i18n keys are new (`mismatch_vendor|product|discriminator`); H1 is done. No error text is cut anywhere — also not for `onboarding_code_invalid`, whose reason is the wrapped error (R4), carried by a port type.
3. **Decision 4's mechanism.** Since 0.10.0 every answer of the API carries `Cache-Control: no-store` (`apiHeaders`, UR-P0-leak-25). The decision stands and is already met; `x-no-store`, the `controllerNoStore` table and the scan test are dropped, G3 is a handler test by name.
4. **0.12.0's conventions in the contract fragment:** `x-admin` as on every other admin operation; 413 and 415 on the PUT (decision 38); the shared 500 (decision 82); the PUT body in `controllerBodies` (strict, decision 43); 501 through `notImplemented` (decisions 47, 79); the two problems get their `### problem-<code>` entries in `docs/contract.md`; no `info.x-accepted-breaks` (nothing breaks).
5. **Order (R8).** This repo's `GET /controller/nodes` is sorted by node id; go-fabric's inventory is in registry order. The adapter sorts the entries by node id.
6. **Error mapping of the three handlers** is their own and closed (decision 2): a host source's other error kinds answer 500, so UR-P4-api-391 does not grow.
7. **Fact 10 (R6):** `Controller.PairDiscovered` (v0.24.0) keeps no code; a device commissioned from the discovery list shows "No code". Fact 9 (R5): a store that fails the erase now fails the decommission.
8. **Citations:** go-fabric facts re-cited at `v0.24.1`, this repo's sources at `f4d625a`.
9. **Guards:** the list is today's pipeline set and the guards the new routes run into (status guard, body table, contract guard, i18n guard, `test-386`, `e2e-web`); G3 is a handler test, G4 holds Q2, G7 the closed mapping (new).
10. **New section** "Review registers this wave touches".

## Why
Issue #68 (owner): a commissioned device's QR or manual pairing code is on a slip of paper that gets lost. go-fabric can keep the code with the node and list it. The page shows one row per commissioned node — name, vendor, product, serial number, the codes, how the code came to be stored — prints one label per node, and attaches a code to a node that has none (typed, scanned, read from a photograph). Codes are secrets: with one and a factory reset anyone can commission the device.

## Sources (read these; nothing else beyond what they point to)
- go-fabric at `v0.24.1` (the module cache, or `git archive v0.24.1`; never the working tree): `controller/inventory.go`, `controller/client/onboarding.go`, `controller/controller.go:147-158,1253-1254,1374-1379`, `controller/client/fabricops.go:157-167,269-271,294-306`, `controller/client/registry.go:504-549,682-690`, `controller/client/node.go:761-788`, `secure/setup/decode.go:372-392`, `secure/setup/error_text_test.go:12-18`, `docs/security.md:204,209,232`. The re-check: `notes/reads/read-U12-recheck.md`.
- This repo at `f4d625a`: `fabricui/controller.go:24` (`ControllerSource`), `:304` (`ControllerCapabilities`), `:369` (`NodeBasic`), `:939-976` (port errors); `fabricui/handlers_controller.go:59` (`controllerAdmin`), `:104` (`controllerBodies`), `:158-190` (`controllerGate`), `:214` (`controllerError`), `:251` (`notImplemented`), `:436` (`ListControllerNodes`), `:669` (`DecommissionNode`); `fabricui/problem.go:93-111` (`terms_required`: a problem with members); `fabricui/handler.go:436-443` (`apiHeaders`); `fabricui/handlers_controller_test.go:195` (`TestControllerAdminMatchesTheContract`), `:260` (`TestController_SecretsStayOutOfLogsAndEvents`); `fabricui/routes_contract_test.go:73,95` (the body table, strict bodies); `fabricui/headers_test.go:42`; `fabricui/status_guard_test.go:18`; `fabricui/undeclared_status_test.go:84`; `fabricui/gofabric/gofabric.go:64` (`ControllerOptions`), `:202` (`Capabilities`), `:218-242` (`Nodes`, sorted), `lookup` below it, `:355` (`Pair`), `:609` (`PairDiscovered`), `:623-632` (`Decommission`), `:766` (`nodeError`); `fabricui/memory_controller.go:175-179,237` (`NoDiagnostics`, `Capabilities`); `contract/openapi.yaml:1287-1317` (`/controller/nodes/{id}`), `:1909-1962` (`/controller/credentials`, the write-only secret route), `:2216-2265` (the response components, `Problem`), `:3033` (`ControllerCapabilities`), `:3087` (`ControllerNodeSummary`); `ui/scripts/contract-guard.mjs:71-91`; `docs/contract.md` "Admin rights" (:402), "Problems" (:516, one `### problem-<code>` per code), "Versioning" (:612), "Identifiers, secrets and events of the controller family" (:967); `ui/src/lib/qr.ts` (72 lines; `renderQrSvg` :46), `ui/src/lib/scan/index.ts` (94; `isPayload` :10, `startScan` :51), `ui/src/lib/pages/controller/QrScan.svelte` (91), `ui/src/lib/pages/controller/Nodes.svelte:99-107` (`decommission`), `ui/src/shell/App.svelte:68,156` (`controllerPages`, `controllerTabs`), `ui/src/lib/context.ts:36-64` (routes), `ui/src/lib/i18n/{en,de}.json`.

## go-fabric facts this rests on (at `v0.24.1`)
1. `controller.Controller.Inventory() []InventoryEntry` (`controller/inventory.go:80-122`): every commissioned node in the order of `Controller.Nodes` (the registry's order, `controller/controller.go:1005-1013`; not sorted); nil once Close has begun; a node without a peer address is left out (`:88-91`), and so, with `Nodes`, a node that failed to load (`controller/client/registry.go:682-690`); a stored code that no longer decodes is listed as none (`:113-118`).
2. `InventoryEntry` (`controller/inventory.go:27-55`): `Node`, `NodeID uint64`, `Label` (NodeLabel), `VendorName`, `ProductName`, `SerialNumber` (Basic Information, "" when the image holds none), `VendorID`, `ProductID` (0 unknown), `HasCode`, `QRCode`, `ManualCode` (`client.SecretCode`), `Source`, `Checks`, `StoredAt`. `QRCode` is zero when the stored code is a manual pairing code: no QR code is made up (`:43-45`, `controller/client/onboarding.go:207-219`). A stored QR code yields both forms; the manual form is 11 digits, or 21 when the code's commissioning flow is not Standard (`onboarding.go:224-233`).
3. `client.SecretCode` (`onboarding.go:122-179`): redacted (`"redacted"`) under every fmt verb, slog, encoding/json and any `encoding.TextMarshaler`; `Reveal()` is the one way to the text; no `UnmarshalText`; not comparable. `InventoryEntry` and `OnboardingCode` redact themselves the same way (`inventory.go:57-70`, `onboarding.go:198-205`).
4. `client.OnboardingSource` (`onboarding.go:94-105`): `OnboardingPaired = 1` (the device accepted the passcode), `OnboardingAttached = 2` (added afterwards, passcode unproven). `client.OnboardingChecks` (`:107-120`), a bit set: `OnboardingCheckedVendor = 1`, `OnboardingCheckedDiscriminator = 2`, `OnboardingCheckedProduct = 4`; 0 for a paired code and for an attached one whose node had nothing to compare with (`:192-195`).
5. `client.Node.SetOnboardingCode(code string) (OnboardingCode, error)` (`onboarding.go:285-326`): trims, validates with the decoders `Controller.Pair` uses, replaces a stored code. Errors: `ErrOnboardingCode` (`:51-54`; unreadable, bad checksum or passcode — wraps the `setup` error with a second `%w`, `:253,259`, so `Unwrap() []error` reaches it), `ErrNotCommissioned` (`controller/client/registry.go:84`), `ErrOnboardingMismatch` (`:55-59`; nothing stored), a store error, `ErrClosed` for a node erased meanwhile (`:432-434`). The very code the node was paired with stays `OnboardingPaired` (`:318-320`). No check proves the passcode or the unit (`:299-304`).
6. The mismatch is a **typed error** (go-fabric#537, v0.23.0): `*client.OnboardingMismatchError{Field OnboardingChecks; Code uint16; Node uint16}` (`onboarding.go:62-77`), for which `errors.Is(err, ErrOnboardingMismatch)` holds (`:92`). `Field` is one of the three `OnboardingChecked*` bits; `Code` is the code's value, `Node` the node's — vendor id, product id, or the discriminator the node advertised. Its text is the old one (`:80-89`); "None of its values is a secret" (`:64-65`). The discriminator check runs only for a record last seen commissionable in Basic mode (`:363`); for a manual code its four bits are compared with the upper four of the advertised one, and **both values of the error are then the 4-bit form** (`:364-369`, `:73-76`). The error does not say which kind of code it was.
7. A manual code's non-digit characters are ignored (`secure/setup/decode.go:373,381-387`); a QR code must start with `MT:` (`onboarding.go:250`). `setup`'s errors name the rule and never the passcode (`secure/setup/error_text_test.go:12-18`, `decode.go:193,428`); they can name a vendor or product id and one offending character with its position (`decode.go:103,227`).
8. `client.Node.ClearOnboardingCode() error` (`onboarding.go:328-344`): removes the stored code, also an unreadable one; nil on an erased node.
9. **The code does not outlive the node.** `Registry.Decommission` (`fabricops.go:294-306`) ends in `deleteNode` (`registry.go:512-549`), which calls `Node.Erase` (`node.go:761-788`): `n.onboarding = nil` (`:769`) and `store.EraseNode` (`:784`). `Registry.Forget` takes the same path (`registry.go:504`). A removal the peer refused keeps the node and so its code. Since v0.23.0 a store that fails to erase the rows fails `Erase`, `Forget` and `Decommission`; the node, closed, stays in the registry for a retry — its code is gone from memory, its row stays until the retry (`node.go:756-760`). This repo's `Decommission` calls `Registry.Decommission` (`fabricui/gofabric/gofabric.go:632`).
10. `controller.Config.KeepOnboardingCodes` (`controller.go:147-158`), off by default: `Pair` keeps the code when the commissioning succeeded — also when the node's start then failed — and the device advertised Basic commissioning mode — not for an Enhanced Commissioning Window, not for a pairing by address, not when `PairOptions.Edit` replaced the passcode (`controller.go:1374-1379`, `fabricops.go:157-167,269-271`, `onboarding.go:399-422`). **`Controller.PairDiscovered` (v0.24.0) keeps none:** it takes a passcode, there is no code (`controller.go:1253-1254`); this repo commissions a device from the discovery list through it (`gofabric.go:609`). The controller exposes no getter for the setting.
11. `docs/security.md:232`: the store holds the code in plain text unless sealed (`Config.Sealer`, v0.23.0, seals the code's row, `:204`; its source, time and checks are not sealed, `:209`); "An API that serves the inventory hands out setup passcodes: give it the authorisation an export of the fabric's keys would need, and call `Reveal()` only where the code is shown or printed".

### Re-check at the tag (done 2026-11-22 at `v0.24.1`, `notes/reads/read-U12-recheck.md`; the list as written on 2026-11-17)
Result: R1, R2, R4 as designed; R3 the typed error (Q1 applies); R5 and R6 as designed with one addition each (facts 9, 10); R7 the pin is `v0.24.1`, already on `main`; R8 as designed in go-fabric, the adapter sorts (decision 7). The implementation does not repeat it unless the pin moves.
- R1. Names and signatures of facts 1, 2, 4, 5, 8: `Controller.Inventory`, every `InventoryEntry` field, the two `OnboardingSource` values, the three `OnboardingChecks` bits and their numeric values, `Node.SetOnboardingCode`, `Node.ClearOnboardingCode`, `Node.OnboardingCode`.
- R2. `SecretCode` still redacts under fmt, slog and JSON and `Reveal()` still returns the text (fact 3).
- R3. The three mismatch texts of fact 6 — or a typed mismatch error if go-fabric took hand-over H1; then decision 6 switches to its fields.
- R4. `ErrOnboardingCode`'s wrapped `setup` texts still carry no passcode (fact 7; `error_text_test.go` present and passing at the tag).
- R5. `Decommission` and `Forget` still erase the code (fact 9), and `SetOnboardingCode` on an erased node still fails rather than stores.
- R6. `Config.KeepOnboardingCodes` keeps its name, default and the three exclusions (fact 10); whether a getter appeared (open question Q3).
- R7. The tag's number and its `CHANGELOG.md` section, for `go.mod`, the changelog line and CLAUDE.md's state sentence; any other facade change between v0.22.2 and the tag that touches `fabricui/gofabric`.
- R8. `Inventory()` still skips nothing but nodes without a peer address, and still returns the entries in `Nodes()` order.

## Decisions (do not reopen)
1. **Contract 0.13.0, additive over 0.12.0; three routes, all `x-admin`, every answer `Cache-Control: no-store` (decision 4).**
   - `GET /controller/inventory` → `OnboardingInventory` `{ entries: OnboardingInventoryEntry[] }`. Admin although it is a GET, as `GET /node/pairing` and `GET /controller/discovery` are: the answer carries setup passcodes. One route, not a public list plus a secret one — the non-secret columns are already on `GET /controller/nodes`.
   - `PUT /controller/nodes/{id}/onboarding-code`, body `OnboardingCodeRequest` `{ code: string }` → 200 with the node's `OnboardingInventoryEntry` (the row the page replaces; it shows `source` and `checks` of what was stored). The body is strict as every body since 0.10.0: the route joins `controllerBodies`; 413 and 415 are declared.
   - `DELETE /controller/nodes/{id}/onboarding-code` → 204, also when the node had none. In scope because a stored secret needs a way out that is not "decommission the node", and a wrong attached code is only otherwise replaceable.
   - The code travels in a request body or a response body, never in a path or a query.
   - Schema names carry `Onboarding`: "inventory" already names the endpoint and cluster inventory of contract 0.4.0 (`EndpointDetail`, `ClusterDetail`, the node page's Inventory tab).
2. **Problems.** Two new codes in `Problem.code` (an enum value added, as `node_limit`, `terms_required` and `icd_multi_admin` were):
   - `onboarding_code_invalid` — 400, port `*OnboardingInvalid` (`errors.Is` `ErrOnboardingCode`). Title "Not an onboarding code"; `detail` is its `Reason`, go-fabric's `setup` text (fact 7). Distinct from `bad_request` (a malformed body, an empty `code`, a bad `{id}`), which the page words differently.
   - `onboarding_code_mismatch` — 409, port `*OnboardingMismatch` (`errors.Is` `ErrOnboardingMismatch`). Title "The code names another device". Members on the problem, as `terms_required` has its own (`fabricui/problem.go:93-111`): `field` (`vendor` | `product` | `discriminator`), `code_value`, `node_value` (Q1). `detail` is a sentence the handler builds from the three (English, e.g. `the code's vendor id is 0x1234, the node's is 0xFFF1`), not go-fabric's text. Nothing was stored. For a manual code's discriminator both values are the 4-bit form, as compared (fact 6); the contract says so.
   - Unknown node: 404 `not_found` (`ErrNodeNotFound`, existing). Node not commissioned or erased meanwhile: 409 `conflict` (`ErrConflict`, existing) with a fixed detail. Source without the capability: 501 `not_implemented` on all three, through `notImplemented` as the diagnostics routes do (`controllerError` alone answers 500 for `ErrNotImplemented`, decision 79). Store failure: 500 `internal`, no detail, logged with the error's text (a store error carries no code: `SecretCode` redacts, and the code is never an argument of the wrapped error — R2).
   - **The three handlers map errors themselves, closed:** `ErrNodeNotFound` 404; `*OnboardingInvalid` 400; `*OnboardingMismatch` 409; `ErrConflict` 409 on the PUT only; `ErrNotImplemented` 501; everything else — also a host source's `ErrNodeFailed`, or `ErrConflict` on the DELETE — 500, logged. No operation answers a status it does not declare (decisions 47, 79, 82; UR-P4-api-391).
3. **Is the differing value of a mismatch a secret? No.** Vendor and product id are the node's public identity (already on `GET /controller/nodes`); the discriminator is what a commissionable device advertises to everyone in range (go-fabric says the same of its typed error, fact 6). None of them is the passcode, and none narrows it. They may stand in the problem's members and `detail` on this admin route and on the page. They are still not logged: the handler logs nothing for a 4xx of these routes.
4. **`Cache-Control: no-store` on every answer of the three routes**, problems and the 403 included. Rebased: the handler sets it on every answer of the API since 0.10.0 — `apiHeaders` before any route writes (`fabricui/handler.go:436-443`), the host check's own 403 (`fabricui/host.go:91`) — and `TestAPI_EveryRouteAnswersWithTheSecurityHeaders` scans every operation of the contract, so the three new ones are in it once they are in the contract. So: no `x-no-store` extension, no `controllerNoStore` table, nothing added to `controllerGate`. G3 names the secret-carrying answers.
5. **Capability and opt-in.** `ControllerCapabilities.onboarding_codes` (bool, required in 0.13.0 with the fixtures updated, as `diagnostics` was in 0.9.0; the contract guard accepts a field new and required on a response-only schema, `ui/scripts/contract-guard.mjs:81-90`, so no `info.x-accepted-breaks` entry). The `gofabric` adapter reports it from a new `ControllerOptions.OnboardingCodes bool`, **false by default**: a host that upgrades does not start handing out passcodes; serving them is the host's stated decision (fact 11). False: the three routes answer 501 with the fixed sentence, the tab is hidden, `has_onboarding_code` is absent. The handler reports it false, whatever the source says, when `Options.Admin` is nil (Q2). A non-admin reads the boolean on `GET /controller/status` like the other capabilities (decision 57); it is no secret.
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

   // OnboardingInvalid is a code that is no onboarding code; Reason says
   // why and never holds the code. errors.Is(err, ErrOnboardingCode) holds.
   type OnboardingInvalid struct{ Reason string }

   // OnboardingMismatch is a readable code that names another device (Q1).
   // errors.Is(err, ErrOnboardingMismatch) holds. No value is a secret.
   type OnboardingMismatch struct {
       Field      string // "vendor" | "product" | "discriminator"
       Code, Node uint16 // the two values as compared (fact 6)
   }
   // both: Error() string, Is(target error) bool, pointer receivers
   ```
   `SetOnboardingCode` takes the code as a plain string, as `Pair` does. The handler checks `Capabilities().OnboardingCodes` first (501) and is the one place outside the adapter that calls `Reveal()`: where it fills the response. `OnboardingEntry` gets `String`/`GoString`/`LogValue` that name the node id and "code redacted"/"no code", as go-fabric's entry does. The two typed port errors carry what the problem shows (`Reason`; `Field`, `Code`, `Node`); the handler reads them with `errors.As` and cuts no error text (Q1).
7. **Adapter (`fabricui/gofabric`, a new `inventory.go`).** `OnboardingInventory`: `ctrl.Inventory()` mapped one to one, then sorted by node id as the adapter's `Nodes()` is (`gofabric.go:218-242`; R8); `NewSecretText(e.QRCode.Reveal())` in one expression, the revealed string bound to no variable; `Source` 1→`paired`, 2→`attached`; `Checks` bit by bit. `SetOnboardingCode`: `lookup(id)` (404), `n.n.SetOnboardingCode(code)`, then the node's entry from a fresh `ctrl.Inventory()` (by node id; missing → `ErrNodeNotFound`). Errors by `errors.Is`, not through `nodeError` (whose default is 502 `node_error`, wrong for a local store): `*client.OnboardingMismatchError` (`errors.As`) → `*OnboardingMismatch`, `Field` by the bit (`OnboardingCheckedVendor` → `vendor`, `OnboardingCheckedProduct` → `product`, `OnboardingCheckedDiscriminator` → `discriminator`), `Code` and `Node` as they are; `client.ErrOnboardingCode` → `*OnboardingInvalid` whose `Reason` is the text of the wrapped `setup` error, the element of `Unwrap() []error` that is not the sentinel (fact 5) — no text is cut; `client.ErrNotCommissioned`, `client.ErrClosed` → `ErrConflict`; anything else unwrapped (500). `ClearOnboardingCode`: `lookup`, `n.n.ClearOnboardingCode()`. With `ControllerOptions.OnboardingCodes` false the three methods are not reached (the handler's 501); they answer `nil` / `ErrNotImplemented` anyway.
8. **The removal warning needs no secret route.** `ControllerNodeSummary.has_onboarding_code` (optional boolean, 0.13.0; existence only, no secret): the `GET /controller/nodes` handler derives it from `OnboardingInventory()`'s `HasCode` when the capability is on, and omits the field otherwise. No new method on `ControllerNode`.
9. **Memory port (`fabricui/memory_controller.go`).** A map node id → stored code, `NoOnboardingCodes` to switch the capability off (as `NoDiagnostics`), `Capabilities().OnboardingCodes = !m.NoOnboardingCodes`. `SetOnboardingCode` there: trimmed text starting `MT:` and matching the scan module's payload shape is a QR code (manual code = a fixed test value held beside it), 11 or 21 digits after dropping non-digits a manual code (no QR), anything else `&OnboardingInvalid{Reason: "Invalid pairing code"}`; a settable `OnboardingMismatch *OnboardingMismatch` makes the next attach answer it. `Decommission` deletes the code. It validates shape only — it is the test port, not a second decoder; the real decoders are tested through the adapter's e2e.
10. **Fixtures**, recorded from the memory port, codes are the spec's public test values (`MT:Y.K9042C00KA0648G00`, `34970112332`), never a real device's: `controller-inventory.200.json` (three nodes: a paired QR code with both forms; an attached 11-digit manual code, `checks.discriminator` true and `checks.vendor`/`checks.product` false — an 11-digit code carries no vendor or product id —, no `qr_code`; a node without a code), `controller-inventory.403.json`, `controller-inventory.501.json`, `controller-node-onboarding-code-set.200.json`, `.400-invalid.json`, `.409-mismatch.json` (with `field`, `code_value`, `node_value`), `.404.json`, `controller-node-onboarding-code-clear.204.json`; `controller-status.200.json` and `controller-status.200-ble.json` gain the capability, `controller-nodes.200.json` one `has_onboarding_code`. `contract-check` on both sides.
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
      The dialog states: a live camera needs a secure page (HTTPS or localhost); a photograph works on any page. A refusal stays in the dialog: the localized title by problem code and, under it, for a mismatch the localized sentence by `field` with `code_value` and `node_value` (vendor and product id as `0x` and four uppercase hex digits, the discriminator decimal; Q1), for an invalid code the module's reason (`detail`, English, as go-fabric words it). Success: the row is replaced, a toast without the code.
    - **Decoder and renderer:** `jsqr` 1.4.0 and `qrcode` 1.5.4, both pinned in `ui/package.json:70-71`, listed in `THIRD-PARTY-NOTICES.md:25-26`, used by `ui/src/lib/scan/index.ts:7,40` and `ui/src/lib/qr.ts:7,46`. No new dependency; `decodeImage` stays in the `scan` chunk (dynamic import).
    - **Removal warning:** `Nodes.svelte:99-107` `decommission()` — the confirm body gains a second sentence when `n.has_onboarding_code`. The Inventory page offers no decommission.
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
      - `controller.inventory.mismatch_vendor`: The code's vendor id is {code}, the device's is {node}. / Die Hersteller-ID des Codes ist {code}, die des Geräts ist {node}.
      - `controller.inventory.mismatch_product`: The code's product id is {code}, the device's is {node}. / Die Produkt-ID des Codes ist {code}, die des Geräts ist {node}.
      - `controller.inventory.mismatch_discriminator`: The code's discriminator is {code}, the device advertised {node}. / Der Discriminator des Codes ist {code}, das Gerät hat {node} gemeldet.
      - `controller.inventory.remove_confirm`: Remove the stored code of {name}? / Den gespeicherten Code von {name} entfernen?
      - `controller.inventory.removed`: The code is removed. / Der Code ist entfernt.
      - `controller.inventory.print`: Print labels / Etiketten drucken
      - `controller.inventory.print_empty`: No node has a code to print. / Kein Knoten hat einen Code zum Drucken.
      - `controller.inventory.admin_required`: The inventory shows secrets and needs admin rights. / Das Inventar zeigt Geheimnisse und braucht Admin-Rechte.
      - `controller.inventory.not_served`: This host does not serve the inventory. / Dieser Host stellt das Inventar nicht bereit.
      - `controller.nodes.remove_confirm_code`: Its stored onboarding code is deleted with it. / Der gespeicherte Onboarding-Code wird mit gelöscht.
12. **Export: printing is the export of this wave; a file export is left out.** The issue states a constraint ("an export is an explicit action"), not a request for a file. Print and the per-row copy are explicit actions the user starts. A CSV or JSON download would put every passcode into a file outside the database's protection (fact 11) for no stated need; it is a later wave's decision with its own route.
13. **Release 0.13.0:** `go.mod` unchanged (go-fabric v0.24.1); `changelog.md` `[0.13.0]`; version files; `contract/previous` = the 0.12.0 contract; `info.description`'s sentence on additive contracts names 0.13.0; `docs/contract.md` a section "Onboarding inventory (v0.13.0)" (routes, the mismatch members, the secrets bullet extended to the three routes) and under "Problems" `### problem-onboarding_code_invalid` and `### problem-onboarding_code_mismatch`; `docs/embedding.md` the three source methods, the two port error types, `ControllerOptions.OnboardingCodes` and the hosts' obligations below; CLAUDE.md's state sentence and the Security rule ("secrets (passcodes, onboarding codes)"); README status.

## Proposed contract (`contract/openapi.yaml`, version 0.13.0; in 0.12.0's form)
```yaml
  /controller/inventory:
    get:
      tags: [controller]
      operationId: getOnboardingInventory
      summary: Every commissioned node with the onboarding code kept for it (0.13.0). The codes are secrets.
      x-admin: true
      responses:
        "200":
          description: The inventory, by node id as `GET /controller/nodes`.
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
      summary: Attach the code printed on the device to a commissioned node (0.13.0); it replaces a stored one. The code is validated and compared with what the node is known to be; its passcode stays unproven.
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
        "409": { $ref: "#/components/responses/Problem" }   # onboarding_code_mismatch (with field, code_value, node_value), conflict
        "413": { $ref: "#/components/responses/Problem" }
        "415": { $ref: "#/components/responses/Problem" }
        "500": { $ref: "#/components/responses/Problem" }
        "501": { $ref: "#/components/responses/Problem" }
    delete:
      tags: [controller]
      operationId: clearNodeOnboardingCode
      x-admin: true
      summary: Remove the node's stored onboarding code (0.13.0); 204 also when it had none.
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
      description: "The commissioned nodes with their kept onboarding codes (0.13.0, go-fabric `controller.Controller.Inventory`)."
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

# additions to existing schemas (as they are in 0.12.0)
#   Problem.code enum (18 values today, the last `icd_multi_admin`): + onboarding_code_invalid, onboarding_code_mismatch
#   Problem, beside `version`/`acknowledgements`/`deadline` of terms_required, three optional members:
#     field:      { type: string, enum: [vendor, product, discriminator], description: "onboarding_code_mismatch (0.13.0): the identity field in which the code differs from the node (go-fabric client.OnboardingMismatchError.Field)." }
#     code_value: { type: integer, minimum: 0, maximum: 65535, description: "onboarding_code_mismatch (0.13.0): the code's value of the field. For the discriminator of a manual pairing code, its 4 bits." }
#     node_value: { type: integer, minimum: 0, maximum: 65535, description: "onboarding_code_mismatch (0.13.0): the node's value — its vendor or product id, or the discriminator it advertised, cut to its upper 4 bits when the code is a manual one." }
#   ControllerCapabilities (required today: ota, credentials, ble, diagnostics): + onboarding_codes (boolean, required): "The source serves the onboarding inventory and the host has an admin check (0.13.0); false: the three routes answer 501."
#   ControllerNodeSummary: + has_onboarding_code (boolean, optional): "An onboarding code is stored for the node (0.13.0); absent when the inventory is not served. Decommissioning the node deletes it."
#   info.version 0.13.0; info.description: "Contracts 0.11.0, 0.12.0 and 0.13.0 are additive over their predecessors."
#   No x-no-store, no info.x-accepted-breaks. Then `make generate`, `make gen-types`.
```

## Hand-over to go-fabric (done)
- **H1, taken:** SukramJ/go-fabric#537, in v0.23.0 — `*client.OnboardingMismatchError{Field, Code, Node}` (fact 6). The contract carries `field`, `code_value`, `node_value`; nothing cuts an error message.

## Guards
- `export PATH=$PATH:$HOME/go/bin`. After the contract change `make generate` and `make gen-types`; after a change under `ui/src` `make web` in the same commit. Then the pipeline's set (CLAUDE.md "Token use"), in the background with a summary log: `make ci` (build, vet, test, lint, generate-check, web-check, dist-check), `make race`, `make cover-check`, `make vuln licenses secrets`, `GOARCH=386 make build`, `make test-386`, `make e2e`, and after `make web-install` `make notices-check`, `make version-check`, `make contract-check`, `make npm-audit`, `make e2e-web`.
- **Existing guards the three routes run into** (each is red until the wave feeds it):
  - the status guard (`fabricui/status_guard_test.go:18`, decisions 47, 82): every status a test of the package gets from the three routes is declared — hence the closed error mapping of decision 2;
  - `TestControllerAdminMatchesTheContract`: the three routes in `controllerAdmin`;
  - `TestControllerBodiesMatchTheContract` and `TestBodies_AreStrict` (`fabricui/routes_contract_test.go:73,95`): the PUT in `controllerBodies`; an unknown member or trailing data is 400 `bad_request`;
  - `TestController_NotImplementedAndNotReadOnlyWhereDeclared` (`fabricui/undeclared_status_test.go:84`): 501 only where declared — the three declare it;
  - `TestAPI_EveryRouteAnswersWithTheSecurityHeaders` and `TestEveryPathAnswers405WithItsMethods`: by the contract scan, nothing to add;
  - the contract guard (`ui/scripts/contract-guard.mjs`, in `make web-check` and `make contract-check`): 0.13.0 against `contract/previous` 0.12.0 with no finding and no `info.x-accepted-breaks`; every new fixture used on both sides;
  - the i18n guard (`ui/src/lib/i18n/catalogs.test.ts`, `coverage.test.ts`): every new key in en and de, every key the page translates in the catalogues;
  - the import guard (`internal/arch/arch_test.go`): `fabricui` imports no `controller/client` — hence the port's own types; `fabricui/gofabric/inventory.go` imports only `controller` and `controller/client`;
  - SA1019 (decision 62): no identifier of this path is deprecated at v0.24.1;
  - `make cover-check`: the floors of `fabricui` and `fabricui/gofabric` with the new files;
  - `make test-386`: the new integers are 16-bit, no `int64` format is needed.
- **New, each with its bite line:**
  - G1 `TestController_SecretsStayOutOfLogsAndEvents` extended: after `GET /controller/inventory`, a successful attach, an invalid one, a mismatch and a store failure, neither the captured log nor any stream event contains the QR text, the manual code or the request's code. Bites when a handler logs a body or an error that wraps the code.
  - G2 `TestSecretTextRedacts`: `%v`, `%+v`, `%#v`, `%s` of a `SecretText`, of an `OnboardingEntry` and of a struct holding one in an unexported field; `slog` text and JSON handlers; `json.Marshal`. Bites when the port type loses a method.
  - G3 a handler test: `Cache-Control: no-store` on the 200 of the GET and of the PUT, and on the 403, 400, 409 and 501 of the three routes (decision 4; the header comes from `apiHeaders`). Bites when a route is mounted outside `apiHeaders`.
  - G4 Q2: with `Options.Admin` nil (token mode, a valid token) the three routes answer 501 and `GET /controller/status` reports `onboarding_codes: false`, though the source serves the inventory; without admin rights 403, fixture `controller-inventory.403.json`. Bites when the handler trusts the source's capability alone.
  - G5 adapter e2e (`fabricui/gofabric/e2e_test.go` pattern, `KeepOnboardingCodes` on): a node paired by code is `paired` with both codes; an attach of another product's code answers `onboarding_code_mismatch` with `field: product` and both ids; a text that is no code answers `onboarding_code_invalid` with go-fabric's reason and without the text; the inventory is in the order of `GET /controller/nodes`; after `DELETE /controller/nodes/{id}` the inventory no longer lists the node.
  - G6 UI: `OnboardingInventory.test.ts` — no QR image for an entry without `qr_code`; the insecure-context sentence; a mismatch's sentence from the three members in the page's language, an invalid code's `detail`; nothing written to `localStorage`. `scan.test.ts` — `decodeImage` returns null for a non-payload QR code. `Nodes.test.ts` — the second sentence of the confirm only with `has_onboarding_code`. The Playwright smoke (`ui/e2e/smoke.spec.ts`, `make e2e-web`) opens the tab.
  - G7 a handler test of the closed mapping: a memory source that answers `ErrNodeFailed` on the PUT and `ErrConflict` on the DELETE gets 500 — the status guard would fail on a 502 or an undeclared 409.

## What a host must do (go-fabric-server wires it; `docs/embedding.md`)
1. Set `gofabric.ControllerOptions.OnboardingCodes = true` to serve the inventory at all — and only behind an access control fit for an export of the fabric's keys (fact 11). In host mode `Options.Admin` decides who reads the codes. **Token mode is no access control**: the token is served by `GET /controller/status` to whoever reaches the handler; it stops a cross-origin page, not a person on the network. A host in token mode that sets the option hands the passcodes to everyone who can open the UI, and must restrict that itself (bind address, proxy authentication).
2. Set `controller.Config.KeepOnboardingCodes` if codes of newly paired devices are to be kept; without it the inventory holds only attached codes. A device commissioned from the discovery list by its passcode has no code to keep (fact 10); its code is attached afterwards. Configure a `Sealer` if the store's secrets are to be sealed at rest.
3. Not log request or response bodies of the three routes in its own middleware or proxy, and not cache them; the handler sends `Cache-Control: no-store` on every answer, a proxy has to honour it.
4. A host with its own `ControllerSource` implements the three methods (or returns `OnboardingCodes: false` and `ErrNotImplemented`) — the interface grows, as it did in 0.9.0. Its refusals are `*OnboardingInvalid`, `*OnboardingMismatch`, `ErrNodeNotFound` and `ErrConflict`; any other error is answered 500.
5. Serve the UI over HTTPS (or localhost) for the live camera; the photograph path and typing work without.

## Left out
A file export (decision 12). An event on the controller stream for a code stored or removed (the page refetches after its own write; a second browser sees it on its next load). Bulk attach. Editing NodeLabel from this page. A setting in the UI for `KeepOnboardingCodes` (host configuration).

## Owner decisions (2026-11-17)
- **Q1 → H1 is asked** (SukramJ/go-fabric#537). go-fabric gets the typed mismatch error before its tag; decision 6 and the mismatch problem switch to its fields (`field`, `code_value`, `node_value`), and the page words the reason in the user's language. No cutting of error strings. If the tag comes without it, the wave stops at R3 and goes back to planning. *2026-11-22: the tag has it (R3); the switch is made in this brief.*
- **Q2 → refuse without `Options.Admin`.** The three routes answer 501 when `Options.Admin` is nil, whatever the capability says; token mode never serves a passcode. A handler test holds it; `onboarding_codes` in the capabilities is false in that case.
- **Q3 → no.** The page does not say whether newly paired codes are kept (no getter; host configuration).
- **Q4 → wrapping grid** on the browser's paper size (A4), no label-printer mode.
- **Q5 → keep** `DELETE …/onboarding-code` and "Remove code".
- **Q6 → behind "Show code"** per row; the print carries every code.

## Review registers this wave touches (`notes/review/2026-11-17/`)
- **UR-P4-api-391** (`phase-4r2-register.md`, open note, go-fabric-ui#85): a host's own `ControllerSource` can answer a status an operation does not declare. The port grows by three methods; decision 2's closed mapping and G7 keep the three new operations out of that note. The note itself stays open for the older operations.
- **UR-P4-err-404** (#85): `PairDiscovered`'s refusals — not touched; the same path is why a discovered device has no kept code (fact 10).
- **UR-P4-ui-396, UR-P4-test-403** (#85): not touched.
- **Decisions 35, 57:** the three routes are `x-admin` like the rest of the controller family; the capability boolean stays readable on the open `GET /controller/status`.
- **Decisions 41, 45, 29:** the controller stream is `x-admin`, read through `fetch`, open only while a controller page is shown — the page's refetch on `node_added` / `node_removed` uses the store's stream as it is, no second one.
- **Decisions 31, 77:** badges, chips and the secondary sentences use the `--fui-*-text` tokens (4.5:1, the ratio guard computes it); the print block is black on white.
- **Decision 38, 43:** the PUT goes through the client's request helper (content type, admin headers) and the strict body table.
- **Decision 62:** no deprecated go-fabric identifier. go-fabric v0.24.1 deprecates the module path in favour of `thingfabric.dev/go-fabric` (from v0.25.0); that move is not this wave.

## Report (fixed shape, short)
Branch, commits, the re-check's result by reference (`notes/reads/read-U12-recheck.md`; again only if the pin moved), guards with HEAD, differences from this brief, left out, questions. Close nothing; the session closes #68 with the PR.
