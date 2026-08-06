# RFC: Commons Onboarding & Governance

- **Status:** draft
- **Author:** Codex
- **Created:** 2026-07-29
- **Design refs:** `design/05 §3, §5` (guild participation, Commons front door and governance), `design/10 §1` (Open Source participation)
- **Depends on:** Commons Compact (server foundation), Client Shell & Sim Loop (implemented),
  account/session bootstrap, WebSocket Transport, faction/incorporation model, guild model
- **Parent / amends:** `archive/commons-compact.md`
- **Supersedes / superseded by:** —
- **Planning:** `planning/commons-onboarding-and-governance/` (once implementing)

## Summary

Complete the player-facing half of the Mutual Aid Compact after its server foundation: expose the
incorporation choice every run, bind Open Source auto-membership, render the cohort and current-standing
surfaces, connect guild participation, and implement the monthly direction-only tithe vote. This split
prevents the server RFC from improvising incorporation, faction, guild, or client contracts that do not
yet exist.

## Specification

- Every run's incorporation contract renders a plain `sign` / `decline` Compact line item. Declining
  leaves a persistent non-signatory affordance; signing sends the authoritative `sign_compact` intent.
- Once the faction/incorporation RFC defines Open Source identity, its incorporation path signs the same
  Compact automatically at the catalog-declared faction tithe. It does not create a private pool.
- The cohort panel consumes server snapshots/events and shows named neighbors, current Health/Capacity,
  labeled NPC fallback, co-op entry points, and current standing. It never materializes a permanent moral
  badge from historical standing. Its formula view renders every Commons formula and the exact
  Enclosure source-weight table from the shipped generated artifact.
- The monthly cohort ballot chooses direction within the server-owned tithe band; it does not accept
  implementation proposals or arbitrary numeric values. The authoritative election window, eligibility,
  tie-break, result event, and effective-boundary schemas remain a `DESIGN-GAP` until account, faction,
  guild, and transport identities are specified.
- Guild participation contributes to the same Commons and supplies the guild Health term. Guildless
  members continue using the already-shipped cohort substitution. Small-guild mercy changes access or
  progress, never the population-normalized Health formula.

## Deviations from design

- None. This RFC is a boundary split from the implemented server foundation, not a mechanics change.

## Acceptance criteria

1. Browser tests prove the Compact line item appears on every incorporation path; decline remains visible,
   sign uses the server intent, and Open Source uses the same pool with its catalog tithe.
2. The cohort panel accurately renders current server snapshots, named neighbors, NPC labels, and ambient
   collapse/recovery/recruitment events without inventing client state; its formula view matches the
   generated artifact exactly.
3. A closed, versioned ballot contract proves one eligible vote per Founder, deterministic resolution,
   bounded results, an auditable result event, and application at the declared accrual boundary.
4. Guild, cohort, and guildless fallback fixtures produce the server-published Health inputs; small-guild
   mercy cannot change the population-invariance calculation.
5. Leaving remains available from every member surface and visibly explains Solidarity reset before the
   authoritative intent is sent.

## Open questions

- `DESIGN-GAP`: ballot identity/eligibility and tie policy await account/session and guild ownership.
- `DESIGN-GAP`: exact cohort-panel interaction design awaits the Client Shell RFC's implementation.
- The faction/incorporation RFC owns how Open Source identity is selected and persisted; this RFC only
  binds that result to the existing Compact intent.

## DESIGN-GAPs blocking acceptance (Codex review, 2026-07-29)

**Status ruling (2026-08-03): UNBLOCKED — every owner contract now exists.** Original 2026-07-29 ruling preserved below for the record; the six blockers are answered in the new contracts section. Blocker #1 is
one-third answered (`account-and-session-bootstrap.md` now owns identity; transport's contracts are
answered T1–T6). The remaining owners — **faction/incorporation model** and **guild model** — are not
drafted, and blockers #2–#5 all hang off them. This RFC stays out of the implementable queue until
those two RFCs exist; nothing here may be improvised around their absence (the reason this split
exists at all).

1. **Missing owner contracts:** ~~account/session~~ (drafted 2026-07-29), faction/incorporation, and
   guild models are not drafted. This RFC cannot define membership-gated UI or governance
   identity until those boundaries exist.
2. **Incorporation contract:** define its state, every Phase-0 path, sign/decline intent timing,
   persistence across a run, Open Source binding, and catalog tithe values. “Every path” has no closed
   set to test today.
3. **Ballot protocol:** replace the acknowledged `DESIGN-GAP` with exact election-window, proposal,
   eligibility snapshot, vote, tie-break, result, audit-event, and effective-accrual-boundary schemas.
   Define “monthly” in server time and behavior across restarts/epoch changes.
4. **Snapshot/event surface:** enumerate the transport payloads and revisions consumed by the cohort
   panel, including recovery ordering for collapse/recovery/recruitment news and current standing.
5. **Guild contribution boundary:** define guild ownership/membership history, the Health input
   snapshot, tithe band, leave/join concurrency, and small-guild mercy's permitted non-formula effects.
6. **Generated formula UI:** identify the exact generated artifact fields exposed to the client and
   add a source-fingerprint/parity test; “renders every formula” must not create a second hand-written
   formula authority.

## Executable contracts (answering the six blockers, 2026-08-03 — all owners now implemented)

1. **Owner contracts (blocker #1): closed.** Account/session (implemented), faction/incorporation
   (implemented — F2's compact binding + F2a's tithe-raise), guild model (implemented — G3's
   Health term, GD3-1's evaluation-based activity), transport (implemented — wire v2, event
   relay). This RFC now binds to real seams only.
2. **Incorporation contract (blocker #2):** the Compact line item renders on the incorporation
   sheet for every faction; `sign`/`decline` at any time from Tier 0 (signing = the existing
   `sign_compact` intent; the incorporation-time render is UI convenience, not a separate path);
   Open Source auto-sign and the existing-member tithe-raise are the implemented F2/F2a paths —
   fixtures consume them, nothing new. Decline persists as a visible affordance (client state,
   re-render each run).
3. **Ballot protocol (blocker #3), executable:** monthly in server time = the FIRST accrual
   boundary on or after 00:00 UTC on the 1st opens a 72-hour window (window identity =
   `(cohort_id, year, month)`, restart-safe because it derives from the calendar, not a timer).
   Ballot = `cast_tithe_ballot {window_id, direction}` intent, `direction ∈ {lower, hold,
   raise}` (direction-only — the design's law), one per Founder per window (idempotency by
   window in the intent record), eligibility = compact member at cast time. Resolution at
   window close: plurality; tie → `hold`; result event `tithe_ballot_resolved {window_id,
   tally, direction, new_tithe_ppm}` with the band-clamped step (`ballot_step_ppm` catalog
   literal, 10_000) applying at the next accrual boundary after the event. Epoch changes
   mid-window: the window completes under its opening epoch's band (windows are short;
   the pin rule is the run rule).
4. **Snapshot/event surface (blocker #4):** the cohort panel consumes exactly: `cohort` scope
   snapshots (existing schema), `compact_sampled`/`compact_left`/collapse/recovery/recruitment
   events (existing kinds), and the generated formula artifact — enumerated, closed, nothing
   new; recovery ordering is the wire-v2 cursor contract (no special case).
5. **Guild contribution boundary (blocker #5):** implemented (G3 + GD3-1 + the membership-period
   clearing rules); this RFC renders it, never recomputes.
6. **Formula UI (blocker #6):** the panel renders the SHIPPED generated artifact verbatim with
   its source fingerprint displayed; a parity test asserts the rendered bytes equal the artifact
   (no second authority — the blocker's own demand, now testable).

Client scope: the cohort panel, incorporation sheet, and ballot card join the Game-UI RFC's
second screen set (same U2 wire-only contract).

## Changelog

- 2026-07-29: split from Commons Compact during implementation review so missing client/governance
  dependencies remain explicit rather than improvised or falsely marked shipped.
- 2026-07-29: updated implemented dependencies; Codex acceptance review confirmed six blocking
  contracts, including the RFC's existing ballot gap.
- 2026-08-03: UNBLOCKED — all six blockers answered against the implemented owner contracts; calendar-derived ballot windows, direction-only voting, band-clamped steps.
- 2026-07-29: ruled blocked-on-owner (faction/incorporation + guild RFCs); account/session and
  transport dependencies now answered.
