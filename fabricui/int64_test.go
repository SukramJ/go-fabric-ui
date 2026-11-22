// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"thingfabric.dev/go-fabric/node"

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

// A vendor-specific device type (0xFFF1xxxx, at or above 2^31) is a
// discovery filter value on a 32-bit host too: the query parameter is
// x-go-type int64 (contract 0.12.0, UR-P4-ctr-260).
func TestDiscoveryOfAHighDeviceTypeOnA32BitHost(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	rec := do(t, mux, http.MethodGet, "/controller/discovery?filter=device_type&value=4293984256&timeout=1", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("device_type 0xFFF10000: %d %s", rec.Code, rec.Body)
	}
	rec = do(t, mux, http.MethodGet, "/controller/discovery?filter=device_type&value=4294967296&timeout=1", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("device_type 2^32: %d %s, want 400", rec.Code, rec.Body)
	}
}

// A cluster whose every read failed lists its attributes by id, a
// manufacturer-specific id (at or above 2^31) last, on a 32-bit host too
// (UR-P4-ctr-262).
func TestFailedAttributesSortOnA32BitHost(t *testing.T) {
	got := failedAttributes(0x0006, []node.AttributeValue{
		{Attribute: 0xFFF10000, Status: 0x86},
		{Attribute: 0x0000, Status: 0x86},
	})
	if len(got) != 2 || got[0].ID != 0x0000 || got[1].ID != 0xFFF10000 {
		t.Fatalf("order %+v; want 0x0, 0xfff10000", got)
	}
}
