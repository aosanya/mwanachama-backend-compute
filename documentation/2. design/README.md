# Design

One page for each part of the library that has been designed. The owner has not reviewed most of them: each page opens with its own status line, which is the place to read how far it is.

| Page | What it holds |
|---|---|
| [formula-language.md](formula-language.md) | The language the body of an agency function is written in: what a formula may contain, its one standard text, how it is read as a tree, and what is refused and with which words |
| [engine-choice.md](engine-choice.md) | The library runs on its own engine, written in this repo on the Go standard library; the limits and speed targets that go with the choice are proposed |
| [batch-evaluation.md](batch-evaluation.md) | One call that calculates a whole month for every member: one answer per row in the order sent, and a bad row never spoils the good ones |
| [library-versioning.md](library-versioning.md) | A published agency keeps the library version it published with until its builder upgrades; a fix ships as a new version and the old one stays untouched |

The test pages that go with these are listed in [the QA README](../4.%20qa/README.md).
