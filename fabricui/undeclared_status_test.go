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
