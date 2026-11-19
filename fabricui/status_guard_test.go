// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"os"
	"testing"

	"github.com/SukramJ/go-fabric-ui/internal/statusguard"
)

// TestMain runs the package's tests through the status guard (decision
// 47): every status a route answered must be declared for its operation
// in contract/openapi.yaml. A path the contract does not name, or a
// method it does not list there, is answered at path level (404, 405)
// and not checked.
func TestMain(m *testing.M) {
	os.Exit(statusguard.Main(m, "../contract/openapi.yaml", DefaultPrefix, "/ui-api"))
}

// scanOperations are the contract's operations (statusguard.ScanOperations).
func scanOperations(doc string) []contractOperation {
	ops := statusguard.ScanOperations(doc)
	out := make([]contractOperation, 0, len(ops))
	for _, op := range ops {
		out = append(out, contractOperation{method: op.Method, path: op.Path, body: op.Body, statuses: op.Statuses})
	}
	return out
}
