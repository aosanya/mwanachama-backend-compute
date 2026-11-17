package library

import (
	"fmt"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/functions"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	validation "github.com/aosanya/mwanachama-backend-compute/problem"
)

func LoadFunctions(registry *Registry) error {
	for _, f := range functions.All() {
		release, err := ReleaseOf(f)
		if err != nil {
			return err
		}
		if err := registry.Publish(release); err != nil {
			return err
		}
	}
	return nil
}

func ReleaseOf(f functions.Function) (Release, error) {
	callable, err := f.Prepare()
	if err != nil {
		return Release{}, err
	}
	tree, problems := formula.Parse(f.Body)
	if len(problems) > 0 {
		return Release{}, fmt.Errorf("%s: the body cannot be read: %s", f.Key(), problems[0].Sentence)
	}
	params := make([]formula.Param, len(f.Definition.Inputs))
	for i, input := range f.Definition.Inputs {
		params[i] = formula.Param{Name: input.Name, Type: input.Type}
	}
	return Release{
		Library:   f.Library,
		Name:      f.Name,
		Version:   f.Version,
		Signature: formula.Signature{Params: params, Result: f.Result, Cost: formula.Cost(tree, nil)},
		Apply:     implementationOf(f, callable),
	}, nil
}

func implementationOf(f functions.Function, callable functions.Callable) Implementation {
	names := f.ParamOrder()
	return func(args []formula.Value) (formula.Value, *formula.Problem) {
		if len(args) != len(names) {
			return formula.Value{}, &formula.Problem{
				Code:     "bad-input",
				Sentence: fmt.Sprintf("%s takes %d values and was given %d.", f.Key(), len(names), len(args)),
			}
		}
		row := function.Row{}
		for i, name := range names {
			row[name] = receivedFrom(args[i])
		}
		outcome := callable.Call(row)
		if len(outcome.Problems) > 0 {
			return formula.Value{}, problemOf(outcome.Problems)
		}
		if outcome.Failure != nil {
			return formula.Value{}, outcome.Failure
		}
		return outcome.Value, nil
	}
}

func receivedFrom(value formula.Value) numtext.Received {
	if value.Type() != formula.TypeListInteger && value.Type() != formula.TypeListDecimal {
		return numtext.One(value.String())
	}
	items := make([]string, value.Len())
	for i := range items {
		items[i] = value.Item(i).String()
	}
	return numtext.Many(items...)
}

func problemOf(problems []validation.Problem) *formula.Problem {
	sentences := make([]string, len(problems))
	for i, p := range problems {
		sentences[i] = p.Sentence
	}
	return &formula.Problem{Code: problems[0].Code, Sentence: strings.Join(sentences, " ")}
}
