// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

// The adapter's mappings that the handler tests (fake node,
// MemoryController) never reach, end to end on fabrictest (UR-P3-test-30):
// the update queue's status, a mistyped attribute write, the ICD state and
// its registration, and a Binding list written and read back.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	icddef "github.com/SukramJ/go-fabric/cluster/spec/icdmanagement"
	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/attestation"
	"github.com/SukramJ/go-fabric/device"
	"github.com/SukramJ/go-fabric/fabrictest"
	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/fabricui"
	"github.com/SukramJ/go-fabric-ui/fabricui/gofabric"
)

// commissionSiteDevice commissions the site's device and waits for its
// subscription; it returns the node's id as the contract carries it.
func commissionSiteDevice(t *testing.T, s *ctrlSite) string {
	t.Helper()
	sp, err := s.dev.Node().SetupPayload()
	if err != nil {
		t.Fatal(err)
	}
	var added nodeDetail
	s.json(http.MethodPost, "/controller/commission", fmt.Sprintf(`{"code":%q}`, sp.Manual), http.StatusOK, &added)
	s.waitEvent(fabricui.EventSubscription)
	return added.Node.NodeIDHex
}

// within polls check every 100 ms for up to 10 s.
func within(t *testing.T, what string, check func() bool) {
	t.Helper()
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(100 * time.Millisecond) {
		if check() {
			return
		}
	}
	t.Fatalf("%s: not within 10 s", what)
}

// The update queue's status is the contract's enum: the manager's
// "in-progress" travels as "in_progress".
func TestEndToEnd_UpdateQueueStatusIsTheContractEnum(t *testing.T) {
	dir := t.TempDir()
	s := newCtrlSiteWith(t, ctrlSiteOptions{
		ota:       &controller.OTAConfig{},
		configure: func(cfg *node.Config) { cfg.OTA = &node.OTAConfig{Dir: dir} },
	})
	id := commissionSiteDevice(t, s)
	b := s.src.Nodes()[0].Basic()
	s.json(http.MethodPost, "/controller/updates/consent",
		fmt.Sprintf(`{"node_id_hex":%q,"vendor_id":%d,"product_id":%d,"target_software_version":2}`, id, b.VendorID, b.ProductID),
		http.StatusNoContent, nil)
	enum := []string{"queued", "in_progress", "stalled"}
	within(t, "status in_progress", func() bool {
		var list struct {
			Queue []struct {
				Status string `json:"status"`
			} `json:"queue"`
		}
		s.json(http.MethodGet, "/controller/updates", "", http.StatusOK, &list)
		for _, q := range list.Queue {
			if !slices.Contains(enum, q.Status) {
				t.Fatalf("queue status %q is not in the contract's enum %v", q.Status, enum)
			}
			if q.Status == "in_progress" {
				return true
			}
		}
		return false
	})
}

// A write value that does not fit the attribute's type is the request's
// error: 400 bad_request, not the node's 502.
func TestEndToEnd_MistypedAttributeWriteIsABadRequest(t *testing.T) {
	s := newCtrlSite(t)
	id := commissionSiteDevice(t, s)
	// NodeLabel (Basic Information 0x0028/0x0005) is a string.
	status, out := s.call(http.MethodPut, "/controller/nodes/"+id+"/attributes",
		`{"endpoint":0,"cluster_hex":"0x0028","attribute_hex":"0x0005","value":12}`)
	var p struct {
		Code string `json:"code"`
	}
	if status != http.StatusBadRequest || json.Unmarshal(out, &p) != nil || p.Code != "bad_request" {
		t.Fatalf("mistyped write: %d %s", status, out)
	}
}

// The ICD state of a LIT-capable ICD: SIT while no client is registered,
// LIT once the controller registered as its check-in client, and
// registered false again after the unregistration.
func TestEndToEnd_ICDOperatingModeFollowsTheRegistration(t *testing.T) {
	s := newCtrlSiteWith(t, ctrlSiteOptions{configure: func(cfg *node.Config) {
		cfg.ICD = &node.ICDConfig{
			Mode: node.ICDModeLIT, IdleModeDuration: 10 * time.Second, ActiveModeDuration: time.Second,
			ActiveModeThreshold: 5 * time.Second, MaximumCheckInBackoff: 10 * time.Second, ClientsSupportedPerFabric: 2,
			UserActiveModeTrigger: &node.ICDUserActiveModeTrigger{Hint: icddef.UserActiveModeTriggerPowerCycle},
		}
	}})
	id := commissionSiteDevice(t, s)
	type icdState struct {
		Supported     bool   `json:"supported"`
		LITSupported  bool   `json:"lit_supported"`
		Registered    bool   `json:"registered"`
		OperatingMode string `json:"operating_mode"`
	}
	read := func() icdState {
		var st icdState
		s.json(http.MethodGet, "/controller/nodes/"+id+"/icd", "", http.StatusOK, &st)
		return st
	}
	if st := read(); !st.Supported || !st.LITSupported || st.Registered || st.OperatingMode != "sit" {
		t.Fatalf("before the registration: %+v", st)
	}
	var reg icdState
	s.json(http.MethodPost, "/controller/nodes/"+id+"/icd/register", `{}`, http.StatusOK, &reg)
	if !reg.Registered {
		t.Fatalf("register answered %+v", reg)
	}
	within(t, "operating mode lit", func() bool { return read().OperatingMode == "lit" })
	var unreg icdState
	s.json(http.MethodPost, "/controller/nodes/"+id+"/icd/unregister", ``, http.StatusOK, &unreg)
	if unreg.Registered {
		t.Fatalf("unregister answered %+v", unreg)
	}
}

// newSwitchSite is a site whose device bridges an on/off light switch,
// whose type mandates a Binding server, with the controller family over
// it; the plug of newCtrlSite has none.
func newSwitchSite(t *testing.T) *ctrlSite {
	t.Helper()
	site := fabrictest.NewSite(t, fabrictest.Options{})
	sw, err := device.NewOnOffLightSwitch(device.Options{StableKey: device.StringKey("switch:1"), Name: "Switch"})
	if err != nil {
		t.Fatal(err)
	}
	dev := site.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{sw}})
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
		hooks.Close()
		_ = site.Run(ctx, ctrl.Close)
	})
	src := gofabric.ControllerOf(ctrl, gofabric.ControllerOptions{
		Hooks: hooks, Credentials: fabricui.NewMemoryCredentials(gofabric.ThreadDatasetName),
		Pair: controller.PairOptions{OnAttestationFailure: attestation.Accept()},
	})
	s := &ctrlSite{t: t, site: site, dev: dev, ctrl: ctrl, src: src, events: make(chan string, 256)}
	s.srv = httptest.NewServer(fabricui.Handler(fabricui.Options{Controller: src, Node: dev.Node()}))
	t.Cleanup(s.srv.Close)
	src.Observe(func(e fabricui.ControllerEvent) {
		select {
		case s.events <- e.Type:
		default:
		}
	})
	var st struct {
		CSRFToken string `json:"csrf_token"`
	}
	s.json(http.MethodGet, "/controller/status", "", http.StatusOK, &st)
	s.token = st.CSRFToken
	return s
}

// A Binding list written through PUT …/bindings reads back as written:
// the adapter builds every entry of the list, not an empty one.
func TestEndToEnd_BindingsWriteReadsBack(t *testing.T) {
	s := newSwitchSite(t)
	id := commissionSiteDevice(t, s)
	type target struct {
		NodeIDHex string `json:"node_id_hex,omitempty"`
		Group     *int   `json:"group,omitempty"`
		Endpoint  *int   `json:"endpoint,omitempty"`
		Cluster   string `json:"cluster_hex,omitempty"`
	}
	type list struct {
		Endpoints []struct {
			Endpoint int      `json:"endpoint"`
			Bindings []target `json:"bindings"`
		} `json:"endpoints"`
	}
	var before list
	s.json(http.MethodGet, "/controller/nodes/"+id+"/bindings", "", http.StatusOK, &before)
	if len(before.Endpoints) == 0 {
		t.Fatal("the switch's node lists no Binding server")
	}
	ep := before.Endpoints[0].Endpoint
	one, seven := 1, 7
	want := []target{
		{NodeIDHex: "0000000000000042", Endpoint: &one, Cluster: "0x0006"},
		{Group: &seven, Cluster: "0x0008"},
	}
	body, err := json.Marshal(map[string]any{"endpoint": ep, "bindings": want})
	if err != nil {
		t.Fatal(err)
	}
	var wr struct {
		Status int `json:"status"`
	}
	s.json(http.MethodPut, "/controller/nodes/"+id+"/bindings", string(body), http.StatusOK, &wr)
	if wr.Status != 0 {
		t.Fatalf("write status %d", wr.Status)
	}
	within(t, "the written bindings", func() bool {
		var after list
		s.json(http.MethodGet, "/controller/nodes/"+id+"/bindings", "", http.StatusOK, &after)
		for _, e := range after.Endpoints {
			if e.Endpoint != ep || len(e.Bindings) != len(want) {
				continue
			}
			for i, b := range e.Bindings {
				w := want[i]
				if b.NodeIDHex != w.NodeIDHex || b.Cluster != w.Cluster || (w.Group == nil) != (b.Group == nil) ||
					w.Group != nil && *b.Group != *w.Group || w.Endpoint != nil && (b.Endpoint == nil || *b.Endpoint != *w.Endpoint) {
					t.Fatalf("binding %d: %+v, want %+v", i, b, w)
				}
			}
			return true
		}
		return false
	})
}
