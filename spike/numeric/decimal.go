package numeric

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
)

const (
	maxFastScale  = 18
	divisionScale = 18
)

var (
	ErrDivideByZero = errors.New("division by zero")
	ErrNotFinite    = errors.New("float is not finite")
)

var powersOfTen = buildPowersOfTen()

func buildPowersOfTen() [maxFastScale + 1]int64 {
	var table [maxFastScale + 1]int64
	table[0] = 1
	for i := 1; i <= maxFastScale; i++ {
		table[i] = table[i-1] * 10
	}
	return table
}

type Decimal struct {
	small int64
	scale int
	big   *big.Int
}

func FromInt(v int64) Decimal {
	return Decimal{small: v}
}

func Parse(text string) (Decimal, error) {
	body, negative := text, false
	if text != "" && (text[0] == '-' || text[0] == '+') {
		body, negative = text[1:], text[0] == '-'
	}
	whole, fraction, hasPoint := strings.Cut(body, ".")
	if whole == "" || (hasPoint && fraction == "") || !onlyDigits(whole) || !onlyDigits(fraction) {
		return Decimal{}, fmt.Errorf("not a decimal: %q", text)
	}
	return fromDigits(negative, whole+fraction, len(fraction)), nil
}

func MustParse(text string) Decimal {
	d, err := Parse(text)
	if err != nil {
		panic(err)
	}
	return d
}

func FromFloat(f float64, places int) (Decimal, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return Decimal{}, ErrNotFinite
	}
	d, err := Parse(strconv.FormatFloat(f, 'f', -1, 64))
	if err != nil {
		return Decimal{}, err
	}
	return d.Round(places), nil
}

func (d Decimal) String() string {
	digits := d.unscaledString()
	negative := strings.HasPrefix(digits, "-")
	digits = strings.TrimPrefix(digits, "-")
	if d.scale > 0 {
		if len(digits) <= d.scale {
			digits = strings.Repeat("0", d.scale-len(digits)+1) + digits
		}
		cut := len(digits) - d.scale
		digits = digits[:cut] + "." + digits[cut:]
	}
	if negative {
		return "-" + digits
	}
	return digits
}

func (d Decimal) DecimalPlaces() int {
	return d.scale
}

func (d Decimal) IsBig() bool {
	return d.big != nil
}

func (d Decimal) Sign() int {
	if d.big != nil {
		return d.big.Sign()
	}
	switch {
	case d.small > 0:
		return 1
	case d.small < 0:
		return -1
	}
	return 0
}

func (d Decimal) Add(o Decimal) Decimal {
	if sum, ok := addFast(d, o); ok {
		return sum
	}
	scale := max(d.scale, o.scale)
	left := scaleUp(d.toBig(), scale-d.scale)
	right := scaleUp(o.toBig(), scale-o.scale)
	return fromBig(new(big.Int).Add(left, right), scale)
}

func (d Decimal) Sub(o Decimal) Decimal {
	return d.Add(o.Neg())
}

func (d Decimal) Mul(o Decimal) Decimal {
	scale := d.scale + o.scale
	if d.big == nil && o.big == nil && scale <= maxFastScale {
		if product, ok := mulInt64(d.small, o.small); ok {
			return Decimal{small: product, scale: scale}
		}
	}
	return fromBig(new(big.Int).Mul(d.toBig(), o.toBig()), scale)
}

func (d Decimal) Div(o Decimal) (Decimal, error) {
	if o.Sign() == 0 {
		return Decimal{}, ErrDivideByZero
	}
	numerator, denominator := d.toBig(), o.toBig()
	shift := divisionScale - d.scale + o.scale
	if shift >= 0 {
		numerator = scaleUp(numerator, shift)
	} else {
		denominator = scaleUp(denominator, -shift)
	}
	return fromBig(divideRounded(numerator, denominator), divisionScale), nil
}

func (d Decimal) Cmp(o Decimal) int {
	if order, ok := cmpFast(d, o); ok {
		return order
	}
	scale := max(d.scale, o.scale)
	return scaleUp(d.toBig(), scale-d.scale).Cmp(scaleUp(o.toBig(), scale-o.scale))
}

func (d Decimal) Neg() Decimal {
	if d.big == nil && d.small != math.MinInt64 {
		return Decimal{small: -d.small, scale: d.scale}
	}
	return fromBig(new(big.Int).Neg(d.toBig()), d.scale)
}

func (d Decimal) Round(places int) Decimal {
	places = max(places, 0)
	switch {
	case places == d.scale:
		return d
	case places > d.scale:
		return d.widen(places)
	}
	return d.narrow(places)
}

func (d Decimal) ToFloat() float64 {
	f, _ := strconv.ParseFloat(d.String(), 64)
	return f
}

func (d Decimal) widen(places int) Decimal {
	by := places - d.scale
	if d.big == nil && places <= maxFastScale {
		if v, ok := mulInt64(d.small, powersOfTen[by]); ok {
			return Decimal{small: v, scale: places}
		}
	}
	return fromBig(scaleUp(d.toBig(), by), places)
}

func (d Decimal) narrow(places int) Decimal {
	by := d.scale - places
	if d.big != nil {
		return fromBig(divideRounded(d.big, pow10Big(by)), places)
	}
	divisor := uint64(powersOfTen[by])
	magnitude := absUint64(d.small)
	quotient, remainder := magnitude/divisor, magnitude%divisor
	if remainder >= divisor-remainder {
		quotient++
	}
	if d.small < 0 {
		return Decimal{small: -int64(quotient), scale: places}
	}
	return Decimal{small: int64(quotient), scale: places}
}

func (d Decimal) toBig() *big.Int {
	if d.big != nil {
		return d.big
	}
	return big.NewInt(d.small)
}

func (d Decimal) unscaledString() string {
	if d.big != nil {
		return d.big.String()
	}
	return strconv.FormatInt(d.small, 10)
}

func fromBig(v *big.Int, scale int) Decimal {
	if scale <= maxFastScale && v.IsInt64() {
		return Decimal{small: v.Int64(), scale: scale}
	}
	return Decimal{scale: scale, big: v}
}

func fromDigits(negative bool, digits string, scale int) Decimal {
	if len(digits) < 19 && scale <= maxFastScale {
		v, _ := strconv.ParseInt(digits, 10, 64)
		if negative {
			v = -v
		}
		return Decimal{small: v, scale: scale}
	}
	v, _ := new(big.Int).SetString(digits, 10)
	if negative {
		v.Neg(v)
	}
	return fromBig(v, scale)
}

func onlyDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func pow10Big(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func scaleUp(v *big.Int, by int) *big.Int {
	if by == 0 {
		return v
	}
	return new(big.Int).Mul(v, pow10Big(by))
}

func divideRounded(numerator, denominator *big.Int) *big.Int {
	negative := numerator.Sign()*denominator.Sign() < 0
	magnitude := new(big.Int).Abs(denominator)
	quotient, remainder := new(big.Int).QuoRem(new(big.Int).Abs(numerator), magnitude, new(big.Int))
	if remainder.Lsh(remainder, 1).Cmp(magnitude) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if negative {
		quotient.Neg(quotient)
	}
	return quotient
}

func alignedFast(a, b Decimal) (int64, int64, int, bool) {
	if a.big != nil || b.big != nil {
		return 0, 0, 0, false
	}
	scale := max(a.scale, b.scale)
	left, okLeft := mulInt64(a.small, powersOfTen[scale-a.scale])
	right, okRight := mulInt64(b.small, powersOfTen[scale-b.scale])
	return left, right, scale, okLeft && okRight
}

func addFast(a, b Decimal) (Decimal, bool) {
	left, right, scale, ok := alignedFast(a, b)
	if !ok {
		return Decimal{}, false
	}
	sum, ok := addInt64(left, right)
	return Decimal{small: sum, scale: scale}, ok
}

func cmpFast(a, b Decimal) (int, bool) {
	left, right, _, ok := alignedFast(a, b)
	switch {
	case !ok:
		return 0, false
	case left < right:
		return -1, true
	case left > right:
		return 1, true
	}
	return 0, true
}

func absUint64(v int64) uint64 {
	if v < 0 {
		return uint64(-v)
	}
	return uint64(v)
}

func mulInt64(a, b int64) (int64, bool) {
	high, low := bits.Mul64(absUint64(a), absUint64(b))
	if high != 0 {
		return 0, false
	}
	if (a < 0) != (b < 0) {
		if low > 1<<63 {
			return 0, false
		}
		return -int64(low), true
	}
	if low > math.MaxInt64 {
		return 0, false
	}
	return int64(low), true
}

func addInt64(a, b int64) (int64, bool) {
	sum := a + b
	if (a >= 0) == (b >= 0) && (sum >= 0) != (a >= 0) {
		return 0, false
	}
	return sum, true
}
