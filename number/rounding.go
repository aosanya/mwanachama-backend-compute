package number

import "math/big"

type Rounding int

const (
	RoundHalfUp Rounding = iota
	RoundHalfDown
	RoundHalfEven
	RoundHalfPosInf
	RoundUp
	RoundDown
	RoundCeiling
	RoundFloor
)

var roundingNames = map[Rounding]string{
	RoundHalfUp:     "half-up",
	RoundHalfDown:   "half-down",
	RoundHalfEven:   "half-even",
	RoundHalfPosInf: "half-+inf",
	RoundUp:         "up",
	RoundDown:       "down",
	RoundCeiling:    "ceiling",
	RoundFloor:      "floor",
}

func (r Rounding) String() string {
	if name, ok := roundingNames[r]; ok {
		return name
	}
	return "unknown"
}

func divideRounded(numerator, denominator *big.Int, mode Rounding) *big.Int {
	negative := (numerator.Sign() < 0) != (denominator.Sign() < 0)
	n := new(big.Int).Abs(numerator)
	d := new(big.Int).Abs(denominator)
	quotient, remainder := new(big.Int).QuoRem(n, d, new(big.Int))
	if remainder.Sign() != 0 && stepsAwayFromZero(mode, negative, quotient, remainder, d) {
		quotient.Add(quotient, big.NewInt(1))
	}
	if negative {
		quotient.Neg(quotient)
	}
	return quotient
}

func stepsAwayFromZero(mode Rounding, negative bool, quotient, remainder, divisor *big.Int) bool {
	againstHalf := new(big.Int).Lsh(remainder, 1).Cmp(divisor)
	switch mode {
	case RoundUp:
		return true
	case RoundDown:
		return false
	case RoundHalfDown:
		return againstHalf > 0
	case RoundHalfEven:
		return againstHalf > 0 || (againstHalf == 0 && quotient.Bit(0) == 1)
	case RoundHalfPosInf:
		return againstHalf > 0 || (againstHalf == 0 && !negative)
	case RoundCeiling:
		return !negative
	case RoundFloor:
		return negative
	default:
		return againstHalf >= 0
	}
}
