# Speed findings: where the built engine spends its allocations

Status: partial. Written for ISS-83 and stopped early at the owner's request. Todos 1 and 2 (profile a run, profile a batch) are done and are on this page. Todo 3 (run the spike's evaluator beside the built one) was not started, and the candidate changes below are sized from the profiles and were not tried. Every figure is invented-input data on one machine. The proposed targets are on [concurrency-and-speed.md](concurrency-and-speed.md); the suite that found the miss is [speed-suite.md](speed-suite.md).

No engine code was changed. Nothing here changes a target.

## The machine and its load

Apple M3 Pro, 11 logical CPUs, darwin/arm64, Go go1.27.3, 2026-11-17. Other agents were building on the machine. The one-minute load average was 87 at the start of the session, 23 at the first profile and about 5 at the batch and compile profiles. Time figures are not claimed as exact. The allocation counts do not depend on load, so this page leans on them.

## What was profiled

| Profile | Case | Count |
|---|---|---|
| Allocation sites, every allocation recorded | one run of formula 8 and of formula 16 (CS-20) | 300,000 runs each |
| Allocation sites, every allocation recorded | a batch of 100,000 rows of formula 8 with one worker (CS-22) | 4 batches |
| Allocation sites, every allocation recorded | one compile of formula 8 (CS-23) | 100,000 compiles |
| CPU | the same batch | 20 batches |

The profile files are in `/tmp/iss83/` and are not kept in the repo.

## One run: nearly all of it is decimal arithmetic

Allocation counts per run match the suite: formula 8 gives 40 allocations and 1,608 bytes, formula 16 gives 161 and 5,352.

| Where | Formula 8 | Formula 16 |
|---|---|---|
| `number.aligned` and what it calls (cumulative share of allocations) | 41.8 percent, about 17 of 40 | 85.4 percent, about 137 of 161 |
| of that, `number.pow10` | 28.3 percent, about 11 | 40.9 percent, about 66 |
| `math/big` `nat.make`, the slices behind every big number | 40.3 percent | 44.1 percent |
| `Program.start`, `NewMeter`, `runEvaluate`, `RunMetered`, `Apply` (flat, the per-run state) | about 24 percent, about 10 | under 5 percent |
| `Decimal.Round` through `divideRounded` | 11.9 percent | 2.5 percent |

`aligned` is called by `Add`, `Sub` and `Cmp` in `number/decimal.go`. It multiplies both numbers by `pow10(scale difference)` every time, and `pow10` builds a new power of ten with `big.Int.Exp` every time, even when the scales are equal and the power is 1. So each addition, subtraction and comparison allocates about six big-number values before it adds anything. Formula 16 does many of them.

If `aligned` allocated nothing, the arithmetic above would leave about 23 allocations for formula 8 and about 24 for formula 16. That is an upper bound from the shares and not a result, but it is at the proposed target of 25 for both.

The CPU profile of a single run was not usable: at 1.5 million iterations it is dominated by the runtime starting the world and returning memory, not by the run. No time share is claimed for a single run.

## One batch row of formula 8 (59.5 allocations, 4.7 KB)

Shares of allocation counts, cumulative, for a batch of 100,000 rows with one worker. They overlap, so they do not add to 100 percent.

| Where | Share | Per row |
|---|---|---|
| `Program.RunMetered`, the run itself | 65.2 percent | about 41 |
| `function.Validator.Validate` | 21.7 percent | about 14 |
| `Compiled.valuesOf`, which reads each input again | 9.9 percent | about 6 |
| `number.aligned` (inside the run) | 27.7 percent | about 17.5 |
| `number.pow10` (inside `aligned`) | 18.8 percent | about 12 |
| `number.ParseDecimal` (in Validate and in `valuesOf`) | 17.8 percent | about 11 |
| `Decimal.Round` | 15.8 percent | about 10 |
| `function.rulesOf` (flat, rebuilt on every Validate) | 5.9 percent | about 4 |
| `strings.NewReader` (flat, in reading a number) | 5.9 percent | about 4 |

CPU, at a load of about 5, of the time inside `runRow`: the run 58 percent, Validate 26 percent, the rest 16 percent. Across the whole profile, the garbage collector's background workers took 36 percent of all samples. The batch allocates 4.7 KB for every row, so about a third of the machine's time goes on collecting what the rows leave behind.

Each input of a row is read from text twice: once in Validate and once in `valuesOf`.

## One compile of formula 8 (113 allocations)

| Where | Share |
|---|---|
| `formula.Plan` | 73.7 percent |
| `formula.Check`, called by Plan | 32.6 percent |
| `formula.Parse` | 25.3 percent |
| the cost count (`costOf`, `Settings.CostOf`, `costNode.Children`) | about 19 percent |
| `childPath`, which builds a path slice for every node | 17.9 percent |

## The five largest costs

| # | Cost | Evidence |
|---|---|---|
| 1 | Decimal alignment builds a power of ten and two products on every add, subtract and compare | 42 to 85 percent of a run's allocations; 28 percent of a batch row's |
| 2 | Per-run state is built fresh on every run (program state, meter, evaluation state) | about 24 percent of formula 8's run allocations |
| 3 | An input is read from text twice in a batch row | 17.8 percent of a row's allocations |
| 4 | Garbage collection of 4.7 KB per row | 36 percent of CPU samples in the batch profile |
| 5 | Validate rebuilds the rules of each input on every call | 5.9 percent of a row's allocations, about 4 per row |

The compile has its own cost, the path slices of the cost count, at about 18 percent.

## Candidate changes, not tried

Sizes are the upper bound the shares allow. None was built, and no time gain is predicted.

| Change | Where | Upper bound on the gain |
|---|---|---|
| Skip the multiplication when the scales are equal, and read powers of ten from a shared read-only table | `number/decimal.go` | removes `pow10` and most of `aligned`: formula 8 from 40 to about 23 allocations, formula 16 from 161 to about 24, a batch row by about 28 percent |
| Reuse the per-run state, or fuse program state, meter and evaluation state into one value | `formula/run_program.go`, `run_meter.go` | about 10 of formula 8's 40 allocations |
| Read each input once per row and hand the value to both Validate and the run | `batch/run.go`, `function/validate.go` | up to half of the 11 reads, about 5 per row |
| Compute `rulesOf` once in `Prepare` | `function/validate.go` | about 4 per row |
| Build cost paths only when a problem needs one | `formula/` cost count | about 20 of a compile's 113 allocations |

The first change alone would bring both profiled formulas to the proposed allocation target of 25. Whether it brings the time under the targets, and whether the batch row targets of 225 and 675 ns can be reached with the real text reading and validation in the row, is not known and needs the comparison with the spike that was not run.

## What was not done

- The spike's evaluator was not run beside the built one (todo 3). For reference, the spike's recorded results in `spike/results/bench-run-2.txt` give its own engine 4 allocations and 368 bytes for formula 8 and 17 allocations and 728 bytes for formula 16, against 40 and 161 here. Those are the spike's recorded figures, not a new measurement.
- A batch profile of formulas 3 and 15, and of eight workers.
- A time profile of one run on a quiet machine.

## Commands to reproduce

Run from the repo root. The suite's cases are named like `CS-20_f8_run`.

```
go test ./concurrency_test/ -run '^$' -bench 'BenchmarkSpeed/CS-20_f8_run$' -benchtime 300000x -benchmem -memprofile mem.prof -memprofilerate 1 -o conc.test
go tool pprof -sample_index=alloc_objects -top -cum conc.test mem.prof

go test ./concurrency_test/ -run '^$' -bench 'BenchmarkSpeed/CS-22_f8_rows_100000_workers_1$' -benchtime 4x -benchmem -memprofile mem.prof -memprofilerate 1 -o conc.test
go test ./concurrency_test/ -run '^$' -bench 'BenchmarkSpeed/CS-22_f8_rows_100000_workers_1$' -benchtime 20x -cpuprofile cpu.prof -o conc.test
go tool pprof -top -cum conc.test cpu.prof

go test ./concurrency_test/ -run '^$' -bench 'BenchmarkSpeed/CS-23_f8_compile$' -benchtime 100000x -benchmem -memprofile mem.prof -memprofilerate 1 -o conc.test
```

`-memprofilerate 1` records every allocation and slows the run about a hundred times, so never read a time from that profile.

## What stays open

- Whether to keep the proposed targets, loosen them, or change the engine. That is the owner's decision after reading this page.
- The comparison with the spike's evaluator on the same rows (todo 3 of ISS-83).
- Re-recording the baseline on a quiet machine (see [speed-suite.md](speed-suite.md)).
