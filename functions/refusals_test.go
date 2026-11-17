package functions

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/function"
)

var refusalCodes = map[string]string{
	"REF-fin.balance_after_payment-07": "no-negatives",
	"REF-fin.balance_after_payment-08": "decimal-places",
	"REF-fin.equal_share-06":           "too-small",
	"REF-fin.equal_share-07":           "list-too-short",
	"REF-fin.equal_share-08":           "no-negatives",
	"REF-date.late_days-07":            "required",
	"REF-date.late_days-08":            "wrong-type",
	"REF-fin.flat_instalment-07":       "too-large",
	"REF-fin.flat_instalment-08":       "too-small",
	"REF-fin.prorate_fee-07":           "too-large",
	"REF-fin.prorate_fee-08":           "no-negatives",
	"REF-fin.straight_line_value-07":   "too-small",
	"REF-fin.straight_line_value-08":   "too-small",
	"REF-fin.coverage_ratio-07":        "too-small",
	"REF-fin.coverage_ratio-08":        "no-negatives",
	"REF-fin.proportional_share-07":    "no-negatives",
	"REF-fin.proportional_share-08":    "run:divide-by-zero",
	"REF-stats.spread-07":              "list-too-short",
	"REF-stats.spread-08":              "no-negatives",
	"REF-math.remainder-06":            "too-small",
	"REF-math.remainder-07":            "no-negatives",
	"REF-math.remainder-08":            "wrong-type",
	"REF-fin.loan_within_limit-07":     "not-allowed",
	"REF-fin.loan_within_limit-08":     "too-small",
	"REF-fin.farmer_net_payout-07":     "decimal-places",
	"REF-fin.farmer_net_payout-08":     "too-small",
	"REF-fin.late_filing_penalty-08":   "decimal-places",
	"REF-fin.interest_step-07":         "no-negatives",
	"REF-fin.interest_step-08":         "decimal-places",
	"REF-fin.monthly_interest-08":      "decimal-places",
	"REF-edu.term_grade-07":            "too-large",
	"REF-edu.term_grade-08":            "list-too-short",
	"REF-fin.interim_payment-07":       "too-large",
	"REF-fin.interim_payment-08":       "too-small",
	"REF-fin.lending_data_ok-07":       "wrong-type",
	"REF-fin.lending_data_ok-08":       "list-too-short",
	"REF-member.voting_right-07":       "required",
	"REF-member.voting_right-08":       "not-allowed",
}

func codeOf(out Outcome) string {
	if len(out.Problems) > 0 {
		return out.Problems[0].Code
	}
	if out.Failure != nil {
		return "run:" + out.Failure.Code
	}
	return ""
}

func TestEveryRefusalIsRefusedForTheRuleItNames(t *testing.T) {
	seen := map[string]bool{}
	for key, cases := range loadReference(t) {
		found, _ := Find(key)
		callable, err := found.Prepare()
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range cases {
			if c.Expected != RefusedText {
				continue
			}
			seen[c.ID] = true
			want, listed := refusalCodes[c.ID]
			if !listed {
				t.Errorf("%s is a refusal with no code listed in the test", c.ID)
				continue
			}
			out := callable.Call(RowOf(c))
			if got := codeOf(out); got != want {
				t.Errorf("%s: refused with %q, want %q (%s)", c.ID, got, want, out.Sentence())
			}
			if len(out.Problems) > 1 {
				t.Errorf("%s: %d problems, want exactly one", c.ID, len(out.Problems))
			}
		}
	}
	for id := range refusalCodes {
		if !seen[id] {
			t.Errorf("%s is listed in the test and is not a recorded refusal", id)
		}
	}
}

func TestEveryDefinitionIsAcceptedOnSave(t *testing.T) {
	for _, candidate := range All() {
		if refused := function.CheckDefinition(candidate.Definition); len(refused) > 0 {
			t.Errorf("%s: %s", candidate.Key(), refused[0].Sentence)
		}
	}
}

func TestEveryFunctionStartsAtVersionOneWithAUniqueKey(t *testing.T) {
	keys := map[string]bool{}
	for _, candidate := range All() {
		if candidate.Version != 1 {
			t.Errorf("%s is not version 1", candidate.Key())
		}
		if keys[candidate.Key()] {
			t.Errorf("%s is in the library twice", candidate.Key())
		}
		keys[candidate.Key()] = true
	}
	if len(keys) != 19 {
		t.Errorf("the library has %d functions, want 19", len(keys))
	}
}

func TestACallGivesTheSameAnswerEveryTime(t *testing.T) {
	versions := loadReference(t)
	for key, cases := range versions {
		found, _ := Find(key)
		callable, _ := found.Prepare()
		for _, c := range cases {
			first := callable.Call(RowOf(c))
			second := callable.Call(RowOf(c))
			if first.Value.String() != second.Value.String() || first.Sentence() != second.Sentence() {
				t.Errorf("%s: two calls differ: %q and %q", c.ID, first.Value.String(), second.Value.String())
			}
		}
	}
}
