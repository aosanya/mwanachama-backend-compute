package functions

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

func decimalArg(t *testing.T, text string) formula.Value {
	t.Helper()
	parsed, err := number.ParseDecimal(text)
	if err != nil {
		t.Fatal(err)
	}
	return formula.DecimalValue(parsed)
}

func TestACalleeIsServedWithItsSignature(t *testing.T) {
	calls := NewCalls()
	signature, found := calls.Function("fin", "interest_step", 1)
	if !found {
		t.Fatal("fin.interest_step@1 is not served")
	}
	if signature.Result != formula.TypeDecimal || len(signature.Params) != 1 || signature.Params[0].Name != "balance" {
		t.Errorf("signature %+v", signature)
	}
	if signature.CallDepth != 0 {
		t.Errorf("a function with no calls has depth %d", signature.CallDepth)
	}
	if signature.Cost < 1 {
		t.Errorf("cost %d", signature.Cost)
	}
}

func TestAFunctionThatCallsAnotherHasACallDepthOfOne(t *testing.T) {
	calls := NewCalls()
	for _, name := range []string{"monthly_interest", "late_filing_penalty"} {
		signature, found := calls.Function("fin", name, 1)
		if !found {
			t.Fatalf("fin.%s@1 is not served", name)
		}
		if signature.CallDepth != 1 {
			t.Errorf("fin.%s@1 has call depth %d, want 1", name, signature.CallDepth)
		}
	}
}

func TestAVersionThatDoesNotExistIsNotServed(t *testing.T) {
	calls := NewCalls()
	if _, found := calls.Function("fin", "interest_step", 2); found {
		t.Error("fin.interest_step@2 was served")
	}
	if versions := calls.Versions("fin", "interest_step"); len(versions) != 1 || versions[0] != 1 {
		t.Errorf("versions %v", versions)
	}
	if _, failure := calls.Apply("fin", "interest_step", 2, []formula.Value{decimalArg(t, "10.00")}); failure == nil || failure.Code != "unknown-function" {
		t.Errorf("failure %+v", failure)
	}
}

func TestACalleeRefusesWhatItsInputRulesRefuse(t *testing.T) {
	calls := NewCalls()
	_, failure := calls.Apply("fin", "interest_step", 1, []formula.Value{decimalArg(t, "-1.00")})
	if failure == nil || failure.Code != "no-negatives" {
		t.Fatalf("failure %+v", failure)
	}
}

func TestACalleeGivesTheSameAnswerThroughACallAndOnItsOwn(t *testing.T) {
	calls := NewCalls()
	value, failure := calls.Apply("fin", "interest_step", 1, []formula.Value{decimalArg(t, "3217.45")})
	if failure != nil || value.String() != "0.015" {
		t.Fatalf("%v %+v", value, failure)
	}
}

func TestACallingFunctionPassesItsCalleesCostToTheCallersMeter(t *testing.T) {
	calls := NewCalls()
	found, _ := Find("fin.monthly_interest@1")
	callable, err := found.PrepareIn(calls)
	if err != nil {
		t.Fatal(err)
	}
	alone, _ := Find("fin.interest_step@1")
	inner, err := alone.PrepareIn(calls)
	if err != nil {
		t.Fatal(err)
	}
	if callable.program.Cost() <= inner.program.Cost() {
		t.Errorf("monthly_interest costs %d and its callee costs %d", callable.program.Cost(), inner.program.Cost())
	}
}
