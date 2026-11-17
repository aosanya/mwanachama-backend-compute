package own

import "fmt"

var precedence = map[string]int{
	"||": 1,
	"&&": 2,
	"==": 3, "!=": 3,
	"<": 4, "<=": 4, ">": 4, ">=": 4,
	"+": 5, "-": 5,
	"*": 6, "/": 6, "%": 6,
}

type parser struct {
	tokens []token
	at     int
}

func Parse(source string) (Node, error) {
	tokens, err := tokenize(source)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	tree, err := p.conditional()
	if err != nil {
		return nil, err
	}
	if end := p.peek(); end.kind != tokenEnd {
		return nil, readError(end.pos, "expected the end of the formula but found "+describe(end))
	}
	return tree, nil
}

func (p *parser) peek() token { return p.tokens[p.at] }

func (p *parser) advance() token {
	current := p.tokens[p.at]
	if current.kind != tokenEnd {
		p.at++
	}
	return current
}

func (p *parser) atSymbol(text string) bool {
	current := p.peek()
	return current.kind == tokenSymbol && current.text == text
}

func (p *parser) accept(text string) (token, bool) {
	if !p.atSymbol(text) {
		return token{}, false
	}
	return p.advance(), true
}

func (p *parser) expect(text string) (token, error) {
	if accepted, ok := p.accept(text); ok {
		return accepted, nil
	}
	found := p.peek()
	return token{}, readError(found.pos, fmt.Sprintf("expected %q but found %s", text, describe(found)))
}

func (p *parser) conditional() (Node, error) {
	test, err := p.binary(1)
	if err != nil {
		return nil, err
	}
	question, ok := p.accept("?")
	if !ok {
		return test, nil
	}
	then, err := p.conditional()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(":"); err != nil {
		return nil, err
	}
	other, err := p.conditional()
	if err != nil {
		return nil, err
	}
	return Conditional{Pos: question.pos, Test: test, Then: then, Else: other}, nil
}

func (p *parser) binary(least int) (Node, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for {
		operator := p.peek()
		level, isOperator := precedence[operator.text]
		if operator.kind != tokenSymbol || !isOperator || level < least {
			return left, nil
		}
		p.advance()
		right, err := p.binary(level + 1)
		if err != nil {
			return nil, err
		}
		left = Binary{Pos: operator.pos, Op: operator.text, Left: left, Right: right}
	}
}

func (p *parser) unary() (Node, error) {
	if p.atSymbol("!") || p.atSymbol("-") {
		operator := p.advance()
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return Unary{Pos: operator.pos, Op: operator.text, Operand: operand}, nil
	}
	return p.primary()
}

func (p *parser) primary() (Node, error) {
	t := p.advance()
	switch t.kind {
	case tokenName:
		return p.nameOrCall(t)
	case tokenSymbol:
		switch t.text {
		case "(":
			return p.group()
		case "[":
			return p.list(t)
		}
	default:
		if literal, ok, err := literalNode(t); ok {
			return literal, err
		}
	}
	return nil, readError(t.pos, "expected a value but found "+describe(t))
}

func (p *parser) group() (Node, error) {
	inner, err := p.conditional()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(")"); err != nil {
		return nil, err
	}
	return inner, nil
}

func (p *parser) list(open token) (Node, error) {
	if closing, ok := p.accept("]"); ok {
		return nil, readError(closing.pos, "a list needs at least one value")
	}
	elements, err := p.separated("]")
	if err != nil {
		return nil, err
	}
	return List{Pos: open.pos, Elements: elements}, nil
}

func (p *parser) nameOrCall(name token) (Node, error) {
	if _, ok := p.accept("("); ok {
		return p.call(name)
	}
	if containsVersion(name.text) {
		return nil, readError(p.peek().pos, `a versioned name must be followed by "("`)
	}
	switch name.text {
	case "true", "false":
		return BooleanLiteral{Pos: name.pos, Value: name.text == "true"}, nil
	}
	return Variable{Pos: name.pos, Name: name.text}, nil
}

func (p *parser) call(name token) (Node, error) {
	if _, ok := p.accept(")"); ok {
		return Call{Pos: name.pos, Name: name.text}, nil
	}
	args, err := p.separated(")")
	if err != nil {
		return nil, err
	}
	return Call{Pos: name.pos, Name: name.text, Args: args}, nil
}

func (p *parser) separated(closing string) ([]Node, error) {
	var items []Node
	for {
		item, err := p.conditional()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if _, ok := p.accept(","); !ok {
			break
		}
	}
	if _, err := p.expect(closing); err != nil {
		return nil, err
	}
	return items, nil
}

func containsVersion(name string) bool {
	for _, r := range name {
		if r == '@' {
			return true
		}
	}
	return false
}
