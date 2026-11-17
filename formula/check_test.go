package formula

import (
	"reflect"
	"strings"
	"testing"
)

type checkCase struct {
	name     string
	text     string
	inputs   []Input
	code     string
	part     string
	path     Path
	sentence string
}

func checkCases() []checkCase {
	manyNames := "a + b + c + d + e + f + g + h + i"
	costly := "sum([" + strings.TrimSuffix(strings.Repeat("sum(x), ", 21), ", ") + "])"
	return []checkCase{
		{"unknown-name", `balance * rate`, inputs("balance", "decimal", "months", "integer"), "unknown-name", "rate", Path{1}, "The name `rate` is not an input of this function. Its inputs are: balance, months."},
		{"unknown-function", `eval("1 + 1")`, nil, "unknown-function", `eval("1 + 1")`, Path{}, "The function eval is not known."},
		{"version-missing", `fin.pmt(12000, 0.01, 24)`, nil, "version-missing", `fin.pmt(12000, 0.01, 24)`, Path{}, "fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2."},
		{"version-not-found", `fin.pmt@7(12000, 0.01, 24)`, nil, "version-not-found", `fin.pmt@7(12000, 0.01, 24)`, Path{}, "`fin.pmt@7` does not exist. Latest is `fin.pmt@2`."},
		{"version-not-allowed", `round@1(balance, 2)`, inputs("balance", "decimal"), "version-not-allowed", `round@1(balance, 2)`, Path{}, "`round` is built in and takes no version."},
		{"wrong-argument-count", `round(balance)`, inputs("balance", "decimal"), "wrong-argument-count", `round(balance)`, Path{}, "`round` needs 2 arguments and this call has 1."},
		{"wrong-argument-type", `round(kilos, 2)`, inputs("kilos", "integer"), "wrong-argument-type", `kilos`, Path{0}, "Argument 1 of `round` must be a decimal or a float, and `kilos` is an integer."},
		{"type-mismatch", `status + 1`, inputs("status", "text"), "type-mismatch", `status + 1`, Path{}, "The part `status + 1` cannot be worked out: `+` needs two numbers, and this has text and integer."},
		{"condition-not-boolean", `balance ? 0.00 : 1.00`, inputs("balance", "decimal"), "condition-not-boolean", `balance`, Path{0}, "The condition `balance` must be true or false, and it is decimal."},
		{"branches-differ", `suspended ? "yes" : 0`, inputs("suspended", "boolean"), "branches-differ", `suspended ? "yes" : 0`, Path{}, "The two answers of `suspended ? \"yes\" : 0` must be the same kind, and they are text and integer."},
		{"float-decimal-mix", `balance - payment`, inputs("balance", "decimal", "payment", "float"), "float-decimal-mix", `balance - payment`, Path{}, "The part `balance - payment` mixes a decimal with a float; convert one with to_float or to_decimal."},
		{"list-items-differ", `max([0.00, "none"])`, nil, "list-items-differ", `[0.00, "none"]`, Path{0}, "The items of `[0.00, \"none\"]` must all be the same kind, and this list has decimal and text."},
		{"too-many-inputs", manyNames, inputs("a", "integer", "b", "integer", "c", "integer", "d", "integer", "e", "integer", "f", "integer", "g", "integer", "h", "integer", "i", "integer"), "too-many-inputs", manyNames, Path{}, "This formula goes over the input limit of 8."},
		{"too-costly", costly, inputs("x", "list of decimal"), "too-costly", costly, Path{}, "This formula goes over the cost limit of 20000."},
		{"call-too-deep", `step_e@1(amount)`, inputs("amount", "decimal"), "call-too-deep", `step_e@1(amount)`, Path{}, "This formula goes over the call depth limit of 4."},
		{"modulo-needs-integers", `balance % 2`, inputs("balance", "decimal"), "type-mismatch", `balance % 2`, Path{}, "The part `balance % 2` cannot be worked out: `%` needs two integers, and this has decimal and integer."},
		{"integer-division", `kilos / 2`, inputs("kilos", "integer"), "type-mismatch", `kilos / 2`, Path{}, "The part `kilos / 2` cannot be worked out: `/` needs a decimal or a float on at least one side, and this has integer and integer."},
		{"chained-comparison", `a < b < c`, inputs("a", "integer", "b", "integer", "c", "integer"), "type-mismatch", `a < b < c`, Path{}, "The part `a < b < c` cannot be worked out: `<` needs two numbers, and this has boolean and integer."},
		{"float-equality", `rate == rate`, inputs("rate", "float"), "type-mismatch", `rate == rate`, Path{}, "The part `rate == rate` cannot be worked out: `==` needs two numbers or two texts, and this has float and float."},
		{"not-needs-boolean", `!balance`, inputs("balance", "decimal"), "type-mismatch", `!balance`, Path{}, "The part `!balance` cannot be worked out: `!` needs a boolean, and this has decimal."},
		{"and-needs-booleans", `suspended && 1`, inputs("suspended", "boolean"), "type-mismatch", `suspended && 1`, Path{}, "The part `suspended && 1` cannot be worked out: `&&` needs two booleans, and this has boolean and integer."},
		{"one-problem-for-one-wrong-name", `round(rate * 2 + rate, 2)`, inputs("balance", "decimal"), "unknown-name", `rate`, Path{0, 0, 0}, "The name `rate` is not an input of this function. Its inputs are: balance."},
		{"library-argument-type", `fin.pmt@1(12000, 0.01, "24")`, nil, "wrong-argument-type", `"24"`, Path{2}, "Argument 3 of `fin.pmt@1` must be an integer, and `\"24\"` is a text."},
	}
}

func TestCheckCodes(t *testing.T) {
	library := testLibrary()
	for _, c := range checkCases() {
		t.Run(c.name, func(t *testing.T) {
			tree, problems := Parse(c.text)
			if len(problems) != 0 {
				t.Fatalf("Parse gave %+v", problems)
			}
			got := Check(tree, c.inputs, library)
			if c.name == "one-problem-for-one-wrong-name" {
				if len(got) == 0 || got[0].Code != c.code || got[0].Sentence != c.sentence || !reflect.DeepEqual(got[0].Path, c.path) {
					t.Fatalf("got %+v", got)
				}
				for _, p := range got {
					if p.Code != "unknown-name" {
						t.Fatalf("a wrong name must not add other problems, got %+v", got)
					}
				}
				return
			}
			if len(got) != 1 {
				t.Fatalf("Check gave %d problems: %+v", len(got), got)
			}
			p := got[0]
			if p.Code != c.code || p.Part != c.part || p.Sentence != c.sentence {
				t.Fatalf("got %+v", p)
			}
			if !reflect.DeepEqual(p.Path, c.path) {
				t.Fatalf("path %#v, want %#v", p.Path, c.path)
			}
			if p.Line != 0 || p.Column != 0 {
				t.Fatalf("a Check problem has no position, got %d:%d", p.Line, p.Column)
			}
		})
	}
}

func TestCheckAcceptsTheCallChainOfFour(t *testing.T) {
	tree := mustParse(t, `step_d@1(amount)`)
	if got := Check(tree, inputs("amount", "decimal"), testLibrary()); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestCheckCountsTheCostOfTheLateFilingFormula(t *testing.T) {
	tree := mustParse(t, examples[5].text)
	if got := Check(tree, examples[5].inputs, testLibrary()); len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
	if cost := Cost(tree, testLibrary()); cost != 58 {
		t.Fatalf("cost %d, want 58", cost)
	}
	termGrade := mustParse(t, examples[9].text)
	if cost := Cost(termGrade, testLibrary()); cost != 3019 {
		t.Fatalf("cost %d, want 3019", cost)
	}
}

func chain(depth int) Node {
	var node Node = Name{Name: "suspended"}
	for i := 0; i < depth; i++ {
		node = Unary{Op: "!", Operand: node}
	}
	return node
}

func TestCheckTestsTheLimitsOfATreeFromAnywhere(t *testing.T) {
	in := inputs("suspended", "boolean")
	long := List{}
	for i := 0; i < 99; i++ {
		long.Items = append(long.Items, Literal{Type: TypeInteger, Value: "1"})
	}
	cases := []struct {
		name     string
		tree     Tree
		code     string
		sentence string
	}{
		{"depth", Tree{Root: chain(12)}, "too-deep", "This formula goes over the depth limit of 12."},
		{"nodes", Tree{Root: Call{Name: "sum", Args: []Node{long}}}, "too-many-nodes", "This formula goes over the node limit of 100."},
		{"integer-digits", Tree{Root: Literal{Type: TypeInteger, Value: "1234567890123456789"}}, "too-many-digits", "This formula goes over the number magnitude limit of 18."},
		{"decimal-places", Tree{Root: Literal{Type: TypeDecimal, Value: "0.12345678901"}}, "too-many-places", "This formula goes over the decimal places limit of 10."},
		{"text-length", Tree{Root: Literal{Type: TypeText, Value: strings.Repeat("x", 201)}}, "too-long", "This formula goes over the text length limit of 200."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Check(c.tree, in, testLibrary())
			if len(got) != 1 || got[0].Code != c.code || got[0].Sentence != c.sentence {
				t.Fatalf("got %+v", got)
			}
			if got[0].Part != Format(c.tree) || !reflect.DeepEqual(got[0].Path, Path{}) {
				t.Fatalf("a limit problem names the whole formula, got part %.30q path %#v", got[0].Part, got[0].Path)
			}
		})
	}
}

func TestCheckRefusesATreeThatIsNotAFormula(t *testing.T) {
	cases := []struct {
		name     string
		tree     Tree
		path     Path
		sentence string
	}{
		{"operator", Tree{Root: Binary{Op: "^", Left: Name{Name: "balance"}, Right: Literal{Type: TypeInteger, Value: "2"}}}, Path{}, "The part at path [] is not a valid part: the operator ^ is not in the language."},
		{"missing-child", Tree{Root: Binary{Op: "+", Left: Name{Name: "balance"}}}, Path{}, "The part at path [] is not a valid part: the field right is missing."},
		{"bad-name", Tree{Root: Unary{Op: "!", Operand: Name{Name: "Balance"}}}, Path{0}, "The part at path [0] is not a valid part: the name Balance is not a valid name."},
		{"bad-value", Tree{Root: Call{Name: "round", Args: []Node{Name{Name: "x"}, Literal{Type: TypeInteger, Value: "2.5"}}}}, Path{1}, "The part at path [1] is not a valid part: the value 2.5 is not a integer."},
		{"no-tree", Tree{}, Path{}, "The part at path [] is not a valid part: the text is not a tree."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Check(c.tree, inputs("balance", "decimal", "x", "decimal"), testLibrary())
			if len(got) != 1 || got[0].Code != "bad-tree" || got[0].Sentence != c.sentence {
				t.Fatalf("got %+v", got)
			}
			if got[0].Part != "" || !reflect.DeepEqual(got[0].Path, c.path) {
				t.Fatalf("part %q path %#v", got[0].Part, got[0].Path)
			}
		})
	}
}

func TestCheckGivesAtMostTwentyProblems(t *testing.T) {
	tree := mustParse(t, "max(["+strings.TrimSuffix(strings.Repeat("rate, ", 30), ", ")+"])")
	got := Check(tree, nil, testLibrary())
	if len(got) != 20 {
		t.Fatalf("got %d problems, want 20", len(got))
	}
}

func TestCheckNeedsNoNumbers(t *testing.T) {
	tree := mustParse(t, `payment >= balance ? 0.00 : balance - payment`)
	got := Check(tree, inputs("payment", "decimal", "balance", "decimal"), nil)
	if len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestEveryCodeOfTheLanguagePageHasACase(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range checkCases() {
		seen[c.code] = true
	}
	for _, code := range []string{"unknown-name", "unknown-function", "version-missing", "version-not-found", "version-not-allowed", "wrong-argument-count", "wrong-argument-type", "type-mismatch", "condition-not-boolean", "branches-differ", "float-decimal-mix", "list-items-differ", "too-many-inputs", "too-costly", "call-too-deep"} {
		if !seen[code] {
			t.Errorf("no Check case for %s", code)
		}
	}
}

func TestCheckRefusesACallWithNoArguments(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		code     string
		sentence string
	}{
		{"unknown-now", `now()`, "unknown-function", "The function now is not known."},
		{"unknown-random", `random()`, "unknown-function", "The function random is not known."},
		{"unknown-in-library", `fin.today@1()`, "unknown-function", "The function fin.today is not known."},
		{"builtin-without-arguments", `round()`, "wrong-argument-count", "`round` needs 2 arguments and this call has 0."},
		{"one-argument-builtin-without-arguments", `sum()`, "wrong-argument-count", "`sum` needs 1 argument and this call has 0."},
		{"library-without-arguments", `late_days@1()`, "wrong-argument-count", "`late_days@1` needs 2 arguments and this call has 0."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Check(mustParse(t, c.text), nil, testLibrary())
			if len(got) != 1 {
				t.Fatalf("Check gave %d problems: %+v", len(got), got)
			}
			if got[0].Code != c.code || got[0].Sentence != c.sentence || got[0].Part != c.text || !reflect.DeepEqual(got[0].Path, Path{}) {
				t.Fatalf("got %+v", got[0])
			}
		})
	}
}

func TestCheckRefusesAnEmptyListStillAsABadTree(t *testing.T) {
	got := Check(Tree{Root: List{}}, nil, testLibrary())
	if len(got) != 1 || got[0].Code != "bad-tree" {
		t.Fatalf("got %+v", got)
	}
}
