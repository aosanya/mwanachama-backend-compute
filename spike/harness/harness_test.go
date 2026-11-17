package harness_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness/stub"
)

func failing(results []harness.Result) []int {
	var numbers []int
	for _, r := range results {
		if !r.Passed {
			numbers = append(numbers, r.Number)
		}
	}
	return numbers
}

func resultFor(t *testing.T, results []harness.Result, number int) harness.Result {
	t.Helper()
	for _, r := range results {
		if r.Number == number {
			return r
		}
	}
	t.Fatalf("no result for formula %d", number)
	return harness.Result{}
}

func TestEveryFormulaPassesOnTheStub(t *testing.T) {
	results := harness.Run(stub.Engine{})
	if len(results) != 19 {
		t.Fatalf("got %d results", len(results))
	}
	if bad := failing(results); len(bad) != 0 {
		t.Errorf("failed formulas %v:\n%s", bad, harness.Format(results))
	}
	for i, r := range results {
		if r.Number != i+1 {
			t.Errorf("result %d is formula %d", i, r.Number)
		}
	}
}

func TestRunTestPassesOnTheStub(t *testing.T) {
	harness.RunTest(t, stub.Engine{})
}

func TestReportSummarisesTheRun(t *testing.T) {
	report := harness.NewReport(stub.Engine{})
	if !report.AllPassed() || report.PassCount() != 19 || report.Engine != "stub" {
		t.Errorf("report is %+v", report)
	}
	if !strings.HasSuffix(report.String(), "stub: 19 of 19 formulas pass\n") {
		t.Errorf("report text is:\n%s", report)
	}
	broken := harness.NewReport(stub.Engine{WrongAt: 4})
	if broken.AllPassed() || broken.PassCount() != 18 {
		t.Errorf("broken report passes %d", broken.PassCount())
	}
}

func TestWrongAnswerFailsOnlyThatFormulaAndNamesIt(t *testing.T) {
	for number := 1; number <= 19; number++ {
		results := harness.Run(stub.Engine{WrongAt: number})
		bad := failing(results)
		if len(bad) != 1 || bad[0] != number {
			t.Errorf("wrong at %d failed %v", number, bad)
			continue
		}
		r := resultFor(t, results, number)
		if !strings.Contains(r.Reason, "want") || !strings.Contains(r.Reason, "got") {
			t.Errorf("formula %d reason is %q", number, r.Reason)
		}
		line := harness.Format(results)
		if !strings.Contains(line, fmt.Sprintf("FAIL %2d  %s", number, r.Title)) {
			t.Errorf("format does not name formula %d:\n%s", number, line)
		}
	}
}

func TestWrongAnswerReasonShowsBothValues(t *testing.T) {
	r := resultFor(t, harness.Run(stub.Engine{WrongAt: 1}), 1)
	for _, part := range []string{`example "partial payment"`, "want 75.25 (decimal)", "got 76.25 (decimal)"} {
		if !strings.Contains(r.Reason, part) {
			t.Errorf("reason %q lacks %q", r.Reason, part)
		}
	}
}

func TestRunErrorFailsTheFormula(t *testing.T) {
	results := harness.Run(stub.Engine{RunErrorAt: 3})
	if bad := failing(results); len(bad) != 1 || bad[0] != 3 {
		t.Fatalf("failed %v", bad)
	}
	r := resultFor(t, results, 3)
	if !strings.Contains(r.Reason, "run error") || !strings.Contains(r.Reason, stub.ErrDeliberate.Error()) {
		t.Errorf("reason is %q", r.Reason)
	}
	if got := strings.Count(r.Reason, "run error"); got != len(harness.Cases()[2].Examples) {
		t.Errorf("reason lists %d run errors", got)
	}
}

func TestCompileErrorFailsTheFormulaAndItsDependants(t *testing.T) {
	results := harness.Run(stub.Engine{CompileErrorAt: 7})
	bad := failing(results)
	if len(bad) != 2 || bad[0] != 7 || bad[1] != 8 {
		t.Fatalf("failed %v", bad)
	}
	if reason := resultFor(t, results, 7).Reason; !strings.HasPrefix(reason, "compile: ") {
		t.Errorf("reason is %q", reason)
	}
	if reason := resultFor(t, results, 8).Reason; !strings.Contains(reason, "interest_step@1") {
		t.Errorf("reason is %q", reason)
	}
}

type orderChecking struct {
	stub.Engine
	seen map[string]bool
}

func (o orderChecking) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	for _, name := range c.Uses {
		if library[name] == nil {
			return nil, fmt.Errorf("formula %d compiled before %s", c.Number, name)
		}
	}
	o.seen[c.Title] = true
	return o.Engine.Compile(c, library)
}

func TestPublishedFormulasAreCompiledAndStoredBeforeTheirUsers(t *testing.T) {
	checker := orderChecking{seen: map[string]bool{}}
	results := harness.Run(checker)
	if bad := failing(results); len(bad) != 0 {
		t.Errorf("failed %v: %s", bad, harness.Format(results))
	}
	if len(checker.seen) != 19 {
		t.Errorf("compiled %d formulas", len(checker.seen))
	}
}

type panicking struct{ stub.Engine }

func (p panicking) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	if c.Number == 2 {
		panic("boom")
	}
	return p.Engine.Compile(c, library)
}

type nilProgram struct{ stub.Engine }

func (n nilProgram) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	if c.Number == 9 {
		return nil, nil
	}
	return n.Engine.Compile(c, library)
}

type runPanics struct{}

func (runPanics) Run(map[string]any) (any, error) { panic("bang") }

type panicOnRun struct{ stub.Engine }

func (p panicOnRun) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	if c.Number == 10 {
		return runPanics{}, nil
	}
	return p.Engine.Compile(c, library)
}

func TestEngineMisbehaviourBecomesAFailureNotACrash(t *testing.T) {
	cases := map[int]struct {
		engine harness.Engine
		reason string
	}{
		2:  {panicking{}, "panic: boom"},
		9:  {nilProgram{}, "no program"},
		10: {panicOnRun{}, "panic: bang"},
	}
	for number, c := range cases {
		results := harness.Run(c.engine)
		if bad := failing(results); len(bad) != 1 || bad[0] != number {
			t.Errorf("formula %d: failed %v", number, bad)
			continue
		}
		if reason := resultFor(t, results, number).Reason; !strings.Contains(reason, c.reason) {
			t.Errorf("formula %d: reason %q lacks %q", number, reason, c.reason)
		}
	}
}

type mutating struct{ stub.Engine }

func (m mutating) Compile(c harness.Case, library harness.Library) (harness.Program, error) {
	inner, err := m.Engine.Compile(c, library)
	if err != nil {
		return nil, err
	}
	return mutatingProgram{inner}, nil
}

type mutatingProgram struct{ inner harness.Program }

func (m mutatingProgram) Run(inputs map[string]any) (any, error) {
	got, err := m.inner.Run(inputs)
	for name := range inputs {
		delete(inputs, name)
	}
	return got, err
}

func TestProgramsCannotDisturbTheExampleInputs(t *testing.T) {
	first := harness.Run(mutating{})
	if bad := failing(first); len(bad) != 0 {
		t.Errorf("failed %v", bad)
	}
	if bad := failing(harness.Run(stub.Engine{})); len(bad) != 0 {
		t.Errorf("cases were disturbed, failed %v", bad)
	}
}

func TestFormatPrintsOneLinePerFormula(t *testing.T) {
	text := harness.Format(harness.Run(stub.Engine{WrongAt: 6}))
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) != 19 {
		t.Fatalf("got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "PASS  1  Balance after payment") {
		t.Errorf("first line is %q", lines[0])
	}
	if !strings.HasPrefix(lines[5], "FAIL  6  Late-filing penalty") {
		t.Errorf("sixth line is %q", lines[5])
	}
}

func TestStubRefusesWhatItDoesNotKnow(t *testing.T) {
	engine := stub.Engine{}
	if _, err := engine.Compile(harness.Case{Number: 99, Source: "1 + 1"}, harness.Library{}); err == nil {
		t.Error("unknown source should not compile")
	}
	program, err := engine.Compile(harness.Cases()[0], harness.Library{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Run(map[string]any{"balance": int64(1)}); err == nil {
		t.Error("unknown inputs should fail")
	}
}
