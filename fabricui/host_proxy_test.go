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
		"A proxy that replaces `Host` — with a name of its own or with the upstream's address — makes the check constant",
		"Behind such a proxy the check is the proxy's: it serves only its own names, with no catch-all server, and refuses any other `Host`",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("docs/embedding.md no longer says: %s", want)
		}
	}
}
