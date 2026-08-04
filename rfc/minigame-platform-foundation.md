# RFC: Minigame Platform Foundation

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/03` (the minigame suite, clocks, economy hooks, AI-fallback law), `design/03 §12b` (the tier→minigame scaling seam + Fairness Law), `design/05 §4` (PvP table, punch-down multiplier)
- **Research:** `neopets-systems.md §4` (the portal faucet model: N-sends × per-game cap × conversion ratio, monthly re-tune), `kol-puzzle-pirates.md §B` (minigame-as-labor, offline-output quality grades), `creature-battler.md` / `lane-pusher-design.md` (the combat instances)
- **Depends on:** Production + Run Genesis + Transport + Gameserver Composition (implemented); Combat Shared Kernel (implemented — the first two platform tenants are the duel/lane engines); Account/Clout (Clout foundation — reward payouts)
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

A minigame session is `{session_id, minigame_id, founder, scaling_inputs (frozen), seed,
mode: solo|async_snapshot|live_pvp, state, result|null}`. Lifecycle: create (server-resolves
scaling inputs, freezes them, seeds from the save stream — the combat C5 pattern) → play (the
minigame's own engine, a platform tenant) → resolve (server-authoritative: solo/snapshot resolve
by replay like combat; live PvP via the match actors on the transport layer) → payout (MP4). No
minigame reads live company state mid-session; the seam is always the frozen integer inputs
(replay-safe by the RA rule, exactly as combat proved).

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

Resolve → payout hook: Clout (achievement/rating), tier-scaled cash (governed by MP3), rating
updates (real Elo, no rubber-band — the combat-bots contract), season facts. **Every minigame
declares its AI fallback** (the design law): solo-by-design, bot backfill, or NPC-partner — a
minigame with no fallback is a load error (the whole game must be playable at zero other players
online). Payouts are intents (evented, replay-logged) — a minigame result enters the economy
through the same closed intent surface as everything else, never a side channel.

### MP5 — The registry

`balance/minigames/*.json`: `{minigame_id, engine_ref, clock: turn|realtime|wallclock|async,
scaling_inputs[], payout, fallback, unlock_condition_ref, era_skins}`. The two combat engines
register as the first tenants (their existing catalogs become platform-conformant); the board
games, Market, and arcade toys are later content on this registry. `unlock_condition_ref` gates
by catalog fact (staggered across the tier arc — the CC hour-5-40-sag fix); the platform never
invents unlock logic.

## Acceptance criteria

1. Session lifecycle: create→play→resolve→payout for a fixture minigame (a trivial platform
   tenant, not a real game) against the composed gameserver; scaling inputs frozen at create,
   unread thereafter (replay-safe proof).
2. Fairness Law: loader rejects a scaling input wired to a ranked power stat; `breadth` scaling
   into ranked play passes; `offline_quality` charges/decays per the grade ladder.
3. Faucet governor: the daily cap forfeits excess (reason-keyed); a bot match pays the reduced
   rate; per-epoch re-tune follows BALANCE-CHANGE.
4. Payout-as-intent: a resolved result enters the economy through the intent surface, evented and
   replay-logged; no side-channel write exists (grep-proven).
5. Fallback law: loader rejects a fallback-less minigame; the fixture tenant's solo path completes
   at zero connected peers.
6. Registry: the combat duel + lane engines register as conformant tenants without engine changes
   (adapter only).

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

- 2026-08-03: created (draft) — the platform; combat engines are its first tenants; the §12b seam
  and the Neopets faucet governor made structural.
- 2026-08-04: Codex acceptance review found C1–C11. The draft is not implementable yet: Clout
  ownership contradicts the ruled single faucet, and session, tenant, scaling, payout, fallback,
  live-PvP, unlock, offline-quality, and replay contracts require owner rulings.
