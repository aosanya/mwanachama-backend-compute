package measure

import (
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

func dec(text string) numeric.Decimal {
	return numeric.MustParse(text)
}

func decimals(texts ...string) []numeric.Decimal {
	values := make([]numeric.Decimal, len(texts))
	for i, text := range texts {
		values[i] = dec(text)
	}
	return values
}

func adHoc(source string, result harness.Kind, params ...harness.Param) harness.Case {
	return harness.Case{Number: 900, Title: "ad hoc", Source: source, Params: params, Result: result}
}

func param(name string, kind harness.Kind) harness.Param {
	return harness.Param{Name: name, Kind: kind}
}

func evaluate(t *testing.T, e harness.Engine, c harness.Case, inputs map[string]any) any {
	t.Helper()
	program, err := e.Compile(c, harness.Library{})
	if err != nil {
		t.Fatalf("%s: compile %q: %v", e.Name(), c.Source, err)
	}
	got, err := program.Run(inputs)
	if err != nil {
		t.Fatalf("%s: run %q: %v", e.Name(), c.Source, err)
	}
	return got
}

func decimalResult(t *testing.T, e harness.Engine, c harness.Case, inputs map[string]any) numeric.Decimal {
	t.Helper()
	got := evaluate(t, e, c, inputs)
	d, ok := got.(numeric.Decimal)
	if !ok {
		t.Fatalf("%s: %q returned %s, want a decimal", e.Name(), c.Source, harness.Describe(got))
	}
	return d
}

func TestFloatBaselineIsInexact(t *testing.T) {
	a, b := 0.1, 0.2
	sum := a + b
	if sum == 0.3 {
		t.Fatalf("float64 0.1 + 0.2 equals 0.3; the baseline premise is gone")
	}
	t.Logf("float64 0.1 + 0.2 = %v, which is not 0.3", sum)
}

func TestTenthsAddExactly(t *testing.T) {
	c := adHoc(`a + b`, harness.KindDecimal, param("a", harness.KindDecimal), param("b", harness.KindDecimal))
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			got := decimalResult(t, e, c, map[string]any{"a": dec("0.1"), "b": dec("0.2")})
			if got.String() != "0.3" {
				t.Errorf("0.1 + 0.2 = %s, want 0.3", got)
			}
		})
	}
}

func TestRoundingHalfAwayFromZero(t *testing.T) {
	cases := []struct {
		x      string
		places int
		want   string
	}{
		{"2.675", 2, "2.68"},
		{"-2.675", 2, "-2.68"},
		{"1.005", 2, "1.01"},
		{"-1.005", 2, "-1.01"},
		{"0.125", 2, "0.13"},
		{"-0.125", 2, "-0.13"},
		{"2.5", 0, "3"},
		{"-2.5", 0, "-3"},
		{"3.5", 0, "4"},
		{"0.4999", 0, "0"},
		{"1.5", 3, "1.500"},
		{"7", 2, "7.00"},
		{"2.99999999999999999999", 2, "3.00"},
	}
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			for _, tc := range cases {
				source := "round(x, " + string(rune('0'+tc.places)) + ")"
				c := adHoc(source, harness.KindDecimal, param("x", harness.KindDecimal))
				got := decimalResult(t, e, c, map[string]any{"x": dec(tc.x)})
				if got.String() != tc.want {
					t.Errorf("%s with x=%s = %s, want %s", source, tc.x, got, tc.want)
				}
				if got.DecimalPlaces() != tc.places {
					t.Errorf("%s with x=%s has %d places, want %d", source, tc.x, got.DecimalPlaces(), tc.places)
				}
			}
		})
	}
}

func TestDivisionThenRound(t *testing.T) {
	rounded := adHoc(`round(10.00 / 3, 2)`, harness.KindDecimal)
	plain := adHoc(`10.00 / 3`, harness.KindDecimal)
	byParams := adHoc(`round(a / b, 2)`, harness.KindDecimal, param("a", harness.KindDecimal), param("b", harness.KindInteger))
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			if got := decimalResult(t, e, rounded, nil); got.String() != "3.33" || got.DecimalPlaces() != 2 {
				t.Errorf("round(10.00 / 3, 2) = %s with %d places, want 3.33", got, got.DecimalPlaces())
			}
			if got := decimalResult(t, e, byParams, map[string]any{"a": dec("10.00"), "b": int64(3)}); got.String() != "3.33" {
				t.Errorf("round(a / b, 2) = %s, want 3.33", got)
			}
			got := decimalResult(t, e, plain, nil)
			t.Logf("%s: 10.00 / 3 = %s (%d places)", e.Name(), got, got.DecimalPlaces())
			if got.String() != "3.333333333333333333" || got.DecimalPlaces() != 18 {
				t.Errorf("10.00 / 3 = %s with %d places, want 3.333333333333333333 with 18", got, got.DecimalPlaces())
			}
		})
	}
}

func TestDecimalAndFloatNeverMix(t *testing.T) {
	mixed := adHoc(`a + f`, harness.KindDecimal, param("a", harness.KindDecimal), param("f", harness.KindFloat))
	mixedCompare := adHoc(`a > f`, harness.KindBoolean, param("a", harness.KindDecimal), param("f", harness.KindFloat))
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			for _, c := range []harness.Case{mixed, mixedCompare} {
				_, err := e.Compile(c, harness.Library{})
				if err == nil {
					t.Errorf("%q compiled; a decimal and a float must not mix", c.Source)
					continue
				}
				t.Logf("%s refuses %q: %s", e.Name(), c.Source, strings.ReplaceAll(err.Error(), "\n", " | "))
			}
		})
	}
}

func TestExplicitConversionsWork(t *testing.T) {
	toDecimal := adHoc(`a + to_decimal(f, 2)`, harness.KindDecimal, param("a", harness.KindDecimal), param("f", harness.KindFloat))
	toFloat := adHoc(`to_float(a) + f`, harness.KindFloat, param("a", harness.KindDecimal), param("f", harness.KindFloat))
	inputs := map[string]any{"a": dec("1.50"), "f": 0.25}
	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			if got := decimalResult(t, e, toDecimal, inputs); got.String() != "1.75" {
				t.Errorf("a + to_decimal(f, 2) = %s, want 1.75", got)
			}
			got := evaluate(t, e, toFloat, inputs)
			f, ok := got.(float64)
			if !ok || math.Abs(f-1.75) > 1e-12 {
				t.Errorf("to_float(a) + f = %s, want 1.75 (float)", harness.Describe(got))
			}
		})
	}
}

func exactProduct(left, right string) *big.Rat {
	a, _ := new(big.Rat).SetString(left)
	b, _ := new(big.Rat).SetString(right)
	return a.Mul(a, b)
}

func sameValue(t *testing.T, got numeric.Decimal, want *big.Rat) {
	t.Helper()
	have, ok := new(big.Rat).SetString(got.String())
	if !ok {
		t.Fatalf("result %q is not a number", got)
	}
	if have.Cmp(want) != 0 {
		t.Errorf("result %s differs from the math/big value %s", got, want.FloatString(24))
	}
}

func TestFastPathOverflowStaysExact(t *testing.T) {
	product := adHoc(`a * b`, harness.KindDecimal, param("a", harness.KindDecimal), param("b", harness.KindDecimal))
	productBack := adHoc(`round(a * b / b, 2)`, harness.KindDecimal, param("a", harness.KindDecimal), param("b", harness.KindDecimal))
	plusOne := adHoc(`x + 1`, harness.KindDecimal, param("x", harness.KindDecimal))
	negate := adHoc(`0 - x`, harness.KindDecimal, param("x", harness.KindDecimal))
	square := adHoc(`x * x`, harness.KindDecimal, param("x", harness.KindDecimal))
	total := adHoc(`sum(xs)`, harness.KindDecimal, param("xs", harness.KindDecimalList))
	totalBackDown := adHoc(`sum(xs) - sum(ys)`, harness.KindDecimal, param("xs", harness.KindDecimalList), param("ys", harness.KindDecimalList))
	mean := adHoc(`sum(xs) / 20`, harness.KindDecimal, param("xs", harness.KindDecimalList))

	twenty := make([]string, 20)
	almost := make([]string, 20)
	for i := range twenty {
		twenty[i] = "4000000000000000000"
		almost[i] = "3999999999999999999"
	}

	for _, e := range engines() {
		t.Run(e.Name(), func(t *testing.T) {
			t.Run("product leaves int64", func(t *testing.T) {
				got := decimalResult(t, e, product, map[string]any{"a": dec("9223372036854775.00"), "b": dec("1000.00")})
				sameValue(t, got, exactProduct("9223372036854775.00", "1000.00"))
				if got.DecimalPlaces() != 4 {
					t.Errorf("product has %d places, want 4", got.DecimalPlaces())
				}
				if !got.IsBig() {
					t.Errorf("product %s should be on the big path", got)
				}
			})
			t.Run("product and back down", func(t *testing.T) {
				got := decimalResult(t, e, productBack, map[string]any{"a": dec("9223372036854775.00"), "b": dec("1000.00")})
				if got.String() != "9223372036854775.00" {
					t.Errorf("round(a * b / b, 2) = %s, want 9223372036854775.00", got)
				}
				if got.IsBig() {
					t.Errorf("result %s fits int64 again and should be on the fast path", got)
				}
			})
			t.Run("int64 maximum plus one", func(t *testing.T) {
				got := decimalResult(t, e, plusOne, map[string]any{"x": dec("9223372036854775807")})
				if got.String() != "9223372036854775808" || !got.IsBig() {
					t.Errorf("x + 1 = %s (big=%v), want 9223372036854775808 on the big path", got, got.IsBig())
				}
			})
			t.Run("negating the int64 minimum", func(t *testing.T) {
				got := decimalResult(t, e, negate, map[string]any{"x": dec("-9223372036854775808")})
				if got.String() != "9223372036854775808" || !got.IsBig() {
					t.Errorf("0 - x = %s (big=%v), want 9223372036854775808 on the big path", got, got.IsBig())
				}
			})
			t.Run("square of a large integer", func(t *testing.T) {
				got := decimalResult(t, e, square, map[string]any{"x": dec("9223372036854775807")})
				sameValue(t, got, exactProduct("9223372036854775807", "9223372036854775807"))
			})
			t.Run("sum of twenty large values", func(t *testing.T) {
				got := decimalResult(t, e, total, map[string]any{"xs": decimals(twenty...)})
				if got.String() != "80000000000000000000" {
					t.Errorf("sum = %s, want 80000000000000000000", got)
				}
				if !got.IsBig() {
					t.Errorf("sum %s should be on the big path", got)
				}
			})
			t.Run("sum chain back to a small result", func(t *testing.T) {
				got := decimalResult(t, e, totalBackDown, map[string]any{"xs": decimals(twenty...), "ys": decimals(almost...)})
				if got.String() != "20" {
					t.Errorf("sum(xs) - sum(ys) = %s, want 20", got)
				}
				if got.IsBig() {
					t.Errorf("result %s fits int64 again and should be on the fast path", got)
				}
			})
			t.Run("mean of twenty large values", func(t *testing.T) {
				got := decimalResult(t, e, mean, map[string]any{"xs": decimals(twenty...)})
				sameValue(t, got, exactProduct("4000000000000000000", "1"))
			})
		})
	}
}
