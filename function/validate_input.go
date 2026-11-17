package function

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

type inputRules struct {
	required    bool
	minItems    *int
	maxItems    *int
	noNegatives bool
	bounds      *Rule
	places      *int
	allowed     []string
	hasAllowed  bool
}

func rulesOf(decl InputDecl) inputRules {
	var rules inputRules
	for i := range decl.Rules {
		rule := decl.Rules[i]
		switch rule.Name {
		case problem.RuleRequired:
			rules.required = true
		case problem.RuleMinItems:
			rules.minItems = rule.Items
		case problem.RuleMaxItems:
			rules.maxItems = rule.Items
		case problem.RuleNoNegatives:
			rules.noNegatives = true
		case problem.RuleRange:
			rules.bounds = &rule
		case problem.RuleDecimalPlaces:
			rules.places = rule.Places
		case problem.RuleAllowedValues:
			rules.allowed = rule.Values
			rules.hasAllowed = true
		}
	}
	return rules
}

type inputOutcome struct {
	problems []problem.Problem
	value    numtext.Value
}

func (o inputOutcome) failed() bool {
	return len(o.problems) > 0
}

func validateInput(decl InputDecl, received numtext.Received, given bool) inputOutcome {
	info, known := infoOf(decl.Type)
	if !known {
		return inputOutcome{}
	}
	rules := rulesOf(decl)
	if info.list {
		return validateList(decl.Name, info, rules, received, given)
	}
	if !given {
		return inputOutcome{problems: []problem.Problem{problem.Required(decl.Name)}}
	}
	return validateSingle(decl, info, rules, received)
}

func validateSingle(decl InputDecl, info typeInfo, rules inputRules, received numtext.Received) inputOutcome {
	subject := problem.Subject{Input: decl.Name}
	value, refusals := numtext.Read(info.kind, received)
	if len(refusals) > 0 {
		shown := shownSingle(decl.Type, received)
		return inputOutcome{problems: []problem.Problem{problem.WrongType(subject, shown, info.words)}}
	}
	var problems []problem.Problem
	if info.numeric {
		problems = numberProblems(subject, rules, received.Text, problem.PlainText(received.Text))
	}
	if rules.hasAllowed && !contains(rules.allowed, received.Text) {
		problems = append(problems, problem.NotAllowed(subject, problem.QuotedText(received.Text), rules.allowed))
	}
	return inputOutcome{problems: problems, value: value}
}

func validateList(name string, info typeInfo, rules inputRules, received numtext.Received, given bool) inputOutcome {
	if !given {
		return listNotGiven(name, info, rules)
	}
	if !received.List {
		shown := problem.PlainText(received.Text)
		return inputOutcome{problems: []problem.Problem{problem.WrongType(problem.Subject{Input: name}, shown, info.words)}}
	}
	var problems []problem.Problem
	shown := problem.List(received.Items)
	count := len(received.Items)
	if rules.minItems != nil && count < *rules.minItems {
		problems = append(problems, problem.ListTooShort(name, true, shown, count, *rules.minItems))
	}
	if count > itemLimit(rules) {
		problems = append(problems, problem.ListTooLong(name, shown, count, itemLimit(rules)))
	}
	items := received.Items
	if len(items) > maxListItems {
		items = items[:maxListItems]
	}
	for position, text := range items {
		problems = append(problems, itemProblems(name, position+1, info, rules, text)...)
	}
	value := numtext.Value{Type: info.kind}
	if len(problems) == 0 {
		value, _ = numtext.Read(info.kind, received)
	}
	return inputOutcome{problems: problems, value: value}
}

func listNotGiven(name string, info typeInfo, rules inputRules) inputOutcome {
	empty := numtext.Value{Type: info.kind}
	if rules.required {
		return inputOutcome{problems: []problem.Problem{problem.Required(name)}, value: empty}
	}
	if rules.minItems != nil {
		return inputOutcome{problems: []problem.Problem{problem.ListTooShort(name, false, problem.Value{}, 0, *rules.minItems)}, value: empty}
	}
	return inputOutcome{value: empty}
}

func itemLimit(rules inputRules) int {
	if rules.maxItems != nil {
		return *rules.maxItems
	}
	return maxListItems
}

func itemProblems(name string, position int, info typeInfo, rules inputRules, text string) []problem.Problem {
	subject := problem.Subject{Input: name, Item: position}
	if _, refusals := numtext.Read(info.element, numtext.One(text)); len(refusals) > 0 {
		return []problem.Problem{problem.WrongType(subject, problem.PlainText(text), info.itemWords())}
	}
	return numberProblems(subject, rules, text, problem.PlainText(text))
}

func numberProblems(subject problem.Subject, rules inputRules, text string, shown problem.Value) []problem.Problem {
	amount, err := number.ParseDecimal(text)
	if err != nil {
		return nil
	}
	var problems []problem.Problem
	if rules.noNegatives && amount.Sign() < 0 {
		problems = append(problems, problem.NoNegatives(subject, shown))
	}
	if rules.bounds != nil {
		problems = append(problems, rangeProblems(subject, *rules.bounds, amount, shown)...)
	}
	if rules.places != nil && amount.Scale() > *rules.places {
		problems = append(problems, problem.DecimalPlaces(subject, shown, *rules.places))
	}
	return problems
}

func rangeProblems(subject problem.Subject, bounds Rule, amount number.Decimal, shown problem.Value) []problem.Problem {
	described := problem.Range{Min: bounds.Min, Max: bounds.Max, MinExclusive: bounds.MinExclusive, MaxExclusive: bounds.MaxExclusive}
	if bounds.Min != nil {
		if lowest, err := number.ParseDecimal(*bounds.Min); err == nil {
			order := amount.Cmp(lowest)
			if order < 0 || order == 0 && bounds.MinExclusive {
				return []problem.Problem{problem.TooSmall(subject, shown, described)}
			}
		}
	}
	if bounds.Max != nil {
		if highest, err := number.ParseDecimal(*bounds.Max); err == nil {
			order := amount.Cmp(highest)
			if order > 0 || order == 0 && bounds.MaxExclusive {
				return []problem.Problem{problem.TooLarge(subject, shown, described)}
			}
		}
	}
	return nil
}

func shownSingle(kind formula.Type, received numtext.Received) problem.Value {
	if received.List {
		return problem.List(received.Items)
	}
	if kind == formula.TypeText {
		return problem.QuotedText(received.Text)
	}
	return problem.PlainText(received.Text)
}

func contains(values []string, text string) bool {
	for _, value := range values {
		if value == text {
			return true
		}
	}
	return false
}
