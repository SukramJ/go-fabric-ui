// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// errController is the in-memory controller whose every port call that
// can fail answers err, as a host's own ControllerSource might.
type errController struct {
	*MemoryController
	err error
}

func (c errController) Node(id uint64) (ControllerNode, bool) {
	n, ok := c.MemoryController.Node(id)
	if !ok {
		return nil, false
	}
	return errNode{n, c.err}, true
}

func (c errController) Pair(context.Context, string, PairOptions) (ControllerNode, error) {
	return nil, c.err
}

func (c errController) Discover(context.Context, DiscoverFilter, time.Duration) ([]Discovered, error) {
	return nil, c.err
}

func (c errController) CommissionDiscovered(context.Context, Discovered, uint32, PairOptions) (ControllerNode, error) {
	return nil, c.err
}

func (c errController) Credentials() CredentialStore          { return errCredentials{c.err} }
func (c errController) SoftwareUpdates() SoftwareUpdateSource { return errUpdates{c.err} }

func (c errController) Decommission(context.Context, uint64) error   { return c.err }
func (c errController) SetFabricLabel(context.Context, string) error { return c.err }
func (c errController) NetworkTopology(context.Context, bool) (NetworkTopology, error) {
	return NetworkTopology{}, c.err
}

type errNode struct {
	ControllerNode
	err error
}

func (n errNode) EndpointDetail(uint16) (EndpointDetail, error) { return EndpointDetail{}, n.err }
func (n errNode) ClusterDetail(uint16, uint32) (ClusterDetail, error) {
	return ClusterDetail{}, n.err
}

func (n errNode) Read(context.Context, []AttributePath, bool) ([]AttributeValue, error) {
	return nil, n.err
}

func (n errNode) Write(context.Context, AttributePath, any, time.Duration) (WriteStatus, error) {
	return WriteStatus{}, n.err
}

func (n errNode) Invoke(context.Context, CommandPath, json.RawMessage, time.Duration) (InvokeResult, error) {
	return InvokeResult{}, n.err
}

func (n errNode) OpenCommissioningWindow(context.Context, time.Duration) (WindowInfo, error) {
	return WindowInfo{}, n.err
}
func (n errNode) RevokeCommissioningWindow(context.Context) error { return n.err }
func (n errNode) Interview(context.Context) error                 { return n.err }
func (n errNode) Ping(context.Context) (PingResult, error)        { return PingResult{}, n.err }
func (n errNode) Enable(context.Context) error                    { return n.err }
func (n errNode) Disable(context.Context) error                   { return n.err }
func (n errNode) SetAutoSubscribe(context.Context, bool) error    { return n.err }
func (n errNode) RegisterICD(context.Context, bool) error         { return n.err }
func (n errNode) UnregisterICD(context.Context, bool) error       { return n.err }

// ErrNotImplemented and ErrNotRead are 501 and 503 only on the operations
// that declare them; from any other route a source's error of either
// kind is 500, as any error the handler has no kind for (decision 79,
// UR-P4-api-267), and every controller operation declares 500 (contract
// 0.12.0, decision 82). The requests go through the full handler, so
// TestMain's status guard checks every answer here too.
func TestController_NotImplementedAndNotReadOnlyWhereDeclared(t *testing.T) {
	routes := append([]struct{ method, path, body string }{
		{http.MethodGet, "/controller/topology", ""},
		{http.MethodGet, "/controller/discovery?filter=any&timeout=1", ""},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":5}`},
		{http.MethodPut, "/controller/fabric/label", `{"label":"x"}`},
		{http.MethodPost, "/controller/nodes/{id}/ping", ""},
	}, controllerRoutes...)
	for _, c := range []struct {
		err      error
		status   int
		declared string // the one route that declares status
	}{
		{ErrNotImplemented, http.StatusNotImplemented, "/controller/topology"},
		{ErrNotRead, http.StatusServiceUnavailable, "/clusters/"},
	} {
		src := errController{NewMemoryController(), c.err}
		_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
		for _, r := range routes {
			path := strings.ReplaceAll(r.path, "{id}", "0000000000000011")
			rec := do(t, mux, r.method, path, r.body)
			declared := strings.Contains(path, c.declared)
			switch {
			case declared && rec.Code != c.status:
				t.Errorf("%v: %s %s: %d, want %d: %s", c.err, r.method, path, rec.Code, c.status, rec.Body)
			case !declared && rec.Code == c.status:
				t.Errorf("%v: %s %s: %d, undeclared there: %s", c.err, r.method, path, rec.Code, rec.Body)
			case !declared && rec.Code >= 500 && rec.Code != http.StatusInternalServerError:
				t.Errorf("%v: %s %s: %d, want 500: %s", c.err, r.method, path, rec.Code, rec.Body)
			}
		}
	}
}

// errCredentials is a host's credential store that fails every call.
type errCredentials struct{ err error }

func (c errCredentials) List(context.Context) ([]Credential, error)            { return nil, c.err }
func (c errCredentials) SetWiFi(context.Context, string, string, string) error { return c.err }
func (c errCredentials) SetThread(context.Context, string, []byte) error       { return c.err }
func (c errCredentials) Remove(context.Context, CredentialKind, string) error  { return c.err }

// errUpdates is a host's update source whose check and consent fail.
type errUpdates struct{ err error }

func (errUpdates) Queue() []PendingUpdate                             { return nil }
func (u errUpdates) Check(context.Context) ([]AvailableUpdate, error) { return nil, u.err }
func (u errUpdates) Consent(context.Context, uint64, uint16, uint16, uint32) error {
	return u.err
}

// fallibleControllerRoutes are the controller routes whose answer
// depends on a port call errController, errCredentials or errUpdates
// fails: controllerRoutes and the routes beside the nodes.
var fallibleControllerRoutes = append([]struct{ method, path, body string }{
	{http.MethodGet, "/controller/topology", ""},
	{http.MethodGet, "/controller/discovery?filter=any&timeout=1", ""},
	{http.MethodPost, "/controller/commission", `{"code":"34970112332","timeout_seconds":5}`},
	{http.MethodPost, "/controller/commission", `{"discovered":{"instance_name":"1A2B3C4D5E6F7788","long_discriminator":3840},"passcode":20202021}`},
	{http.MethodPut, "/controller/fabric/label", `{"label":"x"}`},
	{http.MethodPost, "/controller/nodes/{id}/ping", ""},
	{http.MethodGet, "/controller/credentials", ""},
	{http.MethodPut, "/controller/credentials", `{"kind":"wifi","id":"a","ssid":"b","passphrase":"secret12"}`},
	{http.MethodPut, "/controller/credentials", `{"kind":"thread","id":"t","dataset_hex":"03044d657368"}`},
	{http.MethodDelete, "/controller/credentials?kind=wifi&id=a", ""},
	{http.MethodPost, "/controller/updates/check", ""},
	{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"11","vendor_id":1,"product_id":1,"target_software_version":1}`},
}, controllerRoutes...)

// Every port kind a host's source can answer, on every fallible
// controller route, through the full handler: an operation answers a
// kind with its status only where it declares that status, and 500 with
// the cause in the log elsewhere (UR-P4-api-391, contract 0.16.0;
// pingNode declares 409 and 502). What judges each answer is TestMain's
// status guard (internal/statusguard), which fails the run on a status
// the operation does not declare; this test adds that no kind becomes a
// 5xx other than 500 or a declared 502.
func TestController_SourceKindsOnlyWhereDeclared(t *testing.T) {
	for _, kind := range []error{
		ErrNodeNotFound, ErrNoDevice, ErrUnknownCredential, ErrConflict,
		ErrICDMultiAdmin, ErrNodeFailed, context.DeadlineExceeded,
	} {
		src := errController{NewMemoryController(), kind}
		_, mux := newTestHandler(t, testOptions{noNode: true, controller: src})
		for _, r := range fallibleControllerRoutes {
			path := strings.ReplaceAll(r.path, "{id}", "0000000000000011")
			rec := do(t, mux, r.method, path, r.body)
			if rec.Code >= 500 && rec.Code != http.StatusInternalServerError && rec.Code != http.StatusBadGateway {
				t.Errorf("%v: %s %s: %d: %s", kind, r.method, path, rec.Code, rec.Body)
			}
			if rec.Code == http.StatusBadGateway && !errors.Is(kind, ErrNodeFailed) && !errors.Is(kind, context.DeadlineExceeded) {
				t.Errorf("%v: %s %s: 502: %s", kind, r.method, path, rec.Body)
			}
		}
	}
}
