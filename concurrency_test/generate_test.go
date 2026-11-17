package concurrency

import (
	"fmt"
	"math/rand/v2"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var sharedFormulaIDs = []int{1, 3, 5, 6, 7, 8, 15}

func newStream(formulaID int, caller int) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(formulaID)*7919+20261117, uint64(caller)*104729+1))
}

func decimalText(r *rand.Rand, most int, places int) string {
	whole := r.IntN(most + 1)
	if places == 0 {
		return fmt.Sprint(whole)
	}
	fraction := r.IntN(pow10(places))
	return fmt.Sprintf("%d.%0*d", whole, places, fraction)
}

func pow10(places int) int {
	result := 1
	for i := 0; i < places; i++ {
		result *= 10
	}
	return result
}

func dayText(r *rand.Rand) string {
	return fmt.Sprintf("%04d-%02d-%02d", 2024+r.IntN(4), 1+r.IntN(12), 1+r.IntN(28))
}

func decimalTexts(r *rand.Rand, least int, most int, ceiling int, places int) []string {
	texts := make([]string, least+r.IntN(most-least+1))
	for i := range texts {
		texts[i] = decimalText(r, ceiling, places)
	}
	return texts
}

func generatedValues(formulaID int, r *rand.Rand) map[string]formula.Value {
	switch formulaID {
	case 1:
		return map[string]formula.Value{"balance": dec(decimalText(r, 9000, 2)), "payment": dec(decimalText(r, 9000, 2))}
	case 3:
		return map[string]formula.Value{
			"deposits": decList(decimalTexts(r, 1, 6, 5000, 2)...),
			"expenses": decList(decimalTexts(r, 0, 4, 900, 2)...),
			"members":  whole(int64(1 + r.IntN(500))),
		}
	case 5:
		return map[string]formula.Value{"due": day(dayText(r)), "filed": day(dayText(r))}
	case 6:
		return map[string]formula.Value{"amount": dec(decimalText(r, 20000, 2)), "due": day(dayText(r)), "filed": day(dayText(r))}
	case 7, 8:
		return map[string]formula.Value{"balance": dec(decimalText(r, 9000, 2))}
	case 15:
		units := make([]int64, 2+r.IntN(5))
		for i := range units {
			units[i] = int64(1 + r.IntN(2000))
		}
		return map[string]formula.Value{
			"call_amount": dec(decimalText(r, 90000, 2)),
			"units":       whole(int64(1 + r.IntN(2000))),
			"all_units":   intList(units...),
		}
	}
	panic(fmt.Sprintf("no generator for formula %d", formulaID))
}

func generatedBatch(formulaID int, caller int, count int) []map[string]formula.Value {
	r := newStream(formulaID, caller)
	inputs := make([]map[string]formula.Value, count)
	for i := range inputs {
		inputs[i] = generatedValues(formulaID, r)
	}
	return inputs
}
