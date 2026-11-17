package concurrency

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	baselineFile     = "baselines.json"
	defaultBenchtime = "300ms"
)

type recorded struct {
	Machine   machine           `json:"machine"`
	Date      string            `json:"date"`
	Load      float64           `json:"load"`
	Quiet     bool              `json:"quiet"`
	Benchtime string            `json:"benchtime"`
	Figures   map[string]figure `json:"figures"`
}

type baselines map[string]recorded

func readBaselines(t testing.TB) baselines {
	t.Helper()
	data, err := os.ReadFile(baselineFile)
	if os.IsNotExist(err) {
		return baselines{}
	}
	if err != nil {
		t.Fatalf("read %s: %v", baselineFile, err)
	}
	var stored baselines
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("parse %s: %v", baselineFile, err)
	}
	return stored
}

func writeBaselines(t testing.TB, stored baselines) {
	t.Helper()
	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(baselineFile, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", baselineFile, err)
	}
}

func benchtime() string {
	if chosen := os.Getenv("SPEED_BENCHTIME"); chosen != "" {
		return chosen
	}
	return defaultBenchtime
}

func measureWith(t testing.TB, run func(b *testing.B), units int) figure {
	t.Helper()
	var ns, bytes, allocs []float64
	for i := 0; i < countOfRuns; i++ {
		result := testing.Benchmark(run)
		if result.N == 0 {
			t.Fatal("a benchmark failed to run")
		}
		perUnit := float64(result.N) * float64(units)
		ns = append(ns, float64(result.T.Nanoseconds())/perUnit)
		bytes = append(bytes, float64(result.MemBytes)/perUnit)
		allocs = append(allocs, float64(result.MemAllocs)/perUnit)
	}
	return figure{Ns: median(ns), Bytes: median(bytes), Allocs: median(allocs), Spread: spreadOf(ns)}
}

func measureCase(t testing.TB, c speedCase) figure {
	t.Helper()
	first := measureWith(t, benchmarkOf(c), c.units())
	if first.Spread <= spreadLimit {
		return first
	}
	again := measureWith(t, benchmarkOf(c), c.units())
	if again.Spread < first.Spread {
		return again
	}
	return first
}

func measureAll(t *testing.T) map[string]figure {
	t.Helper()
	if err := flag.Set("test.benchtime", benchtime()); err != nil {
		t.Fatalf("set the benchmark time: %v", err)
	}
	figures := map[string]figure{}
	for _, c := range speedCases() {
		figures[c.name()] = measureCase(t, c)
	}
	return figures
}

type gate struct {
	machine machine
	load    float64
	known   bool
	quiet   bool
}

func speedGate(t *testing.T) gate {
	t.Helper()
	if os.Getenv("SPEED_CHECK") == "" {
		t.Skip("the speed check measures time and runs on its own: make speed-check")
	}
	current := currentMachine()
	load, known := oneMinuteLoad()
	return gate{machine: current, load: load, known: known, quiet: known && load <= quietLimit(current.CPUs)}
}

func (g gate) refuse(t *testing.T) {
	t.Helper()
	if g.quiet || os.Getenv("SPEED_FORCE") != "" {
		return
	}
	if !g.known {
		t.Skip("refused to judge: the load average of this machine cannot be read")
	}
	t.Skipf("refused to judge: the one-minute load average is %.1f and the check judges only at %.1f or lower (half of %d cpus); set SPEED_FORCE=1 to measure anyway", g.load, quietLimit(g.machine.CPUs), g.machine.CPUs)
}

func logFigures(t *testing.T, figures map[string]figure) {
	t.Helper()
	var b strings.Builder
	b.WriteString("\nname | ns per unit | bytes per unit | allocations per unit | spread\n")
	for _, c := range speedCases() {
		f := figures[c.name()]
		fmt.Fprintf(&b, "%s | %.1f | %.1f | %.2f | %.1f%%\n", c.name(), f.Ns, f.Bytes, f.Allocs, f.Spread*100)
	}
	t.Log(b.String())
}

func TestSpeedRecordTheBaselineOfThisMachine(t *testing.T) {
	if os.Getenv("SPEED_RECORD") == "" {
		t.Skip("run only to record a baseline: make baseline")
	}
	g := speedGate(t)
	g.refuse(t)
	figures := measureAll(t)
	logFigures(t, figures)
	stored := readBaselines(t)
	stored[g.machine.key()] = recorded{
		Machine:   g.machine,
		Date:      time.Now().Format("2006-01-02"),
		Load:      g.load,
		Quiet:     g.quiet,
		Benchtime: benchtime(),
		Figures:   figures,
	}
	writeBaselines(t, stored)
	t.Logf("recorded %d figures for %s at load %.1f, quiet %v", len(figures), g.machine.key(), g.load, g.quiet)
}

func TestSpeedCheckAgainstTheBaselineOfThisMachine(t *testing.T) {
	if os.Getenv("SPEED_RECORD") != "" {
		t.Skip("recording a baseline, not judging")
	}
	g := speedGate(t)
	stored, found := readBaselines(t)[g.machine.key()]
	if !found {
		t.Skipf("no baseline for %s; a figure is never compared across machines. Run make baseline on a quiet machine", g.machine.key())
	}
	if !stored.Quiet {
		t.Skipf("the baseline of this machine was recorded at load %.1f on %s, above the quiet limit; run make baseline again on a quiet machine", stored.Load, stored.Date)
	}
	g.refuse(t)
	figures := measureAll(t)
	logFigures(t, figures)
	for _, c := range speedCases() {
		baseline, has := stored.Figures[c.name()]
		if !has {
			t.Errorf("the baseline has no figure for %s; run make baseline", c.name())
			continue
		}
		for _, f := range judgeSlip(c.name(), baseline, figures[c.name()]) {
			t.Errorf("%s (machine %s)", f, g.machine.key())
		}
	}
}

func TestSpeedTargetsOfTheProposedPage(t *testing.T) {
	g := speedGate(t)
	g.refuse(t)
	figures := measureAll(t)
	logFigures(t, figures)
	for _, c := range speedCases() {
		for _, f := range judgeTarget(c, figures[c.name()]) {
			t.Errorf("%s (machine %s, load %.1f)", f, g.machine.key(), g.load)
		}
	}
	for _, f := range judgeMedianRun(figures) {
		t.Errorf("%s (machine %s, load %.1f)", f, g.machine.key(), g.load)
	}
}

func slowedCopy(c speedCase) func(b *testing.B) {
	return func(b *testing.B) {
		p := prepare(b, c)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for repeat := 0; repeat < 3; repeat++ {
				if err := p.step(); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

func TestSpeedCheckFailsOnADeliberateSlowdownInACopy(t *testing.T) {
	if os.Getenv("SPEED_RECORD") != "" {
		t.Skip("recording a baseline, not judging")
	}
	g := speedGate(t)
	g.refuse(t)
	if err := flag.Set("test.benchtime", benchtime()); err != nil {
		t.Fatal(err)
	}
	c := speedCase{check: "CS-20", kind: kindRun, formula: 3}
	baseline := measureWith(t, benchmarkOf(c), 1)
	slowed := measureWith(t, slowedCopy(c), 1)
	found := judgeSlip(c.name(), baseline, slowed)
	t.Logf("baseline %.0f ns %.0f allocs, slowed copy %.0f ns %.0f allocs, findings %v", baseline.Ns, baseline.Allocs, slowed.Ns, slowed.Allocs, found)
	if len(found) == 0 {
		t.Fatal("a copy that runs three times as much work passed the speed check")
	}
}
