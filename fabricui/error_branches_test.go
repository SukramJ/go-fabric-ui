// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"net/http"
	"strings"
	"testing"
)

// The error branches phase 0 of the review found without a test
// (UR-P0-test-01): each answers its problem's status and code.

// Without Options.Registry the rename and the delete answer 503
// service_unready (registry.go, registryTarget).
func TestRegistry_NoRegistry_RenameAndDeleteReturn503(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{})
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/nodes/kitchen/label", `{"label":"x"}`},
		{http.MethodDelete, "/nodes/kitchen", ""},
	} {
		rec := do(t, mux, c.method, c.path, c.body)
		if p := problemOf(t, rec); rec.Code != http.StatusServiceUnavailable || p.Code != codeServiceUnready ||
			p.Title != "Node registry not available" {
			t.Errorf("%s %s: %d %+v", c.method, c.path, rec.Code, p)
		}
	}
}

// A malformed body answers 400 bad_request and one over maxBodyBytes 413
// payload_too_large on the registry's writes (registry.go, readNodeLabel)
// and on PUT /operation/log/level (operations.go, setLogLevel).
func TestMalformedAndOversizedBodies(t *testing.T) {
	_, regMux, _ := registryHandler(t, 0)
	_, opsMux := newTestHandler(t, testOptions{operations: sampleOperations()})
	huge := `{"label":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	for _, c := range []struct {
		mux          *http.ServeMux
		method, path string
	}{
		{regMux, http.MethodPost, "/nodes"},
		{regMux, http.MethodPut, "/nodes/kitchen/label"},
		{opsMux, http.MethodPut, "/operation/log/level"},
	} {
		for _, b := range []struct {
			body   string
			status int
			code   string
		}{
			{`{"label":`, http.StatusBadRequest, codeBadRequest},
			{`{"unknown":1}`, http.StatusBadRequest, codeBadRequest},
			{huge, http.StatusRequestEntityTooLarge, codeTooLarge},
		} {
			rec := do(t, c.mux, c.method, c.path, b.body)
			if p := problemOf(t, rec); rec.Code != b.status || p.Code != b.code {
				t.Errorf("%s %s with %.20q: %d %+v", c.method, c.path, b.body, rec.Code, p)
			}
		}
	}
}
