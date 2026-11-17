package formula

import (
	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type runBuiltin func(args []Value, settings number.Settings) (Value, *runFault)

var runBuiltins = map[string]runBuiltin{
	"round": runRound,
	"sum":   runSum,
	"count": func(args []Value, _ number.Settings) (Value, *runFault) {
		return IntegerValue(int64(len(args[0].members))), nil
	},
	"avg":        runAverage,
	"min":        func(args []Value, _ number.Settings) (Value, *runFault) { return runExtreme(args[0], -1) },
	"max":        func(args []Value, _ number.Settings) (Value, *runFault) { return runExtreme(args[0], 1) },
	"to_float":   runToFloat,
	"to_decimal": runToDecimal,
	"days_between": func(args []Value, _ number.Settings) (Value, *runFault) {
		return runDays(args, datefn.DaysBetween), nil
	},
	"months_between": func(args []Value, _ number.Settings) (Value, *runFault) {
		return runDays(args, datefn.MonthsBetween), nil
	},
	"years_between": func(args []Value, _ number.Settings) (Value, *runFault) {
		return runDays(args, datefn.YearsBetween), nil
	},
}

func runDays(args []Value, between func(a datefn.Date, b datefn.Date) int) Value {
	return IntegerValue(int64(between(args[0].day, args[1].day)))
}

func (c *runCompiler) arguments(node Call, site runSite) []runStep {
	steps := make([]runStep, len(node.Args))
	for i, arg := range node.Args {
		steps[i] = c.compile(arg, childPath(site.path, i))
	}
	return steps
}

func runEvaluate(f *runFrame, steps []runStep) ([]Value, *Problem) {
	values := make([]Value, len(steps))
	for i, step := range steps {
		value, problem := step(f)
		if problem != nil {
			return nil, problem
		}
		values[i] = value
	}
	return values, nil
}

func (c *runCompiler) call(node Call, site runSite) runStep {
	steps := c.arguments(node, site)
	if isBuiltin(node) {
		return c.builtin(node, steps, site)
	}
	return c.function(node, steps, site)
}

func (c *runCompiler) builtin(node Call, steps []runStep, site runSite) runStep {
	apply := runBuiltins[node.Name]
	aggregate := aggregates[node.Name]
	settings := c.options.Numbers
	return func(f *runFrame) (Value, *Problem) {
		args, problem := runEvaluate(f, steps)
		if problem != nil {
			return Value{}, problem
		}
		if aggregate {
			if problem := f.charge(len(args[0].members)); problem != nil {
				return Value{}, problem
			}
		}
		result, fault := apply(args, settings)
		if fault != nil {
			return Value{}, site.fault(*fault)
		}
		return result, nil
	}
}

func (c *runCompiler) function(node Call, steps []runStep, site runSite) runStep {
	libraryName := ""
	if node.Library != nil {
		libraryName = *node.Library
	}
	version := *node.Version
	signature, _ := c.runtime.Function(libraryName, node.Name, version)
	runtime := c.runtime
	return func(f *runFrame) (Value, *Problem) {
		args, problem := runEvaluate(f, steps)
		if problem != nil {
			return Value{}, problem
		}
		for i, param := range signature.Params {
			if param.Type == TypeDecimal {
				args[i] = args[i].widened()
			}
		}
		if problem := f.charge(signature.Cost); problem != nil {
			return Value{}, problem
		}
		result, failure := runtime.Apply(libraryName, node.Name, version, args)
		if failure != nil {
			if failure.Part == "" {
				failure.Part = Format(Tree{Root: node})
				failure.Path = append(Path{}, site.path...)
			}
			return Value{}, failure
		}
		if signature.Result == TypeDecimal {
			result = result.widened()
		}
		if result.kind != signature.Result {
			return Value{}, site.problem(codeBadFunctionResult, "The function `"+callLabel(node)+"` gave "+article(result.kind)+" and its signature says "+article(signature.Result)+".")
		}
		return result, nil
	}
}

func runRound(args []Value, settings number.Settings) (Value, *runFault) {
	places := int(args[1].whole)
	if args[0].kind == TypeFloat {
		decimal, err := args[0].real.ToDecimal(places, settings.Mode, number.ReadShortest)
		if err != nil {
			return Value{}, tooBigFloat()
		}
		back, err := decimal.ToFloat()
		if err != nil {
			return Value{}, tooBigFloat()
		}
		return FloatValue(back), nil
	}
	return runCheckedDecimal(args[0].exact.Round(places, settings.Mode))
}

func runCheckedDecimal(d number.Decimal) (Value, *runFault) {
	if d.Digits() > maxDecimalDigits {
		return Value{}, tooBigDecimal()
	}
	return DecimalValue(d), nil
}

func runSum(args []Value, _ number.Settings) (Value, *runFault) {
	list := args[0]
	if list.kind == TypeListInteger {
		var total int64
		for _, item := range list.members {
			result, fault := runIntegerArithmetic("+", total, item.whole)
			if fault != nil {
				return Value{}, fault
			}
			total = result.whole
		}
		return IntegerValue(total), nil
	}
	total := number.DecimalFromInt(0)
	for _, item := range list.members {
		total = total.Add(item.exact)
	}
	return runCheckedDecimal(total)
}

func runAverage(args []Value, settings number.Settings) (Value, *runFault) {
	list := args[0]
	if len(list.members) == 0 {
		return Value{}, &runFault{code: codeEmptyList}
	}
	total, fault := runSum(args, settings)
	if fault != nil {
		return Value{}, fault
	}
	if list.kind == TypeListInteger {
		mean, err := number.NewFloat(float64(total.whole) / float64(len(list.members)))
		if err != nil {
			return Value{}, tooBigFloat()
		}
		return FloatValue(mean), nil
	}
	mean, err := settings.Divide(total.exact, number.DecimalFromInt(int64(len(list.members))))
	if err != nil {
		return Value{}, dividesByZero()
	}
	return runCheckedDecimal(mean)
}

func runExtreme(list Value, direction int) (Value, *runFault) {
	if len(list.members) == 0 {
		return Value{}, &runFault{code: codeEmptyList}
	}
	chosen := list.members[0]
	for _, item := range list.members[1:] {
		if runOrder(item, chosen)*direction > 0 {
			chosen = item
		}
	}
	return chosen, nil
}

func runToFloat(args []Value, _ number.Settings) (Value, *runFault) {
	converted, err := args[0].exact.ToFloat()
	if err != nil {
		return Value{}, tooBigFloat()
	}
	return FloatValue(converted), nil
}

func runToDecimal(args []Value, settings number.Settings) (Value, *runFault) {
	converted, err := args[0].real.ToDecimal(int(args[1].whole), settings.Mode, number.ReadShortest)
	if err != nil {
		return Value{}, tooBigFloat()
	}
	return runCheckedDecimal(converted)
}
