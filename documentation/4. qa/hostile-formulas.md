# Hostile formulas

Goal: a formula can never hang, crash or exhaust a host. Formulas are written by builders and by AI, some careless, some hostile, and this public library evaluates them. This page designs the test cases, each written as an attacker would try it, with the exact result that must follow. The formula engine is not chosen yet, so no limit has a value yet.

## Limits

| Limit | Value |
|---|---|
| depth | to be set by the engine choice |
| node | to be set by the engine choice |
| cost | to be set by the engine choice |
| time | to be set by the engine choice |
| formula length | to be set by the engine choice |
| text length | to be set by the engine choice |
| number magnitude | to be set by the engine choice |
| decimal places | to be set by the engine choice |
| list length | to be set by the engine choice |

## Refusal sentences

Limit refusal:

> This formula goes over the {limit} limit of {value}.

Parser error:

> The formula cannot be read: {what is wrong} at line {n}, column {m}.

Three more sentences are used by the cases below:

> This formula calls itself: {chain}.

> The function {name} is not known.

> The list has {n} entries and there is no entry {i}.

The same wording is used when the formula is saved and when it is run. `{limit}` is the name in the first column of the Limits table, and `{value}` is that row's value, both filled at run time. A refusal is always one plain sentence and never a stack trace.

In every table, Stage is `save`, `run` or `both`. Where a case says "the limit reached first", the sentence names whichever limit stopped the formula.

## P: Parser fuzzing

Must hold for every input: never panic, always return a plain sentence naming the line and column, return within the time limit, allocate within the limits above.

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| P01 | Random printable text, for example `qz(( ,, 4 4 ;;` | The formula cannot be read: {what is wrong} at line 1, column {m}. | save |
| P02 | Random bytes, including invalid UTF-8 such as `0xFF 0xFE 0xC0` | The formula cannot be read: text is not valid UTF-8 at line {n}, column {m}. | save |
| P03 | A NUL byte in the middle of a good formula, `1 + <NUL> 2` | The formula cannot be read: unexpected NUL character at line 1, column 5. | save |
| P04 | Deeply nested brackets, `((((((...1...))))))` nested past the depth limit | This formula goes over the depth limit of {value}. | save |
| P05 | Deeply nested brackets that are never closed, `((((((((` repeated | The formula cannot be read: missing closing bracket at line {n}, column {m}. Memory stays within the limits; no stack overflow. | save |
| P06 | Deeply nested lists, `[[[[[[...]]]]]]` past the depth limit | This formula goes over the depth limit of {value}. | save |
| P07 | A name longer than the formula length allows, `aaaa...a + 1` | This formula goes over the formula length limit of {value}. | save |
| P08 | A name that fits the formula but is very long, used 1000 times | The result, or the limit sentence for the node limit or the formula length; never a hang. | both |
| P09 | A number written with a very long run of digits, `1111...1` | This formula goes over the number magnitude limit of {value}. | save |
| P10 | A very long run of digits after the decimal point, `0.1111...1` | This formula goes over the decimal places limit of {value}. | save |
| P11 | A zero-width space inside a name, `tot<U+200B>al + 1` | The formula cannot be read: unexpected character U+200B at line 1, column 4. | save |
| P12 | A right-to-left override that makes the formula read backwards on screen, `<U+202E>1 + 2` | The formula cannot be read: unexpected character U+202E at line 1, column 1. | save |
| P13 | An emoji where a name belongs, `rate * 1` with the emoji in place of `rate` | The formula cannot be read: unexpected character at line 1, column 1; or the emoji is accepted as a name and the result is the same on every run. The choice is an open question. | save |
| P14 | Control characters, `1 +<U+0007> 2` and a lone carriage return | The formula cannot be read: unexpected control character at line 1, column 4. | save |
| P15 | A missing closing bracket, `sum(price, quantity` | The formula cannot be read: missing closing bracket at line 1, column 21. | save |
| P16 | A trailing operator, `price * quantity +` | The formula cannot be read: the formula ends after an operator at line 1, column 17. | save |
| P17 | An unfinished string, `"hello` | The formula cannot be read: text is not closed at line 1, column 1. | save |
| P18 | An unfinished escape inside a string, `"hello\` | The formula cannot be read: text ends inside an escape at line 1, column 7. | save |
| P19 | A formula on many lines where the fault is on the last, so line and column must be right | The formula cannot be read: {what is wrong} at line {last line}, column {m}. | save |
| P20 | A formula that is only blanks, or empty | The formula cannot be read: the formula is empty at line 1, column 1. | save |

## L: Limits

Each limit is hit on purpose. After every case the same engine evaluates a small good formula, `2 + 3`, and returns `5`. No goroutine and no memory is left behind.

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| L01 | Over depth, `-(-(-(-(...1...))))` one level past the depth limit | This formula goes over the depth limit of {value}. | save |
| L02 | One level exactly at the depth limit | The normal result. | both |
| L03 | Over nodes, `1 + 1 + 1 + ... + 1` with one node past the node limit | This formula goes over the node limit of {value}. | save |
| L04 | One node exactly at the node limit | The normal result. | both |
| L05 | Over cost, a formula that fits depth and nodes but whose work is past the cost limit, `sum(range(1, 100000)) * sum(range(1, 100000))` with a large enough range | This formula goes over the cost limit of {value}. | both |
| L06 | Over time, a formula under the cost limit on paper whose real work is past the time limit | This formula goes over the time limit of {value}. | run |
| L07 | After L01 to L06, a good formula on the same engine | The normal result, `5` for `2 + 3`. | run |
| L08 | After L05 and L06, the goroutine count and memory in use are read | Both are back to what they were before the case. | run |
| L09 | Over time, with the host's own deadline shorter than the time limit | The host's deadline wins; the sentence names the time limit or the host's deadline, and the engine returns at once. | run |
| L10 | Ten hostile formulas run at the same time on one engine | Each returns its own refusal sentence; a good formula started meanwhile returns its normal result. | run |

## C: Cost tricks

Each trick must be refused when saved. The save-time check might miss some, so each trick also has a run row: it must stop at the cost limit or the time limit, whichever is reached first.

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| C01 | A very large list, `range(1, 1000000000)` | This formula goes over the list length limit of {value}. | save |
| C02 | The same very large list, built from a value only known at run time, `range(1, quantity)` with `quantity` set to a huge number | This formula goes over the list length limit of {value}. | run |
| C03 | A list built from a list, `map(range(1, 1000), x -> range(1, 1000))` | This formula goes over the cost limit of {value}. | save |
| C04 | The same list built from a list, with sizes known only at run time | This formula goes over the cost limit of {value}, or the time limit, whichever is reached first. | run |
| C05 | A repeated aggregate, `sum(map(a, x -> sum(map(b, y -> sum(map(c, z -> 1))))))` | This formula goes over the cost limit of {value}. | save |
| C06 | The same repeated aggregate, with list sizes known only at run time | This formula goes over the cost limit of {value}, or the time limit, whichever is reached first. | run |
| C07 | A call chain between functions, `f1` calls `f2` calls `f3` and so on past the depth limit | This formula goes over the depth limit of {value}. | save |
| C08 | The same call chain, built so the save check cannot count it | This formula goes over the depth limit of {value}, or the cost limit, whichever is reached first. | run |
| C09 | A function that calls itself directly, `f(x) = f(x) + 1` | This formula calls itself: f calls f. | save |
| C10 | The same self-call, slipped past the save check | This formula goes over the depth limit of {value}, or the cost limit, whichever is reached first. | run |
| C11 | A function that calls itself through another, `a(x) = b(x)` and `b(x) = a(x)` | This formula calls itself: a calls b calls a. | save |
| C12 | The same cycle through two functions, slipped past the save check | This formula goes over the depth limit of {value}, or the cost limit, whichever is reached first. | run |
| C13 | A self-call that doubles each time, `f(x) = f(x - 1) + f(x - 1)` with a large `x` | This formula goes over the cost limit of {value}, or the time limit, whichever is reached first. | run |

## I: Size of inputs

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| I01 | Very long text in the formula, `"aaaa...a"` past the text length limit | This formula goes over the text length limit of {value}. | save |
| I02 | Very long text passed in as an input value | This formula goes over the text length limit of {value}. | run |
| I03 | Text grown by the formula, `repeat("ab", 1000000000)` | This formula goes over the text length limit of {value}. | both |
| I04 | A number with a million digits, `1111...1` | This formula goes over the number magnitude limit of {value}. | save |
| I05 | A number with a million digits passed in as an input value | This formula goes over the number magnitude limit of {value}. | run |
| I06 | An exponent of 10^9, `1e1000000000` | This formula goes over the number magnitude limit of {value}. | save |
| I07 | A power that grows past the limit while running, `2 ^ (2 ^ (2 ^ 10))` | This formula goes over the number magnitude limit of {value}, or the cost limit, whichever is reached first. | both |
| I08 | A product repeated until it is huge, `fold(range(1, 100000), 1, (a, x) -> a * x)` | This formula goes over the number magnitude limit of {value}, or the cost limit, whichever is reached first. | run |
| I09 | Many decimal places, `0.` followed by a million digits | This formula goes over the decimal places limit of {value}. | save |
| I10 | A division that never ends in decimals, `1 / 3` taken to extreme places, `round(1 / 3, 1000000)` | This formula goes over the decimal places limit of {value}. | both |
| I11 | A list of a million entries written out in the formula | This formula goes over the list length limit of {value}. | save |
| I12 | A list of a million entries passed in as an input value | This formula goes over the list length limit of {value}. | run |
| I13 | Two lists each at the list length limit, joined, `join(a, b)` | This formula goes over the list length limit of {value}. | run |

## O: No outside reach

A formula only computes from its own text and its inputs. The attempts below are names and syntax an attacker would try.

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| O01 | Read a file, `read_file("/etc/passwd")` | The function read_file is not known. | save |
| O02 | Read the clock, `now()` | The function now is not known. | save |
| O03 | A random number, `random()` | The function random is not known. | save |
| O04 | The network, `http_get("http://example.invalid/")` | The function http_get is not known. | save |
| O05 | An environment variable, `env("HOME")` | The function env is not known. | save |
| O06 | Run a command, `exec("ls")` | The function exec is not known. | save |
| O07 | Load more code, `import("other")` and `eval("1 + 1")` | The function import is not known. And the function eval is not known. | save |
| O08 | A path-like or url-like name used as a name, `/etc/passwd` and `file:///x` | The formula cannot be read: {what is wrong} at line 1, column 1. | save |
| O09 | An index past the end of a list, `[1, 2, 3][99]` | The list has 3 entries and there is no entry 99. | run |
| O10 | A negative index, `[1, 2, 3][-1]` | The list has 3 entries and there is no entry -1. | run |
| O11 | An index that is not a whole number, `[1, 2, 3][1.5]` | The list has 3 entries and there is no entry 1.5. | run |
| O12 | A huge repeat to force memory, `repeat([0], 1000000000)` | This formula goes over the list length limit of {value}. | both |
| O13 | A name that looks like an internal one, `__proto__`, `constructor`, `os` | The function {name} is not known, or the name is an ordinary input name with no value and the sentence says so. | save |
| O14 | Determinism: one formula with the same inputs run many times, on a fresh engine and on a busy one | The same result every run, character for character. | run |
| O15 | Determinism: a formula over a list that sorts entries with equal keys, `sort(rows, x -> x.group)` | The same order of the equal entries every run. | run |
| O16 | Determinism: a sum of many decimals added in the list's order, run in two processes | The same result in both. | run |

## B: Batch

A batch is a list of rows, each holding a formula and its inputs. Hostile rows must not spoil the good ones.

| Id | Attempt | Expected | Stage |
|---|---|---|---|
| B01 | One hostile row among 100 rows, at position 57 | The 99 good rows each return their normal value; row 57 returns its own refusal sentence in position 57; the batch returns without error; the batch time is bounded. | run |
| B02 | Every row hostile, each a different attempt from the groups above | Every row returns its own refusal sentence in its own position; the batch returns without error; the batch time is bounded. | run |
| B03 | The hostile row is first | Every later good row returns its normal value; row 1 returns its refusal sentence; the batch returns without error; the batch time is bounded. | run |
| B04 | The hostile row hits the time limit, and sits in the middle | The rows before and after return their normal values; the middle row returns This formula goes over the time limit of {value}.; the batch returns without error; the batch time is bounded. | run |
| B05 | Two hostile rows that hit the time limit, one after the other | Both return the time limit sentence in their own positions; every good row returns its normal value; the batch time is bounded. | run |

"Bounded" means the batch time is no more than the time limit rule chosen in the open questions below, with the good rows' own time added.

## Open questions

- Does the time limit apply to each formula, or to the whole batch?
- Can a refused formula be saved as a draft, or is it never stored?
- Is an emoji accepted as part of a name, or refused as an unexpected character?
- Is a function that calls itself always refused when saved, or only stopped by the limits when it runs?
- When a host's own deadline is shorter than the time limit, which sentence does the formula return?
- Is a name such as `__proto__` refused when saved, or treated as any other input name?

## Failures found by fuzzing

Each failure found by fuzzing becomes a new row in the matching group, with the next free id, and its input is kept as a permanent case. The run is described in [fuzz-run.md](../3.%20implementation/fuzz-run.md).

The owner reviews this page before the limits are set.
