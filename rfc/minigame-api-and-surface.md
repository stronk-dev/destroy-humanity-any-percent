# RFC: Minigame & Recovery API + Surface (the playability seam)

- **Status:** draft — queued for Codex acceptance review. The named successor that TP-C10 (The
  Pitch) and SR-C1/SR-C3 (Soul Recovery) both depend on for HUMAN playability; **every later
  minigame and coordinator activity inherits this boundary.**
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/06 §transport` (authenticated command surface), `design/11` (first-session
  UX — the onboarding dossier's MUST/MUST-NOT list applies to these surfaces)
- **Depends on:** API Foundation (accepted, implementing — this AMENDS its authenticated `/api/v1/`
  surface, additive-only per its A1 policy), Minigame Platform (archived — the internal
  create/play/resolve services this exposes), Soul Foundation (archival-eligible — the SB19/SB24
  coordinator this exposes), UI Foundation (accepted — the component/token substrate the surfaces
  render on), Game-UI Screens (draft — consumes the surface contracts defined here).
- **Planning:** `planning/minigame-api-and-surface/` (once implementing)

## Summary

Two narrow, authenticated, additive `/api/v1/` surface families exposing machinery that is already
implemented, reviewed, and internal-only: **minigame sessions** (create → command → resolve for
platform tenants, The Pitch first) and the **soul-recovery coordinator** (start/reconnect →
progress heartbeat → cancel/resolve, per SB25–SB27). Plus one UI **surface contract** per family —
the component-boundary shape (inputs, callbacks, lifecycle) that Game-UI screens consume — so
"playable by a human" becomes a defined seam instead of an assumption. No new game mechanics, no
new persistence: this RFC is transport + surface only.

## Motivation

Wave-B's first two content RFCs both bounced on the same missing seam: the platform and coordinator
are internal Go services with no authenticated route, and Game-UI explicitly excludes minigame
surfaces. This RFC owns that seam once, as an API-Foundation amendment, so content RFCs never have
to. Out of scope: the public read API (`/api/public/v1/` — untouched), spectation, any WebSocket
streaming for minigames (bounded commands over HTTP per the established intents pattern; a
streaming variant is a declared successor if a real-time tenant ever needs it), and the Game-UI
screens themselves (that RFC consumes these contracts).

## Specification

### MA1 — The minigame session API (authenticated `/api/v1/minigames/`)

Additive endpoints, exact request/response schemas generated per the API Foundation's schema
authority (A2/A5):
- `POST /api/v1/minigames/{minigame_id}/sessions` — create. Founder identity from the
  authenticated session ONLY (no founder field in the body). Validates unlock (incl. the
  `fiscal_unlock` resolver arm, TP-C6), soul_gate, and the platform's session rules. Response: the
  session descriptor + the engine's initial snapshot.
- `POST /api/v1/minigames/sessions/{session_id}/commands` — one engine command per request (the
  tenant's closed command union, e.g. The Pitch's `draft_card | buy_hack | slot_hack | play_hand |
  end_shop`), idempotent by the platform's existing intent-record discipline; response: the
  post-command snapshot + revision. Typed rejections reuse the closed taxonomy verbatim.
- `POST /api/v1/minigames/sessions/{session_id}/resolve` — triggers the certified resolve through
  the shipped composer; response: the terminal receipt (payout, quality, facts). Retry returns the
  stored receipt (the platform's idempotency, exposed not reinvented).
- `GET /api/v1/minigames/sessions/current` — reconnect: the founder's live session descriptor +
  snapshot, or 404-shaped empty (typed, not an error).
Rate limits per the API Foundation's operational middleware (C16/C20 literals); commands
additionally per-session budgeted (catalog data).

### MA2 — The recovery coordinator API (authenticated `/api/v1/soul-recovery/`)

The SB25–SB27 schemas, verbatim — this RFC adds NO fields:
- `POST /api/v1/soul-recovery/start {activity_id}` — start/reconnect (SB25: returns the existing session with a
  ROTATED progress token when one is active). Exact 7-key response.
- `POST /api/v1/soul-recovery/progress {session_id, progress_token}` — the heartbeat. Exact 5-key response;
  per-session rate limit derived from the catalog beat cadence (SR-C6: cadence = ceiling/3; the
  limiter allows modest jitter above it, rejects flooding).
- `POST /api/v1/soul-recovery/cancel {session_id}` and `/resolve {session_id}` — terminal; retry
  returns the stored receipt. (Route shapes per the SR-C10 ruling — flat command routes, the
  intents pattern.)
Typed rejections: `recovery_token`, `exclusive_activity`, `session_expired`,
`idempotency_conflict` — surfaced exactly as ruled, never exposing claim-lease internals.

### MA3 — The surface contracts (what Game-UI screens consume)

Two component-boundary contracts on the UI Foundation's token/component substrate:
- **`minigame_session` surface:** inputs = the session descriptor + snapshot + the tenant's
  command vocabulary; callbacks = command dispatch + resolve + exit-to-host; lifecycle = mount on
  create/reconnect, snapshot-driven re-render (no client simulation — the server snapshot is the
  only truth), terminal rendering from the receipt. Tenant-specific presentation (The Pitch's card
  table) plugs in as a child component keyed by `engine_ref`; the surface owns the chrome, errors,
  and lifecycle so tenants never touch transport.
- **`soul_recovery` surface (SR-C3, restated here as the owning contract):** inputs = session
  receipt + local toy seed; callbacks = the coordinator commands; one heartbeat scheduler
  (visible-only, cadence per SR-C6, missed ceiling = pause, reconnect-then-resume, never replay
  queued beats); cancel affordance; progress display; terminal return. Toy components receive
  presentation data/callbacks only.
Both surfaces meet the UI Foundation's axe-core/keyboard gates; both follow the onboarding
dossier's MUST/MUST-NOT screen rules where applicable.

### MA4 — Authority & privacy invariants

Founder identity exclusively from the authenticated session; no endpoint accepts a founder,
company, or clock coordinate from the client (the established law). The privacy assertion extends:
an integration test enumerates these endpoints' responses against the same no-hidden-info discipline
as the public API test. Nothing here widens replay surfaces — every mutation already flows through
the shipped, reviewed composers.

## Deviations from design

None — this exposes ruled machinery over the established API discipline.

## Acceptance criteria

1. A human-shaped client (integration test using only these endpoints) plays a full Pitch run:
   create → commands → resolve → receipt, with reconnect mid-run; and completes a full recovery
   session: start → heartbeats → resolve, with token rotation on reconnect and the watchdog path.
2. All schemas generated from the single authority; additive-only against `/api/v1/`; typed
   rejections byte-match the closed taxonomy.
3. Rate limiting proven: command flooding and heartbeat flooding rejected without state mutation.
4. The privacy enumeration test passes for every new endpoint.
5. Both surface contracts implemented as components passing the UI Foundation gates, with the
   tenant child-component seam proven by The Pitch's table mounting into `minigame_session`.

## Open questions

- Whether `GET .../sessions/current` should cover minigame AND recovery in one call for the
  reconnect screen (a convenience composite) — decide at acceptance review.
- The streaming successor's trigger condition (first real-time tenant) — declared, not designed.

## Changelog

- 2026-08-07: created (draft) — the playability seam TP-C10/SR-C1/SR-C3 named.
