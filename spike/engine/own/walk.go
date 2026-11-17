package own

func Children(n Node) []Node {
	switch v := n.(type) {
	case List:
		return v.Elements
	case Unary:
		return []Node{v.Operand}
	case Binary:
		return []Node{v.Left, v.Right}
	case Conditional:
		return []Node{v.Test, v.Then, v.Else}
	case Call:
		return v.Args
	}
	return nil
}

func Depth(n Node) int {
	deepest := 0
	for _, child := range Children(n) {
		deepest = max(deepest, Depth(child))
	}
	return deepest + 1
}

func Count(n Node) int {
	total := 1
	for _, child := range Children(n) {
		total += Count(child)
	}
	return total
}
