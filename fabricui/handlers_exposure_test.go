// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Ported from openccu-loom internal/north/rest/handlers/matter_exposures_test.go
// at b40b289a: the exposable, expose-update and bulk cases, on the
// provider-driven /node/exposure routes. The host-specific cases (the
// five-part key, the kinds, the parameter labeller, the sort order) are
// the provider's now and port to what the handler still decides.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// Ports TestMatterExposable_NilProvider_Returns503.
func TestExposureSchema_NoProvider_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-exposure-schema.503")
}

// Ports TestMatterExposable_NilStore_Returns503.
func TestExposureList_NoProvider_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	if rec := do(t, mux, http.MethodGet, "/node/exposure", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

func TestExposureSchema_Returns200(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	runFixture(t, mux, "node-exposure-schema.200")
}

// Ports TestMatterCandidateProviderFunc_Nil: a provider's nil lists travel
// as empty lists.
func TestExposureSchema_FillsEmptyLists(t *testing.T) {
	p := newMemoryProvider()
	p.schema = Schema{Filters: []Filter{{Field: "x", Label: "X"}}}
	_, mux := newTestHandler(t, testOptions{exposure: p})
	body := strings.TrimSpace(do(t, mux, http.MethodGet, "/node/exposure/schema", "").Body.String())
	if body != `{"columns":[],"groups":false,"filters":[{"field":"x","label":"X","options":[]}]}` {
		t.Fatalf("schema %s", body)
	}
}

// Ports TestMatterExposable_HappyPath_Returns200_MergesData.
func TestExposureList_Returns200(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	runFixture(t, mux, "node-exposure.200")
}

// Ports TestMatterCandidateProviderFunc_NonNil: an item without fields or
// clusters still carries the objects the page reads.
func TestExposureList_EmptyFieldsAreObjects(t *testing.T) {
	p := newMemoryProvider()
	p.items = []Item{{Key: "k"}}
	_, mux := newTestHandler(t, testOptions{exposure: p})
	body := do(t, mux, http.MethodGet, "/node/exposure", "").Body.String()
	if !strings.Contains(body, `"fields":{}`) || !strings.Contains(body, `"clusters":[]`) || !strings.Contains(body, `"state":"unmappable"`) {
		t.Fatalf("item %s", body)
	}
}

// Ports TestMatterExposable_SortOrder: the order is the provider's.
func TestExposureList_KeepsTheProviderOrder(t *testing.T) {
	p := newMemoryProvider()
	slices.Reverse(p.items)
	_, mux := newTestHandler(t, testOptions{exposure: p})
	var got itemList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/exposure", "").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(got.Items))
	for _, it := range got.Items {
		keys = append(keys, it.Key)
	}
	if !slices.Equal(keys, []string{"remote-3", "sensor-2", "lamp-1"}) {
		t.Fatalf("order %v", keys)
	}
}

// Ports TestMatterExposable_ParameterLabel_TitleCasedWhenLabelerIsNil: the
// verdict's device type is named by go-fabric, not by the page.
func TestExposureList_VerdictCarriesTheDeviceTypeName(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	var got itemList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/exposure", "").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Items[0].Verdict.DeviceTypeName == "" || got.Items[0].Verdict.DeviceTypeName == "0x0100" {
		t.Fatalf("device type name %q", got.Items[0].Verdict.DeviceTypeName)
	}
	if got.Items[2].Verdict.DeviceTypeName != "" {
		t.Fatalf("an unmappable item names device type %q", got.Items[2].Verdict.DeviceTypeName)
	}
}

// Ports TestMatterExposable_StoreListError_Returns500.
func TestExposureList_ProviderError_Returns500(t *testing.T) {
	p := newMemoryProvider()
	p.listErr = errors.New("db gone")
	_, mux := newTestHandler(t, testOptions{exposure: p})
	if rec := do(t, mux, http.MethodGet, "/node/exposure", ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}

// Ports TestMatterExposeUpdate_NilStore_Returns503 and
// TestMatterExposeBulk_NilStore_Returns503.
func TestExposureWrites_NoProvider_Return503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for _, route := range [][2]string{{http.MethodPut, "/node/exposure"}, {http.MethodPost, "/node/exposure/preview"}} {
		if rec := do(t, mux, route[0], route[1], `{"items":[]}`); rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s %s: status %d, want 503", route[0], route[1], rec.Code)
		}
	}
}

// Ports TestMatterExposeUpdate_BadBody_Returns400.
func TestExposureApply_BadBody_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	for _, body := range []string{`{`, `{"items":{}}`, `{"items":[],"extra":1}`} {
		if rec := do(t, mux, http.MethodPut, "/node/exposure", body); rec.Code != http.StatusBadRequest {
			t.Fatalf("body %s: status %d, want 400", body, rec.Code)
		}
	}
}

// Ports TestMatterExposeUpdate_InvalidKey_Returns400.
func TestExposureApply_EmptyKey_Returns400(t *testing.T) {
	p := newMemoryProvider()
	_, mux := newTestHandler(t, testOptions{exposure: p})
	runFixture(t, mux, "node-exposure-apply.400")
	if len(p.applied) != 0 {
		t.Fatal("the provider was asked to apply an invalid batch")
	}
}

// Ports TestMatterExposeBulk_BadItemKey_Returns400: the provider's unknown
// key is the caller's error.
func TestExposureApply_UnknownKey_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	rec := do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"nope","enabled":true}]}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(problemOf(t, rec).Detail, "unknown item") {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Ports TestMatterExposeBulk_EmptyItems_Returns204.
func TestExposureApply_EmptyBatch_Returns204(t *testing.T) {
	p := newMemoryProvider()
	_, mux := newTestHandler(t, testOptions{exposure: p})
	runFixture(t, mux, "node-exposure-apply.204")
	if len(p.applied) != 0 {
		t.Fatal("an empty batch reached the provider")
	}
}

// Ports TestMatterExposeBulk_HappyPath_Returns200_AppliedCount.
func TestExposureApply_Returns200WithTheAppliedCount(t *testing.T) {
	p := newMemoryProvider()
	_, mux := newTestHandler(t, testOptions{exposure: p})
	runFixture(t, mux, "node-exposure-apply.200")
	items, _ := p.List(context.Background())
	if !items[0].Enabled || items[1].Enabled || items[1].Name != "Hall sensor" {
		t.Fatalf("items after apply %+v", items)
	}
}

// Ports TestMatterExposeUpdate_HappyPath_Returns204_PublishesEvent and
// TestMatterExposeBulk_PublishesOneFramePerAffectedRow: one
// exposure_changed naming every key, then endpoints_changed after the
// reassembly.
func TestExposureApply_PublishesExposureChangedWithEveryKey(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider(), reassemble: func(context.Context) error { return nil }})
	rec := record(h)
	do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true},{"key":"sensor-2","enabled":false}]}`)
	if !slices.Equal(rec.types(), []string{EventExposureChanged, EventEndpointsChanged}) {
		t.Fatalf("events %v", rec.types())
	}
	if ev, ok := rec.events[0].Data.(ExposureEvent); !ok || !slices.Equal(ev.Keys, []string{"lamp-1", "sensor-2"}) {
		t.Fatalf("payload %+v", rec.events[0].Data)
	}
}

// Ports TestRecordMatterAudit_NonNilRecorder: the change takes effect before
// the response — the endpoints are reassembled synchronously.
func TestExposureApply_Reassembles(t *testing.T) {
	calls := 0
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider(), reassemble: func(context.Context) error { calls++; return nil }})
	do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`)
	if calls != 1 {
		t.Fatalf("reassembled %d times, want 1", calls)
	}
}

// Ports TestMatterExposeBulk_StoreError_Returns500: a failed reassembly is
// logged and does not fail the persisted change.
func TestExposureApply_ReassembleFailureKeepsTheChange(t *testing.T) {
	h, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider(), reassemble: func(context.Context) error { return errors.New("assembler") }})
	rec := record(h)
	if r := do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`); r.Code != http.StatusOK {
		t.Fatalf("status %d", r.Code)
	}
	if !slices.Equal(rec.types(), []string{EventExposureChanged}) {
		t.Fatalf("events %v, want exposure_changed only", rec.types())
	}
}

// Ports TestMatterExposeUpdate_StoreError_Returns500.
func TestExposureApply_ProviderError_Returns500(t *testing.T) {
	p := newMemoryProvider()
	p.applyErr = errors.New("disk full")
	_, mux := newTestHandler(t, testOptions{exposure: p})
	rec := do(t, mux, http.MethodPut, "/node/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "disk full") {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestExposureApply_IsAdminOnly(t *testing.T) {
	p := newMemoryProvider()
	_, mux := newTestHandler(t, testOptions{exposure: p})
	runFixture(t, mux, "node-exposure-apply.403")
	if len(p.applied) != 0 {
		t.Fatal("a refused request reached the provider")
	}
}

// Ports TestValidateExposureKey_AllKnownKinds: the preview shows the
// verdicts after the change.
func TestExposurePreview_Returns200(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	runFixture(t, mux, "node-exposure-preview.200")
}

// Ports TestValidateExposureKey_UnknownKind.
func TestExposurePreview_UnknownKey_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	runFixture(t, mux, "node-exposure-preview.400")
}

// Ports TestValidateExposureKey_EmptyFields: an unmappable item cannot be
// enabled, and the preview says so as 400.
func TestExposurePreview_InvalidUpdate_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider()})
	rec := do(t, mux, http.MethodPost, "/node/exposure/preview", `{"items":[{"key":"remote-3","enabled":true}]}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(problemOf(t, rec).Detail, "cannot be exposed") {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

func TestExposurePreview_ChangesNothing(t *testing.T) {
	p := newMemoryProvider()
	h, mux := newTestHandler(t, testOptions{exposure: p, reassemble: func(context.Context) error { t.Fatal("a preview reassembled"); return nil }})
	rec := record(h)
	do(t, mux, http.MethodPost, "/node/exposure/preview", `{"items":[{"key":"lamp-1","enabled":true,"name":"Pantry"}]}`)
	items, _ := p.List(context.Background())
	if items[0].Enabled || items[0].Name != "Kitchen lamp" || len(rec.types()) != 0 {
		t.Fatalf("a preview changed state: %+v, events %v", items[0], rec.types())
	}
	if body := do(t, mux, http.MethodPost, "/node/exposure/preview", `{"items":[]}`).Body.String(); strings.TrimSpace(body) != `{"items":[]}` {
		t.Fatalf("empty preview %s", body)
	}
}
