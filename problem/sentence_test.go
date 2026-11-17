package problem

import "testing"

type expected struct {
	name     string
	problem  Problem
	code     string
	input    string
	item     int
	given    bool
	value    string
	hasValue bool
	sentence string
}

func inputProblem(name string, p Problem, code, input string, item int, value string, sentence string) expected {
	return expected{name: name, problem: p, code: code, input: input, item: item, given: true, value: value, hasValue: true, sentence: sentence}
}

func TestSentencesOfEveryCodeWordForWord(t *testing.T) {
	whole := "a whole number with at most 18 digits"
	decimal := "a decimal number with at most 30 digits, such as 12.50"
	truth := "true or false"
	list := "a list of decimal numbers"
	cases := []expected{
		{name: "ic-a02", problem: Required("payment"), code: "required", input: "payment", given: false, sentence: "`payment` was not given. It is required."},
		{name: "ic-a04", problem: Required("balance"), code: "required", input: "balance", given: false, sentence: "`balance` was not given. It is required."},
		{name: "ic-a08", problem: Required("note"), code: "required", input: "note", given: false, sentence: "`note` was not given. It is required."},

		inputProblem("ic-b02", ListTooShort("contributions", true, List(nil), 0, 1), "list-too-short", "contributions", 0, "[]",
			"`contributions` has 0 items. It needs at least 1 item."),
		{name: "ic-b03", problem: ListTooShort("contributions", false, List(nil), 0, 1), code: "list-too-short", input: "contributions", given: false,
			sentence: "`contributions` was not given, so it has 0 items. It needs at least 1 item."},
		inputProblem("ic-b07", ListTooLong("scores", List(repeated("70", 13)), 13, 12), "list-too-long", "scores", 0,
			"[70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70]", "`scores` has 13 items. It can have at most 12 items."),
		inputProblem("ic-b08", ListTooLong("scores", List(repeated("70", 21)), 21, 12), "list-too-long", "scores", 0,
			"[70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, ...]", "`scores` has 21 items. It can have at most 12 items."),
		inputProblem("ic-b09", WrongType(Subject{Input: "scores", Item: 13}, PlainText("abc"), whole), "wrong-type", "scores", 13, "abc",
			"Item 13 of `scores` is abc. It must be a whole number with at most 18 digits."),
		inputProblem("ic-b12", WrongType(Subject{Input: "scores"}, PlainText("70"), "a list of whole numbers"), "wrong-type", "scores", 0, "70",
			"`scores` is 70. It must be a list of whole numbers."),

		inputProblem("ic-c02", NoNegatives(Subject{Input: "payment"}, PlainText("-5.00")), "no-negatives", "payment", 0, "-5.00",
			"`payment` is -5.00. It cannot be negative."),
		inputProblem("ic-c04b", DecimalPlaces(Subject{Input: "payment"}, PlainText("-5.125"), 2), "decimal-places", "payment", 0, "-5.125",
			"`payment` is -5.125. It can have at most 2 decimal places."),
		inputProblem("ic-c05", NoNegatives(Subject{Input: "contributions", Item: 2}, PlainText("-50.00")), "no-negatives", "contributions", 2, "-50.00",
			"Item 2 of `contributions` is -50.00. It cannot be negative."),
		inputProblem("ic-c06", TooSmall(Subject{Input: "units"}, PlainText("-3"), Range{Min: ptr("1"), Max: ptr("100")}), "too-small", "units", 0, "-3",
			"`units` is -3. It must be from 1 to 100."),
		inputProblem("ic-c07-open", TooSmall(Subject{Input: "price_per_kilo"}, PlainText("-0.5"), Range{Min: ptr("0.000"), MinExclusive: true}), "too-small", "price_per_kilo", 0, "-0.5",
			"`price_per_kilo` is -0.5. It must be more than 0.000."),

		inputProblem("ic-d01", TooSmall(Subject{Input: "members"}, PlainText("0"), Range{Min: ptr("1"), Max: ptr("500")}), "too-small", "members", 0, "0",
			"`members` is 0. It must be from 1 to 500."),
		inputProblem("ic-d02", TooLarge(Subject{Input: "members"}, PlainText("501"), Range{Min: ptr("1"), Max: ptr("500")}), "too-large", "members", 0, "501",
			"`members` is 501. It must be from 1 to 500."),
		inputProblem("ic-d03", TooSmall(Subject{Input: "requested"}, PlainText("0.00"), Range{Min: ptr("0.00"), MinExclusive: true}), "too-small", "requested", 0, "0.00",
			"`requested` is 0.00. It must be more than 0.00."),
		inputProblem("ic-d05", TooSmall(Subject{Input: "share_pct"}, PlainText("-0.01"), Range{Min: ptr("0.00"), Max: ptr("100.00"), MaxExclusive: true}), "too-small", "share_pct", 0, "-0.01",
			"`share_pct` is -0.01. It must be 0.00 or more."),
		inputProblem("ic-d06", TooLarge(Subject{Input: "share_pct"}, PlainText("100.00"), Range{Min: ptr("0.00"), Max: ptr("100.00"), MaxExclusive: true}), "too-large", "share_pct", 0, "100.00",
			"`share_pct` is 100.00. It must be less than 100.00."),
		inputProblem("ic-d07", TooLarge(Subject{Input: "bonus"}, PlainText("500.01"), Range{Max: ptr("500.00")}), "too-large", "bonus", 0, "500.01",
			"`bonus` is 500.01. It must be 500.00 or less."),
		inputProblem("ic-d09", TooLarge(Subject{Input: "monthly_rate"}, PlainText("0.06"), Range{Min: ptr("0.000"), Max: ptr("0.050")}), "too-large", "monthly_rate", 0, "0.06",
			"`monthly_rate` is 0.06. It must be from 0.000 to 0.050."),
		inputProblem("ic-d10", TooSmall(Subject{Input: "scores", Item: 1}, PlainText("-3"), Range{Min: ptr("0"), Max: ptr("100")}), "too-small", "scores", 1, "-3",
			"Item 1 of `scores` is -3. It must be from 0 to 100."),
		inputProblem("ic-d11", TooLarge(Subject{Input: "amounts", Item: 2}, PlainText("250.00"), Range{Min: ptr("0.00"), Max: ptr("100.00")}), "too-large", "amounts", 2, "250.00",
			"Item 2 of `amounts` is 250.00. It must be from 0.00 to 100.00."),

		inputProblem("ic-e01", DecimalPlaces(Subject{Input: "payment"}, PlainText("5.125"), 2), "decimal-places", "payment", 0, "5.125",
			"`payment` is 5.125. It can have at most 2 decimal places."),
		inputProblem("ic-e02", DecimalPlaces(Subject{Input: "price_per_kilo"}, PlainText("0.4370"), 3), "decimal-places", "price_per_kilo", 0, "0.4370",
			"`price_per_kilo` is 0.4370. It can have at most 3 decimal places."),
		inputProblem("ic-e03", DecimalPlaces(Subject{Input: "bags"}, PlainText("2.5"), 0), "decimal-places", "bags", 0, "2.5",
			"`bags` is 2.5. It cannot have decimal places."),
		inputProblem("ic-e05", DecimalPlaces(Subject{Input: "weight"}, PlainText("2.25"), 1), "decimal-places", "weight", 0, "2.25",
			"`weight` is 2.25. It can have at most 1 decimal place."),
		inputProblem("ic-e06", DecimalPlaces(Subject{Input: "balances", Item: 2}, PlainText("380.505"), 2), "decimal-places", "balances", 2, "380.505",
			"Item 2 of `balances` is 380.505. It can have at most 2 decimal places."),

		inputProblem("ic-f01", NotAllowed(Subject{Input: "status"}, QuotedText("paused"), []string{"active", "lapsed", "suspended"}), "not-allowed", "status", 0, "paused",
			"`status` is \"paused\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		inputProblem("ic-f03", NotAllowed(Subject{Input: "status"}, QuotedText("active "), []string{"active", "lapsed", "suspended"}), "not-allowed", "status", 0, "active ",
			"`status` is \"active \". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		inputProblem("ic-j02", NotAllowed(Subject{Input: "status"}, QuotedText(""), []string{"active", "lapsed", "suspended"}), "not-allowed", "status", 0, "",
			"`status` is \"\". It must be one of \"active\", \"lapsed\", \"suspended\"."),
		inputProblem("ic-j03", NotAllowed(Subject{Input: "status"}, QuotedText("   "), []string{"active", "lapsed", "suspended"}), "not-allowed", "status", 0, "   ",
			"`status` is \"   \". It must be one of \"active\", \"lapsed\", \"suspended\"."),

		inputProblem("ic-g03", WrongType(Subject{Input: "scores", Item: 3}, PlainText("abc"), whole), "wrong-type", "scores", 3, "abc",
			"Item 3 of `scores` is abc. It must be a whole number with at most 18 digits."),
		inputProblem("ic-h12", WrongType(Subject{Input: "balances", Item: 2}, PlainText("abc"), decimal), "wrong-type", "balances", 2, "abc",
			"Item 2 of `balances` is abc. It must be a decimal number with at most 30 digits, such as 12.50."),
		inputProblem("ic-i08", WrongType(Subject{Input: "kilos"}, PlainText("abc"), whole), "wrong-type", "kilos", 0, "abc",
			"`kilos` is abc. It must be a whole number with at most 18 digits."),
		inputProblem("ic-i09", WrongType(Subject{Input: "kilos"}, PlainText("-1.5"), whole), "wrong-type", "kilos", 0, "-1.5",
			"`kilos` is -1.5. It must be a whole number with at most 18 digits."),
		inputProblem("ic-j05", WrongType(Subject{Input: "payment"}, PlainText(""), decimal), "wrong-type", "payment", 0, "",
			"`payment` is . It must be a decimal number with at most 30 digits, such as 12.50."),
		inputProblem("ic-j06", WrongType(Subject{Input: "payment"}, PlainText(" "), decimal), "wrong-type", "payment", 0, " ",
			"`payment` is  . It must be a decimal number with at most 30 digits, such as 12.50."),
		inputProblem("ic-j07", WrongType(Subject{Input: "active"}, PlainText(""), truth), "wrong-type", "active", 0, "",
			"`active` is . It must be true or false."),
		inputProblem("ic-j09", WrongType(Subject{Input: "contributions", Item: 1}, PlainText(""), decimal), "wrong-type", "contributions", 1, "",
			"Item 1 of `contributions` is . It must be a decimal number with at most 30 digits, such as 12.50."),
		inputProblem("ic-m01", WrongType(Subject{Input: "payment"}, List([]string{"45.25"}), decimal), "wrong-type", "payment", 0, "[45.25]",
			"`payment` is [45.25]. It must be a decimal number with at most 30 digits, such as 12.50."),
		inputProblem("ic-m02", WrongType(Subject{Input: "contributions"}, PlainText("200.00"), list), "wrong-type", "contributions", 0, "200.00",
			"`contributions` is 200.00. It must be a list of decimal numbers."),
		inputProblem("ic-m03", WrongType(Subject{Input: "contributions", Item: 2}, PlainText("1,200.50"), decimal), "wrong-type", "contributions", 2, "1,200.50",
			"Item 2 of `contributions` is 1,200.50. It must be a decimal number with at most 30 digits, such as 12.50."),

		inputProblem("ic-i06", UnknownInput("memo", PlainText("hello"), []string{"principal", "monthly_rate", "months"}), "unknown-input", "memo", 0, "hello",
			"`memo` is not an input of this function. Its inputs are: principal, monthly_rate, months."),
		inputProblem("ic-i05", UnknownInput("alpha", PlainText("2"), []string{"balance", "payment"}), "unknown-input", "alpha", 0, "2",
			"`alpha` is not an input of this function. Its inputs are: balance, payment."),
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkInputProblem(t, c)
		})
	}
}

func checkInputProblem(t *testing.T, c expected) {
	t.Helper()
	p := c.problem
	if p.Code != c.code {
		t.Errorf("Code = %q, want %q", p.Code, c.code)
	}
	if p.Input == nil || *p.Input != c.input {
		t.Errorf("Input = %v, want %q", deref(p.Input), c.input)
	}
	if c.item == 0 && p.Item != nil {
		t.Errorf("Item = %d, want null", *p.Item)
	}
	if c.item != 0 && (p.Item == nil || *p.Item != c.item) {
		t.Errorf("Item = %v, want %d", p.Item, c.item)
	}
	if p.Given == nil || *p.Given != c.given {
		t.Errorf("Given = %v, want %v", p.Given, c.given)
	}
	if c.hasValue && (p.Value == nil || *p.Value != c.value) {
		t.Errorf("Value = %v, want %q", deref(p.Value), c.value)
	}
	if !c.hasValue && p.Value != nil {
		t.Errorf("Value = %q, want null", *p.Value)
	}
	if p.Check != nil {
		t.Errorf("Check = %q, want null", *p.Check)
	}
	if p.Sentence != c.sentence {
		t.Errorf("Sentence =\n%q\nwant\n%q", p.Sentence, c.sentence)
	}
}

func TestSentencesOfChecks(t *testing.T) {
	cases := []struct {
		name     string
		problem  Problem
		code     string
		check    string
		sentence string
	}{
		{
			name:     "ic-h01",
			problem:  CheckFailed("payment > 0", "The payment must be more than zero", []NamedValue{{Name: "payment", Value: PlainText("0.00")}}),
			code:     "check-failed",
			check:    "payment > 0",
			sentence: "The payment must be more than zero. `payment` is 0.00.",
		},
		{
			name: "ic-h02",
			problem: CheckFailed("payment <= balance + 1000.00", "The payment cannot be more than 1000.00 above the balance",
				[]NamedValue{{Name: "payment", Value: PlainText("2000.00")}, {Name: "balance", Value: PlainText("100.00")}}),
			code:     "check-failed",
			check:    "payment <= balance + 1000.00",
			sentence: "The payment cannot be more than 1000.00 above the balance. `payment` is 2000.00 and `balance` is 100.00.",
		},
		{
			name: "ic-h04",
			problem: CheckFailed("principal * monthly_rate * months <= principal", "The interest cannot be more than the amount borrowed",
				[]NamedValue{{Name: "principal", Value: PlainText("50.00")}, {Name: "monthly_rate", Value: PlainText("0.050")}, {Name: "months", Value: PlainText("60")}}),
			code:     "check-failed",
			check:    "principal * monthly_rate * months <= principal",
			sentence: "The interest cannot be more than the amount borrowed. `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60.",
		},
		{
			name: "ic-h07",
			problem: CheckFailed("count(balances) >= 6", "At least 6 balances are needed",
				[]NamedValue{{Name: "balances", Value: List([]string{"420.00", "380.50"})}}),
			code:     "check-failed",
			check:    "count(balances) >= 6",
			sentence: "At least 6 balances are needed. `balances` is [420.00, 380.50].",
		},
		{
			name: "text-value-in-a-check",
			problem: CheckFailed("status == \"active\"", "The member must be active",
				[]NamedValue{{Name: "status", Value: QuotedText("lapsed")}}),
			code:     "check-failed",
			check:    "status == \"active\"",
			sentence: "The member must be active. `status` is \"lapsed\".",
		},
		{
			name: "ic-h09",
			problem: CheckError("total / people >= 1.00", "Each person must get at least 1.00",
				"The part `total / people` divides by zero."),
			code:     "check-error",
			check:    "total / people >= 1.00",
			sentence: "The check \"Each person must get at least 1.00\" could not be worked out: The part `total / people` divides by zero.",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := c.problem
			if p.Code != c.code {
				t.Errorf("Code = %q, want %q", p.Code, c.code)
			}
			if p.Rule != "check" {
				t.Errorf("Rule = %q, want check", p.Rule)
			}
			if p.Input != nil || p.Item != nil || p.Given != nil || p.Value != nil {
				t.Errorf("input, item, given and value must be null: %v %v %v %v", p.Input, p.Item, p.Given, p.Value)
			}
			if p.Check == nil || *p.Check != c.check {
				t.Errorf("Check = %v, want %q", deref(p.Check), c.check)
			}
			if p.Sentence != c.sentence {
				t.Errorf("Sentence =\n%q\nwant\n%q", p.Sentence, c.sentence)
			}
		})
	}
}

func TestCheckWithNoValuesEndsAfterTheMessage(t *testing.T) {
	p := CheckFailed("1 > 2", "One must be more than two", nil)
	if p.Sentence != "One must be more than two." {
		t.Fatalf("Sentence = %q", p.Sentence)
	}
}

func TestRangeTexts(t *testing.T) {
	cases := []struct {
		name  string
		r     Range
		small string
		large string
	}{
		{"both bounds", Range{Min: ptr("1"), Max: ptr("500")}, "from 1 to 500", "from 1 to 500"},
		{"min only", Range{Min: ptr("0.00")}, "0.00 or more", ""},
		{"min only exclusive", Range{Min: ptr("0.00"), MinExclusive: true}, "more than 0.00", ""},
		{"max only", Range{Max: ptr("500.00")}, "", "500.00 or less"},
		{"max only exclusive", Range{Max: ptr("100.00"), MaxExclusive: true}, "", "less than 100.00"},
		{"both with min exclusive", Range{Min: ptr("0"), Max: ptr("10"), MinExclusive: true}, "more than 0", "10 or less"},
		{"both with max exclusive", Range{Min: ptr("0.00"), Max: ptr("100.00"), MaxExclusive: true}, "0.00 or more", "less than 100.00"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.small != "" && c.r.TooSmallText() != c.small {
				t.Errorf("TooSmallText() = %q, want %q", c.r.TooSmallText(), c.small)
			}
			if c.large != "" && c.r.TooLargeText() != c.large {
				t.Errorf("TooLargeText() = %q, want %q", c.r.TooLargeText(), c.large)
			}
		})
	}
}

func TestNamedValuesAreJoinedWithCommasAndTheWordAnd(t *testing.T) {
	cases := []struct {
		count int
		want  string
	}{
		{1, "`a` is 1."},
		{2, "`a` is 1 and `b` is 2."},
		{3, "`a` is 1, `b` is 2 and `c` is 3."},
		{4, "`a` is 1, `b` is 2, `c` is 3 and `d` is 4."},
	}
	names := []NamedValue{
		{Name: "a", Value: PlainText("1")},
		{Name: "b", Value: PlainText("2")},
		{Name: "c", Value: PlainText("3")},
		{Name: "d", Value: PlainText("4")},
	}
	for _, c := range cases {
		p := CheckFailed("x", "M", names[:c.count])
		if p.Sentence != "M. "+c.want {
			t.Errorf("%d values: Sentence = %q", c.count, p.Sentence)
		}
	}
}

func TestListCountReadsItemOrItems(t *testing.T) {
	cases := []struct {
		count, items int
		want         string
	}{
		{1, 2, "`x` has 1 item. It needs at least 2 items."},
		{0, 1, "`x` has 0 items. It needs at least 1 item."},
	}
	for _, c := range cases {
		p := ListTooShort("x", true, List(repeated("1", c.count)), c.count, c.items)
		if p.Sentence != c.want {
			t.Errorf("Sentence = %q, want %q", p.Sentence, c.want)
		}
	}
	long := ListTooLong("x", List(repeated("1", 2)), 2, 1)
	if long.Sentence != "`x` has 2 items. It can have at most 1 item." {
		t.Errorf("Sentence = %q", long.Sentence)
	}
}

func TestLongTextIsCutInTheFieldAndInTheSentence(t *testing.T) {
	long := ""
	for i := 0; i < 201; i++ {
		long += "n"
	}
	cut := long[:200] + "..."
	p := NotAllowed(Subject{Input: "note"}, QuotedText(long), []string{"a"})
	if p.Value == nil || *p.Value != cut {
		t.Fatalf("Value = %v", deref(p.Value))
	}
	want := "`note` is \"" + cut + "\". It must be one of \"a\"."
	if p.Sentence != want {
		t.Fatalf("Sentence = %q", p.Sentence)
	}
}

func deref(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}
