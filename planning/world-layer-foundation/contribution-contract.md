# World contribution: close the arithmetic contract before construction

2026-11-05, source `40c14203`. Draft readiness only; World is not accepted or implemented.
Method: [predeclared check](contribution-contract-plan.md), [retained exact probe](contribution-contract-probe.py),
[machine-readable results](contribution-contract.json). No product or authored ruling changes.

## What the sources actually say

- **[V, source]** `design/05-mmo.md` §1 publishes
  `qualifying_output × impactModifier(active_population) × perFactionPercentile`.
  It describes normalization by the active share, per-faction asks/ranking, and
  identical pacing at200 versus20,000 concurrent players. Personal percentile
  rewards and global unlocks are separate axes; labeled NPC fallback remains required.
- **[V, source]** draft WL2 instead writes
  `raw_output × normalizer(active_playerbase_share) × diminishing(player_share)`.
  It calls the multiplier concave and promises identical pacing at200 versus200,000.
  It does not define those two share denominators or bind the design's per-faction
  percentile term explicitly. The different population examples are not themselves
  a contradiction; both examples require an executable definition of normalization.
- **[V, source]** `server/gameserver/world_source.go` populates Commons/population/epoch
  but constructs zero-valued Planet/Milestones fields. This is a static producer
  trace, not a new runtime population study. No current contribution loader or
  implemented World formula is tested here.

## Conditional result, not a universal impossibility claim

Assume the active fraction is1 in both populations, its normalizer is the same
positive value, and `player_share` means a player's raw output divided by total
raw output in the same interval. Let `d` be the same multiplier in each population.
For equal players producing1 each, identical positive total credit requires:

```
200 × d(1/200) = 200000 × d(1/200000)
d(1/200) / d(1/200000) = 1000
```

Now take one player producing1000 and199,000 producing1 each. Total raw output
is200,000, so their shares are exactly those same two endpoints. Their credit ratio is:

```
(1000 × d(1/200)) / (1 × d(1/200000)) = 1000000
```

Thus these two calibration endpoints already force **amplification**, independent
of concavity:1000× raw output becomes1,000,000× credit. It fails the predeclared
non-amplification interpretation `credit ratio <= raw-output ratio`. That is an
explicit interpretation of diminishing relative influence, not an adopted gameplay
threshold or a proof of what an unspecified percentile input would do.

The endpoint proof does not need the plan's additional concavity/domain-at-zero
assumptions. It requires the ratio-only definitions above. If active population
includes cardinality, share means a faction percentile, other context enters the
table, or "same pacing" has a different scaling population, the premise changes.
The result must not be cited as impossibility of population-normalized milestones.

## Controls and scope

All calculations use exact rational arithmetic and a unit positive budget;1 is a
choice of mathematical units, not proposed game balance. No float, table rounding,
Decimal implementation or live throughput is measured.

| Model/control | Uniform200 /20,000 /200,000 totals | 1000× output credit ratio | Result |
|---|---|---|---|
| Multiplier `d(s)=s` | 1 /1 /1 | 1,000,000 | Invariance holds; relative credit amplifies |
| Multiplier `d(s)=1-s` | 199 /19,999 /199,999 | `199000000/199999` (<1000) | Diminishing ratio holds; invariance fails, solo credit is0 |
| Concave final weight `g(s)=2s-s²`, divided by the peer-weight sum | 1 /1 /1 | `133000000/133333` (<1000) | Formal feasibility control with additional denominator/changed object |

The all-zero control is explicitly invalid, so vacuous invariance cannot pass.
The final-weight control shows that a different, fully defined contract can meet
these two modeled properties. It is **not selected or recommended as the game's
formula**. Its new peer sum, budget meaning, finite precision, privacy, aggregation
cost and published artifact semantics are unproved. It is not a fallback that an
implementation agent may quietly substitute for WL2.

## Required contract closure

Before World acceptance, its author and owner need to pin:

1. The actual active-population input: count, fraction, reference population,
   window/late settlement, and treatment of no players/NPCs. Name the scaling
   population used by the invariance criterion; don't tune a curve to one example.
2. Whether concavity describes the multiplier or final credited contribution,
   and the actual anti-domination property. "Concave" alone is not that property.
3. Where the design's **per-faction percentile** enters: cohort/global population,
   rank/tie/zero-output rules and its distinction from raw output share.
4. Whether one formula controls global progress and personal rank or these need
   separately specified axes. Preserve the intended dual rewards, fair small-player
   participation, low/zero-population fallback and published mathematics.

Then close the actual committed-output producer, immutable time/catalog/policy
inputs, exact curve grammar/rounding, reducer/retry ownership, ledger/replay,
throughput alarms, reward and public-consumer contracts as one buildable proposal.
This check supplies a constraint for that work, not implementation authority or
a decision in the owner's name. No existing design/RFC body is rewritten here.
