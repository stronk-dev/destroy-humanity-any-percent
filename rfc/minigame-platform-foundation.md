# RFC: Minigame Platform Foundation

- **Status:** accepted — C1–C18 ruled; normative body reconciled; implementing
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/03` (the minigame suite, clocks, economy hooks, AI-fallback law), `design/03 §12b` (the tier→minigame scaling seam + Fairness Law), `design/05 §4` (PvP table, punch-down multiplier)
- **Research:** `neopets-systems.md §4` (the portal faucet model: N-sends × per-game cap × conversion ratio, monthly re-tune), `kol-puzzle-pirates.md §B` (minigame-as-labor, offline-output quality grades), `creature-battler.md` / `lane-pusher-design.md` (the combat instances)
- **Depends on:** Production + Run Genesis + Transport + Gameserver Composition (implemented); Combat Shared Kernel (implemented — the duel engine is tenant #1). **NOT Clout** (removed per C1).
- **Owner ruling honored:** breadth-first — the PLATFORM every minigame is content on: session lifecycle, the scaling seam as code, the faucet governor, reward hooks, the registry. No individual minigame.
- **Planning:** `planning/minigame-platform-foundation/` (once implementing)

## Summary

Every future minigame — board games, the two combat engines, the Market, the arcade toys, pets'
battles — needs the same platform: a session/match lifecycle, the §12b scaling-seam contract as
executable code, the Neopets faucet governor (so a minigame can never become an economy exploit),
reward payout into the real economy, and a registry. Combat's C5 already prototyped the seam
("the seam is these two integers"); this generalizes it to a platform contract.

## Specification

### MP1 — Session/match lifecycle (closed state machine)

A minigame session is a **Postgres row** (C13's `minigame_sessions` — the verification-queue table
discipline: claim-token + RowsAffected, server-advanced only) — `{session_id, minigame_id,
founder, company_stream_id, run_seq, scaling_inputs (frozen), seed, mode: solo|async_snapshot,
state, result|null}`. **`live_pvp` is a reserved enum arm, NOT shipped (C8)** — Phase-A modes are
`solo` and `async_snapshot` only, the two combat proves replayable. Lifecycle: create (server
resolves scaling inputs, freezes them, seeds from the save stream — the combat C5 pattern) → play
(the tenant engine advances the session via server-validated play commands) → **resolve
(server-authoritative, replay-verified like combat — solo/snapshot both)** → payout (MP4, a
server-authored transition). No minigame reads live company state mid-session; the seam is always
the frozen integer inputs (replay-safe by the RA rule, exactly as combat proved).

### MP2 — The scaling-seam contract (§12b as code)

`scaling_inputs`: closed named integer functions over (company, founder) state, per the §12b
axes (`reward`, `resource_pool`, `challenge`, `breadth`, `era`, `offline_quality`), declared per
minigame in its registry entry, computed server-side at session create, frozen into the snapshot.
**The Fairness Law is loader-enforced:** a scaling input feeding a RANKED PvP power stat is a load
error; tier scaling reaches ranked play only as `breadth` (options inside a fixed ceiling — the
2026-07-28 hardcap decision generalized from pets to all ranked minigames). `offline_quality`
(the Puzzle Pirates grade) is the active/idle bridge: recent session performance charges the
quality tier of the associated automated output, decaying over ~31 days, never zeroed.

### MP3 — The faucet governor (the Neopets shape, mandatory)

Every minigame declares `payout: {sends_per_day, per_send_cap, conversion_ppm}` — N score-sends
per day, each capped, converted at a published ratio. This is the STRUCTURAL guarantee a minigame
cannot become an infinite faucet (the Cloud-Clicker economy laws' answer to "minigame prints
currency"): the daily cap is the governor, published per epoch, re-tunable on the epoch cadence
(the Neopets monthly-25th precedent as our BALANCE-CHANGE cadence). Rewards above the cap are
forfeit (accrual-only, reason-keyed). Bot/AI-fallback matches pay at a declared reduced,
non-ranked rate (anti-farm — the researched universal).

### MP4 — Reward payout & the AI-fallback law

Resolve → payout (rulings C1/C6/C17). **NO Clout** (Feed/Social owns Clout's mint). **The payout
is NOT a client intent** — a minigame result is a SERVER-CERTIFIED outcome: the server resolves
the session (replay-verified), then commits, in ONE Postgres transaction, the session row
(status=resolved, claim-token-guarded) AND a **server-authored Company-stream transition** —
`resolve_minigame_session` (server-only, never client-submitted) — that credits a tier-scaled
economy resource (governed by MP3) plus typed rating/season facts (real Elo, no rubber-band).
Evented and replay-logged like every transition; the certified result is the recorded fact
(replay never re-runs the minigame at verification). The client submits PLAY commands to the
session and NEVER a score. **Every minigame declares its AI fallback** (closed union, C18:
`solo|bot|npc_partner`) — a fallback-less minigame is a load error; the whole game is playable at
zero other players online.

### MP5 — The registry

`balance/minigames/*.json`: `{minigame_id, engine_ref, clock: turn|realtime|wallclock|async,
scaling_inputs[], payout, fallback, unlock_condition_ref}` (era_skins REMOVED per C9 — one token
theme per era, never per-feature). `unlock_condition_ref` uses the closed shell-fact grammar
(UI C5). The combat DUEL engine registers as tenant #1 (their existing catalogs become platform-conformant); the board
games, Market, and arcade toys are later content on this registry. `unlock_condition_ref` gates
by catalog fact (staggered across the tier arc — the CC hour-5-40-sag fix); the platform never
invents unlock logic.

## Owner rulings on C1–C11 (2026-08-04)

- **C1 — accepted, and this one is on me:** MP4 reintroduced the exact Clout-as-payout faucet I
  ruled OUT of Achievements an hour earlier. **Clout removed from payout and dependencies.**
  Phase-A payout arms are only owners that exist at implementation: an economy-resource credit
  (catalog resource ID) and typed rating/season facts (non-resource state). Feed/Social may add a
  social-activity Clout arm by successor RFC through the closed reward union; the generic platform
  never mints a free-form resource/effect.
- **C2 — accepted, the session is DB-authoritative:** sessions are Postgres rows
  (`minigame_sessions`), server-advanced only; concurrency and double-pay prevented by the SAME
  claim-token + RowsAffected discipline the verification queue uses; resolve/payout is one
  transaction keyed by session id (idempotent). No in-memory actor is authoritative for
  async/snapshot modes.
- **C3 — accepted:** `engine_ref` binds to a closed tenant interface — a typed
  `(command, snapshot, result)` descriptor + engine-version identity + deterministic error
  taxonomy; the combat duel adapter and a fixture tenant both implement it and are conformance-
  tested against the same boundary; an engine reading ambient state or returning an unvalidated
  payout is a load/contract error.
- **C4 — accepted:** scaling axes become typed rows (source field, integer range, composition
  order, ranked-power binding) so the Fairness Law is loader-PROVABLE, not asserted.
- **C5 — accepted:** faucet uses **attended-time day boundaries on the absolute grid** (not
  wall-clock — the provision-grid precedent kills the timezone/deploy ambiguity), integer-ppm
  conversion with declared rounding + overflow saturation, cap consumption persisted per session
  (a resolve retry can't double-spend — same idempotency), bot reduction a catalog literal.
- **C6 — accepted, the sharpest correction:** payout is NOT a client intent. A minigame result is
  a SERVER-CERTIFIED outcome: the server resolves the session (replay-verified for solo/snapshot,
  like combat), then the payout enters the economy through a **server-authored** transition inside
  the resolve transaction — evented and replay-logged, but never a client-submitted
  `claim_payout`. The client submits PLAY commands to the session; it never submits a score.
- **C7 — accepted:** fallback is a closed union with exact fields (solo: no peer; bot: catalog
  bot identity + reduced non-ranked rate literal + replay via the deterministic bot policy;
  npc_partner: the virtual-guild pattern). Bot results replay like any tenant result.
- **C8 — accepted, SCOPE RULING: `live_pvp` is DEFERRED out of this foundation.** Phase-A ships
  `solo` and `async_snapshot` ONLY (the modes combat already proves replayable). The `live_pvp`
  mode boundary is DECLARED (the enum arm reserved) but unimplemented; the PvP-service RFC owns
  disconnect/reconnect/authoritative-clock/finalization. AC1 drops to the two shipped modes. This
  keeps the platform honest — it ships what's replay-provable now.
- **C9 — accepted:** `unlock_condition_ref` reuses the closed shell-fact grammar (UI C5); **`era_skins`
  REMOVED** (UI Foundation: one token theme per era, never per-feature skins — my `era_skins` was
  a UI-law violation). AC6 corrected: the combat duel engine registers as tenant #1 via adapter;
  the lane engine follows when implemented (not claimed ready).
- **C10 — accepted, and the labor hazard ruled:** `offline_quality` is stored session state (a
  ppm grade + last-session timestamp), a deterministic decay transition on the attended grid; **it
  decays toward a NEUTRAL floor, never zero** (the ruling that keeps an optional minigame from
  becoming mandatory labor — a lapsed player's automated output degrades to baseline, never below).
- **C11 — accepted:** the registry + tenant schemas + fallback/faucet policy join epoch identity
  (a minigames artifact, mint like the others); session genesis + results join the verifier where
  they affect verification; the DB migration + cross-runtime corpus are named in the closure batch.

## Owner rulings on C12–C18 (2026-08-04)

- **C12 — accepted, body reconciled:** the normative text above is corrected to match the C1–C11
  rulings (live_pvp arm reserved-not-shipped, era_skins removed, duel engine only as tenant #1).
  My rulings appended without fixing the body — corrected now.
- **C13 — session persistence: reuse the verification-queue table discipline exactly.**
  `minigame_sessions(session_id PK, minigame_id, founder_id, company_stream_id, run_seq,
  scaling_inputs jsonb, seed, mode, status, claim_token, state jsonb, result jsonb|null,
  created_at, resolved_at)`; server-advanced only; concurrency + double-resolve prevented by
  claim-token + RowsAffected (the queue's proven pattern); lock order founder-then-session;
  expiry/retention as the queue's. No new invention — the pattern is shipped.
- **C14 — tenant descriptor:** a closed `(command_schema, snapshot_schema, result_schema,
  engine_version, error_taxonomy)` per engine_ref; the result union is `{outcome, score_facts,
  rating_delta|null}` (NO free-form payout — C1); the combat duel adapter and the fixture tenant
  both implement it, conformance-tested against one boundary. Exact byte envelopes are the tenant's
  (combat already has them); the PLATFORM contract is the descriptor shape.
- **C15 — scaling loader-provable:** each scaling row maps its destination to exactly one class
  `power|breadth|presentation`; **the Fairness Law is the loader rule `ranked ∧ power → reject`**;
  bounds/formulas per row are BALANCE DATA (harness-tuned, formula-artifact-exported) — the
  SCHEMA is ruled here, the numbers are not invented.
- **C16 — faucet clock:** attended-grid origin = `run_started_at` (the provision-grid precedent —
  one absolute origin, persisted cursor per session), floor-division ppm conversion, **saturation
  distinguished from the configured cap** (the cap forfeits with a reason key; numeric overflow
  saturates — two different events); bot-reduction literal + reset order are catalog data.
- **C17 — server-authored payout transaction:** the resolve transaction commits, in ONE Postgres
  tx: the session row (status=resolved, claim-token-guarded) AND a server-authored Company-stream
  transition (the payout — a credit + rating facts, evented + replay-logged), same multi-write
  discipline as Exit; the internal command is `resolve_minigame_session` (server-only, never a
  client intent — C6); lock order company-then-session. Replay reads the certified result from the
  log, never re-runs the minigame at verification (the result IS the recorded fact).
- **C18 — fallback/offline_quality:** fallback union exact-keys `{kind: solo|bot|npc_partner,
  bot_ref?, rate_reduction_ppm?}`; `offline_quality` = `{grade_ppm, last_session_at}` on session
  state, decaying on the attended grid toward a **neutral floor (catalog literal, never zero)**
  via a fixed-grid formula (the provision-grid partition-invariance applies); the automation
  destination is the associated tenant's offline output. Grade SOURCE (score→grade curve),
  decay rate, and floor value are BALANCE DATA — the state/transition shape is ruled, the numbers
  deferred.

The through-line of C13–C18: **structure ruled precisely by reusing shipped patterns (queue table,
provision grid, Exit multi-write); balance NUMBERS deferred to harness-tuned catalog data** — the
platform is executable without inventing a single balance value.

## Acceptance criteria

1. Session lifecycle: create→play→resolve→payout for a fixture tenant against the composed
   gameserver, in `solo` and `async_snapshot` modes ONLY (live_pvp deferred, C8); scaling inputs
   frozen at create, unread thereafter (replay-safe proof).
2. Fairness Law: loader rejects a scaling input wired to a ranked power stat; `breadth` scaling
   into ranked play passes; `offline_quality` charges/decays per the grade ladder.
3. Faucet governor: the daily cap forfeits excess (reason-keyed); a bot match pays the reduced
   rate; per-epoch re-tune follows BALANCE-CHANGE.
4. Payout-as-intent: a resolved result enters the economy through the intent surface, evented and
   replay-logged; no side-channel write exists (grep-proven).
5. Fallback law: loader rejects a fallback-less minigame; the fixture tenant's solo path completes
   at zero connected peers.
6. Registry: the combat DUEL engine registers as a conformant tenant via adapter (the lane engine follows when implemented — not claimed ready, C9).

## Post-ruling implementation blockers (Codex review, 2026-08-04)

C1–C11 settle the architecture and remove the Clout/UI/live-PvP contradictions. They do not yet
provide several exact schemas and arithmetic choices their own proposed contracts require. The
following are the narrow remaining contracts needed before code can have one correct shape.

### C12 — The normative body still contradicts three rulings

MP1 still declares `mode: solo|async_snapshot|live_pvp`; MP5 still requires
`unlock_condition_ref` and `era_skins`; AC6 still requires both combat engines. C8 reserves but
rejects live PvP, C9 replaces the unlock grammar/removes skins, and the lane engine is unimplemented.

**Proposed contract:** reconcile decision sites: Phase-A mode is exactly `solo|async_snapshot`
(the wire reserves no accepted `live_pvp` value until its owner RFC); registry unlock is exactly
the UI `always|fact_equals` union; remove `era_skins`; AC6 requires the fixture tenant and duel
adapter only, with lane registration moved to the lane RFC.

### C13 — Session persistence still has no executable row/command contract

C2 names Postgres, claim tokens, and one transaction but not the row keys/columns, transition
commands, revision/hash semantics, allowed concurrency, expiry, retention, or lock order.

**Proposed contract:** define `minigame_sessions` exactly with UUIDv7 session/command IDs,
account/founder ownership, minigame/mode, engine/constants identity, genesis/snapshot/result bytes,
status, revision, attended cursor, claim token/lease, paid revision, and created/updated/expiry
timestamps. Define exact create/play/abandon/resolve server commands and the closed rejection
taxonomy. Reuse canonical SHA256 idempotency over `(session_id,command_id,payload)`. Declare the
maximum active sessions per founder/mode, lease duration, expiry/retention literals, and lock order
relative to founder/company save streams.

### C14 — Tenant descriptors and transition results remain nouns

C3 accepts typed descriptors but supplies no exact descriptor rows, byte envelopes, result/event
union, engine registry interface, or error categories.

**Proposed contract:** enumerate the descriptor JSON and Go interface literally: schema versions,
maximum byte sizes, command/snapshot/result schema references, engine version, allowed modes and
event kinds; transition input/output envelopes with exact keys; and the deterministic rejection
union. State whether snapshot/result bytes are canonical JSON or opaque bytes plus content type.
The fixture tenant's literal schemas/vectors become the conformance oracle.

### C15 — Scaling is still not loader-provable

C4 says typed rows exist but neither lists their exact union nor maps destinations to
`power|breadth|presentation`. “Bounded exact integer” has no bounds/formulas.

**Proposed contract:** provide the exact source union and field names for the Phase-A arms accepted
in C4, per-arm integer range/formula, composition order, destination descriptor, duplicate rule,
and the literal fixture tenant destination registry. Define taint propagation: any tier-derived
source reaching a ranked `power` destination through any composition path rejects.

### C16 — The attended-day faucet clock and arithmetic are ambiguous

An “absolute grid” over attended time needs an origin and persisted cursor. C5 also says declared
rounding/overflow saturation without choosing the rounding or distinguishing a configured payout
cap from numeric overflow. Bot reduction and counter reset order remain open.

**Proposed contract:** identify one monotonically persisted Founder attended-time total and define
window index/origin exactly; if no such authority exists, add it with its save owner/migration.
Spell the checked integer formula and rounding (`floor` or carried remainder), numeric-overflow
behavior (reject/invariant—never silently saturate), order of bot multiplier/conversion/per-send/
window caps, and the exact counter/session idempotency row. Catalog values remain balance data, but
all required keys/ranges and reason-key flow are normative.

### C17 — Server-authored payout has no production transaction envelope

C6 correctly forbids client score claims but does not define the internal command, Company stream
event/receipt payload, replay-input row, or how a minigame DB transaction atomically commits a save
stream revision. “Inside resolve transaction” is not a lock/order/API contract.

**Proposed contract:** provide exact `apply_minigame_result` internal payload/resolved-input/
receipt/event schemas; declare session→founder→company lock order (or the existing canonical order
if different); identify the one repository method owning session resolution + run-log/save/events
+ payout-counter commit; and require fault injection after every write. Replay re-derives the
governed credit from immutable genesis/result/policy bytes and compares receipt/events bytewise.

### C18 — Fallback and offline-quality rows still lack literals/formulas

C7 calls the union exact without enumerating its keys. C10 names a ppm grade/timestamp and neutral
floor but no grade source, grid origin, decay formula/remainder, floor/cap, reset/scope, or
automation destination. These choices change state and replay.

**Proposed contract:** enumerate exact `solo|bot|npc_partner` rows, bot profile/version/policy,
ranked eligibility and payout fields, and the Phase-A fixture values. Either defer
`offline_quality` at loader level, or provide its complete catalog/state/transition contract and
one named neutral automation slot; prove no-play converges to neutral and never reduces canonical
offline production below 90%.

## Open questions

- Live-PvP match-actor lifecycle detail (matchmaking, bot backfill timing) — the PvP-service RFC
  owns it; this platform defines the `live_pvp` mode boundary.
- Whether the arcade (design/03 §9) toys need full sessions or a lightweight score-send-only path
  (recommend: lightweight — a Snake high score doesn't need match state).

## Acceptance blockers (Codex review, 2026-08-04)

The platform boundary is necessary, but the draft is not executable yet. In particular, its reward
owner predates the ruled Achievements correction, and its session/result nouns do not define a
server-authoritative persistence or replay contract. The following proposed closures keep the
platform generic without inventing any minigame.

### C1 — The payout contract reintroduces the forbidden Clout faucet

MP4 says minigames pay “Clout (achievement/rating).” Achievements C1 established that Feed/Social
owns Clout's single mint and no reward system may create a second source. Minigame Platform cannot
depend on an “Account/Clout foundation” that does not exist or mint Clout through a generic hook.

**Proposed contract:** remove Clout from MP4 and from this RFC's dependencies. Phase-A platform
payout arms are only owners that exist when implementation lands: an economy-resource credit with
a catalog-declared resource ID, and typed minigame rating/season facts that are non-resource state.
Feed/Social may add a social-activity Clout arm by successor RFC through the closed reward union;
the generic platform never gains a free-form resource/effect row.

### C2 — A session has no persisted authority or concurrency contract

MP1 lists fields and arrows but not where they live, who may advance them, whether multiple
sessions may coexist, how retries behave, or what prevents two resolve/payout workers from paying
one result twice. An in-memory match actor cannot be the authority for async snapshots or survive a
restart.

**Proposed contract:** define an append-only session command log plus one transactional session
head keyed by UUIDv7 `session_id`. The closed states are `created → active → resolved → paid` plus
terminal `abandoned`; transitions use `(session_id, expected_revision, command_id)` with the
existing canonical-payload/idempotency-hash discipline. Resolve records immutable result bytes;
payout claims that result in the same transaction as its authoritative credit/event writes and is
idempotent by session ID. Exact account/founder ownership, maximum concurrent sessions per mode,
abandon/expiry rules, retention, and lock order must be literal contracts, not handler choices.

### C3 — `engine_ref` does not define a tenant interface

There is no closed command/snapshot/result descriptor, engine-version identity, deterministic
error taxonomy, or proof that combat adapters and the fixture tenant implement the same boundary.
An arbitrary engine callback could read ambient state or return an unvalidated payout.

**Proposed contract:** register a typed tenant descriptor containing named command, snapshot, and
result schema versions plus an immutable engine version. The platform calls a pure deterministic
transition `(snapshot, canonical_command, frozen_inputs, seed/substream) →
(snapshot', result|null, events)`; the tenant cannot access DB, clock, network, or payout APIs.
The platform validates bytes before and after every call and owns the closed rejection taxonomy.
Tenant events are namespaced and registered like production events. Payout derives only from the
validated immutable result through MP3—never from tenant-supplied resource deltas.

### C4 — Scaling axes are labels, not functions

`reward`, `challenge`, and the other names have no row shapes, source fields, integer ranges,
composition order, or ranked-power binding grammar. The loader cannot prove the Fairness Law from
an undeclared relationship between a scaling value and an engine field.

**Proposed contract:** version a closed scaling-source union over implemented committed state
(literal, tier, purchased generator count, Founder carry counter, and attended-quality grade),
each producing a bounded exact integer under a published formula. A tenant declares each input's
axis and exact destination from its schema-owned destination registry. Ranked destinations are
classified `power|breadth|presentation`; ranked rows reject any non-neutral tier-derived path to
`power`. Session creation resolves the complete sorted input map once and stores those canonical
bytes in genesis; unknown sources/destinations and duplicate bindings fail catalog load.

### C5 — The faucet governor lacks time, arithmetic, and persistence semantics

“Per day,” score conversion, cap consumption, bot reduction, and forfeit are not defined precisely.
Wall-clock day boundaries create timezone/deploy ambiguity; a resolve retry can spend the cap twice;
and `score × conversion_ppm` may overflow or round differently across runtimes.

**Proposed contract:** the catalog row names the payout resource and exact integer score domain,
`sends_per_window`, `per_send_cap`, `conversion_ppm`, bot multiplier, and cap reason key. Windows
are UTC epochs `floor(server_ms / 86_400_000)`. Conversion uses checked integer arithmetic with a
persisted remainder only if the owner wants carry; otherwise it declares floor once. A
founder/minigame/window counter is updated in the payout transaction and keyed by session ID so
retry is a no-op. The receipt reports gross, applied, and forfeited amounts plus the visible reason
key. All literal values are balance data and joining the epoch artifact set is a mint.

### C6 — “Payouts are intents” permits a client to claim a result

The current public intent envelope accepts player commands, not server-certified minigame results.
Adding `claim_payout {score:...}` would violate the server-authoritative law; hiding an internal
side write behind the word intent would violate replay/event ownership.

**Proposed contract:** add a server-internal `apply_minigame_result` command that is never decoded
from the player HTTP intent route. It consumes `(session_id, immutable_result_hash)` after the
session repository locks and verifies `resolved`; the production/economy transition receives only
the already-governed applied credit and records its source session in replay inputs, receipt, and
events. The session `paid` mark and Company stream commit share one declared transaction/lock order
or an outbox projection with an exactly-once claim—choose one explicitly and fault-inject every
boundary.

### C7 — AI fallback and ranked identity are not closed data

“Solo-by-design, bot backfill, or NPC-partner” does not state exact rows, timing, bot identity,
rating eligibility, or how a bot result is replayed. “Declared reduced rate” has no literal owner.

**Proposed contract:** make fallback a closed union with exact fields. Solo has no peer; bot
backfill declares deterministic bot profile/version, backfill deadline, ranked eligibility, and
payout multiplier; NPC partner declares profile/version and control boundary. Bot decisions use
the tenant's seeded substream and the same legal-command validator. Every session genesis records
the selected fallback/profile version. Ranked eligibility and payout reduction derive from that
record, never from a client flag.

### C8 — Live PvP is simultaneously deferred and required

The open questions defer match actors/matchmaking, while MP1 and AC1 require `live_pvp` through the
composed server. Without disconnect, reconnect, authoritative-clock, and result-finalization rules,
the state machine cannot support that mode honestly.

**Proposed contract:** Phase A implements `solo` and `async_snapshot` only and reserves—but rejects—
`live_pvp` at registry load until a PvP Match Service RFC owns actor lifecycle, matchmaking,
disconnect/reconnect, clock, spectator/privacy, and bot-backfill timing. Combat adapters can prove
tenant conformance against deterministic solo/snapshot fixtures now. If live PvP must ship in this
RFC, those contracts become blockers here rather than an open question.

### C9 — Registry presentation and unlock fields repeat withdrawn UI paths

`unlock_condition_ref` has no grammar and `era_skins` contradicts UI Foundation: surfaces use one
token theme, never per-feature skins. The combat children are also not implemented tenants today,
so AC6 overclaims their readiness.

**Proposed contract:** replace the fields with the UI-ruled unlock union
`always|fact_equals` over the authoritative shell/snapshot fact manifest; remove `era_skins`.
Registry rows carry a copy key and optional presentation capability IDs only. AC6 uses a fixture
tenant first; duel/lane registration becomes each combat child's acceptance item when that engine
is implemented, not a platform-foundation prerequisite based on draft code.

### C10 — Offline quality has no observable or decay formula

“Recent performance,” “quality tier,” and “~31 days” define neither stored state nor a deterministic
transition. It also risks turning an optional minigame into mandatory labor if automated output
decays toward zero.

**Proposed contract:** either defer `offline_quality` entirely, or specify its owner catalog,
score-to-grade thresholds, exact fixed-grid decay/remainder, floor, cap, save scope/reset, reason
keys, and neutral no-play behavior. The floor must preserve the base offline-production promise;
quality may improve a named automation slot but never reduce the canonical 90% offline rate. Until
those literals are owner-ruled, the scaling-source union rejects `offline_quality`.

### C11 — Artifact/replay identity and migration scope are absent

The registry, tenant schemas, fallback profiles, faucet policy, and session genesis all affect
verification, but none joins epoch identity or the verifier. No DB migration or cross-runtime
corpus is named.

**Proposed contract:** one `minigames` artifact contains the sorted registry and referenced policy
rows; tenant engine versions/schemas are immutable code identity. Session genesis pins constants
hash, artifact bytes/hash, engine version, seed/substream, scaling inputs, fallback selection, and
initial snapshot. A shared verifier replays every command and compares snapshot/result/event bytes.
The implementation lands append-only migrations, crash/fault integration tests, ≥50-command
Go/TS (or single-runtime tenant) corpus, formula artifact rows for scaling/payout, and a protocol-
compliant balance mint before any production tenant or payout is enabled.

## Changelog

- 2026-08-03: created (draft) — the platform.
- 2026-08-04: C12–C18 ruled — body reconciled to the C1–C11 rulings; session table = verification-queue pattern, payout = Exit-style multi-write server transition, faucet = provision-grid clock, offline_quality decays to neutral floor; all STRUCTURE ruled by reusing shipped patterns, all NUMBERS deferred to balance data.
- 2026-08-04: C1–C11 ruled — Clout payout faucet removed (my error, caught), DB-authoritative sessions with queue-style idempotency, server-certified (not client-claimed) payout, live_pvp deferred (solo+async only ship), era_skins removed (UI law), offline_quality decays to neutral floor not zero, epoch/verifier identity named. Accepted, scope narrowed.
- 2026-08-04: Codex acceptance review found C1–C11. The draft is not implementable yet: Clout
  ownership contradicts the ruled single faucet, and session, tenant, scaling, payout, fallback,
  live-PvP, unlock, offline-quality, and replay contracts require owner rulings.
- 2026-08-04: post-ruling implementation pass found C12–C18: design ownership is settled, but the
  active body and exact persistence/wire/arithmetic contracts still need reconciliation/literals.
