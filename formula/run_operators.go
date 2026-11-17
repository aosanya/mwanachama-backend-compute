package formula

import (
	"errors"
	"math"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

const (
	codeDivideByZero = "divide-by-zero"
	codeNumberTooBig = "number-too-big"
	codeEmptyList    = "empty-list"

	maxIntegerDigits = 18
	maxDecimalDigits = 60
	maxIntegerValue  = 999999999999999999
)

type runFault struct {
	code   string
	digits int
}

func (f runFault) sentence(part string) string {
	switch f.code {
	case codeDivideByZero:
		return "The part `" + part + "` divides by zero."
	case codeNumberTooBig:
		if f.digits == 0 {
			return "The part `" + part + "` gives a float that is not a finite number."
		}
		return "The part `" + part + "` gives a number with more than " + strconv.Itoa(f.digits) + " digits."
	case codeEmptyList:
		return "The part `" + part + "` needs a list with at least one item, and the list is empty."
	}
	return "The part `" + part + "` cannot be worked out."
}

func tooBigInteger() *runFault { return &runFault{code: codeNumberTooBig, digits: maxIntegerDigits} }

func tooBigDecimal() *runFault { return &runFault{code: codeNumberTooBig, digits: maxDecimalDigits} }

func tooBigFloat() *runFault { return &runFault{code: codeNumberTooBig} }

func dividesByZero() *runFault { return &runFault{code: codeDivideByZero} }

func floatFault(err error) *runFault {
	if errors.Is(err, number.ErrDivideByZero) {
		return dividesByZero()
	}
	return tooBigFloat()
}

func runOperator(op string, settings number.Settings) func(a Value, b Value) (Value, *runFault) {
	switch op {
	case "+", "-", "*", "/", "%":
		return func(a Value, b Value) (Value, *runFault) { return runArithmetic(op, a, b, settings) }
	case "<", "<=", ">", ">=", "==", "!=":
		return func(a Value, b Value) (Value, *runFault) { return BooleanValue(runCompare(op, a, b)), nil }
	case "&&":
		return func(a Value, b Value) (Value, *runFault) { return BooleanValue(a.truth && b.truth), nil }
	}
	return func(a Value, b Value) (Value, *runFault) { return BooleanValue(a.truth || b.truth), nil }
}

func runArithmetic(op string, a Value, b Value, settings number.Settings) (Value, *runFault) {
	switch {
	case a.kind == TypeFloat || b.kind == TypeFloat:
		return runFloatArithmetic(op, a, b)
	case a.kind == TypeDecimal || b.kind == TypeDecimal:
		return runDecimalArithmetic(op, a.widened().exact, b.widened().exact, settings)
	}
	return runIntegerArithmetic(op, a.whole, b.whole)
}

func runAsFloat(v Value) (number.Float, error) {
	if v.kind == TypeFloat {
		return v.real, nil
	}
	return number.NewFloat(float64(v.whole))
}

func runFloatArithmetic(op string, a Value, b Value) (Value, *runFault) {
	left, err := runAsFloat(a)
	if err != nil {
		return Value{}, floatFault(err)
	}
	right, err := runAsFloat(b)
	if err != nil {
		return Value{}, floatFault(err)
	}
	var result number.Float
	switch op {
	case "+":
		result, err = left.Add(right)
	case "-":
		result, err = left.Sub(right)
	case "*":
		result, err = left.Mul(right)
	default:
		result, err = left.Div(right)
	}
	if err != nil {
		return Value{}, floatFault(err)
	}
	return FloatValue(result), nil
}

func runDecimalArithmetic(op string, a number.Decimal, b number.Decimal, settings number.Settings) (Value, *runFault) {
	var result number.Decimal
	switch op {
	case "+":
		result = a.Add(b)
	case "-":
		result = a.Sub(b)
	case "*":
		result = a.Mul(b)
	default:
		if b.IsZero() {
			return Value{}, dividesByZero()
		}
		quotient, err := settings.Divide(a, b)
		if err != nil {
			return Value{}, dividesByZero()
		}
		result = quotient
	}
	if result.Digits() > maxDecimalDigits {
		return Value{}, tooBigDecimal()
	}
	return DecimalValue(result), nil
}

func runIntegerArithmetic(op string, a int64, b int64) (Value, *runFault) {
	var result int64
	switch op {
	case "+":
		result = a + b
		if (a > 0 && b > 0 && result < 0) || (a < 0 && b < 0 && result >= 0) {
			return Value{}, tooBigInteger()
		}
	case "-":
		result = a - b
		if (a >= 0 && b < 0 && result < 0) || (a < 0 && b > 0 && result >= 0) {
			return Value{}, tooBigInteger()
		}
	case "*":
		result = a * b
		if a != 0 && (result/a != b || (a == -1 && b == math.MinInt64)) {
			return Value{}, tooBigInteger()
		}
	default:
		if b == 0 {
			return Value{}, dividesByZero()
		}
		result = a % b
	}
	if result > maxIntegerValue || result < -maxIntegerValue {
		return Value{}, tooBigInteger()
	}
	return IntegerValue(result), nil
}

func runCompare(op string, a Value, b Value) bool {
	order := runOrder(a, b)
	switch op {
	case "<":
		return order < 0
	case "<=":
		return order <= 0
	case ">":
		return order > 0
	case ">=":
		return order >= 0
	case "==":
		return order == 0
	}
	return order != 0
}

func runOrder(a Value, b Value) int {
	switch {
	case a.kind == TypeText:
		return runTextOrder(a.words, b.words)
	case a.kind == TypeFloat || b.kind == TypeFloat:
		left, _ := runAsFloat(a)
		right, _ := runAsFloat(b)
		return runFloatOrder(left.Value(), right.Value())
	case a.kind == TypeDecimal || b.kind == TypeDecimal:
		return a.widened().exact.Cmp(b.widened().exact)
	case a.whole < b.whole:
		return -1
	case a.whole > b.whole:
		return 1
	}
	return 0
}

func runTextOrder(a string, b string) int {
	if a == b {
		return 0
	}
	return 1
}

func runFloatOrder(a float64, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}
