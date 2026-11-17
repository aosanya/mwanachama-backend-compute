package formula

import (
	"strconv"
	"strings"
)

func Format(tree Tree) string {
	var out strings.Builder
	writeNode(&out, tree.Root)
	return out.String()
}

func levelOf(n Node) int {
	switch node := n.(type) {
	case Unary:
		return levelUnary
	case Binary:
		if op, known := binaryOperators[node.Op]; known {
			return op.level
		}
		return levelOr
	case Conditional:
		return levelConditional
	}
	return levelAtom
}

func writeBracketed(out *strings.Builder, n Node, brackets bool) {
	if brackets {
		out.WriteByte('(')
	}
	writeNode(out, n)
	if brackets {
		out.WriteByte(')')
	}
}

func isConditional(n Node) bool {
	_, yes := n.(Conditional)
	return yes
}

func writeNode(out *strings.Builder, n Node) {
	switch node := n.(type) {
	case nil:
	case Literal:
		writeLiteral(out, node)
	case Name:
		out.WriteString(node.Name)
	case Unary:
		out.WriteString(node.Op)
		operandLevel := levelOf(node.Operand)
		writeBracketed(out, node.Operand, operandLevel < levelUnary)
	case Binary:
		own := levelOf(node)
		writeBracketed(out, node.Left, levelOf(node.Left) < own)
		out.WriteString(" " + node.Op + " ")
		writeBracketed(out, node.Right, levelOf(node.Right) <= own)
	case Conditional:
		writeBracketed(out, node.Condition, isConditional(node.Condition))
		out.WriteString(" ? ")
		writeBracketed(out, node.Then, isConditional(node.Then))
		out.WriteString(" : ")
		writeNode(out, node.Else)
	case Call:
		writeCall(out, node)
	case List:
		out.WriteByte('[')
		writeSequence(out, node.Items)
		out.WriteByte(']')
	}
}

func writeSequence(out *strings.Builder, items []Node) {
	for i, item := range items {
		if i > 0 {
			out.WriteString(", ")
		}
		writeNode(out, item)
	}
}

func writeCall(out *strings.Builder, call Call) {
	if call.Library != nil {
		out.WriteString(*call.Library)
		out.WriteByte('.')
	}
	out.WriteString(call.Name)
	if call.Version != nil {
		out.WriteByte('@')
		out.WriteString(strconv.Itoa(*call.Version))
	}
	out.WriteByte('(')
	writeSequence(out, call.Args)
	out.WriteByte(')')
}

func writeLiteral(out *strings.Builder, literal Literal) {
	if literal.Type != TypeText {
		out.WriteString(literal.Value)
		return
	}
	out.WriteByte('"')
	for _, r := range literal.Value {
		if r == '"' || r == '\\' {
			out.WriteByte('\\')
		}
		out.WriteRune(r)
	}
	out.WriteByte('"')
}
