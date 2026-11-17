package function

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

func compactJSON(t *testing.T, text string) string {
	t.Helper()
	var out bytes.Buffer
	if err := json.Compact(&out, []byte(text)); err != nil {
		t.Fatalf("the test JSON is not valid: %v", err)
	}
	return out.String()
}

func TestEveryDefinitionOfTheQAPageRoundTripsThroughJSON(t *testing.T) {
	definitions := map[string]Definition{"d1": d1, "d2": d2, "d3": d3, "d4": d4, "d5": d5, "d6": d6, "d7": d7, "d8": d8, "d9": d9, "d10": d10, "d11": d11}
	for name, definition := range definitions {
		name, definition := name, definition
		t.Run(name, func(t *testing.T) {
			data := DefinitionToJSON(definition)
			back, problems := DefinitionFromJSON(data)
			if problems == nil || len(problems) != 0 {
				t.Fatalf("DefinitionFromJSON gave %#v for %s", problems, data)
			}
			if !reflect.DeepEqual(back, definition) {
				t.Fatalf("the definition changed\nwant: %#v\ngot:  %#v", definition, back)
			}
			if again := DefinitionToJSON(back); !bytes.Equal(again, data) {
				t.Fatalf("writing again changed the JSON\nfirst:  %s\nsecond: %s", data, again)
			}
		})
	}
}

const worked = `{
  "inputs": [
    {
      "name": "balance",
      "type": "decimal",
      "rules": [
        {"rule": "required"},
        {"rule": "no-negatives"},
        {"rule": "decimal-places", "places": 2}
      ]
    },
    {
      "name": "payment",
      "type": "decimal",
      "rules": [
        {"rule": "required"},
        {"rule": "no-negatives"},
        {"rule": "decimal-places", "places": 2}
      ]
    }
  ],
  "checks": [
    {"formula": "payment > 0", "message": "The payment must be more than zero"},
    {"formula": "payment <= balance + 1000.00", "message": "The payment cannot be more than 1000.00 above the balance"}
  ]
}`

func TestTheWorkedDefinitionOfThePageIsWrittenAndReadAsThePageHasIt(t *testing.T) {
	if got, want := string(DefinitionToJSON(d1)), compactJSON(t, worked); got != want {
		t.Fatalf("JSON differs\nwant: %s\ngot:  %s", want, got)
	}
	back, problems := DefinitionFromJSON([]byte(worked))
	if len(problems) != 0 || !reflect.DeepEqual(back, d1) {
		t.Fatalf("the page's definition is not read back: %+v %#v", problems, back)
	}
}

func TestParametersAreWrittenInTheOrderOfTheTableAndLeftOutWhenNotUsed(t *testing.T) {
	cases := []struct {
		rule Rule
		want string
	}{
		{required(), `{"rule":"required"}`},
		{noNegatives(), `{"rule":"no-negatives"}`},
		{minItems(1), `{"rule":"min-items","items":1}`},
		{maxItems(12), `{"rule":"max-items","items":12}`},
		{between("0", "100"), `{"rule":"range","min":"0","max":"100"}`},
		{moreThan("0.00"), `{"rule":"range","min":"0.00","min_exclusive":true}`},
		{upTo("9.5"), `{"rule":"range","max":"9.5"}`},
		{Rule{Name: "range", Min: str("0"), Max: str("5"), MinExclusive: true, MaxExclusive: true}, `{"rule":"range","min":"0","max":"5","min_exclusive":true,"max_exclusive":true}`},
		{places(2), `{"rule":"decimal-places","places":2}`},
		{places(0), `{"rule":"decimal-places","places":0}`},
		{allowed("active", "lapsed"), `{"rule":"allowed-values","values":["active","lapsed"]}`},
	}
	for _, c := range cases {
		definition := Definition{Inputs: []InputDecl{{Name: "x", Type: formula.TypeDecimal, Rules: []Rule{c.rule}}}}
		want := `{"inputs":[{"name":"x","type":"decimal","rules":[` + c.want + `]}],"checks":[]}`
		if got := string(DefinitionToJSON(definition)); got != want {
			t.Errorf("\nwant: %s\ngot:  %s", want, got)
		}
		back, problems := DefinitionFromJSON([]byte(want))
		if len(problems) != 0 && problems[0].Code == "bad-definition" {
			t.Errorf("%s is refused as badly formed: %+v", c.want, problems)
		}
		if !reflect.DeepEqual(back, definition) {
			t.Errorf("%s is not read back\nwant: %#v\ngot:  %#v", c.want, definition, back)
		}
	}
}

func TestAnEmptyDefinitionAndLeftOutListsAreAccepted(t *testing.T) {
	if got := string(DefinitionToJSON(Definition{})); got != `{"inputs":[],"checks":[]}` {
		t.Fatalf("an empty definition is written as %s", got)
	}
	for _, data := range []string{`{}`, `{"inputs":[],"checks":[]}`} {
		back, problems := DefinitionFromJSON([]byte(data))
		if problems == nil || len(problems) != 0 || !reflect.DeepEqual(back, Definition{}) {
			t.Errorf("%s gave %#v and %#v", data, back, problems)
		}
	}
	back, problems := DefinitionFromJSON([]byte(`{"inputs":[{"name":"flag","type":"boolean"}]}`))
	if !reflect.DeepEqual(back.Inputs, []InputDecl{{Name: "flag", Type: formula.TypeBoolean}}) {
		t.Errorf("an input with no rules list is read as %#v", back.Inputs)
	}
	if len(problems) != 1 || problems[0].Code != CodeInputNeedsRequired {
		t.Errorf("the refusal of the definition is not given: %+v", problems)
	}
}

func TestMarkupCharactersInAFormulaOrMessageAreKept(t *testing.T) {
	definition := Definition{
		Inputs: []InputDecl{input("a", formula.TypeInteger, required())},
		Checks: []CheckDecl{check("a < 5 && a > 0", `Keep "a" < 5 & > 0`)},
	}
	data := DefinitionToJSON(definition)
	if strings.Contains(string(data), escaped("003c")) || strings.Contains(string(data), escaped("0026")) {
		t.Errorf("markup characters are escaped in %s", data)
	}
	back, problems := DefinitionFromJSON(data)
	if len(problems) != 0 || !reflect.DeepEqual(back, definition) {
		t.Errorf("not kept: %+v %#v", problems, back)
	}
}

func notValid(part string, what string) formula.Problem {
	sentence := "The definition is not valid: " + what + "."
	if part != "" {
		sentence = "The definition at " + part + " is not valid: " + what + "."
	}
	return formula.Problem{Code: "bad-definition", Sentence: sentence, Part: part}
}

func TestADefinitionThatIsBadlyFormedIsRefusedWithItsPlace(t *testing.T) {
	cases := []struct {
		id       string
		data     string
		problems []formula.Problem
	}{
		{"not-json", `{`, []formula.Problem{notValid("", "the text is not a definition")}},
		{"empty", ``, []formula.Problem{notValid("", "the text is not a definition")}},
		{"an-array", `[]`, []formula.Problem{notValid("", "the text is not a definition")}},
		{"null", `null`, []formula.Problem{notValid("", "the text is not a definition")}},
		{"trailing-text", `{} {}`, []formula.Problem{notValid("", "the text is not a definition")}},
		{"top-level-field", `{"inputs": [], "owner": "x", "audit": 1}`, []formula.Problem{
			notValid("", "the field audit is not allowed"),
			notValid("", "the field owner is not allowed"),
		}},
		{"inputs-not-a-list", `{"inputs": "x"}`, []formula.Problem{notValid("inputs", "the value x is not a list")}},
		{"checks-not-a-list", `{"checks": {}}`, []formula.Problem{notValid("checks", "the value {} is not a list")}},
		{"input-not-an-object", `{"inputs": ["x"]}`, []formula.Problem{notValid("inputs.1", "the text is not a definition")}},
		{"input-without-name", `{"inputs": [{"type": "decimal"}]}`, []formula.Problem{notValid("inputs.1", "the field name is missing")}},
		{"input-without-name-or-type", `{"inputs": [{"rules": []}]}`, []formula.Problem{
			notValid("inputs.1", "the field name is missing"),
			notValid("inputs.1", "the field type is missing"),
		}},
		{"input-extra-field", `{"inputs": [{"name": "a", "type": "decimal", "colour": "red"}]}`, []formula.Problem{notValid("inputs.1", "the field colour is not allowed")}},
		{"name-not-text", `{"inputs": [{"name": 5, "type": "decimal"}]}`, []formula.Problem{notValid("inputs.1", "the value 5 is not a text")}},
		{"type-not-text", `{"inputs": [{"name": "a", "type": true}]}`, []formula.Problem{notValid("inputs.1", "the value true is not a text")}},
		{"rules-not-a-list", `{"inputs": [{"name": "a", "type": "decimal", "rules": "required"}]}`, []formula.Problem{notValid("inputs.1.rules", "the value required is not a list")}},
		{"rule-without-name", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"places": 2}]}]}`, []formula.Problem{notValid("inputs.1.rules.1", "the field rule is missing")}},
		{"second-input-first-rule", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "required"}]}, {"name": "b", "type": "decimal", "rules": [{"rule": "required", "places": 2}]}]}`,
			[]formula.Problem{notValid("inputs.2.rules.1", "the field places is not allowed")}},
		{"places-as-text", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "decimal-places", "places": "2"}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value 2 is not a whole number")}},
		{"places-fraction", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "decimal-places", "places": 2.5}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value 2.5 is not a whole number")}},
		{"items-as-text", `{"inputs": [{"name": "a", "type": "list of integer", "rules": [{"rule": "max-items", "items": "x"}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value x is not a whole number")}},
		{"min-as-number", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "range", "min": 0}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value 0 is not a text")}},
		{"exclusive-as-text", `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "range", "min": "0", "min_exclusive": "yes"}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value yes is not true or false")}},
		{"values-not-a-list", `{"inputs": [{"name": "a", "type": "text", "rules": [{"rule": "allowed-values", "values": "a"}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value a is not a list")}},
		{"value-not-text", `{"inputs": [{"name": "a", "type": "text", "rules": [{"rule": "allowed-values", "values": ["a", 5]}]}]}`,
			[]formula.Problem{notValid("inputs.1.rules.1", "the value 5 is not a text")}},
		{"check-without-message", `{"checks": [{"formula": "a > 0"}]}`, []formula.Problem{notValid("checks.1", "the field message is missing")}},
		{"check-without-formula", `{"checks": [{"message": "x"}]}`, []formula.Problem{notValid("checks.1", "the field formula is missing")}},
		{"check-extra-field", `{"checks": [{"formula": "a > 0", "message": "x", "level": 2}]}`, []formula.Problem{notValid("checks.1", "the field level is not allowed")}},
		{"second-check-formula-not-text", `{"checks": [{"formula": "a > 0", "message": "x"}, {"formula": 7, "message": "y"}]}`, []formula.Problem{notValid("checks.2", "the value 7 is not a text")}},
		{"two-faults-in-order", `{"inputs": [{"type": "decimal"}], "checks": [{"formula": "a > 0"}]}`, []formula.Problem{
			notValid("inputs.1", "the field name is missing"),
			notValid("checks.1", "the field message is missing"),
		}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			definition, problems := DefinitionFromJSON([]byte(c.data))
			if !reflect.DeepEqual(problems, c.problems) {
				t.Fatalf("problems differ\nwant: %#v\ngot:  %#v", c.problems, problems)
			}
			if !reflect.DeepEqual(definition, Definition{}) {
				t.Fatalf("a badly formed definition gave %#v; it must be empty", definition)
			}
		})
	}
}

func TestADefinitionThatIsRefusedOnSaveGivesTheProblemsOfCheckDefinition(t *testing.T) {
	cases := map[string]Definition{
		"range-reversed":     {Inputs: []InputDecl{input("kilos", formula.TypeInteger, required()), input("bag_size", formula.TypeInteger, between("100", "1"))}},
		"places-too-many":    {Inputs: []InputDecl{input("price", formula.TypeDecimal, required(), places(11))}},
		"check-unknown-name": {Inputs: []InputDecl{input("a", formula.TypeInteger, required())}, Checks: []CheckDecl{check("b > 0", "b must be more than zero")}},
		"needs-required":     {Inputs: []InputDecl{input("a", formula.TypeDecimal)}},
		"rule-repeated":      {Inputs: []InputDecl{input("a", formula.TypeDecimal, required(), required())}},
	}
	for id, definition := range cases {
		id, definition := id, definition
		t.Run(id, func(t *testing.T) {
			want := CheckDefinition(definition)
			if len(want) == 0 {
				t.Fatalf("the test definition %s is not refused by CheckDefinition", id)
			}
			back, problems := DefinitionFromJSON(DefinitionToJSON(definition))
			if !reflect.DeepEqual(problems, want) {
				t.Fatalf("problems differ\nwant: %#v\ngot:  %#v", want, problems)
			}
			if !reflect.DeepEqual(back, definition) {
				t.Fatalf("a refused definition is not given back\nwant: %#v\ngot:  %#v", definition, back)
			}
		})
	}
}

func TestARuleNobodyKnowsIsLeftToCheckDefinition(t *testing.T) {
	data := `{"inputs": [{"name": "a", "type": "decimal", "rules": [{"rule": "required"}, {"rule": "shout"}]}]}`
	back, problems := DefinitionFromJSON([]byte(data))
	if len(problems) != 1 || problems[0].Code != CodeBadRuleParameter {
		t.Fatalf("an unknown rule gave %+v", problems)
	}
	if len(back.Inputs) != 1 || len(back.Inputs[0].Rules) != 2 || back.Inputs[0].Rules[1].Name != "shout" {
		t.Fatalf("the unknown rule is not given back: %#v", back)
	}
}

func TestAtMostTwentyProblemsAreGivenForABadlyFormedDefinition(t *testing.T) {
	var inputs []string
	for i := 0; i < 50; i++ {
		inputs = append(inputs, `{"type": "decimal"}`)
	}
	_, problems := DefinitionFromJSON([]byte(`{"inputs": [` + strings.Join(inputs, ",") + `]}`))
	if len(problems) != 20 {
		t.Fatalf("%d problems given, want 20", len(problems))
	}
}

func TestADeeplyNestedDefinitionIsRefusedWithoutHarm(t *testing.T) {
	const depth = 30000
	data := strings.Repeat(`{"inputs":[`, depth) + strings.Repeat(`]}`, depth)
	definition, problems := DefinitionFromJSON([]byte(data))
	if len(problems) != 1 || problems[0].Code != "bad-definition" || !reflect.DeepEqual(definition, Definition{}) {
		t.Fatalf("a definition %d deep gave %#v and %+v", depth, definition, problems)
	}
}

func escaped(code string) string {
	return "\\u" + code
}
