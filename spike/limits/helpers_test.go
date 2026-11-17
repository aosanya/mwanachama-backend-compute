package limits

import (
	"context"
	"strconv"
	"strings"

	celgo "cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/expr-lang/expr"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const rangeChunk = 4096

type engineFactory struct {
	name  string
	build func(limits Limits, precheck bool) harness.Engine
}

var factories = []engineFactory{
	{"own", func(l Limits, precheck bool) harness.Engine {
		return NewOwn(l, precheck, map[string]own.Function{"range": ownRange(l)})
	}},
	{"cel-go", func(l Limits, precheck bool) harness.Engine { return NewCEL(l, precheck, celRange(l)) }},
	{"expr-lang", func(l Limits, precheck bool) harness.Engine { return NewExpr(l, precheck, exprRange(l)) }},
}

func factoryNamed(name string) engineFactory {
	for _, f := range factories {
		if f.name == name {
			return f
		}
	}
	panic("no engine " + name)
}

func rangeCount(from, to int64) int64 {
	if to < from {
		return 0
	}
	return to - from + 1
}

func rangeSlice(from, to int64) []int64 {
	items := make([]int64, rangeCount(from, to))
	for i := range items {
		items[i] = from + int64(i)
	}
	return items
}

func ownRange(limits Limits) own.Function {
	return own.Function{
		Params: []harness.Kind{harness.KindInteger, harness.KindInteger},
		Result: harness.KindIntegerList,
		Call: func(meter own.Meter, args []any) (any, error) {
			from, to := args[0].(int64), args[1].(int64)
			count := rangeCount(from, to)
			if limits.ListLength > 0 && count > int64(limits.ListLength) {
				return nil, refusal(limits, LimitListLength, LayerFunction, nil)
			}
			items := make([]int64, 0, min(count, rangeChunk))
			for done := int64(0); done < count; done += rangeChunk {
				if err := meter.Charge(int(min(rangeChunk, count-done))); err != nil {
					return nil, err
				}
				for i := done; i < min(done+rangeChunk, count); i++ {
					items = append(items, from+i)
				}
			}
			return items, nil
		},
	}
}

func celRange(limits Limits) celgo.EnvOption {
	return celgo.Function("range", celgo.Overload("range_int_int",
		[]*celgo.Type{celgo.IntType, celgo.IntType}, celgo.ListType(celgo.IntType),
		celgo.BinaryBinding(func(from, to ref.Val) ref.Val {
			low, okLow := from.(types.Int)
			high, okHigh := to.(types.Int)
			if !okLow || !okHigh {
				return types.MaybeNoSuchOverloadErr(from)
			}
			if limits.ListLength > 0 && rangeCount(int64(low), int64(high)) > int64(limits.ListLength) {
				return types.WrapErr(refusal(limits, LimitListLength, LayerFunction, nil))
			}
			return types.NewDynamicList(types.DefaultTypeAdapter, rangeSlice(int64(low), int64(high)))
		})))
}

func exprRange(limits Limits) expr.Option {
	call := func(args ...any) (any, error) {
		ctx := args[0].(context.Context)
		from, to := toInt64(args[1]), toInt64(args[2])
		count := rangeCount(from, to)
		if limits.ListLength > 0 && count > int64(limits.ListLength) {
			return nil, refusal(limits, LimitListLength, LayerFunction, nil)
		}
		items := make([]int64, 0, min(count, rangeChunk))
		for done := int64(0); done < count; done += rangeChunk {
			if err := Charge(ctx, int(min(rangeChunk, count-done))); err != nil {
				return nil, err
			}
			for i := done; i < min(done+rangeChunk, count); i++ {
				items = append(items, from+i)
			}
		}
		return items, nil
	}
	return expr.Function("range", call,
		(func(context.Context, int, int) []int64)(nil),
		(func(context.Context, int64, int64) []int64)(nil),
		(func(context.Context, int, int64) []int64)(nil),
		(func(context.Context, int64, int) []int64)(nil),
	)
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	}
	return 0
}

func nestedMinus(levels int) string {
	return strings.Repeat("-(", levels) + "1" + strings.Repeat(")", levels)
}

func nestedMinusValue(levels int) int64 {
	if levels%2 == 1 {
		return -1
	}
	return 1
}

func balancedSum(leaves int, leaf string) string {
	if leaves == 1 {
		return leaf
	}
	half := leaves / 2
	return "(" + balancedSum(half, leaf) + " + " + balancedSum(leaves-half, leaf) + ")"
}

func formulaWithNodes(nodes int) (string, int64) {
	if nodes%2 == 1 {
		leaves := (nodes + 1) / 2
		return balancedSum(leaves, "1"), int64(leaves)
	}
	leaves := nodes / 2
	return "-" + balancedSum(leaves, "1"), -int64(leaves)
}

func digits(count int) string {
	return "1" + strings.Repeat("7", count-1)
}

func quote(text string) string {
	return strconv.Quote(text)
}
