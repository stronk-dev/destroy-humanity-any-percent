# Server Garden implementation log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude. **Review:** Codex designated cross-party review, pending. Never
self-approved.

Authority: `rfc/minigame-server-garden.md`, accepted 2026-09-25 with every OD at its recommended
default. That includes OD-1 wall-clock, OD-4 24 h catch-up hardcap, OD-5 no death, OD-6 Moore
adjacency, OD-7 lockout, OD-8 effect frozen at maturity, OD-12 hidden server salt, OD-13
`unrelated`, OD-17 one send per harvest, OD-18 per-family read, OD-20 trigger set, and OD-21 the
cumulative clamp.

Declared readings before any code:
- **Version numbering is landing-order.** The RFC's "Founder v22" is next-free, which at landing is
  **v25**. The `server_garden` artifact therefore sits on the scalar Founder chain above
  `cosmetics`: garden pinned ⟹ cosmetics pinned ⟹ … ⟹ `minigame_api`.
- **Clout:** under Clout v1's OD-1 option A, the garden carries no Clout payload.
- **Host generator (OD-14):** Tier 2 exists only as fixture candidates. The fixture host is the
  existing Fiscal `generator_level_rows` row `generator.beige_tower`. SG12 explicitly allows this
  ("the host is the fixture generator row, until Tier-2 content supplies the real one").
- **Fiscal fixture:** the fixture Fiscal artifact gains `{unlock_id: "minigame.server_garden",
  cost: 3}` in a garden fixture bundle only. Production Fiscal bytes are untouched.
- **API C2:** the read endpoint is a new operation (allowed widening). Garden commands are Founder
  intents on `/api/v1/intents`, which is not in the generated registry. Nothing widens an existing
  v1 union; if anything would, it is recorded here as blocked on the owner ruling.
- **Kernel protocol:** each commit touching a guarded path bumps `kernel/VERSION` in the same
  commit. New dirs `server/garden/` and `client/src/garden/` are registered as guarded.
