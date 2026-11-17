package own

import (
	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func listFunctions() map[string][]overload {
	decimals, integers := harness.KindDecimalList, harness.KindIntegerList
	return map[string][]overload{
		"sum": {
			overload1(decimals, harness.KindDecimal, certain1(builtins.SumDecimals)),
			overload1(integers, harness.KindInteger, certain1(builtins.SumInts)),
		},
		"avg": {
			overload1[[]numeric.Decimal, numeric.Decimal](decimals, harness.KindDecimal, builtins.AvgDecimals),
			overload1[[]int64, float64](integers, harness.KindFloat, builtins.AvgInts),
		},
		"min": {
			overload1[[]numeric.Decimal, numeric.Decimal](decimals, harness.KindDecimal, builtins.MinDecimals),
			overload1[[]int64, int64](integers, harness.KindInteger, builtins.MinInts),
		},
		"max": {
			overload1[[]numeric.Decimal, numeric.Decimal](decimals, harness.KindDecimal, builtins.MaxDecimals),
			overload1[[]int64, int64](integers, harness.KindInteger, builtins.MaxInts),
		},
		"count": {
			overload1(decimals, harness.KindInteger, certain1(builtins.CountDecimals)),
			overload1(integers, harness.KindInteger, certain1(builtins.CountInts)),
		},
	}
}
