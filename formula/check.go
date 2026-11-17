package formula

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/limit"
)

const maxProblems = 20

func Check(tree Tree, inputs []Input, library Library) []Problem {
	if faults := shapeFaults(tree.Root); len(faults) > 0 {
		return shapeProblems(faults)
	}
	settings := limit.Proposed()
	whole := Format(tree)
	if refused, over := settings.Refusal(sizeOf(tree.Root, whole)); over {
		return []Problem{wholeFormulaProblem(whole, refused)}
	}
	c := &checker{inputs: inputs, library: library, tree: tree}
	c.typeOf(tree.Root, Path{})
	if len(c.problems) > 0 {
		return capProblems(c.problems)
	}
	if refused, over := settings.Refusal(limit.Measure{CallDepth: callDepth(tree.Root, library)}); over {
		return []Problem{wholeFormulaProblem(whole, refused)}
	}
	if refused, over := settings.Refusal(limit.Measure{Cost: costOf(tree.Root, library, settings)}); over {
		return []Problem{wholeFormulaProblem(whole, refused)}
	}
	return []Problem{}
}

func Cost(tree Tree, library Library) int {
	if tree.Root == nil {
		return 0
	}
	return costOf(tree.Root, library, limit.Proposed())
}

func capProblems(problems []Problem) []Problem {
	if len(problems) > maxProblems {
		return problems[:maxProblems]
	}
	return problems
}

func shapeProblems(faults []shapeFault) []Problem {
	problems := make([]Problem, 0, len(faults))
	for _, fault := range faults {
		problems = append(problems, Problem{
			Code:     codeBadTree,
			Sentence: "The part at path " + pathText(fault.path) + " is not a valid part: " + fault.what + ".",
			Path:     fault.path,
		})
	}
	return capProblems(problems)
}

func wholeFormulaProblem(whole string, refused limit.Refusal) Problem {
	return Problem{Code: limitCode(refused.Limit), Sentence: refused.Sentence(), Part: whole, Path: Path{}}
}

func sizeOf(root Node, whole string) limit.Measure {
	measure := limit.Measure{
		FormulaLength: utf8.RuneCountInString(whole),
		Depth:         treeDepth(root),
		Nodes:         nodeCount(root),
	}
	names := map[string]bool{}
	var walk func(n Node)
	walk = func(n Node) {
		switch node := n.(type) {
		case Name:
			names[node.Name] = true
		case Literal:
			measureLiteral(&measure, node)
		}
		for _, child := range n.Children() {
			walk(child)
		}
	}
	walk(root)
	measure.Inputs = len(names)
	return measure
}

func measureLiteral(measure *limit.Measure, literal Literal) {
	switch literal.Type {
	case TypeText:
		if n := utf8.RuneCountInString(literal.Value); n > measure.TextLiteral {
			measure.TextLiteral = n
		}
	case TypeInteger:
		if n := len(literal.Value); n > measure.IntegerDigits {
			measure.IntegerDigits = n
		}
	case TypeDecimal:
		digits := len(literal.Value) - 1
		places := 0
		if dot := strings.IndexByte(literal.Value, '.'); dot >= 0 {
			places = len(literal.Value) - dot - 1
		}
		if digits > measure.DecimalDigits {
			measure.DecimalDigits = digits
		}
		if places > measure.DecimalPlaces {
			measure.DecimalPlaces = places
		}
	}
}

type checker struct {
	inputs   []Input
	library  Library
	tree     Tree
	problems []Problem
}

func (c *checker) add(code string, sentence string, n Node, path Path) {
	c.problems = append(c.problems, Problem{Code: code, Sentence: sentence, Part: Format(Tree{Root: n}), Path: append(Path{}, path...)})
}

func (c *checker) inputType(name string) (Type, bool) {
	for _, input := range c.inputs {
		if input.Name == name {
			return input.Type, true
		}
	}
	return "", false
}

func (c *checker) inputNames() string {
	names := make([]string, len(c.inputs))
	for i, input := range c.inputs {
		names[i] = input.Name
	}
	return joinNames(names)
}

func article(t Type) string {
	if t == TypeInteger {
		return "an integer"
	}
	return "a " + string(t)
}

func isNumber(t Type) bool {
	return t == TypeInteger || t == TypeDecimal || t == TypeFloat
}

func commonType(a Type, b Type) (Type, bool) {
	switch {
	case a == b:
		return a, true
	case a == TypeInteger && b == TypeDecimal, a == TypeDecimal && b == TypeInteger:
		return TypeDecimal, true
	}
	return "", false
}

func (c *checker) typeOf(n Node, path Path) (Type, bool) {
	switch node := n.(type) {
	case Literal:
		return node.Type, true
	case Name:
		if t, found := c.inputType(node.Name); found {
			return t, true
		}
		sentence := "The name `" + node.Name + "` is not an input of this function. Its inputs are: " + c.inputNames() + "."
		if len(c.inputs) == 0 {
			sentence = "The name `" + node.Name + "` is not an input of this function. It has no inputs."
		}
		c.add(codeUnknownName, sentence, n, path)
		return "", false
	case Unary:
		return c.unaryType(node, path)
	case Binary:
		return c.binaryType(node, path)
	case Conditional:
		return c.conditionalType(node, path)
	case Call:
		return c.callType(node, path)
	case List:
		return c.listType(node, path)
	}
	return "", false
}

func (c *checker) unaryType(node Unary, path Path) (Type, bool) {
	operand, ok := c.typeOf(node.Operand, childPath(path, 0))
	if !ok {
		return "", false
	}
	if operand != TypeBoolean {
		c.add(codeTypeMismatch, mismatchSentence(node, "a boolean", string(operand)), node, path)
		return "", false
	}
	return TypeBoolean, true
}

func mismatchSentence(n Node, needs string, types string) string {
	op := "!"
	if binary, isBinary := n.(Binary); isBinary {
		op = binary.Op
	}
	return "The part `" + Format(Tree{Root: n}) + "` cannot be worked out: `" + op + "` needs " + needs + ", and this has " + types + "."
}

func mixSentence(n Node) string {
	return "The part `" + Format(Tree{Root: n}) + "` mixes a decimal with a float; convert one with to_float or to_decimal."
}

func isMix(a Type, b Type) bool {
	return a == TypeDecimal && b == TypeFloat || a == TypeFloat && b == TypeDecimal
}

func (c *checker) binaryType(node Binary, path Path) (Type, bool) {
	left, leftOK := c.typeOf(node.Left, childPath(path, 0))
	right, rightOK := c.typeOf(node.Right, childPath(path, 1))
	if !leftOK || !rightOK {
		return "", false
	}
	result, code, needs := binaryResult(node.Op, left, right)
	switch code {
	case "":
		return result, true
	case codeFloatDecimalMix:
		c.add(code, mixSentence(node), node, path)
	default:
		c.add(code, mismatchSentence(node, needs, string(left)+" and "+string(right)), node, path)
	}
	return "", false
}

func binaryResult(op string, left Type, right Type) (Type, string, string) {
	switch op {
	case "+", "-", "*":
		return arithmetic(left, right, "two numbers", false)
	case "/":
		return arithmetic(left, right, "a decimal or a float on at least one side", true)
	case "%":
		if left == TypeInteger && right == TypeInteger {
			return TypeInteger, "", ""
		}
		return "", codeTypeMismatch, "two integers"
	case "<", "<=", ">", ">=":
		if isMix(left, right) {
			return "", codeFloatDecimalMix, ""
		}
		if isNumber(left) && isNumber(right) {
			return TypeBoolean, "", ""
		}
		return "", codeTypeMismatch, "two numbers"
	case "==", "!=":
		if isMix(left, right) {
			return "", codeFloatDecimalMix, ""
		}
		exact := func(t Type) bool { return t == TypeInteger || t == TypeDecimal }
		if exact(left) && exact(right) || left == TypeText && right == TypeText {
			return TypeBoolean, "", ""
		}
		return "", codeTypeMismatch, "two numbers or two texts"
	}
	if left == TypeBoolean && right == TypeBoolean {
		return TypeBoolean, "", ""
	}
	return "", codeTypeMismatch, "two booleans"
}

func arithmetic(left Type, right Type, needs string, division bool) (Type, string, string) {
	if !isNumber(left) || !isNumber(right) {
		return "", codeTypeMismatch, needs
	}
	if isMix(left, right) {
		return "", codeFloatDecimalMix, ""
	}
	switch {
	case left == TypeFloat || right == TypeFloat:
		return TypeFloat, "", ""
	case left == TypeDecimal || right == TypeDecimal:
		return TypeDecimal, "", ""
	case division:
		return "", codeTypeMismatch, needs
	}
	return TypeInteger, "", ""
}

func (c *checker) conditionalType(node Conditional, path Path) (Type, bool) {
	condition, conditionOK := c.typeOf(node.Condition, childPath(path, 0))
	then, thenOK := c.typeOf(node.Then, childPath(path, 1))
	otherwise, elseOK := c.typeOf(node.Else, childPath(path, 2))
	if conditionOK && condition != TypeBoolean {
		c.add(codeConditionType, "The condition `"+Format(Tree{Root: node.Condition})+"` must be true or false, and it is "+string(condition)+".", node.Condition, childPath(path, 0))
		return "", false
	}
	if !conditionOK || !thenOK || !elseOK {
		return "", false
	}
	common, same := commonType(then, otherwise)
	if !same {
		c.add(codeBranchesDiffer, "The two answers of `"+Format(Tree{Root: node})+"` must be the same kind, and they are "+string(then)+" and "+string(otherwise)+".", node, path)
		return "", false
	}
	return common, true
}

func (c *checker) listType(node List, path Path) (Type, bool) {
	var first Type
	var common Type
	valid := true
	reported := false
	for i, item := range node.Items {
		t, ok := c.typeOf(item, childPath(path, i))
		if !ok {
			valid = false
			continue
		}
		if i == 0 || first == "" {
			first, common = t, t
			continue
		}
		merged, same := commonType(common, t)
		if !same && !reported {
			c.add(codeListItemsDiffer, "The items of `"+Format(Tree{Root: node})+"` must all be the same kind, and this list has "+string(common)+" and "+string(t)+".", node, path)
			reported = true
		}
		if same {
			common = merged
		}
	}
	if !valid || reported {
		return "", false
	}
	switch common {
	case TypeInteger:
		return TypeListInteger, true
	case TypeDecimal:
		return TypeListDecimal, true
	}
	c.add(codeListItemsDiffer, "The items of `"+Format(Tree{Root: node})+"` must be integers or decimals, and this list has "+string(common)+".", node, path)
	return "", false
}

func callDepth(root Node, library Library) int {
	deepest := 0
	var walk func(n Node)
	walk = func(n Node) {
		if call, isCall := n.(Call); isCall && !isBuiltin(call) && library != nil && call.Version != nil {
			libraryName := ""
			if call.Library != nil {
				libraryName = *call.Library
			}
			if signature, found := library.Function(libraryName, call.Name, *call.Version); found && signature.CallDepth+1 > deepest {
				deepest = signature.CallDepth + 1
			}
		}
		for _, child := range n.Children() {
			walk(child)
		}
	}
	walk(root)
	return deepest
}

func qualifiedName(call Call) string {
	if call.Library != nil {
		return *call.Library + "." + call.Name
	}
	return call.Name
}

func callLabel(call Call) string {
	if call.Version != nil {
		return qualifiedName(call) + "@" + strconv.Itoa(*call.Version)
	}
	return qualifiedName(call)
}
