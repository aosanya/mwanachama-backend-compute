package own

import (
	"fmt"
	"slices"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type program struct {
	params []harness.Param
	result harness.Kind
	body   any
	box    func(*frame) (any, error)
}

func build(tree Node, c harness.Case, library harness.Library) (*program, error) {
	comp, err := newCompiler(c.Params, library)
	if err != nil {
		return nil, err
	}
	root, err := comp.expr(tree)
	if err != nil {
		return nil, err
	}
	if root.kind != c.Result {
		return nil, fmt.Errorf("The formula produces %s but %s is expected.", withArticle(root.kind), withArticle(c.Result))
	}
	return &program{
		params: slices.Clone(c.Params),
		result: root.kind,
		body:   root.fn,
		box:    kindTable[root.kind].box(root.fn),
	}, nil
}

func (p *program) Run(inputs map[string]any) (any, error) {
	fr := newFrame(len(p.params))
	for i, param := range p.params {
		value, err := checkedInput(param, inputs)
		if err != nil {
			return nil, err
		}
		fr.slots[i] = value
	}
	return p.box(fr)
}

func checkedInput(param harness.Param, inputs map[string]any) (any, error) {
	value, present := inputs[param.Name]
	if !present {
		return nil, fmt.Errorf("The input %s is missing.", param.Name)
	}
	kind, known := harness.KindOf(value)
	if !known {
		return nil, fmt.Errorf("The input %s holds a value the formulas do not know.", param.Name)
	}
	if kind != param.Kind {
		return nil, fmt.Errorf("The input %s must be %s but is %s.", param.Name, withArticle(param.Kind), withArticle(kind))
	}
	return value, nil
}
