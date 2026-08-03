# RFC: Clout & Achievements Foundation

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §6` (Clout — the achievement/influence axis; milk-equivalent + attention-economy play; feeds the multiplicative CloutStack), `design/02 §2c.3` + `§6b` (the Neopets possession/burn/provenance law + the Gaia one-mint law), `design/08` (influencer-culture satire)
- **Research:** `neopets-systems.md §5` (achievements-as-currency, the avatar-lending credit market — the possession-check hazard), `gaia-hyperinflation.md §4a` (posting-as-faucet, the single-mint discipline), `cookie-clicker.md` (achievements→milk→kittens, the highest-leverage idea)
- **Depends on:** Production + Run Genesis (implemented — achievements evaluate inside `ApplyLogged`; Clout is Founder state); Meters Foundation (draft — shares the band/registry pattern)
- **Owner ruling honored:** breadth-first — the mint/decay/check MECHANICS and the achievement engine, not the individual achievements (those are content).
- **Planning:** `planning/clout-and-achievements-foundation/` (once implementing)

## Summary

Two coupled foundations the economy already assumes and neither exists: **achievements-as-currency**
(the milk/kittens loop — the design's stated highest-leverage mechanic) and **Clout** (the
Founder-persistent influence axis they feed). Both carry hard-won laws from the enclave research —
the one-mint discipline (Gaia) and the possession-check hazard (Neopets) — that must be structural
from day one or the satirical economy becomes the thing it satirizes.

## Specification

### CA1 — The achievement engine (closed catalog family)

`balance/achievements/*.json`: `{id, condition: closed predicate over committed facts (the routes
condition union + lifetime counters + meter bands + exit history), check: burn|provenance|
possession, clout_grant_ppm, scope: run|founder, copy_key}`. Achievements evaluate at accrual
boundaries inside `ApplyLogged` (state-derived, replay-deterministic, nothing new in
replay_inputs); earning emits `achievement_earned` (registry kind); the earned set is Founder
state (permanent — the milk model: they never un-earn), sorted unique.

- **The check discipline (the Neopets law, ruling C4/§2c.3 made code):** every achievement
  declares how its condition is verified. `burn` = the condition consumed something unfakeable
  (the Kadoatery proof-of-burn — default for prestige achievements). `provenance` = derived from
  the immutable run log / event history (unforgeable by construction). `possession` = checks
  current holdings, and **requires an explicit `possession_justification` catalog field** because
  possession will be rented (the ALP credit-market lesson); the loader rejects a bare
  `possession` check. Phase-0 default: `provenance` (the run log is right there).

### CA2 — Clout: the single-mint axis (the Gaia law made code)

Clout is Founder-persistent ppm. **It has exactly ONE mint: achievement grants** (Phase 0 —
attention-economy play adds mints LATER, each an RFC, each reviewed against this law). The loader/
lint asserts no other source writes Clout; a second faucet is a review-blocking defect by
construction (`design/02 §6b`: "any product that emits Clout is a second faucet the sinks weren't
sized for"). Clout feeds the multiplicative **CloutStack** (design/02 §6): a production-stack slot
whose factor = a declared curve over total Clout (`1 + f(clout_ppm)`, the kitten-multiplier shape,
quantize-once). Clout **persists across Exit** and is **not spendable** (it's an influence
*level*, not a currency — the meter not-spendable pattern; reach/PR surfaces read it, nothing
debits it).

### CA3 — Decay policy (ruled: none at Phase 0)

Clout does not decay in Phase 0 (achievements are permanent, so their Clout is permanent — the
milk model). If attention-economy play later adds *earned* Clout that should fade, that mint
declares its own decay; the achievement mint never decays. This is stated so a future decay
addition is a deliberate per-mint choice, not a global retrofit.

### CA4 — Anti-gaming (the satire must not self-Goodhart)

Achievement conditions are evaluated once and latch (earned-set membership); a condition that
oscillates cannot farm repeated grants. Clout grants are idempotent by achievement id in the
earned set. The relevance harness (drafted) treats achievements as purchasable-adjacent: an
achievement no persona ever earns is dead content flagged like any other.

## Acceptance criteria

1. Achievement engine: condition eval inside `ApplyLogged` byte-parity Go/TS; earning latches
   (re-crossing the condition grants nothing); `achievement_earned` evented; earned-set persists
   across Exit.
2. Check law: loader rejects a bare `possession` check; a `provenance` achievement is verifiable
   from the run log alone (fixture); a `burn` achievement's condition consumed the declared sink.
3. Clout single-mint: the lint/loader fails a seeded second Clout source; CloutStack factor
   golden vectors; Clout survives Exit and cannot be debited (seeded cost-wiring rejected).
4. Anti-gaming: an oscillating condition farms zero repeat grants; idempotent by id.
5. Save migration (Clout + earned-set fields) with corpus; relevance report flags a seeded
   never-earned achievement.

## Open questions

- The CloutStack curve `f` — balance data, harness-tuned; must stay legible (formula artifact).
- Attention-economy Clout mints (posting, viral moments, podcast circuit — `design/02 §6`,
  `design/08`) — each a successor RFC binding to CA2's single-mint law with its own decay.

## Acceptance blockers (Codex review, 2026-08-03)

The draft cannot be accepted because its two central mechanics reverse binding `design/02 §6`
decisions. The remaining gaps also cross the Company/Founder transaction boundary. Proposed
closures follow.

### C1 — The RFC assigns the sole Clout mint to the wrong system

`design/02 §6b` says Clout's exactly-one mint is **social activity through the declared decay
curve**, and explicitly warns that any product/event/reward emitting Clout is a second faucet.
CA2 instead makes achievement grants the sole mint and defers social activity as a later second
mint. Both cannot be true.

**Proposed contract:** achievements do not mint lifetime Clout. They mint a separate permanent,
non-spendable `achievement_score` (the milk-equivalent) used only by achievement-owned surfaces.
The future Feed/Social RFC owns Clout's one and only mint plus decay. If the owner intends to
reverse the Gaia ruling, amend `design/02 §6b` explicitly and re-size the downstream sink/decay
model before acceptance; do not silently redefine “one mint.”

### C2 — Lifetime Clout is forbidden from the production stack

`design/02 §6`'s carry rule states: “Clout never enters the production stack.” Only Clout earned
this run may key production-shaped PR Intern effects; lifetime Clout buys reach. CA2 feeds total
Founder-persistent Clout into a multiplicative CloutStack, recreating the exact prestige snowball
the carry ruling removed.

**Proposed contract:** remove lifetime CloutStack from this RFC. If achievement score keeps the
Cookie Clicker milk loop, its multiplier is a separately named **run-local** contribution whose
provider requires owned PR-Intern content and current-run achievement score; no such content
exists yet, so this foundation owns the typed provider seam and vectors only, neutral in production.
Founder `clout_lifetime` remains reach-only and never imports into `multiplier`/`production`.

### C3 — Units and existing save fields disagree

Clout already exists as non-negative Founder `int64 CloutLifetime`; the RFC calls it PPM and then
uses it as a multiplier input. There is no current-run Clout or achievement set field, and AC5
incorrectly calls existing Clout a new migration.

**Proposed contract:** keep `clout_lifetime` as exact integer units owned by the future social
mint; no PPM relabel. Define achievement score as exact integer units and literal per-achievement
grants in its own catalog (or fixed one point each). Save v16 adds Company
`achievements_earned_run` and `achievement_score_run`, plus Founder
`achievements_earned_lifetime` and `achievement_score_lifetime`; all sets sorted unique on wire,
all counts exact-safe integers. State validation derives score from the catalog+set and rejects
stored mismatch—or omit stored score entirely and derive it, one authority.

### C4 — Ordinary `ApplyLogged` cannot mutate Founder achievements

Like Soul, Founder state is read-only carry during ordinary Company intents. The service commits
one Company stream; it cannot atomically add a permanent Founder achievement/Clout grant from
inside `ApplyLogged`. Declaring a Founder earned set does not create a persistence path.

**Proposed contract:** achievements first latch in Company run state and emit Company events.
`ApplyExitTransaction` settles the complete run-earned delta into Founder lifetime state under its
existing founder→company lock order, idempotent by achievement ID, and starts the next Company run
with an empty run set. The Exit receipt records settled IDs/scores for replay. New Founder
deliberately does not inherit them; account deletion/anonymization follows existing Founder rules.
If immediate cross-run visibility before Exit is required, a new two-stream ordinary-intent
transaction is a separate prerequisite RFC.

### C5 — Scope semantics contradict permanent earning

CA1 permits `scope:run|founder` while saying the earned set is Founder-permanent and achievements
never un-earn. It does not say whether a run-scoped achievement can be earned again, whether its
score settles, or how IDs coexist across scopes.

**Proposed contract:** achievement definition scope means **condition scope**, not ownership:
`run` predicates inspect the current Company/run; `career` predicates inspect the immutable
Founder carry/lifetime counters. Every achievement ID is earned once per Founder lifetime and
settles through C4. Rename the union `condition_scope:run|career`; there is no resettable earned
achievement in Phase A.

### C6 — The condition/check union is not executable

“Routes condition union + lifetime counters + meter bands + exit history” does not declare exact
arms, field names, comparisons, nesting/depth, or catalog/version availability. Meters is itself
blocked. `check: burn|provenance|possession` is a label, not proof: nothing binds a burn to a sink
or provenance to an immutable event.

**Proposed contract:** version a closed predicate union with exact fields and catalog validation.
Phase A arms should be limited to committed boundaries already available without Meters:
`fact_present`, `counter_at_least`, `exit_count_at_least`, `owns_generator_at_least`, and
`all_of` with bounded depth/children. Each definition carries a proof union:
`{kind:"provenance",event_kinds:[...]}` requiring the predicate to reference state derived solely
from those immutable event kinds; `{kind:"burn",event_kind,resource_id,minimum}` requiring the
earning transition's same receipt/event batch to contain the declared debit; or
`{kind:"possession",justification_copy_key}` requiring an ownership predicate. Loader performs
structural compatibility checks between predicate and proof. Meter predicates append only after
Meters lands.

### C7 — Evaluation site and event order are underspecified

Accrual-boundary-only evaluation misses conditions created by the intent action after accrual;
evaluating before/after hooks can change which achievements latch. Multi-achievement event order,
rejections, Exit settlement, and events earned by the terminal action are undefined.

**Proposed contract:** evaluate achievements once after the full applied nonterminal transition
(accrual + action + registered hooks) and once at the terminal pre-settlement boundary after the
Exit action facts are committed to transition state. Rejected intents never evaluate. Evaluate
definitions in byte-order by ID against one pre-achievement snapshot, stage all newly earned IDs,
then commit simultaneously and emit `achievement_earned.v1` in ID order. Achievements cannot
trigger other achievements in the same transition except through the next applied intent. The
terminal set settles to Founder in the same Exit transaction.

### C8 — “Single mint” and not-spendable need structural package boundaries

A loader for achievement rows cannot prove another package never writes Clout, and putting
`spendable:false` in data repeats the constant-as-config problem from Meters. Future event/minigame
effect unions are the actual faucet risk.

**Proposed contract:** one owner package exports the sole Clout transition function to the future
social runtime; current achievement code cannot import it. Closed effect/reward registries do not
include Clout until that owner RFC extends them. Achievement score is a distinct non-resource ID
type, disjoint from economy resources at catalog composition, with no ledger debit API. Build-graph
tests enforce the imports and source-registry tests reject a seeded second Clout transition owner.
Docs state the enforceable claim—one code owner/registered source—not that a lint can prove all
future code behavior.

### C9 — Copy, catalog identity, replay, and kernel ownership are absent

Achievements carry `copy_key` but do not depend on the blocked Copy Pipeline. A new catalog family
inside `ApplyLogged` must be pinned in epoch/replay bundles and ordered in the hook chain. Event
schemas and cross-runtime fixtures are unnamed.

**Proposed contract:** Copy Pipeline is a hard dependency for production achievement definitions;
fixtures may use a local validated copy fixture only. `achievements` is a strict schema-v1 epoch
artifact and adding it is a mint. Replay bundles carry its bytes. Hook order appends Achievements
after Meters (when present) and before future Events; kernel version bumps. Register
`achievement_earned.v1` exact payload `{run_id,achievement_id,condition_scope,score_grant}` and add
sequential Go/TS fixtures for simultaneous, oscillating, terminal, and already-earned cases.

### C10 — The multiplier and relevance gates have no owner/content

The CloutStack curve is left open, no PR-Intern content exists, and Relevance Harness is
unimplemented. Golden multiplier vectors and a relevance failure therefore have no normative
formula or executable report.

**Proposed contract:** keep production contribution neutral in this foundation unless the owner
provides the exact run-local achievement-score formula and fixture content; otherwise move the
multiplier into T0–T1 Purchasable Content. This RFC exports achievement observations to the future
Relevance Harness but does not make an unavailable report an acceptance gate. AC5 requires the
observation fixture/registry now and names report activation as a downstream dependency.

## Changelog

- 2026-08-03: created (draft) — achievements-as-currency + Clout, with the one-mint and
  possession-check laws structural from the first line.
- 2026-08-03: Codex acceptance review recorded C1–C10. Implementation is blocked pending owner
  reconciliation of the Gaia single-mint law, lifetime-Clout carry rule, and Founder settlement
  boundary.
