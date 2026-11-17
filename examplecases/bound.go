package examplecases

import (
	"errors"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	fdef "github.com/aosanya/mwanachama-backend-compute/function"
)

var errBoundNotBuilt = errors.New("the run of a bound case is not built")

func RowOf(inputs []Input) (fdef.Row, error) {
	return nil, errBoundNotBuilt
}

func EvaluateBound(c Case, binding Binding, definition fdef.Definition, runtime formula.Runtime) (Outcome, error) {
	return Outcome{}, errBoundNotBuilt
}

func Run(c Case, definitions Definitions, runtime *StandIn) (Outcome, error) {
	return Evaluate(c, runtime)
}
