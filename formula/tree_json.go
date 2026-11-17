package formula

import (
	"bytes"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/internal/jsonform"
)

func TreeToJSON(tree Tree) []byte {
	var out bytes.Buffer
	writePart(&out, tree.Root)
	return out.Bytes()
}

func writePart(out *bytes.Buffer, node Node) {
	switch n := node.(type) {
	case nil:
		out.WriteString("null")
	case Literal:
		out.WriteString(`{"kind":"literal","type":`)
		out.Write(jsonform.Quote(string(n.Type)))
		out.WriteString(`,"value":`)
		out.Write(jsonform.Quote(n.Value))
		out.WriteString("}")
	case Name:
		out.WriteString(`{"kind":"name","name":`)
		out.Write(jsonform.Quote(n.Name))
		out.WriteString("}")
	case Unary:
		out.WriteString(`{"kind":"unary","op":`)
		out.Write(jsonform.Quote(n.Op))
		out.WriteString(`,"operand":`)
		writePart(out, n.Operand)
		out.WriteString("}")
	case Binary:
		out.WriteString(`{"kind":"binary","op":`)
		out.Write(jsonform.Quote(n.Op))
		out.WriteString(`,"left":`)
		writePart(out, n.Left)
		out.WriteString(`,"right":`)
		writePart(out, n.Right)
		out.WriteString("}")
	case Conditional:
		out.WriteString(`{"kind":"conditional","condition":`)
		writePart(out, n.Condition)
		out.WriteString(`,"then":`)
		writePart(out, n.Then)
		out.WriteString(`,"else":`)
		writePart(out, n.Else)
		out.WriteString("}")
	case Call:
		out.WriteString(`{"kind":"call","library":`)
		writeOptionalName(out, n.Library)
		out.WriteString(`,"name":`)
		out.Write(jsonform.Quote(n.Name))
		out.WriteString(`,"version":`)
		if n.Version == nil {
			out.WriteString("null")
		} else {
			out.WriteString(strconv.Itoa(*n.Version))
		}
		out.WriteString(`,"args":`)
		writeParts(out, n.Args)
		out.WriteString("}")
	case List:
		out.WriteString(`{"kind":"list","items":`)
		writeParts(out, n.Items)
		out.WriteString("}")
	default:
		out.WriteString(`{"kind":`)
		out.Write(jsonform.Quote(string(node.Kind())))
		out.WriteString("}")
	}
}

func writeOptionalName(out *bytes.Buffer, name *string) {
	if name == nil {
		out.WriteString("null")
		return
	}
	out.Write(jsonform.Quote(*name))
}

func writeParts(out *bytes.Buffer, nodes []Node) {
	out.WriteString("[")
	for i, node := range nodes {
		if i > 0 {
			out.WriteString(",")
		}
		writePart(out, node)
	}
	out.WriteString("]")
}

func TreeFromJSON(data []byte) (Tree, []Problem) {
	raw, ok := jsonform.Decode(data)
	reader := &treeReader{}
	if !ok {
		reader.fault(Path{}, notATree)
		return Tree{}, shapeProblems(reader.faults)
	}
	root := reader.part(raw, Path{})
	if len(reader.faults) > 0 || root == nil {
		return Tree{}, shapeProblems(reader.faults)
	}
	tree := Tree{Root: root}
	return tree, shapeProblems(shapeFaults(root))
}

const notATree = "the text is not a tree"

type treeReader struct {
	faults []shapeFault
}

func (r *treeReader) fault(path Path, what string) {
	r.faults = append(r.faults, shapeFault{path, what})
}

func (r *treeReader) full() bool {
	return len(r.faults) >= maxProblems
}

func (r *treeReader) text(object map[string]any, path Path, field string) (string, bool) {
	value, present := object[field]
	if !present {
		r.fault(path, "the field "+field+" is missing")
		return "", false
	}
	text, isText := value.(string)
	if !isText {
		r.fault(path, "the value "+jsonform.Plain(value)+" is not a text")
		return "", false
	}
	return text, true
}

func (r *treeReader) optionalText(object map[string]any, path Path, field string) (*string, bool) {
	value, present := object[field]
	if !present {
		r.fault(path, "the field "+field+" is missing")
		return nil, false
	}
	if value == nil {
		return nil, true
	}
	text, isText := value.(string)
	if !isText {
		r.fault(path, "the value "+jsonform.Plain(value)+" is not a text")
		return nil, false
	}
	return &text, true
}

func (r *treeReader) optionalVersion(object map[string]any, path Path) (*int, bool) {
	value, present := object["version"]
	if !present {
		r.fault(path, "the field version is missing")
		return nil, false
	}
	if value == nil {
		return nil, true
	}
	number, isNumber := jsonform.WholeNumber(value)
	if !isNumber {
		r.fault(path, "the value "+jsonform.Plain(value)+" is not a integer")
		return nil, false
	}
	return &number, true
}

func (r *treeReader) child(object map[string]any, path Path, field string, index int) Node {
	value, present := object[field]
	if !present {
		r.fault(path, "the field "+field+" is missing")
		return nil
	}
	return r.part(value, childPath(path, index))
}

func (r *treeReader) children(object map[string]any, path Path, field string, mayBeEmpty bool) ([]Node, bool) {
	value, present := object[field]
	if !present {
		r.fault(path, "the field "+field+" is missing")
		return nil, false
	}
	items, isList := value.([]any)
	if !isList {
		r.fault(path, "the value "+jsonform.Plain(value)+" is not a list")
		return nil, false
	}
	if len(items) == 0 && !mayBeEmpty {
		r.fault(path, "the part has 0 children and needs 1 or more")
		return nil, false
	}
	nodes := make([]Node, 0, len(items))
	complete := true
	for i, item := range items {
		node := r.part(item, childPath(path, i))
		if node == nil {
			complete = false
		}
		nodes = append(nodes, node)
	}
	return nodes, complete
}

func (r *treeReader) notAllowed(object map[string]any, path Path, allowed ...string) {
	for _, name := range jsonform.UnknownKeys(object, allowed...) {
		r.fault(path, "the field "+name+" is not allowed")
	}
}

func (r *treeReader) part(raw any, path Path) Node {
	if r.full() {
		return nil
	}
	object, isObject := raw.(map[string]any)
	if !isObject {
		r.fault(path, notATree)
		return nil
	}
	kind, present := object["kind"]
	if !present {
		r.fault(path, "the field kind is missing")
		return nil
	}
	name, _ := kind.(string)
	switch NodeKind(name) {
	case KindLiteral:
		return r.literal(object, path)
	case KindName:
		return r.name(object, path)
	case KindUnary:
		return r.unary(object, path)
	case KindBinary:
		return r.binary(object, path)
	case KindConditional:
		return r.conditional(object, path)
	case KindCall:
		return r.call(object, path)
	case KindList:
		return r.list(object, path)
	}
	r.fault(path, "the kind "+jsonform.Plain(kind)+" is not in the language")
	return nil
}

func (r *treeReader) literal(object map[string]any, path Path) Node {
	kind, kindOK := r.text(object, path, "type")
	value, valueOK := r.text(object, path, "value")
	r.notAllowed(object, path, "kind", "type", "value")
	if !kindOK || !valueOK {
		return nil
	}
	return Literal{Type: Type(kind), Value: value}
}

func (r *treeReader) name(object map[string]any, path Path) Node {
	name, nameOK := r.text(object, path, "name")
	r.notAllowed(object, path, "kind", "name")
	if !nameOK {
		return nil
	}
	return Name{Name: name}
}

func (r *treeReader) unary(object map[string]any, path Path) Node {
	op, opOK := r.text(object, path, "op")
	operand := r.child(object, path, "operand", 0)
	r.notAllowed(object, path, "kind", "op", "operand")
	if !opOK || operand == nil {
		return nil
	}
	return Unary{Op: op, Operand: operand}
}

func (r *treeReader) binary(object map[string]any, path Path) Node {
	op, opOK := r.text(object, path, "op")
	left := r.child(object, path, "left", 0)
	right := r.child(object, path, "right", 1)
	r.notAllowed(object, path, "kind", "op", "left", "right")
	if !opOK || left == nil || right == nil {
		return nil
	}
	return Binary{Op: op, Left: left, Right: right}
}

func (r *treeReader) conditional(object map[string]any, path Path) Node {
	condition := r.child(object, path, "condition", 0)
	then := r.child(object, path, "then", 1)
	otherwise := r.child(object, path, "else", 2)
	r.notAllowed(object, path, "kind", "condition", "then", "else")
	if condition == nil || then == nil || otherwise == nil {
		return nil
	}
	return Conditional{Condition: condition, Then: then, Else: otherwise}
}

func (r *treeReader) call(object map[string]any, path Path) Node {
	library, libraryOK := r.optionalText(object, path, "library")
	name, nameOK := r.text(object, path, "name")
	version, versionOK := r.optionalVersion(object, path)
	args, argsOK := r.children(object, path, "args", true)
	r.notAllowed(object, path, "kind", "library", "name", "version", "args")
	if !libraryOK || !nameOK || !versionOK || !argsOK {
		return nil
	}
	return Call{Library: library, Name: name, Version: version, Args: args}
}

func (r *treeReader) list(object map[string]any, path Path) Node {
	items, itemsOK := r.children(object, path, "items", false)
	r.notAllowed(object, path, "kind", "items")
	if !itemsOK {
		return nil
	}
	return List{Items: items}
}
