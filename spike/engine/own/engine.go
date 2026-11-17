package own

import (
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type Options struct {
	MaxDepth int
	MaxNodes int
	MaxSteps int
}

type engine struct {
	options Options
}

func New() harness.Engine { return NewWith(Options{}) }

func NewWith(options Options) harness.Engine { return engine{options: options} }

func (engine) Name() string { return "own" }

func (e engine) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	tree, err := Parse(c.Source)
	if err != nil {
		return nil, err
	}
	if err := e.options.admit(tree); err != nil {
		return nil, err
	}
	compiled, err := build(tree, c, library)
	if err != nil {
		return nil, err
	}
	return compiled, nil
}

func (o Options) admit(tree Node) error {
	if depth := Depth(tree); o.MaxDepth > 0 && depth > o.MaxDepth {
		return fmt.Errorf("The formula is nested too deeply: %d levels where %d are allowed.", depth, o.MaxDepth)
	}
	if nodes := Count(tree); o.MaxNodes > 0 && nodes > o.MaxNodes {
		return fmt.Errorf("The formula is too large: %d parts where %d are allowed.", nodes, o.MaxNodes)
	}
	return nil
}
