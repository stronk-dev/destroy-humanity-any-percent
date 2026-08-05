# RFC: Pet Care Foundation

- **Status:** accepted (C1-C21 ruled; introduces the Founder mutation boundary; implementing)
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

## Owner rulings on C9-C12 (2026-08-05)

- **C9 - accepted, reuse the run-genesis pattern for the Founder stream:** the Founder stream gets
  its OWN genesis (Founder-scope run-genesis analog) written at ApplyFounderLogged activation, and
  the Exit Founder-mutation is ALSO logged into founder_log (Exit becomes a founder_log entry, not
  only a Company one) - so a career-long Founder replay has a durable genesis + complete command
  history despite Company revision pruning. Same solution as run-genesis, applied to the Founder
  stream.
- **C10 - BLOCKED ON the new Founder Attendance Foundation** (drafted 2026-08-05). C4's
  server-timestamp cursor was underdesigned - a timestamp can't be both wall-elapsed and
  zero-on-offline. The shared clock (founder_attended_ms = summed run-attended, flushed at Exit +
  a frozen mid-run partial) is a separate primitive because Minigame C26 needs it too. Pet decay
  reads that clock; it does not define its own.
- **C11 - accepted:** exact wire - `pet_records(pet_id PK, founder_id, species_id, temperament,
  created_at)` + `pet_care_state` keyed by pet_id (the four fixed stat IDs + trust_ppm + mood +
  bond graph); a Founder-scoped save version distinct from the Company chain (VersionForState
  gains a Founder-version axis - the Company version and Founder version advance independently, the
  activation-boundary law per axis). Named tables/JSON/version in the closure batch.
- **C12 - accepted, closed unions enumerated (structure; content names deferred):** mood is a
  closed value set, behavior a closed state/event set with a queue hardcap, status a closed band
  set, eligibility a closed reason set - the ENUMS are ruled (Go/TS byte-identical from them);
  the specific member VALUES and thresholds are balance/content data. No prose-only union survives.

### C12a — The ruling still defers the enum members it says are closed

C12's first sentence says the Go/TypeScript unions are enumerated, but its second sentence defers
the member values. Those values are the wire grammar; they cannot be balance data and a port cannot
derive byte-identical enums from categories alone. The concrete C12 proposal immediately above
already supplies the missing Phase-A members.

**Proposed correction:** adopt the proposal's literals as normative mechanics:
`floor|low|normal|high`; `withdrawn|restless|neutral|engaged`;
`idle|care_response|active|resting`; `grid_tick|care_applied|care_rejected`; queue hardcap `8`;
PRNG label `pet.behavior.v1`; and care rejection details
`cooldown|ineligible|saturated|unknown_pet|unknown_action`. Thresholds, durations, weights, and
stat deltas remain balance/content data. Pet Care remains blocked on this textual correction and
Founder Attendance A1-A5, not on any invented mechanic.

## Owner ruling on C12a (2026-08-05) — adopt the enum literals (my C12 contradiction fixed)

C12a is right: I said the unions were 'enumerated' then deferred their MEMBER VALUES - but enum
members are WIRE GRAMMAR (a byte-identical Go/TS port needs them), NOT balance data. The
structure/numbers line runs BETWEEN the enum (structure) and the thresholds attached to each
member (balance) - I drew it one level too high. **Ruling: the Phase-A enum members are
normative:** status band `floor|low|normal|high`; mood `withdrawn|restless|neutral|engaged`;
behavior state `idle|care_response|active|resting`; behavior event
`grid_tick|care_applied|care_rejected`; behavior queue hardcap `8`; PRNG label `pet.behavior.v1`;
care-rejection detail set as the proposal lists. The per-member THRESHOLDS (what stat value is
'low', decay rate to a band) remain balance data. C12's 'members deferred' clause is withdrawn.

## Owner rulings on C13-C14 (2026-08-05)

- **C13 - accepted:** the fixture catalog rows are closed-key - decay rows, action rows (stat
  target + delta), trust/mood threshold rows, behavior-candidate rows - each exact-key with
  simultaneous-update grouping and uniqueness declared; the fixture loader validates the schema,
  the THRESHOLD/DELTA numbers are balance data. No decorative field survives.
- **C14 - the state-shape contradiction between my C1 and C11, resolved toward replay-owned:** I
  ruled pet state both as a separate mutable `pet_care_state` table (C11) AND as ApplyFounderLogged/
  Founder-replay-owned (C1) - a second mutable table beside Founder state would leave replay
  covering only half the transition. Ruling: **pet care state lives IN the Founder save state (the
  Founder stream's state jsonb), mutated ONLY through ApplyFounderLogged**, so Founder replay
  covers it entirely - NOT a separate mutable table. `pet_records` (immutable identity) stays a
  table; the MUTABLE care state (stats + trust + remainder/cooldown/queue/bond, named JSON keys) is
  Founder-state fields. **Mood is DERIVED (C12), never stored** - the C11 'mood stored' phrasing is
  withdrawn. This is the same discipline as every replay-owned surface: no second mutable authority
  beside the state the transition boundary owns.

## Implementation blockers C15-C16 (Codex, 2026-08-05)

C13-C14 settle catalog/state ownership, but applying them to the strict loader and the existing
per-stream revision version exposes two remaining literal contracts. Neither is a balance value.

### C15 — Two C13 row families still have no exact keys

C13 names an "exact threshold row" and an "exact candidate row" but does not enumerate either
object. In particular it does not state the mood threshold field/order or the behavior row's
from-state/event/duration field names. A loader would still invent which bytes are valid.

**Proposed contract:** mood rows are exact
`{mood,min_average_stat_ppm}` in ascending threshold order, contain every C12a mood exactly once,
start at zero, and have strictly increasing thresholds. Behavior rows are exact
`{temperament,from_state,event,behavior_id,weight,duration_grids_min,duration_grids_max}`; the
key `(temperament,from_state,event,behavior_id)` is unique, rows sort by that byte tuple, weight is
positive, and `1 <= duration_grids_min <= duration_grids_max`. Temperaments are exactly the six
combat values. These are schema/order rules only; all threshold, weight, and duration literals
remain fixture/balance data.

### C16 — Pet activation needs a pinned artifact and a Founder-only version transition

The database already versions each stream revision independently, but Exit currently rejects any
Founder/Company version mismatch and the replay artifact bundle has no pet artifact arm. C14's
"next Founder wire version" therefore cannot activate: writing `pets` would either advance the
Company wire too or make the next Exit fail, while accepting pet IDs/actions without pinned pet
bytes would make replay deployment-dependent.

**Proposed contract:** version 17 is Founder-only and adds the exact C14 `pets` map; Company saves
remain at their pinned Company version. Exit preserves the current Founder version independently
of the terminal/new Company versions, except that the already-shipped paired Meters/Achievements
v16 activation remains atomic when crossing from pre-v16. The replay artifact bundle gains an
optional `pets` artifact whose presence is biconditional with Founder v17. That artifact owns the
C13 catalog bytes and is included in constants identity before any v17 Founder is created. Go and
TypeScript reject Company v17, Founder v17 without the artifact, and pet artifacts with a pre-v17
Founder; mixed Founder-v17/Company-v14-or-v16 Exit and Founder replay are required fixtures.

## Owner rulings on C15-C16 (2026-08-05)

- **C15 - accepted:** the two row families are exact-key. Mood threshold rows
  `{mood_member, floor_ppm}` (the closed mood enum from C12, ascending floors, full set required);
  behavior candidate rows `{from_state, event, to_state, duration_grid_ticks}` over the closed
  behavior state/event enums (C12), with the queue-hardcap-8 bound. The threshold/duration NUMBERS
  are balance data; the field names + enum members are wire grammar. No invented bytes.
- **C16 + Minigame C35 - the SAME seam, ruled together: the Founder save version is an INDEPENDENT
  AXIS from the Company version** (this makes concrete the per-axis activation-boundary law already
  ruled at Pet C11). Exit's current 'Founder version == Company version' equality check is RELAXED
  to 'each version >= its pinned-epoch activation floor, validated independently' - the two axes
  advance separately. Founder-scope mechanics (pet state v-next, minigame_ratings) activate on the
  FOUNDER axis under their own PINNED artifacts (a `pets` epoch artifact; ratings ride the minigame
  artifact), New-Founder-forward or at the first Founder command under a pinned catalog - NEVER
  forcing a Company version bump. Replay reads the pinned pet/rating bytes from the run's pinned
  hash, never deploy-current (activation-boundary law). The Meters(Company)/Achievements(Founder)
  atomic co-activation stays a valid SPECIAL case (both artifacts in one epoch); it is not the
  general rule - independent axes are. Implementation: extend the Exit version-tuple validator to
  per-axis floors + add the `pets` artifact arm to the replay bundle (the missing arm C16 names).

## Implementation blocker C17 (Codex, 2026-08-05)

C16 correctly separates the Founder and Company version axes, and the Exit validator now enforces
independent pinned floors. It does not order two independently optional Founder mechanics on the
one monotonic Founder integer axis. If pets take v17 first, a later minigame-only v18 schema must
either require pet fields or invent a skip migration; reversing them creates the symmetric problem.
The current save format cannot represent arbitrary `{pets,minigames}` feature subsets with one
monotonic version number.

The artifact bytes are also incomplete: C13 names decay/action/Trust policies but supplies no
literal outer object or exact keys for those three row families, while C15 closes only mood and
behavior rows. Treating the C15 fixture as the complete `pets` epoch artifact would falsely pin a
partial catalog and leave state cooldown/action IDs deploy-current.

**Proposed contract:** fix one activation order on the Founder axis. The queue order suggests
`minigames=v17`, then `pets=v18`; v18 requires the still-pinned v17 minigame artifact and adds the
complete pet artifact/state. Company remains v14/v16 and rejects v17/v18. Alternatively, replace
the scalar Founder version with an explicitly versioned feature-vector envelope in a successor
RFC. In either shape, enumerate the complete `pets` artifact's exact top-level, decay, action, and
Trust keys before accepting it into constants identity. This is structural ordering/wire grammar,
not a request for balance literals.

## Owner ruling on C17 (2026-08-05)

**Accepted with the scalar-chain decision made in Minigame C36 (read there for the axis reasoning):
`minigames` = Founder v17, `pets` = Founder v18; v18 requires the v17 `minigames` artifact still
pinned. The feature-vector envelope is deferred to a named successor RFC, to be reached only if a
future epoch needs a higher Founder mechanic without a lower one — YAGNI until then.**

**The complete `pets` epoch artifact is the FULL union below — pinning the C15 fixture alone would
falsely pin a partial catalog (the exact hazard this blocker names), so C15 is a subset, not the
artifact.** The artifact's exact top-level is C13's `{schema_version, stat_policy, actions,
trust_policy, mood_policy, behavior_policy}`, with every row family closed:

- `stat_policy`: absolute `grid_ms`, one exact `{stat_id, initial_ppm, floor_ppm,
  decay_ppm_per_grid}` row per fixed stat, plus the diminishing threshold/factor ppm (C13).
- `actions`: unique exact `{action_id, stat_id, delta_ppm, cooldown_attended_ms, min_eligible_ppm}`
  rows (C13). Cooldown is on the attended watermark, never wall time.
- `trust_policy`: exact keys `{initial_ppm, neutral_ppm, floor_ppm, cap_ppm,
  gain_ppm_per_effective_action, decay_ppm_per_grid}` (C13).
- `mood_policy`: C15's exact mood rows `{mood_member, floor_ppm}`, ascending, full closed set —
  **this SUPERSEDES C13's mood sketch.**
- `behavior_policy`: C15's exact DETERMINISTIC transition rows `{from_state, event, to_state,
  duration_grid_ticks}` over the closed C12 state/event enums — **this SUPERSEDES C13's weighted
  temperament/candidate sketch.** The behavior FSM is deterministic (from_state + event → to_state),
  not weighted-random; that is the replay-determinism-consistent choice and it is final. If
  temperament ever needs to qualify a transition, it enters as a `from_state` qualifier or a future
  ruling — it is NOT invented now, and its absence is correct-by-omission, not a gap.

All numbers in these rows are fixture/balance data; the KEYS, enum members, and top-level are wire
grammar and are hereby pinned. Unknown/duplicate/missing rows fail load (C13). No field is
deploy-current. Pet state itself remains the Founder-save-owned jsonb (C14), added at v18; mood is
DERIVED from `mood_policy`, never stored (C14/C15).

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

## Implementation blockers C13-C14 (Codex, 2026-08-05)

C12a's literal cross-runtime vocabulary is implemented. The next code boundary exposes two exact
wire gaps rather than missing balance values.

### C13 — C3 does not enumerate the fixture catalog rows it calls closed

The RFC names decay, actions, trust, mood thresholds, and behavior candidates, but does not give
their exact keys, simultaneous-update grouping, or uniqueness rules. A fixture-only loader would
still have to invent its schema, so tests could not distinguish content from decorative fields.

**Proposed contract:** a v1 pet-mechanics catalog has exact top-level keys
`{schema_version,stat_policy,actions,trust_policy,mood_policy,behavior_policy}`. `stat_policy`
contains the absolute `grid_ms`, one exact `{stat_id,initial_ppm,floor_ppm,
decay_ppm_per_grid}` row for each fixed stat, and the diminishing threshold/factor ppm. Actions are
unique exact `{action_id,stat_id,delta_ppm,cooldown_attended_ms,min_eligible_ppm}` rows. Trust owns
`{initial_ppm,neutral_ppm,floor_ppm,cap_ppm,gain_ppm_per_effective_action,decay_ppm_per_grid}`.
Mood owns one exact threshold row per C12a mood. Behavior owns exact candidate rows keyed by
temperament/state/event with mechanical behavior ID, positive weight, and duration-grid bounds.
All numbers are fixture/balance data; unknown/duplicate/missing rows fail load.

### C14 — C11 conflicts on the authoritative Pet state shape

C11 names a mutable `pet_care_state` table while C1 requires `ApplyFounderLogged` and Founder replay
to own the mutation. It also says mood is stored, while C12 makes mood derived, and never names
remainder/cooldown/queue/bond JSON keys. Writing a second mutable table beside Founder state would
make replay cover only half the transition.

**Proposed contract:** `pet_records` is immutable relational identity only. Mutable state lives in
the Founder snapshot under an exact `pets` map keyed by pet ID: `{stats_ppm,
stat_decay_remainders_ppm,cooldown_until_attended_ms,trust_ppm,trust_decay_remainder_ppm,
behavior_state,behavior_entered_at_attended_ms,behavior_queue,behavior_prng_cursor}`.
`stats_ppm` and decay remainders have the complete four-stat key set; cooldowns use declared action
IDs; each queue entry is exact `{behavior_id,due_attended_ms}` and the queue obeys hardcap 8. Mood
is derived and not persisted. Writable bonds defer to their cross-Founder successor; no bond graph
ships in this state version. The next Founder wire version activates this map only with a pinned
pet artifact.

## Implementation blockers C18-C21 (Codex, 2026-08-05)

The complete v18 artifact/state grammar and F1 exact-key remediation are independently approved.
The requested care-transition consumer still reaches four mechanical gaps. Implementing through
them would make decay evaluation-frequency-dependent or invent visible pet behavior.

### C18 — Pet state has no attended evaluation watermark and its remainders have no equation

`CareState` stores decay remainders and a behavior-entered cursor, but no last care-evaluation
cursor. Two commands in one run therefore cannot know which part of the frozen Founder-attendance
sample was already applied: recomputing from completed `age_ms` double-decays the first interval,
while treating `behavior_entered_at_attended_ms` as the watermark changes its ruled meaning. C13
also gives integer `decay_ppm_per_grid` values but never defines how the `*_remainder_ppm` fields
participate.

**Proposed contract:** amend pre-mint Founder v18 state with exact
`evaluated_through_attended_ms` per pet. Each resolved care input records the complete A1-A5
Founder-attendance sample plus that stored before-cursor; ApplyFounderLogged requires
`before == state.evaluated_through` and advances exactly to the sample total. Define stat/trust
integration as
`numerator=elapsed_ms*decay_ppm_per_grid+remainder`,
`decay=floor(numerator/grid_ms)`, `remainder=numerator mod grid_ms`, with checked wide-integer
intermediates. Stats saturate at their catalog floors; Trust above neutral decays toward neutral
and never crosses it. Remainder validation becomes `< grid_ms` under the pinned pet catalog. This
gives `advance(a+b)==advance(advance(a),b)` for every split and makes retry/stale-sample behavior
fail closed.

### C19 — Care eligibility and diminishing returns do not define one result

C3/C13 name `min_eligible_ppm`, a diminishing threshold/factor, cooldown, saturation, and
"over-care wastes," but do not state the comparison direction, operation order, or whether a
partially effective action applies. Different reasonable ports produce different Trust gains and
receipts.

**Proposed contract:** an action is eligible when the target stat is at least
`min_eligible_ppm`; catalogs must keep every recovery action eligible at its stat floor (the
no-death law). Resolve decay first, then cooldown/eligibility. If the target is at/above the
diminishing threshold, compute
`effective=floor(delta_ppm*diminishing_factor_ppm/1_000_000)`, otherwise use the full delta;
`applied=min(effective,1_000_000-current)`. Zero applied rejects as `saturated`; positive partials
apply and the unapplied tail is visible waste. Only a positive applied amount starts cooldown and
earns one `gain_ppm_per_effective_action` Trust grant, capped by Trust policy.

### C20 — Mood/status input and deterministic FSM scheduling are unspecified

Mood thresholds do not name the scalar they threshold. The deterministic behavior rows removed
weighted randomness, but state still carries a PRNG cursor and queue of `behavior_id`s while rows
contain only destination states. No rule orders elapsed grid ticks, due queue entries, and the
care event, or bounds a long attended interval without iterating every tick.

**Proposed contract:** derive the care scalar as the minimum of the four current stat ppm values;
the greatest mood threshold not above it wins. Map the four ordered moods one-to-one onto public
status bands `floor|low|normal|high`; snapshots expose only that band and eligible action IDs.
Treat each behavior row's `to_state` as the queued mechanical behavior ID. On transition, drain
due entries in `(due_attended_ms,behavior_id)` order, then apply the current command event; a
matching row replaces any pending entry for the same destination and inserts its absolute due
cursor, sorted, under hardcap 8. `grid_tick` is generated only for crossed absolute grid
boundaries. The implementation must use deterministic cycle skipping over `(state,queue)` for
large intervals; iteration proportional to total attended grids is forbidden. Because C17 removed
weighted choice, `behavior_prng_cursor` is fixed at zero in v18 and should be removed in a later
wire version rather than fake-used.

### C21 — The live Founder command and replay/event envelopes are not exact

PC2 says `care_action {pet,action}` but does not enumerate the authoritative intent fields,
resolved attendance arm, receipt, events, or how the server locates the active Company used by the
attendance sample. Adding an ad-hoc Company ID would let the client select clock context.

**Proposed contract:** add Founder intent exact wire
`{intent_id,kind:"care_action",expected_revision,pet_id,action_id}`. The server resolves the sole
active Company sibling; the client supplies no Company/run coordinate. The Founder replay resolved
arm is exact `{kind:"care_action",attendance,pet_attended_before_ms}`; canonical payload owns the
pet/action IDs and the pinned artifact owns policy. Receipt is exact
`{intent_id,outcome,founder_revision,pet_id,action_id,stat_id,before_ppm,applied_ppm,after_ppm,
trust_before_ppm,trust_after_ppm,mood,status_band,next_eligible_attended_ms}`. Register
`pet_care_applied.v1` and `pet_status_changed.v1`; the latter emits only when the derived public
band changes. Unknown pet/action uses `unknown_id`; cooldown/ineligible/saturated use the existing
`not_eligible` category plus C12a's detail member. Go/TypeScript shared vectors compare state,
receipt, and ordered event bytes.

## Changelog

- 2026-08-04: Codex acceptance review found C1–C8. Founder care cannot mutate inside the
  Company-only `ApplyLogged` boundary; pet identity, care/time arithmetic, trust/combat output,
  behavior FSM, projection privacy, and Founder replay/activation require executable contracts.
- 2026-08-03: created (draft) — the cattery port.
- 2026-08-04: C1-C8 ruled — introduces ApplyFounderLogged (the reusable Founder mutation boundary; I repeated the Meters-C3 Company-scope error), fixed stat IDs, Founder attended cursor, closed trust/mood/FSM contracts, no-death projection. Accepted.

## Owner rulings on C18-C21 (2026-08-05) — the care-transition composer

All four proposed contracts are ACCEPTED. They honor the determinism spine, the no-death law, the
closed-form/no-per-tick law, and the server-authoritative boundary; the notes below ratify the
load-bearing points and add cross-cutting consistency.

- **C18 — accepted.** Add exact `evaluated_through_attended_ms` per pet; each resolved care input
  records the A1-A5 attendance sample + the stored before-cursor; ApplyFounderLogged requires
  `before == state.evaluated_through` and advances to the sample total. The integration is the
  project's **provision-grid partition-invariant primitive**: `numerator = elapsed_ms *
  decay_ppm_per_grid + remainder`, `decay = floor(numerator / grid_ms)`, `remainder = numerator mod
  grid_ms`, checked wide intermediates, `remainder < grid_ms` validated under the pinned catalog.
  This is the SAME primitive minigame C40 uses for offline-quality decay and the faucet window uses
  for conversion carry — **the three must share one tested helper (or byte-identical vectors) so
  `advance(a+b) == advance(advance(a),b)` holds identically everywhere.** The watermark is the
  no-double-decay guard and makes stale/retry samples fail closed. Stats saturate at catalog floors;
  Trust decays toward neutral and never crosses it (monotone bound).
- **C19 — accepted, and the no-death carve-out is load-bearing.** Eligible when target stat >=
  `min_eligible_ppm`; **catalogs MUST keep every recovery action eligible at its stat floor** — a
  floored stat that couldn't be recovered would violate the no-death law. Resolve decay FIRST, then
  cooldown/eligibility (evaluate current state before gating). At/above the diminishing threshold:
  `effective = floor(delta_ppm * diminishing_factor_ppm / 1_000_000)`, else full delta;
  `applied = min(effective, 1_000_000 - current)`. Zero applied rejects `saturated`; positive
  partials apply with the unapplied tail as visible waste; only a positive applied amount starts
  cooldown and earns one `gain_ppm_per_effective_action` Trust grant capped by policy. Numbers are
  balance data; the order-of-operations and comparison direction are hereby ruled.
- **C20 — accepted, and cycle-skipping is the no-per-tick law.** Care scalar = the MIN of the four
  current stat ppm (worst-neglected need dominates mood — the thematically correct aggregator);
  the greatest mood threshold not above it wins; the four ordered moods map one-to-one onto public
  bands `floor|low|normal|high`; snapshots expose ONLY the band + eligible action IDs (projection
  privacy). Behavior `to_state` is the queued mechanical behavior ID; on transition drain due
  entries in `(due_attended_ms, behavior_id)` order, then apply the command event; a matching row
  replaces any pending entry for the same destination, inserting its absolute due cursor sorted
  under hardcap 8. `grid_tick` emits only for crossed absolute grid boundaries. **The implementation
  MUST use deterministic cycle-skipping over `(state,queue)` for large intervals — iteration
  proportional to attended grids is forbidden (the closed-form/never-a-tick-loop binding law).**
  `behavior_prng_cursor` is fixed at zero in v18 (C17 removed weighted choice) and is removed in a
  NAMED later Founder wire version — never fake-used; a vestigial mutable field is a replay-ownership
  hazard, so its removal is a tracked successor, not left to rot.
- **C21 — accepted, and the clock-context rule is security-critical.** Founder intent wire
  `{intent_id, kind:"care_action", expected_revision, pet_id, action_id}`. **The server resolves the
  sole active Company sibling; the client supplies NO Company/run coordinate** — an ad-hoc Company ID
  would let the client select its clock context, which is forbidden (server-authoritative clock).
  Resolved arm `{kind:"care_action", attendance, pet_attended_before_ms}`; canonical payload owns the
  pet/action IDs, the pinned artifact owns policy. Receipt is the exact enumerated set; register
  `pet_care_applied.v1` and `pet_status_changed.v1` (the latter emits only when the derived public
  band changes — privacy projection). Unknown pet/action → `unknown_id`; cooldown/ineligible/
  saturated → `not_eligible` + the C12a detail member. Go/TS shared vectors compare state, receipt,
  and ordered event bytes. Structure ruled; numbers deferred.

**These four complete the C1-ruled ApplyFounderLogged care-transition. Activation stays
New-Founder-forward under the pinned pet artifact; nothing here self-authorizes archival, and
pet-care AC3 (combat C5 cross-verification) remains OPEN until combat's obedience table consumes the
seam.**
