package formula

import "strconv"

func childPath(path Path, index int) Path {
	return append(append(Path{}, path...), index)
}

func pathText(path Path) string {
	text := "["
	for i, index := range path {
		if i > 0 {
			text += ", "
		}
		text += strconv.Itoa(index)
	}
	return text + "]"
}

func validName(name string) bool {
	if len(name) < 1 || len(name) > 40 || !isNameStart(rune(name[0])) {
		return false
	}
	for _, r := range name {
		if !isNamePart(r) {
			return false
		}
	}
	return true
}

func allDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if !isDigit(r) {
			return false
		}
	}
	return true
}

func validInteger(value string) bool {
	return allDigits(value) && (len(value) == 1 || value[0] != '0')
}

func validDecimal(value string) bool {
	for i, r := range value {
		if r != '.' {
			continue
		}
		return validInteger(value[:i]) && allDigits(value[i+1:])
	}
	return false
}

func literalProblem(literal Literal) string {
	valid := false
	switch literal.Type {
	case TypeInteger:
		valid = validInteger(literal.Value)
	case TypeDecimal:
		valid = validDecimal(literal.Value)
	case TypeBoolean:
		valid = literal.Value == "true" || literal.Value == "false"
	case TypeText:
		valid = true
	case "":
		return "the field type is missing"
	default:
		return "the value " + literal.Value + " is not a " + string(literal.Type)
	}
	if !valid {
		return "the value " + literal.Value + " is not a " + string(literal.Type)
	}
	return ""
}

type shapeFault struct {
	path Path
	what string
}

func shapeFaults(root Node) []shapeFault {
	if root == nil {
		return []shapeFault{{Path{}, "the text is not a tree"}}
	}
	var faults []shapeFault
	var walk func(n Node, path Path)
	missing := func(n Node, path Path, field string) bool {
		if n == nil {
			faults = append(faults, shapeFault{path, "the field " + field + " is missing"})
			return true
		}
		return false
	}
	walk = func(n Node, path Path) {
		switch node := n.(type) {
		case Literal:
			if what := literalProblem(node); what != "" {
				faults = append(faults, shapeFault{path, what})
			}
		case Name:
			if !validName(node.Name) {
				faults = append(faults, shapeFault{path, "the name " + node.Name + " is not a valid name"})
			}
		case Unary:
			if node.Op != "!" {
				faults = append(faults, shapeFault{path, "the operator " + node.Op + " is not in the language"})
			}
			if !missing(node.Operand, path, "operand") {
				walk(node.Operand, childPath(path, 0))
			}
		case Binary:
			if _, known := binaryOperators[node.Op]; !known {
				faults = append(faults, shapeFault{path, "the operator " + node.Op + " is not in the language"})
			}
			if !missing(node.Left, path, "left") {
				walk(node.Left, childPath(path, 0))
			}
			if !missing(node.Right, path, "right") {
				walk(node.Right, childPath(path, 1))
			}
		case Conditional:
			if !missing(node.Condition, path, "condition") {
				walk(node.Condition, childPath(path, 0))
			}
			if !missing(node.Then, path, "then") {
				walk(node.Then, childPath(path, 1))
			}
			if !missing(node.Else, path, "else") {
				walk(node.Else, childPath(path, 2))
			}
		case Call:
			if !validName(node.Name) {
				faults = append(faults, shapeFault{path, "the name " + node.Name + " is not a valid name"})
			}
			if node.Library != nil && !validName(*node.Library) {
				faults = append(faults, shapeFault{path, "the name " + *node.Library + " is not a valid name"})
			}
			if node.Version != nil && (*node.Version < 1 || *node.Version > 999999) {
				faults = append(faults, shapeFault{path, "the value " + strconv.Itoa(*node.Version) + " is not a version from 1 to 999999"})
			}
			walkAll(node.Args, path, "args", missing, walk)
		case List:
			walkAll(node.Items, path, "items", missing, walk)
		default:
			kind := "unknown"
			if n != nil {
				kind = string(n.Kind())
			}
			faults = append(faults, shapeFault{path, "the kind " + kind + " is not in the language"})
		}
	}
	walk(root, Path{})
	return faults
}

func walkAll(children []Node, path Path, field string, missing func(Node, Path, string) bool, walk func(Node, Path)) {
	if len(children) == 0 {
		missing(nil, path, field)
		return
	}
	for i, child := range children {
		if !missing(child, path, field) {
			walk(child, childPath(path, i))
		}
	}
}
