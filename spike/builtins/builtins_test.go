package builtins

import (
	"math"
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func decimals(texts ...string) []numeric.Decimal {
	values := make([]numeric.Decimal, len(texts))
	for i, text := range texts {
		values[i] = numeric.MustParse(text)
	}
	return values
}

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

func TestSumDecimals(t *testing.T) {
	if got := SumDecimals(decimals("500.00", "750.00", "1250.50")).String(); got != "2500.50" {
		t.Errorf("got %s", got)
	}
	if got := SumDecimals(decimals("0.1", "0.2")).String(); got != "0.3" {
		t.Errorf("got %s", got)
	}
}

func TestSumOfNoDecimalsIsZero(t *testing.T) {
	if got := SumDecimals(nil); got.Sign() != 0 {
		t.Errorf("got %s", got)
	}
	if got := SumDecimals(nil).Add(numeric.MustParse("2500.50")).String(); got != "2500.50" {
		t.Errorf("got %s", got)
	}
}

func TestSumInts(t *testing.T) {
	if got := SumInts([]int64{1200, 830, 1500, 477}); got != 4007 {
		t.Errorf("got %d", got)
	}
	if got := SumInts(nil); got != 0 {
		t.Errorf("got %d", got)
	}
}

func TestAvgDecimalsGivesDecimal(t *testing.T) {
	got, err := AvgDecimals(decimals("420.00", "380.50", "95.25", "510.00", "440.75", "300.00", "465.50", "389.00"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Round(3).String() != "375.125" {
		t.Errorf("got %s", got)
	}
}

func TestAvgIntsGivesFloat(t *testing.T) {
	got, err := AvgInts([]int64{72, 85, 64, 91, 77})
	if err != nil || math.Abs(got-77.8) > 1e-12 {
		t.Errorf("got %v (%v)", got, err)
	}
}

func TestAvgOfNothingFails(t *testing.T) {
	if _, err := AvgDecimals(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
	if _, err := AvgInts(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
}

func TestMinMaxDecimals(t *testing.T) {
	values := decimals("2400.50", "3100.25", "2750.00")
	low, err := MinDecimals(values)
	if err != nil || low.String() != "2400.50" {
		t.Errorf("min %s (%v)", low, err)
	}
	high, err := MaxDecimals(values)
	if err != nil || high.String() != "3100.25" {
		t.Errorf("max %s (%v)", high, err)
	}
}

func TestMinMaxDecimalsKeepTheElementAsWritten(t *testing.T) {
	high, _ := MaxDecimals(decimals("0.00", "-27.24"))
	if high.String() != "0.00" {
		t.Errorf("got %s", high)
	}
	low, _ := MinDecimals(decimals("1.0", "1.00"))
	if low.String() != "1.0" {
		t.Errorf("got %s", low)
	}
}

func TestMinMaxInts(t *testing.T) {
	values := []int64{72, 85, 64, 91, 77}
	if low, err := MinInts(values); err != nil || low != 64 {
		t.Errorf("min %d (%v)", low, err)
	}
	if high, err := MaxInts(values); err != nil || high != 91 {
		t.Errorf("max %d (%v)", high, err)
	}
	if high, _ := MaxInts([]int64{0, -3}); high != 0 {
		t.Errorf("max %d", high)
	}
}

func TestMinMaxOfNothingFail(t *testing.T) {
	if _, err := MinDecimals(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
	if _, err := MaxDecimals(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
	if _, err := MinInts(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
	if _, err := MaxInts(nil); err != ErrEmptyList {
		t.Errorf("got %v", err)
	}
}

func TestCount(t *testing.T) {
	if got := CountDecimals(decimals("1.00", "2.00", "3.00")); got != 3 {
		t.Errorf("got %d", got)
	}
	if got := CountInts([]int64{1, 2}); got != 2 {
		t.Errorf("got %d", got)
	}
	if CountInts(nil) != 0 || CountDecimals(nil) != 0 {
		t.Error("empty lists count zero")
	}
}

func TestDaysBetween(t *testing.T) {
	cases := []struct {
		from, to time.Time
		want     int64
	}{
		{day(2026, 3, 31), day(2026, 5, 15), 45},
		{day(2026, 6, 12), day(2026, 6, 30), 18},
		{day(2026, 3, 31), day(2026, 3, 20), -11},
		{day(2026, 3, 31), day(2026, 3, 31), 0},
		{day(2024, 2, 28), day(2024, 3, 1), 2},
	}
	for _, c := range cases {
		if got := DaysBetween(c.from, c.to); got != c.want {
			t.Errorf("DaysBetween(%v, %v) = %d, want %d", c.from, c.to, got, c.want)
		}
	}
}

func TestDaysBetweenIgnoresTimeOfDayAndZone(t *testing.T) {
	late := time.Date(2026, 3, 31, 23, 59, 0, 0, time.UTC)
	early := time.Date(2026, 4, 1, 0, 1, 0, 0, time.UTC)
	if got := DaysBetween(late, early); got != 1 {
		t.Errorf("got %d", got)
	}
	plusFive := time.FixedZone("plus5", 5*3600)
	local := time.Date(2026, 4, 1, 2, 0, 0, 0, plusFive)
	if got := DaysBetween(day(2026, 3, 31), local); got != 0 {
		t.Errorf("got %d", got)
	}
}

func TestMonthsBetween(t *testing.T) {
	cases := []struct {
		from, to time.Time
		want     int64
	}{
		{day(2026, 1, 20), day(2026, 9, 15), 7},
		{day(2026, 1, 20), day(2026, 9, 20), 8},
		{day(2026, 5, 1), day(2026, 9, 15), 4},
		{day(2026, 1, 31), day(2026, 2, 28), 0},
		{day(2025, 12, 15), day(2026, 1, 14), 0},
		{day(2025, 12, 15), day(2026, 1, 15), 1},
		{day(2026, 9, 15), day(2026, 1, 20), -7},
		{day(2026, 3, 3), day(2026, 3, 3), 0},
	}
	for _, c := range cases {
		if got := MonthsBetween(c.from, c.to); got != c.want {
			t.Errorf("MonthsBetween(%v, %v) = %d, want %d", c.from.Format("2006-01-02"), c.to.Format("2006-01-02"), got, c.want)
		}
	}
}

func TestYearsBetween(t *testing.T) {
	cases := []struct {
		from, to time.Time
		want     int64
	}{
		{day(2023, 9, 10), day(2026, 11, 17), 3},
		{day(2023, 9, 10), day(2026, 9, 9), 2},
		{day(2023, 9, 10), day(2026, 9, 10), 3},
		{day(2023, 9, 10), day(2033, 9, 10), 10},
		{day(2026, 9, 10), day(2023, 9, 10), -3},
		{day(2026, 1, 1), day(2026, 12, 31), 0},
	}
	for _, c := range cases {
		if got := YearsBetween(c.from, c.to); got != c.want {
			t.Errorf("YearsBetween(%v, %v) = %d, want %d", c.from.Format("2006-01-02"), c.to.Format("2006-01-02"), got, c.want)
		}
	}
}

func TestToFloat(t *testing.T) {
	if got := ToFloat(numeric.MustParse("84500.00")); got != 84500 {
		t.Errorf("got %v", got)
	}
}

func TestToDecimal(t *testing.T) {
	got, err := ToDecimal(0.37, 2)
	if err != nil || got.String() != "0.37" {
		t.Errorf("got %s (%v)", got, err)
	}
	got, err = ToDecimal(0.2, 2)
	if err != nil || got.String() != "0.20" {
		t.Errorf("got %s (%v)", got, err)
	}
	if _, err := ToDecimal(math.NaN(), 2); err == nil {
		t.Error("NaN should fail")
	}
}

func TestRoundDecimal(t *testing.T) {
	if got := RoundDecimal(numeric.MustParse("372.761"), 2).String(); got != "372.76" {
		t.Errorf("got %s", got)
	}
	if got := RoundDecimal(numeric.MustParse("-0.005"), 2).String(); got != "-0.01" {
		t.Errorf("got %s", got)
	}
}

func TestRoundFloat(t *testing.T) {
	cases := []struct {
		in     float64
		places int64
		want   float64
	}{
		{84500.0 / 62000.0, 3, 1.363},
		{0.125, 2, 0.13},
		{-0.125, 2, -0.13},
		{2.5, 0, 3},
		{1.0, 3, 1.0},
	}
	for _, c := range cases {
		if got := RoundFloat(c.in, c.places); got != c.want {
			t.Errorf("RoundFloat(%v, %d) = %v, want %v", c.in, c.places, got, c.want)
		}
	}
	if got := RoundFloat(math.Inf(1), 2); !math.IsInf(got, 1) {
		t.Errorf("got %v", got)
	}
	if got := RoundFloat(math.NaN(), 2); !math.IsNaN(got) {
		t.Errorf("got %v", got)
	}
}
