package function

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func S(sentence string) string {
	return strings.ReplaceAll(sentence, "~", "`")
}

func str(value string) *string { return &value }

func whole(value int) *int { return &value }

func required() Rule { return Rule{Name: problem.RuleRequired} }

func noNegatives() Rule { return Rule{Name: problem.RuleNoNegatives} }

func places(count int) Rule { return Rule{Name: problem.RuleDecimalPlaces, Places: whole(count)} }

func minItems(count int) Rule { return Rule{Name: problem.RuleMinItems, Items: whole(count)} }

func maxItems(count int) Rule { return Rule{Name: problem.RuleMaxItems, Items: whole(count)} }

func between(min string, max string) Rule {
	return Rule{Name: problem.RuleRange, Min: str(min), Max: str(max)}
}

func moreThan(min string) Rule {
	return Rule{Name: problem.RuleRange, Min: str(min), MinExclusive: true}
}

func upTo(max string) Rule { return Rule{Name: problem.RuleRange, Max: str(max)} }

func allowed(values ...string) Rule { return Rule{Name: problem.RuleAllowedValues, Values: values} }

func input(name string, kind formula.Type, rules ...Rule) InputDecl {
	return InputDecl{Name: name, Type: kind, Rules: rules}
}

func check(formulaText string, message string) CheckDecl {
	return CheckDecl{Formula: formulaText, Message: message}
}

func row(pairs ...any) Row {
	received := Row{}
	for i := 0; i < len(pairs); i += 2 {
		name := pairs[i].(string)
		switch value := pairs[i+1].(type) {
		case string:
			received[name] = numtext.One(value)
		case []string:
			received[name] = numtext.Many(value...)
		}
	}
	return received
}

func with(base Row, pairs ...any) Row {
	changed := Row{}
	for name, received := range base {
		changed[name] = received
	}
	for name, received := range row(pairs...) {
		changed[name] = received
	}
	return changed
}

func without(base Row, names ...string) Row {
	changed := Row{}
	for name, received := range base {
		changed[name] = received
	}
	for _, name := range names {
		delete(changed, name)
	}
	return changed
}

func repeated(text string, count int) []string {
	items := make([]string, count)
	for i := range items {
		items[i] = text
	}
	return items
}

var ruleOfCode = map[string]string{
	problem.CodeRequired:      problem.RuleRequired,
	problem.CodeWrongType:     problem.RuleType,
	problem.CodeTooSmall:      problem.RuleRange,
	problem.CodeTooLarge:      problem.RuleRange,
	problem.CodeNoNegatives:   problem.RuleNoNegatives,
	problem.CodeDecimalPlaces: problem.RuleDecimalPlaces,
	problem.CodeNotAllowed:    problem.RuleAllowedValues,
	problem.CodeListTooShort:  problem.RuleMinItems,
	problem.CodeListTooLong:   problem.RuleMaxItems,
	problem.CodeCheckFailed:   problem.RuleCheck,
	problem.CodeCheckError:    problem.RuleCheck,
	problem.CodeUnknownInput:  problem.RuleKnownInput,
}

func at(code string, inputName string, item int, value string, sentence string) problem.Problem {
	given := true
	p := problem.Problem{
		Input:    &inputName,
		Given:    &given,
		Value:    &value,
		Rule:     ruleOfCode[code],
		Sentence: S(sentence),
		Code:     code,
	}
	if item > 0 {
		p.Item = &item
	}
	return p
}

func on(code string, inputName string, value string, sentence string) problem.Problem {
	return at(code, inputName, 0, value, sentence)
}

func missing(inputName string) problem.Problem {
	given := false
	return problem.Problem{
		Input:    &inputName,
		Given:    &given,
		Rule:     problem.RuleRequired,
		Sentence: S("~" + inputName + "~ was not given. It is required."),
		Code:     problem.CodeRequired,
	}
}

func notGiven(code string, inputName string, sentence string) problem.Problem {
	given := false
	return problem.Problem{
		Input:    &inputName,
		Given:    &given,
		Rule:     ruleOfCode[code],
		Sentence: S(sentence),
		Code:     code,
	}
}

func ofCheck(code string, formulaText string, sentence string) problem.Problem {
	return problem.Problem{
		Rule:     problem.RuleCheck,
		Sentence: S(sentence),
		Code:     code,
		Check:    &formulaText,
	}
}

func asJSON(t *testing.T, problems []problem.Problem) string {
	t.Helper()
	data, err := problem.ListJSON(problems)
	if err != nil {
		t.Fatal(err)
	}
	var pretty strings.Builder
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatal(err)
	}
	for i, item := range items {
		pretty.WriteString(fmt.Sprintf("%d %s\n", i+1, item))
	}
	return pretty.String()
}

func compare(t *testing.T, got []problem.Problem, want []problem.Problem) {
	t.Helper()
	if got == nil {
		t.Fatalf("the list is nil; a clean row gives an empty list")
	}
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if asJSON(t, got) != asJSON(t, want) {
		t.Errorf("problem lists differ\nwant:\n%s\ngot:\n%s", asJSON(t, want), asJSON(t, got))
	}
}

type fakeRunner struct {
	calls int
}

func (f *fakeRunner) Run(tree formula.Tree, values map[string]numtext.Value) Outcome {
	f.calls++
	result, failure := evaluate(tree.Root, values)
	if failure != "" {
		return Outcome{Failure: failure}
	}
	truth, isBool := result.(bool)
	if !isBool {
		return Outcome{Failure: "The check did not give true or false."}
	}
	return Outcome{True: truth}
}

func evaluate(node formula.Node, values map[string]numtext.Value) (any, string) {
	switch n := node.(type) {
	case formula.Literal:
		d, err := number.ParseDecimal(n.Value)
		if err != nil {
			return nil, "The test runner cannot read " + n.Value + "."
		}
		return d, ""
	case formula.Name:
		value := values[n.Name]
		if value.Type == numtext.IntegerList || value.Type == numtext.DecimalList {
			items := make([]number.Decimal, len(value.Items))
			for i, item := range value.Items {
				items[i], _ = number.ParseDecimal(item.Text)
			}
			return items, ""
		}
		d, _ := number.ParseDecimal(value.Text)
		return d, ""
	case formula.Call:
		list, failure := evaluate(n.Args[0], values)
		if failure != "" {
			return nil, failure
		}
		return number.DecimalFromInt(int64(len(list.([]number.Decimal)))), ""
	case formula.Binary:
		return evaluateBinary(n, values)
	}
	return nil, "The test runner does not know this part."
}

func evaluateBinary(n formula.Binary, values map[string]numtext.Value) (any, string) {
	left, failure := evaluate(n.Left, values)
	if failure != "" {
		return nil, failure
	}
	right, failure := evaluate(n.Right, values)
	if failure != "" {
		return nil, failure
	}
	if n.Op == "&&" {
		return left.(bool) && right.(bool), ""
	}
	a, b := left.(number.Decimal), right.(number.Decimal)
	switch n.Op {
	case "+":
		return a.Add(b), ""
	case "-":
		return a.Sub(b), ""
	case "*":
		return a.Mul(b), ""
	case "/":
		if b.IsZero() {
			return nil, "The part ~" + formula.Format(formula.Tree{Root: n}) + "~ divides by zero."
		}
		quotient, _ := a.Div(b, 20, number.RoundHalfUp)
		return quotient, ""
	case ">":
		return a.Cmp(b) > 0, ""
	case ">=":
		return a.Cmp(b) >= 0, ""
	case "<":
		return a.Cmp(b) < 0, ""
	case "<=":
		return a.Cmp(b) <= 0, ""
	case "==":
		return a.Cmp(b) == 0, ""
	}
	return nil, "The test runner does not know the operator " + n.Op + "."
}

func runCase(t *testing.T, definition Definition, received Row, want ...problem.Problem) {
	t.Helper()
	runner := &fakeRunner{}
	got := Validate(definition, received, runner)
	compare(t, got, want)
}

var (
	d1 = Definition{
		Inputs: []InputDecl{
			input("balance", formula.TypeDecimal, required(), noNegatives(), places(2)),
			input("payment", formula.TypeDecimal, required(), noNegatives(), places(2)),
		},
		Checks: []CheckDecl{
			check("payment > 0", "The payment must be more than zero"),
			check("payment <= balance + 1000.00", "The payment cannot be more than 1000.00 above the balance"),
		},
	}
	d1Clean = row("balance", "120.50", "payment", "45.25")

	d2 = Definition{
		Inputs: []InputDecl{
			input("kilos", formula.TypeInteger, required(), noNegatives()),
			input("price_per_kilo", formula.TypeDecimal, required(), moreThan("0.000"), places(3)),
			input("credit", formula.TypeDecimal, required(), noNegatives(), places(2)),
		},
	}
	d2Clean = row("kilos", "853", "price_per_kilo", "0.437", "credit", "120.00")

	d3 = Definition{
		Inputs: []InputDecl{
			input("status", formula.TypeText, required(), allowed("active", "lapsed", "suspended")),
			input("requested", formula.TypeDecimal, required(), moreThan("0.00"), places(2)),
			input("contributions", formula.TypeListDecimal, minItems(1), noNegatives()),
		},
	}
	d3Clean = row("status", "active", "requested", "1400.00", "contributions", []string{"200.00", "250.00", "150.00", "100.00"})

	d4 = Definition{
		Inputs: []InputDecl{
			input("scores", formula.TypeListInteger, minItems(1), maxItems(12), between("0", "100")),
		},
	}
	d4Clean = row("scores", []string{"72", "85", "64", "91", "77"})

	d5 = Definition{
		Inputs: []InputDecl{
			input("balances", formula.TypeListDecimal, minItems(1), places(2)),
			input("floor", formula.TypeDecimal, required(), noNegatives(), places(2)),
		},
		Checks: []CheckDecl{check("count(balances) >= 6", "At least 6 balances are needed")},
	}
	d5Clean = row("balances", []string{"420.00", "380.50", "95.25", "510.00", "440.75", "300.00"}, "floor", "350.00")

	d6 = Definition{
		Inputs: []InputDecl{
			input("principal", formula.TypeDecimal, required(), moreThan("0.00"), places(2)),
			input("monthly_rate", formula.TypeDecimal, required(), between("0.000", "0.050"), places(3)),
			input("months", formula.TypeInteger, required(), between("1", "60")),
		},
		Checks: []CheckDecl{
			check("principal >= 100.00", "The loan must be at least 100.00"),
			check("principal * monthly_rate * months <= principal", "The interest cannot be more than the amount borrowed"),
		},
	}
	d6Clean = row("principal", "2400.00", "monthly_rate", "0.012", "months", "7")

	d7 = Definition{
		Inputs: []InputDecl{
			input("total", formula.TypeDecimal, required(), noNegatives(), places(2)),
			input("people", formula.TypeInteger, required()),
		},
		Checks: []CheckDecl{check("total / people >= 1.00", "Each person must get at least 1.00")},
	}
	d7Clean = row("total", "100.00", "people", "3")

	d8 = Definition{
		Inputs: []InputDecl{
			input("kilos", formula.TypeInteger, required()),
			input("payment", formula.TypeDecimal, required()),
			input("ratio", formula.TypeFloat, required()),
			input("note", formula.TypeText, required()),
			input("active", formula.TypeBoolean, required()),
			input("filed", formula.TypeDate, required()),
		},
	}
	d8Clean = row("kilos", "12", "payment", "45.25", "ratio", "0.37", "note", "ok", "active", "true", "filed", "2026-03-31")

	d9 = Definition{
		Inputs: []InputDecl{
			input("deposits", formula.TypeListDecimal, minItems(1), noNegatives()),
			input("expenses", formula.TypeListDecimal, noNegatives()),
			input("members", formula.TypeInteger, required(), between("1", "500")),
		},
	}
	d9Clean = row("deposits", []string{"500.00", "750.00", "1250.50"}, "expenses", []string{"120.25", "80.00"}, "members", "7")

	d10 = Definition{
		Inputs: []InputDecl{
			input("units", formula.TypeInteger, required(), noNegatives(), between("1", "100")),
			input("share_pct", formula.TypeDecimal, required(), Rule{Name: problem.RuleRange, Min: str("0.00"), Max: str("100.00"), MaxExclusive: true}),
			input("bonus", formula.TypeDecimal, required(), upTo("500.00")),
			input("bags", formula.TypeDecimal, required(), places(0)),
			input("weight", formula.TypeDecimal, required(), places(1)),
		},
	}
	d10Clean = row("units", "5", "share_pct", "40.00", "bonus", "100.00", "bags", "3", "weight", "2.5")

	d11 = Definition{
		Inputs: []InputDecl{
			input("amounts", formula.TypeListDecimal, minItems(1), noNegatives(), between("0.00", "100.00"), places(2)),
		},
	}
	d11Clean = row("amounts", []string{"10.00", "25.50"})
)
