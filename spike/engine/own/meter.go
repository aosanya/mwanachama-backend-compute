package own

import (
	"context"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

const defaultCheckEvery = 256

type runLimits struct {
	maxSteps    int
	every       int
	maxList     int
	maxText     int
	maxDigits   int
	maxDecimals int
	metered     bool
}

type runState struct {
	ctx        context.Context
	limits     *runLimits
	steps      int
	sinceCheck int
}

type Meter struct {
	state *runState
}

func (m Meter) Charge(steps int) error { return m.state.charge(steps) }

func (m Meter) ListLimit() int {
	if m.state == nil {
		return 0
	}
	return m.state.limits.maxList
}

func (o Options) runLimits() *runLimits {
	every := o.CheckEvery
	if every <= 0 {
		every = defaultCheckEvery
	}
	limits := &runLimits{
		maxSteps:    o.MaxSteps,
		every:       every,
		maxList:     o.MaxListLength,
		maxText:     o.MaxTextLength,
		maxDigits:   o.MaxDigits,
		maxDecimals: o.MaxDecimals,
		metered:     o.MaxSteps > 0 || o.CheckEvery > 0,
	}
	if limits.metered || limits.maxList > 0 || limits.maxText > 0 || limits.maxDigits > 0 || limits.maxDecimals > 0 || len(o.Functions) > 0 {
		return limits
	}
	return nil
}

func (l *runLimits) begin(ctx context.Context) (*runState, error) {
	if l == nil {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, timeError(context.Cause(ctx))
	}
	return &runState{ctx: ctx, limits: l}, nil
}

func (r *runState) charge(steps int) error {
	if r == nil || !r.limits.metered {
		return nil
	}
	r.steps += steps
	if r.limits.maxSteps > 0 && r.steps > r.limits.maxSteps {
		return costError(r.limits.maxSteps)
	}
	r.sinceCheck += steps
	if r.sinceCheck < r.limits.every {
		return nil
	}
	r.sinceCheck = 0
	if r.ctx.Err() != nil {
		return timeError(context.Cause(r.ctx))
	}
	return nil
}

func (l *runLimits) checkInput(value any) error {
	if l == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		return l.checkText(v)
	case []int64:
		return l.checkList(len(v))
	case []numeric.Decimal:
		return l.checkDecimalList(v)
	case numeric.Decimal:
		return l.checkDecimal(v)
	}
	return nil
}

func (l *runLimits) checkText(text string) error {
	if l.maxText <= 0 || len(text) <= l.maxText {
		return nil
	}
	if count := utf8.RuneCountInString(text); count > l.maxText {
		return textLengthError(l.maxText, count)
	}
	return nil
}

func (l *runLimits) checkList(length int) error {
	if l.maxList > 0 && length > l.maxList {
		return listLengthError(l.maxList, length)
	}
	return nil
}

func (l *runLimits) checkDecimalList(values []numeric.Decimal) error {
	if err := l.checkList(len(values)); err != nil {
		return err
	}
	for _, v := range values {
		if err := l.checkDecimal(v); err != nil {
			return err
		}
	}
	return nil
}

func (l *runLimits) checkDecimal(d numeric.Decimal) error {
	if l.maxDecimals > 0 && d.DecimalPlaces() > l.maxDecimals {
		return decimalPlacesError(l.maxDecimals, d.DecimalPlaces())
	}
	if l.maxDigits > 0 && d.IsBig() {
		if digits := len(d.String()) - d.DecimalPlaces(); digits > l.maxDigits {
			return magnitudeError(l.maxDigits, digits)
		}
	}
	return nil
}

func chargeDecimalList(fn any) any { return chargeList[numeric.Decimal](fn) }

func chargeIntegerList(fn any) any { return chargeList[int64](fn) }

func chargeList[E any](fn any) any {
	body := fn.(thunk[[]E])
	return thunk[[]E](func(fr *frame) ([]E, error) {
		items, err := body(fr)
		if err != nil {
			return nil, err
		}
		if err := fr.run.charge(len(items)); err != nil {
			return nil, err
		}
		return items, nil
	})
}

func limitPlaces(fn any, maxDecimals int) any {
	body := fn.(thunk[int64])
	return thunk[int64](func(fr *frame) (int64, error) {
		places, err := body(fr)
		if err != nil {
			return 0, err
		}
		if places > int64(maxDecimals) {
			return 0, decimalPlacesError(maxDecimals, int(min(places, int64(1<<31))))
		}
		return places, nil
	})
}
