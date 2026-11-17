package basicruns

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	basicPagePath    = "../documentation/4. qa/basic-runs.md"
	examplesPagePath = "../documentation/4. qa/formula-examples.md"
	qaFolder         = "../documentation/4. qa/"
)

type Kind string

const (
	KindOwn       Kind = "own"
	KindAlias     Kind = "alias"
	KindElsewhere Kind = "elsewhere"
	KindNone      Kind = "none"
	KindHeld      Kind = "held"
)

type Case struct {
	Kind    Kind
	Section string
	Part    string
	Name    string
	Formula string
	Inputs  string
	Expect  string
	Note    string
}

type Page struct {
	Cases []Case
	Gaps  map[int]string
}

type table struct {
	section string
	header  []string
	rows    [][]string
}

var (
	headingPattern   = regexp.MustCompile(`^## (?:(\d+)\. )?(.*)$`)
	separatorPattern = regexp.MustCompile(`^\|[\s:|-]+\|$`)
	gapPattern       = regexp.MustCompile(`^(\d+)\. (.*)$`)
)

func splitRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	var cells []string
	var current strings.Builder
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) && runes[i+1] == '|' {
			current.WriteRune('|')
			i++
			continue
		}
		if runes[i] == '|' {
			cells = append(cells, cleanCell(current.String()))
			current.Reset()
			continue
		}
		current.WriteRune(runes[i])
	}
	return append(cells, cleanCell(current.String()))
}

func cleanCell(cell string) string {
	cell = strings.TrimSpace(cell)
	if len(cell) >= 2 && strings.HasPrefix(cell, "`") && strings.HasSuffix(cell, "`") && strings.Count(cell, "`") == 2 {
		cell = cell[1 : len(cell)-1]
	}
	return cell
}

func tablesOf(text string) []table {
	var tables []table
	var section string
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		if match := headingPattern.FindStringSubmatch(lines[i]); match != nil {
			section = match[1]
			if section == "" {
				section = match[2]
			}
			continue
		}
		if !strings.HasPrefix(lines[i], "|") || i+1 >= len(lines) || !separatorPattern.MatchString(lines[i+1]) {
			continue
		}
		current := table{section: section, header: splitRow(lines[i])}
		i += 2
		for ; i < len(lines) && strings.HasPrefix(lines[i], "|"); i++ {
			current.rows = append(current.rows, splitRow(lines[i]))
		}
		i--
		tables = append(tables, current)
	}
	return tables
}

func columnOf(header []string, name string) int {
	for i, cell := range header {
		if strings.EqualFold(cell, name) {
			return i
		}
	}
	return -1
}

func cellAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func sectionFormula(text string, section string) string {
	marker := "## " + section + ". "
	start := strings.Index(text, marker)
	if start < 0 {
		return ""
	}
	rest := text[start:]
	fence := strings.Index(rest, "```formula\n")
	if fence < 0 {
		return ""
	}
	body := rest[fence+len("```formula\n"):]
	end := strings.Index(body, "\n```")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(body[:end])
}

func gapsOf(text string) map[int]string {
	gaps := map[int]string{}
	start := strings.Index(text, "\n## Gaps\n")
	if start < 0 {
		return gaps
	}
	rest := text[start+len("\n## Gaps\n"):]
	if end := strings.Index(rest, "\n## "); end >= 0 {
		rest = rest[:end]
	}
	for _, line := range strings.Split(rest, "\n") {
		if match := gapPattern.FindStringSubmatch(line); match != nil {
			number := 0
			for _, digit := range match[1] {
				number = number*10 + int(digit-'0')
			}
			gaps[number] = match[2]
		}
	}
	return gaps
}

func loadPage(text string) Page {
	page := Page{Gaps: gapsOf(text)}
	for _, current := range tablesOf(text) {
		caseColumn := columnOf(current.header, "Case")
		if caseColumn < 0 {
			continue
		}
		formulaColumn := columnOf(current.header, "Formula")
		inputsColumn := columnOf(current.header, "Inputs")
		expectColumn := columnOf(current.header, "Expected")
		partColumn := columnOf(current.header, "Part")
		settingColumn := columnOf(current.header, "Setting")
		ruleColumn := columnOf(current.header, "Rule")
		shared := ""
		if formulaColumn < 0 && current.section == "12" {
			shared = sectionFormula(text, "12")
		}
		for _, row := range current.rows {
			entry := Case{
				Section: current.section,
				Part:    cellAt(row, partColumn),
				Name:    cellAt(row, caseColumn),
				Formula: cellAt(row, formulaColumn),
				Inputs:  cellAt(row, inputsColumn),
				Expect:  cellAt(row, expectColumn),
				Note:    cellAt(row, ruleColumn),
			}
			if shared != "" {
				entry.Formula = shared
			}
			entry.Kind = kindOf(entry, settingColumn >= 0)
			if settingColumn >= 0 {
				entry.Note = cellAt(row, settingColumn)
			}
			page.Cases = append(page.Cases, entry)
		}
	}
	return page
}

func kindOf(entry Case, heldTable bool) Kind {
	switch {
	case heldTable:
		return KindHeld
	case entry.Name == "none":
		return KindNone
	case entry.Formula != "" && !strings.HasPrefix(entry.Formula, "see section"):
		return KindOwn
	case strings.HasPrefix(entry.Name, "basic-"):
		return KindAlias
	}
	return KindElsewhere
}

func loadExamples(text string) map[string]Case {
	examples := map[string]Case{}
	for _, current := range tablesOf(text) {
		nameColumn := columnOf(current.header, "name")
		formulaColumn := columnOf(current.header, "formula")
		inputsColumn := columnOf(current.header, "inputs")
		expectColumn := columnOf(current.header, "expected")
		if nameColumn < 0 || formulaColumn < 0 || expectColumn < 0 {
			continue
		}
		for _, row := range current.rows {
			name := cellAt(row, nameColumn)
			examples[name] = Case{
				Kind:    KindOwn,
				Name:    name,
				Formula: cellAt(row, formulaColumn),
				Inputs:  cellAt(row, inputsColumn),
				Expect:  cellAt(row, expectColumn),
			}
		}
	}
	return examples
}

func readFile(t testing.TB, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
