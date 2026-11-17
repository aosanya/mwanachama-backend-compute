package hostile

import (
	"regexp"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type library struct {
	functions map[string]formula.Signature
}

func libraryKey(lib string, name string, version int) string {
	return lib + "/" + name + "@" + string(rune('0'+version))
}

func hostileLibrary() library {
	functions := map[string]formula.Signature{
		libraryKey("fin", "pmt", 1): {
			Params: []formula.Param{{Name: "principal", Type: formula.TypeDecimal}, {Name: "rate", Type: formula.TypeDecimal}, {Name: "months", Type: formula.TypeInteger}},
			Result: formula.TypeDecimal,
			Cost:   10,
		},
	}
	for depth, name := range []string{"step_a", "step_b", "step_c", "step_d", "step_e"} {
		functions[libraryKey("", name, 1)] = formula.Signature{
			Params:    []formula.Param{{Name: "p", Type: formula.TypeDecimal}},
			Result:    formula.TypeDecimal,
			Cost:      5 + depth,
			CallDepth: depth,
		}
	}
	return library{functions: functions}
}

func (l library) Function(lib string, name string, version int) (formula.Signature, bool) {
	signature, found := l.functions[libraryKey(lib, name, version)]
	return signature, found
}

func (l library) Versions(lib string, name string) []int {
	var versions []int
	for version := 1; version < 10; version++ {
		if _, found := l.functions[libraryKey(lib, name, version)]; found {
			versions = append(versions, version)
		}
	}
	return versions
}

func (l library) Apply(lib string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	return formula.DecimalValue(number.DecimalFromInt(0)), nil
}

func standardInputs() []formula.Input {
	return []formula.Input{
		{Name: "total", Type: formula.TypeDecimal},
		{Name: "people", Type: formula.TypeInteger},
		{Name: "flag", Type: formula.TypeBoolean},
		{Name: "rows", Type: formula.TypeListDecimal},
		{Name: "label", Type: formula.TypeText},
	}
}

func standardValues(rows int) map[string]formula.Value {
	items := make([]number.Decimal, rows)
	for i := range items {
		items[i] = number.NewDecimal(int64(i%100), 2)
	}
	return map[string]formula.Value{
		"total":  formula.DecimalValue(number.NewDecimal(1000, 2)),
		"people": formula.IntegerValue(3),
		"flag":   formula.BooleanValue(true),
		"rows":   formula.DecimalListValue(items),
		"label":  formula.TextValue("farm"),
	}
}

func atSave(text string, inputs []formula.Input) []formula.Problem {
	tree, problems := formula.Parse(text)
	if len(problems) > 0 {
		return problems
	}
	return formula.Check(tree, inputs, hostileLibrary())
}

func plan(t *testing.T, text string, inputs []formula.Input) *formula.Program {
	t.Helper()
	tree, problems := formula.Parse(text)
	if len(problems) > 0 {
		t.Fatalf("Parse(%.60q) gave %+v", text, problems)
	}
	program, problems := formula.Plan(tree, inputs, hostileLibrary(), formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("Plan(%.60q) gave %+v", text, problems)
	}
	return program
}

func shown(problem formula.Problem) string {
	return problem.Code + ": " + problem.Sentence
}

var plainSentence = regexp.MustCompile(`^[^\n\r]+$`)

func requirePlain(t *testing.T, problem formula.Problem) {
	t.Helper()
	if !plainSentence.MatchString(problem.Sentence) {
		t.Fatalf("sentence is not one plain line: %q", problem.Sentence)
	}
	for _, trace := range []string{"goroutine", "panic", "runtime.", ".go:"} {
		if strings.Contains(problem.Sentence, trace) {
			t.Fatalf("sentence looks like a trace: %q", problem.Sentence)
		}
	}
}

func nested(depth int, open string, close string, core string) string {
	return strings.Repeat(open, depth) + core + strings.Repeat(close, depth)
}

func limitSentence(name string, value int) string {
	return "This formula goes over the " + name + " limit of " + itoa(value) + "."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	return digits
}

func balancedSum(leaves int, leaf string) string {
	if leaves == 1 {
		return leaf
	}
	half := leaves / 2
	return "(" + balancedSum(half, leaf) + " + " + balancedSum(leaves-half, leaf) + ")"
}

func manyItems(count int) string {
	return "sum([" + strings.TrimSuffix(strings.Repeat("1, ", count), ", ") + "])"
}
