package fit

import (
	"fmt"
	"strings"
)

const (
	KindLiteral     = "literal"
	KindName        = "name"
	KindCall        = "call"
	KindBinary      = "binary"
	KindUnary       = "unary"
	KindConditional = "conditional"
	KindList        = "list"
)

type Node struct {
	Kind     string
	Text     string
	Children []Node
}

func leaf(kind, text string) Node {
	return Node{Kind: kind, Text: text}
}

func branch(kind, text string, children ...Node) Node {
	if len(children) == 0 {
		children = nil
	}
	return Node{Kind: kind, Text: text, Children: children}
}

func (n Node) String() string {
	if len(n.Children) == 0 {
		return fmt.Sprintf("%s:%s", n.Kind, n.Text)
	}
	parts := make([]string, len(n.Children))
	for i, child := range n.Children {
		parts[i] = child.String()
	}
	return fmt.Sprintf("%s:%s(%s)", n.Kind, n.Text, strings.Join(parts, " "))
}

func (n Node) Kinds() map[string]bool {
	found := map[string]bool{n.Kind: true}
	for _, child := range n.Children {
		for kind := range child.Kinds() {
			found[kind] = true
		}
	}
	return found
}

func convertAll[T any](items []T, convert func(T) (Node, error)) ([]Node, error) {
	nodes := make([]Node, 0, len(items))
	for _, item := range items {
		node, err := convert(item)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

type NotConvertible struct {
	Engine string
	Raw    string
}

func (e NotConvertible) Error() string {
	return fmt.Sprintf("%s node %s has no neutral kind", e.Engine, e.Raw)
}

func (n Node) Count() int {
	total := 1
	for _, child := range n.Children {
		total += child.Count()
	}
	return total
}
