// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
)

// contractOperation is one operation of contract/openapi.yaml.
type contractOperation struct {
	method, path string
	body         bool
	statuses     []string
}

// contractOps are the operations of contract/openapi.yaml (scanOperations).
func contractOps(t *testing.T) []contractOperation {
	t.Helper()
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	out := scanOperations(string(b))
	if len(out) < 90 {
		t.Fatalf("found %d operations; the scan is not reading the contract", len(out))
	}
	return out
}

// samplePath fills a contract path's parameters with values a route
// pattern matches.
func samplePath(p string) string {
	return strings.NewReplacer("{node}", "second", "{id}", "0000000000000011", "{ep}", "1",
		"{index}", "1", "{cluster_hex}", "0x0006").Replace(p)
}

// Every contract path answers a method it does not list with 405 and an
// Allow of exactly its methods, plus HEAD where GET is (decision 60).
func TestEveryPathAnswers405WithItsMethods(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), operations: sampleOperations(), second: secondNode()})
	methods := map[string][]string{}
	for _, op := range contractOps(t) {
		methods[op.path] = append(methods[op.path], op.method)
	}
	for path, ms := range methods {
		want := slices.Clone(ms)
		if slices.Contains(want, http.MethodGet) {
			want = append(want, http.MethodHead)
		}
		slices.Sort(want)
		rec := do(t, mux, http.MethodPatch, samplePath(path), "")
		if rec.Code != http.StatusMethodNotAllowed || problemOf(t, rec).Code != codeMethodNotAllowed {
			t.Errorf("PATCH %s: %d %s", path, rec.Code, rec.Body)
			continue
		}
		if got := rec.Header().Get("Allow"); got != strings.Join(want, ", ") {
			t.Errorf("PATCH %s: Allow %q, want %q", path, got, strings.Join(want, ", "))
		}
	}
	for _, p := range []string{"/nothing", "/node/nothing", "/controller/nothing", "/operation/nothing"} {
		if rec := do(t, mux, http.MethodGet, p, ""); rec.Code != http.StatusNotFound || rec.Header().Get("Allow") != "" {
			t.Errorf("GET %s: %d Allow %q", p, rec.Code, rec.Header().Get("Allow"))
		}
	}
}

// controllerBodies names every controller operation with a request body.
func TestControllerBodiesMatchTheContract(t *testing.T) {
	var want []string
	for _, op := range contractOps(t) {
		if op.body && strings.HasPrefix(op.path, "/controller/") {
			want = append(want, op.method+" "+op.path)
		}
	}
	slices.Sort(want)
	got := slices.Sorted(func(yield func(string) bool) {
		for k := range controllerBodies {
			if !yield(k) {
				return
			}
		}
	})
	if !slices.Equal(got, want) {
		t.Fatalf("controllerBodies %v\ncontract %v", got, want)
	}
}

// Every body is strict: data after the value, a member named in another
// case and an unknown nested member are 400, on both families.
func TestBodies_AreStrict(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{controller: NewMemoryController(), operations: sampleOperations()})
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/operation/log/level", `{"level":"info"} {}`},
		{http.MethodPut, "/operation/log/level", `{"Level":"info"}`},
		{http.MethodPost, "/node/maintenance/factory-reset", `{"confirm":"remove-all-fabrics"}x`},
		{http.MethodPost, "/node/maintenance/factory-reset", `{"Confirm":"remove-all-fabrics"}`},
		{http.MethodPut, "/controller/fabric/label", `{"label":"a","extra":1}`},
		{http.MethodPut, "/controller/fabric/label", `{"LABEL":"a"}`},
		{http.MethodPut, "/controller/fabric/label", `{"label":"a"}{"label":"b"}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","terms_and_conditions":{"version":1,"acknowledgements":1,"x":1}}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","terms_and_conditions":{"Version":1,"acknowledgements":1}}`},
	} {
		rec := do(t, mux, c.method, c.path, c.body)
		if rec.Code != http.StatusBadRequest || problemOf(t, rec).Code != codeBadRequest {
			t.Errorf("%s %s %s: %d %s", c.method, c.path, c.body, rec.Code, rec.Body)
		}
	}
}

// wireOf is a controller event's payload on the wire without the
// commissioning's id, which is drawn per request.
func wireOf(data any) string {
	if d, ok := data.(controllerEventData); ok {
		d.CommissioningID = nil
		data = d
	}
	return string(mustJSON(data))
}

// Every commissioning_progress of one request carries its id: the
// client's when it names one, else one the handler draws; a malformed id
// is 400 (contract 0.10.0).
func TestCommission_CommissioningIDOnEveryEvent(t *testing.T) {
	h, mux, _ := controllerHandler(t, nil)
	var got []controllerEventData
	defer h.ctrlHub.Subscribe(func(ev Event) {
		if d, ok := ev.Data.(controllerEventData); ok && ev.Type == EventCommissioningProgress {
			got = append(got, d)
		}
	})()
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332","commissioning_id":"page-1"}`); r.Code != http.StatusOK {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
	if len(got) < 2 {
		t.Fatalf("%d events", len(got))
	}
	for _, d := range got {
		if d.CommissioningID == nil || *d.CommissioningID != "page-1" {
			t.Fatalf("event %s", mustJSON(d))
		}
	}
	got = nil
	h, mux, _ = controllerHandler(t, nil)
	defer h.ctrlHub.Subscribe(func(ev Event) {
		if d, ok := ev.Data.(controllerEventData); ok && ev.Type == EventCommissioningProgress {
			got = append(got, d)
		}
	})()
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332"}`); r.Code != http.StatusOK {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
	if len(got) == 0 || got[0].CommissioningID == nil || *got[0].CommissioningID == "" || *got[len(got)-1].CommissioningID != *got[0].CommissioningID {
		t.Fatalf("drawn ids %s", mustJSON(got))
	}
	for _, body := range []string{
		`{"code":"34970112332","commissioning_id":""}`, `{"code":"34970112332","commissioning_id":"a b"}`,
		`{"discovered":{"instance_name":"X"},"passcode":12345678}`, `{"discovered":{"instance_name":"X"},"passcode":11111111}`,
	} {
		if r := do(t, mux, http.MethodPost, "/controller/commission", body); r.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %s", body, r.Code, r.Body)
		}
	}
}

// A fabric label is bounded by 32 bytes, not characters (UR-P1-adpt-40).
func TestFabricLabel_BoundedInBytes(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	if r := do(t, mux, http.MethodPut, "/controller/fabric/label", `{"label":"`+strings.Repeat("ä", 17)+`"}`); r.Code != http.StatusBadRequest {
		t.Fatalf("34 bytes: %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodPut, "/controller/fabric/label", `{"label":"`+strings.Repeat("ä", 16)+`"}`); r.Code != http.StatusNoContent {
		t.Fatalf("32 bytes: %d %s", r.Code, r.Body)
	}
}

// Removing a fabric the node does not have is 404 (decision 43).
func TestRemoveFabric_UnknownIndexIs404(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	if r := do(t, mux, http.MethodDelete, "/node/fabrics/77", ""); r.Code != http.StatusNotFound {
		t.Fatalf("%d %s", r.Code, r.Body)
	}
}

// since_str selects as since does, exactly beyond 2^53, and wins over
// since; a malformed one is 400 (contract 0.10.0, EventSinceStr).
func TestEvents_SinceStr(t *testing.T) {
	n := sampleNode()
	n.logged = sampleEventLog()
	_, mux := newTestHandler(t, testOptions{node: n})
	all := do(t, mux, http.MethodGet, "/node/endpoints/2/events", "").Body.String()
	if got := do(t, mux, http.MethodGet, "/node/endpoints/2/events?since=999&since_str=0", "").Body.String(); got != all {
		t.Fatalf("since_str=0 %s, want %s", got, all)
	}
	if got := do(t, mux, http.MethodGet, "/node/endpoints/2/events?since_str=18446744073709551615", "").Body.String(); strings.TrimSpace(got) != `{"events":[]}` {
		t.Fatalf("since_str=max %s", got)
	}
	for _, q := range []string{"", "01", "-1", "18446744073709551616", "1e3"} {
		if r := do(t, mux, http.MethodGet, "/node/endpoints/2/events?since_str="+q, ""); r.Code != http.StatusBadRequest {
			t.Errorf("since_str=%q: %d %s", q, r.Code, r.Body)
		}
	}
	_, cmux, _ := controllerHandler(t, nil)
	if r := do(t, cmux, http.MethodGet, "/controller/nodes/0000000000000011/events?since_str=x", ""); r.Code != http.StatusBadRequest {
		t.Fatalf("controller since_str=x: %d %s", r.Code, r.Body)
	}
}
