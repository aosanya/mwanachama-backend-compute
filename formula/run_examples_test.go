package formula

import "testing"

const (
	runF06 = "late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)"
	runF18 = "!suspended && role != \"guest\" && (role == \"officer\" || paid_up && months_between(joined, on) >= 6)"
)

func TestTheNineteenExampleFormulasGiveTheirResults(t *testing.T) {
	runRows(t, newRunTestRuntime(t), []runRow{
		{"f01 payment short", "payment >= balance ? 0.00 : balance - payment", "balance=dec:120.50; payment=dec:45.25", "dec:75.25"},
		{"f01 payment covers", "payment >= balance ? 0.00 : balance - payment", "balance=dec:120.50; payment=dec:130.00", "dec:0.00"},
		{"f02 loan limit", "status == \"active\" && requested <= 3 * sum(contributions)", "status=text:\"active\"; requested=dec:1400.00; contributions=list<dec>:[200.00, 250.00, 150.00, 100.00]", "bool:true"},
		{"f03 fund share", "round((sum(deposits) - sum(expenses)) / members, 2)", "deposits=list<dec>:[500.00, 750.00, 1250.50]; expenses=list<dec>:[120.25, 80.00]; members=int:7", "dec:328.61"},
		{"f03 fund share, no expenses", "round((sum(deposits) - sum(expenses)) / members, 2)", "deposits=list<dec>:[500.00, 750.00, 1250.50]; expenses=list<dec>:[]; members=int:7", "dec:357.21"},
		{"f04 farmer net payout", "max([0.00, round(kilos * price_per_kilo, 2) - credit])", "kilos=int:853; price_per_kilo=dec:0.437; credit=dec:120.00", "dec:252.76"},
		{"f04 credit over the payout", "max([0.00, round(kilos * price_per_kilo, 2) - credit])", "kilos=int:853; price_per_kilo=dec:0.437; credit=dec:400.00", "dec:0.00"},
		{"f05 late days", "max([0, days_between(due, filed)])", "due=date:2026-03-31; filed=date:2026-05-15", "int:45"},
		{"f05 filed early", "max([0, days_between(due, filed)])", "due=date:2026-03-31; filed=date:2026-03-20", "int:0"},
		{"f06 penalty over thirty days", runF06, "amount=dec:8400.00; due=date:2026-03-31; filed=date:2026-05-15", "dec:420.00"},
		{"f06 penalty under thirty days", runF06, "amount=dec:8400.00; due=date:2026-03-31; filed=date:2026-04-12", "dec:168.00"},
		{"f06 on time", runF06, "amount=dec:8400.00; due=date:2026-03-31; filed=date:2026-03-31", "dec:0.00"},
		{"f06 over ninety days", runF06, "amount=dec:8400.00; due=date:2026-01-01; filed=date:2026-05-15", "dec:840.00"},
		{"f07 middle band", "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020", "balance=dec:3217.45", "dec:0.015"},
		{"f07 low band", "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020", "balance=dec:640.00", "dec:0.010"},
		{"f07 top band", "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020", "balance=dec:5000.00", "dec:0.020"},
		{"f08 monthly interest", "round(balance * interest_step@1(balance), 2)", "balance=dec:3217.45", "dec:48.26"},
		{"f09 flat instalment", "round((principal + principal * monthly_rate * months) / months, 2)", "principal=dec:2400.00; monthly_rate=dec:0.012; months=int:7", "dec:371.66"},
		{"f10 term grade", "avg(scores) >= 80 ? \"A\" : avg(scores) >= 65 ? \"B\" : avg(scores) >= 50 ? \"C\" : \"F\"", "scores=list<int>:[72, 85, 64, 91, 77]", "text:B"},
		{"f11 prorated fee", "round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)", "monthly_fee=dec:47.00; joined=date:2026-06-12; period_end=date:2026-06-30; days_in_period=int:30", "dec:29.77"},
		{"f12 depreciation", "max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])", "cost=dec:12000.00; salvage=dec:2000.00; life_years=int:7; bought=date:2023-09-10; on=date:2026-11-17", "dec:7714.29"},
		{"f13 coverage ratio", "round(to_float(stock_value) / to_float(facility_balance), 3)", "stock_value=dec:84500.00; facility_balance=dec:62000.00", "float:1.363"},
		{"f14 interim payment", "max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])", "contract=dec:183500.00; percent_complete=float:0.37; paid=dec:52000.00", "dec:9105.50"},
		{"f15 capital call share", "round(call_amount * units / sum(all_units), 2)", "call_amount=dec:50000.00; units=int:1200; all_units=list<int>:[1200, 830, 1500, 477]", "dec:14973.80"},
		{"f16 lending data", "count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00", "balances=list<dec>:[420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00]; floor=dec:350.00", "bool:true"},
		{"f17 yield spread", "max(yields) - min(yields)", "yields=list<dec>:[2400.50, 3100.25, 2750.00]", "dec:699.75"},
		{"f18 voting right", runF18, "joined=date:2026-01-20; on=date:2026-09-15; paid_up=bool:true; suspended=bool:false; role=text:\"member\"", "bool:true"},
		{"f18 suspended", runF18, "joined=date:2026-01-20; on=date:2026-09-15; paid_up=bool:true; suspended=bool:true; role=text:\"member\"", "bool:false"},
		{"f18 officer", runF18, "joined=date:2026-09-01; on=date:2026-09-15; paid_up=bool:false; suspended=bool:false; role=text:\"officer\"", "bool:true"},
		{"f19 seed bags left over", "kilos % bag_size", "kilos=int:1237; bag_size=int:50", "int:37"},
	})
}

func TestTheWorkedLinesOfTheLanguagePage(t *testing.T) {
	runRows(t, newRunTestRuntime(t), []runRow{
		{"tight spacing saved text", "payment>=balance?0.00:balance-payment", "balance=dec:120.50; payment=dec:45.25", "dec:75.25"},
		{"loose spacing is formula four", "max( [0.00 ,round( kilos*price_per_kilo , 2 )-credit] )", "kilos=int:853; price_per_kilo=dec:0.437; credit=dec:120.00", "dec:252.76"},
		{"extra brackets are formula eighteen", "(!suspended) && (role != \"guest\") && ((role == \"officer\") || (paid_up && (months_between(joined, on) >= 6)))", "joined=date:2026-01-20; on=date:2026-09-15; paid_up=bool:true; suspended=bool:false; role=text:\"member\"", "bool:true"},
		{"nested conditional", "suspended?paid_up?1:2:3", "suspended=bool:true; paid_up=bool:false", "int:2"},
		{"decimal division has twenty places", "total / people", "total=dec:10.00; people=int:3", "dec:3.33333333333333333333"},
		{"remainder of a negative dividend", "kilos % bag_size", "kilos=int:-7; bag_size=int:3", "int:-1"},
		{"remainder of a negative divisor", "kilos % bag_size", "kilos=int:7; bag_size=int:-3", "int:1"},
		{"equal by value", "first == second", "first=dec:12.5; second=dec:12.50", "bool:true"},
		{"conditional widens the integer answer", "suspended ? 1 : 2.50", "suspended=bool:true", "dec:1"},
		{"conditional keeps the decimal answer", "suspended ? 1 : 2.50", "suspended=bool:false", "dec:2.50"},
		{"text escapes", "note == \"say \\\"hi\\\" \\\\ bye\"", "note=text:\"say \\\"hi\\\" \\\\ bye\"", "bool:true"},
		{"library call version two", "fin.pmt@2(12000, 0.01, 24)", "", "dec:564.88"},
	})
}

func TestRunProblemsNameTheirPartAndSentence(t *testing.T) {
	cases := []struct {
		name     string
		formula  string
		inputs   string
		code     string
		part     string
		path     Path
		sentence string
	}{
		{"divide by zero", "round(total / people, 2)", "total=dec:10.00; people=int:0", "divide-by-zero", "total / people", Path{0}, "The part `total / people` divides by zero."},
		{"no short circuit in the conditional", "people == 0 ? 0.00 : total / people", "total=dec:10.00; people=int:0", "divide-by-zero", "total / people", Path{2}, "The part `total / people` divides by zero."},
		{"integer result too big", "kilos * kilos", "kilos=int:1000000000", "number-too-big", "kilos * kilos", Path{}, "The part `kilos * kilos` gives a number with more than 18 digits."},
		{"empty list", "max(yields)", "yields=list<dec>:[]", "empty-list", "max(yields)", Path{}, "The part `max(yields)` needs a list with at least one item, and the list is empty."},
	}
	runtime := newRunTestRuntime(t)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, problems, failure := runPlanAndRun(t, c.formula, c.inputs, runtime)
			if len(problems) != 0 {
				t.Fatalf("Plan gave problems %+v", problems)
			}
			if failure == nil {
				t.Fatal("the run gave no problem")
			}
			if failure.Code != c.code || failure.Part != c.part || failure.Sentence != c.sentence || len(failure.Path) != len(c.path) {
				t.Fatalf("problem = %+v, want code %s part %q path %v sentence %q", *failure, c.code, c.part, c.path, c.sentence)
			}
			for i := range c.path {
				if failure.Path[i] != c.path[i] {
					t.Fatalf("path = %v, want %v", failure.Path, c.path)
				}
			}
		})
	}
}

func TestAnIntegerResultAtEighteenDigitsIsAccepted(t *testing.T) {
	runRows(t, nil, []runRow{
		{"nine nines squared", "kilos * kilos", "kilos=int:999999999", "int:999999998000000001"},
	})
}
