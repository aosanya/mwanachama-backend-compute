package harness

import "testing"

func TestCasesAreTheNineteenFormulas(t *testing.T) {
	cases := Cases()
	if len(cases) != 19 {
		t.Fatalf("got %d cases", len(cases))
	}
	for i, c := range cases {
		if c.Number != i+1 {
			t.Errorf("case at index %d is numbered %d", i, c.Number)
		}
		if c.Title == "" || c.Source == "" || len(c.Params) == 0 {
			t.Errorf("formula %d is incomplete", c.Number)
		}
	}
}

func TestEveryExampleMatchesItsCaseTypes(t *testing.T) {
	for _, c := range Cases() {
		if len(c.Examples) == 0 {
			t.Errorf("formula %d has no examples", c.Number)
		}
		seen := map[string]bool{}
		for _, example := range c.Examples {
			if seen[example.Name] {
				t.Errorf("formula %d repeats example name %q", c.Number, example.Name)
			}
			seen[example.Name] = true
			if len(example.Inputs) != len(c.Params) {
				t.Errorf("formula %d example %q has %d inputs for %d params", c.Number, example.Name, len(example.Inputs), len(c.Params))
			}
			for _, p := range c.Params {
				value, ok := example.Inputs[p.Name]
				if !ok {
					t.Errorf("formula %d example %q lacks input %s", c.Number, example.Name, p.Name)
					continue
				}
				if kind, known := KindOf(value); !known || kind != p.Kind {
					t.Errorf("formula %d example %q input %s is %s, want %s", c.Number, example.Name, p.Name, Describe(value), p.Kind)
				}
			}
			if kind, known := KindOf(example.Want); !known || kind != c.Result {
				t.Errorf("formula %d example %q result is %s, want %s", c.Number, example.Name, Describe(example.Want), c.Result)
			}
		}
	}
}

func TestExampleCountsFollowThePage(t *testing.T) {
	atLeast := map[int]int{1: 2, 5: 2, 6: 2, 7: 3}
	for _, c := range Cases() {
		if got := len(c.Examples); got < atLeast[c.Number] {
			t.Errorf("formula %d has %d examples, want at least %d", c.Number, got, atLeast[c.Number])
		}
	}
}

func TestPublishedAndUsedNames(t *testing.T) {
	cases := Cases()
	published := map[string]int{}
	for _, c := range cases {
		if c.PublishedAs != "" {
			published[c.PublishedAs] = c.Number
		}
	}
	if published["late_days@1"] != 5 || published["interest_step@1"] != 7 || len(published) != 2 {
		t.Errorf("published names are %v", published)
	}
	wantUses := map[int][]string{6: {"late_days@1"}, 8: {"interest_step@1"}}
	for _, c := range cases {
		want := wantUses[c.Number]
		if len(c.Uses) != len(want) {
			t.Errorf("formula %d uses %v, want %v", c.Number, c.Uses, want)
			continue
		}
		for i, name := range c.Uses {
			if name != want[i] {
				t.Errorf("formula %d uses %v, want %v", c.Number, c.Uses, want)
			}
		}
	}
}

func TestDependencyOrderPutsPublishedFirst(t *testing.T) {
	ordered := inDependencyOrder(Cases())
	if len(ordered) != 19 {
		t.Fatalf("got %d cases", len(ordered))
	}
	position := map[int]int{}
	for i, c := range ordered {
		position[c.Number] = i
	}
	if position[5] > position[6] || position[7] > position[8] {
		t.Errorf("order is %v", position)
	}
}

func TestDependencyOrderKeepsCyclesAndUnknownNames(t *testing.T) {
	cases := []Case{
		{Number: 1, PublishedAs: "a@1", Uses: []string{"b@1"}},
		{Number: 2, PublishedAs: "b@1", Uses: []string{"a@1"}},
		{Number: 3, Uses: []string{"missing@1"}},
	}
	if got := len(inDependencyOrder(cases)); got != 3 {
		t.Errorf("got %d cases", got)
	}
}
