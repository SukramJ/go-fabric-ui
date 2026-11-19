// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

// errTrailingData refuses a body with more than one JSON value.
var errTrailingData = errors.New("data after the JSON value")

// decodeStrict decodes data into v as every request body is read
// (contract 0.10.0): one JSON value and nothing after it but white space,
// no member the type does not name, and every member name exactly as the
// type spells it — encoding/json alone matches names case-insensitively
// (UR-P3-ctr-06). The same holds for nested objects.
func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errTrailingData
	}
	var raw any
	gen := json.NewDecoder(bytes.NewReader(data))
	gen.UseNumber()
	if err := gen.Decode(&raw); err != nil {
		return err
	}
	return exactNames(raw, reflect.TypeOf(v), "body")
}

// rawMessageType is json.RawMessage, whose content is not checked.
var rawMessageType = reflect.TypeFor[json.RawMessage]()

// exactNames reports the first object member of raw whose name t does not
// spell exactly, walking structs, maps, slices and pointers; a member
// typed json.RawMessage or an interface is not walked.
func exactNames(raw any, t reflect.Type, at string) error {
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t == rawMessageType {
		return nil
	}
	switch v := raw.(type) {
	case map[string]any:
		switch t.Kind() {
		case reflect.Struct:
			fields := jsonFields(t)
			for k, val := range v {
				ft, ok := fields[k]
				if !ok {
					return fmt.Errorf("%s: member %q is not named so in the contract", at, k)
				}
				if err := exactNames(val, ft, at+"."+k); err != nil {
					return err
				}
			}
		case reflect.Map:
			for k, val := range v {
				if err := exactNames(val, t.Elem(), at+"."+k); err != nil {
					return err
				}
			}
		default:
		}
	case []any:
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			for i, val := range v {
				if err := exactNames(val, t.Elem(), fmt.Sprintf("%s[%d]", at, i)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// jsonFields are a struct's members by their JSON name, embedded structs
// flattened.
func jsonFields(t reflect.Type) map[string]reflect.Type {
	out := map[string]reflect.Type{}
	for f := range t.Fields() {
		if !f.IsExported() && !f.Anonymous {
			continue
		}
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			ft := f.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				for k, v := range jsonFields(ft) {
					out[k] = v
				}
				continue
			}
		}
		if name == "" {
			name = f.Name
		}
		out[name] = f.Type
	}
	return out
}

// readBody reads r's body, bounded by maxBodyBytes.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	return io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
}

// allowedMethods are the methods mux serves r's path with other than
// through catchAll, plus HEAD where GET is, sorted: the Allow header of a
// 405 (contract 0.10.0, MethodNotAllowed).
func allowedMethods(mux *http.ServeMux, r *http.Request, catchAll string) []string {
	var out []string
	for _, m := range []string{http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPut} {
		probe := &http.Request{Method: m, URL: r.URL, Host: r.Host, Header: http.Header{}}
		if _, p := mux.Handler(probe); p != "" && p != catchAll {
			out = append(out, m)
		}
	}
	if slices.Contains(out, http.MethodGet) {
		out = append(out, http.MethodHead)
	}
	slices.Sort(out)
	return out
}

// noRoute answers a request no route of mux serves: 405 with Allow when
// the path is served with another method, else 404.
func noRoute(w http.ResponseWriter, r *http.Request, mux *http.ServeMux, catchAll string) {
	if sw, ok := w.(*statusWriter); ok {
		sw.pathLevel = true
	}
	if allow := allowedMethods(mux, r, catchAll); len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
		writeProblem(w, r, http.StatusMethodNotAllowed, codeMethodNotAllowed, "Method not allowed", r.Method+" "+r.URL.Path)
		return
	}
	writeProblem(w, r, http.StatusNotFound, codeNotFound, "No such route", r.Method+" "+r.URL.Path)
}
