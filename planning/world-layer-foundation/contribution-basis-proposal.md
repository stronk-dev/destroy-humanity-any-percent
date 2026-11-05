# World contribution basis — a decision before construction

2026-11-05. **Proposal for an owner choice; not an implementation specification.**
World remains draft. Existing author/design bodies are unchanged. Technical review
and any later owner decision belong in the owning log.

## The decision

Propose two separately defined axes:

- **Shared milestone progress:** a published, population-normalized mean of
  capped qualifying output within each faction, combined using pinned weights.
- **Personal Influence payout:** contribution percentile within the player's
  faction, using separately specified percentile/tie/reward rules.

This changes the current draft's direction. WL2 presently multiplies raw output
by an undefined active-share normalizer and a concave player-share multiplier;
design/05 puts per-faction percentile inside that product. RP-775 demonstrates a
conditional amplification problem, not universal impossibility. This proposal
moves percentile into the personal axis and uses an explicit population count
in the shared axis. It is a candidate requiring Marco's adoption, not a finding
that silently rewrites either original document.

For one closed comparison window, let `x_i >= 0` be a Founder's qualifying output
in the milestone's published units, `C_f > 0` its visible per-Founder hardcap,
`N_f` the explicitly eligible Founder count in faction `f`, and `w_f >= 0` pinned
weights summing to1. Each eligible Founder is counted once, including admitted
zero-output members; Company Exits do not create extra Founders. Define the
**human component**:

```
G_f = 0                                        when N_f = 0
G_f = w_f * sum_i min(x_i, C_f) / N_f          otherwise
G_human = sum_f G_f
c_i = w_f * min(x_i, C_f) / N_f
```

Weights are not silently redistributed when a faction is empty. The count is an
input to be defined from the game's authoritative population/window contract,
not `ConnectionCount()`, registered accounts or whichever rows arrived first.
Offline eligibility and late settlement must be closed explicitly. Different
factions' asked-for outputs need published conversion to commensurate milestone
units; the formula does not invent those rates or qualifying resources.

## What the candidate provides

Replicating each faction's count and output distribution leaves the human total
identical. The total is positive exactly when some positive output belongs to a
positive-weight faction. Within the same faction/cap and positive faction weight,
credit grows with output until the hardcap. For `x_high >= x_low > 0`, the higher-
output to lower-output contribution ratio never exceeds `x_high / x_low`. A Founder's absolute
contribution is bounded by `w_f * C_f / N_f`. This uses an explicit hardcap,
not a hidden curve or a soft ceiling. The model comparison uses cap10 only as
mathematical units; no gameplay cap or weight is selected.

**It does not bound a Founder's fraction of the human total below100%.** With one
positive credited output and all other credited outputs zero, that Founder
supplies the entire positive human subtotal. Strict fractional anti-domination and meaningful lone-player
progress cannot both be inferred from a cap/mean. NPC contributions could change
the denominator but need their own honest, labeled, validated contract; this
proposal supplies none. The original broad “no single output dominates” claim
therefore remains unproved, rather than being renamed as a passing criterion.

Zero-population and all-zero human components are0, not completed milestones and
not evidence of meaningful invariance. The required low/zero-player NPC fallback
and solo viability remain mandatory before World acceptance. Personal payouts,
percentile ties, milestone failure/success and permanent unlocks remain separate
contracts, preserving the intended dual reward axes.

**Denominator dilution is a real tradeoff.** A single eligible Founder producing10
units has human progress10 in a weight1/cap10 faction. Admitting one eligible
zero-output Founder makes `N_f=2` and progress5 even though the first output did
not change. This is not replication of the same output distribution. Zero-weight
factions can likewise have positive raw output while contributing0. Eligibility
and window closure therefore need an explicit rule before final progress or
unlocks can be committed; an incomplete mean cannot later silently retract an
already-finalized milestone. This packet does not choose that rule.

## The producer contract this direction would require

[RP-777's actual SQL population](producer-boundary.md) shows why the producer
must observe the completed logged transition. [Current source traces](../../server/production/replay.go)
include ordinary accrual, offline catch-up and terminal Exit; Soul recovery calls
Evaluate and the time hooks but restores production-bearing outputs in
`ApplySuppressedLogged`. A raw evaluator callback is not an earning fact.
Minigame/Garden credits use their own certified transaction; they cannot become
ordinary passive output merely because they use Ledger.ApplyAccrual.

The complete proposal after a direction is chosen must close:

1. **Qualifying amounts.** Name each eligible source/resource and whether `x_i`
   uses requested, capped or actually credited output. Preserve existing offline,
   banked-time, suppression and reward rules. No arithmetic/partition fix is
   implicitly adopted; current Clout failures remain real.
2. **Population and windows.** Define interval membership, zero-output eligibility,
   faction changes and Exits, offline/late settlement, closure/watermarks, NPC
   treatment and missing evidence. A commit timestamp alone cannot attribute old
   lazy production. Do not prorate aggregate receipts across historical windows
   without an adopted attribution rule.
3. **Committed identity.** Select either transaction-owned typed output records
   or a fully replay-derived reader. Bind the pinned policy, Founder/Company/run,
   interval, semantic source, resource and amount to the committed coordinate.
   Discard previews, suppressed outputs, rejected and rolled-back work. Deduplicate
   immutable source identity rather than callback count. Cover ordinary commands,
   catch-up, both Exit arms and any explicitly qualified reward producer.
4. **Lifecycle and proof.** Close persistence, historical replay, archive/deletion/
   privacy/retention, recovery, failure handling and the public formula consumer.
   A new side table cannot quietly escape those existing obligations. Exact
   grammar/rounding and both runtime vectors precede content/balance adoption.

These are prerequisites, not a hidden list of already implemented behavior.
Selecting this basis would authorize finishing that coherent proposal and
reconciling the *direction* in design/05 and World WL2. It would not authorize
production code, minting, final mathematical bounds, copy, literals or release.

## Evidence and alternative

[Exact comparison](contribution-basis.json), [probe](contribution-basis-probe.py),
[predeclared populations/controls](contribution-basis-plan.md). This is exact
rational model evidence, not a Decimal/ppm implementation or real population.

Marco can instead retain the multiplicative direction. That still requires
explicit population/share inputs and final credited-contribution properties
consistent with RP-775's counterexample; it cannot proceed by guessing tables.
The choice is between product directions before engineering the full contract,
not between completing World and omitting it. The full nine-tier1.0 goal and all
World/solo/fairness/transparent-formula acceptance obligations remain.
