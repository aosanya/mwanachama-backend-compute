package number

import (
	"strings"
	"testing"
)

func TestEM01_PointOnePlusPointTwoIsExactlyPointThree(t *testing.T) {
	sum := dec(t, "0.1").Add(dec(t, "0.2"))
	if !sum.Equal(dec(t, "0.3")) {
		t.Fatalf("0.1 + 0.2 = %s, want a value equal to 0.3", sum)
	}
	wantText(t, sum, "0.3")
}

func TestEM02_OneCentAddedAThousandTimesIsTen(t *testing.T) {
	total := dec(t, "0.00")
	cent := dec(t, "0.01")
	for i := 0; i < 1000; i++ {
		total = total.Add(cent)
	}
	wantText(t, total, "10.00")
}

func TestEM03_SumDoesNotDependOnOrder(t *testing.T) {
	orders := [][]string{
		{"19.99", "0.01", "5.50", "0.10", "0.20"},
		{"0.20", "0.10", "5.50", "0.01", "19.99"},
		{"5.50", "0.20", "19.99", "0.10", "0.01"},
	}
	for _, order := range orders {
		total := dec(t, "0.00")
		for _, amount := range order {
			total = total.Add(dec(t, amount))
		}
		wantText(t, total, "25.80")
	}
}

func TestEM04_ThousandAmountsSumTheSameInEveryOrder(t *testing.T) {
	amounts := make([]Decimal, 1000)
	for i := 1; i <= 1000; i++ {
		cents := int64((37*i)%1000 + 1)
		amounts[i-1] = NewDecimal(cents, 2)
	}
	sum := func(order []int) Decimal {
		total := dec(t, "0.00")
		for _, index := range order {
			total = total.Add(amounts[index])
		}
		return total
	}
	forward := make([]int, 1000)
	reverse := make([]int, 1000)
	shuffled := make([]int, 1000)
	for i := range forward {
		forward[i] = i
		reverse[i] = 999 - i
		shuffled[i] = (i * 7) % 1000
	}
	for _, order := range [][]int{forward, reverse, shuffled} {
		wantValue(t, sum(order), "5005.00")
	}
}

func TestSubtractMultiplyAndCompare(t *testing.T) {
	wantText(t, dec(t, "10.00").Sub(dec(t, "0.01")), "9.99")
	wantText(t, dec(t, "12.50").Mul(dec(t, "0.1")), "1.250")
	if dec(t, "12.5").Cmp(dec(t, "12.50")) != 0 || !dec(t, "12.5").Equal(dec(t, "12.50")) {
		t.Fatal("12.5 and 12.50 must compare equal")
	}
	if dec(t, "0.30").Cmp(dec(t, "0.31")) >= 0 || dec(t, "-1").Cmp(dec(t, "0")) >= 0 {
		t.Fatal("comparison order is wrong")
	}
}

func TestWrittenDigitsAreKept(t *testing.T) {
	wantText(t, dec(t, "0.010"), "0.010")
	wantText(t, dec(t, "12.50"), "12.50")
	if dec(t, "0.010").Scale() != 3 {
		t.Fatal("scale must be the number of written places")
	}
}

func TestParseRefusesWhatIsNotADecimal(t *testing.T) {
	for _, text := range []string{"", ".5", "5.", "+5", "1e3", "1,5", " 5", "5 ", "-", "--1", "1.2.3", "007", "abc"} {
		if _, err := ParseDecimal(text); err == nil {
			t.Fatalf("ParseDecimal(%q) must be refused", text)
		}
	}
}

func TestEM27_ZeroHasNoSign(t *testing.T) {
	wantText(t, dec(t, "-0.004").Round(2, RoundHalfUp), "0.00")
	wantText(t, dec(t, "-0.00"), "0.00")
	wantText(t, dec(t, "0.00").Neg(), "0.00")
}

func TestEM90To94_Extremes(t *testing.T) {
	wantText(t, dec(t, "0.00").Add(dec(t, "0.00")), "0.00")
	wantValue(t, dec(t, "123.45").Mul(dec(t, "0")), "0.00")
	wantText(t, dec(t, "0.01").Add(dec(t, "0.01")), "0.02")
	wantValue(t, dec(t, "1000.00").Mul(dec(t, "0")).Round(2, RoundHalfUp), "0.00")
}

func TestEM81To83And96_HugeFiguresDoNotWrapOrSaturate(t *testing.T) {
	wantText(t, dec(t, "92233720368547758.07").Add(dec(t, "0.01")), "92233720368547758.08")
	wantValue(t, dec(t, "9223372036854775807").Mul(dec(t, "2")), "18446744073709551614")
	total := dec(t, "0.00")
	for i := 0; i < 10; i++ {
		total = total.Add(dec(t, "10000000000000000.00"))
	}
	wantText(t, total, "100000000000000000.00")
	wantText(t, dec(t, "999999999999999999.99").Add(dec(t, "0.01")), "1000000000000000000.00")
	if got := dec(t, "999999999999999999.99").Digits(); got != 20 {
		t.Fatalf("digits = %d, want 20", got)
	}
}

func TestEM100And101_NegativeArithmetic(t *testing.T) {
	wantText(t, dec(t, "-10.00").Add(dec(t, "4.00")), "-6.00")
}

func TestStringKeepsLeadingZeroesOfTheFraction(t *testing.T) {
	wantText(t, dec(t, "0.0001"), "0.0001")
	wantText(t, dec(t, "-0.05"), "-0.05")
	if strings.Contains(dec(t, "-0.00").String(), "-") {
		t.Fatal("zero printed with a sign")
	}
}
