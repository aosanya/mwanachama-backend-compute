package basicruns

import (
	"sort"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type facts struct {
	binary      map[string]bool
	unary       map[string]bool
	builtins    map[string]bool
	inputTypes  map[string]bool
	resultTypes map[string]bool
	flags       map[string]bool
}

func newFacts() facts {
	return facts{
		binary:      map[string]bool{},
		unary:       map[string]bool{},
		builtins:    map[string]bool{},
		inputTypes:  map[string]bool{},
		resultTypes: map[string]bool{},
		flags:       map[string]bool{},
	}
}

var builtinNames = []string{"round", "sum", "count", "avg", "min", "max", "to_float", "to_decimal", "days_between", "months_between", "years_between"}

var binaryOperators = []string{"+", "-", "*", "/", "%", "<", "<=", ">", ">=", "==", "!=", "&&", "||"}

var typeNames = []string{"int", "dec", "float", "text", "bool", "date", "list<int>", "list<dec>"}

func isBuiltin(name string) bool {
	for _, builtin := range builtinNames {
		if builtin == name {
			return true
		}
	}
	return false
}

func (f facts) add(c Case) {
	tree, problems := formula.Parse(c.Formula)
	if len(problems) == 0 {
		walk(tree.Root, func(n formula.Node) {
			switch node := n.(type) {
			case formula.Binary:
				f.binary[node.Op] = true
			case formula.Unary:
				f.unary[node.Op] = true
			case formula.Conditional:
				f.flags["conditional"] = true
			case formula.List:
				f.flags["list written in a formula"] = true
			case formula.Call:
				switch {
				case node.Library != nil:
					f.flags["library function call"] = true
				case isBuiltin(node.Name):
					f.builtins[node.Name] = true
				default:
					f.flags["agency function call"] = true
				}
				if node.Version != nil {
					f.flags["call with a version"] = true
				}
			}
		})
	}
	list, err := parseInputs(c.Inputs)
	if err == nil {
		for _, item := range list {
			f.inputTypes[shortType(item.value.Type())] = true
		}
	}
	if strings.HasPrefix(c.Expect, "result: ") {
		prefix, _, _ := strings.Cut(strings.TrimPrefix(c.Expect, "result: "), ":")
		f.resultTypes[prefix] = true
	}
}

type languagePart struct {
	id   string
	seen func(f facts) bool
}

func languageParts() []languagePart {
	var parts []languagePart
	for _, op := range binaryOperators {
		op := op
		parts = append(parts, languagePart{"operator " + op, func(f facts) bool { return f.binary[op] }})
	}
	parts = append(parts, languagePart{"operator !", func(f facts) bool { return f.unary["!"] }})
	for _, flag := range []string{"conditional", "list written in a formula", "agency function call", "library function call", "call with a version"} {
		flag := flag
		parts = append(parts, languagePart{flag, func(f facts) bool { return f.flags[flag] }})
	}
	for _, name := range builtinNames {
		name := name
		parts = append(parts, languagePart{"built-in " + name, func(f facts) bool { return f.builtins[name] }})
	}
	for _, name := range typeNames {
		name := name
		parts = append(parts, languagePart{"input of type " + name, func(f facts) bool { return f.inputTypes[name] }})
		parts = append(parts, languagePart{"result of type " + name, func(f facts) bool { return f.resultTypes[name] }})
	}
	return parts
}

var partsWaitingOnAGap = map[string]int{
	"result of type date":      3,
	"result of type list<int>": 4,
	"result of type list<dec>": 4,
}

type Coverage struct {
	Covered       []string
	HeldOnly      []string
	NoCase        []string
	StaleGap      []string
	UnknownGap    []string
	GapsOfThePage map[int]string
}

func isResultCase(c Case) bool {
	return strings.HasPrefix(c.Expect, "result: ") || c.Expect == "accepted"
}

func coverageOf(page Page, resolution Resolution) Coverage {
	ran := newFacts()
	for _, c := range resolution.Run {
		if isResultCase(c) {
			ran.add(c)
		}
	}
	waiting := newFacts()
	for _, held := range resolution.Held {
		if isResultCase(held.Case) {
			waiting.add(held.Case)
		}
	}
	report := Coverage{GapsOfThePage: page.Gaps}
	for _, part := range languageParts() {
		gap, listed := partsWaitingOnAGap[part.id]
		switch {
		case part.seen(ran):
			report.Covered = append(report.Covered, part.id)
			if listed {
				report.StaleGap = append(report.StaleGap, part.id)
			}
		case part.seen(waiting):
			report.HeldOnly = append(report.HeldOnly, part.id)
		case listed:
			if _, found := page.Gaps[gap]; !found {
				report.UnknownGap = append(report.UnknownGap, part.id)
			}
		default:
			report.NoCase = append(report.NoCase, part.id)
		}
	}
	return report
}

func TestEveryPartOfTheLanguageHasACaseOrIsAListedGap(t *testing.T) {
	page, resolution := loadResolution(t)
	report := coverageOf(page, resolution)
	for _, id := range report.NoCase {
		t.Errorf("%s has no case on the page and is not listed as a gap", id)
	}
	for _, id := range report.StaleGap {
		t.Errorf("%s now has a case, so its gap is out of date", id)
	}
	for _, id := range report.UnknownGap {
		t.Errorf("%s names a gap number that the page does not list", id)
	}
	t.Logf("covered %d parts; waiting on a held case: %v", len(report.Covered), report.HeldOnly)
}

func namesOfCasesUsing(resolution Resolution, marker string) []string {
	var names []string
	for _, c := range resolution.Run {
		if strings.Contains(" "+c.Formula+" ", marker) {
			names = append(names, c.Name)
		}
	}
	sort.Strings(names)
	return names
}

func withoutRowsNamed(text string, names []string) string {
	var kept []string
	for _, line := range strings.Split(text, "\n") {
		drop := false
		if strings.HasPrefix(line, "|") {
			for _, name := range names {
				if strings.Contains(line, "`"+name+"`") || strings.HasPrefix(line, "| "+name+" ") {
					drop = true
				}
			}
		}
		if !drop {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func TestCoverageFailsWhenTheRowsOfAnOperatorAreRemoved(t *testing.T) {
	_, resolution := loadResolution(t)
	names := namesOfCasesUsing(resolution, " % ")
	if len(names) == 0 {
		t.Fatal("no case uses the remainder operator")
	}
	pageText := withoutRowsNamed(readFile(t, basicPagePath), names)
	examplesText := withoutRowsNamed(readFile(t, examplesPagePath), names)
	page := loadPage(pageText)
	copied := resolve(page, loadExamples(examplesText))
	report := coverageOf(page, copied)
	if !contains(report.NoCase, "operator %") {
		t.Fatalf("removing %v left the remainder operator covered: no-case list is %v", names, report.NoCase)
	}
}

func TestCoverageFailsWhenAGapIsRemovedFromThePage(t *testing.T) {
	text := readFile(t, basicPagePath)
	cut := strings.Index(text, "3. A date as a result.")
	if cut < 0 {
		t.Fatal("the page has no gap 3")
	}
	end := strings.Index(text[cut:], "\n")
	pageText := text[:cut] + text[cut+end+1:]
	page := loadPage(pageText)
	copied := resolve(page, loadExamples(readFile(t, examplesPagePath)))
	report := coverageOf(page, copied)
	if !contains(report.UnknownGap, "result of type date") {
		t.Fatalf("removing gap 3 left the date result unnoticed: %+v", report)
	}
}

func contains(list []string, id string) bool {
	for _, item := range list {
		if item == id {
			return true
		}
	}
	return false
}
