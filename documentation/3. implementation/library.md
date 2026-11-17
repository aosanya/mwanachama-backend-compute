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

The package imports the standard library, `formula/` and, for the ten functions, `functions/` with what that imports. It has no time, random numbers, context, files or network, and a test reads the imports of every non-test file to keep it so.

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
	Metered       MeteredImplementation
	Note          Note
}

func NewRegistry() *Registry
func (r *Registry) Publish(release Release) error
func (r *Registry) Resolve(ref Ref) (Release, *formula.Problem)
func (r *Registry) ApplyMetered(library, name string, version int, args []formula.Value, meter *formula.Meter) (formula.Value, *formula.Problem)
func (r *Registry) MarkKnownWrong(library, name string, version, replacement int) error
func (r *Registry) Status(library, name string, version int) Status
func (r *Registry) NoteOf(library, name string, version int) (Note, bool)
func (r *Registry) ReadStopsFrom(source StopSource)

func Uses(tree formula.Tree) []Use

func LoadFunctions(registry *Registry) error
func ReleaseOf(f functions.Function) (Release, error)
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

## Metering the callee

Built for ISS-77 at the request of ISS-72, whose [How a callee is metered](formula-run.md) section tells a `Runtime` that also has `ApplyMetered` to be given the caller's meter. `*Registry` is such a runtime (a compile-time assertion in `metered_test.go` holds it), so a loop through the registry no longer gets a fresh meter for each call.

- A `Release` may carry a `Metered` form, `func(args []formula.Value, meter *formula.Meter) (formula.Value, *formula.Problem)`. `Apply` stays required. `ApplyMetered` resolves the version, refuses a missing or stopped one with the same codes and sentences as `Apply`, and then runs `Metered` with the meter it was given, or `Apply` when the release has no metered form.
- A release whose body calls back into the registry (an agency function made from a program) sets `Metered` to `Program.ApplyMetered`. The ten functions of `functions/` call nothing in the registry, so they have only `Apply` and take no part in a loop.
- The tests in `library/metered_test.go` publish definitions that call themselves, two that call each other and one that calls itself twice. Each ends with `call-too-deep` or `too-costly` after fewer than 1000 host calls. The direct case ends at `call-too-deep` with the call depth sentence after 6 host calls: the entry, four nested calls, and the fifth nested call, which is refused. A callee declared at exactly what it spends leaves the run's spent total equal to the static cost, and a callee that spends more than it declared is charged the difference.

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

## The first ten functions

Built for ISS-77, in `library/load_functions.go`. The ten functions of [library-functions.md](library-functions.md) (ISS-27) are in `functions/`; the registry serves them once they are loaded.

```go
func LoadFunctions(registry *Registry) error
func ReleaseOf(f functions.Function) (Release, error)
```

- **`LoadFunctions`** publishes `functions.All()` in order, one `Publish` for each. A second load into the same registry is refused with `ErrAlreadyPublished`, and a function that does not prepare (its definition, body or result type is refused) stops the load with its reason.
- **`ReleaseOf`** makes one release from a function. The signature is the definition's inputs in their declared order with their types, the result type of the function, and a cost that is `formula.Cost` of the body. All ten are version 1, so none carries a note.
- **The call.** The release's `Apply` takes the typed values a formula passes. It writes each as the text `Value.String` gives (a list as its items), hands them to the function's `Callable` and so runs `Validate` first: an input rule of the function (a divisor from 1 to 100, money with at most two places, a list with at least one item) refuses a call made inside a formula the same way it refuses a row (decision 8), and the body never runs on a refused call. When the callee gives several problems the call gives one problem whose code is the first code and whose sentence joins all the sentences with a space. A problem of the body itself, such as `divide-by-zero`, comes back as it is.
- **Status.** The registry has no draft state, so the ten are served as released version 1 while the guard manifest still calls them `draft`. Whether a draft version should be served is part of the owner's open question on what makes a version published; nothing here decides it.
- **The package now imports `functions/`**, and through it `function/`, `number/`, `numtext/` and `problem/`. The test that keeps the clock, random numbers, files and the network out of the package still passes.

The tests are in `library/load_functions_test.go`. They load the registry, check the signatures against the definitions, replay every recorded case that can be typed (80 cases; a case with a text that is not a number cannot be typed and is refused by `Validate` in `functions/` already) through `Registry.Apply` with `functions.Judge`, and run three formulas that call the library: `fin.balance_after_payment@1` gives 75.25, `stats.spread@1` over a list gives 9.25, and `math.remainder@1` with a divisor of 0 is refused. A call to `stats.spread@2` is refused on save with `version-not-found`.

The basic runs of [basic-runs-suite.md](basic-runs-suite.md) now run on this registry: the ten `basic-call-*` cases, `lang-library-call-version-two` and `basic-pin-after-new-version` were held for ISS-27 and now pass. `fin.pmt@1` and `fin.pmt@2` are not among the ten, so that suite publishes a two-version `fin.pmt` of its own beside them. The example cases of [example-cases.md](example-cases.md) call no function of the ten; their stand-in library holds the two agency functions and `fin.pmt@2` and is left as it is.

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
5. Whether a function whose manifest state is `draft` should be served at all. Today the ten are served as soon as they are loaded.
6. A call from a formula writes each value as text and reads it back. That keeps the function's input rules in force, and it costs a round trip for a number that is already typed. Proposed: keep it until a measurement says it matters; ISS-36 measures the speed.
