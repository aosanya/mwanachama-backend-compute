package inputchecks

import (
	"sort"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

type engine struct {
	calls int
}

func (e *engine) Run(tree formula.Tree, values map[string]numtext.Value) function.Outcome {
	e.calls++
	result, failure := evaluate(tree, values)
	if failure != "" {
		return function.Outcome{Failure: failure}
	}
	if result.Type() != formula.TypeBoolean {
		return function.Outcome{Failure: "The check did not give true or false."}
	}
	return function.Outcome{True: result.Boolean()}
}

func evaluate(tree formula.Tree, values map[string]numtext.Value) (formula.Value, string) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	inputs := make([]formula.Input, 0, len(names))
	arguments := map[string]formula.Value{}
	for _, name := range names {
		value := toValue(values[name])
		inputs = append(inputs, formula.Input{Name: name, Type: value.Type()})
		arguments[name] = value
	}
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	if len(problems) > 0 {
		return formula.Value{}, problems[0].Sentence
	}
	result, problem := program.Run(arguments)
	if problem != nil {
		return formula.Value{}, problem.Sentence
	}
	return result, ""
}

func toValue(value numtext.Value) formula.Value {
	switch value.Type {
	case numtext.Integer:
		return formula.IntegerValue(value.Integer)
	case numtext.Decimal:
		return formula.DecimalValue(value.Decimal)
	case numtext.Float:
		return formula.FloatValue(value.Float)
	case numtext.Text:
		return formula.TextValue(value.Text)
	case numtext.Boolean:
		return formula.BooleanValue(value.Boolean)
	case numtext.Date:
		return formula.DateValue(value.Date)
	case numtext.IntegerList:
		members := make([]int64, len(value.Items))
		for i, item := range value.Items {
			members[i] = item.Integer
		}
		return formula.IntegerListValue(members)
	case numtext.DecimalList:
		members := make([]number.Decimal, len(value.Items))
		for i, item := range value.Items {
			members[i] = item.Decimal
		}
		return formula.DecimalListValue(members)
	}
	return formula.Value{}
}
