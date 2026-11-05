# World input: a committed command is not a production amount

2026-11-05, runtime source2cc892d0/kernel0.3.228. Draft readiness, not a
World implementation or adopted contribution policy. [Predeclared population](producer-boundary-plan.md),
[actual instrument](../../server/production/world_producer_research_test.go),
[retained results and source hashes](producer-boundary.json).

## Actual boundary observed

The declared Postgres service executes current `Service.Handle`, pinned to the
existing immutable epoch5 fixture: one generator producing1 cash/s, controlled
initial cash100 (1000 for the progress-saturated arm). This is not current-epoch
natural acquisition, player behavior or a population/pacing measurement. A separate
pure `Evaluate` on the restored-state clone supplies the isolated passive-stage
reference. It is **not** an observation of an already committed production event.

| Actual command | Isolated passive reference | Saved balance delta | New revision / event / intent / log / outbox rows |
|---|---:|---:|---|
| Buy one generator, online1s | 1 | -10.3 | 1 /1 /1 /1 /2 |
| One manual action, same timestamp | 0 | +1 | 1 /0 /1 /1 /1 |
| One manual action, online1s | 1 | +2 | 1 /0 /1 /1 /1 |
| Buy one generator, offline1s | 0.9 | -10.4 | 1 /1 /1 /1 /2 |
| Purchase, tier progress already at target | 1 | -10.3 | 1 /1 /1 /1 /2 |
| Unaffordable purchase after1s | 1 hypothetical | 0 | 0 /0 /1 /1 /1 |
| Injected revision-write failure after mutation | 1 hypothetical | 0 | 0 /0 /0 /0 /0 |

Results are canonical Decimal strings in JSON; the table uses ordinary decimal
notation for readability. Counts above are **new** rows; the JSON's count tuple
includes the one initial Company save revision. Applied commands bind their log
to a new revision. The rejected command has a durable receipt/log but no applied
revision; its full loaded save remains unchanged. The injected SQL exception
reaches the actual revision insertion, then leaves every measured durable family
and the complete save unchanged. It is not a simulated commit or host-crash test.

The successful purchase retry, an hour later with the same ID/body, returns the
exact original receipt and retains the entire loaded save and all five counts.
Through the existing EventProjector seam, both calls deliver the **same** event ID
and payload. Delivery is repeatable even though event creation is once-only. This
is measured by the follow-up declared after the first seven-row run; it is not
retroactively attributed to the earlier instrument.

## What follows, and what does not

- **[V, SQL/reference] Net balances are not passive production.** The purchase
  spends11.3 while earning1, and the manual action adds1 separately. Negative
  balance change can accompany positive production; a positive change can occur
  with zero elapsed production. These are normal existing mechanics, not defects.
- **[V, SQL/source] Generic events are not a complete production feed.** Manual
  operations here persist gains without generic events; the purchase event
  describes the purchase/cost, not an explicit accrued-output fact. The existing
  receipt outbox carries the combined command result, not a separate output amount.
  This does not imply every manual/reward gain should count toward a milestone.
- **[V, reference/source] Guild's progress delta is not output.** Current
  `accrualhook.Result` exposes the ledger receipt plus elapsed/production/banked
  time and `ProgressDeltaPPM`. `guild.AccrualHook` consumes the latter only for
  members. In the saturated arm the actual engine's reference delta is0ppm while
  passive production is1. The Guild hook itself is traced, not enabled/measured
  in this population. Its intended XP policy is not changed or declared defective.
- **[V, SQL] Durable intent/log existence is not an applied credit.** Rejected
  requests are deliberately recorded. Source `save.applyIntent` runs mutation
  before insertion/commit; a pure accrual result is provisional until the owned
  transaction commits. A future World callback cannot treat hook invocation as
  commit proof or add progress on every redelivery.
- **[V, source] Existing closed replay has more information than net receipts.**
  `ApplyLogged` uses pinned catalogs, prior state and replay inputs; its returned
  `LoggedTransition` currently has state/outcome/receipt/events/invariants, not an
  isolated output fact. This study does not prove reconstruction is impossible.
  For the purchase, its cost event plus the net amount already separates the two
  terms. A complete replay-derived producer remains an alternative to a new typed
  transaction-owned record; neither is built or selected here.

`databaseWorldSource.SampleWorld` still supplies zero Planet/Milestones fields.
That is a source inspection, not another executed World/transport observation.
Nothing in this study supplies a contribution ledger or modifies those fields.

## Contract work now required

Before adopting World, close one output producer across all relevant commit paths:
ordinary accrual/offline catch-up, exits, suppression, active rewards, minigame and
Garden credits. Explicitly choose which sources qualify; don't classify every
positive ledger entry as production. Bind resource, amount meaning (raw/capped/
credited), scope/run/Founder/faction, pinned catalog, production interval, and a
stable committed coordinate. State rejection/rollback/retry and replay/retention
behavior. A transaction-owned typed record or an explicitly replay-derived reader
must support recovery and deduplicate durable identity, not callback count.

The interval/population issue is separate: lazy settlement can describe past
production, so commit arrival alone does not determine its historical active
population or per-faction comparison window. This was not measured here. Preserve
[RP-775's normalization/concavity constraints](contribution-contract.md) and close
its denominators, percentile/tie rules and output semantics together. No formula,
resource qualification, new outbox schema, clock, balance value or owner choice is
adopted by this research. Full Clout/CI/rights/privacy/clean-host and nine-tier1.0
obligations remain unchanged.

## Reproduce and limitations

From the repository root, choose a **new** ignored output path:

```sh
docker compose -f compose.save-test.yml -f compose.ci-test.yml run --rm test go test -p 1 -run TestWorldOutputProducerIntegrationResearch -count=1 -race -v ./production -args -world-producer-output=/workspace/.cache/world-producer-new.json
```

All seven final arms pass on real Postgres16.15 with Linux-amd64 Go1.26.10,2.022s
including the race wrapper. The local Docker host is ARM64/emulated, not native
amd64 or a supported release-host rehearsal. Normal CI explicitly skips this
manual study; passing that skip is not evidence. Selecting an existing output
refuses before DB work and preserves the earlier artifact. Source hashes include
the final instrument, production/save authorities and the traced Guild/World
sources. No product/schema/data/CI or normative author-body bytes are changed.
