# CI

Built for ISS-80. The repo has no remote and nothing has been pushed, so none of these workflows has run on GitHub yet. Each check was run locally. Status: the owner has not reviewed this page.

## What is in the repo

| File | What it is |
|---|---|
| `.github/workflows/ci.yml` | the checks on every push to `master` and on every pull request |
| `.github/workflows/nightly.yml` | the long runs, once a night at 02:17 UTC and on request |
| `.github/workflows/replay.yml` | the platform replay of the recorded library cases on Ubuntu and macOS, see [platform-check.md](platform-check.md) (ISS-35) |
| `cmd/commentscan/` | the comment scan: `scan.go`, `main.go` and 13 tests |
| `Makefile` | the targets the workflows call |

## The checks of ci.yml

The job runs on `ubuntu-latest` (amd64) and `ubuntu-24.04-arm` (arm64) as two runs side by side, and a failure on one does not stop the other. The Go version is the one in `go.mod`. Each step is a make target, so a person runs the same check at a desk.

| Step | Command | Why it is there |
|---|---|---|
| Vet | `make vet` | `go vet ./...` finds code that compiles and is wrong |
| Format | `make fmt-check` | fails and lists the files when `gofmt -l .` names any |
| Comment scan | `make comment-scan` | the repo's rule 5: no comments in code. See below |
| Tests | `make test-race` | `go test -race -count=1 ./...`: every package, with the race detector, with no cached result |

`make ci` runs the four in this order for a local run.

The two architectures matter because the library gives exact answers that must not depend on the machine. The laws page records an output digest that was the same on arm64 and amd64 on one Mac; a Linux run is the missing part and the job gives it.

## The comment scan

`go run ./cmd/commentscan .` walks every `.go` file under the folder, parses it with `go/parser` and prints `file:line: text` for each comment. It exits 1 when it printed any and 2 when a file does not parse.

It keeps what the owner's rule keeps:

| Kept | How it is recognised |
|---|---|
| Build tags and `//go:embed`, `//go:generate`, any `//go:` directive | the comment starts with `//go:` |
| `//line`, `//nolint`, `//lint:`, `//export` and `// +build` | the comment starts with that word |
| A license header | a comment group before the `package` clause that contains `Copyright`, `SPDX-License-Identifier` or `Licensed under` |
| Generated files as a whole | `ast.IsGenerated`: a `// Code generated ... DO NOT EDIT.` line before the `package` clause |

It skips folders called `testdata` and folders whose name starts with a dot. A comment marker inside a string or a rune is not a comment, because the parser decides what a comment is and not a text search. A directive that sits in the same group as a prose line does not hide the prose line.

Proof: on the repo the scan prints nothing and exits 0. On a copy of `number/` with one comment line added, it printed `decimal.go:166: // planted` and exited 1.

## The nightly run

`make nightly` runs three things one after another:

| Step | What it logs |
|---|---|
| `make -C laws nightly` | 20,000 cases for each law, and the seed range of every law |
| the generated hostile text test with `HOSTILE_SEED` set to the time | the seed, so a failure replays with `HOSTILE_SEED=<n> go test ./hostile_test/ -run Generated -v` |
| `make fuzz-nightly` | `FuzzParse` for 30 minutes (`NIGHTLY_FUZZTIME`, default `30m`) |

Go's fuzzer has no seed to log. When it finds a failing input it writes it as a file under `hostile_test/testdata/fuzz/`, and the workflow uploads that folder as the artifact `failing-fuzz-inputs` when the run fails. The input is then committed as a file and added as a row to the hostile suite, as [hostile-suite.md](hostile-suite.md) describes.

A local run with `make nightly NIGHTLY_FUZZTIME=20s` passed: 20,000 cases for each of the laws with seeds 1794923028 to 1794943027 (about 2 minutes 20 seconds), the hostile generator at seed 1794923169, and 125,930 fuzz executions with no failure.

## What is not in CI

| Not in CI | Why |
|---|---|
| The timed speed checks of [speed-suite.md](speed-suite.md) (`make -C concurrency_test speed-check` and `speed-targets`) | a shared runner is noisy and gives a different machine each night. The speed suite judges a figure only on the machine that recorded it and only when the machine is quiet. In CI the cases run once to show each one works and no time is judged, because the timed tests need `SPEED_CHECK=1` |
| `spike/` | it is a separate Go module (its own `go.mod`) holding the throwaway engine comparison, so `./...` does not reach it. The comment scan does walk it, and it passes |
| Release tags | `v0.1.0` waits for the first working version and for the owner |

## Verified locally

| Check | Result |
|---|---|
| Both workflow files parse as YAML | yes |
| `make fmt-check`, `make comment-scan` | pass on the repo |
| `make vet` and `make test-race` on a copy of the committed tree | vet, format and comment scan pass; the tests passed in every package except `examplecases` and `library`, whose tests were being changed by ISS-79 and ISS-35 at that moment |
| `make nightly NIGHTLY_FUZZTIME=20s` | passes |

The workflows themselves, the arm runner label and the artifact upload could not be run without GitHub.

## Open

- The repo has no remote. The workflows start to run when the owner confirms the public repo and it is pushed (ISS-1 todo 3). The branch is `master`.
- `ubuntu-24.04-arm` is the label GitHub uses for its arm runners on public repos. If the label changes, `ci.yml` changes in one line.
- Whether the nightly run should open an issue on failure is not decided; today it fails the workflow and keeps the failing fuzz input.
- Whether `replay.yml` and `ci.yml` should be one file: they are separate because they were built by different issues.
