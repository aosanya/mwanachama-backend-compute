# Input checks: complete, detailed problems

Status: the owner has not reviewed this page, and the sentences have not been read by the owner. Every number marked proposed is an invented starting point. All names and figures are invented.

## Purpose

A person or an AI that gives a function bad inputs is told everything that is wrong in one answer, in words they can act on. Decision 8 on `documentation/1. requirements/decisions.md` puts the input rules and the checks in the function's definition, so a step, a bulk run and an AI tool all get the same list. The rules, the checks, the problem and every sentence are set by `documentation/2. design/formula-language.md`, under "Input rules, checks and Validate". This page lists the cases that prove them, with the exact problem list each case must give.

What this page covers, and what it leaves to other pages:

| Covered here | Where else |
|---|---|
| Validate: one row against one definition, every rule and every check | The refusal of a definition on save is Check on the definition. Its cases are the `rules-definition-` rows of `4. qa/formula-examples.md`. |
| The problem list of a single run, a batch row and an AI tool call | The batch stage and the row position in a batch result are on `2. design/batch-evaluation.md` and `4. qa/bulk-runs.md`. |
| That a row with problems is never calculated | Law 7 on `4. qa/laws.md` proves it on generated rows. This page gives the hand-picked cases. |
| Problems that appear when a formula is calculated (`divide-by-zero`, `number-too-big`, `empty-list`) | They are codes of a run. Validate gives `check-error` when one of them happens inside a check. |

## How a case is written

A case names a definition (D1 to D11 below), the row received, and the problems the row must give, in order. A row that is not shown for an input holds the clean row of its definition, which is stated with the definition.

A problem is one object with eight fields: `input`, `item`, `given`, `value`, `rule`, `sentence`, `code` and `check`. A problem is written on one line:

```
code · input (and item) · value received — sentence
```

- The `rule` follows from the code: `required` gives rule `required`; `wrong-type` gives `type`; `too-small` and `too-large` give `range`; `list-too-short` gives `min-items`; `list-too-long` gives `max-items`; `no-negatives`, `decimal-places` and `not-allowed` (rule `allowed-values`) as named; `unknown-input` gives `known-input`; `check-failed` and `check-error` give `check`.
- `given` is true and `value` is the text received, written exactly as received. The words `not given` in the value place mean `given` false and `value` null.
- `item` is null unless the line says `item n`, which is the position counted from 1.
- A line of a check reads `code · check "formula" — sentence`. Its `input`, `item`, `given` and `value` are null, and its `check` field holds the standard text of the formula of the check.

How a case is compared (proposed): the test compares the whole list, field by field and in order, and a sentence is compared character for character. A list with a missing problem, an extra problem, a problem out of order or a sentence worded differently fails the case. This is the reading proposed in item 25 of "Differences to settle" on the language page, because two `no-negatives` problems on different inputs cannot be told apart by the code alone.

A row with no problem line gives an empty list and the formula is then calculated. A row with one or more problem lines is never calculated.

## The definitions

Each definition is the definition of an example formula on the language page, or is invented for this page where it says so. Every input that is not a list carries `required`, as the language page demands.

### D1 Balance after payment (formula 1)

| Input | Type | Rules |
|---|---|---|
| `balance` | decimal | required, no-negatives, decimal-places 2 |
| `payment` | decimal | required, no-negatives, decimal-places 2 |

| Check | Formula | Message |
|---|---|---|
| 1 | `payment > 0` | The payment must be more than zero |
| 2 | `payment <= balance + 1000.00` | The payment cannot be more than 1000.00 above the balance |

### D2 Farmer net payout (formula 4)

| Input | Type | Rules |
|---|---|---|
| `kilos` | integer | required, no-negatives |
| `price_per_kilo` | decimal | required, range more than 0.000, decimal-places 3 |
| `credit` | decimal | required, no-negatives, decimal-places 2 |

Clean row: `kilos` 853, `price_per_kilo` 0.437, `credit` 120.00. No checks.

### D3 Loan limit check (formula 2)

| Input | Type | Rules |
|---|---|---|
| `status` | text | required, allowed-values "active", "lapsed", "suspended" |
| `requested` | decimal | required, range more than 0.00, decimal-places 2 |
| `contributions` | list of decimal | min-items 1, no-negatives |

Clean row: `status` active, `requested` 1400.00, `contributions` [200.00, 250.00, 150.00, 100.00]. No checks.

### D4 Term grade (formula 10)

| Input | Type | Rules |
|---|---|---|
| `scores` | list of integer | min-items 1, max-items 12, range 0 to 100 |

Clean row: `scores` [72, 85, 64, 91, 77]. No checks. The `max-items` of 12 is added on the language page to show the rule.

### D5 Lending data check (formula 16)

| Input | Type | Rules |
|---|---|---|
| `balances` | list of decimal | min-items 1, decimal-places 2 |
| `floor` | decimal | required, no-negatives, decimal-places 2 |

| Check | Formula | Message |
|---|---|---|
| 1 | `count(balances) >= 6` | At least 6 balances are needed |

Clean row: `balances` [420.00, 380.50, 95.25, 510.00, 440.75, 300.00], `floor` 350.00. The check is invented for this page.

### D6 Flat-rate loan instalment (formula 9)

| Input | Type | Rules |
|---|---|---|
| `principal` | decimal | required, range more than 0.00, decimal-places 2 |
| `monthly_rate` | decimal | required, range 0.000 to 0.050, decimal-places 3 |
| `months` | integer | required, range 1 to 60 |

| Check | Formula | Message |
|---|---|---|
| 1 | `principal >= 100.00` | The loan must be at least 100.00 |
| 2 | `principal * monthly_rate * months <= principal` | The interest cannot be more than the amount borrowed |

Clean row: `principal` 2400.00, `monthly_rate` 0.012, `months` 7.

### D7 Each person's share (a check that can fail to work out)

| Input | Type | Rules |
|---|---|---|
| `total` | decimal | required, no-negatives, decimal-places 2 |
| `people` | integer | required |

| Check | Formula | Message |
|---|---|---|
| 1 | `total / people >= 1.00` | Each person must get at least 1.00 |

Clean row: `total` 100.00, `people` 3.

### D8 Type probe (invented)

| Input | Type | Rules |
|---|---|---|
| `kilos` | integer | required |
| `payment` | decimal | required |
| `ratio` | float | required |
| `note` | text | required |
| `active` | boolean | required |
| `filed` | date | required |

Clean row: `kilos` 12, `payment` 45.25, `ratio` 0.37, `note` ok, `active` true, `filed` 2026-03-31. No checks. Every case on this definition changes one input and leaves the others clean.

### D9 Fund share per member (formula 3)

| Input | Type | Rules |
|---|---|---|
| `deposits` | list of decimal | min-items 1, no-negatives |
| `expenses` | list of decimal | no-negatives |
| `members` | integer | required, range 1 to 500 |

Clean row: `deposits` [500.00, 750.00, 1250.50], `expenses` [120.25, 80.00], `members` 7. No checks.

### D10 Bounds and places probe (invented)

| Input | Type | Rules |
|---|---|---|
| `units` | integer | required, no-negatives, range 1 to 100 |
| `share_pct` | decimal | required, range 0.00 up to but not including 100.00 |
| `bonus` | decimal | required, range up to 500.00 |
| `bags` | decimal | required, decimal-places 0 |
| `weight` | decimal | required, decimal-places 1 |

Clean row: `units` 5, `share_pct` 40.00, `bonus` 100.00, `bags` 3, `weight` 2.5. No checks.

### D11 Item order probe (invented)

| Input | Type | Rules |
|---|---|---|
| `amounts` | list of decimal | min-items 1, no-negatives, range 0.00 to 100.00, decimal-places 2 |

Clean row: `amounts` [10.00, 25.50]. No checks.

## A. required

The rule says that a name must be in the row. An input that is not given gets `required` and no other problem.

**ic-a01** D1. `balance` 120.50, `payment` 45.25. No problem.

**ic-a02** D1. `balance` 200.00, `payment` not in the row. Both checks read `payment`, so neither runs.

1. `required` · `payment` · not given — `payment` was not given. It is required.

**ic-a03** D1. `balance` not in the row, `payment` 45.25. Check 1 reads only `payment`, so it runs and is true. Check 2 reads `balance`, so it is skipped.

1. `required` · `balance` · not given — `balance` was not given. It is required.

**ic-a04** D1. The row holds no name at all. Both inputs are listed, in the order they are declared.

1. `required` · `balance` · not given — `balance` was not given. It is required.
2. `required` · `payment` · not given — `payment` was not given. It is required.

**ic-a05** D2. `kilos` 853, `price_per_kilo` 0.437, `credit` not in the row.

1. `required` · `credit` · not given — `credit` was not given. It is required.

**ic-a06** D8. `active` not in the row.

1. `required` · `active` · not given — `active` was not given. It is required.

**ic-a07** D8. `filed` not in the row.

1. `required` · `filed` · not given — `filed` was not given. It is required.

**ic-a08** D8. `note` not in the row. A text input that is not given is told apart from an empty text, which is given (section J).

1. `required` · `note` · not given — `note` was not given. It is required.

## B. Lists: min-items and max-items

A list that is not given and has no `required` counts as an empty list. The count of items is tested even when an item has the wrong type.

**ic-b01** D3. The clean row. No problem.

**ic-b02** D3. `contributions` [].

1. `list-too-short` · `contributions` · [] — `contributions` has 0 items. It needs at least 1 item.

**ic-b03** D3. `contributions` not in the row. The list has no `required`, so it counts as empty, and `given` is false.

1. `list-too-short` · `contributions` · not given — `contributions` was not given, so it has 0 items. It needs at least 1 item.

**ic-b04** D9. `expenses` not in the row. The list has no `min-items` and no `required`, so it counts as empty and is accepted. No problem.

**ic-b05** D9. `expenses` []. No problem.

**ic-b06** D4. `scores` with 12 items, each 70. The list is at the limit. No problem.

**ic-b07** D4. `scores` with 13 items, each 70.

1. `list-too-long` · `scores` · [70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70] — `scores` has 13 items. It can have at most 12 items.

**ic-b08** D4. `scores` with 21 items, each 70. The value shows the first 20 items and then `, ...`.

1. `list-too-long` · `scores` · [70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, ...] — `scores` has 21 items. It can have at most 12 items.

**ic-b09** D4. `scores` with 13 items: twelve of 70, then abc. The count is tested although an item is of the wrong type. The list problem comes once, before the items.

1. `list-too-long` · `scores` · [70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, abc] — `scores` has 13 items. It can have at most 12 items.
2. `wrong-type` · `scores` item 13 · abc — Item 13 of `scores` is abc. It must be a whole number with at most 18 digits.

**ic-b10** D4. `scores` [].

1. `list-too-short` · `scores` · [] — `scores` has 0 items. It needs at least 1 item.

**ic-b11** D4. `scores` not in the row.

1. `list-too-short` · `scores` · not given — `scores` was not given, so it has 0 items. It needs at least 1 item.

**ic-b12** D4. `scores` received as the text 70, not as a list. The text cannot be read as a list, so the list gets `wrong-type` and nothing else.

1. `wrong-type` · `scores` · 70 — `scores` is 70. It must be a list of whole numbers.

## C. no-negatives

A value is negative when it is below zero. `-0.00` equals zero and is not negative.

**ic-c01** D1. `balance` 0.00, `payment` 45.25. Zero is not negative. No problem.

**ic-c02** D1. `balance` 80.00, `payment` -5.00. The checks read `payment` and are skipped.

1. `no-negatives` · `payment` · -5.00 — `payment` is -5.00. It cannot be negative.

**ic-c03** D1. `balance` 80.00, `payment` -0.00. `-0.00` is not negative, so no rule is broken. Both inputs are clean, so both checks run. Check 1 is false, because -0.00 is not more than zero. Check 2 is true.

1. `check-failed` · check "payment > 0" — The payment must be more than zero. `payment` is -0.00.

**ic-c04** D1. `balance` -1.00, `payment` -5.125. Three problems at once. `payment` breaks two rules, in the order of the rule table.

1. `no-negatives` · `balance` · -1.00 — `balance` is -1.00. It cannot be negative.
2. `no-negatives` · `payment` · -5.125 — `payment` is -5.125. It cannot be negative.
3. `decimal-places` · `payment` · -5.125 — `payment` is -5.125. It can have at most 2 decimal places.

**ic-c05** D3. `contributions` [200.00, -50.00]. The rule is applied to each item and the problem names the item.

1. `no-negatives` · `contributions` item 2 · -50.00 — Item 2 of `contributions` is -50.00. It cannot be negative.

**ic-c06** D10. `units` -3. The value is both below zero and below the range, so it gets both problems, `no-negatives` first.

1. `no-negatives` · `units` · -3 — `units` is -3. It cannot be negative.
2. `too-small` · `units` · -3 — `units` is -3. It must be from 1 to 100.

**ic-c07** D2. `kilos` 853, `price_per_kilo` -0.5, `credit` 120.00. This case reads the page as written, where an input with no `no-negatives` rule accepts a negative value, so `price_per_kilo` breaks only its range. It is pending until the owner settles open question 1: decision 22 says an input that does not declare negatives refuses them.

1. `too-small` · `price_per_kilo` · -0.5 — `price_per_kilo` is -0.5. It must be more than 0.000.

## D. range

A range compares by value, as exact decimals. Both bounds are allowed unless a bound is exclusive.

**ic-d01** D9. `members` 7. No problem.

**ic-d02** D9. `members` 1. The lower bound is allowed. No problem.

**ic-d03** D9. `members` 500. The upper bound is allowed. No problem.

**ic-d04** D9. `members` 0.

1. `too-small` · `members` · 0 — `members` is 0. It must be from 1 to 500.

**ic-d05** D9. `members` 501.

1. `too-large` · `members` · 501 — `members` is 501. It must be from 1 to 500.

**ic-d06** D3. `requested` 0.01. It is more than 0.00. No problem.

**ic-d07** D3. `requested` 0.00. The bound itself is refused when the minimum is exclusive.

1. `too-small` · `requested` · 0.00 — `requested` is 0.00. It must be more than 0.00.

**ic-d08** D6. `principal` 2400.00, `monthly_rate` 0.05, `months` 7. `0.05` equals the upper bound `0.050` by value. Both checks run and are true. No problem.

**ic-d09** D6. `principal` 2400.00, `monthly_rate` 0.0500, `months` 7. The value is within the range. The text has 4 places, which breaks the places rule. Check 2 reads `monthly_rate` and is skipped. Check 1 reads `principal` and is true.

1. `decimal-places` · `monthly_rate` · 0.0500 — `monthly_rate` is 0.0500. It can have at most 3 decimal places.

**ic-d10** D10. `share_pct` 0.00 and then 99.99. The lower bound is allowed and the upper bound is exclusive. No problem for either row.

**ic-d11** D10. `share_pct` -0.01. With one end exclusive the sentence gives the end that was broken.

1. `too-small` · `share_pct` · -0.01 — `share_pct` is -0.01. It must be 0.00 or more.

**ic-d12** D10. `share_pct` 100.00.

1. `too-large` · `share_pct` · 100.00 — `share_pct` is 100.00. It must be less than 100.00.

**ic-d13** D10. `bonus` 500.00. No problem.

**ic-d14** D10. `bonus` 500.01. A range with a maximum only.

1. `too-large` · `bonus` · 500.01 — `bonus` is 500.01. It must be 500.00 or less.

**ic-d15** D10. `units` 0.

1. `too-small` · `units` · 0 — `units` is 0. It must be from 1 to 100.

**ic-d16** D10. `units` 101.

1. `too-large` · `units` · 101 — `units` is 101. It must be from 1 to 100.

## E. decimal-places

The rule means at most `places` places, counted in the text received. Trailing zeros count.

**ic-e01** D1. `balance` 120.50, `payment` 45.2. One place is below the limit. No problem.

**ic-e02** D1. `balance` 60.00, `payment` 5.125. The checks read `payment` and are skipped.

1. `decimal-places` · `payment` · 5.125 — `payment` is 5.125. It can have at most 2 decimal places.

**ic-e03** D2. `price_per_kilo` 0.437. Three places is the limit. No problem.

**ic-e04** D2. `price_per_kilo` 0.4370. The text has 4 places because the zero is kept.

1. `decimal-places` · `price_per_kilo` · 0.4370 — `price_per_kilo` is 0.4370. It can have at most 3 decimal places.

**ic-e05** D10. `bags` 3. No problem.

**ic-e06** D10. `bags` 2.5. A limit of 0 has its own sentence.

1. `decimal-places` · `bags` · 2.5 — `bags` is 2.5. It cannot have decimal places.

**ic-e07** D10. `bags` 3.0. The zero after the dot is a place.

1. `decimal-places` · `bags` · 3.0 — `bags` is 3.0. It cannot have decimal places.

**ic-e08** D10. `weight` 2.25. A limit of 1 has its own sentence.

1. `decimal-places` · `weight` · 2.25 — `weight` is 2.25. It can have at most 1 decimal place.

**ic-e09** D5. `balances` [420.00, 380.505, 95.25, 510.00, 440.75, 300.00], `floor` 350.00. Check 1 reads `balances` and is skipped.

1. `decimal-places` · `balances` item 2 · 380.505 — Item 2 of `balances` is 380.505. It can have at most 2 decimal places.

## F. allowed-values

Values are compared by their characters. A different case is a different value.

**ic-f01** D3. `status` lapsed. No problem.

**ic-f02** D3. `status` paused.

1. `not-allowed` · `status` · paused — `status` is "paused". It must be one of "active", "lapsed", "suspended".

**ic-f03** D3. `status` Active.

1. `not-allowed` · `status` · Active — `status` is "Active". It must be one of "active", "lapsed", "suspended".

**ic-f04** D3. `status` the text `active` followed by one space.

1. `not-allowed` · `status` · active␣ — `status` is "active ". It must be one of "active", "lapsed", "suspended".

## G. A rule on each item of a list

The problem names the item by its position. The items are tested in position order, and for each item its type first, then `no-negatives`, `range`, `decimal-places` and `allowed-values`.

**ic-g01** D4. `scores` [72, 85, 64, 91, 77]. No problem.

**ic-g02** D4. `scores` [72, 104].

1. `too-large` · `scores` item 2 · 104 — Item 2 of `scores` is 104. It must be from 0 to 100.

**ic-g03** D4. `scores` [-3, 104, abc]. Three problems in item order. The third item cannot be read, so it gets `wrong-type` and nothing else.

1. `too-small` · `scores` item 1 · -3 — Item 1 of `scores` is -3. It must be from 0 to 100.
2. `too-large` · `scores` item 2 · 104 — Item 2 of `scores` is 104. It must be from 0 to 100.
3. `wrong-type` · `scores` item 3 · abc — Item 3 of `scores` is abc. It must be a whole number with at most 18 digits.

**ic-g04** D11. `amounts` [-5.125]. One item breaks three rules, in the order of the rule table.

1. `no-negatives` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It cannot be negative.
2. `too-small` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It must be from 0.00 to 100.00.
3. `decimal-places` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It can have at most 2 decimal places.

**ic-g05** D11. `amounts` [10.00, 250.00, 7.5]. Only item 2 breaks a rule. Item 3 has one place, which is below the limit.

1. `too-large` · `amounts` item 2 · 250.00 — Item 2 of `amounts` is 250.00. It must be from 0.00 to 100.00.

**ic-g06** D11. `amounts` with 1000 items, each -5.125. The most a list can hold is 1000. Each item gives the three problems of ic-g04, so the list holds 3000 problems and none is dropped. The test generates the row and compares item n against the three sentences of ic-g04 with the position n, in order: the problems 1 to 3 are item 1, and the problems 2998 to 3000 are item 1000.

1. `no-negatives` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It cannot be negative.
2. `too-small` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It must be from 0.00 to 100.00.
3. `decimal-places` · `amounts` item 1 · -5.125 — Item 1 of `amounts` is -5.125. It can have at most 2 decimal places.
2998. `no-negatives` · `amounts` item 1000 · -5.125 — Item 1000 of `amounts` is -5.125. It cannot be negative.
2999. `too-small` · `amounts` item 1000 · -5.125 — Item 1000 of `amounts` is -5.125. It must be from 0.00 to 100.00.
3000. `decimal-places` · `amounts` item 1000 · -5.125 — Item 1000 of `amounts` is -5.125. It can have at most 2 decimal places.

## H. Each check on its own

A check runs only when every input it reads had no problem. A skipped check gives no problem and is never invented as a failure. A check that runs gives `check-failed` when it is false. The sentence is the message, a full stop, then the values received of the inputs the check reads, in the order the names first appear in the formula, joined by `, ` and the word `and` before the last, and a full stop.

**ic-h01** D6. The clean row. Both checks run and are true: the interest is 201.60. No problem.

**ic-h02** D1. `balance` 120.50, `payment` 0.00. Both inputs are clean. Check 1 is false. Check 2 is true.

1. `check-failed` · check "payment > 0" — The payment must be more than zero. `payment` is 0.00.

**ic-h03** D1. `balance` 100.00, `payment` 2000.00. Check 1 is true. Check 2 is false, because 2000.00 is more than 1100.00. The check reads `payment` first and then `balance`.

1. `check-failed` · check "payment <= balance + 1000.00" — The payment cannot be more than 1000.00 above the balance. `payment` is 2000.00 and `balance` is 100.00.

**ic-h04** D6. `principal` 50.00, `monthly_rate` 0.050, `months` 60. Both checks run and both are false, so there are two problems, in the order the checks are declared. The first check reads one input and the second reads three.

1. `check-failed` · check "principal >= 100.00" — The loan must be at least 100.00. `principal` is 50.00.
2. `check-failed` · check "principal * monthly_rate * months <= principal" — The interest cannot be more than the amount borrowed. `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60.

**ic-h05** D6. `principal` 100.00, `monthly_rate` 0.050, `months` 20. Both checks are true on their limits: 100.00 is at least 100.00, and the interest of 100.00 is not more than 100.00. No problem.

**ic-h06** D5. `balances` [420.00, 380.50], `floor` 350.00. The check reads a list, which is shown as a list.

1. `check-failed` · check "count(balances) >= 6" — At least 6 balances are needed. `balances` is [420.00, 380.50].

**ic-h07** D7. `total` 10.00, `people` 0. Both inputs are clean, so the check runs and divides by zero. The problem is `check-error`, not the run code `divide-by-zero`. The formula is not calculated.

1. `check-error` · check "total / people >= 1.00" — The check "Each person must get at least 1.00" could not be worked out: The part `total / people` divides by zero.

**ic-h08** D7. The clean row. The check divides to 33.33333333333333333333, which is at least 1.00. No problem.

**ic-h09** D6. `principal` 50.00, `monthly_rate` 0.06, `months` 60. `monthly_rate` has a problem. Check 1 reads only `principal`, so it runs and is false. Check 2 reads `monthly_rate`, so it is skipped and gives nothing. The problems of inputs come before the problems of checks.

1. `too-large` · `monthly_rate` · 0.06 — `monthly_rate` is 0.06. It must be from 0.000 to 0.050.
2. `check-failed` · check "principal >= 100.00" — The loan must be at least 100.00. `principal` is 50.00.

**ic-h10** D6. `principal` 50.00, `monthly_rate` 0.050, `months` 60. This is the row of ic-h09 after `monthly_rate` was fixed. Check 2 now runs and gives a problem that the list of ic-h09 did not show.

1. `check-failed` · check "principal >= 100.00" — The loan must be at least 100.00. `principal` is 50.00.
2. `check-failed` · check "principal * monthly_rate * months <= principal" — The interest cannot be more than the amount borrowed. `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60.

**ic-h11** D6. `principal` -1.00, `monthly_rate` 0.012, `months` 7. `principal` has a problem and both checks read it, so both are skipped. Neither failure is invented, though check 1 would have been false.

1. `too-small` · `principal` · -1.00 — `principal` is -1.00. It must be more than 0.00.

**ic-h12** D6. `principal` 2400.00, `monthly_rate` 0.012, `months` not in the row. Check 1 reads only `principal` and is true. Check 2 reads `months` and is skipped.

1. `required` · `months` · not given — `months` was not given. It is required.

**ic-h13** D5. `balances` [420.00, abc], `floor` 350.00. An item has a problem, so the list has a problem and check 1 is skipped.

1. `wrong-type` · `balances` item 2 · abc — Item 2 of `balances` is abc. It must be a decimal number with at most 30 digits, such as 12.50.

**ic-h14** D6. `principal` 50.00, `monthly_rate` 0.012, `months` 7, and a name `memo` with the text hello that is not an input. An unknown name does not stop a check, and its problem comes before the checks. Check 1 is false. Check 2 is true.

1. `unknown-input` · `memo` · hello — `memo` is not an input of this function. Its inputs are: principal, monthly_rate, months.
2. `check-failed` · check "principal >= 100.00" — The loan must be at least 100.00. `principal` is 50.00.

## I. Several problems at once, and the order

Validate never stops at the first problem. The order is: the inputs in the order they are declared; for each input its rules in the order of the rule table; for a list the list rules once and then the items in position order; then the names in the row that are not inputs, in the order of their characters; then the checks.

**ic-i01** D2. `kilos` -3, `price_per_kilo` 0.0005, `credit` not in the row. Three problems with three different codes, in input order.

1. `no-negatives` · `kilos` · -3 — `kilos` is -3. It cannot be negative.
2. `decimal-places` · `price_per_kilo` · 0.0005 — `price_per_kilo` is 0.0005. It can have at most 3 decimal places.
3. `required` · `credit` · not given — `credit` was not given. It is required.

**ic-i02** D9. `deposits` [-1.00, 5.00], `expenses` [-2.00], `members` 0. The lists come in the order they are declared, and the integer after them.

1. `no-negatives` · `deposits` item 1 · -1.00 — Item 1 of `deposits` is -1.00. It cannot be negative.
2. `no-negatives` · `expenses` item 1 · -2.00 — Item 1 of `expenses` is -2.00. It cannot be negative.
3. `too-small` · `members` · 0 — `members` is 0. It must be from 1 to 500.

**ic-i03** D1. `balance` -1.00, `payment` 45.25, and the names `zeta` with 1 and `alpha` with 2, which are not inputs. The unknown names come after the inputs, in the order of their characters, whatever order the row held them in. Check 1 reads `payment` and is true. Check 2 reads `balance` and is skipped.

1. `no-negatives` · `balance` · -1.00 — `balance` is -1.00. It cannot be negative.
2. `unknown-input` · `alpha` · 2 — `alpha` is not an input of this function. Its inputs are: balance, payment.
3. `unknown-input` · `zeta` · 1 — `zeta` is not an input of this function. Its inputs are: balance, payment.

**ic-i04** D6. `principal` 50.00, `monthly_rate` 0.06, `months` not in the row, and the name `memo` with the text hello. All four kinds of problem in one answer: a rule, a missing input, an unknown name and a check. Check 2 reads `monthly_rate` and `months` and is skipped. Check 1 reads `principal` and is false.

1. `too-large` · `monthly_rate` · 0.06 — `monthly_rate` is 0.06. It must be from 0.000 to 0.050.
2. `required` · `months` · not given — `months` was not given. It is required.
3. `unknown-input` · `memo` · hello — `memo` is not an input of this function. Its inputs are: principal, monthly_rate, months.
4. `check-failed` · check "principal >= 100.00" — The loan must be at least 100.00. `principal` is 50.00.

**ic-i05** D6. The row holds no name. Every input is listed in order and both checks are skipped.

1. `required` · `principal` · not given — `principal` was not given. It is required.
2. `required` · `monthly_rate` · not given — `monthly_rate` was not given. It is required.
3. `required` · `months` · not given — `months` was not given. It is required.

**ic-i06** D2. `kilos` abc, `price_per_kilo` 0.0005, `credit` 120.00. An input that cannot be read does not stop the other inputs from being tested. `kilos` gets `wrong-type` and no `no-negatives`.

1. `wrong-type` · `kilos` · abc — `kilos` is abc. It must be a whole number with at most 18 digits.
2. `decimal-places` · `price_per_kilo` · 0.0005 — `price_per_kilo` is 0.0005. It can have at most 3 decimal places.

**ic-i07** D2. `kilos` -1.5, `price_per_kilo` 0.437, `credit` 120.00. The text cannot be read as a whole number, so it gets `wrong-type` only, though it is also below zero.

1. `wrong-type` · `kilos` · -1.5 — `kilos` is -1.5. It must be a whole number with at most 18 digits.

## J. Missing, empty and space-only are told apart

A name that is not in the row is not given. A name with an empty text, or a text of spaces, is given. For a text input that has no rule against it, an empty text and a text of spaces are accepted. For a number, a boolean or a date, they cannot be read and give `wrong-type`.

**ic-j01** D3. `status` not in the row.

1. `required` · `status` · not given — `status` was not given. It is required.

**ic-j02** D3. `status` the empty text. It is given, so `required` does not refuse it. The allowed values do.

1. `not-allowed` · `status` · (empty) — `status` is "". It must be one of "active", "lapsed", "suspended".

**ic-j03** D3. `status` three spaces. It is given and it is not an allowed value.

1. `not-allowed` · `status` · three spaces — `status` is "   ". It must be one of "active", "lapsed", "suspended".

**ic-j04** D1. `balance` 80.00, `payment` not in the row. The same as ic-a02. It is listed here to sit beside ic-j05 and ic-j06.

1. `required` · `payment` · not given — `payment` was not given. It is required.

**ic-j05** D1. `balance` 80.00, `payment` the empty text. The text is given and cannot be read as a decimal. The sentence follows the template word for word, so the value place is empty (open question 2). The checks read `payment` and are skipped.

1. `wrong-type` · `payment` · (empty) — `payment` is . It must be a decimal number with at most 30 digits, such as 12.50.

**ic-j06** D1. `balance` 80.00, `payment` one space. The sentence holds the space between the word `is` and the full stop, so there are two spaces before the full stop.

1. `wrong-type` · `payment` · one space — `payment` is  . It must be a decimal number with at most 30 digits, such as 12.50.

**ic-j07** D8. `active` the empty text.

1. `wrong-type` · `active` · (empty) — `active` is . It must be true or false.

**ic-j08** D8. `note` the empty text. A text may have 0 characters. No problem.

**ic-j09** D8. `note` three spaces. No problem.

**ic-j10** D3. `contributions` not in the row. See ic-b03.

1. `list-too-short` · `contributions` · not given — `contributions` was not given, so it has 0 items. It needs at least 1 item.

**ic-j11** D3. `contributions` [], a list that is given and empty. See ic-b02.

1. `list-too-short` · `contributions` · [] — `contributions` has 0 items. It needs at least 1 item.

**ic-j12** D3. `contributions` [""], a list of one item that is an empty text. The list has 1 item, so `min-items` is met, and the item cannot be read.

1. `wrong-type` · `contributions` item 1 · (empty) — Item 1 of `contributions` is . It must be a decimal number with at most 30 digits, such as 12.50.

## K. A number written as text

The library reads each input as the text it received. It does not convert `1,200.50`, `95%` or `1 200,50` into a number. It refuses them with `wrong-type`, and the sentence shows the text as received and an example of the shape it needs. Whether the sentence should also say what was wrong with the text is open question 3.

**Whole numbers** (D8 `kilos`). The sentence of a refused text is `` `kilos` is {text}. It must be a whole number with at most 18 digits. ``

| Text received | Result |
|---|---|
| 0 | accepted |
| 12 | accepted |
| -12 | accepted |
| 123456789012345678 (18 digits) | accepted |
| -0 | `wrong-type` — `kilos` is -0. It must be a whole number with at most 18 digits. |
| +5 | `wrong-type` — `kilos` is +5. It must be a whole number with at most 18 digits. |
| 007 | `wrong-type` — `kilos` is 007. It must be a whole number with at most 18 digits. |
| 1237.5 | `wrong-type` — `kilos` is 1237.5. It must be a whole number with at most 18 digits. |
| 1e3 | `wrong-type` — `kilos` is 1e3. It must be a whole number with at most 18 digits. |
| 1,200 | `wrong-type` — `kilos` is 1,200. It must be a whole number with at most 18 digits. |
| 1 200 | `wrong-type` — `kilos` is 1 200. It must be a whole number with at most 18 digits. |
| 1234567890123456789 (19 digits) | `wrong-type` — `kilos` is 1234567890123456789. It must be a whole number with at most 18 digits. |
| abc | `wrong-type` — `kilos` is abc. It must be a whole number with at most 18 digits. |

**Decimal numbers** (D8 `payment`). The sentence of a refused text is `` `payment` is {text}. It must be a decimal number with at most 30 digits, such as 12.50. ``

| Text received | Result |
|---|---|
| 0 | accepted |
| 12.50 | accepted |
| -12.5 | accepted |
| -0.00 | accepted |
| 0.5 | accepted |
| 123456789012345678901234567890 (30 digits) | accepted |
| .5 | `wrong-type` — `payment` is .5. It must be a decimal number with at most 30 digits, such as 12.50. |
| 5. | `wrong-type` — `payment` is 5.. It must be a decimal number with at most 30 digits, such as 12.50. |
| +5 | `wrong-type` — `payment` is +5. It must be a decimal number with at most 30 digits, such as 12.50. |
| 05.5 | `wrong-type` — `payment` is 05.5. It must be a decimal number with at most 30 digits, such as 12.50. |
| 5.5.5 | `wrong-type` — `payment` is 5.5.5. It must be a decimal number with at most 30 digits, such as 12.50. |
| 1e3 | `wrong-type` — `payment` is 1e3. It must be a decimal number with at most 30 digits, such as 12.50. |
| 1,5 | `wrong-type` — `payment` is 1,5. It must be a decimal number with at most 30 digits, such as 12.50. |
| 1,200.50 | `wrong-type` — `payment` is 1,200.50. It must be a decimal number with at most 30 digits, such as 12.50. |
| 95% | `wrong-type` — `payment` is 95%. It must be a decimal number with at most 30 digits, such as 12.50. |
| 1 200,50 | `wrong-type` — `payment` is 1 200,50. It must be a decimal number with at most 30 digits, such as 12.50. |
| 12.50 followed by a space | `wrong-type` — `payment` is 12.50␣. It must be a decimal number with at most 30 digits, such as 12.50. |
| 1234567890123456789012345678901 (31 digits) | `wrong-type` — `payment` is 1234567890123456789012345678901. It must be a decimal number with at most 30 digits, such as 12.50. |
| NaN | `wrong-type` — `payment` is NaN. It must be a decimal number with at most 30 digits, such as 12.50. |

**Approximate numbers** (D8 `ratio`). The text rules are those of a decimal and the words differ. The sentence of a refused text is `` `ratio` is {text}. It must be a number with at most 30 digits, such as 0.37. ``

| Text received | Result |
|---|---|
| 0.37 | accepted |
| 5 | accepted |
| -1.5 | accepted |
| .37 | `wrong-type` — `ratio` is .37. It must be a number with at most 30 digits, such as 0.37. |
| 0,37 | `wrong-type` — `ratio` is 0,37. It must be a number with at most 30 digits, such as 0.37. |
| 1e3 | `wrong-type` — `ratio` is 1e3. It must be a number with at most 30 digits, such as 0.37. |

Each refused row of the three tables is one case. A refused text gives exactly one problem, `wrong-type`, and gives no `no-negatives`, no range and no places problem for the same input. The case ids are `ic-k-int-1` and onward, `ic-k-dec-1` and onward and `ic-k-flt-1` and onward, in table order.

## L. The other types

Each refused row is a case on D8 with the one input changed. A refused text gives one problem, `wrong-type`, and nothing else.

**Text** (`note`). The sentence of a refused text is `` `note` is "{text}". It must be text of at most 200 characters. ``

| Text received | Result |
|---|---|
| ok | accepted |
| the empty text | accepted |
| 200 letters a | accepted |
| 201 letters a | `wrong-type` — the text is cut to its first 200 characters and then `...`. The sentence is `` `note` is "{200 letters a}...". It must be text of at most 200 characters. `` The place of the quotes is open question 4. |
| the text a, a tab, the text b | `wrong-type` — the sentence shows the text as received, with the tab: `` `note` is "a{tab}b". It must be text of at most 200 characters. `` |
| the text a, a line break, the text b | `wrong-type` — `` `note` is "a{line break}b". It must be text of at most 200 characters. `` |

**Yes or no** (`active`). The sentence of a refused text is `` `active` is {text}. It must be true or false. ``

| Text received | Result |
|---|---|
| true | accepted |
| false | accepted |
| True | `wrong-type` — `active` is True. It must be true or false. |
| TRUE | `wrong-type` — `active` is TRUE. It must be true or false. |
| yes | `wrong-type` — `active` is yes. It must be true or false. |
| 1 | `wrong-type` — `active` is 1. It must be true or false. |

**Date** (`filed`). The sentence of a refused text is `` `filed` is {text}. It must be a date written like 2026-03-31. ``

| Text received | Result |
|---|---|
| 2026-03-31 | accepted |
| 2028-02-29 | accepted (a leap year) |
| 2026-02-30 | `wrong-type` — `filed` is 2026-02-30. It must be a date written like 2026-03-31. |
| 2027-02-29 | `wrong-type` — `filed` is 2027-02-29. It must be a date written like 2026-03-31. |
| 2026-13-01 | `wrong-type` — `filed` is 2026-13-01. It must be a date written like 2026-03-31. |
| 2026-3-1 | `wrong-type` — `filed` is 2026-3-1. It must be a date written like 2026-03-31. |
| 31/03/2026 | `wrong-type` — `filed` is 31/03/2026. It must be a date written like 2026-03-31. |

The case ids are `ic-l-text-1` and onward, `ic-l-bool-1` and onward and `ic-l-date-1` and onward, in table order.

**A list where a single value belongs, and the reverse**

**ic-l-list-1** D1. `balance` 80.00, `payment` received as the list [45.25]. A list for an input that is not a list gets `wrong-type` and nothing else.

1. `wrong-type` · `payment` · [45.25] — `payment` is [45.25]. It must be a decimal number with at most 30 digits, such as 12.50.

**ic-l-list-2** D3. `contributions` received as the text 200.00. See ic-b12 for the same case on a list of integers.

1. `wrong-type` · `contributions` · 200.00 — `contributions` is 200.00. It must be a list of decimal numbers.

**ic-l-list-3** D3. `contributions` [200.00, 1,200.50, 100.00]. An item written with a comma is refused like any other text, and the other items are still tested. The row is written as three items.

1. `wrong-type` · `contributions` item 2 · 1,200.50 — Item 2 of `contributions` is 1,200.50. It must be a decimal number with at most 30 digits, such as 12.50.

## M. A row with problems is never calculated, and every caller gets the same list

**ic-m01** D1. `balance` 120.50, `payment` 45.25. No problem, so the formula is calculated. The result is `75.25`.

**ic-m02** D1. `balance` 80.00, `payment` -5.00. The list of ic-c02. The formula is not calculated: the answer holds the problem list and no result, and the engine is not asked to calculate.

**ic-m03** D1. `balance` 120.50, `payment` 0.00. The list of ic-h02. The only problem is a check, and the row is still not calculated.

**ic-m04** A batch of three rows on D1, in this order: row 1 `balance` 120.50 and `payment` 45.25; row 2 `balance` -1.00 and `payment` -5.125; row 3 `balance` 120.50 and `payment` 0.00. The problems of a row are that row's own and do not change the other rows.

| Row | Problems | Calculated |
|---|---|---|
| 1 | none | yes, the result is `75.25` |
| 2 | the three problems of ic-c04, in that order | no |
| 3 | the one problem of ic-h02 | no |

**ic-m05** The same row gives the same problem list through each way in. A way in is the single run (`Validate(definition, row)`), a row of a batch run and the call of an AI tool. The test takes these rows, runs each through the three ways in, and compares the lists field by field and in order: ic-c04, ic-g03, ic-h04 and ic-i04. The only difference allowed is what a way in adds around the list: the position of the row in a batch, and the wrapper of a tool answer. Neither is a field of a problem. What each wrapper holds is open question 6.

**ic-m06** The list is the same each time it is asked for. Each row of ic-m05 is validated twice in one process and once in a second process, and the three lists are equal.

## What the cases cover

| What the issue asks for | Cases |
|---|---|
| Each input rule on its own: missing, wrong type, below the range, above the range, negative where not allowed, too many decimal places, a value not in the allowed list | A (missing), K and L (wrong type), D (range), C (negatives), E (places), F (allowed values), B (list sizes) |
| Each check on its own, with its own message | H |
| Several problems at once, all of them, in declared order, never only the first | I, and the multi-problem cases ic-c04, ic-g03, ic-g04, ic-g06, ic-h04 |
| Each problem has the input's name, the value received, the rule broken, a plain sentence and a code; the sentences are checked word for word | Every case: the whole list is compared, sentence by sentence |
| A blank, a space-only value and a missing input are told apart | J |
| A number given as text is handled as the language page decides, or refused with a sentence | K |
| A row with problems is never calculated, and an input that fails a rule never reaches a check that depends on it | M, and ic-h09, ic-h11, ic-h12, ic-h13 |
| The same problem shape for a single run, a batch row and an AI tool call | ic-m04, ic-m05, ic-m06 |

## Open questions

The owner has not decided any of these. Where the language page is silent, the case above says what it reads and this list says what is unsettled.

1. **Negatives under decision 22.** Decision 22 says an input that does not declare whether it may be negative refuses negatives. The language page has the opposite shape: a `no-negatives` rule refuses them and an input without the rule accepts them. The cases read the page as written (ic-c07, ic-h11 and every input with a range but no `no-negatives`). If decision 22 is built as written, the inputs without a declaration refuse a negative value, an input that allows negatives needs a way to say so, and the cases on `price_per_kilo`, `requested`, `scores`, `principal` and `members` get a second problem or a changed list. The name of the declaration, and the sentence and code for the refusal, are not set on any page.
2. **How an empty value reads in a sentence.** The template `` `{input}` is {value}. `` gives `` `payment` is . `` for an empty text and a sentence with two spaces for a text of one space (ic-j05, ic-j06, ic-j07, ic-j12). A treasurer cannot read that. The proposal is to show an empty or space-only value between double quotes for every type, as a text already is.
3. **A hint when a number is written in a form that is refused.** `1,200.50`, `95%` and `1 200,50` get the ordinary `wrong-type` sentence, which shows the text and `such as 12.50`. It does not say that a comma, a percent sign or a space inside a number is the cause. The owner decides whether the sentence should say so, and whether the library should ever read such a form (the language page says it does not).
4. **A text of more than 200 characters.** The page says the first 200 characters are shown and then `...` (proposed). It does not say whether the `...` is inside the double quotes of the sentence.
5. **`allowed-values` on an integer.** The sentence template writes the value between double quotes (`` `{input}` is "{value}". ``), while the page says only a text is quoted. No case is written for an integer with allowed values until the sentence is settled.
6. **The wrapper of a batch row and of an AI tool answer.** No page gives the shape of a tool answer, and the batch row position is outside the problem. Items 18 and 19 of "Differences to settle" on the language page list the same gap. ic-m05 compares only the list.
7. **A list of more than 1000 items.** The page gives `list-too-long` for it and numbers positions from 1 to 1000. It does not say whether the items after the 1000th are tested. No case is written.
8. **A check that reads a list that was not given.** A list input without `required` and without `min-items` that is not in the row counts as empty and has no problem, so a check that reads it runs. The page does not say what value its sentence shows for it. No case is written.
9. **Whether Validate reports a skipped check.** The page says a skipped check gives nothing (proposed), so ic-h10 shows a problem that the first list did not. The other choice is a problem with the code `check-skipped`, which would change ic-h09, ic-h11, ic-h12 and ic-h13.
10. **The reading of the sentences.** The sentences are in English only on the language page. The owner reads them in the review that follows this page (TD-2982) before they are fixed.
