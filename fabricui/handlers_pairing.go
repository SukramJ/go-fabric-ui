// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// Window duration bounds, Matter §11.19.8.1: 180..900 s, 0 is the 15-minute
// default. Mirrors loom internal/north/rest/handlers/matter.go:174-183
// (matterOpenCommissioningWindow) at b40b289a.
const (
	windowMinSeconds     = 180
	windowMaxSeconds     = 900
	windowDefaultSeconds = 900
)

// ephemeralWindow is a window this handler opened on fresh credentials.
// openedAt is the node's OpenedAt for it, read right after the open; zero
// when the node reported no basic window then.
type ephemeralWindow struct {
	creds    node.WindowCredentials
	qr       string
	manual   string
	openedAt time.Time
}

// pairingResponse is the body of GET /node/pairing: the codes with the
// discriminator and passcode they encode — of the window this handler
// opened on fresh credentials, else the configured ones
// (node.SetupPayload). Loom's setup payload served the same
// (matter.go:51-85 at b40b289a).
type pairingResponse struct {
	Discriminator   *uint16 `json:"discriminator,omitempty"`
	Passcode        *uint32 `json:"passcode,omitempty"`
	QRCode          string  `json:"qr_code"`
	ManualCode      string  `json:"manual_code"`
	Source          string  `json:"source"`
	WindowOpen      bool    `json:"window_open"`
	DurationSeconds uint16  `json:"duration_seconds,omitempty"`
	// WindowRemaining and WindowMatch are absent while no window is open
	// (0.10.0): the seconds until it closes, and whether it admits these
	// codes (false: a window on other credentials, opened elsewhere).
	WindowRemaining *int  `json:"window_remaining_seconds,omitempty"`
	WindowMatch     *bool `json:"window_match,omitempty"`
}

// Pairing sources.
const (
	sourceConfigured = "configured"
	sourceWindow     = "window"
)

// pairing serves GET /node/pairing (admin): the codes of the window this
// handler opened while it is open, else the node's configured setup
// payload (node.SetupPayload). Mirrors loom matter.go:51-85 (MatterSetupPayload),
// whose 503 for a node without commissioning credentials it keeps.
func (h *handler) pairing(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	win := h.node.window()
	var duration uint16
	if win.Open {
		duration = durationSeconds(win.Timeout)
	}
	if eph := h.activeEphemeral(win); eph != nil {
		pc, disc := eph.creds.Passcode, eph.creds.Discriminator
		writeJSON(w, pairingResponse{
			Discriminator: &disc, Passcode: &pc,
			QRCode: eph.qr, ManualCode: eph.manual, Source: sourceWindow,
			WindowOpen: true, DurationSeconds: duration,
			WindowRemaining: remainingSeconds(win, h.node.now()), WindowMatch: ptr(true),
		})
		return
	}
	sp, err := h.node.setupPayload()
	if err != nil {
		h.log.WarnContext(r.Context(), "fabricui.pairing.codes_unavailable", slog.String("error", err.Error()))
		unready(w, r, "Pairing codes not available", "the node has no valid commissioning credentials configured")
		return
	}
	pc, disc := sp.Passcode, sp.Discriminator
	res := pairingResponse{
		Discriminator: &disc, Passcode: &pc,
		QRCode: sp.QR, ManualCode: sp.Manual,
		Source: sourceConfigured, WindowOpen: win.Open, DurationSeconds: duration,
	}
	if win.Open {
		// The configured codes commission through the node's own window
		// only (decision 28).
		res.WindowRemaining = remainingSeconds(win, h.node.now())
		res.WindowMatch = ptr(win.Kind == node.WindowOwn)
	}
	writeJSON(w, res)
}

// activeEphemeral returns the window this handler opened while the node
// still serves it (servesEphemeral) and forgets it otherwise: a window the
// node closed meanwhile (commissioned, timed out, revoked over the wire)
// is no longer the one GET /node/pairing serves.
func (h *handler) activeEphemeral(win node.WindowSnapshot) *ephemeralWindow {
	h.mu.Lock()
	defer h.mu.Unlock()
	eph := h.ephemeral
	if eph == nil {
		return nil
	}
	if !h.servesEphemeral(win, eph) {
		h.ephemeral = nil
		return nil
	}
	return eph
}

// servesEphemeral reports whether win is the window eph names: a basic
// window (node.WindowBasic) opened when eph was, on its discriminator,
// whose deadline has not passed on the node's clock. The window is told
// by its OpenedAt: the snapshot's discriminator is that of the record the
// node published, 0 while the publish failed (UR-P2-err-02), so it decides
// only when the snapshot names one.
func (h *handler) servesEphemeral(win node.WindowSnapshot, eph *ephemeralWindow) bool {
	switch {
	case !win.Open, win.Kind != node.WindowBasic:
		return false
	case !eph.openedAt.IsZero() && !win.OpenedAt.Equal(eph.openedAt):
		return false
	case win.Discriminator != 0 && win.Discriminator != eph.creds.Discriminator:
		return false
	case !win.Deadline.IsZero() && !h.node.now().Before(win.Deadline):
		return false
	}
	return true
}

// windowRequest is the body of POST /node/pairing/window.
type windowRequest struct {
	DurationSeconds uint16 `json:"duration_seconds"`
}

// windowResponse is its answer: what a controller needs to commission the
// node through the window, the discriminator and passcode included.
type windowResponse struct {
	Discriminator   *uint16 `json:"discriminator,omitempty"`
	Passcode        *uint32 `json:"passcode,omitempty"`
	DurationSeconds uint16  `json:"duration_seconds"`
	QRCode          string  `json:"qr_code"`
	ManualCode      string  `json:"manual_code"`
}

// openWindow serves POST /node/pairing/window (admin). Mirrors loom
// matter.go:158-226 (matterOpenCommissioningWindow) and the opener behind
// it (cmd/openccu-loom/matter_window_adapter.go:95-200): the duration
// bounds, 409 when a window is open or commissioning is blocked, 503 with
// Retry-After while no endpoint is exposed, and fresh credentials unless
// [Options].ConfiguredWindow. The passcode is in the response only — never
// in an event, never in a log.
func (h *handler) openWindow(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	var body windowRequest
	if err := decodeJSON(w, r, &body); err != nil {
		badBody(w, r, err)
		return
	}
	duration := body.DurationSeconds
	if duration == 0 {
		duration = windowDefaultSeconds
	}
	if duration < windowMinSeconds || duration > windowMaxSeconds {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "duration_seconds out of range",
			fmt.Sprintf("got %d, want %d..%d", duration, windowMinSeconds, windowMaxSeconds))
		return
	}
	const title = "Failed to open commissioning window"
	if !h.allowEmpty && len(h.node.topology().Devices) == 0 {
		// A controller that subscribes before the first endpoint is in the
		// topology caches an empty parts list (loom matter.go:134-140,
		// ErrBridgeTopologyNotReady).
		w.Header().Set("Retry-After", "30")
		unready(w, r, title, "the node exposes no endpoint yet; retry once the host has assembled its endpoints")
		return
	}
	// openMu holds from the check to the stored credentials: a second
	// open finds the first one's window open, 409, instead of replacing
	// it with codes of its own (UR-P2-conc-07).
	h.openMu.Lock()
	defer h.openMu.Unlock()
	if h.node.window().Open {
		writeProblem(w, r, http.StatusConflict, codeConflict, title, "a commissioning window is already open")
		return
	}
	timeout := time.Duration(duration) * time.Second
	var res windowResponse
	var err error
	if h.configured {
		res, err = h.openConfigured(r, timeout)
	} else {
		res, err = h.openFresh(r, timeout)
	}
	if err != nil {
		switch {
		case errors.Is(err, node.ErrCommissioningWindowAlreadyOpen), errors.Is(err, node.ErrAdministratorWindowOpen):
			writeProblem(w, r, http.StatusConflict, codeConflict, title, "a commissioning window is already open")
		case errors.Is(err, node.ErrCommissioningBlocked):
			// The cause can name file paths; it stays in the log.
			h.log.WarnContext(r.Context(), "fabricui.window.blocked", slog.String("error", err.Error()))
			writeProblem(w, r, http.StatusConflict, codeConflict, title,
				"commissioning is blocked on this node; the node's log names the cause")
		case errors.Is(err, node.ErrCommissioningWindowNotConfigured):
			unready(w, r, title, "the node has no commissioning credentials configured")
		default:
			h.serverError(w, r, title, err)
		}
		return
	}
	if win := h.node.window(); win.Open {
		res.DurationSeconds = durationSeconds(win.Timeout)
	}
	if res.DurationSeconds == 0 {
		res.DurationSeconds = duration
	}
	// window_opened reaches the stream through the node's observer.
	h.log.InfoContext(r.Context(), "fabricui.window.opened",
		slog.Int("discriminator", int(*res.Discriminator)), slog.Int("duration_seconds", int(res.DurationSeconds)),
		slog.Bool("configured", h.configured))
	writeJSON(w, res)
}

// openConfigured opens the node's own window on its configured
// credentials and answers with its setup payload (node.SetupPayload).
func (h *handler) openConfigured(r *http.Request, timeout time.Duration) (windowResponse, error) {
	if err := h.node.openOwnWindow(r.Context(), timeout); err != nil {
		return windowResponse{}, err
	}
	sp, err := h.node.setupPayload()
	if err != nil {
		_ = h.node.closeWindow(r.Context())
		return windowResponse{}, err
	}
	pc, disc := sp.Passcode, sp.Discriminator
	return windowResponse{Discriminator: &disc, Passcode: &pc, QRCode: sp.QR, ManualCode: sp.Manual}, nil
}

// openFresh opens the node's own window on credentials drawn for it by
// node.NewWindowCredentials: a fresh discriminator, passcode and salt.
// Mirrors loom matter_window_adapter.go:143-200 (openEphemeral).
func (h *handler) openFresh(r *http.Request, timeout time.Duration) (windowResponse, error) {
	creds, err := h.fresh(timeout)
	if err != nil {
		return windowResponse{}, err
	}
	qr, manual, err := h.node.pairingCodesFor(creds)
	if err != nil {
		return windowResponse{}, err
	}
	if err := h.node.openBasicWindow(r.Context(), creds); err != nil {
		return windowResponse{}, err
	}
	eph := &ephemeralWindow{creds: creds, qr: qr, manual: manual}
	if win := h.node.window(); win.Open && win.Kind == node.WindowBasic {
		eph.openedAt = win.OpenedAt
	}
	h.mu.Lock()
	h.ephemeral = eph
	h.mu.Unlock()
	h.forgetOnClose(eph)
	pc, disc := creds.Passcode, creds.Discriminator
	return windowResponse{Discriminator: &disc, Passcode: &pc, QRCode: qr, ManualCode: manual}, nil
}

// closeWindow serves POST /node/pairing/window/close (admin): it closes
// whatever window is open; a PASE session already established and an armed
// fail-safe stay with their commissioner (node.CloseWindow). Mirrors loom matter_exposures.go:422-444
// (MatterCommissioningClose).
func (h *handler) closeWindow(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	h.mu.Lock()
	prev := h.ephemeral
	h.mu.Unlock()
	if err := h.node.closeWindow(r.Context()); err != nil {
		h.serverError(w, r, "Failed to close commissioning window", err)
		return
	}
	// window_closed reaches the stream through the node's observer. Only
	// the credentials held before the close are forgotten: a fresh window
	// another request opened meanwhile keeps its own (UR-P1-conc-05); a
	// window this close did end anyway is forgotten by activeEphemeral.
	h.forgetEphemeralIf(prev)
	w.WriteHeader(http.StatusNoContent)
}

// durationSeconds is a window's timeout in whole seconds, clamped to the
// uint16 the contract carries.
func durationSeconds(d time.Duration) uint16 {
	sec := d / time.Second
	switch {
	case sec <= 0:
		return 0
	case sec > math.MaxUint16:
		return math.MaxUint16
	}
	return uint16(sec)
}

// forgetEphemeralIf drops the credentials of the window this handler
// opened while they are still prev.
func (h *handler) forgetEphemeralIf(prev *ephemeralWindow) {
	h.mu.Lock()
	if h.ephemeral == prev {
		h.ephemeral = nil
	}
	h.mu.Unlock()
}

// forgetOnClose drops eph's credentials on the node's transition that ends
// the window — closed, commissioned, timed out, replaced — instead of
// holding them until the next GET /node/pairing (UR-P2-leak-08). The
// observer unregisters itself once eph is gone, off the node's goroutine.
func (h *handler) forgetOnClose(eph *ephemeralWindow) {
	var mu sync.Mutex
	var unobserve func()
	ended := false
	end := func() {
		mu.Lock()
		u := unobserve
		first := !ended
		ended = true
		mu.Unlock()
		if first && u != nil {
			go u()
		}
	}
	u := h.node.observe(node.Observers{OnWindowTransition: func(w node.WindowSnapshot) {
		h.mu.Lock()
		gone := h.ephemeral != eph
		if !gone && !h.servesEphemeral(w, eph) {
			h.ephemeral, gone = nil, true
		}
		h.mu.Unlock()
		if gone {
			end()
		}
	}})
	mu.Lock()
	unobserve = u
	done := ended
	mu.Unlock()
	if done {
		u()
		return
	}
	// A window that ended before the observer was registered.
	if h.activeEphemeral(h.node.window()) != eph {
		end()
	}
}
