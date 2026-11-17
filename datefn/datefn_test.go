package datefn

import (
	"errors"
	"testing"
)

func mustParse(t *testing.T, text string) Date {
	t.Helper()
	d, err := Parse(text)
	if err != nil {
		t.Fatalf("Parse(%q) refused: %v", text, err)
	}
	return d
}

func TestParseReadsAnExistingDay(t *testing.T) {
	for _, text := range []string{"2026-03-31", "2028-02-29", "0001-01-01", "9999-12-31", "2026-12-31", "2000-02-29"} {
		d := mustParse(t, text)
		if d.String() != text {
			t.Errorf("Parse(%q).String() = %q", text, d.String())
		}
	}
}

func TestParseRefusesWhatIsNotADate(t *testing.T) {
	for _, text := range []string{
		"2026-02-30",
		"2027-02-29",
		"2100-02-29",
		"2026-04-31",
		"2026-13-01",
		"2026-00-10",
		"2026-03-00",
		"31/03/2026",
		"",
		" ",
		"2026-3-31",
		"2026-03-3",
		"26-03-31",
		"2026-03-31 ",
		" 2026-03-31",
		"2026-03-31T00:00:00",
		"2026/03/31",
		"abcd-ef-gh",
		"+026-03-31",
		"0000-01-01",
		"２０２６-03-31",
	} {
		_, err := Parse(text)
		if !errors.Is(err, ErrNotADate) {
			t.Errorf("Parse(%q) error = %v, want ErrNotADate", text, err)
		}
	}
}

func TestDaysBetween(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"worked line", "2026-03-31", "2026-05-15", 45},
		{"same day", "2026-03-31", "2026-03-31", 0},
		{"next day", "2026-03-31", "2026-04-01", 1},
		{"leap day crossed", "2028-02-28", "2028-03-01", 2},
		{"no leap day in 2026", "2026-02-28", "2026-03-01", 1},
		{"year boundary", "2026-12-31", "2027-03-01", 60},
		{"new year", "2026-12-31", "2027-01-01", 1},
		{"reversed", "2026-05-15", "2026-03-31", -45},
		{"reversed next day", "2026-04-01", "2026-03-31", -1},
		{"whole leap year", "2028-01-01", "2029-01-01", 366},
		{"whole common year", "2026-01-01", "2027-01-01", 365},
		{"century not leap", "2100-02-28", "2100-03-01", 1},
		{"four hundred year leap", "2000-02-28", "2000-03-01", 2},
	}
	for _, c := range cases {
		got := DaysBetween(mustParse(t, c.a), mustParse(t, c.b))
		if got != c.want {
			t.Errorf("%s: DaysBetween(%s, %s) = %d, want %d", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestMonthsBetween(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"worked line", "2026-01-20", "2026-09-15", 7},
		{"last day of month counts", "2026-01-31", "2026-02-28", 1},
		{"same day of month", "2026-01-15", "2026-02-15", 1},
		{"one day short", "2026-01-15", "2026-02-14", 0},
		{"same day", "2026-03-31", "2026-03-31", 0},
		{"thirty first to thirtieth", "2026-05-31", "2026-06-30", 1},
		{"thirty first to twenty ninth", "2026-05-31", "2026-06-29", 0},
		{"thirtieth to thirty first", "2026-04-30", "2026-05-31", 1},
		{"leap day to last day of february", "2028-02-29", "2029-02-28", 12},
		{"leap day to twenty seventh", "2028-02-29", "2029-02-27", 11},
		{"leap day to leap day", "2024-02-29", "2028-02-29", 48},
		{"year boundary just short", "2025-12-15", "2026-01-14", 0},
		{"year boundary", "2025-12-15", "2026-01-15", 1},
		{"across several years", "2023-09-10", "2026-11-17", 38},
		{"reversed", "2026-09-15", "2026-01-20", -7},
		{"reversed last day of month", "2026-02-28", "2026-01-31", -1},
		{"reversed same day of month", "2026-02-15", "2026-01-15", -1},
		{"reversed short", "2026-02-14", "2026-01-15", 0},
	}
	for _, c := range cases {
		got := MonthsBetween(mustParse(t, c.a), mustParse(t, c.b))
		if got != c.want {
			t.Errorf("%s: MonthsBetween(%s, %s) = %d, want %d", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestYearsBetween(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"worked line", "2023-09-10", "2026-11-17", 3},
		{"same day", "2026-03-31", "2026-03-31", 0},
		{"exact year", "2025-06-15", "2026-06-15", 1},
		{"one day short of a year", "2025-06-15", "2026-06-14", 0},
		{"leap day anniversary on last day of february", "2028-02-29", "2029-02-28", 1},
		{"leap day anniversary one day early", "2028-02-29", "2029-02-27", 0},
		{"ten months", "2026-01-20", "2026-12-19", 0},
		{"eleven months and more", "2026-01-20", "2026-12-31", 0},
		{"twenty three months", "2024-01-20", "2025-12-31", 1},
		{"reversed", "2026-11-17", "2023-09-10", -3},
		{"reversed drops the rest", "2026-06-14", "2025-06-15", 0},
		{"reversed exact year", "2026-06-15", "2025-06-15", -1},
	}
	for _, c := range cases {
		got := YearsBetween(mustParse(t, c.a), mustParse(t, c.b))
		if got != c.want {
			t.Errorf("%s: YearsBetween(%s, %s) = %d, want %d", c.name, c.a, c.b, got, c.want)
		}
	}
}

func TestYearsBetweenIsMonthsBetweenDividedByTwelve(t *testing.T) {
	dates := []string{"2024-02-29", "2025-01-31", "2025-06-15", "2026-02-28", "2026-03-31", "2027-12-01", "2029-02-28"}
	for _, a := range dates {
		for _, b := range dates {
			da, db := mustParse(t, a), mustParse(t, b)
			if got, want := YearsBetween(da, db), MonthsBetween(da, db)/12; got != want {
				t.Errorf("YearsBetween(%s, %s) = %d, MonthsBetween/12 = %d", a, b, got, want)
			}
		}
	}
}

func TestReversedPairIsTheNegative(t *testing.T) {
	dates := []string{"2024-02-29", "2025-01-31", "2025-06-15", "2026-02-28", "2026-03-31", "2027-12-01", "2029-02-28"}
	for _, a := range dates {
		for _, b := range dates {
			da, db := mustParse(t, a), mustParse(t, b)
			if DaysBetween(da, db) != -DaysBetween(db, da) {
				t.Errorf("DaysBetween(%s, %s) is not the negative of the reverse", a, b)
			}
			if MonthsBetween(da, db) != -MonthsBetween(db, da) {
				t.Errorf("MonthsBetween(%s, %s) is not the negative of the reverse", a, b)
			}
		}
	}
}

func TestSameInputGivesTheSameFigureEveryTime(t *testing.T) {
	a, b := mustParse(t, "2026-01-31"), mustParse(t, "2027-02-28")
	first := [3]int{DaysBetween(a, b), MonthsBetween(a, b), YearsBetween(a, b)}
	for i := 0; i < 100; i++ {
		got := [3]int{DaysBetween(a, b), MonthsBetween(a, b), YearsBetween(a, b)}
		if got != first {
			t.Fatalf("run %d gave %v, first gave %v", i, got, first)
		}
	}
}
