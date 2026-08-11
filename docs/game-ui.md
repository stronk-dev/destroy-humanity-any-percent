# Game UI

The Game UI is the Svelte Phase-A play surface mounted by the production client entrypoint. It
consumes the generated `game_ui_snapshot.v1` projection, decoded lifecycle events, bootstrap and
intent operations, and the ratified Copy/Presentation catalogs. Components do not import transport
or replay internals; `client/src/game-ui/runtime.ts` owns HTTP, WebSocket, and envelope decoding.

## Shipped surfaces

- Vision Slide: silently creates the anonymous account through the idempotent bootstrap
  coordinator and persists credentials before entering play.
- Desk: manual action, resources and visible cap explanations, generator purchases, upgrades,
  local splits, the free Horse Armor shelf, shareware registration/order form, and README.TXT.
- Offer Sheet: authoritative exit type and server-clock-relative expiry with Company-only decline.
- Run End: a payload-isolated component that accepts only the decoded `run_ended` event.
- Settings/System: save status, drain notice, and explicit resync action.

The persistent chrome derives its era only from the authoritative tier (`0` is `era_1995`, `1` is
`era_2000`). RTA uses the snapshot's server-time sample plus monotonic elapsed time. Gate splits and
personal-best timing are local display records only and never feed an intent or leaderboard.
Presence is hidden until the real world-channel count arrives; the UI never invents a visitor.

## Boundaries and verification

`make verify-client-boundary` scans the Game UI components alongside the archived UI primitives.
It rejects transport/replay imports, raw network calls, player-facing text literals, and governed
style literals. `make test-browser` applies the WCAG 2.2 AA axe gate to all five surfaces in
Chromium, Firefox, and WebKit and includes the sixty-second observable performance scenario. The
focused `make test-game-ui-performance` command runs that scenario alone.

The deterministic performance lane feeds 1,200 authoritative snapshot updates representing 60
seconds at 20 Hz through 600 shared formatter windows in a 1280×720 Chromium viewport. The shared
renderer may commit a hot Amount at most 600 times and may not produce a long task over 200 ms.
Four-times CPU throttling and the five-percent dropped-frame allowance remain the manual
mid-range-Android release profile; deterministic CI gates the observable commits and long tasks
as ruled.

## Open contract dependencies

Two screen arms remain fail-closed rather than guessing missing authority or copy:

- offer acceptance needs the Founder revision CAS coordinate, which `game_ui_snapshot.v1` does not
  expose; decline remains usable;
- the complete payout terms and run-end delta list need owner-authored presentation rows for the
  existing decoded terms object.

The exact T0→T2 browser progression also remains tied to the ratified epoch-7 content mint. The
current composed path proves real bootstrap, intent, and snapshot behavior without substituting
fixture mechanics for content that is not live.
