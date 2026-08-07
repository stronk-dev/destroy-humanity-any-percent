# Soul Recovery Activities

Soul Recovery ships three fixture-first, zero-output recovery activities on the archived Soul
session machinery: `defrag`, `repot`, and `server_room`. Each activity advances only through
attended presence, suppresses Company production while active, and grants only its catalog-declared
Soul recovery at terminal resolution. No score, resource, achievement, or toy result enters the
authoritative transition.

## Catalog and activation

Each strict activity row declares its mechanical ID, attended duration, recovery amount, toy kind,
reason key, and title/description/disclosure copy keys. Go and TypeScript loaders accept the same
closed grammar, and all copy resolves through the copy pipeline. The implementation remains
fixture-first: no production epoch pins these rows until the owner-gated First Content Epoch mints
the dependency-complete artifact bundle.

The provisional rows are:

- `repot`: 300,000 attended milliseconds, 5 Soul;
- `defrag`: 900,000 attended milliseconds, 12 Soul;
- `server_room`: 2,700,000 attended milliseconds, 30 Soul.

The disclosure is deliberately sincere: the activity produces nothing besides Soul recovery.
Interactive toys and final disclosure rendering remain owned by the UI successor and do not affect
authoritative progress.

## Authenticated coordinator

The public coordinator exposes authenticated start/reconnect, progress, cancel, and resolve
commands. Founder and Company identity come from the session, never request fields. Start creates a
server-owned session; repeating start reconnects and rotates the opaque progress token. Only the
latest token is accepted. Terminal commands are idempotent, and watchdog expiry uses the same
zero-grant cancellation path.

Progress is session-rate-limited with burst 6 and refill interval
`recovery_beat_ceiling_ms / 6`. A limiter rejection neither consumes nor rotates the progress
token. Deterministic gameplay rejections use the closed API error envelope; transient failures
remain server errors.

## Heartbeat lifecycle

The framework-neutral client scheduler beats only while the recovery surface is active and
visible. Its interval is exactly one third of the catalog ceiling. An elapsed gap greater than
`3 * beat_interval_ms`, whether caused by a hidden tab, sleep, clock regression, or transport
failure, enters a reconnect-required pause. No further beat is sent until reconnect rotates the
token. An under-ceiling hidden interval may resume immediately, and queued beats are never replayed.

Only one progress request may be in flight. Stopping with `cancelled`, `resolved`, or `watchdog`
emits one terminal callback and cancels future dispatch. Replay verifies accumulated attended
totals at terminal resolution; heartbeat cadence itself never becomes replay input.

## Review and carried work

The designated review consumed implementation set `{4973c8e, ab9d15e, 3cfc0e6}` plus docs-tier
`{f04c2f3, d1cd39c}` and approved archival in verdict `5754901`. The production mint and UI toy/
disclosure acceptance criteria remain explicit successor work; neither is claimed by this archive.
