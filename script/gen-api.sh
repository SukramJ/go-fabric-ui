#!/bin/sh
# Generates the controller family's router, request and response types
# from contract/openapi.yaml with the pinned oapi-codegen (the tool
# directive of tools/go.mod, a module of its own so that the generator's
# requirements stay out of the root go.mod) and writes them to $1 (default internal/api/api.gen.go) under
# the module's licence header. make generate writes the committed file;
# make generate-check writes a temporary one and compares.
set -eu
out=${1:-internal/api/api.gen.go}
tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT
go tool -modfile=tools/go.mod oapi-codegen -config internal/api/oapi-codegen.yaml contract/openapi.yaml >"$tmp"
{
	printf '// SPDX-License-Identifier: Apache-2.0\n// Copyright (C) 2026 SukramJ.\n\n'
	cat "$tmp"
} >"$out"
