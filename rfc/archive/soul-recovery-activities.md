# RFC: Soul Recovery Activities (the cozy content — touch-grass v1)

- **Status:** implemented fixture-first — SR-C1–SR-C14 ship and passed the designated cross-party
  review. Production activation remains owned by the First Content Epoch; toy rendering and final
  disclosure presentation remain carried to the UI successor.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/02 §8` (recovery: deliberate, opportunity-costed, produces nothing),
  `design/03 §5` (the cozy category is the SOLE Soul-recovery source — board games pay but never
  restore; ruled 2026-08-06)
- **Depends on:** Soul Foundation (SB1–SB27 — the archived session/heartbeat/suppression substrate),
  API Foundation amendment (the authenticated coordinator surface), and UI Foundation successor
  work for the carried toys/disclosures. This RFC adds the fixture catalog/copy, coordinator API,
  rate limiter, and framework-neutral heartbeat scheduler; the Soul transition mechanics remain
  owned by the archived foundation.
- **Planning:** `planning/archive/soul-recovery-activities/`

## Summary

The first three production touch-grass activities — **`defrag`** (a zen tile-layer: watch the disk
come to order), **`server_room`** (a no-goal toybox: arrange the racks, water nothing, help no one),
**`repot`** (a short ritual: repot the server-room plants — "daily" is tonal only, SR-C7) — as
`recovery_activities` catalog rows + the authenticated coordinator API (SR-C1/SR-C10) on the shipped
Soul substrate. Each costs deliberate attended time under production
suppression and **produces nothing** except Soul recovery: no resource, no score, no achievement,
no record of how well you did, because there is no "well." The tooltip says so, sincerely — the one
place the game's curtain-pulling voice goes quiet.

## Motivation

Soul's drain and recovery are both reviewed machinery awaiting owner content; this RFC activates
RECOVERY only (the first production debit source — and the training-data ending's production
reachability — belongs to a later Events/longevity content RFC; SR-C8). The research finding this rests on:
**"produces nothing" and "cheap to serve" are the same property** — a zero-reward toy has nothing
to cheat, so the server needs only the session lifecycle (which exists), and the entire interactive
content is client-side presentation. This is the cheapest content RFC in the game and it completes
the game's emotional core loop (drain → grey pet → touch grass → the pet knows you again).

Out of scope: additional activities (this union is open; rows are data), any activity that grants
any resource (forbidden by the recovery covenant — permanently), and audio (the audio research
commission owns the soundscape later).

## Specification

### SR1 — The three catalog rows

Exact `recovery_activities` rows per the SB10/SB17 grammar extended by SR-C2
(`{activity_id, duration_attended_ms, recovery_amount, toy_kind, reason_key, title_copy_key,
description_copy_key, disclosure_copy_key}` — the literal rows are in the SR-C12 ruling):
- **`defrag`** — the zen tile-layer. Medium duration. Flavor: "Defragment the disk."
- **`server_room`** — the toybox. Long duration (the deep rest). Flavor: "Sit in the server room."
- **`repot`** — the daily ritual. Short duration (the accessible one). Flavor: "Repot the plants."
Durations/recovery amounts are balance data (design intent: short/medium/long tiers so every
playstyle has an accessible entry; the deeper rest restores more). All copy keys resolve through
the copy pipeline; the voice for this content is **sincere** (the pet register, never the corporate
register — `design/08`'s tonal law).

### SR2 — The client toys (presentation-only, deliberately non-authoritative)

Each activity has a client-side interactive toy rendered during the session. **The toy is pure
presentation: it reports nothing, scores nothing, and the server never sees it.** Consequences,
stated explicitly because they are the design:
- The toy MAY use client-local randomness freely (tile shapes, plant colors) — legal because no
  outcome exists to make deterministic; nothing it does is replayed or verified.
- Closing the tab mid-activity loses nothing: the session persists server-side (SB19), the
  heartbeat pauses (SB24 — absence pauses, never kills), reconnect rotates the token and the toy
  resumes fresh. The toy has no state worth saving — that is the point.
- No fail state, no completion meter beyond the session's own progress, no "score screen." The
  session-progress display (from the SB27 progress response) is the only number on screen.
- Toy interaction is NOT required for progress — presence is the activity (the heartbeat is the
  only input that matters). An idle founder staring at the tiles recovers identically. The toy is
  there to be pleasant, not to be performed.

### SR3 — The disclosure (the quiet tooltip)

Each activity's start tooltip states plainly: *"This produces nothing."* — the inverse of the
curtain-pull: everywhere else the game discloses the machinery behind a fake reward; here it
discloses the absence of one, sincerely. Copy through the pipeline; the exact lines are content.

### SR4 — Activation

The rows land FIXTURE-ONLY in this RFC (SR-C13); the production mint of the full dependency
artifact set belongs to the dedicated **First Content Epoch RFC** (owner-gated, minted together —
no partial chain). UI surfacing is the `soul_recovery` surface (the Minigame & Recovery API +
Surface RFC + Game-UI screens); this RFC ships the rows, the coordinator API, the scheduler module,
and the copy — toy components are carried acceptance debt on the UI successor.

## Deviations from design

None — this implements `design/02 §8` recovery and the 2026-08-06 cozy-only-recovery ruling
exactly.

## Acceptance criteria

1. Three exact-key rows validate in both loaders; copy keys resolve; the artifact with rows passes
   the full Soul loader (bands/policy/sources unchanged).
2. A full production-shaped session per activity (fixture epoch): start → heartbeat accumulation →
   resolve grants exactly the catalog amount (saturating at max), byte-replayed; cancel and
   watchdog paths unchanged.
3. Grep-proof: no resource/score/achievement/event grants anywhere in the toy layer; the server
   receives no toy input.
4. **CARRIED TO UI SUCCESSOR:** the toys render and resume-after-reconnect; presence-without-interaction recovers identically
   (no interaction requirement).
5. **CARRIED TO UI SUCCESSOR:** the disclosure tooltip is present for all three, in the sincere register (copy-review flag, not
   CI-provable).

## Open questions

- Whether `repot` seeds a tiny persistent visual (the plant grows across sessions — pure cosmetic,
  client-preference storage, never save state) — nice, cheap, and honest; decide at implementation.
- Ambient audio — deferred to the audio commission; the toys ship silent-but-pleasant first.

## Acceptance-review blockers (Codex, 2026-08-07)

The activity concept matches Soul's covenant, but the current repository can only exercise it from
Go. The draft calls the work “catalog rows + client presentation” while the public coordinator and
presentation catalog do not exist. Those seams must be owned explicitly before production rows mint.

### SR-C1 — The recovery coordinator has no public API or transport surface

`StartSoulRecovery`, `ProgressSoulRecovery`, `CancelSoulRecovery`, and `ResolveSoulRecovery` are
internal production methods with no publicapi/chi route, transport command, or client decoder. A toy
cannot acquire/rotate the progress token or send a heartbeat. “Zero server mechanics” is therefore a
false scope claim.

**Proposed contract:** add a narrow authenticated coordinator API (recommended HTTP like intents,
because heartbeats are bounded commands, not stream events) with exact start/reconnect, progress,
cancel, and resolve schemas from SB25–SB27. Founder identity comes only from the session; rate-limit
progress per session; reconnect stores only the newest token client-side. Map typed rejections,
including `recovery_token`, `exclusive_activity`, and `session_expired`, without exposing claim leases.

### SR-C2 — Recovery rows cannot identify a toy or its copy

The exact row is only `{activity_id,duration_attended_ms,recovery_amount,reason_key}`. It has no
`toy_kind`, title, description, or disclosure key. The copy pipeline cannot infer “Defragment the
disk” or “This produces nothing” from an ID, and hardcoding an ID→component map contradicts the
content-as-data claim.

**Proposed contract:** extend the never-yet-epoch-pinned Soul recovery row in this successor with
exact `toy_kind`, `title_copy_key`, `description_copy_key`, and `disclosure_copy_key` fields, all
closed/registered in Go and TypeScript. Enumerate `toy_kind:defrag|server_room|repot` and register
every key in the tracked copy catalog. If row grammar must remain frozen, create a separately pinned
presentation artifact keyed one-to-one by activity ID instead; do not create two optional authorities.

### SR-C3 — The UI foundation and recovery surface are not available

UI Foundation remains blocked on C9–C11, and Game UI explicitly excludes minigame surfaces. No route,
component input, heartbeat lifecycle, focus/background behavior, or reconnect view exists. AC4 cannot
pass by adding three isolated components.

**Proposed contract:** depend on an accepted UI Foundation and define one `soul_recovery` surface
contract: session receipt + local toy seed in, coordinator callbacks out, one heartbeat scheduler,
visibility/background pause behavior, token replacement on reconnect, cancel affordance, progress
display, and terminal return. Toy components receive presentation data/callbacks only and import no
transport internals.

### SR-C4 — The three “rows” have no literal bytes

Durations, recovery amounts, reason/copy keys, and even final row ordering are absent. A production
artifact mint and balance-harness baseline cannot review an adjective such as short/medium/long.

**Proposed contract:** provide the complete byte-sorted literal rows (`defrag`, `repot`,
`server_room`) with provisional exact integers and copy keys. Validate duration against
`max_session_wall_ms` and the client heartbeat policy; validate recovery against the Soul domain.
Changing these values later is a normal balance mint.

### SR-C5 — The first Soul mint is a multi-artifact activation, not one row edit

No production epoch currently pins Soul. A Soul-bearing bundle requires the Fiscal artifact and the
bumped Minigame/Pet schemas, activating the Founder v17→v20 chain at the next run boundary. SR4 calls
this “the standard epoch mint” without enumerating the accepted artifact set or migration scenario.

**Proposed contract:** enumerate the exact first-Soul artifact bundle and epoch transition, including
the production-safe minigame and pet artifacts it requires. Add a real pre-mint Founder→Exit→v20
activation fixture proving all four scalar versions initialize together and existing runs finish
under old bytes. Do not mint a partial dependency chain.

### SR-C6 — Heartbeat cadence and browser lifecycle are unspecified

SB24 defines the server ceiling, not how often the client beats, what hidden/background tabs do, or
how network retries interact with token rotation. Those choices determine whether ordinary browser
throttling pauses every session.

**Proposed contract:** choose a catalog-derived or literal client cadence strictly below
`recovery_beat_ceiling_ms`, send only while the surface is active/visible, treat a missed ceiling as a
pause, and reconnect via start before resuming. Never replay queued beats after sleep. A fake-clock
browser test covers foreground, duplicate retry, hidden-tab pause, reconnect rotation, and watchdog.

### SR-C7 — `repot` says “daily” but no daily gate exists

The Soul session state has no per-activity cooldown or completion ledger. Calling this a daily ritual
can be flavor or a new authority; the draft does not decide.

**Proposed contract:** make “daily” tonal only in v1: `repot` is repeatable like the other activities,
with no streak, cooldown, or reward multiplier. If once-per-day is intended, it needs a Founder-
attended-day state contract and cannot hide in client presentation.

### SR-C8 — The motivation overstates the production drain state

No production debit source is enabled; only the pure debit machinery and fixture source exist. The
training-data ending is therefore not presently reachable through production content. Saying “the
drain is live” reverses the archival record.

**Proposed contract:** correct the motivation to say recovery and drain machinery are both awaiting
owner content, and this RFC activates recovery only. A later Events/longevity content RFC owns the
first production debit source and ending reachability.

## Changelog

- 2026-08-07: created (draft) — Wave-B; the owner-content mint for Soul recovery.
- 2026-08-07: Codex acceptance review filed SR-C1–SR-C8; implementation blocked pending owner rulings.

## Owner rulings on SR-C1–SR-C8 (2026-08-07)

All accepted; scope decision on C1/C3. Body reconciliations noted.

- **SR-C1 — accepted; the "zero server mechanics" claim is RETRACTED** (false — no public surface
  existed). A narrow authenticated HTTP coordinator API (the intents pattern — heartbeats are
  bounded commands, not stream events) with the exact SB25–SB27 start/reconnect, progress, cancel,
  resolve schemas; Founder identity from the session only; per-session progress rate limit; typed
  rejections (`recovery_token`, `exclusive_activity`, `session_expired`) without exposing claim
  leases. Housed as a declared **API-Foundation amendment** implemented with this RFC.
- **SR-C2 — accepted.** The never-epoch-pinned recovery row is extended IN THIS RFC with exact
  `toy_kind` (closed: `defrag | server_room | repot`), `title_copy_key`, `description_copy_key`,
  `disclosure_copy_key` — one authority (no separate presentation artifact); all keys registered in
  Go + generated TS.
- **SR-C3 — accepted.** This RFC depends on an ACCEPTED UI Foundation (C9–C11 are now on the
  critical path for playability) and defines one `soul_recovery` surface contract: session receipt +
  local toy seed in, coordinator callbacks out, one heartbeat scheduler, visibility/background pause,
  token replacement on reconnect, cancel affordance, progress display, terminal return. Toy
  components receive presentation data/callbacks only.
- **SR-C4 — accepted; provisional literals ruled now** (exact integers, normal balance mints later):
  `defrag {duration 900000, recovery 12}` · `repot {duration 300000, recovery 5}` ·
  `server_room {duration 2700000, recovery 30}` — byte-sorted rows with copy/reason keys; duration
  validated < `max_session_wall_ms`; recovery validated within the Soul domain.
- **SR-C5 — accepted.** The first-Soul mint is enumerated as the FULL bundle transition (fiscal +
  the bumped minigame/pet schemas + soul), with a real pre-mint Founder→Exit→v20 activation fixture
  proving all four scalar versions initialize together and existing runs finish under old bytes. No
  partial chain mints.
- **SR-C6 — accepted.** Client cadence = catalog-derived, strictly below the ceiling (ruled:
  `recovery_beat_ceiling_ms / 3`); beats sent only while the surface is active/visible; a missed
  ceiling is a pause; resume via reconnect-start; queued beats are NEVER replayed after sleep. The
  fake-clock browser test covers foreground, duplicate retry, hidden-tab pause, reconnect rotation,
  and watchdog.
- **SR-C7 — RULED: "daily" is TONAL ONLY in v1.** `repot` is repeatable like the others — no
  streak, cooldown, or multiplier (a once-per-day gate would be a new Founder-attended-day
  authority; if ever wanted, it's an explicit successor, never client presentation).
- **SR-C8 — accepted; the motivation is CORRECTED** (body reconciled): both drain and recovery
  machinery await owner content; this RFC activates RECOVERY only; the first production debit
  source — and with it the training-data ending's production reachability — belongs to a later
  Events/longevity content RFC.

## Implementation blockers (Codex, 2026-08-07)

The first rulings select the right product shape. The remaining blockers are the exact API and mint
bytes, plus contradictions left in the normative body. Implementing them without a second ruling
would invent public authority, error behavior, and an owner-gated balance epoch.

### SR-C9 — The normative body still claims the retracted scope

The header says this RFC adds “ONLY catalog rows + client presentation + copy. Zero server
mechanics,” while SR-C1 puts a new authenticated coordinator API in scope. SR4 says the Game-UI
consumer ships the rows/toys even though UI Foundation remains unaccepted. SR1 still describes copy
keys as an informal suffix rather than the ruled exact row fields.

**Proposed contract:** reconcile SR1/SR4/dependencies/ACs. Mark the implementable landing as catalog,
coordinator API, and internal/browser-independent contract tests; keep actual toy components blocked
on UI Foundation and record them as carried acceptance debt rather than claiming them shipped.

### SR-C10 — Public coordinator request authority and routes are not exact

SB25–SB27 define internal responses, but not HTTP route names or request bodies. The production start
method currently accepts client-supplied session, Founder, and Company IDs; the authenticated API
must decide which are server-derived. Resolve/cancel token semantics and HTTP status/error mapping
are also absent.

**Proposed contract:** enumerate four literal routes and exact JSON bodies. Recommended:
`POST /api/v1/soul-recovery/start {activity_id}` (server creates session ID and resolves active
Founder/Company), `/progress {session_id,progress_token}`, `/cancel {session_id}`, and `/resolve
{session_id}`. Founder/Company IDs never cross the public wire. Define reconnect as repeated start,
and map every deterministic error to one closed `{category,detail}` pair while transient/store errors
remain 5xx and never become gameplay rejections.

### SR-C11 — The per-session heartbeat limiter has no contract

“Rate-limited heartbeats” provides no capacity, refill rule, key lifetime, status, or interaction
with the account limiter. These values affect ordinary browser behavior around the ruled
`ceiling/3` cadence.

**Proposed contract:** make the database heartbeat transition the semantic authority and add a
small API abuse limiter keyed by session ID. Specify literal Phase-0 burst/refill values that safely
admit the catalog-derived cadence, return `429 rate_limited/recovery_progress`, and evict at terminal
or bounded idle TTL. The limiter must not consume/rotate the progress token on rejection.

### SR-C12 — Presentation and copy identifiers are still missing

The ruled row adds four fields, but no literal `reason_key`, title, description, or disclosure keys
are supplied for any activity. The copy catalog cannot be generated from prose labels.

**Proposed contract:** list the exact eight-field rows for all three activities and the exact
copy entries. Recommended mechanical key family: `soul.recovery.<activity>.{title,description,
disclosure}` plus `soul.recovery.<activity>.reason`; the shared disclosure may reuse one key only if
the row explicitly names it. Copy text must pass the existing tone/provenance gates.

### SR-C13 — The first-Soul mint is described but not enumerable

The live epoch manifest contains none of Fiscal/Minigame/Pet/Soul. “Full bundle transition” does not
name the next epoch, artifact order/paths, exact dependency bytes, changelog, or baseline changes.
Those are owner-sensitive balance/history decisions, not implementation details.

**Proposed contract:** enumerate the complete next epoch manifest and every artifact file, or rule
the implementation fixture-only until a dedicated content-mint RFC. Recommended: land row grammar,
literal fixture, API, and activation integration now; keep the production mint owner-gated until the
full dependency artifact bytes are reviewed together.

### SR-C14 — The browser contract is not implementable before UI Foundation

The cadence/focus behavior is ruled, but no accepted component host, routing, request client, or
test environment owns it. Writing toy components now would create the UI seam that SR-C3 says is a
dependency.

**Proposed contract:** keep a framework-neutral scheduler/state-machine module in scope only if its
exact input/output interface is written here; otherwise defer all browser/toy code to the accepted
UI successor. Either choice must leave AC4/AC5 visibly open rather than satisfied by catalog/API
tests.

## Owner rulings on SR-C9–SR-C14 (2026-08-07) — the literals

- **SR-C9 — accepted; bodies reconciled in this edit.** The implementable landing = catalog rows +
  coordinator API + internal/browser-independent contract tests; toy components stay blocked on the
  UI Foundation successor and are recorded as CARRIED ACCEPTANCE DEBT (AC4/AC5 marked open), never
  claimed shipped.
- **SR-C10 — RULED (exact routes; supersedes the `/api/v1/recovery/` naming in the API/Surface
  draft, which is amended to match):** `POST /api/v1/soul-recovery/start {activity_id}` (server
  creates the session ID and resolves the active Founder/Company — no ID crosses the wire; repeated
  start = reconnect with token rotation) · `POST /api/v1/soul-recovery/progress
  {session_id, progress_token}` · `POST /api/v1/soul-recovery/cancel {session_id}` ·
  `POST /api/v1/soul-recovery/resolve {session_id}`. Founder/Company IDs never appear in any public
  request or response. Every deterministic rejection maps to one closed `{category, detail}` pair in
  the API Foundation's error envelope; transient/store errors are 5xx and NEVER become gameplay
  rejections.
- **SR-C11 — accepted (literals).** The database heartbeat transition is the semantic authority; the
  API adds an abuse limiter keyed by session ID: **burst 6, refill 1 token per `ceiling/6` ms**
  (catalog-derived — safely admits the ruled `ceiling/3` cadence with jitter), rejection
  `429 rate_limited/recovery_progress`, evicted at terminal state or a 15-minute idle TTL. A limiter
  rejection NEVER consumes or rotates the progress token.
- **SR-C12 — RULED (the literal eight-field rows + copy identifiers).** Key family
  `soul.recovery.<activity>.{title, description, disclosure}` + `soul.recovery.<activity>.reason` —
  per-activity keys, no shared-key exception. The rows (byte-sorted; durations/amounts provisional
  per SR-C4):
  - `defrag {duration_attended_ms: 900000, recovery_amount: 12, toy_kind: "defrag",
    reason_key: "soul.recovery.defrag.reason", title_copy_key: "soul.recovery.defrag.title",
    description_copy_key: "soul.recovery.defrag.description",
    disclosure_copy_key: "soul.recovery.defrag.disclosure"}`
  - `repot {300000, 5, "repot", "soul.recovery.repot.reason", "soul.recovery.repot.title",
    "soul.recovery.repot.description", "soul.recovery.repot.disclosure"}`
  - `server_room {2700000, 30, "server_room", "soul.recovery.server_room.reason",
    "soul.recovery.server_room.title", "soul.recovery.server_room.description",
    "soul.recovery.server_room.disclosure"}`
  Copy text lands through the pipeline's tone/provenance gates (the sincere register; each
  disclosure states "This produces nothing." in the activity's own voice).
- **SR-C13 — RULED: fixture-only now; the production mint belongs to the dedicated First Content
  Epoch RFC** (shared ruling with TP-C18): row grammar, literal fixture, API, and activation
  integration land in this RFC; the full dependency artifact set (fiscal + minigame/pet bumps +
  soul + pitch) mints together, owner-gated, with its own reviewed epoch bytes and changelog.
- **SR-C14 — RULED: the framework-neutral heartbeat scheduler IS in scope, with this exact
  interface** (no DOM, no framework imports): constructor input `{session_id, progress_token,
  beat_interval_ms, transport: {progress(session_id, token) → Promise<response>},
  now() → ms, visibility: {subscribe(cb(visible: bool)) → unsubscribe}}`; outputs = callbacks
  `{on_progress(response), on_pause(reason: hidden|network), on_resume(), on_token_rotated(token),
  on_terminal(kind)}`; behavior per SR-C6 (beat only while visible, missed ceiling = pause, resume
  via reconnect-start upstream, never queue/replay beats). Toy components and all DOM code defer to
  the UI successor; AC4/AC5 remain visibly open (carried debt), satisfied only by the surface work.
