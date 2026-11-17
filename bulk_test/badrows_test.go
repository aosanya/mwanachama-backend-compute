package bulkruns

import (
	"context"
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func TestBR20BadRowsAtTheEdgesOfChunksNeverTouchTheRowsAroundThem(t *testing.T) {
	compiled := compileM(t)
	rows := []function.Row{
		row("balance", "-10.00", "rate", "0.005"),
		row("balance", "20.00", "rate", "0.005"),
		row("balance", "30.00", "rate", "0.005"),
		row("balance", "40.005", "rate", "0.005"),
		row("rate", "0.005"),
		row("balance", "60.00", "rate", "0.005"),
		row("balance", "70.00", "rate", "0.005"),
		row("balance", "80.00", "rate", "0.005"),
		row("balance", "90.00", "rate", "0.005"),
		row("balance", "-100.00", "rate", "0.00505"),
	}
	want := []answer{
		stopped(batch.StageValidate, bad("balance", "-10.00", problem.CodeNoNegatives)),
		value("0.10"),
		value("0.15"),
		stopped(batch.StageValidate, bad("balance", "40.005", problem.CodeDecimalPlaces)),
		stopped(batch.StageValidate, missing("balance", problem.CodeRequired)),
		value("0.30"),
		value("0.35"),
		value("0.40"),
		value("0.45"),
		stopped(batch.StageValidate, bad("balance", "-100.00", problem.CodeNoNegatives), bad("rate", "0.00505", problem.CodeDecimalPlaces)),
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 4}, want)
	requireSummaryCounts(t, result, 6, 4, 0)
	requireCodeCounts(t, result, map[string]int{problem.CodeNoNegatives: 2, problem.CodeDecimalPlaces: 2, problem.CodeRequired: 1})
}

func TestBR21AOneBadRowAtEveryPositionOfTenRowsWhateverTheChunkSize(t *testing.T) {
	compiled := compileM(t)
	for position := 1; position <= 10; position++ {
		rows := ruleR1Rows(10)
		negative := "-" + cents(1000*position)
		rows[position-1] = row("balance", negative, "rate", "0.005")
		want := ruleR1Answers(10)
		want[position-1] = stopped(batch.StageValidate, bad("balance", negative, problem.CodeNoNegatives))
		for _, chunk := range []int{4, 1, 10, 11} {
			t.Run(fmt.Sprintf("position %d chunk %d", position, chunk), func(t *testing.T) {
				result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: chunk}, want)
				requireSummaryCounts(t, result, 9, 1, 0)
			})
		}
	}
}

func TestBR22OneBadRowAmongAHundred(t *testing.T) {
	compiled := compileM(t)
	rows := ruleR1Rows(100)
	rows[56] = row("balance", "-570.00", "rate", "0.005")
	want := ruleR1Answers(100)
	want[56] = stopped(batch.StageValidate, bad("balance", "-570.00", problem.CodeNoNegatives))
	result := runAndCheck(t, compiled, rows, nil, batch.Options{}, want)
	requireSummaryCounts(t, result, 99, 1, 0)
}

func TestBR23BadRowsOnBothSidesOfEveryChunkEdgeOfTenThousandRows(t *testing.T) {
	compiled := compileM(t)
	rows := ruleR1Rows(10000)
	want := ruleR1Answers(10000)
	badRows := 0
	for i := 1; i <= 10000; i++ {
		if i%100 == 0 || i%100 == 1 {
			negative := "-" + cents(1000*i)
			rows[i-1] = row("balance", negative, "rate", "0.005")
			want[i-1] = stopped(batch.StageValidate, bad("balance", negative, problem.CodeNoNegatives))
			badRows++
		}
	}
	require(t, badRows == 200, "the test built %d bad rows, want 200", badRows)
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100}, want)
	requireSummaryCounts(t, result, 9800, 200, 0)
	requireSum(t, result.Rows, 0, 10000, "2450245.00")
	requireCodeCounts(t, result, map[string]int{problem.CodeNoNegatives: 200})
}

func TestBR24EveryRowBadAndTheCallStillReturnsNoError(t *testing.T) {
	compiled := compileM(t)
	rows := []function.Row{
		row("rate", "0.005"),
		row("balance", "-5.00", "rate", "0.005"),
		row("balance", "12.345", "rate", "0.005"),
		row("balance", "-5.125", "rate", "-0.005"),
		{},
	}
	want := []answer{
		stopped(batch.StageValidate, missing("balance", problem.CodeRequired)),
		stopped(batch.StageValidate, bad("balance", "-5.00", problem.CodeNoNegatives)),
		stopped(batch.StageValidate, bad("balance", "12.345", problem.CodeDecimalPlaces)),
		stopped(batch.StageValidate,
			bad("balance", "-5.125", problem.CodeNoNegatives),
			bad("balance", "-5.125", problem.CodeDecimalPlaces),
			bad("rate", "-0.005", problem.CodeNoNegatives)),
		stopped(batch.StageValidate, missing("balance", problem.CodeRequired), missing("rate", problem.CodeRequired)),
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{}, want)
	requireSummaryCounts(t, result, 0, 5, 0)
	requireCodeCounts(t, result, map[string]int{problem.CodeRequired: 3, problem.CodeNoNegatives: 3, problem.CodeDecimalPlaces: 2})
	total := 0
	for _, count := range result.Summary.ProblemsByCode {
		total += count
	}
	require(t, total == 8, "%d problems in all, want 8", total)
}

func TestBR25EveryRowBadAmong250(t *testing.T) {
	compiled := compileM(t)
	rows := make([]function.Row, 250)
	want := make([]answer, 250)
	for i := range rows {
		rows[i] = row("rate", "0.005")
		want[i] = stopped(batch.StageValidate, missing("balance", problem.CodeRequired))
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100}, want)
	requireSummaryCounts(t, result, 0, 250, 0)
	requireCodeCounts(t, result, map[string]int{problem.CodeRequired: 250})
}

func TestBR30ARowStoppedAtValidateAndARowStoppedAtCalculateInOneBatch(t *testing.T) {
	compiled := compileS(t)
	rows := []function.Row{
		row("total", "100.00", "people", "3"),
		row("total", "10.00", "people", "0"),
		row("total", "-5.00", "people", "2"),
		row("total", "90.00", "people", "4"),
		row("total", "20.00", "people", "0"),
		row("total", "7.005", "people", "1"),
		row("people", "0"),
		row("total", "50.00", "people", "2"),
	}
	want := []answer{
		value("33.33"),
		stopped(batch.StageCalculate, expectedProblem{code: "divide-by-zero"}),
		stopped(batch.StageValidate, bad("total", "-5.00", problem.CodeNoNegatives)),
		value("22.50"),
		stopped(batch.StageCalculate, expectedProblem{code: "divide-by-zero"}),
		stopped(batch.StageValidate, bad("total", "7.005", problem.CodeDecimalPlaces)),
		stopped(batch.StageValidate, missing("total", problem.CodeRequired)),
		value("25.00"),
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 3}, want)
	requireSummaryCounts(t, result, 3, 5, 0)
	requireCodeCounts(t, result, map[string]int{"divide-by-zero": 2, problem.CodeNoNegatives: 1, problem.CodeDecimalPlaces: 1, problem.CodeRequired: 1})
}

func TestBR31ThreeKindsOfRowByRuleOver300Rows(t *testing.T) {
	compiled := compileS(t)
	rows := make([]function.Row, 300)
	want := make([]answer, 300)
	for i := 1; i <= 300; i++ {
		switch i % 3 {
		case 1:
			negative := "-" + cents(1000*i)
			rows[i-1] = row("total", negative, "people", "2")
			want[i-1] = stopped(batch.StageValidate, bad("total", negative, problem.CodeNoNegatives))
		case 2:
			rows[i-1] = row("total", cents(1000*i), "people", "2")
			want[i-1] = value(cents(500 * i))
		default:
			rows[i-1] = row("total", cents(1000*i), "people", "0")
			want[i-1] = stopped(batch.StageCalculate, expectedProblem{code: "divide-by-zero"})
		}
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100}, want)
	requireSummaryCounts(t, result, 100, 200, 0)
	requireSum(t, result.Rows, 0, 300, "75250.00")
	atValidate, atCalculate := 0, 0
	for _, found := range result.Rows {
		switch found.StoppedAt {
		case batch.StageValidate:
			atValidate++
		case batch.StageCalculate:
			atCalculate++
		}
	}
	require(t, atValidate == 100 && atCalculate == 100, "stopped at validate %d and at calculate %d, want 100 and 100", atValidate, atCalculate)
}

func TestBR40TheSameRowTwiceGivesTheSameAnswerTwice(t *testing.T) {
	compiled := compileM(t)
	rows := []function.Row{
		row("balance", "120.50", "rate", "0.005"),
		row("balance", "120.50", "rate", "0.005"),
		row("balance", "-4.00", "rate", "0.005"),
		row("balance", "-4.00", "rate", "0.005"),
		row("balance", "120.50", "rate", "0.005"),
	}
	want := []answer{
		value("0.60"), value("0.60"),
		stopped(batch.StageValidate, bad("balance", "-4.00", problem.CodeNoNegatives)),
		stopped(batch.StageValidate, bad("balance", "-4.00", problem.CodeNoNegatives)),
		value("0.60"),
	}
	runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 2, Workers: 2}, want)
}

func TestBR40AFailureAtCalculateIsTheSameOnBothDuplicates(t *testing.T) {
	compiled := compileS(t)
	rows := []function.Row{row("total", "10.00", "people", "0"), row("total", "10.00", "people", "0")}
	want := []answer{
		stopped(batch.StageCalculate, expectedProblem{code: "divide-by-zero"}),
		stopped(batch.StageCalculate, expectedProblem{code: "divide-by-zero"}),
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{}, want)
	require(t, result.Rows[0].Problems[0].Sentence == result.Rows[1].Problems[0].Sentence, "the two duplicates carry different sentences")
}

func TestBR41TwoHundredFiftyIdenticalRows(t *testing.T) {
	compiled := compileM(t)
	rows := make([]function.Row, 250)
	want := make([]answer, 250)
	for i := range rows {
		rows[i] = row("balance", "120.50", "rate", "0.005")
		want[i] = value("0.60")
	}
	result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100}, want)
	requireSummaryCounts(t, result, 250, 0, 0)
}

func TestARowThatIsInTheBatchAndRunAloneGiveTheSameAnswerForEveryChunkSizeAndWorkerCount(t *testing.T) {
	compiled := compileS(t)
	rows := make([]function.Row, 120)
	for i := range rows {
		switch i % 4 {
		case 0:
			rows[i] = row("total", cents(1000*(i+1)), "people", "3")
		case 1:
			rows[i] = row("total", cents(1000*(i+1)), "people", "0")
		case 2:
			rows[i] = row("total", "-1.00", "people", "2")
		default:
			rows[i] = row("total", "1.005", "people", "2")
		}
	}
	reference, err := compiled.RunBatch(context.Background(), rows, batch.Options{ChunkSize: 100, Workers: 1})
	requireNoError(t, err)
	requireShape(t, reference, len(rows))
	requireEqualsSingle(t, compiled, rows, nil, reference)
	for _, chunk := range []int{1, 7, 120, 121} {
		for _, workers := range []int{1, 4, 64} {
			result, err := compiled.RunBatch(context.Background(), rows, batch.Options{ChunkSize: chunk, Workers: workers})
			requireNoError(t, err)
			requireShape(t, result, len(rows))
			for i := range rows {
				requireSameRow(t, fmt.Sprintf("chunk %d workers %d row %d", chunk, workers, i+1), result.Rows[i], reference.Rows[i])
			}
		}
	}
}
