# Problem record: implementation

The `problem` package holds one Validate problem, the exact-text showing of a received value, the sentence of every code of Validate, and the JSON form. It builds one problem at a time. It does not test a row, order a list of problems or decide what stops what: that is Validate, which is built on top of this package.

The shape and the sentences are set by [formula-language.md](../2.%20design/formula-language.md), sections "The Validate problem", "The order of problems, and what stops what", "Sentences" and "Every code of Validate". The cases that prove them are on [input-checks.md](../4.%20qa/input-checks.md). The rules behind them are decision 8 (every caller gets the same list at once) and decision 18 (Validate covers only the values given at run time) on [decisions.md](../1.%20requirements/decisions.md).

Status on 2026-11-17: built and tested. The owner has not reviewed the sentences, so each sentence is as the language page and the QA page word it.

## The record

`Problem` has eight fields, always written in this order, with `null` where a field does not apply.

| Field | Go type | Meaning |
|---|---|---|
| `input` | `*string` | The name of the input. Null for a problem of a check. |
| `item` | `*int` | The position of the list item, counted from 1. Null when the problem is about the whole input. |
| `given` | `*bool` | True when the name was in the row, false when it was not. Null for a problem of a check. |
| `value` | `*string` | The value received as exact text. Null when the input was not given, and for a problem of a check. |
| `rule` | `string` | The rule broken. |
| `sentence` | `string` | One plain sentence naming the input in backticks and giving the value received. |
| `code` | `string` | A code a program can read. |
| `check` | `*string` | The standard text of the formula of a check. Null for every problem that is not of a check. |

`Problem.JSON()` gives one object and `ListJSON(problems)` gives an array. A nil or empty list is `[]`. The encoder does not escape `<`, `>` or `&`, so a check formula reads as it was written.

`AllCodes()` lists the twelve codes in the order of the language page's table. `ValidCode(code)` is true for 1 to 40 lowercase letters, digits and hyphens. The constants `Code…` and `Rule…` name every code and every rule.

## Showing a received value

A `Value` holds the text as it is shown. Build it with one of three constructors. `Field()` is what goes in the `value` field. `Sentence()` is what goes in a sentence.

| Constructor | Use it for | `Field()` | `Sentence()` |
|---|---|---|---|
| `PlainText(received)` | A number, a date, a yes or no, or text that gets no quotes | The text as received | The same |
| `QuotedText(received)` | The value of an input of type text | The text as received | The text between double quotes |
| `List(items)` | A list | `[a, b, c]` | The same |

- A text is shown character for character, with its spaces: `0.4370` stays `0.4370`, and `active ` keeps its trailing space.
- A text longer than 200 characters is shown as its first 200 characters followed by `...`. The cut counts characters, not bytes. The 200 is proposed on the language page.
- A list is shown between `[` and `]`, separated by `, `. An empty list is `[]`. After the 20th item the text goes on with `, ...` before `]`.
- A problem about one item of a list takes the text of that item as its `Value`, built the same way.

## Building a problem

A `Subject` names the input, and the item when the problem is about one item. `Subject{Input: "scores", Item: 3}` makes a sentence start with `Item 3 of` and the name. With `Item` 0 the sentence starts with the name alone.

| Builder | Code | Rule | Sentence |
|---|---|---|---|
| `Required(input)` | `required` | `required` | `` `payment` was not given. It is required. `` |
| `WrongType(subject, value, typeWords)` | `wrong-type` | `type` | `` `kilos` is abc. It must be a whole number with at most 18 digits. `` |
| `TooSmall(subject, value, range)` | `too-small` | `range` | `` `members` is 0. It must be from 1 to 500. `` |
| `TooLarge(subject, value, range)` | `too-large` | `range` | `` `bonus` is 500.01. It must be 500.00 or less. `` |
| `NoNegatives(subject, value)` | `no-negatives` | `no-negatives` | `` `payment` is -5.00. It cannot be negative. `` |
| `DecimalPlaces(subject, value, places)` | `decimal-places` | `decimal-places` | `` `payment` is 5.125. It can have at most 2 decimal places. `` |
| `NotAllowed(subject, value, allowed)` | `not-allowed` | `allowed-values` | `` `status` is "paused". It must be one of "active", "lapsed", "suspended". `` |
| `ListTooShort(input, given, value, count, items)` | `list-too-short` | `min-items` | `` `contributions` has 0 items. It needs at least 1 item. `` |
| `ListTooLong(input, value, count, items)` | `list-too-long` | `max-items` | `` `scores` has 13 items. It can have at most 12 items. `` |
| `UnknownInput(input, value, inputs)` | `unknown-input` | `known-input` | `` `memo` is not an input of this function. Its inputs are: principal, monthly_rate, months. `` |
| `CheckFailed(formula, message, values)` | `check-failed` | `check` | `` The payment must be more than zero. `payment` is 0.00. `` |
| `CheckError(formula, message, runSentence)` | `check-error` | `check` | `The check "Each person must get at least 1.00" could not be worked out: …` |

Notes on the sentences:

- `Range` holds `Min` and `Max` as exact text, and `MinExclusive` and `MaxExclusive`. `TooSmallText()` and `TooLargeText()` give the words: `from 1 to 500` when both bounds are given and neither is exclusive; otherwise `0.00 or more` or `more than 0.00` for a lower bound, and `500.00 or less` or `less than 100.00` for an upper bound.
- `DecimalPlaces` reads `It cannot have decimal places.` for 0, `It can have at most 1 decimal place.` for 1, and `at most n decimal places` for more.
- `ListTooShort` and `ListTooLong` read `item` for a count of 1 and `items` otherwise, for the count and for the limit. When the list was not given, `ListTooShort` reads `` `x` was not given, so it has 0 items. `` and the problem has `given` false and `value` null.
- `NotAllowed` writes each allowed value between double quotes, joined by `, `. Pass the value as `QuotedText`.
- `CheckFailed` writes the message, a full stop, then each value received, in the order given by the caller, joined by `, ` with `and` before the last: `` `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60. `` With no values it ends after the message. A check problem has `input`, `item`, `given` and `value` null and its formula in `check`.
- `CheckError` carries the sentence of the run problem it wraps. It is built by the caller, because that sentence is a code of a run and not of Validate.

## How it is tested

`go vet ./problem/ && go test ./problem/ -race` runs 91 cases: the showing of text and lists, the JSON field order and nulls, the code and rule of every builder, and the sentences taken word for word from the cases of the QA page.

A throwaway reader, kept outside the repo, read all 105 problem lines of the 104 cases on [input-checks.md](../4.%20qa/input-checks.md), rebuilt each sentence and code through the builders from the line's own fields, and compared them character for character. It found no disagreement. The cases that need a run (which problems a row gives, in what order, and that a row with problems is never calculated) are for Validate to prove.

## Open questions

These are not decided here. The package does what the pages say today.

1. Decision 22 says an input that does not declare negatives refuses them. The language page's `no-negatives` rule refuses them only when it is declared. This package only words the problem, so it does not decide it. Case `ic-c07` is untested.
2. An empty or space-only value reads badly in a sentence: `` `payment` is . ``. It is built as the QA page words it.
3. Whether the `...` of a cut text sits inside the quotes. The package puts it inside, so a cut quoted text reads `"first 200 characters..."`. The page has no case for it.
4. The shape of an AI tool answer is not defined on any page. Only the problem list is built.
5. The language page's table words the list sentences as `{items} items`, and the cases read `1 item` for a count of 1. The package follows the cases.
