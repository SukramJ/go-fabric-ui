// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Ported from openccu-loom internal/north/rest/handlers/matter_sessions_test.go
// and matter_events_test.go at b40b289a, plus the endpoint, mDNS and
// compatibility routes the daemon's adapters served there.

package fabricui

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// Ports TestMatterSessionsReportsPerPeerIdleAge.
func TestSessions_ReportPerPeerIdleAge(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-sessions.200")
}

// Ports TestMatterSessionsWithoutBridgeReports503.
func TestSessions_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-sessions.503")
}

// Ports TestMatterSessionsDistinguishesCommissioningSessions.
func TestSessions_MarkPASE(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	var got sessionList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/sessions", "").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Sessions[0].IsPASE || !got.Sessions[1].IsPASE {
		t.Fatalf("sessions %+v", got.Sessions)
	}
}

// Ports TestMatterSessionsReportsSessionTableOccupancy.
func TestSessions_ReportOccupancy(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	var got sessionList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/sessions", "").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Occupancy != (occupancyResponse{Live: 2, Reserved: 1, Capacity: 65535, Free: 65532}) {
		t.Fatalf("occupancy %+v", got.Occupancy)
	}
}

// Ports TestMatterDiagnosticEventsServesTheTraceNewestFirst.
func TestDiagnosticEvents_ServeTheTraceNewestFirst(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-events.200")
}

// Ports TestADisabledBridgeIsNotAnEmptyTrace.
func TestDiagnosticEvents_NoNode_Returns503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	if rec := do(t, mux, http.MethodGet, "/node/events", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", rec.Code)
	}
}

func TestEndpoints_ListBridgedDevicesAndParts(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-endpoints.200")
}

func TestMdns_ReportsTheAdvertisement(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-mdns.200")
}

func TestCompatibility_JoinsFabricsAndEndpoints(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	runFixture(t, mux, "node-compatibility.200")
}

func TestDiagnostics_NoNode_Return503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	for _, p := range []string{"/node/endpoints", "/node/mdns", "/node/compatibility"} {
		if rec := do(t, mux, http.MethodGet, p, ""); rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s: status %d, want 503", p, rec.Code)
		}
	}
}

func TestUnknownAPIRoute_Returns404Problem(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	rec := do(t, mux, http.MethodGet, "/node/nope", "")
	if rec.Code != http.StatusNotFound || problemOf(t, rec).Code != codeNotFound {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, mux, http.MethodPost, "/node/status", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("POST /node/status: %d, want the API 404", rec.Code)
	}
	if strings.Contains(do(t, mux, http.MethodGet, "/node/nope", "").Body.String(), "<html") {
		t.Fatal("an API 404 served the index")
	}
}
