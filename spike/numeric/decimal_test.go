package numeric

import (
	"math"
	"testing"
)

func TestParseAndStringRoundTrip(t *testing.T) {
	for _, text := range []string{"0", "0.00", "75.25", "-75.25", "-0.05", "120.50", "1000", "0.000000000000000001", "123456789012345678901234567890.123456789", "-9223372036854775808", "9223372036854775807.00"} {
		d, err := Parse(text)
		if err != nil {
			t.Fatalf("Parse(%q): %v", text, err)
		}
		if d.String() != text {
			t.Errorf("round trip of %q gave %q", text, d.String())
		}
	}
}

func TestParseKeepsTrailingZerosAsPlaces(t *testing.T) {
	d := MustParse("75.250")
	if d.DecimalPlaces() != 3 || d.String() != "75.250" {
		t.Errorf("got %s with %d places", d, d.DecimalPlaces())
	}
}

func TestParseRejectsBadText(t *testing.T) {
	for _, text := range []string{"", "-", ".5", "5.", "1.2.3", "abc", "1e5", "--1", "+-1", "1 ", "0x10"} {
		if _, err := Parse(text); err == nil {
			t.Errorf("Parse(%q) should fail", text)
		}
	}
}

func TestZeroValueIsZero(t *testing.T) {
	var d Decimal
	if d.String() != "0" || d.IsBig() || d.Sign() != 0 {
		t.Errorf("zero value is %s", d)
	}
}

func TestFromInt(t *testing.T) {
	if got := FromInt(-42).String(); got != "-42" {
		t.Errorf("got %s", got)
	}
}

func TestPointOneAndPointTwoMakePointThreeExactly(t *testing.T) {
	sum := MustParse("0.1").Add(MustParse("0.2"))
	if sum.Cmp(MustParse("0.3")) != 0 || sum.String() != "0.3" {
		t.Errorf("0.1 + 0.2 = %s", sum)
	}
}

func TestAddAlignsScales(t *testing.T) {
	if got := MustParse("1.5").Add(MustParse("2.25")).String(); got != "3.75" {
		t.Errorf("got %s", got)
	}
	if got := MustParse("0.30").Sub(MustParse("0.10")).String(); got != "0.20" {
		t.Errorf("got %s", got)
	}
}

func TestSubGoesNegative(t *testing.T) {
	if got := MustParse("45.25").Sub(MustParse("120.50")).String(); got != "-75.25" {
		t.Errorf("got %s", got)
	}
}

func TestMulSumsScales(t *testing.T) {
	got := MustParse("853").Mul(MustParse("0.437"))
	if got.String() != "372.761" || got.DecimalPlaces() != 3 {
		t.Errorf("got %s", got)
	}
	got = MustParse("8400.00").Mul(MustParse("0.05"))
	if got.String() != "420.0000" {
		t.Errorf("got %s", got)
	}
}

func TestMulSignCombinations(t *testing.T) {
	cases := map[[2]string]string{
		{"-2.5", "4"}:    "-10.0",
		{"-2.5", "-4"}:   "10.0",
		{"2.5", "-4"}:    "-10.0",
		{"0", "-4.00"}:   "0.00",
		{"-0.01", "0.5"}: "-0.005",
	}
	for operands, want := range cases {
		if got := MustParse(operands[0]).Mul(MustParse(operands[1])).String(); got != want {
			t.Errorf("%s * %s = %s, want %s", operands[0], operands[1], got, want)
		}
	}
}

func TestMulOverflowMovesToBigAndSubDemotesBack(t *testing.T) {
	a := MustParse("9999999999.99")
	product := a.Mul(a)
	if !product.IsBig() {
		t.Fatalf("product %s should sit in the big fallback", product)
	}
	if product.String() != "99999999999800000000.0001" {
		t.Errorf("product is %s", product)
	}
	back := product.Add(MustParse("1")).Sub(product)
	if back.IsBig() || back.String() != "1.0000" {
		t.Errorf("difference %s big=%v", back, back.IsBig())
	}
}

func TestMulOverflowAtExactBoundary(t *testing.T) {
	limit := MustParse("9223372036854775807")
	if !limit.Mul(FromInt(2)).IsBig() {
		t.Error("max int64 times 2 should be big")
	}
	if limit.Mul(FromInt(1)).IsBig() {
		t.Error("max int64 times 1 should stay fast")
	}
	minimum := MustParse("-9223372036854775808")
	if minimum.Mul(FromInt(1)).IsBig() {
		t.Error("min int64 times 1 should stay fast")
	}
	if !minimum.Mul(FromInt(-1)).IsBig() {
		t.Error("min int64 times -1 should be big")
	}
	if got := minimum.Mul(FromInt(-1)).String(); got != "9223372036854775808" {
		t.Errorf("got %s", got)
	}
}

func TestMulScaleBeyondEighteenIsBig(t *testing.T) {
	a := MustParse("0.1234567890")
	got := a.Mul(a)
	if !got.IsBig() || got.DecimalPlaces() != 20 || got.String() != "0.01524157875019052100" {
		t.Errorf("got %s places %d big %v", got, got.DecimalPlaces(), got.IsBig())
	}
}

func TestAddOverflowMovesToBigAndSubDemotesBack(t *testing.T) {
	top := MustParse("92233720368547758.07")
	sum := top.Add(MustParse("0.01"))
	if !sum.IsBig() || sum.String() != "92233720368547758.08" {
		t.Fatalf("sum %s big=%v", sum, sum.IsBig())
	}
	back := sum.Sub(MustParse("0.01"))
	if back.IsBig() || back.Cmp(top) != 0 {
		t.Errorf("back %s big=%v", back, back.IsBig())
	}
}

func TestAddOverflowWhileAligningScales(t *testing.T) {
	large := MustParse("9223372036854775807")
	sum := large.Add(MustParse("0.5"))
	if !sum.IsBig() || sum.String() != "9223372036854775807.5" {
		t.Errorf("got %s", sum)
	}
}

func TestNegAtTheEdge(t *testing.T) {
	minimum := MustParse("-9223372036854775808")
	flipped := minimum.Neg()
	if !flipped.IsBig() || flipped.String() != "9223372036854775808" {
		t.Errorf("got %s", flipped)
	}
	if flipped.Neg().IsBig() {
		t.Error("negating back should demote")
	}
	if got := MustParse("2.50").Neg().String(); got != "-2.50" {
		t.Errorf("got %s", got)
	}
}

func TestCmp(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0", "1.00", 0},
		{"1.01", "1.0", 1},
		{"-1", "1", -1},
		{"0.30", "0.3", 0},
		{"92233720368547758.08", "92233720368547758.07", 1},
		{"9223372036854775807", "9223372036854775807.0000000001", -1},
	}
	for _, c := range cases {
		if got := MustParse(c.a).Cmp(MustParse(c.b)); got != c.want {
			t.Errorf("Cmp(%s, %s) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestRoundHalfAwayFromZero(t *testing.T) {
	cases := []struct {
		in     string
		places int
		want   string
	}{
		{"2.345", 2, "2.35"},
		{"2.344", 2, "2.34"},
		{"-2.345", 2, "-2.35"},
		{"-2.344", 2, "-2.34"},
		{"0.5", 0, "1"},
		{"-0.5", 0, "-1"},
		{"1.5", 0, "2"},
		{"2.5", 0, "3"},
		{"-2.5", 0, "-3"},
		{"0.004", 2, "0.00"},
		{"-0.004", 2, "0.00"},
		{"-0.005", 2, "-0.01"},
		{"372.761", 2, "372.76"},
		{"48.26175", 2, "48.26"},
		{"61105.500000", 2, "61105.50"},
		{"75.2", 2, "75.20"},
		{"7", 2, "7.00"},
	}
	for _, c := range cases {
		if got := MustParse(c.in).Round(c.places).String(); got != c.want {
			t.Errorf("Round(%s, %d) = %s, want %s", c.in, c.places, got, c.want)
		}
	}
}

func TestRoundKeepsThePlacesRequested(t *testing.T) {
	if got := MustParse("5").Round(3).DecimalPlaces(); got != 3 {
		t.Errorf("got %d places", got)
	}
}

func TestRoundOfBigDemotesWhenItFits(t *testing.T) {
	big := MustParse("60000000.00").Mul(MustParse("1.000000000000000001"))
	if !big.IsBig() {
		t.Fatalf("expected big, got %s", big)
	}
	rounded := big.Round(2)
	if rounded.IsBig() || rounded.String() != "60000000.00" {
		t.Errorf("rounded %s big=%v", rounded, rounded.IsBig())
	}
}

func TestRoundNegativePlacesActAsZero(t *testing.T) {
	if got := MustParse("2.5").Round(-1).String(); got != "3" {
		t.Errorf("got %s", got)
	}
}

func TestDivisionCarriesEighteenPlaces(t *testing.T) {
	got, err := MustParse("2300.25").Div(FromInt(7))
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "328.607142857142857143" || got.DecimalPlaces() != 18 {
		t.Errorf("got %s", got)
	}
	if rounded := got.Round(2); rounded.String() != "328.61" || rounded.IsBig() {
		t.Errorf("rounded %s big=%v", rounded, rounded.IsBig())
	}
}

func TestDivisionOfSixtyMillionByFourThousandAndSeven(t *testing.T) {
	got, err := MustParse("60000000.00").Div(FromInt(4007))
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "14973.795857249812827552" {
		t.Errorf("got %s", got)
	}
	if got.Round(2).String() != "14973.80" {
		t.Errorf("rounded %s", got.Round(2))
	}
}

func TestDivisionSignsAndExactResults(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"1", "3", "0.333333333333333333"},
		{"-1", "3", "-0.333333333333333333"},
		{"1", "-3", "-0.333333333333333333"},
		{"2", "3", "0.666666666666666667"},
		{"-2", "3", "-0.666666666666666667"},
		{"10.00", "4", "2.500000000000000000"},
		{"0", "7", "0.000000000000000000"},
		{"1.000000000000000000", "0.001", "1000.000000000000000000"},
	}
	for _, c := range cases {
		got, err := MustParse(c.a).Div(MustParse(c.b))
		if err != nil || got.String() != c.want {
			t.Errorf("%s / %s = %s (%v), want %s", c.a, c.b, got, err, c.want)
		}
	}
}

func TestDivisionOfHighScaleDividend(t *testing.T) {
	a := MustParse("1").Mul(MustParse("0.1234567890")).Mul(MustParse("0.1234567890"))
	got, err := a.Div(MustParse("0.5"))
	if err != nil || got.String() != "0.030483157500381042" {
		t.Errorf("got %s (%v)", got, err)
	}
}

func TestDivisionByZero(t *testing.T) {
	if _, err := FromInt(1).Div(MustParse("0.00")); err != ErrDivideByZero {
		t.Errorf("got %v", err)
	}
}

func TestToFloat(t *testing.T) {
	if got := MustParse("0.37").ToFloat(); got != 0.37 {
		t.Errorf("got %v", got)
	}
	if got := MustParse("-84500.00").ToFloat(); got != -84500 {
		t.Errorf("got %v", got)
	}
}

func TestFromFloatRoundsHalfAwayFromZero(t *testing.T) {
	cases := []struct {
		in     float64
		places int
		want   string
	}{
		{0.37, 2, "0.37"},
		{0.125, 2, "0.13"},
		{-0.125, 2, "-0.13"},
		{2.5, 0, "3"},
		{-2.5, 0, "-3"},
		{1.363, 3, "1.363"},
		{0.1 + 0.2, 2, "0.30"},
		{5, 2, "5.00"},
		{1e21, 0, "1000000000000000000000"},
		{math.Copysign(0, -1), 2, "0.00"},
	}
	for _, c := range cases {
		got, err := FromFloat(c.in, c.places)
		if err != nil || got.String() != c.want {
			t.Errorf("FromFloat(%v, %d) = %s (%v), want %s", c.in, c.places, got, err, c.want)
		}
	}
}

func TestFromFloatRejectsNotFinite(t *testing.T) {
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := FromFloat(f, 2); err != ErrNotFinite {
			t.Errorf("FromFloat(%v) gave %v", f, err)
		}
	}
}

func TestSignAndPlaces(t *testing.T) {
	if MustParse("-0.01").Sign() != -1 || MustParse("0.01").Sign() != 1 || MustParse("0.00").Sign() != 0 {
		t.Error("sign is wrong")
	}
	if MustParse("1.230").DecimalPlaces() != 3 {
		t.Error("places are wrong")
	}
}

func TestIsBigIsTrueOnlyInFallback(t *testing.T) {
	if MustParse("12.50").IsBig() {
		t.Error("small value should not be big")
	}
	if !MustParse("123456789012345678901234567890").IsBig() {
		t.Error("30 digits should be big")
	}
	if !MustParse("0.0000000000000000001").IsBig() {
		t.Error("19 places should be big")
	}
}
