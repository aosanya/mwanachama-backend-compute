# The money suite

The Go tests of [exact-money.md](../4.%20qa/exact-money.md), in the folder `exactmoney_test/`. Built for ISS-33. They run the cases the owner's decisions settle, on the numbers package ([numbers.md](numbers.md)) and on the run of a formula ([formula-run.md](formula-run.md)). Nothing here changes `number/`, `formula/` or the QA page.

Status: the owner has not reviewed this page. All names and figures are invented.

## How to run it

```
go test ./exactmoney_test/ -race
```

The folder holds only test files, in the package `exactmoney_test`, so it is not a library and `go build ./...` has nothing to build in it. `go test ./...` runs it with the rest.

## What the files hold

| File | Cases | What it checks |
|---|---|---|
| `drift_test.go` | EM-01 to EM-04 | exact sums: 0.1 + 0.2, a thousand cents, five amounts and a thousand amounts in several orders |
| `rounding_test.go` | EM-10 to EM-27, EM-101 | every row of the B1 and B2 tables under all eight modes, and zero printed without a sign |
| `split_test.go` | EM-30 to EM-45, EM-50, EM-92, EM-102 | the ten split cases, the six invariants, the left-over staying with the payout |
| `interest_test.go` | EM-52 to EM-56, EM-93 | simple and compound interest, rounding each period against once at the end |
| `mixing_test.go` | EM-60 to EM-66 | a decimal and a float are refused, and the explicit conversion states its places |
| `overflow_test.go` | EM-81 to EM-83, EM-96 | figures past 64 bits against an independent `math/big` count |
| `extremes_test.go` | EM-90 to EM-100 | zero, the smallest unit, a half cent, a division that does not end, four places |
| `formula_test.go` | the same cases through the language | the figures worked out by a formula, not only by the package |
| `generated_test.go` | none of the page's ids | 500 generated splits and 500 generated sums from a fixed seed |
| `helpers_test.go` | | the eight modes, the place counts, and the small helpers |

## Repeated for each place count

Decision 15 makes the places an agency setting, so the page's cases written for 2 places are run for 0, 1, 2, 3 and 4.

- **Rounding rows.** Each row is stored as whole units: the input has one more digit than the places and the expected figure is the same digits with that digit rounded off. `1.005` to 2 places is `1005` at 3 places to `101` at 2, and the same row is run as `1005` at 1 place to `101` at 0, and so on. The expectation is the same whole number at every place count, because dropping the last digit does not depend on where the point is.
- **Split rows.** The total and the shares are whole units. `10000` units among 3 gives `3333` units each and `1` over at every place count.

The cases that name a figure (`1157.625`, `530.6817753125`) are run at 2 places only, since their figures are written with the point.

## Two layers

Each settled case is run on the package that holds the figure (`number/`), and where the language can say it, again through `formula.Plan` and `Run`. The formula layer sets the rounding rule and the digits through `Options.Numbers`, so the same formula text gives the figure of every mode.

| Layer | Cases |
|---|---|
| `number/` | all of them |
| the run of a formula | EM-01, 03, 04 (sums of a list), the rounding tables through `round(x, n)` under all eight modes, EM-27, EM-52, EM-54 to EM-56, EM-60 to EM-63, EM-64, EM-66, EM-90, EM-94, EM-95, EM-97, EM-98, EM-100 |

## Things the language does that the page does not say

Found while writing the formula layer. None is a failure of the engine. Each is a fact of the language page, listed here so the QA page can be corrected by the owner.

1. A formula has no unary minus. A negative amount is written `0.00 - 10.00` or arrives as an input, so EM-100 is run as `0.00 - 10.00 + 4.00`.
2. A decimal literal has at most 10 places, so EM-97's `0.99999999999999999999` (20 places) cannot be written in a formula. The check reads the product in Go instead.
3. There is no float literal, so EM-60 to EM-62 take the float as an input named `rate` and Check refuses the mix before any run. Check, not the run, gives the refusal.
4. `to_decimal` with one argument is refused by Check, so EM-66 is a refusal at save, as the page says.

## What the suite found

Nothing failed. Every expected figure of the page held on `number/` and on the formula run, so no issue was filed against ISS-23 or ISS-25.

The suite does fail when a figure is wrong. In a scratch copy, changing the half-even expectation of EM-20 from `100` to `101` fails the rounding table at every place count, with the id, the input, the mode and both figures in the message. The committed tables were not touched.

A failing generated case names its seed and its index, so it replays: `seed 20261117, split 31: total ... units at ... places, weights [...]`. The generator is `generate(count)` in `generated_test.go`; the same seed gives the same cases.

## Held, with the question that holds each

| Cases | Why not run | Question |
|---|---|---|
| EM-46 to EM-49, EM-51 | ruled out by decision 13 (the left-over stays in the payout's account; no extra unit, no random choice) | none: closed |
| EM-70 to EM-80 | they are relative to the boundary B of the fast path, and B is not set. `number/` keeps every figure as a big integer, so there is no second path to cross | exact-money.md open question 7, the engine choice |
| EM-65 on the formula run | the run reads a float by its shortest text, so `0.1` under `up` gives `0.10`; the page's `0.11` is the exact reading. Both readings are tested on `number/` | open question 8 |
| the size limits of the language (60 digits, 18-digit integers, 30-digit literals) | they belong to Check and Validate, not to the money figures | the language page |
| negative cases on the formula run | run, but only meaningful if negative inputs are allowed (decision 22, open against the language page); they are marked in their names | open question 5 |

## Open for the owner

These stay as the owner left them. The suite pins the behaviour of today where the page says "assumed".

1. The direction of the cut for a negative total (EM-102): the package cuts toward zero and the test says `-33.33` each with `-0.01` over. Cutting toward negative infinity would give `-33.34` and `+0.02`.
2. The rounding list. The three candidate modes (half-+inf, ceiling, floor) are run with the other five. If the owner confirms the proposed list, delete their columns from `rounding_test.go` and the matching rows of the tables in `interest_test.go`.
3. The sentence of the refusal for mixing a decimal and a float. The test checks that it names both kinds and is one sentence, not a fixed string.
4. The places other than 2 for the interest and the division cases, which are written with the point and run once.

## Not in the suite

- The batch cases and the fuzz run: ISS-34 and ISS-39.
- The money figures of the posting step and the ledger, which are the accounting module's and carry their own tests.
