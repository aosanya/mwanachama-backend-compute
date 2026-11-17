package measure

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const (
	batchRows    = 10000
	batchWorkers = 8
	batchSeed    = 20260117
)

var batchFormulas = []int{8, 3, 15}

func BenchmarkRun(b *testing.B) {
	for _, e := range engines() {
		set := mustCompiledFor(b, e)
		b.Run(e.Name(), func(b *testing.B) {
			for _, c := range harness.Cases() {
				program := set.programs[c.Number]
				inputs := c.Examples[0].Inputs
				b.Run(formulaName(c.Number), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if _, err := program.Run(inputs); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

var (
	rowsMutex sync.Mutex
	rowsCache = map[int][]map[string]any{}
)

func batchInputs(number int) []map[string]any {
	rowsMutex.Lock()
	defer rowsMutex.Unlock()
	if rows, ok := rowsCache[number]; ok {
		return rows
	}
	rows := makeInputs(number, batchSeed+int64(number), batchRows)
	rowsCache[number] = rows
	return rows
}

func runRows(program harness.Program, rows []map[string]any) error {
	for _, row := range rows {
		if _, err := program.Run(row); err != nil {
			return err
		}
	}
	return nil
}

func runRowsSplit(program harness.Program, rows []map[string]any, workers int) error {
	var failed atomic.Pointer[error]
	var group sync.WaitGroup
	size := (len(rows) + workers - 1) / workers
	for start := 0; start < len(rows); start += size {
		part := rows[start:min(start+size, len(rows))]
		group.Add(1)
		go func() {
			defer group.Done()
			if err := runRows(program, part); err != nil {
				failed.Store(&err)
			}
		}()
	}
	group.Wait()
	if err := failed.Load(); err != nil {
		return *err
	}
	return nil
}

func benchmarkBatch(b *testing.B, split func(program harness.Program, rows []map[string]any) error) {
	for _, e := range engines() {
		set := mustCompiledFor(b, e)
		b.Run(e.Name(), func(b *testing.B) {
			for _, number := range batchFormulas {
				program := set.programs[number]
				rows := batchInputs(number)
				b.Run(formulaName(number), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if err := split(program, rows); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func BenchmarkBatchSequential(b *testing.B) {
	benchmarkBatch(b, runRows)
}

func BenchmarkBatchParallel(b *testing.B) {
	benchmarkBatch(b, func(program harness.Program, rows []map[string]any) error {
		return runRowsSplit(program, rows, batchWorkers)
	})
}

func BenchmarkCompile(b *testing.B) {
	for _, e := range engines() {
		set := mustCompiledFor(b, e)
		b.Run(e.Name(), func(b *testing.B) {
			for _, c := range harness.Cases() {
				b.Run(formulaName(c.Number), func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						if _, err := e.Compile(c, set.library); err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}

func TestBatchRowsAreDeterministic(t *testing.T) {
	for _, number := range batchFormulas {
		first := makeInputs(number, batchSeed+int64(number), 50)
		second := makeInputs(number, batchSeed+int64(number), 50)
		for i := range first {
			if harness.InputKey(first[i]) != harness.InputKey(second[i]) {
				t.Fatalf("formula %d row %d differs between two generations", number, i)
			}
		}
	}
}

func TestBatchRunsAgreeAcrossEngines(t *testing.T) {
	for _, number := range batchFormulas {
		rows := makeInputs(number, batchSeed+int64(number), 500)
		var reference []any
		for _, e := range engines() {
			program := mustCompiledFor(t, e).programs[number]
			results := make([]any, len(rows))
			for i, row := range rows {
				got, err := program.Run(row)
				if err != nil {
					t.Fatalf("%s %s row %d: %v", e.Name(), formulaName(number), i, err)
				}
				results[i] = got
			}
			if reference == nil {
				reference = results
				continue
			}
			for i := range results {
				if !harness.Equal(reference[i], results[i]) {
					t.Fatalf("%s %s row %d: %s, the first engine gave %s", e.Name(), formulaName(number), i, harness.Describe(results[i]), harness.Describe(reference[i]))
				}
			}
		}
	}
}
