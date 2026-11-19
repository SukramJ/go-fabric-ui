// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"testing"
)

// The Host check sees the Host the proxy sends. Behind a proxy that
// replaces it — with the upstream's address or a name of its own the
// host listed — a request under any name the browser used passes, the
// token included: the check is then the proxy's. Behind a proxy that
// forwards Host, the unlisted name is refused. This is the behaviour
// docs/embedding.md ("Behind an Ingress", Host names) documents, held
// here with the docs' sentence (decision 69).
func TestHostCheck_BehindAProxyThatReplacesHost(t *testing.T) {
	h, err := newHandler(Options{AllowedHosts: []string{"ui.lan"}})
	if err != nil {
		t.Fatal(err)
	}
	h.node, h.configured = sampleNode(), true
	api := http.NewServeMux()
	h.register(api)
	backend := httptest.NewServer(api)
	defer backend.Close()
	target, _ := url.Parse(backend.URL)
	for _, tc := range []struct {
		name    string
		rewrite func(pr *httputil.ProxyRequest)
		want    int
	}{
		{"the upstream's address", func(pr *httputil.ProxyRequest) { pr.SetURL(target) }, http.StatusOK},
		{"a listed name of its own", func(pr *httputil.ProxyRequest) { pr.SetURL(target); pr.Out.Host = "ui.lan" }, http.StatusOK},
		{"Host forwarded", func(pr *httputil.ProxyRequest) { pr.SetURL(target); pr.Out.Host = pr.In.Host }, http.StatusForbidden},
	} {
		proxy := httptest.NewServer(&httputil.ReverseProxy{Rewrite: tc.rewrite})
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, proxy.URL+DefaultPrefix+"/node/status", http.NoBody)
		req.Host = "rebound.attacker.example"
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(res.Body)
		res.Body.Close()
		proxy.Close()
		if res.StatusCode != tc.want {
			t.Errorf("%s: %d, want %d (token issued: %v)", tc.name, res.StatusCode, tc.want, strings.Contains(string(b), h.token))
		}
	}
	doc, err := os.ReadFile("../docs/embedding.md")
	if err != nil {
		t.Fatal(err)
	}
	flat := strings.Join(strings.Fields(string(doc)), " ")
	for _, want := range []string{
		"the host **lists** the names the browser uses in `AllowedHosts`: the handler checks the name — the one setup in which the check is the handler's",
		"every name the proxy forwards passes, and the page's own `Origin` equals its `Host`, so reads and writes pass under any name. The check is then the proxy's",
		"It replaces `Host` — with a name of its own or with the upstream's address: the check is constant",
		"There the proxy authenticates the user, or serves only its own names, with no catch-all server, and refuses any other `Host`",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("docs/embedding.md no longer says: %s", want)
		}
	}
}

// Behind a proxy that forwards Host, AllowHost answered from the proxy's
// source address admits every name the proxy forwards: a rebound name
// reads the token, and its own Origin equals its Host, so its admin
// writes pass too. The check is then the proxy's, as docs/embedding.md
// says (the second kind of proxy; decision 61, UR-P4-doc-190).
func TestHostCheck_AllowHostBySourceAdmitsEveryForwardedName(t *testing.T) {
	bySource := func(r *http.Request) bool { return strings.HasPrefix(r.RemoteAddr, "192.0.2.1:") } // httptest's RemoteAddr
	_, mux := newTestHandler(t, testOptions{allowHost: bySource})
	rec := reqWith(t, mux, http.MethodGet, DefaultPrefix+"/node/status", "", map[string]string{"Host": "rebound.attacker.example"})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), testToken) {
		t.Errorf("status under a forwarded name: %d, token issued: %v", rec.Code, strings.Contains(rec.Body.String(), testToken))
	}
	rec = reqWith(t, mux, http.MethodPost, DefaultPrefix+"/node/pairing/window/close", "", map[string]string{
		"Host": "rebound.attacker.example", "Origin": "http://rebound.attacker.example", TokenHeader: testToken,
	})
	if rec.Code != http.StatusNoContent {
		t.Errorf("an admin write under a forwarded name with its own Origin: %d %s", rec.Code, rec.Body)
	}
}
