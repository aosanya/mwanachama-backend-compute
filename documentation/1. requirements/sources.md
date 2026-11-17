# Sources

Status: draft. Every figure on this page's cases is invented.

The owner's rule: every research source goes into the repo. A figure that a recorded case rests on comes from a source, and the source has a row on this page. A case with no source is not accepted. The cases are on [library-reference.md](../4.%20qa/library-reference.md); the test that protects them is on [library-reference-guard.md](../3.%20implementation/library-reference-guard.md).

No outside source has been opened yet. The table has none. All three rows are inside this repo.

## One row per source

| Id | Kind | Where | What it says | Cases resting on it | Date added |
|---|---|---|---|---|---|
| S-01 | In-repo worked example | `documentation/1. requirements/example-formulas.md`, the "Worked example" paragraph of sections 1 to 19 | Each of the nineteen formulas has an input and the figure it must give. | fin.balance_after_payment 01, 02; fin.equal_share 01; date.late_days 01, 02; fin.flat_instalment 01; fin.prorate_fee 01; fin.straight_line_value 01; fin.coverage_ratio 01; fin.proportional_share 01; stats.spread 01; math.remainder 01; fin.loan_within_limit 01; fin.farmer_net_payout 01, 02; fin.late_filing_penalty 01, 02; fin.interest_step 01 to 03; fin.monthly_interest 01; edu.term_grade 01; fin.interim_payment 01; fin.lending_data_ok 01; member.voting_right 01, 02 | 2026-11-17 |
| S-02 | In-repo input rules and invalid examples | `documentation/1. requirements/example-formulas.md`, the input table and the "Invalid example" paragraph of the same sections | Each input has a type and rules (range, no negatives, places, required), and one input that breaks a rule is named. | fin.balance_after_payment 07, 08; fin.equal_share 06, 07, 08; date.late_days 07, 08; fin.flat_instalment 07, 08; fin.prorate_fee 07, 08; fin.straight_line_value 07, 08; fin.coverage_ratio 07, 08; fin.proportional_share 07; stats.spread 07, 08; math.remainder 06, 07, 08; fin.loan_within_limit 07, 08; fin.farmer_net_payout 07, 08; fin.late_filing_penalty 08; fin.interest_step 07, 08; fin.monthly_interest 08; edu.term_grade 07, 08; fin.interim_payment 07, 08; fin.lending_data_ok 07, 08; member.voting_right 07, 08 | 2026-11-17 |
| S-03 | Hand calculation | The Working column of the case, on `documentation/4. qa/library-reference.md`, shown step by step | The figure is worked out in whole steps from the inputs and the function's rule. | fin.balance_after_payment 03 to 06; fin.equal_share 02 to 05; date.late_days 03 to 06; fin.flat_instalment 02 to 06; fin.prorate_fee 02 to 06; fin.straight_line_value 02 to 06; fin.coverage_ratio 02 to 06; fin.proportional_share 02 to 06 and 08; stats.spread 02 to 06; math.remainder 02 to 05; fin.loan_within_limit 02 to 06; fin.farmer_net_payout 03 to 06; fin.late_filing_penalty 03 to 07; fin.interest_step 04 to 06; fin.monthly_interest 02 to 07; edu.term_grade 02 to 06; fin.interim_payment 02 to 06; fin.lending_data_ok 02 to 06; member.voting_right 03 to 06 | 2026-11-17 |

Each case id is `REF-<function>-<nn>`, so "fin.equal_share 01" is `REF-fin.equal_share-01`.

## How many cases rest on each source

| Id | Cases |
|---|---|
| S-01 | 26 |
| S-02 | 37 |
| S-03 | 89 |
| Total | 152 |

The 152 case ids on [library-reference.md](../4.%20qa/library-reference.md) were checked against the table above, case by case: each id appears in exactly one row, and the count for each source is the count of the case table's Source column.

## The kinds

| Kind | Meaning |
|---|---|
| In-repo worked example | A figure given in a page of this repo. |
| In-repo rules | A rule or an invalid example given in a page of this repo. |
| Hand calculation | A figure worked out in steps by the writer. The steps are kept with the case. |
| Public worked example | A figure printed in a public document, opened and read by the writer. |
| Standard | A published standard, opened and read by the writer. |
| Textbook | A book, opened and read by the writer, with its edition and page. |

## How to add a source

1. Take the next id: `S-04` after `S-03`. An id is never reused.
2. Write the address or path in the Where column. For a page of this repo, the path and the section. For a public document, the address, or the title, edition and page for a book.
3. Write in one sentence, in your own words, what the source says. Never paste a long passage.
4. List the cases resting on it, written as function and case numbers, as the rows above do.
5. Put the date it was added.
6. Check that every case id on the reference page is in exactly one row and the counts above still add up.

Two rules:

- A case with no source is not accepted.
- A public figure is cited only after the source was opened and read. A figure remembered from somewhere is not a source.

## Open questions

1. Can a row change or go? Proposed: a row is never removed while a case rests on it, because a published case is never changed (see [library-versioning.md](../2.%20design/library-versioning.md)). A wrong address would then be corrected in place and the correction dated. The owner has not decided this.
2. The Working column of the reference page is the whole of source S-03. If the cases move into data files (see [library-reference-guard.md](../3.%20implementation/library-reference-guard.md)), the working stays on the page. Does S-03 then keep that page as its place, or does the working move into the data file?
