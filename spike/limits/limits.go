package limits

import (
	"fmt"
	"strconv"
	"time"
)

type Limits struct {
	Depth           int
	Nodes           int
	Cost            int
	Time            time.Duration
	FormulaLength   int
	TextLength      int
	ListLength      int
	NumberMagnitude int
	DecimalPlaces   int
}

func Proposed() Limits {
	return Limits{
		Depth:           64,
		Nodes:           1000,
		Cost:            1_000_000,
		Time:            250 * time.Millisecond,
		FormulaLength:   5000,
		TextLength:      1000,
		ListLength:      10_000,
		NumberMagnitude: 18,
		DecimalPlaces:   18,
	}
}

const (
	LimitDepth           = "depth"
	LimitNode            = "node"
	LimitCost            = "cost"
	LimitTime            = "time"
	LimitFormulaLength   = "formula length"
	LimitTextLength      = "text length"
	LimitListLength      = "list length"
	LimitNumberMagnitude = "number magnitude"
	LimitDecimalPlaces   = "decimal places"
)

func (l Limits) Value(limit string) string {
	switch limit {
	case LimitDepth:
		return strconv.Itoa(l.Depth)
	case LimitNode:
		return strconv.Itoa(l.Nodes)
	case LimitCost:
		return strconv.Itoa(l.Cost)
	case LimitTime:
		return l.Time.String()
	case LimitFormulaLength:
		return strconv.Itoa(l.FormulaLength)
	case LimitTextLength:
		return strconv.Itoa(l.TextLength)
	case LimitListLength:
		return strconv.Itoa(l.ListLength)
	case LimitNumberMagnitude:
		return strconv.Itoa(l.NumberMagnitude)
	case LimitDecimalPlaces:
		return strconv.Itoa(l.DecimalPlaces)
	}
	return ""
}

func (l Limits) Refuse(limit string) *Failure {
	return &Failure{Sentence: LimitSentence(limit, l.Value(limit))}
}

func LimitSentence(limit, value string) string {
	return fmt.Sprintf("This formula goes over the %s limit of %s.", limit, value)
}

func ParserSentence(what string, line, column int) string {
	return fmt.Sprintf("The formula cannot be read: %s at line %d, column %d.", what, line, column)
}

type Layer string

const (
	LayerNative   Layer = "native"
	LayerPrecheck Layer = "precheck"
	LayerWrapper  Layer = "wrapper"
	LayerFunction Layer = "function"
	LayerNone     Layer = "none"
)

type Failure struct {
	Sentence string
	Limit    string
	Layer    Layer
	Native   error
}

func (f *Failure) Error() string { return f.Sentence }

func (f *Failure) Unwrap() error { return f.Native }

func (f *Failure) at(layer Layer, native error) *Failure {
	copied := *f
	copied.Layer, copied.Native = layer, native
	return &copied
}

func refusal(limits Limits, limit string, layer Layer, native error) *Failure {
	return &Failure{Sentence: LimitSentence(limit, limits.Value(limit)), Limit: limit, Layer: layer, Native: native}
}

func clip(text string) string {
	if len(text) > 120 {
		return text[:120] + "..."
	}
	return text
}

func unreadable(what string, line, column int, layer Layer, native error) *Failure {
	return &Failure{Sentence: ParserSentence(clip(what), line, column), Limit: "parser", Layer: layer, Native: native}
}

func plainFailure(native error) *Failure {
	return &Failure{Sentence: "This formula could not be run.", Limit: "other", Layer: LayerWrapper, Native: native}
}
