# Implementation

The repo was created on 2026-11-17 as a pure Go library, the module `github.com/aosanya/mwanachama-backend-compute`. So far it holds an empty `compute` package and five packages that do not wait on the engine choice: numbers, dates, limits, the problem record and the value reader. The throwaway engine comparison is in `spike/`.

## What the repo holds

| File | What it is |
|---|---|
| `go.mod` | the module `github.com/aosanya/mwanachama-backend-compute` |
| `LICENSE` | Apache 2.0 |
| `Makefile` | build, test, vet and clean targets |
| `compute.go` | the empty `compute` package |
| `compute_test.go` | a test that the package compiles |
| `number/` | exact decimal, float, rounding rules and splitting, see [numbers.md](numbers.md) |
| `datefn/` | the date value and the days, months and years between two dates, see [datefn.md](datefn.md) |
| `limit/` | the language's limits as settings and the cost meter, see [limit.md](limit.md) |
| `problem/` | the Validate problem record, its sentences and codes, see [problem.md](problem.md) |
| `numtext/` | text read as a value of the declared type, or refused, see [numtext.md](numtext.md) |
| `CLAUDE.md` | the rules for working in this repo |
| `README.md` | the repo's front page |
| `documentation/` | the four phase folders: requirements, design, implementation, qa |

## Commands

| Command | What it does |
|---|---|
| `make build` | `go build ./...` |
| `make test` | `go test ./...` |
| `make vet` | `go vet ./...` |
| `make clean` | removes `bin/` |

## Status

The repo is local. It is not yet pushed to GitHub and has no remote. There is no release tag yet; `v0.1.0` comes when the first working version passes its tests.

## Related

The design page for the agency's functions will be written in `mwanachama-backend-agency` under `documentation/2. design`. It is not written yet, so no link is given here. It must be linked here once it exists.
