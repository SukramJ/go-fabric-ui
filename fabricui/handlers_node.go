// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/SukramJ/go-fabric/eligibility"
	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"

	"github.com/SukramJ/go-fabric-ui/internal/api"
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
	// ExposureReason is why the host serves no selection
	// (Options.ExposureReason); absent with a selection.
	ExposureReason string `json:"exposure_reason,omitempty"`
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
	if h.exposure == nil {
		res.ExposureReason = h.exposureReason
	}
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
	if err := h.reassemble(r.Context()); err != nil {
		// A silent success would tell the operator the list is correct
		// now, which is the one thing they came to check.
		h.serverError(w, r, "Failed to reassemble the endpoints", err)
		return
	}
	h.publish(Event{Type: EventEndpointsChanged, Data: struct{}{}})
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

// eventWindow is the since and limit of an event route (contract
// EventSince, EventLimit): since absent is 0, limit absent is
// defaultEventLimit and at most maxEventLimit.
const (
	defaultEventLimit = 100
	maxEventLimit     = 1000
)

func eventWindow(sinceQ, limitQ *int) (since uint64, limit int, err error) {
	if sinceQ != nil {
		if *sinceQ < 0 {
			return 0, 0, badRequest("since: %d is not an event number", *sinceQ)
		}
		since = uint64(*sinceQ)
	}
	limit = defaultEventLimit
	if limitQ != nil {
		if *limitQ < 1 || *limitQ > maxEventLimit {
			return 0, 0, badRequest("limit: %d is not within 1..%d", *limitQ, maxEventLimit)
		}
		limit = *limitQ
	}
	return since, limit, nil
}

// eventRecords is the contract's EventRecordList of the port's records.
func eventRecords(in []EventRecord) api.EventRecordList {
	out := api.EventRecordList{Events: make([]api.EventRecord, 0, len(in))}
	for i := range in {
		e := &in[i]
		r := api.EventRecord{
			Number: int(e.Number), Priority: api.EventRecordPriority(e.Priority), Endpoint: int(e.Endpoint), //nolint:gosec // an event number fits int64
			ClusterHex: hexID(e.Cluster), EventHex: hexID(e.Event), Name: optString(e.Name),
		}
		if e.EpochMS != nil {
			r.EpochMs = ptr(int(*e.EpochMS))
		}
		if e.Data != nil {
			r.Data = ptr(e.Data)
		}
		out.Events = append(out.Events, r)
	}
	return out
}

// queryInt is an optional integer query parameter; nil when absent.
func queryInt(r *http.Request, name string) (*int, error) {
	s := r.URL.Query().Get(name)
	if s == "" {
		return nil, nil //nolint:nilnil // absent is no value and no error
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil, badRequest("%s: %q is not an integer", name, s)
	}
	return &v, nil
}

// pathEndpoint is the {ep} path value as an endpoint number.
func pathEndpoint(r *http.Request) (uint16, error) {
	v, err := strconv.ParseUint(r.PathValue("ep"), 10, 16)
	if err != nil {
		return 0, badRequest("ep: %q is not an endpoint number", r.PathValue("ep"))
	}
	return uint16(v), nil
}

// writeAPIError answers a node route's error: an apiProblem as it says,
// ErrEndpointNotFound 404, node.ErrNotStarted 503, anything else 500.
func (h *handler) writeAPIError(w http.ResponseWriter, r *http.Request, title string, err error) {
	var p *apiProblem
	switch {
	case errors.As(err, &p):
		writeProblem(w, r, p.status, p.code, p.title, p.detail)
	case errors.Is(err, ErrEndpointNotFound):
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such endpoint", err.Error())
	case errors.Is(err, node.ErrNotStarted):
		unready(w, r, "Node not started", err.Error())
	default:
		h.serverError(w, r, title, err)
	}
}

// nodeEndpoint serves GET /node/endpoints/{ep}: one endpoint the node
// serves with its values (readEndpointDetail) and the surface report's
// decisions on it — the root's under the key "", the Aggregator's under
// node.AggregatorSurfaceKey, a bridged endpoint's under its stable key
// (go-fabric node/surface.go:33-36 at 894adbcd).
func (h *handler) nodeEndpoint(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	ep, err := pathEndpoint(r)
	if err != nil {
		h.writeAPIError(w, r, "", err)
		return
	}
	d, err := h.node.endpointDetail(r.Context(), ep)
	if err != nil {
		h.writeAPIError(w, r, "Failed to read the endpoint", err)
		return
	}
	out := endpointDetail(d)
	out.Scope = optString(d.Scope)
	decisions, err := h.decisionsOf(surfaceKeyOf(h.node.topology(), ep))
	if err != nil {
		h.serverError(w, r, "Failed to encode the surface report", err)
		return
	}
	out.Decisions = &decisions
	writeJSON(w, out)
}

// surfaceKeyOf is the key endpoint ep has in the node's surface report;
// ok false for an endpoint the report keys by no name of the topology.
func surfaceKeyOf(topo node.BridgeTopology, ep uint16) (key string, ok bool) {
	switch ep {
	case 0:
		return "", true
	case topo.Aggregator:
		return node.AggregatorSurfaceKey, true
	}
	if dev := findDevice(topo, ep); dev != nil && dev.Key != nil {
		return dev.Key.String(), true
	}
	return "", false
}

// decisionsOf are the surface report's decisions on the endpoint keyed
// key, in the report's encoding (surface.Report.MarshalJSON); none when ok
// is false.
func (h *handler) decisionsOf(key string, ok bool) ([]api.SurfaceDecision, error) {
	out := []api.SurfaceDecision{}
	if !ok {
		return out, nil
	}
	b, err := json.Marshal(h.node.surface())
	if err != nil {
		return nil, err
	}
	var rep struct {
		Decisions []api.SurfaceDecision `json:"decisions"`
	}
	if err := json.Unmarshal(b, &rep); err != nil {
		return nil, err
	}
	for i := range rep.Decisions {
		if rep.Decisions[i].Endpoint == key {
			out = append(out, rep.Decisions[i])
		}
	}
	return out, nil
}

// nodeEndpointEvents serves GET /node/endpoints/{ep}/events: the node's
// event log filtered to the endpoint (Node.Events, go-fabric
// node/served.go:159-184 at 894adbcd), oldest first.
func (h *handler) nodeEndpointEvents(w http.ResponseWriter, r *http.Request) {
	if h.node == nil {
		nodeUnready(w, r)
		return
	}
	ep, err := pathEndpoint(r)
	if err != nil {
		h.writeAPIError(w, r, "", err)
		return
	}
	since, err := queryInt(r, "since")
	if err != nil {
		h.writeAPIError(w, r, "", err)
		return
	}
	limit, err := queryInt(r, "limit")
	if err != nil {
		h.writeAPIError(w, r, "", err)
		return
	}
	s, l, err := eventWindow(since, limit)
	if err != nil {
		h.writeAPIError(w, r, "", err)
		return
	}
	writeJSON(w, eventRecords(h.node.eventLog(node.EventQuery{Endpoint: &ep, MinNumber: s, Limit: l})))
}
