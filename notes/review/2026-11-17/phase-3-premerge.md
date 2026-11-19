# Phase 3 pre-merge read — admin and token paths

Reader: Fable. Tree `review/p3` at `7602116`, base `f4b8573`. Read-only:
nothing was run (the local guards were running in the same tree), so every
finding below is **I** (inferred from the lines cited) unless marked M;
"M (read)" means both sides' lines were read, not that a test ran. Each
finding of severity low or above on code still needs its red test, written
by the fix wave in a scratch worktree.

Counts: high 0, medium 2, low 4, note 2.

Answer to the question: on the **server** side the decided shapes are
implemented as decided and I found no gap a non-admin, a foreign page or a
foreign Host can use. Both mediums are on the **web client's token path**
and fail towards less function, with one exception (190: a confirmation
that is skipped for an admin).

## Findings

### UR-P3-web-190 — medium — I — held candidate (class of UR-P0-web-21's token path)

- **Where:** `ui/src/lib/api/client.ts:465-469` (`status()`),
  `:507-511` (`controllerStatus()`); `ui/src/shell/App.svelte:211`
  (`Promise.all([probeNode(), probeController()])`);
  `fabricui/handlers_node.go:146-171`, `fabricui/handlers_controller.go:355-361`;
  consumer `ui/src/lib/pages/controller/node/Acl.svelte:98-101`.
- **Claim:** in token mode the request that *fetches* the token carries
  none, so the server answers it as a non-admin's (decision 57/34, as
  decided): the controller status zeroed (`node_id_hex: ""`,
  `node_count: 0`, `label: ""`), the node status without `config.store`
  and with empty warning messages. The client takes the token from that
  answer and returns the reduced body; nothing repeats the call.
  - The shell's two probes start together, so `controller.loadStatus()`
    always goes out without the token; the controller store keeps the
    zeroed status until the Settings page is opened
    (`Settings.svelte:30`) or a stream gap resyncs. `isAdmin` is true
    (the token is there), so the pages render as an admin's.
  - `Acl.svelte:99-100`: `own = controller.status?.node_id_hex` is `""`,
    and `!own ||` makes `keepsUs` true for any Administer/CASE entry:
    the confirmation before a write that drops this controller's own
    Administer access is skipped. Before this wave the status carried
    the id on the first load.
  - The node family is right in the shell only by order
    (`probeNode` awaits `client.status()`, then `node.loadStatus()` runs
    with the token); a host that uses the stores without the shell's
    probe, and every status after a daemon restart (stale token: the
    answer is the reduced one with the new token), shows no store row
    and empty warnings until the next reload.
  - The fixtures hide it: `controller-status.200*.json` and the node
    status fixtures of an admin are recorded with the token.
- **Proposed fix:** in `status()` and `controllerStatus()`: when the
  answer's `csrf_token` differs from the token the request carried
  (none included), set it and repeat the call once, returning the second
  answer. `Acl.svelte`: an unknown own id asks (fail closed), it does
  not pass. Red tests: a vitest whose fetch answers the reduced body
  without the header and the full one with it (client, controller store,
  ACL save); the Go side needs none.
- **Contract:** no.

### UR-P3-web-191 — medium — I

- **Where:** `ui/src/lib/api/stream.ts:312-317` (403 ends the
  subscription for good), `ui/src/lib/api/client.ts:359-367` (headers
  read at each open from `auth.token`), `:408-410` (only `call` refreshes
  a stale token); `fabricui/stream.go:40` (the node stream's admin test
  runs once, at open); `ui/src/lib/controller-store.svelte.ts:272-279`.
- **Claim:** token mode, the daemon restarts (a new token per process,
  `handler.go:331-335`). The streams drop and reopen with the stale
  token before any call has refreshed it:
  - controller stream and log stream: 403 `forbidden` → `end()`, never
    reopened; `streamDenied` is set. The controller pages stay
    (`isAdmin` is true) but get no event — commissioning progress,
    node changes — until the family is left and entered again or the
    page reloaded. Before the wave the browser's `EventSource`
    reconnected by itself.
  - node stream: 200 as a non-admin's, filtered to the three open types
    for its whole life; the page shows nothing of it. The resync's
    `loadStatus` then gets the new token (and the reduced status of 190),
    but the open stream keeps its filter.
- **Proposed fix:** `FabricClient.openStream` (not the generic
  `fetchEventSource`): on 403 `forbidden` for a request that carried a
  token, refresh the token once (`refreshToken`) and reopen once; a
  second 403, and every `host_not_allowed` / `origin_not_allowed`, ends
  as today. The node store reopens its stream when a status changes the
  token. Red test: vitest with a fetch whose token changes between two
  opens.
- **Contract:** no.

### UR-P3-test-192 — low — M (read)

- **Where:** `fabricui/options.go:390-393` (`return h.hostCheck(mux)`);
  `fabricui/host_test.go:111` and `:251`
  (`mux.Handle("/", h.hostCheck(spaHandler(nil)))`);
  `ui/e2e/smoke.spec.ts` (no test under an unlisted name).
- **Claim:** decision 37's "the bundle of `Handler` / `NewHandler`
  included … and every redirect the mux answers" is held by one line
  that no test reaches: the two bundle tests wrap `spaHandler` by hand,
  and no test sends an unlisted `Host` through `NewHandler` or `Handler`
  (the only uses are `misc_test.go:356`, `spa_csp_test.go:91`, the
  option validation). Dropping the outer wrapper leaves every test
  green and serves the page under any name. The held Playwright tests
  (`held/phase-2-X4.md:356`, `:406`) were not added either.
- **Proposed fix:** `TestNewHandler_UnlistedHostGetsNoPage`: through
  `NewHandler`, in both modes, `GET /`, `/index.html`, one asset, an
  unknown path, `/api/fabric` (the mux's redirect) and `HEAD /` under an
  unlisted Host → 403 `host_not_allowed`, no `<script`, no
  `Location`.
- **Contract:** no.

### UR-P3-test-193 — low — M (read) — the coverage note of the brief

- **Where:** `ui/src/lib/pages/node/NodeConfig.svelte:26` (the store
  row), `ui/src/lib/pages/node/NodeManage.test.ts:30-32`,
  `contract/fixtures/node-status.200-config.json` (recorded without the
  token since `19fb493`).
- **Claim:** the wire side of an admin's `config.store` is tested in Go
  (`fabricui/nonadmin_test.go:168`, `fabricui/registry_test.go:511`).
  The web rendering is not any more: no fixture carries `config.store`,
  and the one test of the block now asserts its absence. A regression
  that drops the row for an admin passes.
- **Proposed fix:** a second fixture `node-status.200-config-admin`
  recorded with the token (used by a Go test and by the web test), and
  the web test asserts the row with `bridge.db`; the existing test keeps
  the non-admin case.
- **Contract:** no (a fixture).

### UR-P3-web-194 — low — I

- **Where:** `ui/src/lib/api/client.ts:597-607` (`isSameOrigin`), `:367`
  (`openStream`).
- **Claim:** UR-P1-web-197's rule "the token only to a base of the
  page's origin" is decided by a pattern, not by the URL parser:
  `if (!absolute) return true` (`:601`) takes every base the pattern
  `^scheme:|^//` does not match as relative. A base that starts with a
  backslash pair or `/\`, or with white space or a control character
  before `//`, is resolved by `fetch` to another host and gets the
  token. The base comes from the embedding host's code, so this is
  robustness of the rule, not a path of a page. Beside it:
  `client.openStream(url)` sends the client's headers, the token
  included, to whatever URL it is handed, not only one under the base.
- **Proposed fix:** with a location, always compare
  `new URL(base, here).origin` with the page's (drop the shortcut);
  `openStream` adds the token only when the URL's origin is the page's.
  Red test: `createClient("\\\\other.example/api")` and
  `"/\\other.example"` send no token.
- **Contract:** no.

### UR-P3-leak-195 — low — I — for the owner (class A-b of decision 34)

- **Where:** `fabricui/handlers_node.go:192-194`
  (`res.AttestationError = err.Error()`), go-fabric
  `node/attestation.go:35-37` at v0.22.2
  (`fmt.Errorf("node: attestation files: %w", err)` over the loader's
  error).
- **Claim:** a non-admin's status still carries the loader's error text
  of the attestation files, which wraps a file error and so, as far as
  read, the files' paths — the class decision 34 withheld
  (`config.store`, the warnings' text). The contract note
  (`notes/briefs/2026-11-18-review-p3-contract.md:76`) names the store,
  the warnings and "counters unchanged"; `attestation_error`,
  `health_note`, `source_note` and `exposure_reason` are not named, so
  they are open by silence, not by decision. Whether
  `LoadChainFiles` puts the path into the text was not read.
- **Proposed fix:** owner decides; the smallest shape is decision 46's:
  for a non-admin `attestation_error` is a fixed text ("the attestation
  files did not load"), the other three stay.
- **Contract:** yes (the field's description), no guard break.

### UR-P3-word-196 — note — the neutral-wording check

Commit subjects and bodies `f4b8573..HEAD`, `contract/openapi.yaml`,
`docs/contract.md`, `docs/embedding.md`, `changelog.md`, `ui/README.md`
and the godocs of the held commits: **clean** — they state the rule, not
its reason. Lines in committed test code and one comment that describe
the scenario; each can be reworded to the rule alone:

| Line | Text |
| --- | --- |
| `fabricui/admin_request_test.go:103-104` | "The admin GETs with an effect, as a page of another site requests them without a header" |
| `fabricui/admin_request_test.go:26,34,42,63` | subtest names "text/plain factory reset / erase / commission / window open" with `Origin: https://other.example`, `Sec-Fetch-Site: cross-site` (`:20`) |
| `fabricui/admin_request_test.go:72` | "form shape against a strict route" |
| `fabricui/host_test.go:36-37` | "a request under a name the handler does not list, sent from a page of that same name" |
| `fabricui/host_test.go:81-83` | "What holds without the Host check: a cross-site admin request without the token is 403, a preflight gets an error …" |
| `fabricui/host_test.go:158` | the sample name `fabric.example.evil` |
| `fabricui/gofabric/gofabric.go:605-607` | "commissioner.Options' own zero value would accept" (says what the adapter did before the fix) |

Not a break of the rule as the brief words it for the public files; the
coordinator decides whether the test comments count.

### UR-P3-bound-197 — note — I

`fabricui/stream.go:86-92`, `fabricui/handler.go:121-135`: the cap of
decision 42 counts the open node stream (no admin rights needed) and the
admin streams in one counter. Requests without admin rights can hold all
64 and an admin's controller or log stream is then 503 until one ends
(the client waits for `Retry-After`). As decided ("per handler"); a
reserve for admin streams would be a new decision. No fix proposed.

## Read without finding

- **Host check (`c3c4534`, `fabricui/host.go`)**: one place; `Mount`
  wraps the API mux (`handler.go:424`), `NewHandler` wraps bundle and
  API (`options.go:393`), `Handler` is `NewHandler`. Nothing is
  registered outside: 404/405 (`noRoute`), the absent controller
  family, HEAD, the four streams and the mux's redirects sit behind it.
  Order a → b → c, then `AllowHost` once; empty name refused before the
  predicate (`host.go:62-65`); zone refused; port, brackets, case, one
  trailing dot. The admission key is unexported, so a host's middleware
  cannot set it. The refusal is problem+json, `no-store`, no token.
  `AllowedHosts` validation as decided; no wildcard, no opt-out.
- **Request header, Origin, 415 (`933137a`)**: `isAdmin` tests the
  header first, then `Admin(r)` (`handler.go:770-773`); no other call of
  `h.admin`. Token mode unchanged, constant-time. Order route → Origin
  → admin test → media type → body in both families
  (`adminOnlyBecause`, `controllerGate`). Origin: a–c, or for a
  predicate admission equality with `r.Host` (host and port,
  case-insensitive); `null`, a path, user info, another scheme refused;
  on an open route a failing Origin makes a non-admin
  (`requestAdmin`). 415 for a body without or with another media type,
  a body of unknown length peeked; bodiless admin POSTs are covered by
  the header and the Origin.
- **Attestation default (`7ad26f4`)**: same condition and place as
  go-fabric `controller/controller.go:845-847` at v0.22.2, before
  `Edit`; `OnFailure = &false` is `attestation.Reject()`
  (`controller/attestation/policy.go:33`). It is the only place the
  adapter builds `commissioner.Options`. Fails closed.
- **Gates (`19fb493`)**: 39 controller + 22×2 node + 3 registry + 3
  operation = 89 admin; 8 open (two statuses, two node streams,
  `GET /nodes`, operation info, log level, controller status). Tables
  held against `x-admin` by tests. HEAD is gated through the GET
  pattern. Node stream: allow-list of three types, host source
  admin-only, `window_opened` carries discriminator and duration only.
  Controller and log streams admin-only, gated before `serveSSE`.
  Status: no `config.store`, warnings as `{code, message: ""}`;
  controller status: `enabled`, `capabilities`, `admin` / token only.
  No passcode, credential name, NOC or ACL on an open route, in an
  event a non-admin receives, or in a problem text of an open route;
  the controller's error details are reachable through admin routes
  only. Fixtures carry the test token only.
- **Web client (`562dffc`)**: header and token on every `call` and on
  the fetch streams; `credentials: "same-origin"`; the token is never
  in a URL or a log; 403 ends a stream, 503 waits at least
  `Retry-After`.
- **Options (`91062a5`)**: the counter is shared by the siblings and
  the registry's nodes, taken after the HEAD answer, released by defer
  on every path; `Done` ends the loop and the deferred unsubscribe
  removes the observers; a negative `MaxStreams` and a non-default
  `Prefix` are errors of `NewHandler`, panics of `Handler`; `Mount`
  keeps `Prefix`.
