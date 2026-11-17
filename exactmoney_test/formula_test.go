package exactmoney_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type declared struct {
	name  string
	value formula.Value
}

func optionsFor(mode number.Rounding, places int) formula.Options {
	options := formula.DefaultOptions()
	options.Numbers = number.Settings{Places: places, Mode: mode, DivisionDigits: 20}
	return options
}

func planFormula(t *testing.T, text string, mode number.Rounding, in ...declared) (*formula.Program, []formula.Problem) {
	t.Helper()
	tree, problems := formula.Parse(text)
	if len(problems) != 0 {
		t.Fatalf("Parse(%q): %+v", text, problems)
	}
	inputs := make([]formula.Input, len(in))
	for i, item := range in {
		inputs[i] = formula.Input{Name: item.name, Type: item.value.Type()}
	}
	return formula.Plan(tree, inputs, nil, optionsFor(mode, 2))
}

func runFormula(t *testing.T, text string, mode number.Rounding, in ...declared) formula.Value {
	t.Helper()
	program, problems := planFormula(t, text, mode, in...)
	if len(problems) != 0 {
		t.Fatalf("Plan(%q): %+v", text, problems)
	}
	values := map[string]formula.Value{}
	for _, item := range in {
		values[item.name] = item.value
	}
	value, problem := program.Run(values)
	if problem != nil {
		t.Fatalf("Run(%q): %+v", text, *problem)
	}
	return value
}

func decimalInput(t *testing.T, name string, text string) declared {
	t.Helper()
	return declared{name, formula.DecimalValue(dec(t, text))}
}

func floatInput(t *testing.T, name string, v float64) declared {
	t.Helper()
	f, err := number.NewFloat(v)
	if err != nil {
		t.Fatal(err)
	}
	return declared{name, formula.FloatValue(f)}
}

func decimalList(t *testing.T, name string, items []number.Decimal) declared {
	t.Helper()
	return declared{name, formula.DecimalListValue(items)}
}

func TestFormula_EM01_AndEM03_ExactSumsInEveryOrder(t *testing.T) {
	if got := runFormula(t, "0.1 + 0.2 == 0.3", number.RoundHalfUp); !got.Boolean() {
		t.Fatalf("0.1 + 0.2 == 0.3 is %s", got)
	}
	orders := map[string][]string{
		"listed":  {"19.99", "0.01", "5.50", "0.10", "0.20"},
		"reverse": {"0.20", "0.10", "5.50", "0.01", "19.99"},
		"mixed":   {"5.50", "0.20", "19.99", "0.10", "0.01"},
	}
	for name, order := range orders {
		t.Run(name, func(t *testing.T) {
			items := make([]number.Decimal, len(order))
			for i, text := range order {
				items[i] = dec(t, text)
			}
			got := runFormula(t, "sum(amounts)", number.RoundHalfUp, decimalList(t, "amounts", items))
			wantEqual(t, got.Decimal(), "25.80")
		})
	}
}

func TestFormula_EM02_AndEM04_LongSumsAreOrderFree(t *testing.T) {
	cents := make([]number.Decimal, 1000)
	for i := range cents {
		cents[i] = units(int64((37*(i+1))%1000)+1, 2)
	}
	ones := make([]number.Decimal, 1000)
	for i := range ones {
		ones[i] = units(1, 2)
	}
	got := runFormula(t, "sum(amounts)", number.RoundHalfUp, decimalList(t, "amounts", ones))
	wantText(t, got.Decimal(), "10.00")
	flipped := make([]number.Decimal, len(cents))
	for i, c := range cents {
		flipped[len(cents)-1-i] = c
	}
	for name, list := range map[string][]number.Decimal{"forward": cents, "reverse": flipped} {
		t.Run(name, func(t *testing.T) {
			got := runFormula(t, "sum(amounts)", number.RoundHalfUp, decimalList(t, "amounts", list))
			wantText(t, got.Decimal(), "5005.00")
		})
	}
}

func TestFormula_RoundingTables_EveryModeThroughTheRoundCall(t *testing.T) {
	for _, row := range halfRows {
		t.Run(row.id+negativeTag(row), func(t *testing.T) {
			places := 2
			if row.id <= "EM-19" {
				places = 0
			}
			scale := places + 1
			input := number.NewDecimal(row.unscaled, scale)
			for i, m := range modes {
				got := runFormula(t, fmt.Sprintf("round(x, %d)", places), m.mode, declared{"x", formula.DecimalValue(input)})
				want := number.NewDecimal(row.expected[i], places)
				if got.Decimal().String() != want.String() {
					t.Errorf("%s: round(%s, %d) under %s is %s, want %s", row.id, input, places, m.name, got.Decimal(), want)
				}
			}
		})
	}
}

func TestFormula_EM27_ZeroHasNoSign(t *testing.T) {
	got := runFormula(t, "round(x, 2)", number.RoundHalfUp, decimalInput(t, "x", "-0.004"))
	wantText(t, got.Decimal(), "0.00")
}

func TestFormula_InterestFigures(t *testing.T) {
	t.Run("EM-52", func(t *testing.T) {
		got := runFormula(t, "1000.00 + 1000.00 * 0.05 * 3", number.RoundHalfUp)
		wantEqual(t, got.Decimal(), "1150.00")
	})
	t.Run("EM-54 unrounded", func(t *testing.T) {
		got := runFormula(t, "1000.00 * 1.05 * 1.05 * 1.05 == 1157.625", number.RoundHalfUp)
		if !got.Boolean() {
			t.Fatal("the unrounded figure is not exactly 1157.625")
		}
	})
	want := map[string]string{
		"half-up": "1157.63", "half-down": "1157.62", "half-even": "1157.62", "half-+inf": "1157.63",
		"up": "1157.63", "down": "1157.62", "ceiling": "1157.63", "floor": "1157.62",
	}
	for _, m := range modes {
		t.Run("EM-54 "+m.name, func(t *testing.T) {
			got := runFormula(t, "round(1000.00 * 1.05 * 1.05 * 1.05, 2)", m.mode)
			wantText(t, got.Decimal(), want[m.name])
		})
	}
	t.Run("EM-56 each period against once", func(t *testing.T) {
		each := runFormula(t, "round(round(round(120.00 * 1.02, 2) * 1.02, 2) * 1.02, 2)", number.RoundHalfUp)
		wantText(t, each.Decimal(), "127.35")
		once := runFormula(t, "round(120.00 * 1.02 * 1.02 * 1.02, 2)", number.RoundHalfUp)
		wantText(t, once.Decimal(), "127.34")
	})
	t.Run("EM-55 once at the end", func(t *testing.T) {
		got := runFormula(t, "round(500.00 * 1.015 * 1.015 * 1.015 * 1.015, 2)", number.RoundHalfUp)
		wantText(t, got.Decimal(), "530.68")
	})
}

func TestFormula_EM60ToEM62_MixingADecimalAndAFloatIsRefusedBeforeARun(t *testing.T) {
	cases := map[string]string{
		"EM-60": "amount + rate",
		"EM-61": "amount * rate",
		"EM-62": "0.3 == rate",
	}
	for id, text := range cases {
		t.Run(id, func(t *testing.T) {
			program, problems := planFormula(t, text, number.RoundHalfUp, decimalInput(t, "amount", "12.50"), floatInput(t, "rate", 0.1))
			if program != nil || len(problems) == 0 {
				t.Fatalf("%q was accepted", text)
			}
			if !strings.Contains(problems[0].Sentence, "decimal") || !strings.Contains(problems[0].Sentence, "float") {
				t.Errorf("the sentence %q does not name both kinds", problems[0].Sentence)
			}
		})
	}
}

func TestFormula_EM63AndEM64_AnExplicitConversionStatesItsPlaces(t *testing.T) {
	got := runFormula(t, "amount + to_decimal(rate, 2)", number.RoundHalfUp, decimalInput(t, "amount", "12.50"), floatInput(t, "rate", 0.1))
	wantEqual(t, got.Decimal(), "12.60")
	got = runFormula(t, "to_decimal(rate, 2)", number.RoundHalfUp, floatInput(t, "rate", 0.1+0.2))
	wantText(t, got.Decimal(), "0.30")
}

func TestFormula_EM66_AConversionWithNoPlacesIsRefused(t *testing.T) {
	program, problems := planFormula(t, "to_decimal(rate)", number.RoundHalfUp, floatInput(t, "rate", 0.1))
	if program != nil || len(problems) == 0 {
		t.Fatal("a conversion with no places was accepted")
	}
}

func TestFormula_EM97AndEM98_ADivisionKeepsTheAgencysDigits(t *testing.T) {
	one, three, two := decimalInput(t, "one", "1.00"), decimalInput(t, "three", "3.00"), decimalInput(t, "two", "2.00")
	third := runFormula(t, "one / three", number.RoundHalfUp, one, three)
	wantText(t, third.Decimal(), "0.33333333333333333333")
	got := runFormula(t, "round(one / three, 2)", number.RoundHalfUp, one, three)
	wantText(t, got.Decimal(), "0.33")
	got = runFormula(t, "one / three * three", number.RoundHalfUp, one, three)
	wantEqual(t, got.Decimal(), "0.99999999999999999999")
	want := map[string]string{"half-up": "0.67", "up": "0.67", "ceiling": "0.67", "down": "0.66", "floor": "0.66"}
	for name, text := range want {
		t.Run(name, func(t *testing.T) {
			mode := modeNamed(t, name)
			got := runFormula(t, "round(two / three, 2)", mode, two, three)
			wantText(t, got.Decimal(), text)
		})
	}
}

func TestFormula_EM90ToEM95_Extremes(t *testing.T) {
	wantText(t, runFormula(t, "0.00 + 0.00", number.RoundHalfUp).Decimal(), "0.00")
	wantText(t, runFormula(t, "0.01 + 0.01", number.RoundHalfUp).Decimal(), "0.02")
	wantText(t, runFormula(t, "round(0.01 * 0.5, 2)", number.RoundHalfUp).Decimal(), "0.01")
	wantText(t, runFormula(t, "round(0.01 * 0.5, 2)", number.RoundHalfDown).Decimal(), "0.00")
	wantText(t, runFormula(t, "round(0.01 * 0.5, 2)", number.RoundHalfEven).Decimal(), "0.00")
	wantText(t, runFormula(t, "0.00 - 10.00 + 4.00", number.RoundHalfUp).Decimal(), "-6.00")
}
