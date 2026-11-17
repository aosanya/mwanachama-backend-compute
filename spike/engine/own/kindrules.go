package own

import (
	"fmt"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func commonKind(a, b harness.Kind) (harness.Kind, error) {
	switch {
	case a == b:
		return a, nil
	case isPair(a, b, harness.KindInteger, harness.KindDecimal):
		return harness.KindDecimal, nil
	case isPair(a, b, harness.KindInteger, harness.KindFloat):
		return harness.KindFloat, nil
	case isPair(a, b, harness.KindDecimal, harness.KindFloat):
		return 0, fmt.Errorf("%s and %s cannot be combined, convert one with to_float or to_decimal", withArticle(a), withArticle(b))
	}
	return 0, fmt.Errorf("%s and %s cannot be combined", withArticle(a), withArticle(b))
}

func isPair(a, b, first, second harness.Kind) bool {
	return (a == first && b == second) || (a == second && b == first)
}

func promote(c code, to harness.Kind) code {
	switch {
	case c.kind == to:
		return c
	case c.kind == harness.KindInteger && to == harness.KindDecimal:
		return code{to, apply1(as[int64](c), certain1(numeric.FromInt))}
	case c.kind == harness.KindInteger && to == harness.KindFloat:
		return code{to, apply1(as[int64](c), certain1(func(v int64) float64 { return float64(v) }))}
	}
	return c
}

func align(pos Pos, left, right code) (code, code, error) {
	kind, err := commonKind(left.kind, right.kind)
	if err != nil {
		return code{}, code{}, typeError(pos, err.Error())
	}
	return promote(left, kind), promote(right, kind), nil
}
