package laws

import (
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var simplestLeaves = []formula.Node{
	formula.Literal{Type: formula.TypeInteger, Value: "0"},
	formula.Literal{Type: formula.TypeDecimal, Value: "0.0"},
	formula.Literal{Type: formula.TypeBoolean, Value: "true"},
	formula.Literal{Type: formula.TypeText, Value: ""},
}

func TreeSize(tree formula.Tree) int {
	return nodeSize(tree.Root)
}

func nodeSize(n formula.Node) int {
	size := 1
	for _, child := range n.Children() {
		size += nodeSize(child)
	}
	return size
}

func Shrink(tree formula.Tree, valid func(formula.Tree) bool, fails func(formula.Tree) bool) formula.Tree {
	current := tree
	for {
		smaller, found := smallerFailure(current, valid, fails)
		if !found {
			return current
		}
		current = smaller
	}
}

func smallerFailure(tree formula.Tree, valid func(formula.Tree) bool, fails func(formula.Tree) bool) (formula.Tree, bool) {
	size := TreeSize(tree)
	for _, path := range paths(tree.Root, formula.Path{}) {
		node, _ := nodeAt(tree.Root, path)
		replacements := append([]formula.Node{}, node.Children()...)
		if len(node.Children()) > 0 {
			replacements = append(replacements, simplestLeaves...)
		}
		for _, replacement := range replacements {
			candidate := formula.Tree{Root: replaceAt(tree.Root, path, replacement)}
			if TreeSize(candidate) < size && valid(candidate) && fails(candidate) {
				return candidate, true
			}
		}
	}
	return tree, false
}

func paths(n formula.Node, prefix formula.Path) []formula.Path {
	out := []formula.Path{append(formula.Path{}, prefix...)}
	for i, child := range n.Children() {
		out = append(out, paths(child, append(append(formula.Path{}, prefix...), i))...)
	}
	return out
}

func nodeAt(n formula.Node, path formula.Path) (formula.Node, bool) {
	for _, index := range path {
		children := n.Children()
		if index >= len(children) {
			return nil, false
		}
		n = children[index]
	}
	return n, true
}

func replaceAt(n formula.Node, path formula.Path, with formula.Node) formula.Node {
	if len(path) == 0 {
		return with
	}
	index := path[0]
	rest := path[1:]
	switch node := n.(type) {
	case formula.Unary:
		node.Operand = replaceAt(node.Operand, rest, with)
		return node
	case formula.Binary:
		if index == 0 {
			node.Left = replaceAt(node.Left, rest, with)
		} else {
			node.Right = replaceAt(node.Right, rest, with)
		}
		return node
	case formula.Conditional:
		switch index {
		case 0:
			node.Condition = replaceAt(node.Condition, rest, with)
		case 1:
			node.Then = replaceAt(node.Then, rest, with)
		default:
			node.Else = replaceAt(node.Else, rest, with)
		}
		return node
	case formula.Call:
		args := append([]formula.Node{}, node.Args...)
		args[index] = replaceAt(args[index], rest, with)
		node.Args = args
		return node
	case formula.List:
		items := append([]formula.Node{}, node.Items...)
		items[index] = replaceAt(items[index], rest, with)
		node.Items = items
		return node
	}
	return n
}

func Show(tree formula.Tree) string {
	text := show(tree.Root)
	if strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") && isBinaryOrConditional(tree.Root) {
		return text[1 : len(text)-1]
	}
	return text
}

func isBinaryOrConditional(n formula.Node) bool {
	switch n.(type) {
	case formula.Binary, formula.Conditional:
		return true
	}
	return false
}

func show(n formula.Node) string {
	switch node := n.(type) {
	case formula.Binary:
		return "(" + show(node.Left) + " " + node.Op + " " + show(node.Right) + ")"
	case formula.Unary:
		return node.Op + show(node.Operand)
	case formula.Conditional:
		return "(" + show(node.Condition) + " ? " + show(node.Then) + " : " + show(node.Else) + ")"
	case formula.Call:
		args := make([]string, len(node.Args))
		for i, arg := range node.Args {
			args[i] = show(arg)
		}
		return callName(node) + "(" + strings.Join(args, ", ") + ")"
	case formula.List:
		items := make([]string, len(node.Items))
		for i, item := range node.Items {
			items[i] = show(item)
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	return formula.Format(formula.Tree{Root: n})
}

func callName(call formula.Call) string {
	name := call.Name
	if call.Library != nil {
		name = *call.Library + "." + name
	}
	if call.Version != nil {
		name += "@" + strconv.Itoa(*call.Version)
	}
	return name
}
