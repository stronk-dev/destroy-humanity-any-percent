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
  - [x] Implement C34's exact offline-quality row, score-to-grade selector, replay-state validator,
    and generated formula contract without production balance literals.
  - [ ] Compose attended-grid offline-quality decay after the Founder-axis artifact activation
    seam lands; thresholds and decay values remain deferred balance literals.
- [ ] Land server-certified resolve/payout, faucet accounting, replay identity, and fault tests.
  - [x] Persist immutable applied-command rows and replay them from genesis before terminal
    resolution.
  - [x] Implement C30's exact payout-policy loader and exact carried-ppm conversion kernel without
    enabling a production faucet.
  - [x] Implement C32's declared certified-score selector and cap-copy validation with fail-closed
    missing/negative fact tests.
  - [x] Append C33's cross-run attended-day window and prove carried conversion, send/per-send
    caps, new-day reset, database bounds, and transaction rollback against Postgres.
  - [ ] Compose Company payout, Founder rating, token-owned session resolution, and the window in
    one Founder→Company→session transaction.
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

Post-C33 status: the payout selector and window authority are implemented. C34 narrows C31's
ellipsis to one proposed exact grade-curve row; C35 carries the Founder-only version/artifact seam
shared by rating, offline-quality, and Pet C16. The atomic resolve composer stays blocked until
that state can be replay-owned without version collision.

Post-C35 infrastructure status: Exit's version-tuple validator now accepts independent,
pinned-bundle-derived Founder and Company floors. The production `minigames` artifact and exact
Founder rating/quality maps still need an owner-assigned version and complete artifact row; the
platform does not infer either from the structural fixture loaders.

Blocked implementation contract carried explicitly: C36 must enumerate the Founder rating row,
season-fact wire, enclosing map, and its ordered version/artifact biconditional before the atomic
resolve composer can write it replay-safely.

Post-C36 status: [x] Founder v17, the exact rating/offline-quality maps, pinned minigames artifact,
Go/TypeScript replay derivation, and reachable Founder-v17/Company-v16 Exit path are implemented
without production rows or a balance mint. The server-certified resolve composer remains next.
