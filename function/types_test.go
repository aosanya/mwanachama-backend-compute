package function

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type typeCase struct {
	text     string
	accepted bool
	shown    string
}

func runTypeCases(t *testing.T, prefix string, name string, words string, quoted bool, base Row, cases []typeCase) {
	t.Helper()
	for i, c := range cases {
		c := c
		id := prefix + "-" + itoaCase(i+1)
		t.Run(id, func(t *testing.T) {
			var want []problem.Problem
			if !c.accepted {
				shown := c.shown
				if shown == "" {
					shown = c.text
				}
				value := shown
				if quoted {
					value = "\"" + shown + "\""
				}
				p := on("wrong-type", name, c.text, "~"+name+"~ is "+value+". "+words)
				want = []problem.Problem{p}
			}
			runCase(t, d8, with(base, name, c.text), want...)
		})
	}
}

func itoaCase(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func TestWholeNumbersAsText(t *testing.T) {
	runTypeCases(t, "ic-k-int", "kilos", "It must be a whole number with at most 18 digits.", false, d8Clean, []typeCase{
		{"0", true, ""}, {"12", true, ""}, {"-12", true, ""}, {"123456789012345678", true, ""},
		{"-0", false, ""}, {"+5", false, ""}, {"007", false, ""}, {"1237.5", false, ""}, {"1e3", false, ""},
		{"1,200", false, ""}, {"1 200", false, ""}, {"1234567890123456789", false, ""}, {"abc", false, ""},
	})
}

func TestDecimalNumbersAsText(t *testing.T) {
	runTypeCases(t, "ic-k-dec", "payment", "It must be a decimal number with at most 30 digits, such as 12.50.", false, d8Clean, []typeCase{
		{"0", true, ""}, {"12.50", true, ""}, {"-12.5", true, ""}, {"-0.00", true, ""}, {"0.5", true, ""},
		{"123456789012345678901234567890", true, ""},
		{".5", false, ""}, {"5.", false, ""}, {"+5", false, ""}, {"05.5", false, ""}, {"5.5.5", false, ""}, {"1e3", false, ""},
		{"1,5", false, ""}, {"1,200.50", false, ""}, {"95%", false, ""}, {"1 200,50", false, ""}, {"12.50 ", false, ""},
		{"1234567890123456789012345678901", false, ""}, {"NaN", false, ""},
	})
}

func TestApproximateNumbersAsText(t *testing.T) {
	runTypeCases(t, "ic-k-flt", "ratio", "It must be a number with at most 30 digits, such as 0.37.", false, d8Clean, []typeCase{
		{"0.37", true, ""}, {"5", true, ""}, {"-1.5", true, ""},
		{".37", false, ""}, {"0,37", false, ""}, {"1e3", false, ""},
	})
}

func TestTextAsText(t *testing.T) {
	long := strings.Repeat("a", 200)
	tooLong := strings.Repeat("a", 201)
	words := "It must be text of at most 200 characters."
	accepted := []string{"ok", "", long}
	for i, text := range accepted {
		text := text
		t.Run("ic-l-text-accepted-"+itoaCase(i+1), func(t *testing.T) {
			runCase(t, d8, with(d8Clean, "note", text))
		})
	}
	t.Run("ic-l-text-4", func(t *testing.T) {
		runCase(t, d8, with(d8Clean, "note", tooLong), on("wrong-type", "note", long+"...", "~note~ is \""+long+"...\". "+words))
	})
	t.Run("ic-l-text-5", func(t *testing.T) {
		runCase(t, d8, with(d8Clean, "note", "a\tb"), on("wrong-type", "note", "a\tb", "~note~ is \"a\tb\". "+words))
	})
	t.Run("ic-l-text-6", func(t *testing.T) {
		runCase(t, d8, with(d8Clean, "note", "a\nb"), on("wrong-type", "note", "a\nb", "~note~ is \"a\nb\". "+words))
	})
}

func TestYesOrNoAsText(t *testing.T) {
	runTypeCases(t, "ic-l-bool", "active", "It must be true or false.", false, d8Clean, []typeCase{
		{"true", true, ""}, {"false", true, ""}, {"True", false, ""}, {"TRUE", false, ""}, {"yes", false, ""}, {"1", false, ""},
	})
}

func TestDatesAsText(t *testing.T) {
	runTypeCases(t, "ic-l-date", "filed", "It must be a date written like 2026-03-31.", false, d8Clean, []typeCase{
		{"2026-03-31", true, ""}, {"2028-02-29", true, ""}, {"2026-02-30", false, ""}, {"2027-02-29", false, ""},
		{"2026-13-01", false, ""}, {"2026-3-1", false, ""}, {"31/03/2026", false, ""},
	})
}

func TestAListWhereASingleValueBelongsAndTheReverse(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-l-list-1", d1, row("balance", "80.00", "payment", []string{"45.25"}), []problem.Problem{on("wrong-type", "payment", "[45.25]", "~payment~ is [45.25]. It must be a decimal number with at most 30 digits, such as 12.50.")}},
		{"ic-l-list-2", d3, with(d3Clean, "contributions", "200.00"), []problem.Problem{on("wrong-type", "contributions", "200.00", "~contributions~ is 200.00. It must be a list of decimal numbers.")}},
		{"ic-l-list-3", d3, with(d3Clean, "contributions", []string{"200.00", "1,200.50", "100.00"}), []problem.Problem{at("wrong-type", "contributions", 2, "1,200.50", "Item 2 of ~contributions~ is 1,200.50. It must be a decimal number with at most 30 digits, such as 12.50.")}},
	})
}
