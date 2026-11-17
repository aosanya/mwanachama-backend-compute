# Library reference cases

Status: draft. Reference cases for the first ten library functions and for the nine that follow them (152 cases). All names and figures are invented. Nothing here chooses a rule the owner has not decided; those are listed under Open questions.

## What this page protects

A library function must give the same answer for the same input for as long as its version exists. Each function below has recorded cases: inputs, the expected result, the source of the figure and the working. A case with no source is not accepted. A published version is never edited (see [library-versioning.md](../2.%20design/library-versioning.md)), so its cases are never changed; a fix is a new version and its cases are added beside the old ones. The test that fails when a recorded case of a published version changes is described on [library-reference-guard.md](../3.%20implementation/library-reference-guard.md). Every source named here is recorded on [sources.md](../1.%20requirements/sources.md) under the same id.

The functions are the nineteen formulas of [example-formulas.md](../1.%20requirements/example-formulas.md) turned into library functions. The ten under The first ten functions are the ones the agency templates need first. The other nine, under The other nine functions, follow the same shape.

## The case shape

The inputs and the expected result are written as text in the notation of example-formulas.md: decimals as `12.50`, floats as `0.75f`, dates as `2026-03-31`, lists as `[1.00, 2.00]`.

| Field | What it holds |
|---|---|
| Id | `REF-<function>-<nn>`, for example `REF-fin.balance_after_payment-01`. The number counts up across the versions of one function: the cases of version 2 take the next numbers and the cases of version 1 keep theirs. |
| Function with version | The heading of the section the case sits in, for example `fin.balance_after_payment@1`. |
| Kind | `normal` (an ordinary input), `edge` (zero, one value, a boundary, or a value that sits on a rounding half) or `refusal` (an input the function must refuse). |
| Inputs | Each input by name, with its value as text. |
| Expected | The result as text, or `refused`. |
| Tolerance | Set once per function, above its case table. |
| Source | A source id (`S-01`, `S-02`, `S-03`), with the section number when the source is a page section. |
| Working | The step-by-step working that gives the expected figure. |

A `refused` case passes when the function gives no result and says so in one sentence. The wording is not fixed (Open question 2).

## Tolerances

| Result type | Tolerance |
|---|---|
| Decimal (every money function) | Exact: difference 0. The result is a decimal with a scale of 2 places unless a function says otherwise, so `0.00` is right and `0` is wrong. |
| Integer | Exact. |
| Boolean and text | Exact: the result is the same text, `true` or `B`. |
| Float (`fin.coverage_ratio@1` only) | A relative error of 1e-9: the result passes when `|result - expected| <= 1e-9 x |expected|`. |
| Refused | No result. |

Provisional: the float tolerance of 1e-9 is a figure of this page, not of the language page. The ratio is rounded to 3 places, so any change in the third place (a relative change of 1e-3 or more) fails, and float noise of about 1e-16 passes. When the expected result is `0.0f` the page requires exactly `0.0f` (Open question 3).

Rounding. Every case that rounds uses the mode example-formulas.md states for `round(x, n)`: half away from zero (named half-up on [exact-money.md](exact-money.md)). Provisional: the mode depends on the formula language page and on the owner's rounding decision (exact-money.md open question 1). A case whose figure sits exactly on a half is marked mode-dependent in its working. These are all of them:

| Case | Half-up (used) | Half-even | Half-down |
|---|---|---|---|
| REF-fin.equal_share-03 | 0.03 | 0.02 | 0.02 |
| REF-fin.flat_instalment-06 | 0.13 | 0.12 | 0.12 |
| REF-fin.prorate_fee-06 | 0.01 | 0.00 | 0.00 |
| REF-fin.straight_line_value-06 | 0.87 | 0.88 | 0.88 |
| REF-fin.proportional_share-05 | 0.13 | 0.12 | 0.12 |
| REF-fin.farmer_net_payout-05 | 0.01 | 0.00 | 0.00 |
| REF-fin.monthly_interest-06 | 0.01 | 0.00 | 0.00 |
| REF-fin.interim_payment-06 | 342.00 | 342.00 | 333.00 |

Cases that round without sitting on a half agree under all three half modes. They would differ under up, down, ceiling and floor, which the owner has not ruled out.

## Edge-case checklist

Every function has cases for the kinds of input it takes. A refusal never reaches the calculation, so an undefined answer is a refusal and not an error halfway through.

| Kind of input | Value | Expected behaviour |
|---|---|---|
| Decimal money | 0.00 | Accepted where the rule says "no negatives". Refused where the rule says "more than 0.00" (principal, cost, facility_balance, call_amount). |
| Decimal money | 0.01 | Accepted. The cent is kept. |
| Decimal money | negative | Refused. The formulas page says negatives are refused by the input rules. |
| Decimal money | 3 places, such as 10.005 | Refused (2 decimal places rule). A rate takes 3 places and a fourth is refused. |
| Decimal rate | outside its range, including -1.000 and 0.051 | Refused (range rule). A rate of minus one never reaches the division. |
| Integer | 0 | Accepted where the range starts at 0 or no range is set (units, value). Refused where the range starts at 1 (members, months, life_years, divisor). |
| Integer | each end of its range, and one outside | The ends are accepted, the value outside is refused. |
| Integer | negative, or not whole (1237.5) | Refused. |
| List | empty | Refused where the rule says "at least one". Accepted where the list may be empty (expenses), and it sums to 0.00. |
| List | one value | Accepted. The sum is that value, and the spread is 0.00. |
| List | all values equal | Accepted. The spread is 0.00. |
| List | one negative element | Refused as a whole. |
| List used as a divisor | sums to 0 | Refused: a division by zero has no answer. |
| Date | same day twice | Accepted. Counts 0 days. |
| Date | second date before the first | Where the formula states a result, that result (late days is 0). Where it does not, see Open question 6. |
| Date | leap day, month end, year end | Counted by the calendar. |
| Date | not a calendar date (2026-02-30), or missing | Refused. |
| Any required input | not given | Refused. |
| Text with allowed values | a value outside the list (`paused`, `visitor`) | Refused. Each allowed value is accepted. |
| Yes or no | not given | Refused. `true` and `false` are accepted. |
| Float | each end of its range, and one outside (0.0f, 1.0f, 1.2f) | The ends are accepted, the value outside is refused. |
| List of decimal without a no negatives rule | one negative element | Accepted as input; the function's own test then decides (`fin.lending_data_ok@1` gives false). |
| A band or a step | each side of every boundary | The boundary value belongs to the band the formula states (a balance of 1000.00 is in the middle band, 90 days late is in the 5 percent step). |
| Answer undefined for the input | variance of one value, rate of minus one, division by zero | Never silently 0 and never a guess: either a defined result written down, or a refusal. `stats.spread@1` of one value is defined and gives 0.00. A sample variance of one value is undefined and gets its refusal case when that function is written. |

## The first ten functions

Numbers in brackets are the formula numbers of example-formulas.md. The input names follow the formulas page, except where a name there belongs to one use (yields, kilos, bag_size): `stats.spread@1` takes `values` and `math.remainder@1` takes `value` and `divisor`. The ranges stay as the formulas page states them.

No function was swapped for another; the ten are the suggested set.

### fin.balance_after_payment@1

Signature: `fin.balance_after_payment@1(balance: decimal, payment: decimal) -> decimal`. Formula 1.

Rule: the balance after a payment, and never below 0.00. Both inputs take no negatives and 2 places.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.balance_after_payment-01 | normal | `balance 120.50, payment 45.25` | `75.25` | S-01 §1 | 45.25 is below 120.50, so 120.50 - 45.25 = 75.25. |
| REF-fin.balance_after_payment-02 | normal | `balance 120.50, payment 130.00` | `0.00` | S-01 §1 | 130.00 is not below 120.50, so the result is 0.00. |
| REF-fin.balance_after_payment-03 | edge | `balance 80.00, payment 0.00` | `80.00` | S-03 | 0.00 is below 80.00, so 80.00 - 0.00 = 80.00. |
| REF-fin.balance_after_payment-04 | edge | `balance 75.25, payment 75.25` | `0.00` | S-03 | The payment equals the balance, which counts as paid, so 0.00. |
| REF-fin.balance_after_payment-05 | edge | `balance 100.00, payment 99.99` | `0.01` | S-03 | In cents 10,000 - 9,999 = 1, which is 0.01. |
| REF-fin.balance_after_payment-06 | edge | `balance 0.00, payment 0.00` | `0.00` | S-03 | 0.00 is not below 0.00, so the result is 0.00. |
| REF-fin.balance_after_payment-07 | refusal | `balance 120.50, payment -5.00` | `refused` | S-02 §1 | -5.00 breaks the no negatives rule of payment. |
| REF-fin.balance_after_payment-08 | refusal | `balance 120.50, payment 10.005` | `refused` | S-02 §1 | 10.005 has 3 places and breaks the 2 decimal places rule. |

### fin.equal_share@1

Signature: `fin.equal_share@1(deposits: list of decimal, expenses: list of decimal, members: integer) -> decimal`. Formula 3.

Rule: what is left after expenses, shared equally and rounded to 2 places. Deposits have at least one value, expenses may be empty, members is 1 to 500, no negatives.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.equal_share-01 | normal | `deposits [500.00, 750.00, 1250.50], expenses [120.25, 80.00], members 7` | `328.61` | S-01 §3 | Deposits 500.00 + 750.00 + 1250.50 = 2500.50. Expenses 120.25 + 80.00 = 200.25. 2500.50 - 200.25 = 2300.25. 2300.25 / 7 = 328.607142..., and 0.007142 is above the half of 0.01, so 328.61. |
| REF-fin.equal_share-02 | normal | `deposits [100.00], expenses [], members 3` | `33.33` | S-03 | Deposits 100.00, no expenses (0.00). 100.00 / 3 = 33.3333..., and 0.0033 is below the half, so 33.33. |
| REF-fin.equal_share-03 | edge | `deposits [0.05], expenses [], members 2` | `0.03` | S-03 | 0.05 / 2 = 0.025, exactly on the half of 0.01. Away from zero gives 0.03. Mode-dependent: half-even gives 0.02. |
| REF-fin.equal_share-04 | edge | `deposits [60.00, 40.25], expenses [10.25], members 1` | `90.00` | S-03 | 60.00 + 40.25 = 100.25. 100.25 - 10.25 = 90.00. 90.00 / 1 = 90.00. One member holds everything. |
| REF-fin.equal_share-05 | edge | `deposits [500.00], expenses [500.00], members 500` | `0.00` | S-03 | 500.00 - 500.00 = 0.00. 0.00 / 500 = 0.00. Members at the top of its range, and a zero result. |
| REF-fin.equal_share-06 | refusal | `deposits [500.00], expenses [], members 0` | `refused` | S-02 §3 | members 0 breaks the range rule (1 to 500), which stops a division by zero. |
| REF-fin.equal_share-07 | refusal | `deposits [], expenses [], members 4` | `refused` | S-02 §3 | An empty deposits list breaks the at least one rule. |
| REF-fin.equal_share-08 | refusal | `deposits [500.00], expenses [-5.00], members 4` | `refused` | S-02 §3 | -5.00 in expenses breaks the no negatives rule. |

### date.late_days@1

Signature: `date.late_days@1(due: date, filed: date) -> integer`. Formula 5.

Rule: whole days from `due` to `filed`, and 0 when filed on or before due.

Provisional: example-formulas.md publishes this as `late_days@1` with no library. The prefix `date.` is proposed here so the name fits `<library>.<function>@<n>`; the owner has not decided it (Open question 8).

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-date.late_days-01 | normal | `due 2026-03-31, filed 2026-05-15` | `45` | S-01 §5 | 2026-03-31 to 2026-04-30 is 30 days. To 2026-05-15 is 15 more. 30 + 15 = 45. |
| REF-date.late_days-02 | normal | `due 2026-03-31, filed 2026-03-20` | `0` | S-01 §5 | Filed is 11 days before due, so the day count is -11 and max(0, -11) = 0. |
| REF-date.late_days-03 | edge | `due 2026-03-31, filed 2026-03-31` | `0` | S-03 | The same day is 0 days apart. |
| REF-date.late_days-04 | edge | `due 2026-03-31, filed 2026-04-01` | `1` | S-03 | One day after due is 1 day late. |
| REF-date.late_days-05 | edge | `due 2028-02-28, filed 2028-03-01` | `2` | S-03 | 2028 is a leap year. 28 Feb to 29 Feb is 1 day and 29 Feb to 1 Mar is 1 more: 2. |
| REF-date.late_days-06 | edge | `due 2026-12-31, filed 2027-03-01` | `60` | S-03 | 2026-12-31 to 2027-01-31 is 31 days. To 2027-02-28 is 28 more, 59. To 2027-03-01 is 1 more: 60. The span crosses a year end. |
| REF-date.late_days-07 | refusal | `due 2026-03-31, filed not given` | `refused` | S-02 §5 | filed is required and is not given. |
| REF-date.late_days-08 | refusal | `due 2026-02-30, filed 2026-03-05` | `refused` | S-02 §5 | February 2026 has 28 days, so 2026-02-30 is not a date and breaks the date type. |

### fin.flat_instalment@1

Signature: `fin.flat_instalment@1(principal: decimal, monthly_rate: decimal, months: integer) -> decimal`. Formula 9.

Rule: the monthly instalment of a loan whose interest is a flat monthly rate on the amount borrowed, rounded to 2 places. Principal is more than 0.00, rate is 0.000 to 0.050 with at most 3 places, months is 1 to 60.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.flat_instalment-01 | normal | `principal 2400.00, monthly_rate 0.012, months 7` | `371.66` | S-01 §9 | Interest 2400.00 x 0.012 x 7 = 201.60. Total 2400.00 + 201.60 = 2601.60. 2601.60 / 7 = 371.657142..., so 371.66. |
| REF-fin.flat_instalment-02 | normal | `principal 1000.00, monthly_rate 0.010, months 10` | `110.00` | S-03 | Interest 1000.00 x 0.010 x 10 = 100.00. Total 1100.00. 1100.00 / 10 = 110.00. |
| REF-fin.flat_instalment-03 | edge | `principal 600.00, monthly_rate 0.000, months 12` | `50.00` | S-03 | Rate 0.000 adds no interest. 600.00 / 12 = 50.00. |
| REF-fin.flat_instalment-04 | edge | `principal 500.00, monthly_rate 0.025, months 1` | `512.50` | S-03 | Interest 500.00 x 0.025 x 1 = 12.50. Total 512.50. 512.50 / 1 = 512.50. One month, so the instalment is the whole total. |
| REF-fin.flat_instalment-05 | edge | `principal 10000.00, monthly_rate 0.050, months 60` | `666.67` | S-03 | Top of the rate and months ranges. Interest 10000.00 x 0.050 x 60 = 30000.00. Total 40000.00. 40000.00 / 60 = 666.6666..., so 666.67. |
| REF-fin.flat_instalment-06 | edge | `principal 1.00, monthly_rate 0.000, months 8` | `0.13` | S-03 | No interest. 1.00 / 8 = 0.125, exactly on the half of 0.01. Away from zero gives 0.13. Mode-dependent: half-even gives 0.12. |
| REF-fin.flat_instalment-07 | refusal | `principal 2400.00, monthly_rate 0.012, months 72` | `refused` | S-02 §9 | months 72 breaks the range rule (1 to 60). |
| REF-fin.flat_instalment-08 | refusal | `principal 2400.00, monthly_rate -1.000, months 7` | `refused` | S-02 §9 | A rate of minus one is outside the range 0.000 to 0.050. |

### fin.prorate_fee@1

Signature: `fin.prorate_fee@1(monthly_fee: decimal, joined: date, period_end: date, days_in_period: integer) -> decimal`. Formula 11.

Rule: the fee for the days from the joining day to the period end, counting the joining day, as a share of `days_in_period`, rounded to 2 places. Fee has no negatives, days_in_period is 28 to 31.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.prorate_fee-01 | normal | `monthly_fee 47.00, joined 2026-06-12, period_end 2026-06-30, days_in_period 30` | `29.77` | S-01 §11 | 18 days between, plus the joining day is 19. 47.00 x 19 = 893.00. 893.00 / 30 = 29.7666..., so 29.77. |
| REF-fin.prorate_fee-02 | normal | `monthly_fee 60.00, joined 2026-03-11, period_end 2026-03-31, days_in_period 31` | `40.65` | S-03 | 11 March to 31 March is 20 days between, plus 1 is 21. 60.00 x 21 = 1260.00. 1260.00 / 31 = 40.64516..., so 40.65. |
| REF-fin.prorate_fee-03 | edge | `monthly_fee 47.00, joined 2026-06-01, period_end 2026-06-30, days_in_period 30` | `47.00` | S-03 | Joined on the first day: 29 days between, plus 1 is 30. 47.00 x 30 / 30 = 47.00, the full fee. |
| REF-fin.prorate_fee-04 | edge | `monthly_fee 47.00, joined 2026-06-30, period_end 2026-06-30, days_in_period 30` | `1.57` | S-03 | Joined on the last day: 0 days between, plus 1 is 1. 47.00 / 30 = 1.5666..., so 1.57. The joining day is counted. |
| REF-fin.prorate_fee-05 | edge | `monthly_fee 0.00, joined 2026-06-12, period_end 2026-06-30, days_in_period 30` | `0.00` | S-03 | 0.00 x 19 = 0.00. 0.00 / 30 = 0.00. |
| REF-fin.prorate_fee-06 | edge | `monthly_fee 0.14, joined 2026-02-28, period_end 2026-02-28, days_in_period 28` | `0.01` | S-03 | 0 days between, plus 1 is 1. 0.14 x 1 / 28 = 0.005, exactly on the half of 0.01. Away from zero gives 0.01. Mode-dependent: half-even gives 0.00. |
| REF-fin.prorate_fee-07 | refusal | `monthly_fee 47.00, joined 2026-06-12, period_end 2026-06-30, days_in_period 32` | `refused` | S-02 §11 | days_in_period 32 breaks the range rule (28 to 31). |
| REF-fin.prorate_fee-08 | refusal | `monthly_fee -47.00, joined 2026-06-12, period_end 2026-06-30, days_in_period 30` | `refused` | S-02 §11 | -47.00 breaks the no negatives rule of monthly_fee. |

### fin.straight_line_value@1

Signature: `fin.straight_line_value@1(cost: decimal, salvage: decimal, life_years: integer, bought: date, on: date) -> decimal`. Formula 12.

Rule: the cost less the yearly depreciation times the whole years from `bought` to `on`, never below `salvage`. The yearly depreciation, (cost - salvage) / life_years, is rounded to 2 places before it is multiplied. Cost is more than 0.00, salvage has no negatives, life_years is 1 to 50.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.straight_line_value-01 | normal | `cost 12000.00, salvage 2000.00, life_years 7, bought 2023-09-10, on 2026-11-17` | `7714.29` | S-01 §12 | Whole years 3 (10 Sep 2023 to 10 Sep 2026, and 17 Nov is later). Yearly (12000.00 - 2000.00) / 7 = 1428.5714..., rounded 1428.57. Times 3 is 4285.71. 12000.00 - 4285.71 = 7714.29, above 2000.00. |
| REF-fin.straight_line_value-02 | normal | `cost 5000.00, salvage 500.00, life_years 3, bought 2024-01-15, on 2026-01-15` | `2000.00` | S-03 | 15 Jan 2024 to 15 Jan 2026 is exactly 2 whole years. Yearly (5000.00 - 500.00) / 3 = 1500.00. Times 2 is 3000.00. 5000.00 - 3000.00 = 2000.00, above 500.00. |
| REF-fin.straight_line_value-03 | edge | `cost 12000.00, salvage 2000.00, life_years 7, bought 2023-09-10, on 2024-09-09` | `12000.00` | S-03 | One day short of a year: 0 whole years. 1428.57 x 0 = 0.00. 12000.00 - 0.00 = 12000.00. |
| REF-fin.straight_line_value-04 | edge | `cost 12000.00, salvage 2000.00, life_years 7, bought 2010-01-01, on 2026-11-17` | `2000.00` | S-03 | 16 whole years. 1428.57 x 16 = 22857.12. 12000.00 - 22857.12 = -10857.12, below salvage, so the result is the salvage 2000.00. |
| REF-fin.straight_line_value-05 | edge | `cost 1000.00, salvage 0.00, life_years 3, bought 2023-01-01, on 2026-01-01` | `0.01` | S-03 | 3 whole years. Yearly 1000.00 / 3 = 333.3333..., rounded 333.33. Times 3 is 999.99. 1000.00 - 999.99 = 0.01. Without rounding the yearly figure first the answer would be 0.00, so this case protects the rounding before the multiplication. |
| REF-fin.straight_line_value-06 | edge | `cost 1.00, salvage 0.00, life_years 8, bought 2025-01-01, on 2026-01-01` | `0.87` | S-03 | 1 whole year. Yearly 1.00 / 8 = 0.125, exactly on the half, rounded away from zero to 0.13. 1.00 - 0.13 = 0.87. Mode-dependent: half-even rounds the yearly figure to 0.12 and gives 0.88. |
| REF-fin.straight_line_value-07 | refusal | `cost 12000.00, salvage 2000.00, life_years 0, bought 2023-09-10, on 2026-11-17` | `refused` | S-02 §12 | life_years 0 breaks the range rule (1 to 50). |
| REF-fin.straight_line_value-08 | refusal | `cost 0.00, salvage 0.00, life_years 7, bought 2023-09-10, on 2026-11-17` | `refused` | S-02 §12 | cost 0.00 breaks the more than 0.00 rule. |

### fin.coverage_ratio@1

Signature: `fin.coverage_ratio@1(stock_value: decimal, facility_balance: decimal) -> float`. Formula 13.

Rule: the stock value divided by the facility balance as floats, rounded to 3 places. Stock has no negatives, balance is more than 0.00.

Tolerance: relative error 1e-9 (see Tolerances).

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.coverage_ratio-01 | normal | `stock_value 84500.00, facility_balance 62000.00` | `1.363f` | S-01 §13 | 84500 / 62000 = 1.362903..., and to 3 places 1.363. |
| REF-fin.coverage_ratio-02 | normal | `stock_value 150.00, facility_balance 100.00` | `1.5f` | S-03 | 150 / 100 = 1.5, which is 1.500 to 3 places. |
| REF-fin.coverage_ratio-03 | edge | `stock_value 0.00, facility_balance 5000.00` | `0.0f` | S-03 | 0 / 5000 = 0. Expected zero, so the result must be exactly 0.0f. |
| REF-fin.coverage_ratio-04 | edge | `stock_value 62000.00, facility_balance 62000.00` | `1.0f` | S-03 | Equal values give exactly 1. |
| REF-fin.coverage_ratio-05 | edge | `stock_value 100.00, facility_balance 300.00` | `0.333f` | S-03 | 100 / 300 = 0.33333..., and to 3 places 0.333. |
| REF-fin.coverage_ratio-06 | edge | `stock_value 2500000.00, facility_balance 0.01` | `250000000.0f` | S-03 | The smallest balance. 2500000 / 0.01 = 250,000,000. The relative error allows 0.25 here. |
| REF-fin.coverage_ratio-07 | refusal | `stock_value 84500.00, facility_balance 0.00` | `refused` | S-02 §13 | facility_balance 0.00 breaks the more than 0.00 rule, which stops a division by zero. |
| REF-fin.coverage_ratio-08 | refusal | `stock_value -1.00, facility_balance 62000.00` | `refused` | S-02 §13 | -1.00 breaks the no negatives rule of stock_value. |

### fin.proportional_share@1

Signature: `fin.proportional_share@1(call_amount: decimal, units: integer, all_units: list of integer) -> decimal`. Formula 15.

Rule: the call amount times the units held, divided by the sum of all units, rounded to 2 places. Call is more than 0.00, units has no negatives, all_units has at least one value and no negatives.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.proportional_share-01 | normal | `call_amount 50000.00, units 1200, all_units [1200, 830, 1500, 477]` | `14973.80` | S-01 §15 | Sum of all units 1200 + 830 + 1500 + 477 = 4007. 50000.00 x 1200 = 60000000.00. 60000000.00 / 4007 = 14973.7959..., so 14973.80. |
| REF-fin.proportional_share-02 | normal | `call_amount 900.00, units 2, all_units [2, 3, 5]` | `180.00` | S-03 | Sum 2 + 3 + 5 = 10. 900.00 x 2 = 1800.00. 1800.00 / 10 = 180.00. |
| REF-fin.proportional_share-03 | edge | `call_amount 900.00, units 0, all_units [2, 3, 5]` | `0.00` | S-03 | 900.00 x 0 = 0.00. 0.00 / 10 = 0.00. An investor with no units pays nothing. |
| REF-fin.proportional_share-04 | edge | `call_amount 900.00, units 7, all_units [7]` | `900.00` | S-03 | One holder with all units. 900.00 x 7 / 7 = 900.00. |
| REF-fin.proportional_share-05 | edge | `call_amount 1.00, units 1, all_units [8]` | `0.13` | S-03 | 1.00 x 1 / 8 = 0.125, exactly on the half. Away from zero gives 0.13. Mode-dependent: half-even gives 0.12. |
| REF-fin.proportional_share-06 | edge | `call_amount 100.00, units 1, all_units [3]` | `33.33` | S-03 | 100.00 / 3 = 33.3333..., so 33.33. |
| REF-fin.proportional_share-07 | refusal | `call_amount 50000.00, units -200, all_units [1200, 830, 1500, 477]` | `refused` | S-02 §15 | units -200 breaks the no negatives rule. |
| REF-fin.proportional_share-08 | refusal | `call_amount 100.00, units 0, all_units [0]` | `refused` | S-03 | The units sum to 0, so the division is by zero and has no answer. The formulas page does not yet have a rule that stops it (Open question 7). |

### stats.spread@1

Signature: `stats.spread@1(values: list of decimal) -> decimal`. Formula 17, where the input is called `yields`.

Rule: the largest value minus the smallest. At least one value, no negatives, 2 places each.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-stats.spread-01 | normal | `values [2400.50, 3100.25, 2750.00]` | `699.75` | S-01 §17 | Largest 3100.25, smallest 2400.50. 3100.25 - 2400.50 = 699.75. |
| REF-stats.spread-02 | normal | `values [10.00, 4.50, 7.25, 4.50]` | `5.50` | S-03 | Largest 10.00. Smallest 4.50 (it appears twice). 10.00 - 4.50 = 5.50. |
| REF-stats.spread-03 | edge | `values [8.40]` | `0.00` | S-03 | One value is both largest and smallest. 8.40 - 8.40 = 0.00. |
| REF-stats.spread-04 | edge | `values [3.00, 3.00, 3.00]` | `0.00` | S-03 | All equal. 3.00 - 3.00 = 0.00. |
| REF-stats.spread-05 | edge | `values [9.99, 0.01]` | `9.98` | S-03 | Largest first and smallest last. In cents 999 - 1 = 998, which is 9.98. |
| REF-stats.spread-06 | edge | `values [0.00, 12.50]` | `12.50` | S-03 | Zero is a value like any other. 12.50 - 0.00 = 12.50. |
| REF-stats.spread-07 | refusal | `values []` | `refused` | S-02 §17 | An empty list breaks the at least one rule. |
| REF-stats.spread-08 | refusal | `values [5.00, -1.00]` | `refused` | S-02 §17 | -1.00 breaks the no negatives rule. |

### math.remainder@1

Signature: `math.remainder@1(value: integer, divisor: integer) -> integer`. Formula 19, where the inputs are called `kilos` and `bag_size`.

Rule: what is left of `value` after taking out whole `divisor`s. Value has no negatives, divisor is 1 to 100.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-math.remainder-01 | normal | `value 1237, divisor 50` | `37` | S-01 §19 | 24 x 50 = 1200 and 1237 - 1200 = 37. |
| REF-math.remainder-02 | normal | `value 100, divisor 7` | `2` | S-03 | 14 x 7 = 98 and 100 - 98 = 2. |
| REF-math.remainder-03 | edge | `value 1200, divisor 50` | `0` | S-03 | 24 x 50 = 1200 exactly, so nothing is left. |
| REF-math.remainder-04 | edge | `value 0, divisor 50` | `0` | S-03 | 0 x 50 = 0, so nothing is left. |
| REF-math.remainder-05 | edge | `value 37, divisor 50` | `37` | S-03 | 37 is below 50, so 0 whole divisors fit and all 37 are left. |
| REF-math.remainder-06 | refusal | `value 1237, divisor 0` | `refused` | S-02 §19 | divisor 0 breaks the range rule (1 to 100). A remainder by zero has no answer. |
| REF-math.remainder-07 | refusal | `value -7, divisor 50` | `refused` | S-02 §19 | -7 breaks the no negatives rule of kilos. |
| REF-math.remainder-08 | refusal | `value 1237.5, divisor 50` | `refused` | S-02 §19 | 1237.5 is not a whole number and breaks the integer type rule. |

## The other nine functions

These are the nine formulas of example-formulas.md that the first ten leave out: 2, 4, 6, 7, 8, 10, 14, 16 and 18. Each is written as a library function at version 1 with eight cases in the same shape: normal, edge and two refusal cases. The input names follow the formulas page.

Two things differ from the first ten. A function may call another: `fin.late_filing_penalty@1` calls `date.late_days@1` and `fin.monthly_interest@1` calls `fin.interest_step@1`, each by name and version, so their cases rest on the called version and never on a newer one. Text and yes-or-no values appear: a text value and a text result are written without quotes in the tables and in the data files (`status active`, result `B`), a yes or no is `true` or `false`, and a float input is written with its trailing `f` on this page and as digits only (`0.37`) in the data files, because the reader of a received float takes digits.

The names, the libraries `edu.` and `member.` and the prefix `fin.` for `interest_step` are proposals (Open question 8). Every function below is a proposal of the same kind as the first ten: the owner has not decided its name or its rules beyond the formulas page.

### fin.loan_within_limit@1

Signature: `fin.loan_within_limit@1(status: text, requested: decimal, contributions: list of decimal) -> boolean`. Formula 2.

Rule: whether a loan may be granted: the member is active and the amount asked is not more than three times the contributions. Status is one of active, lapsed or suspended, requested is more than 0.00 with 2 places, contributions has at least one value and no negatives.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.loan_within_limit-01 | normal | `status active, requested 1400.00, contributions [200.00, 250.00, 150.00, 100.00]` | `true` | S-01 §2 | Contributions 200.00 + 250.00 + 150.00 + 100.00 = 700.00. Three times is 2100.00. 1400.00 is not more than 2100.00 and the status is active, so true. |
| REF-fin.loan_within_limit-02 | normal | `status active, requested 2500.00, contributions [200.00, 250.00, 150.00, 100.00]` | `false` | S-03 | Three times 700.00 is 2100.00. 2500.00 is more than 2100.00, so false. |
| REF-fin.loan_within_limit-03 | edge | `status lapsed, requested 100.00, contributions [500.00]` | `false` | S-03 | 100.00 is far below 1500.00, but the status is lapsed and not active, so false. |
| REF-fin.loan_within_limit-04 | edge | `status active, requested 2100.00, contributions [700.00]` | `true` | S-03 | Three times 700.00 is 2100.00. The amount equals the limit and the test is not more than, so true. |
| REF-fin.loan_within_limit-05 | edge | `status active, requested 2100.01, contributions [700.00]` | `false` | S-03 | 2100.01 is one cent above 2100.00, so false. |
| REF-fin.loan_within_limit-06 | edge | `status active, requested 0.01, contributions [0.00]` | `false` | S-03 | Contributions of 0.00 give a limit of 0.00. 0.01 is above it, so false. |
| REF-fin.loan_within_limit-07 | refusal | `status paused, requested 1400.00, contributions [200.00, 250.00, 150.00, 100.00]` | `refused` | S-02 §2 | paused is not one of the allowed values active, lapsed or suspended. |
| REF-fin.loan_within_limit-08 | refusal | `status active, requested 0.00, contributions [200.00, 250.00, 150.00, 100.00]` | `refused` | S-02 §2 | requested 0.00 breaks the more than 0.00 rule. |

### fin.farmer_net_payout@1

Signature: `fin.farmer_net_payout@1(kilos: integer, price_per_kilo: decimal, credit: decimal) -> decimal`. Formula 4.

Rule: the payment for the kilos delivered, rounded to 2 places, less the input credit, and never below 0.00. Kilos has no negatives, the price is more than 0.000 with at most 3 places, credit has no negatives and 2 places.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.farmer_net_payout-01 | normal | `kilos 853, price_per_kilo 0.437, credit 120.00` | `252.76` | S-01 §4 | 853 x 0.437 = 372.761, rounded to 372.76. 372.76 - 120.00 = 252.76. |
| REF-fin.farmer_net_payout-02 | normal | `kilos 853, price_per_kilo 0.437, credit 400.00` | `0.00` | S-01 §4 | 372.76 - 400.00 = -27.24, which is below zero, so the result is 0.00. |
| REF-fin.farmer_net_payout-03 | edge | `kilos 0, price_per_kilo 0.437, credit 0.00` | `0.00` | S-03 | 0 kilos pay 0.00. 0.00 - 0.00 = 0.00. |
| REF-fin.farmer_net_payout-04 | edge | `kilos 100, price_per_kilo 1.000, credit 100.00` | `0.00` | S-03 | 100 x 1.000 = 100.000, rounded 100.00. 100.00 - 100.00 = 0.00. A credit equal to the payment leaves nothing. |
| REF-fin.farmer_net_payout-05 | edge | `kilos 1, price_per_kilo 0.005, credit 0.00` | `0.01` | S-03 | 1 x 0.005 = 0.005, exactly on the half of 0.01. Away from zero gives 0.01. Mode-dependent: half-even and half-down give 0.00. |
| REF-fin.farmer_net_payout-06 | edge | `kilos 1000000, price_per_kilo 0.001, credit 0.00` | `1000.00` | S-03 | 1000000 x 0.001 = 1000.000, rounded 1000.00. A large number of kilos at the smallest price. |
| REF-fin.farmer_net_payout-07 | refusal | `kilos 853, price_per_kilo 0.4375, credit 120.00` | `refused` | S-02 §4 | 0.4375 has four places and breaks the at most 3 decimal places rule. |
| REF-fin.farmer_net_payout-08 | refusal | `kilos 853, price_per_kilo 0.000, credit 120.00` | `refused` | S-02 §4 | 0.000 breaks the more than 0.000 rule. |

### fin.late_filing_penalty@1

Signature: `fin.late_filing_penalty@1(amount: decimal, due: date, filed: date) -> decimal`. Formula 6.

Rule: a penalty on the amount due that steps up with the days late counted by `date.late_days@1`: 0.00 when on time, 2 percent up to 30 days, 5 percent from 31 to 90, 10 percent above 90, each rounded to 2 places. It calls `date.late_days@1` and so needs that function in the library.

Provisional: example-formulas.md writes the call as `late_days@1`; the name `date.late_days@1` follows the reference of the first ten functions and is a proposal (Open question 8).

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.late_filing_penalty-01 | normal | `amount 8400.00, due 2026-03-31, filed 2026-05-15` | `420.00` | S-01 §6 | Filed 45 days late. 45 is not 0, not above 90, above 30, so 5 percent. 8400.00 x 0.05 = 420.00. |
| REF-fin.late_filing_penalty-02 | normal | `amount 8400.00, due 2026-03-31, filed 2026-04-12` | `168.00` | S-01 §6 | Filed 12 days late. 12 is not above 30, so 2 percent. 8400.00 x 0.02 = 168.00. |
| REF-fin.late_filing_penalty-03 | edge | `amount 8400.00, due 2026-03-31, filed 2026-03-31` | `0.00` | S-03 | 0 days late, so no penalty: 0.00. |
| REF-fin.late_filing_penalty-04 | edge | `amount 8400.00, due 2026-03-31, filed 2026-04-30` | `168.00` | S-03 | 30 days late. 30 is not above 30, so 2 percent: 168.00. |
| REF-fin.late_filing_penalty-05 | edge | `amount 8400.00, due 2026-03-31, filed 2026-05-01` | `420.00` | S-03 | 31 days late. 31 is above 30 and not above 90, so 5 percent: 420.00. |
| REF-fin.late_filing_penalty-06 | edge | `amount 8400.00, due 2026-03-31, filed 2026-06-29` | `420.00` | S-03 | 30 + 31 + 29 = 90 days late. 90 is not above 90, so 5 percent: 420.00. |
| REF-fin.late_filing_penalty-07 | edge | `amount 8400.00, due 2026-03-31, filed 2026-06-30` | `840.00` | S-03 | 91 days late. 91 is above 90, so 10 percent: 840.00. |
| REF-fin.late_filing_penalty-08 | refusal | `amount 8400.005, due 2026-03-31, filed 2026-05-15` | `refused` | S-02 §6 | 8400.005 has three places and breaks the 2 decimal places rule. |

### fin.interest_step@1

Signature: `fin.interest_step@1(balance: decimal) -> decimal`. Formula 7.

Rule: the monthly interest rate of the band the balance falls in: 0.010 under 1000.00, 0.015 from 1000.00 under 5000.00, 0.020 from 5000.00. Balance has no negatives and 2 places.

Provisional: example-formulas.md publishes this as `interest_step@1` with no library. The prefix `fin.` is proposed here (Open question 8).

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.interest_step-01 | normal | `balance 3217.45` | `0.015` | S-01 §7 | 3217.45 is not under 1000.00 but is under 5000.00, so 0.015. |
| REF-fin.interest_step-02 | normal | `balance 640.00` | `0.010` | S-01 §7 | 640.00 is under 1000.00, so 0.010. |
| REF-fin.interest_step-03 | edge | `balance 5000.00` | `0.020` | S-01 §7 | 5000.00 is not under 5000.00, so the top band: 0.020. |
| REF-fin.interest_step-04 | edge | `balance 1000.00` | `0.015` | S-03 | 1000.00 is not under 1000.00, so the middle band: 0.015. |
| REF-fin.interest_step-05 | edge | `balance 999.99` | `0.010` | S-03 | 999.99 is under 1000.00, so 0.010. |
| REF-fin.interest_step-06 | edge | `balance 4999.99` | `0.015` | S-03 | 4999.99 is under 5000.00, so 0.015. |
| REF-fin.interest_step-07 | refusal | `balance -1.00` | `refused` | S-02 §7 | -1.00 breaks the no negatives rule. |
| REF-fin.interest_step-08 | refusal | `balance 1000.005` | `refused` | S-02 §7 | 1000.005 has three places and breaks the 2 decimal places rule. |

### fin.monthly_interest@1

Signature: `fin.monthly_interest@1(balance: decimal) -> decimal`. Formula 8.

Rule: the interest charged for a month: the balance times the rate of `fin.interest_step@1`, rounded to 2 places. It calls `fin.interest_step@1`. Balance has no negatives and 2 places.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.monthly_interest-01 | normal | `balance 3217.45` | `48.26` | S-01 §8 | Rate 0.015. 3217.45 x 0.015 = 48.26175, rounded 48.26. |
| REF-fin.monthly_interest-02 | normal | `balance 640.00` | `6.40` | S-03 | Rate 0.010. 640.00 x 0.010 = 6.400, rounded 6.40. |
| REF-fin.monthly_interest-03 | normal | `balance 5000.00` | `100.00` | S-03 | Rate 0.020. 5000.00 x 0.020 = 100.000, rounded 100.00. |
| REF-fin.monthly_interest-04 | edge | `balance 0.00` | `0.00` | S-03 | Rate 0.010. 0.00 x 0.010 = 0.00. |
| REF-fin.monthly_interest-05 | edge | `balance 999.99` | `10.00` | S-03 | Rate 0.010. 999.99 x 0.010 = 9.9999, rounded 10.00. The top of the lowest band. |
| REF-fin.monthly_interest-06 | edge | `balance 0.50` | `0.01` | S-03 | Rate 0.010. 0.50 x 0.010 = 0.005, exactly on the half of 0.01. Away from zero gives 0.01. Mode-dependent: half-even and half-down give 0.00. |
| REF-fin.monthly_interest-07 | edge | `balance 1000.00` | `15.00` | S-03 | Rate 0.015. 1000.00 x 0.015 = 15.000, rounded 15.00. The first balance of the middle band. |
| REF-fin.monthly_interest-08 | refusal | `balance 3217.455` | `refused` | S-02 §8 | 3217.455 has three places and breaks the 2 decimal places rule. |

### edu.term_grade@1

Signature: `edu.term_grade@1(scores: list of integer) -> text`. Formula 10.

Rule: a letter grade from the average of the term's scores: A from 80, B from 65, C from 50, F below 50. Scores has at least one value and each is 0 to 100.

Provisional: The prefix `edu.` is a proposal (Open question 8).

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-edu.term_grade-01 | normal | `scores [72, 85, 64, 91, 77]` | `B` | S-01 §10 | 72 + 85 + 64 + 91 + 77 = 389. 389 / 5 = 77.8. Not 80 or more, 65 or more, so B. |
| REF-edu.term_grade-02 | normal | `scores [90, 85, 80]` | `A` | S-03 | 90 + 85 + 80 = 255. 255 / 3 = 85. 85 is 80 or more, so A. |
| REF-edu.term_grade-03 | edge | `scores [80]` | `A` | S-03 | One score of 80. The average is 80, which is 80 or more, so A. |
| REF-edu.term_grade-04 | edge | `scores [65]` | `B` | S-03 | One score of 65. The average 65 is not 80 or more and is 65 or more, so B. |
| REF-edu.term_grade-05 | edge | `scores [50]` | `C` | S-03 | One score of 50. The average 50 is not 65 or more and is 50 or more, so C. |
| REF-edu.term_grade-06 | edge | `scores [49]` | `F` | S-03 | One score of 49. The average 49 is under 50, so F. |
| REF-edu.term_grade-07 | refusal | `scores [72, 104]` | `refused` | S-02 §10 | 104 is above 100 and breaks the range rule 0 to 100. |
| REF-edu.term_grade-08 | refusal | `scores []` | `refused` | S-02 §10 | An empty list breaks the at least one rule. |

### fin.interim_payment@1

Signature: `fin.interim_payment@1(contract: decimal, percent_complete: float, paid: decimal) -> decimal`. Formula 14.

Rule: the certified share of a contract, with ten percent held back, less what was already paid, never below 0.00. The float share becomes a decimal of 2 places first. Contract is more than 0.00, percent complete is 0.0f to 1.0f, paid has no negatives; money has 2 places.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.interim_payment-01 | normal | `contract 183500.00, percent_complete 0.37f, paid 52000.00` | `9105.50` | S-01 §14 | The float 0.37f becomes the decimal 0.37. 183500.00 x 0.37 = 67895.00. Times 0.90 is 61105.50. Less 52000.00 is 9105.50. |
| REF-fin.interim_payment-02 | normal | `contract 100000.00, percent_complete 0.50f, paid 0.00` | `45000.00` | S-03 | 100000.00 x 0.50 = 50000.00. Times 0.90 is 45000.00. Less 0.00 is 45000.00. |
| REF-fin.interim_payment-03 | edge | `contract 183500.00, percent_complete 0.37f, paid 70000.00` | `0.00` | S-03 | 61105.50 - 70000.00 = -8894.50, which is below zero, so 0.00. More was paid than is now due. |
| REF-fin.interim_payment-04 | edge | `contract 5000.00, percent_complete 0.0f, paid 0.00` | `0.00` | S-03 | 0.0f is the decimal 0.00. 5000.00 x 0.00 = 0.00. The lowest end of the range. |
| REF-fin.interim_payment-05 | edge | `contract 2000.00, percent_complete 1.0f, paid 0.00` | `1800.00` | S-03 | 1.0f is the decimal 1.00. 2000.00 x 1.00 x 0.90 = 1800.00. The highest end of the range: ten percent is still held back. |
| REF-fin.interim_payment-06 | edge | `contract 1000.00, percent_complete 0.375f, paid 0.00` | `342.00` | S-03 | 0.375f is exactly on the half of 0.01 when made a decimal of 2 places. Away from zero gives 0.38. 1000.00 x 0.38 x 0.90 = 342.00. Mode-dependent: half-even gives 0.38 and 342.00, half-down gives 0.37 and 333.00. |
| REF-fin.interim_payment-07 | refusal | `contract 183500.00, percent_complete 1.2f, paid 52000.00` | `refused` | S-02 §14 | 1.2f is above 1.0f and breaks the range rule 0.0f to 1.0f. |
| REF-fin.interim_payment-08 | refusal | `contract 0.00, percent_complete 0.37f, paid 0.00` | `refused` | S-02 §14 | contract 0.00 breaks the more than 0.00 rule. |

### fin.lending_data_ok@1

Signature: `fin.lending_data_ok@1(balances: list of decimal, floor: decimal) -> boolean`. Formula 16.

Rule: whether the weekly balances of a statement are enough and good enough: at least 6 of them, an average of at least the floor, and none below 0.00. Balances has at least one value with 2 places each and may hold a negative one, which the check then fails; floor has no negatives and 2 places.

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-fin.lending_data_ok-01 | normal | `balances [420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00], floor 350.00` | `true` | S-01 §16 | 8 balances sum to 3001.00 and the average is 375.125. 8 is at least 6, 375.125 is at least 350.00, and the lowest, 95.25, is not below 0.00, so true. |
| REF-fin.lending_data_ok-02 | normal | `balances [420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00], floor 400.00` | `false` | S-03 | The average 375.125 is below the floor 400.00, so false. |
| REF-fin.lending_data_ok-03 | edge | `balances [500.00, 500.00, 500.00, 500.00, 500.00], floor 100.00` | `false` | S-03 | 5 balances are fewer than 6, so false even though the average is above the floor. |
| REF-fin.lending_data_ok-04 | edge | `balances [350.00, 350.00, 350.00, 350.00, 350.00, 350.00], floor 350.00` | `true` | S-03 | Exactly 6 balances, an average of 350.00 that equals the floor, and a lowest of 350.00: true. |
| REF-fin.lending_data_ok-05 | edge | `balances [400.00, 400.00, 400.00, 400.00, 400.00, -0.01], floor 100.00` | `false` | S-03 | 6 balances, an average of 333.33 above the floor, but the lowest is -0.01, which is below 0.00, so false. A negative balance is accepted as input and then fails the check. |
| REF-fin.lending_data_ok-06 | edge | `balances [0.00, 0.00, 0.00, 0.00, 0.00, 0.00], floor 0.00` | `true` | S-03 | 6 balances of 0.00: the average 0.00 equals the floor 0.00 and the lowest 0.00 is not below 0.00, so true. |
| REF-fin.lending_data_ok-07 | refusal | `balances [420.00, n/a], floor 350.00` | `refused` | S-02 §16 | n/a is text in a list of decimal and breaks the type rule. |
| REF-fin.lending_data_ok-08 | refusal | `balances [], floor 350.00` | `refused` | S-02 §16 | An empty list breaks the at least one rule. |

### member.voting_right@1

Signature: `member.voting_right@1(joined: date, on: date, paid_up: boolean, suspended: boolean, role: text) -> boolean`. Formula 18.

Rule: whether a member may vote: not suspended, not a guest, and either an officer or paid up with six or more whole months of membership. Role is one of member, officer or guest.

Provisional: The prefix `member.` is a proposal (Open question 8). How `months_between` counts a joining on the 29th to 31st into a shorter month is not stated (Open question 6).

Tolerance: exact.

| Id | Kind | Inputs | Expected | Source | Working |
|---|---|---|---|---|---|
| REF-member.voting_right-01 | normal | `joined 2026-01-20, on 2026-09-15, paid_up true, suspended false, role member` | `true` | S-01 §18 | 20 Jan to 15 Sep is 7 whole months. Not suspended, not a guest, paid up with 7 months, which is 6 or more, so true. |
| REF-member.voting_right-02 | normal | `joined 2026-01-20, on 2026-09-15, paid_up true, suspended true, role member` | `false` | S-01 §18 | The same member suspended: the first test fails, so false. |
| REF-member.voting_right-03 | edge | `joined 2026-09-15, on 2026-09-15, paid_up false, suspended false, role officer` | `true` | S-03 | An officer needs neither payment nor months: not suspended, not a guest, officer, so true on the day of joining. |
| REF-member.voting_right-04 | edge | `joined 2020-01-01, on 2026-01-01, paid_up true, suspended false, role guest` | `false` | S-03 | A guest never votes, however long and however paid: false. |
| REF-member.voting_right-05 | edge | `joined 2026-03-15, on 2026-09-15, paid_up true, suspended false, role member` | `true` | S-03 | 15 Mar to 15 Sep is exactly 6 whole months, which is 6 or more, so true. |
| REF-member.voting_right-06 | edge | `joined 2026-03-16, on 2026-09-15, paid_up true, suspended false, role member` | `false` | S-03 | 16 Mar to 15 Sep is 5 whole months and 30 days, so 5, which is under 6, so false. |
| REF-member.voting_right-07 | refusal | `joined 2026-01-20, on 2026-09-15, suspended false, role member, paid_up not given` | `refused` | S-02 §18 | paid_up is required and is not given. |
| REF-member.voting_right-08 | refusal | `joined 2026-01-20, on 2026-09-15, paid_up true, suspended false, role visitor` | `refused` | S-02 §18 | visitor is not one of the allowed values member, officer or guest. |

## Sources

One row per source id used. The next writer copies these rows to [sources.md](../1.%20requirements/sources.md). The case lists give the function and the case numbers.

| Id | Kind | Where | Cases resting on it |
|---|---|---|---|
| S-01 | In-repo worked example | `documentation/1. requirements/example-formulas.md`, the "Worked example" paragraph of sections 1 to 19 | fin.balance_after_payment 01, 02; fin.equal_share 01; date.late_days 01, 02; fin.flat_instalment 01; fin.prorate_fee 01; fin.straight_line_value 01; fin.coverage_ratio 01; fin.proportional_share 01; stats.spread 01; math.remainder 01; fin.loan_within_limit 01; fin.farmer_net_payout 01, 02; fin.late_filing_penalty 01, 02; fin.interest_step 01 to 03; fin.monthly_interest 01; edu.term_grade 01; fin.interim_payment 01; fin.lending_data_ok 01; member.voting_right 01, 02 |
| S-02 | In-repo input rules and invalid examples | `documentation/1. requirements/example-formulas.md`, the input table and the "Invalid example" paragraph of the same sections | fin.balance_after_payment 07, 08; fin.equal_share 06, 07, 08; date.late_days 07, 08; fin.flat_instalment 07, 08; fin.prorate_fee 07, 08; fin.straight_line_value 07, 08; fin.coverage_ratio 07, 08; fin.proportional_share 07; stats.spread 07, 08; math.remainder 06, 07, 08; fin.loan_within_limit 07, 08; fin.farmer_net_payout 07, 08; fin.late_filing_penalty 08; fin.interest_step 07, 08; fin.monthly_interest 08; edu.term_grade 07, 08; fin.interim_payment 07, 08; fin.lending_data_ok 07, 08; member.voting_right 07, 08 |
| S-03 | Hand calculation | The Working column of the case, on this page, shown step by step | fin.balance_after_payment 03 to 06; fin.equal_share 02 to 05; date.late_days 03 to 06; fin.flat_instalment 02 to 06; fin.prorate_fee 02 to 06; fin.straight_line_value 02 to 06; fin.coverage_ratio 02 to 06; fin.proportional_share 02 to 06 and 08; stats.spread 02 to 06; math.remainder 02 to 05; fin.loan_within_limit 02 to 06; fin.farmer_net_payout 03 to 06; fin.late_filing_penalty 03 to 07; fin.interest_step 04 to 06; fin.monthly_interest 02 to 07; edu.term_grade 02 to 06; fin.interim_payment 02 to 06; fin.lending_data_ok 02 to 06; member.voting_right 03 to 06 |

No outside source was opened for this page, so no public worked example is cited. A figure taken from one later gets the next id, with its address, and the case is put in the writer's own words.

## Open questions

These are for the owner. None is answered on this page.

1. Rounding mode default. This page uses half away from zero, as example-formulas.md states. Eight cases sit on a half and change with the mode (table under Tolerances). The decision is exact-money.md open question 1.
2. Refusal wording. A refused case passes on "no result and one plain sentence". The sentence is not fixed, and neither is how an input that is not a date, such as 2026-02-30, is worded.
3. Float tolerance against zero. A relative error is meaningless when the expected result is `0.0f`. This page asks for exactly `0.0f`. Is that the rule, or does the float function carry an absolute error too?
4. Float ratio on a half. 100.05 / 100.00 is exactly 1.0005 as a decimal, and half away from zero gives 1.001 at 3 places, but dividing the two as binary floats and rounding gives 1.0. No case is written until the owner says which is right (exact-money.md open question 8, the same question for a conversion).
5. A negative share. `fin.equal_share@1` applies no clamp: expenses above deposits give a negative result by the formula. Whether that is allowed or refused is not stated, so there is no case.
6. Dates in the wrong order and leap anniversaries. `fin.prorate_fee@1` with `joined` after `period_end`, and `fin.straight_line_value@1` with `on` before `bought`, have no stated result. How `years_between` counts an anniversary of 29 February in a year with no 29 February is also not stated. No case until decided.
7. Units above the total, and a total of zero. `fin.proportional_share@1` with `units` above the sum of `all_units` gives a share above the call amount, and `all_units` of `[0]` divides by zero. Case 08 refuses the second. Does the formulas page add an input rule (the sum is more than 0), and is the first allowed?
8. Names. The prefix `date.` for `late_days@1`, the prefix `fin.` for `interest_step@1`, the libraries `edu.` for `term_grade` and `member.` for `voting_right`, the names of the other six functions of the nine, the general names `stats.spread` with `values` and `math.remainder` with `value` and `divisor`, and the range 1 to 100 for the divisor (it comes from bag_size) are proposals. Whether a general `stats` or `math` function takes negatives, as the formulas page's no negatives rule would not allow, is not decided.
9. Places per currency. Every money case uses 2 places, as exact-money.md open question 3 assumes.
10. Negative balances in `fin.lending_data_ok@1`. Formula 16 gives `balances` no no negatives rule, so a negative balance is accepted as input and then fails the function's own test `min(balances) >= 0.00` (case 05 records this). With a no negatives rule the test could never fail and a negative balance would be refused instead. Which is meant?
11. A float made a decimal on a half. `fin.interim_payment@1` case 06 uses 0.375f, which a binary float holds exactly, so it sits exactly on the half of 0.01 when made a decimal of 2 places. A float that is only near a half (0.285f) is the question of open question 4 and has no case.
12. Calling another function. `fin.late_filing_penalty@1` and `fin.monthly_interest@1` call `date.late_days@1` and `fin.interest_step@1` at version 1. A fix to a called function ships as its next version, so the cases of the callers stay true; whether a caller must move to the new version of its callee, and when, is the builder's choice (decision 10) and has no case here.

## How the figures were checked

Every figure was worked by hand first, as shown in the Working column. A throwaway script outside the repo then recomputed all 152 cases with Python's decimal module at 100 digits, rounding half away from zero (ROUND_HALF_UP), applying each function's input rules to decide the refusals, counting whole years and days with the calendar, and computing the float ratio in binary floats against the exact decimal ratio at a relative error of 1e-9. It also re-ran the cases under half-even and half-down to find the eight mode-dependent ones, and it read this page back to check that each case's kind, inputs and expected result match its own, that every id appears once, and that each source id's case list above is right.

The 72 cases of the other nine functions were checked the same way after the first 80: each was worked by hand with the Working shown, a second throwaway script (outside the repo) re-implemented the nine functions with Python's decimal module and the calendar, recomputed every expected figure and every refusal, and checked that each case marked mode-dependent changes under half-even or half-down and that no other case does. The data files and the tables of the section The other nine functions were then written from the same list, so a case has one text. No figure was taken from the library under test.
