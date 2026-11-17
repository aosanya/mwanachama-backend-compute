# The formula language

Every example on this page has a row in `documentation/4. qa/formula-examples.md`, marked by the `case:` line under the example.

This page fixes the language that the body of an agency function is written in.
It states what a formula may contain, how it is written in its one standard text, how it is read as a tree, and what is refused and with which words.
The laws, the hostile cases, the batch run and the library reference are written from it.

Status: written on 2026-11-17 from the nineteen example formulas. The owner has not reviewed it. Every number marked "proposed" is a starting point, not a decision. All names and figures are invented.

## What this page does not decide

- The input rules and checks of a function, and Validate, which tests a row of inputs. A later section of this page covers them.
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
- Every call checks the callee's own input rules. Those rules are specified in a later section of this page.
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
| Items in a list value | 1000 (proposed) | Validate, in the later section | not defined here |
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
- `sum`, `count`, `avg`, `min` and `max` add the number of items of their list. For a list written in the formula that is its item count. For any other list it is 1000, until the input rules of the later section declare a smaller size.
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

The words `{what}` of `bad-syntax` come from this set: `the formula is empty`, `text is not valid UTF-8`, `unexpected character {c}` (a character outside the language is written `U+` and its code, such as `U+200B`), `unexpected control character`, `missing closing bracket`, `a : is missing`, `the formula ends after an operator`, `text is not closed`, `text ends inside an escape`, `text has an unknown escape`, `a number cannot have a leading zero`, `a version is a whole number from 1 to 999999`, `a name must be followed by brackets to be a call`, `two parts are side by side with no operator between them`.

The words `{what}` of `bad-tree` come from this set: `the kind {kind} is not in the language`, `the field {field} is missing`, `the field {field} is not allowed`, `the operator {op} is not in the language`, `the value {value} is not a {type}`, `the name {name} is not a valid name`, `the part has {n} children and needs {m}`, `the text is not a tree`.

The `{needs}` of `type-mismatch` is: `two numbers` for `+ - *` and the four comparisons, `a decimal or a float on at least one side` for `/`, `two integers` for `%`, `two numbers or two texts` for `==` and `!=`, `two booleans` for `&&` and `||`, `a boolean` for `!`.

A call to a name that is none of these, such as `import`, `eval`, `now`, `random`, `env`, `exec` or `read_file`, is `unknown-function`. A formula can only compute from its own text and its inputs.

## Input rules, checks and Validate

Specified in the next section of this page, written separately.

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

func Parse(text string) (Tree, []Problem)

func Format(tree Tree) string

func Check(tree Tree, inputs []Input, library Library) []Problem

func PartAt(tree Tree, path Path) (Part, bool)

func TreeFromJSON(data []byte) (Tree, []Problem)

func TreeToJSON(tree Tree) []byte
```

Each node kind implements `Node`. A call to a built-in or an agency function has a nil `Library`. A call with no version has a nil `Version`. `Line` and `Column` are set by Parse only. For an agency function, `Library` is the empty name in `Function` and `Versions`.
`Signature.CallDepth` is the longest chain of calls below the function: 0 when it calls no function.
There is no Hash here. The publication's hash reads the standard text from `Format`. Which hash function it uses is decided with the publication.
There is no Validate and no run here. They come with the later section and the batch page.

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
