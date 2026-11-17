package exprlang

import (
	"sync"
	"testing"

	"github.com/expr-lang/expr"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func TestFormulas(t *testing.T) {
	harness.RunTest(t, New())
}

func TestTranslate(t *testing.T) {
	cases := map[string]string{
		`balance < 1000.00 ? 0.010 : 0.020`: `balance < dec("1000.00") ? dec("0.010") : dec("0.020")`,
		`amount * 0.75f`:                    `amount * 0.75`,
		`late_days@1(due, filed) > 90`:      `late_days_v1(due, filed) > 90`,
		`role == "a@1 1.50" && x2 > 3`:      `role == "a@1 1.50" && x2 > 3`,
		`role == 'it\'s 2.5' ? 1 : 2`:       `role == 'it\'s 2.5' ? 1 : 2`,
		`months * 3 + p1.5`:                 `months * 3 + p1.5`,
		`round(x, 2) + 10`:                  `round(x, 2) + 10`,
	}
	for source, want := range cases {
		if got := translate(source); got != want {
			t.Errorf("translate(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestTranslationsFlagsChangedFormulas(t *testing.T) {
	changed := 0
	for _, c := range harness.Cases() {
		if _, differs := Translations(c); differs {
			changed++
		}
	}
	if changed == 0 || changed == len(harness.Cases()) {
		t.Fatalf("expected some but not all formulas to change, %d of %d did", changed, len(harness.Cases()))
	}
}

func TestDecimalAndFloatNeverMix(t *testing.T) {
	params := []harness.Param{{Name: "d", Kind: harness.KindDecimal}, {Name: "f", Kind: harness.KindFloat}}
	for _, source := range []string{`d + f`, `f * d`, `d < f`, `d == f`, `d * 0.75f`} {
		_, err := New().Compile(harness.Case{Source: source, Params: params}, harness.Library{})
		if err == nil {
			t.Errorf("%q compiled, want a compile error", source)
		}
	}
}

func TestOptionsReachExpr(t *testing.T) {
	c := harness.Cases()[0]
	_, err := NewWith(Options{ExprOptions: []expr.Option{expr.MaxNodes(2)}}).Compile(c, harness.Library{})
	if err == nil {
		t.Fatal("MaxNodes(2) did not stop a larger formula")
	}
}

func TestConcurrentRunsShareProgramsSafely(t *testing.T) {
	library := harness.Library{}
	engine := New()
	type compiled struct {
		program harness.Program
		c       harness.Case
	}
	var programs []compiled
	for _, c := range inOrder(harness.Cases()) {
		program, err := engine.Compile(c, library)
		if err != nil {
			t.Fatalf("formula %d: %v", c.Number, err)
		}
		if c.PublishedAs != "" {
			library[c.PublishedAs] = program
		}
		programs = append(programs, compiled{program, c})
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 50 {
				for _, p := range programs {
					for _, example := range p.c.Examples {
						got, err := p.program.Run(example.Inputs)
						if err != nil || !harness.Equal(example.Want, got) {
							t.Errorf("formula %d example %q: got %v, err %v", p.c.Number, example.Name, got, err)
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}

func inOrder(cases []harness.Case) []harness.Case {
	var published, rest []harness.Case
	for _, c := range cases {
		if c.PublishedAs != "" {
			published = append(published, c)
		} else {
			rest = append(rest, c)
		}
	}
	return append(published, rest...)
}
