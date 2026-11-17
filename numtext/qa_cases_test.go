package numtext

import (
	"strings"
	"testing"
)

type qaRow struct {
	id       string
	text     string
	accepted bool
}

func runQARows(t *testing.T, kind Type, rows []qaRow) {
	t.Helper()
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			_, refusals := Read(kind, One(row.text))
			if row.accepted && len(refusals) != 0 {
				t.Fatalf("%q should be accepted, got %+v", row.text, refusals)
			}
			if !row.accepted && (len(refusals) != 1 || refusals[0].Rule != RuleType) {
				t.Fatalf("%q should give one type refusal, got %+v", row.text, refusals)
			}
		})
	}
}

func TestInputChecksWholeNumbers(t *testing.T) {
	runQARows(t, Integer, []qaRow{
		{"ic-k-int-accept-1", "0", true},
		{"ic-k-int-accept-2", "12", true},
		{"ic-k-int-accept-3", "-12", true},
		{"ic-k-int-accept-4", "123456789012345678", true},
		{"ic-k-int-1", "-0", false},
		{"ic-k-int-2", "+5", false},
		{"ic-k-int-3", "007", false},
		{"ic-k-int-4", "1237.5", false},
		{"ic-k-int-5", "1e3", false},
		{"ic-k-int-6", "1,200", false},
		{"ic-k-int-7", "1 200", false},
		{"ic-k-int-8", "1234567890123456789", false},
		{"ic-k-int-9", "abc", false},
	})
}

func TestInputChecksDecimalNumbers(t *testing.T) {
	runQARows(t, Decimal, []qaRow{
		{"ic-k-dec-accept-1", "0", true},
		{"ic-k-dec-accept-2", "12.50", true},
		{"ic-k-dec-accept-3", "-12.5", true},
		{"ic-k-dec-accept-4", "-0.00", true},
		{"ic-k-dec-accept-5", "0.5", true},
		{"ic-k-dec-accept-6", "123456789012345678901234567890", true},
		{"ic-k-dec-1", ".5", false},
		{"ic-k-dec-2", "5.", false},
		{"ic-k-dec-3", "+5", false},
		{"ic-k-dec-4", "05.5", false},
		{"ic-k-dec-5", "5.5.5", false},
		{"ic-k-dec-6", "1e3", false},
		{"ic-k-dec-7", "1,5", false},
		{"ic-k-dec-8", "1,200.50", false},
		{"ic-k-dec-9", "95%", false},
		{"ic-k-dec-10", "1 200,50", false},
		{"ic-k-dec-11", "12.50 ", false},
		{"ic-k-dec-12", "1234567890123456789012345678901", false},
		{"ic-k-dec-13", "NaN", false},
	})
}

func TestInputChecksApproximateNumbers(t *testing.T) {
	runQARows(t, Float, []qaRow{
		{"ic-k-flt-accept-1", "0.37", true},
		{"ic-k-flt-accept-2", "5", true},
		{"ic-k-flt-accept-3", "-1.5", true},
		{"ic-k-flt-1", ".37", false},
		{"ic-k-flt-2", "0,37", false},
		{"ic-k-flt-3", "1e3", false},
	})
}

func TestInputChecksOtherTypes(t *testing.T) {
	runQARows(t, Text, []qaRow{
		{"ic-l-text-accept-1", "ok", true},
		{"ic-l-text-accept-2", "", true},
		{"ic-l-text-accept-3", strings.Repeat("a", 200), true},
		{"ic-l-text-1", strings.Repeat("a", 201), false},
		{"ic-l-text-2", "a\tb", false},
		{"ic-l-text-3", "a\nb", false},
	})
	runQARows(t, Boolean, []qaRow{
		{"ic-l-bool-accept-1", "true", true},
		{"ic-l-bool-accept-2", "false", true},
		{"ic-l-bool-1", "True", false},
		{"ic-l-bool-2", "TRUE", false},
		{"ic-l-bool-3", "yes", false},
		{"ic-l-bool-4", "1", false},
	})
	runQARows(t, Date, []qaRow{
		{"ic-l-date-accept-1", "2026-03-31", true},
		{"ic-l-date-accept-2", "2028-02-29", true},
		{"ic-l-date-1", "2026-02-30", false},
		{"ic-l-date-2", "2027-02-29", false},
		{"ic-l-date-3", "2026-13-01", false},
		{"ic-l-date-4", "2026-3-1", false},
		{"ic-l-date-5", "31/03/2026", false},
	})
}

func TestInputChecksListCases(t *testing.T) {
	t.Run("ic-l-list-1", func(t *testing.T) {
		_, refusals := Read(Decimal, Many("45.25"))
		if len(refusals) != 1 || refusals[0].Reason != ReasonListForSingle || refusals[0].Received != "[45.25]" {
			t.Fatalf("got %+v", refusals)
		}
	})
	t.Run("ic-l-list-2", func(t *testing.T) {
		_, refusals := Read(DecimalList, One("200.00"))
		if len(refusals) != 1 || refusals[0].Reason != ReasonTextForList || refusals[0].Received != "200.00" {
			t.Fatalf("got %+v", refusals)
		}
	})
	t.Run("ic-l-list-3", func(t *testing.T) {
		_, refusals := Read(DecimalList, Many("200.00", "1,200.50", "100.00"))
		if len(refusals) != 1 || refusals[0].Item != 2 || refusals[0].Received != "1,200.50" {
			t.Fatalf("got %+v", refusals)
		}
	})
	t.Run("ic-j12", func(t *testing.T) {
		_, refusals := Read(DecimalList, Many(""))
		if len(refusals) != 1 || refusals[0].Item != 1 || refusals[0].Reason != ReasonEmpty || refusals[0].Received != "" {
			t.Fatalf("got %+v", refusals)
		}
	})
}

func TestAnAcceptedValueKeepsEveryDigitAsReceived(t *testing.T) {
	for _, text := range []string{"12.50", "-12.5", "0.4370", "5.125", "45.20"} {
		value, refusals := Read(Decimal, One(text))
		if len(refusals) != 0 || value.Decimal.String() != text {
			t.Errorf("%q came back as %q (%+v)", text, value.Decimal.String(), refusals)
		}
	}
}

func TestANegativeZeroIsAcceptedAndKeepsItsTextAndPlaces(t *testing.T) {
	value, refusals := Read(Decimal, One("-0.00"))
	if len(refusals) != 0 || value.Text != "-0.00" || !value.Decimal.IsZero() || value.Decimal.Scale() != 2 {
		t.Fatalf("got %+v %+v", value, refusals)
	}
}
