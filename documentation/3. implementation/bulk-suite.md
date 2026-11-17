# Bulk run suite

Folder `bulk_test/` (Go package `bulkruns`), built for ISS-39. It runs the cases of [bulk-runs.md](../4.%20qa/bulk-runs.md) against the batch package ([batch.md](batch.md)). Every case is checked the way section 10 of that page and [batch-equals-single.md](../4.%20qa/batch-equals-single.md) ask, on one row shape. The decisions are 9, 16, 17 and 19 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. All figures are invented.

## What it holds

| File | What it is |
|---|---|
| `package.go` | the package clause |
| `helpers_test.go` | the three formulas M, S and F, the rule R1 rows and answers, the answer notation and the shared checks |
| `probe_test.go` | a small runtime function `probe@1` that counts the rows worked and can hold one row until others finish |
| `sizes_test.go` | BR-01 to BR-11 and the defaults |
| `order_test.go` | BR-12 and BR-13 |
| `badrows_test.go` | BR-20 to BR-25, BR-30, BR-31, BR-40, BR-41 and the chunk and worker sweep |
| `pushed_test.go` | BR-50 to BR-57 with a stand-in database of exact fractions |
| `limits_test.go` | BR-60 to BR-64, BR-65 to BR-70 and BR-71 (the group B hostile rows B01 to B05) |
| `held_test.go` | the held cases, each skipped with the question that holds it |

It imports `batch/`, `formula/`, `function/`, `number/`, `numtext/` and `problem/` and nothing else from this repo. It changes none of them.

## How a case is checked

`runAndCheck` runs the batch once and then checks the four things the QA page lists:

1. No error, one result per row, and `Rows[i].Row` equals i+1.
2. Every row equals its written answer: the type and exact text of a value, or the stage and the problems in order, each with its code, its input and the value received (or that it was not given), and a sentence that is not empty.
3. The summary matches the results: `Valued`, `Stopped` and `NotRun` equal the counts read from the rows and add up to `Rows`, and `ProblemsByCode` equals the count of each code. The cases that name counts check them too.
4. Every row equals the same row run alone with `Run` on the shared inputs merged under the row's own values. Value, problems and stage are compared whole, and for the 10,000-row cases all 10,000 rows are compared. A row that was never started is checked to hold no value and exactly one `batch-limit-reached` or `batch-cancelled` problem.

Rows are described by rule where the QA page does so: R1 (balance 10.00 x i, rate 0.005, answer 0.05 x i) for the large cases, and a mod 3 rule for BR-31. The sums of answers are worked with `number.Decimal` and compared with the page's figures.

## Proofs that the checks bite

Made in scratch copies of the folder, never in the committed files:

- Moving every R1 answer by one cent made the suite fail in many places, among them BR-02 to BR-09, BR-13, BR-21 to BR-23, BR-65 to BR-69 and the row limit case. Changing one expected answer in BR-20 failed with `row 2 gave 0.10, want 0.11`.
- Comparing each row with the wrong row in the equals-single check failed at row 1 of the first case that has a bad row.

## Results

`go vet ./bulk_test/` is clean and `go test ./bulk_test/ -race -count=1` passes: 120 tests and subtests pass and 9 are held. No engine failure was found, so no issue was filed against `batch/` and no expectation was changed.

## What the suite does without a hook

The QA page asks for test-only hooks in the library (open question 7). The batch package has them, but they are not reachable from another folder. The suite gets what it needs from a runtime function instead:

- BR-12 holds the row 80.00 until six other rows have finished, with 2 rows to a chunk and 4 workers, so the first chunk finishes last. The test fails if the hold was not released, so it cannot pass without the order being tested.
- BR-60 to BR-62 call a formula that works a row through `probe@1` and check that no row was worked on a refused call.
- The cost of one row (c) is read from a single run of the formula on `formula.Plan` and `RunMetered`, and the test checks that it is the same at row 1 and row 450.

The formula of BR-12 and of the refusal cases is `round(probe@1(balance) * rate, 2)`, formula M with one call added, so these two groups run a variant of M.

## Held cases

Each is a skipped test that names its question.

| Cases | Why | Question |
|---|---|---|
| BR-72 to BR-77 | Whether and how a host's own cancel or deadline is honoured is not decided. `batch/` builds the design page's proposal; this suite builds nothing on it. | Decision 19; open questions 1 and 8 of bulk-runs.md |
| BR-80 | Counting the rows a worker holds needs a hook inside `batch/`. | Open question 7 |
| BR-81, BR-82, BR-83, BR-84 | The memory figures (32 MiB, 1 MiB over one call and over 100 calls) are proposed and wait for the engine choice figures. | Open question 10 |
| BR-85 | Follows BR-73. | Open questions 1 and 8 |
| A smaller agency row limit than the ceiling | The ceiling above which no agency may set its rows per batch is open. The row limit as a setting is tested (50 and 51 rows against a limit of 50). | Open question 11 |
| `Aggregates()` of a program | The language has no declaration syntax for a pushed-down input, so BR-50 to BR-57 pass the total in as a shared input. | batch.md, not built 1 |

BR-57 runs on decision 13: the shares are cut down (the rounding mode of the compiled formula is `RoundDown`), 7 shares of 14.28 add to 99.96 and the left-over 0.04 stays with the payout. A run with the default half-up mode is not asserted, because the split rule belongs to `number.Split`, not to a formula's `round`.

## What stays open

1. The held cases above.
2. BR-65 to BR-70 are marked provisional on the QA page until the engine limits are fixed. They run on `Options.Cost`, which `batch/` builds with the proposal that a chunk starts only while the cost spent is below the limit (open question 5). If "reached" is to mean more than the limit, BR-65 and BR-68 change.
3. The chunk counts of the sizes table (1, 2, 3 and 100 chunks and so on) cannot be seen from outside the package. The suite checks the chunk size and workers that the summary reports and that every answer is the same for every chunk size and worker count, and leaves the counting of chunks to `batch/`'s own tests.
4. BR-71 reads group B of [hostile-formulas.md](../4.%20qa/hostile-formulas.md) as the batch page does. The hostile input is a list of 1,001 items (refused at validate) or a list of 1,000 items over a cost limit of 500 (stopped at calculate with the sentence for that limit). The full hostile page is ISS-34's.
5. The code names of the three refused calls, the Limit and Got of BR-63 and the choice of error in BR-64 are the QA page's proposals (open question 6), and are built as proposed.
6. Decision 22 against the language page's `no-negatives` rule: the suite declares `no-negatives` on every money input of M, S and F, as the QA page does, so it does not depend on which wins.
