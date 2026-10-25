# Brief U1a — go-fabric-ui v0.1.1: list-valued filter fields, host warnings on the status page

Date 2026-10-25. Repository `/home/markus/Dokumente/GitHub/go-fabric-ui`
(main = v0.1.0). Source: go-ha2fabric's report `docs/reports/H3.md`
(§Questions Q2, §Side effect) — the first consumer's findings. Read the
repository's `CLAUDE.md`, `docs/contract.md`, `docs/provider.md`,
`fabricui/provider.go`, `fabricui/handlers_node.go` (status), the Selection
page and its filter code under `ui/src/lib/pages/node/`, `contract/openapi.yaml`.

**Model:** Opus. Branch `u1a/filters-and-warnings`, one PR, no auto-merge,
no label, no attribution lines, English. Version `0.1.1` in
`internal/version`, `ui/package.json`, `changelog.md` (`make version-check`);
`contract/previous/` refreshed in the release commit as the release
checklist says. Additions only: the contract guard must pass without a
removed field.

## 1 List-valued fields (H3 Q2)

An `Item.Fields` value is one string; a multi-valued property (labels)
is a comma list and the filter compares by equality, so "Kitchen, Matter"
is its own filter value. Add `Column.Multi bool` and `Filter.Multi bool`
(JSON `multi`, default false): a multi column renders its value as chips,
and a multi filter matches by membership in the comma-separated list
(trim spaces; an empty item never matches). The filter's option list for
a multi field is built from the individual values across all items, not
the combinations. Contract: the two optional fields; fixtures: one item
with a multi column; TS and Go tests; `docs/provider.md` paragraph.

## 2 Host warnings on the status page

A provider's configuration can be half-valid (a label name that resolves
to nothing); today a host can only log it. Add `Options.Warnings func(ctx)
[]Warning` with `Warning{Code, Message string; Since time.Time}`; the
status response gains `warnings: []` (optional, omitted when the hook is
nil); the Status page shows them in a card above the fabrics card, with
the code as a chip and the message as text; the Selection page shows the
same list collapsed above the table. Fixtures, Go handler test, vitest
test, `docs/embedding.md` paragraph.

## 3 Report

`docs/reports/U1a.md` (PR body): per § with commit hashes; guards with
HEAD (`make ci` incl. `web-check`, `dist-check`, `contract-check`,
`race`); differences; questions. Hand back with the PR number and HEAD.
Copy this brief to `notes/briefs/` in the first commit. If a permission is
denied, say which and stop; never ask the launcher to perform it.
