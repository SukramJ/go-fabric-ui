// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

// failingController is the in-memory controller whose nodes fail every
// interaction and that serves no credentials and no OTA provider.
type failingController struct{ *MemoryController }

func (f failingController) Node(id uint64) (ControllerNode, bool) {
	n, ok := f.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	return failingNode{n}, true
}

func (f failingController) Decommission(context.Context, uint64) error { return ErrNodeFailed }
func (f failingController) SetFabricLabel(context.Context, string) error {
	return ErrInvalidArgument
}
func (f failingController) Credentials() CredentialStore          { return nil }
func (f failingController) SoftwareUpdates() SoftwareUpdateSource { return nil }
func (f failingController) Discover(context.Context, DiscoverFilter, time.Duration) ([]Discovered, error) {
	return nil, ErrNodeFailed
}

func (f failingController) Pair(context.Context, string, PairOptions) (ControllerNode, error) {
	return nil, context.DeadlineExceeded
}

type failingNode struct{ ControllerNode }

func (failingNode) Read(context.Context, []AttributePath, bool) ([]AttributeValue, error) {
	return nil, ErrNodeFailed
}

func (failingNode) Write(context.Context, AttributePath, any, time.Duration) (WriteStatus, error) {
	return WriteStatus{}, ErrNodeFailed
}

func (failingNode) Invoke(context.Context, CommandPath, json.RawMessage, time.Duration) (InvokeResult, error) {
	return InvokeResult{}, ErrNodeFailed
}

func (failingNode) OpenCommissioningWindow(context.Context, time.Duration) (WindowInfo, error) {
	return WindowInfo{}, ErrNodeFailed
}
func (failingNode) RevokeCommissioningWindow(context.Context) error { return ErrNodeFailed }
func (failingNode) Interview(context.Context) error                 { return ErrNodeFailed }
func (failingNode) Enable(context.Context) error                    { return ErrNodeFailed }
func (failingNode) Disable(context.Context) error                   { return ErrNodeFailed }
func (failingNode) SetAutoSubscribe(context.Context, bool) error    { return ErrNodeFailed }
func (failingNode) RegisterICD(context.Context, bool) error         { return ErrNodeFailed }
func (failingNode) UnregisterICD(context.Context, bool) error       { return ErrNodeFailed }

// controllerRoutes is a request for every controller route: method,
// path with {id} for the node, a valid body.
var controllerRoutes = []struct{ method, path, body string }{
	{http.MethodGet, "/controller/nodes/{id}", ""},
	{http.MethodDelete, "/controller/nodes/{id}", ""},
	{http.MethodPost, "/controller/nodes/{id}/interview", ""},
	{http.MethodPost, "/controller/nodes/{id}/enable", ""},
	{http.MethodPost, "/controller/nodes/{id}/disable", ""},
	{http.MethodPost, "/controller/nodes/{id}/auto-subscribe", `{"enabled":true}`},
	{http.MethodPost, "/controller/nodes/{id}/window", `{}`},
	{http.MethodPost, "/controller/nodes/{id}/window/close", ""},
	{http.MethodPost, "/controller/nodes/{id}/attributes", `{"paths":[{"endpoint":1,"cluster_hex":"6","attribute_hex":"0"}]}`},
	{http.MethodPut, "/controller/nodes/{id}/attributes", `{"endpoint":1,"cluster_hex":"6","attribute_hex":"0","value":true,"timed_ms":100}`},
	{http.MethodPost, "/controller/nodes/{id}/commands", `{"endpoint":1,"cluster_hex":"6","command_hex":"2","payload":{},"timed_ms":100}`},
	{http.MethodGet, "/controller/nodes/{id}/acl", ""},
	{http.MethodPut, "/controller/nodes/{id}/acl", `{"entries":[{"privilege":"view","auth_mode":"case","subjects_hex":["5"],"targets":[{"cluster_hex":"6","endpoint":1,"device_type_hex":"100"}]}]}`},
	{http.MethodGet, "/controller/nodes/{id}/bindings", ""},
	{http.MethodPut, "/controller/nodes/{id}/bindings", `{"endpoint":1,"bindings":[{"node_id_hex":"12","endpoint":1,"cluster_hex":"6"}]}`},
	{http.MethodGet, "/controller/nodes/{id}/icd", ""},
	{http.MethodPost, "/controller/nodes/{id}/icd/register", `{"allow_multi_admin":true}`},
	{http.MethodPost, "/controller/nodes/{id}/icd/unregister", `{}`},
	{http.MethodGet, "/controller/nodes/{id}/fabrics", ""},
	{http.MethodDelete, "/controller/nodes/{id}/fabrics/2", ""},
}

// An unknown node is 404 and a malformed id 400 on every node route.
func TestController_NodeRoutesRefuseUnknownAndMalformedIDs(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	for _, r := range controllerRoutes {
		for id, want := range map[string]int{"00000000000000AA": http.StatusNotFound, "zz": http.StatusBadRequest} {
			rec := do(t, mux, r.method, strings.ReplaceAll(r.path, "{id}", id), r.body)
			if rec.Code != want {
				t.Errorf("%s %s (%s): %d %s", r.method, r.path, id, rec.Code, rec.Body)
			}
		}
	}
}

// A node that fails every interaction answers 502 node_error on each
// interacting route; a controller without credentials or OTA answers 503.
func TestController_FailuresMapToProblems(t *testing.T) {
	src := failingController{NewMemoryController()}
	_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
	for _, r := range controllerRoutes {
		path := strings.ReplaceAll(r.path, "{id}", "0000000000000011")
		rec := do(t, mux, r.method, path, r.body)
		switch {
		case r.method == http.MethodGet && (strings.HasSuffix(path, "/icd") || strings.HasSuffix(path, "0011")):
			if rec.Code != http.StatusOK {
				t.Errorf("%s %s: %d %s", r.method, path, rec.Code, rec.Body)
			}
		case rec.Code < 400:
			t.Errorf("%s %s: %d %s", r.method, path, rec.Code, rec.Body)
		}
	}
	for _, c := range []struct {
		method, path, body string
		want               int
	}{
		{http.MethodGet, "/controller/credentials", "", 503},
		{http.MethodPut, "/controller/credentials", `{"kind":"wifi","id":"a","ssid":"b"}`, 503},
		{http.MethodDelete, "/controller/credentials?kind=wifi&id=a", "", 503},
		{http.MethodPost, "/controller/updates/check", "", 503},
		{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"11","vendor_id":1,"product_id":1,"target_software_version":1}`, 503},
		{http.MethodGet, "/controller/discovery?filter=any&timeout=2", "", 502},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":5}`, 502},
		{http.MethodPut, "/controller/fabric/label", `{"label":"x"}`, 400},
	} {
		if rec := do(t, mux, c.method, c.path, c.body); rec.Code != c.want {
			t.Errorf("%s %s: %d, want %d: %s", c.method, c.path, rec.Code, c.want, rec.Body)
		}
	}
	rec := do(t, mux, http.MethodGet, "/controller/updates", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"available":false`) {
		t.Fatalf("updates %d %s", rec.Code, rec.Body)
	}
}

// The handlers' remaining refusals of a bad body or parameter.
func TestController_BadRequests(t *testing.T) {
	_, mux, m := controllerHandler(t, nil)
	n, _ := m.Node(0x11)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPost, "/controller/nodes/11/auto-subscribe", ""},
		{http.MethodPut, "/controller/nodes/11/attributes", `{"endpoint":1,"cluster_hex":"6","attribute_hex":"0","value":true,"timed_ms":70000}`},
		{http.MethodPut, "/controller/nodes/11/attributes", `{"endpoint":1,"cluster_hex":"6","attribute_hex":"0"}`},
		{http.MethodPost, "/controller/nodes/11/commands", `{"endpoint":1,"cluster_hex":"6","command_hex":"zz"}`},
		{http.MethodPost, "/controller/nodes/11/commands", `{"endpoint":-1,"cluster_hex":"6","command_hex":"2"}`},
		{http.MethodPost, "/controller/nodes/11/attributes", `{"paths":[{"endpoint":1,"cluster_hex":"zz","attribute_hex":"0"}]}`},
		{http.MethodPost, "/controller/nodes/11/attributes", `{"paths":[` + strings.Repeat(`{"endpoint":1,"cluster_hex":"6","attribute_hex":"0"},`, 32) + `{"endpoint":1,"cluster_hex":"6","attribute_hex":"0"}]}`},
		{http.MethodPut, "/controller/nodes/11/acl", `{"entries":[{"privilege":"view","auth_mode":"x","subjects_hex":[],"targets":[]}]}`},
		{http.MethodPut, "/controller/nodes/11/acl", `{"entries":[{"privilege":"view","auth_mode":"case","subjects_hex":["zz"],"targets":[]}]}`},
		{http.MethodPut, "/controller/nodes/11/acl", `{"entries":[{"privilege":"view","auth_mode":"case","subjects_hex":[],"targets":[{"cluster_hex":"zz"}]}]}`},
		{http.MethodPut, "/controller/nodes/11/bindings", `{"endpoint":1,"bindings":[{"group":0}]}`},
		{http.MethodPut, "/controller/nodes/11/bindings", `{"endpoint":1,"bindings":[{"node_id_hex":"zz","endpoint":1}]}`},
		{http.MethodPut, "/controller/nodes/11/bindings", `{"endpoint":1,"bindings":[{"group":1,"cluster_hex":"zz"}]}`},
		{http.MethodDelete, "/controller/nodes/11/fabrics/0", ""},
		{http.MethodPost, "/controller/nodes/11/icd/unregister", `{}`},
		{http.MethodGet, "/controller/discovery?filter=colour", ""},
		{http.MethodGet, "/controller/discovery?filter=vendor&value=-1", ""},
		{http.MethodPost, "/controller/commission", `{"discovered":{"instance_name":"x"}}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","passcode":0}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":0}`},
		{http.MethodPut, "/controller/credentials", `{"kind":"wifi","id":""}`},
		{http.MethodPut, "/controller/credentials", `{"kind":"x","id":"a"}`},
		{http.MethodPut, "/controller/credentials", `{"kind":"thread","id":"a"}`},
		{http.MethodDelete, "/controller/credentials?kind=x&id=a", ""},
		{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"zz","vendor_id":1,"product_id":1,"target_software_version":1}`},
		{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"11","vendor_id":-1,"product_id":1,"target_software_version":1}`},
	} {
		if rec := do(t, mux, c.method, c.path, c.body); rec.Code < 400 {
			t.Errorf("%s %s %s: %d %s", c.method, c.path, c.body, rec.Code, rec.Body)
		}
	}
	// What a valid flow still covers: a Thread dataset, an ICD unregister,
	// a commission with a timeout and network, a re-enabled node.
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/controller/credentials", `{"kind":"thread","id":"t","dataset_hex":"0304414243"}`},
		{http.MethodPost, "/controller/nodes/12/icd/unregister", `{"force":false}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","wifi_id":"home","thread_id":"mesh","network_only":true,"timeout_seconds":30}`},
		{http.MethodPost, "/controller/nodes/11/disable", ""},
		{http.MethodPost, "/controller/nodes/11/enable", ""},
		{http.MethodGet, "/controller/discovery?filter=vendor&value=65521&timeout=1", ""},
		{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"11","vendor_id":65521,"product_id":1,"target_software_version":9}`},
	} {
		if rec := do(t, mux, c.method, c.path, c.body); rec.Code >= 400 {
			t.Errorf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
	if ssid, _, ok := m.creds.WiFi("home"); !ok || ssid != "HomeNet" {
		t.Fatal("wifi secret")
	}
	if ds, ok := m.creds.Thread("t"); !ok || len(ds) != 5 {
		t.Fatal("thread secret")
	}
	if n.Disabled() {
		t.Fatal("re-enabled node is disabled")
	}
	for _, e := range []ControllerEvent{
		{Type: EventConnection, Node: 1, Connection: ConnectionConnected},
		{Type: EventSubscription, Node: 1, Active: true},
		{Type: EventCommissioningProgress, Phase: "started", Detail: "x"},
	} {
		if ev := controllerEvent(e); ev.Type != e.Type {
			t.Fatal(ev)
		}
	}
	for _, err := range []error{ErrInvalidArgument, ErrNoDevice, context.DeadlineExceeded, ErrNodeFailed, context.Canceled} {
		if progressDetail(err) == "" {
			t.Fatal(err)
		}
	}
	if (&apiProblem{title: "t", detail: "d"}).Error() != "t: d" {
		t.Fatal("apiProblem.Error")
	}
}
