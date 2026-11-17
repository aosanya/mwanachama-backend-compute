package laws

import (
	"math/rand"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var inputPool = []formula.Input{
	{Name: "count", Type: formula.TypeInteger},
	{Name: "days", Type: formula.TypeInteger},
	{Name: "balance", Type: formula.TypeDecimal},
	{Name: "payment", Type: formula.TypeDecimal},
	{Name: "rate", Type: formula.TypeFloat},
	{Name: "label", Type: formula.TypeText},
	{Name: "active", Type: formula.TypeBoolean},
	{Name: "opened", Type: formula.TypeDate},
	{Name: "closed", Type: formula.TypeDate},
	{Name: "amounts", Type: formula.TypeListInteger},
	{Name: "shares", Type: formula.TypeListDecimal},
}

var textPool = []string{"", "active", "say \"hi\"", "back\\slash", "two words", "a, b", "(x)"}

var arithmeticOps = []string{"+", "-", "*"}
var comparisonOps = []string{"<", "<=", ">", ">="}
var equalityOps = []string{"==", "!="}
var logicOps = []string{"&&", "||"}
var aggregateNames = []string{"sum", "min", "max"}
var dateFunctions = []string{"days_between", "months_between", "years_between"}

const maxAttempts = 400

type Case struct {
	Seed   int64
	Inputs []formula.Input
	Tree   formula.Tree
}

func (c Case) Text() string {
	return formula.Format(c.Tree)
}

func NewCase(seed int64) Case {
	rng := rand.New(rand.NewSource(seed))
	order := rng.Perm(len(inputPool))
	inputs := make([]formula.Input, 0, 8)
	for _, index := range order[:8] {
		inputs = append(inputs, inputPool[index])
	}
	for attempt := 0; attempt < maxAttempts; attempt++ {
		g := &generator{rng: rng, inputs: inputs, budget: 20 + rng.Intn(40)}
		root := g.expression(g.topType(), 3+rng.Intn(4))
		tree := formula.Tree{Root: root}
		if len(formula.Check(tree, inputs, nil)) == 0 {
			return Case{Seed: seed, Inputs: inputs, Tree: tree}
		}
	}
	return Case{Seed: seed, Inputs: inputs, Tree: formula.Tree{Root: formula.Literal{Type: formula.TypeInteger, Value: "1"}}}
}

type generator struct {
	rng    *rand.Rand
	inputs []formula.Input
	budget int
}

func (g *generator) topType() formula.Type {
	tops := []formula.Type{formula.TypeInteger, formula.TypeDecimal, formula.TypeFloat, formula.TypeText, formula.TypeBoolean}
	return tops[g.rng.Intn(len(tops))]
}

func (g *generator) named(kind formula.Type) []string {
	var names []string
	for _, input := range g.inputs {
		if input.Type == kind {
			names = append(names, input.Name)
		}
	}
	return names
}

func (g *generator) pickName(kind formula.Type) (formula.Node, bool) {
	names := g.named(kind)
	if len(names) == 0 {
		return nil, false
	}
	return formula.Name{Name: names[g.rng.Intn(len(names))]}, true
}

func (g *generator) digits(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte('0' + g.rng.Intn(10))
	}
	return string(out)
}

func (g *generator) integerLiteral() formula.Node {
	if g.rng.Intn(4) == 0 {
		return formula.Literal{Type: formula.TypeInteger, Value: "0"}
	}
	return formula.Literal{Type: formula.TypeInteger, Value: string(rune('1'+g.rng.Intn(9))) + g.digits(g.rng.Intn(4))}
}

func (g *generator) decimalLiteral() formula.Node {
	whole := "0"
	if g.rng.Intn(2) == 0 {
		whole = string(rune('1'+g.rng.Intn(9))) + g.digits(g.rng.Intn(3))
	}
	return formula.Literal{Type: formula.TypeDecimal, Value: whole + "." + g.digits(1+g.rng.Intn(3))}
}

func (g *generator) places() formula.Node {
	return formula.Literal{Type: formula.TypeInteger, Value: strconv.Itoa(g.rng.Intn(11))}
}

func (g *generator) leaf(kind formula.Type) formula.Node {
	if g.rng.Intn(2) == 0 {
		if name, found := g.pickName(kind); found {
			return name
		}
	}
	switch kind {
	case formula.TypeInteger:
		return g.integerLiteral()
	case formula.TypeDecimal:
		return g.decimalLiteral()
	case formula.TypeFloat:
		return formula.Call{Name: "to_float", Args: []formula.Node{g.decimalLiteral()}}
	case formula.TypeText:
		return formula.Literal{Type: formula.TypeText, Value: textPool[g.rng.Intn(len(textPool))]}
	case formula.TypeBoolean:
		return formula.Literal{Type: formula.TypeBoolean, Value: []string{"true", "false"}[g.rng.Intn(2)]}
	case formula.TypeListInteger:
		return g.listLiteral(formula.TypeInteger)
	case formula.TypeListDecimal:
		return g.listLiteral(formula.TypeDecimal)
	}
	name, _ := g.pickName(kind)
	return name
}

func (g *generator) listLiteral(item formula.Type) formula.Node {
	items := make([]formula.Node, 1+g.rng.Intn(4))
	for i := range items {
		if item == formula.TypeInteger {
			items[i] = g.integerLiteral()
		} else {
			items[i] = g.decimalLiteral()
		}
	}
	return formula.List{Items: items}
}

func (g *generator) list(kind formula.Type, depth int) formula.Node {
	if depth > 1 && g.rng.Intn(3) == 0 {
		item := formula.TypeInteger
		if kind == formula.TypeListDecimal {
			item = formula.TypeDecimal
		}
		items := make([]formula.Node, 1+g.rng.Intn(3))
		for i := range items {
			items[i] = g.expression(item, depth-1)
		}
		return formula.List{Items: items}
	}
	return g.leaf(kind)
}

func (g *generator) anyList(depth int) (formula.Node, formula.Type) {
	if g.rng.Intn(2) == 0 {
		return g.list(formula.TypeListInteger, depth), formula.TypeListInteger
	}
	return g.list(formula.TypeListDecimal, depth), formula.TypeListDecimal
}

func (g *generator) expression(kind formula.Type, depth int) formula.Node {
	g.budget--
	if depth <= 1 || g.budget <= 0 || g.rng.Intn(10) < 3 {
		return g.leaf(kind)
	}
	switch kind {
	case formula.TypeInteger:
		return g.integer(depth)
	case formula.TypeDecimal:
		return g.decimal(depth)
	case formula.TypeFloat:
		return g.float(depth)
	case formula.TypeBoolean:
		return g.boolean(depth)
	case formula.TypeText:
		return g.conditional(kind, depth)
	}
	return g.leaf(kind)
}

func (g *generator) conditional(kind formula.Type, depth int) formula.Node {
	return formula.Conditional{
		Condition: g.expression(formula.TypeBoolean, depth-1),
		Then:      g.expression(kind, depth-1),
		Else:      g.expression(kind, depth-1),
	}
}

func (g *generator) binary(op string, left formula.Type, right formula.Type, depth int) formula.Node {
	return formula.Binary{Op: op, Left: g.expression(left, depth-1), Right: g.expression(right, depth-1)}
}

func (g *generator) pick(ops []string) string {
	return ops[g.rng.Intn(len(ops))]
}

func (g *generator) dateCall(depth int) (formula.Node, bool) {
	dates := g.named(formula.TypeDate)
	if len(dates) == 0 {
		return nil, false
	}
	pick := func() formula.Node { return formula.Name{Name: dates[g.rng.Intn(len(dates))]} }
	return formula.Call{Name: g.pick(dateFunctions), Args: []formula.Node{pick(), pick()}}, true
}

func (g *generator) integer(depth int) formula.Node {
	switch g.rng.Intn(6) {
	case 0:
		return g.binary(g.pick(arithmeticOps), formula.TypeInteger, formula.TypeInteger, depth)
	case 1:
		return g.binary("%", formula.TypeInteger, formula.TypeInteger, depth)
	case 2:
		return g.conditional(formula.TypeInteger, depth)
	case 3:
		list, _ := g.anyList(depth)
		return formula.Call{Name: "count", Args: []formula.Node{list}}
	case 4:
		return formula.Call{Name: g.pick(aggregateNames), Args: []formula.Node{g.list(formula.TypeListInteger, depth)}}
	}
	if call, found := g.dateCall(depth); found {
		return call
	}
	return g.binary(g.pick(arithmeticOps), formula.TypeInteger, formula.TypeInteger, depth)
}

func (g *generator) decimal(depth int) formula.Node {
	switch g.rng.Intn(9) {
	case 0:
		return g.binary(g.pick(arithmeticOps), formula.TypeDecimal, formula.TypeDecimal, depth)
	case 1:
		return g.binary(g.pick(arithmeticOps), formula.TypeDecimal, formula.TypeInteger, depth)
	case 2:
		return g.binary(g.pick(arithmeticOps), formula.TypeInteger, formula.TypeDecimal, depth)
	case 3:
		right := []formula.Type{formula.TypeDecimal, formula.TypeInteger}[g.rng.Intn(2)]
		return g.binary("/", formula.TypeDecimal, right, depth)
	case 4:
		return g.conditional(formula.TypeDecimal, depth)
	case 5:
		return formula.Call{Name: "round", Args: []formula.Node{g.expression(formula.TypeDecimal, depth-1), g.places()}}
	case 6:
		return formula.Call{Name: g.pick(aggregateNames), Args: []formula.Node{g.list(formula.TypeListDecimal, depth)}}
	case 7:
		return formula.Call{Name: "to_decimal", Args: []formula.Node{g.expression(formula.TypeFloat, depth-1), g.places()}}
	}
	return formula.Call{Name: "avg", Args: []formula.Node{g.list(formula.TypeListDecimal, depth)}}
}

func (g *generator) float(depth int) formula.Node {
	switch g.rng.Intn(6) {
	case 0:
		return g.binary(g.pick(append([]string{"/"}, arithmeticOps...)), formula.TypeFloat, formula.TypeFloat, depth)
	case 1:
		return g.conditional(formula.TypeFloat, depth)
	case 2:
		return formula.Call{Name: "to_float", Args: []formula.Node{g.expression(formula.TypeDecimal, depth-1)}}
	case 3:
		return formula.Call{Name: "avg", Args: []formula.Node{g.list(formula.TypeListInteger, depth)}}
	case 4:
		return formula.Call{Name: "round", Args: []formula.Node{g.expression(formula.TypeFloat, depth-1), g.places()}}
	}
	return g.binary(g.pick(arithmeticOps), formula.TypeFloat, formula.TypeFloat, depth)
}

func (g *generator) boolean(depth int) formula.Node {
	switch g.rng.Intn(7) {
	case 0:
		return formula.Unary{Op: "!", Operand: g.expression(formula.TypeBoolean, depth-1)}
	case 1:
		return g.binary(g.pick(logicOps), formula.TypeBoolean, formula.TypeBoolean, depth)
	case 2:
		pairs := [][2]formula.Type{
			{formula.TypeInteger, formula.TypeInteger},
			{formula.TypeDecimal, formula.TypeDecimal},
			{formula.TypeInteger, formula.TypeDecimal},
			{formula.TypeDecimal, formula.TypeInteger},
			{formula.TypeFloat, formula.TypeFloat},
		}
		pair := pairs[g.rng.Intn(len(pairs))]
		return g.binary(g.pick(comparisonOps), pair[0], pair[1], depth)
	case 3:
		pairs := [][2]formula.Type{
			{formula.TypeInteger, formula.TypeInteger},
			{formula.TypeDecimal, formula.TypeDecimal},
			{formula.TypeInteger, formula.TypeDecimal},
		}
		pair := pairs[g.rng.Intn(len(pairs))]
		return g.binary(g.pick(equalityOps), pair[0], pair[1], depth)
	case 4:
		return g.binary(g.pick(equalityOps), formula.TypeText, formula.TypeText, depth)
	case 5:
		return g.conditional(formula.TypeBoolean, depth)
	}
	return g.binary(g.pick(comparisonOps), formula.TypeInteger, formula.TypeInteger, depth)
}
