# Formula examples

This file holds one case row for every example on `documentation/2. design/formula-language.md`.
The shared case loader, `internal/casefile`, described in `documentation/4. qa/example-cases.md`, reads the table below.
The rows here and the examples on the design page are the same list, matched by case name.
In a row that expects problems, only the type of each input matters; the values are placeholders. The rows named `rules-` are the exception: they are Validate rows, the received values matter, and the section "Definitions of the Validate rows" says what each is validated against.
Status: the owner has not reviewed this page. All names and figures are invented.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| rules-required-valid | payment >= balance ? 0.00 : balance - payment | balance=text:"120.50", payment=text:"45.25" | result: dec:75.25 | A clean row is calculated with the values as received. | documentation/2. design/formula-language.md |
| rules-required-missing | payment >= balance ? 0.00 : balance - payment | balance=text:"200.00" | problems: required | An input that is not in the row gives required and no other problem. | documentation/2. design/formula-language.md |
| rules-required-and-bad-together | payment >= balance ? 0.00 : balance - payment | payment=text:"-5.00" | problems: required, no-negatives | A missing input and a bad input are both listed, in input order. | documentation/2. design/formula-language.md |
| rules-no-negatives-valid | payment >= balance ? 0.00 : balance - payment | balance=text:"0.00", payment=text:"45.25" | result: dec:0.00 | Zero is not negative, so a balance of 0.00 is accepted. | documentation/2. design/formula-language.md |
| rules-no-negatives-invalid | payment >= balance ? 0.00 : balance - payment | balance=text:"80.00", payment=text:"-5.00" | problems: no-negatives | A value below zero breaks no-negatives, and the checks that read it are skipped. | documentation/2. design/formula-language.md |
| rules-decimal-places-valid | payment >= balance ? 0.00 : balance - payment | balance=text:"120.50", payment=text:"45.2" | result: dec:75.30 | A value with fewer places than the limit is accepted. | documentation/2. design/formula-language.md |
| rules-decimal-places-invalid | payment >= balance ? 0.00 : balance - payment | balance=text:"60.00", payment=text:"5.125" | problems: decimal-places | Three places break a limit of two, and 5.125 is shown as received. | documentation/2. design/formula-language.md |
| rules-three-problems-at-once | payment >= balance ? 0.00 : balance - payment | balance=text:"-1.00", payment=text:"-5.125" | problems: no-negatives, no-negatives, decimal-places | One row gives 3 problems and nothing stops at the first. | documentation/2. design/formula-language.md |
| rules-check-failed | payment >= balance ? 0.00 : balance - payment | balance=text:"120.50", payment=text:"0.00" | problems: check-failed | A check that runs and is false gives check-failed with its own message. | documentation/2. design/formula-language.md |
| rules-check-failed-second | payment >= balance ? 0.00 : balance - payment | balance=text:"100.00", payment=text:"2000.00" | problems: check-failed | The second check fails alone, and the first is true. | documentation/2. design/formula-language.md |
| rules-decimal-places-boundary-valid | max([0.00, round(kilos * price_per_kilo, 2) - credit]) | kilos=text:"853", price_per_kilo=text:"0.437", credit=text:"120.00" | result: dec:252.76 | Three places is the limit and is accepted. | documentation/2. design/formula-language.md |
| rules-decimal-places-trailing-zero | max([0.00, round(kilos * price_per_kilo, 2) - credit]) | kilos=text:"853", price_per_kilo=text:"0.4370", credit=text:"120.00" | problems: decimal-places | A trailing zero counts, because the digits are kept as received. | documentation/2. design/formula-language.md |
| rules-different-codes-at-once | max([0.00, round(kilos * price_per_kilo, 2) - credit]) | kilos=text:"-3", price_per_kilo=text:"0.0005" | problems: no-negatives, decimal-places, required | Three inputs give three different codes, in input order. | documentation/2. design/formula-language.md |
| rules-type-valid | kilos % bag_size | kilos=text:"1237", bag_size=text:"50" | result: int:37 | Whole-number text is read as an integer. | documentation/2. design/formula-language.md |
| rules-type-invalid | kilos % bag_size | kilos=text:"1237.5", bag_size=text:"50" | problems: wrong-type | A number with a dot is not an integer, and it gets no other problem. | documentation/2. design/formula-language.md |
| rules-type-other-inputs-still-tested | kilos % bag_size | kilos=text:"abc", bag_size=text:"0" | problems: wrong-type, too-small | A wrong-type input does not stop the other inputs from being tested. | documentation/2. design/formula-language.md |
| rules-unknown-input | kilos % bag_size | kilos=text:"1237", bag_size=text:"50", bag_weight=text:"3" | problems: unknown-input | A name in the row that is not an input of the function is listed. | documentation/2. design/formula-language.md |
| rules-type-date-valid | max([0, days_between(due, filed)]) | due=text:"2026-03-31", filed=text:"2026-05-15" | result: int:45 | A date written year, month, day is read. | documentation/2. design/formula-language.md |
| rules-type-date-invalid | max([0, days_between(due, filed)]) | due=text:"2026-03-31", filed=text:"2026-02-30" | problems: wrong-type | A day that does not exist is not a date. | documentation/2. design/formula-language.md |
| rules-required-date-missing | max([0, days_between(due, filed)]) | due=text:"2026-03-31" | problems: required | A date input that is not in the row gives required. | documentation/2. design/formula-language.md |
| rules-required-boolean-missing | !suspended && role != "guest" && (role == "officer" \|\| paid_up && months_between(joined, on) >= 6) | joined=text:"2026-01-20", on=text:"2026-09-15", suspended=text:"false", role=text:"member" | problems: required | A boolean input that is not in the row gives required. | documentation/2. design/formula-language.md |
| rules-allowed-values-valid | status == "active" && requested <= 3 * sum(contributions) | status=text:"lapsed", requested=text:"1400.00", contributions=list<text>:["200.00", "250.00", "150.00", "100.00"] | result: bool:false | An allowed text is accepted, and the formula then gives its own answer. | documentation/2. design/formula-language.md |
| rules-allowed-values-invalid | status == "active" && requested <= 3 * sum(contributions) | status=text:"paused", requested=text:"1400.00", contributions=list<text>:["200.00", "250.00", "150.00", "100.00"] | problems: not-allowed | A text outside the allowed list is refused. | documentation/2. design/formula-language.md |
| rules-allowed-values-case-differs | status == "active" && requested <= 3 * sum(contributions) | status=text:"Active", requested=text:"1400.00", contributions=list<text>:["200.00"] | problems: not-allowed | Allowed values compare exact characters, so a different case is refused. | documentation/2. design/formula-language.md |
| rules-range-exclusive-valid | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=text:"0.01", contributions=list<text>:["200.00"] | result: bool:true | A value just above an exclusive minimum is accepted. | documentation/2. design/formula-language.md |
| rules-range-exclusive-invalid | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=text:"0.00", contributions=list<text>:["200.00"] | problems: too-small | The bound itself is refused when the minimum is exclusive. | documentation/2. design/formula-language.md |
| rules-no-negatives-list-item-valid | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=text:"1400.00", contributions=list<text>:["200.00", "250.00", "150.00", "100.00"] | result: bool:true | A rule on a list input is applied to each item, and clean items pass. | documentation/2. design/formula-language.md |
| rules-no-negatives-list-item-invalid | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=text:"1400.00", contributions=list<text>:["200.00", "-50.00"] | problems: no-negatives | The problem names item 2 of the list. | documentation/2. design/formula-language.md |
| rules-list-at-least-one-not-given | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=text:"1400.00" | problems: list-too-short | A list not in the row counts as empty and breaks min-items. | documentation/2. design/formula-language.md |
| rules-list-at-least-one-valid | max(yields) - min(yields) | yields=list<text>:["2400.50", "3100.25", "2750.00"] | result: dec:699.75 | A list with at least one item passes min-items. | documentation/2. design/formula-language.md |
| rules-list-at-least-one-invalid | max(yields) - min(yields) | yields=list<text>:[] | problems: list-too-short | An empty list that is given breaks min-items. | documentation/2. design/formula-language.md |
| rules-decimal-places-list-item-invalid | max(yields) - min(yields) | yields=list<text>:["2400.505", "3100.25"] | problems: decimal-places | The problem names item 1 and shows its text alone. | documentation/2. design/formula-language.md |
| rules-range-valid | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<text>:["500.00", "750.00", "1250.50"], expenses=list<text>:["120.25", "80.00"], members=text:"7" | result: dec:328.61 | A value inside the range is accepted. | documentation/2. design/formula-language.md |
| rules-range-edge-valid | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<text>:["500.00", "750.00", "1250.50"], expenses=list<text>:["120.25", "80.00"], members=text:"500" | result: dec:4.60 | The upper bound itself is accepted when it is inclusive. | documentation/2. design/formula-language.md |
| rules-range-too-small | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<text>:["500.00", "750.00", "1250.50"], expenses=list<text>:["120.25", "80.00"], members=text:"0" | problems: too-small | A value below the minimum gives too-small. | documentation/2. design/formula-language.md |
| rules-range-too-large | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<text>:["500.00", "750.00", "1250.50"], expenses=list<text>:["120.25", "80.00"], members=text:"501" | problems: too-large | A value above the maximum gives too-large. | documentation/2. design/formula-language.md |
| rules-list-not-given-is-empty | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<text>:["500.00", "750.00", "1250.50"], members=text:"7" | result: dec:357.21 | A list without required that is not in the row counts as empty. | documentation/2. design/formula-language.md |
| rules-each-item-valid | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<text>:["72", "85", "64", "91", "77"] | result: text:"B" | Every item is in the range, so the row is calculated. | documentation/2. design/formula-language.md |
| rules-each-item-too-large | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<text>:["72", "104"] | problems: too-large | The range is tested on each item and the problem names item 2. | documentation/2. design/formula-language.md |
| rules-each-item-three-problems | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<text>:["-3", "104", "abc"] | problems: too-small, too-large, wrong-type | Three items give three problems, in item order. | documentation/2. design/formula-language.md |
| rules-list-at-most-valid | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<text>:["70", "70", "70", "70", "70", "70", "70", "70", "70", "70", "70", "70"] | result: text:"B" | A list at the max-items limit is accepted. | documentation/2. design/formula-language.md |
| rules-list-too-long | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<text>:["70", "70", "70", "70", "70", "70", "70", "70", "70", "70", "70", "70", "70"] | problems: list-too-long | A list over the max-items limit gives list-too-long. | documentation/2. design/formula-language.md |
| rules-type-list-item-valid | count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00 | balances=list<text>:["420.00", "380.50", "95.25", "510.00", "440.75", "300.00", "465.50", "389.00"], floor=text:"350.00" | result: bool:true | Every item reads as a decimal, so the row is calculated. | documentation/2. design/formula-language.md |
| rules-type-list-item-invalid | count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00 | balances=list<text>:["420.00", "n/a"], floor=text:"350.00" | problems: wrong-type | A text item in a list of decimal gives wrong-type for that item. | documentation/2. design/formula-language.md |
| rules-check-valid | round((principal + principal * monthly_rate * months) / months, 2) | principal=text:"2400.00", monthly_rate=text:"0.012", months=text:"7" | result: dec:371.66 | Both checks run on a clean row and are true. | documentation/2. design/formula-language.md |
| rules-check-two-failed | round((principal + principal * monthly_rate * months) / months, 2) | principal=text:"50.00", monthly_rate=text:"0.050", months=text:"60" | problems: check-failed, check-failed | Two false checks are both listed, in the order declared. | documentation/2. design/formula-language.md |
| rules-check-skipped-input-problem | round((principal + principal * monthly_rate * months) / months, 2) | principal=text:"50.00", monthly_rate=text:"0.06", months=text:"7" | problems: too-large, check-failed | A check runs when the inputs it reads are clean, and one that reads a bad input is skipped. | documentation/2. design/formula-language.md |
| rules-check-skipped-all | round((principal + principal * monthly_rate * months) / months, 2) | principal=text:"-1.00", monthly_rate=text:"0.012", months=text:"7" | problems: too-small | Both checks read the bad input, so both are skipped and none is invented. | documentation/2. design/formula-language.md |
| rules-check-error | round(total / people, 2) | total=text:"10.00", people=text:"0" | problems: check-error | A check that divides by zero gives check-error, not divide-by-zero. | documentation/2. design/formula-language.md |
| rules-check-division-valid | round(total / people, 2) | total=text:"100.00", people=text:"3" | result: dec:33.33 | A check that works out and is true lets the row be calculated. | documentation/2. design/formula-language.md |
| rules-definition-range-reversed | kilos % bag_size |   | problems: range-reversed | A range with a minimum above its maximum is refused on save. | documentation/2. design/formula-language.md |
| rules-definition-places-too-many | payment >= balance ? 0.00 : balance - payment |   | problems: places-too-many | A decimal-places rule above 10 is refused on save. | documentation/2. design/formula-language.md |
| rules-definition-allowed-values-empty | status == "active" && requested <= 3 * sum(contributions) |   | problems: allowed-values-empty | An allowed-values rule with no value is refused on save. | documentation/2. design/formula-language.md |
| rules-definition-check-unknown-name | payment >= balance ? 0.00 : balance - payment |   | problems: unknown-name | A check that reads a name that is not an input is refused on save. | documentation/2. design/formula-language.md |
| rules-definition-check-not-boolean | payment >= balance ? 0.00 : balance - payment |   | problems: check-not-boolean | A check that is not true or false is refused on save. | documentation/2. design/formula-language.md |
| rules-definition-input-needs-required | kilos % bag_size |   | problems: input-needs-required | An input that is not a list and has no required is refused on save. | documentation/2. design/formula-language.md |
| lang-f01-balance-after-payment | payment >= balance ? 0.00 : balance - payment | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | A conditional gives the difference, or zero when the payment covers the balance. | documentation/2. design/formula-language.md |
| lang-f02-loan-limit-check | status == "active" && requested <= 3 * sum(contributions) | status=text:"active", requested=dec:1400.00, contributions=list<dec>:[200.00, 250.00, 150.00, 100.00] | result: bool:true | A text test and a comparison against a sum of a list join with a boolean and. | documentation/2. design/formula-language.md |
| lang-f03-fund-share-per-member | round((sum(deposits) - sum(expenses)) / members, 2) | deposits=list<dec>:[500.00, 750.00, 1250.50], expenses=list<dec>:[120.25, 80.00], members=int:7 | result: dec:328.61 | The share is divided first and rounded once, to two places, at the end. | documentation/2. design/formula-language.md |
| lang-f04-farmer-net-payout | max([0.00, round(kilos * price_per_kilo, 2) - credit]) | kilos=int:853, price_per_kilo=dec:0.437, credit=dec:120.00 | result: dec:252.76 | A max of zero and the net amount stops a payout from going below zero. | documentation/2. design/formula-language.md |
| lang-f05-late-days | max([0, days_between(due, filed)]) | due=date:2026-03-31, filed=date:2026-05-15 | result: int:45 | Days between two dates is a whole number and max with zero hides an early filing. | documentation/2. design/formula-language.md |
| lang-f06-late-filing-penalty | late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2) | amount=dec:8400.00, due=date:2026-03-31, filed=date:2026-05-15 | result: dec:420.00 | Nested conditionals pick a penalty band from the days late given by a versioned agency call. | documentation/2. design/formula-language.md |
| lang-f07-interest-rate-step | balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020 | balance=dec:3217.45 | result: dec:0.015 | A decimal literal keeps its written digits, so the band rate comes out as 0.015. | documentation/2. design/formula-language.md |
| lang-f08-monthly-interest | round(balance * interest_step@1(balance), 2) | balance=dec:3217.45 | result: dec:48.26 | An agency function result feeds a product, and the rounding is to two places. | documentation/2. design/formula-language.md |
| lang-f09-flat-rate-instalment | round((principal + principal * monthly_rate * months) / months, 2) | principal=dec:2400.00, monthly_rate=dec:0.012, months=int:7 | result: dec:371.66 | A decimal times an integer stays exact, and the quotient is rounded half away from zero. | documentation/2. design/formula-language.md |
| lang-f10-term-grade | avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F" | scores=list<int>:[72, 85, 64, 91, 77] | result: text:"B" | The average of integers is a float, and a float compares with an integer literal. | documentation/2. design/formula-language.md |
| lang-f11-prorated-first-month-fee | round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2) | monthly_fee=dec:47.00, joined=date:2026-06-12, period_end=date:2026-06-30, days_in_period=int:30 | result: dec:29.77 | The joining day is counted by adding one to the days between two dates. | documentation/2. design/formula-language.md |
| lang-f12-straight-line-depreciation | max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)]) | cost=dec:12000.00, salvage=dec:2000.00, life_years=int:7, bought=date:2023-09-10, on=date:2026-11-17 | result: dec:7714.29 | Whole years are counted, the yearly step is rounded first, and the value never drops below salvage. | documentation/2. design/formula-language.md |
| lang-f13-collateral-coverage-ratio | round(to_float(stock_value) / to_float(facility_balance), 3) | stock_value=dec:84500.00, facility_balance=dec:62000.00 | result: float:1.363 | Two floats divide and round to three places, and the result stays a float. | documentation/2. design/formula-language.md |
| lang-f14-interim-payment | max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)]) | contract=dec:183500.00, percent_complete=float:0.37, paid=dec:52000.00 | result: dec:9105.50 | A float input is turned into a decimal with to_decimal before it meets a decimal. | documentation/2. design/formula-language.md |
| lang-f15-capital-call-share | round(call_amount * units / sum(all_units), 2) | call_amount=dec:50000.00, units=int:1200, all_units=list<int>:[1200, 830, 1500, 477] | result: dec:14973.80 | A decimal times an integer is divided by an integer sum, then rounded to two places. | documentation/2. design/formula-language.md |
| lang-f16-lending-data-check | count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00 | balances=list<dec>:[420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00], floor=dec:350.00 | result: bool:true | A list is tested by its count, its average and its minimum, and floor is an ordinary name. | documentation/2. design/formula-language.md |
| lang-f17-plot-yield-spread | max(yields) - min(yields) | yields=list<dec>:[2400.50, 3100.25, 2750.00] | result: dec:699.75 | The greatest item less the least item of one list is exact decimal arithmetic. | documentation/2. design/formula-language.md |
| lang-f18-voting-right | !suspended && role != "guest" && (role == "officer" \|\| paid_up && months_between(joined, on) >= 6) | joined=date:2026-01-20, on=date:2026-09-15, paid_up=bool:true, suspended=bool:false, role=text:"member" | result: bool:true | And binds tighter than or, and months between two dates counts whole months. | documentation/2. design/formula-language.md |
| lang-f19-seed-bags-left-over | kilos % bag_size | kilos=int:1237, bag_size=int:50 | result: int:37 | The remainder operator takes two integers and gives an integer. | documentation/2. design/formula-language.md |
| lang-save-tight-spacing | payment>=balance?0.00:balance-payment | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | Typed text with no spaces is read, and saved with one space around each operator. | documentation/2. design/formula-language.md |
| lang-save-loose-spacing | max( [0.00 ,round( kilos*price_per_kilo , 2 )-credit] ) | kilos=int:853, price_per_kilo=dec:0.437, credit=dec:120.00 | result: dec:252.76 | Spaces around brackets and commas are dropped when the formula is saved. | documentation/2. design/formula-language.md |
| lang-save-extra-brackets | (!suspended) && (role != "guest") && ((role == "officer") \|\| (paid_up && (months_between(joined, on) >= 6))) | joined=date:2026-01-20, on=date:2026-09-15, paid_up=bool:true, suspended=bool:false, role=text:"member" | result: bool:true | Brackets that change nothing are dropped, so this saves as formula 18. | documentation/2. design/formula-language.md |
| lang-save-nested-conditional | suspended?paid_up?1:2:3 | suspended=bool:true, paid_up=bool:false | result: int:2 | A conditional in the then branch is saved with brackets, and the result is unchanged. | documentation/2. design/formula-language.md |
| lang-decimal-division-twenty-places | total / people | total=dec:10.00, people=int:3 | result: dec:3.33333333333333333333 | Division of decimals gives exactly 20 places, rounded half away from zero. | documentation/2. design/formula-language.md |
| lang-remainder-negative-dividend | kilos % bag_size | kilos=int:-7, bag_size=int:3 | result: int:-1 | A remainder has the sign of the dividend: -7 % 3 is -1 and 7 % -3 is 1. | documentation/2. design/formula-language.md |
| lang-decimal-equal-by-value | first == second | first=dec:12.5, second=dec:12.50 | result: bool:true | Comparison reads the value of a decimal and not its written digits. | documentation/2. design/formula-language.md |
| lang-conditional-widens-integer | suspended ? 1 : 2.50 | suspended=bool:true | result: dec:1 | An integer answer and a decimal answer have the common type decimal. | documentation/2. design/formula-language.md |
| lang-text-escapes | note == "say \"hi\" \\ bye" | note=text:"say \"hi\" \\ bye" | result: bool:true | Only \" and \\ are escapes in text, and the quotes are not part of the value. | documentation/2. design/formula-language.md |
| lang-library-call-version-two | fin.pmt@2(12000, 0.01, 24) |   | result: dec:564.88 | A library call names its version, and an integer is accepted for a decimal parameter. | documentation/2. design/formula-language.md |
| lang-problem-unknown-name | balance * rate | balance=dec:100.00, months=int:6 | problems: unknown-name | A name that is not an input of the function is a problem and never a zero. | documentation/2. design/formula-language.md |
| lang-problem-unknown-function | eval("1 + 1") |   | problems: unknown-function | A call to a name that is no built-in, library or agency function is refused. | documentation/2. design/formula-language.md |
| lang-problem-version-missing | fin.pmt(12000, 0.01, 24) |   | problems: version-missing | A library call with no version is refused when the formula is saved. | documentation/2. design/formula-language.md |
| lang-problem-version-not-found | fin.pmt@7(12000, 0.01, 24) |   | problems: version-not-found | A library call to a version that does not exist is refused. | documentation/2. design/formula-language.md |
| lang-problem-version-not-allowed | round@1(balance, 2) | balance=dec:100.00 | problems: version-not-allowed | A built-in function takes no version. | documentation/2. design/formula-language.md |
| lang-problem-wrong-argument-count | round(balance) | balance=dec:100.00 | problems: wrong-argument-count | A call with a different number of arguments than its signature is refused. | documentation/2. design/formula-language.md |
| lang-problem-wrong-argument-type | round(kilos, 2) | kilos=int:853 | problems: wrong-argument-type | An argument of a type the signature does not take is refused, and round takes no integer. | documentation/2. design/formula-language.md |
| lang-problem-type-mismatch | status + 1 | status=text:"active" | problems: type-mismatch | An operator that gets types it does not take is refused. | documentation/2. design/formula-language.md |
| lang-problem-condition-not-boolean | balance ? 0.00 : 1.00 | balance=dec:100.00 | problems: condition-not-boolean | The condition of a conditional must be true or false. | documentation/2. design/formula-language.md |
| lang-problem-branches-differ | suspended ? "yes" : 0 | suspended=bool:true | problems: branches-differ | The two answers of a conditional must have a common type. | documentation/2. design/formula-language.md |
| lang-problem-float-decimal-mix | balance - payment | balance=dec:120.50, payment=float:45.25 | problems: float-decimal-mix | A decimal and a float never meet in an operator without a conversion. | documentation/2. design/formula-language.md |
| lang-problem-list-items-differ | max([0.00, "none"]) |   | problems: list-items-differ | The items of a list written in a formula must have a common type. | documentation/2. design/formula-language.md |
| lang-problem-too-many-inputs | a + b + c + d + e + f + g + h + i | a=int:1, b=int:2, c=int:3, d=int:4, e=int:5, f=int:6, g=int:7, h=int:8, i=int:9 | problems: too-many-inputs | A formula may read at most 8 distinct names. | documentation/2. design/formula-language.md |
| lang-problem-too-costly | sum([sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x)]) | x=list<dec>:[1.00] | problems: too-costly | Cost counts every list as 1000 items, so repeated sums go over the cost limit of 20000. | documentation/2. design/formula-language.md |
| lang-problem-call-too-deep | step_e@1(amount) | amount=dec:100.00 | problems: call-too-deep | A chain of five calls is over the call depth limit of 4. | documentation/2. design/formula-language.md |
| lang-problem-too-deep | !!!!!!!!!!!!suspended | suspended=bool:false | problems: too-deep | A tree of depth 13 is over the depth limit of 12. | documentation/2. design/formula-language.md |
| lang-problem-too-many-nodes | sum([1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1]) |   | problems: too-many-nodes | A tree of 101 nodes is over the node limit of 100. | documentation/2. design/formula-language.md |
| lang-problem-too-long | status == "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" | status=text:"active" | problems: too-long | A text literal of 201 characters is over the text length limit of 200. | documentation/2. design/formula-language.md |
| lang-problem-too-many-digits | 1234567890123456789 + kilos | kilos=int:853 | problems: too-many-digits | An integer literal of 19 digits is over the limit of 18. | documentation/2. design/formula-language.md |
| lang-problem-too-many-places | balance * 0.12345678901 | balance=dec:100.00 | problems: too-many-places | A decimal literal with 11 places is over the limit of 10. | documentation/2. design/formula-language.md |
| lang-problem-bad-syntax-bracket | payment >= (balance ? 0.00 : balance - payment |   | problems: bad-syntax | An unclosed bracket is a syntax problem that names the line and the column. | documentation/2. design/formula-language.md |
| lang-problem-bad-syntax-leading-zero | balance * 007 | balance=dec:100.00 | problems: bad-syntax | A number with a leading zero cannot be read. | documentation/2. design/formula-language.md |
| lang-problem-divide-by-zero | round(total / people, 2) | total=dec:10.00, people=int:0 | problems: divide-by-zero | Dividing by zero is a problem and never a value. | documentation/2. design/formula-language.md |
| lang-no-short-circuit | people == 0 ? 0.00 : total / people | total=dec:10.00, people=int:0 | problems: divide-by-zero | A conditional evaluates both answers, so its condition does not guard the division. | documentation/2. design/formula-language.md |
| lang-problem-number-too-big | kilos * kilos | kilos=int:1000000000 | problems: number-too-big | An integer result of more than 18 digits is a problem. | documentation/2. design/formula-language.md |
| lang-problem-empty-list | max(yields) | yields=list<dec>:[] | problems: empty-list | The max of an empty list is a problem and never a value. | documentation/2. design/formula-language.md |

## Definitions of the Validate rows

The rows named `rules-` are Validate rows. They belong to the section "Input rules, checks and Validate" of the design page. Their cells read differently from the other rows:

- The `inputs` cell holds the row as received, so every value is exact text, written `name=text:"..."`, and a list is written `name=list<text>:["...", "..."]`. A name that is left out is not given. The values matter here, unlike the placeholders of the problem rows above.
- The `expected` cell holds `result: ...` when the row is clean, or `problems: ` and the codes in the order Validate lists them. A code that appears twice is listed twice.
- The `rule` cell holds the rule the row shows.
- The table has no column for the definition the row is validated against. The definitions are listed here, and the second table names the one each row uses.
- A row of the last six cases validates a definition and not a row. Its `inputs` cell is empty and its `expected` cell holds the codes of the problems of the definition.

Open point for the case-table owner: how the loader reads a definition, and how it reads a received value. Proposed: a definition column, or one definition file per page, and a `list<text>` type that the loader adds first. Until then these rows are read with the definitions below. The expected cell compares codes only, so a row that holds the same code on two inputs is told apart in the second table.

Rules are written as in the design page: `required`, `no-negatives`, `decimal-places n`, `range`, `min-items n`, `max-items n` and `allowed-values`. A list rule on a list input is applied to each item.

| Id | Formula | Input | Type | Rules |
|---|---|---|---|---|
| V1 | 1, balance after payment | balance | decimal | required, no-negatives, decimal-places 2 |
| V1 | | payment | decimal | required, no-negatives, decimal-places 2 |
| V2 | 2, loan limit check | status | text | required, allowed-values "active", "lapsed", "suspended" |
| V2 | | requested | decimal | required, range more than 0.00, decimal-places 2 |
| V2 | | contributions | list of decimal | min-items 1, no-negatives |
| V3 | 3, fund share per member | deposits | list of decimal | min-items 1, no-negatives |
| V3 | | expenses | list of decimal | no-negatives |
| V3 | | members | integer | required, range 1 to 500 |
| V4 | 4, farmer net payout | kilos | integer | required, no-negatives |
| V4 | | price_per_kilo | decimal | required, range more than 0.000, decimal-places 3 |
| V4 | | credit | decimal | required, no-negatives, decimal-places 2 |
| V5 | 5, late days | due | date | required |
| V5 | | filed | date | required |
| V9 | 9, flat-rate loan instalment | principal | decimal | required, range more than 0.00, decimal-places 2 |
| V9 | | monthly_rate | decimal | required, range 0.000 to 0.050, decimal-places 3 |
| V9 | | months | integer | required, range 1 to 60 |
| V10 | 10, term grade | scores | list of integer | min-items 1, max-items 12, range 0 to 100 |
| V16 | 16, lending data check | balances | list of decimal | min-items 1, decimal-places 2 |
| V16 | | floor | decimal | required, no-negatives, decimal-places 2 |
| V17 | 17, plot yield spread | yields | list of decimal | min-items 1, no-negatives, decimal-places 2 |
| V18 | 18, voting right | joined | date | required |
| V18 | | on | date | required |
| V18 | | paid_up | boolean | required |
| V18 | | suspended | boolean | required |
| V18 | | role | text | required, allowed-values "member", "officer", "guest" |
| V19 | 19, seed bags left over | kilos | integer | required, no-negatives |
| V19 | | bag_size | integer | required, range 1 to 100 |
| VB | round(total / people, 2) | total | decimal | required, no-negatives, decimal-places 2 |
| VB | | people | integer | required |

The `max-items` of 12 in V10 is added to show the rule. The requirements page gives formula 10 only a minimum of one and a range.

| Id | Check | Formula | Message |
|---|---|---|---|
| V1 | 1 | `payment > 0` | The payment must be more than zero |
| V1 | 2 | `payment <= balance + 1000.00` | The payment cannot be more than 1000.00 above the balance |
| V9 | 1 | `principal >= 100.00` | The loan must be at least 100.00 |
| V9 | 2 | `principal * monthly_rate * months <= principal` | The interest cannot be more than the amount borrowed |
| VB | 1 | `total / people >= 1.00` | Each person must get at least 1.00 |

V2, V3, V4, V5, V10, V16, V17, V18 and V19 have no checks.

The definition each row is validated against, and the problems it lists in order. A problem is written as the input, the item position when it is a list item, and the code. A check is written `check n`. A check that is skipped gives no problem and is named here only to show why.

| Case | Definition | Problems in order, as input, item and code, or calculated |
|---|---|---|
| rules-required-valid | V1 | none; calculated |
| rules-required-missing | V1 | payment required |
| rules-required-and-bad-together | V1 | balance required; payment no-negatives |
| rules-no-negatives-valid | V1 | none; calculated |
| rules-no-negatives-invalid | V1 | payment no-negatives; checks 1 and 2 skipped |
| rules-decimal-places-valid | V1 | none; calculated |
| rules-decimal-places-invalid | V1 | payment decimal-places |
| rules-three-problems-at-once | V1 | balance no-negatives; payment no-negatives; payment decimal-places |
| rules-check-failed | V1 | check 1 check-failed |
| rules-check-failed-second | V1 | check 2 check-failed |
| rules-decimal-places-boundary-valid | V4 | none; calculated |
| rules-decimal-places-trailing-zero | V4 | price_per_kilo decimal-places |
| rules-different-codes-at-once | V4 | kilos no-negatives; price_per_kilo decimal-places; credit required |
| rules-type-valid | V19 | none; calculated |
| rules-type-invalid | V19 | kilos wrong-type |
| rules-type-other-inputs-still-tested | V19 | kilos wrong-type; bag_size too-small |
| rules-unknown-input | V19 | bag_weight unknown-input |
| rules-type-date-valid | V5 | none; calculated |
| rules-type-date-invalid | V5 | filed wrong-type |
| rules-required-date-missing | V5 | filed required |
| rules-required-boolean-missing | V18 | paid_up required |
| rules-allowed-values-valid | V2 | none; calculated |
| rules-allowed-values-invalid | V2 | status not-allowed |
| rules-allowed-values-case-differs | V2 | status not-allowed |
| rules-range-exclusive-valid | V2 | none; calculated |
| rules-range-exclusive-invalid | V2 | requested too-small |
| rules-no-negatives-list-item-valid | V2 | none; calculated |
| rules-no-negatives-list-item-invalid | V2 | contributions item 2 no-negatives |
| rules-list-at-least-one-not-given | V2 | contributions (not given) list-too-short |
| rules-list-at-least-one-valid | V17 | none; calculated |
| rules-list-at-least-one-invalid | V17 | yields list-too-short |
| rules-decimal-places-list-item-invalid | V17 | yields item 1 decimal-places |
| rules-range-valid | V3 | none; calculated |
| rules-range-edge-valid | V3 | none; calculated |
| rules-range-too-small | V3 | members too-small |
| rules-range-too-large | V3 | members too-large |
| rules-list-not-given-is-empty | V3 | none; calculated |
| rules-each-item-valid | V10 | none; calculated |
| rules-each-item-too-large | V10 | scores item 2 too-large |
| rules-each-item-three-problems | V10 | scores item 1 too-small; scores item 2 too-large; scores item 3 wrong-type |
| rules-list-at-most-valid | V10 | none; calculated |
| rules-list-too-long | V10 | scores list-too-long |
| rules-type-list-item-valid | V16 | none; calculated |
| rules-type-list-item-invalid | V16 | balances item 2 wrong-type |
| rules-check-valid | V9 | none; calculated |
| rules-check-two-failed | V9 | check 1 check-failed; check 2 check-failed |
| rules-check-skipped-input-problem | V9 | monthly_rate too-large; check 1 check-failed; check 2 skipped |
| rules-check-skipped-all | V9 | principal too-small; checks 1 and 2 skipped |
| rules-check-error | VB | check 1 check-error |
| rules-check-division-valid | VB | none; calculated |
| rules-definition-range-reversed | V19 with bag_size range min 100 and max 1 | range-reversed |
| rules-definition-places-too-many | V1 with payment decimal-places 11 | places-too-many |
| rules-definition-allowed-values-empty | V2 with status allowed-values with no value | allowed-values-empty |
| rules-definition-check-unknown-name | V1 with check 1 `payment <= limit` | unknown-name |
| rules-definition-check-not-boolean | V1 with check 1 `payment + 1` | check-not-boolean |
| rules-definition-input-needs-required | V19 with bag_size range 1 to 100 and no required | input-needs-required |

## Rows that need a value type the loader does not have

Proposed; the owner decides the syntax first (example-cases.md, open question 3). The loader's table has `int`, `dec`, `text`, `bool` and `date` only.

| Type | Proposed syntax | Rows |
|---|---|---|
| float | `float:0.37`, an approximate 64-bit number written as text and read as a float | `lang-f13-collateral-coverage-ratio`, `lang-f14-interim-payment`, `lang-problem-float-decimal-mix` |
| list of integer or decimal | `list<int>:[72, 85]`, `list<dec>:[200.00, 250.00]` and the empty list `list<dec>:[]` | `lang-f02-loan-limit-check`, `lang-f03-fund-share-per-member`, `lang-f10-term-grade`, `lang-f15-capital-call-share`, `lang-f16-lending-data-check`, `lang-f17-plot-yield-spread`, `lang-problem-too-costly`, `lang-problem-empty-list` |
| list of text, for a Validate row | `list<text>:["72", "104"]` and the empty list `list<text>:[]`; each item is a received text | the 22 `rules-` rows with a list input |

The page has no notation of its own for a float value or a list value: it writes the float 0.37 in words, and a list as `[200.00, 250.00]`.
A list value contains `, ` between its items, and a row's inputs are separated by `, ` as well. So the loader would need to split inputs outside brackets as well as outside quotes.
Rows that expect a float result: `lang-f13-collateral-coverage-ratio` (`result: float:1.363`). No row expects a list result.

## Rows longer than the loader allows

The loader's formula cell holds 1 to 200 characters. These formulas are longer; none is shortened here. Either the limit or the example needs to change.

| Row | Formula length, in characters |
|---|---|
| `lang-problem-too-many-nodes` | 302 |
| `lang-problem-too-long` | 213 |

The design page asks that the cell limit be at least the formula text limit of 500 (Differences to settle, item 4). Lengths are of the formula without the `\|` escapes.

## Check against the design page

| Item | Value |
|---|---|
| Formula blocks on the page | 111: 55 `lang-` and 56 `rules-` |
| Rows in the table above | 111: 55 `lang-` and 56 `rules-` |
| Comparison | the names after `case: ` on the page against the first cell of each row of the case table; the command is below. The `rules-` examples come first on the page, so their rows come first in the table. |
| Result | diff gave no output: the two lists of 111 names are identical, in the same order |

The rows of the case table are the ones that end with the page cell. The two tables of "Definitions of the Validate rows" also start with a `rules-` name, so the command leaves them out by that last cell.

```
diff <(grep '^case: ' "documentation/2. design/formula-language.md" | sed 's/^case: //') <(grep '^| \(lang\|rules\)-.* | documentation/2\. design/formula-language\.md |$' "documentation/4. qa/formula-examples.md" | sed 's/^| \([^ ]*\) |.*/\1/')
```
