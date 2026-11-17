package own

import (
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

type slot struct {
	index int
	kind  harness.Kind
}

type compiler struct {
	slots   map[string]slot
	library harness.Library
}

func newCompiler(params []harness.Param, library harness.Library) (*compiler, error) {
	slots := make(map[string]slot, len(params))
	for i, param := range params {
		if _, known := kindTable[param.Kind]; !known {
			return nil, fmt.Errorf("The input %s has no kind the formulas know.", param.Name)
		}
		slots[param.Name] = slot{index: i, kind: param.Kind}
	}
	return &compiler{slots: slots, library: library}, nil
}

func (c *compiler) expr(n Node) (code, error) {
	switch v := n.(type) {
	case Variable:
		return c.variable(v)
	case List:
		return c.list(v)
	case Unary:
		return c.unary(v)
	case Binary:
		return c.binary(v)
	case Conditional:
		return c.conditional(v)
	case Call:
		return c.call(v)
	}
	if literal, ok := literalCode(n); ok {
		return literal, nil
	}
	return code{}, typeError(n.Position(), "this part of the formula is not known")
}

func literalCode(n Node) (code, bool) {
	switch v := n.(type) {
	case IntegerLiteral:
		return code{harness.KindInteger, constant(v.Value)}, true
	case DecimalLiteral:
		return code{harness.KindDecimal, constant(v.Value)}, true
	case FloatLiteral:
		return code{harness.KindFloat, constant(v.Value)}, true
	case TextLiteral:
		return code{harness.KindText, constant(v.Value)}, true
	case DateLiteral:
		return code{harness.KindDate, constant(v.Value)}, true
	case BooleanLiteral:
		return code{harness.KindBoolean, constant(v.Value)}, true
	}
	return code{}, false
}

func (c *compiler) variable(n Variable) (code, error) {
	found, ok := c.slots[n.Name]
	if !ok {
		return code{}, unknownName(n.Name)
	}
	return code{found.kind, kindTable[found.kind].variable(found.index)}, nil
}

func (c *compiler) conditional(n Conditional) (code, error) {
	test, err := c.boolean(n.Test, "the condition before ?")
	if err != nil {
		return code{}, err
	}
	then, err := c.expr(n.Then)
	if err != nil {
		return code{}, err
	}
	other, err := c.expr(n.Else)
	if err != nil {
		return code{}, err
	}
	then, other, err = align(n.Pos, then, other)
	if err != nil {
		return code{}, err
	}
	return code{then.kind, kindTable[then.kind].conditional(as[bool](test), then.fn, other.fn)}, nil
}

func (c *compiler) boolean(n Node, role string) (code, error) {
	value, err := c.expr(n)
	if err != nil {
		return code{}, err
	}
	if value.kind != harness.KindBoolean {
		return code{}, typeError(n.Position(), fmt.Sprintf("%s must be a boolean but is %s", role, withArticle(value.kind)))
	}
	return value, nil
}
