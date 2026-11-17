package exprlang

import (
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"

	"github.com/aosanya/mwanachama-backend-compute/spike/fit/shape"
	"github.com/aosanya/mwanachama-backend-compute/spike/harness"
)

const decimalFunction = "dec"

var _ harness.TreeEngine = engine{}

func (engine) Tree(c harness.Case) (any, error) {
	tree, err := parser.Parse(translate(c.Source))
	if err != nil {
		return nil, err
	}
	return tree.Node, nil
}

func (e engine) Print(c harness.Case) (string, error) {
	tree, err := e.Tree(c)
	if err != nil {
		return "", err
	}
	return Reverse(tree.(ast.Node).String()), nil
}

func Reverse(text string) string {
	return shape.Reverse(text, decimalFunction)
}
