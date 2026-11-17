package function

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const (
	maxInputs         = 8
	maxChecks         = 8
	maxRulesPerInput  = 8
	maxAllowedValues  = 20
	maxAllowedLength  = 200
	maxPlaces         = 10
	maxMessageLength  = 120
	maxSaveProblems   = 20
	maxInputNameLen   = 40
	bareRuleParameter = "no parameters"
)

const (
	CodeInputRepeated        = "input-repeated"
	CodeBadInputName         = "bad-input-name"
	CodeTooManyInputs        = "too-many-inputs"
	CodeRuleRepeated         = "rule-repeated"
	CodeRuleNotForType       = "rule-not-for-type"
	CodeBadRuleParameter     = "bad-rule-parameter"
	CodePlacesTooMany        = "places-too-many"
	CodeRangeReversed        = "range-reversed"
	CodeItemsReversed        = "items-reversed"
	CodeAllowedValuesEmpty   = "allowed-values-empty"
	CodeAllowedValuesTooMany = "allowed-values-too-many"
	CodeAllowedValuesRepeat  = "allowed-values-repeat"
	CodeRulesContradict      = "rules-contradict"
	CodeInputNeedsRequired   = "input-needs-required"
	CodeTooManyChecks        = "too-many-checks"
	CodeBadMessage           = "bad-message"
	CodeCheckNotBoolean      = "check-not-boolean"
)

var ruleNames = []string{
	problem.RuleRequired, problem.RuleMinItems, problem.RuleMaxItems, problem.RuleNoNegatives,
	problem.RuleRange, problem.RuleDecimalPlaces, problem.RuleAllowedValues,
}

type saver struct {
	problems []formula.Problem
}

func (s *saver) add(code string, part string, sentence string) {
	s.problems = append(s.problems, formula.Problem{Code: code, Sentence: sentence, Part: part})
}

func CheckDefinition(definition Definition) []formula.Problem {
	s := &saver{problems: []formula.Problem{}}
	if len(definition.Inputs) > maxInputs {
		s.add(CodeTooManyInputs, "inputs", fmt.Sprintf("This function goes over the input limit of %d.", maxInputs))
	}
	seen := map[string]bool{}
	for _, decl := range definition.Inputs {
		s.checkInput(decl, seen)
	}
	if len(definition.Checks) > maxChecks {
		s.add(CodeTooManyChecks, "checks", fmt.Sprintf("This function goes over the check limit of %d.", maxChecks))
	}
	inputs := checkableInputs(definition)
	for i, decl := range definition.Checks {
		s.checkCheck(i+1, decl, inputs)
	}
	if len(s.problems) > maxSaveProblems {
		return s.problems[:maxSaveProblems]
	}
	return s.problems
}

func checkableInputs(definition Definition) []formula.Input {
	var inputs []formula.Input
	for _, decl := range definition.Inputs {
		if _, known := infoOf(decl.Type); known {
			inputs = append(inputs, formula.Input{Name: decl.Name, Type: decl.Type})
		}
	}
	return inputs
}

func validInputName(name string) bool {
	if len(name) < 1 || len(name) > maxInputNameLen {
		return false
	}
	for i, character := range name {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		if i == 0 && !letter || !letter && !digit && character != '_' {
			return false
		}
	}
	return true
}

func (s *saver) checkInput(decl InputDecl, seen map[string]bool) {
	part := "inputs." + decl.Name
	if !validInputName(decl.Name) {
		s.add(CodeBadInputName, part, "`"+decl.Name+"` is not a valid input name. A name has 1 to 40 lowercase letters, digits and underscores and starts with a letter.")
	}
	if seen[decl.Name] {
		s.add(CodeInputRepeated, part, "The input `"+decl.Name+"` is declared more than once.")
	}
	seen[decl.Name] = true
	info, known := infoOf(decl.Type)
	if !known {
		s.add(CodeBadRuleParameter, part, "`"+decl.Name+"` has the type "+string(decl.Type)+", which is not a type of this library.")
		return
	}
	if len(decl.Rules) > maxRulesPerInput {
		s.add(CodeBadRuleParameter, part, "`"+decl.Name+"` has more than 8 rules.")
	}
	used := map[string]bool{}
	for _, rule := range decl.Rules {
		s.checkRule(decl, info, rule, used)
	}
	s.checkRuleGroups(decl, info)
	if !info.list && !used[problem.RuleRequired] {
		s.add(CodeInputNeedsRequired, part, "`"+decl.Name+"` is not a list, so it needs the rule required.")
	}
}

func knownRule(name string) bool {
	for _, known := range ruleNames {
		if known == name {
			return true
		}
	}
	return false
}

func ruleApplies(rule string, kind formula.Type, info typeInfo) bool {
	switch rule {
	case problem.RuleRequired:
		return true
	case problem.RuleMinItems, problem.RuleMaxItems:
		return info.list
	case problem.RuleNoNegatives, problem.RuleRange:
		return info.numeric
	case problem.RuleDecimalPlaces:
		return kind == formula.TypeDecimal || kind == formula.TypeListDecimal
	case problem.RuleAllowedValues:
		return kind == formula.TypeText || kind == formula.TypeInteger
	}
	return false
}

func (s *saver) checkRule(decl InputDecl, info typeInfo, rule Rule, used map[string]bool) {
	part := "inputs." + decl.Name + "." + rule.Name
	if !knownRule(rule.Name) {
		s.add(CodeBadRuleParameter, part, "The rule "+rule.Name+" of `"+decl.Name+"` is not a rule of this library.")
		return
	}
	if used[rule.Name] {
		s.add(CodeRuleRepeated, part, "The rule "+rule.Name+" is given twice for `"+decl.Name+"`.")
		return
	}
	used[rule.Name] = true
	if !ruleApplies(rule.Name, decl.Type, info) {
		s.add(CodeRuleNotForType, part, "`"+decl.Name+"` is "+string(decl.Type)+", so it cannot have the rule "+rule.Name+".")
		return
	}
	switch rule.Name {
	case problem.RuleRequired, problem.RuleNoNegatives:
		s.checkBareRule(decl, rule, part)
	case problem.RuleMinItems, problem.RuleMaxItems:
		s.checkItemsRule(decl, rule, part)
	case problem.RuleRange:
		s.checkRangeRule(decl, rule, part)
	case problem.RuleDecimalPlaces:
		s.checkPlacesRule(decl, rule, part)
	case problem.RuleAllowedValues:
		s.checkAllowedRule(decl, rule, part)
	}
}

func (s *saver) badParameter(decl InputDecl, rule Rule, part string, needs string) {
	s.add(CodeBadRuleParameter, part, "The rule "+rule.Name+" of `"+decl.Name+"` needs "+needs+".")
}

func (s *saver) checkBareRule(decl InputDecl, rule Rule, part string) {
	if rule.Min != nil || rule.Max != nil || rule.MinExclusive || rule.MaxExclusive || rule.Places != nil || rule.Values != nil || rule.Items != nil {
		s.badParameter(decl, rule, part, bareRuleParameter)
	}
}

func (s *saver) checkItemsRule(decl InputDecl, rule Rule, part string) {
	if rule.Items == nil || *rule.Items < 1 || *rule.Items > maxListItems {
		s.badParameter(decl, rule, part, "items, a whole number from 1 to 1000")
	}
}

func boundText(text *string) (number.Decimal, bool) {
	if text == nil {
		return number.Decimal{}, false
	}
	if _, refusals := numtext.Read(numtext.Decimal, numtext.One(*text)); len(refusals) > 0 {
		return number.Decimal{}, false
	}
	amount, err := number.ParseDecimal(*text)
	if err != nil || amount.Scale() > maxPlaces {
		return number.Decimal{}, false
	}
	return amount, true
}

func (s *saver) checkRangeRule(decl InputDecl, rule Rule, part string) {
	needs := "a minimum or a maximum written as decimal text"
	lowest, hasLowest := boundText(rule.Min)
	highest, hasHighest := boundText(rule.Max)
	missingBound := rule.Min == nil && rule.Max == nil
	unreadable := rule.Min != nil && !hasLowest || rule.Max != nil && !hasHighest
	danglingEnd := rule.MinExclusive && rule.Min == nil || rule.MaxExclusive && rule.Max == nil
	if missingBound || unreadable || danglingEnd {
		s.badParameter(decl, rule, part, needs)
		return
	}
	if !hasLowest || !hasHighest {
		return
	}
	order := lowest.Cmp(highest)
	if order > 0 || order == 0 && (rule.MinExclusive || rule.MaxExclusive) {
		s.add(CodeRangeReversed, part, "The range of `"+decl.Name+"` has a minimum of "+*rule.Min+" above its maximum of "+*rule.Max+".")
	}
}

func (s *saver) checkPlacesRule(decl InputDecl, rule Rule, part string) {
	if rule.Places == nil || *rule.Places < 0 {
		s.badParameter(decl, rule, part, "places, a whole number from 0 to 10")
		return
	}
	if *rule.Places > maxPlaces {
		s.add(CodePlacesTooMany, part, "The rule decimal-places of `"+decl.Name+"` allows "+strconv.Itoa(*rule.Places)+" places. The most is 10.")
	}
}

func (s *saver) checkAllowedRule(decl InputDecl, rule Rule, part string) {
	switch {
	case len(rule.Values) == 0:
		s.add(CodeAllowedValuesEmpty, part, "The allowed values of `"+decl.Name+"` list no value. A list holds 1 to 20 values.")
		return
	case len(rule.Values) > maxAllowedValues:
		s.add(CodeAllowedValuesTooMany, part, "The allowed values of `"+decl.Name+"` list "+strconv.Itoa(len(rule.Values))+" values. The most is 20.")
		return
	}
	for _, value := range rule.Values {
		if utf8.RuneCountInString(value) > maxAllowedLength {
			s.badParameter(decl, rule, part, "values, a list of texts of at most 200 characters")
			return
		}
	}
	reported := map[string]bool{}
	counted := map[string]int{}
	for _, value := range rule.Values {
		counted[value]++
		if counted[value] == 2 && !reported[value] {
			reported[value] = true
			s.add(CodeAllowedValuesRepeat, part, "The value \""+value+"\" is listed twice for `"+decl.Name+"`.")
		}
	}
}

func ruleOf(decl InputDecl, name string) (Rule, bool) {
	for _, rule := range decl.Rules {
		if rule.Name == name {
			return rule, true
		}
	}
	return Rule{}, false
}

func (s *saver) checkRuleGroups(decl InputDecl, info typeInfo) {
	low, hasLow := ruleOf(decl, problem.RuleMinItems)
	high, hasHigh := ruleOf(decl, problem.RuleMaxItems)
	if info.list && hasLow && hasHigh && low.Items != nil && high.Items != nil && *low.Items > *high.Items {
		s.add(CodeItemsReversed, "inputs."+decl.Name+"."+problem.RuleMinItems,
			"`"+decl.Name+"` needs at least "+strconv.Itoa(*low.Items)+" items and at most "+strconv.Itoa(*high.Items)+".")
	}
	if allowed, has := ruleOf(decl, problem.RuleAllowedValues); has && decl.Type == formula.TypeInteger && len(allowed.Values) > 0 && len(allowed.Values) <= maxAllowedValues {
		s.checkContradictions(decl, allowed)
	}
}

func (s *saver) checkContradictions(decl InputDecl, allowed Rule) {
	rules := rulesOf(decl)
	for _, value := range allowed.Values {
		broken := brokenRule(rules, value)
		if broken != "" {
			s.add(CodeRulesContradict, "inputs."+decl.Name+"."+problem.RuleAllowedValues,
				"No value can meet the rules of `"+decl.Name+"`: the allowed value "+value+" breaks the rule "+broken+".")
		}
	}
}

func brokenRule(rules inputRules, value string) string {
	if _, refusals := numtext.Read(numtext.Integer, numtext.One(value)); len(refusals) > 0 {
		return problem.RuleType
	}
	amount, err := number.ParseDecimal(value)
	if err != nil {
		return problem.RuleType
	}
	if rules.noNegatives && amount.Sign() < 0 {
		return problem.RuleNoNegatives
	}
	if rules.bounds != nil && len(rangeProblems(problem.Subject{Input: "x"}, *rules.bounds, amount, problem.PlainText(value))) > 0 {
		return problem.RuleRange
	}
	return ""
}

func (s *saver) checkCheck(position int, decl CheckDecl, inputs []formula.Input) {
	label := "Check " + strconv.Itoa(position)
	part := "checks." + strconv.Itoa(position)
	if length := utf8.RuneCountInString(decl.Message); length < 1 || length > maxMessageLength || strings.ContainsAny(decl.Message, "\r\n") {
		s.add(CodeBadMessage, part, label+" needs a message of 1 to 120 characters on one line.")
	}
	tree, parseProblems := formula.Parse(decl.Formula)
	if len(parseProblems) > 0 {
		s.addInside(label, parseProblems)
		return
	}
	if found := formula.Check(tree, inputs, nil); len(found) > 0 {
		s.addInside(label, found)
		return
	}
	if kind, ok := formula.ResultType(tree, inputs, nil); ok && kind != formula.TypeBoolean {
		s.add(CodeCheckNotBoolean, part, label+" must be true or false, and `"+formula.Format(tree)+"` is "+string(kind)+".")
	}
}

func (s *saver) addInside(label string, found []formula.Problem) {
	for _, p := range found {
		p.Sentence = label + ": " + p.Sentence
		s.problems = append(s.problems, p)
	}
}
