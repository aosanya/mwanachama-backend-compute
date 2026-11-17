package bulkruns

import (
	"context"
	"errors"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func requireRefused(t *testing.T, result batch.Result, err error, code string, limit int, got int) {
	t.Helper()
	var refused *batch.BatchError
	require(t, errors.As(err, &refused), "the call gave error %v, want a BatchError with code %s", err, code)
	require(t, refused.Code == code && refused.Limit == limit && refused.Got == got,
		"refused with %+v, want code %s limit %d got %d", *refused, code, limit, got)
	require(t, refused.Error() != "", "the refusal has no sentence")
	require(t, len(result.Rows) == 0, "a refused call returned %d rows", len(result.Rows))
}

func TestBR60MoreRowsThanTheLimitAreRefusedWholeBeforeAnyRowIsWorked(t *testing.T) {
	hook := &probe{}
	compiled := compile(t, formulaM(), bodyProbed, hook, formula.DefaultOptions())
	result, err := compiled.RunBatch(context.Background(), ruleR1Rows(10001), batch.Options{Workers: 4})
	requireRefused(t, result, err, batch.CodeTooManyRows, 10000, 10001)
	require(t, hook.started.Load() == 0, "%d rows were worked on a refused call", hook.started.Load())
}

func TestBR61AChunkSizeAboveItsMaximumIsRefused(t *testing.T) {
	hook := &probe{}
	compiled := compile(t, formulaM(), bodyProbed, hook, formula.DefaultOptions())
	result, err := compiled.RunBatch(context.Background(), ruleR1Rows(3), batch.Options{ChunkSize: 10001})
	requireRefused(t, result, err, batch.CodeBadChunkSize, 10000, 10001)
	require(t, hook.started.Load() == 0, "%d rows were worked on a refused call", hook.started.Load())
}

func TestBR62WorkersAboveTheirMaximumAreRefused(t *testing.T) {
	hook := &probe{}
	compiled := compile(t, formulaM(), bodyProbed, hook, formula.DefaultOptions())
	result, err := compiled.RunBatch(context.Background(), ruleR1Rows(3), batch.Options{Workers: 65})
	requireRefused(t, result, err, batch.CodeBadWorkers, 64, 65)
	require(t, hook.started.Load() == 0, "%d rows were worked on a refused call", hook.started.Load())
}

func TestBR63ValuesBelowTheRangeAreRefusedWithTheUpperLimitAndTheNumberReceived(t *testing.T) {
	compiled := compileM(t)
	chunk, err := compiled.RunBatch(context.Background(), ruleR1Rows(3), batch.Options{ChunkSize: -1})
	requireRefused(t, chunk, err, batch.CodeBadChunkSize, 10000, -1)
	workers, err := compiled.RunBatch(context.Background(), ruleR1Rows(3), batch.Options{Workers: -1})
	requireRefused(t, workers, err, batch.CodeBadWorkers, 64, -1)
}

func TestBR64RowsAreTestedBeforeWorkersWhenBothAreWrong(t *testing.T) {
	compiled := compileM(t)
	result, err := compiled.RunBatch(context.Background(), ruleR1Rows(10001), batch.Options{Workers: 65})
	requireRefused(t, result, err, batch.CodeTooManyRows, 10000, 10001)
}

func TestTheRowLimitIsTheAgencysSettingAndABatchOfExactlyTheLimitIsAccepted(t *testing.T) {
	compiled := compileM(t)
	over, err := compiled.RunBatch(context.Background(), ruleR1Rows(51), batch.Options{MaxRows: 50})
	requireRefused(t, over, err, batch.CodeTooManyRows, 50, 51)
	runAndCheck(t, compiled, ruleR1Rows(50), nil, batch.Options{MaxRows: 50}, ruleR1Answers(50))
	runAndCheck(t, compiled, ruleR1Rows(10000), nil, batch.Options{}, ruleR1Answers(10000))
}

func rowCostOf(t *testing.T, balance string) int {
	t.Helper()
	tree, problems := formula.Parse(bodyM)
	require(t, len(problems) == 0, "the formula did not parse: %v", problems)
	inputs := []formula.Input{{Name: "balance", Type: formula.TypeDecimal}, {Name: "rate", Type: formula.TypeDecimal}}
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	require(t, len(problems) == 0, "the formula did not plan: %v", problems)
	meter := formula.NewMeter(formula.DefaultOptions().Limits.Cost)
	_, failure := program.RunMetered(map[string]formula.Value{
		"balance": formula.DecimalValue(decimalOf(t, balance)),
		"rate":    formula.DecimalValue(decimalOf(t, "0.005")),
	}, meter)
	require(t, failure == nil, "the single run failed: %v", failure)
	require(t, meter.Spent() > 0, "a row spent %d", meter.Spent())
	return meter.Spent()
}

func costOfOneRow(t *testing.T) int {
	t.Helper()
	first := rowCostOf(t, "10.00")
	last := rowCostOf(t, "4500.00")
	require(t, first == last, "a row costs %d at row 1 and %d at row 450, so 250 x c is not a clean limit", first, last)
	return first
}

func costLimitedAnswers(valued int) []answer {
	want := make([]answer, 450)
	for i := range want {
		if i < valued {
			want[i] = value(ruleR1Answer(i + 1))
		} else {
			want[i] = notRun(batch.CodeBatchLimitReached)
		}
	}
	return want
}

func TestBR65ToBR68TheBatchCostLimitStartsAChunkOnlyWhileTheCostIsBelowIt(t *testing.T) {
	compiled := compileM(t)
	c := costOfOneRow(t)
	rows := ruleR1Rows(450)
	cases := []struct {
		id     string
		chunk  int
		valued int
		sum    string
	}{
		{"BR-65", 100, 300, "2257.50"},
		{"BR-66", 1, 250, "1568.75"},
		{"BR-67", 450, 450, "5073.75"},
		{"BR-68", 50, 250, "1568.75"},
	}
	for _, k := range cases {
		t.Run(k.id, func(t *testing.T) {
			result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: k.chunk, Cost: 250 * c}, costLimitedAnswers(k.valued))
			requireSummaryCounts(t, result, k.valued, 0, 450-k.valued)
			requireSum(t, result.Rows, 0, 450, k.sum)
			if k.valued < 450 {
				requireCodeCounts(t, result, map[string]int{batch.CodeBatchLimitReached: 450 - k.valued})
			}
		})
	}
}

func TestBR69ABadRowInsideTheDoneRowsAndOneInTheRowsLeft(t *testing.T) {
	compiled := compileM(t)
	c := costOfOneRow(t)
	rows := ruleR1Rows(450)
	rows[119] = row("balance", "-1200.00", "rate", "0.005")
	rows[349] = row("balance", "-3500.00", "rate", "0.005")
	want := costLimitedAnswers(300)
	want[119] = stopped(batch.StageValidate, bad("balance", "-1200.00", problem.CodeNoNegatives))
	want[349] = notRun(batch.CodeBatchLimitReached)
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100, Cost: 250 * c}, want)
	requireSummaryCounts(t, result, 299, 1, 150)
	requireSum(t, result.Rows, 0, 450, "2251.50")
}

func TestBR70ACostLimitWithFourWorkersLeavesEveryRowEitherEqualToItsSingleRunOrNotStarted(t *testing.T) {
	compiled := compileM(t)
	c := costOfOneRow(t)
	rows := ruleR1Rows(450)
	for repeat := 0; repeat < 5; repeat++ {
		result, err := compiled.RunBatch(context.Background(), rows, batch.Options{ChunkSize: 10, Workers: 4, Cost: 250 * c})
		requireNoError(t, err)
		requireShape(t, result, 450)
		requireSummary(t, result)
		require(t, result.Summary.Valued+result.Summary.Stopped+result.Summary.NotRun == 450, "summary %+v does not add up to 450", result.Summary)
		for _, found := range result.Rows {
			hasValue := found.Value.String() != ""
			require(t, hasValue != (len(found.Problems) > 0), "row %d holds a value and problems, or neither", found.Row)
			if found.StoppedAt == batch.StageBatch {
				require(t, found.Problems[0].Code == batch.CodeBatchLimitReached, "row %d has code %s", found.Row, found.Problems[0].Code)
			}
		}
		requireEqualsSingle(t, compiled, rows, nil, result)
	}
}

func hostileDefinition() function.Definition {
	return listDefinition("amounts")
}

func listRow(items ...string) function.Row {
	return function.Row{"amounts": manyOf(items...)}
}

func sameItems(count int, text string) []string {
	items := make([]string, count)
	for i := range items {
		items[i] = text
	}
	return items
}

func TestBR71B01OneHostileRowAmongOneHundredAtPositionFiftySeven(t *testing.T) {
	compiled := compile(t, hostileDefinition(), "sum(amounts)", nil, formula.DefaultOptions())
	rows := make([]function.Row, 100)
	want := make([]answer, 100)
	for i := range rows {
		rows[i] = listRow("1.00", "2.00")
		want[i] = value("3.00")
	}
	rows[56] = listRow(sameItems(1001, "1.00")...)
	want[56] = stopped(batch.StageValidate, expectedProblem{code: problem.CodeListTooLong})
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 10, Workers: 4}, want)
	requireSummaryCounts(t, result, 99, 1, 0)
}

func TestBR71B03TheHostileRowIsTheFirstRow(t *testing.T) {
	compiled := compile(t, hostileDefinition(), "sum(amounts)", nil, formula.DefaultOptions())
	rows := []function.Row{listRow(sameItems(1001, "1.00")...), listRow("5.00"), listRow("6.00")}
	want := []answer{stopped(batch.StageValidate, expectedProblem{code: problem.CodeListTooLong}), value("5.00"), value("6.00")}
	runAndCheck(t, compiled, rows, nil, batch.Options{}, want)
}

func TestBR71B02EveryRowHostileGivesEachItsOwnRefusalInItsOwnPosition(t *testing.T) {
	compiled := compile(t, hostileDefinition(), "sum(amounts)", nil, formula.DefaultOptions())
	rows := []function.Row{listRow(sameItems(1001, "1.00")...), listRow(sameItems(2000, "1.00")...), listRow(sameItems(1500, "2.00")...)}
	want := make([]answer, 3)
	for i := range want {
		want[i] = stopped(batch.StageValidate, expectedProblem{code: problem.CodeListTooLong})
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{Workers: 2}, want)
	requireSummaryCounts(t, result, 0, 3, 0)
}

func costLimitedFormula(t *testing.T) *batch.Compiled {
	t.Helper()
	options := formula.DefaultOptions()
	options.Limits.Cost = 500
	return compile(t, hostileDefinition(), "sum(amounts)", nil, options)
}

func requireCostSentence(t *testing.T, found batch.RowResult) {
	t.Helper()
	require(t, found.Problems[0].Sentence == "This formula goes over the cost limit of 500.", "row %d sentence %q", found.Row, found.Problems[0].Sentence)
}

func TestBR71B04OneRowOverTheCostLimitAtRunTimeInTheMiddle(t *testing.T) {
	compiled := costLimitedFormula(t)
	rows := []function.Row{listRow("1.00", "2.00"), listRow("1.00", "2.00"), listRow(sameItems(1000, "1.00")...), listRow("1.00", "2.00"), listRow("1.00", "2.00")}
	want := []answer{value("3.00"), value("3.00"), stopped(batch.StageCalculate, expectedProblem{code: "too-costly"}), value("3.00"), value("3.00")}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 2, Workers: 2}, want)
	requireCostSentence(t, result.Rows[2])
}

func TestBR71B05TwoRowsOverTheCostLimitOneAfterTheOther(t *testing.T) {
	compiled := costLimitedFormula(t)
	rows := []function.Row{listRow("1.00"), listRow(sameItems(1000, "1.00")...), listRow(sameItems(1000, "2.00")...), listRow("4.00")}
	want := []answer{value("1.00"), stopped(batch.StageCalculate, expectedProblem{code: "too-costly"}), stopped(batch.StageCalculate, expectedProblem{code: "too-costly"}), value("4.00")}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 1, Workers: 3}, want)
	requireCostSentence(t, result.Rows[1])
	requireCostSentence(t, result.Rows[2])
	requireSummaryCounts(t, result, 2, 2, 0)
}
