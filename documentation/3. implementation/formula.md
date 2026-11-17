# The formula front end

Package `formula/`, built for ISS-24. It reads a formula's text into a tree, prints a tree in the one standard shape, and checks a tree against everything the language sets. The language is in [formula-language.md](../2.%20design/formula-language.md). The decisions are 5, 6, 12 and 18 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not yet reviewed this page. Everything marked proposed is a starting point.

## Why it does not wait for the engine choice

The engine spike (ISS-4) showed that every option needs our own tree, our own reader and our own Format to reach the standard shape: cel-go and expr-lang/expr both needed a reverse translation and a printer of their own. So the tree, Parse, Format and Check are ours whichever engine is chosen. The choice only decides how a checked tree is run, which is ISS-25. The owner's direction was to build our own front end and take inspiration from the two libraries that were measured, without copying code.

## What it holds

| File | What it is |
|---|---|
| `tree.go` | the tree types of the Go surface, the types of a value, `Problem`, `Library` and `Signature`, and `PartAt` |
| `lex.go` | the reader of tokens, one at a time, with every limit of a single token |
| `parse.go` | `Parse`, the table of operators, the problem codes and the position of a problem |
| `format.go` | `Format`, the standard shape |
| `shape.go` | the first stage of Check: whether every part has the fields of its kind |
| `check.go` | `Check`, `Cost`, the size stage and the type walk of operators, conditionals and lists |
| `calls.go` | the built-in functions, calls to library and agency functions, and the adapter that lets `limit/` count the cost of a tree |

Imports: the standard library and `limit/` only. No comments in the code.

## Parse

- One problem at most. Parse stops at the first problem it meets, reading the text from the left.
- A problem from Parse has no part and a null path. It carries a line and a column counted in characters from 1. The reader keeps a character offset and the line and column are worked out when the problem is made.
- The text is refused first if it is not valid UTF-8 or longer than the formula length limit.
- Limits met while reading are refused at once with the sentence of `limit/`: the text literal, the digits of an integer or a decimal, the places of a decimal, the brackets open at one time. Depth and node count are tested on the finished tree.
- Spaces are allowed between tokens. A space is refused inside a call name: `late_days @ 1(...)` and `fin . pmt@1(x)` are read as a name followed by something that cannot continue it. A space between a call name and its bracket is allowed.
- A name is lowercase letters, digits and underscores, so a capital letter is an unexpected character.

## Format

The rules are the eleven of the standard shape. The brackets are worked out from a table of levels, the same table Parse reads: a left operand gets brackets when it is looser than the operator, a right operand when it is looser or equal, the operand of `!` when it is a binary operation or a conditional, a conditional in a condition or a then branch always, and an else branch never. The laws are tested on 500 generated trees: `Parse(Format(tree))` equals the tree, formatting twice gives the same text, and respacing the standard text gives it back.

## Check

Check reads a tree, never a number. It works in four stages and stops after the first stage that gives a problem:

1. Shape: every part has the fields of its kind and every value fits its pattern, otherwise `bad-tree` with an empty part and the path of the part at fault.
2. Size: formula length, depth, node count, distinct names, and the digits, places and length of literals, with the sentence of `limit/`. The problem names the whole formula and has the path `[]`.
3. Names, functions and kinds, children before their parent and left to right. A part with a problem has no kind, so the parts around it add no problem: a formula with one wrong name gives one problem at that name. At most 20 problems.
4. Call depth, then cost, from the declared depth and cost of each callee (`Signature`) and the meter in `limit/`.

`Cost` gives the cost of a tree on its own. The cost of the late-filing formula is 58 and of the term grade formula is 3019, as the language page says.

## Proposed on this page

None of these is in the language page. All are proposed and tested.

- A name over 40 characters is refused by Parse as an unexpected character at the 41st character.
- A limit problem from Parse carries the position of the part that went over it. A limit on the whole tree (depth, node count) carries line 1, column 1.
- An unknown name in a function with no inputs says `It has no inputs.` instead of listing them.
- A list of text, booleans or dates, or a list whose items have no common kind, gives `list-items-differ`. The first has its own sentence: the items must be integers or decimals.
- The second argument of `round` and `to_decimal` must be an integer written in the formula from 0 to 10, otherwise `wrong-argument-type`.
- `bad-tree` words follow the set on the language page. `the value {value} is not a {type}` keeps the page's article, so it reads `is not a integer`.
- A call has at most 8 arguments through the signature of its callee, so a ninth is `wrong-argument-count`.

## Ideas taken from the two libraries

Both libraries were read for design only, from the Go module cache, and no code was copied. What was read is named here. What was not read is not claimed.

| Idea | From | Where it shows |
|---|---|---|
| A table of binary operators giving each a precedence and a direction of grouping, so the reader and the printer share one table | expr-lang/expr (MIT), `parser/operator` | `binaryOperators` in `parse.go`, read again by `Format` |
| An error that carries a character position, and a later step that turns it into a line and a column | expr-lang/expr (MIT), `file/error.go` | `syntaxFailure` and `position` in `parse.go` |
| The checker as a pass of its own over a finished tree, apart from the reader | expr-lang/expr (MIT) `checker/` and cel-go (Apache 2.0) `checker/` | `check.go`, with Parse and Check kept apart so a tree from the studio's blocks is checked the same way |
| Cost counted before any run, from sizes assumed for lists | cel-go (Apache 2.0), `checker/cost.go` | `limit/` and `Cost`; a list that is not written in the formula counts as 1000 items until a rule declares less |

The planner of cel-go and the compiler and virtual machine of expr were not read for this issue. They are for the evaluator (ISS-25), where the same rule applies: ideas, written fresh.

## Not built

- `TreeFromJSON` and `TreeToJSON` of the Go surface. A tree is built and checked in Go. The studio's block view needs them.
- `CheckDefinition` and `Validate`, which belong to ISS-26.
- A run, which belongs to ISS-25.

## Tests

`go vet ./formula/` is clean and `go test ./formula/ -race` passes, with 134 passing tests and subtests, as does `go test ./...` for the whole module. The tests cover the nineteen example formulas (round trip, `Check` with a fake library), the eight spellings, 20 respaced variants of each formula, every Parse refusal and every limit with its sentence and position, one case per Check code with its sentence word for word, the `bad-tree` cases, the cap of 20 problems, the laws on generated trees, and `PartAt`.
