// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

// controllerHandler is a handler over a fresh in-memory controller, in
// token mode with the fixtures' token.
func controllerHandler(t *testing.T, log *bytes.Buffer) (*handler, *http.ServeMux, *MemoryController) {
	t.Helper()
	m := NewMemoryController()
	h, mux := newTestHandler(t, testOptions{noNode: true, controller: m, log: log})
	return h, mux, m
}

// Every controller fixture, each against a fresh in-memory controller.
func TestControllerFixtures(t *testing.T) {
	for _, name := range []string{
		"controller-status.200",
		"controller-nodes.200",
		"controller-node.200",
		"controller-node.200-icd",
		"controller-node.400",
		"controller-node.404",
		"controller-node-decommission.204",
		"controller-node-decommission.403",
		"controller-node-interview.204",
		"controller-node-enable.204",
		"controller-node-disable.204",
		"controller-node-auto-subscribe.204",
		"controller-node-window.200",
		"controller-node-window.400",
		"controller-node-window-close.204",
		"controller-node-attributes.200",
		"controller-node-attributes.400",
		"controller-node-attribute-write.200",
		"controller-node-attribute-write.403",
		"controller-node-command.200",
		"controller-node-acl.200",
		"controller-node-acl-set.200",
		"controller-node-acl-set.400",
		"controller-node-bindings.200",
		"controller-node-bindings-set.200",
		"controller-node-icd.200",
		"controller-node-icd-register.200",
		"controller-node-icd-register.409",
		"controller-node-icd-unregister.200",
		"controller-node-fabrics.200",
		"controller-node-fabric-remove.204",
		"controller-node-fabric-remove.400",
		"controller-commission.200",
		"controller-commission.200-discovered",
		"controller-commission.400",
		"controller-commission.404",
		"controller-commission.502",
		"controller-discovery.200",
		"controller-discovery.400",
		"controller-credentials.200",
		"controller-credentials-set.204",
		"controller-credentials-set.400",
		"controller-credentials-remove.204",
		"controller-credentials-remove.404",
		"controller-fabric-label.204",
		"controller-fabric-label.400",
		"controller-updates.200",
		"controller-updates-consent.204",
		"controller-updates-check.200",
		"controller-unknown.404",
	} {
		t.Run(name, func(t *testing.T) {
			_, mux, _ := controllerHandler(t, nil)
			runFixture(t, mux, name)
		})
	}
}

// Without Options.Controller every controller path is the API's 404, the
// signal the shell hides the family by.
func TestController_AbsentAnswers404(t *testing.T) {
	_, mux := newTestHandler(t, testOptions{noNode: true})
	for _, p := range []string{"/controller/status", "/controller/nodes", "/controller/stream"} {
		rec := do(t, mux, http.MethodGet, p, "")
		if rec.Code != http.StatusNotFound || problemOf(t, rec).Code != codeNotFound {
			t.Fatalf("%s: %d %s", p, rec.Code, rec.Body)
		}
	}
}

// adminRoutesOfContract are the controller operations contract/openapi.yaml
// marks x-admin, as "METHOD /path": a line scan of the contract's layout
// (a path at two spaces, a method at four, x-admin at six).
func adminRoutesOfContract(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pathRe := regexp.MustCompile(`^ {2}(/\S+):\s*$`)
	methodRe := regexp.MustCompile(`^ {4}(get|put|post|delete|patch):\s*$`)
	var out []string
	var path, method string
	for line := range strings.SplitSeq(string(b), "\n") {
		if m := pathRe.FindStringSubmatch(line); m != nil {
			path, method = m[1], ""
			continue
		}
		if m := methodRe.FindStringSubmatch(line); m != nil {
			method = strings.ToUpper(m[1])
			continue
		}
		if strings.TrimSpace(line) == "x-admin: true" && method != "" && strings.HasPrefix(path, "/controller/") {
			out = append(out, method+" "+path)
		}
	}
	slices.Sort(out)
	return out
}

func TestControllerAdminMatchesTheContract(t *testing.T) {
	want := adminRoutesOfContract(t)
	if len(want) < 15 {
		t.Fatalf("found %d admin operations in the contract; the scan is not reading it", len(want))
	}
	got := slices.Sorted(func(yield func(string) bool) {
		for k := range controllerAdmin {
			if !yield(k) {
				return
			}
		}
	})
	if !slices.Equal(got, want) {
		t.Fatalf("controllerAdmin differs from the contract's x-admin\n got: %v\nwant: %v", got, want)
	}
}

// Every admin route refuses a request without the token before it binds
// or decodes anything; with the host's Admin it refuses what the host
// does not admit.
func TestController_AdminRoutesRefuseWithoutRights(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	_, hostMux := newTestHandler(t, testOptions{noNode: true, controller: NewMemoryController(), admin: func(*http.Request) bool { return false }})
	for route := range controllerAdmin {
		method, path, _ := strings.Cut(route, " ")
		path = strings.NewReplacer("{id}", "0000000000000011", "{index}", "2").Replace(path)
		for _, m := range []http.Handler{mux, hostMux} {
			req := httptest.NewRequestWithContext(context.Background(), method, DefaultPrefix+path, strings.NewReader("{not json"))
			rec := httptest.NewRecorder()
			m.ServeHTTP(rec, req)
			if rec.Code != http.StatusForbidden || problemOf(t, rec).Code != codeForbidden {
				t.Fatalf("%s: %d %s", route, rec.Code, rec.Body)
			}
		}
	}
}

func TestController_HexIDs(t *testing.T) {
	for _, c := range []struct {
		in   string
		want uint64
		ok   bool
	}{
		{"0000000000000011", 0x11, true},
		{"11", 0x11, true},
		{"0x11", 0x11, true},
		{"ffffffffffffffff", 0xFFFFFFFFFFFFFFFF, true},
		{"10000000000000000", 0, false},
		{"", 0, false},
		{"0x", 0, false},
		{"-1", 0, false},
		{"12g", 0, false},
	} {
		got, err := parseHex(c.in, 64)
		if (err == nil) != c.ok || got != c.want {
			t.Errorf("parseHex(%q) = %#x, %v; want %#x, ok %v", c.in, got, err, c.want, c.ok)
		}
	}
	if hex16(0x1B669) != "000000000001B669" || hexID(6) != "0x0006" || hexID(0xFFF1FC01) != "0xFFF1FC01" {
		t.Fatal("hex formatting")
	}
}

// The window's passcode and a credential's secret reach the admin answer
// only: no log line, no event, no read of the credentials.
func TestController_SecretsStayOutOfLogsAndEvents(t *testing.T) {
	var log bytes.Buffer
	h, mux, _ := controllerHandler(t, &log)
	rec := record(h)
	h.subscribeController(rec.add)
	if r := do(t, mux, http.MethodPost, "/controller/nodes/0000000000000011/window", `{}`); r.Code != http.StatusOK || !strings.Contains(r.Body.String(), "20202021") {
		t.Fatalf("window %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodPut, "/controller/credentials", `{"kind":"wifi","id":"x","ssid":"S","passphrase":"pa55-word"}`); r.Code != http.StatusNoContent {
		t.Fatalf("credential %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"discovered":{"instance_name":"1A2B3C4D5E6F7788"},"passcode":11223344}`); r.Code != http.StatusBadGateway {
		t.Fatalf("commission %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332"}`); r.Code != http.StatusOK {
		t.Fatalf("commission %d %s", r.Code, r.Body)
	}
	creds := do(t, mux, http.MethodGet, "/controller/credentials", "")
	for _, secret := range []string{"20202021", "pa55-word", "11223344", "34970112332", "secret-passphrase"} {
		if strings.Contains(log.String(), secret) || strings.Contains(creds.Body.String(), secret) {
			t.Fatalf("%q in the log or the credential list:\n%s\n%s", secret, log.String(), creds.Body)
		}
		for _, ev := range rec.events {
			if strings.Contains(string(mustJSON(ev.Data)), secret) {
				t.Fatalf("%q in event %s", secret, ev.Type)
			}
		}
	}
	if got := rec.types(); !slices.Contains(got, EventCommissioningProgress) || !slices.Contains(got, EventNodeAdded) {
		t.Fatalf("events %v", got)
	}
}

// The controller stream carries the source's events and the handler's
// own, each with its payload, and heartbeats.
func TestControllerStream_DeliversEvents(t *testing.T) {
	h, _, m := controllerHandler(t, nil)
	h.heartbeat = 20 * time.Millisecond
	mux := http.NewServeMux()
	h.register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+DefaultPrefix+"/controller/stream", http.NoBody)
	resp, err := srv.Client().Do(req) //nolint:bodyclose // closed below; the reader goroutine drains it
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	lines := make(chan string, 64)
	go func() {
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	wait := func(want string) {
		t.Helper()
		deadline := time.After(5 * time.Second)
		for {
			select {
			case l, ok := <-lines:
				if !ok {
					t.Fatalf("stream closed waiting for %q", want)
				}
				if l == want {
					return
				}
			case <-deadline:
				t.Fatalf("no %q", want)
			}
		}
	}
	wait(": connected")
	n, _ := m.Node(0x11)
	_ = n.Disable(context.Background())
	wait("event: connection")
	wait(`data: {"node_id_hex":"0000000000000011","connection":"disconnected"}`)
	_ = m.Decommission(context.Background(), 0x12)
	wait("event: node_removed")
	wait(`data: {"node_id_hex":"0000000000000012"}`)
	wait(": heartbeat")
}

// A port error the handler has no kind for is 500 without detail; a
// node's failure is 502 with the node's text.
func TestControllerError_Kinds(t *testing.T) {
	var log bytes.Buffer
	h, _, _ := controllerHandler(t, &log)
	for _, c := range []struct {
		err    error
		status int
		code   string
	}{
		{ErrNodeNotFound, 404, codeNotFound},
		{ErrNoDevice, 404, codeNotFound},
		{ErrUnknownCredential, 404, codeNotFound},
		{ErrInvalidArgument, 400, codeBadRequest},
		{ErrInvalidCredential, 400, codeBadRequest},
		{ErrConflict, 409, codeConflict},
		{ErrNodeFailed, 502, codeNodeError},
		{context.DeadlineExceeded, 502, codeNodeError},
		{os.ErrPermission, 500, codeInternal},
	} {
		rec := httptest.NewRecorder()
		h.controllerError(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x", nil), c.err)
		p := problemOf(t, rec)
		if rec.Code != c.status || p.Code != c.code {
			t.Errorf("%v: %d %s", c.err, rec.Code, p.Code)
		}
		if c.status == 500 && p.Detail != "" {
			t.Errorf("500 with detail %q", p.Detail)
		}
	}
}

// A body over the limit is 413, as on the node routes.
func TestController_BodyTooLarge(t *testing.T) {
	_, mux, _ := controllerHandler(t, nil)
	big := `{"label":"` + strings.Repeat("x", maxBodyBytes) + `"}`
	rec := do(t, mux, http.MethodPut, "/controller/fabric/label", big)
	if rec.Code != http.StatusRequestEntityTooLarge || problemOf(t, rec).Code != codeTooLarge {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

// The in-memory controller's flows the fixtures do not show: a second
// commissioning finds no device, the ACL write drops a subject naming the
// node itself, the stream sees the write.
func TestMemoryController_Flows(t *testing.T) {
	_, mux, m := controllerHandler(t, nil)
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"code":"MT:Y.K9042C00KA0648G00"}`); r.Code != http.StatusOK {
		t.Fatalf("first %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodPost, "/controller/commission", `{"code":"34970112332"}`); r.Code != http.StatusNotFound {
		t.Fatalf("second %d %s", r.Code, r.Body)
	}
	if r := do(t, mux, http.MethodGet, "/controller/discovery", ""); r.Code != http.StatusOK || !strings.Contains(r.Body.String(), `"devices":[]`) {
		t.Fatalf("discovery after commissioning %d %s", r.Code, r.Body)
	}
	body := `{"entries":[{"privilege":"view","auth_mode":"case","subjects_hex":["0000000000000011"],"targets":[]},` +
		`{"privilege":"administer","auth_mode":"case","subjects_hex":["0000000000000011","0000000000000020"],"targets":[]}]}`
	if r := do(t, mux, http.MethodPut, "/controller/nodes/0000000000000011/acl", body); r.Code != http.StatusOK {
		t.Fatalf("acl %d %s", r.Code, r.Body)
	}
	r := do(t, mux, http.MethodGet, "/controller/nodes/0000000000000011/acl", "")
	if !strings.Contains(r.Body.String(), `"subjects_hex":["0000000000000020"]`) || strings.Contains(r.Body.String(), `"privilege":"view"`) {
		t.Fatalf("acl after write %s", r.Body)
	}
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/controller/nodes/0000000000000011/bindings", `{"endpoint":1,"bindings":[{"group":1,"endpoint":1}]}`},
		{http.MethodPut, "/controller/nodes/0000000000000011/bindings", `{"endpoint":1,"bindings":[{"node_id_hex":"12"}]}`},
		{http.MethodPut, "/controller/nodes/0000000000000011/attributes", `{"endpoint":70000,"cluster_hex":"6","attribute_hex":"0","value":true}`},
		{http.MethodPost, "/controller/nodes/0000000000000011/commands", `{"endpoint":1,"cluster_hex":"zz","command_hex":"0"}`},
		{http.MethodPost, "/controller/nodes/0000000000000011/icd/register", `{}`},
		{http.MethodGet, "/controller/discovery?timeout=100", ""},
		{http.MethodPut, "/controller/credentials", `{"kind":"thread","id":"t","dataset_hex":"zz"}`},
		{http.MethodPost, "/controller/commission", `{}`},
		{http.MethodPost, "/controller/commission", `{"code":"34970112332","discovered":{"instance_name":"x"},"passcode":1}`},
		{http.MethodPost, "/controller/updates/consent", `{"node_id_hex":"AA","vendor_id":1,"product_id":1,"target_software_version":1}`},
	} {
		if r := do(t, mux, c.method, c.path, c.body); r.Code < 400 {
			t.Errorf("%s %s %s: %d %s", c.method, c.path, c.body, r.Code, r.Body)
		}
	}
	if n, ok := m.Node(0x11); !ok || n.Disabled() {
		t.Fatal("node state")
	}
}
