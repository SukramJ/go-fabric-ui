// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Ported from openccu-loom internal/north/rest/handlers/matter_test.go,
// matter_exposures_test.go and matter_commissioning_emitter_test.go at
// b40b289a: the setup-payload, window open/close and share cases, on
// GET /node/pairing and POST /node/pairing/window[/close].

package fabricui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// Ports TestMatterSetupPayload_ValidConfig_Returns200: the configured codes
// while no window of this handler is open; no passcode.
func TestPairing_Configured_Returns200(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-pairing.200")
}

// Ports TestMatterSetupPayload_PasscodeZero_Returns503.
func TestPairing_NoCredentials_Returns503(t *testing.T) {
	n := sampleNode()
	n.codesErr = errors.New("node: invalid passcode 0")
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-pairing.503")
}

// Ports TestMatterShare_NilOpener_Returns503: no node, no pairing.
func TestPairing_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodGet, "/node/pairing", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

func TestPairing_ServesTheOpenFreshWindow(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`); rec.Code != http.StatusOK {
		t.Fatalf("open: %d %s", rec.Code, rec.Body)
	}
	runFixture(t, mux, "node-pairing.200-window")
}

func TestPairing_ForgetsAClosedFreshWindow(t *testing.T) {
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{node: n})
	do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`)
	n.open = false // closed by the node itself (a controller commissioned, the timer ran out)
	var p pairingResponse
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/pairing", "").Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Source != sourceConfigured || p.Passcode != nil {
		t.Fatalf("pairing %+v, want the configured codes without a passcode", p)
	}
	if h.ephemeral != nil {
		t.Fatal("the closed window's credentials are still held")
	}
}

func TestPairing_IsAdminOnly(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-pairing.403")
}

// Ports TestMatterCommissioningWindow_NilOpener_Returns503.
func TestOpenWindow_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{}`); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

// Ports TestMatterCommissioningWindow_InvalidBody_Returns400.
func TestOpenWindow_InvalidBody_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for _, body := range []string{`{`, `{"duration_seconds":"x"}`, `{"unknown":1}`} {
		if rec := do(t, mux, http.MethodPost, "/node/pairing/window", body); rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s: status %d, want 400", body, rec.Code)
		}
	}
}

// Ports TestMatterCommissioningWindow_DurationTooShort_Returns400.
func TestOpenWindow_DurationTooShort_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-pairing-window.400")
}

// Ports TestMatterCommissioningWindow_DurationTooLong_Returns400.
func TestOpenWindow_DurationTooLong_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":901}`)
	if rec.Code != http.StatusBadRequest || problemOf(t, rec).Detail != "got 901, want 180..900" {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Ports TestMatterCommissioningWindow_DurationZero_UsesDefault900.
func TestOpenWindow_DurationZero_UsesDefault900(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":0}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body)
	}
	if len(n.basic) != 1 || n.basic[0].Timeout != 900*time.Second {
		t.Fatalf("opened %+v, want one window of 900 s", n.basic)
	}
}

// Ports TestMatterCommissioningWindow_CommissioningInProgress_Returns409.
func TestOpenWindow_WindowOpen_Returns409(t *testing.T) {
	n := sampleNode()
	n.open, n.duration = true, 600
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-pairing-window.409")
	if len(n.basic) != 0 {
		t.Fatal("a second window was opened")
	}
}

// Ports TestMatterCommissioningWindow_CommissioningBlocked_Returns409: the
// cause (which can name paths) stays in the log.
func TestOpenWindow_CommissioningBlocked_Returns409(t *testing.T) {
	n := sampleNode()
	n.openErr = fmt.Errorf("%w: /etc/node/dac.pem: no such file", node.ErrCommissioningBlocked)
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{node: n, log: &log})
	rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d, want 409", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "dac.pem") || !strings.Contains(log.String(), "dac.pem") {
		t.Fatalf("the cause belongs in the log only: body %s", rec.Body)
	}
}

// Ports TestMatterCommissioningWindow_OpenerError_Returns500.
func TestOpenWindow_NodeError_Returns500(t *testing.T) {
	n := sampleNode()
	n.openErr = errors.New("bridge: socket closed")
	_, mux := newTestHandler(t, testOptions{node: n})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}

// Ports TestMatterCommissioningWindow_HappyPath_Returns200: fresh
// credentials, the codes for them, the passcode in the response.
func TestOpenWindow_HappyPath_Returns200(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-pairing-window.200")
	if len(n.basic) != 1 || n.basic[0].Passcode != 20202021 || n.basic[0].Iterations != ephemeralIterations {
		t.Fatalf("opened %+v", n.basic)
	}
}

// Ports TestMatterCommissioningWindow_HappyPath_RecordsAudit: the open is
// logged with its discriminator and duration, never with the passcode.
func TestOpenWindow_LogsWithoutThePasscode(t *testing.T) {
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{log: &log})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`); rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(log.String(), "fabricui.window.opened") {
		t.Fatalf("the open is not logged: %s", log.String())
	}
	if strings.Contains(log.String(), "20202021") {
		t.Fatalf("the passcode reached the log: %s", log.String())
	}
}

// Ports TestMatterCommissioningWindow_HappyPath_PublishesCommissioningWindowOpened:
// the event carries the discriminator and duration, never the passcode.
func TestOpenWindow_PublishesWindowOpenedWithoutThePasscode(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{})
	rec := record(h)
	do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":300}`)
	if !slices.Equal(rec.types(), []string{EventWindowOpened}) {
		t.Fatalf("events %v", rec.types())
	}
	raw, _ := json.Marshal(rec.events[0].Data)
	if strings.Contains(string(raw), "20202021") || !strings.Contains(string(raw), `"discriminator":1234`) {
		t.Fatalf("payload %s", raw)
	}
}

// Ports TestMatterShare_HappyPath_RecordsAuditUnderShareAction: sharing the
// node with a further controller is the same route; with ConfiguredWindow
// it opens the node's own window on its configured codes.
func TestOpenWindow_ConfiguredOpensTheOwnWindow(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n, configured: true})
	rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{"duration_seconds":600}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body)
	}
	var w windowResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &w); err != nil {
		t.Fatal(err)
	}
	if w.Passcode != nil || w.Discriminator != nil || w.QRCode != n.qr || w.DurationSeconds != 600 {
		t.Fatalf("window %+v", w)
	}
	if len(n.own) != 1 || n.own[0] != 600*time.Second || len(n.basic) != 0 {
		t.Fatalf("own %v basic %v", n.own, n.basic)
	}
}

func TestOpenWindow_RefusedWhileNoEndpointIsExposed(t *testing.T) {
	n := sampleNode()
	n.topo = node.BridgeTopology{Aggregator: 1}
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := runFixture(t, mux, "node-pairing-window.503")
	if rec.Header().Get("Retry-After") != "30" {
		t.Fatalf("Retry-After %q", rec.Header().Get("Retry-After"))
	}
	_, mux = newTestHandler(t, testOptions{node: n, allowEmpty: true})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window", `{}`); rec.Code != http.StatusOK {
		t.Fatalf("AllowEmptyTopology: status %d", rec.Code)
	}
}

// Ports TestMatterCommissioningClose_NilCloser_Returns503.
func TestCloseWindow_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window/close", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

// Ports TestMatterCommissioningClose_HappyPath_Returns204.
func TestCloseWindow_HappyPath_Returns204(t *testing.T) {
	n := sampleNode()
	n.open = true
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-pairing-window-close.204")
	if n.closed != 1 {
		t.Fatalf("closed %d times", n.closed)
	}
}

// Ports TestMatterCommissioningClose_CloserError_Returns500.
func TestCloseWindow_Error_Returns500(t *testing.T) {
	n := sampleNode()
	n.closeErr = errors.New("boom")
	_, mux := newTestHandler(t, testOptions{node: n})
	if rec := do(t, mux, http.MethodPost, "/node/pairing/window/close", ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}

// Ports TestMatterCommissioningClose_HappyPath_PublishesCommissioningProgress.
func TestCloseWindow_PublishesWindowClosed(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{})
	rec := record(h)
	do(t, mux, http.MethodPost, "/node/pairing/window/close", "")
	if !slices.Equal(rec.types(), []string{EventWindowClosed}) {
		t.Fatalf("events %v", rec.types())
	}
}

func TestFreshCredentials_AreValidAndDistinct(t *testing.T) {
	seen := map[uint32]bool{}
	for range 200 {
		c, err := freshCredentials(time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if !validPasscode(c.Passcode) || c.Passcode == 0 || c.Discriminator > node.MaxDiscriminator ||
			len(c.Salt) != 32 || c.Iterations != ephemeralIterations || c.Timeout != time.Minute {
			t.Fatalf("credentials %+v", c)
		}
		seen[c.Passcode] = true
	}
	if len(seen) < 190 {
		t.Fatalf("only %d distinct passcodes in 200 draws", len(seen))
	}
	for _, bad := range []uint32{0, 11111111, 12345678, 87654321, 99999999, 100000000} {
		if validPasscode(bad) {
			t.Fatalf("passcode %d accepted", bad)
		}
	}
}
