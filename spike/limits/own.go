package limits

import (
	"context"
	"errors"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const ownCheckEvery = 128

func OwnOptions(limits Limits, functions map[string]own.Function) own.Options {
	return own.Options{
		MaxDepth:      limits.Depth,
		MaxNodes:      limits.Nodes,
		MaxLength:     limits.FormulaLength,
		MaxSteps:      limits.Cost,
		MaxListLength: limits.ListLength,
		MaxTextLength: limits.TextLength,
		MaxDigits:     limits.NumberMagnitude,
		MaxDecimals:   limits.DecimalPlaces,
		CheckEvery:    ownCheckEvery,
		Functions:     functions,
	}
}

func NewOwn(limits Limits, precheck bool, functions map[string]own.Function) harness.Engine {
	return guard{
		inner: own.NewWith(OwnOptions(limits, functions)),
		setup: Setup{Limits: limits, Precheck: precheck},
		as:    ownTranslator{},
		start: func(p harness.Program) runner { return ownRunner{p} },
	}
}

type ownRunner struct {
	program harness.Program
}

func (r ownRunner) run(ctx context.Context, inputs map[string]any) (any, error) {
	if contextual, ok := r.program.(interface {
		RunContext(context.Context, map[string]any) (any, error)
	}); ok {
		return contextual.RunContext(ctx, inputs)
	}
	return r.program.Run(inputs)
}

var ownNames = map[string]string{
	"depth":            LimitDepth,
	"formula length":   LimitFormulaLength,
	"node":             LimitNode,
	"cost":             LimitCost,
	"time":             LimitTime,
	"list length":      LimitListLength,
	"text length":      LimitTextLength,
	"number magnitude": LimitNumberMagnitude,
	"decimal places":   LimitDecimalPlaces,
}

type ownTranslator struct{}

func (ownTranslator) translate(limits Limits, err error) *Failure {
	var limit *own.LimitError
	if errors.As(err, &limit) {
		return refusal(limits, ownNames[limit.Name], LayerNative, err)
	}
	return passThrough(err)
}

func passThrough(err error) *Failure {
	limit := "other"
	if strings.HasPrefix(err.Error(), "The formula cannot be read") {
		limit = "parser"
	}
	return &Failure{Sentence: clip(firstLine(err.Error())), Limit: limit, Layer: LayerNative, Native: err}
}
