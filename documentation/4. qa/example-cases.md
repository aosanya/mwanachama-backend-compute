# Example cases

Every example formula on a design page is also a row a test loads, so the docs and the tests cannot disagree. This page designs the one case table and the one loader that every test issue shares: laws, money, hostile formulas, reference values, input checks and speed.

The three design pages are named by their intended file names: `documentation/2. design/formula-language.md`, `documentation/2. design/batch.md` and `documentation/2. design/posting.md`. They do not exist yet.

## Decided by the owner

- One table shape for every case, with the six columns below.
- One plain file per design page, so a reviewer reads the same file the test runs.
- A check fails when a design page has an example with no row, or a row names a page that no longer has it.
- Every figure is invented. The repo is public: no real names, emails, organisations or figures.

Everything else on this page is a proposal until the owner confirms it. Limits are proposals too.

## The columns

| Column | Required | What it holds |
|---|---|---|
| `name` | yes | Unique id of the case. Lowercase letters, digits and hyphens, starting with a letter, 3 to 60 characters. Unique across all case files, not only within one. |
| `formula` | yes | The formula as one line of text, 1 to 200 characters. A literal pipe is written `\|`. |
| `inputs` | yes, may be empty | The named values the formula reads. Empty means the formula reads none. |
| `expected` | yes | Either one result or one problems list. Never both, never neither. |
| `rule` | yes | The one rule the case shows, as one sentence of 1 to 120 characters. |
| `page` | yes | The repo-relative path of the design page the example comes from. |

## Typed values

An input is written `name=type:value`. Several inputs are separated by `, ` outside quotes. An input name is lowercase letters, digits and underscores, starting with a letter, 1 to 40 characters, and appears once per row.

| Type | Written as | Notes |
|---|---|---|
| `int` | `int:3`, `int:-7` | Whole number, optional leading minus, no thousands separator, at most 18 digits. |
| `dec` | `dec:12.50`, `dec:0.05` | Digits with an optional dot and optional leading minus. Read as text and kept exact: never a float, no exponent. The digits after the dot are part of the value, so `12.5` and `12.50` are written as they are. At most 30 digits in all. |
| `text` | `text:"red"` | Double quotes always. A quote inside is `\"`, a backslash is `\\`. At most 200 characters. |
| `bool` | `bool:true`, `bool:false` | Lowercase only. |
| `date` | `date:2026-03-12` | `YYYY-MM-DD`, a real calendar day, no time and no zone. |

The same syntax writes an expected result. A new type is added here first, then in the loader, never the other way round.

## Expected result and expected problems

| Written as | Meaning |
|---|---|
| `result: dec:37.50` | The formula evaluates, and the value equals this typed value, type and text included. |
| `problems: unknown-name` | The formula does not evaluate, and the problems reported are exactly this list. |
| `problems: unknown-name, divide-by-zero` | Several problems, compared in the order the library reports them. |

A problem is written by its code: lowercase letters, digits and hyphens, 1 to 40 characters. Codes are not message texts. The list of codes is owned by the formula language page; a row may only use a code that page defines.

## Case files

One case file per design page, in one folder.

| Design page | Case file (proposed) |
|---|---|
| `documentation/2. design/formula-language.md` | `documentation/4. qa/cases/formula-language.cases.md` |
| `documentation/2. design/batch.md` | `documentation/4. qa/cases/batch.cases.md` |
| `documentation/2. design/posting.md` | `documentation/4. qa/cases/posting.cases.md` |

The file name is the design page name, then `.cases.md`. The `page` column of every row in a file must be the page that file is named for; any other value is a malformed file. A file holds at most 200 rows.

A case file is a markdown file, so a reviewer reads it rendered or raw. It holds, in this order: an optional first line `# <title>`, blank lines, the header row, the separator row, then one row per line. The header row is exactly `| name | formula | inputs | expected | rule | page |`. Nothing else may sit between the header and the last row, and a row is one line.

Rows as they appear in a file, with invented figures. The formula syntax shown is illustrative until the formula language page fixes it.

```
| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| line-total | qty * price | qty=int:3, price=dec:12.50 | result: dec:37.50 | Multiplying a whole number by a decimal keeps the decimal exact. | documentation/2. design/formula-language.md |
| even-share | round(total / people, 2) | total=dec:100.00, people=int:3 | result: dec:33.33 | Rounding to two places happens once, at the end. | documentation/2. design/formula-language.md |
| late-fee | if(paid_on > due_on, fee, 0.00) | paid_on=date:2026-03-12, due_on=date:2026-03-05, fee=dec:5.00 | result: dec:5.00 | A later date compares greater than an earlier one. | documentation/2. design/formula-language.md |
| rate-missing | qty * rate | qty=int:3 | problems: unknown-name | A name with no input is a problem, never a zero. | documentation/2. design/formula-language.md |
| zero-people | total / people | total=dec:10.00, people=int:0 | problems: divide-by-zero | Dividing by zero is a problem, never a value. | documentation/2. design/formula-language.md |
```

## How an example links to its row

An example on a design page is a fenced code block whose info string is `formula`. The line right after the closing fence is the marker: `case: ` and the case name. Nothing else, no blank line between.

````
```formula
qty * price
```
case: line-total
````

Only `formula` blocks are examples. Other fences and prose are not scanned. A name used as a marker twice on one page is a malformed page.

## The check

It is a Go test in the repo root, run by `go test ./...`, so `make test` runs it. For each of the three design pages it reads the page and its case file and compares.

| Failure | Message (proposed text) |
|---|---|
| Example with no row | `<page>:<line>: example "<name>" has no row in <case file>`. A `formula` block with no marker line reports `example has no case line`. |
| Row naming a missing page | `<case file>:<line>: row "<name>" names page "<page>" which does not exist` |
| Row naming a page with no such example | `<case file>:<line>: row "<name>" names page "<page>" which has no example marked "case: <name>"` |

Until a design page exists, its case file may be absent, and the check passes for that page. Once the page exists with an example, a missing case file is the first failure for each example.

## The loader

One package, proposed as `internal/casefile`, used by every test issue. It reads a folder and returns the cases, each with the six columns plus its file and line, inputs as an ordered list of name, type and text. Tests range over the cases and use `name` as the subtest name, so a failing row is found by searching the case file for that name.

A malformed file fails the whole test run. The message names the file and the line, as `<file>:<line>: <reason>`, and the loader never skips a row or a file. Reasons it reports:

| Reason | When |
|---|---|
| `bad header` | The header row is not exactly the six titles in order. |
| `wrong column count` | A row does not have six cells. |
| `empty cell` | Any required cell is empty. |
| `bad name` | A case or input name breaks its pattern or length. |
| `duplicate name` | A case name is used twice across the folder, or an input name twice in a row. |
| `bad input` | An input is not `name=type:value`, or a value does not parse as its type. |
| `bad expected` | The cell is neither `result: ` nor `problems: `, or has both, or a problem code breaks its pattern. |
| `wrong page` | The `page` cell is not the page the file is named for. |
| `too long` | A limit above is exceeded. |
| `too many rows` | A file has more than 200 rows. |

## Sentence each design page carries

Each page carries exactly one sentence pointing at its case file, placed under its title.

| Design page | Sentence |
|---|---|
| `documentation/2. design/formula-language.md` | Every example on this page has a row in `documentation/4. qa/cases/formula-language.cases.md`, marked by the `case:` line under the example. |
| `documentation/2. design/batch.md` | Every example on this page has a row in `documentation/4. qa/cases/batch.cases.md`, marked by the `case:` line under the example. |
| `documentation/2. design/posting.md` | Every example on this page has a row in `documentation/4. qa/cases/posting.cases.md`, marked by the `case:` line under the example. |

## Open questions

1. Does a decimal result compare by exact text or by value? Proposed: exact text, so `12.5` does not equal `12.50`, because for money the scale is part of the answer.
2. Can a case that is not an example on a design page (a hostile formula, a speed case) live in these files? Proposed: yes, if it names the design page whose rule it tests, and that page carries a matching `formula` example. A case with no page that owns its rule waits for that page.
3. Do batch and posting need list or record values? Proposed: no list type now. When the batch page needs one, the owner decides the syntax here first.
4. Where do problem codes live? Proposed: a table on the formula language page, and the loader fails a row that uses a code not in it.
5. Is the case folder `documentation/4. qa/cases/` right, given `documentation/` holds only the four phase folders? Proposed: yes, it is a subfolder of the QA phase, and the loader path is a constant in `internal/casefile`.
6. Should the check also fail on a case file with rows but no design page yet? Proposed: no; it fails only once a row names a page that does not exist (the second failure), which is the same moment.
