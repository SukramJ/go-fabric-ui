// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Ported from openccu-loom internal/north/rest/handlers/matter_sessions_test.go
// and matter_events_test.go at b40b289a, plus the endpoint, mDNS and
// compatibility routes the daemon's adapters served there.

package fabricui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// The sessions carry the session parameters the node reports
// (Node.SessionParameters, contract 0.7.0): the specification version as
// 8 hex digits, the TCP size only with TCP.
func TestSessions_ReportTheSessionParameters(t *testing.T) {
	n := sampleNode()
	n.params.SupportedTransports = 0x04 // TCP server
	n.params.MaxTCPMessageSize = 64000
	_, mux := newTestHandler(t, testOptions{node: n})
	var got sessionList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/sessions", "").Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	p := got.Parameters
	if p == nil || p.SpecificationVersionHex != "01050000" || !p.TCP || p.MaxTCPMessageSize == nil || *p.MaxTCPMessageSize != 64000 {
		t.Fatalf("parameters %+v", p)
	}
	_, mux = newTestHandler(t, testOptions{})
	if strings.Contains(do(t, mux, http.MethodGet, "/node/sessions", "").Body.String(), "max_tcp_message_size") {
		t.Fatal("max_tcp_message_size without TCP")
	}
}

// POST /node/mdns/reannounce asks the node to announce now (contract
// 0.7.0): an admin route, 503 without a node, 404 for an unknown key.
func TestReannounce(t *testing.T) {
	n := sampleNode()
	_, mux := newTestHandler(t, testOptions{node: n})
	runFixture(t, mux, "node-mdns-reannounce.204")
	if n.reannounceCount() != 1 {
		t.Fatalf("reannounced %d times, want 1", n.reannounceCount())
	}
	runFixture(t, mux, "nodes-mdns-reannounce.404")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, DefaultPrefix+"/node/mdns/reannounce", http.NoBody)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden || n.reannounceCount() != 1 {
		t.Fatalf("without the token: %d, reannounced %d times", rec.Code, n.reannounceCount())
	}
	_, mux = newTestHandler(t, testOptions{noNode: true})
	runFixture(t, mux, "node-mdns-reannounce.503")
}
