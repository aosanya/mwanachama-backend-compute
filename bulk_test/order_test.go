package bulkruns

import (
	"context"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

func TestBR12ChunksThatFinishLastDoNotMoveTheirRows(t *testing.T) {
	balances := []string{"80.00", "10.00", "60.00", "30.00", "50.00", "20.00", "70.00", "40.00"}
	answers := []string{"0.40", "0.05", "0.30", "0.15", "0.25", "0.10", "0.35", "0.20"}
	rows := make([]function.Row, len(balances))
	want := make([]answer, len(balances))
	for i, balance := range balances {
		rows[i] = row("balance", balance, "rate", "0.005")
		want[i] = value(answers[i])
	}
	hold := &probe{holdValue: "80.00", waitFor: 6}
	compiled := compile(t, formulaM(), bodyProbed, hold, formula.DefaultOptions())
	result, err := compiled.RunBatch(context.Background(), rows, batch.Options{ChunkSize: 2, Workers: 4})
	requireNoError(t, err)
	require(t, hold.released.Load(), "the first chunk was never held until the later chunks finished, so the order was not tested")
	requireAnswers(t, result, want)
	requireSummary(t, result)
	requireSummaryCounts(t, result, 8, 0, 0)
	requireEqualsSingle(t, compiled, rows, nil, result)
}

func TestBR13TenThousandRowsComeBackOnceEachAndInOrderWithFourWorkers(t *testing.T) {
	compiled := compileM(t)
	result := runAndCheck(t, compiled, ruleR1Rows(10000), nil, batch.Options{ChunkSize: 100, Workers: 4}, ruleR1Answers(10000))
	seen := make(map[int]bool, 10000)
	for _, found := range result.Rows {
		require(t, !seen[found.Row], "row number %d appears twice", found.Row)
		seen[found.Row] = true
	}
	require(t, len(seen) == 10000, "%d distinct row numbers, want 10000", len(seen))
}
