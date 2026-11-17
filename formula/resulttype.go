package formula

func ResultType(tree Tree, inputs []Input, library Library) (Type, bool) {
	if tree.Root == nil {
		return "", false
	}
	c := &checker{inputs: inputs, library: library, tree: tree}
	return c.typeOf(tree.Root, Path{})
}
