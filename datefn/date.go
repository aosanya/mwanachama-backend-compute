package datefn

import (
	"errors"
	"fmt"
)

var ErrNotADate = errors.New("not a date written like 2026-03-31")

const (
	minYear = 1
	maxYear = 9999
)

type Date struct {
	year  int
	month int
	day   int
}

func Parse(text string) (Date, error) {
	if len(text) != 10 || text[4] != '-' || text[7] != '-' {
		return Date{}, ErrNotADate
	}
	year, ok := digits(text[0:4])
	if !ok {
		return Date{}, ErrNotADate
	}
	month, ok := digits(text[5:7])
	if !ok {
		return Date{}, ErrNotADate
	}
	day, ok := digits(text[8:10])
	if !ok {
		return Date{}, ErrNotADate
	}
	if year < minYear || year > maxYear || month < 1 || month > 12 || day < 1 || day > daysInMonth(year, month) {
		return Date{}, ErrNotADate
	}
	return Date{year: year, month: month, day: day}, nil
}

func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
}

func digits(text string) (int, bool) {
	value := 0
	for i := 0; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return 0, false
		}
		value = value*10 + int(text[i]-'0')
	}
	return value, true
}

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func daysInMonth(year, month int) int {
	switch month {
	case 2:
		if isLeap(year) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	}
	return 31
}

func (d Date) isLastDayOfMonth() bool {
	return d.day == daysInMonth(d.year, d.month)
}

func (d Date) before(other Date) bool {
	if d.year != other.year {
		return d.year < other.year
	}
	if d.month != other.month {
		return d.month < other.month
	}
	return d.day < other.day
}

func (d Date) dayNumber() int {
	year := d.year
	month := d.month
	if month <= 2 {
		year--
	}
	era := year / 400
	yearOfEra := year - era*400
	shifted := month + 9
	if month > 2 {
		shifted = month - 3
	}
	dayOfYear := (153*shifted+2)/5 + d.day - 1
	dayOfEra := yearOfEra*365 + yearOfEra/4 - yearOfEra/100 + dayOfYear
	return era*146097 + dayOfEra
}
