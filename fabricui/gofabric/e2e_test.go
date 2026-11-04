// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/attestation"
	"github.com/SukramJ/go-fabric/controller/topology"
	"github.com/SukramJ/go-fabric/device"
	"github.com/SukramJ/go-fabric/fabrictest"
	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/fabricui"
	"github.com/SukramJ/go-fabric-ui/fabricui/gofabric"
)

// ctrlSite is a go-fabric controller (controller.New on a fabrictest site)
// behind the controller family's handlers, next to one device. Every
// HTTP call runs under Site.Run, which moves the site's fake clock while
// the controller waits on the network.
type ctrlSite struct {
	t      *testing.T
	site   *fabrictest.Site
	dev    *fabrictest.Device
	plug   *device.OnOffPlugInUnit
	ctrl   *controller.Controller
	srv    *httptest.Server
	token  string
	events chan string
}

func newCtrlSite(t *testing.T) *ctrlSite {
	t.Helper()
	site := fabrictest.NewSite(t, fabrictest.Options{})
	plug, err := device.NewOnOffPlugInUnit(device.Options{StableKey: device.StringKey("plug:1"), Name: "Plug"})
	if err != nil {
		t.Fatal(err)
	}
	dev := site.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{plug}})
	hooks := gofabric.NewHooks()
	cfg := site.ControllerConfig(fabrictest.ControllerOptions{Label: "e2e"})
	cfg.HooksFor = hooks.Wrap(cfg.HooksFor)
	// The network graph is the host's service: its changes are installed
	// before controller.New, the service built after it and closed first.
	var changes topology.Changes
	changes.Install(&cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var ctrl *controller.Controller
	if err := site.Run(ctx, func(ctx context.Context) (err error) {
		ctrl, err = controller.New(ctx, cfg)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	svc := topology.New(ctrl, topology.Config{Changes: &changes, Clock: cfg.Clock})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		svc.Close()
		hooks.Close() // before the controller goes: no timer outlives it
		_ = site.Run(ctx, ctrl.Close)
	})
	src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{
		Hooks:       hooks,
		Topology:    svc,
		Credentials: fabricui.NewMemoryCredentials(gofabric.ThreadDatasetName),
		// The site's device carries the module's test attestation.
		Pair: controller.PairOptions{OnAttestationFailure: attestation.Accept()},
	})
	s := &ctrlSite{t: t, site: site, dev: dev, plug: plug, ctrl: ctrl, events: make(chan string, 256)}
	// The node family beside the controller family: the device's node
	// is the node the controller commissions.
	s.srv = httptest.NewServer(fabricui.Handler(fabricui.Options{Controller: src, Node: dev.Node()}))
	t.Cleanup(s.srv.Close)
	src.Observe(func(e fabricui.ControllerEvent) {
		select {
		case s.events <- e.Type:
		default:
		}
	})
	var st struct {
		Enabled   bool   `json:"enabled"`
		Label     string `json:"label"`
		CSRFToken string `json:"csrf_token"`
	}
	s.json(http.MethodGet, "/controller/status", "", http.StatusOK, &st)
	if !st.Enabled || st.Label != "e2e" || st.CSRFToken == "" {
		t.Fatalf("status %+v", st)
	}
	s.token = st.CSRFToken
	return s
}

// call sends one request under Site.Run.
func (s *ctrlSite) call(method, path, body string) (status int, out []byte) {
	s.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := s.site.Run(ctx, func(ctx context.Context) error {
		var rd io.Reader
		if body != "" {
			rd = strings.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, s.srv.URL+fabricui.DefaultPrefix+path, rd)
		if err != nil {
			return err
		}
		req.Header.Set(fabricui.TokenHeader, s.token)
		resp, err := s.srv.Client().Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		status = resp.StatusCode
		out, err = io.ReadAll(resp.Body)
		return err
	})
	if err != nil {
		s.t.Fatalf("%s %s: %v", method, path, err)
	}
	return status, out
}

// json sends one request, checks the status and decodes the answer into v.
func (s *ctrlSite) json(method, path, body string, want int, v any) {
	s.t.Helper()
	status, out := s.call(method, path, body)
	if status != want {
		s.t.Fatalf("%s %s: %d, want %d: %s", method, path, status, want, out)
	}
	if v != nil {
		if err := json.Unmarshal(out, v); err != nil {
			s.t.Fatalf("%s %s: %v: %s", method, path, err, out)
		}
	}
}

// waitEvent waits for an event of the type from the source.
func (s *ctrlSite) waitEvent(typ string) {
	s.t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		select {
		case got := <-s.events:
			if got == typ {
				return
			}
		case <-deadline:
			s.t.Fatalf("no %s event", typ)
		}
	}
}

type nodeDetail struct {
	Node struct {
		NodeIDHex  string `json:"node_id_hex"`
		Connection string `json:"connection"`
		VendorID   int    `json:"vendor_id"`
	} `json:"node"`
	Record struct {
		FabricIndex int      `json:"fabric_index"`
		Addresses   []string `json:"addresses"`
	} `json:"record"`
	Endpoints []struct {
		Endpoint  int   `json:"endpoint"`
		Reachable *bool `json:"reachable"`
		Clusters  []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"clusters"`
	} `json:"endpoints"`
}

// TestEndToEnd_Controller: the controller family on a real go-fabric
// controller — commission a fabrictest device by its manual code, list
// and read the node, open and close a window, write and read back the
// ACL, read an attribute, list the node's fabrics, decommission.
func TestEndToEnd_Controller(t *testing.T) {
	s := newCtrlSite(t)
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}

	var added nodeDetail
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.Manual), http.StatusOK, &added)
	id := added.Node.NodeIDHex
	if len(id) != 16 || added.Record.FabricIndex == 0 {
		t.Fatalf("commissioned %+v", added)
	}
	s.waitEvent(fabricui.EventNodeAdded)

	// The network graph: the controller and the node, joined by a session.
	type edge struct {
		Source string `json:"source"`
		Target string `json:"target"`
		Kind   string `json:"kind"`
	}
	var graph struct {
		Family string `json:"family"`
		Nodes  []struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
		} `json:"nodes"`
		Edges []edge `json:"edges"`
	}
	s.json(http.MethodGet, "/controller/topology", "", http.StatusOK, &graph)
	peer := "peer:" + id
	if graph.Family != "controller" || len(graph.Nodes) != 2 || graph.Nodes[0].ID != "node" ||
		graph.Nodes[1].ID != peer || graph.Nodes[1].Kind != "peer" ||
		!slices.Contains(graph.Edges, edge{Source: "node", Target: peer, Kind: "session"}) {
		t.Fatalf("graph %+v", graph)
	}

	var list struct {
		Nodes []struct {
			NodeIDHex string `json:"node_id_hex"`
		} `json:"nodes"`
	}
	s.json(http.MethodGet, "/controller/nodes", "", http.StatusOK, &list)
	if len(list.Nodes) != 1 || list.Nodes[0].NodeIDHex != id {
		t.Fatalf("nodes %+v", list)
	}

	var detail nodeDetail
	s.json(http.MethodGet, "/controller/nodes/"+id, "", http.StatusOK, &detail)
	onOff := -1
	for _, ep := range detail.Endpoints {
		for _, c := range ep.Clusters {
			if c.ID == 0x0006 && c.Name == "OnOff" {
				onOff = ep.Endpoint
			}
		}
	}
	if len(detail.Endpoints) < 2 || onOff < 1 {
		t.Fatalf("detail endpoints %+v", detail.Endpoints)
	}
	for _, ep := range detail.Endpoints {
		if ep.Endpoint == onOff && (ep.Reachable == nil || !*ep.Reachable) {
			t.Fatalf("the bridged plug's endpoint is not reachable: %s", mustJSON(t, ep))
		}
	}

	// The ping: a Matter read of DataModelRevision through the session
	// answers, and the one outcome stands for every known address.
	var ping struct {
		Method     string   `json:"method"`
		PerAddress bool     `json:"per_address"`
		Reachable  bool     `json:"reachable"`
		RttMs      *float64 `json:"rtt_ms"`
		Addresses  []struct {
			IP        string `json:"ip"`
			Reachable bool   `json:"reachable"`
		} `json:"addresses"`
	}
	s.json(http.MethodPost, "/controller/nodes/"+id+"/ping", "", http.StatusOK, &ping)
	if ping.Method != "matter-read" || ping.PerAddress || !ping.Reachable || ping.RttMs == nil || len(ping.Addresses) == 0 {
		t.Fatalf("ping %+v", ping)
	}
	for _, a := range ping.Addresses {
		if !a.Reachable {
			t.Fatalf("ping address %+v", a)
		}
	}

	// The window for another administrator: codes and the passcode.
	var w struct {
		ManualCode     string `json:"manual_code"`
		QRCode         string `json:"qr_code"`
		Passcode       int    `json:"passcode"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	s.json(http.MethodPost, "/controller/nodes/"+id+"/window", `{"timeout_seconds":300}`, http.StatusOK, &w)
	if w.ManualCode == "" || !strings.HasPrefix(w.QRCode, "MT:") || w.Passcode == 0 || w.TimeoutSeconds != 300 {
		t.Fatalf("window %+v", w)
	}
	s.json(http.MethodPost, "/controller/nodes/"+id+"/window/close", "", http.StatusNoContent, nil)

	// The ACL: ours, then ours and an operator on the plug's OnOff.
	type acl struct {
		Entries []struct {
			Privilege   string   `json:"privilege"`
			AuthMode    string   `json:"auth_mode"`
			SubjectsHex []string `json:"subjects_hex"`
			Targets     []struct {
				ClusterHex *string `json:"cluster_hex"`
				Endpoint   *int    `json:"endpoint"`
			} `json:"targets"`
			FabricIndex int `json:"fabric_index"`
		} `json:"entries"`
	}
	var before acl
	s.json(http.MethodGet, "/controller/nodes/"+id+"/acl", "", http.StatusOK, &before)
	if len(before.Entries) != 1 || before.Entries[0].Privilege != "administer" || before.Entries[0].AuthMode != "case" {
		t.Fatalf("acl before %+v", before)
	}
	own := before.Entries[0].SubjectsHex
	body := fmt.Sprintf(`{"entries":[{"privilege":"administer","auth_mode":"case","subjects_hex":%s,"targets":[]},`+
		`{"privilege":"operate","auth_mode":"case","subjects_hex":["0000000000000099"],"targets":[{"cluster_hex":"0x0006","endpoint":%d}]}]}`,
		mustJSON(t, own), onOff)
	var wr struct {
		Status int `json:"status"`
	}
	s.json(http.MethodPut, "/controller/nodes/"+id+"/acl", body, http.StatusOK, &wr)
	if wr.Status != 0 {
		t.Fatalf("acl write status %#x", wr.Status)
	}
	var after acl
	s.json(http.MethodGet, "/controller/nodes/"+id+"/acl", "", http.StatusOK, &after)
	if len(after.Entries) != 2 || after.Entries[1].Privilege != "operate" || !slices.Equal(after.Entries[1].SubjectsHex, []string{"0000000000000099"}) ||
		len(after.Entries[1].Targets) != 1 || *after.Entries[1].Targets[0].ClusterHex != "0x0006" || *after.Entries[1].Targets[0].Endpoint != onOff ||
		after.Entries[1].FabricIndex != before.Entries[0].FabricIndex {
		t.Fatalf("acl after %s", mustJSON(t, after))
	}

	// A targeted read: the plug's OnOff, and a path the node lacks.
	var read struct {
		Values []struct {
			Endpoint int             `json:"endpoint"`
			Value    json.RawMessage `json:"value"`
			Status   *int            `json:"status"`
		} `json:"values"`
	}
	s.json(http.MethodPost, "/controller/nodes/"+id+"/attributes",
		fmt.Sprintf(`{"paths":[{"endpoint":%d,"cluster_hex":"0x0006","attribute_hex":"0x0000"},{"endpoint":%d,"cluster_hex":"0x0006","attribute_hex":"0x7777"}]}`, onOff, onOff),
		http.StatusOK, &read)
	if len(read.Values) != 2 || string(read.Values[0].Value) != "false" || read.Values[1].Status == nil {
		t.Fatalf("read %s", mustJSON(t, read))
	}

	// The inventory from the controller's image (no interaction): the
	// plug's endpoint with its DeviceTypeList, Reachable, no TagList, its
	// clusters without values; OnOff with its value, feature and commands.
	var epd struct {
		Reachable   bool `json:"reachable"`
		DeviceTypes []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"device_types"`
		Tags []struct {
			Namespace int    `json:"namespace"`
			Tag       int    `json:"tag"`
			Label     string `json:"label"`
			MfgCode   *int   `json:"mfg_code"`
		} `json:"tags"`
		Clusters []struct {
			ID         int `json:"id"`
			Attributes []struct {
				Value json.RawMessage `json:"value"`
			} `json:"attributes"`
		} `json:"clusters"`
		Conformance *struct {
			Status string `json:"status"`
		} `json:"conformance"`
	}
	s.json(http.MethodGet, fmt.Sprintf("/controller/nodes/%s/endpoints/%d", id, onOff), "", http.StatusOK, &epd)
	types := map[int]string{}
	for _, dt := range epd.DeviceTypes {
		types[dt.ID] = dt.Name
	}
	if !epd.Reachable || types[0x010A] != "OnOffPlugInUnit" || types[0x0013] != "BridgedNode" || epd.Tags == nil || len(epd.Tags) != 0 ||
		epd.Conformance == nil || epd.Conformance.Status != "clean" || len(epd.Clusters) == 0 {
		t.Fatalf("endpoint detail %s", mustJSON(t, epd))
	}
	for _, c := range epd.Clusters {
		for _, a := range c.Attributes {
			if a.Value != nil {
				t.Fatalf("the endpoint detail carries a value: %s", mustJSON(t, c))
			}
		}
	}
	type element struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	var cd struct {
		Name          string `json:"name"`
		FeatureMapHex string `json:"feature_map_hex"`
		Features      []struct {
			Code  string `json:"code"`
			Title string `json:"title"`
			Bit   int    `json:"bit"`
		} `json:"features"`
		DataVersion *int `json:"data_version"`
		Attributes  []struct {
			ID    int             `json:"id"`
			Name  string          `json:"name"`
			Value json.RawMessage `json:"value"`
		} `json:"attributes"`
		Accepted []element `json:"accepted"`
	}
	s.json(http.MethodGet, fmt.Sprintf("/controller/nodes/%s/endpoints/%d/clusters/0x0006", id, onOff), "", http.StatusOK, &cd)
	var onOffValue string
	for _, a := range cd.Attributes {
		if a.ID == 0 && a.Name == "OnOff" {
			onOffValue = string(a.Value)
		}
	}
	if cd.Name != "OnOff" || cd.FeatureMapHex != "0x00000001" || len(cd.Features) != 1 || cd.Features[0].Code != "LT" ||
		cd.Features[0].Title != "Lighting" || cd.Features[0].Bit != 0 || cd.DataVersion == nil || onOffValue != "false" ||
		!slices.Contains(cd.Accepted, element{ID: 2, Name: "Toggle"}) {
		t.Fatalf("cluster detail %s", mustJSON(t, cd))
	}
	if st, out := s.call(http.MethodGet, fmt.Sprintf("/controller/nodes/%s/endpoints/%d/clusters/0x0300", id, onOff), ""); st != http.StatusNotFound {
		t.Fatalf("an absent cluster answers %d %s", st, out)
	}
	if st, out := s.call(http.MethodGet, "/controller/nodes/"+id+"/endpoints/99", ""); st != http.StatusNotFound {
		t.Fatalf("an absent endpoint answers %d %s", st, out)
	}

	// An invoke the plug answers, then the read shows it.
	s.json(http.MethodPost, "/controller/nodes/"+id+"/commands", fmt.Sprintf(`{"endpoint":%d,"cluster_hex":"0x0006","command_hex":"0x01"}`, onOff), http.StatusOK, &wr)
	if wr.Status != 0 {
		t.Fatalf("On status %#x", wr.Status)
	}

	var fabrics struct {
		Fabrics []struct {
			Own   bool   `json:"own"`
			Label string `json:"label"`
		} `json:"fabrics"`
	}
	s.json(http.MethodGet, "/controller/nodes/"+id+"/fabrics", "", http.StatusOK, &fabrics)
	if len(fabrics.Fabrics) != 1 || !fabrics.Fabrics[0].Own {
		t.Fatalf("fabrics %+v", fabrics)
	}
	if st, out := s.call(http.MethodGet, "/controller/nodes/"+id+"/icd", ""); st != http.StatusOK || !bytes.Contains(out, []byte(`"supported":false`)) {
		t.Fatalf("icd %d %s", st, out)
	}

	s.json(http.MethodDelete, "/controller/nodes/"+id, "", http.StatusNoContent, nil)
	s.waitEvent(fabricui.EventNodeRemoved)
	s.json(http.MethodGet, "/controller/nodes", "", http.StatusOK, &list)
	if len(list.Nodes) != 0 {
		t.Fatalf("nodes after decommission %+v", list)
	}
	if st, _ := s.call(http.MethodGet, "/controller/nodes/"+id, ""); st != http.StatusNotFound {
		t.Fatalf("decommissioned node answers %d", st)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// eventRecord is one EventRecord of an event route.
type eventRecord struct {
	Number   int             `json:"number"`
	Priority string          `json:"priority"`
	Endpoint int             `json:"endpoint"`
	Cluster  string          `json:"cluster_hex"`
	Event    string          `json:"event_hex"`
	Name     string          `json:"name"`
	EpochMS  *int64          `json:"epoch_ms"`
	Data     json.RawMessage `json:"data"`
}

// events reads an event route.
func (s *ctrlSite) eventList(path string) []eventRecord {
	s.t.Helper()
	var l struct {
		Events []eventRecord `json:"events"`
	}
	s.json(http.MethodGet, path, "", http.StatusOK, &l)
	return l.Events
}

// find is the first event of cluster and event, nil without one.
func find(evs []eventRecord, cluster, event string) *eventRecord {
	for i := range evs {
		if evs[i].Cluster == cluster && evs[i].Event == event {
			return &evs[i]
		}
	}
	return nil
}

// endpointDetail is the part of an EndpointDetail the test reads.
type endpointDetail struct {
	Name        string `json:"name"`
	Parent      *int   `json:"parent_endpoint_id"`
	Parts       []int  `json:"parts"`
	Scope       string `json:"scope"`
	DeviceTypes []struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Revision int    `json:"revision"`
	} `json:"device_types"`
	Clusters []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Revision   int    `json:"revision"`
		FeatureMap string `json:"feature_map_hex"`
		Features   []struct {
			Code string `json:"code"`
		} `json:"features"`
		DataVersion *int `json:"data_version"`
		Attributes  []struct {
			ID    int             `json:"id"`
			Name  string          `json:"name"`
			Value json.RawMessage `json:"value"`
		} `json:"attributes"`
		Accepted []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"accepted"`
	} `json:"clusters"`
	Decisions []struct {
		Endpoint string `json:"endpoint"`
	} `json:"decisions"`
}

// value is the JSON value of attribute attr of cluster in d; "" without one.
func (d *endpointDetail) value(cluster, attr int) string {
	for i := range d.Clusters {
		c := &d.Clusters[i]
		if c.ID != cluster {
			continue
		}
		for _, a := range c.Attributes {
			if a.ID == attr {
				return string(a.Value)
			}
		}
	}
	return ""
}

// TestEndToEnd_DetailsAndEvents: the node family's endpoint detail and
// event log and the controller family's event history, on the real
// adapters of a fabrictest pair — the root, the Aggregator and the plug
// with OnOff's value; an On the controller invokes reaches the node's
// attribute_changed and the plug's detail; the root's StartUp reaches the
// controller's history with the newly-commissioned read, and the plug's
// ReachableChanged with the default subscription.
func TestEndToEnd_DetailsAndEvents(t *testing.T) {
	s := newCtrlSite(t)
	changes := make(chan fabricui.AttributeChangedEvent, 64)
	defer s.dev.Node().Observe(fabricui.NodeObservers(func(e fabricui.Event) {
		if c, ok := e.Data.(fabricui.AttributeChangedEvent); ok {
			select {
			case changes <- c:
			default:
			}
		}
	}, false))()
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var added nodeDetail
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.Manual), http.StatusOK, &added)
	id := added.Node.NodeIDHex
	plugEP, ok := s.plug.EndpointNumber()
	if !ok {
		t.Fatal("the plug has no endpoint")
	}
	agg := s.dev.Node().Topology().Aggregator

	// The root and the Aggregator answer, each with its direct children.
	var root, aggr, plug endpointDetail
	s.json(http.MethodGet, "/node/endpoints/0", "", http.StatusOK, &root)
	if len(root.DeviceTypes) == 0 || root.DeviceTypes[0].Name != "RootNode" || root.DeviceTypes[0].Revision == 0 ||
		!slices.Equal(root.Parts, []int{int(agg)}) || root.Parent != nil || len(root.Decisions) == 0 || root.value(0x0028, 0x0003) == "" {
		t.Fatalf("root %s", mustJSON(t, root))
	}
	s.json(http.MethodGet, fmt.Sprintf("/node/endpoints/%d", agg), "", http.StatusOK, &aggr)
	if aggr.Name != "Aggregator" || !slices.Equal(aggr.Parts, []int{int(plugEP)}) || aggr.Parent != nil {
		t.Fatalf("aggregator %s", mustJSON(t, aggr))
	}
	// The plug: its device types, OnOff's value, feature and commands by
	// name, its parent, scope and decisions.
	s.json(http.MethodGet, fmt.Sprintf("/node/endpoints/%d", plugEP), "", http.StatusOK, &plug)
	if plug.Name != "Plug" || plug.Parent == nil || *plug.Parent != int(agg) || plug.Scope == "" || len(plug.Decisions) == 0 ||
		plug.value(0x0006, 0x0000) != "false" {
		t.Fatalf("plug %s", mustJSON(t, plug))
	}
	for i := range plug.Clusters {
		c := &plug.Clusters[i]
		if c.ID != 0x0006 {
			continue
		}
		if c.Name != "OnOff" || c.Revision == 0 || c.FeatureMap != "0x00000001" || len(c.Features) != 1 || c.Features[0].Code != "LT" ||
			c.DataVersion == nil || c.Attributes[0].Name != "OnOff" || !slices.ContainsFunc(c.Accepted, func(a struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		},
		) bool {
			return a.ID == 2 && a.Name == "Toggle"
		}) {
			t.Fatalf("plug OnOff %s", mustJSON(t, c))
		}
	}
	if st, out := s.call(http.MethodGet, "/node/endpoints/99", ""); st != http.StatusNotFound {
		t.Fatalf("an endpoint the node does not serve: %d %s", st, out)
	}

	// The controller's On reaches the node: attribute_changed names
	// OnOff's attribute 0, the detail reads true.
	var wr struct {
		Status int `json:"status"`
	}
	s.json(http.MethodPost, "/controller/nodes/"+id+"/commands", fmt.Sprintf(`{"endpoint":%d,"cluster_hex":"0x0006","command_hex":"0x01"}`, plugEP), http.StatusOK, &wr)
	deadline := time.After(10 * time.Second)
	for seen := false; !seen; {
		select {
		case c := <-changes:
			seen = c.Endpoint == plugEP && c.ClusterHex == "0x0006" && slices.Contains(c.Attributes, "0x0000")
		case <-deadline:
			t.Fatal("no attribute_changed for the plug's OnOff")
		}
	}
	s.json(http.MethodGet, fmt.Sprintf("/node/endpoints/%d", plugEP), "", http.StatusOK, &plug)
	if plug.value(0x0006, 0x0000) != "true" {
		t.Fatalf("plug after On %s", mustJSON(t, plug))
	}

	// The root's StartUp: in the node's log and, read when the controller
	// commissioned the node, in the controller's history.
	startUp := find(s.eventList("/node/endpoints/0/events"), "0x0028", "0x0000")
	if startUp == nil || startUp.Name != "StartUp" || startUp.Priority != "critical" || startUp.EpochMS == nil ||
		!strings.Contains(string(startUp.Data), `"softwareVersion"`) {
		t.Fatalf("node StartUp %+v", startUp)
	}
	history := s.eventList("/controller/nodes/" + id + "/events")
	got := find(history, "0x0028", "0x0000")
	if got == nil || got.Number != startUp.Number || got.Name != "StartUp" || string(got.Data) != string(startUp.Data) {
		t.Fatalf("controller StartUp %+v, node's %+v", got, startUp)
	}
	last := history[len(history)-1].Number
	if newer := s.eventList(fmt.Sprintf("/controller/nodes/%s/events?since=%d", id, last+1)); len(newer) != 0 {
		t.Fatalf("events after the last %+v", newer)
	}

	// The plug's ReachableChanged reaches the controller through the
	// default subscription.
	s.json(http.MethodPost, "/controller/nodes/"+id+"/auto-subscribe", `{"enabled":true}`, http.StatusNoContent, nil)
	s.plug.SetReachable(context.Background(), false)
	reach := find(s.eventList(fmt.Sprintf("/node/endpoints/%d/events", plugEP)), "0x0039", "0x0003")
	if reach == nil || reach.Name != "ReachableChanged" || string(reach.Data) != `{"reachableNewValue":false}` {
		t.Fatalf("node ReachableChanged %+v", reach)
	}
	// The route reads the controller's memory, no interaction: it is
	// called while Eventually moves the site's clock to the report.
	got = nil
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := s.site.Eventually(ctx, "the plug's ReachableChanged in the controller's history", func() bool {
		got = find(s.rawEvents(fmt.Sprintf("/controller/nodes/%s/events?since=%d", id, last+1)), "0x0039", "0x0003")
		return got != nil
	}); err != nil {
		t.Fatal(err)
	}
	if got.Endpoint != int(plugEP) || got.Number != reach.Number || string(got.Data) != string(reach.Data) {
		t.Fatalf("controller ReachableChanged %+v, node's %+v", got, reach)
	}
}

// rawEvents reads an event route outside Site.Run; nil on any failure.
func (s *ctrlSite) rawEvents(path string) []eventRecord {
	resp, err := s.srv.Client().Get(s.srv.URL + fabricui.DefaultPrefix + path)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var l struct {
		Events []eventRecord `json:"events"`
	}
	if resp.StatusCode != http.StatusOK || json.NewDecoder(resp.Body).Decode(&l) != nil {
		return nil
	}
	return l.Events
}
