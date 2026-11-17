package concurrency

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	evenShareBody = "round(total / people, 2)"
	tickBody      = "round(balance * tick@1(balance), 2)"
	mixedRows     = 2000
)

func decimalInput(name string, places int) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeDecimal, Rules: []function.Rule{
		{Name: problem.RuleRequired},
		{Name: problem.RuleNoNegatives},
		{Name: problem.RuleDecimalPlaces, Places: &places},
	}}
}

func integerInput(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeInteger, Rules: []function.Rule{{Name: problem.RuleRequired}}}
}

func evenShareDefinition() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("total", 2), integerInput("people")}}
}

func tickDefinition() function.Definition {
	return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2)}}
}

func row(pairs ...string) function.Row {
	result := function.Row{}
	for i := 0; i+1 < len(pairs); i += 2 {
		result[pairs[i]] = numtext.One(pairs[i+1])
	}
	return result
}

func cents(total int) string {
	return fmt.Sprintf("%d.%02d", total/100, total%100)
}

func mixedBatch(count int) []function.Row {
	rows := make([]function.Row, count)
	for i := range rows {
		switch i % 7 {
		case 4:
			rows[i] = row("total", cents(1000*(i+1)), "people", "0")
		case 5:
			rows[i] = row("total", fmt.Sprintf("%d.%03d", i+1, i%1000), "people", "3")
		case 6:
			rows[i] = row("total", "-"+cents(100*(i+1)), "people", "2")
		default:
			rows[i] = row("total", cents(1000*(i+1)), "people", fmt.Sprint(1+i%9))
		}
	}
	return rows
}

func tickBatch(count int) []function.Row {
	rows := make([]function.Row, count)
	for i := range rows {
		rows[i] = row("balance", cents(1000*(i+1)))
	}
	return rows
}

func rowKey(result batch.RowResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d|%s|", result.Row, result.StoppedAt)
	if len(result.Problems) == 0 {
		fmt.Fprintf(&b, "%s:%s", result.Value.Type(), result.Value.String())
		return b.String()
	}
	for _, found := range result.Problems {
		fmt.Fprintf(&b, "%s/%s/%s;", found.Code, inputName(found), found.Sentence)
	}
	return b.String()
}

func listKey(result batch.Result) []string {
	keys := make([]string, len(result.Rows))
	for i, found := range result.Rows {
		keys[i] = rowKey(found)
	}
	return keys
}

func sameLists(t *testing.T, got []string, want []string, what string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: %d rows, want %d", what, len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: row %d is %q, want %q", what, i+1, got[i], want[i])
		}
	}
}

type tickRuntime struct {
	calls *atomic.Int64
}

func newTickRuntime() tickRuntime { return tickRuntime{calls: &atomic.Int64{}} }

var tickSignature = formula.Signature{
	Params: []formula.Param{{Name: "balance", Type: formula.TypeDecimal}},
	Result: formula.TypeDecimal,
	Cost:   5,
}

func (r tickRuntime) Function(library string, name string, version int) (formula.Signature, bool) {
	if library == "" && name == "tick" && version == 1 {
		return tickSignature, true
	}
	return formula.Signature{}, false
}

func (r tickRuntime) Versions(library string, name string) []int {
	if library == "" && name == "tick" {
		return []int{1}
	}
	return nil
}

func (r tickRuntime) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	r.calls.Add(1)
	return dec("0.005"), nil
}

func compileOrFail(t *testing.T, definition function.Definition, body string, runtime formula.Runtime) *batch.Compiled {
	t.Helper()
	compiled, problems := batch.Compile(definition, body, runtime, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("compile %q: %s", body, problems[0].Sentence)
	}
	return compiled
}

func runOrFail(t *testing.T, compiled *batch.Compiled, rows []function.Row, options batch.Options) batch.Result {
	t.Helper()
	result, err := compiled.RunBatch(context.Background(), rows, options)
	if err != nil {
		t.Fatalf("the batch was refused: %v", err)
	}
	return result
}

func TestCS07WorkersAndChunkSizeChangeNoAnswer(t *testing.T) {
	compiled := compileOrFail(t, evenShareDefinition(), evenShareBody, nil)
	rows := mixedBatch(mixedRows)
	reference := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 100, Workers: 1}))
	stages := map[batch.Stage]int{}
	for _, found := range runOrFail(t, compiled, rows, batch.Options{ChunkSize: 100, Workers: 1}).Rows {
		stages[found.StoppedAt]++
	}
	if stages[batch.StageValidate] == 0 || stages[batch.StageCalculate] == 0 || stages[""] == 0 {
		t.Fatalf("the rows must mix values, validate problems and calculation problems, got %v", stages)
	}
	for _, workers := range []int{1, 8} {
		for _, chunk := range []int{1, 7, mixedRows, mixedRows + 1} {
			t.Run(fmt.Sprintf("workers %d chunk %d", workers, chunk), func(t *testing.T) {
				got := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: chunk, Workers: workers}))
				sameLists(t, got, reference, "the result list")
			})
		}
	}
	for i := 0; i < len(rows); i += 97 {
		alone := compiled.Run(rows[i])
		alone.Row = i + 1
		if rowKey(alone) != reference[i] {
			t.Fatalf("row %d alone is %q, in the batch %q", i+1, rowKey(alone), reference[i])
		}
	}
}

func TestCS08TheSameRowsRunTwiceWithEightWorkersGiveTheSameList(t *testing.T) {
	compiled := compileOrFail(t, evenShareDefinition(), evenShareBody, nil)
	rows := mixedBatch(mixedRows)
	first := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 7, Workers: 8}))
	second := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 7, Workers: 8}))
	sameLists(t, second, first, "the second run")
}

func TestCS09ABatchOverTheRowSettingIsRefusedWholeAndNothingStarts(t *testing.T) {
	ticks := newTickRuntime()
	compiled := compileOrFail(t, tickDefinition(), tickBody, ticks)
	before := goroutineCount()
	result, err := compiled.RunBatch(context.Background(), tickBatch(batch.DefaultMaxRows+1), batch.Options{Workers: 8})
	var refusal *batch.BatchError
	if !errors.As(err, &refusal) || refusal.Code != batch.CodeTooManyRows {
		t.Fatalf("the error was %v, wanted too-many-rows", err)
	}
	if len(result.Rows) != 0 {
		t.Fatalf("the refusal came with %d rows", len(result.Rows))
	}
	if worked := ticks.calls.Load(); worked != 0 {
		t.Fatalf("%d rows were worked out before the refusal", worked)
	}
	if after := goroutineCount(); after > before {
		t.Fatalf("the refusal left %d goroutines running, there were %d", after, before)
	}
}

func goroutineCount() int { return runtime.NumGoroutine() }

func rowCost(t *testing.T) int {
	t.Helper()
	program, problems := formula.Plan(mustTree(t, tickBody), []formula.Input{{Name: "balance", Type: formula.TypeDecimal}}, newTickRuntime(), formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("plan: %s", problems[0].Sentence)
	}
	return program.Cost()
}

type limitedBatch struct {
	valued, notRun, other int
	startedPositions      []int
}

func readLimited(result batch.Result) limitedBatch {
	var counted limitedBatch
	for i, found := range result.Rows {
		switch {
		case found.HasValue():
			counted.valued++
			counted.startedPositions = append(counted.startedPositions, i)
		case found.StoppedAt == batch.StageBatch && len(found.Problems) == 1 && found.Problems[0].Code == batch.CodeBatchLimitReached:
			counted.notRun++
		default:
			counted.other++
		}
	}
	return counted
}

func TestCS10ARowsNotStartedUnderTheBatchCostLimitCarryTheirCode(t *testing.T) {
	cost := rowCost(t)
	const limitedRows = 10000
	const startedRows = 25
	compiled := compileOrFail(t, tickDefinition(), tickBody, newTickRuntime())
	rows := tickBatch(limitedRows)
	result := runOrFail(t, compiled, rows, batch.Options{ChunkSize: 1, Workers: 1, MaxRows: limitedRows, Cost: startedRows * cost})
	counted := readLimited(result)
	if counted.valued != startedRows || counted.notRun != limitedRows-startedRows || counted.other != 0 {
		t.Fatalf("valued %d, not started %d, other %d; wanted %d, %d, 0", counted.valued, counted.notRun, counted.other, startedRows, limitedRows-startedRows)
	}
	for position, started := range counted.startedPositions {
		if started != position {
			t.Fatalf("the rows started are not the first in order: the %dth started row is row %d", position+1, started+1)
		}
	}
	if counted.valued+counted.notRun+counted.other != limitedRows {
		t.Fatal("the counts do not add up to the rows sent")
	}
}

func TestCS11TheRowsStartedWithEightWorkersStillEqualTheirSingleRuns(t *testing.T) {
	cost := rowCost(t)
	const limitedRows = 10000
	compiled := compileOrFail(t, tickDefinition(), tickBody, newTickRuntime())
	rows := tickBatch(limitedRows)
	result := runOrFail(t, compiled, rows, batch.Options{ChunkSize: 1, Workers: 8, MaxRows: limitedRows, Cost: 25 * cost})
	counted := readLimited(result)
	if counted.other != 0 || counted.valued+counted.notRun != limitedRows {
		t.Fatalf("valued %d, not started %d, other %d", counted.valued, counted.notRun, counted.other)
	}
	if counted.valued < 25 {
		t.Fatalf("only %d rows started, at least 25 are owed to the limit", counted.valued)
	}
	for _, position := range counted.startedPositions {
		alone := compiled.Run(rows[position])
		alone.Row = position + 1
		if rowKey(alone) != rowKey(result.Rows[position]) {
			t.Fatalf("started row %d is %q, alone it is %q", position+1, rowKey(result.Rows[position]), rowKey(alone))
		}
	}
}

func TestCS12TheCallReturnsOnlyWhenTheWholeBatchIsFinishedAndLeavesNoGoroutine(t *testing.T) {
	ticks := newTickRuntime()
	compiled := compileOrFail(t, tickDefinition(), tickBody, ticks)
	const batchRows = 5000
	before := goroutineCount()
	result := runOrFail(t, compiled, tickBatch(batchRows), batch.Options{ChunkSize: 50, Workers: 8, MaxRows: batchRows})
	after := goroutineCount()
	if after > before {
		t.Fatalf("%d goroutines after the call, %d before", after, before)
	}
	if worked := ticks.calls.Load(); worked != batchRows {
		t.Fatalf("the call returned with %d of %d rows worked out", worked, batchRows)
	}
	if result.Summary.Valued != batchRows {
		t.Fatalf("the summary counts %d valued rows of %d", result.Summary.Valued, batchRows)
	}
}

func TestCS13RunningTheCalculationStageAgainChangesNothing(t *testing.T) {
	compiled := compileOrFail(t, evenShareDefinition(), evenShareBody, nil)
	rows := mixedBatch(mixedRows)
	first := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 100, Workers: 1}))
	second := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 100, Workers: 1}))
	third := listKey(runOrFail(t, compiled, rows, batch.Options{ChunkSize: 100, Workers: 8}))
	sameLists(t, second, first, "the second run")
	sameLists(t, third, first, "the run with eight workers")
}

func TestCS14ACancelledBatchIsHeld(t *testing.T) {
	t.Skip("held: whether and how the library honours a host's own cancel is open under decision 19; batch/ reads the context at each chunk edge as the design page proposes, and the pass rule waits on the owner")
}

func inputName(found problem.Problem) string {
	if found.Input == nil {
		return ""
	}
	return *found.Input
}
