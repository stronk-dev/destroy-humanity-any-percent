# RFC: Active-Play Buff Windows Foundation

- **Status:** implemented — Company v18 scheduler, claims, Lucky saturation, combo hardcap, replay,
  schema-v2 cap events, and Exit reset are shipped and independently reviewed.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-05
- **Design refs:** `design/02 §2.3` (active play: golden opportunities on the shaped t⁵·exp spawn, the
  Lucky-bank formula, multiplicative buff stacking, the click clamp + timing-is-the-skill,
  the daemon idle mechanic gated behind the enshittification slider)
- **Depends on:** Production + Save + Run Genesis (implemented — buffs/opportunities/clicks are
  Company intents inside `ApplyLogged`); Numeric Core (implemented — Lucky/combo payouts are Decimal)
- **Owner ruling honored:** breadth-first — the spawn/Lucky/combo/click MECHANICS and wire; the
  spawn cadence, buff magnitudes, and combo ceiling are balance data. Clicking is RETAINED as a
  timing skill (owner ruling 2026-08-05); the rhythm-timing enhancement is a declared successor.
- **Planning:** `planning/active-play-buff-windows/` (once implementing)

## Summary

The active-build skill layer, made server-authoritative and deterministic: **golden opportunities**
that spawn on a seeded shaped distribution and grant short **multiplicative buff windows**, the
**Lucky-bank payout formula** (the bank-management meta), and the **click intent** with
its server-side rate clamp (timing, not raw speed, is the skill). Company-scoped, inside `ApplyLogged`,
activating new-run-forward at a **Company save-version bump to v18** under a pinned
`opportunities` artifact. This RFC specifies the spawn/payout/combo/click mechanics and wire, not the
spawn cadence, buff magnitudes, or combo ceiling (all balance data). The **daemon** idle
mechanic and the **rhythm-timing** click enhancement are declared successors (see §Successors).

## Motivation

The design wants active play to be a real, separate discipline from idle — a skill ceiling of buff
sequencing and bank management that active players climb, without an autoclicker arms race (the
well-documented failure mode where raw click automation, not timing, becomes the winning play).
Three things must be right: (1) **server authority** — clients
send click and activation *intents*, never payouts; the server owns the RNG, the clock, and the math;
(2) **determinism** — spawns are seeded and replayable, buff expiry is closed-form (no tick loop);
(3) **the clamp** — clicks are rate-validated so that *timing* is the skill, keeping automation out of
the cheat space and into sanctioned progression later (the AGI-tier macro bench, a successor).

Out of scope (successors / data): the daemon mechanic (gated on the enshittification toggle,
which does not yet exist), the rhythm-timing click enhancement (a declared successor), the AGI-tier
macro bench, and every magnitude/cadence/cap number.

## Specification

### AB1 — Golden-opportunity spawn (seeded, shaped, server-authoritative)

Opportunities spawn on a schedule the **server computes and LOGS** (A3 — there is no generic run-RNG
stream): a Company-run SplitMix64 substream `active_play.spawn.v1` (seeded from the run seed, cursor
`spawn_seq`) feeds a **server-only** t⁵·exp inverse-CDF sampler (`design/02 §2.3` — t⁵ suppresses
instant re-spawns, exp cuts the long tail); the sampled `next_spawn_attended_ms` is
LOGGED in the resolved arm and replay READS it (TS never recomputes the float). Spawn cadence numbers
are balance data. On spawn the server records `{opportunity_id, spawned_attended_ms,
expires_attended_ms, effect_ref, selected_generator?}` in **Company ATTENDED ms** (A2 — never wall
time, so absence never expires it); the opportunity is claimable via `claim_opportunity` (A4) until
`expires_attended_ms` (closed-form, lazy scheduler, no tick). At most ONE opportunity pending; a missed
one advances the schedule without a queue; an unclaimed opportunity expires with no penalty.

`effect_ref` selects a member of a **closed effect-type union** — effect + generator selection use
INTEGER weight tables over the same SplitMix64 substream (byte-identical Go/TS, A3):
- `production_frenzy` — ×N production for a window (design: ~×7 / ~77 s);
- `click_frenzy` — ×M per-click value (buffs the existing `perform_manual_batch`, A1) for a short
  window (design: ~×777 / ~13 s);
- `building_special` — +P%/owned of ONE selected generator (`1 + owned·per_owned_ppm/1e6`, owned =
  purchased count, the generator logged) for a window (design: +10%/owned, 30 s);
- `lucky_payout` — an instant cash grant per the AB2 formula (no window).
(`instant_payout` is REMOVED from v1, A5 — `lucky_payout` covers the instant-cash role; a named
successor if ever needed, rather than shipping a vague "bounded fraction".)

### AB2 — The Lucky-bank payout (Decimal-exact)

The `lucky_payout` grant is `min(lucky_bank_frac · bank, lucky_rate_cap · rate) + epsilon`, the
Lucky formula (`design/02 §2.3`: `min(0.15·bank, 900·rate)+ε`). `bank` (current Company cash) and
`rate` (current production/s) are **Decimal**; the payout is computed Decimal-exact and credited
through the normal resource path; wire format is **strings** (the big-number law). `lucky_bank_frac`,
`lucky_rate_cap`, and `epsilon` are balance data; the `min(fraction·bank, cap·rate)+ε` SHAPE is ruled
— it is what creates the ~`(cap/frac)`×rate bank-management target (design: ~6,000×rate) and the
hoard-vs-spend tension. The payout reads live `bank`/`rate` at resolution (server-owned), so the
client cannot inflate it.

### AB3 — Multiplicative buff stacking + the combo hardcap

Active buffs compose **multiplicatively** into the production stack (documented order,
`design/02 §2`). Buff state is the MINIMAL replay-owned set (A6) `{buff_instance_id, effect_row_id,
selected_target?, activated_attended_ms, expires_attended_ms}` on the Company save — the fixed factor
is **DERIVED from the pinned artifact, never stored** (so save and catalog can't disagree). A buff is
active iff `attended_now < expires_attended_ms` — **closed-form, evaluated on read, never ticked** (all
coordinates are Company ATTENDED ms, A2, so absence never expires a buff). Multiple active buffs
multiply in raw-byte instance order, quantized at the existing contribution boundary; the **event-buff
slot product is then clamped to `combo_cap`** — including the effective per-generator product of an
`all` production frenzy and generator-specific building specials — a visible HARDCAP + `reason_key`
(never a softcap) —
BEFORE the remaining production slots. Contributions register in `event_buffs` with exact source IDs;
each effect declares which targets it affects. Expired rows are filtered from the math immediately and
removed/evented by the next lazy transition. Server-owned throughout; the client renders, never decides.

### AB4 — The click intent and the rate clamp

**No new click intent (A1 — my `click_batch` duplicated the shipped surface).** The
server-authoritative click surface already exists: `perform_manual_batch`, backed by a persisted
milli-token bucket with an honest `applied_count`. That token bucket is the **SOLE** clamp (never a
client-supplied `window_ms` — that would be a client-authority hole); `applied_count` is the
non-punitive feedback. This RFC does NOT add a second intent — instead, an active `click_frenzy` buff
**modifies the per-click value of the existing `perform_manual_batch`** (server-derived, reading live
buff state). So above the token clamp, **timing (which clicks land inside which buff window), not raw
rate, is the skill.**

**The timing-skill enhancement is a successor** (owner ruling 2026-08-05: enrich clicking with
rhythm/timing mechanics). The server-validatable model is resolved:
**the 20 Hz sim tick (50 ms) IS the beat grid.** The server judges a coarse
±1-tick on-beat layer trusting NO wall-clock (the click already lands inside a known sim tick), and a
fine sub-tick "Perfect" layer stays cosmetic/solo-only so no unprovable timestamp ever gates farmable
power. The click clamp (~20/s = one legal click per tick) thus becomes the *instrument*: "click on the
beat-tick" is the skill, mashers lose nothing. Combo becomes a hardcapped visible "Groove" multiplier
(resets on a true miss, holds on a near-miss) that multiplies into the buff stack. This foundation
ships the clamped click intent + `client_window_ms`; the Groove/on-beat judgement lands as the
successor built on this model. (Note: a note-highway minigame variant needs a rhythm-game patent
freedom-to-operate check before its own RFC — flagged in the dossier.)

### AB5 — Scope, activation, wire

Company-scoped (opportunities/buffs are within-run; they reset on Exit). State (the pending
opportunity, active-buff array, scheduler cursor + `spawn_seq`, clamp accounting — all in ATTENDED
coordinates) lives in the Company save under a **save-version bump to v18** (A5/A8/D9 — NOT v17;
Doctrine took Company v17, so Active-Play sequences at v18 and requires the meters+achievements+
doctrines chain pinned), activating new-run-forward under a pinned `opportunities` artifact
(biconditional with floor ≥ 18; the activation-boundary law), Company axis only. The claim intent is
`claim_opportunity {intent_id, kind, expected_revision, opportunity_id}` (A4); there is NO click
intent (A1 — `perform_manual_batch` is reused). Resolved arms, receipts, and events
(`opportunity_spawned.v1`, `opportunity_claimed.v1`, `buff_expired.v1`) are exact byte-grammar; the
spawn schedule is server-logged and replay-read (A3); Go/TS shared vectors compare state, receipt, and
ordered event bytes — including the logged spawn schedule, the Lucky Decimal payout at boundary
banks/rates, combo-cap saturation, offline pause, and Exit reset.

## Successors (declared, not in this foundation)

- **Daemon idle mechanic** (`design/02 §2.3`): daemons attach to tenants (Tier 3+), each
  draining a % of visible rate, popping for ×(>1) of the drained total — quadratic-in-count patience
  bonus (the idle crown jewel). **Gated on the enshittification toggle** (the opt-in
  make-it-worse-for-a-multiplier slider), which is itself an unbuilt design-only mechanic
  (gap-backlog). Deferred until that toggle has a foundation; declared here so the buff-stack knows a
  drain-then-refund buff family is coming.
- **Rhythm-timing click judgement** (AB4): a declared successor built on the resolved beat-grid model.
- **AGI-tier macro bench:** sanctioned click automation as late-tier progression (`design/02 §2.3`).

## Deviations from design

- Mechanical naming (`production_frenzy`, `click_frenzy`, `building_special`, `combo_cap`) over the
  flavor; magnitudes/cadence/cap as data (naming + balance-data laws). Nothing else diverges;
  spawn/Lucky/combo shapes follow `design/02 §2.3` exactly.

## Acceptance criteria

1. Spawns are seeded from the run stream on the shaped t⁵·exp distribution; replay reproduces the
   schedule byte-identically; expiry is closed-form (no tick loop); an unclaimed opportunity expires
   with no penalty.
2. The Lucky payout is Decimal-exact `min(frac·bank, cap·rate)+ε`, reads live server state, wire as
   strings; boundary banks/rates verified in shared vectors.
3. Buffs multiply into the documented stack; the combo ceiling is a visible hardcap saturating with a
   `reason_key`; expiry is server-owned and closed-form.
4. The click clamp drops over-cap counts silently; within-clamp clicks feed click-value + combo;
   client sends the batch, never the value; `client_window_ms` is carried for the successor judgement.
5. Company save v18 activation under the pinned `opportunities` artifact, New-Founder-forward, Company
   axis only (Founder axis untouched); migration + corpus.
6. Byte-parity Go/TS shared vectors across spawn schedule, Lucky payout, combo cap, and click clamp.

## Historical open questions (resolved by A1-A8)

- **Company save version:** proposed v17 — confirm against any other pending Company-v17 claimant
  (the acceptance review should check no other Company mechanic is mid-flight for the same bump; if so,
  the atomic co-activation pattern (cf. meters v15 + achievements v16) applies).
- **Spawn while offline:** do opportunities spawn/accumulate during offline spans, or only while
  present? Proposed: they are an ACTIVE-play mechanic, so spawns accrue only on attended time (an
  absent player misses no idle progress — offline is the idle build's domain, active buffs are the
  active build's). Confirm.

## Acceptance-review blockers (2026-08-05)

The two flagged direction calls are sound: no other active RFC claims Company v17, and active-play
opportunities should advance only on attended time. The latter requires attended-coordinate state,
not wall timestamps that continue expiring while absent. The exact transition remains blocked on
A1-A8.

### A1 — `click_batch` duplicates the implemented manual-action clamp and trusts the wrong input

`perform_manual_batch` already supplies the server-authoritative click/action surface, backed by a
persisted milli-token bucket and an honest `applied_count`. A second `click_batch` creates two ways
to mint the same manual output. Worse, AB4's clamp is based on client-supplied `client_window_ms`, so
a client can claim a larger window and authorize more clicks; that contradicts server authority.
The written object also omits `expected_revision` despite claiming the C1 envelope.

**Proposed contract:** do not add `click_batch`. Bind click buffs to selected catalog manual-action
IDs and continue using `perform_manual_batch`; its server cursor/token bucket is the sole clamp and
`window_ms` remains audit/UX-only. The applied count already provides non-punitive feedback. If a
separate action kind is genuinely required, state why and give it an independent server-owned token
cursor; client window must never determine accepted count.

### A2 — Attended-only spawn/expiry has no coordinate or lazy catch-up rule

AB1/AB3 store wall `spawned_at`/`expires_at`, which makes an already-spawned opportunity or buff
expire while the founder is absent. That prices absence and contradicts the accepted attended-only
direction. There is also no rule for several spawn/expiry boundaries crossed between commands, or
for whether a rejected command commits lazy spawn/expiry cleanup.

**Proposed contract:** store all opportunity and active-buff coordinates in exact Company attended
milliseconds, resolved from the already-implemented offline-span ledger at each command. Advance a
single lazy scheduler before the command-specific action, preserving phase and processing a bounded
number of due transitions; absent wall time advances neither schedule nor expiry. Specify event
ordering and the same rejection/compound-transition rule needed by Fiscal F1. The catalog must cap
the number of due transitions within the online catch-up horizon or the implementation needs a
closed-form skip rule.

### A3 — The run RNG and t^5*exp sampler are not executable cross-runtime contracts

The Company save has no generic mutable run-RNG stream. The RFC gives a density shape but no inverse
sampler/discretization, seed derivation, substream labels, rejection rules, interval rounding, draw
order, or cursor state. Native `log`/`pow` in Go and JS cannot be assumed byte-identical. Effect and
building selection draw order is also unspecified.

**Owner ruling required:** define an integer-deterministic sampler or make the server log a resolved
schedule that replay verifies against an exact shared algorithm. Enumerate run seed source,
SplitMix64 substream labels, cursor/sequence state, uniform-to-interval transform, millisecond
rounding, effect-selection weights, building-selection order, and shared vectors. A float-based
sampler may be server-only only if replay records and validates sufficient deterministic evidence;
TS must not independently approximate it and call that parity.

### A4 — There is no claim intent or claim lifecycle

The summary says activation intents and AB1 says opportunities are claimable "by intent," but no
intent kind/keys, ordering, expired/unknown/already-claimed result, or pending-opportunity invariant
is specified. It is also unclear whether claiming an opportunity first accrues production, which
bank/rate Lucky reads, and whether one pending opportunity blocks later spawns.

**Proposed contract:** add exact `claim_opportunity {intent_id,kind,expected_revision,
opportunity_id}`. Resolve the lazy scheduler and normal accrual first, then validate the pending ID
and attended expiry, then atomically clear it and apply its effect. At most one opportunity may be
pending; missed opportunities advance the schedule without accumulating a queue. Reuse closed
rejection categories/details and define idempotent retry through the standard intent record.

### A5 — The `opportunities` artifact and effect union are incomplete

No exact artifact root exists for spawn policy, effect weights, duration/multiplier rows, eligible
manual actions/generators, instant-payout formula/cap, combo cap/reason key, or hardcap bounds.
`instant_payout` is only "a bounded fraction" and therefore not a mechanic. `building_special`
needs an exact factor equation and a stable selected generator.

**Proposed contract:** enumerate schema v1 with one schedule policy, weighted effect rows, explicit
effect-specific exact keys, combo policy, and click-action bindings. Define building factor
`1 + owned*per_owned_ppm/1e6` with purchased-vs-total count chosen explicitly; define the instant
payout equation or remove that arm from v1. Loader cross-validates every generator/action ID against
the same pinned economy catalog. Adding the artifact is the Company-v17 activation mint.

### A6 — Buff state duplicates catalog authority and lacks slot/expiry semantics

Storing `effect_ref` plus `multiplier` allows pinned policy and save state to disagree. The RFC does
not name the multiplier slot/source/target, within-slot ordering, whether production frenzy affects
manual output, or how expired rows are removed/evented. Saturating the product before or after
per-target selection can yield different results.

**Owner ruling required:** declare the minimal replay-owned state (recommended: buff instance ID,
effect row ID, selected target if any, activated/expiry attended coordinates; derive fixed factors
from the pinned artifact). Register contributions in `event_buffs` with exact mechanical source IDs;
state which targets each effect affects; multiply active rows in raw-byte instance order, quantize at
the existing contribution boundary, then clamp the event-buff slot product to `combo_cap` before the
remaining production slots. Expired rows are filtered for math immediately and removed/evented by
the next applied lazy transition.

### A7 — Faucet saturation and Lucky rate ownership are unstated

Lucky/instant payouts can hit the Company cash hardcap. Ordinary ledger writes reject overflow,
while passive faucets use accrual-only saturation; a claim at cap must not brick the stream. "Live
rate" also needs an exact target: cash base rate before or after currently active production buffs,
and excluding the opportunity being claimed.

**Proposed contract:** compute rate through the authoritative production rate path using all buffs
active immediately before the claim, excluding the pending effect itself; apply the payout through a
named faucet-saturation boundary that clamps to the resource hardcap and returns the actual credited
delta plus cap reason. Define zero-rate/zero-bank behavior and Decimal quantize order in shared
vectors.

### A8 — Save/wire/event state is asserted but not enumerated

Company v17 is available, but the exact new state keys and completeness rules are absent: scheduler
cursor/sequence, RNG position, pending opportunity, active buff array, and any clamp accounting.
AB5 names events without exact payloads and omits the claim intent/receipt resolved arm. A
`buff_expired` event generated lazily also needs an intent/event coordinate and stable order.

**Proposed contract:** enumerate the v17 save keys and null/empty representations, strict migration
and new-run activation, exact Go/TS intent/resolved/receipt/event objects, raw-byte buff ordering,
event registry/DB constraint changes, and a sequential fixture combining spawn, miss, claim, Lucky,
overlapping buffs, cap saturation, manual clicks, expiry, offline pause, Exit reset, and replay.

## Changelog

- 2026-08-05: created (draft) — Wave-A foundation; active-play skill layer, `design/02 §2.3`. Click
  clamp + timing-skill confirmed by owner ruling; rhythm-timing enhancement and daemon mechanic
  declared as successors.
- 2026-08-05: Codex acceptance review — Company v17 availability and attended-only direction
  confirmed; implementation blocked on A1-A8, including the duplicate/insecure click surface and
  the absent deterministic scheduler/RNG/claim contracts.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.

## Owner rulings on A1-A8 (2026-08-06)

All accepted. Two owner-calls (A3, A6) + the Company-version correction (v17 → **v18**, per Doctrine
D9). Several fix real errors in my draft (the duplicate click surface, the nonexistent run-RNG, the
wall-clock spawn/expiry).

- **A1 — accepted; DO NOT add `click_batch` (it duplicated the shipped clamp).** `perform_manual_batch`
  already IS the server-authoritative click surface with a persisted milli-token bucket and honest
  `applied_count` — a second intent would mint the same output twice, and my `client_window_ms` clamp
  was a client-authority hole. Click buffs (`click_frenzy`) bind to selected catalog manual-action IDs
  and MODIFY the existing `perform_manual_batch` value; its server cursor/token bucket is the SOLE
  clamp; `window_ms` stays audit/UX-only. AB4 reconciled (the `click_batch` intent is removed).
- **A2 — accepted; attended-ms coordinates + lazy scheduler (fixes my wall-clock error).** All
  opportunity + active-buff coordinates are exact Company ATTENDED milliseconds (resolved from the
  offline-span ledger), NOT wall time — so absence never expires an opportunity/buff (the attended-only
  law). One lazy scheduler advances BEFORE the command action, phase-preserving, processing a bounded
  number of due transitions; absent wall time advances neither schedule nor expiry. Same
  compound-transition/rejection rule as Fiscal F1. The catalog caps due transitions within the online
  catch-up horizon (or a closed-form skip rule). AB1/AB3 reconciled.
- **A3 — RULED (server-logs-the-schedule; fixes the nonexistent run-RNG + the Go/JS float problem).**
  There is NO generic mutable run-RNG stream. Introduce a Company-run SplitMix64 substream
  `active_play.spawn.v1` seeded from the run seed with a persisted `spawn_seq` cursor. **The t⁵·exp
  interval sampler is SERVER-ONLY (Go float inverse-CDF) and its RESULT is LOGGED in the resolved arm;
  replay READS the logged schedule and validates against it — TS never recomputes the float** (this is
  the established server-stamps-and-logs pattern, so no cross-runtime float parity is required for the
  interval). Effect-selection and building-selection use INTEGER weight tables over the same SplitMix64
  substream (byte-identical Go/TS — those ARE reproduced). The resolved arm logs `spawn_seq`, the
  sampled `next_spawn_attended_ms`, the selected `effect_ref`, and (for `building_special`) the
  selected generator; boundary vectors included.
- **A4 — accepted (claim lifecycle).** Add exact `claim_opportunity {intent_id, kind,
  expected_revision, opportunity_id}`. Resolve the lazy scheduler + normal accrual FIRST, then validate
  the pending ID + attended expiry, then atomically clear it and apply its effect. **At most ONE
  opportunity pending**; a missed opportunity advances the schedule without accumulating a queue. Reuse
  closed rejection categories; idempotent retry via the standard intent record.
- **A5 — accepted (the `opportunities` artifact); `instant_payout` REMOVED from v1.** Schema v1:
  schedule policy, weighted effect rows, effect-specific exact keys, combo policy, click-action
  bindings. `building_special` factor = `1 + owned·per_owned_ppm/1e6` using the OWNED (purchased) count
  with the selected generator logged (A3). `lucky_payout` covers the instant-cash role, so
  `instant_payout` is dropped from v1 (a named successor if ever needed) rather than shipped vague. The
  loader cross-validates every generator/action ID against the pinned economy catalog. **The artifact
  is the Company-v18 activation mint (NOT v17 — Doctrine D9 took v17; Active-Play sequences at v18,
  requiring the meters+achievements+doctrines chain pinned).**
- **A6 — RULED (minimal buff state; don't store the multiplier).** Store only `{buff_instance_id,
  effect_row_id, selected_target?, activated_attended_ms, expires_attended_ms}` — the fixed factor is
  DERIVED from the pinned artifact (so save state and catalog can never disagree, the flaw A6 names).
  Register contributions in `event_buffs` with exact mechanical source IDs; each effect declares which
  targets it affects (state whether `production_frenzy` touches manual output). Multiply active rows in
  raw-byte instance order, quantize at the existing contribution boundary, THEN clamp the event-buff
  slot product to `combo_cap` (visible hardcap + reason key) BEFORE the remaining production slots.
  Expired rows are filtered from the math immediately and removed/evented by the next lazy transition.
- **A7 — accepted (Lucky rate + faucet saturation).** The Lucky payout reads rate through the
  authoritative production-rate path using all buffs active immediately BEFORE the claim, EXCLUDING the
  pending effect itself (no self-reference). Apply the payout through a named faucet-saturation boundary
  that clamps to the cash hardcap and returns the actual credited Decimal delta + cap reason (accrual
  saturation, never overflow-reject — a claim at cap must not brick). Define zero-rate/zero-bank
  behavior + Decimal quantize order in shared vectors.
- **A8 — accepted (v18 save/wire/events).** Enumerate the Company-v18 save keys (scheduler cursor +
  `spawn_seq`, pending opportunity, active-buff array, clamp accounting) with null/empty reps; strict
  migration + new-run activation (v18 requires v17 doctrines pinned); exact Go/TS
  intent/resolved/receipt/event objects; raw-byte buff ordering; event registry/DB changes; one
  sequential fixture combining spawn, miss, claim, Lucky, overlapping buffs, cap saturation, manual
  clicks, expiry, offline pause, Exit reset, and replay.

A1-A8 fully ruled. Numbers (spawn cadence, magnitudes, combo_cap, Lucky constants) stay data; the
daemon mechanic + rhythm-timing judgement remain declared successors.

## Implementation acceptance recheck (2026-08-06) — blocked A9-A16

A1-A8 settle the intended play loop, but comparing them with the shipped Company save, `ApplyLogged`,
contribution-provider, and intent-store contracts exposes eight remaining executable gaps. No
Active-Play mechanic code started: filling these in locally would choose persistence, replay, and
numeric authority that the RFC still leaves open.

### A9 — The `opportunities` artifact is still a list of families, not an exact schema

A5 promises a schema-v1 artifact but does not enumerate exact root or nested keys, numeric wire types,
row identities, targets, weight totals, or ordering. The loader therefore cannot distinguish a missing
policy from a legal zero value, cross-check effect bindings, or generate a byte-identical TypeScript
catalog. `production_frenzy`, `click_frenzy`, and `building_special` also have different required
fields, but their closed tagged union is only prose.

**Proposed contract:** enumerate exact root `{schema_version:1,schedule_policy,effects,combo_policy}`.
Schedule policy names the sampler version, substream label, interval parameters, lifetime, and maximum
due transitions. Effects are a raw-byte-sorted tagged union keyed by `effect_row_id` and carrying
`weight`; `production_frenzy` has `{factor,duration_ms,targets}`, `click_frenzy` has
`{factor,duration_ms,action_ids}`, `building_special` has
`{per_owned_ppm,duration_ms,eligible_generator_ids}`, and `lucky_payout` has
`{lucky_bank_frac,lucky_rate_cap,epsilon,resource_id,hardcap_reason_key}`. Combo policy is
`{combo_cap,hardcap_reason_key}`. Decimal factors/constants use canonical Decimal strings; durations,
weights, ppm, and maximum-due are positive safe integers. Exact-key, sorted/unique, positive-weight,
target, resource, generator, action, and multiplier-source checks run in both loaders against the same
pinned economy artifact. Adding `opportunities` extends the epoch artifact authority and is a mint.

### A10 — The scheduler still has no reproducible seed/draw/identifier algorithm

A3 selects server-side floating-point interval sampling and integer effect selection, but it does not
define the run-seed function already used by the repository, how `spawn_seq` addresses a substream,
the inverse-CDF equation/rounding/domain, draw order, or how `opportunity_id` and `buff_instance_id`
are produced. A mutable PRNG cursor and a sequence-addressed pure draw are different save contracts.
"Replay reads the logged schedule" also does not say which fields replay recomputes and which it treats
as server-resolved evidence.

**Proposed contract:** derive each spawn independently from
`base=runidentity.Seed(founder_id,run_seq)` and
`determinism.Substream(base xor uint64(spawn_seq),"active_play.spawn.v1")`; do not persist a second
mutable PRNG state. Specify one Go sampler function and its exact finite-domain checks and
floor/ceiling-to-millisecond rule; only its sampled interval is trusted as a logged resolved input by
TS. After that draw, use `Bound(total_weight)` for effect and, only for `building_special`,
`Bound(total_generator_weight)` in that order; both runtimes recompute those integer selections and
reject logged disagreement. Derive UUIDv7-compatible opportunity/buff IDs from a separate named
substream plus the attended spawn/activation coordinate, with collision/domain vectors. Increment
`spawn_seq` exactly once per SPAWN; a post-miss reschedule reuses the already-advanced sequence (INFO-1 ruling 2026-08-06: every opportunity consumes a unique substream, byte-agreed both runtimes).

### A11 — Lazy scheduler mutation beside a rejected command is not representable

A2 says the scheduler runs before every command and cites Fiscal F1's compound-transition rule, but
Fiscal F11 remains unresolved. The store forbids events on `IntentRejected`, and `ApplyLogged` restores
the entire Company snapshot for every non-applied outcome. Therefore a due spawn/expiry cannot both
commit and preserve a rejected `claim_opportunity` receipt under the current boundary.

**Owner ruling required:** choose one contract rather than inheriting the unresolved Fiscal shape.
Recommended for this Company-scoped foundation: lazy spawn/expiry is part of the ordinary command and
rolls back with any semantic rejection; expired rows are still ignored immediately by all math, and
the next applied command persists cleanup/events. This keeps the global rejected-intent invariant and
needs no compound receipt. If time transitions must commit beside rejection, enumerate a new outer
applied/action-rejected receipt and event ordering for every Company intent before implementation.

### A12 — Active-buff contributions have no owner inside the replay boundary

Live contributions are currently frozen into `replay_inputs` before `ApplyLogged`; Active-Play buffs
are instead replay-owned Company state and may spawn/expire during the same transition. Computing them
in the external provider gives it a stale pre-transition view, while storing them in replay inputs
duplicates state/catalog authority. A6 also says each effect declares its targets but never actually
rules whether production frenzy affects manual output, how a building special targets a generator's
rate, or how click frenzy composes with existing manual-role contributions.

**Proposed contract:** add an internal `activePlayContributions(state,artifact,attended_now)` assembler
inside `ApplyLogged` and its TS mirror. It is the sole owner of `event_buffs` contributions and is not
serialized in `replay_inputs`. `production_frenzy` targets generator production only and never manual
actions; `click_frenzy` targets only its declared manual action IDs; `building_special` targets only
the logged selected generator. Source IDs are
`active_play.<effect_row_id>.<buff_instance_id>` and sort by raw bytes through the existing slot rule.
The transition advances/filters the scheduler before assembling contributions for command effects;
the pre-claim authoritative rate for Lucky uses the same active set but excludes the pending effect.
State, artifact, and resolved schedule are the only inputs.

### A13 — Lucky saturation and Decimal rounding remain directional

A7 names a new faucet-saturation boundary but neither the ledger nor production package currently
exports it. The RFC does not say whether each product is quantized before `min`, whether epsilon is
added before or after saturation, how a zero-bank/zero-rate payout behaves, or whether a zero actual
credit still consumes the opportunity and emits the claim event. These choices change receipt and
state bytes at magnitude boundaries.

**Proposed contract:** compute `bank_term=Quantize(frac*bank)` and
`rate_term=Quantize(rate_cap*rate)`, then `requested=Quantize(min(bank_term,rate_term)+epsilon)`;
apply one named accrual-only ledger method that returns `(actual_delta,saturated)` and never rejects at
the hardcap. Claim always consumes the opportunity, including actual delta zero, and its receipt/event
carry requested delta, actual credited delta, saturation boolean, and nullable cap reason (required iff
saturated). Add zero/zero, epsilon-only, one-ulp-below-cap, at-cap, and overflow-scale shared vectors.

### A14 — Company v18 state is not byte-enumerated

A8 says the keys are enumerated, but no keys or types follow it. The codec needs exact absent/null/empty
rules and invariants for the scheduler, pending object, and buff list; “clamp accounting” is named even
though A1 reuses the existing manual token fields and therefore may add no state at all. Raw-byte buff
ordering cannot be enforced without a canonical array rule.

**Proposed contract:** Company v18 appends exactly
`opportunity_spawn_seq`, `next_opportunity_attended_ms`, `pending_opportunity`, and `active_buffs` to
Company v17; no second clamp fields. Sequence/coordinate are non-negative safe integers. Pending is
`null` or exact `{opportunity_id,spawned_attended_ms,expires_attended_ms,effect_row_id,
selected_generator_id}` with the selected generator nullable and biconditional on effect kind.
`active_buffs` is a raw-byte-ascending array of exact A6 objects; IDs are unique, coordinates ordered,
and every row resolves in the pinned artifact. Pre-v18 states must contain none; v18 requires all four,
using `0`, `null`, and `[]` for a newly activated run.

### A15 — Intent/resolved/receipt/event bytes and rejection details are not enumerated

The current exact parsers, replay union, event registry, and DB CHECK cannot implement A8 from event
names alone. Missing shapes include schedule evidence on every command, claim result arms, Lucky
credit evidence, buff creation/expiry, miss cleanup, and deterministic ordering when several expire at
one coordinate. The existing rejection taxonomy also has no chosen details for unknown, stale,
expired, or mismatched opportunity IDs.

**Proposed contract:** enumerate one replay-inputs version bump whose resolved object contains an
`active_play` arm on every v18 Company command: scheduler before/after sequence and coordinate,
zero-or-more ordered expiry records, optional spawn record with sampled interval plus integer draw
evidence, and optional claim resolution. `claim_opportunity` has the A4 exact request. Reuse
`unknown_id/opportunity_id`, `not_eligible/opportunity_expired`, and
`not_eligible/opportunity_not_pending`; exact applied receipt adds a typed `opportunity` result while
retaining the standard receipt envelope. Event order is expired rows by buff-instance raw bytes,
missed pending expiry, newly spawned opportunity, then the command-specific claim/event. Enumerate
exact payload keys for `opportunity_spawned.v1`, `opportunity_expired.v1`,
`opportunity_claimed.v1`, `buff_started.v1`, and `buff_expired.v1`, then register them in Go and the
DB constraint. Shared sequential vectors compare raw state, receipt, and event bytes.

### A16 — Activation, initialization, and Exit reset are not wired as a closed invariant

`CatalogBundle.valid`, epoch identity, version floors, save codecs, and Exit activation currently know
Company v17 as the maximum. A5 says v18 requires the earlier foundation/doctrine chain, but does not
state whether an `opportunities` artifact without doctrines is invalid, how a current v17 run carries
to v18, or which initial schedule coordinate is used. Exit reset is asserted but no next-run schedule
initialization input is identified.

**Proposed contract:** `opportunities` presence is biconditional with Company floor 18 and requires
meters+achievements+doctrines; it is forbidden on Founder state. The Company codec maximum becomes 18
without changing the Founder axis. At an activating Exit/New Founder, initialize empty pending/buffs,
`spawn_seq=0`, and derive the first schedule from attended coordinate zero using the server-resolved
run identity; record that schedule in the new-run genesis/Exit evidence so its bytes are replayable.
An Exit from active v18 discards all old scheduler/pending/buff state and initializes the next run from
its own run seed and pinned next artifact. Artifact addition is a mint and all artifact-set/hash/KV-1
registries change in the same balance commit.

Until A9-A16 are ruled and reconciled into AB1-AB5, implementation would invent catalog, state,
replay, and transition semantics. A1-A8's gameplay decisions remain intact.

## Owner rulings on A9-A16 (2026-08-06) — exact catalog/state/replay/wire under A1-A8

All accepted (Codex's proposed contracts are executable and sound). Owner-call: A11.

- **A9 — accepted (exact artifact).** Root `{schema_version:1, schedule_policy, effects, combo_policy}`.
  `schedule_policy` names sampler version, substream label, interval params, lifetime, max-due-
  transitions. `effects` = raw-byte-sorted tagged union keyed by `effect_row_id` + `weight`:
  `production_frenzy {factor,duration_ms,targets}`, `click_frenzy {factor,duration_ms,action_ids}`,
  `building_special {per_owned_ppm,duration_ms,eligible_generator_ids}`, `lucky_payout
  {lucky_bank_frac,lucky_rate_cap,epsilon,resource_id,hardcap_reason_key}`. `combo_policy
  {combo_cap,hardcap_reason_key}`. Decimal canonical strings; durations/weights/ppm/max-due positive
  safe ints; exact-key/sorted/unique/target/resource/generator/action/multiplier-source checks in both
  loaders vs the pinned economy artifact. Adding `opportunities` is a mint.
- **A10 — accepted (exact scheduler seed/draw).** `base = runidentity.Seed(founder_id, run_seq)`;
  `determinism.Substream(base ⊕ uint64(spawn_seq), "active_play.spawn.v1")`; NO second mutable PRNG
  state. One Go sampler (finite-domain checks + floor/ceil-to-ms rule); only its sampled interval is a
  trusted logged input for TS. Then `Bound(total_weight)` for effect, and only for `building_special`
  `Bound(total_generator_weight)`, in that order; both runtimes recompute the integer selections and
  reject logged disagreement. UUIDv7-compatible opportunity/buff IDs from a separate named substream +
  the attended coordinate (collision/domain vectors). `spawn_seq` increments once per SPAWN; the post-miss reschedule reuses the already-advanced sequence (INFO-1 ruling 2026-08-06). (Consistent with the Fiscal F10 framing.)
- **A11 — RULED: ROLLBACK (and Fiscal F11 revised to match — one consistent model).** Lazy
  spawn/expiry is part of the ordinary command and ROLLS BACK with any semantic rejection; expired rows
  are still ignored immediately by all math, and the next APPLIED command persists cleanup/events. This
  keeps the global rejected-intent invariant and needs NO compound receipt. (The over-engineered
  commit-under-rejection wrapper is rejected for BOTH Active-Play and Fiscal.)
- **A12 — accepted (buff contributions owned inside `ApplyLogged`).** An internal
  `activePlayContributions(state, artifact, attended_now)` assembler inside `ApplyLogged` + its TS
  mirror is the SOLE owner of `event_buffs` contributions and is NOT serialized in `replay_inputs`
  (no stale external-provider view, no duplicated authority). Targets (resolving A6): `production_frenzy`
  → generator production ONLY (never manual); `click_frenzy` → only its declared manual action IDs;
  `building_special` → only the logged selected generator. Source IDs
  `active_play.<effect_row_id>.<buff_instance_id>`, raw-byte sorted. The transition advances/filters the
  scheduler BEFORE assembling contributions; the pre-claim Lucky rate uses the same active set EXCLUDING
  the pending effect.
- **A13 — accepted (Lucky saturation exact).** `bank_term = Quantize(frac*bank)`, `rate_term =
  Quantize(rate_cap*rate)`, `requested = Quantize(min(bank_term, rate_term) + epsilon)`; one named
  accrual-only ledger method returns `(actual_delta, saturated)` and NEVER rejects at the hardcap. A
  claim ALWAYS consumes the opportunity (incl. actual delta zero); receipt/event carry requested delta,
  actual credited delta, `saturated`, and a nullable cap reason (required iff saturated). Vectors:
  zero/zero, epsilon-only, one-ulp-below-cap, at-cap, overflow-scale.
- **A14 — accepted (v18 state bytes).** Exact absent/null/empty rules + invariants for the scheduler,
  pending object, and buff list; a canonical raw-byte array rule for buff ordering. NOTE: since A1
  reuses the existing `perform_manual_batch` token fields, "clamp accounting" adds NO new state.
- **A15 — accepted (wire).** A replay-inputs version bump whose resolved object carries an `active_play`
  arm on every APPLIED v18 command: scheduler before/after seq + coordinate, zero-or-more ordered expiry
  records, an optional spawn record (sampled interval + integer draw evidence), and optional claim
  resolution. `claim_opportunity` = the A4 request. Reuse `unknown_id/opportunity_id`,
  `not_eligible/opportunity_expired`, `not_eligible/opportunity_not_pending`; a typed `opportunity`
  result in the standard receipt. Event order: expired rows by buff-instance raw bytes → missed pending
  expiry → newly spawned → command-specific claim/event. Exact payloads for `opportunity_spawned.v1`,
  `opportunity_expired.v1`, `opportunity_claimed.v1`, `buff_started.v1`, `buff_expired.v1`; registered
  in Go + the DB constraint; shared sequential vectors.
- **A16 — accepted (activation/init/Exit).** `opportunities` presence biconditional with Company floor
  18, requires meters+achievements+doctrines, forbidden on Founder state; the Company codec maximum
  becomes 18 (Founder axis unchanged). At an activating Exit/New Founder: init empty pending/buffs,
  `spawn_seq=0`, derive the first schedule from attended coordinate zero using the server-resolved run
  identity, and record that schedule in the new-run genesis/Exit evidence (replayable). An Exit from
  active v18 DISCARDS all old scheduler/pending/buff state and inits the next run from its own run seed
  + pinned next artifact. Artifact addition is a mint; all artifact-set/hash/KV-1 registries change in
  the same balance commit.

A9-A16 fully ruled. AB1-AB5 refined (not contradicted); the compound-transition inconsistency with
Fiscal is resolved (both ROLLBACK). Numbers stay data.

### Designated-review amendments F2–F7 (2026-08-06)

The effective `event_buffs` product for a generator is the product of active `all` contributions and
that generator's specific contributions; `combo_cap` limits that combined product. The shared `all`
product consumes cap headroom before generator-specific products. A buff claim that first saturates
the product carries nullable `cap_reason_key` in its typed receipt and in schema-v2
`opportunity_claimed.v1` / `buff_started.v1` payloads (required exactly when the clamp bites); schema
v1 remains valid for existing rows. Schedule replay permits and verifies the ruled compound transition
where one applied command records a miss and the successor spawn. A13's five named Lucky boundary
vectors are shared Go/TS fixtures, not prose-only acceptance debt.
