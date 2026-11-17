# Implementation

The repo was created on 2026-11-17 as a pure Go library, the module `github.com/aosanya/mwanachama-backend-compute`. It holds the building blocks (numbers, dates, limits, the problem record, the value reader), the formula engine (front end, run, JSON forms), function definitions with Validate, the version registry and the first nineteen library functions, batch evaluation, and a test suite for each part of the design. The engine is our own (decision 23). The throwaway engine comparison is in `spike/`.

## What the repo holds

| File | What it is |
|---|---|
| `go.mod` | the module `github.com/aosanya/mwanachama-backend-compute` |
| `LICENSE` | Apache 2.0 |
| `Makefile` | build, test, vet, format and comment checks, the CI target, and the fuzz and nightly runs, see [ci.md](ci.md) |
| `compute.go` | the empty `compute` package |
| `compute_test.go` | a test that the package compiles |
| `number/` | exact decimal, float, rounding rules and splitting, see [numbers.md](numbers.md) |
| `datefn/` | the date value and the days, months and years between two dates, see [datefn.md](datefn.md) |
| `limit/` | the language's limits as settings and the cost meter, see [limit.md](limit.md) |
| `problem/` | the Validate problem record, its sentences and codes, see [problem.md](problem.md) |
| `numtext/` | text read as a value of the declared type, or refused, see [numtext.md](numtext.md) |
| `formula/` | the formula front end: read a formula's text into a tree, print it in the standard shape, and check it, see [formula.md](formula.md) |
| `formula/` (run) | the run of a checked formula: plan once, run many times, the cost limit and how a callee is metered, see [formula-run.md](formula-run.md) |
| `formula/` (JSON) | the JSON forms of a formula tree and of a function definition, see [json-forms.md](json-forms.md) |
| `function/` | function definitions with their input rules and checks, `Validate` and the refusal of a bad definition, see [function.md](function.md) |
| `functions/` | the first ten library functions with their recorded cases and the guard that stops a published version's cases from changing, see [library-functions.md](library-functions.md) and [library-reference-guard.md](library-reference-guard.md) |
| `functions/` (platform) | the nine functions after the first ten, and the check that every recorded case gives the same answer on every platform, see [platform-check.md](platform-check.md) |
| `library/` | the registry that serves library functions by name and version, known-wrong marks, the stop source and the loader of the nineteen functions, see [library.md](library.md) |
| `batch/` | one compiled function over many rows, a bad row never fails the batch, see [batch.md](batch.md) |
| `examplecases/` and `testdata/` | the loader that runs every documented example formula as a test case, with the 111 cases and the input definitions of the Validate rows, see [example-cases.md](example-cases.md) |
| `laws/` | the seven laws run on generated formulas and rows, with a nightly run, see [laws.md](laws.md) |
| `exactmoney_test/` | the exact money tests: shares, rounding and balances, see [money-suite.md](money-suite.md) |
| `hostile_test/` | the hostile and runaway formula tests and the parser fuzz run, see [hostile-suite.md](hostile-suite.md) and [fuzz-run.md](fuzz-run.md) |
| `basicruns_test/` | every part of the language run once on its own, with a check that no part lacks a case, see [basic-runs-suite.md](basic-runs-suite.md) |
| `inputchecks_test/` | the input check cases run against `function/`, problem by problem, see [input-checks-suite.md](input-checks-suite.md) |
| `bulk_test/` | the bulk run cases run against `batch/`, row by row, see [bulk-suite.md](bulk-suite.md) |
| `concurrency_test/` | many callers on one compiled program, and the speed benchmarks with their recorded baseline, see [speed-suite.md](speed-suite.md) and [concurrency-and-speed.md](concurrency-and-speed.md) |
| `cmd/` | `commentscan` (the comment scan), `refmanifest` (extends the reference manifest) and `refplatform` (the platform replay), see [ci.md](ci.md) and [platform-check.md](platform-check.md) |
| `internal/jsonform/` | the small helpers that the two JSON forms share, see [json-forms.md](json-forms.md) |
| `.github/` | the CI, nightly and platform replay workflows, written locally and not yet run on GitHub, see [ci.md](ci.md) |
| `spike/` | the throwaway engine comparison, a nested Go module kept so its figures can be re-run, see [engine-spike-results.md](engine-spike-results.md) |
| `CLAUDE.md` | the rules for working in this repo |
| `README.md` | the repo's front page |
| `documentation/` | the four phase folders: requirements, design, implementation, qa |

## Commands

| Command | What it does |
|---|---|
| `make build` | `go build ./...` |
| `make test` | `go test ./...` |
| `make test-race` | `go test -race -count=1 ./...` |
| `make vet` | `go vet ./...` |
| `make fmt-check` | fails if `gofmt` would change a file |
| `make comment-scan` | fails on a comment line in a Go file |
| `make ci` | vet, format check, comment scan and the race tests |
| `make fuzz` | the parser fuzz run for 60 seconds |
| `make nightly` | the laws, the hostile generator and the long fuzz run, with their seeds logged |
| `make clean` | removes `bin/` |

`make speed-check` runs the timed speed check of `concurrency_test/`; it and `make -C concurrency_test speed-targets` are not part of `make test` or `make ci` and are described in [speed-suite.md](speed-suite.md).

## Status

The repo is local. It is not yet pushed to GitHub and has no remote. There is no release tag yet; `v0.1.0` comes when the first working version passes its tests.

## Related

The design page for the agency's functions will be written in `mwanachama-backend-agency` under `documentation/2. design`. It is not written yet, so no link is given here. It must be linked here once it exists.
