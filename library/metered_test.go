package library

import (
	"sync/atomic"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var _ formula.MeteredRuntime = (*Registry)(nil)

var loopSignature = formula.Signature{
	Params: []formula.Param{{Name: "total", Type: formula.TypeDecimal}},
	Result: formula.TypeDecimal,
	Cost:   1,
}

type definedProgram struct {
	name    string
	body    string
	program *formula.Program
}

func registryOfDefinitions(t *testing.T, calls *atomic.Int64, declared int, definitions map[string]string) *Registry {
	t.Helper()
	signature := loopSignature
	signature.Cost = declared
	registry := NewRegistry()
	held := map[string]*definedProgram{}
	for name, body := range definitions {
		held[name] = &definedProgram{name: name, body: body}
		entry := held[name]
		release := Release{
			Library:   "loop",
			Name:      name,
			Version:   1,
			Signature: signature,
			Apply: func(args []formula.Value) (formula.Value, *formula.Problem) {
				calls.Add(1)
				return entry.program.Apply(args)
			},
			Metered: func(args []formula.Value, meter *formula.Meter) (formula.Value, *formula.Problem) {
				calls.Add(1)
				return entry.program.ApplyMetered(args, meter)
			},
		}
		if err := registry.Publish(release); err != nil {
			t.Fatalf("publish loop.%s@1: %v", name, err)
		}
	}
	inputs := []formula.Input{{Name: "total", Type: formula.TypeDecimal}}
	for name, entry := range held {
		program, problems := formula.Plan(parsed(t, entry.body), inputs, registry, formula.DefaultOptions())
		if len(problems) > 0 {
			t.Fatalf("loop.%s@1 does not plan: %v", name, problems)
		}
		entry.program = program
	}
	return registry
}

func TestARegistryCallLoopEndsInsteadOfRunningForever(t *testing.T) {
	cases := []struct {
		name        string
		entry       string
		definitions map[string]string
	}{
		{"a function that calls itself", "again", map[string]string{"again": "loop.again@1(total) + 1"}},
		{"two functions that call each other", "ping", map[string]string{"ping": "loop.pong@1(total)", "pong": "loop.ping@1(total)"}},
		{"a self-call that doubles each time", "twice", map[string]string{"twice": "loop.twice@1(total) + loop.twice@1(total)"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var calls atomic.Int64
			registry := registryOfDefinitions(t, &calls, 1, c.definitions)
			_, problem := registry.Apply("loop", c.entry, 1, []formula.Value{decimalOf(t, "100.00")})
			if problem == nil {
				t.Fatal("the loop returned a value")
			}
			if problem.Code != "call-too-deep" && problem.Code != "too-costly" {
				t.Fatalf("got %s: %s after %d calls", problem.Code, problem.Sentence, calls.Load())
			}
			if calls.Load() > 1000 {
				t.Fatalf("the loop made %d calls before it stopped", calls.Load())
			}
		})
	}
}

func TestASelfCallingDefinitionEndsAtCallTooDeep(t *testing.T) {
	var calls atomic.Int64
	registry := registryOfDefinitions(t, &calls, 1, map[string]string{"again": "loop.again@1(total) + 1"})
	_, problem := registry.ApplyMetered("loop", "again", 1, []formula.Value{decimalOf(t, "100.00")}, formula.NewMeter(formula.DefaultOptions().Limits.Cost))
	if problem == nil {
		t.Fatal("the self-calling definition returned a value")
	}
	if problem.Code != "call-too-deep" {
		t.Fatalf("got %s: %s", problem.Code, problem.Sentence)
	}
	if problem.Sentence != "This formula goes over the call depth limit of 4." {
		t.Errorf("sentence %q", problem.Sentence)
	}
	if calls.Load() != 6 {
		t.Errorf("the host was called %d times; the entry and four nested calls run and the fifth nested call is refused, so 6", calls.Load())
	}
}

func TestAFormulaThatCallsTheRegistryGivesTheCalleeAShareOfItsMeter(t *testing.T) {
	var calls atomic.Int64
	registry := registryOfDefinitions(t, &calls, 3, map[string]string{"plain": "total + 1"})
	caller, problems := formula.Plan(parsed(t, "loop.plain@1(total)"), []formula.Input{{Name: "total", Type: formula.TypeDecimal}}, registry, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("plan: %v", problems)
	}
	meter := formula.NewMeter(formula.DefaultOptions().Limits.Cost)
	value, problem := caller.RunMetered(map[string]formula.Value{"total": decimalOf(t, "100.00")}, meter)
	if problem != nil {
		t.Fatalf("run: %v", problem)
	}
	if value.String() != "101.00" {
		t.Errorf("the call gave %s", value.String())
	}
	if meter.Spent() != caller.Cost() {
		t.Errorf("the run spent %d and the static cost is %d", meter.Spent(), caller.Cost())
	}
}

func TestAReleaseWithoutAMeteredFormStillRunsThroughApplyMetered(t *testing.T) {
	registry := NewRegistry()
	release := Release{
		Library: "fin", Name: "flat", Version: 1, Signature: loopSignature,
		Apply: func(args []formula.Value) (formula.Value, *formula.Problem) { return args[0], nil },
	}
	if err := registry.Publish(release); err != nil {
		t.Fatal(err)
	}
	value, problem := registry.ApplyMetered("fin", "flat", 1, []formula.Value{decimalOf(t, "5.00")}, formula.NewMeter(100))
	if problem != nil || value.String() != "5.00" {
		t.Fatalf("got %v %v", value, problem)
	}
}

func TestApplyMeteredRefusesAStoppedVersionLikeApply(t *testing.T) {
	registry := finRegistry(t)
	stops := newStopTable()
	stops.put(Ref{Library: "fin", Name: "pmt", Version: 1}, "it pays out too little")
	registry.ReadStopsFrom(stops)
	_, problem := registry.ApplyMetered("fin", "pmt", 1, []formula.Value{decimalOf(t, "12000.00"), decimalOf(t, "0.01"), formula.IntegerValue(24)}, formula.NewMeter(100))
	if problem == nil || problem.Code != "version-stopped" {
		t.Fatalf("got %v", problem)
	}
}
