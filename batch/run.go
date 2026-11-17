package batch

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	batchLimitSentence  = "The batch reached its cost limit before this row was started."
	batchCancelSentence = "The batch was cancelled before this row was started."
)

func (c *Compiled) Run(row function.Row) RowResult {
	result, _ := c.runRow(row)
	return result
}

func (c *Compiled) runRow(row function.Row) (RowResult, int) {
	if problems := c.validator.Validate(row, c.runner); len(problems) > 0 {
		return RowResult{Problems: problems, StoppedAt: StageValidate}, 0
	}
	meter := formula.NewMeter(c.options.Limits.Cost)
	value, failure := c.program.RunMetered(c.valuesOf(row), meter)
	if failure != nil {
		return RowResult{Problems: []problem.Problem{calculationProblem(*failure)}, StoppedAt: StageCalculate}, meter.Spent()
	}
	return RowResult{Value: value}, meter.Spent()
}

func (c *Compiled) valuesOf(row function.Row) map[string]formula.Value {
	values := make(map[string]formula.Value, len(c.inputs))
	for _, input := range c.inputs {
		received, given := row[input.Name]
		if !given {
			if empty, isList := emptyList(input.Type); isList {
				values[input.Name] = empty
			}
			continue
		}
		read, refusals := numtext.Read(numtextTypes[input.Type], received)
		if len(refusals) > 0 {
			continue
		}
		if converted, readable := formulaValue(read); readable {
			values[input.Name] = converted
		}
	}
	return values
}

func calculationProblem(failure formula.Problem) problem.Problem {
	return problem.Problem{Rule: RuleCalculation, Sentence: failure.Sentence, Code: failure.Code}
}

func notRun(code string, sentence string) RowResult {
	return RowResult{
		Problems:  []problem.Problem{{Rule: RuleBatch, Sentence: sentence, Code: code}},
		StoppedAt: StageBatch,
	}
}
