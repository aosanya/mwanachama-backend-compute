package number

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var (
	ErrDivideByZero = errors.New("division by zero")
	ErrNotFinite    = errors.New("the number is not finite")
)

var (
	bigZero = new(big.Int)
	bigTen  = big.NewInt(10)
)

type Decimal struct {
	unscaled *big.Int
	scale    int
}

func NewDecimal(unscaled int64, scale int) Decimal {
	return Decimal{unscaled: big.NewInt(unscaled), scale: scale}
}

func DecimalFromInt(value int64) Decimal {
	return NewDecimal(value, 0)
}

func ParseDecimal(text string) (Decimal, error) {
	body := strings.TrimPrefix(text, "-")
	negative := len(body) != len(text)
	whole, fraction, hasPoint := strings.Cut(body, ".")
	if !isDigits(whole) || (len(whole) > 1 && whole[0] == '0') || (hasPoint && !isDigits(fraction)) {
		return Decimal{}, fmt.Errorf("%q is not a decimal", text)
	}
	unscaled, _ := new(big.Int).SetString(whole+fraction, 10)
	if negative {
		unscaled.Neg(unscaled)
	}
	return Decimal{unscaled: unscaled, scale: len(fraction)}, nil
}

func isDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (d Decimal) value() *big.Int {
	if d.unscaled == nil {
		return bigZero
	}
	return d.unscaled
}

func pow10(exponent int) *big.Int {
	return new(big.Int).Exp(bigTen, big.NewInt(int64(exponent)), nil)
}

func (d Decimal) Scale() int { return d.scale }

func (d Decimal) Sign() int { return d.value().Sign() }

func (d Decimal) IsZero() bool { return d.Sign() == 0 }

func (d Decimal) Digits() int {
	return len(new(big.Int).Abs(d.value()).String())
}

func (d Decimal) String() string {
	digits := new(big.Int).Abs(d.value()).String()
	sign := ""
	if d.Sign() < 0 {
		sign = "-"
	}
	if d.scale <= 0 {
		return sign + digits
	}
	if len(digits) <= d.scale {
		digits = strings.Repeat("0", d.scale-len(digits)+1) + digits
	}
	cut := len(digits) - d.scale
	return sign + digits[:cut] + "." + digits[cut:]
}

func aligned(a, b Decimal) (*big.Int, *big.Int, int) {
	scale := max(a.scale, b.scale)
	left := new(big.Int).Mul(a.value(), pow10(scale-a.scale))
	right := new(big.Int).Mul(b.value(), pow10(scale-b.scale))
	return left, right, scale
}

func (d Decimal) Add(other Decimal) Decimal {
	left, right, scale := aligned(d, other)
	return Decimal{unscaled: left.Add(left, right), scale: scale}
}

func (d Decimal) Sub(other Decimal) Decimal {
	left, right, scale := aligned(d, other)
	return Decimal{unscaled: left.Sub(left, right), scale: scale}
}

func (d Decimal) Mul(other Decimal) Decimal {
	return Decimal{unscaled: new(big.Int).Mul(d.value(), other.value()), scale: d.scale + other.scale}
}

func (d Decimal) Neg() Decimal {
	return Decimal{unscaled: new(big.Int).Neg(d.value()), scale: d.scale}
}

func (d Decimal) Cmp(other Decimal) int {
	left, right, _ := aligned(d, other)
	return left.Cmp(right)
}

func (d Decimal) Equal(other Decimal) bool {
	return d.Cmp(other) == 0
}

func (d Decimal) Round(places int, mode Rounding) Decimal {
	places = max(places, 0)
	shift := d.scale - places
	if shift <= 0 {
		return Decimal{unscaled: new(big.Int).Mul(d.value(), pow10(-shift)), scale: places}
	}
	return Decimal{unscaled: divideRounded(d.value(), pow10(shift), mode), scale: places}
}

func (d Decimal) Div(divisor Decimal, digits int, mode Rounding) (Decimal, error) {
	if divisor.IsZero() {
		return Decimal{}, ErrDivideByZero
	}
	if digits < 0 {
		return Decimal{}, fmt.Errorf("a division keeps 0 or more digits, not %d", digits)
	}
	numerator := new(big.Int).Set(d.value())
	denominator := new(big.Int).Set(divisor.value())
	exponent := digits + divisor.scale - d.scale
	if exponent >= 0 {
		numerator.Mul(numerator, pow10(exponent))
	} else {
		denominator.Mul(denominator, pow10(-exponent))
	}
	return Decimal{unscaled: divideRounded(numerator, denominator, mode), scale: digits}, nil
}

func (d Decimal) ToFloat() (Float, error) {
	parsed, err := strconv.ParseFloat(d.String(), 64)
	if err != nil {
		return Float{}, ErrNotFinite
	}
	return NewFloat(parsed)
}
