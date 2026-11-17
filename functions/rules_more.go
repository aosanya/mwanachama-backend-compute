package functions

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func allowed(values ...string) function.Rule {
	return function.Rule{Name: problem.RuleAllowedValues, Values: values}
}

func word(name string, values ...string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeText, Rules: []function.Rule{required(), allowed(values...)}}
}

func flag(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeBoolean, Rules: []function.Rule{required()}}
}

func ratio(name string, lowest string, highest string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeFloat, Rules: []function.Rule{required(), between(lowest, highest)}}
}

func scoreList(name string, lowest string, highest string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeListInteger, Rules: []function.Rule{atLeastOne(), between(lowest, highest)}}
}
