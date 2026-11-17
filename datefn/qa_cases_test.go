package datefn

import "testing"

type qaCase struct {
	name     string
	function string
	a, b     string
	want     int
}

func TestCasesOfTheQAPages(t *testing.T) {
	cases := []qaCase{
		{"lang-f05-late-days", "days", "2026-03-31", "2026-05-15", 45},
		{"rules-type-date-valid", "days", "2026-03-31", "2026-05-15", 45},
		{"lang-f11-prorated-first-month-fee", "days", "2026-06-12", "2026-06-30", 18},
		{"lang-f12-straight-line-depreciation", "years", "2023-09-10", "2026-11-17", 3},
		{"lang-f18-voting-right", "months", "2026-01-20", "2026-09-15", 7},
		{"lang-save-extra-brackets", "months", "2026-01-20", "2026-09-15", 7},
		{"REF-date.late_days-01", "days", "2026-03-31", "2026-05-15", 45},
		{"REF-date.late_days-02", "days", "2026-03-31", "2026-03-20", -11},
		{"REF-date.late_days-03", "days", "2026-03-31", "2026-03-31", 0},
		{"REF-date.late_days-04", "days", "2026-03-31", "2026-04-01", 1},
		{"REF-date.late_days-05", "days", "2028-02-28", "2028-03-01", 2},
		{"REF-date.late_days-06", "days", "2026-12-31", "2027-03-01", 60},
	}
	for _, c := range cases {
		a, b := mustParse(t, c.a), mustParse(t, c.b)
		var got int
		switch c.function {
		case "days":
			got = DaysBetween(a, b)
		case "months":
			got = MonthsBetween(a, b)
		case "years":
			got = YearsBetween(a, b)
		}
		if got != c.want {
			t.Errorf("%s: %s between %s and %s = %d, want %d", c.name, c.function, c.a, c.b, got, c.want)
		}
	}
}

func TestRefusalsOfTheQAPages(t *testing.T) {
	for _, c := range []struct{ name, text string }{
		{"rules-type-date-invalid", "2026-02-30"},
		{"REF-date.late_days-08", "2026-02-30"},
	} {
		if _, err := Parse(c.text); err == nil {
			t.Errorf("%s: %q was read as a date", c.name, c.text)
		}
	}
}
