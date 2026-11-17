package own

import (
	"strconv"
	"strings"
)

const atomLevel = 100

func Format(n Node) string {
	switch v := n.(type) {
	case IntegerLiteral:
		return strconv.FormatInt(v.Value, 10)
	case DecimalLiteral:
		return v.Value.String()
	case FloatLiteral:
		return strconv.FormatFloat(v.Value, 'f', -1, 64) + "f"
	case TextLiteral:
		return quoteText(v.Value)
	case DateLiteral:
		return v.Value.Format(dateLayout)
	case BooleanLiteral:
		return strconv.FormatBool(v.Value)
	case Variable:
		return v.Name
	case List:
		return "[" + formatAll(v.Elements) + "]"
	case Call:
		return v.Name + "(" + formatAll(v.Args) + ")"
	case Unary:
		return v.Op + bracketedIf(isOperation(v.Operand), v.Operand)
	case Binary:
		return formatBinary(v)
	case Conditional:
		return formatConditional(v)
	}
	return ""
}

func formatBinary(b Binary) string {
	level := precedence[b.Op]
	left := bracketedIf(operandLevel(b.Left) < level, b.Left)
	right := bracketedIf(operandLevel(b.Right) <= level, b.Right)
	return left + " " + b.Op + " " + right
}

func formatConditional(c Conditional) string {
	test := bracketedIf(isConditional(c.Test), c.Test)
	then := bracketedIf(isConditional(c.Then), c.Then)
	return test + " ? " + then + " : " + Format(c.Else)
}

func formatAll(nodes []Node) string {
	parts := make([]string, len(nodes))
	for i, node := range nodes {
		parts[i] = Format(node)
	}
	return strings.Join(parts, ", ")
}

func bracketedIf(condition bool, n Node) string {
	if condition {
		return "(" + Format(n) + ")"
	}
	return Format(n)
}

func operandLevel(n Node) int {
	switch v := n.(type) {
	case Binary:
		return precedence[v.Op]
	case Conditional:
		return 0
	}
	return atomLevel
}

func isOperation(n Node) bool {
	switch n.(type) {
	case Binary, Conditional:
		return true
	}
	return false
}

func isConditional(n Node) bool {
	_, ok := n.(Conditional)
	return ok
}

func quoteText(value string) string {
	escaper := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + escaper.Replace(value) + `"`
}
