// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/SukramJ/go-fabric/node"
)

// Window duration bounds, Matter §11.19.8.1: 180..900 s, 0 is the 15-minute
// default. Mirrors loom internal/north/rest/handlers/matter.go:169-180
// (matterOpenCommissioningWindow) at b40b289a.
const (
	windowMinSeconds     = 180
	windowMaxSeconds     = 900
	windowDefaultSeconds = 900
)

// ephemeralIterations is the PBKDF2 iteration count of a window opened on
// fresh credentials — loom's default (cmd/openccu-loom/matter_window_adapter.go:
// 159-162), the minimum the spec allows.
const ephemeralIterations = 1000

// ephemeralWindow is a window this handler opened on fresh credentials.
type ephemeralWindow struct {
	creds   node.WindowCredentials
	qr      string
	manual  string
	expires time.Time
}

// pairingResponse is the body of GET /node/pairing. Discriminator and
// passcode are set only for a window this handler opened on fresh
// credentials: the facade hands out neither the configured passcode nor the
// discriminator of the node's own window (CK questions); the codes carry
// both.
type pairingResponse struct {
	Discriminator   *uint16 `json:"discriminator,omitempty"`
	Passcode        *uint32 `json:"passcode,omitempty"`
	QRCode          string  `json:"qr_code"`
	ManualCode      string  `json:"manual_code"`
	Source          string  `json:"source"`
	WindowOpen      bool    `json:"window_open"`
	DurationSeconds uint16  `json:"duration_seconds,omitempty"`
}

// Pairing sources.
const (
	sourceConfigured = "configured"
	sourceWindow     = "window"
)

// pairing serves GET /node/pairing (admin): the codes of the window this
// handler opened while it is open, else the node's configured codes
// (node.PairingCodes). Mirrors loom matter.go:51-85 (MatterSetupPayload),
// whose 503 for a node without commissioning credentials it keeps.
func (h *handler) pairing(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	open, duration := h.node.window()
	if eph := h.activeEphemeral(open); eph != nil {
		pc, disc := eph.creds.Passcode, eph.creds.Discriminator
		writeJSON(w, pairingResponse{
			Discriminator: &disc, Passcode: &pc,
			QRCode: eph.qr, ManualCode: eph.manual, Source: sourceWindow,
			WindowOpen: true, DurationSeconds: duration,
		})
		return
	}
	qr, manual, err := h.node.pairingCodes()
	if err != nil {
		h.log.WarnContext(r.Context(), "fabricui.pairing.codes_unavailable", slog.String("error", err.Error()))
		unready(w, r, "Pairing codes not available", "the node has no valid commissioning credentials configured")
		return
	}
	res := pairingResponse{
		QRCode: qr, ManualCode: manual,
		Source: sourceConfigured, WindowOpen: open,
	}
	if open {
		res.DurationSeconds = duration
	}
	writeJSON(w, res)
}

// activeEphemeral returns the window this handler opened while a window is
// open and its time has not run out, and forgets it otherwise. A window
// closed meanwhile is forgotten when the route closes it or the node's
// observers report the close (forgetOnClose).
func (h *handler) activeEphemeral(open bool) *ephemeralWindow {
	h.mu.Lock()
	defer h.mu.Unlock()
	eph := h.ephemeral
	if eph == nil {
		return nil
	}
	if !open || !h.now().Before(eph.expires) {
		h.ephemeral = nil
		return nil
	}
	return eph
}

// windowRequest is the body of POST /node/pairing/window.
type windowRequest struct {
	DurationSeconds uint16 `json:"duration_seconds"`
}

// windowResponse is its answer: what a controller needs to commission the
// node through the window. Discriminator and passcode are absent for a
// window on the configured credentials, which the facade does not report.
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
		// topology caches an empty parts list (loom matter.go:131-139,
		// ErrBridgeTopologyNotReady).
		w.Header().Set("Retry-After", "30")
		unready(w, r, title, "the node exposes no endpoint yet; retry once the host has assembled its endpoints")
		return
	}
	if open, _ := h.node.window(); open {
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
	_, res.DurationSeconds = h.node.window()
	if res.DurationSeconds == 0 {
		res.DurationSeconds = duration
	}
	ev := WindowEvent{DurationSeconds: res.DurationSeconds}
	if res.Discriminator != nil {
		ev.Discriminator = *res.Discriminator
	}
	h.log.InfoContext(r.Context(), "fabricui.window.opened",
		slog.Int("discriminator", int(ev.Discriminator)), slog.Int("duration_seconds", int(res.DurationSeconds)),
		slog.Bool("configured", h.configured))
	h.publishNodeEvent(Event{Type: EventWindowOpened, Data: ev})
	writeJSON(w, res)
}

// openConfigured opens the node's own window on its configured credentials.
func (h *handler) openConfigured(r *http.Request, timeout time.Duration) (windowResponse, error) {
	if err := h.node.openOwnWindow(r.Context(), timeout); err != nil {
		return windowResponse{}, err
	}
	qr, manual, err := h.node.pairingCodes()
	if err != nil {
		_ = h.node.closeWindow(r.Context())
		return windowResponse{}, err
	}
	return windowResponse{QRCode: qr, ManualCode: manual}, nil
}

// openFresh opens the node's own window on credentials drawn for it: a
// fresh discriminator, passcode and salt. Mirrors loom
// matter_window_adapter.go:143-200 (openEphemeral).
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
	h.mu.Lock()
	h.ephemeral = &ephemeralWindow{creds: creds, qr: qr, manual: manual, expires: h.now().Add(timeout)}
	h.mu.Unlock()
	pc, disc := creds.Passcode, creds.Discriminator
	return windowResponse{Discriminator: &disc, Passcode: &pc, QRCode: qr, ManualCode: manual}, nil
}

// closeWindow serves POST /node/pairing/window/close (admin): it revokes
// whatever window is open. Mirrors loom matter_exposures.go:422-444
// (MatterCommissioningClose).
func (h *handler) closeWindow(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	if err := h.node.closeWindow(r.Context()); err != nil {
		h.serverError(w, r, "Failed to close commissioning window", err)
		return
	}
	h.forgetEphemeral()
	h.publishNodeEvent(Event{Type: EventWindowClosed, Data: WindowEvent{}})
	w.WriteHeader(http.StatusNoContent)
}

// freshCredentials draws a window's credentials: a 12-bit discriminator, a
// passcode setup.IsValidSetupPIN accepts and a 32-byte salt.
func freshCredentials(timeout time.Duration) (node.WindowCredentials, error) {
	var b [4 + 4 + 32]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			return node.WindowCredentials{}, err
		}
		disc := uint16(binary.BigEndian.Uint32(b[0:4]) & uint32(node.MaxDiscriminator))
		pc := binary.BigEndian.Uint32(b[4:8])%99999998 + 1
		if !validPasscode(pc) {
			continue
		}
		return node.WindowCredentials{
			Discriminator: disc,
			Passcode:      pc,
			Iterations:    ephemeralIterations,
			Salt:          append([]byte(nil), b[8:]...),
			Timeout:       timeout,
		}, nil
	}
}

// validPasscode applies the setup PIN rule of Matter §5.1.7.1: 1..99999998
// and none of the trivial values. OpenBasicWindow checks it again.
func validPasscode(pc uint32) bool {
	switch pc {
	case 0, 11111111, 22222222, 33333333, 44444444, 55555555, 66666666, 77777777, 88888888, 99999999,
		12345678, 87654321:
		return false
	}
	return pc <= 99999998
}

// forgetEphemeral drops the credentials of the window this handler opened.
func (h *handler) forgetEphemeral() {
	h.mu.Lock()
	h.ephemeral = nil
	h.mu.Unlock()
}
