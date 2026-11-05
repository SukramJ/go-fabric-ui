// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"
)

// NodeSpec is what the operator gives for a new node (contract 0.6.0):
// its label. Everything else — key, port, discriminator, passcode, store
// — the host allocates.
type NodeSpec struct {
	// Label is the operator's name for the node: trimmed, 1 to
	// [MaxNodeLabel] characters, no control character.
	Label string
}

// MaxNodeLabel is the longest label POST /nodes and PUT
// /nodes/{key}/label accept, in characters.
const MaxNodeLabel = 64

// NodeRegistry is the host's registry of the nodes beside the default node
// (contract 0.6.0, [Options].Registry): the first node is the host's
// configuration, every further one is registry state in the host's store,
// created, renamed and deleted through the page.
//
// The host allocates a new node's key — a slug of the label, unique,
// matching ^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$, no page name of the node
// family and not [DefaultNodeKey] — and its port, discriminator, passcode
// and store; it starts the node in Create and stops it in Delete. Delete
// is final: the host removes the node's fabrics (its controllers lose the
// node), its store and its selection. The methods are called from the
// request goroutines, concurrently.
type NodeRegistry interface {
	// List is every node of the registry, in the selector's order; it is
	// called on every GET /nodes and on every request under
	// /nodes/{key}/…, so it answers from memory. The handler of a node
	// keeps its state (a window's credentials, its stream's subscribers)
	// while its entry's Source, Exposure and Events stay the same values.
	List() []NodeEntry
	// Create allocates, persists and starts a node and returns its entry.
	// An error wrapping [ErrNodeLimit] answers 409 node_limit; any other
	// 500.
	Create(ctx context.Context, spec NodeSpec) (NodeEntry, error)
	// Rename sets the label of the node key. An error wrapping
	// [ErrDefaultNode] answers 409 node_is_default; any other 500.
	Rename(ctx context.Context, key, label string) error
	// Delete stops the node key and removes its fabrics, store and
	// selection. An error wrapping [ErrDefaultNode] answers 409
	// node_is_default; any other 500.
	Delete(ctx context.Context, key string) error
}

var (
	// ErrNodeLimit is what a [NodeRegistry] wraps when it serves no further
	// node: POST /nodes answers 409 node_limit with the error's text.
	ErrNodeLimit = errors.New("fabricui: the host serves no further node")
	// ErrDefaultNode is what a [NodeRegistry] wraps when a node is its
	// configuration rather than registry state: 409 node_is_default. The
	// handler answers so itself for [Options].Node.
	ErrDefaultNode = errors.New("fabricui: the node is the host's configuration, not registry state")
)

// The problem codes of the node registry (contract 0.6.0).
const (
	codeNodeLimit     = "node_limit"
	codeNodeIsDefault = "node_is_default"
)

// EventNodesChanged carries a [NodesChangedEvent]: the node list changed
// through POST /nodes, PUT /nodes/{key}/label or DELETE /nodes/{key}
// (contract 0.6.0). Every node's stream carries it.
const EventNodesChanged = "nodes_changed"

// NodesChangedEvent is the payload of nodes_changed: the keys of the nodes
// created, renamed or deleted. A page reloads GET /nodes.
type NodesChangedEvent struct {
	Keys []string `json:"keys"`
}

// nodeLabelRequest is the body of POST /nodes and PUT /nodes/{key}/label.
type nodeLabelRequest struct {
	Label string `json:"label"`
}

// nodeLabel validates a label from a request body.
func nodeLabel(raw string) (string, error) {
	l := strings.TrimSpace(raw)
	switch {
	case l == "":
		return "", errors.New("label is empty")
	case utf8.RuneCountInString(l) > MaxNodeLabel:
		return "", fmt.Errorf("label is longer than %d characters", MaxNodeLabel)
	case strings.ContainsFunc(l, unicode.IsControl):
		return "", errors.New("label holds a control character")
	}
	return l, nil
}

// managedNode is a registry node's handler and the entry it was built on.
type managedNode struct {
	entry NodeEntry
	h     *handler
}

// sameValue compares two port values; one of a type that is not
// comparable is never the same, so its handler is rebuilt.
func sameValue(a, b any) (eq bool) {
	defer func() {
		if recover() != nil {
			eq = false
		}
	}()
	return a == b
}

func sameEntry(a, b *NodeEntry) bool {
	return sameValue(a.Source, b.Source) && sameValue(a.Exposure, b.Exposure) && sameValue(a.Events, b.Events)
}

// registryEntries is the registry's list as the handler serves it: each
// entry's key checked as Options.Nodes' are, one that is invalid or
// collides with an earlier key logged and left out. The handlers of keys
// the list no longer holds are dropped.
func (h *handler) registryEntries() []NodeEntry {
	if h.registry == nil {
		return nil
	}
	list := h.registry.List()
	out := make([]NodeEntry, 0, len(list))
	seen := make(map[string]bool, len(list)+len(h.keyed))
	for k := range h.keyed {
		seen[k] = true
	}
	for i := range list {
		e := list[i]
		if err := checkNodeKey(e.Key, seen); err != nil {
			h.log.Warn("fabricui.registry.entry_skipped", slog.String("error", err.Error()))
			continue
		}
		seen[e.Key] = true
		out = append(out, e)
	}
	h.regMu.Lock()
	for k := range h.regCache {
		if !seen[k] {
			delete(h.regCache, k)
		}
	}
	h.regMu.Unlock()
	return out
}

// managedHandler is the handler of the registry node e: the one built for
// the key while e's Source, Exposure and Events are the same values, a new
// sibling otherwise.
func (h *handler) managedHandler(e *NodeEntry) *handler {
	h.regMu.Lock()
	defer h.regMu.Unlock()
	ne := entryOf(e, true)
	if c, ok := h.regCache[e.Key]; ok && sameEntry(&c.entry, e) {
		c.h.config.Store(ne.config)
		return c.h
	}
	n := h.sibling(ne)
	h.regCache[e.Key] = managedNode{entry: *e, h: n}
	return n
}

// nodeHandler is the handler of the node key: a configured node's, or a
// registry node's while the registry lists it.
func (h *handler) nodeHandler(key string) (*handler, bool) {
	if n, ok := h.keyed[key]; ok {
		return n, true
	}
	for _, e := range h.registryEntries() {
		if e.Key == key {
			return h.managedHandler(&e), true
		}
	}
	return nil, false
}

// listedInRegistry reports whether the registry lists key.
func (h *handler) listedInRegistry(key string) bool {
	for _, e := range h.registryEntries() {
		if e.Key == key {
			return true
		}
	}
	return false
}

// registryUnready answers a registry route while the host keeps no
// registry.
func registryUnready(w http.ResponseWriter, r *http.Request) {
	unready(w, r, "Node registry not available", "the host keeps no node registry (Options.Registry is nil)")
}

// registryTarget resolves the {node} of a registry route: false when it
// answered already — 503 without a registry, 409 node_is_default for a
// configured node, 404 for a key the registry does not list.
func (h *handler) registryTarget(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.registry == nil {
		registryUnready(w, r)
		return "", false
	}
	key := r.PathValue("node")
	if _, ok := h.keyed[key]; ok {
		writeProblem(w, r, http.StatusConflict, codeNodeIsDefault, "Node is configuration",
			"the node "+key+" is the host's configuration; only a node created through POST /nodes can be renamed or deleted")
		return "", false
	}
	if !h.listedInRegistry(key) {
		writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such node", "GET /nodes lists no node "+key)
		return "", false
	}
	return key, true
}

// registryError answers an error of the host's registry.
func (h *handler) registryError(w http.ResponseWriter, r *http.Request, title string, err error) {
	switch {
	case errors.Is(err, ErrNodeLimit):
		writeProblem(w, r, http.StatusConflict, codeNodeLimit, "Node limit reached", err.Error())
	case errors.Is(err, ErrDefaultNode):
		writeProblem(w, r, http.StatusConflict, codeNodeIsDefault, "Node is configuration", err.Error())
	default:
		h.serverError(w, r, title, err)
	}
}

// readNodeLabel decodes and validates the body of a registry route; false
// when it answered already.
func readNodeLabel(w http.ResponseWriter, r *http.Request) (string, bool) {
	var body nodeLabelRequest
	if err := decodeJSON(w, r, &body); err != nil {
		badBody(w, r, err)
		return "", false
	}
	label, err := nodeLabel(body.Label)
	if err != nil {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid label", err.Error())
		return "", false
	}
	return label, true
}

// createNode serves POST /nodes (contract 0.6.0): the registry creates
// and starts the node; 201 with its summary.
func (h *handler) createNode(w http.ResponseWriter, r *http.Request) {
	if h.registry == nil {
		registryUnready(w, r)
		return
	}
	label, ok := readNodeLabel(w, r)
	if !ok {
		return
	}
	e, err := h.registry.Create(r.Context(), NodeSpec{Label: label})
	if err != nil {
		h.registryError(w, r, "Failed to create the node", err)
		return
	}
	seen := make(map[string]bool, len(h.keyed))
	for k := range h.keyed {
		seen[k] = true
	}
	if err := checkNodeKey(e.Key, seen); err != nil {
		h.serverError(w, r, "The host created a node with an invalid key", err)
		return
	}
	h.log.InfoContext(r.Context(), "fabricui.registry.created", slog.String("key", e.Key))
	h.publishNodesChanged(e.Key)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(summaryOf(entryOf(&e, true)))
}

// renameNode serves PUT /nodes/{node}/label (contract 0.6.0).
func (h *handler) renameNode(w http.ResponseWriter, r *http.Request) {
	key, ok := h.registryTarget(w, r)
	if !ok {
		return
	}
	label, ok := readNodeLabel(w, r)
	if !ok {
		return
	}
	if err := h.registry.Rename(r.Context(), key, label); err != nil {
		h.registryError(w, r, "Failed to rename the node", err)
		return
	}
	h.publishNodesChanged(key)
	w.WriteHeader(http.StatusNoContent)
}

// deleteNode serves DELETE /nodes/{node} (contract 0.6.0): the registry
// stops the node and removes its fabrics, store and selection.
func (h *handler) deleteNode(w http.ResponseWriter, r *http.Request) {
	key, ok := h.registryTarget(w, r)
	if !ok {
		return
	}
	if err := h.registry.Delete(r.Context(), key); err != nil {
		h.registryError(w, r, "Failed to delete the node", err)
		return
	}
	h.regMu.Lock()
	delete(h.regCache, key)
	h.regMu.Unlock()
	h.log.InfoContext(r.Context(), "fabricui.registry.deleted", slog.String("key", key))
	h.publishNodesChanged(key)
	w.WriteHeader(http.StatusNoContent)
}

// publishNodesChanged hands nodes_changed to every node's stream.
func (h *handler) publishNodesChanged(keys ...string) {
	if h.nodesHub != nil {
		h.nodesHub.Publish(Event{Type: EventNodesChanged, Data: NodesChangedEvent{Keys: keys}})
	}
}
