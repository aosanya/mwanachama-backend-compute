package concurrency

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var runPathPackages = []string{"formula", "batch", "function", "number", "numtext", "datefn", "limit", "problem", "library"}

func packageLevelVariables(file *ast.File) map[string]bool {
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		general, isGeneral := declaration.(*ast.GenDecl)
		if !isGeneral || general.Tok != token.VAR {
			continue
		}
		for _, spec := range general.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if name.Name != "_" {
					names[name.Name] = true
				}
			}
		}
	}
	return names
}

func rootIdentifier(expression ast.Expr) string {
	for {
		switch e := expression.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.IndexExpr:
			expression = e.X
		case *ast.SelectorExpr:
			expression = e.X
		case *ast.StarExpr:
			expression = e.X
		case *ast.ParenExpr:
			expression = e.X
		case *ast.SliceExpr:
			expression = e.X
		default:
			return ""
		}
	}
}

func isWrittenIn(body *ast.BlockStmt, names map[string]bool) []string {
	var written []string
	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				return true
			}
			for _, left := range n.Lhs {
				if root := rootIdentifier(left); names[root] {
					written = append(written, root)
				}
			}
		case *ast.IncDecStmt:
			if root := rootIdentifier(n.X); names[root] {
				written = append(written, root)
			}
		case *ast.CallExpr:
			if function, isIdent := n.Fun.(*ast.Ident); isIdent && function.Name == "delete" && len(n.Args) > 0 {
				if root := rootIdentifier(n.Args[0]); names[root] {
					written = append(written, root)
				}
			}
		}
		return true
	})
	return written
}

func TestCS13NoPackageOnTheRunPathWritesAPackageLevelVariableAfterStartup(t *testing.T) {
	declared := 0
	for _, directory := range runPathPackages {
		entries, err := os.ReadDir(filepath.Join("..", directory))
		if err != nil {
			t.Fatalf("read %s: %v", directory, err)
		}
		files := token.NewFileSet()
		var parsed []*ast.File
		names := map[string]bool{}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			file, err := parser.ParseFile(files, filepath.Join("..", directory, entry.Name()), nil, 0)
			if err != nil {
				t.Fatalf("parse %s/%s: %v", directory, entry.Name(), err)
			}
			parsed = append(parsed, file)
			for name := range packageLevelVariables(file) {
				names[name] = true
			}
		}
		declared += len(names)
		for _, file := range parsed {
			for _, declaration := range file.Decls {
				function, isFunction := declaration.(*ast.FuncDecl)
				if !isFunction || function.Body == nil || function.Name.Name == "init" {
					continue
				}
				if written := isWrittenIn(function.Body, names); len(written) > 0 {
					sort.Strings(written)
					t.Errorf("%s: %s writes the package-level variable %s", directory, function.Name.Name, strings.Join(written, ", "))
				}
			}
		}
	}
	if declared == 0 {
		t.Fatal("no package-level variable was found, so the check read nothing")
	}
}

const writingSource = `package sample

var counter int
var table = map[string]int{}
var readOnly = []string{"a"}

func init() { counter = 1 }

func bump() { counter++ }

func store(key string) { table[key] = 2 }

func remove(key string) { delete(table, key) }

func read() string { return readOnly[0] }
`

func TestTheStateCheckSeesAWriteAndLeavesAReadAlone(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "sample.go", writingSource, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := packageLevelVariables(file)
	written := map[string]bool{}
	for _, declaration := range file.Decls {
		function, isFunction := declaration.(*ast.FuncDecl)
		if !isFunction || function.Name.Name == "init" {
			continue
		}
		written[function.Name.Name] = len(isWrittenIn(function.Body, names)) > 0
	}
	for name, want := range map[string]bool{"bump": true, "store": true, "remove": true, "read": false} {
		if written[name] != want {
			t.Errorf("%s: writes = %v, want %v", name, written[name], want)
		}
	}
}
