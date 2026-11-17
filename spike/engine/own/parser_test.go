package own

import (
	"strconv"
	"strings"
	"testing"
)

func shape(n Node) string {
	switch v := n.(type) {
	case IntegerLiteral:
		return strconv.FormatInt(v.Value, 10)
	case DecimalLiteral:
		return v.Value.String()
	case FloatLiteral:
		return strconv.FormatFloat(v.Value, 'g', -1, 64) + "f"
	case TextLiteral:
		return strconv.Quote(v.Value)
	case DateLiteral:
		return v.Value.Format(dateLayout)
	case BooleanLiteral:
		return strconv.FormatBool(v.Value)
	case Variable:
		return v.Name
	case List:
		return "[" + shapes(v.Elements) + "]"
	case Unary:
		return "(" + v.Op + " " + shape(v.Operand) + ")"
	case Binary:
		return "(" + v.Op + " " + shape(v.Left) + " " + shape(v.Right) + ")"
	case Conditional:
		return "(? " + shape(v.Test) + " " + shape(v.Then) + " " + shape(v.Else) + ")"
	case Call:
		return v.Name + "(" + shapes(v.Args) + ")"
	}
	return "?"
}

func shapes(nodes []Node) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = shape(n)
	}
	return strings.Join(parts, ", ")
}

func TestParserGroupsByPrecedenceAndAssociativity(t *testing.T) {
	cases := []struct{ source, want string }{
		{"1 + 2 * 3", "(+ 1 (* 2 3))"},
		{"(1 + 2) * 3", "(* (+ 1 2) 3)"},
		{"a - b - c", "(- (- a b) c)"},
		{"a / b * c % d", "(% (* (/ a b) c) d)"},
		{"!a && b", "(&& (! a) b)"},
		{"!(a && b)", "(! (&& a b))"},
		{"-a * b", "(* (- a) b)"},
		{"- - a", "(- (- a))"},
		{"a || b && c", "(|| a (&& b c))"},
		{"a && b || c && d", "(|| (&& a b) (&& c d))"},
		{"a == b && c != d", "(&& (== a b) (!= c d))"},
		{"a + 1 < b * 2", "(< (+ a 1) (* b 2))"},
		{"a < b == c", "(== (< a b) c)"},
		{"a ? 1 : b ? 2 : 3", "(? a 1 (? b 2 3))"},
		{"a ? b ? 1 : 2 : 3", "(? a (? b 1 2) 3)"},
		{"a || b ? 1 + 2 : 3 * 4", "(? (|| a b) (+ 1 2) (* 3 4))"},
		{"f(a, 1 + 2)", "f(a, (+ 1 2))"},
		{"f()", "f()"},
		{"late_days@1(due, filed) == 0", "(== late_days@1(due, filed) 0)"},
		{"max([0, x]) + 1", "(+ max([0, x]) 1)"},
		{`"+" == s`, `(== "+" s)`},
		{"true && false", "(&& true false)"},
		{"0.75f * 2.50", "(* 0.75f 2.50)"},
		{"d < 2026-03-31", "(< d 2026-03-31)"},
	}
	for _, c := range cases {
		tree, err := Parse(c.source)
		if err != nil {
			t.Errorf("%q: %v", c.source, err)
			continue
		}
		if got := shape(tree); got != c.want {
			t.Errorf("%q: got %s, want %s", c.source, got, c.want)
		}
	}
}

func TestParserRefusesWithLineAndColumn(t *testing.T) {
	cases := []struct{ source, message string }{
		{"1 +", "The formula cannot be read: expected a value but found the end of the formula at line 1, column 4."},
		{"", "The formula cannot be read: expected a value but found the end of the formula at line 1, column 1."},
		{"a +\n  * b", `The formula cannot be read: expected a value but found "*" at line 2, column 3.`},
		{"(1 + 2", `The formula cannot be read: expected ")" but found the end of the formula at line 1, column 7.`},
		{"1 2", `The formula cannot be read: expected the end of the formula but found "2" at line 1, column 3.`},
		{"a ? 1", `The formula cannot be read: expected ":" but found the end of the formula at line 1, column 6.`},
		{"a ? : 2", `The formula cannot be read: expected a value but found ":" at line 1, column 5.`},
		{"f(1,", "The formula cannot be read: expected a value but found the end of the formula at line 1, column 5."},
		{"f(1 2)", `The formula cannot be read: expected ")" but found "2" at line 1, column 5.`},
		{"[]", "The formula cannot be read: a list needs at least one value at line 1, column 2."},
		{"[1,]", `The formula cannot be read: expected a value but found "]" at line 1, column 4.`},
		{"[1 2]", `The formula cannot be read: expected "]" but found "2" at line 1, column 4.`},
		{"late@1", `The formula cannot be read: a versioned name must be followed by "(" at line 1, column 7.`},
		{"99999999999999999999", "The formula cannot be read: the number 99999999999999999999 is too large at line 1, column 1."},
		{"2026-02-30", "The formula cannot be read: 2026-02-30 is not a real date at line 1, column 1."},
		{"a +\n\n  )", `The formula cannot be read: expected a value but found ")" at line 3, column 3.`},
		{`1 + "x`, "The formula cannot be read: the text is not closed at line 1, column 5."},
	}
	for _, c := range cases {
		_, err := Parse(c.source)
		if err == nil || err.Error() != c.message {
			t.Errorf("%q:\n got  %v\n want %s", c.source, err, c.message)
		}
	}
}

func TestNodesCarryTheirPosition(t *testing.T) {
	tree, err := Parse("a +\n  b * c")
	if err != nil {
		t.Fatal(err)
	}
	plus := tree.(Binary)
	times := plus.Right.(Binary)
	if plus.Position() != (Pos{1, 3}) || times.Position() != (Pos{2, 5}) || times.Left.Position() != (Pos{2, 3}) {
		t.Errorf("positions are %v %v %v", plus.Position(), times.Position(), times.Left.Position())
	}
}

func TestDepthAndCountWalkTheTree(t *testing.T) {
	cases := []struct {
		source       string
		depth, count int
	}{
		{"1", 1, 1},
		{"1 + 2 * 3", 3, 5},
		{"a ? 1 : b ? 2 : 3", 3, 7},
		{"max([0, f(x, 1)])", 4, 6},
		{"-(-(-a))", 4, 4},
	}
	for _, c := range cases {
		tree, err := Parse(c.source)
		if err != nil {
			t.Fatal(err)
		}
		if Depth(tree) != c.depth || Count(tree) != c.count {
			t.Errorf("%q: depth %d count %d, want %d and %d", c.source, Depth(tree), Count(tree), c.depth, c.count)
		}
	}
}
