package number

import (
	"errors"
	"fmt"
	"math/big"
)

type SplitResult struct {
	Shares   []Decimal
	LeftOver Decimal
}

func EqualSplit(total Decimal, members int, places int) (SplitResult, error) {
	if members < 1 {
		return SplitResult{}, errors.New("a split needs at least one member")
	}
	weights := make([]int64, members)
	for i := range weights {
		weights[i] = 1
	}
	return Split(total, weights, places)
}

func Split(total Decimal, weights []int64, places int) (SplitResult, error) {
	if len(weights) == 0 {
		return SplitResult{}, errors.New("a split needs at least one weight")
	}
	if places < 0 {
		return SplitResult{}, fmt.Errorf("a split keeps 0 or more places, not %d", places)
	}
	sum := new(big.Int)
	for _, weight := range weights {
		if weight < 0 {
			return SplitResult{}, errors.New("a weight cannot be negative")
		}
		sum.Add(sum, big.NewInt(weight))
	}
	if sum.Sign() == 0 {
		return SplitResult{}, errors.New("the weights of a split cannot all be zero")
	}
	denominator := new(big.Int).Mul(sum, pow10(total.scale))
	shares := make([]Decimal, len(weights))
	paid := Decimal{unscaled: new(big.Int), scale: places}
	for i, weight := range weights {
		numerator := new(big.Int).Mul(total.value(), big.NewInt(weight))
		numerator.Mul(numerator, pow10(places))
		shares[i] = Decimal{unscaled: divideRounded(numerator, denominator, RoundDown), scale: places}
		paid = paid.Add(shares[i])
	}
	return SplitResult{Shares: shares, LeftOver: total.Sub(paid)}, nil
}
