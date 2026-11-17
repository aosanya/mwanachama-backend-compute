package number

import (
	"errors"
	"math"
	"strings"
	"testing"
)

func mustFloat(t *testing.T, v float64) Float {
	t.Helper()
	f, err := NewFloat(v)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFloatRefusesWhatIsNotFinite(t *testing.T) {
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := NewFloat(v); !errors.Is(err, ErrNotFinite) {
			t.Fatalf("NewFloat(%v) must be refused as not finite, got %v", v, err)
		}
	}
	huge := mustFloat(t, math.MaxFloat64)
	if _, err := huge.Add(huge); !errors.Is(err, ErrNotFinite) {
		t.Fatalf("overflowing a float must be refused as not finite, got %v", err)
	}
	if _, err := mustFloat(t, 1).Div(mustFloat(t, 0)); err == nil {
		t.Fatal("float division by zero must be refused")
	}
}

func TestFloatArithmeticIsBinaryAndShowsItsDrift(t *testing.T) {
	sum, err := mustFloat(t, 0.1).Add(mustFloat(t, 0.2))
	if err != nil {
		t.Fatal(err)
	}
	if sum.Value() == 0.3 {
		t.Fatal("float 0.1 + 0.2 must not equal 0.3, that is the drift decimals avoid")
	}
}

func TestEM63_ExplicitConversionOfFloatPointOne(t *testing.T) {
	converted, err := mustFloat(t, 0.1).ToDecimal(2, RoundHalfUp, ReadShortest)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, converted, "0.10")
	wantText(t, dec(t, "12.50").Add(converted), "12.60")
}

func TestEM64_ExplicitConversionOfTheDriftedSum(t *testing.T) {
	for _, mode := range []Rounding{RoundHalfUp, RoundHalfDown, RoundHalfEven} {
		for _, reading := range []FloatReading{ReadShortest, ReadExact} {
			converted, err := mustFloat(t, 0.30000000000000004).ToDecimal(2, mode, reading)
			if err != nil {
				t.Fatal(err)
			}
			wantText(t, converted, "0.30")
		}
	}
}

func TestEM65_UpAndCeilingOfFloatPointOneDependOnTheReading(t *testing.T) {
	for _, mode := range []Rounding{RoundUp, RoundCeiling} {
		exact, err := mustFloat(t, 0.1).ToDecimal(2, mode, ReadExact)
		if err != nil {
			t.Fatal(err)
		}
		wantText(t, exact, "0.11")
		shortest, err := mustFloat(t, 0.1).ToDecimal(2, mode, ReadShortest)
		if err != nil {
			t.Fatal(err)
		}
		wantText(t, shortest, "0.10")
	}
}

func TestOpenQuestionEight_FloatOnePointZeroZeroFive(t *testing.T) {
	shortest, _ := mustFloat(t, 1.005).ToDecimal(2, RoundHalfUp, ReadShortest)
	exact, _ := mustFloat(t, 1.005).ToDecimal(2, RoundHalfUp, ReadExact)
	wantText(t, shortest, "1.01")
	wantText(t, exact, "1.00")
}

func TestEM66_ConversionStatesItsPlaces(t *testing.T) {
	if _, err := mustFloat(t, 0.1).ToDecimal(-1, RoundHalfUp, ReadShortest); err == nil {
		t.Fatal("a conversion without valid places must be refused")
	}
}

func TestDecimalToFloatIsExplicitAndFinite(t *testing.T) {
	f, err := dec(t, "1.50").ToFloat()
	if err != nil {
		t.Fatal(err)
	}
	if f.Value() != 1.5 {
		t.Fatalf("got %v, want 1.5", f.Value())
	}
	huge := dec(t, "1"+strings.Repeat("0", 400))
	if _, err := huge.ToFloat(); !errors.Is(err, ErrNotFinite) {
		t.Fatalf("a decimal too big for a float must be refused as not finite, got %v", err)
	}
}

func TestEM60To62_MixingTheKindsIsRefusedInOneSentence(t *testing.T) {
	amount := DecimalValue(dec(t, "12.50"))
	rate := FloatValue(mustFloat(t, 0.1))
	operations := map[string]func() error{
		"add":      func() error { _, err := Add(amount, rate); return err },
		"subtract": func() error { _, err := Sub(rate, amount); return err },
		"multiply": func() error { _, err := Mul(amount, rate); return err },
		"compare":  func() error { _, err := Compare(DecimalValue(dec(t, "0.3")), FloatValue(mustFloat(t, 0.3))); return err },
	}
	for name, run := range operations {
		err := run()
		var mix *KindMixError
		if !errors.As(err, &mix) {
			t.Fatalf("%s: want a kind mix error, got %v", name, err)
		}
		sentence := err.Error()
		if !strings.Contains(sentence, "decimal") || !strings.Contains(sentence, "float") {
			t.Fatalf("%s: the sentence must name both kinds: %q", name, sentence)
		}
		if strings.Count(sentence, ".") != 1 || strings.Contains(sentence, "\n") {
			t.Fatalf("%s: the refusal must be one sentence: %q", name, sentence)
		}
	}
}

func TestSameKindOperationsWorkThroughValues(t *testing.T) {
	sum, err := Add(DecimalValue(dec(t, "12.50")), DecimalValue(dec(t, "0.10")))
	if err != nil {
		t.Fatal(err)
	}
	if sum.Kind() != KindDecimal {
		t.Fatal("a decimal and a decimal give a decimal")
	}
	wantText(t, sum.Decimal(), "12.60")
	product, err := Mul(FloatValue(mustFloat(t, 2)), FloatValue(mustFloat(t, 0.5)))
	if err != nil {
		t.Fatal(err)
	}
	if product.Kind() != KindFloat || product.Float().Value() != 1 {
		t.Fatal("a float and a float give a float")
	}
	order, err := Compare(DecimalValue(dec(t, "0.30")), DecimalValue(dec(t, "0.3")))
	if err != nil || order != 0 {
		t.Fatalf("0.30 and 0.3 compare equal, got %d %v", order, err)
	}
}
