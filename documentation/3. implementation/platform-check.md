# The nine other functions and the platform check

Built for ISS-35, in the package `functions/`. It adds the nine functions that follow the first ten, with their recorded cases, and a check that every recorded case gives the same answer on every machine. The first ten and the guard are on [library-functions.md](library-functions.md). The cases are on [library-reference.md](../4.%20qa/library-reference.md), the sources on [sources.md](../1.%20requirements/sources.md), the rule that a published version is never edited in decision 10 of [decisions.md](../1.%20requirements/decisions.md). All names and figures are invented.

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## Where it lives

| Path | What it holds |
|---|---|
| `functions/definitions_more.go` | the nine definitions |
| `functions/rules_more.go` | the small constructors the nine need: allowed values, a yes-or-no input, a float range, a list of whole numbers with a range |
| `functions/calls.go` | `Calls`, the runtime that lets a function call another by name and version |
| `functions/call.go` | `Prepare` now calls `PrepareIn(NewCalls())`, so a function whose body calls another is planned with a runtime |
| `functions/platform.go` | `Replay`, `Platform`, `ComparePlatforms` and the file of recorded results |
| `functions/reference/<library>.<function>@1.json` | 72 new cases, eight for each of the nine |
| `functions/reference/manifest.json` | 72 new hashes, every new version `draft` |
| `functions/reference/platform-results.txt` | the result of every case on the platform it was recorded on (152 lines and a header) |
| `cmd/refplatform/` | `go run ./cmd/refplatform` compares this platform with the recorded one, `-record` writes it |
| `.github/workflows/replay.yml` | the smallest CI file: the replay and the platform check on Ubuntu and macOS |

Imports: the standard library, `formula/`, `function/`, `number/`, `numtext/` and `problem/`. No comments in the code.

## The nine functions

Each starts at version 1. The formula number is the one of [example-formulas.md](../1.%20requirements/example-formulas.md), and the signature is the one on the reference page; a test reads the page and fails when a definition differs from it.

| Function | Formula | Result | Body |
|---|---|---|---|
| `fin.loan_within_limit@1` | 2 | boolean | `status == "active" && requested <= 3 * sum(contributions)` |
| `fin.farmer_net_payout@1` | 4 | decimal | `max([0.00, round(kilos * price_per_kilo, 2) - credit])` |
| `fin.late_filing_penalty@1` | 6 | decimal | `date.late_days@1(due, filed) == 0 ? 0.00 : date.late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : date.late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)` |
| `fin.interest_step@1` | 7 | decimal | `balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020` |
| `fin.monthly_interest@1` | 8 | decimal | `round(balance * fin.interest_step@1(balance), 2)` |
| `edu.term_grade@1` | 10 | text | `avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"` |
| `fin.interim_payment@1` | 14 | decimal | `max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])` |
| `fin.lending_data_ok@1` | 16 | boolean | `count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00` |
| `member.voting_right@1` | 18 | boolean | `!suspended && role != "guest" && (role == "officer" \|\| (paid_up && months_between(joined, on) >= 6))` |

The library prefixes `edu.` and `member.`, the prefix `fin.` for `interest_step`, and the names themselves are proposals (reference page, open question 8). The rules of each input are the ones the formulas page states: a status or a role is one of its allowed values, a float share is 0.0f to 1.0f, scores are whole numbers 0 to 100, a price per kilo is more than 0.000 with at most 3 places, and the balances of formula 16 have no no negatives rule.

The input texts of a case are read by `numtext`: a float input is the digits (`0.37`), not `0.37f`, and a yes or no is `true` or `false`. The expected text of a text result is the text (`B`), of a float result the digits and an `f` (`1.363f`), as on the first ten.

## One function calling another

`fin.late_filing_penalty@1` calls `date.late_days@1` and `fin.monthly_interest@1` calls `fin.interest_step@1`. `formula.Plan` needs a runtime that knows the callee, so `Function.PrepareIn(calls)` takes a `*Calls` and `Prepare()` makes a fresh one.

| What `Calls` does | How |
|---|---|
| Serves a signature | `Function(library, name, version)` prepares the callee from `All()` once, keeps it, and gives its parameters, result, cost and call depth (1 plus the deepest callee) |
| Serves the versions | `Versions(library, name)` lists the versions `All()` holds |
| Runs a callee | `Apply` and `ApplyMetered` read the received values as text, run `Validate` first and never calculate a row that has a problem (decision 8), and hand the caller's meter to the callee so its work counts against the caller's cost limit |
| Refuses | a callee that is not in `All()` gives `unknown-function` with its name and version; a callee that is being prepared when it is asked for again (a loop of calls) is not found |

The version registry of `library/` (ISS-77) loads the same nineteen functions with `LoadFunctions`, and a formula that runs through it reaches the callee through the registry's own `ApplyMetered`. `Calls` is for the replay in this package, which cannot import `library/` because `library/` imports `functions/`. A formula `fin.monthly_interest@1(balance)` planned on the registry gives 48.26 for 3217.45; this was checked once with a temporary test and is not kept as a test of this package.

## The cases

72 cases for the nine, eight each, in the shape of the first ten: id, kind, inputs, expected, tolerance and source. The working stays on the reference page. Sources are the three rows of `sources.md`:

| Source | Cases | Rests on |
|---|---|---|
| S-01 | 14 | the worked example of the formulas page |
| S-02 | 16 | the input tables and invalid examples of the formulas page |
| S-03 | 42 | the hand calculation in the Working column |

Every figure was worked by hand, and a second throwaway script outside the repo, which re-implemented the nine functions with Python's decimal module and the calendar, recomputed all 72 expected results and every refusal. No expected figure was taken from the library under test. The script also found the three cases that change under half-even or half-down (`fin.farmer_net_payout` 05, `fin.monthly_interest` 06 and `fin.interim_payment` 06) and checked that no other case does; they are in the table under Tolerances of the reference page.

Every refusal is pinned to the rule it names (`refusalCodes` in `refusals_test.go`): `not-allowed`, `too-small`, `too-large`, `no-negatives`, `decimal-places`, `list-too-short`, `wrong-type` and `required`. No new refusal comes from a run.

## The platform check

A library function must give the same answer on every machine. The check replays the 152 recorded cases on the platform it runs on, writes down what each gave, and compares that with the results recorded on a reference platform.

| Piece | What it does |
|---|---|
| `Replay(versions)` | runs every case through `Prepare` and `Call`, and keeps for each id the function key and a text: the result as the engine prints it, or `refused:<code>` for a refusal, with the code of the first problem or `run:<code>` |
| `platform-results.txt` | line 1 is `platform<TAB>darwin/arm64`; then one line `id<TAB>function<TAB>result` for each case, sorted. It was written by `go run ./cmd/refplatform -record` on `darwin/arm64` |
| `ComparePlatforms(recorded, got, cases)` | a decimal, an integer, a boolean, a text and a refusal must be the same text. A float result agrees when `abs(got - recorded) <= 1e-9 x abs(recorded)` and, when the recorded one is exactly 0, only when it is exactly 0 (the reference page's tolerance and its open question 3) |
| The sentence | `REF-fin.balance_after_payment-01: fin.balance_after_payment@1 gave 76.25 on linux/amd64 and 75.25 on darwin/arm64. A library function gives the same answer on every machine.` A case that is missing on one side gives `... recorded on <platform> and not run on <platform>.` or `... has no result recorded on <platform>.` |
| `go test ./functions/` | `TestEveryCaseGivesTheRecordedResultOnThisPlatform` runs on whatever platform runs the tests, so a difference fails the run there |
| `go run ./cmd/refplatform` | the same comparison as a command, for CI: it prints each failure and exits 1 |
| `.github/workflows/replay.yml` | on every push and pull request, on Ubuntu and macOS: `go vet`, `go test -race` and the command |

The sentence is a proposal; the page named none. A refusal is compared by its code and not by its sentence, because the wording of a refusal is open (reference page, open question 2).

### Where it was run

| Platform | How | Result |
|---|---|---|
| darwin/arm64 | `go test ./functions/ -race` and `go run ./cmd/refplatform` | passes; the reference platform |
| darwin/amd64 | the same without `-race`, as an Intel binary under Rosetta on the same Mac | 152 results agree |
| linux/arm64 | a static binary built with `GOOS=linux CGO_ENABLED=0`, run in an Alpine container (`postgres:16-alpine`, aarch64) against the repo mounted read only: the command and the package's test binary | 152 results agree; the tests pass |
| linux/amd64 | built and vetted with `GOOS=linux GOARCH=amd64`, not run | not run |
| the workflow on Ubuntu and macOS | not run: the repo has no remote, so no workflow has run | not run |

So "all recorded cases pass on both platforms" is shown for macOS and for Linux on ARM; it is not shown for Linux on Intel or in the CI the workflow describes. Both are written to run when the repo is pushed.

## Checks in this package

`go vet ./functions/ ./cmd/...` is clean and `go test ./functions/ -race` passes: the replay of all 152 cases with the judge of the first ten, the signatures read from the reference page (19), the refusal codes (the 22 of the first ten and the 16 of the nine), the guard tests against the real files with 152 hashes, the calls runtime (a callee's signature and call depth, a version that does not exist, a callee that refuses what its rules refuse, the answer of a callee on its own and through a call) and the platform tests (the comparison, its sentence, a float inside and outside the tolerance, an exact zero, a refusal for another rule, a missing result, a malformed file).

The new tests that were written before the code: the cases and the 19-function counts (commit `148d63f`) failed only because the nine functions did not exist; the platform tests failed first because `platform-results.txt` did not exist, and passed once it was recorded. The calls tests were written after the runtime.

## What stays open

These are the owner's, or another issue's. Nothing here is decided.

1. **The count of `library/`.** `TestEveryRecordedCaseGivesTheRecordedAnswerThroughTheRegistry` in `library/load_functions_test.go` says `10 functions are recorded` and fails now that `functions/reference` holds 19. The line is the only failure of `go test ./library/` and needs `10` changed to `19` by the owner of that package (ISS-77); it was not edited from here. Whether the registry's replay then passes for the nine is shown by `LoadFunctions` loading all nineteen without error and by the one call above, not yet by that test.
2. **All nineteen versions are `draft`.** What makes a version `published` is still the first open question of [library-reference-guard.md](library-reference-guard.md). Once it is decided the guard protects the 152 cases.
3. **The platform results file can be rewritten.** `go run ./cmd/refplatform -record` replaces the recorded results of a platform. Nothing stops a change to that file from hiding a real difference. Proposed: the file is part of the published release and a change to it is reviewed like a change to a published case. This is not decided.
4. **Linux on Intel and the CI.** See the table above.
5. **A reference platform.** The recorded results come from one machine. Whether the recorded platform should be the CI's Linux, and who records it again after a change of the Go toolchain, is open.
6. **The questions of the reference page that are new**: negative balances in `fin.lending_data_ok@1` (10), a float made a decimal on a half (11) and calling another function (12). Open questions 1 to 9 of that page are unchanged, and the count of cases on a half is now eight.
7. **A binary float on the platform check.** The float cases are few (`fin.coverage_ratio@1`), and they agree to the last digit on the three platforms run, so the tolerance has not been needed yet. It is the page's, not tuned.
8. **The README index.** The row for `platform-check.md` is not added, because another agent has `documentation/3. implementation/README.md` open:
   `| platform-check.md | the nine functions after the first ten, a function calling another, and the platform check of the recorded cases (ISS-35) |`
