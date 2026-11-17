package formula

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/limit"
)

const (
	codeBadSyntax       = "bad-syntax"
	codeTooLong         = "too-long"
	codeTooDeep         = "too-deep"
	codeTooManyNodes    = "too-many-nodes"
	codeTooManyDigits   = "too-many-digits"
	codeTooManyPlaces   = "too-many-places"
	codeTooManyInputs   = "too-many-inputs"
	codeCallTooDeep     = "call-too-deep"
	codeTooCostly       = "too-costly"
	codeBadTree         = "bad-tree"
	codeUnknownName     = "unknown-name"
	codeUnknownFunction = "unknown-function"
	codeVersionMissing  = "version-missing"
	codeVersionNotFound = "version-not-found"
	codeVersionBuiltIn  = "version-not-allowed"
	codeArgumentCount   = "wrong-argument-count"
	codeArgumentType    = "wrong-argument-type"
	codeTypeMismatch    = "type-mismatch"
	codeConditionType   = "condition-not-boolean"
	codeBranchesDiffer  = "branches-differ"
	codeFloatDecimalMix = "float-decimal-mix"
	codeListItemsDiffer = "list-items-differ"
)

func limitCode(name limit.Name) string {
	switch name {
	case limit.FormulaLengthLimit, limit.TextLengthLimit:
		return codeTooLong
	case limit.DepthLimit:
		return codeTooDeep
	case limit.NodeLimit:
		return codeTooManyNodes
	case limit.InputLimit:
		return codeTooManyInputs
	case limit.NumberMagnitudeLimit:
		return codeTooManyDigits
	case limit.DecimalPlacesLimit:
		return codeTooManyPlaces
	case limit.CallDepthLimit:
		return codeCallTooDeep
	case limit.CostLimit:
		return codeTooCostly
	}
	return codeBadTree
}

type associativity int

const (
	leftToRight associativity = iota
)

type operator struct {
	level int
	assoc associativity
}

const (
	levelConditional = 0
	levelOr          = 1
	levelAnd         = 2
	levelEquality    = 3
	levelComparison  = 4
	levelSum         = 5
	levelProduct     = 6
	levelUnary       = 7
	levelAtom        = 8
)

var binaryOperators = map[string]operator{
	"||": {levelOr, leftToRight},
	"&&": {levelAnd, leftToRight},
	"==": {levelEquality, leftToRight},
	"!=": {levelEquality, leftToRight},
	"<":  {levelComparison, leftToRight},
	"<=": {levelComparison, leftToRight},
	">":  {levelComparison, leftToRight},
	">=": {levelComparison, leftToRight},
	"+":  {levelSum, leftToRight},
	"-":  {levelSum, leftToRight},
	"*":  {levelProduct, leftToRight},
	"/":  {levelProduct, leftToRight},
	"%":  {levelProduct, leftToRight},
}

type parser struct {
	lex       *lexer
	current   token
	previous  token
	brackets  int
	settings  limit.Settings
	lastError *syntaxFailure
}

func Parse(text string) (Tree, []Problem) {
	return parseWith(text, limit.Proposed())
}

func parseWith(text string, settings limit.Settings) (Tree, []Problem) {
	runes := []rune(text)
	if !utf8.ValidString(text) {
		return Tree{}, []Problem{syntaxProblem(runes, &syntaxFailure{offset: 0, code: codeBadSyntax, message: "text is not valid UTF-8"})}
	}
	if len(runes) > settings.FormulaLength {
		return Tree{}, []Problem{syntaxProblem(runes, refusal(0, limit.FormulaLengthLimit, settings.FormulaLength))}
	}
	p := &parser{lex: newLexer(text, settings), settings: settings}
	if failure := p.advance(); failure != nil {
		return Tree{}, []Problem{syntaxProblem(runes, failure)}
	}
	if p.current.kind == tokenEnd {
		return Tree{}, []Problem{syntaxProblem(runes, &syntaxFailure{offset: 0, code: codeBadSyntax, message: "the formula is empty"})}
	}
	root, failure := p.conditional()
	if failure == nil && p.current.kind != tokenEnd {
		failure = p.unexpectedToken(p.current)
	}
	if failure != nil {
		return Tree{}, []Problem{syntaxProblem(runes, failure)}
	}
	tree := Tree{Root: root}
	measure := limit.Measure{Depth: treeDepth(root), Nodes: nodeCount(root)}
	if refused, over := settings.Refusal(measure); over {
		return Tree{}, []Problem{syntaxProblem(runes, refusal(0, refused.Limit, refused.Value))}
	}
	return tree, nil
}

func syntaxProblem(runes []rune, failure *syntaxFailure) Problem {
	line, column := position(runes, failure.offset)
	sentence := failure.message
	if failure.code == codeBadSyntax {
		sentence = "The formula cannot be read: " + failure.message + " at line " + strconv.Itoa(line) + ", column " + strconv.Itoa(column) + "."
	}
	return Problem{Code: failure.code, Sentence: sentence, Line: line, Column: column}
}

func position(runes []rune, offset int) (int, int) {
	line, column := 1, 1
	for i := 0; i < offset && i < len(runes); i++ {
		if runes[i] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func (p *parser) advance() *syntaxFailure {
	p.previous = p.current
	next, failure := p.lex.next()
	if failure != nil {
		return failure
	}
	p.current = next
	return nil
}

func (p *parser) isSymbol(text string) bool {
	return p.current.kind == tokenSymbol && p.current.text == text
}

func startsOperand(t token) bool {
	switch t.kind {
	case tokenInteger, tokenDecimal, tokenText, tokenName:
		return true
	case tokenSymbol:
		return t.text == "(" || t.text == "[" || t.text == "!"
	}
	return false
}

func (p *parser) unexpectedToken(t token) *syntaxFailure {
	if startsOperand(t) {
		return &syntaxFailure{offset: t.start, code: codeBadSyntax, message: "two parts are side by side with no operator between them"}
	}
	r := p.lex.runes[t.start]
	return unexpected(r, t.start)
}

func (p *parser) missingClosing() *syntaxFailure {
	if p.current.kind == tokenEnd {
		return &syntaxFailure{offset: p.current.start, code: codeBadSyntax, message: "missing closing bracket"}
	}
	return p.unexpectedToken(p.current)
}

func (p *parser) operandMissing() *syntaxFailure {
	if p.current.kind != tokenEnd {
		return p.unexpectedToken(p.current)
	}
	message := "the formula ends after an operator"
	if p.previous.kind == tokenSymbol && (p.previous.text == "(" || p.previous.text == "[" || p.previous.text == ",") {
		message = "missing closing bracket"
	}
	return &syntaxFailure{offset: p.current.start, code: codeBadSyntax, message: message}
}

func (p *parser) conditional() (Node, *syntaxFailure) {
	condition, failure := p.binary(levelOr)
	if failure != nil {
		return nil, failure
	}
	if !p.isSymbol("?") {
		return condition, nil
	}
	if failure := p.advance(); failure != nil {
		return nil, failure
	}
	then, failure := p.conditional()
	if failure != nil {
		return nil, failure
	}
	if !p.isSymbol(":") {
		if startsOperand(p.current) {
			return nil, p.unexpectedToken(p.current)
		}
		return nil, &syntaxFailure{offset: p.current.start, code: codeBadSyntax, message: "a : is missing"}
	}
	if failure := p.advance(); failure != nil {
		return nil, failure
	}
	otherwise, failure := p.conditional()
	if failure != nil {
		return nil, failure
	}
	return Conditional{Condition: condition, Then: then, Else: otherwise}, nil
}

func (p *parser) binary(minimum int) (Node, *syntaxFailure) {
	left, failure := p.unary()
	if failure != nil {
		return nil, failure
	}
	for p.current.kind == tokenSymbol {
		op, known := binaryOperators[p.current.text]
		if !known || op.level < minimum {
			break
		}
		symbol := p.current.text
		if failure := p.advance(); failure != nil {
			return nil, failure
		}
		right, failure := p.binary(op.level + 1)
		if failure != nil {
			return nil, failure
		}
		left = Binary{Op: symbol, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) unary() (Node, *syntaxFailure) {
	if p.isSymbol("!") {
		if failure := p.advance(); failure != nil {
			return nil, failure
		}
		operand, failure := p.unary()
		if failure != nil {
			return nil, failure
		}
		return Unary{Op: "!", Operand: operand}, nil
	}
	return p.primary()
}

func (p *parser) open() *syntaxFailure {
	p.brackets++
	if p.brackets > p.settings.BracketsOpen {
		return refusal(p.current.start, limit.DepthLimit, p.settings.BracketsOpen)
	}
	return p.advance()
}

func (p *parser) primary() (Node, *syntaxFailure) {
	t := p.current
	switch t.kind {
	case tokenInteger:
		return p.literal(TypeInteger, t.text)
	case tokenDecimal:
		return p.literal(TypeDecimal, t.text)
	case tokenText:
		return p.literal(TypeText, t.text)
	case tokenName:
		return p.nameOrCall()
	case tokenSymbol:
		switch t.text {
		case "(":
			return p.group()
		case "[":
			return p.list()
		}
	}
	return nil, p.operandMissing()
}

func (p *parser) literal(kind Type, value string) (Node, *syntaxFailure) {
	if failure := p.advance(); failure != nil {
		return nil, failure
	}
	return Literal{Type: kind, Value: value}, nil
}

func (p *parser) group() (Node, *syntaxFailure) {
	if failure := p.open(); failure != nil {
		return nil, failure
	}
	inner, failure := p.conditional()
	if failure != nil {
		return nil, failure
	}
	if !p.isSymbol(")") {
		return nil, p.missingClosing()
	}
	p.brackets--
	if failure := p.advance(); failure != nil {
		return nil, failure
	}
	return inner, nil
}

func (p *parser) list() (Node, *syntaxFailure) {
	items, failure := p.sequence("]")
	if failure != nil {
		return nil, failure
	}
	return List{Items: items}, nil
}

func (p *parser) sequence(closing string) ([]Node, *syntaxFailure) {
	if failure := p.open(); failure != nil {
		return nil, failure
	}
	var items []Node
	for {
		item, failure := p.conditional()
		if failure != nil {
			return nil, failure
		}
		items = append(items, item)
		switch {
		case p.isSymbol(","):
			if failure := p.advance(); failure != nil {
				return nil, failure
			}
		case p.isSymbol(closing):
			p.brackets--
			if failure := p.advance(); failure != nil {
				return nil, failure
			}
			return items, nil
		default:
			return nil, p.missingClosing()
		}
	}
}

func (p *parser) adjacent(symbol string) bool {
	return p.current.kind == tokenSymbol && p.current.text == symbol && p.current.start == p.previous.finish
}

func (p *parser) nameOrCall() (Node, *syntaxFailure) {
	first := p.current
	if first.text == "true" || first.text == "false" {
		return p.literal(TypeBoolean, first.text)
	}
	if failure := p.advance(); failure != nil {
		return nil, failure
	}
	name := first.text
	var library *string
	compound := false
	if p.adjacent(".") {
		compound = true
		if failure := p.advance(); failure != nil {
			return nil, failure
		}
		if p.current.kind != tokenName || p.current.start != p.previous.finish {
			return nil, p.callNameMissing()
		}
		libraryName := name
		library = &libraryName
		name = p.current.text
		if failure := p.advance(); failure != nil {
			return nil, failure
		}
	}
	var version *int
	if p.adjacent("@") {
		compound = true
		if failure := p.advance(); failure != nil {
			return nil, failure
		}
		number, failure := p.version()
		if failure != nil {
			return nil, failure
		}
		version = &number
	}
	if !p.isSymbol("(") {
		if compound {
			return nil, p.callNameMissing()
		}
		return Name{Name: name}, nil
	}
	args, failure := p.sequence(")")
	if failure != nil {
		return nil, failure
	}
	return Call{Library: library, Name: name, Version: version, Args: args}, nil
}

func (p *parser) callNameMissing() *syntaxFailure {
	return &syntaxFailure{offset: p.current.start, code: codeBadSyntax, message: "a name must be followed by brackets to be a call"}
}

func (p *parser) version() (int, *syntaxFailure) {
	bad := &syntaxFailure{offset: p.current.start, code: codeBadSyntax, message: "a version is a whole number from 1 to 999999"}
	if p.current.kind != tokenInteger || p.current.start != p.previous.finish {
		return 0, bad
	}
	number, err := strconv.Atoi(p.current.text)
	if err != nil || number < 1 || number > 999999 {
		return 0, bad
	}
	if failure := p.advance(); failure != nil {
		return 0, failure
	}
	return number, nil
}

func treeDepth(n Node) int {
	deepest := 0
	for _, child := range n.Children() {
		if d := treeDepth(child); d > deepest {
			deepest = d
		}
	}
	return deepest + 1
}

func nodeCount(n Node) int {
	total := 1
	for _, child := range n.Children() {
		total += nodeCount(child)
	}
	return total
}

func joinNames(names []string) string {
	return strings.Join(names, ", ")
}
