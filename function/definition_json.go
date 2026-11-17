package function

import (
	"bytes"
	"strconv"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/internal/jsonform"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

const CodeBadDefinition = "bad-definition"

const notADefinition = "the text is not a definition"

func DefinitionToJSON(definition Definition) []byte {
	var out bytes.Buffer
	out.WriteString(`{"inputs":[`)
	for i, decl := range definition.Inputs {
		if i > 0 {
			out.WriteString(",")
		}
		writeInput(&out, decl)
	}
	out.WriteString(`],"checks":[`)
	for i, decl := range definition.Checks {
		if i > 0 {
			out.WriteString(",")
		}
		out.WriteString(`{"formula":`)
		out.Write(jsonform.Quote(decl.Formula))
		out.WriteString(`,"message":`)
		out.Write(jsonform.Quote(decl.Message))
		out.WriteString("}")
	}
	out.WriteString("]}")
	return out.Bytes()
}

func writeInput(out *bytes.Buffer, decl InputDecl) {
	out.WriteString(`{"name":`)
	out.Write(jsonform.Quote(decl.Name))
	out.WriteString(`,"type":`)
	out.Write(jsonform.Quote(string(decl.Type)))
	out.WriteString(`,"rules":[`)
	for i, rule := range decl.Rules {
		if i > 0 {
			out.WriteString(",")
		}
		writeRule(out, rule)
	}
	out.WriteString("]}")
}

func writeRule(out *bytes.Buffer, rule Rule) {
	out.WriteString(`{"rule":`)
	out.Write(jsonform.Quote(rule.Name))
	if rule.Items != nil {
		out.WriteString(`,"items":` + strconv.Itoa(*rule.Items))
	}
	if rule.Min != nil {
		out.WriteString(`,"min":`)
		out.Write(jsonform.Quote(*rule.Min))
	}
	if rule.Max != nil {
		out.WriteString(`,"max":`)
		out.Write(jsonform.Quote(*rule.Max))
	}
	if rule.MinExclusive {
		out.WriteString(`,"min_exclusive":true`)
	}
	if rule.MaxExclusive {
		out.WriteString(`,"max_exclusive":true`)
	}
	if rule.Places != nil {
		out.WriteString(`,"places":` + strconv.Itoa(*rule.Places))
	}
	if rule.Values != nil {
		out.WriteString(`,"values":[`)
		for i, value := range rule.Values {
			if i > 0 {
				out.WriteString(",")
			}
			out.Write(jsonform.Quote(value))
		}
		out.WriteString("]")
	}
	out.WriteString("}")
}

func DefinitionFromJSON(data []byte) (Definition, []formula.Problem) {
	reader := &definitionReader{}
	raw, ok := jsonform.Decode(data)
	if !ok {
		reader.fault("", notADefinition)
		return Definition{}, reader.problems()
	}
	definition := reader.definition(raw)
	if len(reader.faults) > 0 {
		return Definition{}, reader.problems()
	}
	return definition, CheckDefinition(definition)
}

type definitionFault struct {
	place string
	what  string
}

type definitionReader struct {
	faults []definitionFault
}

func (r *definitionReader) fault(place string, what string) {
	r.faults = append(r.faults, definitionFault{place, what})
}

func (r *definitionReader) full() bool {
	return len(r.faults) >= maxSaveProblems
}

func (r *definitionReader) problems() []formula.Problem {
	problems := make([]formula.Problem, 0, len(r.faults))
	for _, fault := range r.faults {
		sentence := "The definition is not valid: " + fault.what + "."
		if fault.place != "" {
			sentence = "The definition at " + fault.place + " is not valid: " + fault.what + "."
		}
		problems = append(problems, formula.Problem{Code: CodeBadDefinition, Sentence: sentence, Part: fault.place})
	}
	if len(problems) > maxSaveProblems {
		return problems[:maxSaveProblems]
	}
	return problems
}

func (r *definitionReader) object(raw any, place string) (map[string]any, bool) {
	object, isObject := raw.(map[string]any)
	if !isObject {
		r.fault(place, notADefinition)
	}
	return object, isObject
}

func (r *definitionReader) notAllowed(object map[string]any, place string, allowed ...string) {
	for _, name := range jsonform.UnknownKeys(object, allowed...) {
		r.fault(place, "the field "+name+" is not allowed")
	}
}

func (r *definitionReader) required(object map[string]any, place string, field string) (string, bool) {
	value, present := object[field]
	if !present {
		r.fault(place, "the field "+field+" is missing")
		return "", false
	}
	return r.text(value, place)
}

func (r *definitionReader) text(value any, place string) (string, bool) {
	text, isText := value.(string)
	if !isText {
		r.fault(place, "the value "+jsonform.Plain(value)+" is not a text")
	}
	return text, isText
}

func (r *definitionReader) list(object map[string]any, place string, field string) ([]any, bool) {
	value, present := object[field]
	if !present {
		return nil, false
	}
	items, isList := value.([]any)
	if !isList {
		r.fault(place, "the value "+jsonform.Plain(value)+" is not a list")
		return nil, false
	}
	return items, true
}

func (r *definitionReader) definition(raw any) Definition {
	object, ok := r.object(raw, "")
	if !ok {
		return Definition{}
	}
	r.notAllowed(object, "", "inputs", "checks")
	var definition Definition
	if items, present := r.list(object, "inputs", "inputs"); present {
		for i, item := range items {
			if r.full() {
				break
			}
			definition.Inputs = append(definition.Inputs, r.input(item, "inputs."+strconv.Itoa(i+1)))
		}
	}
	if items, present := r.list(object, "checks", "checks"); present {
		for i, item := range items {
			if r.full() {
				break
			}
			definition.Checks = append(definition.Checks, r.check(item, "checks."+strconv.Itoa(i+1)))
		}
	}
	return definition
}

func (r *definitionReader) input(raw any, place string) InputDecl {
	object, ok := r.object(raw, place)
	if !ok {
		return InputDecl{}
	}
	name, _ := r.required(object, place, "name")
	kind, _ := r.required(object, place, "type")
	r.notAllowed(object, place, "name", "type", "rules")
	decl := InputDecl{Name: name, Type: formula.Type(kind)}
	if items, present := r.list(object, place+".rules", "rules"); present {
		for j, item := range items {
			if r.full() {
				break
			}
			decl.Rules = append(decl.Rules, r.rule(item, place+".rules."+strconv.Itoa(j+1)))
		}
	}
	return decl
}

func (r *definitionReader) check(raw any, place string) CheckDecl {
	object, ok := r.object(raw, place)
	if !ok {
		return CheckDecl{}
	}
	text, _ := r.required(object, place, "formula")
	message, _ := r.required(object, place, "message")
	r.notAllowed(object, place, "formula", "message")
	return CheckDecl{Formula: text, Message: message}
}

var ruleParameters = map[string][]string{
	problem.RuleRequired:      nil,
	problem.RuleNoNegatives:   nil,
	problem.RuleMinItems:      {"items"},
	problem.RuleMaxItems:      {"items"},
	problem.RuleRange:         {"min", "max", "min_exclusive", "max_exclusive"},
	problem.RuleDecimalPlaces: {"places"},
	problem.RuleAllowedValues: {"values"},
}

func allRuleParameters() []string {
	return []string{"items", "min", "max", "min_exclusive", "max_exclusive", "places", "values"}
}

func (r *definitionReader) rule(raw any, place string) Rule {
	object, ok := r.object(raw, place)
	if !ok {
		return Rule{}
	}
	name, _ := r.required(object, place, "rule")
	allowed, known := ruleParameters[name]
	if !known {
		allowed = allRuleParameters()
	}
	r.notAllowed(object, place, append([]string{"rule"}, allowed...)...)
	rule := Rule{Name: name}
	permitted := map[string]bool{}
	for _, parameter := range allowed {
		permitted[parameter] = true
	}
	if permitted["items"] {
		rule.Items = r.whole(object, place, "items")
	}
	if permitted["min"] {
		rule.Min = r.optionalText(object, place, "min")
	}
	if permitted["max"] {
		rule.Max = r.optionalText(object, place, "max")
	}
	if permitted["min_exclusive"] {
		rule.MinExclusive = r.flag(object, place, "min_exclusive")
	}
	if permitted["max_exclusive"] {
		rule.MaxExclusive = r.flag(object, place, "max_exclusive")
	}
	if permitted["places"] {
		rule.Places = r.whole(object, place, "places")
	}
	if permitted["values"] {
		rule.Values = r.values(object, place)
	}
	return rule
}

func (r *definitionReader) whole(object map[string]any, place string, field string) *int {
	value, present := object[field]
	if !present {
		return nil
	}
	number, isWhole := jsonform.WholeNumber(value)
	if !isWhole {
		r.fault(place, "the value "+jsonform.Plain(value)+" is not a whole number")
		return nil
	}
	return &number
}

func (r *definitionReader) optionalText(object map[string]any, place string, field string) *string {
	value, present := object[field]
	if !present {
		return nil
	}
	text, isText := r.text(value, place)
	if !isText {
		return nil
	}
	return &text
}

func (r *definitionReader) flag(object map[string]any, place string, field string) bool {
	value, present := object[field]
	if !present {
		return false
	}
	set, isFlag := value.(bool)
	if !isFlag {
		r.fault(place, "the value "+jsonform.Plain(value)+" is not true or false")
	}
	return set
}

func (r *definitionReader) values(object map[string]any, place string) []string {
	items, present := r.list(object, place, "values")
	if !present {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		if text, isText := r.text(item, place); isText {
			values = append(values, text)
		}
	}
	return values
}
