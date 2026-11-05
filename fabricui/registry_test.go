// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

// portSource is a NodeEntry.Source of the registry tests: the fake node,
// which the handler takes as its nodePort (entryOf); the NodeSource
// methods are never called.
type portSource struct {
	stubSource
	*fakeNode
}

// memoryRegistry is the NodeRegistry double of the tests and the e2e
// server: a node per Create on a fake node of its own with a memory
// selection, the key a slug of the label, at most limit nodes (0 is no
// limit).
type memoryRegistry struct {
	mu      sync.Mutex
	limit   int
	nodes   []NodeEntry
	nextErr error
	// reassembled counts each node's rebuilds.
	reassembled map[string]int
}

func newMemoryRegistry(limit int) *memoryRegistry {
	return &memoryRegistry{limit: limit, reassembled: map[string]int{}}
}

func (m *memoryRegistry) List() []NodeEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.nodes)
}

// slug is the label lower-cased, every run of other characters one "-".
func slug(label string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(label) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	s := strings.TrimSuffix(b.String(), "-")
	if s == "" {
		s = "node"
	}
	return s
}

func (m *memoryRegistry) taken(key string) bool {
	if key == DefaultNodeKey || slices.Contains(reservedNodeKeys, key) {
		return true
	}
	return slices.ContainsFunc(m.nodes, func(e NodeEntry) bool { return e.Key == key })
}

func (m *memoryRegistry) Create(_ context.Context, spec NodeSpec) (NodeEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.nextErr; err != nil {
		m.nextErr = nil
		return NodeEntry{}, err
	}
	if m.limit > 0 && len(m.nodes) >= m.limit {
		return NodeEntry{}, fmt.Errorf("%w: %d nodes beside the default", ErrNodeLimit, m.limit)
	}
	key := slug(spec.Label)
	for i := 2; m.taken(key); i++ {
		key = fmt.Sprintf("%s-%d", slug(spec.Label), i)
	}
	n := secondNode()
	port := 5541 + len(m.nodes)
	n.addr = fmt.Sprintf("[::]:%d", port)
	e := NodeEntry{
		Key: key, Label: spec.Label, Kind: "bridge",
		Config: &NodeConfigView{
			Listen: n.addr, Discriminator: uint16(3000 + port%1000), VendorID: 0xFFF1, VendorName: "go-fabric",
			ProductID: 0x8000, ProductName: "Bridge", Store: "bridge-" + key + ".db",
		},
		Source:   portSource{fakeNode: n},
		Exposure: newMemoryProvider(),
		Reassemble: func(context.Context) error {
			m.mu.Lock()
			m.reassembled[key]++
			m.mu.Unlock()
			return nil
		},
	}
	m.nodes = append(m.nodes, e)
	return e, nil
}

func (m *memoryRegistry) Rename(_ context.Context, key, label string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.nextErr; err != nil {
		m.nextErr = nil
		return err
	}
	for i := range m.nodes {
		if m.nodes[i].Key == key {
			m.nodes[i].Label = label
			return nil
		}
	}
	return fmt.Errorf("no node %s", key)
}

func (m *memoryRegistry) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := m.nextErr; err != nil {
		m.nextErr = nil
		return err
	}
	i := slices.IndexFunc(m.nodes, func(e NodeEntry) bool { return e.Key == key })
	if i < 0 {
		return fmt.Errorf("no node %s", key)
	}
	m.nodes = slices.Delete(m.nodes, i, i+1)
	return nil
}

// registryHandler is a handler with the default fake node, its memory
// selection and a registry holding one node, "kitchen".
func registryHandler(t *testing.T, limit int) (*handler, *http.ServeMux, *memoryRegistry) {
	t.Helper()
	reg := newMemoryRegistry(limit)
	if _, err := reg.Create(context.Background(), NodeSpec{Label: "Kitchen"}); err != nil {
		t.Fatal(err)
	}
	h, mux := newTestHandler(t, testOptions{exposure: newMemoryProvider(), registry: reg, nodeConfig: defaultNodeConfig()})
	return h, mux, reg
}

// defaultNodeConfig is the default node's configuration view of the
// registry tests and the e2e server.
func defaultNodeConfig() *NodeConfigView {
	return &NodeConfigView{
		SourceNote: "the --bridge-* flags", Label: "Main bridge", Listen: "[::]:5540", Discriminator: 3840,
		VendorID: 0xFFF1, VendorName: "go-fabric", ProductID: 0x8000, ProductName: "Bridge", Store: "bridge.db",
	}
}

func TestRegistry_Fixtures(t *testing.T) {
	_, mux, reg := registryHandler(t, 2)
	for _, name := range []string{
		"nodes.200-registry", "nodes-status.200-registry", "node-status.200-config",
		"nodes-create.201", "nodes-create.400", "nodes-create.403", "nodes-create.409",
		"nodes-label.204", "nodes-label.400", "nodes-label.404", "nodes-label.409",
		"nodes-delete.409", "nodes-delete.404", "nodes-delete.204",
	} {
		runFixture(t, mux, name)
	}
	if got := reg.List(); len(got) != 1 || got[0].Key != "kitchen" || got[0].Label != "Kitchen and dining" {
		t.Fatalf("the registry after the fixtures: %+v", got)
	}
	// Without a registry the writing routes answer 503.
	_, mux = newTestHandler(t, testOptions{})
	runFixture(t, mux, "nodes-create.503")
}

func TestRegistry_NewRefusesAnInconsistentConfiguration(t *testing.T) {
	reg := newMemoryRegistry(0)
	for _, c := range []struct {
		name string
		o    Options
		err  string
	}{
		{"with nodes", Options{Node: stubSource{}, Registry: reg, Nodes: []NodeEntry{{Key: "a"}}}, "both set"},
		{"without the default", Options{Registry: reg}, "needs Options.Node"},
	} {
		if _, err := newHandler(c.o); err == nil || !strings.Contains(err.Error(), c.err) {
			t.Errorf("%s: %v", c.name, err)
		}
	}
	if err := Mount(http.NewServeMux(), Options{Registry: reg}); err == nil {
		t.Fatal("Mount accepted a registry without Options.Node")
	}
}

// GET /nodes lists the default and then the registry on every request;
// the registry's node serves its routes, its own selection and its own
// rebuild under /nodes/{key}/…, and its handler keeps its state between
// requests.
func TestRegistry_ListAndServe(t *testing.T) {
	h, mux, reg := registryHandler(t, 0)
	ctx := context.Background()
	if _, err := reg.Create(ctx, NodeSpec{Label: "Garden"}); err != nil {
		t.Fatal(err)
	}
	rec := do(t, mux, http.MethodGet, "/nodes", "")
	want := `{"nodes":[` +
		`{"key":"default","label":"default","enabled":true,"managed":false,"exposure":true},` +
		`{"key":"kitchen","label":"Kitchen","kind":"bridge","enabled":true,"managed":true,"exposure":true},` +
		`{"key":"garden","label":"Garden","kind":"bridge","enabled":true,"managed":true,"exposure":true}]}`
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Fatalf("GET /nodes\n got %s\nwant %s", got, want)
	}
	var st statusResponse
	rec = do(t, mux, http.MethodGet, "/nodes/garden/status", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st.ListenAddr != "[::]:5542" || !st.Exposure || !st.NodeRegistry {
		t.Fatalf("garden's status: %+v", st)
	}
	if rec := do(t, mux, http.MethodGet, "/nodes/garden/exposure", ""); rec.Code != http.StatusOK {
		t.Fatalf("garden's selection: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, mux, http.MethodPut, "/nodes/garden/exposure", `{"items":[{"key":"lamp-1","enabled":true}]}`); rec.Code != http.StatusOK {
		t.Fatalf("garden's selection write: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, mux, http.MethodPost, "/nodes/kitchen/maintenance/reassemble", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("kitchen's rebuild: %d %s", rec.Code, rec.Body)
	}
	reg.mu.Lock()
	rebuilt := countsOf(reg.reassembled)
	reg.mu.Unlock()
	if rebuilt != "garden=1 kitchen=1" {
		t.Fatalf("rebuilds %s", rebuilt)
	}
	// One handler per node while its entry stays the same.
	a, _ := h.nodeHandler("garden")
	b, _ := h.nodeHandler("garden")
	if a != b || a == h {
		t.Fatal("the registry node's handler is rebuilt between requests")
	}
	// A new Source is a new handler.
	reg.mu.Lock()
	reg.nodes[1].Source = portSource{fakeNode: secondNode()}
	reg.mu.Unlock()
	if c, _ := h.nodeHandler("garden"); c == a {
		t.Fatal("the handler survived a new Source")
	}
	// A deleted node is gone from the list, its routes and the cache.
	if err := reg.Delete(ctx, "garden"); err != nil {
		t.Fatal(err)
	}
	if rec := do(t, mux, http.MethodGet, "/nodes/garden/status", ""); rec.Code != http.StatusNotFound {
		t.Fatalf("a deleted node's status: %d", rec.Code)
	}
	h.regMu.Lock()
	_, cached := h.regCache["garden"]
	h.regMu.Unlock()
	if cached {
		t.Fatal("the deleted node's handler is still cached")
	}
}

// countsOf prints a count map sorted by key.
func countsOf(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s=%d", k, m[k]))
	}
	return strings.Join(out, " ")
}

// An entry with an invalid key, a page name or the default's key is left
// out of the list and its routes.
func TestRegistry_SkipsAnInvalidEntry(t *testing.T) {
	_, mux, reg := registryHandler(t, 0)
	reg.mu.Lock()
	reg.nodes = append(reg.nodes, NodeEntry{Key: "a/b"}, NodeEntry{Key: "status"}, NodeEntry{Key: DefaultNodeKey}, NodeEntry{Key: "kitchen"})
	reg.mu.Unlock()
	var list deviceNodeList
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/nodes", "").Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, n := range list.Nodes {
		keys = append(keys, n.Key)
	}
	if !slices.Equal(keys, []string{"default", "kitchen"}) {
		t.Fatalf("keys %v", keys)
	}
}

// The host's errors: ErrNodeLimit and ErrDefaultNode are 409 with their
// codes, any other 500 without its text; a key the host returns that the
// contract refuses is 500.
func TestRegistry_HostErrors(t *testing.T) {
	_, mux, reg := registryHandler(t, 0)
	for _, c := range []struct {
		err          error
		method, path string
		body         string
		status       int
		code         string
	}{
		{fmt.Errorf("%w: full", ErrNodeLimit), http.MethodPost, "/nodes", `{"label":"x"}`, http.StatusConflict, codeNodeLimit},
		{fmt.Errorf("%w: kitchen", ErrDefaultNode), http.MethodPut, "/nodes/kitchen/label", `{"label":"x"}`, http.StatusConflict, codeNodeIsDefault},
		{fmt.Errorf("%w: kitchen", ErrDefaultNode), http.MethodDelete, "/nodes/kitchen", "", http.StatusConflict, codeNodeIsDefault},
		{errors.New("/var/lib/store: disk full"), http.MethodPost, "/nodes", `{"label":"x"}`, http.StatusInternalServerError, codeInternal},
		{errors.New("/var/lib/store: disk full"), http.MethodPut, "/nodes/kitchen/label", `{"label":"x"}`, http.StatusInternalServerError, codeInternal},
		{errors.New("/var/lib/store: disk full"), http.MethodDelete, "/nodes/kitchen", "", http.StatusInternalServerError, codeInternal},
	} {
		reg.mu.Lock()
		reg.nextErr = c.err
		reg.mu.Unlock()
		rec := do(t, mux, c.method, c.path, c.body)
		p := problemOf(t, rec)
		if rec.Code != c.status || p.Code != c.code || strings.Contains(p.Detail, "/var/lib") {
			t.Errorf("%s %s with %v: %d %+v", c.method, c.path, c.err, rec.Code, p)
		}
	}
	bad := &badKeyRegistry{memoryRegistry: newMemoryRegistry(0)}
	_, mux = newTestHandler(t, testOptions{registry: bad})
	if rec := do(t, mux, http.MethodPost, "/nodes", `{"label":"x"}`); rec.Code != http.StatusInternalServerError {
		t.Fatalf("a created node with an invalid key: %d %s", rec.Code, rec.Body)
	}
}

// badKeyRegistry creates a node under a page name.
type badKeyRegistry struct{ *memoryRegistry }

func (b *badKeyRegistry) Create(context.Context, NodeSpec) (NodeEntry, error) {
	return NodeEntry{Key: "status"}, nil
}

func TestRegistry_LabelValidation(t *testing.T) {
	for _, c := range []struct {
		in, want, err string
	}{
		{"  Living room  ", "Living room", ""},
		{"", "", "empty"},
		{"   ", "", "empty"},
		{strings.Repeat("ä", MaxNodeLabel), strings.Repeat("ä", MaxNodeLabel), ""},
		{strings.Repeat("a", MaxNodeLabel+1), "", "longer"},
		{"a\nb", "", "control"},
	} {
		got, err := nodeLabel(c.in)
		if c.err == "" && (err != nil || got != c.want) || c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)) {
			t.Errorf("%q: %q %v", c.in, got, err)
		}
	}
}

// nodes_changed reaches every node's stream — the default's, a configured
// and a registry node's — after each of the three writes.
func TestRegistry_NodesChangedOnEveryStream(t *testing.T) {
	h, mux, _ := registryHandler(t, 0)
	kitchen, _ := h.nodeHandler("kitchen")
	def, other := record(h), record(kitchen)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPost, "/nodes", `{"label":"Garden"}`},
		{http.MethodPut, "/nodes/garden/label", `{"label":"Garden shed"}`},
		{http.MethodDelete, "/nodes/garden", ""},
	} {
		if rec := do(t, mux, c.method, c.path, c.body); rec.Code >= 300 {
			t.Fatalf("%s %s: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
	for name, r := range map[string]*recorder{"default": def, "kitchen": other} {
		if got := r.types(); !slices.Equal(got, []string{EventNodesChanged, EventNodesChanged, EventNodesChanged}) {
			t.Fatalf("%s's stream: %v", name, got)
		}
		if ev := r.events[0].Data.(NodesChangedEvent); !slices.Equal(ev.Keys, []string{"garden"}) {
			t.Fatalf("%s's payload %+v", name, ev)
		}
	}
	// Without a registry no stream subscribes to the node list.
	h, _ = newTestHandler(t, testOptions{})
	if h.nodesHub != nil {
		t.Fatal("a nodes hub without a registry")
	}
}

// A configured node beside the default serves the selection, the rebuild
// and the events of its entry (0.6.0); GET /nodes reports its exposure.
func TestNodes_EntryServesItsOwnSelection(t *testing.T) {
	rebuilt := 0
	ev := NewEvents()
	h, err := newHandler(Options{
		Node: portSource{fakeNode: sampleNode()},
		Nodes: []NodeEntry{{
			Key: "second", Source: portSource{fakeNode: secondNode()}, Exposure: newMemoryProvider(),
			Reassemble: func(context.Context) error { rebuilt++; return nil }, Events: ev,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	h.token = testToken
	mux := http.NewServeMux()
	h.register(mux)
	if rec := do(t, mux, http.MethodGet, "/node/exposure", ""); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("the default without a selection: %d", rec.Code)
	}
	if rec := do(t, mux, http.MethodGet, "/nodes/second/exposure", ""); rec.Code != http.StatusOK {
		t.Fatalf("second's selection: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, mux, http.MethodPost, "/nodes/second/maintenance/reassemble", ""); rec.Code != http.StatusNoContent || rebuilt != 1 {
		t.Fatalf("second's rebuild: %d, %d rebuilds", rec.Code, rebuilt)
	}
	r := record(h.keyed["second"])
	ev.Publish(Event{Type: EventExposureChanged, Data: ExposureEvent{Keys: []string{"lamp"}}})
	if got := r.types(); !slices.Equal(got, []string{EventExposureChanged}) {
		t.Fatalf("second's stream: %v", got)
	}
	rec := do(t, mux, http.MethodGet, "/nodes", "")
	if !strings.Contains(rec.Body.String(), `{"key":"second","label":"second","enabled":true,"managed":false,"exposure":true}`) {
		t.Fatalf("GET /nodes: %s", rec.Body)
	}
	// The default out of Options.Nodes serves its entry's selection when
	// the Options-level one is nil.
	h, err = newHandler(Options{Nodes: []NodeEntry{{Key: "x", Source: stubSource{}, Exposure: newMemoryProvider()}}})
	if err != nil {
		t.Fatal(err)
	}
	if h.exposure == nil {
		t.Fatal("the first of Nodes does not serve its selection as the default")
	}
}

// The status reports node_registry with a registry, and admin in host
// mode.
func TestStatus_RegistryAndAdmin(t *testing.T) {
	_, mux, _ := registryHandler(t, 0)
	var st map[string]any
	if err := json.Unmarshal(do(t, mux, http.MethodGet, "/node/status", "").Body.Bytes(), &st); err != nil {
		t.Fatal(err)
	}
	if st["node_registry"] != true || st["admin"] != nil {
		t.Fatalf("token mode with a registry: node_registry %v, admin %v", st["node_registry"], st["admin"])
	}
	_, mux = newTestHandler(t, testOptions{admin: func(r *http.Request) bool { return r.Header.Get("X-Admin") == "1" }})
	for _, c := range []struct {
		header string
		want   bool
	}{{"1", true}, {"", false}} {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, DefaultPrefix+"/node/status", http.NoBody)
		req.Header.Set("X-Admin", c.header)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		st = nil
		if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
			t.Fatal(err)
		}
		if st["admin"] != c.want || st["node_registry"] != nil {
			t.Fatalf("host mode %q: admin %v, node_registry %v", c.header, st["admin"], st["node_registry"])
		}
	}
}

// The registry's routes are admin routes: 403 before the body is read.
func TestRegistry_AdminOnly(t *testing.T) {
	_, mux, reg := registryHandler(t, 0)
	for _, c := range []struct{ method, path string }{
		{http.MethodPost, "/nodes"}, {http.MethodPut, "/nodes/kitchen/label"}, {http.MethodDelete, "/nodes/kitchen"},
	} {
		req := httptest.NewRequestWithContext(context.Background(), c.method, DefaultPrefix+c.path, strings.NewReader(`{"label":"x"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s %s without the token: %d", c.method, c.path, rec.Code)
		}
	}
	if got := reg.List(); len(got) != 1 || got[0].Label != "Kitchen" {
		t.Fatalf("a refused request changed the registry: %+v", got)
	}
}

// The configuration block: the default node's from Options.NodeConfig as
// configuration, a registry node's as registry with its entry's label, a
// rename shown at once; never a passcode; absent without a view.
func TestStatus_ConfigBlock(t *testing.T) {
	_, mux, _ := registryHandler(t, 0)
	config := func(path string) map[string]any {
		t.Helper()
		var st map[string]any
		if err := json.Unmarshal(do(t, mux, http.MethodGet, path, "").Body.Bytes(), &st); err != nil {
			t.Fatal(err)
		}
		c, _ := st["config"].(map[string]any)
		return c
	}
	if c := config("/node/status"); c["source"] != NodeConfigSourceConfiguration || c["source_note"] != "the --bridge-* flags" || c["discriminator"] != 3840.0 {
		t.Fatalf("default: %v", c)
	}
	if c := config("/nodes/kitchen/status"); c["source"] != NodeConfigSourceRegistry || c["label"] != "Kitchen" || c["store"] != "bridge-kitchen.db" {
		t.Fatalf("kitchen: %v", c)
	}
	if rec := do(t, mux, http.MethodPut, "/nodes/kitchen/label", `{"label":"Kitchen and dining"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("rename: %d", rec.Code)
	}
	if c := config("/nodes/kitchen/status"); c["label"] != "Kitchen and dining" {
		t.Fatalf("kitchen after the rename: %v", c)
	}
	for _, path := range []string{"/node/status", "/nodes/kitchen/status"} {
		if rec := do(t, mux, http.MethodGet, path, ""); strings.Contains(rec.Body.String(), "passcode") {
			t.Fatalf("%s names a passcode: %s", path, rec.Body)
		}
	}
	_, mux = newTestHandler(t, testOptions{})
	if c := config("/node/status"); c != nil {
		t.Fatalf("without a view: %v", c)
	}
}
