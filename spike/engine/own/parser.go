package own

import (
	"fmt"
	"unicode/utf8"
)

var precedence = map[string]int{
	"||": 1,
	"&&": 2,
	"==": 3, "!=": 3,
	"<": 4, "<=": 4, ">": 4, ">=": 4,
	"+": 5, "-": 5,
	"*": 6, "/": 6, "%": 6,
}

const nestingCeiling = 10000

type parser struct {
	tokens     []token
	at         int
	nesting    int
	maxNesting int
	nodes      int
	maxNodes   int
}

func Parse(source string) (Node, error) {
	return ParseWith(source, Options{})
}

func ParseWith(source string, options Options) (Node, error) {
	if options.MaxLength > 0 && len(source) > options.MaxLength {
		if length := utf8.RuneCountInString(source); length > options.MaxLength {
			return nil, formulaLengthError(options.MaxLength, length)
		}
	}
	tokens, err := tokenizeWith(source, options.lexRules())
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens, maxNesting: options.nestingLimit(), maxNodes: options.MaxNodes}
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

func (o Options) nestingLimit() int {
	if o.MaxDepth > 0 {
		return min(o.MaxDepth, nestingCeiling)
	}
	return nestingCeiling
}

func (p *parser) spend() error {
	p.nodes++
	if p.maxNodes > 0 && p.nodes > p.maxNodes {
		return nodeError(p.maxNodes, p.nodes)
	}
	return nil
}

func (p *parser) conditional() (Node, error) {
	p.nesting++
	defer func() { p.nesting-- }()
	if p.nesting > p.maxNesting {
		return nil, depthError(p.maxNesting, p.nesting)
	}
	return p.conditionalBody()
}

func (p *parser) conditionalBody() (Node, error) {
	test, err := p.binary(1)
	if err != nil {
		return nil, err
	}
	question, ok := p.accept("?")
	if !ok {
		return test, nil
	}
	if err := p.spend(); err != nil {
		return nil, err
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
		if err := p.spend(); err != nil {
			return nil, err
		}
		right, err := p.binary(level + 1)
		if err != nil {
			return nil, err
		}
		left = Binary{Pos: operator.pos, Op: operator.text, Left: left, Right: right}
	}
}

func (p *parser) unary() (Node, error) {
	var operators []token
	for p.atSymbol("!") || p.atSymbol("-") {
		if err := p.spend(); err != nil {
			return nil, err
		}
		operators = append(operators, p.advance())
	}
	operand, err := p.primary()
	if err != nil {
		return nil, err
	}
	for i := len(operators) - 1; i >= 0; i-- {
		operand = Unary{Pos: operators[i].pos, Op: operators[i].text, Operand: operand}
	}
	return operand, nil
}

func (p *parser) primary() (Node, error) {
	t := p.advance()
	if !(t.kind == tokenSymbol && t.text == "(") {
		if err := p.spend(); err != nil {
			return nil, err
		}
	}
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
