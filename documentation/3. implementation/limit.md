# The limits package

Folder `limit/`. Built for ISS-42. It holds every limit of the formula language as one named number in one place, the refusal sentence, and a cost meter that counts work. It reads no clock. Every number below is proposed until the owner confirms it.

Rules it implements: [decision 18](../1.%20requirements/decisions.md) (Check covers limits, depth and cost from the text alone) and [decision 19](../1.%20requirements/decisions.md) (a cost limit only; the library sets no time limit of its own). Built from the Limits and Cost sections of [the language page](../2.%20design/formula-language.md) and the cases of [the hostile formulas page](../4.%20qa/hostile-formulas.md).

## The settings

`limit.Proposed()` returns a `Settings`. Each field is a number a host can change, so no limit is a constant of the library.

| Field | Proposed value | Meaning |
|---|---|---|
| `FormulaLength` | 500 | characters of formula text |
| `Depth` | 12 | tree depth, the root being 1 |
| `BracketsOpen` | 12 | brackets open at one time |
| `Nodes` | 100 | nodes in a tree |
| `Inputs` | 8 | distinct names a formula reads |
| `CallArguments` | 8 | arguments in one call |
| `TextLiteral` | 200 | characters in a text literal |
| `IntegerDigits` | 18 | digits in an integer literal |
| `DecimalDigits` | 30 | digits in a decimal literal |
| `DecimalPlaces` | 10 | places of a decimal literal |
| `NameLength` | 40 | characters in a name |
| `CallDepth` | 4 | calls inside calls |
| `Cost` | 20000 | the cost of a formula |
| `ListItems` | 1000 | items in a list value, and the size the cost meter assumes for a list it cannot count |

## Refusing

`Settings.Refusal(measure)` takes a `Measure`, a struct of the sizes a parser or checker observed, and returns the first limit that is over, in the order of the table above, or no refusal. A size equal to its limit is accepted; one over is refused.

A `Refusal` has the limit's `Name` and the limit's value. `Sentence()` gives the one plain sentence the hostile page uses, and `Error()` gives the same text, so a refusal can be returned as an error:

> This formula goes over the depth limit of 12.

| Name constant | Text in the sentence | Measures that use it |
|---|---|---|
| `FormulaLengthLimit` | formula length | `FormulaLength` |
| `DepthLimit` | depth | `Depth`, `BracketsOpen` |
| `NodeLimit` | node | `Nodes` |
| `InputLimit` | input | `Inputs` |
| `TextLengthLimit` | text length | `TextLiteral` |
| `NumberMagnitudeLimit` | number magnitude | `IntegerDigits` (value 18), `DecimalDigits` (value 30) |
| `DecimalPlacesLimit` | decimal places | `DecimalPlaces` |
| `NameLengthLimit` | name length | `NameLength` |
| `CallDepthLimit` | call depth | `CallDepth` |
| `CostLimit` | cost | `Cost` |
| `ListLengthLimit` | list length | `ListItems` |

`Settings.AllowsArguments(count)` answers the arguments in one call. It has no limit sentence, because the language page gives that problem its own code, `wrong-argument-count`.

## The cost meter

The meter walks a tree through a small interface, `limit.Node`, so it does not depend on the formula front end that a later issue builds:

| Method | Gives |
|---|---|
| `Children()` | the node's children, in order |
| `Role()` | `RolePlain`, `RoleAggregate` (`sum`, `count`, `avg`, `min`, `max`), `RoleListLiteral` or `RoleCall` |
| `Items()` | for a list literal, its item count; for any other node, the items its function's `max-items` rule declares, or 0 |
| `CalleeCost()` | for a call, the callee's declared cost |

`Settings.CostOf(node)` counts as the language page says:

- each node costs 1;
- an aggregate adds the item count of its first child, which is its list: the written count for a list literal, otherwise the declared `max-items` when it is smaller than `ListItems`, otherwise `ListItems`;
- a call adds the callee's declared cost.

`NodeCount(node)` and `TreeDepth(node)` give the node count and the depth to fill a `Measure`. The count is the same on every run and every machine: nothing in the package imports `time`, `math/rand` or `context`, and a test reads the source to keep it so.

Checked on hand-built trees: the late-filing penalty of the margin table costs 58 (34 nodes and three calls at a callee cost of 8) and the term grade costs 3019 (19 nodes and three aggregates over lists the formula cannot count).

## What stays open

These are the points where the pages disagree or the owner has not decided. No QA page was edited.

1. The hostile cases C07, C08, C10 and C12 call the call chain refusal the depth limit; the language page names it call depth, and the package says call depth.
2. The language page holds a written list at 99 items by the node limit, but the hostile cases C01 and I11 refuse a million-entry written list with the list length sentence. The package uses list length for list values only.
3. The language page's list of limit names has no list length and no name length. The package defines both and the language page needs both added, or removed here.
4. The hostile page's Limits table lacks inputs, brackets open, name length, call depth and call arguments.
5. Number magnitude has two values, 18 and 30. The sentence names the value of the literal that went over, the integer literal first.
6. The package counts cost from the tree and stops nothing at run time. Whether and how a run is stopped by the count (hostile C02, C04, C06 and C13), and whether a host's own cancel is honoured, is left to the issue that runs formulas (decision 19).
7. `ListItems` is both the list value limit and the default list size the meter assumes. Whether those are two settings is open.
8. Problems listed by one Check (20, proposed) and the digits of a result (18 and 60) belong to Check and to a run, not to this package.
