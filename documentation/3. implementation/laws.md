# Laws as generated-input tests

The seven laws of [4. qa/laws.md](../4.%20qa/laws.md) run on generated trees and rows in the `laws/` folder. Every case has a seed, a failure prints the seed, a failing tree is shrunk, and the nightly run logs the seed it starts from. Built for ISS-32.

Every count, limit and name below that the QA page marks proposed is still proposed. All figures in the generators are invented.

## What runs

| Law | What the test checks | State |
|---|---|---|
| 1 Round trip | `Parse(Format(tree))` equals `tree` on 500 trees | runs in `make test` |
| 2 Standard shape is stable | `Format(Parse(Format(x)))` equals `Format(x)` | runs in `make test` |
| 3 Spacing does not change the hash | 2 spellings of one tree give one hash | runs in `make test` |
| 4 Batch equals single | needs the batch evaluation build, ISS-29 | waiting, listed as a skipped test |
| 5 Same inputs, same output | each tree is run over 20 generated rows twice in one process, once in a second process, and a digest of all outputs can be compared across machines | runs in `make test`; pinned library versions and the cross-machine check wait, see below |
| 6 A published function never changes | needs the first ten library functions with recorded results, ISS-27 | waiting, listed as a skipped test |
| 7 Validation first | a row with a planted missing or wrong-type input is never calculated | the never-calculated half runs; the complete problem list needs Validate and waits (ISS-26) |

`go test ./laws` lists the waiting ones as skipped tests with the build they wait for in the skip line, so they show up in a verbose run.

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
| 5 | `output moved between runs, seed N, shrunk to T: run 1 gave X, run 2 gave Y` |
| 7 | `row R has 1 problem and was calculated, seed N, shrunk to T` |

`Engine` holds the five things the laws call (Parse, Format, Check, Hash and Run), so the real engine is `RealEngine()` and a broken one is a copy with one field replaced. Law 7 adds the seed at the end of its sentence, which the page does not have.

## Each law fails on a broken engine

`laws/proofs_test.go` holds one broken engine per law and checks that the law finds it within 500 cases, with the page's sentence, a seed, and a seed that reproduces the same message while the real engine passes that seed.

| Law | Broken engine |
|---|---|
| 1 | Format drops every bracket |
| 2 | Format trims a trailing zero of a decimal each time it runs |
| 3 | The hash is taken of the typed text and not of the standard text |
| 5 | Run adds something that changes from one call to the next |
| 7 | Run turns the missing-input problem into a value |

The page asks for the proof in a test copy of the engine. The proof here replaces one function of the engine in the test instead, so nothing is copied and the committed code is never edited.

## Cases and the nightly run

| Setting | Meaning |
|---|---|
| `LAWS_CASES` | cases per law; 500 when not set (proposed) |
| `LAWS_SEED` | the number the seeds count up from; 0 when not set, so a normal run uses seeds 1 to 500 |

Every law logs `N cases, seeds A to B` first, so a verbose run names its seeds. `make -C laws nightly` runs 20,000 cases per law from the current time as the base seed, and the seed range is in its output. One nightly run took about 70 seconds. A failing seed is replayed with `LAWS_SEED=<seed - 1> LAWS_CASES=1`.

The root `Makefile` is shared with other work in progress, so the `nightly` target lives in `laws/Makefile`; a line `nightly: ; $(MAKE) -C laws nightly` in the root `Makefile` is left to add when that file is free.

## Law 5 across processes and machines

The second process is the test binary run again with `LAWS_CHILD=1`; it prints the digest of every outcome over the same seeds and the first process compares it with its own.

Across machines the test compares the digest of seeds 1 to 500 with `laws/testdata/output-digest.txt`. It is off unless `LAWS_PLATFORM_GOLDEN=1`, because a legitimate change of the engine moves the digest and would then fail every test run until the file is regenerated. The CI turns it on for both platforms it offers (Linux and macOS, proposed). Regenerate the file with `LAWS_CHILD=1 go test ./laws -run TestPrintOutputDigest -v` and copy the digest. The digest was recorded on arm64 and also matched when the tests were built for amd64 on the same Mac, which is the one place a float result could differ between machines; a run on Linux has not been made.

## Hash function

The hash is open on the QA page. The test takes it as a field of `Engine` and the real engine uses SHA-256 of the standard text, as a stand-in until the publication's own hash is chosen.

## Still open

- Laws 4 and 6, and the complete-problem-list half of law 7, switch on when ISS-29, ISS-27 and ISS-26 are done.
- Law 5 over pinned library versions needs the version generator and a library; it waits for ISS-27 and ISS-28.
- How many past versions laws 5 and 6 replay is open on the QA page.
- Case counts are proposed (500 and 20,000).
- The `nightly` line in the root `Makefile`.
- The 2 platforms for the machine check are proposed; only arm64 and amd64 on one Mac have been compared.
