# RFC: The Pitch (minigame content — THE TEMPLATE)

- **Status:** accepted; implementing (TP-C1–TP-C25 ruled). v1 scope: engine + pinned catalog +
  internal integration; playability lands with the Minigame API & Surface successor. **This is the exemplar
  minigame-content RFC**: its structure (tenant row → engine contract → certified result → economy
  hooks → content-as-data) is the template the other minigame content RFCs replicate.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/03 §8/§9` (casino tier + Demo Disc Arcade — the two homes), `design/03 §12b`
  (the Fairness Law — ranked outcomes never scale with tier power), `design/02 §2.2` (the
  multiplicative production stack this game deliberately mirrors)
- **Depends on:** Minigame Platform Foundation (implemented + archived — C1–C40; this registers as a
  tenant), Fiscal Quarters (archived — the unlock sink), Soul Foundation (`soul_gate` field,
  archival-eligible). This RFC adds a TENANT, an ENGINE, CONTENT, and **two named platform/
  composition amendments** (TP-C2's pinned `pitch` artifact + `CatalogBundle.Pitch`; TP-C6/TP-C15's
  `fiscal_unlock` resolver arm).
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
`{minigame_id: "pitch", engine_ref: "pitch", engine_version: "1.0.0", modes: ["solo"],
result_score_fact_ids: ["pitch.final_round", "pitch.best_hand_exponent"], scaling, payout,
fallback, offline_quality, rating_policy, unlock_condition}` — every sub-object per the
already-ruled closed grammars, with the complete literal bindings in the TP-C16 ruling (grade curve
on rounds reached, cap.minigame_faucet, season s1, neutral Elo 1000 [0,3000], fallback
{kind:"solo"}). Specifics:
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
constraint: **exactly 12 `metric_card`s and 8 `growth_hack`s** (the literal rows are in the TP-C13
ruling), each individually flat and legible (depth comes from hack INTERACTIONS, which the closed
effect union makes enumerable and testable). Card/hack rows are exact-key catalog data with copy keys through the copy
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
4. Content: launch catalog rows exact-key validated; copy keys resolve; the Pitch-owned content
   corpus covers every card/hack and its ruled interaction controls.
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

## Owner rulings on TP-C11–TP-C18 (2026-08-07) — the literals

- **TP-C11 — accepted; bodies reconciled in this edit** (TP1 literals, TP4 exact counts, the
  header's "no new platform mechanics" corrected to name the two amendments).
- **TP-C12 — RULED, and the command set is NARROWED for v1** (per TP-C3's "owner-selected closed
  set"): three arms — `play_hand {card_ids[]}` (a sorted, duplicate-free subset of the current hand,
  size ≤ `play_size`), `buy_hack {offer_id}`, `end_shop {}`. `draft_card`/`slot_hack` are REMOVED
  from v1 (hands are dealt automatically; purchased hacks auto-slot up to the slot hardcap; a
  drafting variant is a successor). **Phases:** `playing → shop → (playing | terminal)`; genesis
  enters `playing` at round 1. **Genesis:** deck = every card × its `copies`, shuffled via substream
  `pitch.deck.v1`; hand of `hand_size` dealt; `run_currency = start_currency`; shop offers drawn via
  `pitch.shop.v1` on each shop entry. **Snapshot exact keys:** `{phase, round, hands_remaining,
  deck_count, hand[], slotted_hacks[], run_currency, shop_offers[], funding_target,
  round_best_valuation, revision}` — IDs are strings, all Decimal values canonical strings; `hand`
  and `slotted_hacks` sorted raw-byte; cards selected by ID (never index). Playing a hand scores it,
  decrements `hands_remaining`, and updates `round_best_valuation`; meeting `funding_target` within
  the round's hands enters `shop`; exhausting hands below target is the terminal failure; clearing
  the final funding row is the terminal success. One terminal transition emits the result. **Sorted
  rejection enum:** `duplicate_card | hack_slots_full | hand_too_large | illegal_phase |
  insufficient_currency | unknown_card | unknown_offer` — all reject without mutation.
- **TP-C13 — RULED (exact schemas + the 20 launch rows + the curve; values provisional bytes).**
  Effect arms (discriminator `kind`): `flat_add {amount}` · `card_factor {factor}` (per played
  card) · `shape_factor {shape ∈ pair|full_hand, factor}` (narrowed by TP-C22 — `triple`/`flush_kind` are successor arms) · `chain_factor
  {partner_hack_id, factor}` (applies only when the partner is also slotted; evaluation position =
  the ordered-interactions stage). Amount/factor are canonical Decimal strings.
  `metric_card {card_id, base_metric, copies, copy_key}` — the 12 (base_metric provisional):
  `api_call 15` · `beta_invite 30` · `cache_hit 20` · `demo_day 90` · `newsletter 25` ·
  `page_views 10` · `patch_release 40` · `press_mention 75` · `referral_loop 60` ·
  `testimonial 50` · `uptime_nines 45` · `users_signup 35`; `copies: 2` each; copy keys
  `pitch.card.<card_id>`.
  `growth_hack {hack_id, price, draft_weight, effect, copy_key}` — the 8 (prices/weights
  provisional): `ab_test {3, 10, card_factor 1.5}` · `buzzword {2, 12, flat_add 25}` ·
  `dark_pattern {4, 8, shape_factor pair ×3}` · `growth_loop {5, 6, card_factor 2}` ·
  `infinite_scroll {3, 10, flat_add 50}` · `pivot {6, 4, shape_factor full_hand ×4}` ·
  `stealth_mode {4, 6, chain_factor pivot ×2.5}` · `synergy_deck {4, 6, chain_factor ab_test ×2}`;
  copy keys `pitch.hack.<hack_id>`. Every `chain_factor` partner exists (loader-checked).
  **Funding curve** `{round, funding_target}`: `1:100 · 2:300 · 3:1000 · 4:5000 · 5:25000 ·
  6:200000 · 7:2000000 · 8:50000000` (8 rounds; provisional). **Global policy:** `hand_size 7,
  play_size 4, hands_per_round 3, hack_slots 4, start_currency 4, shop_size 3` — all visible
  hardcaps with reason keys.
- **TP-C14 — accepted.** Artifact `balance/pitch.json`, `schema_version: 1`; canonical hash by the
  platform's established constants-hash construction; new `CatalogBundle.Pitch`; **content identity
  lives in exact genesis keys** `{pitch_content_hash, pitch_schema_version}` (no DB column — no
  queryable consumer exists); start-time equality checked among epoch, bundle, definition, and
  genesis; replay resolves through the constants-hash-pinned bundle. No session-table migration.
- **TP-C15 — accepted.** The third unlock arm, exact wire
  `{"kind":"fiscal_unlock","unlock_id":"minigame.pitch"}`; the literal Fiscal row
  `{unlock_id:"minigame.pitch", cost: 3}` (provisional); start resolves the pinned Founder state's
  byte-sorted unlocked set and rejects `not_eligible/fiscal_unlock_required` otherwise; Go/TS loader
  parity + composed before/after tests.
- **TP-C16 — RULED.** The complete literal definition ships IN THE RFC BODY'S TP1 (reconciled this
  edit) with these bindings: `engine_version "1.0.0"`, modes `["solo"]`, resource `cash`,
  payout fact `pitch.final_round`, grade curve `{1: 200000, 3: 500000, 5: 800000, 8: 1000000}`
  (round-reached → grade ppm, provisional), cap reason `cap.minigame_faucet`, season member `s1`,
  neutral Elo `1000` bounds `[0, 3000]` (untouched — solo), breadth source = unlocked card-set
  variants (v1: the single launch set), `fallback {kind:"solo"}`, `soul_gate "human_hobby"`.
  **Where any literal field name here drifts from the shipped schema-v3 grammar, the shipped grammar
  governs and the value mapping is 1:1** — the implementation copies bytes, never authors policy in
  Go.
- **TP-C17 — accepted.** Scoring binds to the core's canonical significant-digit rule at ONE named
  final boundary (the `play_hand` valuation quantize); intermediates use ordinary core Decimal ops;
  the domain is non-negative (zero valid). `pitch.best_hand_exponent` = the canonical base-10
  exponent after final quantization; zero → exponent 0; hardcap `1_000_000` with reason
  `cap.pitch_exponent`, saturating. Boundary vectors (zero, sub-unit, 1e12±, the hardcap) in the
  shared corpus.
- **TP-C18 — RULED (option a): fixture-first.** This implementation is fixture-only; the production
  mint of the ENTIRE new artifact set (fiscal + minigame/pet schema bumps + soul + pitch) is
  owner-gated in a dedicated **First Content Epoch RFC** (shared with SR-C13), minted only after the
  Pitch content gate passes. No partial chain, no epoch bytes here.

## Implementation-readiness blockers (Codex, 2026-08-07) — TP-C19–TP-C25

TP-C11–TP-C18 settle the literal content and product scope. A source walk against the shipped
tenant boundary found seven remaining executable gaps. These are not balance retuning: the current
contracts permit multiple incompatible byte histories, and two literal rows are unreachable under
the ruled command grammar. Implementing through them would require inventing mechanics.

### TP-C19 — The pinned Pitch artifact still cannot reach the pure tenant

`CatalogBundle.Pitch` names an owner, but the shipped tenant receives only
`CreateInput{mode,seed,scaling_inputs}` and `ApplyInput{mode,revision,snapshot,command,scaling_inputs}`.
`TenantRegistry` is process-static and `Service.Play` does not resolve catalog bytes. C14 therefore
cannot perform its required epoch/bundle/definition/genesis equality check, and replay cannot give
the engine the hash-pinned cards it must recompute.

**Proposed contract:** add one platform-owned `TenantContentResolver` keyed by
`(constants_hash,engine_ref,engine_version)`. It returns canonical artifact bytes plus the artifact
SHA-256 and schema version. `CreateInput` and `ApplyInput` receive cloned immutable content bytes,
content hash, schema version, and the server-owned session seed; live play and replay resolve them
from the session's immutable `constants_hash` and `seed`. The Pitch tenant rejects any mismatch
between those resolved values and the snapshot's identity. No tenant reads process-current data and
no new session column is needed.

### TP-C20 — The exact snapshot and the deterministic deal cannot both hold

C12's “exact” snapshot omits C14's required `pitch_content_hash` and
`pitch_schema_version`. It also has no seed, draw cursor, or deck order, while `ApplyInput` currently
has no seed. Even after C19, the rules do not say whether the 24-card deck persists across rounds,
when it reshuffles, or how `deck_count` changes. More sharply: `play_hand.card_ids` is duplicate-free
while every catalog card has two copies and `dark_pattern` requires a pair. If `card_ids` means base
IDs, the launch hack is unreachable.

**Proposed contract:** every snapshot has exactly thirteen keys: C12's eleven plus
`pitch_content_hash` and `pitch_schema_version`. Represent each physical card as the stable instance
ID `<card_id>#<copy_ordinal>`; `hand[]` and command `card_ids[]` contain instance IDs, remain
duplicate-free, and scoring resolves their base card IDs. At each round, derive a fresh full-deck
Fisher–Yates permutation using SplitMix64 with mandated rejection sampling and
`Substream(seed,"pitch.deck.v1",round)`. Deal positions 0–6, 7–13, and 14–20 for the three hands;
the four unused cards stay in the deck. `deck_count` is the number remaining after the current deal.
The snapshot needs no mutable PRNG cursor because `(seed,round,hand_number)` reproduces the draw.

### TP-C21 — Shop identity and run-currency income are absent

The engine starts with four currency and can spend it, but no transition earns currency. Hacks cost
up to six, so legal catalog rows may be permanently unaffordable. `shop_offers[]` also has no item
schema, weighted-without-replacement rule, ownership exclusion, purchase-removal rule, or stable
`offer_id`; the same seed can legally produce different shops in two runtimes.

**Proposed contract:** add one explicit `round_clear_currency` integer to global policy (owner must
choose the provisional literal) and grant it exactly once when a non-final round clears, before
entering shop. Shop offers are exact objects `{offer_id,hack_id,price}`. Generate `shop_size`
unowned hacks without replacement by `draft_weight` using SplitMix64 rejection sampling under
`Substream(seed,"pitch.shop.v1",round)`; `offer_id = "pitch.offer.<round>.<slot>.<hack_id>"`.
Buying removes the offer and auto-slots the hack; owned hacks never reappear. `end_shop` discards
unbought offers and advances the round. Currency and prices are exact nonnegative safe integers.

### TP-C22 — The effect formulas and shape vocabulary still admit divergent scores

`flat_add` does not say whether it applies once per hand or once per card; `card_factor` says “per
played card” but has no target selector; and the order between those arms is not a byte equation.
The shape union includes `flush_kind`, but cards have no kind field, and includes `triple` while the
launch deck has only two copies per base ID. Go and TypeScript can therefore implement different
legal valuations, and two union members have no computable predicate.

**Proposed contract:** narrow schema-v1 shapes to the used `pair | full_hand` members. `pair` means
at least two selected instances share a base card ID; `full_hand` means exactly `play_size`
instances. For each selected card, compute `base_metric + sum(flat_add amounts)`, then multiply that
card by every slotted `card_factor`; sum the per-card values; multiply once by each satisfied
`shape_factor`; then multiply once by each `chain_factor` whose partner is slotted, visiting hacks
in raw-byte `hack_id` order. Quantize once after the final multiplication. `triple` and
`flush_kind` require successor content/schema fields and are not reserved executable v1 arms.

### TP-C23 — Terminal result semantics are not named

C12 says “final round” and “round reached” in different places, but does not say whether a failed
round is counted, nor name `Result.outcome`. It also does not define the terminal snapshot's
`phase`, `hands_remaining`, or retained Decimal valuation. Those bytes feed payout and are compared
during replay.

**Proposed contract:** `pitch.final_round` is the highest round entered (failure on round 1 emits
1; clearing round 8 emits 8). Outcomes are the closed `funded | funding_failed`. The command that
fails the last available hand or clears round 8 writes `phase:"terminal"`, retains the final
`round_best_valuation`, leaves `hands_remaining:0` on failure and the post-play count on success,
sets `shop_offers:[]`, and emits score facts byte-sorted as
`pitch.best_hand_exponent`, `pitch.final_round`; `rating_delta:null`.

### TP-C24 — TP-C16 still does not form a loadable schema-v3 definition

The ruled binding omits required payout literals (`sends_per_day`, `per_send_cap`,
`conversion_ppm`), offline-quality decay literals and automation destination, provisional-games,
and the exact scaling row. “Breadth source = unlocked card-set variants” is not a shipped scaling
source kind, and `resource cash` is not the registered `company.cash` ID. The current loader also
has no `fiscal_unlock` arm, so a row matching C15 cannot load.

**Proposed contract:** provide the complete literal schema-v3 JSON row, including all six payout
keys, all six offline-quality keys, the one scaling row, all five rating keys, and the exact
automation destination. For v1's single launch set, use a literal breadth input of 1 unless a real
card-set state owner is added. Bind `company.cash` exactly. Extend the closed unlock union in both
loaders to `fiscal_unlock {unlock_id}`, and make the composed resolver read the pinned Founder
fiscal unlock set. No implementation chooses the missing balance integers.

### TP-C25 — The Pitch-owned content gate has no reproducible corpus

TP-C9 requires every row to affect a declared golden scenario, but no scenario file, seed, command
sequence, expected valuation, or gate budget exists. “Reachable in seeded generation” is also not a
finite CI assertion without a seed set. A test author would have to invent the acceptance evidence.

**Proposed contract:** check in a versioned Pitch content-gate corpus. Each of the twelve cards and
eight hacks names at least one exact `(seed,commands)` scenario and expected terminal snapshot;
chain hacks include both partner-present and partner-absent rows, and `dark_pattern`/`pivot` include
their triggering and control hands. The gate validates row coverage structurally and byte-compares
Go/TypeScript outputs. Declare a fixed maximum transition count equal to the sum of corpus command
counts; content changes regenerate this corpus in a reviewable balance-change commit.

## Owner rulings on TP-C19–TP-C25 (2026-08-07)

All accepted; owner literals supplied where demanded.

- **TP-C19 — accepted.** One platform-owned `TenantContentResolver` keyed
  `(constants_hash, engine_ref, engine_version)` returning canonical artifact bytes + SHA-256 +
  schema version; `CreateInput`/`ApplyInput` receive cloned immutable content bytes, content hash,
  schema version, and the server-owned session seed; live and replay both resolve from the session's
  immutable identity; the tenant rejects any identity mismatch. No process-current reads, no new
  session column.
- **TP-C20 — accepted.** The snapshot has exactly THIRTEEN keys (the C12 eleven +
  `pitch_content_hash`, `pitch_schema_version`). Physical cards are instance IDs
  `<card_id>#<copy_ordinal>`; `hand[]`/`card_ids[]` carry instance IDs (duplicate-free), scoring
  resolves base IDs — which makes `dark_pattern`'s pair reachable via the two copies (the exact
  defect this catches). Per-round full-deck Fisher–Yates via SplitMix64 with mandated rejection
  sampling under `Substream(seed, "pitch.deck.v1", round)`; deal positions 0–6 / 7–13 / 14–20 for
  the three hands; `deck_count` = remaining after the current deal; no mutable PRNG cursor
  (`(seed, round, hand_number)` reproduces every draw).
- **TP-C21 — accepted; the income literal is `round_clear_currency: 3` (provisional).** Granted
  exactly once when a non-final round clears, before entering shop. Shop offers are exact
  `{offer_id, hack_id, price}`; `shop_size` unowned hacks drawn without replacement by
  `draft_weight` (SplitMix64 rejection sampling, `Substream(seed, "pitch.shop.v1", round)`);
  `offer_id = "pitch.offer.<round>.<slot>.<hack_id>"`; buying removes the offer and auto-slots;
  owned hacks never reappear; `end_shop` discards and advances. All currency/prices nonnegative
  safe integers.
- **TP-C22 — accepted; the shape union NARROWS to `pair | full_hand`** (supersedes TP-C13's
  four-member list — `triple`/`flush_kind` need successor content/schema fields and are NOT reserved
  v1 arms; no launch hack used them). `pair` = ≥2 selected instances share a base ID; `full_hand` =
  exactly `play_size` instances. The byte equation: per selected card `base_metric + Σ(flat_add)`,
  × every slotted `card_factor`; Σ the per-card values; × each satisfied `shape_factor`; × each
  `chain_factor` whose partner is slotted, hacks visited in raw-byte `hack_id` order; ONE quantize
  after the final multiplication.
- **TP-C23 — accepted.** `pitch.final_round` = the highest round ENTERED (failing round 1 emits 1;
  clearing round 8 emits 8). Outcomes: closed `funded | funding_failed`. The terminal transition
  writes `phase:"terminal"`, retains the final `round_best_valuation`, `hands_remaining: 0` on
  failure / the post-play count on success, `shop_offers: []`, score facts byte-sorted
  (`pitch.best_hand_exponent`, `pitch.final_round`), `rating_delta: null`.
- **TP-C24 — accepted; the missing literals (all provisional balance data):** payout
  `{sends_per_day: 5, per_send_cap: 300, conversion_ppm: 500000}`; offline-quality
  `{decay_grid_ms: 3600000, decay_ppm_per_grid: 10000}` with the automation destination bound 1:1
  to the shipped destination kind the fixture tenant uses, targeted at `minigame.pitch`; rating
  `{neutral: 1000, floor: 0, ceiling: 3000, season: "s1", provisional_games: 10}` (inert — solo);
  the scaling row is a literal breadth input of `1` (the single launch set; a real card-set state
  owner is a successor); the resource binds exactly `company.cash`. **The closed unlock union in
  BOTH loaders extends with `fiscal_unlock {unlock_id}`** and the composed resolver reads the
  pinned Founder fiscal unlock set (the TP-C15 arm made loadable). The shipped-grammar-governs
  1:1-mapping clause (TP-C16) applies to every literal here.
- **TP-C25 — accepted.** A versioned Pitch content-gate corpus is checked in: every one of the 12
  cards and 8 hacks names ≥1 exact `(seed, commands)` scenario with its expected terminal snapshot;
  chain hacks include partner-present AND partner-absent rows; `dark_pattern`/`pivot` include
  triggering and control hands; the gate validates coverage structurally and byte-compares Go/TS
  outputs; the fixed maximum transition budget = the sum of corpus command counts; content changes
  regenerate the corpus in a reviewable balance-change commit.
