package cel

import (
	"strings"
	"sync"
	"testing"

	"cel.dev/cel-go/cel"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func TestEveryFormulaPasses(t *testing.T) {
	harness.RunTest(t, New())
}

func TestEngineIsNamedCelGo(t *testing.T) {
	if got := New().Name(); got != "cel-go" {
		t.Errorf("name is %q", got)
	}
}

func TestTranslateRewritesLiteralsAndLibraryCalls(t *testing.T) {
	cases := map[string]string{
		`balance < 1000.00 ? 0.010 : 0.020`:            `balance < decimal("1000.00") ? decimal("0.010") : decimal("0.020")`,
		`round(balance * interest_step@1(balance), 2)`: `round(balance * interest_step_v1(balance), 2)`,
		`status == "A 1.50 @1" && rate > 0.75f`:        `status == "A 1.50 @1" && rate > 0.75`,
		`rate * 2f`:                                    `rate * 2.0`,
		`kilos % 50`:                                   `kilos % 50`,
	}
	for source, want := range cases {
		if got := translate(source); got != want {
			t.Errorf("translate(%s) = %s, want %s", source, got, want)
		}
	}
}

func TestTranslationsReportsWhetherTheTextDiffers(t *testing.T) {
	differs := 0
	for _, c := range harness.Cases() {
		text, changed := Translations(c)
		if text == "" {
			t.Errorf("formula %d translated to nothing", c.Number)
		}
		if changed != (text != c.Source) {
			t.Errorf("formula %d: changed flag disagrees with the text", c.Number)
		}
		if changed {
			differs++
		}
	}
	if differs == 0 || differs == len(harness.Cases()) {
		t.Errorf("%d of %d formulas differ", differs, len(harness.Cases()))
	}
}

func compileSource(t *testing.T, source string, params ...harness.Param) (harness.Program, error) {
	t.Helper()
	return New().Compile(harness.Case{Source: source, Params: params}, harness.Library{})
}

func TestCheckerRefusesDecimalMixedWithDouble(t *testing.T) {
	params := []harness.Param{{Name: "amount", Kind: harness.KindDecimal}, {Name: "rate", Kind: harness.KindFloat}}
	for _, source := range []string{`amount + rate`, `amount * rate`, `rate < amount`, `amount + 0.75f`} {
		if _, err := compileSource(t, source, params...); err == nil {
			t.Errorf("%s compiled", source)
		}
	}
}

func TestCheckerAcceptsDecimalWithIntEitherSide(t *testing.T) {
	params := []harness.Param{{Name: "amount", Kind: harness.KindDecimal}, {Name: "count", Kind: harness.KindInteger}}
	for _, source := range []string{`amount * count`, `count * amount`, `amount / count`, `count < amount`, `amount >= count`} {
		if _, err := compileSource(t, source, params...); err != nil {
			t.Errorf("%s: %v", source, err)
		}
	}
}

func TestResultKindMismatchIsRefused(t *testing.T) {
	c := harness.Case{Source: `1 + 1`, Result: harness.KindDecimal}
	if _, err := New().Compile(c, harness.Library{}); err == nil {
		t.Error("an int result should not satisfy a decimal formula")
	}
}

func TestMissingLibraryFunctionIsRefused(t *testing.T) {
	c := harness.Cases()[5]
	if _, err := New().Compile(c, harness.Library{}); err == nil {
		t.Error("formula 6 compiled without late_days@1")
	}
}

func TestDivisionByZeroIsARunError(t *testing.T) {
	program, err := compileSource(t, `amount / count`, harness.Param{Name: "amount", Kind: harness.KindDecimal}, harness.Param{Name: "count", Kind: harness.KindInteger})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Run(map[string]any{"amount": numeric.MustParse("1.00"), "count": int64(0)}); err == nil {
		t.Error("division by zero should fail")
	}
}

func TestDecimalEqualityIgnoresScale(t *testing.T) {
	program, err := compileSource(t, `amount == 1.5`, harness.Param{Name: "amount", Kind: harness.KindDecimal})
	if err != nil {
		t.Fatal(err)
	}
	got, err := program.Run(map[string]any{"amount": numeric.MustParse("1.50")})
	if err != nil || got != true {
		t.Errorf("got %v, %v", got, err)
	}
}

func TestOptionsReachTheEnvironmentAndTheProgram(t *testing.T) {
	engine := NewWith(Options{
		EnvOptions:     []cel.EnvOption{cel.Variable("extra", cel.IntType)},
		ProgramOptions: []cel.ProgramOption{cel.CostLimit(1)},
	})
	program, err := engine.Compile(harness.Case{Source: `extra + extra + extra`}, harness.Library{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Run(map[string]any{"extra": int64(1)}); err == nil || !strings.Contains(err.Error(), "cost") {
		t.Errorf("cost limit did not apply: %v", err)
	}
}

func TestOneProgramRunsFromManyGoroutines(t *testing.T) {
	library := harness.Library{}
	engine := New()
	cases := harness.Cases()
	late, err := engine.Compile(cases[4], library)
	if err != nil {
		t.Fatal(err)
	}
	library["late_days@1"] = late
	program, err := engine.Compile(cases[5], library)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for g := 0; g < 16; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				for _, example := range cases[5].Examples {
					got, err := program.Run(example.Inputs)
					if err != nil || !harness.Equal(example.Want, got) {
						t.Errorf("%s: got %v, %v", example.Name, got, err)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}
