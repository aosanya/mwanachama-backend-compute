package formula

type Type string

const (
	TypeInteger     Type = "integer"
	TypeDecimal     Type = "decimal"
	TypeFloat       Type = "float"
	TypeText        Type = "text"
	TypeBoolean     Type = "boolean"
	TypeDate        Type = "date"
	TypeListInteger Type = "list of integer"
	TypeListDecimal Type = "list of decimal"
)

type NodeKind string

const (
	KindLiteral     NodeKind = "literal"
	KindName        NodeKind = "name"
	KindUnary       NodeKind = "unary"
	KindBinary      NodeKind = "binary"
	KindConditional NodeKind = "conditional"
	KindCall        NodeKind = "call"
	KindList        NodeKind = "list"
)

type Path []int

type Node interface {
	Kind() NodeKind
	Children() []Node
}

type Literal struct {
	Type  Type
	Value string
}

type Name struct {
	Name string
}

type Unary struct {
	Op      string
	Operand Node
}

type Binary struct {
	Op    string
	Left  Node
	Right Node
}

type Conditional struct {
	Condition Node
	Then      Node
	Else      Node
}

type Call struct {
	Library *string
	Name    string
	Version *int
	Args    []Node
}

type List struct {
	Items []Node
}

type Tree struct {
	Root Node
}

type Part struct {
	Path Path
	Text string
}

type Problem struct {
	Code     string
	Sentence string
	Part     string
	Path     Path
	Line     int
	Column   int
}

type Input struct {
	Name string
	Type Type
}

type Param struct {
	Name string
	Type Type
}

type Signature struct {
	Params    []Param
	Result    Type
	Cost      int
	CallDepth int
}

type Library interface {
	Function(library string, name string, version int) (Signature, bool)
	Versions(library string, name string) []int
}

func (Literal) Kind() NodeKind     { return KindLiteral }
func (Name) Kind() NodeKind        { return KindName }
func (Unary) Kind() NodeKind       { return KindUnary }
func (Binary) Kind() NodeKind      { return KindBinary }
func (Conditional) Kind() NodeKind { return KindConditional }
func (Call) Kind() NodeKind        { return KindCall }
func (List) Kind() NodeKind        { return KindList }

func (Literal) Children() []Node { return nil }
func (Name) Children() []Node    { return nil }

func (u Unary) Children() []Node  { return []Node{u.Operand} }
func (b Binary) Children() []Node { return []Node{b.Left, b.Right} }

func (c Conditional) Children() []Node { return []Node{c.Condition, c.Then, c.Else} }

func (c Call) Children() []Node { return c.Args }
func (l List) Children() []Node { return l.Items }

func PartAt(tree Tree, path Path) (Part, bool) {
	node := tree.Root
	if node == nil {
		return Part{}, false
	}
	for _, index := range path {
		children := node.Children()
		if index < 0 || index >= len(children) || children[index] == nil {
			return Part{}, false
		}
		node = children[index]
	}
	return Part{Path: append(Path{}, path...), Text: Format(Tree{Root: node})}, true
}
