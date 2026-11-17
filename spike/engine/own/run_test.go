package own

import (
	"sync"
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func harnessDate() time.Time {
	return time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
}

func runOK(t *testing.T, source string, result harness.Kind, inputs map[string]any) any {
	t.Helper()
	got, err := compileOK(t, source, result).Run(inputs)
	if err != nil {
		t.Fatalf("%q: %v", source, err)
	}
	return got
}

func withInputs(overrides map[string]any) map[string]any {
	inputs := sampleInputs()
	for name, value := range overrides {
		inputs[name] = value
	}
	return inputs
}

func TestIntegerDivisionIsExactAndGivesADecimal(t *testing.T) {
	got := runOK(t, "7 / 2", decimalKind, sampleInputs())
	if !harness.Equal(numeric.MustParse("3.500000000000000000"), got) {
		t.Errorf("got %s", harness.Describe(got))
	}
}

func TestIntegerArithmeticStaysIntegerAndWidensWithDecimals(t *testing.T) {
	cases := []struct {
		source string
		kind   harness.Kind
		want   any
	}{
		{"n * 3 + 1", integerKind, int64(13)},
		{"n % 3", integerKind, int64(1)},
		{"-n + 10", integerKind, int64(6)},
		{"n * 0.50", decimalKind, numeric.MustParse("2.00")},
		{"d + n", decimalKind, numeric.MustParse("6.50")},
		{"n * 0.5f", floatKind, 2.0},
		{"d < n", boolKind, true},
		{"1 + 2 * 3 == 7", boolKind, true},
		{`s == "x" && !(n > 4)`, boolKind, true},
		{"when < 2026-04-01", boolKind, true},
		{"when == 2026-03-31", boolKind, true},
		{"when > 2026-03-31", boolKind, false},
		{"sum(ds) + sum(ns)", decimalKind, numeric.MustParse("9.00")},
		{"b ? 1 : 2 + 3", integerKind, int64(1)},
		{`s != "y" ? "yes" : "no"`, textKind, "yes"},
		{"b == true", boolKind, true},
	}
	for _, c := range cases {
		if got := runOK(t, c.source, c.kind, sampleInputs()); !harness.Equal(c.want, got) {
			t.Errorf("%q: want %s, got %s", c.source, harness.Describe(c.want), harness.Describe(got))
		}
	}
}

func TestRunTimeFailuresAreSentences(t *testing.T) {
	cases := []struct {
		source string
		kind   harness.Kind
		inputs map[string]any
		want   string
	}{
		{"d / d", decimalKind, withInputs(map[string]any{"d": numeric.MustParse("0.00")}), "The formula divides by zero."},
		{"n % n", integerKind, withInputs(map[string]any{"n": int64(0)}), "The formula divides by zero."},
		{"n / n", decimalKind, withInputs(map[string]any{"n": int64(0)}), "The formula divides by zero."},
		{"f / f", floatKind, withInputs(map[string]any{"f": 0.0}), "The formula divides by zero."},
		{"avg(ds)", decimalKind, withInputs(map[string]any{"ds": []numeric.Decimal{}}), "The formula takes a value from an empty list."},
		{"max(ns)", integerKind, withInputs(map[string]any{"ns": []int64{}}), "The formula takes a value from an empty list."},
		{"n * 4611686018427387904", integerKind, sampleInputs(), "The formula produced a whole number that is too large."},
	}
	for _, c := range cases {
		_, err := compileOK(t, c.source, c.kind).Run(c.inputs)
		if err == nil || err.Error() != c.want {
			t.Errorf("%q: got %v, want %s", c.source, err, c.want)
		}
	}
}

func TestConditionalAndLogicalOperatorsOnlyEvaluateWhatTheyNeed(t *testing.T) {
	zero := withInputs(map[string]any{"n": int64(0)})
	cases := []struct {
		source string
		kind   harness.Kind
		want   any
	}{
		{"n == 0 || 10 / n > 1", boolKind, true},
		{"n != 0 && 10 / n > 1", boolKind, false},
		{"n != 0 ? 10 / n : 0.00", decimalKind, numeric.MustParse("0.00")},
	}
	for _, c := range cases {
		if got := runOK(t, c.source, c.kind, zero); !harness.Equal(c.want, got) {
			t.Errorf("%q: want %s, got %s", c.source, harness.Describe(c.want), harness.Describe(got))
		}
	}
}

func TestRunRefusesInputsThatDoNotFitTheDeclaredParams(t *testing.T) {
	program := compileOK(t, "d + 1", decimalKind)
	cases := []struct {
		inputs map[string]any
		want   string
	}{
		{withInputs(map[string]any{"d": "text"}), "The input d must be a decimal but is a text."},
		{withInputs(map[string]any{"n": 3}), "The input n holds a value the formulas do not know."},
		{withInputs(map[string]any{"ns": []numeric.Decimal{}}), "The input ns must be an integer list but is a decimal list."},
		{map[string]any{}, "The input d is missing."},
	}
	for _, c := range cases {
		if _, err := program.Run(c.inputs); err == nil || err.Error() != c.want {
			t.Errorf("got %v, want %s", err, c.want)
		}
	}
}

func TestRunDoesNotChangeTheInputs(t *testing.T) {
	inputs := sampleInputs()
	runOK(t, "sum(ds) + d", decimalKind, inputs)
	if len(inputs) != len(sampleInputs()) {
		t.Errorf("inputs changed: %v", inputs)
	}
}

func TestAProgramIsSafeForManyGoroutinesAtOnce(t *testing.T) {
	library := harness.Library{}
	lateDays := harness.Case{Number: 5, Source: "max([0, days_between(due, filed)])", Result: integerKind,
		Params: []harness.Param{{Name: "due", Kind: dateKind}, {Name: "filed", Kind: dateKind}}}
	compiled, err := New().Compile(lateDays, library)
	if err != nil {
		t.Fatal(err)
	}
	library["late_days@1"] = compiled
	penalty := harness.Case{Number: 6, Result: decimalKind,
		Source: "late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : 0.00",
		Params: []harness.Param{{Name: "amount", Kind: decimalKind}, {Name: "due", Kind: dateKind}, {Name: "filed", Kind: dateKind}}}
	program, err := New().Compile(penalty, library)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for round := range 200 {
				days := (worker*200 + round) % 90
				due := harnessDate()
				got, err := program.Run(map[string]any{
					"amount": numeric.MustParse("8400.00"), "due": due, "filed": due.AddDate(0, 0, days),
				})
				want := numeric.MustParse("0.00")
				if days > 30 {
					want = numeric.MustParse("420.00")
				}
				if err != nil || !harness.Equal(want, got) {
					t.Errorf("days %d: got %v %v", days, got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}
