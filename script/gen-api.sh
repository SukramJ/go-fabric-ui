#!/bin/sh
# Generates the controller family's router, request and response types
# from contract/openapi.yaml with the pinned oapi-codegen (go.mod tool
# directive) and writes them to $1 (default internal/api/api.gen.go) under
# the module's licence header. make generate writes the committed file;
# make generate-check writes a temporary one and compares.
set -eu
out=${1:-internal/api/api.gen.go}
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
go tool oapi-codegen -config internal/api/oapi-codegen.yaml contract/openapi.yaml >"$tmp"
{
	printf '// SPDX-License-Identifier: Apache-2.0\n// Copyright (C) 2026 SukramJ.\n\n'
	cat "$tmp"
} >"$out"
