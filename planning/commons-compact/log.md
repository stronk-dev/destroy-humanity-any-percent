# Commons Compact — append-only log

## 2026-07-29 — acceptance and start

- Re-read RFC-0000, AGENTS.md, the complete Commons RFC, and the normative research formula/starting-value sections.
- Owner direction to implement the RFC queue is explicit acceptance. C1-C8 close the draft's executable gaps without changing its design: bounded ratios are integer ppm, large values remain RFC-0001 Decimals, and every provisional value is balance data.
- Marked the RFC implementing. No push will be performed.

## 2026-07-29 — pure catalog and numeric boundary

- Added the strict Commons catalog in Go/TypeScript and shipped the research starting values as data. Bounded ratios are ppm; the multiplier remains canonical Decimal.
- Implemented source-derived Enclosure and the Health/Solidarity modifier in both runtimes with shared edge vectors. Non-members return an absent contribution; members return exactly `commons.member` in the fixed Commons slot.
- Declared the one Commons multiplier source in the economy catalog and added schema/cross-catalog checks plus a compile-enforced two-way production boundary.
- Focused Go, TypeScript, 6,396 Node tests, schema verification, and the import gate pass. Review correction: the economy schema's `oneOf` made target `all` match both alternatives; changed it to `anyOf`, matching the already-canonical Go/TypeScript loader behavior.

## 2026-07-29 — save v6, membership intents, and cohort assignment

- Added save v6 company membership/tithe/Solidarity state with strict hourly sample validation and a corpus-backed v1-v5 migration to empty non-member state.
- Added strict `sign_compact` and `leave_compact` intent parsing, accrue-before-change semantics, typed repeat/non-member rejections, revision-tied membership events, and leave-time Solidarity erasure. Direct transition tests cover sign, duplicate sign, leave, and clean re-sign.
- Generalized post-commit event projection to support multiple independent projectors without coupling production to Commons.
- Added migrations and an idempotent Commons projector. Founder cohort assignment is server-context-derived, protected by a per-server/bracket advisory transaction lock, fills the oldest cohort below its catalog target, and persists across runs/leaves.
- Real Postgres concurrency tests prove two simultaneous first signs share one eligible cohort, replay does not inflate membership, leaving updates only run membership, and founder cohort identity remains stable.
- Batch review found no unresolved correctness issue. All Go packages and the full Postgres integration target pass.
