package examplecases

import "testing"

const heldPath = "../testdata/formula-language.held.md"

func TestEveryCaseRunsAndGivesItsExpectedResultOrProblems(t *testing.T) {
	cases := loadAll(t)
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewStandIn()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if reason, isHeld := held[c.Name]; isHeld {
				t.Skipf("held: %s", reason)
			}
			outcome, err := Evaluate(c, runtime)
			if err != nil {
				t.Fatalf("%s:%d: %v", c.File, c.Line, err)
			}
			compare(t, c, outcome)
		})
	}
}

func compare(t *testing.T, c Case, outcome Outcome) {
	t.Helper()
	if message := Mismatch(c, outcome); message != "" {
		t.Fatal(message)
	}
}

func TestEveryHeldNameIsACaseOfTheFile(t *testing.T) {
	held, err := LoadHeld(heldPath)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, c := range loadAll(t) {
		names[c.Name] = true
	}
	for name := range held {
		if !names[name] {
			t.Errorf("%s holds %q, which is not a case", heldPath, name)
		}
	}
}
