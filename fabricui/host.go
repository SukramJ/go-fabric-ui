// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"unicode"
)

// hostAdmission is how the Host check admitted a request: byPredicate
// when Options.AllowHost did, not a name rule.
type hostAdmission struct{ byPredicate bool }

// hostAdmissionKey carries the admission in the request's context; its
// presence also says the check ran (a second wrapper does not repeat it).
type hostAdmissionKey struct{}

// admissionOf is the request's admission; ok is false before the check.
func admissionOf(r *http.Request) (hostAdmission, bool) {
	a, ok := r.Context().Value(hostAdmissionKey{}).(hostAdmission)
	return a, ok
}

// hostName is a Host header's or an origin's name: the port dropped (a
// bracketed IPv6 address included), lower case, one trailing dot dropped.
func hostName(hostport string) string {
	name := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		name = h
	} else if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
		name = name[1 : len(name)-1]
	}
	return strings.TrimSuffix(strings.ToLower(name), ".")
}

// nameAllowed reports whether name passes the name rules: a loopback
// name (localhost, *.localhost), an IP address without a zone, or an
// entry of Options.AllowedHosts. *.localhost is admitted on the
// assumption that the client resolves it to loopback (RFC 6761 §6.3); a
// client that asks a resolver for it trusts that resolver (decision 37
// (a), UR-P4-auth-05).
func (h *handler) nameAllowed(name string) bool {
	switch {
	case name == "":
		return false
	case name == "localhost", strings.HasSuffix(name, ".localhost"):
		return true
	}
	if a, err := netip.ParseAddr(name); err == nil {
		return a.Zone() == ""
	}
	return h.allowedHosts[name]
}

// admitHost decides r's Host: the name rules first, then
// Options.AllowHost, asked once; an empty Host is refused unasked.
func (h *handler) admitHost(r *http.Request) (hostAdmission, bool) {
	name := hostName(r.Host)
	if name == "" {
		return hostAdmission{}, false
	}
	if h.nameAllowed(name) {
		return hostAdmission{}, true
	}
	if h.allowHost != nil && h.allowHost(r) {
		return hostAdmission{byPredicate: true}, true
	}
	return hostAdmission{}, false
}

// hostCheck answers 403 host_not_allowed to a request whose Host the
// handler does not admit, on every route and in every mode, before any
// route runs; a request it admits carries its admission on. A request
// already checked by an outer hostCheck is passed on as it is.
func (h *handler) hostCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := admissionOf(r); ok {
			next.ServeHTTP(w, r)
			return
		}
		a, ok := h.admitHost(r)
		if !ok {
			hdr := w.Header()
			hdr.Set("Cache-Control", "no-store")
			setCommonHeaders(hdr)
			writeProblem(w, r, http.StatusForbidden, codeHostNotAllowed, "Host not allowed",
				"the handler does not serve this host name")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), hostAdmissionKey{}, a)))
	})
}

// allowedHostSet validates Options.AllowedHosts: host names only.
func allowedHostSet(in []string) (map[string]bool, error) {
	out := make(map[string]bool, len(in))
	for _, e := range in {
		switch {
		case e == "":
			return nil, errors.New("fabricui: Options.AllowedHosts has an empty entry")
		case strings.Contains(e, "://"):
			return nil, fmt.Errorf("fabricui: Options.AllowedHosts entry %q has a scheme; name the host only", e)
		case strings.ContainsAny(e, ":[]"):
			return nil, fmt.Errorf("fabricui: Options.AllowedHosts entry %q has a port; name the host only", e)
		case strings.Contains(e, "/"):
			return nil, fmt.Errorf("fabricui: Options.AllowedHosts entry %q has a path; name the host only", e)
		case strings.Contains(e, "*"):
			return nil, fmt.Errorf("fabricui: Options.AllowedHosts entry %q has a wildcard; name each host", e)
		case strings.ContainsFunc(e, unicode.IsSpace):
			return nil, fmt.Errorf("fabricui: Options.AllowedHosts entry %q has white space", e)
		}
		out[strings.TrimSuffix(strings.ToLower(e), ".")] = true
	}
	return out, nil
}
