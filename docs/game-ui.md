# Game UI

The Game UI is the Svelte Phase-A play surface mounted by the production client entrypoint. It
consumes the generated `game_ui_snapshot.v2` projection, decoded lifecycle events, bootstrap and
intent operations, and the ratified Copy/Presentation catalogs. Components do not import transport
or replay internals; `client/src/game-ui/runtime.ts` owns HTTP, WebSocket, and envelope decoding.

## Shipped surfaces

- Vision Slide: silently creates the anonymous account through the idempotent bootstrap
  coordinator and persists credentials before entering play.
- Desk: manual action, resources and visible cap explanations, generator purchases, upgrades,
  local splits, the free Horse Armor shelf, shareware registration/order form, and README.TXT.
- Offer Sheet: authoritative exit type, complete payout terms, server-clock-relative expiry,
  Company-only decline, and Founder-CAS-guarded acceptance.
- Run End: a payload-isolated component that accepts only the decoded `run_ended` event.
- Settings/System: save status, drain notice, and explicit resync action.

The persistent chrome derives its era only from the authoritative tier (`0` is `era_1995`, `1` is
`era_2000`). RTA uses the snapshot's server-time sample plus monotonic elapsed time. Gate splits and
personal-best timing are local display records only and never feed an intent or leaderboard.
Presence is hidden until the real world-channel count arrives; the UI never invents a visitor.
Presentation schema v3 owns the literal `$0.00` and pre-naming `Founder` constants, so copy
placeholders never substitute a formatted zero or an unrelated company label. Missing constants
throw. Payout labels and any shipped network-slot titles also resolve only through that catalog;
unknown future slot IDs are withheld rather than rendered mechanically.

Live sync requires snapshot v2 and its positive `founder_revision`. Encrypted bootstrap receipts
remain replayable under the schema version they were minted with: a stored v1 snapshot legally
lacks that coordinate, and offer acceptance stays disabled until the next live v2 sync supplies it.
The runtime persists positioned player/world subscriptions, recovers missed publications after a
drop, and falls back to the same authenticated live snapshot operation on a revision gap, expired
history, queue overflow, or invalid frame. Recovery snapshots are delivered into the existing
`bindSnapshot` path; there is no parallel API client or snapshot schema. Drain delays reconnect by
the server-advertised bound. Auth-expired and replaced sockets surface offline instead of inventing
Account token-rotation or multi-tab arbitration behavior.

## Boundaries and verification

`make verify-client-boundary` scans the Game UI components alongside the archived UI primitives.
It rejects transport/replay imports, raw network calls, player-facing text literals, and governed
style literals. `make test-browser` applies the WCAG 2.2 AA axe gate to all five surfaces in
Chromium, Firefox, and WebKit and includes the sixty-second observable performance scenario. The
focused `make test-game-ui-performance` command runs that scenario alone.
`make test-game-ui-composed` additionally drives Chromium through the real Vite proxy, composed
gameserver, Postgres bootstrap transaction, authenticated live snapshot-v2 route, and Centrifuge
world subscription; its schema/revision and visible visitor-counter assertions prove the production
HTTP synchronization and WebSocket handshake completed. The composed witness then closes the
production browser socket, commits an intent while disconnected, and requires the reconnect to send
the exact persisted player epoch/offset and advance it by replaying the missed receipt. This current
composed test does not yet prove the three server-authorized transition/continuation controls; that
narrowed GU-C25–GU-C28 browser proof remains the Game UI closeout item.

The deterministic performance lane runs in an isolated Chromium process after the functional
Chromium/Firefox/WebKit matrix, then feeds 1,200 authoritative snapshot updates representing 60
seconds at 20 Hz through 600 shared formatter windows in a 1280×720 viewport. Isolation keeps
concurrent browser engines from becoming part of the measurement. The shared renderer may commit
a hot Amount at most 600 times and may not produce a long task over 200 ms.
Four-times CPU throttling and the five-percent dropped-frame allowance remain the manual
mid-range-Android release profile; deterministic CI gates the observable commits and long tasks
as ruled.

## Live-content boundary

Epochs 7 and 8 provide the live T0–T1 economy and first-hour curriculum. The content contract's
pinned-seed proof drives that sequence through the real composed gameserver and Postgres. The Game
UI does not duplicate that two-hour/full-script proof in Chromium. Its remaining browser gate is
deliberately narrower: server-authorized Gate and Wind Down controls, both terminal surfaces, and
Run-End continuation into the already-created next run, with discriminating disconnect and
suppression failures.
