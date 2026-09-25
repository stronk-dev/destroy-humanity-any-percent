# Minigame Platform

The implemented foundation owns durable sessions, pure tenant engines, immutable pinned policy,
and the server-authored resolution transaction. No public intent accepts a score or result: a
tenant certifies the result, and the platform resolves its payout and Founder effects.

## Authoritative sessions

`minigame_sessions` is the Postgres authority for Phase-A `solo` and `async_snapshot` sessions.
`live_pvp` is deliberately rejected until its separate service owns realtime lifecycle rules.
Every row freezes the Founder/Company run identity, pinned constants hash, minigame and engine
versions, unsigned seed, complete integer scaling-input object, mode, and genesis snapshot. A
database trigger rejects any later mutation of those fields.

Active sessions begin at revision 1. A server-side play or resolve command claims the row with a
database-generated UUID token after locking Founder then session. Concurrent workers cannot both
claim it; a crashed claim can be replaced after the same five-minute lease used by replay
verification. A completed play advances the revision exactly once and returns the row to active.
A resolved result advances once, clears the claim, records its completion time, the exact terminal
receipt, and the committed Company and Founder revisions, and is immutable.

Resolution exposes a transaction-bound service write rather than a public client command. A
terminal play returns an opaque certification whose identity and result fields cannot be populated
outside the platform package. The coordinator locks Founder then Company, verifies the session's
claim and immutable command history, advances the attended-day faucet, applies the Company payout,
updates Founder rating and offline quality, finalizes the session, and appends both save revisions,
events, and replay logs in one transaction. Fault injection covers every write boundary. A retry is
answered from the durable session-ID idempotency receipt without executing the tenant or faucet
again.

## Tier unlock arm (TT-PA2)

`unlock_condition` also accepts `{"kind":"tier_at_least","tier":0..9}`, with an optional
`"exit_history_at_least": n ≥ 0`. The Go and TS loaders reject unknown keys, non-integers and
out-of-domain values. The composed start coordinator evaluates `UnlockCondition.TierUnlockFailure`
from pinned server state only: the Company `tier` and `len(founder.exit_history)`. It rejects
before tenant creation with `ErrMinigameTierRequired` or `ErrMinigameCurriculumExitRequired`. The
public error details `not_eligible/tier_required` and `not_eligible/curriculum_exit_required` are
mapped in the API batch (TT-PA4).

## Server-sampled command time (TT-PA1)

The claim transaction samples the database clock once, as
`floor(extract(epoch FROM clock_timestamp())*1000)`, before the tenant runs. The sampled value
reaches the tenant as `ApplyInput.ServerTimeMs`. Both command-append statements, nonterminal and
terminal, insert exactly that value as `server_ts_ms` and never read the clock again at insert
time. Verification replay passes each row's persisted `server_ts_ms`. A rejected command appends
nothing, so its sample is discarded. Tenants that don't measure time, such as Pitch, ignore the
field. No tenant reads a clock, and no API accepts a client time.

## Authenticated composed API

The composed server mounts the generated private-v1 registry as the only route authority for
minigame create, current, command, and resolve requests and for Soul-Recovery start, progress,
cancel, and resolve requests. Founder identity always comes from the access token. Request schemas
reject Founder IDs, Company stream IDs, and server-clock coordinates, and a foreign session ID is
indistinguishable from a well-formed missing session ID.

The real-socket integration path proves a complete Pitch session and a complete Soul-Recovery
session through the composed binary. Recovery reconnect rotates the progress token; the old token
fails, attended heartbeats reach eligibility, terminal retry returns identical durable bytes, and
an over-age session terminates through the watchdog path. Authenticated command flooding reaches
the shared account limiter without changing the active session; after refill, the exact prior
snapshot remains current and ordinary play resumes. Recovery heartbeats retain their additional
per-session limiter.

The internal recovery command kind remains `resolve_soul_recovery` or
`cancel_soul_recovery` for replay. The public durable terminal receipt uses the API grammar's
`action: "resolve" | "cancel"`; the distinction is explicit so internal command vocabulary cannot
leak across the public schema boundary.

## Client surface

`client/src/game-ui/minigame/` holds the MA3/MA-C9 client surface. Components never call `fetch`:
`MinigameSessionPort` (`session-port.ts`) is the only transport seam. It is built from the
generated operation table and typed with the generated DTOs. A non-2xx response whose body is
exactly `{category, detail}` with a known category becomes a `MinigameAPIError`; any other failure
becomes a `MinigameTransportError`.

`MinigameSessionSurface.svelte` owns the chrome, lifecycle and errors. Its closed states are
`loading | active | paused_reconnect | required_terminal | error`, plus a host-level launcher for a
`{kind:"none"}` current answer.

- **Rejections:** `session-surface.ts` maps exactly the server's `minigameErrorJSON` pairs to one
  of five effects: stay, refetch, launcher, pause or error. Each effect has a copy notice and a
  flag saying whether it clears the selection. An unit test checks that the table's keys match the
  server source. Unlisted pairs, `invalid/*` and `internal_invariant/*` render the error state.
- **Idempotent retries:** a transport failure moves the surface to `paused_reconnect`, which
  retries automatically every second (at most five times) and also offers a manual Retry. A retry
  resends the same `command_id` and body, relying on the MA-C13 receipt. A create keeps its
  idempotency key until it succeeds.
- **Terminal receipt:** it is rendered once and moves focus to the surface heading. The host's
  `onTerminal` then runs a single authoritative refresh.

The tenant-surface registry (`tenant-registry.ts`) is keyed by the same `(engine_ref,
engine_version)` arm as the pinned `balance/minigame-api/first-content.json` tenant rows. Loading
fails in either direction: a tenant row with no child, or a child with no tenant row. A session
whose arm or `minigame_id` is not registered renders the error state. Pitch `1.0.0` registers one
presentation-only child, `PitchTable.svelte`: native checkboxes (capped at `play_size`, with
`cap.pitch_play` text shown), a play button, shop offer buttons, and close-shop.

The Pitch snapshot carries only instance IDs. The table resolves card and hack copy from the pinned
`balance/pitch.json` bytes bundled with the client, and only when their SHA-256 equals the
snapshot's `pitch_content_hash`. A mismatch renders `minigame.error.content_mismatch`; there is no
deploy-current fallback.

Surface copy lives in `copy/catalog/minigame-surface-candidate.json`. It is plain functional
wording drafted by the implementer and awaits owner adoption; it is not ruled copy.
`test/minigame-surface-browser.test.ts` runs in all three browsers and covers:

- axe checks in every state;
- keyboard selection with the play-size cap;
- the complete launcher → table → shop → terminal flow;
- same-ID retry;
- selection clearing on a revision conflict;
- failing closed on an unknown tenant or foreign content;
- the create-key hold.

`client/src/game-ui/soul/SoulRecoverySurface.svelte` is the SR-C3 `soul_recovery` surface. It uses
`SoulRecoveryPort` (`recovery-surface.ts`) with the generated start/progress/resolve/cancel
operations, and the pinned `balance/soul/first-content.json` plus the prestige catch-up ceiling
bundled with the client.

- **Picker:** lists every pinned activity with its title, description, duration and disclosure
  small print. The disclosure is also the Begin button's `aria-describedby`.
- **Heartbeat:** one `RecoveryScheduler` beats every `recovery_beat_ceiling_ms / 3` (SR-C6), and
  only while the page is visible. Progress is shown with a native `<progress>`.
- **Pauses:** a hidden tab pauses the session. A transport failure or a gap longer than the
  ceiling requires reconnecting. A required reconnect stays visible across a background/foreground
  cycle. Reconnect calls start again, which rotates the token for the same session. A different
  session ID rebinds the scheduler.
- **Session gone:** a `404 unknown_id/recovery_session` ends the surface without a reconnect-start,
  so a watchdog-ended session is never silently replaced.
- **Finish and Stop:** Finish appears once the session is eligible. "Stop early" states that
  nothing is restored. The terminal message shows the Soul before/after or the watchdog
  cancellation, focus moves to the heading, and the host refreshes once.
- **Rejections:** `RECOVERY_REJECTIONS` maps exact recovery-handler pairs. A unit test checks that
  every mapped pair is written by `server/account/api.go`.
- **Toy:** `RecoveryToy.svelte` is decorative only (`aria-hidden`, no animation). Its cells follow a
  client-local seed that is generated on mount and never sent.

`test/soul-recovery-surface-browser.test.ts` runs in all three browsers with a controllable clock
and visibility. It covers axe checks in every state, beats, the hidden pause, the ceiling-gap
pause, a network pause with no beats replayed, reconnect token rotation, finish, the
gone/watchdog/not-ready paths, and reconnect precedence across a background/foreground cycle.
`make verify-client-boundary` now scans the `minigame/` and `soul/` Svelte components with the
same literal, style, import and network rules as the Game UI.

## Tenant boundary

A tenant registers one immutable descriptor: engine/version identity, command/snapshot/result
schema references, shipped modes, a sorted closed rejection taxonomy, and a complete map of frozen
scaling inputs to `power`, `breadth`, or `presentation` destinations. The registry rejects duplicate
engines, duplicate modes/errors, `live_pvp`, unknown destination classes, and incomplete scaling
maps. Creation, play, and terminal-result validation dispatch by the exact frozen
`(engine_ref, engine_version)` pair. If a deployment no longer carries that version, the session
defers unchanged; a newer engine under the same reference cannot execute it under the old label.

Tenant creation and transitions are pure calls. Their only inputs are mode, seed or revision,
canonical snapshot/command JSON, and a defensive copy of the frozen exact-integer scaling map.
They cannot emit economy deltas. A terminal result is limited to a mechanical outcome, sorted
typed integer score facts, and an optional exact-integer rating delta; payout remains platform
owned. Descriptor schema references are backed by tenant validators invoked before and after every
call. Noncanonical or wrong-schema snapshots, undeclared rejection codes, malformed results, and
unknown modes fail closed as rejection or tenant divergence. JSON numbers have one accepted
grammar: exact safe integers only; decimal/exponent aliases are rejected before the JSONB seam.

Every applied play appends its canonical command to `minigame_session_commands` in the same
transaction as the session-head revision. Rows are immutable; they may disappear only with their
parent session's retention cascade. Terminal resolution locks the certified session, replays the
complete ordered command log from genesis through the exact frozen engine version, byte-compares
the pre-terminal snapshot and terminal snapshot/result, then appends the terminal command and
resolves the head in the payout transaction. A same-version engine change that alters an honest
history therefore fails closed before any payout can commit.

The current conformance tenant is test-only. The combat duel adapter will register when its engine
RFC supplies an implemented transition surface; the deferred lane engine is not fabricated by the
platform.

New-schema minigame definitions declare `soul_gate: human_hobby|unrelated`. A bundle containing
Soul requires that schema and the same pinned Soul artifact. The production start boundary resolves
the Founder revision and pinned constants hash; near-zero Soul rejects human-hobby sessions while
unrelated sessions remain available. Historical minigame artifacts keep their prior grammar and
are accepted only in Soul-less bundles.

## Scaling policy grammar

The platform can load and resolve structural scaling policies without enabling a production
minigame catalog. A policy contains one exact-key row per destination:
`destination`, `destination_class`, `source_kind`, `source_ref`, `op`, `operand`, `clamp_min`, and
`clamp_max`. Destination classes are `power`, `breadth`, or `presentation`; source kinds are
`literal`, `tier`, `purchased_generator_count`, `founder_carry_counter`, or
`attended_quality_grade`; operations are `identity`, `add`, `mul`, or `floordiv`.

Resolution reads the source, applies the one declared integer operation, then clamps to the
declared range. Intermediate arithmetic is exact and unbounded; `floordiv` uses mathematical floor
for negative values as well as positive ones. The final values must fit the numeric core's exact
integer domain. Duplicate destinations, unknown keys or source paths, malformed integer literals,
and ranked policies with a `power` destination fail catalog load. The generated formula artifact
publishes this grammar and operation order and fingerprints the loader and resolver source.

No production policy rows or balance values ship in this foundation slice. Those rows join the
minigames epoch artifact only after harness tuning and an owner-approved balance mint.

## Fallback policy grammar

Every enabled minigame must load exactly one fallback arm. `solo` has only its discriminator.
`bot` carries an exact `{policy_id, version}` identity plus `rate_reduction_ppm`;
`npc_partner` carries an exact `{profile_id, version}` identity plus the same reduction field.
Policy/profile IDs use mechanical identifiers, versions are semantic-version identities frozen in
session genesis, and reductions are integer ppm in `[0, 1_000_000]`. Missing, mixed-arm, unknown,
or extra fields fail load. This grammar is published and source-fingerprinted with the scaling
contract.

The offline-quality policy has an exact structural loader but no production balance row. Its
grade curve consists only of `{score_threshold, grade_ppm}` rows in strictly ascending score
order with nondecreasing grades. Scores use the last threshold they meet; values below the first
threshold use the declared neutral floor, which must equal the curve's lowest grade. The outer
row also binds one declared score fact and one tenant-registered automation destination. Unknown
keys, undeclared identities, ambiguous JSON, invalid ppm domains, and noncanonical curves fail
load. Go and TypeScript consume one shared policy/state/threshold fixture and select identical
grades at every boundary.

The replay-owned state shape is `{grade_ppm, last_founder_attended_ms,
decay_remainder_ppm}`. This slice validates that wire and the score-to-grade selection only; the
attended-grid decay transition and production policy literals remain disabled until the Founder
version/artifact activation seam is composed.

The activation seam is now structural: the pinned `minigames` artifact contains every immutable
definition row (engine identity, modes, result facts, scaling, payout, fallback, offline-quality,
rating, and unlock policy), closes sorted minigame-ID and rating-season domains, and its presence
derives Founder save v17 while Company remains v16.
Founder v17 stores exact current rating rows (`elo`, `season_member`, `games_counted`) and exact
offline-quality watermark rows. The artifact and maps are biconditional in both replay runtimes.
The checked fixture exercises the full grammar without minting a production balance artifact.

## Payout policy and conversion kernel

The structural payout row has exactly six keys: `credited_resource_id`, `sends_per_day`,
`per_send_cap`, `conversion_ppm`, `payout_score_fact_id`, and `cap_reason_key`. Loading requires
the resource, typed score fact, and visible copy key to exist in their owning declarations; the
policy cannot create a free-form currency, select an arbitrary result value, or invent cap copy.
Count and cap values stay in the exact-integer domain and conversion is integer ppm.

Payout first selects exactly the declared, nonnegative fact from the certified tenant result;
missing, malformed, or negative facts fail before writes. The pure conversion kernel then applies
the fallback reduction with floor rounding, followed by
the conversion ratio with the prior `conversion_remainder_ppm`, and returns the next modulo
remainder. It uses exact integer intermediates, so the product at the maximum legal score cannot
overflow native arithmetic or change the remainder. The formula artifact publishes and
fingerprints both the loader and this operation order.

Cross-run quota and carry live on one `minigame_faucet_window` row keyed by Founder, minigame, and
Founder-attended day. Under the Founder lock, conversion updates that row's modulo remainder,
applies the per-send cap only while `quota_used < sends_per_day`, and increments quota once for an
admitted send. Converted units beyond either configured limit are returned as forfeited with the
declared cap reason. A new attended day starts a new zeroed row; no wall clock participates.

The window mutation is transaction-owned and intentionally not exported. Its caller owns the
Founder→Company locks and the session claim in the same transaction; the token-owned terminal
session update is the exactly-once authority. Rolling the transaction back removes the window,
payout, rating/quality changes, both log arms, both events, and all save revisions.

The Company replay arm independently recomputes the certified-result hash, selected score, faucet
conversion, configured forfeit, and saturating Decimal credit from the pinned policy. The Founder
arm independently recomputes attended-grid quality decay followed by grade replacement and applies
the engine-certified rating delta with exact checked arithmetic and catalog bounds. Both Go and
TypeScript replay the same shared fixture to byte-identical receipts, event order, and post-state;
the two immutable logs bind the same certified-result hash and are joined by a deferrable relational
source coordinate.

## Persistent tenants

Server Garden is the first persistent (non-session) tenant. It uses only the faucet window, keyed by
faucet owner `server_garden` through `ApplyPersistentFaucetWindowTx`, and the coordinator's
`FounderIdempotency` / `CompanyCanonicalPayload` options. See [Server Garden](minigame-server-garden.md).
