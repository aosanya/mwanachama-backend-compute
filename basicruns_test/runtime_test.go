package basicruns

import (
	"sort"
	"strconv"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/library"
)

type described struct {
	library string
	name    string
	version int
	sig     formula.Signature
}

type signatures struct {
	functions []described
}

func (s signatures) Function(library string, name string, version int) (formula.Signature, bool) {
	for _, fn := range s.functions {
		if fn.library == library && fn.name == name && fn.version == version {
			return fn.sig, true
		}
	}
	return formula.Signature{}, false
}

func (s signatures) Versions(library string, name string) []int {
	var versions []int
	for _, fn := range s.functions {
		if fn.library == library && fn.name == name {
			versions = append(versions, fn.version)
		}
	}
	sort.Ints(versions)
	return versions
}

func decimalParams(count int) []formula.Param {
	params := make([]formula.Param, count)
	for i := range params {
		params[i] = formula.Param{Name: "p", Type: formula.TypeDecimal}
	}
	return params
}

func describedFunctions() []described {
	functions := []described{
		{"", "late_days", 1, formula.Signature{Params: []formula.Param{{Name: "due", Type: formula.TypeDate}, {Name: "filed", Type: formula.TypeDate}}, Result: formula.TypeInteger, Cost: 8}},
		{"", "interest_step", 1, formula.Signature{Params: decimalParams(1), Result: formula.TypeDecimal, Cost: 5}},
	}
	for depth, name := range []string{"step_a", "step_b", "step_c", "step_d", "step_e"} {
		functions = append(functions, described{"", name, 1, formula.Signature{Params: decimalParams(1), Result: formula.TypeDecimal, Cost: 5 + depth, CallDepth: depth}})
	}
	return functions
}

type runtimeOf struct {
	signatures
	programs map[string]*formula.Program
	registry *library.Registry
}

func (r runtimeOf) Function(library string, name string, version int) (formula.Signature, bool) {
	if signature, found := r.registry.Function(library, name, version); found {
		return signature, true
	}
	return r.signatures.Function(library, name, version)
}

func (r runtimeOf) Versions(library string, name string) []int {
	if versions := r.registry.Versions(library, name); len(versions) > 0 {
		return versions
	}
	return r.signatures.Versions(library, name)
}

func programKey(library string, name string, version int) string {
	return library + "." + name + "@" + strconv.Itoa(version)
}

func (r runtimeOf) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	if _, served := r.registry.Function(library, name, version); served {
		return r.registry.Apply(library, name, version, args)
	}
	if program, found := r.programs[programKey(library, name, version)]; found {
		return program.Apply(args)
	}
	return formula.Value{}, &formula.Problem{Code: "function-not-built", Sentence: "The function " + programKey(library, name, version) + " is not built here."}
}

var agencyFunctions = []struct {
	name    string
	formula string
	inputs  []formula.Input
}{
	{"late_days", "max([0, days_between(due, filed)])", []formula.Input{{Name: "due", Type: formula.TypeDate}, {Name: "filed", Type: formula.TypeDate}}},
	{"interest_step", "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020", []formula.Input{{Name: "balance", Type: formula.TypeDecimal}}},
}

func newRuntime(t *testing.T) runtimeOf {
	t.Helper()
	runtime := runtimeOf{
		signatures: signatures{functions: describedFunctions()},
		programs:   map[string]*formula.Program{},
		registry:   newRegistry(t),
	}
	for _, agency := range agencyFunctions {
		tree, problems := formula.Parse(agency.formula)
		if len(problems) != 0 {
			t.Fatalf("%s did not parse: %+v", agency.name, problems)
		}
		program, problems := formula.Plan(tree, agency.inputs, nil, formula.DefaultOptions())
		if len(problems) != 0 {
			t.Fatalf("%s did not plan: %+v", agency.name, problems)
		}
		runtime.programs[programKey("", agency.name, 1)] = program
	}
	return runtime
}
