# RFC: Achievements Foundation (Clout deferred — see C1/C2 rulings)

- **Status:** accepted (C1–C10 ruled; SCOPE NARROWED to Achievements; implementing)
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/02 §6` (Clout — the achievement/influence axis; milk-equivalent + attention-economy play; feeds the multiplicative CloutStack), `design/02 §2c.3` + `§6b` (the Neopets possession/burn/provenance law + the Gaia one-mint law), `design/08` (influencer-culture satire)
- **Research:** `neopets-systems.md §5` (achievements-as-currency, the avatar-lending credit market — the possession-check hazard), `gaia-hyperinflation.md §4a` (posting-as-faucet, the single-mint discipline), `cookie-clicker.md` (achievements→milk→kittens, the highest-leverage idea)
- **Depends on:** Production + Run Genesis + Copy Pipeline (implemented); Meters Foundation
  (implementing — shares the strict registry pattern; Phase-A predicates do not yet consume meters)
- **Owner ruling honored:** breadth-first — the mint/decay/check MECHANICS and the achievement engine, not the individual achievements (those are content).
- **Planning:** `planning/clout-and-achievements-foundation/`

## Summary

This foundation implements achievements-as-score: permanent earned IDs and their exact,
non-spendable `achievement_score`. It deliberately does **not** own Clout. Clout's only mint is
future social activity, and lifetime Clout never enters production. The foundation preserves the
possession-check warning from Neopets and exports a neutral, typed observation seam for the future
run-local PR-Intern multiplier and Relevance Harness.

## Specification

### CA1 — The achievement engine (closed catalog family)

`balance/achievements/*.json` is strict schema v1. Rows are
`{id,condition_scope,condition,proof,score_grant,copy_key}`. `condition_scope` is `run|career` and
describes what the predicate may observe, never reset ownership. `condition` is the bounded closed
union from C6: `fact_present`, `counter_at_least`, `exit_count_at_least`,
`owns_generator_at_least`, or `all_of`. `proof` is exactly one of `provenance`, `burn`, or
`possession`; the loader enforces structural predicate/proof compatibility. `score_grant` is a
positive exact-safe integer balance literal. `copy_key` resolves through the implemented Copy
Pipeline before a production catalog may ship.

Achievements evaluate after the complete applied nonterminal transition and at the terminal
pre-settlement boundary. Rejected intents never evaluate. Definitions evaluate in byte order
against one pre-achievement snapshot; newly earned IDs commit simultaneously and emit
`achievement_earned.v1` in ID order. The Company run set latches immediately. Founder lifetime
ownership settles atomically at Exit, so achievements never mutate Founder scope during an
ordinary Company intent.

- **The check discipline (the Neopets law, ruling C4/§2c.3 made code):** every achievement
  declares how its condition is verified. `burn` = the earning transition consumed something unfakeable
  (the Kadoatery proof-of-burn — default for prestige achievements). `provenance` = derived from
  the immutable run log / event history (unforgeable by construction). `possession` = checks
  current holdings, and **requires an explicit `possession_justification` catalog field** because
  possession will be rented (the ALP credit-market lesson); the loader rejects a bare
  `possession` check. Phase-0 default: `provenance` (the run log is right there).

### CA2 — Achievement score and the neutral provider seam

`achievement_score` is an exact non-negative integer derived from the catalog grants of the
earned-ID set. It is not an economy resource, has no debit API, and never aliases `clout_lifetime`.
Save v16 persists Company `achievements_earned_run`/`achievement_score_run` and Founder
`achievements_earned_lifetime`/`achievement_score_lifetime`, with sorted-unique sets and derived-
score validation. The future PR-Intern owner may consume the typed run observation, but this
foundation registers no production contribution and therefore ships neutral.

### CA3 — Founder settlement and reset

The Company run set accumulates newly earned IDs. `ApplyExitTransaction` unions them into the
Founder lifetime set under the existing founder→company lock order, derives the lifetime score
from the pinned achievement catalog, records the settled IDs and score in the Exit receipt for
replay, and starts the next Company run with an empty run set/zero run score. The union is
idempotent by achievement ID. New Founder starts empty and deliberately inherits nothing.

### CA4 — Anti-gaming (the satire must not self-Goodhart)

Achievement conditions latch by lifetime ID; an oscillating condition cannot farm repeated
grants. The run set may not contain an ID already present in Founder carry. The Relevance Harness
receives typed earned observations; activation of its dead-achievement report remains downstream.

### CA5 — Identity, replay, and ownership boundaries

`achievements` is a strict epoch artifact. Production definitions require verified copy keys;
test fixtures may use a validated local copy registry. The artifact joins constants identity and
replay bundles in the mint that introduces content. Hook order is Purchasable Content → Meters →
Achievements → future Events. Achievement code does not import the future Clout transition owner,
and closed reward/effect registries expose neither Clout nor achievement score as a generic grant.

## Owner rulings on C1–C10 (2026-08-03) — TWO BINDING-LAW REVERSALS CORRECTED

Codex caught this RFC contradicting two binding design/02 §6 laws. Both catches are correct; the
RFC is restructured, not patched:

- **C1 — accepted: Clout's ONE mint is SOCIAL ACTIVITY, not achievements** (design/02 §6b, binding
  — I reversed it). Achievements mint a separate permanent non-spendable **`achievement_score`**
  (the true milk-equivalent). **Clout is removed from this RFC entirely** — the Feed/Social
  foundation owns Clout's single mint + decay. This RFC becomes **Achievements Foundation**.
- **C2 — accepted: lifetime Clout never enters production** (design/02 §6 carry rule, binding — my
  lifetime CloutStack recreated the exact snowball the carry rule removed). Removed. The milk-loop
  multiplier, IF kept, is a **run-local** contribution requiring owned PR-Intern content + current-
  run achievement score; no such content exists, so **this foundation ships the typed provider
  seam + vectors only, NEUTRAL in production**. Founder `clout_lifetime` stays reach-only, never
  imported into multiplier/production.
- **C3 — accepted:** `achievement_score` is a new non-negative int64 Founder field (not ppm, not
  the existing `CloutLifetime`); the earned-set is a new sorted-unique Founder field; both named
  in the save migration.
- **C4–C10 — accepted as proposed** (the check-discipline predicate grammar, the latch/idempotency
  proof, the run-vs-founder scope of the earned set, the achievement_score save shape, the
  relevance-harness dead-achievement hook, and the Company/Founder transaction boundary — earned-set
  and achievement_score are Founder-scope, so per Meters-C3 they accumulate as a Company pending
  delta settled at Exit OR the achievement engine writes founder scope only via the exit path;
  ruled: **achievement earning accumulates on Company state and the earned-set/score flush to the
  Founder stream at Exit**, the same multi-stream discipline, so ordinary intents stay Company-only).

## Acceptance criteria

1. Achievement engine: condition evaluation inside `ApplyLogged` is byte-identical Go/TS; earning
   latches, simultaneous IDs order by byte value, oscillation grants nothing, and
   `achievement_earned.v1` uses the exact registered payload.
2. Check law: loader rejects a bare `possession` check; a `provenance` achievement is verifiable
   from the run log alone (fixture); a `burn` achievement's condition consumed the declared sink.
3. Achievement score is structurally non-spendable and disjoint from economy resources; the
   foundation has no Clout writer/import and its production contribution seam is neutral.
4. Exit settlement unions IDs exactly once, records the settled result for replay, resets run
   state, and preserves Founder lifetime state across later Exits.
5. Save-v16 Company/Founder migration corpus, artifact/replay identity, sequential Go/TS fixtures,
   and the typed downstream Relevance observation registry all pass.

## Open questions

- Literal production achievement definitions and grants are content data. The foundation may land
  against discriminating fixtures; the artifact mint waits for owner-authored content.
- Social Clout mint/decay and the PR-Intern run-local multiplier belong to their successor RFCs.

### C11 — Save v16 and live evaluation depend on unresolved meter/save activation

The implementation plan treats literal achievement rows as mint-only debt, but v16 follows meter
save v15 and derives run/lifetime scores from the run-pinned achievement artifact. Current runs are
v14 and pinned to epochs containing neither `meters` nor `achievements`. There is no deterministic
way to write complete v16 state or evaluate achievements during those runs without reading
deploy-current artifacts. Meter C13 also leaves the v14→v15 activation boundary owner-gated, so
v16 cannot activate independently without skipping or redefining the save chain.

**Proposed contract:** use the same new-run activation boundary as Meter C13. Pre-foundation runs
remain on their pinned save semantics and execute neither Meters nor Achievements. Exit into the
first epoch containing both artifacts assembles complete v15 meter state and empty/derived v16
achievement state in one new-run transaction; subsequent ordinary writes require both pinned
artifacts. No achievement is retroactively earned from pre-activation history. The migration
corpus proves old-run replay through Exit, atomic v14→v16 new-run assembly, derived-score closure,
and rejection when either artifact is missing. If retroactive achievements are desired, owner must
define the immutable historical event/counter scan and artifact source for every old hash.

## Owner ruling on the activation boundary (Achievements C11, 2026-08-04) — ACCEPTED, generalized

Accepted as proposed, and lifted to a **reusable law** because every save-version foundation
hits it: **new-mechanic activation is NEW-RUN-BOUND at the first epoch whose PINNED catalog
carries the mechanic's artifact.** A run pinned to a pre-artifact epoch finishes under its pinned
save semantics and executes NO hook for the new mechanic (no retroactive gain — the same shape as
L2b version-drift and P6a pre-timer runs). The first Exit into an epoch containing the artifact(s)
assembles the new save version's complete state IN THE NEW-RUN TRANSACTION from that pinned
artifact (Standing axes from the Notoriety reseed, everything else from catalog initials);
subsequent ordinary writes then require the pinned artifact. Replay reads the run's PINNED
artifact bytes, never deploy-current — so activation never depends on deploy timing.

**Meters v15 and Achievements v16 activate ATOMICALLY** at the first Exit into an epoch containing
BOTH `meters` and `achievements` artifacts — one new-run transaction assembles v15 meter state +
empty/derived v16 achievement state together; no run is ever v15-with-meters-but-v14-achievements.
The migration corpus proves: old-run replay through Exit, atomic v14→v16 new-run assembly,
derived-score closure, and no retroactive earning.

This law is now the template for Pet Care and every subsequent save-version mechanic.

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

- 2026-08-03: created (draft) — achievements-as-currency + Clout.
- 2026-08-03: C1–C10 ruled — TWO binding-law reversals corrected: Clout's sole mint is social activity (deferred to Feed/Social, removed here) and lifetime Clout never enters production (CloutStack removed). Renamed Achievements Foundation; ships achievement_score + a neutral run-local provider seam; earned-set/score flush at Exit. Accepted.
- 2026-08-03: owner reconciled C1–C10; implementation unblocked. Stale Clout-era summary,
  specification, acceptance criteria, open questions, and blocked changelog were replaced before
  implementation.
- 2026-08-04: C11 records the save-v16 activation dependency: current v14 runs have neither the
  meter nor achievement artifact, so live state must activate at a ruled new-run boundary rather
  than from deploy-current catalogs.
