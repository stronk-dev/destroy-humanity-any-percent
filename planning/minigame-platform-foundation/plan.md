# Minigame Platform Foundation implementation plan

RFC: `rfc/minigame-platform-foundation.md`

- [x] Acceptance-review the draft and reconcile owner rulings C1–C18 into the normative body.
- [x] Land the Postgres-authoritative session lifecycle and claim-token repository boundary.
  - [x] Append the `minigame_sessions` migration with closed modes/statuses and immutable frozen
    genesis fields.
  - [x] Prove concurrent claims, token-owned writes, terminal immutability, and frozen scaling
    inputs against real Postgres.
- [x] Land the closed tenant descriptor/registry boundary.
  - [x] Conformance-test a deterministic fixture tenant.
  - [x] Obtain independent approval of the claim-release, certified-result, schema-validation,
    identity-constraint, and canonical-integer remediation round.
  - [ ] Register the combat duel adapter only when its implemented engine surface satisfies the
    same boundary; do not invent the deferred lane engine.
- [ ] Land scaling-source validation, the ranked-power prohibition, fallback rows, and formula
  artifact output without production balance literals.
- [ ] Land server-certified resolve/payout, faucet accounting, replay identity, and fault tests.
- [ ] Compose the platform into the gameserver and prove solo/async-snapshot lifecycle ACs.
- [ ] Mint the production artifact only after owner/harness-supplied balance rows.
- [ ] Update canonical docs, pass normal repository-root verification, obtain independent
  full-range adversarial review, and archive.

Carried acceptance debt: production registry/faucet/offline-quality numbers are balance data and
remain a DESIGN-GAP until supplied by the owner/harness. The persistence and tenant boundaries use
only structural contracts already ruled in C13–C18.
