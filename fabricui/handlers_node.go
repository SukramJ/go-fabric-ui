// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/surface"
)

// statusResponse is the body of GET /node/status. Mirrors loom's
// MatterStatusResponse (internal/north/rest/handlers/matter_exposures.go:
// 350-360 at b40b289a) plus the health probe's verdict, whether the
// selection page is served and the token of token mode.
type statusResponse struct {
	Enabled        bool   `json:"enabled"`
	Listening      bool   `json:"listening"`
	ListenAddr     string `json:"listen_addr,omitempty"`
	Healthy        *bool  `json:"healthy,omitempty"`
	HealthNote     string `json:"health_note,omitempty"`
	EndpointCount  int    `json:"endpoint_count"`
	FabricCount    int    `json:"fabric_count"`
	EnabledCount   *int   `json:"enabled_count,omitempty"`
	Advertising    bool   `json:"advertising"`
	WindowOpen     bool   `json:"commissioning_window_open"`
	WindowDuration uint16 `json:"commissioning_window_duration_seconds,omitempty"`
	// WindowKind is which window is open (node.WindowKind: none, own,
	// basic, enhanced); absent without a node.
	WindowKind string `json:"window_kind,omitempty"`
	// WindowDeadline is when the open window's timer closes it, on the
	// node's clock (node.Window); absent while none is open.
	WindowDeadline string `json:"window_deadline,omitempty"`
	Exposure       bool   `json:"exposure"`
	CSRFToken      string `json:"csrf_token,omitempty"`
	// Warnings is nil, and omitted, without Options.Warnings; with the
	// hook it is a list, empty when the host finds nothing.
	Warnings []Warning `json:"warnings,omitzero"`
	// Surface summarises the policies the node's served surface is
	// composed under (Node.Surface); absent without a node.
	Surface *surfaceSummary `json:"surface,omitempty"`
}

// surfaceSummary is the head of the node's surface.Report and the counts
// a reader needs to know whether to look further. Profiles and Extensions
// are lists, empty rather than null.
type surfaceSummary struct {
	Level         string   `json:"level"`
	LevelOrigin   string   `json:"level_origin"`
	Profiles      []string `json:"profiles"`
	Extensions    []string `json:"extensions"`
	Assembled     bool     `json:"assembled"`
	WithheldCount int      `json:"withheld_count"`
	WarningCount  int      `json:"warning_count"`
	ChangedCount  int      `json:"changed_count"`
}

// summariseSurface reads the summary from a report. A profile warning is
// a Warn decision: a controller profile's policy keeps an element a warn
// rule matches — or a remove rule below Reproduced — with surface.Warn
// (go-fabric surface/profile.go:259-273, profilePolicy.Judge), and the
// verdict, not the reason's text, is what a reader counts
// (surface/policy.go:25-30).
func summariseSurface(r *surface.Report) *surfaceSummary {
	out := &surfaceSummary{
		Level:         string(r.Level),
		LevelOrigin:   r.LevelOrigin,
		Profiles:      append([]string{}, r.Profiles...),
		Extensions:    append([]string{}, r.Extensions...),
		Assembled:     r.Assembled,
		WithheldCount: len(r.Withheld),
		ChangedCount:  len(r.Changed),
	}
	for i := range r.Decisions {
		if r.Decisions[i].Verdict == surface.Warn {
			out.WarningCount++
		}
	}
	return out
}

// status serves GET /node/status. A host without a node is answered with
// 200 and enabled false, not 503: the status is what tells a page the
// family is off (loom matter_exposures.go:363-372).
func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	res := statusResponse{Exposure: h.exposure != nil, CSRFToken: h.token}
	if h.warnings != nil {
		// The host's findings are about its configuration, not the node,
		// so a host without a node reports them too.
		res.Warnings = h.warnings(r.Context())
		if res.Warnings == nil {
			res.Warnings = []Warning{}
		}
	}
	if h.node == nil {
		writeJSON(w, res)
		return
	}
	res.Enabled = true
	res.ListenAddr = h.node.localAddr()
	res.Listening = res.ListenAddr != ""
	if healthy, note, ok := h.node.health(); ok {
		res.Healthy = &healthy
		res.HealthNote = note
	}
	res.EndpointCount = len(h.node.topology().Devices)
	rep := h.node.surface()
	res.Surface = summariseSurface(&rep)
	res.Advertising = h.node.mdns().Advertising
	win := h.node.window()
	res.WindowOpen = win.Open
	res.WindowKind = win.Kind.String()
	if win.Open {
		res.WindowDuration = durationSeconds(win.Timeout)
		if !win.Deadline.IsZero() {
			res.WindowDeadline = win.Deadline.UTC().Format(time.RFC3339)
		}
	}
	if fabrics, err := h.node.fabrics(r.Context()); err == nil {
		res.FabricCount = len(fabrics)
	}
	if h.exposure != nil {
		if items, err := h.exposure.List(r.Context()); err == nil {
			n := 0
			for i := range items {
				if items[i].Enabled {
					n++
				}
			}
			res.EnabledCount = &n
		}
	}
	writeJSON(w, res)
}

// fabricResponse is one entry of GET /node/fabrics. Mirrors loom's
// MatterFabricResponse (matter.go:232-255): a JSON number carries 53 bits,
// so the 64-bit ids travel exactly as 16-digit uppercase hex, and the
// numeric fields are informational only.
type fabricResponse struct {
	FabricIndex   uint8  `json:"fabric_index"`
	FabricID      uint64 `json:"fabric_id"`
	FabricIDHex   string `json:"fabric_id_hex"`
	NodeID        uint64 `json:"node_id"`
	NodeIDHex     string `json:"node_id_hex"`
	VendorID      uint16 `json:"vendor_id"`
	VendorName    string `json:"vendor_name"`
	Label         string `json:"label,omitempty"`
	CompressedID  string `json:"compressed_id"`
	RootPublicKey string `json:"root_public_key"`
}

type fabricList struct {
	Fabrics []fabricResponse `json:"fabrics"`
}

// fabrics serves GET /node/fabrics. Mirrors loom matter.go:266-296
// (MatterFabrics) on node.FabricInfos.
func (h *handler) fabrics(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	infos, err := h.node.fabrics(r.Context())
	if err != nil {
		h.serverError(w, r, "Failed to list fabrics", err)
		return
	}
	out := make([]fabricResponse, 0, len(infos))
	for _, f := range infos {
		out = append(out, fabricResponse{
			FabricIndex:   f.FabricIndex,
			FabricID:      f.FabricID,
			FabricIDHex:   hex64(f.FabricID),
			NodeID:        f.NodeID,
			NodeIDHex:     hex64(f.NodeID),
			VendorID:      f.VendorID,
			VendorName:    eligibility.VendorName(f.VendorID),
			Label:         f.Label,
			CompressedID:  hex.EncodeToString(f.CompressedFabricID[:]),
			RootPublicKey: hex.EncodeToString(f.RootPublicKey),
		})
	}
	writeJSON(w, fabricList{Fabrics: out})
}

// removeFabric serves DELETE /node/fabrics/{index}: node.RemoveFabric is
// the RemoveFabric command's own path (the fabric's rows, sessions,
// subscriptions and records go). Mirrors loom matter_exposures.go:383-410
// (MatterFabricRevoke).
func (h *handler) removeFabric(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	idx, err := strconv.ParseUint(r.PathValue("index"), 10, 8)
	if err != nil || idx == 0 || idx == 255 {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid fabric index", "fabric_index must be 1..254")
		return
	}
	if err := h.node.removeFabric(r.Context(), uint8(idx)); err != nil {
		h.serverError(w, r, "Failed to remove fabric", err)
		return
	}
	// fabric_removed reaches the stream through the node's observer.
	h.log.InfoContext(r.Context(), "fabricui.fabric.removed", slog.Int("fabric_index", int(idx)))
	w.WriteHeader(http.StatusNoContent)
}

// maintenanceReassemble serves POST /node/maintenance/reassemble: it
// rebuilds the endpoints from the host's current selection without
// touching a pairing. Mirrors loom matter_maintenance.go:27-48
// (MatterForceSync).
func (h *handler) maintenanceReassemble(w http.ResponseWriter, r *http.Request) {
	if h.reassemble == nil {
		unready(w, r, "Reassembly not available", "the host wires no Options.Reassemble")
		return
	}
	before := h.announcedChanges()
	if err := h.reassemble(r.Context()); err != nil {
		// A silent success would tell the operator the list is correct
		// now, which is the one thing they came to check.
		h.serverError(w, r, "Failed to reassemble the endpoints", err)
		return
	}
	h.publish(Event{Type: EventEndpointsChanged, Data: struct{}{}})
	h.publishSurfaceChanged(before)
	w.WriteHeader(http.StatusNoContent)
}

// factoryResetConfirmation is the word the body has to carry. It names the
// action rather than agreeing with it, so a client replaying a generic
// {"confirm":"yes"} cannot unpair a house (loom matter_maintenance.go:58-61).
const factoryResetConfirmation = "remove-all-fabrics"

// factoryReset serves POST /node/maintenance/factory-reset: every fabric
// is removed and each controller has to commission the node again. The
// written confirmation is part of the contract. Mirrors loom
// matter_maintenance.go:67-116 (MatterFactoryReset).
func (h *handler) factoryReset(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	var body struct {
		Confirm string `json:"confirm"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Confirm != factoryResetConfirmation {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Factory reset not confirmed",
			`send {"confirm":"`+factoryResetConfirmation+`"} — this removes every fabric `+
				`and each paired controller has to commission the node again`)
		return
	}
	infos, err := h.node.fabrics(r.Context())
	if err != nil {
		h.serverError(w, r, "Failed to list fabrics", err)
		return
	}
	removed := make([]uint8, 0, len(infos))
	for _, f := range infos {
		if err := h.node.removeFabric(r.Context(), f.FabricIndex); err != nil {
			// Reported rather than skipped: a partial reset leaves the
			// node paired to a controller the operator believes removed.
			h.log.WarnContext(r.Context(), "fabricui.factory_reset.partial",
				slog.Int("removed", len(removed)), slog.Int("of", len(infos)), slog.Any("fabric_indexes", removed))
			h.serverError(w, r, fmt.Sprintf("Factory reset incomplete: removed %d of %d fabrics", len(removed), len(infos)), err)
			return
		}
		removed = append(removed, f.FabricIndex)
	}
	h.log.InfoContext(r.Context(), "fabricui.factory_reset.done", slog.Int("removed", len(removed)))
	w.WriteHeader(http.StatusNoContent)
}

// hex64 is a 64-bit id as the 16-digit uppercase hex controllers and
// chip-tool print.
func hex64(v uint64) string { return fmt.Sprintf("%016X", v) }
