package limits

import (
	"context"
	"sync/atomic"
)

type meterKey struct{}

type meter struct {
	limits Limits
	spent  atomic.Int64
}

func withMeter(ctx context.Context, limits Limits) context.Context {
	return context.WithValue(ctx, meterKey{}, &meter{limits: limits})
}

func Charge(ctx context.Context, steps int) error {
	m, metered := ctx.Value(meterKey{}).(*meter)
	if err := ctx.Err(); err != nil {
		if metered {
			return refusal(m.limits, LimitTime, LayerFunction, err)
		}
		return err
	}
	if !metered || m.limits.Cost <= 0 {
		return nil
	}
	if m.spent.Add(int64(steps)) > int64(m.limits.Cost) {
		return refusal(m.limits, LimitCost, LayerFunction, nil)
	}
	return nil
}

func ListLimitOf(ctx context.Context) int {
	if m, ok := ctx.Value(meterKey{}).(*meter); ok {
		return m.limits.ListLength
	}
	return 0
}

func ListRefusal(ctx context.Context) error {
	m, ok := ctx.Value(meterKey{}).(*meter)
	if !ok {
		return &Failure{Sentence: LimitSentence(LimitListLength, "0"), Limit: LimitListLength, Layer: LayerFunction}
	}
	return refusal(m.limits, LimitListLength, LayerFunction, nil)
}
