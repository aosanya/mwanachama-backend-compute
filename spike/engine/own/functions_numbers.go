package own

import (
	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func numberFunctions() map[string][]overload {
	return map[string][]overload{
		"round": {
			overload2(harness.KindDecimal, harness.KindInteger, harness.KindDecimal, certain2(builtins.RoundDecimal)),
			overload2(harness.KindFloat, harness.KindInteger, harness.KindFloat, certain2(builtins.RoundFloat)),
		},
		"to_float": {
			overload1(harness.KindDecimal, harness.KindFloat, certain1(builtins.ToFloat)),
			overload1(harness.KindInteger, harness.KindFloat, certain1(func(v int64) float64 { return float64(v) })),
		},
		"to_decimal": {
			overload2[float64, int64, numeric.Decimal](harness.KindFloat, harness.KindInteger, harness.KindDecimal, builtins.ToDecimal),
		},
	}
}
