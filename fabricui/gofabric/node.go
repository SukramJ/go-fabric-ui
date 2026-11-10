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
	"slices"
	"sync"
	"time"

	"github.com/SukramJ/go-fabric/cluster/spec"
	"github.com/SukramJ/go-fabric/cluster/spec/accesscontrol"
	"github.com/SukramJ/go-fabric/cluster/spec/basicinformation"
	"github.com/SukramJ/go-fabric/cluster/spec/binding"
	"github.com/SukramJ/go-fabric/cluster/spec/icdmanagement"
	"github.com/SukramJ/go-fabric/cluster/spec/operationalcredentials"
	"github.com/SukramJ/go-fabric/controller/client"
	"github.com/SukramJ/go-fabric/schema"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// The Matter 1.6.1 ids of the attributes the inventory reads from the
// image: Descriptor (0x001D) DeviceTypeList (0x0000), ServerList
// (0x0001) and TagList (0x0004); BridgedDeviceBasicInformation (0x0039)
// Reachable (0x0011); FixedLabel (0x0040) and UserLabel (0x0041)
// LabelList (0x0000) — the ids of go-fabric's generated cluster/spec
// packages, which the import rule (internal/arch) keeps this adapter
// from importing beyond the ones it encodes with.
const (
	clusterDescriptor       uint32 = 0x001D
	attrDeviceTypeList      uint32 = 0x0000
	attrServerList          uint32 = 0x0001
	attrTagList             uint32 = 0x0004
	clusterBridgedBasicInfo uint32 = 0x0039
	attrReachable           uint32 = 0x0011
	clusterFixedLabel       uint32 = 0x0040
	clusterUserLabel        uint32 = 0x0041
	attrLabelList           uint32 = 0x0000
)

// node is a commissioned client.Node as a fabricui.ControllerNode.
type node struct {
	s  *source
	n  *client.Node
	id uint64
}

func (w *node) ID() uint64 { return w.id }

func (w *node) Record() fabricui.NodeRecord { return nodeRecord(w.n.Record()) }

// nodeRecord is the port's record of go-fabric's.
func nodeRecord(r client.Record) fabricui.NodeRecord {
	out := fabricui.NodeRecord{
		FabricIndexOnPeer: r.FabricIndexOnPeer, Hostname: r.Hostname, OperationalHost: r.OperationalHost,
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
	return endpointValue(e, cluster, attr, v)
}

// endpointValue decodes an attribute of an endpoint of the image into v,
// rendered by name as Read renders it; false when the image lacks it or
// it does not fit.
func endpointValue(e *client.Endpoint, cluster, attr uint32, v any) bool {
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
		e := fabricui.NodeEndpoint{Number: ep.Number(), Name: ep.Name(), Parent: parentOf(ep), DeviceTypes: deviceTypes(ep), Clusters: []fabricui.NodeCluster{}}
		var reachable bool
		if endpointValue(ep, clusterBridgedBasicInfo, attrReachable, &reachable) {
			e.Reachable = &reachable
		}
		for _, c := range ep.Clusters() {
			e.Clusters = append(e.Clusters, fabricui.NodeCluster{ID: c.ID(), Name: c.Name(), Revision: c.Revision(), Features: append([]string{}, c.Features()...)})
		}
		out = append(out, e)
	}
	return out
}

// parentOf is the endpoint whose parts list names ep; nil for the root
// and a top-level endpoint.
func parentOf(ep *client.Endpoint) *uint16 {
	if o := ep.Owner(); o != nil && o.Number() != 0 {
		p := o.Number()
		return &p
	}
	return nil
}

// deviceTypes is the endpoint's Descriptor DeviceTypeList in its order,
// each with the schema's name; the endpoint's device type alone while the
// image holds no list.
func deviceTypes(ep *client.Endpoint) []fabricui.NodeDeviceType {
	var list []struct {
		DeviceType uint32 `json:"deviceType"`
		Revision   uint16 `json:"revision"`
	}
	out := []fabricui.NodeDeviceType{}
	if endpointValue(ep, clusterDescriptor, attrDeviceTypeList, &list) && len(list) > 0 {
		for _, dt := range list {
			out = append(out, fabricui.NodeDeviceType{ID: dt.DeviceType, Name: deviceTypeName(dt.DeviceType), Revision: dt.Revision})
		}
		return out
	}
	if dt := ep.DeviceType(); dt != 0 {
		t := fabricui.NodeDeviceType{ID: dt}
		if d := ep.Device(); d != nil {
			t.Name = d.Name
		}
		if rev, ok := ep.DeviceRevision(); ok {
			t.Revision = rev
		}
		out = append(out, t)
	}
	return out
}

// deviceTypeName is the schema's name of a device type; empty for a
// manufacturer-specific or unknown one.
func deviceTypeName(id uint32) string {
	if dt, ok := schema.DeviceTypeIDOf(id); ok {
		if name, ok := schema.DeviceTypeName(dt); ok {
			return name
		}
	}
	return ""
}

// clusterName is a cluster's name: its definition's, else the schema's;
// empty for one neither names.
func clusterName(id uint32, def *spec.Cluster) string {
	if def != nil {
		return def.Name
	}
	name, _ := schema.ClusterName(id)
	return name
}

// globalAttributeNames are the global attributes' names (Matter Core
// §7.13), which no generated definition lists.
var globalAttributeNames = map[uint32]string{
	spec.AttrGeneratedCommandList: "GeneratedCommandList",
	spec.AttrAcceptedCommandList:  "AcceptedCommandList",
	spec.AttrEventList:            "EventList",
	spec.AttrAttributeList:        "AttributeList",
	spec.AttrFeatureMap:           "FeatureMap",
	spec.AttrClusterRevision:      "ClusterRevision",
}

// EndpointDetail is one endpoint of the image with its inventory, read
// from the controller's cache only (owner decision U4 0.1): the
// Descriptor's DeviceTypeList and TagList, FixedLabel and UserLabel,
// BridgedDeviceBasicInformation Reachable, the parts on the image's tree,
// the device type conformance (client.Endpoint.Conformance) and every
// ready cluster without its values.
func (w *node) EndpointDetail(number uint16) (fabricui.EndpointDetail, error) {
	ep := w.n.Endpoint(number)
	if ep == nil {
		return fabricui.EndpointDetail{}, fmt.Errorf("%w: %d", fabricui.ErrEndpointNotFound, number)
	}
	d := fabricui.EndpointDetail{
		Number: number, Parent: parentOf(ep), Name: ep.Name(), Reachable: true, DeviceTypes: deviceTypes(ep),
		Tags: tagsOf(ep), Labels: labelsOf(ep), Parts: []uint16{}, Clusters: []fabricui.ClusterDetail{},
	}
	endpointValue(ep, clusterBridgedBasicInfo, attrReachable, &d.Reachable)
	for _, p := range ep.Parts() {
		d.Parts = append(d.Parts, p.Number())
	}
	for _, c := range ep.Clusters() {
		d.Clusters = append(d.Clusters, clusterDetail(c, false))
	}
	d.Conformance = conformanceOf(ep.Conformance())
	return d, nil
}

// ClusterDetail is one ready cluster of an endpoint of the image with the
// values the controller holds, rendered by name as Read renders them.
func (w *node) ClusterDetail(number uint16, cluster uint32) (fabricui.ClusterDetail, error) {
	ep := w.n.Endpoint(number)
	if ep == nil {
		return fabricui.ClusterDetail{}, fmt.Errorf("%w: %d", fabricui.ErrEndpointNotFound, number)
	}
	c := ep.Cluster(cluster)
	if c == nil {
		// Until the Descriptor is read the image cannot tell a cluster
		// the endpoint lacks from one not read yet.
		if ep.Cluster(clusterDescriptor) == nil {
			return fabricui.ClusterDetail{}, fmt.Errorf("%w: the Descriptor of endpoint %d", fabricui.ErrNotRead, number)
		}
		var servers []uint32
		if endpointValue(ep, clusterDescriptor, attrServerList, &servers) && slices.Contains(servers, cluster) {
			return fabricui.ClusterDetail{}, fmt.Errorf("%w: cluster 0x%04X on endpoint %d", fabricui.ErrNotRead, cluster, number)
		}
		return fabricui.ClusterDetail{}, fmt.Errorf("%w: 0x%04X on endpoint %d", fabricui.ErrClusterNotFound, cluster, number)
	}
	return clusterDetail(c, true), nil
}

// tagsOf is the endpoint's Descriptor TagList; empty without one.
func tagsOf(ep *client.Endpoint) []fabricui.SemanticTag {
	var list []struct {
		MfgCode     *uint16 `json:"mfgCode"`
		NamespaceID uint8   `json:"namespaceId"`
		Tag         uint8   `json:"tag"`
		Label       *string `json:"label"`
	}
	endpointValue(ep, clusterDescriptor, attrTagList, &list)
	out := make([]fabricui.SemanticTag, 0, len(list))
	for _, t := range list {
		st := fabricui.SemanticTag{MfgCode: t.MfgCode, Namespace: t.NamespaceID, Tag: t.Tag}
		if t.Label != nil {
			st.Label = *t.Label
		}
		out = append(out, st)
	}
	return out
}

// labelsOf is the endpoint's FixedLabel LabelList, then its UserLabel
// LabelList.
func labelsOf(ep *client.Endpoint) []fabricui.EndpointLabel {
	return append(labelList(ep, clusterFixedLabel, false), labelList(ep, clusterUserLabel, true)...)
}

// labelList is the LabelList of a FixedLabel or UserLabel cluster.
func labelList(ep *client.Endpoint, cluster uint32, user bool) []fabricui.EndpointLabel {
	var list []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	endpointValue(ep, cluster, attrLabelList, &list)
	out := make([]fabricui.EndpointLabel, 0, len(list))
	for _, l := range list {
		out = append(out, fabricui.EndpointLabel{Label: l.Label, Value: l.Value, User: user})
	}
	return out
}

// conformanceOf is the port's form of client.ConformanceReport; its
// informational client findings stay out.
func conformanceOf(r client.ConformanceReport) *fabricui.Conformance {
	out := &fabricui.Conformance{Status: r.Status.String(), Violations: []fabricui.ConformanceViolation{}, Unjudged: []fabricui.ClusterElement{}}
	for i := range r.Violations {
		v := &r.Violations[i]
		out.Violations = append(out.Violations, fabricui.ConformanceViolation{
			DeviceType: v.DeviceType, DeviceTypeID: v.DeviceTypeID, Requirement: v.Requirement, Kind: string(v.Kind), Detail: v.Detail,
		})
	}
	for _, id := range r.Unjudged {
		out.Unjudged = append(out.Unjudged, fabricui.ClusterElement{ID: id, Name: clusterName(id, nil)})
	}
	return out
}

// clusterDetail is a ready cluster's inventory from the image: revision,
// FeatureMap with its features (the codes of client.Cluster.Features, the
// bit from the schema, the title from the definition), DataVersion,
// AttributeList — with the values when values is set — the command lists
// and the EventList where the peer reports one, each named by the
// definition where it has one.
func clusterDetail(c *client.Cluster, values bool) fabricui.ClusterDetail {
	def := c.Definition()
	d := fabricui.ClusterDetail{
		ID: c.ID(), Name: clusterName(c.ID(), def), Revision: c.Revision(), FeatureMap: c.FeatureMap(),
		Features: []fabricui.ClusterFeature{}, Attributes: []fabricui.ClusterAttribute{},
		Accepted: commands(def, c.AcceptedCommands(), spec.Request), Generated: commands(def, c.GeneratedCommands(), spec.Response),
		Events: []fabricui.ClusterElement{},
	}
	if v, ok := c.Version(); ok {
		d.DataVersion = &v
	}
	known := schema.ClusterFeatures(c.ID())
	for _, code := range c.Features() {
		f := fabricui.ClusterFeature{Code: code}
		if i := slices.IndexFunc(known, func(k schema.ClusterFeature) bool { return k.Name == code }); i >= 0 {
			f.Bit = known[i].Bit
		}
		if def != nil {
			if df := def.Feature(code); df != nil {
				f.Title = df.Title
			}
		}
		d.Features = append(d.Features, f)
	}
	for _, id := range c.AttributeList() {
		a := fabricui.ClusterAttribute{ID: id, Name: globalAttributeNames[id]}
		if def != nil {
			if da := def.Attribute(id); da != nil {
				a.Name = da.Name
			}
		}
		if values {
			if v, ok := c.Value(id); ok {
				// A value that does not render is left out, not null.
				if raw, err := client.DecodeJSON(client.AttributeField(c.ID(), id), v, client.NameBased); err == nil {
					a.Value = raw
				}
			}
		}
		d.Attributes = append(d.Attributes, a)
	}
	var events []uint32
	if v, ok := c.Value(spec.AttrEventList); ok {
		if raw, err := client.DecodeJSON(client.AttributeField(c.ID(), spec.AttrEventList), v, client.NameBased); err == nil {
			_ = json.Unmarshal(raw, &events)
		}
	}
	for _, id := range events {
		e := fabricui.ClusterElement{ID: id}
		if def != nil {
			if de := def.Event(id); de != nil {
				e.Name = de.Name
			}
		}
		d.Events = append(d.Events, e)
	}
	return d
}

// commands names a command list by the definition's commands of the
// direction.
func commands(def *spec.Cluster, ids []uint32, dir spec.Direction) []fabricui.ClusterElement {
	out := make([]fabricui.ClusterElement, 0, len(ids))
	for _, id := range ids {
		e := fabricui.ClusterElement{ID: id}
		if def != nil {
			if dc := def.Command(id, dir); dc != nil {
				e.Name = dc.Name
			}
		}
		out = append(out, e)
	}
	return out
}

// Events are the events the controller received from the node, from its
// event history (client.Node.EventHistory, go-fabric
// controller/client/event_history.go:100-115 at 894adbcd), each named by
// the schema (schema.EventName) and its data rendered by name as Read
// renders a value (client.DecodeJSON with client.EventField); data that
// does not render is left out.
func (w *node) Events(since uint64, limit int) []fabricui.EventRecord {
	hist := w.n.EventHistory(since, limit)
	out := make([]fabricui.EventRecord, 0, len(hist))
	for i := range hist {
		e := &hist[i]
		p := e.Path
		r := fabricui.EventRecord{
			Number: e.Number, Priority: fabricui.EventPriorityName(uint8(e.Priority)),
			Endpoint: p.Endpoint, Cluster: p.Cluster, Event: p.Event,
		}
		r.Name, _ = schema.EventName(p.Cluster, p.Event)
		if e.EpochTimestamp != nil {
			ms := int64(*e.EpochTimestamp) //nolint:gosec // an epoch in milliseconds fits int64
			r.EpochMS = &ms
		}
		if e.HasValue {
			if raw, err := client.DecodeJSON(client.EventField(p.Cluster, p.Event), e.Value, client.NameBased); err == nil {
				r.Data = raw
			}
		}
		out = append(out, r)
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

// pingTimeout bounds a ping, as go-fabric-server's ping_node bounds each
// attempt (controllerapi/cmd_node.go:187): the session read of the
// fallback, or the whole round of per-address probes.
const pingTimeout = 10 * time.Second

// pingParallel bounds how many addresses of one node are probed at once.
const pingParallel = 4

// Ping probes every address of the node's record on its own with
// client.Node.PingAddress (go-fabric controller/client/ping.go at
// 8feaa508, issue #398): one empty read over the node's held session sent
// to that address — matter.js ClientInteraction.probe with
// addressOverride — measured on the node's clock; nothing migrates, and a
// silent address is an outcome, not an error (PerAddress true). The
// result is reachable when one address answered, with the fastest round
// trip.
//
// A node without addresses or without a session to probe on
// (client.ErrNoSession, ErrNotCommissioned, ErrClosed) keeps the single
// outcome: a read of BasicInformation DataModelRevision (0x0028/0x0000)
// through the session without data version filters — go-fabric-server's
// ping_node (controllerapi/cmd_node.go:167-204 at e8bcd86), one attempt —
// standing for every address (PerAddress false). The end of the request's
// ctx is its error.
func (w *node) Ping(ctx context.Context) (fabricui.PingResult, error) {
	rec := w.n.Record().Addresses
	addrs := make([]netip.AddrPort, 0, len(rec))
	for _, a := range rec {
		addrs = append(addrs, netip.AddrPortFrom(a.IP, a.Port))
	}
	rctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if res, ok := pingAddresses(rctx, addrs, w.n.PingAddress); ok {
		if err := ctx.Err(); err != nil {
			return fabricui.PingResult{}, err
		}
		return res, nil
	}
	return w.pingSession(rctx)
}

// pingAddresses probes each address with ping, at most pingParallel at
// once, and keeps the record's order. It reports false when no address is
// known or a probe failed for the node rather than for its address — no
// session to probe on, a node not commissioned or closed — so the caller
// falls back to the session read.
func pingAddresses(ctx context.Context, addrs []netip.AddrPort, ping func(context.Context, netip.AddrPort) (time.Duration, error)) (fabricui.PingResult, bool) {
	if len(addrs) == 0 {
		return fabricui.PingResult{}, false
	}
	type outcome struct {
		rtt time.Duration
		err error
	}
	outs := make([]outcome, len(addrs))
	sem := make(chan struct{}, pingParallel)
	var wg sync.WaitGroup
	for i, a := range addrs {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			rtt, err := ping(ctx, a)
			outs[i] = outcome{rtt, err}
		})
	}
	wg.Wait()
	res := fabricui.PingResult{PerAddress: true, Addresses: make([]fabricui.PingAddress, 0, len(addrs))}
	for i, addr := range addrs {
		o := outs[i]
		if errors.Is(o.err, client.ErrNoSession) || errors.Is(o.err, client.ErrNotCommissioned) || errors.Is(o.err, client.ErrClosed) {
			return fabricui.PingResult{}, false
		}
		pa := fabricui.PingAddress{Address: addr.String()}
		if o.err != nil {
			pa.Error = o.err.Error()
		} else {
			pa.Reachable, pa.RTT = true, o.rtt
			if !res.Reachable || o.rtt < res.RTT {
				res.Reachable, res.RTT = true, o.rtt
			}
		}
		res.Addresses = append(res.Addresses, pa)
	}
	if !res.Reachable {
		res.Error = fmt.Sprintf("none of the node's %d addresses answered", len(addrs))
	}
	return res, true
}

// pingSession is the single-outcome ping: one read through the node's
// session stands for every address the record lists.
func (w *node) pingSession(ctx context.Context) (fabricui.PingResult, error) {
	req := client.ReadRequest{IncludeKnownVersions: true}
	req.Attributes = []client.AttributePath{client.AttributeAt(0, basicinformation.ClusterID, basicinformation.AttrDataModelRevision)}
	start := time.Now()
	_, err := w.n.Read(ctx, req)
	rtt := time.Since(start)
	// An answer is reachable, a status in place of the value included, as
	// ping_node counts it (cmd_node.go:190-193: err == nil).
	res := fabricui.PingResult{Addresses: []fabricui.PingAddress{}}
	if err != nil {
		res.Error = err.Error()
	} else {
		res.Reachable, res.RTT = true, rtt
	}
	for _, a := range w.Record().Addresses {
		pa := fabricui.PingAddress{Address: a, Reachable: res.Reachable, RTT: res.RTT}
		if err != nil {
			pa.Error = res.Error
		}
		res.Addresses = append(res.Addresses, pa)
	}
	return res, nil
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
