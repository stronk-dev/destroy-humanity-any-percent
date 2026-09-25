# Garage Player Surfaces implementation plan

RFC: `rfc/garage-player-surfaces.md` (accepted 2026-09-25; owner decisions at recommended defaults).
Scope of this plan: release-manifest Batch A slices over already-live backends only. The pet (GS4),
cosmetic, Clout, T2-era and ticker slices are out of scope until their producers land.

Constraint for this lane: no kernel-guarded path (`kernel/affecting-paths.json`) is edited. A slice
that would need one is recorded as a blocker in `log.md` instead.

- [x] GS0.2 runtime: `intent()` returns the typed outcome; non-2xx throws `GameUIRequestError`;
  `act()` scopes `expected_revision` (Company vs Founder); rejected intents render their reason.
- [x] GS0.1 projection: Game UI snapshot v4 (`features` arms + `generators[].provision_cap` +
  feature facts), registered schema, generated types, client parser; live sync requires v4.
- [ ] GS3 Meters arm + surface.
- [ ] GS2 Achievements arm + surface.
- [ ] GS1 Fiscal arm + surface.
- [ ] GS7 minigames availability arm wired into the existing Pitch surface/nav.
- [ ] GS5 active-play arm + Desk opportunity region.
- [ ] GS6 provisioning caps + owned-upgrade text on the Desk.
- [ ] GS0.3 event decoders (announcements only).
- [ ] Canonical docs, cold gates, hand-off for Codex designated review (never self-archive).
