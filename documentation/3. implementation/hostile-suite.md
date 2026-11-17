# The hostile formula suite

Built for ISS-34. It runs the cases of [hostile-formulas.md](../4.%20qa/hostile-formulas.md) as Go tests against the formula front end ([formula.md](formula.md)) and the run ([formula-run.md](formula-run.md)), and adds the fuzz run described in [fuzz-run.md](fuzz-run.md). The limits are the language page's proposed values: formula length 500, depth 12, brackets open 12, nodes 100, text literal 200, integer 18 digits, decimal 30 digits and 10 places, call depth 4, cost 20000. The decisions it protects are 12, 18 and 19 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. All figures and sentences are invented or proposed.

## What it holds

The folder is `hostile_test/`, one Go package named `hostile` made of test files only.

| File | What it is |
|---|---|
| `helpers_test.go` | a small library of invented functions (`fin.pmt@1`, `step_a@1` to `step_e@1` with call depths 0 to 4), the inputs and values the cases share, and the builders of nested, long and many-item text |
| `save_cases_test.go` | the cases that are refused when the formula is saved: groups P, L, C, I and O. Each case names its id, its text, the code and the exact sentence (or the start of it) |
| `run_cases_test.go` | the cases that run: the normal results at the limits, the run-time stop at the cost limit, ten hostile formulas at once, the loops through a host, the same result on every run, and the sum of many decimals in two processes |
| `coverage_test.go` | the list of held cases, and the checks that every row of the QA page is built or held with a reason, and that no file of the suite imports `time` |
| `fuzz_test.go` | `FuzzParse`, the seeded generator of hostile text, and the properties both check |
| `testdata/fuzz/FuzzParse/` | the seed corpus in Go's own layout |

Imports: the standard library and the packages of this module. Nothing reads a clock, and a test reads the imports of every file to keep it so.

## Commands

| Command | What it does |
|---|---|
| `go test ./hostile_test/ -race` | the whole suite, and a replay of the seeds and of saved failures. This is part of `make test` |
| `make fuzz` | `FuzzParse` for 60 seconds with `GOMEMLIMIT=1GiB`, before a merge |
| `make fuzz-nightly` | the same for 30 minutes |
| `HOSTILE_SEED=<n> go test ./hostile_test/ -run Generated -v` | replays the generator with a seed. Without the variable the seed is 20261117 and is written in the log either way |

`make test` does not search for new inputs, so it stays at a few seconds.

## What the cases check

A case in a save-time table passes when `Parse` and then `Check` refuse the text with the first problem having the code and the sentence the case names, and the sentence is one plain line with no trace in it. A run case passes when the program gives the normal result, or the one problem the case names. After every case the same engine still runs `2 + 3` and gives `5` (L07), and the number of goroutines is the same before and after, with the heap within 32 MiB (L08).

Cost is checked in the two places the run page names. A formula whose count is over 20000 is refused at save with `This formula goes over the cost limit of 20000.` A formula that passes by count and meets a list longer than the count assumed (the length of a list that is not written in the formula counts as 1000) is stopped at run time with the same sentence. A test holds the two sentences equal, and another that twenty runs spend the same count.

## How the 77 rows of the page are covered

| Kind | Rows |
|---|---|
| Built and passing | P01, P02, P04 to P20, L01 to L05, L07, L08, L10, C01, C03, C05, C07, C09, C11, C02, C04, C06, I01, I03, I04, I06 to I11, O01, O04 to O16 |
| Built, failing, tracked by an issue | P03 (ISS-75), O02 and O03 (ISS-74), C10, C12 and C13 (ISS-72) |
| Held, with the reason | L06, L09, C08, I02, I05, I12, I13, B01 to B05 |

A failing row is skipped with the issue code in the message. If it starts to pass, the suite fails and says to remove it from the list, so the list cannot go stale.

### Held

| Row | Reason |
|---|---|
| L06 | withdrawn on the page: there is no time limit (decision 19) |
| L09 | open (decision 19): whether and how a host's own cancel is honoured |
| C08 | a callee must already be published, so no chain past the call depth can be built to slip past the save check. C10, C12 and C13 cover a host that loops |
| I02, I05, I12 | an input value that is too long or too large is refused by Validate, which ISS-26 builds. The run itself has no list length limit |
| I13 | the language has no `join`, and the list length limit of a value belongs to Validate |
| B01 to B05 | batch evaluation is ISS-29 |

### Where the expected result is not the page's sentence

Where a case uses a part that is not in the language (`range`, `map`, `fold`, `repeat`, `join`, `sort`, `^`, an exponent, a position in a list, a field of a list item), the case expects what the language page's Differences to settle 8 to 13 say: `unknown-function` for a call to a name that is not known, `bad-syntax` for a text that cannot be read, `wrong-argument-type` for `round(1.00 / 3, 1000000)`. The rows C03 and C05 also have a second case in this language that gets the same intent, a formula of 21 and of 24 `sum(rows)` over an input list, which is refused at save for its cost. Three more differences follow the language page:

- P05 has three cases. Eight brackets and a 1 give `missing closing bracket` at column 10. Three hundred brackets give the depth sentence, because Parse stops at the 13th open bracket first (difference 9). Eight brackets alone give a plain syntax sentence.
- P15 and P16 give columns 20 and 19, one past the last character (difference 10).
- A text over 500 characters is refused first with the formula length sentence, so a very long run of digits (P09, I04, I09) gives that sentence and a short one over the limit gives the magnitude or places sentence (P09, P10, I09).

## Failures found

| Issue | What | Rows |
|---|---|---|
| ISS-72 | A host that calls a program from its Runtime in a loop is never stopped. `Program.Apply` makes a new meter for each call and `Runtime.Apply` carries none. The test runtime had to stop the loop itself after 2001 calls | C10, C12, C13 |
| ISS-73 | A lone carriage return was read as a blank, and the page says it is an unexpected control character. Fixed: Parse refuses a lone carriage return, and P14 passes for both control characters | P14 |
| ISS-74 | `now()` and `random()` give `unexpected character )` because a call with no argument is not in the language. The hostile page and the language page's list of codes say `unknown-function`. The owner decides | O02, O03 |
| ISS-75 | A NUL byte gives `unexpected control character`, the page says `unexpected NUL character`. The owner decides | P03 |

No expected sentence was changed to make a case pass.

## The fuzz run

`FuzzParse` takes a text and runs it through `Parse`, `Check`, `Plan` and `Run` on five invented inputs. Four properties are checked on every input:

| Property | Fails when |
|---|---|
| It never panics | any step panics |
| It does bounded work | a valid text over 500 characters is not refused as too long; `Check` lists more than 20 problems; `Check` passes a formula whose count is over the cost limit; a run that gives a value spent more than the limit |
| A refusal is a plain sentence | a sentence has a line break or looks like a trace, or a syntax refusal has no line and column, or the sentence and the position fields differ |
| It is deterministic | the same text gives a different outcome the second time |

The seeds are 29 texts in the code (formulas in the style of the examples, near-parse cases and the hostile groups) and 12 files in `testdata/fuzz/FuzzParse/`. `TestGeneratedHostileTextHoldsEveryProperty` builds 3000 texts from a seeded generator of pieces (brackets, operators, names, long runs, control characters, bidirectional marks, an emoji) and checks the same properties, so `make test` has a fixed search of its own.

Two short runs were made on a busy machine (load average above 60): 60 seconds gave 273,809 executions and 30 seconds with four workers gave 67,041, with no panic, no crash and no failing input. The count stopped for many seconds at a time, and replaying all 216 inputs the run had kept took under 50 milliseconds each, so the stalls were the load and not an input. A quiet-machine run and the 30-minute nightly run are still to be done.

## Open

- A host's own cancel (L09) and how a callee is metered (ISS-72).
- Where the nightly run is scheduled. The repo has no CI and no remote.
- Go's fuzzer keeps the failing input as a file and has no seed to log, so the seed of the page is logged only for the generator. A failing fuzz input is committed as a file in `testdata/fuzz/FuzzParse` and added as a row of the QA page.
- B01 to B05 and I02, I05, I12 wait on ISS-29 and ISS-26.
