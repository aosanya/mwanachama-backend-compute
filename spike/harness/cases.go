package harness

import (
	"time"

	"github.com/aosanya/mwanachama-backend-compute/spike/numeric"
)

type Example struct {
	Name   string
	Inputs map[string]any
	Want   any
}

type Case struct {
	Number      int
	Title       string
	Source      string
	Params      []Param
	Result      Kind
	PublishedAs string
	Uses        []string
	Examples    []Example
}

type inputs = map[string]any

func dec(text string) numeric.Decimal {
	return numeric.MustParse(text)
}

func decs(texts ...string) []numeric.Decimal {
	values := make([]numeric.Decimal, len(texts))
	for i, text := range texts {
		values[i] = dec(text)
	}
	return values
}

func date(text string) time.Time {
	t, err := time.Parse("2006-01-02", text)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

func Cases() []Case {
	return []Case{
		{
			Number: 1, Title: "Balance after payment",
			Source: `payment >= balance ? 0.00 : balance - payment`,
			Params: []Param{{"balance", KindDecimal}, {"payment", KindDecimal}},
			Result: KindDecimal,
			Examples: []Example{
				{"partial payment", inputs{"balance": dec("120.50"), "payment": dec("45.25")}, dec("75.25")},
				{"payment above balance", inputs{"balance": dec("120.50"), "payment": dec("130.00")}, dec("0.00")},
				{"tenths stay exact", inputs{"balance": dec("0.30"), "payment": dec("0.10")}, dec("0.20")},
				{"payment equals balance", inputs{"balance": dec("120.50"), "payment": dec("120.50")}, dec("0.00")},
			},
		},
		{
			Number: 2, Title: "Loan limit check",
			Source: `status == "active" && requested <= 3 * sum(contributions)`,
			Params: []Param{{"status", KindText}, {"requested", KindDecimal}, {"contributions", KindDecimalList}},
			Result: KindBoolean,
			Examples: []Example{
				{"within three times", inputs{"status": "active", "requested": dec("1400.00"), "contributions": decs("200.00", "250.00", "150.00", "100.00")}, true},
				{"one cent over", inputs{"status": "active", "requested": dec("2100.01"), "contributions": decs("200.00", "250.00", "150.00", "100.00")}, false},
				{"exactly three times", inputs{"status": "active", "requested": dec("2100.00"), "contributions": decs("200.00", "250.00", "150.00", "100.00")}, true},
				{"lapsed member", inputs{"status": "lapsed", "requested": dec("100.00"), "contributions": decs("200.00", "250.00", "150.00", "100.00")}, false},
			},
		},
		{
			Number: 3, Title: "Fund share per member",
			Source: `round((sum(deposits) - sum(expenses)) / members, 2)`,
			Params: []Param{{"deposits", KindDecimalList}, {"expenses", KindDecimalList}, {"members", KindInteger}},
			Result: KindDecimal,
			Examples: []Example{
				{"seven members", inputs{"deposits": decs("500.00", "750.00", "1250.50"), "expenses": decs("120.25", "80.00"), "members": int64(7)}, dec("328.61")},
				{"no expenses", inputs{"deposits": decs("100.00"), "expenses": decs(), "members": int64(3)}, dec("33.33")},
				{"one member", inputs{"deposits": decs("100.00"), "expenses": decs("0.25"), "members": int64(1)}, dec("99.75")},
			},
		},
		{
			Number: 4, Title: "Farmer net payout",
			Source: `max([0.00, round(kilos * price_per_kilo, 2) - credit])`,
			Params: []Param{{"kilos", KindInteger}, {"price_per_kilo", KindDecimal}, {"credit", KindDecimal}},
			Result: KindDecimal,
			Examples: []Example{
				{"credit taken off", inputs{"kilos": int64(853), "price_per_kilo": dec("0.437"), "credit": dec("120.00")}, dec("252.76")},
				{"credit above payout", inputs{"kilos": int64(853), "price_per_kilo": dec("0.437"), "credit": dec("400.00")}, dec("0.00")},
			},
		},
		{
			Number: 5, Title: "Late days", PublishedAs: "late_days@1",
			Source: `max([0, days_between(due, filed)])`,
			Params: []Param{{"due", KindDate}, {"filed", KindDate}},
			Result: KindInteger,
			Examples: []Example{
				{"filed 45 days late", inputs{"due": date("2026-03-31"), "filed": date("2026-05-15")}, int64(45)},
				{"filed early", inputs{"due": date("2026-03-31"), "filed": date("2026-03-20")}, int64(0)},
			},
		},
		{
			Number: 6, Title: "Late-filing penalty", Uses: []string{"late_days@1"},
			Source: `late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)`,
			Params: []Param{{"amount", KindDecimal}, {"due", KindDate}, {"filed", KindDate}},
			Result: KindDecimal,
			Examples: []Example{
				{"45 days late", inputs{"amount": dec("8400.00"), "due": date("2026-03-31"), "filed": date("2026-05-15")}, dec("420.00")},
				{"12 days late", inputs{"amount": dec("8400.00"), "due": date("2026-03-31"), "filed": date("2026-04-12")}, dec("168.00")},
				{"on time", inputs{"amount": dec("8400.00"), "due": date("2026-03-31"), "filed": date("2026-03-31")}, dec("0.00")},
				{"122 days late", inputs{"amount": dec("8400.00"), "due": date("2026-03-31"), "filed": date("2026-07-31")}, dec("840.00")},
			},
		},
		{
			Number: 7, Title: "Interest rate step", PublishedAs: "interest_step@1",
			Source: `balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020`,
			Params: []Param{{"balance", KindDecimal}},
			Result: KindDecimal,
			Examples: []Example{
				{"middle band", inputs{"balance": dec("3217.45")}, dec("0.015")},
				{"lowest band", inputs{"balance": dec("640.00")}, dec("0.010")},
				{"top band at the edge", inputs{"balance": dec("5000.00")}, dec("0.020")},
			},
		},
		{
			Number: 8, Title: "Monthly interest", Uses: []string{"interest_step@1"},
			Source: `round(balance * interest_step@1(balance), 2)`,
			Params: []Param{{"balance", KindDecimal}},
			Result: KindDecimal,
			Examples: []Example{
				{"middle band", inputs{"balance": dec("3217.45")}, dec("48.26")},
				{"lowest band", inputs{"balance": dec("640.00")}, dec("6.40")},
			},
		},
		{
			Number: 9, Title: "Flat-rate loan instalment",
			Source: `round((principal + principal * monthly_rate * months) / months, 2)`,
			Params: []Param{{"principal", KindDecimal}, {"monthly_rate", KindDecimal}, {"months", KindInteger}},
			Result: KindDecimal,
			Examples: []Example{
				{"seven months", inputs{"principal": dec("2400.00"), "monthly_rate": dec("0.012"), "months": int64(7)}, dec("371.66")},
				{"one month", inputs{"principal": dec("2400.00"), "monthly_rate": dec("0.012"), "months": int64(1)}, dec("2428.80")},
			},
		},
		{
			Number: 10, Title: "Term grade",
			Source: `avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"`,
			Params: []Param{{"scores", KindIntegerList}},
			Result: KindText,
			Examples: []Example{
				{"average 77.8", inputs{"scores": []int64{72, 85, 64, 91, 77}}, "B"},
				{"average 87.5", inputs{"scores": []int64{90, 85}}, "A"},
				{"average 55", inputs{"scores": []int64{50, 60}}, "C"},
				{"average 45", inputs{"scores": []int64{40, 50}}, "F"},
			},
		},
		{
			Number: 11, Title: "Pro-rated first-month fee",
			Source: `round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)`,
			Params: []Param{{"monthly_fee", KindDecimal}, {"joined", KindDate}, {"period_end", KindDate}, {"days_in_period", KindInteger}},
			Result: KindDecimal,
			Examples: []Example{
				{"joined on the 12th", inputs{"monthly_fee": dec("47.00"), "joined": date("2026-06-12"), "period_end": date("2026-06-30"), "days_in_period": int64(30)}, dec("29.77")},
				{"joined on the 1st", inputs{"monthly_fee": dec("47.00"), "joined": date("2026-06-01"), "period_end": date("2026-06-30"), "days_in_period": int64(30)}, dec("47.00")},
			},
		},
		{
			Number: 12, Title: "Equipment value after straight-line depreciation",
			Source: `max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])`,
			Params: []Param{{"cost", KindDecimal}, {"salvage", KindDecimal}, {"life_years", KindInteger}, {"bought", KindDate}, {"on", KindDate}},
			Result: KindDecimal,
			Examples: []Example{
				{"three full years", inputs{"cost": dec("12000.00"), "salvage": dec("2000.00"), "life_years": int64(7), "bought": date("2023-09-10"), "on": date("2026-11-17")}, dec("7714.29")},
				{"fully depreciated", inputs{"cost": dec("12000.00"), "salvage": dec("2000.00"), "life_years": int64(7), "bought": date("2023-09-10"), "on": date("2033-09-10")}, dec("2000.00")},
			},
		},
		{
			Number: 13, Title: "Collateral coverage ratio",
			Source: `round(to_float(stock_value) / to_float(facility_balance), 3)`,
			Params: []Param{{"stock_value", KindDecimal}, {"facility_balance", KindDecimal}},
			Result: KindFloat,
			Examples: []Example{
				{"ratio 1.363", inputs{"stock_value": dec("84500.00"), "facility_balance": dec("62000.00")}, 1.363},
				{"ratio 1", inputs{"stock_value": dec("62000.00"), "facility_balance": dec("62000.00")}, 1.0},
			},
		},
		{
			Number: 14, Title: "Interim payment on percent complete",
			Source: `max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])`,
			Params: []Param{{"contract", KindDecimal}, {"percent_complete", KindFloat}, {"paid", KindDecimal}},
			Result: KindDecimal,
			Examples: []Example{
				{"37 percent complete", inputs{"contract": dec("183500.00"), "percent_complete": 0.37, "paid": dec("52000.00")}, dec("9105.50")},
				{"already paid more", inputs{"contract": dec("183500.00"), "percent_complete": 0.2, "paid": dec("52000.00")}, dec("0.00")},
			},
		},
		{
			Number: 15, Title: "Capital call share",
			Source: `round(call_amount * units / sum(all_units), 2)`,
			Params: []Param{{"call_amount", KindDecimal}, {"units", KindInteger}, {"all_units", KindIntegerList}},
			Result: KindDecimal,
			Examples: []Example{
				{"1200 of 4007 units", inputs{"call_amount": dec("50000.00"), "units": int64(1200), "all_units": []int64{1200, 830, 1500, 477}}, dec("14973.80")},
				{"no units held", inputs{"call_amount": dec("50000.00"), "units": int64(0), "all_units": []int64{1200, 830, 1500, 477}}, dec("0.00")},
				{"all units held", inputs{"call_amount": dec("50000.00"), "units": int64(4007), "all_units": []int64{1200, 830, 1500, 477}}, dec("50000.00")},
			},
		},
		{
			Number: 16, Title: "Lending data check",
			Source: `count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00`,
			Params: []Param{{"balances", KindDecimalList}, {"floor", KindDecimal}},
			Result: KindBoolean,
			Examples: []Example{
				{"eight weeks above the floor", inputs{"balances": decs("420.00", "380.50", "95.25", "510.00", "440.75", "300.00", "465.50", "389.00"), "floor": dec("350.00")}, true},
				{"average below the floor", inputs{"balances": decs("420.00", "380.50", "95.25", "510.00", "440.75", "300.00", "465.50", "389.00"), "floor": dec("400.00")}, false},
				{"too few weeks", inputs{"balances": decs("500.00", "500.00", "500.00", "500.00", "500.00"), "floor": dec("100.00")}, false},
				{"one negative week", inputs{"balances": decs("1000.00", "1000.00", "1000.00", "1000.00", "1000.00", "-1.00"), "floor": dec("100.00")}, false},
			},
		},
		{
			Number: 17, Title: "Plot yield spread",
			Source: `max(yields) - min(yields)`,
			Params: []Param{{"yields", KindDecimalList}},
			Result: KindDecimal,
			Examples: []Example{
				{"three plots", inputs{"yields": decs("2400.50", "3100.25", "2750.00")}, dec("699.75")},
				{"one plot", inputs{"yields": decs("100.00")}, dec("0.00")},
			},
		},
		{
			Number: 18, Title: "Voting right",
			Source: `!suspended && role != "guest" && (role == "officer" || (paid_up && months_between(joined, on) >= 6))`,
			Params: []Param{{"joined", KindDate}, {"on", KindDate}, {"paid_up", KindBoolean}, {"suspended", KindBoolean}, {"role", KindText}},
			Result: KindBoolean,
			Examples: []Example{
				{"paid-up member of seven months", inputs{"joined": date("2026-01-20"), "on": date("2026-09-15"), "paid_up": true, "suspended": false, "role": "member"}, true},
				{"suspended", inputs{"joined": date("2026-01-20"), "on": date("2026-09-15"), "paid_up": true, "suspended": true, "role": "member"}, false},
				{"guest", inputs{"joined": date("2026-01-20"), "on": date("2026-09-15"), "paid_up": true, "suspended": false, "role": "guest"}, false},
				{"unpaid officer", inputs{"joined": date("2026-08-01"), "on": date("2026-09-15"), "paid_up": false, "suspended": false, "role": "officer"}, true},
				{"member not paid up", inputs{"joined": date("2026-01-20"), "on": date("2026-09-15"), "paid_up": false, "suspended": false, "role": "member"}, false},
				{"member of four months", inputs{"joined": date("2026-05-01"), "on": date("2026-09-15"), "paid_up": true, "suspended": false, "role": "member"}, false},
			},
		},
		{
			Number: 19, Title: "Seed bags left over",
			Source: `kilos % bag_size`,
			Params: []Param{{"kilos", KindInteger}, {"bag_size", KindInteger}},
			Result: KindInteger,
			Examples: []Example{
				{"37 kilos left", inputs{"kilos": int64(1237), "bag_size": int64(50)}, int64(37)},
				{"nothing left", inputs{"kilos": int64(1200), "bag_size": int64(50)}, int64(0)},
			},
		},
	}
}
