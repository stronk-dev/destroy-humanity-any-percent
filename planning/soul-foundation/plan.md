# Soul Foundation implementation plan

RFC: `rfc/soul-foundation.md`

- [x] Resolve SB1-SB9 and reconcile the normative body in the same ruling edit.
- [x] Resolve implementation blockers SB10-SB16, including once-only state and the Company
  suppression authority.
- [x] Wait for the exact Fiscal v19 artifact/activation chain before enabling Soul v20.
- [x] Resolve implementation blockers SB17-SB23: literal artifact enums, debit errors/events,
  recovery persistence/commands, the suppressed transition, v20 activation bytes, consumer schema
  versions, and cross-stream event order.
- [x] Implement the strict Soul artifact, dormant-to-active save transition, and pure band/gate API.
- [x] Implement the ruled atomic debit component and real recovery activity coordinator, including
  the SB24-SB27 progress capability, partition-safe presence accumulation, and lazy watchdog.
- [x] Add pet/minigame consumers, exact wire/event/replay vectors, and migration corpus.
- [x] Route the Trust/Soul correlation to the Balance Harness successor and ending semantic-copy
  protection to the Copy Pipeline successor (SB7-SB8).
- [ ] Run normal root gates and obtain both mandatory full-range reviews before docs/archive.

Current state: SB1-SB27 are implemented. Normal root verification and both mandatory full-range
reviews remain before archival; production recovery rows remain unminted until that gate closes.
