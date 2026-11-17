package laws

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func TestLawBatchEqualsSingle(t *testing.T) {
	runLaw(t, BatchEqualsSingle)
}

func TestBrokenBatchThatCopiesAProblemOntoLaterRowsFailsTheBatchLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Batch.Batch
	broken.Batch.Batch = func(compiled *batch.Compiled, rows []function.Row, chunk int, workers int) []string {
		outcomes := real(compiled, rows, chunk, workers)
		first := -1
		for i, outcome := range outcomes {
			if strings.HasPrefix(outcome, "problems:") {
				first = i
				break
			}
		}
		if first < 0 {
			return outcomes
		}
		for i := first + 1; i < len(outcomes); i++ {
			outcomes[i] = outcomes[first]
		}
		return outcomes
	}
	failure := proveLawFails(t, BatchEqualsSingle, broken, "batch differs from single, seed ")
	if !strings.Contains(failure.Message, ": alone gave ") || !strings.Contains(failure.Message, ", in the batch gave ") {
		t.Fatalf("message %q lacks the page's closing words", failure.Message)
	}
}

func TestBrokenBatchThatReordersRowsFailsTheBatchLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Batch.Batch
	broken.Batch.Batch = func(compiled *batch.Compiled, rows []function.Row, chunk int, workers int) []string {
		outcomes := real(compiled, rows, chunk, workers)
		for i, j := 0, len(outcomes)-1; i < j; i, j = i+1, j-1 {
			outcomes[i], outcomes[j] = outcomes[j], outcomes[i]
		}
		return outcomes
	}
	proveLawFails(t, BatchEqualsSingle, broken, "batch differs from single, seed ")
}

func TestBrokenBatchThatDropsARowFailsTheBatchLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Batch.Batch
	broken.Batch.Batch = func(compiled *batch.Compiled, rows []function.Row, chunk int, workers int) []string {
		outcomes := real(compiled, rows, chunk, workers)
		if len(outcomes) > 1 {
			return outcomes[:len(outcomes)-1]
		}
		return outcomes
	}
	failure := proveLawFails(t, BatchEqualsSingle, broken, "batch differs from single, seed ")
	if !strings.Contains(failure.Message, "row count") {
		t.Fatalf("message %q does not say the row count moved", failure.Message)
	}
}

func TestBatchLawRowsEveryInputTypeBecomesAReceivedText(t *testing.T) {
	seen := map[string]bool{}
	for seed := int64(1); seed <= 200; seed++ {
		c := NewCase(seed)
		for _, row := range NewRows(seed, c.Inputs) {
			for name, received := range ReceivedRow(row) {
				if received.List {
					seen["list"] = true
				} else if received.Text != "" {
					seen["text"] = true
				}
				if _, given := row.Values[name]; !given {
					t.Fatalf("seed %d: %s was written for a value that is not in the row", seed, name)
				}
			}
			if len(ReceivedRow(row)) != len(row.Values) {
				t.Fatalf("seed %d: %d values became %d received texts", seed, len(row.Values), len(ReceivedRow(row)))
			}
		}
	}
	if !seen["list"] || !seen["text"] {
		t.Fatalf("200 seeds gave %v, wanted both text and list values", seen)
	}
}

func TestBrokenValidateThatStopsAtTheFirstProblemFailsTheProblemListLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Validate
	broken.Validate = func(definition function.Definition, row function.Row) []problem.Problem {
		found := real(definition, row)
		if len(found) > 1 {
			return found[:1]
		}
		return found
	}
	failure := proveLawFails(t, ProblemListIsComplete, broken, "problem list incomplete, seed ")
	if !strings.Contains(failure.Message, ", missing: ") {
		t.Fatalf("message %q lacks the page's closing words", failure.Message)
	}
}

func TestBrokenBatchThatCalculatesARowWithProblemsFailsTheProblemListLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Batch.Single
	broken.Batch.Single = func(compiled *batch.Compiled, row function.Row) string {
		outcome := real(compiled, row)
		if strings.HasPrefix(outcome, "problems:validate:") {
			return "value:integer:40"
		}
		return outcome
	}
	failure := proveLawFails(t, ProblemListIsComplete, broken, "row ")
	if !strings.Contains(failure.Message, " and was calculated, seed ") {
		t.Fatalf("message %q lacks the page's closing words", failure.Message)
	}
}

func TestBrokenValidateThatNamesAnInputNobodyPlantedFailsTheProblemListLaw(t *testing.T) {
	broken := RealEngine()
	real := broken.Validate
	broken.Validate = func(definition function.Definition, row function.Row) []problem.Problem {
		found := real(definition, row)
		if len(found) == 0 && len(definition.Inputs) > 0 {
			name := definition.Inputs[0].Name
			return []problem.Problem{{Input: &name, Code: problem.CodeRequired, Rule: problem.RuleRequired, Sentence: "invented"}}
		}
		return found
	}
	proveLawFails(t, ProblemListIsComplete, broken, "problem list wrong, seed ")
}
