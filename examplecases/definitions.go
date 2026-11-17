package examplecases

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	fdef "github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	TargetRow        = "row"
	TargetDefinition = "definition"

	definitionHeaderRow = "| id | part | name | detail | message |"
	definitionSeparator = "|---|---|---|---|---|"
	bindingHeaderRow    = "| case | definition | target |"
	bindingSeparator    = "|---|---|---|"
	definitionColumns   = 5
	bindingColumns      = 3

	partInput = "input"
	partCheck = "check"
	partFrom  = "from"
)

var (
	numberPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)
	countPattern  = regexp.MustCompile(`^[0-9]+$`)

	definitionTypes = map[string]formula.Type{
		"decimal":         formula.TypeDecimal,
		"integer":         formula.TypeInteger,
		"float":           formula.TypeFloat,
		"text":            formula.TypeText,
		"date":            formula.TypeDate,
		"boolean":         formula.TypeBoolean,
		"list of decimal": formula.TypeListDecimal,
		"list of integer": formula.TypeListInteger,
	}
)

type Binding struct {
	Case       string
	Definition string
	Target     string
	Line       int
}

type Definitions struct {
	Definitions map[string]fdef.Definition
	Bindings    map[string]Binding
}

type draft struct {
	id         string
	definition fdef.Definition
	copied     map[string]bool
}

func LoadDefinitions(path string) (Definitions, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return Definitions{}, err
	}
	return ParseDefinitions(path, string(text))
}

func ParseDefinitions(file string, text string) (Definitions, error) {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	result := Definitions{Definitions: map[string]fdef.Definition{}, Bindings: map[string]Binding{}}
	var current *draft
	finished := map[string]bool{}
	table := ""
	for index := 0; index < len(lines); index++ {
		line := strings.TrimRight(lines[index], " \t")
		number := index + 1
		fail := func(reason string, detail string) (Definitions, error) {
			return Definitions{}, FileError{File: file, Line: number, Reason: reason, Detail: detail}
		}
		switch {
		case line == "" || (!strings.HasPrefix(line, "|") && table == ""):
		case line == definitionHeaderRow:
			if index+1 >= len(lines) || strings.TrimRight(lines[index+1], " \t") != definitionSeparator {
				return fail("bad header", "the separator row must follow the header")
			}
			table = "definitions"
			index++
		case line == bindingHeaderRow:
			if table == "" {
				return fail("bad header", "the definitions table must come first")
			}
			if index+1 >= len(lines) || strings.TrimRight(lines[index+1], " \t") != bindingSeparator {
				return fail("bad header", "the separator row must follow the header")
			}
			if current != nil {
				result.Definitions[current.id] = current.definition
				current = nil
			}
			table = "bindings"
			index++
		case !strings.HasPrefix(line, "|"):
		case table == "":
			return fail("bad header", "the header row must be "+definitionHeaderRow)
		case table == "definitions":
			next, err := definitionRow(file, number, line, current, finished, result.Definitions)
			if err != nil {
				return Definitions{}, err
			}
			current = next
		default:
			if err := bindingRow(file, number, line, &result); err != nil {
				return Definitions{}, err
			}
		}
	}
	if table == "" {
		return Definitions{}, FileError{File: file, Line: 1, Reason: "bad header", Detail: "the header row must be " + definitionHeaderRow}
	}
	if current != nil {
		result.Definitions[current.id] = current.definition
	}
	return result, nil
}

func definitionRow(file string, number int, line string, current *draft, finished map[string]bool, done map[string]fdef.Definition) (*draft, error) {
	fail := func(reason string, detail string) (*draft, error) {
		return nil, FileError{File: file, Line: number, Reason: reason, Detail: detail}
	}
	cells, ok := splitCells(line)
	if !ok || len(cells) != definitionColumns {
		return fail("wrong column count", fmt.Sprintf("a row has %d cells and one line", definitionColumns))
	}
	id, part, name, detail, message := cells[0], cells[1], cells[2], cells[3], cells[4]
	if id == "" || name == "" || part == "" {
		return fail("empty cell", "the id, the part and the name are required")
	}
	if current == nil || current.id != id {
		if finished[id] {
			return fail("definition split", fmt.Sprintf("the rows of %q are not together", id))
		}
		if current != nil {
			done[current.id] = current.definition
		}
		finished[id] = true
		current = &draft{id: id, copied: map[string]bool{}}
	}
	switch part {
	case partFrom:
		return current, copyFrom(file, number, current, name, done)
	case partInput:
		return current, addInput(file, number, current, name, detail)
	case partCheck:
		return current, addCheck(file, number, current, name, detail, message)
	}
	return fail("bad part", fmt.Sprintf("the part %q is input, check or from", part))
}

func copyFrom(file string, number int, current *draft, name string, done map[string]fdef.Definition) error {
	if len(current.definition.Inputs) != 0 || len(current.definition.Checks) != 0 {
		return FileError{File: file, Line: number, Reason: "bad part", Detail: "from must be the first row of a definition"}
	}
	base, found := done[name]
	if !found {
		return FileError{File: file, Line: number, Reason: "unknown definition", Detail: fmt.Sprintf("%q is not a definition above this row", name)}
	}
	current.definition = fdef.Definition{
		Inputs: append([]fdef.InputDecl(nil), base.Inputs...),
		Checks: append([]fdef.CheckDecl(nil), base.Checks...),
	}
	for _, input := range base.Inputs {
		current.copied["input:"+input.Name] = true
	}
	for i := range base.Checks {
		current.copied["check:"+strconv.Itoa(i+1)] = true
	}
	return nil
}

func addInput(file string, number int, current *draft, name string, detail string) error {
	fail := func(reason string, text string) error {
		return FileError{File: file, Line: number, Reason: reason, Detail: text}
	}
	if detail == "" {
		return fail("empty cell", "an input needs its type")
	}
	typeWords, rulesText, _ := strings.Cut(detail, ": ")
	kind, known := definitionTypes[strings.TrimSpace(typeWords)]
	if !known {
		return fail("bad type", fmt.Sprintf("the type %q is not one of the eight", typeWords))
	}
	declared := fdef.InputDecl{Name: name, Type: kind}
	if strings.TrimSpace(rulesText) != "" {
		parts, ok := splitOutside(rulesText, "; ")
		if !ok {
			return fail("bad rule", "a quote is not closed")
		}
		for _, part := range parts {
			rule, text := readRule(strings.TrimSpace(part))
			if text != "" {
				return fail("bad rule", text)
			}
			declared.Rules = append(declared.Rules, rule)
		}
	}
	key := "input:" + name
	for i, existing := range current.definition.Inputs {
		if existing.Name == name {
			if !current.copied[key] {
				return fail("duplicate name", fmt.Sprintf("input %q is written twice", name))
			}
			current.definition.Inputs[i] = declared
			current.copied[key] = false
			return nil
		}
	}
	current.definition.Inputs = append(current.definition.Inputs, declared)
	return nil
}

func addCheck(file string, number int, current *draft, name string, formulaText string, message string) error {
	fail := func(reason string, text string) error {
		return FileError{File: file, Line: number, Reason: reason, Detail: text}
	}
	position, err := strconv.Atoi(name)
	if err != nil || position < 1 {
		return fail("bad name", fmt.Sprintf("a check is named by its number, not %q", name))
	}
	if formulaText == "" || message == "" {
		return fail("empty cell", "a check needs its formula and its message")
	}
	declared := fdef.CheckDecl{Formula: formulaText, Message: message}
	key := "check:" + name
	if position <= len(current.definition.Checks) {
		if !current.copied[key] {
			return fail("duplicate name", fmt.Sprintf("check %d is written twice", position))
		}
		current.definition.Checks[position-1] = declared
		current.copied[key] = false
		return nil
	}
	if position != len(current.definition.Checks)+1 {
		return fail("bad name", fmt.Sprintf("check %d comes before check %d", position, len(current.definition.Checks)+1))
	}
	current.definition.Checks = append(current.definition.Checks, declared)
	return nil
}

func readRule(text string) (fdef.Rule, string) {
	word, rest, _ := strings.Cut(text, " ")
	rest = strings.TrimSpace(rest)
	switch word {
	case problem.RuleRequired, problem.RuleNoNegatives:
		if rest != "" {
			return fdef.Rule{}, fmt.Sprintf("the rule %s takes nothing after it", word)
		}
		return fdef.Rule{Name: word}, ""
	case problem.RuleDecimalPlaces:
		count, ok := countOf(rest)
		return fdef.Rule{Name: word, Places: &count}, failureIf(!ok, text)
	case problem.RuleMinItems, problem.RuleMaxItems:
		count, ok := countOf(rest)
		return fdef.Rule{Name: word, Items: &count}, failureIf(!ok, text)
	case problem.RuleRange:
		return readRange(rest)
	case problem.RuleAllowedValues:
		return readAllowedValues(rest)
	}
	return fdef.Rule{}, fmt.Sprintf("%q is not a rule", text)
}

func failureIf(failed bool, text string) string {
	if failed {
		return fmt.Sprintf("%q is not written as the notation says", text)
	}
	return ""
}

func countOf(text string) (int, bool) {
	if !countPattern.MatchString(text) {
		return 0, false
	}
	count, err := strconv.Atoi(text)
	return count, err == nil
}

func readRange(text string) (fdef.Rule, string) {
	rule := fdef.Rule{Name: problem.RuleRange}
	bad := fmt.Sprintf("%q is not a range", "range "+text)
	set := func(target **string, value string) bool {
		if !numberPattern.MatchString(value) {
			return false
		}
		*target = &value
		return true
	}
	switch {
	case strings.HasPrefix(text, "more than "):
		rule.MinExclusive = true
		if !set(&rule.Min, strings.TrimPrefix(text, "more than ")) {
			return fdef.Rule{}, bad
		}
	case strings.HasPrefix(text, "at least "):
		if !set(&rule.Min, strings.TrimPrefix(text, "at least ")) {
			return fdef.Rule{}, bad
		}
	case strings.HasPrefix(text, "less than "):
		rule.MaxExclusive = true
		if !set(&rule.Max, strings.TrimPrefix(text, "less than ")) {
			return fdef.Rule{}, bad
		}
	case strings.HasPrefix(text, "at most "):
		if !set(&rule.Max, strings.TrimPrefix(text, "at most ")) {
			return fdef.Rule{}, bad
		}
	default:
		low, high, found := strings.Cut(text, " to ")
		if !found || !set(&rule.Min, low) || !set(&rule.Max, high) {
			return fdef.Rule{}, bad
		}
	}
	return rule, ""
}

func readAllowedValues(text string) (fdef.Rule, string) {
	rule := fdef.Rule{Name: problem.RuleAllowedValues}
	if text == "" {
		return rule, ""
	}
	parts, ok := splitOutside(text, ", ")
	if !ok {
		return fdef.Rule{}, "a quote is not closed"
	}
	for _, part := range parts {
		value, rest, quoted := readQuoted(strings.TrimSpace(part))
		if !quoted || rest != "" {
			return fdef.Rule{}, fmt.Sprintf("%q is not a quoted allowed value", part)
		}
		rule.Values = append(rule.Values, value)
	}
	return rule, ""
}

func bindingRow(file string, number int, line string, result *Definitions) error {
	fail := func(reason string, detail string) error {
		return FileError{File: file, Line: number, Reason: reason, Detail: detail}
	}
	cells, ok := splitCells(line)
	if !ok || len(cells) != bindingColumns {
		return fail("wrong column count", fmt.Sprintf("a row has %d cells and one line", bindingColumns))
	}
	for _, cell := range cells {
		if cell == "" {
			return fail("empty cell", "every cell of a binding is required")
		}
	}
	name, definition, target := cells[0], cells[1], cells[2]
	if _, found := result.Definitions[definition]; !found {
		return fail("unknown definition", fmt.Sprintf("%q is not a definition of the file", definition))
	}
	if target != TargetRow && target != TargetDefinition {
		return fail("bad target", fmt.Sprintf("the target %q is row or definition", target))
	}
	if first, twice := result.Bindings[name]; twice {
		return fail("duplicate name", fmt.Sprintf("%q is also bound on line %d", name, first.Line))
	}
	result.Bindings[name] = Binding{Case: name, Definition: definition, Target: target, Line: number}
	return nil
}
