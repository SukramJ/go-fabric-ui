// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"
)

// distFS is the built web interface (ui/ builds into dist/; make web).
//
//go:embed all:dist
var distFS embed.FS

// hashedAssetDir holds the content-hashed files Vite emits; they never
// change under one name.
const hashedAssetDir = "assets/"

// spaHandler serves the bundle: the root as index.html, a file that exists
// as itself, anything else 404. The shell routes by hash and names its
// assets relatively ("./assets/…"), so the index under any other path
// would load its assets below that path and get no script: a blank page
// (UR-P2-web-91). Mirrors loom internal/north/ui/spa.go:29-111 at
// b40b289a, including the 503 page while the bundle is not built; the 404
// for an unknown path is this module's. frameAncestors are
// the parents admitted beside the page's own origin (Options.FrameAncestors).
func spaHandler(frameAncestors []string) http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return spaHandlerFS(sub, frameAncestors)
}

func spaHandlerFS(sub fs.FS, frameAncestors []string) http.Handler {
	files := http.FileServer(http.FS(sub))
	idx := readIndex(sub, frameAncestors)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setCommonHeaders(w.Header())
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "index.html" {
			idx.serve(w)
			return
		}
		if strings.HasSuffix(path, "/") {
			http.NotFound(w, r)
			return
		}
		f, err := sub.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		st, err := f.Stat()
		_ = f.Close()
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", cacheControlFor(path))
		files.ServeHTTP(w, r)
	})
}

// cacheControlFor caches a hashed asset for a year and anything else for an
// hour with revalidation (loom spa.go:73-85).
func cacheControlFor(path string) string {
	if strings.HasPrefix(path, hashedAssetDir) {
		return "public, max-age=31536000, immutable"
	}
	return "public, max-age=3600, must-revalidate"
}

// index is the page served for the shell's paths: index.html, or the
// not-built page with 503, with the policy derived from it.
type index struct {
	body   []byte
	status int
	csp    string
}

// readIndex reads index.html once, at start (the bundle is embedded), and
// derives its policy from the bytes it serves.
func readIndex(sub fs.FS, frameAncestors []string) index {
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		data = []byte(notBuiltHTML)
		return index{body: data, status: http.StatusServiceUnavailable, csp: contentSecurityPolicy(data, frameAncestors)}
	}
	return index{body: data, status: http.StatusOK, csp: contentSecurityPolicy(data, frameAncestors)}
}

// serve writes the page uncached with its policy.
func (i index) serve(w http.ResponseWriter) {
	hdr := w.Header()
	hdr.Set("Content-Type", "text/html; charset=utf-8")
	hdr.Set("Cache-Control", "no-store")
	hdr.Set("Content-Security-Policy", i.csp)
	w.WriteHeader(i.status)
	_, _ = w.Write(i.body)
}

// inlineScript and inlineStyle find a page's inline <script> (one without
// src) and <style> blocks; the hash covers the bytes between the tags, as
// the browser hashes them.
var (
	inlineScript = regexp.MustCompile(`(?is)<script\b([^>]*)>(.*?)</script>`)
	inlineStyle  = regexp.MustCompile(`(?is)<style\b[^>]*>(.*?)</style>`)
	srcAttr      = regexp.MustCompile(`(?i)\ssrc\s*=`)
)

// contentSecurityPolicy is the page's policy (UR-P0-web-23, -24): its own
// origin for everything, the inline pre-paint script admitted by its hash
// (derived from the page served, so it cannot drift from it), no plugins,
// no <base>, forms to the own origin, framing by the own origin and
// frameAncestors. Nothing inline is admitted for styles but a <style>
// element of the page itself, by its hash (the not-built page): the
// Svelte pages set styles through the CSSOM (style: directives,
// style.cssText), which the policy does not govern, and no template holds
// a style attribute. The camera scan needs no source: the video's
// srcObject is a MediaStream, not a URL.
func contentSecurityPolicy(page []byte, frameAncestors []string) string {
	scripts := []string{"'self'"}
	for _, m := range inlineScript.FindAllSubmatch(page, -1) {
		if !srcAttr.Match(m[1]) {
			scripts = append(scripts, hashSource(m[2]))
		}
	}
	styles := append(make([]string, 0, 2), "'self'")
	for _, m := range inlineStyle.FindAllSubmatch(page, -1) {
		styles = append(styles, hashSource(m[1]))
	}
	return strings.Join([]string{
		"default-src 'self'",
		"script-src " + strings.Join(scripts, " "),
		"style-src " + strings.Join(styles, " "),
		"img-src 'self'",
		"object-src 'none'",
		"base-uri 'none'",
		"form-action 'self'",
		"frame-ancestors " + strings.Join(append([]string{"'self'"}, frameAncestors...), " "),
	}, "; ")
}

// hashSource is the CSP hash-source of b.
func hashSource(b []byte) string {
	sum := sha256.Sum256(b)
	return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
}

// cspSource is a CSP scheme-source ("https:") or host-source
// ("https://*.example.com:8123/path"), CSP Level 3 section 2.3.1; an IPv6
// literal is none.
var cspSource = regexp.MustCompile(`^(?:[A-Za-z][A-Za-z0-9+.-]*:` +
	`|(?:[A-Za-z][A-Za-z0-9+.-]*://)?(?:\*|(?:\*\.)?[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*)(?::(?:[0-9]+|\*))?(?:/[^\s;,']*)?)$`)

// validFrameAncestors refuses an entry of Options.FrameAncestors that is
// no host-source: a keyword ('self' is always there), a separator or white
// space would change the policy's meaning, and a scheme-source ("https:")
// or a bare host wildcard ("*", "https://*:443") admits any parent, which
// is what the default prevents (UR-P0-web-23, owner decision 22).
func validFrameAncestors(list []string) error {
	for _, s := range list {
		if !cspSource.MatchString(s) {
			return fmt.Errorf("fabricui: Options.FrameAncestors: %q is not a CSP host-source", s)
		}
		if anyParent(s) {
			return fmt.Errorf("fabricui: Options.FrameAncestors: %q admits any parent; name the parent's origin", s)
		}
	}
	return nil
}

// anyParent reports whether the CSP source s, already matched by
// cspSource, is a scheme-source, a host-source whose host is "*", or one
// whose wildcard spans a single label ("*.com", "*.local"), which admits
// nearly any parent (UR-P1-web-16).
func anyParent(s string) bool {
	if _, rest, ok := strings.Cut(s, "://"); ok {
		s = rest
	} else if strings.HasSuffix(s, ":") {
		return true // a scheme-source such as "https:"
	}
	host, _, _ := strings.Cut(s, "/")
	host, _, _ = strings.Cut(host, ":")
	if rest, ok := strings.CutPrefix(host, "*."); ok {
		return !strings.Contains(rest, ".")
	}
	return host == "*"
}

const notBuiltHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>go-fabric-ui — bundle not built</title>
    <style>
      body { font: 14px/1.5 system-ui, sans-serif; max-width: 42rem; margin: 4rem auto; padding: 0 1rem; color: #1e293b; }
      code, pre { background: #f1f5f9; padding: 0.1rem 0.4rem; border-radius: 0.25rem; }
      pre { padding: 0.75rem 1rem; }
    </style>
  </head>
  <body>
    <h1>Bundle not built</h1>
    <p>The web interface has not been built into <code>fabricui/dist/</code>. Build it with:</p>
    <pre>make web-install   # once
make web</pre>
    <p>The API under <code>/api/v1/node/</code> answers regardless.</p>
  </body>
</html>
`
