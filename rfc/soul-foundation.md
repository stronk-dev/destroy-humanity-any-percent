# RFC: Soul Foundation (the personal ledger)

- **Status:** accepted — SB1-SB16 ruled (exact artifact, soul_exhausted_source_ids eligibility, the
  ApplyDebit component, touch-grass `soul_recovery_sessions` with NO Company bump, human-content
  classification, v20 activation+wire). Implementation DEPENDENCY-BLOCKED on Fiscal v19 being
  implemented; the pure catalog/band layer may land earlier.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-05
- **Design refs:** `design/02 §8` (Soul — the personal ledger, distinct from the moral Trust axis;
  drains via Faustian contracts/crunch/longevity, recovers via deliberate touch-grass, gates the
  human content via the pet proxy, the transcendence-at-zero "training data" ending)
- **Depends on:** Save + Run Genesis (implemented), Founder Attendance (implemented —
  `ApplyFounderLogged` + the Founder attended cursor), Pet Care Foundation (implementing — the pet
  is Soul's *display proxy*, the greying). The Founder save-version chain (proposes v20, after
  Fiscal Quarters v19).
- **Owner rulings honored (2026-08-05):** model = **meter-shown, currency-triggered hybrid**;
  recovery = **deliberate, opportunity-costed** (no idle refill); gating = **graduated**
  (soft→hard). Drain *sources* (specific contracts/crunch/longevity) are content; some are gated on
  the events engine — declared successors.
- **Planning:** `planning/soul-foundation/` (once implementing)

## Summary

Soul — "what's left of *you*", distinct from the moral Trust axis — as a Founder-scoped ledger that
is **displayed like a meter but moved like a currency**: an always-visible condition read *through
the pet* (never a wallet UI), that only changes via **discrete, opt-in, clicked** debits at
Faustian/crunch/longevity moments and **deliberate, opportunity-costed** recovery — never a passive
drain and never a passive idle refill. It gates the human content on a **graduated** ramp (soft
greying early → hard locks near-zero) and its balance at Transcendence selects the ending. This RFC
specifies the ledger, the opt-in debit and recovery intents, the graduated-gating derivation, and
the ending hook — not the drain-source catalog or any threshold/amount (content and data).

## Motivation

The design wants Soul to carry sincere weight (the pet layer is sincere, never a joke) without
becoming "prestige-currency #2" to min-max — the failure mode a pure spendable
currency invites. The hybrid resolves it: **the meter presentation gives the sincerity** (you read your Soul
in how the pet treats you, not in a number), while **the opt-in clicked debit gives the ownership**
(you *chose* to sign — that hits harder than an imposed drain). Three things must hold: (1) Soul is
never rendered as a wallet or spent freely — only debited at declared opt-in moments, each with a
law-10 curtain tooltip; (2) recovery is a deliberate act that costs time and produces nothing — no
passive idle refill (or the drain is toothless and the low-Soul endings unreachable); (3) the
gating is graduated and telegraphed — no cliff (the no-unexplained-caps sensibility).

Out of scope (content / successors / data): the drain-source catalog (specific Faustian contracts,
crunch events, longevity rungs — several gated on the events engine), the touch-grass activity
catalog, the exact band thresholds / debit amounts / recovery magnitudes, and the ending's authored
copy. Also out of scope: any depiction beyond institutional metaphor (see S5).

## Specification

### S1 — The Soul ledger (Founder-scoped, meter-*presented*)

Soul is the **already-existing** signed `soul` field on the **Founder save** (persists across Exit —
the founder's personal ledger; the game remembers across runs), bounded `[soul_floor, soul_max]`
(balance data). **v20 ACTIVATES this dormant field (SB1) — it does not add a new one:** pre-v20 keeps
`soul == 0` and no Soul-owned state (rejects nonzero); the first Exit/New-Founder activation under the
pinned Soul artifact initializes it to `soul_initial` (in `[floor,max]`, recommended = max), recorded
in the Founder Exit resolved arm so replay derives it from the pinned bundle. **The EXACT bounded meter
+ band IS visible and published via the pet panel (SB6 — moral meters are transparent, the
published-formula law); Soul is simply never a currency/shop wallet.** One ledger, no second authority;
history lives in the Founder log/events, not the save.

### S2 — Debits: opt-in, currency-*triggered*, never passive

Soul decreases only via an **opt-in debit COMPONENT invoked INSIDE the owner transition (SB2 — NOT a
standalone rail):** an Event / longevity / contract owner debits Soul in the SAME transaction that
grants its benefit, so benefit + debit commit atomically and the offer's eligibility is proven by that
owner. **No public standalone `pay_soul_price` ships in production**; a **fixture-only** Founder command
exercises the pure component for tests. The debit resolves the source's amount + the **mandatory
curtain-tooltip copy key** (law 10 — the deal states in plain text that it costs Soul). **There is NO
passive Soul drain.** Affordability (SB3): a price must be FULLY affordable — reject `unaffordable/soul`
if it would cross the floor; ONLY a source declared `may_exhaust: true` may consume the exact remaining
balance to the floor once (loader-enforced single-use). The **source catalog** (VC term-sheet clause
7(c), crunch/founder-mode, longevity rungs, skipping the recital) is content whose owners (mostly the
events engine) are declared successors; this foundation ships the debit COMPONENT + a fixture source.

### S3 — Recovery: deliberate, opportunity-costed, never idle

Soul increases only via a **real touch-grass ACTIVITY with a production-suppression interval (SB4 — NOT
an instant command; an instant grant would be free).** The activity mirrors the minigame session:
**start** against the active Company (server-owned attended start + duration + a Company
production-suppression interval), then **resolve** after sufficient Founder attended time in ONE
Company+Founder transaction that grants the catalog Soul amount (saturating at `soul_max`) AND records
ZERO resource output for that interval. It is **mutually exclusive with production and other exclusive
activities** — the opportunity cost is real (the whole point of the recovery ruling). Cancel /
reconnect / Exit / concurrency are enumerated. **Soul does NOT passively refill while idle** — only a
resolved activity restores it. The per-playstyle asymmetry is inherent (cheap for an idle build, a real
tax for an active build), but it is always a deliberate, production-forgoing act. The activity catalog
is content; the lifecycle (start → attended-duration + production-suppression → resolve) is ruled.

### S4 — Graduated gating, derived through the pet proxy

Soul's consequences are **DERIVED** from the ledger value (never stored), on a **graduated** ramp of
closed bands (STRUCTURE ruled; thresholds are data): `whole → dimming → hollow → near-zero`. The
band drives:
- **The pet proxy (the display):** the pet-care trust/greying rendering tech renders the band — a whole-Soul
  pet recognizes and greets you; as Soul dims, recognition fades and flavor cools; near-zero, the
  pet's UI greys out. Soul is *shown* here and only here — this is the meter-presentation.
- **Soft consequences (early bands):** flavor/greying/tooltip changes only — no mechanical lock.
- **Hard consequences (near-zero band ONLY):** hobby minigames and pet interactions lock behind
  "you no longer remember why this was fun" (a derived predicate the UI/minigame gate reads). The
  hard lock bites only at the bottom — graduated, telegraphed by the soft bands above it, no cliff.

The gating predicate is a pure read of the ledger against the pinned band thresholds — deterministic,
never stored, byte-parity Go/TS.

### S5 — The transcendence ending hook (institutional metaphor only)

**The ending keys on a DATED FACT, not a live band (SB7 — reconciles the `design/02 §7` dated-facts
law):** the FIRST time an opt-in debit reaches the exact floor, emit an immutable `soul.depleted`
Founder fact/event; **Transcendence keys on that dated fact plus current band** (full Soul → earnest
ascension; the dated `soul.depleted` → the "you become training data" ending) — never an arbitrary live
threshold. This RFC exports only the ending ENUM + the dated-fact emission, not the authored content.
**Content constraint (law/safety, ruled):** the zero-Soul ending stays **institutional metaphor**
(training data, a dataset, a model input) and **never depicts self-harm or suicide** (`AGENTS.md §4`).
**But this foundation CANNOT claim semantic copy safety (SB7):** the current copy pipeline enforces
provenance/denylist, not semantic self-harm detection — so protected-copy denylist/required-review keys
are a **Copy-Pipeline successor**; do not assert the pipeline flags it here.

### S6 — Scope, activation, wire

Founder-scoped. The `soul` field already exists; **v20 ACTIVATES it (SB1)** under the pinned `soul`
artifact (chain: minigames 17 → pets 18 → fiscal 19 → soul 20), biconditional with floor ≥ 20, requires
v19 pinned, New-Founder-forward, Founder axis only (Company rejects it). History lives in the Founder
log/events, not the save. **Soul is dependency-blocked on Fiscal v19 being IMPLEMENTED (SB9)** — the
pure Soul catalog/band functions may land earlier, but v20 activation waits on v19. Intents (the
fixture-only debit component) / `touch_grass` activity /
`touch_grass`, resolved arms, receipts (naming before/after Soul, band, source/activity, curtain-copy
key), and events (`soul_price_paid.v1`, `soul_recovered.v1`, `soul_band_changed.v1` — the last emits
only when the derived band changes) are exact byte-grammar; Go/TS shared vectors compare state,
receipt, ordered event bytes, and the band-derivation at every threshold boundary and at the floor.

## Deviations from design

- **Model refinement (owner-ruled):** `design/02 §8` framed Soul loosely as a drain/recover ledger;
  this pins it as the meter-shown/currency-triggered hybrid (opt-in debits + opt-in recovery, no
  passive movement), and reconciles §8's touch-grass line (idle refill removed). No mechanic is added
  beyond the design; the drains/gates/ending are §8 verbatim.
- Mechanical naming (`pay_soul_price`, `touch_grass`, band enum) with flavor as localization keys.

## Acceptance criteria

1. Soul is a Founder-scoped bounded ledger, one authority, persists across Exit; rendered ONLY via
   the pet proxy (no wallet/currency UI anywhere — grep-proven absent).
2. Debits occur ONLY through the Soul-debit component inside an eligible owner transition (opt-in,
   curtain-tooltip key required); no public standalone debit rail and no passive drain path exist;
   full debit is required except the declared once-only exhaust case.
3. Recovery occurs ONLY through the ruled touch-grass activity lifecycle (opt-in, costs attended time,
   suppresses Company production, produces nothing); no passive
   idle refill exists (grep-proven); saturates at max.
4. Graduated gating derived (never stored): soft bands = flavor/greying only; hard locks bite only in
   the near-zero band; the pet proxy renders the band; the minigame/pet-interaction lock predicate is
   a pure read; byte-parity Go/TS at every band boundary.
5. Transcendence reads the dated `soul.depleted` fact plus current band → ending-variant enum. Copy
   semantic review is a successor obligation; this foundation does not claim the current copy gate
   proves it.
6. Founder save v20 activation under the pinned `soul` artifact, New-Founder-forward, requires v19
   pinned, Company axis rejects v20; migration + corpus; byte-parity wire vectors.

## Open questions (resolve before `accepted`)

- **Founder version ordering:** proposed v20 after Fiscal Quarters v19. If Soul and Fiscal Quarters
  implement out of order, the chain numbers swap — the acceptance review confirms the final order
  against the scalar-chain discipline (C36).
- **Drain-source dependency on the events engine:** the Faustian/crunch/longevity sources fire from
  the events engine (draft). This foundation ships the debit mechanism + a fixture source; confirm no
  production source ships before its event exists (activation-boundary discipline).
- **Trust↔Soul relationship:** the earlier meters work referenced a Trust↔Soul correlation gate
  (now unbuilt). Confirm whether Soul and the Trust meter interact mechanically or are fully
  independent axes (design: independent — Trust = to others, Soul = to self); if independent, this
  RFC needs no meters dependency.

## Acceptance-review blockers (2026-08-05)

The version order is correct: Fiscal v19 precedes Soul v20 on the scalar Founder axis. Soul and
Trust should remain mechanically independent at runtime. Production drain sources must not ship
before their Events/content owners; only a test fixture may exercise the generic mechanism here.
Those direction calls do not close SB1-SB9.

### SB1 — `soul` already exists, but activation/initialization is undefined

The Founder save has carried a signed `soul` field since the legacy schema, defaulting to zero and
accepting the whole exact-integer range. Version 20 therefore does not "add the ledger"; it changes
the meaning and valid domain of an existing dormant field. The RFC does not say how a pre-v20 zero
becomes the active initial value, or how legacy/new Founder genesis differs.

**Proposed contract:** pre-v20 Founder state must keep `soul == 0` and no Soul-owned state. At the
first Exit/New-Founder activation under a pinned Soul artifact, v20 initializes Soul to
`soul_initial` (required within `[floor,max]`, recommended equal to max). Version 20 requires the
artifact and validates the bounded domain; pre-v20 rejects any nonzero Soul. Record the activation
in the Founder Exit resolved arm so replay derives it from the pinned result bundle, never current
deployment data. Do not store event history in the save—the immutable Founder log/events already
own history.

### SB2 — A generic debit intent has no eligibility or atomic benefit boundary

`source_id` alone lets a client invoke any catalog source without proving that its contract/event/
longevity choice is currently offered. Debiting in one Founder command and granting the associated
benefit in a later Company/Event transition can partially commit or be retried independently. The
foundation would create a payment rail without an owned purchase.

**Owner ruling required:** recommended shape is a reusable Soul-debit component invoked inside the
owner transition (Events, longevity, contract) so benefit + debit commit atomically; no public
standalone `pay_soul_price` ships in production. A fixture-only Founder command may prove the pure
component. If the public intent remains, specify a server-authored offer/eligibility token,
single-use state, expiry, and the multi-stream transaction applying its benefit.

### SB3 — Floor saturation makes expensive deals cheaper at low Soul

If a source costs 10 and the founder has 3 above the floor, saturation grants the same downstream
benefit for 3. Repeating a source at the floor can cost zero. That conflicts with calling Soul a
price and can make the strongest low-Soul route free.

**Owner ruling required:** choose exact affordability. Recommended: require the full debit and
reject `unaffordable/soul` if it would cross the floor; a specifically declared `may_exhaust: true`
source may consume the exact remaining balance to the floor once, with loader-enforced single-use
eligibility. Saturation remains appropriate for recovery at max, not generic price payment.

### SB4 — `touch_grass` does not actually cost time or stop production

A Founder-only instant command cannot make a player spend attended time or forgo Company output.
Increasing the historical attendance counter would fabricate presence; merely reading it charges
nothing; and Company production continues unless a Company-owned pause is recorded. The proposed
intent therefore grants Soul for free despite the core recovery law.

**Owner ruling required:** define a real activity lifecycle and cross-stream coordinator. A sound
shape mirrors minigames: start an activity against the active Company, record a server-owned
attended start/duration and a Company production-suppression interval, then resolve after sufficient
Founder attended time in one Company+Founder transaction that grants Soul and records zero resource
output for that interval. Enumerate cancel/reconnect/Exit/concurrency behavior and prove the player
cannot run production or another exclusive activity simultaneously. This cannot be implemented as
the current single instantaneous `ApplyFounderLogged` command.

### SB5 — The pinned Soul artifact is not an exact grammar

The loader needs initial/floor/max, ordered band thresholds, debit-source definitions, recovery
activity definitions/durations, copy keys, one-use/exhaustion flags, gate policy, and ending enum
mapping. None of the exact keys, ordering, bounds, or cross-artifact references are stated.

**Proposed contract:** enumerate schema v1 with policy, ordered bands, source rows, and activity
rows. Production artifact rows may be empty until their owner RFCs; tests use a fixture artifact
that is never epoch-seeded. Exact-key Go/TS loaders enforce sorted unique IDs, threshold partition
of the full domain, positive safe-integer amounts/durations, verified copy keys, and only registered
event/minigame/pet references. Add `soul` to the single artifact authority; v20 presence is
biconditional and requires the exact v19 Fiscal artifact.

### SB6 — Pet/minigame gating and visibility are claimed beyond implemented surfaces

Pet care can read Founder Soul, but the minigame start path and future UI need explicit consumers.
"Rendered only via the pet proxy" also risks hiding the exact meter, contradicting design/02's law
that moral meters are visible/published. The Game UI stage is not implemented, so grep cannot prove
the final presentation in this foundation.

**Proposed contract:** ship a pure Soul-band projection plus one closed `human_content_locked`
predicate. Pet care keeps recovery/essential-care paths available at the floor; hobby minigame
session start rejects near-zero through a named resolver, while unrelated minigames remain governed
by catalog classification. The public pet panel exposes the exact bounded meter and band/reason,
but never lists Soul in currency/shop wallets. Split AC1/AC4 into foundation-level projection/gate
proof and later Game-UI rendering proof; do not claim the latter early.

### SB7 — The ending hook conflicts with the dated-facts gating law

Design/02 section 7 says endings key on dated ledger facts, never meter thresholds, while S5 selects
the ending directly from the current Soul band. Section 8 specifically names zero Soul, so the
documents need one authoritative reconciliation. The existing copy pipeline also does not
semantically detect self-harm/suicide; it enforces provenance/denylist/known-red terms, so AC5's
"copy-pipeline flag proven" is currently false.

**Owner ruling required:** recommended reconciliation: emit an immutable `soul.depleted` Founder
fact/event the first time an opt-in debit reaches the exact floor; Transcendence keys on that dated
fact plus current band as explicitly ruled, rather than an arbitrary threshold. Separately add
specific protected-copy denylist/required-review keys in a Copy Pipeline successor; this foundation
may export the ending enum but cannot claim semantic copy safety it cannot prove.

### SB8 — Trust independence still leaves a correlation-gate obligation

No runtime formula should couple Trust and Soul, so no Meters package import is needed. But design/02
explicitly requires a CI correlation gate preventing content authors from making the axes move
together. Declaring full independence without assigning that gate silently drops a design law.

**Proposed contract:** state runtime independence as normative and route a balance-harness successor
that measures signed Trust/Soul deltas across declared content choices, with a catalog threshold and
fail-closed fixture. It activates when the first production Soul source lands; fixture-only sources
in this RFC demonstrate axis separation but do not pretend to establish live correlation.

### SB9 — Wire/events/rejections and upstream sequencing are incomplete

Intent/resolved/receipt/event exact keys are not enumerated, nor are band-change ordering, actual
credited/debited delta under caps, activity lifecycle events, or existing rejection categories.
Implementation also cannot satisfy the v20 biconditional until Fiscal v19's artifact grammar and
activation are ruled and implemented.

**Proposed contract:** after FQ v19 closes, enumerate strict v20 codec/migration and every command/
resolved/receipt/event object; reuse the existing rejection taxonomy; emit `soul_band_changed.v1`
after the debit/recovery event only on a real derived-band change; compare raw state, receipt, and
ordered event bytes in Go/TS. Keep status dependency-blocked until Fiscal is implemented, even if the
pure Soul catalog/band functions land earlier.

## Changelog

- 2026-08-05: created (draft) — Wave-A foundation; the meter-shown/currency-triggered Soul hybrid
  from `soul-mechanic.md` + owner rulings; reconciles `design/02 §8` touch-grass recovery.
- 2026-08-05: Codex acceptance review — v19->v20 ordering, runtime Trust independence, and
  fixture-only source direction confirmed; blocked on SB1-SB9, chiefly existing-field activation,
  atomic debit ownership, real opportunity-costed recovery, and honest gate/copy contracts.

## Owner rulings on SB1-SB9 (2026-08-06)

All accepted. Four owner-calls (SB2/SB3/SB4/SB7). Several fix real errors in my draft (the standalone
debit rail, the free instant recovery, the "hidden meter" vs the moral-meter transparency law).

- **SB1 — accepted; v20 ACTIVATES the existing `soul` field, does not add it.** Pre-v20 keeps
  `soul == 0` and no Soul-owned state (rejects nonzero). At the first Exit/New-Founder activation under
  the pinned Soul artifact, v20 initializes Soul to `soul_initial` (required in `[floor,max]`,
  recommended = max) and validates the bounded domain; the activation is recorded in the Founder Exit
  resolved arm so replay derives it from the pinned result bundle (never deploy data). History stays in
  the Founder log/events — NOT the save. (S1/S6 reconciled.)
- **SB2 — RULED: Soul-debit is a COMPONENT inside the owner transition, NOT a standalone rail.** A
  generic public `pay_soul_price {source_id}` would be a payment rail with no owned purchase and could
  partially commit across streams. **No standalone `pay_soul_price` ships in production.** Instead a
  reusable **Soul-debit component** is invoked INSIDE the owner transition (an Event / longevity /
  contract) so the benefit and the debit commit ATOMICALLY; the source's offer/eligibility is proven by
  that owner (the events engine etc., successors). A **fixture-only** Founder command exercises the pure
  component for tests. (S2 reconciled — the standalone intent becomes a fixture-only component.)
- **SB3 — RULED: full-debit-or-reject; no cheap-at-floor exploit.** A Soul price must be FULLY
  affordable — reject `unaffordable/soul` if the debit would cross the floor (floor-saturation would let
  a low-Soul founder buy the same benefit for pennies, and repeat-at-floor for free). The ONLY exception
  is a source explicitly declared `may_exhaust: true`, which may consume the exact remaining balance to
  the floor ONCE, loader-enforced single-use. Saturation stays for recovery-at-max only, never price
  payment.
- **SB4 — RULED: `touch_grass` is a REAL activity lifecycle with production suppression (not an instant
  command).** An instant Founder command grants Soul for free — it can't make the player forgo output.
  Recovery mirrors the minigame session: **start** an activity against the active Company (server-owned
  attended start + duration + a Company **production-suppression interval**), **resolve** after
  sufficient Founder attended time in ONE Company+Founder transaction that grants Soul AND records ZERO
  resource output for that interval. Enumerate cancel/reconnect/Exit/concurrency; **prove the player
  cannot run production or another exclusive activity simultaneously** (the opportunity cost is real —
  the whole point of the recovery ruling). This is NOT the single instantaneous `ApplyFounderLogged`.
  (S3 reconciled.)
- **SB5 — accepted (Soul artifact grammar).** Schema v1: policy, ordered bands, source rows, activity
  rows; production rows may be empty until their owner RFCs; tests use a fixture artifact NEVER
  epoch-seeded. Exact-key Go/TS loaders enforce sorted-unique IDs, a band-threshold PARTITION of the
  full domain, positive safe-int amounts/durations, verified copy keys, and only registered
  event/minigame/pet references. `soul` joins the artifact authority; v20 presence biconditional,
  requires the exact v19 Fiscal artifact.
- **SB6 — accepted; the EXACT meter IS visible (fixes my "hidden meter" error).** Ship a pure Soul-band
  projection + one closed `human_content_locked` predicate. Pet care keeps recovery/essential-care
  available at the floor; hobby-minigame session-start rejects near-zero via a named resolver; unrelated
  minigames stay governed by catalog classification. **The public pet panel exposes the EXACT bounded
  meter + band/reason** — moral meters are visible/published (the transparency law; my "never rendered
  as a number" was wrong) — but Soul is **never** a currency/shop wallet. Split AC1/AC4 into
  foundation-level projection/gate proof and a later Game-UI rendering proof. (S1/S4 reconciled: exact
  meter shown+published, not a wallet.)
- **SB7 — RULED: the ending keys on a DATED FACT, not a live band (reconciles the §7 dated-facts law).**
  Emit an immutable `soul.depleted` Founder fact/event the FIRST time an opt-in debit reaches the exact
  floor; **Transcendence keys on that dated fact plus current band** — never an arbitrary live
  threshold (design/02 §7). Copy safety: the current pipeline enforces provenance/denylist, NOT semantic
  self-harm detection, so this foundation exports the ending ENUM but **cannot claim semantic copy
  safety**; protected-copy denylist/required-review keys are a Copy-Pipeline successor. (S5/AC5
  reconciled — I over-claimed "copy pipeline flags it".)
- **SB8 — accepted; the Trust↔Soul correlation gate is ROUTED, not dropped.** Runtime independence is
  normative (no Meters import). The design-required CI correlation gate (preventing content authors from
  coupling the axes) is a **balance-harness successor**: it measures signed Trust/Soul deltas across
  declared content choices with a catalog threshold + fail-closed fixture, activating when the first
  production Soul source lands. Fixture-only sources here demonstrate separation but don't establish
  live correlation.
- **SB9 — accepted; Soul is dependency-blocked on Fiscal v19 implementation.** After FQ v19 closes,
  enumerate the strict v20 codec/migration + every command/resolved/receipt/event object; reuse the
  existing rejection taxonomy; emit `soul_band_changed.v1` after a debit/recovery only on a REAL derived
  band change; Go/TS byte compare. **Status stays dependency-blocked until Fiscal is implemented** (the
  pure Soul catalog/band functions may land earlier, but v20 activation waits on v19).

SB1-SB9 fully ruled. Production drain-sources (events/longevity/contracts) and recovery activities are
successors gated on their owner engines; the correlation gate + protected-copy keys are named
successors; numbers (bands/costs/durations) are data.

## Implementation acceptance recheck (2026-08-06) — blocked SB10-SB16

SB1-SB9 choose the intended Soul model, but the shipped Founder/save/minigame/Company transition
surfaces cannot implement that model without seven further contracts. Fiscal remains a real ordering
dependency, but waiting for Fiscal would not resolve these Soul-specific gaps. No Soul mechanic code
started.

### SB10 — The Soul artifact is not an exact schema, so even the “pure” layer cannot land

SB5 names policy, bands, source rows, and activity rows, but gives no exact nested keys, ending enum,
human-content classification, copy-key fields, source-owner binding, or null/empty rules. “Threshold
partition” also has no interval convention, so two loaders can disagree at every boundary.

**Proposed contract:** exact schema-v1 root
`{schema_version:1,policy,bands,debit_sources,recovery_activities,ending_policy}`. Policy is
`{soul_floor,soul_initial,soul_max}`. Bands are a raw-byte-order-independent but value-ordered complete
array of exact `{band_member,min_inclusive,max_inclusive,human_content_locked,reason_key}` using the
closed members `whole|dimming|hollow|near_zero`; intervals are contiguous, non-overlapping, and cover
`[floor,max]`, with only `near_zero` locked. Debit sources are raw-byte-sorted exact
`{source_id,owner_kind,amount,may_exhaust,single_use,curtain_copy_key}` with the biconditional
`may_exhaust == single_use`; fixture owner kind is allowed only outside epoch-seeded artifacts.
Recovery activities are raw-byte-sorted exact
`{activity_id,duration_attended_ms,recovery_amount,reason_key}`. Ending policy is exact
`{whole_variant,depleted_variant}` over closed mechanical IDs. Safe-integer/domain, exact-key,
sorted/unique, verified-copy-key, owner-kind, and full-partition checks run byte-identically in Go/TS.

### SB11 — Once-only exhaust sources require state that SB1 currently forbids

SB3 says a `may_exhaust` source may consume to the floor once, but neither the Founder save nor an
owner table records that the source was used. Loader validation can prove a row is declared
single-use; it cannot prove a founder has not used it. Reusing immutable ledger facts as mutable
eligibility would require every owner transition to scan history and would make the save snapshot
insufficient for replay.

**Owner ruling required:** persist a raw-byte-sorted `soul_exhausted_source_ids` set in Founder v20,
limited to catalog rows with `may_exhaust:true`; the debit component inserts on the same atomic
transition and rejects `not_eligible/soul_source_consumed` thereafter. This is current eligibility
state, not a second Soul balance/history authority. If owner-specific state instead owns single use,
enumerate that state and freeze it into the debit component's resolved input; do not claim a generic
component can enforce it alone.

### SB12 — The debit component has no exact API, evidence, or dated-fact ordering

“Invoked inside the owner transition” is the correct boundary but not a callable contract. It leaves
before/after band derivation, actual debit, eligibility evidence, source lookup, the curtain key, and
first-depletion detection open. The fixture-only command is also undefined: adding it to the public
intent parser would accidentally ship the forbidden standalone rail.

**Proposed contract:** export a pure
`ApplyDebit(state,artifact,{source_id,owner_kind,eligibility_ref}) -> {soul_before,debit,
soul_after,band_before,band_after,curtain_copy_key,depleted_first_time}` component. It validates source
owner, full affordability, once-only state, and canonical bands; the owner transition supplies and
persists its own eligibility reference and benefit atomically. Emit the owner's benefit event first,
then `soul_price_paid.v1`, optional `soul_band_changed.v1`, and, only on first exact-floor debit,
`soul_depleted.v1`; insert `soul.depleted` in Founder `LedgerFactKinds` in that same transition.
Exercise the component through a package-private simulation/test entrypoint, never a production parser
kind. Enumerate exact result/event payload keys and rejection details before an owner integrates it.

### SB13 — Touch-grass suppression has no persistence owner or Company-version contract

SB4 requires a Company production-suppression interval and cross-stream resolution, but Soul is
otherwise declared Founder-v20/Company-untouched. An in-memory interval vanishes on reconnect; a
Founder-only field cannot stop Company accrual; a database activity row beside an unchanged Company
save is an ambient read outside `ApplyLogged`. The current production engine accrues the entire gap at
the next command and has no “zero-output segment” input.

**Owner ruling required:** choose the authoritative activity/suppression shape. Recommended: a
Postgres `soul_recovery_sessions` lifecycle patterned on minigames, plus an immutable resolved
suppression segment frozen into every affected Company command's `replay_inputs`. Start and resolve
are explicit Company+Founder transactions under the established lock order; while active, ordinary
Company intents reject `not_eligible/exclusive_activity`, and resolve advances `evaluated_through`
over the suppression interval with zero ledger output before granting Founder Soul. State whether this
requires Company v19 (after Active-Play v18) or is completely represented by the session+resolved
input; enumerate the resulting activation dependency. Without this ruling, “Founder axis only” and
“Company production suppression” contradict each other.

### SB14 — Recovery lifecycle semantics are still only a requirement list

Start/resolve/cancel/reconnect/Exit/concurrency are required but not decided. There is no answer for
early resolve, duplicate resolve, reconnect after completion, cancel output, Exit while active,
whether attended time pauses offline, or racing an ordinary Company command with resolve.

**Proposed contract:** start creates exactly one active session per founder and records Founder and
Company stream/revision/run coordinates, activity ID, attended start, required duration, and claim
token state. It rejects if another exclusive session exists. Attended time uses the race-safe Founder
attendance resolver and pauses offline. Early resolve rejects without mutation; eligible resolve
claims then locks Company→Founder in the declared order, validates unchanged run/session identity,
commits the zero-output suppression segment, saturating recovery, both replay logs, receipts/events,
and terminal session state atomically. Identical retry replays; expired claim leases are recoverable.
Cancel is allowed before resolution, grants zero Soul, ends suppression at the cancel coordinate, and
never backfills output. Exit rejects while active (recommended; avoids a three-outcome terminal
transaction) unless a successor explicitly defines cancel-on-Exit. Reconnect resumes the same row.

### SB15 — Human-content gating has no catalog classification or composed resolver

SB6 says hobby minigame starts lock at near-zero while unrelated minigames do not. The current
minigame catalog has no ruled Soul-gating classification, and its session service cannot read Founder
state through a named resolver. Pet essential/recovery exceptions likewise need a closed action
classification; “keeps available” is not loadable data.

**Proposed contract:** extend the pinned minigame artifact with a closed
`soul_gate: human_hobby|unrelated` field and the pet action rows with
`soul_gate: essential|recovery|ordinary`. The Soul package exports only
`HumanContentLocked(soul,artifact)` and band projection. A composed Founder-state resolver supplies
the pinned Soul snapshot/bundle to minigame start and pet eligibility; human-hobby and ordinary pet
actions reject `not_eligible/human_content_locked`, while unrelated, essential, and recovery paths
remain available. Artifact cross-validation refuses gated consumers unless the same bundle pins Soul.
These schema additions are balance changes owned by their respective RFCs, not silently added here.

### SB16 — v20 activation and exact wire/event grammar remain intentionally deferred, not executable

SB9 says “after Fiscal closes, enumerate” rather than enumerating. The Founder codec needs exact v20
key rules (including SB11 if accepted), activation evidence for dormant `soul`, replay-input version,
receipt/event bytes, and New-Founder behavior. The current scalar chain only supports Founder v18.

**Proposed contract:** after Fiscal v19 is implemented, Founder v20 retains the existing `soul` key
but changes its validation to `[floor,max]` and appends only the ruled eligibility/activity-owned state.
Soul artifact presence is biconditional with Founder floor 20 and requires fiscal+minigames+pets;
Company rejects v20. Exit/New-Founder activation resolves `{soul_initial,band_member}` from the pinned
bundle into the Founder log; pre-v20 rejects nonzero Soul. Enumerate exact start/cancel/resolve and
owner-debit resolved/receipt/event objects, event registry/DB migration, raw Go/TS save corpus, and a
sequential Founder replay fixture crossing every band and the first-depletion fact. Until Fiscal and
SB10-SB15 close, SB16 cannot honestly be implemented.

Until SB10-SB16 and Fiscal F9-F15 are ruled, only conceptual package boundaries are known. Shipping a
catalog or activity now would invent persistence and cross-stream semantics despite the explicit
DESIGN-GAP law.

## Owner rulings on SB10-SB16 (2026-08-06) — exact schema/state/lifecycle/wire under SB1-SB9

All accepted (Codex's proposed contracts are sound). Owner-calls: SB11, SB13.

- **SB10 — accepted (exact Soul artifact).** Root `{schema_version:1, policy, bands, debit_sources,
  recovery_activities, ending_policy}`. `policy {soul_floor, soul_initial, soul_max}`. `bands` = a
  value-ordered COMPLETE array of `{band_member, min_inclusive, max_inclusive, human_content_locked,
  reason_key}` over closed members `whole|dimming|hollow|near_zero`, contiguous + non-overlapping +
  covering `[floor,max]`, ONLY `near_zero` locked. `debit_sources` raw-byte-sorted `{source_id,
  owner_kind, amount, may_exhaust, single_use, curtain_copy_key}` with the biconditional `may_exhaust ==
  single_use`; fixture owner_kind only outside epoch-seeded artifacts. `recovery_activities`
  raw-byte-sorted `{activity_id, duration_attended_ms, recovery_amount, reason_key}`. `ending_policy
  {whole_variant, depleted_variant}` over closed IDs. All checks byte-identical Go/TS.
- **SB11 — RULED (once-only eligibility state).** Persist a raw-byte-sorted `soul_exhausted_source_ids`
  set in Founder v20, limited to catalog rows with `may_exhaust:true`; the debit component INSERTS on
  the same atomic transition and rejects `not_eligible/soul_source_consumed` thereafter. This is
  current-eligibility state, NOT a second Soul balance/history authority (history stays in the Founder
  log).
- **SB12 — accepted (the debit component API).** Export a pure `ApplyDebit(state, artifact, {source_id,
  owner_kind, eligibility_ref}) -> {soul_before, debit, soul_after, band_before, band_after,
  curtain_copy_key, depleted_first_time}`. It validates owner, full affordability, once-only state, and
  canonical bands; the OWNER transition supplies + persists its eligibility reference and benefit
  atomically. Event order: the owner's benefit event FIRST, then `soul_price_paid.v1`, optional
  `soul_band_changed.v1`, and ONLY on the first exact-floor debit `soul_depleted.v1` + insert
  `soul.depleted` into the Founder `LedgerFactKinds` in that same transaction (SB7's dated fact).
  Exercise via a package-private test entrypoint — NEVER a production parser kind (that would ship the
  forbidden standalone rail).
- **SB13 — RULED (touch-grass suppression: NO Company save-version bump — the Fiscal-F12 pattern).** A
  Postgres `soul_recovery_sessions` lifecycle (patterned on minigames, a SEPARATE table — NOT Company
  save state) plus an immutable resolved **suppression segment frozen into the affected Company
  command's `replay_inputs`**. So — exactly like Fiscal F12's frozen contributions — the Company save
  axis is UNTOUCHED (no v19): the session is a separate table, the zero-output segment is a
  replay-input, and resolve advances the EXISTING `evaluated_through` over the suppression interval. While
  a session is active, ordinary Company intents reject `not_eligible/exclusive_activity` (a
  session-existence check, not save state). Start + resolve are Company+Founder transactions under the
  **ratified Founder-then-Company lock order** (C38).
- **SB14 — accepted (lifecycle); lock order corrected to Founder-then-Company.** Start creates exactly
  ONE active session per founder (records Founder+Company stream/revision/run coordinates, activity ID,
  attended start, required duration, claim-token state) and rejects if another exclusive session exists.
  Attended time uses the race-safe Founder attendance resolver and PAUSES OFFLINE. Early resolve rejects
  without mutation; eligible resolve claims, then locks **Founder-then-Company** (the ratified order —
  correcting SB14's "Company→Founder" wording), validates unchanged run/session identity, and atomically
  commits the zero-output suppression segment, saturating recovery, both replay logs, receipts/events,
  and terminal session state. Identical retry replays; expired claim leases recover. Cancel (before
  resolution) grants ZERO Soul, ends suppression at the cancel coordinate, never backfills output. Exit
  REJECTS while a session is active (recommended — avoids a three-outcome terminal transaction) unless a
  successor defines cancel-on-Exit. Reconnect resumes the same row.
- **SB15 — accepted (human-content gating classification).** Extend the pinned minigame artifact with a
  closed `soul_gate: human_hobby|unrelated` field and pet action rows with `soul_gate:
  essential|recovery|ordinary`. The Soul package exports ONLY `HumanContentLocked(soul, artifact)` +
  band projection. A composed Founder-state resolver supplies the pinned Soul snapshot/bundle to
  minigame-start and pet eligibility; at `near_zero`, `human_hobby` minigames and `ordinary` pet actions
  reject `not_eligible/human_content_locked`, while `unrelated`, `essential`, and `recovery` paths remain
  available. Artifact cross-validation refuses gated consumers unless the same bundle pins Soul.
- **SB16 — accepted (v20 activation + wire); stays dependency-blocked on Fiscal v19 impl.** After Fiscal
  v19 is implemented, Founder v20 RETAINS the existing `soul` key but changes its validation to
  `[floor,max]` and appends only the ruled eligibility (`soul_exhausted_source_ids`) / activity-owned
  state. Soul artifact presence biconditional with Founder floor 20, requires fiscal+minigames+pets;
  Company rejects v20. Exit/New-Founder activation resolves `{soul_initial, band_member}` from the pinned
  bundle into the Founder log; pre-v20 rejects nonzero Soul. Enumerate the exact start/cancel/resolve +
  owner-debit resolved/receipt/event objects, event registry/DB migration, Go/TS save corpus, and a
  sequential Founder replay fixture crossing every band + the first-depletion fact.

SB10-SB16 fully ruled. S1-S6 refined (not contradicted). Soul implementation remains DEPENDENCY-BLOCKED
on Fiscal v19 being implemented; the pure catalog/band/`HumanContentLocked` layer may land earlier.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
