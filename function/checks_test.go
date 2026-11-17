package function

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func TestEachCheckOnItsOwn(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-h01", d6, d6Clean, nil},
		{"ic-h02", d1, row("balance", "120.50", "payment", "0.00"), []problem.Problem{ofCheck("check-failed", "payment > 0", "The payment must be more than zero. ~payment~ is 0.00.")}},
		{"ic-h03", d1, row("balance", "100.00", "payment", "2000.00"), []problem.Problem{ofCheck("check-failed", "payment <= balance + 1000.00", "The payment cannot be more than 1000.00 above the balance. ~payment~ is 2000.00 and ~balance~ is 100.00.")}},
		{"ic-h04", d6, row("principal", "50.00", "monthly_rate", "0.050", "months", "60"), []problem.Problem{
			ofCheck("check-failed", "principal >= 100.00", "The loan must be at least 100.00. ~principal~ is 50.00."),
			ofCheck("check-failed", "principal * monthly_rate * months <= principal", "The interest cannot be more than the amount borrowed. ~principal~ is 50.00, ~monthly_rate~ is 0.050 and ~months~ is 60."),
		}},
		{"ic-h05", d6, row("principal", "100.00", "monthly_rate", "0.050", "months", "20"), nil},
		{"ic-h06", d5, row("balances", []string{"420.00", "380.50"}, "floor", "350.00"), []problem.Problem{ofCheck("check-failed", "count(balances) >= 6", "At least 6 balances are needed. ~balances~ is [420.00, 380.50].")}},
		{"ic-h07", d7, row("total", "10.00", "people", "0"), []problem.Problem{ofCheck("check-error", "total / people >= 1.00", "The check \"Each person must get at least 1.00\" could not be worked out: The part ~total / people~ divides by zero.")}},
		{"ic-h08", d7, d7Clean, nil},
		{"ic-h09", d6, row("principal", "50.00", "monthly_rate", "0.06", "months", "60"), []problem.Problem{
			on("too-large", "monthly_rate", "0.06", "~monthly_rate~ is 0.06. It must be from 0.000 to 0.050."),
			ofCheck("check-failed", "principal >= 100.00", "The loan must be at least 100.00. ~principal~ is 50.00."),
		}},
		{"ic-h10", d6, row("principal", "50.00", "monthly_rate", "0.050", "months", "60"), []problem.Problem{
			ofCheck("check-failed", "principal >= 100.00", "The loan must be at least 100.00. ~principal~ is 50.00."),
			ofCheck("check-failed", "principal * monthly_rate * months <= principal", "The interest cannot be more than the amount borrowed. ~principal~ is 50.00, ~monthly_rate~ is 0.050 and ~months~ is 60."),
		}},
		{"ic-h11", d6, row("principal", "-1.00", "monthly_rate", "0.012", "months", "7"), []problem.Problem{on("too-small", "principal", "-1.00", "~principal~ is -1.00. It must be more than 0.00.")}},
		{"ic-h12", d6, without(d6Clean, "months"), []problem.Problem{missing("months")}},
		{"ic-h13", d5, row("balances", []string{"420.00", "abc"}, "floor", "350.00"), []problem.Problem{at("wrong-type", "balances", 2, "abc", "Item 2 of ~balances~ is abc. It must be a decimal number with at most 30 digits, such as 12.50.")}},
		{"ic-h14", d6, row("principal", "50.00", "monthly_rate", "0.012", "months", "7", "memo", "hello"), []problem.Problem{
			on("unknown-input", "memo", "hello", "~memo~ is not an input of this function. Its inputs are: principal, monthly_rate, months."),
			ofCheck("check-failed", "principal >= 100.00", "The loan must be at least 100.00. ~principal~ is 50.00."),
		}},
	})
}

func TestSeveralProblemsAtOnceAndTheirOrder(t *testing.T) {
	runCases(t, []validateCase{
		{"ic-i01", d2, row("kilos", "-3", "price_per_kilo", "0.0005"), []problem.Problem{
			on("no-negatives", "kilos", "-3", "~kilos~ is -3. It cannot be negative."),
			on("decimal-places", "price_per_kilo", "0.0005", "~price_per_kilo~ is 0.0005. It can have at most 3 decimal places."),
			missing("credit"),
		}},
		{"ic-i02", d9, row("deposits", []string{"-1.00", "5.00"}, "expenses", []string{"-2.00"}, "members", "0"), []problem.Problem{
			at("no-negatives", "deposits", 1, "-1.00", "Item 1 of ~deposits~ is -1.00. It cannot be negative."),
			at("no-negatives", "expenses", 1, "-2.00", "Item 1 of ~expenses~ is -2.00. It cannot be negative."),
			on("too-small", "members", "0", "~members~ is 0. It must be from 1 to 500."),
		}},
		{"ic-i03", d1, row("balance", "-1.00", "payment", "45.25", "zeta", "1", "alpha", "2"), []problem.Problem{
			on("no-negatives", "balance", "-1.00", "~balance~ is -1.00. It cannot be negative."),
			on("unknown-input", "alpha", "2", "~alpha~ is not an input of this function. Its inputs are: balance, payment."),
			on("unknown-input", "zeta", "1", "~zeta~ is not an input of this function. Its inputs are: balance, payment."),
		}},
		{"ic-i04", d6, row("principal", "50.00", "monthly_rate", "0.06", "memo", "hello"), []problem.Problem{
			on("too-large", "monthly_rate", "0.06", "~monthly_rate~ is 0.06. It must be from 0.000 to 0.050."),
			missing("months"),
			on("unknown-input", "memo", "hello", "~memo~ is not an input of this function. Its inputs are: principal, monthly_rate, months."),
			ofCheck("check-failed", "principal >= 100.00", "The loan must be at least 100.00. ~principal~ is 50.00."),
		}},
		{"ic-i05", d6, row(), []problem.Problem{missing("principal"), missing("monthly_rate"), missing("months")}},
		{"ic-i06", d2, row("kilos", "abc", "price_per_kilo", "0.0005", "credit", "120.00"), []problem.Problem{
			on("wrong-type", "kilos", "abc", "~kilos~ is abc. It must be a whole number with at most 18 digits."),
			on("decimal-places", "price_per_kilo", "0.0005", "~price_per_kilo~ is 0.0005. It can have at most 3 decimal places."),
		}},
		{"ic-i07", d2, with(d2Clean, "kilos", "-1.5"), []problem.Problem{on("wrong-type", "kilos", "-1.5", "~kilos~ is -1.5. It must be a whole number with at most 18 digits.")}},
	})
}

func TestMissingEmptyAndSpaceOnlyAreToldApart(t *testing.T) {
	decimalWords := "It must be a decimal number with at most 30 digits, such as 12.50."
	runCases(t, []validateCase{
		{"ic-j01", d3, without(d3Clean, "status"), []problem.Problem{missing("status")}},
		{"ic-j02", d3, with(d3Clean, "status", ""), []problem.Problem{on("not-allowed", "status", "", "~status~ is \"\". It must be one of \"active\", \"lapsed\", \"suspended\".")}},
		{"ic-j03", d3, with(d3Clean, "status", "   "), []problem.Problem{on("not-allowed", "status", "   ", "~status~ is \"   \". It must be one of \"active\", \"lapsed\", \"suspended\".")}},
		{"ic-j04", d1, row("balance", "80.00"), []problem.Problem{missing("payment")}},
		{"ic-j05", d1, row("balance", "80.00", "payment", ""), []problem.Problem{on("wrong-type", "payment", "", "~payment~ is . "+decimalWords)}},
		{"ic-j06", d1, row("balance", "80.00", "payment", " "), []problem.Problem{on("wrong-type", "payment", " ", "~payment~ is  . "+decimalWords)}},
		{"ic-j07", d8, with(d8Clean, "active", ""), []problem.Problem{on("wrong-type", "active", "", "~active~ is . It must be true or false.")}},
		{"ic-j08", d8, with(d8Clean, "note", ""), nil},
		{"ic-j09", d8, with(d8Clean, "note", "   "), nil},
		{"ic-j10", d3, without(d3Clean, "contributions"), []problem.Problem{notGiven("list-too-short", "contributions", "~contributions~ was not given, so it has 0 items. It needs at least 1 item.")}},
		{"ic-j11", d3, with(d3Clean, "contributions", []string{}), []problem.Problem{on("list-too-short", "contributions", "[]", "~contributions~ has 0 items. It needs at least 1 item.")}},
		{"ic-j12", d3, with(d3Clean, "contributions", []string{""}), []problem.Problem{at("wrong-type", "contributions", 1, "", "Item 1 of ~contributions~ is . "+decimalWords)}},
	})
}

func TestNeverCalculatedAndTheSameListForEveryWayIn(t *testing.T) {
	rows := []struct {
		id         string
		definition Definition
		received   Row
	}{
		{"ic-c04", d1, row("balance", "-1.00", "payment", "-5.125")},
		{"ic-g03", d4, row("scores", []string{"-3", "104", "abc"})},
		{"ic-h04", d6, row("principal", "50.00", "monthly_rate", "0.050", "months", "60")},
		{"ic-i04", d6, row("principal", "50.00", "monthly_rate", "0.06", "memo", "hello")},
	}
	for _, c := range rows {
		c := c
		t.Run(c.id, func(t *testing.T) {
			first := asJSON(t, Validate(c.definition, c.received, &fakeRunner{}))
			second := asJSON(t, Validate(c.definition, c.received, &fakeRunner{}))
			if first != second {
				t.Errorf("two calls gave different lists:\n%s\n%s", first, second)
			}
			if first == "" {
				t.Errorf("the row has problems and the list is empty")
			}
		})
	}
}

func TestABatchOfRowsGivesEachRowItsOwnList(t *testing.T) {
	runner := &fakeRunner{}
	one := Validate(d1, row("balance", "120.50", "payment", "45.25"), runner)
	two := Validate(d1, row("balance", "-1.00", "payment", "-5.125"), runner)
	three := Validate(d1, row("balance", "120.50", "payment", "0.00"), runner)
	compare(t, one, nil)
	compare(t, two, []problem.Problem{
		on("no-negatives", "balance", "-1.00", "~balance~ is -1.00. It cannot be negative."),
		on("no-negatives", "payment", "-5.125", "~payment~ is -5.125. It cannot be negative."),
		on("decimal-places", "payment", "-5.125", "~payment~ is -5.125. It can have at most 2 decimal places."),
	})
	compare(t, three, []problem.Problem{ofCheck("check-failed", "payment > 0", "The payment must be more than zero. ~payment~ is 0.00.")})
}

func TestACheckIsNotRunWhenAnInputItReadsHasAProblem(t *testing.T) {
	runner := &fakeRunner{}
	Validate(d6, row("principal", "-1.00", "monthly_rate", "0.012", "months", "7"), runner)
	if runner.calls != 0 {
		t.Errorf("the runner was asked %d times; both checks read principal and must be skipped", runner.calls)
	}
	Validate(d6, row("principal", "50.00", "monthly_rate", "0.06", "months", "60"), runner)
	if runner.calls != 1 {
		t.Errorf("only the check that reads principal alone may run, got %d runs", runner.calls)
	}
}

func TestACheckIsNeverRunWithoutARunner(t *testing.T) {
	got := Validate(d1, d1Clean, nil)
	if len(got) != 2 {
		t.Fatalf("with no runner both checks give check-error, got %d problems", len(got))
	}
	for _, p := range got {
		if p.Code != problem.CodeCheckError {
			t.Errorf("code %q, want check-error", p.Code)
		}
	}
}
