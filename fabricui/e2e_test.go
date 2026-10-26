// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui_test

import (
	"bufio"
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

	"github.com/SukramJ/go-fabric/controller/commissioner"
	"github.com/SukramJ/go-fabric/device"
	"github.com/SukramJ/go-fabric/fabrictest"
	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// site is one device of a fabrictest site behind the handler on an HTTP
// test server, with the stream read into events. The device has started
// before the handler is built: the handler learns of its transitions
// through node.Observe, not through observers wired into its
// configuration.
type site struct {
	t      *testing.T
	site   *fabrictest.Site
	device *fabrictest.Device
	srv    *httptest.Server
	token  string
	events chan string
}

func newSite(t *testing.T, configured bool) *site {
	t.Helper()
	s := &site{t: t, site: fabrictest.NewSite(t, fabrictest.Options{}), events: make(chan string, 32)}
	plug, err := device.NewOnOffPlugInUnit(device.Options{StableKey: device.StringKey("plug:1"), Name: "Plug"})
	if err != nil {
		t.Fatal(err)
	}
	// A bridged node with a part: the part's clusters come from the
	// topology too (node.BridgedDevice.Clusters), not from what was added.
	owner, err := device.NewBridgedNode(device.Options{StableKey: device.StringKey("owner:1"), Name: "Owner"})
	if err != nil {
		t.Fatal(err)
	}
	light, err := device.NewOnOffLight(device.Options{StableKey: device.StringKey("owner:1/light"), Name: "Light"})
	if err != nil {
		t.Fatal(err)
	}
	if err := owner.AddPart(light.Device); err != nil {
		t.Fatal(err)
	}
	s.device = s.site.AddDevice(fabrictest.Topology{Bridged: []node.Endpoint{plug, owner}})
	s.srv = httptest.NewServer(fabricui.Handler(fabricui.Options{
		Node: s.device.Node(), ConfiguredWindow: configured,
	}))
	t.Cleanup(s.srv.Close)

	var st struct {
		Enabled   bool   `json:"enabled"`
		CSRFToken string `json:"csrf_token"`
	}
	s.getJSON("/node/status", &st)
	if !st.Enabled || st.CSRFToken == "" {
		t.Fatalf("status %+v", st)
	}
	s.token = st.CSRFToken

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.srv.URL+"/api/v1/node/stream", http.NoBody)
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	connected := make(chan struct{})
	go func() {
		defer resp.Body.Close()
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			line := sc.Text()
			if line == ": connected" {
				close(connected)
			}
			if typ, ok := strings.CutPrefix(line, "event: "); ok {
				s.events <- typ
			}
		}
	}()
	<-connected
	return s
}

func (s *site) do(method, path, body string) (status int, respBody []byte) {
	s.t.Helper()
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req, _ := http.NewRequestWithContext(context.Background(), method, s.srv.URL+"/api/v1"+path, rd)
	req.Header.Set(fabricui.TokenHeader, s.token)
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		s.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b
}

func (s *site) getJSON(path string, v any) {
	s.t.Helper()
	code, b := s.do(http.MethodGet, path, "")
	if code != http.StatusOK {
		s.t.Fatalf("GET %s: %d %s", path, code, b)
	}
	if err := json.Unmarshal(b, v); err != nil {
		s.t.Fatal(err)
	}
}

func (s *site) expect(code int, method, path, body string) []byte {
	s.t.Helper()
	got, b := s.do(method, path, body)
	if got != code {
		s.t.Fatalf("%s %s: %d %s, want %d", method, path, got, b, code)
	}
	return b
}

// waitEvent waits for the stream to deliver typ, skipping others.
func (s *site) waitEvent(typ string) {
	s.t.Helper()
	timeout := time.After(30 * time.Second)
	for {
		select {
		case got := <-s.events:
			if got == typ {
				return
			}
		case <-timeout:
			s.t.Fatalf("no %s event on the stream", typ)
		}
	}
}

type fabricList struct {
	Fabrics []struct {
		FabricIndex int    `json:"fabric_index"`
		VendorID    int    `json:"vendor_id"`
		NodeIDHex   string `json:"node_id_hex"`
		Label       string `json:"label"`
	} `json:"fabrics"`
}

type session struct {
	FabricIndex int  `json:"fabric_index"`
	IsPASE      bool `json:"is_pase"`
}

type sessionList struct {
	Sessions  []session `json:"sessions"`
	Occupancy struct {
		Live int `json:"live"`
	} `json:"occupancy"`
}

// commissionAndRemove is the end-to-end path the brief names: the window
// the handler opened commissions a controller, the fabric is listed, the
// DELETE removes it, and the stream reported both transitions.
func (s *site) commissionAndRemove(passcode uint32) {
	s.t.Helper()
	c := s.site.AddController(fabrictest.ControllerOptions{Label: "Controller"})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	opts := fabrictest.CommissionOptions{}
	if passcode != 0 {
		opts.Edit = func(o *commissioner.Options) { o.Passcode = passcode }
	}
	if _, err := c.Commission(ctx, s.device, opts); err != nil {
		s.t.Fatal(err)
	}
	s.waitEvent(fabricui.EventFabricAdded)
	s.waitEvent(fabricui.EventWindowClosed)

	// The controller's CASE session is listed, on the node's own
	// diagnostics (node.SessionDiagnostics).
	var sl sessionList
	s.getJSON("/node/sessions", &sl)
	if sl.Occupancy.Live == 0 || !slices.ContainsFunc(sl.Sessions, func(x session) bool { return !x.IsPASE && x.FabricIndex != 0 }) {
		s.t.Fatalf("sessions after commissioning %+v", sl)
	}

	var fl fabricList
	s.getJSON("/node/fabrics", &fl)
	if len(fl.Fabrics) != 1 || fl.Fabrics[0].VendorID != 0xFFF1 || fl.Fabrics[0].Label != "Controller" || len(fl.Fabrics[0].NodeIDHex) != 16 {
		s.t.Fatalf("fabrics after commissioning %+v", fl)
	}
	s.expect(http.StatusNoContent, http.MethodDelete, fmt.Sprintf("/node/fabrics/%d", fl.Fabrics[0].FabricIndex), "")
	s.waitEvent(fabricui.EventFabricRemoved)
	s.getJSON("/node/fabrics", &fl)
	if len(fl.Fabrics) != 0 {
		s.t.Fatalf("fabrics after DELETE %+v", fl)
	}
}

type status struct {
	WindowOpen     bool   `json:"commissioning_window_open"`
	WindowKind     string `json:"window_kind"`
	WindowDeadline string `json:"window_deadline"`
	WindowDuration int    `json:"commissioning_window_duration_seconds"`
}

// reopen closes the uncommissioned node's own window and opens one through
// the handler; the status names each window by its kind and deadline.
func (s *site) reopen(wantKind string) []byte {
	s.t.Helper()
	st := s.status()
	if !st.WindowOpen || st.WindowKind != "own" || st.WindowDeadline == "" {
		s.t.Fatalf("status of the uncommissioned node %+v", st)
	}
	s.expect(http.StatusConflict, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`)
	s.expect(http.StatusNoContent, http.MethodPost, "/node/pairing/window/close", "")
	s.waitEvent(fabricui.EventWindowClosed)
	st = s.status()
	if st.WindowOpen || st.WindowKind != "none" || st.WindowDeadline != "" {
		s.t.Fatalf("status after the close %+v", st)
	}
	b := s.expect(http.StatusOK, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`)
	s.waitEvent(fabricui.EventWindowOpened)
	st = s.status()
	// The deadline is on the node's clock (the site's), not wall time.
	deadline, err := time.Parse(time.RFC3339, st.WindowDeadline)
	left := deadline.Sub(s.device.Node().Clock().Now())
	if !st.WindowOpen || st.WindowKind != wantKind || st.WindowDuration != 300 || err != nil ||
		left <= 4*time.Minute || left > 5*time.Minute {
		s.t.Fatalf("status after the open %+v (%v)", st, err)
	}
	return b
}

// status reads GET /node/status into a fresh value.
func (s *site) status() status {
	s.t.Helper()
	var st status
	s.getJSON("/node/status", &st)
	return st
}

// checkEndpoints reads GET /node/endpoints: every endpoint, the part
// included, lists its mounted clusters with their schema names.
func (s *site) checkEndpoints() {
	s.t.Helper()
	var el struct {
		Endpoints []struct {
			EndpointID         int    `json:"endpoint_id"`
			ParentEndpointID   int    `json:"parent_endpoint_id"`
			DeviceTypeRevision int    `json:"device_type_revision"`
			Name               string `json:"name"`
			Clusters           []struct {
				ID       uint32 `json:"id"`
				Name     string `json:"name"`
				Revision int    `json:"revision"`
			} `json:"clusters"`
		} `json:"endpoints"`
	}
	s.getJSON("/node/endpoints", &el)
	if len(el.Endpoints) != 3 {
		s.t.Fatalf("endpoints %+v, want the plug, the owner and its part", el)
	}
	const descriptor, onOff = 0x001D, 0x0006
	for _, ep := range el.Endpoints {
		ids := make([]uint32, 0, len(ep.Clusters))
		for _, c := range ep.Clusters {
			if c.Name == "" || c.Revision == 0 {
				s.t.Fatalf("endpoint %d cluster %+v without name or revision", ep.EndpointID, c)
			}
			ids = append(ids, c.ID)
		}
		if ep.DeviceTypeRevision == 0 || !slices.Contains(ids, descriptor) {
			s.t.Fatalf("endpoint %+v", ep)
		}
		if ep.Name == "Light" && (ep.ParentEndpointID == 0 || !slices.Contains(ids, onOff)) {
			s.t.Fatalf("the part %+v", ep)
		}
	}
}

func TestEndToEnd_ConfiguredWindowCommissionsAndRemoves(t *testing.T) {
	s := newSite(t, true)
	var fl fabricList
	s.getJSON("/node/fabrics", &fl)
	if len(fl.Fabrics) != 0 {
		t.Fatalf("fabrics before commissioning %+v", fl)
	}
	s.checkEndpoints()
	var p struct {
		QRCode        string `json:"qr_code"`
		ManualCode    string `json:"manual_code"`
		Source        string `json:"source"`
		WindowOpen    bool   `json:"window_open"`
		Passcode      uint32 `json:"passcode"`
		Discriminator uint16 `json:"discriminator"`
	}
	s.getJSON("/node/pairing", &p)
	if !strings.HasPrefix(p.QRCode, "MT:") || len(p.ManualCode) != 11 || p.Source != "configured" || !p.WindowOpen ||
		p.Passcode != s.device.Passcode() || p.Discriminator != s.device.Discriminator() {
		t.Fatalf("pairing %+v", p)
	}
	var w struct {
		QRCode   string `json:"qr_code"`
		Passcode uint32 `json:"passcode"`
	}
	if err := json.Unmarshal(s.reopen("own"), &w); err != nil {
		t.Fatal(err)
	}
	if w.QRCode != p.QRCode || w.Passcode != s.device.Passcode() {
		t.Fatalf("configured window %+v, want the configured codes and passcode", w)
	}
	s.commissionAndRemove(0)
}

func TestEndToEnd_FreshWindowCommissionsAndRemoves(t *testing.T) {
	s := newSite(t, false)
	var w struct {
		Passcode   uint32 `json:"passcode"`
		QRCode     string `json:"qr_code"`
		ManualCode string `json:"manual_code"`
	}
	if err := json.Unmarshal(s.reopen("basic"), &w); err != nil {
		t.Fatal(err)
	}
	if w.Passcode == 0 || w.Passcode == s.device.Passcode() || !strings.HasPrefix(w.QRCode, "MT:") {
		t.Fatalf("fresh window %+v", w)
	}
	var p struct {
		Passcode uint32 `json:"passcode"`
		Source   string `json:"source"`
	}
	s.getJSON("/node/pairing", &p)
	if p.Source != "window" || p.Passcode != w.Passcode {
		t.Fatalf("pairing while the fresh window is open %+v", p)
	}
	s.commissionAndRemove(w.Passcode)
}
