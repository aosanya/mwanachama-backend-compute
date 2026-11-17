# The basic run suite

Built for ISS-38, in the folder `basicruns_test/`, package `basicruns`. It reads [basic-runs.md](../4.%20qa/basic-runs.md), the QA page that lists every part of the formula language run alone, and runs each case through the formula package of [formula.md](formula.md) and [formula-run.md](formula-run.md). The language is in [formula-language.md](../2.%20design/formula-language.md).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## What it does

The suite has no list of cases of its own. The page is the list. A change to a formula, an input or an expected result on the page changes the test; a row added to the page is run the next time `go test` runs.

| Step | What reads what |
|---|---|
| Load | `loader_test.go` reads every table of basic-runs.md that has a `Case` column, the shared formula of section 12, the held table and the numbered Gaps list. It reads the rows of [formula-examples.md](../4.%20qa/formula-examples.md) by name. |
| Resolve | `resolve_test.go` sorts each row: it has its own formula (`basic-` rows), it points at another row (a `basic-` name with no formula), it names a row of formula-examples.md (`lang-` rows), it names a row of another QA page, or it is `none` (a gap). A name that no page holds is an error. |
| Run | `execute_test.go` reads the formula, inputs and expected cell, plans the formula with the typed inputs and runs it. The result is written the way the page writes it (`dec:3.50`, `float:3`, `text:"low"`). `problems:` lists the codes of Plan, or the code of the one problem of a run, in order. |
| Cover | `coverage_test.go` lists 46 parts of the language and checks each against the cases that run. |

The inputs are typed values (`int:7`, `dec:12.50`, `list<dec>:[1.00, 2.00]`), the way a host would hand them over after reading its inputs. Agency functions called by name (`late_days@1`, `interest_step@1`) are planned from their formulas in `runtime_test.go`, so those cases run for real.

## What runs and what is held

The run of `go test ./basicruns_test/ -race -v` has 162 cases that run and pass, and 25 that are listed as held. A held case is shown as skipped with its reason; it is never dropped.

| Held | Count | Reason |
|---|---|---|
| The ten library calls of section 10 and `lang-library-call-version-two` | 11 | The library functions are not built (ISS-27). A case that expects a problem from a library call still runs, because Check needs the signature and not the function. |
| `pmt-version-1-loan`, `same-host-newer-minus-older`, `old-version-still-served` | 3 | Cases of [library-versioning.md](../4.%20qa/library-versioning.md) that need the registry. |
| `basic-pin-after-new-version` | 1 | A library call of version 1 with a result. |
| `basic-call-chain-at-limit` | 1 | Gap 9: `step_a@1` to `step_e@1` have no definition. |
| `rules-no-negatives-invalid`, `rules-required-missing`, `rules-decimal-places-invalid` | 3 | They read text into values: Validate, ISS-26. |
| The six rows of the page's held table | 6 | The agency's division digits and rounding rule have no shape yet (decisions 20 and 21). |

When ISS-27 and ISS-26 land, the rows move out of the held list by removing the rule in `heldReason` in `resolve_test.go`; no case is rewritten.

## The coverage check

`TestEveryPartOfTheLanguageHasACaseOrIsAListedGap` lists 46 parts: the 13 binary operators, `!`, the conditional, a list written in a formula, an agency call, a library call, a call with a version, the 11 built-in functions and the eight types as an input and as a result. A part is covered when a case that runs and gives a result uses it.

| Outcome | What the check does |
|---|---|
| Covered | 42 parts. |
| Only in a held case | The library call. It is reported and does not fail the check. |
| No case, and not a gap | Fails, naming the part. |
| A gap listed in the test that now has a case | Fails, so the gap is taken out. |
| A gap number that the page does not list | Fails. |

The three parts with no case are the date result and the two list results, held by Gaps 3 and 4 of the page.

Two tests prove the check by working on copies in memory. Removing every row that uses `%` makes the remainder operator fail it, and removing Gap 3 from the page makes the date result fail it. Two more tests prove the run: a changed expected figure and a changed problem code each fail exactly that case.

## Result

Every runnable case passed on the first run, so no engine bug was filed and no figure of the page was found wrong. That includes the figures that depend on open questions: a negative half rounds away from zero (Gap 5) and an empty sum is `0` with no places (Gap 6), both as the language page writes them.

## Proposed on this page

- The coverage list and the gap numbers of the three open parts are in the test, because the language page has no machine-readable list of its parts. A new part of the language needs a line in `languageParts`.
- A `problems:` cell is read as the codes Plan gives, or the one code of a failed run. The page does not say which of the two a code belongs to.

## Open

- The page's three open questions are unchanged: whether the loader reads the `basic-` rows here or they move into formula-examples.md (this suite reads them here, so there is one reader and no copy), the `1e200` notation (it ran as written) and the worked case with the whole number `0`.
- Gaps 1 to 12 of the page wait on the owner.
- The `rules-` rows of section 12 need Validate. The same text-to-value reading is tested for the other rows on [input-checks.md](../4.%20qa/input-checks.md).

## Files

| File | What it holds |
|---|---|
| `package.go` | The package name. |
| `loader_test.go` | The reader of the page tables, the held table and the Gaps list. |
| `values_test.go` | The notation of values: `int:`, `dec:`, `float:`, `text:`, `bool:`, `date:`, `list<int>:`, `list<dec>:`. |
| `runtime_test.go` | The signatures the checker needs and the agency functions planned from their formulas. |
| `execute_test.go` | Plan and run of one case, and the judging of its expected cell. |
| `resolve_test.go` | The sorting of rows, the held rules and their reasons. |
| `cases_test.go` | The run of every case, the held list and the checks of the page itself. |
| `coverage_test.go` | The coverage check and its proofs. |
