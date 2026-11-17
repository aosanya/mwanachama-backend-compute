# The first ten library functions

Built for ISS-27, in the package `functions/`. It holds the ten functions the agency templates need first, each as a definition with a one-line body, and the recorded cases that protect them, with the guard that stops a published version's cases from changing. The cases are on [library-reference.md](../4.%20qa/library-reference.md), the rule and the test design on [library-reference-guard.md](library-reference-guard.md), the rule that a published version is never edited in decision 10 of [decisions.md](../1.%20requirements/decisions.md) and in [library-versioning.md](../2.%20design/library-versioning.md). All names and figures are invented.

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

Since ISS-35 the package also holds the nine functions after these ten (nineteen in all, 152 recorded cases), a runtime that lets a function call another, and a check that every case gives the same answer on every platform. They are on [platform-check.md](platform-check.md). The counts below (ten functions, 80 cases) describe the first ten.

## Where it lives

| Path | What it holds |
|---|---|
| `functions/library.go` | `Function`, `All` and `Find`: a function is a library, a name, a version, a `function.Definition`, a one-line body and its result type |
| `functions/definitions.go` | the ten definitions |
| `functions/rules.go` | the small constructors of input rules the definitions use |
| `functions/call.go` | `Prepare` and `Callable.Call`: validate first, read the values, run the body |
| `functions/judge.go` | how a result is compared with a recorded case |
| `functions/reference.go` | the case shape, the canonical form, the hash and the loader |
| `functions/guard.go` | the manifest, the guard and `ExtendManifest` |
| `functions/reference/<library>.<function>@<n>.json` | the cases of one function version |
| `functions/reference/manifest.json` | the state of every version and a hash of every case |
| `cmd/refmanifest/` | `go run ./cmd/refmanifest` adds the cases the manifest does not hold |

Proposed, and a difference from the issue text and the design page. The issue named `library/` for the functions and `testdata/` for the cases. `library/` is the library versions of ISS-28 and `testdata/` is the example cases of ISS-31, so both live in `functions/` here and the data files sit in `functions/reference/`. The design page's `testdata/reference/` is one `git mv` away if the owner prefers it.

Imports: the standard library, `formula/`, `function/`, `number/`, `numtext/` and `problem/`. No comments in the code.

## The ten functions

| Function | Result | Body |
|---|---|---|
| `fin.balance_after_payment@1` | decimal | `payment >= balance ? 0.00 : balance - payment` |
| `fin.equal_share@1` | decimal | `round((sum(deposits) - sum(expenses)) / members, 2)` |
| `date.late_days@1` | integer | `max([0, days_between(due, filed)])` |
| `fin.flat_instalment@1` | decimal | `round((principal + principal * monthly_rate * months) / months, 2)` |
| `fin.prorate_fee@1` | decimal | `round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)` |
| `fin.straight_line_value@1` | decimal | `max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])` |
| `fin.coverage_ratio@1` | float | `round(to_float(stock_value) / to_float(facility_balance), 3)` |
| `fin.proportional_share@1` | decimal | `round(call_amount * units / sum(all_units), 2)` |
| `stats.spread@1` | decimal | `max(values) - min(values)` |
| `math.remainder@1` | integer | `value % divisor` |

Each starts at version 1. The inputs, their types and their order are the signatures of the reference page, and a test reads the page and fails when a definition differs from it. The rules are the ones the page states in words: money is `required`, `no-negatives` and at most 2 places; principal, cost, facility balance and call amount are more than 0.00 instead of `no-negatives`; a rate is 0.000 to 0.050 with at most 3 places; months is 1 to 60, members 1 to 500, `life_years` 1 to 50, `days_in_period` 28 to 31, `divisor` 1 to 100; deposits, `values` and `all_units` need at least one item and `expenses` may be empty.

## How a function is called

```go
found, _ := functions.Find("fin.balance_after_payment@1")
callable, err := found.Prepare()
outcome := callable.Call(function.Row{"balance": numtext.One("120.50"), "payment": numtext.One("45.25")})
```

`Prepare` refuses a function whose definition `function.CheckDefinition` refuses, whose body `formula.Parse` or `formula.Plan` refuses, or whose body gives a type other than the one the function says. It plans the body once. `Call` runs `Validate` first and never calculates a row that has a problem (decision 8), then reads the received texts into typed values and runs the program. An `Outcome` holds the value, or the problems of `Validate`, or the one problem of the run. `Refused` says whether there is no result and `Sentence` gives the first sentence. A `Callable` holds no state and may be called from many goroutines.

## How a recorded case is checked

One case in a data file is the shape of the design page: `id`, `kind`, `inputs`, `expected`, `tolerance` and `source`. Inputs are texts, a list is an array of texts, and an input that is not given is left out. The working stays on the reference page only.

| Result | Passes when |
|---|---|
| Decimal and integer | the text of the result equals the expected text, so `0.00` is not `0` |
| Float | `abs(result - expected) <= 1e-9 x abs(expected)`, and exactly `0.0f` when `0.0f` is expected. The trailing `f` of the expected text is dropped |
| `refused` | there is no result: `Validate` gave a problem, or the run did |

A failed case prints the design page's sentence, for example `REF-fin.pmt-01: fin.pmt@1 gave 565, the case records 564.` A test also pins the rule each refusal is refused for, so a case that refuses for another reason fails: 21 are refused by a rule of `Validate` and one by the run (below).

## The guard

`Guard.Check` gives failures and notices. It is run by `go test ./functions/` on the real files and by tests that change the files in memory and on disk.

| # on the design page | What the guard does |
|---|---|
| 1 | A case of a published version whose hash differs from the manifest fails: `REF-...: this case of published fin.x@1 has changed (recorded 345008ae79f4, now ...). A published version is never edited. Ship the fix as the next version of fin.x.` |
| 2 | A recorded case missing from the data file fails |
| 3 | A published version with no data file, or that `All` no longer lists, fails |
| 4 | A case with no source id, or whose source id has no row on [sources.md](../1.%20requirements/sources.md), fails |
| 5 | The cases of a draft version are not checked against the manifest |
| 6 | The replay: every case of every version is run (`TestEveryRecordedCaseReplaysAgainstTheLibrary`) |
| 7 | The ids on the reference page and in the data files must be the same for each version |

A case added to a published version is a notice and not a failure. `ExtendManifest` and `go run ./cmd/refmanifest` add the cases the manifest lacks, give a new version the state `draft`, and stop with failure 1 if an id is already recorded with another hash. They never change or delete an entry.

The canonical form is the design page's, and a test reproduces its worked hash for the first case, `345008ae79f497d9c4999c1075c9ba95b63f68866d2c3ab2bfa6ad9dd893daab`, with the Go code and not with the outside script. The manifest holds 80 hashes, all computed by the command and none typed by hand.

All ten versions are `draft` in the manifest. What makes a version `published` is the owner's open question 1 of the guard page (proposed: the commit of the release tag). The guard tests mark every version published in memory and on a copy of the files, so the failures are proved. Flipping the state in `manifest.json` turns failures 1 to 3 on for the real files.

## How to add function number eleven

1. Write its cases in the reference page, with a source id from sources.md, and its section with the `Signature:` line.
2. Write the definition in `definitions.go`, add it to `All`, and write its data file `functions/reference/<library>.<function>@1.json` with the same ids.
3. Run `go run ./cmd/refmanifest` and then `go test ./functions/`.

A fix to a published function is never an edit: add `@2` as a new `Function` beside `@1`, give its cases the next numbers of the same function (`REF-fin.x-09`), and leave every case of `@1` as it is.

## What stays open

These are the owner's, copied from the reference page and the guard page. Nothing here is decided.

1. **A division by zero in `fin.proportional_share@1`.** Case 08 (`units 0, all_units [0]`) is refused by the run with `divide-by-zero`, not by an input rule, because the formulas page has no rule that the sum is more than 0 (reference page question 7). If the owner adds the rule it is a check of the definition and a new version.
2. **The rounding of halves.** Five of these ten functions' cases (eight in all nineteen, see [platform-check.md](platform-check.md)) sit on a half and pass with half away from zero, the setting `number.DefaultSettings` uses (reference page question 1, exact-money question 1). With another default they fail and the cases would be a new version.
3. **The wording of a refusal** (reference page question 2). A refused case passes on any sentence, and the test pins only the code.
4. **The float tolerance against zero** (question 3): the guard asks for exactly `0.0f`.
5. **Cases not written** for the open questions 4, 5 and 6 of the reference page: a ratio on a half, a negative share and dates in the wrong order.
6. **Names** (question 8): `date.` for `late_days`, and the general `stats.spread` and `math.remainder`.
7. **What makes a version published**, who may mark it, and where the manifest lives (guard page questions 1, 2 and 4).
8. **Whether the guard compares with the manifest of the last release tag** (guard page question 5). It does not: a commit that edits a published case and its hash together passes, and only review stops it.
9. **Where the functions are registered.** They are registered in `library/` (ISS-28, ISS-77): `library.LoadFunctions` publishes `All()`, all nineteen functions, into the registry, and a formula that calls `fin.balance_after_payment@1(...)` runs on them. See [library.md](library.md). Every version is still served as released version 1 while the manifest says `draft`, because what makes a version published is open (question 7 above).
10. **The guard's question 6**: how the test asks the library which versions it serves is answered by `All()` here.

## Tests

`go vet ./functions/ ./cmd/...` is clean and `go test ./functions/ -race` passes: 80 recorded cases replayed (58 results and 22 refusals), the refusal reasons, that every definition is accepted on save, that every key is unique and at version 1, that two calls give the same answer, the page's signatures, and the guard tests above.
