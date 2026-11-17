package cel

import (
	"fmt"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/common/types/ref"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

var publishedCases = indexPublished(harness.Cases())

func indexPublished(cases []harness.Case) map[string]harness.Case {
	index := map[string]harness.Case{}
	for _, c := range cases {
		if c.PublishedAs != "" {
			index[c.PublishedAs] = c
		}
	}
	return index
}

func libraryFunctions(c harness.Case, library harness.Library) ([]cel.EnvOption, error) {
	options := make([]cel.EnvOption, len(c.Uses))
	for i, name := range c.Uses {
		published, known := publishedCases[name]
		callee, loaded := library[name]
		if !known || !loaded {
			return nil, fmt.Errorf("%s is not in the library", name)
		}
		options[i] = libraryFunction(name, published, callee)
	}
	return options, nil
}

func libraryFunction(name string, published harness.Case, callee harness.Program) cel.EnvOption {
	functionName := translate(name)
	return cel.Function(functionName,
		cel.Overload(functionName+"_call", argumentTypes(published.Params), typeOf(published.Result), cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			out, err := callee.Run(inputsFor(published.Params, args))
			if err != nil {
				return types.WrapErr(err)
			}
			return adapter{}.NativeToValue(out)
		})),
	)
}

func argumentTypes(params []harness.Param) []*cel.Type {
	list := make([]*cel.Type, len(params))
	for i, p := range params {
		list[i] = typeOf(p.Kind)
	}
	return list
}

func inputsFor(params []harness.Param, args []ref.Val) map[string]any {
	inputs := make(map[string]any, len(params))
	for i, p := range params {
		inputs[p.Name] = nativeOf(args[i], p.Kind)
	}
	return inputs
}
