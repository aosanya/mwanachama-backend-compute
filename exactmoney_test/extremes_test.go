package exactmoney_test

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func TestEM90_ZeroPlusZero(t *testing.T) {
	wantText(t, dec(t, "0.00").Add(dec(t, "0.00")), "0.00")
}

func TestEM91_ZeroTimesAnything(t *testing.T) {
	wantEqual(t, dec(t, "123.45").Mul(number.DecimalFromInt(0)), "0.00")
}

func TestEM94_TheSmallestAmountIsNotLost(t *testing.T) {
	wantText(t, dec(t, "0.01").Add(dec(t, "0.01")), "0.02")
}

func TestEM95_AHalfCentIsARoundingCaseNotALoss(t *testing.T) {
	half := dec(t, "0.01").Mul(dec(t, "0.5"))
	wantEqual(t, half, "0.005")
	cases := map[string]string{"half-up": "0.01", "half-down": "0.00", "half-even": "0.00"}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			wantText(t, half.Round(2, modeNamed(t, name)), want)
		})
	}
}

func TestEM97_OneThirdIsKeptToTheAgencysDigits(t *testing.T) {
	settings := number.DefaultSettings()
	third, err := settings.Divide(dec(t, "1"), dec(t, "3"))
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, third, "0.33333333333333333333")
	wantText(t, settings.Round(third), "0.33")
	wantText(t, third.Mul(number.DecimalFromInt(3)), "0.99999999999999999999")
}

func TestEM97_TheDigitsKeptFollowTheSetting(t *testing.T) {
	for _, digits := range []int{2, 5, 20, 30} {
		settings := number.Settings{Places: 2, Mode: number.RoundHalfUp, DivisionDigits: digits}
		third, err := settings.Divide(dec(t, "1"), dec(t, "3"))
		if err != nil {
			t.Fatal(err)
		}
		if third.Scale() != digits {
			t.Errorf("%d digits kept, got %d", digits, third.Scale())
		}
	}
}

func TestEM98_TwoThirdsToTwoPlacesNamesTheMode(t *testing.T) {
	cases := map[string]string{"half-up": "0.67", "up": "0.67", "ceiling": "0.67", "down": "0.66", "floor": "0.66"}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			mode := modeNamed(t, name)
			kept, err := dec(t, "2").Div(dec(t, "3"), 20, mode)
			if err != nil {
				t.Fatal(err)
			}
			wantText(t, kept.Round(2, mode), want)
		})
	}
}

func TestEM99_TheSmallestUnitFollowsTheCurrencysPlaces(t *testing.T) {
	settings := number.Settings{Places: 4, Mode: number.RoundHalfUp, DivisionDigits: 20}
	one := dec(t, "0.0001")
	wantText(t, one.Add(one), "0.0002")
	half := one.Mul(dec(t, "0.5"))
	wantEqual(t, half, "0.00005")
	wantText(t, settings.Round(half), "0.0001")
}

func TestEM100_ANegativeAmountAddsExactly(t *testing.T) {
	wantText(t, dec(t, "-10.00").Add(dec(t, "4.00")), "-6.00")
}

func TestDivisionByZeroIsRefused(t *testing.T) {
	if _, err := number.DefaultSettings().Divide(dec(t, "1"), dec(t, "0")); err == nil {
		t.Fatal("a division by zero was accepted")
	}
}
