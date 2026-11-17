# Basic runs: every part of the language once

Status: the owner has not reviewed this page. Every number marked "proposed" is a starting point, not a decision. All names and figures are invented.

## What this page protects

A formula is built from small parts: operators, comparisons, a conditional, seven types, conversions, aggregates, built-in functions and calls. When a long formula gives a wrong answer, the fault should be found at the part that broke. So every part of the language on [formula-language.md](../2.%20design/formula-language.md) is run alone at least once, with exact inputs and an exact expected result.

The page is a coverage table. Each row is one part, and the row names the case that covers it. A part with no case is listed under "Gaps", never left out silently. The page holds single runs only. Many rows in one call are on [bulk-runs.md](bulk-runs.md), refusals of limits and hostile text on [hostile-formulas.md](hostile-formulas.md), money rounding on [exact-money.md](exact-money.md) and the reference values of library functions on [library-reference.md](library-reference.md).

## How to read a row

- A case name that starts with `lang-` or `rules-` is a row of [formula-examples.md](formula-examples.md). Its formula, inputs and result are there and are not copied here.
- A case name that starts with `basic-` is new on this page. Its formula, inputs and expected result are in the table of its section, so the case loader reads them from here once it reads this page (open question 1).
- Values are written in the notation of formula-examples.md: `int:7`, `dec:12.50`, `float:1.5`, `text:"a"`, `bool:true`, `date:2026-03-31`, `list<dec>:[1.00, 2.00]`. The loader does not have `float` or `list` yet; the proposed syntax is the one on formula-examples.md, "Rows that need a value type the loader does not have".
- `problems: code` lists the codes in the order the engine lists them. The words of each sentence are on the language page.
- Inputs that reach the formula as text are Validate rows and belong to [formula-examples.md](formula-examples.md). The `basic-` cases here give typed values, the way a caller that has already read its inputs would.

## What the cases assume

| Setting | Value taken here | From |
|---|---|---|
| Rounding rule | Half up, so 2.345 to 2 places is 2.35. | Decisions 14 and 20 |
| Division | Exactly 20 places, rounded at the 20th. | Language page; decision 21 proposes 20 as the default of an agency setting |
| Negative inputs | Refused unless the function declares the input may be negative. | Decision 22 |
| Decimal places | Two, as every case here is written. | Decision 15 |

Where a case depends on a setting an agency can change, it is written for the default and the page says so. Cases for other settings are under "Held cases".

## 1. Operators

The language has the operators `+ - * / %` and the four comparisons and `== !=` and `&& || !` and the conditional. It has no negate operator and no power operator, so those two rows of the issue are covered by a refusal and by the way a negative is written (the language page, "Left out on purpose").

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| add, integers | `basic-add-integers` | `a + b` | a=int:7, b=int:5 | result: int:12 | Two integers give an integer. |
| add, decimals | `basic-add-decimals` | `a + b` | a=dec:120.50, b=dec:45.2 | result: dec:165.70 | The result has as many places as the operand with more places. |
| add, integer and decimal | `basic-add-integer-decimal` | `a + b` | a=int:3, b=dec:0.50 | result: dec:3.50 | An integer and a decimal give a decimal. |
| subtract, integers | `basic-subtract-integers` | `a - b` | a=int:5, b=int:8 | result: int:-3 | A result may be negative; only a literal has no sign. |
| subtract, decimals | `lang-f01-balance-after-payment` | | | | See formula-examples.md. |
| multiply, integers | `basic-multiply-integers` | `a * b` | a=int:6, b=int:7 | result: int:42 | Two integers give an integer. |
| multiply, decimal and integer | `lang-f04-farmer-net-payout` | | | | `0.437 * 853` is `372.761` inside it. |
| multiply, decimals | `basic-multiply-decimals` | `a * b` | a=dec:1.25, b=dec:0.40 | result: dec:0.5000 | Multiplying adds the places: 2 and 2 give 4. |
| divide, decimal by integer | `lang-decimal-division-twenty-places` | | | | 10.00 / 3 gives 20 places. |
| divide, integer by decimal | `basic-divide-integer-decimal` | `a / b` | a=int:7, b=dec:2.0 | result: dec:3.50000000000000000000 | A decimal quotient has exactly 20 places. |
| divide, floats | `basic-divide-floats` | `a / b` | a=float:1.5, b=float:0.5 | result: float:3 | Two floats give a float. |
| divide, two integers refused | `basic-divide-integers-refused` | `a / b` | a=int:7, b=int:2 | problems: type-mismatch | `/` needs a decimal or a float on at least one side. |
| divide by zero | `lang-problem-divide-by-zero` | | | | Never a value. |
| remainder | `lang-f19-seed-bags-left-over` | | | | 1237 % 50 is 37. |
| remainder, negative dividend | `lang-remainder-negative-dividend` | | | | -7 % 3 is -1. |
| remainder, negative divisor | `basic-remainder-negative-divisor` | `a % b` | a=int:7, b=int:-3 | result: int:1 | The sign is the dividend's, and the size is below the divisor's. |
| remainder by zero | `basic-remainder-by-zero` | `a % b` | a=int:5, b=int:0 | problems: divide-by-zero | `%` by zero is a problem, never a value. |
| remainder on a decimal refused | `basic-remainder-decimal-refused` | `a % b` | a=dec:5.5, b=int:2 | problems: type-mismatch | `%` takes two integers. |
| negate | `basic-negate-by-subtraction` | `0 - x` | x=dec:12.50 | result: dec:-12.50 | There is no unary minus. A negative is written as a subtraction from `0`. |
| negate, written with a minus sign | none | | | | Gap 1. |
| power | `basic-power-sign-refused` | `a ^ 2` | a=int:3 | problems: bad-syntax | `^` is not in the language. Sentence: `The formula cannot be read: unexpected character ^ at line 1, column 3.` |
| power, as a function | `basic-power-function-refused` | `pow(a, 2)` | a=int:3 | problems: unknown-function | There is no math function beyond `round`. |

## 2. Precedence and grouping

Each pair gives different results from the same inputs, so a wrong precedence changes the answer.

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| `*` before `+` | `basic-precedence-multiply-first` | `a + b * c` | a=int:2, b=int:3, c=int:4 | result: int:14 | 3 * 4 is worked out first. |
| brackets over `*` | `basic-precedence-brackets-first` | `(a + b) * c` | a=int:2, b=int:3, c=int:4 | result: int:20 | Brackets group. |
| `-` left to right | `basic-precedence-subtract-left` | `a - b - c` | a=int:10, b=int:3, c=int:2 | result: int:5 | Read as `(a - b) - c`. |
| `-` with brackets right | `basic-precedence-subtract-bracket-right` | `a - (b - c)` | a=int:10, b=int:3, c=int:2 | result: int:9 | The brackets are kept. |
| `/` and `*` left to right | `basic-precedence-divide-then-multiply` | `a / b * c` | a=dec:12.00, b=int:4, c=int:3 | result: dec:9.00000000000000000000 | (12.00 / 4) * 3: 3.00000000000000000000 times 3. |
| `/` with brackets right | `basic-precedence-divide-bracket-right` | `a / (b * c)` | a=dec:12.00, b=int:4, c=int:3 | result: dec:1.00000000000000000000 | 12.00 / 12. |
| `%` and `*` left to right | `basic-precedence-remainder-then-multiply` | `a % b * c` | a=int:17, b=int:5, c=int:3 | result: int:6 | (17 % 5) * 3 is 2 * 3. |
| arithmetic before `==` | `basic-precedence-arithmetic-before-equal` | `a + b == c * d` | a=int:2, b=int:3, c=int:1, d=int:5 | result: bool:true | 5 == 5. |
| comparisons do not chain | `basic-precedence-chained-comparison` | `a < b < c` | a=int:1, b=int:2, c=int:3 | problems: type-mismatch | Read as `(a < b) < c`, a boolean against a number. |
| `!` before `&&` | `basic-precedence-not-first` | `!a && b` | a=bool:true, b=bool:false | result: bool:false | (!true) && false. |
| `!` over a bracket | `basic-precedence-not-bracket` | `!(a && b)` | a=bool:true, b=bool:false | result: bool:true | !(true && false). |
| `&&` before `\|\|` | `basic-precedence-and-before-or` | `a \|\| b && c` | a=bool:true, b=bool:false, c=bool:false | result: bool:true | true \|\| (false && false). |
| `\|\|` with brackets first | `basic-precedence-or-bracket` | `(a \|\| b) && c` | a=bool:true, b=bool:false, c=bool:false | result: bool:false | (true \|\| false) && false. |
| `\|\|` before `?` | `basic-precedence-or-before-conditional` | `a \|\| b ? 1 : 2` | a=bool:false, b=bool:true | result: int:1 | The whole `a \|\| b` is the condition. |
| conditional, else to the right | `basic-precedence-conditional-chain` | `a ? 1 : b ? 2 : 3` | a=bool:false, b=bool:true | result: int:2 | Read as `a ? 1 : (b ? 2 : 3)`. |
| conditional, then is whole | `basic-precedence-conditional-in-then` | `a ? b ? 1 : 2 : 3` | a=bool:true, b=bool:false | result: int:2 | Read as `a ? (b ? 1 : 2) : 3`. |
| spacing does not change meaning | `lang-save-tight-spacing` | | | | Also `lang-save-extra-brackets`. |

## 3. Comparison and boolean logic

Every comparison is run for a true case and a false case, with a boundary where the two operands are equal.

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| `<` true | `basic-less-true` | `a < b` | a=int:2, b=int:3 | result: bool:true | |
| `<` equal | `basic-less-equal-operands` | `a < b` | a=int:3, b=int:3 | result: bool:false | An equal pair is not less. |
| `<=` equal | `basic-less-or-equal-equal-operands` | `a <= b` | a=int:3, b=int:3 | result: bool:true | |
| `<=` false | `basic-less-or-equal-false` | `a <= b` | a=int:4, b=int:3 | result: bool:false | |
| `>` true | `basic-greater-true` | `a > b` | a=int:4, b=int:3 | result: bool:true | |
| `>` equal | `basic-greater-equal-operands` | `a > b` | a=int:3, b=int:3 | result: bool:false | |
| `>=` equal | `basic-greater-or-equal-equal-operands` | `a >= b` | a=int:3, b=int:3 | result: bool:true | The loan formula depends on this one. |
| `>=` false | `basic-greater-or-equal-false` | `a >= b` | a=int:2, b=int:3 | result: bool:false | |
| `==` on numbers | `lang-decimal-equal-by-value` | | | | 12.5 equals 12.50. |
| `==` on integer and decimal | `basic-equal-integer-decimal` | `a == b` | a=int:3, b=dec:3.00 | result: bool:true | Compared by value. |
| `==` on texts | `lang-f02-loan-limit-check` | | | | `status == "active"` inside it. |
| `==` on texts, different case | `basic-equal-text-case` | `a == b` | a=text:"Active", b=text:"active" | result: bool:false | Exact characters. |
| `!=` true | `basic-not-equal-true` | `a != b` | a=dec:1.00, b=dec:1.01 | result: bool:true | |
| `!=` false | `basic-not-equal-false` | `a != b` | a=dec:12.5, b=dec:12.50 | result: bool:false | Equal by value. |
| comparison of decimal and float | `basic-compare-decimal-float` | `a < b` | a=dec:1.00, b=float:2.5 | problems: float-decimal-mix | Convert one side first. |
| `==` on floats | `basic-equal-floats-refused` | `a == b` | a=float:1.5, b=float:1.5 | problems: type-mismatch | `==` on floats is left out. |
| `==` on an integer and a float | `basic-equal-integer-float-refused` | `a == b` | a=int:1, b=float:1.0 | problems: type-mismatch | Left out. |
| ordering of texts | `basic-order-text-refused` | `a < b` | a=text:"ant", b=text:"bee" | problems: type-mismatch | Texts are compared with `==` and `!=` only. |
| text against a number | `basic-equal-text-number-refused` | `a == 1` | a=text:"1" | problems: type-mismatch | |
| `&&` true and true | `basic-and-true-true` | `a && b` | a=bool:true, b=bool:true | result: bool:true | |
| `&&` true and false | `basic-and-true-false` | `a && b` | a=bool:true, b=bool:false | result: bool:false | |
| `&&` false and true | `basic-and-false-true` | `a && b` | a=bool:false, b=bool:true | result: bool:false | |
| `&&` false and false | `basic-and-false-false` | `a && b` | a=bool:false, b=bool:false | result: bool:false | |
| `\|\|` true and true | `basic-or-true-true` | `a \|\| b` | a=bool:true, b=bool:true | result: bool:true | |
| `\|\|` true and false | `basic-or-true-false` | `a \|\| b` | a=bool:true, b=bool:false | result: bool:true | |
| `\|\|` false and true | `basic-or-false-true` | `a \|\| b` | a=bool:false, b=bool:true | result: bool:true | |
| `\|\|` false and false | `basic-or-false-false` | `a \|\| b` | a=bool:false, b=bool:false | result: bool:false | |
| `!` of true | `basic-not-true` | `!a` | a=bool:true | result: bool:false | |
| `!` of false | `basic-not-false` | `!a` | a=bool:false | result: bool:true | |
| `!` twice | `basic-not-twice` | `!!a` | a=bool:true | result: bool:true | |
| a mix of all three | `lang-f18-voting-right` | | | | `!`, `&&`, `\|\|` and a comparison in one formula. |
| `&&` on numbers refused | `basic-and-numbers-refused` | `a && b` | a=int:1, b=int:0 | problems: type-mismatch | `&&` needs two booleans. |
| `&&` does not skip its right side | `basic-and-no-short-circuit` | `a && total / n > 0.50` | a=bool:false, total=dec:1.00, n=int:0 | problems: divide-by-zero | A false left side does not guard the right side. |

## 4. The conditional

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| true branch | `basic-loan-balance-payment-covers` | see section 12 | | | The then branch is chosen. |
| false branch | `lang-f01-balance-after-payment` | | | | The else branch is chosen. |
| nested, three answers, low | `basic-conditional-nested-low` | `x < 10 ? "low" : x < 20 ? "mid" : "high"` | x=int:5 | result: text:"low" | |
| nested, boundary | `basic-conditional-nested-boundary` | `x < 10 ? "low" : x < 20 ? "mid" : "high"` | x=int:10 | result: text:"mid" | 10 < 10 is false, 10 < 20 is true. |
| nested, mid | `basic-conditional-nested-mid` | `x < 10 ? "low" : x < 20 ? "mid" : "high"` | x=int:15 | result: text:"mid" | |
| nested, high | `basic-conditional-nested-high` | `x < 10 ? "low" : x < 20 ? "mid" : "high"` | x=int:25 | result: text:"high" | |
| nested, in the then branch | `lang-save-nested-conditional` | | | | Saved with brackets; the result is `2`. |
| nested, three bands of a penalty | `lang-f06-late-filing-penalty` | | | | Calls an agency function inside. |
| a band table | `lang-f07-interest-rate-step` | | | | Decimal literals keep their digits. |
| the two answers meet in a common type | `lang-conditional-widens-integer` | | | | `1` and `2.50` give a decimal. |
| a branch that would fail is not skipped | `lang-no-short-circuit` | | | | The condition does not guard the division. Both answers are worked out. |
| a condition that is not true or false | `lang-problem-condition-not-boolean` | | | | |
| answers of different kinds | `lang-problem-branches-differ` | | | | |

## 5. Types, as an input and as a result

| Type | As an input | As a result |
|---|---|---|
| integer | `lang-f19-seed-bags-left-over` | `basic-add-integers` |
| decimal | `lang-f01-balance-after-payment` | `lang-f01-balance-after-payment` |
| float | `lang-f14-interim-payment` | `lang-f13-collateral-coverage-ratio` and `basic-float-input-result` |
| text | `lang-f02-loan-limit-check` | `lang-f10-term-grade` |
| boolean | `lang-f18-voting-right` | `lang-f02-loan-limit-check` |
| date | `lang-f05-late-days` | none, Gap 3 |
| list of integer | `lang-f10-term-grade` | none, Gap 4 |
| list of decimal | `lang-f03-fund-share-per-member` | none, Gap 4 |

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| float in, float out | `basic-float-input-result` | `a * b` | a=float:1.5, b=int:2 | result: float:3 | A float and an integer give a float. |
| text out | `basic-text-result` | `status == "active" ? "open" : "closed"` | status=text:"active" | result: text:"open" | |
| text with escapes | `lang-text-escapes` | | | | `\"` and `\\`. |
| boolean out | `basic-and-true-true` | | | | Section 3. |
| a decimal literal keeps its digits | `lang-f07-interest-rate-step` | | | | `0.015` stays three places. |
| a float is always finite | `basic-float-overflow` | `a * b` | a=float:1e200, b=float:1e200 | problems: number-too-big | The notation `1e200` is proposed; the loader does not read it yet. |
| an integer result over 18 digits | `lang-problem-number-too-big` | | | | |

## 6. Conversion between decimal and float

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| decimal to float | `basic-to-float` | `to_float(x)` | x=dec:1.50 | result: float:1.5 | |
| float to decimal | `basic-to-decimal` | `to_decimal(x, 2)` | x=float:0.37 | result: dec:0.37 | The float is read from its shortest text. |
| float to decimal, more places than the text | `basic-to-decimal-three-places` | `to_decimal(x, 3)` | x=float:0.37 | result: dec:0.370 | Exactly `n` places. |
| float to decimal, round at the place | `basic-to-decimal-rounds` | `to_decimal(x, 2)` | x=float:1.005 | result: dec:1.01 | Proposed: read from its shortest text, 1.005, then round half up. |
| float to decimal, no places | `basic-to-decimal-no-places` | `to_decimal(x, 0)` | x=float:2.5 | result: dec:3 | Proposed. |
| both ways, decimal first | `basic-convert-round-trip` | `to_decimal(to_float(x), 2)` | x=dec:1.50 | result: dec:1.50 | |
| a conversion inside a formula | `lang-f13-collateral-coverage-ratio` | | | | `to_float` on two decimals. |
| a conversion that makes a mix work | `lang-f14-interim-payment` | | | | `to_decimal` before a decimal product. |
| a mix without conversion | `lang-problem-float-decimal-mix` | | | | Also the other order: `basic-mix-float-first` below. |
| a mix, float on the left | `basic-mix-float-first` | `a - b` | a=float:45.25, b=dec:120.50 | problems: float-decimal-mix | Either order is refused. |
| `to_float` of an integer | `basic-to-float-integer-refused` | `to_float(x)` | x=int:5 | problems: wrong-argument-type | `to_float` takes a decimal. |
| `to_decimal` of a decimal | `basic-to-decimal-decimal-refused` | `to_decimal(x, 2)` | x=dec:1.50 | problems: wrong-argument-type | `to_decimal` takes a float. |

## 7. Built-in functions: `round`

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| round, dropped digits | `basic-round-down` | `round(x, 2)` | x=dec:372.761 | result: dec:372.76 | The third place is 1, below the half. |
| round, exactly half | `basic-round-half` | `round(x, 2)` | x=dec:2.345 | result: dec:2.35 | Half up (decision 14, 20). |
| round, just under the half | `basic-round-just-under` | `round(x, 2)` | x=dec:2.344 | result: dec:2.34 | |
| round, negative half | `basic-round-negative-half` | `round(x, 2)` | x=dec:-2.345 | result: dec:-2.35 | As the language page writes it, half away from zero. Gap 5. |
| round, no places | `basic-round-no-places` | `round(x, 0)` | x=dec:2.5 | result: dec:3 | |
| round, no places, below the half | `basic-round-no-places-below` | `round(x, 0)` | x=dec:0.4 | result: dec:0 | |
| round, more places than the value has | `basic-round-pads-places` | `round(x, 4)` | x=dec:1.5 | result: dec:1.5000 | Exactly `n` places. |
| round, a float | `basic-round-float` | `round(x, 3)` | x=float:1.3629 | result: float:1.363 | A float stays a float. |
| round once at the end | `lang-f03-fund-share-per-member` | | | | |
| round with one argument | `lang-problem-wrong-argument-count` | | | | |
| round of an integer | `lang-problem-wrong-argument-type` | | | | `round` takes no integer. |
| round with a version | `lang-problem-version-not-allowed` | | | | |

## 8. Aggregates over a list

Each aggregate is run on an empty list, one item and many items. Lists are not rounded; `avg` of a decimal list has 20 places.

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| `sum`, many | `basic-sum-many` | `sum(xs)` | xs=list<dec>:[200.00, 250.00] | result: dec:450.00 | |
| `sum`, one item | `basic-sum-one` | `sum(xs)` | xs=list<dec>:[75.50] | result: dec:75.50 | |
| `sum`, empty | `basic-sum-empty` | `sum(xs)` | xs=list<dec>:[] | result: dec:0 | The value is 0. The written places are Gap 6. |
| `sum`, integers | `basic-sum-integers` | `sum(xs)` | xs=list<int>:[3, 4, 5] | result: int:12 | |
| `sum`, a list written in the formula | `basic-sum-written-list` | `sum([1, 2.50])` | | result: dec:3.50 | An integer and a decimal item widen to a decimal list (proposed). |
| `count`, many | `basic-count-many` | `count(xs)` | xs=list<dec>:[420.00, 380.50] | result: int:2 | |
| `count`, one item | `basic-count-one` | `count(xs)` | xs=list<dec>:[9.99] | result: int:1 | |
| `count`, empty | `basic-count-empty` | `count(xs)` | xs=list<dec>:[] | result: int:0 | |
| `avg`, many decimals | `basic-avg-many` | `avg(xs)` | xs=list<dec>:[10.00, 20.00, 25.00] | result: dec:18.33333333333333333333 | 55.00 / 3, 20 places. |
| `avg`, exact | `basic-avg-exact` | `avg(xs)` | xs=list<dec>:[1.00, 2.00] | result: dec:1.50000000000000000000 | |
| `avg`, one item | `basic-avg-one` | `avg(xs)` | xs=list<dec>:[5.00] | result: dec:5.00000000000000000000 | |
| `avg`, integers | `basic-avg-integers` | `avg(xs)` | xs=list<int>:[72, 85] | result: float:78.5 | The average of integers is a float. |
| `avg`, one integer | `basic-avg-integers-one` | `avg(xs)` | xs=list<int>:[70] | result: float:70 | |
| `avg`, empty | `basic-avg-empty` | `avg(xs)` | xs=list<dec>:[] | problems: empty-list | Never a value. |
| `min`, many | `basic-min-many` | `min(xs)` | xs=list<dec>:[10.00, 4.50, 7.25, 4.50] | result: dec:4.50 | |
| `min`, one item | `basic-min-one` | `min(xs)` | xs=list<dec>:[4.10] | result: dec:4.10 | |
| `min`, integers | `basic-min-integers` | `min(xs)` | xs=list<int>:[5, 3, 9] | result: int:3 | The item type. |
| `min`, a tie shows the first item | `basic-min-tie-first` | `min(xs)` | xs=list<dec>:[4.50, 4.5] | result: dec:4.50 | The first item with the least value. |
| `min`, empty | `basic-min-empty` | `min(xs)` | xs=list<dec>:[] | problems: empty-list | |
| `max`, many | `basic-max-many` | `max(xs)` | xs=list<dec>:[10.00, 4.50, 7.25] | result: dec:10.00 | |
| `max`, one item | `basic-max-one` | `max(xs)` | xs=list<dec>:[4.10] | result: dec:4.10 | |
| `max`, a tie shows the first item | `basic-max-tie-first` | `max(xs)` | xs=list<dec>:[4.5, 4.50] | result: dec:4.5 | |
| `max`, empty | `lang-problem-empty-list` | | | | `max` of an empty list. |
| `max` of a list written with a name | `lang-f04-farmer-net-payout` | | | | `max([0.00, x])`. |
| `max` and `min` together | `lang-f17-plot-yield-spread` | | | | |
| aggregates in a test | `lang-f16-lending-data-check` | | | | `count`, `avg` and `min` in one boolean. |
| `sum` of a list in a ratio | `lang-f15-capital-call-share` | | | | |

## 9. Built-in functions: dates

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| `days_between` | `lang-f05-late-days` | | | | 45 days. |
| `days_between`, backwards | `basic-days-between-backwards` | `days_between(a, b)` | a=date:2026-05-15, b=date:2026-03-31 | result: int:-45 | Negative when `b` is before `a`. |
| `days_between`, the same day | `basic-days-between-same-day` | `days_between(a, b)` | a=date:2026-03-31, b=date:2026-03-31 | result: int:0 | |
| `days_between`, a leap year | `basic-days-between-leap-year` | `days_between(a, b)` | a=date:2028-02-28, b=date:2028-03-01 | result: int:2 | 29 February exists in 2028. |
| `months_between` | `lang-f18-voting-right` | | | | 7 whole months inside it. |
| `months_between`, month end | `basic-months-between-month-end` | `months_between(a, b)` | a=date:2026-01-31, b=date:2026-02-28 | result: int:1 | `b` is the last day of its month. |
| `months_between`, a day short | `basic-months-between-day-short` | `months_between(a, b)` | a=date:2026-01-31, b=date:2026-02-27 | result: int:0 | The 27th is before the 31st and not the last day. |
| `months_between`, backwards | `basic-months-between-backwards` | `months_between(a, b)` | a=date:2026-09-15, b=date:2026-01-20 | result: int:-7 | Proposed: the same count as forwards, with a minus sign. Gap 7. |
| `years_between` | `lang-f12-straight-line-depreciation` | | | | 3 years inside it. |
| `years_between`, one day short | `basic-years-between-day-short` | `years_between(a, b)` | a=date:2023-09-10, b=date:2026-09-09 | result: int:2 | 35 months divided by 12, the rest dropped. |
| `years_between`, on the day | `basic-years-between-on-day` | `years_between(a, b)` | a=date:2023-09-10, b=date:2026-09-10 | result: int:3 | 36 months. |
| a date given where a number is needed | `basic-days-between-wrong-type` | `days_between(a, 5)` | a=date:2026-03-31 | problems: wrong-argument-type | Both arguments are dates. |
| a date inside a day count with a floor | `lang-f11-prorated-first-month-fee` | | | | `days_between` plus one. |

## 10. Calls: library functions by name and version

Each function of the first ten is called once by name and version, with one ordinary case. The call here proves the name, the version and the arguments are read. The recorded reference values, their working and their sources stay on [library-reference.md](library-reference.md), and the ordinary case below repeats the case numbered 01 there.

| Part | Case | Formula | Inputs | Expected | Reference |
|---|---|---|---|---|---|
| `fin.balance_after_payment@1` | `basic-call-balance-after-payment` | `fin.balance_after_payment@1(balance, payment)` | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | REF-fin.balance_after_payment-01 |
| `fin.equal_share@1` | `basic-call-equal-share` | `fin.equal_share@1(deposits, expenses, members)` | deposits=list<dec>:[500.00, 750.00, 1250.50], expenses=list<dec>:[120.25, 80.00], members=int:7 | result: dec:328.61 | REF-fin.equal_share-01 |
| `date.late_days@1` | `basic-call-late-days` | `date.late_days@1(due, filed)` | due=date:2026-03-31, filed=date:2026-05-15 | result: int:45 | REF-date.late_days-01 |
| `fin.flat_instalment@1` | `basic-call-flat-instalment` | `fin.flat_instalment@1(principal, monthly_rate, months)` | principal=dec:2400.00, monthly_rate=dec:0.012, months=int:7 | result: dec:371.66 | REF-fin.flat_instalment-01 |
| `fin.prorate_fee@1` | `basic-call-prorate-fee` | `fin.prorate_fee@1(monthly_fee, joined, period_end, days_in_period)` | monthly_fee=dec:47.00, joined=date:2026-06-12, period_end=date:2026-06-30, days_in_period=int:30 | result: dec:29.77 | REF-fin.prorate_fee-01 |
| `fin.straight_line_value@1` | `basic-call-straight-line-value` | `fin.straight_line_value@1(cost, salvage, life_years, bought, on)` | cost=dec:12000.00, salvage=dec:2000.00, life_years=int:7, bought=date:2023-09-10, on=date:2026-11-17 | result: dec:7714.29 | REF-fin.straight_line_value-01 |
| `fin.coverage_ratio@1` | `basic-call-coverage-ratio` | `fin.coverage_ratio@1(stock_value, facility_balance)` | stock_value=dec:84500.00, facility_balance=dec:62000.00 | result: float:1.363 | REF-fin.coverage_ratio-01 |
| `fin.proportional_share@1` | `basic-call-proportional-share` | `fin.proportional_share@1(call_amount, units, all_units)` | call_amount=dec:50000.00, units=int:1200, all_units=list<int>:[1200, 830, 1500, 477] | result: dec:14973.80 | REF-fin.proportional_share-01 |
| `stats.spread@1` | `basic-call-spread` | `stats.spread@1(values)` | values=list<dec>:[2400.50, 3100.25, 2750.00] | result: dec:699.75 | REF-stats.spread-01 |
| `math.remainder@1` | `basic-call-remainder` | `math.remainder@1(value, divisor)` | value=int:1237, divisor=int:50 | result: int:37 | REF-math.remainder-01 |
| `fin.pmt@2` | `lang-library-call-version-two` | | | | 564.88. |
| `fin.pmt@1` | `pmt-version-1-loan` | | | | In [library-versioning.md](library-versioning.md): 564. |
| a call with no version | `lang-problem-version-missing` | | | | Also `unversioned-call-refused` on the versioning page. |
| a version that does not exist | `lang-problem-version-not-found` | | | | Gap 8: the versioning page gives it the code `version-unknown`. |
| the wrong argument count | `lang-problem-wrong-argument-count` | | | | |
| the wrong argument type | `lang-problem-wrong-argument-type` | | | | |
| a name that is no function | `lang-problem-unknown-function` | | | | |
| an integer where a decimal is taken | `lang-library-call-version-two` | | | | `12000` is accepted for a decimal parameter. |

## 11. One function calling another, and the version a call pins

| Part | Case | Formula | Inputs | Expected | Rule |
|---|---|---|---|---|---|
| an agency function called once | `lang-f08-monthly-interest` | | | | `interest_step@1` inside a product. |
| an agency function called three times | `lang-f06-late-filing-penalty` | | | | `late_days@1` three times, each a call. |
| a call chain at the limit | `basic-call-chain-at-limit` | `step_d@1(amount)` | amount=dec:100.00 | accepted | A chain of 4 is accepted. The result depends on the definitions of `step_a@1` to `step_d@1`, which the language page lists but does not define. Gap 9. |
| a call chain over the limit | `lang-problem-call-too-deep` | | | | A chain of 5 is refused. |
| two versions of one function in one formula | `same-host-newer-minus-older` | | | | In [library-versioning.md](library-versioning.md): `0.88`. |
| the call keeps the version it names | `old-version-still-served` | | | | In [library-versioning.md](library-versioning.md). |
| a new version is published, the call keeps its pin | `basic-pin-after-new-version` | `fin.pmt@1(principal, rate, months)` | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:564 | Run once with `fin.pmt@1` alone served and once with `fin.pmt@2` published too; both runs give `564`. |

## 12. The worked case: balance after a payment

The first design of the engine wrote a loan balance this way, with the whole number `0` in the then branch:

```formula
payment >= balance ? 0 : balance - payment
```

It is saved in this standard text. The number `0` is an integer and `balance - payment` is a decimal, so the two answers have the common type decimal, and the result is a decimal. A decimal result compares by value, so `0`, `0.0` and `0.00` are the same amount. The same formula with `0.00` is formula 1 of [formula-examples.md](formula-examples.md). Every case below reads two inputs, `balance` and `payment`, that are required, have no negatives and have 2 places (the definition V1).

| Part | Case | Inputs | Expected | Rule |
|---|---|---|---|---|
| payment below the balance | `basic-loan-balance-payment-short` | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | 120.50 - 45.25. |
| payment above the balance | `basic-loan-balance-payment-covers` | balance=dec:120.50, payment=dec:130.00 | result: dec:0 | The payment covers it, so the balance is zero. |
| payment equal to the balance | `basic-loan-balance-payment-equal` | balance=dec:120.50, payment=dec:120.50 | result: dec:0 | `>=` makes an equal payment clear the balance. |
| nothing owed and nothing paid | `basic-loan-balance-zero-zero` | balance=dec:0.00, payment=dec:0.00 | result: dec:0 | 0.00 >= 0.00. |
| nothing paid | `basic-loan-balance-nothing-paid` | balance=dec:80.00, payment=dec:0.00 | result: dec:80.00 | |
| one cent short | `basic-loan-balance-one-cent` | balance=dec:100.00, payment=dec:99.99 | result: dec:0.01 | |
| the formula as typed with no spaces | `lang-save-tight-spacing` | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | Saved with one space around each operator. |
| a negative payment | `rules-no-negatives-invalid` | balance=text:"80.00", payment=text:"-5.00" | problems: no-negatives | Refused before the formula runs (decision 22). |
| a missing balance | `rules-required-missing` | payment=text:"45.25" with no balance | problems: required | |
| three places | `rules-decimal-places-invalid` | balance=text:"60.00", payment=text:"5.125" | problems: decimal-places | |
| a balance allowed to be negative, payment covers it | `basic-loan-balance-negative-allowed-covers` | balance=dec:-10.00, payment=dec:5.00 | result: dec:0 | Needs the definition V1n: `balance` declares negatives allowed. |
| both allowed to be negative, payment lower | `basic-loan-balance-negative-allowed-lower` | balance=dec:-10.00, payment=dec:-20.00 | result: dec:10.00 | Needs V1n with `payment` also declaring negatives allowed. -20.00 >= -10.00 is false, so -10.00 - (-20.00). |

## Held cases

These cases need a setting whose shape the owner has not decided. They are written so they are not forgotten, and no loader reads them yet.

| Case | Setting | Formula | Inputs | Expected |
|---|---|---|---|---|
| `basic-held-division-five-digits` | the agency keeps 5 digits inside a division (decision 21; the range is open) | `a / b` | a=dec:2.00, b=int:3 | result: dec:0.66667 |
| `basic-held-division-default` | the default of 20 digits | `a / b` | a=dec:2.00, b=int:3 | result: dec:0.66666666666666666667 |
| `basic-held-rounding-half-down` | the agency chooses half down (decision 20) | `round(x, 2)` | x=dec:2.345 | result: dec:2.34 |
| `basic-held-rounding-half-even` | the agency chooses half even | `round(x, 2)` | x=dec:2.345 | result: dec:2.34 |
| `basic-held-rounding-up` | the agency chooses up | `round(x, 2)` | x=dec:2.341 | result: dec:2.35 |
| `basic-held-rounding-down` | the agency chooses down | `round(x, 2)` | x=dec:2.349 | result: dec:2.34 |

How a formula names a rounding rule is not on the language page, so the formula text of the last four rows is a placeholder. The exact-money page holds the full cases for every rule.

## Gaps

A part the language page names or the issue asks for that has no case. Each is open for the owner.

1. A negative written with a minus sign. The language page leaves out the unary minus and a signed literal, and does not give the words of the refusal of `balance * -5`. The case waits for those words.
2. A power operator. The language has none. Cases refuse `^` and `pow`; there is nothing to run.
3. A date as a result. The page says a date comes only from an input and goes only to `days_between`, `months_between` and `years_between`. Whether a formula may return a date, for example `due` alone, is not said.
4. A list as a result. A list can be written in a formula, but no row of formula-examples.md expects a list, and the page does not say whether `[a, b]` may be a whole formula.
5. A half on a negative value. The language page rounds half away from zero. Decision 14 and 20 name half up, and a list with down, up, half up, half down and half even. For a positive value half up and half away from zero agree. For a negative half they differ if "up" means toward positive. `basic-round-negative-half` follows the language page until the owner names the rule.
6. The written places of the sum of an empty list. The page says the sum of an empty list is `0`. It does not say whether a decimal list gives `0` or `0.00`.
7. The sign of `months_between` and `years_between` going backwards. The page says negative when `b` is before `a`, and does not say it is the same count as forwards.
8. The code for a version that does not exist. The language page says `version-not-found`. [library-versioning.md](library-versioning.md) says `version-unknown`.
9. The functions `step_a@1` to `step_e@1`. The language page lists their kinds and chain depth, and no definition, so a chain at the limit cannot be given a result.
10. A `round` or `to_decimal` with `n` outside 0 to 10, or `n` given as an input name instead of written in the formula. The page says `n` is an integer written in the formula, from 0 to 10, and gives no code for a call that breaks it.
11. A decimal literal or an integer at its digit limit as a result. The limits are on [hostile-formulas.md](hostile-formulas.md).
12. A rounding rule other than half away from zero inside a formula. Decision 20 makes it an agency choice. The language page has no way to name one. See "Held cases".

## Open questions for the owner

1. Does the case loader read the `basic-` rows on this page, or are they moved into formula-examples.md with the rest? Proposed: moved, so there is one table (formula-examples.md has the loader's columns).
2. Is the `1e200` notation the proposed way to write a very large float?
3. Is the first-design `0` (a whole number in the then branch) kept as the worked case, or only `0.00`? Proposed: kept, because it shows the common type.

## How the cases were worked

Every figure was worked by hand from the rules of the language page, then read again against the rules for its type: the places of a sum and a product, the 20 places of a division, half up at the 20th place, and the first item of a tie. Nothing was run, because the engine is not chosen.
