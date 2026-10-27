// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/SukramJ/go-fabric/eligibility"

	"github.com/SukramJ/go-fabric-ui/internal/api"
)

// The cluster, attribute and command ids the typed controller routes
// address: AccessControl ACL (0x001F/0x0000), Binding (0x001E/0x0000),
// OperationalCredentials Fabrics (0x003E/0x0001), CurrentFabricIndex
// (0x003E/0x0005) and RemoveFabric (0x003E/0x0A) — the Matter 1.6.1 ids
// of go-fabric's generated cluster/spec packages, which this package does
// not import (internal/arch); the end-to-end test reads and writes them on
// a go-fabric device.
const (
	clusterAccessControl          uint32 = 0x001F
	attributeACL                  uint32 = 0x0000
	clusterBinding                uint32 = 0x001E
	attributeBinding              uint32 = 0x0000
	clusterOperationalCredentials uint32 = 0x003E
	attributeFabrics              uint32 = 0x0001
	attributeCurrentFabricIndex   uint32 = 0x0005
	commandRemoveFabric           uint32 = 0x000A
)

// maxReadPaths bounds a targeted read (decision 0.4: no full inventory).
const maxReadPaths = 32

// controllerAdmin holds the controller routes the contract marks
// x-admin, by method and path relative to the prefix
// (TestControllerAdminMatchesTheContract holds it against
// contract/openapi.yaml).
var controllerAdmin = map[string]bool{
	"DELETE /controller/nodes/{id}":                 true,
	"POST /controller/nodes/{id}/interview":         true,
	"POST /controller/nodes/{id}/enable":            true,
	"POST /controller/nodes/{id}/disable":           true,
	"POST /controller/nodes/{id}/auto-subscribe":    true,
	"POST /controller/nodes/{id}/window":            true,
	"POST /controller/nodes/{id}/window/close":      true,
	"PUT /controller/nodes/{id}/attributes":         true,
	"POST /controller/nodes/{id}/commands":          true,
	"PUT /controller/nodes/{id}/acl":                true,
	"PUT /controller/nodes/{id}/bindings":           true,
	"POST /controller/nodes/{id}/icd/register":      true,
	"POST /controller/nodes/{id}/icd/unregister":    true,
	"DELETE /controller/nodes/{id}/fabrics/{index}": true,
	"POST /controller/commission":                   true,
	"GET /controller/discovery":                     true,
	"PUT /controller/credentials":                   true,
	"DELETE /controller/credentials":                true,
	"PUT /controller/fabric/label":                  true,
	"POST /controller/updates/consent":              true,
	"POST /controller/updates/check":                true,
}

// controllerServer implements the generated strict server interface: the
// handler bodies of the controller family, each a call of the port and a
// mapping of its answer. Router, path and query binding, body decoding
// and the response types are generated from the contract
// (internal/api, make generate).
type controllerServer struct {
	h   *handler
	src ControllerSource
}

var _ api.StrictServerInterface = (*controllerServer)(nil)

// mountController registers the controller family on mux under the
// prefix: the generated routes on a mux of their own behind the admin
// gate, and a 404 problem for any other /controller path.
func (h *handler) mountController(mux *http.ServeMux) {
	srv := &controllerServer{h: h, src: h.ctrl}
	strict := api.NewStrictHandlerWithOptions(srv, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  func(w http.ResponseWriter, r *http.Request, err error) { badBody(w, r, err) },
		ResponseErrorHandlerFunc: h.controllerError,
	})
	sub := http.NewServeMux()
	api.HandlerWithOptions(strict, api.StdHTTPServerOptions{
		BaseURL:    h.prefix,
		BaseRouter: sub,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request parameter", err.Error())
		},
	})
	sub.HandleFunc(h.prefix+"/controller/", func(w http.ResponseWriter, r *http.Request) {
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
	})
	mux.Handle(h.prefix+"/controller/", h.controllerGate(sub))
}

// controllerGate refuses an admin route the request may not use before
// anything is bound or decoded, and bounds the body.
func (h *handler) controllerGate(sub *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := sub.Handler(r)
		method, path, _ := strings.Cut(pattern, " ")
		if controllerAdmin[method+" "+strings.TrimPrefix(path, h.prefix)] && !h.isAdmin(r) {
			writeProblem(w, r, http.StatusForbidden, codeForbidden, "Admin rights required",
				"this route writes, commissions or reads a secret")
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
		}
		sub.ServeHTTP(w, r)
	})
}

// apiProblem is a problem a handler body returns as its error.
type apiProblem struct {
	status      int
	code, title string
	detail      string
}

func (p *apiProblem) Error() string { return p.title + ": " + p.detail }

func badRequest(format string, args ...any) error {
	return &apiProblem{http.StatusBadRequest, codeBadRequest, "Invalid request", fmt.Sprintf(format, args...)}
}

func serviceUnready(title, detail string) error {
	return &apiProblem{http.StatusServiceUnavailable, codeServiceUnready, title, detail}
}

// controllerError writes the problem of a handler body's error: its own
// problem, a port error by its kind, 500 for anything else (logged, no
// detail).
func (h *handler) controllerError(w http.ResponseWriter, r *http.Request, err error) {
	if p, ok := errors.AsType[*apiProblem](err); ok {
		writeProblem(w, r, p.status, p.code, p.title, p.detail)
		return
	}
	switch {
	case errors.Is(err, ErrNodeNotFound):
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", err.Error())
	case errors.Is(err, ErrNoDevice):
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No device answered", err.Error())
	case errors.Is(err, ErrUnknownCredential):
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such credential", err.Error())
	case errors.Is(err, ErrInvalidArgument), errors.Is(err, ErrInvalidCredential):
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request", err.Error())
	case errors.Is(err, ErrConflict):
		writeProblem(w, r, http.StatusConflict, codeConflict, "Refused by the node", err.Error())
	case errors.Is(err, ErrNodeFailed), errors.Is(err, context.DeadlineExceeded):
		h.log.WarnContext(r.Context(), "fabricui.controller.node_failed", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		writeProblem(w, r, http.StatusBadGateway, codeNodeError, "The node failed", err.Error())
	default:
		h.serverError(w, r, "Controller failed", err)
	}
}

// --- identifiers ---

// hex16 is a 64-bit id as the contract carries it.
func hex16(v uint64) string { return fmt.Sprintf("%016X", v) }

// hexID is a cluster, attribute, command or device type id as the
// contract carries it.
func hexID(v uint32) string { return fmt.Sprintf("0x%04X", v) }

func parseHex(s string, bits int) (uint64, error) {
	t := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(s), "0x"), "0X")
	if t == "" || len(t) > bits/4 {
		return 0, fmt.Errorf("%q is not a %d-bit hex id", s, bits)
	}
	v, err := strconv.ParseUint(t, 16, bits)
	if err != nil {
		return 0, fmt.Errorf("%q is not a %d-bit hex id", s, bits)
	}
	return v, nil
}

func parseHex32(field, s string) (uint32, error) {
	v, err := parseHex(s, 32)
	if err != nil {
		return 0, badRequest("%s: %v", field, err)
	}
	return uint32(v), nil //nolint:gosec // parsed with bitSize 32
}

func parseNodeID(field, s string) (uint64, error) {
	v, err := parseHex(s, 64)
	if err != nil {
		return 0, badRequest("%s: %v", field, err)
	}
	return v, nil
}

func endpointOf(field string, v int) (uint16, error) {
	if v < 0 || v > 0xFFFF {
		return 0, badRequest("%s: %d is not an endpoint number", field, v)
	}
	return uint16(v), nil
}

// timedOf is a request's timed_ms.
func timedOf(ms *int) (time.Duration, error) {
	if ms == nil {
		return 0, nil
	}
	if *ms < 0 || *ms > 0xFFFF {
		return 0, badRequest("timed_ms: %d is not 0 to 65535", *ms)
	}
	return time.Duration(*ms) * time.Millisecond, nil
}

func (s *controllerServer) node(id string) (ControllerNode, error) {
	nid, err := parseNodeID("id", id)
	if err != nil {
		return nil, err
	}
	n, ok := s.src.Node(nid)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, hex16(nid))
	}
	return n, nil
}

func ptr[T any](v T) *T { return &v }

// optString is s, nil when empty.
func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	t = t.UTC()
	return &t
}

func optUint[T ~uint8 | ~uint16 | ~uint32](v *T) *int {
	if v == nil {
		return nil
	}
	return ptr(int(*v))
}

func nonZero[T ~uint8 | ~uint16 | ~uint32](v T) *int {
	if v == 0 {
		return nil
	}
	return ptr(int(v))
}

// --- status, nodes ---

func (s *controllerServer) GetControllerStatus(context.Context, api.GetControllerStatusRequestObject) (api.GetControllerStatusResponseObject, error) {
	f := s.src.Fabric()
	st := api.GetControllerStatus200JSONResponse{
		Enabled:               true,
		FabricIDHex:           hex16(f.FabricID),
		CompressedFabricIDHex: hex16(f.CompressedFabricID),
		FabricIndex:           int(f.FabricIndex),
		NodeIDHex:             hex16(f.NodeID),
		VendorID:              int(f.VendorID),
		VendorName:            eligibility.VendorName(f.VendorID),
		Label:                 f.Label,
		NodeCount:             len(s.src.Nodes()),
		Capabilities: api.ControllerCapabilities{
			Ota:         s.src.SoftwareUpdates() != nil,
			Credentials: s.src.Credentials() != nil,
			Ble:         s.src.Capabilities().BLE,
		},
	}
	if s.h.admin == nil {
		st.CsrfToken = ptr(s.h.token)
	}
	return st, nil
}

func (s *controllerServer) ListControllerNodes(context.Context, api.ListControllerNodesRequestObject) (api.ListControllerNodesResponseObject, error) {
	out := api.ListControllerNodes200JSONResponse{Nodes: []api.ControllerNodeSummary{}}
	for _, n := range s.src.Nodes() {
		out.Nodes = append(out.Nodes, nodeSummary(n))
	}
	return out, nil
}

func nodeSummary(n ControllerNode) api.ControllerNodeSummary {
	b, rec := n.Basic(), n.Record()
	_, icd := n.ICD()
	return api.ControllerNodeSummary{
		NodeIDHex:          hex16(n.ID()),
		VendorID:           nonZero(b.VendorID),
		VendorName:         optString(b.VendorName),
		ProductID:          nonZero(b.ProductID),
		ProductName:        optString(b.ProductName),
		NodeLabel:          optString(b.NodeLabel),
		Connection:         api.ConnectionState(n.ConnectionState()),
		SubscriptionActive: n.SubscriptionActive(),
		Disabled:           n.Disabled(),
		Icd:                icd,
		Addresses:          nonNil(rec.Addresses),
		CommissionedAt:     optTime(rec.CommissionedAt),
	}
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

func nodeDetail(n ControllerNode) api.ControllerNodeDetail {
	rec, p, b := n.Record(), n.Properties(), n.Basic()
	d := api.ControllerNodeDetail{
		Node: nodeSummary(n),
		Record: api.NodeRecord{
			FabricIndex:         int(rec.FabricIndex),
			FabricIndexOnPeer:   nonZero(rec.FabricIndexOnPeer),
			Addresses:           nonNil(rec.Addresses),
			Hostname:            optString(rec.Hostname),
			DiscoveredAt:        optTime(rec.DiscoveredAt),
			CommissionedAt:      optTime(rec.CommissionedAt),
			Discriminator:       optUint(rec.Discriminator),
			AnnouncedVendorID:   optUint(rec.VendorID),
			AnnouncedProductID:  optUint(rec.ProductID),
			AnnouncedDeviceType: optUint(rec.DeviceType),
			DeviceName:          optString(rec.DeviceName),
		},
		Properties: api.NodeProperties{
			SupportsThread:          p.SupportsThread,
			SupportsWifi:            p.SupportsWiFi,
			SupportsEthernet:        p.SupportsEthernet,
			MainsPowered:            p.MainsPowered,
			BatteryPowered:          p.BatteryPowered,
			IntermittentlyConnected: p.IntermittentlyConnected,
			SpecificationVersion:    nonZero(p.SpecificationVersion),
			SerialNumber:            optString(b.SerialNumber),
			SoftwareVersion:         optString(b.SoftwareVersion),
			DeviceTypes:             []int{},
		},
		Endpoints: []api.NodeEndpoint{},
	}
	if p.LongIdleTimeOperating {
		d.Properties.LongIdleTimeOperating = ptr(true)
	}
	if p.ThreadSleepyEndDevice {
		d.Properties.ThreadSleepyEndDevice = ptr(true)
	}
	if p.IdleModeDuration > 0 {
		d.Properties.IdleModeDurationSeconds = ptr(int(p.IdleModeDuration / time.Second))
	}
	for _, t := range p.DeviceTypes {
		d.Properties.DeviceTypes = append(d.Properties.DeviceTypes, int(t))
	}
	for _, ep := range n.Endpoints() {
		e := api.NodeEndpoint{Endpoint: int(ep.Number), Name: optString(ep.Name), DeviceTypes: []api.NodeDeviceType{}, Clusters: []api.NodeCluster{}}
		if ep.Parent != nil {
			e.Parent = ptr(int(*ep.Parent))
		}
		for _, dt := range ep.DeviceTypes {
			e.DeviceTypes = append(e.DeviceTypes, api.NodeDeviceType{ID: int(dt.ID), IDHex: hexID(dt.ID), Name: dt.Name, Revision: nonZero(dt.Revision)})
		}
		for _, c := range ep.Clusters {
			e.Clusters = append(e.Clusters, api.NodeCluster{ID: int(c.ID), IDHex: hexID(c.ID), Name: c.Name, Revision: int(c.Revision), Features: nonNil(c.Features)})
		}
		d.Endpoints = append(d.Endpoints, e)
	}
	return d
}

func (s *controllerServer) GetControllerNode(_ context.Context, req api.GetControllerNodeRequestObject) (api.GetControllerNodeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	return api.GetControllerNode200JSONResponse(nodeDetail(n)), nil
}

func (s *controllerServer) DecommissionNode(ctx context.Context, req api.DecommissionNodeRequestObject) (api.DecommissionNodeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if err := s.src.Decommission(ctx, n.ID()); err != nil {
		return nil, err
	}
	return api.DecommissionNode204Response{}, nil
}

func (s *controllerServer) InterviewNode(ctx context.Context, req api.InterviewNodeRequestObject) (api.InterviewNodeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if err := n.Interview(ctx); err != nil {
		return nil, err
	}
	return api.InterviewNode204Response{}, nil
}

func (s *controllerServer) EnableNode(ctx context.Context, req api.EnableNodeRequestObject) (api.EnableNodeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if err := n.Enable(ctx); err != nil {
		return nil, err
	}
	return api.EnableNode204Response{}, nil
}

func (s *controllerServer) DisableNode(ctx context.Context, req api.DisableNodeRequestObject) (api.DisableNodeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if err := n.Disable(ctx); err != nil {
		return nil, err
	}
	return api.DisableNode204Response{}, nil
}

func (s *controllerServer) SetNodeAutoSubscribe(ctx context.Context, req api.SetNodeAutoSubscribeRequestObject) (api.SetNodeAutoSubscribeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, badRequest("the body {enabled} is required")
	}
	if err := n.SetAutoSubscribe(ctx, req.Body.Enabled); err != nil {
		return nil, err
	}
	return api.SetNodeAutoSubscribe204Response{}, nil
}

// --- window ---

func (s *controllerServer) OpenNodeWindow(ctx context.Context, req api.OpenNodeWindowRequestObject) (api.OpenNodeWindowResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	var timeout time.Duration
	if req.Body != nil && req.Body.TimeoutSeconds != nil {
		t := *req.Body.TimeoutSeconds
		if t < 0 || t > 0xFFFF {
			return nil, badRequest("timeout_seconds: %d is not 0 to 65535", t)
		}
		timeout = time.Duration(t) * time.Second
	}
	w, err := n.OpenCommissioningWindow(ctx, timeout)
	if err != nil {
		return nil, err
	}
	// The answer carries the passcode; it is not logged.
	return api.OpenNodeWindow200JSONResponse{
		ManualCode: w.ManualCode, QrCode: w.QRCode, Passcode: int(w.Passcode),
		Discriminator: int(w.Discriminator), VendorID: int(w.VendorID), ProductID: int(w.ProductID),
		TimeoutSeconds: int(w.Timeout / time.Second),
	}, nil
}

func (s *controllerServer) CloseNodeWindow(ctx context.Context, req api.CloseNodeWindowRequestObject) (api.CloseNodeWindowResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if err := n.RevokeCommissioningWindow(ctx); err != nil {
		return nil, err
	}
	return api.CloseNodeWindow204Response{}, nil
}

// --- attributes, commands ---

func attributePathOf(field string, endpoint int, cluster, attribute string) (AttributePath, error) {
	ep, err := endpointOf(field+".endpoint", endpoint)
	if err != nil {
		return AttributePath{}, err
	}
	c, err := parseHex32(field+".cluster_hex", cluster)
	if err != nil {
		return AttributePath{}, err
	}
	a, err := parseHex32(field+".attribute_hex", attribute)
	if err != nil {
		return AttributePath{}, err
	}
	return AttributePath{Endpoint: ep, Cluster: c, Attribute: a}, nil
}

func (s *controllerServer) ReadNodeAttributes(ctx context.Context, req api.ReadNodeAttributesRequestObject) (api.ReadNodeAttributesResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if req.Body == nil || len(req.Body.Paths) == 0 {
		return nil, badRequest("paths: at least one path is required")
	}
	if len(req.Body.Paths) > maxReadPaths {
		return nil, badRequest("paths: %d paths, at most %d", len(req.Body.Paths), maxReadPaths)
	}
	paths := make([]AttributePath, 0, len(req.Body.Paths))
	for i, p := range req.Body.Paths {
		ap, err := attributePathOf(fmt.Sprintf("paths[%d]", i), p.Endpoint, p.ClusterHex, p.AttributeHex)
		if err != nil {
			return nil, err
		}
		paths = append(paths, ap)
	}
	filtered := req.Body.FabricFiltered == nil || *req.Body.FabricFiltered
	values, err := n.Read(ctx, paths, filtered)
	if err != nil {
		return nil, err
	}
	out := api.ReadNodeAttributes200JSONResponse{Values: []api.AttributeValue{}}
	for _, v := range values {
		e := api.AttributeValue{Endpoint: int(v.Path.Endpoint), ClusterHex: hexID(v.Path.Cluster), AttributeHex: hexID(v.Path.Attribute)}
		if v.Status != 0 {
			e.Status = ptr(int(v.Status))
		} else {
			e.Value = ptr(v.Value)
		}
		out.Values = append(out.Values, e)
	}
	return out, nil
}

func writeResult(st WriteStatus) api.WriteResult {
	r := api.WriteResult{Status: int(st.Status)}
	if st.ClusterStatus != nil {
		r.ClusterStatus = ptr(int(*st.ClusterStatus))
	}
	return r
}

func (s *controllerServer) WriteNodeAttribute(ctx context.Context, req api.WriteNodeAttributeRequestObject) (api.WriteNodeAttributeResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	b := req.Body
	if b == nil || len(b.Value) == 0 {
		return nil, badRequest("value is required")
	}
	path, err := attributePathOf("path", b.Endpoint, b.ClusterHex, b.AttributeHex)
	if err != nil {
		return nil, err
	}
	timed, err := timedOf(b.TimedMs)
	if err != nil {
		return nil, err
	}
	st, err := n.Write(ctx, path, b.Value, timed)
	if err != nil {
		return nil, err
	}
	return api.WriteNodeAttribute200JSONResponse(writeResult(st)), nil
}

func (s *controllerServer) InvokeNodeCommand(ctx context.Context, req api.InvokeNodeCommandRequestObject) (api.InvokeNodeCommandResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	b := req.Body
	if b == nil {
		return nil, badRequest("the body {endpoint, cluster_hex, command_hex} is required")
	}
	ep, err := endpointOf("endpoint", b.Endpoint)
	if err != nil {
		return nil, err
	}
	c, err := parseHex32("cluster_hex", b.ClusterHex)
	if err != nil {
		return nil, err
	}
	cmd, err := parseHex32("command_hex", b.CommandHex)
	if err != nil {
		return nil, err
	}
	timed, err := timedOf(b.TimedMs)
	if err != nil {
		return nil, err
	}
	var payload json.RawMessage
	if b.Payload != nil {
		payload = *b.Payload
	}
	res, err := n.Invoke(ctx, CommandPath{Endpoint: ep, Cluster: c, Command: cmd}, payload, timed)
	if err != nil {
		return nil, err
	}
	out := api.InvokeNodeCommand200JSONResponse{Status: int(res.Status)}
	if res.ClusterStatus != nil {
		out.ClusterStatus = ptr(int(*res.ClusterStatus))
	}
	if res.HasResponse {
		out.ResponseCommandHex = ptr(hexID(res.Command))
		if len(res.Response) > 0 {
			out.Response = ptr(res.Response)
		}
	}
	return out, nil
}

// readOne reads one attribute and answers its JSON; a status the node
// answers for it is a node error.
func readOne(ctx context.Context, n ControllerNode, path AttributePath, filtered bool) (json.RawMessage, error) {
	values, err := n.Read(ctx, []AttributePath{path}, filtered)
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		if v.Path != path {
			continue
		}
		if v.Status != 0 {
			return nil, fmt.Errorf("%w: status %#x for %s/%s/%s", ErrNodeFailed, v.Status, strconv.Itoa(int(path.Endpoint)), hexID(path.Cluster), hexID(path.Attribute))
		}
		return v.Value, nil
	}
	return nil, fmt.Errorf("%w: no value for %s/%s/%s", ErrNodeFailed, strconv.Itoa(int(path.Endpoint)), hexID(path.Cluster), hexID(path.Attribute))
}

// --- ACL ---

var (
	privilegeNames = map[uint8]api.ACLEntryPrivilege{
		1: api.ACLEntryPrivilegeView, 2: api.ACLEntryPrivilegeProxyView, 3: api.ACLEntryPrivilegeOperate,
		4: api.ACLEntryPrivilegeManage, 5: api.ACLEntryPrivilegeAdminister,
	}
	authModeNames = map[uint8]api.ACLEntryAuthMode{1: api.ACLEntryAuthModePase, 2: api.ACLEntryAuthModeCase, 3: api.ACLEntryAuthModeGroup}
)

func keyOf[K, V comparable](m map[K]V, v V) (K, bool) {
	for k, x := range m {
		if x == v {
			return k, true
		}
	}
	var zero K
	return zero, false
}

// aclJSON is an AccessControlEntryStruct as the read renders it, by field
// name.
type aclJSON struct {
	Privilege uint8    `json:"privilege"`
	AuthMode  uint8    `json:"authMode"`
	Subjects  []uint64 `json:"subjects"`
	Targets   []struct {
		Cluster    *uint32 `json:"cluster"`
		Endpoint   *uint16 `json:"endpoint"`
		DeviceType *uint32 `json:"deviceType"`
	} `json:"targets"`
	FabricIndex uint8 `json:"fabricIndex"`
}

func (s *controllerServer) GetNodeACL(ctx context.Context, req api.GetNodeACLRequestObject) (api.GetNodeACLResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	raw, err := readOne(ctx, n, AttributePath{Endpoint: 0, Cluster: clusterAccessControl, Attribute: attributeACL}, true)
	if err != nil {
		return nil, err
	}
	var entries []aclJSON
	if err := json.Unmarshal(raw, &entries); err != nil {
		return nil, fmt.Errorf("%w: the ACL does not read as a list of entries: %w", ErrNodeFailed, err)
	}
	out := api.GetNodeACL200JSONResponse{Entries: []api.ACLEntry{}}
	for _, e := range entries {
		a := api.ACLEntry{
			Privilege: privilegeNames[e.Privilege], AuthMode: authModeNames[e.AuthMode],
			SubjectsHex: []string{}, Targets: []api.ACLTarget{}, FabricIndex: nonZero(e.FabricIndex),
		}
		for _, sub := range e.Subjects {
			a.SubjectsHex = append(a.SubjectsHex, hex16(sub))
		}
		for _, t := range e.Targets {
			var at api.ACLTarget
			if t.Cluster != nil {
				at.ClusterHex = ptr(hexID(*t.Cluster))
			}
			if t.Endpoint != nil {
				at.Endpoint = ptr(int(*t.Endpoint))
			}
			if t.DeviceType != nil {
				at.DeviceTypeHex = ptr(hexID(*t.DeviceType))
			}
			a.Targets = append(a.Targets, at)
		}
		out.Entries = append(out.Entries, a)
	}
	return out, nil
}

// SetNodeACL writes our fabric's entries in one write of the ACL. A
// subject naming the node itself is left out, and an entry left without
// subjects dropped — a node always has access to itself — as
// go-fabric-server's set_acl_entry does after matterjs-server
// (ControllerCommandHandler.ts:2618-2668).
func (s *controllerServer) SetNodeACL(ctx context.Context, req api.SetNodeACLRequestObject) (api.SetNodeACLResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, badRequest("the body {entries} is required")
	}
	entries := make([]ACLEntry, 0, len(req.Body.Entries))
	for i, e := range req.Body.Entries {
		priv, ok := keyOf(privilegeNames, e.Privilege)
		if !ok {
			return nil, badRequest("entries[%d].privilege: %q is unknown", i, e.Privilege)
		}
		mode, ok := keyOf(authModeNames, e.AuthMode)
		if !ok {
			return nil, badRequest("entries[%d].auth_mode: %q is unknown", i, e.AuthMode)
		}
		entry := ACLEntry{Privilege: priv, AuthMode: mode}
		for j, sh := range e.SubjectsHex {
			sub, err := parseNodeID(fmt.Sprintf("entries[%d].subjects_hex[%d]", i, j), sh)
			if err != nil {
				return nil, err
			}
			if sub != n.ID() {
				entry.Subjects = append(entry.Subjects, sub)
			}
		}
		if len(e.SubjectsHex) > 0 && len(entry.Subjects) == 0 {
			continue
		}
		for j, t := range e.Targets {
			var at ACLTarget
			if t.ClusterHex != nil {
				c, err := parseHex32(fmt.Sprintf("entries[%d].targets[%d].cluster_hex", i, j), *t.ClusterHex)
				if err != nil {
					return nil, err
				}
				at.Cluster = &c
			}
			if t.Endpoint != nil {
				ep, err := endpointOf(fmt.Sprintf("entries[%d].targets[%d].endpoint", i, j), *t.Endpoint)
				if err != nil {
					return nil, err
				}
				at.Endpoint = &ep
			}
			if t.DeviceTypeHex != nil {
				d, err := parseHex32(fmt.Sprintf("entries[%d].targets[%d].device_type_hex", i, j), *t.DeviceTypeHex)
				if err != nil {
					return nil, err
				}
				at.DeviceType = &d
			}
			entry.Targets = append(entry.Targets, at)
		}
		entries = append(entries, entry)
	}
	st, err := n.Write(ctx, AttributePath{Endpoint: 0, Cluster: clusterAccessControl, Attribute: attributeACL}, entries, 0)
	if err != nil {
		return nil, err
	}
	return api.SetNodeACL200JSONResponse(writeResult(st)), nil
}

// --- bindings ---

// bindingJSON is a Binding TargetStruct as the read renders it.
type bindingJSON struct {
	Node        *uint64 `json:"node"`
	Group       *uint16 `json:"group"`
	Endpoint    *uint16 `json:"endpoint"`
	Cluster     *uint32 `json:"cluster"`
	FabricIndex uint8   `json:"fabricIndex"`
}

func (s *controllerServer) GetNodeBindings(ctx context.Context, req api.GetNodeBindingsRequestObject) (api.GetNodeBindingsResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	var paths []AttributePath
	for _, ep := range n.Endpoints() {
		for _, c := range ep.Clusters {
			if c.ID == clusterBinding {
				paths = append(paths, AttributePath{Endpoint: ep.Number, Cluster: clusterBinding, Attribute: attributeBinding})
			}
		}
	}
	out := api.GetNodeBindings200JSONResponse{Endpoints: []api.EndpointBindings{}}
	if len(paths) == 0 {
		return out, nil
	}
	values, err := n.Read(ctx, paths, true)
	if err != nil {
		return nil, err
	}
	for _, v := range values {
		if v.Status != 0 {
			continue
		}
		var targets []bindingJSON
		if err := json.Unmarshal(v.Value, &targets); err != nil {
			return nil, fmt.Errorf("%w: endpoint %d's bindings do not read as a list: %w", ErrNodeFailed, v.Path.Endpoint, err)
		}
		eb := api.EndpointBindings{Endpoint: int(v.Path.Endpoint), Bindings: []api.BindingTarget{}}
		for _, t := range targets {
			var b api.BindingTarget
			if t.Node != nil {
				b.NodeIDHex = ptr(hex16(*t.Node))
			}
			b.Group = optUint(t.Group)
			b.Endpoint = optUint(t.Endpoint)
			if t.Cluster != nil {
				b.ClusterHex = ptr(hexID(*t.Cluster))
			}
			b.FabricIndex = nonZero(t.FabricIndex)
			eb.Bindings = append(eb.Bindings, b)
		}
		out.Endpoints = append(out.Endpoints, eb)
	}
	return out, nil
}

func (s *controllerServer) SetNodeBindings(ctx context.Context, req api.SetNodeBindingsRequestObject) (api.SetNodeBindingsResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, badRequest("the body {endpoint, bindings} is required")
	}
	ep, err := endpointOf("endpoint", req.Body.Endpoint)
	if err != nil {
		return nil, err
	}
	targets := make([]BindingTarget, 0, len(req.Body.Bindings))
	for i, b := range req.Body.Bindings {
		var t BindingTarget
		if b.NodeIDHex != nil {
			id, err := parseNodeID(fmt.Sprintf("bindings[%d].node_id_hex", i), *b.NodeIDHex)
			if err != nil {
				return nil, err
			}
			t.Node = &id
		}
		if b.Group != nil {
			if *b.Group < 1 || *b.Group > 0xFFFF {
				return nil, badRequest("bindings[%d].group: %d is not a group id", i, *b.Group)
			}
			g := uint16(*b.Group)
			t.Group = &g
		}
		if b.Endpoint != nil {
			e, err := endpointOf(fmt.Sprintf("bindings[%d].endpoint", i), *b.Endpoint)
			if err != nil {
				return nil, err
			}
			t.Endpoint = &e
		}
		if b.ClusterHex != nil {
			c, err := parseHex32(fmt.Sprintf("bindings[%d].cluster_hex", i), *b.ClusterHex)
			if err != nil {
				return nil, err
			}
			t.Cluster = &c
		}
		switch {
		case t.Group != nil && (t.Node != nil || t.Endpoint != nil):
			return nil, badRequest("bindings[%d]: a group target names no node and no endpoint", i)
		case t.Group == nil && (t.Node == nil || t.Endpoint == nil):
			return nil, badRequest("bindings[%d]: a unicast target names node_id_hex and endpoint", i)
		}
		targets = append(targets, t)
	}
	st, err := n.Write(ctx, AttributePath{Endpoint: ep, Cluster: clusterBinding, Attribute: attributeBinding}, targets, 0)
	if err != nil {
		return nil, err
	}
	return api.SetNodeBindings200JSONResponse(writeResult(st)), nil
}

// --- ICD ---

func icdState(n ControllerNode) api.IcdState {
	st, ok := n.ICD()
	if !ok {
		return api.IcdState{}
	}
	out := api.IcdState{
		Supported: true, LitSupported: st.LITSupported, Registered: st.Registered,
		Awake: ptr(st.Awake), Available: ptr(st.Available),
		NextCheckIn: optTime(st.NextCheckIn), LastCheckIn: optTime(st.LastCheckIn),
	}
	switch st.OperatingMode {
	case "sit":
		out.OperatingMode = ptr(api.IcdStateOperatingModeSit)
	case "lit":
		out.OperatingMode = ptr(api.IcdStateOperatingModeLit)
	}
	return out
}

func (s *controllerServer) GetNodeIcd(_ context.Context, req api.GetNodeIcdRequestObject) (api.GetNodeIcdResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	return api.GetNodeIcd200JSONResponse(icdState(n)), nil
}

func (s *controllerServer) RegisterNodeIcd(ctx context.Context, req api.RegisterNodeIcdRequestObject) (api.RegisterNodeIcdResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if _, ok := n.ICD(); !ok {
		return nil, badRequest("the node has no ICD Management")
	}
	multi := req.Body != nil && req.Body.AllowMultiAdmin != nil && *req.Body.AllowMultiAdmin
	if err := n.RegisterICD(ctx, multi); err != nil {
		return nil, err
	}
	return api.RegisterNodeIcd200JSONResponse(icdState(n)), nil
}

func (s *controllerServer) UnregisterNodeIcd(ctx context.Context, req api.UnregisterNodeIcdRequestObject) (api.UnregisterNodeIcdResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if _, ok := n.ICD(); !ok {
		return nil, badRequest("the node has no ICD Management")
	}
	force := req.Body != nil && req.Body.Force != nil && *req.Body.Force
	if err := n.UnregisterICD(ctx, force); err != nil {
		return nil, err
	}
	return api.UnregisterNodeIcd200JSONResponse(icdState(n)), nil
}

// --- fabrics on the node ---

// fabricJSON is a FabricDescriptorStruct as the read renders it.
type fabricJSON struct {
	VendorID    uint16 `json:"vendorId"`
	FabricID    uint64 `json:"fabricId"`
	NodeID      uint64 `json:"nodeId"`
	Label       string `json:"label"`
	FabricIndex uint8  `json:"fabricIndex"`
}

func (s *controllerServer) ownIndex(ctx context.Context, n ControllerNode) (uint8, error) {
	raw, err := readOne(ctx, n, AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeCurrentFabricIndex}, true)
	if err != nil {
		return 0, err
	}
	var idx uint8
	if err := json.Unmarshal(raw, &idx); err != nil {
		return 0, fmt.Errorf("%w: CurrentFabricIndex: %w", ErrNodeFailed, err)
	}
	return idx, nil
}

func (s *controllerServer) ListNodeFabrics(ctx context.Context, req api.ListNodeFabricsRequestObject) (api.ListNodeFabricsResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	raw, err := readOne(ctx, n, AttributePath{Endpoint: 0, Cluster: clusterOperationalCredentials, Attribute: attributeFabrics}, false)
	if err != nil {
		return nil, err
	}
	own, err := s.ownIndex(ctx, n)
	if err != nil {
		return nil, err
	}
	var fabrics []fabricJSON
	if err := json.Unmarshal(raw, &fabrics); err != nil {
		return nil, fmt.Errorf("%w: the fabrics do not read as a list: %w", ErrNodeFailed, err)
	}
	out := api.ListNodeFabrics200JSONResponse{Fabrics: []api.NodeFabric{}}
	for _, f := range fabrics {
		out.Fabrics = append(out.Fabrics, api.NodeFabric{
			FabricIndex: int(f.FabricIndex), FabricIDHex: hex16(f.FabricID), NodeIDHex: hex16(f.NodeID),
			VendorID: int(f.VendorID), VendorName: eligibility.VendorName(f.VendorID), Label: f.Label,
			Own: f.FabricIndex == own,
		})
	}
	return out, nil
}

func (s *controllerServer) RemoveNodeFabric(ctx context.Context, req api.RemoveNodeFabricRequestObject) (api.RemoveNodeFabricResponseObject, error) {
	n, err := s.node(req.ID)
	if err != nil {
		return nil, err
	}
	if req.Index < 1 || req.Index > 254 {
		return nil, badRequest("index: %d is not a fabric index (1 to 254)", req.Index)
	}
	own, err := s.ownIndex(ctx, n)
	if err != nil {
		return nil, err
	}
	if req.Index == int(own) {
		return nil, badRequest("fabric %d is this controller's; decommission the node instead", req.Index)
	}
	payload := json.RawMessage(`{"fabricIndex":` + strconv.Itoa(req.Index) + `}`)
	res, err := n.Invoke(ctx, CommandPath{Endpoint: 0, Cluster: clusterOperationalCredentials, Command: commandRemoveFabric}, payload, 0)
	if err != nil {
		return nil, err
	}
	var noc struct {
		StatusCode uint8 `json:"statusCode"`
	}
	if res.Status != 0 {
		return nil, fmt.Errorf("%w: RemoveFabric answered status %#x", ErrNodeFailed, res.Status)
	}
	if res.HasResponse && json.Unmarshal(res.Response, &noc) == nil && noc.StatusCode != 0 {
		return nil, fmt.Errorf("%w: RemoveFabric answered NOCResponse status %d", ErrNodeFailed, noc.StatusCode)
	}
	return api.RemoveNodeFabric204Response{}, nil
}

// --- commissioning, discovery ---

func pairOptionsOf(b *api.CommissionRequest) PairOptions {
	var o PairOptions
	if b.WifiID != nil {
		o.WiFiID = *b.WifiID
	}
	if b.ThreadID != nil {
		o.ThreadID = *b.ThreadID
	}
	o.NetworkOnly = b.NetworkOnly != nil && *b.NetworkOnly
	return o
}

// CommissionNode commissions from a code, or a discovered device and its
// passcode, and reports the commissioning on the stream. Neither the code
// nor the passcode reaches a log line or an event.
func (s *controllerServer) CommissionNode(ctx context.Context, req api.CommissionNodeRequestObject) (api.CommissionNodeResponseObject, error) {
	b := req.Body
	if b == nil {
		return nil, badRequest("the body {code} or {discovered, passcode} is required")
	}
	hasCode := b.Code != nil && strings.TrimSpace(*b.Code) != ""
	switch {
	case hasCode && b.Discovered != nil:
		return nil, badRequest("code and discovered name two devices; send one")
	case !hasCode && b.Discovered == nil:
		return nil, badRequest("code or discovered is required")
	case b.Discovered != nil && b.Passcode == nil:
		return nil, badRequest("discovered needs passcode")
	case b.Passcode != nil && (*b.Passcode < 1 || *b.Passcode > 99999998):
		return nil, badRequest("passcode: not 1 to 99999998")
	}
	if b.TimeoutSeconds != nil {
		if *b.TimeoutSeconds < 1 || *b.TimeoutSeconds > 600 {
			return nil, badRequest("timeout_seconds: %d is not 1 to 600", *b.TimeoutSeconds)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*b.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	opts := pairOptionsOf(b)
	s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("started")}})
	var n ControllerNode
	var err error
	if hasCode {
		n, err = s.src.Pair(ctx, strings.TrimSpace(*b.Code), opts)
	} else {
		d := Discovered{InstanceName: b.Discovered.InstanceName}
		if b.Discovered.LongDiscriminator != nil {
			d.LongDiscriminator = uint16(*b.Discovered.LongDiscriminator) //nolint:gosec // a 12-bit discriminator
		}
		n, err = s.src.CommissionDiscovered(ctx, d, uint32(*b.Passcode), opts) //nolint:gosec // bounded above
	}
	if err != nil {
		s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("failed"), Detail: ptr(progressDetail(err))}})
		return nil, err
	}
	s.h.ctrlHub.Publish(Event{Type: EventCommissioningProgress, Data: controllerEventData{Phase: ptr("succeeded"), NodeIDHex: ptr(hex16(n.ID()))}})
	return api.CommissionNode200JSONResponse(nodeDetail(n)), nil
}

// progressDetail is a failed commissioning's kind for the stream — the
// kind only, so no error text that might echo a code travels in an event.
func progressDetail(err error) string {
	switch {
	case errors.Is(err, ErrInvalidArgument):
		return "invalid_argument"
	case errors.Is(err, ErrNoDevice):
		return "no_device"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, ErrNodeFailed):
		return "node_failed"
	default:
		return "failed"
	}
}

func (s *controllerServer) DiscoverDevices(ctx context.Context, req api.DiscoverDevicesRequestObject) (api.DiscoverDevicesResponseObject, error) {
	f := DiscoverFilter{Kind: "any"}
	if req.Params.Filter != nil {
		f.Kind = string(*req.Params.Filter)
		if !req.Params.Filter.Valid() {
			return nil, badRequest("filter: %q is unknown", f.Kind)
		}
	}
	if f.Kind != "any" {
		if req.Params.Value == nil || *req.Params.Value < 0 || int64(*req.Params.Value) > 0xFFFFFFFF { // int64: int is 32 bits on arm/386
			return nil, badRequest("value: filter %s needs a value", f.Kind)
		}
		f.Value = uint32(*req.Params.Value)
	}
	timeout := 3 * time.Second
	if req.Params.Timeout != nil {
		if *req.Params.Timeout < 1 || *req.Params.Timeout > 60 {
			return nil, badRequest("timeout: %d is not 1 to 60", *req.Params.Timeout)
		}
		timeout = time.Duration(*req.Params.Timeout) * time.Second
	}
	found, err := s.src.Discover(ctx, f, timeout)
	if err != nil {
		return nil, err
	}
	out := api.DiscoverDevices200JSONResponse{Devices: []api.Discovered{}}
	for _, d := range found {
		e := api.Discovered{
			InstanceName: d.InstanceName, Hostname: optString(d.Hostname), Addresses: nonNil(d.Addresses),
			LongDiscriminator: int(d.LongDiscriminator), CommissioningMode: int(d.CommissioningMode),
			VendorID: nonZero(d.VendorID), ProductID: nonZero(d.ProductID), DeviceType: nonZero(d.DeviceType),
			DeviceName: optString(d.DeviceName), PairingHint: nonZero(d.PairingHint),
			PairingInstruction: optString(d.PairingInstruction), Port: nonZero(d.Port),
		}
		if d.VendorID != 0 {
			e.VendorName = optString(eligibility.VendorName(d.VendorID))
		}
		out.Devices = append(out.Devices, e)
	}
	return out, nil
}

// --- credentials ---

func (s *controllerServer) credentials() (CredentialStore, error) {
	c := s.src.Credentials()
	if c == nil {
		return nil, serviceUnready("No credential store", "the host stores no Wi-Fi or Thread credentials")
	}
	return c, nil
}

func (s *controllerServer) ListCredentials(ctx context.Context, _ api.ListCredentialsRequestObject) (api.ListCredentialsResponseObject, error) {
	c, err := s.credentials()
	if err != nil {
		return nil, err
	}
	list, err := c.List(ctx)
	if err != nil {
		return nil, err
	}
	out := api.ListCredentials200JSONResponse{Credentials: []api.Credential{}}
	for _, e := range list {
		out.Credentials = append(out.Credentials, api.Credential{Kind: api.CredentialKind(e.Kind), ID: e.ID, Name: e.Name})
	}
	return out, nil
}

// SetCredential stores a credential; the passphrase and the dataset are
// not logged and never read back.
func (s *controllerServer) SetCredential(ctx context.Context, req api.SetCredentialRequestObject) (api.SetCredentialResponseObject, error) {
	c, err := s.credentials()
	if err != nil {
		return nil, err
	}
	b := req.Body
	if b == nil {
		return nil, badRequest("the body {kind, id, …} is required")
	}
	if b.ID == "" || utf8.RuneCountInString(b.ID) > 64 {
		return nil, badRequest("id: 1 to 64 characters")
	}
	switch b.Kind {
	case api.CredentialRequestKindWifi:
		if b.Ssid == nil || *b.Ssid == "" {
			return nil, badRequest("ssid is required for a Wi-Fi credential")
		}
		pass := ""
		if b.Passphrase != nil {
			pass = *b.Passphrase
		}
		err = c.SetWiFi(ctx, b.ID, *b.Ssid, pass)
	case api.CredentialRequestKindThread:
		if b.DatasetHex == nil || *b.DatasetHex == "" {
			return nil, badRequest("dataset_hex is required for a Thread dataset")
		}
		ds, herr := hex.DecodeString(*b.DatasetHex)
		if herr != nil {
			return nil, badRequest("dataset_hex is not hex")
		}
		err = c.SetThread(ctx, b.ID, ds)
	default:
		return nil, badRequest("kind: %q is not wifi or thread", b.Kind)
	}
	if err != nil {
		return nil, err
	}
	return api.SetCredential204Response{}, nil
}

func (s *controllerServer) RemoveCredential(ctx context.Context, req api.RemoveCredentialRequestObject) (api.RemoveCredentialResponseObject, error) {
	c, err := s.credentials()
	if err != nil {
		return nil, err
	}
	if !req.Params.Kind.Valid() {
		return nil, badRequest("kind: %q is not wifi or thread", req.Params.Kind)
	}
	if err := c.Remove(ctx, CredentialKind(req.Params.Kind), req.Params.ID); err != nil {
		return nil, err
	}
	return api.RemoveCredential204Response{}, nil
}

// --- fabric label ---

// maxFabricLabel is the longest fabric label in characters (matter.js
// Fabric.setLabel, packages/protocol/src/fabric/Fabric.ts:193-196 at
// 862ad75f, as go-fabric's commissioner.MaxLabelLength).
const maxFabricLabel = 32

func (s *controllerServer) SetFabricLabel(ctx context.Context, req api.SetFabricLabelRequestObject) (api.SetFabricLabelResponseObject, error) {
	if req.Body == nil {
		return nil, badRequest("the body {label} is required")
	}
	if utf8.RuneCountInString(req.Body.Label) > maxFabricLabel {
		return nil, badRequest("label: longer than %d characters", maxFabricLabel)
	}
	if err := s.src.SetFabricLabel(ctx, req.Body.Label); err != nil {
		return nil, err
	}
	return api.SetFabricLabel204Response{}, nil
}

// --- updates ---

func (s *controllerServer) ListUpdates(context.Context, api.ListUpdatesRequestObject) (api.ListUpdatesResponseObject, error) {
	u := s.src.SoftwareUpdates()
	out := api.ListUpdates200JSONResponse{Available: u != nil, Queue: []api.PendingUpdate{}}
	if u == nil {
		return out, nil
	}
	for _, p := range u.Queue() {
		e := api.PendingUpdate{
			NodeIDHex: hex16(p.NodeID), VendorID: int(p.VendorID), ProductID: int(p.ProductID),
			TargetSoftwareVersion: int(p.TargetSoftwareVersion), Status: api.PendingUpdateStatus(p.Status),
			ProgressStatus: optString(p.ProgressStatus), ProgressAt: optTime(p.ProgressAt),
		}
		if p.Transferred != nil {
			e.Transferred = ptr(int(*p.Transferred)) //nolint:gosec // a transfer's byte count
		}
		if p.Length != nil {
			e.Length = ptr(int(*p.Length)) //nolint:gosec // an image's byte count
		}
		out.Queue = append(out.Queue, e)
	}
	return out, nil
}

func (s *controllerServer) updates() (SoftwareUpdateSource, error) {
	u := s.src.SoftwareUpdates()
	if u == nil {
		return nil, serviceUnready("No OTA provider", "the controller runs no OTA provider")
	}
	return u, nil
}

func (s *controllerServer) ConsentUpdate(ctx context.Context, req api.ConsentUpdateRequestObject) (api.ConsentUpdateResponseObject, error) {
	u, err := s.updates()
	if err != nil {
		return nil, err
	}
	b := req.Body
	if b == nil {
		return nil, badRequest("the body is required")
	}
	id, err := parseNodeID("node_id_hex", b.NodeIDHex)
	if err != nil {
		return nil, err
	}
	if b.VendorID < 0 || b.VendorID > 0xFFFF || b.ProductID < 0 || b.ProductID > 0xFFFF || b.TargetSoftwareVersion < 0 || int64(b.TargetSoftwareVersion) > 0xFFFFFFFF { // int64: int is 32 bits on arm/386
		return nil, badRequest("vendor_id, product_id or target_software_version out of range")
	}
	if _, ok := s.src.Node(id); !ok {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, hex16(id))
	}
	if err := u.Consent(ctx, id, uint16(b.VendorID), uint16(b.ProductID), uint32(b.TargetSoftwareVersion)); err != nil { //nolint:gosec // bounded above
		return nil, err
	}
	s.h.ctrlHub.Publish(Event{Type: EventUpdate, Data: controllerEventData{NodeIDHex: ptr(hex16(id))}})
	return api.ConsentUpdate204Response{}, nil
}

func (s *controllerServer) CheckUpdates(ctx context.Context, _ api.CheckUpdatesRequestObject) (api.CheckUpdatesResponseObject, error) {
	u, err := s.updates()
	if err != nil {
		return nil, err
	}
	list, err := u.Check(ctx)
	if err != nil {
		return nil, err
	}
	out := api.CheckUpdates200JSONResponse{Updates: []api.AvailableUpdate{}}
	for _, a := range list {
		out.Updates = append(out.Updates, api.AvailableUpdate{
			NodeIDHex: hex16(a.NodeID), VendorID: int(a.VendorID), ProductID: int(a.ProductID),
			SoftwareVersion: int(a.SoftwareVersion), SoftwareVersionString: a.SoftwareVersionString,
		})
	}
	s.h.ctrlHub.Publish(Event{Type: EventUpdate, Data: controllerEventData{}})
	return out, nil
}

// --- stream ---

// controllerStream is the stream's response: the strict server hands it
// the writer, and it serves server-sent events until the request ends.
type controllerStream struct {
	h   *handler
	ctx context.Context
}

func (c controllerStream) VisitControllerStreamResponse(w http.ResponseWriter) error {
	c.h.serveSSE(c.ctx, w, c.h.subscribeController)
	return nil
}

func (s *controllerServer) ControllerStream(ctx context.Context, _ api.ControllerStreamRequestObject) (api.ControllerStreamResponseObject, error) {
	return controllerStream{h: s.h, ctx: ctx}, nil
}

// subscribeController registers fn for the controller's events and the
// handler's own (commissioning progress, update) and returns the function
// that unregisters both.
func (h *handler) subscribeController(fn func(Event)) (unsubscribe func()) {
	unsubs := []func(){h.ctrlHub.Subscribe(fn)}
	if h.ctrl != nil {
		unsubs = append(unsubs, h.ctrl.Observe(func(e ControllerEvent) { fn(controllerEvent(e)) }))
	}
	return func() {
		for _, u := range unsubs {
			u()
		}
	}
}

// controllerEventData is the payload of every controller stream event,
// the contract's ControllerEvent (hand-written as the node family's event
// payloads are: the stream's schema is text, so the generator does not
// reach it).
type controllerEventData struct {
	NodeIDHex  *string              `json:"node_id_hex,omitempty"`
	Connection *api.ConnectionState `json:"connection,omitempty"`
	Active     *bool                `json:"active,omitempty"`
	Phase      *string              `json:"phase,omitempty"`
	Detail     *string              `json:"detail,omitempty"`
}

// controllerEvent is a port event as the stream carries it.
func controllerEvent(e ControllerEvent) Event {
	d := controllerEventData{}
	if e.Node != 0 {
		d.NodeIDHex = ptr(hex16(e.Node))
	}
	switch e.Type {
	case EventConnection:
		d.Connection = ptr(api.ConnectionState(e.Connection))
	case EventSubscription:
		d.Active = ptr(e.Active)
	case EventCommissioningProgress:
		if e.Phase != "" {
			d.Phase = ptr(e.Phase)
		}
		d.Detail = optString(e.Detail)
	}
	return Event{Type: e.Type, Data: d}
}
