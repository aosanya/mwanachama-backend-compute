# Library versioning cases

Status: draft. Cases for [library-versioning.md](../2.%20design/library-versioning.md). All names and figures are invented.

The loan used throughout: 12,000 borrowed at 12% a year, paid monthly over 24 months, so the monthly rate is 0.01 and n is 24. `fin.pmt@1` gives 564 and `fin.pmt@2` gives 564.88.

## Notes

- This file is named for the design page without the `.cases.md` suffix that [example-cases.md](example-cases.md) proposes. The owner can rename it.
- The problem codes `version-missing` and `version-unknown` are proposals. The formula language page must define them before the check accepts them.
- The design page has no `formula` blocks and no `case:` marker lines, so the example-to-row check does not apply to this file.
- Limits as in example-cases.md: a case name is 3 to 60 characters, a rule is 1 to 120 characters, a formula is at most 200 characters.
- The scale of a decimal result (`564` against `564.00`) follows the open question on exact text in example-cases.md. The rows write the figures as the design page does.

## Part 1: cases a formula engine decides alone

| name | formula | inputs | expected | rule | page |
|---|---|---|---|---|---|
| unversioned-call-refused | fin.pmt(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | problems: version-missing | A call that names no version is refused. | documentation/2. design/library-versioning.md |
| unknown-version-refused | fin.pmt@7(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | problems: version-unknown | A call to a version the library does not have is refused. | documentation/2. design/library-versioning.md |
| pmt-version-1-loan | fin.pmt@1(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:564 | Version 1 rounds the payment down to a whole number. | documentation/2. design/library-versioning.md |
| pmt-version-2-loan | fin.pmt@2(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:564.88 | Version 2 rounds the payment to the nearest 0.01. | documentation/2. design/library-versioning.md |
| same-host-newer-minus-older | fin.pmt@2(principal, rate, months) - fin.pmt@1(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:0.88 | One host runs both versions in one formula. | documentation/2. design/library-versioning.md |
| same-host-older-minus-newer | fin.pmt@1(principal, rate, months) - fin.pmt@2(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:-0.88 | The same two versions with the order swapped give the opposite sign. | documentation/2. design/library-versioning.md |
| known-wrong-still-evaluates | fin.pmt@1(principal, rate, months) | principal=dec:12000.00, rate=dec:0.01, months=int:24 | result: dec:564 | A known-wrong version keeps its result; only a notice and a warning are added. | documentation/2. design/library-versioning.md |

## Part 2: cases that need a publication and a draft

Each row is one rule of the design page, in the order of its sections, then two rows the page implies. The scan texts are the sentences the design page gives.

| name | setup | action | expected | rule |
|---|---|---|---|---|
| unversioned-refused-on-save | Hillside Seed Buyers has a draft. Library versions are `fin.pmt@1` and `fin.pmt@2`. | Save the formula `fin.pmt(12000, 0.01, 24)`, then save `fin.pmt@2(12000, 0.01, 24)`. | First save refused with `fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2.` Second save succeeds. | A version is always a choice (section 1). |
| hash-covers-each-call-version | Lakeside Savings Circle's March publication records `fin.pmt@1` for the loan screen and `fin.pmt@1` for the arrears report: 2 calls. | Change only the loan screen to `fin.pmt@2` and compute the content hash of the draft. | The hash differs from March's. The change publishes as a new publication. Nothing else in the draft changed. | The publication records a version per call and the hash covers them (section 2). |
| old-version-still-served | The host serves `fin.pmt@1` and `fin.pmt@2`. March records `fin.pmt@1`. Hillside's new publication records `fin.pmt@2`. | In July, run the 12,000 loan on both publications. | Lakeside gives 564. Hillside gives 564.88. The host lists both versions. | A released version is never edited and never removed (section 3). |
| fix-ships-as-next-version-with-note | In June `fin.pmt@1` is found to round down to a whole number. | Read the release of the fix. | The fix is `fin.pmt@2`. `fin.pmt@1` is unchanged. The note has 3 parts: what changed, for whom it matters, and one differing result (12,000 at 12% over 24 months, was 564, now 564.88). | A fix ships as the next version with a three-part note (section 4). |
| notice-and-one-step-move | July. Lakeside's draft holds `fin.pmt@1` in 2 calls. `fin.pmt@2` exists. | Open the function in the studio and the review. Press the move. Publish with a new tag. | Both notices carry the note. Before the publish, March still gives 564. The move rewrites 2 calls and changes nothing else. The new publication gives 564.88. March is unchanged and still available. | The builder is told and moves in one step; no published agency changes until that publish (section 5). |
| known-wrong-louder-and-warns | August. `fin.pmt@1` is marked known-wrong with `fin.pmt@2` as the replacement. Riverbend Market Women's April publication records `fin.pmt@1` in 1 place. | Open the notice. Publish a change to an unrelated screen. | The known-wrong notice shows above the ordinary kind and names `fin.pmt@2`. The warning names 1 place. The publish goes through. April still gives 564. | A known-wrong version is louder and warns, and does not block (section 6). |
| unknown-version-blocks-publish | A formula holds `fin.pmt@7(12000, 0.01, 24)`. The library has versions 1 and 2. | Run the agency scan, then publish. | The scan reports `Error: fin.pmt@7 does not exist. Latest is fin.pmt@2. Publish refused.` No publication is made. | A call to a version that does not exist blocks the publish (section 7). |
| no-library-call-hashes-as-before | A publication made before versions were recorded, with no library call. | Compute its content hash with the current code. | The hash equals the stored hash. | With no call there is nothing recorded and nothing added to the hash. |
| two-versions-in-one-draft-kept | A draft calls `fin.pmt@1` on the loan screen and `fin.pmt@2` on the arrears report. | Publish, then read the publication. | The publication records 2 calls, one at version 1 and one at version 2. The loan screen gives 564 and the arrears report gives 564.88. | Each call keeps its own version; versions are not merged per function. |

## Open for the owner

Can a known-wrong version ever be made to stop running (for example when a result is unsafe), and who may decide that?

This is not decided. Until it is, a known-wrong version keeps running and only warns. No case on this page assumes either answer: the case `known-wrong-still-evaluates` and the row `known-wrong-louder-and-warns` describe the default only, and both change when the owner answers.
