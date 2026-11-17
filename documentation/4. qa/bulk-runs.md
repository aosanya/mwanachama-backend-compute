# Bulk runs: test design

Status: draft for owner review. The owner has not reviewed this page. Every number marked proposed is an invented starting point. A sentence that starts with "Provisional:" depends on the formula language page (`documentation/2. design/formula-language.md`, still being written), on the Validate problem shape, on the limits or on the engine choice; the marker is dropped or the sentence corrected when those are settled.

All names and figures are invented. The repo is public.

## What this protects

End goal: a bulk run such as month-end interest is correct for every row, whatever its size, and a restart changes nothing. Speed is the concurrency and speed issue; this page is about correctness.

It is written against [batch evaluation](../2.%20design/batch-evaluation.md) and [batch equals single](batch-equals-single.md) and uses their numbers and names without restating the rules:

| Rule | Value |
|---|---|
| Chunk k | rows (k-1) x `ChunkSize` + 1 to the smaller of k x `ChunkSize` and the row count; only the last chunk may be shorter |
| `ChunkSize` | default 100 (0 gives the default); 1 to 10,000 |
| `Workers` | default 1 (0 gives the default); 1 to 64 |
| Rows per call | at most the agency's setting (decision 16), proposed default 10,000; more is refused whole with `too-many-rows` before any row is worked out |
| Row number | position in the list plus 1 |
| Answer of a row | exactly one of a value or a non-empty problem list, with the stage `validate`, `calculate` or `batch` |
| Stages of a bulk run | a calculation batch that finishes as a whole, then an actualization batch that posts or releases the results. The library does only the calculation stage (decision 17); every case here is a calculation batch |

## How a case is read

Formulas and inputs of each case are set once below. A case that is written out in full lists every row with its exact answer; a large one is described by a rule that gives the answer of every row. Written-out cases load into the case table as rows `<case>-r1` to `<case>-rN` as described on [batch equals single](batch-equals-single.md); the `formula`, `rule` and `page` cells repeat the setup, and the `expected` cell holds `result: dec:...` or `problems: <codes>`. The stage column below is checked against the single run, not the table.

Provisional: the formula syntax is illustrative until the language page fixes it. The codes (`required`, `no-negatives`, `decimal-places`, `divide-by-zero`, `batch-limit-reached`, `batch-cancelled`) and the order of problems inside a row (input declaration order, then rule order within an input) are the batch page's proposals. A problem is written `input (value received) code`.

### Formula M: month-end interest

`round(balance * rate, 2)`

| Input | Type | Rules |
|---|---|---|
| balance | decimal | required, no negatives, 2 decimal places |
| rate | decimal | required, no negatives, at most 4 decimal places |

Rule R1, used by every large case: row i has balance 10.00 x i and rate 0.005, so its answer is 0.05 x i, exact at 2 places with no rounding. Every row has its own answer, so a row that is moved, doubled or missing shows. Sum of the answers of rows 1 to n is 0.05 x n x (n + 1) / 2.

### Formula S: even share

`round(total / people, 2)`

| Input | Type | Rules |
|---|---|---|
| total | decimal | required, no negatives, 2 decimal places |
| people | integer | required; no range rule, so 0 passes validation |

### Formula F: fund share

`round(payout * contribution / total_contributions, 2)`

Multiplication comes first so a quotient that does not end is divided once and rounded once, as in example B of the batch page. Inputs: `contribution`, `payout` (decimals, required, no negatives, 2 places) and `total_contributions` (decimal, required, pushed down: sum of contributions).

### Checks every case runs

1. No error, one result per input row, and `Rows[i].Row` equals i+1.
2. Every row equals its expected answer, type and exact text (0.10 is not 0.1).
3. The summary matches the results: Valued + Stopped + NotRun equals Rows, and `ProblemsByCode` counts every problem.
4. Every row equals the same row run alone (section 10).

## 1. Sizes

All with rule R1, `ChunkSize` 100 and 1 worker.

| Id | Rows | Chunks | Last chunk | Row 1 | Last row | Sum of answers |
|---|---|---|---|---|---|---|
| BR-01 | 0 | 0 | none | none | none | 0 results, no error, summary all 0 |
| BR-02 | 1 | 1 | rows 1 to 1 | 0.05 | 0.05 | 0.05 |
| BR-03 | 3 | 1 | rows 1 to 3 | 0.05 | 0.15 | 0.30 |
| BR-04 | 99 (one short of a chunk) | 1 | rows 1 to 99 | 0.05 | 4.95 | 247.50 |
| BR-05 | 100 (exactly a chunk) | 1 | rows 1 to 100 | 0.05 | 5.00 | 252.50 |
| BR-06 | 101 (one over) | 2 | rows 101 to 101 | 0.05 | 5.05 | 257.55 |
| BR-07 | 200 | 2 | rows 101 to 200 | 0.05 | 10.00 | 1,005.00 |
| BR-08 | 201 | 3 | rows 201 to 201 | 0.05 | 10.05 | 1,015.05 |
| BR-09 | 10,000 (the maximum) | 100 | rows 9,901 to 10,000 | 0.05 | 500.00 | 2,500,250.00 |

In each, Valued equals the row count and Stopped and NotRun are 0. BR-03 written out: balances 10.00, 20.00, 30.00 give 0.05, 0.10, 0.15.

BR-10, chunk edges written out. 7 rows by rule R1, `ChunkSize` 3: chunks are rows 1 to 3, 4 to 6 and 7 to 7. Answers 0.05, 0.10, 0.15, 0.20, 0.25, 0.30, 0.35. The same 7 answers come back with `ChunkSize` 1 (7 chunks), 2 (4 chunks), 7 and 8 (1 chunk), and with 1 and 4 workers.

BR-11, chunk sweep. 250 rows by rule R1: with `ChunkSize` 100 the chunks are rows 1 to 100, 101 to 200, 201 to 250. With `ChunkSize` 1 there are 250 chunks, with 7 there are 36 (the last is rows 246 to 250), with 250 and 251 there is 1. All answers are 0.05 x i whatever the chunk size and worker count (1, 4, 64).

## 2. Results come back in order, one per row

BR-12. Rows written out, `ChunkSize` 2, 4 workers, so 4 chunks run together. Formula M, rate 0.005. The test holds chunk 1 until chunks 2 to 4 have finished, so the first rows finish last.

| Row | balance | Answer |
|---|---|---|
| 1 | 80.00 | 0.40 |
| 2 | 10.00 | 0.05 |
| 3 | 60.00 | 0.30 |
| 4 | 30.00 | 0.15 |
| 5 | 50.00 | 0.25 |
| 6 | 20.00 | 0.10 |
| 7 | 70.00 | 0.35 |
| 8 | 40.00 | 0.20 |

8 results, row numbers 1 to 8 each once and in order, Valued 8. Provisional: holding a chunk needs a test-only hook (open question 7).

BR-13. 10,000 rows by rule R1, `ChunkSize` 100, 4 workers. The row numbers are exactly 1 to 10,000, none missing and none twice, and result i holds 0.05 x i.

## 3. Bad rows

A bad row never fails the batch, carries every problem it has, and does not touch the rows around it. Provisional: the problem codes and order are those of the batch page.

BR-20. Formula M, `ChunkSize` 4: chunks are rows 1 to 4, 5 to 8, 9 to 10. Bad rows sit at the start (1), at the end of a chunk (4), at the start of the next chunk (5) and at the end of the batch (10).

| Row | balance | rate | Answer | Stage |
|---|---|---|---|---|
| 1 | -10.00 | 0.005 | balance (-10.00) no-negatives | validate |
| 2 | 20.00 | 0.005 | 0.10 | |
| 3 | 30.00 | 0.005 | 0.15 | |
| 4 | 40.005 | 0.005 | balance (40.005) decimal-places | validate |
| 5 | not given | 0.005 | balance (not given) required | validate |
| 6 | 60.00 | 0.005 | 0.30 | |
| 7 | 70.00 | 0.005 | 0.35 | |
| 8 | 80.00 | 0.005 | 0.40 | |
| 9 | 90.00 | 0.005 | 0.45 | |
| 10 | -100.00 | 0.00505 | balance (-100.00) no-negatives, then rate (0.00505) decimal-places | validate |

Valued 6, Stopped 4, NotRun 0. Problems: 5 in all, `no-negatives` 2, `decimal-places` 2, `required` 1.

BR-21, every position. 10 batches of 10 rows by rule R1, with one bad row at position p, from 1 to 10: its balance is -(10.00 x p). Run with `ChunkSize` 4, 1, 10 and 11. Row p carries the one problem `balance (-10.00 x p) no-negatives` at stage `validate`; the other 9 rows carry 0.05 x i. Valued 9, Stopped 1.

BR-22, one bad among many. 100 rows by rule R1 except row 57, balance -570.00. Row 57 carries `balance (-570.00) no-negatives`; the other 99 carry 0.05 x i. Valued 99, Stopped 1.

BR-23, both sides of every chunk edge. 10,000 rows by rule R1, `ChunkSize` 100, except that rows with i mod 100 equal to 0 or 1 have balance -(10.00 x i). That makes the last row of each chunk and the first row of the next one bad (row 1 and row 10,000 have no neighbour across an edge), 200 bad rows. Valued 9,800 with a sum of answers of 2,450,245.00 (the 2,500,250.00 of BR-09 less 50,005.00 of the bad rows). Stopped 200, `no-negatives` 200.

BR-24, every row bad, written out. Formula M.

| Row | balance | rate | Answer | Stage |
|---|---|---|---|---|
| 1 | not given | 0.005 | balance (not given) required | validate |
| 2 | -5.00 | 0.005 | balance (-5.00) no-negatives | validate |
| 3 | 12.345 | 0.005 | balance (12.345) decimal-places | validate |
| 4 | -5.125 | -0.005 | balance (-5.125) no-negatives, balance (-5.125) decimal-places, rate (-0.005) no-negatives | validate |
| 5 | not given | not given | balance (not given) required, rate (not given) required | validate |

Valued 0, Stopped 5. Problems: 8 in all, `required` 3, `no-negatives` 3, `decimal-places` 2. The call returns no error.

BR-25, every row bad, large. 250 rows with `balance` not given and rate 0.005. Each row carries the one problem `balance (not given) required`. Stopped 250, `required` 250.

## 4. Both kinds of failure in one batch

Provisional: the stage names and the order of problems follow the batch page. A row stopped by an input rule has stage `validate` and is never calculated. A row that passes validation and fails during the calculation has stage `calculate`.

BR-30. Formula S, `ChunkSize` 3: chunks are rows 1 to 3, 4 to 6, 7 to 8.

| Row | total | people | Answer | Stage |
|---|---|---|---|---|
| 1 | 100.00 | 3 | 33.33 | |
| 2 | 10.00 | 0 | divide-by-zero | calculate |
| 3 | -5.00 | 2 | total (-5.00) no-negatives | validate |
| 4 | 90.00 | 4 | 22.50 | |
| 5 | 20.00 | 0 | divide-by-zero | calculate |
| 6 | 7.005 | 1 | total (7.005) decimal-places | validate |
| 7 | not given | 0 | total (not given) required | validate |
| 8 | 50.00 | 2 | 25.00 | |

Row 7 would also divide by zero, but it is never calculated, so it carries only `required`. Valued 3, Stopped 5, of which 3 at validate and 2 at calculate. Problems: `divide-by-zero` 2, `no-negatives` 1, `decimal-places` 1, `required` 1.

BR-31, by rule. Formula S, 300 rows, `ChunkSize` 100. Row i by i mod 3:

| i mod 3 | total | people | Answer | Stage |
|---|---|---|---|---|
| 1 | -(10.00 x i) | 2 | total (-10.00 x i) no-negatives | validate |
| 2 | 10.00 x i | 2 | 5.00 x i | |
| 0 | 10.00 x i | 0 | divide-by-zero | calculate |

Valued 100 with a sum of answers of 75,250.00, Stopped 200: 100 at validate and 100 at calculate.

## 5. Duplicates

BR-40. The same row twice gives the same answer twice. Formula M, rate 0.005, `ChunkSize` 2, 2 workers: chunks are rows 1 to 2, 3 to 4, 5 to 5. Rows 1 and 2 share a chunk; rows 1 and 5 do not.

| Row | balance | Answer | Stage |
|---|---|---|---|
| 1 | 120.50 | 0.60 | |
| 2 | 120.50 | 0.60 | |
| 3 | -4.00 | balance (-4.00) no-negatives | validate |
| 4 | -4.00 | balance (-4.00) no-negatives | validate |
| 5 | 120.50 | 0.60 | |

120.50 x 0.005 is 0.6025, which rounds to 0.60. The answers are identical apart from the row number. The same holds for a failure at calculate: formula S with total 10.00 and people 0 on two rows gives `divide-by-zero` at stage `calculate` on both.

BR-41. 250 identical rows, balance 120.50, rate 0.005, `ChunkSize` 100. All 250 answers are 0.60, row numbers 1 to 250, Valued 250.

## 6. Totals pushed down

Provisional: everything in this section depends on the language page, which has no declaration syntax for a pushed-down input, and on the batch page's list of what may be pushed down. Formula F; the total goes in as a shared input. The stand-in database in the test holds exact decimals.

Left-over cents: the owner has decided a split's left-over cent is not paid out and stays in the account the payout came from. The library returns each share and nothing else; the test works out the left-over as the payout less the sum of the shares and checks that no row holds it.

| Id | Setup | Answer per row | Sum of shares | Left-over |
|---|---|---|---|---|
| BR-50 | contributions 250.00, 500.00, 250.00; total 1,000.00; payout 400.00 | 100.00, 200.00, 100.00 | 400.00 | 0.00 |
| BR-51 | contributions 100.00 on each of 3 rows; total 300.00; payout 100.00 | 33.33, 33.33, 33.33 | 99.99 | 0.01 |
| BR-52 | 1,000 rows, contribution 10.00 each; total 10,000.00; payout 5,000.00; `ChunkSize` 100 | 5.00 on every row | 5,000.00 | 0.00 |

BR-51: 100.00 x 100.00 / 300.00 is 33.3333..., which gives 33.33 whether rounding goes to the nearest cent or down. 3 x 33.33 is 99.99, so 0.01 stays in the account.

BR-53, aggregates match the rows. Contributions 250.00, 500.00, 250.00, 125.50. The stand-in database and the test, working from the same rows, give: sum 1,125.50, count 4, minimum 125.50, maximum 500.00. Each pair must be equal in value and exact text. Passing the sum in as `total_contributions` gives the same answers as a formula given the list itself (G09 on [batch equals single](batch-equals-single.md)).

BR-54, exact sum. 1,000 contributions of 0.01. The total is 10.00 exactly, not the 9.999999999999831 a binary float gives (EM-02 on [exact money](exact-money.md)). With payout 10.00 each share is 0.01 and the 1,000 shares add to 10.00.

BR-55, average as two pushed inputs. Provisional: average is a pushed sum and a pushed count divided in the formula. Formula `round(contribution - total_sum / member_count, 2)`, contributions 200.00, 300.00, 400.00, `total_sum` 900.00 and `member_count` 3 shared. Answers -100.00, 0.00, 100.00.

BR-56, the limit stated on the batch page. The library cannot check that the total covers the same rows. Formula F with only the rows 250.00 and 500.00, total 1,000.00, payout 400.00: answers 100.00 and 200.00 and no problem. The test's own check finds the rows sum to 750.00, not 1,000.00; matching them is the host's duty.

BR-57, a split that does not divide evenly. 100.00 among 7 members of equal contribution, total 700.00, payout 100.00. Each share is 14.2857..., and by decision 13 each share is cut down to the smallest unit: 14.28 on every row. The 7 shares add to 99.96, which is the payout, and the left-over 0.04 stays in the account the payout came from (EM-31 on [exact money](exact-money.md)). Rounding to the nearest cent would give 14.29 and 7 shares of 100.03, more than the payout, so a split never does that.

## 7. Whole-batch limits

### A refused call

The call returns an error, no row is calculated and no result list comes back.

| Id | Call | Error code | Limit | Got |
|---|---|---|---|---|
| BR-60 | 10,001 rows | `too-many-rows` | 10,000 | 10,001 |
| BR-61 | 3 rows, `ChunkSize` 10,001 | `bad-chunk-size` | 10,000 | 10,001 |
| BR-62 | 3 rows, `Workers` 65 | `bad-workers` | 64 | 65 |
| BR-63 | 3 rows, `ChunkSize` -1; and 3 rows, `Workers` -1 | `bad-chunk-size`; `bad-workers` | 10,000; 64 | -1; -1 |
| BR-64 | 10,001 rows and `Workers` 65 | `too-many-rows` | 10,000 | 10,001 |

Provisional: the row limit is the agency's setting (decision 16) and not a constant of the library; these cases run with an agency setting of 10,000, which is the proposed default, and a call over the setting is refused whole before any row is worked out. The chunk-size and worker limits are the batch page's proposals, and the Limit and Got of BR-63 and the choice of error in BR-64 are proposals (open question 6). A batch of exactly the agency's limit is accepted (BR-09). A case with a smaller agency setting is held until the owner names the ceiling (open question 11).

### Cost limit of the whole batch

Provisional: all of BR-65 to BR-69 wait for the limits, which the engine choice sets. The case uses a batch cost limit of 250 x c, where c is the cost of one row of formula M read from a single run, so the cases do not depend on the engine's units. Rows are by rule R1, 450 rows, 1 worker, so which rows are left unstarted is the same on every run.

Provisional: a chunk is started only while the cost spent is below the limit. A chunk that has started finishes its rows. A row never started carries `batch-limit-reached` at stage `batch` and no value. The call returns no error.

| Id | `ChunkSize` | Chunks started | Rows with a value | Rows with `batch-limit-reached` | Sum of values |
|---|---|---|---|---|---|
| BR-65 | 100 | 3 (spent 100c, 200c, 300c) | 1 to 300 | 301 to 450 (150) | 2,257.50 |
| BR-66 | 1 | 250 | 1 to 250 | 251 to 450 (200) | 1,568.75 |
| BR-67 | 450 | 1 | 1 to 450 | none | 5,073.75 |
| BR-68 | 50 | 5 (250c, equal to the limit) | 1 to 250 | 251 to 450 (200) | 1,568.75 |

In BR-65 the third chunk starts at 200c, below the limit, and runs to 300c; the fourth does not start. Valued 300, NotRun 150, `batch-limit-reached` 150. The rows with a value are kept and named by their row numbers in the result list; the rows not started are named the same way, so nothing is lost or counted twice. The chunk size changes which rows are left unstarted, never the answer of a row that was run (BR-66 and BR-68 give the same 250 rows).

BR-69, a bad row inside the done rows and one in the rows left. As BR-65, but rows 120 and 350 have a negative balance. Row 120 is in a started chunk and carries `balance (-1,200.00) no-negatives` at stage `validate`. Row 350 is never started and carries `batch-limit-reached` at stage `batch`, not a validation problem, because it was never looked at. Valued 299 with a sum of 2,251.50, Stopped 1, NotRun 150. This holds whether or not a row stopped at validate spends any of the limit: chunk 3 starts with at least 199c spent, below the limit, and chunk 4 does not start because the 299 valued rows alone spend 299c.

BR-70, a cost limit with more than 1 worker. Provisional: the same 450 rows with the batch cost limit and 4 workers (there is no time limit, decision 19). Which rows are left is not repeatable with more than 1 worker, so the test does not ask for any particular rows. It asserts: no error; Valued + Stopped + NotRun equals 450; every row either equals its single run or carries `batch-limit-reached` at stage `batch`; no row carries both a value and a problem.

BR-71, a row over its own limit. Provisional: group B of [hostile formulas](hostile-formulas.md), B01 to B05, read as the batch page reads them: one formula, many rows, the hostile row being the hostile input. B01: 100 rows, one hostile at row 57; the other 99 rows carry what they carry alone, row 57 carries its own refusal sentence, the call returns no error, the batch work is bounded. B03: the hostile row is row 1. B04 and B05: one and two rows over the cost limit at run time in the middle, each carrying "This formula goes over the cost limit of {value}." at stage `calculate`; the rows around them carry values. The formula and the hostile input are those of the hostile page (for example C02); nothing is invented here.

## 8. Cancel and restart

The library keeps no state between calls, so "restart" is the host calling again. Provisional: whether and how the library honours a host's own cancel or deadline is open (decision 19; open question 8). If it does, a cancel stops the batch at the next chunk edge, rows not started carry `batch-cancelled` at stage `batch`, and the call returns no error. BR-72 to BR-77 follow that proposal and wait for the owner's answer. The test cancels in a fixed place with a test-only hook (open question 7). Rows by rule R1, 450 rows, `ChunkSize` 100, 1 worker.

BR-72, cancelled before the call. Every one of the 450 rows carries `batch-cancelled`. Valued 0, NotRun 450. Called again with a live context, the 450 rows give 0.05 x i, summing to 5,073.75, exactly as a call never preceded by a cancel.

BR-73, cancelled after chunk 2 finishes. Rows 1 to 200 carry 0.05 x i, summing to 1,005.00. Rows 201 to 450 carry `batch-cancelled`. Valued 200, NotRun 250. Then two ways to go on:

| Way | What is run | Result |
|---|---|---|
| Run all again | The 450 rows | The same 450 answers as an uncancelled run, row by row. Sum 5,073.75 |
| Run the rest | Rows 201 to 450, renumbered 1 to 250 | Row j carries 0.05 x (200 + j), summing to 4,068.75. Joined to the first call: 200 + 250 = 450 rows valued, each original row once, sum 1,005.00 + 4,068.75 = 5,073.75 |

BR-74, counted twice. Run the rest of BR-73 a second time and add its rows again: the total is 9,142.50, which is wrong. The final total must be 5,073.75, with every original row valued once. The library returns the same answers both times; keeping them from being added twice is the host's rule (open question 1). The test's own join of two calls rejects a row that appears twice.

BR-75, a bad row among the rows left. As BR-73 with row 350 having a negative balance. In the first call row 350 carries `batch-cancelled`, not a problem. In the call that runs the rest it carries `balance (-3,500.00) no-negatives` at stage `validate`. Joined, the 450 rows equal an uncancelled run of the same rows: 449 values and 1 problem.

BR-76, more than 1 worker. Provisional: which rows are left unstarted is not repeatable, so the test asserts only: row numbers 1 to 450 each exactly once; every row either has its value 0.05 x i or `batch-cancelled`; the two groups do not overlap; and after the rows left are run again and joined, all 450 rows equal an uncancelled run.

BR-77, a deadline instead of a cancel, with 1 worker and a deadline that passes after chunk 2. Same as BR-73: rows 201 to 450 carry `batch-cancelled`.

## 9. Memory

A call does not hold more than the amounts below at once. The rows in work are those a worker has taken and not finished; each worker holds one chunk.

| Id | Case | Expected |
|---|---|---|
| BR-80 | 10,000 rows by rule R1, `ChunkSize` 100, 1 worker; then 4 workers; `ChunkSize` 1 with 64 workers; `ChunkSize` 10,000 with 1 worker | Most rows in work at once: 100, 400, 64 and 10,000. In general never more than `Workers` x `ChunkSize`, and never more than the row count. Provisional: counting rows in work needs a test-only hook |
| BR-81 | 10,000 rows, each with 2 problems (balance -10.00 and rate 0.00505), `ChunkSize` 100, 4 workers | Stopped 10,000, 20,000 problems; the memory held beyond the input rows and the result list is at most 32 MiB (proposed). Provisional: the figure waits for the engine choice |
| BR-82 | After BR-80 and BR-81 have returned and a collection has run | Heap in use within 1 MiB (proposed) of its size before the call; goroutine count the same as before |
| BR-83 | 100 calls of 10,000 rows, the host splitting 1,000,000 rows: row i of call c is balance 10.00 x ((c - 1) x 10,000 + i) | Heap in use after call 100 within 1 MiB (proposed) of after call 1. Sum of all answers 25,000,025,000.00 |
| BR-84 | A refused call, 10,001 rows (BR-60) | Heap in use after the call within 1 MiB (proposed) of before; no result list was built |
| BR-85 | A cancelled call (BR-73) | Same goroutine count and heap as BR-82 after it returns |

## 10. Every row equals the same row run alone

Every case above is run through the test on [batch equals single](batch-equals-single.md) as it stands: after the batch, each row is run alone with `Run`, with the shared inputs merged under the row's own values, and compared on value (type and exact text), problems (count, order, input, received value, rule, sentence, code) and stage. For a case of 10,000 rows, all 10,000 rows are compared, not a sample.

Rows that carry `batch-limit-reached` or `batch-cancelled` are exempt, because a single run never meets a batch limit. Instead the test checks that such a row carries no value, stage `batch`, and exactly one problem. Every other row of BR-65 to BR-77 is compared, including the rows that were run before the limit or the cancel.

Neither the chunk size nor the number of workers may change the answer of a row that was run. Only which rows are left unstarted by a limit may change with them (BR-65 against BR-66).

## Open questions

The owner has not decided any of these. Each has a proposed answer.

1. When a run is stopped halfway, are the rows already done kept and named, and does the restart run only the rows left? Proposed: yes; the library keeps no state, the host keeps each finished row once and posts it once, and restarts with the rows that carry `batch-cancelled` or `batch-limit-reached`. This is the rule the owner reviews with the treasurer's month-end moment (todo TD-2988); BR-72 to BR-77 follow this proposal and wait for it.
2. Settled by decision 13: the shares of a split are cut down to the smallest unit and the left-over stays in the account the payout came from. BR-57 is run on that rule. The number is kept so the other references hold.
3. [Laws](laws.md) draws 1 to 200 rows per batch for the random run, and this page uses 10,000. Raised by the batch page, not settled here. Proposed: keep 200 for the random run and use the fixed cases at 10,000 and 10,001 on this page.
4. The hostile page describes a batch as rows each holding a formula; the batch page runs one formula over many rows. Raised by the batch page, not settled here. This page uses the batch page's reading (BR-71).
5. Does "the limit has been reached" mean spent is at least the limit? Proposed: yes, as BR-65 and BR-68 assume. If it means more than the limit, one more chunk starts in each.
6. Which Limit and Got does a refused call report for a value below the range (BR-63), and which error wins when two are wrong at once (BR-64)? Proposed: the upper limit and the number received; rows first, then chunk size, then workers.
7. May the test build have hooks to hold a chunk, cancel at a chunk edge, count rows in work and read the cost of one row? Proposed: yes, compiled only into tests, never into the library a host imports.
8. Is a host's own cancel or deadline honoured at all? Decision 19 leaves it open. The batch page lets a started chunk finish after a cancel; the hostile page (L09) waits on the same answer. Proposed: yes, and a started chunk finishes its rows, each bound by its own cost limit, as the batch page says.
9. What stage does a row that goes over its own cost limit at run time carry? Proposed: `calculate`, and it follows the step's failure choice (decision 19).
10. Are 32 MiB beyond the rows and results for 10,000 rows, and 1 MiB of drift over 100 calls, the right memory limits? Proposed: yes as a start; the engine choice sets the real figures.
11. What is the ceiling above which no agency may set its rows per batch, and who cuts a run larger than the agency's limit into several calculation batches, the platform or the builder? (decisions 16 and 17; the agency steps decision 11). Proposed for the ceiling: 10,000, the batch page's proposed default. No proposal is made for who cuts the run.

## How the figures were checked

Every figure was worked by hand, then re-checked with a throwaway script using exact decimals, kept outside the repo. Rounding there was to the nearest cent unless a case says down, and the script confirmed that no figure of formula M falls on a half cent, so no case depends on the rounding mode except BR-57.
