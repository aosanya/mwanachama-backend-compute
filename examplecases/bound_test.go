package examplecases

import (
	"strings"
	"testing"
)

func TestARowIsGivenToValidateAsReceivedTextAndLists(t *testing.T) {
	cases := loadAll(t)
	var c Case
	for _, candidate := range cases {
		if candidate.Name == "rules-allowed-values-valid" {
			c = candidate
		}
	}
	row, err := RowOf(c.Inputs)
	if err != nil {
		t.Fatal(err)
	}
	status := row["status"]
	if status.List || status.Text != "lapsed" {
		t.Errorf("status %+v", status)
	}
	contributions := row["contributions"]
	if !contributions.List || strings.Join(contributions.Items, "|") != "200.00|250.00|150.00|100.00" {
		t.Errorf("contributions %+v", contributions)
	}
	if len(row) != 3 {
		t.Errorf("%d names in the row", len(row))
	}
}

func TestAValueThatIsNotReceivedTextIsRefusedAsARow(t *testing.T) {
	for name, input := range map[string]Input{
		"integer": {Name: "kilos", Value: Value{Type: "int", Text: "3"}},
		"decimal": {Name: "kilos", Value: Value{Type: "dec", Text: "3.00"}},
		"boolean": {Name: "kilos", Value: Value{Type: "bool", Text: "true"}},
		"list":    {Name: "kilos", Value: Value{Type: "list<int>", Items: []string{"1"}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := RowOf([]Input{input}); err == nil {
				t.Fatal("no refusal")
			}
		})
	}
}

func TestEveryValidateCaseRunsUnlessItIsHeldAndTheHeldOnesAreOnlyThoseThatWaitForADecision(t *testing.T) {
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	definitions := loadDefinitions(t)
	var ran, waiting int
	for _, c := range loadAll(t) {
		if !strings.HasPrefix(c.Name, "rules-") {
			continue
		}
		if reason, isHeld := held[c.Name]; isHeld {
			waiting++
			if !strings.Contains(reason, "decision") && !strings.Contains(reason, "bug") {
				t.Errorf("%s is held for %q, which is neither a decision nor a filed bug", c.Name, reason)
			}
			continue
		}
		outcome, err := Run(c, definitions, runtime)
		if err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		if message := Mismatch(c, outcome); message != "" {
			t.Error(message)
		}
		ran++
	}
	if ran+waiting != 56 {
		t.Fatalf("%d ran and %d waiting", ran, waiting)
	}
	if ran < 50 {
		t.Fatalf("only %d of the 56 Validate cases run", ran)
	}
}

func TestADefinitionCaseNeverRunsTheFormula(t *testing.T) {
	definitions := loadDefinitions(t)
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	var found int
	for _, c := range loadAll(t) {
		binding, bound := definitions.Bindings[c.Name]
		if !bound || binding.Target != TargetDefinition {
			continue
		}
		found++
		if len(c.Inputs) != 0 {
			t.Errorf("%s has inputs", c.Name)
		}
		outcome, err := Run(c, definitions, runtime)
		if err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		if outcome.Result != nil || len(outcome.Problems) != 1 {
			t.Errorf("%s: %+v", c.Name, outcome)
		}
	}
	if found != 6 {
		t.Fatalf("%d definition cases", found)
	}
}
