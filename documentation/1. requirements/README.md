# Requirements

Nothing is here yet beyond this README, apart from the decisions below.

## Decisions taken on 2026-11-17

| # | Decision |
|---|---|
| 1 | The repo is named `mwanachama-backend-compute` |
| 2 | The repo is public |
| 3 | It is a pure library: no database, no HTTP, no imports from any mwanachama repo |
| 4 | The first release tag comes later, once the first working version passes its tests |
| 5 | It computes values and does not schedule or run workloads |
| 6 | A published function is never edited: a fix ships as a new version and the old one stays |
