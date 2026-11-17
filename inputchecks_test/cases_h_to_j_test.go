package inputchecks

import "testing"

func casesH() []testCase {
	return []testCase{
		clean("ic-h01", d6, d6.clean, 2),
		bad("ic-h02", d1, row("balance", "120.50", "payment", "0.00"), 2,
			"check-failed · check \"payment > 0\" — The payment must be more than zero. ~payment~ is 0.00."),
		bad("ic-h03", d1, row("balance", "100.00", "payment", "2000.00"), 2,
			"check-failed · check \"payment <= balance + 1000.00\" — The payment cannot be more than 1000.00 above the balance. ~payment~ is 2000.00 and ~balance~ is 100.00."),
		bad("ic-h04", d6, row("principal", "50.00", "monthly_rate", "0.050", "months", "60"), 2,
			"check-failed · check \"principal >= 100.00\" — The loan must be at least 100.00. ~principal~ is 50.00.",
			"check-failed · check \"principal * monthly_rate * months <= principal\" — The interest cannot be more than the amount borrowed. ~principal~ is 50.00, ~monthly_rate~ is 0.050 and ~months~ is 60."),
		clean("ic-h05", d6, row("principal", "100.00", "monthly_rate", "0.050", "months", "20"), 2),
		bad("ic-h06", d5, row("balances", []string{"420.00", "380.50"}, "floor", "350.00"), 1,
			"check-failed · check \"count(balances) >= 6\" — At least 6 balances are needed. ~balances~ is [420.00, 380.50]."),
		bad("ic-h07", d7, row("total", "10.00", "people", "0"), 1,
			"check-error · check \"total / people >= 1.00\" — The check \"Each person must get at least 1.00\" could not be worked out: The part ~total / people~ divides by zero."),
		clean("ic-h08", d7, d7.clean, 1),
		bad("ic-h09", d6, row("principal", "50.00", "monthly_rate", "0.06", "months", "60"), 1,
			"too-large · monthly_rate · 0.06 — ~monthly_rate~ is 0.06. It must be from 0.000 to 0.050.",
			"check-failed · check \"principal >= 100.00\" — The loan must be at least 100.00. ~principal~ is 50.00."),
		bad("ic-h10", d6, row("principal", "50.00", "monthly_rate", "0.050", "months", "60"), 2,
			"check-failed · check \"principal >= 100.00\" — The loan must be at least 100.00. ~principal~ is 50.00.",
			"check-failed · check \"principal * monthly_rate * months <= principal\" — The interest cannot be more than the amount borrowed. ~principal~ is 50.00, ~monthly_rate~ is 0.050 and ~months~ is 60."),
		hold(bad("ic-h11", d6, row("principal", "-1.00", "monthly_rate", "0.012", "months", "7"), 0,
			"too-small · principal · -1.00 — ~principal~ is -1.00. It must be more than 0.00."), decision22),
		bad("ic-h11-s", d6, row("principal", "0.00", "monthly_rate", "0.012", "months", "7"), 0,
			"too-small · principal · 0.00 — ~principal~ is 0.00. It must be more than 0.00."),
		bad("ic-h12", d6, row("principal", "2400.00", "monthly_rate", "0.012"), 1,
			"required · months · not given — ~months~ was not given. It is required."),
		bad("ic-h13", d5, row("balances", []string{"420.00", "abc"}, "floor", "350.00"), 0,
			"wrong-type · balances item 2 · abc — Item 2 of ~balances~ is abc. It must be a decimal number with at most 30 digits, such as 12.50."),
		bad("ic-h14", d6, row("principal", "50.00", "monthly_rate", "0.012", "months", "7", "memo", "hello"), 2,
			"unknown-input · memo · hello — ~memo~ is not an input of this function. Its inputs are: principal, monthly_rate, months.",
			"check-failed · check \"principal >= 100.00\" — The loan must be at least 100.00. ~principal~ is 50.00."),
	}
}

func casesI() []testCase {
	return []testCase{
		bad("ic-i01", d2, row("kilos", "-3", "price_per_kilo", "0.0005"), unchecked,
			"no-negatives · kilos · -3 — ~kilos~ is -3. It cannot be negative.",
			"decimal-places · price_per_kilo · 0.0005 — ~price_per_kilo~ is 0.0005. It can have at most 3 decimal places.",
			"required · credit · not given — ~credit~ was not given. It is required."),
		bad("ic-i02", d9, row("deposits", []string{"-1.00", "5.00"}, "expenses", []string{"-2.00"}, "members", "0"), unchecked,
			"no-negatives · deposits item 1 · -1.00 — Item 1 of ~deposits~ is -1.00. It cannot be negative.",
			"no-negatives · expenses item 1 · -2.00 — Item 1 of ~expenses~ is -2.00. It cannot be negative.",
			"too-small · members · 0 — ~members~ is 0. It must be from 1 to 500."),
		bad("ic-i03", d1, row("balance", "-1.00", "payment", "45.25", "zeta", "1", "alpha", "2"), 1,
			"no-negatives · balance · -1.00 — ~balance~ is -1.00. It cannot be negative.",
			"unknown-input · alpha · 2 — ~alpha~ is not an input of this function. Its inputs are: balance, payment.",
			"unknown-input · zeta · 1 — ~zeta~ is not an input of this function. Its inputs are: balance, payment."),
		bad("ic-i04", d6, row("principal", "50.00", "monthly_rate", "0.06", "memo", "hello"), 1,
			"too-large · monthly_rate · 0.06 — ~monthly_rate~ is 0.06. It must be from 0.000 to 0.050.",
			"required · months · not given — ~months~ was not given. It is required.",
			"unknown-input · memo · hello — ~memo~ is not an input of this function. Its inputs are: principal, monthly_rate, months.",
			"check-failed · check \"principal >= 100.00\" — The loan must be at least 100.00. ~principal~ is 50.00."),
		bad("ic-i05", d6, row(), 0,
			"required · principal · not given — ~principal~ was not given. It is required.",
			"required · monthly_rate · not given — ~monthly_rate~ was not given. It is required.",
			"required · months · not given — ~months~ was not given. It is required."),
		bad("ic-i06", d2, row("kilos", "abc", "price_per_kilo", "0.0005", "credit", "120.00"), unchecked,
			"wrong-type · kilos · abc — ~kilos~ is abc. It must be a whole number with at most 18 digits.",
			"decimal-places · price_per_kilo · 0.0005 — ~price_per_kilo~ is 0.0005. It can have at most 3 decimal places."),
		bad("ic-i07", d2, row("kilos", "-1.5", "price_per_kilo", "0.437", "credit", "120.00"), unchecked,
			"wrong-type · kilos · -1.5 — ~kilos~ is -1.5. It must be a whole number with at most 18 digits."),
	}
}

func casesJ() []testCase {
	return []testCase{
		bad("ic-j01", d3, without(d3.clean, "status"), unchecked,
			"required · status · not given — ~status~ was not given. It is required."),
		bad("ic-j02", d3, with(d3.clean, "status", ""), unchecked,
			"not-allowed · status · (empty) — ~status~ is \"\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		bad("ic-j03", d3, with(d3.clean, "status", "   "), unchecked,
			"not-allowed · status · three spaces — ~status~ is \"   \". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		bad("ic-j04", d1, row("balance", "80.00"), 0,
			"required · payment · not given — ~payment~ was not given. It is required."),
		bad("ic-j05", d1, row("balance", "80.00", "payment", ""), 0,
			"wrong-type · payment · (empty) — ~payment~ is . It must be a decimal number with at most 30 digits, such as 12.50."),
		bad("ic-j06", d1, row("balance", "80.00", "payment", " "), 0,
			"wrong-type · payment · one space — ~payment~ is  . It must be a decimal number with at most 30 digits, such as 12.50."),
		bad("ic-j07", d8, with(d8.clean, "active", ""), unchecked,
			"wrong-type · active · (empty) — ~active~ is . It must be true or false."),
		clean("ic-j08", d8, with(d8.clean, "note", ""), unchecked),
		clean("ic-j09", d8, with(d8.clean, "note", "   "), unchecked),
		bad("ic-j10", d3, without(d3.clean, "contributions"), unchecked,
			"list-too-short · contributions · not given — ~contributions~ was not given, so it has 0 items. It needs at least 1 item."),
		bad("ic-j11", d3, with(d3.clean, "contributions", []string{}), unchecked,
			"list-too-short · contributions · [] — ~contributions~ has 0 items. It needs at least 1 item."),
		bad("ic-j12", d3, with(d3.clean, "contributions", []string{""}), unchecked,
			"wrong-type · contributions item 1 · (empty) — Item 1 of ~contributions~ is . It must be a decimal number with at most 30 digits, such as 12.50."),
	}
}

func TestSectionHEachCheckOnItsOwn(t *testing.T)            { runCases(t, casesH()) }
func TestSectionISeveralProblemsAndTheirOrder(t *testing.T) { runCases(t, casesI()) }
func TestSectionJMissingEmptyAndSpaceOnly(t *testing.T)     { runCases(t, casesJ()) }
