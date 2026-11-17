package problem

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInputProblemFieldsAreWrittenInOrderWithNulls(t *testing.T) {
	p := NoNegatives(Subject{Input: "payment"}, PlainText("-5.00"))
	got, err := p.JSON()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"input\":\"payment\",\"item\":null,\"given\":true,\"value\":\"-5.00\",\"rule\":\"no-negatives\"," +
		"\"sentence\":\"`payment` is -5.00. It cannot be negative.\",\"code\":\"no-negatives\",\"check\":null}"
	if string(got) != want {
		t.Fatalf("JSON() =\n%s\nwant\n%s", got, want)
	}
}

func TestCheckProblemFieldsAreWrittenInOrderWithNulls(t *testing.T) {
	p := CheckFailed("payment > 0", "The payment must be more than zero", []NamedValue{{Name: "payment", Value: PlainText("0.00")}})
	got, err := p.JSON()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"input\":null,\"item\":null,\"given\":null,\"value\":null,\"rule\":\"check\"," +
		"\"sentence\":\"The payment must be more than zero. `payment` is 0.00.\",\"code\":\"check-failed\",\"check\":\"payment > 0\"}"
	if string(got) != want {
		t.Fatalf("JSON() =\n%s\nwant\n%s", got, want)
	}
}

func TestRequiredProblemHasGivenFalseAndValueNull(t *testing.T) {
	p := Required("credit")
	got, err := p.JSON()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"input\":\"credit\",\"item\":null,\"given\":false,\"value\":null,\"rule\":\"required\"," +
		"\"sentence\":\"`credit` was not given. It is required.\",\"code\":\"required\",\"check\":null}"
	if string(got) != want {
		t.Fatalf("JSON() =\n%s\nwant\n%s", got, want)
	}
}

func TestItemProblemCarriesThePositionAndTheTextOfThatItem(t *testing.T) {
	p := NoNegatives(Subject{Input: "contributions", Item: 2}, PlainText("-50.00"))
	got, err := p.JSON()
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"input\":\"contributions\",\"item\":2,\"given\":true,\"value\":\"-50.00\",\"rule\":\"no-negatives\"," +
		"\"sentence\":\"Item 2 of `contributions` is -50.00. It cannot be negative.\",\"code\":\"no-negatives\",\"check\":null}"
	if string(got) != want {
		t.Fatalf("JSON() =\n%s\nwant\n%s", got, want)
	}
}

func TestListOfProblemsIsAnArrayInTheGivenOrder(t *testing.T) {
	list := []Problem{Required("balance"), Required("payment")}
	got, err := ListJSON(list)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 || decoded[0]["input"] != "balance" || decoded[1]["input"] != "payment" {
		t.Fatalf("ListJSON() = %s", got)
	}
}

func TestEmptyListOfProblemsIsAnEmptyArray(t *testing.T) {
	for _, list := range [][]Problem{nil, {}} {
		got, err := ListJSON(list)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "[]" {
			t.Fatalf("ListJSON() = %s", got)
		}
	}
}

func TestJSONKeepsGreaterThanAndAmpersandAsWritten(t *testing.T) {
	p := CheckFailed("a > 1 && b < 2", "A must be more than 1", []NamedValue{{Name: "a", Value: PlainText("0")}})
	got, err := p.JSON()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"check":"a > 1 && b < 2"`) {
		t.Fatalf("JSON() escaped the check: %s", got)
	}
}

func TestEveryCodeIsLowercaseLettersDigitsAndHyphens(t *testing.T) {
	for _, code := range AllCodes() {
		if !ValidCode(code) {
			t.Errorf("code %q is not valid", code)
		}
	}
}

func TestValidCode(t *testing.T) {
	valid := []string{"required", "check-failed", "a", "no-negatives", strings.Repeat("a", 40)}
	invalid := []string{"", "Required", "check_failed", "has space", strings.Repeat("a", 41), "é"}
	for _, code := range valid {
		if !ValidCode(code) {
			t.Errorf("ValidCode(%q) = false", code)
		}
	}
	for _, code := range invalid {
		if ValidCode(code) {
			t.Errorf("ValidCode(%q) = true", code)
		}
	}
}

func TestTheTwelveCodesOfValidate(t *testing.T) {
	want := []string{
		"required", "wrong-type", "too-small", "too-large", "no-negatives", "decimal-places",
		"not-allowed", "list-too-short", "list-too-long", "check-failed", "check-error", "unknown-input",
	}
	got := AllCodes()
	if len(got) != len(want) {
		t.Fatalf("AllCodes() = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AllCodes()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestEveryProblemUsesTheRuleOfItsCode(t *testing.T) {
	s := Subject{Input: "x"}
	rules := map[string]Problem{
		"required":       Required("x"),
		"type":           WrongType(s, PlainText("abc"), "true or false"),
		"range":          TooSmall(s, PlainText("0"), Range{Min: ptr("1")}),
		"max-range":      TooLarge(s, PlainText("9"), Range{Max: ptr("5")}),
		"no-negatives":   NoNegatives(s, PlainText("-1")),
		"decimal-places": DecimalPlaces(s, PlainText("1.5"), 0),
		"allowed-values": NotAllowed(s, QuotedText("z"), []string{"a"}),
		"min-items":      ListTooShort("x", true, List(nil), 0, 1),
		"max-items":      ListTooLong("x", List([]string{"1", "2"}), 2, 1),
		"check":          CheckFailed("x > 0", "X must be more than zero", nil),
		"known-input":    UnknownInput("memo", PlainText("hi"), []string{"x"}),
	}
	for name, p := range rules {
		want := name
		if name == "max-range" {
			want = "range"
		}
		if p.Rule != want {
			t.Errorf("%s: Rule = %q, want %q", name, p.Rule, want)
		}
		if !ValidCode(p.Code) {
			t.Errorf("%s: Code = %q is not valid", name, p.Code)
		}
	}
	if got := CheckError("x / y > 1", "X over Y", "The part `x / y` divides by zero.").Rule; got != "check" {
		t.Errorf("check-error Rule = %q", got)
	}
}

func ptr(s string) *string {
	return &s
}
