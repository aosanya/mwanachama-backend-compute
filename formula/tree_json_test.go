package formula

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func compacted(t *testing.T, text string) string {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, []byte(text)); err != nil {
		t.Fatalf("the test JSON is not valid: %v", err)
	}
	return out.String()
}

func TestTheNineteenExamplesRoundTripThroughJSON(t *testing.T) {
	for _, ex := range examples {
		ex := ex
		t.Run(ex.name, func(t *testing.T) {
			tree := mustParse(t, ex.text)
			data := TreeToJSON(tree)
			back, problems := TreeFromJSON(data)
			if len(problems) != 0 {
				t.Fatalf("TreeFromJSON gave %+v for %s", problems, data)
			}
			if !reflect.DeepEqual(back, tree) {
				t.Fatalf("the tree changed\nwant: %#v\ngot:  %#v", tree, back)
			}
			if Format(back) != Format(tree) {
				t.Fatalf("Format changed: %q and %q", Format(tree), Format(back))
			}
			if again := TreeToJSON(back); !bytes.Equal(again, data) {
				t.Fatalf("writing again changed the JSON\nfirst:  %s\nsecond: %s", data, again)
			}
			if problems := Check(back, ex.inputs, testLibrary()); len(problems) != 0 {
				t.Fatalf("Check gave %+v on a tree read from JSON", problems)
			}
		})
	}
}

func TestTheWorkedTreeOfThePageIsWrittenAsThePageWritesIt(t *testing.T) {
	want := compacted(t, `{
  "kind": "conditional",
  "condition": {
    "kind": "binary",
    "op": ">=",
    "left": {"kind": "name", "name": "payment"},
    "right": {"kind": "name", "name": "balance"}
  },
  "then": {"kind": "literal", "type": "decimal", "value": "0.00"},
  "else": {
    "kind": "binary",
    "op": "-",
    "left": {"kind": "name", "name": "balance"},
    "right": {"kind": "name", "name": "payment"}
  }
}`)
	tree := mustParse(t, `payment >= balance ? 0.00 : balance - payment`)
	if got := string(TreeToJSON(tree)); got != want {
		t.Fatalf("JSON differs\nwant: %s\ngot:  %s", want, got)
	}
	back, problems := TreeFromJSON([]byte(want))
	if len(problems) != 0 || !reflect.DeepEqual(back, tree) {
		t.Fatalf("the page's tree is not read back: %+v %+v", problems, back)
	}
}

func TestFieldsAreWrittenInTheOrderOfTheTable(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{`!ready`, `{"kind":"unary","op":"!","operand":{"kind":"name","name":"ready"}}`},
		{`round(x, 2)`, `{"kind":"call","library":null,"name":"round","version":null,"args":[{"kind":"name","name":"x"},{"kind":"literal","type":"integer","value":"2"}]}`},
		{`fin.pmt@2(a)`, `{"kind":"call","library":"fin","name":"pmt","version":2,"args":[{"kind":"name","name":"a"}]}`},
		{`late_days@1(due, filed)`, `{"kind":"call","library":null,"name":"late_days","version":1,"args":[{"kind":"name","name":"due"},{"kind":"name","name":"filed"}]}`},
		{`[1, 2]`, `{"kind":"list","items":[{"kind":"literal","type":"integer","value":"1"},{"kind":"literal","type":"integer","value":"2"}]}`},
		{`now()`, `{"kind":"call","library":null,"name":"now","version":null,"args":[]}`},
		{`flag == true`, `{"kind":"binary","op":"==","left":{"kind":"name","name":"flag"},"right":{"kind":"literal","type":"boolean","value":"true"}}`},
	}
	for _, c := range cases {
		tree := mustParse(t, c.text)
		if got := string(TreeToJSON(tree)); got != c.want {
			t.Errorf("%s\nwant: %s\ngot:  %s", c.text, c.want, got)
		}
		back, problems := TreeFromJSON([]byte(c.want))
		if len(problems) != 0 || !reflect.DeepEqual(back, tree) {
			t.Errorf("%s is not read back: %+v %#v", c.text, problems, back)
		}
	}
}

func TestTextIsKeptWhateverItHolds(t *testing.T) {
	for _, value := range []string{`a<b>&c`, `say "hi" \ now`, "café   end", ``, `  spaced  `} {
		tree := Tree{Root: Binary{Op: "==", Left: Name{Name: "label"}, Right: Literal{Type: TypeText, Value: value}}}
		data := TreeToJSON(tree)
		if strings.Contains(string(data), `<`) || strings.Contains(string(data), `&`) {
			t.Errorf("markup characters are escaped in %s", data)
		}
		back, problems := TreeFromJSON(data)
		if len(problems) != 0 || !reflect.DeepEqual(back, tree) {
			t.Errorf("text %q is not kept: %+v %#v", value, problems, back)
		}
	}
}

func TestAnEmptyTreeIsNull(t *testing.T) {
	if got := string(TreeToJSON(Tree{})); got != "null" {
		t.Fatalf("an empty tree is written as %q", got)
	}
}

type treeJSONCase struct {
	id       string
	data     string
	problems []Problem
}

func badTree(path Path, what string) Problem {
	return Problem{
		Code:     "bad-tree",
		Sentence: "The part at path " + pathText(path) + " is not a valid part: " + what + ".",
		Path:     path,
	}
}

func TestATreeThatIsNotAFormulaIsRefusedWithItsPath(t *testing.T) {
	cases := []treeJSONCase{
		{"page-example-operator", `{"kind": "binary", "op": "^", "left": {"kind": "name", "name": "balance"}, "right": {"kind": "literal", "type": "integer", "value": "2"}}`,
			[]Problem{badTree(Path{}, "the operator ^ is not in the language")}},
		{"unknown-kind", `{"kind": "loop"}`, []Problem{badTree(Path{}, "the kind loop is not in the language")}},
		{"kind-not-text", `{"kind": 7}`, []Problem{badTree(Path{}, "the kind 7 is not in the language")}},
		{"kind-missing", `{"name": "x"}`, []Problem{badTree(Path{}, "the field kind is missing")}},
		{"field-missing", `{"kind": "name"}`, []Problem{badTree(Path{}, "the field name is missing")}},
		{"fields-missing-in-table-order", `{"kind": "literal"}`, []Problem{
			badTree(Path{}, "the field type is missing"),
			badTree(Path{}, "the field value is missing"),
		}},
		{"field-not-allowed", `{"kind": "name", "name": "x", "colour": "red", "bold": true}`, []Problem{
			badTree(Path{}, "the field bold is not allowed"),
			badTree(Path{}, "the field colour is not allowed"),
		}},
		{"nested-literal", `{"kind": "binary", "op": "+", "left": {"kind": "name", "name": "a"}, "right": {"kind": "literal", "type": "decimal", "value": "1.5.0"}}`,
			[]Problem{badTree(Path{1}, "the value 1.5.0 is not a decimal")}},
		{"deeper-path", `{"kind": "conditional", "condition": {"kind": "name", "name": "a"}, "then": {"kind": "name", "name": "b"}, "else": {"kind": "unary", "op": "!", "operand": {"kind": "name", "name": "Bad"}}}`,
			[]Problem{badTree(Path{2, 0}, "the name Bad is not a valid name")}},
		{"value-not-text", `{"kind": "literal", "type": "integer", "value": 5}`, []Problem{badTree(Path{}, "the value 5 is not a text")}},
		{"bad-integer", `{"kind": "literal", "type": "integer", "value": "007"}`, []Problem{badTree(Path{}, "the value 007 is not a integer")}},
		{"bad-boolean", `{"kind": "literal", "type": "boolean", "value": "yes"}`, []Problem{badTree(Path{}, "the value yes is not a boolean")}},
		{"version-zero", `{"kind": "call", "library": null, "name": "late_days", "version": 0, "args": []}`,
			[]Problem{badTree(Path{}, "the value 0 is not a version from 1 to 999999")}},
		{"version-fraction", `{"kind": "call", "library": null, "name": "late_days", "version": 1.5, "args": []}`,
			[]Problem{badTree(Path{}, "the value 1.5 is not a integer")}},
		{"version-text", `{"kind": "call", "library": null, "name": "late_days", "version": "1", "args": []}`,
			[]Problem{badTree(Path{}, "the value 1 is not a integer")}},
		{"library-bad-name", `{"kind": "call", "library": "Fin", "name": "pmt", "version": 1, "args": []}`,
			[]Problem{badTree(Path{}, "the name Fin is not a valid name")}},
		{"args-not-a-list", `{"kind": "call", "library": null, "name": "round", "version": null, "args": "x"}`,
			[]Problem{badTree(Path{}, "the value x is not a list")}},
		{"argument-path", `{"kind": "call", "library": null, "name": "round", "version": null, "args": [{"kind": "name", "name": "x"}, {"kind": "name"}]}`,
			[]Problem{badTree(Path{1}, "the field name is missing")}},
		{"empty-list", `{"kind": "list", "items": []}`, []Problem{badTree(Path{}, "the part has 0 children and needs 1 or more")}},
		{"child-not-an-object", `{"kind": "unary", "op": "!", "operand": "x"}`, []Problem{badTree(Path{0}, "the text is not a tree")}},
		{"two-faults-in-tree-order", `{"kind": "binary", "op": "^", "left": {"kind": "name", "name": "A"}, "right": {"kind": "literal", "type": "integer", "value": "x"}}`, []Problem{
			badTree(Path{}, "the operator ^ is not in the language"),
			badTree(Path{0}, "the name A is not a valid name"),
			badTree(Path{1}, "the value x is not a integer"),
		}},
		{"not-json", `{`, []Problem{badTree(Path{}, "the text is not a tree")}},
		{"empty-input", ``, []Problem{badTree(Path{}, "the text is not a tree")}},
		{"an-array", `[1, 2]`, []Problem{badTree(Path{}, "the text is not a tree")}},
		{"null", `null`, []Problem{badTree(Path{}, "the text is not a tree")}},
		{"trailing-text", `{"kind": "name", "name": "x"} extra`, []Problem{badTree(Path{}, "the text is not a tree")}},
		{"two-trees", `{"kind": "name", "name": "x"}{"kind": "name", "name": "y"}`, []Problem{badTree(Path{}, "the text is not a tree")}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			_, problems := TreeFromJSON([]byte(c.data))
			if !reflect.DeepEqual(problems, c.problems) {
				t.Fatalf("problems differ\nwant: %#v\ngot:  %#v", c.problems, problems)
			}
		})
	}
}

func TestATreeThatCannotBeReadIsEmptyAndOneWithBadValuesIsKept(t *testing.T) {
	tree, problems := TreeFromJSON([]byte(`{"kind": "name"}`))
	if len(problems) == 0 || tree.Root != nil {
		t.Fatalf("an unreadable tree gave %#v and %+v", tree, problems)
	}
	tree, problems = TreeFromJSON([]byte(`{"kind": "name", "name": "Bad"}`))
	if len(problems) != 1 || !reflect.DeepEqual(tree, Tree{Root: Name{Name: "Bad"}}) {
		t.Fatalf("a readable tree with a bad value gave %#v and %+v", tree, problems)
	}
}

func TestAGoodTreeGivesAnEmptyNonNilList(t *testing.T) {
	_, problems := TreeFromJSON([]byte(`{"kind": "name", "name": "x"}`))
	if problems == nil || len(problems) != 0 {
		t.Fatalf("a good tree gave %#v; it must give an empty list that is not nil", problems)
	}
}

func TestAtMostTwentyProblemsAreGiven(t *testing.T) {
	var items []string
	for i := 0; i < 60; i++ {
		items = append(items, `{"kind": "name", "name": "Bad"}`)
	}
	_, problems := TreeFromJSON([]byte(`{"kind": "list", "items": [` + strings.Join(items, ",") + `]}`))
	if len(problems) != 20 {
		t.Fatalf("%d problems given, want 20", len(problems))
	}
	if problems[0].Path[0] != 0 || problems[19].Path[0] != 19 {
		t.Fatalf("the first twenty in tree order were not kept: %+v ... %+v", problems[0], problems[19])
	}
}

func TestADeeplyNestedTreeIsRefusedWithoutHarm(t *testing.T) {
	const depth = 30000
	data := strings.Repeat(`{"kind":"unary","op":"!","operand":`, depth) + `{"kind":"name","name":"x"}` + strings.Repeat(`}`, depth)
	tree, problems := TreeFromJSON([]byte(data))
	if tree.Root != nil || len(problems) != 1 || problems[0].Code != "bad-tree" {
		t.Fatalf("a tree %d deep gave %#v and %+v", depth, tree, problems)
	}
}

func TestATreeReadFromJSONGivesTheSameCheckProblemsAsItsText(t *testing.T) {
	text := `balance * rate`
	tree := mustParse(t, text)
	back, problems := TreeFromJSON(TreeToJSON(tree))
	if len(problems) != 0 {
		t.Fatalf("TreeFromJSON gave %+v", problems)
	}
	known := inputs("balance", "decimal")
	if !reflect.DeepEqual(Check(back, known, testLibrary()), Check(tree, known, testLibrary())) {
		t.Fatalf("Check differs between the tree and its JSON round trip")
	}
}
