package own

import (
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type Pos struct {
	Line   int
	Column int
}

func (p Pos) Position() Pos { return p }

type Node interface {
	Position() Pos
}

type IntegerLiteral struct {
	Pos
	Value int64
}

type DecimalLiteral struct {
	Pos
	Value numeric.Decimal
}

type FloatLiteral struct {
	Pos
	Value float64
}

type TextLiteral struct {
	Pos
	Value string
}

type DateLiteral struct {
	Pos
	Value time.Time
}

type BooleanLiteral struct {
	Pos
	Value bool
}

type Variable struct {
	Pos
	Name string
}

type List struct {
	Pos
	Elements []Node
}

type Unary struct {
	Pos
	Op      string
	Operand Node
}

type Binary struct {
	Pos
	Op    string
	Left  Node
	Right Node
}

type Conditional struct {
	Pos
	Test Node
	Then Node
	Else Node
}

type Call struct {
	Pos
	Name string
	Args []Node
}
