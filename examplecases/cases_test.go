package examplecases

import (
	"strconv"
	"strings"
	"testing"
)

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
	where := c.File + ":" + strconv.Itoa(c.Line)
	if c.Expected.Result != nil {
		if outcome.Result == nil {
			t.Fatalf("%s: expected %s, got problems %v: %v", where, c.Expected.Result, outcome.Problems, outcome.Where)
		}
		if outcome.Result.String() != c.Expected.Result.String() {
			t.Fatalf("%s: expected %s, got %s", where, c.Expected.Result, outcome.Result)
		}
		return
	}
	if outcome.Result != nil {
		t.Fatalf("%s: expected problems %v, got the result %s", where, c.Expected.Problems, outcome.Result)
	}
	if strings.Join(outcome.Problems, ", ") != strings.Join(c.Expected.Problems, ", ") {
		t.Fatalf("%s: expected problems %v, got %v: %v", where, c.Expected.Problems, outcome.Problems, outcome.Where)
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
