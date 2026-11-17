package fit

import (
	"fmt"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
)

func FromOwn(raw any) (Node, error) {
	node, ok := raw.(own.Node)
	if !ok {
		return Node{}, NotConvertible{"own", fmt.Sprintf("%T", raw)}
	}
	switch v := node.(type) {
	case own.IntegerLiteral:
		return leaf(KindLiteral, strconv.FormatInt(v.Value, 10)), nil
	case own.DecimalLiteral:
		return leaf(KindLiteral, v.Value.String()), nil
	case own.FloatLiteral:
		return leaf(KindLiteral, strconv.FormatFloat(v.Value, 'f', -1, 64)+"f"), nil
	case own.TextLiteral:
		return leaf(KindLiteral, strconv.Quote(v.Value)), nil
	case own.DateLiteral:
		return leaf(KindLiteral, v.Value.Format("2006-01-02")), nil
	case own.BooleanLiteral:
		return leaf(KindLiteral, strconv.FormatBool(v.Value)), nil
	case own.Variable:
		return leaf(KindName, v.Name), nil
	case own.List:
		return ownBranch(KindList, "", v.Elements)
	case own.Call:
		return ownBranch(KindCall, v.Name, v.Args)
	case own.Unary:
		return ownBranch(KindUnary, v.Op, []own.Node{v.Operand})
	case own.Binary:
		return ownBranch(KindBinary, v.Op, []own.Node{v.Left, v.Right})
	case own.Conditional:
		return ownBranch(KindConditional, "?:", []own.Node{v.Test, v.Then, v.Else})
	}
	return Node{}, NotConvertible{"own", fmt.Sprintf("%T", node)}
}

func ownBranch(kind, text string, children []own.Node) (Node, error) {
	nodes, err := convertAll(children, func(n own.Node) (Node, error) { return FromOwn(n) })
	if err != nil {
		return Node{}, err
	}
	return branch(kind, text, nodes...), nil
}
