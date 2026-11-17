package exprlang

import (
	"fmt"
	"time"

	"github.com/expr-lang/expr"

	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type function struct {
	name       string
	call       func(args ...any) (any, error)
	signatures []any
}

var clashingBuiltins = []string{"round", "sum", "count", "avg", "min", "max"}

var functions = []function{
	{"round", round, roundSignatures()},
	{"sum", sum, []any{signature(decimalType, decimalsType), signature(int64Type, int64sType)}},
	{"count", count, []any{signature(int64Type, decimalsType), signature(int64Type, int64sType)}},
	{"avg", avg, []any{signature(decimalType, decimalsType), signature(floatType, int64sType)}},
	{"min", extreme(builtins.MinDecimals, builtins.MinInts), extremeSignatures()},
	{"max", extreme(builtins.MaxDecimals, builtins.MaxInts), extremeSignatures()},
	{"to_float", toFloat, []any{signature(floatType, decimalType)}},
	{"to_decimal", toDecimal, toDecimalSignatures()},
	{"days_between", dateDistance(builtins.DaysBetween), []any{signature(int64Type, timeType, timeType)}},
	{"months_between", dateDistance(builtins.MonthsBetween), []any{signature(int64Type, timeType, timeType)}},
	{"years_between", dateDistance(builtins.YearsBetween), []any{signature(int64Type, timeType, timeType)}},
}

func functionOptions() []expr.Option {
	var options []expr.Option
	for _, name := range clashingBuiltins {
		options = append(options, expr.DisableBuiltin(name))
	}
	for _, f := range functions {
		options = append(options, expr.Function(f.name, f.call, f.signatures...))
	}
	return append(options, expr.ConstExpr("dec"))
}

func unsupported(name string, args ...any) error {
	return fmt.Errorf("%s does not accept %T", name, args)
}

func roundSignatures() []any {
	var signatures []any
	for _, places := range integerTypes {
		signatures = append(signatures, signature(decimalType, decimalType, places), signature(floatType, floatType, places))
	}
	return signatures
}

func round(args ...any) (any, error) {
	places, _ := asInt64(args[1])
	switch v := args[0].(type) {
	case numeric.Decimal:
		return builtins.RoundDecimal(v, places), nil
	case float64:
		return builtins.RoundFloat(v, places), nil
	}
	return nil, unsupported("round", args[0])
}

func sum(args ...any) (any, error) {
	switch list := args[0].(type) {
	case []numeric.Decimal:
		return builtins.SumDecimals(list), nil
	case []int64:
		return builtins.SumInts(list), nil
	}
	return nil, unsupported("sum", args[0])
}

func count(args ...any) (any, error) {
	switch list := args[0].(type) {
	case []numeric.Decimal:
		return builtins.CountDecimals(list), nil
	case []int64:
		return builtins.CountInts(list), nil
	}
	return nil, unsupported("count", args[0])
}

func avg(args ...any) (any, error) {
	switch list := args[0].(type) {
	case []numeric.Decimal:
		return builtins.AvgDecimals(list)
	case []int64:
		return builtins.AvgInts(list)
	}
	return nil, unsupported("avg", args[0])
}

func extremeSignatures() []any {
	return []any{signature(decimalType, decimalsType), signature(int64Type, int64sType), signature(anyType, anysType)}
}

func extreme(
	decimals func([]numeric.Decimal) (numeric.Decimal, error),
	ints func([]int64) (int64, error),
) func(args ...any) (any, error) {
	return func(args ...any) (any, error) {
		switch list := args[0].(type) {
		case []numeric.Decimal:
			return decimals(list)
		case []int64:
			return ints(list)
		case []any:
			return extremeOfMixed(list, decimals, ints)
		}
		return nil, unsupported("min or max", args[0])
	}
}

func extremeOfMixed(
	list []any,
	decimals func([]numeric.Decimal) (numeric.Decimal, error),
	ints func([]int64) (int64, error),
) (any, error) {
	if asDecimals, ok := allAs(list, asDecimal); ok && hasDecimal(list) {
		return decimals(asDecimals)
	}
	if asInts, ok := allAs(list, asInt64); ok {
		return ints(asInts)
	}
	return nil, unsupported("min or max", list)
}

func allAs[T any](list []any, convert func(any) (T, bool)) ([]T, bool) {
	converted := make([]T, len(list))
	for i, v := range list {
		item, ok := convert(v)
		if !ok {
			return nil, false
		}
		converted[i] = item
	}
	return converted, true
}

func hasDecimal(list []any) bool {
	for _, v := range list {
		if _, ok := v.(numeric.Decimal); ok {
			return true
		}
	}
	return false
}

func toFloat(args ...any) (any, error) {
	return builtins.ToFloat(args[0].(numeric.Decimal)), nil
}

func toDecimalSignatures() []any {
	var signatures []any
	for _, places := range integerTypes {
		signatures = append(signatures, signature(decimalType, floatType, places))
	}
	return signatures
}

func toDecimal(args ...any) (any, error) {
	places, _ := asInt64(args[1])
	return builtins.ToDecimal(args[0].(float64), places)
}

func dateDistance(distance func(from, to time.Time) int64) func(args ...any) (any, error) {
	return func(args ...any) (any, error) {
		return distance(args[0].(time.Time), args[1].(time.Time)), nil
	}
}
