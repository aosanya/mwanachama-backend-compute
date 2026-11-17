# The date functions

The package `datefn/` reads a calendar day and counts the days, whole months and whole years between two days. It uses the standard library only, reads no clock and knows no time zone, so the same two days give the same whole number on every machine. The rules come from the Built-in functions section of [formula-language.md](../2.%20design/formula-language.md).

## The date value

`Parse(text)` reads a day written `YYYY-MM-DD` and gives a `Date`, or `ErrNotADate`. The text must be exactly ten characters: four digits, a dash, two digits, a dash, two digits, with ASCII digits only. The day must exist.

| Text | Result |
|---|---|
| `2026-03-31` | read |
| `2028-02-29` | read: 2028 is a leap year |
| `2027-02-29`, `2100-02-29` | refused: not a leap year |
| `2026-02-30`, `2026-04-31` | refused: no such day |
| `2026-13-01`, `2026-00-10`, `2026-03-00` | refused |
| `31/03/2026`, `2026/03/31`, `2026-3-31`, `26-03-31` | refused: not the written shape |
| an empty text, a text with a space around it, `2026-03-31T00:00:00` | refused |
| `0000-01-01` | refused: the first year is 1 (proposed) |

Years run from 0001 to 9999 (proposed). A leap year is a year divisible by 4, except a year divisible by 100 that is not divisible by 400. `Date.String()` writes the same shape back, so `Parse(d.String())` gives `d`.

## The three functions

All three take two dates, `a` then `b`, and give a whole number. All three are negative when `b` is before `a`.

| Function | What it counts | Worked line |
|---|---|---|
| `DaysBetween(a, b)` | the days from `a` to `b` | 2026-03-31 to 2026-05-15 is 45 |
| `MonthsBetween(a, b)` | the whole months from `a` to `b` | 2026-01-20 to 2026-09-15 is 7 |
| `YearsBetween(a, b)` | `MonthsBetween(a, b)` divided by 12, dropping the rest | 2023-09-10 to 2026-11-17 is 3 |

## The month rule

A month is whole when the day of the month of `b` is not before the day of the month of `a`, or when `b` is the last day of its month. The count starts from the difference in months, `(year of b - year of a) * 12 + month of b - month of a`, and takes one off when the month is not whole.

| From | To | Months | Why |
|---|---|---|---|
| 2026-01-20 | 2026-09-15 | 7 | the difference is 8 and 15 is before 20 |
| 2026-01-31 | 2026-02-28 | 1 | 28 is before 31 but 28 February is the last day of its month |
| 2026-05-31 | 2026-06-30 | 1 | 30 June is the last day of its month |
| 2026-05-31 | 2026-06-29 | 0 | 29 is before 31 and 29 June is not the last day |
| 2028-02-29 | 2029-02-28 | 12 | 28 February 2029 is the last day of its month |
| 2028-02-29 | 2029-02-27 | 11 | 27 is before 29 and not the last day |
| 2025-12-15 | 2026-01-14 | 0 | the year boundary, one day short |
| 2025-12-15 | 2026-01-15 | 1 | the year boundary, the same day of the month |

A reversed pair counts the pair the right way round and gives the negative: from 2026-09-15 to 2026-01-20 is -7. A reversed pair is therefore always the negative of the pair as written, for days and for months. Years drop the rest toward zero, so -13 months is -1 year.

## What the tests hold

`datefn_test.go` holds the worked lines above, a leap day, the last day of a month, the same day, a reversed pair, a year boundary, the refused texts, that the years are the months divided by 12, that a reversed pair is the negative, and that the same input gives the same figure on every run. `qa_cases_test.go` runs the date lines of [formula-examples.md](../4.%20qa/formula-examples.md) and the cases of `date.late_days@1` in [library-reference.md](../4.%20qa/library-reference.md): all agree.

Run them with `go test ./datefn/ -race`.

## What stays open

Nothing in this package waits on the engine choice. Three points sit on other pages:

- A date input that is not given, and a date input that is the wrong shape, are told apart by Validate with the codes `required` and `wrong-type`. This package only refuses the text; the problem sentence belongs to the problem record.
- [library-reference.md](../4.%20qa/library-reference.md), open question 6, says a reversed pair and a leap-day anniversary have no stated result. The rules above give one, from the language page: a reversed pair is negative, and a leap day counts as a full year on the last day of February of the next common year.
- The first year (1) and the last (9999) are proposed here; the language page names no range.
