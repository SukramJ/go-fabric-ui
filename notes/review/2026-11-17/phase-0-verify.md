# Phase 0 — verification

Independent verifier (Fable), baseline `25f6e7b4`, branch `review/p0`.
Scope (README "Method per phase" step 3, decision 17): every high and
medium of phase 0, and a sample of the lows — two of the threat model's
five, the evidence unit's one. Each cited line was re-read with the aim of
refuting the row. "Measured" = shown by a temporary Go test against the
handler (`newTestHandler` / `Handler`, `httptest`), or by a run of the one
tool named; "inferred" = read at the line, or browser and proxy behaviour.
The temporary test file is deleted; no production code or existing test was
touched; nothing committed.

Two rows are held back (decisions 5 and 14); their detail, and one new
observation that belongs with them, are in `held/phase-0-verify-held.md`
(git-excluded).

## Verdicts

| ID | Reviewer | Verdict | Final | Deciding `file:line` | M / I | Contract | Reason |
| --- | --- | --- | --- | --- | --- | --- | --- |
| UR-P0-auth-20 | high | confirmed — held | high | held | held | y | held |
| UR-P0-web-21 | medium | confirmed — held | medium | held | held | y | held |
| UR-P0-doc-22 | medium | confirmed | medium | `docs/contract.md:370-374`; `docs/embedding.md:356-359`; `fabricui/handlers_node.go:138`; `fabricui/handlers_controller.go:319-320` | M (token served to a request with no credential: 200 with `csrf_token`); docs read | n | Neither doc, nor the `Options.Admin` comment (`options.go:218-225`), says that reaching the listener is being admin; the only sentence on the token's purpose (`contract.md:372-374`) reads as a protection. A doc finding by decision 7, medium because it is the one property an embedder has to know to place the listener. The fix text must be written together with the fix of the held high (it must not repeat `contract.md:372-374` unqualified) |
| UR-P0-web-23 | medium | confirmed | medium | `fabricui/spa.go:37`, `:70-71` | M (absence: the index answers `Cache-Control`, `Content-Type`, `X-Content-Type-Options` only, also to `Sec-Fetch-Dest: iframe`, `Sec-Fetch-Site: cross-site`); the click-through I | n for the API; host-visible | Not covered by decision 7: the framing page acts through the operator's browser. `frame-ancestors 'self'` is sound for the Ingress (same-origin parent, `docs/embedding.md:349-350`), but a host or user that frames the standalone page from another origin (a dashboard's web-page card) stops working — so the default and the option belong in the wizard although no contract field moves. Set it on the 503 not-built page too (`spa.go:72-76` shares the path, so it does) |
| UR-P0-auth-26 | medium | confirmed | medium | `fabricui/handlers_controller.go:51-74` (the four absent), `:721-722`, `:875`, `:1184`; `contract/openapi.yaml` (no `x-admin` on them) | M (host mode, `Admin` false: `GET …/acl`, `…/bindings`, `…/fabrics` 200; `POST …/attributes` with `fabric_filtered:false` 200 and both fabrics' descriptors; `GET /controller/topology?refresh=true` 403) | y | The handler's own rule at `:130-134` (talking to nodes needs admin) is not applied to these four. In token mode everyone is admin, so this is a host-mode finding; it is real on a pinned host: go-fabric-server's default `--web` is read-only for every peer (`srv:internal/daemon/web.go:106`). Gating is sound; "serve from the last read" needs a cache that X1 has to show exists |
| UR-P0-leak-27 | medium | confirmed; **fix amended** | medium | `fabricui/handler.go:354-355`; `fabricui/operations.go:44-53`, `:227-241`; `ui/src/lib/api/stream.ts:70`, `:114`, `:138` | M (host mode, `Admin` false: `GET /operation/log` 200); what a host's ring holds I | y if gated | Claim holds, and go-fabric-server wires `Operations` in its read-only default (`srv:internal/daemon/web.go:37`, `:106`). The proposed `x-admin` on the log **stream** is not sound as written: the page opens streams with `EventSource`, which cannot send `X-Fabric-UI-Token`, so in token mode the page's own log stream would be refused. Gating the stream needs a fetch-based reader in the client (or the stream stays open and only `GET /operation/log` is gated, which gates nothing). To the wizard with that cost stated |
| UR-P0-ctr-28 | low | confirmed | low | runtime `v1.8.0` `bindjson.go:36-45`; `fabricui/handlers_node.go:326`, `:377`; `fabricui/problem.go:128` | M (`PUT /controller/fabric/label` with `{"label":"x","zzz":1} trailing` → 204; `POST /node/pairing/window` with an unknown field → 400) | y if made strict | 13 generated bindings counted in `internal/api/api.gen.go`; the gate at `handlers_controller.go:137-139` bounds size only, so the plan's "Bodies" bullet is wrong as the row says. Of the two options, the strict decode is the one that also serves the held medium; a docs-only line does not |
| UR-P0-ui-30 | low | confirmed | low | `ui/src/shell/route.ts:37`; called at `ui/src/shell/App.svelte:71-72` | I (read; `decodeURIComponent("%")` throws by specification; not run) | n | `%` is not in `NODE_PAGES`, so the segment reaches the unguarded call; `:71` is the component's initial state, so the throw is at mount, `:72` on every later hash change. Fix sound. Regression test (vitest on `routeOf`) decides by decision 17 |
| UR-P0-sup-02 | low | confirmed | low | `Makefile:77` | M (`go-licenses` v1.6.0 `check` with `unknown` added to `--disallowed_types`: `Unknown license type found for library modernc.org/mathutil`, exit 1; without it exit 0 per the evidence run) | n | The fix does what it says; it needs the `--ignore modernc.org/mathutil` in the same change (and the same line in `.github/workflows/ci.yml:81-83` if that job does not call `make licenses`), or the guard goes red |

Severity changes: none. Struck: none. One fix amended (27), one fix with a
host-visible default to decide (23).

## Decisions 7 and 8 against the severities

- Decision 7 covers finding 22 (doc, not auth) and nothing else in this
  set: 23 is a page of another origin acting through the operator's
  browser, which the assumption does not cover; the two held rows are judged
  in the held file.
- Decision 8: 26 and 27 are host-mode disclosures to a non-admin; both are
  reachable on a pinned host's default configuration (read-only `--web`),
  so medium stands until the owner decides the class.

## Contract changes (decision 13 → wizard, phase 3)

20 (held), 21 (held), 26, 27 if gated, 28 if made strict. 23 changes no
contract field but what a framing host sees; decide its default in the
same wizard. 22, 30 and sup-02 change no contract.

## False-positive rate

Sample of lows: 3 re-read (ctr-28, ui-30, sup-02), 0 struck — **0 of 3**.
All rows verified: 9, 0 struck, 0 severity changes, 1 fix amended.
The sample is small: 0 of 3 bounds the rate loosely, not at zero.

## Missed by the reports

- One observation next to the held high — held, see the held file.
- For finding 27's fix: the `EventSource` limit above (no report names it).
  The same limit applies to any later proposal to gate one of the four
  streams.
- Nothing else; none was hunted for.
