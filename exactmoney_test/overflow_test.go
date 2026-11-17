package exactmoney_test

import (
	"math"
	"math/big"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func bigText(value *big.Int, places int) string {
	sign := ""
	digits := new(big.Int).Abs(value).String()
	if value.Sign() < 0 {
		sign = "-"
	}
	for len(digits) <= places {
		digits = "0" + digits
	}
	if places == 0 {
		return sign + digits
	}
	return sign + digits[:len(digits)-places] + "." + digits[len(digits)-places:]
}

func TestEM81_TheSixtyFourBitLimitPlusOneCentDoesNotWrapOrSaturate(t *testing.T) {
	limit := units(math.MaxInt64, 2)
	got := limit.Add(units(1, 2))
	want := new(big.Int).Add(big.NewInt(math.MaxInt64), big.NewInt(1))
	wantText(t, got, bigText(want, 2))
	wantText(t, got, "92233720368547758.08")
}

func TestEM82_TheSixtyFourBitLimitTimesTwo(t *testing.T) {
	got := units(math.MaxInt64, 2).Mul(number.DecimalFromInt(2))
	want := new(big.Int).Mul(big.NewInt(math.MaxInt64), big.NewInt(2))
	wantText(t, got, bigText(want, 2))
	wantText(t, got, "184467440737095516.14")
}

func TestEM83_TenAmountsOfTenToTheEighteenCents(t *testing.T) {
	each := units(1_000_000_000_000_000_000, 2)
	total := number.NewDecimal(0, 2)
	for i := 0; i < 9; i++ {
		total = total.Add(each)
	}
	wantText(t, total, "90000000000000000.00")
	total = total.Add(each)
	wantText(t, total, "100000000000000000.00")
}

func TestEM96_ACarryThroughEveryDigit(t *testing.T) {
	got := dec(t, "999999999999999999.99").Add(dec(t, "0.01"))
	wantText(t, got, "1000000000000000000.00")
}

func TestOverflow_ReverseAndForwardSumsAgreeBeyondSixtyFourBits(t *testing.T) {
	amounts := make([]number.Decimal, 0, 40)
	independent := new(big.Int)
	for i := 0; i < 40; i++ {
		units64 := int64(math.MaxInt64) - int64(i)
		amounts = append(amounts, units(units64, 2))
		independent.Add(independent, big.NewInt(units64))
	}
	forward := number.NewDecimal(0, 2)
	backward := number.NewDecimal(0, 2)
	for i := range amounts {
		forward = forward.Add(amounts[i])
		backward = backward.Add(amounts[len(amounts)-1-i])
	}
	wantText(t, forward, bigText(independent, 2))
	wantText(t, backward, bigText(independent, 2))
}
