package cel

import (
	"fmt"

	"cel.dev/cel-go/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const engineName = "cel-go"

type Options struct {
	EnvOptions     []cel.EnvOption
	ProgramOptions []cel.ProgramOption
}

type engine struct {
	options Options
}

func New() harness.Engine {
	return NewWith(Options{})
}

func NewWith(options Options) harness.Engine {
	return engine{options: options}
}

func (e engine) Name() string {
	return engineName
}

func (e engine) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	env, err := e.environment(c, library)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(translate(c.Source))
	if err := issues.Err(); err != nil {
		return nil, err
	}
	if err := checkResult(c, ast); err != nil {
		return nil, err
	}
	compiled, err := env.Program(ast, e.programOptions()...)
	if err != nil {
		return nil, err
	}
	return program{compiled}, nil
}

func (e engine) environment(c harness.Case, library harness.Library) (*cel.Env, error) {
	libraryOptions, err := libraryFunctions(c, library)
	if err != nil {
		return nil, err
	}
	options := []cel.EnvOption{cel.CustomTypeAdapter(adapter{}), cel.CrossTypeNumericComparisons(true)}
	options = append(options, variables(c)...)
	options = append(options, decimalOperators()...)
	options = append(options, builtinFunctions()...)
	options = append(options, libraryOptions...)
	return cel.NewEnv(append(options, e.options.EnvOptions...)...)
}

func (e engine) programOptions() []cel.ProgramOption {
	options := []cel.ProgramOption{cel.CustomDecoratorV2(foldDecimalLiterals), cel.CustomDecoratorV2(bindDecimalOperators)}
	return append(options, e.options.ProgramOptions...)
}

func checkResult(c harness.Case, ast *cel.Ast) error {
	want := typeOf(c.Result)
	if want == nil || ast.OutputType().IsExactType(want) {
		return nil
	}
	return fmt.Errorf("formula gives %s, want %s", ast.OutputType(), want)
}

type program struct {
	compiled cel.Program
}

func (p program) Run(inputs map[string]any) (any, error) {
	out, _, err := p.compiled.Eval(inputs)
	if err != nil {
		return nil, err
	}
	return out.Value(), nil
}
