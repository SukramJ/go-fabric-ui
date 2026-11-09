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
| `Node` | The default node (`*node.Node` satisfies `NodeSource`, or a `*NodeSwitch`, the source a host swaps after an erase), served under `/node/*` and listed first by `GET /nodes` under the key `default`; nil serves the family as disabled unless `Nodes` names one. |
| `Nodes` | Further device nodes (0.5.0), `[]NodeEntry{Key, Label, Kind, Source, Exposure, Reassemble, Events, Erase, Config}` in the selector's order, each under `/nodes/{key}/*`; with `Node` nil the first is the default. See [Several nodes](#several-nodes). |
| `Registry` | The host's node registry (0.6.0): further nodes the operator creates, renames and deletes in the page; needs `Node`, excludes `Nodes`. See [Nodes created in the page](#nodes-created-in-the-page). |
| `NodeConfig` | The default node's configuration as its status page shows it, read-only (0.6.0); nil omits the block. Never a passcode. |
| `Operations` | The daemon's log level, log ring and version (0.5.0) for the Operation page; nil answers 503 there and hides the page. See [Operation](#operation). |
| `ExposureReason` | Why the host serves no selection (0.5.0), in the host's language; the selection page shows it instead of hiding. |
| `Exposure` | The host's `ExposureProvider` ([provider.md](provider.md)); nil hides the selection page. A list-valued (`Multi`) column reads `Item.Lists`, one entry per value; the comma-separated `Item.Fields` form is deprecated. |
| `Controller` | The controller family's source (`gofabric.ControllerOf` over a `*controller.Controller`, see below); nil serves no controller family and every `/controller/*` path answers 404. |
| `Admin` | Who may write, read the pairing secret and use a controller route marked admin; nil selects token mode ([contract.md](contract.md#admin-rights)). |
| `Reassemble` | Rebuilds the endpoints from the current selection (maintenance route, and after an exposure change). |
| `Erase` | Erases the default node and brings it back (0.7.0): `POST /node/maintenance/erase`. Nil answers 501 and hides the page's button. See [Erasing a node](#erasing-a-node). A node of `Nodes` or of the registry carries its own in `NodeEntry.Erase`. |
| `Events` | Events the host publishes beside the handler's own and the node's (an exposure change made outside the page). |
| `Prefix` | The API prefix; default `/api/v1`. |
| `Logger` | The handlers' logger; nothing it receives carries a passcode. |
| `ConfiguredWindow` | Open the window on the node's configured passcode instead of fresh credentials. |
| `AllowEmptyTopology` | Allow a window while no endpoint is exposed. |
| `Warnings` | The host's findings about its own configuration, shown on the status and selection pages. |
| `Counters` | The host's own counters (0.8.0), shown on the status page and served to a scraper on `GET /node/status`. |

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

A host also counts things a scraper wants beside the health — commands it
answered with success and acted on nothing, for one. `Options.Counters`
(0.8.0) returns them as `Counter{Code, Value, Message}`; it is called on
every `GET /node/status` and `GET /nodes/{node}/status`, which answer them
as `counters` (absent without the hook, `[]` when the host has none, also
when the host serves no node; host-wide, so every node's status carries
the same list). `Code` is the counter's stable machine name (`discards`),
`Value` the count since the host started, `Message` the host's one-line
meaning in its own language. The status page shows them in a small table
under the warnings card, code, value and message verbatim, and hides it
when the list is absent or empty; every node event that reloads the
status (a health verdict, a fabric, a window, an endpoint change) reloads
the counters with it, as it does the warnings — a counter alone sends no
event.

The node's transitions need no wiring: each open stream registers an
observer with `Node.Observe` and removes it when the connection closes, so
a handler built after `Node.Start` reports a controller commissioning the
node, a fabric removed over the wire or from a route, every window
opened or closed, and every endpoint whose surface an assembly changed
(`surface_changed`, from `Observers.OnSurfaceChanged`) — each once, whoever
caused it, a reassembly the host runs on its own included — and every CASE
session opened or closed and every change of its subscriptions
(`session_opened`, `session_closed`, `subscriptions_changed`, from
`OnSessionOpened`, `OnSessionClosed`, `OnSubscriptionsChanged`) and every
attribute change of a cluster server (`attribute_changed`, from
`OnAttributeChanged`: the ids, never the values). Since 0.5.0 each stream
also registers `Node.OnHealthSample` and `Node.OnDiagnosticEvent`
(`health_sample`, `diagnostic_event`), so the status and diagnostics pages
update without a reload.
`node.Config.Observers` stays the host's own. A host that carries these
events over a channel of its own registers them with
`fabricui.ObserveNode(n, push)` — `NodeObservers` through `Node.Observe`
plus the two callbacks `node.Observers` does not hold; publishing them
into `Options.Events` instead would hand every stream each one twice.

Everything the handlers read is on the node facade: the window from
`Node.Window` (its kind, discriminator and deadline on the node's clock),
the configured codes from `Node.SetupPayload`, the clusters of every
endpoint, parts included, from `Topology`, fresh window credentials from
`node.NewWindowCredentials`, the session table from
`Node.SessionDiagnostics`, the surface summary of the status (level,
profiles, extensions and counts) from `Node.Surface`, an endpoint's values
from `Node.ReadCluster` and `Node.ReadAttribute` and its events from
`Node.Events` (0.5.0), and the stream's health samples and trace entries
from `Node.OnHealthSample` and `Node.OnDiagnosticEvent` (0.5.0). A host
wrapping its node in its own `NodeSource` implements those methods — the
two callbacks are new in 0.5.0 and break such a wrapper until it forwards
them:

```go
func (w wrapped) OnHealthSample(fn func(node.HealthSample)) func()       { return w.n.OnHealthSample(fn) }
func (w wrapped) OnDiagnosticEvent(fn func(node.DiagnosticEvent)) func() { return w.n.OnDiagnosticEvent(fn) }
```

0.7.0 adds four methods of the facade the same way: `AttestationError`
(the status's `attestation_error`), `BootReason` (`boot_reason`),
`SessionParameters` (the sessions' `parameters`) and `TriggerReannounce`
(`POST /node/mdns/reannounce`). `*node.Node` has them; a wrapper does not
compile until it forwards them. The port follows the facade in one shape
on purpose: an optional interface would let a wrapper that forgets a
method lose a page field without a word. A host that wraps its node only
to follow it across an erase uses `fabricui.NodeSwitch` (0.7.1) and writes
no forwarder: the switch is this repository's, and its build fails when
`NodeSource` gains a method it does not delegate.

```go
func (w wrapped) AttestationError() error                   { return w.n.AttestationError() }
func (w wrapped) BootReason() uint8                         { return w.n.BootReason() }
func (w wrapped) SessionParameters() node.SessionParameters { return w.n.SessionParameters() }
func (w wrapped) TriggerReannounce(ctx context.Context)     { w.n.TriggerReannounce(ctx) }
```

### Erasing a node

`POST /node/maintenance/factory-reset` removes every fabric through the
`NodeSource` and leaves the node running. Erasing is more: go-fabric's
`Node.FactoryReset` removes the fabrics, stops the node and erases its
store and endpoint records, and a Go node is single-use, so the host has
to bring a node back. That is a lifecycle action the handlers cannot take
on a `NodeSource`; the host wires it as `Options.Erase` (the default
node) or `NodeEntry.Erase` (0.7.0), as it wires `Reassemble`.
`fabrictest.Device.FactoryReset` is the pattern: `node.FactoryReset`,
then a new node on the erased store and the same address, started — or
the daemon's exit for a supervisor to restart it. The handler keeps the
`NodeSource` it was given; how a host that builds a new node in place
hands the pages that node is the next section. A registry node may
instead appear in `Registry.List` with a new `Source`, which gets a fresh
handler. Without an `Erase` the route answers 501 and `Status.erase` is
absent, so the page offers no erase.

#### Erase and a node that comes back

The host builds a `fabricui.NewNodeSwitch(first)` over its first node and
hands the switch to `Options.Node` or `NodeEntry.Source` (0.7.1). Its
`Erase` runs `FactoryReset` and starts the new node, then calls
`sw.Set(newNode)`; every method of the switch reaches the current
target, swapped atomically. A failed restart calls nothing: the pages
keep reading the stopped node — its last health verdict and its erased
store's errors — and what the host shows about the failure is the
host's. A stream open across `Set` keeps the node it was opened on (its
observers were registered when it began) until the page reconnects; the
erase flow polls `GET /node/status`, which reads through the switch,
until the new node answers. As a registry node's `Source` the same
`*NodeSwitch` stays the same value across `Set`, so the node keeps its
handler. `NewNodeSwitch(nil)` and `Set(nil)` panic: a nil target would
fail inside the first request, where a nil `Options.Node` answers 503.
The end-to-end test `TestEndToEnd_EraseBringsAFreshNode` is the pattern:

```go
sw := fabricui.NewNodeSwitch(dev.Node())
opts.Node = sw
opts.Erase = func(ctx context.Context) error {
	err := dev.FactoryReset(ctx)
	if err == nil {
		sw.Set(dev.Node())
	}
	return err
}
```

### Several nodes

A daemon that hosts more than one device node lists them all (0.5.0):

```go
fabricui.Handler(fabricui.Options{
	Nodes: []fabricui.NodeEntry{
		{Key: "bridge", Label: "Bridge", Kind: "bridge", Source: bridgeNode},
		{Key: "lights", Label: "Lights", Kind: "device", Source: lightsNode},
	},
	Exposure: provider, // the default node's selection: the first entry's
})
```

`GET /nodes` lists them in that order; `/node/*` serves the first, and
`/nodes/{key}/*` every one. A key is stable across restarts, matches
`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$` and is no page name of the node family
(`status`, `pair`, …); `Handler` panics and `Mount` fails on a bad or
repeated key. `Options.Node` is the shorthand for one node under the key
`default`; with both, `Node` comes first. `Options.Exposure`, `Reassemble` and
`Events` belong to the default node; another node serves the selection,
rebuild and events of its entry (0.6.0), or reports no selection and
`ExposureReason` without them. The shell shows a node selector in its header once
there are two nodes, and routes a node's pages under `#/node/<key>/…`; a
host with its own app builds a node's client with `client.forNode(key)`.

### Operation

`Options.Operations` gives the Operation page its data (0.5.0): the level
the daemon logs at and a way to set it, a bounded ring of its recent log
records with a way to observe new ones, and its version and uptime. The
port is five methods:

```go
type Operations interface {
	LogLevel() slog.Level
	SetLogLevel(slog.Level) error
	Log(since uint64, limit int) []LogRecord
	ObserveLog(fn func(LogRecord)) (unsubscribe func())
	Info() OperationInfo
}
```

The level and the ring belong in front of the host's logger, the one
`node.Config.Logger` receives: a `slog.Handler` with a `slog.LevelVar` that
keeps the newest records. go-fabric's `logring` package is that handler
once it lands (go-fabric issue #393, concept U5 C.4); until then a host
writes the adapter itself, and `fabricui.MemoryOperations` is the test
double the handler tests and the browser smoke use. `fabricui.GoFabricVersion()`
reads the go-fabric version from the binary's build information for
`OperationInfo.GoFabric`. The ring holds no passcode because nothing the
host and go-fabric log carries one; the handler adds nothing to it.

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
| `Hooks` | Turns the nodes' `client.Hooks` into the controller stream's events (`node_added`, `node_removed`, `connection`, `subscription`, `node_updated`, coalesced — a structure change, a report or an event the node delivered). `Wrap` chains them after the host's own `HooksFor`, its `Event` hook included; `Close` (before the controller's) stops the pending timers; events after it, or while the controller has no registry, are dropped. Nil leaves the stream with the handler's own events. |
| `Credentials` | The host's credential store: `fabricui.CredentialStore` (list names, set write-only, remove) plus `WiFi(id)` and `Thread(id)` for a commissioning that names one. `Credential.Default` (0.7.1) is the host's statement of its default entry — the one its own interfaces use without a choice; the pages preselect it. The lookup stays by the id a request names: an empty id names none. `fabricui.NewMemoryCredentials(gofabric.ThreadDatasetName)` is an in-memory one that checks a dataset with go-fabric's decoder. Nil answers 503 on the credential routes. |
| `Pair` | The template of every commissioning: the attestation policy, the default subscription, the BLE central (`Capabilities().BLE`). The network a request names is set on a copy; the request's step reports run after the template's `Progress`, and its Terms and Conditions acknowledgement is set on `PairOptions.TermsAndConditions` before the template's `Edit` runs, as go-fabric runs it — the template's acknowledgement applies when the request sends none (0.8.0; 0.7.x set it through `Edit`, after the template's). A device that requires terms the request does not acknowledge fails with go-fabric's `commissioner.TermsRequiredError`, which the adapter reports as `*fabricui.ErrTermsRequired` (409 `terms_required`). The central is used when the code allows BLE, `network_only` is off and the request names a credential (0.7.1); the decision reaches `PairOptions.Transport` on every pairing from a code. go-fabric's network steps provision by the device's NetworkCommissioning features: a Wi-Fi credential is skipped on a device without the Wi-Fi feature, a Thread dataset on one without Thread (`controller/commissioner/network.go:217-235`, `:308-325` at v0.18.0). |
| `Logger` | The adapter's logger; nothing it logs carries a passcode or a credential. `fabricui.pair.ble_skipped` (Info, `reason` `network_only`, `code_without_ble` or `no_network`, 0.7.1) says why a commissioning did not use the central. |

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
| `POST /controller/commission` | `Controller.Pair`, or `Registry().ForDevice` and `Commission` for a discovered device; `PairOptions.Progress` (commissioner `Options.Progress`) for the steps, `PairOptions.TermsAndConditions` (commissioner `Options.TermsAndConditions` for a discovered device; 0.8.0) |
| `GET /controller/discovery` | `Controller.Discover` |
| `PUT /controller/fabric/label` | `Fabric().SetLabel`, then `Registry().SetFabricLabel` in the background |
| `/controller/updates…` | `SoftwareUpdates()`: `QueuedUpdates`, `QueryUpdates`, `AddUpdateConsent`; the stream's `update` events from `Observe`, subscribed in `ControllerOf` and released by `Hooks.Close` (0.7.0) |
| `GET /controller/topology` | `ControllerOptions.Topology`, the host's `*topology.Service`: `Topology()`, or `Refresh` with `refresh=true`; its `OnChanged` becomes `topology_changed`. The host wires it: `topology.Changes.Install(&cfg)` before `controller.New`, `topology.New(ctrl, topology.Config{Changes: &changes, Clock: cfg.Clock})`, `ControllerOf(…, ControllerOptions{Topology: svc})`, and `svc.Close()` on shutdown (the subscription goes with `Hooks.Close`). The border routers (0.8.0) need nothing more: with `topology.Config.BorderRouters` nil the service builds on the controller's own registry, `Controller.BorderRouters` (go-fabric `controller/topology/service.go:57-60`, `:163-165` at v0.19.1). Without it `NetworkTopology` answers `fabricui.ErrNotImplemented` (501) |

A host without go-fabric's controller — a test, a demonstration —
implements `fabricui.ControllerSource` itself; `fabricui.NewMemoryController()`
is an in-memory one with two nodes, one ICD and a pending update.

Two port changes of 0.7.0 concern such a source. `PairOptions` gains
`TermsAndConditions` (the request's acknowledgement) and `Progress`, which
the handler sets and the source calls with each `PairingStep` as it
starts and ends; a source that never calls it shows no steps.
`ControllerEvent` gains `Step`, `UpdateKind`, `UpdateVersion` and
`UpdateErr`, and the `update` events are the source's: the handler no
longer publishes one after a check or a consent, so a source of its own
emits them through `Observe` (the memory controller emits
`check_finished` and `consent_changed`).

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
