// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import "github.com/SukramJ/go-fabric-ui/fabricui"

// ProductKeysOf are the product keys the update check groups src's
// commissioned nodes by, through the real closure of newUpdates.
func ProductKeysOf(src fabricui.ControllerSource) []string {
	s, ok := src.(*source)
	if !ok {
		return nil
	}
	var out []string
	for _, n := range newUpdates(s, nil).nodes() {
		out = append(out, n.product)
	}
	return out
}
