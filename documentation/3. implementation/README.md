# Implementation

The repo was created on 2026-11-17 as a pure Go library, the module `github.com/aosanya/mwanachama-backend-compute`. So far it holds an empty `compute` package.

## What the repo holds

| File | What it is |
|---|---|
| `go.mod` | the module `github.com/aosanya/mwanachama-backend-compute` |
| `LICENSE` | Apache 2.0 |
| `Makefile` | build, test, vet and clean targets |
| `compute.go` | the empty `compute` package |
| `compute_test.go` | a test that the package compiles |
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
