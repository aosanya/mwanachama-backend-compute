# The speed and many-callers suite

Built for ISS-36, in the folder `concurrency_test/` (package `concurrency`, test files only). It runs the cases of [concurrency-and-speed.md](../4.%20qa/concurrency-and-speed.md) on `formula/`, `batch/` and `library/`: many callers on one compiled program, and speed that is measured, recorded and judged. The targets it carries are on [concurrency-and-speed.md](concurrency-and-speed.md) in this folder. The decisions are 9 (bulk runs), 17 (two stages) and 19 (a cost limit and no clock) in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. Every target, margin and count is proposed. All names and figures are invented.

## What the folder holds

| File | What it is |
|---|---|
| `fixtures_test.go` | the nineteen example formulas with their worked example and result, and an agency runtime that serves `late_days@1` and `interest_step@1` as programs |
| `generate_test.go` | a seeded stream of inputs for formulas 1, 3, 5, 6, 7, 8 and 15 |
| `sharing_test.go` | CS-01 to CS-06: one program, 64 callers |
| `batchsharing_test.go` | CS-07 to CS-14: batches, workers, chunks, the limits and the refusal |
| `state_test.go` | CS-13: no package on the run path writes a package-level variable |
| `speedcases_test.go` | the 35 speed cases, the benchmarks and the proposed targets |
| `judge_test.go` | the margins, the median and spread of five counts, and the judging of a slip and of a target |
| `machine_test.go` | the machine, and the one-minute load average |
| `speedcheck_test.go` | recording a baseline, the speed check, the target check and the proof of a slowdown |
| `baselines.json` | the recorded figures, one entry for each machine |
| `Makefile` | `test`, `speed-check`, `speed-targets` and `baseline` |

Imports: the standard library, `formula/`, `batch/`, `function/`, `library/`, `number/`, `numtext/`, `datefn/` and `problem/`. The folder reads no clock except to date a baseline.

## Commands

| Command | What it does |
|---|---|
| `make -C concurrency_test test` or `go test -race ./concurrency_test/` | the many-callers cases, the judging rules, and one pass of every speed case to show each runs. No time is judged. It also runs in `make test` |
| `make -C concurrency_test baseline` | measures all 35 cases and writes the figures for this machine into `baselines.json` |
| `make -C concurrency_test speed-check` | measures again and fails when a figure is over its baseline by more than the margin |
| `make -C concurrency_test speed-targets` | measures and fails when a figure is over its proposed target by more than the margin |

The root `Makefile` has no line for these. It carries another builder's uncommitted edits, so none was added. The line to add is `speed-check: ; $(MAKE) -C concurrency_test speed-check`.

Environment: `SPEED_CHECK=1` turns on the timed tests (the make targets set it), `SPEED_RECORD=1` records, `SPEED_FORCE=1` measures even when the machine is loaded, and `SPEED_BENCHTIME` (default 300ms) is the time of one count.

## The many-callers cases

| Case | State | Test |
|---|---|---|
| CS-01 | built | 64 callers start together, each with 2,000 inputs (200 under `-short`), on one program of each of formulas 1, 3, 5, 6, 7, 8 and 15; each result equals the single run. A second test holds the nineteen worked examples against the results on the page |
| CS-02 | built | the nineteen worked examples and 2,000 generated inputs of each of the seven formulas keep their inputs, compared by an exact key before and after |
| CS-03 | built | formulas 6 and 8, with their callee programs shared by 64 callers; `interest_step@1` is also called directly from 64 goroutines |
| CS-04 | built | each example run 200 times in one process, and once in a second process (the test binary started again) |
| CS-05 | partly built | 32 goroutines compile the same text at once and every program has the same cost, the same standard text and the same result. **Held:** "exactly one program is built" |
| CS-06 | partly built | two pinned versions, `fin.interest_step@1` and `@2`, in one registry, alternated by 64 goroutines; neither gets the other's result, and the test refuses to pass if the two never differ on its inputs. **Held:** the count of cache entries |
| CS-07 | built | 2,000 rows (values, input problems and a division by zero), 1 and 8 workers, chunk sizes 1, 7, 2,000 and 2,001: the same result list, and every 97th row equals its run alone |
| CS-08 | built | the same rows twice with 8 workers give the same list |
| CS-09 | built | 10,001 rows with the default row setting are refused whole with `too-many-rows`; a counting function shows no row was worked and no goroutine is left over |
| CS-10 | built | 10,000 rows, chunk size 1, one worker, a cost limit of 25 rows: exactly the first 25 rows have a value, the rest carry `batch-limit-reached` at stage `batch`, and the counts add up |
| CS-11 | built | the same with 8 workers: at least 25 rows started, every started row equals its single run, every other row is not started |
| CS-12 | built | 5,000 rows and 8 workers: all 5,000 rows were worked when the call returned, and the goroutine count is not above the count before |
| CS-13 | built | three runs (two with one worker, one with 8) give the same list; a read of the source finds no write to a package-level variable outside `init` in nine packages, and a test shows the reader sees a write |
| CS-14 | held | whether a host's own cancel is honoured is open under decision 19 |

Things that differ from the QA page:

- The library keeps no cache of compiled programs, so the cache cases (the "exactly one program" half of CS-05 and the entry count of CS-06) have nothing to count. They are skipped tests with that reason. A host that keeps a cache owns them.
- The registry needs a library name, so CS-06 uses `fin.interest_step`. The page names `interest_step@1` and `@2`, which are agency functions; the agency functions of CS-03 are programs run through `Program.Apply`.
- CS-12 reads the goroutine count before and after and fails only when the count is higher after. Goroutines of earlier tests may still be ending, so "equal" would fail by chance.
- The goroutines of CS-09 and CS-12 are counted with `runtime.NumGoroutine`; the tests of the folder never run in parallel with each other.

## The speed cases

35 cases: one run of each of the 19 formulas (CS-20), a batch of 1,000 rows with one worker for formulas 1, 8, 3 and 15 (CS-21), a batch of 100,000 rows with 1 and with 8 workers for the same four (CS-22), and one compile of formulas 1, 8, 15 and 3 (CS-23). A run uses the first example's inputs. A batch goes through `batch.RunBatch` with a chunk size of 100: every row is read from text, validated, run and written to its own slot, and the figure is per row. A compile is `Parse` and `Plan`, which calls `Check`.

Every figure is the median of five counts of `testing.Benchmark`, with the spread of the five ((largest minus least) over the median). A case whose spread is over 25 percent is measured once more and the one with the smaller spread is kept. Bytes and allocations come from the same counts.

### How a figure is judged

| Rule | Value |
|---|---|
| Slip against the baseline | time over the baseline by more than 25 percent; bytes or allocations over it by more than 10 percent |
| Miss against the target | time over the target by more than 25 percent; allocations over the target by more than 10 percent (a target of zero is no target) |
| The median target | the median of the 19 runs, at most 450 ns plus 25 percent |
| Machine | a baseline is found by the machine's name (processor, CPU count, system, Go version); no baseline is used on another machine |
| Quiet | the one-minute load average at the start must be at most half of the CPU count (5.5 on this machine); otherwise the timed tests say they refuse to judge and skip, and `SPEED_FORCE=1` overrides it |
| A baseline recorded under load | marked `quiet: false`; the speed check skips against it and says to record it again |

The judging rules are plain functions with tests that run in `make test`: a figure at the margin passes, one just over it fails, a doubled time fails, a rise of 11 percent in bytes or in allocations fails, and formula 1 has no batch target. The slowdown proof, `TestSpeedCheckFailsOnADeliberateSlowdownInACopy`, measures the program of formula 3, then a copy that runs the same program three times for each iteration, and requires the check to report the slip. It did: 1,929 ns became 8,780 ns, and 92 allocations became 276.

## The figures recorded

Measured on 2026-11-17 by `make baseline` with `SPEED_FORCE=1`:

| Part | Value |
|---|---|
| Processor | Apple M3 Pro, 11 logical CPUs (5 performance, 6 efficiency) |
| System | darwin/arm64, Go go1.27.3, GOMAXPROCS 11 |
| One-minute load average at the start | 9.4 (the limit for a quiet machine is 5.5) |
| Count time | 300 ms, 5 counts |

Other agents were building and testing on the machine, so the file is marked not quiet and the speed check will not judge against it. The allocation figures do not depend on load; the time figures do (formula 1 gave 617 ns at a load of 27 and 543 ns at 9.4). Time per unit, bytes per unit and allocations per unit; target is the proposed target of [concurrency-and-speed.md](concurrency-and-speed.md) (ns, and allocations where there is one).

| Case | ns | Bytes | Allocs | Spread | Target |
|---|---|---|---|---|---|
| CS-20 f1 run | 543 | 808 | 21.0 | 10% | 1,500 ns and 25 |
| CS-20 f2 run | 1,042 | 1,936 | 52.0 | 14% | 1,500 ns and 25 |
| CS-20 f3 run | 1,825 | 3,376 | 92.0 | 12% | 1,500 ns and 25 |
| CS-20 f4 run | 977 | 1,776 | 37.0 | 8% | 1,500 ns and 25 |
| CS-20 f5 run | 311 | 920 | 6.0 | 5% | 1,500 ns and 25 |
| CS-20 f6 run | 3,031 | 5,424 | 66.0 | 12% | 1,500 ns and 25 |
| CS-20 f7 run | 470 | 680 | 19.0 | 2% | 1,500 ns and 25 |
| CS-20 f8 run | 1,131 | 1,608 | 40.0 | 6% | 1,500 ns and 25 |
| CS-20 f9 run | 1,391 | 2,104 | 52.0 | 18% | 1,500 ns and 25 |
| CS-20 f10 run | 639 | 584 | 6.0 | 2% | 1,500 ns and 25 |
| CS-20 f11 run | 1,120 | 1,984 | 37.0 | 8% | 1,500 ns and 25 |
| CS-20 f12 run | 2,131 | 3,264 | 67.0 | 56% | 1,500 ns and 25 |
| CS-20 f13 run | 1,002 | 1,296 | 31.0 | 2% | 1,500 ns and 25 |
| CS-20 f14 run | 1,573 | 2,488 | 55.0 | 33% | 1,500 ns and 25 |
| CS-20 f15 run | 1,134 | 1,792 | 37.0 | 28% | 1,500 ns and 25 |
| CS-20 f16 run | 5,575 | 5,352 | 161.0 | 63% | 1,500 ns and 25 |
| CS-20 f17 run | 2,833 | 1,672 | 47.0 | 73% | 1,500 ns and 25 |
| CS-20 f18 run | 1,386 | 952 | 4.0 | 24% | 1,500 ns and 25 |
| CS-20 f19 run | 434 | 312 | 3.0 | 22% | 1,500 ns and 25 |
| CS-21 f1 1,000 rows workers 1 | 2,107 | 4,340 | 55.0 | 51% | none |
| CS-21 f3 1,000 rows workers 1 | 7,023 | 9,813 | 196.8 | 19% | 725 ns |
| CS-21 f8 1,000 rows workers 1 | 2,560 | 4,676 | 59.5 | 29% | 225 ns |
| CS-21 f15 1,000 rows workers 1 | 3,899 | 7,387 | 107.0 | 9% | 675 ns |
| CS-22 f1 100,000 rows workers 1 | 2,019 | 4,336 | 55.0 | 4% | none |
| CS-22 f1 100,000 rows workers 8 | 640 | 4,336 | 55.0 | 19% | none |
| CS-22 f3 100,000 rows workers 1 | 6,009 | 9,810 | 196.8 | 15% | 725 ns |
| CS-22 f3 100,000 rows workers 8 | 1,598 | 9,810 | 196.8 | 12% | 425 ns |
| CS-22 f8 100,000 rows workers 1 | 2,198 | 4,679 | 59.5 | 4% | 225 ns |
| CS-22 f8 100,000 rows workers 8 | 654 | 4,679 | 59.5 | 16% | 150 ns |
| CS-22 f15 100,000 rows workers 1 | 3,629 | 7,384 | 107.0 | 11% | 675 ns |
| CS-22 f15 100,000 rows workers 8 | 1,032 | 7,384 | 107.0 | 19% | 400 ns |
| CS-23 f1 compile | 2,927 | 4,264 | 134.0 | 4% | 2450 ns and 45 |
| CS-23 f3 compile | 4,123 | 5,440 | 176.0 | 11% | 3850 ns and 70 |
| CS-23 f8 compile | 3,239 | 3,656 | 113.0 | 2% | 2525 ns and 55 |
| CS-23 f15 compile | 3,474 | 4,568 | 145.0 | 2% | 2850 ns and 60 |

The median over the nineteen runs is 1131 ns (target 450).

## What the figures show against the proposed targets

`make speed-targets` with `SPEED_FORCE=1`, at a load of about 9, gave 34 findings, and the page's targets are not met by this build. The targets were set from the spike's engine in `spike/`, which is a different implementation, and the batch rows of the spike did not read text or validate.

- **One run (CS-20):** 13 of the 19 formulas are over 25 allocations (the largest is formula 16 with 161); only formulas 1, 5, 7, 10, 18 and 19 are within. In the recorded figures formulas 3, 6, 12, 14, 16 and 17 are over 1,500 ns, and the median is 1,131 ns against 450. The allocation findings do not depend on load.
- **A batch row (CS-21, CS-22):** every case with a target is over it, by 5.4 to 11.4 times with one worker (formula 8 about 2,200 to 2,600 ns against 225; formula 3 about 6,000 to 7,000 against 725; formula 15 about 3,600 to 3,900 against 675) and by 2.6 to 4.4 times with 8 workers. The speed-up with 8 workers is 3.2 to 3.8 times.
- **A compile (CS-23):** time is 7 to 28 percent over the target before the margin, and allocations are 2.1 to 3.0 times it (134, 113, 145 and 176 against 45, 55, 60 and 70).

Why is not measured here. Two causes are likely and both belong to the owner's reading: the targets came from another engine, and a real batch row also reads its text with `numtext` and validates it, which the spike's per-row figure did not. This suite does not split them. Until the owner decides, the target check stays its own command and the speed check, which compares a build with itself on one machine, is what guards against a slip.

## Where this differs from the QA page

- **Baseline:** the QA page says a check compares against a baseline measured on the same machine in the same job, just before. Comparing two commits in one job needs two builds. This suite keeps a baseline per machine in `baselines.json` and compares only on the same machine; a baseline is refreshed by `make baseline` on a quiet machine. The same-job form is the owner's choice.
- **Two checks:** the QA page has one check that fails a slip and a miss together. Here they are `speed-check` (a slip) and `speed-targets` (a miss), because the targets are not met today and one red check would hide a slip.
- **The speed cases are benchmarks and tests together:** `go test -bench BenchmarkSpeed ./concurrency_test/` runs the same 35 cases with `-benchmem`.

## What stays open

- A quiet machine. The baseline must be recorded again at a load of 5.5 or less, and the targets re-read (open question 7 of the QA page).
- Whether the targets are right for this build, or the build is to meet them (above).
- The baseline form (per machine, or in the same job) and a machine the team keeps for it (open question 6).
- Whether the speed check runs on every change or before a release (open question 3). It is not in `make test`.
- CS-14, the cancel, and the cache half of CS-05 and CS-06.
- The batch cases of 1,000 and 100,000 rows carry the per-row targets of 10,000 rows (open question 8), and no batch of formula 1 has a target.
- The line for the root `Makefile`, and the row for this page in the implementation README.
