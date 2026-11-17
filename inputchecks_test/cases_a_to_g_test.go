package inputchecks

import (
	"fmt"
	"strings"
	"testing"
)

func listOf(text string, count int) string {
	shown := count
	tail := ""
	if count > 20 {
		shown = 20
		tail = ", ..."
	}
	return "[" + strings.Join(repeat(text, shown), ", ") + tail + "]"
}

func casesA() []testCase {
	return []testCase{
		clean("ic-a01", d1, row("balance", "120.50", "payment", "45.25"), 2),
		bad("ic-a02", d1, row("balance", "200.00"), 0,
			"required · payment · not given — ~payment~ was not given. It is required."),
		bad("ic-a03", d1, row("payment", "45.25"), 1,
			"required · balance · not given — ~balance~ was not given. It is required."),
		bad("ic-a04", d1, row(), 0,
			"required · balance · not given — ~balance~ was not given. It is required.",
			"required · payment · not given — ~payment~ was not given. It is required."),
		bad("ic-a05", d2, without(d2.clean, "credit"), unchecked,
			"required · credit · not given — ~credit~ was not given. It is required."),
		bad("ic-a06", d8, without(d8.clean, "active"), unchecked,
			"required · active · not given — ~active~ was not given. It is required."),
		bad("ic-a07", d8, without(d8.clean, "filed"), unchecked,
			"required · filed · not given — ~filed~ was not given. It is required."),
		bad("ic-a08", d8, without(d8.clean, "note"), unchecked,
			"required · note · not given — ~note~ was not given. It is required."),
	}
}

func casesB() []testCase {
	thirteenAndABC := append(repeat("70", 12), "abc")
	return []testCase{
		clean("ic-b01", d3, d3.clean, unchecked),
		bad("ic-b02", d3, with(d3.clean, "contributions", []string{}), unchecked,
			"list-too-short · contributions · [] — ~contributions~ has 0 items. It needs at least 1 item."),
		bad("ic-b03", d3, without(d3.clean, "contributions"), unchecked,
			"list-too-short · contributions · not given — ~contributions~ was not given, so it has 0 items. It needs at least 1 item."),
		clean("ic-b04", d9, without(d9.clean, "expenses"), unchecked),
		clean("ic-b05", d9, with(d9.clean, "expenses", []string{}), unchecked),
		clean("ic-b06", d4, row("scores", repeat("70", 12)), unchecked),
		bad("ic-b07", d4, row("scores", repeat("70", 13)), unchecked,
			"list-too-long · scores · "+listOf("70", 13)+" — ~scores~ has 13 items. It can have at most 12 items."),
		bad("ic-b08", d4, row("scores", repeat("70", 21)), unchecked,
			"list-too-long · scores · "+listOf("70", 21)+" — ~scores~ has 21 items. It can have at most 12 items."),
		bad("ic-b09", d4, row("scores", thirteenAndABC), unchecked,
			"list-too-long · scores · [70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, abc] — ~scores~ has 13 items. It can have at most 12 items.",
			"wrong-type · scores item 13 · abc — Item 13 of ~scores~ is abc. It must be a whole number with at most 18 digits."),
		bad("ic-b10", d4, row("scores", []string{}), unchecked,
			"list-too-short · scores · [] — ~scores~ has 0 items. It needs at least 1 item."),
		bad("ic-b11", d4, row(), unchecked,
			"list-too-short · scores · not given — ~scores~ was not given, so it has 0 items. It needs at least 1 item."),
		bad("ic-b12", d4, row("scores", "70"), unchecked,
			"wrong-type · scores · 70 — ~scores~ is 70. It must be a list of whole numbers."),
	}
}

func casesC() []testCase {
	return []testCase{
		clean("ic-c01", d1, row("balance", "0.00", "payment", "45.25"), 2),
		bad("ic-c02", d1, row("balance", "80.00", "payment", "-5.00"), 0,
			"no-negatives · payment · -5.00 — ~payment~ is -5.00. It cannot be negative."),
		bad("ic-c03", d1, row("balance", "80.00", "payment", "-0.00"), 2,
			"check-failed · check \"payment > 0\" — The payment must be more than zero. ~payment~ is -0.00."),
		bad("ic-c04", d1, row("balance", "-1.00", "payment", "-5.125"), 0,
			"no-negatives · balance · -1.00 — ~balance~ is -1.00. It cannot be negative.",
			"no-negatives · payment · -5.125 — ~payment~ is -5.125. It cannot be negative.",
			"decimal-places · payment · -5.125 — ~payment~ is -5.125. It can have at most 2 decimal places."),
		bad("ic-c05", d3, with(d3.clean, "contributions", []string{"200.00", "-50.00"}), unchecked,
			"no-negatives · contributions item 2 · -50.00 — Item 2 of ~contributions~ is -50.00. It cannot be negative."),
		bad("ic-c06", d10, with(d10.clean, "units", "-3"), unchecked,
			"no-negatives · units · -3 — ~units~ is -3. It cannot be negative.",
			"too-small · units · -3 — ~units~ is -3. It must be from 1 to 100."),
		hold(bad("ic-c07", d2, with(d2.clean, "price_per_kilo", "-0.5"), unchecked,
			"too-small · price_per_kilo · -0.5 — ~price_per_kilo~ is -0.5. It must be more than 0.000."), decision22),
	}
}

func casesD() []testCase {
	return []testCase{
		clean("ic-d01", d9, d9.clean, unchecked),
		clean("ic-d02", d9, with(d9.clean, "members", "1"), unchecked),
		clean("ic-d03", d9, with(d9.clean, "members", "500"), unchecked),
		bad("ic-d04", d9, with(d9.clean, "members", "0"), unchecked,
			"too-small · members · 0 — ~members~ is 0. It must be from 1 to 500."),
		bad("ic-d05", d9, with(d9.clean, "members", "501"), unchecked,
			"too-large · members · 501 — ~members~ is 501. It must be from 1 to 500."),
		clean("ic-d06", d3, with(d3.clean, "requested", "0.01"), unchecked),
		bad("ic-d07", d3, with(d3.clean, "requested", "0.00"), unchecked,
			"too-small · requested · 0.00 — ~requested~ is 0.00. It must be more than 0.00."),
		clean("ic-d08", d6, row("principal", "2400.00", "monthly_rate", "0.05", "months", "7"), 2),
		bad("ic-d09", d6, row("principal", "2400.00", "monthly_rate", "0.0500", "months", "7"), 1,
			"decimal-places · monthly_rate · 0.0500 — ~monthly_rate~ is 0.0500. It can have at most 3 decimal places."),
		clean("ic-d10-1", d10, with(d10.clean, "share_pct", "0.00"), unchecked),
		clean("ic-d10-2", d10, with(d10.clean, "share_pct", "99.99"), unchecked),
		hold(bad("ic-d11", d10, with(d10.clean, "share_pct", "-0.01"), unchecked,
			"too-small · share_pct · -0.01 — ~share_pct~ is -0.01. It must be 0.00 or more."), decision22),
		bad("ic-d12", d10, with(d10.clean, "share_pct", "100.00"), unchecked,
			"too-large · share_pct · 100.00 — ~share_pct~ is 100.00. It must be less than 100.00."),
		clean("ic-d13", d10, with(d10.clean, "bonus", "500.00"), unchecked),
		bad("ic-d14", d10, with(d10.clean, "bonus", "500.01"), unchecked,
			"too-large · bonus · 500.01 — ~bonus~ is 500.01. It must be 500.00 or less."),
		bad("ic-d15", d10, with(d10.clean, "units", "0"), unchecked,
			"too-small · units · 0 — ~units~ is 0. It must be from 1 to 100."),
		bad("ic-d16", d10, with(d10.clean, "units", "101"), unchecked,
			"too-large · units · 101 — ~units~ is 101. It must be from 1 to 100."),
	}
}

func casesE() []testCase {
	return []testCase{
		clean("ic-e01", d1, row("balance", "120.50", "payment", "45.2"), 2),
		bad("ic-e02", d1, row("balance", "60.00", "payment", "5.125"), 0,
			"decimal-places · payment · 5.125 — ~payment~ is 5.125. It can have at most 2 decimal places."),
		clean("ic-e03", d2, d2.clean, unchecked),
		bad("ic-e04", d2, with(d2.clean, "price_per_kilo", "0.4370"), unchecked,
			"decimal-places · price_per_kilo · 0.4370 — ~price_per_kilo~ is 0.4370. It can have at most 3 decimal places."),
		clean("ic-e05", d10, d10.clean, unchecked),
		bad("ic-e06", d10, with(d10.clean, "bags", "2.5"), unchecked,
			"decimal-places · bags · 2.5 — ~bags~ is 2.5. It cannot have decimal places."),
		bad("ic-e07", d10, with(d10.clean, "bags", "3.0"), unchecked,
			"decimal-places · bags · 3.0 — ~bags~ is 3.0. It cannot have decimal places."),
		bad("ic-e08", d10, with(d10.clean, "weight", "2.25"), unchecked,
			"decimal-places · weight · 2.25 — ~weight~ is 2.25. It can have at most 1 decimal place."),
		bad("ic-e09", d5, with(d5.clean, "balances", []string{"420.00", "380.505", "95.25", "510.00", "440.75", "300.00"}), 0,
			"decimal-places · balances item 2 · 380.505 — Item 2 of ~balances~ is 380.505. It can have at most 2 decimal places."),
	}
}

func casesF() []testCase {
	return []testCase{
		clean("ic-f01", d3, with(d3.clean, "status", "lapsed"), unchecked),
		bad("ic-f02", d3, with(d3.clean, "status", "paused"), unchecked,
			"not-allowed · status · paused — ~status~ is \"paused\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		bad("ic-f03", d3, with(d3.clean, "status", "Active"), unchecked,
			"not-allowed · status · Active — ~status~ is \"Active\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		bad("ic-f04", d3, with(d3.clean, "status", "active "), unchecked,
			"not-allowed · status · active␣ — ~status~ is \"active␣\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
	}
}

func thousandItems() testCase {
	received := row("amounts", repeat("-5.125", 1000))
	want := lines()
	for item := 1; item <= 1000; item++ {
		want = append(want,
			line(fmt.Sprintf("no-negatives · amounts item %d · -5.125 — Item %d of ~amounts~ is -5.125. It cannot be negative.", item, item)),
			line(fmt.Sprintf("too-small · amounts item %d · -5.125 — Item %d of ~amounts~ is -5.125. It must be from 0.00 to 100.00.", item, item)),
			line(fmt.Sprintf("decimal-places · amounts item %d · -5.125 — Item %d of ~amounts~ is -5.125. It can have at most 2 decimal places.", item, item)),
		)
	}
	return testCase{id: "ic-g06", def: d11, row: received, want: want, ran: unchecked}
}

func casesG() []testCase {
	return []testCase{
		clean("ic-g01", d4, d4.clean, unchecked),
		bad("ic-g02", d4, row("scores", []string{"72", "104"}), unchecked,
			"too-large · scores item 2 · 104 — Item 2 of ~scores~ is 104. It must be from 0 to 100."),
		hold(bad("ic-g03", d4, row("scores", []string{"-3", "104", "abc"}), unchecked,
			"too-small · scores item 1 · -3 — Item 1 of ~scores~ is -3. It must be from 0 to 100.",
			"too-large · scores item 2 · 104 — Item 2 of ~scores~ is 104. It must be from 0 to 100.",
			"wrong-type · scores item 3 · abc — Item 3 of ~scores~ is abc. It must be a whole number with at most 18 digits."), decision22),
		bad("ic-g04", d11, row("amounts", []string{"-5.125"}), unchecked,
			"no-negatives · amounts item 1 · -5.125 — Item 1 of ~amounts~ is -5.125. It cannot be negative.",
			"too-small · amounts item 1 · -5.125 — Item 1 of ~amounts~ is -5.125. It must be from 0.00 to 100.00.",
			"decimal-places · amounts item 1 · -5.125 — Item 1 of ~amounts~ is -5.125. It can have at most 2 decimal places."),
		bad("ic-g05", d11, row("amounts", []string{"10.00", "250.00", "7.5"}), unchecked,
			"too-large · amounts item 2 · 250.00 — Item 2 of ~amounts~ is 250.00. It must be from 0.00 to 100.00."),
		thousandItems(),
		bad("ic-g03-s", d4, row("scores", []string{"101", "104", "abc"}), unchecked,
			"too-large · scores item 1 · 101 — Item 1 of ~scores~ is 101. It must be from 0 to 100.",
			"too-large · scores item 2 · 104 — Item 2 of ~scores~ is 104. It must be from 0 to 100.",
			"wrong-type · scores item 3 · abc — Item 3 of ~scores~ is abc. It must be a whole number with at most 18 digits."),
	}
}

func TestSectionARequired(t *testing.T)              { runCases(t, casesA()) }
func TestSectionBListSizes(t *testing.T)             { runCases(t, casesB()) }
func TestSectionCNoNegatives(t *testing.T)           { runCases(t, casesC()) }
func TestSectionDRange(t *testing.T)                 { runCases(t, casesD()) }
func TestSectionEDecimalPlaces(t *testing.T)         { runCases(t, casesE()) }
func TestSectionFAllowedValues(t *testing.T)         { runCases(t, casesF()) }
func TestSectionGRuleOnEachItemOfAList(t *testing.T) { runCases(t, casesG()) }
