package concurrency

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/library"
)

const (
	callers          = 64
	inputsPerCaller  = 2000
	shortInputsCount = 200
	repeatedRuns     = 200
	childEnvironment = "CONCURRENCY_CHILD"
	resultPrefix     = "RESULT "
)

func inputsPerRun() int {
	if testing.Short() {
		return shortInputsCount
	}
	return inputsPerCaller
}

func releaseTogether(callersCount int, work func(caller int)) {
	var ready, done sync.WaitGroup
	start := make(chan struct{})
	ready.Add(callersCount)
	done.Add(callersCount)
	for caller := 0; caller < callersCount; caller++ {
		go func(caller int) {
			defer done.Done()
			ready.Done()
			<-start
			work(caller)
		}(caller)
	}
	ready.Wait()
	close(start)
	done.Wait()
}

func TestCS01OneCompiledProgramSharedBy64CallersGivesWhatOneCallerGets(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, id := range sharedFormulaIDs {
		c := formulaNamed(id)
		program := planCase(t, c, runtime)
		inputs := make([][]map[string]formula.Value, callers)
		expected := make([][]string, callers)
		for caller := range inputs {
			inputs[caller] = generatedBatch(id, caller, inputsPerRun())
			expected[caller] = make([]string, len(inputs[caller]))
			for i, values := range inputs[caller] {
				expected[caller][i] = runKey(program, values)
			}
		}
		t.Run(c.name, func(t *testing.T) {
			var mu sync.Mutex
			var wrong []string
			releaseTogether(callers, func(caller int) {
				for i, values := range inputs[caller] {
					if got := runKey(program, values); got != expected[caller][i] {
						mu.Lock()
						wrong = append(wrong, "caller "+strconv.Itoa(caller)+" input "+strconv.Itoa(i)+": "+got+" differs from "+expected[caller][i])
						mu.Unlock()
						return
					}
				}
			})
			if len(wrong) > 0 {
				t.Fatalf("%d callers differ from a single run; first: %s", len(wrong), wrong[0])
			}
		})
	}
}

func TestCS01TheWorkedExampleOfEveryFormulaIsTheBaselineOfTheSharedRuns(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, c := range nineteenFormulas() {
		t.Run(c.name, func(t *testing.T) {
			program := planCase(t, c, runtime)
			if got := runKey(program, c.values()); got != c.result {
				t.Fatalf("formula %d gave %s, the page gives %s", c.id, got, c.result)
			}
		})
	}
}

func TestCS02ARunNeverChangesItsInputs(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, c := range nineteenFormulas() {
		t.Run(c.name, func(t *testing.T) {
			program := planCase(t, c, runtime)
			values := c.values()
			before := valuesKey(values)
			program.Run(values)
			if after := valuesKey(values); after != before {
				t.Fatalf("the inputs changed from %q to %q", before, after)
			}
		})
	}
	for _, id := range sharedFormulaIDs {
		c := formulaNamed(id)
		program := planCase(t, c, runtime)
		for i, values := range generatedBatch(id, 0, inputsPerRun()) {
			before := valuesKey(values)
			program.Run(values)
			if after := valuesKey(values); after != before {
				t.Fatalf("formula %d input %d changed from %q to %q", id, i, before, after)
			}
		}
	}
}

func TestCS03FormulasThatCallAFunctionShareItsProgramSafely(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, id := range []int{6, 8} {
		c := formulaNamed(id)
		program := planCase(t, c, runtime)
		inputs := make([][]map[string]formula.Value, callers)
		expected := make([][]string, callers)
		for caller := range inputs {
			inputs[caller] = generatedBatch(id, caller, inputsPerRun())
			expected[caller] = make([]string, len(inputs[caller]))
			for i, values := range inputs[caller] {
				expected[caller][i] = runKey(program, values)
			}
		}
		t.Run(c.name, func(t *testing.T) {
			var mu sync.Mutex
			var wrong []string
			releaseTogether(callers, func(caller int) {
				for i, values := range inputs[caller] {
					if got := runKey(program, values); got != expected[caller][i] {
						mu.Lock()
						wrong = append(wrong, "caller "+strconv.Itoa(caller)+" input "+strconv.Itoa(i))
						mu.Unlock()
						return
					}
				}
			})
			if len(wrong) > 0 {
				t.Fatalf("%d callers differ from a single run; first: %s", len(wrong), wrong[0])
			}
		})
	}
}

func TestCS03TheFunctionCalledDirectlyAtTheSameTimeEqualsItsSingleRun(t *testing.T) {
	runtime := newAgencyRuntime()
	inputs := make([][]map[string]formula.Value, callers)
	expected := make([][]string, callers)
	for caller := range inputs {
		inputs[caller] = generatedBatch(8, caller, inputsPerRun())
		expected[caller] = make([]string, len(inputs[caller]))
		for i, values := range inputs[caller] {
			value, failure := runtime.Apply("", "interest_step", 1, []formula.Value{values["balance"]})
			expected[caller][i] = outcomeKey(value, failure)
		}
	}
	var mu sync.Mutex
	wrong := 0
	releaseTogether(callers, func(caller int) {
		for i, values := range inputs[caller] {
			value, failure := runtime.Apply("", "interest_step", 1, []formula.Value{values["balance"]})
			if outcomeKey(value, failure) != expected[caller][i] {
				mu.Lock()
				wrong++
				mu.Unlock()
				return
			}
		}
	})
	if wrong > 0 {
		t.Fatalf("%d callers differ from the single call", wrong)
	}
}

func TestCS04TheSameFormulaGivesTheSameOutputEveryTime(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, c := range nineteenFormulas() {
		t.Run(c.name, func(t *testing.T) {
			program := planCase(t, c, runtime)
			values := c.values()
			first := runKey(program, values)
			for i := 1; i < repeatedRuns; i++ {
				if got := runKey(program, values); got != first {
					t.Fatalf("run %d gave %s, the first gave %s", i, got, first)
				}
			}
		})
	}
}

func TestChildReportsTheNineteenResults(t *testing.T) {
	if os.Getenv(childEnvironment) == "" {
		t.Skip("run only as the second process of CS-04")
	}
	runtime := newAgencyRuntime()
	for _, c := range nineteenFormulas() {
		program := planCase(t, c, runtime)
		os.Stdout.WriteString(resultPrefix + strconv.Itoa(c.id) + " " + runKey(program, c.values()) + "\n")
	}
}

func TestCS04ASecondProcessGivesTheSameOutput(t *testing.T) {
	if os.Getenv(childEnvironment) != "" {
		t.Skip("this is the second process")
	}
	command := exec.Command(os.Args[0], "-test.run=^TestChildReportsTheNineteenResults$", "-test.count=1")
	command.Env = append(os.Environ(), childEnvironment+"=1")
	var out bytes.Buffer
	command.Stdout = &out
	if err := command.Run(); err != nil {
		t.Fatalf("the second process failed: %v\n%s", err, out.String())
	}
	reported := map[int]string{}
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, resultPrefix) {
			continue
		}
		id, key, found := strings.Cut(strings.TrimPrefix(line, resultPrefix), " ")
		if !found {
			t.Fatalf("a result line has no key: %q", line)
		}
		number, err := strconv.Atoi(id)
		if err != nil {
			t.Fatalf("a result line has no number: %q", line)
		}
		reported[number] = key
	}
	runtime := newAgencyRuntime()
	for _, c := range nineteenFormulas() {
		program := planCase(t, c, runtime)
		if want := runKey(program, c.values()); reported[c.id] != want {
			t.Fatalf("formula %d gave %q in the second process and %q here", c.id, reported[c.id], want)
		}
	}
}

func TestCS05CompilingTheSameTextAtTheSameMomentGivesEqualPrograms(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, id := range []int{1, 3, 8, 15} {
		c := formulaNamed(id)
		reference := planCase(t, c, runtime)
		wantKey := runKey(reference, c.values())
		wantCost := reference.Cost()
		var mu sync.Mutex
		var wrong []string
		releaseTogether(32, func(caller int) {
			tree, problems := formula.Parse(c.body)
			if len(problems) > 0 {
				mu.Lock()
				wrong = append(wrong, "parse: "+problems[0].Sentence)
				mu.Unlock()
				return
			}
			program, problems := formula.Plan(tree, c.inputs(), runtime, formula.DefaultOptions())
			if len(problems) > 0 {
				mu.Lock()
				wrong = append(wrong, "plan: "+problems[0].Sentence)
				mu.Unlock()
				return
			}
			if program.Cost() != wantCost || runKey(program, c.values()) != wantKey || formula.Format(tree) != formula.Format(mustTree(t, c.body)) {
				mu.Lock()
				wrong = append(wrong, "caller "+strconv.Itoa(caller)+" built a program that differs")
				mu.Unlock()
			}
		})
		if len(wrong) > 0 {
			t.Fatalf("formula %d: %d of 32 compiles differ; first: %s", id, len(wrong), wrong[0])
		}
	}
}

func mustTree(t testing.TB, body string) formula.Tree {
	t.Helper()
	tree, problems := formula.Parse(body)
	if len(problems) > 0 {
		t.Fatalf("parse %q: %s", body, problems[0].Sentence)
	}
	return tree
}

func TestCS05ExactlyOneProgramIsBuiltForOneText(t *testing.T) {
	t.Skip("held: the library keeps no cache of compiled programs, so there is nothing to count; a host that keeps one owns this case")
}

func interestStepRegistry(t testing.TB) *library.Registry {
	t.Helper()
	registry := library.NewRegistry()
	signature := formula.Signature{
		Params: []formula.Param{{Name: "balance", Type: formula.TypeDecimal}},
		Result: formula.TypeDecimal,
		Cost:   5,
	}
	bands := func(low string, high string, lowRate string, midRate string, highRate string) library.Implementation {
		lowLimit, midLimit := dec(low).Decimal(), dec(high).Decimal()
		return func(args []formula.Value) (formula.Value, *formula.Problem) {
			balance := args[0].Decimal()
			switch {
			case balance.Cmp(lowLimit) < 0:
				return dec(lowRate), nil
			case balance.Cmp(midLimit) < 0:
				return dec(midRate), nil
			}
			return dec(highRate), nil
		}
	}
	releases := []library.Release{
		{Library: "fin", Name: "interest_step", Version: 1, Signature: signature, Apply: bands("1000.00", "5000.00", "0.010", "0.015", "0.020")},
		{
			Library: "fin", Name: "interest_step", Version: 2, Signature: signature, Apply: bands("2000.00", "8000.00", "0.008", "0.012", "0.018"),
			Note: library.Note{
				Changed: "The bands moved up to 2000.00 and 8000.00 and every rate is lower.",
				Matters: "Any agency that charges interest on a balance.",
				Example: library.Example{Input: "balance 3217.45", Old: "0.015", New: "0.012"},
			},
		},
	}
	for _, release := range releases {
		if err := registry.Publish(release); err != nil {
			t.Fatalf("publish %s@%d: %v", release.Name, release.Version, err)
		}
	}
	return registry
}

func TestCS06TwoPinnedVersionsOfOneFunctionNeverMix(t *testing.T) {
	registry := interestStepRegistry(t)
	balanceInput := []formula.Input{{Name: "balance", Type: formula.TypeDecimal}}
	plan := func(body string) *formula.Program {
		program, problems := formula.Plan(mustTree(t, body), balanceInput, registry, formula.DefaultOptions())
		if len(problems) > 0 {
			t.Fatalf("plan %q: %s", body, problems[0].Sentence)
		}
		return program
	}
	older, newer := plan("fin.interest_step@1(balance)"), plan("fin.interest_step@2(balance)")
	inputs := generatedBatch(8, 0, inputsPerRun())
	expectedOlder := make([]string, len(inputs))
	expectedNewer := make([]string, len(inputs))
	for i, values := range inputs {
		expectedOlder[i] = runKey(older, values)
		expectedNewer[i] = runKey(newer, values)
	}
	differ := 0
	for i := range inputs {
		if expectedOlder[i] != expectedNewer[i] {
			differ++
		}
	}
	if differ == 0 {
		t.Fatal("the two versions never differ on these inputs, so the case would prove nothing")
	}
	var mu sync.Mutex
	wrong := 0
	releaseTogether(callers, func(caller int) {
		for i, values := range inputs {
			var got, want string
			if (caller+i)%2 == 0 {
				got, want = runKey(older, values), expectedOlder[i]
			} else {
				got, want = runKey(newer, values), expectedNewer[i]
			}
			if got != want {
				mu.Lock()
				wrong++
				mu.Unlock()
				return
			}
		}
	})
	if wrong > 0 {
		t.Fatalf("%d callers got the program of the other version", wrong)
	}
}

func TestCS06TheCacheHoldsOneEntryForEachNameAndVersion(t *testing.T) {
	t.Skip("held: the library keeps no cache of compiled programs; the registry holds one release for each name and version, which TestCS06TwoPinnedVersionsOfOneFunctionNeverMix reads through two pinned programs")
}
