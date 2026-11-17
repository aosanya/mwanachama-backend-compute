package library

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/functions"
	"github.com/aosanya/mwanachama-backend-compute/number"
	"github.com/aosanya/mwanachama-backend-compute/numtext"
)

const referenceFolder = "../functions/reference"

var typedKinds = map[formula.Type]numtext.Type{
	formula.TypeInteger:     numtext.Integer,
	formula.TypeDecimal:     numtext.Decimal,
	formula.TypeFloat:       numtext.Float,
	formula.TypeText:        numtext.Text,
	formula.TypeBoolean:     numtext.Boolean,
	formula.TypeDate:        numtext.Date,
	formula.TypeListInteger: numtext.IntegerList,
	formula.TypeListDecimal: numtext.DecimalList,
}

func typedValue(v numtext.Value) formula.Value {
	switch v.Type {
	case numtext.Integer:
		return formula.IntegerValue(v.Integer)
	case numtext.Decimal:
		return formula.DecimalValue(v.Decimal)
	case numtext.Float:
		return formula.FloatValue(v.Float)
	case numtext.Text:
		return formula.TextValue(v.Text)
	case numtext.Boolean:
		return formula.BooleanValue(v.Boolean)
	case numtext.Date:
		return formula.DateValue(v.Date)
	case numtext.IntegerList:
		items := make([]int64, len(v.Items))
		for i, item := range v.Items {
			items[i] = item.Integer
		}
		return formula.IntegerListValue(items)
	}
	items := make([]number.Decimal, len(v.Items))
	for i, item := range v.Items {
		items[i] = item.Decimal
	}
	return formula.DecimalListValue(items)
}

func loadedRegistry(t *testing.T) *Registry {
	t.Helper()
	registry := NewRegistry()
	if err := LoadFunctions(registry); err != nil {
		t.Fatalf("LoadFunctions: %v", err)
	}
	return registry
}

func TestLoadFunctionsPublishesEveryFunctionAtVersionOne(t *testing.T) {
	registry := loadedRegistry(t)
	for _, f := range functions.All() {
		release, problem := registry.Resolve(Ref{Library: f.Library, Name: f.Name, Version: f.Version})
		if problem != nil {
			t.Errorf("%s: %s", f.Key(), problem.Sentence)
			continue
		}
		if release.Signature.Result != f.Result {
			t.Errorf("%s: result %s, the function says %s", f.Key(), release.Signature.Result, f.Result)
		}
		if len(release.Signature.Params) != len(f.Definition.Inputs) {
			t.Errorf("%s: %d params, the definition has %d", f.Key(), len(release.Signature.Params), len(f.Definition.Inputs))
			continue
		}
		for i, input := range f.Definition.Inputs {
			param := release.Signature.Params[i]
			if param.Name != input.Name || param.Type != input.Type {
				t.Errorf("%s: param %d is %s %s, the definition has %s %s", f.Key(), i, param.Name, param.Type, input.Name, input.Type)
			}
		}
	}
}

func TestLoadingTheSameFunctionsTwiceIsRefused(t *testing.T) {
	registry := loadedRegistry(t)
	if err := LoadFunctions(registry); err == nil {
		t.Fatal("a second load of the same versions was accepted")
	}
}

func TestEveryRecordedCaseGivesTheRecordedAnswerThroughTheRegistry(t *testing.T) {
	registry := loadedRegistry(t)
	versions, err := functions.LoadCases(filepath.FromSlash(referenceFolder))
	if err != nil {
		t.Fatalf("the recorded cases cannot be read: %v", err)
	}
	if len(versions) != len(functions.All()) {
		t.Fatalf("%d functions are in the library and %d data files were read", len(functions.All()), len(versions))
	}
	run, skipped := 0, 0
	for _, key := range functions.SortedKeys(versions) {
		f, found := functions.Find(key)
		if !found {
			t.Fatalf("%s is recorded and the library does not have it", key)
		}
		for _, c := range versions[key] {
			args, typed := argumentsOf(f, c)
			if !typed {
				skipped++
				continue
			}
			value, problem := registry.Apply(f.Library, f.Name, f.Version, args)
			outcome := functions.Outcome{Value: value, Failure: problem}
			if message := functions.Judge(key, c, outcome); message != "" {
				t.Error(message)
			}
			run++
		}
	}
	if run < 60 {
		t.Errorf("only %d recorded cases could be run through the registry (%d could not be typed)", run, skipped)
	}
}

func argumentsOf(f functions.Function, c functions.Case) ([]formula.Value, bool) {
	args := make([]formula.Value, 0, len(f.Definition.Inputs))
	row := functions.RowOf(c)
	for _, input := range f.Definition.Inputs {
		received, given := row[input.Name]
		if !given {
			return nil, false
		}
		read, refusals := numtext.Read(typedKinds[input.Type], received)
		if len(refusals) > 0 {
			return nil, false
		}
		args = append(args, typedValue(read))
	}
	return args, true
}

func TestAFormulaCallsALoadedFunction(t *testing.T) {
	registry := loadedRegistry(t)
	inputs := []formula.Input{
		{Name: "balance", Type: formula.TypeDecimal},
		{Name: "payment", Type: formula.TypeDecimal},
	}
	tree := parsed(t, "fin.balance_after_payment@1(balance, payment)")
	program, problems := formula.Plan(tree, inputs, registry, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("plan: %v", problems)
	}
	value, problem := program.Run(map[string]formula.Value{
		"balance": decimalOf(t, "120.50"),
		"payment": decimalOf(t, "45.25"),
	})
	if problem != nil {
		t.Fatalf("run: %v", problem)
	}
	if value.String() != "75.25" {
		t.Errorf("the formula gave %s, the recorded case says 75.25", value.String())
	}
}

func TestAListArgumentReachesTheFunction(t *testing.T) {
	registry := loadedRegistry(t)
	inputs := []formula.Input{{Name: "values", Type: formula.TypeListDecimal}}
	program, problems := formula.Plan(parsed(t, "stats.spread@1(values)"), inputs, registry, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("plan: %v", problems)
	}
	list := formula.DecimalListValue([]number.Decimal{decimalOf(t, "12.50").Decimal(), decimalOf(t, "3.25").Decimal(), decimalOf(t, "8.00").Decimal()})
	value, problem := program.Run(map[string]formula.Value{"values": list})
	if problem != nil {
		t.Fatalf("run: %v", problem)
	}
	if value.String() != "9.25" {
		t.Errorf("the spread of 12.50, 3.25 and 8.00 is 9.25 and the formula gave %s", value.String())
	}
}

func TestAnInputRuleOfTheFunctionRefusesTheCallInsideAFormula(t *testing.T) {
	registry := loadedRegistry(t)
	inputs := []formula.Input{
		{Name: "value", Type: formula.TypeInteger},
		{Name: "divisor", Type: formula.TypeInteger},
	}
	program, problems := formula.Plan(parsed(t, "math.remainder@1(value, divisor)"), inputs, registry, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("plan: %v", problems)
	}
	_, problem := program.Run(map[string]formula.Value{"value": formula.IntegerValue(10), "divisor": formula.IntegerValue(0)})
	if problem == nil {
		t.Fatal("a divisor of 0 is outside the 1 to 100 the function accepts and the call was not refused")
	}
	if strings.TrimSpace(problem.Sentence) == "" {
		t.Error("the refusal has no sentence")
	}
}

func TestAnUnknownVersionOfALoadedFunctionIsRefusedOnSave(t *testing.T) {
	registry := loadedRegistry(t)
	inputs := []formula.Input{{Name: "values", Type: formula.TypeListDecimal}}
	_, problems := formula.Plan(parsed(t, "stats.spread@2(values)"), inputs, registry, formula.DefaultOptions())
	if len(problems) == 0 {
		t.Fatal("stats.spread@2 does not exist and the formula was accepted")
	}
	if problems[0].Code != "version-not-found" {
		t.Errorf("the code is %s, want version-not-found", problems[0].Code)
	}
}

func TestTheLoadedFunctionsReportTheirCostToTheCaller(t *testing.T) {
	registry := loadedRegistry(t)
	for _, f := range functions.All() {
		signature, found := registry.Function(f.Library, f.Name, f.Version)
		if !found {
			t.Fatalf("%s is not served", f.Key())
		}
		if signature.Cost < 1 {
			t.Errorf("%s declares cost %d", f.Key(), signature.Cost)
		}
	}
}
