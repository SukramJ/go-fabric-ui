# Review phase 3 — carry-over inventory (script, read-only map)

Generated at `f4b8573` by an awk script over the four registers
(`notes/review/2026-11-17/phase-{0,1a,1b,2}-register.md`): every row whose
disposition or "where" names phase 3 (or "phase 1b / 3"). A map, never a
citation. `unit` is the script's guess from the row's words: `fix` = the
phase 3 fix bundle, `wizard` = the row asks for an owner decision,
`api`/`cite` = a phase 3 review unit, `upstream+fix` = waits on a go-fabric
issue. Held rows stay "(held)" (decisions 5, 14).

| phase | ID | sev | unit | disposition | where |
| --- | --- | --- | --- | --- | --- |
| 0 | (held) | high | fix | open → phase 3 | decision 18 |
| 0 | (held) | medium | fix | open → phase 3 | decision 18 |
| 0 | UR-P0-doc-22 | medium | fix | open → phase 3, with the held rows | decision 18: its text depends on their fix |
| 0 | UR-P0-auth-26 | medium | fix | open → phase 3 | decision 19: all four routes gated (contract) |
| 0 | UR-P0-leak-27 | medium | fix | open → phase 3 | decision 20: log and log stream gated, the client's log stream through `fetch` (contract, web) |
| 0 | UR-P0-ctr-28 | low | wizard | open → phase 3's wizard | contract if made strict (decision 13); the plan's "Bodies" bullet corrected |
| 0 | UR-P0-bound-29 | low | wizard | open → phase 3's wizard | a new 503 case (contract) |
| 0 | UR-P0-leak-31 | note | fix | open → phase 3, with UR-P0-leak-27 | decision 20 |
| 1a | (held) | medium | fix | open → phase 3 | decision 25 |
| 1a | UR-P1-err-04 | low | fix | open → phase 1b / 3 | the page reading the window's kind is 1b; a field is a contract change |
| 1a | UR-P1-err-07 | low | fix | log part fixed `12e7c6e`; optional `fabric_count` → phase 3 |  |
| 1a | UR-P1-leak-09 | low | fix | open → phase 3 | with UR-P0-leak-27/-31 (decision 20) |
| 1a | UR-P1-adpt-65 | low | fix | open → phase 3 | needs a presence flag on the port's discovered device; bundled with phase 3's commissioning changes |
| 1a | UR-P1-ctr-32 | low | fix | 502 part fixed `e791329`; enum `unknown` → phase 3 |  |
| 1a | UR-P1-err-33 | low | upstream+fix | open → phase 3 / upstream | node route 404 is a contract change; the controller route needs go-fabric's fabric-not-found error, which the facade does not export (issue with phase 3) |
| 1a | UR-P1-adpt-35 | low | fix | fixed `a92a4e8` (no-contract part) | 65535 answers the unknown endpoint's empty list; 400 or `maximum: 65534` → phase 3 |
| 1a | UR-P1-err-36 | low | fix | log part fixed `12e7c6e`; a 500 → phase 3 |  |
| 1a | UR-P1-bound-37 | low | fix | fixed `b80bd20` (no-contract part) | chunked read, a refused endpoint is 502; a per-endpoint status → phase 3 |
| 1a | UR-P1-web-20 | note | api | open → phase 3 api | from the pre-merge read: `FrameAncestors` admits a wildcard over a public suffix (`https://*.co.uk`); refusing it needs a suffix list |
| 1a | UR-P1-conc-21 | note | api | open → phase 3 api | from the pre-merge read: on the path a deletion overtook, `NodeRegistry.List` is called under the handler's registry lock; the port's doc does not forbid a call back into the handler |
| 1a | UR-P1-ctr-13 | note | fix | open → phase 3 | 405 is a new status |
| 1a | UR-P1-ctr-17 | note | fix | open → phase 3 | with UR-P0-ctr-28 |
| 1a | UR-P1-api-18 | note | fix | open → phase 3 | a correlation field |
| 1a | UR-P1-ui-38 | note | fix | open → phase 3 | the fixtures carry both names; either single name changes a fixture |
| 1a | UR-P1-adpt-40 | note | cite | open → phase 3 `cite` | rune or byte limit of the fabric label: a reference question for Fable |
| 1a | UR-P1-leak-66 | note | fix | open → phase 3 | with UR-P0-leak-27/-31 (decision 20) |
| 1b | UR-P1-ctr-113 | low | fix | open → phase 3 | the event number as a string is a contract change; with UR-P1-ctr-177 |
| 1b | UR-P1-ui-133 | low | fix | fixed `33b4d13`, `1f9609e`; controller `admin` field → phase 3 | with UR-P1-ui-196 |
| 1b | UR-P1-ctr-168 | low | fix | open → phase 3 | needs a contract field for the auto-subscribe setting |
| 1b | UR-P1-api-171 | low | api | open → phase 3 `api` | exporting the controller Diagnostics page; name clash with the type |
| 1b | UR-P1-web-115 | note | fix | fixed `7b57f41` (narrowed) | a hex pattern on the contract's `NodeID` → phase 3 |
| 1b | UR-P1-ctr-177 | note | fix | open → phase 3 | with UR-P1-ctr-113 |
| 1b | UR-P1-ui-178 | note | cite | open → phase 3 `cite` | the invalid passcodes rest on a reading of the specification |
| 1b | UR-P1-ui-179 | note | cite | open → phase 3 `cite` | whether an ACL entry may name `pase` |
| 1b | UR-P1-ui-196 | note | fix | fixed `8829c9a`; `admin` on the controller status → phase 3 | pre-merge read |
| 1b | UR-P1-web-197 | note | api | open → phase 3 `api` | the token header follows a host's `baseUrl`; a docs sentence or an origin check |
| 2 | UR-P2-ui-05 | low | fix | fixed `8b94c01` | the flow takes its own window's kind from the status; the residue (a failed status read and a replacement before the next) → phase 3 with decision 28's field |
| 2 | UR-P2-ui-06 | note | fix | open → phase 3 | the countdown needs a remaining-seconds field (contract) |
| 2 | UR-P2-conc-34 | low | api | docs `56d1682`; an `Options` field → phase 3 `api` | the host's shutdown of open streams |
| 2 | UR-P2-ui-35 | note | fix | open → phase 3 | an exposure apply whose rebuild fails needs a field (contract) |
| 2 | UR-P2-err-37 | note | upstream+fix | upstream go-fabric, then phase 3 | the facade exports neither error that tells a refused write from a node failure |
| 2 | UR-P2-auth-60 | note (verifier; X3 low) | fix | open → phase 3 | decision 33: the cached ACL, bindings, fabrics and NOCs gated |
| 2 | UR-P2-leak-61 | note | fix | open → phase 3 | decision 33: endpoint 0's NOCs and ACL events gated |
| 2 | UR-P2-doc-62 | low | fix | open → phase 3, with UR-P0-doc-22 | what a false `Admin` serves, written once the gates of decisions 19, 20, 33–35 land |
| 2 | UR-P2-leak-63 | — | fix | merged into UR-P1-leak-66 → phase 3 | verifier; its test becomes row 66's red test |
| 2 | UR-P2-ctr-64 | note | fix | open → phase 3 | node ids of 2^63 and more printed negative by the diagnostics (contract) |

Count: 46 rows (3 held). Not rows here but phase 3 work as well: decisions
19, 20, 28, 33–35 (gates and the window field), the contract bump with a
`docs/contract.md` line (decision 13), decision 24 (CSP builder, `api`),
and the go-fabric issues #543 (err-09) and #544 (err-37).

The script (`p3inv.sh`, run from the register directory):

```sh
#!/bin/sh
# Inventory of every register row that hands work to phase 3.
# Usage: p3inv.sh <dir with phase-*-register.md>
cd "$1" || exit 1
for f in phase-0-register.md phase-1a-register.md phase-1b-register.md phase-2-register.md; do
  awk -v f="$f" -F'|' '
    /^\| *(UR-|\(held\))/ && /phase (1b \/ )?3/ {
      id=$2; sev=$3; disp=$4; where=$5
      gsub(/^ +| +$/,"",id); gsub(/^ +| +$/,"",sev); gsub(/^ +| +$/,"",disp); gsub(/^ +| +$/,"",where)
      unit="fix"
      if (disp ~ /phase 3 `?api/ || where ~ /phase 3 `?api/) unit="api"
      if (disp ~ /`cite`/) unit="cite"
      if (disp ~ /wizard/) unit="wizard"
      if (disp ~ /upstream/) unit="upstream+fix"
      printf "| %s | %s | %s | %s | %s | %s |\n", substr(f,7,length(f)-18), id, sev, unit, disp, where
    }' "$f"
done
```
