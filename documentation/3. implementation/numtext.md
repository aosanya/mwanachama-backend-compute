# The value reader

The package `numtext/` reads each input of a row as the text it was received in and gives the value of the type the input declares, or says which rule the text broke. It reads the rules of "How a value is read" in [formula-language.md](../2.%20design/formula-language.md) and is tested against sections K and L of [input-checks.md](../4.%20qa/input-checks.md). It uses the standard library, `number/` for the exact decimal and the float, and `datefn/` for the date. It reads no clock and no locale, so the same text gives the same value on every machine.

## Using it

```go
value, refusals := numtext.Read(numtext.Decimal, numtext.One("45.20"))
```

| Name | What it is |
|---|---|
| `Read(kind, received)` | Reads one received input. Gives the value, or a list of refusals and an empty value. |
| `One(text)` | A received text. An empty text is a text that was given. |
| `Many(items...)` | A received list of texts. An empty list is a list that was given. |
| `Type` | `Integer`, `Decimal`, `Float`, `Text`, `Boolean`, `Date`, `IntegerList`, `DecimalList`. |
| `Value` | The type, the text as received, and the value in the field of its type: `Integer`, `Decimal`, `Float`, `Boolean`, `Date`, or `Items` for a list. |
| `Refusal` | The rule broken (`RuleType`, the text `type`), the type, the reason, the position of the item (from 1, 0 for the whole input) and the text as received. |

A program's number type is never read: the reader takes only text, so `5.125` and `45.20` are never turned into a binary number on the way in. A host that holds a number passes the digits as they were written.

A name that is not in the row is not given, and the reader is not called for it: the `required` rule belongs to the definition. An empty text is given, and a type that cannot be empty (a number, a boolean, a date) refuses it with the reason `empty`.

## What each type accepts and refuses

| Type | Accepted | Refused, with the reason |
|---|---|---|
| `Integer` | An optional `-`, then 1 to 18 digits, no leading zero unless the number is `0`. | `-0` `negative-zero`; `+5` `sign`; `007` `leading-zero`; 19 digits `too-many-digits`; `1237.5`, `1e3`, `1,200`, `1 200`, `abc`, a lone `-` `not-digits`; an empty text `empty`. |
| `Decimal` | An optional `-`, digits, and optionally a dot with 1 or more digits after it. At most 30 digits in all, the whole part counted. No leading zero unless the whole part is `0`. | `.5` `missing-whole-part`; `5.` `missing-fraction`; `+5` `sign`; `05.5` `leading-zero`; 31 digits `too-many-digits`; `1e3`, `1,5`, `1,200.50`, `95%`, `1 200,50`, `5.5.5`, `NaN`, a trailing space `not-digits`; an empty text `empty`. |
| `Float` | The same text as a decimal, read as the nearest approximate number after it passes. | The same as a decimal. |
| `Text` | 0 to 200 characters with no control character. A character is one code point. | More than 200 characters `too-long`; a tab, line break, bell or NUL `control-character`; bytes that are not valid text `invalid-text`. |
| `Boolean` | `true` or `false`. | `True`, `TRUE`, `yes`, `1` and a text with a space around it `not-boolean`; an empty text `empty`. |
| `Date` | `YYYY-MM-DD`, a day that exists, years 0001 to 9999 (the year range is `datefn/`'s and is proposed). | `31/03/2026`, `2026-3-1` `not-date`; `2026-02-30`, `2027-02-29`, `2026-13-01` `no-such-day`; an empty text `empty`. |
| `IntegerList`, `DecimalList` | A received list whose items each read as the element type. | A text received for a list input `text-for-list`. Each bad item is its own refusal with its position and the element type. |

A single-value type that receives a list is refused with `list-for-single`, and the text of the refusal is the list written `[45.25]`. A type the package does not know is refused with `unknown-type`.

Nothing is converted. `1,200.50`, `95%`, `1 200,50`, `1e3`, `+5`, `.5` and `5.` are refused for the three number types, never read as a number.

## Places are kept as written

The decimal keeps the digits after the dot as they were received, trailing zeros included: `45.20` has 2 places, `5.125` has 3, `0.4370` has 4, and `Decimal.String()` writes them back unchanged. The number of places an agency allows is a rule of the input (decision 15), applied after reading, not a rule of the reader. The value of the text is read as an exact decimal by the range, no-negatives and check rules.

## One refusal for a refused text

A refused text gives exactly one refusal and no other problem for the same input. A list gives one refusal for each refused item, in the order of the items, and the items that read are not reported. The reader does not write the sentence and does not choose the code: the problem record (ISS-41, `problem/`) turns the rule `type` into the code `wrong-type` and writes the sentence from the type and the text.

## What is open

| # | Question | What the package does now |
|---|---|---|
| 1 | A negative zero. The page accepts `-0.00` for a decimal. The decimal of `number/` keeps no sign for a zero, so it prints `0.00`. | The reader accepts it, keeps the text as received in `Value.Text` (which is what a problem shows), and the decimal is zero with 2 places. Whether `number/` should keep the sign is for ISS-23's owner. |
| 2 | `-0` for a decimal. The page refuses `-0` for a whole number only. | A decimal `-0` is accepted, as the table reads. |
| 3 | The reasons. The page names one code, `wrong-type`, for every refused text. The reasons here are the reader's own and are proposed: they let the problem record say what was wrong if the owner decides the sentence should (input-checks open question 3). | Every reason is listed above. None reaches a person. |
| 4 | What a control character is. | Any code point the standard library calls a control character, which includes the tab, line break, NUL, DEL and the C1 controls (proposed). |
| 5 | Text that is not valid UTF-8. The page does not say. | Refused with `invalid-text` (proposed). |
| 6 | The most items in a list. The language page gives a position from 1 to 1000 in a problem. | Not enforced here: `max-items` belongs to the definition. |
| 7 | The year range of a date. | `datefn/`'s: 0001 to 9999 (proposed). |
