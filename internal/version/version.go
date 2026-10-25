// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package version holds the release version of the module. The Go module
// and the npm package @sukramj/fabric-ui share it; make version-check ties
// it to ui/package.json and the changelog.
package version

// Version is the release version. Bumping it, ui/package.json's version
// and the changelog section happen in one commit.
const Version = "0.1.2"
