# Phase 1a — verification

Independent verifier (Fable), HEAD `49cc081`, branch `review/p1a`, go-fabric
`v0.22.2` (the pinned export, `gf:` below). Scope: the four mediums in full (one held back),
six lows as the sample (decision 17). **M** = run as a temporary test
(copied into the package, run, deleted; sources kept under
`scratchpad/verify/`) or read at the line where the line alone decides;
**I** = inferred. `git status` afterwards shows only the two untracked notes
files that were there before (`notes/briefs/2026-11-18-review-p1a.md`,
`notes/reads/read-review-p1a-inventory.md`).

## Verdicts

| ID | Reviewer's sev | Verdict | My sev | Reason, `file:line` (M/I) |
| --- | --- | --- | --- | --- |
| UR-P1-conc-01 | medium | confirmed | medium | `fabricui/stream.go:31-86`; routes `handler.go:313`, `:448`, `internal/api/api.gen.go:3047`. **M**, all three streams: `HEAD` answers 200 and runs `serveSSE`; after one more byte on the connection and a close, the subscriber (node stream) and the goroutine (all three) are still there, `GET` under the same input ends. Without the extra byte the `HEAD` loop ends on close. Unauthenticated, permanent per request, but small per leak: medium, not high |
| UR-P1-adpt-30 | medium | confirmed | medium | `gf:diagevent/ring.go:210-231`, `gf:controller/diagnostics.go:134-142` (newest first); `fabricui/gofabric/diagnostics.go:85-94` passes through, says "oldest first"; `fabricui/handlers_controller.go:1849` reverses; contract `openapi.yaml:1845-1849`. **M** (read; the lines decide). `GET /node/events` is right (`handlers_diagnostics.go:379-384`, no reversal), so the two traces disagree on one page family |
| UR-P1-err-31 | medium | confirmed, **cause corrected** | low | Consequence holds (`handlers_controller.go:1172-1176`, `:1219`, `:879-883`): JSON `null` unmarshals to 0 / empty without an error. The stated precondition does not: `client.DecodeJSON` cannot fail at the pin (`gf:internal/valuejson/valuejson.go:56-64`, a mismatching shape is rendered untyped, `:80-133`), so `gofabric/node.go:462-465` is dead code. `null` arrives another way: a node that reports **TLV null** for the attribute (`valuejson.go:82`, `:150`). **M** (read) for both halves; end to end **I**. Only a non-conformant node does that, and a hostile one gains nothing by it: low |
| UR-P1-err-04 | low | confirmed | low | `fabricui/handlers_pairing.go:77-88`, `:97-110`. **M** (read): `win.Kind` is used only to forget the ephemeral; any other open window answers `source: configured, window_open: true`. Right for a window opened on the configured credentials, wrong for an enhanced or a foreign basic one |
| UR-P1-conc-06 | low | confirmed | low | `fabricui/registry.go:125-127`, `:165-176` (only `config` refreshed on a hit); `handler.go:491-492`; `handlers_node.go:138`. **M** (read). The doc (`registry.go:48`) names the three compared fields and is silent on the two frozen ones |
| UR-P1-adpt-34 | low | confirmed | low | `fabricui/handlers_controller.go:939` runs before any look at `mode`; `:943` then drops the entry. **M** (read) |
| UR-P1-adpt-35 | low | confirmed | low | `fabricui/handlers_node.go:455-461` admits 65535; `gf:node/served.go:176-180`, `gf:im/eventlog.go:443`, `:468`. **M** (read). Harmless (the log is unguarded either way); note would also be defensible |
| UR-P1-conc-61 | low | confirmed | low | `fabricui/gofabric/gofabric.go:566`, `:573-583`; `gf:controller/client/label.go:128-133` sets the captured label again. **M** (read); the reviewer's rate (138–187 of 300) not re-run |
| UR-P1-err-63 | low | confirmed in part | low | `ErrMismatch` → `nodeError` → 502: `fabricui/gofabric/updates.go:88-96` (the report's `:195-202` is the wrong range), `gofabric.go:613-626`, `gf:controller/softwareupdate/manager.go:1247-1249`. **M** (read). The `ErrTransferOptions` half is unreachable: the adapter's `Target` sets neither option (`updates.go:94`, `gf:…/manager.go:78-86`) |

False positives in the sample: none of six lows struck; one of four mediums
lowered (err-31), none struck.

Duplicates noticed while reading:

- UR-P1-doc-41 (G-go-A) and UR-P1-doc-69 (G-go-B): one root cause, CLAUDE.md's import lists against `internal/arch/arch_test.go`; one fix.
- UR-P1-err-10 (S-go) and UR-P1-adpt-65 (G-go-B): two defects at one site (`handlers_controller.go:1304-1307`, the optional `long_discriminator` narrowed to `uint16` with 0 as "absent"); fix together (carry the pointer or a `set` flag, range-check 0–4095).
- UR-P1-err-31's third fix item ("the adapter answers a decode failure as an error") has no reachable case at the pin; see above.

## The four mediums

### UR-P1-conc-01 — confirmed, medium

Go's mux matches `HEAD` on a `GET` pattern, and net/http discards a `HEAD`
response's body without an error, so `stream.go:61`, `:78` and `:82` never
fail; only `ctx.Done()` (`:58`) ends the loop. Probe (node stream): `HEAD`,
then one byte or a whole pipelined request, then close → subscriber count 1
two seconds and 100 heartbeats later; `GET` with the same input → 0; `HEAD`
without the extra byte → 0. Same probe by goroutine count on
`/controller/stream` and `/operation/log/stream`: +1 after `HEAD`, +0 after
`GET`. `httptest.Server.Close` reports the connection still `active`. A
plain `http.Client` `HEAD` also holds the subscription until the client
drops its idle connection. No rights needed. Not high: each request leaks a
goroutine, a ticker and a 64-slot channel, not a secret or a crash; it does
turn any later cap on open streams (UR-P0-bound-29) into a permanent
refusal, so the method check must land with or before that cap.

Red test: `TestStream_HEADDoesNotSubscribe` — raw connection, `HEAD
…/node/stream`, one further byte, close; assert `len(h.hub.subs) == 0`
within a second (and the same by goroutine count for the controller and log
streams).

### UR-P1-adpt-30 — confirmed, medium

`gf:diagevent/ring.go:210` "newest first", index `next-1-i` at `:227`;
`gf:controller/diagnostics.go:134-142` returns that snapshot. The adapter
copies it in order (`gofabric/diagnostics.go:88-92`) under a comment that
says oldest first (`:85-86`), the port promises oldest first
(`controller.go:71-72`), and the handler reverses (`handlers_controller.go:1849`).
With the real controller the route is oldest first against
`openapi.yaml:1845-1849`; the page shows the list as served
(`ui/src/lib/pages/controller/Diagnostics.svelte:60`, `:254`). The memory
controller is oldest first, so every handler test passes, and the e2e test
checks membership only (`gofabric/e2e_test.go:311-319`). Medium by the
scale: wrong behaviour every user of a real host meets; nothing is lost.

Red test: in `TestEndToEnd_DetailsAndEvents` (or a new
`TestEndToEnd_ControllerEventsNewestFirst`), after two trace entries, assert
`events[i].at >= events[i+1].at` for the whole list and that the first entry
is the last one recorded.

### UR-P1-err-31 — confirmed, cause corrected, low

What the handlers do with `null` is as reported: `ownIndex` returns 0 with a
nil error (`handlers_controller.go:1172-1176`), the guard at `:1219` compares
against an index that `:1212` already excludes, `GetNodeACL` answers
`{"entries":[]}` (`:879-883`). What is not as reported is how `null` gets
there. `valuejson.Decode` (`gf:internal/valuejson/valuejson.go:56-64`) only
marshals maps, slices, strings, bools and `json.Number`s built from parsed
integers and finite floats; a value whose shape does not match the
definition falls through to the untyped rendering (`:133`). It has no
failing path, so the adapter's `raw = null` branch (`gofabric/node.go:462-465`)
is never taken at the pin. The reachable path is a node that answers the
read with a **TLV null** (`valuejson.go:82`): rendered as JSON `null`,
Status 0. A mismatching shape instead (a string, a struct) reaches
`json.Unmarshal` as that shape and is the intended 502. So the precondition
is narrower: a node that reports null for the non-nullable
CurrentFabricIndex, ACL, Fabrics or Binding. A conformant node never does; a
hostile node loses nothing it does not already control. That is an edge-case
defect with a destructive outcome on a broken node, hence low. The fix
stands minus its adapter item: `readOne` refuses `null` for these four as
`ErrNodeFailed`, `ownIndex` refuses 0.

Red test: `TestRemoveNodeFabric_NullCurrentFabricIndexIsNodeError` — a
`ControllerNode` wrapper returning `null` for CurrentFabricIndex: `DELETE
…/fabrics/1` answers 502 and no Invoke is sent; `TestGetNodeACL_NullIs502`
likewise.

### (held) — confirmed, medium; held back until its fix (decision 25)

