// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package statusguard is the status guard of decision 47 for the tests of
// every package that serves the API: a TestMain runs the tests through
// Main, which records each status the API answers (internal/statusobs)
// and fails the run on one the contract does not declare for the
// operation. Imported by tests only.
package statusguard

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/SukramJ/go-fabric-ui/internal/statusobs"
)

// Operation is one operation of contract/openapi.yaml.
type Operation struct {
	Method, Path string
	Body         bool
	Statuses     []string
}

// Main runs m with every API status recorded and, when the tests passed
// and are not -short, requires each status to be declared for its
// operation in the contract at contractPath. prefixes are the mount
// prefixes the tests use. It returns the exit code.
func Main(m *testing.M, contractPath string, prefixes ...string) int {
	var r recorder
	statusobs.Observe = r.observe
	code := m.Run()
	statusobs.Observe = nil
	if code != 0 || testing.Short() {
		return code
	}
	b, err := os.ReadFile(contractPath) //nolint:gosec // the test's own contract path
	if err != nil {
		fmt.Fprintln(os.Stderr, "status guard:", err)
		return 1
	}
	if bad := r.undeclared(string(b), prefixes); len(bad) > 0 {
		fmt.Fprintf(os.Stderr, "status guard: %d answered status(es) the contract does not declare for the operation:\n", len(bad))
		for _, s := range bad {
			fmt.Fprintln(os.Stderr, "  - "+s)
		}
		return 1
	}
	return 0
}

type recorder struct {
	sync.Mutex
	seen map[string]bool // "METHOD /request/path STATUS"
}

func (r *recorder) observe(method, path string, status int) {
	r.Lock()
	defer r.Unlock()
	if r.seen == nil {
		r.seen = map[string]bool{}
	}
	r.seen[fmt.Sprintf("%s %s %d", method, path, status)] = true
}

// declaredOp is one contract operation with its path as a pattern.
type declaredOp struct {
	method, path string
	re           *regexp.Regexp
	literal      int
	statuses     map[string]bool
}

// undeclared are the recorded statuses no operation of doc declares.
func (r *recorder) undeclared(doc string, prefixes []string) []string {
	param := regexp.MustCompile(`\{[^}]+\}`)
	ops := ScanOperations(doc)
	declared := make([]declaredOp, 0, len(ops))
	for _, op := range ops {
		// A path parameter is one segment; the rest is literal.
		var sb strings.Builder
		last := 0
		for _, loc := range param.FindAllStringIndex(op.Path, -1) {
			sb.WriteString(regexp.QuoteMeta(op.Path[last:loc[0]]))
			sb.WriteString(`[^/]+`)
			last = loc[1]
		}
		sb.WriteString(regexp.QuoteMeta(op.Path[last:]))
		d := declaredOp{
			method: op.Method, path: op.Path, re: regexp.MustCompile("^" + sb.String() + "$"),
			literal: len(param.ReplaceAllString(op.Path, "")), statuses: map[string]bool{},
		}
		for _, s := range op.Statuses {
			d.statuses[s] = true
		}
		declared = append(declared, d)
	}
	r.Lock()
	defer r.Unlock()
	var bad []string
	for k := range r.seen {
		parts := strings.SplitN(k, " ", 3)
		method, path, status := parts[0], parts[1], parts[2]
		if method == "HEAD" {
			method = "GET"
		}
		rel, ok := relativePath(path, prefixes)
		if !ok {
			continue
		}
		var best *declaredOp
		for i := range declared {
			d := &declared[i]
			if d.method == method && d.re.MatchString(rel) && (best == nil || d.literal > best.literal) {
				best = d
			}
		}
		if best == nil {
			continue
		}
		if !best.statuses[status] {
			bad = append(bad, fmt.Sprintf("%s %s answered %s (%s)", method, best.path, status, path))
		}
	}
	slices.Sort(bad)
	return slices.Compact(bad)
}

// relativePath is path under one of the prefixes.
func relativePath(path string, prefixes []string) (string, bool) {
	for _, p := range prefixes {
		if rest, ok := strings.CutPrefix(path, p); ok && strings.HasPrefix(rest, "/") {
			return rest, true
		}
	}
	return "", false
}

// ScanOperations are the operations of the contract's paths with their
// declared statuses and whether they take a request body.
func ScanOperations(doc string) []Operation {
	pathRe := regexp.MustCompile(`^ {2}(/\S+):\s*$`)
	methodRe := regexp.MustCompile(`^ {4}(get|put|post|delete|patch):\s*$`)
	statusRe := regexp.MustCompile(`^ {8}"(\d{3})":`)
	var out []Operation
	var path string
	inResponses := false
	for line := range strings.SplitSeq(doc, "\n") {
		if line == "components:" {
			break
		}
		if m := pathRe.FindStringSubmatch(line); m != nil {
			path = m[1]
			continue
		}
		if m := methodRe.FindStringSubmatch(line); m != nil {
			out = append(out, Operation{Method: strings.ToUpper(m[1]), Path: path})
			inResponses = false
			continue
		}
		if len(out) == 0 {
			continue
		}
		op := &out[len(out)-1]
		switch {
		case line == "      requestBody:":
			op.Body = true
		case line == "      responses:":
			inResponses = true
		case strings.HasPrefix(line, "      ") && !strings.HasPrefix(line, "       "):
			inResponses = false
		case inResponses:
			if m := statusRe.FindStringSubmatch(line); m != nil {
				op.Statuses = append(op.Statuses, m[1])
			}
		}
	}
	return out
}
