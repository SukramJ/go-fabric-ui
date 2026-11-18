// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package gofabric binds the controller page family to go-fabric's
// controller: [ControllerOf] is a [fabricui.ControllerSource] over a
// *controller.Controller, the one place in this module that knows the
// controller facade. Beside the facade every package may import, the
// import guard (internal/arch, its adapter list) allows it exactly: the
// controller facade controller, controller/client,
// controller/commissioner and controller/softwareupdate, the network
// graph controller/topology, the diagnostic shapes diag, the DNS-SD
// identifiers client/discovery, the Thread dataset codec thread, and the
// generated cluster definitions it encodes ACL and Binding entries with
// (cluster/spec and its accesscontrol, basicinformation, binding,
// icdmanagement and operationalcredentials). A host passes it in
// fabricui.Options.Controller:
//
//	hooks := gofabric.NewHooks()
//	cfg.HooksFor = hooks.Wrap(cfg.HooksFor) // before controller.New
//	var changes topology.Changes
//	changes.Install(&cfg) // before controller.New
//	ctrl, err := controller.New(ctx, cfg)
//	svc := topology.New(ctrl, topology.Config{Changes: &changes, Clock: cfg.Clock})
//	src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{Hooks: hooks, Credentials: store, Topology: svc})
//	mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Controller: src}))
//	// … on shutdown: svc.Close() and hooks.Close() (or src.Close()
//	// without hooks) before ctrl.Close(ctx)
//
// The facade calls are those go-fabric-server's controllerapi makes per
// WebSocket command (cmd_node.go, cmd_commission.go, cmd_interact.go,
// cmd_acl.go, cmd_icd.go at v0.3.2), made once here.
package gofabric

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SukramJ/go-fabric/client/discovery"
	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/controller/commissioner"
	"github.com/SukramJ/go-fabric/controller/topology"
	"github.com/SukramJ/go-fabric/thread"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// Credentials is the host's credential store as a commissioning reads it:
// the routes' [fabricui.CredentialStore] and the secrets a pairing that
// names a credential needs. *fabricui.MemoryCredentials satisfies it.
type Credentials interface {
	fabricui.CredentialStore
	WiFi(id string) (ssid, passphrase string, ok bool)
	Thread(id string) (dataset []byte, ok bool)
}

// ControllerOptions configure [ControllerOf].
type ControllerOptions struct {
	// Hooks deliver the nodes' transitions to the controller stream; the
	// host wires them into controller.Config.HooksFor before
	// controller.New. Nil leaves the stream with the handler's own
	// events (commissioning progress) and the update events the source
	// observes on the controller's update manager; the host then closes
	// the source itself ([Source.Close]), which Hooks.Close does
	// otherwise.
	Hooks *Hooks
	// Credentials is the host's Wi-Fi and Thread credential store; nil
	// serves none (the credential routes answer 503).
	Credentials Credentials
	// Pair is the template of every commissioning: the attestation
	// policy, the default subscription, the BLE central. The network a
	// request names is set on a copy.
	Pair controller.PairOptions
	// Logger receives the adapter's logs; nil is slog.Default(). Nothing
	// it logs carries a passcode or a credential.
	Logger *slog.Logger
	// Topology is the host's network graph of the controller's nodes,
	// which GET /controller/topology serves and whose changes reach the
	// stream as topology_changed. It is the host's: its Changes are
	// installed on controller.Config before controller.New, and the host
	// closes it on shutdown. Nil serves no graph (the route answers 501).
	Topology *topology.Service
}

// ThreadDatasetName is a Thread operational dataset's network name, read
// with go-fabric's thread decoder (matter.js
// OperationalDataset.decode); a dataset that does not decode wraps
// fabricui.ErrInvalidCredential. It is the threadName of
// fabricui.NewMemoryCredentials for a store that checks what it keeps.
func ThreadDatasetName(dataset []byte) (string, error) {
	ds, err := thread.DecodeDataset(dataset)
	if err != nil {
		return "", fmt.Errorf("%w: %w", fabricui.ErrInvalidCredential, err)
	}
	if ds.NetworkName == nil || *ds.NetworkName == "" {
		return "", fmt.Errorf("%w: the dataset names no network", fabricui.ErrInvalidCredential)
	}
	return *ds.NetworkName, nil
}

// Source is the controller family's source over a controller, with what
// ends the work it runs beside the requests.
type Source interface {
	fabricui.ControllerSource
	// Close releases the source's subscriptions to the topology service
	// and the update manager, cancels a running label push and waits for
	// it. [Hooks.Close] calls it; a host without hooks calls it before it
	// closes the controller. It is idempotent.
	Close()
}

// ControllerOf is the controller family's source over ctrl.
func ControllerOf(ctrl *controller.Controller, opts ControllerOptions) Source {
	s := &source{ctrl: ctrl, opts: opts, events: fabricui.NewEvents(), log: opts.Logger, now: time.Now}
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.push = s.pushCurrentLabel
	if s.log == nil {
		s.log = slog.Default()
	}
	if opts.Topology != nil {
		s.watchTopology(opts.Topology)
	}
	if m := ctrl.SoftwareUpdates(); m != nil {
		s.watchUpdates(m)
	}
	if opts.Hooks != nil {
		opts.Hooks.bind(s)
	}
	return s
}

// source is the adapter.
type source struct {
	ctrl   *controller.Controller
	opts   ControllerOptions
	events *fabricui.Events
	log    *slog.Logger

	// mu guards discovered and discoveredAt: the devices the last
	// discovery answered, by instance name, for CommissionDiscovered, and
	// when it answered. now is the clock; a test sets it.
	mu           sync.Mutex
	discovered   map[string]client.Device
	discoveredAt time.Time
	now          func() time.Time

	// topoState holds the subscriptions to opts.Topology and to the
	// controller's update manager, released by Close.
	topoState

	// ctx is the source's life: Close cancels it, and the label push
	// derives from it.
	ctx    context.Context //nolint:containedctx // the source's own lifetime, not a request's
	cancel context.CancelFunc
	// labelMu guards the label push's state: closed after Close, pushing
	// while the push worker runs, again when a label was set during a
	// push. pushes counts the running worker; Close waits for it.
	labelMu        sync.Mutex
	closed         bool
	pushing, again bool
	pushes         sync.WaitGroup
	// push sends the fabric's current label to the nodes; a test
	// replaces it.
	push func(ctx context.Context) error
}

func (s *source) publish(e fabricui.ControllerEvent) {
	s.events.Publish(fabricui.Event{Type: e.Type, Data: e})
}

func (s *source) Fabric() fabricui.ControllerFabric {
	f := s.ctrl.Fabric()
	if f == nil {
		return fabricui.ControllerFabric{}
	}
	out := fabricui.ControllerFabric{
		FabricID: f.FabricID, FabricIndex: f.FabricIndex, NodeID: f.RootNodeID,
		VendorID: f.RootVendorID, Label: f.CurrentLabel(),
	}
	if f.Identity != nil {
		id := f.Identity.CompressedFabricID
		out.CompressedFabricID = beUint64(id[:])
	}
	return out
}

func beUint64(b []byte) uint64 {
	var v uint64
	for _, x := range b {
		v = v<<8 | uint64(x)
	}
	return v
}

func (s *source) Capabilities() fabricui.ControllerCapabilities {
	return fabricui.ControllerCapabilities{BLE: s.opts.Pair.BLE != nil, Diagnostics: true}
}

// nodeOf wraps a commissioned node; false while it has no peer address.
func (s *source) nodeOf(n *client.Node) (*node, bool) {
	if n == nil {
		return nil, false
	}
	addr, ok := n.PeerAddress()
	if !ok {
		return nil, false
	}
	return &node{s: s, n: n, id: addr.NodeID}, true
}

func (s *source) Nodes() []fabricui.ControllerNode {
	var nodes []*node
	reg := s.ctrl.Registry()
	if reg == nil {
		return []fabricui.ControllerNode{}
	}
	for _, n := range reg.Commissioned() {
		if w, ok := s.nodeOf(n); ok {
			nodes = append(nodes, w)
		}
	}
	slices.SortFunc(nodes, func(a, b *node) int {
		switch {
		case a.id < b.id:
			return -1
		case a.id > b.id:
			return 1
		}
		return 0
	})
	out := make([]fabricui.ControllerNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n)
	}
	return out
}

// lookup is the commissioned node with the id — go-fabric-server's
// Server.node (controllerapi/cmd_node.go:24-35 at v0.3.2).
func (s *source) lookup(id uint64) (*node, bool) {
	f, reg := s.ctrl.Fabric(), s.ctrl.Registry()
	if f == nil || reg == nil {
		return nil, false
	}
	n := reg.GetByAddress(f.AddressOf(id))
	if n == nil || !n.IsCommissioned() {
		return nil, false
	}
	return s.nodeOf(n)
}

func (s *source) Node(id uint64) (fabricui.ControllerNode, bool) {
	n, ok := s.lookup(id)
	if !ok {
		return nil, false
	}
	return n, true
}

// networkOf is the network a commissioning provisions, from the stored
// credentials the request names.
func (s *source) networkOf(o fabricui.PairOptions) (*commissioner.WiFiNetwork, *commissioner.ThreadNetwork, error) {
	if o.WiFiID == "" && o.ThreadID == "" {
		return nil, nil, nil
	}
	if s.opts.Credentials == nil {
		return nil, nil, fmt.Errorf("%w: the host stores no credentials", fabricui.ErrInvalidArgument)
	}
	var wifi *commissioner.WiFiNetwork
	var th *commissioner.ThreadNetwork
	if o.WiFiID != "" {
		ssid, pass, ok := s.opts.Credentials.WiFi(o.WiFiID)
		if !ok {
			return nil, nil, fmt.Errorf("%w: wifi %q", fabricui.ErrUnknownCredential, o.WiFiID)
		}
		wifi = &commissioner.WiFiNetwork{SSID: ssid, Credentials: pass}
	}
	if o.ThreadID != "" {
		ds, ok := s.opts.Credentials.Thread(o.ThreadID)
		if !ok {
			return nil, nil, fmt.Errorf("%w: thread %q", fabricui.ErrUnknownCredential, o.ThreadID)
		}
		th = &commissioner.ThreadNetwork{OperationalDataset: ds}
	}
	return wifi, th, nil
}

// capabilityBLE is the BLE bit of a QR payload's discovery capabilities
// (controller.PairingCode.Capabilities).
const capabilityBLE = 1 << 1

// bleFor is a commissioning's BLE decision from the host-side inputs: a
// central, the code's discovery capabilities (a manual code names none and
// allows every transport; a QR code allows BLE with bit 1, go-fabric
// client/discovery/session.go:73-93 at v0.18.0), network_only and a named
// credential — go-fabric refuses BLE without a network
// (controller/controller.go:674-676). The switch is the reasons'
// precedence: the operator's network_only before the device's code before
// the missing network (brief U8 §1).
func bleFor(central bool, pc controller.PairingCode, networkOnly, network bool) fabricui.BLEDecision {
	switch {
	case !central:
		return fabricui.BLEDecision{Reason: fabricui.BLESkipNoCentral}
	case networkOnly:
		return fabricui.BLEDecision{Reason: fabricui.BLESkipNetworkOnly}
	case !pc.Manual && pc.Capabilities&capabilityBLE == 0:
		return fabricui.BLEDecision{Reason: fabricui.BLESkipCode}
	case !network:
		return fabricui.BLEDecision{Reason: fabricui.BLESkipNoNetwork}
	default:
		return fabricui.BLEDecision{Used: true}
	}
}

// Pair pairs the device code names — controller.Controller.Pair, over BLE
// when bleFor says so: the template has a central, the request does not
// rule it out, the code allows it and a named credential puts the device
// on its network, the four inputs go-fabric-server's commission_with_code
// decides on (controllerapi/cmd_commission.go:132-135 at main b947c93;
// brief U8 §1). The decision is stated, never silent: through
// PairOptions.Transport always, and as the log line
// fabricui.pair.ble_skipped with its reason when a central is not used.
func (s *source) Pair(ctx context.Context, code string, o fabricui.PairOptions) (fabricui.ControllerNode, error) {
	pc, err := controller.ParsePairingCode(code)
	if err != nil {
		return nil, fmt.Errorf("%w: not a QR payload or a manual pairing code", fabricui.ErrInvalidArgument)
	}
	wifi, th, err := s.networkOf(o)
	if err != nil {
		return nil, err
	}
	d := bleFor(s.opts.Pair.BLE != nil, pc, o.NetworkOnly, wifi != nil || th != nil)
	opts := s.opts.Pair
	opts.BLE, opts.WiFi, opts.Thread = nil, wifi, th
	if d.Used {
		opts.BLE = s.opts.Pair.BLE
	}
	if s.opts.Pair.BLE != nil && !d.Used {
		s.log.Info("fabricui.pair.ble_skipped", slog.String("reason", string(d.Reason)))
	}
	if o.Transport != nil {
		o.Transport(d)
	}
	var started atomic.Bool
	opts.Progress = markSubscribe(progressFor(opts.Progress, o.Progress), &started)
	opts.TermsAndConditions = termsOf(o.TermsAndConditions, opts.TermsAndConditions)
	n, err := s.ctrl.Pair(ctx, code, opts)
	if err != nil && !s.commissionedDespite(n, err, &started) {
		return nil, commissionError(err)
	}
	w, ok := s.nodeOf(n)
	if !ok {
		return nil, fmt.Errorf("%w: the commissioned node has no peer address", fabricui.ErrNodeFailed)
	}
	return w, nil
}

// progressFor is the commissioning's progress hook: the template's, then
// the port's with the step in the port's shape (go-fabric
// controller.PairOptions.Progress, controller/controller.go:611-618 at
// v0.18.0; the phases commissioner.StepStarted…StepSkipped,
// controller/commissioner/progress.go:15-32). The step's error stays here:
// the port's step carries none. Nil when neither is set.
func progressFor(host func(controller.PairingStep), port func(fabricui.PairingStep)) func(controller.PairingStep) {
	if port == nil {
		return host
	}
	return func(st controller.PairingStep) {
		if host != nil {
			host(st)
		}
		port(fabricui.PairingStep{Number: st.Number, Sub: st.Sub, Name: st.Name, Phase: st.Phase.String()})
	}
}

// markSubscribe wraps progress so that started is set when this
// commissioning reports step Subscribe 21.1 started: go-fabric reports it
// once the commissioning completed and the node starts
// (commissioner.ReportSubscribe).
func markSubscribe(progress func(controller.PairingStep), started *atomic.Bool) func(controller.PairingStep) {
	return func(st controller.PairingStep) {
		if st.Number == 21 && st.Sub == 1 && st.Phase == commissioner.StepStarted {
			started.Store(true)
		}
		if progress != nil {
			progress(st)
		}
	}
}

// termsOf is the request's Terms and Conditions acknowledgement as
// go-fabric's (controller.PairOptions.TermsAndConditions, controller/
// controller.go:654-663 at v0.19.1; commissioner.Options.TermsAndConditions,
// controller/commissioner/commissioner.go:154-167), else the host
// template's. The host's Edit runs after it, as go-fabric's Pair runs it
// (controller/controller.go:752-768): what the host edits stays the
// host's.
func termsOf(tc *fabricui.TermsAcknowledgement, template *commissioner.TermsAndConditionsAcknowledgement) *commissioner.TermsAndConditionsAcknowledgement {
	if tc == nil {
		return template
	}
	return &commissioner.TermsAndConditionsAcknowledgement{Version: tc.Version, Acknowledgements: tc.Acknowledgements}
}

// commissionError maps a commissioning's error to the port's kinds. A
// device that requires terms and conditions is the port's
// *ErrTermsRequired with its version and deadline
// (commissioner.TermsRequiredError, controller/commissioner/
// errors.go:158-176 at v0.19.1).
func commissionError(err error) error {
	var terms *commissioner.TermsRequiredError
	if errors.As(err, &terms) {
		return &fabricui.ErrTermsRequired{Version: terms.Version, Deadline: terms.Deadline}
	}
	switch {
	case errors.Is(err, controller.ErrPairingCode), errors.Is(err, controller.ErrConfiguration):
		return fmt.Errorf("%w: %w", fabricui.ErrInvalidArgument, err)
	case errors.Is(err, controller.ErrNoDevice):
		return fabricui.ErrNoDevice
	case errors.Is(err, context.Canceled):
		return err
	}
	return fmt.Errorf("%w: %w", fabricui.ErrNodeFailed, err)
}

func identifierOf(f fabricui.DiscoverFilter) (discovery.Identifier, error) {
	switch f.Kind {
	case "", "any":
		return discovery.AnyCommissionable(), nil
	case "long_discriminator":
		if f.Value > 0x0FFF {
			return discovery.Identifier{}, fmt.Errorf("%w: a long discriminator has 12 bits", fabricui.ErrInvalidArgument)
		}
		return discovery.ByLongDiscriminator(uint16(f.Value)), nil
	case "short_discriminator":
		if f.Value > 0x0F {
			return discovery.Identifier{}, fmt.Errorf("%w: a short discriminator has 4 bits", fabricui.ErrInvalidArgument)
		}
		return discovery.ByShortDiscriminator(uint8(f.Value)), nil
	case "vendor":
		if f.Value > 0xFFFF {
			return discovery.Identifier{}, fmt.Errorf("%w: a vendor id has 16 bits", fabricui.ErrInvalidArgument)
		}
		return discovery.ByVendor(uint16(f.Value)), nil
	case "device_type":
		return discovery.ByDeviceType(f.Value), nil
	}
	return discovery.Identifier{}, fmt.Errorf("%w: filter %q", fabricui.ErrInvalidArgument, f.Kind)
}

// Discover lists the commissionable devices — controller.Controller.Discover,
// as go-fabric-server's discover (controllerapi/cmd_commission.go:335-347).
func (s *source) Discover(ctx context.Context, f fabricui.DiscoverFilter, timeout time.Duration) ([]fabricui.Discovered, error) {
	id, err := identifierOf(f)
	if err != nil {
		return nil, err
	}
	found, err := s.ctrl.Discover(ctx, id, timeout)
	if err != nil {
		return nil, err
	}
	s.remember(found)
	out := make([]fabricui.Discovered, 0, len(found))
	for i := range found {
		out = append(out, discoveredOf(&found[i]))
	}
	return out, nil
}

func discoveredOf(d *client.Device) fabricui.Discovered {
	out := fabricui.Discovered{
		InstanceName: d.InstanceID, Hostname: d.Hostname, LongDiscriminator: d.D, CommissioningMode: d.CM,
		Addresses: []string{},
	}
	for _, a := range d.Addresses {
		out.Addresses = append(out.Addresses, a.IP.String())
		if out.Port == 0 {
			out.Port = a.Port
		}
	}
	if d.VP != nil {
		vendor, product, _ := cutPlus(*d.VP)
		if v, ok := discovery.ParseTxtDecimal(vendor, 0xFFFF); ok {
			out.VendorID = uint16(v) //nolint:gosec // ParseTxtDecimal bounds it by 0xFFFF
		}
		if p, ok := discovery.ParseTxtDecimal(product, 0xFFFF); ok {
			out.ProductID = uint16(p) //nolint:gosec // ParseTxtDecimal bounds it by 0xFFFF
		}
	}
	if d.DT != nil {
		out.DeviceType = *d.DT
	}
	if d.DN != nil {
		out.DeviceName = *d.DN
	}
	if d.PH != nil {
		out.PairingHint = *d.PH
	}
	if d.PI != nil {
		out.PairingInstruction = *d.PI
	}
	return out
}

func cutPlus(s string) (a, b string, ok bool) {
	for i := range len(s) {
		if s[i] == '+' {
			return s[:i], s[i+1:], true
		}
	}
	return s, "", false
}

// rediscoverWindow is how long CommissionDiscovered searches again for a
// device the adapter no longer holds.
const rediscoverWindow = 5 * time.Second

// discoveredMaxAge is how long a discovery's answers stand in for the
// device: an older one may have moved, so CommissionDiscovered finds it
// again.
const discoveredMaxAge = 2 * time.Minute

// remember replaces the cached devices with a discovery's answers: only
// the last discovery's are kept, so anything on the network that
// advertises does not grow the cache.
func (s *source) remember(found []client.Device) {
	m := make(map[string]client.Device, len(found))
	for i := range found {
		m[found[i].InstanceID] = found[i]
	}
	s.mu.Lock()
	s.discovered, s.discoveredAt = m, s.now()
	s.mu.Unlock()
}

// cached is the device the last discovery answered by instance name,
// false when it did not or the answer is older than discoveredMaxAge.
func (s *source) cached(instance string) (client.Device, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.discovered[instance]
	if !ok || s.now().Sub(s.discoveredAt) > discoveredMaxAge {
		return client.Device{}, false
	}
	return d, true
}

// forget drops a commissioned device from the cache.
func (s *source) forget(instance string) {
	s.mu.Lock()
	delete(s.discovered, instance)
	s.mu.Unlock()
}

// CommissionDiscovered commissions a discovered device with its passcode
// through the registry (client.Registry.ForDevice and Commission), as
// go-fabric-server's commission_on_network does
// (controllerapi/cmd_commission.go:159-232). The BLE decision does not
// apply — the device came from a DNS-SD discovery — so it reports no
// PairOptions.Transport.
func (s *source) CommissionDiscovered(ctx context.Context, d fabricui.Discovered, passcode uint32, o fabricui.PairOptions) (fabricui.ControllerNode, error) {
	dev, ok := s.cached(d.InstanceName)
	if !ok {
		if d.LongDiscriminator == 0 {
			return nil, fabricui.ErrNoDevice
		}
		found, err := s.ctrl.Discover(ctx, discovery.ByLongDiscriminator(d.LongDiscriminator), rediscoverWindow)
		if err != nil {
			return nil, commissionError(err)
		}
		i := slices.IndexFunc(found, func(f client.Device) bool { return d.InstanceName == "" || f.InstanceID == d.InstanceName })
		if i < 0 {
			return nil, fabricui.ErrNoDevice
		}
		dev = found[i]
	}
	wifi, th, err := s.networkOf(o)
	if err != nil {
		return nil, err
	}
	reg := s.ctrl.Registry()
	if reg == nil {
		return nil, fmt.Errorf("%w: the controller is closed", fabricui.ErrNodeFailed)
	}
	n, err := reg.ForDevice(ctx, dev, "")
	if err != nil {
		return nil, commissionError(err)
	}
	po := s.opts.Pair
	copts := client.CommissionOptions{
		Options: commissioner.Options{
			Passcode: passcode, OnAttestationFailure: po.OnAttestationFailure,
			WiFiNetwork: wifi, ThreadNetwork: th,
			TermsAndConditions: termsOf(o.TermsAndConditions, po.TermsAndConditions),
			Progress:           progressFor(po.Progress, o.Progress),
		},
		AutoSubscribe: po.AutoSubscribe,
	}
	if po.Edit != nil {
		po.Edit(&copts.Options)
	}
	var started atomic.Bool
	copts.Progress = markSubscribe(copts.Progress, &started)
	if err := reg.Commission(ctx, n, copts); err != nil && !s.commissionedDespite(n, err, &started) {
		return nil, commissionError(err)
	}
	s.forget(d.InstanceName)
	w, ok := s.nodeOf(n)
	if !ok {
		return nil, fmt.Errorf("%w: the commissioned node has no peer address", fabricui.ErrNodeFailed)
	}
	return w, nil
}

// Decommission removes our fabric from the node — client.Registry.Decommission
// (go-fabric-server's remove_node, controllerapi/cmd_node.go:81-94).
func (s *source) Decommission(ctx context.Context, id uint64) error {
	n, ok := s.lookup(id)
	if !ok {
		return fabricui.ErrNodeNotFound
	}
	reg := s.ctrl.Registry()
	if reg == nil {
		return fabricui.ErrNodeNotFound
	}
	if err := reg.Decommission(ctx, n.n); err != nil {
		return nodeError(err)
	}
	return nil
}

// labelPushTimeout bounds the background push of a new fabric label to
// the connected nodes.
const labelPushTimeout = 2 * time.Minute

// SetFabricLabel sets the fabric's label at once and pushes it to the
// connected nodes in the background (client.Registry.SetFabricLabel), as
// go-fabric-server's set_default_fabric_label answers before the push
// (controllerapi/cmd_server.go:217-286).
func (s *source) SetFabricLabel(_ context.Context, label string) error {
	f := s.ctrl.Fabric()
	if f == nil {
		return fmt.Errorf("%w: the controller has no fabric", fabricui.ErrNodeFailed)
	}
	if err := f.SetLabel(label); err != nil {
		if errors.Is(err, commissioner.ErrLabel) {
			return fmt.Errorf("%w: %w", fabricui.ErrInvalidArgument, err)
		}
		return err
	}
	s.pushLabel() //nolint:contextcheck // the push outlives the request by design
	return nil
}

// pushLabel starts the label push worker, or asks the running one for
// another round. The push outlives the request on purpose: the answer
// does not wait. One worker at a time sends the label as it stands when
// a round starts, so the pushes end at the last label set, never at an
// older one a late goroutine still held.
func (s *source) pushLabel() {
	s.labelMu.Lock()
	defer s.labelMu.Unlock()
	switch {
	case s.closed:
		return
	case s.pushing:
		s.again = true
		return
	}
	s.pushing = true
	s.pushes.Add(1)
	go s.pushRounds() //nolint:contextcheck // the push outlives the request by design
}

// pushRounds is the worker: a round per label set while it runs, each
// bounded by labelPushTimeout and cancelled by Close.
func (s *source) pushRounds() {
	defer s.pushes.Done()
	for {
		s.labelMu.Lock()
		s.again = false
		s.labelMu.Unlock()
		ctx, cancel := context.WithTimeout(s.ctx, labelPushTimeout)
		if err := s.push(ctx); err != nil {
			s.log.Warn("fabricui.gofabric.label_push_failed", slog.String("error", err.Error()))
		}
		cancel()
		s.labelMu.Lock()
		if !s.again || s.closed {
			s.pushing = false
			s.labelMu.Unlock()
			return
		}
		s.labelMu.Unlock()
	}
}

// pushCurrentLabel sends the fabric's current label to the connected
// nodes (client.Registry.SetFabricLabel, which sets that label on the
// fabrics again: the current one, so nothing older returns).
func (s *source) pushCurrentLabel(ctx context.Context) error {
	f, reg := s.ctrl.Fabric(), s.ctrl.Registry()
	if f == nil || reg == nil {
		return nil
	}
	return reg.SetFabricLabel(ctx, f.CurrentLabel())
}

// Close implements [Source].
func (s *source) Close() {
	s.labelMu.Lock()
	s.closed = true
	s.labelMu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.pushes.Wait()
	s.closeTopology()
}

func (s *source) Credentials() fabricui.CredentialStore {
	if s.opts.Credentials == nil {
		return nil
	}
	return s.opts.Credentials
}

func (s *source) SoftwareUpdates() fabricui.SoftwareUpdateSource {
	m := s.ctrl.SoftwareUpdates()
	if m == nil {
		return nil
	}
	return &updates{s: s, m: m}
}

func (s *source) Observe(fn func(fabricui.ControllerEvent)) (unsubscribe func()) {
	return s.events.Subscribe(func(e fabricui.Event) {
		if ce, ok := e.Data.(fabricui.ControllerEvent); ok {
			fn(ce)
		}
	})
}

// nodeError maps an interaction's error to the port's kinds: a value
// that does not fit its type is an invalid argument, an ICD refusal over
// other administrators a conflict, anything else the node's failure.
func nodeError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, client.ErrJSONValue):
		return fmt.Errorf("%w: %w", fabricui.ErrInvalidArgument, err)
	case errors.Is(err, context.Canceled):
		return err
	}
	if _, ok := errors.AsType[*client.ICDMultiAdminError](err); ok {
		return fmt.Errorf("%w: %w", fabricui.ErrConflict, err)
	}
	return fmt.Errorf("%w: %w", fabricui.ErrNodeFailed, err)
}

// commissionedDespite reports a commissioning that happened although it
// returned err: go-fabric returns the node with the error of its start
// after CommissioningComplete (the Subscribe step, already reported as
// failed through the progress hook), and the node stays commissioned
// (controller/client/fabricops.go:229-235 at v0.22.2). The commissioning
// is answered as done and the failed start is logged (UR-P2-adpt-03).
// started is this call's mark of step Subscribe 21.1 (markSubscribe):
// a node commissioned before the call is returned with its refusal too
// (ErrAlreadyCommissioned), and that call commissioned nothing
// (UR-P2-adpt-190).
func (s *source) commissionedDespite(n *client.Node, err error, started *atomic.Bool) bool {
	if !started.Load() || n == nil || !n.IsCommissioned() {
		return false
	}
	s.log.Warn("fabricui.pair.start_failed", slog.String("node", n.ID()), slog.String("error", err.Error()))
	return true
}
