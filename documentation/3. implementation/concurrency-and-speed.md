# Concurrency and speed: the recorded figures

The cases are on [concurrency-and-speed.md](../4.%20qa/concurrency-and-speed.md). This page records the target figures they carry, where each comes from, and what stays open. Every figure is proposed. The engine is the library's own (decision 23), and the figures come from [the engine choice page](../2.%20design/engine-choice.md) and from the spike measurements behind it ([engine-spike-results.md](engine-spike-results.md), `spike/results/`). All names and figures are invented.

## The machine and its noise

The spike ran on an Apple M3 Pro with 11 logical CPUs and 18 GB of memory, while other agents ran tests on it. The one-minute load average had a median of 13.3 in the first run and 6.0 in the second. The own engine's per-run figures moved up to 4.4 times between the two runs, its batch figures by at most 4.2 percent. A figure measured on this machine is never compared with a figure from another machine.

## One run, one compiled program, one thread (CS-20)

| Target | Proposed | Measured for the own engine, second run (first run) |
|---|---|---|
| Nanoseconds per run, median over the 19 formulas | at most 450 | 206 (406) |
| Nanoseconds per run, the slowest formula | at most 1,500 | 648 (2,843), formula 12 |
| Allocations per run, every formula | at most 25 | 25 (25), formula 12; median 4 |

The four classes of the cases, measured per run, second run (first run): formula 1 97 (135) ns with 2 allocations, formula 8 146 (185) ns with 4, formula 15 497 (589) ns with 24, formula 3 538 (798) ns with 24. The first run does not meet the slowest-formula target (formula 12 at 2,843 ns). The targets are for a machine with a load average near 6 or lower.

## A batch, per row (CS-21 and CS-22)

Measured on batches of 10,000 rows, second run (first run), nanoseconds per row.

| Formula | One worker: proposed | One worker: measured | 8 workers: proposed | 8 workers: measured |
|---|---|---|---|---|
| 8 | at most 225 | 141.5 (147.5) | at most 150 | 60.6 (67.7) |
| 3 | at most 725 | 474.8 (464.1) | at most 425 | 180.1 (204.3) |
| 15 | at most 675 | 433.8 (433.9) | at most 400 | 166.1 (193.5) |
| 1 | none | not run as a batch | none | not run as a batch |

One worker: 1.5 times the larger of the two runs, rounded up to a multiple of 25. 8 workers: 2 times the larger, rounded up to a multiple of 25. The speed-up with 8 workers was 2.3 to 2.6 times. The cases carry the per-row figures of the 10,000-row batch to batches of 1,000 and 100,000 rows. That is an assumption, not a measurement (open question 8 of the cases page).

## One compile (CS-23)

The engine choice page gives no compile target. The targets below are derived here with the same rule as the one-worker batch figures: 1.5 times the larger of the two runs, rounded up to a multiple of 25 ns. Allocations per compile were the same in both runs, so the target is the measured count rounded up to a multiple of 5.

| Formula | Measured, second run (first run), ns | Proposed, ns | Allocations measured | Allocations proposed |
|---|---|---|---|---|
| 1 | 1,630 (1,579) | at most 2,450 | 45 | at most 45 |
| 8 | 1,680 (1,643) | at most 2,525 | 53 | at most 55 |
| 15 | 1,898 (1,863) | at most 2,850 | 57 | at most 60 |
| 3 | 2,558 (2,453) | at most 3,850 | 69 | at most 70 |

## The margins

A figure over its target by more than 25 percent for time, or by more than 10 percent for bytes and allocations, fails the check. Both margins are proposed. The check compares against a baseline measured on the same machine in the same job, and refuses to judge when the load average is above half the CPU count.

## What stays open

- A re-run of the benchmarks on a quiet machine before the owner accepts the targets.
- The 1,000-row and 100,000-row batches and the batch of formula 1 were not measured.
- Whether the compile targets derived here are accepted.
- CS-14, the cancel of a batch, stays held: whether a host's own cancel is honoured is open under decision 19, and with it what "promptly" means without a clock.
- Whether the speed check runs on every change or only before a release, and which machine it runs on.
