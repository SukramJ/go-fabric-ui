# Phase 3 carry-over — the held rows

| row | ID | fold-ins |
| --- | --- | --- |
| phase 0 high | UR-P0-auth-20 | verifier observation (rebinding outside token mode), UR-P2-auth-92 |
| phase 0 medium | UR-P0-web-21 | UR-P2-web-93, UR-P2-web-94 |
| phase 1a medium | UR-P1-adpt-60 | — (adpt-65 same function) |

## Fix shape decided 2026-11-18 (decisions 37, 38)

- **37 (auth-20 + observation + auth-92):** a `Host` check in every mode on
  every route, the bundle included, before routing and before the token is
  issued: admit loopback names, IP literals and the names of a new
  `Options` field (e.g. `AllowedHosts`); anything else 403. A host under a
  DNS name or behind a proxy that rewrites `Host` lists it; changelog and
  `docs/embedding.md` (the proxy case, verifier remark 3). No opt-out.
- **38 (web-21 + web-93 + web-94):** every admin route requires a custom
  request header in both modes (token mode: `X-Fabric-UI-Token`; host mode:
  a fixed header, e.g. `X-Fabric-UI`) — covers forms, `no-cors`, `<img>`
  GETs, old browsers; 415 on a body whose `Content-Type` is not
  `application/json`; 403 when `Origin` is present and not an allowed name
  (row 20's list, not `Host` alone — proxy case, web-94). `Sec-Fetch-Site`
  is not relied on. The docs paragraph on what `Admin` must guarantee is
  part of the fix. Red tests: `held/phase-2-X4.md` "Held red tests".
- **adpt-60 (decision 25):** as `held/phase-1a-verify.md`: default the
  attestation policy to reject in `CommissionDiscovered` before `po.Edit`;
  red test `TestEndToEnd_CommissionDiscoveredRejectsFailedAttestationByDefault`.
