# Hand-over 2026-11-27: #85 close-out (next wave)

Not a brief yet. The next session's Fable agent writes the brief
(`notes/briefs/<date>-<wave>-85-closeout.md`) from this hand-over.

## State

- main at `51bd01c` (#95). Released: `v0.14.2` (go-fabric v0.28.0, the
  other dependencies up to date, #94) and `v0.15.0` (contract 0.15.0:
  `setFabricLabel` declares 409 `conflict`, "Refused by the host", #95).
- go-fabric pin: `thingfabric.dev/go-fabric v0.28.0`. The five identifiers
  deprecated in v0.27.0 (removable in v0.29.0) are not used here.
- Issue #85 is open with four notes; the comment of 2026-11-27 records the
  `setFabricLabel` part as done. Rows and evidence:
  `notes/review/2026-11-17/phase-4r2-register.md:99,104,115,116`.

## Open points

| Note | Question | Where the answer comes from | Lands in |
| --- | --- | --- | --- |
| ui-396 | Does a failing assembly (a rebuild after a selection apply) raise the observer, i.e. send `surface_changed` or `endpoints_changed`, at go-fabric v0.28.0? | Fable fact note, go-fabric `node` at v0.28.0 (the assembly and its observer) | `ui/src/lib/store.svelte.ts:480-486` (clears `exposureNotInEffect` on both events); register: clear on `endpoints_changed` only, once read |
| err-404 | Can `client/discovery` / the controller's `Discover` return a device without an address? Which `controller.ErrConfiguration` cases reach `PairDiscovered`? | Fable fact note, go-fabric `controller`, `client/discovery` at v0.28.0 | `fabricui/gofabric/gofabric.go:444` maps `ErrConfiguration` to `ErrInvalidArgument` (400); the registry path answered 502. Either align here, or a question to go-fabric to split the error |
| api-391, rest | A host's own `ControllerSource` can answer 409, 502 or 404 on operations that do not declare them (13 answers by a temporary guard case). Declare them, or map undeclared kinds to 500 per operation as the onboarding routes do (`opaqueError`, `fabricui/handlers_controller.go:2206-2245`)? | Fable decision proposal; owner decides | contract (then 0.16.0) or the handler bodies; `fabricui/undeclared_status_test.go` |
| test-403 | No test lets go-fabric's real update manager fail one product and reads the route's `errors` and the single `check_finished` | none, implementation (Opus): an image source that refuses the download as a fixture | `fabricui/gofabric/updates.go` tests; the seam (`updates.go:24-30`) stays |

## Next step

1. Fresh session, Opus as coordinator.
2. One Fable agent, bounded: read go-fabric v0.28.0 for ui-396 and err-404
   (fact note with `file:line` to `notes/reads/read-85-closeout.md`), and
   propose the api-391 policy with its contract consequence. Then write the
   brief.
3. Owner decides api-391.
4. Opus implements all four in one PR on a branch; guards as `CLAUDE.md`
   lists; release as `v0.16.0` (contract change) or `v0.15.1` (none).

## Lessons from this session (local guards on macOS)

- `GOARCH=386` alone fails on darwin ("unsupported GOOS/GOARCH pair");
  run `GOOS=linux GOARCH=386 make build`; `make test-386` executes on CI only.
- After a dependabot npm bump, `ui/node_modules` is stale: run
  `make web-install` before `dist-check` / `notices-check`. If `npm ci`
  fails with ENOTEMPTY, `rm -rf ui/node_modules` first.
- `npm install pkg@x` writes `^x`; the repo pins exact versions. Pin with
  `--save-exact` and leave `peerDependencies` (`svelte: ^5.57.0`) alone.
- zsh does not split `$VAR` words: no `set -- $M`.
