package examplecases

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	formulaFence = "```formula"
	closingFence = "```"
	markerPrefix = "case: "
)

var designPages = []string{"formula-language", "batch", "posting"}

type Example struct {
	Name string
	Line int
}

func ScanPage(page string, text string) ([]Example, []string) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var examples []Example
	var failures []string
	seen := map[string]bool{}
	for i := 0; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") != formulaFence {
			continue
		}
		opening := i + 1
		closing := i + 1
		for closing < len(lines) && strings.TrimRight(lines[closing], " \t") != closingFence {
			closing++
		}
		i = closing
		if closing+1 >= len(lines) || !strings.HasPrefix(lines[closing+1], markerPrefix) {
			failures = append(failures, fmt.Sprintf("%s:%d: example has no case line", page, opening))
			continue
		}
		name := strings.TrimSpace(strings.TrimPrefix(lines[closing+1], markerPrefix))
		if seen[name] {
			failures = append(failures, fmt.Sprintf("%s:%d: marker %q is used twice", page, closing+2, name))
		}
		seen[name] = true
		examples = append(examples, Example{Name: name, Line: opening})
	}
	return examples, failures
}

func CheckPages(root string, caseDir string, cases []Case) []string {
	pages := map[string]bool{}
	for _, name := range designPages {
		pages[pageFolder+name+".md"] = true
	}
	byPage := map[string][]Case{}
	for _, c := range cases {
		pages[c.Page] = true
		byPage[c.Page] = append(byPage[c.Page], c)
	}
	var ordered []string
	for page := range pages {
		ordered = append(ordered, page)
	}
	sort.Strings(ordered)
	var failures []string
	for _, page := range ordered {
		failures = append(failures, checkPage(root, caseDir, page, byPage[page])...)
	}
	return failures
}

func checkPage(root string, caseDir string, page string, rows []Case) []string {
	text, err := os.ReadFile(filepath.Join(root, page))
	if err != nil {
		var failures []string
		for _, row := range rows {
			failures = append(failures, fmt.Sprintf("%s/%s:%d: row %q names page %q which does not exist", caseDir, row.File, row.Line, row.Name, page))
		}
		return failures
	}
	examples, failures := ScanPage(page, string(text))
	caseFile := caseDir + "/" + strings.TrimSuffix(filepath.Base(page), ".md") + caseFileSuffix
	rowNames := map[string]bool{}
	for _, row := range rows {
		rowNames[row.Name] = true
	}
	marked := map[string]bool{}
	for _, example := range examples {
		marked[example.Name] = true
		if !rowNames[example.Name] {
			failures = append(failures, fmt.Sprintf("%s:%d: example %q has no row in %s", page, example.Line, example.Name, caseFile))
		}
	}
	for _, row := range rows {
		if !marked[row.Name] {
			failures = append(failures, fmt.Sprintf("%s/%s:%d: row %q names page %q which has no example marked \"case: %s\"", caseDir, row.File, row.Line, row.Name, page, row.Name))
		}
	}
	return failures
}
