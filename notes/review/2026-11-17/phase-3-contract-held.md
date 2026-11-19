# Phase 3 contract shape — the held part

Labels H-a … H-d are those of `notes/briefs/2026-11-18-review-p3-contract.md`.
Fix shapes: `held/phase-3-carryover-held.md:9-27` (decisions 37, 38, 25).

| Label | Artefact |
| --- | --- |
| H-a | 403 problem code `host_not_allowed` |
| H-b | 403 problem code `origin_not_allowed` |
| H-c | `Options.AllowedHosts []string`, default nil |
| H-d | `const RequestHeader = "X-Fabric-UI"` (TS `REQUEST_HEADER`) |

## Order of checks (one place, before the route's handler)

1. Host (H-a) → 2. route: 404 / 405 → 3. on `x-admin` routes: Origin
(H-b) → admin test (`forbidden`) → 4. Content-Type (415) → 5. body:
413 / 400.

## Decision 37 — UR-P0-auth-20 (+ observation, UR-P2-auth-92)

- **Where:** every mode, every route, the bundle of `Handler` /
  `NewHandler` included, HEAD too; before routing and before any status
  writes `csrf_token` (`handler.go:624-630` is the only admin signal
  today; the statuses issue the token, `docs/contract.md:371-372`).
  Under `Mount` it covers the API pattern; the host's own pages are the
  host's.
- **Rule:** take `r.Host`, drop the port (bracketed IPv6 included),
  lower-case, drop one trailing dot. Admit when it (a) is `localhost` or
  ends in `.localhost`, (b) parses as an IP literal (`netip.ParseAddr`,
  no zone), or (c) equals an entry of `AllowedHosts`. An empty `Host` is
  refused. Else 403 `host_not_allowed`, problem+json, no token, no page.
- **`AllowedHosts`:** host names only, compared case-insensitively.
  `NewHandler` / `Mount` return an error (and `Handler` panics) on an
  entry that is empty, has a scheme, port, path, wildcard or white
  space. No opt-out, no `*`. Godoc: "AllowedHosts are the DNS names
  besides loopback names and IP addresses under which requests reach
  the handler; a host behind a proxy lists the name the proxy sends as
  Host."
- **Contract:** `"403"` on all 97 operations (`Problem` component);
  `host_not_allowed` in `Problem.code`; `### problem-host_not_allowed`.
- **Docs:** `docs/embedding.md` (the proxy case), changelog,
  UR-P0-doc-22 / UR-P2-doc-62.

## Decision 38 — UR-P0-web-21 (+ UR-P2-web-93, -94)

- **One admin test** (decision 45), for routes, the admin query
  parameter and the node stream's filter, and for `Status.admin` /
  `ControllerStatus.admin`:
  - host mode: `X-Fabric-UI: 1` present (exactly `1`) **and** `Admin(r)`;
    the header is tested first;
  - token mode: `X-Fabric-UI-Token` equals the token, as today;
    `X-Fabric-UI` is not required there (the shipped client sends both
    on every request; the token only to a same-origin base).
  A missing header is 403 `forbidden` on an `x-admin` route (GET and
  HEAD included); on an open route the request is a non-admin's.
- **Origin:** on an `x-admin` route, when `Origin` is present its host
  must pass rule (a)–(c) above (the list, not equality with `Host`:
  UR-P2-web-94's proxy case); `null` or unparsable is refused. Else 403
  `origin_not_allowed`. Both modes. On an open route a failing Origin
  only makes the request a non-admin's. `Sec-Fetch-Site` is not read.
- **415:** a request with a non-empty body whose `Content-Type` media
  type is not `application/json` (case-insensitive, parameters allowed)
  answers 415 `unsupported_media_type`; a missing `Content-Type` with a
  body too. Declared on the 26 body operations.
- **Contract:** `x-admin` paragraph of `info.description`
  (`openapi.yaml:19-21`) names both headers; `origin_not_allowed`,
  `unsupported_media_type` in `Problem.code`.
- **Docs:** the paragraph on what `Admin` must guarantee
  (`docs/contract.md:374-383`, `docs/embedding.md`); `ui/README.md`: a
  host's own client and its own `EventSourceFactory` send the header.

## UR-P1-adpt-60 (decision 25)

No contract artefact: `CommissionDiscovered` defaults the attestation
policy to reject before `po.Edit`; the refusal is the existing
commissioning failure (502 `node_error`, `commissioning_progress`
`failed`). If G needs a new failure kind, stop and ask.

## Red tests that must turn green (`held/phase-2-X4.md`)

| Test | Line | Row |
| --- | --- | --- |
| `TestX4Held_Auth20_TokenModeARequestUnderAForeignHostGetsNoToken` | 89 | auth-20 |
| `TestX4Held_Auth20_EveryStatusRouteIssuesTheTokenUnderAForeignHost` | 129 | auth-20 |
| `TestX4Held_Auth20Obs_HostModeAForeignHostReadsEverythingUngated` | 196 | observation |
| Playwright "RED UR-P0-auth-20: a page under a rebound name reads the token, the passcode, and resets the node" | 356 | auth-20 |
| Playwright "RED UR-P0-auth-20 (observation): the bundle itself is served under the rebound name and runs as admin" | 406 | observation |
| `TestX4Held_Web21_HostModeACrossSiteWriteIsAdmitted` (six subtests) | 476 | web-21 |
| Playwright "RED UR-P0-web-21: a text/plain form of another site erases the node (host mode)", "… a no-cors fetch of another site opens a window and starts an update check (host mode)" | 625, 646 | web-21 |
| `TestX4_Auth92_HostModeAddressAdminARebindingPageReadsThePairingSecret` (+ Playwright "RED UR-P2-auth-92 …") | 737, 770 | auth-92 |
| `TestX4_Web93_HostModeACrossSiteGETWithAnEffectIsAdmitted` | 812 | web-93 |
| `TestEndToEnd_CommissionDiscoveredRejectsFailedAttestationByDefault` | `held/phase-3-carryover-held.md:27` | adpt-60 |

Stays green: `TestX4Held_Auth20_ControlsThatHold` (141). The two
"control" subtests of the Web21 test (527-541) expect 400 today and
answer 403 `origin_not_allowed` afterwards: change their expectation.
Every existing host-mode test that calls an admin route gains the
header (one helper).

## Open (held)

- **Ingress.** Whether the pinned Ingress host receives the browser's
  `Host` or the proxy's own was not checked here. If it is the
  browser's, that host cannot know the names and "no opt-out" bites.
  Recommendation: G measures it at the pinned host before coding; if it
  is the browser's name, back to the owner (a per-request
  `func(host string) bool` beside the list is the smallest way out).
- **IP-literal Origin.** Rule (b) admits any IP-literal origin, as
  decided; the header requirement is what stops a cross-site page then.
  No change proposed; the fuzz targets of decision 53 cover the host and
  origin parsing and the media type.

## Measurement 2026-11-19 (G, coordinator) → decision 61

HA core `homeassistant/components/hassio/ingress.py:266-272` (`_init_header`, filter `:28-37`) and Supervisor `supervisor/api/ingress.py` (`_init_header`, `_REQUEST_HEADERS_FILTER`, main, read via raw.githubusercontent) both forward the browser's `Host` unchanged; Supervisor only appends X-Forwarded-For. So an Ingress add-on sees `homeassistant.local:8123` or a remote-access domain and cannot list it. Owner (decision 61): a per-request predicate in `Options` beside the allow-list admits a request on the host's own trusted signal (e.g. the Supervisor's source address); the check stays mode-independent and on every route.

## Decision 61 shape

Label **H-e**: `Options.AllowHost func(r *http.Request) bool`, default nil.
It takes the request, not the name: the host's trusted signal is
usually not the name (`r.RemoteAddr`, a header its proxy sets).

- **Combination (rule d).** Rules a–c first, as above. Only when none
  admits and `AllowHost` is non-nil it is called, once per request;
  true admits, without the list (`AllowedHosts` may be empty). False,
  or nil, is 403 `host_not_allowed` as before. An empty `Host` is
  refused before the predicate is asked. A panic in it is not caught.
- **nil** means the list only: decision 37's shape unchanged.
- **Every mode, every route**: the predicate is part of the one Host
  check (step 1 of the order), so token mode, host mode, the bundle,
  HEAD and the streams run it; it has no opt-out value — a host that
  returns true always has written its own opt-out, and the godoc says
  what that means.
- **Origin (decision 38) follows.** Behind such a proxy `Origin` names
  the same unlisted name. On an `x-admin` route an `Origin` passes when
  its host meets a–c, **or** when rule d admitted this request and the
  Origin's host and port equal `r.Host` exactly (case-insensitive).
  Any other Origin on a request admitted by d is 403
  `origin_not_allowed`. Without this the Ingress page's own writes
  fail; with it a page of another name through the same proxy is still
  refused.
- **Admin is separate.** `AllowHost` true says nothing about admin
  rights; `Admin(r)` and the request header are tested as before.
- **Validation:** none (a func); nothing new for `NewHandler` to refuse.
- **Godoc (neutral):** "AllowHost admits a request whose Host is neither
  a loopback name, an IP address nor in AllowedHosts. A host behind a
  proxy that forwards the browser's Host unchanged cannot list the
  names; it answers here from what only its proxy can produce, such as
  the proxy's source address. It is asked once per request, on every
  route and in every mode, and must not answer from the Host header or
  anything else the browser chooses. Nil admits by name only."
- **Docs:** `docs/embedding.md`, the proxy paragraph: list the name when
  the proxy rewrites `Host`, set `AllowHost` when it forwards it.

### Tests G adds (red first, beside the X4 held tests)

| Test | Asserts |
| --- | --- |
| `TestAllowHost_PredicateAdmitsAnUnlistedHost` | foreign `Host`, predicate true by `RemoteAddr`: `GET /node/status` 200 with the token (token mode) and 200 (host mode); the predicate ran once |
| `TestAllowHost_PredicateRefuses` | predicate false: 403 `host_not_allowed`, no `csrf_token` in the body, on status, an admin route and `GET /` |
| `TestAllowHost_NilMeansTheListOnly` | nil: listed name 200, loopback 200, IP literal 200, unlisted 403; the predicate is not needed for a–c (a counting predicate stays at 0 for them) |
| `TestAllowHost_EveryModeAndRoute` | table over token mode / host mode × every entry of `nodeRoutes`, `registryRoutes`, the operation and controller routes, the four streams (HEAD) and the bundle: unlisted `Host` with predicate false is 403 everywhere; the route list is taken from the contract so a new route cannot miss it |
| `TestAllowHost_OriginUnderAnAdmittedHost` | admitted by d: `Origin` equal to `Host` passes an admin write; another `Origin` is 403 `origin_not_allowed`; an Origin under a–c passes as before |
| `TestAllowHost_EmptyHostIsRefusedUnasked` | empty `Host`: 403, predicate call count 0 |

The X4 held tests stay as listed above: none sets the predicate, so each
still turns green by rules a–c alone. Fuzz (decision 53): the host
parser's target runs with a predicate that records its calls.

### Public note — exact change for G

In `notes/briefs/2026-11-18-review-p3-contract.md`, section 5, replace
the table row

`| **H-c** | see held/ | see held/ | held row UR-P0-auth-20 |`

with

`| **H-c**, **H-e** | see held/ | see held/ | held row UR-P0-auth-20 (decisions 37, 61): two fields |`

and in the header paragraph "**H-a … H-d**" becomes "**H-a … H-e**".
No other public line changes; no status, problem code or contract
shape is new.
