# RFC: The Pitch (minigame content — THE TEMPLATE)

- **Status:** acceptance blocked on TP-C11–TP-C18 (TP-C1–TP-C10 ruled). v1 scope: engine + pinned catalog +
  internal integration; playability lands with the Minigame API & Surface successor). **This is the exemplar
  minigame-content RFC**: its structure (tenant row → engine contract → certified result → economy
  hooks → content-as-data) is the template the other minigame content RFCs replicate.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/03 §8/§9` (casino tier + Demo Disc Arcade — the two homes), `design/03 §12b`
  (the Fairness Law — ranked outcomes never scale with tier power), `design/02 §2.2` (the
  multiplicative production stack this game deliberately mirrors)
- **Depends on:** Minigame Platform Foundation (implemented + archived — C1–C40; this registers as a
  tenant), Fiscal Quarters (archived — the unlock sink), Soul Foundation (`soul_gate` field,
  archival-eligible). No new platform mechanics: this RFC adds a TENANT, an ENGINE, and CONTENT.
- **Planning:** `planning/minigame-the-pitch/` (once implementing)

## Summary

A solo score-multiplier card game: draft **metric cards**, play scoring hands, and stack
**growth-hack modifiers** whose effects multiply — `valuation = metrics × hype`, computed
**Decimal-exact** and deliberately pushed into the big-number regime the numeric core already
governs. Rounds are **funding-round targets**: beat the round's valuation threshold or the run ends.
It is turn-based, deterministic, PvE-against-a-threshold — the AI-fallback law is satisfied by
construction (solo by design), and the certified result is a pure recomputation of
`(seed, ordered choices)`. Mechanically this is the game's own multiplicative buff-stack exposed as
a card game — and computing it in exact `Decimal` where the inspiration genre lives with float
overflow is itself the satire.

## Motivation

The minigame platform ships with a test-only fixture tenant — zero real games. The research ranked
this shape first on every axis that matters (server-validatable, bottable, economy-hook, clock) and
uniquely on core-alignment: the score engine IS our production formula. It is the cheapest real
minigame we can ship and the best template, because it exercises every platform seam (tenant row,
certified result, faucet, offline-quality, unlock, soul_gate) with the least new engine surface.

Out of scope: multiplayer/ranked modes (solo only in v1; an async daily-seed board is a declared
successor), the card/hack CATALOGS beyond the launch set (content/data), and all numbers.

## Specification

### TP1 — The tenant row (the platform's C37 grammar, filled in)

One row in the pinned `minigames` artifact:
`{minigame_id: "pitch", engine_ref: "pitch", engine_version: 1, modes: ["solo"],
result_score_fact_ids: [...], scaling, payout, fallback, offline_quality, rating_policy,
unlock_condition}` — every sub-object per the already-ruled closed grammars. Specifics:
- **`scaling`:** the Fairness Law applies in full — NOTHING here scales with tier power. The only
  scaling input is `breadth` (unlocked card-set variants), never a power stat. Loader-rejected
  otherwise (the platform already enforces this).
- **`fallback`:** `solo` (the closed fallback row for solo-by-design games — no bot exists because
  no opponent exists; the threshold is the opponent).
- **`soul_gate`: `human_hobby`** — the founder's hobby; locks at near-zero Soul per SB15.
- **`unlock_condition`:** the Tier-5 casino home, unlocked by the purchased **Fiscal
  Investor-Confidence unlock** via the `fiscal_unlock {unlock_id}` resolver arm (TP-C6 — the
  composed server resolver reads the pinned Founder fiscal flag; the unlock_id registers in the
  Fiscal artifact). The Demo Disc Arcade variant is REMOVED from v1 (a successor with the arcade
  content RFC).
- **`rating_policy`:** a real neutral rating-season row with certified `rating_delta: null` (per
  C40 — Elo/count provably unchanged; resolution + offline-quality still commit). There is no
  `none` arm in the loader grammar (TP-C5).

### TP2 — The engine contract (deterministic, Decimal-exact)

A new engine package (`server/pitch` + `client/src/pitch`) behind the platform's engine seam:
- **A run** = `(seed, ordered choices)`. The seed derives from the platform's session identity via
  the established substream discipline (a named `pitch.run.v1` substream). The daily-seed variant is
  REMOVED from v1 (TP-C7 — the async-snapshot successor owns the calendar seed and its board).
- **The loop (structure ruled; all magnitudes are catalog data):** each round presents a drafted
  hand of `metric_card`s (each carries a base `metric` value) and the player's slotted
  `growth_hack`s (each a modifier in a **closed effect union**: flat metric adds, per-card
  multipliers, hand-shape multipliers, and hack-order interactions). A played hand scores
  `valuation = Σ(metrics after adds) × Π(hack multipliers)` — **computed in `Decimal`, never
  float** (the big-number law; scores are EXPECTED to enter the `break_infinity` regime — that is
  the point and the joke). Round N's `funding_target` (a Decimal threshold from the catalog curve)
  must be met within the round's hand budget or the run ends. Between rounds: a shop/draft step
  (spend run-local score-currency on hacks/cards — run-local, NEVER company resources).
- **Hardcaps, visibly:** hack slots, hand size, and rounds are visible hardcaps with reason keys.
  The score itself is uncapped (the numeric core governs it — watching the number leave the solar
  system is the reward).
- **The certified result** = the platform's integer-fact grammar (TP-C1): `pitch.final_round` (the
  payout/quality fact) + `pitch.best_hand_exponent` (display-only, hardcapped); the full Decimal
  valuation lives in the terminal ENGINE SNAPSHOT, which the platform byte-compares — recomputable
  byte-for-byte from `(seed, choices)`; `engine_version` ("1.0.0") pins the calc;
  the platform's resolve composer (C37–C40) owns everything after certification. Go/TS byte-parity
  golden vectors over the scoring math, including hack-order interactions and the
  big-number boundary.

### TP3 — Economy hooks (all via the platform, nothing new)

Payout = the certified score through the platform's grade curve + **faucet governor** (daily cap,
forfeit-with-reason above it — already shipped). `offline_quality` charges/decays per the C34
ladder. No Clout (the platform's ruled no-Clout law). Numbers (grade thresholds, payout rates,
funding-target curve, card/hack magnitudes) are balance data.

### TP4 — Content scope (the deckbuilder discipline)

Launch catalog: **small and flat** — the research's content-obligation warning is adopted as a
constraint: ≥~12 `metric_card`s and ≥~8 `growth_hack`s, each individually flat and legible, rather
than a large pool (depth comes from hack INTERACTIONS, which the closed effect union makes
enumerable and testable). Card/hack rows are exact-key catalog data with copy keys through the copy
pipeline (flavor lives in data; `pitch`/`metric_card`/`growth_hack`/`valuation` are the mechanical
names). Dead content is gated by the Pitch-owned content gate (TP-C9): every row must be reachable in
seeded generation AND affect at least one declared golden scenario; interaction arms require
pairwise fixtures. (The production relevance harness does NOT measure engine content — a
statistical strategy-relevance harness is a named successor.)

### TP5 — The template notes (for the ~10 RFCs that replicate this)

The reusable skeleton this RFC establishes: ① one tenant row filling the C37 grammar (never new
platform mechanics); ② one engine package with a versioned pure certified-result function over
`(seed, choices)`; ③ economy hooks exclusively through the shipped platform seams; ④ content as
exact-key catalog data with copy-pipeline keys; ⑤ Fairness-Law scaling declaration; ⑥ soul_gate
classification; ⑦ numbers-as-data. A content RFC that needs anything outside this skeleton is
touching platform territory and must say so explicitly.

## Deviations from design

None — `design/03`'s casino/arcade homes, clock taxonomy (session-skill), and the Fairness Law are
followed; mechanical naming per the naming law.

## Acceptance criteria

1. Tenant row loads under the pinned artifact in both runtimes; Fairness-Law loader rejection
   proven (a power-stat scaling input fails); soul_gate honored (near-zero Soul rejects start).
2. Engine determinism: `(seed, choices)` recomputes the certified result byte-identically Go/TS;
   golden vectors cover scoring incl. hack interactions and a big-number-regime hand.
3. Full platform path: create→play→resolve→payout for a real run through the shipped composer;
   faucet cap forfeits with reason; offline_quality charges; unlock via the Fiscal sink proven.
4. Content: launch catalog rows exact-key validated; copy keys resolve; relevance floors green once
   the production baseline exists.
5. No Company resource enters the run; no run-local currency leaves it (grep + test).

## Open questions

- The async daily-seed leaderboard variant (same seed, compare certified scores) — a declared
  successor once the platform's async_snapshot mode has a first consumer.
- Which of the four researched turn-based/snapshot surfaces ship (the redundancy watch) — a
  `design/03` roster decision owed before the NEXT combat-shaped minigame RFC; The Pitch is not a
  combat surface and does not wait on it.

## Acceptance-review blockers (Codex, 2026-08-07)

The concept fits the platform, but the draft cannot yet serve as an executable template. Several
claimed seams do not exist in the archived platform, and the engine/content byte contracts are not
enumerated. Implementing through these gaps would either add platform mechanics under a content RFC
or make deploy-current content part of a pinned replay.

### TP-C1 — The certified Decimal result cannot cross the platform result boundary

The shipped `minigame.Result` is exactly `{outcome,rating_delta,score_facts[]}`, and every score fact
is an exact signed `int64` within the JavaScript-safe domain. It cannot carry
`best_hand_valuation: canonical Decimal string`; payout also selects one nonnegative integer fact.
TP2/AC2 require the Decimal string in the certified result while TP5 forbids a platform change.

**Proposed contract:** keep the full Decimal valuation in the terminal engine snapshot (which the
platform already byte-compares), and enumerate bounded integer result facts separately. Recommended:
`pitch.final_round` is the payout/quality fact and `pitch.best_hand_exponent` is display/analytics
only, with an exact exponent derivation and hardcap. If the Decimal string must live in `Result`,
declare a Minigame Platform Result-v2 successor instead; The Pitch cannot add it invisibly.

### TP-C2 — Versioned content has no pinned tenant input

The tenant receives only `(mode,seed,revision,snapshot,command,scaling_inputs)`. The registry is keyed
only by `(engine_ref,engine_version)`, and the replay bundle has no `pitch` artifact. A tenant cannot
read a hot-reloadable card/hack/target catalog without ambient deploy-current state, while putting
rows in the existing minigame definition violates its exact schema.

**Proposed contract:** choose one replay-safe owner. Recommended: add a separately hash-pinned
`pitch` artifact to `CatalogBundle`, construct a constants-hash-aware tenant resolver, and freeze the
exact content hash/engine-content version in session genesis. An alternative is compiling the entire
launch set into engine version `1.0.0`, but that explicitly deviates from balance-data law 4 and
requires owner approval. No unpinned process-global catalog is acceptable.

### TP-C3 — The engine transition grammar is directional, not executable

There are no exact genesis/snapshot/command schemas, phase/state machine, rejection taxonomy,
shuffle/draw algorithm, shop transaction, hand-budget transition, target lookup, or terminal rule.
“Ordered choices” cannot be validated without naming each choice arm and its legal phase.

**Proposed contract:** enumerate command arms (`draft_card`, `buy_hack`, `slot_hack`, `play_hand`,
`end_shop`, or the owner-selected closed set), the exact snapshot keys, phase enum, sorted collection
rules, SplitMix64 substream labels/counters, terminal outcomes, and sorted tenant error taxonomy.
Every applied command advances one revision; illegal phase/index/insufficient run-currency rejects
without mutation. The result is emitted by exactly one terminal transition.

### TP-C4 — The effect union and Decimal operation order are not byte contracts

“Flat adds, per-card multipliers, hand-shape multipliers, and hack-order interactions” does not name
row keys, target selectors, rational/Decimal factor grammar, applicability, ordering, or quantization.
Go and TypeScript could satisfy the prose with different scores.

**Proposed contract:** enumerate every effect arm with exact keys and one published evaluation order:
card base → flat adds → per-card factors → hand-shape factors → ordered hack interactions → one
canonical quantize. Factors use RFC-0001 canonical Decimal strings; all tie/order behavior is bytewise.
Define the target-curve row and run-currency debit arithmetic with the same precision rules.

### TP-C5 — The tenant row is incomplete and `rating_policy: none` is not legal grammar

The loader has no `none` rating arm. Every definition needs literal score-fact IDs, a six-key payout
row, offline-quality row, valid rating-season row, exact scaling row(s), unlock row, and semver engine
identity. `engine_version: 1`, `[...]`, “economy-resource credits,” and “breadth” do not load.

**Proposed contract:** provide one complete literal schema-v3 definition and descriptor. Use
`engine_version:"1.0.0"`, `fallback:{kind:"solo"}`, nil certified `rating_delta`, and a real neutral
rating row whose state remains unchanged. Name the credited resource, payout/quality score fact,
automation destination, cap reason, breadth destination/source, season member, and every schema/error
ID. Balance magnitudes may remain provisional catalog values; structural IDs may not.

### TP-C6 — The Fiscal unlock is parsed but not enforced, and the two homes conflict

Production start checks Soul but never evaluates `Definition.Unlock`; no resolver maps
`fiscal_unlocks` to the generic `fact_equals` row. One row also cannot be both the free T0–T1 Demo
Disc tutorial and the purchased Tier-5 casino unlock.

**Proposed contract:** decide the product shape. Recommended: one free tutorial definition
`pitch.demo` and one full `pitch` definition, or explicitly make the demo a non-platform tutorial
inside the full game's client. Add a composed server unlock resolver that reads the pinned Founder
Fiscal flag (never a client fact), register its exact `unlock_id` in the Fiscal artifact, and prove
start rejects before purchase and succeeds after it. This is a named platform/composition amendment,
not tenant-engine code.

### TP-C7 — Daily seed is simultaneously out of scope and an acceptance criterion

TP2 introduces and publishes a daily seed, AC2 requires it in the wire, but TP1 ships only `solo` and
the Open Questions defer the daily/async variant. No calendar input exists in the tenant boundary.

**Proposed contract:** remove daily-seed behavior and AC text from v1. The solo session continues to
use the server-authored platform seed and `pitch.run.v1` substream. A later async-snapshot RFC owns
the calendar seed, publication, board, and fairness rules.

### TP-C8 — The launch content set is not a catalog

`≥~12` and `≥~8` are neither exact counts nor literal rows. No IDs, copy keys, effects, rarity/draft
weights, prices, or interaction declarations exist, so an engine fixture and loader cannot be built.

**Proposed contract:** check in the complete byte-sorted launch rows as provisional balance data.
Every card and hack has exact keys, copy keys, price/draft policy, and one declared effect arm. State
the exact launch count and require every referenced interaction partner to exist.

### TP-C9 — Relevance Harness does not measure minigame cards or hacks

The shipped harness ablates production purchasables through `SimulateAdvance`; it has no Pitch
engine, card-pool, draft, or hack-interaction boundary. TP4's “dead content fails CI” claim is false.

**Proposed contract:** replace that claim with a Pitch-owned exhaustive/enumerated content gate:
every row must be reachable in seeded generation and must affect at least one declared golden
scenario; interaction arms require pairwise fixtures. A statistical strategy-relevance harness is a
successor after real balance data exists, not an invented reuse of the production harness.

### TP-C10 — No player-facing command/surface reaches the platform

The authoritative platform and tenant registry are internal services; Game UI explicitly excludes
minigame surfaces. This draft requires create/play/resolve and a client engine but names neither an
authenticated coordinator API nor a surface/component contract.

**Proposed contract:** either scope this RFC honestly to engine+catalog+internal integration and
defer “playable” UI, or add dependencies/contracts for the minigame API and one UI surface: exact
create/play/resolve request/response schemas, authenticated Founder authority, reconnect snapshot,
command dispatcher, and terminal rendering. The exemplar should make this boundary explicit because
every later minigame will inherit it.

## Changelog

- 2026-08-07: created (draft) — Wave-B opener; the exemplar minigame-content RFC.
- 2026-08-07: Codex acceptance review filed TP-C1–TP-C10; implementation blocked pending owner rulings.

## Owner rulings on TP-C1–TP-C10 (2026-08-07)

All accepted; product decisions on C6/C10. Body reconciliations noted inline.

- **TP-C1 — accepted.** The Decimal valuation lives in the terminal ENGINE SNAPSHOT (already
  byte-compared); the `Result` carries bounded integer facts only: `pitch.final_round` is the
  payout/quality fact; `pitch.best_hand_exponent` (exact derivation: the canonical Decimal's base-10
  exponent, hardcapped with a reason key) is display/analytics only. NO platform Result-v2.
- **TP-C2 — accepted (recommended arm).** A separately hash-pinned `pitch` artifact joins
  `CatalogBundle`; a constants-hash-aware tenant resolver; the exact content hash + engine-content
  version frozen into session genesis. Law 4 (balance-data, hot-reloadable) preserved; no
  process-global catalog, no compiled-in content.
- **TP-C3 — accepted.** Closed command arms `draft_card | buy_hack | slot_hack | play_hand |
  end_shop`; exact genesis/snapshot keys; a phase enum; sorted collection rules; named SplitMix64
  substream labels + counters for shuffle/draw; one terminal transition emits the result; sorted
  tenant error taxonomy; every applied command advances one revision; illegal phase/index/insufficient
  run-currency rejects without mutation.
- **TP-C4 — accepted.** Every effect arm enumerated with exact keys and ONE published evaluation
  order: card base → flat adds → per-card factors → hand-shape factors → ordered hack interactions →
  one canonical quantize. Factors are RFC-0001 canonical Decimal strings; all tie/order behavior
  bytewise; the funding-target curve row and run-currency arithmetic use the same precision rules.
- **TP-C5 — accepted.** One complete LITERAL schema-v3 definition ships in this RFC's
  implementation: `engine_version: "1.0.0"`, `fallback: {kind:"solo"}`, nil certified `rating_delta`
  with a real neutral rating-season row (state provably unchanged), literal score-fact IDs, the
  six-key payout row, offline-quality row, exact scaling row (`breadth` with named
  destination/source), unlock row, and every schema/error ID. Structural IDs literal; magnitudes
  provisional catalog values.
- **TP-C6 — RULED (product shape): ONE platform definition.** `pitch` is the Tier-5 casino game,
  unlocked by the purchased **Fiscal unlock** via a new composed server unlock resolver: a
  `fiscal_unlock {unlock_id}` arm in the platform's unlock grammar, resolving the pinned Founder
  fiscal flag server-side (never a client fact), with the `unlock_id` registered in the Fiscal
  artifact — **a named platform/composition amendment, explicitly in this RFC's scope** (the
  exemplar establishes the seam). **The Demo Disc Arcade appearance is REMOVED from v1** (my
  draft's overreach — the tutorial-arcade variant becomes a successor when the arcade content RFC
  exists). Start rejects before purchase, succeeds after — test-proven.
- **TP-C7 — accepted; daily seed REMOVED from v1** (body + AC reconciled). Solo sessions use the
  server-authored platform seed + `pitch.run.v1`. The async-snapshot successor owns the calendar
  seed, its publication, the board, and fairness.
- **TP-C8 — accepted.** The complete byte-sorted launch rows are checked in as provisional balance
  data: **exactly 12 `metric_card`s and 8 `growth_hack`s**, each with exact keys, copy keys,
  price/draft policy, one declared effect arm; every referenced interaction partner must exist
  (loader-checked).
- **TP-C9 — accepted; the relevance-harness claim is RETRACTED** (it was false — the harness
  ablates production purchasables, not engine content). Replaced with a Pitch-owned content gate:
  every row reachable in seeded generation AND affecting ≥1 declared golden scenario; interaction
  arms require pairwise fixtures. A statistical strategy-relevance harness is a named successor.
- **TP-C10 — RULED (honest scope): v1 = engine + pinned catalog + internal platform integration**
  ("playable from Go", proven by the composed integration test). The authenticated minigame API +
  the UI surface are a NAMED DEPENDENCY — the **Minigame API & Surface** successor (an API-Foundation
  amendment + a Game-UI consumer), which every later minigame inherits; this template declares that
  boundary explicitly instead of pretending it away. Player-facing playability lands there.

Reconciliations applied to the body: daily seed removed (TP2/AC2), arcade home removed (TP1),
relevance claim replaced (TP4), Result grammar corrected (TP2/AC).

## Implementation blockers (Codex, 2026-08-07)

The first ruling round settled every product fork, but it did not supply the byte contracts it says
are exact. The normative body also retains several original placeholders. This narrower bounce asks
only for the literals and state-machine rules needed to implement the chosen product shape.

### TP-C11 — The normative body still contradicts the accepted rulings

TP1 still specifies `engine_version: 1`, `result_score_fact_ids: [...]`, and an incomplete object;
TP4 still says `≥~12` and `≥~8`; the dependency/header text still says “No new platform mechanics”
despite the ruled Pitch artifact and Fiscal unlock amendment. A reader cannot tell whether the body
or appended rulings govern.

**Proposed contract:** reconcile TP1–TP5 and AC1–AC5 in the ruling edit itself. Replace every
placeholder and approximate count with the final literal contract, and describe the two platform
amendments explicitly. Historical blocker text may retain the old wording.

### TP-C12 — The exact engine snapshot and command wire remain absent

TP-C3 accepted a closed command set, but no command payload keys, snapshot keys, phase transitions,
collection representation, or error IDs were supplied. Even genesis is undefined. Go and TypeScript
cannot independently implement byte-identical state from the current prose.

**Proposed contract:** add literal JSON examples or tables for genesis and every snapshot phase;
define exact payload keys for all five commands, legal source/target phases, revision/counter effects,
terminal success/failure rules, and the complete sorted rejection enum. State whether indexes or IDs
select cards/hacks and how a shop offer is represented and consumed.

### TP-C13 — The effect and content row unions are still unnamed

TP-C4 accepted four effect families without specifying their discriminator names or exact fields.
TP-C8 requires 12+8 literal rows, but none are present. Prices, weights, base values, interaction
partners, hand shapes, funding targets, hand budgets, slot caps, and reason/copy keys remain choices.

**Proposed contract:** enumerate exact schemas for `metric_card`, `growth_hack`, each effect arm,
funding-round rows, and global policy; then include all 20 byte-sorted launch rows and the complete
funding curve. Values may be marked provisional, but they must be literal owner-approved bytes.

### TP-C14 — The pinned Pitch artifact/session identity amendment is not specified

The platform session stores the minigame constants hash and genesis, but has no separately named
Pitch content hash/version. `CatalogBundle` and epoch artifact authority likewise have no `pitch`
member. “Freeze it in genesis” does not define which fields, hash algorithm, loader, or consistency
checks own that identity.

**Proposed contract:** specify artifact name/path/schema version; canonical hash construction; the
new `CatalogBundle.Pitch` type; exact genesis identity keys; start-time equality checks among epoch,
bundle, definition, tenant, and genesis; replay lookup; and the migration/immutability rule if any
session columns change. Prefer storing content identity in exact genesis keys unless a queryable DB
column has a named consumer.

### TP-C15 — The Fiscal unlock grammar and state predicate are incomplete

The current platform grammar is only `always | fact_equals`; the current Fiscal catalog has sorted
`{unlock_id,cost}` rows and Founder state stores unlocked IDs. The ruling names
`fiscal_unlock {unlock_id}` but not its exact JSON keys, the literal Pitch unlock ID/cost, or the
start-time predicate and error.

**Proposed contract:** add the exact third unlock arm and literal Fiscal row. Recommended wire:
`{"kind":"fiscal_unlock","unlock_id":"minigame.pitch"}`; start resolves the pinned Founder state
and requires that ID in its byte-sorted unlocked set, otherwise returns the existing typed
not-eligible category with a named detail. Add Go/TS loader parity and composed before/after tests.

### TP-C16 — The schema-v3 minigame definition is still not literal

TP-C5 says a complete row will ship, but the RFC does not provide it. Resource ID, payout fact,
cap key, grade curve, automation destination, quality decay, neutral Elo bounds, season member,
breadth source/destination, and exact modes remain implementation choices.

**Proposed contract:** include the complete JSON definition in the RFC. All structural strings and
every provisional integer must be explicit; the implementation copies those bytes into the fixture
and production artifact rather than authoring policy in Go.

### TP-C17 — Scoring normalization and exponent projection are ambiguous

“One canonical quantize” does not state the significant-digit constant, whether intermediate
Decimal operations use the core's ordinary quantization, or how the base-10 exponent is derived for
zero/sub-unit/scientific values. The exponent hardcap and reason key are unnamed.

**Proposed contract:** bind scoring to the existing Decimal canonical significant-digit rule at the
named final boundary; define zero and sign domain; define exponent as the canonical Decimal base-10
exponent after final quantization; name its integer min/max, saturation behavior, and reason key;
provide boundary vectors in the shared corpus.

### TP-C18 — No production mint is defined for the new artifact set

The live epoch manifest currently has no Fiscal, Minigame, Pet, Soul, or Pitch artifacts. Adding
Pitch is a balance mint and artifact-set growth, but the RFC gives no epoch ID/name, artifact order,
paths, dependency-complete bytes, or baseline regeneration duties.

**Proposed contract:** either (a) declare this implementation fixture-only and defer the production
mint, or (b) enumerate the complete next epoch artifact list and changelog. Recommended for honest
scope: implement engine/catalog/internal integration fixture-first; mint Pitch together with the
dependency-complete content epoch only after its balance rows pass the content gate.
