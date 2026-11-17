package builtins

import (
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

var ErrEmptyList = errors.New("list is empty")

func SumDecimals(values []numeric.Decimal) numeric.Decimal {
	total := numeric.FromInt(0)
	for _, v := range values {
		total = total.Add(v)
	}
	return total
}

func SumInts(values []int64) int64 {
	var total int64
	for _, v := range values {
		total += v
	}
	return total
}

func AvgDecimals(values []numeric.Decimal) (numeric.Decimal, error) {
	if len(values) == 0 {
		return numeric.Decimal{}, ErrEmptyList
	}
	return SumDecimals(values).Div(numeric.FromInt(int64(len(values))))
}

func AvgInts(values []int64) (float64, error) {
	if len(values) == 0 {
		return 0, ErrEmptyList
	}
	return float64(SumInts(values)) / float64(len(values)), nil
}

func MinDecimals(values []numeric.Decimal) (numeric.Decimal, error) {
	return pick(values, func(candidate, best numeric.Decimal) bool { return candidate.Cmp(best) < 0 })
}

func MaxDecimals(values []numeric.Decimal) (numeric.Decimal, error) {
	return pick(values, func(candidate, best numeric.Decimal) bool { return candidate.Cmp(best) > 0 })
}

func MinInts(values []int64) (int64, error) {
	return pick(values, func(candidate, best int64) bool { return candidate < best })
}

func MaxInts(values []int64) (int64, error) {
	return pick(values, func(candidate, best int64) bool { return candidate > best })
}

func CountDecimals(values []numeric.Decimal) int64 {
	return int64(len(values))
}

func CountInts(values []int64) int64 {
	return int64(len(values))
}

func DaysBetween(from, to time.Time) int64 {
	return int64(dateOf(to).Sub(dateOf(from)) / (24 * time.Hour))
}

func MonthsBetween(from, to time.Time) int64 {
	from, to = dateOf(from), dateOf(to)
	if to.Before(from) {
		return -MonthsBetween(to, from)
	}
	months := int64(to.Year()-from.Year())*12 + int64(to.Month()-from.Month())
	if to.Day() < from.Day() {
		months--
	}
	return months
}

func YearsBetween(from, to time.Time) int64 {
	return MonthsBetween(from, to) / 12
}

func ToFloat(d numeric.Decimal) float64 {
	return d.ToFloat()
}

func ToDecimal(f float64, places int64) (numeric.Decimal, error) {
	return numeric.FromFloat(f, int(places))
}

func RoundDecimal(d numeric.Decimal, places int64) numeric.Decimal {
	return d.Round(int(places))
}

func RoundFloat(f float64, places int64) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f
	}
	rounded, err := numeric.FromFloat(f, int(places))
	if err != nil {
		return f
	}
	back, err := strconv.ParseFloat(rounded.String(), 64)
	if err != nil {
		return f
	}
	return back
}

func pick[T any](values []T, better func(candidate, best T) bool) (T, error) {
	var best T
	if len(values) == 0 {
		return best, ErrEmptyList
	}
	best = values[0]
	for _, v := range values[1:] {
		if better(v, best) {
			best = v
		}
	}
	return best, nil
}

func dateOf(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
