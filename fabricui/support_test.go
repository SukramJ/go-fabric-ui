// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"
)

// update rewrites the response half of every fixture a test runs from the
// handler's answer: go test ./fabricui -run Fixture -update. Review the
// diff before committing it — the fixtures are the contract.
var update = flag.Bool("update", false, "rewrite fixture responses from the handlers")

// fixtureDir is contract/fixtures, read by these tests and by the
// TypeScript client's.
const fixtureDir = "../contract/fixtures"

// testToken is the admin token the fixtures' requests carry.
const testToken = "test-token"

// testNow is the clock of every handler test.
var testNow = time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

// fixture is one contract/fixtures file.
type fixture struct {
	Route   string `json:"route"`
	Request struct {
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Headers map[string]string `json:"headers,omitempty"`
		Body    json.RawMessage   `json:"body,omitempty"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body,omitempty"`
	} `json:"response"`
}

func readFixture(t *testing.T, name string) (f fixture, path string) {
	t.Helper()
	path = filepath.Join(fixtureDir, name+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}
	return f, path
}

// runFixture sends the fixture's request to mux and checks the status and
// the JSON body after normalisation; with -update it writes the answer
// into the fixture instead. It returns the recorder for further checks.
func runFixture(t *testing.T, mux http.Handler, name string) *httptest.ResponseRecorder {
	t.Helper()
	f, path := readFixture(t, name)
	var body io.Reader
	if len(f.Request.Body) > 0 {
		body = bytes.NewReader(f.Request.Body)
	}
	req := httptest.NewRequestWithContext(context.Background(), f.Request.Method, DefaultPrefix+f.Request.Path, body)
	for k, v := range f.Request.Headers {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	got := normaliseJSON(t, rec.Body.Bytes())
	if *update {
		f.Response.Status = rec.Code
		f.Response.Body = got
		out, err := json.MarshalIndent(f, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(out, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
		return rec
	}
	if rec.Code != f.Response.Status {
		t.Fatalf("%s: status %d, fixture says %d; body %s", name, rec.Code, f.Response.Status, rec.Body)
	}
	want := normaliseJSON(t, f.Response.Body)
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: body differs from the fixture\n got: %s\nwant: %s", name, got, want)
	}
	return rec
}

// normaliseJSON re-encodes a JSON document with sorted keys; empty stays
// empty.
func normaliseJSON(t *testing.T, raw []byte) json.RawMessage {
	t.Helper()
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	// UseNumber keeps a 64-bit numeric twin exact in the fixture text.
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("not JSON: %v: %s", err, raw)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// fakeNode is the handler tests' node, as loom's handler tests fake each
// narrow port (internal/north/rest/handlers/matter_test.go at b40b289a).
type fakeNode struct {
	mu sync.Mutex

	healthOK bool
	healthy  bool
	note     string
	addr     string

	fabricList []node.FabricInfo
	fabricsErr error
	removeErr  error
	// removeFailAt fails the removal of this fabric index.
	removeFailAt uint8
	removed      []uint8

	// open, duration, kind and disc are the window node.Window reports;
	// an open window without a kind is the node's own.
	open     bool
	duration uint16
	kind     node.WindowKind
	disc     uint16
	openErr  error
	basic    []node.WindowCredentials
	own      []time.Duration
	closeErr error
	closed   int

	qr, manual string
	codesErr   error

	topo   node.BridgeTopology
	sess   node.SessionDiagnostics
	md     mdnsResponse
	events []node.DiagnosticEvent
	rep    surface.Report

	// observers are what node.Observe registered; the fake reports its
	// route-caused transitions to them as the node does.
	observers map[int]node.Observers
	nextObs   int
}

// notify runs fn on every registered observer set, outside the lock.
func (f *fakeNode) notify(fn func(node.Observers)) {
	f.mu.Lock()
	list := make([]node.Observers, 0, len(f.observers))
	for _, o := range f.observers {
		list = append(list, o)
	}
	f.mu.Unlock()
	for _, o := range list {
		fn(o)
	}
}

func (f *fakeNode) observe(o node.Observers) (unsubscribe func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.observers == nil {
		f.observers = map[int]node.Observers{}
	}
	id := f.nextObs
	f.nextObs++
	f.observers[id] = o
	return func() {
		f.mu.Lock()
		delete(f.observers, id)
		f.mu.Unlock()
	}
}

// observerCount is how many observer sets are registered.
func (f *fakeNode) observerCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.observers)
}

// transition reports the current window to the observers.
func (f *fakeNode) transition() {
	w := f.window()
	f.notify(func(o node.Observers) {
		if o.OnWindowTransition != nil {
			_ = o.OnWindowTransition(w)
		}
	})
}

func (f *fakeNode) health() (healthy bool, note string, ok bool) {
	return f.healthy, f.note, f.healthOK
}
func (f *fakeNode) localAddr() string { return f.addr }

func (f *fakeNode) fabrics(context.Context) ([]node.FabricInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]node.FabricInfo(nil), f.fabricList...), f.fabricsErr
}

func (f *fakeNode) removeFabric(_ context.Context, idx uint8) error {
	f.mu.Lock()
	if f.removeErr != nil || (f.removeFailAt != 0 && idx == f.removeFailAt) {
		defer f.mu.Unlock()
		if f.removeErr != nil {
			return f.removeErr
		}
		return errors.New("store: disk /var/lib/node.db is read-only")
	}
	f.removed = append(f.removed, idx)
	f.mu.Unlock()
	f.notify(func(o node.Observers) {
		if o.OnFabricRemoved != nil {
			_ = o.OnFabricRemoved(node.FabricInfo{FabricIndex: idx})
		}
	})
	return nil
}

// fakeDiscriminator and fakePasscode are the fake's configured
// credentials, the ones its configured codes encode.
const (
	fakeDiscriminator = 3840
	fakePasscode      = 20202021
)

func (f *fakeNode) window() node.WindowSnapshot {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.open {
		return node.WindowSnapshot{}
	}
	w := node.WindowSnapshot{
		Open: true, Own: true, Kind: f.kind, Discriminator: f.disc,
		Timeout: time.Duration(f.duration) * time.Second,
	}
	if w.Kind == node.WindowNone {
		w.Kind, w.Discriminator = node.WindowOwn, fakeDiscriminator
	}
	w.OpenedAt = testNow
	w.Deadline = testNow.Add(w.Timeout)
	return w
}

func (f *fakeNode) openBasicWindow(_ context.Context, creds node.WindowCredentials) error {
	f.mu.Lock()
	if f.openErr != nil {
		defer f.mu.Unlock()
		return f.openErr
	}
	f.basic = append(f.basic, creds)
	f.open, f.duration = true, uint16(creds.Timeout/time.Second)
	f.kind, f.disc = node.WindowBasic, creds.Discriminator
	f.mu.Unlock()
	f.transition()
	return nil
}

func (f *fakeNode) openOwnWindow(_ context.Context, timeout time.Duration) error {
	f.mu.Lock()
	if f.openErr != nil {
		defer f.mu.Unlock()
		return f.openErr
	}
	f.own = append(f.own, timeout)
	f.open, f.duration = true, uint16(timeout/time.Second)
	f.kind, f.disc = node.WindowOwn, fakeDiscriminator
	f.mu.Unlock()
	f.transition()
	return nil
}

func (f *fakeNode) closeWindow(context.Context) error {
	f.mu.Lock()
	if f.closeErr != nil {
		defer f.mu.Unlock()
		return f.closeErr
	}
	f.closed++
	f.open, f.duration, f.kind, f.disc = false, 0, node.WindowNone, 0
	f.mu.Unlock()
	f.transition()
	return nil
}

func (f *fakeNode) setupPayload() (node.SetupPayload, error) {
	if f.codesErr != nil {
		return node.SetupPayload{}, f.codesErr
	}
	return node.SetupPayload{
		Discriminator: fakeDiscriminator, Passcode: fakePasscode, VendorID: 0xFFF1, ProductID: 0x8000,
		QR: f.qr, Manual: f.manual,
	}, nil
}

func (f *fakeNode) pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
	return fmt.Sprintf("MT:FRESH%04d", creds.Discriminator), fmt.Sprintf("%011d", creds.Passcode), nil
}

func (f *fakeNode) topology() node.BridgeTopology            { return f.topo }
func (f *fakeNode) sessions() node.SessionDiagnostics        { return f.sess }
func (f *fakeNode) now() time.Time                           { return testNow }
func (f *fakeNode) mdns() mdnsResponse                       { return f.md }
func (f *fakeNode) diagnosticEvents() []node.DiagnosticEvent { return f.events }
func (f *fakeNode) surface() surface.Report                  { return f.rep }

// stringKey is a node.BridgedDevice key.
type stringKey string

func (k stringKey) String() string { return string(k) }

// sampleSurface is the surface report of the node-status.200 and
// node-surface.200 fixtures: an explicit level, two profiles, two
// extensions, assembled; the effective surface of the two top-level
// bridged endpoints of sampleNode ("lamp-1" on 2, "sensor-2" on 3); a
// decision of each verdict the pages tell apart — a silent keep, a
// profile's Warn, a profile's Remove, a removal the chain kept as
// mandatory (Keep with a reason) — beside one withheld element and one
// announced change of "lamp-1" with two element changes. Every element
// carries a kind and an origin: the report's encoding has no name for
// either's zero value (go-fabric surface/element.go:135-140).
func sampleSurface() surface.Report {
	r := surface.Policies{
		Level:      "1.5.1",
		Profiles:   []string{"home-assistant", "apple-home"},
		Extensions: surface.NewExtensions("BD-chip-Thermostat-Schedules", "BD-chip-Thermostat-Events"),
	}.Report()
	r.Assembled = true
	r.Effective = surface.Surface{Endpoints: []surface.Endpoint{
		{
			Key: "lamp-1", Number: 2,
			DeviceTypes: []surface.DeviceTypeRef{{ID: 0x0100, Revision: 4}},
			Clusters: []surface.Cluster{
				{ID: 0x0003, Revision: 6, Attributes: []uint32{0x0000, 0x0001, 0xFFFD}, Accepted: []uint32{0x00}},
				{ID: 0x0006, Revision: 6, FeatureMap: 0x1, Attributes: []uint32{0x0000, 0x4000, 0x4003, 0xFFFC, 0xFFFD}, Accepted: []uint32{0x00, 0x01, 0x02}},
			},
		},
		{
			Key: "sensor-2", Number: 3,
			DeviceTypes: []surface.DeviceTypeRef{{ID: 0x0302, Revision: 3}},
			Clusters: []surface.Cluster{
				{ID: 0x0402, Revision: 4, Attributes: []uint32{0x0000, 0x0001, 0x0002, 0xFFFD}},
			},
		},
	}}
	onOff := surface.Element{Endpoint: "lamp-1", Cluster: 0x0006, Kind: surface.KindCluster, ID: 0x0006, Origin: surface.OriginDeviceType, Mandatory: true}
	startUp := surface.Element{Endpoint: "lamp-1", Cluster: 0x0006, Kind: surface.KindAttribute, ID: 0x4003, Origin: surface.OriginHostDeclared}
	tolerance := surface.Element{Endpoint: "sensor-2", Cluster: 0x0402, Kind: surface.KindAttribute, ID: 0x0003, Origin: surface.OriginHostDeclared}
	r.Decisions = []surface.Decision{
		{Element: onOff, Policy: "level", Verdict: surface.Keep},
		{Element: startUp, Policy: "profile:apple-home", Verdict: surface.Warn, Reason: "warning (Reported) profile apple-home: not reproduced in isolation (owner report)"},
		{Element: tolerance, Policy: "profile:home-assistant", Verdict: surface.Remove, Reason: "removed by profile home-assistant: synthetic (test)"},
		{Element: onOff, Policy: "profile:home-assistant", Verdict: surface.Keep, Reason: "kept: mandatory at pin (removed by profile home-assistant: synthetic (test))"},
	}
	r.Withheld = []surface.Element{{Endpoint: "air-1", Kind: surface.KindDeviceType, ID: 0x002C, Origin: surface.OriginDeviceType, Mandatory: true}}
	r.Changed = []surface.SurfaceChange{{
		Endpoint: "lamp-1",
		Before:   [32]byte{0x1A, 0x2B, 0x3C, 0x4D, 0x5E, 0x6F},
		After:    [32]byte{0x9A, 0x8B, 0x7C, 0x6D, 0x5E, 0x4F},
		Changes: []surface.Change{
			{Endpoint: "lamp-1", Cluster: 0x0006, Kind: surface.KindCluster, ID: 0x0006, Before: "revision 5", After: "revision 6"},
			{Endpoint: "lamp-1", Cluster: 0x0006, Kind: surface.KindAttribute, ID: 0x4003, After: "served"},
		},
		At: testNow.Add(-time.Minute),
	}}
	return r
}

// sampleNode is a node with two fabrics, two bridged devices (one with a
// part), one session pair, an advertisement, a trace and a surface report.
func sampleNode() *fakeNode {
	return &fakeNode{
		rep:      sampleSurface(),
		healthOK: true, healthy: true, note: "listener ok",
		addr: "[::]:5540",
		fabricList: []node.FabricInfo{
			{
				FabricIndex: 1, FabricID: 0x0000000000000001, NodeID: 0xB2C3D4E5F6071829,
				RootPublicKey:      []byte{0x04, 0xAA, 0xBB},
				CompressedFabricID: [8]byte{0x9C, 0x71, 0xD3, 0x8F, 0xBE, 0x48, 0xF2, 0xE5},
				VendorID:           0x1349, Label: "Home",
			},
			{
				FabricIndex: 2, FabricID: 0x00000000000000AB, NodeID: 0x0000000000000042,
				RootPublicKey:      []byte{0x04, 0x01},
				CompressedFabricID: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
				VendorID:           0x6006,
			},
		},
		qr: "MT:Y.K9042C00KA0648G00", manual: "34970112332",
		topo: node.BridgeTopology{
			Aggregator: 1,
			Devices: []node.BridgedDevice{
				{
					Key: stringKey("lamp-1"), Endpoint: 2, DeviceType: 0x0100, DeviceTypeRevision: 4, Name: "Kitchen lamp", Reachable: true,
					Clusters: []node.ClusterInfo{
						{ID: 0x0003, Name: "Identify", Revision: 6},
						{ID: 0x0006, Name: "OnOff", Revision: 6},
						{ID: 0x001D, Name: "Descriptor", Revision: 3},
						{ID: 0x0039, Name: "BridgedDeviceBasicInformation", Revision: 6},
					},
				},
				{
					Key: stringKey("sensor-2"), Endpoint: 3, DeviceType: 0x0302, DeviceTypeRevision: 3, Name: "Hall sensor", Reachable: false,
					Clusters: []node.ClusterInfo{
						{ID: 0x001D, Name: "Descriptor", Revision: 3},
						{ID: 0x0039, Name: "BridgedDeviceBasicInformation", Revision: 6},
						{ID: 0x0402, Name: "TemperatureMeasurement", Revision: 6},
					},
					Parts: []node.BridgedDevice{{
						Key: stringKey("sensor-2/humidity"), Endpoint: 4, DeviceType: 0x0307, DeviceTypeRevision: 3, Name: "Hall humidity", Reachable: true,
						Clusters: []node.ClusterInfo{{ID: 0x001D, Name: "Descriptor", Revision: 3}, {ID: 0x0405, Name: "RelativeHumidityMeasurement", Revision: 5}},
					}},
				},
			},
		},
		sess: node.SessionDiagnostics{
			AsOf: testNow,
			Sessions: []node.SessionInfo{
				{
					ID: 17, FabricIndex: 1, PeerNodeID: 0x000000000001B669, LocalNodeID: 0xB2C3D4E5F6071829,
					Subscriptions: 2, LastActivity: testNow.Add(-5 * time.Second), LastPeerActivity: testNow.Add(-45 * time.Second),
					Idle: 5 * time.Second, PeerIdle: 45 * time.Second,
				},
				{
					ID: 18, FabricIndex: 0, PASE: true,
					LastActivity: testNow.Add(-2 * time.Second), LastPeerActivity: testNow.Add(-2 * time.Second),
					Idle: 2 * time.Second, PeerIdle: 2 * time.Second,
				},
			},
			Occupancy: node.Occupancy{Live: 2, Reserved: 1, Capacity: 65535, Free: 65532},
		},
		md: mdnsResponse{
			Advertising: true,
			Services: []mdnsService{{
				ServiceType: "_matter._tcp", InstanceName: "9C71D38FBE48F2E5-B2C3D4E5F6071829", HostName: "node-host",
				Port: 5540, Addresses: []string{"fe80::1"}, Subtypes: []string{"_I9C71D38FBE48F2E5"},
				TXT: map[string]string{"SII": "5000", "SAI": "300"},
			}},
			Findings: []mdnsFinding{{Severity: "warning", Code: "link_local_only", Message: "Only a link-local address is advertised.", Service: "_matter._tcp"}},
		},
		events: []node.DiagnosticEvent{
			{At: testNow.Add(-time.Minute), Kind: "pairing", Severity: "error", Message: "A controller's pairing attempt was refused.", Detail: map[string]string{"reason": "busy"}},
			{At: testNow.Add(-2 * time.Minute), Kind: "session", Severity: "info", Message: "A controller opened a session."},
		},
	}
}

// testOptions are a handler's options in token mode with testToken.
type testOptions struct {
	node       *fakeNode
	noNode     bool
	exposure   ExposureProvider
	admin      func(*http.Request) bool
	reassemble func(context.Context) error
	events     EventSource
	configured bool
	allowEmpty bool
	warnings   func(context.Context) []Warning
	log        *bytes.Buffer
	controller ControllerSource
}

// newTestHandler builds a handler on the fake node and returns it with its
// routes.
func newTestHandler(t *testing.T, o testOptions) (*handler, *http.ServeMux) {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	if o.log != nil {
		logger = slog.New(slog.NewTextHandler(o.log, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	h, err := newHandler(Options{
		Exposure: o.exposure, Admin: o.admin, Reassemble: o.reassemble, Events: o.events,
		Logger: logger, ConfiguredWindow: o.configured, AllowEmptyTopology: o.allowEmpty,
		Warnings: o.warnings, Controller: o.controller,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !o.noNode {
		if o.node == nil {
			o.node = sampleNode()
		}
		h.node = o.node
	}
	if o.admin == nil {
		h.token = testToken
	}
	h.fresh = func(timeout time.Duration) (node.WindowCredentials, error) {
		return node.WindowCredentials{
			Discriminator: 1234, Passcode: 20202021, Iterations: node.PBKDFIterationsMin,
			Salt: bytes.Repeat([]byte{7}, 32), Timeout: timeout,
		}, nil
	}
	mux := http.NewServeMux()
	h.register(mux)
	return h, mux
}

// do sends one request with the admin token.
func do(t *testing.T, mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var rd io.Reader
	if body != "" {
		rd = strings.NewReader(body)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, DefaultPrefix+path, rd)
	req.Header.Set(TokenHeader, testToken)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// problemOf decodes a problem body.
func problemOf(t *testing.T, rec *httptest.ResponseRecorder) problem {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("content type %q, want application/problem+json; body %s", ct, rec.Body)
	}
	var p problem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// recorder collects the events a handler publishes.
type recorder struct {
	mu     sync.Mutex
	events []Event
}

func (r *recorder) add(ev Event) {
	r.mu.Lock()
	r.events = append(r.events, ev)
	r.mu.Unlock()
}

func (r *recorder) types() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, len(r.events))
	for _, ev := range r.events {
		out = append(out, ev.Type)
	}
	return out
}

// record subscribes as a stream does: the handler's own events, the
// host's source and the node's observers.
func record(h *handler) *recorder {
	r := &recorder{}
	h.subscribe(r.add)
	return r
}
