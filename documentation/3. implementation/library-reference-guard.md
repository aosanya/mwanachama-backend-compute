# Library reference guard

Status: design, not built. Nothing on this page exists as code or as data files yet. The paths, names and sentences below are proposals until the owner reviews them. All names and figures are invented.

The rule and the test that protects it. The cases being protected are on [library-reference.md](../4.%20qa/library-reference.md); the sources behind them are on [sources.md](../1.%20requirements/sources.md). The versioning decision is on [library-versioning.md](../2.%20design/library-versioning.md), whose section 3 leaves this guard to the library reference issue.

## The rule

A published function version is never edited and never removed (CLAUDE.md rule 4, library-versioning.md section 3). Its recorded cases are part of it. A case of a published version is never changed and never removed. A fix is the next version, with its own cases beside the old ones.

## Where the cases live

What exists today: the repo root holds `compute.go`, `compute_test.go`, `go.mod`, `Makefile`, `LICENSE`, `README.md`, `CLAUDE.md` and `documentation/`, and an untracked `spike/` folder that is not part of this design. There is no `testdata/` folder and no library function. The layout below is a proposal because the repo layout for library code is not decided.

| Path | What it holds |
|---|---|
| `testdata/reference/<library>.<function>@<n>.json` | The cases of one function version. One file per version. |
| `testdata/reference/manifest.json` | The state of every version and a hash of every case. |

For example `testdata/reference/fin.balance_after_payment@1.json`.

Each case holds the fields of the case shape on the reference page: `id`, `kind`, `inputs`, `expected`, `tolerance` and `source`. The working and the section number of a source stay on the markdown page only. At most 200 cases per data file (proposed, the same limit as the case files on [example-cases.md](example-cases.md)). Today the reference page has 10 functions and 80 cases, 8 per function.

One case in a data file, as the loader would read it:

```json
{
  "id": "REF-fin.balance_after_payment-01",
  "kind": "normal",
  "inputs": {"balance": "120.50", "payment": "45.25"},
  "expected": "75.25",
  "tolerance": "exact",
  "source": "S-01"
}
```

The values are text in the notation of example-formulas.md, so a decimal is never turned into a float on the way in. An input that is not given is left out of `inputs`.

Second check: the markdown page and the data files must list the same ids. A case id on the page with no case in the data file, or the other way round, is failure 7 below. The page keeps the working; the data file keeps what the test needs.

## The manifest

For every function version the manifest holds a state, `draft` or `published`, and for each case its id and a SHA-256 of the case's canonical form.

### The canonical form

One line of JSON, written the same way every time:

1. The fields in this order: `id`, `kind`, `inputs`, `expected`, `tolerance`, `source`.
2. `inputs` as an object whose keys are in the function's parameter order, the order of its signature on the reference page.
3. Every value is a string. A list is an array of strings.
4. The `tolerance` is the function's tolerance, copied into every case of it: `exact`, or `rel:1e-9` for the float ratio.
5. `source` is the source id only, such as `S-01`.
6. No spaces and no line breaks outside the strings, no trailing newline.
7. UTF-8. Characters are written as themselves; only the quote, the backslash and control characters are escaped.

The working and the section number are not hashed. A case that is hashed is what the function must do; the working is only how the writer found it.

Worked example, the first case of `fin.balance_after_payment@1`, with the source written as its id only. The canonical form is this single line:

```
{"id":"REF-fin.balance_after_payment-01","kind":"normal","inputs":{"balance":"120.50","payment":"45.25"},"expected":"75.25","tolerance":"exact","source":"S-01"}
```

Its SHA-256 is `345008ae79f497d9c4999c1075c9ba95b63f68866d2c3ab2bfa6ad9dd893daab`. This figure was computed with a throwaway script kept outside the repo (a few lines of Python that build the line above and hash it as UTF-8) and checked with `shasum -a 256` on the same line. The guard test computes every hash itself; the hashes in the manifest are never typed by hand.

The manifest then holds, in this shape (one case shown; the real entry has all 8):

```json
{
  "fin.balance_after_payment@1": {
    "state": "published",
    "cases": {
      "REF-fin.balance_after_payment-01": "345008ae79f497d9c4999c1075c9ba95b63f68866d2c3ab2bfa6ad9dd893daab"
    }
  }
}
```

A hash is 64 hexadecimal characters. The messages below print the first 12.

## What the guard test checks

A Go test in the repo root, run by `go test ./...` and so by `make test`. It reads the data files, the manifest and the markdown pages. Each failure prints one sentence and the case id; the test fails the build. Sentences are proposals.

| # | Failure | The sentence |
|---|---|---|
| 1 | A case of a published version has a different hash from the manifest. | `REF-fin.pmt-01: this case of published fin.pmt@1 has changed (recorded 15228064b72f, now 79b8599388a4). A published version is never edited. Ship the fix as fin.pmt@2.` |
| 2 | A case of a published version is in the manifest and missing from its data file. | `REF-fin.pmt-01: this case of published fin.pmt@1 is missing. A case of a published version is never removed.` |
| 3 | A published version has no data file, or the library no longer has the version. | `fin.pmt@1 is published but testdata/reference/fin.pmt@1.json does not exist.` and `fin.pmt@1 is published but the library no longer has it. A published version is never removed.` |
| 4 | A case, in a data file or on the page, has no source id, or its source id has no row on sources.md. | `REF-fin.pmt-03: this case has no source id. A case with no source is not accepted.` and `REF-fin.pmt-03: source S-09 has no row in documentation/1. requirements/sources.md.` |
| 5 | Not a failure: a draft version's cases that are not yet in the manifest. | Nothing is printed. A draft may change freely and its hashes are recorded when it is published. |
| 6 | The library function's output differs from a recorded case. This is the reference test itself. | `REF-fin.pmt-01: fin.pmt@1 gave 565, the case records 564.` |
| 7 | The page and the data files list different ids. | `REF-fin.pmt-04 is on the page and not in fin.pmt@1.json.` or `REF-fin.pmt-04 is in fin.pmt@1.json and not on the page.` |

How failure 6 compares, as the reference page states:

| Result type | Passes when |
|---|---|
| Decimal and integer | The text is equal. Exact, difference 0, and `0.00` is not `0`. |
| Float | `abs(result - expected) <= 1e-9 x abs(expected)`, and exactly `0.0f` when the expected result is `0.0f`. |
| `refused` | The function gives no result. |

Failure 6 applies to every version, draft or published. It also names a result that should have been refused: `REF-fin.pmt-08: fin.pmt@1 gave 564, the case records refused.`

### A case added to a published version

Allowed. A new case records what the version does; it cannot change what it does. The guard passes it and the reference test runs it. It prints one notice: `REF-fin.pmt-03: new case of published fin.pmt@1 is not in the manifest yet.` The manifest gains the case through a small command, not yet designed, that reads the data files and adds an entry for each case the manifest does not hold. The command never changes or deletes an entry: if a case's id is in the manifest with a different hash, it stops and prints failure 1.

### A case changed or removed in a published version

Not allowed. The only way forward is the next version.

## How a version becomes published

Proposed: the manifest state of a version flips from `draft` to `published` in the commit on which a release tag is cut. The repo cuts release tags, agencies pin to them, and `v0.1.0` comes when the first working version passes its tests (CLAUDE.md). From that commit on, failures 1 to 3 apply to the version. This is an open question below; the owner has not decided what makes a version published.

## A worked story

The invented function `fin.pmt` from library-versioning.md. `fin.pmt@1` gives 564 for 12,000 at 12% a year over 24 months, and `fin.pmt@2` gives 564.88. The cases below are written only for this story, and the source `S-03` stands in; a real `fin.pmt` case would need its own row on sources.md.

`fin.pmt@1` is published. Its case, with the manifest entry made from it:

```json
{
  "id": "REF-fin.pmt-01",
  "kind": "normal",
  "inputs": {"principal": "12000.00", "rate": "0.01", "months": "24"},
  "expected": "564",
  "tolerance": "exact",
  "source": "S-03"
}
```

```json
{
  "fin.pmt@1": {
    "state": "published",
    "cases": {
      "REF-fin.pmt-01": "15228064b72fe995697f7ab7d8f3c53c3c31a7fbafc5f2a03759c7a72cf257cb"
    }
  }
}
```

### The wrong move

A writer sees that the right payment is 564.88 and changes `expected` in `fin.pmt@1.json` to `564.88`. The hash of the changed case is `79b8599388a46b725d892b4eb15dffcf0731f3262a3ba4e69cdddecaa9680ab8`. The guard fails with failure 1:

`REF-fin.pmt-01: this case of published fin.pmt@1 has changed (recorded 15228064b72f, now 79b8599388a4). A published version is never edited. Ship the fix as fin.pmt@2.`

Failure 6 also fails if the function was changed to give 564.88, because `fin.pmt@1` would no longer give 564 to the agencies that pinned it.

### The right move

Add `fin.pmt@2` as a new function and `testdata/reference/fin.pmt@2.json` with its own cases. The case numbers count on across versions, so the first case of version 2 is `REF-fin.pmt-02`:

```json
{
  "id": "REF-fin.pmt-02",
  "kind": "normal",
  "inputs": {"principal": "12000.00", "rate": "0.01", "months": "24"},
  "expected": "564.88",
  "tolerance": "exact",
  "source": "S-03"
}
```

Both versions go in the manifest. `fin.pmt@1` keeps its case and its hash. `fin.pmt@2` stays `draft` until it is published:

```json
{
  "fin.pmt@1": {
    "state": "published",
    "cases": {
      "REF-fin.pmt-01": "15228064b72fe995697f7ab7d8f3c53c3c31a7fbafc5f2a03759c7a72cf257cb"
    }
  },
  "fin.pmt@2": {
    "state": "draft",
    "cases": {
      "REF-fin.pmt-02": "891eb53c82e9fd8ba746aa97d295958437925e8f1f387b1f57d7696a8dbca783"
    }
  }
}
```

The hashes above were computed with the same outside script as the first one. The guard passes: `fin.pmt@1` gives 564, `fin.pmt@2` gives 564.88, and the old case is unchanged.

## What the guard does not do

- It does not choose the rounding mode. The five cases that sit on a half stay as the reference page records them, until the owner decides (exact-money.md open question 1).
- It does not check the formula language. It only compares what a function gives with what a case records.
- It does not stop a commit that changes a case and its hash in the manifest together. See the open questions.
- It does not say whether a version is right. A known-wrong version keeps its cases and they keep passing.

## Open questions

1. What makes a version published? The proposal above is the release tag commit. The owner has not decided.
2. Who may mark a version published, and who may add a version to the manifest?
3. Does a known-wrong version's cases stay in the manifest? Yes: a released version is never removed, per [library-versioning.md](../2.%20design/library-versioning.md). Whether a known-wrong version can ever be stopped from running is the owner's open question on that page, and not decided here.
4. Where does the manifest live if the repo layout is not `testdata/reference/`? The path is a proposal.
5. Does the guard compare against the manifest as it was at the last release tag? Without it, a commit that edits a published case and its hash in the manifest together passes, and only review stops it. With it, the test needs the git history, which a pure library may not want to read.
6. How does the test ask the library which versions it serves? Failure 3 needs a list of function versions from the library, which does not exist yet.
