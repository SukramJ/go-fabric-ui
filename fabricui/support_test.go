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

	open     bool
	duration uint16
	openErr  error
	basic    []node.WindowCredentials
	own      []time.Duration
	closeErr error
	closed   int

	qr, manual string
	codesErr   error

	topo   node.BridgeTopology
	sess   []sessionInfo
	occ    occupancyResponse
	md     mdnsResponse
	events []node.DiagnosticEvent
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
	defer f.mu.Unlock()
	if f.removeErr != nil || (f.removeFailAt != 0 && idx == f.removeFailAt) {
		if f.removeErr != nil {
			return f.removeErr
		}
		return errors.New("store: disk /var/lib/node.db is read-only")
	}
	f.removed = append(f.removed, idx)
	return nil
}

func (f *fakeNode) window() (open bool, durationSeconds uint16) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.open, f.duration
}

func (f *fakeNode) openBasicWindow(_ context.Context, creds node.WindowCredentials) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.openErr != nil {
		return f.openErr
	}
	f.basic = append(f.basic, creds)
	f.open, f.duration = true, uint16(creds.Timeout/time.Second)
	return nil
}

func (f *fakeNode) openOwnWindow(_ context.Context, timeout time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.openErr != nil {
		return f.openErr
	}
	f.own = append(f.own, timeout)
	f.open, f.duration = true, uint16(timeout/time.Second)
	return nil
}

func (f *fakeNode) closeWindow(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closeErr != nil {
		return f.closeErr
	}
	f.closed++
	f.open, f.duration = false, 0
	return nil
}

func (f *fakeNode) pairingCodes() (qr, manual string, err error) { return f.qr, f.manual, f.codesErr }

func (f *fakeNode) pairingCodesFor(creds node.WindowCredentials) (qr, manual string, err error) {
	return fmt.Sprintf("MT:FRESH%04d", creds.Discriminator), fmt.Sprintf("%011d", creds.Passcode), nil
}

func (f *fakeNode) topology() node.BridgeTopology                { return f.topo }
func (f *fakeNode) sessions() ([]sessionInfo, occupancyResponse) { return f.sess, f.occ }
func (f *fakeNode) mdns() mdnsResponse                           { return f.md }
func (f *fakeNode) diagnosticEvents() []node.DiagnosticEvent     { return f.events }

// stringKey is a node.BridgedDevice key.
type stringKey string

func (k stringKey) String() string { return string(k) }

// sampleNode is a node with two fabrics, two bridged devices (one with a
// part), one session pair, an advertisement and a trace.
func sampleNode() *fakeNode {
	return &fakeNode{
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
				{Key: stringKey("lamp-1"), Endpoint: 2, DeviceType: 0x0100, Name: "Kitchen lamp", Reachable: true},
				{
					Key: stringKey("sensor-2"), Endpoint: 3, DeviceType: 0x0302, Name: "Hall sensor", Reachable: false,
					Parts: []node.BridgedDevice{{Key: stringKey("sensor-2/humidity"), Endpoint: 4, DeviceType: 0x0307, Name: "Hall humidity", Reachable: true}},
				},
			},
		},
		sess: []sessionInfo{
			{
				SessionID: 17, FabricIndex: 1, PeerNodeID: 0x000000000001B669, LocalNodeID: 0xB2C3D4E5F6071829,
				Subscriptions: 2, LastActivity: testNow.Add(-5 * time.Second), LastPeerActivity: testNow.Add(-45 * time.Second),
			},
			{
				SessionID: 18, FabricIndex: 0, IsPASE: true,
				LastActivity: testNow.Add(-2 * time.Second), LastPeerActivity: testNow.Add(-2 * time.Second),
			},
		},
		occ: occupancyResponse{Live: 2, Reserved: 1, Capacity: 65535, Free: 65532},
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
	log        *bytes.Buffer
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
	h.now = func() time.Time { return testNow }
	h.fresh = func(timeout time.Duration) (node.WindowCredentials, error) {
		return node.WindowCredentials{
			Discriminator: 1234, Passcode: 20202021, Iterations: ephemeralIterations,
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

func record(h *handler) *recorder {
	r := &recorder{}
	h.hub.Subscribe(r.add)
	return r
}
