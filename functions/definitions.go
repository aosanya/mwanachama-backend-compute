package functions

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

var balanceAfterPayment = Function{
	Library: "fin", Name: "balance_after_payment", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{money("balance"), money("payment")}},
	Body:       "payment >= balance ? 0.00 : balance - payment",
	Result:     formula.TypeDecimal,
}

var equalShare = Function{
	Library: "fin", Name: "equal_share", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		moneyList("deposits", atLeastOne()),
		moneyList("expenses"),
		whole("members", between("1", "500")),
	}},
	Body:   "round((sum(deposits) - sum(expenses)) / members, 2)",
	Result: formula.TypeDecimal,
}

var lateDays = Function{
	Library: "date", Name: "late_days", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{day("due"), day("filed")}},
	Body:       "max([0, days_between(due, filed)])",
	Result:     formula.TypeInteger,
}

var flatInstalment = Function{
	Library: "fin", Name: "flat_instalment", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		positiveMoney("principal"),
		{Name: "monthly_rate", Type: formula.TypeDecimal, Rules: []function.Rule{required(), between("0.000", "0.050"), places(3)}},
		whole("months", between("1", "60")),
	}},
	Body:   "round((principal + principal * monthly_rate * months) / months, 2)",
	Result: formula.TypeDecimal,
}

var prorateFee = Function{
	Library: "fin", Name: "prorate_fee", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		money("monthly_fee"),
		day("joined"),
		day("period_end"),
		whole("days_in_period", between("28", "31")),
	}},
	Body:   "round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)",
	Result: formula.TypeDecimal,
}

var straightLineValue = Function{
	Library: "fin", Name: "straight_line_value", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		positiveMoney("cost"),
		money("salvage"),
		whole("life_years", between("1", "50")),
		day("bought"),
		day("on"),
	}},
	Body:   "max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])",
	Result: formula.TypeDecimal,
}

var coverageRatio = Function{
	Library: "fin", Name: "coverage_ratio", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{money("stock_value"), positiveMoney("facility_balance")}},
	Body:       "round(to_float(stock_value) / to_float(facility_balance), 3)",
	Result:     formula.TypeFloat,
}

var proportionalShare = Function{
	Library: "fin", Name: "proportional_share", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		positiveMoney("call_amount"),
		whole("units", noNegatives()),
		{Name: "all_units", Type: formula.TypeListInteger, Rules: []function.Rule{atLeastOne(), noNegatives()}},
	}},
	Body:   "round(call_amount * units / sum(all_units), 2)",
	Result: formula.TypeDecimal,
}

var spread = Function{
	Library: "stats", Name: "spread", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{moneyList("values", atLeastOne())}},
	Body:       "max(values) - min(values)",
	Result:     formula.TypeDecimal,
}

var remainder = Function{
	Library: "math", Name: "remainder", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		whole("value", noNegatives()),
		whole("divisor", between("1", "100")),
	}},
	Body:   "value % divisor",
	Result: formula.TypeInteger,
}
