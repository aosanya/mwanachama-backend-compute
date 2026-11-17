# Laws the formula engine must always keep

Status: the owner has not yet reviewed this page. Every number marked "proposed" is an invented starting point.

## Purpose

Hand-picked examples show that the engine works on the cases someone thought of. The laws below must hold on every input, so each one is tested on generated inputs. This repo computes values and agencies pin to release tags, so a broken law is a broken promise to every agency that pinned the version.

Terms used on this page. The meanings of Check and Validate are decision 18; the other meanings are proposed:

| Term | Meaning |
|---|---|
| Parse | Text in, formula tree out |
| Format | Formula tree in, standard text out |
| Check | Tests the formula text alone against everything the language sets: limits, shapes, names, functions, kinds, depth and cost (decision 18) |
| Validate | Tests a row of inputs against a formula and lists every problem found |
| batch run | Running one formula over many rows in one call |

## The formula generator

The formula generator produces random trees within the language's limits. Every tree it produces passes Check.

- Limits, equal to the language page's proposed limits so that every generated tree passes Check: depth at most 12, at most 100 nodes, at most 8 inputs per formula, cost at most 20000, formula text at most 500 characters.
- Proposed batch size: 1 to 200 rows per batch run. An agency sets its own row limit (decision 16), which is larger; the fixed cases at the limit and one over it are on the bulk runs page.
- Each case has a fixed seed. A failure prints the seed, and running that seed again produces the same tree and the same failure.
- A failing tree is shrunk to the smallest tree that still fails, and the shrunk tree is what the report shows.
- The nightly run logs its seed at the start, so any night can be replayed.

Other generators used below build on it:

| Generator | Produces |
|---|---|
| Formula generator | A random tree |
| Spacing generator | The standard text of a tree with random spaces, tabs and line breaks added between tokens, never inside a number or a name |
| Row generator | 1 to 200 rows of inputs for a tree, some valid, some with missing, extra or wrong-type values |
| Version generator | One of the pinned library versions that a publication may name |

## Law 1: Round trip

Statement: Parse(Format(tree)) equals tree. Saving always stores Format(Parse(text)), never the text as typed.

Generator: the formula generator, then Format each tree. For the saving half, the spacing generator.

Failing example:

| | |
|---|---|
| Input tree | `a - (b - c)` |
| Broken engine | Format drops the brackets and writes `a - b - c` |
| Parse of that text | `(a - b) - c`, a different tree |
| Test reports | `round trip failed, seed 41807, shrunk to a - (b - c): Parse(Format(tree)) read back as (a - b) - c` |

## Law 2: Standard shape is stable

Statement: formatting an already standard formula changes nothing. Format(Parse(Format(x))) equals Format(x).

Generator: the formula generator, then Format each tree to get standard text.

Failing example:

| | |
|---|---|
| Input tree | `max(a, 1.50)` |
| Broken engine | The first Format writes `max(a, 1.50)`, the second writes `max(a, 1.5)` |
| Test reports | `standard shape moved, seed 20931, shrunk to max(a, 1.50): first Format gave max(a, 1.50), second gave max(a, 1.5)` |

## Law 3: Spacing does not change the hash

Statement: the same formula typed with different spacing hashes the same. The hash is taken of Format(Parse(text)).

Generator: the spacing generator. Each case makes 2 spellings of one tree and compares their hashes.

Failing example:

| | |
|---|---|
| Spelling one | `a+b` |
| Spelling two | `a + b` |
| Broken engine | Hashes the typed text instead of the standard text |
| Test reports | `hash differs, seed 77215, shrunk to a+b and a + b: 2 spellings, 2 hashes` |

## Law 4: Batch equals single

Statement: a batch run gives, row by row, what running each row alone gives. Row order, row count and the problems per row are the same.

Generator: the formula generator for the tree and the row generator for the rows.

Failing example:

| | |
|---|---|
| Formula | `total / count` |
| Rows | row 1 `total 10, count 2`, row 2 `total 5, count 0`, row 3 `total 9, count 3` |
| Broken engine | The divide-by-zero in row 2 stops the batch run and row 3 gets row 2's problem |
| Test reports | `batch differs from single, seed 5530, row 3: alone gave 3, in the batch gave a division problem` |

## Law 5: Same inputs, same output, every time

Statement: one formula, one row and one library version give the same output on every run, on every machine, and for every library version a publication pins.

Generator: the formula generator and the row generator, run twice in one process, once in a second process, and once for each pinned version from the version generator. The machine check runs on the 2 platforms the CI offers (proposed: Linux and macOS).

Failing example:

| | |
|---|---|
| Formula | `sum(items)` over 3 items valued 0.1, 0.2, 0.3 |
| Broken engine | Adds in the order a map gives them, so the order differs from run to run |
| Test reports | `output moved between runs, seed 90412, run 1 gave 0.6000000000000001, run 2 gave 0.6` |

## Law 6: A published library function never changes

Statement: for every published function, the reference results recorded when it was published stay as recorded. A fix ships as a new version and the old one stays.

Generator: no random tree. The recorded reference results are the input: each function keeps a file of inputs and the outputs it gave at publication, and the test replays all of them against the current code. The row generator adds new rows to the recorded set only through a new version.

Failing example:

| | |
|---|---|
| Function | `round_half(x)` in `v0.1.0`, recorded as `round_half(2.5)` gives 3 |
| Broken engine | A fix edits the function in place so it gives 2 |
| Test reports | `reference changed, round_half v0.1.0, input 2.5, recorded 3, now 2; ship the fix as a new version` |

## Law 7: Validation first

Statement: a row with problems is never calculated, and its problem list is complete. Validate lists every problem in the row, not only the first.

Generator: the row generator, with rows that carry 0 to 5 planted problems (proposed). The test knows the planted problems, so it knows the full list.

Failing example:

| | |
|---|---|
| Formula | `price * quantity` |
| Row | `price` is the text "abc" and `quantity` is missing |
| Broken engine | Validate stops at the first problem |
| Test reports | `problem list incomplete, seed 12066, row 1: planted 2, listed 1, missing: quantity not given` |

A second failing example for the same law: a row with `quantity` equal to -1 where the language allows only 0 or more. The broken engine lists the problem and still calculates 40, and the test reports `row 1 has 1 problem and was calculated`.

## Cases per law

All counts are proposed.

- Normal run: 500 cases per law, started by `make test`.
- Nightly run: 20,000 cases per law, with a new seed each night. The seed is logged.
- Law 6 does not use random cases. Every recorded reference result is replayed in both runs.

## Summary

| Law | Generator | Normal cases | Nightly cases |
|---|---|---|---|
| 1 Round trip | Formula generator, spacing generator | 500 (proposed) | 20,000 (proposed) |
| 2 Standard shape is stable | Formula generator | 500 (proposed) | 20,000 (proposed) |
| 3 Spacing does not change the hash | Spacing generator | 500 (proposed) | 20,000 (proposed) |
| 4 Batch equals single | Formula generator, row generator | 500 (proposed) | 20,000 (proposed) |
| 5 Same inputs, same output | Formula generator, row generator, version generator | 500 (proposed) | 20,000 (proposed) |
| 6 A published function never changes | Recorded reference results | All recorded | All recorded |
| 7 Validation first | Row generator with planted problems | 500 (proposed) | 20,000 (proposed) |

## Open questions

The owner has not decided:

1. The final generator limits. Depth, nodes, inputs, cost and text length now equal the language page's proposed values and move with it. The random batch size of 1 to 200 rows stays a test choice; the agency's limit and its ceiling are open on the batch page.
2. The case counts of a normal run and a nightly run.
3. How many past library versions laws 5 and 6 replay.
4. The hash function for law 3.
5. Settled by decision 18: Check reads the formula text alone and covers everything about it, and Validate covers only the values given when the formula runs. The number is kept so the other references hold.
