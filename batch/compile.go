package batch

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

type Compiled struct {
	inputs    []formula.Input
	validator function.Validator
	runner    function.Runner
	program   *formula.Program
	options   formula.Options
	hash      string
}

func Compile(definition function.Definition, body string, runtime formula.Runtime, options formula.Options) (*Compiled, []formula.Problem) {
	if problems := function.CheckDefinition(definition); len(problems) > 0 {
		return nil, problems
	}
	tree, problems := formula.Parse(body)
	if len(problems) > 0 {
		return nil, problems
	}
	inputs := make([]formula.Input, len(definition.Inputs))
	for i, declared := range definition.Inputs {
		inputs[i] = formula.Input{Name: declared.Name, Type: declared.Type}
	}
	program, problems := formula.Plan(tree, inputs, runtime, options)
	if len(problems) > 0 {
		return nil, problems
	}
	digest := sha256.Sum256([]byte(formula.Format(tree)))
	return &Compiled{
		inputs:    inputs,
		validator: function.Prepare(definition),
		runner:    newCheckRunner(options),
		program:   program,
		options:   options,
		hash:      hex.EncodeToString(digest[:]),
	}, []formula.Problem{}
}

func Merge(shared function.Row, own function.Row) function.Row {
	merged := make(function.Row, len(shared)+len(own))
	for name, received := range shared {
		merged[name] = received
	}
	for name, received := range own {
		merged[name] = received
	}
	return merged
}
