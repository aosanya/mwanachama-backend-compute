package own

import (
	"fmt"
	"math"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var integerOps = map[string]func(int64, int64) (int64, error){
	"+": addInts,
	"-": subtractInts,
	"*": multiplyInts,
	"%": remainderInts,
}

var decimalOps = map[string]func(numeric.Decimal, numeric.Decimal) (numeric.Decimal, error){
	"+": certain2(numeric.Decimal.Add),
	"-": certain2(numeric.Decimal.Sub),
	"*": certain2(numeric.Decimal.Mul),
	"/": numeric.Decimal.Div,
}

var floatOps = map[string]func(float64, float64) (float64, error){
	"+": certain2(func(a, b float64) float64 { return a + b }),
	"-": certain2(func(a, b float64) float64 { return a - b }),
	"*": certain2(func(a, b float64) float64 { return a * b }),
	"/": divideFloats,
}

func arithmetic(n Binary, left, right code) (code, error) {
	left, right, err := align(n.Pos, left, right)
	if err != nil {
		return code{}, err
	}
	if left.kind == harness.KindInteger && n.Op == "/" {
		left, right = promote(left, harness.KindDecimal), promote(right, harness.KindDecimal)
	}
	if result, ok := fold(n.Op, left, right); ok {
		return result, nil
	}
	return code{}, typeError(n.Pos, fmt.Sprintf("%s cannot be used on %s values", n.Op, left.kind))
}

func fold(op string, left, right code) (code, bool) {
	switch left.kind {
	case harness.KindInteger:
		return operate(integerOps, op, left, right)
	case harness.KindDecimal:
		return operate(decimalOps, op, left, right)
	case harness.KindFloat:
		return operate(floatOps, op, left, right)
	}
	return code{}, false
}

func operate[T any](ops map[string]func(T, T) (T, error), op string, left, right code) (code, bool) {
	f, ok := ops[op]
	if !ok {
		return code{}, false
	}
	return code{left.kind, apply2(as[T](left), as[T](right), f)}, true
}

func addInts(a, b int64) (int64, error) {
	sum := a + b
	if (a >= 0) == (b >= 0) && (sum >= 0) != (a >= 0) {
		return 0, errOverflow
	}
	return sum, nil
}

func subtractInts(a, b int64) (int64, error) {
	difference := a - b
	if (a >= 0) != (b >= 0) && (difference >= 0) != (a >= 0) {
		return 0, errOverflow
	}
	return difference, nil
}

func multiplyInts(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	product := a * b
	minimumTimesMinusOne := (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64)
	if minimumTimesMinusOne || product/b != a {
		return 0, errOverflow
	}
	return product, nil
}

func remainderInts(a, b int64) (int64, error) {
	if b == 0 {
		return 0, errDivideByZero
	}
	return a % b, nil
}

func divideFloats(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errDivideByZero
	}
	return a / b, nil
}
