// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package fabricui serves the web interface of a go-fabric node: the HTTP
// API of the node page family on the node facade ([NodeSource], which
// *node.Node satisfies), the host's selection through an
// [ExposureProvider], server-sent events, and the built Svelte bundle
// embedded from dist/.
//
// A daemon without a web app of its own serves everything from one
// handler:
//
//	mux.Handle("/", fabricui.Handler(fabricui.Options{Node: n, Exposure: p}))
//
// A daemon with its own single-page app mounts only the API and embeds the
// pages from the npm package @sukramj/fabric-ui:
//
//	err := fabricui.Mount(mux, fabricui.Options{Node: n, Admin: isAdmin})
//
// The routes and their bodies are contract/openapi.yaml; docs/contract.md
// is the route table, docs/embedding.md the wiring and docs/provider.md how
// to write a provider.
package fabricui
