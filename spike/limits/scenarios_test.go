package limits

import (
	"strings"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type expectation struct {
	limits []string
	value  any
	strict []string
}

type scenario struct {
	id            string
	title         string
	formula       func(l Limits) string
	params        []harness.Param
	inputs        func(l Limits) map[string]any
	result        harness.Kind
	adjust        func(l Limits) Limits
	expect        expectation
	maxRefusal    time.Duration
	notApplicable map[string]string
}

var allEngines = []string{"own", "cel-go", "expr-lang"}

func fixed(text string) func(Limits) string {
	return func(Limits) string { return text }
}

func noInputs(Limits) map[string]any { return map[string]any{} }

func integerParam(name string) []harness.Param {
	return []harness.Param{{Name: name, Kind: harness.KindInteger}}
}

func refused(strict []string, limits ...string) expectation {
	return expectation{limits: limits, strict: strict}
}

func accepted(value any, strict []string) expectation {
	return expectation{value: value, strict: strict}
}

func scenarios() []scenario {
	return []scenario{
		{id: "P04", title: "brackets nested past the depth limit",
			formula: func(l Limits) string { return nestedMinus(l.Depth + 8) },
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitDepth)},
		{id: "P04b", title: "plain brackets nested past the depth limit",
			formula: func(l Limits) string {
				return strings.Repeat("(", l.Depth+8) + "1" + strings.Repeat(")", l.Depth+8)
			},
			inputs: noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitDepth)},
		{id: "P05", title: "a million unclosed brackets",
			formula: fixed(strings.Repeat("(", 1_000_000)),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitFormulaLength, LimitDepth, "parser"), maxRefusal: 2 * time.Second},
		{id: "P05s", title: "unclosed brackets inside the formula length",
			formula: func(l Limits) string { return strings.Repeat("(", l.FormulaLength-100) },
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitDepth, "parser")},
		{id: "P20", title: "empty formula",
			formula: fixed(""),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, "parser")},
		{id: "P20b", title: "blank formula",
			formula: fixed("  \n "),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, "parser")},
		{id: "L01", title: "one level over the depth limit",
			formula: func(l Limits) string { return nestedMinus(l.Depth) },
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitDepth)},
		{id: "L02", title: "exactly at the depth limit",
			formula: func(l Limits) string { return nestedMinus(l.Depth - 1) },
			inputs:  noInputs, result: harness.KindInteger,
			expect: accepted(nestedMinusValue(63), allEngines)},
		{id: "L03", title: "one node over the node limit",
			formula: func(l Limits) string { text, _ := formulaWithNodes(l.Nodes + 1); return text },
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused([]string{"own"}, LimitNode)},
		{id: "L04", title: "exactly at the node limit",
			formula: func(l Limits) string { text, _ := formulaWithNodes(l.Nodes); return text },
			inputs:  noInputs, result: harness.KindInteger,
			expect: accepted(int64(-500), []string{"own"})},
		{id: "L05", title: "work past the cost limit",
			formula: func(Limits) string { return balancedSum(128, "sum(range(1, 10000))") },
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitCost), maxRefusal: 500 * time.Millisecond},
		{id: "L06", title: "work past the time limit",
			formula: func(Limits) string { return balancedSum(16, "sum(range(1, 4000000))") },
			inputs:  noInputs, result: harness.KindInteger,
			adjust: func(l Limits) Limits {
				l.Cost, l.ListLength, l.Time = 0, 5_000_000, 30*time.Millisecond
				return l
			},
			expect: refused(allEngines, LimitTime), maxRefusal: 180 * time.Millisecond},
		{id: "C01", title: "a list of a billion entries",
			formula: fixed("count(range(1, 1000000000))"),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitListLength), maxRefusal: 500 * time.Millisecond},
		{id: "C02", title: "a list of a billion entries from an input",
			formula: fixed("count(range(1, quantity))"),
			params:  integerParam("quantity"),
			inputs:  func(Limits) map[string]any { return map[string]any{"quantity": int64(1_000_000_000)} },
			result:  harness.KindInteger,
			expect:  refused(allEngines, LimitListLength), maxRefusal: 500 * time.Millisecond},
		{id: "I01", title: "a long text in the formula",
			formula: func(l Limits) string { return quote(strings.Repeat("a", l.TextLength+100)) },
			inputs:  noInputs, result: harness.KindText,
			expect: refused(allEngines, LimitTextLength)},
		{id: "I02", title: "a long text as an input",
			formula: fixed(`s == "x"`),
			params:  []harness.Param{{Name: "s", Kind: harness.KindText}},
			inputs:  func(Limits) map[string]any { return map[string]any{"s": strings.Repeat("a", 1_000_000)} },
			result:  harness.KindBoolean,
			expect:  refused(allEngines, LimitTextLength)},
		{id: "I04", title: "a million-digit number",
			formula: fixed(digits(1_000_000)),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitFormulaLength, LimitNumberMagnitude), maxRefusal: 2 * time.Second},
		{id: "I04b", title: "a hundred-digit number",
			formula: fixed(digits(100)),
			inputs:  noInputs, result: harness.KindInteger,
			expect: refused(allEngines, LimitNumberMagnitude)},
		{id: "I05", title: "a number of 200000 digits as an input",
			formula: fixed("d"),
			params:  []harness.Param{{Name: "d", Kind: harness.KindDecimal}},
			inputs:  func(Limits) map[string]any { return map[string]any{"d": numeric.MustParse(digits(200_000))} },
			result:  harness.KindDecimal,
			expect:  refused(allEngines, LimitNumberMagnitude)},
		{id: "I06", title: "an exponent of a billion",
			formula: fixed("1e1000000000"),
			inputs:  noInputs, result: harness.KindDecimal,
			expect: refused(allEngines, LimitNumberMagnitude)},
		{id: "I09", title: "a million decimal places",
			formula: fixed("0." + strings.Repeat("1", 1_000_000)),
			inputs:  noInputs, result: harness.KindDecimal,
			expect: refused(allEngines, LimitFormulaLength, LimitDecimalPlaces), maxRefusal: 2 * time.Second},
		{id: "I09b", title: "forty decimal places",
			formula: fixed("0." + strings.Repeat("1", 40)),
			inputs:  noInputs, result: harness.KindDecimal,
			expect: refused(allEngines, LimitDecimalPlaces)},
		{id: "I10", title: "round to a million places",
			formula: fixed("round(1.0 / 3.0, 1000000)"),
			inputs:  noInputs, result: harness.KindDecimal,
			expect: refused(allEngines, LimitDecimalPlaces)},
		{id: "I10b", title: "round to a million places given as an input",
			formula: fixed("round(1.0 / 3.0, places)"),
			params:  integerParam("places"),
			inputs:  func(Limits) map[string]any { return map[string]any{"places": int64(1_000_000)} },
			result:  harness.KindDecimal,
			expect:  refused(allEngines, LimitDecimalPlaces)},
		{id: "I12", title: "a list input over the list length limit",
			formula: fixed("count(xs)"),
			params:  []harness.Param{{Name: "xs", Kind: harness.KindIntegerList}},
			inputs:  func(l Limits) map[string]any { return map[string]any{"xs": make([]int64, l.ListLength+1)} },
			result:  harness.KindInteger,
			expect:  refused(allEngines, LimitListLength)},
	}
}

func scenarioNamed(id string) scenario {
	for _, s := range scenarios() {
		if s.id == id {
			return s
		}
	}
	panic("no scenario " + id)
}

func (s scenario) limits(base Limits) Limits {
	if s.adjust != nil {
		return s.adjust(base)
	}
	return base
}
