package batch

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func decimalInput(name string, places int) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeDecimal, Rules: []function.Rule{
		{Name: problem.RuleRequired},
		{Name: problem.RuleNoNegatives},
		{Name: problem.RuleDecimalPlaces, Places: &places},
	}}
}

func integerInput(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeInteger, Rules: []function.Rule{{Name: problem.RuleRequired}}}
}

func monthEnd() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2), decimalInput("rate", 4)}}
}

const monthEndBody = "round(balance * rate, 2)"

func evenShare() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("total", 2), integerInput("people")}}
}

const evenShareBody = "round(total / people, 2)"

func loanBalance() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2), decimalInput("payment", 2)}}
}

const loanBalanceBody = "payment >= balance ? 0.00 : balance - payment"

func fundShare() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{
		decimalInput("contribution", 2), decimalInput("total_contributions", 2), decimalInput("payout", 2),
	}}
}

const fundShareBody = "round(contribution / total_contributions * payout, 2)"

func fundShareFromList() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{
		decimalInput("contribution", 2),
		{Name: "contributions", Type: formula.TypeListDecimal, Rules: []function.Rule{{Name: problem.RuleRequired}}},
		decimalInput("payout", 2),
	}}
}

const fundShareFromListBody = "round(contribution / sum(contributions) * payout, 2)"

func mustCompile(t *testing.T, definition function.Definition, body string) *Compiled {
	t.Helper()
	compiled, problems := Compile(definition, body, nil, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("Compile(%q) refused: %v", body, problems)
	}
	return compiled
}

func row(pairs ...string) function.Row {
	result := function.Row{}
	for i := 0; i+1 < len(pairs); i += 2 {
		result[pairs[i]] = numtext.One(pairs[i+1])
	}
	return result
}

func cents(total int) string {
	return fmt.Sprintf("%d.%02d", total/100, total%100)
}

func monthEndRows(count int) []function.Row {
	rows := make([]function.Row, count)
	for i := range rows {
		rows[i] = row("balance", cents(1000*(i+1)), "rate", "0.005")
	}
	return rows
}

func monthEndAnswer(position int) string {
	return cents(5 * position)
}

func codesOf(problems []problem.Problem) []string {
	codes := make([]string, len(problems))
	for i, found := range problems {
		codes[i] = found.Code
	}
	return codes
}

func valueText(t *testing.T, result RowResult) string {
	t.Helper()
	if len(result.Problems) > 0 {
		t.Fatalf("row %d has problems %v, wanted a value", result.Row, codesOf(result.Problems))
	}
	return result.Value.String()
}

func requireValue(t *testing.T, result RowResult, want string) {
	t.Helper()
	if got := valueText(t, result); got != want {
		t.Fatalf("row %d gave %s, want %s", result.Row, got, want)
	}
}

func requireProblems(t *testing.T, result RowResult, stage Stage, codes ...string) {
	t.Helper()
	if result.StoppedAt != stage {
		t.Fatalf("row %d stopped at %q, want %q (problems %v)", result.Row, result.StoppedAt, stage, codesOf(result.Problems))
	}
	if got := codesOf(result.Problems); !reflect.DeepEqual(got, codes) {
		t.Fatalf("row %d problem codes %v, want %v", result.Row, got, codes)
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("the call returned an error: %v", err)
	}
}

func requireRowNumbers(t *testing.T, result Result, rows int) {
	t.Helper()
	if len(result.Rows) != rows {
		t.Fatalf("batch row count differs: sent %d rows, got %d results", rows, len(result.Rows))
	}
	for i, found := range result.Rows {
		if found.Row != i+1 {
			t.Fatalf("row number wrong: result at position %d carries Row %d", i+1, found.Row)
		}
	}
}

func requireSameAnswer(t *testing.T, label string, got RowResult, want RowResult) {
	t.Helper()
	if got.StoppedAt != want.StoppedAt {
		t.Fatalf("%s: stage differs, alone %q, in the batch %q", label, want.StoppedAt, got.StoppedAt)
	}
	if len(got.Problems) != len(want.Problems) {
		t.Fatalf("%s: problem count differs, alone %d, in the batch %d", label, len(want.Problems), len(got.Problems))
	}
	for i := range got.Problems {
		if !reflect.DeepEqual(got.Problems[i], want.Problems[i]) {
			t.Fatalf("%s: problem %d differs, alone %+v, in the batch %+v", label, i+1, want.Problems[i], got.Problems[i])
		}
	}
	if len(got.Problems) == 0 {
		if got.Value.Type() != want.Value.Type() || got.Value.String() != want.Value.String() {
			t.Fatalf("%s: value differs, alone %s %s, in the batch %s %s", label,
				want.Value.Type(), want.Value.String(), got.Value.Type(), got.Value.String())
		}
	}
}

func exempt(result RowResult) bool {
	return result.StoppedAt == StageBatch
}

func requireEqualsSingle(t *testing.T, compiled *Compiled, rows []function.Row, shared function.Row, result Result) {
	t.Helper()
	for i, found := range result.Rows {
		if exempt(found) {
			if len(found.Problems) != 1 || found.Value.String() != "" {
				t.Fatalf("row %d is exempt but carries %d problems and value %q", found.Row, len(found.Problems), found.Value.String())
			}
			continue
		}
		alone := compiled.Run(Merge(shared, rows[i]))
		requireSameAnswer(t, fmt.Sprintf("batch differs from single, row %d", i+1), found, alone)
	}
}

func requireSummaryMatches(t *testing.T, result Result) {
	t.Helper()
	valued, stopped, notRun := 0, 0, 0
	byCode := map[string]int{}
	for _, found := range result.Rows {
		switch {
		case found.StoppedAt == StageBatch:
			notRun++
		case len(found.Problems) > 0:
			stopped++
		default:
			valued++
		}
		for _, one := range found.Problems {
			byCode[one.Code]++
		}
	}
	summary := result.Summary
	if summary.Rows != len(result.Rows) || summary.Valued != valued || summary.Stopped != stopped || summary.NotRun != notRun {
		t.Fatalf("summary differs from results: summary %+v, results show valued %d stopped %d not run %d of %d",
			summary, valued, stopped, notRun, len(result.Rows))
	}
	if summary.Valued+summary.Stopped+summary.NotRun != summary.Rows {
		t.Fatalf("summary counts %d + %d + %d do not add up to %d", summary.Valued, summary.Stopped, summary.NotRun, summary.Rows)
	}
	if !reflect.DeepEqual(summary.ProblemsByCode, byCode) {
		t.Fatalf("problems by code differ: summary %v, results %v", summary.ProblemsByCode, byCode)
	}
}

func answersOf(result Result) []string {
	answers := make([]string, len(result.Rows))
	for i, found := range result.Rows {
		if len(found.Problems) > 0 {
			answers[i] = "problems:" + fmt.Sprint(codesOf(found.Problems)) + "@" + string(found.StoppedAt)
			continue
		}
		answers[i] = found.Value.String()
	}
	return answers
}

func listOf(items ...string) numtext.Received {
	return numtext.Many(items...)
}

func defaultOptions() formula.Options {
	return formula.DefaultOptions()
}
