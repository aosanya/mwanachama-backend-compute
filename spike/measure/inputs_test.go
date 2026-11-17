package measure

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type inputMaker func(r *rand.Rand) map[string]any

func money(r *rand.Rand, maxUnits int64) numeric.Decimal {
	return numeric.MustParse(fmt.Sprintf("%d.%02d", r.Int63n(maxUnits), r.Intn(100)))
}

func moneyList(r *rand.Rand, minLen, maxLen int, maxUnits int64) []numeric.Decimal {
	values := make([]numeric.Decimal, minLen+r.Intn(maxLen-minLen+1))
	for i := range values {
		values[i] = money(r, maxUnits)
	}
	return values
}

func someDay(r *rand.Rand, spread int) time.Time {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return base.AddDate(0, 0, r.Intn(spread))
}

func formulaThreeInputs(r *rand.Rand) map[string]any {
	return map[string]any{
		"deposits": moneyList(r, 1, 6, 5000),
		"expenses": moneyList(r, 0, 3, 100),
		"members":  int64(1 + r.Intn(60)),
	}
}

func formulaFiveInputs(r *rand.Rand) map[string]any {
	due := someDay(r, 200)
	return map[string]any{"due": due, "filed": due.AddDate(0, 0, r.Intn(160)-20)}
}

func formulaSixInputs(r *rand.Rand) map[string]any {
	inputs := formulaFiveInputs(r)
	inputs["amount"] = money(r, 100000)
	return inputs
}

func formulaSevenInputs(r *rand.Rand) map[string]any {
	return map[string]any{"balance": money(r, 9000)}
}

func formulaFifteenInputs(r *rand.Rand) map[string]any {
	held := make([]int64, 2+r.Intn(8))
	var total int64
	for i := range held {
		held[i] = int64(1 + r.Intn(2000))
		total += held[i]
	}
	return map[string]any{
		"call_amount": money(r, 200000),
		"units":       int64(r.Intn(int(total) + 1)),
		"all_units":   held,
	}
}

var makers = map[int]inputMaker{
	3:  formulaThreeInputs,
	5:  formulaFiveInputs,
	6:  formulaSixInputs,
	7:  formulaSevenInputs,
	8:  formulaSevenInputs,
	15: formulaFifteenInputs,
}

func makeInputs(number int, seed int64, count int) []map[string]any {
	r := rand.New(rand.NewSource(seed))
	rows := make([]map[string]any, count)
	for i := range rows {
		rows[i] = makers[number](r)
	}
	return rows
}
