# Design: the first-hour payoff — branched endings that hand run 2 a head start

- **Status: RULED 2026-08-15.** Mechanism, data authority (epoch 8 + `curriculum` artifact) and
  wire arm are owner-ruled; the three copy texts are owner-adopted verbatim. Literals `A/B/R/S/G`
  are MEASURED and PROPOSED but **not yet ratified** — they ratify with the epoch-8 payoff batch.
  Resolves T01-C31 (branch), T01-C35 (run-2 accelerator) and T01-C36 (production authority + wire)
  as ONE mechanism.
- **Owner direction 2026-08-15:** absorb both into T01; the accelerator is a starter package
  derived from which ending you got.
- **Author:** Claude (design lane). Literals measured by Codex before ratification.

## The problem, stated once

Exiting gives nothing mechanical. `prestige.NewRunState` builds a zero-balance, zero-generator,
zero-upgrade Company; Reputation and Route Knowledge appear only in exit bookkeeping and feed no
multiplier. At the scripted first exit Reputation computes to level **0** (cube-root of a tiny
lifetime value against the `1e12` threshold) and the 25 Route Knowledge granted cannot afford the
50-cost hint. So run 2 is byte-identical to run 1, and design/02's prestige loop — the game's
central structure — has no payoff in the first hour.

## The mechanism (one new production behavior, not two)

**A run-start starter package**, declared as data, applied by `NewRunState`, selected by the
ending you earned. That single seam delivers the branch's meaning *and* the faster run 2.

| Ending | Predicate (unchanged from the adopted design) | Starter package | Why it is that |
|---|---|---|---|
| **Acquihire** | `GeneratorPurchasedTotal` ≥ A, ≥1 upgrade owned | **Seed capital** — `company.cash` grant of **S** | They bought the team. You got paid. |
| **Burnout** | cash < B × cheapest unowned generator price | **Partial rebuild** — **G** *generated* `generator.beige_tower` | Every death teaches you the map: you know what to build and some of it is already standing. |
| **Pivot** | default | **Contacts** — `upgrade.reply_all_macro` pre-owned | You were the company; what carried over is your own throughput. |

Doctrine compliance (§11b family 1) is the reason for the exact shapes: burnout's generators are
**generated, not purchased** — free, priceless, earning no per-count multipliers and paying no
cost — so the purchased/generated split is preserved rather than quietly violated. Acquihire
grants a resource, pivot grants an upgrade; neither touches the split at all.

## Why this satisfies AC1 honestly

`run2_garage_gate < garage_gate` becomes true **for a content-derived reason** rather than by RNG
luck: every branch starts run 2 strictly ahead on a real production input. The relation is then a
measurement of the mechanism, which is what the acceptance criterion was always for.

## Data authority (resolves T01-C36)

- **Epoch 7 is immutable history.** Its snapshot is committed at
  `sha256:6c7fab29…f789` and `TestEpochGuardRejectsArtifactAdditionAsHotfix` explicitly rejects
  adding an artifact to a minted epoch. Therefore this content ships as **epoch 8**, with a
  nineteenth artifact `curriculum` (`balance/curriculum/t0-t1.json`) carrying the branch
  predicates (`A`,`B`), the Route-Knowledge bonus (`R`), and the three starter packages.
- Design law 4 forbids the alternative: `A`/`B`/`R`/`S`/`G` are balance data in a hot-reloadable
  declarative file, **never Go constants**. The "versioned kernel constants" arm is rejected on
  that law, not on preference.
- No epoch-7 byte, hash, or ratification is touched. T01 mints twice: epoch 7 (the retuned
  economy, done) and epoch 8 (the first-hour payoff).

## Wire and replay (resolves the rest of T01-C36)

- `run_ended` widens **append-only** to v3 carrying `branch` (closed union:
  `acquihire`/`burnout`/`pivot`) and the applied starter package. Historical v2 events stay
  accepted and byte-identical; the client's exact-object decoder gains the arm rather than
  loosening.
- Terminal replay inputs freeze the branch decision alongside `selected_exit_type`; the branch is
  a pure function of persisted state and replays byte-identically.
- Copy keys are pinned per branch: `curriculum.first_failure.<branch>.{title,body}`, with the
  shared `…next_run` line unchanged. The three adopted copy texts bind to those keys verbatim.

## Acceptance

1. Each branch is reached by at least one ruled persona at the pinned seed; an unreachable branch
   is a finding, not a feature.
2. Each branch's run 2 crosses `gate.t0_to_t1` **strictly faster** than its run 1 at the same
   seed, and the report records the **per-branch savings distribution** (min/p50/max) so a
   regression is visible. *(Revised 2026-08-15, owner-adopted: the original text said "materially
   faster" without defining materiality, which acceptance could not test. Measured reality: the
   weakest pair is Reference/Pivot at 10,000 ms while pivot's p50 saving is 284,071 ms — the small
   margin is a tail, not the typical case, and pivot's starter is a pre-owned upgrade whose value
   no literal in `A/B/R/S/G` can enlarge.)*
3. Reputation is identical and full across branches (the ruled no-tax property); a mutation
   reducing it in any branch fails.
4. Removing the branch selector, or the starter-package application, must each fail a test.
5. Historical v2 `run_ended` events and pre-branch replay inputs still decode byte-identically.

## Open literals (measured, then owner-ratified)

`A`, `B` (branch predicates), `R` (Route-Knowledge bonus), `S` (seed capital), `G` (rebuilt
generators). Derived so that all three branches are reachable AND every same-seed run 2 is
strictly faster without trivialising the tier — Codex measures across the three personas and returns
the tuple with its pacing evidence. **Measured 2026-08-15 and proposed:** `A=200`, `B=2e0`,
`R=50`, `S=1e4`, `G=10`; all three branches reachable (64/16/17 of 97 runs), zero invariant
failures, every same-seed run 2 strictly faster.
