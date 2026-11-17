package bulkruns

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type expectedProblem struct {
	input    string
	given    bool
	received string
	code     string
}

type answer struct {
	value    string
	stage    batch.Stage
	problems []expectedProblem
}

func value(text string) answer {
	return answer{value: text}
}

func stopped(stage batch.Stage, problems ...expectedProblem) answer {
	return answer{stage: stage, problems: problems}
}

func bad(input string, received string, code string) expectedProblem {
	return expectedProblem{input: input, given: true, received: received, code: code}
}

func missing(input string, code string) expectedProblem {
	return expectedProblem{input: input, code: code}
}

func notRun(code string) answer {
	return stopped(batch.StageBatch, expectedProblem{code: code})
}

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

func formulaM() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2), decimalInput("rate", 4)}}
}

const bodyM = "round(balance * rate, 2)"

func formulaS() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("total", 2), integerInput("people")}}
}

const bodyS = "round(total / people, 2)"

func formulaF() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{
		decimalInput("contribution", 2), decimalInput("payout", 2), decimalInput("total_contributions", 2),
	}}
}

const bodyF = "round(payout * contribution / total_contributions, 2)"

func compile(t *testing.T, definition function.Definition, body string, runtime formula.Runtime, options formula.Options) *batch.Compiled {
	t.Helper()
	compiled, problems := batch.Compile(definition, body, runtime, options)
	if len(problems) > 0 {
		t.Fatalf("Compile(%q) refused: %v", body, problems)
	}
	return compiled
}

func compileM(t *testing.T) *batch.Compiled {
	t.Helper()
	return compile(t, formulaM(), bodyM, nil, formula.DefaultOptions())
}

func compileS(t *testing.T) *batch.Compiled {
	t.Helper()
	return compile(t, formulaS(), bodyS, nil, formula.DefaultOptions())
}

func compileF(t *testing.T, options formula.Options) *batch.Compiled {
	t.Helper()
	return compile(t, formulaF(), bodyF, nil, options)
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

func ruleR1Row(position int) function.Row {
	return row("balance", cents(1000*position), "rate", "0.005")
}

func ruleR1Answer(position int) string {
	return cents(5 * position)
}

func ruleR1Rows(count int) []function.Row {
	rows := make([]function.Row, count)
	for i := range rows {
		rows[i] = ruleR1Row(i + 1)
	}
	return rows
}

func ruleR1Answers(count int) []answer {
	answers := make([]answer, count)
	for i := range answers {
		answers[i] = value(ruleR1Answer(i + 1))
	}
	return answers
}

func require(t *testing.T, ok bool, format string, args ...any) {
	t.Helper()
	if !ok {
		t.Fatalf(format, args...)
	}
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	require(t, err == nil, "the call returned an error: %v", err)
}

func requireShape(t *testing.T, result batch.Result, rows int) {
	t.Helper()
	require(t, len(result.Rows) == rows, "sent %d rows, got %d results", rows, len(result.Rows))
	for i, found := range result.Rows {
		require(t, found.Row == i+1, "result at position %d carries Row %d", i+1, found.Row)
	}
}

func requireAnswer(t *testing.T, label string, got batch.RowResult, want answer) {
	t.Helper()
	if want.stage == "" {
		require(t, len(got.Problems) == 0, "%s: row %d has problems %v, wanted the value %s", label, got.Row, codesOf(got.Problems), want.value)
		require(t, got.Value.Type() == formula.TypeDecimal, "%s: row %d value has type %s, want decimal", label, got.Row, got.Value.Type())
		require(t, got.Value.String() == want.value, "%s: row %d gave %s, want %s", label, got.Row, got.Value.String(), want.value)
		return
	}
	require(t, got.StoppedAt == want.stage, "%s: row %d stopped at %q, want %q (problems %v)", label, got.Row, got.StoppedAt, want.stage, codesOf(got.Problems))
	require(t, len(got.Problems) == len(want.problems), "%s: row %d has problems %v, want %d problems", label, got.Row, codesOf(got.Problems), len(want.problems))
	for i, wanted := range want.problems {
		found := got.Problems[i]
		require(t, found.Code == wanted.code, "%s: row %d problem %d has code %s, want %s", label, got.Row, i+1, found.Code, wanted.code)
		require(t, found.Sentence != "", "%s: row %d problem %d has no sentence", label, got.Row, i+1)
		if wanted.input != "" {
			require(t, found.Input != nil && *found.Input == wanted.input, "%s: row %d problem %d is about %v, want %s", label, got.Row, i+1, derefText(found.Input), wanted.input)
			if wanted.given {
				require(t, found.Given != nil && *found.Given, "%s: row %d problem %d says the input was not given", label, got.Row, i+1)
				require(t, found.Value != nil && *found.Value == wanted.received, "%s: row %d problem %d received %v, want %s", label, got.Row, i+1, derefText(found.Value), wanted.received)
			} else {
				require(t, found.Given != nil && !*found.Given, "%s: row %d problem %d says the input was given", label, got.Row, i+1)
			}
		}
	}
}

func derefText(text *string) string {
	if text == nil {
		return "<none>"
	}
	return *text
}

func codesOf(problems []problem.Problem) []string {
	codes := make([]string, len(problems))
	for i, found := range problems {
		codes[i] = found.Code
	}
	return codes
}

func requireAnswers(t *testing.T, result batch.Result, want []answer) {
	t.Helper()
	requireShape(t, result, len(want))
	for i, wanted := range want {
		requireAnswer(t, "case", result.Rows[i], wanted)
	}
}

func requireSummary(t *testing.T, result batch.Result) {
	t.Helper()
	valued, stoppedRows, notRunRows := 0, 0, 0
	byCode := map[string]int{}
	for _, found := range result.Rows {
		switch {
		case found.StoppedAt == batch.StageBatch:
			notRunRows++
		case len(found.Problems) > 0:
			stoppedRows++
		default:
			valued++
		}
		for _, one := range found.Problems {
			byCode[one.Code]++
		}
	}
	summary := result.Summary
	require(t, summary.Rows == len(result.Rows), "summary says %d rows, results hold %d", summary.Rows, len(result.Rows))
	require(t, summary.Valued == valued && summary.Stopped == stoppedRows && summary.NotRun == notRunRows,
		"summary valued %d stopped %d not run %d, results show %d %d %d", summary.Valued, summary.Stopped, summary.NotRun, valued, stoppedRows, notRunRows)
	require(t, summary.Valued+summary.Stopped+summary.NotRun == summary.Rows, "summary counts do not add up to %d", summary.Rows)
	require(t, reflect.DeepEqual(summary.ProblemsByCode, byCode), "problems by code: summary %v, results %v", summary.ProblemsByCode, byCode)
}

func requireSummaryCounts(t *testing.T, result batch.Result, valued int, stoppedRows int, notRunRows int) {
	t.Helper()
	summary := result.Summary
	require(t, summary.Valued == valued && summary.Stopped == stoppedRows && summary.NotRun == notRunRows,
		"summary valued %d stopped %d not run %d, want %d %d %d", summary.Valued, summary.Stopped, summary.NotRun, valued, stoppedRows, notRunRows)
}

func requireCodeCounts(t *testing.T, result batch.Result, want map[string]int) {
	t.Helper()
	got := result.Summary.ProblemsByCode
	if len(want) == 0 {
		require(t, len(got) == 0, "problems by code %v, want none", got)
		return
	}
	require(t, reflect.DeepEqual(got, want), "problems by code %v, want %v", got, want)
}

func requireSameRow(t *testing.T, label string, got batch.RowResult, alone batch.RowResult) {
	t.Helper()
	require(t, got.StoppedAt == alone.StoppedAt, "%s: stage differs, alone %q, in the batch %q", label, alone.StoppedAt, got.StoppedAt)
	require(t, len(got.Problems) == len(alone.Problems), "%s: problem count differs, alone %d, in the batch %d", label, len(alone.Problems), len(got.Problems))
	for i := range got.Problems {
		require(t, reflect.DeepEqual(got.Problems[i], alone.Problems[i]), "%s: problem %d differs, alone %+v, in the batch %+v", label, i+1, alone.Problems[i], got.Problems[i])
	}
	if len(got.Problems) == 0 {
		require(t, got.Value.Type() == alone.Value.Type() && got.Value.String() == alone.Value.String(),
			"%s: value differs, alone %s %s, in the batch %s %s", label, alone.Value.Type(), alone.Value.String(), got.Value.Type(), got.Value.String())
	}
}

func requireEqualsSingle(t *testing.T, compiled *batch.Compiled, rows []function.Row, shared function.Row, result batch.Result) {
	t.Helper()
	for i, found := range result.Rows {
		if found.StoppedAt == batch.StageBatch {
			require(t, len(found.Problems) == 1 && found.Value.String() == "",
				"row %d never started but carries %d problems and value %q", found.Row, len(found.Problems), found.Value.String())
			require(t, found.Problems[0].Code == batch.CodeBatchLimitReached || found.Problems[0].Code == batch.CodeBatchCancelled,
				"row %d never started and carries code %s", found.Row, found.Problems[0].Code)
			continue
		}
		alone := compiled.Run(batch.Merge(shared, rows[i]))
		requireSameRow(t, fmt.Sprintf("batch differs from single, row %d", i+1), found, alone)
	}
}

func runAndCheck(t *testing.T, compiled *batch.Compiled, rows []function.Row, shared function.Row, options batch.Options, want []answer) batch.Result {
	t.Helper()
	options.Shared = shared
	result, err := compiled.RunBatch(context.Background(), rows, options)
	requireNoError(t, err)
	requireAnswers(t, result, want)
	requireSummary(t, result)
	requireEqualsSingle(t, compiled, rows, shared, result)
	return result
}

func decimalOf(t *testing.T, text string) number.Decimal {
	t.Helper()
	parsed, err := number.ParseDecimal(text)
	require(t, err == nil, "%q is not a decimal: %v", text, err)
	return parsed
}

func sumOf(t *testing.T, rows []batch.RowResult, from int, to int) string {
	t.Helper()
	total := decimalOf(t, "0")
	for i := from; i < to; i++ {
		if len(rows[i].Problems) > 0 {
			continue
		}
		require(t, rows[i].Value.Type() == formula.TypeDecimal, "row %d is not a decimal", rows[i].Row)
		total = total.Add(rows[i].Value.Decimal())
	}
	return total.String()
}

func requireSum(t *testing.T, rows []batch.RowResult, from int, to int, want string) {
	t.Helper()
	got := sumOf(t, rows, from, to)
	wanted := decimalOf(t, want)
	require(t, decimalOf(t, got).Equal(wanted), "sum of answers is %s, want %s", got, want)
}

func positionText(position int) string {
	return strconv.Itoa(position)
}

func manyOf(items ...string) numtext.Received {
	return numtext.Many(items...)
}
