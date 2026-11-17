package function

import (
	"strconv"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type validateCase struct {
	id         string
	definition Definition
	received   Row
	want       []problem.Problem
}

func runCases(t *testing.T, cases []validateCase) {
	t.Helper()
	for _, c := range cases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			runCase(t, c.definition, c.received, c.want...)
		})
	}
}

func scoresOf(text string, count int) []string { return repeated(text, count) }

func TestRequired(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-a01", d1, d1Clean, nil},
		{"ic-a02", d1, row("balance", "200.00"), []problem.Problem{missing("payment")}},
		{"ic-a03", d1, row("payment", "45.25"), []problem.Problem{missing("balance")}},
		{"ic-a04", d1, row(), []problem.Problem{missing("balance"), missing("payment")}},
		{"ic-a05", d2, without(d2Clean, "credit"), []problem.Problem{missing("credit")}},
		{"ic-a06", d8, without(d8Clean, "active"), []problem.Problem{missing("active")}},
		{"ic-a07", d8, without(d8Clean, "filed"), []problem.Problem{missing("filed")}},
		{"ic-a08", d8, without(d8Clean, "note"), []problem.Problem{missing("note")}},
	})
}

func TestListSizes(t *testing.T) {
	tooShort := "~contributions~ has 0 items. It needs at least 1 item."
	scores13 := "[" + strings.Join(scoresOf("70", 13), ", ") + "]"
	scores21 := "[" + strings.Join(scoresOf("70", 20), ", ") + ", ...]"
	runCases(t, []validateCase{
		{"ic-b01", d3, d3Clean, nil},
		{"ic-b02", d3, with(d3Clean, "contributions", []string{}), []problem.Problem{on("list-too-short", "contributions", "[]", tooShort)}},
		{"ic-b03", d3, without(d3Clean, "contributions"), []problem.Problem{notGiven("list-too-short", "contributions", "~contributions~ was not given, so it has 0 items. It needs at least 1 item.")}},
		{"ic-b04", d9, without(d9Clean, "expenses"), nil},
		{"ic-b05", d9, with(d9Clean, "expenses", []string{}), nil},
		{"ic-b06", d4, row("scores", scoresOf("70", 12)), nil},
		{"ic-b07", d4, row("scores", scoresOf("70", 13)), []problem.Problem{on("list-too-long", "scores", scores13, "~scores~ has 13 items. It can have at most 12 items.")}},
		{"ic-b08", d4, row("scores", scoresOf("70", 21)), []problem.Problem{on("list-too-long", "scores", scores21, "~scores~ has 21 items. It can have at most 12 items.")}},
		{"ic-b09", d4, row("scores", append(scoresOf("70", 12), "abc")), []problem.Problem{
			on("list-too-long", "scores", "["+strings.Join(scoresOf("70", 12), ", ")+", abc]", "~scores~ has 13 items. It can have at most 12 items."),
			at("wrong-type", "scores", 13, "abc", "Item 13 of ~scores~ is abc. It must be a whole number with at most 18 digits."),
		}},
		{"ic-b10", d4, row("scores", []string{}), []problem.Problem{on("list-too-short", "scores", "[]", "~scores~ has 0 items. It needs at least 1 item.")}},
		{"ic-b11", d4, row(), []problem.Problem{notGiven("list-too-short", "scores", "~scores~ was not given, so it has 0 items. It needs at least 1 item.")}},
		{"ic-b12", d4, row("scores", "70"), []problem.Problem{on("wrong-type", "scores", "70", "~scores~ is 70. It must be a list of whole numbers.")}},
	})
}

func TestNoNegatives(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-c01", d1, row("balance", "0.00", "payment", "45.25"), nil},
		{"ic-c02", d1, row("balance", "80.00", "payment", "-5.00"), []problem.Problem{on("no-negatives", "payment", "-5.00", "~payment~ is -5.00. It cannot be negative.")}},
		{"ic-c03", d1, row("balance", "80.00", "payment", "-0.00"), []problem.Problem{ofCheck("check-failed", "payment > 0", "The payment must be more than zero. ~payment~ is -0.00.")}},
		{"ic-c04", d1, row("balance", "-1.00", "payment", "-5.125"), []problem.Problem{
			on("no-negatives", "balance", "-1.00", "~balance~ is -1.00. It cannot be negative."),
			on("no-negatives", "payment", "-5.125", "~payment~ is -5.125. It cannot be negative."),
			on("decimal-places", "payment", "-5.125", "~payment~ is -5.125. It can have at most 2 decimal places."),
		}},
		{"ic-c05", d3, with(d3Clean, "contributions", []string{"200.00", "-50.00"}), []problem.Problem{at("no-negatives", "contributions", 2, "-50.00", "Item 2 of ~contributions~ is -50.00. It cannot be negative.")}},
		{"ic-c06", d10, with(d10Clean, "units", "-3"), []problem.Problem{
			on("no-negatives", "units", "-3", "~units~ is -3. It cannot be negative."),
			on("too-small", "units", "-3", "~units~ is -3. It must be from 1 to 100."),
		}},
	})
}

func TestNegativesWithoutTheRulePendingDecision22(t *testing.T) {
	t.Log("ic-c07 is pending: decision 22 says an input without a declaration refuses negatives; the language page accepts them. Built as the page is written.")
	runCases(t, []validateCase{
		{"ic-c07", d2, with(d2Clean, "price_per_kilo", "-0.5"), []problem.Problem{on("too-small", "price_per_kilo", "-0.5", "~price_per_kilo~ is -0.5. It must be more than 0.000.")}},
	})
}

func TestRange(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-d01", d9, with(d9Clean, "members", "7"), nil},
		{"ic-d02", d9, with(d9Clean, "members", "1"), nil},
		{"ic-d03", d9, with(d9Clean, "members", "500"), nil},
		{"ic-d04", d9, with(d9Clean, "members", "0"), []problem.Problem{on("too-small", "members", "0", "~members~ is 0. It must be from 1 to 500.")}},
		{"ic-d05", d9, with(d9Clean, "members", "501"), []problem.Problem{on("too-large", "members", "501", "~members~ is 501. It must be from 1 to 500.")}},
		{"ic-d06", d3, with(d3Clean, "requested", "0.01"), nil},
		{"ic-d07", d3, with(d3Clean, "requested", "0.00"), []problem.Problem{on("too-small", "requested", "0.00", "~requested~ is 0.00. It must be more than 0.00.")}},
		{"ic-d08", d6, with(d6Clean, "monthly_rate", "0.05"), nil},
		{"ic-d09", d6, with(d6Clean, "monthly_rate", "0.0500"), []problem.Problem{on("decimal-places", "monthly_rate", "0.0500", "~monthly_rate~ is 0.0500. It can have at most 3 decimal places.")}},
		{"ic-d10a", d10, with(d10Clean, "share_pct", "0.00"), nil},
		{"ic-d10b", d10, with(d10Clean, "share_pct", "99.99"), nil},
		{"ic-d11", d10, with(d10Clean, "share_pct", "-0.01"), []problem.Problem{on("too-small", "share_pct", "-0.01", "~share_pct~ is -0.01. It must be 0.00 or more.")}},
		{"ic-d12", d10, with(d10Clean, "share_pct", "100.00"), []problem.Problem{on("too-large", "share_pct", "100.00", "~share_pct~ is 100.00. It must be less than 100.00.")}},
		{"ic-d13", d10, with(d10Clean, "bonus", "500.00"), nil},
		{"ic-d14", d10, with(d10Clean, "bonus", "500.01"), []problem.Problem{on("too-large", "bonus", "500.01", "~bonus~ is 500.01. It must be 500.00 or less.")}},
		{"ic-d15", d10, with(d10Clean, "units", "0"), []problem.Problem{on("too-small", "units", "0", "~units~ is 0. It must be from 1 to 100.")}},
		{"ic-d16", d10, with(d10Clean, "units", "101"), []problem.Problem{on("too-large", "units", "101", "~units~ is 101. It must be from 1 to 100.")}},
	})
}

func TestDecimalPlaces(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-e01", d1, row("balance", "120.50", "payment", "45.2"), nil},
		{"ic-e02", d1, row("balance", "60.00", "payment", "5.125"), []problem.Problem{on("decimal-places", "payment", "5.125", "~payment~ is 5.125. It can have at most 2 decimal places.")}},
		{"ic-e03", d2, with(d2Clean, "price_per_kilo", "0.437"), nil},
		{"ic-e04", d2, with(d2Clean, "price_per_kilo", "0.4370"), []problem.Problem{on("decimal-places", "price_per_kilo", "0.4370", "~price_per_kilo~ is 0.4370. It can have at most 3 decimal places.")}},
		{"ic-e05", d10, with(d10Clean, "bags", "3"), nil},
		{"ic-e06", d10, with(d10Clean, "bags", "2.5"), []problem.Problem{on("decimal-places", "bags", "2.5", "~bags~ is 2.5. It cannot have decimal places.")}},
		{"ic-e07", d10, with(d10Clean, "bags", "3.0"), []problem.Problem{on("decimal-places", "bags", "3.0", "~bags~ is 3.0. It cannot have decimal places.")}},
		{"ic-e08", d10, with(d10Clean, "weight", "2.25"), []problem.Problem{on("decimal-places", "weight", "2.25", "~weight~ is 2.25. It can have at most 1 decimal place.")}},
		{"ic-e09", d5, with(d5Clean, "balances", []string{"420.00", "380.505", "95.25", "510.00", "440.75", "300.00"}), []problem.Problem{at("decimal-places", "balances", 2, "380.505", "Item 2 of ~balances~ is 380.505. It can have at most 2 decimal places.")}},
	})
}

func TestAllowedValues(t *testing.T) {
	notAllowed := func(shown string, value string) problem.Problem {
		return on("not-allowed", "status", value, "~status~ is \""+shown+"\". It must be one of \"active\", \"lapsed\", \"suspended\".")
	}
	runCases(t, []validateCase{
		{"ic-f01", d3, with(d3Clean, "status", "lapsed"), nil},
		{"ic-f02", d3, with(d3Clean, "status", "paused"), []problem.Problem{notAllowed("paused", "paused")}},
		{"ic-f03", d3, with(d3Clean, "status", "Active"), []problem.Problem{notAllowed("Active", "Active")}},
		{"ic-f04", d3, with(d3Clean, "status", "active "), []problem.Problem{notAllowed("active ", "active ")}},
	})
}

func TestARuleOnEachItem(t *testing.T) {
	var many []string
	var manyWant []problem.Problem
	for item := 1; item <= 1000; item++ {
		many = append(many, "-5.125")
		position := strconv.Itoa(item)
		manyWant = append(manyWant,
			at("no-negatives", "amounts", item, "-5.125", "Item "+position+" of ~amounts~ is -5.125. It cannot be negative."),
			at("too-small", "amounts", item, "-5.125", "Item "+position+" of ~amounts~ is -5.125. It must be from 0.00 to 100.00."),
			at("decimal-places", "amounts", item, "-5.125", "Item "+position+" of ~amounts~ is -5.125. It can have at most 2 decimal places."),
		)
	}
	runCases(t, []validateCase{
		{"ic-g01", d4, d4Clean, nil},
		{"ic-g02", d4, row("scores", []string{"72", "104"}), []problem.Problem{at("too-large", "scores", 2, "104", "Item 2 of ~scores~ is 104. It must be from 0 to 100.")}},
		{"ic-g03", d4, row("scores", []string{"-3", "104", "abc"}), []problem.Problem{
			at("too-small", "scores", 1, "-3", "Item 1 of ~scores~ is -3. It must be from 0 to 100."),
			at("too-large", "scores", 2, "104", "Item 2 of ~scores~ is 104. It must be from 0 to 100."),
			at("wrong-type", "scores", 3, "abc", "Item 3 of ~scores~ is abc. It must be a whole number with at most 18 digits."),
		}},
		{"ic-g04", d11, row("amounts", []string{"-5.125"}), []problem.Problem{
			at("no-negatives", "amounts", 1, "-5.125", "Item 1 of ~amounts~ is -5.125. It cannot be negative."),
			at("too-small", "amounts", 1, "-5.125", "Item 1 of ~amounts~ is -5.125. It must be from 0.00 to 100.00."),
			at("decimal-places", "amounts", 1, "-5.125", "Item 1 of ~amounts~ is -5.125. It can have at most 2 decimal places."),
		}},
		{"ic-g05", d11, row("amounts", []string{"10.00", "250.00", "7.5"}), []problem.Problem{at("too-large", "amounts", 2, "250.00", "Item 2 of ~amounts~ is 250.00. It must be from 0.00 to 100.00.")}},
		{"ic-g06", d11, row("amounts", many), manyWant},
	})
}
