package formula

import (
	"sort"
	"testing"
)

type fakeFunction struct {
	library string
	name    string
	version int
	sig     Signature
}

type fakeLibrary struct {
	functions []fakeFunction
}

func (f fakeLibrary) Function(library string, name string, version int) (Signature, bool) {
	for _, fn := range f.functions {
		if fn.library == library && fn.name == name && fn.version == version {
			return fn.sig, true
		}
	}
	return Signature{}, false
}

func (f fakeLibrary) Versions(library string, name string) []int {
	var versions []int
	for _, fn := range f.functions {
		if fn.library == library && fn.name == name {
			versions = append(versions, fn.version)
		}
	}
	sort.Ints(versions)
	return versions
}

func decimalParams(count int) []Param {
	params := make([]Param, count)
	for i := range params {
		params[i] = Param{Name: "p", Type: TypeDecimal}
	}
	return params
}

func testLibrary() fakeLibrary {
	functions := []fakeFunction{
		{"", "late_days", 1, Signature{Params: []Param{{"due", TypeDate}, {"filed", TypeDate}}, Result: TypeInteger, Cost: 8}},
		{"", "interest_step", 1, Signature{Params: decimalParams(1), Result: TypeDecimal, Cost: 5}},
		{"fin", "pmt", 1, Signature{Params: []Param{{"principal", TypeDecimal}, {"rate", TypeDecimal}, {"months", TypeInteger}}, Result: TypeDecimal, Cost: 10}},
		{"fin", "pmt", 2, Signature{Params: []Param{{"principal", TypeDecimal}, {"rate", TypeDecimal}, {"months", TypeInteger}}, Result: TypeDecimal, Cost: 10}},
	}
	names := []string{"step_a", "step_b", "step_c", "step_d", "step_e"}
	for depth, name := range names {
		functions = append(functions, fakeFunction{"", name, 1, Signature{Params: decimalParams(1), Result: TypeDecimal, Cost: 5 + depth, CallDepth: depth}})
	}
	return fakeLibrary{functions: functions}
}

func mustParse(t *testing.T, text string) Tree {
	t.Helper()
	tree, problems := Parse(text)
	if len(problems) != 0 {
		t.Fatalf("Parse(%q) gave problems %+v", text, problems)
	}
	return tree
}

func inputs(pairs ...string) []Input {
	var list []Input
	for i := 0; i+1 < len(pairs); i += 2 {
		list = append(list, Input{Name: pairs[i], Type: Type(pairs[i+1])})
	}
	return list
}
