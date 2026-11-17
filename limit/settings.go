package limit

import "fmt"

type Name string

const (
	FormulaLengthLimit   Name = "formula length"
	DepthLimit           Name = "depth"
	NodeLimit            Name = "node"
	InputLimit           Name = "input"
	TextLengthLimit      Name = "text length"
	NumberMagnitudeLimit Name = "number magnitude"
	DecimalPlacesLimit   Name = "decimal places"
	NameLengthLimit      Name = "name length"
	CallDepthLimit       Name = "call depth"
	CostLimit            Name = "cost"
	ListLengthLimit      Name = "list length"
)

type Settings struct {
	FormulaLength int
	Depth         int
	BracketsOpen  int
	Nodes         int
	Inputs        int
	CallArguments int
	TextLiteral   int
	IntegerDigits int
	DecimalDigits int
	DecimalPlaces int
	NameLength    int
	CallDepth     int
	Cost          int
	ListItems     int
}

func Proposed() Settings {
	return Settings{
		FormulaLength: 500,
		Depth:         12,
		BracketsOpen:  12,
		Nodes:         100,
		Inputs:        8,
		CallArguments: 8,
		TextLiteral:   200,
		IntegerDigits: 18,
		DecimalDigits: 30,
		DecimalPlaces: 10,
		NameLength:    40,
		CallDepth:     4,
		Cost:          20000,
		ListItems:     1000,
	}
}

type Measure struct {
	FormulaLength int
	Depth         int
	BracketsOpen  int
	Nodes         int
	Inputs        int
	TextLiteral   int
	IntegerDigits int
	DecimalDigits int
	DecimalPlaces int
	NameLength    int
	CallDepth     int
	Cost          int
	ListItems     int
}

type Refusal struct {
	Limit Name
	Value int
}

func (r Refusal) Sentence() string {
	return fmt.Sprintf("This formula goes over the %s limit of %d.", r.Limit, r.Value)
}

func (r Refusal) Error() string {
	return r.Sentence()
}

func (s Settings) AllowsArguments(count int) bool {
	return count <= s.CallArguments
}

func (s Settings) Refusal(m Measure) (Refusal, bool) {
	checks := []struct {
		observed int
		limit    int
		name     Name
	}{
		{m.FormulaLength, s.FormulaLength, FormulaLengthLimit},
		{m.Depth, s.Depth, DepthLimit},
		{m.BracketsOpen, s.BracketsOpen, DepthLimit},
		{m.Nodes, s.Nodes, NodeLimit},
		{m.Inputs, s.Inputs, InputLimit},
		{m.TextLiteral, s.TextLiteral, TextLengthLimit},
		{m.IntegerDigits, s.IntegerDigits, NumberMagnitudeLimit},
		{m.DecimalDigits, s.DecimalDigits, NumberMagnitudeLimit},
		{m.DecimalPlaces, s.DecimalPlaces, DecimalPlacesLimit},
		{m.NameLength, s.NameLength, NameLengthLimit},
		{m.CallDepth, s.CallDepth, CallDepthLimit},
		{m.Cost, s.Cost, CostLimit},
		{m.ListItems, s.ListItems, ListLengthLimit},
	}
	for _, check := range checks {
		if check.observed > check.limit {
			return Refusal{Limit: check.name, Value: check.limit}, true
		}
	}
	return Refusal{}, false
}
