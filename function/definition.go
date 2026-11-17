package function

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

type Rule struct {
	Name         string
	Min          *string
	Max          *string
	MinExclusive bool
	MaxExclusive bool
	Places       *int
	Values       []string
	Items        *int
}

type InputDecl struct {
	Name  string
	Type  formula.Type
	Rules []Rule
}

type CheckDecl struct {
	Formula string
	Message string
}

type Definition struct {
	Inputs []InputDecl
	Checks []CheckDecl
}

type Row map[string]numtext.Received

type Outcome struct {
	True    bool
	Failure string
}

type Runner interface {
	Run(tree formula.Tree, values map[string]numtext.Value) Outcome
}
