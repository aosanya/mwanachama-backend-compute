package concurrency

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/batch"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type speedKind string

const (
	kindRun     speedKind = "run"
	kindBatch   speedKind = "batch"
	kindCompile speedKind = "compile"
)

type speedCase struct {
	check   string
	kind    speedKind
	formula int
	workers int
	rows    int
}

func (c speedCase) name() string {
	switch c.kind {
	case kindBatch:
		return fmt.Sprintf("%s f%d rows %d workers %d", c.check, c.formula, c.rows, c.workers)
	}
	return fmt.Sprintf("%s f%d %s", c.check, c.formula, c.kind)
}

func (c speedCase) units() int {
	if c.kind == kindBatch {
		return c.rows
	}
	return 1
}

type target struct {
	ns     float64
	allocs float64
}

func (c speedCase) target() target {
	switch c.check {
	case "CS-20":
		return target{ns: 1500, allocs: 25}
	case "CS-21":
		return map[int]target{8: {ns: 225}, 3: {ns: 725}, 15: {ns: 675}}[c.formula]
	case "CS-22":
		if c.workers == 1 {
			return map[int]target{8: {ns: 225}, 3: {ns: 725}, 15: {ns: 675}}[c.formula]
		}
		return map[int]target{8: {ns: 150}, 3: {ns: 425}, 15: {ns: 400}}[c.formula]
	case "CS-23":
		return map[int]target{1: {ns: 2450, allocs: 45}, 8: {ns: 2525, allocs: 55}, 15: {ns: 2850, allocs: 60}, 3: {ns: 3850, allocs: 70}}[c.formula]
	}
	return target{}
}

const medianRunTarget = 450.0

var batchFormulas = []int{1, 8, 3, 15}

func speedCases() []speedCase {
	var cases []speedCase
	for id := 1; id <= 19; id++ {
		cases = append(cases, speedCase{check: "CS-20", kind: kindRun, formula: id})
	}
	for _, id := range batchFormulas {
		cases = append(cases, speedCase{check: "CS-21", kind: kindBatch, formula: id, workers: 1, rows: 1000})
	}
	for _, id := range batchFormulas {
		for _, workers := range []int{1, 8} {
			cases = append(cases, speedCase{check: "CS-22", kind: kindBatch, formula: id, workers: workers, rows: 100000})
		}
	}
	for _, id := range []int{1, 8, 15, 3} {
		cases = append(cases, speedCase{check: "CS-23", kind: kindCompile, formula: id})
	}
	return cases
}

func batchDefinition(id int) function.Definition {
	switch id {
	case 1:
		return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2), decimalInput("payment", 2)}}
	case 8:
		return function.Definition{Inputs: []function.InputDecl{decimalInput("balance", 2)}}
	case 3:
		return function.Definition{Inputs: []function.InputDecl{
			{Name: "deposits", Type: formula.TypeListDecimal, Rules: []function.Rule{{Name: problem.RuleRequired}}},
			{Name: "expenses", Type: formula.TypeListDecimal},
			integerInput("members"),
		}}
	case 15:
		return function.Definition{Inputs: []function.InputDecl{
			decimalInput("call_amount", 2),
			integerInput("units"),
			{Name: "all_units", Type: formula.TypeListInteger, Rules: []function.Rule{{Name: problem.RuleRequired}}},
		}}
	}
	panic(fmt.Sprintf("no batch definition for formula %d", id))
}

func batchRowFor(id int, i int) function.Row {
	switch id {
	case 1:
		return row("balance", cents(10000+37*i%900000), "payment", cents(500+11*i%700000))
	case 8:
		return row("balance", cents(10000+37*i%900000))
	case 3:
		return function.Row{
			"deposits": numtext.Many("500.00", "750.00", cents(100000+i%90000)),
			"expenses": numtext.Many("120.25", "80.00"),
			"members":  numtext.One(fmt.Sprint(1 + i%500)),
		}
	case 15:
		return function.Row{
			"call_amount": numtext.One(cents(500000 + 13*i%9000000)),
			"units":       numtext.One(fmt.Sprint(1 + i%2000)),
			"all_units":   numtext.Many("1200", "830", "1500", "477"),
		}
	}
	panic(fmt.Sprintf("no batch row for formula %d", id))
}

var (
	rowsMu    sync.Mutex
	rowsCache = map[string][]function.Row{}
)

func batchRows(id int, count int) []function.Row {
	rowsMu.Lock()
	defer rowsMu.Unlock()
	key := fmt.Sprintf("%d/%d", id, count)
	if rows, found := rowsCache[key]; found {
		return rows
	}
	rows := make([]function.Row, count)
	for i := range rows {
		rows[i] = batchRowFor(id, i)
	}
	rowsCache[key] = rows
	return rows
}

type prepared struct {
	units int
	step  func() error
}

func prepare(t testing.TB, c speedCase) prepared {
	t.Helper()
	runtime := newAgencyRuntime()
	example := formulaNamed(c.formula)
	switch c.kind {
	case kindRun:
		program := planCase(t, example, runtime)
		values := example.values()
		return prepared{units: 1, step: func() error {
			if _, failure := program.Run(values); failure != nil {
				return fmt.Errorf("%s: %s", failure.Code, failure.Sentence)
			}
			return nil
		}}
	case kindCompile:
		inputs := example.inputs()
		return prepared{units: 1, step: func() error {
			tree, problems := formula.Parse(example.body)
			if len(problems) > 0 {
				return fmt.Errorf("%s", problems[0].Sentence)
			}
			if _, problems = formula.Plan(tree, inputs, runtime, formula.DefaultOptions()); len(problems) > 0 {
				return fmt.Errorf("%s", problems[0].Sentence)
			}
			return nil
		}}
	}
	compiled, problems := batch.Compile(batchDefinition(c.formula), example.body, runtime, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("compile formula %d for the batch: %s", c.formula, problems[0].Sentence)
	}
	rows := batchRows(c.formula, c.rows)
	options := batch.Options{ChunkSize: 100, Workers: c.workers, MaxRows: c.rows}
	return prepared{units: c.rows, step: func() error {
		result, err := compiled.RunBatch(context.Background(), rows, options)
		if err != nil {
			return err
		}
		if result.Summary.Valued != c.rows {
			return fmt.Errorf("%d of %d rows have a value", result.Summary.Valued, c.rows)
		}
		return nil
	}}
}

func benchmarkOf(c speedCase) func(b *testing.B) {
	return func(b *testing.B) {
		p := prepare(b, c)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := p.step(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkSpeed(b *testing.B) {
	for _, c := range speedCases() {
		b.Run(c.name(), benchmarkOf(c))
	}
}

func TestEverySpeedCaseRunsOnceAndWorks(t *testing.T) {
	for _, c := range speedCases() {
		c := c
		if c.rows > 1000 {
			c.rows = 1000
		}
		t.Run(c.name(), func(t *testing.T) {
			if err := prepare(t, c).step(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTheBatchFormulasGiveTheirValuesThroughTheBatchPath(t *testing.T) {
	runtime := newAgencyRuntime()
	for _, id := range batchFormulas {
		compiled, problems := batch.Compile(batchDefinition(id), formulaNamed(id).body, runtime, formula.DefaultOptions())
		if len(problems) > 0 {
			t.Fatalf("formula %d: %s", id, problems[0].Sentence)
		}
		result, err := compiled.RunBatch(context.Background(), batchRows(id, 50), batch.Options{Workers: 4})
		if err != nil {
			t.Fatal(err)
		}
		if result.Summary.Valued != 50 {
			t.Fatalf("formula %d: %d of 50 rows have a value, first problems %v", id, result.Summary.Valued, result.Rows[0].Problems)
		}
	}
}
