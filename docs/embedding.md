# Embedding

go-fabric-ui ships twice in one version: a Go module with the handlers and
the built bundle, and an npm package with the pages for a host that has a
single-page app of its own.

## A daemon without a web app: the Go module

```go
import "github.com/SukramJ/go-fabric-ui/fabricui"

n, err := node.New(cfg)
// … n.Start(ctx) — before or after the handler is built
h, err := fabricui.NewHandler(fabricui.Options{Node: n, Exposure: provider})
if err != nil {
	return err
}
mux.Handle("/", h)
```

`NewHandler` (0.10.0) serves the API under `/api/v1` and the standalone
shell at `/` from the bundle embedded in the module (`fabricui/dist/`, built and
committed here): no Node.js toolchain in the daemon's build, no second
image. The root (`/`, `/index.html`) answers the shell's index, a file of
the bundle answers itself, and any other path answers 404 — the shell
routes by hash (`/#/node/status`), and its relative asset paths would not
load under a deeper path; an unknown path under the API prefix answers a
404 problem, a served API path asked with another method a 405 problem.
`NewHandler` fails on invalid options; `Handler(o)` is the same for
options known to be valid and panics on the error. Both serve the API
under `/api/v1` only and refuse any other `Prefix`: the shell calls
`./api/v1` relative to the page. A host that mounts the API under
another prefix uses [`Mount`](#mounting-the-api-alone), which keeps any
valid prefix and serves no page.

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
| `Admin` | Whether a request that carries `X-Fabric-UI: 1` has admin rights: every route the contract marks `x-admin` (the writes, the pairing secret and, from 0.10.0, the reading routes and the controller and log streams) and the full answer of the open ones; nil selects token mode. See [Admin rights](#admin-rights) and [contract.md](contract.md#admin-rights). |
| `Reassemble` | Rebuilds the endpoints from the current selection (maintenance route, and after an exposure change). |
| `Erase` | Erases the default node and brings it back (0.7.0): `POST /node/maintenance/erase`. Nil answers 501 and hides the page's button. See [Erasing a node](#erasing-a-node). A node of `Nodes` or of the registry carries its own in `NodeEntry.Erase`. |
| `Events` | Events the host publishes beside the handler's own and the node's (an exposure change made outside the page); they reach admin streams only, whatever their type. The handler keeps a hub of its own (0.10.0): a host that reads the `*fabricui.Events` it passes here no longer sees the handler's events in it. |
| `Prefix` | The API prefix for `Mount`; default `/api/v1`. `NewHandler` and `Handler` take `""` or `/api/v1` only. |
| `Logger` | The handlers' logger; nothing it receives carries a passcode. |
| `ConfiguredWindow` | Open the window on the node's configured passcode instead of fresh credentials. Without it, a node without a fabric has its own window on the configured passcode open again once the window on fresh credentials closes or expires (go-fabric's design; `Commissioning.WindowOnDemand` turns it off); with it, the closed window stays closed. |
| `AllowEmptyTopology` | Allow a window while no endpoint is exposed. |
| `FrameAncestors` | Further origins that may frame the page, each a CSP host-source. See [Response headers and framing](#response-headers-and-framing). |
| `Warnings` | The host's findings about its own configuration, shown on the status and selection pages. |
| `AllowedHosts` | The DNS names besides loopback names and IP addresses under which the handler answers (0.10.0); host names only, no scheme, port, path, wildcard or white space. Any other `Host` is 403 `host_not_allowed`. The check sees the `Host` the proxy sends: a proxy that replaces it makes the check constant (stated from 0.11.0). See [Behind an Ingress](#behind-an-ingress). |
| `AllowHost` | Admits a request whose `Host` the names do not (0.10.0), asked once per request; nil admits by name only. It answers from what only the host's proxy produces, never from `Host` or anything else the browser chooses. Every name it admits passes, reads and writes: the Host check is then the proxy's (its authentication, its server names); see [Behind an Ingress](#behind-an-ingress). |
| `MaxStreams` | The event streams the handler serves at once, every kind and node together (0.10.0); 0 is `DefaultMaxStreams` (64), negative an error. One more answers 503 `too_many_streams` with `Retry-After: 5`. |
| `Done` | Closing it ends every open stream and removes its observers (0.10.0); a host closes it from its server's shutdown hook. |
| `Counters` | The host's own counters (0.8.0), shown on the status page and served to a scraper on `GET /node/status`. |

### Admin rights

`Options.Admin` answers one question per request: may this requester
operate the daemon? The handler asks it only for a request that carries
`X-Fabric-UI: 1` (`fabricui.RequestHeader`) and passed the Host and
Origin checks ([contract.md](contract.md#admin-rights)), on the
request's goroutine and on every request of a page that polls. Its yes
opens every route the contract marks `x-admin`: the pairing codes and
passcodes, every write — fabrics, windows, the selection, the log level,
the node registry, erase and factory reset, and on the controller side
commissioning, attributes, commands, ACL, bindings, credentials,
updates — the reading routes of both families and the controller and log
streams. So it answers yes only from what the requester cannot choose,
such as the source address of an authenticating proxy or a session the
host verified itself, never from a header, query or body the browser
sends, and from memory.

A request it answers no — or one without the header — is served:
`GET /node/status` and `/nodes/{node}/status` without the configuration
block's `store` and with every warning reduced to its code, the rest as
for an admin (health, addresses, window state, counts, surface summary,
counters); `GET /nodes`; `GET /operation/info` and
`/operation/log/level`; `GET /controller/status` with `enabled`,
`capabilities` and `admin`, every other member empty or 0; the node
stream with `nodes_changed`, `window_opened` and `window_closed` only;
and the bundle. Everything else answers 403 `forbidden`. In token mode
(`Admin` nil) the token decides the same routes and answers.

A host's own client of the API — a script, its own app — sends
`X-Fabric-UI: 1` on every request in host mode (in token mode the
`X-Fabric-UI-Token` header instead); the npm package's client sends it
on its own. The browser's `EventSource` sends no header, so a stream it
opens is a non-admin's node stream, and the controller and log streams
answer it 403. A host that passes the pages an `EventSourceFactory` of
its own makes it send the same headers over `fetch`.

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

`GET /node/status` is not an admin route and the page polls it, so any
client that passes the Host check calls four host ports per request, on
the request's goroutine: `Warnings`, `Counters`, the node's
`FabricInfos` and the exposure provider's `List` (to count the enabled
items). Nothing bounds the rate; keep them cheap — answers from memory,
no network round trip, no store scan.

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

The streams are long requests: `http.Server.Shutdown` does not cancel a
running request, so an open page holds the host's shutdown for its whole
budget, and the stream's observers outlive it while the host stops its
node. From 0.10.0 the host closes `Options.Done` when the shutdown
begins, and every stream ends and removes its observers:

```go
done := make(chan struct{})
h, err := fabricui.NewHandler(fabricui.Options{Node: n, Done: done})
// …
srv := &http.Server{Handler: mux}
srv.RegisterOnShutdown(func() { close(done) })
```

The earlier pattern holds as well: a handler stream ends when its
request's context does, so a host may serve with a base context it
cancels when the shutdown begins:

```go
base, stopStreams := context.WithCancel(context.Background())
srv := &http.Server{Handler: mux, BaseContext: func(net.Listener) context.Context { return base }}
srv.RegisterOnShutdown(stopStreams)
```

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
node) or `NodeEntry.Erase` (0.7.0), as it wires `Reassemble`. The
handler runs it detached from the request, bounded by two minutes, so a
client that leaves does not stop it half-way, and logs a failure at
error; the same holds for `NodeRegistry.Delete` and the controller
port's `Decommission`.
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
host's. `Set` ends every node stream a handler serves while the switch is
its `Options.Node` or a `NodeEntry.Source` (a stream's observers are
registered on the node it was opened on), so the
browser's `EventSource` reconnects and the new stream observes the new
node; the erase flow polls `GET /node/status`, which reads through the
switch, until the new node answers. As a registry node's `Source` the same
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
(`status`, `pair`, …); `NewHandler` and `Mount` fail and `Handler`
panics on a bad or repeated key. `Options.Node` is the shorthand for one node under the key
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
(`logring.New`, `Handler.Level`/`SetLevel`, `Records`, `Observe`;
go-fabric v0.22.2 `logring/logring.go:116`, `:142-146`, `:239`, `:331`):
a host maps its `Record` onto `fabricui.LogRecord` and adds `Info`.
`fabricui.MemoryOperations` is the test double the handler tests and the browser smoke use. `fabricui.GoFabricVersion()`
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
| `OnboardingCodes` | Serves the onboarding inventory (0.13.0): `GET /controller/inventory` hands out the onboarding codes go-fabric keeps or a user attached, `PUT`/`DELETE /controller/nodes/{id}/onboarding-code` attach and remove one. Off by default — a host that upgrades does not start handing out passcodes. See "The onboarding inventory" below for what a host must do before it sets it. |
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
| `POST /controller/commission` | `Controller.Pair`, or `Controller.PairDiscovered` for a discovered device (0.12.0, go-fabric v0.24.0; before, `Registry().ForDevice` and `Commission`); `PairOptions.Progress` for the steps, `PairOptions.TermsAndConditions` (0.8.0). Both run go-fabric's one pairing pipeline: the template's attestation policy, an unset one rejecting a failed attestation, the pairing entries in the trace `GET /controller/events` serves and the `pair.failures` metric; a discovered device's commissioning reports Discovery 2.1 as skipped. The template's `BLE` and `Address` are not handed to `PairDiscovered`, which reaches the device at its discovered addresses |
| `GET /controller/discovery` | `Controller.Discover`, always with a timeout (1–60 s, default 3 s): the facade's 0 would search for 15 minutes (go-fabric `controller/code.go:92-99` at v0.22.2) |
| `PUT /controller/fabric/label` | `Fabric().SetLabel`, then `Registry().SetFabricLabel` in the background |
| `/controller/updates…` | `SoftwareUpdates()`: `QueuedUpdates`, `QueryUpdates` over every node, its failed products' nodes (`softwareupdate.ProductCheckErrors`, go-fabric v0.24.0) the check's `errors` (0.12.0; 0.11.0 checked one node per product again), `AddUpdateConsent`; the stream's `update` events from `Observe`, subscribed in `ControllerOf` and released by `Hooks.Close` (0.7.0) |
| `GET /controller/topology` | `ControllerOptions.Topology`, the host's `*topology.Service`: `Topology()`, or `Refresh` with `refresh=true`; its `OnChanged` becomes `topology_changed`. The host wires it: `topology.Changes.Install(&cfg)` before `controller.New`, `topology.New(ctrl, topology.Config{Changes: &changes, Clock: cfg.Clock})`, `ControllerOf(…, ControllerOptions{Topology: svc})`, and `svc.Close()` on shutdown (the subscription goes with `Hooks.Close`). The border routers (0.8.0) need nothing more: with `topology.Config.BorderRouters` nil the service builds on the controller's own registry, `Controller.BorderRouters` (go-fabric `controller/topology/service.go:57-60`, `:163-165` at v0.19.1). Without it `NetworkTopology` answers `fabricui.ErrNotImplemented` (501) |
| `GET /controller/diagnostics`, `/mdns`, `/events`; `controller_diagnostic_event` | `Controller.SessionDiagnostics`, `DiagnoseMDNS`, `DiagnosticEvents` and `OnDiagnosticEvent` (go-fabric `controller/diagnostics.go:46`, `:547`, `:140`, `:154` at v0.21.0), mapped one to one onto the port's types; nothing to wire. `DiagnoseMDNS` waits for the DNS-SD cache and is called from the request only; `OnDiagnosticEvent` runs on the goroutine that recorded the event, and the handler only queues it (0.9.0) |
| `GET /controller/inventory`, `PUT`/`DELETE /controller/nodes/{id}/onboarding-code` | `Controller.Inventory`, sorted by node id as the node list is; `client.Node.SetOnboardingCode`, then the node's entry from a fresh inventory; `client.Node.ClearOnboardingCode` (go-fabric v0.23.0). `*client.OnboardingMismatchError` becomes `*fabricui.OnboardingMismatch` (its field bit named), `client.ErrOnboardingCode` `*fabricui.OnboardingInvalid` with the wrapped decoder's text, `ErrNotCommissioned` and `ErrClosed` `ErrConflict`; a store error stays as it is (500). Only with `ControllerOptions.OnboardingCodes` (0.13.0) |

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

The port change of 0.9.0: `ControllerSource` gains four methods for the
controller's diagnostics, with types of `fabricui`, not go-fabric's.
A source states that it serves them with
`ControllerCapabilities.Diagnostics` in `Capabilities()` — the flag
`capabilities.diagnostics` reports, read without a snapshot, since
`GET /controller/status` is polled; the `gofabric` adapter sets it. With
the flag false, `GET /controller/diagnostics`, `/controller/mdns` and
`/controller/events` answer 501 without calling the four methods, and the
stream does not subscribe. `Diagnostics() (ControllerDiagnostics, bool)`
is the snapshot of the sessions (`ControllerSession`) and peers
(`ControllerPeer`); `false` there answers 501 as well. `DiagnoseMDNS() []MDNSFinding` is called from a
request handler only, errors first. `DiagnosticEvents() []DiagnosticEvent`
is the trace oldest first; the route serves it newest first.
`OnDiagnosticEvent(fn)` registers fn for every new entry and returns its
unsubscribe function, or nil when the source serves no diagnostics; fn
must not block. A source of its own that has no diagnostics leaves the
flag false and may return `false`, nil, nil and nil;
`MemoryController.NoDiagnostics` stands for one.

### The onboarding inventory (0.13.0)

The port change of 0.13.0: `ControllerSource` gains three methods, with
types of `fabricui`:

- `OnboardingInventory() []OnboardingEntry` — every commissioned node by
  node id; `OnboardingEntry` carries the codes as `SecretText`, redacted
  under every fmt verb, slog, encoding/json and
  `encoding.TextMarshaler`; `Reveal()` is the one way to the text, and the
  handler calls it only where it fills the response.
- `SetOnboardingCode(ctx, id, code string) (OnboardingEntry, error)` —
  refusals are `*fabricui.OnboardingInvalid` (`Reason`, never the code;
  400 `onboarding_code_invalid`), `*fabricui.OnboardingMismatch` (`Field`
  `vendor`, `product` or `discriminator`, `Code` and `Node`; 409
  `onboarding_code_mismatch`), `ErrNodeNotFound` (404) and `ErrConflict`
  (409, the node is no longer commissioned). Any other error is 500.
- `ClearOnboardingCode(ctx, id) error` — nil also without a code;
  `ErrNodeNotFound` is 404, any other error 500.

A source never puts the code into an error — not into a refusal's
`Reason`, not into a wrapped store error: the handler logs the text of
every 500, and a refusal's `Reason` is the problem's `detail`.

A source states that it serves them with
`ControllerCapabilities.OnboardingCodes`; a source of its own that keeps
no codes leaves it false and returns nil and `ErrNotImplemented`.
`MemoryController.NoOnboardingCodes` stands for one; the memory
controller keeps the code a commissioning by code used, and its
`OnboardingMismatch` makes the next attach a mismatch — set before the
handler serves or between requests, since the field is not guarded.

What a host must do before it serves the inventory:

1. Set `gofabric.ControllerOptions.OnboardingCodes = true`, and only
   behind an access control fit for an export of the fabric's keys: an
   API that serves the inventory hands out setup passcodes. In host mode
   `Options.Admin` decides who reads them. **Token mode is no access
   control** — its token is served by `GET /controller/status` to
   whoever reaches the handler — so the handler serves the inventory only
   with `Options.Admin` and answers 501 in token mode, whatever the source
   says.
2. Set `controller.Config.KeepOnboardingCodes` if the codes of newly
   paired devices are to be kept; without it the inventory holds only
   attached codes. A device commissioned from the discovery list by its
   passcode has no code to keep; its code is attached afterwards.
   Configure a `Sealer` if the store's secrets are to be sealed at rest.
3. Not log request or response bodies of the three routes in its own
   middleware or proxy, and not cache them; the handler sends
   `Cache-Control: no-store` on every answer, a proxy has to honour it.
4. A host with its own `ControllerSource` implements the three methods
   (or reports `OnboardingCodes: false` and returns `ErrNotImplemented`);
   the interface grows, as it did in 0.9.0.
5. Serve the UI over HTTPS (or localhost) for the live camera scan; typing
   the code and choosing a photograph, which the page decodes in the
   browser and uploads nowhere, work without.
6. A host that mounts the Inventory page from the npm package includes
   `@sukramj/fabric-ui/theme/tokens.css`: the label print's rule
   (`html.fui-print-labels`, set only for the Print button's print) is
   there; without it the button prints the page as it is.

## Mounting the API alone

A daemon with a mux of its own and a different layout mounts only the API:

```go
if err := fabricui.Mount(mux, fabricui.Options{Node: n, Admin: isAdmin, Prefix: "/api/v1"}); err != nil {
    return err
}
```

`Mount` puts the API on the mux as the one pattern `Prefix + "/"`, under
any valid prefix. It
fails — and leaves the mux as it was — on invalid options, when the mux
already serves one of the API's routes under the prefix, and when that
pattern conflicts with one of the mux's. A route the host registers under
the prefix after `Mount` is more specific and shadows the API's.

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
- **Host names.** The handler answers only a `Host` it admits — a
  loopback name, an IP address, a name in `Options.AllowedHosts`, or a
  request `Options.AllowHost` admits — and answers any other 403
  `host_not_allowed`, the page included. The check sees the `Host` the
  proxy sends, not the browser's. Three kinds of proxy:
  - It forwards the browser's `Host` unchanged and the host **lists**
    the names the browser uses in `AllowedHosts`: the handler checks
    the name — the one setup in which the check is the handler's.
  - It forwards the browser's `Host` and the host sets `AllowHost`,
    answered from what only the proxy produces, such as its source
    address — never from `Host` or anything else the browser chooses:
    every name the proxy forwards passes, and the page's own `Origin`
    equals its `Host`, so reads and writes pass under any name. The
    check is then the proxy's — its authentication of the user, its
    server names — as for the Ingress, whose proxy authenticates.
  - It replaces `Host` — with a name of its own, which the host lists in
    `AllowedHosts` (an unlisted one is refused on every request, 403
    `host_not_allowed`), or with the upstream's address: the check is
    constant, every request passes it whatever name the browser used.

  In the second and third case a page under a name that resolves to the
  proxy (DNS rebinding) reads what a page under the proxy's own name
  reads, the token and the admin answers included, and in the second it
  writes too. There the proxy authenticates the user, or serves only its
  own names, with no catch-all server, and refuses any other `Host`.
- **Admin gate by the host.** The proxy authenticates the user; the host's
  `Options.Admin` admits a request when it came through the proxy (for
  example by the proxy's source address), and the page's client sends
  `X-Fabric-UI: 1` with it. Token mode is for a daemon that serves the
  page directly.
- **Framing.** The page's policy says `frame-ancestors 'self'`, and the
  Ingress parent is the same origin, so the Ingress panel needs no
  setting. See the next section for another parent.

## Response headers and framing

What the handler sets on its responses, so a host does not have to:

- **The page** (`index.html`, served for the prefix root alone, and the
  503 page while the bundle is not built) carries `Content-Security-Policy`:
  `default-src 'self'`; `script-src 'self'` and the hash of the inline
  pre-paint script, derived from the embedded `index.html` when the handler
  is built; `style-src 'self'`; `img-src 'self'`; `object-src 'none'`;
  `base-uri 'none'`; `form-action 'self'`;
  `frame-ancestors 'self'` and `Options.FrameAncestors`. The API, the
  streams and the camera scan are the page's own origin and need nothing
  more.
- **Every response** of `NewHandler`, `Handler` and `Mount` — page, bundle, JSON,
  problem, stream — carries `X-Content-Type-Options: nosniff` and
  `Referrer-Policy: no-referrer`; every API answer carries
  `Cache-Control: no-store`, since answers hold pairing codes, the token
  and the daemon's state.

`Options.FrameAncestors` admits further parents that may frame the page —
a dashboard on another origin that shows it in a frame:

```go
h := fabricui.Handler(fabricui.Options{Node: n, FrameAncestors: []string{"https://dash.example:8123"}})
```

Each entry is a CSP host-source (`https://dash.example:8123`,
`https://*.example.com`); `NewHandler` and `Mount` fail and `Handler`
panics on anything else, a quoted keyword such as `'none'` included, and
on an entry that admits any parent (`*`, a scheme-source such as
`https:`, `https://*`, a wildcard over one label such as
`https://*.com`). A wildcard over a public suffix of more than one label
(`https://*.co.uk`) is not recognised as such and admits every site
under it: name the parent's own domain. Under an Ingress the parent is
the same origin and needs no entry. `Mount` serves no page; it checks the field
and the host's own page sets its own policy.
