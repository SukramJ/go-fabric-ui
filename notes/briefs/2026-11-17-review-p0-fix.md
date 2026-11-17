# Brief: review phase 0, fix wave

Branch `review/p0` (worktree `../go-fabric-ui-review`), baseline `25f6e7b4`.
Sources: `notes/review/2026-11-17/threat-model.md` (finding rows at the
end), `phase-0-evidence.md` ("Findings"), `phase-0-verify.md`,
`decisions.md` rows 13, 18–21. One implementing agent (Opus), Go and web.
Everything that changes the contract waits for phase 3 (decisions 13, 18–20).

| ID | What | Done when |
| --- | --- | --- |
| UR-P0-web-23 | `Content-Security-Policy: frame-ancestors 'self' <extra>` on the index; new `Options.FrameAncestors []string` (each a CSP host-source or scheme-source, validated in `New`/`Mount`, error otherwise) | red test first; `docs/embedding.md` names the field and the Ingress case |
| UR-P0-web-24 | A CSP on the index: `default-src 'self'`, the inline prepaint script by its hash (derived from the embedded `index.html` at start or guarded by a test that fails when the script changes — never a hand-copied hash that can drift), `object-src 'none'`, `base-uri 'none'`, `form-action 'self'`, the rest as narrow as the page allows (camera scan, inline styles, `data:`/`blob:` images); `X-Content-Type-Options: nosniff` and `Referrer-Policy: no-referrer` on every response of the handler, API and streams included | `make e2e-web` green and no CSP violation in Chromium's console on any page (add an assertion to the smoke if it fits the pattern) |
| UR-P0-leak-25 | `Cache-Control: no-store` on every JSON and problem answer (`writeJSON`, `writeProblemBody`, the hand-set sites, the generated routes through the controller gate) | test over every route of the mux |
| UR-P0-web-33 | An event whose type is not `[a-z_]+` is dropped (and counted or logged without its payload) before the SSE writer | red test with a line break in the type |
| UR-P0-api-32 | `Mount` registers on a scratch mux first, so a duplicate pattern leaves the host's mux untouched; or, if that cannot work with `http.ServeMux`, the error says the mux is unusable and the doc comment says so | test |
| UR-P0-ui-30 | `route.ts:37`: a malformed `%` sequence falls back to the default route instead of throwing | vitest, red first; `make web` in the same commit |
| UR-P0-sup-02 | `make licenses` fails on a dependency whose licence go-licenses cannot classify, with an explicit, commented allow-list for the ones checked by hand (`modernc.org/mathutil`: BSD-3-Clause, say where read) | the target fails without the list entry |
| UR-P0-test-01 | Tests for the four untested error branches (`handlers_pairing.go:184`, `registry.go:214`, `:247`, `operations.go:209`) | each asserts the problem's status and code |
| UR-P0-err-01 | The eight unreachable `req.Body == nil` branches of `handlers_controller.go`: reach them with a test if a request can make them true; if the generated code makes them unreachable, leave them and say why in the report | report says which |

Not in this wave: UR-P0-auth-20, -web-21, -doc-22 (decision 18), -auth-26
(19), -leak-27 and -leak-31 (20), -ctr-28 and -bound-29 (contract: phase 3's
wizard), -sup-01 (no fixed `picomatch` release; Dependabot picks it up).

Changelog: entries under `## [Unreleased]`. No release (decision 15).
Report: `notes/review/2026-11-17/phase-0-fix.md` in the fixed shape.
