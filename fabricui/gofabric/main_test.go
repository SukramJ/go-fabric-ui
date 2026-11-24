// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import (
	"os"
	"testing"

	"thingfabric.dev/go-fabric-ui/fabricui"
	"thingfabric.dev/go-fabric-ui/internal/statusguard"
)

// TestMain runs the adapter's tests through the status guard (decision
// 47): a status only the real adapter's errors produce must be declared
// for its operation too (UR-P4-test-33).
func TestMain(m *testing.M) {
	os.Exit(statusguard.Main(m, "../../contract/openapi.yaml", fabricui.DefaultPrefix))
}
