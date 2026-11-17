package problem

import (
	"fmt"
	"strings"
)

type Subject struct {
	Input string
	Item  int
}

type Range struct {
	Min          *string
	Max          *string
	MinExclusive bool
	MaxExclusive bool
}

type NamedValue struct {
	Name  string
	Value Value
}

func (s Subject) lead() string {
	if s.Item > 0 {
		return fmt.Sprintf("Item %d of `%s`", s.Item, s.Input)
	}
	return "`" + s.Input + "`"
}

func (s Subject) position() *int {
	if s.Item > 0 {
		return number(s.Item)
	}
	return nil
}

func (r Range) bothBoundsInclusive() bool {
	return r.Min != nil && r.Max != nil && !r.MinExclusive && !r.MaxExclusive
}

func (r Range) TooSmallText() string {
	if r.bothBoundsInclusive() {
		return "from " + *r.Min + " to " + *r.Max
	}
	if r.Min == nil {
		return ""
	}
	if r.MinExclusive {
		return "more than " + *r.Min
	}
	return *r.Min + " or more"
}

func (r Range) TooLargeText() string {
	if r.bothBoundsInclusive() {
		return "from " + *r.Min + " to " + *r.Max
	}
	if r.Max == nil {
		return ""
	}
	if r.MaxExclusive {
		return "less than " + *r.Max
	}
	return *r.Max + " or less"
}

func Required(input string) Problem {
	return Problem{
		Input:    text(input),
		Given:    flag(false),
		Rule:     RuleRequired,
		Sentence: "`" + input + "` was not given. It is required.",
		Code:     CodeRequired,
	}
}

func WrongType(s Subject, v Value, typeWords string) Problem {
	return received(s, v, RuleType, CodeWrongType, s.lead()+" is "+v.Sentence()+". It must be "+typeWords+".")
}

func TooSmall(s Subject, v Value, r Range) Problem {
	return received(s, v, RuleRange, CodeTooSmall, s.lead()+" is "+v.Sentence()+". It must be "+r.TooSmallText()+".")
}

func TooLarge(s Subject, v Value, r Range) Problem {
	return received(s, v, RuleRange, CodeTooLarge, s.lead()+" is "+v.Sentence()+". It must be "+r.TooLargeText()+".")
}

func NoNegatives(s Subject, v Value) Problem {
	return received(s, v, RuleNoNegatives, CodeNoNegatives, s.lead()+" is "+v.Sentence()+". It cannot be negative.")
}

func DecimalPlaces(s Subject, v Value, places int) Problem {
	return received(s, v, RuleDecimalPlaces, CodeDecimalPlaces, s.lead()+" is "+v.Sentence()+". "+placesSentence(places))
}

func NotAllowed(s Subject, v Value, allowed []string) Problem {
	quoted := make([]string, len(allowed))
	for i, value := range allowed {
		quoted[i] = `"` + value + `"`
	}
	return received(s, v, RuleAllowedValues, CodeNotAllowed, s.lead()+" is "+v.Sentence()+". It must be one of "+strings.Join(quoted, ", ")+".")
}

func ListTooShort(input string, given bool, v Value, count, items int) Problem {
	opening := fmt.Sprintf("`%s` has %d %s.", input, count, itemWord(count))
	if !given {
		opening = fmt.Sprintf("`%s` was not given, so it has %d %s.", input, count, itemWord(count))
	}
	sentence := fmt.Sprintf("%s It needs at least %d %s.", opening, items, itemWord(items))
	return listProblem(input, given, v, RuleMinItems, CodeListTooShort, sentence)
}

func ListTooLong(input string, v Value, count, items int) Problem {
	sentence := fmt.Sprintf("`%s` has %d %s. It can have at most %d %s.", input, count, itemWord(count), items, itemWord(items))
	return listProblem(input, true, v, RuleMaxItems, CodeListTooLong, sentence)
}

func UnknownInput(input string, v Value, inputs []string) Problem {
	sentence := "`" + input + "` is not an input of this function. Its inputs are: " + strings.Join(inputs, ", ") + "."
	return received(Subject{Input: input}, v, RuleKnownInput, CodeUnknownInput, sentence)
}

func CheckFailed(formula, message string, values []NamedValue) Problem {
	sentence := message + "."
	if len(values) > 0 {
		sentence = message + ". " + namedValues(values) + "."
	}
	return checkProblem(formula, CodeCheckFailed, sentence)
}

func CheckError(formula, message, runSentence string) Problem {
	return checkProblem(formula, CodeCheckError, `The check "`+message+`" could not be worked out: `+runSentence)
}

func received(s Subject, v Value, rule, code, sentence string) Problem {
	return Problem{
		Input:    text(s.Input),
		Item:     s.position(),
		Given:    flag(true),
		Value:    text(v.Field()),
		Rule:     rule,
		Sentence: sentence,
		Code:     code,
	}
}

func listProblem(input string, given bool, v Value, rule, code, sentence string) Problem {
	p := Problem{
		Input:    text(input),
		Given:    flag(given),
		Rule:     rule,
		Sentence: sentence,
		Code:     code,
	}
	if given {
		p.Value = text(v.Field())
	}
	return p
}

func checkProblem(formula, code, sentence string) Problem {
	return Problem{
		Rule:     RuleCheck,
		Sentence: sentence,
		Code:     code,
		Check:    text(formula),
	}
}

func placesSentence(places int) string {
	switch places {
	case 0:
		return "It cannot have decimal places."
	case 1:
		return "It can have at most 1 decimal place."
	}
	return fmt.Sprintf("It can have at most %d decimal places.", places)
}

func itemWord(count int) string {
	if count == 1 {
		return "item"
	}
	return "items"
}

func namedValues(values []NamedValue) string {
	parts := make([]string, len(values))
	for i, named := range values {
		parts[i] = "`" + named.Name + "` is " + named.Value.Sentence()
	}
	if len(parts) == 1 {
		return parts[0]
	}
	last := len(parts) - 1
	return strings.Join(parts[:last], ", ") + " and " + parts[last]
}
