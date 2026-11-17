# Example cases: implementation

This page records how the case loader and its check are to be built. The design of the case table, the typed values and the rules behind them is on the QA page, [example-cases](../4.%20qa/example-cases.md). This page does not repeat it.

Status on 2026-11-17: nothing described here is built, and none of the three design pages exists on disk yet. The shapes below are a proposal until the owner reviews them.

## What is built

Nothing yet. The shapes the build will produce:

| Shape | Meaning |
|---|---|
| `internal/casefile` | The one package every test issue uses to read cases. The package name and path are proposed. |
| `Case` | One row. Holds the six columns (`name`, `formula`, `inputs`, `expected`, `rule`, `page`) plus the file it came from and its line number. |
| `Value` | A typed value. Holds its type (`int`, `dec`, `text`, `bool` or `date`) and its exact text. A `dec` is never turned into a float. |
| `Load` | Takes a folder and returns the cases of every case file in it, or the first malformed-file failure. |
| the check test | A Go test in the repo root that compares each design page with its case file. Runs under `go test ./...` and so under `make test`. |

The inputs of a case are an ordered list of name, type and text. The folder the loader is pointed at is a constant in `internal/casefile`.

## Loader behaviour

- It reads the case files in the folder sorted by name, and the rows of each file in order, so the cases come back in the same order on every run.
- It returns the cases, each with its file and line.
- A malformed file fails the whole run with `<file>:<line>: <reason>`. The loader never skips a row and never skips a file.
- The reasons, by name: `bad header`, `wrong column count`, `empty cell`, `bad name`, `duplicate name`, `bad input`, `bad expected`, `wrong page`, `too long`, `too many rows`. When each applies is on the QA page.
- A test ranges over the cases and uses the case `name` as the subtest name, so a failing row is found by searching the case file for that name.

Limits the loader enforces:

| Limit | Value |
|---|---|
| Columns per row | 6 |
| Rows per case file | 200 |
| Case name | 3 to 60 characters |
| Formula | 1 to 200 characters |
| Rule | 1 to 120 characters |
| Input name | 1 to 40 characters |
| `int` | at most 18 digits |
| `dec` | at most 30 digits in all |
| `text` | at most 200 characters |
| Problem code | 1 to 40 characters |

## The check

For each of the three design pages the check reads the page and its case file and compares them.

- It scans the page for fenced blocks whose info string is `formula`. Other fences and prose are not scanned.
- The line right after the closing fence is the marker: `case: ` and the case name, with no blank line between.
- A name used as a marker twice on one page is a malformed page.

The three failures, by name:

| Failure | Where it is reported |
|---|---|
| Example with no row | the design page and the line of the example. A `formula` block with no marker line reports `example has no case line`. |
| Row naming a missing page | the case file and the line of the row |
| Row naming a page with no such example | the case file and the line of the row |

When a page or case file is absent: until a design page exists, its case file may be absent and the check passes for that page. Once the page exists with an example, a missing case file is the first failure for each example. The message texts are proposed on the QA page.

## Where the shape is recorded and who points at it

The shape is recorded on the QA page. Each of the three design pages points at its case file with one sentence under its title.

| Design page | Case file | Status |
|---|---|---|
| `documentation/2. design/formula-language.md` | `documentation/4. qa/cases/formula-language.cases.md` | page not written yet, sentence waiting |
| `documentation/2. design/batch.md` | `documentation/4. qa/cases/batch.cases.md` | page not written yet, sentence waiting |
| `documentation/2. design/posting.md` | `documentation/4. qa/cases/posting.cases.md` | page not written yet, sentence waiting |

The sentence each page must carry:

| Design page | Sentence |
|---|---|
| `documentation/2. design/formula-language.md` | Every example on this page has a row in `documentation/4. qa/cases/formula-language.cases.md`, marked by the `case:` line under the example. |
| `documentation/2. design/batch.md` | Every example on this page has a row in `documentation/4. qa/cases/batch.cases.md`, marked by the `case:` line under the example. |
| `documentation/2. design/posting.md` | Every example on this page has a row in `documentation/4. qa/cases/posting.cases.md`, marked by the `case:` line under the example. |

## Status

Nothing of the loader or the check is built. The package name, the shapes and the message texts are a proposal until the owner reviews them. The questions still open are listed on the QA page under [Open questions](../4.%20qa/example-cases.md#open-questions); they are not copied here.
