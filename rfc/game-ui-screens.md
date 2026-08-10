# RFC: Game UI Screens (the play surfaces)

- **Status:** draft
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-03
- **Design refs:** `design/11-ux-writing.md` (first-session narrative, voice), `design/08 §speedrun` (run-title bar, splits panel, PB/gold deltas), `design/06` (DOM-first, Svelte 5 runes, `$derived` bound to visible tab only), client-shell docs (reconciliation, `reason_key` caps, activity_ppm never-frozen)
- **Depends on:** Client Shell (implemented), Transport (implemented), **UI Foundation (accepted —
  C1–C11 ruled; implementing)** — screens are built FROM its token matrix and components and
  inherit its gates, T0–T1 Content (draft — ships together; screens without content are furniture)
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
   metric; showing it live would teach clicking-for-the-timer). **First-run variant per the
   design/11 §1b adoptions (binding):** run 1 shows `PB: —` and NO world-record line (the WR line
   appears only once a PB exists to compare against).
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
7. **First-session elements (design/11 §1b adoptions, all binding):** the **Vision Slide**
   (REQUIRED v1.0 — the pre-run framing screen at `[BEGIN ATTEMPT]`; account creation is silent
   server-anonymous, so this screen has NO signup form); the run-1 presence element is the
   **visitor counter ONLY** (no feed rendering — the feed unlocks at the scripted first failure,
   which is T0–T1 content, not a Phase-A screen); the T0 era skin carries the §1b satire beats
   (the arcade `ORDER NOW $0.00` element, the `UNREGISTERED` count-up nag, the README.TXT skin)
   as copy/data through the pipeline, never as component literals.

### U2 — Contracts

- Every screen consumes ONLY: shell state, decoded wire envelopes, and event payloads — zero
  new server surfaces, zero client-computed authority. A screen needing a field the wire lacks
  is a DESIGN-GAP escalation, never a client-side derivation of authoritative data.
- DOM-first: no canvas in Phase A. `$derived` subscriptions bound per-visible-tab (the
  tech-stack rule); number formatting throttled ~10 Hz, rendered through the UI Foundation's
  pinned notations binding (`@antimatter-dimensions/notations`, Standard, 3 significant digits) —
  no second formatter.
- All visual values from the UI Foundation token matrix (era_1995 for Phase A — including its
  deliberate all-zero motion tokens); no raw colors/spacing in components.
- All copy through the copy system (T0–T1 RFC's pipeline); zero literals in components.
- Accessibility: the UI Foundation's C11 gate applies to every Phase-A screen (axe-core WCAG 2.2
  AA, zero serious/critical) IN ADDITION to keyboard-completable first hour and OS reduced-motion
  — the era_1995 zero-motion tokens satisfy reduced-motion by construction.

### U3 — What Phase A does NOT ship

Boards UI, commons/guild panels (Commons Onboarding RFC owns those), pet surfaces, the
`minigame_session` and `soul_recovery` surfaces (**Minigame & Recovery API + Surface** owns those
contracts — MA3; their screens mount as later additions on this RFC's shell), cosmetic shop
beyond the static Horse Armor shelf, feed/presence rendering beyond the §1b visitor counter. Each
is a later screen-set RFC on the same U2 contract.

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
- 2026-08-07: pre-acceptance hardening — UI Foundation (C1–C11) added as a binding dependency
  (token matrix, pinned notations, C11 axe gate); design/11 §1b first-session adoptions bound
  (Vision Slide, `PB: —` first-run title bar, visitor-counter-only presence, T0 satire beats);
  U3 reconciled with the Minigame & Recovery API + Surface RFC's MA3 surface ownership.

## Codex acceptance-review blockers (2026-08-10 — GU-C1–GU-C8)

UI Foundation now supplies the governed primitives, but the screen RFC does not yet define the
data and lifecycle contracts needed to connect those primitives to the composed server. These are
product/wire decisions; implementation remains blocked rather than filling them in locally.

### GU-C1 — The shipped shell snapshot cannot represent the Desk

`AuthoritativeSnapshot` contains resources, discrete facts, and progress coordinates only. It has
no generator purchased/provisioned counts or contribution rows, upgrade ownership/eligibility,
manual-token state, tier, run sequence, Founder exit count, or server time sample. U1 requires all
of those, and the production entry currently mounts the UI fixture without composing a stream.

**Proposed contract:** enumerate one exact generated `game_ui_snapshot.v1` DTO owned by the server
projection and decoded by the shell. It carries the revision/hash/time envelope plus closed `run`,
`manual_action`, `generators`, `upgrades`, `resources`, `facts`, and `progress` fields; every row is
mechanical-ID sorted. Buttons submit existing intents and treat affordability as advisory display
only—the receipt remains authoritative. Pin which HTTP/socket snapshot operation supplies it and
how it composes into `ShellRuntime`; do not let screens read save state or the economy catalog
directly.

### GU-C2 — The screen/surface registry is not a literal catalog

U1 names seven conceptual areas, but no `surface_id`, `mount_id`, unlock row, default surface, or
authoritative fact manifest exists. UI Foundation rejects unknown facts and duplicate mounts, so
the prose cannot load through it.

**Proposed contract:** add the complete Phase-A surface document with exact rows and default/fallback
selection. State explicitly which U1 items are surfaces versus components inside a surface, and
name every `fact_equals` ID in the snapshot manifest. Run-end/offer navigation driven by events
must have one deterministic precedence rule over player-selected tabs.

### GU-C3 — Required event payloads are opaque or incomplete at the client boundary

The transport validates event envelopes, not the per-kind payload shapes the screens require.
The RFC assumes exact gate times, complete `run_ended` curriculum/delta data, offer terms and
expiry, while no generated TypeScript discriminated union owns those bytes.

**Proposed contract:** enumerate and generate the exact v1 payloads for `gate_crossed`,
`run_ended`, `exit_offer_spawned`/resolved, `server_restarting`, and `resync_required`, including
schema-version behavior and cursor effect. The run-end component receives only the decoded
`run_ended` object in its type; a compile-time fixture proves no snapshot parameter is available.

### GU-C4 — The timer, PB, and split persistence model is under-specified

There is no server-clock-offset sample on the wire and no local PB/split storage schema. Browser
wall time alone cannot implement the stated RTA law across reconnect/sleep, while an unspecified
client store makes reset, imported runs, and run identity ambiguous.

**Proposed contract:** bind RTA to a frozen server sample `{server_now_ms, run_started_at_ms}` plus
monotonic elapsed time, resampled on every authoritative snapshot and snapped on terminal data.
Define a versioned local-only record keyed by Founder/run/category with exact PB and split fields,
write timing, corruption fallback, and deletion semantics. It must never feed an intent or board
submission.

### GU-C5 — First-session/bootstrap ownership has no executable transition

The Vision Slide says `[BEGIN ATTEMPT]` silently creates an anonymous account, but no command,
success receipt, retry/idempotency rule, or failure screen is named. The current UI fixture has no
account/bootstrap composition.

**Proposed contract:** bind the CTA to the existing account/bootstrap operation by exact generated
API name and response, or add the missing operation in the API Foundation. Specify idempotent retry,
credential persistence before navigation, and the authoritative fact that selects Vision versus
Desk. No local click may enter play before bootstrap succeeds.

### GU-C6 — Screen copy and presentation bindings are absent

Copy Pipeline has no keys for the Vision Slide, navigation, generator titles/descriptions,
settings, drain/resync story beats, Horse Armor, or the named T0 satire elements. Generator rows
also have no presentation key—the carried PT-C4 debt. The UI boundary correctly rejects component
text literals, so implementation has nothing legal to render.

**Proposed contract:** T0–T1 owns one byte-sorted screen-copy/presentation manifest binding every
surface, action, generator, upgrade, system state, and first-session element to registered Copy
keys. Game UI consumes only that manifest. Supply literal launch text before implementation; no
mechanical-ID title derivation.

### GU-C7 — Era selection contradicts the two-era first hour

U2 says `era_1995 for Phase A`, while T0–T1 and UI Foundation explicitly map T0 to `era_1995` and
T1 to `era_2000`. A fixed era would make the UI fail its own core satire contract.

**Proposed contract:** derive era exclusively from the authoritative tier fact using the closed
mapping `{0:era_1995, 1:era_2000}`; later tiers fail closed until their theme artifacts exist.
The switch changes only the UI root tokens/attribute and preserves active surface/focus where the
element still exists.

### GU-C8 — The performance acceptance criterion has no executable budget

“Reference low-end profile” and “without frame drops” name neither hardware emulation nor a
number, so two implementations can claim opposite results.

**Proposed contract:** define a repository fixture with exact viewport, CPU-throttle factor,
duration, 20-Hz input count, maximum 10-Hz formatted commits, long-task ceiling, and dropped-frame
allowance. If deterministic browser timing is not reliable in CI, gate observable update counts
and long tasks in CI and record the hardware profile as a manual release check.
