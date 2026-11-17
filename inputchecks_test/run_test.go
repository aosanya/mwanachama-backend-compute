package inputchecks

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const unchecked = -1

type testCase struct {
	id   string
	def  definition
	row  function.Row
	want []problem.Problem
	ran  int
	held string
	open string
}

func clean(id string, d definition, received function.Row, ran int) testCase {
	return testCase{id: id, def: d, row: received, want: []problem.Problem{}, ran: ran}
}

func bad(id string, d definition, received function.Row, ran int, texts ...string) testCase {
	return testCase{id: id, def: d, row: received, want: lines(texts...), ran: ran}
}

func withoutValues(problems []problem.Problem) []problem.Problem {
	stripped := make([]problem.Problem, len(problems))
	for i, p := range problems {
		p.Value = nil
		stripped[i] = p
	}
	return stripped
}

func unspecified(c testCase, question string) testCase {
	c.open = question
	return c
}

func sameList(got []problem.Problem, want []problem.Problem) bool {
	return reflect.DeepEqual(got, want)
}

func hold(c testCase, question string) testCase {
	c.held = question
	return c
}

const decision22 = "decision 22 against the language page's no-negatives rule (input-checks open question 1)"

func show(problems []problem.Problem) string {
	var out strings.Builder
	for i, p := range problems {
		subject := "-"
		if p.Input != nil {
			subject = *p.Input
		}
		if p.Item != nil {
			subject += fmt.Sprintf(" item %d", *p.Item)
		}
		if p.Check != nil {
			subject = "check " + *p.Check
		}
		value := "not given"
		if p.Value != nil {
			value = fmt.Sprintf("%q", *p.Value)
		}
		out.WriteString(fmt.Sprintf("%d. %s · %s · %s · rule %s — %s\n", i+1, p.Code, subject, value, p.Rule, p.Sentence))
	}
	return out.String()
}

func runCase(t *testing.T, c testCase) {
	t.Helper()
	if c.held != "" {
		t.Skip("held: " + c.held)
	}
	runner := &engine{}
	got := function.Validate(c.def.body, c.row, runner)
	if got == nil {
		t.Fatalf("%s: the list is nil; a clean row gives an empty list", c.id)
	}
	want := c.want
	if c.open != "" {
		got, want = withoutValues(got), withoutValues(want)
		t.Log("value field not compared: " + c.open)
	}
	if !sameList(got, want) {
		t.Errorf("%s: the problem list differs\nwant:\n%s\ngot:\n%s", c.id, show(want), show(got))
	}
	if c.ran != unchecked && runner.calls != c.ran {
		t.Errorf("%s: %d checks ran, want %d", c.id, runner.calls, c.ran)
	}
}

func runCases(t *testing.T, cases []testCase) {
	t.Helper()
	seen := map[string]bool{}
	for _, c := range cases {
		if seen[c.id] {
			t.Fatalf("case id %s is used twice", c.id)
		}
		seen[c.id] = true
		t.Run(c.id, func(t *testing.T) { runCase(t, c) })
	}
}

func heldOf(cases []testCase) []testCase {
	var held []testCase
	for _, c := range cases {
		if c.held != "" {
			held = append(held, c)
		}
	}
	return held
}

func TestHeldCasesAreListedWithTheirQuestions(t *testing.T) {
	var all []testCase
	for _, group := range allGroups() {
		all = append(all, group...)
	}
	held := heldOf(all)
	if len(held) == 0 {
		t.Fatal("no case is held, but ic-c07 must be")
	}
	for _, c := range held {
		t.Logf("held %s: %s", c.id, c.held)
	}
	found := false
	for _, c := range held {
		if c.id == "ic-c07" {
			found = true
		}
	}
	if !found {
		t.Fatal("ic-c07 is not held")
	}
}

func TestTheComparisonSeesAWrongSentenceAnOrderAndAMissingProblem(t *testing.T) {
	want := lines(
		"no-negatives · balance · -1.00 — ~balance~ is -1.00. It cannot be negative.",
		"no-negatives · payment · -5.125 — ~payment~ is -5.125. It cannot be negative.",
	)
	got := function.Validate(d1.body, row("balance", "-1.00", "payment", "-5.125"), &engine{})
	if len(got) != 3 {
		t.Fatalf("the row of ic-c04 gives %d problems", len(got))
	}
	if sameList(got, want) {
		t.Fatal("a list with a missing problem was taken as equal")
	}
	reworded := lines(
		"no-negatives · balance · -1.00 — ~balance~ is -1.00. It must not be negative.",
		"no-negatives · payment · -5.125 — ~payment~ is -5.125. It cannot be negative.",
		"decimal-places · payment · -5.125 — ~payment~ is -5.125. It can have at most 2 decimal places.",
	)
	if sameList(got, reworded) {
		t.Fatal("a sentence worded differently was taken as equal")
	}
	swapped := []problem.Problem{got[1], got[0], got[2]}
	if sameList(got, swapped) {
		t.Fatal("a problem out of order was taken as equal")
	}
	if !sameList(got, append([]problem.Problem{}, got...)) {
		t.Fatal("equal lists were taken as different")
	}
}
