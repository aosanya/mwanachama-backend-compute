# Parser fuzz run

How the parser is fuzzed, and how to start each run. The run is built (ISS-34) in `hostile_test/`, see [hostile-suite.md](hostile-suite.md). The engine is our own (decision 23), and the parser package is `formula/`.

Fuzzing uses Go's native fuzzing: `go test -fuzz`. No extra tool is installed.

## What is fuzzed

The target is `FuzzParse` in `hostile_test/`. It calls `Parse`, `Check`, `Plan` and `Run` with the input as a string.

Three properties are checked on every input:

| Property | Fails when |
|---|---|
| It never panics | any input makes the parser panic |
| It does bounded work | any input makes the parser work past what the formula length limit allows. The library has no time limit of its own (decision 19), so the bound is a count of work and not a clock |
| A refusal is a plain sentence with line and column | the parser refuses and the message has no sentence, no line or no column |

The sentence form is the parser-error sentence in [hostile-formulas.md](../4.%20qa/hostile-formulas.md).

## Seeds

The seed corpus lives in `hostile_test/testdata/fuzz/FuzzParse`. This is Go's own layout. Seeds are:

- the valid example formulas
- the near-parse cases (formulas one character away from valid)
- the nested-bracket and long-name cases from groups P and I of [hostile-formulas.md](../4.%20qa/hostile-formulas.md)

All seed values are invented.

## Time budget

| Run | When | Command flag | Time |
|---|---|---|---|
| Normal check | every `go test ./...` | none (no `-fuzz`) | seconds |
| Short run | before a merge | `-fuzztime=60s` | 60 seconds |
| Nightly run | once a night | `-fuzztime=30m` | 30 minutes |

These three numbers are decided here.

| Limit | Value |
|---|---|
| Work per input | bounded by the formula length limit, 500 characters proposed on the language page; no clock |
| Memory the run may use | 1 GiB |

Go's own fuzzing reports an input that hangs the run. That belongs to the test harness and is not a limit of the library.

Without `-fuzz`, `go test ./...` only replays the seeds and the saved failures. It does not search for new inputs.

## Commands

Run from the repo root, or use `make fuzz` and `make fuzz-nightly`.

Normal check:

```
go test ./...
```

Short run, before a merge:

```
GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=60s ./hostile_test
```

Nightly run:

```
GOMEMLIMIT=1GiB go test -fuzz=FuzzParse -fuzztime=30m ./hostile_test
```

`go test -fuzz` accepts one package and one target per run, so the package path is required.

## When a run fails

1. Go writes the failing input to `hostile_test/testdata/fuzz/FuzzParse/<hash>`.
2. Commit that file. It is a permanent case: every later `go test ./...` replays it.
3. Add the same input as a new row in the matching group of [hostile-formulas.md](../4.%20qa/hostile-formulas.md), with the next free id.
4. Fix it as a new version of the library. A published function is never edited (CLAUDE.md rule 4).

## Open questions

- Where the nightly run is scheduled. No CI exists for this repo yet and the repo has no remote.

## Related

- [hostile-formulas.md](../4.%20qa/hostile-formulas.md): the hostile formula groups and the parser-error sentence.
