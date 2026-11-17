package examplecases

import (
	"errors"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type Outcome struct {
	Result   *Value
	Problems []string
	Where    []formula.Problem
}

var ErrUnsupported = errors.New("the formula package has no value of this type")

var inputTypes = map[string]formula.Type{
	"int":       formula.TypeInteger,
	"dec":       formula.TypeDecimal,
	"float":     formula.TypeFloat,
	"text":      formula.TypeText,
	"bool":      formula.TypeBoolean,
	"date":      formula.TypeDate,
	"list<int>": formula.TypeListInteger,
	"list<dec>": formula.TypeListDecimal,
}

var resultTypes = map[formula.Type]string{
	formula.TypeInteger:     "int",
	formula.TypeDecimal:     "dec",
	formula.TypeFloat:       "float",
	formula.TypeText:        "text",
	formula.TypeBoolean:     "bool",
	formula.TypeDate:        "date",
	formula.TypeListInteger: "list<int>",
	formula.TypeListDecimal: "list<dec>",
}

func Evaluate(c Case, runtime formula.Runtime) (Outcome, error) {
	declared, values, err := runInputs(c.Inputs)
	if err != nil {
		return Outcome{}, err
	}
	tree, parseProblems := formula.Parse(c.Formula)
	if len(parseProblems) != 0 {
		return problemsOutcome(parseProblems), nil
	}
	program, planProblems := formula.Plan(tree, declared, runtime, formula.DefaultOptions())
	if len(planProblems) != 0 {
		return problemsOutcome(planProblems), nil
	}
	result, problem := program.Run(values)
	if problem != nil {
		return problemsOutcome([]formula.Problem{*problem}), nil
	}
	shown := shownResult(result)
	return Outcome{Result: &shown}, nil
}

func problemsOutcome(problems []formula.Problem) Outcome {
	codes := make([]string, len(problems))
	for i, p := range problems {
		codes[i] = p.Code
	}
	return Outcome{Problems: codes, Where: problems}
}

func shownResult(v formula.Value) Value {
	kind := resultTypes[v.Type()]
	if v.Type() == formula.TypeText {
		return Value{Type: kind, Text: v.Text()}
	}
	return Value{Type: kind, Text: v.String()}
}

func runInputs(inputs []Input) ([]formula.Input, map[string]formula.Value, error) {
	declared := make([]formula.Input, 0, len(inputs))
	values := make(map[string]formula.Value, len(inputs))
	for _, input := range inputs {
		kind, known := inputTypes[input.Value.Type]
		if !known {
			return nil, nil, ErrUnsupported
		}
		value, err := runValue(input.Value)
		if err != nil {
			return nil, nil, err
		}
		declared = append(declared, formula.Input{Name: input.Name, Type: kind})
		values[input.Name] = value
	}
	return declared, values, nil
}

func runValue(v Value) (formula.Value, error) {
	switch v.Type {
	case "int":
		n, err := strconv.ParseInt(v.Text, 10, 64)
		return formula.IntegerValue(n), err
	case "dec":
		d, err := number.ParseDecimal(v.Text)
		return formula.DecimalValue(d), err
	case "float":
		f, err := strconv.ParseFloat(v.Text, 64)
		if err != nil {
			return formula.Value{}, err
		}
		n, err := number.NewFloat(f)
		return formula.FloatValue(n), err
	case "text":
		return formula.TextValue(v.Text), nil
	case "bool":
		return formula.BooleanValue(v.Text == "true"), nil
	case "date":
		d, err := datefn.Parse(v.Text)
		return formula.DateValue(d), err
	case "list<int>":
		items := make([]int64, len(v.Items))
		for i, item := range v.Items {
			n, err := strconv.ParseInt(item, 10, 64)
			if err != nil {
				return formula.Value{}, err
			}
			items[i] = n
		}
		return formula.IntegerListValue(items), nil
	case "list<dec>":
		items := make([]number.Decimal, len(v.Items))
		for i, item := range v.Items {
			d, err := number.ParseDecimal(item)
			if err != nil {
				return formula.Value{}, err
			}
			items[i] = d
		}
		return formula.DecimalListValue(items), nil
	}
	return formula.Value{}, ErrUnsupported
}
