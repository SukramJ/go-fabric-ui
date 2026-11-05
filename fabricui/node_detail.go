// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/SukramJ/go-fabric/contract"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/schema"
)

// The Matter 1.6.1 ids the node's endpoint detail reads: Descriptor
// (0x001D) ServerList (0x0001) and TagList (0x0004); FixedLabel (0x0040) and UserLabel
// (0x0041) LabelList (0x0000); the global attributes (Matter Core §7.13)
// GeneratedCommandList (0xFFF8), AcceptedCommandList (0xFFF9), EventList
// (0xFFFA), AttributeList (0xFFFB), FeatureMap (0xFFFC) and
// ClusterRevision (0xFFFD). The import rule keeps the generated cluster
// definitions that name them out of this package (internal/arch).
const (
	clusterDescriptor     uint32 = 0x001D
	attrServerList        uint32 = 0x0001
	attrTagList           uint32 = 0x0004
	clusterFixedLabel     uint32 = 0x0040
	clusterUserLabel      uint32 = 0x0041
	attrLabelList         uint32 = 0x0000
	attrGeneratedCommands uint32 = 0xFFF8
	attrAcceptedCommands  uint32 = 0xFFF9
	attrEventList         uint32 = 0xFFFA
	attrAttributeList     uint32 = 0xFFFB
	attrFeatureMap        uint32 = 0xFFFC
	attrClusterRevision   uint32 = 0xFFFD
)

// statusUnsupportedEndpoint is the IM status UNSUPPORTED_ENDPOINT (Matter
// Core §8.10.1, 0x7F), the contract.StatusError Node.ReadCluster returns
// for an endpoint the node does not serve (go-fabric node/served.go:70-94
// at 894adbcd).
const statusUnsupportedEndpoint uint8 = 0x7F

// readEndpointDetail builds an endpoint's detail from what the node
// serves: the Descriptor's TagList and ServerList and every server
// cluster's values through Node.ReadCluster (go-fabric
// node/served.go:20-40 and :70-94 at 82591648), a failed attribute with
// its status; FixedLabel and UserLabel from their values; the place in the
// tree, the DeviceTypeList, the name and the scope from the endpoint's
// node.EndpointInfo (Node.Endpoints, node/endpoints.go:14-66 at 82591648)
// — the root and the Aggregator from the same source as a bridged
// endpoint — and Reachable from the topology (Node.Topology), true for an
// endpoint the topology does not list. ErrEndpointNotFound for an
// endpoint the node does not serve, node.ErrNotStarted before the node
// started.
func readEndpointDetail(ctx context.Context, n NodeSource, ep uint16) (EndpointDetail, error) {
	desc, err := n.ReadCluster(ctx, ep, clusterDescriptor)
	if err != nil {
		var se contract.StatusError
		if errors.As(err, &se) && se.Status == statusUnsupportedEndpoint {
			return EndpointDetail{}, fmt.Errorf("%w: %d", ErrEndpointNotFound, ep)
		}
		return EndpointDetail{}, err
	}
	infos := n.Endpoints()
	i := slices.IndexFunc(infos, func(e node.EndpointInfo) bool { return e.Endpoint == ep })
	if i < 0 {
		// Served at the Descriptor read, gone from the tree since: the
		// node reassembled in between.
		return EndpointDetail{}, fmt.Errorf("%w: %d", ErrEndpointNotFound, ep)
	}
	info := &infos[i]
	d := EndpointDetail{
		Number: ep, Parent: treeParent(info), Name: info.Name, Reachable: true, Scope: info.Scope,
		DeviceTypes: deviceTypesOf(info.DeviceTypes), Tags: tagsOf(desc),
		Labels: []EndpointLabel{}, Parts: childrenOf(infos, ep), Clusters: []ClusterDetail{},
	}
	if d.Name == "" && len(d.DeviceTypes) > 0 {
		d.Name = d.DeviceTypes[0].Name
	}
	if dev, _ := findDevice(n.Topology(), ep); dev != nil {
		d.Reachable = dev.Reachable
	}
	var servers []uint32
	valueOf(desc, attrServerList, &servers)
	slices.Sort(servers)
	for _, id := range servers {
		vals := desc
		if id != clusterDescriptor {
			// A cluster whose read fails is listed with its id and the
			// schema's name, revision and features, without values.
			vals, _ = n.ReadCluster(ctx, ep, id)
		}
		d.Clusters = append(d.Clusters, servedCluster(id, vals))
		switch id {
		case clusterFixedLabel:
			d.Labels = append(d.Labels, labelsOf(vals, false)...)
		case clusterUserLabel:
			d.Labels = append(d.Labels, labelsOf(vals, true)...)
		}
	}
	// FixedLabel sorts before UserLabel, so the labels are in the
	// contract's order already.
	return d, nil
}

// treeParent is the contract's parent of an endpoint: its parent in the
// node's endpoint tree (EndpointInfo.Parent where HasParent is set,
// node/endpoints.go:19-24 at 82591648), absent for the root and for a
// top-level endpoint — one whose parent is the root, the Aggregator — as
// the controller family reads it (gofabric parentOf).
func treeParent(info *node.EndpointInfo) *uint16 {
	if !info.HasParent || info.Parent == 0 {
		return nil
	}
	p := info.Parent
	return &p
}

// childrenOf are the direct children of endpoint ep — the endpoints whose
// tree parent it is — in endpoint order (Node.Endpoints lists the root
// first, then by number). Not the PartsList, which names every descendant
// on an endpoint that serves the full-family list.
func childrenOf(infos []node.EndpointInfo, ep uint16) []uint16 {
	out := []uint16{}
	for i := range infos {
		if infos[i].HasParent && infos[i].Parent == ep {
			out = append(out, infos[i].Endpoint)
		}
	}
	return out
}

// deviceTypesOf is a DeviceTypeList in its order, each named by the schema.
func deviceTypesOf(dts []node.DeviceTypeInfo) []NodeDeviceType {
	out := make([]NodeDeviceType, 0, len(dts))
	for _, dt := range dts {
		out = append(out, NodeDeviceType{ID: dt.ID, Name: schemaDeviceTypeName(dt.ID), Revision: dt.Revision})
	}
	return out
}

// findDevice is the bridged device or part at endpoint ep with its parent
// (the Aggregator for a top-level device); nil when the topology does not
// list ep.
func findDevice(topo node.BridgeTopology, ep uint16) (dev *node.BridgedDevice, parent *uint16) {
	var walk func(p uint16, devs []node.BridgedDevice) bool
	walk = func(p uint16, devs []node.BridgedDevice) bool {
		for i := range devs {
			if devs[i].Endpoint == ep {
				dev, parent = &devs[i], &p
				return true
			}
			if walk(devs[i].Endpoint, devs[i].Parts) {
				return true
			}
		}
		return false
	}
	walk(topo.Aggregator, topo.Devices)
	return dev, parent
}

// servedCluster is a cluster's inventory from the values the node serves:
// ClusterRevision, FeatureMap with the features the schema names for its
// set bits, the DataVersion, the AttributeList in id order with each
// value rendered by name (node.AttributeValue.JSON) — absent where the
// read left the attribute out or the value does not render — and the
// command and event lists, each named by go-fabric's schema
// (schema.AttributeName, CommandName, ResponseCommandName, EventName,
// ClusterFeatures; schema/lookup.go at 894adbcd). An attribute whose read
// failed (AttributeValue.Status, go-fabric node/served.go:20-40 at
// 82591648) is listed with its status and no value, the AttributeList's or
// not. Without a value (a read that failed as a whole, or every attribute
// failed) it is the schema's revision and the cluster's name.
func servedCluster(id uint32, vals []node.AttributeValue) ClusterDetail {
	c := ClusterDetail{
		ID: id, Features: []ClusterFeature{}, Attributes: []ClusterAttribute{},
		Accepted: []ClusterElement{}, Generated: []ClusterElement{}, Events: []ClusterElement{},
	}
	c.Name, _ = schema.ClusterName(id)
	first := slices.IndexFunc(vals, func(v node.AttributeValue) bool { return v.Status == 0 })
	if first < 0 {
		c.Revision, _ = schema.ClusterRevision(id)
		c.Attributes = failedAttributes(id, vals)
		return c
	}
	// A failed attribute carries DataVersion 0: the version is a value's.
	v := vals[first].DataVersion
	c.DataVersion = &v
	valueOf(vals, attrClusterRevision, &c.Revision)
	valueOf(vals, attrFeatureMap, &c.FeatureMap)
	for _, f := range schema.ClusterFeatures(id) {
		if c.FeatureMap&(1<<f.Bit) != 0 {
			c.Features = append(c.Features, ClusterFeature{Code: f.Name, Title: f.Title, Bit: f.Bit})
		}
	}
	var attrs, accepted, generated, events []uint32
	valueOf(vals, attrAttributeList, &attrs)
	valueOf(vals, attrAcceptedCommands, &accepted)
	valueOf(vals, attrGeneratedCommands, &generated)
	valueOf(vals, attrEventList, &events)
	for _, v := range vals {
		if v.Status != 0 && !slices.Contains(attrs, v.Attribute) {
			attrs = append(attrs, v.Attribute)
		}
	}
	slices.Sort(attrs)
	for _, a := range attrs {
		e := ClusterAttribute{ID: a}
		e.Name, _ = schema.AttributeName(id, a)
		if i := slices.IndexFunc(vals, func(v node.AttributeValue) bool { return v.Attribute == a }); i >= 0 {
			if vals[i].Status != 0 {
				e.Status = vals[i].Status
			} else if raw, err := vals[i].JSON(); err == nil {
				e.Value = raw
			}
		}
		c.Attributes = append(c.Attributes, e)
	}
	c.Accepted = named(id, accepted, schema.CommandName)
	c.Generated = named(id, generated, schema.ResponseCommandName)
	c.Events = named(id, events, schema.EventName)
	return c
}

// failedAttributes are the attributes of vals whose read failed, in id
// order, each with its status.
func failedAttributes(cluster uint32, vals []node.AttributeValue) []ClusterAttribute {
	out := []ClusterAttribute{}
	for _, v := range vals {
		if v.Status != 0 {
			e := ClusterAttribute{ID: v.Attribute, Status: v.Status}
			e.Name, _ = schema.AttributeName(cluster, v.Attribute)
			out = append(out, e)
		}
	}
	slices.SortFunc(out, func(a, b ClusterAttribute) int { return int(a.ID) - int(b.ID) })
	return out
}

// named lists ids with the name lookup gives each.
func named(cluster uint32, ids []uint32, lookup func(cluster, id uint32) (string, bool)) []ClusterElement {
	out := make([]ClusterElement, 0, len(ids))
	for _, id := range ids {
		name, _ := lookup(cluster, id)
		out = append(out, ClusterElement{ID: id, Name: name})
	}
	return out
}

// tagsOf is the Descriptor's TagList; empty while it serves none.
func tagsOf(desc []node.AttributeValue) []SemanticTag {
	var list []struct {
		MfgCode     *uint16 `json:"mfgCode"`
		NamespaceID uint8   `json:"namespaceId"`
		Tag         uint8   `json:"tag"`
		Label       *string `json:"label"`
	}
	valueOf(desc, attrTagList, &list)
	out := make([]SemanticTag, 0, len(list))
	for _, t := range list {
		st := SemanticTag{MfgCode: t.MfgCode, Namespace: t.NamespaceID, Tag: t.Tag}
		if t.Label != nil {
			st.Label = *t.Label
		}
		out = append(out, st)
	}
	return out
}

// labelsOf is a FixedLabel or UserLabel cluster's LabelList.
func labelsOf(vals []node.AttributeValue, user bool) []EndpointLabel {
	var list []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	valueOf(vals, attrLabelList, &list)
	out := make([]EndpointLabel, 0, len(list))
	for _, l := range list {
		out = append(out, EndpointLabel{Label: l.Label, Value: l.Value, User: user})
	}
	return out
}

// valueOf decodes the attribute attr of vals, rendered by name, into v;
// v stays as it is when vals lacks it, its read failed or it does not fit.
func valueOf(vals []node.AttributeValue, attr uint32, v any) {
	if i := slices.IndexFunc(vals, func(x node.AttributeValue) bool { return x.Attribute == attr && x.Status == 0 }); i >= 0 {
		jsonOf(vals[i], v)
	}
}

func jsonOf(av node.AttributeValue, v any) bool {
	raw, err := av.JSON()
	return err == nil && json.Unmarshal(raw, v) == nil
}

// schemaDeviceTypeName is the schema's name of a device type; empty for a
// manufacturer-specific or unknown one, as the controller family names it.
func schemaDeviceTypeName(id uint32) string {
	if dt, ok := schema.DeviceTypeIDOf(id); ok {
		if name, ok := schema.DeviceTypeName(dt); ok {
			return name
		}
	}
	return ""
}

// eventRecordOf is the port's EventRecord of an event the node served: its
// name from the schema, its time as epoch milliseconds, its payload
// rendered by name (node.Event.JSON, go-fabric node/served.go:140-144 at
// 894adbcd) — absent without one or when it does not render.
func eventRecordOf(e *node.Event) EventRecord {
	r := EventRecord{
		Number: e.Number, Priority: EventPriorityName(uint8(e.Priority)),
		Endpoint: e.Endpoint, Cluster: e.Cluster, Event: e.Event,
	}
	r.Name, _ = schema.EventName(e.Cluster, e.Event)
	if ms := e.Time.UnixMilli(); ms > 0 {
		r.EpochMS = &ms
	}
	if e.Payload != nil {
		if raw, err := e.JSON(); err == nil {
			r.Data = raw
		}
	}
	return r
}
