package fit

import (
	"fmt"
	"regexp"
	"strconv"

	celast "cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/types"
)

var versionedName = regexp.MustCompile(`^(.+)_v(\d+)$`)

func FromCEL(raw any) (Node, error) {
	expr, ok := raw.(celast.Expr)
	if !ok {
		return Node{}, NotConvertible{"cel", fmt.Sprintf("%T", raw)}
	}
	switch expr.Kind() {
	case celast.LiteralKind:
		return celLiteral(expr)
	case celast.IdentKind:
		return leaf(KindName, expr.AsIdent()), nil
	case celast.ListKind:
		return celChildren(KindList, "", expr.AsList().Elements())
	case celast.CallKind:
		return celCall(expr)
	}
	return Node{}, NotConvertible{"cel", fmt.Sprintf("kind %d", expr.Kind())}
}

func celLiteral(expr celast.Expr) (Node, error) {
	switch v := expr.AsLiteral().(type) {
	case types.Int:
		return leaf(KindLiteral, strconv.FormatInt(int64(v), 10)), nil
	case types.Double:
		return leaf(KindLiteral, strconv.FormatFloat(float64(v), 'f', -1, 64)+"f"), nil
	case types.String:
		return leaf(KindLiteral, strconv.Quote(string(v))), nil
	case types.Bool:
		return leaf(KindLiteral, strconv.FormatBool(bool(v))), nil
	}
	return Node{}, NotConvertible{"cel", fmt.Sprintf("literal %T", expr.AsLiteral())}
}

func celCall(expr celast.Expr) (Node, error) {
	call := expr.AsCall()
	name, args := call.FunctionName(), call.Args()
	if call.IsMemberFunction() {
		return Node{}, NotConvertible{"cel", "member call " + name}
	}
	switch {
	case name == "decimal" && len(args) == 1 && args[0].Kind() == celast.LiteralKind:
		text, isText := args[0].AsLiteral().(types.String)
		if isText {
			return leaf(KindLiteral, string(text)), nil
		}
	case name == operators.Conditional:
		return celChildren(KindConditional, "?:", args)
	case name == operators.LogicalNot || name == operators.Negate:
		symbol, _ := operators.FindReverse(name)
		return celChildren(KindUnary, symbol, args)
	}
	if symbol, isOperator := operators.FindReverseBinaryOperator(name); isOperator && len(args) == 2 {
		if !binaryOperators[symbol] {
			return Node{}, NotConvertible{"cel", "operator " + symbol}
		}
		return celChildren(KindBinary, symbol, args)
	}
	if match := versionedName.FindStringSubmatch(name); match != nil {
		name = match[1] + "@" + match[2]
	}
	return celChildren(KindCall, name, args)
}

func celChildren(kind, text string, children []celast.Expr) (Node, error) {
	nodes, err := convertAll(children, func(e celast.Expr) (Node, error) { return FromCEL(e) })
	if err != nil {
		return Node{}, err
	}
	return branch(kind, text, nodes...), nil
}
