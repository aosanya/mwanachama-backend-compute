package own

import "github.com/aosanya/mwanachama-backend-compute/spike/harness"

var _ harness.TreeEngine = engine{}

func (engine) Tree(c harness.Case) (any, error) {
	return Parse(c.Source)
}

func (engine) Print(c harness.Case) (string, error) {
	tree, err := Parse(c.Source)
	if err != nil {
		return "", err
	}
	return Format(tree), nil
}
