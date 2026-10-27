// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package gofabric binds the controller page family to go-fabric's
// controller: [ControllerOf] is a [fabricui.ControllerSource] over a
// *controller.Controller, the one place in this module that knows the
// controller facade (controller, controller/client,
// controller/softwareupdate, controller/thread and the generated cluster
// definitions it encodes ACL and Binding entries with). A host passes it
// in fabricui.Options.Controller:
//
//	hooks := gofabric.NewHooks()
//	cfg.HooksFor = hooks.Wrap(cfg.HooksFor) // before controller.New
//	ctrl, err := controller.New(ctx, cfg)
//	src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{Hooks: hooks, Credentials: store})
//	mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Controller: src}))
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
	"time"

	"github.com/SukramJ/go-fabric/client/discovery"
	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/controller/commissioner"
	"github.com/SukramJ/go-fabric/controller/thread"

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
	// events (commissioning progress, update).
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
}

// ThreadDatasetName is a Thread operational dataset's network name, read
// with go-fabric's controller/thread decoder (matter.js
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

// ControllerOf is the controller family's source over ctrl.
func ControllerOf(ctrl *controller.Controller, opts ControllerOptions) fabricui.ControllerSource {
	s := &source{ctrl: ctrl, opts: opts, events: fabricui.NewEvents(), discovered: map[string]client.Device{}, log: opts.Logger}
	if s.log == nil {
		s.log = slog.Default()
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

	// mu guards discovered: the devices the last discoveries answered,
	// by instance name, for CommissionDiscovered.
	mu         sync.Mutex
	discovered map[string]client.Device
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
	return fabricui.ControllerCapabilities{BLE: s.opts.Pair.BLE != nil}
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
	for _, n := range s.ctrl.Registry().Commissioned() {
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
	f := s.ctrl.Fabric()
	if f == nil {
		return nil, false
	}
	n := s.ctrl.Registry().GetByAddress(f.AddressOf(id))
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

// Pair pairs the device code names — controller.Controller.Pair, over BLE
// when the template has a central, the request does not rule it out, the
// code allows it and a credential puts the device on its network, as
// go-fabric-server's commission_with_code decides
// (controllerapi/cmd_commission.go:98-141 at v0.3.2).
func (s *source) Pair(ctx context.Context, code string, o fabricui.PairOptions) (fabricui.ControllerNode, error) {
	pc, err := controller.ParsePairingCode(code)
	if err != nil {
		return nil, fmt.Errorf("%w: not a QR payload or a manual pairing code", fabricui.ErrInvalidArgument)
	}
	wifi, th, err := s.networkOf(o)
	if err != nil {
		return nil, err
	}
	opts := s.opts.Pair
	opts.BLE, opts.WiFi, opts.Thread = nil, wifi, th
	if s.opts.Pair.BLE != nil && !o.NetworkOnly && (pc.Manual || pc.Capabilities&capabilityBLE != 0) && (wifi != nil || th != nil) {
		opts.BLE = s.opts.Pair.BLE
	}
	n, err := s.ctrl.Pair(ctx, code, opts)
	if err != nil {
		return nil, commissionError(err)
	}
	w, ok := s.nodeOf(n)
	if !ok {
		return nil, fmt.Errorf("%w: the commissioned node has no peer address", fabricui.ErrNodeFailed)
	}
	return w, nil
}

// commissionError maps a commissioning's error to the port's kinds.
func commissionError(err error) error {
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
	s.mu.Lock()
	for i := range found {
		s.discovered[found[i].InstanceID] = found[i]
	}
	s.mu.Unlock()
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

// CommissionDiscovered commissions a discovered device with its passcode
// through the registry (client.Registry.ForDevice and Commission), as
// go-fabric-server's commission_on_network does
// (controllerapi/cmd_commission.go:159-232).
func (s *source) CommissionDiscovered(ctx context.Context, d fabricui.Discovered, passcode uint32, o fabricui.PairOptions) (fabricui.ControllerNode, error) {
	s.mu.Lock()
	dev, ok := s.discovered[d.InstanceName]
	s.mu.Unlock()
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
	n, err := reg.ForDevice(ctx, dev)
	if err != nil {
		return nil, commissionError(err)
	}
	po := s.opts.Pair
	copts := client.CommissionOptions{
		Options: commissioner.Options{
			Passcode: passcode, OnAttestationFailure: po.OnAttestationFailure,
			WiFiNetwork: wifi, ThreadNetwork: th,
		},
		AutoSubscribe: po.AutoSubscribe,
	}
	if po.Edit != nil {
		po.Edit(&copts.Options)
	}
	if err := reg.Commission(ctx, n, copts); err != nil {
		return nil, commissionError(err)
	}
	s.mu.Lock()
	delete(s.discovered, d.InstanceName)
	s.mu.Unlock()
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
	if err := s.ctrl.Registry().Decommission(ctx, n.n); err != nil {
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
	// The push outlives the request on purpose: the answer does not wait.
	go func() { //nolint:gosec,contextcheck // G118: the push outlives the request by design
		ctx, cancel := context.WithTimeout(context.Background(), labelPushTimeout)
		defer cancel()
		if err := s.ctrl.Registry().SetFabricLabel(ctx, label); err != nil {
			s.log.Warn("fabricui.gofabric.label_push_failed", slog.String("error", err.Error()))
		}
	}()
	return nil
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
