package examplecases

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type function struct {
	library   string
	name      string
	version   int
	signature formula.Signature
}

type StandIn struct {
	functions []function
	programs  map[string]*formula.Program
	natives   map[string]func([]formula.Value) (formula.Value, *formula.Problem)
}

func key(library string, name string, version int) string {
	return library + "." + name + "@" + strconv.Itoa(version)
}

func (s *StandIn) Function(library string, name string, version int) (formula.Signature, bool) {
	for _, f := range s.functions {
		if f.library == library && f.name == name && f.version == version {
			return f.signature, true
		}
	}
	return formula.Signature{}, false
}

func (s *StandIn) Versions(library string, name string) []int {
	var versions []int
	for _, f := range s.functions {
		if f.library == library && f.name == name {
			versions = append(versions, f.version)
		}
	}
	sort.Ints(versions)
	return versions
}

func (s *StandIn) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	id := key(library, name, version)
	if program, found := s.programs[id]; found {
		return program.Apply(args)
	}
	if native, found := s.natives[id]; found {
		return native(args)
	}
	return formula.Value{}, &formula.Problem{Code: "unknown-function", Sentence: "The function " + id + " is not known."}
}

func decimalParams(count int) []formula.Param {
	params := make([]formula.Param, count)
	for i := range params {
		params[i] = formula.Param{Name: "p", Type: formula.TypeDecimal}
	}
	return params
}

func NewStandIn() (*StandIn, error) {
	s := &StandIn{
		programs: map[string]*formula.Program{},
		natives:  map[string]func([]formula.Value) (formula.Value, *formula.Problem){},
	}
	s.functions = []function{
		{"", "late_days", 1, formula.Signature{Params: []formula.Param{{Name: "due", Type: formula.TypeDate}, {Name: "filed", Type: formula.TypeDate}}, Result: formula.TypeInteger, Cost: 8}},
		{"", "interest_step", 1, formula.Signature{Params: decimalParams(1), Result: formula.TypeDecimal, Cost: 5}},
		{"fin", "pmt", 1, formula.Signature{Params: paymentParams(), Result: formula.TypeDecimal, Cost: 10}},
		{"fin", "pmt", 2, formula.Signature{Params: paymentParams(), Result: formula.TypeDecimal, Cost: 10}},
	}
	for depth, name := range []string{"step_a", "step_b", "step_c", "step_d", "step_e"} {
		s.functions = append(s.functions, function{"", name, 1, formula.Signature{Params: decimalParams(1), Result: formula.TypeDecimal, Cost: 5 + depth, CallDepth: depth}})
	}
	if err := s.agencyProgram("late_days", "max([0, days_between(due, filed)])", []formula.Input{{Name: "due", Type: formula.TypeDate}, {Name: "filed", Type: formula.TypeDate}}); err != nil {
		return nil, err
	}
	if err := s.agencyProgram("interest_step", "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020", []formula.Input{{Name: "balance", Type: formula.TypeDecimal}}); err != nil {
		return nil, err
	}
	s.natives[key("fin", "pmt", 2)] = payment
	return s, nil
}

func paymentParams() []formula.Param {
	return []formula.Param{{Name: "principal", Type: formula.TypeDecimal}, {Name: "rate", Type: formula.TypeDecimal}, {Name: "months", Type: formula.TypeInteger}}
}

func (s *StandIn) agencyProgram(name string, text string, inputs []formula.Input) error {
	tree, problems := formula.Parse(text)
	if len(problems) != 0 {
		return errProblems(name, problems)
	}
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	if len(problems) != 0 {
		return errProblems(name, problems)
	}
	s.programs[key("", name, 1)] = program
	return nil
}

func payment(args []formula.Value) (formula.Value, *formula.Problem) {
	principal, _ := args[0].Decimal().ToFloat()
	rate, _ := args[1].Decimal().ToFloat()
	factor := 1.0
	for i := int64(0); i < args[2].Integer(); i++ {
		factor /= 1 + rate.Value()
	}
	amount := principal.Value() * rate.Value() / (1 - factor)
	decimal, err := number.ParseDecimal(strconv.FormatFloat(amount, 'f', 2, 64))
	if err != nil {
		return formula.Value{}, &formula.Problem{Code: "bad-function-result", Sentence: "The payment could not be written as a decimal."}
	}
	return formula.DecimalValue(decimal), nil
}

func errProblems(name string, problems []formula.Problem) error {
	return fmt.Errorf("the stand-in %s does not plan: %s", name, problems[0].Sentence)
}
