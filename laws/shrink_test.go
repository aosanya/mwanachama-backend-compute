package laws

import (
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

func containsCall(tree formula.Tree, name string) bool {
	found := false
	walk(tree.Root, func(n formula.Node) {
		if call, isCall := n.(formula.Call); isCall && call.Name == name {
			found = true
		}
	})
	return found
}

func TestShrinkReachesTheSmallestTreeThatStillFails(t *testing.T) {
	engine := RealEngine()
	shrunkCases := 0
	for seed := int64(1); seed <= 300; seed++ {
		c := NewCase(seed)
		if !containsCall(c.Tree, "round") {
			continue
		}
		fails := func(tree formula.Tree) bool { return containsCall(tree, "round") }
		valid := func(tree formula.Tree) bool { return len(engine.Check(tree, c.Inputs)) == 0 }
		small := Shrink(c.Tree, valid, fails)
		if !fails(small) || !valid(small) {
			t.Fatalf("seed %d: shrunk tree %s no longer fails or is not valid", seed, Show(small))
		}
		if TreeSize(small) != 3 {
			t.Fatalf("seed %d: shrunk to %s with %d nodes, want a round call with two arguments", seed, Show(small), TreeSize(small))
		}
		if TreeSize(small) < TreeSize(c.Tree) {
			shrunkCases++
		}
	}
	if shrunkCases == 0 {
		t.Fatal("no generated tree got smaller")
	}
}

func TestShrinkLeavesATreeAloneWhenNothingSmallerFails(t *testing.T) {
	engine := RealEngine()
	c := NewCase(7)
	valid := func(tree formula.Tree) bool { return len(engine.Check(tree, c.Inputs)) == 0 }
	same := func(tree formula.Tree) bool { return TreeSize(tree) == TreeSize(c.Tree) }
	if got := Shrink(c.Tree, valid, same); Show(got) != Show(c.Tree) {
		t.Fatalf("shrunk %s to %s though nothing smaller fails", Show(c.Tree), Show(got))
	}
}

func TestShowBracketsEveryOperatorSoGroupingIsVisible(t *testing.T) {
	left := formula.Binary{Op: "-", Left: formula.Binary{Op: "-", Left: formula.Name{Name: "a"}, Right: formula.Name{Name: "b"}}, Right: formula.Name{Name: "c"}}
	right := formula.Binary{Op: "-", Left: formula.Name{Name: "a"}, Right: formula.Binary{Op: "-", Left: formula.Name{Name: "b"}, Right: formula.Name{Name: "c"}}}
	if got := Show(formula.Tree{Root: left}); got != "(a - b) - c" {
		t.Fatalf("left grouping shown as %q", got)
	}
	if got := Show(formula.Tree{Root: right}); got != "a - (b - c)" {
		t.Fatalf("right grouping shown as %q", got)
	}
}
