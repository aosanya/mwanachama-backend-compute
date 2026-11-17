package functions

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

func text(value string) *string {
	return &value
}

func count(value int) *int {
	return &value
}

func required() function.Rule {
	return function.Rule{Name: problem.RuleRequired}
}

func noNegatives() function.Rule {
	return function.Rule{Name: problem.RuleNoNegatives}
}

func places(n int) function.Rule {
	return function.Rule{Name: problem.RuleDecimalPlaces, Places: count(n)}
}

func between(lowest string, highest string) function.Rule {
	return function.Rule{Name: problem.RuleRange, Min: text(lowest), Max: text(highest)}
}

func moreThan(lowest string) function.Rule {
	return function.Rule{Name: problem.RuleRange, Min: text(lowest), MinExclusive: true}
}

func atLeastOne() function.Rule {
	return function.Rule{Name: problem.RuleMinItems, Items: count(1)}
}

func money(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeDecimal, Rules: []function.Rule{required(), noNegatives(), places(2)}}
}

func positiveMoney(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeDecimal, Rules: []function.Rule{required(), moreThan("0.00"), places(2)}}
}

func moneyList(name string, rules ...function.Rule) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeListDecimal, Rules: append(rules, noNegatives(), places(2))}
}

func whole(name string, rules ...function.Rule) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeInteger, Rules: append([]function.Rule{required()}, rules...)}
}

func day(name string) function.InputDecl {
	return function.InputDecl{Name: name, Type: formula.TypeDate, Rules: []function.Rule{required()}}
}
