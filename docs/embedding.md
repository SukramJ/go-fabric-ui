# Embedding

go-fabric-ui ships twice in one version: a Go module with the handlers and
the built bundle, and an npm package with the pages for a host that has a
single-page app of its own.

## A daemon without a web app: the Go module

```go
import "github.com/SukramJ/go-fabric-ui/fabricui"

n, err := node.New(cfg)
// … n.Start(ctx) — before or after the handler is built
mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Exposure: provider}))
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
| `Exposure` | The host's `ExposureProvider` ([provider.md](provider.md)); nil hides the selection page. A list-valued (`Multi`) column reads `Item.Lists`, one entry per value; the comma-separated `Item.Fields` form is deprecated. |
| `Controller` | The controller family's source (`gofabric.ControllerOf` over a `*controller.Controller`, see below); nil serves no controller family and every `/controller/*` path answers 404. |
| `Admin` | Who may write, read the pairing secret and use a controller route marked admin; nil selects token mode ([contract.md](contract.md#admin-rights)). |
| `Reassemble` | Rebuilds the endpoints from the current selection (maintenance route, and after an exposure change). |
| `Events` | Events the host publishes beside the handler's own and the node's (an exposure change made outside the page). |
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

The node's transitions need no wiring: each open stream registers an
observer with `Node.Observe` and removes it when the connection closes, so
a handler built after `Node.Start` reports a controller commissioning the
node, a fabric removed over the wire or from a route, every window
opened or closed, and every endpoint whose surface an assembly changed
(`surface_changed`, from `Observers.OnSurfaceChanged`) — each once, whoever
caused it, a reassembly the host runs on its own included.
`node.Config.Observers` stays the host's own. A host that carries these
events over a channel of its own registers `fabricui.NodeObservers` with
`Node.Observe` the same way; publishing them into `Options.Events` instead
would hand every stream each one twice.

Everything the handlers read is on the node facade: the window from
`Node.Window` (its kind, discriminator and deadline on the node's clock),
the configured codes from `Node.SetupPayload`, the clusters of every
endpoint, parts included, from `Topology`, fresh window credentials from
`node.NewWindowCredentials`, the session table from
`Node.SessionDiagnostics`, and the surface summary of the status (level,
profiles, extensions and counts) from `Node.Surface`. A host wrapping its node in its own
`NodeSource` implements those methods.

## The controller family

A daemon that runs go-fabric's controller hands it to the handlers through
the adapter in `fabricui/gofabric`, the one package of this module that
knows the controller facade:

```go
import (
    "github.com/SukramJ/go-fabric-ui/fabricui"
    "github.com/SukramJ/go-fabric-ui/fabricui/gofabric"
)

hooks := gofabric.NewHooks()
cfg.HooksFor = hooks.Wrap(cfg.HooksFor) // before controller.New; keeps the host's own hooks
ctrl, err := controller.New(ctx, cfg)
// …
src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{
    Hooks:       hooks,
    Credentials: store,                                    // the host's Wi-Fi and Thread store
    Pair:        controller.PairOptions{OnAttestationFailure: policy},
})
mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Controller: src}))
// … on shutdown: hooks.Close() before ctrl.Close(ctx)
```

`ControllerOptions`:

| Field | Purpose |
| --- | --- |
| `Hooks` | Turns the nodes' `client.Hooks` into the controller stream's events (`node_added`, `node_removed`, `connection`, `subscription`, `node_updated`, coalesced). `Wrap` chains them after the host's own `HooksFor`; `Close` (before the controller's) stops the pending timers; events after it, or while the controller has no registry, are dropped. Nil leaves the stream with the handler's own events. |
| `Credentials` | The host's credential store: `fabricui.CredentialStore` (list names, set write-only, remove) plus `WiFi(id)` and `Thread(id)` for a commissioning that names one. `fabricui.NewMemoryCredentials(gofabric.ThreadDatasetName)` is an in-memory one that checks a dataset with go-fabric's decoder. Nil answers 503 on the credential routes. |
| `Pair` | The template of every commissioning: the attestation policy, the default subscription, the BLE central (`Capabilities().BLE`). The network a request names is set on a copy. |
| `Logger` | The adapter's logger; nothing it logs carries a passcode or a credential. |

What each route calls on the facade:

| Route | go-fabric |
| --- | --- |
| `GET /controller/status` | `Controller.Fabric` (fabric id, compressed id, index, root node id, vendor, `CurrentLabel`), `Registry().Commissioned`, `SoftwareUpdates` |
| `GET /controller/nodes`, `/nodes/{id}` | `Registry().Commissioned`, `GetByAddress(Fabric().AddressOf(id))`; `Node.Record`, `Properties`, `ConnectionState`, `SubscriptionActive`, `Disabled`, `Endpoints`, BasicInformation from the image |
| `DELETE /controller/nodes/{id}` | `Registry().Decommission` |
| `POST …/interview`, `/enable`, `/disable`, `/auto-subscribe` | `Node.Read` (wildcard, no version filters), `Enable`, `Disable`, `SetAutoSubscribe` |
| `POST …/window`, `/window/close` | `Node.OpenCommissioningWindow` (enhanced), `RevokeCommissioningWindow` |
| `POST`/`PUT …/attributes`, `POST …/commands` | `Node.Read` / `Write` / `Invoke`, values through `client.EncodeJSON` and `DecodeJSON` by the generated cluster definitions |
| `GET`/`PUT …/acl`, `…/bindings` | `Node.Read` and one `Node.Write` of the whole list, built from `cluster/spec/accesscontrol` and `binding` with our fabric index |
| `…/icd`, `/icd/register`, `/icd/unregister` | `Node.ICD()` (`Register`, `Unregister`, `Forget`) |
| `GET …/fabrics`, `DELETE …/fabrics/{index}` | `Node.Read` of OperationalCredentials `Fabrics` and `CurrentFabricIndex`, `Node.Invoke` RemoveFabric |
| `POST /controller/commission` | `Controller.Pair`, or `Registry().ForDevice` and `Commission` for a discovered device |
| `GET /controller/discovery` | `Controller.Discover` |
| `PUT /controller/fabric/label` | `Fabric().SetLabel`, then `Registry().SetFabricLabel` in the background |
| `/controller/updates…` | `SoftwareUpdates()`: `QueuedUpdates`, `QueryUpdates`, `AddUpdateConsent` |

A host without go-fabric's controller — a test, a demonstration —
implements `fabricui.ControllerSource` itself; `fabricui.NewMemoryController()`
is an in-memory one with two nodes, one ICD and a pending update.

## Mounting the API alone

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
  theme variables onto the page and follows its light and dark mode; the
  shell's pre-paint script does the same before the first frame, so the
  page never shows another palette first. Under any other origin the
  bridge stays inactive and the tokens' Home Assistant defaults apply,
  light or dark by `prefers-color-scheme`.
- **Admin gate by the host.** The proxy authenticates the user; the host's
  `Options.Admin` admits a request when it came through the proxy (for
  example by the proxy's source address). Token mode is for a daemon that
  serves the page directly.
