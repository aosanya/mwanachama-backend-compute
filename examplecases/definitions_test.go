package examplecases

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	fdef "github.com/aosanya/mwanachama-backend-compute/function"
)

const definitionsPath = "../testdata/formula-language.definitions.md"

func loadDefinitions(t *testing.T) Definitions {
	t.Helper()
	file, err := LoadDefinitions(definitionsPath)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func inputOf(t *testing.T, definition fdef.Definition, name string) fdef.InputDecl {
	t.Helper()
	for _, input := range definition.Inputs {
		if input.Name == name {
			return input
		}
	}
	t.Fatalf("no input %q", name)
	return fdef.InputDecl{}
}

func ruleOf(input fdef.InputDecl, name string) (fdef.Rule, bool) {
	for _, rule := range input.Rules {
		if rule.Name == name {
			return rule, true
		}
	}
	return fdef.Rule{}, false
}

func TestTheFileHoldsTheTwelveDefinitionsOfTheQAPageAndSixDerivedOnes(t *testing.T) {
	file := loadDefinitions(t)
	if len(file.Definitions) != 18 {
		t.Fatalf("%d definitions", len(file.Definitions))
	}
	counts := map[string][2]int{"V1": {2, 2}, "V2": {3, 0}, "V3": {3, 0}, "V4": {3, 0}, "V5": {2, 0}, "V9": {3, 2}, "V10": {1, 0}, "V16": {2, 0}, "V17": {1, 0}, "V18": {5, 0}, "V19": {2, 0}, "VB": {2, 1}}
	for id, want := range counts {
		definition, found := file.Definitions[id]
		if !found {
			t.Fatalf("no definition %s", id)
		}
		if len(definition.Inputs) != want[0] || len(definition.Checks) != want[1] {
			t.Errorf("%s has %d inputs and %d checks, want %d and %d", id, len(definition.Inputs), len(definition.Checks), want[0], want[1])
		}
	}
}

func TestTheRulesAreReadWithTheirParameters(t *testing.T) {
	file := loadDefinitions(t)
	price := inputOf(t, file.Definitions["V4"], "price_per_kilo")
	rule, found := ruleOf(price, "range")
	if !found || rule.Min == nil || *rule.Min != "0.000" || !rule.MinExclusive || rule.Max != nil {
		t.Errorf("price_per_kilo range: %+v", rule)
	}
	places, found := ruleOf(price, "decimal-places")
	if !found || places.Places == nil || *places.Places != 3 {
		t.Errorf("price_per_kilo decimal-places: %+v", places)
	}
	members := inputOf(t, file.Definitions["V3"], "members")
	rule, _ = ruleOf(members, "range")
	if rule.Min == nil || *rule.Min != "1" || rule.Max == nil || *rule.Max != "500" || rule.MinExclusive || rule.MaxExclusive {
		t.Errorf("members range: %+v", rule)
	}
	scores := inputOf(t, file.Definitions["V10"], "scores")
	if scores.Type != formula.TypeListInteger {
		t.Errorf("scores type %s", scores.Type)
	}
	most, found := ruleOf(scores, "max-items")
	if !found || most.Items == nil || *most.Items != 12 {
		t.Errorf("scores max-items: %+v", most)
	}
	status := inputOf(t, file.Definitions["V2"], "status")
	allowed, found := ruleOf(status, "allowed-values")
	if !found || strings.Join(allowed.Values, "|") != "active|lapsed|suspended" {
		t.Errorf("status allowed-values: %+v", allowed)
	}
	for _, id := range []string{"V1", "V2", "V3"} {
		for _, input := range file.Definitions[id].Inputs {
			if input.Name == "contributions" || input.Name == "deposits" {
				if input.Type != formula.TypeListDecimal {
					t.Errorf("%s type %s", input.Name, input.Type)
				}
			}
		}
	}
}

func TestADerivedDefinitionCopiesItsBaseAndReplacesWhatItNames(t *testing.T) {
	file := loadDefinitions(t)
	reversed := file.Definitions["V19-range-reversed"]
	if len(reversed.Inputs) != 2 {
		t.Fatalf("%d inputs", len(reversed.Inputs))
	}
	rule, _ := ruleOf(inputOf(t, reversed, "bag_size"), "range")
	if rule.Min == nil || *rule.Min != "100" || rule.Max == nil || *rule.Max != "1" {
		t.Errorf("range %+v", rule)
	}
	base := file.Definitions["V19"]
	rule, _ = ruleOf(inputOf(t, base, "bag_size"), "range")
	if *rule.Min != "1" || *rule.Max != "100" {
		t.Errorf("the base changed: %+v", rule)
	}
	if _, required := ruleOf(inputOf(t, file.Definitions["V19-input-needs-required"], "bag_size"), "required"); required {
		t.Error("input-needs-required still has required")
	}
	if got := file.Definitions["V1-check-unknown-name"].Checks[0].Formula; got != "payment <= limit" {
		t.Errorf("check %q", got)
	}
	if len(file.Definitions["V1-check-unknown-name"].Checks) != 2 {
		t.Error("the second check of V1 was lost")
	}
	empty, _ := ruleOf(inputOf(t, file.Definitions["V2-allowed-values-empty"], "status"), "allowed-values")
	if len(empty.Values) != 0 {
		t.Errorf("values %v", empty.Values)
	}
}

func TestEveryBaseDefinitionIsAcceptedOnSave(t *testing.T) {
	file := loadDefinitions(t)
	for id, definition := range file.Definitions {
		if strings.Contains(id, "-") {
			continue
		}
		if problems := fdef.CheckDefinition(definition); len(problems) != 0 {
			t.Errorf("%s is refused: %v", id, problems)
		}
	}
}

func TestEveryRulesCaseHasOneBindingAndNoOtherCaseHasOne(t *testing.T) {
	file := loadDefinitions(t)
	rules := 0
	for _, c := range loadAll(t) {
		_, bound := file.Bindings[c.Name]
		if strings.HasPrefix(c.Name, "rules-") {
			rules++
			if !bound {
				t.Errorf("%s has no binding", c.Name)
			}
		} else if bound {
			t.Errorf("%s is bound and is not a rules case", c.Name)
		}
	}
	if rules != 56 || len(file.Bindings) != 56 {
		t.Errorf("%d rules cases, %d bindings", rules, len(file.Bindings))
	}
}

func TestEveryBindingNamesADefinitionAndATarget(t *testing.T) {
	file := loadDefinitions(t)
	var targetRows, targetDefinitions int
	for name, binding := range file.Bindings {
		if _, found := file.Definitions[binding.Definition]; !found {
			t.Errorf("%s names the definition %q, which is not in the file", name, binding.Definition)
		}
		switch binding.Target {
		case TargetRow:
			targetRows++
		case TargetDefinition:
			targetDefinitions++
		default:
			t.Errorf("%s has the target %q", name, binding.Target)
		}
	}
	if targetRows != 50 || targetDefinitions != 6 {
		t.Errorf("%d row cases, %d definition cases", targetRows, targetDefinitions)
	}
}

const (
	definitionHeader = "| id | part | name | detail | message |\n|---|---|---|---|---|\n"
	bindingHeader    = "| case | definition | target |\n|---|---|---|\n"
)

func TestAMalformedDefinitionFileIsRefusedWithItsLineAndReason(t *testing.T) {
	good := "| A | input | x | decimal: required | |\n"
	refusals := []struct {
		name   string
		text   string
		reason string
		line   int
	}{
		{"no header", "# Definitions\n\n| A | input | x | decimal | |\n", "bad header", 3},
		{"wrong column count", definitionHeader + "| A | input | x |\n", "wrong column count", 3},
		{"unknown part", definitionHeader + "| A | outputs | x | decimal | |\n", "bad part", 3},
		{"unknown type", definitionHeader + "| A | input | x | money: required | |\n", "bad type", 3},
		{"unknown rule", definitionHeader + "| A | input | x | decimal: positive | |\n", "bad rule", 3},
		{"rule without a number", definitionHeader + "| A | input | x | decimal: decimal-places | |\n", "bad rule", 3},
		{"range without an end", definitionHeader + "| A | input | x | decimal: range 1 | |\n", "bad rule", 3},
		{"input twice", definitionHeader + good + good, "duplicate name", 4},
		{"check without a message", definitionHeader + good + "| A | check | 1 | x > 0 | |\n", "empty cell", 4},
		{"from an unknown definition", definitionHeader + "| B | from | A | | |\n", "unknown definition", 3},
		{"definition twice", definitionHeader + good + "| B | input | x | decimal | |\n| A | input | y | decimal | |\n", "definition split", 5},
		{"binding to an unknown definition", definitionHeader + good + "\n" + bindingHeader + "| c1 | Z | row |\n", "unknown definition", 7},
		{"binding with a bad target", definitionHeader + good + "\n" + bindingHeader + "| c1 | A | both |\n", "bad target", 7},
		{"case bound twice", definitionHeader + good + "\n" + bindingHeader + "| c1 | A | row |\n| c1 | A | row |\n", "duplicate name", 8},
	}
	for _, refusal := range refusals {
		t.Run(refusal.name, func(t *testing.T) {
			_, err := ParseDefinitions("x.definitions.md", refusal.text)
			failure, ok := err.(FileError)
			if !ok {
				t.Fatalf("got %v", err)
			}
			if failure.Reason != refusal.reason || failure.Line != refusal.line {
				t.Fatalf("got %s:%d %s, want %d %s", failure.File, failure.Line, failure.Reason, refusal.line, refusal.reason)
			}
		})
	}
}
