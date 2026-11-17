package formula

import (
	goparser "go/parser"
	gotoken "go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/number"
)

func runBigList(items int) Value {
	list := make([]number.Decimal, items)
	for i := range list {
		list[i] = number.NewDecimal(int64(i), 2)
	}
	return DecimalListValue(list)
}

func runPlanned(t *testing.T, formula string, declared []Input, runtime Runtime, options Options) *Program {
	t.Helper()
	program, problems := Plan(mustParse(t, formula), declared, runtime, options)
	if len(problems) != 0 {
		t.Fatalf("Plan(%q) gave %+v", formula, problems)
	}
	return program
}

func TestARunSpendsWhatItDoesAndNeverMoreThanItsCount(t *testing.T) {
	runtime := newRunTestRuntime(t)
	cases := []struct {
		name    string
		formula string
		inputs  string
	}{
		{"f01", "payment >= balance ? 0.00 : balance - payment", "balance=dec:120.50; payment=dec:45.25"},
		{"f03", "round((sum(deposits) - sum(expenses)) / members, 2)", "deposits=list<dec>:[500.00, 750.00, 1250.50]; expenses=list<dec>:[120.25, 80.00]; members=int:7"},
		{"f06", runF06, "amount=dec:8400.00; due=date:2026-03-31; filed=date:2026-05-15"},
		{"f10", "avg(scores) >= 80 ? \"A\" : avg(scores) >= 65 ? \"B\" : avg(scores) >= 50 ? \"C\" : \"F\"", "scores=list<int>:[72, 85, 64, 91, 77]"},
		{"f16", "count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00", "balances=list<dec>:[420.00, 380.50, 95.25, 510.00]; floor=dec:350.00"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list := runInputsOf(t, c.inputs)
			program := runPlanned(t, c.formula, runDeclared(list), runtime, DefaultOptions())
			meter := NewMeter(20000)
			if _, failure := program.RunMetered(runValuesOf(list), meter); failure != nil {
				t.Fatalf("run failed with %+v", failure)
			}
			if meter.Spent() == 0 || meter.Spent() > program.Cost() {
				t.Fatalf("spent %d, count %d", meter.Spent(), program.Cost())
			}
		})
	}
}

func TestEachNodeCostsOneAndAnAggregateAddsItsItems(t *testing.T) {
	program := runPlanned(t, "count(xs)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	meter := NewMeter(20000)
	if _, failure := program.RunMetered(map[string]Value{"xs": runBigList(7)}, meter); failure != nil {
		t.Fatalf("run failed with %+v", failure)
	}
	if meter.Spent() != 2+7 {
		t.Fatalf("spent %d, want 9: the call, the name and 7 items", meter.Spent())
	}
	written := runPlanned(t, "sum([1, 2.50])", nil, nil, DefaultOptions())
	meter = NewMeter(20000)
	if _, failure := written.RunMetered(nil, meter); failure != nil {
		t.Fatalf("run failed with %+v", failure)
	}
	if meter.Spent() != 1+1+2+2 || written.Cost() != 6 {
		t.Fatalf("spent %d, count %d, want 6 and 6", meter.Spent(), written.Cost())
	}
}

func TestACallAddsTheCalleesDeclaredCost(t *testing.T) {
	runtime := newRunTestRuntime(t)
	list := runInputsOf(t, "balance=dec:3217.45")
	program := runPlanned(t, "round(balance * interest_step@1(balance), 2)", runDeclared(list), runtime, DefaultOptions())
	meter := NewMeter(20000)
	if _, failure := program.RunMetered(runValuesOf(list), meter); failure != nil {
		t.Fatalf("run failed with %+v", failure)
	}
	want := 1 + 1 + 1 + 1 + 1 + 5 + 1
	if meter.Spent() != want || program.Cost() != want {
		t.Fatalf("spent %d, count %d, want %d for both", meter.Spent(), program.Cost(), want)
	}
}

func TestAListLargerThanAssumedStopsTheRunAtTheCostLimit(t *testing.T) {
	program := runPlanned(t, "sum(xs)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	if program.Cost() > 20000 {
		t.Fatalf("the count %d is already past the limit", program.Cost())
	}
	meter := NewMeter(20000)
	value, failure := program.RunMetered(map[string]Value{"xs": runBigList(25000)}, meter)
	if failure == nil {
		t.Fatalf("the run gave %s", runShown(value))
	}
	if failure.Code != "too-costly" || failure.Sentence != "This formula goes over the cost limit of 20000." || failure.Part != "sum(xs)" || len(failure.Path) != 0 {
		t.Fatalf("problem = %+v", *failure)
	}
	if meter.Spent() > 20000+25000 {
		t.Fatalf("spent %d: the work was not bounded", meter.Spent())
	}
}

func TestAListAtTheCostLimitStillRuns(t *testing.T) {
	program := runPlanned(t, "count(xs)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	value, failure := program.Run(map[string]Value{"xs": runBigList(19998)})
	if failure != nil || runShown(value) != "int:19998" {
		t.Fatalf("gave %s, %+v", runShown(value), failure)
	}
	_, failure = program.Run(map[string]Value{"xs": runBigList(19999)})
	if failure == nil || failure.Code != "too-costly" {
		t.Fatalf("one item over gave %+v", failure)
	}
}

func TestAHugeListIsRefusedWithoutBeingWalked(t *testing.T) {
	program := runPlanned(t, "sum(xs) + sum(xs) + sum(xs)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	meter := NewMeter(20000)
	_, failure := program.RunMetered(map[string]Value{"xs": runBigList(1000000)}, meter)
	if failure == nil || failure.Code != "too-costly" {
		t.Fatalf("problem = %+v", failure)
	}
	if meter.Spent() > 1000000+10 {
		t.Fatalf("spent %d: a second aggregate ran after the first crossed the limit", meter.Spent())
	}
}

func TestTheLimitIsASettingNotAConstant(t *testing.T) {
	options := DefaultOptions()
	options.Limits.Cost = 5
	program := runPlanned(t, "a + b * c - d", inputs("a", "integer", "b", "integer", "c", "integer", "d", "integer"), nil, options)
	list := runInputsOf(t, "a=int:1; b=int:2; c=int:3; d=int:4")
	_, failure := program.Run(runValuesOf(list))
	if failure == nil || failure.Sentence != "This formula goes over the cost limit of 5." {
		t.Fatalf("problem = %+v", failure)
	}
}

func TestOneStoppedRunDoesNotHarmTheNext(t *testing.T) {
	program := runPlanned(t, "count(xs)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	if _, failure := program.Run(map[string]Value{"xs": runBigList(30000)}); failure == nil {
		t.Fatal("the big run was not stopped")
	}
	value, failure := program.Run(map[string]Value{"xs": runBigList(3)})
	if failure != nil || runShown(value) != "int:3" {
		t.Fatalf("the next run gave %s, %+v", runShown(value), failure)
	}
}

func TestTheSameInputSpendsTheSameEveryRun(t *testing.T) {
	program := runPlanned(t, "round(sum(xs) / count(xs), 2)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	values := map[string]Value{"xs": runBigList(50)}
	var first int
	for i := 0; i < 20; i++ {
		meter := NewMeter(20000)
		if _, failure := program.RunMetered(values, meter); failure != nil {
			t.Fatalf("run failed with %+v", failure)
		}
		if i == 0 {
			first = meter.Spent()
		}
		if meter.Spent() != first {
			t.Fatalf("run %d spent %d, the first spent %d", i, meter.Spent(), first)
		}
	}
}

func TestOneProgramIsRunByManyAtOnce(t *testing.T) {
	program := runPlanned(t, "round(sum(xs) / count(xs), 2)", inputs("xs", "list of decimal"), nil, DefaultOptions())
	var group sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		group.Add(1)
		go func(size int) {
			defer group.Done()
			for i := 0; i < 50; i++ {
				value, failure := program.Run(map[string]Value{"xs": runBigList(size)})
				if failure != nil || value.Type() != TypeDecimal {
					t.Errorf("worker %d gave %v, %+v", size, value, failure)
					return
				}
			}
		}(worker + 2)
	}
	group.Wait()
}

func TestAFormulaOverItsCountIsRefusedByPlanAndNeverRuns(t *testing.T) {
	long := "max([sum(xs)"
	for i := 0; i < 20; i++ {
		long += ", sum(xs)"
	}
	long += "])"
	_, problems := Plan(mustParse(t, long), inputs("xs", "list of decimal"), nil, DefaultOptions())
	if len(problems) != 1 || problems[0].Code != "too-costly" {
		t.Fatalf("problems = %+v", problems)
	}
}

func TestNothingInTheRunReadsAClockOrARandomNumber(t *testing.T) {
	files, err := filepath.Glob("run*.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no run files found")
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := goparser.ParseFile(gotoken.NewFileSet(), file, source, goparser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range parsed.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if path == "time" || path == "math/rand" || path == "context" || path == "os" || path == "net/http" {
				t.Fatalf("%s imports %s", file, path)
			}
		}
	}
}
