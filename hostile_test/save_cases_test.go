package hostile

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type saveCase struct {
	id       string
	text     string
	inputs   []formula.Input
	code     string
	sentence string
	prefix   string
	known    string
}

func cannotBeRead(what string, line int, column int) string {
	return "The formula cannot be read: " + what + " at line " + itoa(line) + ", column " + itoa(column) + "."
}

func unknownFunction(name string) string {
	return "The function " + name + " is not known."
}

var booleanInput = []formula.Input{{Name: "flag", Type: formula.TypeBoolean}}

var rowsInput = []formula.Input{{Name: "rows", Type: formula.TypeListDecimal}}

var totalInput = []formula.Input{{Name: "total", Type: formula.TypeDecimal}}

var parserCases = []saveCase{
	{id: "P01", text: "qz(( ,, 4 4 ;;", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "P02", text: "1 + \xff\xfe\xc0", code: "bad-syntax", prefix: "The formula cannot be read: text is not valid UTF-8 at line "},
	{id: "P03", text: "1 + \x00 2", code: "bad-syntax", sentence: cannotBeRead("unexpected control character", 1, 5)},
	{id: "P04", text: nested(13, "(", ")", "1"), code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "P04", text: nested(240, "(", ")", "1"), code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "P05", text: strings.Repeat("(", 8) + "1", code: "bad-syntax", sentence: cannotBeRead("missing closing bracket", 1, 10)},
	{id: "P05", text: strings.Repeat("(", 300), code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "P05", text: strings.Repeat("(", 8), code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "P06", text: nested(13, "[", "]", "1"), code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "P06", text: nested(240, "[", "]", "1"), code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "P07", text: strings.Repeat("a", 600) + " + 1", code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "P08", text: strings.TrimSuffix(strings.Repeat(strings.Repeat("a", 40)+" + ", 1000), " + "), code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "P09", text: strings.Repeat("1", 19), code: "too-many-digits", sentence: limitSentence("number magnitude", 18)},
	{id: "P09", text: strings.Repeat("1", 600), code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "P10", text: "0." + strings.Repeat("1", 11), code: "too-many-places", sentence: limitSentence("decimal places", 10)},
	{id: "P10", text: "0." + strings.Repeat("1", 600), code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "P11", text: "tot​al + 1", code: "bad-syntax", sentence: cannotBeRead("unexpected character U+200B", 1, 4)},
	{id: "P12", text: "‮1 + 2", code: "bad-syntax", sentence: cannotBeRead("unexpected character U+202E", 1, 1)},
	{id: "P13", text: "\U0001F600 * 1", code: "bad-syntax", prefix: "The formula cannot be read: unexpected character"},
	{id: "P14", text: "1 +\a 2", code: "bad-syntax", sentence: cannotBeRead("unexpected control character", 1, 4)},
	{id: "P14", text: "1 +\r2", code: "bad-syntax", sentence: cannotBeRead("unexpected control character", 1, 4)},
	{id: "P15", text: "sum(price, quantity", code: "bad-syntax", sentence: cannotBeRead("missing closing bracket", 1, 20)},
	{id: "P16", text: "price * quantity +", code: "bad-syntax", sentence: cannotBeRead("the formula ends after an operator", 1, 19)},
	{id: "P17", text: `"hello`, code: "bad-syntax", sentence: cannotBeRead("text is not closed", 1, 1)},
	{id: "P18", text: `"hello\`, code: "bad-syntax", sentence: cannotBeRead("text ends inside an escape", 1, 7)},
	{id: "P19", text: "1 +\n2 +\n3 4", code: "bad-syntax", sentence: cannotBeRead("two parts are side by side with no operator between them", 3, 3)},
	{id: "P20", text: "", code: "bad-syntax", sentence: cannotBeRead("the formula is empty", 1, 1)},
	{id: "P20", text: "     ", code: "bad-syntax", sentence: cannotBeRead("the formula is empty", 1, 1)},
}

var limitCases = []saveCase{
	{id: "L01", text: nested(12, "!(", ")", "flag"), inputs: booleanInput, code: "too-deep", sentence: limitSentence("depth", 12)},
	{id: "L03", text: manyItems(99), code: "too-many-nodes", sentence: limitSentence("node", 100)},
	{id: "L05", text: "sum(rows) + " + balancedSum(31, "sum(rows)"), inputs: rowsInput, code: "too-costly", sentence: limitSentence("cost", 20000)},
	{id: "L05", text: balancedSum(32, "sum(rows)"), inputs: rowsInput, code: "too-costly", sentence: limitSentence("cost", 20000)},
}

var costCases = []saveCase{
	{id: "C01", text: "range(1, 1000000000)", code: "unknown-function", sentence: unknownFunction("range")},
	{id: "C03", text: "map(range(1, 1000), x -> range(1, 1000))", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "C03", text: balancedSum(21, "sum(rows)"), inputs: rowsInput, code: "too-costly", sentence: limitSentence("cost", 20000)},
	{id: "C05", text: "sum(map(a, x -> sum(map(b, y -> sum(map(c, z -> 1))))))", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "C05", text: balancedSum(24, "sum(rows)"), inputs: rowsInput, code: "too-costly", sentence: limitSentence("cost", 20000)},
	{id: "C07", text: "step_e@1(total)", inputs: totalInput, code: "call-too-deep", sentence: limitSentence("call depth", 4)},
	{id: "C09", text: "f@1(total) + 1", inputs: totalInput, code: "unknown-function", sentence: unknownFunction("f")},
	{id: "C11", text: "a@1(total)", inputs: totalInput, code: "unknown-function", sentence: unknownFunction("a")},
}

var sizeCases = []saveCase{
	{id: "I01", text: `"` + strings.Repeat("a", 201) + `"`, code: "too-long", sentence: limitSentence("text length", 200)},
	{id: "I03", text: `repeat("ab", 1000000000)`, code: "unknown-function", sentence: unknownFunction("repeat")},
	{id: "I04", text: strings.Repeat("1", 1000000), code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "I06", text: "1e1000000000", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "I07", text: "2 ^ (2 ^ (2 ^ 10))", code: "bad-syntax", prefix: "The formula cannot be read: unexpected character"},
	{id: "I08", text: "fold(range(1, 100000), 1, (a, x) -> a * x)", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "I09", text: "0." + strings.Repeat("1", 1000000), code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "I09", text: "0." + strings.Repeat("1", 11), code: "too-many-places", sentence: limitSentence("decimal places", 10)},
	{id: "I10", text: "round(1.00 / 3, 1000000)", code: "wrong-argument-type", prefix: "Argument 2 of `round` must be"},
	{id: "I11", text: "sum([" + strings.TrimSuffix(strings.Repeat("1, ", 1000000), ", ") + "])", code: "too-long", sentence: limitSentence("formula length", 500)},
	{id: "I11", text: manyItems(100), code: "too-many-nodes", sentence: limitSentence("node", 100)},
}

var reachCases = []saveCase{
	{id: "O01", text: `read_file("/etc/passwd")`, code: "unknown-function", sentence: unknownFunction("read_file")},
	{id: "O02", text: "now()", code: "unknown-function", sentence: unknownFunction("now"), known: "ISS-74"},
	{id: "O03", text: "random()", code: "unknown-function", sentence: unknownFunction("random"), known: "ISS-74"},
	{id: "O04", text: `http_get("http://example.invalid/")`, code: "unknown-function", sentence: unknownFunction("http_get")},
	{id: "O05", text: `env("HOME")`, code: "unknown-function", sentence: unknownFunction("env")},
	{id: "O06", text: `exec("ls")`, code: "unknown-function", sentence: unknownFunction("exec")},
	{id: "O07", text: `import("other")`, code: "unknown-function", sentence: unknownFunction("import")},
	{id: "O07", text: `eval("1 + 1")`, code: "unknown-function", sentence: unknownFunction("eval")},
	{id: "O08", text: "/etc/passwd", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "O08", text: "file:///x", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "O09", text: "[1, 2, 3][99]", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "O10", text: "[1, 2, 3][-1]", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "O11", text: "[1, 2, 3][1.5]", code: "bad-syntax", prefix: "The formula cannot be read: "},
	{id: "O12", text: "repeat([0], 1000000000)", code: "unknown-function", sentence: unknownFunction("repeat")},
	{id: "O13", text: "__proto__", code: "bad-syntax", prefix: "The formula cannot be read: unexpected character"},
	{id: "O13", text: "constructor", inputs: totalInput, code: "unknown-name", prefix: "The name `constructor` is not an input of this function."},
	{id: "O13", text: "os", inputs: totalInput, code: "unknown-name", prefix: "The name `os` is not an input of this function."},
	{id: "O15", text: "sort(rows, x -> x.group)", inputs: rowsInput, code: "bad-syntax", prefix: "The formula cannot be read: "},
}

func allSaveCases() []saveCase {
	var all []saveCase
	for _, group := range [][]saveCase{parserCases, limitCases, costCases, sizeCases, reachCases} {
		all = append(all, group...)
	}
	return all
}

func checkSave(c saveCase) string {
	problems := atSave(c.text, c.inputs)
	if len(problems) == 0 {
		return "the formula was accepted"
	}
	first := problems[0]
	if !plainSentence.MatchString(first.Sentence) {
		return "sentence is not one plain line: " + first.Sentence
	}
	if first.Code != c.code {
		return "code " + first.Code + ", want " + c.code + " (" + first.Sentence + ")"
	}
	if c.sentence != "" && first.Sentence != c.sentence {
		return "sentence " + first.Sentence + ", want " + c.sentence
	}
	if c.prefix != "" && !strings.HasPrefix(first.Sentence, c.prefix) {
		return "sentence " + first.Sentence + ", want it to start with " + c.prefix
	}
	return ""
}

func settle(t *testing.T, id string, known string, failure string) {
	t.Helper()
	switch {
	case failure == "" && known != "":
		t.Fatalf("%s passes now: remove it from the known failures of %s", id, known)
	case failure != "" && known != "":
		t.Skipf("%s: known failure of %s: %s", id, known, failure)
	case failure != "":
		t.Fatalf("%s: %s", id, failure)
	}
}

func runSaveCases(t *testing.T, cases []saveCase) {
	for index, c := range cases {
		c := c
		t.Run(c.id+"/"+itoa(index), func(t *testing.T) {
			settle(t, c.id, c.known, checkSave(c))
		})
	}
}

func TestParserFuzzingCases(t *testing.T) { runSaveCases(t, parserCases) }

func TestLimitCases(t *testing.T) { runSaveCases(t, limitCases) }

func TestCostTrickCases(t *testing.T) { runSaveCases(t, costCases) }

func TestSizeOfInputCases(t *testing.T) { runSaveCases(t, sizeCases) }

func TestNoOutsideReachCases(t *testing.T) { runSaveCases(t, reachCases) }
