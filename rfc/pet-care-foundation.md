# RFC: Pet Care Foundation

- **Status:** accepted (C1-C8 ruled; introduces the Founder mutation boundary; implementing)
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/04 §1` (the tamagotchi layer — 4-stat decay, care actions + diminishing returns, personality behavior FSM, trust/mood two-tier, bonds), `design/04 §Neopets adoptions` (no-death canon, public-awkwardness-not-loss), `design/03 §10` / combat C5 (the pet is the On-Call Leader — the seam awaits its two integers)
- **Research:** `cattery-reusables.md` (the port source: decay + care + FSM + CSS-sprite tech), `neopets-systems.md §3` (no-death, care barely punishes), `creature-battler.md §8.3` (care→options-not-stats)
- **Depends on:** Save + Run Genesis (implemented); Combat Shared Kernel (implemented). **NOT the Company `ApplyLogged` path — C1: this RFC introduces the Founder mutation boundary.**
- **Owner ruling honored:** breadth-first — the care/trust/mood/FSM MECHANICS, not pets' content (species, cosmetics, the battle content).
- **Planning:** `planning/pet-care-foundation/` (once implementing)

## Summary

The cattery port, made deterministic and server-authoritative. Care stats with diminishing-return
actions, the two-tier trust/mood model, the personality behavior FSM, and bonds — all as
Founder-scoped state mutated by care intents through `ApplyFounderLogged`. Critically: this RFC
CLOSES combat's C5 fixture-only boundary by producing the real `(trust_ppm, soul)` inputs the duel
and lane engines consume.

## Specification

### PC1 — Care stats & the no-death law

Four mechanically fixed stats per pet (`hunger`, `energy`, `cleanliness`, `affection`),
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
spend no Company resource, attended time, manual token, or meter. Catalog cooldowns and
diminishing returns are their only scarcity. Diminishing-return math is integer-ppm, byte-parity
in both runtimes.

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

## Owner rulings on C1-C8 (2026-08-04) - introduces the FOUNDER MUTATION BOUNDARY

- **C1 - accepted, and I repeated the Meters-C3 mistake:** I wrote Founder-scoped care as mutating
  inside Company `ApplyLogged`, which commits the Company stream only. **Ruling: this RFC
  introduces `ApplyFounderLogged` - the Founder-scoped mirror of the Company transition boundary,
  the project's FIRST Founder-scoped intent surface.** Same discipline as Company: canonical
  command, resolved-input closed union, receipt, event envelopes, idempotency, immutable Founder
  run-log rows, and a Founder-replay path owning pet state INDEPENDENTLY of run replay. Care
  commands lock+commit the Founder stream ONLY, read no live Company state, spend no Company
  resource/token. Company-needing actions are declared multi-stream successors (none in Phase A).
  This boundary is REUSABLE - every future Founder-scoped mechanic (Meters-C3 Soul verbs,
  live-writing achievement surfaces) uses it instead of the Exit-only multi-stream path.
- **C2 - accepted:** stat IDs FIXED as hunger|energy|cleanliness|affection; immutable pet_records
  (pet_id, founder_id, species_id, temperament, created_at) + mutable care state keyed by pet_id;
  fixture starter for tests; production starter creation activates only under a pinned pet catalog,
  atomically with New Founder (activation-boundary law applies) - no invented species/temperament.
- **C3 - accepted:** care/decay STRUCTURE ruled (fixed-grid decay on the attended clock - the
  provision-grid partition-invariance; per-action stat deltas; the >90%->0.5x diminishing curve as
  integer-ppm; closed cooldown/eligibility grammar); exact rate/floor/deltas are BALANCE DATA.
- **C4 - accepted, the Founder attended cursor:** the Founder stream carries its OWN
  founder_attended_ms cursor, server-stamped per Founder command (mirroring the Company attended
  derivation); pet decay uses that cursor's delta, so Founder replay is independent of run replay
  (C1's requirement). Offline spans contribute zero.
- **C5 - accepted:** Trust/Mood/combat-output become closed stored fields - trust_ppm (persistent,
  the combat C5 input), mood (session, reset rule declared), integer-ppm gain/decay; the
  (trust_ppm, soul) output cross-verified against the combat C5 golden vectors (soul READ from
  Meters via the carry seam, never written here per Meters C3).
- **C6 - accepted:** the FSM gets a closed transition contract - states, events, queue bounds,
  timing, PRNG label + draw order (save-seeded stream, replay-safe), persisted-vs-transient split;
  behavior stays display/flavor, never authoritative economy.
- **C7 - accepted:** no-death visibility is a declared projection - guild-visible pet STATUS is a
  read model (not raw Founder-save exposure), care-option greying returns a typed eligibility
  reason (reason-key pattern); only the public status projection is guild-visible.
- **C8 - accepted:** pet data is a Founder-scoped save version (its own migration, NOT the Company
  chain); named tables/JSON, schema version, Founder-replay corpus, activation-boundary law.

The through-line: **C1's Founder mutation boundary is the real deliverable** - a reusable
architectural addition the whole Founder layer has needed; the rest is structure-ruled/
numbers-deferred as usual.

## Implementation blockers C9-C12 (Codex, 2026-08-05)

The persistence envelope can be implemented without pet mechanics, and that slice has landed.
Applying the ruled text to the existing Exit/save topology exposes four remaining structural gaps.
These are not balance literals: choosing any of them in code would change replay, attendance, save
activation, or the visible state machine.

### C9 — Founder replay needs segment/genesis semantics across Exit

`founder_log` records Founder commands, but Exit also mutates the Founder stream outside
`ApplyFounderLogged`; retained save revisions are pruned. A single career-long replay therefore
has neither a durable genesis nor a complete command history. Logging only care commands would
false-diverge after the first Exit, while replaying all Exits would contradict the existing
cross-run-Founder-verification non-goal.

**Proposed contract:** replay immutable Founder *segments*. A forward migration adds an immutable
segment row containing exact pre-command Founder state bytes/version/hash and binds each
`founder_log` row to a segment. The Store lazily opens a new segment whenever the current Founder
revision is not the last applied revision of the open segment (Exit and any other owned
multi-stream mutation therefore form a boundary). Within a segment, applied revisions are
contiguous; rejected rows retain the current revision. Verification replays only the segment,
byte-compares receipts/events/order, and compares its terminal state to the next segment genesis
or current head. It never claims to verify the external Exit transition. Migration 00054 is
already applied locally, so this lands forward-only.

### C10 — A server timestamp is not an attended-time authority

C4 says the Founder cursor is server-stamped per command and offline spans contribute zero. Two
timestamps alone cannot satisfy both: elapsed wall time counts offline gaps, while counting only
the command instant advances by zero. No persisted Founder attendance source exists today.

**Proposed contract:** gameserver authentication/presence owns one persisted, monotonic Founder
attendance accumulator. Overlapping authenticated connections count elapsed time once; periods
with no active connection add zero. Presence updates lock only that accumulator. A Founder command
samples it after locking the Founder stream, records `{attended_before_ms, attended_after_ms}` in
resolved inputs, and advances the Founder save cursor exactly to the sampled total. Exit samples
the same authority rather than deriving a second clock. The lock graph has no reverse edge from
presence into save streams, and fixtures prove reconnect/offline, overlapping sessions,
partition-invariance, and an Exit between two care commands.

### C11 — Founder pet persistence and activation still have no exact wire

C2/C8 name `pet_records`, mutable care state, an “own” Founder save version, and New-Founder
activation, but do not define the table keys, save JSON, version number, or how the current shared
`VersionForState`/Exit transition permits a Founder-only version while Company remains on its
existing version. There is no safe loader or migration target yet.

**Proposed contract:** append immutable relational identity
`pet_records(pet_id,founder_id,species_id,temperament,created_at,catalog_hash)` and one closed
Founder-save pet map keyed by `pet_id`, with complete four-stat/remainder/cooldown/trust/FSM key
sets. Introduce the next Founder wire version without advancing Company state; refactor version
selection/Exit validation to accept that declared Founder/Company version pair. Starter creation
inserts identity plus initial Founder state in the New-Founder transaction only when the pinned
epoch contains the pet artifact. Pre-artifact founders synthesize nothing; later acquisition is a
successor contract. The exact JSON and migration corpus land before any production species row.

### C12 — Mood/FSM/status unions remain names, not closed contracts

C5-C7 say “closed” but enumerate no mood values, behavior states/events, queue hardcap, status
bands, or eligibility details. A Go/TS port cannot be byte-identical from prose.

**Proposed contract:** mechanically close Phase-A to four stat bands
`floor|low|normal|high`, derived mood `withdrawn|restless|neutral|engaged`, behavior states
`idle|care_response|active|resting`, and transition events `grid_tick|care_applied|care_rejected`.
The persisted queue hardcap is 8; candidates sort by mechanical behavior ID; PRNG label is
`pet.behavior.v1` with draws ordered candidate-index then duration. Persist current state,
entered-at attended cursor, bounded queued behavior IDs/due cursors, and PRNG cursor; mood remains
derived. Care rejection details are `cooldown|ineligible|saturated|unknown_pet|unknown_action`.
Balance rows own thresholds, durations, deltas, and candidate weights.

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
- 2026-08-03: created (draft) — the cattery port.
- 2026-08-04: C1-C8 ruled — introduces ApplyFounderLogged (the reusable Founder mutation boundary; I repeated the Meters-C3 Company-scope error), fixed stat IDs, Founder attended cursor, closed trust/mood/FSM contracts, no-death projection. Accepted.
