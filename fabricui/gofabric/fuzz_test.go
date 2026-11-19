// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

// The standing fuzz targets of the adapter (review decision 53): the
// phase 0 targets as permanent Fuzz functions. `go test` runs each on its
// seeds; `make fuzz` runs each for FUZZTIME.

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/SukramJ/go-fabric/client/discovery"
	"github.com/SukramJ/go-fabric/controller"
	"github.com/SukramJ/go-fabric/controller/client"

	"github.com/SukramJ/go-fabric-ui/fabricui"
)

// FuzzThreadDatasetName: a dataset that does not decode, or names no
// network, is an error wrapping ErrInvalidCredential; one that does
// names a network.
func FuzzThreadDatasetName(f *testing.F) {
	for _, s := range []string{
		"03044d657368" + "01021234", "0304", "01021234", "", "00", "0300", "0301ff", "03104d6573684d6573684d6573684d657368",
		"0e080000000000010000" + "000300000f" + "0102face" + "030a4f70656e546872656164",
	} {
		b, err := hex.DecodeString(s)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, dataset []byte) {
		name, err := ThreadDatasetName(dataset)
		if err != nil {
			if !errors.Is(err, fabricui.ErrInvalidCredential) {
				t.Fatalf("%x: %v does not wrap ErrInvalidCredential", dataset, err)
			}
			return
		}
		if name == "" {
			t.Fatalf("%x: accepted without a network name", dataset)
		}
	})
}

// FuzzDiscoveredAndCode: discoveredOf's vendor and product from the mDNS
// VP value agree with go-fabric's ParseTxtDecimal, with no product
// without a "+"; bleFor after ParsePairingCode uses BLE exactly when there
// is a central and a network, the request is not network-only and the
// code allows BLE (a manual code, or a QR code with the BLE bit), and
// states a reason otherwise.
func FuzzDiscoveredAndCode(f *testing.F) {
	for _, s := range [][2]string{
		{"4996+22", "34970112332"},
		{"65521+32769", "MT:Y.K9042C00KA0648G00"},
		{"65536+1", "MT:-24J0AFN00KA0648G00"},
		{"+", "749701123365521327694"},
		{"12", "MT:"},
		{"1+2+3", ""},
		{" 1+ 2", "3497-011-2332"},
		{"0x10+010", "MT:Y.K90AFN00KA0648G00"},
	} {
		f.Add(s[0], s[1], true, false, true)
		f.Add(s[0], s[1], false, true, false)
	}
	f.Fuzz(func(t *testing.T, vp, code string, central, networkOnly, network bool) {
		d := discoveredOf(&client.Device{VP: &vp})
		vendor, product, plus := strings.Cut(vp, "+")
		wantV, okV := discovery.ParseTxtDecimal(vendor, 0xFFFF)
		wantP, okP := discovery.ParseTxtDecimal(product, 0xFFFF)
		if !okV {
			wantV = 0
		}
		if !okP || !plus {
			wantP = 0
		}
		if uint64(d.VendorID) != wantV || uint64(d.ProductID) != wantP {
			t.Fatalf("VP %q: vendor %d product %d, want %d %d", vp, d.VendorID, d.ProductID, wantV, wantP)
		}
		pc, err := controller.ParsePairingCode(code)
		if err != nil {
			return
		}
		got := bleFor(central, pc, networkOnly, network)
		allowed := pc.Manual || pc.Capabilities&capabilityBLE != 0
		want := central && network && !networkOnly && allowed
		if got.Used != want || got.Used == (got.Reason != "") {
			t.Fatalf("code %q (manual %v, capabilities %#x) central %v network-only %v network %v: %+v",
				code, pc.Manual, pc.Capabilities, central, networkOnly, network, got)
		}
	})
}
