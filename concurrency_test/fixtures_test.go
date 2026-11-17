package concurrency

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/aosanya/mwanachama-backend-compute/datefn"
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/number"
)

type argument struct {
	name  string
	value formula.Value
}

func arg(name string, value formula.Value) argument {
	return argument{name: name, value: value}
}

type formulaCase struct {
	id     int
	name   string
	body   string
	args   []argument
	result string
}

func (c formulaCase) inputs() []formula.Input {
	declared := make([]formula.Input, len(c.args))
	for i, a := range c.args {
		declared[i] = formula.Input{Name: a.name, Type: a.value.Type()}
	}
	return declared
}

func (c formulaCase) values() map[string]formula.Value {
	values := make(map[string]formula.Value, len(c.args))
	for _, a := range c.args {
		values[a.name] = a.value
	}
	return values
}

func dec(text string) formula.Value {
	d, err := number.ParseDecimal(text)
	if err != nil {
		panic("fixture decimal " + text + ": " + err.Error())
	}
	return formula.DecimalValue(d)
}

func whole(n int64) formula.Value { return formula.IntegerValue(n) }

func day(text string) formula.Value {
	d, err := datefn.Parse(text)
	if err != nil {
		panic("fixture date " + text + ": " + err.Error())
	}
	return formula.DateValue(d)
}

func flag(truth bool) formula.Value { return formula.BooleanValue(truth) }

func word(text string) formula.Value { return formula.TextValue(text) }

func approx(v float64) formula.Value {
	f, err := number.NewFloat(v)
	if err != nil {
		panic("fixture float: " + err.Error())
	}
	return formula.FloatValue(f)
}

func decList(texts ...string) formula.Value {
	items := make([]number.Decimal, len(texts))
	for i, text := range texts {
		items[i] = dec(text).Decimal()
	}
	return formula.DecimalListValue(items)
}

func intList(items ...int64) formula.Value { return formula.IntegerListValue(items) }

const (
	lateBody     = "max([0, days_between(due, filed)])"
	stepBody     = "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020"
	penaltyBody  = "late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)"
	interestBody = "round(balance * interest_step@1(balance), 2)"
	votingBody   = "!suspended && role != \"guest\" && (role == \"officer\" || (paid_up && months_between(joined, on) >= 6))"
)

func nineteenFormulas() []formulaCase {
	return []formulaCase{
		{1, "balance after payment", "payment >= balance ? 0.00 : balance - payment",
			[]argument{arg("balance", dec("120.50")), arg("payment", dec("45.25"))}, "decimal:75.25"},
		{2, "loan limit check", "status == \"active\" && requested <= 3 * sum(contributions)",
			[]argument{arg("status", word("active")), arg("requested", dec("1400.00")), arg("contributions", decList("200.00", "250.00", "150.00", "100.00"))}, "boolean:true"},
		{3, "fund share per member", "round((sum(deposits) - sum(expenses)) / members, 2)",
			[]argument{arg("deposits", decList("500.00", "750.00", "1250.50")), arg("expenses", decList("120.25", "80.00")), arg("members", whole(7))}, "decimal:328.61"},
		{4, "farmer net payout", "max([0.00, round(kilos * price_per_kilo, 2) - credit])",
			[]argument{arg("kilos", whole(853)), arg("price_per_kilo", dec("0.437")), arg("credit", dec("120.00"))}, "decimal:252.76"},
		{5, "late days", lateBody,
			[]argument{arg("due", day("2026-03-31")), arg("filed", day("2026-05-15"))}, "integer:45"},
		{6, "late-filing penalty", penaltyBody,
			[]argument{arg("amount", dec("8400.00")), arg("due", day("2026-03-31")), arg("filed", day("2026-05-15"))}, "decimal:420.00"},
		{7, "interest rate step", stepBody,
			[]argument{arg("balance", dec("3217.45"))}, "decimal:0.015"},
		{8, "monthly interest", interestBody,
			[]argument{arg("balance", dec("3217.45"))}, "decimal:48.26"},
		{9, "flat-rate loan instalment", "round((principal + principal * monthly_rate * months) / months, 2)",
			[]argument{arg("principal", dec("2400.00")), arg("monthly_rate", dec("0.012")), arg("months", whole(7))}, "decimal:371.66"},
		{10, "term grade", "avg(scores) >= 80 ? \"A\" : avg(scores) >= 65 ? \"B\" : avg(scores) >= 50 ? \"C\" : \"F\"",
			[]argument{arg("scores", intList(72, 85, 64, 91, 77))}, "text:B"},
		{11, "pro-rated first-month fee", "round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)",
			[]argument{arg("monthly_fee", dec("47.00")), arg("joined", day("2026-06-12")), arg("period_end", day("2026-06-30")), arg("days_in_period", whole(30))}, "decimal:29.77"},
		{12, "straight-line depreciation", "max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])",
			[]argument{arg("cost", dec("12000.00")), arg("salvage", dec("2000.00")), arg("life_years", whole(7)), arg("bought", day("2023-09-10")), arg("on", day("2026-11-17"))}, "decimal:7714.29"},
		{13, "collateral coverage ratio", "round(to_float(stock_value) / to_float(facility_balance), 3)",
			[]argument{arg("stock_value", dec("84500.00")), arg("facility_balance", dec("62000.00"))}, "float:1.363"},
		{14, "interim payment", "max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])",
			[]argument{arg("contract", dec("183500.00")), arg("percent_complete", approx(0.37)), arg("paid", dec("52000.00"))}, "decimal:9105.50"},
		{15, "capital call share", "round(call_amount * units / sum(all_units), 2)",
			[]argument{arg("call_amount", dec("50000.00")), arg("units", whole(1200)), arg("all_units", intList(1200, 830, 1500, 477))}, "decimal:14973.80"},
		{16, "lending data check", "count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00",
			[]argument{arg("balances", decList("420.00", "380.50", "95.25", "510.00", "440.75", "300.00", "465.50", "389.00")), arg("floor", dec("350.00"))}, "boolean:true"},
		{17, "plot yield spread", "max(yields) - min(yields)",
			[]argument{arg("yields", decList("2400.50", "3100.25", "2750.00"))}, "decimal:699.75"},
		{18, "voting right", votingBody,
			[]argument{arg("joined", day("2026-01-20")), arg("on", day("2026-09-15")), arg("paid_up", flag(true)), arg("suspended", flag(false)), arg("role", word("member"))}, "boolean:true"},
		{19, "seed bags left over", "kilos % bag_size",
			[]argument{arg("kilos", whole(1237)), arg("bag_size", whole(50))}, "integer:37"},
	}
}

func formulaNamed(id int) formulaCase {
	for _, c := range nineteenFormulas() {
		if c.id == id {
			return c
		}
	}
	panic("no formula " + strconv.Itoa(id))
}

type agencyRuntime struct {
	signatures map[string]formula.Signature
	programs   map[string]*formula.Program
}

func functionKey(library string, name string, version int) string {
	return library + "." + name + "@" + strconv.Itoa(version)
}

func newAgencyRuntime() agencyRuntime {
	runtime := agencyRuntime{
		signatures: map[string]formula.Signature{},
		programs:   map[string]*formula.Program{},
	}
	runtime.add("late_days", lateBody, formula.TypeInteger, 8, []formula.Input{{Name: "due", Type: formula.TypeDate}, {Name: "filed", Type: formula.TypeDate}})
	runtime.add("interest_step", stepBody, formula.TypeDecimal, 5, []formula.Input{{Name: "balance", Type: formula.TypeDecimal}})
	return runtime
}

func (r agencyRuntime) add(name string, body string, result formula.Type, cost int, inputs []formula.Input) {
	tree, problems := formula.Parse(body)
	if len(problems) > 0 {
		panic("agency function " + name + ": " + problems[0].Sentence)
	}
	program, problems := formula.Plan(tree, inputs, nil, formula.DefaultOptions())
	if len(problems) > 0 {
		panic("agency function " + name + ": " + problems[0].Sentence)
	}
	params := make([]formula.Param, len(inputs))
	for i, input := range inputs {
		params[i] = formula.Param{Name: input.Name, Type: input.Type}
	}
	key := functionKey("", name, 1)
	r.signatures[key] = formula.Signature{Params: params, Result: result, Cost: cost}
	r.programs[key] = program
}

func (r agencyRuntime) Function(library string, name string, version int) (formula.Signature, bool) {
	signature, found := r.signatures[functionKey(library, name, version)]
	return signature, found
}

func (r agencyRuntime) Versions(library string, name string) []int {
	var versions []int
	for version := 1; version < 10; version++ {
		if _, found := r.signatures[functionKey(library, name, version)]; found {
			versions = append(versions, version)
		}
	}
	return versions
}

func (r agencyRuntime) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	program, found := r.programs[functionKey(library, name, version)]
	if !found {
		return formula.Value{}, &formula.Problem{Code: "unknown-function", Sentence: "The function " + functionKey(library, name, version) + " is not known."}
	}
	return program.Apply(args)
}

func planCase(t testing.TB, c formulaCase, runtime formula.Runtime) *formula.Program {
	t.Helper()
	tree, problems := formula.Parse(c.body)
	if len(problems) > 0 {
		t.Fatalf("formula %d %q: %s", c.id, c.name, problems[0].Sentence)
	}
	program, problems := formula.Plan(tree, c.inputs(), runtime, formula.DefaultOptions())
	if len(problems) > 0 {
		t.Fatalf("formula %d %q: %s", c.id, c.name, problems[0].Sentence)
	}
	return program
}

func outcomeKey(value formula.Value, failure *formula.Problem) string {
	if failure != nil {
		return "problem:" + failure.Code + ":" + failure.Sentence
	}
	return string(value.Type()) + ":" + value.String()
}

func valuesKey(values map[string]formula.Value) string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		b.WriteString(name)
		b.WriteByte('=')
		b.WriteString(string(values[name].Type()))
		b.WriteByte(':')
		b.WriteString(values[name].String())
		b.WriteByte(';')
	}
	return b.String()
}

func runKey(program *formula.Program, values map[string]formula.Value) string {
	value, failure := program.Run(values)
	return outcomeKey(value, failure)
}
