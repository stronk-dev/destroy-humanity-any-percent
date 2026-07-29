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

## 2026-07-29 — lazy Health/Capacity pipeline and population gate

- Added a neutral post-accrual hook contract. Production calls the hook but imports no Commons code; the composition adapter derives Enclosure from accepted multiplier contributions, advances hourly 30-day Solidarity buckets, computes tithed Capacity from committed accrual receipts, and emits `compact_sampled`.
- Added idempotent sample projection, member/cohort/server read models, labeled NPC denominator weight below 40 members, guildless effective-Health fallback, asymmetric closed-form Health smoothing, and the `commons.member` contribution provider. The first sample uses the declared neutral NPC Health; subsequent intents consume the projected snapshot.
- Added explicit collapse merge. Founder/member/sample rows move under one advisory lock, source cohorts close, and target standing is recomputed from member numerator/denominator inputs rather than averaged from rounded Health.
- Extended the production integration test through sign → member accrual → sample projection → Health snapshot; the real-Postgres suite and replay checks pass.
- Added the SplitMix64 population harness gate over 128 seeds. Mean modifiers at 200 vs 20,000 members remain within the shipped 100-ppm bound and their 95% intervals overlap.
- Formula generation now fingerprints the executable Enclosure, modifier, aggregate-Health, and smoothing authorities and publishes the Commons formula family plus the exact source-weight table from balance data.

## 2026-07-29 — ambient events and canonical documentation

- Added catalog-owned collapsed/strained/healthy thresholds. Server Health band crossings now append immutable band events; entering/leaving collapse adds cascade/recovery events.
- Added the idempotent mid-T3 recruitment operator. It verifies the authoritative company stream and refuses Founders who have ever signed or already received the offer; Postgres uniqueness proves once-per-career delivery.
- Expanded the closed Go/DB event registry and strict payload tests for the entire Commons family.
- Published `docs/commons.md`, updated production/save canonical docs for the six-intent surface and save v6, and updated repository/onboarding status. Formula generation publishes an explicit empty Phase-0 source-weight array rather than JSON null.
- Focused Go/TypeScript/schema tests and the full Postgres integration target pass after the dispatch migration.
