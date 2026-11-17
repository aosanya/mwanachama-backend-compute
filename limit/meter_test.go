package limit

import "testing"

type testNode struct {
	role     Role
	items    int
	callee   int
	children []Node
}

func (n testNode) Children() []Node { return n.children }
func (n testNode) Role() Role       { return n.role }
func (n testNode) Items() int       { return n.items }
func (n testNode) CalleeCost() int  { return n.callee }

func leaf() Node { return testNode{role: RolePlain} }

func listInput(declaredMax int) Node { return testNode{role: RolePlain, items: declaredMax} }

func listLiteral(count int) Node {
	children := make([]Node, count)
	for i := range children {
		children[i] = leaf()
	}
	return testNode{role: RoleListLiteral, items: count, children: children}
}

func aggregateOver(list Node) Node {
	return testNode{role: RoleAggregate, children: []Node{list}}
}

func callCosting(callee int) Node { return testNode{role: RoleCall, callee: callee} }

func chainOfDepth(depth int) Node {
	node := leaf()
	for i := 1; i < depth; i++ {
		node = testNode{role: RolePlain, children: []Node{node}}
	}
	return node
}

func rootWithLeaves(nodes int, special ...Node) Node {
	children := append([]Node{}, special...)
	total := 1
	for _, child := range children {
		total += NodeCount(child)
	}
	for total < nodes {
		children = append(children, leaf())
		total++
	}
	return testNode{role: RolePlain, children: children}
}

func TestEachNodeCostsOne(t *testing.T) {
	tree := rootWithLeaves(34)
	if got := Proposed().CostOf(tree); got != 34 {
		t.Fatalf("cost %d, want 34", got)
	}
	if got := NodeCount(tree); got != 34 {
		t.Fatalf("nodes %d, want 34", got)
	}
}

func TestACallAddsTheCalleesDeclaredCost(t *testing.T) {
	tree := rootWithLeaves(34, callCosting(8), callCosting(8), callCosting(8))
	if got := NodeCount(tree); got != 34 {
		t.Fatalf("nodes %d, want 34", got)
	}
	if got := Proposed().CostOf(tree); got != 58 {
		t.Fatalf("cost %d, want 58, the margin table's late-filing penalty", got)
	}
}

func TestAnAggregateOverAListInputCountsTheDefaultListSize(t *testing.T) {
	tree := rootWithLeaves(19,
		aggregateOver(listInput(0)),
		aggregateOver(listInput(0)),
		aggregateOver(listInput(0)),
	)
	if got := NodeCount(tree); got != 19 {
		t.Fatalf("nodes %d, want 19", got)
	}
	if got := Proposed().CostOf(tree); got != 3019 {
		t.Fatalf("cost %d, want 3019, the margin table's term grade", got)
	}
}

func TestAnAggregateOverAWrittenListCountsItsItems(t *testing.T) {
	tree := aggregateOver(listLiteral(5))
	nodes := NodeCount(tree)
	if nodes != 7 {
		t.Fatalf("nodes %d, want 7", nodes)
	}
	if got := Proposed().CostOf(tree); got != 7+5 {
		t.Fatalf("cost %d, want 12", got)
	}
}

func TestADeclaredMaxItemsReplacesTheDefaultWhenSmaller(t *testing.T) {
	small := aggregateOver(listInput(40))
	if got := Proposed().CostOf(small); got != 2+40 {
		t.Fatalf("cost %d, want 42", got)
	}
	large := aggregateOver(listInput(5000))
	if got := Proposed().CostOf(large); got != 2+1000 {
		t.Fatalf("cost %d, want 1002, the default list size caps a larger declaration", got)
	}
}

func TestTheDefaultListSizeFollowsTheSetting(t *testing.T) {
	settings := Proposed()
	settings.ListItems = 200
	tree := aggregateOver(listInput(0))
	if got := settings.CostOf(tree); got != 2+200 {
		t.Fatalf("cost %d, want 202", got)
	}
}

func TestOnlyTheListArgumentOfAnAggregateIsCountedOnce(t *testing.T) {
	tree := testNode{role: RoleAggregate, children: []Node{listInput(0), leaf()}}
	if got := Proposed().CostOf(tree); got != 3+1000 {
		t.Fatalf("cost %d, want 1003", got)
	}
}

func TestTheSameTreeGivesTheSameCostOnEveryRun(t *testing.T) {
	tree := rootWithLeaves(40, callCosting(8), aggregateOver(listInput(0)), aggregateOver(listLiteral(3)))
	first := Proposed().CostOf(tree)
	for i := 0; i < 200; i++ {
		if got := Proposed().CostOf(tree); got != first {
			t.Fatalf("run %d gave %d, first gave %d", i, got, first)
		}
	}
}

func TestDepthCountsTheRootAsOne(t *testing.T) {
	cases := map[int]int{1: 1, 6: 6, 12: 12, 13: 13}
	for depth, want := range cases {
		if got := TreeDepth(chainOfDepth(depth)); got != want {
			t.Fatalf("depth of a chain of %d is %d, want %d", depth, got, want)
		}
	}
}

func TestATreeAtTheDepthLimitPassesAndOneDeeperIsRefused(t *testing.T) {
	settings := Proposed()
	at := Measure{Depth: TreeDepth(chainOfDepth(12)), Nodes: NodeCount(chainOfDepth(12))}
	if refusal, refused := settings.Refusal(at); refused {
		t.Fatalf("depth 12 refused: %s", refusal.Sentence())
	}
	over := Measure{Depth: TreeDepth(chainOfDepth(13)), Nodes: NodeCount(chainOfDepth(13))}
	refusal, refused := settings.Refusal(over)
	if !refused || refusal.Sentence() != "This formula goes over the depth limit of 12." {
		t.Fatalf("got %+v refused %v", refusal, refused)
	}
}

func TestATreeAtTheNodeLimitPassesAndOneMoreIsRefused(t *testing.T) {
	settings := Proposed()
	at := rootWithLeaves(100)
	if refusal, refused := settings.Refusal(Measure{Nodes: NodeCount(at), Cost: settings.CostOf(at)}); refused {
		t.Fatalf("100 nodes refused: %s", refusal.Sentence())
	}
	over := rootWithLeaves(101)
	refusal, refused := settings.Refusal(Measure{Nodes: NodeCount(over), Cost: settings.CostOf(over)})
	if !refused || refusal.Sentence() != "This formula goes over the node limit of 100." {
		t.Fatalf("got %+v refused %v", refusal, refused)
	}
}

func TestATreeWhoseCostIsPastTheCostLimitIsRefusedWithTheCostSentence(t *testing.T) {
	settings := Proposed()
	aggregates := make([]Node, 21)
	for i := range aggregates {
		aggregates[i] = aggregateOver(listInput(0))
	}
	tree := rootWithLeaves(0, aggregates...)
	cost := settings.CostOf(tree)
	if cost <= 20000 {
		t.Fatalf("cost %d, want past 20000", cost)
	}
	refusal, refused := settings.Refusal(Measure{Cost: cost})
	if !refused || refusal.Sentence() != "This formula goes over the cost limit of 20000." {
		t.Fatalf("got %+v refused %v", refusal, refused)
	}
}
