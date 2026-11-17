package inputchecks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const cutTextValue = "the page cuts the sentence but does not say whether the value field of a text over 200 characters is cut (new question 11 on the suite page)"

type typeTable struct {
	prefix   string
	input    string
	words    string
	quoted   bool
	accepted []string
	negative []string
	refused  []string
}

func wrongType(inputName string, received string, words string, quoted bool) problem.Problem {
	given := true
	shown := received
	if quoted {
		shown = "\"" + received + "\""
	}
	value := received
	return problem.Problem{
		Input:    &inputName,
		Given:    &given,
		Value:    &value,
		Rule:     "type",
		Sentence: fmt.Sprintf("`%s` is %s. It must be %s.", inputName, shown, words),
		Code:     "wrong-type",
	}
}

func tableCases(table typeTable) []testCase {
	var cases []testCase
	for i, text := range table.accepted {
		cases = append(cases, clean(fmt.Sprintf("ic-%s-ok-%d", table.prefix, i+1), d8, with(d8.clean, table.input, text), unchecked))
	}
	for i, text := range table.negative {
		c := clean(fmt.Sprintf("ic-%s-neg-%d", table.prefix, i+1), d8, with(d8.clean, table.input, text), unchecked)
		cases = append(cases, hold(c, decision22))
	}
	for i, text := range table.refused {
		cases = append(cases, testCase{
			id:   fmt.Sprintf("ic-%s-%d", table.prefix, i+1),
			def:  d8,
			row:  with(d8.clean, table.input, text),
			want: []problem.Problem{wrongType(table.input, text, table.words, table.quoted)},
			ran:  unchecked,
		})
	}
	return cases
}

func casesK() []testCase {
	whole := typeTable{
		prefix:   "k-int",
		input:    "kilos",
		words:    "a whole number with at most 18 digits",
		accepted: []string{"0", "12", "123456789012345678"},
		negative: []string{"-12"},
		refused:  []string{"-0", "+5", "007", "1237.5", "1e3", "1,200", "1 200", "1234567890123456789", "abc"},
	}
	exact := typeTable{
		prefix:   "k-dec",
		input:    "payment",
		words:    "a decimal number with at most 30 digits, such as 12.50",
		accepted: []string{"0", "12.50", "-0.00", "0.5", "123456789012345678901234567890"},
		negative: []string{"-12.5"},
		refused: []string{".5", "5.", "+5", "05.5", "5.5.5", "1e3", "1,5", "1,200.50", "95%", "1 200,50", "12.50 ",
			"1234567890123456789012345678901", "NaN"},
	}
	approximate := typeTable{
		prefix:   "k-flt",
		input:    "ratio",
		words:    "a number with at most 30 digits, such as 0.37",
		accepted: []string{"0.37", "5"},
		negative: []string{"-1.5"},
		refused:  []string{".37", "0,37", "1e3"},
	}
	var cases []testCase
	for _, table := range []typeTable{whole, exact, approximate} {
		cases = append(cases, tableCases(table)...)
	}
	return cases
}

func casesL() []testCase {
	letters := strings.Repeat("a", 200)
	words := typeTable{
		prefix:   "l-text",
		input:    "note",
		words:    "text of at most 200 characters",
		quoted:   true,
		accepted: []string{"ok", "", letters},
	}
	cases := tableCases(words)
	cut := wrongType("note", letters+"a", "text of at most 200 characters", true)
	cut.Sentence = "`note` is \"" + letters + "...\". It must be text of at most 200 characters."
	cases = append(cases,
		unspecified(testCase{id: "ic-l-text-1", def: d8, row: with(d8.clean, "note", letters+"a"), want: []problem.Problem{cut}, ran: unchecked}, cutTextValue),
		testCase{id: "ic-l-text-2", def: d8, row: with(d8.clean, "note", "a\tb"), want: []problem.Problem{wrongType("note", "a\tb", "text of at most 200 characters", true)}, ran: unchecked},
		testCase{id: "ic-l-text-3", def: d8, row: with(d8.clean, "note", "a\nb"), want: []problem.Problem{wrongType("note", "a\nb", "text of at most 200 characters", true)}, ran: unchecked},
	)
	yesNo := typeTable{
		prefix:   "l-bool",
		input:    "active",
		words:    "true or false",
		accepted: []string{"true", "false"},
		refused:  []string{"True", "TRUE", "yes", "1"},
	}
	cases = append(cases, tableCases(yesNo)...)
	date := typeTable{
		prefix:   "l-date",
		input:    "filed",
		words:    "a date written like 2026-03-31",
		accepted: []string{"2026-03-31", "2028-02-29"},
		refused:  []string{"2026-02-30", "2027-02-29", "2026-13-01", "2026-3-1", "31/03/2026"},
	}
	cases = append(cases, tableCases(date)...)
	cases = append(cases,
		bad("ic-l-list-1", d1, row("balance", "80.00", "payment", []string{"45.25"}), 0,
			"wrong-type · payment · [45.25] — ~payment~ is [45.25]. It must be a decimal number with at most 30 digits, such as 12.50."),
		bad("ic-l-list-2", d3, with(d3.clean, "contributions", "200.00"), unchecked,
			"wrong-type · contributions · 200.00 — ~contributions~ is 200.00. It must be a list of decimal numbers."),
		bad("ic-l-list-3", d3, with(d3.clean, "contributions", []string{"200.00", "1,200.50", "100.00"}), unchecked,
			"wrong-type · contributions item 2 · 1,200.50 — Item 2 of ~contributions~ is 1,200.50. It must be a decimal number with at most 30 digits, such as 12.50."),
	)
	return cases
}

func TestSectionKANumberWrittenAsText(t *testing.T) { runCases(t, casesK()) }
func TestSectionLTheOtherTypes(t *testing.T)        { runCases(t, casesL()) }

func allGroups() [][]testCase {
	return [][]testCase{casesA(), casesB(), casesC(), casesD(), casesE(), casesF(), casesG(), casesH(), casesI(), casesJ(), casesK(), casesL()}
}

type calculation struct {
	problems []problem.Problem
	result   string
	worked   bool
}

type calculator struct {
	runs int
}

func (c *calculator) work(body formula.Tree, d definition, received function.Row) (string, bool) {
	c.runs++
	values := map[string]numtext.Value{}
	for _, decl := range d.body.Inputs {
		kind := numtext.Decimal
		if decl.Type == formula.TypeInteger {
			kind = numtext.Integer
		}
		value, refusals := numtext.Read(kind, received[decl.Name])
		if len(refusals) > 0 {
			return refusals[0].Error(), false
		}
		values[decl.Name] = value
	}
	result, failure := evaluate(body, values)
	if failure != "" {
		return failure, false
	}
	return result.String(), true
}

func answer(t *testing.T, calc *calculator, checks *engine, validator function.Validator, d definition, bodyText string, received function.Row) calculation {
	t.Helper()
	problems := validator.Validate(received, checks)
	if len(problems) > 0 {
		return calculation{problems: problems}
	}
	body, parseProblems := formula.Parse(bodyText)
	if len(parseProblems) > 0 {
		t.Fatalf("the body %q does not read: %v", bodyText, parseProblems)
	}
	result, worked := calc.work(body, d, received)
	return calculation{problems: problems, result: result, worked: worked}
}

const balanceAfterPayment = "balance - payment"

func TestIcM01AFormulaWithNoProblemIsCalculated(t *testing.T) {
	calc, checks := &calculator{}, &engine{}
	got := answer(t, calc, checks, function.Prepare(d1.body), d1, balanceAfterPayment, d1.clean)
	if len(got.problems) != 0 || !got.worked || got.result != "75.25" {
		t.Fatalf("got %+v, want the result 75.25", got)
	}
	if calc.runs != 1 {
		t.Fatalf("the calculation ran %d times, want 1", calc.runs)
	}
}

func TestIcM02ARowWithAProblemIsNotCalculated(t *testing.T) {
	calc, checks := &calculator{}, &engine{}
	got := answer(t, calc, checks, function.Prepare(d1.body), d1, balanceAfterPayment, row("balance", "80.00", "payment", "-5.00"))
	want := lines("no-negatives · payment · -5.00 — ~payment~ is -5.00. It cannot be negative.")
	if show(got.problems) != show(want) || got.worked || got.result != "" {
		t.Fatalf("got %+v, want only the problem list\n%s", got, show(want))
	}
	if calc.runs != 0 {
		t.Fatalf("the calculation ran %d times, want 0", calc.runs)
	}
	if checks.calls != 0 {
		t.Fatalf("a check ran %d times, want 0: both read the broken input", checks.calls)
	}
}

func TestIcM03ARowWhoseOnlyProblemIsACheckIsNotCalculated(t *testing.T) {
	calc, checks := &calculator{}, &engine{}
	got := answer(t, calc, checks, function.Prepare(d1.body), d1, balanceAfterPayment, row("balance", "120.50", "payment", "0.00"))
	want := lines("check-failed · check \"payment > 0\" — The payment must be more than zero. ~payment~ is 0.00.")
	if show(got.problems) != show(want) || got.worked {
		t.Fatalf("got %+v, want\n%s", got, show(want))
	}
	if calc.runs != 0 {
		t.Fatalf("the calculation ran %d times, want 0", calc.runs)
	}
}

func TestIcM04ABatchOfThreeRowsKeepsEachRowsOwnProblems(t *testing.T) {
	calc, checks := &calculator{}, &engine{}
	validator := function.Prepare(d1.body)
	rows := []function.Row{
		row("balance", "120.50", "payment", "45.25"),
		row("balance", "-1.00", "payment", "-5.125"),
		row("balance", "120.50", "payment", "0.00"),
	}
	wantProblems := [][]problem.Problem{
		{},
		lines(
			"no-negatives · balance · -1.00 — ~balance~ is -1.00. It cannot be negative.",
			"no-negatives · payment · -5.125 — ~payment~ is -5.125. It cannot be negative.",
			"decimal-places · payment · -5.125 — ~payment~ is -5.125. It can have at most 2 decimal places."),
		lines("check-failed · check \"payment > 0\" — The payment must be more than zero. ~payment~ is 0.00."),
	}
	wantWorked := []bool{true, false, false}
	for i, received := range rows {
		got := answer(t, calc, checks, validator, d1, balanceAfterPayment, received)
		if show(got.problems) != show(wantProblems[i]) {
			t.Errorf("row %d problems differ\nwant:\n%s\ngot:\n%s", i+1, show(wantProblems[i]), show(got.problems))
		}
		if got.worked != wantWorked[i] {
			t.Errorf("row %d worked is %v, want %v", i+1, got.worked, wantWorked[i])
		}
		if i == 0 && got.result != "75.25" {
			t.Errorf("row 1 result is %q, want 75.25", got.result)
		}
	}
	if calc.runs != 1 {
		t.Errorf("the calculation ran %d times, want once for row 1", calc.runs)
	}
}
