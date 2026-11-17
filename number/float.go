package number

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

type Float struct {
	v float64
}

type FloatReading int

const (
	ReadShortest FloatReading = iota
	ReadExact
)

func NewFloat(v float64) (Float, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return Float{}, ErrNotFinite
	}
	return Float{v: v}, nil
}

func (f Float) Value() float64 { return f.v }

func (f Float) Add(other Float) (Float, error) { return NewFloat(f.v + other.v) }

func (f Float) Sub(other Float) (Float, error) { return NewFloat(f.v - other.v) }

func (f Float) Mul(other Float) (Float, error) { return NewFloat(f.v * other.v) }

func (f Float) Div(other Float) (Float, error) {
	if other.v == 0 {
		return Float{}, ErrDivideByZero
	}
	return NewFloat(f.v / other.v)
}

func (f Float) ToDecimal(places int, mode Rounding, reading FloatReading) (Decimal, error) {
	if places < 0 {
		return Decimal{}, fmt.Errorf("a conversion states 0 or more places, not %d", places)
	}
	var read Decimal
	if reading == ReadExact {
		read = exactDecimal(f.v)
	} else {
		parsed, err := ParseDecimal(strconv.FormatFloat(f.v, 'f', -1, 64))
		if err != nil {
			return Decimal{}, err
		}
		read = parsed
	}
	return read.Round(places, mode), nil
}

func exactDecimal(v float64) Decimal {
	rat := new(big.Rat).SetFloat64(v)
	scale := rat.Denom().BitLen() - 1
	unscaled := new(big.Int).Mul(rat.Num(), new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(scale)), nil))
	return Decimal{unscaled: unscaled, scale: scale}
}
