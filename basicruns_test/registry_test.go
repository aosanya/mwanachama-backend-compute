package basicruns

import (
	"math"
	"strconv"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/library"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

var pmtSignature = formula.Signature{
	Params: []formula.Param{
		{Name: "principal", Type: formula.TypeDecimal},
		{Name: "rate", Type: formula.TypeDecimal},
		{Name: "months", Type: formula.TypeInteger},
	},
	Result: formula.TypeDecimal,
	Cost:   10,
}

func monthlyPayment(args []formula.Value) float64 {
	principal, _ := strconv.ParseFloat(args[0].Decimal().String(), 64)
	rate, _ := strconv.ParseFloat(args[1].Decimal().String(), 64)
	months := float64(args[2].Integer())
	return principal * rate / (1 - math.Pow(1+rate, -months))
}

func paymentAs(t *testing.T, amount string) formula.Value {
	t.Helper()
	decimal, err := number.ParseDecimal(amount)
	if err != nil {
		t.Fatalf("decimal %q: %v", amount, err)
	}
	return formula.DecimalValue(decimal)
}

func newRegistry(t *testing.T) *library.Registry {
	t.Helper()
	registry := library.NewRegistry()
	if err := library.LoadFunctions(registry); err != nil {
		t.Fatalf("the ten library functions did not load: %v", err)
	}
	releases := []library.Release{
		{
			Library: "fin", Name: "pmt", Version: 1, Signature: pmtSignature,
			Apply: func(args []formula.Value) (formula.Value, *formula.Problem) {
				return paymentAs(t, strconv.Itoa(int(math.Floor(monthlyPayment(args))))), nil
			},
		},
		{
			Library: "fin", Name: "pmt", Version: 2, Signature: pmtSignature,
			Apply: func(args []formula.Value) (formula.Value, *formula.Problem) {
				return paymentAs(t, strconv.FormatFloat(math.Round(monthlyPayment(args)*100)/100, 'f', 2, 64)), nil
			},
			Note: library.Note{
				Changed: "The payment is rounded to the nearest 0.01. It was rounded down to a whole number.",
				Matters: "Any agency that shows or collects a loan payment.",
				Example: library.Example{Input: "12,000 at 1% a month over 24 months", Old: "564", New: "564.88"},
			},
		},
	}
	for _, release := range releases {
		if err := registry.Publish(release); err != nil {
			t.Fatalf("publish fin.pmt@%d: %v", release.Version, err)
		}
	}
	return registry
}
