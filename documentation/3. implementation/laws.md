# Laws as generated-input tests

The seven laws of [4. qa/laws.md](../4.%20qa/laws.md) run on generated trees and rows in the `laws/` folder. Every case has a seed, a failure prints the seed, a failing tree is shrunk, and the nightly run logs the seed it starts from. Built for ISS-32; laws 4 and 6, law 5 over pinned library versions and the complete problem list of law 7 were switched on in ISS-76.

Every count, limit and name below that the QA page marks proposed is still proposed. All figures in the generators are invented.

## What runs

| Law | What the test checks | State |
|---|---|---|
| 1 Round trip | `Parse(Format(tree))` equals `tree` on 500 trees | runs in `make test` |
| 2 Standard shape is stable | `Format(Parse(Format(x)))` equals `Format(x)` | runs in `make test` |
| 3 Spacing does not change the hash | 2 spellings of one tree give one hash | runs in `make test` |
| 4 Batch equals single | each tree becomes a function whose every input is required; the rows of the row generator are written as received text and run alone and in one batch (a random chunk size of 1 to 50 and 1 to 4 workers), and the two outcomes must match row by row, in the same order and the same count | runs in `make test` |
| 5 Same inputs, same output | each tree is run over 20 generated rows twice in one process, once in a second process, and a digest of all outputs can be compared across machines; for every published version of `fin.pmt`, a pinned call is run twice, on two registries built apart, and once in a second process | runs in `make test`; the cross-machine check is opt-in, see below |
| 6 A published function never changes | the results recorded in `laws/testdata/references.txt` are replayed against the library and must come out as recorded; every published version must have a record | runs in `make test` |
| 7 Validation first | a row with a planted missing or wrong-type input is never calculated; Validate lists every planted problem and no other, and the batch run stops that row at validation | runs in `make test` |

No law is waiting on a build any more. `go test ./laws -v` still lists as skipped the checks that are opt-in or print-only: the cross-machine digest, the two digest printers and the reference recorder.

## The generators

All in `laws/`, none read the clock, none keep state between cases.

| File | Holds |
|---|---|
| `gen.go` | `NewCase(seed)`: 8 inputs drawn from a pool of 11 (integers, decimals, a float, text, a boolean, two dates, a list of integers, a list of decimals) and a tree over them |
| `spacing.go` | `SpellDifferently(text, seed)`: random spaces, tabs and line breaks between the tokens of standard text, never inside a number, a name, a call name or a text, and never past the formula length limit |
| `rows.go` | `NewRows(seed, inputs)`: 1 to 200 rows; about 4 rows in 10 carry 1 to 5 planted problems, each a missing input or a value of the wrong type, recorded on the row |
| `shrink.go` | `Shrink`: replaces a node by one of its children or by the simplest literal and keeps the replacement when the tree is smaller, still passes Check and still fails; repeats until nothing smaller fails. `Show` prints a tree with every operator bracketed, so `a - (b - c)` and `(a - b) - c` can be told apart in a failure |

How a tree is generated: it is built by type, so a decimal slot gets a decimal expression, a boolean slot a comparison or a logic operator, a list slot a name or a list of items, and the conditional's two answers always have the same kind. Limits are the language page's proposed values. A tree is kept only when `Check` passes it (depth at most 12, at most 100 nodes, cost at most 20000, text at most 500 characters, at most 8 inputs); the generator tries again with the same random source when it does not, so the same seed always ends on the same tree. Library calls are not generated yet, because no library function exists; they come with ISS-27.

The same seed gives the same case twice. A test checks it on 200 seeds, and checks on 1000 seeds that every tree passes Check, stays inside the limits, and that every kind of node, every built-in function except the library calls and every literal type turns up.

## How a failure is reported

Each law is a function `func(Engine, seed) *Failure`. The message is the page's sentence:

| Law | Message |
|---|---|
| 1 | `round trip failed, seed N, shrunk to T: Parse(Format(tree)) read back as U` |
| 2 | `standard shape moved, seed N, shrunk to T: first Format gave X, second gave Y` |
| 3 | `hash differs, seed N, shrunk to A and B: 2 spellings, 2 hashes` |
| 4 | `batch differs from single, seed N, shrunk to T, row R: alone gave X, in the batch gave Y` (`row count` takes the place of `row R` when the batch gives a different number of rows) |
| 5 | `output moved between runs, seed N, shrunk to T: run 1 gave X, run 2 gave Y` |
| 5, pinned | `output moved between runs, seed N, pinned fin.pmt@V, row R: run 1 gave X, run 2 gave Y` |
| 6 | `reference changed, fin.pmt@V, input I, recorded X, now Y; ship the fix as a new version`, or `no reference recorded for fin.pmt@V, recorded up to fin.pmt@W; record the results it gives when it is published` |
| 7 | `row R has 1 problem and was calculated, seed N, shrunk to T` |
| 7, list | `problem list incomplete, seed N, row R: planted P, listed L, missing: I not given` (or `I has the wrong type`), `problem list wrong, seed N, row R: planted P, listed L, not planted: I`, and `row R has N problems and was calculated, seed N` |

`Engine` holds the things the laws call: Parse, Format, Check, Hash and Run for laws 1 to 3, 5 and 7; `Batch` (a `Single` and a `Batch` function over a compiled function of `batch/`), `Validate` (`function/`) and `Library` (a function that gives a fresh registry of `library/`) for laws 4, 5 pinned, 6 and 7 list. The real engine is `RealEngine()` and a broken one is a copy with one field replaced. Law 7 adds the seed at the end of its sentence, which the page does not have.

## Each law fails on a broken engine

`laws/proofs_test.go` holds one broken engine per law and checks that the law finds it within 500 cases, with the page's sentence, a seed, and a seed that reproduces the same message while the real engine passes that seed.

| Law | Broken engine |
|---|---|
| 1 | Format drops every bracket |
| 2 | Format trims a trailing zero of a decimal each time it runs |
| 3 | The hash is taken of the typed text and not of the standard text |
| 5 | Run adds something that changes from one call to the next |
| 7 | Run turns the missing-input problem into a value |
| 4 | the batch gives a later row the problem of an earlier row; the batch gives its rows in reverse; the batch drops its last row |
| 5, pinned | `fin.pmt@2` gives a different figure on every second call |
| 6 | `fin.pmt@1` is edited in place to round to cents; a third version is published with no record |
| 7, list | Validate stops at the first problem; Validate names an input nobody planted; the batch calculates a row that failed validation |

The page asks for the proof in a test copy of the engine. The proof here replaces one function of the engine in the test instead, so nothing is copied and the committed code is never edited.

## Law 4: how a row reaches the batch

`DefinitionOf(inputs)` declares every input of a case as required, with its type and no other rule, so what the batch refuses is what the row generator plants. `ReceivedRow` writes each generated value as the text a caller would send: integers, decimals, floats, text, yes or no and dates as one text, the two list types as a list of texts. A planted wrong-type value for an input of type text is an integer that reads as text, so it is not a problem, and the list of planted problems of a text input is left out of the expected list for law 7.

The same compiled function is used for the single and the batch run, so the law compares `Compiled.Run` row by row with `Compiled.RunBatch`, which is the shared surface; the ISS-29 builder's own tests cover the cost limit and cancel. A row's outcome is its value with its type, or its stage and its problems with code, input and sentence, so a changed sentence counts as a difference.

## Law 6: references

`laws/testdata/references.txt` holds one line for each recorded result: `fin.pmt@1 12000.00 0.01 24 => 564`, the reference of a published version, its inputs in the order of its signature and the result it gave when it was published. There are five lines for each of the two published versions. A line is only added for a new version, never changed. To record a new version, publish it in `FixtureReleases` and run `LAWS_RECORD=1 go test ./laws -run TestRecordReferences -v`, which prints five lines for every published version; the lines of the new version are copied into the file.

The law fails when a recorded result moves, and when a published version has no record at all. It cannot see a whole function that has no record for any version, because the registry has no listing of functions; that is open below.

`laws/library_law.go` holds `FixtureLibrary`, an invented `fin.pmt` in two versions (rounded down to a whole number, and rounded to the cent) as the laws' own stand-in, so that nothing here waits for the real functions.

## Law 5 over pinned versions

The version generator picks one of the published versions of `fin.pmt` from the seed, builds `fin.pmt@V(principal, rate, months)` and runs it over up to 20 generated rows twice, each time on a registry built afresh, then once in a second process (`TestPrintPinnedDigest` is the child). The registry is a field of `Engine`, so a broken registry replaces it in a test.

## Cases and the nightly run

| Setting | Meaning |
|---|---|
| `LAWS_CASES` | cases per law; 500 when not set (proposed) |
| `LAWS_SEED` | the number the seeds count up from; 0 when not set, so a normal run uses seeds 1 to 500 |

Every law logs `N cases, seeds A to B` first, so a verbose run names its seeds. `make -C laws nightly` runs 20,000 cases per law from the current time as the base seed, and the seed range is in its output for every law, because the target's filter matches every test file of the folder. One nightly run now takes about 2 minutes 30 seconds; law 4 is 70 seconds of it. A failing seed is replayed with `LAWS_SEED=<seed - 1> LAWS_CASES=1`.

The root `Makefile` is shared with other work in progress, so the `nightly` target lives in `laws/Makefile`; a line `nightly: ; $(MAKE) -C laws nightly` in the root `Makefile` is left to add when that file is free.

## Law 5 across processes and machines

The second process is the test binary run again with `LAWS_CHILD=1`; it prints the digest of every outcome over the same seeds and the first process compares it with its own.

Across machines the test compares the digest of seeds 1 to 500 with `laws/testdata/output-digest.txt`. It is off unless `LAWS_PLATFORM_GOLDEN=1`, because a legitimate change of the engine moves the digest and would then fail every test run until the file is regenerated. The CI turns it on for both platforms it offers (Linux and macOS, proposed). Regenerate the file with `LAWS_CHILD=1 go test ./laws -run TestPrintOutputDigest -v` and copy the digest. The digest was recorded on arm64 and also matched when the tests were built for amd64 on the same Mac, which is the one place a float result could differ between machines; a run on Linux has not been made.

## Hash function

The hash is open on the QA page. The test takes it as a field of `Engine` and the real engine uses SHA-256 of the standard text, as a stand-in until the publication's own hash is chosen.

## Still open

- The laws use the stand-in `fin.pmt` of `FixtureLibrary`. When the first ten library functions (ISS-27) load through `Publish`, `RealEngine().Library` should return the registry that loads them, and their reference files should be replayed here; ISS-27 keeps its own manifest of references in `functions/reference/`, and whether this file or that one is the record is for the owner.
- The registry has no listing of its functions, so law 6 can only check the versions of functions that have at least one record. A listing in `library/` would close that.
- Law 5 over pinned library versions replays every published version; how many past versions to replay is open on the QA page.
- Case counts are proposed (500 and 20,000).
- The `nightly` line in the root `Makefile`.
- The 2 platforms for the machine check are proposed; only arm64 and amd64 on one Mac have been compared. The cross-machine digest does not cover pinned library calls, whose float step (`math.Pow`) is the one place a result could differ between machines.
- The hash function of law 3 is a stand-in (SHA-256 of the standard text).
- A text input's wrong-type plant is not a problem, so law 7's expected list leaves it out; the row generator could plant a rule break for text (a text over a length) instead.
