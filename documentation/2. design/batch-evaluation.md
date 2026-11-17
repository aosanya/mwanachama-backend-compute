# Batch evaluation

Status: the owner has not reviewed this page beyond the decisions listed below. Every number marked proposed is invented. A sentence that starts with "Provisional:" depends on a page or choice that does not exist yet: the formula language page (`documentation/2. design/formula-language.md`), the shape of a Validate problem, the limits, or the engine. When those are settled the marker is dropped or the sentence is corrected.

End goal: a host calculates a whole month for every member in one call, gets one answer per row in the order it sent them, and never loses good rows because of a bad one.

All figures on this page are invented. The repo is public.

## Decided by the owner

Decided on 2026-11-17:

1. Bulk runs are designed in from the start.
2. A batch run takes one compiled formula and a list of input rows. It returns, per row and in the same order, either a result or the list of problems that stopped the row.
3. Every row is validated against the input rules and checks before it is calculated. A row with problems is not calculated and carries its full list of problems. Each problem names the input, the value received and the rule broken, plus a plain sentence and a code.
4. A bad row never fails the batch. A row that fails during calculation carries its reason the same way.
5. The program is immutable, so rows may be worked in parallel. The chunk size and the number of workers are the caller's to set.
6. The cost limit applies to the whole batch as well as to each row. Changed by decision 19: the library has no time limit of its own, so the earlier time limit is gone.
7. Aggregates over many rows (sum, count, average) are worked out by the database where the data lives, and only the total is passed in. This page says which aggregates may be pushed down and how a formula declares one.
8. A batch result equals running each row alone. A test proves it on invented data.
9. The host records one summary entry for a batch, not one entry per row, because the inputs are people's financial figures. How long anything is kept is the host's rule and a separate issue.
10. Each agency sets the largest number of rows one batch may hold, and a batch over the limit is refused whole before any row is worked out (decision 16).
11. A bulk run has two stages: a calculation batch that works out every row and finishes as a whole, then an actualization batch that posts or releases the results. This page and the library are the calculation stage only (decision 17).

Everything else on this page is a proposal or an open question and says so.

## The rules in numbers

| Rule | Value | Status |
|---|---|---|
| Row number | 1-based; row number is the position in the list plus 1 | Proposed |
| Order | Result i answers input row i; result count equals input row count; never reordered, whatever order workers finish | Proposed |
| Chunk rule | Rows are cut into consecutive chunks of `ChunkSize` rows; one worker takes a whole chunk and works its rows in order | Proposed |
| Chunk size | Default 100 when 0 is given; allowed 1 to 10,000 and not above the rows-per-call maximum; anything else is refused with `bad-chunk-size` | Proposed |
| Workers | Default 1 when 0 is given; allowed 1 to 64; anything else is refused with `bad-workers` | Proposed |
| Rows per call | At most the agency's own setting (decision 16); more is refused with `too-many-rows`. 10,000 is the proposed default and the figure the tests use | The agency's setting is decided; the default and the ceiling are Provisional |
| Empty batch | 0 rows gives 0 results, no error, and a summary of zeros | Proposed |
| Shared inputs | One set of inputs applied to every row; the row's own value wins when a name is in both | Proposed |

Provisional: the limits below are not set until the engine is chosen.

| Limit | Value |
|---|---|
| Cost of one row | to be set by the engine choice |
| Time of one row | to be set by the engine choice |
| Cost of a whole batch | to be set by the engine choice |
| Time of a whole batch | to be set by the engine choice |
| Input size of a whole batch | to be set by the engine choice |

## 1. Each row has its own answer, and its own problems

Row i gets exactly one of two answers: a value, or a non-empty list of problems. It never gets both and never neither. The answer also names the stage that stopped the row: `validate`, `calculate`, or `batch` (the row was never started, see section 4).

Provisional: a problem holds six parts: the input name, whether the input was given, the value received as exact text, the rule broken, a plain sentence and a code. The Validate problem shape on the language page decides the final parts.

Provisional: problems inside one row come in the order the language page sets. Proposed until then: input declaration order, then rule order within an input.

Provisional: the codes named below (`required`, `no-negatives`, `decimal-places`, `divide-by-zero`) are the ones the case files already use; the language page owns the list.

The library keeps no row key. The host maps a row number to a member. This is an open question below.

Example A. A month-end run of formula 1 from `documentation/1. requirements/example-formulas.md`, `payment >= balance ? 0.00 : balance - payment`. Both inputs are required, may not be negative and have 2 decimal places. Five rows:

| Row | balance | payment | Answer | Stage |
|---|---|---|---|---|
| 1 | 120.50 | 45.25 | 75.25 | |
| 2 | 80.00 | -5.00 | Problem: `payment`, received `-5.00`, rule no negatives, code `no-negatives` | validate |
| 3 | 60.00 | 70.00 | 0.00 | |
| 4 | 200.00 | not given | Problem: `payment`, not given, rule required, code `required` | validate |
| 5 | -1.00 | 5.125 | Two problems: `balance` (`no-negatives`), then `payment` (`decimal-places`), in input order | validate |

Five rows come back, in the order sent. Rows 1 and 3 have values, rows 2, 4 and 5 stopped at validate, and there are 4 problems in total. The batch returns without error.

Example B. A failure during calculation, `round(total / people, 2)`, with `total` a decimal and `people` an integer and no range rule on `people`:

| Row | total | people | Answer | Stage |
|---|---|---|---|---|
| 1 | 100.00 | 3 | 33.33 | |
| 2 | 10.00 | 0 | Problem: code `divide-by-zero` | calculate |
| 3 | 90.00 | 4 | 22.50 | |

Row 2 passes validation, because nothing forbids 0, and stops while it is calculated. Rows 1 and 3 are not touched by it.

## 2. Validation comes first

Every row is validated before it is calculated. A row with one problem or more is never calculated, and its list holds every problem found, not only the first. This is law 7 on [laws](../4.%20qa/laws.md).

A row that passes validation is calculated. If the calculation fails, the row carries that single reason with stage `calculate`, in the same problem form.

Example A, row 5, shows it: the row breaks 2 rules and lists both, and the formula is never worked on `-1.00` and `5.125`.

## 3. Rows and chunks are worked in parallel

The program is immutable. A row is worked with its own state, no mutable state is shared between rows, and each result is written to its own position in the result list.

Proposed chunk rule: rows are cut into consecutive chunks of `ChunkSize` rows. Chunk k holds rows (k-1) x `ChunkSize` + 1 up to the smaller of k x `ChunkSize` and the row count. Only the last chunk may be shorter. One worker takes a whole chunk and works its rows in order. With 250 rows and `ChunkSize` 100 there are 3 chunks: rows 1 to 100, 101 to 200, 201 to 250.

The chunk size and the number of workers are the caller's to set, within the ranges in the table above. The defaults are proposals: chunk size 100, workers 1. A value outside its range refuses the whole call before any row is worked; the error is a batch error with the code, the limit and the number received.

Neither number may change any row's answer. The test runs the same rows with chunk sizes 1, 7, the row count and the row count plus 1, and with 1 and 4 workers, and the results must be identical.

## 4. Limits apply to the whole batch and to each row

Provisional: a row over its own cost limit at run time gets a problem with the refusal sentence from [hostile formulas](../4.%20qa/hostile-formulas.md), for example "This formula goes over the cost limit of {value}." It is a failed calculation and follows the step's failure choice (decision 19). The rest of the batch carries on. Group B on that page (B01 to B05) must hold.

Provisional: the batch has its own cost limit, and so does each row. There is no time limit (decision 19). The values are to be set by the engine choice, as the limits table says.

Provisional: a chunk is started only while the batch limit has not been reached. A chunk that has started finishes its rows, each still bound by its own row limits. Rows never started carry the problem code `batch-limit-reached` with stage `batch` and are not calculated.

Provisional: whether and how the library honours a host's own cancel or deadline is open (decision 19, open question 10). If it does, rows not started then carry the code `batch-cancelled` with stage `batch`.

The batch call itself still returns with no error in both cases. Only a refused call (too many rows, bad chunk size, bad workers) returns an error, and then no row is calculated and nothing partial comes back.

Rows per call: at most the agency's setting (decision 16), with 10,000 as the Provisional default. A call over the setting is refused whole. Who cuts a run larger than the setting into several calculation batches, the platform or the builder, is open (agency steps decision 11). An empty batch returns 0 results and a summary of zeros.

Law 4 exemption: rows that carry `batch-limit-reached` or `batch-cancelled` are not compared with a single run, because a single run never meets a batch limit. Every other row is compared.

Provisional: which rows are left unstarted is repeatable only for a cost limit with 1 worker. With more than 1 worker it is not repeatable, and a test must not depend on which rows were left.

## 5. Aggregates are pushed down to the database

The library is a pure library and never reads a database. When a formula needs the sum of everyone's contributions, the host asks its database once, where the data lives, and passes the total in as an ordinary input. Passing it in Shared inputs applies it to every row.

Everything in this section is Provisional: it depends on the language page, which has no declaration syntax yet.

Provisional: how a formula declares one. The input table has a rule cell reading `pushed down: <function> of <measure>`, for example `pushed down: sum of contributions`. The measure is a name the host maps to its own column.

Provisional: the program exposes its declarations, so the host knows what to ask its database. Validate treats a pushed-down input as an ordinary input of its type.

Provisional: which aggregates may be pushed down:

| May be pushed down | Over |
|---|---|
| sum, count, min, max | decimal or integer values only |
| average | only as two pushed inputs, a sum and a count, divided inside the formula |

Average is divided in the formula because the formula's own rounding stays in the formula; a database average might round differently or use floating point.

Provisional: what is not pushed down. Float sums, because the order the database adds in can change the answer (law 5). Median and percentiles. Distinct counts. Anything that depends on order.

The host must use an exact decimal type in the database for a pushed-down sum.

Stated limit: the library cannot check that the total covers the same rows as the batch. Matching them is the host's duty.

Example C. Calculate each member's share of a fund, `round(contribution / total_contributions * payout, 2)`:

| Input | Type | Rule |
|---|---|---|
| contribution | decimal | required, no negatives, 2 decimal places |
| total_contributions | decimal | required, `pushed down: sum of contributions` |
| payout | decimal | required, no negatives, 2 decimal places |

The host asks its database once for the sum of contributions and gets 1000.00. It passes `total_contributions` 1000.00 and `payout` 400.00 as Shared inputs, with 3 rows:

| Row | contribution | Answer |
|---|---|---|
| 1 | 250.00 | 100.00 |
| 2 | 500.00 | 200.00 |
| 3 | 250.00 | 100.00 |

The same formula written with the list itself, `sum(contributions)` over 250.00, 500.00 and 250.00, gives the same 3 answers. A test checks that a total passed in equals the formula given the list, on invented data.

## 6. A batch equals running each row alone

Row i of a batch result equals the result of running row i alone, with the shared inputs merged under the row's own values. Provisional: equal means the following, and the parts of a problem follow the Validate problem shape:

| Compared | How |
|---|---|
| Value | Type and exact text, so 12.5 differs from 12.50 |
| Problems | Same count, order, input, received value, rule, sentence and code |
| Stage | Same stage |

Duration and the summary are not compared. This is law 4 on [laws](../4.%20qa/laws.md), with the exemption in section 4. The signatures and the test cases are on [batch equals single](../4.%20qa/batch-equals-single.md).

## 7. What the host records

The host records one summary entry for a batch. The library returns the summary and the host writes it; the library stores nothing.

The summary holds:

| Part | Note |
|---|---|
| Formula hash | And the name and version of the formula and of every library version it calls |
| Row count | |
| Valued, stopped and not-run counts | They add up to the row count |
| Problems per code | A count per code, never the problems themselves |
| Chunk size and workers | |
| Duration | |

The summary never holds input values, per-row results, per-row problems (they hold the value received) or who the rows are.

The result of the call goes back to the caller for that call. Whether and how long the host keeps the result or the summary is the host's retention rule, a separate issue. This page does not decide it.

## Open questions

The owner has not decided any of these. Each has a proposed answer.

1. The limit is each agency's own setting (decision 16). Is 10,000 right as the default, and what ceiling may no agency go above? Also: default chunk size 100, default 1 worker, at most 64 workers? Proposed: yes, as a starting point. Provisional: rows per call stays open until the limits are set.
2. [Laws](../4.%20qa/laws.md) uses 1 to 200 rows per batch for the random row generator and says the generator limits must equal the language's limits. Proposed: keep 200 for the random run and add fixed cases at 10,000 rows (accepted) and 10,001 rows (refused with `too-many-rows`).
3. The hostile page describes a batch as "a list of rows, each holding a formula and its inputs", while this page runs one formula over many rows. Proposed: a row is a row of inputs, and a mix of formulas is several calls.
4. Settled by decision 19: there is no time limit, so the question of whether it applies to each formula or to the whole batch falls away. The cost limit applies to both. The number is kept so the other references hold.
5. Should a row carry a key from the host, such as a member id? Proposed: no, only the row number.
6. Should results stream back instead of coming as one list? Proposed: not now, because a call holds at most the agency's row limit, 10,000 by default.
7. The case-file page names the design page `batch.md` and its case file `batch.cases.md`; this page is `batch-evaluation.md`. Proposed: settle one name before the loader is built. Until then the examples here are written inline and are not `formula` blocks with case markers.
8. Problem codes: the language page owns the list. Provisional: this page needs `required`, `no-negatives`, `decimal-places` and `divide-by-zero`. Proposed new codes for the language page to adopt: `batch-limit-reached` and `batch-cancelled` as row problem codes, and `too-many-rows`, `bad-chunk-size` and `bad-workers` as codes of a refused call.
9. Should the summary also keep the pushed-down totals? Proposed: yes, because they are fund totals and not one person's figure. The owner decides.
10. Is a host's own cancel or deadline honoured, and how? Decision 19 leaves it open. Proposed: yes, a started chunk finishes its rows and rows not started carry `batch-cancelled`, as the bulk-runs page assumes in BR-72 to BR-77.
11. Who cuts a run larger than the agency's row limit into several calculation batches? (agency steps decision 11) No proposal is made here.
