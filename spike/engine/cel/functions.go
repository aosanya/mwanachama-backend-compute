package cel

import (
	"time"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"cel.dev/cel-go/interpreter"
	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func builtinFunctions() []cel.EnvOption {
	return []cel.EnvOption{
		decimalConstructor(),
		roundFunction(),
		toFloatFunction(),
		toDecimalFunction(),
		dateFunction("days_between", builtins.DaysBetween),
		dateFunction("months_between", builtins.MonthsBetween),
		dateFunction("years_between", builtins.YearsBetween),
		listFunction("sum", sumDecimals, decimalType, sumInts, cel.IntType),
		listFunction("count", countDecimals, cel.IntType, countInts, cel.IntType),
		listFunction("avg", avgDecimals, decimalType, avgInts, cel.DoubleType),
		listFunction("min", minDecimals, decimalType, minInts, cel.IntType),
		listFunction("max", maxDecimals, decimalType, maxInts, cel.IntType),
	}
}

func decimalConstructor() cel.EnvOption {
	return cel.Function(decimalFunction,
		cel.Overload("decimal_string", []*cel.Type{cel.StringType}, decimalType, cel.UnaryBinding(func(v ref.Val) ref.Val {
			text, ok := v.(types.String)
			if !ok {
				return types.MaybeNoSuchOverloadErr(v)
			}
			d, err := numeric.Parse(string(text))
			if err != nil {
				return types.WrapErr(err)
			}
			return decimalValue{d}
		})),
	)
}

func roundFunction() cel.EnvOption {
	return cel.Function("round",
		cel.Overload("round_decimal", []*cel.Type{decimalType, cel.IntType}, decimalType, cel.BinaryBinding(func(value, places ref.Val) ref.Val {
			d, okValue := asDecimal(value)
			p, okPlaces := places.(types.Int)
			if !okValue || !okPlaces {
				return types.MaybeNoSuchOverloadErr(value)
			}
			return decimalValue{builtins.RoundDecimal(d, int64(p))}
		})),
		cel.Overload("round_double", []*cel.Type{cel.DoubleType, cel.IntType}, cel.DoubleType, cel.BinaryBinding(func(value, places ref.Val) ref.Val {
			f, okValue := value.(types.Double)
			p, okPlaces := places.(types.Int)
			if !okValue || !okPlaces {
				return types.MaybeNoSuchOverloadErr(value)
			}
			return types.Double(builtins.RoundFloat(float64(f), int64(p)))
		})),
	)
}

func toFloatFunction() cel.EnvOption {
	return cel.Function("to_float",
		cel.Overload("to_float_decimal", []*cel.Type{decimalType}, cel.DoubleType, cel.UnaryBinding(func(v ref.Val) ref.Val {
			d, ok := asDecimal(v)
			if !ok {
				return types.MaybeNoSuchOverloadErr(v)
			}
			return types.Double(builtins.ToFloat(d))
		})),
	)
}

func toDecimalFunction() cel.EnvOption {
	return cel.Function("to_decimal",
		cel.Overload("to_decimal_double", []*cel.Type{cel.DoubleType, cel.IntType}, decimalType, cel.BinaryBinding(func(value, places ref.Val) ref.Val {
			f, okValue := value.(types.Double)
			p, okPlaces := places.(types.Int)
			if !okValue || !okPlaces {
				return types.MaybeNoSuchOverloadErr(value)
			}
			d, err := builtins.ToDecimal(float64(f), int64(p))
			if err != nil {
				return types.WrapErr(err)
			}
			return decimalValue{d}
		})),
	)
}

func dateFunction(name string, apply func(from, to time.Time) int64) cel.EnvOption {
	return cel.Function(name,
		cel.Overload(name+"_timestamp", []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.IntType, cel.BinaryBinding(func(from, to ref.Val) ref.Val {
			start, okFrom := from.(types.Timestamp)
			end, okTo := to.(types.Timestamp)
			if !okFrom || !okTo {
				return types.MaybeNoSuchOverloadErr(from)
			}
			return types.Int(apply(start.Time, end.Time))
		})),
	)
}

func listFunction(name string, onDecimals func([]numeric.Decimal) ref.Val, decimalResult *cel.Type, onInts func([]int64) ref.Val, intResult *cel.Type) cel.EnvOption {
	return cel.Function(name,
		cel.Overload(name+"_decimal_list", []*cel.Type{cel.ListType(decimalType)}, decimalResult, cel.UnaryBinding(func(v ref.Val) ref.Val {
			items, ok := decimalsOf(v)
			if !ok {
				return types.MaybeNoSuchOverloadErr(v)
			}
			return onDecimals(items)
		})),
		cel.Overload(name+"_int_list", []*cel.Type{cel.ListType(cel.IntType)}, intResult, cel.UnaryBinding(func(v ref.Val) ref.Val {
			items, ok := intsOf(v)
			if !ok {
				return types.MaybeNoSuchOverloadErr(v)
			}
			return onInts(items)
		})),
	)
}

func sumDecimals(items []numeric.Decimal) ref.Val {
	return decimalValue{builtins.SumDecimals(items)}
}

func sumInts(items []int64) ref.Val {
	return types.Int(builtins.SumInts(items))
}

func countDecimals(items []numeric.Decimal) ref.Val {
	return types.Int(builtins.CountDecimals(items))
}

func countInts(items []int64) ref.Val {
	return types.Int(builtins.CountInts(items))
}

func avgDecimals(items []numeric.Decimal) ref.Val {
	return decimalOrError(builtins.AvgDecimals(items))
}

func avgInts(items []int64) ref.Val {
	average, err := builtins.AvgInts(items)
	if err != nil {
		return types.WrapErr(err)
	}
	return types.Double(average)
}

func minDecimals(items []numeric.Decimal) ref.Val {
	return decimalOrError(builtins.MinDecimals(items))
}

func maxDecimals(items []numeric.Decimal) ref.Val {
	return decimalOrError(builtins.MaxDecimals(items))
}

func minInts(items []int64) ref.Val {
	return intOrError(builtins.MinInts(items))
}

func maxInts(items []int64) ref.Val {
	return intOrError(builtins.MaxInts(items))
}

func decimalOrError(d numeric.Decimal, err error) ref.Val {
	if err != nil {
		return types.WrapErr(err)
	}
	return decimalValue{d}
}

func intOrError(i int64, err error) ref.Val {
	if err != nil {
		return types.WrapErr(err)
	}
	return types.Int(i)
}

func decimalsOf(v ref.Val) ([]numeric.Decimal, bool) {
	if native, ok := v.Value().([]numeric.Decimal); ok {
		return native, true
	}
	return collect(v, asDecimal)
}

func intsOf(v ref.Val) ([]int64, bool) {
	if native, ok := v.Value().([]int64); ok {
		return native, true
	}
	return collect(v, func(item ref.Val) (int64, bool) {
		i, ok := item.(types.Int)
		return int64(i), ok
	})
}

func collect[T any](v ref.Val, convert func(ref.Val) (T, bool)) ([]T, bool) {
	lister, ok := v.(traits.Lister)
	if !ok {
		return nil, false
	}
	size, ok := lister.Size().(types.Int)
	if !ok {
		return nil, false
	}
	items := make([]T, size)
	for i := range items {
		item, ok := convert(lister.Get(types.Int(i)))
		if !ok {
			return nil, false
		}
		items[i] = item
	}
	return items, true
}

func foldDecimalLiterals(i interpreter.InterpretableV2) (interpreter.InterpretableV2, error) {
	call, ok := i.(interpreter.InterpretableCall)
	if !ok || call.Function() != decimalFunction || len(call.Args()) != 1 {
		return i, nil
	}
	if _, constant := call.Args()[0].(interpreter.InterpretableConst); !constant {
		return i, nil
	}
	value := call.Eval(interpreter.EmptyActivation())
	if err, failed := value.(*types.Err); failed {
		return nil, err
	}
	return interpreter.NewConstValue(call.ID(), value), nil
}
