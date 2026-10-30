// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/SukramJ/go-fabric/contract"
)

// verdictResponse is an item's eligibility on the wire: the state token of
// contract.EligibilityState.String and the device type's name.
type verdictResponse struct {
	State          string   `json:"state"`
	DeviceType     uint16   `json:"device_type"`
	DeviceTypeName string   `json:"device_type_name"`
	Clusters       []uint32 `json:"clusters"`
	Reason         string   `json:"reason"`
}

// itemResponse is one item of GET /node/exposure and of a preview.
type itemResponse struct {
	Key    string            `json:"key"`
	Name   string            `json:"name"`
	Group  string            `json:"group"`
	Fields map[string]string `json:"fields"`
	// Lists travels only when the provider fills it, so a host on Fields
	// alone answers byte for byte as before.
	Lists   map[string][]string `json:"lists,omitempty"`
	Verdict verdictResponse     `json:"verdict"`
	Enabled bool                `json:"enabled"`
	Source  string              `json:"source"`
}

type itemList struct {
	Items []itemResponse `json:"items"`
}

// updateList is the body of PUT /node/exposure and POST
// /node/exposure/preview.
type updateList struct {
	Items []Update `json:"items"`
}

// listsOf copies an item's lists for the response: a nil list answers
// [] rather than null, so every entry the page reads is an array.
func listsOf(in map[string][]string) map[string][]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]string, len(in))
	for k, v := range in {
		out[k] = append([]string{}, v...)
	}
	return out
}

func itemsOf(items []Item) []itemResponse {
	out := make([]itemResponse, 0, len(items))
	for i := range items {
		it := &items[i]
		fields := it.Fields
		if fields == nil {
			fields = map[string]string{}
		}
		clusters := append([]uint32{}, it.Verdict.Clusters...)
		out = append(out, itemResponse{
			Key: it.Key, Name: it.Name, Group: it.Group, Fields: fields, Lists: listsOf(it.Lists),
			Verdict: verdictResponse{
				State:          it.Verdict.State.String(),
				DeviceType:     it.Verdict.DeviceType,
				DeviceTypeName: contract.DeviceTypeName(it.Verdict.DeviceType),
				Clusters:       clusters,
				Reason:         it.Verdict.Reason,
			},
			Enabled: it.Enabled,
			Source:  it.Source,
		})
	}
	return out
}

func exposureUnready(w http.ResponseWriter, r *http.Request) {
	unready(w, r, "Selection not available", "the host serves no selection (Options.Exposure is nil)")
}

// exposureSchema serves GET /node/exposure/schema.
func (h *handler) exposureSchema(w http.ResponseWriter, r *http.Request) {
	if h.exposure == nil {
		exposureUnready(w, r)
		return
	}
	s, err := h.exposure.Schema(r.Context())
	if err != nil {
		h.serverError(w, r, "Failed to read the selection schema", err)
		return
	}
	if s.Columns == nil {
		s.Columns = []Column{}
	}
	if s.Filters == nil {
		s.Filters = []Filter{}
	}
	for i := range s.Filters {
		if s.Filters[i].Options == nil {
			s.Filters[i].Options = []FilterOption{}
		}
	}
	writeJSON(w, s)
}

// exposureList serves GET /node/exposure. Mirrors loom
// internal/north/rest/handlers/matter_exposures.go:86-150 (MatterExposable)
// with the host's merge of candidates and stored rows behind the provider.
func (h *handler) exposureList(w http.ResponseWriter, r *http.Request) {
	if h.exposure == nil {
		exposureUnready(w, r)
		return
	}
	items, err := h.exposure.List(r.Context())
	if err != nil {
		h.serverError(w, r, "Failed to list the selection", err)
		return
	}
	writeJSON(w, itemList{Items: itemsOf(items)})
}

// readUpdates decodes and checks an update list; false when it answered.
func (h *handler) readUpdates(w http.ResponseWriter, r *http.Request) ([]Update, bool) {
	var body updateList
	if err := decodeJSON(w, r, &body); err != nil {
		badBody(w, r, err)
		return nil, false
	}
	for i := range body.Items {
		if body.Items[i].Key == "" {
			writeProblem(w, r, http.StatusBadRequest, codeBadRequest, fmt.Sprintf("Invalid item %d", i), "key is required")
			return nil, false
		}
	}
	return body.Items, true
}

// providerError answers a provider's refusal: 400 for an unknown key or a
// malformed update, 500 (logged) otherwise.
func (h *handler) providerError(w http.ResponseWriter, r *http.Request, title string, err error) {
	if errors.Is(err, ErrUnknownItem) || errors.Is(err, ErrInvalidUpdate) {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, title, err.Error())
		return
	}
	h.serverError(w, r, title, err)
}

// exposurePreview serves POST /node/exposure/preview (admin): the verdicts
// the items would carry after the updates, before anything is written.
func (h *handler) exposurePreview(w http.ResponseWriter, r *http.Request) {
	if h.exposure == nil {
		exposureUnready(w, r)
		return
	}
	updates, ok := h.readUpdates(w, r)
	if !ok {
		return
	}
	if len(updates) == 0 {
		writeJSON(w, itemList{Items: []itemResponse{}})
		return
	}
	items, err := h.exposure.Preview(r.Context(), updates)
	if err != nil {
		h.providerError(w, r, "Failed to preview the selection", err)
		return
	}
	writeJSON(w, itemList{Items: itemsOf(items)})
}

// appliedResponse is the answer of PUT /node/exposure.
type appliedResponse struct {
	Applied int `json:"applied"`
}

// exposureApply serves PUT /node/exposure (admin): the host applies the
// updates, the node's endpoints are reassembled so the change takes effect
// before the response, and exposure_changed names the keys. Mirrors loom
// matter_exposures.go:253-311 (MatterExposeBulk): an empty batch is 204,
// a failed reassembly is logged and does not fail the persisted change.
func (h *handler) exposureApply(w http.ResponseWriter, r *http.Request) {
	if h.exposure == nil {
		exposureUnready(w, r)
		return
	}
	updates, ok := h.readUpdates(w, r)
	if !ok {
		return
	}
	if len(updates) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	applied, err := h.exposure.Apply(r.Context(), updates)
	if err != nil {
		h.providerError(w, r, "Failed to apply the selection", err)
		return
	}
	keys := make([]string, 0, len(updates))
	for i := range updates {
		keys = append(keys, updates[i].Key)
	}
	h.publish(Event{Type: EventExposureChanged, Data: ExposureEvent{Keys: keys}})
	if h.reassemble != nil {
		if err := h.reassemble(r.Context()); err != nil {
			h.log.WarnContext(r.Context(), "fabricui.exposure.reassemble_failed", slog.String("error", err.Error()))
		} else {
			h.publish(Event{Type: EventEndpointsChanged, Data: struct{}{}})
		}
	}
	h.log.InfoContext(r.Context(), "fabricui.exposure.applied", slog.Int("applied", applied))
	writeJSON(w, appliedResponse{Applied: applied})
}
