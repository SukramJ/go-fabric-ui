# M1: the problem types of go-fabric-ui (for thingfabric.dev/go-fabric-ui/problems)

For the coordinator (brief M1 row 6). From contract 0.14.0 every problem
answers `type: https://thingfabric.dev/go-fabric-ui/problems#problem-<code>`
(`fabricui/problem.go`, `problemTypeBase`); the page at that address needs
one anchor per row below, named exactly as the anchor column, so every
`type` resolves. The text of each entry is `docs/contract.md`, section
"Problems" (the heading column is the heading there, an `###` under
`## Problems`). The page title column is a proposal: the handlers send a
per-route `title` (examples in the last column, from `fabricui/*.go`), so
the page names the type, not one route's title.

| Anchor | Page title (proposal) | HTTP status | `docs/contract.md` heading | Handler titles sent (examples) |
| --- | --- | --- | --- | --- |
| `problem-bad_request` | Bad request | 400 | `### problem-bad_request` | Invalid request, Invalid request body, Invalid request parameter, Invalid label, Invalid fabric index, Invalid log level, Erase not confirmed, Factory reset not confirmed, duration_seconds out of range |
| `problem-forbidden` | Admin rights required | 403 | `### problem-forbidden` | Admin rights required |
| `problem-host_not_allowed` | Host not allowed | 403 (0.10.0) | `### problem-host_not_allowed` | Host not allowed |
| `problem-origin_not_allowed` | Origin not allowed | 403 (0.10.0) | `### problem-origin_not_allowed` | Origin not allowed |
| `problem-method_not_allowed` | Method not allowed | 405 (0.10.0) | `### problem-method_not_allowed` | Method not allowed |
| `problem-unsupported_media_type` | Unsupported media type | 415 (0.10.0) | `### problem-unsupported_media_type` | Unsupported media type |
| `problem-too_many_streams` | Too many streams | 503 (0.10.0), `Retry-After: 5` | `### problem-too_many_streams` | Too many streams |
| `problem-conflict` | Conflict | 409 | `### problem-conflict` | Refused by the node; the pairing routes' own titles ("a commissioning window is already open") |
| `problem-icd_multi_admin` | Other administrators may not support LIT | 409 (0.12.0) | `### problem-icd_multi_admin` | Other administrators may not support LIT |
| `problem-not_found` | Not found | 404 | `### problem-not_found` | No such route, No such node, No such endpoint, No such fabric, No such cluster, No such credential, No device answered |
| `problem-service_unready` | Service not ready | 503 | `### problem-service_unready` | Not read yet, Node not started, Node not assembled, Node not available, Node registry not available, No credential store, No OTA provider, Operation not available, Pairing codes not available, Reassembly not available, Selection not available |
| `problem-internal` | Internal error | 500 (no detail) | `### problem-internal` | the failing operation's title; the cause goes to the log only |
| `problem-payload_too_large` | Request body too large | 413 | `### problem-payload_too_large` | Request body too large |
| `problem-node_error` | The node failed | 502 | `### problem-node_error` | The node failed |
| `problem-node_limit` | Node limit reached | 409 (0.6.0) | `### problem-node_limit` | Node limit reached |
| `problem-node_is_default` | Node is configuration | 409 (0.6.0) | `### problem-node_is_default` | Node is configuration |
| `problem-terms_required` | Terms and conditions required | 409 (0.8.0); members `version`, `acknowledgements`, `deadline` | `### problem-terms_required` | Terms and conditions required |
| `problem-not_implemented` | Not implemented | 501 | `### problem-not_implemented` | Not implemented, Erase not available |
| `problem-onboarding_code_invalid` | Not an onboarding code | 400 (0.13.0) | `### problem-onboarding_code_invalid` | Not an onboarding code |
| `problem-onboarding_code_mismatch` | The code names another device | 409 (0.13.0); members `field`, `code_value`, `node_value` | `### problem-onboarding_code_mismatch` | The code names another device |

20 types; the same 20 codes as the `Problem.code` enum of
`contract/openapi.yaml` (0.14.0). Every problem also carries `code` and the
header `X-Problem-Code`; a client matches on those, not on `type`. The old
base, served until v0.13.0, was
`https://github.com/SukramJ/go-fabric-ui/blob/main/docs/contract.md#problem-`
(the same anchors, on `docs/contract.md`); it keeps resolving through
GitHub's redirect to `github.com/thingfabric/go-fabric-ui`.
