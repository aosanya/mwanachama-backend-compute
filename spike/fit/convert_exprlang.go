package fit

import (
	"fmt"
	"strconv"

	"github.com/expr-lang/expr/ast"
)

func FromExprLang(raw any) (Node, error) {
	node, ok := raw.(ast.Node)
	if !ok {
		return Node{}, NotConvertible{"expr-lang", fmt.Sprintf("%T", raw)}
	}
	switch v := node.(type) {
	case *ast.IdentifierNode:
		return leaf(KindName, v.Value), nil
	case *ast.IntegerNode:
		return leaf(KindLiteral, strconv.Itoa(v.Value)), nil
	case *ast.FloatNode:
		return leaf(KindLiteral, strconv.FormatFloat(v.Value, 'f', -1, 64)+"f"), nil
	case *ast.StringNode:
		return leaf(KindLiteral, strconv.Quote(v.Value)), nil
	case *ast.BoolNode:
		return leaf(KindLiteral, strconv.FormatBool(v.Value)), nil
	case *ast.ArrayNode:
		return exprChildren(KindList, "", v.Nodes)
	case *ast.UnaryNode:
		return exprUnary(v)
	case *ast.BinaryNode:
		return exprBinary(v)
	case *ast.ConditionalNode:
		return exprConditional(v)
	case *ast.CallNode:
		return exprCall(v)
	case *ast.BuiltinNode:
		return exprChildren(KindCall, v.Name, v.Arguments)
	}
	return Node{}, NotConvertible{"expr-lang", fmt.Sprintf("%T", node)}
}

func exprUnary(v *ast.UnaryNode) (Node, error) {
	if !unaryOperators[v.Operator] {
		return Node{}, NotConvertible{"expr-lang", "operator " + v.Operator}
	}
	return exprChildren(KindUnary, v.Operator, []ast.Node{v.Node})
}

func exprBinary(v *ast.BinaryNode) (Node, error) {
	if !binaryOperators[v.Operator] {
		return Node{}, NotConvertible{"expr-lang", "operator " + v.Operator}
	}
	return exprChildren(KindBinary, v.Operator, []ast.Node{v.Left, v.Right})
}

func exprConditional(v *ast.ConditionalNode) (Node, error) {
	if !v.Ternary {
		return Node{}, NotConvertible{"expr-lang", "if block"}
	}
	return exprChildren(KindConditional, "?:", []ast.Node{v.Cond, v.Exp1, v.Exp2})
}

func exprCall(v *ast.CallNode) (Node, error) {
	callee, named := v.Callee.(*ast.IdentifierNode)
	if !named {
		return Node{}, NotConvertible{"expr-lang", fmt.Sprintf("callee %T", v.Callee)}
	}
	if text, isWrapper := decimalWrapper(callee, v.Arguments); isWrapper {
		return leaf(KindLiteral, text), nil
	}
	name := callee.Value
	if match := versionedName.FindStringSubmatch(name); match != nil {
		name = match[1] + "@" + match[2]
	}
	return exprChildren(KindCall, name, v.Arguments)
}

func decimalWrapper(callee *ast.IdentifierNode, args []ast.Node) (string, bool) {
	if callee.Value != "dec" || len(args) != 1 {
		return "", false
	}
	text, isText := args[0].(*ast.StringNode)
	if !isText {
		return "", false
	}
	return text.Value, true
}

func exprChildren(kind, text string, children []ast.Node) (Node, error) {
	nodes, err := convertAll(children, func(n ast.Node) (Node, error) { return FromExprLang(n) })
	if err != nil {
		return Node{}, err
	}
	return branch(kind, text, nodes...), nil
}
