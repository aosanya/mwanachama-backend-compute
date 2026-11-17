# The library runs on its own engine

The library reads, prints, checks and runs a formula on an engine written in this repo on the Go standard library. cel-go and expr-lang are not used.

End goal: one tree serves the reader, the printer, the block view and the evaluator, and a runaway formula is stopped by a count of work that is the same on every machine.

Decided by the owner on 2026-11-17 (decision 23 on [decisions.md](../1.%20requirements/decisions.md)). The decision is the engine. The limit values and the speed targets below are proposed, and every statement that depends on [formula-language.md](formula-language.md) is marked "Provisional:". All names and figures are invented.

## The decision

Decision 23: the library runs formulas on its own engine, written in this repo on the Go standard library: its own tree, reader, standard-shape printer and evaluator. Nothing is translated to or from another engine's form, and the library depends on no expression engine.

What it costs: the repo owns and maintains a parser and an evaluator. In the spike that is 2186 lines (1749 before the limit code, 472 lines of limit code added), every built-in is wired by hand, the parser is ours to fuzz, and the language is ours to extend. The costs are set out in "What the choice costs".

## The measured figures

A spike ran three candidates on the 19 formulas of [example-formulas.md](../1.%20requirements/example-formulas.md): cel-go v0.33.2, expr-lang v1.17.8 and an own engine. The spike is throwaway and kept in `spike/`. The method, the machine and every figure are on [engine-spike-results.md](../3.%20implementation/engine-spike-results.md). The files of the figures are in `spike/results/`.

The runs were not on a quiet machine. Other agents ran at the same time. The 1-minute load average on 11 cores had a median of 5.98 in run 2 and 13.27 in run 1. Rows with figures from both runs show run 2 first and run 1 in brackets. The largest spread (largest minus smallest of 5 runs, over the median) of the run 2 per-run figures is 104.2 % for cel-go, 49.4 % for expr-lang and 26.2 % for the own engine. The spread of every row is on the results page.

| Measure | cel-go | expr-lang | own engine | File |
|---|---|---|---|---|
| Exact on 0.1 + 0.2 / rounding (13 cases) / overflow (8 cases against `math/big`) | yes / yes / yes | yes / yes / yes | yes / yes / yes | `exactness-run.txt` |
| Decimal with float refused at compile (`a + f`, `a > f`) | yes | yes | yes | `exactness-run.txt` |
| One compiled program from 64 goroutines x 2000 runs, formulas 3, 5, 6, 7, 8, 15, under `-race` | pass | pass | pass | `exactness-run.txt` |
| ns/op median, formulas 3 / 6 / 8 / 15 / 19, run 2 | 1370 / 1894 / 1076 / 681 / 67 | 652 / 1351 / 350 / 567 / 37 | 538 / 559 / 146 / 497 / 68 | `summary.txt`, `bench-run-2.txt` |
| The same, run 1 | 2119 / 3152 / 775 / 946 / 81 | 1005 / 1310 / 362 / 569 / 40 | 798 / 1583 / 185 / 589 / 80 | `summary.txt`, `bench-run.txt` |
| ns/op, median over the 19 formulas, run 2 (run 1) | 583 (711) | 350 (362) | 206 (406) | computed from `summary.txt` |
| Formulas where the engine has the lowest median, of 19, run 2 (run 1) | 0 (0) | 1 (8) | 18 (11) | computed from `summary.txt` |
| allocs/op, formulas 3 / 6 / 8 / 15 / 19 (same in both runs) | 40 / 46 / 13 / 33 / 1 | 28 / 24 / 6 / 27 / 0 | 24 / 14 / 4 / 24 / 1 | `summary.txt` |
| Batch of 10,000 rows, one goroutine, ns/row, formulas 8 / 3 / 15, run 2 (run 1) | 397.4 / 1069.1 / 633.7 (447.5 / 1097.9 / 624.9) | 325.7 / 636.6 / 526.3 (349.7 / 655.0 / 531.2) | 141.5 / 474.8 / 433.8 (147.5 / 464.1 / 433.9) | `summary.txt` |
| Batch of 10,000 rows, 8 workers, ns/row, formulas 8 / 3 / 15, run 2 (run 1) | 154.1 / 313.0 / 206.6 (145.2 / 300.3 / 206.2) | 123.4 / 199.6 / 190.9 (129.8 / 314.8 / 284.9) | 60.6 / 180.1 / 166.1 (67.7 / 204.3 / 193.5) | `summary.txt` |
| Compile ns, formulas 3 / 6 / 8 / 15 / 19, run 2 | 86697 / 185225 / 82034 / 80345 / 63070 | 20146 / 50343 / 15735 / 19496 / 12417 | 2558 / 7328 / 1680 / 1898 / 672 | `summary.txt`, `bench-run-2.txt` |
| Depth limit enforced by the engine itself (limit 64) | yes | no: no refusal up to the 128 brackets tried | yes | `limits-table.txt`, L01, L01b |
| Node limit enforced by the engine itself (limit 1000) | yes | yes | yes | `limits-table.txt`, L03 |
| Cost limit enforced by the engine itself (limit 1,000,000) | yes | no: charged by the `range` function of the limits tests, plus a memory budget | yes | `limits-table.txt`, L05 |
| Time limit enforced by the engine itself | no: wrapper | no: wrapper | yes | `limits-table.txt`, L06 |
| Formula length limit enforced by the engine itself (limit 5000) | yes | no | yes | `limits-table.txt`, P05, I04 |
| Crash or hang found | none | crash: fatal stack overflow, the whole process died, on 1,000,000 unclosed `(` without the shared pre-check | none | `limits-table.txt`, P05 |
| Printed text equals the formula as written, of 19 | 14 | 16 | 18 | `fit-table.txt` |
| Tree-changing probes, of 25 bracket probes | 4 | 2 | 0 | `fit-table.txt`, section "brackets" |
| Nodes folded away turning the engine's tree into the neutral tree, of 220 | 14 (234 raw) | 43 (263 raw) | 0 (220 raw) | `fit-table.txt` |
| Engine code, non-test lines in the engine folder | 734 | 553 | 2186 (1749 before limits) | counted in `spike/engine/` |
| Translation steps from the page text to the engine's text | 3 | 3 | 0 | `spike/engine/*/translate.go`, `fit-table.txt` |

Notes on the table:

- All three engines use the same decimal package, `spike/numeric` (340 lines), so the exactness rows test how each engine is wired to it. They do not compare three arithmetics. The division gave 18 places on all three (`10.00 / 3` is `3.333333333333333333`).
- The own engine has the lowest median on 18 of 19 formulas in run 2 and on 11 in run 1. On formula 19 expr-lang is faster in both runs (37 and 40 ns against 68 and 80 ns).
- The own engine's per-run figures moved more with the load than the others: run 1 against run 2 is up to 4.39 times (formula 12) for the own engine, 1.90 for cel-go and 1.69 for expr-lang. Its batch figures moved by at most 4.2 % sequentially.
- With 8 workers the three engines are closer than sequentially (formula 3: 313.0, 199.6 and 180.1 ns/row), because allocation grows with the work: cel-go allocates 383852 times per batch of formula 3, expr-lang 271510 and the own engine 231509 (`summary.txt`).
- "Enforced by the engine itself" is the `native` mode of `limits-table.txt`, where only the engine's own options and the spike's adapter code are used. In `full` mode, with a shared pre-check of 248 lines and shared wrappers, all three gave the same sentence for every refusal case. The limits alone therefore did not rule out cel-go or expr-lang. The pre-check is a second reader of the formula text that has to follow each engine's syntax.
- The time limit is measured in the spike as a host-side safety. Under decision 19 the library has no time limit of its own. In the spike the wrapper for cel-go and expr-lang abandons the running goroutine: right after the refusals the goroutine count was 3 against 2 before, and it was back to 2 after 22.25 ms (cel-go) and 2.27 ms (expr-lang). The own engine stayed at 2 (`limits-table.txt`, L08).

## Why each other candidate lost

Decisions 5, 6 and 7: the formula text is the stored form, it is saved in one standard shape so that parsing the printed form gives back the same tree, and the block view comes from the same text. Decisions 12 and 18: Check reads the text alone and covers limits, depth and cost, and a failing formula is saved as a draft. Decision 19: a runaway formula is stopped by a cost count, and the library sets no time limit.

cel-go lost on the tree and on the cost of translation.

- Decisions 5, 6, 7: it needs 3 translation steps and a reverse step. Of the 19 formulas the printed text equals the written one for 14. It folds 14 of 234 nodes when its tree is turned into the neutral tree. 4 of 25 bracket probes change the tree: `!!a` is printed `a`, `a && (b && c)` is printed `a && b && c` and reads back as another tree, and two probes with `==` and `<` read in another order than the own engine's. Provisional: the own engine's order is the one in the operator table of the language page.
- Decisions 12, 18: it enforces depth, nodes, cost and formula length itself. It does not enforce text length, list length or decimal places itself (wrapper or function layer), and a text of 1100 characters in a formula was not refused (`limits-table.txt`, I01). It gives parser sentences of its own (`invalid int literal`), not the limit sentence, until the shared code translates them.
- Decision 19: it counts cost natively through a cost model written in the spike. The time limit is a wrapper that leaves a goroutine behind for 22.25 ms.
- Adapter findings: operators are singleton functions, so the decimal operators are declared for the checker only and swapped at plan time in a decorator keyed by overload ID (`spike/engine/cel/operators.go`). `==` between a decimal and an integer and unary minus on a decimal are refused at compile. Every decimal operation boxes its result into an interface.
- Speed: median 583 ns/op over the 19 formulas against 206 for the own engine, and a compile of 63070 to 185225 ns against 672 to 7328 (run 2).

expr-lang lost on safety and on the tree.

- Decisions 12, 18: it has no depth limit and no formula length limit of its own. 1,000,000 unclosed `(` ended the process with a fatal stack overflow, which a recover cannot catch (`limits-table.txt`, P05). Its safety then depends on the shared pre-check being run first. Its cost is charged only by the `range` function of the limits tests (`spike/limits/helpers_test.go`) and by a memory budget set to the same number, so a formula of plain operators is not counted.
- Decisions 5, 6, 7: 3 translation steps; the lexer rejects `@` (`unrecognized character: U+0040 '@'`); `2f` is printed `2`, so the float marker is lost. It folds 43 of 263 nodes. 2 of 25 bracket probes change the tree. Printed text equals the written one for 16 of 19. Provisional: its printer keeps a bracket around `&&` inside `||` and adds a bracket around a conditional in the else branch, where the language page prints the fewest brackets.
- Adapter findings: a patcher `forgetCallTypes` clears the type of every call node and relies on the checker's internals (`spike/engine/exprlang/operators.go`). Operator and function arguments arrive as boxed `...any`. Unary minus on a decimal is refused at compile.
- Speed: it is the fastest on formula 19 and smaller than cel-go (553 lines against 734), with a median of 350 ns/op over the 19 formulas against 206 for the own engine.

Both lost on the same point under decision 7. The language is the owner's to define, and each engine has its own: its own operator order for `==` and `<`, no `@` in names, and a decimal literal that has to be wrapped so that `0.010` keeps its trailing zero. The own engine has none of these gaps: 0 translation steps, 0 folded nodes, 0 of 25 tree-changing probes and 25 of 25 printed as the standard text.

The own engine's one miss in the 19 (formula 18) is that it prints `!suspended && role != "guest" && (role == "officer" || paid_up && months_between(joined, on) >= 6)` without the brackets that `example-formulas.md` writes around `paid_up && months_between(joined, on) >= 6`. Provisional: the language page prints formula 18 the same way (its section on formula 18).

## What the choice costs

| Cost | Measured in the spike |
|---|---|
| Lines to maintain | 2186 non-test lines in `spike/engine/own/` (1749 before the limit code and 472 lines of limit code added). cel-go 734 and expr-lang 553, plus 642 lines of shared limit code and 128 and 64 lines of adapter code for limits (`spike/limits/`) that the own engine needs only in part (79 lines of adapter) |
| A parser to fuzz | The parser and the lexer are ours: 271 and 224 lines in the spike. The fuzz run on [fuzz-run.md](../3.%20implementation/fuzz-run.md) applies to them. The spike fuzzed nothing; its one crash (expr-lang) came from a plain test with 1,000,000 brackets |
| Every built-in wired by hand | 11 built-ins and the operators: 227 lines of function wiring in the own engine against 238 for cel-go and 176 for expr-lang. Each new built-in is a function, a type rule and a test |
| The language is ours | A new part of the language needs the lexer, the parser, the tree, the printer, the type rules, the evaluator and the block. The tree and printer files of the own engine are 195 lines (84, 17 and 94) |
| Fixes by nobody else | A bug in the reader or the evaluator is fixed in this repo, and a published function is never edited (decision 10), so a fix to a built-in ships as a new version. Provisional: the language page proposes that built-ins are versioned as one language version |

## Limits (proposed)

The Limits table of [hostile-formulas.md](../4.%20qa/hostile-formulas.md) filled from `Proposed()` in `spike/limits/limits.go`. Every value is proposed. The values on the hostile page today are those of the language page.

| Limit | Value (proposed) | On the hostile page today | The 19 formulas |
|---|---|---|---|
| depth | 64 | 12 | 8 |
| node | 1000 | 100 | 34 |
| cost | 1,000,000 | 20000 | not measured |
| formula length | 5000 characters | 500 characters | 175 characters |
| text length | 1000 characters in a text literal | 200 characters | not measured |
| number magnitude | 18 digits | 18 digits in an integer literal, 30 in a decimal literal | not measured |
| decimal places | 18 | 10 in a literal | not measured |
| list length | 10,000 items in a list value | 1000 items | not measured |

- Cost is the library's limit for a runaway formula (decision 19). Its unit is the engine's own count. In the spike the own engine charges 1 for each node it evaluates and the item count for each list argument; the cost of the 19 formulas was not measured in that unit. Provisional: the language page counts cost from the tree before any run (each node 1, plus list sizes, plus the callee's cost), and its value is 20000. The two counts are different units, so 1,000,000 and 20000 cannot be compared.
- There is no time limit in the library (decision 19). The spike's `Proposed()` has a time of 250 ms and its test runs used 30 ms (cases L06) and a 15 ms host deadline (L09). These are host-side safeties for the test of a host's cancel, not library limits.
- The spike's values are 8 times (depth), 29.4 times (nodes) and 28.6 times (formula length) the largest of the 19 formulas. Provisional: the language page's values are 1.5, 2.9 and 2.9 times. The owner chooses the margin.
- The spike has one setting for number magnitude (18) and one for decimal places (18), and no separate 30-digit setting for a decimal literal.

## Speed targets (proposed)

For the own engine, on the 19 formulas and the batch of 10,000 rows of the spike. All are proposed. The figures they rest on are the own engine's in run 2 and run 1 (`summary.txt`).

One formula, one compiled program, one thread:

| Target | Proposed | Measured, run 2 (run 1) | Headroom |
|---|---|---|---|
| ns/op, median over the 19 formulas | at most 450 | 206 (406) | 2.2 times run 2 and 1.1 times run 1 |
| ns/op, the slowest of the 19 | at most 1500 | 648 (2843), formula 12 | 2.3 times run 2. Run 1 does not meet it (formula 12 at 2843, formula 6 at 1583) |
| allocs/op, every formula | at most 25 | 25 (25), formula 12; median 4 (4) | none. The count was the same in both runs for every formula, so no noise to cover |

The headroom rests on the spread: the own engine's median spread in run 2 is 4.4 % and the largest 26.2 %, and run 1 against run 2 moved up to 4.39 times on one formula. The targets are for a machine with a 1-minute load average near run 2's median (5.98) or lower.

A batch of 10,000 rows, one compiled program:

| Formula | ns/row, one goroutine: proposed | Measured, run 2 (run 1) | ns/row, 8 workers: proposed | Measured, run 2 (run 1) |
|---|---|---|---|---|
| 8 | at most 225 | 141.5 (147.5) | at most 150 | 60.6 (67.7) |
| 3 | at most 725 | 474.8 (464.1) | at most 425 | 180.1 (204.3) |
| 15 | at most 675 | 433.8 (433.9) | at most 400 | 166.1 (193.5) |

One goroutine: 1.5 times the larger of the two runs, rounded up to a multiple of 25. The runs differ by at most 4.2 % and the spread within run 2 is 5.3 % to 16.7 %. 8 workers: 2 times the larger of the two runs, rounded up to a multiple of 25. The spreads of the parallel rows are larger (10.2 % to 37.5 % in run 2), and the speed-up with 8 workers was 2.3 to 2.6 times, not 8 times. The same targets in ns per batch are the figures above times 10,000: for formula 3, at most 7.25 ms with one goroutine and 4.25 ms with 8 workers.

## Where the spike differs from the language page and the decisions

Provisional: the language page may still change. The spike ran on the syntax of `example-formulas.md`.

| Part | The spike | The language page or the decision |
|---|---|---|
| Rounding | Half away from zero in all three engines (`round(-2.675, 2)` is -2.68, `round(-2.5, 0)` is -3). No other rule was run | Decision 14: halves round up. The language page: half away from zero. The two agree for positive numbers and not for negative halves, if "up" means toward positive infinity. Decision 20 proposes down, up, half up, half down and half even; none of the five is in the spike |
| Division digits | 18 places (`spike/numeric/decimal.go`, `divisionScale`), fixed, not an agency setting | Decision 21 proposes 20 and an agency setting. The language page gives 20. The fast path of the spike's decimal stops at scale 18 (`maxFastScale`), so 20 places would use the big-integer path; that is read from the code, not measured |
| Operator order | The own engine reads `a == b < c` as `a == (b < c)`. cel-go and expr-lang read it the other way | The operator table of the language page: `<` binds tighter than `==`, `!` binds tightest |
| Printing | The own engine prints formula 18 without the brackets of `example-formulas.md` | The language page prints formula 18 the same way and states the rule of the fewest brackets. The 14, 16 and 18 of 19 are counted against `example-formulas.md`, not against that rule |
| `&&`, `||` | The own engine skips the right side when the left decides | The language page evaluates both sides |
| Float literal | `0.75f` and `2f` are read | The language page leaves the float literal out |
| Unary minus | Read by the own engine (`-d` on a decimal gave `-2.00`), refused for decimals by cel-go and expr-lang | The language page has no unary minus |
| `range` | The hostile cases used `range(1, n)` for lists and cost | The language page has no `range`, `map`, `fold` or `join` |
| Limits | Depth 64, nodes 1000, formula length 5000, list length 10,000 | The language page: 12, 100, 500 and 1000, plus inputs 8 and call depth 4, which the spike has not |
| Cost | Counted while the formula runs | The language page counts from the tree in Check, before any run. Decision 18 asks Check to cover cost. The spike did not build a count from the tree |
| Tree form | A neutral tree written as text | The language page writes the tree as JSON |
| Calls | `late_days@1(due, filed)` and `interest_step@1(balance)` | The language page also has `fin.pmt@2(a, b, c)` with a library name. The spike did not run a library name |
| Functions | All 11 built-ins the page names (`round`, `sum`, `count`, `avg`, `min`, `max`, `to_float`, `to_decimal`, `days_between`, `months_between`, `years_between`) and `%` appear in the 19 formulas and ran | The same names. The page's rules for each (the result type of `avg`, the empty list, problem codes) were not compared one by one |
| Problem codes | A refusal is a sentence and the name of the limit | The language page gives each problem a code, a part and a path |

## Open questions for the owner

1. Is a host's own cancel or deadline honoured, and with which sentence? In the spike all three engines stopped a run at a host deadline of 15 ms (after 15.09 to 15.67 ms) and reported the time-limit sentence, but that is the spike's choice. This is open under decision 19 and case L09 of the hostile page waits on it.
2. Is cost counted from the tree in Check, while the formula runs, or both? Decision 18 says Check covers cost and decision 19 allows a stop at run time. The own engine of the spike counts only at run time.
3. Which limit values: the spike's (64, 1000, 1,000,000, 5000, 1000, 10,000) or the language page's (12, 100, 20000, 500, 200, 1000)?
4. Do halves round up toward positive infinity (decision 14 read strictly) or away from zero (the spike and the language page)? Negative halves differ.
5. Is the spike's own engine the starting point of the library's engine, or is the engine written again from the language page?
6. Does the division carry 18 places as in the spike, or 20 as decision 21 proposes, and is it a setting of the agency passed into the evaluator for every run?

The raw figures and the method: [engine-spike-results.md](../3.%20implementation/engine-spike-results.md).
