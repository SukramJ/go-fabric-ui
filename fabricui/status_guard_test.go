// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
)

// answered collects every status the API answered during the package's
// tests, by operation (decision 47).
var answered struct {
	sync.Mutex
	seen map[string]bool // "METHOD /contract/path STATUS"
}

// TestMain runs the package's tests with observeStatus set and then
// requires every status a route answered to be declared for its
// operation in contract/openapi.yaml. A path the contract does not name,
// or a method it does not list there, is answered at path level (404,
// 405) and not checked here.
func TestMain(m *testing.M) {
	answered.seen = map[string]bool{}
	observeStatus = func(method, path string, status int) {
		answered.Lock()
		answered.seen[fmt.Sprintf("%s %s %d", method, path, status)] = true
		answered.Unlock()
	}
	code := m.Run()
	if code == 0 && !testing.Short() {
		if bad := undeclaredStatuses(); len(bad) > 0 {
			fmt.Fprintf(os.Stderr, "status guard: %d answered status(es) the contract does not declare for the operation:\n", len(bad))
			for _, b := range bad {
				fmt.Fprintln(os.Stderr, "  - "+b)
			}
			code = 1
		}
	}
	os.Exit(code)
}

// declaredOp is one contract operation with its path as a pattern.
type declaredOp struct {
	method, path string
	re           *regexp.Regexp
	literal      int
	statuses     map[string]bool
}

// undeclaredStatuses are the answered statuses no operation declares.
func undeclaredStatuses() []string {
	b, err := os.ReadFile("../contract/openapi.yaml")
	if err != nil {
		return []string{err.Error()}
	}
	ops := scanOperations(string(b))
	param := regexp.MustCompile(`\{[^}]+\}`)
	var declared []declaredOp
	for _, op := range ops {
		// A path parameter is one segment; the rest is literal.
		var sb strings.Builder
		last := 0
		for _, loc := range param.FindAllStringIndex(op.path, -1) {
			sb.WriteString(regexp.QuoteMeta(op.path[last:loc[0]]))
			sb.WriteString(`[^/]+`)
			last = loc[1]
		}
		sb.WriteString(regexp.QuoteMeta(op.path[last:]))
		d := declaredOp{
			method: op.method, path: op.path, re: regexp.MustCompile("^" + sb.String() + "$"),
			literal: len(param.ReplaceAllString(op.path, "")), statuses: map[string]bool{},
		}
		for _, s := range op.statuses {
			d.statuses[s] = true
		}
		declared = append(declared, d)
	}
	answered.Lock()
	defer answered.Unlock()
	var bad []string
	for k := range answered.seen {
		parts := strings.SplitN(k, " ", 3)
		method, path, status := parts[0], parts[1], parts[2]
		if method == "HEAD" {
			method = "GET"
		}
		rel, ok := relativePath(path)
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

// relativePath is path under the API prefixes the tests mount.
func relativePath(path string) (string, bool) {
	for _, p := range []string{DefaultPrefix, "/ui-api"} {
		if rest, ok := strings.CutPrefix(path, p); ok && strings.HasPrefix(rest, "/") {
			return rest, true
		}
	}
	return "", false
}

// scanOperations is contractOps without a *testing.T.
func scanOperations(doc string) []contractOperation {
	pathRe := regexp.MustCompile(`^ {2}(/\S+):\s*$`)
	methodRe := regexp.MustCompile(`^ {4}(get|put|post|delete|patch):\s*$`)
	statusRe := regexp.MustCompile(`^ {8}"(\d{3})":`)
	var out []contractOperation
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
			out = append(out, contractOperation{method: strings.ToUpper(m[1]), path: path})
			inResponses = false
			continue
		}
		if len(out) == 0 {
			continue
		}
		op := &out[len(out)-1]
		switch {
		case line == "      requestBody:":
			op.body = true
		case line == "      responses:":
			inResponses = true
		case strings.HasPrefix(line, "      ") && !strings.HasPrefix(line, "       "):
			inResponses = false
		case inResponses:
			if m := statusRe.FindStringSubmatch(line); m != nil {
				op.statuses = append(op.statuses, m[1])
			}
		}
	}
	return out
}
