package own

import "github.com/aosanya/mwanachama-backend-compute/spike/harness"

type Options struct {
	MaxDepth      int
	MaxNodes      int
	MaxSteps      int
	MaxLength     int
	MaxListLength int
	MaxTextLength int
	MaxDigits     int
	MaxDecimals   int
	CheckEvery    int
	Functions     map[string]Function
}

type engine struct {
	options Options
}

func New() harness.Engine { return NewWith(Options{}) }

func NewWith(options Options) harness.Engine { return engine{options: options} }

func (engine) Name() string { return "own" }

func (e engine) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	tree, err := ParseWith(c.Source, e.options)
	if err != nil {
		return nil, err
	}
	if err := e.options.admit(tree); err != nil {
		return nil, err
	}
	compiled, err := build(tree, c, library, e.options)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

func (o Options) admit(tree Node) error {
	if depth := Depth(tree); o.MaxDepth > 0 && depth > o.MaxDepth {
		return depthError(o.MaxDepth, depth)
	}
	if nodes := Count(tree); o.MaxNodes > 0 && nodes > o.MaxNodes {
		return nodeError(o.MaxNodes, nodes)
	}
	return nil
}
