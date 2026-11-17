package formula

import (
	"strconv"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type runRow struct {
	name    string
	formula string
	inputs  string
	expect  string
}

func runShortType(t Type) string {
	switch t {
	case TypeInteger:
		return "int"
	case TypeDecimal:
		return "dec"
	case TypeFloat:
		return "float"
	case TypeText:
		return "text"
	case TypeBoolean:
		return "bool"
	case TypeDate:
		return "date"
	case TypeListInteger:
		return "list<int>"
	case TypeListDecimal:
		return "list<dec>"
	}
	return string(t)
}

func runShown(v Value) string {
	return runShortType(v.Type()) + ":" + v.String()
}

func runValue(t *testing.T, text string) Value {
	t.Helper()
	prefix, body, found := strings.Cut(text, ":")
	if !found {
		t.Fatalf("value %q has no type prefix", text)
	}
	switch prefix {
	case "int":
		n, err := strconv.ParseInt(body, 10, 64)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		return IntegerValue(n)
	case "dec":
		d, err := number.ParseDecimal(body)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		return DecimalValue(d)
	case "float":
		f, err := strconv.ParseFloat(body, 64)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		n, err := number.NewFloat(f)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		return FloatValue(n)
	case "text":
		unquoted, err := strconv.Unquote(body)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		return TextValue(unquoted)
	case "bool":
		return BooleanValue(body == "true")
	case "date":
		d, err := datefn.Parse(body)
		if err != nil {
			t.Fatalf("value %q: %v", text, err)
		}
		return DateValue(d)
	case "list<int>":
		var items []int64
		for _, part := range runListParts(body) {
			n, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				t.Fatalf("value %q: %v", text, err)
			}
			items = append(items, n)
		}
		return IntegerListValue(items)
	case "list<dec>":
		var items []number.Decimal
		for _, part := range runListParts(body) {
			d, err := number.ParseDecimal(part)
			if err != nil {
				t.Fatalf("value %q: %v", text, err)
			}
			items = append(items, d)
		}
		return DecimalListValue(items)
	}
	t.Fatalf("value %q has the unknown type %q", text, prefix)
	return Value{}
}

func runListParts(body string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(body, "["), "]")
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

type runInput struct {
	name  string
	value Value
}

func runInputsOf(t *testing.T, text string) []runInput {
	t.Helper()
	var list []runInput
	if strings.TrimSpace(text) == "" {
		return list
	}
	for _, pair := range strings.Split(text, "; ") {
		name, value, found := strings.Cut(pair, "=")
		if !found {
			t.Fatalf("input %q has no =", pair)
		}
		list = append(list, runInput{name: name, value: runValue(t, value)})
	}
	return list
}

func runDeclared(list []runInput) []Input {
	declared := make([]Input, len(list))
	for i, item := range list {
		declared[i] = Input{Name: item.name, Type: item.value.Type()}
	}
	return declared
}

func runValuesOf(list []runInput) map[string]Value {
	values := make(map[string]Value, len(list))
	for _, item := range list {
		values[item.name] = item.value
	}
	return values
}

type runTestRuntime struct {
	fakeLibrary
	programs map[string]*Program
	natives  map[string]func(args []Value) (Value, *Problem)
}

func runKey(library string, name string, version int) string {
	return library + "." + name + "@" + strconv.Itoa(version)
}

func (r runTestRuntime) Apply(library string, name string, version int, args []Value) (Value, *Problem) {
	key := runKey(library, name, version)
	if program, found := r.programs[key]; found {
		return program.Apply(args)
	}
	if native, found := r.natives[key]; found {
		return native(args)
	}
	return Value{}, &Problem{Code: "unknown-function", Sentence: "The function " + key + " is not known."}
}

func newRunTestRuntime(t *testing.T) runTestRuntime {
	t.Helper()
	runtime := runTestRuntime{
		fakeLibrary: testLibrary(),
		programs:    map[string]*Program{},
		natives:     map[string]func(args []Value) (Value, *Problem){},
	}
	late := mustParse(t, "max([0, days_between(due, filed)])")
	lateProgram, problems := Plan(late, inputs("due", "date", "filed", "date"), nil, DefaultOptions())
	if len(problems) != 0 {
		t.Fatalf("late_days did not plan: %+v", problems)
	}
	runtime.programs[runKey("", "late_days", 1)] = lateProgram
	step := mustParse(t, "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020")
	stepProgram, problems := Plan(step, inputs("balance", "decimal"), nil, DefaultOptions())
	if len(problems) != 0 {
		t.Fatalf("interest_step did not plan: %+v", problems)
	}
	runtime.programs[runKey("", "interest_step", 1)] = stepProgram
	runtime.natives[runKey("fin", "pmt", 2)] = runPayment
	return runtime
}

func runPayment(args []Value) (Value, *Problem) {
	principal, _ := args[0].Decimal().ToFloat()
	rate, _ := args[1].Decimal().ToFloat()
	months := float64(args[2].Integer())
	factor := 1.0
	for i := 0.0; i < months; i++ {
		factor /= 1 + rate.Value()
	}
	payment := principal.Value() * rate.Value() / (1 - factor)
	text := strconv.FormatFloat(payment, 'f', 2, 64)
	decimal, _ := number.ParseDecimal(text)
	return DecimalValue(decimal), nil
}

func runPlanAndRun(t *testing.T, formula string, text string, runtime Runtime) (Value, []Problem, *Problem) {
	t.Helper()
	tree := mustParse(t, formula)
	list := runInputsOf(t, text)
	program, problems := Plan(tree, runDeclared(list), runtime, DefaultOptions())
	if len(problems) != 0 {
		return Value{}, problems, nil
	}
	value, failure := program.Run(runValuesOf(list))
	return value, nil, failure
}

func runCodes(problems []Problem) []string {
	codes := make([]string, len(problems))
	for i, problem := range problems {
		codes[i] = problem.Code
	}
	return codes
}

func runRows(t *testing.T, runtime Runtime, rows []runRow) {
	t.Helper()
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			value, problems, failure := runPlanAndRun(t, row.formula, row.inputs, runtime)
			switch {
			case strings.HasPrefix(row.expect, "!"):
				want := strings.Split(strings.TrimPrefix(row.expect, "!"), ",")
				if strings.Join(runCodes(problems), ",") != strings.Join(want, ",") {
					t.Fatalf("Plan problems = %v (run failure %v), want %v", runCodes(problems), failure, want)
				}
			case strings.HasPrefix(row.expect, "?"):
				if len(problems) != 0 {
					t.Fatalf("Plan gave problems %+v", problems)
				}
				if failure == nil || failure.Code != strings.TrimPrefix(row.expect, "?") {
					t.Fatalf("run failure = %+v, want code %s (value %v)", failure, strings.TrimPrefix(row.expect, "?"), runShown(value))
				}
			default:
				if len(problems) != 0 {
					t.Fatalf("Plan gave problems %+v", problems)
				}
				if failure != nil {
					t.Fatalf("run failed with %+v", failure)
				}
				if got := runShown(value); got != row.expect {
					t.Fatalf("result = %s, want %s", got, row.expect)
				}
			}
		})
	}
}
