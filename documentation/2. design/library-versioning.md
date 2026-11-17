# Library versions are pinned per publication

When a library function is found wrong, an agency that already published keeps the version it published with until its builder upgrades. The fix ships as a new version. The old version stays untouched.

End goal: an agency's numbers never change under it without its builder choosing, and a builder always learns when a better version of a function exists.

Decided by the owner on 2026-11-17. All names and figures below are invented.

## The five facts

| Fact | Rule |
|---|---|
| Name form | `<library>.<function>@<whole number version>`. Versions start at 1 and go up by 1. |
| Where the version is recorded | In the publication document, once per call. |
| What is hashed | The publication's content hash covers every recorded version. |
| What is never done | A released version is never edited and never removed. |
| What a notice carries | The release note of the newer version: what changed, for whom it matters, one differing result. |

## The story used on this page

`fin.pmt` returns the monthly payment on a loan. In March, the cooperative "Lakeside Savings Circle" published its agency with `fin.pmt@1`. In June the library found that `fin.pmt@1` rounds the payment down to a whole number instead of to the nearest 0.01. The fix is `fin.pmt@2`.

The loan used throughout: 12,000 borrowed at 12% a year, paid monthly over 24 months, so the monthly rate is 0.01 and n is 24.

| Version | Monthly payment | Over 24 months |
|---|---|---|
| `fin.pmt@1` | 564 | 13,536.00 |
| `fin.pmt@2` | 564.88 | 13,557.12 |

The exact value is 564.8816667. The difference is 0.88 a month and 21.12 over the loan.

## 1. A function is named with its version

A library function is called as `fin.pmt@1`. A call that names no version (`fin.pmt`) is refused when the formula is saved, so a version is always a choice.

Example: the cooperative "Hillside Seed Buyers" saves the formula `fin.pmt(12000, 0.01, 24)`. The save is refused with: `fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2.` The builder rewrites it as `fin.pmt@2(12000, 0.01, 24)` and the save succeeds.

## 2. The publication records the version of every call

The publication records the version of every library function its formulas call, and the publication's content hash covers them. Changing only a version is therefore a change that can be published.

Example: Lakeside's March publication records `fin.pmt@1` for the loan screen and `fin.pmt@1` for the arrears report: 2 calls, 2 recorded versions. If the builder changes only the loan screen to `fin.pmt@2`, the content hash differs from March's and the change publishes as a new publication.

## 3. One running host serves every version still in use

One running host serves every version still in use. This means a released version is never edited and never removed. A guard test that fails the build if a released version's behaviour changes or a version disappears belongs to the library reference issue and is not specified here.

Example: in July the host serves `fin.pmt@1` for Lakeside's March publication and `fin.pmt@2` for Hillside's new publication, side by side. Lakeside's loan of 12,000 still comes out at 564 a month.

## 4. A fix ships as the next version, with a note

The fix for `fin.pmt@1` ships as `fin.pmt@2`. Every new version carries a plain note with three parts: what changed, for whom it matters, and one result that differs (input, old result, new result).

Example note for `fin.pmt@2`:

| Part | Text |
|---|---|
| What changed | The payment is rounded to the nearest 0.01. It was rounded down to a whole number. |
| For whom it matters | Any agency that shows or collects a loan payment. |
| A result that differs | Input: 12,000 at 12% a year over 24 months. Old result: 564. New result: 564.88. |

## 5. The builder is told, and moves in one step

When a newer version exists, the builder sees a notice on the function in the studio and on the agency's review. The notice carries the note. The builder can move to the new version in one step. Moving is an edit of the draft and is published like any other change. Nothing changes for a published agency until that publish.

Example, July: Lakeside's builder opens the loan formula in the studio and sees:

> `fin.pmt@2` is available. Changed: payment rounded to the nearest 0.01, not down to a whole number. Matters for: any agency that shows or collects a loan payment. Example: 12,000 at 12% a year over 24 months, was 564, now 564.88. [Move to `fin.pmt@2`]

The same notice shows on the agency's review. Until the builder presses the button and publishes, the March publication keeps calculating 564 a month. Pressing the button edits the draft: `fin.pmt@1` becomes `fin.pmt@2` in 2 calls. The builder publishes in July as a new publication with its own tag. The March publication is not changed and stays available.

## 6. A version found wrong is marked known-wrong

A version found wrong is marked known-wrong, with its replacement named. The notice for a known-wrong version is louder than the ordinary one. A publish that still uses a known-wrong version shows a warning and is not refused.

Example, August: the library marks `fin.pmt@1` known-wrong and names `fin.pmt@2` as its replacement. The cooperative "Riverbend Market Women" published in April with `fin.pmt@1` and has not upgraded.

Notice in the studio and on the review, shown above the ordinary kind:

> `fin.pmt@1` is known wrong. Use `fin.pmt@2`. Loan payments of 12,000 at 12% a year over 24 months are 564, not 564.88. [Move to `fin.pmt@2`]

Warning when Riverbend publishes a change to an unrelated screen while still on `fin.pmt@1`:

> Warning: this publication calls known-wrong `fin.pmt@1` in 1 place. Replacement: `fin.pmt@2`. Publish anyway or move first.

The publish goes through if the builder chooses it. Riverbend's April publication keeps calculating 564 until the builder moves.

## 7. A call to a version that does not exist blocks the publish

The agency scan reports a call to a version that does not exist as a blocking error. The publish is refused.

Example: a formula holds `fin.pmt@7(12000, 0.01, 24)`, but the library has versions 1 and 2. The scan reports:

> Error: `fin.pmt@7` does not exist. Latest is `fin.pmt@2`. Publish refused.

## 8. Stopping a version found unsafe

Decided by the owner on 2026-11-17 (decision 11 on `1. requirements/decisions.md`):

- A known-wrong version keeps running. The builder sees a loud notice and a one-step upgrade, as in section 6.
- Only the platform owner may stop a version found unsafe. The stop is for every agency at once. Each step that calls the stopped version then follows its own failure choice (the calculating step's choice of going to a named step or stopping and telling the person).
- No version is stopped automatically after a set time. There is no clock in this page's rules.
- A builder or an agency cannot stop a version for itself or for others; it moves to a newer version instead.

## Open for the owner

Still open, with no proposal made here:

- What counts as unsafe.
- How a stop and its reason are recorded.
- What people see when a step cannot run because its version is stopped.

This section needs the owner's answer before the page is final.
