# Example cases: implementation

The case loader, the check of the design page against its case file, and the test that runs every case. Built for ISS-31 in the package `examplecases/`, with its data in `testdata/`. The design of the case table, the typed values and the rules behind them is on the QA page, [example-cases](../4.%20qa/example-cases.md). This page says what was built and where it differs from that design.

Status: the owner has not reviewed this page. The shapes marked proposed on the QA page stay proposals.

## What is built

| Where | What it is |
|---|---|
| `examplecases/value.go` | `Value`, the typed value notation: `int`, `dec`, `float`, `text`, `bool`, `date`, `list<int>`, `list<dec>`, `list<text>`. A value keeps its exact text and a decimal is never turned into a float. |
| `examplecases/loader.go` | `Case`, `Input`, `Expected`, `Load(folder)` and `Parse(file, text)`. A malformed file is `<file>:<line>: <reason>` and the loader never skips a row or a file. |
| `examplecases/page.go` | `ScanPage` and `CheckPages`, the check of a design page against its case file. |
| `examplecases/evaluate.go` | `Evaluate`: one case through `formula.Parse`, `formula.Plan` and `Program.Run`, giving the typed result or the problem codes in order. |
| `examplecases/library.go` | `StandIn`, the library the cases call (see below). |
| `examplecases/compare.go` | `Mismatch` and `Failing`: the comparison of an outcome with the expected cell. |
| `examplecases/held.go` | `LoadHeld`: the list of cases that are read but not run, with a reason. |
| `testdata/formula-language.cases.md` | The 111 cases of the language page. |
| `testdata/formula-language.held.md` | The 56 held cases and why. |

The tests are in `loader_test.go`, `page_test.go`, `data_test.go`, `cases_test.go` and `proof_test.go`. They run under `go test ./...` and so under `make test`. A failing case is found by its name, because the name is the name of the subtest.

## Loader behaviour

- It reads the files named `*.cases.md` in the folder, sorted by name, and the rows of each file in order, so the cases come back in the same order on every run.
- A case file is an optional first line `# <title>`, blank lines, the header row `| name | formula | inputs | expected | rule | page |`, the separator row, then one row per line. A pipe in a cell is written `\|`.
- Every reason of the QA page is used: `bad header`, `wrong column count`, `empty cell`, `bad name`, `duplicate name`, `bad input`, `bad expected`, `wrong page`, `too long`, `too many rows`. A reason may be followed by `: ` and a short detail. A name used twice is refused across the files of a folder as well as within one.
- `inputs` may be empty. Inputs are written `name=type:value` and separated by `, ` outside quotes and brackets, so a list value such as `list<dec>:[200.00, 250.00]` and a text value with a comma stay whole.
- The page of a row must be the page its file is named for: `formula-language.cases.md` holds rows of `documentation/2. design/formula-language.md`.
- The loader is tested with 36 refusals, each with its file, line and reason, with the order of the rows, with every notation, with the escaped pipe, and with a folder of two files.

## The check of a design page

`CheckPages(root, caseDir, cases)` reads each design page (the three named on the QA page, and any page a row names) and its case file and gives a list of failures. Only fences whose info string is `formula` are scanned, and the line right after the closing fence must be `case: <name>`. The three failures use the proposed message texts of the QA page, plus one that the QA page does not name:

| Failure | Message |
|---|---|
| Example with no row | `<page>:<line>: example "<name>" has no row in <case file>` |
| Example with no marker line | `<page>:<line>: example has no case line` |
| Row naming a missing page | `<case file>:<line>: row "<name>" names page "<page>" which does not exist` |
| Row naming a page with no such example | `<case file>:<line>: row "<name>" names page "<page>" which has no example marked "case: <name>"` |
| A marker used twice on one page (proposed) | `<page>:<line>: marker "<name>" is used twice` |

A page that does not exist with no case file passes. The check runs on the real language page: its 111 examples and the 111 rows match both ways.

## How a case is run

`Evaluate` reads the inputs as typed values, parses the formula, plans it with the inputs' types and runs it. A problem from Parse, from Plan (which runs Check) or from the run is the outcome, and only its code is compared, in order. A result is compared as its type and its exact text, so `12.5` does not equal `12.50` (open question 1 of the QA page, taken as proposed).

The library a case calls is `StandIn`: `late_days@1` and `interest_step@1` are programs planned from the formulas of the language page, `fin.pmt@2` is written for the test, and `fin.pmt@1` and `step_a@1` to `step_e@1` have signatures and no body (they exist so that Check can give `call-too-deep` and the version codes). It stands in until the ten library functions of ISS-27 exist. A case that names a library function that has no body here fails at run time with `unknown-function`.

## Notation for the definitions of the Validate rows (ISS-79, proposed)

A Validate row is checked against a function definition, which the case table has no column for. The definitions are in a second file beside the case file, `testdata/formula-language.definitions.md`, with two tables.

The first table has one row for each input and each check of a definition: `| id | part | name | detail | message |`.

| Part | Name | Detail | Message |
|---|---|---|---|
| `input` | the input's name | `<type>: <rule>; <rule>`, with the type written `decimal`, `integer`, `float`, `text`, `date`, `boolean`, `list of decimal` or `list of integer` | empty |
| `check` | the check's number | the check's formula | the check's message |
| `from` | the id of another definition | empty | empty |

- The rules are written as on the QA page and separated by `; `: `required`, `no-negatives`, `decimal-places 2`, `min-items 1`, `max-items 12`, `range 1 to 500`, `range more than 0.00` and `allowed-values "a", "b"`. An `allowed-values` with no value is a legal row, because a definition with no allowed value is itself a case.
- A definition whose rows begin with `from` copies the definition it names. The rows after it with the same part and name replace the copied ones. The six definitions that are refused on save are written this way, as the definition they come from with one change.
- The rows of one definition are together. The same input twice in one definition is refused.

The second table says which definition each case is checked against and what it checks: `| case | definition | target |`. The target `row` validates the row of the case's `inputs` cell, and runs the formula when the row is clean. The target `definition` runs `CheckDefinition` on the definition and compares its codes, with an empty `inputs` cell.

The loader is `LoadDefinitions` and `ParseDefinitions` in `examplecases/definitions.go`. A malformed file is refused as `<file>:<line>: <reason>`, with the reasons `bad header`, `wrong column count`, `empty cell`, `bad part`, `bad type`, `bad rule`, `duplicate name`, `definition split`, `unknown definition` and `bad target`.

## Results

- **Run and passing:** the 55 `lang-` cases, with their exact typed results or problem codes in order. No expected result was changed to make a case pass, and none failed.
- **Held:** the 56 `rules-` cases. They are Validate rows, and each is checked against an input definition (V1 to VB on [formula-examples](../4.%20qa/formula-examples.md)) that has no notation in the case file yet. That is the open point the QA page names (a definition column, or one definition file per page). They are listed in `testdata/formula-language.held.md`, shown as skipped with the reason in a test run, and kept in the case file. A held row that is not a case of the file is a failure.
- **Proofs, made in copies and never in the committed files:** changing one expected result in a copy of the case file makes exactly that case fail (five changes tried: a decimal, an integer, a problem code, a boolean and a float), and the unchanged copy has no failure. In a copy of the design page, a formula block with no case line, a marked example with no row, and a removed example whose row is left are each reported.

## Differences from the proposal

- **Package and folder.** The package is `examplecases/`, not `internal/casefile`, and the case files are in `testdata/`, not in `documentation/4. qa/cases/`. The QA page proposed both and left them open (its question 5).
- **Where the table lives twice.** The 111 rows are on the QA page [formula-examples](../4.%20qa/formula-examples.md) and in `testdata/formula-language.cases.md`, because a case file holds nothing but its table. A test holds the two equal, row by row, so a change to one without the other fails.
- **Formula length.** The cell holds up to 500 characters, not the 200 proposed, because two rows are longer (302 and 213 characters) and the language page's text limit is 500.
- **Types.** `float`, `list<int>`, `list<dec>` and `list<text>` are added, with the notation that formula-examples.md proposes. `list<text>` is read by the loader but has no value in the formula package, so `Evaluate` returns `ErrUnsupported` for it; only held rows use it.
- **Design page names.** The three design pages named on the QA page are `formula-language.md`, `batch.md` and `posting.md`. Only the first exists with examples. The batch page of the repo is `batch-evaluation.md`, so the check will look for `batch.md` until the owner settles the name.
- **Sentence under the title.** The sentence each design page is to carry, pointing at its case file, is not added: the pages are not part of this issue, and the case file is in `testdata/`, so the sentence would have to change first.

## Still open

The open questions of the QA page are not copied here: see [Open questions](../4.%20qa/example-cases.md#open-questions). What this build adds to them:

1. How a held Validate row is to be run: the definitions V1 to VB need a notation. When they have one, the held rows go to the run and the held file shrinks.
2. Whether the stand-in library should be replaced by the real functions when ISS-27 lands, or kept for the rows that name a version no release has.
3. Whether the table on the QA page should be generated from the case file, so that there is one copy.
4. The cases of the other test issues (laws, money, hostile formulas, basic runs) have their own tables on their QA pages and are not read by this loader yet.
