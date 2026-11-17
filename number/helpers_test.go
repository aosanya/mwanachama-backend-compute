package number

import "testing"

func dec(t *testing.T, text string) Decimal {
	t.Helper()
	d, err := ParseDecimal(text)
	if err != nil {
		t.Fatalf("ParseDecimal(%q): %v", text, err)
	}
	return d
}

func wantText(t *testing.T, got Decimal, want string) {
	t.Helper()
	if got.String() != want {
		t.Fatalf("got %s, want %s", got.String(), want)
	}
}

func wantValue(t *testing.T, got Decimal, want string) {
	t.Helper()
	if !got.Equal(dec(t, want)) {
		t.Fatalf("got %s, want a value equal to %s", got.String(), want)
	}
}

var allModes = []struct {
	name string
	mode Rounding
}{
	{"half-up", RoundHalfUp},
	{"half-down", RoundHalfDown},
	{"half-even", RoundHalfEven},
	{"half-+inf", RoundHalfPosInf},
	{"up", RoundUp},
	{"down", RoundDown},
	{"ceiling", RoundCeiling},
	{"floor", RoundFloor},
}
