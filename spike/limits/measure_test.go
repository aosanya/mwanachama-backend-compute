package limits

import (
	"fmt"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/spike/engine/own"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

func TestLargestAmongTheNineteenFormulas(t *testing.T) {
	deepest, biggest, longest := 0, 0, 0
	for _, c := range harness.Cases() {
		tree, err := own.Parse(c.Source)
		if err != nil {
			t.Fatal(err)
		}
		deepest = max(deepest, own.Depth(tree))
		biggest = max(biggest, own.Count(tree))
		longest = max(longest, len([]rune(c.Source)))
	}
	summary := fmt.Sprintf("largest among the 19 by the own engine's count: depth %d, nodes %d, formula length %d characters", deepest, biggest, longest)
	record(tableRow{Case: "F19", Engine: "own", Mode: "native", Layer: string(LayerNone), Time: "-", Sentence: summary})
	t.Log(summary)
	limits := Proposed()
	if deepest*4 > limits.Depth || biggest*10 > limits.Nodes || longest*10 > limits.FormulaLength {
		t.Errorf("the proposed limits leave too little room: %s", summary)
	}
}

func smallestPassing(f engineFactory, set func(*Limits, int)) int {
	for n := 1; n <= 400; n++ {
		limits := Proposed()
		set(&limits, n)
		if harness.NewReport(f.build(limits, false)).AllPassed() {
			return n
		}
	}
	return -1
}

func TestSmallestNativeLimitsThatPassTheNineteenFormulas(t *testing.T) {
	for _, f := range factories {
		depth := smallestPassing(f, func(l *Limits, n int) { l.Depth = n })
		nodes := smallestPassing(f, func(l *Limits, n int) { l.Nodes = n })
		summary := fmt.Sprintf("smallest depth limit that lets all 19 pass natively: %d; smallest node limit: %d", depth, nodes)
		if f.name == "expr-lang" {
			summary = fmt.Sprintf("no native depth limit (the depth setting changes nothing); smallest node limit that lets all 19 pass: %d", nodes)
		}
		record(tableRow{Case: "F19", Engine: f.name, Mode: "native", Layer: string(LayerNative), Time: "-", Sentence: summary})
		t.Log(f.name + ": " + summary)
		if depth < 0 || nodes < 0 || depth*2 > Proposed().Depth || nodes*4 > Proposed().Nodes {
			t.Errorf("%s: %s", f.name, summary)
		}
	}
}
