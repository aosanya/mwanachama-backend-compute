package exactmoney_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func floatValue(t *testing.T, v float64) number.Value {
	t.Helper()
	f, err := number.NewFloat(v)
	if err != nil {
		t.Fatal(err)
	}
	return number.FloatValue(f)
}

func wantMixRefused(t *testing.T, err error) {
	t.Helper()
	var mix *number.KindMixError
	if !errors.As(err, &mix) {
		t.Fatalf("got %v, want a refusal of the mix", err)
	}
	text := err.Error()
	if !strings.Contains(text, "decimal") || !strings.Contains(text, "float") {
		t.Errorf("the sentence %q does not name both kinds", text)
	}
	if strings.Count(text, ". ") != 0 || !strings.HasSuffix(text, ".") {
		t.Errorf("the sentence %q is not one sentence", text)
	}
}

func TestEM60_ADecimalAndAFloatLiteralAreRefused(t *testing.T) {
	_, err := number.Add(number.DecimalValue(dec(t, "12.50")), floatValue(t, 0.1))
	wantMixRefused(t, err)
}

func TestEM61_ADecimalTimesAFloatIsRefused(t *testing.T) {
	_, err := number.Mul(number.DecimalValue(dec(t, "12.50")), floatValue(t, 0.1))
	wantMixRefused(t, err)
}

func TestEM62_ComparingADecimalWithAFloatIsRefused(t *testing.T) {
	_, err := number.Compare(number.DecimalValue(dec(t, "0.3")), floatValue(t, 0.3))
	wantMixRefused(t, err)
}

func TestEM63_AnExplicitConversionIsAccepted(t *testing.T) {
	f, _ := number.NewFloat(0.1)
	converted, err := f.ToDecimal(2, number.RoundHalfUp, number.ReadShortest)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, converted, "0.10")
	wantText(t, dec(t, "12.50").Add(converted), "12.60")
}

func TestEM64_TheFloatSumOfPointOneAndPointTwoConvertsToPointThree(t *testing.T) {
	a, _ := number.NewFloat(0.1)
	b, _ := number.NewFloat(0.2)
	sum, err := a.Add(b)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"half-up", "half-down", "half-even", "half-+inf"} {
		t.Run(name, func(t *testing.T) {
			converted, err := sum.ToDecimal(2, modeNamed(t, name), number.ReadShortest)
			if err != nil {
				t.Fatal(err)
			}
			wantText(t, converted, "0.30")
		})
	}
}

func TestEM65_UnderUpAndCeilingTheReadingOfAFloatDecidesTheCent_OpenQuestion8(t *testing.T) {
	f, _ := number.NewFloat(0.1)
	for _, name := range []string{"up", "ceiling"} {
		t.Run(name, func(t *testing.T) {
			exact, err := f.ToDecimal(2, modeNamed(t, name), number.ReadExact)
			if err != nil {
				t.Fatal(err)
			}
			wantText(t, exact, "0.11")
			shortest, err := f.ToDecimal(2, modeNamed(t, name), number.ReadShortest)
			if err != nil {
				t.Fatal(err)
			}
			wantText(t, shortest, "0.10")
		})
	}
}

func TestEM66_AConversionStatesItsPlaces(t *testing.T) {
	f, _ := number.NewFloat(0.1)
	if _, err := f.ToDecimal(-1, number.RoundHalfUp, number.ReadShortest); err == nil {
		t.Fatal("a conversion with no places was accepted")
	}
}

func TestFloatTrapsAreNotInTheDecimalPath(t *testing.T) {
	f, _ := number.NewFloat(1.005)
	exact, _ := f.ToDecimal(2, number.RoundHalfUp, number.ReadExact)
	wantText(t, exact, "1.00")
	shortest, _ := f.ToDecimal(2, number.RoundHalfUp, number.ReadShortest)
	wantText(t, shortest, "1.01")
	wantText(t, dec(t, "1.005").Round(2, number.RoundHalfUp), "1.01")
}
