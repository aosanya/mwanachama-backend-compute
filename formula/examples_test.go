package formula

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

type example struct {
	name   string
	text   string
	inputs []Input
}

var examples = []example{
	{"f01", `payment >= balance ? 0.00 : balance - payment`, inputs("payment", "decimal", "balance", "decimal")},
	{"f02", `status == "active" && requested <= 3 * sum(contributions)`, inputs("status", "text", "requested", "decimal", "contributions", "list of decimal")},
	{"f03", `round((sum(deposits) - sum(expenses)) / members, 2)`, inputs("deposits", "list of decimal", "expenses", "list of decimal", "members", "integer")},
	{"f04", `max([0.00, round(kilos * price_per_kilo, 2) - credit])`, inputs("kilos", "integer", "price_per_kilo", "decimal", "credit", "decimal")},
	{"f05", `max([0, days_between(due, filed)])`, inputs("due", "date", "filed", "date")},
	{"f06", `late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)`, inputs("due", "date", "filed", "date", "amount", "decimal")},
	{"f07", `balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020`, inputs("balance", "decimal")},
	{"f08", `round(balance * interest_step@1(balance), 2)`, inputs("balance", "decimal")},
	{"f09", `round((principal + principal * monthly_rate * months) / months, 2)`, inputs("principal", "decimal", "monthly_rate", "decimal", "months", "integer")},
	{"f10", `avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"`, inputs("scores", "list of integer")},
	{"f11", `round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)`, inputs("monthly_fee", "decimal", "joined", "date", "period_end", "date", "days_in_period", "integer")},
	{"f12", `max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])`, inputs("salvage", "decimal", "cost", "decimal", "life_years", "integer", "bought", "date", "on", "date")},
	{"f13", `round(to_float(stock_value) / to_float(facility_balance), 3)`, inputs("stock_value", "decimal", "facility_balance", "decimal")},
	{"f14", `max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])`, inputs("contract", "decimal", "percent_complete", "float", "paid", "decimal")},
	{"f15", `round(call_amount * units / sum(all_units), 2)`, inputs("call_amount", "decimal", "units", "integer", "all_units", "list of integer")},
	{"f16", `count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00`, inputs("balances", "list of decimal", "floor", "decimal")},
	{"f17", `max(yields) - min(yields)`, inputs("yields", "list of decimal")},
	{"f18", `!suspended && role != "guest" && (role == "officer" || paid_up && months_between(joined, on) >= 6)`, inputs("suspended", "boolean", "role", "text", "paid_up", "boolean", "joined", "date", "on", "date")},
	{"f19", `kilos % bag_size`, inputs("kilos", "integer", "bag_size", "integer")},
}

func TestNineteenExamplesRoundTrip(t *testing.T) {
	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			tree := mustParse(t, ex.text)
			if got := Format(tree); got != ex.text {
				t.Fatalf("Format gave %q, want %q", got, ex.text)
			}
			again := mustParse(t, Format(tree))
			if !reflect.DeepEqual(tree, again) {
				t.Fatalf("Parse(Format(tree)) is not the tree")
			}
		})
	}
}

func TestNineteenExamplesPassCheck(t *testing.T) {
	library := testLibrary()
	for _, ex := range examples {
		t.Run(ex.name, func(t *testing.T) {
			tree := mustParse(t, ex.text)
			if problems := Check(tree, ex.inputs, library); len(problems) != 0 {
				t.Fatalf("Check gave %+v", problems)
			}
		})
	}
}

func TestEightSpellings(t *testing.T) {
	spellings := []struct{ typed, standard string }{
		{`payment>=balance?0.00:balance-payment`, `payment >= balance ? 0.00 : balance - payment`},
		{`payment  >=  balance ? 0.00 :  balance - payment`, `payment >= balance ? 0.00 : balance - payment`},
		{`max( [0.00 ,round( kilos*price_per_kilo , 2 )-credit] )`, `max([0.00, round(kilos * price_per_kilo, 2) - credit])`},
		{`round (amount*0.10,2)`, `round(amount * 0.10, 2)`},
		{`(balance - payment) - credit`, `balance - payment - credit`},
		{`balance - ( payment - credit )`, `balance - (payment - credit)`},
		{`! suspended&&role!="guest"`, `!suspended && role != "guest"`},
		{`suspended?paid_up?1:2:3`, `suspended ? (paid_up ? 1 : 2) : 3`},
	}
	for _, s := range spellings {
		t.Run(s.typed, func(t *testing.T) {
			if got := Format(mustParse(t, s.typed)); got != s.standard {
				t.Fatalf("Format(Parse(%q)) = %q, want %q", s.typed, got, s.standard)
			}
		})
	}
}

func TestExtraBracketsAreDropped(t *testing.T) {
	typed := `(!suspended) && (role != "guest") && ((role == "officer") || (paid_up && (months_between(joined, on) >= 6)))`
	want := examples[17].text
	if got := Format(mustParse(t, typed)); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := Format(mustParse(t, `((a))`)); got != "a" {
		t.Fatalf("((a)) gave %q", got)
	}
}

func TestBracketRules(t *testing.T) {
	cases := map[string]string{
		`a - (b - c)`:             `a - (b - c)`,
		`(a - b) - c`:             `a - b - c`,
		`a || (b && c)`:           `a || b && c`,
		`(a || b) && c`:           `(a || b) && c`,
		`!(a && b)`:               `!(a && b)`,
		`!!a`:                     `!!a`,
		`a < (b < c)`:             `a < (b < c)`,
		`(a ? b : c) ? d : e`:     `(a ? b : c) ? d : e`,
		`a ? b : (c ? d : e)`:     `a ? b : c ? d : e`,
		`(a ? b : c) + 1`:         `(a ? b : c) + 1`,
		`f(a ? b : c)`:            `f(a ? b : c)`,
		`[a ? b : c]`:             `[a ? b : c]`,
		`a * (b + c)`:             `a * (b + c)`,
		`a * (b * c)`:             `a * (b * c)`,
		`(a * b) * c`:             `a * b * c`,
		`fin.pmt@1(a, b)`:         `fin.pmt@1(a, b)`,
		`late_days@1(due, filed)`: `late_days@1(due, filed)`,
		`round (x, 2)`:            `round(x, 2)`,
		`"say \"hi\" \\ now"`:     `"say \"hi\" \\ now"`,
		"a\t+\n b\r\n- c":         `a + b - c`,
		`balance * 0.010 + 12.50`: `balance * 0.010 + 12.50`,
		`x == 12.5 || y != 12.50`: `x == 12.5 || y != 12.50`,
		`true && !false`:          `true && !false`,
	}
	for typed, want := range cases {
		if got := Format(mustParse(t, typed)); got != want {
			t.Errorf("Format(Parse(%q)) = %q, want %q", typed, got, want)
		}
	}
}

func TestDigitsAreNeverRewritten(t *testing.T) {
	a := Format(mustParse(t, `12.5`))
	b := Format(mustParse(t, `12.50`))
	if a == b {
		t.Fatalf("12.5 and 12.50 gave the same text %q", a)
	}
}

func respaced(standard string, rng *rand.Rand) string {
	var out strings.Builder
	inText := false
	for i := 0; i < len(standard); i++ {
		c := standard[i]
		switch {
		case c == '"':
			inText = !inText
			out.WriteByte(c)
		case c == ' ' && !inText:
			switch rng.Intn(3) {
			case 0:
			default:
				for n := rng.Intn(3) + 1; n > 0; n-- {
					out.WriteByte(" \t\n\r"[rng.Intn(4)])
				}
			}
		default:
			out.WriteByte(c)
		}
	}
	return out.String()
}

func TestRespacedExamplesGiveTheSameStandardText(t *testing.T) {
	rng := rand.New(rand.NewSource(20261117))
	for _, ex := range examples {
		for i := 0; i < 20; i++ {
			typed := respaced(ex.text, rng)
			tree, problems := Parse(typed)
			if len(problems) != 0 {
				t.Fatalf("%s: Parse(%q) gave %+v", ex.name, typed, problems)
			}
			if got := Format(tree); got != ex.text {
				t.Fatalf("%s: Format(Parse(%q)) = %q, want %q", ex.name, typed, got, ex.text)
			}
		}
	}
}
