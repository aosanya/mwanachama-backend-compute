package library

import (
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

var pmtSignature = formula.Signature{
	Params: []formula.Param{
		{Name: "principal", Type: formula.TypeDecimal},
		{Name: "rate", Type: formula.TypeDecimal},
		{Name: "months", Type: formula.TypeInteger},
	},
	Result: formula.TypeDecimal,
	Cost:   20,
}

func exactPayment(args []formula.Value) float64 {
	principal, _ := strconv.ParseFloat(args[0].Decimal().String(), 64)
	rate, _ := strconv.ParseFloat(args[1].Decimal().String(), 64)
	months := float64(args[2].Integer())
	return principal * rate / (1 - math.Pow(1+rate, -months))
}

func decimalOf(t *testing.T, text string) formula.Value {
	t.Helper()
	d, err := number.ParseDecimal(text)
	if err != nil {
		t.Fatalf("decimal %q: %v", text, err)
	}
	return formula.DecimalValue(d)
}

func pmtRoundedDown(t *testing.T) func([]formula.Value) (formula.Value, *formula.Problem) {
	return func(args []formula.Value) (formula.Value, *formula.Problem) {
		return decimalOf(t, strconv.Itoa(int(math.Floor(exactPayment(args))))), nil
	}
}

func pmtRoundedToCents(t *testing.T) func([]formula.Value) (formula.Value, *formula.Problem) {
	return func(args []formula.Value) (formula.Value, *formula.Problem) {
		return decimalOf(t, strconv.FormatFloat(math.Round(exactPayment(args)*100)/100, 'f', 2, 64)), nil
	}
}

var pmtNote = Note{
	Changed: "The payment is rounded to the nearest 0.01. It was rounded down to a whole number.",
	Matters: "Any agency that shows or collects a loan payment.",
	Example: Example{Input: "12,000 at 12% a year over 24 months", Old: "564", New: "564.88"},
}

func pmtVersion(t *testing.T, version int) Release {
	release := Release{Library: "fin", Name: "pmt", Version: version, Signature: pmtSignature}
	if version == 1 {
		release.Apply = pmtRoundedDown(t)
		return release
	}
	release.Apply = pmtRoundedToCents(t)
	release.Note = pmtNote
	return release
}

func finRegistry(t *testing.T) *Registry {
	t.Helper()
	registry := NewRegistry()
	for _, version := range []int{1, 2} {
		if err := registry.Publish(pmtVersion(t, version)); err != nil {
			t.Fatalf("publish fin.pmt@%d: %v", version, err)
		}
	}
	return registry
}

var loanInputs = []formula.Input{
	{Name: "principal", Type: formula.TypeDecimal},
	{Name: "rate", Type: formula.TypeDecimal},
	{Name: "months", Type: formula.TypeInteger},
}

func loanValues(t *testing.T) map[string]formula.Value {
	return map[string]formula.Value{
		"principal": decimalOf(t, "12000.00"),
		"rate":      decimalOf(t, "0.01"),
		"months":    formula.IntegerValue(24),
	}
}

func parsed(t *testing.T, text string) formula.Tree {
	t.Helper()
	tree, problems := formula.Parse(text)
	if len(problems) > 0 {
		t.Fatalf("parse %q: %v", text, problems)
	}
	return tree
}

func planned(t *testing.T, registry *Registry, text string) (*formula.Program, []formula.Problem) {
	t.Helper()
	return formula.Plan(parsed(t, text), loanInputs, registry, formula.DefaultOptions())
}

func runLoan(t *testing.T, registry *Registry, text string) string {
	t.Helper()
	program, problems := planned(t, registry, text)
	if len(problems) > 0 {
		t.Fatalf("plan %q: %v", text, problems)
	}
	value, problem := program.Run(loanValues(t))
	if problem != nil {
		t.Fatalf("run %q: %v", text, problem)
	}
	return value.String()
}

type stopTable struct {
	mu    sync.Mutex
	stops map[Ref]Stop
}

func newStopTable() *stopTable { return &stopTable{stops: map[Ref]Stop{}} }

func (s *stopTable) put(ref Ref, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stops[ref] = Stop{Reason: reason}
}

func (s *stopTable) Stopped(library string, name string, version int) (Stop, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	stop, found := s.stops[Ref{Library: library, Name: name, Version: version}]
	return stop, found
}
