package own

import (
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/builtins"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func dateFunctions() map[string][]overload {
	between := func(f func(from, to time.Time) int64) []overload {
		return []overload{overload2(harness.KindDate, harness.KindDate, harness.KindInteger, certain2(f))}
	}
	return map[string][]overload{
		"days_between":   between(builtins.DaysBetween),
		"months_between": between(builtins.MonthsBetween),
		"years_between":  between(builtins.YearsBetween),
	}
}
