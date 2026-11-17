# The library registry

Built for ISS-28, in the package `library/`. It serves library functions by name and version, so an agency's numbers never change under it without its builder choosing, and a call to a version that does not exist is refused. The design is [library-versioning.md](../2.%20design/library-versioning.md), the cases are [library-versioning.md](../4.%20qa/library-versioning.md), and the decisions are 10 (a function is called by name and version, a published version is never edited or removed) and 11 (a known-wrong version keeps running, only the platform owner may stop a version) in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## What the library does and what the host does

| The library | The host (the agency repos and the wakala API) |
|---|---|
| Holds every released version of every function, and serves each one | Records the version of each call in the publication and hashes it |
| Refuses a missing or unknown version with the page's sentence | Shows the scan error and blocks the publish |
| Says whether a version is known wrong, which version replaces it, whether a newer one exists and the note of the newer one | Draws the notice, the louder known-wrong notice, the warning on publish and the one-step move |
| Reads a stop from a place the host gives it, and fails a call to a stopped version with a sentence | Decides who may stop, records the stop and its reason, and sends the failing call to the step's failure choice |

The package imports the standard library and `formula/`. It has no time, random numbers, context, files or network, and a test reads the imports of every non-test file to keep it so.

## The surface

```go
type Ref struct{ Library, Name string; Version int }
func ParseRef(text string) (Ref, *formula.Problem)
func (r Ref) String() string

type Release struct {
	Library, Name string
	Version       int
	Signature     formula.Signature
	Apply         Implementation
	Note          Note
}

func NewRegistry() *Registry
func (r *Registry) Publish(release Release) error
func (r *Registry) Resolve(ref Ref) (Release, *formula.Problem)
func (r *Registry) MarkKnownWrong(library, name string, version, replacement int) error
func (r *Registry) Status(library, name string, version int) Status
func (r *Registry) NoteOf(library, name string, version int) (Note, bool)
func (r *Registry) ReadStopsFrom(source StopSource)

func Uses(tree formula.Tree) []Use
```

`*Registry` also has `Function`, `Versions` and `Apply`, so it is a `formula.Runtime`: it is handed to `formula.Plan`, and `Check` then refuses an unversioned call or a missing version while the formula is saved.

- **The name form.** `ParseRef` reads `library.function@N`. A name is lower-case letters, digits and underscores, starting with a letter, up to 40 characters. `N` is a whole number from 1 with no leading zero. A reference with no `@N` has `Version` 0, meaning none was written. A bad reference gives the code `bad-reference` (a proposal).
- **Publishing.** `Publish` takes one release. Versions start at 1 and go up by 1. Publishing a version that exists gives `ErrAlreadyPublished`, and skipping a number gives `ErrOutOfOrder`; in both cases nothing changes. A release needs a valid name, a version, a result type and an implementation (`ErrBadRelease`). The registry copies the signature, and `Function` and `Resolve` hand out copies, so a caller cannot edit what is stored.
- **The note.** Every version from 2 carries the three-part note of design section 4: what changed, for whom it matters, and one differing result (input, old, new). A version from 2 with any part empty is refused with `ErrNoteIncomplete`. Version 1 has no note.
- **No way to remove or edit.** The registry has no method that removes, replaces, edits or stops a version. A test reads the method set and fails if one appears.
- **Resolving.** `Resolve` gives the release, or one problem:

| Case | Code | Sentence |
|---|---|---|
| No version written | `version-missing` | `fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2.` |
| The version is not there | `version-not-found` | `` `fin.pmt@7` does not exist. Latest is `fin.pmt@2`. `` |
| The function is not there | `unknown-function` | `The function fin.later is not known.` |

These are the same codes and sentences `Check` gives for a saved formula. The QA page calls the second code `version-unknown`; the language page and the code say `version-not-found`, and that disagreement is open (below).

- **Known wrong.** `MarkKnownWrong(library, name, version, replacement)` marks a version and names its replacement, which must exist and be newer. A mark is made once and not changed. A known-wrong version runs exactly as before: no time passes into the library, so 30 days or 400 days change nothing, and a test runs it again and again. The notice and the warning belong to the host.
- **Status.** `Status` is what a notice needs in one read: `Exists`, `Latest`, `Newer`, `KnownWrong`, `Replacement`, `Stopped`, `StopReason` and `Note`. `Note` is the note of the replacement for a known-wrong version, otherwise the note of the latest version when a newer one exists.
- **Stopped.** The host gives the registry a `StopSource` with `ReadStopsFrom`. A call to a version the source reports as stopped does not run: `Apply` gives the problem `version-stopped`. Every program planned on the registry sees the stop at once, so two agencies follow it together, each by its own failure choice. How the stop is kept and who may set it is the host's. The sentence is proposed:

> `` `fin.pmt@1` is stopped by the platform owner and cannot run. Reason: it pays out too little. ``

  With no reason it ends after "cannot run."
- **Uses.** `Uses(tree)` lists every library call of a formula in reading order, with its library, name, version (0 when none was written) and its place as a `formula.Path`. Built-in functions are not listed. A host records these versions in the publication, counts the places of a known-wrong version for its warning (`formula.PartAt` gives the text of a place), and reports a version that does not exist as a blocking error.

## Concurrency

A `Registry` is safe for many goroutines. Reads take a shared lock, `Publish` and `MarkKnownWrong` take an exclusive one, and the function and the stop source are called with no lock held. A test publishes while eight goroutines read, and another runs two versions of one function from sixteen goroutines, under the race detector.

## Tests

The tests are in `library/registry_test.go` and `helpers_test.go`. They hold a small fixture of `fin.pmt` in two versions (564 and 564.88 for the loan of the QA page) and run the cases of Part 1 of [the QA page](../4.%20qa/library-versioning.md) through `formula.Plan`: the unversioned and unknown calls, version 1, version 2, and the two differences of 0.88 and -0.88, and a known-wrong version that keeps its result. The cases of Part 2 are host cases, listed below.

| QA case (Part 2) | Where it is covered |
|---|---|
| `unversioned-refused-on-save` | Here: `Check` through `Plan`. The save belongs to the agency. |
| `hash-covers-each-call-version` | Agency repo. The library gives `Uses`. |
| `old-version-still-served` | Here: two versions run side by side. |
| `fix-ships-as-next-version-with-note` | Here: the note is required and read back with `NoteOf`. |
| `notice-and-one-step-move` | Agency repo and studio. The library gives `Status`. |
| `known-wrong-louder-and-warns` | Agency repo. The library gives `Status` and `Uses`. |
| `unknown-version-blocks-publish` | Agency repo scan. The library gives `Resolve` and `Uses`. |
| `no-library-call-hashes-as-before` | Agency repo. |
| `two-versions-in-one-draft-kept` | Here: `Uses` lists each call with its own version. |
| `known-wrong-keeps-running` | Here. |
| `owner-stops-unsafe-version` | Here for the stop of every program at once. The failure choice is the host's. |
| `builder-cannot-stop-a-version` | Here there is no way to stop. The refusal and its wording belong to the host. |

## What the first ten functions need

ISS-27 builds the first ten library functions. They arrive here as releases, one `Publish` for each version, with the signature the language page gives them. This package has none of its own, and the fixture's `fin.pmt` lives in the tests only.

## Still open

Copied from [the design page](../2.%20design/library-versioning.md), decided by nobody here:

- What counts as unsafe.
- How a stop and its reason are recorded. The library takes a `StopSource` and records nothing.
- What people see when a step cannot run because its version is stopped. The sentence above is a proposal.

Open from this build:

1. `version-not-found` (language page and code) or `version-unknown` (QA page). Proposed: `version-not-found`.
2. Whether a known-wrong mark may be changed, for instance to name a better replacement. Proposed: no, a mark is made once, in the spirit of decision 10. A better replacement would be a newer version with a new mark.
3. Whether version 1 should also carry a note. Proposed: no, there is nothing earlier to differ from.
4. Whether `Publish` should be the only way a host loads releases, or whether the registry should also read them from a file. Proposed: only `Publish`, so the library stays free of files.
