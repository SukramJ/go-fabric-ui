// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Ported from openccu-loom internal/north/rest/handlers/matter_test.go,
// matter_exposures_test.go and matter_maintenance_test.go at b40b289a:
// the fabric, status, revoke and maintenance cases, renamed onto the node
// routes. Each test names the loom test it ports.

package fabricui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/node"
	"github.com/SukramJ/go-fabric/surface"
)

// Ports TestMatterStatus_NilReader_Returns200_Disabled.
func TestStatus_NoNode_Returns200Disabled(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true, admin: func(*http.Request) bool { return true }})
	runFixture(t, mux, "node-status.200-disabled")
}

// Ports TestMatterStatus_WithReader_Returns200_WithSnapshot.
func TestStatus_WithNode_Returns200WithSnapshot(t *testing.T) {
	n := sampleNode()
	n.open, n.duration = true, 300
	_, mux := newTestHandler(t, testOptions{
		node: n, exposure: newMemoryProvider(), admin: func(*http.Request) bool { return true },
		warnings: sampleWarnings,
	})
	runFixture(t, mux, "node-status.200")
}

// sampleWarnings are the host findings of the node-status.200 fixture:
// one with the time the host first saw it, one without.
func sampleWarnings(context.Context) []Warning {
	return []Warning{
		{Code: "label_unresolved", Message: `The label "Basement" names no item.`, Since: testNow.Add(-time.Hour)},
		{Code: "rule_unused", Message: "The rule for remotes matches no item."},
	}
}

// The warnings are omitted without the hook, an empty list when the hook
// finds nothing, and reported by a host without a node too.
func TestStatus_Warnings(t *testing.T) {
	get := func(o testOptions) map[string]json.RawMessage {
		t.Helper()
		o.admin = func(*http.Request) bool { return true }
		_, mux := newTestHandler(t, o)
		rec := do(t, mux, http.MethodGet, "/node/status", "")
		var m map[string]json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
			t.Fatal(err)
		}
		return m
	}
	if w, ok := get(testOptions{})["warnings"]; ok {
		t.Fatalf("warnings without the hook: %s, want absent", w)
	}
	if w := string(get(testOptions{warnings: func(context.Context) []Warning { return nil }})["warnings"]); w != "[]" {
		t.Fatalf("warnings of an empty hook: %q, want []", w)
	}
	var ws []Warning
	if err := json.Unmarshal(get(testOptions{noNode: true, warnings: sampleWarnings})["warnings"], &ws); err != nil || len(ws) != 2 {
		t.Fatalf("warnings without a node: %v %v, want the two findings", ws, err)
	}
}

// A node composed under the zero policies reports the pin's level, origin
// "pin", and its profiles and extensions as empty lists, not null; a Keep
// decision of a profile policy counts as a warning only with a warning
// reason, and a warning reason of a non-profile policy does not count.
func TestStatus_SurfaceOfTheZeroPolicies(t *testing.T) {
	n := sampleNode()
	n.rep = surface.Policies{}.Report()
	n.rep.Decisions = []surface.Decision{
		{Policy: "profile:alexa", Verdict: surface.Keep},
		{Policy: "custom", Verdict: surface.Keep, Reason: "warning (Reported) custom"},
		{Policy: "profile:alexa", Verdict: surface.Refuse, Reason: "warning (Reported) refused"},
	}
	_, mux := newTestHandler(t, testOptions{node: n, admin: func(*http.Request) bool { return true }})
	rec := do(t, mux, http.MethodGet, "/node/status", "")
	var m struct {
		Surface map[string]json.RawMessage `json:"surface"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"level": `"` + string(surface.PinLevel()) + `"`, "level_origin": `"pin"`,
		"profiles": "[]", "extensions": "[]", "assembled": "false",
		"withheld_count": "0", "warning_count": "0", "changed_count": "0",
	}
	if len(m.Surface) != len(want) {
		t.Fatalf("surface %v, want the keys of %v", m.Surface, want)
	}
	for k, v := range want {
		if got := string(m.Surface[k]); got != v {
			t.Errorf("surface.%s = %s, want %s", k, got, v)
		}
	}
}

func TestStatus_TokenModeIssuesTheToken(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	rec := do(t, mux, http.MethodGet, "/node/status", "")
	var s statusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatal(err)
	}
	if s.CSRFToken != testToken {
		t.Fatalf("csrf_token %q, want the handler's token", s.CSRFToken)
	}
	if s.Healthy == nil || !*s.Healthy {
		t.Fatalf("healthy = %v, want true from the probe", s.Healthy)
	}
}

// Ports TestMatterExposeUpdate_AttributesTheAuthenticatedSubject: the host's
// admin check decides, and then no token is issued.
func TestStatus_HostModeIssuesNoToken(t *testing.T) {
	h, err := newHandler(Options{Admin: func(*http.Request) bool { return true }})
	if err != nil {
		t.Fatal(err)
	}
	if h.token != "" {
		t.Fatalf("token %q issued although the host decides admin rights", h.token)
	}
	h2, err := newHandler(Options{})
	if err != nil {
		t.Fatal(err)
	}
	h3, err := newHandler(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(h2.token) != 32 || h2.token == h3.token {
		t.Fatalf("token mode tokens %q, %q: want 32 random hex digits per handler", h2.token, h3.token)
	}
}

// Ports TestMatterFabrics_NilStore_Returns503.
func TestFabrics_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-fabrics.503")
}

// Ports TestMatterFabrics_StoreError_Returns500: the cause stays in the log.
func TestFabrics_StoreError_Returns500(t *testing.T) {
	n := sampleNode()
	n.fabricsErr = errors.New("sqlite: /var/lib/node.db: locked")
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{node: n, log: &log})
	rec := do(t, mux, http.MethodGet, "/node/fabrics", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
	if p := problemOf(t, rec); p.Detail != "" || strings.Contains(rec.Body.String(), "/var/lib") {
		t.Fatalf("problem leaks the cause: %s", rec.Body)
	}
	if !strings.Contains(log.String(), "/var/lib/node.db") {
		t.Fatalf("the cause is not in the log: %s", log.String())
	}
}

// Ports TestMatterFabrics_EmptyList_Returns200: an empty list, not null.
func TestFabrics_EmptyList_Returns200(t *testing.T) {
	n := sampleNode()
	n.fabricList = nil
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := do(t, mux, http.MethodGet, "/node/fabrics", "")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"fabrics":[]}` {
		t.Fatalf("got %d %s, want 200 {\"fabrics\":[]}", rec.Code, rec.Body)
	}
}

// Ports TestMatterFabrics_SingleRecord_Returns200: every field, the vendor
// name from eligibility.VendorName, the keys as hex.
func TestFabrics_SingleRecord_Returns200(t *testing.T) {
	n := sampleNode()
	n.fabricList = n.fabricList[:1]
	_, mux := newTestHandler(t, testOptions{node: n})
	var got fabricList
	rec := do(t, mux, http.MethodGet, "/node/fabrics", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Fabrics) != 1 {
		t.Fatalf("fabrics %d, want 1", len(got.Fabrics))
	}
	f := got.Fabrics[0]
	if f.FabricIndex != 1 || f.VendorID != 0x1349 || f.VendorName == "" || f.Label != "Home" ||
		f.CompressedID != "9c71d38fbe48f2e5" || f.RootPublicKey != "04aabb" {
		t.Fatalf("fabric %+v", f)
	}
}

// Ports TestMatterFabrics_MultipleRecords_Returns200.
func TestFabrics_MultipleRecords_Returns200(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-fabrics.200")
}

// Ports TestMatterFabrics_LargeIDsSurviveAsHexOnTheWire: a node id above
// 2^53 is exact in its hex twin.
func TestFabrics_LargeIDsSurviveAsHexOnTheWire(t *testing.T) {
	n := sampleNode()
	n.fabricList = []node.FabricInfo{{FabricIndex: 1, FabricID: 0xFFFFFFFFFFFFFFFE, NodeID: 0xB2C3D4E5F6071829}}
	_, mux := newTestHandler(t, testOptions{node: n})
	rec := do(t, mux, http.MethodGet, "/node/fabrics", "")
	body := rec.Body.String()
	if !strings.Contains(body, `"node_id_hex":"B2C3D4E5F6071829"`) || !strings.Contains(body, `"fabric_id_hex":"FFFFFFFFFFFFFFFE"`) {
		t.Fatalf("hex ids missing or inexact: %s", body)
	}
}

// Ports TestMatterFabricRevoke_NilRevoker_Returns503.
func TestRemoveFabric_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodDelete, "/node/fabrics/1", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

// Ports TestMatterFabricRevoke_NonNumericID_Returns400.
func TestRemoveFabric_NonNumericIndex_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for _, idx := range []string{"abc", "-1", "256", "255"} {
		if rec := do(t, mux, http.MethodDelete, "/node/fabrics/"+idx, ""); rec.Code != http.StatusBadRequest {
			t.Fatalf("index %s: status %d, want 400", idx, rec.Code)
		}
	}
}

// Ports TestMatterFabricRevoke_ZeroID_Returns400.
func TestRemoveFabric_ZeroIndex_Returns400(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-fabric-remove.400")
}

// Ports TestMatterFabricRevoke_ValidID_Returns204: the fabric goes through
// the facade and fabric_removed is published.
func TestRemoveFabric_ValidIndex_Returns204(t *testing.T) {
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{node: n})
	rec := record(h)
	runFixture(t, mux, "node-fabric-remove.204")
	if !slices.Equal(n.removed, []uint8{2}) {
		t.Fatalf("removed %v, want [2]", n.removed)
	}
	if !slices.Equal(rec.types(), []string{EventFabricRemoved}) {
		t.Fatalf("events %v, want fabric_removed", rec.types())
	}
}

// Ports TestMatterFabricRevoke_RevokerError_Returns500.
func TestRemoveFabric_Error_Returns500(t *testing.T) {
	n := sampleNode()
	n.removeErr = errors.New("boom")
	_, mux := newTestHandler(t, testOptions{node: n})
	if rec := do(t, mux, http.MethodDelete, "/node/fabrics/1", ""); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
}

// Ports TestForceSyncReassemblesTheTopology.
func TestReassemble_RebuildsTheEndpoints(t *testing.T) {
	calls := 0
	h, mux := newTestHandler(t, testOptions{reassemble: func(context.Context) error { calls++; return nil }})
	rec := record(h)
	runFixture(t, mux, "node-maintenance-reassemble.204")
	if calls != 1 {
		t.Fatalf("reassemble ran %d times, want 1", calls)
	}
	if !slices.Equal(rec.types(), []string{EventEndpointsChanged}) {
		t.Fatalf("events %v, want endpoints_changed", rec.types())
	}
}

// Ports TestForceSyncReportsAFailedReassembly.
func TestReassemble_Failure_Returns500(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{reassemble: func(context.Context) error { return errors.New("assembler: bad spec") }})
	rec := do(t, mux, http.MethodPost, "/node/maintenance/reassemble", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
	if p := problemOf(t, rec); p.Code != codeInternal {
		t.Fatalf("code %q", p.Code)
	}
}

// Ports TestFactoryResetRefusesWithoutTheWrittenConfirmation: a generic
// "yes" and an empty body remove nothing.
func TestFactoryReset_RefusesWithoutTheWrittenConfirmation(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-maintenance-factory-reset.400")
	if rec := do(t, mux, http.MethodPost, "/node/maintenance/factory-reset", ""); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty body: status %d, want 400", rec.Code)
	}
	if len(n.removed) != 0 {
		t.Fatalf("removed %v without confirmation", n.removed)
	}
}

// Ports TestFactoryResetRemovesEveryFabricWhenConfirmed.
func TestFactoryReset_RemovesEveryFabricWhenConfirmed(t *testing.T) {
	n := sampleNode()
	h, mux := newTestHandler(t, testOptions{node: n})
	rec := record(h)
	runFixture(t, mux, "node-maintenance-factory-reset.204")
	if !slices.Equal(n.removed, []uint8{1, 2}) {
		t.Fatalf("removed %v, want [1 2]", n.removed)
	}
	if !slices.Equal(rec.types(), []string{EventFabricRemoved, EventFabricRemoved}) {
		t.Fatalf("events %v", rec.types())
	}
}

// Ports TestMaintenanceActionsNeedAnEnabledBridge.
func TestMaintenance_NeedsItsDependencies(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-maintenance-reassemble.503")
	if rec := do(t, mux, http.MethodPost, "/node/maintenance/factory-reset", `{"confirm":"remove-all-fabrics"}`); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("factory reset without a node: status %d, want 503", rec.Code)
	}
}

// Ports TestFactoryResetAuditsThePartOfTheResetThatLanded: the partial
// reset is reported with the count, and logged.
func TestFactoryReset_ReportsAPartialReset(t *testing.T) {
	n := sampleNode()
	n.removeFailAt = 2
	var log bytes.Buffer
	_, mux := newTestHandler(t, testOptions{node: n, log: &log})
	rec := do(t, mux, http.MethodPost, "/node/maintenance/factory-reset", `{"confirm":"remove-all-fabrics"}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d, want 500", rec.Code)
	}
	if p := problemOf(t, rec); p.Title != "Factory reset incomplete: removed 1 of 2 fabrics" {
		t.Fatalf("title %q", p.Title)
	}
	if !strings.Contains(log.String(), "fabricui.factory_reset.partial") {
		t.Fatalf("the partial reset is not logged: %s", log.String())
	}
}
