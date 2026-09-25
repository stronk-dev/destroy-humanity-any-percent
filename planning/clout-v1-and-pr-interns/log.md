# Clout v1 + PR Interns — implementation log

## 2026-09-25 — Predeclaration (Claude)

**Implemented by:** Claude, on the owner's direction that Claude implements accepted RFCs.
Awaiting Codex's designated cross-party review; nothing here is self-approved.

**Authority:** `rfc/clout-v1-and-pr-interns.md`, accepted 2026-09-25 at recommended defaults.

**Numbering at predeclaration (HEAD `c20d23f1`):** kernel 0.3.125; economy catalog schema 4 is
current and no artifact or candidate claims v5 (`balance/testdata/t2/economy-candidate-v1.json` is
schema 4), so this RFC takes **economy v5**; `LatestCompanyVersion = 18`, so this RFC takes
**Company v19**.

**Scope under A:** CV1–CV5, CV8 (fixture rows only), CV9 producer, CV10, AC4 with an empty writer
set. CV6/CV7 are not built (A rejects them). `pr_intern_3` stays out of all data (OD-5: it lands
with Tier 2 content or is dropped).

**Readings recorded before code:**
- **R1 — Input enum.** The loader accepts `achievement_attainment_run` (the ruled input) and
  `achievement_score_run` (needed as AC3's discriminating contrast arm; it needs no save change).
  `clout_run` always rejects, because no CV6 ledger exists under A (AC1 row "clout_run without the
  CV6 ledger").
- **R2 — `axis_at_least` storage.** It is extracted from an upgrade's `requires` before the
  remainder reaches `routes.ParseConditions`, and stored as the upgrade's own axis minimum. The route
  condition union never learns the kind, so route predicates and gates reject it by construction and
  the route context version is unchanged (CV1 item 3).
- **R3 — Rejection detail.** CV1 says an axis shortfall is `requirement_not_met`, while the shipped
  `buy_upgrade` reports every `requires` failure as `not_eligible/requires`. DESIGN-GAP DG-A: the
  implementation keeps the shipped `not_eligible/requires` pair (the shortfall is a `requires`
  failure; a new detail would widen the rejection taxonomy without a ruling). Owner/RFC author to
  confirm.
