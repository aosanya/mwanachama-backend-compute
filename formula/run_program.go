package formula

import (
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/limit"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

const codeBadInput = "bad-input"

type Runtime interface {
	Library
	Apply(library string, name string, version int, args []Value) (Value, *Problem)
}

type MeteredRuntime interface {
	Runtime
	ApplyMetered(library string, name string, version int, args []Value, meter *Meter) (Value, *Problem)
}

type Options struct {
	Limits  limit.Settings
	Numbers number.Settings
}

func DefaultOptions() Options {
	return Options{Limits: limit.Proposed(), Numbers: number.DefaultSettings()}
}

type runFrame struct {
	slots   []Value
	meter   *Meter
	program *Program
}

type runStep func(*runFrame) (Value, *Problem)

type Program struct {
	whole   string
	inputs  []Input
	root    runStep
	options Options
	cost    int
}

func Plan(tree Tree, inputs []Input, runtime Runtime, options Options) (*Program, []Problem) {
	var library Library
	if runtime != nil {
		library = runtime
	}
	if problems := Check(tree, inputs, library); len(problems) > 0 {
		return nil, problems
	}
	compiler := &runCompiler{runtime: runtime, options: options, slots: map[string]int{}}
	for i, input := range inputs {
		if _, taken := compiler.slots[input.Name]; !taken {
			compiler.slots[input.Name] = i
		}
	}
	return &Program{
		whole:   Format(tree),
		inputs:  append([]Input{}, inputs...),
		root:    compiler.compile(tree.Root, Path{}),
		options: options,
		cost:    Cost(tree, library),
	}, []Problem{}
}

func (p *Program) Cost() int { return p.cost }

func (p *Program) Run(values map[string]Value) (Value, *Problem) {
	return p.RunMetered(values, NewMeter(p.options.Limits.Cost))
}

func (p *Program) RunMetered(values map[string]Value, meter *Meter) (Value, *Problem) {
	slots := make([]Value, len(p.inputs))
	for i, input := range p.inputs {
		value, found := values[input.Name]
		if !found {
			return Value{}, runInputProblem(input.Name, "The input `"+input.Name+"` is missing.")
		}
		slots[i] = value
	}
	return p.start(slots, meter)
}

func (p *Program) Apply(args []Value) (Value, *Problem) {
	if len(args) != len(p.inputs) {
		return Value{}, &Problem{Code: codeBadInput, Sentence: "The function needs " + strconv.Itoa(len(p.inputs)) + " arguments and was given " + strconv.Itoa(len(args)) + ".", Path: Path{}}
	}
	return p.start(append([]Value{}, args...), NewMeter(p.options.Limits.Cost))
}

func (p *Program) ApplyMetered(args []Value, meter *Meter) (Value, *Problem) {
	if len(args) != len(p.inputs) {
		return Value{}, &Problem{Code: codeBadInput, Sentence: "The function needs " + strconv.Itoa(len(p.inputs)) + " arguments and was given " + strconv.Itoa(len(args)) + ".", Path: Path{}}
	}
	if meter.depth > p.options.Limits.CallDepth {
		refused := limit.Refusal{Limit: limit.CallDepthLimit, Value: p.options.Limits.CallDepth}
		return Value{}, &Problem{Code: codeCallTooDeep, Sentence: refused.Sentence(), Part: p.whole, Path: Path{}}
	}
	return p.start(append([]Value{}, args...), meter)
}

func (p *Program) start(slots []Value, meter *Meter) (Value, *Problem) {
	for i, input := range p.inputs {
		if slots[i].kind != input.Type {
			return Value{}, runInputProblem(input.Name, "The input `"+input.Name+"` must be "+article(input.Type)+" and is "+article(slots[i].kind)+".")
		}
	}
	return p.root(&runFrame{slots: slots, meter: meter, program: p})
}

func runInputProblem(name string, sentence string) *Problem {
	return &Problem{Code: codeBadInput, Sentence: sentence, Part: name, Path: Path{}}
}

func (f *runFrame) charge(units int) *Problem {
	if f.meter.Charge(units) {
		return nil
	}
	refused := limit.Refusal{Limit: limit.CostLimit, Value: f.meter.Limit()}
	return &Problem{Code: codeTooCostly, Sentence: refused.Sentence(), Part: f.program.whole, Path: Path{}}
}
