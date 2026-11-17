# Engine spike results

What the throwaway engine spike in `spike/` holds, how to re-run it, and every figure it produced. The choice made from these figures (decision 23) is on [engine-choice.md](../2.%20design/engine-choice.md). All names and values are invented.

`spike/` is throwaway. It is a nested Go module (`spike/go.mod`), outside the library's public interface, and it never changes the root `go.mod`. It is kept so the figures can be re-run.

The spike ran on the syntax of [example-formulas.md](../1.%20requirements/example-formulas.md), not on [formula-language.md](../2.%20design/formula-language.md). Every statement here that depends on the language page is marked "Provisional:".

## What spike/ holds

| Folder | What it does |
|---|---|
| `spike/numeric/` | The exact decimal used by all three engines (340 lines, non-test). Integers on a 64-bit fast path that moves to a big integer when a result leaves it. Division carries 18 places |
| `spike/builtins/` | The built-in functions that do not depend on an engine (`sum`, `avg`, date counts and so on) |
| `spike/harness/` | The 19 example formulas as cases with inputs and expected results, the `Engine` and `Program` interfaces, value equality, and `stub/`, a fake engine used to test the harness |
| `spike/engine/cel/` | The adapter for cel-go v0.33.2 |
| `spike/engine/exprlang/` | The adapter for expr-lang/expr v1.17.8 |
| `spike/engine/own/` | The own engine: lexer, parser, tree, printer, type rules and evaluator on the Go standard library |
| `spike/measure/` | Benchmarks (per run, per compile, batch sequential, batch with 8 workers), the exactness tests and the many-goroutine tests. `measure/cmd/summarise/` turns benchmark output into the per-run tables |
| `spike/limits/` | The proposed limits (`Proposed()`), the shared pre-check, one wrapper per engine, the hostile-formula cases and the limits matrix |
| `spike/fit/` | The tree, print-back and round-trip tests: each engine's tree converted to one neutral tree, printed back, respaced, compared with the page text. `fit/shape/` holds the reverse of the text translation |
| `spike/results/` | The raw output of every run, linked at the end of this page |

## How to re-run

All commands run from `spike/`.

| What | Command |
|---|---|
| Every test | `cd spike && go test ./... -count=1` |
| Exactness and many goroutines under the race detector | `go test -race ./measure/` |
| Benchmarks | `go test -bench . -benchmem -count=5 ./measure/` |
| Fit results written to a folder | `FIT_RESULTS_DIR=$PWD/results go test ./fit -v` |
| Limits table written to a file | `LIMITS_TABLE=$PWD/results/limits-table.txt go test -count=1 ./limits` |
| Per-run summary of benchmark files | `go run ./measure/cmd/summarise results/bench-run-2.txt results/bench-run.txt` |

On 2026-11-17 at 11:54, with a 1-minute load average of 3.21, `go test ./... -count=1` passed in every package. The limits table written by that run had the same 173 rows, with the same layers and outcomes in the first four columns, as `limits-table.txt`. `go run ./measure/cmd/summarise results/bench-run-2.txt results/bench-run.txt` gave output identical to `summary.txt`.

## The machine

From [machine.txt](../../spike/results/machine.txt):

| Item | Value |
|---|---|
| CPU | Apple M3 Pro, 11 logical cores (5 performance, 6 efficiency) |
| Memory | 19327352832 bytes |
| OS | macOS 15.7.3 (build 24G419), Darwin 24.6.0, arm64 |
| Go | go1.27.3 darwin/arm64 (`spike/go.mod` says go 1.25) |
| Modules | `cel.dev/cel-go` v0.33.2, `github.com/expr-lang/expr` v1.17.8 |
| Date | 2026-11-17 |
| Benchmark binary | `go test -c ./measure/`, run with `-test.bench . -test.benchmem -test.count=5`, GOMAXPROCS = 11 |

The runs were not on a quiet machine. Other agents ran tests and builds on the same machine during both runs. The 1-minute load average was sampled every 20 seconds:

| Run | Started | Samples | Min | Median | Max |
|---|---|---|---|---|---|
| Run 1 (`bench-run.txt`) | 10:56:59 | 57 | 6.81 | 13.27 | 29.22 |
| Run 2 (`bench-run-2.txt`) | 11:24:10 | 51 | 4.55 | 5.98 | 19.34 |

On 11 cores a load average above 11 means more runnable work than cores. Run 1 sat above that for most of its length (median 13.27). Run 2 stayed below it at the median (5.98) but reached 19.34. Neither run is quiet. Run 2 is the quieter one, and the figures used on the choice page come from it. Both runs are shown here so the noise is visible.

## The method

| Item | What was done |
|---|---|
| Inputs | The first example of each of the 19 formulas, from `spike/harness/cases.go`, for the per-run figures |
| Per run | One compiled program runs the same inputs `b.N` times. `go test` picks `b.N`. The figure is ns/op, B/op and allocs/op |
| Per compile | One parse, check and build of the formula, repeated `b.N` times |
| Batch rows | 10,000 rows of random but seeded inputs (seed 20260117 plus the formula number), for formulas 8, 3 and 15, made once and reused |
| Sequential batch | One goroutine runs the 10,000 rows in order |
| Parallel batch | 8 goroutines each take a block of the rows, on one shared compiled program |
| Repeats | 5 runs of every benchmark (`-count=5`) |
| Reported | The median of the 5 runs. `spread %` is the largest minus the smallest of the 5 runs, divided by the median, times 100 |
| Allocation figures | B/op and allocs/op are medians of the 5 runs. They were the same in run 1 and run 2 for every formula and engine in the per-run table |

A spread of 100 % means the largest and the smallest of the 5 runs differ by as much as the median itself. Read any figure with a spread over 20 % as indicative only.

## Per run, one compiled program (run 2, the quieter run)

Source: `summary.txt`, the section for `bench-run-2.txt`. Formula numbers are those of `example-formulas.md`.

| Formula | cel-go ns/op | cel-go spread % | expr-lang ns/op | expr-lang spread % | own ns/op | own spread % |
|---|---|---|---|---|---|---|
| 01 | 175 | 15.6 | 123 | 49.4 | 97 | 7.0 |
| 02 | 583 | 85.2 | 170 | 15.9 | 124 | 5.6 |
| 03 | 1370 | 13.8 | 652 | 22.7 | 538 | 3.3 |
| 04 | 359 | 27.3 | 260 | 38.3 | 146 | 2.3 |
| 05 | 279 | 8.9 | 205 | 26.8 | 131 | 7.6 |
| 06 | 1894 | 15.2 | 1351 | 22.9 | 559 | 4.1 |
| 07 | 124 | 24.3 | 103 | 28.9 | 88 | 24.7 |
| 08 | 1076 | 19.7 | 350 | 5.4 | 146 | 7.1 |
| 09 | 1622 | 18.2 | 657 | 16.1 | 514 | 1.8 |
| 10 | 372 | 104.2 | 118 | 4.6 | 89 | 2.8 |
| 11 | 883 | 44.4 | 627 | 1.9 | 562 | 3.3 |
| 12 | 1048 | 13.4 | 841 | 20.1 | 648 | 3.1 |
| 13 | 492 | 11.7 | 399 | 2.7 | 377 | 7.1 |
| 14 | 482 | 4.5 | 403 | 2.4 | 236 | 5.4 |
| 15 | 681 | 9.9 | 567 | 0.8 | 497 | 2.7 |
| 16 | 1295 | 15.8 | 557 | 3.1 | 461 | 13.4 |
| 17 | 615 | 1.5 | 147 | 2.4 | 106 | 4.4 |
| 18 | 303 | 4.5 | 207 | 2.1 | 206 | 1.8 |
| 19 | 67 | 2.5 | 37 | 2.9 | 68 | 26.2 |

| Formula | cel-go B/op | cel-go allocs/op | expr-lang B/op | expr-lang allocs/op | own B/op | own allocs/op |
|---|---|---|---|---|---|---|
| 01 | 144 | 6 | 88 | 2 | 184 | 2 |
| 02 | 280 | 10 | 128 | 3 | 160 | 1 |
| 03 | 1264 | 40 | 984 | 28 | 920 | 24 |
| 04 | 376 | 13 | 312 | 8 | 232 | 3 |
| 05 | 216 | 7 | 168 | 5 | 176 | 2 |
| 06 | 2138 | 46 | 1850 | 24 | 856 | 14 |
| 07 | 72 | 3 | 64 | 1 | 184 | 2 |
| 08 | 592 | 13 | 528 | 6 | 368 | 4 |
| 09 | 952 | 31 | 1016 | 28 | 920 | 24 |
| 10 | 280 | 10 | 64 | 3 | 176 | 2 |
| 11 | 904 | 29 | 936 | 26 | 920 | 24 |
| 12 | 1248 | 40 | 1192 | 32 | 968 | 25 |
| 13 | 184 | 15 | 192 | 13 | 264 | 10 |
| 14 | 432 | 17 | 432 | 12 | 240 | 5 |
| 15 | 1000 | 33 | 928 | 27 | 920 | 24 |
| 16 | 1216 | 37 | 728 | 19 | 728 | 17 |
| 17 | 480 | 16 | 136 | 4 | 184 | 2 |
| 18 | 80 | 4 | 32 | 1 | 160 | 1 |
| 19 | 8 | 1 | 0 | 0 | 160 | 1 |

Across the 19 formulas, medians computed from the table above: cel-go 583 ns/op, expr-lang 350 ns/op, own engine 206 ns/op. The own engine has the lowest median on 18 of the 19 formulas in this run and expr-lang on formula 19 (37 ns against 68 ns). Median of the allocs/op across the 19: cel-go 15, expr-lang 8, own 4. The largest allocs/op: cel-go 46 (formula 6), expr-lang 32 (formula 12), own 25 (formula 12). The largest spread: cel-go 104.2 % (formula 10), expr-lang 49.4 % (formula 1), own 26.2 % (formula 19). The median spread: cel-go 15.2 %, expr-lang 5.4 %, own 4.4 %.

## Per run (run 1, the noisier run)

Source: `summary.txt`, the section for `bench-run.txt`. Same method, load median 13.27.

| Formula | cel-go ns/op | cel-go spread % | expr-lang ns/op | expr-lang spread % | own ns/op | own spread % |
|---|---|---|---|---|---|---|
| 01 | 315 | 46.4 | 164 | 33.4 | 135 | 76.6 |
| 02 | 1068 | 60.5 | 251 | 115.5 | 145 | 22.5 |
| 03 | 2119 | 72.0 | 1005 | 40.2 | 798 | 96.0 |
| 04 | 414 | 22.1 | 255 | 7.8 | 205 | 74.0 |
| 05 | 529 | 184.1 | 188 | 10.7 | 406 | 94.0 |
| 06 | 3152 | 60.3 | 1310 | 162.7 | 1583 | 35.6 |
| 07 | 139 | 61.0 | 166 | 45.1 | 112 | 61.8 |
| 08 | 775 | 16.9 | 362 | 4.6 | 185 | 24.6 |
| 09 | 1061 | 37.5 | 1109 | 68.2 | 554 | 7.1 |
| 10 | 399 | 13.0 | 125 | 9.3 | 92 | 5.0 |
| 11 | 789 | 10.5 | 674 | 18.6 | 595 | 359.6 |
| 12 | 1138 | 17.0 | 890 | 4.4 | 2843 | 87.2 |
| 13 | 524 | 6.3 | 436 | 9.9 | 912 | 53.2 |
| 14 | 600 | 4.0 | 398 | 15.3 | 616 | 44.5 |
| 15 | 946 | 50.7 | 569 | 15.9 | 589 | 32.0 |
| 16 | 1651 | 12.7 | 586 | 12.1 | 527 | 11.6 |
| 17 | 711 | 13.2 | 151 | 9.5 | 125 | 8.2 |
| 18 | 356 | 4.1 | 211 | 4.6 | 236 | 9.5 |
| 19 | 81 | 5.9 | 40 | 20.1 | 80 | 3.0 |

Medians over the 19 formulas, computed from the table above: cel-go 711 ns/op, expr-lang 362 ns/op, own engine 406 ns/op. The own engine has the lowest median on 11 formulas in this run and expr-lang on 8. The largest spread: cel-go 184.1 %, expr-lang 162.7 %, own engine 359.6 % (formula 11).

Run 1 against run 2, per formula: the own engine's median is higher in run 1 for every formula, by at most 4.39 times (formula 12: 2843 against 648) and by more than 2 times on 5 formulas (5, 6, 12, 13, 14). cel-go's largest ratio is 1.90 (formula 5) and expr-lang's 1.69 (formula 9). So the own engine's per-run figures moved more with the load than the other two did. The own engine's batch figures did not (see below).

## Per compile (run 2)

Source: `summary.txt`, the section for `bench-run-2.txt`. A compile is a parse, a check and building one program. Medians over the 19 formulas, computed from the table: cel-go 92215 ns, expr-lang 26260 ns, own engine 2483 ns.

| Formula | cel-go ns/op | cel-go spread % | expr-lang ns/op | expr-lang spread % | own ns/op | own spread % |
|---|---|---|---|---|---|---|
| 01 | 87247 | 12.8 | 24203 | 2.9 | 1630 | 39.5 |
| 02 | 85317 | 2.5 | 31351 | 12.7 | 1990 | 14.8 |
| 03 | 86697 | 4.6 | 20146 | 5.5 | 2558 | 5.1 |
| 04 | 92215 | 2.7 | 54444 | 5.7 | 2649 | 9.1 |
| 05 | 77629 | 4.2 | 32086 | 0.4 | 1765 | 22.9 |
| 06 | 185225 | 11.5 | 50343 | 6.7 | 7328 | 5.7 |
| 07 | 118124 | 35.4 | 31168 | 22.4 | 2483 | 9.8 |
| 08 | 82034 | 16.6 | 15735 | 35.3 | 1680 | 10.4 |
| 09 | 100423 | 21.7 | 23518 | 8.8 | 2244 | 4.5 |
| 10 | 118190 | 8.5 | 26260 | 2.6 | 3545 | 3.4 |
| 11 | 94363 | 37.7 | 20719 | 4.0 | 3079 | 28.5 |
| 12 | 109714 | 6.3 | 101481 | 3.5 | 3900 | 3.8 |
| 13 | 81615 | 42.6 | 15056 | 5.1 | 2072 | 10.5 |
| 14 | 109716 | 3.5 | 62570 | 6.5 | 3561 | 3.8 |
| 15 | 80345 | 20.6 | 19496 | 6.8 | 1898 | 6.6 |
| 16 | 110958 | 20.6 | 36061 | 32.4 | 2897 | 13.3 |
| 17 | 70876 | 4.7 | 18339 | 3.3 | 1316 | 13.7 |
| 18 | 109679 | 9.3 | 27823 | 5.4 | 3407 | 3.8 |
| 19 | 63070 | 2.3 | 12417 | 3.6 | 672 | 3.8 |

| Formula | cel-go B/op | cel-go allocs/op | expr-lang B/op | expr-lang allocs/op | own B/op | own allocs/op |
|---|---|---|---|---|---|---|
| 01 | 115075 | 1555 | 30788 | 400 | 2936 | 45 |
| 02 | 114048 | 1553 | 32188 | 409 | 3304 | 54 |
| 03 | 118899 | 1587 | 28556 | 294 | 4744 | 69 |
| 04 | 122578 | 1642 | 69834 | 1099 | 4936 | 73 |
| 05 | 108426 | 1389 | 47135 | 629 | 2912 | 53 |
| 06 | 210209 | 2892 | 40822 | 487 | 11824 | 218 |
| 07 | 140755 | 1935 | 34389 | 469 | 3488 | 71 |
| 08 | 112617 | 1412 | 24651 | 226 | 2992 | 53 |
| 09 | 121136 | 1670 | 30492 | 337 | 3552 | 66 |
| 10 | 144439 | 2028 | 27684 | 295 | 5864 | 107 |
| 11 | 123055 | 1672 | 28396 | 295 | 5120 | 74 |
| 12 | 140138 | 1894 | 113017 | 1946 | 5848 | 104 |
| 13 | 109870 | 1431 | 24284 | 223 | 3104 | 55 |
| 14 | 140354 | 1882 | 80388 | 1212 | 5520 | 95 |
| 15 | 111147 | 1474 | 27900 | 292 | 3192 | 57 |
| 16 | 133483 | 1865 | 36269 | 481 | 5232 | 80 |
| 17 | 104233 | 1368 | 27836 | 305 | 2512 | 40 |
| 18 | 134859 | 1836 | 25227 | 242 | 5752 | 90 |
| 19 | 95590 | 1214 | 18824 | 148 | 1176 | 20 |

## Batch of 10,000 rows, one goroutine

Run 2 (load median 5.98):

| Engine | Formula | ns/batch median | ns/row | spread % | B/batch | allocs/batch |
|---|---|---|---|---|---|---|
| cel-go | 03 | 10690591 | 1069.1 | 83.2 | 12185797 | 383852 |
| cel-go | 08 | 3973828 | 397.4 | 4.4 | 5896103 | 128918 |
| cel-go | 15 | 6336968 | 633.7 | 16.6 | 9976705 | 327757 |
| expr-lang | 03 | 6365992 | 636.6 | 133.3 | 9589243 | 271510 |
| expr-lang | 08 | 3256809 | 325.7 | 16.0 | 5280824 | 60003 |
| expr-lang | 15 | 5263020 | 526.3 | 12.2 | 9272113 | 269693 |
| own | 03 | 4748257 | 474.8 | 5.4 | 8948246 | 231509 |
| own | 08 | 1415332 | 141.5 | 16.7 | 3680003 | 40000 |
| own | 15 | 4338364 | 433.8 | 5.3 | 9191272 | 239708 |

Run 1 (load median 13.27):

| Engine | Formula | ns/batch median | ns/row | spread % | B/batch | allocs/batch |
|---|---|---|---|---|---|---|
| cel-go | 03 | 10978937 | 1097.9 | 16.3 | 12185797 | 383852 |
| cel-go | 08 | 4474784 | 447.5 | 4.2 | 5896091 | 128918 |
| cel-go | 15 | 6248596 | 624.9 | 18.4 | 9976700 | 327757 |
| expr-lang | 03 | 6549928 | 655.0 | 17.2 | 9589252 | 271510 |
| expr-lang | 08 | 3497488 | 349.7 | 50.3 | 5280810 | 60003 |
| expr-lang | 15 | 5311764 | 531.2 | 31.4 | 9272109 | 269693 |
| own | 03 | 4641369 | 464.1 | 10.0 | 8948246 | 231509 |
| own | 08 | 1475219 | 147.5 | 13.4 | 3680003 | 40000 |
| own | 15 | 4339420 | 433.9 | 4.5 | 9191271 | 239708 |

## Batch of 10,000 rows, 8 workers splitting the rows

Run 2:

| Engine | Formula | ns/batch median | ns/row | spread % | B/batch | allocs/batch |
|---|---|---|---|---|---|---|
| cel-go | 03 | 3129859 | 313.0 | 14.8 | 12186181 | 383873 |
| cel-go | 08 | 1541409 | 154.1 | 4.1 | 5895701 | 128934 |
| cel-go | 15 | 2066241 | 206.6 | 80.5 | 9977380 | 327778 |
| expr-lang | 03 | 1995853 | 199.6 | 8.3 | 9590023 | 271531 |
| expr-lang | 08 | 1234022 | 123.4 | 9.4 | 5281889 | 60026 |
| expr-lang | 15 | 1909198 | 190.9 | 5.0 | 9272955 | 269715 |
| own | 03 | 1800609 | 180.1 | 37.5 | 8948932 | 231527 |
| own | 08 | 606341 | 60.6 | 20.8 | 3680673 | 40018 |
| own | 15 | 1660590 | 166.1 | 10.2 | 9191956 | 239726 |

Run 1:

| Engine | Formula | ns/batch median | ns/row | spread % | B/batch | allocs/batch |
|---|---|---|---|---|---|---|
| cel-go | 03 | 3003279 | 300.3 | 3.7 | 12186177 | 383873 |
| cel-go | 08 | 1452267 | 145.2 | 23.1 | 5895687 | 128934 |
| cel-go | 15 | 2062440 | 206.2 | 9.7 | 9977373 | 327778 |
| expr-lang | 03 | 3148388 | 314.8 | 118.7 | 9590052 | 271531 |
| expr-lang | 08 | 1298374 | 129.8 | 85.9 | 5281892 | 60026 |
| expr-lang | 15 | 2848899 | 284.9 | 156.9 | 9272957 | 269715 |
| own | 03 | 2042620 | 204.3 | 25.7 | 8948944 | 231527 |
| own | 08 | 677131 | 67.7 | 24.6 | 3680674 | 40018 |
| own | 15 | 1934650 | 193.5 | 13.1 | 9191964 | 239726 |

The 8 workers gave the own engine 2.6 times, 2.3 times and 2.6 times the sequential speed on formulas 3, 8 and 15 in run 2 (474.8 / 180.1, 141.5 / 60.6, 433.8 / 166.1 ns/row). On 11 cores with a load average of 6 to 19 the speed-up was never 8 times, and the spreads of the parallel rows are larger than those of the sequential rows. The sequential figures of the own engine differ between run 1 and run 2 by at most 4.2 % (formula 8: 147.5 and 141.5 ns/row).

## Exactness

Source: `exactness-run.txt` (`go test -v ./measure/` and `go test -race -v ./measure/`, both pass).

| Test | What it does | cel-go | expr-lang | own |
|---|---|---|---|---|
| 0.1 + 0.2 | Adds the decimals 0.1 and 0.2 and compares with 0.3. The baseline in the same file shows that `float64` gives 0.30000000000000004 | pass | pass | pass |
| Rounding | 13 cases such as `round(2.675, 2)` is 2.68, `round(-2.675, 2)` is -2.68, `round(2.5, 0)` is 3, `round(-2.5, 0)` is -3, and the places of the result are exact | pass | pass | pass |
| Fast-path overflow | 8 cases that leave and return to 64-bit integers (product leaves int64, product and back down, int64 maximum plus one, negating the int64 minimum, a square, a sum of twenty large values, a sum chain back to a small result, a mean), each compared with `math/big` | pass | pass | pass |
| Division | `10.00 / 3` gives `3.333333333333333333`, 18 places, and `round(10.00 / 3, 2)` gives `3.33` | pass | pass | pass |
| Decimal and float never mix | `a + f` and `a > f` with `a` a decimal and `f` a float are refused when the formula is compiled | pass | pass | pass |
| Explicit conversion | `to_float` and `to_decimal` work | pass | pass | pass |
| One program, many goroutines | 64 goroutines run 2000 inputs each on one compiled program, for formulas 3, 5, 6, 7, 8 and 15, and every result equals the single-threaded result. Also done for the shared library programs. Run once without and once with `-race` | pass | pass | pass |
| Inputs not changed | A run does not change its input map | pass | pass | pass |

All three engines use the same `spike/numeric` decimal, so the first five rows test how each engine is wired to it, not three arithmetics.

The refusal sentences of the decimal and float mix, as the engines gave them:

| Engine | Sentence |
|---|---|
| cel-go | `ERROR: <input>:1:3: found no matching overload for '_+_' applied to '(decimal, double)'` followed by the formula and a marker line |
| expr-lang | `invalid operation: + (mismatched types numeric.Decimal and float64) (1:3)` followed by the formula and a marker line |
| own | `The formula cannot be used: a decimal and a float cannot be combined, convert one with to_float or to_decimal at line 1, column 3.` |

## Limits

Source: `limits-run.txt` and `limits-table.txt`. The limits used in the runs are `Proposed()` in `spike/limits/limits.go`: depth 64, nodes 1000, cost 1,000,000, formula length 5000, text length 1000, list length 10,000, number magnitude 18, decimal places 18, and a time limit of 250 ms. The time-limit cases (L06) used 30 ms and L09 used a host deadline of 15 ms, so they finish quickly.

Under decision 19 the library has no time limit of its own. The spike measured one only because the first draft of the limits had one. In the results below the time limit is a host-side safety, not a library limit.

Every case ran in two modes. `native` uses only what the engine itself offers. `full` adds the shared pre-check (a scan of the formula text, 248 lines) and the shared wrappers. The layer says who refused: `native` is the engine, `precheck` is the shared scan, `wrapper` is the shared code around a run, `function` is a limit counted inside a built-in function the spike wrote, `none` is no refusal.

In `full` mode every refusal case gave the same sentence for the three engines, for example `This formula goes over the depth limit of 64.` In `native` mode they differed. The table gives the `native` result of the cases that stand for each limit.

| Limit | Case | cel-go | expr-lang | own |
|---|---|---|---|---|
| depth | L01, one level past 64 | native, 621 us | none: no refusal, result 1, 218 us | native, 115 us |
| depth | L01b, how deep is accepted | up to 63 brackets | no refusal up to the 128 tried | up to 63 brackets |
| nodes | L03, one node past 1000 | native, 12.50 ms | native, 246 us | native, 253 us |
| cost | L05, 128 copies of `sum(range(1, 10000))` added together | native, 5.04 ms | function, 4.80 ms | native, 2.53 ms |
| time | L06, 30 ms | wrapper, 32.30 ms | wrapper, 31.26 ms | native, 30.70 ms |
| formula length | P05, a million `(` | native, 13.47 ms | CRASH: fatal error: stack overflow, the whole process died, 854.50 ms | native, 390 us |
| formula length | I04, a million-digit number | native, 3.04 ms | no limit: read all 1,000,000 characters and refused with a number-parsing sentence, 16.21 ms | native, 344 us |
| text length | I01, 1100 characters in the formula | none, 834 us | none, 125 us | native, 43 us |
| text length | I02, 1,000,000 characters as an input | wrapper, 1.03 ms | wrapper, 360 us | native, 345 us |
| list length | C01, `count(range(1, 1000000000))` | function, 787 us | function, 162 us | native, 46 us |
| list length | I12, a million-entry list as an input | wrapper, 619 us | wrapper, 89 us | native, 55 us |
| number magnitude | I04b, a hundred digits | native, 629 us, with its own sentence (`invalid int literal`) | native, 85 us, with a parsing sentence (`strconv.ParseInt`) | native, 37 us, with the limit sentence |
| number magnitude | I05, a million digits as an input | wrapper, 12.26 ms | wrapper, 11.58 ms | native, 10.88 ms |
| decimal places | I09b, places past 18 in a literal | wrapper, 689 us | wrapper, 110 us | native, 49 us |
| decimal places | I10, `round(1 / 3, 1000000)` | wrapper, 21.68 ms | wrapper, 20.68 ms | native, 46 us |

Smallest limits that let all 19 formulas pass natively (`TestSmallestNativeLimitsThatPassTheNineteenFormulas`): cel-go depth 7 and nodes 38; expr-lang no native depth limit and nodes 48; own depth 8 and nodes 34. The largest among the 19 by the own engine's count: depth 8, nodes 34, formula length 175 characters. Each engine counts depth and nodes in its own way, so the same number does not mean the same formulas on every engine.

Other cases, all engines in `full` mode:

| Case | What was checked | cel-go | expr-lang | own |
|---|---|---|---|---|
| L02, L04 | A formula exactly at the depth limit and exactly at the node limit is accepted | results -1 and -500 | results -1 and -500 | results -1 and -500 |
| L07 | `2 + 3` returns 5 after six hostile formulas and a time-limit run | pass | pass | pass |
| L08 | Goroutines before, right after the refusals, after waiting | 2, 3, 2 after 22.25 ms | 2, 3, 2 after 2.27 ms | 2, 2, 2 after 0 us |
| L09 | A host deadline of 15 ms arrives while a run is under way (time limit 400 ms) | stopped after 15.67 ms | stopped after 15.53 ms | stopped after 15.09 ms |
| L10 | Ten hostile formulas at once; a good formula answers meanwhile | 5 in 331 us | 5 in 788 us | 5 in 38 us |

In L09 the spike reported the host's deadline with the time-limit sentence (`This formula goes over the time limit of 15ms.`). That is the spike's choice, not a decision. Whether and how the library honours a host's cancel is still open under decision 19.

Cost is counted at run time in all three engines of the spike. The own engine charges 1 for each node evaluated and the item count for each list argument (`spike/engine/own/kinds.go`, `guards.go`). cel-go uses its own cost tracking with a cost model that adds the list size for `sum`, `count`, `avg`, `min`, `max` and `range` (`spike/limits/cel.go`). expr-lang is charged only by the `range` function that the limits tests define (`spike/limits/helpers_test.go`, using `Charge` in `spike/limits/meter.go`), and its memory budget is set to the same number (`spike/limits/expr.go`). A formula of plain operators is not counted by expr-lang in the spike. The three counts are in different units, so `1,000,000` does not mean the same work on each. The spike did not measure the cost of the 19 formulas in any of the three units.

## Fit: tree, print-back and round trip

Source: `fit-run.txt` and `fit-table.txt`. Each engine's tree is converted to one neutral tree. The formula is printed back from the tree, respaced in 6 ways, read again and compared.

| Result | cel-go | expr-lang | own |
|---|---|---|---|
| Tree converts, of 19 | 19 | 19 | 19 |
| Tree equals the own engine's tree, of 19 | 19 | 19 | 19 |
| Print-back idempotent, of 19 | 19 | 19 | 19 |
| Canonical over 6 respacings, of 114 variants | 114 | 114 | 114 |
| Printed text read again gives the same tree, of 19 | 19 | 19 | 19 |
| Printed text equals the formula as written on `example-formulas.md`, of 19 | 14 | 16 | 18 |
| Nodes in the 19 formulas as the engine gives them, and in the neutral tree | 234 and 220 | 263 and 220 | 220 and 220 |
| Nodes folded away in the conversion | 14 | 43 | 0 |
| Raw node kinds met | 33 | 9 | 9 |
| Translation steps from the page text to the engine's text | 3 | 3 | 0 |
| Reverse step exact on the 19 | 19 | 19 | not needed |
| Reverse step changed a probe | `2.0f` became `2f` | `2f` became `2`, `2.0f` became `2f` | none |

The three translation steps of cel-go and expr-lang (`spike/engine/cel/translate.go`, `spike/engine/exprlang/translate.go`): a decimal literal such as `12.50` is wrapped (`decimal("12.50")` and `dec("12.50")`) so its written digits are kept, a float literal with `f` is rewritten, and `@` is replaced by `_v` because neither lexer accepts `@` (expr-lang gives `unrecognized character: U+0040 '@'`).

The formula that the own engine does not print as written is formula 18. `example-formulas.md` writes `(paid_up && months_between(joined, on) >= 6)` with brackets and the own engine prints them without. Provisional: `formula-language.md` (line 871) prints formula 18 without them, as the own engine does, and states the rule of the fewest brackets. cel-go prints formula 18 without them too. The other misses are formulas 1, 6, 7 and 10 on cel-go (it adds brackets around every comparison, arithmetic part and conditional branch) and formulas 6, 7 and 10 on expr-lang (it adds brackets around a conditional in the else branch). expr-lang keeps the brackets of formula 18 because its printer keeps a bracket around `&&` inside `||` (probe `a || (b && c)` in the bracket table), so its 16 of 19 is not a sign of the fewest-brackets rule.

Bracket probes (25 formulas of brackets, negation and the conditional, `fit-table.txt`, section "brackets"). A row counts as tree-changing when the engine's tree differs from the own engine's tree, when the printed text read again gives another tree, or when the printed text read by the own engine gives another tree.

| Engine | Printed text equals the standard text | Tree-changing rows |
|---|---|---|
| cel-go | 19 of 25 | 4 |
| expr-lang | 19 of 25 | 2 |
| own | 25 of 25 | 0 |

The tree-changing rows:

| Engine | Source | Printed | What changed |
|---|---|---|---|
| cel-go | `!!a` | `a` | The double negation is folded away by the parser |
| cel-go | `(a == b) < c` | `a == b < c` | The printed text read by the own engine is another tree |
| cel-go | `a && (b && c)` | `a && b && c` | The printed text read again is another tree |
| cel-go | `a == b < c` | `a == b < c` | The tree differs from the own engine's tree |
| expr-lang | `(a == b) < c` | `a == b < c` | The printed text read by the own engine is another tree |
| expr-lang | `a == b < c` | `a == b < c` | The tree differs from the own engine's tree |

Provisional: the own engine reads `a == b < c` as `a == (b < c)`, which is the order of the operator table on `formula-language.md` (`<` binds tighter than `==`). cel-go and expr-lang read it the other way.

Literals: `"a \"q\" b"`, `"two  spaces"`, `[0.00, 1000.00]`, `round(x * 0.010, 2)`, `true` and `x * 0.90` were kept as written by all three. A float literal `2f` is printed as `2f` by cel-go and the own engine and as `2` by expr-lang.

## Code size and findings from the adapters

Non-test lines, counted with `cat $(ls *.go | grep -v _test) | wc -l` in each folder.

| Folder or files | Lines |
|---|---|
| `spike/engine/cel/` | 734 |
| `spike/engine/exprlang/` | 553 |
| `spike/engine/own/` | 2186 |
| `spike/engine/own/` before the limits commit a3fe97a | 1749 |
| Added to `own/` by the limits commit (non-test insertions; 35 lines deleted) | 472 |
| Tree and print files: cel `tree.go` and `print.go` | 29 and 24 |
| Tree and print files: exprlang `tree.go` | 33 |
| Tree and print files: own `ast.go`, `tree.go`, `print.go` | 84, 17 and 94 |
| Converters to the neutral tree: `fit/convert_cel.go`, `convert_exprlang.go`, `convert_own.go` (plus `fit/node.go`, 83, shared) | 83, 95 and 50 |
| Function wiring: cel `functions.go`; exprlang `functions.go`; own `functions*.go` (4 files) | 238, 176 and 227 |
| Limits: `limits/cel.go`, `expr.go`, `own.go` | 128, 64 and 79 |
| Limits shared by the three: `limits.go`, `guard.go`, `precheck.go`, `meter.go` | 127, 218, 248 and 49 (642) |
| `spike/numeric/` and `spike/builtins/`, used by all three | 340 and 132 |

The own engine is 1749 lines before the limits and 2186 with them.

Findings from building the adapters, each checked against the code or run:

| Engine | Finding | Where |
|---|---|---|
| cel-go | Operators are singleton functions, so the decimal operators are declared for the checker only and swapped at plan time in a decorator keyed by overload ID | `spike/engine/cel/operators.go` (`bindDecimalOperators`) |
| cel-go | `==` between a decimal and an integer is refused at compile: `found no matching overload for '_==_' applied to '(decimal, int)'`. `==` between two decimals works through the value's `Equal` | run on 2026-11-17 against `spike/engine/cel` |
| cel-go | Unary minus on a decimal is refused: `found no matching overload for '-_' applied to '(decimal)'` | same run |
| cel-go | Every decimal operation boxes its result into the `ref.Val` interface (`decimalValue{...}`), which shows in the allocs/op column | `operators.go`, per-run table |
| expr-lang | A patcher `forgetCallTypes` clears the type of every call node so the operator functions can be used. It depends on expr's checker internals | `spike/engine/exprlang/operators.go` |
| expr-lang | Operator and function arguments arrive as boxed `...any` | `operators.go` (`call(args ...any)`) |
| expr-lang | The lexer rejects `@`, so the text is always translated. The float suffix `f` is removed, so `2f` is lost | `translate.go`; `fit-run.txt` |
| expr-lang | Unary minus on a decimal is refused: `invalid operation: - (mismatched type numeric.Decimal)` | run on 2026-11-17 |
| expr-lang | No depth limit, no formula length limit, and a stack overflow that kills the process on 1,000,000 unclosed brackets | `limits-table.txt` (P05, P04, L01) |
| own | Typed closures, exact line and column on read and type errors, tree and print from its own tree, and every built-in is wired by hand | `spike/engine/own/` |
| own | `&&` and `||` skip the right side when the left decides. Provisional: the language page evaluates both sides | `compile_binary.go` (`both`, `either`) |

The own engine reads and prints unary minus (`-d` on a decimal gave `-2.00`) and `d == 2` with an integer (true).

## Caveats

- Noise. Neither benchmark run was quiet (load medians 13.27 and 5.98 on 11 cores). The spreads in the tables are the measure of it. The per-run figures of the own engine moved by up to 4.39 times between the runs; the batch figures by at most 4.2 % (sequential) and more in parallel.
- One machine, one operating system, one Go version. Nothing was run on a second machine, so the claim of decision 19 that a cost count gives the same outcome on every machine rests on the count being a count and not a clock. It was not checked across machines.
- Throwaway code. `spike/` was written fast, is outside the library's public interface, and is not reviewed as library code. The own engine of the spike is not the library's engine.
- The spike did not fuzz any parser. The only crash found came from a plain test with 1,000,000 brackets.
- Provisional syntax. The 19 formulas are those of `example-formulas.md`. The spike supports a float literal (`0.75f`), unary minus (in the own engine) and `range`, which the language page leaves out, and its hostile cases used `range`. The differences are listed on the choice page.
- The time limit is not a library limit (decision 19).
- Not measured: the cost of the 19 formulas in any engine's unit; the memory held by a compiled program; compile time under the limits of `full` mode; a second batch size.

## Raw files

| File | What it holds |
|---|---|
| [machine.txt](../../spike/results/machine.txt) | The machine, the versions and the load samples |
| [summary.txt](../../spike/results/summary.txt) | The per-run, per-compile and batch tables of both runs |
| [bench-run.txt](../../spike/results/bench-run.txt) | Run 1, raw benchmark output |
| [bench-run-2.txt](../../spike/results/bench-run-2.txt) | Run 2, raw benchmark output |
| [exactness-run.txt](../../spike/results/exactness-run.txt) | Exactness, many-goroutine and race runs |
| [limits-run.txt](../../spike/results/limits-run.txt) | The limits run, with the raw engine sentences |
| [limits-table.txt](../../spike/results/limits-table.txt) | One row per case, engine and mode |
| [fit-run.txt](../../spike/results/fit-run.txt) | The fit run |
| [fit-table.txt](../../spike/results/fit-table.txt) | The fit tables and the bracket probes |
