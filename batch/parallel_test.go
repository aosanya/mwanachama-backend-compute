package batch

import (
	"context"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/function"
)

func mixedRows(count int) []function.Row {
	rows := make([]function.Row, count)
	for i := range rows {
		switch i % 5 {
		case 0:
			rows[i] = row("total", cents(1000*(i+1)), "people", "3")
		case 1:
			rows[i] = row("total", cents(700*(i+1)), "people", "0")
		case 2:
			rows[i] = row("total", "-1.00", "people", "4")
		case 3:
			rows[i] = row("total", "12.345", "people", "7")
		default:
			rows[i] = row("total", cents(333*(i+1)), "people", "7")
		}
	}
	return rows
}

func TestWhateverTheChunkSizeAndWorkersTheResultsAreTheSame(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := mixedRows(250)
	reference, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 100, Workers: 1})
	requireNoError(t, err)
	requireRowNumbers(t, reference, 250)
	requireEqualsSingle(t, compiled, rows, nil, reference)
	requireSummaryMatches(t, reference)
	for _, chunk := range []int{1, 7, 250, 251} {
		for _, workers := range []int{1, 4, 64} {
			result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: chunk, Workers: workers})
			requireNoError(t, err)
			if !reflect.DeepEqual(result.Rows, reference.Rows) {
				t.Fatalf("chunk size %d with %d workers gave other rows than chunk size 100 with 1 worker", chunk, workers)
			}
			if result.Summary.ChunkSize != chunk || result.Summary.Workers != workers {
				t.Fatalf("summary says chunk size %d workers %d, used %d and %d", result.Summary.ChunkSize, result.Summary.Workers, chunk, workers)
			}
		}
	}
}

func TestChunksThatFinishLastDoNotMoveTheirRows(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	balances := []string{"80.00", "10.00", "60.00", "30.00", "50.00", "20.00", "70.00", "40.00"}
	want := []string{"0.40", "0.05", "0.30", "0.15", "0.25", "0.10", "0.35", "0.20"}
	rows := make([]function.Row, len(balances))
	for i, balance := range balances {
		rows[i] = row("balance", balance, "rate", "0.005")
	}
	release := make(chan struct{})
	var finished int32
	options := Options{ChunkSize: 2, Workers: 4,
		beforeChunk: func(index int) {
			if index == 1 {
				<-release
			}
		},
		afterChunk: func(index int) {
			if index != 1 && atomic.AddInt32(&finished, 1) == 3 {
				close(release)
			}
		},
	}
	result, err := compiled.RunBatch(context.Background(), rows, options)
	requireNoError(t, err)
	requireRowNumbers(t, result, 8)
	if got := answersOf(result); !reflect.DeepEqual(got, want) {
		t.Fatalf("answers %v, want %v", got, want)
	}
}

func TestTenThousandRowsComeBackOnceEachAndInOrderWithFourWorkers(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	result, err := compiled.RunBatch(context.Background(), monthEndRows(10000), Options{ChunkSize: 100, Workers: 4})
	requireNoError(t, err)
	requireRowNumbers(t, result, 10000)
	for i, found := range result.Rows {
		if got := valueText(t, found); got != monthEndAnswer(i+1) {
			t.Fatalf("row %d gave %s, want %s", i+1, got, monthEndAnswer(i+1))
		}
	}
}

func TestNoChunkIsWorkedTwiceOrLeftOut(t *testing.T) {
	compiled := mustCompile(t, monthEnd(), monthEndBody)
	var mu sync.Mutex
	started := map[int]int{}
	options := Options{ChunkSize: 7, Workers: 8, beforeChunk: func(index int) {
		mu.Lock()
		started[index]++
		mu.Unlock()
	}}
	_, err := compiled.RunBatch(context.Background(), monthEndRows(100), options)
	requireNoError(t, err)
	if len(started) != 15 {
		t.Fatalf("%d chunks started, want 15", len(started))
	}
	for index, times := range started {
		if times != 1 || index < 1 || index > 15 {
			t.Fatalf("chunk %d started %d times", index, times)
		}
	}
}

func TestOneCompiledFormulaServesManyCallsAtOnce(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := mixedRows(120)
	reference, err := compiled.RunBatch(context.Background(), rows, Options{})
	requireNoError(t, err)
	var wait sync.WaitGroup
	failures := make(chan string, 16)
	for caller := 0; caller < 16; caller++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 5, Workers: 4})
			if err != nil || !reflect.DeepEqual(result.Rows, reference.Rows) {
				failures <- "a concurrent call gave other rows"
			}
		}()
	}
	wait.Wait()
	close(failures)
	for failure := range failures {
		t.Fatal(failure)
	}
}

func TestRowsOfABatchShareNoState(t *testing.T) {
	compiled := mustCompile(t, evenShare(), evenShareBody)
	rows := []function.Row{
		row("total", "10.00", "people", "0"),
		row("total", "90.00", "people", "4"),
		row("total", "10.00", "people", "0"),
		row("total", "100.00", "people", "3"),
	}
	first, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 1, Workers: 4})
	requireNoError(t, err)
	second, err := compiled.RunBatch(context.Background(), rows, Options{ChunkSize: 4, Workers: 1})
	requireNoError(t, err)
	if !reflect.DeepEqual(first.Rows, second.Rows) {
		t.Fatal("the same rows gave other answers on a second call")
	}
	if first.Rows[1].Value.String() != "22.50" || first.Rows[3].Value.String() != "33.33" {
		t.Fatalf("answers %v", answersOf(first))
	}
}

func TestTheInputRowsAreNotChanged(t *testing.T) {
	compiled := mustCompile(t, loanBalance(), loanBalanceBody)
	rows := []function.Row{row("balance", "120.50"), row("balance", "60.00", "payment", "99.00")}
	shared := row("payment", "10.00")
	before := []function.Row{row("balance", "120.50"), row("balance", "60.00", "payment", "99.00")}
	_, err := compiled.RunBatch(context.Background(), rows, Options{Shared: shared, Workers: 2, ChunkSize: 1})
	requireNoError(t, err)
	if !reflect.DeepEqual(rows, before) || len(shared) != 1 {
		t.Fatalf("the call changed its inputs: rows %v shared %v", rows, shared)
	}
}
