package formula

import (
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

const codeBadFunctionResult = "bad-function-result"

type runCompiler struct {
	runtime Runtime
	options Options
	slots   map[string]int
}

type runSite struct {
	node Node
	path Path
}

func (s runSite) problem(code string, sentence string) *Problem {
	return &Problem{Code: code, Sentence: sentence, Part: Format(Tree{Root: s.node}), Path: append(Path{}, s.path...)}
}

func (s runSite) fault(fault runFault) *Problem {
	return s.problem(fault.code, fault.sentence(Format(Tree{Root: s.node})))
}

func (c *runCompiler) compile(n Node, path Path) runStep {
	body := c.body(n, path)
	return func(f *runFrame) (Value, *Problem) {
		if problem := f.charge(1); problem != nil {
			return Value{}, problem
		}
		return body(f)
	}
}

func (c *runCompiler) body(n Node, path Path) runStep {
	site := runSite{node: n, path: path}
	switch node := n.(type) {
	case Literal:
		return c.literal(node, site)
	case Name:
		slot := c.slots[node.Name]
		return func(f *runFrame) (Value, *Problem) { return f.slots[slot], nil }
	case Unary:
		return c.unary(node, path)
	case Binary:
		return c.binary(node, site)
	case Conditional:
		return c.conditional(node, path)
	case List:
		return c.list(node, site)
	case Call:
		return c.call(node, site)
	}
	return func(*runFrame) (Value, *Problem) {
		return Value{}, site.problem(codeBadTree, "The part is not a valid part.")
	}
}

func (c *runCompiler) literal(node Literal, site runSite) runStep {
	value, ok := runLiteralValue(node)
	if !ok {
		return func(*runFrame) (Value, *Problem) {
			return Value{}, site.problem(codeBadTree, "The literal `"+node.Value+"` is not a valid "+string(node.Type)+".")
		}
	}
	return func(*runFrame) (Value, *Problem) { return value, nil }
}

func runLiteralValue(node Literal) (Value, bool) {
	switch node.Type {
	case TypeInteger:
		n, err := strconv.ParseInt(node.Value, 10, 64)
		return IntegerValue(n), err == nil
	case TypeDecimal:
		d, err := number.ParseDecimal(node.Value)
		return DecimalValue(d), err == nil
	case TypeText:
		return TextValue(node.Value), true
	case TypeBoolean:
		return BooleanValue(node.Value == "true"), node.Value == "true" || node.Value == "false"
	}
	return Value{}, false
}

func (c *runCompiler) unary(node Unary, path Path) runStep {
	operand := c.compile(node.Operand, childPath(path, 0))
	return func(f *runFrame) (Value, *Problem) {
		value, problem := operand(f)
		if problem != nil {
			return Value{}, problem
		}
		return BooleanValue(!value.truth), nil
	}
}

func (c *runCompiler) binary(node Binary, site runSite) runStep {
	left := c.compile(node.Left, childPath(site.path, 0))
	right := c.compile(node.Right, childPath(site.path, 1))
	apply := runOperator(node.Op, c.options.Numbers)
	return func(f *runFrame) (Value, *Problem) {
		a, problem := left(f)
		if problem != nil {
			return Value{}, problem
		}
		b, problem := right(f)
		if problem != nil {
			return Value{}, problem
		}
		result, fault := apply(a, b)
		if fault != nil {
			return Value{}, site.fault(*fault)
		}
		return result, nil
	}
}

func (c *runCompiler) conditional(node Conditional, path Path) runStep {
	condition := c.compile(node.Condition, childPath(path, 0))
	then := c.compile(node.Then, childPath(path, 1))
	otherwise := c.compile(node.Else, childPath(path, 2))
	return func(f *runFrame) (Value, *Problem) {
		test, problem := condition(f)
		if problem != nil {
			return Value{}, problem
		}
		yes, problem := then(f)
		if problem != nil {
			return Value{}, problem
		}
		no, problem := otherwise(f)
		if problem != nil {
			return Value{}, problem
		}
		chosen, other := yes, no
		if !test.truth {
			chosen, other = no, yes
		}
		if chosen.kind == TypeInteger && other.kind == TypeDecimal {
			return chosen.widened(), nil
		}
		return chosen, nil
	}
}

func (c *runCompiler) list(node List, site runSite) runStep {
	items := make([]runStep, len(node.Items))
	for i, item := range node.Items {
		items[i] = c.compile(item, childPath(site.path, i))
	}
	return func(f *runFrame) (Value, *Problem) {
		members := make([]Value, len(items))
		kind := TypeListInteger
		for i, item := range items {
			value, problem := item(f)
			if problem != nil {
				return Value{}, problem
			}
			members[i] = value
			if value.kind == TypeDecimal {
				kind = TypeListDecimal
			}
		}
		if kind == TypeListDecimal {
			for i := range members {
				members[i] = members[i].widened()
			}
		}
		return Value{kind: kind, members: members}, nil
	}
}
