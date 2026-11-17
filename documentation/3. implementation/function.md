# Function definitions and Validate

Package `function/`, built for ISS-26. A function's definition holds its inputs, the rules of each input and its checks. `Validate` tests one row of received values against a definition and gives every problem at once. `CheckDefinition` refuses a bad definition when the function is saved. The rules and sentences are in [formula-language.md](../2.%20design/formula-language.md) under "Input rules, checks and Validate". The cases are on [input-checks.md](../4.%20qa/input-checks.md). The decisions are 8, 12, 18 and 22 in [decisions.md](../1.%20requirements/decisions.md).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## Open question first: negative inputs

Decision 22 says an input refuses negatives unless the function declares that it allows them. The language page has the opposite shape: a `no-negatives` rule refuses them, and an input without the rule accepts them. The package is built as the language page is written, so `no-negatives` is a rule like the others and an input without it accepts a negative value.

Case `ic-c07` (`price_per_kilo` of -0.5 with a range but no `no-negatives`) reads the page as written and gives only `too-small`. It runs as `TestNegativesWithoutTheRulePendingDecision22` and logs that it is pending. If the owner chooses decision 22, three things change: a declaration that allows negatives has to be named, the refusal of an undeclared input needs a sentence and a code, and the cases of `price_per_kilo`, `requested`, `scores`, `principal` and `members` get a second problem.

## What it holds

| File | What it is |
|---|---|
| `definition.go` | `Rule`, `InputDecl`, `CheckDecl`, `Definition`, `Row`, and the `Runner` interface with its `Outcome` |
| `types.go` | the eight input types with the `numtext/` type each is read as and the words of each in a sentence |
| `validate_input.go` | one input: not given, wrong type, the list rules, then no-negatives, range, places and allowed values |
| `validate.go` | `Validate`, `Prepare` and `Validator`: the inputs in order, unknown names, then the checks |
| `save.go` | `CheckDefinition` and the codes of a refused definition |

Imports: the standard library, `formula/`, `number/`, `numtext/` and `problem/`. No comments in the code.

## Using it

```go
definition := function.Definition{
    Inputs: []function.InputDecl{
        {Name: "payment", Type: formula.TypeDecimal, Rules: []function.Rule{
            {Name: "required"}, {Name: "no-negatives"},
            {Name: "decimal-places", Places: &two},
        }},
    },
    Checks: []function.CheckDecl{{Formula: "payment > 0", Message: "The payment must be more than zero"}},
}
saved := function.CheckDefinition(definition)
problems := function.Validate(definition, function.Row{"payment": numtext.One("-5.00")}, runner)
```

| Name | What it does |
|---|---|
| `CheckDefinition(definition)` | Gives the problems of a definition, as `formula.Problem` (`code`, `sentence`, `part`, `path`). An empty, non-nil list means it is accepted. At most 20 problems. A refused definition is stored as a draft that cannot be published (decision 12); storing is the caller's. |
| `Validate(definition, row, runner)` | Gives `[]problem.Problem` in the order of the language page. An empty, non-nil list means the row is clean and the formula may be calculated. A row with one problem or more is never calculated. |
| `Prepare(definition)` and `Validator.Validate(row, runner)` | The same, with the check formulas read once. A batch uses one `Validator` for every row. |
| `Row` | A map from input name to `numtext.Received`. `numtext.One(text)` is a text, `numtext.Many(items...)` a list, and a name left out is not given. A host that holds a number passes the digits as written. |

`Validate` is called only with a definition that `CheckDefinition` accepted. A row problem is the `problem.Problem` of [problem.md](problem.md): the same eight fields for a single run, a batch row and an AI tool call. What a batch or a tool adds around the list is not a field of a problem.

## The runner

A check is a formula that gives true or false, and Validate has to run it. The run belongs to ISS-25, so this package does not run formulas. It asks a `Runner`:

```go
type Runner interface {
    Run(tree formula.Tree, values map[string]numtext.Value) Outcome
}

type Outcome struct {
    True    bool
    Failure string
}
```

`Failure` is the plain sentence of a run that could not be worked out, such as a division by zero. It becomes `check-error`, with the message of the check and that sentence. A false answer becomes `check-failed`. A nil runner gives `check-error` for every check that would have run, with the sentence "No way to run this check was given." The tests use a small runner of their own for the operators the check cases need. The real one is wired in when ISS-25 lands: its adapter has to take a tree and the typed values and give an `Outcome`.

`Validate` differs from the Go surface on the language page in two ways, both proposed: it takes the runner as a third argument, and its problem is `problem.Problem` where the page names `RowProblem`.

## What Validate does

- The inputs come in the order of the definition. For each input the rules come in the order of the rule table. A list gets its list rules once and then its items in position order, each item with its type first.
- An input that is not given gets `required` and nothing else. A list without `required` counts as empty, gets `min-items` when that rule is present, and shows `given` false.
- A text that cannot be read as the type gets `wrong-type` and nothing else, so it never reaches no-negatives, range, places or allowed values. Other inputs are still tested.
- Values are compared as exact decimals, never as binary numbers. Places are counted in the text received, trailing zeros included. `-0.00` is not negative.
- A text input is shown between double quotes in a sentence and without quotes in the `value` field. A list is shown between brackets, at most 20 items and then `, ...`.
- Names in the row that are not inputs come after the inputs, in the order of their characters, each as `unknown-input`.
- A check runs only when every input it reads had no problem. A skipped check gives nothing. A check that runs gives `check-failed` or `check-error`. The values in a failed check's sentence are those of the inputs it reads, in the order the names first appear in its standard text.
- A list of more than 1000 items gets `list-too-long`, and only the first 1000 items are tested.

## What CheckDefinition refuses

The codes are those of the language page, with these places (`part`):

| Code | Part |
|---|---|
| `too-many-inputs`, `too-many-checks` | `inputs`, `checks` |
| `bad-input-name`, `input-repeated`, `input-needs-required` | `inputs.{name}` |
| `rule-repeated`, `rule-not-for-type`, `bad-rule-parameter`, `places-too-many`, `range-reversed`, `allowed-values-*`, `rules-contradict`, `items-reversed` | `inputs.{name}.{rule}` (`items-reversed` is on `min-items`, `rules-contradict` on `allowed-values`) |
| `bad-message`, `check-not-boolean` | `checks.{n}` |
| any problem of a check formula | the code, part, path, line and column of Parse or Check, with the sentence starting `Check {n}: ` |

A check formula reads only the definition's inputs and may call built-in functions only, so Check is called with no library. When Parse or Check gives a problem for a check, that is the only problem of that check: the type test for `check-not-boolean` runs only on a formula Check accepted. It uses `formula.ResultType`, a small exported function added in `formula/resulttype.go` over the checker's own type walk.

`rules-contradict` is built for an integer input with `allowed-values`: an allowed value that is not a whole number (rule `type`), is negative while `no-negatives` is on, or is outside the `range` is named with the rule it breaks. A text input has no other rule that can refuse a text, so none is checked.

## Proposed on this page

None of these is in the language page. All are tested.

- `bad-rule-parameter` also covers a rule name this library does not have (`The rule positive of `kilos` is not a rule of this library.`), an input type that is not one of the eight, more than 8 rules on one input, a rule that takes no parameter but is given one, a minimum or maximum that is not decimal text of at most 30 digits and 10 places, an exclusive end with no bound, and an allowed value over 200 characters. The words after `needs` are: `a minimum or a maximum written as decimal text`, `places, a whole number from 0 to 10`, `items, a whole number from 1 to 1000`, `values, a list of texts of at most 200 characters` and `no parameters`.
- `allowed-values` on an integer shows the received value between double quotes, as the sentence template does.
- A list that is not given and whose check reads it is shown as `[]` in the check's sentence.
- `rule-not-for-type` names the type as written: `` `status` is text, so it cannot have the rule no-negatives. ``

## What is not built, and what stays open

1. Negative inputs under decision 22, as above.
2. The run of a check. The `Runner` is the seam; ISS-25 supplies the real one.
3. A `max-items` rule lowers the cost of its list input from 1000 items to that number (language page, "Refusing a definition on save"). `formula.Input` has no field for it, so Check still counts a list input as 1000 items. It needs a field on `formula.Input`.
4. The language page's `TreeFromJSON`, `TreeToJSON` and the JSON form of a definition. A definition is built in Go here.
5. How an empty or space-only value reads in a sentence (`` `payment` is . ``), whether the `...` of a cut text is inside the quotes, and the shape of an AI tool answer: input-checks open questions 2, 4 and 6. The package words them as the page does.
6. Whether the items after the 1000th of a list are tested (input-checks open question 7). They are not.
7. What a check shows for a list that was not given (input-checks open question 8). It is shown as `[]`.
8. Whether a skipped check is reported (input-checks open question 9). It is not.
9. `allowed-values` on an integer and the quotes of its sentence (input-checks open question 5).

## Tests

`go vet ./function/` is clean and `go test ./function/ -race` passes with 218 passing tests and subtests:

- every case of sections A to M of input-checks.md except those that need the batch or a tool, including the 3000-problem case `ic-g06`, the whole-number, decimal and approximate-number text tables, the text, yes-or-no and date tables, and the list-for-single cases;
- `ic-c07`, which runs and logs that it is pending;
- the six refused definitions of the language page and 28 more refusals, the cap of 20, and the acceptance of the eleven definitions D1 to D11 of the QA page;
- that a check is not run when an input it reads has a problem, that a missing runner gives `check-error`, and that two calls give the same list.

The cases `ic-m04` to `ic-m06` are covered as far as this package goes: one `Validate` call per row, the same list on a second call. The three ways in (a single run, a batch row, an AI tool call) share the one function, so they cannot differ; the wrappers are not built.
