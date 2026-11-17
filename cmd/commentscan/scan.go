package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Finding struct {
	File string
	Line int
	Text string
}

var directivePrefixes = []string{
	"//go:",
	"//line ",
	"//nolint",
	"//lint:",
	"//export ",
	"// +build ",
}

var licenseMarkers = []string{
	"Copyright",
	"SPDX-License-Identifier",
	"Licensed under",
}

func isDirective(text string) bool {
	for _, prefix := range directivePrefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func isLicenseHeader(group *ast.CommentGroup, file *ast.File) bool {
	if group.End() > file.Package {
		return false
	}
	text := group.Text()
	for _, marker := range licenseMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func Scan(filename string, src []byte) ([]Finding, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if ast.IsGenerated(file) {
		return nil, nil
	}
	var findings []Finding
	for _, group := range file.Comments {
		if isLicenseHeader(group, file) {
			continue
		}
		for _, comment := range group.List {
			if isDirective(comment.Text) {
				continue
			}
			findings = append(findings, Finding{
				File: filename,
				Line: fset.Position(comment.Slash).Line,
				Text: comment.Text,
			})
		}
	}
	return findings, nil
}

func skippedDirectory(root, path, name string) bool {
	if path == root {
		return false
	}
	return name == "testdata" || strings.HasPrefix(name, ".")
}

func Walk(root string) ([]Finding, error) {
	var findings []Finding
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if skippedDirectory(root, path, entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		found, err := Scan(path, src)
		if err != nil {
			return err
		}
		findings = append(findings, found...)
		return nil
	})
	return findings, err
}
