// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// The problem codes of the contract (docs/contract.md, "Problems").
const (
	codeBadRequest     = "bad_request"
	codeForbidden      = "forbidden"
	codeConflict       = "conflict"
	codeNotFound       = "not_found"
	codeServiceUnready = "service_unready"
	codeInternal       = "internal"
	codeTooLarge       = "payload_too_large"
	codeNodeError      = "node_error"
)

// problemTypeBase is the base of every problem's type URI: the contract
// page's anchor for the code.
const problemTypeBase = "https://github.com/SukramJ/go-fabric-ui/blob/main/docs/contract.md#problem-"

// maxBodyBytes bounds a request body; a selection of a few thousand items
// fits well inside it.
const maxBodyBytes = 1 << 20

// problem is an RFC 9457 problem+json body. Mirrors loom's
// internal/north/rest/problem/problem.go:50-59 (Details) at b40b289a,
// without its host-specific members.
type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
	Code     string `json:"code"`
}

// writeProblem writes a problem with the code, title and detail; the
// instance is the request path. Mirrors loom problem.go:61-85 (Write, New).
func writeProblem(w http.ResponseWriter, r *http.Request, status int, code, title, detail string) {
	p := problem{
		Type:   problemTypeBase + code,
		Title:  title,
		Status: status,
		Detail: detail,
		Code:   code,
	}
	if r != nil {
		p.Instance = r.URL.Path
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("X-Problem-Code", code)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}

// serverError logs err and answers 500 internal with an empty detail: an
// error's text can carry store paths or socket detail, which is for the
// operator's log, not for a caller. Mirrors loom
// internal/north/rest/handlers/write.go:59-81 (writeServerError).
func (h *handler) serverError(w http.ResponseWriter, r *http.Request, title string, err error) {
	h.log.ErrorContext(r.Context(), title, slog.String("error", err.Error()),
		slog.String("method", r.Method), slog.String("path", r.URL.Path))
	writeProblem(w, r, http.StatusInternalServerError, codeInternal, title, "")
}

// unready answers 503 service_unready for a dependency the host did not
// wire.
func unready(w http.ResponseWriter, r *http.Request, title, detail string) {
	writeProblem(w, r, http.StatusServiceUnavailable, codeServiceUnready, title, detail)
}

// writeJSON answers 200 with v as the JSON body. Mirrors loom
// write.go:96-100.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON reads the body into v, refusing unknown fields and a body
// over maxBodyBytes. Mirrors loom write.go:109-114 (DecodeJSON).
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// badBody answers a body decodeJSON refused: 413 for one too large, 400
// otherwise. Mirrors loom write.go:44-49 (DecodeJSONStatus).
func badBody(w http.ResponseWriter, r *http.Request, err error) {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		writeProblem(w, r, http.StatusRequestEntityTooLarge, codeTooLarge, "Request body too large", err.Error())
		return
	}
	writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid request body", err.Error())
}
