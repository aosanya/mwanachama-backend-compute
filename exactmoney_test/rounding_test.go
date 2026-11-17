package exactmoney_test

import (
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

type roundingRow struct {
	id       string
	unscaled int64
	expected [8]int64
	negative bool
}

var halfRows = []roundingRow{
	{"EM-10", 25, [8]int64{3, 2, 2, 3, 3, 2, 3, 2}, false},
	{"EM-11", 15, [8]int64{2, 1, 2, 2, 2, 1, 2, 1}, false},
	{"EM-12", 5, [8]int64{1, 0, 0, 1, 1, 0, 1, 0}, false},
	{"EM-13", -5, [8]int64{-1, 0, 0, 0, -1, 0, 0, -1}, true},
	{"EM-14", -15, [8]int64{-2, -1, -2, -1, -2, -1, -1, -2}, true},
	{"EM-15", -25, [8]int64{-3, -2, -2, -2, -3, -2, -2, -3}, true},
	{"EM-16", 24, [8]int64{2, 2, 2, 2, 3, 2, 3, 2}, false},
	{"EM-17", 26, [8]int64{3, 3, 3, 3, 3, 2, 3, 2}, false},
	{"EM-18", -24, [8]int64{-2, -2, -2, -2, -3, -2, -2, -3}, true},
	{"EM-19", -26, [8]int64{-3, -3, -3, -3, -3, -2, -2, -3}, true},
	{"EM-20", 1005, [8]int64{101, 100, 100, 101, 101, 100, 101, 100}, false},
	{"EM-21", 1015, [8]int64{102, 101, 102, 102, 102, 101, 102, 101}, false},
	{"EM-22", 1025, [8]int64{103, 102, 102, 103, 103, 102, 103, 102}, false},
	{"EM-23", 1004, [8]int64{100, 100, 100, 100, 101, 100, 101, 100}, false},
	{"EM-24", 1006, [8]int64{101, 101, 101, 101, 101, 100, 101, 100}, false},
	{"EM-25", -1005, [8]int64{-101, -100, -100, -100, -101, -100, -100, -101}, true},
	{"EM-26", -1025, [8]int64{-103, -102, -102, -102, -103, -102, -102, -103}, true},
}

func negativeTag(row roundingRow) string {
	if row.negative {
		return " (only if negative inputs are allowed)"
	}
	return ""
}

func TestRoundingTables_B1AndB2_EveryModeAtEveryPlaceCount(t *testing.T) {
	for _, row := range halfRows {
		t.Run(row.id+negativeTag(row), func(t *testing.T) {
			perPlaces(t, func(t *testing.T, places int) {
				input := number.NewDecimal(row.unscaled, places+1)
				for i, m := range modes {
					got := input.Round(places, m.mode)
					want := number.NewDecimal(row.expected[i], places)
					if got.String() != want.String() {
						t.Errorf("%s: %s rounded to %d places under %s is %s, want %s", row.id, input, places, m.name, got, want)
					}
				}
			})
		})
	}
}

func TestRoundingTables_AtTheWrittenFigures(t *testing.T) {
	cases := []struct {
		id     string
		text   string
		places int
		mode   string
		want   string
	}{
		{"EM-10", "2.5", 0, "half-up", "3"},
		{"EM-10", "2.5", 0, "half-even", "2"},
		{"EM-13", "-0.5", 0, "half-up", "-1"},
		{"EM-13", "-0.5", 0, "half-+inf", "0"},
		{"EM-20", "1.005", 2, "half-up", "1.01"},
		{"EM-20", "1.005", 2, "half-even", "1.00"},
		{"EM-21", "1.015", 2, "half-even", "1.02"},
		{"EM-22", "1.025", 2, "half-even", "1.02"},
		{"EM-25", "-1.005", 2, "half-up", "-1.01"},
		{"EM-26", "-1.025", 2, "half-down", "-1.02"},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %s %s", c.id, c.text, c.mode), func(t *testing.T) {
			wantText(t, dec(t, c.text).Round(c.places, modeNamed(t, c.mode)), c.want)
		})
	}
}

func TestEM27_ZeroHasNoSign(t *testing.T) {
	input := dec(t, "-0.004")
	for _, name := range []string{"half-up", "half-down", "half-even", "half-+inf", "down", "ceiling"} {
		t.Run(name+" (only if negative inputs are allowed)", func(t *testing.T) {
			got := input.Round(2, modeNamed(t, name))
			wantText(t, got, "0.00")
			if got.Sign() != 0 {
				t.Fatalf("sign of zero is %d", got.Sign())
			}
		})
	}
	for _, name := range []string{"up", "floor"} {
		t.Run(name+" (only if negative inputs are allowed)", func(t *testing.T) {
			wantText(t, input.Round(2, modeNamed(t, name)), "-0.01")
		})
	}
}

func TestEM101_AHalfBelowZeroRoundsAsTheEM13Row(t *testing.T) {
	for i, m := range modes {
		t.Run(m.name+" (only if negative inputs are allowed)", func(t *testing.T) {
			want := number.DecimalFromInt(halfRows[3].expected[i]).String()
			wantText(t, dec(t, "-0.5").Round(0, m.mode), want)
		})
	}
}
