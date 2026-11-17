# Concurrency and speed

Status: the owner has not reviewed this page. Every number marked "proposed" is a starting point. Every target figure is marked "to be set" until the owner chooses the engine (the engine spike, ISS-4). All names and figures are invented; the repo is public.

## What this page protects

A host runs one compiled formula for many callers at once, and a month-end run works thousands of rows in one call. Two promises follow, and a change that breaks either must be caught before it ships:

1. Sharing is safe. Many callers running one compiled program at the same moment each get exactly what a single caller gets.
2. Speed does not slip. A change that makes a run, a batch or a compile slower than its target by more than a stated margin fails a check.

This page designs the cases. It builds nothing. The cases that need the engine's own figures say "to be set" and are filled in when the engine page exists (todo TD-2980 of the same issue).

## How the cases are read

- A case id starts with `CS-` (concurrency and speed).
- Setup says what is run. Pass rule says what must be true. Judged by names the figure or the comparison the rule rests on.
- A case marked **Held** cannot be written yet because a decision it needs is open. The reason is given.
- Formula numbers are the formulas of [example formulas](../1.%20requirements/example-formulas.md). Formula 1 is the loan balance, 3 the fund share per member (two lists of decimals), 8 the monthly interest (calls the library function `interest_step@1`) and 15 the capital call share (a decimal and a list of integers).

## Facts the cases rest on

All of these are from the engine comparison spike (`spike/`). They are measurements, not targets.

### The machine

| Part | Value |
|---|---|
| Processor | Apple M3 Pro, 11 logical CPUs (5 performance, 6 efficiency) |
| Memory | 18 GB |
| System | macOS 15.7.3, arm64 |
| Go | as recorded in `spike/results/machine.txt` |
| Command | `go test -c ./measure/`, run with `-test.bench . -test.benchmem -test.count=5`, `GOMAXPROCS` equal to the CPU count |

Other agents ran tests and builds on the same machine during both runs. The one-minute load average was 6.8 to 29.2 (median 13.3) in the first run and 4.5 to 19.3 (median 6.0) in the second, on 11 CPUs. The spread of five runs of one case reached over 100 percent for a single run of one formula and was often 2 to 20 percent. So the figures below carry about 20 percent noise and are never to be read as targets. The tables below are from the second run (`spike/results/bench-run-2.txt`); the raw output of both runs is in `spike/results/`.

### What was measured, formula 8 (a library call), medians

| Engine | One run, ns | One row in a batch of 10,000, one worker, ns | One row, 8 workers, ns | One compile, ns |
|---|---|---|---|---|
| cel-go | 1,076 | 397 | 154 | 82,034 |
| expr-lang | 350 | 326 | 123 | 15,735 |
| own | 146 | 142 | 61 | 1,680 |

### The same per row, formulas 3 and 15 (decimal heavy), batch of 10,000, one worker then 8 workers

| Engine | Formula 3, 1 worker, ns | Formula 3, 8 workers, ns | Formula 15, 1 worker, ns | Formula 15, 8 workers, ns |
|---|---|---|---|---|
| cel-go | 1,069 | 313 | 634 | 207 |
| expr-lang | 637 | 200 | 526 | 191 |
| own | 475 | 180 | 434 | 166 |

### What the spike already showed about sharing

64 goroutines each ran 2,000 inputs of formulas 3, 5, 6, 7, 8 and 15 on one compiled program, under the race detector, and every result equalled the result of the same input run alone. The same was done for the library programs shared by their callers. A run did not change its inputs. These are the structure of cases CS-01 to CS-04.

## Concurrency cases

| Id | Setup | Pass rule | Judged by |
|---|---|---|---|
| CS-01 | One compiled program of each of formulas 1, 3, 5, 6, 7, 8 and 15. 64 goroutines start together, each with its own 2,000 inputs, under `go test -race` | Every result equals the result of the same input run alone, in the same type and exact text. The race detector reports nothing | Equality with the single run; race detector clean |
| CS-02 | Run every example of every formula and compare the inputs before and after the run | The inputs are unchanged, whatever the formula does | Equality of an exact key of the inputs, before and after |
| CS-03 | A formula that calls a library function (formulas 6 and 8), its library program shared by all callers, 64 goroutines each with its own inputs for the caller and the callee | Every caller result equals the single run. Calling the library function directly from the goroutines at the same time also equals the single run | Equality with the single run; race detector clean |
| CS-04 | The same formula run 200 times in one process and once in a second process | Same output every time (law 5 on [laws](laws.md)) | Equality of exact text |
| CS-05 | Compile the same formula text twice at the same moment from 32 goroutines | Exactly one compiled program is built and every caller gets it. No caller waits on a lock held for a run | Count of compiles seen by a counting hook; race detector clean |
| CS-06 | Two publications pin different versions of one library function (`interest_step@1` and `interest_step@2`, see [library versioning](library-versioning.md)). 64 goroutines alternate between them | A call for version 1 never gets a program built for version 2 and the other way round, and the cache holds one entry per name and version | Result equality with a single run of the named version; entry count of the cache |
| CS-07 | A batch of 2,000 rows (a mix of valid rows and rows with problems, as in [batch equals single](batch-equals-single.md)) run with 1 worker and with 8 workers, chunk sizes 1, 7, the row count and the row count plus 1 | Same rows, same order, same answers and same problems per row, whatever the number of workers and the chunk size | Equality of the whole result list |
| CS-08 | The same 2,000 rows run twice with 8 workers | The same result list both times, though the workers finish in a different order | Equality of the whole result list |
| CS-09 | A batch over the agency's row setting (10,001 rows with the proposed default of 10,000) | Refused whole with `too-many-rows` before any row is worked out; no row is calculated and no goroutine started for rows | Error code; count of rows worked is 0 |
| CS-10 | A batch of 10,000 rows whose chunk is started while the batch cost limit is reached (a cost limit set low enough for the test), 1 worker | The rows never started carry `batch-limit-reached` with stage `batch`; the rows started are the first ones in order | Row codes; the number of rows with a value plus the rows stopped plus the rows not started equals the row count |
| CS-11 | The same case as CS-10 with 8 workers | The answers of the rows that were started still equal their single runs. The test does not depend on which rows were left unstarted | Equality per row for the started rows only |
| CS-12 | A batch of 5,000 rows, then the goroutine count is read after the call returns | The call returns only when the whole calculation batch has finished (decision 17, the calculation stage finishes as a whole), and no goroutine started by the call is still running | Goroutine count after the call equals the count before it, after the call has returned and nothing else is waited for |
| CS-13 | A run of the calculation stage changes nothing real: the same batch is run twice with the same inputs, then once more with 8 workers | The three result lists are equal, and nothing the library holds has changed between runs (it holds nothing) | Equality of the result lists; the package holds no package-level state a run can write to |
| CS-14 | Cancelling a batch, with a host cancel signal sent while 8 workers work | **Held.** Whether and how the library honours a host's own cancel is open (decision 19, open question 10 on [batch evaluation](../2.%20design/batch-evaluation.md)). If it does, the rows not started carry `batch-cancelled` with stage `batch`, the call returns promptly and leaves no work running | To be set once the question is decided |

Notes on the concurrency cases:

- CS-01 to CS-03, CS-05 and CS-06 run under the race detector. A race report fails the case even when every result is right.
- No concurrency case measures time. A case that passes or fails depends on the input and the code, never on how fast the machine is (decision 19: the library has no time limit of its own, so a test must not pass and fail on the same input).
- CS-10 and CS-11 depend on the batch cost limit, whose value is to be set by the engine choice. Until then the test uses a small invented limit.

## Speed cases

Speed cases are benchmarks, not tests. They run in a separate check (proposed: `make bench-check`) and never in `make test`, because a figure depends on the machine.

| Id | Setup | Reported | Target |
|---|---|---|---|
| CS-20 | One run of one compiled program, first example inputs, for each formula class below | Median nanoseconds, bytes and allocations per run of 5 counts | to be set |
| CS-21 | One batch of 1,000 rows, one worker | Median nanoseconds, bytes and allocations per row | to be set |
| CS-22 | One batch of 100,000 rows, one worker, then 8 workers. The agency's row setting is raised for the test above its default (the ceiling is an open question) | Median nanoseconds, bytes and allocations per row, and the ratio of the 8 workers figure to the 1 worker figure | to be set |
| CS-23 | One compile of each formula class (parse, check and build one program) | Median nanoseconds, bytes and allocations per compile | to be set |

The formula classes, used in CS-20 to CS-23 and chosen from the spike:

| Class | Formula | Why |
|---|---|---|
| Plain | 1, the loan balance, `payment >= balance ? 0.00 : balance - payment` | One comparison and one subtraction on decimals |
| Decimal heavy | 3, the fund share per member, and 15, the capital call share | Lists, sums, a division and a rounding on exact decimals |
| Calls the library | 8, the monthly interest, `round(balance * interest_step@1(balance), 2)` | A call to a published function by name and version |

Rules for every speed case:

1. Every figure is the median of 5 counts (`-test.count=5`), reported with the spread of the 5. A case whose spread is over 25 percent (proposed) is run again before it is judged.
2. Every figure is stored together with the machine it was measured on (the table above, with the Go version and the date). A figure measured on one machine is never compared with a figure from another machine.
3. A check compares a run against a baseline that was measured on the same machine in the same job, run just before it. It never compares against a figure from a table written on another machine.
4. The machine is quiet: the check records the one-minute load average at its start and refuses to judge when the load is above half the CPU count (proposed), and says so, instead of passing or failing.
5. Memory is judged as well as time: bytes and allocations per run are reported, and a rise over the margin fails the same way a rise in time does.

## The target and the margin

| Rule | Value | Status |
|---|---|---|
| Target for one run, per class | to be set from the engine choice page | To be set |
| Target for one row of a batch, per class, 1 worker and 8 workers | to be set from the engine choice page | To be set |
| Target for one compile, per class | to be set from the engine choice page | To be set |
| Margin: a figure over its target by more than this fails the check | 25 percent | Proposed. The spike's own noise was about 20 percent on a busy machine, so a margin below 20 would fail on noise |
| Margin for bytes and allocations per run | 10 percent | Proposed. These figures do not depend on load, so the margin can be tighter |
| Count of runs a figure is the median of | 5 | Proposed |

A run over the target by more than the margin fails the check and says which class, which figure, the target, the figure measured and the machine.

## What the cases do not cover

- Whether one engine is faster than another. That is the engine choice page.
- Row-by-row equality with a single run, on its own. That is [batch equals single](batch-equals-single.md) and law 4 on [laws](laws.md). CS-07 and CS-08 only add the number of workers and the chunk size to it.
- The cost limit as a refusal of a hostile formula. That is [hostile formulas](hostile-formulas.md).
- How long a host keeps a result or a summary. That is the host's rule.

## Open questions

The owner has not decided any of these. Each has a proposed answer.

1. The target figures themselves. They wait for the engine choice. Proposed: a target for each class is the median measured for the chosen engine on a quiet machine, rounded up to 2 significant figures, and the margin above is added to it.
2. The margin, 25 percent for time and 10 percent for memory. Proposed as stated above.
3. Whether the speed check runs on every change or only before a release. Proposed: before a release and on a nightly run, never in `make test`.
4. Whether a host's own cancel is honoured, and then what "promptly" means without a clock (decision 19). Proposed: say how many rows may still finish after the signal, as a count and not a time, and hold CS-14 until the owner answers.
5. The 100,000-row batch of CS-22 is above the proposed default of 10,000 rows per call. Proposed: the test raises the agency's setting for that run only, and the ceiling an agency may set stays open on [batch evaluation](../2.%20design/batch-evaluation.md).
6. A machine for the check that is the same every time, for example a machine the team keeps for this. Proposed: the check records the machine and refuses to compare across machines (rule 2 above), and a fixed machine is chosen later.
