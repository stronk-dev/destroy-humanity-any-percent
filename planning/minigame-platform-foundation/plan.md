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
  - [x] Implement C28's exact one-row transform grammar, exact integer resolver, ranked-power
    rejection, and generated formula contract with focused tests.
  - [x] Implement C27's exact `solo|bot|npc_partner` fallback arms, identity/version validation,
    reduction bounds, and published formula contract.
  - [ ] Implement offline quality after the exact nested `grade_curve` row keys are owner-ruled;
    thresholds and decay values remain deferred balance literals.
- [ ] Land server-certified resolve/payout, faucet accounting, replay identity, and fault tests.
  - [x] Persist immutable applied-command rows and replay them from genesis before terminal
    resolution.
  - [x] Implement C30's exact payout-policy loader and exact carried-ppm conversion kernel without
    enabling a production faucet.
  - [x] Implement C32's declared certified-score selector and cap-copy validation with fail-closed
    missing/negative fact tests.
  - [ ] Compose C33 cap/window accounting and the multi-stream transaction.
- [ ] Compose the platform into the gameserver and prove solo/async-snapshot lifecycle ACs.
- [ ] Mint the production artifact only after owner/harness-supplied balance rows.
- [ ] Update canonical docs, pass normal repository-root verification, obtain independent
  full-range adversarial review, and archive.

Carried acceptance debt: production registry/faucet/offline-quality numbers are balance data and
remain a DESIGN-GAP until supplied by the owner/harness. The persistence and tenant boundaries use
only structural contracts already ruled in C13–C18.

Implementation blockers C19–C23: exact scaling-source rows, immutable session command history,
the production payout envelope, attended-window/counter arithmetic, and complete fallback/
offline-quality rows require owner rulings before the remaining platform slices can be written.

Post-ruling implementation blockers C24-C27: the transform grammar is still not exact-key, the
Company+Founder+session resolution has conflicting transaction/lock ownership, its faucet depends
on the absent Founder attendance authority, and fallback/offline rows still contain unnamed
objects. C20's command-history slice is complete; the remaining platform is blocked on these
structural rulings rather than balance literals.

Founder Attendance now closes C26's clock dependency. C28-C30 carry the remaining literal
contracts: the actual transform key/operation grammar, Founder rating persistence/replay, and the
faucet policy/idempotency schema. No payout or production catalog row lands before those rulings.

Post-C27 implementation gap: the fallback arms are exact and implemented, but the offline policy's
`grade_curve` is still only a noun. The RFC must enumerate that nested row's keys and ordering/
duplicate rules before the loader can distinguish valid data from an invented schema.

Post-C30 implementation gaps were ruled as C31-C33. C32's score/copy ownership is implemented;
C33's cross-run window and C31's still-incomplete curve wire remain open.
