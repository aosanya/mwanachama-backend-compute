package functions

import (
	"errors"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type Callable struct {
	function  Function
	validator function.Validator
	program   *formula.Program
	inputs    []formula.Input
}

type Outcome struct {
	Value    formula.Value
	Problems []problem.Problem
	Failure  *formula.Problem
}

func (o Outcome) Refused() bool {
	return len(o.Problems) > 0 || o.Failure != nil
}

func (o Outcome) Sentence() string {
	if len(o.Problems) > 0 {
		return o.Problems[0].Sentence
	}
	if o.Failure != nil {
		return o.Failure.Sentence
	}
	return ""
}

var numtextKinds = map[formula.Type]numtext.Type{
	formula.TypeInteger:     numtext.Integer,
	formula.TypeDecimal:     numtext.Decimal,
	formula.TypeFloat:       numtext.Float,
	formula.TypeText:        numtext.Text,
	formula.TypeBoolean:     numtext.Boolean,
	formula.TypeDate:        numtext.Date,
	formula.TypeListInteger: numtext.IntegerList,
	formula.TypeListDecimal: numtext.DecimalList,
}

func (f Function) Prepare() (Callable, error) {
	if refused := function.CheckDefinition(f.Definition); len(refused) > 0 {
		return Callable{}, errors.New(f.Key() + ": the definition is refused: " + refused[0].Sentence)
	}
	tree, problems := formula.Parse(f.Body)
	if len(problems) > 0 {
		return Callable{}, errors.New(f.Key() + ": the body cannot be read: " + problems[0].Sentence)
	}
	inputs := make([]formula.Input, len(f.Definition.Inputs))
	for i, decl := range f.Definition.Inputs {
		inputs[i] = formula.Input{Name: decl.Name, Type: decl.Type}
	}
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	if len(problems) > 0 {
		return Callable{}, errors.New(f.Key() + ": the body is refused: " + problems[0].Sentence)
	}
	if result, known := formula.ResultType(tree, inputs, nil); !known || result != f.Result {
		return Callable{}, errors.New(f.Key() + ": the body gives " + string(result) + " and the function says " + string(f.Result))
	}
	return Callable{function: f, validator: function.Prepare(f.Definition), program: program, inputs: inputs}, nil
}

func (c Callable) Function() Function {
	return c.function
}

func (c Callable) Call(row function.Row) Outcome {
	if problems := c.validator.Validate(row, nil); len(problems) > 0 {
		return Outcome{Problems: problems}
	}
	values := map[string]formula.Value{}
	for _, input := range c.inputs {
		value, read := c.read(input, row)
		if !read {
			return Outcome{Failure: &formula.Problem{Code: "bad-input", Sentence: "The input `" + input.Name + "` could not be read."}}
		}
		values[input.Name] = value
	}
	result, failure := c.program.Run(values)
	if failure != nil {
		return Outcome{Failure: failure}
	}
	return Outcome{Value: result}
}

func (c Callable) read(input formula.Input, row function.Row) (formula.Value, bool) {
	received, given := row[input.Name]
	kind := numtextKinds[input.Type]
	if !given {
		received = numtext.Many()
	}
	read, refusals := numtext.Read(kind, received)
	if len(refusals) > 0 {
		return formula.Value{}, false
	}
	return fromRead(read)
}

func fromRead(v numtext.Value) (formula.Value, bool) {
	switch v.Type {
	case numtext.Integer:
		return formula.IntegerValue(v.Integer), true
	case numtext.Decimal:
		return formula.DecimalValue(v.Decimal), true
	case numtext.Float:
		return formula.FloatValue(v.Float), true
	case numtext.Text:
		return formula.TextValue(v.Text), true
	case numtext.Boolean:
		return formula.BooleanValue(v.Boolean), true
	case numtext.Date:
		return formula.DateValue(v.Date), true
	case numtext.IntegerList:
		items := make([]int64, len(v.Items))
		for i, item := range v.Items {
			items[i] = item.Integer
		}
		return formula.IntegerListValue(items), true
	case numtext.DecimalList:
		items := make([]number.Decimal, len(v.Items))
		for i, item := range v.Items {
			items[i] = item.Decimal
		}
		return formula.DecimalListValue(items), true
	}
	return formula.Value{}, false
}
