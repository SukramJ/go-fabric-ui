// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package gofabric

import "github.com/SukramJ/go-fabric-ui/fabricui"

// WaitLabelPush waits until the label push worker of src is idle: every
// round asked for has run. The caller sets no label meanwhile.
func WaitLabelPush(src fabricui.ControllerSource) {
	src.(*source).pushes.Wait()
}
