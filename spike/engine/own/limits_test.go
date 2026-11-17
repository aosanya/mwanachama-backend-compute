package own

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func limitNameOf(err error) string {
	var limit *LimitError
	if errors.As(err, &limit) {
		return limit.Name
	}
	return ""
}

func zeroInput() map[string]any { return map[string]any{"n": int64(0)} }

func compileProgram(t *testing.T, options Options, source string, result harness.Kind) harness.Program {
	t.Helper()
	c := harness.Case{Source: source, Result: result, Params: []harness.Param{{Name: "n", Kind: integerKind}}}
	program, err := NewWith(options).Compile(c, nil)
	if err != nil {
		t.Fatalf("%q: %v", source, err)
	}
	return program
}

func TestParseNestingGuardStopsAMillionBrackets(t *testing.T) {
	source := make([]byte, 1_000_000)
	for i := range source {
		source[i] = '('
	}
	_, err := Parse(string(source))
	if limitNameOf(err) != "depth" {
		t.Errorf("got %v", err)
	}
}

func TestParseNodeLimitStopsALongChainEarly(t *testing.T) {
	source := "1"
	for range 5000 {
		source += " + 1"
	}
	_, err := ParseWith(source, Options{MaxNodes: 100})
	if limitNameOf(err) != "node" {
		t.Errorf("got %v", err)
	}
}

func TestLiteralLimits(t *testing.T) {
	cases := []struct {
		source  string
		options Options
		name    string
	}{
		{`"abcdef"`, Options{MaxTextLength: 5}, "text length"},
		{"123456", Options{MaxDigits: 5}, "number magnitude"},
		{"0.123456", Options{MaxDecimals: 5}, "decimal places"},
		{"1 + 1 + 1", Options{MaxLength: 4}, "formula length"},
	}
	for _, c := range cases {
		if _, err := ParseWith(c.source, c.options); limitNameOf(err) != c.name {
			t.Errorf("%q: got %v", c.source, err)
		}
	}
}

func TestStepLimitCountsEveryEvaluatedNode(t *testing.T) {
	if _, err := compileProgram(t, Options{MaxSteps: 5}, "1 + 2 + 3", integerKind).Run(zeroInput()); err != nil {
		t.Errorf("five steps should fit: %v", err)
	}
	_, err := compileProgram(t, Options{MaxSteps: 4}, "1 + 2 + 3", integerKind).Run(zeroInput())
	if limitNameOf(err) != "cost" {
		t.Errorf("got %v", err)
	}
}

func TestStepLimitCountsListElementsOfAggregates(t *testing.T) {
	options := Options{MaxSteps: 20}
	if _, err := compileProgram(t, options, "sum([1, 2, 3])", integerKind).Run(zeroInput()); err != nil {
		t.Errorf("got %v", err)
	}
	options.MaxSteps = 7
	_, err := compileProgram(t, options, "sum([1, 2, 3])", integerKind).Run(zeroInput())
	if limitNameOf(err) != "cost" {
		t.Errorf("got %v", err)
	}
}

func TestContextDeadlineStopsAMeteredRun(t *testing.T) {
	options := Options{CheckEvery: 1, Functions: map[string]Function{"wait": {
		Result: integerKind,
		Call: func(meter Meter, _ []any) (any, error) {
			time.Sleep(20 * time.Millisecond)
			return int64(1), meter.Charge(1)
		},
	}}}
	program := compileProgram(t, options, "wait() + 1", integerKind).(interface {
		RunContext(context.Context, map[string]any) (any, error)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err := program.RunContext(ctx, zeroInput())
	if limitNameOf(err) != "time" || !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v", err)
	}
}

func TestInputLimits(t *testing.T) {
	options := Options{MaxTextLength: 3, MaxListLength: 2}
	program := func(source string, kind harness.Kind, param harness.Kind) harness.Program {
		c := harness.Case{Source: source, Result: kind, Params: []harness.Param{{Name: "x", Kind: param}}}
		compiled, err := NewWith(options).Compile(c, nil)
		if err != nil {
			t.Fatal(err)
		}
		return compiled
	}
	if _, err := program(`x == "a"`, boolKind, textKind).Run(map[string]any{"x": "abcd"}); limitNameOf(err) != "text length" {
		t.Errorf("got %v", err)
	}
	if _, err := program("count(x)", integerKind, harness.KindIntegerList).Run(map[string]any{"x": []int64{1, 2, 3}}); limitNameOf(err) != "list length" {
		t.Errorf("got %v", err)
	}
}

func TestPlacesArgumentIsLimited(t *testing.T) {
	options := Options{MaxDecimals: 6}
	c := harness.Case{Source: "round(1.0 / 3.0, n)", Result: decimalKind, Params: []harness.Param{{Name: "n", Kind: integerKind}}}
	program, err := NewWith(options).Compile(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Run(map[string]any{"n": int64(6)}); err != nil {
		t.Errorf("got %v", err)
	}
	if _, err := program.Run(map[string]any{"n": int64(1_000_000)}); limitNameOf(err) != "decimal places" {
		t.Errorf("got %v", err)
	}
	c.Source = "round(1.0 / 3.0, 1000000)"
	if _, err := NewWith(options).Compile(c, nil); limitNameOf(err) != "decimal places" {
		t.Errorf("got %v", err)
	}
}
