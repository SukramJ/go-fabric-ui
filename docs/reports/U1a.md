# Report U1a — list-valued fields and host warnings (go-fabric-ui v0.1.1)

Brief: `notes/briefs/2026-10-25-U1a-fabric-ui-patch.md`. Source: go-ha2fabric's
report H3 (§Questions Q2, §Side effect). Additions only; contract v1 keeps
every field.

## Per section: what landed

### §1 List-valued fields — `e62f102`

- `Column.Multi` and `Filter.Multi` (`fabricui/provider.go`, JSON `multi`,
  `omitempty`, so a v0.1.0 schema answers byte for byte as before).
- Contract: `multi` on `Column` and `Filter`, optional (no `default:` in the
  YAML, which would have made the generated TS field required; the default
  is stated in the description).
- Page: `ui/src/lib/pages/node/fields.ts` (`listValues`, `fieldMatches`,
  `distinctValues`); `Selection.svelte` renders a multi column as chips,
  builds a multi column filter's options from the individual values, and
  matches multi filters (column or fixed) by membership, trimmed, an empty
  value never matching.
- Fixtures: the in-memory provider gained a `labels` column (`Filter`,
  `Multi`) with `"Evening, Ceiling"`, `"Evening"` and `""`; the schema,
  list and preview fixtures re-recorded.
- Tests: Go `TestExposureSchema_MultiIsOptional`; vitest `fields.test.ts`
  and two Selection cases (chips plus membership filter with individual
  options; fixed filter with `multi` true vs. false).
- Docs: `docs/provider.md` §Schema paragraph.

### §2 Host warnings — `e62f102`

- `Options.Warnings func(ctx) []Warning`, `Warning{Code, Message string;
  Since time.Time}` (`since` with `omitzero`). `GET /node/status` answers
  `warnings` (`omitzero`): absent without the hook, `[]` when the hook
  returns nil or empty, and reported also when the host serves no node.
- Contract: `Status.warnings`, schema `Warning`; the client exports the
  `Warning` type.
- Page: `HostWarnings.svelte` — a card on the Status page (code as a
  warning chip, message as text, "since" when set), a collapsed `<details>`
  above the Selection list. Selection loads the status itself when no one
  has yet (a host that mounts the page alone).
- Fixture `node-status.200` carries two warnings (one with `since`).
- Tests: Go `TestStatus_Warnings` (absent / `[]` / without a node) plus the
  fixture test; vitest Status (card above the status card, no card without
  warnings) and Selection (collapsed list).
- Docs: `docs/embedding.md` Options row and paragraph; `docs/contract.md`
  route table.

### Release — `18b2ad6`

Version `0.1.1` in `internal/version`, `ui/package.json`,
`ui/package-lock.json`, `contract/openapi.yaml` `info.version`,
`changelog.md`; README and CLAUDE.md state lines; `contract/previous/`
refreshed. `make contract-check` was run against the v0.1.0 copy before the
refresh (exit 0) and again after.

## Guards (HEAD 18b2ad6, clean tree)

| Guard | Result |
| --- | --- |
| `make ci` (build vet test lint web-check dist-check) | exit 0 |
| `make race` | exit 0 |
| `make contract-check` (before refresh, against v0.1.0) | exit 0 |
| `make contract-check` (after refresh) | exit 0 |
| `make version-check` | exit 0 |

## Differences from the brief

- "A card above the fabrics card": the Status page has no fabrics card; its
  one card is the status card, which carries the fabric count. The warnings
  card sits above it. On a host without a node, it sits above the
  "disabled" card.
- §1 and §2 share one commit: they touch the same files (contract,
  generated types, Selection page, test support, bundle), so splitting them
  would have meant committing an intermediate bundle.
- The Selection page now calls `loadStatus` on mount when the status is not
  loaded; before, it read the status only if someone else had loaded it.

## Questions

None.

## by_design candidates

None.
