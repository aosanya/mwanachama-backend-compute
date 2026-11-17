package exactmoney_test

import (
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

var placeCounts = []int{0, 1, 2, 3, 4}

var modes = []struct {
	name string
	mode number.Rounding
}{
	{"half-up", number.RoundHalfUp},
	{"half-down", number.RoundHalfDown},
	{"half-even", number.RoundHalfEven},
	{"half-+inf", number.RoundHalfPosInf},
	{"up", number.RoundUp},
	{"down", number.RoundDown},
	{"ceiling", number.RoundCeiling},
	{"floor", number.RoundFloor},
}

func modeNamed(t *testing.T, name string) number.Rounding {
	t.Helper()
	for _, m := range modes {
		if m.name == name {
			return m.mode
		}
	}
	t.Fatalf("no rounding mode named %q", name)
	return 0
}

func dec(t *testing.T, text string) number.Decimal {
	t.Helper()
	d, err := number.ParseDecimal(text)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", text, err)
	}
	return d
}

func wantText(t *testing.T, got number.Decimal, want string) {
	t.Helper()
	if got.String() != want {
		t.Fatalf("got %s, want %s", got.String(), want)
	}
}

func wantEqual(t *testing.T, got number.Decimal, want string) {
	t.Helper()
	if !got.Equal(dec(t, want)) {
		t.Fatalf("got %s, want a value equal to %s", got.String(), want)
	}
}

func units(count int64, places int) number.Decimal {
	return number.NewDecimal(count, places)
}

func perPlaces(t *testing.T, run func(t *testing.T, places int)) {
	t.Helper()
	for _, places := range placeCounts {
		t.Run(fmt.Sprintf("places=%d", places), func(t *testing.T) { run(t, places) })
	}
}

func sumOf(parts ...number.Decimal) number.Decimal {
	total := number.NewDecimal(0, 0)
	for _, part := range parts {
		total = total.Add(part)
	}
	return total
}
