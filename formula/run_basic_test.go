package formula

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func TestBasicOperators(t *testing.T) {
	runRows(t, nil, []runRow{
		{"add integers", "a + b", "a=int:7; b=int:5", "int:12"},
		{"add decimals", "a + b", "a=dec:120.50; b=dec:45.2", "dec:165.70"},
		{"add integer and decimal", "a + b", "a=int:3; b=dec:0.50", "dec:3.50"},
		{"subtract integers", "a - b", "a=int:5; b=int:8", "int:-3"},
		{"multiply integers", "a * b", "a=int:6; b=int:7", "int:42"},
		{"multiply decimals", "a * b", "a=dec:1.25; b=dec:0.40", "dec:0.5000"},
		{"multiply decimal and integer", "a * b", "a=dec:0.437; b=int:853", "dec:372.761"},
		{"divide integer by decimal", "a / b", "a=int:7; b=dec:2.0", "dec:3.50000000000000000000"},
		{"divide decimal by integer", "a / b", "a=dec:2.00; b=int:3", "dec:0.66666666666666666667"},
		{"divide negative decimal by integer", "a / b", "a=dec:-2.00; b=int:3", "dec:-0.66666666666666666667"},
		{"divide floats", "a / b", "a=float:1.5; b=float:0.5", "float:3"},
		{"divide integers refused", "a / b", "a=int:7; b=int:2", "!type-mismatch"},
		{"remainder negative divisor", "a % b", "a=int:7; b=int:-3", "int:1"},
		{"remainder by zero", "a % b", "a=int:5; b=int:0", "?divide-by-zero"},
		{"remainder on a decimal refused", "a % b", "a=dec:5.5; b=int:2", "!type-mismatch"},
		{"divide decimal by zero", "a / b", "a=dec:5.5; b=int:0", "?divide-by-zero"},
		{"divide float by zero", "a / b", "a=float:5.5; b=float:0.0", "?divide-by-zero"},
		{"negate by subtraction", "0 - x", "x=dec:12.50", "dec:-12.50"},
		{"power sign refused", "a ^ 2", "a=int:3", "!bad-syntax"},
		{"power function refused", "pow(a, 2)", "a=int:3", "!unknown-function"},
		{"integer and float add", "a + b", "a=int:2; b=float:1.5", "float:3.5"},
		{"float and decimal refused", "a - b", "a=float:45.25; b=dec:120.50", "!float-decimal-mix"},
	})
}

func TestBasicPrecedenceAndGrouping(t *testing.T) {
	runRows(t, nil, []runRow{
		{"multiply first", "a + b * c", "a=int:2; b=int:3; c=int:4", "int:14"},
		{"brackets first", "(a + b) * c", "a=int:2; b=int:3; c=int:4", "int:20"},
		{"subtract left", "a - b - c", "a=int:10; b=int:3; c=int:2", "int:5"},
		{"subtract bracket right", "a - (b - c)", "a=int:10; b=int:3; c=int:2", "int:9"},
		{"divide then multiply", "a / b * c", "a=dec:12.00; b=int:4; c=int:3", "dec:9.00000000000000000000"},
		{"divide bracket right", "a / (b * c)", "a=dec:12.00; b=int:4; c=int:3", "dec:1.00000000000000000000"},
		{"remainder then multiply", "a % b * c", "a=int:17; b=int:5; c=int:3", "int:6"},
		{"arithmetic before equal", "a + b == c * d", "a=int:2; b=int:3; c=int:1; d=int:5", "bool:true"},
		{"comparisons do not chain", "a < b < c", "a=int:1; b=int:2; c=int:3", "!type-mismatch"},
		{"not first", "!a && b", "a=bool:true; b=bool:false", "bool:false"},
		{"not over a bracket", "!(a && b)", "a=bool:true; b=bool:false", "bool:true"},
		{"and before or", "a || b && c", "a=bool:true; b=bool:false; c=bool:false", "bool:true"},
		{"or bracket first", "(a || b) && c", "a=bool:true; b=bool:false; c=bool:false", "bool:false"},
		{"or before conditional", "a || b ? 1 : 2", "a=bool:false; b=bool:true", "int:1"},
		{"conditional else to the right", "a ? 1 : b ? 2 : 3", "a=bool:false; b=bool:true", "int:2"},
		{"conditional then is whole", "a ? b ? 1 : 2 : 3", "a=bool:true; b=bool:false", "int:2"},
	})
}

func TestBasicComparisonAndBooleanLogic(t *testing.T) {
	runRows(t, nil, []runRow{
		{"less true", "a < b", "a=int:2; b=int:3", "bool:true"},
		{"less equal operands", "a < b", "a=int:3; b=int:3", "bool:false"},
		{"less or equal equal operands", "a <= b", "a=int:3; b=int:3", "bool:true"},
		{"less or equal false", "a <= b", "a=int:4; b=int:3", "bool:false"},
		{"greater true", "a > b", "a=int:4; b=int:3", "bool:true"},
		{"greater equal operands", "a > b", "a=int:3; b=int:3", "bool:false"},
		{"greater or equal equal operands", "a >= b", "a=int:3; b=int:3", "bool:true"},
		{"greater or equal false", "a >= b", "a=int:2; b=int:3", "bool:false"},
		{"equal integer and decimal", "a == b", "a=int:3; b=dec:3.00", "bool:true"},
		{"equal texts different case", "a == b", "a=text:\"Active\"; b=text:\"active\"", "bool:false"},
		{"equal texts", "a == b", "a=text:\"active\"; b=text:\"active\"", "bool:true"},
		{"not equal true", "a != b", "a=dec:1.00; b=dec:1.01", "bool:true"},
		{"not equal false", "a != b", "a=dec:12.5; b=dec:12.50", "bool:false"},
		{"not equal texts", "a != b", "a=text:\"a\"; b=text:\"b\"", "bool:true"},
		{"decimal against float refused", "a < b", "a=dec:1.00; b=float:2.5", "!float-decimal-mix"},
		{"float equal refused", "a == b", "a=float:1.5; b=float:1.5", "!type-mismatch"},
		{"integer equal float refused", "a == b", "a=int:1; b=float:1.0", "!type-mismatch"},
		{"float ordering", "a < b", "a=float:1.5; b=float:2.5", "bool:true"},
		{"integer ordering against float", "a < b", "a=int:2; b=float:2.5", "bool:true"},
		{"decimal ordering against integer", "a < b", "a=dec:2.50; b=int:3", "bool:true"},
		{"text ordering refused", "a < b", "a=text:\"ant\"; b=text:\"bee\"", "!type-mismatch"},
		{"text against number refused", "a == 1", "a=text:\"1\"", "!type-mismatch"},
		{"and true true", "a && b", "a=bool:true; b=bool:true", "bool:true"},
		{"and true false", "a && b", "a=bool:true; b=bool:false", "bool:false"},
		{"and false true", "a && b", "a=bool:false; b=bool:true", "bool:false"},
		{"and false false", "a && b", "a=bool:false; b=bool:false", "bool:false"},
		{"or true true", "a || b", "a=bool:true; b=bool:true", "bool:true"},
		{"or true false", "a || b", "a=bool:true; b=bool:false", "bool:true"},
		{"or false true", "a || b", "a=bool:false; b=bool:true", "bool:true"},
		{"or false false", "a || b", "a=bool:false; b=bool:false", "bool:false"},
		{"not true", "!a", "a=bool:true", "bool:false"},
		{"not false", "!a", "a=bool:false", "bool:true"},
		{"not twice", "!!a", "a=bool:true", "bool:true"},
		{"and on numbers refused", "a && b", "a=int:1; b=int:0", "!type-mismatch"},
		{"and does not skip its right side", "a && total / n > 0.50", "a=bool:false; total=dec:1.00; n=int:0", "?divide-by-zero"},
		{"or does not skip its right side", "a || total / n > 0.50", "a=bool:true; total=dec:1.00; n=int:0", "?divide-by-zero"},
	})
}

func TestBasicConditional(t *testing.T) {
	ladder := "x < 10 ? \"low\" : x < 20 ? \"mid\" : \"high\""
	runRows(t, nil, []runRow{
		{"nested low", ladder, "x=int:5", "text:low"},
		{"nested boundary", ladder, "x=int:10", "text:mid"},
		{"nested mid", ladder, "x=int:15", "text:mid"},
		{"nested high", ladder, "x=int:25", "text:high"},
		{"text result", "status == \"active\" ? \"open\" : \"closed\"", "status=text:\"active\"", "text:open"},
		{"condition that is not boolean", "a ? 1 : 2", "a=int:1", "!condition-not-boolean"},
		{"answers of different kinds", "a ? 1 : \"x\"", "a=bool:true", "!branches-differ"},
		{"loan covers with the whole number zero", "payment >= balance ? 0 : balance - payment", "balance=dec:120.50; payment=dec:130.00", "dec:0"},
		{"loan equal payment", "payment >= balance ? 0 : balance - payment", "balance=dec:120.50; payment=dec:120.50", "dec:0"},
		{"loan zero and zero", "payment >= balance ? 0 : balance - payment", "balance=dec:0.00; payment=dec:0.00", "dec:0"},
		{"loan nothing paid", "payment >= balance ? 0 : balance - payment", "balance=dec:80.00; payment=dec:0.00", "dec:80.00"},
		{"loan one cent short", "payment >= balance ? 0 : balance - payment", "balance=dec:100.00; payment=dec:99.99", "dec:0.01"},
		{"loan negative balance covered", "payment >= balance ? 0 : balance - payment", "balance=dec:-10.00; payment=dec:5.00", "dec:0"},
		{"loan negative both lower", "payment >= balance ? 0 : balance - payment", "balance=dec:-10.00; payment=dec:-20.00", "dec:10.00"},
	})
}

func TestBasicTypesAndConversion(t *testing.T) {
	runRows(t, nil, []runRow{
		{"float in float out", "a * b", "a=float:1.5; b=int:2", "float:3"},
		{"float stays finite", "a * b", "a=float:1e200; b=float:1e200", "?number-too-big"},
		{"to float", "to_float(x)", "x=dec:1.50", "float:1.5"},
		{"to decimal", "to_decimal(x, 2)", "x=float:0.37", "dec:0.37"},
		{"to decimal three places", "to_decimal(x, 3)", "x=float:0.37", "dec:0.370"},
		{"to decimal rounds at the place", "to_decimal(x, 2)", "x=float:1.005", "dec:1.01"},
		{"to decimal no places", "to_decimal(x, 0)", "x=float:2.5", "dec:3"},
		{"convert round trip", "to_decimal(to_float(x), 2)", "x=dec:1.50", "dec:1.50"},
		{"to float of an integer refused", "to_float(x)", "x=int:5", "!wrong-argument-type"},
		{"to decimal of a decimal refused", "to_decimal(x, 2)", "x=dec:1.50", "!wrong-argument-type"},
		{"conversion makes a mix work", "to_float(a) - b", "a=dec:1.50; b=float:0.25", "float:1.25"},
	})
}

func TestBasicRound(t *testing.T) {
	runRows(t, nil, []runRow{
		{"dropped digits", "round(x, 2)", "x=dec:372.761", "dec:372.76"},
		{"exactly half", "round(x, 2)", "x=dec:2.345", "dec:2.35"},
		{"just under the half", "round(x, 2)", "x=dec:2.344", "dec:2.34"},
		{"negative half", "round(x, 2)", "x=dec:-2.345", "dec:-2.35"},
		{"no places", "round(x, 0)", "x=dec:2.5", "dec:3"},
		{"no places below the half", "round(x, 0)", "x=dec:0.4", "dec:0"},
		{"pads places", "round(x, 4)", "x=dec:1.5", "dec:1.5000"},
		{"float stays a float", "round(x, 3)", "x=float:1.3629", "float:1.363"},
		{"integer refused", "round(x, 2)", "x=int:5", "!wrong-argument-type"},
		{"one argument refused", "round(x)", "x=dec:5.00", "!wrong-argument-count"},
		{"a version refused", "round@1(x, 2)", "x=dec:5.00", "!version-not-allowed"},
	})
}

func TestBasicAggregates(t *testing.T) {
	runRows(t, nil, []runRow{
		{"sum many", "sum(xs)", "xs=list<dec>:[200.00, 250.00]", "dec:450.00"},
		{"sum one", "sum(xs)", "xs=list<dec>:[75.50]", "dec:75.50"},
		{"sum empty", "sum(xs)", "xs=list<dec>:[]", "dec:0"},
		{"sum integers", "sum(xs)", "xs=list<int>:[3, 4, 5]", "int:12"},
		{"sum empty integers", "sum(xs)", "xs=list<int>:[]", "int:0"},
		{"sum a written list", "sum([1, 2.50])", "", "dec:3.50"},
		{"count many", "count(xs)", "xs=list<dec>:[420.00, 380.50]", "int:2"},
		{"count one", "count(xs)", "xs=list<dec>:[9.99]", "int:1"},
		{"count empty", "count(xs)", "xs=list<dec>:[]", "int:0"},
		{"avg many", "avg(xs)", "xs=list<dec>:[10.00, 20.00, 25.00]", "dec:18.33333333333333333333"},
		{"avg exact", "avg(xs)", "xs=list<dec>:[1.00, 2.00]", "dec:1.50000000000000000000"},
		{"avg one", "avg(xs)", "xs=list<dec>:[5.00]", "dec:5.00000000000000000000"},
		{"avg integers", "avg(xs)", "xs=list<int>:[72, 85]", "float:78.5"},
		{"avg one integer", "avg(xs)", "xs=list<int>:[70]", "float:70"},
		{"avg empty", "avg(xs)", "xs=list<dec>:[]", "?empty-list"},
		{"avg empty integers", "avg(xs)", "xs=list<int>:[]", "?empty-list"},
		{"min many", "min(xs)", "xs=list<dec>:[10.00, 4.50, 7.25, 4.50]", "dec:4.50"},
		{"min one", "min(xs)", "xs=list<dec>:[4.10]", "dec:4.10"},
		{"min integers", "min(xs)", "xs=list<int>:[5, 3, 9]", "int:3"},
		{"min tie shows the first item", "min(xs)", "xs=list<dec>:[4.50, 4.5]", "dec:4.50"},
		{"min empty", "min(xs)", "xs=list<dec>:[]", "?empty-list"},
		{"max many", "max(xs)", "xs=list<dec>:[10.00, 4.50, 7.25]", "dec:10.00"},
		{"max one", "max(xs)", "xs=list<dec>:[4.10]", "dec:4.10"},
		{"max tie shows the first item", "max(xs)", "xs=list<dec>:[4.5, 4.50]", "dec:4.5"},
		{"max empty", "max(xs)", "xs=list<dec>:[]", "?empty-list"},
		{"max of mixed written items", "max([1, 2.50, 2])", "", "dec:2.50"},
		{"a list as the whole formula", "[1, 2]", "", "list<int>:[1, 2]"},
	})
}

func TestBasicDates(t *testing.T) {
	runRows(t, nil, []runRow{
		{"days between", "days_between(a, b)", "a=date:2026-03-31; b=date:2026-05-15", "int:45"},
		{"days between backwards", "days_between(a, b)", "a=date:2026-05-15; b=date:2026-03-31", "int:-45"},
		{"days between the same day", "days_between(a, b)", "a=date:2026-03-31; b=date:2026-03-31", "int:0"},
		{"days between across a leap day", "days_between(a, b)", "a=date:2028-02-28; b=date:2028-03-01", "int:2"},
		{"months between", "months_between(a, b)", "a=date:2026-01-20; b=date:2026-09-15", "int:7"},
		{"months between at a month end", "months_between(a, b)", "a=date:2026-01-31; b=date:2026-02-28", "int:1"},
		{"months between a day short", "months_between(a, b)", "a=date:2026-01-31; b=date:2026-02-27", "int:0"},
		{"months between backwards", "months_between(a, b)", "a=date:2026-09-15; b=date:2026-01-20", "int:-7"},
		{"years between a day short", "years_between(a, b)", "a=date:2023-09-10; b=date:2026-09-09", "int:2"},
		{"years between on the day", "years_between(a, b)", "a=date:2023-09-10; b=date:2026-09-10", "int:3"},
		{"a number where a date is needed", "days_between(a, 5)", "a=date:2026-03-31", "!wrong-argument-type"},
	})
}

func TestTheResultKeepsTheWrittenDigits(t *testing.T) {
	cases := map[string]string{
		"0.010":   "dec:0.010",
		"12.50":   "dec:12.50",
		"0":       "int:0",
		"0.00":    "dec:0.00",
		"007":     "!bad-syntax",
		"1 + 0.0": "dec:1.0",
	}
	for formula, expect := range cases {
		t.Run(formula, func(t *testing.T) {
			if expect == "!bad-syntax" {
				if _, problems := Parse(formula); len(problems) == 0 {
					t.Fatal("Parse accepted a leading zero")
				}
				return
			}
			runRows(t, nil, []runRow{{formula, formula, "", expect}})
		})
	}
}

func TestSettingsChooseTheRoundingAndTheDivisionDigits(t *testing.T) {
	tree := mustParse(t, "round(x, 2)")
	options := DefaultOptions()
	options.Numbers = number.Settings{Places: 2, Mode: number.RoundHalfDown, DivisionDigits: 5}
	program, problems := Plan(tree, inputs("x", "decimal"), nil, options)
	if len(problems) != 0 {
		t.Fatalf("Plan gave %+v", problems)
	}
	halfDown, failure := program.Run(map[string]Value{"x": runValue(t, "dec:2.345")})
	if failure != nil || runShown(halfDown) != "dec:2.34" {
		t.Fatalf("half down gave %s, %v", runShown(halfDown), failure)
	}
	division := mustParse(t, "a / b")
	divider, problems := Plan(division, inputs("a", "decimal", "b", "integer"), nil, options)
	if len(problems) != 0 {
		t.Fatalf("Plan gave %+v", problems)
	}
	five, failure := divider.Run(map[string]Value{"a": runValue(t, "dec:2.00"), "b": runValue(t, "int:3")})
	if failure != nil || runShown(five) != "dec:0.66667" {
		t.Fatalf("five digits gave %s, %v", runShown(five), failure)
	}
}

func TestARunNeedsEveryInputOfTheRightType(t *testing.T) {
	program, problems := Plan(mustParse(t, "a + b"), inputs("a", "integer", "b", "integer"), nil, DefaultOptions())
	if len(problems) != 0 {
		t.Fatalf("Plan gave %+v", problems)
	}
	if _, failure := program.Run(map[string]Value{"a": runValue(t, "int:1")}); failure == nil || failure.Code != "bad-input" {
		t.Fatalf("a missing input gave %+v", failure)
	}
	if _, failure := program.Run(map[string]Value{"a": runValue(t, "int:1"), "b": runValue(t, "dec:2.0")}); failure == nil || failure.Code != "bad-input" {
		t.Fatalf("a wrong type gave %+v", failure)
	}
}

func TestTheSameInputGivesTheSameResultEveryRun(t *testing.T) {
	program, problems := Plan(mustParse(t, "round(sum(xs) / count(xs), 2)"), inputs("xs", "list of decimal"), nil, DefaultOptions())
	if len(problems) != 0 {
		t.Fatalf("Plan gave %+v", problems)
	}
	values := map[string]Value{"xs": runValue(t, "list<dec>:[10.00, 20.00, 25.00]")}
	first, _ := program.Run(values)
	for i := 0; i < 100; i++ {
		again, failure := program.Run(values)
		if failure != nil || runShown(again) != runShown(first) {
			t.Fatalf("run %d gave %s, %v; the first gave %s", i, runShown(again), failure, runShown(first))
		}
	}
}
