# RFC: Soul Recovery Activities (the cozy content — touch-grass v1)

- **Status:** draft — Wave-B content; queued for Codex acceptance review. The **owner-content
  mint** the Soul verdict requires before production `recovery_activities` rows may exist.
- **Author:** Marco (drafted by Claude)
- **Created:** 2026-08-07
- **Design refs:** `design/02 §8` (recovery: deliberate, opportunity-costed, produces nothing),
  `design/03 §5` (the cozy category is the SOLE Soul-recovery source — board games pay but never
  restore; ruled 2026-08-06)
- **Depends on:** Soul Foundation (SB1–SB27 — the complete session/heartbeat/suppression substrate;
  archival-eligible). **This RFC adds ONLY catalog rows + client presentation + copy. Zero server
  mechanics** — the coordinator, watchdog, suppression, and replay already exist and are reviewed.
- **Planning:** `planning/soul-recovery-activities/` (once implementing)

## Summary

The first three production touch-grass activities — **`defrag`** (a zen tile-layer: watch the disk
come to order), **`server_room`** (a no-goal toybox: arrange the racks, water nothing, help no one),
**`repot`** (a short daily ritual: repot the server-room plants) — as `recovery_activities` catalog
rows on the shipped Soul substrate. Each costs deliberate attended time under production
suppression and **produces nothing** except Soul recovery: no resource, no score, no achievement,
no record of how well you did, because there is no "well." The tooltip says so, sincerely — the one
place the game's curtain-pulling voice goes quiet.

## Motivation

Soul's drain is live machinery awaiting its counterweight: without production activity rows, the
training-data ending is reachable but recovery is fixture-only. The research finding this rests on:
**"produces nothing" and "cheap to serve" are the same property** — a zero-reward toy has nothing
to cheat, so the server needs only the session lifecycle (which exists), and the entire interactive
content is client-side presentation. This is the cheapest content RFC in the game and it completes
the game's emotional core loop (drain → grey pet → touch grass → the pet knows you again).

Out of scope: additional activities (this union is open; rows are data), any activity that grants
any resource (forbidden by the recovery covenant — permanently), and audio (the audio research
commission owns the soundscape later).

## Specification

### SR1 — The three catalog rows

Exact `recovery_activities` rows per the SB10/SB17 grammar
(`{activity_id, duration_attended_ms, recovery_amount, reason_key}` + copy keys):
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

The rows enter the pinned Soul artifact (the SB17 grammar allows production rows; the fixture-only
containment lifts exactly here, per the Soul closing verdict: **after Soul's archival + this
content**). Activation is the standard epoch mint under the activation-boundary law. UI surfacing
(where the activities appear) is the Game-UI screens RFC's consumer seam; this RFC ships the rows,
the toys, and the copy.

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
4. The toys render and resume-after-reconnect; presence-without-interaction recovers identically
   (no interaction requirement).
5. The disclosure tooltip present for all three, in the sincere register (copy-review flag, not
   CI-provable).

## Open questions

- Whether `repot` seeds a tiny persistent visual (the plant grows across sessions — pure cosmetic,
  client-preference storage, never save state) — nice, cheap, and honest; decide at implementation.
- Ambient audio — deferred to the audio commission; the toys ship silent-but-pleasant first.

## Changelog

- 2026-08-07: created (draft) — Wave-B; the owner-content mint for Soul recovery.
