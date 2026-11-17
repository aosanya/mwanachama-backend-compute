package formula

import (
	"strings"
	"testing"
)

type parseCase struct {
	name     string
	text     string
	code     string
	sentence string
	line     int
	column   int
}

func TestParseRefusals(t *testing.T) {
	cases := []parseCase{
		{"bracket", `payment >= (balance ? 0.00 : balance - payment`, "bad-syntax", "The formula cannot be read: missing closing bracket at line 1, column 47.", 1, 47},
		{"leading-zero", `balance * 007`, "bad-syntax", "The formula cannot be read: a number cannot have a leading zero at line 1, column 11.", 1, 11},
		{"empty", ``, "bad-syntax", "The formula cannot be read: the formula is empty at line 1, column 1.", 1, 1},
		{"ends-after-operator", `a +`, "bad-syntax", "The formula cannot be read: the formula ends after an operator at line 1, column 4.", 1, 4},
		{"colon-missing", `a ? b`, "bad-syntax", "The formula cannot be read: a : is missing at line 1, column 6.", 1, 6},
		{"text-not-closed", `"abc`, "bad-syntax", "The formula cannot be read: text is not closed at line 1, column 1.", 1, 1},
		{"unknown-escape", `"a\q"`, "bad-syntax", "The formula cannot be read: text has an unknown escape at line 1, column 3.", 1, 3},
		{"escape-at-end", `"a\`, "bad-syntax", "The formula cannot be read: text ends inside an escape at line 1, column 3.", 1, 3},
		{"side-by-side", `a b`, "bad-syntax", "The formula cannot be read: two parts are side by side with no operator between them at line 1, column 3.", 1, 3},
		{"unexpected-character", `a # b`, "bad-syntax", "The formula cannot be read: unexpected character # at line 1, column 3.", 1, 3},
		{"unexpected-outside-language", "a ​ b", "bad-syntax", "The formula cannot be read: unexpected character U+200B at line 1, column 3.", 1, 3},
		{"control-character", "a \x01 b", "bad-syntax", "The formula cannot be read: unexpected control character at line 1, column 3.", 1, 3},
		{"second-line", "a +\n  # b", "bad-syntax", "The formula cannot be read: unexpected character # at line 2, column 3.", 2, 3},
		{"not-utf-8", "a + \xff", "bad-syntax", "The formula cannot be read: text is not valid UTF-8 at line 1, column 1.", 1, 1},
		{"dot-before-digits", `.5`, "bad-syntax", "The formula cannot be read: unexpected character . at line 1, column 1.", 1, 1},
		{"dot-after-digits", `5.`, "bad-syntax", "The formula cannot be read: unexpected character . at line 1, column 2.", 1, 2},
		{"name-needs-brackets", `fin.pmt`, "bad-syntax", "The formula cannot be read: a name must be followed by brackets to be a call at line 1, column 8.", 1, 8},
		{"version-not-a-number", `late_days@0(a, b)`, "bad-syntax", "The formula cannot be read: a version is a whole number from 1 to 999999 at line 1, column 11.", 1, 11},
		{"version-too-big", `late_days@1000000(a, b)`, "bad-syntax", "The formula cannot be read: a version is a whole number from 1 to 999999 at line 1, column 11.", 1, 11},
		{"space-before-at", `late_days @ 1(due, filed)`, "bad-syntax", "", 0, 0},
		{"space-around-dot", `fin . pmt@1(x)`, "bad-syntax", "", 0, 0},
		{"empty-brackets-of-call", `round()`, "bad-syntax", "", 0, 0},
		{"empty-list", `max([])`, "bad-syntax", "", 0, 0},
		{"depth", "!!!!!!!!!!!!suspended", "too-deep", "This formula goes over the depth limit of 12.", 0, 0},
		{"brackets-open", strings.Repeat("(", 13) + "a" + strings.Repeat(")", 13), "too-deep", "This formula goes over the depth limit of 12.", 0, 0},
		{"nodes", "sum([" + strings.TrimSuffix(strings.Repeat("1, ", 99), ", ") + "])", "too-many-nodes", "This formula goes over the node limit of 100.", 0, 0},
		{"text-length", `status == "` + strings.Repeat("x", 201) + `"`, "too-long", "This formula goes over the text length limit of 200.", 0, 0},
		{"formula-length", "x" + strings.Repeat("+x", 250), "too-long", "This formula goes over the formula length limit of 500.", 0, 0},
		{"integer-digits", `1234567890123456789 + kilos`, "too-many-digits", "This formula goes over the number magnitude limit of 18.", 0, 0},
		{"decimal-digits", `12345678901234567890123456789.50`, "too-many-digits", "This formula goes over the number magnitude limit of 30.", 0, 0},
		{"decimal-places", `balance * 0.12345678901`, "too-many-places", "This formula goes over the decimal places limit of 10.", 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, problems := Parse(c.text)
			if len(problems) != 1 {
				t.Fatalf("Parse gave %d problems: %+v", len(problems), problems)
			}
			p := problems[0]
			if p.Code != c.code {
				t.Fatalf("code %q, want %q (%s)", p.Code, c.code, p.Sentence)
			}
			if c.sentence != "" && p.Sentence != c.sentence {
				t.Fatalf("sentence %q, want %q", p.Sentence, c.sentence)
			}
			if c.line != 0 && (p.Line != c.line || p.Column != c.column) {
				t.Fatalf("position %d:%d, want %d:%d", p.Line, p.Column, c.line, c.column)
			}
			if p.Part != "" || p.Path != nil {
				t.Fatalf("a Parse problem has no part and a null path, got %q and %v", p.Part, p.Path)
			}
		})
	}
}

func TestParseAcceptsTheLimitsExactly(t *testing.T) {
	accepted := []string{
		"!!!!!!!!!!!suspended",
		strings.Repeat("(", 12) + "a" + strings.Repeat(")", 12),
		"sum([" + strings.TrimSuffix(strings.Repeat("1, ", 98), ", ") + "])",
		`status == "` + strings.Repeat("x", 200) + `"`,
		`123456789012345678 + kilos`,
		`1234567890123456789012345678.50`,
		`balance * 0.1234567890`,
		`late_days@999999(a, b)`,
		`round (x, 2)`,
	}
	for _, text := range accepted {
		if _, problems := Parse(text); len(problems) != 0 {
			t.Errorf("Parse(%.40q) gave %+v", text, problems)
		}
	}
}

func TestParseKeepsTheWrittenDigits(t *testing.T) {
	tree := mustParse(t, `0.010`)
	literal, ok := tree.Root.(Literal)
	if !ok || literal.Type != TypeDecimal || literal.Value != "0.010" {
		t.Fatalf("got %#v", tree.Root)
	}
}

func TestParseReadsTheTreeOfTheExample(t *testing.T) {
	tree := mustParse(t, `payment >= balance ? 0.00 : balance - payment`)
	want := Conditional{
		Condition: Binary{Op: ">=", Left: Name{Name: "payment"}, Right: Name{Name: "balance"}},
		Then:      Literal{Type: TypeDecimal, Value: "0.00"},
		Else:      Binary{Op: "-", Left: Name{Name: "balance"}, Right: Name{Name: "payment"}},
	}
	if got, ok := tree.Root.(Conditional); !ok || got != want {
		t.Fatalf("got %#v", tree.Root)
	}
}

func TestPartAtGivesThePathAndTheStandardText(t *testing.T) {
	tree := mustParse(t, `payment >= balance ? 0.00 : balance - payment`)
	cases := []struct {
		path Path
		text string
	}{
		{Path{}, `payment >= balance ? 0.00 : balance - payment`},
		{Path{0}, `payment >= balance`},
		{Path{2}, `balance - payment`},
		{Path{2, 0}, `balance`},
	}
	for _, c := range cases {
		part, ok := PartAt(tree, c.path)
		if !ok || part.Text != c.text {
			t.Errorf("PartAt(%v) = %+v, %v; want %q", c.path, part, ok, c.text)
		}
	}
	if _, ok := PartAt(tree, Path{3}); ok {
		t.Errorf("PartAt([3]) should not exist")
	}
}
