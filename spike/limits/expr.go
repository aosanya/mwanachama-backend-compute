package limits

import (
	"context"
	"errors"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/file"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/exprlang"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const ExprContextName = "ctx"

func ExprOptions(limits Limits, extra ...expr.Option) exprlang.Options {
	options := []expr.Option{
		expr.MaxNodes(uint(limits.Nodes)),
		expr.WithContext(ExprContextName),
		expr.AllowUndefinedVariables(),
	}
	return exprlang.Options{ExprOptions: append(options, extra...), MemoryBudget: uint(limits.Cost)}
}

func NewExpr(limits Limits, precheck bool, extra ...expr.Option) harness.Engine {
	return guard{
		inner: exprlang.NewWith(ExprOptions(limits, extra...)),
		setup: Setup{Limits: limits, Precheck: precheck, InputChecks: true, ResultChecks: true, AbandonOnTime: true},
		as:    exprTranslator{},
		start: func(p harness.Program) runner { return exprRunner{p, limits} },
	}
}

type exprRunner struct {
	program harness.Program
	limits  Limits
}

func (r exprRunner) run(ctx context.Context, inputs map[string]any) (any, error) {
	withContext := make(map[string]any, len(inputs)+1)
	for name, value := range inputs {
		withContext[name] = value
	}
	withContext[ExprContextName] = withMeter(ctx, r.limits)
	return r.program.Run(withContext)
}

type exprTranslator struct{}

func (exprTranslator) translate(limits Limits, err error) *Failure {
	text := err.Error()
	switch {
	case strings.Contains(text, "exceeds maximum allowed nodes"):
		return refusal(limits, LimitNode, LayerNative, err)
	case strings.Contains(text, "memory budget exceeded"):
		return refusal(limits, LimitCost, LayerNative, err)
	}
	var located *file.Error
	if errors.As(err, &located) && located.Line > 0 {
		return unreadable(located.Message, located.Line, located.Column+1, LayerNative, err)
	}
	return passThrough(err)
}
