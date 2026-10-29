// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

// Package arch holds the import guard: this module reaches go-fabric only
// through its facade, and depends on nothing but go-fabric and the
// standard library. A facade gap is a question for go-fabric, not a
// reason to import a package below it. Two places are wider by decision
// (brief U2, 0.1): the controller adapter fabricui/gofabric is the one
// package that knows go-fabric's controller facade, and internal/api, the
// handlers generated from the contract, uses oapi-codegen's runtime.
package arch

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	module   = "github.com/SukramJ/go-fabric-ui"
	goFabric = "github.com/SukramJ/go-fabric"
)

// facade is what production code may import from go-fabric.
var facade = map[string]bool{
	goFabric + "/node":        true,
	goFabric + "/eligibility": true,
	goFabric + "/contract":    true,
	goFabric + "/clock":       true,
}

// adapterDir is the controller adapter; adapter is what it may import
// beside the facade: the controller facade, the Thread dataset codec and
// the generated cluster definitions it encodes ACL and Binding entries
// with.
const adapterDir = "fabricui/gofabric"

var adapter = map[string]bool{
	goFabric + "/controller":                          true,
	goFabric + "/controller/client":                   true,
	goFabric + "/controller/commissioner":             true,
	goFabric + "/controller/softwareupdate":           true,
	goFabric + "/client/discovery":                    true,
	goFabric + "/thread":                              true,
	goFabric + "/cluster/spec":                        true,
	goFabric + "/cluster/spec/accesscontrol":          true,
	goFabric + "/cluster/spec/basicinformation":       true,
	goFabric + "/cluster/spec/binding":                true,
	goFabric + "/cluster/spec/icdmanagement":          true,
	goFabric + "/cluster/spec/operationalcredentials": true,
}

// generatedDir holds the generated handlers; generated is the runtime
// they import.
const generatedDir = "internal/api"

var generated = map[string]bool{"github.com/oapi-codegen/runtime": true}

// testOnly is what a test may import beside the facade: the site the
// end-to-end test runs on, the device constructors it bridges, and the
// commissioner options fabrictest's Commission edits.
var testOnly = map[string]bool{
	goFabric + "/fabrictest":              true,
	goFabric + "/device":                  true,
	goFabric + "/controller/commissioner": true,
	goFabric + "/controller":              true,
	goFabric + "/controller/attestation":  true,
}

func TestImportsStayOnTheFacade(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found at %s: %v", root, err)
	}
	files := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "ui", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		files++
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		isTest := strings.HasSuffix(path, "_test.go")
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if msg := check(filepath.ToSlash(filepath.Dir(rel)), p, isTest); msg != "" {
				t.Errorf("%s imports %s: %s", rel, p, msg)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files < 10 {
		t.Fatalf("walked only %d Go files; the guard is not looking at the module", files)
	}
}

// check returns why an import of p by a file of dir is not allowed, or "".
func check(dir, p string, isTest bool) string {
	first, _, _ := strings.Cut(p, "/")
	switch {
	case dir == adapterDir && adapter[p]:
		return ""
	case dir == generatedDir && generated[p]:
		return ""
	case !strings.Contains(first, "."):
		return "" // the standard library
	case p == module || strings.HasPrefix(p, module+"/"):
		return ""
	case facade[p]:
		return ""
	case isTest && testOnly[p]:
		return ""
	case p == goFabric || strings.HasPrefix(p, goFabric+"/"):
		return "below go-fabric's facade (node, eligibility, contract, clock); report the gap to go-fabric (wave CK)"
	default:
		return "a dependency beyond go-fabric and the standard library"
	}
}

func TestCheckClassifies(t *testing.T) {
	for _, c := range []struct {
		dir    string
		path   string
		isTest bool
		ok     bool
	}{
		{"fabricui/gofabric", goFabric + "/controller", false, true},
		{"fabricui", goFabric + "/controller", false, false},
		{"fabricui/gofabric", goFabric + "/secure/setup", false, false},
		{"internal/api", "github.com/oapi-codegen/runtime", false, true},
		{"fabricui", "github.com/oapi-codegen/runtime", false, false},
		{"", "net/http", false, true},
		{"fabricui", goFabric + "/node", false, true},
		{"fabricui", goFabric + "/bridge", false, false},
		{"fabricui", goFabric + "/secure/setup", false, false},
		{"fabricui", goFabric + "/fabrictest", false, false},
		{"fabricui", goFabric + "/fabrictest", true, true},
		{"fabricui", goFabric + "/mdns", true, false},
		{"fabricui", "github.com/go-chi/chi/v5", false, false},
		{"fabricui", module + "/fabricui", false, true},
	} {
		if got := check(c.dir, c.path, c.isTest) == ""; got != c.ok {
			t.Errorf("check(%q, %q, test=%v) allowed=%v, want %v", c.dir, c.path, c.isTest, got, c.ok)
		}
	}
}

// TestLicenseHeaderOnEveryGoFile holds the Apache-2.0 SPDX header on every
// Go file of the module.
func TestLicenseHeaderOnEveryGoFile(t *testing.T) {
	const header = "// SPDX-License-Identifier: Apache-2.0\n// Copyright (C) 2026 SukramJ.\n"
	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(string(b), header) {
			t.Errorf("%s does not open with the Apache-2.0 SPDX header", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
