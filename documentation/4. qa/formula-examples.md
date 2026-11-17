# Formula examples

This file holds one case row for every example on `documentation/2. design/formula-language.md`.
The shared case loader, `internal/casefile`, described in `documentation/4. qa/example-cases.md`, reads the table below.
The rows here and the examples on the design page are the same list, matched by case name.
In a row that expects problems, only the type of each input matters; the values are placeholders.
Status: the owner has not reviewed this page. All names and figures are invented.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
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

## Rows that need a value type the loader does not have

Proposed; the owner decides the syntax first (example-cases.md, open question 3). The loader's table has `int`, `dec`, `text`, `bool` and `date` only.

| Type | Proposed syntax | Rows |
|---|---|---|
| float | `float:0.37`, an approximate 64-bit number written as text and read as a float | `lang-f13-collateral-coverage-ratio`, `lang-f14-interim-payment`, `lang-problem-float-decimal-mix` |
| list of integer or decimal | `list<int>:[72, 85]`, `list<dec>:[200.00, 250.00]` and the empty list `list<dec>:[]` | `lang-f02-loan-limit-check`, `lang-f03-fund-share-per-member`, `lang-f10-term-grade`, `lang-f15-capital-call-share`, `lang-f16-lending-data-check`, `lang-f17-plot-yield-spread`, `lang-problem-too-costly`, `lang-problem-empty-list` |

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
| Formula blocks on the page | 55 |
| Rows in the table above | 55 |
| Comparison | the names after `case: ` on the page against the first cell of each row; the command is below |
| Result | diff gave no output: the two lists of 55 names are identical, in the same order |

```
diff <(grep '^case: ' "documentation/2. design/formula-language.md" | sed 's/^case: //') <(grep '^| lang-' "documentation/4. qa/formula-examples.md" | sed 's/^| \([^ ]*\) |.*/\1/')
```
