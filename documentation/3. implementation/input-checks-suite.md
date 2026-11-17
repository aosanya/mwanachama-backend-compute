# Input checks suite

Folder `inputchecks_test/`, built for ISS-37. It runs the cases of [input-checks.md](../4.%20qa/input-checks.md) against the `function/` package ([function.md](function.md)) and compares each problem list field by field and sentence by sentence. The rules are decisions 8, 18 and 22 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page or the sentences it checks (TD-2982 of ISS-14). All names and figures are invented.

## What it is, and how it differs from the tests in `function/`

`function/` has its own tests for sections A to M, written by the builder of the package with helpers of that package. This suite is a second, independent reading of the same page. It differs in three ways:

| In `function/` tests | In this suite |
|---|---|
| Expected problems are built with Go constructors (`on`, `at`, `missing`) that fill the rule from a table | Expected problems are the page's own lines, `code · input · value — sentence`, read by a small notation (`line` in `notation_test.go`), so the page's wording is the only source of a sentence and a rule |
| Checks run on a small runner written for the tests, which knows a few operators | Checks run on the real formula run of ISS-25 (`formula.Plan` and `Program.Run`) through a 30-line adapter in `engine_test.go`, so the division by zero of `ic-h07` is the run's own sentence |
| The package's own definitions D1 to D11 are shared with the cases | Definitions D1 to D11 are written again from the page's tables and each is also given to `CheckDefinition` |

The overlap is the whole of sections A to M: every case here is also a case there. The suite is meant to fail when one of the two readings of the page is wrong and the other is not.

## What it holds

| File | What it is |
|---|---|
| `notation_test.go` | the notation, the eleven definitions with their clean rows, and tests that each definition is accepted on save and that its clean row gives an empty list |
| `engine_test.go` | the adapter that turns the typed values of `numtext/` into values of `formula/` and runs a check on the real run |
| `run_test.go` | the case type, how a case is compared (the whole list with `reflect.DeepEqual`, and the number of checks that ran), the held cases and their listing |
| `cases_a_to_g_test.go` | sections A to G: required, list sizes, no-negatives, range, decimal places, allowed values, a rule on each item |
| `cases_h_to_j_test.go` | sections H to J: each check, several problems and their order, missing against empty against space-only |
| `cases_k_to_m_test.go` | sections K and L (the texts of each type, built from the page's tables), and `ic-m01` to `ic-m04` |
| `entrypoints_test.go` | `ic-m05` and `ic-m06`: one list from every way in, and the same list in a second process |

The folder holds test files only. Its package is `inputchecks`. Run it with `go test ./inputchecks_test/ -race`.

## How a case is compared

- The whole list is compared: the same number of problems, in the same order, each with the same `input`, `item`, `given`, `value`, `rule`, `sentence`, `code` and `check`. A missing problem, an extra one, one out of order or a sentence worded differently fails the case. `TestTheComparisonSeesAWrongSentenceAnOrderAndAMissingProblem` shows the comparison does fail on each.
- A clean row must give an empty list that is not `nil`.
- Where the page says which checks run and which are skipped, the case also counts the runs of the check runner: `ic-a02` runs none, `ic-a03` one, `ic-h04` two. A skipped check is therefore proved skipped, not only absent from the list.
- A text of a table in section K or L is one case on D8 with one input changed. The refused sentence is the page's template for the type, with the text as received.

## What was run and what it showed

`go vet ./inputchecks_test/` is clean and `go test ./inputchecks_test/ -race -count=1` passes:

| Group | Case tests |
|---|---|
| A required | 8 |
| B list sizes | 12 |
| C no-negatives | 7 (1 held) |
| D range | 17 (1 held; `ic-d10` is two rows) |
| E decimal places | 9 |
| F allowed values | 4 |
| G a rule on each item | 7 (1 held; the 3000-problem `ic-g06` included) |
| H each check | 15 (1 held) |
| I several problems | 7 |
| J missing, empty, space-only | 12 |
| K a number written as text | 38 (3 held) |
| L the other types | 22 |
| M never calculated, same list | 4 cases (`ic-m01` to `ic-m04`) and 3 tests (`ic-m05`, `ic-m06` twice) |

In all, 206 tests and subtests pass and 8 are skipped: the 7 held cases, and the helper test that only runs inside the second process of `ic-m06`. No sentence of `function/` differed from the page, so no issue was raised against it.

Two things were added to the page's cases, both marked with a suffix:

- `ic-g03-s` and `ic-h11-s` are variants of `ic-g03` and `ic-h11` that use a value above the range or at its excluded bound instead of a negative value, so the item order and the skipped checks stay covered while the originals are held.
- `ic-m05` is run on `ic-g03-s`, not `ic-g03`, for the same reason.

## Held cases

Cases that give a negative value to an input that does not declare `no-negatives` read the page as written and would change if decision 22 is built as written. They are listed by `TestHeldCasesAreListedWithTheirQuestions` and skipped, with the question in the skip message:

| Case | What it holds | Question |
|---|---|---|
| `ic-c07` | `price_per_kilo` of -0.5 on D2 | decision 22 against the no-negatives rule (input-checks open question 1) |
| `ic-d11` | `share_pct` of -0.01 on D10 | the same |
| `ic-g03` | `scores` [-3, 104, abc] on D4 | the same |
| `ic-h11` | `principal` of -1.00 on D6 | the same |
| `ic-k-int-neg-1` | `kilos` of -12 accepted on D8 | the same |
| `ic-k-dec-neg-1` | `payment` of -12.5 accepted on D8 | the same |
| `ic-k-flt-neg-1` | `ratio` of -1.5 accepted on D8 | the same |

## A field that is not compared

`ic-l-text-1` (201 letters in `note`): the page gives the sentence, cut to 200 characters and `...`. It also says the `value` field holds the text exactly as received, but its cases on lists show a cut value (`ic-b08`). It does not say which holds for a long text, and `function/` cuts the value. The case compares everything but `value` and logs why. This is the new question 11 below.

## Entry points

`ic-m05` takes `ic-c04`, `ic-g03-s`, `ic-h04` and `ic-i04` and compares the list from:

- a single run: `function.Validate(definition, row, runner)`;
- a batch row: one `function.Prepare(definition)` used for the row three times, the way a batch uses one `Validator`;
- a tool answer: the list written with `problem.ListJSON` and read back, which is how a list travels in a tool answer.

The lists are equal as values and as JSON. `ic-m06` shows two calls in one process and one call in a second process (the test binary starts itself with a switch and prints the lists) give the same text.

## What is proved about never calculating

The library does not calculate: the host does. `ic-m01` to `ic-m04` therefore use a small host in the test (`answer`): it validates, and only an empty list reaches a calculation (`balance - payment` on the real run). The cases show the contract the host must keep: `ic-m02` and `ic-m03` give the problem list and no result, the calculation counter stays at 0, and in `ic-m02` no check is run because both read the broken input. A batch of three rows (`ic-m04`) calculates row 1 only and keeps rows 2 and 3 with their own lists.

## What stays open

These are the owner's. The suite is built on the page's reading and says so.

1. Decision 22 against the no-negatives rule, as above (input-checks open question 1). When it is settled, the seven held cases and every case that has a range but no `no-negatives` change or are lifted.
2. How an empty or space-only value reads in a sentence (question 2). `ic-j05`, `ic-j06`, `ic-j07` and `ic-j12` expect `` `payment` is . `` as the page words it.
3. Whether `...` is inside the quotes for a cut text (question 4). `ic-l-text-1` expects it inside.
4. The wrapper of a batch row and of an AI tool answer (question 6). Only the list is compared; no tool call is built and the batch package (ISS-29) is not used.
5. The owner's reading of the sentences (question 10 and TD-2982 of ISS-14).
6. **New, question 11.** The `value` field of a text over 200 characters: the text as received (page, section "How a case is written") or cut like the sentence and like a list. `function/` cuts it.
7. Items after the 1000th of a list and a check that reads a list not given (questions 7 and 8): no case is written.
8. The seven held cases wait for 1.

## Index row

To add to [README.md](README.md): `| input-checks-suite.md | the independent suite of the input check cases, the held cases and what stays open (ISS-37) |`
