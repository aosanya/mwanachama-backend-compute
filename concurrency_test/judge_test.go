package concurrency

import (
	"fmt"
	"sort"
	"testing"
)

const (
	timeMargin   = 0.25
	memoryMargin = 0.10
	spreadLimit  = 0.25
	countOfRuns  = 5
)

type figure struct {
	Ns     float64 `json:"ns"`
	Bytes  float64 `json:"bytes"`
	Allocs float64 `json:"allocs"`
	Spread float64 `json:"spread"`
}

type finding struct {
	kind     string
	name     string
	what     string
	limit    float64
	measured float64
}

func (f finding) String() string {
	return fmt.Sprintf("%s: %s, %s is %.1f against %.1f", f.kind, f.name, f.what, f.measured, f.limit)
}

func median(values []float64) float64 {
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[middle]
	}
	return (sorted[middle-1] + sorted[middle]) / 2
}

func spreadOf(values []float64) float64 {
	middle := median(values)
	if middle == 0 {
		return 0
	}
	least, most := values[0], values[0]
	for _, v := range values {
		least, most = min(least, v), max(most, v)
	}
	return (most - least) / middle
}

func judgeSlip(name string, baseline figure, measured figure) []finding {
	var found []finding
	if limit := baseline.Ns * (1 + timeMargin); measured.Ns > limit {
		found = append(found, finding{"slip", name, "ns per unit", limit, measured.Ns})
	}
	if limit := baseline.Bytes * (1 + memoryMargin); measured.Bytes > limit {
		found = append(found, finding{"slip", name, "bytes per unit", limit, measured.Bytes})
	}
	if limit := baseline.Allocs * (1 + memoryMargin); measured.Allocs > limit {
		found = append(found, finding{"slip", name, "allocations per unit", limit, measured.Allocs})
	}
	return found
}

func judgeTarget(c speedCase, measured figure) []finding {
	goal := c.target()
	var found []finding
	if goal.ns > 0 {
		if limit := goal.ns * (1 + timeMargin); measured.Ns > limit {
			found = append(found, finding{"target", c.name(), "ns per unit", limit, measured.Ns})
		}
	}
	if goal.allocs > 0 {
		if limit := goal.allocs * (1 + memoryMargin); measured.Allocs > limit {
			found = append(found, finding{"target", c.name(), "allocations per unit", limit, measured.Allocs})
		}
	}
	return found
}

func judgeMedianRun(figures map[string]figure) []finding {
	var runs []float64
	for _, c := range speedCases() {
		if c.check == "CS-20" {
			runs = append(runs, figures[c.name()].Ns)
		}
	}
	if len(runs) == 0 {
		return nil
	}
	if middle, limit := median(runs), medianRunTarget*(1+timeMargin); middle > limit {
		return []finding{{"target", "CS-20 median of the 19 formulas", "ns per run", limit, middle}}
	}
	return nil
}

func TestASlowdownOverTheMarginFailsTheCheckAndASmallOneDoesNot(t *testing.T) {
	baseline := figure{Ns: 1000, Bytes: 2000, Allocs: 40}
	cases := []struct {
		name     string
		measured figure
		kinds    int
	}{
		{"the same figures", figure{Ns: 1000, Bytes: 2000, Allocs: 40}, 0},
		{"a rise inside both margins", figure{Ns: 1240, Bytes: 2190, Allocs: 43}, 0},
		{"faster and smaller", figure{Ns: 600, Bytes: 1000, Allocs: 20}, 0},
		{"time up 26 percent", figure{Ns: 1260, Bytes: 2000, Allocs: 40}, 1},
		{"time doubled", figure{Ns: 2000, Bytes: 2000, Allocs: 40}, 1},
		{"bytes up 11 percent", figure{Ns: 1000, Bytes: 2220, Allocs: 40}, 1},
		{"allocations up 11 percent", figure{Ns: 1000, Bytes: 2000, Allocs: 44.5}, 1},
		{"everything worse", figure{Ns: 3000, Bytes: 5000, Allocs: 90}, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := judgeSlip("sample", baseline, c.measured); len(got) != c.kinds {
				t.Fatalf("got %d findings %v, want %d", len(got), got, c.kinds)
			}
		})
	}
}

func TestATargetIsMissedOnlyBeyondItsMargin(t *testing.T) {
	c := speedCase{check: "CS-21", kind: kindBatch, formula: 8, workers: 1, rows: 1000}
	if got := judgeTarget(c, figure{Ns: 225 * 1.25}); len(got) != 0 {
		t.Fatalf("exactly at the margin must pass, got %v", got)
	}
	if got := judgeTarget(c, figure{Ns: 225*1.25 + 1}); len(got) != 1 {
		t.Fatalf("just over the margin must fail, got %v", got)
	}
	withoutTarget := speedCase{check: "CS-21", kind: kindBatch, formula: 1, workers: 1, rows: 1000}
	if got := judgeTarget(withoutTarget, figure{Ns: 1e9}); len(got) != 0 {
		t.Fatalf("formula 1 has no batch target, got %v", got)
	}
	run := speedCase{check: "CS-20", kind: kindRun, formula: 3}
	if got := judgeTarget(run, figure{Ns: 100, Allocs: 27.5}); len(got) != 0 {
		t.Fatalf("27.5 allocations is inside 25 plus 10 percent, got %v", got)
	}
	if got := judgeTarget(run, figure{Ns: 100, Allocs: 27.6}); len(got) != 1 {
		t.Fatalf("27.6 allocations is over 25 plus 10 percent, got %v", got)
	}
}

func TestTheMedianTargetIsReadOverTheNineteenFormulas(t *testing.T) {
	figures := map[string]figure{}
	for _, c := range speedCases() {
		if c.check == "CS-20" {
			figures[c.name()] = figure{Ns: 400}
		}
	}
	if got := judgeMedianRun(figures); len(got) != 0 {
		t.Fatalf("a median of 400 is inside 450, got %v", got)
	}
	for _, c := range speedCases() {
		if c.check == "CS-20" {
			figures[c.name()] = figure{Ns: 600}
		}
	}
	if got := judgeMedianRun(figures); len(got) != 1 {
		t.Fatalf("a median of 600 is over 450 plus 25 percent, got %v", got)
	}
}

func TestMedianAndSpreadOfFiveCounts(t *testing.T) {
	values := []float64{100, 120, 90, 110, 105}
	if got := median(values); got != 105 {
		t.Fatalf("median = %v, want 105", got)
	}
	if got := spreadOf(values); got < 0.28 || got > 0.29 {
		t.Fatalf("spread = %v, want (120-90)/105", got)
	}
}
