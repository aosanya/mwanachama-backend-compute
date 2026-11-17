package datefn

func DaysBetween(a, b Date) int {
	return b.dayNumber() - a.dayNumber()
}

func MonthsBetween(a, b Date) int {
	if b.before(a) {
		return -wholeMonths(b, a)
	}
	return wholeMonths(a, b)
}

func YearsBetween(a, b Date) int {
	return MonthsBetween(a, b) / 12
}

func wholeMonths(from, to Date) int {
	months := (to.year-from.year)*12 + to.month - from.month
	if to.day < from.day && !to.isLastDayOfMonth() {
		months--
	}
	return months
}
