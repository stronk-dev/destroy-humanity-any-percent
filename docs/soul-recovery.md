# Soul Recovery Activities

Soul Recovery ships three minted (epoch 6), zero-output recovery activities on the archived Soul
session machinery: `defrag`, `repot`, and `server_room`. Each activity advances only through
attended presence, suppresses Company production while active, and grants only its catalog-declared
Soul recovery at terminal resolution. No score, resource, achievement, or toy result enters the
authoritative transition.

## Catalog and activation

Each strict activity row declares its mechanical ID, attended duration, recovery amount, toy kind,
reason key, and title/description/disclosure copy keys. Go and TypeScript loaders accept the same
closed grammar, and all copy resolves through the copy pipeline. The rows are LIVE: the epoch-6
mint (First Content, 2026-08-10) pins them via `balance/soul/first-content.json`; retunes ride
the normal epoch lane.

The minted rows are:

- `repot`: 300,000 attended milliseconds, 5 Soul;
- `defrag`: 900,000 attended milliseconds, 12 Soul;
- `server_room`: 2,700,000 attended milliseconds, 30 Soul.

The disclosure is deliberately sincere: the activity produces nothing besides Soul recovery.
The UI successor implements decorative toys and disclosure rendering in the Recovery surface;
neither affects authoritative progress. See [Minigame platform](minigame-platform.md#client-surface)
for current implementation and bounded real-service player verification.

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
visible. Its interval is contracted (SR-C6) to be one third of the catalog ceiling; the shipped
scheduler accepts any interval ≥ 1 ms; `loadSoulRecoveryContent` derives the actual surface's
interval from the pinned catalog ceiling. An elapsed gap greater than
`3 * beat_interval_ms`, whether caused by a hidden tab, sleep, clock regression, or transport
failure, enters a reconnect-required pause. No further beat is sent until reconnect rotates the
token. An under-ceiling hidden interval may resume immediately, and queued beats are never replayed.

Reconnect can return a replacement/existing session with another activity. The surface binds
its returned activity ID to the mounted catalog for title, toy and disclosure, retaining the
server's duration/progress and replacement session/token. The requested old activity is not
display authority. Unknown returned activities show the existing error and dispatch no new
heartbeat, matching Begin's existing owner lookup; no automatic replacement/cancel is added.

The Recovery surface declares zero Tab indices for Begin, Reconnect, Finish, Stop and Back,
preserving document-order keyboard traversal even when a browser omits implicit button stops.
Disabled controls retain their native disabled behavior. Terminal receipts restore the heading
when focus belonged to Recovery or was unclaimed, unless a newer surviving control receives
focus before the DOM update finishes. A player who tabs to another host control while Finish/
Stop is pending keeps that focus; terminal text and the once-only host notification still arrive.
Tab from the ordinarily restored heading reaches Back. Mounted Chromium/WebKit checks exercise actual Tab/
Shift-Tab and Enter/Space with controlled coordinator replies, not direct button focus or clicks.
They do not establish real assistive-technology or whole-player accessibility acceptance.

The built-client/real-server Recovery journey also traverses from the current focus using
Tab before Enter, rather than assigning focus to an action button. Its test-side negative
removes Begin from Tab order and requires refusal with no Recovery request or new stored
session, then restores the original DOM attribute before ordinary play. The same driver
serves bounded cancellation/reconnect in normal composition and the explicit five-minute
Finish population. This Chromium integration is distinct from the controlled multi-engine
checks above; neither substitutes for assistive-technology acceptance.

Only one progress request may be in flight. Stopping with `cancelled`, `resolved`, or `watchdog`
emits one terminal callback and cancels future dispatch. Replay verifies accumulated attended
totals at terminal resolution; heartbeat cadence itself never becomes replay input.

## Review and carried work

The designated review consumed implementation set `{4973c8e, ab9d15e, 3cfc0e6}` plus docs-tier
`{f04c2f3, d1cd39c}` and approved archival in verdict `5754901`. At that archive coordinate, the
production mint and UI toy/disclosure acceptance criteria were explicit successor work. First
Content subsequently minted the three activity rows in epoch 6; the UI successor subsequently
implemented the toys/disclosure. Its current player verification and designated review remain
separate from this archive. Neither successor changes the archive's fixture-first implementation
and review boundary.
