// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"
)

// MemoryCredentials is an in-memory [CredentialStore]: a host without a
// store of its own, the in-memory controller and the tests use it. WiFi
// and Thread hand a stored secret to the commissioning that names it; no
// route reads them.
type MemoryCredentials struct {
	mu         sync.Mutex
	threadName func([]byte) (string, error)
	wifi       map[string]wifiCredential
	thread     map[string]threadCredential
}

type wifiCredential struct{ ssid, passphrase string }

type threadCredential struct {
	name    string
	dataset []byte
}

// NewMemoryCredentials returns an empty store. threadName reads a Thread
// dataset's network name and refuses a dataset no device could join
// (gofabric.ThreadDatasetName, over go-fabric's thread); nil
// names a dataset by its id and accepts any non-empty one.
func NewMemoryCredentials(threadName func(dataset []byte) (string, error)) *MemoryCredentials {
	return &MemoryCredentials{threadName: threadName, wifi: map[string]wifiCredential{}, thread: map[string]threadCredential{}}
}

// List implements [CredentialStore]: Wi-Fi first, each kind by id.
func (m *MemoryCredentials) List(context.Context) ([]Credential, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Credential, 0, len(m.wifi)+len(m.thread))
	for _, id := range slices.Sorted(maps.Keys(m.wifi)) {
		out = append(out, Credential{Kind: CredentialWiFi, ID: id, Name: m.wifi[id].ssid})
	}
	for _, id := range slices.Sorted(maps.Keys(m.thread)) {
		out = append(out, Credential{Kind: CredentialThread, ID: id, Name: m.thread[id].name})
	}
	return out, nil
}

// SetWiFi implements [CredentialStore].
func (m *MemoryCredentials) SetWiFi(_ context.Context, id, ssid, passphrase string) error {
	if id == "" || ssid == "" {
		return fmt.Errorf("%w: a Wi-Fi credential needs an id and an SSID", ErrInvalidCredential)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.wifi[id] = wifiCredential{ssid: ssid, passphrase: passphrase}
	return nil
}

// SetThread implements [CredentialStore].
func (m *MemoryCredentials) SetThread(_ context.Context, id string, dataset []byte) error {
	if id == "" || len(dataset) == 0 {
		return fmt.Errorf("%w: a Thread dataset needs an id and a dataset", ErrInvalidCredential)
	}
	name := id
	if m.threadName != nil {
		n, err := m.threadName(dataset)
		if err != nil {
			return err
		}
		name = n
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thread[id] = threadCredential{name: name, dataset: slices.Clone(dataset)}
	return nil
}

// Remove implements [CredentialStore].
func (m *MemoryCredentials) Remove(_ context.Context, kind CredentialKind, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch kind {
	case CredentialWiFi:
		if _, ok := m.wifi[id]; ok {
			delete(m.wifi, id)
			return nil
		}
	case CredentialThread:
		if _, ok := m.thread[id]; ok {
			delete(m.thread, id)
			return nil
		}
	}
	return fmt.Errorf("%w: %s %q", ErrUnknownCredential, kind, id)
}

// WiFi is the stored Wi-Fi credential id names, for a commissioning.
func (m *MemoryCredentials) WiFi(id string) (ssid, passphrase string, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.wifi[id]
	return c.ssid, c.passphrase, ok
}

// Thread is the stored Thread dataset id names, for a commissioning.
func (m *MemoryCredentials) Thread(id string) (dataset []byte, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.thread[id]
	return slices.Clone(c.dataset), ok
}

// The in-memory controller's fixed values. The codes are the example
// codes of the Matter specification's onboarding chapter (passcode
// 20202021, discriminator 3840), which no real device uses.
const (
	memoryPasscode   = 20202021
	memoryManualCode = "34970112332"
	memoryQRCode     = "MT:Y.K9042C00KA0648G00"
	memoryInstance   = "1A2B3C4D5E6F7788"
	memoryLongDisc   = 3840
)

// memoryTime is the in-memory controller's clock: every time it reports
// is relative to it, so recorded answers stay the same.
var memoryTime = time.Date(2026, 10, 20, 8, 0, 0, 0, time.UTC)

// MemoryController is an in-memory [ControllerSource]: two commissioned
// nodes — a light on Wi-Fi with a pending update and a Binding cluster,
// and a sleepy ICD contact sensor on Thread — one commissionable device a
// discovery finds, a credential store and an update queue. The handler
// tests record the contract's fixtures from it; a host shows the pages
// with it before it has a controller. It is safe for concurrent use.
type MemoryController struct {
	mu      sync.Mutex
	fabric  ControllerFabric
	nodes   map[uint64]*memoryNode
	creds   *MemoryCredentials
	updates *memoryUpdates
	next    uint64
	subs    *Events
	// paired is whether the commissionable device was commissioned.
	paired bool
}

var _ ControllerSource = (*MemoryController)(nil)

// NewMemoryController returns the in-memory controller in its initial
// state.
func NewMemoryController() *MemoryController {
	m := &MemoryController{
		fabric: ControllerFabric{
			FabricID: 0x0000000000000001, CompressedFabricID: 0x87E1B004E235A130, FabricIndex: 1,
			NodeID: 0x000000000001B669, VendorID: 0xFFF1, Label: "fabric-ui",
		},
		nodes: map[uint64]*memoryNode{},
		creds: NewMemoryCredentials(nil),
		next:  0x13,
		subs:  NewEvents(),
	}
	m.updates = &memoryUpdates{m: m}
	_ = m.creds.SetWiFi(context.Background(), "home", "HomeNet", "secret-passphrase")
	_ = m.creds.SetThread(context.Background(), "mesh", []byte{0x03, 0x04, 'M', 'e', 's', 'h'})
	light := newMemoryNode(m, 0x11, memoryLight())
	sensor := newMemoryNode(m, 0x12, memorySensor())
	m.nodes[light.id], m.nodes[sensor.id] = light, sensor
	m.updates.queue = []PendingUpdate{{
		NodeID: light.id, VendorID: 0xFFF1, ProductID: 0x8001, TargetSoftwareVersion: 2,
		Status: "queued", ProgressStatus: "WaitForConsent",
	}}
	return m
}

func (m *MemoryController) publish(e ControllerEvent) {
	m.subs.Publish(Event{Type: e.Type, Data: e})
}

// Fabric implements [ControllerSource].
func (m *MemoryController) Fabric() ControllerFabric {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fabric
}

// Capabilities implements [ControllerSource]; the in-memory controller
// has no radio.
func (m *MemoryController) Capabilities() ControllerCapabilities { return ControllerCapabilities{} }

// Nodes implements [ControllerSource].
func (m *MemoryController) Nodes() []ControllerNode {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]ControllerNode, 0, len(m.nodes))
	for _, id := range slices.Sorted(maps.Keys(m.nodes)) {
		out = append(out, m.nodes[id])
	}
	return out
}

// Node implements [ControllerSource].
func (m *MemoryController) Node(id uint64) (ControllerNode, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[id]
	return n, ok
}

// Pair implements [ControllerSource]: the commissionable device's manual
// or QR code commissions it once; another code that reads as a code finds
// no device; anything else is an invalid code.
func (m *MemoryController) Pair(_ context.Context, code string, _ PairOptions) (ControllerNode, error) {
	code = strings.ReplaceAll(strings.TrimSpace(code), "-", "")
	if !strings.HasPrefix(code, "MT:") && !isDigits(code) {
		return nil, fmt.Errorf("%w: not a QR payload or a manual pairing code", ErrInvalidArgument)
	}
	if code != memoryManualCode && code != memoryQRCode {
		return nil, ErrNoDevice
	}
	return m.commission()
}

func isDigits(s string) bool {
	if len(s) != 11 && len(s) != 21 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (m *MemoryController) commission() (ControllerNode, error) {
	m.mu.Lock()
	if m.paired {
		m.mu.Unlock()
		return nil, ErrNoDevice
	}
	m.paired = true
	n := newMemoryNode(m, m.next, memoryPlug())
	m.next++
	m.nodes[n.id] = n
	m.mu.Unlock()
	m.publish(ControllerEvent{Type: EventNodeAdded, Node: n.id})
	return n, nil
}

// Discover implements [ControllerSource]: the commissionable device,
// until it is commissioned, when the filter selects it.
func (m *MemoryController) Discover(_ context.Context, f DiscoverFilter, _ time.Duration) ([]Discovered, error) {
	m.mu.Lock()
	paired := m.paired
	m.mu.Unlock()
	d := Discovered{
		InstanceName: memoryInstance, Hostname: "plug-1.local.", Addresses: []string{"fd00::12"}, Port: 5540,
		LongDiscriminator: memoryLongDisc, VendorID: 0xFFF1, ProductID: 0x8002, DeviceType: 0x010A,
		DeviceName: "Plug", CommissioningMode: 1, PairingHint: 33,
	}
	if paired {
		return []Discovered{}, nil
	}
	switch f.Kind {
	case "", "any":
	case "long_discriminator":
		if f.Value != uint32(d.LongDiscriminator) {
			return []Discovered{}, nil
		}
	case "short_discriminator":
		if f.Value != uint32(d.LongDiscriminator>>8) {
			return []Discovered{}, nil
		}
	case "vendor":
		if f.Value != uint32(d.VendorID) {
			return []Discovered{}, nil
		}
	case "device_type":
		if f.Value != d.DeviceType {
			return []Discovered{}, nil
		}
	default:
		return nil, fmt.Errorf("%w: filter %q", ErrInvalidArgument, f.Kind)
	}
	return []Discovered{d}, nil
}

// CommissionDiscovered implements [ControllerSource].
func (m *MemoryController) CommissionDiscovered(_ context.Context, d Discovered, passcode uint32, _ PairOptions) (ControllerNode, error) {
	if d.InstanceName != memoryInstance {
		return nil, ErrNoDevice
	}
	if passcode != memoryPasscode {
		return nil, fmt.Errorf("%w: PASE failed: the device refused the passcode", ErrNodeFailed)
	}
	return m.commission()
}

// Decommission implements [ControllerSource].
func (m *MemoryController) Decommission(_ context.Context, id uint64) error {
	m.mu.Lock()
	if _, ok := m.nodes[id]; !ok {
		m.mu.Unlock()
		return ErrNodeNotFound
	}
	delete(m.nodes, id)
	m.mu.Unlock()
	m.publish(ControllerEvent{Type: EventNodeRemoved, Node: id})
	return nil
}

// SetFabricLabel implements [ControllerSource].
func (m *MemoryController) SetFabricLabel(_ context.Context, label string) error {
	m.mu.Lock()
	m.fabric.Label = label
	m.mu.Unlock()
	return nil
}

// Credentials implements [ControllerSource].
func (m *MemoryController) Credentials() CredentialStore { return m.creds }

// SoftwareUpdates implements [ControllerSource].
func (m *MemoryController) SoftwareUpdates() SoftwareUpdateSource { return m.updates }

// Observe implements [ControllerSource].
func (m *MemoryController) Observe(fn func(ControllerEvent)) (unsubscribe func()) {
	return m.subs.Subscribe(func(e Event) {
		if ce, ok := e.Data.(ControllerEvent); ok {
			fn(ce)
		}
	})
}

// memoryUpdates is the in-memory update queue.
type memoryUpdates struct {
	m     *MemoryController
	queue []PendingUpdate
}

func (u *memoryUpdates) Queue() []PendingUpdate {
	u.m.mu.Lock()
	defer u.m.mu.Unlock()
	return slices.Clone(u.queue)
}

func (u *memoryUpdates) Check(context.Context) ([]AvailableUpdate, error) {
	return []AvailableUpdate{{NodeID: 0x11, VendorID: 0xFFF1, ProductID: 0x8001, SoftwareVersion: 2, SoftwareVersionString: "1.1.0"}}, nil
}

func (u *memoryUpdates) Consent(_ context.Context, node uint64, vendorID, productID uint16, version uint32) error {
	u.m.mu.Lock()
	defer u.m.mu.Unlock()
	for i, p := range u.queue {
		if p.NodeID == node && p.VendorID == vendorID && p.ProductID == productID && p.TargetSoftwareVersion == version {
			u.queue[i].Status, u.queue[i].ProgressStatus, u.queue[i].ProgressAt = "in_progress", "Querying", memoryTime
			return nil
		}
	}
	u.queue = append(u.queue, PendingUpdate{NodeID: node, VendorID: vendorID, ProductID: productID, TargetSoftwareVersion: version, Status: "queued"})
	return nil
}

// memorySpec is what a memory node is made of.
type memorySpec struct {
	basic     NodeBasic
	record    NodeRecord
	props     NodeProperties
	endpoints []NodeEndpoint
	values    map[AttributePath]json.RawMessage
	conn      ConnectionState
	icd       *ICDState
}

// memoryNode is one node of the in-memory controller.
type memoryNode struct {
	m    *MemoryController
	id   uint64
	mu   sync.Mutex
	spec memorySpec
	// disabled and subscribed are the node's management state.
	disabled, subscribed bool
}

func newMemoryNode(m *MemoryController, id uint64, s memorySpec) *memoryNode {
	s.record.FabricIndex = 1
	// Every node holds our fabric and the controller's Administer entry.
	s.values[AttributePath{0, clusterOperationalCredentials, attributeCurrentFabricIndex}] = json.RawMessage(`1`)
	s.values[AttributePath{0, clusterOperationalCredentials, attributeFabrics}] = json.RawMessage(fmt.Sprintf(
		`[{"rootPublicKey":"BA==","vendorId":65521,"fabricId":1,"nodeId":%d,"label":"fabric-ui","fabricIndex":1},`+
			`{"rootPublicKey":"BA==","vendorId":4937,"fabricId":2,"nodeId":%d,"label":"Home","fabricIndex":2}]`, id, id+0x1000))
	s.values[AttributePath{0, clusterAccessControl, attributeACL}] = json.RawMessage(
		`[{"privilege":5,"authMode":2,"subjects":[112233],"targets":null,"fabricIndex":1}]`)
	return &memoryNode{m: m, id: id, spec: s, subscribed: s.conn == ConnectionConnected}
}

func memoryLight() memorySpec {
	return memorySpec{
		basic: NodeBasic{VendorID: 0xFFF1, VendorName: "Test Vendor", ProductID: 0x8001, ProductName: "Light", NodeLabel: "Kitchen", SerialNumber: "L-0001", SoftwareVersion: "1.0.0"},
		record: NodeRecord{
			FabricIndexOnPeer: 1, Addresses: []string{"[fd00::11]:5540"}, Hostname: "light-1.local.",
			DiscoveredAt: memoryTime.Add(-time.Hour), CommissionedAt: memoryTime.Add(-time.Hour),
		},
		props: NodeProperties{SupportsWiFi: true, MainsPowered: true, SpecificationVersion: 0x01060100, DeviceTypes: []uint32{0x0016, 0x0100}},
		endpoints: []NodeEndpoint{
			{Number: 0, DeviceTypes: []NodeDeviceType{{ID: 0x0016, Name: "RootNode", Revision: 4}}, Clusters: []NodeCluster{
				{ID: 0x001D, Name: "Descriptor", Revision: 3},
				{ID: clusterAccessControl, Name: "AccessControl", Revision: 3},
				{ID: 0x0028, Name: "BasicInformation", Revision: 5},
				{ID: clusterOperationalCredentials, Name: "OperationalCredentials", Revision: 2},
			}},
			{Number: 1, DeviceTypes: []NodeDeviceType{{ID: 0x0100, Name: "OnOffLight", Revision: 3}}, Clusters: []NodeCluster{
				{ID: 0x0003, Name: "Identify", Revision: 6},
				{ID: 0x0006, Name: "OnOff", Revision: 6, Features: []string{"Lighting"}},
				{ID: clusterBinding, Name: "Binding", Revision: 1},
			}},
		},
		values: map[AttributePath]json.RawMessage{
			{0, 0x0028, 0x0003}: json.RawMessage(`"Light"`),
			{0, 0x0028, 0x0005}: json.RawMessage(`"Kitchen"`),
			{1, 0x0006, 0x0000}: json.RawMessage(`true`),
			{1, clusterBinding, attributeBinding}: json.RawMessage(
				`[{"node":18,"endpoint":1,"cluster":6,"fabricIndex":1}]`),
		},
		conn: ConnectionConnected,
	}
}

func memorySensor() memorySpec {
	return memorySpec{
		basic: NodeBasic{VendorID: 0xFFF1, VendorName: "Test Vendor", ProductID: 0x8003, ProductName: "Contact Sensor", SoftwareVersion: "2.3"},
		record: NodeRecord{
			FabricIndexOnPeer: 1, Addresses: []string{"[fd00::12:1]:5540"},
			DiscoveredAt: memoryTime.Add(-2 * time.Hour), CommissionedAt: memoryTime.Add(-2 * time.Hour),
		},
		props: NodeProperties{
			SupportsThread: true, BatteryPowered: true, IntermittentlyConnected: true, LongIdleTimeOperating: true,
			ThreadSleepyEndDevice: true, IdleModeDuration: time.Hour, SpecificationVersion: 0x01040000, DeviceTypes: []uint32{0x0016, 0x0015},
		},
		endpoints: []NodeEndpoint{
			{Number: 0, DeviceTypes: []NodeDeviceType{{ID: 0x0016, Name: "RootNode", Revision: 4}}, Clusters: []NodeCluster{
				{ID: 0x001D, Name: "Descriptor", Revision: 3}, {ID: 0x0046, Name: "IcdManagement", Revision: 3, Features: []string{"CheckInProtocolSupport", "LongIdleTimeSupport"}},
			}},
			{Number: 1, DeviceTypes: []NodeDeviceType{{ID: 0x0015, Name: "ContactSensor", Revision: 2}}, Clusters: []NodeCluster{
				{ID: 0x0045, Name: "BooleanState", Revision: 2},
			}},
		},
		values: map[AttributePath]json.RawMessage{
			{1, 0x0045, 0x0000}: json.RawMessage(`false`),
		},
		conn: ConnectionWaitingForDiscovery,
		icd: &ICDState{
			LITSupported: true, Registered: true, OperatingMode: "lit", Available: true,
			NextCheckIn: memoryTime.Add(time.Hour), LastCheckIn: memoryTime,
		},
	}
}

func memoryPlug() memorySpec {
	return memorySpec{
		basic: NodeBasic{VendorID: 0xFFF1, VendorName: "Test Vendor", ProductID: 0x8002, ProductName: "Plug"},
		record: NodeRecord{
			FabricIndexOnPeer: 1, Addresses: []string{"[fd00::12]:5540"}, Hostname: "plug-1.local.",
			DiscoveredAt: memoryTime, CommissionedAt: memoryTime, Discriminator: ptr(uint16(memoryLongDisc)),
			VendorID: ptr(uint16(0xFFF1)), ProductID: ptr(uint16(0x8002)), DeviceType: ptr(uint32(0x010A)), DeviceName: "Plug",
		},
		props: NodeProperties{SupportsEthernet: true, MainsPowered: true, DeviceTypes: []uint32{0x0016, 0x010A}},
		endpoints: []NodeEndpoint{
			{Number: 0, DeviceTypes: []NodeDeviceType{{ID: 0x0016, Name: "RootNode", Revision: 4}}},
			{Number: 1, DeviceTypes: []NodeDeviceType{{ID: 0x010A, Name: "OnOffPlugInUnit", Revision: 4}}, Clusters: []NodeCluster{{ID: 0x0006, Name: "OnOff", Revision: 6}}},
		},
		values: map[AttributePath]json.RawMessage{{1, 0x0006, 0x0000}: json.RawMessage(`false`)},
		conn:   ConnectionConnected,
	}
}

func (n *memoryNode) ID() uint64 { return n.id }

func (n *memoryNode) Record() NodeRecord {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.spec.record
}

func (n *memoryNode) Properties() NodeProperties {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.spec.props
}

func (n *memoryNode) Basic() NodeBasic {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.spec.basic
}

func (n *memoryNode) ConnectionState() ConnectionState {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.disabled {
		return ConnectionDisconnected
	}
	return n.spec.conn
}

func (n *memoryNode) SubscriptionActive() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.subscribed && !n.disabled && n.spec.conn == ConnectionConnected
}

func (n *memoryNode) Disabled() bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.disabled
}

func (n *memoryNode) Endpoints() []NodeEndpoint {
	n.mu.Lock()
	defer n.mu.Unlock()
	return slices.Clone(n.spec.endpoints)
}

// statusUnsupportedAttribute and statusUnsupportedCommand are the
// Interaction Model statuses the in-memory node answers for a path it
// does not have.
const (
	statusUnsupportedAttribute = 0x86
	statusUnsupportedCommand   = 0x81
)

func (n *memoryNode) Read(_ context.Context, paths []AttributePath, _ bool) ([]AttributeValue, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.disabled {
		return nil, fmt.Errorf("%w: the node is disabled", ErrNodeFailed)
	}
	out := make([]AttributeValue, 0, len(paths))
	for _, p := range paths {
		if v, ok := n.spec.values[p]; ok {
			out = append(out, AttributeValue{Path: p, Value: v})
		} else {
			out = append(out, AttributeValue{Path: p, Status: statusUnsupportedAttribute})
		}
	}
	return out, nil
}

func (n *memoryNode) Write(_ context.Context, p AttributePath, value any, _ time.Duration) (WriteStatus, error) {
	var raw json.RawMessage
	switch v := value.(type) {
	case json.RawMessage:
		raw = v
	case []ACLEntry:
		raw = mustJSON(aclValue(v))
	case []BindingTarget:
		raw = mustJSON(bindingValue(v))
	default:
		return WriteStatus{}, fmt.Errorf("%w: a value of type %T", ErrInvalidArgument, value)
	}
	n.mu.Lock()
	_, ok := n.spec.values[p]
	if ok {
		n.spec.values[p] = raw
	}
	n.mu.Unlock()
	if !ok {
		return WriteStatus{Status: statusUnsupportedAttribute}, nil
	}
	n.m.publish(ControllerEvent{Type: EventNodeUpdated, Node: n.id})
	return WriteStatus{}, nil
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// aclValue is an ACL as a read renders it, by field name, with our
// fabric's index.
func aclValue(entries []ACLEntry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		m := map[string]any{"privilege": e.Privilege, "authMode": e.AuthMode, "subjects": nil, "targets": nil, "fabricIndex": 1}
		if len(e.Subjects) > 0 {
			m["subjects"] = e.Subjects
		}
		if len(e.Targets) > 0 {
			ts := make([]map[string]any, 0, len(e.Targets))
			for _, t := range e.Targets {
				ts = append(ts, map[string]any{"cluster": t.Cluster, "endpoint": t.Endpoint, "deviceType": t.DeviceType})
			}
			m["targets"] = ts
		}
		out = append(out, m)
	}
	return out
}

func bindingValue(targets []BindingTarget) []map[string]any {
	out := make([]map[string]any, 0, len(targets))
	for _, t := range targets {
		m := map[string]any{"fabricIndex": 1}
		if t.Node != nil {
			m["node"] = *t.Node
		}
		if t.Group != nil {
			m["group"] = *t.Group
		}
		if t.Endpoint != nil {
			m["endpoint"] = *t.Endpoint
		}
		if t.Cluster != nil {
			m["cluster"] = *t.Cluster
		}
		out = append(out, m)
	}
	return out
}

func (n *memoryNode) Invoke(_ context.Context, p CommandPath, payload json.RawMessage, _ time.Duration) (InvokeResult, error) {
	switch {
	case p.Endpoint == 0 && p.Cluster == clusterOperationalCredentials && p.Command == commandRemoveFabric:
		var req struct {
			FabricIndex uint8 `json:"fabricIndex"`
		}
		if err := json.Unmarshal(payload, &req); err != nil {
			return InvokeResult{}, fmt.Errorf("%w: RemoveFabric: %w", ErrInvalidArgument, err)
		}
		n.mu.Lock()
		path := AttributePath{0, clusterOperationalCredentials, attributeFabrics}
		var fabrics []map[string]any
		_ = json.Unmarshal(n.spec.values[path], &fabrics)
		kept := fabrics[:0]
		status := 0x0B // NOCResponse InvalidFabricIndex
		for _, f := range fabrics {
			if idx, _ := f["fabricIndex"].(float64); uint8(idx) == req.FabricIndex {
				status = 0
				continue
			}
			kept = append(kept, f)
		}
		n.spec.values[path] = mustJSON(kept)
		n.mu.Unlock()
		return InvokeResult{HasResponse: true, Command: 0x08, Response: mustJSON(map[string]any{"statusCode": status, "fabricIndex": req.FabricIndex})}, nil
	case p.Endpoint == 1 && p.Cluster == 0x0006 && p.Command == 0x02:
		n.mu.Lock()
		path := AttributePath{1, 0x0006, 0x0000}
		v, ok := n.spec.values[path]
		if ok {
			n.spec.values[path] = mustJSON(string(v) != "true")
		}
		n.mu.Unlock()
		if !ok {
			return InvokeResult{Status: statusUnsupportedCommand}, nil
		}
		return InvokeResult{}, nil
	}
	return InvokeResult{Status: statusUnsupportedCommand}, nil
}

func (n *memoryNode) OpenCommissioningWindow(_ context.Context, timeout time.Duration) (WindowInfo, error) {
	if timeout == 0 {
		timeout = 900 * time.Second
	}
	return WindowInfo{
		ManualCode: memoryManualCode, QRCode: memoryQRCode, Passcode: memoryPasscode, Discriminator: memoryLongDisc,
		VendorID: n.Basic().VendorID, ProductID: n.Basic().ProductID, Timeout: timeout,
	}, nil
}

func (n *memoryNode) RevokeCommissioningWindow(context.Context) error { return nil }

func (n *memoryNode) Interview(context.Context) error {
	if n.Disabled() {
		return fmt.Errorf("%w: the node is disabled", ErrNodeFailed)
	}
	n.m.publish(ControllerEvent{Type: EventNodeUpdated, Node: n.id})
	return nil
}

func (n *memoryNode) Enable(context.Context) error {
	n.setDisabled(false)
	return nil
}

func (n *memoryNode) Disable(context.Context) error {
	n.setDisabled(true)
	return nil
}

func (n *memoryNode) setDisabled(d bool) {
	n.mu.Lock()
	n.disabled = d
	n.mu.Unlock()
	n.m.publish(ControllerEvent{Type: EventConnection, Node: n.id, Connection: n.ConnectionState()})
}

func (n *memoryNode) SetAutoSubscribe(_ context.Context, on bool) error {
	n.mu.Lock()
	n.subscribed = on
	n.mu.Unlock()
	n.m.publish(ControllerEvent{Type: EventSubscription, Node: n.id, Active: n.SubscriptionActive()})
	return nil
}

func (n *memoryNode) ICD() (ICDState, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.spec.icd == nil {
		return ICDState{}, false
	}
	return *n.spec.icd, true
}

func (n *memoryNode) RegisterICD(_ context.Context, allowMultiAdmin bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.spec.icd == nil {
		return fmt.Errorf("%w: the node has no ICD Management", ErrInvalidArgument)
	}
	if !allowMultiAdmin {
		return fmt.Errorf("%w: the node has administrators from other vendors that may not support LIT", ErrConflict)
	}
	n.spec.icd.Registered = true
	return nil
}

func (n *memoryNode) UnregisterICD(context.Context, bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.spec.icd == nil {
		return fmt.Errorf("%w: the node has no ICD Management", ErrInvalidArgument)
	}
	n.spec.icd.Registered = false
	return nil
}
