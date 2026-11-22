// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package statusguard

import (
	"slices"
	"testing"
)

const guardDoc = `paths:
  /node/status:
    get:
      responses:
        "200":
          description: ok
components:
`

// A recorded status on a path no contract operation names is reported,
// with the request's own path; a declared status of a named path and a
// path outside the prefixes are not (UR-P4-test-266, UR-P4-test-394).
func TestUndeclared_APathNoOperationNames(t *testing.T) {
	var r recorder
	r.observe("GET", "/api/v1/node/status", 200)
	r.observe("GET", "/api/v1/node/unnamed", 200)
	r.observe("GET", "/elsewhere/x", 500)
	got := r.undeclared(guardDoc, []string{"/api/v1"})
	want := []string{"GET /api/v1/node/unnamed answered 200, and no contract operation names it"}
	if !slices.Equal(got, want) {
		t.Fatalf("undeclared = %q, want %q", got, want)
	}
}
