# RFC: Pet Care Foundation

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/04 §1` (the tamagotchi layer — 4-stat decay, care actions + diminishing returns, personality behavior FSM, trust/mood two-tier, bonds), `design/04 §Neopets adoptions` (no-death canon, public-awkwardness-not-loss), `design/03 §10` / combat C5 (the pet is the On-Call Leader — the seam awaits its two integers)
- **Research:** `cattery-reusables.md` (the port source: decay + care + FSM + CSS-sprite tech), `neopets-systems.md §3` (no-death, care barely punishes), `creature-battler.md §8.3` (care→options-not-stats)
- **Depends on:** Save + Production + Run Genesis (implemented — pet state is Founder-scoped, care actions are intents inside `ApplyLogged`); Combat Shared Kernel (implemented — this RFC produces the `(trust_ppm, soul)` integers combat's C5 already consumes)
- **Owner ruling honored:** breadth-first — the care/trust/mood/FSM MECHANICS, not pets' content (species, cosmetics, the battle content).
- **Planning:** `planning/pet-care-foundation/` (once implementing)

## Summary

The cattery port, made deterministic and server-authoritative. Care stats with diminishing-return
actions, the two-tier trust/mood model, the personality behavior FSM, and bonds — all as
Founder-scoped state mutated by care intents inside the replay boundary. Critically: this RFC
CLOSES combat's C5 fixture-only boundary by producing the real `(trust_ppm, soul)` inputs the duel
and lane engines consume.

## Specification

### PC1 — Care stats & the no-death law

Four stats per pet (the cattery four: hunger/energy/cleanliness/affection — final names content),
each ppm, decaying on ATTENDED time toward a low band (the attended clock, like everything).
**No-death (the Neopets canon, ruled):** a fully-neglected stat floors and STAYS floored — the pet
never dies, never leaves; neglect costs the pet's public status display (guild-visible), the FSM
mood, and greyed-out care options, never the pet. A dormant founder returns to a living pet
(the researched retention property).

### PC2 — Care actions with diminishing returns

Care actions are intents (`care_action {pet, action}` — C1, evented, replay-logged). Each raises a
stat, with the cattery diminishing-returns curve: effectiveness scales down as the stat rises
(`>90% → 0.5×`, the researched shape), so over-care wastes — care buys OPTIONS and TEMPO, never a
higher ceiling (the hardcap decision: every pet of an identity shares the stat ceiling). Actions
cost attended time / manual tokens, never a spendable meter. Diminishing-return math is
integer-ppm, byte-parity both runtimes.

### PC3 — Trust & mood (two-tier, the combat seam)

**Trust** (persistent, server, decays slowly toward neutral absent care — the cattery two-day
decay generalized to attended time) and **Mood** (session, volatile). Trust is the durable
relationship; mood is the moment. **This is combat's C5 input:** `trust_ppm` (from care quality
over time) → the Obedience curve (50%→30% across Trust 1.00→0.80, already specced in the combat
kernel); `soul` (from the Meters foundation) → the leader-ability modulation. This RFC's output is
exactly the two integers combat froze as fixtures — combat's fixture boundary becomes real.

### PC4 — The personality behavior FSM

Each pet has a personality (the six Temperaments — shared with combat's type layer, one identity
system) driving a behavior state machine: activity → behavior-chain queue (the cattery FSM
ported). Behaviors are DETERMINISTIC given (personality, stats, seed) — drawn from the save-seeded
stream, replay-safe. Behavior is display + flavor (it drives the pet's visible activity and the
On-Call Leader's in-match reactions), never authoritative economy. Bonds/affinity between pets are
a Founder-state graph (the cattery bond system), affecting behavior only.

### PC5 — Scope & the CSS-sprite tech

Pets are FOUNDER-scoped (they persist across Exit — a pet is not a company asset; the founder's
companion). Care state, trust, bonds all persist; save-version bump, corpus fixtures. The
CSS-recolorable sprite system (cattery tech, `cattery-reusables.md`) is the rendering approach —
declared here as the pet-visual contract for the UI foundation to consume, not implemented as
screens.

## Acceptance criteria

1. Care stats decay on attended time toward the low band (offline-span fixture: no decay while
   away); no-death (a floored stat floors, pet persists, options grey — never removal).
2. Care actions: diminishing-returns curve byte-parity Go/TS; over-care wastes; the ceiling is
   the shared identity hardcap (care can't exceed it).
3. Trust/mood: trust decays slowly, mood is volatile; **the produced `(trust_ppm, soul)` drive a
   combat duel's Obedience exactly as the combat C5 golden vectors specify** — combat's fixture
   boundary replaced by this RFC's real output, cross-verified.
4. FSM: behavior deterministic given (personality, stats, seed), replay-safe; bonds affect
   behavior only, never economy.
5. Founder-scope persistence across Exit; migration + corpus; the sprite contract typed for UI.

## Open questions

- Pet acquisition (how you get your first pet, breeding/rarity) — content, later RFC; this
  foundation assumes one starter pet exists.
- Pet battles' full content (rosters, seasons) — the combat engines own the mechanics; this
  produces their care inputs only.

## Acceptance blockers (Codex review, 2026-08-04)

The design direction is coherent, but the draft cannot yet be accepted without inventing a new
Founder mutation/replay boundary and most of the state machine. The blockers below separate
mechanical foundation from later pet content while preserving the no-death and options-not-power
laws.

### C1 — Founder-scoped care cannot run inside Company `ApplyLogged`

The summary and dependency line put Founder pet mutations “inside `ApplyLogged`,” but that boundary
commits only the Company stream and receives Founder carry read-only. Meters C3 already ruled the
same topology for Soul: an ordinary Company command cannot mutate Founder state. A care action
implemented as described would either be a hidden cross-stream write or replay only half its
effects.

**Proposed contract:** introduce one named Founder mutation boundary with canonical command,
resolved-input, receipt, and event envelopes, plus idempotency and immutable log rows matching the
Company boundary. Care commands lock and commit the Founder stream only. Any action that genuinely
needs Company state becomes a declared multi-stream successor; Phase-A care reads no live Company
state and spends no Company resource/manual token. Founder replay owns pet state independently of
run replay.

### C2 — Pet identity and starter creation are missing authority

The RFC assumes a starter pet but supplies no persisted pet key, species/personality catalog,
ownership rule, or creation transaction. “Final names content” also leaves the four stat keys
unstable even though saves and commands must name them.

**Proposed contract:** fix the mechanical stat IDs as `hunger|energy|cleanliness|affection`; define
an immutable pet record `{pet_id, founder_id, species_id, temperament, created_at}` and mutable care
state keyed by `pet_id`. A fixture-only starter definition is available to tests. Production
starter creation activates only with a pinned pet catalog and occurs atomically with New Founder
(or through a separately ruled acquisition command); no loader invents a species or temperament.

### C3 — Care actions and decay are shapes without arithmetic

PC1/PC2 provide one breakpoint (`>90% → 0.5×`) but no exact decay grid, low-band floor, action
rows, affected-stat deltas, secondary effects, cooldown/eligibility grammar, simultaneous-update
order, rounding, saturation events, or reason keys. “Costs attended time / manual tokens” is both
ambiguous and cross-scope.

**Proposed contract:** make decay and actions strict catalog rows. The schema fixes a ppm stat
domain, absolute Founder-attended grid, per-stat decay numerator/denominator/remainder, floor, and
simultaneous commit order. Each action declares exact signed stat deltas, eligibility predicate,
diminishing-return piecewise-linear table, cooldown, and visible no-op/saturation reason keys.
Actions never spend time or Company manual tokens; their only scarcity is the catalog cooldown and
the diminishing-return curve. Numeric values remain balance data and join epoch identity before
production activation.

### C4 — Founder attended time has no live persisted cursor

Pet decay and trust use attended time across runs, but Founder `AgeMS` advances at Exit while care
must evaluate during a run. The draft does not define how a Founder-only transaction obtains the
current run's attended delta without reading mutable Company state, how multiple founders/runs are
ordered, or how partition invariance survives Exit.

**Proposed contract:** name one Founder-owned monotonic `pet_attended_ms` authority advanced only by
an explicit, idempotent settlement from the Company stream (with source stream/revision watermark),
then run decay/FSM grids from that value. Alternatively, make care a declared Founder+Company
transaction with the existing lock order. In either shape, specify the Exit handoff and prove
`advance(a+b) == advance(advance(a),b)` across a mid-interval Exit.

### C5 — Trust, mood, and combat output are not closed state

The draft gives prose ranges (“Trust 1.00→0.80”), calls mood session-volatile, and says Soul
modulates combat, but supplies no stored fields, reset rule, trust gain/decay equation, mood state
union, or exact output function. `trust_ppm` also risks confusion with the ten constituency Trust
meters unless mechanically namespaced.

**Proposed contract:** persist `pet_trust_ppm` on Founder pet state with `[0,1_000_000]`, catalog
initial/neutral/floor/cap, exact care gains, and fixed-grid decay/remainder. Define mood as a closed
derived enum over current care/FSM state, not a second authority, unless a resettable persisted
field is justified. Export one pure `CombatInputs(pet_state, founder_soul)` returning exact-safe
`{pet_trust_ppm,soul}`; the combat catalog alone maps those integers to Obedience. Until the duel
engine exists, conformance ends at shared vectors rather than claiming a live duel.

### C6 — The behavior FSM has no transition contract

“Activity → behavior-chain queue” does not define states, events, queue bounds, transition timing,
PRNG labels/draw order, or what becomes persisted. Without those, behavior can diverge across
replay or grow an unbounded save. The bond graph also has no mutation owner and would require
cross-founder/guild writes.

**Proposed contract:** enumerate a closed behavior-state union and bounded queue envelope, absolute
Founder-attended transition grid, deterministic candidate ordering, labeled PRNG substreams, and
persisted cursor/seed/remainder. Behavior may emit presentation facts only. Defer bonds to a
successor that owns cross-founder/guild interactions; Phase A may keep a fixture-only affinity
input but stores no writable bond graph.

### C7 — No-death visibility and greyed options need explicit projections

PC1 makes neglect guild-visible and says care options grey out, but does not name the privacy
surface, event/read model, or eligibility reason returned to clients. A Founder save must not
become directly queryable merely because a UI promise mentions guilds.

**Proposed contract:** care command rejection uses a closed detail discriminator and copy key;
snapshot output exposes only the pet's current display band and eligible-action IDs. Guild-visible
status is deferred to Feed/Social or Guild projection through an explicit presentation event,
with no raw stat values. The no-death invariant is enforced in the Founder transition and tested
independently of any projection.

### C8 — Persistence, replay, and activation are unnamed

PC5 says “save-version bump, corpus fixtures” but pet data is Founder-scoped, not part of the
Company save-version migration pattern. It does not name tables/JSON, schema version, replay
bundle, TS boundary, or how old pinned epochs activate the new mechanic.

**Proposed contract:** define a versioned Founder pet snapshot and immutable Founder command log;
pin the pet catalog/hash on pet creation and carry it in replay inputs. Apply the established
activation law: old founders/runs do not synthesize pet state from deploy-current catalogs;
production starter creation begins only at the first epoch carrying the pet artifact. Supply
Go/TS canonical-state and transition vectors for decay, care, trust, FSM, Exit handoff, and
no-death saturation before implementation status changes to accepted.

## Changelog

- 2026-08-04: Codex acceptance review found C1–C8. Founder care cannot mutate inside the
  Company-only `ApplyLogged` boundary; pet identity, care/time arithmetic, trust/combat output,
  behavior FSM, projection privacy, and Founder replay/activation require executable contracts.
- 2026-08-03: created (draft) — the cattery port deterministic + server-authoritative; closes
  combat C5's fixture-only `(trust_ppm, soul)` boundary with real output.
