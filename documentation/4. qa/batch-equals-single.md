# Batch equals single

Status: the owner has not yet reviewed this page. Every number marked "proposed" is an invented starting point. A sentence that starts with "Provisional:" depends on the formula language page (`documentation/2. design/formula-language.md`, not written yet), on the shape of a Validate problem, on the limits or on the engine; the marker is dropped or the sentence corrected when those are settled.

All figures on this page are invented. The repo is public.

## Purpose

[Batch evaluation](../2.%20design/batch-evaluation.md) promises that a batch result equals running each row alone, and that a test proves it on invented data. This page writes that test down: the Go surface it calls, the steps it takes, the cases it loads from the case table, the cases it generates itself, and what the later bulk-run cases may rely on. It extends law 4 on [laws](laws.md).

The owner decided the promise (decision 8 on the design page). Everything else here is a proposal or an open question and says so.

## The Go surface

These are the signatures the test calls, signatures only.

```go
type Inputs map[string]Value

type Problem struct {
	Input    string
	Given    bool
	Received string
	Rule     string
	Sentence string
	Code     string
}

type StoppedAt string

const (
	StoppedAtValidate  StoppedAt = "validate"
	StoppedAtCalculate StoppedAt = "calculate"
	StoppedAtBatch     StoppedAt = "batch"
)

type RowResult struct {
	Row       int
	Value     Value
	Problems  []Problem
	StoppedAt StoppedAt
}

type BatchOptions struct {
	ChunkSize int
	Workers   int
	Shared    Inputs
}

type BatchSummary struct {
	FormulaHash    string
	Rows           int
	Valued         int
	Stopped        int
	NotRun         int
	ProblemsByCode map[string]int
	ChunkSize      int
	Workers        int
	Duration       time.Duration
}

type BatchResult struct {
	Rows    []RowResult
	Summary BatchSummary
}

type BatchError struct {
	Code  string
	Limit int
	Got   int
}

type Aggregate struct {
	Input    string
	Function string
	Measure  string
}

func Run(ctx context.Context, program *Program, inputs Inputs) RowResult
func RunBatch(ctx context.Context, program *Program, rows []Inputs, options BatchOptions) (BatchResult, error)
func (e *BatchError) Error() string
func (p *Program) Aggregates() []Aggregate
```

Provisional: `Value`, `Program`, `Run` and the shape of `Problem` depend on the formula language page. `Parse`, `Format`, `Check` and `Validate` keep their own names from [laws](laws.md); this page only keeps the naming consistent with them and defines no signature for them.

What each part means:

| Part | Meaning |
|---|---|
| `Inputs` | The named values of one row |
| `Problem.Input` | The name of the input the problem is about |
| `Problem.Given` | Whether the input was given at all |
| `Problem.Received` | The value received, as exact text |
| `Problem.Rule` | The rule that was broken |
| `Problem.Sentence` | A plain sentence for a person to read |
| `Problem.Code` | The short code of the problem, such as `no-negatives` |
| `StoppedAt` | The stage that stopped a row: validate, calculate, or batch (never started) |
| `RowResult.Row` | The row number, 1-based: the position in the list plus 1 |
| `RowResult.Value` | The value, when the row has one |
| `RowResult.Problems` | The problems, when the row has any; a row has a value or a non-empty list, never both and never neither |
| `BatchOptions.ChunkSize` | Rows per chunk; 0 means the default 100 |
| `BatchOptions.Workers` | Number of workers; 0 means the default 1 |
| `BatchOptions.Shared` | Inputs applied to every row; the row's own value wins when a name is in both |
| `BatchSummary.FormulaHash` | The hash of the formula that was run |
| `BatchSummary.Rows` | The number of rows sent |
| `BatchSummary.Valued` | Rows that came back with a value |
| `BatchSummary.Stopped` | Rows stopped at validate or at calculate |
| `BatchSummary.NotRun` | Rows never started, stage batch |
| `BatchSummary.ProblemsByCode` | A count of problems per code |
| `BatchSummary.ChunkSize`, `Workers` | The values actually used, after defaults |
| `BatchSummary.Duration` | How long the call took |
| `BatchResult.Rows` | One result per input row, in the order sent |
| `BatchError.Code` | `too-many-rows`, `bad-chunk-size` or `bad-workers` |
| `BatchError.Limit` | The limit that was broken |
| `BatchError.Got` | The number received |
| `Aggregate` | One pushed-down input: its name, the function (sum, count, min, max) and the measure the host maps to a column |
| `(*Program).Aggregates` | The aggregates the program declares, so the host knows what to ask its database |

Valued + Stopped + NotRun = Rows. The three groups do not overlap.

## The test

For each case, from the case table or generated:

1. Build the rows, as a list of `Inputs`, and the options.
2. Run `RunBatch` once and keep the result.
3. Check the shape: no error, `len(Rows)` equals the number of input rows, and `Rows[i].Row` equals `i+1` for every `i`.
4. For each row, merge the shared inputs under the row's own values and run `Run` alone on the merged row.
5. Compare the batch result of that row with the single result:

| Compared | How |
|---|---|
| Value | Type and exact text, so 12.5 differs from 12.50 |
| Problems | Same count, same order, and per problem the same input, received value, rule, sentence and code |
| Stage | The same `StoppedAt` |

6. A row that carries the code `batch-limit-reached` or `batch-cancelled` is exempt from step 5, because a single run never meets a batch limit. Every other row is compared.
7. Duration and the summary are not compared with the single run.
8. Check the summary against the results: `Rows` equals the number of rows; `Valued`, `Stopped` and `NotRun` equal the counts found by reading the results; the three add up to `Rows`; `ProblemsByCode` equals the count of each code over all problems; `ChunkSize` and `Workers` equal the values used after defaults.
9. For a case from the case table, also compare each row with the `expected` cell of its table row, so the test fails when batch and single agree on a wrong answer.

A failure is reported in the style of [laws](laws.md):

| Failure | Test reports |
|---|---|
| A row differs | `batch differs from single, row 3: alone gave 3, in the batch gave a division problem` |
| Wrong count | `batch row count differs: sent 5 rows, got 4 results` |
| Wrong row number | `row number wrong: result at position 3 carries Row 5` |
| Summary differs | `summary differs from results: Valued says 3, results show 2` |
| Differs from the table | `batch row 2 differs from case balance-bad-row-r2: expected problems no-negatives, got 0.00` |

A generated case adds its seed to the line, as the laws do: `batch differs from single, seed 5530, row 3: ...`.

## How a batch case is written in the case table

Proposal, not decided by the owner. The six-column shape on [example cases](example-cases.md) has one row of inputs per case, and its open question 3 leaves list syntax to the owner. A batch needs many rows of inputs for one formula. Proposed way to write one without a new syntax:

- A batch case is consecutive table rows named `<case>-r1`, `<case>-r2`, up to `<case>-rN`. Names stay lowercase letters, digits and hyphens.
- The `formula` cell is the same on every row of the case.
- Each row's `inputs` cell holds its full inputs. Shared inputs are written out on each row.
- Each row's `expected` cell holds the expected answer of that row.
- The loader groups the rows of a case in order of N into one `RunBatch` call, and also runs each row alone with `Run`. So one table feeds both sides of the comparison.
- A row that is not given an input has that input left out of its `inputs` cell.

The `page` column of every row below is `documentation/2. design/batch-evaluation.md`. The case file name is open (open question 7 on the design page).

Provisional: the problem codes below (`required`, `no-negatives`, `decimal-places`, `divide-by-zero`) belong to the formula language page, which owns the list. The case table holds the code only; the stage (validate or calculate) is checked against the single run, not against the table.

Provisional: the formula syntax shown is the one used on the design page, and stays illustrative until the language page fixes it.

### (a) All rows good

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| share-all-good-r1 | round(total / people, 2) | total=dec:100.00, people=int:3 | result: dec:33.33 | A batch of good rows returns one value per row. | documentation/2. design/batch-evaluation.md |
| share-all-good-r2 | round(total / people, 2) | total=dec:90.00, people=int:4 | result: dec:22.50 | A batch of good rows returns one value per row. | documentation/2. design/batch-evaluation.md |
| share-all-good-r3 | round(total / people, 2) | total=dec:50.00, people=int:2 | result: dec:25.00 | A batch of good rows returns one value per row. | documentation/2. design/batch-evaluation.md |

### (b) One bad row at position 2 of 3

Formula 1 of `documentation/1. requirements/example-formulas.md`.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| balance-bad-row-r1 | payment >= balance ? 0.00 : balance - payment | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | A bad row never spoils the good rows around it. | documentation/2. design/batch-evaluation.md |
| balance-bad-row-r2 | payment >= balance ? 0.00 : balance - payment | balance=dec:80.00, payment=dec:-5.00 | problems: no-negatives | A bad row never spoils the good rows around it. | documentation/2. design/batch-evaluation.md |
| balance-bad-row-r3 | payment >= balance ? 0.00 : balance - payment | balance=dec:60.00, payment=dec:70.00 | result: dec:0.00 | A bad row never spoils the good rows around it. | documentation/2. design/batch-evaluation.md |

### (c) Example A of the design page, 5 rows

Rows 1 and 3 have values, rows 2, 4 and 5 stop at validate, 4 problems in all. Row 4 leaves `payment` out because it was not given.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| month-end-r1 | payment >= balance ? 0.00 : balance - payment | balance=dec:120.50, payment=dec:45.25 | result: dec:75.25 | Five rows come back in the order sent. | documentation/2. design/batch-evaluation.md |
| month-end-r2 | payment >= balance ? 0.00 : balance - payment | balance=dec:80.00, payment=dec:-5.00 | problems: no-negatives | Five rows come back in the order sent. | documentation/2. design/batch-evaluation.md |
| month-end-r3 | payment >= balance ? 0.00 : balance - payment | balance=dec:60.00, payment=dec:70.00 | result: dec:0.00 | Five rows come back in the order sent. | documentation/2. design/batch-evaluation.md |
| month-end-r4 | payment >= balance ? 0.00 : balance - payment | balance=dec:200.00 | problems: required | Five rows come back in the order sent. | documentation/2. design/batch-evaluation.md |
| month-end-r5 | payment >= balance ? 0.00 : balance - payment | balance=dec:-1.00, payment=dec:5.125 | problems: no-negatives, decimal-places | Five rows come back in the order sent. | documentation/2. design/batch-evaluation.md |

### (d) A failure at calculate, example B

Row 2 passes validation and stops at calculate; rows 1 and 3 are untouched.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| share-calc-fail-r1 | round(total / people, 2) | total=dec:100.00, people=int:3 | result: dec:33.33 | A failure at calculate stops only its own row. | documentation/2. design/batch-evaluation.md |
| share-calc-fail-r2 | round(total / people, 2) | total=dec:10.00, people=int:0 | problems: divide-by-zero | A failure at calculate stops only its own row. | documentation/2. design/batch-evaluation.md |
| share-calc-fail-r3 | round(total / people, 2) | total=dec:90.00, people=int:4 | result: dec:22.50 | A failure at calculate stops only its own row. | documentation/2. design/batch-evaluation.md |

### (e) Pushed down

The total and the payout are written out on each row, as the table has no shared column. The test passes them once in `Shared` for a second run of the same rows, and both runs must give the same answers.

Provisional: the pushed-down declaration and `Aggregates()` depend on the language page, which has no declaration syntax yet.

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| fund-share-r1 | round(contribution / total_contributions * payout, 2) | contribution=dec:250.00, total_contributions=dec:1000.00, payout=dec:400.00 | result: dec:100.00 | A total passed in works as an input on every row. | documentation/2. design/batch-evaluation.md |
| fund-share-r2 | round(contribution / total_contributions * payout, 2) | contribution=dec:500.00, total_contributions=dec:1000.00, payout=dec:400.00 | result: dec:200.00 | A total passed in works as an input on every row. | documentation/2. design/batch-evaluation.md |
| fund-share-r3 | round(contribution / total_contributions * payout, 2) | contribution=dec:250.00, total_contributions=dec:1000.00, payout=dec:400.00 | result: dec:100.00 | A total passed in works as an input on every row. | documentation/2. design/batch-evaluation.md |

The list form of this case, the same formula given the list itself, is G09 below, because the case table has no list type.

## Cases the table cannot hold, generated in the test

These cases are built in Go, not loaded from the table. Row counts and sweeps are proposals. Each runs the test steps above unless it says otherwise.

| Id | Case | Rows loaded | Expected |
|---|---|---|---|
| G01 | Empty batch | 0 rows | 0 results, no error, a summary with every count 0 |
| G02 | Exactly the rows-per-call maximum | 10,000 rows of example B's formula, valid rows from a fixed seed (Provisional: 10,000 is a proposed limit) | Accepted; 10,000 results in order, `Rows[i].Row` equals `i+1` |
| G03 | One row over the maximum | 10,001 rows (Provisional: 10,000 is a proposed limit) | Refused with a `BatchError`, code `too-many-rows`, `Limit` 10,000, `Got` 10,001; the result holds no `Rows` and no row was calculated |
| G04 | Chunk size and workers sweep | 250 rows of formula 1 from a fixed seed: good rows, rows with 1 and 2 problems, and rows that fail at calculate | With chunk sizes 1, 7, 250 and 251 and workers 1 and 4, the results are identical to the run with workers 1 and chunk size 100, and each also equals single |
| G05 | Chunk size over the maximum | 3 rows, chunk size 10,001 (Provisional: 10,000 is a proposed limit) | Refused with code `bad-chunk-size`, `Limit` 10,000, `Got` 10,001; no row calculated |
| G06 | Workers over the maximum | 3 rows, workers 65 | Refused with code `bad-workers`, `Limit` 64, `Got` 65; no row calculated |
| G07 | Defaults | 250 rows, chunk size 0 and workers 0 | No error; the summary reports `ChunkSize` 100 and `Workers` 1; the results equal those of G04's reference run |
| G08 | Shared name also in the row | 3 rows of formula 1; `Shared` holds `payment` 10.00; row 2 also holds `payment` 99.00 | Rows 1 and 3 use 10.00 and row 2 uses 99.00, the row's own value; each equals `Run` on the merged row |
| G09 | Pushed-down total equals the list | The 3 rows of case (e) with `Shared` holding `total_contributions` 1000.00 and `payout` 400.00; and the formula `round(contribution / sum(contributions) * payout, 2)` run alone for each row with the list 250.00, 500.00, 250.00 (Provisional: the list input type and the declaration depend on the language page) | Both give 100.00, 200.00, 100.00, with the same type and exact text |
| G10 | Batch limit reached | Enough rows of formula 1 to pass the batch cost limit, 1 worker (Provisional: needs the limits set by the engine choice) | Cost limit with 1 worker, so the unstarted rows are the same on every run; unstarted rows carry `batch-limit-reached`, stage `batch`, and are counted in `NotRun`; started rows equal single; the call returns no error |
| G11 | Cancelled context | 250 rows, a context cancelled before the call and a second case cancelled after the first chunk is started | Rows not started carry `batch-cancelled`, stage `batch`; started rows equal single; the call returns no error; the second case does not depend on which rows were started |
| G12 | Hostile rows in a batch | Hostile page group B, B01 to B05, run through `RunBatch` (Provisional: the hostile inputs depend on the language page and the limits) | Each row holds as written on [hostile formulas](hostile-formulas.md); the batch returns without error and the good rows equal single |
| G13 | Random run | Law 4 generators from [laws](laws.md): the formula generator and the row generator, 1 to 200 rows (proposed) | 500 cases in a normal run and 20,000 in the nightly run (both proposed, from laws); every case passes the test steps with its seed in any failure line |

The G11 second case keeps its check to the rule in the design page: which rows are left unstarted is repeatable only for a cost limit with 1 worker, so the test compares started rows and checks that every row is either equal to single or carries one of the two exempt codes.

## Feeds the bulk-run cases

The later bulk-run cases (issue ISS-17) rely on these, and no more. They are defined on [batch evaluation](../2.%20design/batch-evaluation.md) (sections 1 and 3 and the numbers table) and are not restated beyond this list:

- The chunk size rule: chunk k holds rows (k-1) x `ChunkSize` + 1 to the smaller of k x `ChunkSize` and the row count; only the last chunk may be shorter; the default is 100 and the range is 1 to 10,000.
- The order of rows: result i answers input row i, and `Row` is i+1.
- The per-row answer: exactly one of a value or a non-empty problem list, with the stage.

## Where it is recorded

`documentation/3. implementation` should link [batch evaluation](../2.%20design/batch-evaluation.md) and this page; that edit is made by the lead after the pages are written, not by this page.

## Open questions

The owner has not decided any of these. Each has a proposed answer.

1. Is the batch-case naming right: consecutive rows `<case>-r1` to `<case>-rN`, the same formula on each, the loader grouping them into one call? Proposed: yes, because it needs no new syntax and one table feeds both sides. The loader must also fail a case whose rows skip a number or whose formulas differ.
2. Is `Run` the name of the single-row run? Proposed: yes, for this page; the language page may rename it, and then only this page's signatures and the test change.
3. Does the test live in the root package? Proposed: yes, next to the case check that example cases proposes, run by `go test ./...`.
4. Are the generated row counts right: 250 rows for the sweep and defaults, 10,000 and 10,001 for the limit, 1 to 200 for the random run? Proposed: yes, as a starting point; the random run's counts must equal the language's limits once that page fixes them.
