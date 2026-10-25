# Embedding

go-fabric-ui ships twice in one version: a Go module with the handlers and
the built bundle, and an npm package with the pages for a host that has a
single-page app of its own.

## A daemon without a web app: the Go module

```go
import "github.com/SukramJ/go-fabric-ui/fabricui"

ev := fabricui.NewEvents()
cfg.Observers = ev.Observers(cfg.Observers) // before node.New
n, err := node.New(cfg)
// …
mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Exposure: provider, Events: ev}))
```

`Handler` serves the API under `/api/v1` and the standalone shell at `/`
from the bundle embedded in the module (`fabricui/dist/`, built and
committed here): no Node.js toolchain in the daemon's build, no second
image. Any path that is not a file of the bundle answers the shell's index;
an unknown path under the API prefix answers a 404 problem.

`Options`:

| Field | Purpose |
| --- | --- |
| `Node` | The node (`*node.Node` satisfies `NodeSource`); nil serves the family as disabled. |
| `Exposure` | The host's `ExposureProvider` ([provider.md](provider.md)); nil hides the selection page. |
| `Admin` | Who may write and read the pairing secret; nil selects token mode ([contract.md](contract.md#admin-rights)). |
| `Reassemble` | Rebuilds the endpoints from the current selection (maintenance route, and after an exposure change). |
| `Events` | Events beside the handler's own; an `*Events` wired into `node.Config.Observers` carries fabric and window transitions. |
| `Prefix` | The API prefix; default `/api/v1`. |
| `Logger` | The handlers' logger; nothing it receives carries a passcode. |
| `ConfiguredWindow` | Open the window on the node's configured passcode instead of fresh credentials. |
| `AllowEmptyTopology` | Allow a window while no endpoint is exposed. |
| `Warnings` | The host's findings about its own configuration, shown on the status and selection pages. |

A provider's configuration can be half-valid — a label name that resolves
to nothing, a rule that matches no item — and a log line is where nobody
looks. `Options.Warnings` returns those findings as `Warning{Code,
Message, Since}`; it is called on every `GET /node/status`, which answers
them as `warnings` (absent without the hook, `[]` when it finds nothing,
also when the host serves no node). The status page shows them in a card
above the status, the code as a chip and the message as text; the
selection page shows the same list collapsed above the table. `Code` is
the host's stable identifier for the kind of finding, `Message` is already
in the host's language, and a zero `Since` is left out.

`node.Config.Observers` is set before `node.New`; the facade offers no
registration afterwards, so a host that wants the stream to report
fabrics and windows wires `Events.Observers` there. Without it the stream
still reports what the handler's own routes do (a removed fabric, an
opened or closed window), but not a controller commissioning the node.

A daemon with a mux of its own and a different layout mounts only the API:

```go
if err := fabricui.Mount(mux, fabricui.Options{Node: n, Admin: isAdmin, Prefix: "/api/v1"}); err != nil {
    return err
}
```

## A host with its own single-page app: the npm package

```sh
npm install @sukramj/fabric-ui
```

The package ships Svelte 5 sources (pages, kit, client, translator, theme);
the host's Vite build compiles them. See
[`ui/README.md`](https://github.com/SukramJ/go-fabric-ui/blob/main/ui/README.md)
for the context the pages read (`provideFabricUI`), the stylesheet, and
feeding the event stream into a host's own hub. The host serves the API
with `fabricui.Mount`.

## Behind an Ingress

The bundle and the shell are built for a proxy that serves them under a
path prefix it chooses:

- **Relative paths.** The bundle's base is `./`, the shell routes by hash
  (`#/node/pair`), and the client calls `./api/v1`, so the page works under
  any prefix without configuration.
- **Theme bridge.** Inside Home Assistant's Ingress frame the page is same
  origin with Home Assistant's frontend; `startHaBridge` copies its live
  theme variables onto the page and follows its light and dark mode. Under
  any other origin the bridge stays inactive and the tokens' Home
  Assistant defaults apply.
- **Admin gate by the host.** The proxy authenticates the user; the host's
  `Options.Admin` admits a request when it came through the proxy (for
  example by the proxy's source address). Token mode is for a daemon that
  serves the page directly.
