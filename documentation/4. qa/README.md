# QA

Test designs and recorded cases. Most pages are drafts: each opens with its own status line, and a number marked proposed is a starting point, not a decision.

## The shared case table

| Page | What it holds |
|---|---|
| [example-cases.md](example-cases.md) | The one case table and the one loader that every test issue shares, so the docs and the tests cannot disagree |
| [formula-examples.md](formula-examples.md) | One case row for every example on the formula language design page |

## What must always hold

| Page | What it holds |
|---|---|
| [laws.md](laws.md) | The laws the formula engine must keep on every input, each tested on generated inputs |
| [exact-money.md](exact-money.md) | Test design for exact money: no balance, interest or share is ever off by a fraction |
| [hostile-formulas.md](hostile-formulas.md) | Cases written as an attacker would try them, so a formula can never hang, crash or exhaust a host |
| [input-checks.md](input-checks.md) | A person or an AI that gives a function bad inputs is told everything that is wrong, in one answer, in words they can act on |

## Running formulas

| Page | What it holds |
|---|---|
| [basic-runs.md](basic-runs.md) | Every part of the language run once on its own, as a coverage table with its gaps listed |
| [bulk-runs.md](bulk-runs.md) | Test design for a bulk run such as month-end interest: correct for every row, and a restart changes nothing |
| [batch-equals-single.md](batch-equals-single.md) | The test that a batch result equals running each row alone |
| [concurrency-and-speed.md](concurrency-and-speed.md) | Many callers at once get what one caller gets, and the speed targets that must not slip |

## The library

| Page | What it holds |
|---|---|
| [library-reference.md](library-reference.md) | Recorded reference cases for the first ten library functions, each with its source |
| [library-versioning.md](library-versioning.md) | Cases for the versioning design page, on the loan whose payment is 564 in version 1 and 564.88 in version 2 |

The design pages these test are listed in [the design README](../2.%20design/README.md).
