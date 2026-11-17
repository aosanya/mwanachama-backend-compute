package exprlang

import (
	"fmt"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type decimalOperator struct {
	symbol string
	name   string
	result reflect.Type
	apply  func(a, b numeric.Decimal) (any, error)
}

func compared(test func(order int) bool) func(a, b numeric.Decimal) (any, error) {
	return func(a, b numeric.Decimal) (any, error) { return test(a.Cmp(b)), nil }
}

var decimalOperators = []decimalOperator{
	{"+", "decimal_add", decimalType, func(a, b numeric.Decimal) (any, error) { return a.Add(b), nil }},
	{"-", "decimal_sub", decimalType, func(a, b numeric.Decimal) (any, error) { return a.Sub(b), nil }},
	{"*", "decimal_mul", decimalType, func(a, b numeric.Decimal) (any, error) { return a.Mul(b), nil }},
	{"/", "decimal_div", decimalType, func(a, b numeric.Decimal) (any, error) { return a.Div(b) }},
	{"<", "decimal_lt", boolType, compared(func(o int) bool { return o < 0 })},
	{"<=", "decimal_le", boolType, compared(func(o int) bool { return o <= 0 })},
	{">", "decimal_gt", boolType, compared(func(o int) bool { return o > 0 })},
	{">=", "decimal_ge", boolType, compared(func(o int) bool { return o >= 0 })},
	{"==", "decimal_eq", boolType, compared(func(o int) bool { return o == 0 })},
	{"!=", "decimal_ne", boolType, compared(func(o int) bool { return o != 0 })},
}

type forgetCallTypes struct{}

func (forgetCallTypes) Visit(node *ast.Node) {
	if call, ok := (*node).(*ast.CallNode); ok {
		call.SetType(nil)
	}
}

func (forgetCallTypes) Reset() {}

func (forgetCallTypes) ShouldRepeat() bool {
	return false
}

func operatorOptions() []expr.Option {
	options := []expr.Option{expr.Patch(forgetCallTypes{})}
	for _, op := range decimalOperators {
		options = append(options, expr.Function(op.name, op.call, op.signatures()...), expr.Operator(op.symbol, op.name))
	}
	return options
}

func (op decimalOperator) call(args ...any) (any, error) {
	a, aok := asDecimal(args[0])
	b, bok := asDecimal(args[1])
	if !aok || !bok {
		return nil, fmt.Errorf("operator %s needs decimals or integers, got %T and %T", op.symbol, args[0], args[1])
	}
	return op.apply(a, b)
}

func (op decimalOperator) signatures() []any {
	var signatures []any
	for _, other := range append([]reflect.Type{decimalType}, integerTypes...) {
		signatures = append(signatures, signature(op.result, decimalType, other))
		if other != decimalType {
			signatures = append(signatures, signature(op.result, other, decimalType))
		}
	}
	return signatures
}
