// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/SukramJ/go-fabric/node"

	"github.com/SukramJ/go-fabric-ui/fabricui/gofabric"
)

// The update check's re-check groups the nodes by vendor, product and
// the numeric SoftwareVersion, as go-fabric's QueryUpdates groups them
// (controller/softwareupdate/manager.go:995 at v0.23.0), not by the
// SoftwareVersionString: two releases that share a display string are two
// products there. Through the real closure, on a commissioned node whose
// two versions differ (UR-P4-adpt-197).
func TestUpdates_ProductKeyIsTheNumericSoftwareVersion(t *testing.T) {
	s := newCtrlSiteWith(t, ctrlSiteOptions{configure: func(cfg *node.Config) {
		cfg.BasicInformation.SoftwareVersion = 7
		cfg.BasicInformation.SoftwareVersionString = "1.0-beta"
	}})
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	s.pairOne(ctx)
	if got, want := gofabric.ProductKeysOf(s.src), []string{"FFF1/8001/7"}; !slices.Equal(got, want) {
		t.Fatalf("product keys %q, want %q", got, want)
	}
}
