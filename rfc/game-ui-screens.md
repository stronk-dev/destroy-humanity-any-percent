# RFC: Game UI Screens (the play surfaces)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/11-ux-writing.md` (first-session narrative, voice), `design/08 §speedrun` (run-title bar, splits panel, PB/gold deltas), `design/06` (DOM-first, Svelte 5 runes, `$derived` bound to visible tab only), client-shell docs (reconciliation, `reason_key` caps, activity_ppm never-frozen)
- **Depends on:** Client Shell (implemented), Transport (implemented), T0–T1 Content (draft — ships together; screens without content are furniture)
- **Planning:** `planning/game-ui-screens/` (once implementing)

## Summary

The shell (sim loop, prediction, reconciliation, wire decoding) is implemented and reviewed; no
player can see it. This RFC specifies the Phase-A screen set — the minimum surfaces for the T0–T1
first hour — as components over the shell's existing state, with the same closed-contract
discipline as every other boundary.

## Specification

### U1 — The screen set (closed, Phase A)

1. **The Desk (main play surface):** manual action button (token bucket visible as the
   in-fiction "energy" it already is), generator list (owned count, rate contribution, buy 1/max
   with the existing afford-fast-path), upgrade shelf, balance header using
   `@antimatter-dimensions/notations`. Capped resources render the shell's `reason_key`
   explanation (never a frozen number — already law).
2. **The Run-Title Bar (always visible):** `COMPANY — Any% — RTA h:mm:ss — Tier N` — RTA derived
   client-side from `run_started_at` vs server clock offset (display only; the server's timer
   facts stay authoritative). Attended time appears post-run, never live (it's an intent-cadence
   metric; showing it live would teach clicking-for-the-timer).
3. **The Splits Panel (collapsible):** one row per crossed gate with time; PB/gold deltas begin
   as local-only (client-persisted) until boards ship UI — the panel renders from `gate_crossed`
   events already in the stream.
4. **The Wind-Down / Run-End screen:** renders EXCLUSIVELY from the `run_ended` event payload
   (AC7's guarantee, now consumed); the scripted-first variant carries its curriculum copy;
   "run 2 opens with" delta list from the D6 assembly facts.
5. **The Offer Sheet:** exit-offer terms from the offer event's terms object; accept/decline
   through the existing intents; countdown from `expires_at`.
6. **Settings/system:** save status, session, drain-notice rendering (`server_restarting` →
   diegetic "scheduled maintenance"), the `resync_required` full-sync flow surfaced as a story
   beat not an error (already ruled in shell D2 — this screen is where it lands).

### U2 — Contracts

- Every screen consumes ONLY: shell state, decoded wire envelopes, and event payloads — zero
  new server surfaces, zero client-computed authority. A screen needing a field the wire lacks
  is a DESIGN-GAP escalation, never a client-side derivation of authoritative data.
- DOM-first: no canvas in Phase A. `$derived` subscriptions bound per-visible-tab (the
  tech-stack rule); number formatting throttled ~10 Hz.
- All copy through the copy system (T0–T1 RFC's pipeline); zero literals in components.
- Accessibility floor: keyboard-completable first hour; reduced-motion honors the OS setting.

### U3 — What Phase A does NOT ship

Boards UI, commons/guild panels (Commons Onboarding RFC owns those), pet/minigame surfaces,
cosmetic shop beyond the static Horse Armor shelf, feed/presence rendering. Each is a later
screen-set RFC on the same U2 contract.

## Acceptance criteria

1. First-session browser test: the full T2 script (T0–T1 RFC) completed through the real UI
   against the composed gameserver — the first end-to-end human-path test in the repo.
2. Wire-only proof: components compile against the decoded-envelope types with no imports from
   transport internals; a lint boundary (the combat-gate pattern) enforces it.
3. Run-end screen renders byte-only from a `run_ended` fixture (no shell state read) — AC7
   consumed as designed.
4. Cap explanation, drain notice, and resync story-beat each have a browser test.
5. Performance: 20 Hz sim + 10 Hz format on the reference low-end profile without frame drops
   (budget in the scenario file).

## Changelog

- 2026-08-03: created (draft) — Phase A screens; ships with T0–T1 content.
- 2026-08-06: non-normative reference cleanup for publication; no spec change.
