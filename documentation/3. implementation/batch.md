# Batch evaluation

Package `batch/`, built for ISS-29. It takes one compiled function and a list of rows and gives back one answer per row, in the order sent. A bad row never fails the batch. The design is [batch-evaluation.md](../2.%20design/batch-evaluation.md), the cases are [bulk-runs.md](../4.%20qa/bulk-runs.md) and [batch-equals-single.md](../4.%20qa/batch-equals-single.md). The decisions are 9, 16, 17 and 19 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## What it holds

| File | What it is |
|---|---|
| `types.go` | `Stage`, `RowResult`, `Options`, `Summary`, `Result`, `BatchError`, the codes and the proposed limits |
| `compile.go` | `Compile` and `Merge` |
| `run.go` | `Run`, the one-row path every batch row takes, and the rows that were not started |
| `batch.go` | `RunBatch`: the plan, the chunks, the workers and the summary |
| `check_runner.go` | the `function.Runner` that works out a function's checks with `formula.Plan` |
| `convert.go` | text read by `numtext` turned into the typed values of `formula` |

Imports: the standard library, `formula/`, `function/`, `number/`, `numtext/` and `problem/`. No comments in the code, and no clock is read except to time the call for the summary.

## Using it

```go
compiled, problems := batch.Compile(definition, "round(balance * rate, 2)", nil, formula.DefaultOptions())
result, err := compiled.RunBatch(ctx, rows, batch.Options{ChunkSize: 100, Workers: 4, MaxRows: agency.MaxBatchRows})
alone := compiled.Run(batch.Merge(shared, rows[0]))
```

| Name | What it does |
|---|---|
| `Compile(definition, body, runtime, options)` | Refuses a bad definition (`function.CheckDefinition`), reads and checks the body, plans it once. Gives the problems and no program when any is found. Many callers may share one `Compiled`. |
| `(*Compiled).RunBatch(ctx, rows, options)` | One answer per row. An error comes only from a refused call, and then no row is worked and no result comes back. |
| `(*Compiled).Run(row)` | The same path for one row. `Row` is 0. A row of a batch equals this on the merged row. |
| `Merge(shared, own)` | The shared inputs under the row's own values. Neither argument is changed. |

A row is a `function.Row`: a map from input name to `numtext.Received`, so the host passes the digits as written and the library reads them.

## The answer of a row

`RowResult` holds `Row` (1-based, the position plus 1), `Value`, `Problems` and `StoppedAt`. A row has a value or a non-empty list of problems, never both and never neither (`HasValue` says which).

| Stage | When | Problems |
|---|---|---|
| `validate` | The row broke an input rule or a check. Every problem comes at once, in the order of [problem.md](problem.md). The row is not calculated. | `problem.Problem` from `function.Validate` |
| `calculate` | The row was valid and the run failed (`divide-by-zero`, `too-costly`, `empty-list` and the others of [formula-run.md](formula-run.md)). It is a failed calculation and follows the step's failure choice (decision 19). | One `problem.Problem` with rule `calculation`, the code and sentence of the run |
| `batch` | The row was never started. | One problem with rule `batch` and the code `batch-limit-reached` or `batch-cancelled` |

An input that is a list and is not given is an empty list for the formula. A check is worked out by planning its formula once for the names it reads and caching the program, so a check costs one plan for the whole batch.

## The two limits of a call

| Setting | Value | Refusal |
|---|---|---|
| `MaxRows` | The agency's setting (decision 16), passed in on every call. 0 reads as the proposed default 10,000. | `too-many-rows`, limit `MaxRows`, got the rows sent |
| `ChunkSize` | 0 reads as 100; 1 to 10,000 and not above `MaxRows` | `bad-chunk-size`, limit the smaller of 10,000 and `MaxRows` |
| `Workers` | 0 reads as 1; 1 to 64 | `bad-workers`, limit 64 |

The rows are tested first, then the chunk size, then the workers, so 10,001 rows with 65 workers gives `too-many-rows`. A batch of exactly `MaxRows` is accepted. An empty batch gives no rows, no error and a summary of zeros. Where the default chunk size is above a small `MaxRows`, the default is cut to `MaxRows`.

`Options.Cost` is the cost limit of the whole batch, in the units of one run (`Meter.Spent`). 0 means no batch limit. A chunk is started only while the cost spent by the rows so far is below the limit, so a limit reached exactly (BR-68) starts no more. A chunk that has started finishes its rows. A row stopped at validate spends nothing. Rows never started carry `batch-limit-reached` and are counted in `NotRun`. With 1 worker the rows left are the same on every run; with more they are not, and the tests ask only that every row is either equal to its single run or not started.

## Cancel

`RunBatch` reads `ctx` at each chunk edge. A chunk that finds the context done carries `batch-cancelled` on every row, and one that has started finishes. The call returns no error. A cancel before the call leaves every row not run. This is the proposal of the design page (open question 10) and the owner has not decided it: it is one function, `refusalOf` in `batch.go`, and the hooks `beforeChunk` and `afterChunk` are test-only. The library keeps no state, so a restart is the host calling again with the rows that carry one of the two codes; the host keeps each finished row once and posts it once.

## Parallel work

Chunk k holds rows (k-1) x `ChunkSize` + 1 up to the smaller of k x `ChunkSize` and the row count. Each worker takes the next chunk from one counter and works its rows in order. A row is written to its own position in the result list, and the only shared values are the atomic counters of the next chunk and of the cost spent, so no row can see another. The program is immutable and every row has its own meter. The chunk size and the workers change no answer: the tests run 250 rows with chunk sizes 1, 7, 250 and 251 and with 1, 4 and 64 workers and compare the rows.

## Two stages (decision 17)

This package is the calculation stage only. It changes nothing real, so a host may hold, repeat or cancel the result before the actualization batch posts or releases it. Cutting a run larger than `MaxRows` into several calculation batches belongs to the host.

## The summary

`Summary` holds the formula hash (the SHA-256 of the standard text, so a respaced formula has the same hash), the counts `Rows`, `Valued`, `Stopped` and `NotRun` (they add up to `Rows`), `ProblemsByCode` (a count per code and never the problems themselves), the chunk size and workers actually used, and the duration. It holds no input value and no row's answer. The library returns it and the host writes the one entry for the batch.

## Proposed on this page

- The sentences of the three refused calls: `This batch has {got} rows and at most {limit} are allowed.`, `The chunk size is {got} and it must be from 1 to {limit}.` and `The number of workers is {got} and it must be from 1 to {limit}.`
- The sentences of the two not-run codes and the rule names `calculation` and `batch`.
- A calculation problem has no `input` and no `check`: the part and path of the formula problem are not carried.
- The chunk limit follows a small `MaxRows`, as above.
- The cost of the whole batch is the sum of the costs of the rows that were run, read from each row's own meter. The validation of a row and of its checks is not counted.

## Not built, and what stays open

1. `(*Program).Aggregates` and the declaration `pushed down: <function> of <measure>`. The language page has no declaration syntax yet. Pushed-down totals work as shared inputs, and a test shows a total passed in equals the formula given the list.
2. The name and version of the library functions in the summary. A `formula.Runtime` is the host's, so the host adds them.
3. The default and the ceiling of the agency's row limit (open question 1 of the design page): 10,000 is a proposed default only.
4. Whether and how a host's cancel is honoured (decision 19; open question 10), as above.
5. Who cuts a larger run into several calculation batches (open question 11).
6. Whether a row carries a key from the host, and whether results stream (open questions 5 and 6): no and no.
7. A row that is stopped at validate does not use up any of the batch limit. BR-69 holds either way, but the owner has not said which is meant.
8. Decision 22 against the language page's `no-negatives` rule, which `function/` builds as the language page is written.
9. The test-only hooks that the QA page asks for to count rows in work (BR-80 to BR-85). No memory case is built.
10. The batch-equals-single test lives in `batch/`, not in the root package (open question 3 of batch-equals-single.md), because the root package would import `batch/`.

## Tests

`go vet ./batch/` is clean and `go test ./batch/ -race` passes. The files are `helpers_test.go`, `rows_test.go` (examples A and B, the pushed-down example C, shared inputs, the summary, every row equal to its single run), `limits_test.go` (BR-60 to BR-64, BR-65 to BR-70, BR-72 to BR-77 with the join of two calls) and `parallel_test.go` (the chunk and worker sweep, BR-12, BR-13, no chunk twice, one compiled formula used by 16 callers at once). The hostile group B cases (BR-71) are covered by a row over its own cost limit and a list over 1000 items; the full hostile page is ISS-34's.
