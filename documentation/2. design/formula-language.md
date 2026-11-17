# The formula language

Every example on this page has a row in `documentation/4. qa/formula-examples.md`, marked by the `case:` line under the example.

This page fixes the language that the body of an agency function is written in.
It states what a formula may contain, how it is written in its one standard text, how it is read as a tree, and what is refused and with which words.
The laws, the hostile cases, the batch run and the library reference are written from it.

Status: written on 2026-11-17 from the nineteen example formulas. The owner has not reviewed it. Every number marked "proposed" is a starting point, not a decision. All names and figures are invented.

## What this page does not decide

- The engine that runs a formula, and the time limit of a run.
- The batch run, its row limit and its result shape.
- The hash function. This page only says what the hash is taken of.
- What the libraries contain. This page only says how a library function is called and checked.
- How the studio draws the blocks. This page only says that every part of the language has one.

## What the language is for

The owner decided that a formula is stored and published as text written like a spreadsheet formula, for example `payment >= balance ? 0.00 : balance - payment`.
The studio shows the same formula as blocks that a builder can edit. Text is the only stored form.
So every part of the language below can be shown as one block and edited as one block, and the page leaves out every part that none of the nineteen example formulas needs.

## Values and types

| Type | Written in a formula | Meaning | Limits |
|---|---|---|---|
| integer | `3`, `90`, `0` | A whole number, exact. | At most 18 digits. No sign, no leading zero. |
| decimal | `12.50`, `0.010` | An exact decimal. A literal keeps its written digits, so `0.010` and `0.01` are two different literals of equal value. | At most 30 digits in all. At most 10 places in a literal. |
| float | no literal | An approximate 64-bit binary number. It comes only from an input or from `to_float`. | Always finite. |
| text | `"active"` | The characters between double quotes. `\"` and `\\` are the only escapes. | At most 200 characters. No control characters. |
| boolean | `true`, `false` | True or false. | None. |
| date | no literal | A calendar day such as 2026-03-31, with no time and no zone. It comes only from an input. | None. |
| list of integer, list of decimal | `[0.00, x]` | Items of one type, in order. | At most 1000 items (proposed). A list written in a formula has at least 1 item. |

A digit string has no sign. A negative number is written with a subtraction, such as `0 - x`.
A decimal literal has digits on both sides of the dot. `.5` and `5.` are refused.
An integer or decimal literal does not start with `0` unless it is `0` or starts `0.`.

Two words are reserved: `true` and `false`. Every other name is free, including `floor` and `on`.

### One type per list and per conditional

Two types have a common type when they are the same type, or when one is an integer and the other a decimal. The common type is then the decimal.
The common type is used in three places only:

- the items of a list written in a formula,
- the two answers of a conditional,
- an argument given to a library function or an agency function when the parameter is a decimal.

No example needs the widening of an integer item or answer. It is stated so the rule is one rule (proposed).

### Mixing types in an operator

| Left | Right | `+` `-` `*` | `/` | `<` `<=` `>` `>=` | `==` `!=` |
|---|---|---|---|---|---|
| integer | integer | integer | refused | boolean | boolean |
| integer | decimal, either order | decimal | decimal | boolean | boolean |
| decimal | decimal | decimal | decimal | boolean | boolean |
| integer | float, either order | float | float | boolean | refused |
| float | float | float | float | boolean | refused |
| decimal | float, either order | refused, `float-decimal-mix` | refused, `float-decimal-mix` | refused, `float-decimal-mix` | refused, `float-decimal-mix` |
| text | text | refused | refused | refused | boolean |
| any other pair | | refused, `type-mismatch` | refused, `type-mismatch` | refused, `type-mismatch` | refused, `type-mismatch` |

A refusal in the table without a code is `type-mismatch`.
To use a decimal and a float together, convert one with `to_float` or `to_decimal`.

### Exact arithmetic

- `+` and `-` of two decimals give as many places as the operand with more places: `120.50 - 45.25` is `75.25`.
- `*` adds the places: `0.437 * 853` is `372.761`. An integer has 0 places.
- `/` of decimals gives exactly 20 places, rounded half away from zero: `10.00 / 3` is `3.33333333333333333333`.
- Comparison reads values, not digits: `12.5 == 12.50` is true.
- `round(x, n)` rounds half away from zero to exactly `n` places: `round(2.345, 2)` is `2.35` and `round(-2.345, 2)` is `-2.35`. `n` is an integer written in the formula, from 0 to 10.
- `%` takes two integers. The result has the sign of the dividend and a size below the divisor's: `-7 % 3` is `-1` and `7 % -3` is `1`.
- Division by zero, by `/` or `%`, is a problem and never a value. A float is never infinite and never not-a-number.
- An integer result of more than 18 digits, a decimal result of more than 60 digits, or a float result that is not finite is the problem `number-too-big`.
- `sum` of an empty list is `0` and `count` of an empty list is `0`. `min`, `max` and `avg` of an empty list is the problem `empty-list`.
- A float converted to a decimal is read from its shortest text, so the float 1.005 becomes `1.01` at 2 places (proposed).

## Operators and precedence

From tightest to loosest:

| Level | Operators | Grouping | Operand types |
|---|---|---|---|
| 1 | `!` before its operand | prefix | boolean |
| 2 | `*` `/` `%` | left to right | numbers; `%` integers only |
| 3 | `+` `-` | left to right | numbers |
| 4 | `<` `<=` `>` `>=` | left to right | numbers |
| 5 | `==` `!=` | left to right | numbers or two texts |
| 6 | `&&` | left to right | boolean |
| 7 | `\|\|` | left to right | boolean |
| 8 | `condition ? then : else` | else branch to the right | see below |

Brackets `( )` group. A call and a list are written like a name and need no brackets.
Left to right means `a - b - c` is `(a - b) - c`.
Comparison levels chain by the same rule, so `a < b < c` reads as `(a < b) < c` and is refused as a type mismatch.

The conditional:

- It is right-associative in the else branch: `a ? b : c ? d : e` is `a ? b : (c ? d : e)`.
- The then branch is a whole expression, so `a ? b ? c : d : e` reads as `a ? (b ? c : d) : e`.
- The condition must be a boolean. Otherwise the problem is `condition-not-boolean`.
- The two answers must have a common type. Otherwise the problem is `branches-differ`.

Text is compared with `==` and `!=` only. Ordering of text is left out.

`&&` and `||` evaluate both sides, and a conditional evaluates both answers. Nothing is skipped.
A problem in any part is a problem of the formula, even in a part the condition did not choose.
No example needs a part to be skipped (proposed).
A run stops at the first problem it meets, reading the tree left to right with children before their parent.

## The tree

Parse turns text into a tree. Format turns a tree into the one standard text. The tree is the stable, documented form that the studio, the tests and the checks share.
A tree has no position data, no spacing data and no bracket data. Brackets only change the shape of the tree. So two formulas that mean the same, however they were typed, give equal trees.

A tree is written as JSON. Each part is an object with a `kind`.

| Kind | Fields | Children, in index order |
|---|---|---|
| literal | `type` (`integer`, `decimal`, `text` or `boolean`), `value` (string) | none |
| name | `name` (string) | none |
| unary | `op` (`!`), `operand` | 0 operand |
| binary | `op` (one of `+ - * / % < <= > >= == != && \|\|`), `left`, `right` | 0 left, 1 right |
| conditional | `condition`, `then`, `else` | 0 condition, 1 then, 2 else |
| call | `library` (string or null), `name` (string), `version` (integer or null), `args` (1 to 8 parts) | one per argument, in order |
| list | `items` (1 or more parts) | one per item, in order |

Rules for the fields:

- A literal keeps its type and its written digits. `0.00` is the decimal literal with the value `"0.00"`, never the number 0. `0.010` stays `"0.010"`.
- The value of a text literal is the characters between the quotes after the escapes are read.
- A literal has no sign.
- A name matches the pattern in the standard shape below.
- A call to a built-in has `library` null and `version` null. A call to an agency function has `library` null and a version. A call to a library function has a library name and a version. A call with no version has `version` null; Check refuses it.
- Fields are written in the order of the table. Equality is equality of values.

### Example tree

The formula `payment >= balance ? 0.00 : balance - payment` is this tree.

```json
{
  "kind": "conditional",
  "condition": {
    "kind": "binary",
    "op": ">=",
    "left": {"kind": "name", "name": "payment"},
    "right": {"kind": "name", "name": "balance"}
  },
  "then": {"kind": "literal", "type": "decimal", "value": "0.00"},
  "else": {
    "kind": "binary",
    "op": "-",
    "left": {"kind": "name", "name": "balance"},
    "right": {"kind": "name", "name": "payment"}
  }
}
```

It has 8 nodes and depth 3.

### Addressing a part

A part is addressed by its path: the list of child indexes from the root. The root has the path `[]`.
In the tree above, `[0]` is `payment >= balance`, `[2]` is `balance - payment` and `[2, 0]` is the name `balance` inside it.

A part is shown to a person as its path and its standard text. The standard text is the Format of that part alone, so the part `[2]` is shown as `balance - payment`.
Every problem names its part in both ways. A block in the studio can use the path to mark the block at fault.

## The standard shape

Saving stores `Format(Parse(text))`, never the text as typed. Hand-typed spacing is rewritten, so two people who write the same formula get identical text. The publication's content hash is taken of the standard text, so a respaced formula is not read as a change.

Format follows these rules:

1. The text is one line, with no line break, no tab, and no space at the start or the end.
2. A binary operator has one space on each side. `?` and `:` have one space on each side.
3. There is no space after `!`, after `(` or `[`, before `)` or `]`, or before `,`.
4. Arguments and list items are separated by `, `: a comma and one space.
5. A call is written `name(args)`, `name@1(args)` or `library.name@1(args)`. There is no space before the bracket and none around `.` or `@`. A call with no version in its tree is written without `@`.
6. A name is written as it is. A name has lowercase letters, digits and underscores only, starts with a letter, and has 1 to 40 characters. A library name follows the same pattern.
7. A literal is written as it is: integer digits, decimal digits with their trailing zeros, `true` or `false`, and text between double quotes with `\"` and `\\` as the only escapes. Spaces inside text are kept.
8. Brackets are the fewest that keep the tree. A left operand gets brackets only when its level is looser than the operator's. A right operand gets brackets when its level is looser than or equal to the operator's. So `a - (b - c)` keeps its brackets, `(a - b) - c` loses them, and `a || (b && c)` loses them.
9. The operand of `!` gets brackets when it is a binary operation or a conditional. `!!a` and `!(a && b)` are standard.
10. In a conditional, an else branch that is a conditional gets no brackets. A conditional in the condition or in the then branch gets brackets. A conditional that is an operand of an operator gets brackets. A conditional that is a whole formula, a call argument or a list item gets none.
11. Brackets that change nothing are dropped: `((a))` is `a`.

Parse accepts any space, tab, line feed or carriage return between two tokens. It refuses a space inside a number, a name or a call name, so `late_days @ 1(due, filed)` and `fin . pmt@1(x)` are refused.
A space is allowed between a call name and its bracket: `round (x, 2)` is read and saved as `round(x, 2)`.

### Eight spellings and their standard text

| Typed | Standard text |
|---|---|
| `payment>=balance?0.00:balance-payment` | `payment >= balance ? 0.00 : balance - payment` |
| `payment  >=  balance ? 0.00 :  balance - payment` | `payment >= balance ? 0.00 : balance - payment` |
| `max( [0.00 ,round( kilos*price_per_kilo , 2 )-credit] )` | `max([0.00, round(kilos * price_per_kilo, 2) - credit])` |
| `round (amount*0.10,2)` | `round(amount * 0.10, 2)` |
| `(balance - payment) - credit` | `balance - payment - credit` |
| `balance - ( payment - credit )` | `balance - (payment - credit)` |
| `! suspended&&role!="guest"` | `!suspended && role != "guest"` |
| `suspended?paid_up?1:2:3` | `suspended ? (paid_up ? 1 : 2) : 3` |

Digits are never rewritten. `12.5` and `12.50` are two different formulas with two different hashes, because they give two different results when a result is shown.

### The laws

- `Parse(Format(tree))` equals `tree`, for every tree that passes Check.
- `Format(Parse(Format(x)))` equals `Format(x)`, for every text `x` that Parse accepts.
- Saving stores `Format(Parse(text))`. When Parse or Check gives a problem, nothing is stored.
- Two spellings of one formula have one standard text, so they have one hash.

## Calls

| Written | What it calls | Version |
|---|---|---|
| `round(x, 2)` | A built-in function. | None. |
| `fin.pmt@2(a, b, c)` | A function of a library. | Required. |
| `late_days@1(due, filed)` | A function of an agency. | Required. |

The rules for a call:

- A library call or an agency call names a version, a whole number from 1 to 999999. A call without one is refused when the formula is saved: `fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2.`
- Arguments are positional. There are 1 to 8 of them. Their count and types are checked against the callee's signature. An integer is accepted for a decimal parameter.
- A function that calls another may call only a version that is already saved and published. So a function can never call itself, and there are no cycles.
- A chain of calls is at most 4 calls long. The chain counts from this formula to the last function that calls no other. A formula that calls a function which calls three more has a chain of 4.
- Every call checks the callee's own input rules. Those rules are specified under Input rules, checks and Validate.
- A call costs the callee's cost, as defined under Limits.

### Built-in functions

Built-ins have no version in the formula text.

| Call | Arguments | Result | Worked line |
|---|---|---|---|
| `round(x, n)` | `x` decimal; `n` an integer written in the formula, 0 to 10 | decimal with exactly `n` places | `round(372.761, 2)` is `372.76` |
| `round(x, n)` | `x` float; `n` as above | float | `round(x, 3)` of the float 1.3629 is `1.363` |
| `sum(list)` | list of integer, or list of decimal | integer, or decimal | `sum([200.00, 250.00])` is `450.00` |
| `count(list)` | list of integer, or list of decimal | integer | `count([420.00, 380.50])` is `2` |
| `avg(list)` | list of decimal | decimal, 20 places | `avg([1.00, 2.00])` is `1.50000000000000000000` |
| `avg(list)` | list of integer | float | `avg([72, 85])` is `78.5` |
| `min(list)`, `max(list)` | list of integer, or list of decimal | the item type | `min([0.00, 27.24])` is `0.00` |
| `to_float(x)` | decimal | float | `to_float(1.50)` is `1.5` |
| `to_decimal(x, n)` | `x` float; `n` an integer written in the formula, 0 to 10 | decimal with exactly `n` places | `to_decimal(0.37, 2)` of the float 0.37 is `0.37` |
| `days_between(a, b)` | date, date | integer | 2026-03-31 to 2026-05-15 is `45` |
| `months_between(a, b)` | date, date | integer | 2026-01-20 to 2026-09-15 is `7` |
| `years_between(a, b)` | date, date | integer | 2023-09-10 to 2026-11-17 is `3` |

`min` and `max` give the first item with the least or the greatest value.

How the date functions count:

- `days_between(a, b)` is the number of days from `a` to `b`. It is negative when `b` is before `a`.
- `months_between(a, b)` is the number of whole months from `a` to `b`. A month is whole when the day of the month of `b` is not before the day of the month of `a`, or when `b` is the last day of its month. So 2026-01-31 to 2026-02-28 is 1 month. It is negative when `b` is before `a`.
- `years_between(a, b)` is `months_between(a, b)` divided by 12, dropping the rest.

### The language version

The built-ins are not versioned in the formula text. They belong to the language version, a whole number that starts at 1.
Proposed: the publication records the language version once, the content hash covers it, and a change to what a built-in gives ships as language version 2. A publication keeps the language version it was published with until its builder moves, in the same way as for a library function. See the open questions.

## Limits

Every limit is a number. Parse stops at the first limit it meets and gives one problem. Check tests a tree that came from anywhere, such as the studio's blocks, so it tests the same limits again, and the ones Parse cannot know.

| Limit | Value | Tested by | Code |
|---|---|---|---|
| Formula text length, in characters | 500 | Parse on the typed text, Check on the standard text | `too-long` |
| Tree depth, the root being 1 | 12 | Parse, Check | `too-deep` |
| Brackets open at one time | 12 | Parse | `too-deep` |
| Node count | 100 | Parse, Check | `too-many-nodes` |
| Distinct names a formula reads, which are its inputs | 8 | Check | `too-many-inputs` |
| Arguments in one call | 8 | Check | `wrong-argument-count` |
| Items in a list written in a formula | 99, held by the node limit | Parse, Check | `too-many-nodes` |
| Items in a list value | 1000 (proposed) | Validate, under Input rules, checks and Validate | `list-too-long` |
| Text literal, in characters | 200 | Parse, Check | `too-long` |
| Integer literal, in digits | 18 | Parse, Check | `too-many-digits` |
| Decimal literal, in digits | 30 | Parse, Check | `too-many-digits` |
| Places of a decimal literal | 10 | Parse, Check | `too-many-places` |
| Name length, in characters | 40 | Parse, Check | `bad-syntax`, `bad-tree` |
| Version number | 1 to 999999 | Parse, Check | `bad-syntax`, `bad-tree` |
| Call depth, in calls | 4 | Check | `call-too-deep` |
| Cost | 20000 | Check | `too-costly` |
| Problems listed by one Check | 20 (proposed) | Check | none |
| Digits of an integer result | 18 | a run | `number-too-big` |
| Digits of a decimal result | 60 | a run | `number-too-big` |

The sentence of a limit refusal is the one the hostile formulas page uses: `This formula goes over the {limit} limit of {value}.`
The limit names are `formula length`, `depth`, `node`, `input`, `text length`, `number magnitude`, `decimal places`, `call depth` and `cost`.

### Cost

Cost is counted from the tree, before any run.

- Each node costs 1.
- `sum`, `count`, `avg`, `min` and `max` add the number of items of their list. For a list written in the formula that is its item count. For any other list it is 1000, until a `max-items` rule of the function's definition declares a smaller size.
- A call to an agency function or a library function adds the callee's cost. The cost of an agency function is counted when it is saved. A library declares the cost of each of its functions.

A run never costs more than its count, because the count already assumes the largest lists. So a run has no cost problem.

### The margin of the examples

| Formula | Depth | Nodes | Cost | Text length | Inputs |
|---|---|---|---|---|---|
| 6, late-filing penalty | 6 | 34 | 58 | 175 | 3 |
| 18, voting right | 6 | 18 | 18 | 98 | 5 |
| 12, depreciation | 8 | 16 | 18 | 90 | 5 |
| 10, term grade | 6 | 19 | 3019 | 81 | 1 |
| Limit | 12 | 100 | 20000 | 500 | 8 |

Over all nineteen formulas: the deepest tree is 8 (formulas 12 and 14), the most nodes is 34 (formula 6), the highest cost is 3019 (formula 10), the longest standard text is 175 characters (formula 6), the most inputs is 5 (formulas 12 and 18) and the longest call chain is 1 (formulas 6 and 8).
The limits are 1.5 times, 2.9 times, 6.6 times, 2.9 times, 1.6 times and 4 times those.

Formula 6 costs 58: its 34 nodes plus 3 calls of `late_days@1`, each adding the callee's cost of 8.

## Check

Check tests a tree. It does not run it and it does not see any input row.
It takes the names and types of the function's inputs and the library of callable functions, and gives a list of problems. An empty list means the tree is fine.

Check works in stages and stops after the first stage that gives a problem:

1. Shape: every part has the fields of its kind and nothing else, and every value fits its pattern. A part that does not is `bad-tree`.
2. Size: depth, node count, distinct inputs and standard text length.
3. Names, functions and types, in tree order, children before their parent and left to right.
4. Call depth, then cost.

Inside stage 3, a part that has a problem has no type, so the parts around it add no problem of their own. A formula with one wrong name gives one problem, not five.
Check gives at most 20 problems (proposed).

### The problem

A problem is one object.

```json
{
  "code": "unknown-name",
  "sentence": "The name `rate` is not an input of this function. Its inputs are: balance, months.",
  "part": "rate",
  "path": [1]
}
```

| Field | Meaning |
|---|---|
| `code` | Lowercase letters, digits and hyphens, 1 to 40 characters. The codes on this page are the only codes. |
| `sentence` | One plain sentence a treasurer can read. It names the part at fault by its text. |
| `part` | The standard text of the part at fault. For a limit on the whole formula it is the whole formula. |
| `path` | The path of the part at fault, as defined under the tree. |

A problem from Parse has no tree to point into. Its `part` is empty, its `path` is null, and it carries two more fields, `line` and `column`, both counted from 1. A column counts characters. The position is the first character that cannot continue the text, or one past the last character when the text ends too soon. For an unclosed text it is the opening quote, and for a broken escape it is the backslash.

### Every code

| Code | Where | When | Sentence |
|---|---|---|---|
| `bad-syntax` | Parse | The text cannot be read. | `The formula cannot be read: {what} at line {n}, column {m}.` |
| `too-long` | Parse, Check | Formula text over 500 characters, or a text literal over 200. | `This formula goes over the formula length limit of 500.` or `This formula goes over the text length limit of 200.` |
| `too-deep` | Parse, Check | Tree depth over 12, or more than 12 brackets open. | `This formula goes over the depth limit of 12.` |
| `too-many-nodes` | Parse, Check | More than 100 nodes. | `This formula goes over the node limit of 100.` |
| `too-many-digits` | Parse, Check | An integer literal over 18 digits, or a decimal literal over 30. | `This formula goes over the number magnitude limit of {18 or 30}.` |
| `too-many-places` | Parse, Check | A decimal literal with more than 10 places. | `This formula goes over the decimal places limit of 10.` |
| `bad-tree` | Check | A part has the wrong shape. | `The part at path {path} is not a valid part: {what}.` |
| `too-many-inputs` | Check | More than 8 distinct names. | `This formula goes over the input limit of 8.` |
| `unknown-name` | Check | A name is not an input of the function. | `` The name `{name}` is not an input of this function. Its inputs are: {names}. `` |
| `unknown-function` | Check | No built-in, library function or agency function has that name. | `The function {name} is not known.` |
| `version-missing` | Check | A library or agency call has no version. | `{name} needs a version. Available: {list}.` |
| `version-not-found` | Check | The name exists but not that version. | `` `{call}` does not exist. Latest is `{latest}`. `` |
| `version-not-allowed` | Check | A built-in is given a version. | `` `{name}` is built in and takes no version. `` |
| `wrong-argument-count` | Check | The number of arguments is not the signature's. | `` `{name}` needs {n} arguments and this call has {m}. `` |
| `wrong-argument-type` | Check | An argument has a type the signature does not take. | `` Argument {i} of `{name}` must be {needed}, and `{argument}` is {actual}. `` |
| `type-mismatch` | Check | An operator gets types it does not take. | `` The part `{part}` cannot be worked out: `{op}` needs {needs}, and this has {types}. `` |
| `condition-not-boolean` | Check | A condition is not a boolean. | `` The condition `{part}` must be true or false, and it is {type}. `` |
| `branches-differ` | Check | The two answers have no common type. | `` The two answers of `{part}` must be the same kind, and they are {a} and {b}. `` |
| `float-decimal-mix` | Check | A decimal and a float meet in an operator. | `` The part `{part}` mixes a decimal with a float; convert one with to_float or to_decimal. `` |
| `list-items-differ` | Check | The items of a list have no common type. | `` The items of `{part}` must all be the same kind, and this list has {a} and {b}. `` |
| `call-too-deep` | Check | A chain of more than 4 calls. | `This formula goes over the call depth limit of 4.` |
| `too-costly` | Check | Cost over 20000. | `This formula goes over the cost limit of 20000.` |
| `divide-by-zero` | A run | `/` or `%` by zero. | `` The part `{part}` divides by zero. `` |
| `number-too-big` | A run | A result over its digit limit, or a float that is not finite. | `` The part `{part}` gives a number with more than {18 or 60} digits. `` |
| `empty-list` | A run | `min`, `max` or `avg` of an empty list. | `` The part `{part}` needs a list with at least one item, and the list is empty. `` |

The codes of Validate and of a refused definition are listed under Input rules, checks and Validate.

The words `{what}` of `bad-syntax` come from this set: `the formula is empty`, `text is not valid UTF-8`, `unexpected character {c}` (a character outside the language is written `U+` and its code, such as `U+200B`), `unexpected control character`, `missing closing bracket`, `a : is missing`, `the formula ends after an operator`, `text is not closed`, `text ends inside an escape`, `text has an unknown escape`, `a number cannot have a leading zero`, `a version is a whole number from 1 to 999999`, `a name must be followed by brackets to be a call`, `two parts are side by side with no operator between them`.

The words `{what}` of `bad-tree` come from this set: `the kind {kind} is not in the language`, `the field {field} is missing`, `the field {field} is not allowed`, `the operator {op} is not in the language`, `the value {value} is not a {type}`, `the name {name} is not a valid name`, `the part has {n} children and needs {m}`, `the text is not a tree`.

The `{needs}` of `type-mismatch` is: `two numbers` for `+ - *` and the four comparisons, `a decimal or a float on at least one side` for `/`, `two integers` for `%`, `two numbers or two texts` for `==` and `!=`, `two booleans` for `&&` and `||`, `a boolean` for `!`.

A call to a name that is none of these, such as `import`, `eval`, `now`, `random`, `env`, `exec` or `read_file`, is `unknown-function`. A formula can only compute from its own text and its inputs.

## Input rules, checks and Validate

The owner decided that the rules for the inputs of a function and the checks over them are part of the function's definition, not of the step that calls it. So every caller, whether a step in a workflow, a bulk run or an AI chat tool, is held to the same rules.
Validate runs before any calculation and lists every problem at once, never only the first. Each problem names the input, the value received and the rule broken, as a plain sentence and a code a program can read.
Everything below that goes beyond those decisions is proposed.

### What a definition holds

A definition holds the function's inputs and its checks. It is stored with the function's version, so a changed rule ships as a new version, as a changed formula does (proposed).

| Field | Meaning | Limits |
|---|---|---|
| `inputs` | The inputs, in declaration order. | 0 to 8 inputs. |
| `inputs[].name` | The name the formula reads. The pattern of a name in the standard shape. | 1 to 40 characters, unique in the definition. |
| `inputs[].type` | One of `integer`, `decimal`, `float`, `text`, `boolean`, `date`, `list of integer`, `list of decimal`. | One type. |
| `inputs[].rules` | The rules of the input, each an object with a `rule` field and its parameters. | 0 to 8 rules, each rule name at most once. |
| `checks` | The checks, in declaration order. | 0 to 8 checks. |
| `checks[].formula` | A boolean formula in the language of this page, saved in its standard text. It reads inputs of this definition only. It obeys every limit under Limits, so at most 500 characters, 100 nodes, depth 12 and cost 20000. | See Limits. |
| `checks[].message` | The plain words a person reads when the check fails, written without a final full stop. | 1 to 120 characters, no line break. |

The worked definition is formula 1, balance after payment. Both inputs are required, may not be negative and have at most 2 decimal places. The two checks are invented for this example.

```json
{
  "inputs": [
    {
      "name": "balance",
      "type": "decimal",
      "rules": [
        {"rule": "required"},
        {"rule": "no-negatives"},
        {"rule": "decimal-places", "places": 2}
      ]
    },
    {
      "name": "payment",
      "type": "decimal",
      "rules": [
        {"rule": "required"},
        {"rule": "no-negatives"},
        {"rule": "decimal-places", "places": 2}
      ]
    }
  ],
  "checks": [
    {"formula": "payment > 0", "message": "The payment must be more than zero"},
    {"formula": "payment <= balance + 1000.00", "message": "The payment cannot be more than 1000.00 above the balance"}
  ]
}
```

### The rules

Six rules are decided: required, type, allowed range, no negatives, number of decimal places and allowed values. The list rules at least one element and at most n elements are the two list rules the example formulas use. A rule on a list input that is written once for the items, such as a range, is applied to every item.

| Order | Rule | Applies to | Parameters | Limits | Code |
|---|---|---|---|---|---|
| 1 | `required` | every type | none | none | `required` |
| 2 | `type` | every type. It is the `type` field of the input, not a rule object. | none | see "How a value is read" | `wrong-type` |
| 3 | `min-items` | list of integer, list of decimal | `items`, a whole number | 1 to 1000 | `list-too-short` |
| 4 | `max-items` | list of integer, list of decimal | `items`, a whole number | 1 to 1000. A list with no `max-items` rule is held to 1000. | `list-too-long` |
| 5 | `no-negatives` | integer, decimal, float, list of integer, list of decimal | none | none | `no-negatives` |
| 6 | `range` | the same types as `no-negatives` | `min` and `max`, each decimal text; `min_exclusive` and `max_exclusive`, each true or false, false when left out | at least one of `min` and `max`; a bound has at most 30 digits and at most 10 places | `too-small`, `too-large` |
| 7 | `decimal-places` | decimal, list of decimal | `places`, a whole number | 0 to 10 | `decimal-places` |
| 8 | `allowed-values` | text, integer | `values`, a list of texts | 1 to 20 values, each at most 200 characters, all different | `not-allowed` |

Notes on the rules:

- `required` on a list input means the list must be present. An empty list that is present is given. A list input without `required` that is not present counts as an empty list.
- Every input that is not a list must carry `required`. A definition that leaves it out is refused on save with `input-needs-required`, because the formula cannot be calculated without a value (proposed; see the open questions).
- `min-items` and `max-items` count the items of the list. The count is tested even when some items have the wrong type.
- `no-negatives`, `range`, `decimal-places` and `allowed-values` on a list input are applied to each item. The problem names the item by its position, counted from 1.
- `range` compares by value, as exact decimals, never as binary floats. `12.5` and `12.50` are equal. With `min_exclusive` true the bound itself is refused, which reads `more than 0.00`. A date has no range rule (proposed).
- `decimal-places` means at most `places` places, counted in the text received. A value with fewer places is accepted: `45.2` passes 2 places. `0.4370` has 4 places and fails 3 places, because the digits received are kept (proposed).
- `allowed-values` compares the characters exactly, and a different case is a different value: `"Active"` is not `"active"`.
- A value is negative when it is below zero. `-0.00` equals zero and is not negative.

What each rule says, with one valid and one invalid invented row. The rows are formulas named in the examples below.

| Rule | Sentence | Valid | Invalid |
|---|---|---|---|
| `required` | `` `payment` was not given. It is required. `` | payment 45.25 (`rules-required-valid`) | payment not given (`rules-required-missing`) |
| `type` | `` `kilos` is 1237.5. It must be a whole number with at most 18 digits. `` | kilos 1237 (`rules-type-valid`) | kilos 1237.5 (`rules-type-invalid`) |
| `min-items` | `` `yields` has 0 items. It needs at least 1 item. `` | 3 yields (`rules-list-at-least-one-valid`) | no yields (`rules-list-at-least-one-invalid`) |
| `max-items` | `` `scores` has 13 items. It can have at most 12 items. `` | 12 scores (`rules-list-at-most-valid`) | 13 scores (`rules-list-too-long`) |
| `no-negatives` | `` `payment` is -5.00. It cannot be negative. `` | balance 0.00 (`rules-no-negatives-valid`) | payment -5.00 (`rules-no-negatives-invalid`) |
| `range` | `` `members` is 0. It must be from 1 to 500. `` | members 7 and 500 (`rules-range-valid`, `rules-range-edge-valid`) | members 0 and 501 (`rules-range-too-small`, `rules-range-too-large`) |
| `decimal-places` | `` `payment` is 5.125. It can have at most 2 decimal places. `` | payment 45.2 (`rules-decimal-places-valid`) | payment 5.125 (`rules-decimal-places-invalid`) |
| `allowed-values` | `` `status` is "paused". It must be one of "active", "lapsed", "suspended". `` | status "lapsed" (`rules-allowed-values-valid`) | status "paused" (`rules-allowed-values-invalid`) |
| a rule on each item | `` Item 2 of `scores` is 104. It must be from 0 to 100. `` | 5 scores in range (`rules-each-item-valid`) | an item of 104 (`rules-each-item-too-large`) |
| a check | `` The payment must be more than zero. `payment` is 0.00. `` | payment 2400.00 over 7 months (`rules-check-valid`) | payment 0.00 (`rules-check-failed`) |

### How a value is read

A row gives each input as received text. The library never reads a number from a program's number type, because a binary number cannot keep `5.125` or `45.20` as written. A host that holds a number passes the digits as they were written.

| What the row holds | Meaning |
|---|---|
| The name is not in the row | The input is not given. |
| The name with a text | The input is given, and the text is the value received. An empty text is given. |
| The name with a list of texts | A list input is given, and each item is a text. An empty list is given. |

The value received is shown in a problem as exact text:

- A text is shown as it was received, character for character: `5.125` stays `5.125`, `-5.00` stays `-5.00`, `0.4370` stays `0.4370`. A text input is shown without quotes in the `value` field, and with double quotes inside a sentence.
- An input that is not given has `given` false and `value` null.
- A list is shown as its items between `[` and `]`, separated by `, `: `[2400.50, 3100.25]`. An empty list is `[]`. After the 20th item the text goes on with `, ...` before `]`.
- A problem about one item of a list names the item by its position, counted from 1, and shows the text of that item alone.
- A text longer than 200 characters is shown as its first 200 characters followed by `...` (proposed).

How the type rule reads the text:

| Type | Accepted text | Refused |
|---|---|---|
| integer | An optional `-`, then 1 to 18 digits. No leading zero unless the number is `0`. | `-0`, `+5`, `1237.5`, `1e3`, `1,200`, a space, an empty text. |
| decimal | An optional `-`, digits, and optionally a dot with 1 or more digits after it. At most 30 digits in all. No leading zero unless the whole part is `0`. | `.5`, `5.`, `1e3`, `1,5`, `+5`, a space, an empty text, more than 30 digits. |
| float | The same text as a decimal. It is read as an approximate number after it passes. | The same as a decimal. |
| text | 0 to 200 characters with no control character. | More than 200 characters, any control character. |
| boolean | `true` or `false`, in lowercase. | `True`, `yes`, `1`, an empty text. |
| date | `YYYY-MM-DD`, a day that exists. | `2026-02-30`, `31/03/2026`, an empty text. |
| list of integer, list of decimal | A list whose items each read as the element type. | A text received for a list input, a list received for another type. |

How the other rules read the text:

- Decimal places are the digits after the dot in the text, trailing zeros included. An integer has 0 places.
- Range, no negatives and the comparisons in checks read the value of the text as an exact decimal.
- Allowed values compare the characters of the text.

A value that has passed its type rule is calculated as that type, with its digits as received: `5.125` is the decimal `5.125` and `45.2` is the decimal `45.2`.

### The Validate problem

Validate tests one row against one definition and gives a list of problems. A problem is one object. The fields are always written, in this order, and the value is null where it does not apply.

```json
{
  "input": "payment",
  "item": null,
  "given": true,
  "value": "-5.00",
  "rule": "no-negatives",
  "sentence": "`payment` is -5.00. It cannot be negative.",
  "code": "no-negatives",
  "check": null
}
```

| Field | Meaning |
|---|---|
| `input` | The name of the input. Null for a problem of a check. |
| `item` | The position, from 1 to 1000, of the list item the problem is about. Null when the problem is about the whole input. |
| `given` | True when the name was in the row, false when it was not. Null for a problem of a check. |
| `value` | The value received, as exact text, or null when the input was not given. For a list it is the list as shown above, and for a list item it is the text of that item. Null for a problem of a check. |
| `rule` | The rule broken: `required`, `type`, `min-items`, `max-items`, `no-negatives`, `range`, `decimal-places`, `allowed-values`, `check`, `known-input` or `batch`. |
| `sentence` | One plain sentence a treasurer can read. It names the input in backticks and gives the value received. |
| `code` | A code from the table below, lowercase letters, digits and hyphens, 1 to 40 characters. |
| `check` | The standard text of the formula of the check. Null for any problem that is not of a check. |

A problem of a check has `input`, `given` and `value` null, because a check can read several inputs. Its sentence gives the message of the check and then the values received of the inputs the check reads, such as `` The payment must be more than zero. `payment` is 0.00. ``

```json
{
  "input": null,
  "item": null,
  "given": null,
  "value": null,
  "rule": "check",
  "sentence": "The payment must be more than zero. `payment` is 0.00.",
  "code": "check-failed",
  "check": "payment > 0"
}
```

This is the shape that the per-row results of a batch run carry. The batch page names six parts: the input, whether it was given, the value received, the rule broken, a sentence and a code. They are the same fields, in the same order. This page adds `item` and `check`, which are null in an ordinary problem. See "Differences to settle".

### The order of problems, and what stops what

Validate tests every rule of every input and never stops at the first problem. The problems come in this order:

1. The inputs in the order of the definition. For each input, its rules in the order of the table of rules: `required`, `type`, `min-items`, `max-items`, `no-negatives`, `range`, `decimal-places`, `allowed-values`.
2. For a list input, `required`, `type`, `min-items` and `max-items` come once for the list. Then the items come in position order, and for each item its `type`, then `no-negatives`, `range`, `decimal-places` and `allowed-values`.
3. Every name in the row that is not an input of the definition, in the order of its characters. Each gives `unknown-input`.
4. The checks in the order of the definition.

What one problem of an input does to the other rules of the same input:

- An input that is not given gets `required` and nothing else. It gets no `type`, no range and no other problem. The exception is a list input without `required`: it counts as an empty list and gets `min-items` when that rule is present, with `given` false.
- An input whose text fails the type rule gets `wrong-type` and nothing else, because the text cannot be read as a number, a date or the like. It gets no `no-negatives`, no range, no places and no allowed values.
- An item of a list that fails the type rule gets `wrong-type` and nothing else for that item. The other items are still tested, and so are `min-items` and `max-items`.
- A list received for an input that is not a list, or a text received for a list input, gets `wrong-type` and nothing else.
- Every other pair of rules is independent. A value that is both below zero and below the range gets both `no-negatives` and `too-small`. A value of `-5.125` with 2 places gets `no-negatives` and `decimal-places`, in that order.
- A problem in one input never stops another input from being tested.

What a problem does to the checks:

- A check runs only when every input it reads had no problem. A list input has a problem when the list or any of its items has one.
- A check that is skipped gives no problem. A skipped check is never invented as a failure, so the list is complete for what could be tested. After the inputs are fixed, a check can then give a problem that the first list did not show.
- A check that runs gives `check-failed` when it is false.
- A check that runs and cannot be worked out, for example by a division by zero, gives `check-error`. Its sentence carries the sentence of that problem. The codes `divide-by-zero`, `number-too-big` and `empty-list` are never codes of Validate; they stay codes of a run, as listed under Every code.
- An unknown name does not stop a check.

Validate returns an empty list when the row is clean. The formula is calculated only then. A row with one problem or more is never calculated (law 7 on `documentation/4. qa/laws.md`). The most one row can give from declared inputs is 24016 problems: 8 list inputs of 3001 problems each, which is 1000 items with 3 each and 1 for the list, and 8 checks. Unknown names add one each. There is no cap on the list (proposed), because a list that is cut is not complete.

Rules on one input when it is a list, shown on formula 10: the input `scores` is a list of integer with `min-items` 1, `max-items` 12 and a `range` of 0 to 100. The row `["-3", "104", "abc"]` gives 3 problems at once, one per item, in position order (`rules-each-item-three-problems`).

### Sentences

A sentence names the input in backticks and gives the value received. A value of type text is written between double quotes in a sentence, and every other value as received. A sentence for an item starts with `Item {n} of` and the name.

Check messages are written by the builder. Three examples with the sentence each gives:

| Message | Check formula | Sentence when it fails |
|---|---|---|
| The payment must be more than zero | `payment > 0` | `` The payment must be more than zero. `payment` is 0.00. `` |
| The loan must be at least 100.00 | `principal >= 100.00` | `` The loan must be at least 100.00. `principal` is 50.00. `` |
| The interest cannot be more than the amount borrowed | `principal * monthly_rate * months <= principal` | `` The interest cannot be more than the amount borrowed. `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60. `` |

The values are given for each name the check reads, in the order the names first appear in the standard text of the check, joined by `, ` and the word `and` before the last.

### Every code of Validate

Codes of a Validate problem.

| Code | Rule | When | Sentence | Example |
|---|---|---|---|---|
| `required` | `required` | The name is not in the row. | `` `{input}` was not given. It is required. `` | `rules-required-missing` |
| `wrong-type` | `type` | The text cannot be read as the type. | `` `{input}` is {value}. It must be {type words}. `` | `rules-type-invalid` |
| `too-small` | `range` | The value is below `min`, or equal to it when `min_exclusive`. | `` `{input}` is {value}. It must be {min text}. `` | `rules-range-too-small` |
| `too-large` | `range` | The value is above `max`, or equal to it when `max_exclusive`. | `` `{input}` is {value}. It must be {max text}. `` | `rules-range-too-large` |
| `no-negatives` | `no-negatives` | The value is below zero. | `` `{input}` is {value}. It cannot be negative. `` | `rules-no-negatives-invalid` |
| `decimal-places` | `decimal-places` | The text has more places than `places`. | `` `{input}` is {value}. It can have at most {n} decimal places. `` | `rules-decimal-places-invalid` |
| `not-allowed` | `allowed-values` | The text is not one of the values. | `` `{input}` is "{value}". It must be one of {values}. `` | `rules-allowed-values-invalid` |
| `list-too-short` | `min-items` | The list has fewer than `items` items. | `` `{input}` has {n} items. It needs at least {items} items. `` | `rules-list-at-least-one-invalid` |
| `list-too-long` | `max-items` | The list has more than `items` items, or more than 1000. | `` `{input}` has {n} items. It can have at most {items} items. `` | `rules-list-too-long` |
| `check-failed` | `check` | A check that ran is false. | `{message}. {values}.` | `rules-check-failed` |
| `check-error` | `check` | A check that ran cannot be worked out. | `The check "{message}" could not be worked out: {run sentence}` | `rules-check-error` |
| `unknown-input` | `known-input` | A name in the row is not an input of the definition. | `` `{input}` is not an input of this function. Its inputs are: {names}. `` | `rules-unknown-input` |

The words in the sentences:

- `{type words}` is: `a whole number with at most 18 digits`, `a decimal number with at most 30 digits, such as 12.50`, `a number with at most 30 digits, such as 0.37`, `text of at most 200 characters`, `true or false`, `a date written like 2026-03-31`, `a list of whole numbers`, `a list of decimal numbers`. For an item of a list the element's words are used.
- `{min text}` and `{max text}` are: `from {min} to {max}` when both bounds are given and neither is exclusive; otherwise for `too-small` `{min} or more`, or `more than {min}` when `min_exclusive`; and for `too-large` `{max} or less`, or `less than {max}` when `max_exclusive`.
- The places sentence reads `It can have at most {n} decimal places.` for 2 or more, `It can have at most 1 decimal place.` for 1, and `It cannot have decimal places.` for 0.
- `{items}` and `{n}` read `1 item` for 1 and `{n} items` otherwise.
- A list that is not given gives `` `{input}` was not given, so it has 0 items. It needs at least {items} items. ``

Codes of a row that the batch stopped. They belong to the batch page and are proposed. The row has no input, item, given or value, its `rule` is `batch`, and its stage is `batch` on that page.

| Code | When | Sentence | Status |
|---|---|---|---|
| `batch-limit-reached` | The row was never started because the batch reached its limit. | `This row was not worked because the batch reached its limit.` | proposed |
| `batch-cancelled` | The row was never started because the caller cancelled the batch. | `This row was not worked because the batch was cancelled.` | proposed |

`divide-by-zero`, `number-too-big` and `empty-list` are codes of a run, not of Validate, and they stay as listed under Every code. A check that meets one of them gives `check-error`.

### Refusing a definition on save

A definition is tested when the function is saved. This is a Check rule for the definition itself. A definition that cannot be satisfied, or that contradicts itself, is refused and nothing is stored. The problems are the problem of Check: `code`, `sentence`, `part` and `path`. For a problem of the definition, `part` is its place, such as `inputs.bag_size.range` or `checks.1`, and `path` is null. For a problem inside the formula of a check, the code, part and path are those of Check on that formula and the sentence starts with `Check {n}: `, where n counts from 1.

| Code | When | Sentence | Example |
|---|---|---|---|
| `input-repeated` | Two inputs have the same name. | `` The input `{name}` is declared more than once. `` | none |
| `bad-input-name` | A name breaks the pattern of a name. | `` `{name}` is not a valid input name. A name has 1 to 40 lowercase letters, digits and underscores and starts with a letter. `` | none |
| `too-many-inputs` | More than 8 inputs. | `This function goes over the input limit of 8.` | none |
| `rule-repeated` | A rule appears twice in one input. | `` The rule {rule} is given twice for `{input}`. `` | none |
| `rule-not-for-type` | A rule on a type it does not apply to. | `` `{input}` is {type}, so it cannot have the rule {rule}. `` | none |
| `bad-rule-parameter` | A parameter is missing, is not a number or text of the right kind, or is outside its limit. | `` The rule {rule} of `{input}` needs {what}. `` | none |
| `places-too-many` | `places` above 10. | `` The rule decimal-places of `{input}` allows {n} places. The most is 10. `` | `rules-definition-places-too-many` |
| `range-reversed` | `min` above `max`, or equal to it with an exclusive end. | `` The range of `{input}` has a minimum of {min} above its maximum of {max}. `` | `rules-definition-range-reversed` |
| `items-reversed` | `min-items` above `max-items`. | `` `{input}` needs at least {a} items and at most {b}. `` | none |
| `allowed-values-empty` | `values` has 0 values. | `` The allowed values of `{input}` list no value. A list holds 1 to 20 values. `` | `rules-definition-allowed-values-empty` |
| `allowed-values-too-many` | `values` has more than 20 values. | `` The allowed values of `{input}` list {n} values. The most is 20. `` | none |
| `allowed-values-repeat` | A value is listed twice. | `` The value "{value}" is listed twice for `{input}`. `` | none |
| `rules-contradict` | An allowed value that another rule of the input refuses. | `` No value can meet the rules of `{input}`: the allowed value {value} breaks the rule {rule}. `` | none |
| `input-needs-required` | An input that is not a list has no `required`. | `` `{input}` is not a list, so it needs the rule required. `` | `rules-definition-input-needs-required` |
| `too-many-checks` | More than 8 checks. | `This function goes over the check limit of 8.` | none |
| `bad-message` | A message with 0 characters, more than 120, or a line break. | `Check {n} needs a message of 1 to 120 characters on one line.` | none |
| `check-not-boolean` | The formula of a check is not true or false. | `` Check {n} must be true or false, and `{formula}` is {type}. `` | `rules-definition-check-not-boolean` |
| `unknown-name` | A check reads a name that is not an input. This is the code of Check, with `Check {n}: ` before its sentence. | `` Check {n}: The name `{name}` is not an input of this function. Its inputs are: {names}. `` | `rules-definition-check-unknown-name` |

A check formula that has any other problem of Check keeps its own code. A check may call built-in functions only (proposed). Check gives at most 20 problems for a definition, as it does for a formula.

A definition with a `max-items` rule lowers the cost of its list input from 1000 items to that number, for the cost of a formula that reads it.

### Examples

Each example below is a formula with its case line. The definition the row is validated against is in the table before it, and the row received is on the lines after it. A row that is clean gives the result of the formula. A row with problems gives the list, in order, and the formula is not calculated.

#### Balance after payment, formula 1

| Input | Type | Rules |
|---|---|---|
| `balance` | decimal | required, no-negatives, decimal-places 2 |
| `payment` | decimal | required, no-negatives, decimal-places 2 |

| Check | Formula | Message |
|---|---|---|
| 1 | `payment > 0` | The payment must be more than zero |
| 2 | `payment <= balance + 1000.00` | The payment cannot be more than 1000.00 above the balance |

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-required-valid

Received: `balance` 120.50 and `payment` 45.25. No problem. The result is `75.25`.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-required-missing

Received: `balance` 200.00, and `payment` not in the row. Problem `required`, input `payment`, not given, value null: `payment` was not given. It is required. Both checks read `payment`, so neither runs.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-required-and-bad-together

Received: `payment` -5.00, and `balance` not in the row. 2 problems, in input order. Problem `required`, input `balance`, not given: `balance` was not given. It is required. Problem `no-negatives`, input `payment`, received -5.00: `payment` is -5.00. It cannot be negative.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-no-negatives-valid

Received: `balance` 0.00 and `payment` 45.25. Zero is not negative. No problem. The result is `0.00`.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-no-negatives-invalid

Received: `balance` 80.00 and `payment` -5.00. Problem `no-negatives`, input `payment`, received -5.00: `payment` is -5.00. It cannot be negative. The checks read `payment` and are skipped.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-decimal-places-valid

Received: `balance` 120.50 and `payment` 45.2. One place is below the limit of 2. No problem. The result is `75.30`.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-decimal-places-invalid

Received: `balance` 60.00 and `payment` 5.125. Problem `decimal-places`, input `payment`, received 5.125: `payment` is 5.125. It can have at most 2 decimal places.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-three-problems-at-once

Received: `balance` -1.00 and `payment` -5.125. The list has 3 problems and nothing stops at the first. Problem `no-negatives`, input `balance`, received -1.00: `balance` is -1.00. It cannot be negative. Problem `no-negatives`, input `payment`, received -5.125: `payment` is -5.125. It cannot be negative. Problem `decimal-places`, input `payment`, received -5.125: `payment` is -5.125. It can have at most 2 decimal places.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-check-failed

Received: `balance` 120.50 and `payment` 0.00. Both inputs are clean, so both checks run. Check 1 is false and check 2 is true. Problem `check-failed`, rule `check`, check `payment > 0`: The payment must be more than zero. `payment` is 0.00.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-check-failed-second

Received: `balance` 100.00 and `payment` 2000.00. Check 1 is true. Check 2 is false, because 2000.00 is more than 1100.00. Problem `check-failed`, check `payment <= balance + 1000.00`: The payment cannot be more than 1000.00 above the balance. `payment` is 2000.00 and `balance` is 100.00.

#### Farmer net payout, formula 4

| Input | Type | Rules |
|---|---|---|
| `kilos` | integer | required, no-negatives |
| `price_per_kilo` | decimal | required, range more than 0.000, decimal-places 3 |
| `credit` | decimal | required, no-negatives, decimal-places 2 |

```formula
max([0.00, round(kilos * price_per_kilo, 2) - credit])
```
case: rules-decimal-places-boundary-valid

Received: `kilos` 853, `price_per_kilo` 0.437 and `credit` 120.00. Three places is the limit and is accepted. No problem. The result is `252.76`.

```formula
max([0.00, round(kilos * price_per_kilo, 2) - credit])
```
case: rules-decimal-places-trailing-zero

Received: `kilos` 853, `price_per_kilo` 0.4370 and `credit` 120.00. The text has 4 places because the zero is kept. Problem `decimal-places`, input `price_per_kilo`, received 0.4370: `price_per_kilo` is 0.4370. It can have at most 3 decimal places.

```formula
max([0.00, round(kilos * price_per_kilo, 2) - credit])
```
case: rules-different-codes-at-once

Received: `kilos` -3 and `price_per_kilo` 0.0005, and `credit` not in the row. 3 problems with 3 different codes, in input order. Problem `no-negatives`, input `kilos`, received -3: `kilos` is -3. It cannot be negative. Problem `decimal-places`, input `price_per_kilo`, received 0.0005: `price_per_kilo` is 0.0005. It can have at most 3 decimal places. Problem `required`, input `credit`, not given: `credit` was not given. It is required.

#### Seed bags left over, formula 19

| Input | Type | Rules |
|---|---|---|
| `kilos` | integer | required, no-negatives |
| `bag_size` | integer | required, range 1 to 100 |

```formula
kilos % bag_size
```
case: rules-type-valid

Received: `kilos` 1237 and `bag_size` 50. No problem. The result is `37`.

```formula
kilos % bag_size
```
case: rules-type-invalid

Received: `kilos` 1237.5 and `bag_size` 50. Problem `wrong-type`, input `kilos`, received 1237.5: `kilos` is 1237.5. It must be a whole number with at most 18 digits. `kilos` gets no `no-negatives` problem, because its text cannot be read as a whole number.

```formula
kilos % bag_size
```
case: rules-type-other-inputs-still-tested

Received: `kilos` abc and `bag_size` 0. 2 problems. Problem `wrong-type`, input `kilos`, received abc: `kilos` is abc. It must be a whole number with at most 18 digits. Problem `too-small`, input `bag_size`, received 0: `bag_size` is 0. It must be from 1 to 100.

```formula
kilos % bag_size
```
case: rules-unknown-input

Received: `kilos` 1237, `bag_size` 50 and `bag_weight` 3. Problem `unknown-input`, input `bag_weight`, given, received 3, rule `known-input`: `bag_weight` is not an input of this function. Its inputs are: kilos, bag_size. The formula is not calculated.

#### Late days, formula 5

| Input | Type | Rules |
|---|---|---|
| `due` | date | required |
| `filed` | date | required |

```formula
max([0, days_between(due, filed)])
```
case: rules-type-date-valid

Received: `due` 2026-03-31 and `filed` 2026-05-15. No problem. The result is `45`.

```formula
max([0, days_between(due, filed)])
```
case: rules-type-date-invalid

Received: `due` 2026-03-31 and `filed` 2026-02-30. There is no 30 February. Problem `wrong-type`, input `filed`, received 2026-02-30: `filed` is 2026-02-30. It must be a date written like 2026-03-31.

```formula
max([0, days_between(due, filed)])
```
case: rules-required-date-missing

Received: `due` 2026-03-31, and `filed` not in the row. Problem `required`, input `filed`, not given: `filed` was not given. It is required.

#### Voting right, formula 18

| Input | Type | Rules |
|---|---|---|
| `joined` | date | required |
| `on` | date | required |
| `paid_up` | boolean | required |
| `suspended` | boolean | required |
| `role` | text | required, allowed-values "member", "officer", "guest" |

```formula
!suspended && role != "guest" && (role == "officer" || paid_up && months_between(joined, on) >= 6)
```
case: rules-required-boolean-missing

Received: `joined` 2026-01-20, `on` 2026-09-15, `suspended` false and `role` member, and `paid_up` not in the row. Problem `required`, input `paid_up`, not given: `paid_up` was not given. It is required.

#### Loan limit check, formula 2

| Input | Type | Rules |
|---|---|---|
| `status` | text | required, allowed-values "active", "lapsed", "suspended" |
| `requested` | decimal | required, range more than 0.00, decimal-places 2 |
| `contributions` | list of decimal | min-items 1, no-negatives |

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-allowed-values-valid

Received: `status` lapsed, `requested` 1400.00 and `contributions` [200.00, 250.00, 150.00, 100.00]. No problem. The status is allowed, though it is not "active". The result is `false`.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-allowed-values-invalid

Received: `status` paused, `requested` 1400.00 and `contributions` [200.00, 250.00, 150.00, 100.00]. Problem `not-allowed`, input `status`, received paused: `status` is "paused". It must be one of "active", "lapsed", "suspended".

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-allowed-values-case-differs

Received: `status` Active, `requested` 1400.00 and `contributions` [200.00]. A different case is a different text. Problem `not-allowed`, input `status`, received Active: `status` is "Active". It must be one of "active", "lapsed", "suspended".

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-range-exclusive-valid

Received: `status` active, `requested` 0.01 and `contributions` [200.00]. 0.01 is more than 0.00. No problem. The result is `true`.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-range-exclusive-invalid

Received: `status` active, `requested` 0.00 and `contributions` [200.00]. The bound itself is refused. Problem `too-small`, input `requested`, received 0.00: `requested` is 0.00. It must be more than 0.00.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-no-negatives-list-item-valid

Received: `status` active, `requested` 1400.00 and `contributions` [200.00, 250.00, 150.00, 100.00]. No problem. The result is `true`.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-no-negatives-list-item-invalid

Received: `status` active, `requested` 1400.00 and `contributions` [200.00, -50.00]. Problem `no-negatives`, input `contributions`, item 2, received -50.00: Item 2 of `contributions` is -50.00. It cannot be negative.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-list-at-least-one-not-given

Received: `status` active and `requested` 1400.00, and `contributions` not in the row. The list has no `required` rule, so it counts as empty. Problem `list-too-short`, input `contributions`, not given, value null, rule `min-items`: `contributions` was not given, so it has 0 items. It needs at least 1 item.

#### Plot yield spread, formula 17

| Input | Type | Rules |
|---|---|---|
| `yields` | list of decimal | min-items 1, no-negatives, decimal-places 2 |

```formula
max(yields) - min(yields)
```
case: rules-list-at-least-one-valid

Received: `yields` [2400.50, 3100.25, 2750.00]. No problem. The result is `699.75`.

```formula
max(yields) - min(yields)
```
case: rules-list-at-least-one-invalid

Received: `yields` []. The list is given and has 0 items. Problem `list-too-short`, input `yields`, received [], rule `min-items`: `yields` has 0 items. It needs at least 1 item.

```formula
max(yields) - min(yields)
```
case: rules-decimal-places-list-item-invalid

Received: `yields` [2400.505, 3100.25]. Problem `decimal-places`, input `yields`, item 1, received 2400.505: Item 1 of `yields` is 2400.505. It can have at most 2 decimal places.

#### Fund share per member, formula 3

| Input | Type | Rules |
|---|---|---|
| `deposits` | list of decimal | min-items 1, no-negatives |
| `expenses` | list of decimal | no-negatives |
| `members` | integer | required, range 1 to 500 |

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: rules-range-valid

Received: `deposits` [500.00, 750.00, 1250.50], `expenses` [120.25, 80.00] and `members` 7. No problem. The result is `328.61`.

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: rules-range-edge-valid

Received: the same lists and `members` 500. The upper bound is allowed. 2300.25 / 500 is 4.6005, rounded to `4.60`.

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: rules-range-too-small

Received: the same lists and `members` 0. Problem `too-small`, input `members`, received 0: `members` is 0. It must be from 1 to 500.

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: rules-range-too-large

Received: the same lists and `members` 501. Problem `too-large`, input `members`, received 501: `members` is 501. It must be from 1 to 500.

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: rules-list-not-given-is-empty

Received: `deposits` [500.00, 750.00, 1250.50] and `members` 7, and `expenses` not in the row. `expenses` has no `min-items` and no `required`, so it counts as empty. No problem. The result is `357.21`.

#### Term grade, formula 10

| Input | Type | Rules |
|---|---|---|
| `scores` | list of integer | min-items 1, max-items 12, range 0 to 100 |

The `max-items` of 12 is added here to show the rule. The requirements page gives only at least one and the range.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: rules-each-item-valid

Received: `scores` [72, 85, 64, 91, 77]. No problem. The result is `"B"`.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: rules-each-item-too-large

Received: `scores` [72, 104]. Problem `too-large`, input `scores`, item 2, received 104: Item 2 of `scores` is 104. It must be from 0 to 100.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: rules-each-item-three-problems

Received: `scores` [-3, 104, abc]. 3 problems, in item order. Problem `too-small`, item 1, received -3: Item 1 of `scores` is -3. It must be from 0 to 100. Problem `too-large`, item 2, received 104: Item 2 of `scores` is 104. It must be from 0 to 100. Problem `wrong-type`, item 3, received abc: Item 3 of `scores` is abc. It must be a whole number with at most 18 digits.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: rules-list-at-most-valid

Received: `scores` with 12 items, each 70. The list is at the limit of 12. No problem. The average is 70 and the result is `"B"`.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: rules-list-too-long

Received: `scores` with 13 items, each 70. Problem `list-too-long`, input `scores`, received [70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70, 70], rule `max-items`: `scores` has 13 items. It can have at most 12 items.

#### Lending data check, formula 16

| Input | Type | Rules |
|---|---|---|
| `balances` | list of decimal | min-items 1, decimal-places 2 |
| `floor` | decimal | required, no-negatives, decimal-places 2 |

```formula
count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00
```
case: rules-type-list-item-valid

Received: `balances` [420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00] and `floor` 350.00. No problem. The result is `true`.

```formula
count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00
```
case: rules-type-list-item-invalid

Received: `balances` [420.00, n/a] and `floor` 350.00. Problem `wrong-type`, input `balances`, item 2, received n/a: Item 2 of `balances` is n/a. It must be a decimal number with at most 30 digits, such as 12.50.

#### Flat-rate loan instalment, formula 9

| Input | Type | Rules |
|---|---|---|
| `principal` | decimal | required, range more than 0.00, decimal-places 2 |
| `monthly_rate` | decimal | required, range 0.000 to 0.050, decimal-places 3 |
| `months` | integer | required, range 1 to 60 |

| Check | Formula | Message |
|---|---|---|
| 1 | `principal >= 100.00` | The loan must be at least 100.00 |
| 2 | `principal * monthly_rate * months <= principal` | The interest cannot be more than the amount borrowed |

```formula
round((principal + principal * monthly_rate * months) / months, 2)
```
case: rules-check-valid

Received: `principal` 2400.00, `monthly_rate` 0.012 and `months` 7. Both checks run and are true: the interest is 201.60. The result is `371.66`.

```formula
round((principal + principal * monthly_rate * months) / months, 2)
```
case: rules-check-two-failed

Received: `principal` 50.00, `monthly_rate` 0.050 and `months` 60. The inputs are clean. Both checks run and both are false, so there are 2 problems, in the order the checks are declared. Problem `check-failed`, check `principal >= 100.00`: The loan must be at least 100.00. `principal` is 50.00. Problem `check-failed`, check `principal * monthly_rate * months <= principal`: The interest cannot be more than the amount borrowed. `principal` is 50.00, `monthly_rate` is 0.050 and `months` is 60.

```formula
round((principal + principal * monthly_rate * months) / months, 2)
```
case: rules-check-skipped-input-problem

Received: `principal` 50.00, `monthly_rate` 0.06 and `months` 7. `monthly_rate` has a problem. Check 1 reads only `principal`, so it runs and is false. Check 2 reads `monthly_rate`, so it is skipped and gives nothing. Problem `too-large`, input `monthly_rate`, received 0.06: `monthly_rate` is 0.06. It must be from 0.000 to 0.050. Problem `check-failed`, check `principal >= 100.00`: The loan must be at least 100.00. `principal` is 50.00.

```formula
round((principal + principal * monthly_rate * months) / months, 2)
```
case: rules-check-skipped-all

Received: `principal` -1.00, `monthly_rate` 0.012 and `months` 7. `principal` has a problem and both checks read it, so both are skipped. Neither failure is invented, though check 1 would have been false. Problem `too-small`, input `principal`, received -1.00: `principal` is -1.00. It must be more than 0.00.

#### A check that cannot be worked out

| Input | Type | Rules |
|---|---|---|
| `total` | decimal | required, no-negatives, decimal-places 2 |
| `people` | integer | required |

| Check | Formula | Message |
|---|---|---|
| 1 | `total / people >= 1.00` | Each person must get at least 1.00 |

```formula
round(total / people, 2)
```
case: rules-check-error

Received: `total` 10.00 and `people` 0. Both inputs are clean, so the check runs and divides by zero. Problem `check-error`, check `total / people >= 1.00`: The check "Each person must get at least 1.00" could not be worked out: The part `total / people` divides by zero.

```formula
round(total / people, 2)
```
case: rules-check-division-valid

Received: `total` 100.00 and `people` 3. The check divides to 33.33333333333333333333, which is at least 1.00, so it is true. No problem. The result is `33.33`.

#### Definitions that are refused

Each definition below is refused on save. The formula is the one of the function, and the problem is the problem of the definition.

```formula
kilos % bag_size
```
case: rules-definition-range-reversed

`bag_size` has the rule range with `min` 100 and `max` 1. Problem `range-reversed`, part `inputs.bag_size.range`: The range of `bag_size` has a minimum of 100 above its maximum of 1.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-definition-places-too-many

`payment` has the rule decimal-places with `places` 11. Problem `places-too-many`, part `inputs.payment.decimal-places`: The rule decimal-places of `payment` allows 11 places. The most is 10.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: rules-definition-allowed-values-empty

`status` has the rule allowed-values with `values` []. Problem `allowed-values-empty`, part `inputs.status.allowed-values`: The allowed values of `status` list no value. A list holds 1 to 20 values.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-definition-check-unknown-name

Check 1 is `payment <= limit`, and `limit` is not an input. Problem `unknown-name`, part `limit`, path `[1]`: Check 1: The name `limit` is not an input of this function. Its inputs are: balance, payment.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: rules-definition-check-not-boolean

Check 1 is `payment + 1`. Problem `check-not-boolean`, part `checks.1`: Check 1 must be true or false, and `payment + 1` is decimal.

```formula
kilos % bag_size
```
case: rules-definition-input-needs-required

`bag_size` is an integer with the rule range 1 to 100 and no `required`. Problem `input-needs-required`, part `inputs.bag_size`: `bag_size` is not a list, so it needs the rule required.

## Examples

Each example is a formula, its case line, and a reminder of its inputs and its result. The functions the examples call are these.

| Function | Kind | Takes | Gives |
|---|---|---|---|
| `late_days@1` | agency, formula 5 | date, date | integer |
| `interest_step@1` | agency, formula 7 | decimal | decimal |
| `fin.pmt@1`, `fin.pmt@2` | library | decimal, decimal, integer | decimal |
| `step_a@1` to `step_e@1` | agency | decimal | decimal |

`step_a@1` calls no function. Each later step calls the one before it, so `step_e@1` has a chain of 4 below it.

### The nineteen example formulas

These are the formulas of `documentation/1. requirements/example-formulas.md`, in their standard text.

Formula 1, balance after payment.

```formula
payment >= balance ? 0.00 : balance - payment
```
case: lang-f01-balance-after-payment

Balance 120.50 and payment 45.25 give `75.25`. Payment 130.00 gives `0.00`.

Formula 2, loan limit check.

```formula
status == "active" && requested <= 3 * sum(contributions)
```
case: lang-f02-loan-limit-check

Status "active", requested 1400.00 and contributions [200.00, 250.00, 150.00, 100.00] give `true`.

Formula 3, fund share per member.

```formula
round((sum(deposits) - sum(expenses)) / members, 2)
```
case: lang-f03-fund-share-per-member

Deposits [500.00, 750.00, 1250.50], expenses [120.25, 80.00] and members 7 give `328.61`. With no expenses the result is `357.21`.

Formula 4, farmer net payout.

```formula
max([0.00, round(kilos * price_per_kilo, 2) - credit])
```
case: lang-f04-farmer-net-payout

Kilos 853, price 0.437 and credit 120.00 give `252.76`. Credit 400.00 gives `0.00`.

Formula 5, late days. Published as `late_days@1`.

```formula
max([0, days_between(due, filed)])
```
case: lang-f05-late-days

Due 2026-03-31 and filed 2026-05-15 give `45`. Filed 2026-03-20 gives `0`.

Formula 6, late-filing penalty.

```formula
late_days@1(due, filed) == 0 ? 0.00 : late_days@1(due, filed) > 90 ? round(amount * 0.10, 2) : late_days@1(due, filed) > 30 ? round(amount * 0.05, 2) : round(amount * 0.02, 2)
```
case: lang-f06-late-filing-penalty

Amount 8400.00, due 2026-03-31 and filed 2026-05-15 give `420.00`. Filed 2026-04-12 gives `168.00`.

Formula 7, interest rate step. Published as `interest_step@1`.

```formula
balance < 1000.00 ? 0.010 : balance < 5000.00 ? 0.015 : 0.020
```
case: lang-f07-interest-rate-step

Balance 3217.45 gives `0.015`. Balance 640.00 gives `0.010`. Balance 5000.00 gives `0.020`.

Formula 8, monthly interest.

```formula
round(balance * interest_step@1(balance), 2)
```
case: lang-f08-monthly-interest

Balance 3217.45 gives `48.26`.

Formula 9, flat-rate loan instalment.

```formula
round((principal + principal * monthly_rate * months) / months, 2)
```
case: lang-f09-flat-rate-instalment

Principal 2400.00, monthly rate 0.012 and months 7 give `371.66`.

Formula 10, term grade.

```formula
avg(scores) >= 80 ? "A" : avg(scores) >= 65 ? "B" : avg(scores) >= 50 ? "C" : "F"
```
case: lang-f10-term-grade

Scores [72, 85, 64, 91, 77] have the average 77.8 and give `"B"`.

Formula 11, pro-rated first-month fee.

```formula
round(monthly_fee * (days_between(joined, period_end) + 1) / days_in_period, 2)
```
case: lang-f11-prorated-first-month-fee

Fee 47.00, joined 2026-06-12, period end 2026-06-30 and 30 days in the period give `29.77`.

Formula 12, equipment value after straight-line depreciation.

```formula
max([salvage, cost - round((cost - salvage) / life_years, 2) * years_between(bought, on)])
```
case: lang-f12-straight-line-depreciation

Cost 12000.00, salvage 2000.00, life 7 years, bought 2023-09-10 and on 2026-11-17 give `7714.29`.

Formula 13, collateral coverage ratio.

```formula
round(to_float(stock_value) / to_float(facility_balance), 3)
```
case: lang-f13-collateral-coverage-ratio

Stock value 84500.00 and facility balance 62000.00 give the float `1.363`.

Formula 14, interim payment on percent complete.

```formula
max([0.00, round(contract * to_decimal(percent_complete, 2) * 0.90 - paid, 2)])
```
case: lang-f14-interim-payment

Contract 183500.00, percent complete the float 0.37 and paid 52000.00 give `9105.50`.

Formula 15, capital call share.

```formula
round(call_amount * units / sum(all_units), 2)
```
case: lang-f15-capital-call-share

Call amount 50000.00, units 1200 and all units [1200, 830, 1500, 477] give `14973.80`.

Formula 16, lending data check.

```formula
count(balances) >= 6 && avg(balances) >= floor && min(balances) >= 0.00
```
case: lang-f16-lending-data-check

Eight balances from [420.00, 380.50, 95.25, 510.00, 440.75, 300.00, 465.50, 389.00] and floor 350.00 give `true`.

Formula 17, plot yield spread.

```formula
max(yields) - min(yields)
```
case: lang-f17-plot-yield-spread

Yields [2400.50, 3100.25, 2750.00] give `699.75`.

Formula 18, voting right.

```formula
!suspended && role != "guest" && (role == "officer" || paid_up && months_between(joined, on) >= 6)
```
case: lang-f18-voting-right

Joined 2026-01-20, on 2026-09-15, paid up, not suspended and role "member" give `true`. The same member suspended gives `false`.

Formula 19, seed bags left over.

```formula
kilos % bag_size
```
case: lang-f19-seed-bags-left-over

Kilos 1237 and bag size 50 give `37`.

### Saving rewrites the spacing

Each formula below is what a person typed. The saved text is on the line after it.

```formula
payment>=balance?0.00:balance-payment
```
case: lang-save-tight-spacing

Saved as `payment >= balance ? 0.00 : balance - payment`. Balance 120.50 and payment 45.25 give `75.25`.

```formula
max( [0.00 ,round( kilos*price_per_kilo , 2 )-credit] )
```
case: lang-save-loose-spacing

Saved as `max([0.00, round(kilos * price_per_kilo, 2) - credit])`. It is formula 4, with the same result.

```formula
(!suspended) && (role != "guest") && ((role == "officer") || (paid_up && (months_between(joined, on) >= 6)))
```
case: lang-save-extra-brackets

Saved as `!suspended && role != "guest" && (role == "officer" || paid_up && months_between(joined, on) >= 6)`. It is formula 18. The brackets around `paid_up && ...` are dropped because `&&` binds tighter than `||`.

```formula
suspended?paid_up?1:2:3
```
case: lang-save-nested-conditional

Saved as `suspended ? (paid_up ? 1 : 2) : 3`. Suspended true and paid up false give `2`.

### What the operators give

```formula
total / people
```
case: lang-decimal-division-twenty-places

Total 10.00 and people 3 give `3.33333333333333333333`, 20 places.

```formula
kilos % bag_size
```
case: lang-remainder-negative-dividend

Kilos -7 and bag size 3 give `-1`. Kilos 7 and bag size -3 give `1`. Inputs may be negative when their rules allow it.

```formula
first == second
```
case: lang-decimal-equal-by-value

First 12.5 and second 12.50, both decimals, give `true`.

```formula
suspended ? 1 : 2.50
```
case: lang-conditional-widens-integer

Suspended true gives the decimal `1`. The answer is a decimal because the two answers have the common type decimal.

```formula
note == "say \"hi\" \\ bye"
```
case: lang-text-escapes

The text in quotes is `say "hi" \ bye`. Note `say "hi" \ bye` gives `true`.

```formula
fin.pmt@2(12000, 0.01, 24)
```
case: lang-library-call-version-two

The integer 12000 is accepted for the decimal parameter. The result is `564.88`.

### Check problems

Each formula below is checked against the inputs and functions named under it. The problem it gives is on the line after.

```formula
balance * rate
```
case: lang-problem-unknown-name

Inputs: balance (decimal), months (integer). Problem `unknown-name`, part `rate`, path `[1]`: The name `rate` is not an input of this function. Its inputs are: balance, months.

```formula
eval("1 + 1")
```
case: lang-problem-unknown-function

No inputs. Problem `unknown-function`, part `eval("1 + 1")`, path `[]`: The function eval is not known.

```formula
fin.pmt(12000, 0.01, 24)
```
case: lang-problem-version-missing

Problem `version-missing`, part `fin.pmt(12000, 0.01, 24)`, path `[]`: fin.pmt needs a version. Available: fin.pmt@1, fin.pmt@2.

```formula
fin.pmt@7(12000, 0.01, 24)
```
case: lang-problem-version-not-found

Problem `version-not-found`, part `fin.pmt@7(12000, 0.01, 24)`, path `[]`: `fin.pmt@7` does not exist. Latest is `fin.pmt@2`.

```formula
round@1(balance, 2)
```
case: lang-problem-version-not-allowed

Inputs: balance (decimal). Problem `version-not-allowed`, part `round@1(balance, 2)`, path `[]`: `round` is built in and takes no version.

```formula
round(balance)
```
case: lang-problem-wrong-argument-count

Inputs: balance (decimal). Problem `wrong-argument-count`, part `round(balance)`, path `[]`: `round` needs 2 arguments and this call has 1.

```formula
round(kilos, 2)
```
case: lang-problem-wrong-argument-type

Inputs: kilos (integer). Problem `wrong-argument-type`, part `kilos`, path `[0]`: Argument 1 of `round` must be a decimal or a float, and `kilos` is an integer.

```formula
status + 1
```
case: lang-problem-type-mismatch

Inputs: status (text). Problem `type-mismatch`, part `status + 1`, path `[]`: The part `status + 1` cannot be worked out: `+` needs two numbers, and this has text and integer.

```formula
balance ? 0.00 : 1.00
```
case: lang-problem-condition-not-boolean

Inputs: balance (decimal). Problem `condition-not-boolean`, part `balance`, path `[0]`: The condition `balance` must be true or false, and it is decimal.

```formula
suspended ? "yes" : 0
```
case: lang-problem-branches-differ

Inputs: suspended (boolean). Problem `branches-differ`, part `suspended ? "yes" : 0`, path `[]`: The two answers of `suspended ? "yes" : 0` must be the same kind, and they are text and integer.

```formula
balance - payment
```
case: lang-problem-float-decimal-mix

Inputs: balance (decimal), payment (float). Problem `float-decimal-mix`, part `balance - payment`, path `[]`: The part `balance - payment` mixes a decimal with a float; convert one with to_float or to_decimal.

```formula
max([0.00, "none"])
```
case: lang-problem-list-items-differ

No inputs. Problem `list-items-differ`, part `[0.00, "none"]`, path `[0]`: The items of `[0.00, "none"]` must all be the same kind, and this list has decimal and text.

```formula
a + b + c + d + e + f + g + h + i
```
case: lang-problem-too-many-inputs

Inputs: a to i, nine integers. Problem `too-many-inputs`, part the whole formula, path `[]`: This formula goes over the input limit of 8.

```formula
sum([sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x), sum(x)])
```
case: lang-problem-too-costly

Inputs: x (list of decimal). The formula has 44 nodes and depth 4. Each `sum(x)` costs 2 nodes plus 1000 items, 21 of them cost 21042, and the outer `sum` adds 21 items and 2 nodes: 21065. Problem `too-costly`, part the whole formula, path `[]`: This formula goes over the cost limit of 20000.

```formula
step_e@1(amount)
```
case: lang-problem-call-too-deep

Inputs: amount (decimal). `step_e@1` calls `step_d@1`, which calls `step_c@1`, then `step_b@1`, then `step_a@1`: a chain of 5 calls. Problem `call-too-deep`, part the whole formula, path `[]`: This formula goes over the call depth limit of 4. `step_d@1(amount)` has a chain of 4 and is accepted.

```formula
!!!!!!!!!!!!suspended
```
case: lang-problem-too-deep

Inputs: suspended (boolean). Twelve `!` and one name make a tree of depth 13. Parse gives problem `too-deep`: This formula goes over the depth limit of 12. Eleven `!` are accepted.

```formula
sum([1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1])
```
case: lang-problem-too-many-nodes

No inputs. A `sum` of a list of 99 ones has 101 nodes and depth 3. Parse gives problem `too-many-nodes`: This formula goes over the node limit of 100. A list of 98 ones has 100 nodes and is accepted.

```formula
status == "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
```
case: lang-problem-too-long

Inputs: status (text). The text between the quotes has 201 characters. Parse gives problem `too-long`: This formula goes over the text length limit of 200. A formula of 501 characters gives the same code with the sentence: This formula goes over the formula length limit of 500.

```formula
1234567890123456789 + kilos
```
case: lang-problem-too-many-digits

Inputs: kilos (integer). The first number has 19 digits. Parse gives problem `too-many-digits`: This formula goes over the number magnitude limit of 18.

```formula
balance * 0.12345678901
```
case: lang-problem-too-many-places

Inputs: balance (decimal). The decimal has 11 places. Parse gives problem `too-many-places`: This formula goes over the decimal places limit of 10.

### Parse failures

```formula
payment >= (balance ? 0.00 : balance - payment
```
case: lang-problem-bad-syntax-bracket

The text has 46 characters, so the end is column 47. Parse gives problem `bad-syntax`: The formula cannot be read: missing closing bracket at line 1, column 47.

```formula
balance * 007
```
case: lang-problem-bad-syntax-leading-zero

Parse gives problem `bad-syntax`: The formula cannot be read: a number cannot have a leading zero at line 1, column 11.

### Problems that appear in a run

```formula
round(total / people, 2)
```
case: lang-problem-divide-by-zero

Total 10.00 and people 0 give problem `divide-by-zero`, part `total / people`, path `[0]`: The part `total / people` divides by zero.

```formula
people == 0 ? 0.00 : total / people
```
case: lang-no-short-circuit

Total 10.00 and people 0 give problem `divide-by-zero`, part `total / people`, path `[2]`. The conditional evaluates both answers, so the condition does not guard the division. An input rule of the later section does.

```formula
kilos * kilos
```
case: lang-problem-number-too-big

Kilos 1000000000 give problem `number-too-big`, part `kilos * kilos`, path `[]`: The part `kilos * kilos` gives a number with more than 18 digits. The result 10 to the power 18 has 19 digits. Kilos 999999999 give `999999998000000001`.

```formula
max(yields)
```
case: lang-problem-empty-list

Yields [] give problem `empty-list`, part `max(yields)`, path `[]`: The part `max(yields)` needs a list with at least one item, and the list is empty.

### A tree that is not a formula

A tree from the studio's blocks can be wrong in a way no text can. Check gives `bad-tree`. This tree has an operator that is not in the language, and it has no formula example because no text produces it.

```json
{"kind": "binary", "op": "^", "left": {"kind": "name", "name": "balance"}, "right": {"kind": "literal", "type": "integer", "value": "2"}}
```

Problem `bad-tree`, part empty, path `[]`: The part at path [] is not a valid part: the operator ^ is not in the language.

## The Go surface

These are the signatures the language implies. They show shapes only. The package has no engine in it and no import of another mwanachama repo.

```go
package formula

type Type string

type NodeKind string

const (
	KindLiteral     NodeKind = "literal"
	KindName        NodeKind = "name"
	KindUnary       NodeKind = "unary"
	KindBinary      NodeKind = "binary"
	KindConditional NodeKind = "conditional"
	KindCall        NodeKind = "call"
	KindList        NodeKind = "list"
)

type Path []int

type Node interface {
	Kind() NodeKind
	Children() []Node
}

type Literal struct {
	Type  Type
	Value string
}

type Name struct {
	Name string
}

type Unary struct {
	Op      string
	Operand Node
}

type Binary struct {
	Op    string
	Left  Node
	Right Node
}

type Conditional struct {
	Condition Node
	Then      Node
	Else      Node
}

type Call struct {
	Library *string
	Name    string
	Version *int
	Args    []Node
}

type List struct {
	Items []Node
}

type Tree struct {
	Root Node
}

type Part struct {
	Path Path
	Text string
}

type Problem struct {
	Code     string
	Sentence string
	Part     string
	Path     Path
	Line     int
	Column   int
}

type Input struct {
	Name string
	Type Type
}

type Param struct {
	Name string
	Type Type
}

type Signature struct {
	Params    []Param
	Result    Type
	Cost      int
	CallDepth int
}

type Library interface {
	Function(library string, name string, version int) (Signature, bool)
	Versions(library string, name string) []int
}

type Rule struct {
	Name         string
	Min          *string
	Max          *string
	MinExclusive bool
	MaxExclusive bool
	Places       *int
	Values       []string
	Items        *int
}

type InputDecl struct {
	Name  string
	Type  Type
	Rules []Rule
}

type CheckDecl struct {
	Formula string
	Message string
}

type Definition struct {
	Inputs []InputDecl
	Checks []CheckDecl
}

type Received struct {
	Text  string
	Items []string
	List  bool
}

type Row map[string]Received

type RowProblem struct {
	Input    *string
	Item     *int
	Given    *bool
	Value    *string
	Rule     string
	Sentence string
	Code     string
	Check    *string
}

func Parse(text string) (Tree, []Problem)

func Format(tree Tree) string

func Check(tree Tree, inputs []Input, library Library) []Problem

func PartAt(tree Tree, path Path) (Part, bool)

func TreeFromJSON(data []byte) (Tree, []Problem)

func TreeToJSON(tree Tree) []byte

func CheckDefinition(definition Definition) []Problem

func Validate(definition Definition, row Row) []RowProblem
```

Each node kind implements `Node`. A call to a built-in or an agency function has a nil `Library`. A call with no version has a nil `Version`. `Line` and `Column` are set by Parse only. For an agency function, `Library` is the empty name in `Function` and `Versions`.
`Signature.CallDepth` is the longest chain of calls below the function: 0 when it calls no function.
There is no Hash here. The publication's hash reads the standard text from `Format`. Which hash function it uses is decided with the publication.
There is no run here. It comes with the batch page.
`CheckDefinition` tests a definition when a function is saved. It gives the problems of Check, and it calls `Check` on each check formula with the definition's inputs as its `[]Input`.
`Validate` gives an empty, non-nil list when the row is clean, and the formula is calculated only then. It is called only with a definition that `CheckDefinition` accepted. A `RowProblem` holds the fields of the Validate problem in their order; a pointer is nil where the JSON has null.
The page's `Input` and `Problem` already belong to Check, so the definition's input is `InputDecl` and the problem of a row is `RowProblem` (see Differences to settle).

## Left out on purpose

No formula needs any of these, so the language does not have them. Each is an open question for the owner.

- A float literal such as `0.75f`. The requirements page lists it for formulas 13 and 14, but those formulas contain none: `0.37f` and `1.363f` are an input and a result, and a float is made by `to_float` or given as an input.
- A date literal. Dates come from inputs and go only to `days_between`, `months_between` and `years_between`. Comparing, adding to, or testing dates for equality is left out.
- A minus sign before a number or a name, and the unary minus. No formula writes one.
- Integer divided by integer. `/` needs a decimal or a float on at least one side.
- `%` on decimals, a power operator, and any math function beyond `round`.
- Joining texts, ordering texts with `<`, and any function on text.
- `==` and `!=` on floats, booleans, dates and lists.
- Lists of float, text, boolean or date, lists of lists, an empty list written in a formula, and picking an item by position.
- `round` on an integer, `to_float` of an integer, and `to_decimal` of anything but a float.
- A call with no argument.
- Short-circuit evaluation and skipping the answer a condition did not choose.
- Names for partial results, anonymous functions, `map`, `range`, `fold`, `sort` and loops. Formula 6 repeats `late_days@1(due, filed)` three times rather than naming it.
- Comments and line breaks inside a formula.
- Reading a clock, a random number, a file, the network or an environment value.
- A rounding mode other than half away from zero.

## Open questions for the owner

Each is chosen provisionally on this page and is marked proposed.

1. List size. 1000 items at most in a list value (proposed).
2. Every number under Limits: formula text 500, depth 12, 100 nodes, 8 inputs, call depth 4, cost 20000, 20 problems per Check, 60 digits in a decimal result (all proposed). The margins are in the table above.
3. Evaluate everything, with no skipping in `&&`, `||` or a conditional (proposed). A guard such as `people == 0 ? 0.00 : total / people` then protects nothing, and the guard is an input rule.
4. Integer widens to decimal in list items, conditional answers and arguments, though no example needs it (proposed).
5. A float becomes a decimal through its shortest text, so 1.005 at 2 places is `1.01` (proposed). The other choice reads the exact binary value and gives `1.00`.
6. Decimal division gives exactly 20 places, half away from zero, and keeps the trailing zeros (proposed).
7. The language version of the built-ins: a whole number recorded once in the publication, covered by the hash, moved by the builder (proposed).
8. Whether to add a float literal, a date literal, a negative number, integer division, or any other part under "Left out on purpose".
9. Whether a name that equals a built-in name, such as an input called `max`, is allowed. It is allowed here because a call always has brackets (proposed).
10. Whether the problem sentences should be translated. They are English only here.
11. Whether an input that is not a list may be optional. Here every such input needs `required`, and a definition without it is refused with `input-needs-required`, because no stand-in value exists (proposed). The other choice is a `default` field on the input.
12. Whether `decimal-places` means at most `places` or exactly `places`. Here it is at most, and trailing zeros in the text count: `0.4370` fails 3 places (proposed).
13. Whether a problem of a check should carry the names and values of the inputs it reads as fields. Here `input`, `given` and `value` are null and the sentence carries the values (proposed).
14. Whether Validate may list no problem for a check it skipped. Here a skipped check gives nothing, so a second Validate after a fix can show a new `check-failed` (proposed). The other choice is a problem with the code `check-skipped`.
15. Whether the list of problems of a row is capped. Here it is not, and the most is 24016 plus one per unknown name (proposed).
16. Whether a check may call an agency function or a library function. Here it may call built-ins only, because Validate takes no library (proposed).
17. Whether a date input may have a range, and whether `allowed-values` may apply to a list of text. Neither is needed by an example, so neither exists here.
18. Whether a text longer than 200 characters, or a list of more than 20 items, is cut when shown in a problem. Here it is cut, with `...` (proposed).
19. Whether a name in a row that is not an input is a problem. Here it is `unknown-input`, so a host that merges shared inputs into rows must pass only declared names (proposed).
20. Whether a blank text for a text input, such as a text of spaces, is given. Here it is given, and `required` does not refuse it (proposed).

## Differences to settle

Pages other than this one disagree with it in these places. The owners of those pages can update them. This page is the language, so it does not change them.

1. `4. qa/laws.md` proposes depth 6, 40 nodes, 8 inputs and 1 to 200 rows. Here depth is 12, because formulas 12 and 14 have depth 8 and formulas 4, 6, 9, 10, 11 and 18 sit at 6 with no margin. Nodes are 100, because formula 6 has 34 and 40 leaves a margin of 6. Inputs stay 8. The rows per batch are not set here. The page also reads Check as testing "limits and shapes". Here Check also tests names, functions, types, call depth and cost. Its spacing generator must not add spaces inside a call name, around `.` or `@`.
2. `4. qa/laws.md` open question 4 asks for the hash function. This page only says the hash is taken of the standard text.
3. `4. qa/example-cases.md` and `3. implementation/example-cases.md` name the case file `documentation/4. qa/cases/formula-language.cases.md`. This page uses `documentation/4. qa/formula-examples.md`, as assigned. The sentence under the title is changed to match.
4. The case table limits a formula cell to 200 characters. The standard text of formula 6 has 175, but the examples of `too-many-nodes` and `too-long` have 302 and 213. The cell limit should be at least the formula text limit of 500.
5. The case table has no `float` typed value, and formula 13 gives a float. A float type must be added there first, or the case holds the result as text.
6. The case table's `expected` problem cell lists codes only. The codes are the ones in the table under Check. A problem also carries a sentence, a part and a path. The table must say whether a row compares those.
7. `4. qa/hostile-formulas.md` has limits not yet set. Here: formula length 500, depth 12, node 100, cost 20000, text length 200, number magnitude 30 digits for a decimal and 18 for an integer, decimal places 10 in a literal, list length 1000. The time limit is not set here, because the language has no loops and the engine page owns it. The sentences are the same.
8. Hostile P04 and P06 name a depth limit for nested brackets and lists. Brackets and `[` are counted by the same limit, so both give the depth sentence. A list of lists is also refused as a type.
9. Hostile P05 expects `missing closing bracket` for brackets never closed. Parse stops at the 13th open bracket first and gives the depth sentence.
10. Hostile P15 and P16 give columns 21 and 17. By the rule here the end of the text is one past the last character, which is column 20 for `sum(price, quantity` and column 19 for `price * quantity +`.
11. Hostile P13 asks whether an emoji is a name. Names are ASCII only, so it is `unexpected character`. Hostile O13 `__proto__` is refused because a name starts with a letter. `constructor` and `os` are ordinary names and are `unknown-name` unless they are inputs.
12. Hostile L01 uses `-(-(-(...)))`. This language has no unary minus, so `!` stands in for it. The hostile cases that use `range`, `map`, `fold`, `join`, `repeat`, `sort`, `^`, `e` exponents, field access on a list item, or `[1, 2, 3][99]` have no such part here. They give `unknown-function` or `bad-syntax` with the sentences above, not the run sentences of O09 to O11, C01 to C06, I03, I07, I08, I10 and I13. `round(1 / 3, 1000000)` gives `wrong-argument-type`, because the places are a literal from 0 to 10.
13. Hostile C07 gives the depth sentence for a call chain. Here it is `call-too-deep` with the call depth limit sentence. Hostile C09 and C11 expect `This formula calls itself`. Here no check can see a cycle, because a function may call only a published version. A self-call names a function not yet published and gives `unknown-function` or `version-not-found`.
14. `1. requirements/example-formulas.md` writes formula 18 with brackets around `paid_up && months_between(joined, on) >= 6`. The standard text drops them. The same page lists a float literal for formulas 13 and 14, which have none. Its conventions mention `sum`, `min`, `max` and `avg` of floats and lists of float; this language has only lists of integer and decimal. Its `%` convention says "remainder of integer division" and now has a sign rule. Its table lists brackets for formulas 3, 9, 11, 12, 14 and 18; formula 18 still has one bracket pair.
15. `2. design/library-versioning.md` shows the scan sentence with the prefix `Error:` and a mention of publish. Here the Check sentence has no prefix. The wording after it is the same. It does not say how a built-in is versioned. Here a built-in has no version and belongs to the language version.
16. `4. qa/exact-money.md` names several rounding modes and leaves division and float conversion open. Here `round` has only half away from zero, division gives 20 places, and a float is converted through its shortest text. For positive numbers half away from zero equals half-up. For negative numbers it does not equal "half toward positive infinity". The decimal and float refusal there has a free wording. Here it is the sentence of `float-decimal-mix`.
17. Check takes the function's input names and types as a second argument. The assignment named two arguments, a tree and a library, but a tree cannot be type-checked without the input types.
18. `2. design/batch-evaluation.md` names six parts of a problem. The Validate problem here has eight fields: the same six in the same order (input, given, value, rule, sentence, code), with `item` after `input` and `check` at the end, which are null in an ordinary problem. A problem of a check has `input`, `given` and `value` null. The batch page and `4. qa/bulk-runs.md` can keep writing `input (value) code`, and add the item position for a list.
19. The batch page says a calculation failure carries its reason "in the same problem form". A run problem here has the shape `code`, `sentence`, `part` and `path`. Proposed: in a batch row the run problem is carried as a Validate problem with `input`, `item`, `given`, `value` and `check` null, `rule` `calculation`, its sentence and its code, so a row has one form. The same holds for `batch-limit-reached` and `batch-cancelled`, with `rule` `batch`.
20. The batch page orders problems by input declaration order, then rule order. Here the order adds two steps after the inputs: unknown names, then checks. The rule order inside an input is the order of the rule table, which gives `no-negatives` before `decimal-places`, as `4. qa/bulk-runs.md` BR-24 row 4 lists them.
21. Batch example A and `4. qa/bulk-runs.md` give formula 1 inputs that are required, not negative and of 2 places, with no checks. Here "2 decimal places" means at most 2, and the worked definition adds two checks, so a `payment` of 0.00 gives `check-failed` there and not here. The batch rows still hold for the definition without the checks.
22. Batch example C writes the rule cell `pushed down: sum of contributions`. This page has no such rule. `total_contributions` is an ordinary decimal input here with `required`. Where a pushed-down declaration lives is left to the batch page, and a field of the definition is the proposed home.
23. `1. requirements/example-formulas.md` writes rules in words. They map to this page as follows: "more than 0.00" is `range` with `min` 0.00 and `min_exclusive` true; "at least one" is `min-items` 1; "each in range 0 to 100" is a `range` rule written once on the list input; "may be empty" is a list with no `required` and no `min-items`; "2 decimal places each" is `decimal-places` on a list; "range 0.0f to 1.0f" is `range` with `min` 0.0 and `max` 1.0, because a float literal has no suffix here; "2 decimal places" is at most 2.
24. `4. qa/laws.md` law 7 plants problems and expects the full list. A check problem is listed only when every input the check reads is clean, so the row generator must not expect a check problem for a row that also has a problem in an input the check reads. The code names `required`, `no-negatives` and `decimal-places` are as the batch page uses them.
25. `4. qa/example-cases.md`: the `expected` cell compares codes only, so two problems of the same code on different inputs, such as `no-negatives` twice, cannot be told apart. Proposed: a Validate row also compares input and item. A Validate row needs a definition, and the case table has no column for one; `4. qa/formula-examples.md` lists the definitions in a section of its own until the case-table owner decides.
26. The names `Input` and `Problem` are taken in the Go surface by Check, so the definition's input is `InputDecl` and the Validate problem is `RowProblem`. The signature is `Validate(definition Definition, row Row) []RowProblem`, with `CheckDefinition(definition Definition) []Problem` for the save.
