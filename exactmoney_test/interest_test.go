package exactmoney_test

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func TestEM52_SimpleInterestHasNoRounding(t *testing.T) {
	yearly := dec(t, "1000.00").Mul(dec(t, "0.05"))
	wantEqual(t, yearly, "50.00")
	interest := yearly.Mul(number.DecimalFromInt(3))
	wantEqual(t, interest, "150.00")
	wantEqual(t, dec(t, "1000.00").Add(interest), "1150.00")
}

func TestEM53_ARateWithOneDecimalIsExact(t *testing.T) {
	yearly := dec(t, "2500.00").Mul(dec(t, "0.045"))
	wantEqual(t, yearly, "112.50")
	interest := yearly.Mul(number.DecimalFromInt(2))
	wantEqual(t, interest, "225.00")
	wantEqual(t, dec(t, "2500.00").Add(interest), "2725.00")
}

func compoundUnrounded(t *testing.T, start string, factor string, periods int) number.Decimal {
	t.Helper()
	value := dec(t, start)
	for i := 0; i < periods; i++ {
		value = value.Mul(dec(t, factor))
	}
	return value
}

func compoundRoundedEach(t *testing.T, start string, factor string, periods int, mode number.Rounding) number.Decimal {
	t.Helper()
	value := dec(t, start)
	for i := 0; i < periods; i++ {
		value = value.Mul(dec(t, factor)).Round(2, mode)
	}
	return value
}

func TestEM54_CompoundCarriesEveryDigitUntilTheEnd(t *testing.T) {
	unrounded := compoundUnrounded(t, "1000.00", "1.05", 3)
	wantEqual(t, unrounded, "1157.625")
	want := map[string]string{
		"half-up": "1157.63", "half-down": "1157.62", "half-even": "1157.62", "half-+inf": "1157.63",
		"up": "1157.63", "down": "1157.62", "ceiling": "1157.63", "floor": "1157.62",
	}
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			wantText(t, unrounded.Round(2, m.mode), want[m.name])
		})
	}
}

func TestEM55_RoundingEachPeriodAgainstOnceAtTheEnd(t *testing.T) {
	unrounded := compoundUnrounded(t, "500.00", "1.015", 4)
	wantEqual(t, unrounded, "530.6817753125")
	cases := []struct {
		mode string
		each string
		once string
	}{
		{"half-up", "530.68", "530.68"},
		{"half-down", "530.68", "530.68"},
		{"half-even", "530.68", "530.68"},
		{"up", "530.70", "530.69"},
		{"ceiling", "530.70", "530.69"},
		{"down", "530.67", "530.68"},
		{"floor", "530.67", "530.68"},
	}
	for _, c := range cases {
		t.Run(c.mode, func(t *testing.T) {
			mode := modeNamed(t, c.mode)
			wantText(t, compoundRoundedEach(t, "500.00", "1.015", 4, mode), c.each)
			wantText(t, unrounded.Round(2, mode), c.once)
		})
	}
}

func TestEM55_ThePeriodTable(t *testing.T) {
	value := dec(t, "500.00")
	want := []string{"507.50", "515.11", "522.84", "530.68"}
	for _, text := range want {
		value = value.Mul(dec(t, "1.015")).Round(2, number.RoundHalfUp)
		wantText(t, value, text)
	}
}

func TestEM56_TheTwoApproachesDiffer(t *testing.T) {
	unrounded := compoundUnrounded(t, "120.00", "1.02", 3)
	wantEqual(t, unrounded, "127.344960")
	cases := []struct {
		mode string
		each string
		once string
	}{
		{"half-up", "127.35", "127.34"},
		{"half-down", "127.35", "127.34"},
		{"half-even", "127.35", "127.34"},
		{"half-+inf", "127.35", "127.34"},
		{"up", "127.35", "127.35"},
		{"ceiling", "127.35", "127.35"},
		{"down", "127.33", "127.34"},
		{"floor", "127.33", "127.34"},
	}
	for _, c := range cases {
		t.Run(c.mode, func(t *testing.T) {
			mode := modeNamed(t, c.mode)
			wantText(t, compoundRoundedEach(t, "120.00", "1.02", 3, mode), c.each)
			wantText(t, unrounded.Round(2, mode), c.once)
		})
	}
}

func TestEM93_ZeroRateAddsNothing(t *testing.T) {
	interest := dec(t, "1000.00").Mul(dec(t, "0.00")).Mul(number.DecimalFromInt(3))
	wantEqual(t, interest, "0.00")
	wantEqual(t, dec(t, "1000.00").Add(interest), "1000.00")
}
