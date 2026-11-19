// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package statusobs is the hook through which tests see every status the
// API answers. fabricui calls Observe when it is set; only tests set it
// (internal/statusguard), so the guard that each status is declared for
// its operation in the contract (decision 47) also sees the answers of
// the controller adapter's tests (UR-P4-test-33).
package statusobs

// Observe receives, when set, the method, the request path and the
// status of every API answer that is not a path-level 404 or 405.
var Observe func(method, path string, status int)
