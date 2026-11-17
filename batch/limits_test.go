package batch

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func requireRefused(t *testing.T, result Result, err error, code string, limit int, got int) {
	t.Helper()
	var refused *BatchError
	if !errors.As(err, &refused) {
		t.Fatalf("the call gave error %v, want a BatchError with code %s", err, code)
	}
	if refused.Code != code || refused.Limit != limit || refused.Got != got {
		t.Fatalf("refused with %+v, want code %s limit %d got %d", *refused, code, limit, got)
	}
	if refused.Error() == "" {
		t.Fatal("the refusal has no sentence")
	}
	if len(result.Rows) != 0 {
		t.Fatalf("a refused call returned %d rows", len(result.Rows))
	}
}

func TestMoreRowsThanTheLimitAreRefusedWholeBeforeAnyRowIsWorked(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	touched := 0
	options := Options{beforeChunk: func(int) { touched++ }}
	result, err := compiled.RunBatch(context.Background(), monthEndRows(10001), options)
	requireRefused(t, result, err, CodeTooManyRows, 10000, 10001)
	if touched != 0 {
		t.Fatalf("%d chunks were started on a refused call", touched)
	}
}

func TestTheRowLimitIsTheAgencysSettingAndNotAConstant(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(51), Options{MaxRows: 50})
	requireRefused(t, result, err, CodeTooManyRows, 50, 51)
	exact, err := compiled.RunBatch(context.Background(), monthEndRows(50), Options{MaxRows: 50})
	requireNoError(t, err)
	requireRowNumbers(t, exact, 50)
	larger, err := compiled.RunBatch(context.Background(), monthEndRows(10001), Options{MaxRows: 20000})
	requireNoError(t, err)
	requireRowNumbers(t, larger, 10001)
}

func TestAChunkSizeOutsideItsRangeIsRefused(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	cases := []struct {
		name      string
		options   Options
		limit     int
		got       int
		wantsCode string
	}{
		{"over the maximum", Options{ChunkSize: 10001}, 10000, 10001, CodeBadChunkSize},
		{"negative", Options{ChunkSize: -1}, 10000, -1, CodeBadChunkSize},
		{"over the agency's rows", Options{ChunkSize: 51, MaxRows: 50}, 50, 51, CodeBadChunkSize},
	}
	for _, c := range cases {
		result, err := compiled.RunBatch(context.Background(), monthEndRows(3), c.options)
		t.Run(c.name, func(t *testing.T) { requireRefused(t, result, err, c.wantsCode, c.limit, c.got) })
	}
}

func TestWorkersOutsideTheirRangeAreRefused(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	over, err := compiled.RunBatch(context.Background(), monthEndRows(3), Options{Workers: 65})
	requireRefused(t, over, err, CodeBadWorkers, 64, 65)
	negative, err := compiled.RunBatch(context.Background(), monthEndRows(3), Options{Workers: -1})
	requireRefused(t, negative, err, CodeBadWorkers, 64, -1)
	most, err := compiled.RunBatch(context.Background(), monthEndRows(3), Options{Workers: 64})
	requireNoError(t, err)
	requireRowNumbers(t, most, 3)
}

func TestTooManyRowsIsTheErrorWhenSeveralLimitsAreBroken(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(10001), Options{Workers: 65})
	requireRefused(t, result, err, CodeTooManyRows, 10000, 10001)
}

func rowCost(t *testing.T, compiled *Compiled) int {
	t.Helper()
	_, spent := compiled.runRow(monthEndRows(1)[0])
	if spent <= 0 {
		t.Fatalf("a row spent %d", spent)
	}
	return spent
}

func sumOfAnswers(rows []RowResult, from int, to int) int {
	total := 0
	for i := from; i < to; i++ {
		total += 5 * (i + 1)
	}
	return total
}

func requireValuedThenNotRun(t *testing.T, result Result, valued int, total int) {
	t.Helper()
	for i, found := range result.Rows {
		if i < valued {
			requireValue(t, found, monthEndAnswer(i+1))
			continue
		}
		requireProblems(t, found, StageBatch, CodeBatchLimitReached)
	}
	if result.Summary.Valued != valued || result.Summary.NotRun != total-valued || result.Summary.Stopped != 0 {
		t.Fatalf("summary %+v, want valued %d not run %d", result.Summary, valued, total-valued)
	}
	if result.Summary.ProblemsByCode[CodeBatchLimitReached] != total-valued {
		t.Fatalf("problems by code %v", result.Summary.ProblemsByCode)
	}
}

func TestAChunkIsStartedOnlyWhileTheBatchCostIsBelowTheLimit(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	c := rowCost(t, compiled)
	rows := monthEndRows(450)
	cases := []struct {
		name   string
		chunk  int
		valued int
	}{
		{"chunks of 100", 100, 300},
		{"chunks of 1", 1, 250},
		{"one chunk of 450", 450, 450},
		{"chunks of 50 reach the limit exactly", 50, 250},
	}
	for _, k := range cases {
		t.Run(k.name, func(t *testing.T) {
			result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: k.chunk, Cost: 250 * c})
			requireNoError(t, err)
			requireRowNumbers(t, result, 450)
			requireValuedThenNotRun(t, result, k.valued, 450)
			requireEqualsSingle(t, compiled, rows, nil, result)
			requireSummaryMatches(t, result)
		})
	}
}

func TestTheBatchCostLimitKeepsEveryStartedRowsAnswer(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	c := rowCost(t, compiled)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(450), Options{ChunkSize: 100, Cost: 250 * c})
	requireNoError(t, err)
	if got := sumOfAnswers(result.Rows, 0, 300); got != 225750 {
		t.Fatalf("the test's own sum %d, want 225750", got)
	}
}

func TestARowStoppedAtValidateDoesNotSpendTheBatchLimit(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	c := rowCost(t, compiled)
	rows := monthEndRows(450)
	rows[119] = row("balance", "-1200.00", "rate", "0.005")
	rows[349] = row("balance", "-3500.00", "rate", "0.005")
	result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 100, Cost: 250 * c})
	requireNoError(t, err)
	requireProblems(t, result.Rows[119], StageValidate, problem.CodeNoNegatives)
	requireProblems(t, result.Rows[349], StageBatch, CodeBatchLimitReached)
	if result.Summary.Valued != 299 || result.Summary.Stopped != 1 || result.Summary.NotRun != 150 {
		t.Fatalf("summary %+v, want valued 299 stopped 1 not run 150", result.Summary)
	}
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestWithSeveralWorkersTheLimitLeavesRowsThatEitherEqualSingleOrWereNotRun(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	c := rowCost(t, compiled)
	rows := monthEndRows(450)
	result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 10, Workers: 4, Cost: 250 * c})
	requireNoError(t, err)
	requireRowNumbers(t, result, 450)
	requireSummaryMatches(t, result)
	if result.Summary.Valued+result.Summary.Stopped+result.Summary.NotRun != 450 {
		t.Fatalf("summary %+v does not add up to 450", result.Summary)
	}
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestWithNoBatchCostLimitEveryRowIsWorked(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(450), Options{ChunkSize: 10})
	requireNoError(t, err)
	if result.Summary.Valued != 450 {
		t.Fatalf("summary %+v", result.Summary)
	}
}

func TestACallCancelledBeforeItStartsLeavesEveryRowNotRunAndNoError(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	rows := monthEndRows(450)
	result, err := compiled.RunBatch(cancelled, rows, Options{ChunkSize: 100})
	requireNoError(t, err)
	requireRowNumbers(t, result, 450)
	for _, found := range result.Rows {
		requireProblems(t, found, StageBatch, CodeBatchCancelled)
	}
	if result.Summary.Valued != 0 || result.Summary.NotRun != 450 {
		t.Fatalf("summary %+v", result.Summary)
	}
	again, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 100})
	requireNoError(t, err)
	if again.Summary.Valued != 450 {
		t.Fatalf("the second call %+v", again.Summary)
	}
	requireValue(t, again.Rows[449], "22.50")
}

func TestACancelAfterTheSecondChunkLeavesTheRestNotRun(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rows := monthEndRows(450)
	options := Options{ChunkSize: 100, afterChunk: func(index int) {
		if index == 2 {
			cancel()
		}
	}}
	result, err := compiled.RunBatch(ctx, rows, options)
	requireNoError(t, err)
	for i, found := range result.Rows {
		if i < 200 {
			requireValue(t, found, monthEndAnswer(i+1))
			continue
		}
		requireProblems(t, found, StageBatch, CodeBatchCancelled)
	}
	if result.Summary.Valued != 200 || result.Summary.NotRun != 250 {
		t.Fatalf("summary %+v", result.Summary)
	}
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestTheRowsLeftAfterACancelCanBeRunAgainAndJoined(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rows := monthEndRows(450)
	rows[349] = row("balance", "-3500.00", "rate", "0.005")
	first, err := compiled.RunBatch(ctx, rows, Options{ChunkSize: 100, afterChunk: func(index int) {
		if index == 2 {
			cancel()
		}
	}})
	requireNoError(t, err)
	requireProblems(t, first.Rows[349], StageBatch, CodeBatchCancelled)
	var leftPositions []int
	var left []function.Row
	for i, found := range first.Rows {
		if found.StoppedAt == StageBatch {
			leftPositions = append(leftPositions, i)
			left = append(left, rows[i])
		}
	}
	second, err := compiled.RunBatch(context.Background(), left, Options{ChunkSize: 100})
	requireNoError(t, err)
	joined := append([]RowResult{}, first.Rows...)
	seen := map[int]bool{}
	for j, found := range second.Rows {
		position := leftPositions[j]
		if seen[position] {
			t.Fatalf("row %d appears twice", position+1)
		}
		seen[position] = true
		found.Row = position + 1
		joined[position] = found
	}
	whole, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 100})
	requireNoError(t, err)
	if !reflect.DeepEqual(answersOf(Result{Rows: joined}), answersOf(whole)) {
		t.Fatal("the joined rows differ from a run that was never cancelled")
	}
	requireProblems(t, joined[349], StageValidate, problem.CodeNoNegatives)
}

func TestWithSeveralWorkersACancelLeavesEveryRowEitherAnsweredOrNotRun(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rows := monthEndRows(450)
	result, err := compiled.RunBatch(ctx, rows, Options{ChunkSize: 10, Workers: 4, afterChunk: func(index int) {
		if index == 5 {
			cancel()
		}
	}})
	requireNoError(t, err)
	requireRowNumbers(t, result, 450)
	for _, found := range result.Rows {
		if found.StoppedAt == StageBatch {
			requireProblems(t, found, StageBatch, CodeBatchCancelled)
			continue
		}
		requireValue(t, found, monthEndAnswer(found.Row))
	}
	requireSummaryMatches(t, result)
}

func TestACancelAndALimitTogetherCarryTheCancelCode(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	result, err := compiled.RunBatch(cancelled, monthEndRows(5), Options{Cost: 1})
	requireNoError(t, err)
	for _, found := range result.Rows {
		requireProblems(t, found, StageBatch, CodeBatchCancelled)
	}
}
