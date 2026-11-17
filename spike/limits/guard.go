package limits

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var errTimeLimit = errors.New("formula time limit")

type Setup struct {
	Limits        Limits
	Precheck      bool
	InputChecks   bool
	ResultChecks  bool
	AbandonOnTime bool
}

type translator interface {
	translate(limits Limits, err error) *Failure
}

type runner interface {
	run(ctx context.Context, inputs map[string]any) (any, error)
}

type guard struct {
	inner harness.Engine
	setup Setup
	as    translator
	start func(harness.Program) runner
}

func (g guard) Name() string { return g.inner.Name() }

func (g guard) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	if g.setup.Precheck {
		if err := Precheck(c.Source, g.setup.Limits); err != nil {
			return nil, err
		}
	}
	compiled, err := safely(func() (harness.Program, error) { return g.inner.Compile(c, unwrapped(library)) })
	if err != nil {
		return nil, g.as.translate(g.setup.Limits, err)
	}
	return &Program{guard: g, inner: compiled, runner: g.start(compiled)}, nil
}

func unwrapped(library harness.Library) harness.Library {
	inner := make(harness.Library, len(library))
	for name, entry := range library {
		if guarded, ok := entry.(*Program); ok {
			entry = guarded.inner
		}
		inner[name] = entry
	}
	return inner
}

type Program struct {
	guard  guard
	inner  harness.Program
	runner runner
}

func (p *Program) Run(inputs map[string]any) (any, error) {
	return p.RunContext(context.Background(), inputs)
}

type ran struct {
	value any
	err   error
}

func (p *Program) RunContext(parent context.Context, inputs map[string]any) (any, error) {
	setup := p.guard.setup
	if setup.InputChecks {
		if err := CheckInputs(inputs, setup.Limits); err != nil {
			return nil, err
		}
	}
	hostDeadline, hasHost := parent.Deadline()
	started := time.Now()
	ctx, cancel := context.WithCancelCause(parent)
	defer cancel(nil)
	if setup.Limits.Time > 0 {
		timer := time.AfterFunc(setup.Limits.Time, func() { cancel(errTimeLimit) })
		defer timer.Stop()
	}
	value, err := p.execute(ctx, inputs)
	if err != nil {
		return nil, p.refusal(err, ctx, started, hostDeadline, hasHost)
	}
	if setup.ResultChecks {
		if err := CheckResult(value, setup.Limits); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (p *Program) execute(ctx context.Context, inputs map[string]any) (any, error) {
	if !p.guard.setup.AbandonOnTime {
		return safely(func() (any, error) { return p.runner.run(ctx, inputs) })
	}
	done := make(chan ran, 1)
	go func() {
		value, err := safely(func() (any, error) { return p.runner.run(ctx, inputs) })
		done <- ran{value, err}
	}()
	select {
	case o := <-done:
		return o.value, o.err
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}

func (p *Program) refusal(err error, ctx context.Context, started time.Time, hostDeadline time.Time, hasHost bool) error {
	limits := p.guard.setup.Limits
	failure := p.classify(err)
	hostCaused := ctx.Err() != nil && !errors.Is(context.Cause(ctx), errTimeLimit)
	if failure.Limit == LimitTime && hostCaused {
		host := limits.Time
		if hasHost {
			host = hostDeadline.Sub(started).Round(time.Millisecond)
		}
		failure = &Failure{Sentence: LimitSentence(LimitTime, host.String()), Limit: LimitTime, Layer: failure.Layer, Native: err}
	}
	return failure
}

func (p *Program) classify(err error) *Failure {
	limits := p.guard.setup.Limits
	var failure *Failure
	if errors.As(err, &failure) {
		return failure
	}
	if translated := p.guard.as.translate(limits, err); translated.Limit != "other" {
		return translated
	}
	if errors.Is(err, errTimeLimit) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return refusal(limits, LimitTime, LayerWrapper, err)
	}
	return p.guard.as.translate(limits, err)
}

func safely[T any](f func() (T, error)) (value T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return f()
}

func CheckInputs(inputs map[string]any, limits Limits) error {
	for _, value := range inputs {
		if err := checkValue(value, limits); err != nil {
			return err
		}
	}
	return nil
}

func checkValue(value any, limits Limits) error {
	switch v := value.(type) {
	case string:
		if limits.TextLength > 0 && len(v) > limits.TextLength && len([]rune(v)) > limits.TextLength {
			return refusal(limits, LimitTextLength, LayerWrapper, nil)
		}
	case []int64:
		return checkLength(len(v), limits)
	case []numeric.Decimal:
		if err := checkLength(len(v), limits); err != nil {
			return err
		}
		for _, d := range v {
			if err := checkDecimal(d, limits); err != nil {
				return err
			}
		}
	case numeric.Decimal:
		return checkDecimal(v, limits)
	}
	return nil
}

func checkLength(length int, limits Limits) error {
	if limits.ListLength > 0 && length > limits.ListLength {
		return refusal(limits, LimitListLength, LayerWrapper, nil)
	}
	return nil
}

func checkDecimal(d numeric.Decimal, limits Limits) error {
	if limits.DecimalPlaces > 0 && d.DecimalPlaces() > limits.DecimalPlaces {
		return refusal(limits, LimitDecimalPlaces, LayerWrapper, nil)
	}
	if limits.NumberMagnitude > 0 && d.IsBig() && len(d.String())-d.DecimalPlaces() > limits.NumberMagnitude {
		return refusal(limits, LimitNumberMagnitude, LayerWrapper, nil)
	}
	return nil
}

func CheckResult(value any, limits Limits) error {
	switch v := value.(type) {
	case numeric.Decimal:
		return checkDecimal(v, limits)
	case []int64:
		return checkLength(len(v), limits)
	}
	return nil
}
