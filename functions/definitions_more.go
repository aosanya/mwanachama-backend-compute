package functions

import (
	"github.com/aosanya/mwanachama-backend-compute/formula"
	"github.com/aosanya/mwanachama-backend-compute/function"
)

var loanWithinLimit = Function{
	Library: "fin", Name: "loan_within_limit", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		word("status", "active", "lapsed", "suspended"),
		positiveMoney("requested"),
		{Name: "contributions", Type: formula.TypeListDecimal, Rules: []function.Rule{atLeastOne(), noNegatives()}},
	}},
	Body:   `status == "active" && requested <= 3 * sum(contributions)`,
	Result: formula.TypeBoolean,
}

var farmerNetPayout = Function{
	Library: "fin", Name: "farmer_net_payout", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		whole("kilos", noNegatives()),
		{Name: "price_per_kilo", Type: formula.TypeDecimal, Rules: []function.Rule{required(), moreThan("0.000"), places(3)}},
		money("credit"),
	}},
	Body:   "max([0.00, round(kilos * price_per_kilo, 2) - credit])",
	Result: formula.TypeDecimal,
}

var latePenalty = Function{
	Library: "fin", Name: "late_filing_penalty", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{money("amount"), day("due"), day("filed")}},
	Body: "date.late_days@1(due, filed) == 0 ? 0.00 : date.late_days@1(due, filed) > 90 ? round(amount * 0.10, 2)" +
		" : date.late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)",
	Result: formula.TypeDecimal,
}

var interestStep = Function{
	Library: "fin", Name: "interest_step", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{money("balance")}},
	Body:       "balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020",
	Result:     formula.TypeDecimal,
}

var monthlyInterest = Function{
	Library: "fin", Name: "monthly_interest", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{money("balance")}},
	Body:       "round(balance * fin.interest_step@1(balance), 2)",
	Result:     formula.TypeDecimal,
}

var termGrade = Function{
	Library: "edu", Name: "term_grade", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{scoreList("scores", "0", "100")}},
	Body:       `avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"`,
	Result:     formula.TypeText,
}

var interimPayment = Function{
	Library: "fin", Name: "interim_payment", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		positiveMoney("contract"),
		ratio("percent_complete", "0.0", "1.0"),
		money("paid"),
	}},
	Body:   "max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])",
	Result: formula.TypeDecimal,
}

var lendingDataOk = Function{
	Library: "fin", Name: "lending_data_ok", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		{Name: "balances", Type: formula.TypeListDecimal, Rules: []function.Rule{atLeastOne(), places(2)}},
		money("floor"),
	}},
	Body:   "count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00",
	Result: formula.TypeBoolean,
}

var votingRight = Function{
	Library: "member", Name: "voting_right", Version: 1,
	Definition: function.Definition{Inputs: []function.InputDecl{
		day("joined"),
		day("on"),
		flag("paid_up"),
		flag("suspended"),
		word("role", "member", "officer", "guest"),
	}},
	Body:   `!suspended && role != "guest" && (role == "officer" || (paid_up && months_between(joined, on) >= 6))`,
	Result: formula.TypeBoolean,
}
