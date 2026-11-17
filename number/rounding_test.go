package number

import "testing"

type roundingCase struct {
	id     string
	input  string
	places int
	want   map[string]string
}

func table(halfUp, halfDown, halfEven, halfPosInf, up, down, ceiling, floor string) map[string]string {
	return map[string]string{
		"half-up": halfUp, "half-down": halfDown, "half-even": halfEven, "half-+inf": halfPosInf,
		"up": up, "down": down, "ceiling": ceiling, "floor": floor,
	}
}

func TestEM10To19_RoundingToZeroPlaces(t *testing.T) {
	cases := []roundingCase{
		{"EM-10", "2.5", 0, table("3", "2", "2", "3", "3", "2", "3", "2")},
		{"EM-11", "1.5", 0, table("2", "1", "2", "2", "2", "1", "2", "1")},
		{"EM-12", "0.5", 0, table("1", "0", "0", "1", "1", "0", "1", "0")},
		{"EM-13", "-0.5", 0, table("-1", "0", "0", "0", "-1", "0", "0", "-1")},
		{"EM-14", "-1.5", 0, table("-2", "-1", "-2", "-1", "-2", "-1", "-1", "-2")},
		{"EM-15", "-2.5", 0, table("-3", "-2", "-2", "-2", "-3", "-2", "-2", "-3")},
		{"EM-16", "2.4", 0, table("2", "2", "2", "2", "3", "2", "3", "2")},
		{"EM-17", "2.6", 0, table("3", "3", "3", "3", "3", "2", "3", "2")},
		{"EM-18", "-2.4", 0, table("-2", "-2", "-2", "-2", "-3", "-2", "-2", "-3")},
		{"EM-19", "-2.6", 0, table("-3", "-3", "-3", "-3", "-3", "-2", "-2", "-3")},
	}
	runRoundingCases(t, cases)
}

func TestEM20To26_RoundingToTwoPlacesWithoutTheFloatTrap(t *testing.T) {
	cases := []roundingCase{
		{"EM-20", "1.005", 2, table("1.01", "1.00", "1.00", "1.01", "1.01", "1.00", "1.01", "1.00")},
		{"EM-21", "1.015", 2, table("1.02", "1.01", "1.02", "1.02", "1.02", "1.01", "1.02", "1.01")},
		{"EM-22", "1.025", 2, table("1.03", "1.02", "1.02", "1.03", "1.03", "1.02", "1.03", "1.02")},
		{"EM-23", "1.004", 2, table("1.00", "1.00", "1.00", "1.00", "1.01", "1.00", "1.01", "1.00")},
		{"EM-24", "1.006", 2, table("1.01", "1.01", "1.01", "1.01", "1.01", "1.00", "1.01", "1.00")},
		{"EM-25", "-1.005", 2, table("-1.01", "-1.00", "-1.00", "-1.00", "-1.01", "-1.00", "-1.00", "-1.01")},
		{"EM-26", "-1.025", 2, table("-1.03", "-1.02", "-1.02", "-1.02", "-1.03", "-1.02", "-1.02", "-1.03")},
	}
	runRoundingCases(t, cases)
}

func runRoundingCases(t *testing.T, cases []roundingCase) {
	t.Helper()
	for _, c := range cases {
		for _, m := range allModes {
			got := dec(t, c.input).Round(c.places, m.mode)
			if got.String() != c.want[m.name] {
				t.Errorf("%s %s to %d places under %s: got %s, want %s", c.id, c.input, c.places, m.name, got, c.want[m.name])
			}
		}
	}
}

func TestEM27_NegativeFractionBelowTheHalfLandsOnZeroOrOneUnit(t *testing.T) {
	for _, m := range allModes {
		got := dec(t, "-0.004").Round(2, m.mode)
		want := "0.00"
		if m.name == "up" || m.name == "floor" {
			want = "-0.01"
		}
		if got.String() != want {
			t.Errorf("-0.004 under %s: got %s, want %s", m.name, got, want)
		}
	}
}

func TestRoundingPadsWithZeroesWhenPlacesGrow(t *testing.T) {
	wantText(t, dec(t, "1.5").Round(3, RoundHalfUp), "1.500")
	wantText(t, dec(t, "7").Round(2, RoundDown), "7.00")
}

func TestEM95_HalfACentIsARoundingCase(t *testing.T) {
	half := dec(t, "0.01").Mul(dec(t, "0.5"))
	wantValue(t, half, "0.005")
	wantText(t, half.Round(2, RoundHalfUp), "0.01")
	wantText(t, half.Round(2, RoundHalfDown), "0.00")
	wantText(t, half.Round(2, RoundHalfEven), "0.00")
}

func TestEM97_OneThirdKeepsTheSetDigits(t *testing.T) {
	third, err := dec(t, "1").Div(dec(t, "3"), 20, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, third, "0.33333333333333333333")
	wantText(t, third.Round(2, RoundHalfUp), "0.33")
	wantText(t, third.Mul(dec(t, "3")), "0.99999999999999999999")
}

func TestEM98_TwoThirdsUnderEachMode(t *testing.T) {
	two, err := dec(t, "2").Div(dec(t, "3"), 20, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"half-up": "0.67", "up": "0.67", "ceiling": "0.67", "down": "0.66", "floor": "0.66"}
	for _, m := range allModes {
		expected, listed := want[m.name]
		if !listed {
			continue
		}
		wantText(t, two.Round(2, m.mode), expected)
	}
}

func TestDivisionKeepsExactlyTheDigitsAsked(t *testing.T) {
	got, err := dec(t, "10.00").Div(dec(t, "3"), 20, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, got, "3.33333333333333333333")
	short, err := dec(t, "1").Div(dec(t, "3"), 4, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, short, "0.3333")
	exact, err := dec(t, "10.00").Div(dec(t, "4"), 5, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, exact, "2.50000")
}

func TestDivisionRoundsTheKeptDigitsByTheMode(t *testing.T) {
	up, _ := dec(t, "2").Div(dec(t, "3"), 3, RoundHalfUp)
	down, _ := dec(t, "2").Div(dec(t, "3"), 3, RoundDown)
	wantText(t, up, "0.667")
	wantText(t, down, "0.666")
}

func TestDivisionByZeroIsRefused(t *testing.T) {
	if _, err := dec(t, "1").Div(dec(t, "0.00"), 20, RoundHalfUp); err == nil {
		t.Fatal("division by zero must be refused")
	}
}

func TestNegativeDivisionRoundsAwayFromZeroAtTheHalf(t *testing.T) {
	got, err := dec(t, "-1").Div(dec(t, "8"), 2, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	wantText(t, got, "-0.13")
}

func TestEM99_SmallestUnitFollowsThePlaces(t *testing.T) {
	wantText(t, dec(t, "0.0001").Add(dec(t, "0.0001")), "0.0002")
	wantValue(t, dec(t, "0.0001").Mul(dec(t, "0.5")), "0.00005")
}

func TestRoundingNamesAreStable(t *testing.T) {
	names := map[Rounding]string{
		RoundDown: "down", RoundUp: "up", RoundHalfUp: "half-up", RoundHalfDown: "half-down", RoundHalfEven: "half-even",
		RoundHalfPosInf: "half-+inf", RoundCeiling: "ceiling", RoundFloor: "floor",
	}
	for mode, want := range names {
		if mode.String() != want {
			t.Errorf("name of %d = %s, want %s", mode, mode.String(), want)
		}
	}
}
