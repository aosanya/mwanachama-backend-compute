package library

import "github.com/aosanya/mwanachama-backend-compute/formula"

type Use struct {
	Ref
	Path formula.Path
}

func Uses(tree formula.Tree) []Use {
	uses := []Use{}
	collectUses(tree.Root, formula.Path{}, &uses)
	return uses
}

func collectUses(node formula.Node, path formula.Path, uses *[]Use) {
	if node == nil {
		return
	}
	if call, isCall := node.(formula.Call); isCall && call.Library != nil {
		ref := Ref{Library: *call.Library, Name: call.Name}
		if call.Version != nil {
			ref.Version = *call.Version
		}
		*uses = append(*uses, Use{Ref: ref, Path: append(formula.Path{}, path...)})
	}
	for index, child := range node.Children() {
		collectUses(child, append(append(formula.Path{}, path...), index), uses)
	}
}
