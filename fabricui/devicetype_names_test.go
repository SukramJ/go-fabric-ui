// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import "testing"

func TestDeviceTypeName(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   uint16
		want string
	}{
		{"no device type", 0, ""},
		{"operator label", 0x0100, "On/Off Light"},
		{"schema name", 0x0301, "Thermostat"},
		{"unknown id in hex", 0xABCD, "0xABCD"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := deviceTypeName(tc.id); got != tc.want {
				t.Errorf("deviceTypeName(0x%04X) = %q, want %q", tc.id, got, tc.want)
			}
		})
	}
}
