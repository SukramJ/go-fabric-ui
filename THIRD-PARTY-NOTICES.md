# Third-party notices

This repository is Apache-2.0 (see `LICENSE`). Third-party code it ships is
listed here with its licence.

## npm packages in the built bundle

The committed bundle `fabricui/dist/` (and so every binary that embeds the
Go module) carries these packages, compiled into its `vendor` chunk. The
list is checked in CI against the chunk itself (`make notices-check`:
the build writes the chunk's package list and `ui/scripts/check-notices.mjs`
fails on a package missing here or without its licence text). The full
licence texts are under `licenses/npm/`.

| Package | Version | Licence | Text |
| --- | --- | --- | --- |
| `@lucide/svelte` | 1.57.0 | ISC | [`licenses/npm/lucide-svelte.txt`](./licenses/npm/lucide-svelte.txt) |
| `clsx` | 2.1.1 | MIT | [`licenses/npm/clsx.txt`](./licenses/npm/clsx.txt) |
| `dijkstrajs` | 1.0.3 | MIT | [`licenses/npm/dijkstrajs.txt`](./licenses/npm/dijkstrajs.txt) |
| `qrcode` | 1.5.4 | MIT | [`licenses/npm/qrcode.txt`](./licenses/npm/qrcode.txt) |
| `svelte` | 5.57.2 | MIT | [`licenses/npm/svelte.txt`](./licenses/npm/svelte.txt) |
| `tailwind-merge` | 3.7.0 | MIT | [`licenses/npm/tailwind-merge.txt`](./licenses/npm/tailwind-merge.txt) |

The npm package `@sukramj/fabric-ui` itself ships sources, not this bundle;
its runtime dependencies (`@lucide/svelte`, `clsx`, `qrcode`,
`tailwind-merge`; `svelte` is a peer dependency) are installed by the
consumer's package manager with their own licence files. Build-time
tools (Vite, Tailwind CSS, TypeScript, the test stack) are not shipped.

## Go modules

The Go module depends on go-fabric and, through it, on the modules below.
They are fetched by the Go toolchain from their own repositories with
their own licence files; none is vendored here. `make licenses` (go-licenses)
fails on a forbidden, restricted or reciprocal licence.

| Module | Licence |
| --- | --- |
| `github.com/SukramJ/go-fabric` | Apache-2.0 |
| `filippo.io/nistec` | BSD-3-Clause |
| `github.com/cenkalti/backoff` | MIT |
| `github.com/dustin/go-humanize` | MIT |
| `github.com/google/uuid` | BSD-3-Clause |
| `github.com/grandcat/zeroconf` | MIT |
| `github.com/mattn/go-isatty` | MIT |
| `github.com/miekg/dns` | BSD-3-Clause |
| `github.com/ncruces/go-strftime` | MIT |
| `github.com/remyoudompheng/bigfft` | BSD-3-Clause |
| `golang.org/x/net` | BSD-3-Clause |
| `golang.org/x/sys` | BSD-3-Clause |
| `modernc.org/libc` | BSD-3-Clause |
| `modernc.org/mathutil` | BSD-3-Clause |
| `modernc.org/memory` | BSD-3-Clause |
| `modernc.org/sqlite` | BSD-3-Clause |

## Ported source

The node pages, their handlers, the UI kit, the design tokens and the
theme bridge are ported from [openccu-loom](https://github.com/SukramJ/openccu-loom)
at commit `b40b289a` (Apache-2.0, same owner); every ported file names its
origin with the line range at that commit.
