# Phase 3, unit C — docs and supply chain (reviewer report)

Read at `f4b8573` (review worktree; nothing modified but this file).
Brief `notes/briefs/2026-11-18-review-p3.md`, IDs `UR-P3-<theme>-60…89`.
Carried rows (UR-P0-doc-22, UR-P2-doc-62 and the rest of
`read-review-p3-carryover.md`) and decided shapes (decisions 33–44) are not
re-reported. No held candidate in this unit.

Evidence runs (read-only, review worktree): `go list -deps` over
`./fabricui/...` for linux, windows, darwin, freebsd; `npm pack --dry-run`;
`npm audit` and `npm audit --omit=dev` (2026-11-18); a Python comparison of
`docs/contract.md`'s route tables against `contract/openapi.yaml` (method,
statuses, `x-admin`); `gh api repos/SukramJ/go-fabric-ui/branches/main/protection`
and `…/rulesets` (GET only). No scratch worktree was needed: every finding
is a docs, workflow or configuration finding and cites both sides' lines.

## Findings

### UR-P3-sup-60 — low — `contract/previous/` is not the last release's copy

- **Where:** `contract/previous/openapi.yaml:4`; `docs/contract.md:454`, `:464-465`; `README.md:84-85`; `.github/workflows/release-on-tag.yml` (no step)
- **Claim:** the contract guard compares against contract 0.8.1, not the last release (v0.9.2, contract 0.9.0), so a removal of anything 0.9.0 added passes `make contract-check`. The docs say the opposite. **M.**
- **Docs side:** `docs/contract.md:454` "`contract/previous/openapi.yaml` is the copy of the last release, and the contract guard … fails on a removal against it"; `:464-465` "The release commit that tags a version copies the contract to `contract/previous/`."; `README.md:84-85` "a guard fails on any removal against the last release's copy".
- **Code side:** `contract/previous/openapi.yaml:4` `version: 0.8.1`, while `contract/openapi.yaml:4` is `version: 0.9.0` and `git tag` lists `v0.9.0`, `v0.9.1` and `v0.9.2`. `diff contract/openapi.yaml contract/previous/openapi.yaml` shows that `/controller/diagnostics`, `/controller/mdns`, `/controller/events` and their schemas (0.9.0) are missing from `previous`. The last commit that touched `contract/previous` was `e5cfa1a` (U11, which itself moved it to 0.8.1). `release-on-tag.yml` checks the version and the bundle but not `previous`.
- **Proposed fix:** phase 3's contract bump first copies the 0.9.0 contract (`git show v0.9.2:contract/openapi.yaml`) to `contract/previous/`, so the guard covers the 0.9.0 additions. Then add a standing guard: `make version-check` (run by the CI job `web` and by the release job) fails unless `contract/previous/openapi.yaml` is byte-equal to `git show $(git describe --tags --abbrev=0):contract/openapi.yaml` on a release commit, or to the newest tag's copy on any other commit. Then reword `docs/contract.md:464-465` to say what that guard enforces.
- **Contract:** no (the guard's baseline file only; no served route or field changes).

### UR-P3-sup-61 — low — the release job gives every step the npm publish identity and a write token

- **Where:** `.github/workflows/release-on-tag.yml:19-21`, `:37`, `:44`, `:57-58`, `:59-60`, `:75-76`
- **Claim:** `permissions` is set for the whole workflow: `contents: write` plus `id-token: write`. So every step of the single job can mint the OIDC token that npm's Trusted Publisher accepts, and the checkout leaves the write token in `.git/config`. Those steps include `npm ci`, `make dist-check` (Vite and its plugins, all development dependencies that Dependabot auto-merges on minor and patch), an unpinned `npm install -g npm@11` and actions pinned by mutable tag. Code that runs in any of them can publish `@sukramj/fabric-ui` with valid provenance, or push to the repository. **I** (the precondition is a compromised dependency, tag or registry release; nothing like that was observed).
- **Workflow side:** `:19-21` `permissions:` / `contents: write` / `id-token: write # npm trusted publishing and provenance`; `:37` `- uses: actions/checkout@v7` (no `persist-credentials: false`); `:44` `actions/setup-go@v7.0.0`; `:57-58` `name: npm ci` / `run: make web-install`; `:59-60` `run: make dist-check`; `:75-76` `run: npm install -g npm@11`. The header comment (`:8-11`) states that the publish "carries no token: npmjs trusts this workflow through GitHub's OIDC token".
- **Proposed fix:** split the workflow into three jobs:
  1. `verify` with `contents: read`: checkout with `persist-credentials: false`, version-check, `npm ci --ignore-scripts`, dist-check, and the release notes uploaded as an artifact.
  2. `release` (`needs: verify`) with `contents: write` only, which creates the GitHub release.
  3. `publish` (`needs: verify`) with `id-token: write` and `contents: read`, which runs `npm ci --ignore-scripts` and `npm publish --provenance`, with npm pinned to an exact version (`npm@11.x.y`).

  Pin every action by SHA (UR-P3-sup-62). Optionally, a GitHub environment `npm` with the owner as required reviewer guards the publish job.
- **Contract:** no.

### UR-P3-sup-62 — low — actions pinned by mutable tag, against the stated SHA policy

- **Where:** `.github/workflows/ci.yml:45-46`, `:70`, `:73`, `:105-106`, `:119-120`, `:133-134`, `:145`, `:166-167`, `:193-194`; `nightly.yml:32-33`; `release-on-tag.yml:37`, `:44`; `.github/dependabot.yml:24-25`
- **Claim:** `actions/checkout` and `actions/setup-go` are referenced by tag in ci, nightly and release. Five references use the floating major `@v7`: `ci.yml:133`, `:145`, `:166`, `:193` and `release-on-tag.yml:37`. Meanwhile `setup-node`, `cache`, `codeql-action`, `action-gh-release` and `fetch-metadata` are pinned by SHA. Dependabot's configuration states that every action is pinned by SHA. **M.**
- **Docs/config side:** `dependabot.yml:24-25` "# GitHub Actions (pinned by commit SHA; Dependabot bumps the SHA and keeps the `# vX` version comment in sync)."
- **Workflow side:** `ci.yml:133` `- uses: actions/checkout@v7`; `ci.yml:134` `- uses: actions/setup-go@v7.0.0`; `release-on-tag.yml:37` `- uses: actions/checkout@v7`. By contrast, `codeql.yml:32` is `actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1`. `dependabot-auto-merge.yml:17` pins `dependabot/fetch-metadata@25dd0e34…` without the `# vX` comment that the policy names.
- **Proposed fix:** pin every `uses:` by a 40-hex SHA with a `# vX.Y.Z` comment; `codeql.yml:32-33` already has the right SHAs for checkout v7.0.1 and setup-go v7.0.0. Add a standing guard, a `lint`-job step (or a Go test under `internal/arch`) that fails on any `uses:` line not matching `@[0-9a-f]{40} # v`.
- **Contract:** no.

### UR-P3-sup-63 — note — a `v*` tag on any commit publishes, untested

- **Where:** `.github/workflows/release-on-tag.yml:12-16`, `:52-60`; `gh api …/rulesets`
- **Claim:** the release job runs `version-check` and `dist-check` only, without the test, lint, security or contract jobs. It also does not check that the tagged commit is on `main`, or that the commit's required checks passed. No ruleset protects `v*` tags (`gh api repos/SukramJ/go-fabric-ui/rulesets` → `[]`), so a tag pushed on an unmerged branch commit releases the Go module and publishes the npm package. **M** for the workflow and the empty ruleset list; **I** for impact (only the owner holds write access today).
- **Workflow side:** `:12-16` `on: push: tags: - 'v*'`; the only gates are `:52-56` (version) and `:59-60` (`make dist-check`).
- **Proposed fix:** in `verify` (UR-P3-sup-61), fail unless `git merge-base --is-ancestor "$GITHUB_SHA" origin/main` holds and `gh api repos/{repo}/commits/$GITHUB_SHA/check-runs` shows each of the ten required contexts with conclusion `success`. Also add a tag ruleset that restricts creation of `v*` to the owner.
- **Contract:** no.

### UR-P3-sup-64 — note — six required checks accept a status from any source

- **Where:** branch protection of `main` (`gh api repos/SukramJ/go-fabric-ui/branches/main/protection`, read 2026-11-18)
- **Claim:** of the ten required contexts, `build`, `lint`, `test (ubuntu-latest)`, `test (windows-latest)`, `codeql (go)` and `security (vuln + licenses + secrets)` have `"app_id": null`, so any integration or token with `statuses: write` can satisfy them with a commit status. Only `coverage`, `web`, `contract` and `e2e-web (…)` are bound to GitHub Actions (`app_id` 15368). `enforce_admins` is false and no review is required. The ten contexts match the ten jobs of `ci.yml` and `codeql.yml` exactly. **M** for the settings; **I** for impact (no workflow here grants `statuses: write`).
- **Proposed fix:** bind all ten checks to app 15368 (GitHub Actions). That is an owner action in the settings; no file changes.
- **Contract:** no.

### UR-P3-sup-65 — low — Dependabot auto-merges a go-fabric minor (0.x) bump

- **Where:** `.github/dependabot.yml:21-22`; `.github/workflows/dependabot-auto-merge.yml:19`
- **Claim:** go-fabric is `v0.22.2`, so Dependabot classes 0.22 → 0.23 as `version-update:semver-minor`. That bump falls into the `go-minor-patch` group and is auto-merged on green CI, because only `semver-major` is excluded. Under semver 0.x a minor is breaking. This repository treats each go-fabric version as a wave: CLAUDE.md "State" ties each release to a go-fabric version, and the docs cite go-fabric "at v0.22.2" and other tags. An auto-merge moves the pin under every such citation without the Fable read or a changelog line. **I** (no go-fabric bump has been auto-merged yet; `git log --author=dependabot` shows only npm, Actions and kin-openapi bumps).
- **Config side:** `dependabot.yml:21-22` `go-minor-patch:` / `update-types: ["minor", "patch"]`; `dependabot-auto-merge.yml:19` `if: steps.meta.outputs.update-type != 'version-update:semver-major'`.
- **Proposed fix:** exclude `github.com/SukramJ/go-fabric` from auto-merge with `&& !contains(steps.meta.outputs.dependency-names, 'github.com/SukramJ/go-fabric')`, or give it a Dependabot group of its own that the auto-merge skips. Auto-merge then covers only patch bumps of 0.x modules.
- **Contract:** no.

### UR-P3-sup-66 — note — CodeQL scans the Go half only

- **Where:** `.github/workflows/codeql.yml:38-41`; `ui/src/lib/pages/node/Pair.svelte:174`, `ui/src/lib/pages/controller/node/Window.svelte:102`
- **Claim:** `languages: go`. The web half (TypeScript and Svelte, two `{@html qrSvg}` sinks, the hash router, the stream parser) has no static analysis besides svelte-check. **M.**
- **Proposed fix:** add `javascript-typescript` with `build-mode: none` as a second matrix entry, and add its context to the required checks.
- **Contract:** no.

### UR-P3-sup-67 — low — Tailwind's generated CSS is in the bundle, not in the notices

- **Where:** `THIRD-PARTY-NOTICES.md:27-31`; `ui/vite.config.ts:27-29`; `fabricui/dist/assets/index-B-JMWn6y.css`
- **Claim:** the committed bundle's stylesheet contains Tailwind CSS 4's generated layers, including its preflight (MIT, Tailwind Labs, derived from modern-normalize, MIT). The notices say that build-time tools such as Tailwind are not shipped. The vendor-list plugin that `notices-check` relies on reads only JS chunks' modules, so the guard cannot see a package that reaches the bundle as CSS. **M** for presence; **I** for the licence obligation, which depends on whether the preflight counts as a substantial portion of the software.
- **Docs side:** `THIRD-PARTY-NOTICES.md:30-31` "Build-time tools (Vite, Tailwind CSS, TypeScript, the test stack) are not shipped."
- **Code side:** `ui/src/shell/app.css:7` `@import "tailwindcss" source(none);`. In the bundle CSS: `@layer theme`, `@layer base` and the preflight rule `::file-selector-button{box-sizing:border-box;border:0 solid;margin:0;padding:0}`. `vite.config.ts:27-29`: `for (const chunk of Object.values(bundle)) { if (chunk.type !== "chunk") continue; for (const id of Object.keys(chunk.modules))`, so CSS assets are skipped.
- **Proposed fix:** list `tailwindcss` (MIT) in the bundled table with `licenses/npm/tailwindcss.txt`, and correct the sentence. Make `check-notices.mjs` require a fixed list of CSS-origin packages (today `tailwindcss`) beside the JS vendor list, so the guard covers it. Side observation (**I**): the plugin reads each package's version from the top-level `node_modules/<name>/package.json`, so a nested copy would be listed with the wrong version. No bundled package is nested today (`package-lock.json`).
- **Contract:** no.

### UR-P3-sup-68 — note — the Go table in the notices is stale

- **Where:** `THIRD-PARTY-NOTICES.md:38-40`, `:48`, `:51`
- **Claim:** the table lists `github.com/cenkalti/backoff` and `github.com/grandcat/zeroconf`, which are in neither `go.mod` nor `go.sum` (`grep -c` → 0). `go list -deps ./fabricui/...` for linux, windows, darwin and freebsd links 15 third-party modules, all listed; the two extras appear in none of those lists. The sentence about `make licenses` predates UR-P0-sup-02's fix. **M.**
- **Docs side:** `:38-40` "`make licenses` (go-licenses) fails on a forbidden, restricted or reciprocal licence."
- **Code side:** `Makefile:93` `--disallowed_types=forbidden,restricted,reciprocal,unknown`.
- **Proposed fix:** drop the two rows and add "or unclassified". Optionally add a guard that compares the table with `go list -deps -f '{{.Module.Path}}'` over the four GOOS.
- **Contract:** no.

### UR-P3-sup-69 — note — the generator's module graph is in every consumer's `go.mod` graph

- **Where:** `go.mod:9-40`, `:42`
- **Claim:** `tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen` puts the generator's dependencies into the module's `require` block as `// indirect`: kin-openapi, speakeasy-api/openapi, oasdiff/yaml{,3}, go-yit, yaml-jsonpath, jsonschema, `golang.org/x/tools v0.52.0`, `x/mod`, `x/text`, `x/sync`, about 17 of 31 indirect lines. None of them is linked into `fabricui` (`go list -deps`, above). With module graph pruning, a consumer still loads this `go.mod`'s requirements, so these versions raise the consumer's minimum versions of `x/tools`, `x/text`, yaml and the rest in MVS. **I** (Go's documented pruning rules; not reproduced in a consumer).
- **Proposed fix:** move the generator to its own module (`tools/go.mod` with the `tool` line; `script/gen-api.sh` runs `go tool -modfile=tools/go.mod oapi-codegen`), so the root `go.mod` requires only go-fabric, the runtime and their graph. Or accept it and record that in `THIRD-PARTY-NOTICES.md:62-64`.
- **Contract:** no.

### UR-P3-sup-70 — note — no standing audit of the npm packages (decision 11)

- **Where:** `Makefile:105` (`npm ci --no-audit`); `.github/workflows/ci.yml` (no audit step)
- **Claim:** Go dependencies have a standing `govulncheck` (`ci.yml` job `security`). The npm packages, both the five that land in the bundle and the development stack that builds it in the release job (UR-P3-sup-61), have no audit anywhere in the pipeline. Phase 0 found 4 high advisories, all development-only (`phase-0-evidence.md:31-32`). Today's runs (2026-11-18) report 0 for both `npm audit --omit=dev` and `npm audit`. **M.**
- **Proposed fix (the decision 11 job):** add `npm audit --omit=dev --audit-level=low` to the `web` job, so it fails. That matches decision 11: an advisory on a bundled package is a finding. Add a non-failing `npm audit` of the whole tree to `nightly.yml`, which writes to the job summary; a development-only advisory is a note and must not turn an unrelated pull request red, as the tools' pinning comment in `ci.yml:30-31` argues.
- **Contract:** no.

### UR-P3-sup-71 — note — Makefile and README describe a `make ci` that is not the pipeline; dead `FUZZTIME`

- **Where:** `Makefile:10`, `:135`; `README.md:91`; `ci.yml:9-11`
- **Claim:** `README.md:91` says `make ci` runs "build vet test lint web-check dist-check" (CLAUDE.md repeats this), but `Makefile:135` also runs `generate-check`. The target's help text says "everything the pipeline's main jobs run", yet it leaves out race, cover-check, vuln, licenses, secrets, notices-check, version-check, contract-check, e2e-web and the 386 build. `FUZZTIME ?= 10000x` (`Makefile:10`) is used by no target. The `ci.yml:9-11` header says that there is no fuzz job, which decision 9 is about (for unit B). **M.**
- **Proposed fix:** correct `README.md:91` and the help text ("the main Go and web checks; the pipeline also runs …"). Delete `FUZZTIME`, or use it in the fuzz target that decision 9 asks for.
- **Contract:** no.

### UR-P3-doc-72 — low — `Options.Prefix` with `Handler` gives a page that cannot reach its API

- **Where:** `docs/embedding.md:42`, `:380-382`; `ui/src/shell/App.svelte:62`; `fabricui/options.go:256-257`, `:336-345`; `fabricui/spa.go:212`
- **Claim:** the Options table offers `Prefix` as a free setting. With `Handler`, however, the embedded shell always calls `./api/v1` relative to the page. Under any other prefix every call of the page reaches the bundle's file server and answers 404, so the shell shows its "server did not answer" state. Nothing in the docs or in `Handler` says or enforces that `Handler` needs the default prefix. **M** by reading both sides; not run.
- **Docs side:** `embedding.md:42` "| `Prefix` | The API prefix; default `/api/v1`. |". `:380-382` "the client calls `./api/v1`, so the page works under any prefix without configuration" (the proxy's prefix, not `Options.Prefix`).
- **Code side:** `App.svelte:62` `createClient("./api/v1")`. `options.go:336-345`: `Handler` registers the API under `o.Prefix` and the SPA at `/`. `spa.go:53-58`: a path that is no file of the bundle → `http.NotFound`. `spa.go:212`: the not-built page names `/api/v1/node/`.
- **Proposed fix:** either `Handler` refuses a non-default `Prefix` (an error from `newHandler` when the SPA is served; `Mount` keeps the field), or the docs row says "with `Handler` the shell calls `./api/v1`; set `Prefix` only with `Mount`". Fix the not-built page's text with it.
- **Contract:** no.

### UR-P3-doc-73 — low — README and CONCEPT report the state of v0.4.1 and v0.5.0

- **Where:** `README.md:18-42`; `CONCEPT.md:31-41`
- **Claim:** README's status banner says `v0.4.1`, and calls 0.5.0, U5a and U7 "unreleased". The module and the package are `0.9.2` (`internal/version/version.go`, `ui/package.json:3`, `changelog.md:135`) and tags `v0.5.0` to `v0.9.2` exist. CONCEPT's wave table ends at U5, marks U4a and U4b "`v0.5.0` (unreleased)", and its U5 row is German ("Betrieb: Live-Status, Betriebsseite …"), against CLAUDE.md's "Everything in the repository is English". The hosts read the README first. **M.**
- **Docs side:** `README.md:18` "> **Status: `v0.4.1`**"; `:24` "`0.5.0` (waves U4b and U4a, unreleased)"; `:36-37` "Wave U7 (unreleased, contract 0.7.0)"; `CONCEPT.md:41` "| U5 | Betrieb: …".
- **Proposed fix:** the banner names `v0.9.2` (contract 0.9.0) and keeps one line per wave with its release. Add the CONCEPT rows for U5a to U11 with their versions, in English.
- **Contract:** no.

### UR-P3-doc-74 — note — `Handler`'s godoc still promises the index for any path

- **Where:** `fabricui/options.go:333-335`
- **Claim:** the exported doc says every other path falls back to the index. Decision 36 (UR-P2-web-91) made such a path 404, and `docs/embedding.md:20-24` and `changelog.md:86-90` say so; the godoc, which pkg.go.dev shows to hosts, was missed. **M.**
- **Docs side:** `options.go:333-335` "Handler returns the API routes and the embedded web interface: the API under o.Prefix, the bundle at "/" with a fallback to its index for any other path."
- **Code side:** `spa.go:26-27` "the root as index.html, a file that exists as itself, anything else 404"; `spa.go:48-65`.
- **Proposed fix:** "…the bundle at "/": the root and /index.html answer the index, a file of the bundle answers itself, any other path 404."
- **Contract:** no.

### UR-P3-doc-75 — note — "A 5xx problem carries no detail" is true of 500 only

- **Where:** `docs/contract.md:389-390`
- **Claim:** 503 `service_unready` and 501 `not_implemented` carry a fixed sentence as detail, and 502 `node_error` carries the node's error text. The document itself says so later (`:714-716`, "Unlike the other 5xx problems it carries a detail"). **M.**
- **Docs side:** `:389-390` "A 5xx problem carries no detail — the cause, which can name paths, goes to the log."
- **Code side:** `handler.go:661` `unready(w, r, "Node not available", "the host serves no node (Options.Node is nil)")`; `handlers_node.go:386` 501 with "the host wires no Erase"; `handlers_controller.go:197` `writeProblem(w, r, http.StatusBadGateway, codeNodeError, "The node failed", err.Error())`; only `problem.go:106-124` (500) writes `""`.
- **Proposed fix:** "A 500 problem carries no detail …; a 501 or 503 carries a fixed sentence, a 502 the node's answer."
- **Contract:** no (prose only).

### UR-P3-doc-76 — note — "every POST, PUT and DELETE" is admin; one POST is not

- **Where:** `fabricui/options.go:219-222`; `docs/contract.md:370-372`
- **Claim:** `POST /controller/nodes/{id}/attributes` is a read and not `x-admin`. Conversely, `POST /node/exposure/preview` is admin, though it writes nothing. **M.** If decision 40 (attribute values gated) makes that route `x-admin`, the sentence becomes true for this route.
- **Docs side:** `options.go:219-222` "Admin decides whether a request may write (every POST, PUT and DELETE route) …"; `contract.md:370` "A writing route (POST, PUT, DELETE), `GET /node/pairing` and every controller route marked `x-admin` … run only for an admin".
- **Code side:** `contract/openapi.yaml:1356` `operationId: readNodeAttributes` without `x-admin` (the PUT at `:1380-1381` has it); `docs/contract.md:531` "| `/controller/nodes/{id}/attributes` | POST | – |".
- **Proposed fix:** after the phase 3 gates land, word both places as "every route the contract marks `x-admin`" (one rule), as decision 41's allow-list will.
- **Contract:** no.

### UR-P3-doc-77 — note — `FrameAncestors` missing from the Options table; the one-label wildcard refusal undocumented

- **Where:** `docs/embedding.md:26-47`, `:424-428`; `changelog.md:12-15`
- **Claim:** the Options table lists 21 fields but not `FrameAncestors`, which is described only at `:417-431`. The prose and the changelog list the refused any-parent forms as `*`, `https:` and `https://*`. The code and the godoc also refuse a wildcard over one label (`https://*.com`, `*.local`), so a host reading the docs expects that form to be accepted. **M.**
- **Docs side:** `embedding.md:426-427` "and on an entry that admits any parent (`*`, a scheme-source such as `https:`, `https://*`)"; `changelog.md:14-15` likewise.
- **Code side:** `spa.go:183-195` `anyParent`: "one whose wildcard spans a single label ("*.com", "*.local")"; `options.go:276-277`.
- **Proposed fix:** add the row `| FrameAncestors | … see Response headers and framing |`, and add "a wildcard over one label (`https://*.com`)" to both lists.
- **Contract:** no.

### UR-P3-doc-78 — note — `/node/surface` answers 500, which the docs list and the contract does not (for A)

- **Where:** `docs/contract.md:28`; `contract/openapi.yaml:203-221`; `fabricui/handlers_surface.go:25`
- **Claim:** the docs table says `200, 500, 503`, and the handler can answer 500 (`h.serverError(w, r, "Failed to encode the surface report", err)`). The operation's responses in the contract are 200 and 503 only. This is the one status mismatch between `docs/contract.md`'s tables and the contract; the comparison script found none in method or admin. **M.**
- **Proposed fix:** add `"500"` to `/node/surface` and `/nodes/{node}/surface` in the contract (unit A owns the contract side).
- **Contract:** yes (an added response; minor).

## For unit D (claims about go-fabric, not checked here)

- `docs/embedding.md:251-252`: "go-fabric's `logring` package is that handler once it lands (go-fabric issue #393)". Is it still unlanded at v0.22.2?
- `docs/contract.md:656`: the controller "tracks no session-id occupancy, so `occupancy` is all zero". The adapter passes go-fabric's occupancy through (`gofabric/diagnostics.go:27-28`) and `contract/openapi.yaml` says "capacity 0 means …", so this is a go-fabric claim at v0.21.0/v0.22.2.
- `docs/embedding.md:288-311` (the facade table) and `docs/contract.md:595-625`: the adapter's go-fabric calls.

## Covered without finding

- `docs/contract.md` route tables against `contract/openapi.yaml` (script): every method, admin mark and status set agrees, except UR-P3-doc-78. The rows that combine two methods (`acl`, `bindings`, `credentials`, `consent`/`check`, `enable`/`disable`, `icd/register`/`unregister`) agree when read by hand. `/nodes/{node}/…` is documented as "as `/node/…`".
- `docs/contract.md`: the node stream's event list against every event type the module publishes (`events.go`, handlers); heartbeat 15 s (`handler.go:268`); buffer 64 (`stream.go:20`); window 180..900 s, default 900, and `Retry-After: 30` (`handlers_pairing.go:22-24`, `:179`); the confirmation bodies `remove-all-fabrics` and `erase-node` (`handlers_node.go:321`, `:371`); the event limit of 1000; the log limits 200/2000 (`operations.go:152-153`); `%016X` ids; `maxBodyBytes = 1 << 20`; `X-Problem-Code`; the problem code list (`problem.go:16-25`); the label limit of 64; the key pattern and reserved page names (`handler.go:165`, `:169`); the 404 for `/controller/*` without a controller (`handler.go:466-473`); `capabilities.diagnostics` gating (`handlers_controller.go:1730`, `:1958`); `csrf_token` on the controller status (`handlers_controller.go:324`).
- `docs/embedding.md`: every `Options` field but `FrameAncestors` (UR-P3-doc-77) and the defaults against `options.go:176-296`; `NodeEntry` fields; the root, `/index.html`, files and 404 (`spa.go:46-65`); the API catch-all 404; the CSP directives (`spa.go:129-150`); nosniff, no-referrer and no-store (`handler.go:376-393`, `spa.go:45`, `:104`); detached two-minute runs for erase, registry delete, decommission and the reassembly after apply (`handlers_node.go:613-620`, `registry.go:338`, `handlers_controller.go:562`, `handlers_exposure.go:216`); `NewNodeSwitch(nil)` and `Set(nil)` panic (`nodeswitch.go:61`, `:73`); the four host ports on `GET /node/status` (`handlers_node.go:139-210`); every named export (`GoFabricVersion`, `MemoryOperations`, `NewMemoryController`, `NoDiagnostics`, `NewMemoryCredentials`, `ObserveNode`, `ErrTermsRequired`, `ErrNotImplemented`, `ErrNodeLimit`, `ErrDefaultNode`, `gofabric.NewHooks`, `ControllerOf`, `ThreadDatasetName`) exists.
- `docs/provider.md`: the error mapping (`handlers_exposure.go:149`); empty batch 204; the order `exposure_changed`, then reassembly, then `endpoints_changed` on success only (`handlers_exposure.go:198-225`).
- `ui/README.md`: every page and function it names is exported from `@sukramj/fabric-ui` (`src/lib/index.ts`, `pages/index.ts`, `context.ts`); `subscribeStream`, `subscribeControllerStream` and `subscribeLogStream` signatures; `applySkin` and `startHaBridge` from `./theme`; `createTranslator(locale, extra)`; dev server `:5174` with `FABRIC_UI_API` default `http://localhost:8080` (`vite.config.ts:65-68`). The catalogue prefixes leave out `kit.*` and `select.*`, which is too small to report.
- `changelog.md` `[Unreleased]` against the code (CSP, headers, 404 for unknown paths, detached runs, `LICENSES_CHECKED`, the stream type filter); version 0.9.2 agrees in `internal/version`, `ui/package.json` and the newest changelog section.
- Reports (`docs/reports/*.md`): grep for "fallback", "deep path" and "every deep" found no claim on current behaviour that `embedding.md` or `contract.md` relies on.
- Workflows: `ci.yml`, `nightly.yml` and `codeql.yml` have read-only top-level permissions (`security-events: write` on codeql only). `pull_request_target` appears only in the auto-merge workflow, which checks out nothing, runs no PR code and is limited to `dependabot[bot]`. The only secret is `GITHUB_TOKEN`. Tool installs are version-pinned through the Go checksum database (`ci.yml:30-35`, `Makefile:12-16`). Dependabot has a 7-day cooldown on all three ecosystems, and majors are excluded from auto-merge (grouped Actions majors report the highest update type).
- Branch protection: the ten required contexts are exactly the ten jobs (ci 9, codeql 1), with `strict: true` and force-push and deletion off (see UR-P3-sup-64 for `app_id`).
- The npm package (`npm pack --dry-run`, 89 files): only `src/lib` sources, the two catalogues, README, LICENSE and package.json; no test file, fixture, `.npmrc` or sourcemap. No install script reaches a Linux runner (the only `hasInstallScript` is `fsevents`, darwin-only optional). `npm publish --provenance` without a token, as the header says.
- The committed bundle: no sourcemap, absolute path or `console.log`/`debug`. `dist-check` diffs both directions (`diff -r`) on the pinned Node 22.22.1, in the `web` job and the release. `notices-check`: 7 JS-bundled packages with accepted licences and texts (see UR-P3-sup-67 for the CSS).
- `schema.gen.ts` freshness is guarded by `ui/src/lib/api/schema.test.ts` (`gen-types` needs no CI target); `generate-check` runs in `lint`; `e2e` is a subset of `test`.
- Go licences: every linked module is MIT, BSD-3-Clause or Apache-2.0; `unknown` is disallowed (`Makefile:93`).

## Counts

| Severity | Count | IDs |
| --- | --- | --- |
| high | 0 | — |
| medium | 0 | — |
| low | 7 | sup-60, sup-61, sup-62, sup-65, sup-67, doc-72, doc-73 |
| note | 12 | sup-63, sup-64, sup-66, sup-68, sup-69, sup-70, sup-71, doc-74, doc-75, doc-76, doc-77, doc-78 |

19 findings: 13 measured (sup-60, -62, -66, -68, -70, -71, doc-72 to -78), 3 inferred (sup-61, -65, -69), 3 measured with an inferred impact (sup-63, -64, -67). One is a contract change (doc-78, for A). No held candidate.
