# Review phase 0 — fix wave report

Brief: `notes/briefs/2026-11-17-review-p0-fix.md`. Branch `review/p0`,
start `bc19f52`. One agent (Opus), Go and web. No push, no pull request,
no tag. Neither `contract/openapi.yaml` nor `internal/api` changed.

## Per ID

### UR-P0-ui-30 — malformed `%` in the hash

- Commit `3de7c98` shell: a malformed % in the hash falls back to the default route.
- Red: `ui/src/shell/route.test.ts` "falls back to the default node's
  status for a malformed % sequence" (`URIError: URI malformed`).
- Change: `route.ts` decodes the key in `try`; a failure returns the
  default node's status. `make web` in the same commit.
- Deviation: none.

### UR-P0-sup-02 — `make licenses` passes an unclassified licence

- Commit `43735ec` make licenses: an unclassified licence fails, a commented allow-list.
- Red: `make licenses` with `unknown` added to `--disallowed_types` and no
  list: `Unknown license type found for library modernc.org/mathutil`,
  exit 2.
- Change: `unknown` disallowed; `LICENSES_CHECKED` in the `Makefile`
  (passed as `--ignore`), commented with version, licence (BSD-3-Clause,
  three clauses incl. "Neither the name") and where it was read (module
  cache, `modernc.org/mathutil@v1.7.1/LICENSE`). The CI job calls
  `make licenses` (`.github/workflows/ci.yml:84`), so it needs no change.
- Deviation: none.

### UR-P0-test-01 — four untested error branches

- Commit `2fd3a9a` fabricui: tests for four untested error branches.
- Test: `fabricui/error_branches_test.go` — 503 `service_unready` for
  `ErrCommissioningWindowNotConfigured` (ConfiguredWindow mode); 503
  `service_unready` "Node registry not available" on PUT
  `/nodes/{node}/label` and DELETE `/nodes/{node}`; 400 `bad_request`
  (malformed, unknown field) and 413 `payload_too_large` on POST `/nodes`,
  PUT `/nodes/{node}/label`, PUT `/operation/log/level`.
- Red: not applicable — the tests cover existing behaviour and passed on
  first run. Reach confirmed with a cover profile: every branch named in
  the brief has count ≥ 1.
- Deviation: no red step (none possible); no changelog line (tests only).

### UR-P0-err-01 — the eight `req.Body == nil` branches

- No commit. A request cannot make them true: every strict wrapper in
  `internal/api/api.gen.go` (13 sites, e.g. `:5835-5840` for
  CommissionNode) calls `runtime.BindJSONBody` and answers a missing or
  invalid body itself through `RequestErrorHandlerFunc` → `badBody`, then
  sets `request.Body = &body` unconditionally. They are reachable only by
  calling the `controllerServer` methods directly, which is not a
  request. Left as they are, as the brief says.

### UR-P0-web-33 — SSE event type unchecked

- Commit `8d0f7ac` stream: drop an event whose type is not [a-z_]+.
- Red: `TestStream_DropsAnEventWithAnInvalidType` (`misc_test.go`) — a type
  with a line break forged `event: fabric_removed`.
- Change: `serveSSE` drops an event whose type is not a non-empty
  `[a-z_]+` and logs `fabricui.stream.invalid_type` with the type only
  (slog quotes it), never the payload. Covers all three streams (node,
  operation log, controller). Every type the module publishes passes.
- Deviation: logged, not counted.

### UR-P0-api-32 — a failed `Mount` changed the host's mux

- Commits `37bde08` (and the lint fix `1c04fcf`) fabricui: Mount registers one pattern after a lookup, a failed Mount leaves the mux untouched.
- Red: `TestMount_AFailedMountLeavesTheMuxUntouched` — with the host
  holding `GET /ui-api/operation/info`, the routes registered before the
  panic answered 200 after the failed `Mount`.
- Change: `handler.routes()` builds every route on a mux of the API's own
  and returns its patterns; `mount` looks each pattern up on the host's
  mux (`ServeMux.Handler` with a synthetic request, wildcards as plain
  segments) and refuses one the host already serves under the prefix,
  then registers the single pattern `Prefix + "/"`, recovering its panic.
  Either the one registration happens or nothing does; the error says
  "nothing was registered". `Handler` uses the same path. Doc comment on
  `Mount` and `docs/embedding.md` ("Mounting the API alone") say so.
- Deviation: the brief's first option (scratch mux first, then the host's)
  cannot detect a conflict with the host's patterns, because `ServeMux`
  neither lists nor removes patterns; registering one delegating pattern
  is the form of it that is atomic. Behaviour change for hosts: a route a
  host registers under the prefix *after* `Mount` now shadows the API's
  route instead of panicking (documented).

### UR-P0-leak-25 — no `Cache-Control` on JSON answers

- Commit `73157ee` fabricui: no-store, nosniff and no-referrer on every API answer.
- Red: `TestAPI_EveryRouteAnswersWithTheSecurityHeaders`
  (`fabricui/headers_test.go`) — every operation of the contract (scanned
  from `contract/openapi.yaml`, ≥ 80) plus two unknown paths, with node,
  controller, registry and operations wired; 293 header misses before.
- Change: `apiHeaders` wraps the API's single pattern and sets
  `Cache-Control: no-store`, `X-Content-Type-Options: nosniff`,
  `Referrer-Policy: no-referrer` before any route writes.
- Deviation: one wrapper instead of the brief's sites (`writeJSON`,
  `writeProblemBody`, the two hand-set sites, `controllerGate`); it covers
  all of them, the streams and the generated routes, and a new site
  cannot miss it.

### UR-P0-web-23, UR-P0-web-24 — CSP, framing, headers on the bundle

- Commits `73157ee` (API half of web-24, above) and `65a9c9e` spa: CSP with
  the pre-paint script's hash, frame-ancestors and Options.FrameAncestors.
- Red: `fabricui/spa_csp_test.go` (did not compile against the old
  `spaHandlerFS`; then the tests decide): `TestSPA_IndexCarriesTheCSP`,
  `TestSPA_FrameAncestors`, `TestSPA_NotBuiltCarriesTheCSP`,
  `TestSPA_EmbeddedIndexScriptHash`, `TestOptions_FrameAncestorsValidated`.
- Policy on the index, every deep path and the 503 page:
  `default-src 'self'; script-src 'self' 'sha256-…'; style-src 'self';
  img-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self';
  frame-ancestors 'self' <FrameAncestors>`. The script hash is derived
  from the embedded `index.html` when the handler is built (every inline
  `<script>` without `src`), and `TestSPA_EmbeddedIndexScriptHash` reads
  the committed `dist/index.html` independently and compares — no
  hand-copied hash. The 503 page's inline `<style>` is admitted by its
  hash the same way. `nosniff` and `no-referrer` on every bundle response.
- `Options.FrameAncestors []string`: each a CSP host-source or
  scheme-source (CSP3 grammar; IPv6 literals are none), checked in
  `newHandler`, so `Handler` panics and `Mount` errs. `docs/embedding.md`
  gains "Response headers and framing" and a framing bullet in "Behind an
  Ingress".
- Chromium: the smoke (`ui/e2e/smoke.spec.ts`) now collects
  `securitypolicyviolation` events from an init script on every page
  smoke and asserts the policy header; two new smokes open the pairing
  window (QR code and countdown ring) and the QR scan with Chromium's fake
  camera (launch flags in `playwright.config.ts`, camera permission in
  the test). No violation on any page. Positive control: with the script
  hash removed, 35 smokes failed on `script-src-elem inline …:9`.
- Deviation, narrower than the brief expected: no `'unsafe-inline'`, no
  `style-src-attr`, no `data:`/`blob:`. The bundle holds no `data:` or
  `blob:` URL; the camera's video takes a `MediaStream` through
  `srcObject`, which CSP does not govern; Svelte sets dynamic styles
  through `style.cssText` (CSSOM, not governed). The one static style
  attribute in a template (`Pair.svelte`, the ring's transition) made
  Chromium log a violation when the template is parsed (probe, though the
  cloned style still applied), so it became `style:transition` — a web
  change with `make web` in the same commit.

## Guards

Run 1 at `65a9c9e`: `make ci` exit 2 — `golangci-lint` found two issues in
this wave's code (gocritic `unnamedResult` on `handler.routes`, `prealloc`
on `styles` in `spa.go`); all other nine guards exit 0. Fixed in `1c04fcf`
fabricui: named results on routes, preallocated styles (lint; UR-P0-api-32,
UR-P0-web-24).

Run 2, the full set again, one guard at a time, HEAD before and after
`1c04fcf` (the only untracked file was this report):

| Guard | exit |
| --- | --- |
| `make ci` | 0 |
| `make race` | 0 |
| `make cover-check` | 0 |
| `make vuln licenses secrets` | 0 |
| `GOARCH=386 make build` | 0 |
| `make notices-check` | 0 |
| `make version-check` | 0 |
| `make contract-check` | 0 |
| `make e2e` | 0 |
| `make e2e-web` | 0 (42 tests, no CSP violation) |

The commit of this report is notes only; no third run.

## Decisions

- `Mount` as one delegating pattern after a lookup (api-32, above).
- One header wrapper for the API instead of per-site headers (leak-25).
- The script hash derived at start *and* guarded by a test against the
  committed bundle; both forms the brief allowed.
- `FrameAncestors` follows the CSP grammar, so `*` and `https:` are
  accepted; the doc comment and `docs/embedding.md` say that they admit
  any parent.
- The fake-camera flags are set for the whole Chromium project; they only
  matter where a test grants the camera.

## Left out

- UR-P0-err-01: no change (unreachable, above).
- Everything under "Not in this wave" in the brief.
- `Cache-Control` set again in `writeJSON` / `writeProblemBody`: covered
  by the wrapper.

## Questions for the owner

1. Should `FrameAncestors` refuse `*` and bare scheme-sources, which undo
   the default? Today they pass, as the grammar allows.
2. `Mount`: is it acceptable that a host route registered under the
   prefix after `Mount` shadows the API's route (before: a panic at the
   host's registration)?
3. Hosts that serve the npm package's pages with their own index
   (`Mount` only) get no CSP from this module. Export the policy builder
   (or document a policy) for them — phase 3 or later?
