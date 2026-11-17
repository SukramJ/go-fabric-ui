# Phase 0 — mechanical evidence

- **Code:** baseline `25f6e7b4`; run on branch `review/p0` at HEAD
  `ee88372f` (three notes-only commits on top). HEAD unchanged before and
  after every run.
- **Toolchain:** go1.27.4 linux/amd64, Node v22.22.1, npm 9.2.0,
  Playwright 1.65.0 (Chromium 1254 already installed, nothing installed).
- **Date:** 2026-11-17. Guards 20:33:26–20:42:51 (sequential, one job at a
  time; `ui/node_modules` was missing, so `make web-install` ran first);
  fuzzing 20:44:10–21:28:48 (eight targets one after another).
- **Machine:** 5 cores, 19 GB.

## Guards

| Guard | Exit | Duration | Summary |
| --- | --- | --- | --- |
| `make web-install` | 0 | 23 s | `npm ci` in `ui/` |
| `make ci` | 0 | 221 s | build, vet; `go test`: ok `fabricui` (1.0 s), `fabricui/gofabric` (2.0 s), `internal/arch`; gofumpt clean; golangci-lint **0 issues**; `go mod tidy -diff` clean, `go mod verify` all verified; generate-check ok; svelte-check **965 files, 0 errors, 0 warnings**; vitest **33 files, 432 tests passed**; dist-check: byte-identical |
| `make race` | 0 | 31 s | no `DATA RACE`; ok `fabricui` (4.0 s), `fabricui/gofabric` (6.9 s), `internal/arch` |
| `make cover-check` | 0 | 17 s | `fabricui` **94.2 %** (floor 88), `fabricui/gofabric` **71.4 %** (floor 50), `internal/api` 42.1 % (floor 0), `script/coverfloor` 0 % (floor 0) — every floor met |
| `make vuln` | 0 | 17 s | govulncheck: **No vulnerabilities found** |
| `make licenses` | 0 | 22 s | go-licenses passed; one `E` line: no licence classified for `modernc.org/mathutil` v1.7.1 (its `LICENSE` is BSD-3-Clause text; reached via go-fabric `store` → `modernc.org/sqlite` → `libc`) — see UR-P0-sup-02 |
| `make secrets` | 0 | 25 s | gitleaks v8.30.1: 71 commits, ~3.45 MB, **no leaks found** |
| `GOARCH=386 make build` | 0 | 126 s | builds |
| `make contract-check` | 0 | 13 s | contract guard keeps every path/method/response/field of `contract/previous`; `TestEveryFixtureIsUsedByTheGoTests` ok; vitest fixtures + guard **2 files, 9 tests passed** |
| `make generate-check` | 0 | 1 s | `internal/api` is the generation of the contract |
| `make notices-check` | 0 | 7 s | 7 bundled packages listed with their licence texts |
| `make version-check` | 0 | 1 s | 0.9.2 |
| `make e2e` | 0 | 5 s | `TestEndToEnd*`: 8 top-level tests pass (`fabricui`, `fabricui/gofabric`) |
| `make e2e-web` | 0 | 51 s | Playwright, 1 worker: **40 passed** (40.8 s) |
| `npm audit --json` | 1 | 3 s | 4 high, all development-only (table below) |
| `npm audit --omit=dev --json` | 0 | 2 s | **0** advisories |

Caveat (measured, process): the first fuzz file was written into
`fabricui/` at about 20:35–20:37 while `make ci` was in its lint stage and
moved out again before `make race`. The Go test stage of `make ci` ran
before it (its duration matches a run without the fuzz seeds), lint
reported 0 issues; `make race` and `make cover-check` ran on the clean
tree and repeat the same tests green.

## npm audit

213 dependencies (34 prod, 180 dev, 45 optional).

| Package | Severity | Advisory | Path | In the bundle | Fix |
| --- | --- | --- | --- | --- | --- |
| `picomatch` 4.0.7 | high | GHSA-fc5q-39xc-vjpf, ReDoS through single-branch repeated extglobs (`<=4.0.7`) | transitive (tinyglobby, vite, vitest) | no (`--omit=dev`: 0) | none published: 4.0.7 is the newest version; npm's suggestion is a downgrade of vite/vitest |
| `tinyglobby` | high | via picomatch | transitive | no | as above |
| `vite` | high | via picomatch, tinyglobby | direct dev | no | as above |
| `vitest` | high | via picomatch, tinyglobby | direct dev | no | as above |

Decision 11: development-only, so a note (UR-P0-sup-01), not a finding.

## Error-path coverage (from `coverage.out` of `make cover-check`)

Block-level reading of the merged profile (`-coverpkg=./...`, a block
counts as reached when any test binary reached it). A `case a, b:` is one
block, so which disjunct a test took is not visible.

**Fully reached:** `writeProblem`, `newProblem`, `writeProblemBody`,
`writeTermsRequired`, `serverError`, `unready`, `writeJSON`, `decodeJSON`,
`badBody` (both 413 and 400) (`fabricui/problem.go:48-140`); every case of
`controllerError` (`fabricui/handlers_controller.go:164-198`); every case
of `writeAPIError` (`fabricui/handlers_node.go:465-478`), `registryError`
(`fabricui/registry.go:231-240`), `providerError`
(`fabricui/handlers_exposure.go:148-156`).

**Unreached error branches:**

| file:line | Maps | Client-visible problem | Reading |
| --- | --- | --- | --- |
| `fabricui/handlers_pairing.go:177` | `node.ErrCommissioningWindowAlreadyOpen` / `ErrAdministratorWindowOpen` from the open call | 409 `conflict` "a commissioning window is already open" | same problem as the pre-check at :172 (reached) — not distinct |
| `fabricui/handlers_pairing.go:184` | `node.ErrCommissioningWindowNotConfigured` | 503 `service_unready` "no commissioning credentials configured" | distinct → UR-P0-test-01 |
| `fabricui/registry.go:214` | registry nil on rename/delete (`registryTarget`) | 503 `service_unready` "Node registry not available" | distinct for PUT `/nodes/{node}/label` and DELETE `/nodes/{node}` → UR-P0-test-01 |
| `fabricui/registry.go:247` | `decodeJSON` refusal on POST `/nodes`, PUT `/nodes/{node}/label` | 400 / 413 | `badBody` itself is reached elsewhere; per route unreached → UR-P0-test-01 |
| `fabricui/operations.go:209` | `decodeJSON` refusal on PUT `/operation/log/level` | 400 / 413 | as above → UR-P0-test-01 |
| `fabricui/operations.go:219` | `Operations.SetLogLevel` error | 500 `internal` "Setting the log level failed" | generic 500 — not distinct |
| `fabricui/handlers_node.go:335` | `fabrics()` error during factory reset | 500 `internal` | generic 500 |
| `fabricui/handlers_node.go:503` | surface report encoding | 500 `internal` | generic 500, practically unreachable |
| `fabricui/handlers_exposure.go:97` | `ExposureProvider.Schema` error | 500 `internal` | generic 500 |
| `fabricui/handlers_controller.go:638, 778, 921, 1045, 1262, 1419, 1475, 1526` | `req.Body == nil` in SetNodeAutoSubscribe, InvokeNodeCommand, SetNodeACL, SetNodeBindings, CommissionNode, SetCredential, SetFabricLabel, ConsentUpdate | 400 `bad_request` "the body … is required" | **dead code** (measured): the generated strict handler always sets `request.Body = &body` after `runtime.BindJSONBody` and answers a missing/invalid body itself through `RequestErrorHandlerFunc` → `badBody` (`internal/api/api.gen.go:5835-5840` for CommissionNode, same shape for the others) → UR-P0-err-01 |

**Uncovered statements per production file** (`fabricui`, `fabricui/gofabric`):

| Uncovered / total | File |
| --- | --- |
| 125 / 391 | `fabricui/gofabric/node.go` |
| 122 / 256 | `fabricui/gofabric/gofabric.go` |
| 46 / 857 | `fabricui/handlers_controller.go` |
| 40 / 453 | `fabricui/memory_controller.go` |
| 16 / 33 | `fabricui/gofabric/updates.go` |
| 16 / 38 | `fabricui/nodeswitch.go` |
| 9 / 112 | `fabricui/gofabric/hooks.go` |
| 9 / 165 | `fabricui/gofabric/topology.go` |
| 9 / 145 | `fabricui/operations.go` |
| 8 / 105 | `fabricui/handlers_pairing.go` |
| 6 / 221 | `fabricui/handlers_node.go` |
| 6 / 135 | `fabricui/node_detail.go` |
| 6 / 48 | `fabricui/ports.go` |
| 6 / 29 | `fabricui/stream.go` |
| 5 / 121 | `fabricui/registry.go` |
| 4 / 89 | `fabricui/handlers_exposure.go` |
| 3 / 33 | `fabricui/spa.go` |
| 1 / 53 | `fabricui/events.go` |
| 1 / 28 | `fabricui/gofabric/diagnostics.go` |
| 1 / 84 | `fabricui/handlers_diagnostics.go` |
| 1 / 92 | `fabricui/handlers_topology.go` |
| 1 / 18 | `fabricui/options.go` |

Every other production file of the two packages: 0 uncovered.

## Fuzzing (decision 9)

Native Go fuzz targets, `-fuzztime=5m` each, one at a time, default
workers (5), `CGO_ENABLED=0`. Written as temporary test files
(`fabricui/fuzz_p0_test.go`, `fabricui/gofabric/fuzz_p0_test.go`), **removed
after the run** — no target hit, so nothing is left in the tree (the
corpora stayed in the Go build cache only). The first target's wall time
(539 s) includes the coverage-instrumented build. Seeds: the request
bodies and paths of all 142 fixtures under `contract/fixtures/` where the
target takes a body or a path, plus hand-written edge cases.

| Target | Under test (file:line) | Oracle | Execs | New interesting | Result |
| --- | --- | --- | --- | --- | --- |
| `FuzzP0ParseHex` | `parseHex` / `parseHex32` / `parseNodeID`, `fabricui/handlers_controller.go:208-234` | 64-bit round trip through `hex16`; 32-bit bound; errors are 400 `apiProblem` | 983 313 | 149 | pass |
| `FuzzP0QueryWindows` | `queryInt`, `eventWindow`, `pathEndpoint` (`fabricui/handlers_node.go:423-461`), `logWindow` (`fabricui/operations.go:245`), `endpointOf` (`handlers_controller.go:236`) | limits within 1..max; every error a 400 `bad_request` `apiProblem` | 877 216 | 141 | pass |
| `FuzzP0Strings` | `nodeLabel` (`fabricui/registry.go:96`), `checkNodeKey` (`fabricui/handler.go:140`), `parseLevel`/`levelName` (`fabricui/operations.go:100-118`) | accepted label trimmed, non-empty, ≤ 64 runes, no control character; key matches the contract pattern; level round trip | 864 736 | 175 | pass |
| `FuzzP0DecodeBodies` | `decodeJSON` + `badBody` (`fabricui/problem.go:125-140`) into the window, update-list, log-level and label bodies | a refusal is 400/413 `application/problem+json` agreeing with status and `X-Problem-Code` | 816 522 | 213 | pass |
| `FuzzP0ControllerBodies` | every controller write route a fixture names (strict handler → `CommissionNode`, `SetCredential` with `dataset_hex`, ACL, bindings, attribute write, invoke, window, …; `fabricui/handlers_controller.go`), on `MemoryController`, with the token | no 500/5xx outside 501–503; non-2xx is problem+json; no `code`, `passcode`, `passphrase` or `dataset_hex` of the body in the log | 387 311 | 230 | pass |
| `FuzzP0Mux` | the whole mux as `fabricui.Handler` builds it (`newTestHandler` + `spaHandler`, `fabricui/options.go:324-332`): fake node + second node, registry double, memory provider, `MemoryOperations`, `MemoryController` behind a topology-refresh spy; token mode and `Options.Admin` mode, with and without rights; fuzzed method, request target, one header, body | no panic; no 500 or >503; every non-2xx/3xx under `/api/v1/` is problem+json agreeing with status and code header; without rights no admin route (route tables + the controller's `x-admin` map) answers 2xx and no `refresh=true` topology reaches the controller | 320 196 | 199 | pass |
| `FuzzP0ThreadDatasetName` | `ThreadDatasetName`, `fabricui/gofabric/gofabric.go:88` | an error wraps `ErrInvalidCredential`; success names a network | 539 124 | 96 | pass |
| `FuzzP0DiscoveredAndCode` | `discoveredOf` / `cutPlus` (`fabricui/gofabric/gofabric.go:425-468`, the mDNS `VP` TXT value) and `bleFor` after `controller.ParsePairingCode` (`gofabric.go:268-292`) | vendor/product agree with `ParseTxtDecimal`, no product without `+`; BLE used only with a central, a network, not network-only, and a BLE-capable QR or a manual code | 696 455 | 185 | pass |

No crash, no assertion failure, no hit. 3xx answers (the mux's
path-cleaning redirects) were not held to the problem+json rule.

## Findings

No row held back (no security vulnerability found in this phase's
mechanical evidence). Measured = seen in a guard, the profile or the code;
inferred = reasoned from it.

| ID | Severity | file:line | Claim | Measured / inferred | Proposed fix |
| --- | --- | --- | --- | --- | --- |
| UR-P0-sup-02 | low | `Makefile:76-77` (`licenses`) | `go-licenses check` passes a dependency whose licence it cannot classify: `modernc.org/mathutil` v1.7.1 (via go-fabric `store` → `modernc.org/sqlite`) logs `Failed to find license` and the guard exits 0, so a future dependency with an unknown or missing licence would pass the same way | **Measured:** the `E` line and exit 0 in this run; the `LICENSE` file is BSD-3-Clause text (acceptable). **Inferred:** `--disallowed_types=forbidden,restricted,reciprocal` does not include `unknown` | add `unknown` to `--disallowed_types` and list `modernc.org/mathutil` as a reviewed exception (`--ignore modernc.org/mathutil` with a comment naming BSD-3-Clause), so an unclassified licence fails the guard |
| UR-P0-sup-01 | note | `ui/package-lock.json` (`picomatch` 4.0.7) | `npm audit`: 4 high — GHSA-fc5q-39xc-vjpf (picomatch ReDoS, `<=4.0.7`) through tinyglobby, vite, vitest; none in the bundle (`--omit=dev`: 0) | **Measured:** both audit runs; `npm view picomatch versions` ends at 4.0.7, so no fixed version is published. **Inferred:** reachable only through build/test glob patterns from the repository itself | none now (decision 11: dev-only = note); take the patched picomatch through Dependabot once published (an `overrides` entry if vite/vitest lag); the workflow job is phase 3's decision |
| UR-P0-err-01 | note | `fabricui/handlers_controller.go:638, 778, 921, 1045, 1262, 1419, 1475, 1526` | the eight `req.Body == nil` → 400 branches are unreachable: the generated strict handler sets `request.Body = &body` unconditionally after `runtime.BindJSONBody` and answers a missing or invalid body itself through `RequestErrorHandlerFunc` → `badBody` | **Measured:** the blocks are unreached in `coverage.out`; all 13 strict bodies in `internal/api/api.gen.go` assign `&body` (e.g. :5835-5840) | drop the branches, or keep them with a comment naming the generated guarantee; no client-visible change |
| UR-P0-test-01 | note | `fabricui/handlers_pairing.go:184`; `fabricui/registry.go:214`, `:247`; `fabricui/operations.go:209` | four error branches that map a distinct client-visible problem have no test: 503 "no commissioning credentials configured" (`ErrCommissioningWindowNotConfigured`); 503 "Node registry not available" on PUT `/nodes/{node}/label` and DELETE `/nodes/{node}`; 400/413 on a malformed body for the registry routes and PUT `/operation/log/level` | **Measured:** count 0 in `coverage.out` | one table test per branch (fake node returning the error; no registry; a malformed and an oversized body), each asserting status, `code` and problem+json |

Severity count: high 0, medium 0, low 1, note 3. Guards: all green
(`npm audit` exit 1 = the dev-only advisories above). Fuzz: no hit.
