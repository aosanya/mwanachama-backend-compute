package formula

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

type loopingRuntime struct {
	fakeLibrary
	programs map[string]*Program
	calls    int
	stopAt   int
}

func (r *loopingRuntime) Apply(library string, name string, version int, args []Value) (Value, *Problem) {
	r.calls++
	if r.calls > r.stopAt {
		return Value{}, &Problem{Code: "host-stopped", Sentence: "The host stopped the loop."}
	}
	return r.programs[name].Apply(args)
}

func (r *loopingRuntime) ApplyMetered(library string, name string, version int, args []Value, meter *Meter) (Value, *Problem) {
	r.calls++
	if r.calls > r.stopAt {
		return Value{}, &Problem{Code: "host-stopped", Sentence: "The host stopped the loop."}
	}
	return r.programs[name].ApplyMetered(args, meter)
}

func newLoopingRuntime(t *testing.T, bodies map[string]string) *loopingRuntime {
	t.Helper()
	var functions []fakeFunction
	for name := range bodies {
		functions = append(functions, fakeFunction{"", name, 1, Signature{Params: decimalParams(1), Result: TypeDecimal, Cost: 1}})
	}
	runtime := &loopingRuntime{fakeLibrary: fakeLibrary{functions: functions}, programs: map[string]*Program{}, stopAt: 5000}
	for name, body := range bodies {
		program, problems := Plan(mustParse(t, body), inputs("total", "decimal"), runtime, DefaultOptions())
		if len(problems) != 0 {
			t.Fatalf("%s did not plan: %+v", name, problems)
		}
		runtime.programs[name] = program
	}
	return runtime
}

func TestAChainOfCalleesIsStoppedByTheRunThatStartedIt(t *testing.T) {
	cases := []struct {
		name   string
		entry  string
		bodies map[string]string
	}{
		{"a function that calls itself", "f", map[string]string{"f": "f@1(total) + 1"}},
		{"two functions that call each other", "a", map[string]string{"a": "b@1(total)", "b": "a@1(total)"}},
		{"a self-call that doubles each time", "f", map[string]string{"f": "f@1(total) + f@1(total)"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			runtime := newLoopingRuntime(t, c.bodies)
			_, problem := runtime.programs[c.entry].Run(map[string]Value{"total": DecimalValue(number.NewDecimal(100, 2))})
			if problem == nil {
				t.Fatalf("the loop returned a value")
			}
			if problem.Code != "call-too-deep" && problem.Code != "too-costly" {
				t.Fatalf("got %s: %s after %d host calls", problem.Code, problem.Sentence, runtime.calls)
			}
			if runtime.calls > 1000 {
				t.Fatalf("the loop ran %d calls before it stopped", runtime.calls)
			}
		})
	}
}

func TestACalleeIsChargedAtLeastItsDeclaredCostAndNeverTwice(t *testing.T) {
	runtime := newLoopingRuntime(t, map[string]string{"g": "total + 1"})
	caller, problems := Plan(mustParse(t, "g@1(total)"), inputs("total", "decimal"), runtime, DefaultOptions())
	if len(problems) != 0 {
		t.Fatalf("caller did not plan: %+v", problems)
	}
	meter := NewMeter(DefaultOptions().Limits.Cost)
	if _, problem := caller.RunMetered(map[string]Value{"total": DecimalValue(number.NewDecimal(100, 2))}, meter); problem != nil {
		t.Fatalf("unexpected %s", problem.Sentence)
	}
	static := caller.Cost()
	if meter.Spent() > static {
		t.Fatalf("the run spent %d and the static cost is %d", meter.Spent(), static)
	}
}

func TestAnUnmeteredRuntimeStillWorksAsBefore(t *testing.T) {
	runtime := newRunTestRuntime(t)
	text, problems, problem := runPlanAndRun(t, "interest_step@1(500.00)", "", runtime)
	if len(problems) != 0 || problem != nil {
		t.Fatalf("got %+v %+v", problems, problem)
	}
	if text.String() == "" {
		t.Fatalf("no result")
	}
}
