package bulkruns

import (
	"context"
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

func TestBR01ZeroRowsGiveNoResultsNoErrorAndASummaryOfZeros(t *testing.T) {
	compiled := compileM(t)
	result, err := compiled.RunBatch(context.Background(), nil, batch.Options{ChunkSize: 100, Workers: 1})
	requireNoError(t, err)
	require(t, result.Rows != nil && len(result.Rows) == 0, "rows %v, want an empty list", result.Rows)
	requireSummaryCounts(t, result, 0, 0, 0)
	require(t, result.Summary.Rows == 0, "summary rows %d", result.Summary.Rows)
	requireCodeCounts(t, result, nil)
}

func TestBR02ToBR09SizesAroundAChunkGiveEveryRowItsOwnAnswer(t *testing.T) {
	compiled := compileM(t)
	cases := []struct {
		id      string
		rows    int
		lastRow string
		sum     string
	}{
		{"BR-02", 1, "0.05", "0.05"},
		{"BR-03", 3, "0.15", "0.30"},
		{"BR-04", 99, "4.95", "247.50"},
		{"BR-05", 100, "5.00", "252.50"},
		{"BR-06", 101, "5.05", "257.55"},
		{"BR-07", 200, "10.00", "1005.00"},
		{"BR-08", 201, "10.05", "1015.05"},
		{"BR-09", 10000, "500.00", "2500250.00"},
	}
	for _, c := range cases {
		t.Run(c.id, func(t *testing.T) {
			result := runAndCheck(t, compiled, ruleR1Rows(c.rows), nil, batch.Options{ChunkSize: 100, Workers: 1}, ruleR1Answers(c.rows))
			requireAnswer(t, c.id, result.Rows[0], value("0.05"))
			requireAnswer(t, c.id, result.Rows[c.rows-1], value(c.lastRow))
			requireSum(t, result.Rows, 0, c.rows, c.sum)
			requireSummaryCounts(t, result, c.rows, 0, 0)
			requireCodeCounts(t, result, nil)
		})
	}
}

func TestBR03WrittenOut(t *testing.T) {
	compiled := compileM(t)
	rows := []function.Row{
		row("balance", "10.00", "rate", "0.005"),
		row("balance", "20.00", "rate", "0.005"),
		row("balance", "30.00", "rate", "0.005"),
	}
	runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: 100, Workers: 1}, []answer{value("0.05"), value("0.10"), value("0.15")})
}

func TestBR10ChunkEdgesGiveTheSameSevenAnswersWhateverTheChunkSizeAndWorkers(t *testing.T) {
	compiled := compileM(t)
	rows := ruleR1Rows(7)
	want := []answer{value("0.05"), value("0.10"), value("0.15"), value("0.20"), value("0.25"), value("0.30"), value("0.35")}
	for _, chunk := range []int{3, 1, 2, 7, 8} {
		for _, workers := range []int{1, 4} {
			t.Run(fmt.Sprintf("chunk %d workers %d", chunk, workers), func(t *testing.T) {
				result := runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: chunk, Workers: workers}, want)
				require(t, result.Summary.ChunkSize == chunk && result.Summary.Workers == workers,
					"summary says chunk size %d workers %d, used %d and %d", result.Summary.ChunkSize, result.Summary.Workers, chunk, workers)
			})
		}
	}
}

func TestBR11ChunkSweepOver250RowsNeverChangesAnAnswer(t *testing.T) {
	compiled := compileM(t)
	rows := ruleR1Rows(250)
	want := ruleR1Answers(250)
	for _, chunk := range []int{100, 1, 7, 250, 251} {
		for _, workers := range []int{1, 4, 64} {
			t.Run(fmt.Sprintf("chunk %d workers %d", chunk, workers), func(t *testing.T) {
				runAndCheck(t, compiled, rows, nil, batch.Options{ChunkSize: chunk, Workers: workers}, want)
			})
		}
	}
}

func TestDefaultsAreUsedForZeroAndReportedInTheSummary(t *testing.T) {
	compiled := compileM(t)
	result := runAndCheck(t, compiled, ruleR1Rows(250), nil, batch.Options{}, ruleR1Answers(250))
	require(t, result.Summary.ChunkSize == 100 && result.Summary.Workers == 1,
		"summary chunk size %d workers %d, want 100 and 1", result.Summary.ChunkSize, result.Summary.Workers)
}
