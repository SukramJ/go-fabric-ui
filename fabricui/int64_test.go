// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/SukramJ/go-fabric-ui/internal/api"
)

// A uint32 at or above 2^31 stays a positive JSON number on a 32-bit
// host (GOARCH=386/arm, the armv7 add-on): the contract's fields of that
// range are format int64 (0.11.0, UR-P4-ctr-30). make test-386 runs this
// package under GOARCH=386.
func TestUint32FieldsSurviveA32BitHost(t *testing.T) {
	dv := uint32(0xDEADBEEF)
	b, err := json.Marshal(clusterDetail(ClusterDetail{ID: 0xFFF1FC00, DataVersion: &dv}))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"data_version":3735928559`, `"id":4294048768`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("want %s in %s", want, b)
		}
	}
}

// The consent of a version at or above 2^31 decodes on a 32-bit host,
// as the update list offers it (UR-P4-ctr-30).
func TestConsentOfAHighVersionDecodesOnA32BitHost(t *testing.T) {
	var b api.UpdateConsentRequest
	if err := decodeStrict([]byte(`{"node_id_hex":"0000000000000001","vendor_id":1,"product_id":1,"target_software_version":3000000000}`), &b); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if b.TargetSoftwareVersion != 3_000_000_000 {
		t.Errorf("version %d", b.TargetSoftwareVersion)
	}
}
