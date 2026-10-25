// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// distFS is the built web interface (ui/ builds into dist/; make web).
//
//go:embed all:dist
var distFS embed.FS

// hashedAssetDir holds the content-hashed files Vite emits; they never
// change under one name.
const hashedAssetDir = "assets/"

// spaHandler serves the bundle: a file that exists as itself, anything else
// as index.html (the shell routes by hash, so a deep path is a reload of
// the shell). Mirrors loom internal/north/ui/spa.go:29-111 at b40b289a,
// including the 503 page while the bundle is not built.
func spaHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return spaHandlerFS(sub)
}

func spaHandlerFS(sub fs.FS) http.Handler {
	files := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" || path == "index.html" || strings.HasSuffix(path, "/") {
			serveIndex(sub, w)
			return
		}
		f, err := sub.Open(path)
		if err != nil {
			serveIndex(sub, w)
			return
		}
		st, err := f.Stat()
		_ = f.Close()
		if err != nil || st.IsDir() {
			serveIndex(sub, w)
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

// serveIndex writes index.html uncached, or the not-built page with 503.
func serveIndex(sub fs.FS, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	data, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(notBuiltHTML))
		return
	}
	_, _ = w.Write(data)
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
