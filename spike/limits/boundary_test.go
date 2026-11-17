package limits

import (
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func accepts(engine harness.Engine, source string, want int64) bool {
	got, err := compileRun(engine, source, nil, nil, harness.KindInteger)
	return err == nil && got == want
}

func TestNativeBoundaries(t *testing.T) {
	limits := Proposed()
	for _, f := range factories {
		engine := f.build(limits, false)
		deepest := 0
		for k := 1; k <= limits.Depth*2; k++ {
			if accepts(engine, nestedMinus(k), nestedMinusValue(k)) {
				deepest = k
			}
		}
		firstRefusedDepth := deepest + 1
		summary := fmt.Sprintf("native depth: nested minus accepted up to %d brackets (%d levels counting the number), limit %d", deepest, deepest+1, limits.Depth)
		if deepest == limits.Depth*2 {
			summary = fmt.Sprintf("native depth: no refusal found; every nesting tried up to %d brackets was accepted, limit %d", deepest, limits.Depth)
		}
		record(tableRow{Case: "L01b", Engine: f.name, Mode: "native", Layer: string(LayerNative), Time: "-", Sentence: summary})
		t.Log(f.name + ": " + summary)
		if f.name == "own" && deepest+1 != limits.Depth {
			t.Errorf("own depth boundary is %d, want %d", deepest+1, limits.Depth)
		}
		_ = firstRefusedDepth

		largest := 0
		for n := limits.Nodes - 40; n <= limits.Nodes+40; n++ {
			text, value := formulaWithNodes(n)
			if accepts(engine, text, value) {
				largest = n
			}
		}
		summary = fmt.Sprintf("native nodes: formulas of balanced additions accepted up to %d nodes by this engine's own count, limit %d", largest, limits.Nodes)
		record(tableRow{Case: "L03b", Engine: f.name, Mode: "native", Layer: string(LayerNative), Time: "-", Sentence: summary})
		t.Log(f.name + ": " + summary)
		if f.name == "own" && largest != limits.Nodes {
			t.Errorf("own node boundary is %d, want %d", largest, limits.Nodes)
		}
	}
}
