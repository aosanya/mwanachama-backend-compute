package examplecases

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	caseFileSuffix = ".cases.md"
	pageFolder     = "documentation/2. design/"
	headerRow      = "| name | formula | inputs | expected | rule | page |"
	separatorRow   = "|---|---|---|---|---|---|"
	columnCount    = 6
	maxRows        = 200
	minNameLength  = 3
	maxNameLength  = 60
	maxFormula     = 500
	maxRule        = 120
	maxInputName   = 40
	maxProblemCode = 40
)

var (
	caseNamePattern  = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	inputNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	problemPattern   = regexp.MustCompile(`^[a-z0-9-]+$`)
)

type Input struct {
	Name  string
	Value Value
}

type Expected struct {
	Result   *Value
	Problems []string
}

type Case struct {
	Name     string
	Formula  string
	Inputs   []Input
	Expected Expected
	Rule     string
	Page     string
	File     string
	Line     int
}

type FileError struct {
	File   string
	Line   int
	Reason string
	Detail string
}

func (e FileError) Error() string {
	message := fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Reason)
	if e.Detail != "" {
		message += ": " + e.Detail
	}
	return message
}

func PageOf(file string) string {
	return pageFolder + strings.TrimSuffix(filepath.Base(file), caseFileSuffix) + ".md"
}

func Load(dir string) ([]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), caseFileSuffix) {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	var all []Case
	seen := map[string]Case{}
	for _, name := range names {
		text, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		cases, err := Parse(name, string(text))
		if err != nil {
			return nil, err
		}
		for _, c := range cases {
			if first, found := seen[c.Name]; found {
				return nil, FileError{File: c.File, Line: c.Line, Reason: "duplicate name", Detail: fmt.Sprintf("%q is also in %s:%d", c.Name, first.File, first.Line)}
			}
			seen[c.Name] = c
		}
		all = append(all, cases...)
	}
	return all, nil
}

func Parse(file string, text string) ([]Case, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	index, err := findTable(file, lines)
	if err != nil {
		return nil, err
	}
	var cases []Case
	names := map[string]int{}
	for ; index < len(lines); index++ {
		line := strings.TrimRight(lines[index], " \t")
		number := index + 1
		if line == "" {
			continue
		}
		if len(cases) == maxRows {
			return nil, FileError{File: file, Line: number, Reason: "too many rows", Detail: fmt.Sprintf("a file holds at most %d", maxRows)}
		}
		c, err := parseRow(file, number, line)
		if err != nil {
			return nil, err
		}
		if first, found := names[c.Name]; found {
			return nil, FileError{File: file, Line: number, Reason: "duplicate name", Detail: fmt.Sprintf("%q is also on line %d", c.Name, first)}
		}
		names[c.Name] = number
		cases = append(cases, c)
	}
	return cases, nil
}

func findTable(file string, lines []string) (int, error) {
	index := 0
	for index < len(lines) && (strings.TrimSpace(lines[index]) == "" || (index == 0 && strings.HasPrefix(lines[index], "# "))) {
		index++
	}
	if index >= len(lines) || strings.TrimRight(lines[index], " \t") != headerRow {
		return 0, FileError{File: file, Line: min(index+1, max(len(lines), 1)), Reason: "bad header", Detail: "the header row must be " + headerRow}
	}
	index++
	if index >= len(lines) || strings.TrimRight(lines[index], " \t") != separatorRow {
		return 0, FileError{File: file, Line: index + 1, Reason: "bad header", Detail: "the separator row must follow the header"}
	}
	return index + 1, nil
}

func splitCells(line string) ([]string, bool) {
	if !strings.HasPrefix(line, "|") {
		return nil, false
	}
	var cells []string
	var cell strings.Builder
	closed := false
	for i := 1; i < len(line); i++ {
		switch {
		case line[i] == '\\' && i+1 < len(line) && line[i+1] == '|':
			cell.WriteByte('|')
			i++
		case line[i] == '|':
			cells = append(cells, strings.TrimSpace(cell.String()))
			cell.Reset()
			closed = true
		default:
			cell.WriteByte(line[i])
			closed = false
		}
	}
	if !closed {
		return nil, false
	}
	return cells, true
}

func parseRow(file string, number int, line string) (Case, error) {
	fail := func(reason string, detail string) (Case, error) {
		return Case{}, FileError{File: file, Line: number, Reason: reason, Detail: detail}
	}
	cells, ok := splitCells(line)
	if !ok || len(cells) != columnCount {
		return fail("wrong column count", fmt.Sprintf("a row has %d cells and one line", columnCount))
	}
	for i, cell := range cells {
		if cell == "" && i != 2 {
			return fail("empty cell", "cell "+fmt.Sprint(i+1)+" is required")
		}
	}
	name := cells[0]
	if !caseNamePattern.MatchString(name) || len(name) < minNameLength || len(name) > maxNameLength {
		return fail("bad name", fmt.Sprintf("case name %q must be lowercase letters, digits and hyphens, %d to %d characters, starting with a letter", name, minNameLength, maxNameLength))
	}
	if utf8.RuneCountInString(cells[1]) > maxFormula {
		return fail("too long", fmt.Sprintf("the formula is over %d characters", maxFormula))
	}
	inputs, reason, detail := parseInputs(cells[2])
	if reason != "" {
		return fail(reason, detail)
	}
	expected, reason, detail := parseExpected(cells[3])
	if reason != "" {
		return fail(reason, detail)
	}
	if utf8.RuneCountInString(cells[4]) > maxRule {
		return fail("too long", fmt.Sprintf("the rule is over %d characters", maxRule))
	}
	if cells[5] != PageOf(file) {
		return fail("wrong page", fmt.Sprintf("the page of this file is %q", PageOf(file)))
	}
	return Case{Name: name, Formula: cells[1], Inputs: inputs, Expected: expected, Rule: cells[4], Page: cells[5], File: file, Line: number}, nil
}

func parseInputs(cell string) ([]Input, string, string) {
	if cell == "" {
		return nil, "", ""
	}
	parts, ok := splitOutside(cell, ", ")
	if !ok {
		return nil, "bad input", "an input must be name=type:value"
	}
	var inputs []Input
	seen := map[string]bool{}
	for _, part := range parts {
		name, typed, found := strings.Cut(part, "=")
		if !found {
			return nil, "bad input", fmt.Sprintf("%q is not name=type:value", part)
		}
		if !inputNamePattern.MatchString(name) || len(name) > maxInputName {
			return nil, "bad name", fmt.Sprintf("input name %q must be lowercase letters, digits and underscores, 1 to %d characters, starting with a letter", name, maxInputName)
		}
		if seen[name] {
			return nil, "duplicate name", fmt.Sprintf("input %q is written twice", name)
		}
		seen[name] = true
		value, reason := parseTyped(typed)
		if reason == reasonTooLong {
			return nil, "too long", fmt.Sprintf("the value of input %q is over a limit", name)
		}
		if reason != "" {
			return nil, "bad input", fmt.Sprintf("the value of input %q is not a typed value", name)
		}
		inputs = append(inputs, Input{Name: name, Value: value})
	}
	return inputs, "", ""
}

func parseExpected(cell string) (Expected, string, string) {
	switch {
	case strings.HasPrefix(cell, "result:"):
		value, reason := parseTyped(strings.TrimSpace(strings.TrimPrefix(cell, "result:")))
		if reason == reasonTooLong {
			return Expected{}, "too long", "the expected value is over a limit"
		}
		if reason != "" {
			return Expected{}, "bad expected", "the result is not a typed value"
		}
		return Expected{Result: &value}, "", ""
	case strings.HasPrefix(cell, "problems:"):
		list := strings.TrimSpace(strings.TrimPrefix(cell, "problems:"))
		if list == "" {
			return Expected{}, "bad expected", "a problems list names at least one code"
		}
		var codes []string
		for _, code := range strings.Split(list, ",") {
			code = strings.TrimSpace(code)
			if !problemPattern.MatchString(code) || len(code) > maxProblemCode {
				return Expected{}, "bad expected", fmt.Sprintf("problem code %q must be lowercase letters, digits and hyphens, 1 to %d characters", code, maxProblemCode)
			}
			codes = append(codes, code)
		}
		return Expected{Problems: codes}, "", ""
	}
	return Expected{}, "bad expected", "the cell starts with result: or problems:"
}
