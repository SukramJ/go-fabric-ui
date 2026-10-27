// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/SukramJ/go-fabric/cluster/spec"
	"github.com/SukramJ/go-fabric/cluster/spec/accesscontrol"
	"github.com/SukramJ/go-fabric/cluster/spec/basicinformation"
	"github.com/SukramJ/go-fabric/cluster/spec/binding"
	"github.com/SukramJ/go-fabric/cluster/spec/icdmanagement"
	"github.com/SukramJ/go-fabric/cluster/spec/operationalcredentials"
	"github.com/SukramJ/go-fabric/controller/client"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// node is a commissioned client.Node as a fabricui.ControllerNode.
type node struct {
	s  *source
	n  *client.Node
	id uint64
}

func (w *node) ID() uint64 { return w.id }

func (w *node) Record() fabricui.NodeRecord {
	r := w.n.Record()
	out := fabricui.NodeRecord{
		FabricIndexOnPeer: r.FabricIndexOnPeer, Hostname: r.Hostname,
		DiscoveredAt: r.DiscoveredAt, CommissionedAt: r.CommissionedAt,
		Discriminator: r.Discriminator, VendorID: r.VendorID, ProductID: r.ProductID, DeviceType: r.DeviceType,
		DeviceName: r.DeviceName, Addresses: []string{},
	}
	if r.PeerAddress != nil {
		out.FabricIndex = r.PeerAddress.FabricIndex
	}
	for _, a := range r.Addresses {
		out.Addresses = append(out.Addresses, netip.AddrPortFrom(a.IP, a.Port).String())
	}
	return out
}

func (w *node) Properties() fabricui.NodeProperties {
	p := w.n.Properties()
	out := fabricui.NodeProperties{
		SupportsThread: p.SupportsThread, SupportsWiFi: p.SupportsWifi, SupportsEthernet: p.SupportsEthernet,
		MainsPowered: p.MainsPowered, BatteryPowered: p.BatteryPowered,
		IntermittentlyConnected: p.IntermittentlyConnected, LongIdleTimeOperating: p.LongIdleTimeOperating,
		ThreadSleepyEndDevice: p.ThreadSleepyEndDevice, DeviceTypes: append([]uint32{}, p.DeviceTypes...),
	}
	if p.IdleModeDuration != nil {
		out.IdleModeDuration = *p.IdleModeDuration
	}
	if p.SpecificationVersion != nil {
		out.SpecificationVersion = *p.SpecificationVersion
	}
	return out
}

// imageValue decodes an attribute of the image's root endpoint into v through its
// definition; false when the image lacks it or it does not fit.
func (w *node) imageValue(cluster, attr uint32, v any) bool {
	e := w.n.Endpoint(0)
	if e == nil {
		return false
	}
	c := e.Cluster(cluster)
	if c == nil {
		return false
	}
	val, ok := c.Value(attr)
	if !ok {
		return false
	}
	raw, err := client.DecodeJSON(client.AttributeField(cluster, attr), val, client.NameBased)
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, v) == nil
}

// Basic reads BasicInformation from the image, no interaction.
func (w *node) Basic() fabricui.NodeBasic {
	var b fabricui.NodeBasic
	const bi = basicinformation.ClusterID
	w.imageValue(bi, basicinformation.AttrVendorId, &b.VendorID)
	w.imageValue(bi, basicinformation.AttrVendorName, &b.VendorName)
	w.imageValue(bi, basicinformation.AttrProductId, &b.ProductID)
	w.imageValue(bi, basicinformation.AttrProductName, &b.ProductName)
	w.imageValue(bi, basicinformation.AttrNodeLabel, &b.NodeLabel)
	w.imageValue(bi, basicinformation.AttrSerialNumber, &b.SerialNumber)
	w.imageValue(bi, basicinformation.AttrSoftwareVersionString, &b.SoftwareVersion)
	return b
}

func (w *node) ConnectionState() fabricui.ConnectionState { return connectionOf(w.n.ConnectionState()) }

func (w *node) SubscriptionActive() bool { return w.n.SubscriptionActive() }
func (w *node) Disabled() bool           { return w.n.Disabled() }

func (w *node) Endpoints() []fabricui.NodeEndpoint {
	eps := w.n.Endpoints()
	out := make([]fabricui.NodeEndpoint, 0, len(eps))
	for _, ep := range eps {
		e := fabricui.NodeEndpoint{Number: ep.Number(), Name: ep.Name(), DeviceTypes: []fabricui.NodeDeviceType{}, Clusters: []fabricui.NodeCluster{}}
		if o := ep.Owner(); o != nil && o.Number() != 0 {
			p := o.Number()
			e.Parent = &p
		}
		if dt := ep.DeviceType(); dt != 0 {
			t := fabricui.NodeDeviceType{ID: dt}
			if d := ep.Device(); d != nil {
				t.Name = d.Name
			}
			if rev, ok := ep.DeviceRevision(); ok {
				t.Revision = rev
			}
			e.DeviceTypes = append(e.DeviceTypes, t)
		}
		for _, c := range ep.Clusters() {
			e.Clusters = append(e.Clusters, fabricui.NodeCluster{ID: c.ID(), Name: c.Name(), Revision: c.Revision(), Features: append([]string{}, c.Features()...)})
		}
		out = append(out, e)
	}
	return out
}

// Read reads concrete paths without data version filters, fabric-filtered
// or not — go-fabric-server's read_attribute (controllerapi/cmd_interact.go:68-110
// at v0.3.2) — and renders each value by name.
func (w *node) Read(ctx context.Context, paths []fabricui.AttributePath, fabricFiltered bool) ([]fabricui.AttributeValue, error) {
	req := client.ReadRequest{IncludeKnownVersions: true}
	for _, p := range paths {
		req.Attributes = append(req.Attributes, client.AttributeAt(p.Endpoint, p.Cluster, p.Attribute))
	}
	req.NoFabricFilter = !fabricFiltered
	rep, err := w.n.Read(ctx, req)
	if err != nil {
		return nil, nodeError(err)
	}
	got := map[fabricui.AttributePath]fabricui.AttributeValue{}
	for i := range rep.Attributes {
		d := &rep.Attributes[i]
		p := fabricui.AttributePath{Endpoint: d.Path.Endpoint, Cluster: d.Path.Cluster, Attribute: d.Path.Attribute}
		raw, err := client.DecodeJSON(client.AttributeField(p.Cluster, p.Attribute), d.Value, client.NameBased)
		if err != nil {
			raw = json.RawMessage("null")
		}
		got[p] = fabricui.AttributeValue{Path: p, Value: raw}
	}
	for _, st := range rep.Statuses {
		p := fabricui.AttributePath{Endpoint: st.Path.Endpoint, Cluster: st.Path.Cluster, Attribute: st.Path.Attribute}
		got[p] = fabricui.AttributeValue{Path: p, Status: uint8(st.Status)}
	}
	out := make([]fabricui.AttributeValue, 0, len(paths))
	for _, p := range paths {
		if v, ok := got[p]; ok {
			out = append(out, v)
		}
	}
	return out, nil
}

// fabricIndex is our fabric's index on the node: the record's, else the
// node's CurrentFabricIndex — the index a fabric-scoped entry carries.
func (w *node) fabricIndex(ctx context.Context) (uint8, error) {
	if idx := w.n.Record().FabricIndexOnPeer; idx != 0 {
		return idx, nil
	}
	path := fabricui.AttributePath{Endpoint: 0, Cluster: operationalcredentials.ClusterID, Attribute: operationalcredentials.AttrCurrentFabricIndex}
	values, err := w.Read(ctx, []fabricui.AttributePath{path}, true)
	if err != nil {
		return 0, err
	}
	for _, v := range values {
		var idx uint8
		if v.Status == 0 && json.Unmarshal(v.Value, &idx) == nil && idx != 0 {
			return idx, nil
		}
	}
	return 0, fmt.Errorf("%w: the node reports no CurrentFabricIndex", fabricui.ErrNodeFailed)
}

// aclList builds the AccessControl ACL from the generated definition's
// AccessControlEntryStruct, our fabric's index in every entry.
func aclList(entries []fabricui.ACLEntry, idx uint8) spec.List[accesscontrol.AccessControlEntryStruct] {
	out := make(spec.List[accesscontrol.AccessControlEntryStruct], 0, len(entries))
	for _, e := range entries {
		s := accesscontrol.AccessControlEntryStruct{
			Privilege:   accesscontrol.AccessControlEntryPrivilegeEnum(e.Privilege),
			AuthMode:    accesscontrol.AccessControlEntryAuthModeEnum(e.AuthMode),
			FabricIndex: idx,
		}
		s.Subjects, s.Targets = spec.NullOf[[]uint64](), spec.NullOf[[]accesscontrol.AccessControlTargetStruct]()
		if len(e.Subjects) > 0 {
			s.Subjects = spec.Nullable[[]uint64]{Value: append([]uint64{}, e.Subjects...)}
		}
		if len(e.Targets) > 0 {
			ts := make([]accesscontrol.AccessControlTargetStruct, 0, len(e.Targets))
			for _, t := range e.Targets {
				ts = append(ts, accesscontrol.AccessControlTargetStruct{
					Cluster: nullable(t.Cluster), Endpoint: nullable(t.Endpoint), DeviceType: nullable(t.DeviceType),
				})
			}
			s.Targets = spec.Nullable[[]accesscontrol.AccessControlTargetStruct]{Value: ts}
		}
		out = append(out, s)
	}
	return out
}

// nullable is v as a nullable member: nil is null, the wildcard.
func nullable[T any](v *T) spec.Nullable[T] {
	if v == nil {
		return spec.NullOf[T]()
	}
	return spec.Nullable[T]{Value: *v}
}

// bindingList builds the Binding list from the generated definition's
// TargetStruct.
func bindingList(targets []fabricui.BindingTarget, idx uint8) spec.List[binding.TargetStruct] {
	out := make(spec.List[binding.TargetStruct], 0, len(targets))
	for _, t := range targets {
		out = append(out, binding.TargetStruct{Node: t.Node, Group: t.Group, Endpoint: t.Endpoint, Cluster: t.Cluster, FabricIndex: idx})
	}
	return out
}

// Write writes one attribute in one write — go-fabric-server's
// write_attribute and, for the typed lists, set_acl_entry and
// set_node_binding (controllerapi/cmd_interact.go:285-340,
// cmd_acl.go:37-62 at v0.3.2): a JSON value by the attribute's definition
// (client.EncodeJSON), an ACL or a Binding list from the generated
// structs. The node's status for the path is the answer, not an error.
func (w *node) Write(ctx context.Context, p fabricui.AttributePath, value any, timed time.Duration) (fabricui.WriteStatus, error) {
	var enc client.Encodable
	switch v := value.(type) {
	case json.RawMessage:
		e, err := client.EncodeJSON(client.AttributeField(p.Cluster, p.Attribute), v)
		if err != nil {
			return fabricui.WriteStatus{}, nodeError(err)
		}
		enc = e
	case []fabricui.ACLEntry:
		if p.Cluster != accesscontrol.ClusterID || p.Attribute != accesscontrol.AttrAcl {
			return fabricui.WriteStatus{}, fmt.Errorf("%w: ACL entries go to AccessControl ACL", fabricui.ErrInvalidArgument)
		}
		idx, err := w.fabricIndex(ctx)
		if err != nil {
			return fabricui.WriteStatus{}, err
		}
		enc = aclList(v, idx)
	case []fabricui.BindingTarget:
		if p.Cluster != binding.ClusterID || p.Attribute != binding.AttrBinding {
			return fabricui.WriteStatus{}, fmt.Errorf("%w: binding targets go to Binding", fabricui.ErrInvalidArgument)
		}
		idx, err := w.fabricIndex(ctx)
		if err != nil {
			return fabricui.WriteStatus{}, err
		}
		enc = bindingList(v, idx)
	default:
		return fabricui.WriteStatus{}, fmt.Errorf("%w: a value of type %T", fabricui.ErrInvalidArgument, value)
	}
	opts := client.WriteOptions{Timed: timed > 0 || attributeTimed(p.Cluster, p.Attribute), TimedTimeout: timed}
	path := client.AttributeAt(p.Endpoint, p.Cluster, p.Attribute)
	statuses, err := w.n.Write(ctx, []client.WriteRequest{{Path: path, Value: enc}}, opts)
	if len(statuses) > 0 {
		return fabricui.WriteStatus{Status: uint8(statuses[0].Status), ClusterStatus: statuses[0].ClusterStatus}, nil
	}
	if se, ok := errors.AsType[*client.StatusError](err); ok {
		return fabricui.WriteStatus{Status: uint8(se.Status), ClusterStatus: se.ClusterStatus}, nil
	}
	if err != nil {
		return fabricui.WriteStatus{}, nodeError(err)
	}
	return fabricui.WriteStatus{}, nil
}

// attributeTimed reports an attribute whose write must be timed by its
// definition's access ("T").
func attributeTimed(cluster, attr uint32) bool {
	if def := spec.Lookup(cluster); def != nil {
		if a := def.Attribute(attr); a != nil {
			return a.Access.Timed
		}
	}
	return false
}

func commandTimed(cluster, cmd uint32) bool {
	if def := spec.Lookup(cluster); def != nil {
		if c := def.Command(cmd, spec.Request); c != nil {
			return c.Access.Timed
		}
	}
	return false
}

// Invoke invokes one command with its fields by name (client.EncodeJSON
// by the command's definition) — go-fabric-server's device_command
// (controllerapi/cmd_interact.go:344-400 at v0.3.2); the node's status is
// the answer, the response rendered by name.
func (w *node) Invoke(ctx context.Context, p fabricui.CommandPath, payload json.RawMessage, timed time.Duration) (fabricui.InvokeResult, error) {
	field := client.CommandField(p.Cluster, p.Command)
	var fields client.Encodable
	if t := bytes.TrimSpace(payload); field.Typed() || (len(t) > 0 && string(t) != "null") {
		e, err := client.EncodeJSON(field, payload)
		if err != nil {
			return fabricui.InvokeResult{}, nodeError(err)
		}
		fields = e
	}
	opts := client.InvokeOptions{Timed: timed > 0 || commandTimed(p.Cluster, p.Command), TimedTimeout: timed}
	res, err := w.n.Invoke(ctx, client.CommandAt(p.Endpoint, p.Cluster, p.Command), fields, opts)
	if se, ok := errors.AsType[*client.StatusError](err); ok {
		return fabricui.InvokeResult{Status: uint8(se.Status), ClusterStatus: se.ClusterStatus}, nil
	}
	if err != nil {
		return fabricui.InvokeResult{}, nodeError(err)
	}
	out := fabricui.InvokeResult{Status: uint8(res.Status), ClusterStatus: res.ClusterStatus}
	if res.HasFields {
		out.HasResponse, out.Command = true, res.Command
		raw, err := client.DecodeJSON(client.ResponseField(p.Cluster, res.Command), res.Fields, client.NameBased)
		if err == nil {
			out.Response = raw
		}
	}
	return out, nil
}

// OpenCommissioningWindow opens an enhanced window — client.Node.OpenCommissioningWindow,
// as go-fabric-server's open_commissioning_window (controllerapi/cmd_node.go:303-334).
func (w *node) OpenCommissioningWindow(ctx context.Context, timeout time.Duration) (fabricui.WindowInfo, error) {
	cw, err := w.n.OpenCommissioningWindow(ctx, timeout, client.WindowOptions{})
	if err != nil {
		return fabricui.WindowInfo{}, nodeError(err)
	}
	return fabricui.WindowInfo{
		ManualCode: cw.ManualCode, QRCode: cw.QRCode, Passcode: cw.Passcode, Discriminator: cw.Discriminator,
		VendorID: cw.VendorID, ProductID: cw.ProductID, Timeout: cw.Timeout,
	}, nil
}

func (w *node) RevokeCommissioningWindow(ctx context.Context) error {
	_, err := w.n.RevokeCommissioningWindow(ctx)
	return nodeError(err)
}

// Interview reads the whole node again without data version filters —
// go-fabric-server's interview_node (controllerapi/cmd_node.go:99-124).
func (w *node) Interview(ctx context.Context) error {
	req := client.ReadRequest{IncludeKnownVersions: true}
	req.Attributes = []client.AttributePath{{}}
	if _, err := w.n.Read(ctx, req); err != nil {
		return nodeError(err)
	}
	w.s.publish(fabricui.ControllerEvent{Type: fabricui.EventNodeUpdated, Node: w.id})
	return nil
}

func (w *node) Enable(ctx context.Context) error  { return nodeError(w.n.Enable(ctx)) }
func (w *node) Disable(ctx context.Context) error { return nodeError(w.n.Disable(ctx)) }

func (w *node) SetAutoSubscribe(ctx context.Context, on bool) error {
	return nodeError(w.n.SetAutoSubscribe(ctx, on))
}

// ICD is the node's ICD state — go-fabric-server's get_icd_state
// (controllerapi/cmd_icd.go:47-74 at v0.3.2): a node without ICD
// Management is not supported; the operating mode is the image's.
func (w *node) ICD() (fabricui.ICDState, bool) {
	ep := w.n.Endpoint(0)
	if ep == nil || ep.Cluster(icdmanagement.ClusterID) == nil {
		return fabricui.ICDState{}, false
	}
	c := w.n.ICD()
	if c == nil {
		return fabricui.ICDState{}, true
	}
	st := fabricui.ICDState{LITSupported: c.PeerSupportsLIT(), Registered: c.Registered(), Awake: c.Awake(), Available: c.Available()}
	var mode uint8
	if w.imageValue(icdmanagement.ClusterID, icdmanagement.AttrOperatingMode, &mode) {
		st.OperatingMode = "sit"
		if mode == 1 {
			st.OperatingMode = "lit"
		}
	}
	if t, ok := c.NextExpectedCheckIn(); ok {
		st.NextCheckIn = t
	}
	if t, ok := c.LastCheckInAt(); ok {
		st.LastCheckIn = t
	}
	return st, true
}

// RegisterICD registers us as the node's check-in client
// (client.ICDClient.Register; go-fabric-server's register_icd,
// controllerapi/cmd_icd.go:121-150).
func (w *node) RegisterICD(ctx context.Context, allowMultiAdmin bool) error {
	c := w.n.ICD()
	if c == nil {
		return fmt.Errorf("%w: the node has no ICD client", fabricui.ErrInvalidArgument)
	}
	return nodeError(c.Register(ctx, client.ICDRegisterOptions{AllowMultiAdmin: allowMultiAdmin}))
}

// UnregisterICD unregisters us, or with force forgets the registration
// without the node (client.ICDClient.Unregister / Forget; unregister_icd,
// controllerapi/cmd_icd.go:153-180).
func (w *node) UnregisterICD(ctx context.Context, force bool) error {
	c := w.n.ICD()
	if c == nil {
		return fmt.Errorf("%w: the node has no ICD client", fabricui.ErrInvalidArgument)
	}
	if force {
		return nodeError(c.Forget(ctx))
	}
	return nodeError(c.Unregister(ctx))
}
