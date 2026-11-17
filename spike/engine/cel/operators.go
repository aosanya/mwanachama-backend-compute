package cel

import (
	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/interpreter"
	"cel.dev/cel-go/interpreter/functions"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type operandPair struct {
	suffix      string
	left, right *cel.Type
}

type operator struct {
	function string
	id       string
	result   *cel.Type
	apply    func(a, b numeric.Decimal) ref.Val
}

var operandPairs = []operandPair{
	{"decimal_decimal", decimalType, decimalType},
	{"decimal_int", decimalType, cel.IntType},
	{"int_decimal", cel.IntType, decimalType},
}

var decimalOperatorList = []operator{
	{operators.Add, "add", decimalType, func(a, b numeric.Decimal) ref.Val { return decimalValue{a.Add(b)} }},
	{operators.Subtract, "subtract", decimalType, func(a, b numeric.Decimal) ref.Val { return decimalValue{a.Sub(b)} }},
	{operators.Multiply, "multiply", decimalType, func(a, b numeric.Decimal) ref.Val { return decimalValue{a.Mul(b)} }},
	{operators.Divide, "divide", decimalType, divide},
	{operators.Less, "less", cel.BoolType, compared(func(order int) bool { return order < 0 })},
	{operators.LessEquals, "less_equals", cel.BoolType, compared(func(order int) bool { return order <= 0 })},
	{operators.Greater, "greater", cel.BoolType, compared(func(order int) bool { return order > 0 })},
	{operators.GreaterEquals, "greater_equals", cel.BoolType, compared(func(order int) bool { return order >= 0 })},
}

var operatorBindings = indexBindings(decimalOperatorList)

func indexBindings(list []operator) map[string]functions.BinaryOp {
	bindings := map[string]functions.BinaryOp{}
	for _, op := range list {
		for _, pair := range operandPairs {
			bindings[op.id+"_"+pair.suffix] = onDecimals(op.apply)
		}
	}
	return bindings
}

func decimalOperators() []cel.EnvOption {
	options := make([]cel.EnvOption, len(decimalOperatorList))
	for i, op := range decimalOperatorList {
		options[i] = declareOperator(op)
	}
	return options
}

func declareOperator(op operator) cel.EnvOption {
	overloads := make([]cel.FunctionOpt, len(operandPairs))
	for i, pair := range operandPairs {
		overloads[i] = cel.Overload(op.id+"_"+pair.suffix, []*cel.Type{pair.left, pair.right}, op.result)
	}
	return cel.Function(op.function, overloads...)
}

func divide(a, b numeric.Decimal) ref.Val {
	quotient, err := a.Div(b)
	if err != nil {
		return types.WrapErr(err)
	}
	return decimalValue{quotient}
}

func compared(accept func(order int) bool) func(a, b numeric.Decimal) ref.Val {
	return func(a, b numeric.Decimal) ref.Val {
		return types.Bool(accept(a.Cmp(b)))
	}
}

func onDecimals(apply func(a, b numeric.Decimal) ref.Val) functions.BinaryOp {
	return func(left, right ref.Val) ref.Val {
		a, okLeft := asDecimal(left)
		b, okRight := asDecimal(right)
		if !okLeft || !okRight {
			return types.MaybeNoSuchOverloadErr(left)
		}
		return apply(a, b)
	}
}

type binaryCall struct {
	id          int64
	left, right interpreter.InterpretableV2
	apply       functions.BinaryOp
}

func (b binaryCall) ID() int64 {
	return b.id
}

func (b binaryCall) Eval(activation interpreter.Activation) ref.Val {
	return b.apply(b.left.Eval(activation), b.right.Eval(activation))
}

func (b binaryCall) Exec(frame *interpreter.ExecutionFrame) ref.Val {
	return b.apply(b.left.Exec(frame), b.right.Exec(frame))
}

func bindDecimalOperators(i interpreter.InterpretableV2) (interpreter.InterpretableV2, error) {
	call, ok := i.(interpreter.InterpretableCall)
	if !ok || len(call.Args()) != 2 {
		return i, nil
	}
	apply, found := operatorBindings[call.OverloadID()]
	if !found {
		return i, nil
	}
	return binaryCall{call.ID(), call.Args()[0], call.Args()[1], apply}, nil
}
