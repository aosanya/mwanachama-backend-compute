package examplecases

import (
	"errors"
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	fdef "github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

var ErrNotReceived = errors.New("a Validate row gives every value as received text or a list of received text")

func RowOf(inputs []Input) (fdef.Row, error) {
	row := make(fdef.Row, len(inputs))
	for _, input := range inputs {
		switch input.Value.Type {
		case "text":
			row[input.Name] = numtext.One(input.Value.Text)
		case "list<text>":
			row[input.Name] = numtext.Many(input.Value.Items...)
		default:
			return nil, fmt.Errorf("input %q is %s: %w", input.Name, input.Value.Type, ErrNotReceived)
		}
	}
	return row, nil
}

func EvaluateBound(c Case, binding Binding, definition fdef.Definition, runtime formula.Runtime) (Outcome, error) {
	if binding.Target == TargetDefinition {
		return problemsOutcome(fdef.CheckDefinition(definition)), nil
	}
	row, err := RowOf(c.Inputs)
	if err != nil {
		return Outcome{}, err
	}
	compiled, problems := batch.Compile(definition, c.Formula, runtime, formula.DefaultOptions())
	if len(problems) != 0 {
		return problemsOutcome(problems), nil
	}
	result := compiled.Run(row)
	if result.HasValue() {
		shown := shownResult(result.Value)
		return Outcome{Result: &shown}, nil
	}
	outcome := Outcome{Problems: make([]string, len(result.Problems))}
	for i, found := range result.Problems {
		outcome.Problems[i] = found.Code
		outcome.Where = append(outcome.Where, formula.Problem{Code: found.Code, Sentence: found.Sentence})
	}
	return outcome, nil
}

func Run(c Case, definitions Definitions, runtime *StandIn) (Outcome, error) {
	binding, bound := definitions.Bindings[c.Name]
	if !bound {
		return Evaluate(c, runtime)
	}
	definition, found := definitions.Definitions[binding.Definition]
	if !found {
		return Outcome{}, fmt.Errorf("%s names the definition %q, which is not in the file", c.Name, binding.Definition)
	}
	return EvaluateBound(c, binding, definition, runtime)
}
