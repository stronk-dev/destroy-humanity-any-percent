# The Pitch

The Pitch is the first real tenant of the minigame platform: a deterministic solo card game where
metric cards and growth hacks produce an exact Decimal valuation against eight funding rounds. The
implementation is fixture-first. Its engine, catalog, copy, policy row, and verification corpus
ship, but no production epoch pins the content until the owner-gated First Content Epoch.

## Pinned identity and session boundary

The platform resolves Pitch content by the immutable `(constants_hash, engine_ref,
engine_version)` tuple. Creation, every applied command, and replay receive cloned canonical
artifact bytes, their SHA-256, schema version 1, and the server-owned session seed. Genesis records
the content hash and schema version in the exact thirteen-key snapshot. A missing, relabeled, or
deploy-current artifact fails closed; no process-global Pitch catalog participates.

Pitch registers engine `pitch` version `1.0.0`, solo mode, the closed command union
`play_hand | buy_hack | end_shop`, and the sorted rejection taxonomy. Physical cards use stable
`<card_id>#<copy_ordinal>` identities, so two copies of one metric can form a pair without allowing
duplicate command IDs.

## Deterministic deck, shop, and scoring

The run coordinate is
`run_seed = Substream(seed, "pitch.run.v1").Next()`. Each round derives deck and shop PRNGs as
`Substream(run_seed XOR uint64(round), label)`. A rejection-sampled SplitMix64 Fisher–Yates shuffle
deals positions 0–6, 7–13, and 14–20; the shop performs weighted draws without replacement and
uses stable `pitch.offer.<round>.<slot>.<hack_id>` identities. No mutable PRNG cursor is saved.

Scoring is Decimal-only:

1. For each selected card, add every flat modifier to its base metric and apply every card factor.
2. Deterministically sum the per-card values.
3. Visit hacks once in raw-byte ID order, multiplying each satisfied shape factor or
   partner-present chain factor at its encountered position.
4. Quantize once after the final multiplication.

Shapes are `pair` and `full_hand`. The score itself is uncapped; only the display fact
`pitch.best_hand_exponent` saturates at 1,000,000 with `cap.pitch_exponent`. The shared big-number
vector evaluates two `1e300` cards under a `1e100` factor to exactly `2e400` in Go and TypeScript.

## Progression and economy composition

Pitch lives behind Fiscal unlock `minigame.pitch` and Soul gate `human_hobby`. Start resolves both
from the pinned Founder state: a missing Fiscal purchase rejects before tenant creation, and a
near-zero Soul band rejects with `human_content_locked`. Neither identity is client supplied.

The terminal result remains inside the platform's integer grammar: outcome
`funded | funding_failed`, display exponent, and `pitch.final_round` as the highest round entered.
The platform—not the tenant—selects that fact, applies the attended-day faucet to `company.cash`,
updates Founder offline quality/rating, and writes Company and Founder replay evidence atomically.
Live Founder mutation routes through `ApplyFounderLogged`, so Fiscal auto-sweep, receipt decoration,
and replay execute one transition boundary.

## Content and verification

The launch fixture contains exactly twelve two-copy metric cards, eight growth hacks, eight funding
targets, and visible reason keys for every hardcap. The versioned content-gate corpus covers every
card and hack, partner-present/absent chains, pair/full-hand trigger controls, terminal snapshot
bytes, and a fixed 108-transition budget in both runtimes. A real Postgres integration test proves
Fiscal reject/purchase/start, play, certification, payout, retry idempotency, and Founder replay.

The designated review consumed `{0eb3772, 853ef93, 2a55e12}` and approved archival under verdict
`c76101a`. Human playability remains owned by the Minigame & Recovery API + Surface successor.
