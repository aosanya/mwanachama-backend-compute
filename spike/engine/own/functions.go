package own

import (
	"fmt"
	"maps"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type overload struct {
	params []harness.Kind
	result harness.Kind
	build  func(args []code) any
}

var builtinFunctions = joinCatalogs(listFunctions(), dateFunctions(), numberFunctions())

func joinCatalogs(catalogs ...map[string][]overload) map[string][]overload {
	joined := map[string][]overload{}
	for _, catalog := range catalogs {
		maps.Copy(joined, catalog)
	}
	return joined
}

func overload1[A, R any](param, result harness.Kind, f func(A) (R, error)) overload {
	return overload{
		params: []harness.Kind{param},
		result: result,
		build:  func(args []code) any { return apply1(as[A](args[0]), f) },
	}
}

func overload2[A, B, R any](first, second, result harness.Kind, f func(A, B) (R, error)) overload {
	return overload{
		params: []harness.Kind{first, second},
		result: result,
		build:  func(args []code) any { return apply2(as[A](args[0]), as[B](args[1]), f) },
	}
}

func (c *compiler) call(n Call) (code, error) {
	options, err := c.overloadsOf(n)
	if err != nil {
		return code{}, err
	}
	args := make([]code, len(n.Args))
	for i, arg := range n.Args {
		if args[i], err = c.expr(arg); err != nil {
			return code{}, err
		}
	}
	chosen, ok := pick(options, args)
	if !ok {
		return code{}, typeError(n.Pos, signatureMessage(n.Name, options, args))
	}
	if err := c.guardArguments(n, args); err != nil {
		return code{}, err
	}
	return code{chosen.result, chosen.build(args)}, nil
}

func (c *compiler) overloadsOf(n Call) ([]overload, error) {
	if strings.Contains(n.Name, "@") {
		return c.libraryOverloads(n.Name)
	}
	if options, ok := builtinFunctions[n.Name]; ok {
		return options, nil
	}
	if extra, ok := c.options.Functions[n.Name]; ok {
		return []overload{extra.overload()}, nil
	}
	return nil, unknownFunction(n.Name)
}

func pick(options []overload, args []code) (overload, bool) {
	for _, option := range options {
		if matches(option.params, args) {
			return option, true
		}
	}
	return overload{}, false
}

func matches(params []harness.Kind, args []code) bool {
	if len(params) != len(args) {
		return false
	}
	for i, arg := range args {
		if arg.kind != params[i] {
			return false
		}
	}
	return true
}

func signatureMessage(name string, options []overload, args []code) string {
	accepted := make([]string, len(options))
	for i, option := range options {
		accepted[i] = signature(option.params)
	}
	given := make([]harness.Kind, len(args))
	for i, arg := range args {
		given[i] = arg.kind
	}
	return fmt.Sprintf("the function %s accepts %s but was given %s", name, strings.Join(accepted, " or "), signature(given))
}

func signature(kinds []harness.Kind) string {
	names := make([]string, len(kinds))
	for i, kind := range kinds {
		names[i] = kind.String()
	}
	return "(" + strings.Join(names, ", ") + ")"
}

func (c *compiler) libraryOverloads(name string) ([]overload, error) {
	entry, found := c.library[name]
	if !found {
		return nil, unknownFunction(name)
	}
	callee, ok := entry.(*program)
	if !ok {
		return nil, fmt.Errorf("The function %s was not built by this engine.", name)
	}
	return []overload{callee.asOverload()}, nil
}

func (p *program) asOverload() overload {
	return overload{
		params: kindsOf(p.params),
		result: p.result,
		build:  p.invocation,
	}
}

func (p *program) invocation(args []code) any {
	setters := make([]setter, len(args))
	for i, arg := range args {
		setters[i] = kindTable[arg.kind].store(arg.fn, i)
	}
	return kindTable[p.result].invoke(p.body, setters, len(p.params))
}

func kindsOf(params []harness.Param) []harness.Kind {
	kinds := make([]harness.Kind, len(params))
	for i, param := range params {
		kinds[i] = param.Kind
	}
	return kinds
}
