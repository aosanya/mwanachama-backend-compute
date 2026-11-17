package function

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

func refused(code string, part string, sentence string) formula.Problem {
	return formula.Problem{Code: code, Sentence: S(sentence), Part: part}
}

func refusedInside(code string, part string, path formula.Path, sentence string) formula.Problem {
	return formula.Problem{Code: code, Sentence: S(sentence), Part: part, Path: path}
}

type saveCase struct {
	id         string
	definition Definition
	want       []formula.Problem
}

func runSaveCases(t *testing.T, cases []saveCase) {
	t.Helper()
	for _, c := range cases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			got := CheckDefinition(c.definition)
			if got == nil {
				t.Fatalf("the list is nil; an accepted definition gives an empty list")
			}
			if len(c.want) == 0 && len(got) == 0 {
				return
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("problems differ\nwant: %#v\ngot:  %#v", c.want, got)
			}
		})
	}
}

func TestEveryDefinitionOfTheQAPageIsAccepted(t *testing.T) {
	definitions := map[string]Definition{"d1": d1, "d2": d2, "d3": d3, "d4": d4, "d5": d5, "d6": d6, "d7": d7, "d8": d8, "d9": d9, "d10": d10, "d11": d11}
	for name, definition := range definitions {
		got := CheckDefinition(definition)
		if got == nil || len(got) != 0 {
			t.Errorf("%s is refused: %#v", name, got)
		}
	}
}

func TestDefinitionsThatAreRefused(t *testing.T) {
	kilosInput := input("kilos", formula.TypeInteger, required())
	runSaveCases(t, []saveCase{
		{"rules-definition-range-reversed", Definition{Inputs: []InputDecl{kilosInput, input("bag_size", formula.TypeInteger, between("100", "1"))}}, []formula.Problem{
			refused("range-reversed", "inputs.bag_size.range", "The range of ~bag_size~ has a minimum of 100 above its maximum of 1."),
			refused("input-needs-required", "inputs.bag_size", "~bag_size~ is not a list, so it needs the rule required."),
		}},
		{"rules-definition-places-too-many", Definition{Inputs: []InputDecl{input("payment", formula.TypeDecimal, required(), places(11))}}, []formula.Problem{
			refused("places-too-many", "inputs.payment.decimal-places", "The rule decimal-places of ~payment~ allows 11 places. The most is 10."),
		}},
		{"rules-definition-allowed-values-empty", Definition{Inputs: []InputDecl{input("status", formula.TypeText, required(), allowed())}}, []formula.Problem{
			refused("allowed-values-empty", "inputs.status.allowed-values", "The allowed values of ~status~ list no value. A list holds 1 to 20 values."),
		}},
		{"rules-definition-check-unknown-name", Definition{Inputs: d1.Inputs, Checks: []CheckDecl{check("payment <= limit", "The payment must stay under the limit")}}, []formula.Problem{
			refusedInside("unknown-name", "limit", formula.Path{1}, "Check 1: The name ~limit~ is not an input of this function. Its inputs are: balance, payment."),
		}},
		{"rules-definition-check-not-boolean", Definition{Inputs: d1.Inputs, Checks: []CheckDecl{check("payment + 1", "The payment plus one")}}, []formula.Problem{
			refused("check-not-boolean", "checks.1", "Check 1 must be true or false, and ~payment + 1~ is decimal."),
		}},
		{"rules-definition-input-needs-required", Definition{Inputs: []InputDecl{input("bag_size", formula.TypeInteger, between("1", "100"))}}, []formula.Problem{
			refused("input-needs-required", "inputs.bag_size", "~bag_size~ is not a list, so it needs the rule required."),
		}},
	})
}

func TestTheOtherRefusalsOfTheLanguagePage(t *testing.T) {
	many := make([]InputDecl, 9)
	for i := range many {
		many[i] = input(fmt.Sprintf("field_%d", i+1), formula.TypeInteger, required())
	}
	manyChecks := make([]CheckDecl, 9)
	for i := range manyChecks {
		manyChecks[i] = check("kilos > 0", "Kilos must be more than zero")
	}
	twentyOne := make([]string, 21)
	for i := range twentyOne {
		twentyOne[i] = fmt.Sprintf("value_%d", i+1)
	}
	kilos := Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required())}}
	withChecks := func(checks ...CheckDecl) Definition {
		return Definition{Inputs: kilos.Inputs, Checks: checks}
	}
	runSaveCases(t, []saveCase{
		{"input-repeated", Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required()), input("kilos", formula.TypeDecimal, required())}}, []formula.Problem{
			refused("input-repeated", "inputs.kilos", "The input ~kilos~ is declared more than once."),
		}},
		{"bad-input-name", Definition{Inputs: []InputDecl{input("Kilos", formula.TypeInteger, required())}}, []formula.Problem{
			refused("bad-input-name", "inputs.Kilos", "~Kilos~ is not a valid input name. A name has 1 to 40 lowercase letters, digits and underscores and starts with a letter."),
		}},
		{"bad-input-name-too-long", Definition{Inputs: []InputDecl{input(strings.Repeat("a", 41), formula.TypeInteger, required())}}, []formula.Problem{
			refused("bad-input-name", "inputs."+strings.Repeat("a", 41), "~"+strings.Repeat("a", 41)+"~ is not a valid input name. A name has 1 to 40 lowercase letters, digits and underscores and starts with a letter."),
		}},
		{"too-many-inputs", Definition{Inputs: many}, []formula.Problem{refused("too-many-inputs", "inputs", "This function goes over the input limit of 8.")}},
		{"rule-repeated", Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required(), noNegatives(), noNegatives())}}, []formula.Problem{
			refused("rule-repeated", "inputs.kilos.no-negatives", "The rule no-negatives is given twice for ~kilos~."),
		}},
		{"rule-not-for-type", Definition{Inputs: []InputDecl{input("status", formula.TypeText, required(), noNegatives())}}, []formula.Problem{
			refused("rule-not-for-type", "inputs.status.no-negatives", "~status~ is text, so it cannot have the rule no-negatives."),
		}},
		{"rule-not-for-date-range", Definition{Inputs: []InputDecl{input("filed", formula.TypeDate, required(), between("1", "2"))}}, []formula.Problem{
			refused("rule-not-for-type", "inputs.filed.range", "~filed~ is date, so it cannot have the rule range."),
		}},
		{"rule-not-for-list-allowed", Definition{Inputs: []InputDecl{input("scores", formula.TypeListInteger, allowed("1"))}}, []formula.Problem{
			refused("rule-not-for-type", "inputs.scores.allowed-values", "~scores~ is list of integer, so it cannot have the rule allowed-values."),
		}},
		{"bad-rule-parameter-range-none", Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required(), Rule{Name: "range"})}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.kilos.range", "The rule range of ~kilos~ needs a minimum or a maximum written as decimal text."),
		}},
		{"bad-rule-parameter-range-text", Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required(), between("one", "5"))}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.kilos.range", "The rule range of ~kilos~ needs a minimum or a maximum written as decimal text."),
		}},
		{"bad-rule-parameter-places-missing", Definition{Inputs: []InputDecl{input("payment", formula.TypeDecimal, required(), Rule{Name: "decimal-places"})}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.payment.decimal-places", "The rule decimal-places of ~payment~ needs places, a whole number from 0 to 10."),
		}},
		{"bad-rule-parameter-items-zero", Definition{Inputs: []InputDecl{input("scores", formula.TypeListInteger, minItems(0))}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.scores.min-items", "The rule min-items of ~scores~ needs items, a whole number from 1 to 1000."),
		}},
		{"bad-rule-parameter-unknown-rule", Definition{Inputs: []InputDecl{input("kilos", formula.TypeInteger, required(), Rule{Name: "positive"})}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.kilos.positive", "The rule positive of ~kilos~ is not a rule of this library."),
		}},
		{"bad-rule-parameter-bad-type", Definition{Inputs: []InputDecl{input("kilos", "number", required())}}, []formula.Problem{
			refused("bad-rule-parameter", "inputs.kilos", "~kilos~ has the type number, which is not a type of this library."),
		}},
		{"items-reversed", Definition{Inputs: []InputDecl{input("scores", formula.TypeListInteger, minItems(5), maxItems(2))}}, []formula.Problem{
			refused("items-reversed", "inputs.scores.min-items", "~scores~ needs at least 5 items and at most 2."),
		}},
		{"allowed-values-too-many", Definition{Inputs: []InputDecl{input("status", formula.TypeText, required(), allowed(twentyOne...))}}, []formula.Problem{
			refused("allowed-values-too-many", "inputs.status.allowed-values", "The allowed values of ~status~ list 21 values. The most is 20."),
		}},
		{"allowed-values-repeat", Definition{Inputs: []InputDecl{input("status", formula.TypeText, required(), allowed("active", "lapsed", "active"))}}, []formula.Problem{
			refused("allowed-values-repeat", "inputs.status.allowed-values", "The value \"active\" is listed twice for ~status~."),
		}},
		{"rules-contradict-range", Definition{Inputs: []InputDecl{input("grade", formula.TypeInteger, required(), between("1", "5"), allowed("3", "9"))}}, []formula.Problem{
			refused("rules-contradict", "inputs.grade.allowed-values", "No value can meet the rules of ~grade~: the allowed value 9 breaks the rule range."),
		}},
		{"rules-contradict-negatives", Definition{Inputs: []InputDecl{input("grade", formula.TypeInteger, required(), noNegatives(), allowed("-1"))}}, []formula.Problem{
			refused("rules-contradict", "inputs.grade.allowed-values", "No value can meet the rules of ~grade~: the allowed value -1 breaks the rule no-negatives."),
		}},
		{"rules-contradict-type", Definition{Inputs: []InputDecl{input("grade", formula.TypeInteger, required(), allowed("abc"))}}, []formula.Problem{
			refused("rules-contradict", "inputs.grade.allowed-values", "No value can meet the rules of ~grade~: the allowed value abc breaks the rule type."),
		}},
		{"too-many-checks", Definition{Inputs: kilos.Inputs, Checks: manyChecks}, []formula.Problem{refused("too-many-checks", "checks", "This function goes over the check limit of 8.")}},
		{"bad-message-empty", withChecks(check("kilos > 0", "")), []formula.Problem{refused("bad-message", "checks.1", "Check 1 needs a message of 1 to 120 characters on one line.")}},
		{"bad-message-long", withChecks(check("kilos > 0", strings.Repeat("a", 121))), []formula.Problem{refused("bad-message", "checks.1", "Check 1 needs a message of 1 to 120 characters on one line.")}},
		{"bad-message-line-break", withChecks(check("kilos > 0", "one\ntwo")), []formula.Problem{refused("bad-message", "checks.1", "Check 1 needs a message of 1 to 120 characters on one line.")}},
		{"check-keeps-its-own-code", withChecks(check("kilos > 0", "Kilos are needed"), check("grab(kilos) > 0", "Kilos are grabbed")), []formula.Problem{
			refusedInside("unknown-function", "grab(kilos)", formula.Path{0}, "Check 2: The function grab is not known."),
		}},
		{"check-that-cannot-be-read", withChecks(check("kilos >", "Kilos are needed")), []formula.Problem{
			{Code: "bad-syntax", Sentence: "Check 1: The formula cannot be read: the formula ends after an operator at line 1, column 8.", Line: 1, Column: 8},
		}},
	})
}

func TestADefinitionWithManyProblemsGivesAtMostTwenty(t *testing.T) {
	inputs := make([]InputDecl, 8)
	for i := range inputs {
		inputs[i] = input(fmt.Sprintf("field_%d", i+1), formula.TypeInteger, between("9", "1"))
	}
	got := CheckDefinition(Definition{Inputs: inputs})
	if len(got) != 16 {
		t.Fatalf("8 inputs with 2 problems each give 16 problems, got %d", len(got))
	}
	more := make([]InputDecl, 8)
	for i := range more {
		more[i] = input(fmt.Sprintf("field_%d", i+1), formula.TypeInteger, between("9", "1"), allowed(), noNegatives(), noNegatives())
	}
	got = CheckDefinition(Definition{Inputs: more})
	if len(got) != 20 {
		t.Errorf("the list is cut at 20 problems, got %d", len(got))
	}
}
