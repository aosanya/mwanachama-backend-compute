package limit

type Role int

const (
	RolePlain Role = iota
	RoleAggregate
	RoleListLiteral
	RoleCall
)

type Node interface {
	Children() []Node
	Role() Role
	Items() int
	CalleeCost() int
}

func NodeCount(n Node) int {
	total := 1
	for _, child := range n.Children() {
		total += NodeCount(child)
	}
	return total
}

func TreeDepth(n Node) int {
	deepest := 0
	for _, child := range n.Children() {
		if d := TreeDepth(child); d > deepest {
			deepest = d
		}
	}
	return deepest + 1
}

func (s Settings) CostOf(n Node) int {
	total := 1
	switch n.Role() {
	case RoleAggregate:
		total += s.listItemsOf(n)
	case RoleCall:
		total += n.CalleeCost()
	}
	for _, child := range n.Children() {
		total += s.CostOf(child)
	}
	return total
}

func (s Settings) listItemsOf(aggregate Node) int {
	children := aggregate.Children()
	if len(children) == 0 {
		return 0
	}
	list := children[0]
	if list.Role() == RoleListLiteral {
		return list.Items()
	}
	declared := list.Items()
	if declared > 0 && declared < s.ListItems {
		return declared
	}
	return s.ListItems
}
