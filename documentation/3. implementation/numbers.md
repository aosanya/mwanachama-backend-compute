# The numbers package

The package `number/` holds the figures a formula works on: an exact decimal, a float, the two ways to convert between them, the rounding rules, the digits kept in a division, and the split of a total. It imports only the Go standard library and knows nothing of formulas, rows or agencies. A host or the engine passes in the settings an agency chose.

Status: built for issue ISS-23 and awaiting the owner's review. Every figure marked proposed is a starting point, not a decision. All names and figures in the tests are invented.

Built from [exact-money.md](../4.%20qa/exact-money.md) and from decisions 13, 14, 15, 20 and 21 in [decisions.md](../1.%20requirements/decisions.md). The kinds and their limits are on [formula-language.md](../2.%20design/formula-language.md).

## What is in it

| Name | What it is |
|---|---|
| `Decimal` | An exact decimal: a whole number of units and a count of places. It keeps its written digits, so `0.010` has 3 places and equals `0.01`. It has no limit of its own on size; the limits are the engine's (below). |
| `Float` | An approximate 64-bit number that is always finite. |
| `Value` | A decimal or a float, for the operations that must refuse a mix. |
| `Rounding` | The rule for the fraction: `RoundHalfUp` (the default), `RoundHalfDown`, `RoundHalfEven`, `RoundUp`, `RoundDown`, and the candidates `RoundHalfPosInf`, `RoundCeiling`, `RoundFloor`. |
| `Settings` | The places of the currency, the rounding rule and the digits kept in a division. `DefaultSettings()` gives 2 places, half up and 20 digits. |
| `Split`, `EqualSplit` | Share a total by weights, or equally, and report the left-over. |

## Decimal

| Call | Result |
|---|---|
| `ParseDecimal(text)` | A decimal, or an error. The text is an optional `-`, whole digits and optionally a dot with one or more digits. `.5`, `5.`, `+5`, `1e3`, `1,5`, spaces, an empty text and a leading zero such as `007` are refused. Counting digits and the 30 digit limit belong to Check and Validate, not here. |
| `NewDecimal(units, places)`, `DecimalFromInt(n)` | A decimal from whole units. `NewDecimal(1999, 2)` is `19.99`. |
| `Add`, `Sub`, `Mul`, `Neg` | Exact. A sum or difference has the larger number of places; a product has the sum of the places. Nothing is rounded and nothing wraps or saturates. |
| `Cmp`, `Equal` | Compare values, not digits: `12.5` equals `12.50`. |
| `Round(places, rule)` | A decimal with exactly that many places. More places pad with zeros. |
| `Div(divisor, digits, rule)` | A decimal with exactly `digits` places, the last rounded by the rule. Dividing by zero is `ErrDivideByZero`. |
| `String()` | The digits as kept. Zero never has a sign: `-0.004` rounded to 2 places prints `0.00`. |
| `Digits()` | The number of digits of the whole figure, for the engine's size limit. |
| `ToFloat()` | The explicit conversion to a float. A decimal too big for a float is `ErrNotFinite`. |

## Float

`NewFloat` refuses `NaN` and the infinities with `ErrNotFinite`; `Add`, `Sub`, `Mul` and `Div` refuse a result that is not finite, and `Div` refuses a zero divisor. A float is never read from a program's number type by a host: the library reads digits as text (the language page, Validate).

## Conversion and mixing

A decimal and a float never meet by themselves.

- Inside Go the two are different types, so one cannot be passed for the other.
- On `Value`, `Add`, `Sub`, `Mul` and `Compare` refuse a decimal with a float with a `*KindMixError`. Its sentence names both kinds and says to convert one: `A decimal and a float cannot be combined; convert one with to_float or to_decimal.` The wording is proposed (exact-money.md open question 6).
- `Float.ToDecimal(places, rule, reading)` is the only way from a float to a decimal. It states its places; a negative number of places is refused.
- `Decimal.ToFloat()` is the only way back.

`FloatReading` chooses what a float means on conversion. `ReadShortest` reads the shortest text that gives the float back (`0.1` is `0.1`); `ReadExact` reads the float's exact binary value (`0.1` is `0.1000000000000000055...`). The two differ at the half: float `1.005` to 2 places half up is `1.01` read shortest and `1.00` read exact, and float `0.1` to 2 places under up is `0.10` read shortest and `0.11` read exact. The package offers both; which one the language uses is open (below). `ReadShortest` is the zero value and the one the tests use for the ordinary cases.

## Rounding

A rule names what happens at the half or at any fraction.

| Rule | 2.5 | -2.5 | 2.4 | Status |
|---|---|---|---|---|
| `RoundHalfUp` | 3 | -3 | 2 | default (decision 14); at the half, away from zero |
| `RoundHalfDown` | 2 | -2 | 2 | proposed list |
| `RoundHalfEven` | 2 | -2 | 2 | proposed list |
| `RoundUp` | 3 | -3 | 3 | proposed list; any fraction, away from zero |
| `RoundDown` | 2 | -2 | 2 | proposed list; any fraction, toward zero |
| `RoundHalfPosInf` | 3 | -2 | 2 | candidate only |
| `RoundCeiling` | 3 | -2 | 3 | candidate only |
| `RoundFloor` | 2 | -3 | 2 | candidate only |

The proposed list is down, up, half up, half down and half even (decision 20). The three candidates are in the package so the tests can run all eight; they are dropped from the package if the owner confirms the list. Rounding is exact: `1.005` and `1.015` are decimals, so the float trap of exact-money.md group B does not arise.

## Division

`Div` keeps exactly the digits asked and rounds the last one by the rule, so `1 / 3` to 20 digits is `0.33333333333333333333` and times 3 is `0.99999999999999999999`. There is no refusal and no rational number (decision 21). `Settings.Divide` uses the agency's `DivisionDigits` and `Mode`.

## Settings

| Setting | Meaning | Default | Status |
|---|---|---|---|
| `Places` | Places of the agency's currency (decision 15). | 2 | the default is the usual currency; the agency sets it |
| `Mode` | The rounding rule (decisions 14 and 20). | half up | half up is decided as the default; the list is proposed |
| `DivisionDigits` | Digits kept in a division before the final rounding (decision 21). | 20 | proposed; the range is open |

`Settings.Round(d)` is `d.Round(Places, Mode)`. No setting is read from anywhere: the caller passes them in, so the package makes no choice for an agency.

## Splitting a total

`Split(total, weights, places)` shares a total. Each share is the total times its weight over the sum of the weights, cut down to the smallest unit of `places` toward zero. The left-over is the total minus the sum of the shares, so shares plus left-over always add back to the total. Decision 13: nobody gets an extra unit, there is no rounding account and no random choice, the payout is the sum of the shares and the left-over stays in the account the payout came from. `EqualSplit` is `Split` with equal weights. Weights are whole numbers; a negative weight, no weights, or weights that all are zero are refused.

For `-100.00` among 3 the shares are `-33.33` each and the left-over is `-0.01`: the cut is toward zero. That direction for a negative total is not decided (below).

## What is not here

- Size limits. The engine refuses a decimal result of more than 60 digits, an integer result of more than 18 and a literal of more than 30 digits (language page, `number-too-big` and `too-many-digits`). `Decimal.Digits()` gives it the count. The cases on the boundary B of the fast path (exact-money.md EM-70 to EM-80) are not written: this package keeps every figure as a big integer, and B belongs to the engine choice.
- The reading of a number from text with spaces, commas or a percent sign, and the input rules (Validate).
- The formula language, the library functions and batch runs.

## Open questions

These stay as the owner left them. The package takes them as settings or as a choice so none is hard coded.

1. The list of rounding rules (proposed: down, up, half up, half down, half even), and whether an agency can change its own default from half up.
2. The digits kept in a division: the 20 digit default is proposed and the range an agency may choose from is open.
3. The direction of the cut for a negative total in `Split`. The package cuts toward zero; cutting toward negative infinity would give `-33.34` and a left-over of `+0.02` for `-100.00` among 3.
4. How a float is read on conversion, `ReadShortest` or `ReadExact` (exact-money.md open question 8).
5. The exact sentence for mixing a decimal with a float (open question 6).
6. Places other than 2: the tests run `Split` for 0 and 3 places and `Settings` for 0 and 4, but the full set of cases is not repeated per place count (open question 3).
7. Negative inputs: whether a function may take one is declared per input (decision 22, open against the language page); the package itself handles negative figures.
