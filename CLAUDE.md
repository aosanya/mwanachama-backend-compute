# CLAUDE.md

Rules for working in this repo.

## Rules

1. This repo computes values. It does not schedule or run workloads.
2. It is a pure library: no database, no HTTP, and it imports no mwanachama repo.
3. The repo is public. Every example, fixture and test value is invented: no real names, emails or figures.
4. A published library function is never edited. A fix ships as a new version and the old one stays. Agencies pin to release tags. `v0.1.0` is cut when the first working version passes its tests.
5. No comments in code. A name, a type or a smaller function carries it. Anything that needs explaining goes in `documentation/` under its phase folder.
6. Tests: `make test` or `go test ./...`.

## Documentation layout

| Folder | Contents |
|---|---|
| `documentation/1. requirements` | Decisions and scope |
| `documentation/2. design` | Design records |
| `documentation/3. implementation` | What the repo holds and how to build it |
| `documentation/4. qa` | Test plans and QA records |

`documentation/` holds only these four folders and its `README.md`.
