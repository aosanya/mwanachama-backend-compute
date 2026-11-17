package number

import "testing"

func TestEM52And53_SimpleInterestHasNoRounding(t *testing.T) {
	yearly := dec(t, "1000.00").Mul(dec(t, "0.05"))
	wantValue(t, yearly, "50.00")
	interest := yearly.Mul(dec(t, "3"))
	wantValue(t, interest, "150.00")
	wantValue(t, dec(t, "1000.00").Add(interest), "1150.00")
	second := dec(t, "2500.00").Mul(dec(t, "0.045"))
	wantValue(t, second, "112.50")
	wantValue(t, second.Mul(dec(t, "2")), "225.00")
}

func TestEM54_CompoundCarriesEveryDigitUntilTheEnd(t *testing.T) {
	value := dec(t, "1000.00")
	for year := 0; year < 3; year++ {
		value = value.Mul(dec(t, "1.05"))
	}
	wantValue(t, value, "1157.625")
	want := map[string]string{
		"half-up": "1157.63", "half-down": "1157.62", "half-even": "1157.62", "half-+inf": "1157.63",
		"up": "1157.63", "down": "1157.62", "ceiling": "1157.63", "floor": "1157.62",
	}
	for _, m := range allModes {
		wantText(t, value.Round(2, m.mode), want[m.name])
	}
}

func compound(t *testing.T, start, factor string, periods int, mode Rounding, roundEach bool) Decimal {
	t.Helper()
	value := dec(t, start)
	for i := 0; i < periods; i++ {
		value = value.Mul(dec(t, factor))
		if roundEach {
			value = value.Round(2, mode)
		}
	}
	return value.Round(2, mode)
}

func TestEM55_RoundingEachPeriodAgainstOnceAtTheEnd(t *testing.T) {
	unrounded := dec(t, "500.00")
	for i := 0; i < 4; i++ {
		unrounded = unrounded.Mul(dec(t, "1.015"))
	}
	wantValue(t, unrounded, "530.6817753125")
	for _, mode := range []Rounding{RoundHalfUp, RoundHalfDown, RoundHalfEven} {
		wantText(t, compound(t, "500.00", "1.015", 4, mode, true), "530.68")
		wantText(t, compound(t, "500.00", "1.015", 4, mode, false), "530.68")
	}
	eachAndEnd := map[Rounding][2]string{
		RoundUp:      {"530.70", "530.69"},
		RoundCeiling: {"530.70", "530.69"},
		RoundDown:    {"530.67", "530.68"},
		RoundFloor:   {"530.67", "530.68"},
	}
	for mode, want := range eachAndEnd {
		wantText(t, compound(t, "500.00", "1.015", 4, mode, true), want[0])
		wantText(t, compound(t, "500.00", "1.015", 4, mode, false), want[1])
	}
}

func TestEM56_TheTwoApproachesDifferByACent(t *testing.T) {
	for _, mode := range []Rounding{RoundHalfUp, RoundHalfDown, RoundHalfEven, RoundHalfPosInf} {
		wantText(t, compound(t, "120.00", "1.02", 3, mode, true), "127.35")
		wantText(t, compound(t, "120.00", "1.02", 3, mode, false), "127.34")
	}
	for _, mode := range []Rounding{RoundUp, RoundCeiling} {
		wantText(t, compound(t, "120.00", "1.02", 3, mode, true), "127.35")
		wantText(t, compound(t, "120.00", "1.02", 3, mode, false), "127.35")
	}
	for _, mode := range []Rounding{RoundDown, RoundFloor} {
		wantText(t, compound(t, "120.00", "1.02", 3, mode, true), "127.33")
		wantText(t, compound(t, "120.00", "1.02", 3, mode, false), "127.34")
	}
}
