# The run of a checked formula

Built for ISS-25, in the package `formula/` next to the front end of [formula.md](formula.md). It plans a checked tree once and runs it on typed values, as many times as a host likes. The language is in [formula-language.md](../2.%20design/formula-language.md), the cases in [basic-runs.md](../4.%20qa/basic-runs.md). The decisions are 18 (Check reads the text, Validate and the run read values) and 19 (a cost limit and no clock) in [decisions.md](../1.%20requirements/decisions.md), and 23 (our own engine).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## Where it lives

The issue proposed a folder `run/`. The run is in `formula/` instead, in files named `run_*.go`, because it needs the tree, `Check`, `Cost` and the library signatures that live there, and a second package would have to import them and be imported back by `Plan`. The outside surface is the same either way. Imports: the standard library, `number/`, `datefn/` and `limit/`. No time, no random numbers, no context, and a test reads the imports of every `run_*.go` file to keep it so.

## The surface

```go
func Plan(tree Tree, inputs []Input, runtime Runtime, options Options) (*Program, []Problem)

func (p *Program) Run(values map[string]Value) (Value, *Problem)
func (p *Program) RunMetered(values map[string]Value, meter *Meter) (Value, *Problem)
func (p *Program) Apply(args []Value) (Value, *Problem)
func (p *Program) Cost() int
```

- `Plan` calls `Check` first. If Check gives any problem, `Plan` returns them and no program. So a program is always a checked one.
- `Run` takes the values by input name and gives the result, or one problem. It never gives both. A run stops at the first problem it meets, reading the tree left to right with children before their parent.
- A `Program` holds no state of its own. One program may be run by many callers at once, and each run has its own frame and its own meter. A test runs one program from eight goroutines under the race detector.
- `Apply` takes the values by position, in the order of the inputs given to `Plan`. It is how a program stands in as an agency function: a `Runtime` can hand the arguments of a call to the program of the callee.
- `RunMetered` takes the meter from the caller, so a host can see what a run spent (`Meter.Spent`).
- `ApplyMetered` takes the values by position and the meter of the run that is calling. A `Runtime` that also implements `MeteredRuntime` (`ApplyMetered(library, name, version, args, meter)`) is given a callee meter by the run, and hands it to `Program.ApplyMetered`. See How a callee is metered below.

### Values

`Value` is one typed value. The constructors are `IntegerValue`, `DecimalValue` (a `number.Decimal`), `FloatValue` (a `number.Float`), `TextValue`, `BooleanValue`, `DateValue` (a `datefn.Date`), `IntegerListValue` and `DecimalListValue`. `Type()` says which, and `String()` writes it the way the QA pages do: a decimal with its places (`75.25`, `0.5000`), a float in its shortest form (`3`, `1.363`), a list as `[1, 2]`.

A run needs every input of the plan, with exactly the declared type. A missing input, or one of another type, is the problem `bad-input` and the formula is not run. An integer is not accepted for a decimal input here, because the host has already read its values (Validate is ISS-26); the widening of an integer to a decimal happens inside the formula, where the language page puts it.

### Calls

```go
type Runtime interface {
	Library
	Apply(library string, name string, version int, args []Value) (Value, *Problem)
}
```

`Runtime` is the `Library` of the front end plus the one method that does the call. Built-ins are done by the package and never reach the runtime. For a library or an agency function the run widens an integer argument to a decimal where the signature takes a decimal, charges the callee's declared cost, and calls `Apply`. The result must have the type the signature says, with an integer widened to a decimal; otherwise the problem is `bad-function-result`. A problem the callee returns without a part is given the part and path of the call.

### Settings

```go
type Options struct {
	Limits  limit.Settings
	Numbers number.Settings
}
```

`DefaultOptions` is `limit.Proposed()` and `number.DefaultSettings()`. The rounding rule of `round` and `to_decimal` is `Numbers.Mode`, and the digits kept by a division are `Numbers.DivisionDigits`. The defaults are half away from zero and 20 digits, so a different agency setting (decisions 20 and 21) changes the figures with no change to the formula. The cost limit of a run is `Limits.Cost`. `Check` still tests the proposed limits, so a lower cost limit stops a run and does not refuse the plan.

## What each part gives

| Part | Result |
|---|---|
| `+ - *` | integer, decimal or float by the table of the language page; an integer mixed with a float is read as a float |
| `/` | decimal with `DivisionDigits` places, or a float; two integers are refused by Check |
| `%` | integer with the sign of the dividend |
| comparisons | boolean, read by value: `12.5 == 12.50` is true; text is equal only when every character is |
| `&&`, `\|\|`, a conditional | both sides are always worked out, so a problem in the answer not chosen is a problem of the formula |
| a conditional with an integer and a decimal answer | the chosen answer is a decimal: `1` stays `1` with no places |
| a list written in the formula | a list of integer, or a list of decimal when any item is a decimal, with the integers widened |
| `round(x, n)` | a decimal with exactly `n` places; a float goes through its shortest text, is rounded, and comes back as a float |
| `sum` | integer or decimal; an empty list is `0` (a decimal `0` has no places) |
| `count` | integer |
| `avg` | decimal with 20 places for decimals; a float for integers; an empty list is `empty-list` |
| `min`, `max` | the first item with the least or greatest value, as it was written; an empty list is `empty-list` |
| `to_float`, `to_decimal(x, n)` | the explicit conversions; a float is read from its shortest text |
| `days_between`, `months_between`, `years_between` | integers from `datefn/` |

The three problems of a run are `divide-by-zero`, `number-too-big` (an integer over 18 digits, a decimal over 60 digits, a float that is not finite) and `empty-list`, with the sentences of the language page. `too-costly` is the fourth and is below.

## How cost is counted

The count of the run mirrors the count of `Check`, and a test holds the two together.

- Each part costs 1, charged when the run reaches it, before its children.
- `sum`, `count`, `avg`, `min` and `max` charge the number of items of the list they were given, after the list is worked out and before any item is read. `Check` counts 1000 for a list that is not written in the formula. The run counts the real size.
- A call to a library function or an agency function charges the callee's declared cost, from its signature.

A run stops as soon as the meter goes past `Limits.Cost`, with the problem `too-costly`, the sentence `This formula goes over the cost limit of 20000.`, the whole formula as its part and the path `[]`. It is a failed calculation, so the step's failure choice applies (decision 19). Because an aggregate charges before it reads, a list of a million items is refused without being walked, and a second aggregate after the first crossed the limit never runs.

So a run costs no more than the count of `Check` while the lists are no larger than `Check` assumed. A list larger than that is what the run-time stop is for: `sum(xs)` is counted 1002 at save, and a run over 25000 items is stopped. This answers the question the hostile page left open ("does a formula over the cost limit reach a run at all"): a formula over its limit by count is refused at save, and a formula over it only by the size of a list is stopped at run time.

There is no clock. The same input spends the same amount on every machine, and a test checks that twenty runs spend the same.

## Ideas taken from the two libraries

Read for design only, from the module cache. No code is copied.

| Idea | From | Where it shows |
|---|---|---|
| A program planned once into steps that are run many times, with the names resolved to slots before the first run | cel-go (Apache 2.0) `interpreter/interpretable.go`, expr-lang/expr (MIT) `vm/program.go` | `Plan` and the `runStep` closures; a name is read from a slot, never looked up by text at run time |
| A cost tracker made fresh for each evaluation, charging as the work is done, with the run stopped when it goes over; the actual cost kept in step with the estimate of the checker | cel-go `interpreter/runtimecost.go`, whose header warns that a change there needs the same change in `checker/cost.go` | `Meter`, `runFrame.charge`, and the test that the run never spends more than `Cost` |
| A budget that is tested while the program runs and not only before it | expr-lang/expr `vm/vm.go` (`MemoryBudget`) | the aggregate charge before the loop |

The evaluators were read at their heads and outlines: `runtimecost.go`, `interpretable.go`, `vm.go` and `program.go`. The spike's own evaluator in `spike/engine/own` was read too, for its slots in a frame and for the order of evaluation. What was not read is not claimed.

## Proposed on this page

- `bad-input` and `bad-function-result` are new codes, not in the language page. Both are errors of the host and not of the formula.
- The sentence of `number-too-big` for a float is `The part `{part}` gives a float that is not a finite number.` The page gives only the digits sentence.
- An integer with a float is read as a float by converting the integer to a 64-bit float, which keeps 15 to 16 digits of a large integer.
- The average of integers is the integer sum divided as floats. The sum is checked against the 18 digits first.
- A callee is charged its declared cost at the call. See How a callee is metered for what a program used as a callee spends.
- The cost problem names the whole formula and the path `[]`, as the save-time `too-costly` does, and not the part where the limit was met.
- The decimal zero of an empty `sum` has no places (`0`), as basic-runs.md Gap 6 writes it.

## Differences from the issue text

- The issue's done-when says "a branch not taken that would fail does not fail". The language page says the opposite: a conditional, `&&` and `||` work out both sides, and `people == 0 ? 0.00 : total / people` is `divide-by-zero` when `people` is 0 (case `lang-no-short-circuit`, `basic-and-no-short-circuit`; open question 3 of the language page, proposed). The run follows the language page. The two cannot both hold, and the owner decides.

## Not built

- The cases of basic-runs.md sections 10 and 11 that name `fin.*` and `stats.*` library functions. They need the library functions of ISS-27. The run is tested with an agency function made from a program and with a native `fin.pmt@2` of the test.
- A host's own cancel (decision 19, open on the batch page).
- Validate, which reads text into values and checks the rules: ISS-26.
- Gaps 1 to 12 of basic-runs.md, which wait on the owner.

## Tests

`go vet ./formula/` is clean and `go test ./formula/ -race` passes. The run tests are `run_examples_test.go` (the nineteen example formulas with their expected results, the worked lines, and the problems with their part, path and sentence), `run_basic_test.go` (sections 1 to 9 and 12 of basic-runs.md, the settings, the inputs of a run, and the same result every run) and `run_cost_test.go` (what a run spends against the count, the stop at the limit, no clock, one program run by eight goroutines).

## How a callee is metered

Built for ISS-72. `Runtime.Apply` carries no meter, so a program that a host calls back through `Program.Apply` starts a fresh meter every time and a host that loops is never stopped by the library.

- A `Runtime` that implements `MeteredRuntime` is called through `ApplyMetered`. The run charges the callee's declared cost at the call, then gives the callee a meter of its own whose limit is what was left of the caller's limit before that charge, and whose depth is the caller's depth plus one.
- When the callee returns, the caller is charged the difference if the callee spent more than it declared. So a call costs the larger of the declared cost and the actual cost, and is never charged twice. A callee that spends exactly what it declared leaves the run's total equal to the static cost from `Check`.
- `Program.ApplyMetered` refuses a callee whose depth is over the call depth limit (4, proposed) with `call-too-deep` and the call depth sentence. So a function that calls itself, two functions that call each other and a self-call that doubles each time end at depth 5, with at most 2^5 calls for the doubling case, and the host does not have to stop them.
- A `Runtime` without `ApplyMetered` is called through `Apply` as before, with a fresh meter for each call. Such a host is the host's own to guard. `library/` and `examplecases/` are still in this group until they implement `ApplyMetered`.
- The callee's failure is returned as it is. A callee that runs out of its share gives `too-costly` with the callee's own formula as its part.

Open: the declared cost is still the callee's own claim. A callee that spends more is charged the difference after it has run, and the run's own limit stops it as it goes, because the callee's meter carries the remaining share.
