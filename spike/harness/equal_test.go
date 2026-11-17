package harness

import (
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func TestEqualComparesDecimalsByText(t *testing.T) {
	if !Equal(dec("75.25"), dec("75.25")) {
		t.Error("same text should be equal")
	}
	if Equal(dec("75.25"), dec("75.250")) {
		t.Error("75.25 and 75.250 should differ")
	}
	if Equal(dec("75.25"), 75.25) {
		t.Error("decimal and float should differ")
	}
}

func TestEqualComparesFloatsWithinTolerance(t *testing.T) {
	if !Equal(1.363, 1.363+1e-10) {
		t.Error("1e-10 apart should be equal")
	}
	if Equal(1.363, 1.363+1e-8) {
		t.Error("1e-8 apart should differ")
	}
	if Equal(1.0, int64(1)) {
		t.Error("float and integer should differ")
	}
}

func TestEqualComparesTheOtherKindsByValue(t *testing.T) {
	cases := []struct {
		name      string
		want, got any
		equal     bool
	}{
		{"integers", int64(45), int64(45), true},
		{"different integers", int64(45), int64(46), false},
		{"int and int64", int64(45), 45, false},
		{"text", "B", "B", true},
		{"different text", "B", "C", false},
		{"booleans", true, true, true},
		{"different booleans", true, false, false},
		{"dates", date("2026-03-31"), date("2026-03-31"), true},
		{"different dates", date("2026-03-31"), date("2026-04-01"), false},
		{"same date in another zone", date("2026-03-31"), time.Date(2026, 3, 31, 5, 0, 0, 0, time.FixedZone("x", 5*3600)), true},
		{"decimal lists", decs("1.00", "2.00"), decs("1.00", "2.00"), true},
		{"decimal lists by text", decs("1.00"), decs("1.0"), false},
		{"decimal lists by length", decs("1.00"), decs("1.00", "2.00"), false},
		{"integer lists", []int64{1, 2}, []int64{1, 2}, true},
		{"different integer lists", []int64{1, 2}, []int64{2, 1}, false},
		{"nil got", int64(1), nil, false},
	}
	for _, c := range cases {
		if got := Equal(c.want, c.got); got != c.equal {
			t.Errorf("%s: Equal = %v, want %v", c.name, got, c.equal)
		}
	}
}

func TestDescribeNamesTheKind(t *testing.T) {
	cases := map[string]any{
		"75.25 (decimal)":             dec("75.25"),
		"1.363 (float)":               1.363,
		"45 (integer)":                int64(45),
		`"B" (text)`:                  "B",
		"true (boolean)":              true,
		"2026-03-31 (date)":           date("2026-03-31"),
		"[1.00, 2.50] (decimal list)": decs("1.00", "2.50"),
		"[1 2] (integer list)":        []int64{1, 2},
		"7 (int)":                     7,
		"<nil> (<nil>)":               nil,
		"[] (decimal list)":           []numeric.Decimal{},
	}
	for want, v := range cases {
		if got := Describe(v); got != want {
			t.Errorf("Describe = %q, want %q", got, want)
		}
	}
}

func TestInputKeyIgnoresMapOrderAndSeesScale(t *testing.T) {
	a := InputKey(inputs{"balance": dec("1.0"), "payment": dec("2.0")})
	b := InputKey(inputs{"payment": dec("2.0"), "balance": dec("1.0")})
	if a != b {
		t.Errorf("%q != %q", a, b)
	}
	if a == InputKey(inputs{"balance": dec("1.00"), "payment": dec("2.0")}) {
		t.Error("scale should change the key")
	}
}

func TestKindOfRejectsUnknownTypes(t *testing.T) {
	if _, ok := KindOf(7); ok {
		t.Error("int is not a value type")
	}
	if _, ok := KindOf([]string{"a"}); ok {
		t.Error("string list is not a value type")
	}
	if Kind(0).String() != "unknown" || KindDecimalList.String() != "decimal list" {
		t.Error("kind names are wrong")
	}
}
