# The JSON forms of a formula tree and of a function definition

Built for ISS-78. A formula is stored as one line of text (decision 5). The studio's blocks, an AI tool and the host move the same formula as a tree, and a function's inputs, rules and checks as a definition. Both are JSON, and both read back to exactly the Go value the engine uses, so nothing is translated to or from another form. The shapes are in [formula-language.md](../2.%20design/formula-language.md) under "The tree" and "What a definition holds". The decisions are 5, 6 and 7 in [decisions.md](../1.%20requirements/decisions.md). The Go surface is in [formula.md](formula.md) and [function.md](function.md).

Status: the owner has not reviewed this page. Everything marked proposed is a starting point.

## What it holds

| File | What it is |
|---|---|
| `formula/tree_json.go` | `TreeToJSON` and `TreeFromJSON` |
| `function/definition_json.go` | `DefinitionToJSON`, `DefinitionFromJSON` and the code `bad-definition` |
| `internal/jsonform/` | the small shared parts: `Quote`, `Decode`, `Plain`, `WholeNumber` and `UnknownKeys` |

Imports: the standard library. No comments in the code.

## A tree

```go
func TreeToJSON(tree formula.Tree) []byte
func TreeFromJSON(data []byte) (formula.Tree, []formula.Problem)
```

`TreeToJSON` writes one line with no spaces. Each part is an object with a `kind`, and its fields are written in the order of the table on the language page. Text is written as it is: `<`, `>` and `&` are not escaped. An empty tree is written as `null`.

The formula `payment >= balance ? 0.00 : balance - payment` is written, in one line, as the tree on the language page:

```json
{"kind":"conditional","condition":{"kind":"binary","op":">=","left":{"kind":"name","name":"payment"},"right":{"kind":"name","name":"balance"}},"then":{"kind":"literal","type":"decimal","value":"0.00"},"else":{"kind":"binary","op":"-","left":{"kind":"name","name":"balance"},"right":{"kind":"name","name":"payment"}}}
```

A call writes both `library` and `version`, as `null` where it has none, and always `args`:

```json
{"kind":"call","library":"fin","name":"pmt","version":2,"args":[{"kind":"name","name":"a"}]}
{"kind":"call","library":null,"name":"round","version":null,"args":[{"kind":"name","name":"x"},{"kind":"literal","type":"integer","value":"2"}]}
```

`TreeFromJSON` reads the same shape and nothing else. It works in two steps, like Check.

1. **Structure.** Every part must be an object with a `kind`, every field of its kind must be there, and no other field may be. A field of the wrong kind of value is refused. If this step finds a problem, the tree is not built: the first return value is an empty `Tree` and the problems are those of this step, in tree order.
2. **Values.** A tree whose structure is fine goes through the same value checks as a tree read from text: a literal whose digits do not fit its type, a name that breaks the pattern of a name, an operator that is not in the language, a version outside 1 to 999999. These problems come back with the tree, so a block editor can show the tree and mark the blocks at fault.

A tree read from JSON is the same tree that `Parse` gives, so `Format` of it is the standard text and `Check` gives the same problems. A good tree gives an empty list that is not nil.

### The problems

Every problem is `bad-tree`, with the sentence `The part at path {path} is not a valid part: {what}.`, an empty `part` and the path of the part at fault. At most 20 are given, in tree order. The words `{what}` are those of the language page, and these are the ones this step uses:

| `{what}` | When |
|---|---|
| `the text is not a tree` | the data is not JSON, is empty, holds more than one value, is nested more than 10000 deep, or a part is not an object |
| `the field kind is missing` | a part has no `kind` |
| `the kind {kind} is not in the language` | the `kind` is none of the seven |
| `the field {field} is missing` | a field of the kind is not there |
| `the field {field} is not allowed` | a field is there that the kind does not have, in the order of its characters |
| `the value {value} is not a text` | a field that must be text holds something else (proposed wording) |
| `the value {value} is not a integer` | `version` is not a whole number. The wording keeps the page's `a {type}`. |
| `the value {value} is not a list` | `args` or `items` is not a list (proposed wording) |
| `the part has 0 children and needs 1 or more` | a list with no items (proposed wording) |
| `the operator {op} is not in the language`, `the value {value} is not a {type}`, `the name {name} is not a valid name` | the value checks of step 2 |

Example. The tree with an operator that is not in the language, from the language page, gives one problem:

```json
{"kind":"binary","op":"^","left":{"kind":"name","name":"balance"},"right":{"kind":"literal","type":"integer","value":"2"}}
```

`bad-tree`, path `[]`: The part at path [] is not a valid part: the operator ^ is not in the language.

A problem inside a call's second argument has the path `[1]`, and one in the else branch of a conditional, inside a `!`, has the path `[2, 0]`.

## A definition

```go
func DefinitionToJSON(definition function.Definition) []byte
func DefinitionFromJSON(data []byte) (function.Definition, []formula.Problem)
```

The keys are those of the language page. `DefinitionToJSON` always writes `inputs` and `checks`, and `rules` for each input. Each rule is an object with a `rule` field, then the parameters it uses, in this order, and none that it does not use:

| Rule | Parameters, in order |
|---|---|
| `required`, `no-negatives` | none |
| `min-items`, `max-items` | `items` |
| `range` | `min`, `max`, `min_exclusive`, `max_exclusive`. The two flags are written only when true. |
| `decimal-places` | `places` |
| `allowed-values` | `values` |

The worked definition of the language page (formula 1, balance after payment) is written, in one line, exactly as the page shows it.

`DefinitionFromJSON` reads the same shape. On reading, `inputs`, `checks` and each input's `rules` may be left out (proposed): a missing list is an empty one. A badly formed definition gives an empty `Definition` and its problems. A well formed one is given back whole, followed by the problems of `CheckDefinition` on it, so a definition that the save would refuse is still returned for the editor to show. A definition that is read and accepted gives an empty list that is not nil.

### The problems of a badly formed definition

The code is `bad-definition` (proposed). The sentence is `The definition at {place} is not valid: {what}.`, and for the whole text `The definition is not valid: {what}.`. The `part` is the place, counted from 1 like `checks.1` on the language page: `inputs.2`, `inputs.2.rules.1`, `checks.1`. At most 20 are given.

| `{what}` | When |
|---|---|
| `the text is not a definition` | the data is not JSON, is empty, holds more than one value, is nested too deep, or a part is not an object |
| `the field {field} is missing` | `name` or `type` of an input, `rule` of a rule, `formula` or `message` of a check |
| `the field {field} is not allowed` | a field the object does not have. For a rule these are the fields of the other rules. |
| `the value {value} is not a text` | a name, type, bound, formula, message or allowed value that is not text |
| `the value {value} is not a whole number` | `items` or `places` |
| `the value {value} is not true or false` | `min_exclusive` or `max_exclusive` |
| `the value {value} is not a list` | `inputs`, `checks`, `rules` or `values` |

A rule whose name is not one of the eight is read with the parameters of all the rules, and `CheckDefinition` refuses it with `bad-rule-parameter`. A parameter that is missing, such as `places` on `decimal-places`, is not a problem of the form: `CheckDefinition` refuses it with the sentence it already has.

## Tests

`formula/tree_json_test.go`, `function/definition_json_test.go` and `internal/jsonform/jsonform_test.go`.

- The nineteen example formulas of [formula-examples.md](../4.%20qa/formula-examples.md) are written, read back and compared with `reflect.DeepEqual`, written again to the same bytes, and checked.
- The eleven definitions D1 to D11 of [input-checks.md](../4.%20qa/input-checks.md) go the same way.
- The worked tree and the worked definition are compared with the language page's JSON, byte for byte after the page's spaces are taken out.
- 27 trees that are not formulas and 29 definitions that are badly formed are refused with the exact sentence and place.
- A tree or definition nested 30000 deep is refused with one problem and no crash. A list of 60 bad parts gives 20 problems, the first 20 in tree order.
- Copies of the code with a flag not written, or the value checks left out, fail the tests.

## What stays open

1. **Key names.** The tree keys are the language page's. The definition keys are the page's too; the parameter names of a rule and `min_exclusive` and `max_exclusive` follow the rules table. Nothing here is new, but the page does not say that a missing `rules`, `inputs` or `checks` is an empty list, nor that the flags are left out when false. Both are proposed.
2. **The code `bad-definition`**, the places `inputs.2.rules.1` and the wording of the problems of a badly formed definition are proposed. The language page gives codes for a definition that is refused on save, and none for one that cannot be read.
3. **The words of `bad-tree`** that are not in the language page's list: `the value {value} is not a text`, `the value {value} is not a list` and `the part has 0 children and needs 1 or more`. The page's set has `the part has {n} children and needs {m}`, which is not written for a list with no items.
4. **`a integer`.** The page's sentence is `the value {value} is not a {type}`, so `version` gives `is not a integer`. It reads badly. [formula.md](formula.md) already keeps the page's article for the same reason.
5. **A tree with a bad value.** It comes back with its problems. Whether the host may save such a tree as a draft is decision 12's question, not this page's.
6. **Duplicate keys.** If a JSON object repeats a key, the last one wins. Nothing here refuses it.
7. **Size.** Neither function limits the size of the data it is given. The host limits the body of a request. A tree that is read is then checked for depth, nodes and cost by `Check`.
