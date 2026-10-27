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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var ctrl *controller.Controller
	if err := site.Run(ctx, func(ctx context.Context) (err error) {
		ctrl, err = controller.New(ctx, cfg)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		hooks.Close() // before the controller goes: no timer outlives it
		_ = site.Run(ctx, ctrl.Close)
	})
	src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{
		Hooks:       hooks,
		Credentials: fabricui.NewMemoryCredentials(gofabric.ThreadDatasetName),
		// The site's device carries the module's test attestation.
		Pair: controller.PairOptions{OnAttestationFailure: attestation.Accept()},
	})
	s := &ctrlSite{t: t, site: site, dev: dev, ctrl: ctrl, events: make(chan string, 256)}
	s.srv = httptest.NewServer(fabricui.Handler(fabricui.Options{Controller: src}))
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
		Endpoint int `json:"endpoint"`
		Clusters []struct {
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
