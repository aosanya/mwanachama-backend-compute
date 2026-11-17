package own

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

const (
	decimalKind = harness.KindDecimal
	floatKind   = harness.KindFloat
	integerKind = harness.KindInteger
	textKind    = harness.KindText
	boolKind    = harness.KindBoolean
	dateKind    = harness.KindDate
)

var sampleParams = []harness.Param{
	{Name: "d", Kind: decimalKind},
	{Name: "f", Kind: floatKind},
	{Name: "n", Kind: integerKind},
	{Name: "s", Kind: textKind},
	{Name: "b", Kind: boolKind},
	{Name: "when", Kind: dateKind},
	{Name: "ds", Kind: harness.KindDecimalList},
	{Name: "ns", Kind: harness.KindIntegerList},
}

func compileWith(source string, result harness.Kind, library harness.Library) (harness.Program, error) {
	c := harness.Case{Number: 1, Source: source, Params: sampleParams, Result: result}
	return New().Compile(c, library)
}

func compileOK(t *testing.T, source string, result harness.Kind) harness.Program {
	t.Helper()
	program, err := compileWith(source, result, nil)
	if err != nil {
		t.Fatalf("%q should compile: %v", source, err)
	}
	return program
}

func sampleInputs() map[string]any {
	return map[string]any{
		"d": numeric.MustParse("2.50"), "f": 0.5, "n": int64(4), "s": "x", "b": true,
		"when": harnessDate(),
		"ds":   []numeric.Decimal{numeric.MustParse("1.00"), numeric.MustParse("2.00")},
		"ns":   []int64{1, 2, 3},
	}
}

var everyResultKind = []harness.Kind{decimalKind, floatKind, integerKind, textKind, boolKind, dateKind, harness.KindDecimalList}

func TestDecimalAndFloatNeverMixWithoutAConversion(t *testing.T) {
	mixed := []string{"d + f", "d - f", "d * f", "d / f", "d < f", "d == f", "d != f", "[d, f]", "[d, 1.5f]", "b ? d : f", "d + 0.5f", "d * 2.5f"}
	for _, source := range mixed {
		for _, result := range everyResultKind {
			if _, err := compileWith(source, result, nil); err == nil {
				t.Errorf("%q should not compile as %s", source, result)
			}
		}
	}
	_, err := compileWith("d + f", decimalKind, nil)
	want := "The formula cannot be used: a decimal and a float cannot be combined, convert one with to_float or to_decimal at line 1, column 3."
	if err == nil || err.Error() != want {
		t.Errorf("got %v", err)
	}
}

func TestConversionsMakeMixedFormulasLegal(t *testing.T) {
	compileOK(t, "to_float(d) + f", floatKind)
	compileOK(t, "d + to_decimal(f, 2)", decimalKind)
	compileOK(t, "to_float(n) * f", floatKind)
}

func TestIntegersWidenToDecimalOrFloat(t *testing.T) {
	compileOK(t, "n + d", decimalKind)
	compileOK(t, "d * n", decimalKind)
	compileOK(t, "n * f", floatKind)
	compileOK(t, "n < d", boolKind)
	compileOK(t, "n < f", boolKind)
	compileOK(t, "b ? n : d", decimalKind)
	compileOK(t, "max([n, d])", decimalKind)
	compileOK(t, "avg(ns) >= 80", boolKind)
	compileOK(t, "n / n", decimalKind)
	compileOK(t, "n % n", integerKind)
}

func TestTheResultKindMustEqualTheCaseResult(t *testing.T) {
	_, err := compileWith("n + 1", decimalKind, nil)
	if err == nil || err.Error() != "The formula produces an integer but a decimal is expected." {
		t.Errorf("got %v", err)
	}
	_, err = compileWith("d > 1", textKind, nil)
	if err == nil || err.Error() != "The formula produces a boolean but a text is expected." {
		t.Errorf("got %v", err)
	}
}

func TestUnknownNamesAndFunctionsAreRefusedInPlainSentences(t *testing.T) {
	cases := []struct{ source, message string }{
		{"nope(1)", "The function nope is not known."},
		{"late_days@1(when, when)", "The function late_days@1 is not known."},
		{"missing + 1", "The name missing is not known."},
		{"round(missing, 2)", "The name missing is not known."},
	}
	for _, c := range cases {
		_, err := compileWith(c.source, decimalKind, nil)
		if err == nil || err.Error() != c.message {
			t.Errorf("%q: got %v, want %s", c.source, err, c.message)
		}
	}
}

func TestWrongArgumentsNameWhatTheFunctionAccepts(t *testing.T) {
	_, err := compileWith(`round(s, 2)`, decimalKind, nil)
	want := "The formula cannot be used: the function round accepts (decimal, integer) or (float, integer) but was given (text, integer) at line 1, column 1."
	if err == nil || err.Error() != want {
		t.Errorf("got %v", err)
	}
	for _, source := range []string{"sum(d)", "sum()", "days_between(when)", "count(n, n)", "to_decimal(d, 2)", "round(n, 2)"} {
		if _, err := compileWith(source, decimalKind, nil); err == nil || !strings.Contains(err.Error(), "accepts") {
			t.Errorf("%q: got %v", source, err)
		}
	}
}

func TestOperatorsRefuseKindsTheyDoNotFit(t *testing.T) {
	cases := []struct{ source, part string }{
		{"s + s", "+ cannot be used on text values"},
		{"d % d", "% cannot be used on decimal values"},
		{"f % f", "% cannot be used on float values"},
		{"b < b", "< cannot put booleans in order"},
		{"ds == ds", "== cannot compare decimal list values"},
		{"n + s", "an integer and a text cannot be combined"},
		{"s < n", "a text and an integer cannot be combined"},
		{"n && b", "&& needs booleans but was given an integer"},
		{"!n", "! needs a boolean but was given an integer"},
		{"-s", "- needs a number but was given a text"},
		{"n ? 1 : 2", "the condition before ? must be a boolean but is an integer"},
		{"[s]", "a list can hold only integers or decimals, not text values"},
		{"[d, s]", "the values in a list must agree"},
		{"b ? d : s", "a decimal and a text cannot be combined"},
	}
	for _, c := range cases {
		_, err := compileWith(c.source, decimalKind, nil)
		if err == nil || !strings.Contains(err.Error(), c.part) || !strings.Contains(err.Error(), "line 1, column") {
			t.Errorf("%q: got %v, want it to contain %q", c.source, err, c.part)
		}
	}
}

func TestParamsWithoutAKnownKindAreRefused(t *testing.T) {
	c := harness.Case{Source: "x", Params: []harness.Param{{Name: "x"}}, Result: decimalKind}
	if _, err := New().Compile(c, nil); err == nil || err.Error() != "The input x has no kind the formulas know." {
		t.Errorf("got %v", err)
	}
}

func TestFormulasAreCompiledOnceAndCarryNoCompileTimeState(t *testing.T) {
	program := compileOK(t, "d * 2", decimalKind)
	first, err := program.Run(sampleInputs())
	second, err2 := program.Run(sampleInputs())
	if err != nil || err2 != nil || !harness.Equal(first, second) {
		t.Errorf("runs differ: %v %v %v %v", first, second, err, err2)
	}
}
