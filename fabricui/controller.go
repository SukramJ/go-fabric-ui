// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// ControllerSource is what the controller page family reads and drives: a
// go-fabric controller seen through types of this package, so the port
// carries no go-fabric type and the module's controller half does not
// move with go-fabric's facade. The adapter
// [github.com/SukramJ/go-fabric-ui/fabricui/gofabric.ControllerOf] is the
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
	// answer within timeout.
	Discover(ctx context.Context, filter DiscoverFilter, timeout time.Duration) ([]Discovered, error)
	// CommissionDiscovered commissions a device a Discover answered with
	// its passcode.
	CommissionDiscovered(ctx context.Context, d Discovered, passcode uint32, opts PairOptions) (ControllerNode, error)
	// Decommission removes our fabric from the node and forgets it.
	Decommission(ctx context.Context, id uint64) error
	// SetFabricLabel sets the fabric label; the nodes get it in the
	// background.
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
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	SetAutoSubscribe(ctx context.Context, on bool) error
	// ICD is the node's ICD state; false for a node without ICD
	// Management.
	ICD() (ICDState, bool)
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
}

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
	Addresses      []string
	Hostname       string
	DiscoveredAt   time.Time
	CommissionedAt time.Time
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
	Parent      *uint16
	Name        string
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
	InstanceName       string
	Hostname           string
	Addresses          []string
	Port               uint16
	LongDiscriminator  uint16
	VendorID           uint16
	ProductID          uint16
	DeviceType         uint32
	DeviceName         string
	CommissioningMode  uint8
	PairingHint        uint32
	PairingInstruction string
}

// CredentialStore is the host's store of network credentials for a
// commissioning: Wi-Fi SSIDs with their passphrase, Thread operational
// datasets. The routes read names only; a secret goes in and never comes
// back out.
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
}

// SoftwareUpdateSource is the controller's OTA provider.
type SoftwareUpdateSource interface {
	// Queue is the update queue.
	Queue() []PendingUpdate
	// Check asks every node for an available update.
	Check(ctx context.Context) ([]AvailableUpdate, error)
	// Consent consents to node's update to the version.
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
	// ErrConflict: the node refuses the action in its state (409), an
	// ICD registration other administrators' vendors may not support.
	ErrConflict = errors.New("fabricui: conflict")
	// ErrNodeFailed: the interaction with the node failed (502); its
	// text is the problem's detail.
	ErrNodeFailed = errors.New("fabricui: the node failed")
	// ErrInvalidCredential: a credential the store refuses (400).
	ErrInvalidCredential = errors.New("fabricui: invalid credential")
	// ErrUnknownCredential: no stored credential with the kind and id
	// (404).
	ErrUnknownCredential = errors.New("fabricui: no such credential")
)
