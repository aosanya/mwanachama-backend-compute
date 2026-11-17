package formula

import (
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/limit"
)

type argumentSpec struct {
	accepts []Type
	needed  string
	places  bool
}

type builtinSpec struct {
	args   []argumentSpec
	result func(args []Type) Type
}

var (
	numberOrFloat = argumentSpec{accepts: []Type{TypeDecimal, TypeFloat}, needed: "a decimal or a float"}
	placesCount   = argumentSpec{accepts: []Type{TypeInteger}, needed: "an integer from 0 to 10 written in the formula", places: true}
	anyList       = argumentSpec{accepts: []Type{TypeListInteger, TypeListDecimal}, needed: "a list of integer or a list of decimal"}
	decimalList   = argumentSpec{accepts: []Type{TypeListInteger, TypeListDecimal}, needed: "a list of integer or a list of decimal"}
	dateArgument  = argumentSpec{accepts: []Type{TypeDate}, needed: "a date"}
)

func itemType(list Type) Type {
	if list == TypeListInteger {
		return TypeInteger
	}
	return TypeDecimal
}

var builtins = map[string]builtinSpec{
	"round": {[]argumentSpec{numberOrFloat, placesCount}, func(a []Type) Type { return a[0] }},
	"sum":   {[]argumentSpec{anyList}, func(a []Type) Type { return itemType(a[0]) }},
	"count": {[]argumentSpec{anyList}, func([]Type) Type { return TypeInteger }},
	"avg": {[]argumentSpec{decimalList}, func(a []Type) Type {
		if a[0] == TypeListInteger {
			return TypeFloat
		}
		return TypeDecimal
	}},
	"min":            {[]argumentSpec{anyList}, func(a []Type) Type { return itemType(a[0]) }},
	"max":            {[]argumentSpec{anyList}, func(a []Type) Type { return itemType(a[0]) }},
	"to_float":       {[]argumentSpec{{accepts: []Type{TypeDecimal}, needed: "a decimal"}}, func([]Type) Type { return TypeFloat }},
	"to_decimal":     {[]argumentSpec{{accepts: []Type{TypeFloat}, needed: "a float"}, placesCount}, func([]Type) Type { return TypeDecimal }},
	"days_between":   {[]argumentSpec{dateArgument, dateArgument}, func([]Type) Type { return TypeInteger }},
	"months_between": {[]argumentSpec{dateArgument, dateArgument}, func([]Type) Type { return TypeInteger }},
	"years_between":  {[]argumentSpec{dateArgument, dateArgument}, func([]Type) Type { return TypeInteger }},
}

var aggregates = map[string]bool{"sum": true, "count": true, "avg": true, "min": true, "max": true}

func isBuiltin(call Call) bool {
	if call.Library != nil {
		return false
	}
	_, known := builtins[call.Name]
	return known
}

func countSentence(label string, needed int, given int) string {
	word := "arguments"
	if needed == 1 {
		word = "argument"
	}
	return "`" + label + "` needs " + strconv.Itoa(needed) + " " + word + " and this call has " + strconv.Itoa(given) + "."
}

func accepted(spec []Type, actual Type) bool {
	for _, t := range spec {
		if t == actual {
			return true
		}
	}
	return false
}

func writtenPlaces(n Node) bool {
	literal, isLiteral := n.(Literal)
	if !isLiteral || literal.Type != TypeInteger {
		return false
	}
	value, err := strconv.Atoi(literal.Value)
	return err == nil && value >= 0 && value <= 10
}

func (c *checker) argumentTypeProblem(call Call, path Path, index int, needed string, actual Type) {
	argument := call.Args[index]
	text := Format(Tree{Root: argument})
	c.add(codeArgumentType, "Argument "+strconv.Itoa(index+1)+" of `"+callLabel(call)+"` must be "+needed+", and `"+text+"` is "+article(actual)+".", argument, childPath(path, index))
}

func (c *checker) callType(node Call, path Path) (Type, bool) {
	types := make([]Type, len(node.Args))
	known := make([]bool, len(node.Args))
	allKnown := true
	for i, arg := range node.Args {
		types[i], known[i] = c.typeOf(arg, childPath(path, i))
		allKnown = allKnown && known[i]
	}
	if isBuiltin(node) {
		return c.builtinType(node, path, types, known, allKnown)
	}
	return c.libraryType(node, path, types, known, allKnown)
}

func (c *checker) builtinType(node Call, path Path, types []Type, known []bool, allKnown bool) (Type, bool) {
	if node.Version != nil {
		c.add(codeVersionBuiltIn, "`"+node.Name+"` is built in and takes no version.", node, path)
		return "", false
	}
	spec := builtins[node.Name]
	if len(node.Args) != len(spec.args) {
		c.add(codeArgumentCount, countSentence(node.Name, len(spec.args), len(node.Args)), node, path)
		return "", false
	}
	faulty := false
	for i, arg := range spec.args {
		if !known[i] {
			continue
		}
		if !accepted(arg.accepts, types[i]) || arg.places && !writtenPlaces(node.Args[i]) {
			c.argumentTypeProblem(node, path, i, arg.needed, types[i])
			faulty = true
		}
	}
	if faulty || !allKnown {
		return "", false
	}
	return spec.result(types), true
}

func (c *checker) libraryType(node Call, path Path, types []Type, known []bool, allKnown bool) (Type, bool) {
	libraryName := ""
	if node.Library != nil {
		libraryName = *node.Library
	}
	var versions []int
	if c.library != nil {
		versions = c.library.Versions(libraryName, node.Name)
	}
	if len(versions) == 0 {
		c.add(codeUnknownFunction, "The function "+qualifiedName(node)+" is not known.", node, path)
		return "", false
	}
	if node.Version == nil {
		available := make([]string, len(versions))
		for i, v := range versions {
			available[i] = qualifiedName(node) + "@" + strconv.Itoa(v)
		}
		c.add(codeVersionMissing, qualifiedName(node)+" needs a version. Available: "+joinNames(available)+".", node, path)
		return "", false
	}
	signature, found := c.library.Function(libraryName, node.Name, *node.Version)
	if !found {
		latest := versions[len(versions)-1]
		c.add(codeVersionNotFound, "`"+callLabel(node)+"` does not exist. Latest is `"+qualifiedName(node)+"@"+strconv.Itoa(latest)+"`.", node, path)
		return "", false
	}
	if len(node.Args) != len(signature.Params) {
		c.add(codeArgumentCount, countSentence(callLabel(node), len(signature.Params), len(node.Args)), node, path)
		return "", false
	}
	faulty := false
	for i, param := range signature.Params {
		if !known[i] {
			continue
		}
		if types[i] != param.Type && !(param.Type == TypeDecimal && types[i] == TypeInteger) {
			c.argumentTypeProblem(node, path, i, article(param.Type), types[i])
			faulty = true
		}
	}
	if faulty || !allKnown {
		return "", false
	}
	return signature.Result, true
}

type costNode struct {
	node    Node
	library Library
}

func wrapForCost(n Node, library Library) limit.Node {
	return costNode{node: n, library: library}
}

func (c costNode) Children() []limit.Node {
	children := c.node.Children()
	wrapped := make([]limit.Node, len(children))
	for i, child := range children {
		wrapped[i] = wrapForCost(child, c.library)
	}
	return wrapped
}

func (c costNode) Role() limit.Role {
	switch node := c.node.(type) {
	case List:
		return limit.RoleListLiteral
	case Call:
		if isBuiltin(node) {
			if aggregates[node.Name] {
				return limit.RoleAggregate
			}
			return limit.RolePlain
		}
		return limit.RoleCall
	}
	return limit.RolePlain
}

func (c costNode) Items() int {
	if list, isList := c.node.(List); isList {
		return len(list.Items)
	}
	return 0
}

func (c costNode) CalleeCost() int {
	call, isCall := c.node.(Call)
	if !isCall || c.library == nil || call.Version == nil {
		return 0
	}
	libraryName := ""
	if call.Library != nil {
		libraryName = *call.Library
	}
	signature, found := c.library.Function(libraryName, call.Name, *call.Version)
	if !found {
		return 0
	}
	return signature.Cost
}

func costOf(root Node, library Library, settings limit.Settings) int {
	return settings.CostOf(wrapForCost(root, library))
}
