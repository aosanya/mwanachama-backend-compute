package limits

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	celgo "cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types/ref"
	"cel.dev/cel-go/common/types/traits"
	"cel.dev/cel-go/interpreter"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const celInterruptFrequency = 128

var celCostFunctions = map[string]bool{"sum": true, "count": true, "avg": true, "min": true, "max": true}

type celCostModel struct{}

func (celCostModel) CallCost(function, overloadID string, args []ref.Val, result ref.Val) *uint64 {
	var size ref.Val
	switch {
	case celCostFunctions[function] && len(args) == 1:
		size = args[0]
	case function == "range":
		size = result
	default:
		return nil
	}
	sizer, ok := size.(traits.Sizer)
	if !ok {
		return nil
	}
	count, ok := sizer.Size().Value().(int64)
	if !ok || count < 0 {
		return nil
	}
	cost := uint64(count) + 1
	return &cost
}

func CELOptions(limits Limits, env ...celgo.EnvOption) cel.Options {
	var envOptions []celgo.EnvOption
	if limits.Depth > 0 {
		envOptions = append(envOptions, celgo.ParserRecursionLimit(limits.Depth))
	}
	if limits.FormulaLength > 0 {
		envOptions = append(envOptions, celgo.ParserExpressionSizeLimit(limits.FormulaLength))
	}
	if limits.Nodes > 0 {
		envOptions = append(envOptions, celgo.ExpressionNodeLimit(limits.Nodes))
	}
	programOptions := []celgo.ProgramOption{celgo.InterruptCheckFrequency(celInterruptFrequency)}
	if limits.Cost > 0 {
		programOptions = append(programOptions, celgo.CostTracking(celCostModel{}), celgo.CostLimit(uint64(limits.Cost)))
	}
	return cel.Options{EnvOptions: append(envOptions, env...), ProgramOptions: programOptions}
}

func NewCEL(limits Limits, precheck bool, env ...celgo.EnvOption) harness.Engine {
	return guard{
		inner: cel.NewWith(CELOptions(limits, env...)),
		setup: Setup{Limits: limits, Precheck: precheck, InputChecks: true, ResultChecks: true, AbandonOnTime: true},
		as:    celTranslator{},
		start: func(p harness.Program) runner { return plainRunner{p} },
	}
}

type plainRunner struct {
	program harness.Program
}

func (r plainRunner) run(_ context.Context, inputs map[string]any) (any, error) {
	return r.program.Run(inputs)
}

var celIssue = regexp.MustCompile(`(?s)^ERROR: <input>:(-?\d+):(\d+): (.*)$`)

var celExpecting = regexp.MustCompile(`\s+expecting \{.*$`)

type celTranslator struct{}

func (celTranslator) translate(limits Limits, err error) *Failure {
	var cancelled interpreter.EvalCancelledError
	if errors.As(err, &cancelled) {
		switch cancelled.Cause {
		case interpreter.CostLimitExceeded:
			return refusal(limits, LimitCost, LayerNative, err)
		case interpreter.MemoryLimitExceeded:
			return refusal(limits, LimitListLength, LayerNative, err)
		}
		return refusal(limits, LimitTime, LayerNative, err)
	}
	text := err.Error()
	switch {
	case strings.Contains(text, "expression recursion limit exceeded"):
		return refusal(limits, LimitDepth, LayerNative, err)
	case strings.Contains(text, "expression node count exceeds limit"):
		return refusal(limits, LimitNode, LayerNative, err)
	case strings.Contains(text, "expression code point size exceeds limit"):
		return refusal(limits, LimitFormulaLength, LayerNative, err)
	}
	if found := celIssue.FindStringSubmatch(text); found != nil {
		line, _ := strconv.Atoi(found[1])
		column, _ := strconv.Atoi(found[2])
		if line > 0 {
			what := strings.TrimPrefix(firstLine(found[3]), "Syntax error: ")
			return unreadable(celWhat(what), line, max(column, 1), LayerNative, err)
		}
	}
	return passThrough(err)
}

func celWhat(text string) string {
	text = celExpecting.ReplaceAllString(text, "")
	text = strings.ReplaceAll(text, "'<EOF>'", "the end of the formula")
	return text
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}
