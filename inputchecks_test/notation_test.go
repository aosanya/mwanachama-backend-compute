package inputchecks

import (
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
	"github.com/aosanya/mwanachama-backend-compute/problem"
)

var ruleOfCode = map[string]string{
	"required":       "required",
	"wrong-type":     "type",
	"too-small":      "range",
	"too-large":      "range",
	"list-too-short": "min-items",
	"list-too-long":  "max-items",
	"no-negatives":   "no-negatives",
	"decimal-places": "decimal-places",
	"not-allowed":    "allowed-values",
	"unknown-input":  "known-input",
	"check-failed":   "check",
	"check-error":    "check",
}

var valueWords = map[string]string{
	"(empty)":      "",
	"one space":    " ",
	"three spaces": "   ",
}

func readable(text string) string {
	text = strings.ReplaceAll(text, "~", "`")
	return strings.ReplaceAll(text, "␣", " ")
}

func line(text string) problem.Problem {
	text = readable(text)
	left, sentence, found := strings.Cut(text, " — ")
	if !found {
		panic("no sentence in: " + text)
	}
	parts := strings.SplitN(left, " · ", 3)
	code := parts[0]
	rule, known := ruleOfCode[code]
	if !known {
		panic("unknown code in: " + text)
	}
	p := problem.Problem{Rule: rule, Sentence: sentence, Code: code}
	if strings.HasPrefix(parts[1], "check \"") {
		formulaText := strings.TrimSuffix(strings.TrimPrefix(parts[1], "check \""), "\"")
		p.Check = &formulaText
		return p
	}
	name, itemText, hasItem := strings.Cut(parts[1], " item ")
	p.Input = &name
	if hasItem {
		item := 0
		for _, digit := range itemText {
			item = item*10 + int(digit-'0')
		}
		p.Item = &item
	}
	given := parts[2] != "not given"
	p.Given = &given
	if given {
		value := parts[2]
		if words, isWords := valueWords[value]; isWords {
			value = words
		}
		p.Value = &value
	}
	return p
}

func lines(texts ...string) []problem.Problem {
	problems := make([]problem.Problem, len(texts))
	for i, text := range texts {
		problems[i] = line(text)
	}
	return problems
}

func row(pairs ...any) function.Row {
	received := function.Row{}
	for i := 0; i < len(pairs); i += 2 {
		name := pairs[i].(string)
		switch value := pairs[i+1].(type) {
		case string:
			received[name] = numtext.One(value)
		case []string:
			received[name] = numtext.Many(value...)
		}
	}
	return received
}

func with(base function.Row, pairs ...any) function.Row {
	changed := function.Row{}
	for name, received := range base {
		changed[name] = received
	}
	for name, received := range row(pairs...) {
		changed[name] = received
	}
	return changed
}

func without(base function.Row, names ...string) function.Row {
	changed := function.Row{}
	for name, received := range base {
		changed[name] = received
	}
	for _, name := range names {
		delete(changed, name)
	}
	return changed
}

func repeat(text string, count int) []string {
	items := make([]string, count)
	for i := range items {
		items[i] = text
	}
	return items
}

func text(value string) *string { return &value }

func places(count int) function.Rule {
	return function.Rule{Name: "decimal-places", Places: &count}
}

func items(name string, count int) function.Rule {
	return function.Rule{Name: name, Items: &count}
}

func between(low string, high string) function.Rule {
	return function.Rule{Name: "range", Min: text(low), Max: text(high)}
}

func over(low string) function.Rule {
	return function.Rule{Name: "range", Min: text(low), MinExclusive: true}
}

func upTo(high string) function.Rule {
	return function.Rule{Name: "range", Max: text(high)}
}

func below(low string, high string) function.Rule {
	return function.Rule{Name: "range", Min: text(low), Max: text(high), MaxExclusive: true}
}

func oneOf(values ...string) function.Rule {
	return function.Rule{Name: "allowed-values", Values: values}
}

var (
	required    = function.Rule{Name: "required"}
	noNegatives = function.Rule{Name: "no-negatives"}
)

func in(name string, kind formula.Type, rules ...function.Rule) function.InputDecl {
	return function.InputDecl{Name: name, Type: kind, Rules: rules}
}

func guard(formulaText string, message string) function.CheckDecl {
	return function.CheckDecl{Formula: formulaText, Message: message}
}

type definition struct {
	name  string
	body  function.Definition
	clean function.Row
}

var (
	d1 = definition{
		name: "D1",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("balance", formula.TypeDecimal, required, noNegatives, places(2)),
				in("payment", formula.TypeDecimal, required, noNegatives, places(2)),
			},
			Checks: []function.CheckDecl{
				guard("payment > 0", "The payment must be more than zero"),
				guard("payment <= balance + 1000.00", "The payment cannot be more than 1000.00 above the balance"),
			},
		},
		clean: row("balance", "120.50", "payment", "45.25"),
	}
	d2 = definition{
		name: "D2",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("kilos", formula.TypeInteger, required, noNegatives),
				in("price_per_kilo", formula.TypeDecimal, required, over("0.000"), places(3)),
				in("credit", formula.TypeDecimal, required, noNegatives, places(2)),
			},
		},
		clean: row("kilos", "853", "price_per_kilo", "0.437", "credit", "120.00"),
	}
	d3 = definition{
		name: "D3",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("status", formula.TypeText, required, oneOf("active", "lapsed", "suspended")),
				in("requested", formula.TypeDecimal, required, over("0.00"), places(2)),
				in("contributions", formula.TypeListDecimal, items("min-items", 1), noNegatives),
			},
		},
		clean: row("status", "active", "requested", "1400.00", "contributions", []string{"200.00", "250.00", "150.00", "100.00"}),
	}
	d4 = definition{
		name: "D4",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("scores", formula.TypeListInteger, items("min-items", 1), items("max-items", 12), between("0", "100")),
			},
		},
		clean: row("scores", []string{"72", "85", "64", "91", "77"}),
	}
	d5 = definition{
		name: "D5",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("balances", formula.TypeListDecimal, items("min-items", 1), places(2)),
				in("floor", formula.TypeDecimal, required, noNegatives, places(2)),
			},
			Checks: []function.CheckDecl{guard("count(balances) >= 6", "At least 6 balances are needed")},
		},
		clean: row("balances", []string{"420.00", "380.50", "95.25", "510.00", "440.75", "300.00"}, "floor", "350.00"),
	}
	d6 = definition{
		name: "D6",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("principal", formula.TypeDecimal, required, over("0.00"), places(2)),
				in("monthly_rate", formula.TypeDecimal, required, between("0.000", "0.050"), places(3)),
				in("months", formula.TypeInteger, required, between("1", "60")),
			},
			Checks: []function.CheckDecl{
				guard("principal >= 100.00", "The loan must be at least 100.00"),
				guard("principal * monthly_rate * months <= principal", "The interest cannot be more than the amount borrowed"),
			},
		},
		clean: row("principal", "2400.00", "monthly_rate", "0.012", "months", "7"),
	}
	d7 = definition{
		name: "D7",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("total", formula.TypeDecimal, required, noNegatives, places(2)),
				in("people", formula.TypeInteger, required),
			},
			Checks: []function.CheckDecl{guard("total / people >= 1.00", "Each person must get at least 1.00")},
		},
		clean: row("total", "100.00", "people", "3"),
	}
	d8 = definition{
		name: "D8",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("kilos", formula.TypeInteger, required),
				in("payment", formula.TypeDecimal, required),
				in("ratio", formula.TypeFloat, required),
				in("note", formula.TypeText, required),
				in("active", formula.TypeBoolean, required),
				in("filed", formula.TypeDate, required),
			},
		},
		clean: row("kilos", "12", "payment", "45.25", "ratio", "0.37", "note", "ok", "active", "true", "filed", "2026-03-31"),
	}
	d9 = definition{
		name: "D9",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("deposits", formula.TypeListDecimal, items("min-items", 1), noNegatives),
				in("expenses", formula.TypeListDecimal, noNegatives),
				in("members", formula.TypeInteger, required, between("1", "500")),
			},
		},
		clean: row("deposits", []string{"500.00", "750.00", "1250.50"}, "expenses", []string{"120.25", "80.00"}, "members", "7"),
	}
	d10 = definition{
		name: "D10",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("units", formula.TypeInteger, required, noNegatives, between("1", "100")),
				in("share_pct", formula.TypeDecimal, required, below("0.00", "100.00")),
				in("bonus", formula.TypeDecimal, required, upTo("500.00")),
				in("bags", formula.TypeDecimal, required, places(0)),
				in("weight", formula.TypeDecimal, required, places(1)),
			},
		},
		clean: row("units", "5", "share_pct", "40.00", "bonus", "100.00", "bags", "3", "weight", "2.5"),
	}
	d11 = definition{
		name: "D11",
		body: function.Definition{
			Inputs: []function.InputDecl{
				in("amounts", formula.TypeListDecimal, items("min-items", 1), noNegatives, between("0.00", "100.00"), places(2)),
			},
		},
		clean: row("amounts", []string{"10.00", "25.50"}),
	}
)

var allDefinitions = []definition{d1, d2, d3, d4, d5, d6, d7, d8, d9, d10, d11}

func TestTheElevenDefinitionsAreAcceptedOnSave(t *testing.T) {
	for _, d := range allDefinitions {
		t.Run(d.name, func(t *testing.T) {
			if problems := function.CheckDefinition(d.body); len(problems) != 0 {
				t.Fatalf("%s is refused on save: %v", d.name, problems)
			}
		})
	}
}

func TestTheCleanRowOfEachDefinitionGivesNoProblem(t *testing.T) {
	for _, d := range allDefinitions {
		t.Run(d.name, func(t *testing.T) {
			got := function.Validate(d.body, d.clean, &engine{})
			if got == nil || len(got) != 0 {
				t.Fatalf("the clean row of %s gives %v", d.name, got)
			}
		})
	}
}

func TestTheNotationReadsALineOfThePage(t *testing.T) {
	got := line("no-negatives · payment · -5.00 — ~payment~ is -5.00. It cannot be negative.")
	if *got.Input != "payment" || got.Item != nil || !*got.Given || *got.Value != "-5.00" || got.Rule != "no-negatives" || got.Check != nil {
		t.Fatalf("read as %+v", got)
	}
	if got.Sentence != "`payment` is -5.00. It cannot be negative." {
		t.Fatalf("sentence %q", got.Sentence)
	}
	item := line("wrong-type · scores item 13 · abc — Item 13 of ~scores~ is abc.")
	if *item.Item != 13 || *item.Input != "scores" {
		t.Fatalf("item read as %+v", item)
	}
	missing := line("required · payment · not given — ~payment~ was not given. It is required.")
	if *missing.Given || missing.Value != nil {
		t.Fatalf("not given read as %+v", missing)
	}
	guardLine := line("check-failed · check \"payment > 0\" — The payment must be more than zero. ~payment~ is 0.00.")
	if guardLine.Input != nil || guardLine.Given != nil || guardLine.Value != nil || *guardLine.Check != "payment > 0" || guardLine.Rule != "check" {
		t.Fatalf("check read as %+v", guardLine)
	}
	space := line("not-allowed · status · active␣ — ~status~ is \"active␣\". It must be one of \"active\".")
	if *space.Value != "active " {
		t.Fatalf("value %q", *space.Value)
	}
}
