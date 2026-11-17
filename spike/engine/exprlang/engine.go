package exprlang

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type Options struct {
	ExprOptions  []expr.Option
	MemoryBudget uint
}

type engine struct {
	options Options
	pool    *sync.Pool
}

func New() harness.Engine {
	return NewWith(Options{})
}

func NewWith(options Options) harness.Engine {
	pool := &sync.Pool{New: func() any { return &vm.VM{MemoryBudget: options.MemoryBudget} }}
	return engine{options: options, pool: pool}
}

func (engine) Name() string {
	return "expr-lang"
}

func (e engine) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	options, err := e.compileOptions(c, library)
	if err != nil {
		return nil, err
	}
	compiled, err := expr.Compile(translate(c.Source), options...)
	if err != nil {
		return nil, err
	}
	return &program{compiled: compiled, pool: e.pool, params: c.Params, result: c.Result}, nil
}

func (e engine) compileOptions(c harness.Case, library harness.Library) ([]expr.Option, error) {
	env, err := environment(c.Params)
	if err != nil {
		return nil, err
	}
	options := append([]expr.Option{expr.Env(env)}, languageOptions...)
	for name, entry := range library {
		published, ok := entry.(*program)
		if !ok {
			return nil, fmt.Errorf("library entry %s was not compiled by %s", name, e.Name())
		}
		options = append(options, expr.Function(callableName(name), published.call, published.signature()))
	}
	return append(options, e.options.ExprOptions...), nil
}

var languageOptions = append(operatorOptions(), functionOptions()...)

func callableName(published string) string {
	return strings.Replace(published, "@", "_v", 1)
}

func environment(params []harness.Param) (map[string]any, error) {
	env := map[string]any{"dec": numeric.Parse}
	for _, p := range params {
		t, ok := kindTypes[p.Kind]
		if !ok {
			return nil, fmt.Errorf("parameter %s has no expr type for kind %s", p.Name, p.Kind)
		}
		env[p.Name] = reflect.Zero(t).Interface()
	}
	return env, nil
}

type program struct {
	compiled *vm.Program
	pool     *sync.Pool
	params   []harness.Param
	result   harness.Kind
}

func (p *program) Run(inputs map[string]any) (any, error) {
	machine := p.pool.Get().(*vm.VM)
	defer p.pool.Put(machine)
	out, err := machine.Run(p.compiled, inputs)
	if err != nil {
		return nil, err
	}
	if n, ok := out.(int); ok {
		return int64(n), nil
	}
	return out, nil
}

func (p *program) call(args ...any) (any, error) {
	inputs := make(map[string]any, len(p.params))
	for i, param := range p.params {
		inputs[param.Name] = args[i]
	}
	return p.Run(inputs)
}

func (p *program) signature() any {
	in := make([]reflect.Type, len(p.params))
	for i, param := range p.params {
		in[i] = kindTypes[param.Kind]
	}
	return signature(kindTypes[p.result], in...)
}
