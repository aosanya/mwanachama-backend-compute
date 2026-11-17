# Example formulas

Agency functions carry a body of one line of spreadsheet-like text, such as `payment >= balance ? 0.00 : balance - payment`.
This page lists nineteen invented formulas, each a calculation an agency template asks for, with typed inputs, rules, a worked example and an invalid example.
The language page (`formula-language.md`) is written from this page, so every language part it describes must be needed by a formula below.

How the formulas were found: the agency templates were searched for the words calculate, interest, reduce, deduct, accrue, prorate, depreciat, balance, ratio, total and average, then for pro-rata, installment, penalty, GPA, yield, percent, repayment and arrears. The needs that came up were rewritten as invented formulas with invented names and figures.

Conventions used below:

- Types are decimal (exact, `12.50`), float (approximate, `0.75f`), integer, text, boolean, date and list of any of these. Dates are written `2026-03-31`.
- `%` is the remainder of integer division. Rates are written as decimals, for example `0.015`.
- `round(x, n)` rounds half away from zero to `n` places and keeps the type of `x`.
- `sum`, `min` and `max` of a list give the element type. `count` gives an integer. `avg` of decimals gives a decimal, `avg` of integers or floats gives a float.
- `days_between(a, b)`, `months_between(a, b)` and `years_between(a, b)` count whole days, months and years from `a` to `b`. `to_float(x)` turns a decimal into a float. `to_decimal(x, n)` turns a float into a decimal with `n` places.
- A function is named `name@version`. Another function calls it by that name.

## 1. Balance after payment

Need: a member pays towards what they owe, and the balance never goes below zero.

| Input | Type | Rule |
| --- | --- | --- |
| balance | decimal | required, no negatives, 2 decimal places |
| payment | decimal | required, no negatives, 2 decimal places |

Result: decimal

Formula: `payment >= balance ? 0.00 : balance - payment`

Worked example: balance 120.50 and payment 45.25 give 75.25. Payment 130.00 gives 0.00.

Invalid example: payment -5.00 breaks the no negatives rule.

## 2. Loan limit check

Need: a loan may not exceed three times the member's contributions, and only an active member can borrow.

| Input | Type | Rule |
| --- | --- | --- |
| status | text | required, allowed values "active", "lapsed", "suspended" |
| requested | decimal | required, more than 0.00, 2 decimal places |
| contributions | list of decimal | at least one, no negatives |

Result: boolean

Formula: `status == "active" && requested <= 3 * sum(contributions)`

Worked example: status "active", requested 1400.00, contributions [200.00, 250.00, 150.00, 100.00]. The sum is 700.00, three times is 2100.00, and 1400.00 is not more than that, so the result is true.

Invalid example: status "paused" breaks the allowed values rule.

## 3. Fund share per member

Need: what is left in a savings fund after expenses, shared equally between members.

| Input | Type | Rule |
| --- | --- | --- |
| deposits | list of decimal | at least one, no negatives |
| expenses | list of decimal | no negatives, may be empty |
| members | integer | required, range 1 to 500 |

Result: decimal

Formula: `round((sum(deposits) - sum(expenses)) / members, 2)`

Worked example: deposits [500.00, 750.00, 1250.50] sum to 2500.50. Expenses [120.25, 80.00] sum to 200.25. The difference is 2300.25. With 7 members, 2300.25 / 7 = 328.607142 and the result is 328.61.

Invalid example: members 0 breaks the range rule (1 to 500).

## 4. Farmer net payout

Need: pay a farmer for delivered kilos and take off the input credit, never paying less than zero.

| Input | Type | Rule |
| --- | --- | --- |
| kilos | integer | required, no negatives |
| price_per_kilo | decimal | required, more than 0.000, at most 3 decimal places |
| credit | decimal | required, no negatives, 2 decimal places |

Result: decimal

Formula: `max([0.00, round(kilos * price_per_kilo, 2) - credit])`

Worked example: kilos 853 and price 0.437 give 372.761, rounded to 372.76. With credit 120.00 the result is 252.76. With credit 400.00 the result is 0.00.

Invalid example: price_per_kilo 0.4375 has four decimal places and breaks the decimal places rule.

## 5. Late days

Need: the number of days a filing came after its due date, zero when it was on time.

| Input | Type | Rule |
| --- | --- | --- |
| due | date | required |
| filed | date | required |

Result: integer. Published as `late_days@1`.

Formula: `max([0, days_between(due, filed)])`

Worked example: due 2026-03-31 and filed 2026-05-15 are 30 days to 2026-04-30 plus 15 more, so the result is 45. Filed 2026-03-20 gives 0.

Invalid example: filed not given breaks the required rule.

## 6. Late-filing penalty

Need: a penalty on the amount due that steps up the later the filing is.

| Input | Type | Rule |
| --- | --- | --- |
| amount | decimal | required, no negatives, 2 decimal places |
| due | date | required |
| filed | date | required |

Result: decimal

Formula: `late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)`

Worked example: amount 8400.00 filed 45 days late (formula 5). 45 is not 0, not over 90, over 30, so 8400.00 * 0.05 = 420.00. Filed 12 days late gives 8400.00 * 0.02 = 168.00.

Invalid example: amount 8400.005 has three decimal places and breaks the decimal places rule.

## 7. Interest rate step

Need: the monthly interest rate depends on the band the balance falls in.

| Input | Type | Rule |
| --- | --- | --- |
| balance | decimal | required, no negatives, 2 decimal places |

Result: decimal. Published as `interest_step@1`.

Formula: `balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020`

Worked example: balance 3217.45 is not under 1000.00 but is under 5000.00, so the result is 0.015. Balance 640.00 gives 0.010 and 5000.00 gives 0.020.

Invalid example: balance -1.00 breaks the no negatives rule.

## 8. Monthly interest

Need: the interest charged for a month, using the stepped rate.

| Input | Type | Rule |
| --- | --- | --- |
| balance | decimal | required, no negatives, 2 decimal places |

Result: decimal

Formula: `round(balance * interest_step@1(balance), 2)`

Worked example: balance 3217.45 with rate 0.015 gives 48.26175, rounded to 48.26.

Invalid example: balance 3217.455 has three decimal places and breaks the decimal places rule.

## 9. Flat-rate loan instalment

Need: the monthly instalment of a loan whose interest is a flat rate on the amount borrowed.

| Input | Type | Rule |
| --- | --- | --- |
| principal | decimal | required, more than 0.00, 2 decimal places |
| monthly_rate | decimal | required, range 0.000 to 0.050, at most 3 decimal places |
| months | integer | required, range 1 to 60 |

Result: decimal

Formula: `round((principal + principal * monthly_rate * months) / months, 2)`

Worked example: principal 2400.00, rate 0.012, 7 months. Interest is 2400.00 * 0.012 * 7 = 201.60, the total is 2601.60, and 2601.60 / 7 = 371.657142 rounds to 371.66.

Invalid example: months 72 breaks the range rule (1 to 60).

## 10. Term grade

Need: a student's letter grade from the average of the term's scores.

| Input | Type | Rule |
| --- | --- | --- |
| scores | list of integer | at least one, each in range 0 to 100 |

Result: text, one of "A", "B", "C", "F"

Formula: `avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"`

Worked example: scores [72, 85, 64, 91, 77] sum to 389, the average is 77.8, which is not 80 or more but is 65 or more, so the result is "B".

Invalid example: scores [72, 104] has an element above 100 and breaks the range rule.

## 11. Pro-rated first-month fee

Need: a member who joins part-way through a month pays only for the days left, counting the joining day.

| Input | Type | Rule |
| --- | --- | --- |
| monthly_fee | decimal | required, no negatives, 2 decimal places |
| joined | date | required |
| period_end | date | required |
| days_in_period | integer | required, range 28 to 31 |

Result: decimal

Formula: `round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)`

Worked example: fee 47.00, joined 2026-06-12, period end 2026-06-30, 30 days in the period. The days between are 18, plus 1 is 19. 47.00 * 19 = 893.00, and 893.00 / 30 = 29.7666 rounds to 29.77.

Invalid example: days_in_period 32 breaks the range rule (28 to 31).

## 12. Equipment value after straight-line depreciation

Need: the book value of equipment after the full years since purchase, never below its salvage value.

| Input | Type | Rule |
| --- | --- | --- |
| cost | decimal | required, more than 0.00, 2 decimal places |
| salvage | decimal | required, no negatives, 2 decimal places |
| life_years | integer | required, range 1 to 50 |
| bought | date | required |
| on | date | required |

Result: decimal

Formula: `max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])`

Worked example: cost 12000.00, salvage 2000.00, life 7 years, bought 2023-09-10, on 2026-11-17. Full years are 3. Yearly depreciation is 10000.00 / 7 = 1428.5714, rounded to 1428.57. Three years is 4285.71. 12000.00 - 4285.71 = 7714.29, which is above 2000.00, so the result is 7714.29.

Invalid example: life_years 0 breaks the range rule (1 to 50).

## 13. Collateral coverage ratio

Need: how many times the value of pledged stock covers the facility balance, as an approximate ratio.

| Input | Type | Rule |
| --- | --- | --- |
| stock_value | decimal | required, no negatives, 2 decimal places |
| facility_balance | decimal | required, more than 0.00, 2 decimal places |

Result: float

Formula: `round(to_float(stock_value) / to_float(facility_balance), 3)`

Worked example: stock 84500.00 and balance 62000.00 give 1.362903, rounded to 1.363f.

Invalid example: facility_balance 0.00 breaks the more than 0.00 range rule, which stops a division by zero.

## 14. Interim payment on percent complete

Need: pay for the certified share of a contract, hold back ten percent, and take off what was already paid.

| Input | Type | Rule |
| --- | --- | --- |
| contract | decimal | required, more than 0.00, 2 decimal places |
| percent_complete | float | required, range 0.0f to 1.0f |
| paid | decimal | required, no negatives, 2 decimal places |

Result: decimal

Formula: `max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])`

Worked example: contract 183500.00, percent complete 0.37f, paid 52000.00. The float becomes the decimal 0.37. 183500.00 * 0.37 = 67895.00, times 0.90 is 61105.50, minus 52000.00 is 9105.50.

Invalid example: percent_complete 1.2f breaks the range rule (0.0f to 1.0f).

## 15. Capital call share

Need: an investor's share of a call, in proportion to the units they hold out of all units.

| Input | Type | Rule |
| --- | --- | --- |
| call_amount | decimal | required, more than 0.00, 2 decimal places |
| units | integer | required, no negatives |
| all_units | list of integer | at least one, no negatives |

Result: decimal

Formula: `round(call_amount * units / sum(all_units), 2)`

Worked example: call 50000.00, units 1200, all units [1200, 830, 1500, 477] sum to 4007. 50000.00 * 1200 = 60000000.00, and 60000000.00 / 4007 = 14973.7959, rounded to 14973.80.

Invalid example: units -200 breaks the no negatives rule.

## 16. Lending data check

Need: weekly balances from a statement are enough and good enough to consider an application.

| Input | Type | Rule |
| --- | --- | --- |
| balances | list of decimal | at least one, 2 decimal places each |
| floor | decimal | required, no negatives, 2 decimal places |

Result: boolean

Formula: `count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00`

Worked example: balances [420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00] number 8 and sum to 3001.00. The average is 375.125, which is 350.00 or more with floor 350.00, and the lowest is 95.25, so the result is true.

Invalid example: balances [420.00, "n/a"] has a text element and breaks the type rule (list of decimal).

## 17. Plot yield spread

Need: the gap between a farmer's best and worst plot.

| Input | Type | Rule |
| --- | --- | --- |
| yields | list of decimal | at least one, no negatives, 2 decimal places |

Result: decimal

Formula: `max(yields) - min(yields)`

Worked example: yields [2400.50, 3100.25, 2750.00] give 3100.25 - 2400.50 = 699.75.

Invalid example: yields [] breaks the at least one rule (required list).

## 18. Voting right

Need: a member may vote when not suspended, not a guest, and either an officer or paid up with six or more months of membership.

| Input | Type | Rule |
| --- | --- | --- |
| joined | date | required |
| on | date | required |
| paid_up | boolean | required |
| suspended | boolean | required |
| role | text | required, allowed values "member", "officer", "guest" |

Result: boolean

Formula: `!suspended && role != "guest" && (role == "officer" || (paid_up && months_between(joined, on) >= 6))`

Worked example: joined 2026-01-20, on 2026-09-15, paid up, not suspended, role "member". Full months are 7, so the result is true. The same member with suspended true gives false.

Invalid example: paid_up not given breaks the required rule.

## 19. Seed bags left over

Need: the kilos left over after filling whole bags.

| Input | Type | Rule |
| --- | --- | --- |
| kilos | integer | required, no negatives |
| bag_size | integer | required, range 1 to 100 |

Result: integer

Formula: `kilos % bag_size`

Worked example: kilos 1237 and bag size 50 fill 24 bags (1200 kilos) and leave 37.

Invalid example: kilos 1237.5 is not a whole number and breaks the type rule (integer).

## Parts of the language the examples need

| Part | Formulas |
| --- | --- |
| Integer, decimal, float, text, boolean and date values | 1 to 19 |
| Float literal such as `0.75f` | 13, 14 |
| List values, written `[1, 2, 3]` | 4, 5, 12, 14 (in formulas); 2, 3, 10, 15, 16, 17 (as inputs) |
| `+` | 9, 11 |
| `-` | 1, 3, 4, 12, 14, 17 |
| `*` | 2, 4, 6, 8, 9, 11, 12, 14, 15 |
| `/` | 3, 9, 11, 12, 13, 15 |
| `%` (remainder) | 19 |
| Comparison `>=` | 1, 10, 11, 16, 18 |
| Comparison `<=` | 2 |
| Comparison `>` | 6 |
| Comparison `<` | 7 |
| Comparison `==` | 2, 6, 18 |
| Comparison `!=` | 18 |
| Boolean `&&` | 2, 16, 18 |
| Boolean `\|\|` | 18 |
| Boolean `!` | 18 |
| Brackets to group | 3, 9, 11, 12, 14, 18 |
| Conditional `cond ? a : b` | 1, 6, 7, 10 |
| Nested conditional | 6, 7, 10 |
| `round(x, n)` on decimal | 3, 4, 6, 8, 9, 11, 12, 14, 15 |
| `round(x, n)` on float | 13 |
| `sum(x)` | 2, 3, 15 |
| `count(x)` | 16 |
| `avg(x)` | 10, 16 |
| `min(x)` | 16, 17 |
| `max(x)` | 4, 5, 12, 14, 17 |
| `to_float(x)` | 13 |
| `to_decimal(x, n)` | 14 |
| `days_between(a, b)` | 5, 11 |
| `months_between(a, b)` | 18 |
| `years_between(a, b)` | 12 |
| Named library call with a version, `name@1(...)` | 6, 8 |
| One function calling another | 6 (calls 5), 8 (calls 7) |
| The same call repeated in one formula | 6 |
| Input rule: required | 1 to 19 |
| Input rule: type | 16, 19 |
| Input rule: range | 3, 4, 9, 10, 11, 12, 13, 14, 19 |
| Input rule: no negatives | 1, 2, 3, 4, 6, 7, 8, 11, 12, 13, 14, 15, 16, 17, 19 |
| Input rule: decimal places | 1 to 4, 6 to 9, 11 to 17 |
| Input rule: allowed values | 2, 18 |
| Input rule: list has at least one element | 2, 10, 15, 16, 17 |
| Rule applied to each list element | 2, 3, 10, 15, 17 |
