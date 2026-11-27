// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// ControllerSource is what the controller page family reads and drives: a
// go-fabric controller seen through types of this package, so the port
// carries no go-fabric type and the module's controller half does not
// move with go-fabric's facade. The adapter
// [thingfabric.dev/go-fabric-ui/fabricui/gofabric.ControllerOf] is the
// one implementation over *controller.Controller; [NewMemoryController] is
// an in-memory one for tests and demonstrations.
//
// 64-bit identifiers are uint64 here and hex strings only on the wire.
type ControllerSource interface {
	// Fabric is the controller's fabric.
	Fabric() ControllerFabric
	// Capabilities says what the controller can do beyond the core
	// routes; the credential and update routes also ask Credentials and
	// SoftwareUpdates.
	Capabilities() ControllerCapabilities
	// Nodes are the commissioned nodes, by node id.
	Nodes() []ControllerNode
	// Node is the commissioned node with the id; false when there is none.
	Node(id uint64) (ControllerNode, bool)
	// Pair commissions the device a QR payload or a manual pairing code
	// names. The code carries the passcode; it is never logged.
	Pair(ctx context.Context, code string, opts PairOptions) (ControllerNode, error)
	// Discover lists the commissionable devices filter selects that
	// answer within timeout. Kinds answered as such: [ErrInvalidArgument]
	// (400), [ErrNodeFailed] and context.DeadlineExceeded (502); any other
	// kind is 500 with the cause in the log (contract 0.16.0).
	Discover(ctx context.Context, filter DiscoverFilter, timeout time.Duration) ([]Discovered, error)
	// CommissionDiscovered commissions a device a Discover answered with
	// its passcode.
	CommissionDiscovered(ctx context.Context, d Discovered, passcode uint32, opts PairOptions) (ControllerNode, error)
	// Decommission removes our fabric from the node and forgets it. The
	// handler runs it detached from the request, bounded by two minutes.
	Decommission(ctx context.Context, id uint64) error
	// SetFabricLabel sets the fabric label; the nodes get it in the
	// background. An error wrapping [ErrConflict] says the host refuses
	// the change, a label it pins: 409 conflict (contract 0.15.0).
	// [ErrInvalidArgument] is 400; any other kind is 500 with the cause in
	// the log (contract 0.16.0).
	SetFabricLabel(ctx context.Context, label string) error
	// Credentials is the host's store of Wi-Fi and Thread credentials;
	// nil when the host keeps none.
	Credentials() CredentialStore
	// SoftwareUpdates is the controller's OTA provider; nil when it runs
	// none.
	SoftwareUpdates() SoftwareUpdateSource
	// Observe registers fn for every controller event from now on and
	// returns the function that unregisters it. fn must not block.
	Observe(fn func(ControllerEvent)) (unsubscribe func())
	// NetworkTopology is the controller's network as a graph: the
	// controller, its nodes and the links between them. refresh re-reads
	// the nodes' network diagnostics first. [ErrNotImplemented] when the
	// source serves no graph (GET /controller/topology answers 501).
	// [ErrInvalidArgument] is 400, [ErrNodeFailed] and
	// context.DeadlineExceeded 502; any other kind is 500 with the cause
	// in the log (contract 0.16.0).
	NetworkTopology(ctx context.Context, refresh bool) (NetworkTopology, error)
	// Diagnostics is a snapshot of the controller's sessions and peers
	// (GET /controller/diagnostics, contract 0.9.0). It is called only
	// when Capabilities().Diagnostics is true; false there answers 501
	// as well (a source that claims the capability and has nothing).
	Diagnostics() (ControllerDiagnostics, bool)
	// DiagnoseMDNS is what keeps the controller's DNS-SD from working,
	// errors first (GET /controller/mdns). It is called from a request
	// handler only, never from a callback of the source.
	DiagnoseMDNS() []MDNSFinding
	// DiagnosticEvents is the controller's diagnostic trace, newest
	// first, as go-fabric's ring returns it (GET /controller/events
	// serves it in this order).
	DiagnosticEvents() []DiagnosticEvent
	// OnDiagnosticEvent registers fn for every entry the trace records
	// from now on (the stream's controller_diagnostic_event) and returns
	// the function that unregisters it; nil when the source serves no
	// diagnostics. fn must not block.
	OnDiagnosticEvent(fn func(DiagnosticEvent)) (unsubscribe func())
	// OnboardingInventory is every commissioned node with the onboarding
	// code kept for it, by node id (GET /controller/inventory, contract
	// 0.13.0). The codes are secrets: [SecretText] keeps them out of
	// every rendering. It is called only when
	// Capabilities().OnboardingCodes is true.
	OnboardingInventory() []OnboardingEntry
	// SetOnboardingCode attaches the code printed on the device to the
	// commissioned node and replaces a stored one; it answers the node's
	// entry with what was stored. The code is a plain string, as Pair's,
	// and never logged. Refusals: [*OnboardingInvalid],
	// [*OnboardingMismatch], [ErrNodeNotFound], [ErrConflict] (the node
	// is no longer commissioned); any other error answers 500.
	SetOnboardingCode(ctx context.Context, id uint64, code string) (OnboardingEntry, error)
	// ClearOnboardingCode removes the node's stored code; nil also when
	// it had none. [ErrNodeNotFound] for an unknown node; any other
	// error answers 500.
	ClearOnboardingCode(ctx context.Context, id uint64) error
}

// ControllerDiagnostics is the controller's diagnostic snapshot as GET
// /controller/diagnostics serves it (the contract's ControllerDiagnostics):
// the port's own shape of go-fabric's diag.SessionDiagnostics, so the
// port carries no go-fabric type.
type ControllerDiagnostics struct {
	// AsOf is when the snapshot was taken; every age is measured against
	// it.
	AsOf time.Time
	// Occupancy is the session-id space's usage; Capacity 0 when the
	// controller does not track it.
	Occupancy SessionOccupancy
	// Sessions are the open sessions, by peer, then session.
	Sessions []ControllerSession
	// Peers are the controller's peers, in the order they joined.
	Peers []ControllerPeer
}

// SessionOccupancy is how much of the session-id space is held.
type SessionOccupancy struct {
	Live, Reserved, Capacity, Free int
}

// ControllerSession is one open PASE or CASE session of the controller
// (the contract's Session with the fields of 0.9.0).
type ControllerSession struct {
	ID                      uint16
	FabricIndex             uint8
	PeerNodeID, LocalNodeID uint64
	PASE                    bool
	Subscriptions           int
	LastActivity            time.Time
	LastPeerActivity        time.Time
	Idle, PeerIdle          time.Duration
	CreatedAt               time.Time
	Resumed                 bool
	// Transport is "udp", "tcp" or "ble"; "" when unknown.
	Transport string
	PeerLost  bool
}

// ControllerPeer is one of the controller's peers (the contract's
// ControllerPeer, go-fabric diag.PeerInfo).
type ControllerPeer struct {
	FabricIndex uint8
	NodeID      uint64
	// Registered: a commissioned node of the controller's registry has
	// this address.
	Registered bool
	// State is the node's connection state; "" without a registered node.
	State     ConnectionState
	LoadError string
	// TimeOffline is how long the peer's connection process has been
	// trying; 0 when none runs.
	TimeOffline                         time.Duration
	OperationalAddress, OperationalHost string
	TCPUnsupported, Reachable           bool
	DiscoveredNames                     []string
	// LastCheckInAt and NextExpectedCheckIn are zero when unknown.
	LastCheckInAt, NextExpectedCheckIn time.Time
	SubscriptionActive                 bool
	SessionIDs                         []uint16
}

// MDNSFinding is one condition that keeps the controller's DNS-SD from
// working (the contract's MdnsFinding).
type MDNSFinding struct {
	// Severity is "error" or "warning".
	Severity, Code, Message string
	// Service is the instance the finding is about; "" for none.
	Service string
}

// DiagnosticEvent is one entry of the controller's diagnostic trace (the
// contract's DiagnosticEvent).
type DiagnosticEvent struct {
	At time.Time
	// Kind is pairing, session, discovery, reachability, icd, ota or
	// camera; Severity info, warning or error.
	Kind, Severity, Message string
	Detail                  map[string]string
}

// NetworkTopology is a controller's network as GET /controller/topology
// serves it (the contract's TopologyGraph with family "controller"): the
// port's own types, so the graph builder's go-fabric types stay in the
// adapter.
type NetworkTopology struct {
	// CollectedAt is when the diagnostics the graph rests on were read.
	CollectedAt time.Time
	// Nodes are the graph's nodes; Connections its edges between them.
	Nodes       []NetworkNode
	Connections []NetworkConnection
}

// NetworkNode is one node of a [NetworkTopology].
type NetworkNode struct {
	// ID is unique in the graph: "node" for the controller,
	// "peer:<16-digit hex node id>" for a commissioned node, any other
	// string for an access point or a router the controller did not
	// commission.
	ID string
	// Kind is one of the contract's TopologyNode kinds: node, peer,
	// thread_router, thread_end_device, wifi_ap, ethernet (the go-fabric
	// adapter makes an Ethernet node a peer with role ethernet), and
	// border_router (contract 0.8.0): a Thread border router, labelled by
	// its mDNS host name, ok while seen within 24 hours.
	Kind string
	// Label names the node; State is ok, unreachable, idle or unknown.
	Label, State string
	// NodeID is a commissioned node's operational id; 0 for the
	// controller and a node it did not commission.
	NodeID uint64
	// Role is the Thread routing role or the network kind; RLOC16 the
	// Thread RLOC16, nil when unknown; SSID and BSSID the Wi-Fi network.
	Role        string
	RLOC16      *uint16
	SSID, BSSID string
	// Details are the ordered key/value pairs a page shows for the node.
	Details []TopologyDetail
}

// NetworkConnection is one edge of a [NetworkTopology].
type NetworkConnection struct {
	// Source and Target are node ids of the graph.
	Source, Target string
	// Kind is contains, session or link; Strength one of strong, medium,
	// weak, none (observed dead), unknown, empty when not a link.
	Kind, Strength string
	// LQI and RSSI are a Thread link's quality indicator and signal
	// strength in dBm; nil when not reported.
	LQI  *uint8
	RSSI *int8
	// Label is an optional caption.
	Label string
}

// TopologyDetail is one key/value pair a topology page shows for a node.
type TopologyDetail struct {
	Key, Value string
}

// ControllerNode is one commissioned node.
type ControllerNode interface {
	// ID is the node's operational node id on our fabric.
	ID() uint64
	Record() NodeRecord
	Properties() NodeProperties
	// Basic is what the node's BasicInformation says, from the
	// controller's image of it (no read).
	Basic() NodeBasic
	ConnectionState() ConnectionState
	SubscriptionActive() bool
	Disabled() bool
	// Endpoints are the image's endpoints in number order.
	Endpoints() []NodeEndpoint
	// EndpointDetail is one endpoint of the image with its inventory,
	// its clusters without attribute values; nothing is read from the
	// node. ErrEndpointNotFound when the image holds no such endpoint.
	// The 404 kinds ([ErrNodeNotFound], [ErrEndpointNotFound],
	// [ErrClusterNotFound], [ErrNoDevice], [ErrUnknownCredential]) and
	// [ErrInvalidArgument] (400) are answered as such; any other kind is
	// 500 with the cause in the log (contract 0.16.0).
	EndpointDetail(ep uint16) (EndpointDetail, error)
	// ClusterDetail is one server cluster of an endpoint of the image
	// with the attribute values the controller holds; nothing is read
	// from the node. ErrEndpointNotFound or ErrClusterNotFound when the
	// image lacks it, ErrNotRead (503) while the endpoint's ServerList
	// names the cluster but the controller has not read it yet. The 404
	// kinds and [ErrInvalidArgument] (400) are answered as such; any other
	// kind is 500 with the cause in the log (contract 0.16.0).
	ClusterDetail(ep uint16, cluster uint32) (ClusterDetail, error)
	// Events are the events the controller received from the node, those
	// numbered since or later, oldest first; with limit > 0 only the
	// newest limit of them. Nothing is read from the node.
	Events(since uint64, limit int) []EventRecord
	// Read reads concrete attribute paths; a path the node refuses is an
	// AttributeValue with its status, not an error.
	Read(ctx context.Context, paths []AttributePath, fabricFiltered bool) ([]AttributeValue, error)
	// Write writes one attribute. value is a JSON value
	// (json.RawMessage) in the shape a read answers, or the typed list of
	// an ACL ([]ACLEntry) or a Binding ([]BindingTarget) attribute. A
	// timed of 0 writes untimed unless the attribute needs a timed write.
	Write(ctx context.Context, path AttributePath, value any, timed time.Duration) (WriteStatus, error)
	// Invoke invokes one command; payload is its fields by name as JSON,
	// nil for none.
	Invoke(ctx context.Context, path CommandPath, payload json.RawMessage, timed time.Duration) (InvokeResult, error)
	// OpenCommissioningWindow opens an enhanced commissioning window for
	// another administrator; 0 is the node's default timeout. The result
	// carries the window's passcode.
	OpenCommissioningWindow(ctx context.Context, timeout time.Duration) (WindowInfo, error)
	RevokeCommissioningWindow(ctx context.Context) error
	// Interview reads the whole node again.
	Interview(ctx context.Context) error
	// Ping measures whether the node answers a Matter read, and how fast:
	// each address of its record on its own where the node holds a
	// session (PerAddress true), otherwise one read of BasicInformation
	// DataModelRevision through the session for every address. A silent
	// address or a failed read is a result with Reachable false, not an
	// error. The 404 kinds, [ErrInvalidArgument] (400), [ErrConflict]
	// (409) and [ErrNodeFailed] or context.DeadlineExceeded (502) are
	// answered as such (409 and 502 declared since contract 0.16.0); any
	// other kind is 500 with the cause in the log.
	Ping(ctx context.Context) (PingResult, error)
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	SetAutoSubscribe(ctx context.Context, on bool) error
	// ICD is the node's ICD state; false for a node without ICD
	// Management.
	ICD() (ICDState, bool)
	// RegisterICD registers this controller as the node's check-in
	// client; without allowMultiAdmin a node whose other administrators'
	// vendors may not support LIT is [ErrICDMultiAdmin].
	RegisterICD(ctx context.Context, allowMultiAdmin bool) error
	UnregisterICD(ctx context.Context, force bool) error
}

// ControllerFabric is the controller's fabric.
type ControllerFabric struct {
	FabricID           uint64
	CompressedFabricID uint64
	FabricIndex        uint8
	// NodeID is the controller's own operational node id.
	NodeID   uint64
	VendorID uint16
	Label    string
}

// ControllerCapabilities are what a controller can do beyond the core
// routes.
type ControllerCapabilities struct {
	// BLE: the controller commissions over Bluetooth LE.
	BLE bool
	// Diagnostics: the source serves the controller's diagnostics
	// (contract 0.9.0, capabilities.diagnostics) — Diagnostics,
	// DiagnoseMDNS, DiagnosticEvents and OnDiagnosticEvent. False: the
	// three diagnostic routes answer 501 without calling them, and the
	// stream carries no controller_diagnostic_event.
	Diagnostics bool
	// OnboardingCodes: the source serves the onboarding inventory
	// (contract 0.13.0, capabilities.onboarding_codes) —
	// OnboardingInventory, SetOnboardingCode and ClearOnboardingCode.
	// False, or a handler without Options.Admin: the three routes answer
	// 501 without calling them.
	OnboardingCodes bool
}

// SecretText is a text the port carries and nothing renders: an
// onboarding code. It is redacted under every fmt verb, in a log record
// of either slog handler, in encoding/json and every encoding that takes
// an encoding.TextMarshaler — also as a field of a struct, exported or
// not. [SecretText.Reveal] is the one way to the text. The zero value is
// no text. It is not comparable; compare the revealed texts.
type SecretText struct {
	_ [0]func()
	// text is behind a pointer because fmt prints a struct's unexported
	// field without calling its methods: a pointer prints as an address.
	text *string
}

// secretRedacted is what a SecretText renders as.
const secretRedacted = "redacted"

// NewSecretText wraps s; "" is the zero value.
func NewSecretText(s string) SecretText {
	if s == "" {
		return SecretText{}
	}
	return SecretText{text: &s}
}

// Reveal is the text — the secret. Call it where the code is shown, and
// nowhere a log or a serialiser can see the result.
func (s SecretText) Reveal() string {
	if s.text == nil {
		return ""
	}
	return *s.text
}

// IsZero reports no text.
func (s SecretText) IsZero() bool { return s.text == nil }

// String keeps the text out of a formatted value.
func (s SecretText) String() string {
	if s.text == nil {
		return ""
	}
	return secretRedacted
}

// GoString keeps the text out of a %#v.
func (s SecretText) GoString() string { return "fabricui.SecretText(" + s.String() + ")" }

// LogValue keeps the text out of a log record.
func (s SecretText) LogValue() slog.Value { return slog.StringValue(s.String()) }

// MarshalText keeps the text out of encoding/json and every encoding that
// takes an encoding.TextMarshaler. There is no UnmarshalText.
func (s SecretText) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

// OnboardingSource tells how a node's onboarding code came to be stored
// (the contract's OnboardingSource).
type OnboardingSource string

// The two sources, as go-fabric's client.OnboardingSource names them.
const (
	// OnboardingPaired: the device accepted the code's passcode at
	// commissioning.
	OnboardingPaired OnboardingSource = "paired"
	// OnboardingAttached: added afterwards; its passcode is unproven.
	OnboardingAttached OnboardingSource = "attached"
)

// OnboardingChecks are the identity checks an attached code passed; a
// check runs only when the node's side of it is known. All false for a
// paired code. No check proves the passcode or the unit.
type OnboardingChecks struct {
	Vendor, Product, Discriminator bool
}

// OnboardingEntry is one commissioned node of the onboarding inventory
// (the contract's OnboardingInventoryEntry, go-fabric
// controller.InventoryEntry). It renders as its node id and whether it
// has a code, never the code.
type OnboardingEntry struct {
	NodeID uint64
	// Label is the node's NodeLabel; "" when unset or not read yet.
	Label                                 string
	VendorName, ProductName, SerialNumber string
	// VendorID and ProductID are 0 when unknown.
	VendorID, ProductID uint16
	// HasCode reports a stored code; the fields below are zero without
	// one.
	HasCode bool
	// QRCode is the QR code's text ("MT:…"); zero when the stored code is
	// a manual pairing code, from which no QR code is made up.
	QRCode SecretText
	// ManualCode is the manual pairing code, 11 or 21 digits.
	ManualCode SecretText
	Source     OnboardingSource
	Checks     OnboardingChecks
	StoredAt   time.Time
}

// String keeps the codes out of a formatted entry.
func (e OnboardingEntry) String() string {
	code := "no code"
	if e.HasCode {
		code = "code redacted"
	}
	return fmt.Sprintf("OnboardingEntry(node %016X, %s)", e.NodeID, code)
}

// GoString keeps the codes out of a %#v.
func (e OnboardingEntry) GoString() string { return e.String() }

// LogValue keeps the codes out of a log record.
func (e OnboardingEntry) LogValue() slog.Value { return slog.StringValue(e.String()) }

// ConnectionState is a node's connection state.
type ConnectionState string

// The connection states, as matter.js NodeConnectionState names them.
const (
	ConnectionConnected           ConnectionState = "connected"
	ConnectionDisconnected        ConnectionState = "disconnected"
	ConnectionReconnecting        ConnectionState = "reconnecting"
	ConnectionWaitingForDiscovery ConnectionState = "waiting_for_discovery"
)

// NodeRecord is a node's commissioning record.
type NodeRecord struct {
	// FabricIndex is our fabric's index on the controller.
	FabricIndex uint8
	// FabricIndexOnPeer is the index the node gave our fabric; 0 when
	// unknown.
	FabricIndexOnPeer uint8
	// Addresses are the node's known addresses ("[fe80::1%eth0]:5540").
	Addresses []string
	// Hostname is the host the commissionable advertisement's SRV record
	// named.
	Hostname string
	// OperationalHost is the operational SRV target that owned the node's
	// address (Addresses' first) when it was last reached; empty when
	// unknown. Not Hostname.
	OperationalHost string
	DiscoveredAt    time.Time
	CommissionedAt  time.Time
	// Discriminator, VendorID, ProductID and DeviceType are what the node
	// announced when it was found; nil when unknown.
	Discriminator *uint16
	VendorID      *uint16
	ProductID     *uint16
	DeviceType    *uint32
	DeviceName    string
}

// NodeProperties are what the controller derived about a node.
type NodeProperties struct {
	SupportsThread, SupportsWiFi, SupportsEthernet bool
	MainsPowered, BatteryPowered                   bool
	IntermittentlyConnected, LongIdleTimeOperating bool
	ThreadSleepyEndDevice                          bool
	// IdleModeDuration is the ICD idle mode duration; 0 when unknown.
	IdleModeDuration time.Duration
	// SpecificationVersion is the node's BasicInformation
	// SpecificationVersion; 0 when unknown.
	SpecificationVersion uint32
	DeviceTypes          []uint32
}

// NodeBasic is a node's BasicInformation as the controller's image holds
// it; empty fields are unknown.
type NodeBasic struct {
	VendorID        uint16
	VendorName      string
	ProductID       uint16
	ProductName     string
	NodeLabel       string
	SerialNumber    string
	SoftwareVersion string
}

// NodeEndpoint is one endpoint of a node's image.
type NodeEndpoint struct {
	Number uint16
	// Parent is the endpoint whose parts list names this one; nil for the
	// root and top-level endpoints.
	Parent *uint16
	Name   string
	// Reachable is BridgedDeviceBasicInformation Reachable; nil for an
	// endpoint that does not serve it.
	Reachable   *bool
	DeviceTypes []NodeDeviceType
	Clusters    []NodeCluster
}

// NodeDeviceType is a device type an endpoint declares.
type NodeDeviceType struct {
	ID       uint32
	Name     string
	Revision uint16
}

// NodeCluster is a server cluster of an endpoint.
type NodeCluster struct {
	ID       uint32
	Name     string
	Revision uint16
	Features []string
}

// EndpointDetail is one endpoint of a node's image with its inventory
// (contract 0.4.0, EndpointDetail).
type EndpointDetail struct {
	Number uint16
	// Parent is the endpoint whose parts list names this one; nil for the
	// root and top-level endpoints.
	Parent *uint16
	Name   string
	// Reachable is BridgedDeviceBasicInformation Reachable where the
	// endpoint serves it, true elsewhere.
	Reachable bool
	// DeviceTypes is the Descriptor's DeviceTypeList in its order.
	DeviceTypes []NodeDeviceType
	// Tags is the Descriptor's TagList.
	Tags []SemanticTag
	// Labels are the FixedLabel entries, then the UserLabel ones.
	Labels []EndpointLabel
	// Parts are the numbers of the endpoint's direct children.
	Parts []uint16
	// Scope is the snapshot scope a node-side endpoint was assembled from;
	// empty on the controller side.
	Scope string
	// Clusters are the server clusters in id order; EndpointDetail leaves
	// their attribute values out.
	Clusters []ClusterDetail
	// Conformance is the device type conformance verdict; nil when the
	// source judges none.
	Conformance *Conformance
}

// SemanticTag is one entry of a Descriptor's TagList.
type SemanticTag struct {
	// MfgCode is the manufacturer code; nil when null.
	MfgCode   *uint16
	Namespace uint8
	Tag       uint8
	Label     string
}

// EndpointLabel is one FixedLabel or UserLabel entry.
type EndpointLabel struct {
	Label, Value string
	// User marks a UserLabel entry.
	User bool
}

// The conformance statuses (ConformanceVerdict.status).
const (
	ConformanceUnknown    = "unknown"
	ConformanceIncomplete = "incomplete"
	ConformanceClean      = "clean"
	ConformanceViolated   = "violated"
)

// Conformance is an endpoint's device type conformance verdict.
type Conformance struct {
	// Status is one of the Conformance* constants.
	Status     string
	Violations []ConformanceViolation
	// Unjudged are the server clusters whose inside was not judged.
	Unjudged []ClusterElement
}

// ConformanceViolation is one departure from a device type requirement.
type ConformanceViolation struct {
	DeviceType string
	// DeviceTypeID is 0 for Base and where the violation names none.
	DeviceTypeID uint32
	Requirement  string
	Kind         string
	Detail       string
}

// ClusterDetail is one server cluster with its inventory (contract 0.4.0,
// ClusterDetail).
type ClusterDetail struct {
	ID uint32
	// Name is empty for a cluster no definition names.
	Name       string
	Revision   uint16
	FeatureMap uint32
	// Features are the features the FeatureMap sets, in bit order.
	Features []ClusterFeature
	// DataVersion is nil while unknown.
	DataVersion *uint32
	// Attributes are the AttributeList in id order.
	Attributes []ClusterAttribute
	Accepted   []ClusterElement
	Generated  []ClusterElement
	// Events are the EventList where the peer reports one.
	Events []ClusterElement
}

// ClusterFeature is one FeatureMap bit a cluster sets.
type ClusterFeature struct {
	// Code is the short name ("LT"), Title the long one ("Lighting"),
	// empty where no definition names it.
	Code, Title string
	Bit         uint8
}

// ClusterAttribute is one attribute of a ClusterDetail.
type ClusterAttribute struct {
	ID   uint32
	Name string
	// Value is the value as JSON by field name; nil when left out or
	// unknown.
	Value json.RawMessage
	// Status is the status the attribute was answered with; 0 on success.
	Status uint8
}

// ClusterElement is a command or event id with its name.
type ClusterElement struct {
	ID   uint32
	Name string
}

// EventRecord is one Matter event (contract 0.4.1, EventRecord): one the
// node served or one the controller received from a node.
type EventRecord struct {
	// Number is the node-wide event number.
	Number uint64
	// Priority is one of "debug", "info" and "critical"
	// ([EventPriorityName]).
	Priority string
	Endpoint uint16
	Cluster  uint32
	Event    uint32
	// Name is the event's name; empty where the schema names none.
	Name string
	// EpochMS is the epoch timestamp in milliseconds; nil when the event
	// carries none or one above the int64 range.
	EpochMS *int64
	// Data is the event's data as JSON by field name; nil without data or
	// when it does not render.
	Data json.RawMessage
}

// EventPriorityName is the contract's name of a Matter event priority
// (Matter Core §10.6.6.1: 0 debug, 1 info, 2 critical); a value the
// specification does not define is "debug".
func EventPriorityName(p uint8) string {
	switch p {
	case 1:
		return "info"
	case 2:
		return "critical"
	default:
		return "debug"
	}
}

// AttributePath is a concrete attribute path.
type AttributePath struct {
	Endpoint  uint16
	Cluster   uint32
	Attribute uint32
}

// CommandPath is a concrete command path.
type CommandPath struct {
	Endpoint uint16
	Cluster  uint32
	Command  uint32
}

// AttributeValue is a read's answer for one path: the value as JSON, or the
// node's status for it.
type AttributeValue struct {
	Path  AttributePath
	Value json.RawMessage
	// Status is the node's status for the path; 0 with a value.
	Status uint8
}

// WriteStatus is the node's answer to a write.
type WriteStatus struct {
	Status        uint8
	ClusterStatus *uint8
}

// InvokeResult is the node's answer to an invoke.
type InvokeResult struct {
	Status        uint8
	ClusterStatus *uint8
	// HasResponse reports a response command; Command is its id and
	// Response its fields by name.
	HasResponse bool
	Command     uint32
	Response    json.RawMessage
}

// PingResult is the outcome of [ControllerNode.Ping]: a Matter read, not
// an ICMP echo.
type PingResult struct {
	// PerAddress says whether each address was probed on its own; false
	// when one read through the node's session stands for every address.
	PerAddress bool
	// Reachable is whether the read answered — with PerAddress, whether
	// one address answered.
	Reachable bool
	// RTT is the read's round trip, with PerAddress the fastest address's;
	// zero when it failed.
	RTT time.Duration
	// Error is why the read failed, with PerAddress why no address
	// answered; empty when it answered.
	Error string
	// Addresses are the node's known addresses, each with its outcome.
	Addresses []PingAddress
}

// PingAddress is one address of a [PingResult].
type PingAddress struct {
	// Address is the address with its port, as NodeRecord.Addresses
	// lists it.
	Address   string
	Reachable bool
	RTT       time.Duration
	// Error is why this address did not answer; empty when it answered.
	Error string
}

// WindowInfo is an open commissioning window on a node. Passcode is a
// secret: it appears only in the answer of the admin route that opened
// the window.
type WindowInfo struct {
	ManualCode, QRCode  string
	Passcode            uint32
	Discriminator       uint16
	VendorID, ProductID uint16
	Timeout             time.Duration
}

// ICDState is a node's intermittently-connected-device state.
type ICDState struct {
	LITSupported bool
	Registered   bool
	// OperatingMode is "sit" or "lit"; empty when unknown.
	OperatingMode string
	Awake         bool
	Available     bool
	NextCheckIn   time.Time
	LastCheckIn   time.Time
}

// ACLEntry is one access control entry of our fabric on a node. Subjects
// are node ids, group ids or CATs; empty is every subject of the auth
// mode. Targets empty is the whole node.
type ACLEntry struct {
	// Privilege is 1 view, 2 proxy view, 3 operate, 4 manage, 5 administer.
	Privilege uint8
	// AuthMode is 1 PASE, 2 CASE, 3 group.
	AuthMode    uint8
	Subjects    []uint64
	Targets     []ACLTarget
	FabricIndex uint8
}

// ACLTarget narrows an ACL entry; a nil member is any.
type ACLTarget struct {
	Cluster    *uint32
	Endpoint   *uint16
	DeviceType *uint32
}

// BindingTarget is one binding of an endpoint: a node and endpoint, or a
// group; Cluster narrows it.
type BindingTarget struct {
	Node        *uint64
	Group       *uint16
	Endpoint    *uint16
	Cluster     *uint32
	FabricIndex uint8
}

// PairOptions are the network and bounds of one commissioning.
type PairOptions struct {
	// WiFiID and ThreadID name a stored credential ([CredentialStore])
	// to put the device on.
	WiFiID, ThreadID string
	// NetworkOnly never commissions over Bluetooth LE.
	NetworkOnly bool
	// TermsAndConditions is the Terms and Conditions acknowledgement the
	// commissioning sends with GeneralCommissioning's
	// SetTCAcknowledgements (contract 0.7.0; go-fabric
	// controller.PairOptions.TermsAndConditions, controller/controller.go:
	// 654-663 at v0.19.1); nil sends none, and a device that requires
	// terms then fails the commissioning with [*ErrTermsRequired].
	TermsAndConditions *TermsAcknowledgement
	// Progress, when set, is called on the commissioning's goroutine as
	// each step starts and ends (go-fabric controller.PairOptions.Progress,
	// controller/controller.go:720-727 at v0.22.2); it must not block. The handler
	// sets it to report commissioning_progress with the step.
	Progress func(PairingStep)
	// Transport, when set, is called once on the commissioning's
	// goroutine, before the first Progress report, with whether the
	// search runs over Bluetooth LE and, if not, why; it must not block.
	// A source without a central reports Used false with
	// [BLESkipNoCentral]. Pair reports it; CommissionDiscovered does not
	// (the device came from a DNS-SD discovery). The handler sets it to
	// report commissioning_progress with ble (contract 0.7.1).
	Transport func(BLEDecision)
	// CommissionTimeout, when positive, bounds the commissioning of a
	// device once it is reached, apart from the search for it (go-fabric
	// controller.PairOptions.CommissionTimeout, controller/controller.go:
	// 1135-1149 at v0.24.1): the search keeps its own bound. The handler
	// sets it to 600 s for a code without timeout_seconds (decision 71);
	// zero leaves the source's own default. The go-fabric adapter takes
	// the smaller of it and a positive ControllerOptions.Pair.
	// CommissionTimeout of the host, so a host's shorter bound stays.
	CommissionTimeout time.Duration
}

// BLEDecision is whether a commissioning from a code searches over
// Bluetooth LE and, if not, why (contract 0.7.1, BLEDecision). It is the
// source's decision from its own inputs — a central, the request's
// network_only and named credential, the code's discovery capabilities —
// not a step: go-fabric reports the BLE search's own Discovery.BLE 2.1
// step when it runs.
type BLEDecision struct {
	Used bool
	// Reason is why BLE is not used; empty when Used.
	Reason BLESkipReason
}

// BLESkipReason is why a commissioning does not search over Bluetooth
// LE: the contract's BLEDecision.reason, one vocabulary for the stream,
// the adapter's log line and the page.
type BLESkipReason string

// The reasons, in their precedence when several hold: the host has no
// central; the operator's network_only; the device's QR code without the
// BLE bit; no credential named, and a device on BLE has no network yet.
const (
	BLESkipNoCentral   BLESkipReason = "no_central"
	BLESkipNetworkOnly BLESkipReason = "network_only"
	BLESkipCode        BLESkipReason = "code_without_ble"
	BLESkipNoNetwork   BLESkipReason = "no_network"
)

// TermsAcknowledgement is the version of the terms a user accepted and
// the bitmap of the accepted terms (contract 0.7.0, TermsAcknowledgement).
type TermsAcknowledgement struct {
	Version, Acknowledgements uint16
}

// PairingStep is one step of a commissioning as it starts and ends
// (contract 0.7.0, PairingStep): matter.js's step number, sub-step and
// name verbatim, plus go-fabric's Discovery 2.1, PASE 2.2 and Subscribe
// 21.1 (go-fabric controller.PairingStep = commissioner.Step,
// controller/commissioner/progress.go:67-73 at v0.18.0), and its phase:
// "started", "succeeded", "failed" or "skipped". It carries no error
// text, so nothing a code echoes travels in an event.
type PairingStep struct {
	Number, Sub uint8
	Name, Phase string
}

// DiscoverFilter selects commissionable devices: Kind "any" (or empty),
// "long_discriminator", "short_discriminator", "vendor" or "device_type",
// with Value the discriminator, vendor id or device type.
type DiscoverFilter struct {
	Kind  string
	Value uint32
}

// Discovered is a commissionable device a discovery answered.
type Discovered struct {
	InstanceName      string
	Hostname          string
	Addresses         []string
	Port              uint16
	LongDiscriminator uint16
	// HasLongDiscriminator says LongDiscriminator was given: 0 is a valid
	// discriminator (UR-P1-adpt-65). A discovery sets it with every
	// device it lists; a commissioning request when it names one.
	HasLongDiscriminator bool
	VendorID             uint16
	ProductID            uint16
	DeviceType           uint32
	DeviceName           string
	CommissioningMode    uint8
	PairingHint          uint32
	PairingInstruction   string
}

// CredentialStore is the host's store of network credentials for a
// commissioning: Wi-Fi SSIDs with their passphrase, Thread operational
// datasets. The routes read names only; a secret goes in and never comes
// back out.
//
// A method's error answers its status only where the route declares it
// (contract 0.16.0): [ErrInvalidArgument] and [ErrInvalidCredential] 400
// on SetWiFi, SetThread and Remove, the 404 kinds ([ErrUnknownCredential]
// among them) on Remove; any other kind, and every error of List, is 500
// with the cause in the log.
type CredentialStore interface {
	// List is every stored credential, by kind and id; Name is the SSID
	// or the Thread network name.
	List(ctx context.Context) ([]Credential, error)
	SetWiFi(ctx context.Context, id, ssid, passphrase string) error
	// SetThread stores an operational dataset (MeshCoP TLVs); an
	// invalid one wraps [ErrInvalidCredential].
	SetThread(ctx context.Context, id string, dataset []byte) error
	// Remove forgets a credential; one the store does not hold wraps
	// [ErrUnknownCredential].
	Remove(ctx context.Context, kind CredentialKind, id string) error
}

// CredentialKind is "wifi" or "thread".
type CredentialKind string

// The credential kinds.
const (
	CredentialWiFi   CredentialKind = "wifi"
	CredentialThread CredentialKind = "thread"
)

// Credential is a stored credential's name, never its secret.
type Credential struct {
	Kind CredentialKind
	ID   string
	Name string
	// Default marks the host's default entry: the one a commissioning
	// started without a choice would use on the host's own interfaces
	// (contract 0.7.1, Credential.default); the pages preselect it. A
	// host that sets none lists no default. The lookup of a commissioning
	// stays by the id the request names: an empty id names nothing.
	Default bool
}

// SoftwareUpdateSource is the controller's OTA provider.
type SoftwareUpdateSource interface {
	// Queue is the update queue.
	Queue() []PendingUpdate
	// Check asks every node for an available update. When the check
	// failed for some nodes it returns the updates of the others with an
	// [UpdateCheckFailures] naming the nodes that failed (contract
	// 0.11.0, AvailableUpdateList.errors); any other error fails the
	// whole check.
	Check(ctx context.Context) ([]AvailableUpdate, error)
	// Consent consents to node's update to the version. The 404 kinds
	// ([ErrNodeNotFound] among them) and [ErrInvalidArgument] (400) are
	// answered as such; any other kind is 500 with the cause in the log
	// (contract 0.16.0).
	Consent(ctx context.Context, node uint64, vendorID, productID uint16, version uint32) error
}

// PendingUpdate is one entry of the update queue.
type PendingUpdate struct {
	NodeID                uint64
	VendorID, ProductID   uint16
	TargetSoftwareVersion uint32
	// Status is "queued", "in_progress" or "stalled".
	Status string
	// ProgressStatus is the last status the node reported; empty before
	// one.
	ProgressStatus string
	ProgressAt     time.Time
	// Transferred and Length are an in-progress transfer's bytes; nil
	// otherwise.
	Transferred, Length *uint64
}

// UpdateCheckFailures is the error of a [SoftwareUpdateSource.Check]
// that failed for some nodes only: the updates Check returns with it are
// those of the other nodes.
type UpdateCheckFailures []UpdateCheckFailure

// UpdateCheckFailure is one node whose update check failed.
type UpdateCheckFailure struct {
	NodeID uint64
	Err    error
}

func (f UpdateCheckFailures) Error() string {
	parts := make([]string, 0, len(f))
	for _, e := range f {
		parts = append(parts, fmt.Sprintf("node %016X: %v", e.NodeID, e.Err))
	}
	return "update check failed for " + strings.Join(parts, "; ")
}

// AvailableUpdate is an update a node can take.
type AvailableUpdate struct {
	NodeID                uint64
	VendorID, ProductID   uint16
	SoftwareVersion       uint32
	SoftwareVersionString string
}

// Controller event types the controller stream carries.
const (
	EventNodeAdded             = "node_added"
	EventNodeUpdated           = "node_updated"
	EventNodeRemoved           = "node_removed"
	EventConnection            = "connection"
	EventSubscription          = "subscription"
	EventCommissioningProgress = "commissioning_progress"
	EventUpdate                = "update"
	// EventTopologyChanged tells a page to reload GET /controller/topology.
	EventTopologyChanged = "topology_changed"
	// EventControllerDiagnosticEvent carries one [DiagnosticEvent] the
	// controller's trace recorded (contract 0.9.0); the handler publishes
	// it from [ControllerSource.OnDiagnosticEvent].
	EventControllerDiagnosticEvent = "controller_diagnostic_event"
)

// ControllerEvent is one event of the controller: Type is one of the
// Event* controller constants; Node is the node it is about, 0 for a
// controller-wide one.
type ControllerEvent struct {
	Type string
	Node uint64
	// Connection is the new state of a connection event.
	Connection ConnectionState
	// Active is the default subscription's state of a subscription event.
	Active bool
	// Phase ("started", "succeeded", "failed") and Detail are a
	// commissioning_progress event's.
	Phase, Detail string
	// Step is a commissioning_progress event's step while Phase is
	// "started" (contract 0.7.0); nil for the commissioning's own phases.
	Step *PairingStep
	// UpdateKind is an update event's kind — available, consent_changed,
	// check_finished, applying, done, failed, queued, transfer_started
	// (go-fabric softwareupdate.UpdateEventKind, controller/softwareupdate/
	// observe.go:29-59 at v0.18.0); UpdateVersion its target version for
	// the kinds that carry one, UpdateErr the text of its error (contract
	// 0.7.0). A ControllerSource emits the update events itself: the
	// handler publishes none.
	UpdateKind    string
	UpdateVersion *uint32
	UpdateErr     string
}

// The errors a ControllerSource reports and the routes map to problems.
var (
	// ErrNodeNotFound: no commissioned node with the id (404).
	ErrNodeNotFound = errors.New("fabricui: no such node")
	// ErrInvalidArgument: the request does not fit (400), for a pairing
	// code, a value that does not fit its attribute, an unknown filter.
	ErrInvalidArgument = errors.New("fabricui: invalid argument")
	// ErrNoDevice: no device answered a commissioning's discovery (404).
	ErrNoDevice = errors.New("fabricui: no device answered")
	// ErrConflict: the node or the controller's client refuses the action
	// in its state (409, problem conflict).
	ErrConflict = errors.New("fabricui: conflict")
	// ErrICDMultiAdmin: an ICD registration without allowMultiAdmin on a
	// node whose other administrators' vendors may not support LIT (409,
	// problem icd_multi_admin, contract 0.12.0). It wraps [ErrConflict],
	// so errors.Is(err, ErrConflict) holds for it too.
	ErrICDMultiAdmin = fmt.Errorf("%w: other administrators' vendors may not support LIT", ErrConflict)
	// ErrNodeFailed: the interaction with the node failed (502); its
	// text is the problem's detail.
	ErrNodeFailed = errors.New("fabricui: the node failed")
	// ErrInvalidCredential: a credential the store refuses (400).
	ErrInvalidCredential = errors.New("fabricui: invalid credential")
	// ErrUnknownCredential: no stored credential with the kind and id
	// (404).
	ErrUnknownCredential = errors.New("fabricui: no such credential")
	// ErrNotImplemented: the source does not serve the route yet. 501
	// from NetworkTopology (GET /controller/topology), the one port method
	// whose operation declares it (the diagnostics reads answer 501 from
	// Capabilities); from any other method it is 500, as an error of no
	// kind (decision 79).
	ErrNotImplemented = errors.New("fabricui: not implemented")
	// ErrEndpointNotFound: the node's image holds no such endpoint (404).
	ErrEndpointNotFound = errors.New("fabricui: no such endpoint")
	// ErrClusterNotFound: the endpoint holds no such server cluster (404).
	ErrClusterNotFound = errors.New("fabricui: no such cluster")
	// ErrNotRead: the controller has not read the element yet. 503 from
	// ControllerNode.ClusterDetail (GET
	// /controller/nodes/{id}/endpoints/{ep}/clusters/{cluster_hex}), the
	// one operation that declares it; from any other method it is 500
	// (decision 79).
	ErrNotRead = errors.New("fabricui: not read yet")
	// ErrOnboardingCode: a text that is no onboarding code (contract
	// 0.13.0); a source reports it as an [*OnboardingInvalid].
	ErrOnboardingCode = errors.New("fabricui: not an onboarding code")
	// ErrOnboardingMismatch: a readable onboarding code that names
	// another device (contract 0.13.0); a source reports it as an
	// [*OnboardingMismatch].
	ErrOnboardingMismatch = errors.New("fabricui: the onboarding code names another device")
)

// OnboardingInvalid is a text SetOnboardingCode cannot read as an
// onboarding code (400 onboarding_code_invalid). Reason says why — the
// problem's detail — and never holds the code. errors.Is(err,
// ErrOnboardingCode) holds.
type OnboardingInvalid struct {
	Reason string
}

func (e *OnboardingInvalid) Error() string { return ErrOnboardingCode.Error() + ": " + e.Reason }

// Is reports target as [ErrOnboardingCode].
func (e *OnboardingInvalid) Is(target error) bool { return target == ErrOnboardingCode }

// The identity fields of an [OnboardingMismatch], as the contract's
// Problem.field names them.
const (
	OnboardingFieldVendor        = "vendor"
	OnboardingFieldProduct       = "product"
	OnboardingFieldDiscriminator = "discriminator"
)

// OnboardingMismatch is a readable onboarding code that names another
// device (409 onboarding_code_mismatch); nothing was stored.
// errors.Is(err, ErrOnboardingMismatch) holds. None of its values is a
// secret: vendor and product ids are public, the discriminator is
// advertised (go-fabric client.OnboardingMismatchError).
type OnboardingMismatch struct {
	// Field is OnboardingFieldVendor, OnboardingFieldProduct or
	// OnboardingFieldDiscriminator.
	Field string
	// Code is the code's value of the field, Node the node's, as
	// compared: for the discriminator of a manual pairing code both are
	// the 4-bit form.
	Code, Node uint16
}

func (e *OnboardingMismatch) Error() string {
	return fmt.Sprintf("%v: %s", ErrOnboardingMismatch, e.Sentence())
}

// Sentence is the mismatch in English words: vendor and product ids as
// 0x and four hex digits, the discriminator decimal.
func (e *OnboardingMismatch) Sentence() string {
	switch e.Field {
	case OnboardingFieldVendor:
		return fmt.Sprintf("the code's vendor id is 0x%04X, the node's is 0x%04X", e.Code, e.Node)
	case OnboardingFieldProduct:
		return fmt.Sprintf("the code's product id is 0x%04X, the node's is 0x%04X", e.Code, e.Node)
	default:
		return fmt.Sprintf("the code's discriminator is %d, the node advertised %d", e.Code, e.Node)
	}
}

// Is reports target as [ErrOnboardingMismatch].
func (e *OnboardingMismatch) Is(target error) bool { return target == ErrOnboardingMismatch }

// ErrTermsRequired is a commissioning the device refused for want of an
// accepted version of its terms and conditions (contract 0.8.0, problem
// terms_required, 409): the device serves GeneralCommissioning's TC
// feature, holds no satisfying acceptance, and the request sent no
// acknowledgement or one below Version. go-fabric ends such a
// commissioning right after its information step, before ArmFailSafe and
// AddNOC (commissioner.TermsRequiredError, controller/commissioner/
// errors.go:158-176 at v0.19.1), so a second commissioning with
// [PairOptions.TermsAndConditions] set finds the device as the first did.
// A source reports it as a *ErrTermsRequired.
type ErrTermsRequired struct {
	// Version is the device's TCMinRequiredVersion, the lowest version an
	// acknowledgement has to name.
	Version uint16
	// Deadline is the device's TCUpdateDeadline in Matter epoch seconds
	// (from 2000-01-01T00:00:00Z); nil when the attribute is null.
	Deadline *uint32
}

func (e *ErrTermsRequired) Error() string {
	return fmt.Sprintf("fabricui: the device requires acknowledgement of its terms and conditions, version %d or later", e.Version)
}
