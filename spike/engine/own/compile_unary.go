package own

import (
	"math"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func (c *compiler) unary(n Unary) (code, error) {
	operand, err := c.expr(n.Operand)
	if err != nil {
		return code{}, err
	}
	if n.Op == "!" {
		return not(n, operand)
	}
	return minus(n, operand)
}

func not(n Unary, operand code) (code, error) {
	if operand.kind != harness.KindBoolean {
		return code{}, typeError(n.Pos, "! needs a boolean but was given "+withArticle(operand.kind))
	}
	invert := certain1(func(v bool) bool { return !v })
	return code{harness.KindBoolean, apply1(as[bool](operand), invert)}, nil
}

func minus(n Unary, operand code) (code, error) {
	switch operand.kind {
	case harness.KindInteger:
		return code{operand.kind, apply1(as[int64](operand), negateInt)}, nil
	case harness.KindDecimal:
		return code{operand.kind, apply1(as[numeric.Decimal](operand), certain1(numeric.Decimal.Neg))}, nil
	case harness.KindFloat:
		return code{operand.kind, apply1(as[float64](operand), certain1(func(v float64) float64 { return -v }))}, nil
	}
	return code{}, typeError(n.Pos, "- needs a number but was given "+withArticle(operand.kind))
}

func negateInt(v int64) (int64, error) {
	if v == math.MinInt64 {
		return 0, errOverflow
	}
	return -v, nil
}
