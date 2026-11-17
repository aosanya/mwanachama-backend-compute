package own

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type foreignProgram struct{}

func (foreignProgram) Run(map[string]any) (any, error) { return nil, nil }

func publishedHalf(t *testing.T) harness.Library {
	t.Helper()
	c := harness.Case{Source: "d / 2", Result: decimalKind, Params: []harness.Param{{Name: "d", Kind: decimalKind}}}
	program, err := New().Compile(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	return harness.Library{"half@1": program}
}

func TestLibraryFunctionsAreCalledByPublishedName(t *testing.T) {
	program, err := compileWith("round(half@1(d) + half@1(half@1(d)), 3)", decimalKind, publishedHalf(t))
	if err != nil {
		t.Fatal(err)
	}
	got, err := program.Run(sampleInputs())
	if err != nil || !harness.Equal(numeric.MustParse("1.875"), got) {
		t.Errorf("got %s, %v", harness.Describe(got), err)
	}
}

func TestLibraryCallsAreCheckedAgainstTheirParams(t *testing.T) {
	_, err := compileWith("half@1(n)", decimalKind, publishedHalf(t))
	if err == nil || !strings.Contains(err.Error(), "the function half@1 accepts (decimal) but was given (integer)") {
		t.Errorf("got %v", err)
	}
	_, err = compileWith("half@1(d)", integerKind, publishedHalf(t))
	if err == nil || err.Error() != "The formula produces a decimal but an integer is expected." {
		t.Errorf("got %v", err)
	}
	_, err = compileWith("half@2(d)", decimalKind, publishedHalf(t))
	if err == nil || err.Error() != "The function half@2 is not known." {
		t.Errorf("got %v", err)
	}
}

func TestLibraryFailuresReachTheCaller(t *testing.T) {
	library := publishedHalf(t)
	c := harness.Case{Source: "1 / n", Result: decimalKind, Params: []harness.Param{{Name: "n", Kind: integerKind}}}
	inverse, err := New().Compile(c, nil)
	if err != nil {
		t.Fatal(err)
	}
	library["inverse@1"] = inverse
	program, err := compileWith("half@1(inverse@1(n))", decimalKind, library)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Run(withInputs(map[string]any{"n": int64(0)})); err == nil || err.Error() != "The formula divides by zero." {
		t.Errorf("got %v", err)
	}
}

func TestOnlyProgramsOfThisEngineCanBeCalled(t *testing.T) {
	library := harness.Library{"other@1": foreignProgram{}}
	_, err := compileWith("other@1(d)", decimalKind, library)
	if err == nil || err.Error() != "The function other@1 was not built by this engine." {
		t.Errorf("got %v", err)
	}
}

func TestBuiltinNamesDoNotTakeVersions(t *testing.T) {
	_, err := compileWith("sum@1(ds)", decimalKind, nil)
	if err == nil || err.Error() != "The function sum@1 is not known." {
		t.Errorf("got %v", err)
	}
}
