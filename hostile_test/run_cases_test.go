package hostile

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type runCase struct {
	id       string
	text     string
	rows     int
	code     string
	sentence string
	result   string
}

var costSentence = limitSentence("cost", 20000)

var runCases = []runCase{
	{id: "L02", text: nested(11, "!(", ")", "flag"), result: "false"},
	{id: "L04", text: manyItems(98), result: "98"},
	{id: "L05", text: balancedSum(16, "sum(rows)"), rows: 25000, code: "too-costly", sentence: costSentence},
	{id: "L07", text: "2 + 3", result: "5"},
	{id: "C02", text: "sum(rows)", rows: 25000, code: "too-costly", sentence: costSentence},
	{id: "C04", text: "sum(rows) + sum(rows)", rows: 15000, code: "too-costly", sentence: costSentence},
	{id: "C06", text: balancedSum(8, "sum(rows)"), rows: 5000, code: "too-costly", sentence: costSentence},
	{id: "I08", text: "sum(rows) + sum(rows) + sum(rows)", rows: 9000, code: "too-costly", sentence: costSentence},
}

func TestRunCases(t *testing.T) {
	for index, c := range runCases {
		c := c
		t.Run(c.id+"/"+itoa(index), func(t *testing.T) {
			program := plan(t, c.text, standardInputs())
			value, problem := program.Run(standardValues(c.rows))
			if c.code == "" {
				if problem != nil {
					t.Fatalf("%s: unexpected problem %s", c.id, shown(*problem))
				}
				if value.String() != c.result {
					t.Fatalf("%s: result %q, want %q", c.id, value.String(), c.result)
				}
				return
			}
			if problem == nil {
				t.Fatalf("%s: the run gave %q and no problem", c.id, value.String())
			}
			requirePlain(t, *problem)
			if problem.Code != c.code || problem.Sentence != c.sentence {
				t.Fatalf("%s: got %s, want %s: %s", c.id, shown(*problem), c.code, c.sentence)
			}
		})
	}
}

func TestEveryCostRefusalIsTheSameSentenceAtSaveAndAtRun(t *testing.T) {
	saved := atSave(balancedSum(32, "sum(rows)"), rowsInput)
	program := plan(t, balancedSum(16, "sum(rows)"), rowsInput)
	_, ran := program.Run(map[string]formula.Value{"rows": standardValues(25000)["rows"]})
	if len(saved) == 0 || ran == nil || saved[0].Sentence != ran.Sentence {
		t.Fatalf("save %+v and run %+v differ", saved, ran)
	}
}

func TestCostIsACountAndTheSameOnEveryRun(t *testing.T) {
	program := plan(t, "sum(rows) + sum(rows)", standardInputs())
	var spent []int
	for i := 0; i < 20; i++ {
		meter := formula.NewMeter(20000)
		if _, problem := program.RunMetered(standardValues(500), meter); problem != nil {
			t.Fatalf("unexpected %s", shown(*problem))
		}
		spent = append(spent, meter.Spent())
	}
	for _, n := range spent {
		if n != spent[0] {
			t.Fatalf("spent %v differs between runs", spent)
		}
	}
}

func TestAGoodFormulaStillRunsAfterEveryRefusal(t *testing.T) {
	for _, c := range allSaveCases() {
		atSave(c.text, c.inputs)
	}
	for _, c := range runCases {
		program := plan(t, c.text, standardInputs())
		program.Run(standardValues(c.rows))
	}
	program := plan(t, "2 + 3", nil)
	value, problem := program.Run(nil)
	if problem != nil || value.String() != "5" {
		t.Fatalf("2 + 3 gave %q and %v", value.String(), problem)
	}
}

func TestNothingIsLeftBehindAfterTheHostileCases(t *testing.T) {
	runtime.GC()
	goroutines := runtime.NumGoroutine()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	for _, c := range allSaveCases() {
		atSave(c.text, c.inputs)
	}
	for _, c := range runCases {
		program := plan(t, c.text, standardInputs())
		program.Run(standardValues(c.rows))
	}
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	if got := runtime.NumGoroutine(); got != goroutines {
		t.Fatalf("goroutines %d before and %d after", goroutines, got)
	}
	if after.HeapAlloc > before.HeapAlloc+32<<20 {
		t.Fatalf("heap %d bytes before and %d after", before.HeapAlloc, after.HeapAlloc)
	}
}

func TestTenHostileFormulasAtOnceOnOneEngine(t *testing.T) {
	hostile := []runCase{runCases[2], runCases[4], runCases[5], runCases[6], runCases[7]}
	var wait sync.WaitGroup
	failures := make(chan string, 64)
	good := plan(t, "2 + 3", nil)
	programs := make([]*formula.Program, len(hostile))
	for i, c := range hostile {
		programs[i] = plan(t, c.text, standardInputs())
	}
	for round := 0; round < 2; round++ {
		for i, c := range hostile {
			c := c
			program := programs[i]
			wait.Add(1)
			go func() {
				defer wait.Done()
				_, problem := program.Run(standardValues(c.rows))
				if problem == nil || problem.Sentence != c.sentence {
					failures <- c.id + ": wrong outcome"
				}
			}()
		}
	}
	wait.Add(1)
	go func() {
		defer wait.Done()
		value, problem := good.Run(nil)
		if problem != nil || value.String() != "5" {
			failures <- "the good formula was spoiled"
		}
	}()
	wait.Wait()
	close(failures)
	for failure := range failures {
		t.Error(failure)
	}
}

type loopRuntime struct {
	library
	programs map[string]*formula.Program
	calls    int
}

const hostStopped = "host-stopped"

func (l *loopRuntime) Function(lib string, name string, version int) (formula.Signature, bool) {
	if _, found := l.programs[name]; found && lib == "" && version == 1 {
		return formula.Signature{Params: []formula.Param{{Name: "p", Type: formula.TypeDecimal}}, Result: formula.TypeDecimal, Cost: 1}, true
	}
	return l.library.Function(lib, name, version)
}

func (l *loopRuntime) Apply(lib string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	l.calls++
	if l.calls > 2000 {
		return formula.Value{}, &formula.Problem{Code: hostStopped, Sentence: "The host stopped the loop."}
	}
	return l.programs[name].Apply(args)
}

func loopProgram(t *testing.T, bodies map[string]string) (*loopRuntime, map[string]*formula.Program) {
	t.Helper()
	runtimeValue := &loopRuntime{library: hostileLibrary(), programs: map[string]*formula.Program{}}
	for name := range bodies {
		runtimeValue.programs[name] = nil
	}
	for name, body := range bodies {
		tree, problems := formula.Parse(body)
		if len(problems) > 0 {
			t.Fatalf("Parse(%q) gave %+v", body, problems)
		}
		program, problems := formula.Plan(tree, totalInput, runtimeValue, formula.DefaultOptions())
		if len(problems) > 0 {
			t.Fatalf("Plan(%q) gave %+v", body, problems)
		}
		runtimeValue.programs[name] = program
	}
	return runtimeValue, runtimeValue.programs
}

func TestSelfCallsSlippedPastTheSaveCheckAreStopped(t *testing.T) {
	cases := []struct {
		id     string
		entry  string
		bodies map[string]string
	}{
		{"C10", "f", map[string]string{"f": "f@1(total) + 1"}},
		{"C12", "a", map[string]string{"a": "b@1(total)", "b": "a@1(total)"}},
		{"C13", "f", map[string]string{"f": "f@1(total) + f@1(total)"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.id, func(t *testing.T) {
			runtimeValue, programs := loopProgram(t, c.bodies)
			_, problem := programs[c.entry].Run(map[string]formula.Value{"total": formula.DecimalValue(number.NewDecimal(100, 2))})
			if problem == nil {
				t.Fatalf("%s: the loop returned a value", c.id)
			}
			if problem.Code == hostStopped {
				t.Fatalf("%s: the library never stopped the loop; the host had to after %d calls", c.id, runtimeValue.calls)
			}
			if problem.Code != "too-costly" && problem.Code != "call-too-deep" {
				t.Fatalf("%s: got %s", c.id, shown(*problem))
			}
		})
	}
}

func TestSameResultOnEveryRunFreshAndBusy(t *testing.T) {
	program := plan(t, "round(total / people, 2) + sum(rows)", standardInputs())
	values := standardValues(1000)
	first, problem := program.Run(values)
	if problem != nil {
		t.Fatalf("unexpected %s", shown(*problem))
	}
	var wait sync.WaitGroup
	results := make(chan string, 400)
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := 0; i < 50; i++ {
				value, problem := program.Run(values)
				if problem != nil {
					results <- "problem"
					continue
				}
				results <- value.String()
			}
		}()
	}
	wait.Wait()
	close(results)
	for result := range results {
		if result != first.String() {
			t.Fatalf("result %q differs from %q", result, first.String())
		}
	}
}

const childVariable = "HOSTILE_CHILD_SUM"

func TestChildProcessSum(t *testing.T) {
	if os.Getenv(childVariable) == "" {
		t.Skip("run only as the child of the two process case")
	}
	program := plan(t, "sum(rows)", rowsInput)
	value, problem := program.Run(map[string]formula.Value{"rows": standardValues(1000)["rows"]})
	if problem != nil {
		t.Fatalf("unexpected %s", shown(*problem))
	}
	os.Stdout.WriteString("RESULT:" + value.String() + "\n")
}

func TestSumOfManyDecimalsIsTheSameInTwoProcesses(t *testing.T) {
	program := plan(t, "sum(rows)", rowsInput)
	value, problem := program.Run(map[string]formula.Value{"rows": standardValues(1000)["rows"]})
	if problem != nil {
		t.Fatalf("unexpected %s", shown(*problem))
	}
	command := exec.Command(os.Args[0], "-test.run=^TestChildProcessSum$", "-test.v")
	command.Env = append(os.Environ(), childVariable+"=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, output)
	}
	want := "RESULT:" + value.String()
	if !strings.Contains(string(output), want) {
		t.Fatalf("child output %q does not hold %q", output, want)
	}
}
