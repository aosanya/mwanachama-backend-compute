package laws

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	batchSeedOffset = 1 << 21
	mostChunkSize   = 50
	mostWorkers     = 4
)

func ReceivedRow(row Row) function.Row {
	received := make(function.Row, len(row.Values))
	for name, value := range row.Values {
		received[name] = receivedOf(value)
	}
	return received
}

func receivedOf(value formula.Value) numtext.Received {
	if value.Type() == formula.TypeListInteger || value.Type() == formula.TypeListDecimal {
		items := make([]string, value.Len())
		for i := range items {
			items[i] = value.Item(i).String()
		}
		return numtext.Many(items...)
	}
	return numtext.One(value.String())
}

func DefinitionOf(inputs []formula.Input) function.Definition {
	declared := make([]function.InputDecl, len(inputs))
	for i, input := range inputs {
		declared[i] = function.InputDecl{Name: input.Name, Type: input.Type, Rules: []function.Rule{{Name: problem.RuleRequired}}}
	}
	return function.Definition{Inputs: declared}
}

func rowOutcome(result batch.RowResult) string {
	if result.HasValue() {
		return "value:" + string(result.Value.Type()) + ":" + result.Value.String()
	}
	parts := make([]string, len(result.Problems))
	for i, found := range result.Problems {
		name := ""
		if found.Input != nil {
			name = *found.Input
		}
		parts[i] = found.Code + "@" + name + ":" + found.Sentence
	}
	return "problems:" + string(result.StoppedAt) + ":" + strings.Join(parts, "|")
}

func singleOutcome(compiled *batch.Compiled, row function.Row) string {
	return rowOutcome(compiled.Run(row))
}

func batchOutcomes(compiled *batch.Compiled, rows []function.Row, chunk int, workers int) []string {
	result, err := compiled.RunBatch(context.Background(), rows, batch.Options{ChunkSize: chunk, Workers: workers})
	if err != nil {
		return []string{"refused:" + err.Error()}
	}
	outcomes := make([]string, len(result.Rows))
	for i, row := range result.Rows {
		outcomes[i] = rowOutcome(row)
	}
	return outcomes
}

func validateRow(definition function.Definition, row function.Row) []problem.Problem {
	return function.Validate(definition, row, nil)
}

func compileCase(inputs []formula.Input, tree formula.Tree) (*batch.Compiled, []formula.Problem) {
	return batch.Compile(DefinitionOf(inputs), formula.Format(tree), nil, formula.DefaultOptions())
}

func receivedRows(rows []Row) []function.Row {
	received := make([]function.Row, len(rows))
	for i, row := range rows {
		received[i] = ReceivedRow(row)
	}
	return received
}

func BatchEqualsSingle(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	rows := receivedRows(NewRows(seed, c.Inputs))
	shape := rand.New(rand.NewSource(seed ^ batchSeedOffset))
	chunk, workers := 1+shape.Intn(mostChunkSize), 1+shape.Intn(mostWorkers)
	difference := func(tree formula.Tree) (string, string, string, bool) {
		compiled, problems := compileCase(c.Inputs, tree)
		if len(problems) > 0 {
			return "0", "a formula that passes Check", "a refusal: " + problems[0].Sentence, true
		}
		alone := make([]string, len(rows))
		for i, row := range rows {
			alone[i] = engine.Batch.Single(compiled, row)
		}
		together := engine.Batch.Batch(compiled, rows, chunk, workers)
		if len(together) != len(alone) {
			return "row count", fmt.Sprintf("%d rows", len(alone)), fmt.Sprintf("%d rows", len(together)), true
		}
		for i := range alone {
			if alone[i] != together[i] {
				return fmt.Sprintf("row %d", i+1), alone[i], together[i], true
			}
		}
		return "", "", "", false
	}
	fails := func(tree formula.Tree) bool {
		_, _, _, failed := difference(tree)
		return failed
	}
	if !fails(c.Tree) {
		return nil
	}
	shrunk := Shrink(c.Tree, validIn(engine, c.Inputs), fails)
	where, alone, together, _ := difference(shrunk)
	return failure("batch equals single", seed, shrunk, fmt.Sprintf("batch differs from single, seed %d, shrunk to %s, %s: alone gave %s, in the batch gave %s", seed, Show(shrunk), where, alone, together))
}

func expectedProblemInputs(inputs []formula.Input, row Row) map[string]string {
	kinds := map[string]formula.Type{}
	for _, input := range inputs {
		kinds[input.Name] = input.Type
	}
	expected := map[string]string{}
	for _, planted := range row.Planted {
		switch {
		case planted.Kind == PlantedMissing:
			expected[planted.Input] = "not given"
		case kinds[planted.Input] != formula.TypeText:
			expected[planted.Input] = "has the wrong type"
		}
	}
	return expected
}

func listedInputs(problems []problem.Problem) map[string]bool {
	listed := map[string]bool{}
	for _, found := range problems {
		if found.Input != nil {
			listed[*found.Input] = true
		}
	}
	return listed
}

func ProblemListIsComplete(engine Engine, seed int64) *Failure {
	c := NewCase(seed)
	definition := DefinitionOf(c.Inputs)
	rows := NewRows(seed, c.Inputs)
	compiled, refused := compileCase(c.Inputs, c.Tree)
	if len(refused) > 0 {
		return &Failure{Law: "problem list", Seed: seed, Message: fmt.Sprintf("problem list could not be tried, seed %d: %s", seed, refused[0].Sentence)}
	}
	for index, row := range rows {
		expected := expectedProblemInputs(c.Inputs, row)
		received := ReceivedRow(row)
		listed := listedInputs(engine.Validate(definition, received))
		for _, input := range c.Inputs {
			if reason, planted := expected[input.Name]; planted && !listed[input.Name] {
				return &Failure{Law: "problem list", Seed: seed, Message: fmt.Sprintf("problem list incomplete, seed %d, row %d: planted %d, listed %d, missing: %s %s", seed, index+1, len(expected), len(listed), input.Name, reason)}
			}
			if _, planted := expected[input.Name]; !planted && listed[input.Name] {
				return &Failure{Law: "problem list", Seed: seed, Message: fmt.Sprintf("problem list wrong, seed %d, row %d: planted %d, listed %d, not planted: %s", seed, index+1, len(expected), len(listed), input.Name)}
			}
		}
		if len(expected) > 0 && !strings.HasPrefix(engine.Batch.Single(compiled, received), "problems:validate:") {
			return &Failure{Law: "problem list", Seed: seed, Message: fmt.Sprintf("row %d has %s and was calculated, seed %d", index+1, plural(len(expected)), seed)}
		}
	}
	return nil
}
